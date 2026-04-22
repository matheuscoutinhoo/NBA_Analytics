import { create } from "zustand";
import type { User } from "@/lib/types";

interface AuthState {
   user: User | null;
   isAuthenticated: boolean;
   isLoading: boolean;
   setUser: (user: User | null) => void;
   setTokens: (accessToken: string, refreshToken: string) => void;
   logout: () => void;
   setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
   user: null,
   isAuthenticated: false,
   isLoading: true,
   setUser: (user) =>
      set({ user, isAuthenticated: !!user, isLoading: false }),
   setTokens: (accessToken, refreshToken) => {
      if (typeof window !== "undefined") {
         localStorage.setItem("access_token", accessToken);
         localStorage.setItem("refresh_token", refreshToken);
      }
   },
   logout: () => {
      if (typeof window !== "undefined") {
         localStorage.removeItem("access_token");
         localStorage.removeItem("refresh_token");
      }
      set({ user: null, isAuthenticated: false });
   },
   setLoading: (loading) => set({ isLoading: loading }),
}));

interface UIState {
   sidebarOpen: boolean;
   toggleSidebar: () => void;
   setSidebarOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
   sidebarOpen: true,
   toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
   setSidebarOpen: (open) => set({ sidebarOpen: open }),
}));
