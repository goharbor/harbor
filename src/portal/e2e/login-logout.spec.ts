import { test } from '@playwright/test';

const user: string = process.env.HARBOR_ADMIN || 'admin';
const pwd: string = process.env.HARBOR_ADMIN_PASSWD || 'Harbor12345';

test('login and logout', async ({ page }) => {
    // login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(user);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // logout
    await page.getByRole('button', { name: user, exact: true }).waitFor();
    await page.getByRole('button', { name: user, exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});
