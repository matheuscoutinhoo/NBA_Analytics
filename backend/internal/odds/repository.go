package odds

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetOddsByGameID(gameID int64) ([]models.GameOdds, error) {
	rows, err := r.db.Query(
		"SELECT id, game_id, bookmaker, market_type, home_odd, away_odd, draw_odd, over_under_line, over_odd, under_odd, fetched_at FROM game_odds WHERE game_id = ? ORDER BY fetched_at DESC",
		gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get odds: %w", err)
	}
	defer rows.Close()

	var oddsList []models.GameOdds
	for rows.Next() {
		var o models.GameOdds
		if err := rows.Scan(&o.ID, &o.GameID, &o.Bookmaker, &o.MarketType, &o.HomeOdd, &o.AwayOdd, &o.DrawOdd, &o.OverUnderLine, &o.OverOdd, &o.UnderOdd, &o.FetchedAt); err != nil {
			return nil, err
		}
		oddsList = append(oddsList, o)
	}
	return oddsList, nil
}

func (r *Repository) UpsertOdds(o *models.GameOdds) error {
	_, err := r.db.Exec(
		`INSERT INTO game_odds (game_id, bookmaker, market_type, home_odd, away_odd, draw_odd, over_under_line, over_odd, under_odd, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.GameID, o.Bookmaker, o.MarketType, o.HomeOdd, o.AwayOdd, o.DrawOdd, o.OverUnderLine, o.OverOdd, o.UnderOdd, time.Now(),
	)
	return err
}
