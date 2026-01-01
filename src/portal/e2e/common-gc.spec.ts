import { expect, Page, test } from '@playwright/test'
import { execSync } from 'child_process';

interface PushImageOptions {
  ip: string;
  user: string;
  pwd: string;
  project: string;
  image: string;
  tag?: string;
  needPullFirst?: boolean;
  localRegistry?: string;
  localNamespace?: string;
}

const harborIp = process.env.HARBOR_BASE_URL?.replace(/^https?:\/\//, '') || 'localhost';
const harborUser = process.env.HARBOR_USERNAME || 'admin';
const harborPassword = process.env.HARBOR_USERNAME || 'Harbor12345';
const localRegistryName = process.env.LOCAL_REGISTRY || 'docker.io';
const localRegistryNamespace = process.env.LOCAL_REGISTRY_NAMESPACE || 'library';

async function loginAsAdmin(page: Page) {
    await page.goto(harborIp);
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
    await page.getByRole('button', { name: 'LOG IN' }).click();
    await expect(page.getByRole('link', { name: 'Projects' })).toBeVisible();
}

async function createProject(page: Page, projectName: string, isPublic: boolean = false) {
    await page.getByRole("link", { name: "Projects" }).click();
    await page.getByRole("button", { name: "NEW PROJECT" }).click();
    await page.locator("//*[@id='create_project_name']").fill(projectName);
    if (isPublic) {
        await page.locator(`xpath=//input[@name='public']/..//label[contains(@class,'clr-control-label')]`).check();
    }
    await page.getByRole('button', { name: 'OK' }).click();
    await expect(page.getByRole('link', {name: projectName})).toBeVisible()
}

async function pushImage(options: PushImageOptions) {
  const {
    ip,
    user,
    pwd,
    project,
    image,
    tag = 'latest',
    needPullFirst = true,
    localRegistry = localRegistryName,
    localNamespace = localRegistryNamespace,
  } = options;

  const imageWithTag = `${image}:${tag}`;
  const sourceImage = `${localRegistry}/${localNamespace}/${imageWithTag}`;
  const targetImage = `${ip}/${project}/${imageWithTag}`;

  try {
    if (needPullFirst) {
      console.log(`Pulling ${sourceImage}...`);
      execSync(`docker pull ${sourceImage}`, { stdio: 'inherit' });
    }
    
    console.log(`Logging in to ${ip}...`);
    execSync(`docker login -u ${user} -p ${pwd} ${ip}`, { stdio: 'inherit' });

    const srcImage = needPullFirst ? sourceImage : imageWithTag;
    execSync(`docker tag ${srcImage} ${targetImage}`, { stdio: 'inherit' });

    console.log(`Pushing ${targetImage}...`);
    execSync(`docker push ${targetImage}`, { stdio: 'inherit' });

    execSync(`docker logout ${ip}`, { stdio: 'inherit' });
  } catch (error) {
    console.error('Docker operation failed:', error);
    throw error;
  }
}

async function goIntoProject(page: Page, projectName: string) {
    await page.getByRole('link', { name: 'Projects' }).click();
    await expect(page.getByRole('link', {name: projectName})).toBeVisible()
    await page.getByRole('link', {name: projectName}).click();
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
        await page.locator('#delete_untagged').click();
    }

    if (gc_now) {
        await page.getByRole("button", { name: 'DRY RUN' }).click()
    } else {
        await page.getByRole('button', { name: 'GC NOW' }).click();
    }
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

    await pushImage({
      ip: harborIp,
      user: harborUser,
      pwd: harborPassword,
      project: project1,
      image: smaller_repo,
      needPullFirst: true,
      tag: smaller_repo_tag,
    });

    const timestamp2 = Date.now();
    const project2 = `project${timestamp2}`;
    console.log(project2);
    await createProject(page, project2);

    await pushImage({
      ip: harborIp,
      user: harborUser,
      pwd: harborPassword,
      project: project2,
      image: larger_repo,
      needPullFirst: true,
      tag: larger_repo_tag,
    });

    await switchToProjectQuotas(page);
    await checkProjectQuotaSorting(page, project1, project2);

    await deleteRepo(page, project1, smaller_repo);
    await deleteRepo(page, project2, larger_repo);
    await runGC(page)
})