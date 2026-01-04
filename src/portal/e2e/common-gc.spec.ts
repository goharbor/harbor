import { expect, Page, test } from '@playwright/test'
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


const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
const harborUser = process.env.HARBOR_USERNAME || 'admin';
const harborPassword = process.env.HARBOR_USERNAME || 'Harbor12345';
const localRegistryName = process.env.LOCAL_REGISTRY || 'docker.io';
const localRegistryNamespace = process.env.LOCAL_REGISTRY_NAMESPACE || 'library';


function execCommand(command: string): string {
  try {
    return execSync(command, { encoding: 'utf-8', stdio: 'pipe' });
  } catch (error: any) {
    throw new Error(`Command failed: ${command}\n${error.message}`);
  }
}

function dockerLogin(ip: string, username: string, password: string) {
  console.log(`Logging in to ${ip}...`);
  execCommand(`docker login -u '${username}' -p '${password}' ${ip}`);
}

function dockerLogout(ip: string) {
  execCommand(`docker logout ${ip}`);
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
    localRegistry = 'docker.io',
    localNamespace = 'library'
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
  await page.goto(harborIp);
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();
  await expect(page.getByRole('link', { name: 'Projects' })).toBeVisible();
}

async function createProject(page: Page, projectName: string, isPublic: boolean = false, storageQuota?: number, storageQuotaUnit?: string) {
  await page.getByRole("link", { name: "Projects" }).click();
  await page.getByRole("button", { name: "NEW PROJECT" }).click();
  await page.locator("//*[@id='create_project_name']").fill(projectName);
  
  if (isPublic) {
      await page.locator(`xpath=//input[@name='public']/..//label[contains(@class,'clr-control-label')]`).check();
  }
  
  if (storageQuota !== undefined && storageQuotaUnit) {
      // Enable storage quota
      await page.locator("xpath=//*[@id='create_project_storage_limit']").fill(storageQuota.toString());
      await page.locator("xpath=//*[@id='create_project_storage_limit_unit']").selectOption(storageQuotaUnit);
  }
  
  await page.getByRole('button', { name: 'OK' }).click();
  await expect(page.getByRole('link', {name: projectName})).toBeVisible()
}


async function goIntoProject(page: Page, projectName: string) {
  await page.getByRole('link', { name: 'Projects' }).click();
  await expect(page.getByRole('link', {name: projectName})).toBeVisible()
  await page.getByRole('link', {name: projectName}).click();
}

async function goIntoRepo(page: Page, projectName: string, repoName: string) {
  
  await expect(page.getByRole('link', {name: `${projectName}/${repoName}`})).toBeVisible()
  await page.getByRole('link', {name: `${projectName}/${repoName}`}).click();
  
  await expect(page.locator(`xpath=//artifact-list-page//h2[contains(., '${repoName}')]`)).toBeVisible();
}

async function goIntoArtifact(page: Page, tag: string) {
  await page.locator('xpath=//clr-datagrid//clr-spinner').waitFor({ state: 'hidden' }).catch((() => {}));
  
  const artifactLink = page.locator(`xpath=//clr-dg-row[contains(.,'${tag}')]//a[contains(.,'sha256')]`);
  await expect(artifactLink).toBeVisible();
  await artifactLink.click();
  
  await expect(page.locator('xpath=//artifact-tag')).toBeVisible();
  await page.locator('xpath=//clr-datagrid//clr-spinner').waitFor({ state: 'hidden' }).catch(() => {});
}

async function shouldContainTag(page: Page, tag: string) {
  await expect(page.locator(`xpath=//artifact-tag//clr-dg-row//clr-dg-cell[contains(.,'${tag}')]`)).toBeVisible();
}

async function shouldNotContainTag(page: Page, tag: string) {
  await expect(page.locator(`xpath=//artifact-tag//clr-dg-row//clr-dg-cell[contains(.,'${tag}')]`)).not.toBeVisible();
}

async function deleteTag(page: Page, tag: string) {
  const tagCheckbox = page.locator(`xpath=//clr-dg-row[contains(.,'${tag}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]`);
  await tagCheckbox.click();
  
  await page.locator('xpath=//*[@id="delete-tag"]').click();
  
  await expect(page.getByRole('button', { name: 'DELETE' })).toBeVisible();
  await page.getByRole('button', { name: 'DELETE' }).click();
  
  await shouldNotContainTag(page, tag);
}

async function shouldContainArtifact(page: Page) {
  await expect(page.locator('xpath=//artifact-list-tab//clr-dg-row//a[contains(.,"sha256")]')).toBeVisible();
}

async function shouldNotContainAnyArtifact(page: Page) {
  await expect(page.locator('xpath=//artifact-list-tab//clr-dg-row')).not.toBeVisible();
}

