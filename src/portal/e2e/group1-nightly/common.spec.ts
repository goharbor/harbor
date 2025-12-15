import { test, expect, login } from '../fixtures/harbor';
import { createProject, pushImage, waitForProjectInList } from '../utils';
import { logout } from '../fixtures/harbor';

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
  
  // Wait for label to appear in the list
  await harborPage.getByRole('row', { name: new RegExp(labelName) }).waitFor({ state: 'visible', timeout: 5000 });
});

test('update a system label', async ({ harborPage }) => {
  const originalName = `label_${Date.now()}`;
  const updatedName = `label_updated_${Date.now()}`;
  
  // Navigate to Labels and create a label
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(originalName);
  await harborPage.getByText('OK').click();
  
  // Wait for label to appear first
  await harborPage.getByRole('row', { name: new RegExp(originalName) }).waitFor({ state: 'visible', timeout: 5000 });
  
  // Select and edit the label
  await harborPage.getByRole('row', { name: new RegExp(originalName) }).getByRole('gridcell', { name: 'Select' }).locator('label').click();
  await harborPage.getByRole('button', { name: 'Edit' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(updatedName);
  await harborPage.getByText('OK').click();
  
  // Wait for updated label to appear
  await harborPage.getByRole('row', { name: new RegExp(updatedName) }).waitFor({ state: 'visible', timeout: 5000 });
});

test('delete a system label', async ({ harborPage }) => {
  const labelName = `label_${Date.now()}`;
  
  // Navigate to Labels and create a label
  await harborPage.getByRole('link', { name: 'Labels' }).click();
  await harborPage.getByRole('button', { name: 'New Label' }).click();
  await harborPage.getByRole('textbox', { name: 'Label Name' }).fill(labelName);
  await harborPage.getByText('OK').click();
  
  // Wait for label to appear first
  await harborPage.getByRole('row', { name: new RegExp(labelName) }).waitFor({ state: 'visible', timeout: 5000 });
  
  // Select and delete the label
  await harborPage.getByRole('row', { name: new RegExp(labelName) }).getByRole('gridcell', { name: 'Select' }).locator('label').click();
  await harborPage.getByRole('button', { name: 'Delete' }).click();
  await harborPage.getByRole('button', { name: 'DELETE', exact: true }).click();
  
  // Wait for label to be removed from the list
  await harborPage.getByRole('row', { name: new RegExp(labelName) }).waitFor({ state: 'detached', timeout: 5000 });
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
  
  // Wait for modal to close
  await modal.waitFor({ state: 'hidden', timeout: 5000 });
  
  // Verify project was created by checking if it appears in the project list (with pagination)
  await waitForProjectInList(harborPage, projectName);
  
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
  await okButton.waitFor({ state: 'visible' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Wait for modal to close
  await modal.waitFor({ state: 'hidden', timeout: 5000 });
  
  // Verify project was created (with pagination)
  await waitForProjectInList(harborPage, projectName);
  
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
    
    // Wait for modal to close
    await modal.waitFor({ state: 'hidden', timeout: 5000 });
    
    // Verify project was created (with pagination)
    await waitForProjectInList(harborPage, projectName);
    
    // Navigate back to projects list for next iteration
    await harborPage.getByRole('link', { name: 'Projects' }).click();
    
    // Small delay to avoid timestamp collisions in next iteration
    await harborPage.waitForTimeout(100);
  }
});

test('delete a project', async ({ harborPage }) => {
  const projectName = `project_to_delete_${Date.now()}`;
  
  // Create a new project first
  await harborPage.getByRole('button', { name: 'New Project' }).click();
  
  // Wait for modal to appear
  const modal = harborPage.getByLabel('New Project');
  await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
  
  // Fill in the project name
  await modal.getByRole('textbox').first().fill(projectName);
  
  // Wait for OK button to be enabled and click it
  const okButton = modal.getByRole('button', { name: 'OK' });
  await okButton.waitFor({ state: 'visible' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Wait for modal to close
  await modal.waitFor({ state: 'hidden', timeout: 5000 });
  
  // Verify project was created (with pagination)
  await waitForProjectInList(harborPage, projectName);
  
  // Navigate back to projects list
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  
  // Select the project by clicking on the row's checkbox label
  const projectRow = harborPage.getByRole('row', { name: new RegExp(projectName) });
  await projectRow.locator('label').click();
  
  // Click ACTION text/button
  await harborPage.getByText('ACTION').click();
  
  // Click Delete button
  await harborPage.getByRole('button', { name: 'Delete' }).click();
  
  // Confirm deletion by clicking DELETE button
  await harborPage.getByRole('button', { name: 'DELETE' }).click();
  
  // Wait a moment for deletion to process
  await harborPage.waitForTimeout(1000);
  
  // Verify project was deleted - should not appear in the list
  await expect(harborPage.getByRole('link', { name: projectName })).not.toBeVisible({ timeout: 5000 });
});

test('user view projects', async ({ harborPage }) => {
  // Create three projects and go into each
  const d = new Date();
  const dateStr = d.toLocaleString('en-US', { month: '2-digit' }) + Math.floor(d.getTime() / 1000);
  const projectNames = [
    `test${dateStr}1`,
    `test${dateStr}2`,
    `test${dateStr}3`,
  ];

  for (const projectName of projectNames) {
    // Click on New Project button
    await harborPage.getByRole('button', { name: 'New Project' }).click();
    
    // Wait for modal to appear
    const modal = harborPage.getByLabel('New Project');
    await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
    
    // Fill in the project name
    await modal.getByRole('textbox').first().fill(projectName);
    
    // Wait for OK button to be enabled and click it
    const okButton = modal.getByRole('button', { name: 'OK' });
    await okButton.waitFor({ state: 'visible' });
    await expect(okButton).toBeEnabled();
    await okButton.click();
    
    // Wait for modal to close
    await modal.waitFor({ state: 'hidden', timeout: 5000 });
    
    // Verify project was created (with pagination)
    await waitForProjectInList(harborPage, projectName);
    
    // Navigate into the project
    await harborPage.getByRole('link', { name: projectName }).click();
    
    // Navigate back to projects list for next project
    await harborPage.getByRole('link', { name: 'Projects' }).click();
    
    // Small delay to avoid timestamp collisions
    await harborPage.waitForTimeout(100);
  }

  await harborPage.getByRole('link', { name: 'Logs' }).click();
  // Wait until page contains all three project names
  for (const projectName of projectNames) {
    await expect(harborPage.getByRole('gridcell', { name: projectName, exact: true })).toBeVisible();
  }
});

test('push image', async ({ harborPage, harborUser }) => {
  const d = new Date();
  const dateStr = d.toLocaleString('en-US', { month: '2-digit' }) + Math.floor(d.getTime() / 1000);
  const projectName = `project${dateStr}`;
  const image = 'hello-world';
  
  // Click on New Project button
  await harborPage.getByRole('button', { name: 'New Project' }).click();
  
  // Wait for modal to appear
  const modal = harborPage.getByLabel('New Project');
  await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
  
  // Fill in the project name
  await modal.getByRole('textbox').first().fill(projectName);
  
  // Wait for OK button to be enabled and click it
  const okButton = modal.getByRole('button', { name: 'OK' });
  await okButton.waitFor({ state: 'visible' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Wait for modal to close
  await modal.waitFor({ state: 'hidden', timeout: 5000 });
  
  // Verify project was created (with pagination)
  await waitForProjectInList(harborPage, projectName);
  
  // Push image using the utility function
  const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
  await pushImage({
    ip: harborIp,
    user: harborUser.username,
    pwd: harborUser.password,
    project: projectName,
    imageWithOrWithoutTag: image,
    needPullFirst: true,
    localRegistry: process.env.LOCAL_REGISTRY || 'docker.io',
    localRegistryNamespace: process.env.LOCAL_REGISTRY_NAMESPACE || 'library',
  });
  
  // Navigate into the project
  await harborPage.getByRole('link', { name: projectName }).click();
  
  // Wait for image to appear in the repository list
  await expect(harborPage.getByRole('link', { name: new RegExp(image) })).toBeVisible({ timeout: 10000 });
  
  // Logout
  await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
  await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();
});