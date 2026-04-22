package games

import (
	"database/sql"
	"fmt"

	"github.com/matheuscoutinhoo/better/internal/abacus"
	"github.com/matheuscoutinhoo/better/internal/models"
	"github.com/matheuscoutinhoo/better/internal/odds"
)

type InsightService struct {
	gamesRepo *Repository
	oddsRepo  *odds.Repository
	aiClient  *abacus.Client
	db        *sql.DB
}

func NewInsightService(gamesRepo *Repository, oddsRepo *odds.Repository, aiClient *abacus.Client, db *sql.DB) *InsightService {
	return &InsightService{
		gamesRepo: gamesRepo,
		oddsRepo:  oddsRepo,
		aiClient:  aiClient,
		db:        db,
	}
}

func (s *InsightService) GenerateInsight(gameID, userID int64) (*models.AIInsight, error) {
	game, err := s.gamesRepo.GetGameByID(gameID)
	if err != nil || game == nil {
		return nil, fmt.Errorf("game not found")
	}

	// Get context data for the prompt
	homeRecent, _ := s.gamesRepo.GetTeamRecentGames(game.HomeTeam, 10)
	awayRecent, _ := s.gamesRepo.GetTeamRecentGames(game.AwayTeam, 10)
	h2h, _ := s.gamesRepo.GetHeadToHead(game.HomeTeam, game.AwayTeam, 6)
	gameOdds, _ := s.oddsRepo.GetOddsByGameID(gameID)

	prompt := s.aiClient.BuildGameAnalysisPrompt(game, homeRecent, awayRecent, h2h, gameOdds)

	response, err := s.aiClient.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	// Store insight
	result, err := s.db.Exec(
		"INSERT INTO ai_insights (game_id, user_id, prompt_used, insight_text, confidence_score) VALUES (?, ?, ?, ?, ?)",
		gameID, userID, prompt, response, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store insight: %w", err)
	}

	id, _ := result.LastInsertId()
	insight := &models.AIInsight{
		ID:          id,
		GameID:      gameID,
		UserID:      userID,
		PromptUsed:  prompt,
		InsightText: response,
	}

	return insight, nil
}
