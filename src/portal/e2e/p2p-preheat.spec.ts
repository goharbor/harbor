import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

// Environment variables
const LOCAL_REGISTRY: string = process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string = process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip: string = process.env.IP || 'localhost';
const user: string = process.env.HARBOR_ADMIN || 'admin';
const pwd: string = process.env.HARBOR_PASSWORD || 'Harbor12345';
const dragonflyAuthToken: string = process.env.DRAGONFLY_AUTH_TOKEN || '';
const distributionEndpoint: string = process.env.DISTRIBUTION_ENDPOINT || 'https://127.0.0.1';

/**
 * Executes a shell command and returns the output.
 */
function runCommand(command: string): string {
    console.log(`\n$ ${command}`);
    try {
        const output = execSync(command, {
            encoding: 'utf-8',
            stdio: ['pipe', 'pipe', 'pipe'],
        });
        console.log('✅ Command output:\n', output.trim());
        return output.trim();
    } catch (error: any) {
        console.error(`❌ Command failed: ${command}`);
        console.error('--- STDOUT ---\n', error.stdout?.toString()?.trim() || '');
        console.error('--- STDERR ---\n', error.stderr?.toString()?.trim() || '');
        throw error;
    }
}

/**
 * Tags and pushes a single image to Harbor.
 */
function pushImageWithTag(
    ip: string,
    user: string,
    pwd: string,
    project: string,
    image: string,
    tag: string,
    tag1: string = 'latest'
): void {
    console.log(`\n🚀 Running docker push for ${image}...`);

    const sourceImage = `${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1}`;
    const targetImage = `${ip}/${project}/${image}:${tag}`;

    runCommand(`docker pull ${sourceImage}`);
    runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);
    runCommand(`docker tag ${sourceImage} ${targetImage}`);
    runCommand(`docker push ${targetImage}`);
    runCommand(`docker logout ${ip}`);
}

