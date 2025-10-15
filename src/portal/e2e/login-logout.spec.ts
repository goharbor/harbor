import { expect, test } from '@playwright/test';

test('login and logout', async ({ page }) => {
    // Recording...await page.getByRole('button', { name: 'Advanced' }).click();
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill('harbor-cli');
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');

    await page.getByRole('button', { name: 'LOG IN' }).click();

    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(5000);
    const userButton = page.getByRole('button', { name: 'harbor-cli' });
    await expect(userButton).toBeVisible();
    await expect(userButton).toBeEnabled();

    await page.getByRole('button', { name: 'harbor-cli' }).click();
    await page.getByRole('button', { name: 'harbor-cli' }).dblclick();
    await page.getByRole('menuitem', { name: 'Log Out' }).dblclick();
});
