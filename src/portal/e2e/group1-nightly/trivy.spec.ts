import {
    expect,
    request as playwrightRequest,
    test,
    type APIRequestContext,
    type Locator,
    type Page,
} from '@playwright/test';
import {
    execFileSync,
    type ExecFileSyncOptionsWithStringEncoding,
} from 'child_process';
import {
    createServer,
    type IncomingMessage,
    type Server,
    type ServerResponse,
} from 'http';

const LOCAL_REGISTRY: string =
    process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string =
    process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip = process.env.IP || '';
const user = process.env.HARBOR_ADMIN || 'admin';
const pwd =
    process.env.HARBOR_PASSWORD || 'Harbor12345';
const scannerPort = Number(process.env.SCANNER_PORT || 8081);
const scannerEndpoint = process.env.SCANNER_ENDPOINT || `http://${ip}:${scannerPort}`;
const stepTimeout = 30 * 1000;
const scanResultTimeout = stepTimeout;
const fakeScannerReports = new Map<string, FakeScanArtifact>();

let fakeScannerServer: Server | undefined;

test.beforeAll(async () => {
    if (process.env.SCANNER_ENDPOINT) {
        return;
    }

    fakeScannerServer = await startFakeScannerAdapter(scannerPort);
});

test.afterAll(async () => {
    if (!fakeScannerServer) {
        return;
    }

    await new Promise<void>((resolve, reject) => {
        fakeScannerServer?.close(error => error ? reject(error) : resolve());
    });
});

function log(message: string): void {
    console.log(`[trivy] ${new Date().toISOString()} ${message}`);
}

async function step<T>(
    name: string,
    action: () => Promise<T>,
    timeout = stepTimeout
): Promise<T> {
    return test.step(name, async () => {
        log(`START ${name}`);
        try {
            const result = await action();
            log(`DONE ${name}`);
            return result;
        } catch (error) {
            log(`FAIL ${name}: ${(error as Error).message}`);
            throw error;
        }
    }, { timeout });
}

async function retryStep<T>(
    name: string,
    action: () => Promise<T>,
    attempts = 3
): Promise<T> {
    let lastError: unknown;
    const deadline = Date.now() + stepTimeout;
    for (let attempt = 1; attempt <= attempts; attempt += 1) {
        const remaining = deadline - Date.now();
        if (remaining <= 0) {
            break;
        }
        try {
            log(`RETRY ${name}: attempt ${attempt}/${attempts}`);
            return await step(`${name} attempt ${attempt}`, action, remaining);
        } catch (error) {
            lastError = error;
            if (attempt < attempts) {
                const delay = Math.min(attempt * 1000, Math.max(0, deadline - Date.now()));
                await new Promise(resolve => setTimeout(resolve, delay));
            }
        }
    }
    throw lastError || new Error(`${name} did not complete within ${stepTimeout / 1000}s`);
}

test('Test Case - Get Harbor Version', async ({ request }) => {
    const systemInfo = await getSystemInfoFromAPI(request, user, pwd);
    expect(systemInfo.harbor_version).toBeTruthy();
    log(`Harbor version: ${systemInfo.harbor_version}`);
});

test('Test Case - Trivy Is Default Scanner And It Is Immutable', async ({ page }) => {
    await login(page);
    await openScannersPage(page);
    const trivyRow = page.getByRole('row').filter({ hasText: 'Trivy' });
    await expect(trivyRow.filter({ hasText: 'Default' })).toBeVisible();

    await trivyRow.locator('label.clr-control-label').click();
    await openActionMenu(page);
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE', exact: true }).click();
    await expect(page.getByText(/default scanner can not be deleted|immutable/i)).toBeVisible();
});

test('Test Case - Disable Scan Schedule', async ({ page }) => {
    await login(page);
    await openVulnerabilityPage(page);

    await page.getByRole('button', { name: 'EDIT' }).click();
    await page.getByRole('combobox').selectOption('None');
    await page.getByRole('button', { name: 'SAVE' }).click();

    await logout(page);
    await login(page);
    await openVulnerabilityPage(page);
    await expect(page.getByText('None', { exact: true })).toBeVisible();
});

test('Test Case - Security Hub', async ({
    page,
    request,
}) => {
    test.setTimeout(180 * 1000);
    expect(ip, 'IP must be set to the Harbor registry host').toBeTruthy();

    const tag = 'v2.2.0';
    const highSeverity = 'High';
    const indexRepo = 'index1';
    const images = [
        'goharbor/harbor-log-base',
        'goharbor/harbor-prepare-base',
        'goharbor/harbor-redis-base',
        'goharbor/harbor-nginx-base',
        'goharbor/harbor-registry-base',
    ];
    const project = `aproject-${Date.now()}`;

    log(`start project=${project} registry=${ip}`);
    await login(page);
    log('logged in');
    await createProject(page, project);
    log(`created project ${project}`);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    log(`project id=${projectId}`);

    for (const image of images) {
        log(`push ${image}:${tag} to ${project}`);
        pushImageWithTag(ip, user, pwd, project, image, tag, tag);
        log(`scan ${project}/${image}`);
        await scanRepository(page, projectId, project, image);
        log(`scanned ${project}/${image}`);
    }

    log(`push manifest ${project}/${indexRepo}:${tag}`);
    pushManifestList(ip, user, pwd, `${ip}/${project}/${indexRepo}:${tag}`, [
        `${ip}/${project}/${images[0]}:${tag}`,
        `${ip}/${project}/${images[1]}:${tag}`,
    ]);

    for (const image of images.slice(0, 2)) {
        log(`delete source repo ${project}/${image}`);
        await deleteRepository(page, projectId, project, image);
    }

    log(`scan manifest repo ${project}/${indexRepo}`);
    await scanRepository(page, projectId, project, indexRepo);
    log('open Security Hub');
    await openSecurityHub(page);

    log('fetch vulnerability summary API');
    const summary = await getVulnerabilitySummaryFromAPI(request, user, pwd);
    expect(summary.dangerous_artifacts.length).toBeGreaterThanOrEqual(2);
    expect(summary.dangerous_cves.length).toBeGreaterThanOrEqual(1);

    const dangerousCve = summary.dangerous_cves[0];
    log('fetch vulnerabilities');
    const vulnerabilities = await getVulnerabilitiesFromAPI(
        request,
        user,
        pwd,
        [`project_id=${projectId}`, `tag=${tag}`]
    );
    const filterVulnerability = getCompleteVulnerability(vulnerabilities);

    const digest = filterVulnerability.digest;
    const cveId = filterVulnerability.cve_id;
    const packageName = filterVulnerability.package;
    const cvssScore = String(filterVulnerability.cvss_v3_score);
    const vulnerabilitySeverity = filterVulnerability.severity;

    await assertSummaryCounts(page, summary);
    await assertDangerousArtifacts(page, summary.dangerous_artifacts);
    await assertDangerousCVEs(page, summary.dangerous_cves);

    await assertQuickSearchByArtifact(page, summary.dangerous_artifacts[0]);
    await assertQuickSearchByCve(page, dangerousCve.cve_id);

    await searchByTextFilter(page, 'project_id', project);
    await expectGridCellVisible(page, project);

    await searchByTextFilter(
        page,
        'repository_name',
        `${project}/${images[2]}`
    );
    await expectGridCellVisible(page, `${project}/${images[2]}`);

    await searchByTextFilter(page, 'digest', digest);
    await expectGridCellVisible(page, shortDigest(digest));

    await searchByTextFilter(page, 'cve_id', cveId);
    await expectGridCellVisible(page, cveId);

    await searchByTextFilter(page, 'package', packageName);
    await expectGridCellVisible(page, packageName);

    await searchByTextFilter(page, 'tag', tag);
    await expectGridCellVisible(page, tag);

    await searchByCvssScore(page, cvssScore);
    await expectGridCellVisible(page, cvssScore);

    await assertSeverityFilter(page, 'Critical', summary.critical_cnt);
    await assertSeverityFilter(page, highSeverity, summary.high_cnt);
    await assertSeverityFilter(page, 'Medium', summary.medium_cnt);
    await assertSeverityFilter(page, 'Low', summary.low_cnt);

    await assertSeverityFilter(page, 'Unknown', summary.unknown_cnt);
    await assertSeverityFilter(page, 'None', 0);

    await searchByAllFilters(page, {
        project,
        repositoryName: filterVulnerability.repository_name,
        digest,
        cveId,
        packageName,
        tag,
        cvssScore,
        severity: vulnerabilitySeverity,
    });
    await expectGridCellVisible(page, vulnerabilitySeverity);

    await page.getByRole('button', { name: 'Open' }).first().click();
    if (filterVulnerability.desc) {
        await expect(page.locator('clr-datagrid')).toContainText(
            filterVulnerability.desc
        );
    }

    await deleteRepository(page, projectId, project, indexRepo);
    await openSecurityHub(page);
    await expect(vulnerabilitySummary(page)).not.toContainText(
        `${project}/${indexRepo}`,
        { timeout: scanResultTimeout }
    );

    await searchByTextFilter(
        page,
        'repository_name',
        `${project}/${indexRepo}`
    );
    await expectNoVulnerabilities(page);

    await logout(page);
    log('done');
});

