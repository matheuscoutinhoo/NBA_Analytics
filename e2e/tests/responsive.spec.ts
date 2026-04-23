import { test, expect } from "@playwright/test";

test.describe("Responsive Design", () => {
   test("login page should be responsive on mobile", async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 812 }); // iPhone X
      await page.goto("/login");
      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
      await expect(page.getByPlaceholder(/email/i)).toBeVisible();
   });

   test("login page should be responsive on tablet", async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto("/login");
      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
   });

   test("login page should be responsive on desktop", async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      await page.goto("/login");
      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
   });
});
