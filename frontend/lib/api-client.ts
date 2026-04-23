const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class ApiClient {
   private accessToken: string | null = null;
   private onUnauthorized: (() => void) | null = null;

   setAccessToken(token: string | null) {
      this.accessToken = token;
   }

   setOnUnauthorized(callback: () => void) {
      this.onUnauthorized = callback;
   }

   private async request<T>(
      path: string,
      options: RequestInit = {}
   ): Promise<T> {
      const headers: Record<string, string> = {
         "Content-Type": "application/json",
         ...((options.headers as Record<string, string>) || {}),
      };

      if (this.accessToken) {
         headers["Authorization"] = `Bearer ${this.accessToken}`;
      }

      const response = await fetch(`${API_URL}${path}`, {
         ...options,
         headers,
         credentials: "include",
      });

      if (response.status === 401) {
         // Try to refresh token
         const refreshed = await this.refreshToken();
         if (refreshed) {
            headers["Authorization"] = `Bearer ${this.accessToken}`;
            const retryResponse = await fetch(`${API_URL}${path}`, {
               ...options,
               headers,
               credentials: "include",
            });
            if (retryResponse.ok) {
               return retryResponse.json();
            }
         }
         this.onUnauthorized?.();
         throw new Error("Unauthorized");
      }

      if (!response.ok) {
         const error = await response.json().catch(() => ({ error: "Request failed" }));
         throw new Error(error.error || `HTTP ${response.status}`);
      }

      return response.json();
   }

   private async refreshToken(): Promise<boolean> {
      try {
         const response = await fetch(`${API_URL}/api/v1/auth/refresh`, {
            method: "POST",
            credentials: "include",
         });
         if (response.ok) {
            const data = await response.json();
            this.accessToken = data.access_token;
            if (typeof window !== "undefined") {
               window.dispatchEvent(
                  new CustomEvent("token-refreshed", {
                     detail: { accessToken: data.access_token },
                  })
               );
            }
            return true;
         }
      } catch {
         // Refresh failed
      }
      return false;
   }

   // Auth
   async register(email: string, password: string) {
      return this.request<{ message: string; user: import("@/types").User }>(
         "/api/v1/auth/register",
         { method: "POST", body: JSON.stringify({ email, password }) }
      );
   }

   async login(email: string, password: string) {
      return this.request<import("@/types").LoginResponse>(
         "/api/v1/auth/login",
         { method: "POST", body: JSON.stringify({ email, password }) }
      );
   }

   async logout() {
      return this.request<{ message: string }>("/api/v1/auth/logout", {
         method: "POST",
      });
   }

   async deleteAccount() {
      return this.request<{ message: string }>("/api/v1/auth/account", {
         method: "DELETE",
      });
   }

   // Games
   async getRecentGames() {
      return this.request<import("@/types").NBAGame[]>("/api/v1/games/recent");
   }

   async getUpcomingGames() {
      return this.request<import("@/types").NBAGame[]>("/api/v1/games/upcoming");
   }

   async getGameStats(gameId: number) {
      return this.request<{
         game: import("@/types").NBAGame;
         team_stats: import("@/types").GameStats[];
         player_stats: import("@/types").PlayerStats[];
      }>(`/api/v1/games/${gameId}/stats`);
   }

   async getGameOdds(gameId: number) {
      return this.request<import("@/types").GameOdds[]>(
         `/api/v1/games/${gameId}/odds`
      );
   }

   async generateInsights(gameId: number) {
      return this.request<import("@/types").AIInsight>(
         `/api/v1/games/${gameId}/insights`,
         { method: "POST" }
      );
   }

   // Bets
   async getBets() {
      return this.request<import("@/types").UserBet[]>("/api/v1/bets");
   }

   async createBet(bet: {
      game_id: number;
      bet_type: string;
      selection: string;
      odd_value: number;
      stake: number;
   }) {
      return this.request<import("@/types").UserBet>("/api/v1/bets", {
         method: "POST",
         body: JSON.stringify(bet),
      });
   }

   async updateBetResult(betId: number, result: string) {
      return this.request<{ message: string }>(
         `/api/v1/bets/${betId}/result`,
         { method: "PATCH", body: JSON.stringify({ result }) }
      );
   }

   async deleteBet(betId: number) {
      return this.request<{ message: string }>(`/api/v1/bets/${betId}`, {
         method: "DELETE",
      });
   }

   // Bankroll
   async getBankroll() {
      return this.request<{
         bankroll: import("@/types").Bankroll | null;
         transactions: import("@/types").BankrollTransaction[];
      }>("/api/v1/bankroll");
   }

   async deposit(amount: number) {
      return this.request<import("@/types").Bankroll>(
         "/api/v1/bankroll/deposit",
         { method: "POST", body: JSON.stringify({ amount }) }
      );
   }

   async withdraw(amount: number) {
      return this.request<import("@/types").Bankroll>(
         "/api/v1/bankroll/withdraw",
         { method: "POST", body: JSON.stringify({ amount }) }
      );
   }

   async getBankrollStats() {
      return this.request<import("@/types").BankrollStats>(
         "/api/v1/bankroll/stats"
      );
   }

   // Account
   async getProfile() {
      return this.request<import("@/types").User>("/api/v1/account/profile");
   }

   async updateProfile(email: string) {
      return this.request<import("@/types").User>("/api/v1/account/profile", {
         method: "PATCH",
         body: JSON.stringify({ email }),
      });
   }

   async changePassword(currentPassword: string, newPassword: string) {
      return this.request<{ message: string }>("/api/v1/account/password", {
         method: "PATCH",
         body: JSON.stringify({
            current_password: currentPassword,
            new_password: newPassword,
         }),
      });
   }

   // Predictions
   async getPredictions() {
      return this.request<import("@/types").AIPrediction[]>(
         "/api/v1/predictions"
      );
   }
}

export const apiClient = new ApiClient();
export default apiClient;
