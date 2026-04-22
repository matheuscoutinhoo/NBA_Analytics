import { test, expect } from '@playwright/test';

test.describe('Bankroll Management', () => {
   test.beforeEach(async ({ page }) => {
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

   test('should display bankroll page with entry form', async ({ page }) => {
      await page.goto('/dashboard/bankroll');
      await expect(page.getByText('New Entry')).toBeVisible();
      await expect(page.getByText('DEPOSIT')).toBeVisible();
      await expect(page.getByText('WITHDRAWAL')).toBeVisible();
      await expect(page.getByText('BET')).toBeVisible();
      await expect(page.getByText('WIN')).toBeVisible();
   });

   test('should display performance page with charts', async ({ page }) => {
      await page.goto('/dashboard/performance');
      await expect(page.getByText('Balance Evolution')).toBeVisible();
      await expect(page.getByText('Monthly Performance')).toBeVisible();
   });
});
