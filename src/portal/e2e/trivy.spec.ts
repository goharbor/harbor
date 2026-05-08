import {
    expect,
    test,
    type APIRequestContext,
    type Locator,
    type Page,
} from '@playwright/test';
import {
    execFileSync,
    type ExecFileSyncOptionsWithStringEncoding,
} from 'child_process';

const LOCAL_REGISTRY: string =
    process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string =
    process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip = process.env.IP || '';
const user = process.env.HARBOR_ADMIN || 'admin';
const pwd =
    process.env.HARBOR_PASSWORD ||
    process.env.HARBOR_ADMIN_PASSWD ||
    'Harbor12345';

test('shows scanned artifact vulnerabilities in Security Hub', async ({
    page,
    request,
}) => {
    test.setTimeout(60 * 60 * 1000);
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

    await login(page);
    await createProject(page, project);

    for (const image of images) {
        pushImageWithTag(ip, user, pwd, project, image, tag, tag);
        await scanRepository(page, project, image);
    }

    pushManifestList(ip, user, pwd, `${ip}/${project}/${indexRepo}:${tag}`, [
        `${ip}/${project}/${images[0]}:${tag}`,
        `${ip}/${project}/${images[1]}:${tag}`,
    ]);

    for (const image of images.slice(0, 2)) {
        await deleteRepository(page, project, image);
    }

    await scanRepository(page, project, indexRepo);
    await openSecurityHub(page);

    const summary = await getVulnerabilitySummaryFromAPI(request, user, pwd);
    expect(summary.dangerous_artifacts.length).toBeGreaterThanOrEqual(2);
    expect(summary.dangerous_cves.length).toBeGreaterThanOrEqual(1);

    const dangerousArtifact = summary.dangerous_artifacts[0];
    const secondDangerousArtifact = summary.dangerous_artifacts[1];
    const dangerousCve = summary.dangerous_cves[0];
    const projectId = await getProjectIdFromAPI(request, user, pwd, project);
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

    await assertQuickSearchByArtifact(page, dangerousArtifact);
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

    await searchBySeverity(page, 'Unknown');
    await expectNoVulnerabilities(page);
    await searchBySeverity(page, 'None');
    await expectNoVulnerabilities(page);

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

    await openSecurityHub(page);
    await expectRepositoryJump(
        page,
        dangerousArtifact.repository_name,
        indexRepo
    );

    await openSecurityHub(page);
    await expectSecurityHubDigestJump(page, dangerousArtifact.digest);

    await openSecurityHub(page);
    await expectSummaryDigestJump(page, dangerousArtifact.digest);

    await openSecurityHub(page);
    await expectSummaryDigestJump(page, secondDangerousArtifact.digest);

    await deleteRepository(page, project, indexRepo);
    await openSecurityHub(page);
    await expect(vulnerabilitySummary(page)).not.toContainText(
        `${project}/${indexRepo}`,
        { timeout: 60000 }
    );

    await searchByTextFilter(
        page,
        'repository_name',
        `${project}/${indexRepo}`
    );
    await expectNoVulnerabilities(page);

    await logout(page);
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

const scanResultTimeout = 5 * 60 * 1000;

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
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).fill(user);
    await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
    await page.getByRole('button', { name: 'LOG IN' }).click();
}

async function logout(page: Page): Promise<void> {
    await page.goto('/');
    const userMenu = page.getByRole('button', { name: user, exact: true });
    await expect(userMenu).toBeVisible();
    await userMenu.click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
}

async function createProject(page: Page, project: string): Promise<void> {
    await page.getByRole('button', { name: 'New Project' }).click();
    await page.locator('#create_project_name').fill(project);
    await page.getByRole('button', { name: 'OK' }).click();
}

async function openProject(page: Page, project: string): Promise<void> {
    await page.goto('/');
    await page.getByRole('link', { name: project }).click();
}

async function scanRepository(
    page: Page,
    project: string,
    repository: string
): Promise<void> {
    await openProject(page, project);
    await page.getByRole('link', { name: `${project}/${repository}` }).click();
    await scanCurrentArtifact(page);
}

async function scanCurrentArtifact(page: Page): Promise<void> {
    await page
        .getByRole('gridcell', { name: 'Select Select' })
        .locator('label')
        .click();
    await page.getByRole('checkbox', { name: 'Select', exact: true }).check();

    const scanButton = page.getByRole('button', { name: 'Scan vulnerability' });
    await expect(scanButton).toBeEnabled();
    await scanButton.click();
    await page
        .getByRole('gridcell', { name: /Total/ })
        .waitFor({ timeout: scanResultTimeout });
}

async function deleteRepository(
    page: Page,
    project: string,
    repository: string
): Promise<void> {
    await openProject(page, project);
    await page.locator('.refresh-btn > clr-icon').click();

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
    await expect(page.locator('clr-dg-footer')).toContainText(
        count > 1000 ? '1000+ CVEs' : `${count} CVEs`
    );
    await expectGridCellVisible(page, severity);
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
    await expect(page.locator('clr-dg-placeholder')).toContainText(
        'We could not find any vulnerability'
    );
}

async function expectRepositoryJump(
    page: Page,
    repositoryName: string,
    expectedHeading: string
): Promise<void> {
    const repositoryLink = page
        .getByRole('link', { name: repositoryName })
        .last();
    await expect(repositoryLink).toBeVisible();
    await repositoryLink.click();

    const repositoryDetailLink = page
        .getByRole('link', { name: repositoryName })
        .last();
    await expect(repositoryDetailLink).toBeVisible();
    await repositoryDetailLink.click();
    await expect(
        page.locator('h2').filter({ hasText: expectedHeading })
    ).toBeVisible();
}

async function expectSecurityHubDigestJump(
    page: Page,
    digest: string
): Promise<void> {
    const digestText = shortDigest(digest);
    const digestLink = page.getByRole('link', { name: digestText }).last();
    await expect(digestLink).toBeVisible();
    await digestLink.click();
    await expect(
        page.locator('h2').filter({ hasText: digestText })
    ).toBeVisible();
}

async function expectSummaryDigestJump(
    page: Page,
    digest: string
): Promise<void> {
    const digestText = shortDigest(digest);
    await expect(vulnerabilitySummary(page)).toContainText(digestText);
    await vulnerabilitySummary(page)
        .getByRole('link', { name: digestText })
        .click();
    await expect(
        page.locator('h2').filter({ hasText: digestText })
    ).toBeVisible();
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

function runDockerCommand(
    args: string[],
    options: Partial<ExecFileSyncOptionsWithStringEncoding> = {}
): string {
    try {
        return execFileSync('docker', args, {
            encoding: 'utf-8',
            stdio: ['pipe', 'pipe', 'pipe'],
            ...options,
        });
    } catch {
        throw new Error(`Docker command failed: ${args[0]}`);
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
    tag1: string = 'latest'
): void {
    const sourceImage = `${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1}`;
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