test('Test Case - Manual Scan All', async ({ page, request }) => {
    test.setTimeout(150 * 1000);
    const project = 'library';
    const image = 'redis';
    const digest = 'e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c';

    await pushImage(ip, user, pwd, project, image, digest);
    await login(page);

    await openVulnerabilityPage(page);
    await triggerScanNowAndWait(page);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await scanResultShouldDisplayInListRow(page, digest);
    await page.getByRole('link', { name: 'sha256' }).click();
    await viewRepoScanDetails(page, ['Critical']);
});

test('Test Case - Scan A Tag In The Repo', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'vmware/photon';
    const tag = '1.0';

    await login(page);
    await createProject(page, project);
    await pushImage(ip, user, pwd, project, `${image}:${tag}`);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await scanRepository(page, projectId, project, image);
    await scanResultShouldDisplayInListRow(page, tag);
    pullImage(ip, user, pwd, project, image, tag);
});

test('Test Case - Scan As An Unprivileged User', async ({ page, request }) => {
    const suffix = Date.now();
    const project = `project-${suffix}`;
    const image = 'hello-world';
    const unprivilegedUser = `user${suffix}`;
    const unprivilegedPassword = 'Test1@01';

    await login(page);
    await createProject(page, project, true);
    await pushImage(ip, user, pwd, project, image);
    await logout(page);

    await createUser(page, unprivilegedUser, unprivilegedPassword);
    await loginAs(page, unprivilegedUser, unprivilegedPassword);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await page
        .getByRole('gridcell', { name: 'Select Select' })
        .locator('label')
        .click();
    await page.getByRole('checkbox', { name: 'Select', exact: true }).check();
    await expect(page.getByRole('button', { name: 'Scan vulnerability' })).toBeDisabled();
});

test('Test Case - Scan Image With Empty Vul', async ({ page, request }) => {
    test.setTimeout(90 * 1000);
    const project = 'library';
    const image = 'photon';
    const tag = '4.0_scan';

    await pushImage(ip, user, pwd, project, `${image}:${tag}`);
    await login(page);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await scanRepository(page, projectId, project, image);
    await scanResultShouldDisplayAnyInListRow(page, tag);
});

test('Test Case - Scan Image On Push', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'memcached';
    const tag = 'latest';

    await login(page);
    await createProject(page, project, true);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await openProjectConfig(page);
    await enableScanOnPush(page);

    await pushImage(ip, user, pwd, project, image);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await scanResultShouldDisplayInListRow(page, tag);
    await page.getByRole('link', { name: 'sha256' }).click();
    await viewRepoScanDetails(page, ['Critical', 'High']);
});

test('Test Case - View Scan Results', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'tomcat';
    const tag = 'latest';
    const signinUser = `user${Date.now()}`;
    const signinPassword = 'Test1@34';

    await createUserFromAPI(request, signinUser, signinPassword);
    await loginAs(page, signinUser, signinPassword);
    await createProject(page, project);
    await pushImage(ip, signinUser, signinPassword, project, image);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await scanRepository(page, projectId, project, image);
    await scanResultShouldDisplayInListRow(page, tag);
    await page.getByRole('link', { name: 'sha256' }).click();
    await viewRepoScanDetails(page, ['Critical']);
});

test('Test Case - Project Level Image Serverity Policy', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'redis';
    const digest = '0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a';

    await login(page);
    await createProject(page, project);
    await pushImage(ip, user, pwd, project, image, digest);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await scanRepository(page, projectId, project, image);
    await scanResultShouldDisplayInListRow(page, digest);
    await openProject(page, projectId);
    await setPreventVulnerabilityPolicy(page, 3);
    cannotPullImage(ip, user, pwd, project, image, digest, 'allowlist');
});

test('Test Case - Verfiy System Level CVE Allowlist', async ({ page, request }) => {
    test.setTimeout(90 * 1000);
    const project = `project-${Date.now()}`;
    const image = 'goharbor/harbor-portal';
    const digest = '55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d';
    const signinUser = `user${Date.now()}`;
    const signinPassword = 'Test1@34';

    await createUserFromAPI(request, signinUser, signinPassword);
    await resetSystemCveAllowlistFromAPI();
    await loginAs(page, signinUser, signinPassword);
    await createProject(page, project);
    await pushImage(ip, signinUser, signinPassword, project, image, digest);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await setPreventVulnerabilityPolicy(page, 2);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    const cveAllowlist = await getCveAllowlistDataFromAPI(request, user, pwd, project, image, digest);

    await logout(page, signinUser);
    await login(page);
    await openConfigurationSecurity(page);
    await addSystemCveAllowlist(page, cveAllowlist.partial);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await addSystemCveAllowlist(page, cveAllowlist.last);
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openConfigurationSecurity(page);
    await setCveAllowlistExpires(page, true);
    await expect(page.getByText(/system CVE allowlist has expired/i)).toBeVisible();
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await setCveAllowlistExpires(page, false);
    await expect(page.getByText(/system CVE allowlist has expired/i)).not.toBeVisible();
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openConfigurationSecurity(page);
    await deleteSystemCveAllowlistItem(page, cveAllowlist.last);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
});

test('Test Case - Verfiy Project Level CVE Allowlist', async ({ page, request }) => {
    test.setTimeout(90 * 1000);
    const project = `project-${Date.now()}`;
    const image = 'goharbor/harbor-portal';
    const digest = '55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d';
    const signinUser = `user${Date.now()}`;
    const signinPassword = 'Test1@34';

    await createUserFromAPI(request, signinUser, signinPassword);
    await loginAs(page, signinUser, signinPassword);
    await createProject(page, project);
    await pushImage(ip, signinUser, signinPassword, project, image, digest);
    pullImage(ip, signinUser, signinPassword, project, image, digest);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await setPreventVulnerabilityPolicy(page, 2);
    await useProjectLevelCveAllowlist(page);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    const cveAllowlist = await getCveAllowlistDataFromAPI(request, user, pwd, project, image, digest);
    await openProject(page, projectId);
    await addProjectCveAllowlist(page, cveAllowlist.partial);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await addProjectCveAllowlist(page, cveAllowlist.last);
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openProjectConfig(page);
    await setCveAllowlistExpires(page, true);
    await expect(page.getByText(/project CVE allowlist has expired/i)).toBeVisible();
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await setCveAllowlistExpires(page, false);
    await expect(page.getByText(/project CVE allowlist has expired/i)).not.toBeVisible();
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openProjectConfig(page);
    await deleteProjectCveAllowlistItem(page, cveAllowlist.last);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
});

test('Test Case - Verfiy Project Level CVE Allowlist By Quick Way of Add System', async ({ page, request }) => {
    test.setTimeout(90 * 1000);
    const project = `project-${Date.now()}`;
    const image = 'goharbor/harbor-portal';
    const digest = '55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d';
    const signinUser = `user${Date.now()}`;
    const signinPassword = 'Test1@34';

    await createUserFromAPI(request, signinUser, signinPassword);
    await resetSystemCveAllowlistFromAPI();
    await loginAs(page, signinUser, signinPassword);
    await createProject(page, project);
    await pushImage(ip, signinUser, signinPassword, project, image, digest);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await setPreventVulnerabilityPolicy(page, 2);
    await scanRepository(page, projectId, project, image);
    const cveAllowlist = await getCveAllowlistDataFromAPI(request, user, pwd, project, image, digest);
    await logout(page, signinUser);
    await login(page);
    await openConfigurationSecurity(page);
    await addSystemCveAllowlist(page, cveAllowlist.all);
    await logout(page);
    await loginAs(page, signinUser, signinPassword);
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openProject(page, projectId);
    await useProjectLevelCveAllowlist(page);
    await scanRepository(page, projectId, project, image);
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await addSystemCvesToProjectAllowlist(page);
    await scanRepository(page, projectId, project, image);
    pullImage(ip, signinUser, signinPassword, project, image, digest);
    await openProjectConfig(page);
    await setCveAllowlistExpires(page, true);
    await expect(page.getByText(/project CVE allowlist has expired/i)).toBeVisible();
    cannotPullImage(ip, signinUser, signinPassword, project, image, digest, 'policy');
    await setCveAllowlistExpires(page, false);
    await expect(page.getByText(/project CVE allowlist has expired/i)).not.toBeVisible();
});

