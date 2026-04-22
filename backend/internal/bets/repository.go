package bets

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

func (r *Repository) GetUserBets(userID int64) ([]models.UserBet, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, game_id, bet_type, selection, odd_value, stake, potential_return, result, placed_at, settled_at FROM user_bets WHERE user_id = ? ORDER BY placed_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get bets: %w", err)
	}
	defer rows.Close()

	var betsList []models.UserBet
	for rows.Next() {
		var b models.UserBet
		if err := rows.Scan(&b.ID, &b.UserID, &b.GameID, &b.BetType, &b.Selection, &b.OddValue, &b.Stake, &b.PotentialReturn, &b.Result, &b.PlacedAt, &b.SettledAt); err != nil {
			return nil, err
		}
		betsList = append(betsList, b)
	}
	return betsList, nil
}

func (r *Repository) GetBetByID(id, userID int64) (*models.UserBet, error) {
	bet := &models.UserBet{}
	err := r.db.QueryRow(
		"SELECT id, user_id, game_id, bet_type, selection, odd_value, stake, potential_return, result, placed_at, settled_at FROM user_bets WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&bet.ID, &bet.UserID, &bet.GameID, &bet.BetType, &bet.Selection, &bet.OddValue, &bet.Stake, &bet.PotentialReturn, &bet.Result, &bet.PlacedAt, &bet.SettledAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return bet, nil
}

func (r *Repository) CreateBet(bet *models.UserBet) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO user_bets (user_id, game_id, bet_type, selection, odd_value, stake, potential_return, result) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')",
		bet.UserID, bet.GameID, bet.BetType, bet.Selection, bet.OddValue, bet.Stake, bet.PotentialReturn,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create bet: %w", err)
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateBetResult(id, userID int64, result string) error {
	_, err := r.db.Exec(
		"UPDATE user_bets SET result = ?, settled_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		result, id, userID,
	)
	return err
}

func (r *Repository) DeleteBet(id, userID int64) error {
	_, err := r.db.Exec("DELETE FROM user_bets WHERE id = ? AND user_id = ?", id, userID)
	return err
}
