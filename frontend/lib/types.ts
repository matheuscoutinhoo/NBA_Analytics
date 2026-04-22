export interface User {
   id: string;
   username: string;
   email: string;
   role: string;
   is_active: boolean;
   age_verified: boolean;
   consent_given: boolean;
   xp_points: number;
   created_at: string;
   updated_at: string;
}

export interface TokenPair {
   access_token: string;
   refresh_token: string;
   expires_at: number;
}

export interface AuthResponse {
   user: User;
   tokens: TokenPair;
}

export interface Team {
   id: string;
   name: string;
   abbreviation: string;
   city: string;
   conference: string;
   logo_url?: string;
}

export interface Odds {
   id: string;
   game_id: string;
   source: string;
   home_moneyline: number;
   away_moneyline: number;
   home_spread: number;
   away_spread: number;
   over_under: number;
   home_implied_prob: number;
   away_implied_prob: number;
   fetched_at: string;
}

export interface Analysis {
   id: string;
   game_id: string;
   home_win_prob: number;
   away_win_prob: number;
   confidence: "LOW" | "MEDIUM" | "HIGH";
   justification: string;
   risk_alerts: string[];
   home_ev?: number;
   away_ev?: number;
   is_home_value: boolean;
   is_away_value: boolean;
   model_version: string;
   analyzed_at: string;
}

export interface Injury {
   player_name: string;
   status: string;
   description: string;
}

export interface Game {
   id: string;
   external_id: string;
   home_team_id: string;
   away_team_id: string;
   game_date: string;
   season: string;
   status: "SCHEDULED" | "LIVE" | "FINAL" | "POSTPONED";
   home_score?: number;
   away_score?: number;
   home_team?: Team;
   away_team?: Team;
   analysis?: Analysis;
   latest_odds?: Odds;
   home_injuries?: Injury[];
   away_injuries?: Injury[];
}

export interface ValueAnalysis {
   model_prob: number;
   implied_prob: number;
   ev: number;
   is_value: boolean;
   divergence_pct: number;
}

export interface GameAnalysisResponse {
   game: Game;
   home_value?: ValueAnalysis;
   away_value?: ValueAnalysis;
   disclaimer: string;
   responsible_gaming: string;
}

export interface BankrollEntry {
   id: string;
   user_id: string;
   game_id?: string;
   entry_type: "DEPOSIT" | "WITHDRAWAL" | "BET" | "WIN" | "LOSS";
   bet_type?: string;
   odd?: number;
   stake: number;
   result?: number;
   balance_after: number;
   notes?: string;
   created_at: string;
}

export interface DashboardStats {
   current_balance: number;
   total_deposits: number;
   total_withdrawals: number;
   total_bets: number;
   win_count: number;
   loss_count: number;
   win_rate: number;
   roi: number;
   max_drawdown: number;
   profit_loss: number;
   balance_history: { date: string; balance: number }[];
   monthly_stats: { month: string; profit: number; bets: number; win_rate: number }[];
}

export interface UserPreferences {
   id: string;
   user_id: string;
   daily_limit?: number;
   weekly_limit?: number;
   monthly_limit?: number;
   alert_threshold: number;
   dark_mode: boolean;
   notifications_enabled: boolean;
}

export interface APIResponse<T> {
   success: boolean;
   data?: T;
   error?: { code: string; message: string };
   meta?: { page: number; per_page: number; total: number; total_pages: number };
}
