package bankroll

import (
	"database/sql"
	"fmt"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetBankroll(userID int64) (*models.Bankroll, error) {
	b := &models.Bankroll{}
	err := r.db.QueryRow(
		"SELECT id, user_id, initial_amount, current_amount, updated_at FROM bankroll WHERE user_id = ?",
		userID,
	).Scan(&b.ID, &b.UserID, &b.InitialAmount, &b.CurrentAmount, &b.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (r *Repository) CreateBankroll(userID int64, initialAmount float64) (*models.Bankroll, error) {
	_, err := r.db.Exec(
		"INSERT INTO bankroll (user_id, initial_amount, current_amount) VALUES (?, ?, ?)",
		userID, initialAmount, initialAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bankroll: %w", err)
	}
	return r.GetBankroll(userID)
}

func (r *Repository) UpdateBalance(userID int64, amount float64) error {
	_, err := r.db.Exec(
		"UPDATE bankroll SET current_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?",
		amount, userID,
	)
	return err
}

func (r *Repository) GetTransactions(userID int64) ([]models.BankrollTransaction, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, bet_id, type, amount, balance_after, created_at FROM bankroll_transactions WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.BankrollTransaction
	for rows.Next() {
		var tx models.BankrollTransaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.BetID, &tx.Type, &tx.Amount, &tx.BalanceAfter, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (r *Repository) CreateTransaction(tx *models.BankrollTransaction) error {
	_, err := r.db.Exec(
		"INSERT INTO bankroll_transactions (user_id, bet_id, type, amount, balance_after) VALUES (?, ?, ?, ?, ?)",
		tx.UserID, tx.BetID, tx.Type, tx.Amount, tx.BalanceAfter,
	)
	return err
}
