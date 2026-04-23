package predictions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/matheuscoutinhoo/better/internal/abacus"
	"github.com/matheuscoutinhoo/better/internal/games"
	"github.com/matheuscoutinhoo/better/internal/models"
	"github.com/matheuscoutinhoo/better/internal/odds"
)

type Service struct {
	gamesRepo *games.Repository
	oddsRepo  *odds.Repository
	aiClient  *abacus.Client
	db        *sql.DB
}

func NewService(gamesRepo *games.Repository, oddsRepo *odds.Repository, aiClient *abacus.Client, db *sql.DB) *Service {
	return &Service{
		gamesRepo: gamesRepo,
		oddsRepo:  oddsRepo,
		aiClient:  aiClient,
		db:        db,
	}
}

type aiAnalysisResult struct {
	WinProbability struct {
		Home float64 `json:"home"`
		Away float64 `json:"away"`
	} `json:"win_probability"`
	RecommendedBets []struct {
		Market     string `json:"market"`
		Pick       string `json:"pick"`
		Confidence string `json:"confidence"`
		Reasoning  string `json:"reasoning"`
	} `json:"recommended_bets"`
	RiskLevel  string   `json:"risk_level"`
	KeyFactors []string `json:"key_factors"`
	Summary    string   `json:"summary"`
}

// RunAnalysis analyzes all upcoming games with AI + odds context
func (s *Service) RunAnalysis() {
	log.Println("Starting hourly AI predictions analysis...")

	upcoming, err := s.gamesRepo.GetUpcomingGames(7)
	if err != nil {
		log.Printf("Failed to get upcoming games for analysis: %v", err)
		return
	}

	if len(upcoming) == 0 {
		log.Println("No upcoming games to analyze")
		return
	}

	// Clear old predictions
	_, _ = s.db.Exec("DELETE FROM ai_predictions")

	analyzed := 0
	for _, game := range upcoming {
		if err := s.analyzeGame(game); err != nil {
			log.Printf("Failed to analyze game %s vs %s: %v", game.HomeTeam, game.AwayTeam, err)
			continue
		}
		analyzed++
	}

	log.Printf("AI predictions analysis complete: %d/%d games analyzed", analyzed, len(upcoming))
}

func (s *Service) analyzeGame(game models.NBAGame) error {
	// Gather context
	homeRecent, _ := s.gamesRepo.GetTeamRecentGames(game.HomeTeam, 10)
	awayRecent, _ := s.gamesRepo.GetTeamRecentGames(game.AwayTeam, 10)
	h2h, _ := s.gamesRepo.GetHeadToHead(game.HomeTeam, game.AwayTeam, 6)
	gameOdds, _ := s.oddsRepo.GetOddsByGameID(game.ID)

	// Build prompt
	prompt := s.aiClient.BuildGameAnalysisPrompt(&game, homeRecent, awayRecent, h2h, gameOdds)

	// Call AI
	response, err := s.aiClient.Chat(prompt)
	if err != nil {
		return fmt.Errorf("AI call failed: %w", err)
	}

	// Parse AI response
	var result aiAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Try to extract JSON from markdown code blocks
		cleaned := extractJSON(response)
		if err2 := json.Unmarshal([]byte(cleaned), &result); err2 != nil {
			// Store raw response as summary with default probabilities
			return s.storePrediction(game, gameOdds, aiAnalysisResult{
				Summary: response,
				WinProbability: struct {
					Home float64 `json:"home"`
					Away float64 `json:"away"`
				}{Home: 0.5, Away: 0.5},
				KeyFactors: []string{"Analysis available - see summary"},
			})
		}
	}

	return s.storePrediction(game, gameOdds, result)
}