async function cannotPushImage(ip: string, user: string, pwd: string, project: string, imageWithTag: string, expectedErrorMessage: string) {
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
      // Verify the error message contains the expected text
      if (!error.message.includes(expectedErrorMessage)) {
        throw new Error(`Expected error message to contain "${expectedErrorMessage}", but got: ${error.message}`);
      }
      console.log(`Push correctly failed with expected error: ${expectedErrorMessage}`);
    }
  } finally {
    dockerLogout(ip);
  }
}

async function getProjectStorageQuota(page: Page, projectName: string): Promise<string> {
  await switchToProjectQuotas(page);
  
  const quotaCell = page.locator(`xpath=//project-quotas//clr-datagrid//clr-dg-row[contains(.,'${projectName}')]//clr-dg-cell[3]//label`);
  await quotaCell.waitFor();
  return await quotaCell.textContent() || '';
}

async function switchToGarbageCollection(page: Page) {
  await page.locator("//clr-main-container//clr-vertical-nav-group//span[contains(.,'Clean Up')]").click();
  await page.getByRole('link', { name: 'Garbage Collection' }).click();
}

async function deleteRepo(page: Page, projectName: string, repoName: string) {
  await goIntoProject(page, projectName);
  const repoRow = page.locator(` xpath=//clr-dg-row[contains(.,'${projectName}/${repoName}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]`);
  await repoRow.check();

  await page.getByRole('button', { name: 'DELETE' }).click();
  await page.locator("xpath=//button[contains(.,'DELETE')]").click();
  await expect(repoRow).not.toBeVisible();
}

async function switchToProjectQuotas(page: Page) {
  // Navigate to Administration → Project Quotas
  await page.locator("//clr-vertical-nav-group-children/a[contains(.,'Project Quotas')]").click();
  await page.waitForTimeout(1000);
}

async function checkProjectQuotaSorting(
  page: Page, 
  smaller_proj: string, 
  larger_proj: string
) {
  const storageHeader = page.locator(
    "//div[@class='datagrid-table']//div[@class='datagrid-header']//button[normalize-space()='Storage']"
  );
  
  console.log(`smaller project: ${smaller_proj}`);
  console.log(`larger project: ${larger_proj}`);
  
  // Ascending (smaller first)
  await storageHeader.click();
  
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[1]//clr-dg-cell[1]//a[contains(text(), '${smaller_proj}')]`)
  ).toBeVisible();
  
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[2]//clr-dg-cell[1]//a[contains(text(), '${larger_proj}')]`)
  ).toBeVisible();

  // Descending (larger first)
  await storageHeader.click();
  
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[1]//clr-dg-cell[1]//a[contains(text(), '${larger_proj}')]`)
  ).toBeVisible();
  
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[2]//clr-dg-cell[1]//a[contains(text(), '${smaller_proj}')]`)
  ).toBeVisible();
}

async function runGC(page: Page, workers?: number, deleteUntagged: boolean = false, gc_now: boolean = false) {
  await page.locator(" //clr-main-container//clr-vertical-nav-group//span[contains(.,'Clean Up')]").click();
  await page.getByRole('link', { name: 'Garbage Collection' }).click();

  if (workers) {
      await page.selectOption('#workers', workers.toString())
  }

  if (deleteUntagged) {
      await page.locator('label[for="delete_untagged"]').click();
  }

  if (gc_now) {
      await page.getByRole("button", { name: 'DRY RUN' }).click()
  } else {
      await page.getByRole('button', { name: 'GC NOW' }).click();
  }
}


async function getLatestGCJobId(page: Page): Promise<string> {
  const jobId = page.locator('//clr-datagrid//div//clr-dg-row[1]//clr-dg-cell[1]');
  await jobId.waitFor();
  return await jobId.textContent() || '';
}

async function verifyGCSuccess(page: Page, jobId: string, expectedMessage: string) {
  const response = await page.request.get(`https://${harborIp}/api/v2.0/system/gc/${jobId}/log`, {
    headers: {
      'Authorization': `Basic ${Buffer.from(`${harborUser}:${harborPassword}`).toString('base64')}`,
    },
  });

  expect(response.ok()).toBeTruthy();
  const logText = await response.text();
  
  expect(logText).toContain(expectedMessage);
  expect(logText).toContain('success to run gc in job.');
}

async function waitForGCToComplete(page: Page, jobId: string, timeoutMs: number = 120000) {
  // Poll the GC job log until it shows completion
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeoutMs) {
    try {
      const response = await page.request.get(`https://${harborIp}/api/v2.0/system/gc/${jobId}/log`, {
        headers: {
          'Authorization': `Basic ${Buffer.from(`${harborUser}:${harborPassword}`).toString('base64')}`,
        },
      });

      if (response.ok()) {
        const logText = await response.text();
        if (logText.includes('success to run gc in job.')) {
          return;
        }
      }
    } catch (e) {
      // Ignore errors and retry
    }
    await page.waitForTimeout(2000);
  }
  
  throw new Error(`GC job ${jobId} did not complete within ${timeoutMs}ms`);
}