test('Distribution CRUD', async ({ page }) => {
    test.setTimeout(10 * 60 * 1000); // 10 minutes
    const d = Date.now();
    const name = `distribution${d}`;
    const endpoint = 'https://32.1.1.2';
    const endpointNew = 'https://10.65.65.42';

    // Login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(user);
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Navigate to Distributions
    await page.getByRole('link', { name: 'Distributions' }).click();

    // Create Distribution
    await page.getByRole('button', { name: 'NEW INSTANCE' }).click();
    await page.locator('#provider').selectOption('Dragonfly');
    await page.locator('#name').fill(name);
    await page.locator('#endpoint').fill(endpoint);
    if (dragonflyAuthToken) {
        await page.locator('#auth_data_token').fill(dragonflyAuthToken);
    }
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(name)).toBeVisible();

    // Edit Distribution
    await page.getByRole('row', { name: new RegExp(name) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Edit' }).click();
    await page.locator('#endpoint').fill(endpointNew);
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(endpointNew)).toBeVisible();

    // Delete Distribution
    await page.getByRole('row', { name: new RegExp(name) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(name)).not.toBeVisible();

    // Logout
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});

test('P2P Preheat Policy CRUD', async ({ page }) => {
    test.setTimeout(15 * 60 * 1000); // 15 minutes
    const d = Date.now();
    const projectName = `project_p2p${d}`;
    const distName = `distribution${d}`;
    const endpoint = 'https://20.76.1.2';
    const policyName = `policy${d}`;
    const repo = 'alpine';
    const repoNew = 'redis*';
    const tag = 'v1.0';

    // Login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).fill(user);
    await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Create Distribution
    await page.getByRole('link', { name: 'Distributions' }).click();
    await page.getByRole('button', { name: 'NEW INSTANCE' }).click();
    await page.locator('#provider').selectOption('Dragonfly');
    await page.locator('#name').fill(distName);
    await page.locator('#endpoint').fill(endpoint);
    if (dragonflyAuthToken) {
        await page.locator('#auth_data_token').fill(dragonflyAuthToken);
    }
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(distName)).toBeVisible();

    // Create Project
    await page.getByRole('link', { name: 'Projects' }).click();
    await page.getByRole('button', { name: 'New Project' }).click();
    await page.locator('#create_project_name').fill(projectName);
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);

    // Go into project
    await page.getByRole('link', { name: projectName }).click();

    // Create P2P Preheat Policy
    await page.getByRole('link', { name: 'P2P Preheat' }).click();
    await page.getByRole('button', { name: 'NEW POLICY' }).click();
    await page.locator('#provider').selectOption({ label: distName });
    await page.locator('#name').fill(policyName);
    await page.locator('#repo').fill(repo);
    await page.locator('#tag').fill(tag);
    await page.getByRole('button', { name: 'ADD' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(policyName)).toBeVisible();

    // Edit P2P Preheat Policy
    await page.getByRole('row', { name: new RegExp(policyName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Edit' }).click();
    await page.locator('#repo').fill(repoNew);
    await page.getByRole('button', { name: 'SAVE' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(repoNew)).toBeVisible();

    // Try to delete distribution (should fail because policy is using it)
    await page.getByRole('link', { name: 'Distributions' }).click();
    await page.getByRole('row', { name: new RegExp(distName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);
    // Distribution should still exist (deletion blocked by policy)
    await expect(page.getByText(distName)).toBeVisible();

    // Go back and delete policy
    await page.getByRole('link', { name: 'Projects' }).click();
    await page.getByRole('link', { name: projectName }).click();
    await page.getByRole('link', { name: 'P2P Preheat' }).click();
    await page.getByRole('row', { name: new RegExp(policyName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(policyName)).not.toBeVisible();

    // Now delete distribution
    await page.getByRole('link', { name: 'Distributions' }).click();
    await page.getByRole('row', { name: new RegExp(distName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByText(distName)).not.toBeVisible();

    // Logout
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});

test('P2P Preheat By Manual', async ({ page }) => {
    test.skip(!process.env.DISTRIBUTION_ENDPOINT, 'Requires DISTRIBUTION_ENDPOINT env var (need_distribution_endpoint tag)');
    test.setTimeout(30 * 60 * 1000); // 30 minutes

    const d = Date.now();
    const projectName = `project_p2p${d}`;
    const distName = `distribution${d}`;
    const policyName = `policy${d}`;
    const image1 = 'busybox';
    const image2 = 'hello-world';
    const tag1 = 'latest';
    const tag2 = 'stable';

    // Login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).fill(user);
    await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
    await page.getByRole('button', { name: 'LOG IN' }).click();

    // Create Distribution
    await page.getByRole('link', { name: 'Distributions' }).click();
    await page.getByRole('button', { name: 'NEW INSTANCE' }).click();
    await page.locator('#provider').selectOption('Dragonfly');
    await page.locator('#name').fill(distName);
    await page.locator('#endpoint').fill(distributionEndpoint);
    if (dragonflyAuthToken) {
        await page.locator('#auth_data_token').fill(dragonflyAuthToken);
    }
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);

    // Create Project
    await page.getByRole('link', { name: 'Projects' }).click();
    await page.getByRole('button', { name: 'New Project' }).click();
    await page.locator('#create_project_name').fill(projectName);
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);

    // Push images to project
    pushImageWithTag(ip, user, pwd, projectName, image1, tag1, tag1);
    pushImageWithTag(ip, user, pwd, projectName, image1, tag2, tag2);
    pushImageWithTag(ip, user, pwd, projectName, image2, tag1, tag1);

    // Go into project and create policy
    await page.getByRole('link', { name: projectName }).click();
    await page.getByRole('link', { name: 'P2P Preheat' }).click();
    await page.getByRole('button', { name: 'NEW POLICY' }).click();
    await page.locator('#provider').selectOption({ label: distName });
    await page.locator('#name').fill(policyName);
    await page.locator('#repo').fill(image1);
    await page.locator('#tag').fill(tag1);
    await page.getByRole('button', { name: 'ADD' }).click();
    await page.waitForTimeout(2000);

    // Execute P2P Preheat
    await page.getByRole('row', { name: new RegExp(policyName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Execute' }).click();
    await page.getByRole('button', { name: 'CONFIRM' }).click();

    // Wait for execution with retries
    let verified = false;
    for (let i = 0; i < 10; i++) {
        await page.waitForTimeout(5000);
        await page.locator('.refresh-btn').click();
        await page.waitForTimeout(3000);
        
        const successVisible = await page.getByText('Success').isVisible().catch(() => false);
        if (successVisible) {
            verified = true;
            break;
        }
    }
    expect(verified).toBeTruthy();

    // Check that the correct image was preheated
    await expect(page.getByText(`${projectName}/${image1}:${tag1}`)).toBeVisible();

    // Verify: Check that other images were NOT preheated
    await expect(page.getByText(`${projectName}/${image1}:${tag2}`)).not.toBeVisible();
    await expect(page.getByText(`${projectName}/${image2}:${tag1}`)).not.toBeVisible();

    // Cleanup: delete policy
    await page.getByRole('link', { name: 'P2P Preheat' }).click();
    await page.getByRole('row', { name: new RegExp(policyName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);

    // Cleanup: delete distribution
    await page.getByRole('link', { name: 'Distributions' }).click();
    await page.getByRole('row', { name: new RegExp(distName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);

    // Cleanup: delete project
    await page.getByRole('link', { name: 'Projects' }).click();
    await page.getByRole('row', { name: new RegExp(projectName) }).locator('label').click();
    await page.getByRole('button', { name: 'ACTION' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE' }).click();
    await page.waitForTimeout(2000);

    // Logout
    await page.getByRole('button', { name: user, exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});
