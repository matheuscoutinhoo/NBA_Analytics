import { test, expect } from "@playwright/test";

// Helper to login before tests
async function login(page: any) {
  await page.goto("/login");
  await page.fill('input[type="email"]', "e2e-dashboard@test.com");
  await page.fill('input[type="password"]', "TestPass123!");
  await page.click('button[type="submit"]');
  await page.waitForURL("**/dashboard", { timeout: 10000 });
}

test.describe("Dashboard", () => {
  test.beforeAll(async ({ browser }) => {
    // Register the test user
    const page = await browser.newPage();
    await page.goto("/register");
    await page.fill('input[type="email"]', "e2e-dashboard@test.com");
    const passwordInputs = page.locator('input[type="password"]');
    await passwordInputs.nth(0).fill("TestPass123!");
    await passwordInputs.nth(1).fill("TestPass123!");
    await page.click('button[type="submit"]');
    await page.waitForURL("**/dashboard", { timeout: 10000 });
    await page.close();
  });

  test("should display stats cards", async ({ page }) => {
    await login(page);
    await expect(page.getByText("Win Rate")).toBeVisible();
    await expect(page.getByText("ROI")).toBeVisible();
    await expect(page.getByText("Total Bets")).toBeVisible();
  });

  test("should have navigation sidebar", async ({ page }) => {
    await login(page);
    await expect(page.getByText("Dashboard")).toBeVisible();
    await expect(page.getByText("Games")).toBeVisible();
    await expect(page.getByText("Bets")).toBeVisible();
    await expect(page.getByText("Bankroll")).toBeVisible();
    await expect(page.getByText("Account")).toBeVisible();
  });

  test("should navigate to games page", async ({ page }) => {
    await login(page);
    await page.click("text=Games");
    await page.waitForURL("**/games");
    await expect(page.getByRole("heading", { name: /games/i })).toBeVisible();
  });

  test("should navigate to bets page", async ({ page }) => {
    await login(page);
    await page.click("text=Bets");
    await page.waitForURL("**/bets");
    await expect(page.getByRole("heading", { name: /bet tracking/i })).toBeVisible();
  });

  test("should navigate to bankroll page", async ({ page }) => {
    await login(page);
    await page.click("text=Bankroll");
    await page.waitForURL("**/bankroll");
    await expect(page.getByRole("heading", { name: /bankroll/i })).toBeVisible();
  });
});
