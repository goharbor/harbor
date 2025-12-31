import { expect, Page, test } from '@playwright/test'

async function loginAsAdmin(page: Page) {
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
    await page.getByRole('button', { name: 'LOG IN' }).click();
    await expect(page.getByRole('link', { name: 'Projects' })).toBeVisible();
}

test('Project Quota Sorting', async ({ page }) => {
    await loginAsAdmin(page);
})