test('Test Case - Stop Scan And Stop Scan All', async ({ page, request }) => {
    test.setTimeout(90 * 1000);
    const project = `project-${Date.now()}`;
    const image = 'goharbor/harbor-e2e-engine';
    const tag = 'test-ui';

    await login(page);
    await createProject(page, project);
    pushImageWithTag(ip, user, pwd, project, image, tag, tag);

    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await page.getByRole('row').filter({ hasText: tag }).locator('label.clr-control-label').click();
    await page.getByRole('button', { name: 'Scan vulnerability' }).click();
    await openActionMenu(page);
    await stopScanIfRunning(page);

    await openVulnerabilityPage(page);
    await page.getByRole('button', { name: 'SCAN NOW' }).click();
    await stopScanAllIfRunning(page);
});

test('Test Case - Verify SBOM Manual Generation', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'alpine';
    const tag = '3.10';

    await login(page);
    await createProject(page, project);
    await pushImage(ip, user, pwd, project, `${image}:${tag}`);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await generateRepoSbom(page, tag);
    await checkoutAndReviewSbomDetails(page, tag);
});

test('Test Case - Generate Image SBOM On Push', async ({ page, request }) => {
    const project = `project-${Date.now()}`;
    const image = 'memcached';
    const tag = 'latest';

    await login(page);
    await createProject(page, project);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProject(page, projectId);
    await openProjectConfig(page);
    await enableSbomOnPush(page);
    await pushImage(ip, user, pwd, project, image);
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await checkoutAndReviewSbomDetails(page, tag);
});

test('Test Case - External Scanner CRUD', async ({ page }) => {
    test.setTimeout(90 * 1000);
    const scanner = `scanner-${Date.now()}`;
    const endpoint = scannerEndpointFor(scanner);

    await login(page);
    await openScannersPage(page);
    await addScanner(page, scanner, endpoint, 'Basic', 'For testing', {
        skipCertVerification: true,
        internalRegistryAddress: true,
        username: 'scanner_name',
        password: 'scanner_password',
    });
    await updateScanner(page, scanner, `${scanner}-edit1`, `${endpoint}1`, 'Bearer', 'For testing-edit1', {
        skipCertVerification: true,
        internalRegistryAddress: true,
        token: 'scanner_token',
    });
    await updateScanner(page, `${scanner}-edit1`, `${scanner}-edit2`, `${endpoint}2`, 'APIKey', 'For testing-edit2', {
        skipCertVerification: true,
        internalRegistryAddress: true,
        apiKey: 'scanner_api_key',
    });
    await updateScanner(page, `${scanner}-edit2`, scanner, endpoint, 'None', 'For testing');
    await filterScannerByName(page, scanner);
    await filterScannerByEndpoint(page, endpoint);
    await expect(page.getByRole('row').filter({ hasText: scanner }).filter({ hasText: endpoint })).toBeVisible();
    await deleteScanner(page, scanner);
});

test('Test Case - Set External Scanner As Default And Scan', async ({ page, request }) => {
    test.setTimeout(150 * 1000);
    const project = `project-${Date.now()}`;
    const scanner = `scanner-${Date.now()}`;
    const endpoint = scannerEndpointFor(scanner);
    const image = 'hello-world';
    const tag = 'latest';

    await setDefaultScannerFromAPI('Trivy');
    await login(page);
    await createProject(page, project);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openProjectScanner(page, projectId);
    await expect(page.locator('#scanner-name')).toHaveText('Trivy');
    await openScannersPage(page);
    await expect(page.getByRole('row').filter({ hasText: 'Trivy' }).filter({ hasText: 'Default' })).toBeVisible();
    await addScanner(page, scanner, endpoint, 'None', 'For testing');
    await setScannerAsDefault(page, scanner);
    await openProjectScanner(page, projectId);
    await expect(page.locator('#scanner-name')).toHaveText(scanner);
    await pushImage(ip, user, pwd, project, image);
    await scanRepository(page, projectId, project, image);
    await expectArtifactScanResult(page, tag, '10');
    await openScannersPage(page);
    await setScannerAsDefault(page, 'Trivy');
    await openProjectScanner(page, projectId);
    await expect(page.locator('#scanner-name')).toHaveText('Trivy');
    await scanRepository(page, projectId, project, image);
    await expectArtifactScanResult(page, tag, /No vulnerability/i);
    await openProjectScanner(page, projectId);
    await selectProjectScanner(page, scanner);
    await scanRepository(page, projectId, project, image);
    await expectArtifactScanResult(page, tag, '10');
    await openProjectScanner(page, projectId);
    await selectProjectScanner(page, 'Trivy');
    await scanRepository(page, projectId, project, image);
    await expectArtifactScanResult(page, tag, /No vulnerability/i);
    await openScannersPage(page);
    await deleteScanner(page, scanner);
    await openProjectScanner(page, projectId);
    await page.locator('#edit-scanner').click();
    await expect(page.getByRole('row').filter({ hasText: scanner })).not.toBeVisible();
    await expect(page.getByRole('row').filter({ hasText: 'Trivy' })).toBeVisible();
});

test('Test Case - Enable And Deactivate Scanner', async ({ page, request }) => {
    test.setTimeout(120 * 1000);
    const project = `project-${Date.now()}`;
    const scanner = `scanner-${Date.now()}`;
    const endpoint = scannerEndpointFor(scanner);
    const image = 'hello-world';
    const tag = 'latest';

    await setDefaultScannerFromAPI('Trivy');
    await login(page);
    await createProject(page, project);
    await pushImage(ip, user, pwd, project, image);
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
    await openScannersPage(page);
    await addScanner(page, scanner, endpoint, 'None', 'For testing');
    await openProjectScanner(page, projectId);
    await selectProjectScanner(page, scanner);
    await scanRepository(page, projectId, project, image);
    await openScannersPage(page);
    await enableOrDeactivateScanner(page, scanner, 'DEACTIVATE');
    await openProjectScanner(page, projectId);
    await expect(page.getByText('Deactivated')).toBeVisible();
    await openProject(page, projectId);
    await page.getByRole('link', { name: `${project}/${image}` }).click();
    await page.getByRole('row').filter({ hasText: tag }).locator('label.clr-control-label').click();
    await expect(page.getByRole('button', { name: 'Scan vulnerability' })).toBeDisabled();
    await openScannersPage(page);
    await enableOrDeactivateScanner(page, scanner, 'ENABLE');
    await openProjectScanner(page, projectId);
    await expect(page.getByText('Deactivated')).not.toBeVisible();
    await scanRepository(page, projectId, project, image);
    await openScannersPage(page);
    await deleteScanner(page, scanner);
    await openProjectScanner(page, projectId);
    await expect(page.locator('#scanner-name')).toHaveText('Trivy');
});

type VulnerabilitySummary = {
    critical_cnt: number;
    high_cnt: number;
    medium_cnt: number;
    low_cnt: number;
    unknown_cnt: number;
    dangerous_artifacts: Array<{
        repository_name: string;
        digest: string;
    }>;
    dangerous_cves: Array<{
        cve_id: string;
        repository_name: string;
        package: string;
        severity: string;
        version: string;
        cvss_score_v3: number;
    }>;
};

type Project = {
    project_id: number;
};

type SystemInfo = {
    harbor_version: string;
};

type VulnerabilityItem = {
    repository_name: string;
    digest: string;
    tags?: string[];
    cve_id: string;
    package: string;
    severity: string;
    cvss_v3_score: number;
    desc?: string;
};

type MultipleFilterSearchData = {
    project: string;
    repositoryName: string;
    digest: string;
    cveId: string;
    packageName: string;
    tag: string;
    cvssScore: string;
    severity: string;
};

type ScannerOptions = {
    skipCertVerification?: boolean;
    internalRegistryAddress?: boolean;
    username?: string;
    password?: string;
    token?: string;
    apiKey?: string;
};

type ScannerRegistration = {
    uuid: string;
    name: string;
    url: string;
    is_default: boolean;
};

type CveAllowlistData = {
    partial: string;
    last: string;
    all: string;
};

type FakeScanArtifact = {
    repository: string;
    digest: string;
    mime_type: string;
};

async function startFakeScannerAdapter(port: number): Promise<Server> {
    const server = createServer(handleFakeScannerRequest);
    await new Promise<void>((resolve, reject) => {
        server.once('error', reject);
        server.listen(port, '0.0.0.0', () => {
            server.off('error', reject);
            resolve();
        });
    });

    log(`fake scanner adapter listening on ${scannerEndpoint}`);
    return server;
}

