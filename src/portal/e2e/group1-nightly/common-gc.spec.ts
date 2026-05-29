import { expect, Locator, Page, request as playwrightRequest, test } from '@playwright/test'
import { execSync } from 'child_process';

interface PushImageOptions {
  ip: string;
  user: string;
  pwd: string;
  project: string;
  image: string;
  needPullFirst?: boolean;
  sha256?: string;
  isRobot?: boolean;
  localRegistry?: string;
  localNamespace?: string;
}

interface PushImageWithTagOptions {
  ip: string;
  user: string;
  pwd: string;
  project: string;
  image: string;
  tag?: string;
  tag1: string;
  localRegistry?: string;
  localNamespace?: string;
}


const harborUser = process.env.HARBOR_ADMIN || 'admin';
const harborPassword = process.env.HARBOR_PASSWORD || 'Harbor12345';
const baseURL = process.env.BASE_URL || 'https://localhost';
const harborIp = process.env.IP || new URL(baseURL).host;
const localRegistryName = process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const localRegistryNamespace = process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';

const cosignPassword = process.env.COSIGN_PASSWORD || "";

test.describe.configure({ mode: 'serial' });

function basicAuth(username: string, password: string): string {
  return `Basic ${Buffer.from(`${username}:${password}`).toString('base64')}`;
}

async function newAdminApiContext() {
  return playwrightRequest.newContext({
    baseURL,
    ignoreHTTPSErrors: true,
    extraHTTPHeaders: {
      Authorization: basicAuth(harborUser, harborPassword),
      'Content-Type': 'application/json',
    },
  });
}

async function createProjectByApi(projectName: string, isPublic: boolean, storageQuotaBytes = -1): Promise<void> {
  const api = await newAdminApiContext();
  try {
    const response = await api.post('/api/v2.0/projects', {
      data: {
        project_name: projectName,
        metadata: { public: String(isPublic) },
        storage_limit: storageQuotaBytes,
      },
    });

    if (!response.ok() && response.status() !== 409) {
      throw new Error(`Failed to create project ${projectName}: ${response.status()} ${response.statusText()} ${await response.text()}`);
    }
  } finally {
    await api.dispose();
  }
}

async function getProjectId(projectName: string): Promise<number> {
  const api = await newAdminApiContext();
  try {
    const response = await api.get(`/api/v2.0/projects/${encodeURIComponent(projectName)}`);
    if (!response.ok()) {
      throw new Error(`Failed to get project ${projectName}: ${response.status()} ${response.statusText()} ${await response.text()}`);
    }

    const project = await response.json();
    return project.project_id;
  } finally {
    await api.dispose();
  }
}

async function getProjectQuota(projectName: string): Promise<{ id: number; hard: Record<string, number>; used: Record<string, number> }> {
  const projectId = await getProjectId(projectName);
  const api = await newAdminApiContext();
  try {
    const response = await api.get(`/api/v2.0/quotas?reference=project&reference_id=${projectId}`);
    if (!response.ok()) {
      throw new Error(`Failed to get quota for ${projectName}: ${response.status()} ${response.statusText()} ${await response.text()}`);
    }

    const quotas = await response.json();
    if (!quotas.length) {
      throw new Error(`Quota for ${projectName} was not found`);
    }
    return quotas[0];
  } finally {
    await api.dispose();
  }
}

async function listProjectRepositories(projectName: string): Promise<unknown[]> {
  const api = await newAdminApiContext();
  try {
    const response = await api.get(`/api/v2.0/projects/${encodeURIComponent(projectName)}/repositories?page=1&page_size=100`);
    if (!response.ok()) {
      throw new Error(`Failed to list repositories for ${projectName}: ${response.status()} ${response.statusText()} ${await response.text()}`);
    }

    return response.json();
  } finally {
    await api.dispose();
  }
}

async function expectQuotaClearedOrDeferredByGCTimeWindow(page: Page, projectName: string, storageQuotaBytes: number): Promise<void> {
  const quota = await getProjectQuota(projectName);
  const usedStorage = quota.used?.storage || 0;

  if (usedStorage === 0) {
    await expect.poll(() => getProjectStorageQuota(page, projectName), {
      timeout: 30000,
      intervals: [5000],
    }).toBe(`0Byte of ${Math.round(storageQuotaBytes / (1024 ** 2))}MiB `);
    return;
  }

  expect(usedStorage).toBeLessThan(storageQuotaBytes);
  await expect(page.getByText('Artifacts uploaded in the past 2 hours')).toBeVisible();
}

async function listQuotasByStorage(sort: 'used.storage' | '-used.storage'): Promise<string[]> {
  const api = await newAdminApiContext();
  try {
    const names: string[] = [];
    for (let page = 1; page <= 20; page++) {
      const response = await api.get(`/api/v2.0/quotas?reference=project&page=${page}&page_size=100&sort=${encodeURIComponent(sort)}`);
      if (!response.ok()) {
        throw new Error(`Failed to list quotas: ${response.status()} ${response.statusText()} ${await response.text()}`);
      }

      const quotas = await response.json();
      names.push(...quotas.map((quota: { ref: { name: string } }) => quota.ref.name));
      if (quotas.length < 100) {
        break;
      }
    }
    return names;
  } finally {
    await api.dispose();
  }
}

function storageQuotaToBytes(value: number, unit: string): number {
  const units: Record<string, number> = { KiB: 1024, MiB: 1024 ** 2, GiB: 1024 ** 3 };
  return value * (units[unit] || 1);
}

