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
      image: 'busybox',
      needPullFirst: false,
      tag: 'latest'
    });

    await deleteRepo(page, project1, 'busybox');
})