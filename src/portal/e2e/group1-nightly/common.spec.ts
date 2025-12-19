import { test, expect, login } from '../fixtures/harbor';
import { createProject, pushImage, waitForProjectInList } from '../utils';

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
  
  // Create project and navigate into it
  await createProject(harborPage, projectName, true);
});

test('create a new public project', async ({ harborPage }) => {
  const projectName = `public_project_${Date.now()}`;
  
  await createProject(harborPage, projectName, false, true);
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
  
  // Create a new project
  await createProject(harborPage, projectName);
  
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

test('delete multiple projects', async ({ harborPage, harborUser }) => {
  const d = new Date();
  const dateStr = d.toLocaleString('en-US', { month: '2-digit' }) + Math.floor(d.getTime() / 1000);
  const projectWithArtifacts = `projecta${dateStr}`;
  const projectWithoutArtifacts = `projectb${dateStr}`;
  const image = 'hello-world';
  
  // Create public projects
  await createProject(harborPage, projectWithArtifacts, false, true);
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  await harborPage.waitForTimeout(100);
  
  await createProject(harborPage, projectWithoutArtifacts, false, true);
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  
  // Push image to first project only
  const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
  await pushImage({
    ip: harborIp,
    user: harborUser.username,
    pwd: harborUser.password,
    project: projectWithArtifacts,
    imageWithOrWithoutTag: image,
    needPullFirst: true,
    localRegistry: process.env.LOCAL_REGISTRY || 'docker.io',
    localRegistryNamespace: process.env.LOCAL_REGISTRY_NAMESPACE || 'library',
  });
  
  // Navigate back to projects list
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  
  // Select both projects
  const projectARow = harborPage.getByRole('row', { name: new RegExp(projectWithArtifacts) });
  const projectBRow = harborPage.getByRole('row', { name: new RegExp(projectWithoutArtifacts) });
  
  await projectARow.locator('label').click();
  await projectBRow.locator('label').click();
  
  // Click ACTION and Delete
  await harborPage.getByText('ACTION').click();
  await harborPage.getByRole('button', { name: 'Delete' }).click();
  await harborPage.getByRole('button', { name: 'DELETE' }).click();
  
  // Wait for deletion to process
  await harborPage.waitForTimeout(1000);
  
  // Verify project with artifacts still exists (deletion should fail)
  await expect(harborPage.getByRole('link', { name: projectWithArtifacts })).toBeVisible({ timeout: 5000 });
  
  // Verify project without artifacts was deleted
  await expect(harborPage.getByRole('link', { name: projectWithoutArtifacts })).not.toBeVisible({ timeout: 5000 });
});

test('delete multi repos', async ({ harborPage, harborUser }) => {
  const d = new Date();
  const dateStr = d.toLocaleString('en-US', { month: '2-digit' }) + Math.floor(d.getTime() / 1000);
  const projectName = `project${dateStr}`;
  const repos = ['hello-world', 'busybox'];
  
  // Create project and push images
  await createProject(harborPage, projectName, false);
  
  const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
  for (const repo of repos) {
    await pushImage({
      ip: harborIp,
      user: harborUser.username,
      pwd: harborUser.password,
      project: projectName,
      imageWithOrWithoutTag: repo,
      needPullFirst: true,
      localRegistry: process.env.LOCAL_REGISTRY || 'docker.io',
      localRegistryNamespace: process.env.LOCAL_REGISTRY_NAMESPACE || 'library',
    });
  }
  
  // Navigate into the project
  await waitForProjectInList(harborPage, projectName, 15000, true);
  
  // Select both repositories
  for (const repo of repos) {
    const repoRow = harborPage.getByRole('row', { name: new RegExp(repo) });
    await repoRow.locator('label').click();
  }
  
  // Click ACTION and Delete
  await harborPage.getByRole('button', { name: 'Delete' }).click();
  await harborPage.getByRole('button', { name: 'DELETE', exact: true }).click();

  // Wait for deletion to process
  await harborPage.waitForTimeout(1000);
  
  // Verify both repositories were deleted
  for (const repo of repos) {
    await expect(harborPage.getByRole('link', { name: repo })).not.toBeVisible({ timeout: 5000 });
  }
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
    // Create project and navigate into it
    await createProject(harborPage, projectName, true);
    
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
  
  // Create project
  await createProject(harborPage, projectName);
  
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
  await waitForProjectInList(harborPage, projectName, 15000, true);
  
  // Wait for image to appear in the repository list
  await expect(harborPage.getByRole('link', { name: new RegExp(image) })).toBeVisible({ timeout: 10000 });
  
  // Logout
  await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
  await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();
});


test('project level policy public', async ({ harborPage, harborUser }) => {
  const d = new Date();
  const time = d.getTime();
  const projectName = `test_project_${time}`;

  // Create project and navigate into it
  await createProject(harborPage, projectName, true);
  
  // Wait for project page to load
  await harborPage.waitForLoadState('networkidle');
  
  // Click the application button to access Configuration
  const appButton = harborPage.getByRole('application').locator('button');
  await appButton.waitFor({ state: 'visible', timeout: 5000 });
  await appButton.click();
  
  // Navigate to Configuration tab and make project public
  await harborPage.getByRole('tab', { name: 'Configuration' }).locator('a').click();
  await harborPage.getByText('Public', { exact: true }).click();
  await harborPage.getByRole('button', { name: 'SAVE' }).click();
  
  // Wait for save to complete
  await harborPage.waitForTimeout(1000);
  
  // Logout
  await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
  await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();

  // Login again to verify
  await login(harborPage, process.env.HARBOR_BASE_URL, harborUser);
  
  // Navigate to Projects page
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  await harborPage.waitForLoadState('networkidle');
  
  // Search through all pages to verify the project is now public
  const startTime = Date.now();
  const timeout = 15000;
  let found = false;
  
  while (Date.now() - startTime < timeout && !found) {
    // Check if project row with Public status is visible on current page
    const projectRow = harborPage.getByRole('row', { name: new RegExp(projectName) });
    
    if (await projectRow.isVisible()) {
      // Verify it shows Public status
      await expect(projectRow.getByText('Public')).toBeVisible({ timeout: 5000 });
      found = true;
      break;
    }
    
    // Check if Next Page button is enabled
    const nextButton = harborPage.getByRole('button', { name: 'Next Page' });
    const isNextEnabled = await nextButton.isEnabled().catch(() => false);
    
    if (isNextEnabled) {
      // Click next page and wait for content to load
      await nextButton.click();
      await harborPage.waitForTimeout(500);
    } else {
      // No more pages, check one final time
      if (await projectRow.isVisible()) {
        await expect(projectRow.getByText('Public')).toBeVisible({ timeout: 5000 });
        found = true;
      } else {
        throw new Error(`Project "${projectName}" not found in project list after checking all pages`);
      }
      break;
    }
  }
  
  if (!found) {
    throw new Error(`Timeout waiting for project "${projectName}" to appear in project list`);
  }
});

test('goto harbor API docs', async ({ harborPage, context }) => {
  // Navigate to API Docs - this may open in a new tab
  const [newPage] = await Promise.all([
    context.waitForEvent('page'),
    harborPage.getByRole('link', { name: 'Harbor API V2.0' }).click()
  ]);
  
  // Wait for the new page to load
  await newPage.waitForLoadState();
  
  // Wait for API Docs page to load by checking for Swagger UI element
  await expect(newPage.locator('.swagger-ui')).toBeVisible({ timeout: 10000 });
});

test('repo size', async ({ harborPage, harborUser }) => {
  const projectName = `project${Date.now()}`;
  const image = 'alpine';

  await createProject(harborPage, projectName);
  
  // Push image with specific tag using the utility function
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
  
  // Navigate to the project
  await waitForProjectInList(harborPage, projectName, 15000, true);
  
  // Navigate into the repository
  await expect(harborPage.getByRole('link', { name: new RegExp(image) })).toBeVisible({ timeout: 10000 });
  await harborPage.getByRole('link', { name: new RegExp(image) }).click();
  
  // Wait for and verify the repo size is displayed (alpine 2.6 is approximately 3.68MiB)
  await expect(harborPage.getByText(/3\.6[0-9]MiB/)).toBeVisible({ timeout: 10000 });
});

test('edit token expire', async ({ harborPage, harborUser }) => {
  // Navigate to Configuration -> System Settings
  await harborPage.getByRole('link', { name: 'Configuration' }).click();
  await harborPage.getByRole('button', { name: 'System Settings' }).click();
  
  // Modify token expiration to 20 minutes
  const tokenInput = harborPage.getByRole('spinbutton', { name: 'Token Expiration (Minutes) *' });
  await tokenInput.fill('20');
  
  // Save the configuration
  await harborPage.getByRole('button', { name: 'SAVE' }).click();
  
  // Wait for save to complete
  await harborPage.waitForTimeout(1000);
  
  // Logout
  await harborPage.getByRole('button', { name: harborUser.username, exact: true }).click();
  await harborPage.getByRole('menuitem', { name: 'Log Out' }).click();
  
  // Login again to verify
  await login(harborPage, process.env.HARBOR_BASE_URL, harborUser);
  
  // Navigate to Configuration -> System Settings
  await harborPage.getByRole('link', { name: 'Configuration' }).click();
  await harborPage.getByRole('button', { name: 'System Settings' }).click();
  
  // Verify token expiration is 20 minutes
  const tokenInputVerify = harborPage.getByRole('spinbutton', { name: 'Token Expiration (Minutes) *' });
  await expect(tokenInputVerify).toHaveValue('20');
  
  // Reset to default (30 minutes)
  await tokenInputVerify.fill('30');
  await harborPage.getByRole('button', { name: 'SAVE' }).click();
  
  // Wait for save to complete
  await harborPage.waitForTimeout(1000);
});

test('statistics info', async ({ harborPage, harborUser }) => {
  const d = new Date();
  const dateStr = d.toLocaleString('en-US', { month: '2-digit' }) + Math.floor(d.getTime() / 1000);
  
  // Navigate to Projects page to see statistics
  await harborPage.getByRole('link', { name: 'Projects' }).click();
  
  // Get initial statistics counts
  const getPrivateRepoCount = async () => {
    const text = await harborPage.locator('statistics-panel').getByText('Private').nth(1).locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  const getPrivateProjectCount = async () => {
    const text = await harborPage.locator('statistics-panel').getByText('Private').first().locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  const getPublicRepoCount = async () => {
    const text = await harborPage.getByText('Public').nth(1).locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  const getPublicProjectCount = async () => {
    const text = await harborPage.getByText('Public').first().locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  const getTotalRepoCount = async () => {
    const text = await harborPage.getByText('Total').nth(1).locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  const getTotalProjectCount = async () => {
    const text = await harborPage.getByText('Total').first().locator('..').textContent();
    return parseInt(text?.match(/\d+/)?.[0] || '0');
  };
  
  // Capture initial counts
  const privateRepoCount1 = await getPrivateRepoCount();
  const privateProjectCount1 = await getPrivateProjectCount();
  const publicRepoCount1 = await getPublicRepoCount();
  const publicProjectCount1 = await getPublicProjectCount();
  const totalRepoCount1 = await getTotalRepoCount();
  const totalProjectCount1 = await getTotalProjectCount();

  console.log('Initial Counts:', {
    privateRepoCount1,
    privateProjectCount1,
    publicRepoCount1,
    publicProjectCount1,
    totalRepoCount1,
    totalProjectCount1,
  });
  
  // Create private and public projects
  const privateProjectName = `private${dateStr}`;
  const publicProjectName = `public${dateStr}`;
  const image = 'hello-world';
  
  // Create private project and push image
  await createProject(harborPage, privateProjectName, false, false);
  
  const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
  await pushImage({
    ip: harborIp,
    user: harborUser.username,
    pwd: harborUser.password,
    project: privateProjectName,
    imageWithOrWithoutTag: image,
    needPullFirst: true,
    localRegistry: process.env.LOCAL_REGISTRY || 'docker.io',
    localRegistryNamespace: process.env.LOCAL_REGISTRY_NAMESPACE || 'library',
  });
  
  // Create public project and push image
  await createProject(harborPage, publicProjectName, false, true);
  
  await pushImage({
    ip: harborIp,
    user: harborUser.username,
    pwd: harborUser.password,
    project: publicProjectName,
    imageWithOrWithoutTag: image,
    needPullFirst: true,
    localRegistry: process.env.LOCAL_REGISTRY || 'docker.io',
    localRegistryNamespace: process.env.LOCAL_REGISTRY_NAMESPACE || 'library',
  });
  
  // Calculate expected counts
  const expectedPrivateProjectCount = privateProjectCount1 + 1;
  const expectedPrivateRepoCount = privateRepoCount1 + 1;
  const expectedPublicProjectCount = publicProjectCount1 + 1;
  const expectedPublicRepoCount = publicRepoCount1 + 1;
  const expectedTotalRepoCount = totalRepoCount1 + 2;
  const expectedTotalProjectCount = totalProjectCount1 + 2;

  console.log('Expected Counts:', {
    expectedPrivateProjectCount,
    expectedPrivateRepoCount,
    expectedPublicProjectCount,
    expectedPublicRepoCount,
    expectedTotalRepoCount,
    expectedTotalProjectCount,
  });
  
  // Refresh the page to update statistics
  await harborPage.reload({ waitUntil: 'networkidle' });
  
  // Wait for statistics to update
  await harborPage.waitForTimeout(2000);
  
  // Get updated statistics counts
  const privateRepoCount2 = await getPrivateRepoCount();
  const privateProjectCount2 = await getPrivateProjectCount();
  const publicRepoCount2 = await getPublicRepoCount();
  const publicProjectCount2 = await getPublicProjectCount();
  const totalRepoCount2 = await getTotalRepoCount();
  const totalProjectCount2 = await getTotalProjectCount();

  console.log('Updated Counts:', {
    privateRepoCount2,
    privateProjectCount2,
    publicRepoCount2,
    publicProjectCount2,
    totalRepoCount2,
    totalProjectCount2,
  });
  
  // Verify all statistics match expected values
  expect(privateProjectCount2).toBe(expectedPrivateProjectCount);
  expect(privateRepoCount2).toBe(expectedPrivateRepoCount);
  expect(publicProjectCount2).toBe(expectedPublicProjectCount);
  expect(publicRepoCount2).toBe(expectedPublicRepoCount);
  expect(totalProjectCount2).toBe(expectedTotalProjectCount);
  expect(totalRepoCount2).toBe(expectedTotalRepoCount);
});