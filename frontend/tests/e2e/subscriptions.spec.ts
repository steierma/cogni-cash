import { test, expect } from '@playwright/test';

test.describe('Subscriptions Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('admin');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page).toHaveURL('/');
    
    // Navigate to Subscriptions page
    await page.goto('/subscriptions');
    await expect(page).toHaveURL('/subscriptions');
  });

  test('should display subscriptions title', async ({ page }) => {
    // The heading is "Subscriptions" (from layout.subscriptions)
    await expect(page.getByRole('heading', { name: 'Subscriptions' })).toBeVisible();
    await expect(page.getByText('Manage your recurring fixed costs and discover new ones.')).toBeVisible();
  });

  test('should display discovered subscriptions', async ({ page }) => {
    // The discovery section should show identified patterns
    // We saw these in the previous run's dump
    await expect(page.locator('body')).toContainText('Cloud Services');
    await expect(page.locator('body')).toContainText('REWE Supermarket');
  });

  test('should display discovery status', async ({ page }) => {
    // The discovery settings button should be visible
    await expect(page.getByRole('button', { name: /Discovery Algorithm Settings/i })).toBeVisible();
    await expect(page.getByText(/New Patterns Detected/i)).toBeVisible();
  });
});