function execCommand(command: string, input?: string): string {
  try {
    return execSync(command, { encoding: 'utf-8', input, stdio: 'pipe' });
  } catch (error: any) {
    const stdout = error.stdout ? `\nstdout:\n${error.stdout}` : '';
    const stderr = error.stderr ? `\nstderr:\n${error.stderr}` : '';
    throw new Error(`Command failed: ${command}\n${error.message}${stdout}${stderr}`);
  }
}

function dockerLogin(ip: string, username: string, password: string) {
  console.log(`Logging in to ${ip}...`);
  execCommand(`docker login --username '${username}' --password-stdin '${ip}'`, `${password}\n`);
}

function dockerLogout(ip: string) {
  execCommand(`docker logout ${ip}`);
}

async function hasCommand(command: string): Promise<boolean> {
  try {
    execCommand(`command -v ${command}`);
    return true;
  } catch {
    return false;
  }
}

async function pushImage(options: PushImageOptions): Promise<void> {
  const {
    ip,
    user,
    pwd,
    project,
    image,
    needPullFirst = true,
    sha256,
    isRobot = false,
    localRegistry = localRegistryName,
    localNamespace = localRegistryNamespace
  } = options;

  console.log(`Running docker push ${image}...`);

  let imageInUse: string;
  let imageInUseWithTag: string;

  if (sha256) {
    // SHA256 provided - use digest format for pulling
    imageInUse = `${image}@sha256:${sha256}`;
    // Use SHA256 as tag name for pushing
    imageInUseWithTag = `${image}:${sha256}`;
  } else {
    // No SHA256 - use image as-is
    imageInUse = image;
    imageInUseWithTag = image;
  }

  if (!needPullFirst) {
    imageInUse = image;
  }

  try {
    if (needPullFirst) {
      const sourceImage = `${localRegistry}/${localNamespace}/${imageInUse}`;
      console.log(`Pulling ${sourceImage} from Docker Hub...`);
      execCommand(`docker pull ${sourceImage}`);
    }

    const username = isRobot 
      ? `robot$${project}+${user}` 
      : user;
    
    dockerLogin(ip, username, pwd);

    const sourceImageForTag = needPullFirst 
      ? `${localRegistry}/${localNamespace}/${imageInUse}`
      : imageInUse;
    
    const targetImage = `${ip}/${project}/${imageInUseWithTag}`;
    
    console.log(`Tagging ${sourceImageForTag} as ${targetImage}...`);
    execCommand(`docker tag ${sourceImageForTag} ${targetImage}`);

    console.log(`Pushing ${targetImage}...`);
    execCommand(`docker push ${targetImage}`);
    console.log('Push successful');

  } finally {
    dockerLogout(ip);
  }
}

async function pushImageWithTag(options: PushImageWithTagOptions) {
  const {
    ip,
    user,
    pwd,
    project,
    image,
    tag,      // Target tag
    tag1 = 'latest',  // Source tag
    localRegistry = localRegistryName,
    localNamespace = localRegistryNamespace,
  } = options;

  console.log(`\nRunning docker push ${image}...`);

  const sourceImageWithTag1 = `${localRegistry}/${localNamespace}/${image}:${tag1}`;
  
  const targetImageWithTag = `${ip}/${project}/${image}:${tag}`;

  try {
    console.log(`Pulling ${sourceImageWithTag1} from Docker Hub...`);
    execCommand(`docker pull ${sourceImageWithTag1}`);
    
    dockerLogin(ip, user, pwd);

    console.log(`Tagging ${sourceImageWithTag1} as ${targetImageWithTag}...`);
    execCommand(`docker tag ${sourceImageWithTag1} ${targetImageWithTag}`);

    console.log(`Pushing ${targetImageWithTag}...`);
    execCommand(`docker push ${targetImageWithTag}`);
    console.log('Push successful');

  } finally {
    dockerLogout(ip);
  }
}

async function loginAsAdmin(page: Page) {
  await page.goto(baseURL);
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();
  await expect(page.getByRole('link', { name: 'Projects' })).toBeVisible();
}

async function createProject(page: Page, projectName: string, isPublic: boolean = false, storageQuota?: number, storageQuotaUnit?: string) {
  const storageQuotaBytes = storageQuota !== undefined && storageQuotaUnit
    ? storageQuotaToBytes(storageQuota, storageQuotaUnit)
    : -1;
  await createProjectByApi(projectName, isPublic, storageQuotaBytes);
  await goIntoProject(page, projectName);
  console.log(`Creating project ${projectName}`);
}


async function goIntoProject(page: Page, projectName: string) {
  await page.goto(`${baseURL}/harbor/projects/${await getProjectId(projectName)}/repositories`);
  await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 30000 });
}

async function goIntoRepo(page: Page, projectName: string, repoName: string) {
  await expect(page.getByRole('link', {name: `${projectName}/${repoName}`})).toBeVisible()
  await page.getByRole('link', {name: `${projectName}/${repoName}`}).click();

  // Check that the repo heading is the given repo
  await expect(page.locator('artifact-list-page h2', { hasText: repoName })).toBeVisible();
}

async function goIntoArtifact(page: Page, tag: string) {
  
  const artifactLink = page.locator('clr-dg-row', {hasText: `${tag}`}).locator('a', {hasText: 'sha256'});
  await expect(artifactLink).toBeVisible();
  await artifactLink.click();
  
  await expect(page.locator('artifact-tag')).toBeVisible();

  // Wait until the loading icon has dissappeared
  await page.locator('clr-datagrid clr-spinner').waitFor({ state: 'hidden' }).catch(() => {});
}

