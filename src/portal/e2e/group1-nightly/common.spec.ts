import { test, expect } from '../fixtures/harbor';

test('sign-in and sign-out', async ({ harborPage, harborUser }) => {
    await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
    await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();
});

test('create a system label', async ({ harborPage }) => {
  const labelName = `label_${Date.now()}`;
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(labelName);
  await harborPage.getByText('OK').click();
});