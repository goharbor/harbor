import { expect, test } from '@playwright/test';

const harborUser = process.env.HARBOR_ADMIN || 'admin';
const harborPassword = process.env.HARBOR_PASSWORD || 'Harbor12345';

test('login and logout', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).fill(harborUser);
    await page.getByRole('textbox', { name: 'Password' }).fill(harborPassword);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    const userMenu = page.getByRole('button', {
        name: harborUser,
        exact: true,
    });
    await expect(userMenu).toBeVisible();
    await userMenu.click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});
