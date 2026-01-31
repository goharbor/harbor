import { test, expect, Page, BrowserContext } from '@playwright/test';
import { execSync } from 'child_process';

// Environment variables
const LOCAL_REGISTRY: string =
    process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string =
    process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip: string = process.env.IP;
const user: string = process.env.HARBOR_ADMIN || 'admin';
const pwd: string = process.env.HARBOR_PASSWORD || 'Harbor12345';
const WEBHOOK_ENDPOINT_UI: string = process.env.WEBHOOK_ENDPOINT_UI || '';

/**
 * Test Case: Tag Retention And Replication Event Type Webhook Functionality By CloudEvents Format
 *
 * Migrated from: tests/robot-cases/Group1-Nightly/Webhook.robot:191
 *
 * This test verifies that webhooks are correctly triggered and received in CloudEvents format
 * for Tag Retention and Replication events.
 */
test.describe('Tag Retention and Replication Webhook - CloudEvents Format', () => {
    test.setTimeout(60 * 60 * 1000); // 60 minutes

    test('should trigger webhooks for tag retention and replication events', async ({
        browser,
    }) => {
        // Test variables
        const d = Date.now().toString();
        const project = `project${d}`;
        const projectDest = `project_dest${d}`;
        const projectPushDest = `project_push_dest${d}`;
        const image = 'busybox';
        const tag1 = 'latest';
        const tag2 = 'stable';
        const payloadFormat = 'CloudEvents';
        const webhookName = `webhook${d}`;
        const endpointName = `e${d}`;
        const replicationRuleName = `rule_push_${d}`;

        // Create browser context for multiple pages
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
        });

        // ============================================
        // Step 1: Setup Webhook Server Tab
        // ============================================
        const webhookPage = await context.newPage();
        await webhookPage.goto(`http://${WEBHOOK_ENDPOINT_UI}`);

        // Get webhook endpoint URL from the page
        const webhookEndpointUrl = await webhookPage
            .locator('//p//code')
            .textContent();
        console.log('Webhook Endpoint URL:', webhookEndpointUrl);
        expect(webhookEndpointUrl).toBeTruthy();

        // ============================================
        // Step 2: Setup Harbor Tab and Login
        // ============================================
        const harborPage = await context.newPage();
        await harborPage.goto('/');

        // Login to Harbor
        await harborPage.getByRole('textbox', { name: 'Username' }).fill(user);
        await harborPage.getByRole('textbox', { name: 'Password' }).fill(pwd);
        await harborPage.getByRole('button', { name: 'LOG IN' }).click();
        await harborPage.waitForTimeout(2000);

        // ============================================
        // Step 3: Create Projects
        // ============================================
        // Create destination project first
        await createProject(harborPage, projectDest);

        // Create main project
        await createProject(harborPage, project);

        // ============================================
        // Step 4: Configure Tag Retention
        // ============================================
        // go into project
        await goIntoProject(harborPage, project);

        // Navigate to Tag Retention
        await harborPage.waitForTimeout(1000);
        await harborPage.getByRole('application').locator('button').click();
        await harborPage.getByText('Policy').click();
        await harborPage.getByRole('button', { name: 'Tag Retention' }).click();
        await harborPage.waitForTimeout(1000);

        // Add a tag retention rule
        await harborPage.getByRole('button', { name: 'ADD RULE' }).click();
        await harborPage.locator('#template').selectOption('always');
        await harborPage
            .getByRole('button', { name: 'ADD', exact: true })
            .click();
        await harborPage.getByRole('button', { name: 'ACTION' }).click();
        await harborPage
            .getByRole('button', { name: 'Edit', exact: true })
            .click();
        await harborPage.locator('#repos').click();
        // await harborPage.locator('#repos').press('ControlOrMeta+a');
        await harborPage.locator('#repos').fill('**');
        await harborPage.locator('#tags').click();
        // await harborPage.locator('#tags').press('ControlOrMeta+a');
        await harborPage.locator('#tags').fill(tag1);
        await harborPage.getByRole('button', { name: 'SAVE' }).click();
        await harborPage.waitForTimeout(1000);

        // Verify rule was created with the tag
        await expect(harborPage.getByRole('listitem')).toContainText(tag1);

        // ============================================
        // Step 5: Configure Webhook
        // ============================================
        // Navigate to Webhooks
        await harborPage.getByRole('application').locator('button').click();
        await harborPage.getByText('Webhooks').click();
        await harborPage.waitForTimeout(1000);

        // Create new webhook
        await harborPage.locator('#new-webhook').click();
        await harborPage.waitForTimeout(500);

        // Fill webhook details
        await harborPage.locator('#name').fill(webhookName);
        await harborPage.locator('#edit_endpoint_url').fill(webhookEndpointUrl);

        // Select CloudEvents format
        await harborPage.locator('#payload-format').selectOption(payloadFormat);

        // Uncheck all events first, then select specific ones
        const eventLabels = harborPage.locator(
            'form .clr-control-inline label.clr-control-label'
        );
        const count = await eventLabels.count();
        for (let i = 0; i < count; i++) {
            await eventLabels.nth(i).click();
        }

        // Check desired events
        await harborPage
            .locator('label.clr-control-label')
            .filter({ hasText: 'Tag retention finished' })
            .click();
        await harborPage
            .locator('label.clr-control-label')
            .filter({ hasText: 'Replication status changed' })
            .click();

        // Save webhook
        await harborPage.locator('#new-webhook-continue').click();
        await harborPage.waitForTimeout(2000);

        // Verify webhook was created
        await expect(harborPage.getByText(webhookName)).toBeVisible();

        // ============================================
        // Step 6: Configure Replication
        // ============================================
        // Navigate to Administration > Registries
        // await harborPage.getByText('Administration').click();
        await harborPage.getByRole('link', { name: 'Registries' }).click();
        // await harborPage.locator('//span[contains(.,"Registries")]').click();
        await harborPage.waitForTimeout(500);

        // Create new endpoint
        await harborPage.getByRole('button', { name: 'New Endpoint' }).click();
        await harborPage.waitForTimeout(500);
        await harborPage
            .getByLabel('New Registry Endpoint')
            .getByText('Provider')
            .click();
        await harborPage.locator('#adapter').selectOption('harbor');
        await harborPage.locator('#destination_name').fill(endpointName);
        await harborPage.locator('#destination_url').fill(`https://${ip}`);
        await harborPage.locator('#destination_access_key').fill(user);
        await harborPage.locator('#destination_password').fill(pwd);

        // Disable certificate verification (for self-signed certs)
        await harborPage.locator('#destination_insecure_checkbox').click();

        await harborPage.getByRole('button', { name: 'OK' }).click();
        await harborPage.waitForTimeout(2000);

        // Navigate to Replications
        await harborPage.getByRole('link', { name: 'Replications' }).click();
        // await harborPage.locator('//span[contains(.,"Replications")]').click();
        await harborPage.waitForTimeout(1000);

        // Create replication rule
        await harborPage.locator('#new_replication_rule_id').click();
        await harborPage.waitForTimeout(500);

        // Fill rule details
        await harborPage.locator('#ruleName').fill(replicationRuleName);

        // Select push-based replication
        await harborPage
            .locator('label.clr-control-label')
            .filter({ hasText: 'Push-based' })
            .click();

        // Select destination endpoint (partial match)
        await harborPage.locator('#dest_registry').selectOption('1: Object');
        await expect(harborPage.locator('#dest_registry')).toContainText(
            endpointName
        );

        // Set filter
        await harborPage.locator('#filter_name').fill(`${project}/*`);

        // Set destination namespace
        await harborPage.locator('#dest_namespace').fill(projectPushDest);

        // Select flattening (default: Flatten 1 Level)
        await harborPage
            .locator('#dest_namespace_replace_count')
            .selectOption('Flatten 1 Level');

        // Save rule
        await harborPage.getByRole('button', { name: 'SAVE' }).click();
        await harborPage.waitForTimeout(2000);

        // ============================================
        // Step 7: Verify Tag Retention Webhook
        // ============================================
        console.log('Starting Tag Retention Webhook Verification...');

        // Clear webhook receiver
        await webhookPage.bringToFront();
        await deleteAllRequests(webhookPage);

        // Push images with two different tags
        pushImageWithTag(ip, user, pwd, project, image, tag1, tag1);
        pushImageWithTag(ip, user, pwd, project, image, tag2, tag2);

        // Go back to Harbor and execute tag retention
        await harborPage.bringToFront();
        await harborPage.goto('/');
        await goIntoProject(harborPage, project);

        // Navigate to Tag Retention
        await harborPage.getByText('Policy').click();
        await harborPage.getByRole('button', { name: 'Tag Retention' }).click();
        await harborPage.waitForTimeout(1000);
        await harborPage.waitForTimeout(1000);

        // Execute run
        await harborPage.locator('#run-now').click();
        await harborPage.waitForTimeout(500);
        await harborPage.locator('#execute-run').click();
        await harborPage.waitForTimeout(5000);

        // Expand to see results
        await harborPage
            .locator('//clr-expandable-animation//button')
            .first()
            .click();
        await harborPage.waitForTimeout(2000);

        // Navigate to webhooks and verify execution
        await harborPage.goto('/');
        await goIntoProject(harborPage, project);
        await harborPage.getByRole('application').locator('button').click();
        await harborPage.getByText('Webhooks').click();
        await harborPage.waitForTimeout(1000);

        // TODO: update this to selector.
        // Select webhook row

        await harborPage.getByRole('gridcell', { name: 'Select' }).click();
        // await expect(harborPage.locator('#datagrid')).toContainText('111');
        // await page.getByRole('gridcell', { name: 'Error' }).click();
        // await expect(harborPage.locator('#datagrid')).toContainText('Error');

        // await expect(page.locator('#clr-dg-row12')).toMatchAriaSnapshot(`- gridcell "Tag retention finished"`);
        // await page.getByRole('gridcell', { name: 'Tag retention finished', exact: true }).click();
        await expect(
            harborPage.getByRole('gridcell', {
                name: 'Tag retention finished',
                exact: true,
            })
        ).toBeVisible();
        // await expect(harborPage.locator('#datagrid')).toContainText('Tag retention finished');
        // await page.getByRole('button', { name: 'Toggle vertical navigation' }).click();

        // await harborPage.getByText(/ WEBHOOK /).click();

        // get and print cells in the row
        const row = harborPage.getByRole('row').filter({ hasText: /WEBHOOK/ });
        const cells = row.getByRole('gridcell');
        const cellCount = await cells.count();

        for (let i = 0; i < cellCount; i++) {
            const cellText = await cells.nth(i).textContent();
            console.log(`Cell ${i}:`, cellText);
        }

        // await harborPage
        //     .locator(
        //         `//clr-dg-row[contains(.,'${webhookName}')]//div[contains(@class,'datagrid-select')]`
        //     )
        //     .click();
        await harborPage.waitForTimeout(1000);

        // Get latest webhook execution ID
        const webhookExecutionId = cells[0];
        console.log('Tag Retention Webhook Execution ID:', webhookExecutionId);
        await harborPage.waitForTimeout(10000);
        await expect(
            harborPage.getByRole('gridcell', {
                name: 'Success',
                exact: true,
            })
        ).toBeVisible();
        // Verify webhook execution status
        // await expect(
        //     harborPage.locator(
        //         `//clr-dg-row[.//clr-dg-cell/a[text()=${webhookExecutionId}]]//clr-dg-cell[3]`
        //     )
        // ).toContainText('Success');
        // await expect(
        //     harborPage.locator(
        //         `//clr-dg-row[.//clr-dg-cell/a[text()=${webhookExecutionId}]]//clr-dg-cell[4]`
        //     )
        // ).toContainText('Tag retention finished');

        // Verify webhook payload on webhook server
        await webhookPage.bringToFront();
        await webhookPage.waitForTimeout(2000);

        // Verify CloudEvents format properties
        await expect(webhookPage.locator('body')).toContainText(
            '"specversion":"1.0"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"type":"harbor.tag_retention.finished"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"datacontenttype":"application/json"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            `"operator":"${user}"`
        );
        await expect(webhookPage.locator('body')).toContainText(
            `"project_name":"${project}"`
        );
        await expect(webhookPage.locator('body')).toContainText(
            `"name_tag":"${image}:${tag2}"`
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"status":"SUCCESS"'
        );
        await expect(webhookPage.locator('body')).toContainText('"total":2');
        await expect(webhookPage.locator('body')).toContainText('"retained":1');

        console.log('Tag Retention Webhook Verification PASSED');

        // ============================================
        // Step 8: Verify Replication Webhook
        // ============================================
        console.log('Starting Replication Webhook Verification...');

        // Clear webhook receiver
        await deleteAllRequests(webhookPage);

        // Go back to Harbor and trigger replication
        await harborPage.bringToFront();
        // await harborPage.getByRole('link', { name: 'Administration' }).click();
        await harborPage.waitForTimeout(500);
        await harborPage.getByRole('link', { name: 'Replications' }).click();
        await harborPage.waitForTimeout(1000);

        // Select rule and replicate

        // await harborPage.getByRole('gridcell', { name: endpointName }).click();
        await harborPage
            .getByRole('gridcell', { name: 'Select' })
            .locator('label')
            .click();
        await harborPage
            .getByRole('radio', { name: 'Select' })
            .setChecked(true);
        // await harborPage
        //     .getByRole('row', { name: `/${endpointName}/` })
        //     .locator('label')
        //     .click();

        await harborPage.getByRole('button', { name: 'Replicate' }).click();
        await harborPage
            .getByRole('button', { name: 'REPLICATE', exact: true })
            .click();

        // await harborPage
        //     .locator(
        //         `//clr-dg-row[contains(.,'${replicationRuleName}')]//div[contains(@class,'datagrid-select')]`
        //     )
        //     .first()
        //     .click();
        // await harborPage.waitForTimeout(500);
        // await harborPage.locator('#replication_exe_id').click();
        // await harborPage.waitForTimeout(500);
        // await harborPage
        //     .locator('//clr-modal//button[contains(.,"REPLICATE")]')
        //     .click();
        await harborPage.waitForTimeout(20000);

        // Verify replication succeeded
        // await expect(
        //     harborPage.locator(
        //         '//hbr-replication//div[contains(@class,"datagrid")]//clr-dg-row[1]'
        //     )
        // ).toContainText('Succeeded');
        await harborPage
            .locator('.execution-select > .refresh-btn > clr-icon')
            .click();

        await expect(
            harborPage.getByRole('gridcell', { name: 'Succeeded' }).first()
        ).toBeVisible();
        // Navigate to project webhooks and verify
        await harborPage.goto('/');
        await goIntoProject(harborPage, project);
        await harborPage.getByText('Webhooks').click();
        await harborPage.waitForTimeout(1000);

        // Select webhook row
        await harborPage.getByRole('gridcell', { name: 'Select' }).click();
        // await harborPage
        //     .locator(
        //         `//clr-dg-row[contains(.,'${webhookName}')]//div[contains(@class,'datagrid-select')]`
        //     )
        //     .click();
        await harborPage.waitForTimeout(1000);

        // Get latest webhook execution ID for replication
        const row1 = harborPage.getByRole('row').filter({ hasText: /WEBHOOK/ });
        const cells1 = row1.getByRole('gridcell');
        const cellCount1 = await cells1.count();

        for (let i = 0; i < cellCount1; i++) {
            const cellText = await cells.nth(i).textContent();
            console.log(`Cell ${i}:`, cellText);
        }

        // await harborPage
        //     .locator(
        //         `//clr-dg-row[contains(.,'${webhookName}')]//div[contains(@class,'datagrid-select')]`
        //     )
        //     .click();
        await harborPage.waitForTimeout(1000);

        // Get latest webhook execution ID
        const replicationWebhookExecutionId = cells1[0];
        console.log(
            'Tag Retention Webhook Execution ID:',
            replicationWebhookExecutionId
        );

        // const replicationWebhookExecutionId = await harborPage
        //     .locator('//clr-dg-row[1]//clr-dg-cell[1]//a')
        //     .textContent();
        // console.log(
        //     'Replication Webhook Execution ID:',
        //     replicationWebhookExecutionId
        // );

        // Verify webhook execution status
        await expect(
            harborPage
                .getByRole('gridcell', {
                    name: 'Success',
                    exact: true,
                })
                .first()
        ).toBeVisible();

        // await expect(
        //     harborPage.locator(
        //         `//clr-dg-row[.//clr-dg-cell/a[text()=${replicationWebhookExecutionId}]]//clr-dg-cell[3]`
        //     )
        // ).toContainText('Success');
        // await expect(
        //     harborPage.locator(
        //         `//clr-dg-row[.//clr-dg-cell/a[text()=${replicationWebhookExecutionId}]]//clr-dg-cell[4]`
        //     )
        // ).toContainText('Replication status changed');

        await expect(
            harborPage.getByRole('gridcell', {
                name: 'Replication status changed',
                exact: true,
            }).first()
        ).toBeVisible();
        // Verify webhook payload on webhook server
        await webhookPage.bringToFront();
        await webhookPage.waitForTimeout(2000);

        // Verify CloudEvents format properties for replication
        await expect(webhookPage.locator('body')).toContainText(
            '"specversion":"1.0"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"type":"harbor.replication.status.changed"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"datacontenttype":"application/json"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            `"operator":"${user}"`
        );
        await expect(webhookPage.locator('body')).toContainText(
            '"trigger_type":"MANUAL"'
        );
        await expect(webhookPage.locator('body')).toContainText(
            `"namespace":"${project}"`
        );

        console.log('Replication Webhook Verification PASSED');

        // ============================================
        // Cleanup
        // ============================================
        await context.close();
    });
});

