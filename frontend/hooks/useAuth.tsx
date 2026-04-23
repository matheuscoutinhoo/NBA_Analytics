"use client";

import {
   createContext,
   useContext,
   useState,
   useEffect,
   useCallback,
   type ReactNode,
} from "react";
import { apiClient } from "@/lib/api-client";
import type { User } from "@/types";

interface AuthContextType {
   user: User | null;
   isLoading: boolean;
   isAuthenticated: boolean;
   login: (email: string, password: string) => Promise<void>;
   register: (email: string, password: string) => Promise<void>;
   logout: () => Promise<void>;
   deleteAccount: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
   const [user, setUser] = useState<User | null>(null);
   const [isLoading, setIsLoading] = useState(true);

   const clearAuth = useCallback(() => {
      setUser(null);
      apiClient.setAccessToken(null);
   }, []);

   useEffect(() => {
      apiClient.setOnUnauthorized(clearAuth);

      // Listen for token refresh events
      const handleTokenRefresh = (e: Event) => {
         const detail = (e as CustomEvent).detail;
         apiClient.setAccessToken(detail.accessToken);
      };
      window.addEventListener("token-refreshed", handleTokenRefresh);

      // Try to load user profile (will use refresh token cookie)
      apiClient
         .getProfile()
         .then((u) => {
            setUser(u);
         })
         .catch(() => {
            clearAuth();
         })
         .finally(() => setIsLoading(false));

      return () => {
         window.removeEventListener("token-refreshed", handleTokenRefresh);
      };
   }, [clearAuth]);

   const login = async (email: string, password: string) => {
      const response = await apiClient.login(email, password);
      apiClient.setAccessToken(response.access_token);
      setUser(response.user);
   };

   const register = async (email: string, password: string) => {
      await apiClient.register(email, password);
   };

   const logout = async () => {
      try {
         await apiClient.logout();
      } finally {
         clearAuth();
      }
   };

   const deleteAccount = async () => {
      await apiClient.deleteAccount();
      clearAuth();
   };

   return (
      <AuthContext.Provider
         value={{
            user,
            isLoading,
            isAuthenticated: !!user,
            login,
            register,
            logout,
            deleteAccount,
         }}
      >
         {children}
      </AuthContext.Provider>
   );
}

export function useAuth() {
   const context = useContext(AuthContext);
   if (!context) {
      throw new Error("useAuth must be used within an AuthProvider");
   }
   return context;
}
