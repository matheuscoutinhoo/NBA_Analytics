export interface User {
   id: number;
   email: string;
   role: string;
   created_at: string;
   updated_at: string;
}

export interface NBAGame {
   id: number;
   external_id: string;
   game_date: string;
   home_team: string;
   away_team: string;
   home_score: number | null;
   away_score: number | null;
   status: "scheduled" | "live" | "final" | "postponed";
   season: string;
   scraped_at: string;
}

export interface GameStats {
   id: number;
   game_id: number;
   team: string;
   field_goal_pct: number | null;
   three_point_pct: number | null;
   rebounds: number | null;
   assists: number | null;
   turnovers: number | null;
}

export interface PlayerStats {
   id: number;
   game_id: number;
   player_name: string;
   team: string;
   points: number | null;
   rebounds: number | null;
   assists: number | null;
   minutes: number | null;
}

export interface GameOdds {
   id: number;
   game_id: number;
   bookmaker: string;
   market_type: string;
   home_odd: number | null;
   away_odd: number | null;
   draw_odd: number | null;
   over_under_line: number | null;
   over_odd: number | null;
   under_odd: number | null;
   fetched_at: string;
}

export interface AIInsight {
   id: number;
   game_id: number;
   user_id: number;
   insight_text: string;
   confidence_score: number | null;
   generated_at: string;
}

export interface AIPrediction {
   id: number;
   game_id: number;
   home_team: string;
   away_team: string;
   game_date: string;
   home_win_prob: number;
   away_win_prob: number;
   recommended_pick: string;
   confidence: string;
   bookmaker: string;
   home_odd: number | null;
   away_odd: number | null;
   over_under_line: number | null;
   key_factors: string;
   summary: string;
   analyzed_at: string;
}

export interface UserBet {
   id: number;
   user_id: number;
   game_id: number;
   bet_type: string;
   selection: string;
   odd_value: number;
   stake: number;
   potential_return: number;
   result: "pending" | "won" | "lost" | "void";
   placed_at: string;
   settled_at: string | null;
}

export interface Bankroll {
   id: number;
   user_id: number;
   initial_amount: number;
   current_amount: number;
   updated_at: string;
}

export interface BankrollTransaction {
   id: number;
   user_id: number;
   bet_id: number | null;
   type: string;
   amount: number;
   balance_after: number;
   created_at: string;
}

export interface BankrollStats {
   total_bets: number;
   won_bets: number;
   lost_bets: number;
   pending_bets: number;
   win_rate: number;
   total_staked: number;
   total_returns: number;
   profit_loss: number;
   roi: number;
   initial_amount: number;
   current_amount: number;
}

export interface LoginResponse {
   access_token: string;
   user: User;
}

export interface AuthState {
   user: User | null;
   accessToken: string | null;
   isLoading: boolean;
}