func (s *Service) storePrediction(game models.NBAGame, gameOdds []models.GameOdds, result aiAnalysisResult) error {
	// Find bet365 odds (or best available)
	var bestOdds *models.GameOdds
	for i, o := range gameOdds {
		if o.Bookmaker == "Bet365" || o.Bookmaker == "bet365" {
			bestOdds = &gameOdds[i]
			break
		}
	}
	// Fallback to first h2h odds
	if bestOdds == nil {
		for i, o := range gameOdds {
			if o.MarketType == "h2h" && o.HomeOdd != nil {
				bestOdds = &gameOdds[i]
				break
			}
		}
	}

	// Determine recommended pick
	pick := game.HomeTeam
	confidence := "medium"
	if result.WinProbability.Away > result.WinProbability.Home {
		pick = game.AwayTeam
	}

	if len(result.RecommendedBets) > 0 {
		pick = result.RecommendedBets[0].Pick
		confidence = result.RecommendedBets[0].Confidence
	}

	// Serialize key factors
	keyFactorsJSON, _ := json.Marshal(result.KeyFactors)

	bookmaker := ""
	var homeOdd, awayOdd, ouLine *float64
	if bestOdds != nil {
		bookmaker = bestOdds.Bookmaker
		homeOdd = bestOdds.HomeOdd
		awayOdd = bestOdds.AwayOdd
		ouLine = bestOdds.OverUnderLine
	}

	_, err := s.db.Exec(
		`INSERT INTO ai_predictions (game_id, home_team, away_team, game_date, home_win_prob, away_win_prob, recommended_pick, confidence, bookmaker, home_odd, away_odd, over_under_line, key_factors, summary, analyzed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		game.ID, game.HomeTeam, game.AwayTeam, game.GameDate,
		result.WinProbability.Home, result.WinProbability.Away,
		pick, confidence, bookmaker, homeOdd, awayOdd, ouLine,
		string(keyFactorsJSON), result.Summary, time.Now(),
	)
	return err
}

// GetPredictions returns the latest AI predictions
func (s *Service) GetPredictions() ([]models.AIPrediction, error) {
	rows, err := s.db.Query(
		`SELECT id, game_id, home_team, away_team, game_date, home_win_prob, away_win_prob, recommended_pick, confidence, bookmaker, home_odd, away_odd, over_under_line, key_factors, summary, analyzed_at 
		FROM ai_predictions ORDER BY home_win_prob + away_win_prob DESC, game_date ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []models.AIPrediction
	for rows.Next() {
		var p models.AIPrediction
		if err := rows.Scan(&p.ID, &p.GameID, &p.HomeTeam, &p.AwayTeam, &p.GameDate,
			&p.HomeWinProb, &p.AwayWinProb, &p.RecommendedPick, &p.Confidence,
			&p.Bookmaker, &p.HomeOdd, &p.AwayOdd, &p.OverUnderLine,
			&p.KeyFactors, &p.Summary, &p.AnalyzedAt); err != nil {
			return nil, err
		}
		predictions = append(predictions, p)
	}
	return predictions, nil
}

// StartCronJob runs analysis every intervalHours
func (s *Service) StartCronJob(intervalHours int) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	go func() {
		// Run on startup with a small delay to let scraper finish first
		time.Sleep(30 * time.Second)
		s.RunAnalysis()
		for range ticker.C {
			s.RunAnalysis()
		}
	}()
}

func extractJSON(s string) string {
	// Try to find JSON between ```json and ``` markers
	start := -1
	for i := 0; i < len(s)-6; i++ {
		if s[i:i+7] == "```json" {
			start = i + 7
			break
		}
		if s[i:i+3] == "```" && start == -1 {
			start = i + 3
		}
	}
	if start == -1 {
		// Try to find raw JSON object
		for i := 0; i < len(s); i++ {
			if s[i] == '{' {
				start = i
				break
			}
		}
	}
	if start == -1 {
		return s
	}

	end := len(s)
	for i := len(s) - 1; i >= start; i-- {
		if s[i] == '}' {
			end = i + 1
			break
		}
	}

	return s[start:end]
}