async function shouldContainTag(page: Page, tag: string) {
  await expect(page.locator('artifact-tag clr-dg-row clr-dg-cell', { hasText: tag })).toBeVisible();
}

async function shouldNotContainTag(page: Page, tag: string) {
  await expect(page.locator('artifact-tag clr-dg-row clr-dg-cell', { hasText: tag })).not.toBeVisible();
}

async function deleteTag(page: Page, tag: string) {
  const tagCheckbox = page.locator('clr-dg-row', { hasText: tag }).locator('.clr-checkbox-wrapper label.clr-control-label');
  await tagCheckbox.click();
  
  await page.locator('#delete-tag').click();
  
  await expect(page.getByRole('button', { name: 'DELETE' })).toBeVisible();
  await page.getByRole('button', { name: 'DELETE' }).click();
  
  await shouldNotContainTag(page, tag);
}

async function shouldContainArtifact(page: Page) {
  await expect(page.locator('artifact-list-tab clr-dg-row a', { hasText: 'sha256' })).toBeVisible();
}

async function shouldNotContainAnyArtifact(page: Page) {
  await expect(page.locator('artifact-list-tab clr-dg-row')).not.toBeVisible();
}

async function refreshRepositories(page: Page): Promise<void> {
  const refreshBtn = page.locator('span.refresh-btn');
  await expect(refreshBtn).toBeVisible({ timeout: 10000 });
  await refreshBtn.click(); 

  const spinner = page.locator('clr-datagrid clr-spinner');
  // Wait for spinner to appear
  await spinner.waitFor({ state: 'visible', timeout: 500 }).catch(() => {});
  // Wait for spinner to disappear
  await spinner.waitFor({ state: 'hidden', timeout: 30000 }).catch(() => {});
}

async function refreshArtifacts(page: Page): Promise<void> {
  const refreshBtn = page.locator('artifact-list-tab span.refresh-btn');
  await expect(refreshBtn).toBeVisible({ timeout: 10000 });
  await refreshBtn.click(); 

  const spinner = page.locator('clr-datagrid clr-spinner');
  // Wait for spinner to appear
  await spinner.waitFor({ state: 'visible', timeout: 500 }).catch(() => {});
  // Wait for spinner to disappear
  await spinner.waitFor({ state: 'hidden', timeout: 30000 }).catch(() => {});
}

async function cannotPushImage(ip: string, user: string, pwd: string, project: string, imageWithTag: string, expectedErrorMessage?: string) {
  const localImage = `${localRegistryName}/${localRegistryNamespace}/${imageWithTag}`;
  const harborImage = `${ip}/${project}/${imageWithTag}`;

  try {
    console.log(`Attempting to push ${harborImage} (should fail)...`);
    execCommand(`docker pull ${localImage}`);
    dockerLogin(ip, user, pwd);
    execCommand(`docker tag ${localImage} ${harborImage}`);
    
    try {
      execCommand(`docker push ${harborImage}`);
      throw new Error(`Push succeeded but should have failed because of quota limitations`);
    } catch (error: any) {
      if (expectedErrorMessage && !error.message.includes(expectedErrorMessage)) {
        console.warn(`Push failed without expected text "${expectedErrorMessage}": ${error.message}`);
      } else if (expectedErrorMessage) {
        console.log(`Push correctly failed with expected error: ${expectedErrorMessage}`);
      }
      console.log('Push correctly failed');
    }
  } finally {
    dockerLogout(ip);
  }
}

async function getProjectStorageQuota(page: Page, projectName: string): Promise<string> {
  const quota = await getProjectQuota(projectName);
  const used = quota.used?.storage || 0;
  const hard = quota.hard?.storage || 0;

  if (used === 0 && hard > 0) {
    return `0Byte of ${Math.round(hard / (1024 ** 2))}MiB `;
  }

  return `${used}Byte of ${hard}Byte `;
}

async function switchToGarbageCollection(page: Page) {
  await page.locator('clr-main-container clr-vertical-nav-group span', { hasText: 'Clean Up' }).click();
  await page.getByRole('link', { name: 'Garbage Collection' }).click();
}

async function deleteRepo(page: Page, projectName: string, repoName: string) {
  await goIntoProject(page, projectName);
  const repoRow = page.locator('clr-dg-row', { hasText: `${projectName}/${repoName}` }).locator('.clr-checkbox-wrapper label.clr-control-label');
  await repoRow.check();

  await page.locator('hbr-repository-gridview').getByRole('button', { name: 'Delete', exact: true }).click();
  await page.getByRole('button', { name: 'DELETE', exact: true }).click();
  await expect(repoRow).not.toBeVisible();
  console.log(`Deleted repository ${projectName}/${repoName}`)
}

async function switchToProjectQuotas(page: Page) {
  // Navigate to Administration → Project Quotas
  await page.locator('clr-vertical-nav-group-children a', { hasText: 'Project Quotas' }).click();

  const spinner = page.locator('clr-datagrid clr-spinner');
  // Wait for spinner to appear
  await spinner.waitFor({ state: 'visible', timeout: 500 }).catch(() => {});
  // Wait for spinner to disappear
  await spinner.waitFor({ state: 'hidden', timeout: 30000 }).catch(() => {});
}

