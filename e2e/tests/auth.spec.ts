import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
   test('should display login page', async ({ page }) => {
      await page.goto('/login');
      await expect(page.getByText('Welcome back')).toBeVisible();
      await expect(page.getByLabel('Email')).toBeVisible();
      await expect(page.getByLabel('Password')).toBeVisible();
      await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
   });

   test('should display register page with compliance elements', async ({ page }) => {
      await page.goto('/register');
      await expect(page.getByText('Create Account')).toBeVisible();
      await expect(page.getByLabel('Username')).toBeVisible();
      await expect(page.getByLabel('Email')).toBeVisible();
      await expect(page.getByText('18 years or older')).toBeVisible();
      await expect(page.getByText('consent to the processing')).toBeVisible();
      await expect(page.getByText('informational purposes only')).toBeVisible();
   });

   test('should require age verification', async ({ page }) => {
      await page.goto('/register');
      await page.getByLabel('Username').fill('testuser');
      await page.getByLabel('Email').fill('test@example.com');
      await page.getByLabel('Password').fill('password123');
      // Do NOT check age verification
      await page.getByRole('button', { name: /create account/i }).click();
      await expect(page.getByText('18 or older')).toBeVisible();
   });

   test('should navigate from login to register', async ({ page }) => {
      await page.goto('/login');
      await page.getByText('Create one').click();
      await expect(page).toHaveURL(/\/register/);
   });
});

test.describe('Landing Page', () => {
   test('should display landing page with key elements', async ({ page }) => {
      await page.goto('/');
      await expect(page.getByText('NBA Analytics & Intelligence')).toBeVisible();
      await expect(page.getByText('Informational purposes only')).toBeVisible();
      await expect(page.getByText('Responsible Gaming')).toBeVisible();
      await expect(page.getByText('does not accept')).toBeVisible();
   });

   test('should have navigation links', async ({ page }) => {
      await page.goto('/');
      await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible();
      await expect(page.getByRole('link', { name: /get started/i })).toBeVisible();
   });
});