async function handleFakeScannerRequest(
    request: IncomingMessage,
    response: ServerResponse
): Promise<void> {
    const path = request.url?.split('?')[0] || '';
    const route = fakeScannerRoute(path);

    if (request.method === 'GET' && route === '/api/v1/metadata') {
        writeJson(response, 200, fakeScannerMetadata());
        return;
    }

    if (request.method === 'POST' && route === '/api/v1/scan') {
        const body = await readJsonBody(request);
        const id = `scan-${Date.now()}-${Math.random().toString(16).slice(2)}`;
        fakeScannerReports.set(id, body.artifact || {});
        writeJson(response, 202, { id });
        return;
    }

    const reportMatch = route.match(/^\/api\/v1\/scan\/([^/]+)\/report$/);
    if (request.method === 'GET' && reportMatch) {
        const artifact = fakeScannerReports.get(reportMatch[1]);
        if (!artifact) {
            writeJson(response, 404, { error: { message: 'scan report not found' } });
            return;
        }

        writeJson(response, 200, fakeVulnerabilityReport(artifact));
        return;
    }

    writeJson(response, 404, { error: { message: 'not found' } });
}

function scannerEndpointFor(name: string): string {
    if (process.env.SCANNER_ENDPOINT) {
        return scannerEndpoint;
    }

    return `${scannerEndpoint}/${name}`;
}

function fakeScannerRoute(path: string): string {
    const routeStart = path.indexOf('/api/v1/');
    return routeStart === -1 ? path : path.slice(routeStart);
}

function fakeScannerMetadata() {
    return {
        scanner: {
            name: 'Fake Scanner',
            vendor: 'Harbor',
            version: '1.0.0',
        },
        capabilities: [
            {
                type: 'vulnerability',
                consumes_mime_types: [
                    'application/vnd.oci.image.manifest.v1+json',
                    'application/vnd.docker.distribution.manifest.v2+json',
                ],
                produces_mime_types: [
                    'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0',
                    'application/vnd.security.vulnerability.report; version=1.1',
                ],
            },
        ],
        properties: {
            'harbor.scanner-adapter/registry-authorization-type': 'Bearer',
        },
    };
}

function fakeVulnerabilityReport(artifact: FakeScanArtifact) {
    return {
        generated_at: new Date().toISOString(),
        artifact: {
            repository: artifact.repository,
            digest: artifact.digest,
            mime_type: artifact.mime_type,
        },
        scanner: {
            name: 'Fake Scanner',
            vendor: 'Harbor',
            version: '1.0.0',
        },
        severity: 'Critical',
        vulnerabilities: Array.from({ length: 10 }, (_, index) => ({
            id: `CVE-2026-${String(index + 1).padStart(4, '0')}`,
            package: `fake-package-${index + 1}`,
            version: '1.0.0',
            fix_version: '1.0.1',
            severity: 'Critical',
            description: 'Synthetic vulnerability from Playwright fake scanner adapter.',
            links: [`https://example.test/CVE-2026-${String(index + 1).padStart(4, '0')}`],
            preferred_cvss: {
                score_v3: 9.8,
                vector_v3: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
            },
        })),
    };
}

async function readJsonBody(request: IncomingMessage): Promise<any> {
    const chunks: Buffer[] = [];
    for await (const chunk of request) {
        chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
    }

    if (!chunks.length) {
        return {};
    }

    return JSON.parse(Buffer.concat(chunks).toString('utf-8'));
}

function writeJson(response: ServerResponse, statusCode: number, body: unknown): void {
    response.writeHead(statusCode, {
        'Content-Type': 'application/json',
    });
    response.end(JSON.stringify(body));
}

function vulnerabilitySummary(page: Page): Locator {
    return page.locator('app-vulnerability-summary');
}

function vulnerabilityFilter(page: Page): Locator {
    return page.locator('app-vulnerability-filter');
}

function shortDigest(digest: string): string {
    return digest.slice(0, 12);
}

function repositoryRowName(project: string, repository: string): RegExp {
    return new RegExp(`Select\\s+Select\\s+${project}/${repository}`, 'i');
}

function getCompleteVulnerability(
    vulnerabilities: VulnerabilityItem[]
): VulnerabilityItem {
    const vulnerability = vulnerabilities.find(
        item =>
            item.repository_name &&
            item.digest &&
            item.cve_id &&
            item.package &&
            item.severity &&
            typeof item.cvss_v3_score === 'number'
    );

    if (!vulnerability) {
        throw new Error('No complete vulnerability found for filter checks');
    }

    return vulnerability;
}

async function login(page: Page): Promise<void> {
    await loginAs(page, user, pwd);
}

async function loginAs(page: Page, username: string, password: string): Promise<void> {
    await step(`Sign in as ${username}`, async () => {
        await page.goto('/');
        await page.getByRole('textbox', { name: 'Username' }).fill(username);
        await page.getByRole('textbox', { name: 'Password', exact: true }).fill(password);
        await page.getByRole('button', { name: 'LOG IN' }).click();
    });
}

async function logout(page: Page, username = user): Promise<void> {
    await step(`Log out of Harbor as ${username}`, async () => {
        await page.goto('/');
        const userMenu = page.getByRole('button', { name: username, exact: true });
        await expect(userMenu).toBeVisible();
        await userMenu.click();
        await page.getByRole('menuitem', { name: 'Log Out' }).click();
    });
}

async function createProject(page: Page, project: string, isPublic = false): Promise<void> {
    await step(`Create project ${project}`, async () => {
        const result = await page.evaluate(
            async ({ projectName, publicProject }) => {
                const csrfToken = localStorage.getItem('__csrf') || '';
                const response = await fetch('/api/v2.0/projects', {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-Harbor-CSRF-Token': csrfToken,
                    },
                    body: JSON.stringify({
                        project_name: projectName,
                        metadata: { public: publicProject ? 'true' : 'false' },
                    }),
                });

                return {
                    ok: response.ok,
                    status: response.status,
                    statusText: response.statusText,
                    body: await response.text(),
                };
            },
            { projectName: project, publicProject: isPublic }
        );

        if (!result.ok && result.status !== 409) {
            throw new Error(
                `Failed to create project ${project}: ${result.status} ${result.statusText} ${result.body}`
            );
        }
        await page.goto('/harbor/projects');
        await expect(page.getByRole('heading', { name: 'Projects' }).first()).toBeVisible();
    });
}

async function createUser(page: Page, username: string, password: string): Promise<void> {
    await page.goto('/');
    await page.getByRole('link', { name: 'Sign up for an account' }).click();
    await page.locator('#username').fill(username);
    await page.locator('#email').fill(`${username}@example.com`);
    await page.getByRole('textbox', { name: 'First and last name*' }).fill(username);
    await page.getByRole('textbox', { name: 'Password*', exact: true }).fill(password);
    await page.getByRole('textbox', { name: 'Confirm Password*' }).fill(password);
    await page.getByRole('button', { name: 'SIGN UP' }).click();
    await expect(page.getByRole('textbox', { name: 'Username' })).toBeVisible();
}

async function openVulnerabilityPage(page: Page): Promise<void> {
    await step('Open vulnerability page', async () => {
        await page.goto('/harbor/interrogation-services/vulnerability');
        await expect(page.getByRole('button', { name: 'SCAN NOW' })).toBeVisible();
    });
}

async function triggerScanNowAndWait(page: Page): Promise<void> {
    const scanNow = page.getByRole('button', { name: 'SCAN NOW' });
    await scanNow.click();
    await expect(scanNow).toBeDisabled({ timeout: 5000 }).catch(() => undefined);
    await expect(scanNow).toBeEnabled({ timeout: 120000 });
}

async function openScannersPage(page: Page): Promise<void> {
    await step('Open scanners page', async () => {
        await page.goto('/harbor/interrogation-services/scanners');
        await expect(page.getByRole('button', { name: /NEW SCANNER/i })).toBeVisible();
    });
}

async function openProjectConfig(page: Page): Promise<void> {
    await step('Open project configuration', async () => {
        const projectId = page.url().match(/\/projects\/(\d+)/)?.[1];
        if (projectId) {
            await page.goto(`/harbor/projects/${projectId}/configs`);
        } else {
            await page.getByRole('tab', { name: 'Configuration' }).click();
        }
        await expect(page.locator('hbr-project-policy-config')).toBeVisible();
    });
}

async function openProjectScanner(page: Page, projectId: number): Promise<void> {
    await step(`Open project scanner projectId=${projectId}`, async () => {
        await page.goto(`/harbor/projects/${projectId}/scanner`);
        await expect(page.locator('scanner')).toBeVisible();
    });
}

async function openProject(page: Page, projectId: number): Promise<void> {
    await step(`Open project repositories projectId=${projectId}`, async () => {
        await page.goto(`/harbor/projects/${projectId}/repositories`);
    });
}

async function enableScanOnPush(page: Page): Promise<void> {
    const checkbox = page.locator('#scan-image-on-push-wrapper input');
    if (!(await checkbox.isChecked())) {
        await page.locator('#scan-image-on-push-wrapper label.clr-control-label').click();
    }
    await expect(checkbox).toBeChecked();
    await saveProjectConfig(page);
}