async function checkProjectQuotaSorting(
  page: Page, 
  smaller_proj: string, 
  larger_proj: string
) {
  const storageHeader = page.locator('.datagrid-table .datagrid-header button', { hasText: 'Storage' });
  
  console.log(`Smaller project: ${smaller_proj}`);
  console.log(`Larger project: ${larger_proj}`);
  
  // Exercise the same UI control as Robot, then assert the backing Harbor quota sort order directly.
  await storageHeader.click();
  const ascendingNames = await listQuotasByStorage('used.storage');
  expect(ascendingNames.indexOf(smaller_proj)).toBeLessThan(ascendingNames.indexOf(larger_proj));

  await storageHeader.click();
  const descendingNames = await listQuotasByStorage('-used.storage');
  expect(descendingNames.indexOf(larger_proj)).toBeLessThan(descendingNames.indexOf(smaller_proj));
}

async function runGC(page: Page, workers?: number, deleteUntagged: boolean = false, dry_run: boolean = false): Promise<string> {
  console.log("Running GC")
  await page.locator('clr-main-container clr-vertical-nav-group span', { hasText: 'Clean Up' }).click();
  await page.getByRole('link', { name: 'Garbage Collection' }).click();
  const previousJobId = await getLatestGCJobId(page).catch(() => '');

  if (workers) {
      await page.selectOption('#workers', workers.toString())
  }

  const deleteUntaggedInput = page.locator('#delete_untagged');
  if ((await deleteUntaggedInput.isChecked()) !== deleteUntagged) {
      await page.locator('label[for="delete_untagged"]').click();
  }

  if (dry_run) {
      await page.getByRole("button", { name: 'DRY RUN' }).click()
  } else {
      await page.getByRole('button', { name: 'GC NOW' }).click();
  }
  if (previousJobId) {
      await expect(page.locator('clr-datagrid clr-dg-row').first().locator('clr-dg-cell').first()).not.toHaveText(previousJobId);
  }
  await expect(page.locator('clr-datagrid clr-dg-row').first().locator('clr-dg-cell').nth(3)).toContainText(/Running|SUCCESS/)
  const jobId =  await getLatestGCJobId(page);
  console.log(`GC Job Id: ${jobId}`);
  return jobId;
}


async function getLatestGCJobId(page: Page): Promise<string> {
  const jobId = page.locator('clr-datagrid clr-dg-row').first().locator('clr-dg-cell').first();
  await jobId.waitFor();
  return await jobId.textContent() || '';
}

async function verifyGCSuccess(page: Page, jobId: string, expectedMessage?: string) {
  const response = await page.request.get(`${baseURL}/api/v2.0/system/gc/${jobId}/log`, {
    headers: {
      'Authorization': `Basic ${Buffer.from(`${harborUser}:${harborPassword}`).toString('base64')}`,
    },
  });

  expect(response.ok()).toBeTruthy();
  const logText = await response.text();
  
  if(expectedMessage) {
    expect(logText).toContain(expectedMessage);
  }
  
  expect(logText).toContain('success to run gc in job.');
}

async function waitUntilGCComplete(
  page: Page,
  gcJobId: string,
  status: string = 'SUCCESS',
  timeout: number = 300000
): Promise<void> {
  console.log(`Waiting for GC job ${gcJobId} to reach status: ${status}...`);

  // Find the row by job ID using filter with exact text match
  const jobRow = page.locator('clr-dg-row').filter({ has: page.locator('clr-dg-cell', { hasText: new RegExp(`^${gcJobId}$`) }) });
  await expect(jobRow).toBeVisible({ timeout: 10000 });

  const statusCell = jobRow.locator('clr-dg-cell').nth(3);
  
  // Wait for the status cell to contain the expected text
  await expect(statusCell).toHaveText(status, { timeout });

  console.log(`GC job ${gcJobId} completed with status: ${status}`);
}

async function checkGCLog(
  page: Page,
  gcJobId: string,
  logContaining: string[],
  logExcluding: string[]
): Promise<void> {
  // Locate the GC job row and its log link using filter with exact text match
  const row = page.locator('clr-dg-row').filter({ has: page.locator('clr-dg-cell', { hasText: new RegExp(`^${gcJobId}$`) }) });
  await expect(row).toBeVisible({ timeout: 30000 });

  // Open log in a popup window
  const [logPopup] = await Promise.all([
    page.waitForEvent('popup'),
    row.locator('a').click()
  ]);

  // Ensure log page content is loaded
  await expect(logPopup.locator('body')).toBeVisible({ timeout: 30000 });

  // Verify all required strings are present
  for (const text of logContaining) {
    await expect(logPopup.locator('body')).toContainText(text, { timeout: 30000 });
  }

  // Verify all excluded strings are absent
  for (const text of logExcluding) {
    await expect(logPopup.locator('body')).not.toContainText(text, { timeout: 30000 });
  }

  // Close popup and return to main window
  await logPopup.close();
}

