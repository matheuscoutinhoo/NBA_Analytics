package games

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

func (r *Repository) GetRecentGames(days int) ([]models.NBAGame, error) {
	since := time.Now().AddDate(0, 0, -days)
	now := time.Now()
	rows, err := r.db.Query(
		"SELECT id, external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at FROM nba_games WHERE game_date >= ? AND game_date <= ? ORDER BY game_date DESC",
		since, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent games: %w", err)
	}
	defer rows.Close()
	return scanGames(rows)
}

func (r *Repository) GetUpcomingGames(days int) ([]models.NBAGame, error) {
	now := time.Now()
	until := now.AddDate(0, 0, days)
	rows, err := r.db.Query(
		"SELECT id, external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at FROM nba_games WHERE game_date >= ? AND game_date <= ? AND status = 'scheduled' ORDER BY game_date ASC",
		now, until,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming games: %w", err)
	}
	defer rows.Close()
	return scanGames(rows)
}

func (r *Repository) GetGameByID(id int64) (*models.NBAGame, error) {
	game := &models.NBAGame{}
	err := r.db.QueryRow(
		"SELECT id, external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at FROM nba_games WHERE id = ?",
		id,
	).Scan(&game.ID, &game.ExternalID, &game.GameDate, &game.HomeTeam, &game.AwayTeam, &game.HomeScore, &game.AwayScore, &game.Status, &game.Season, &game.ScrapedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return game, nil
}

func (r *Repository) UpsertGame(game *models.NBAGame) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO nba_games (external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(external_id) DO UPDATE SET
			home_score = excluded.home_score,
			away_score = excluded.away_score,
			status = excluded.status,
			scraped_at = excluded.scraped_at`,
		game.ExternalID, game.GameDate, game.HomeTeam, game.AwayTeam, game.HomeScore, game.AwayScore, game.Status, game.Season, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert game: %w", err)
	}
	return result.LastInsertId()
}

func (r *Repository) GetGameStats(gameID int64) ([]models.GameStats, error) {
	rows, err := r.db.Query(
		"SELECT id, game_id, team, field_goal_pct, three_point_pct, rebounds, assists, turnovers, created_at FROM game_stats WHERE game_id = ?",
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.GameStats
	for rows.Next() {
		var s models.GameStats
		if err := rows.Scan(&s.ID, &s.GameID, &s.Team, &s.FieldGoalPct, &s.ThreePointPct, &s.Rebounds, &s.Assists, &s.Turnovers, &s.CreatedAt); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *Repository) UpsertGameStats(stats *models.GameStats) error {
	_, err := r.db.Exec(
		`INSERT INTO game_stats (game_id, team, field_goal_pct, three_point_pct, rebounds, assists, turnovers)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		stats.GameID, stats.Team, stats.FieldGoalPct, stats.ThreePointPct, stats.Rebounds, stats.Assists, stats.Turnovers,
	)
	return err
}

func (r *Repository) GetPlayerStats(gameID int64) ([]models.PlayerStats, error) {
	rows, err := r.db.Query(
		"SELECT id, game_id, player_name, team, points, rebounds, assists, minutes, created_at FROM player_stats WHERE game_id = ?",
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PlayerStats
	for rows.Next() {
		var s models.PlayerStats
		if err := rows.Scan(&s.ID, &s.GameID, &s.PlayerName, &s.Team, &s.Points, &s.Rebounds, &s.Assists, &s.Minutes, &s.CreatedAt); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *Repository) UpsertPlayerStats(stats *models.PlayerStats) error {
	_, err := r.db.Exec(
		`INSERT INTO player_stats (game_id, player_name, team, points, rebounds, assists, minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		stats.GameID, stats.PlayerName, stats.Team, stats.Points, stats.Rebounds, stats.Assists, stats.Minutes,
	)
	return err
}

func (r *Repository) GetTeamRecentGames(team string, limit int) ([]models.NBAGame, error) {
	rows, err := r.db.Query(
		"SELECT id, external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at FROM nba_games WHERE (home_team = ? OR away_team = ?) AND status = 'final' ORDER BY game_date DESC LIMIT ?",
		team, team, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}

func (r *Repository) GetHeadToHead(team1, team2 string, limit int) ([]models.NBAGame, error) {
	rows, err := r.db.Query(
		`SELECT id, external_id, game_date, home_team, away_team, home_score, away_score, status, season, scraped_at 
		FROM nba_games 
		WHERE ((home_team = ? AND away_team = ?) OR (home_team = ? AND away_team = ?)) AND status = 'final' 
		ORDER BY game_date DESC LIMIT ?`,
		team1, team2, team2, team1, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}

func scanGames(rows *sql.Rows) ([]models.NBAGame, error) {
	var games []models.NBAGame
	for rows.Next() {
		var g models.NBAGame
		if err := rows.Scan(&g.ID, &g.ExternalID, &g.GameDate, &g.HomeTeam, &g.AwayTeam, &g.HomeScore, &g.AwayScore, &g.Status, &g.Season, &g.ScrapedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}