async function enableSbomOnPush(page: Page): Promise<void> {
    const checkbox = page.locator('#generate-sbom-on-push-wrapper input');
    if (!(await checkbox.isChecked())) {
        await page.locator('#generate-sbom-on-push-wrapper label.clr-control-label').click();
    }
    await expect(checkbox).toBeChecked();
    await saveProjectConfig(page);
}

async function setPreventVulnerabilityPolicy(
    page: Page,
    levelIndex: number
): Promise<void> {
    await openProjectConfig(page);
    const checkbox = page.getByRole('checkbox', {
        name: 'Prevent vulnerable images from running.',
    });
    if (!(await checkbox.isChecked())) {
        await checkbox.check({ force: true });
    }
    await expect(checkbox).toBeChecked();
    await page
        .getByRole('combobox', {
            name: /Prevent images with vulnerability severity/i,
        })
        .selectOption({ index: levelIndex });
    await saveProjectConfig(page);
}

async function saveProjectConfig(page: Page): Promise<void> {
    const saveButton = page.locator('hbr-project-policy-config').getByRole('button', { name: 'SAVE' });
    const projectId = page.url().match(/\/projects\/(\d+)/)?.[1];
    await expect(saveButton).toBeEnabled();
    if (!projectId) {
        await saveButton.click();
        return;
    }

    await Promise.all([
        page.waitForResponse(response =>
            response.url().includes(`/api/v2.0/projects/${projectId}`) &&
            response.request().method() === 'PUT' &&
            response.ok()
        ),
        saveButton.click(),
    ]);
}

async function openConfigurationSecurity(page: Page): Promise<void> {
    await page.goto('/harbor/configs/security');
    await expect(page.getByText('Deployment security')).toBeVisible();
}

async function addSystemCveAllowlist(page: Page, cves: string): Promise<void> {
    await openConfigurationSecurity(page);
    await page.locator('#show-add-modal-button').click();
    await page.locator('#allowlist-textarea').fill(cves);
    await page.locator('#add-to-system').click();
    await saveSecurityConfig(page);
}

async function addProjectCveAllowlist(page: Page, cves: string): Promise<void> {
    await openProjectConfig(page);
    await useProjectLevelCveAllowlist(page);
    await page.locator('#show-add-modal').click();
    await page.locator('#allowlist-textarea').fill(cves);
    await page.locator('#add-to-allowlist').click();
    await saveProjectConfig(page);
}

async function useProjectLevelCveAllowlist(page: Page): Promise<void> {
    await openProjectConfig(page);
    const projectAllowlist = page.getByRole('radio', { name: 'Project allowlist' });
    if (!(await projectAllowlist.isChecked())) {
        await projectAllowlist.check({ force: true });
        await saveProjectConfig(page);
    }
    await expect(projectAllowlist).toBeChecked();
}

async function addSystemCvesToProjectAllowlist(page: Page): Promise<void> {
    await openProjectConfig(page);
    await page.locator('#add-system').click();
    await saveProjectConfig(page);
}

async function deleteSystemCveAllowlistItem(
    page: Page,
    cve: string
): Promise<void> {
    const allowlist = page.locator('.allowlist-window').first();
    await allowlist.locator('li', { hasText: cve }).locator('a.float-lg-right').click({ force: true });
    await saveSecurityConfig(page);
}

async function deleteProjectCveAllowlistItem(page: Page, cve: string): Promise<void> {
    await openProjectConfig(page);
    await page
        .locator('hbr-project-policy-config .allowlist-window')
        .last()
        .locator('li', { hasText: cve })
        .locator('a.float-lg-right')
        .click({ force: true });
    await saveProjectConfig(page);
}

async function saveSecurityConfig(page: Page): Promise<void> {
    const saveButton = page.locator('#security_save');
    if (await saveButton.isDisabled()) {
        return;
    }
    await expect(saveButton).toBeEnabled();
    await Promise.all([
        page.waitForResponse(response =>
            response.url().includes('/api/v2.0/system/CVEAllowlist') &&
            response.request().method() === 'PUT' &&
            response.ok()
        ),
        saveButton.click(),
    ]);
}

async function setCveAllowlistExpires(
    page: Page,
    expired: boolean
): Promise<void> {
    const date = new Date();
    date.setDate(date.getDate() + (expired ? -1 : 1));
    const value = formatDateInputValue(date);
    const dateInput = page.getByRole('textbox', { name: /Never expires|Expires at/i }).last();
    const neverExpires = page.locator('#neverExpires').last();

    if (await neverExpires.isChecked()) {
        await neverExpires.uncheck({ force: true });
    }
    await dateInput.evaluate((element, nextValue) => {
        const input = element as HTMLInputElement;
        input.removeAttribute('readonly');
        input.value = nextValue;
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
    }, value);

    if (await page.locator('#security_save').isVisible()) {
        await saveSecurityConfig(page);
    } else {
        await saveProjectConfig(page);
    }
}

