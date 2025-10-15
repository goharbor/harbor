import { test, expect } from '@playwright/test';

test('test', async ({ page }) => {
  // Recording...await page.getByRole('button', { name: 'Advanced' }).click();
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('harbor-cli');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();
  await page.getByRole('button', { name: 'harbor-cli' }).click();
  await page.getByRole('button', { name: 'harbor-cli' }).dblclick();
  await page.getByRole('menuitem', { name: 'Log Out' }).dblclick();
});