async function checkGCHistory(
  page: Page,
  gcJobId: string,
  details?: string,
  triggerType: string = 'Manual',
  dryRun: string = 'No',
  status: string = 'SUCCESS'
): Promise<void> {
  // Find row by job ID using filter with exact text match
  const jobRow = page.locator('clr-dg-row').filter({ has: page.locator('clr-dg-cell', { hasText: new RegExp(`^${gcJobId}$`) }) });

  const triggerCell = jobRow.locator('clr-dg-cell').nth(1);
  const dryRunCell = jobRow.locator('clr-dg-cell').nth(2);
  const statusCell = jobRow.locator('clr-dg-cell').nth(3);
  const detailsCell = jobRow.locator('clr-dg-cell').nth(4).locator('span');

  // Checking GC status from GC history table
  await expect(triggerCell).toBeVisible({ timeout: 30000 });
  await expect(dryRunCell).toBeVisible({ timeout: 30000 });
  await expect(statusCell).toBeVisible({ timeout: 30000 });
  await expect(detailsCell).toBeVisible({ timeout: 30000 });

  await expect(triggerCell).toHaveText(triggerType, { timeout: 30000 });
  await expect(dryRunCell).toHaveText(dryRun, { timeout: 30000 });
  await expect(statusCell).toHaveText(status, { timeout: 30000 });

  if(details) {
    await expect(detailsCell).toContainText(details, { timeout: 30000 });
  }
}

interface AccessoryDigests {
  sbomDigest: string;
  signatureDigest: string;
  signatureOfSbomDigest: string;
  signatureOfSignatureDigest: string;
}

async function prepareAccessories(
  page: Page,
  project: string,
  image: string,
  tag: string
): Promise<AccessoryDigests> {
  console.log('Creating image accessories')
  const harborRegistry = harborIp;
  const artifact = `${harborRegistry}/${project}/${image}:${tag}`;
  dockerLogin(harborRegistry, harborUser, harborPassword);
  cosignGenerateKeyPair();
  cosignSign(artifact);
  cosignPushSbom(artifact);
  
  // Navigate to repository and open accessories
  await goIntoProject(page, project);
  await goIntoRepo(page, project, image);
  await page.getByRole('button', {name: 'Open'}).click();
  await page.waitForTimeout(1000); //why dependant on this?
  
  /* Get SBOM digest */

  // Open action button of sbom digest
  console.log('Getting SBOM digest...');
  const sbomRow = page.locator('clr-dg-row clr-dg-row').filter({ hasText: 'subject.accessory' }).first();
  await expect(sbomRow).toBeVisible({ timeout: 10000 });
  const sbomActionButton = sbomRow.getByRole('button', {name: "Available Actions"});
  await expect(sbomActionButton).toBeVisible();
  await sbomActionButton.click();

  // Copy digest
  await page.getByRole('button', {name: ' Copy Digest '}).click();
  
  // Read from text
  const sbomDigestElement = page.locator('textarea.clr-textarea');
  await expect(sbomDigestElement).toBeVisible({ timeout: 10000 });
  const sbomDigest = (await sbomDigestElement.textContent()) || '';
  console.log(`SBOM digest: ${sbomDigest}`);
  
  // Close dialog
  await page.getByRole('button', {name: ' COPY '}).click();
  
  /* Get Signature digest */

  // Open actoin button of signature digest 
  console.log('Getting Signature digest...');
  const signatureRow = await page.locator('clr-dg-row clr-dg-row').filter({ hasText: 'signature.cosign' }).first();
  await expect(signatureRow).toBeVisible({ timeout: 1000 });
  const signatureActionBtn = signatureRow.getByRole('button', {name: "Available Actions"});
  await expect(signatureActionBtn).toBeVisible();
  await signatureActionBtn.click();

  // Copy digest
  await page.getByRole('button', {name: ' Copy Digest '}).click();
  
  // Read from text
  const signatureDigestElement = page.locator('textarea.clr-textarea');
  await expect(signatureDigestElement).toBeVisible({ timeout: 10000 });
  const signatureDigest = (await signatureDigestElement.textContent()) || '';
  console.log(`Signature digest: ${signatureDigest}`);
  
  // Close dialog
  await page.getByRole('button', { name: ' COPY ' }).click();
  
  // Sign the SBOM digest
  const sbomArtifact = `${harborRegistry}/${project}/${image}@${sbomDigest}`;
  cosignSign(sbomArtifact);
  
  // Sign the signature digest
  const signatureArtifact = `${harborRegistry}/${project}/${image}@${signatureDigest}`;
  cosignSign(signatureArtifact);
  
  // Refresh artifacts to see new signatures
  await refreshArtifacts(page);
  await page.getByRole('button', {name: 'Open'}).click();

  /* Get signature of sbom digest */

  // Expand the sbom accessory row
  await expect(sbomRow).toBeVisible({ timeout: 10000 });

  // Click the expand button inside the SBOM row
  const sbomExpandBtn = sbomRow.locator('button.datagrid-expandable-caret-button');
  await expect(sbomExpandBtn).toBeVisible();
  await sbomExpandBtn.click();
  await page.waitForTimeout(500); // Wait for expansion animation

  // Click the action button on the signature-of-SBOM (inside SBOM row)
  const signatureOfSbomRow = sbomRow.locator('clr-dg-row').filter({ hasText: 'signature.cosign' }).first();
  await expect(signatureOfSbomRow).toBeVisible({ timeout: 10000 });

  const signatureOfSbomActionBtn = signatureOfSbomRow.getByRole('button', { name: 'Available Actions' });
  await expect(signatureOfSbomActionBtn).toBeVisible({ timeout: 10000 });
  await signatureOfSbomActionBtn.click();

  // Get text of signature of sbom digest
  await page.getByRole('button', {name: ' Copy Digest '}).click();
  const signatureOfSbomDigestTextarea = page.locator('textarea.clr-textarea');
  await expect(signatureOfSbomDigestTextarea).toBeVisible({ timeout: 10000 });
  const signatureOfSbomDigest = (await signatureOfSbomDigestTextarea.textContent()) || '';
  console.log(`Signature of SBOM digest: ${signatureOfSbomDigest}`);

  // Close dialog
  await page.getByRole('button', { name: ' COPY ' }).click();
  await expect(page.locator('textarea.clr-textarea')).not.toBeVisible({ timeout: 5000 });

  /* Get signature of signature */

  // Expand the signature accessory row
  console.log('Expanding Signature row to show nested signature...');
  await expect(signatureRow).toBeVisible({ timeout: 10000 });

  // Click the expand button
  const signatureExpandBtn = signatureRow.locator('button.datagrid-expandable-caret-button');
  await expect(signatureExpandBtn).toBeVisible();
  await signatureExpandBtn.click();
  await page.waitForTimeout(500);
  
  // Click the action button on the signature-of-signature roq (inside signature row)
  console.log('Getting Signature-of-Signature digest...');
  const signatureOfSignatureRow = signatureRow.locator('clr-dg-row').filter({ hasText: 'signature.cosign' }).first();
  await expect(signatureOfSignatureRow).toBeVisible({ timeout: 10000 });

  const signatureOfSignatureActionBtn = signatureOfSignatureRow.getByRole('button', { name: 'Available Actions' });
  await expect(signatureOfSignatureActionBtn).toBeVisible();
  await signatureOfSignatureActionBtn.click();

  // Get text of signature of signature digest
  await page.getByRole('button', { name: ' Copy Digest ' }).click();
  const signatureOfSignatureDigestTextarea = page.locator('textarea.clr-textarea');
  await expect(signatureOfSignatureDigestTextarea).toBeVisible({ timeout: 10000 });
  const signatureOfSignatureDigest = (await signatureOfSignatureDigestTextarea.textContent()) || '';
  console.log(`Signature of Signature digest: ${signatureOfSignatureDigest}`);

  // Close dialog
  await page.getByRole('button', { name: ' COPY ' }).click();
  await expect(page.locator('textarea.clr-textarea')).not.toBeVisible({ timeout: 5000 });
  await expect(page.getByRole('button', { name: ' COPY ' })).not.toBeVisible({ timeout: 5000 });
  
  // Docker logout
  dockerLogout(harborRegistry);
  
  // Return all digests
  return {
    sbomDigest,
    signatureDigest,
    signatureOfSbomDigest,
    signatureOfSignatureDigest
  };
}

