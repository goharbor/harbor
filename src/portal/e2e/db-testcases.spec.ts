import { test, expect } from '@playwright/test';

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

    return username
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

test('Update User Comment', async ({ page }) => {
    // Creating user
    const username = await createUser(page)

    //Login with user credentials
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(username);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill("Harbor12345");

    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Updating user comment
    await page.getByRole('button', { name: username }).click();
    await page.getByRole('menuitem', { name: 'User Profile' }).click();
    await page.getByRole('textbox', { name: 'Comments' }).click();
    await page.getByRole('textbox', { name: 'Comments' }).fill('test1234');
    await page.getByRole('button', { name: 'OK' }).click();
})

test('Update User Password', async ({ page }) => {
    // Create user
    const username = await createUser(page)

    // Login with user credentials
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(username);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill("Harbor12345");

    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Update Password
    await page.getByRole('button', { name: username }).click();
    await page.getByRole('menuitem', { name: 'Change Password' }).click();
    await page.getByRole('textbox', { name: 'Current Password*' }).click();
    await page.getByRole('textbox', { name: 'Current Password*' }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'New Password*' }).click();
    await page.getByRole('textbox', { name: 'New Password*' }).fill('Test1234');
    await page.getByRole('textbox', { name: 'Confirm Password*' }).click();
    await page.getByRole('textbox', { name: 'Confirm Password*' }).fill('Test1234');
    await page.getByRole('button', { name: 'OK' }).click();

    // Logout after update
    await page.getByRole('button', { name: username }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();

    // Login with Updated Password
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(username);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Test1234');

    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Logout
    await page.getByRole('button', { name: username }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
})

test('Edit Self Registration', async ({ page }) => {
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
    if (await page.locator('clr-checkbox-wrapper label').isChecked()) {
        await page.locator('clr-checkbox-wrapper label').click();
        await page.getByRole('button', { name: 'SAVE' }).click();
    }

    //Logout
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).dblclick();

    // Checks whether Signup is visible or not
    await expect(page.getByText('Sign up for an account')).not.toBeVisible();

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

    await expect(page.locator('clr-checkbox-wrapper label')).not.toBeChecked();

    //Update self-registration Status
    if (!(await page.locator('clr-checkbox-wrapper label').isChecked())) {
        await page.locator('clr-checkbox-wrapper label').click();
        await page.getByRole('button', { name: 'SAVE' }).click();
    }

    //Logout
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).dblclick();
})

test('Admin Add New Users', async ({ page }) => {
    await page.goto('/');

    //Login with Admin Credentials
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Check Self configuration is Enabled
    await page.getByRole('link', { name: 'Configuration' }).click();
    await expect(page.locator('clr-checkbox-wrapper label')).toBeChecked();

    // Add users with self registration is enabled
    await page.getByRole('link', { name: 'Users' }).click();
    await page.getByRole('button', { name: 'New User' }).click();

    let timestamp = Date.now();
    let username = "harbor-user" + timestamp    
    await page.locator('#username').click();
    await page.locator('#username').fill(username);

    let email = username + "@example.com"
    await page.locator('#email').click();
    await page.locator('#email').fill(email);
    await page.getByRole('textbox', { name: 'First and last name*' }).click();
    await page.getByRole('textbox', { name: 'First and last name*' }).fill(username);
    await page.getByRole('textbox', { name: 'Password*', exact: true }).click();
    await page.getByRole('textbox', { name: 'Password*', exact: true }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'Confirm Password*' }).click();
    await page.getByRole('textbox', { name: 'Confirm Password*' }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'Comments' }).click();
    await page.getByRole('textbox', { name: 'Comments' }).fill('harbortest');

    await page.getByRole('button', { name: 'OK' }).click();

    await expect(page.getByText('New user created successfully.')).toBeVisible();

    // Add users with self registration is disabled
    await page.getByRole('link', { name: 'Configuration' }).click();

    await page.locator('clr-checkbox-wrapper label').click();
    await expect(page.locator('clr-checkbox-wrapper label')).not.toBeChecked();

    await page.getByRole('button', { name: 'SAVE' }).click();

    await page.getByRole('link', { name: 'Users' }).click();
    await page.getByRole('button', { name: 'New User' }).click();
    timestamp = Date.now();
    username = "harbor-user" + timestamp  
    await page.locator('#username').click();
    await page.locator('#username').fill(username);

    email = username + "@example.com"
    await page.locator('#email').click();
    await page.locator('#email').fill(email);
    await page.getByRole('textbox', { name: 'First and last name*' }).click();
    await page.getByRole('textbox', { name: 'First and last name*' }).fill(username);
    await page.getByRole('textbox', { name: 'Password*', exact: true }).click();
    await page.getByRole('textbox', { name: 'Password*', exact: true }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'Confirm Password*' }).click();
    await page.getByRole('textbox', { name: 'Confirm Password*' }).fill('Harbor12345');
    await page.getByRole('textbox', { name: 'Comments' }).click();
    await page.getByRole('textbox', { name: 'Comments' }).fill('harbortest');

    await page.getByRole('button', { name: 'OK' }).click();

    await expect(page.getByText('New user created successfully.')).toBeVisible();
})