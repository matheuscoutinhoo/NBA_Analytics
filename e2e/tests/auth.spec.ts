import { test, expect } from "@playwright/test";

test.describe("Auth Flow", () => {
   const testEmail = `test-${Date.now()}@example.com`;
   const testPassword = "TestPass123!";

   test("should show login page", async ({ page }) => {
      await page.goto("/login");
      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
      await expect(page.getByPlaceholder(/email/i)).toBeVisible();
      await expect(page.getByPlaceholder(/password/i)).toBeVisible();
   });

   test("should navigate to register", async ({ page }) => {
      await page.goto("/login");
      await page.click("text=Create an account");
      await expect(page).toHaveURL(/.*register/);
   });

   test("should register a new user", async ({ page }) => {
      await page.goto("/register");
      await page.fill('input[type="email"]', testEmail);
      const passwordInputs = page.locator('input[type="password"]');
      await passwordInputs.nth(0).fill(testPassword);
      await passwordInputs.nth(1).fill(testPassword);
      await page.click('button[type="submit"]');
      await page.waitForURL("**/dashboard", { timeout: 10000 });
      await expect(page).toHaveURL(/.*dashboard/);
   });

   test("should show validation errors for weak password", async ({ page }) => {
      await page.goto("/register");
      await page.fill('input[type="email"]', "weak@test.com");
      const passwordInputs = page.locator('input[type="password"]');
      await passwordInputs.nth(0).fill("123");
      await passwordInputs.nth(1).fill("123");
      await page.click('button[type="submit"]');
      await expect(page.locator(".bg-red-500\\/10")).toBeVisible({ timeout: 5000 });
   });

   test("should login with existing credentials", async ({ page }) => {
      await page.goto("/login");
      await page.fill('input[type="email"]', testEmail);
      await page.fill('input[type="password"]', testPassword);
      await page.click('button[type="submit"]');
      await page.waitForURL("**/dashboard", { timeout: 10000 });
      await expect(page).toHaveURL(/.*dashboard/);
   });

   test("should show error for wrong password", async ({ page }) => {
      await page.goto("/login");
      await page.fill('input[type="email"]', testEmail);
      await page.fill('input[type="password"]', "WrongPassword1!");
      await page.click('button[type="submit"]');
      await expect(page.locator(".bg-red-500\\/10")).toBeVisible({ timeout: 5000 });
   });

   test("should redirect unauthenticated users to login", async ({ page }) => {
      await page.goto("/dashboard");
      await page.waitForURL("**/login", { timeout: 10000 });
      await expect(page).toHaveURL(/.*login/);
   });
});