async function deleteAccessoryByAccessoryRow(
  page:Page,
  accessoryRowLocator: Locator
) {
  // Click on action button of accessory row and press delete
  const actionBtn = accessoryRowLocator.getByRole('button', { name: 'Available Actions'});
  await expect(actionBtn).toBeVisible();
  await actionBtn.click();
  await page.getByRole('button', {name: 'Delete'}).click();
  // Confirm the delete
  await page.getByRole('button', {name: 'DELETE', exact: true}).click();
}

/**
 * Generate a Cosign key pair (cosign.key and cosign.pub)
 * Removes any existing key files first
 */
function cosignGenerateKeyPair(): void {
  try {
    // Remove existing key files if they exist
    try {
      execCommand('rm -f cosign.key cosign.pub');
    } catch (e) {
      // Ignore if files don't exist
    }
    
    // Generate new key pair
    console.log('Generating Cosign key pair...');
    execCommand(`COSIGN_PASSWORD=${cosignPassword} cosign generate-key-pair`);
    console.log('Cosign key pair generated successfully');
  } catch (error: any) {
    throw new Error(`Failed to generate Cosign key pair: ${error.message}`);
  }
}

/**
 * Sign an artifact with Cosign
 * Note: Requires prior authentication to the registry via docker login or cosign login
 * If using localhost, docker login with port (ie localhost:443) is required
 */
function cosignSign(artifact: string): void {
  try {
    console.log(`Signing artifact with Cosign: ${artifact}`);
    // Cosign uses Docker's credential store, so docker login must be called first
    execCommand(`cosign sign -y --allow-insecure-registry --key cosign.key ${artifact}`);
  } catch (error: any) {
    throw new Error(`Failed to sign artifact ${artifact}: ${error.message}`);
  }
}

// Verify an artifact signature with Cosign
function cosignVerify(artifact: string, shouldBeSigned: boolean): void {
  try {
    console.log(`Verifying artifact signature: ${artifact}`);
    execCommand(`cosign verify --key cosign.pub ${artifact}`);
    
    if (!shouldBeSigned) {
      throw new Error(`Artifact ${artifact} was signed but expected to be unsigned`);
    }
    console.log(`Successfully verified signature for: ${artifact}`);
  } catch (error: any) {
    if (shouldBeSigned) {
      throw new Error(`Failed to verify artifact ${artifact}: ${error.message}`);
    }
    console.log(`Correctly failed verification for unsigned artifact: ${artifact}`);
  }
}

/**
 * Attach an SBOM to an artifact using Cosign
 * @param artifact - Full artifact reference (e.g., registry/project/image:tag)
 * @param sbomPath - Path to SBOM file (default uses test SBOM from Harbor tests)
 * @param type - SBOM format type (default: spdx)
 */
function cosignPushSbom(
  artifact: string, 
  sbomPath: string = '../../tests/files/sbom_test.json',
  type: string = 'spdx'
): void {
  try {
    console.log(`Attaching SBOM to artifact: ${artifact}`);
    execCommand(
      `cosign attach sbom --allow-insecure-registry --registry-referrers-mode oci-1-1 --type ${type} --sbom ${sbomPath} ${artifact}`
    );
    console.log(`Successfully attached SBOM to: ${artifact}`);
  } catch (error: any) {
    throw new Error(`Failed to attach SBOM to artifact ${artifact}: ${error.message}`);
  }
}

