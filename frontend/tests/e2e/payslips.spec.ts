import { test, expect } from '@playwright/test';

test.describe('Payslips Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('admin');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page).toHaveURL('/');
    
    // Navigate to Payslips page
    await page.goto('/payslips');
    await expect(page).toHaveURL('/payslips');
  });

  test('should display payslips title and subtitle', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Payslips' })).toBeVisible();
    await expect(page.getByText('Manage your HR documents and track your true net income.')).toBeVisible();
  });

  test('should display at least one payslip in the table', async ({ page }) => {
    // Wait for the table to load (the spinner should disappear)
    await expect(page.locator('table')).toBeVisible();
    
    // Check if there are rows in the tbody (excluding the "No payslips found" row)
    const rows = page.locator('tbody tr');
    
    // In CI, it might take a moment to load
    await expect(rows.first()).toBeVisible();
    
    // We expect at least 1 real data row. 
    // If it says "No payslips found", the test should fail.
    await expect(page.getByText('No payslips found')).not.toBeVisible();
    
    const rowCount = await rows.count();
    expect(rowCount).toBeGreaterThan(0);
    
    // Verify one of the seeded content fields (€ is used for EUR in en locale)
    await expect(page.locator('tbody')).toContainText('€');
  });

  test('should open the "Add Payslip" modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Manual Override' }).click();
    
    // Check if the modal is visible
    const modal = page.locator('div[role="dialog"], .fixed.inset-0').filter({ hasText: 'Import with Overrides' });
    await expect(modal).toBeVisible();
    
    // Close the modal using the X button in the modal header
    // The modal header has the title and the X button
    await modal.locator('button').filter({ has: page.locator('svg') }).first().click();
    
    await expect(modal).not.toBeVisible();
  });

  test('should filter by employer', async ({ page }) => {
    // The select for employer (3rd select in the filters bar)
    const employerSelect = page.locator('select').nth(2);
    await expect(employerSelect).toBeVisible();
    
    // Select "All Employers" initially
    await expect(employerSelect).toHaveValue('All');
  });
  
  test('should toggle column visibility', async ({ page }) => {
    const columnsButton = page.getByRole('button', { name: 'Columns' });
    await columnsButton.click();
    
    // Scope search to the open menu
    const menu = page.locator('div.absolute.right-0.mt-2');
    await expect(menu).toBeVisible();
    
    // Based on translations: payslips.modals.gross is "Gross (Gesamtbrutto)"
    const grossOption = menu.getByRole('button', { name: 'Gross (Gesamtbrutto)' });
    await expect(grossOption).toBeVisible();
    
    // Toggle Gross (it is visible by default)
    await grossOption.click();
    
    // Verify table header doesn't contain Gross anymore 
    // Header uses payslips.table.gross which is "Gross (Brutto)"
    const tableHeader = page.locator('thead');
    
    // Use a custom timeout or just wait for the condition
    await expect(tableHeader).not.toContainText('Gross (Brutto)', { timeout: 10000 });
    
    // Toggle it back on
    await grossOption.click();
    await expect(tableHeader).toContainText('Gross (Brutto)');
  });
});
