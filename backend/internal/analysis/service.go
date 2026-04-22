package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/better/backend/pkg/logger"
	"github.com/better/backend/pkg/response"
)

type Game struct {
	ID           uuid.UUID `json:"id"`
	ExternalID   string    `json:"external_id"`
	HomeTeamID   uuid.UUID `json:"home_team_id"`
	AwayTeamID   uuid.UUID `json:"away_team_id"`
	GameDate     time.Time `json:"game_date"`
	Season       string    `json:"season"`
	Status       string    `json:"status"`
	HomeScore    *int      `json:"home_score"`
	AwayScore    *int      `json:"away_score"`
	HomeTeam     *Team     `json:"home_team,omitempty"`
	AwayTeam     *Team     `json:"away_team,omitempty"`
	Analysis     *Analysis `json:"analysis,omitempty"`
	LatestOdds   *Odds     `json:"latest_odds,omitempty"`
	HomeInjuries []Injury  `json:"home_injuries,omitempty"`
	AwayInjuries []Injury  `json:"away_injuries,omitempty"`
}

type Team struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Abbreviation string    `json:"abbreviation"`
	City         string    `json:"city"`
	Conference   string    `json:"conference"`
	LogoURL      *string   `json:"logo_url"`
}

type Analysis struct {
	ID            uuid.UUID       `json:"id"`
	GameID        uuid.UUID       `json:"game_id"`
	HomeWinProb   float64         `json:"home_win_prob"`
	AwayWinProb   float64         `json:"away_win_prob"`
	Confidence    string          `json:"confidence"`
	Justification string          `json:"justification"`
	RiskAlerts    json.RawMessage `json:"risk_alerts"`
	HomeEV        *float64        `json:"home_ev"`
	AwayEV        *float64        `json:"away_ev"`
	IsHomeValue   bool            `json:"is_home_value"`
	IsAwayValue   bool            `json:"is_away_value"`
	ModelVersion  string          `json:"model_version"`
	AnalyzedAt    time.Time       `json:"analyzed_at"`
}

type Odds struct {
	ID              uuid.UUID `json:"id"`
	GameID          uuid.UUID `json:"game_id"`
	Source          string    `json:"source"`
	HomeMoneyline   float64   `json:"home_moneyline"`
	AwayMoneyline   float64   `json:"away_moneyline"`
	HomeSpread      float64   `json:"home_spread"`
	AwaySpread      float64   `json:"away_spread"`
	OverUnder       float64   `json:"over_under"`
	HomeImpliedProb float64   `json:"home_implied_prob"`
	AwayImpliedProb float64   `json:"away_implied_prob"`
	FetchedAt       time.Time `json:"fetched_at"`
}

type Injury struct {
	PlayerName  string `json:"player_name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type ValueAnalysis struct {
	ModelProb   float64 `json:"model_prob"`
	ImpliedProb float64 `json:"implied_prob"`
	EV          float64 `json:"ev"`
	IsValue     bool    `json:"is_value"`
	Divergence  float64 `json:"divergence_pct"`
}

type Service struct {
	db           *pgxpool.Pool
	redis        *redis.Client
	logger       *logger.Logger
	abacusAPIKey string
	abacusAPIURL string
}

func NewService(db *pgxpool.Pool, redisClient *redis.Client, log *logger.Logger, abacusKey, abacusURL string) *Service {
	return &Service{
		db:           db,
		redis:        redisClient,
		logger:       log,
		abacusAPIKey: abacusKey,
		abacusAPIURL: abacusURL,
	}
}

func (s *Service) GetTodaysGames(ctx context.Context) ([]Game, error) {
	// Check Redis cache first
	cacheKey := fmt.Sprintf("games:today:%s", time.Now().UTC().Format("2006-01-02"))
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var games []Game
		if json.Unmarshal([]byte(cached), &games) == nil {
			return games, nil
		}
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.external_id, g.home_team_id, g.away_team_id, g.game_date, g.season, g.status, g.home_score, g.away_score,
			ht.id, ht.name, ht.abbreviation, ht.city, ht.conference, ht.logo_url,
			at.id, at.name, at.abbreviation, at.city, at.conference, at.logo_url
		FROM games g
		JOIN teams ht ON g.home_team_id = ht.id
		JOIN teams at ON g.away_team_id = at.id
		WHERE g.game_date >= $1 AND g.game_date < $2
		ORDER BY g.game_date ASC`, today, tomorrow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var ht, at Team
		err := rows.Scan(
			&g.ID, &g.ExternalID, &g.HomeTeamID, &g.AwayTeamID, &g.GameDate, &g.Season, &g.Status, &g.HomeScore, &g.AwayScore,
			&ht.ID, &ht.Name, &ht.Abbreviation, &ht.City, &ht.Conference, &ht.LogoURL,
			&at.ID, &at.Name, &at.Abbreviation, &at.City, &at.Conference, &at.LogoURL,
		)
		if err != nil {
			s.logger.Error("scanning game row", map[string]interface{}{"error": err.Error()})
			continue
		}
		g.HomeTeam = &ht
		g.AwayTeam = &at

		// Get latest odds
		odds, _ := s.getLatestOdds(ctx, g.ID)
		g.LatestOdds = odds

		// Get analysis
		analysis, _ := s.getLatestAnalysis(ctx, g.ID)
		g.Analysis = analysis

		games = append(games, g)
	}

	// Cache for 5 minutes
	if data, err := json.Marshal(games); err == nil {
		s.redis.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return games, nil
}