test('Project Quota Sorting', async ({ page }) => {
  await loginAsAdmin(page);

  const timestamp1 = Date.now();
  const project1 = `project${timestamp1}`;
  console.log(project1);
  await createProject(page, project1);

  const smaller_repo = 'alpine';
  const smaller_repo_tag = 'latest';
  const larger_repo = 'photon';
  const larger_repo_tag = 'latest';

  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project1,
    image: smaller_repo,
    tag: smaller_repo_tag,
    tag1: 'latest',
  });

  const timestamp2 = Date.now();
  const project2 = `project${timestamp2}`;
  console.log(project2);
  await createProject(page, project2);

  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project2,
    image: larger_repo,
    tag: larger_repo_tag,
    tag1: 'latest'
  });

  await switchToProjectQuotas(page);
  await checkProjectQuotaSorting(page, project1, project2);

  await deleteRepo(page, project1, smaller_repo);
  await deleteRepo(page, project2, larger_repo);
  await runGC(page)
})

test('Garbage Collection', async ({ page }) => {
  const timestamp1 = Date.now();
  await loginAsAdmin(page);
  const project1 = `project${timestamp1}`;
  
  await runGC(page);
  
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
  await runGC(page, 5);
  const latestJobId = await getLatestGCJobId(page);
  console.log(`Latest GC Job ID: ${latestJobId}`);
  await waitForGCToComplete(page, latestJobId);
  await verifyGCSuccess(page, latestJobId, '7 blobs and 1 manifests eligible for deletion');
  await verifyGCSuccess(page, latestJobId, 'The GC job actual frees up 34 MB space');
})

test('GC Untagged Images', async ({ page }) => {
  const timestamp = Date.now();
  await loginAsAdmin(page);
  const project = `project${timestamp}`;
  
  await runGC(page, 4);
  
  await createProject(page, project);
  await pushImageWithTag({
    ip: harborIp,
    user: harborUser,
    pwd: harborPassword,
    project: project,
    image: 'hello-world',
    tag: 'latest',
    tag1: 'latest'
  });
  
  // Make hello-world untagged by deleting the 'latest' tag
  await goIntoProject(page, project);
  await goIntoRepo(page, project, 'hello-world');
  await goIntoArtifact(page, 'latest');
  await shouldContainTag(page, 'latest');
  await deleteTag(page, 'latest');
  await shouldNotContainTag(page, 'latest');
  
  // Run GC without delete untagged artifacts (should not delete hello-world)
  await switchToGarbageCollection(page);
  await runGC(page, 3);
  let jobId = await getLatestGCJobId(page);
  await waitForGCToComplete(page, jobId);
  
  // Verify artifact still exists
  await goIntoProject(page, project);
  await goIntoRepo(page, project, 'hello-world');
  await shouldContainArtifact(page);
  
  // Run GC WITH delete untagged artifacts (should delete hello-world)
  await switchToGarbageCollection(page);
  await runGC(page, 2, true);
  jobId = await getLatestGCJobId(page);
  await waitForGCToComplete(page, jobId);
  
  // Verify no artifacts exist
  await goIntoProject(page, project);
  await goIntoRepo(page, project, 'hello-world');
  await shouldNotContainAnyArtifact(page);
})

test('Project Quotas Control Under GC', async ({ page }) => {
  const timestamp = Date.now();
  await loginAsAdmin(page);
  const project = `project${timestamp}`;
  const storageQuota:number = 20.0;
  const storageQuotaUnit:string = 'MiB';
  const image = 'redis';
  const imageTag = '8.4.0';
  
  await runGC(page);
  
  // Create project has insufficient storage quota
  await createProject(page, project, true, storageQuota, storageQuotaUnit);
  
  // Try to push redis:8.4.0 - should fail due to quota
  await cannotPushImage(
    harborIp,
    harborUser,
    harborPassword,
    project,
    `${image}:${imageTag}`,
    `will exceed the configured upper limit of ${storageQuota.toFixed(1)} ${storageQuotaUnit}.`
  );
  
  // Run GC multiple times until quota shows 0 Byte
  const expectedQuota = `0Byte of ${storageQuota}${storageQuotaUnit} `;
  let quotaMatches = false;
  
  for (let i = 0; i < 10; i++) {
    console.log(`GC iteration ${i + 1}/10`);
    
    await switchToGarbageCollection(page);
    await runGC(page);
    const jobId = await getLatestGCJobId(page);
    await waitForGCToComplete(page, jobId);
    
    const actualQuota = await getProjectStorageQuota(page, project);
    console.log(`Quota check: expected="${expectedQuota}", actual="${actualQuota}"`);
    
    if (actualQuota === expectedQuota) {
      quotaMatches = true;
      break;
    }
    
    await page.waitForTimeout(5000);
  }
  
  expect(quotaMatches).toBeTruthy();
})