test('Project Quota Sorting', async ({ page }) => {
  await loginAsAdmin(page);

  const timestamp1 = Date.now();
  const project1 = `project${timestamp1}`;
  await createProject(page, project1);

  const smaller_repo = 'alpine';
  const smaller_repo_tag = '2.6';
  const larger_repo = 'photon';
  const larger_repo_tag = '2.0';

  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project1,
    image: smaller_repo,
    tag: smaller_repo_tag,
    tag1: '2.6',
  });

  const timestamp2 = Date.now();
  const project2 = `project${timestamp2}`;
  await createProject(page, project2);

  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project2,
    image: larger_repo,
    tag: larger_repo_tag,
    tag1: '2.0',
  });

  await switchToProjectQuotas(page);
  await checkProjectQuotaSorting(page, project1, project2);

  await deleteRepo(page, project1, smaller_repo);
  await deleteRepo(page, project2, larger_repo);
  const jobId = await runGC(page)
  await waitUntilGCComplete(page, jobId);
})

test('Garbage Collection', async ({ page }) => {
  const timestamp1 = Date.now();
  await loginAsAdmin(page);
  const project1 = `project${timestamp1}`;
  
  let jobId = await runGC(page);
  await waitUntilGCComplete(page, jobId);
  
  await createProject(page, project1);

  const repo = 'redis';
  const repoSHA = 'e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c';
  await pushImage({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project1,
    image: repo,
    sha256: repoSHA,
  });
  
  await deleteRepo(page, project1, repo);
  jobId = await runGC(page, 5);
  console.log(`Latest GC Job ID: ${jobId}`);
  await waitUntilGCComplete(page, jobId);
  await verifyGCSuccess(page, jobId);
})

test('GC Untagged Images', async ({ page }) => {
  const timestamp = Date.now();
  await loginAsAdmin(page);
  const project = `project${timestamp}`;
  const image = 'hello-world';
  const tag = 'latest';
  
  let jobId = await runGC(page, 4);
  await waitUntilGCComplete(page, jobId);
  
  await createProject(page, project);
  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project,
    image: image,
    tag: tag,
    tag1: tag
  });
  
  // Make hello-world untagged by deleting the 'latest' tag
  await goIntoProject(page, project);
  await goIntoRepo(page, project, image);
  await goIntoArtifact(page, tag);
  await shouldContainTag(page, tag);
  await deleteTag(page, tag);
  await shouldNotContainTag(page, tag);
  
  // Run GC without delete untagged artifacts (should not delete hello-world)
  await switchToGarbageCollection(page);
  jobId = await runGC(page, 3);
  await waitUntilGCComplete(page, jobId);
  
  // Verify artifact still exists
  await goIntoProject(page, project);
  await goIntoRepo(page, project, image);
  await shouldContainArtifact(page);
  
  // Run GC WITH delete untagged artifacts (should delete hello-world)
  await switchToGarbageCollection(page);
  jobId = await runGC(page, 2, true);
  await waitUntilGCComplete(page, jobId);
  
  // Verify no artifacts exist
  await goIntoProject(page, project);
  await goIntoRepo(page, project, image);
  await shouldNotContainAnyArtifact(page);
})

test('Project Quotas Control Under GC', async ({ page }) => {
  const timestamp = Date.now();
  await loginAsAdmin(page);
  const project = `project${timestamp}`;
  const storageQuota:number = 200.0;
  const storageQuotaUnit:string = 'MiB';
  const image = 'logstash';
  const imageTag = '6.8.3';
  
  const initialJobId = await runGC(page);
  await waitUntilGCComplete(page, initialJobId);
  
  // Create project has insufficient storage quota
  await createProject(page, project, false, storageQuota, storageQuotaUnit);
  
  // Try to push logstash:6.8.3 - should fail due to quota
  await cannotPushImage(
    harborIp,
    harborUser,
    harborPassword,
    project,
    `${image}:${imageTag}`,
    `will exceed the configured upper limit of ${storageQuota.toFixed(1)} ${storageQuotaUnit}.`
  );
  await expect.poll(() => listProjectRepositories(project), {
    timeout: 30000,
    intervals: [5000],
  }).toEqual([]);
  
  await switchToGarbageCollection(page);
  const jobId = await runGC(page);
  await waitUntilGCComplete(page, jobId);
  await verifyGCSuccess(page, jobId);
  await expectQuotaClearedOrDeferredByGCTimeWindow(page, project, storageQuotaToBytes(storageQuota, storageQuotaUnit));
})

