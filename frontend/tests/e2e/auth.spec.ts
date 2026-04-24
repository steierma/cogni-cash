import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('should login successfully with admin credentials', async ({ page }) => {
    // Go to the login page
    await page.goto('/login');

    // Check if we are on the login page
    await expect(page.getByText('Welcome Back')).toBeVisible();

    // Fill in the credentials
    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password').fill('admin');

    // Click the sign in button
    await page.getByRole('button', { name: 'Sign In' }).click();

    // Should redirect to dashboard
    await expect(page).toHaveURL('/');
    
    // Check for some dashboard content (e.g., "Good morning" or "Cash Flow Overview")
    // Note: The greeting depends on the time of day, so we check for something static
    await expect(page.getByText('Cash Flow Overview')).toBeVisible();
  });

  test('should show error with invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel('Username').fill('wronguser');
    await page.getByLabel('Password').fill('wrongpassword');
    await page.getByRole('button', { name: 'Sign In' }).click();

    // Error message should be visible
    await expect(page.getByText('Invalid username or password.')).toBeVisible();
    
    // Should stay on login page
    await expect(page).toHaveURL('/login');
  });
});
