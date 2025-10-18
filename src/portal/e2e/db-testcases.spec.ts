import { test } from '@playwright/test';

async function createUser(page) {
    await page.goto('/');
    await page.getByRole('link', { name: 'Sign up for an account' }).click();
    await page.locator('#username').click();
    await page.locator('#username').fill('user');
    await page.locator('#username').press('Home');
    const timestamp = Date.now();
    const username = "harbor-user" + timestamp
    await page.locator('#username').fill(username);
    await page.locator('#username').press('ArrowDown');
    await page.locator('new-user-form div').filter({ hasText: 'Email is only used for' }).nth(3).click();
    const email = username + "@example.com"
    await page.locator('#email').fill(email);
    await page.locator('clr-input-container').filter({ hasText: 'First and last name' }).locator('div').nth(1).click();
    await page.getByRole('textbox', { name: 'First and last name*' }).fill(username);
    await page.getByRole('textbox', { name: 'Password*', exact: true }).click();
    await page.getByRole('textbox', { name: 'Password*', exact: true }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'Confirm Password*' }).click();
    await page.getByRole('textbox', { name: 'Confirm Password*' }).fill('Harbor12345');
    await page.getByRole('button', { name: 'SIGN UP' }).click();
}


test('Create An New User', async ({ page }) => {
    // Login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');

    await page.getByRole('button', { name: 'LOG IN' }).click();

    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(5000);

    //Select Configuration
    await page.getByRole('link', { name: 'Configuration' }).click();

    //Update self-registration Status
    if (!(await page.locator('clr-checkbox-wrapper label').isChecked())) {
        await page.locator('clr-checkbox-wrapper label').click();
        await page.getByRole('button', { name: 'SAVE' }).click();
    }

    //Logout
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).dblclick();

    //Creating user
    await createUser(page)
});