func (s *Service) GetGameAnalysis(ctx context.Context, gameID uuid.UUID) (*Game, error) {
	cacheKey := fmt.Sprintf("analysis:%s", gameID)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var game Game
		if json.Unmarshal([]byte(cached), &game) == nil {
			return &game, nil
		}
	}

	var g Game
	var ht, at Team
	err = s.db.QueryRow(ctx, `
		SELECT g.id, g.external_id, g.home_team_id, g.away_team_id, g.game_date, g.season, g.status, g.home_score, g.away_score,
			ht.id, ht.name, ht.abbreviation, ht.city, ht.conference, ht.logo_url,
			at.id, at.name, at.abbreviation, at.city, at.conference, at.logo_url
		FROM games g
		JOIN teams ht ON g.home_team_id = ht.id
		JOIN teams at ON g.away_team_id = at.id
		WHERE g.id = $1`, gameID).Scan(
		&g.ID, &g.ExternalID, &g.HomeTeamID, &g.AwayTeamID, &g.GameDate, &g.Season, &g.Status, &g.HomeScore, &g.AwayScore,
		&ht.ID, &ht.Name, &ht.Abbreviation, &ht.City, &ht.Conference, &ht.LogoURL,
		&at.ID, &at.Name, &at.Abbreviation, &at.City, &at.Conference, &at.LogoURL,
	)
	if err != nil {
		return nil, err
	}

	g.HomeTeam = &ht
	g.AwayTeam = &at

	g.LatestOdds, _ = s.getLatestOdds(ctx, g.ID)
	g.Analysis, _ = s.getLatestAnalysis(ctx, g.ID)

	// Get injuries
	g.HomeInjuries, _ = s.getTeamInjuries(ctx, g.HomeTeamID)
	g.AwayInjuries, _ = s.getTeamInjuries(ctx, g.AwayTeamID)

	// Cache for 10 minutes
	if data, err := json.Marshal(g); err == nil {
		s.redis.Set(ctx, cacheKey, string(data), 10*time.Minute)
	}

	return &g, nil
}

// CalculateEV: EV = (P_model * odd) - 1
func CalculateEV(modelProb, odd float64) float64 {
	if odd <= 0 {
		return 0
	}
	return (modelProb * odd) - 1
}

// CalculateImpliedProbability: P = 1 / odd
func CalculateImpliedProbability(odd float64) float64 {
	if odd <= 0 {
		return 0
	}
	return 1.0 / odd
}

// AnalyzeValue compares model probability with market odds
func AnalyzeValue(modelProb, odd float64) ValueAnalysis {
	impliedProb := CalculateImpliedProbability(odd)
	ev := CalculateEV(modelProb, odd)
	divergence := math.Abs(modelProb-impliedProb) / impliedProb * 100

	return ValueAnalysis{
		ModelProb:   modelProb,
		ImpliedProb: impliedProb,
		EV:          ev,
		IsValue:     ev > 0,
		Divergence:  math.Round(divergence*100) / 100,
	}
}

