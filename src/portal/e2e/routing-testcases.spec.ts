import { test, expect, Page, Locator } from '@playwright/test';

async function Login(page: Page) {
    // Login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');

    await page.getByRole('button', { name: 'LOG IN' }).click();
}

test('Main Menu Routing', async ({ page }) => {
    // override the default timeout for this test, As it expects more time than the default one.
    test.setTimeout(60_000);

    // Login with admin credentials
    await Login(page);

    // Page Checks Dictionary, Path as a Key and a element in that path as a Value
    const pageChecks: Record<string, (page: Page) => Locator> = {
        'harbor/projects': (page: Page) => page.locator('projects div h2:has-text("Projects")'),
        'harbor/logs': (page: Page) => page.locator('app-logs h2:has-text("Logs")'),
        'harbor/users': (page: Page) => page.locator('harbor-user div h2:has-text("Users")'),
        'harbor/robot-accounts': (page: Page) => page.locator('system-robot-accounts h2:has-text("Robot Accounts")'),
        'harbor/registries': (page: Page) => page.locator('hbr-endpoint h2:has-text("Registries")'),
        'harbor/replications': (page: Page) => page.locator('total-replication h2:has-text("Replications")'),
        'harbor/distribution/instances': (page: Page) => page.locator('dist-instances div h2:has-text("Instances")'),
        'harbor/labels': (page: Page) => page.locator('app-labels h2:has-text("Labels")'),
        'harbor/project-quotas': (page: Page) => page.locator('app-project-quotas h2:has-text("Project Quotas")'),
        'harbor/interrogation-services/scanners': (page: Page) => page.locator('config-scanner div h4:has-text("Image Scanners")'),
        'harbor/interrogation-services/vulnerability': (page: Page) => page.locator('vulnerability-config div button:has-text("SCAN NOW")'),
        'harbor/interrogation-services/security-hub': (page: Page) => page.locator('h1:has-text("Vulnerabilities")'),
        'harbor/clearing-job/gc': (page: Page) => page.locator('gc-history h5:has-text("GC History")'),
        'harbor/clearing-job/audit-log-purge': (page: Page) => page.locator('app-purge-history  h5:has-text("Purge History")'),
        'harbor/job-service-dashboard/pending-jobs': (page: Page) => page.locator('app-pending-job-list button span:has-text("STOP")'),
        'harbor/job-service-dashboard/schedules': (page: Page) => page.locator('app-schedule-list clr-dg-cell:has-text("SYSTEM_ARTIFACT_CLEANUP")'),
        'harbor/job-service-dashboard/workers': (page: Page) => page.locator('app-worker-list button span:has-text("Free")'),
        'harbor/configs/auth': (page: Page) => page.locator('config config-auth label:has-text("Auth Mode")'),
        'harbor/configs/security': (page: Page) => page.locator('config app-security span:has-text("CVE allowlist")'),
        'harbor/configs/setting': (page: Page) => page.locator('config system-settings label:has-text("Project Creation")'),
    };

    // Iterate through the dictionary and expect the locators to be visible
    for (const [path, locatorFn] of Object.entries(pageChecks)) {
        await page.goto(`/${path}`);
        const element = locatorFn(page);
        await expect(element).toBeVisible();
    }
});