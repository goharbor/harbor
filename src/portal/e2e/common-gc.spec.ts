import { expect, Page, test } from '@playwright/test'

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

    await deleteRepo(page, project1, 'busybox');
})