func (s *Service) getLatestOdds(ctx context.Context, gameID uuid.UUID) (*Odds, error) {
	var odds Odds
	err := s.db.QueryRow(ctx,
		`SELECT id, game_id, source, home_moneyline, away_moneyline, home_spread, away_spread, over_under, home_implied_prob, away_implied_prob, fetched_at
		FROM odds WHERE game_id = $1 ORDER BY fetched_at DESC LIMIT 1`, gameID,
	).Scan(&odds.ID, &odds.GameID, &odds.Source, &odds.HomeMoneyline, &odds.AwayMoneyline,
		&odds.HomeSpread, &odds.AwaySpread, &odds.OverUnder, &odds.HomeImpliedProb, &odds.AwayImpliedProb, &odds.FetchedAt)
	if err != nil {
		return nil, err
	}
	return &odds, nil
}

func (s *Service) getLatestAnalysis(ctx context.Context, gameID uuid.UUID) (*Analysis, error) {
	var a Analysis
	err := s.db.QueryRow(ctx,
		`SELECT id, game_id, home_win_prob, away_win_prob, confidence, justification, risk_alerts, home_ev, away_ev, is_home_value, is_away_value, model_version, analyzed_at
		FROM analyses WHERE game_id = $1 AND expires_at > NOW() ORDER BY analyzed_at DESC LIMIT 1`, gameID,
	).Scan(&a.ID, &a.GameID, &a.HomeWinProb, &a.AwayWinProb, &a.Confidence, &a.Justification,
		&a.RiskAlerts, &a.HomeEV, &a.AwayEV, &a.IsHomeValue, &a.IsAwayValue, &a.ModelVersion, &a.AnalyzedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Service) getTeamInjuries(ctx context.Context, teamID uuid.UUID) ([]Injury, error) {
	rows, err := s.db.Query(ctx,
		`SELECT p.name, i.status, i.description FROM injuries i JOIN players p ON i.player_id = p.id WHERE i.team_id = $1`,
		teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var injuries []Injury
	for rows.Next() {
		var inj Injury
		if err := rows.Scan(&inj.PlayerName, &inj.Status, &inj.Description); err != nil {
			continue
		}
		injuries = append(injuries, inj)
	}
	return injuries, nil
}

// Handler

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetTodaysGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.service.GetTodaysGames(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"games":      games,
		"date":       time.Now().UTC().Format("2006-01-02"),
		"disclaimer": "This analysis is for informational purposes only. Past performance does not guarantee future results. No financial advice is provided.",
	})
}

func (h *Handler) GetGameAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := uuid.Parse(vars["gameId"])
	if err != nil {
		response.BadRequest(w, "invalid game ID")
		return
	}

	game, err := h.service.GetGameAnalysis(r.Context(), gameID)
	if err != nil {
		response.NotFound(w, "game not found")
		return
	}

	// Calculate value analysis if odds and analysis available
	var homeValue, awayValue *ValueAnalysis
	if game.Analysis != nil && game.LatestOdds != nil {
		hv := AnalyzeValue(game.Analysis.HomeWinProb, game.LatestOdds.HomeMoneyline)
		av := AnalyzeValue(game.Analysis.AwayWinProb, game.LatestOdds.AwayMoneyline)
		homeValue = &hv
		awayValue = &av
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"game":               game,
		"home_value":         homeValue,
		"away_value":         awayValue,
		"disclaimer":         "This analysis is for informational purposes only. Past performance does not guarantee future results.",
		"responsible_gaming": "Please set personal limits and gamble responsibly. If you need help, visit www.ncpgambling.org",
	})
}

func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.db.Query(r.Context(),
		`SELECT id, name, abbreviation, city, conference, division, logo_url FROM teams ORDER BY name`)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var t Team
		var division string
		if err := rows.Scan(&t.ID, &t.Name, &t.Abbreviation, &t.City, &t.Conference, &division, &t.LogoURL); err != nil {
			continue
		}
		teams = append(teams, t)
	}

	response.JSON(w, http.StatusOK, teams)
}
