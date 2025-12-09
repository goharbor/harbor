import { test, expect } from '../fixtures/harbor';

test('sign-out', async ({ harborPage, harborUser }) => {
  // Sign-out if already signed in
  await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
  await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();
});

test('create a system label', async ({ harborPage }) => {
  const labelName = `label_${Date.now()}`;

  // Navigate to Labels and create a label
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(labelName);
  await harborPage.getByText('OK').click();
});

test('update a system label', async ({ harborPage }) => {
  const originalName = `label_${Date.now()}`;
  const updatedName = `label_updated_${Date.now()}`;
  
  // Navigate to Labels and create a label
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(originalName);
  await harborPage.getByText('OK').click();
  
  // Select and edit the label
  await harborPage.getByRole('row', { name: new RegExp(originalName) }).getByRole('gridcell', { name: 'Select' }).locator('label').click();
  await harborPage.getByRole('button', { name: 'Edit' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(updatedName);
  await harborPage.getByText('OK').click();
});

test('delete a system label', async ({ harborPage }) => {
  const labelName = `label_${Date.now()}`;
  
  // Navigate to Labels and create a label
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(labelName);
  await harborPage.getByText('OK').click();
  
  // Select and delete the label
  await harborPage.getByRole('row', { name: new RegExp(labelName) }).getByRole('gridcell', { name: 'Select' }).locator('label').click();
  await harborPage.getByRole('button', { name: 'Delete' }).click();
  await harborPage.getByRole('button', { name: 'DELETE', exact: true }).click();
});