package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/matheuscoutinhoo/better/internal/bets"
	"github.com/matheuscoutinhoo/better/internal/database"
	"github.com/matheuscoutinhoo/better/internal/models"
	_ "modernc.org/sqlite"
)

func setupBetsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	if err := database.InitWithDB(db); err != nil {
		t.Fatalf("Failed to init test database: %v", err)
	}
	return db
}

func createBetTestUser(t *testing.T, db *sql.DB) int64 {
	result, err := db.Exec("INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)", "bettest@example.com", "hash", "user")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func createTestGame(t *testing.T, db *sql.DB) int64 {
	result, err := db.Exec(
		"INSERT INTO nba_games (external_id, game_date, home_team, away_team, status, season) VALUES (?, ?, ?, ?, ?, ?)",
		"test_game_1", time.Now(), "Lakers", "Celtics", "scheduled", "2025-2026",
	)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestCreateBet(t *testing.T) {
	db := setupBetsTestDB(t)
	defer db.Close()
	userID := createBetTestUser(t, db)
	gameID := createTestGame(t, db)

	repo := bets.NewRepository(db)

	bet := &models.UserBet{
		UserID:          userID,
		GameID:          gameID,
		BetType:         "moneyline",
		Selection:       "Lakers",
		OddValue:        1.85,
		Stake:           100.0,
		PotentialReturn: 185.0,
	}

	id, err := repo.CreateBet(bet)
	if err != nil {
		t.Fatalf("CreateBet failed: %v", err)
	}
	if id <= 0 {
		t.Error("bet ID should be positive")
	}
}

func TestGetUserBets(t *testing.T) {
	db := setupBetsTestDB(t)
	defer db.Close()
	userID := createBetTestUser(t, db)
	gameID := createTestGame(t, db)

	repo := bets.NewRepository(db)

	// Create two bets
	repo.CreateBet(&models.UserBet{UserID: userID, GameID: gameID, BetType: "moneyline", Selection: "Lakers", OddValue: 1.85, Stake: 100, PotentialReturn: 185})
	repo.CreateBet(&models.UserBet{UserID: userID, GameID: gameID, BetType: "over_under", Selection: "Over 220.5", OddValue: 1.90, Stake: 50, PotentialReturn: 95})

	betsList, err := repo.GetUserBets(userID)
	if err != nil {
		t.Fatalf("GetUserBets failed: %v", err)
	}
	if len(betsList) != 2 {
		t.Errorf("expected 2 bets, got %d", len(betsList))
	}
}

func TestUpdateBetResult(t *testing.T) {
	db := setupBetsTestDB(t)
	defer db.Close()
	userID := createBetTestUser(t, db)
	gameID := createTestGame(t, db)

	repo := bets.NewRepository(db)

	id, _ := repo.CreateBet(&models.UserBet{UserID: userID, GameID: gameID, BetType: "moneyline", Selection: "Lakers", OddValue: 1.85, Stake: 100, PotentialReturn: 185})

	err := repo.UpdateBetResult(id, userID, "won")
	if err != nil {
		t.Fatalf("UpdateBetResult failed: %v", err)
	}

	bet, _ := repo.GetBetByID(id, userID)
	if bet.Result != "won" {
		t.Errorf("expected result 'won', got '%s'", bet.Result)
	}
	if bet.SettledAt == nil {
		t.Error("settled_at should be set")
	}
}

func TestDeleteBet(t *testing.T) {
	db := setupBetsTestDB(t)
	defer db.Close()
	userID := createBetTestUser(t, db)
	gameID := createTestGame(t, db)

	repo := bets.NewRepository(db)

	id, _ := repo.CreateBet(&models.UserBet{UserID: userID, GameID: gameID, BetType: "moneyline", Selection: "Lakers", OddValue: 1.85, Stake: 100, PotentialReturn: 185})

	err := repo.DeleteBet(id, userID)
	if err != nil {
		t.Fatalf("DeleteBet failed: %v", err)
	}

	bet, _ := repo.GetBetByID(id, userID)
	if bet != nil {
		t.Error("bet should be deleted")
	}
}

func TestGetBetByID_WrongUser(t *testing.T) {
	db := setupBetsTestDB(t)
	defer db.Close()
	userID := createBetTestUser(t, db)
	gameID := createTestGame(t, db)

	repo := bets.NewRepository(db)

	id, _ := repo.CreateBet(&models.UserBet{UserID: userID, GameID: gameID, BetType: "moneyline", Selection: "Lakers", OddValue: 1.85, Stake: 100, PotentialReturn: 185})

	bet, _ := repo.GetBetByID(id, 9999) // wrong user
	if bet != nil {
		t.Error("should not find bet for wrong user")
	}
}
