import { test, expect } from '@playwright/test';

test.describe('Games & Analysis', () => {
   test.beforeEach(async ({ page }) => {
      // Mock auth state via localStorage
      await page.goto('/login');
      await page.evaluate(() => {
         localStorage.setItem('auth-storage', JSON.stringify({
            state: {
               user: { id: '1', username: 'testuser', email: 'test@test.com', role: 'user' },
               tokens: { access_token: 'mock-token', refresh_token: 'mock-refresh' },
            },
         }));
      });
   });

   test('should display games list page', async ({ page }) => {
      await page.goto('/dashboard/games');
      await expect(page.getByText("Today's Games")).toBeVisible();
   });

   test('should display disclaimer on game analysis page', async ({ page }) => {
      await page.goto('/dashboard/games/mock-game-id');
      await expect(page.getByText('informational purposes')).toBeVisible();
   });
});
