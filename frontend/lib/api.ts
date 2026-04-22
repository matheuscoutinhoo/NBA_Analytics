const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
   private baseUrl: string;

   constructor(baseUrl: string) {
      this.baseUrl = baseUrl;
   }

   private getToken(): string | null {
      if (typeof window === "undefined") return null;
      return localStorage.getItem("access_token");
   }

   private async request<T>(
      endpoint: string,
      options: RequestInit = {}
   ): Promise<T> {
      const token = this.getToken();
      const headers: Record<string, string> = {
         "Content-Type": "application/json",
         ...(options.headers as Record<string, string>),
      };

      if (token) {
         headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch(`${this.baseUrl}${endpoint}`, {
         ...options,
         headers,
      });

      if (response.status === 401) {
         // Try refresh
         const refreshed = await this.tryRefresh();
         if (refreshed) {
            headers["Authorization"] = `Bearer ${this.getToken()}`;
            const retryResponse = await fetch(`${this.baseUrl}${endpoint}`, {
               ...options,
               headers,
            });
            if (!retryResponse.ok) {
               throw new Error(`API Error: ${retryResponse.status}`);
            }
            return retryResponse.json();
         }
         // Redirect to login
         if (typeof window !== "undefined") {
            localStorage.removeItem("access_token");
            localStorage.removeItem("refresh_token");
            window.location.href = "/login";
         }
         throw new Error("Unauthorized");
      }

      if (!response.ok) {
         const errorData = await response.json().catch(() => ({}));
         throw new Error(
            errorData?.error?.message || `API Error: ${response.status}`
         );
      }

      return response.json();
   }

   private async tryRefresh(): Promise<boolean> {
      const refreshToken = localStorage.getItem("refresh_token");
      if (!refreshToken) return false;

      try {
         const response = await fetch(`${this.baseUrl}/auth/refresh`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: refreshToken }),
         });

         if (!response.ok) return false;

         const data = await response.json();
         if (data.data) {
            localStorage.setItem("access_token", data.data.access_token);
            localStorage.setItem("refresh_token", data.data.refresh_token);
            return true;
         }
         return false;
      } catch {
         return false;
      }
   }

   // Auth
   async register(data: {
      username: string;
      email: string;
      password: string;
      age_verified: boolean;
      consent: boolean;
   }) {
      return this.request<{ data: { user: any; tokens: any } }>("/auth/register", {
         method: "POST",
         body: JSON.stringify(data),
      });
   }

   async login(data: { email: string; password: string }) {
      return this.request<{ data: { user: any; tokens: any } }>("/auth/login", {
         method: "POST",
         body: JSON.stringify(data),
      });
   }

   async logout() {
      return this.request("/auth/logout", { method: "POST" });
   }

   async getMe() {
      return this.request<{ data: any }>("/auth/me");
   }

   // Analysis
   async getTodaysGames() {
      return this.request<{ data: { games: any[]; date: string; disclaimer: string } }>(
         "/analysis/games/today"
      );
   }

   async getGameAnalysis(gameId: string) {
      return this.request<{ data: any }>(`/analysis/games/${gameId}`);
   }

   async getTeams() {
      return this.request<{ data: any[] }>("/analysis/teams");
   }

   // Bankroll
   async createBankrollEntry(data: {
      game_id?: string;
      entry_type: string;
      bet_type?: string;
      odd?: number;
      stake: number;
      result?: number;
      notes?: string;
   }) {
      return this.request<{ data: any }>("/bankroll/entries", {
         method: "POST",
         body: JSON.stringify(data),
      });
   }

   async getBankrollEntries(page = 1, perPage = 20) {
      return this.request<{ data: any[]; meta: any }>(
         `/bankroll/entries?page=${page}&per_page=${perPage}`
      );
   }

   async deleteBankrollEntry(entryId: string) {
      return this.request(`/bankroll/entries/${entryId}`, {
         method: "DELETE",
      });
   }

   async getDashboard() {
      return this.request<{ data: any }>("/bankroll/dashboard");
   }

   // User
   async getPreferences() {
      return this.request<{ data: any }>("/user/preferences");
   }

   async updatePreferences(data: Record<string, any>) {
      return this.request<{ data: any }>("/user/preferences", {
         method: "PUT",
         body: JSON.stringify(data),
      });
   }

   async exportData() {
      return this.request<{ data: any }>("/user/export");
   }

   async deleteAccount() {
      return this.request("/user/delete", { method: "DELETE" });
   }
}

export const api = new ApiClient(API_URL);
