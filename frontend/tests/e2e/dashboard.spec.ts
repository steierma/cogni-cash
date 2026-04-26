import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('admin');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page).toHaveURL('/');
  });

  test('should display dashboard components', async ({ page }) => {
    // Check for greeting or title - this ensures the page has loaded its main data
    await expect(page.getByText('Cash Flow Overview')).toBeVisible({ timeout: 15000 });

    // Check for summary cards
    await expect(page.getByText('Net Savings')).toBeVisible();
    await expect(page.getByText('Total Expenses')).toBeVisible();
    await expect(page.getByText('Total Income')).toBeVisible();

    // Check for charts - these are custom flex-based charts
    await expect(page.getByText('Cash Flow Overview')).toBeVisible();
    await expect(page.getByText('Top Spending Categories')).toBeVisible();
  });


  test('should navigate to different pages from sidebar', async ({ page }) => {
    // Check sidebar navigation
    await page.getByRole('link', { name: 'Transactions' }).click();
    await expect(page).toHaveURL('/transactions');

    await page.getByRole('link', { name: 'Forecasting' }).click();
    await expect(page).toHaveURL('/forecasting');

    await page.getByRole('link', { name: 'Payslips' }).click();
    await expect(page).toHaveURL('/payslips');
  });
});
