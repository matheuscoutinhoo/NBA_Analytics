import { renderHook, act } from "@testing-library/react";
import { useAuthStore, useUIStore } from "@/store";

describe("useAuthStore", () => {
   beforeEach(() => {
      useAuthStore.setState({
         user: null,
         isAuthenticated: false,
         isLoading: true,
      });
   });

   it("initializes with default state", () => {
      const { result } = renderHook(() => useAuthStore());
      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.isLoading).toBe(true);
   });

   it("sets user and authenticates", () => {
      const { result } = renderHook(() => useAuthStore());
      const mockUser = {
         id: "123",
         username: "testuser",
         email: "test@example.com",
         role: "USER",
         is_active: true,
         age_verified: true,
         consent_given: true,
         xp_points: 50,
         created_at: "2024-01-01",
         updated_at: "2024-01-01",
      };

      act(() => {
         result.current.setUser(mockUser);
      });

      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.isLoading).toBe(false);
   });

   it("clears user on logout", () => {
      const { result } = renderHook(() => useAuthStore());

      act(() => {
         result.current.setUser({
            id: "123",
            username: "testuser",
            email: "test@example.com",
            role: "USER",
            is_active: true,
            age_verified: true,
            consent_given: true,
            xp_points: 0,
            created_at: "",
            updated_at: "",
         });
      });

      act(() => {
         result.current.logout();
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
   });
});

describe("useUIStore", () => {
   it("toggles sidebar", () => {
      const { result } = renderHook(() => useUIStore());

      const initial = result.current.sidebarOpen;

      act(() => {
         result.current.toggleSidebar();
      });

      expect(result.current.sidebarOpen).toBe(!initial);
   });

   it("sets sidebar state directly", () => {
      const { result } = renderHook(() => useUIStore());

      act(() => {
         result.current.setSidebarOpen(false);
      });

      expect(result.current.sidebarOpen).toBe(false);

      act(() => {
         result.current.setSidebarOpen(true);
      });

      expect(result.current.sidebarOpen).toBe(true);
   });
});
