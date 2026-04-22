package models

import "time"

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginAttempt struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	IPAddress   string    `json:"ip_address"`
	AttemptedAt time.Time `json:"attempted_at"`
	Success     bool      `json:"success"`
}

type NBAGame struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"external_id,omitempty"`
	GameDate   time.Time `json:"game_date"`
	HomeTeam   string    `json:"home_team"`
	AwayTeam   string    `json:"away_team"`
	HomeScore  *int      `json:"home_score,omitempty"`
	AwayScore  *int      `json:"away_score,omitempty"`
	Status     string    `json:"status"`
	Season     string    `json:"season"`
	ScrapedAt  time.Time `json:"scraped_at"`
}

type GameStats struct {
	ID             int64     `json:"id"`
	GameID         int64     `json:"game_id"`
	Team           string    `json:"team"`
	FieldGoalPct   *float64  `json:"field_goal_pct,omitempty"`
	ThreePointPct  *float64  `json:"three_point_pct,omitempty"`
	Rebounds       *int      `json:"rebounds,omitempty"`
	Assists        *int      `json:"assists,omitempty"`
	Turnovers      *int      `json:"turnovers,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PlayerStats struct {
	ID         int64     `json:"id"`
	GameID     int64     `json:"game_id"`
	PlayerName string    `json:"player_name"`
	Team       string    `json:"team"`
	Points     *int      `json:"points,omitempty"`
	Rebounds   *int      `json:"rebounds,omitempty"`
	Assists    *int      `json:"assists,omitempty"`
	Minutes    *int      `json:"minutes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AIInsight struct {
	ID              int64     `json:"id"`
	GameID          int64     `json:"game_id"`
	UserID          int64     `json:"user_id"`
	PromptUsed      string    `json:"prompt_used,omitempty"`
	InsightText     string    `json:"insight_text"`
	ConfidenceScore *float64  `json:"confidence_score,omitempty"`
	GeneratedAt     time.Time `json:"generated_at"`
}

type GameOdds struct {
	ID            int64     `json:"id"`
	GameID        int64     `json:"game_id"`
	Bookmaker     string    `json:"bookmaker"`
	MarketType    string    `json:"market_type"`
	HomeOdd       *float64  `json:"home_odd,omitempty"`
	AwayOdd       *float64  `json:"away_odd,omitempty"`
	DrawOdd       *float64  `json:"draw_odd,omitempty"`
	OverUnderLine *float64  `json:"over_under_line,omitempty"`
	OverOdd       *float64  `json:"over_odd,omitempty"`
	UnderOdd      *float64  `json:"under_odd,omitempty"`
	FetchedAt     time.Time `json:"fetched_at"`
}

type UserBet struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	GameID          int64      `json:"game_id"`
	BetType         string     `json:"bet_type"`
	Selection       string     `json:"selection"`
	OddValue        float64    `json:"odd_value"`
	Stake           float64    `json:"stake"`
	PotentialReturn float64    `json:"potential_return"`
	Result          string     `json:"result"`
	PlacedAt        time.Time  `json:"placed_at"`
	SettledAt       *time.Time `json:"settled_at,omitempty"`
}

type Bankroll struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	InitialAmount float64   `json:"initial_amount"`
	CurrentAmount float64   `json:"current_amount"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BankrollTransaction struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	BetID        *int64    `json:"bet_id,omitempty"`
	Type         string    `json:"type"`
	Amount       float64   `json:"amount"`
	BalanceAfter float64   `json:"balance_after"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	UserID    *int64    `json:"user_id,omitempty"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  *int64    `json:"entity_id,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
