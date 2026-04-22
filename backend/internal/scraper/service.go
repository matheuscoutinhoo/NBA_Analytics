package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/better/backend/pkg/logger"
)

type Service struct {
	db          *pgxpool.Pool
	logger      *logger.Logger
	nbaStatsURL string
	oddsAPIKey  string
	httpClient  *http.Client
}

func NewService(db *pgxpool.Pool, log *logger.Logger, nbaStatsURL, oddsAPIKey string) *Service {
	return &Service{
		db:          db,
		logger:      log,
		nbaStatsURL: nbaStatsURL,
		oddsAPIKey:  oddsAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type NBAGame struct {
	GameID     string    `json:"gameId"`
	HomeTeamID string    `json:"homeTeamId"`
	AwayTeamID string    `json:"awayTeamId"`
	GameDate   time.Time `json:"gameDate"`
	Season     string    `json:"season"`
	Status     string    `json:"status"`
	HomeScore  *int      `json:"homeScore"`
	AwayScore  *int      `json:"awayScore"`
}

type TeamStats struct {
	TeamID    string  `json:"teamId"`
	FGPct     float64 `json:"fgPct"`
	FG3Pct    float64 `json:"fg3Pct"`
	FTPct     float64 `json:"ftPct"`
	Rebounds  int     `json:"rebounds"`
	Assists   int     `json:"assists"`
	Turnovers int     `json:"turnovers"`
	Steals    int     `json:"steals"`
	Blocks    int     `json:"blocks"`
	Points    int     `json:"points"`
}

type OddsData struct {
	GameID        string  `json:"game_id"`
	Source        string  `json:"source"`
	HomeMoneyline float64 `json:"home_moneyline"`
	AwayMoneyline float64 `json:"away_moneyline"`
	HomeSpread    float64 `json:"home_spread"`
	AwaySpread    float64 `json:"away_spread"`
	OverUnder     float64 `json:"over_under"`
}

type InjuryData struct {
	PlayerID    string `json:"player_id"`
	TeamID      string `json:"team_id"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// RunDailyScrape fetches today's games, stats, injuries
func (s *Service) RunDailyScrape(ctx context.Context) error {
	idempotencyKey := fmt.Sprintf("daily_scrape_%s", time.Now().UTC().Format("2006-01-02"))

	// Check if already run
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM scraping_jobs WHERE idempotency_key = $1 AND status = 'COMPLETED')`,
		idempotencyKey,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("checking idempotency: %w", err)
	}
	if exists {
		s.logger.Info("daily scrape already completed for today")
		return nil
	}

	jobID := uuid.New()
	_, err = s.db.Exec(ctx,
		`INSERT INTO scraping_jobs (id, job_type, status, idempotency_key, started_at) VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (idempotency_key) DO NOTHING`,
		jobID, "DAILY_SCRAPE", "RUNNING", idempotencyKey,
	)
	if err != nil {
		return fmt.Errorf("creating scraping job: %w", err)
	}

	var recordsProcessed int

	// 1. Fetch today's games
	games, err := s.fetchTodaysGames(ctx)
	if err != nil {
		s.markJobFailed(ctx, idempotencyKey, err.Error())
		return fmt.Errorf("fetching games: %w", err)
	}

	for _, game := range games {
		if err := s.upsertGame(ctx, game); err != nil {
			s.logger.Error("failed to upsert game", map[string]interface{}{
				"game_id": game.GameID,
				"error":   err.Error(),
			})
			continue
		}
		recordsProcessed++
	}

	// 2. Fetch odds
	odds, err := s.fetchOdds(ctx)
	if err != nil {
		s.logger.Warn("failed to fetch odds, continuing", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, odd := range odds {
			if err := s.upsertOdds(ctx, odd); err != nil {
				s.logger.Error("failed to upsert odds", map[string]interface{}{"error": err.Error()})
				continue
			}
			recordsProcessed++
		}
	}

	// 3. Fetch injuries
	injuries, err := s.fetchInjuries(ctx)
	if err != nil {
		s.logger.Warn("failed to fetch injuries, continuing", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, injury := range injuries {
			if err := s.upsertInjury(ctx, injury); err != nil {
				s.logger.Error("failed to upsert injury", map[string]interface{}{"error": err.Error()})
				continue
			}
			recordsProcessed++
		}
	}

	// Mark completed
	s.db.Exec(ctx,
		`UPDATE scraping_jobs SET status = 'COMPLETED', records_processed = $2, completed_at = NOW() WHERE idempotency_key = $1`,
		idempotencyKey, recordsProcessed,
	)

	s.logger.Info("daily scrape completed", map[string]interface{}{
		"records_processed": recordsProcessed,
	})

	return nil
}

// RunResultsUpdate fetches final scores for completed games
func (s *Service) RunResultsUpdate(ctx context.Context) error {
	idempotencyKey := fmt.Sprintf("results_update_%s_%s", time.Now().UTC().Format("2006-01-02"), time.Now().UTC().Format("15"))

	var exists bool
	s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM scraping_jobs WHERE idempotency_key = $1 AND status = 'COMPLETED')`,
		idempotencyKey,
	).Scan(&exists)
	if exists {
		return nil
	}

	jobID := uuid.New()
	s.db.Exec(ctx,
		`INSERT INTO scraping_jobs (id, job_type, status, idempotency_key, started_at) VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (idempotency_key) DO NOTHING`,
		jobID, "RESULTS_UPDATE", "RUNNING", idempotencyKey,
	)

	// Fetch yesterday's and today's game results
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().UTC().Format("2006-01-02")

	games, err := s.fetchGameResults(ctx, yesterday, today)
	if err != nil {
		s.markJobFailed(ctx, idempotencyKey, err.Error())
		return err
	}

	recordsProcessed := 0
	for _, game := range games {
		if err := s.updateGameResult(ctx, game); err != nil {
			s.logger.Error("failed to update game result", map[string]interface{}{
				"game_id": game.GameID,
				"error":   err.Error(),
			})
			continue
		}
		recordsProcessed++
	}

	s.db.Exec(ctx,
		`UPDATE scraping_jobs SET status = 'COMPLETED', records_processed = $2, completed_at = NOW() WHERE idempotency_key = $1`,
		idempotencyKey, recordsProcessed,
	)

	return nil
}

func (s *Service) fetchTodaysGames(ctx context.Context) ([]NBAGame, error) {
	today := time.Now().UTC().Format("2006-01-02")
	url := fmt.Sprintf("https://stats.nba.com/stats/scoreboardv3?GameDate=%s&LeagueID=00", today)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.nba.com/")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NBA API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Parse the NBA API response into NBAGame structs
	var games []NBAGame
	scoreboard, ok := result["scoreboard"].(map[string]interface{})
	if !ok {
		return games, nil
	}

	gamesList, ok := scoreboard["games"].([]interface{})
	if !ok {
		return games, nil
	}

	for _, g := range gamesList {
		gameMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		game := NBAGame{
			GameID: fmt.Sprintf("%v", gameMap["gameId"]),
			Season: fmt.Sprintf("%v", gameMap["season"]),
		}

		if homeTeam, ok := gameMap["homeTeam"].(map[string]interface{}); ok {
			game.HomeTeamID = fmt.Sprintf("%v", homeTeam["teamId"])
			if score, ok := homeTeam["score"].(float64); ok && score > 0 {
				scoreInt := int(score)
				game.HomeScore = &scoreInt
			}
		}
		if awayTeam, ok := gameMap["awayTeam"].(map[string]interface{}); ok {
			game.AwayTeamID = fmt.Sprintf("%v", awayTeam["teamId"])
			if score, ok := awayTeam["score"].(float64); ok && score > 0 {
				scoreInt := int(score)
				game.AwayScore = &scoreInt
			}
		}

		if status, ok := gameMap["gameStatusText"].(string); ok {
			if status == "Final" {
				game.Status = "FINAL"
			} else {
				game.Status = "SCHEDULED"
			}
		}

		if dateStr, ok := gameMap["gameDateTimeUTC"].(string); ok {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				game.GameDate = t
			}
		}

		games = append(games, game)
	}

	return games, nil
}

func (s *Service) fetchOdds(ctx context.Context) ([]OddsData, error) {
	if s.oddsAPIKey == "" {
		return nil, fmt.Errorf("odds API key not configured")
	}

	url := fmt.Sprintf("https://api.the-odds-api.com/v4/sports/basketball_nba/odds/?apiKey=%s&regions=us&markets=h2h,spreads,totals&oddsFormat=decimal", s.oddsAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("odds API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiGames []map[string]interface{}
	if err := json.Unmarshal(body, &apiGames); err != nil {
		return nil, err
	}

	var odds []OddsData
	for _, game := range apiGames {
		odd := OddsData{
			GameID: fmt.Sprintf("%v", game["id"]),
			Source: "the-odds-api",
		}

		if bookmakers, ok := game["bookmakers"].([]interface{}); ok && len(bookmakers) > 0 {
			if bm, ok := bookmakers[0].(map[string]interface{}); ok {
				if markets, ok := bm["markets"].([]interface{}); ok {
					for _, m := range markets {
						market, ok := m.(map[string]interface{})
						if !ok {
							continue
						}
						key := fmt.Sprintf("%v", market["key"])
						outcomes, ok := market["outcomes"].([]interface{})
						if !ok || len(outcomes) < 2 {
							continue
						}
						switch key {
						case "h2h":
							if o, ok := outcomes[0].(map[string]interface{}); ok {
								if price, ok := o["price"].(float64); ok {
									odd.HomeMoneyline = price
								}
							}
							if o, ok := outcomes[1].(map[string]interface{}); ok {
								if price, ok := o["price"].(float64); ok {
									odd.AwayMoneyline = price
								}
							}
						case "spreads":
							if o, ok := outcomes[0].(map[string]interface{}); ok {
								if point, ok := o["point"].(float64); ok {
									odd.HomeSpread = point
								}
							}
							if o, ok := outcomes[1].(map[string]interface{}); ok {
								if point, ok := o["point"].(float64); ok {
									odd.AwaySpread = point
								}
							}
						case "totals":
							if o, ok := outcomes[0].(map[string]interface{}); ok {
								if point, ok := o["point"].(float64); ok {
									odd.OverUnder = point
								}
							}
						}
					}
				}
			}
		}

		odds = append(odds, odd)
	}

	return odds, nil
}

func (s *Service) fetchInjuries(ctx context.Context) ([]InjuryData, error) {
	// This would call a real injury feed. Placeholder implementation.
	return []InjuryData{}, nil
}

func (s *Service) fetchGameResults(ctx context.Context, from, to string) ([]NBAGame, error) {
	// Fetch games for both dates
	var allGames []NBAGame
	for _, date := range []string{from, to} {
		games, err := s.fetchGamesForDate(ctx, date)
		if err != nil {
			s.logger.Warn("failed to fetch games for date", map[string]interface{}{"date": date, "error": err.Error()})
			continue
		}
		allGames = append(allGames, games...)
	}
	return allGames, nil
}

func (s *Service) fetchGamesForDate(ctx context.Context, date string) ([]NBAGame, error) {
	url := fmt.Sprintf("https://stats.nba.com/stats/scoreboardv3?GameDate=%s&LeagueID=00", date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.nba.com/")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	var games []NBAGame
	// Parse similar to fetchTodaysGames
	scoreboard, ok := result["scoreboard"].(map[string]interface{})
	if !ok {
		return games, nil
	}
	gamesList, ok := scoreboard["games"].([]interface{})
	if !ok {
		return games, nil
	}

	for _, g := range gamesList {
		gameMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		game := NBAGame{
			GameID: fmt.Sprintf("%v", gameMap["gameId"]),
			Status: "FINAL",
		}
		if homeTeam, ok := gameMap["homeTeam"].(map[string]interface{}); ok {
			game.HomeTeamID = fmt.Sprintf("%v", homeTeam["teamId"])
			if score, ok := homeTeam["score"].(float64); ok {
				scoreInt := int(score)
				game.HomeScore = &scoreInt
			}
		}
		if awayTeam, ok := gameMap["awayTeam"].(map[string]interface{}); ok {
			game.AwayTeamID = fmt.Sprintf("%v", awayTeam["teamId"])
			if score, ok := awayTeam["score"].(float64); ok {
				scoreInt := int(score)
				game.AwayScore = &scoreInt
			}
		}
		games = append(games, game)
	}

	return games, nil
}

func (s *Service) upsertGame(ctx context.Context, game NBAGame) error {
	// Lookup team UUIDs
	var homeTeamUUID, awayTeamUUID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT id FROM teams WHERE external_id = $1`, game.HomeTeamID).Scan(&homeTeamUUID)
	if err != nil {
		return fmt.Errorf("home team not found: %s", game.HomeTeamID)
	}
	err = s.db.QueryRow(ctx, `SELECT id FROM teams WHERE external_id = $1`, game.AwayTeamID).Scan(&awayTeamUUID)
	if err != nil {
		return fmt.Errorf("away team not found: %s", game.AwayTeamID)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO games (id, external_id, home_team_id, away_team_id, game_date, season, status, home_score, away_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (external_id) DO UPDATE SET status = $7, home_score = $8, away_score = $9, updated_at = NOW()`,
		uuid.New(), game.GameID, homeTeamUUID, awayTeamUUID, game.GameDate, game.Season, game.Status, game.HomeScore, game.AwayScore,
	)
	return err
}

func (s *Service) upsertOdds(ctx context.Context, odd OddsData) error {
	var gameUUID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT id FROM games WHERE external_id = $1`, odd.GameID).Scan(&gameUUID)
	if err != nil {
		return fmt.Errorf("game not found for odds: %s", odd.GameID)
	}

	homeImplied := 0.0
	awayImplied := 0.0
	if odd.HomeMoneyline > 0 {
		homeImplied = 1.0 / odd.HomeMoneyline
	}
	if odd.AwayMoneyline > 0 {
		awayImplied = 1.0 / odd.AwayMoneyline
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO odds (game_id, source, home_moneyline, away_moneyline, home_spread, away_spread, over_under, home_implied_prob, away_implied_prob)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		gameUUID, odd.Source, odd.HomeMoneyline, odd.AwayMoneyline, odd.HomeSpread, odd.AwaySpread, odd.OverUnder, homeImplied, awayImplied,
	)
	return err
}

func (s *Service) upsertInjury(ctx context.Context, injury InjuryData) error {
	var playerUUID, teamUUID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT id FROM players WHERE external_id = $1`, injury.PlayerID).Scan(&playerUUID)
	if err != nil {
		return fmt.Errorf("player not found: %s", injury.PlayerID)
	}
	err = s.db.QueryRow(ctx, `SELECT id FROM teams WHERE external_id = $1`, injury.TeamID).Scan(&teamUUID)
	if err != nil {
		return fmt.Errorf("team not found: %s", injury.TeamID)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO injuries (player_id, team_id, status, description, updated_at)
		VALUES ($1, $2, $3, $4, NOW())`,
		playerUUID, teamUUID, injury.Status, injury.Description,
	)
	return err
}

func (s *Service) updateGameResult(ctx context.Context, game NBAGame) error {
	_, err := s.db.Exec(ctx,
		`UPDATE games SET status = $2, home_score = $3, away_score = $4, updated_at = NOW() WHERE external_id = $1`,
		game.GameID, game.Status, game.HomeScore, game.AwayScore,
	)
	return err
}

func (s *Service) markJobFailed(ctx context.Context, key, errMsg string) {
	s.db.Exec(ctx,
		`UPDATE scraping_jobs SET status = 'FAILED', error_message = $2, completed_at = NOW() WHERE idempotency_key = $1`,
		key, errMsg,
	)
}
