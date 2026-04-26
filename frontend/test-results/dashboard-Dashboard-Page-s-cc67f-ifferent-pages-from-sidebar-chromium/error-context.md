# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: dashboard.spec.ts >> Dashboard Page >> should navigate to different pages from sidebar
- Location: tests/e2e/dashboard.spec.ts:28:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/login
Call log:
  - navigating to "http://localhost:5173/login", waiting until "load"

```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | test.describe('Dashboard Page', () => {
  4  |   test.beforeEach(async ({ page }) => {
  5  |     // Login before each test
> 6  |     await page.goto('/login');
     |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/login
  7  |     await page.getByLabel('Username').fill('admin');
  8  |     await page.getByLabel('Password').fill('admin');
  9  |     await page.getByRole('button', { name: 'Sign In' }).click();
  10 |     await expect(page).toHaveURL('/');
  11 |   });
  12 | 
  13 |   test('should display dashboard components', async ({ page }) => {
  14 |     // Check for greeting or title - this ensures the page has loaded its main data
  15 |     await expect(page.getByText('Cash Flow Overview')).toBeVisible({ timeout: 15000 });
  16 | 
  17 |     // Check for summary cards
  18 |     await expect(page.getByText('Net Savings')).toBeVisible();
  19 |     await expect(page.getByText('Total Expenses')).toBeVisible();
  20 |     await expect(page.getByText('Total Income')).toBeVisible();
  21 | 
  22 |     // Check for charts - these are custom flex-based charts
  23 |     await expect(page.getByText('Cash Flow Overview')).toBeVisible();
  24 |     await expect(page.getByText('Top Spending Categories')).toBeVisible();
  25 |   });
  26 | 
  27 | 
  28 |   test('should navigate to different pages from sidebar', async ({ page }) => {
  29 |     // Check sidebar navigation
  30 |     await page.getByRole('link', { name: 'Transactions' }).click();
  31 |     await expect(page).toHaveURL('/transactions');
  32 | 
  33 |     await page.getByRole('link', { name: 'Forecasting' }).click();
  34 |     await expect(page).toHaveURL('/forecasting');
  35 | 
  36 |     await page.getByRole('link', { name: 'Payslips' }).click();
  37 |     await expect(page).toHaveURL('/payslips');
  38 |   });
  39 | });
  40 | 
```