function formatDateInputValue(date: Date): string {
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${month}/${day}/${date.getFullYear()}`;
}

async function generateRepoSbom(page: Page, tag: string): Promise<void> {
    const row = page.getByRole('row').filter({ hasText: tag });
    await row.locator('label.clr-control-label').click();
    await page.locator('#generate-sbom-btn').click();
    await expect(page.getByRole('link', { name: 'SBOM details' })).toBeVisible({
        timeout: scanResultTimeout,
    });
}

async function checkoutAndReviewSbomDetails(page: Page, tag: string): Promise<void> {
    await page.getByRole('row').filter({ hasText: tag }).getByRole('link', { name: 'SBOM details' }).click();
    const downloadPromise = page.waitForEvent('download');
    await page.locator('#sbom-btn').click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/\.json$/);
    await expect(page.getByRole('row').filter({ hasText: /Package|Version|License/i }).first()).toBeVisible();
}

async function scanRepository(
    page: Page,
    projectId: number,
    project: string,
    repository: string
): Promise<void> {
    await openProject(page, projectId);
    await retryStep(`Open repository ${project}/${repository}`, async () => {
        await page.getByRole('link', { name: `${project}/${repository}` }).click();
        await expect(page.getByRole('heading', { name: repository })).toBeVisible();
    });
    await scanCurrentArtifact(page);
}

async function scanCurrentArtifact(page: Page): Promise<void> {
    await step('Select artifact and trigger vulnerability scan', async () => {
        await page
            .getByRole('gridcell', { name: 'Select Select' })
            .locator('label')
            .click();
        await page.getByRole('checkbox', { name: 'Select', exact: true }).check();

        const scanButton = page.getByRole('button', { name: 'Scan vulnerability' });
        await expect(scanButton).toBeEnabled();
        await scanButton.click();
    });
    await retryStep('Wait for vulnerability scan result', async () => {
        await page
            .getByRole('gridcell', { name: /Total|No vulnerability/ })
            .waitFor({ timeout: scanResultTimeout });
    }, 6);
}

async function openActionMenu(page: Page): Promise<void> {
    await step('Open action menu', async () => {
        const actionMenu = page
            .locator('clr-dropdown, clr-dg-action-overflow')
            .filter({ hasText: /action/i })
            .first();
        await actionMenu.click();
    });
}

async function stopScanIfRunning(page: Page): Promise<void> {
    const stopScan = page.getByRole('menuitem', { name: /Stop Scan/i });
    if (await stopScan.isEnabled()) {
        await stopScan.click();
        await expect(page.getByText(/Stopped/i).first()).toBeVisible({ timeout: scanResultTimeout });
        return;
    }

    await page.keyboard.press('Escape').catch(() => undefined);
    await expect(page.getByRole('gridcell', { name: /Total|No vulnerability/ })).toBeVisible({ timeout: scanResultTimeout });
}

async function stopScanAllIfRunning(page: Page): Promise<void> {
    const stopScan = page.getByRole('button', { name: /STOP SCAN/i });
    if (await stopScan.isEnabled()) {
        await stopScan.click();
        await expect(page.getByRole('button', { name: 'SCAN NOW' })).toBeVisible({ timeout: scanResultTimeout });
        return;
    }

    await expect(page.getByRole('button', { name: 'SCAN NOW' })).toBeVisible();
}

async function deleteRepository(
    page: Page,
    projectId: number,
    project: string,
    repository: string
): Promise<void> {
    await openProject(page, projectId);
    await page.locator('.refresh-btn > clr-icon').click({ force: true });

    const row = page.getByRole('row', {
        name: repositoryRowName(project, repository),
    });
    await row.waitFor({ state: 'visible', timeout: 30000 });
    await row.locator('label').click();

    const deleteButton = page.getByRole('button', { name: 'Delete' });
    await expect(deleteButton).toBeEnabled();
    await deleteButton.click();
    await page.getByRole('button', { name: 'DELETE', exact: true }).click();
}

async function openSecurityHub(page: Page): Promise<void> {
    await page.goto('/');
    await page.getByRole('link', { name: 'Interrogation Services' }).click();
    await page.getByRole('link', { name: 'Security Hub' }).click();
    await expect(vulnerabilitySummary(page)).toBeVisible();
}

async function assertSummaryCounts(
    page: Page,
    summary: VulnerabilitySummary
): Promise<void> {
    for (const count of [
        summary.critical_cnt,
        summary.high_cnt,
        summary.medium_cnt,
        summary.low_cnt,
    ]) {
        await expect(vulnerabilitySummary(page)).toContainText(String(count));
    }
}

async function assertDangerousArtifacts(
    page: Page,
    artifacts: VulnerabilitySummary['dangerous_artifacts']
): Promise<void> {
    for (const artifact of artifacts) {
        await expect(vulnerabilitySummary(page)).toContainText(
            artifact.repository_name
        );
        await expect(vulnerabilitySummary(page)).toContainText(
            shortDigest(artifact.digest)
        );
    }
}

async function assertDangerousCVEs(
    page: Page,
    cves: VulnerabilitySummary['dangerous_cves']
): Promise<void> {
    for (const cve of cves) {
        await expect(vulnerabilitySummary(page)).toContainText(cve.cve_id);
        await expect(vulnerabilitySummary(page)).toContainText(cve.severity);
        await expect(vulnerabilitySummary(page)).toContainText(
            String(cve.cvss_score_v3)
        );
        await expect(vulnerabilitySummary(page)).toContainText(
            `${cve.package}@${cve.version}`
        );
    }
}

async function assertQuickSearchByArtifact(
    page: Page,
    artifact: VulnerabilitySummary['dangerous_artifacts'][number]
): Promise<void> {
    await vulnerabilitySummary(page)
        .getByRole('link', { name: artifact.repository_name })
        .first()
        .click();

    await expect(
        vulnerabilityFilter(page).getByRole('textbox').first()
    ).toHaveValue(artifact.repository_name);
    await expect(
        vulnerabilityFilter(page).getByRole('textbox').nth(1)
    ).toHaveValue(artifact.digest);
    await expectGridCellVisible(page, artifact.repository_name);
    await expectGridCellVisible(page, shortDigest(artifact.digest));
}

async function assertQuickSearchByCve(
    page: Page,
    cveId: string
): Promise<void> {
    await vulnerabilitySummary(page)
        .getByRole('link', { name: cveId })
        .first()
        .click();

    await expect(vulnerabilityFilter(page).getByRole('textbox')).toHaveValue(
        cveId
    );
    await expect(vulnerabilityFilter(page).getByRole('combobox')).toHaveValue(
        'cve_id'
    );
    await expectGridCellVisible(page, cveId);
}

async function searchByTextFilter(
    page: Page,
    option: string,
    value: string
): Promise<void> {
    const filter = vulnerabilityFilter(page);
    await filter.getByRole('combobox').first().selectOption(option);
    await filter.getByRole('textbox').first().fill(value);
    await search(page);
}

async function searchByCvssScore(page: Page, score: string): Promise<void> {
    const filter = vulnerabilityFilter(page);
    await filter.getByRole('combobox').first().selectOption('cvss_score_v3');
    await filter.getByRole('textbox').first().fill(score);
    await filter.getByRole('textbox').nth(1).fill(score);
    await search(page);
}

async function searchBySeverity(page: Page, severity: string): Promise<void> {
    const filter = vulnerabilityFilter(page);
    await filter.getByRole('combobox').first().selectOption('severity');
    await filter.getByRole('combobox').nth(1).selectOption(severity);
    await search(page);
}

async function assertSeverityFilter(
    page: Page,
    severity: string,
    count: number
): Promise<void> {
    await searchBySeverity(page, severity);
    const expectedFooter = count > 1000 ? '1000+ CVEs' : `${count} CVEs`;
    await expect(page.locator('clr-dg-footer')).toContainText(expectedFooter);

    if (count === 0) {
        await expectNoVulnerabilities(page);
        return;
    }

    await expectGridCellVisible(page, severity === 'Unknown' ? 'n/a' : severity);
}

async function searchByAllFilters(
    page: Page,
    data: MultipleFilterSearchData
): Promise<void> {
    await setTextFilterAtRow(page, 0, 'project_id', data.project);
    await expect(
        vulnerabilityFilter(page).locator('clr-icon').nth(2)
    ).toBeVisible();
    await search(page);

    await addFilterCondition(page);
    await setTextFilterAtRow(page, 1, 'repository_name', data.repositoryName);
    await search(page);

    await addFilterCondition(page);
    await setTextFilterAtRow(page, 2, 'digest', data.digest);
    await search(page);

    await addFilterCondition(page);
    await setTextFilterAtRow(page, 3, 'cve_id', data.cveId);
    await search(page);

    await addFilterCondition(page);
    await setTextFilterAtRow(page, 4, 'package', data.packageName);
    await search(page);

    await addFilterCondition(page);
    await setTextFilterAtRow(page, 5, 'tag', data.tag);
    await search(page);

    await addFilterCondition(page);
    await vulnerabilityFilter(page)
        .getByRole('combobox')
        .nth(6)
        .selectOption('cvss_score_v3');
    await fillCvssRange(page, data.cvssScore);

    await addFilterCondition(page);
    await page
        .locator('.clr-select-wrapper.ml-1 > .clr-select')
        .selectOption(data.severity);
    await search(page);
}

async function setTextFilterAtRow(
    page: Page,
    rowIndex: number,
    option: string,
    value: string
): Promise<void> {
    const filter = vulnerabilityFilter(page);
    await filter.getByRole('combobox').nth(rowIndex).selectOption(option);
    await filter.getByRole('textbox').nth(rowIndex).fill(value);
}

async function fillCvssRange(page: Page, score: string): Promise<void> {
    const range = page.locator('div').filter({ hasText: /^FromTo$/ });
    await range.getByRole('textbox').first().fill(score);
    await range.getByRole('textbox').nth(1).fill(score);
}

async function addFilterCondition(page: Page): Promise<void> {
    await vulnerabilityFilter(page).locator('clr-icon').nth(1).click();
}

async function search(page: Page): Promise<void> {
    await page.getByRole('button', { name: 'SEARCH' }).click();
}

async function expectGridCellVisible(page: Page, name: string): Promise<void> {
    await expect(page.getByRole('gridcell', { name }).first()).toBeVisible();
}

async function expectNoVulnerabilities(page: Page): Promise<void> {
    await expect(page.locator('clr-dg-row')).toHaveCount(0);
}

function sleepSync(ms: number): void {
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, ms);
}

function authorizationHeader(user: string, password: string): string {
    return `Basic ${Buffer.from(`${user}:${password}`).toString('base64')}`;
}

async function getVulnerabilitySummaryFromAPI(
    request: APIRequestContext,
    user: string,
    password: string
): Promise<VulnerabilitySummary> {
    const response = await request.get('/api/v2.0/security/summary', {
        params: {
            with_dangerous_cve: 'true',
            with_dangerous_artifact: 'true',
        },
        headers: {
            Authorization: authorizationHeader(user, password),
            'Content-Type': 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to fetch vulnerability summary: ${response.status()} ${response.statusText()}`
        );
    }

    return response.json();
}

