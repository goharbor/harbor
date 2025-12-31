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

async function loginAsAdmin(page: Page) {
    await page.goto('/');
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
    localRegistry = 'docker.io',
    localNamespace = 'library'
  } = options;

  const imageWithTag = `${image}:${tag}`;
  const sourceImage = `${localRegistry}/${localNamespace}/${imageWithTag}`;
  const targetImage = `${ip}/${project}/${imageWithTag}`;

  try {
    // Pull from source if needed
    if (needPullFirst) {
      console.log(`Pulling ${sourceImage}...`);
      execSync(`sudo docker pull ${sourceImage}`, { stdio: 'inherit' });
    }

    // Login to Harbor
    console.log(`Logging in to ${ip}...`);
    execSync(`sudo docker login -u ${user} -p ${pwd} ${ip}`, { stdio: 'pipe' });

    // Tag image
    const srcImage = needPullFirst ? sourceImage : imageWithTag;
    execSync(`sudo docker tag ${srcImage} ${targetImage}`, { stdio: 'inherit' });

    // Push to Harbor
    console.log(`Pushing ${targetImage}...`);
    execSync(`sudo docker push ${targetImage}`, { stdio: 'inherit' });

    // Logout
    execSync(`sudo docker logout ${ip}`, { stdio: 'inherit' });
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

  // Ascending (smaller first)
  await storageHeader.click();
  console.log(`smaller project: ${smaller_proj}`);
  console.log(`larger project: ${larger_proj}`);

  
  // smaller proj : in row 1
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[1]//clr-dg-cell[1]//a[contains(text(), '${smaller_proj}')]`)
  ).toBeVisible();
  
  // larger proj : in row 2
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[2]//clr-dg-cell[1]//a[contains(text(), '${larger_proj}')]`)
  ).toBeVisible();

  // Descending (larger first)
  await storageHeader.click();
  
  // larger proj : in row 1
  await expect(
    page.locator(`//div[@class='datagrid-table']//clr-dg-row[1]//clr-dg-cell[1]//a[contains(text(), '${larger_proj}')]`)
  ).toBeVisible();
  
  // smaller proj : in row 2
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

    await pushImage({
      ip: 'localhost:80',
      user: 'admin',
      pwd: 'Harbor12345',
      project: project1,
      image: 'photon',
      needPullFirst: false,
      tag: '2.0'
    });

    const timestamp2 = Date.now();
    const project2 = `project${timestamp2}`;
    console.log(project2);
    await createProject(page, project2);

    await pushImage({
      ip: 'localhost:80',
      user: 'admin',
      pwd: 'Harbor12345',
      project: project2,
      image: 'alpine',
      needPullFirst: false,
      tag: 'latest'
    });

    // alpine < photon
    await switchToProjectQuotas(page);
    await checkProjectQuotaSorting(page, project2, project1);

    await deleteRepo(page, project1, 'photon');
    await deleteRepo(page, project2, 'alpine');
    await runGC(page)
})