// ============================================
// Helper Functions
// ============================================

/**
 * Creates a new project in Harbor
 */
async function createProject(page: Page, projectName: string): Promise<void> {
    await page.goto('/');
    await page.getByRole('button', { name: 'New Project' }).click();
    await page.locator('#create_project_name').fill(projectName);
    await page.getByRole('button', { name: 'OK' }).click();
    await page.waitForTimeout(2000);
    console.log(`Created project: ${projectName}`);
}

/**
 * Navigates into a project
 */
async function goIntoProject(page: Page, projectName: string): Promise<void> {
    await page.goto('/');
    await page.getByRole('link', { name: projectName }).click();
    await page.waitForTimeout(1000);
}

/**
 * Deletes all requests on the webhook receiver page
 */
async function deleteAllRequests(page: Page): Promise<void> {
    await page.bringToFront();
    await page.waitForTimeout(1000);
    const deleteButton = page.getByRole('button', {
        name: 'Delete all requests',
    });
    if (await deleteButton.isVisible()) {
        await deleteButton.click();
        await page.waitForTimeout(1000);
    } else {
        console.log('No requests to delete or button not found');
    }
}

/**
 * Runs a shell command and returns the output
 */
function runCommand(command: string): string {
    console.log(`\n$ ${command}`);

    try {
        const output = execSync(command, {
            encoding: 'utf-8',
            stdio: ['pipe', 'pipe', 'pipe'],
        });

        console.log('Command output:\n', output.trim());
        return output.trim();
    } catch (error: any) {
        console.error(`Command failed: ${command}`);
        console.error('STDOUT:', error.stdout?.toString()?.trim() || '');
        console.error('STDERR:', error.stderr?.toString()?.trim() || '');
        throw error;
    }
}

/**
 * Tags and pushes a single image to Harbor
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
    console.log(`\nPushing image ${image}:${tag} to ${project}...`);

    const sourceImage = `${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1}`;
    const targetImage = `${ip}/${project}/${image}:${tag}`;

    // Pull image from local registry
    runCommand(`docker pull ${sourceImage}`);

    // Login to Harbor
    runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);

    // Tag image for Harbor project
    runCommand(`docker tag ${sourceImage} ${targetImage}`);

    // Push image to Harbor
    runCommand(`docker push ${targetImage}`);

    // Logout after push
    runCommand(`docker logout ${ip}`);

    console.log(`Successfully pushed ${targetImage}`);
}
