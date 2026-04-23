package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/matheuscoutinhoo/better/internal/games"
	"github.com/matheuscoutinhoo/better/internal/models"
	"github.com/matheuscoutinhoo/better/internal/odds"
)

type Scraper struct {
	gamesRepo *games.Repository
	oddsRepo  *odds.Repository
	userAgent string
	oddsKey   string
	oddsURL   string
	httpCli   *http.Client
}

func NewScraper(gamesRepo *games.Repository, oddsRepo *odds.Repository, userAgent, oddsKey, oddsURL string) *Scraper {
	return &Scraper{
		gamesRepo: gamesRepo,
		oddsRepo:  oddsRepo,
		userAgent: userAgent,
		oddsKey:   oddsKey,
		oddsURL:   oddsURL,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// balldontlie.io free API for NBA data
const nbaAPIBase = "https://api.balldontlie.io/v1"

type bdlGamesResponse struct {
	Data []bdlGame `json:"data"`
	Meta bdlMeta   `json:"meta"`
}

type bdlGame struct {
	ID               int     `json:"id"`
	Date             string  `json:"date"`
	HomeTeam         bdlTeam `json:"home_team"`
	VisitorTeam      bdlTeam `json:"visitor_team"`
	HomeTeamScore    int     `json:"home_team_score"`
	VisitorTeamScore int     `json:"visitor_team_score"`
	Status           string  `json:"status"`
	Season           int     `json:"season"`
}

type bdlTeam struct {
	ID           int    `json:"id"`
	Abbreviation string `json:"abbreviation"`
	City         string `json:"city"`
	FullName     string `json:"full_name"`
	Name         string `json:"name"`
}

type bdlMeta struct {
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	NextPage    *int `json:"next_page"`
	PerPage     int  `json:"per_page"`
	TotalCount  int  `json:"total_count"`
}

func (s *Scraper) ScrapeRecentGames() error {
	log.Println("Starting NBA games scrape...")
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	page := 1
	for {
		url := fmt.Sprintf("%s/games?start_date=%s&end_date=%s&per_page=100&page=%d",
			nbaAPIBase,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			page,
		)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", s.userAgent)

		resp, err := s.httpCli.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch games: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("NBA API returned status %d, using simulated data", resp.StatusCode)
			return s.seedSimulatedGames()
		}

		var gamesResp bdlGamesResponse
		if err := json.Unmarshal(body, &gamesResp); err != nil {
			log.Printf("Failed to parse NBA API response, using simulated data: %v", err)
			return s.seedSimulatedGames()
		}

		for _, g := range gamesResp.Data {
			gameDate, _ := time.Parse("2006-01-02T15:04:05.000Z", g.Date)
			if gameDate.IsZero() {
				gameDate, _ = time.Parse("2006-01-02", g.Date)
			}

			status := "final"
			if g.HomeTeamScore == 0 && g.VisitorTeamScore == 0 {
				status = "scheduled"
			}

			homeScore := g.HomeTeamScore
			awayScore := g.VisitorTeamScore

			game := &models.NBAGame{
				ExternalID: fmt.Sprintf("bdl_%d", g.ID),
				GameDate:   gameDate,
				HomeTeam:   g.HomeTeam.FullName,
				AwayTeam:   g.VisitorTeam.FullName,
				HomeScore:  &homeScore,
				AwayScore:  &awayScore,
				Status:     status,
				Season:     fmt.Sprintf("%d-%d", g.Season, g.Season+1),
			}

			if _, err := s.gamesRepo.UpsertGame(game); err != nil {
				log.Printf("Failed to upsert game %s: %v", game.ExternalID, err)
			}
		}

		if gamesResp.Meta.NextPage == nil {
			break
		}
		page = *gamesResp.Meta.NextPage
	}

	log.Printf("NBA games scrape completed")
	return nil
}

func (s *Scraper) ScrapeOdds() error {
	if s.oddsKey == "" {
		log.Println("No odds API key configured, seeding simulated odds")
		return s.seedSimulatedOdds()
	}

	url := fmt.Sprintf("%s/sports/basketball_nba/odds/?apiKey=%s&regions=us,eu,uk&markets=h2h,totals&oddsFormat=decimal",
		s.oddsURL, s.oddsKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		log.Printf("Failed to fetch odds, using simulated data: %v", err)
		return s.seedSimulatedOdds()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Odds API returned %d: %s, using simulated data", resp.StatusCode, string(body))
		return s.seedSimulatedOdds()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var oddsEvents []oddsAPIEvent
	if err := json.Unmarshal(body, &oddsEvents); err != nil {
		return fmt.Errorf("failed to parse odds: %w", err)
	}

	// Pre-fetch all games once
	recentGames, _ := s.gamesRepo.GetRecentGames(30)
	upcomingGames, _ := s.gamesRepo.GetUpcomingGames(7)
	allGames := append(recentGames, upcomingGames...)

	log.Printf("Odds API returned %d events, matching against %d games in DB", len(oddsEvents), len(allGames))

	matched := 0
	for _, event := range oddsEvents {
		foundMatch := false
		for _, game := range allGames {
			if matchesGame(event, game) {
				for _, bookmaker := range event.Bookmakers {
					for _, market := range bookmaker.Markets {
						odd := &models.GameOdds{
							GameID:     game.ID,
							Bookmaker:  bookmaker.Title,
							MarketType: market.Key,
						}
						for _, outcome := range market.Outcomes {
							switch outcome.Name {
							case event.HomeTeam:
								odd.HomeOdd = &outcome.Price
							case event.AwayTeam:
								odd.AwayOdd = &outcome.Price
							case "Over":
								odd.OverOdd = &outcome.Price
								odd.OverUnderLine = outcome.Point
							case "Under":
								odd.UnderOdd = &outcome.Price
							}
						}
						s.oddsRepo.UpsertOdds(odd)
					}
				}
				matched++
				foundMatch = true
				break
			}
		}
		// If no matching game, create one from the odds event
		if !foundMatch {
			commenceTime := time.Now().Add(24 * time.Hour) // default future
			if event.CommenceTime != "" {
				if t, err := time.Parse(time.RFC3339, event.CommenceTime); err == nil {
					commenceTime = t
				}
			}
			status := "scheduled"
			if commenceTime.Before(time.Now()) {
				status = "live"
			}
			newGame := &models.NBAGame{
				ExternalID: event.ID,
				GameDate:   commenceTime,
				HomeTeam:   event.HomeTeam,
				AwayTeam:   event.AwayTeam,
				Status:     status,
				Season:     "2025-2026",
			}
			gameID, err := s.gamesRepo.UpsertGame(newGame)
			if err != nil {
				log.Printf("Failed to create game for odds event %s vs %s: %v", event.HomeTeam, event.AwayTeam, err)
				continue
			}
			for _, bookmaker := range event.Bookmakers {
				for _, market := range bookmaker.Markets {
					odd := &models.GameOdds{
						GameID:     gameID,
						Bookmaker:  bookmaker.Title,
						MarketType: market.Key,
					}
					for _, outcome := range market.Outcomes {
						switch outcome.Name {
						case event.HomeTeam:
							odd.HomeOdd = &outcome.Price
						case event.AwayTeam:
							odd.AwayOdd = &outcome.Price
						case "Over":
							odd.OverOdd = &outcome.Price
							odd.OverUnderLine = outcome.Point
						case "Under":
							odd.UnderOdd = &outcome.Price
						}
					}
					s.oddsRepo.UpsertOdds(odd)
				}
			}
			matched++
			log.Printf("Created game + odds for: %s vs %s", event.HomeTeam, event.AwayTeam)
		}
	}

	log.Printf("Odds scrape completed: %d events from API, %d matched to games", len(oddsEvents), matched)
	return nil
}

type oddsAPIEvent struct {
	ID            string          `json:"id"`
	CommenceTime  string          `json:"commence_time"`
	HomeTeam      string          `json:"home_team"`
	AwayTeam      string          `json:"away_team"`
	Bookmakers    []oddsBookmaker `json:"bookmakers"`
}

type oddsBookmaker struct {
	Key     string       `json:"key"`
	Title   string       `json:"title"`
	Markets []oddsMarket `json:"markets"`
}

type oddsMarket struct {
	Key      string        `json:"key"`
	Outcomes []oddsOutcome `json:"outcomes"`
}

type oddsOutcome struct {
	Name  string   `json:"name"`
	Price float64  `json:"price"`
	Point *float64 `json:"point,omitempty"`
}

func matchesGame(event oddsAPIEvent, game models.NBAGame) bool {
	eHome := strings.ToLower(event.HomeTeam)
	gHome := strings.ToLower(game.HomeTeam)
	eAway := strings.ToLower(event.AwayTeam)
	gAway := strings.ToLower(game.AwayTeam)
	return (eHome == gHome || strings.Contains(eHome, gHome) || strings.Contains(gHome, eHome)) &&
		(eAway == gAway || strings.Contains(eAway, gAway) || strings.Contains(gAway, eAway))
}

// Simulated data for when APIs aren't available
func (s *Scraper) seedSimulatedGames() error {
	log.Println("Seeding simulated NBA games...")

	teams := []struct{ name string }{
		{"Los Angeles Lakers"}, {"Boston Celtics"}, {"Golden State Warriors"},
		{"Milwaukee Bucks"}, {"Denver Nuggets"}, {"Phoenix Suns"},
		{"Philadelphia 76ers"}, {"Miami Heat"}, {"Dallas Mavericks"},
		{"Memphis Grizzlies"}, {"Sacramento Kings"}, {"Cleveland Cavaliers"},
		{"New York Knicks"}, {"Brooklyn Nets"}, {"Atlanta Hawks"},
		{"Minnesota Timberwolves"}, {"Oklahoma City Thunder"}, {"New Orleans Pelicans"},
		{"Los Angeles Clippers"}, {"Chicago Bulls"},
	}

	now := time.Now()
	gameCount := 0

	// Past 30 days
	for day := 30; day >= 1; day-- {
		date := now.AddDate(0, 0, -day)
		numGames := 3 + (day % 4) // 3-6 games per day
		for i := 0; i < numGames && i*2+1 < len(teams); i++ {
			homeIdx := (day*3 + i*2) % len(teams)
			awayIdx := (day*3 + i*2 + 1) % len(teams)
			if homeIdx == awayIdx {
				awayIdx = (awayIdx + 1) % len(teams)
			}

			homeScore := 95 + (day+i*7)%30
			awayScore := 90 + (day+i*5)%35

			game := &models.NBAGame{
				ExternalID: fmt.Sprintf("sim_%d_%d", day, i),
				GameDate:   date,
				HomeTeam:   teams[homeIdx].name,
				AwayTeam:   teams[awayIdx].name,
				HomeScore:  &homeScore,
				AwayScore:  &awayScore,
				Status:     "final",
				Season:     "2025-2026",
			}
			if _, err := s.gamesRepo.UpsertGame(game); err != nil {
				log.Printf("Failed to seed game: %v", err)
			}
			gameCount++
		}
	}

	// Upcoming 7 days
	for day := 1; day <= 7; day++ {
		date := now.AddDate(0, 0, day)
		numGames := 3 + (day % 3)
		for i := 0; i < numGames && i*2+1 < len(teams); i++ {
			homeIdx := (day*5 + i*2) % len(teams)
			awayIdx := (day*5 + i*2 + 1) % len(teams)
			if homeIdx == awayIdx {
				awayIdx = (awayIdx + 1) % len(teams)
			}

			game := &models.NBAGame{
				ExternalID: fmt.Sprintf("sim_future_%d_%d", day, i),
				GameDate:   date,
				HomeTeam:   teams[homeIdx].name,
				AwayTeam:   teams[awayIdx].name,
				Status:     "scheduled",
				Season:     "2025-2026",
			}
			if _, err := s.gamesRepo.UpsertGame(game); err != nil {
				log.Printf("Failed to seed future game: %v", err)
			}
			gameCount++
		}
	}

	log.Printf("Seeded %d simulated games", gameCount)
	return nil
}

func (s *Scraper) seedSimulatedOdds() error {
	log.Println("Seeding simulated odds...")

	upcoming, err := s.gamesRepo.GetUpcomingGames(7)
	if err != nil {
		return err
	}

	for i, game := range upcoming {
		homeOdd := 1.5 + float64(i%5)*0.3
		awayOdd := 2.0 + float64(i%4)*0.25
		overLine := 215.5 + float64(i%10)
		overOdd := 1.90 + float64(i%3)*0.05
		underOdd := 1.90 + float64(i%4)*0.05

		// Moneyline odds
		s.oddsRepo.UpsertOdds(&models.GameOdds{
			GameID:     game.ID,
			Bookmaker:  "simulated",
			MarketType: "h2h",
			HomeOdd:    &homeOdd,
			AwayOdd:    &awayOdd,
		})

		// Over/under odds
		s.oddsRepo.UpsertOdds(&models.GameOdds{
			GameID:        game.ID,
			Bookmaker:     "simulated",
			MarketType:    "totals",
			OverUnderLine: &overLine,
			OverOdd:       &overOdd,
			UnderOdd:      &underOdd,
		})
	}

	log.Println("Simulated odds seeded")
	return nil
}

func (s *Scraper) StartCronJob(intervalHours int) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	go func() {
		// Run immediately on start
		s.RunAll()
		for range ticker.C {
			s.RunAll()
		}
	}()
}

func (s *Scraper) RunAll() {
	log.Println("Running scheduled scrape...")
	if err := s.ScrapeRecentGames(); err != nil {
		log.Printf("Error scraping games: %v", err)
	}
	if err := s.ScrapeOdds(); err != nil {
		log.Printf("Error scraping odds: %v", err)
	}
}
