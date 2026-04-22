-- 001_create_tables.sql
-- Users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('user', 'admin')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX idx_users_email ON users(email);

-- Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    revoked INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Login Attempts (brute force protection)
CREATE TABLE IF NOT EXISTS login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    attempted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_login_attempts_email ON login_attempts(email);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address);

-- NBA Games
CREATE TABLE IF NOT EXISTS nba_games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE,
    game_date DATETIME NOT NULL,
    home_team TEXT NOT NULL,
    away_team TEXT NOT NULL,
    home_score INTEGER,
    away_score INTEGER,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK(status IN ('scheduled', 'live', 'final', 'postponed')),
    season TEXT NOT NULL,
    scraped_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_nba_games_date ON nba_games(game_date);
CREATE INDEX idx_nba_games_teams ON nba_games(home_team, away_team);

-- Game Stats
CREATE TABLE IF NOT EXISTS game_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    team TEXT NOT NULL,
    field_goal_pct REAL,
    three_point_pct REAL,
    rebounds INTEGER,
    assists INTEGER,
    turnovers INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES nba_games(id) ON DELETE CASCADE
);

CREATE INDEX idx_game_stats_game_id ON game_stats(game_id);

-- Player Stats
CREATE TABLE IF NOT EXISTS player_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    player_name TEXT NOT NULL,
    team TEXT NOT NULL,
    points INTEGER,
    rebounds INTEGER,
    assists INTEGER,
    minutes INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES nba_games(id) ON DELETE CASCADE
);

CREATE INDEX idx_player_stats_game_id ON player_stats(game_id);
CREATE INDEX idx_player_stats_player ON player_stats(player_name);

-- AI Insights
CREATE TABLE IF NOT EXISTS ai_insights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    prompt_used TEXT NOT NULL,
    insight_text TEXT NOT NULL,
    confidence_score REAL,
    generated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES nba_games(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_ai_insights_game_id ON ai_insights(game_id);
CREATE INDEX idx_ai_insights_user_id ON ai_insights(user_id);

-- Game Odds
CREATE TABLE IF NOT EXISTS game_odds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    bookmaker TEXT NOT NULL,
    market_type TEXT NOT NULL,
    home_odd REAL,
    away_odd REAL,
    draw_odd REAL,
    over_under_line REAL,
    over_odd REAL,
    under_odd REAL,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES nba_games(id) ON DELETE CASCADE
);

CREATE INDEX idx_game_odds_game_id ON game_odds(game_id);

-- User Bets
CREATE TABLE IF NOT EXISTS user_bets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    game_id INTEGER NOT NULL,
    bet_type TEXT NOT NULL,
    selection TEXT NOT NULL,
    odd_value REAL NOT NULL,
    stake REAL NOT NULL,
    potential_return REAL NOT NULL,
    result TEXT CHECK(result IN ('pending', 'won', 'lost', 'void')),
    placed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    settled_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (game_id) REFERENCES nba_games(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_bets_user_id ON user_bets(user_id);
CREATE INDEX idx_user_bets_game_id ON user_bets(game_id);

-- Bankroll
CREATE TABLE IF NOT EXISTS bankroll (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    initial_amount REAL NOT NULL DEFAULT 0,
    current_amount REAL NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_bankroll_user_id ON bankroll(user_id);

-- Bankroll Transactions
CREATE TABLE IF NOT EXISTS bankroll_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    bet_id INTEGER,
    type TEXT NOT NULL CHECK(type IN ('deposit', 'withdraw', 'bet_placed', 'bet_won', 'bet_lost', 'bet_void')),
    amount REAL NOT NULL,
    balance_after REAL NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (bet_id) REFERENCES user_bets(id) ON DELETE SET NULL
);

CREATE INDEX idx_bankroll_tx_user_id ON bankroll_transactions(user_id);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    entity TEXT NOT NULL,
    entity_id INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
