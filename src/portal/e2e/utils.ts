import { Page, expect } from '@playwright/test';

const { exec } = require('child_process');
const util = require('util');
const execAsync = util.promisify(exec);

export async function runCommand(cmd) {
  try {
    const { stdout, stderr } = await execAsync(cmd);
    return stdout + stderr;
  } catch (error) {
    return error.stdout + error.stderr;
  }
}

export async function pullImage({ ip, user, pwd, project, image, tag = null, isRobot = false }) {
  console.log(`\nRunning docker pull ${image}...`);
  const imageWithTag = tag === null ? image : `${image}:${tag}`;
  const loginCmd = isRobot
    ? `docker login -u robot\\$${project}+${user} -p ${pwd} ${ip}`
    : `docker login -u ${user} -p ${pwd} ${ip}`;
  await runCommand(loginCmd);

  const pullCmd = `docker pull ${ip}/${project}/${imageWithTag}`;
  const output = await runCommand(pullCmd);
  console.log(output);

  if (!output.includes('Digest:')) throw new Error('Output missing Digest');
  if (!output.includes('Status:')) throw new Error('Output missing Status');
  if (output.includes('No such image:')) throw new Error('Image not found');
}

export async function pushImage({
  ip, user, pwd, project, imageWithOrWithoutTag,
  needPullFirst = true, sha256 = null, isRobot = false,
  localRegistry, localRegistryNamespace
}) {
  const d = Date.now();
  const imageInUse = sha256 === null
    ? imageWithOrWithoutTag
    : `${imageWithOrWithoutTag}@sha256:${sha256}`;
  const imageInUseWithTag = sha256 === null
    ? imageWithOrWithoutTag
    : `${imageWithOrWithoutTag}:${sha256}`;

  await new Promise(r => setTimeout(r, 3000));
  console.log(`\nRunning docker push ${imageWithOrWithoutTag}...`);

  let imageToTag = imageWithOrWithoutTag;
  if (needPullFirst) {
    await runCommand(`docker pull ${localRegistry}/${localRegistryNamespace}/${imageInUse}`);
    imageToTag = imageInUse;
  }

  const loginCmd = isRobot
    ? `docker login -u robot\\$${project}+${user} -p ${pwd} ${ip}`
    : `docker login -u ${user} -p ${pwd} ${ip}`;
  await runCommand(loginCmd);

  if (needPullFirst) {
    await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  } else {
    await runCommand(`docker tag ${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  }

  await runCommand(`docker push ${ip}/${project}/${imageInUseWithTag}`);
  await runCommand(`docker logout ${ip}`);
  await new Promise(r => setTimeout(r, 1000));
}

export async function pushImageWithTag({
  ip, user, pwd, project, image, tag, tag1 = 'latest',
  localRegistry, localRegistryNamespace
}) {
  console.log(`\nRunning docker push ${image}...`);
  await runCommand(`docker pull ${localRegistry}/${localRegistryNamespace}/${image}:${tag1}`);
  await runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);
  await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${image}:${tag1} ${ip}/${project}/${image}:${tag}`);
  await runCommand(`docker push ${ip}/${project}/${image}:${tag}`);
  await runCommand(`docker logout ${ip}`);
}

export async function waitForProjectInList(harborPage: Page, projectName: string, timeout: number = 15000, goto: boolean = false) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    // Check if project is visible on current page
    const projectLink = harborPage.getByRole('link', { name: projectName });
    if (await projectLink.isVisible()) {
      if (goto) {
        await projectLink.click();
      }
      return;
    }
    
    // Check if Next Page button is enabled
    const nextButton = harborPage.getByRole('button', { name: 'Next Page' });
    const isNextEnabled = await nextButton.isEnabled().catch(() => false);
    
    if (isNextEnabled) {
      // Click next page and wait for content to load
      await nextButton.click();
      await harborPage.waitForTimeout(500);
    } else {
      // No more pages, wait a bit and check one more time
      await harborPage.waitForTimeout(500);
      if (await projectLink.isVisible()) {
        if (goto) {
          await projectLink.click();
        }
        return;
      }
      throw new Error(`Project "${projectName}" not found in project list after checking all pages`);
    }
  }
  
  throw new Error(`Timeout waiting for project "${projectName}" to appear in project list`);
}

export async function createProject(harborPage: Page, projectName: string, goto: boolean = false, isPublic: boolean = false) {
  // Click on New Project button
  await harborPage.getByRole('button', { name: 'New Project' }).click();
  
  // Wait for modal to appear
  const modal = harborPage.getByLabel('New Project');
  await expect(modal.getByRole('heading', { name: 'New Project', level: 3 })).toBeVisible();
  
  // Fill in the project name
  await modal.getByRole('textbox').first().fill(projectName);

  if (isPublic) {
    // Set project as public - check the Public checkbox
    await modal.getByText('Public').click();
  }
  
  // Wait for OK button to be enabled and click it
  const okButton = modal.getByRole('button', { name: 'OK' });
  await okButton.waitFor({ state: 'visible' });
  await expect(okButton).toBeEnabled();
  await okButton.click();
  
  // Wait for modal to close
  await modal.waitFor({ state: 'hidden', timeout: 5000 });
  
  // Verify project was created by checking if it appears in the project list (with pagination support)
  await waitForProjectInList(harborPage, projectName, 15000, goto);
}