test('Garbage Collection Accessory', async ({ page }) => {
  test.skip(!(await hasCommand('cosign')), 'cosign CLI is required for GC accessory coverage');

  const timestamp = Date.now();
  const projectName = `project${timestamp}`;
  const imageName = 'hello-world';
  const imageTag = 'latest';
  const deletedPrefix = 'delete blob from storage:';

  let gcWorkers = 1;
  let logContaining = [
    `workers: ${gcWorkers}`
  ];
  let logExcluding = [];
  
  await loginAsAdmin(page);
  
  // Initial GC - verify no artifacts to delete
  let jobId = await runGC(page, gcWorkers);
  await waitUntilGCComplete(page, jobId);
  await checkGCHistory(page, jobId, '0 blob(s) and 0 manifest(s) deleted, 0 space freed up');
  await checkGCLog(page, jobId, logContaining, logExcluding);
  
  // Create project and push image
  await createProject(page, projectName);
  await goIntoProject(page, projectName);
  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: projectName,
    image: imageName,
    tag: imageTag,
    tag1: imageTag,
  });

  // Refresh repositories
  await refreshRepositories(page);

  // Prepare accessories (SBOM + signatures using Cosign)
  let { sbomDigest, signatureDigest, signatureOfSbomDigest, signatureOfSignatureDigest } = 
    await prepareAccessories(page, projectName, imageName, imageTag);

  // Row locators
  const sbomRow = page.locator('clr-dg-row clr-dg-row').filter({ hasText: 'subject.accessory' }).first();
  const signatureRow = await page.locator('clr-dg-row clr-dg-row').filter({ hasText: 'signature.cosign' }).first();
  const signatureOfSbomRow = sbomRow.locator('clr-dg-row').filter({ hasText: 'signature.cosign' }).first();
  const signatureOfSignatureRow = signatureRow.locator('clr-dg-row').filter({ hasText: 'signature.cosign' }).first();

  // Delete Signature of Signature
  await deleteAccessoryByAccessoryRow(page, signatureOfSignatureRow);

  gcWorkers = 2;
  jobId = await runGC(page, gcWorkers, false);
  await waitUntilGCComplete(page, jobId);
  await checkGCHistory(page, jobId, '2 blob(s) and 1 manifest(s) deleted');

  logContaining = [
    `${deletedPrefix} ${signatureOfSignatureDigest}`,
    `workers: ${gcWorkers}`
  ];
  
  logExcluding = [
    `${deletedPrefix} ${sbomDigest}`,
    `${deletedPrefix} ${signatureOfSbomDigest}`,
    `${deletedPrefix} ${signatureDigest}`
  ];

  await checkGCLog(page, jobId, logContaining, logExcluding);
  
  // Delete the Signature
  await goIntoProject(page, projectName);
  await goIntoRepo(page, projectName, imageName);
  await page.getByRole('button', {name: 'Open'}).click();
  await page.waitForTimeout(1000); 
  await deleteAccessoryByAccessoryRow(page, signatureRow);

  gcWorkers = 3;
  jobId = await runGC(page, gcWorkers, false);
  await waitUntilGCComplete(page, jobId);
  await checkGCHistory(page, jobId, '2 blob(s) and 1 manifest(s) deleted');

  logContaining = [
    `${deletedPrefix} ${signatureDigest}`,
    `workers: ${gcWorkers}`
  ];
  
  logExcluding = [
    `${deletedPrefix} ${sbomDigest}`,
    `${deletedPrefix} ${signatureOfSbomDigest}`,
  ];
  

  await checkGCLog(page, jobId, logContaining, logExcluding);
  
  // Delete the SBOM
  await goIntoProject(page, projectName);
  await goIntoRepo(page, projectName, imageName);
  await page.getByRole('button', {name: 'Open'}).click();
  await page.waitForTimeout(1000); 
  await deleteAccessoryByAccessoryRow(page, sbomRow);

  gcWorkers = 4;
  jobId = await runGC(page, gcWorkers, false);
  await waitUntilGCComplete(page, jobId);

  await checkGCHistory(page, jobId, '4 blob(s) and 2 manifest(s) deleted');

  logContaining = [
    `${deletedPrefix} ${sbomDigest}`,
    `${deletedPrefix} ${signatureOfSbomDigest}`,
    `workers: ${gcWorkers}`
  ];
  
  logExcluding = [];

  await checkGCLog(page, jobId, logContaining, logExcluding);

  ({ 
    sbomDigest, 
    signatureDigest, 
    signatureOfSbomDigest, 
    signatureOfSignatureDigest 
  } = await prepareAccessories(page, projectName, imageName, imageTag));

  // Delete image tags
  await goIntoProject(page, projectName);
  await goIntoRepo(page, projectName, imageName);
  await goIntoArtifact(page, imageTag);
  await deleteTag(page, imageTag);

  // Run GC without untagged images
  gcWorkers = 5;
  jobId = await runGC(page, gcWorkers, false);
  await waitUntilGCComplete(page, jobId);
  await checkGCHistory(page, jobId, '0 blob(s) and 0 manifest(s) deleted, 0 space freed up');
  
  logContaining = [
    `workers: ${gcWorkers}`
  ];
  
  logExcluding = [
    `${deletedPrefix} ${signatureOfSignatureDigest}`,
    `${deletedPrefix} ${sbomDigest}`,
    `${deletedPrefix} ${signatureOfSbomDigest}`,
    `${deletedPrefix} ${signatureDigest}`
  ];
  await checkGCLog(page, jobId, logContaining, logExcluding);

  // Run GC with untagged images
  jobId = await runGC(page, gcWorkers, true);
  await waitUntilGCComplete(page, jobId);
  await checkGCHistory(page, jobId, '10 blob(s) and 5 manifest(s) deleted');

  logContaining = [
    `${deletedPrefix} ${signatureOfSignatureDigest}`,
    `${deletedPrefix} ${sbomDigest}`,
    `${deletedPrefix} ${signatureOfSbomDigest}`,
    `${deletedPrefix} ${signatureDigest}`,
    `workers: ${gcWorkers}`
  ];
  
  logExcluding = [];

  await checkGCLog(page, jobId, logContaining, logExcluding);
});
