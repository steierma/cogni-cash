import { test, expect } from '@playwright/test';

test.describe('Transactions Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('admin');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page).toHaveURL('/');
    
    // Navigate to Transactions page
    await page.goto('/transactions');
    await expect(page).toHaveURL('/transactions');
  });

  test('should display transactions table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Transactions' })).toBeVisible();
    
    // Click Search to load transactions (required by TransactionsPage logic)
    await page.getByRole('button', { name: 'Search' }).first().click();
    
    // Wait for the table to load
    const table = page.locator('table');
    await expect(table).toBeVisible({ timeout: 10000 });
    
    // Check for seeded data
    await expect(table).toContainText('Salary admin Corp');
    await expect(table).toContainText('REWE Supermarket');
  });

  test('should filter transactions by description', async ({ page }) => {
    // Click Search to load transactions
    await page.getByRole('button', { name: 'Search' }).first().click();
    
    const table = page.locator('table');
    await expect(table).toBeVisible({ timeout: 10000 });

    const searchInput = page.getByPlaceholder(/Search description or reference/i);
    await searchInput.fill('REWE');
    
    // Click Search again to apply filter
    await page.getByRole('button', { name: 'Search' }).first().click();
    
    // Should show REWE transactions
    await expect(table).toContainText('REWE Supermarket');
    
    // Should NOT show Salary transactions
    await expect(table).not.toContainText('Salary admin Corp');
  });
});
