# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: payslips.spec.ts >> Payslips Page >> should open the "Add Payslip" modal
- Location: tests/e2e/payslips.spec.ts:43:3

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
  3  | test.describe('Payslips Page', () => {
  4  |   test.beforeEach(async ({ page }) => {
  5  |     // Login before each test
> 6  |     await page.goto('/login');
     |                ^ Error: page.goto: Test timeout of 30000ms exceeded.
  7  |     await page.getByLabel('Username').fill('admin');
  8  |     await page.getByLabel('Password').fill('admin');
  9  |     await page.getByRole('button', { name: 'Sign In' }).click();
  10 |     await expect(page).toHaveURL('/');
  11 |     
  12 |     // Navigate to Payslips page
  13 |     await page.goto('/payslips');
  14 |     await expect(page).toHaveURL('/payslips');
  15 |   });
  16 | 
  17 |   test('should display payslips title and subtitle', async ({ page }) => {
  18 |     await expect(page.getByRole('heading', { name: 'Payslips' })).toBeVisible();
  19 |     await expect(page.getByText('Manage your HR documents and track your true net income.')).toBeVisible();
  20 |   });
  21 | 
  22 |   test('should display at least one payslip in the table', async ({ page }) => {
  23 |     // Wait for the table to load (the spinner should disappear)
  24 |     await expect(page.locator('table')).toBeVisible();
  25 |     
  26 |     // Check if there are rows in the tbody (excluding the "No payslips found" row)
  27 |     const rows = page.locator('tbody tr');
  28 |     
  29 |     // In CI, it might take a moment to load
  30 |     await expect(rows.first()).toBeVisible();
  31 |     
  32 |     // We expect at least 1 real data row. 
  33 |     // If it says "No payslips found", the test should fail.
  34 |     await expect(page.getByText('No payslips found')).not.toBeVisible();
  35 |     
  36 |     const rowCount = await rows.count();
  37 |     expect(rowCount).toBeGreaterThan(0);
  38 |     
  39 |     // Verify one of the seeded content fields (€ is used for EUR in en locale)
  40 |     await expect(page.locator('tbody')).toContainText('€');
  41 |   });
  42 | 
  43 |   test('should open the "Add Payslip" modal', async ({ page }) => {
  44 |     await page.getByRole('button', { name: 'Manual Override' }).click();
  45 |     
  46 |     // Check if the modal is visible
  47 |     const modal = page.locator('div[role="dialog"], .fixed.inset-0').filter({ hasText: 'Import with Overrides' });
  48 |     await expect(modal).toBeVisible();
  49 |     
  50 |     // Close the modal using the X button in the modal header
  51 |     // The modal header has the title and the X button
  52 |     await modal.locator('button').filter({ has: page.locator('svg') }).first().click();
  53 |     
  54 |     await expect(modal).not.toBeVisible();
  55 |   });
  56 | 
  57 |   test('should filter by employer', async ({ page }) => {
  58 |     // The select for employer (3rd select in the filters bar)
  59 |     const employerSelect = page.locator('select').nth(2);
  60 |     await expect(employerSelect).toBeVisible();
  61 |     
  62 |     // Select "All Employers" initially
  63 |     await expect(employerSelect).toHaveValue('All');
  64 |   });
  65 |   
  66 |   test('should toggle column visibility', async ({ page }) => {
  67 |     const columnsButton = page.getByRole('button', { name: 'Columns' });
  68 |     await columnsButton.click();
  69 |     
  70 |     // Scope search to the open menu
  71 |     const menu = page.locator('div.absolute.right-0.mt-2');
  72 |     await expect(menu).toBeVisible();
  73 |     
  74 |     // Based on translations: payslips.modals.gross is "Gross (Gesamtbrutto)"
  75 |     const grossOption = menu.getByRole('button', { name: 'Gross (Gesamtbrutto)' });
  76 |     await expect(grossOption).toBeVisible();
  77 |     
  78 |     // Toggle Gross (it is visible by default)
  79 |     await grossOption.click();
  80 |     
  81 |     // Verify table header doesn't contain Gross anymore 
  82 |     // Header uses payslips.table.gross which is "Gross (Brutto)"
  83 |     const tableHeader = page.locator('thead');
  84 |     
  85 |     // Use a custom timeout or just wait for the condition
  86 |     await expect(tableHeader).not.toContainText('Gross (Brutto)', { timeout: 10000 });
  87 |     
  88 |     // Toggle it back on
  89 |     await grossOption.click();
  90 |     await expect(tableHeader).toContainText('Gross (Brutto)');
  91 |   });
  92 | });
  93 | 
```