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

function shellQuote(value) {
  return `'${String(value).replace(/'/g, `'\\''`)}'`;
}

async function dockerLogin(ip, user, pwd) {
  await runCommand(`printf '%s\n' ${shellQuote(pwd)} | docker login --username ${shellQuote(user)} --password-stdin ${shellQuote(ip)}`);
}

async function dockerPush(imageRef) {
  const help = await runCommand('docker push --help');
  const pushCmd = help.includes('--format string')
    ? `docker push --remove-signatures --format v2s2 ${shellQuote(imageRef)}`
    : `docker push ${shellQuote(imageRef)}`;
  const output = await runCommand(pushCmd);
  if (/denied|unauthorized|error|failed|invalid/i.test(output)) {
    throw new Error(`Image push failed: ${output}`);
  }
  return output;
}

async function ensureImage(imageRef) {
  const inspectOutput = await runCommand(`docker image inspect ${shellQuote(imageRef)}`);
  if (!/No such image|not found|Error response from daemon|image not known/i.test(inspectOutput)) {
    return;
  }

  const pullOutput = await runCommand(`docker pull ${shellQuote(imageRef)}`);
  if (/toomanyrequests|denied|unauthorized|error|failed/i.test(pullOutput)) {
    throw new Error(`Image pull failed: ${pullOutput}`);
  }
}

export async function pullImage({ ip, user, pwd, project, image, tag = null, isRobot = false }) {
  console.log(`\nRunning docker pull ${image}...`);
  const imageWithTag = tag === null ? image : `${image}:${tag}`;
  await dockerLogin(ip, isRobot ? `robot$${project}+${user}` : user, pwd);

  const pullCmd = `docker pull ${ip}/${project}/${imageWithTag}`;
  const output = await runCommand(pullCmd);
  console.log(output);

  if (/denied|unauthorized|not found|No such image|Error response from daemon/i.test(output)) {
    throw new Error(`Image pull failed: ${output}`);
  }
}

export async function cannotPullImage({ ip, user, pwd, project, image, tag = null, expectedError = null }) {
  console.log(`\nVerifying docker pull should fail for ${image}...`);
  const imageWithTag = tag === null ? image : `${image}:${tag}`;
  await dockerLogin(ip, user, pwd);

  const pullCmd = `docker pull ${ip}/${project}/${imageWithTag}`;
  const output = await runCommand(pullCmd);
  console.log(output);

  // Verify that the pull failed
  if (output.includes('Digest:') && output.includes('Status: Downloaded')) {
    throw new Error('Pull should have failed but succeeded');
  }

  // Verify expected error message if provided
  if (expectedError && !output.includes(expectedError)) {
    if (expectedError.includes('unauthorized') && /Cannot perform an interactive login|Login prior to pull|requested access/i.test(output)) {
      return output;
    }
    throw new Error(`Expected error message "${expectedError}" not found in output: ${output}`);
  }

  return output;
}

export async function cannotPushImage({ ip, user, pwd, project, image, expectedError = null, expectedError2 = null, localRegistry, localRegistryNamespace }) {
  console.log(`\nVerifying docker push should fail for ${image}...`);
  
  // Pull the image from local registry
  await ensureImage(`${localRegistry}/${localRegistryNamespace}/${image}`);
  
  // Login to Harbor
  await dockerLogin(ip, user, pwd);
  
  // Tag the image
  await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${image} ${ip}/${project}/${image}`);
  
  // Try to push and expect failure
  const output = await runCommand(`docker push ${ip}/${project}/${image}`);
  console.log(output);

  // Verify that the push failed
  if (output.includes('Pushed') || output.includes('digest:')) {
    throw new Error('Push should have failed but succeeded');
  }

  // Verify expected error messages if provided
  if (expectedError && !output.includes(expectedError)) {
    if (expectedError.includes('read only mode') && /^Error response from daemon:\s*$/i.test(output.trim())) {
      return output;
    }
    throw new Error(`Expected error message "${expectedError}" not found in output: ${output}`);
  }

  if (expectedError2 && !output.includes(expectedError2)) {
    throw new Error(`Expected error message "${expectedError2}" not found in output: ${output}`);
  }

  // Logout
  await runCommand(`docker logout ${ip}`);

  return output;
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
    await ensureImage(`${localRegistry}/${localRegistryNamespace}/${imageInUse}`);
    imageToTag = imageInUse;
  }

  await dockerLogin(ip, isRobot ? `robot$${project}+${user}` : user, pwd);

  if (needPullFirst) {
    await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  } else {
    await runCommand(`docker tag ${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  }

  await dockerPush(`${ip}/${project}/${imageInUseWithTag}`);
  await runCommand(`docker logout ${ip}`);
  await new Promise(r => setTimeout(r, 1000));
}

export async function pushImageWithTag({
  ip, user, pwd, project, image, tag, tag1 = 'latest',
  localRegistry, localRegistryNamespace
}) {
  console.log(`\nRunning docker push ${image}...`);
  await ensureImage(`${localRegistry}/${localRegistryNamespace}/${image}:${tag1}`);
  await dockerLogin(ip, user, pwd);
  await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${image}:${tag1} ${ip}/${project}/${image}:${tag}`);
  await dockerPush(`${ip}/${project}/${image}:${tag}`);
  await runCommand(`docker logout ${ip}`);
}

export async function waitForProjectInList(harborPage: Page, projectName: string, timeout: number = 15000, goto: boolean = false) {
  const startTime = Date.now();

  // Normalize pagination before walking forward through the project list.
  const previousButton = harborPage.getByRole('button', { name: 'Previous Page' });
  while (await previousButton.isEnabled().catch(() => false)) {
    await previousButton.click();
    await harborPage.waitForTimeout(300);
  }
  
  while (Date.now() - startTime < timeout) {
    // Check if project is visible on current page
    const projectLink = harborPage.getByRole('link', { name: projectName, exact: true });
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
  await modal.waitFor({ state: 'hidden', timeout: 15000 });
  
  // Verify project was created by checking if it appears in the project list (with pagination support)
  await waitForProjectInList(harborPage, projectName, 15000, goto);
}

export async function cosignGenerateKeyPair() {
  console.log("Generating Cosign key pair...");
  await runCommand("rm -f cosign.key cosign.pub");
  await runCommand("COSIGN_PASSWORD=\"\" cosign generate-key-pair");
}

export async function cosignSign(artifact: string) {
  console.log(`Signing artifact with Cosign: ${artifact}...`);
  await runCommand(`COSIGN_PASSWORD="" cosign sign -y --allow-insecure-registry --key cosign.key ${artifact}`);
}

export async function cosignVerify(artifact: string, shouldPass: boolean = true) {
  console.log(`Verifying artifact with Cosign: ${artifact}...`);
  const output = await runCommand(`COSIGN_PASSWORD="" cosign verify --key cosign.pub --allow-insecure-registry ${artifact}`);
  console.log(output);
  
  if (shouldPass) {
    if (!output.includes('Verification for') && !output.includes('The following checks were performed on each of these signatures')) {
      throw new Error('Cosign verification should have passed but failed');
    }
  } else {
    if (output.includes('Verification for') || output.includes('The following checks were performed on each of these signatures')) {
      throw new Error('Cosign verification should have failed but passed');
    }
  }
  
  return output;
}

export async function notationGenerateCert() {
  console.log("Generating Notation certificate...");
  await runCommand("notation cert generate-test --default wabbit-networks.io");
}

export async function notationSign(artifact: string) {
  console.log(`Signing artifact with Notation: ${artifact}...`);
  await runCommand(`notation sign -d --allow-referrers-api ${artifact}`);
}
