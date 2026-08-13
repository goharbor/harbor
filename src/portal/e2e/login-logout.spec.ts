import { test } from '@playwright/test';

const harborUser = process.env.HARBOR_ADMIN || 'admin';
const harborPassword = process.env.HARBOR_PASSWORD || 'Harbor12345';

test('login and logout', async ({ page }) => {
    // login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(harborUser);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill(harborPassword);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // logout
    await page.getByRole('button', { name: harborUser, exact: true }).waitFor();
    await page.getByRole('button', { name: harborUser, exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});
