package unit

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/matheuscoutinhoo/better/internal/bankroll"
	"github.com/matheuscoutinhoo/better/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	if err := database.InitWithDB(db); err != nil {
		t.Fatalf("Failed to init test database: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *sql.DB) int64 {
	result, err := db.Exec("INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)", "test@example.com", "hash", "user")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func createTestGameAndBet(t *testing.T, db *sql.DB, userID int64) int64 {
	result, err := db.Exec("INSERT INTO nba_games (external_id, game_date, home_team, away_team, status, season) VALUES (?, datetime('now'), ?, ?, ?, ?)",
		"test_br_game", "Lakers", "Celtics", "scheduled", "2025-2026")
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}
	gameID, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO user_bets (user_id, game_id, bet_type, selection, odd_value, stake, potential_return, result) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, gameID, "moneyline", "Lakers", 1.85, 100.0, 185.0, "pending")
	if err != nil {
		t.Fatalf("Failed to create test bet: %v", err)
	}
	betID, _ := result.LastInsertId()
	return betID
}

func TestBankrollDeposit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	// First deposit creates bankroll
	b, err := svc.Deposit(userID, 1000.0)
	if err != nil {
		t.Fatalf("Deposit failed: %v", err)
	}
	if b.CurrentAmount != 1000.0 {
		t.Errorf("expected 1000.0, got %f", b.CurrentAmount)
	}
	if b.InitialAmount != 1000.0 {
		t.Errorf("expected initial 1000.0, got %f", b.InitialAmount)
	}

	// Second deposit adds to existing
	b, err = svc.Deposit(userID, 500.0)
	if err != nil {
		t.Fatalf("Second deposit failed: %v", err)
	}
	if b.CurrentAmount != 1500.0 {
		t.Errorf("expected 1500.0, got %f", b.CurrentAmount)
	}
}

func TestBankrollWithdraw(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	// Deposit first
	svc.Deposit(userID, 1000.0)

	// Valid withdraw
	b, err := svc.Withdraw(userID, 300.0)
	if err != nil {
		t.Fatalf("Withdraw failed: %v", err)
	}
	if b.CurrentAmount != 700.0 {
		t.Errorf("expected 700.0, got %f", b.CurrentAmount)
	}
}

func TestBankrollWithdraw_InsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	svc.Deposit(userID, 100.0)

	_, err := svc.Withdraw(userID, 200.0)
	if err == nil {
		t.Fatal("should fail with insufficient funds")
	}
}

func TestBankrollWithdraw_NoBankroll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	_, err := svc.Withdraw(userID, 100.0)
	if err == nil {
		t.Fatal("should fail with no bankroll")
	}
}

func TestBankrollDeductBet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)
	betID := createTestGameAndBet(t, db, userID)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	svc.Deposit(userID, 1000.0)

	err := svc.DeductBet(userID, betID, 100.0)
	if err != nil {
		t.Fatalf("DeductBet failed: %v", err)
	}

	b, _ := repo.GetBankroll(userID)
	if b.CurrentAmount != 900.0 {
		t.Errorf("expected 900.0, got %f", b.CurrentAmount)
	}
}

func TestBankrollCreditWin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)
	betID := createTestGameAndBet(t, db, userID)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	svc.Deposit(userID, 1000.0)
	svc.DeductBet(userID, betID, 100.0) // 900

	err := svc.CreditWin(userID, betID, 250.0)
	if err != nil {
		t.Fatalf("CreditWin failed: %v", err)
	}

	b, _ := repo.GetBankroll(userID)
	if b.CurrentAmount != 1150.0 {
		t.Errorf("expected 1150.0, got %f", b.CurrentAmount)
	}
}

func TestBankrollRefundBet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := createTestUser(t, db)
	betID := createTestGameAndBet(t, db, userID)

	repo := bankroll.NewRepository(db)
	svc := bankroll.NewService(repo)

	svc.Deposit(userID, 1000.0)
	svc.DeductBet(userID, betID, 100.0) // 900

	err := svc.RefundBet(userID, betID, 100.0)
	if err != nil {
		t.Fatalf("RefundBet failed: %v", err)
	}

	b, _ := repo.GetBankroll(userID)
	if b.CurrentAmount != 1000.0 {
		t.Errorf("expected 1000.0, got %f", b.CurrentAmount)
	}
}
