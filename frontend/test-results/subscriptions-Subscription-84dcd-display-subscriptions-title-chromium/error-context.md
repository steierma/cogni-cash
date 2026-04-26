# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: subscriptions.spec.ts >> Subscriptions Page >> should display subscriptions title
- Location: tests/e2e/subscriptions.spec.ts:17:3

# Error details

```
Test timeout of 30000ms exceeded while running "beforeEach" hook.
```

```
Error: page.goto: Test timeout of 30000ms exceeded.
Call log:
  - navigating to "http://localhost:5173/login", waiting until "load"

```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | test.describe('Subscriptions Page', () => {
  4  |   test.beforeEach(async ({ page }) => {
  5  |     // Login before each test
> 6  |     await page.goto('/login');
     |                ^ Error: page.goto: Test timeout of 30000ms exceeded.
  7  |     await page.getByLabel('Username').fill('admin');
  8  |     await page.getByLabel('Password').fill('admin');
  9  |     await page.getByRole('button', { name: 'Sign In' }).click();
  10 |     await expect(page).toHaveURL('/');
  11 |     
  12 |     // Navigate to Subscriptions page
  13 |     await page.goto('/subscriptions');
  14 |     await expect(page).toHaveURL('/subscriptions');
  15 |   });
  16 | 
  17 |   test('should display subscriptions title', async ({ page }) => {
  18 |     // The heading is "Subscriptions" (from layout.subscriptions)
  19 |     await expect(page.getByRole('heading', { name: 'Subscriptions' })).toBeVisible();
  20 |     await expect(page.getByText('Manage your recurring fixed costs and discover new ones.')).toBeVisible();
  21 |   });
  22 | 
  23 |   test('should display discovered subscriptions', async ({ page }) => {
  24 |     // The discovery section should show identified patterns
  25 |     // We saw these in the previous run's dump
  26 |     await expect(page.locator('body')).toContainText('Cloud Services');
  27 |     await expect(page.locator('body')).toContainText('REWE Supermarket');
  28 |   });
  29 | 
  30 |   test('should display discovery status', async ({ page }) => {
  31 |     // The discovery settings button should be visible
  32 |     await expect(page.getByRole('button', { name: /Discovery Algorithm Settings/i })).toBeVisible();
  33 |     await expect(page.getByText(/New Patterns Detected/i)).toBeVisible();
  34 |   });
  35 | });
  36 | 
```