async function getProjectIdFromAPI(
    request: APIRequestContext,
    user: string,
    password: string,
    projectName: string
): Promise<number> {
    const response = await request.get('/api/v2.0/projects', {
        params: {
            name: projectName,
            page: '1',
            page_size: '1',
        },
        headers: {
            Authorization: authorizationHeader(user, password),
            'Content-Type': 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to fetch project: ${response.status()} ${response.statusText()}`
        );
    }

    const projects: Project[] = await response.json();
    if (!projects.length) {
        throw new Error(`Project not found: ${projectName}`);
    }

    return projects[0].project_id;
}

async function getVulnerabilitiesFromAPI(
    request: APIRequestContext,
    user: string,
    password: string,
    queryConditions: string[]
): Promise<VulnerabilityItem[]> {
    const response = await request.get('/api/v2.0/security/vul', {
        params: {
            with_tag: 'true',
            page: '1',
            page_size: '15',
            q: queryConditions.join(','),
        },
        headers: {
            Authorization: authorizationHeader(user, password),
            'Content-Type': 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to fetch vulnerabilities: ${response.status()} ${response.statusText()}`
        );
    }

    return response.json();
}

async function getCveAllowlistDataFromAPI(
    request: APIRequestContext,
    user: string,
    password: string,
    project: string,
    repository: string,
    reference: string
): Promise<CveAllowlistData> {
    const blockingSeverities = new Set(['Critical', 'High', 'Medium']);
    let cves = new Set<string>();
    let blockingCves = new Set<string>();

    for (let attempt = 0; attempt < 10; attempt += 1) {
        cves = new Set<string>();
        blockingCves = new Set<string>();
        const response = await request.get(
            `/api/v2.0/projects/${encodeURIComponent(project)}/repositories/${encodeRepositoryName(repository)}/artifacts/${reference}/additions/vulnerabilities`,
            {
                headers: {
                    Authorization: authorizationHeader(user, password),
                    'Content-Type': 'application/json',
                },
            }
        );

        if (!response.ok()) {
            throw new Error(
                `Failed to fetch CVE allowlist data: ${response.status()} ${response.statusText()}`
            );
        }

        const reports: Record<string, { vulnerabilities?: Array<{ id: string; severity: string }> }> = await response.json();
        for (const report of Object.values(reports)) {
            for (const vulnerability of report.vulnerabilities || []) {
                cves.add(vulnerability.id);
                if (blockingSeverities.has(vulnerability.severity)) {
                    blockingCves.add(vulnerability.id);
                }
            }
        }

        if (cves.size >= 2 && blockingCves.size > 0) {
            break;
        }

        await new Promise(resolve => setTimeout(resolve, 2000));
    }

    const allCves = Array.from(cves);
    const blockingCveList = Array.from(blockingCves);
    if (allCves.length < 2 || blockingCveList.length < 2) {
        throw new Error(
            `Expected at least two CVEs and one blocking CVE, got ${allCves.length} CVEs and ${blockingCves.size} blocking CVEs`
        );
    }
    const lastBlockingCve = blockingCveList[blockingCveList.length - 1];

    return {
        partial: blockingCveList.slice(0, -1).join('\n'),
        last: lastBlockingCve,
        all: blockingCveList.join('\n'),
    };
}

async function resetSystemCveAllowlistFromAPI(): Promise<void> {
    const baseURL = process.env.BASE_URL;
    if (!baseURL) {
        throw new Error('BASE_URL is required to reset the system CVE allowlist');
    }
    const api = await playwrightRequest.newContext({
        baseURL,
        extraHTTPHeaders: {
            Authorization: authorizationHeader(user, pwd),
            'Content-Type': 'application/json',
        },
    });

    const response = await api.put('/api/v2.0/system/CVEAllowlist', {
        data: {
            items: [],
            expires_at: null,
        },
    });

    try {
        if (!response.ok()) {
            throw new Error(
                `Failed to reset system CVE allowlist: ${response.status()} ${response.statusText()}`
            );
        }
    } finally {
        await api.dispose();
    }
}

function encodeRepositoryName(repository: string): string {
    return encodeURIComponent(repository).replace(/%2F/g, '%252F');
}

async function createUserFromAPI(
    request: APIRequestContext,
    username: string,
    password: string
): Promise<void> {
    await step(`Create Harbor user ${username}`, async () => {
        const response = await request.post('/api/v2.0/users', {
            headers: {
                Authorization: authorizationHeader(user, pwd),
                'Content-Type': 'application/json',
            },
            data: {
                username,
                email: `${username}@example.com`,
                realname: username,
                password,
                comment: 'Playwright Trivy test user',
            },
        });

        if (!response.ok() && response.status() !== 409) {
            throw new Error(
                `Failed to create user ${username}: ${response.status()} ${response.statusText()}`
            );
        }
    });
}

async function getSystemInfoFromAPI(
    request: APIRequestContext,
    user: string,
    password: string
): Promise<SystemInfo> {
    const response = await request.get('/api/v2.0/systeminfo', {
        headers: {
            Authorization: authorizationHeader(user, password),
            Accept: 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to fetch system info: ${response.status()} ${response.statusText()}`
        );
    }

    return response.json();
}

async function addScanner(
    page: Page,
    name: string,
    endpoint: string,
    auth: string,
    description?: string,
    options: ScannerOptions = {}
): Promise<void> {
    await deleteStaleScanners(page, name);
    await page.getByRole('button', { name: /NEW SCANNER/i }).click();
    await fillScannerForm(page, name, endpoint, auth, description, options);
    await page.locator('#button-test').click();
    await expect(page.getByText('Test passed')).toBeVisible({ timeout: scanResultTimeout });
    await page.locator('#button-add').click();
    await expectScannerRowVisible(page, name, endpoint);
}

async function updateScanner(
    page: Page,
    originalName: string,
    name: string,
    endpoint: string,
    auth: string,
    description?: string,
    options: ScannerOptions = {}
): Promise<void> {
    await selectScannerRow(page, originalName);
    await page.locator('#action-scanner').click();
    await page.getByRole('menuitem', { name: 'EDIT' }).click();
    await fillScannerForm(page, name, endpoint, auth, description, options);
    await page.locator('#button-save').click();
    await expectScannerRowVisible(page, name, endpoint, auth);
}

async function fillScannerForm(
    page: Page,
    name: string,
    endpoint: string,
    auth: string,
    description?: string,
    options: ScannerOptions = {}
): Promise<void> {
    await page.locator('#scanner-name').fill(name);
    if (description !== undefined) {
        await page.locator('#description').fill(description);
    }
    await page.locator('#scanner-endpoint').fill(endpoint);
    await page.locator('#scanner-authorization').selectOption(auth);

    if (auth === 'Basic') {
        await page.locator('#scanner-username').fill(options.username || '');
        await page.locator('#scanner-password').fill(options.password || '');
    }
    if (auth === 'Bearer') {
        await page.locator('#scanner-token').fill(options.token || '');
    }
    if (auth === 'APIKey') {
        await page.locator('#scanner-apiKey').fill(options.apiKey || '');
    }
    if (options.skipCertVerification) {
        await setCheckbox(page, 'scanner-skipCertVerify', true);
    }
    if (options.internalRegistryAddress) {
        await setCheckbox(page, 'scanner-use-inner', true);
    }
}

async function filterScannerByName(page: Page, name: string): Promise<void> {
    await expectScannerRowVisible(page, name);
}

async function filterScannerByEndpoint(page: Page, endpoint: string): Promise<void> {
    await expectScannerRowVisible(page, endpoint);
}

async function deleteScanner(page: Page, name: string): Promise<void> {
    await selectScannerRow(page, name);
    await page.locator('#action-scanner').click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'DELETE', exact: true }).click();
    await page.reload();
    await expect(scannerRow(page, name)).not.toBeVisible();
}

async function setScannerAsDefault(page: Page, name: string): Promise<void> {
    await selectScannerRow(page, name);
    await page.locator('#set-default').click();
    await expectScannerRowVisible(page, name, 'Default');
}

async function enableOrDeactivateScanner(
    page: Page,
    name: string,
    action: 'ENABLE' | 'DEACTIVATE'
): Promise<void> {
    await selectScannerRow(page, name);
    await page.locator('#action-scanner').click();
    await page.getByRole('menuitem', { name: action }).click();
    const expectedState = action === 'ENABLE' ? 'Enabled' : 'Deactivated';
    await expectScannerRowVisible(page, name, expectedState);
}

async function selectProjectScanner(
    page: Page,
    name: string,
    scannerCount?: number
): Promise<void> {
    await page.locator('#edit-scanner').click();
    if (scannerCount !== undefined) {
        await expect(page.locator('clr-dg-row')).toHaveCount(scannerCount);
    }
    await page.getByRole('row').filter({ hasText: name }).locator('label.clr-control-label').click();
    await page.locator('#save-scanner').click();
    await expect(page.locator('#scanner-name')).toHaveText(name);
}

async function selectScannerRow(page: Page, name: string): Promise<void> {
    await expectScannerRowVisible(page, name);
    const row = scannerRow(page, name);
    await row.locator('label.clr-control-label').click();
}

function scannerRow(page: Page, name: string): Locator {
    return page.getByRole('row').filter({ hasText: name });
}

async function expectScannerRowVisible(
    page: Page,
    name: string,
    ...texts: string[]
): Promise<void> {
    let lastError: unknown;

    for (let attempt = 1; attempt <= 4; attempt += 1) {
        const row = texts.reduce(
            (locator, text) => locator.filter({ hasText: text }),
            scannerRow(page, name)
        );

        try {
            await expect(row).toBeVisible({ timeout: 5000 });
            return;
        } catch (error) {
            lastError = error;
            if (attempt < 4) {
                await page.reload();
            }
        }
    }

    throw lastError;
}

async function deleteStaleScanners(
    page: Page,
    name: string
): Promise<void> {
    const response = await page.request.get('/api/v2.0/scanners', {
        headers: {
            Authorization: authorizationHeader(user, pwd),
            Accept: 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to list scanners: ${response.status()} ${response.statusText()}`
        );
    }

    const scanners = await response.json() as ScannerRegistration[];
    for (const scanner of scanners) {
        if (scanner.name !== name) {
            continue;
        }

        const deleteResponse = await page.request.delete(`/api/v2.0/scanners/${scanner.uuid}`, {
            headers: {
                Authorization: authorizationHeader(user, pwd),
            },
        });

        if (!deleteResponse.ok() && deleteResponse.status() !== 404) {
            throw new Error(
                `Failed to delete stale scanner ${scanner.name}: ${deleteResponse.status()} ${deleteResponse.statusText()}`
            );
        }
    }
}

async function setDefaultScannerFromAPI(name: string): Promise<void> {
    const baseURL = process.env.BASE_URL;
    if (!baseURL) {
        throw new Error('BASE_URL is required to set the default scanner');
    }

    const api = await playwrightRequest.newContext({
        baseURL,
        extraHTTPHeaders: {
            Authorization: authorizationHeader(user, pwd),
            'Content-Type': 'application/json',
        },
    });

    try {
        const scanners = await getScannersFromAPI(api);
        const scanner = scanners.find(candidate => candidate.name === name);
        if (!scanner) {
            throw new Error(`Scanner ${name} was not found`);
        }

        if (scanner.is_default) {
            return;
        }

        const response = await api.patch(`/api/v2.0/scanners/${scanner.uuid}`, {
            data: {
                is_default: true,
            },
        });

        if (!response.ok()) {
            throw new Error(
                `Failed to set scanner ${name} as default: ${response.status()} ${response.statusText()}`
            );
        }
    } finally {
        await api.dispose();
    }
}

async function getScannersFromAPI(
    request: APIRequestContext
): Promise<ScannerRegistration[]> {
    const response = await request.get('/api/v2.0/scanners', {
        headers: {
            Authorization: authorizationHeader(user, pwd),
            Accept: 'application/json',
        },
    });

    if (!response.ok()) {
        throw new Error(
            `Failed to list scanners: ${response.status()} ${response.statusText()}`
        );
    }

    return response.json();
}

async function setCheckbox(
    page: Page,
    id: string,
    checked: boolean
): Promise<void> {
    const checkbox = page.locator(`#${id}`);
    if ((await checkbox.isChecked()) === checked) {
        return;
    }

    await page.locator(`label[for="${id}"]`).click();
    if (checked) {
        await expect(checkbox).toBeChecked();
    } else {
        await expect(checkbox).not.toBeChecked();
    }
}

async function expectArtifactScanResult(
    page: Page,
    tag: string,
    expected: string | RegExp
): Promise<void> {
    const row = page.getByRole('row').filter({ hasText: tag });
    await expect(row.getByRole('gridcell', { name: /Total.*Fixable|No vulnerability/ })).toBeVisible({
        timeout: scanResultTimeout,
    });
    await expect(row).toContainText(expected, {
        timeout: scanResultTimeout,
    });
}

function cannotPullImage(
    ip: string,
    user: string,
    pwd: string,
    project: string,
    image: string,
    tag: string,
    expectedMessage?: string
): void {
    dockerLogin(ip, user, pwd);
    try {
        let failure = '';
        for (let attempt = 0; attempt < 5; attempt += 1) {
            try {
                runDockerCommand(['pull', `${ip}/${project}/${image}:${tag}`]);
            } catch (error) {
                failure = String((error as Error).message);
                break;
            }
            sleepSync(500);
        }
        expect(failure).toBeTruthy();
        if (expectedMessage) {
            expect(failure).toContain(expectedMessage);
        }
    } finally {
        runDockerCommand(['logout', ip]);
    }
}

function runDockerCommand(
    args: string[],
    options: Partial<ExecFileSyncOptionsWithStringEncoding> = {}
): string {
    log(`docker ${args.join(' ')}`);
    try {
        return execFileSync('docker', args, {
            encoding: 'utf-8',
            stdio: ['pipe', 'pipe', 'pipe'],
            ...options,
        });
    } catch (error) {
        const execError = error as Error & {
            stderr?: Buffer | string;
            stdout?: Buffer | string;
            status?: number;
        };
        const stderr = execError.stderr?.toString().trim();
        const stdout = execError.stdout?.toString().trim();
        throw new Error(
            [
                `Docker command failed: docker ${args.join(' ')}`,
                execError.status === undefined ? '' : `exit status: ${execError.status}`,
                stderr ? `stderr: ${stderr}` : '',
                stdout ? `stdout: ${stdout}` : '',
            ]
                .filter(Boolean)
                .join('\n')
        );
    }
}

function dockerLogin(registry: string, user: string, password: string): void {
    runDockerCommand(
        ['login', registry, '--username', user, '--password-stdin'],
        {
            input: `${password}\n`,
        }
    );
}

function pushManifestList(
    registry: string,
    user: string,
    password: string,
    manifest: string,
    images: string[]
): void {
    dockerLogin(registry, user, password);

    try {
        runDockerCommand(['manifest', 'create', manifest, ...images]);
        runDockerCommand(['manifest', 'push', manifest]);
    } finally {
        runDockerCommand(['logout', registry]);
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
    tag1: string = 'latest',
    localRegistry: string = LOCAL_REGISTRY,
    localNamespace: string = LOCAL_REGISTRY_NAMESPACE
): void {
    const sourceImage = `${localRegistry}/${localNamespace}/${image}:${tag1}`;
    const targetImage = `${ip}/${project}/${image}:${tag}`;

    dockerLogin(ip, user, pwd);

    try {
        runDockerCommand(['pull', sourceImage]);
        runDockerCommand(['tag', sourceImage, targetImage]);
        runDockerCommand(['push', targetImage]);
    } finally {
        runDockerCommand(['logout', ip]);
    }
}

function pullImage(
    ip: string,
    user: string,
    pwd: string,
    project: string,
    image: string,
    tag: string
): void {
    dockerLogin(ip, user, pwd);
    try {
        let lastError: unknown;
        for (let attempt = 0; attempt < 3; attempt += 1) {
            try {
                runDockerCommand(['pull', `${ip}/${project}/${image}:${tag}`]);
                return;
            } catch (error) {
                lastError = error;
                sleepSync(500);
            }
        }
        throw lastError;
    } finally {
        runDockerCommand(['logout', ip]);
    }
}

async function pushImage(
    ip: string,
    user: string,
    pwd: string,
    project: string,
    image: string,
    digest?: string,
    needPullFirst = true,
    isRobot = false,
    localRegistry: string = LOCAL_REGISTRY,
    localNamespace: string = LOCAL_REGISTRY_NAMESPACE
): Promise<void> {
    const imageInUse = digest ? `${image}@sha256:${digest}` : image;
    const imageInUseWithTag = digest ? `${image}:${digest}` : image;
    const sourceImage = needPullFirst
        ? `${localRegistry}/${localNamespace}/${imageInUse}`
        : imageInUse;
    const targetImage = `${ip}/${project}/${imageInUseWithTag}`;
    const username = isRobot ? `robot$${project}+${user}` : user;

    dockerLogin(ip, username, pwd);
    try {
        if (needPullFirst) {
            runDockerCommand(['pull', sourceImage]);
        }
        runDockerCommand(['tag', sourceImage, targetImage]);
        runDockerCommand(['push', targetImage]);
    } finally {
        runDockerCommand(['logout', ip]);
    }
}

async function scanResultShouldDisplayInListRow(
    page: Page,
    tagOrDigest: string,
    hasNoVulnerability = false
): Promise<void> {
    const row = page.getByRole('row').filter({ hasText: tagOrDigest });
    const vulnerabilityCell = hasNoVulnerability
        ? row.getByRole('gridcell', { name: 'No vulnerability' })
        : row.getByRole('gridcell', { name: /Total.*Fixable/ });

    await expect(vulnerabilityCell).toBeVisible({ timeout: scanResultTimeout });
}

async function scanResultShouldDisplayAnyInListRow(
    page: Page,
    tagOrDigest: string
): Promise<void> {
    const row = page.getByRole('row').filter({ hasText: tagOrDigest });
    await expect(row.getByRole('gridcell', { name: /Total.*Fixable|No vulnerability/ })).toBeVisible({
        timeout: scanResultTimeout,
    });
}

async function viewRepoScanDetails(page: Page, vulnerabilityLevels: string[]): Promise<void> {
    await step(`Review vulnerability details: ${vulnerabilityLevels.join(', ')}`, async () => {
        const vulnerabilityTable = page.locator('hbr-artifact-vulnerabilities');
        await expect(vulnerabilityTable).toBeVisible({ timeout: 10000 });

        const pageSize = page.getByRole('combobox', { name: 'Page size' }).last();
        if (await pageSize.isVisible()) {
            await pageSize.selectOption('100');
        }

        for (const vulnerabilityLevel of vulnerabilityLevels) {
            const row = vulnerabilityTable
                .getByRole('row')
                .filter({ hasText: vulnerabilityLevel })
                .first();

            await expect(row).toBeVisible({ timeout: scanResultTimeout });
        }
        await page.getByRole('tab', { name: /Build History/ }).click();
        await expect(page.getByText(/Build History/i)).toBeVisible();
    });
}
