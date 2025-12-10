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

test('create a new project', async ({ harborPage }) => {
  const projectName = `test_project_${Date.now()}`;
  
  // Click on New Project button
  await harborPage.getByRole('button', { name: 'New Project' }).click();
  
  // Wait for modal to appear
  const modal = harborPage.getByLabel('New Project');
  await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
  
  // Fill in the project name - using the first textbox in the modal
  await modal.getByRole('textbox').first().fill(projectName);
  
  // Wait for OK button to be enabled and click it
  const okButton = modal.getByRole('button', { name: 'OK' });
  await okButton.waitFor({ state: 'visible' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Verify project was created by checking if it appears in the project list
  await harborPage.getByRole('link', { name: projectName }).waitFor({ state: 'visible', timeout: 5000 });
  
  // Navigate into the project
  await harborPage.getByRole('link', { name: projectName }).click();
});

test('create a new public project', async ({ harborPage }) => {
  const projectName = `public_project_${Date.now()}`;
  
  // Click on New Project button
  await harborPage.getByRole('button', { name: 'New Project' }).click();
  
  // Wait for modal to appear
  const modal = harborPage.getByLabel('New Project');
  await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
  
  // Fill in the project name
  await modal.getByRole('textbox').first().fill(projectName);
  
  // Set project as public - check the Public checkbox
  
  await modal.getByText('Public').click();
  
  // Click OK button to create the project
  const okButton = modal.getByRole('button', { name: 'OK' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Verify project was created
  await harborPage.getByRole('link', { name: projectName }).waitFor({ state: 'visible', timeout: 5000 });
  
  // Navigate into the project
  await harborPage.getByRole('link', { name: projectName }).click();
  
  // Verify it's public by checking the grid shows "Public"
  await harborPage.getByRole('link', { name: 'Projects' }).click();
});

test('create projects with different storage quotas', async ({ harborPage }) => {
  const testCases = [
    { quota: '100', unit: 'MiB' },
    { quota: '500', unit: 'MiB' },
    { quota: '1', unit: 'GiB' },
    { quota: '5', unit: 'GiB' },
    { quota: '1', unit: 'TiB' },
  ];

  for (const testCase of testCases) {
    const projectName = `quota_${testCase.quota}${testCase.unit.toLowerCase()}_${Date.now()}`;
    
    // Click on New Project button
    await harborPage.getByRole('button', { name: 'New Project' }).click();
    
    // Wait for modal to appear
    const modal = harborPage.getByLabel('New Project');
    await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
    
    // Fill in the project name
    await modal.getByRole('textbox').first().fill(projectName);
    
    // Set storage quota
    const quotaTextbox = modal.getByRole('textbox', { name: /Project quota limits.*Proxy Cache/ });
    await quotaTextbox.fill(testCase.quota);
    
    // Select storage quota unit
    const unitCombobox = modal.getByRole('combobox');
    await unitCombobox.selectOption(testCase.unit);
    
    // Click OK button to create the project
    const okButton = modal.getByRole('button', { name: 'OK' });
    await expect(okButton).toBeEnabled();
    await okButton.click();
    
    // Verify project was created
    await harborPage.getByRole('link', { name: projectName }).waitFor({ state: 'visible', timeout: 5000 });
    
    // Navigate back to projects list for next iteration
    await harborPage.getByRole('link', { name: 'Projects' }).click();
  }
});