
import { test, expect, Page } from '@playwright/test';
import { execSync } from 'child_process';

// variables
const LOCAL_REGISTRY: string = process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string = process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip: string = process.env.IP;
const user: string = process.env.HARBOR_ADMIN || 'admin';
const pwd: string = process.env.HARBOR_PASSWORD || 'Harbor12345';

test('Trivy is default scanner and it is immutable', async ({ page }) => {
  test.setTimeout(60 * 1000); //1 minutes
  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // switch to scanners page
  await page.getByText(' Interrogation Services ').click();
  await page.getByRole('link', {name: 'Scanners', exact: true}).click();
  await expect(page.getByRole('button', {name: ' SET AS DEFAULT '})).toBeVisible();

  // Should display default trivy scanner
  await expect(page.getByRole('gridcell', {name: 'Trivy Default'})).toBeVisible();

  // Trivy scanner should be immutable
  const trivyRow = page.getByRole('row').filter({ 
    has: page.getByRole('gridcell', { name: 'Trivy' }) 
  });
  await expect(trivyRow).toBeVisible();
  await trivyRow.getByRole('gridcell', {name: 'Select'}).click();
  await page.locator('#action-scanner').click();
  await page.getByRole('menuitem', {name: 'DELETE'}).click();
  await page.getByRole('button', {name: 'DELETE'}).click();
  await expect(page.getByText("registration Trivy is not allowed to delete as it is immutable: scanner API: delete")).toBeVisible();
})

test('Disable Scan Schedule', async ({ page }) => {
  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // switch to vulnerabilities page
  await page.getByText(' Interrogation Services ').click();
  await page.getByRole('link', {name: 'Vulnerability', exact: true}).click();
  await expect(page.getByRole('button', {name: 'SCAN NOW'})).toBeVisible();

  // Disable Scan Schedule
  await page.getByRole('button', {name: 'EDIT'}).click();
  const selectPolicyDropDown = page.getByRole('combobox');
  await selectPolicyDropDown.click();
  await selectPolicyDropDown.selectOption({value: 'None'});
  await page.getByRole('button', {name: 'SAVE'}).click();

  // logout
  await page.goto('/');
  await page.getByRole('button', { name: 'admin', exact: true }).waitFor();
  await page.getByRole('button', { name: 'admin', exact: true }).click();
  await page.getByRole('menuitem', { name: 'Log Out' }).click();

  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // switch to vulnerabilities page
  await page.getByText(' Interrogation Services ').click();
  await page.getByRole('link', {name: 'Vulnerability', exact: true}).click();
  await expect(page.getByRole('button', {name: 'SCAN NOW'})).toBeVisible();

  // check that scan schedule is disabled
  await expect(page.getByText('None', {exact: true})).toBeVisible();
})

test('login and scan the things', async ({ page }) => {
  test.setTimeout(60 * 60 * 1000); // 60 minutes
const tag = 'v2.2.0';
const digest = 'sha256:7c3f03db32f9a89b47faedb69cb6ea10741cec203ec76eb45add65e58baa2a82';
const cve_id = 'CVE-2021-22926';
const packageName = 'curl';
const cvss_score_v3_from = 6.5;
const cvss_score_v3_to = 7.5;
const severity = 'High';
const d = '1';
const index_repo = `index${d}`;
const cve_description = 'Description: libcurl-using applications can ask for a specific client certificate to be used in a transfer';

// now we need to tag and push these images to harbor.
  const images = [
  'goharbor/harbor-log-base',
  'goharbor/harbor-prepare-base',
  'goharbor/harbor-redis-base',
  'goharbor/harbor-nginx-base',
  'goharbor/harbor-registry-base'
];

const project: string = 'aproject-'+ Date.now();
    // login
    await page.goto('/');
    await page.getByRole('textbox', { name: 'Username' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill('admin');
    await page.getByRole('textbox', { name: 'Password' }).click();
    await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
    await page.getByRole('button', { name: 'LOG IN' }).click();

  // create project
  await page.getByRole('button', { name: 'New Project' }).click();
  await page.locator('#create_project_name').click();
  await page.locator('#create_project_name').fill(project);
  await page.getByRole('button', { name: 'OK' }).click();

for (const image of images) {

  pushImageWithTag(ip, user, pwd, project, image, tag, tag);

  // go into project repo page to verify the image is there.
  await page.getByRole('link', { name: project }).click();
  await page.getByRole('link', { name: project + '/' + image }).click();


  // scan the repo
//   const tagname = process.env.TAG_NAME || 'v2.2.0';
// const row = page.getByRole('row', { name: new RegExp(tagname) });
await page.waitForTimeout(1000);
await page.getByRole('gridcell', { name: 'Select Select' }).locator('label').click();
await page.waitForTimeout(1000);
await page.getByRole('checkbox', { name: 'Select', exact: true }).check();
await page.waitForTimeout(5000);
await page.getByRole('button', { name: 'Scan vulnerability' }).click();
await page.getByRole('gridcell', { name: /Total/ }).waitFor();


await page.goto('/');
}

  // // get back to the project page
  // await page.getByText(project).click();


const command = `./e2e/scripts/docker_push_manifest_list.sh \
  ${ip} ${user} ${pwd} \
  "${ip}/${project}/${index_repo}:${tag}" \
  "${ip}/${project}/${images[0]}:${tag}" \
  "${ip}/${project}/${images[1]}:${tag}"`;

  const output = runCommand(command);

  expect(output).not.toContain('Error');

// delete first two repos
for (let i = 0; i < 2; i++) {
  const image = images[i];
    await page.goto('/');
    await page.getByRole('link', { name: project }).click();
    await page.locator('.refresh-btn > clr-icon').click();

const rowRegex = new RegExp(`Select\\s+Select\\s+${project}/${image}`, 'i');

// Wait for row to appear and click its checkbox (label)
const row = page.getByRole('row', { name: rowRegex });
await row.waitFor({ state: 'visible', timeout: 10000 }); // ensure it's visible
await row.locator('label').click();
await page.waitForTimeout(5000);
await page.getByRole('button', { name: 'Delete' }).click();
await page.getByRole('button', { name: 'DELETE', exact: true }).click();
}

// go into the index repo and scan the manifest list
await page.getByRole('link', { name: project + '/' + index_repo }).click();
//scan the repo
await page.waitForTimeout(1000);
await page.getByRole('gridcell', { name: 'Select Select' }).locator('label').click();
await page.waitForTimeout(1000);
await page.getByRole('checkbox', { name: 'Select', exact: true }).check();
await page.waitForTimeout(5000);
await page.getByRole('button', { name: 'Scan vulnerability' }).click();
await page.getByRole('gridcell', { name: /Total/ }).waitFor();

// go to security hub
await page.getByRole('link', { name: 'Interrogation Services' }).click();
await page.getByRole('link', { name: 'Security Hub' }).click();
// await page.getByRole('button', { name: 'SEARCH' }).click();

// get vuln summary from api and compare it with the ui display
const summary = await getVulnerabilitySummaryFromAPI(ip, user, pwd);
console.log('Vulnerability Summary from API:', summary);
  // Map expected counts
  const expectedCounts = [
    summary.critical_cnt, // 1st div
    summary.high_cnt,     // 2nd div
    summary.medium_cnt,   // 3rd div
    summary.low_cnt,      // 4th div
    summary.unknown_cnt,                    // 5th div
    0,                    // 6th div
  ];

  // Loop through and verify UI elements
  for (let i = 0; i < expectedCounts.length - 2; i++) {
    await page.waitForTimeout(1000);
    console.log(`Verifying count for severity index ${i}: Expected ${expectedCounts[i]}`);
    await expect(page.locator('app-vulnerability-summary')).toContainText(`${expectedCounts[i]}`);
  }

  // check the top 5 dangerous artifacts
  const dangerousArtifacts = summary.dangerous_artifacts;

await page.waitForTimeout(2000);
for (const artifact of dangerousArtifacts) {
  // repository name
  await expect(page.locator('app-vulnerability-summary')).toContainText(artifact.repository_name);

  // shortened digest (first few chars for parity with UI)
  const shortDigest = artifact.digest.slice(0, 14); // e.g. sha256:f4215ab2
  await expect(page.locator('app-vulnerability-summary')).toContainText(shortDigest);

  // log for debug visibility
  console.log(`✅ Verified artifact: ${artifact.repository_name} (${shortDigest})`);
}
await page.waitForTimeout(2000);

// // check the top 5 dangerous CVEs
const dangerousCVEs = summary.dangerous_cves;
console.log('Dangerous CVEs from API:', dangerousCVEs);

for (const cve of dangerousCVEs) {
  const pkgVersion = `${cve.package}@${cve.version}`;

  // ✅ check dynamically for each CVE’s values
  await expect(page.locator('app-vulnerability-summary')).toContainText(cve.cve_id);
  await expect(page.locator('app-vulnerability-summary')).toContainText(cve.severity);
  await expect(page.locator('app-vulnerability-summary')).toContainText(String(cve.cvss_score_v3));
  await expect(page.locator('app-vulnerability-summary')).toContainText(pkgVersion);

  console.log(`✅ Verified CVE ${cve.cve_id} (${cve.severity}) - ${pkgVersion}`);
}

// check the quick search
  // select the first repo name on the right side - dangerous artifacts
  await page.locator('app-vulnerability-summary').getByRole('link', { name: summary.dangerous_cves[0].repository_name }).first().click();
  // check if the below element got the right repo and digest
  console.log('Checking quick search values for repo:', dangerousCVEs[0].repository_name);
  console.log('Checking quick search values for repo:', summary.dangerous_cves[0].repository_name);
  console.log('checking value of dangerous arts:', summary.dangerous_artifacts[0].repository_name);

  await expect(page.locator('app-vulnerability-filter form div').filter({ hasText: 'Filter by All Repository Name' }).getByRole('textbox')).toHaveValue(summary.dangerous_artifacts[0].repository_name);
  await expect(page.getByRole('textbox').nth(2)).toHaveValue(summary.dangerous_artifacts[0].digest);
  // check if the table shows the right info
  await page.locator('.datagrid-inner-wrapper').click();
  await page.waitForTimeout(3000);
  // await expect(page.locator('#clr-dg-row33')).toContainText('CVE-2021-37600');
  await expect(page.getByText(summary.dangerous_artifacts[0].repository_name).nth(2)).toBeVisible(); // this works no need for fuzzy i guess
  // const repo = summary.dangerous_cves[0].cve_id;
  // // const repo = summary.dangerous_artifacts[0].repository_name;:w

  // // create fuzzy version (partial match)
  // const fuzzyRepo = new RegExp(repo.replace(/[.*+?^${}()|[\]\\]/g, ''), 'i');

  // // check visibility using fuzzy matching
  // await expect(page.getByText(fuzzyRepo).first()).toBeVisible();

  // await expect(page.locator('#clr-dg-row33')).toContainText('CVE-2021-37600');
  // await expect(page.locator('#clr-dg-row33')).toContainText(summary.dangerous_artifacts[0].repository_name);
  await page.waitForTimeout(3000);

  await expect(page.getByRole('gridcell', { name: summary.dangerous_artifacts[0].repository_name }).first()).toBeVisible();
  await expect(page.getByRole('gridcell', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) }).first()).toBeVisible();
  // await expect(page.getByRole('gridcell', { name: summary.dangerous_cves[0].version }).first()).toBeVisible();
  // await expect(page.getByRole('gridcell', { name: summary.dangerous_cves[0].cvss_score_v3.toString() }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row33')).toContainText(summary.dangerous_artifacts[0].digest.substring(0, 12));
  // previously the below 2 lines were using cves
  // await expect(page.locator('#clr-dg-row33')).toContainText(summary.dangerous_cves[0].version);
  // await expect(page.locator('#clr-dg-row33')).toContainText(summary.dangerous_cves[0].cvss_score_v3.toString());
 await page.waitForTimeout(2000);

 // check for the cve id
//  // ---  remove this once test passes -- start
//   await page.goto('/');
//   // login
//   await page.goto('/');
//   await page.getByRole('textbox', { name: 'Username' }).click();
//   await page.getByRole('textbox', { name: 'Username' }).fill('admin');
//   await page.getByRole('textbox', { name: 'Password' }).click();
//   await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
//   await page.getByRole('button', { name: 'LOG IN' }).click();
//   await page.getByRole('link', { name: 'Interrogation Services' }).click();
//   // --- remove this once test passes -- end

  // search for dangerous cves
  await page.getByRole('link', { name: summary.dangerous_cves[0].cve_id }).first().click();
  await page.waitForTimeout(2000);
  // await page.getByText('Top 5 Most Dangerous CVEs CVE').click();
  await page.getByText('Top 5 Most Dangerous CVEs').click();
  const value = await page.locator('div:nth-child(3) > .card-block > div > div').first();
  // console.log("what the hell", value.textContent());
  // console.log("what is this", value);

  // TODO: this should be dynamic
  await page.getByRole('link', { name: dangerousCVEs[0].cve_id }).first().click();
  await expect(page.locator('app-vulnerability-filter').getByRole('textbox')).toHaveValue(dangerousCVEs[0].cve_id);
  // await page.locator('.datagrid-inner-wrapper').click();
  await expect(page.locator('app-vulnerability-filter').getByRole('combobox')).toHaveValue('cve_id');
  // await page.locator('.datagrid').click();
  await expect(page.getByRole('gridcell', { name: summary.dangerous_cves[0].cve_id }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row31')).toContainText(dangerousCVEs[0].cve_id);
  await page.waitForTimeout(2000);
  // quick search done ---

  // check the search by one condition
  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('project_id');
  await page.locator('app-vulnerability-filter').getByRole('textbox').dblclick();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(project);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  await expect(page.getByRole('gridcell', { name: project }).first()).toBeVisible();

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('repository_name');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(`${project}/${images[2]}`);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  // await expect(page.locator('#clr-dg-row58')).toContainText(`${project}/${images[2]}`);
  await expect(page.getByRole('gridcell', { name: `${project}/${images[2]}` }).first()).toBeVisible();

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('digest');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(digest);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  const shortDigest = digest.slice(0, 14);
  await expect(page.getByRole('gridcell', { name: `${shortDigest}` }).first()).toBeVisible();

  // search by cve id
  await page.waitForTimeout(2000);
  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('cve_id');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(cve_id);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  // await expect(page.locator('#clr-dg-row118')).toContainText('CVE-2022-29155');
  await expect(page.getByRole('gridcell', { name: `${cve_id}` }).first()).toBeVisible();

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('package');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill('curl');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  await expect(page.getByRole('gridcell', { name: "curl" }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row129')).toContainText('curl');

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('tag');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(tag);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  await expect(page.getByRole('gridcell', { name: tag }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row144')).toContainText('v2.2.0');
  // await page.getByText('CVE-2022-32207aproject-1764110949942/goharbor/harbor-redis-basesha256:').click();
  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('cvss_score_v3');
  await page.getByRole('textbox').nth(1).click();
  await page.getByRole('textbox').nth(1).fill('6.5');
  await page.getByRole('textbox').nth(2).click();
  await page.getByRole('textbox').nth(2).fill('7.5');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  // await expect(page.locator('#clr-dg-row159')).toContainText('7.5');
  await expect(page.getByRole('gridcell', { name: '7.5' }).first()).toBeVisible();

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('severity');
  await page.getByRole('combobox').nth(1).selectOption('Critical');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  if (summary.critical_cnt > 1000) {
    await expect(page.locator('clr-dg-footer')).toContainText('1000+ CVEs');
  } else {
    await expect(page.locator('clr-dg-footer')).toContainText(summary.critical_cnt + ' CVEs');
  }
  await expect(page.getByRole('gridcell', { name: 'Critical' }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row174')).toContainText('Critical');
  await page.getByRole('combobox').nth(1).selectOption('High');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  if (summary.high_cnt > 1000) {
    await expect(page.locator('clr-dg-footer')).toContainText('1000+ CVEs');
  } else {
    await expect(page.locator('clr-dg-footer')).toContainText(summary.high_cnt + ' CVEs');
  }
  // await expect(page.locator('clr-dg-footer')).toContainText(summary.high_cnt + ' CVEs');
  await expect(page.getByRole('gridcell', { name: 'High' }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row189')).toContainText('High');
  await page.getByRole('combobox').nth(1).selectOption('Medium');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  if (summary.medium_cnt > 1000) {
    await expect(page.locator('clr-dg-footer')).toContainText('1000+ CVEs');
  } else {
    await expect(page.locator('clr-dg-footer')).toContainText(summary.medium_cnt + ' CVEs');
  }
  await expect(page.getByRole('gridcell', { name: 'Medium' }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row204')).toContainText('Medium');
  await page.getByRole('combobox').nth(1).selectOption('Low');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  if (summary.low_cnt > 1000) {
    await expect(page.locator('clr-dg-footer')).toContainText('1000+ CVEs');
  } else {
    await expect(page.locator('clr-dg-footer')).toContainText(summary.low_cnt + ' CVEs');
  }
  await expect(page.getByRole('gridcell', { name: 'Low' }).first()).toBeVisible();
  // await expect(page.locator('#clr-dg-row219')).toContainText('Low');
  await page.getByRole('combobox').nth(1).selectOption('Unknown');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  await expect(page.locator('clr-dg-placeholder')).toContainText('We could not find any vulnerability');
  await page.getByRole('combobox').nth(1).selectOption('None');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.waitForTimeout(2000);
  await expect(page.locator('clr-dg-placeholder')).toContainText('We could not find any vulnerability');

  // search by multiple conditions
  await page.getByRole('combobox').first().selectOption('project_id');
  await page.locator('app-vulnerability-filter').getByRole('textbox').dblclick();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(project);
  await expect(page.locator('app-vulnerability-filter clr-icon').nth(2)).toBeVisible();
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.getByRole('combobox').nth(1).selectOption('repository_name');
  await page.getByRole('textbox').nth(2).dblclick();
  await page.getByRole('textbox').nth(2).fill(`${project}/${images[2]}`);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.getByRole('combobox').nth(2).selectOption('digest');
  await page.getByRole('textbox').nth(3).fill(digest);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.locator('div:nth-child(4) > .clr-control-container > div > .clr-select').first().selectOption('cve_id');
  await page.locator('app-vulnerability-filter').getByRole('textbox').nth(3).dblclick();
  await page.locator('app-vulnerability-filter').getByRole('textbox').nth(3).fill(cve_id);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.locator('div:nth-child(5) > .clr-control-container > div > .clr-select').first().selectOption('package');
  await page.locator('app-vulnerability-filter').getByRole('textbox').nth(4).dblclick();
  await page.locator('app-vulnerability-filter').getByRole('textbox').nth(4).fill('curl');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.locator('div:nth-child(6) > .clr-control-container > div > .clr-select').first().selectOption('tag');
  await page.locator('.clr-input.ng-untouched').dblclick();
  await page.locator('.clr-input.ng-untouched').fill(tag);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.locator('div:nth-child(7) > .clr-control-container > div > .clr-select').first().selectOption('cvss_score_v3');
  await page.locator('div').filter({ hasText: /^FromTo$/ }).getByRole('textbox').first().dblclick();
  await page.locator('div').filter({ hasText: /^FromTo$/ }).getByRole('textbox').first().fill('6.5');
  await page.locator('div').filter({ hasText: /^FromTo$/ }).getByRole('textbox').nth(1).dblclick();
  await page.locator('div').filter({ hasText: /^FromTo$/ }).getByRole('textbox').nth(1).fill('7.5');
  await page.locator('app-vulnerability-filter clr-icon').nth(1).click();
  await page.locator('.clr-select-wrapper.ml-1 > .clr-select').selectOption('High');
  await page.getByRole('button', { name: 'SEARCH' }).click();
  // await expect(page.locator('#clr-dg-row281')).toContainText('High');
  await expect(page.getByRole('gridcell', { name: 'High' }).first()).toBeVisible();

  await page.getByRole('button', { name: 'Open' }).first().click();
  // await expect(page.locator('#clr-dg-expandable-row-281')).toContainText('Description: libcurl-using applications can ask for a specific client certificate to be used in a transfer');
  // await page.getByRole('gridcell', { name: 'Close' }).click();
  await expect(page.locator('clr-datagrid')).toContainText(cve_description);

// do page jump test
  // await page.getByRole('link', { name: 'Projects' }).click();
  // await page.getByRole('link', { name: 'Interrogation Services' }).click();
  // await page.getByText('CVE-2022-22823').click();
  await page.getByRole('link', { name: cve_id }).first().click();
  await page.waitForTimeout(2000);

  // do repo jump test
  await page.goto('/');
  await page.getByRole('link', { name: 'Interrogation Services' }).click();
  // await page.getByRole('link', { name: 'Security Hub' }).click();
  // await page.locator('#clr-dg-row317').getByRole('link', { name: 'aproject-1764149771585/' }).click();
  // await expect(page.locator('h2')).toContainText('goharbor/harbor-registry-base');
  await expect(page.getByRole('link', { name: summary.dangerous_artifacts[0].repository_name }).last()).toBeVisible();
  await page.getByRole('link', { name: summary.dangerous_artifacts[0].repository_name }).last().click();
  await page.waitForTimeout(1000);
  await page.getByRole('link', { name: summary.dangerous_artifacts[0].repository_name }).last().click();
  // await page.getByRole('link', { name: summary.dangerous_artifacts[0].repository_name }).last().click();
  await expect(page.locator('h2')).toContainText(index_repo);

  // do digest jump test
  await page.goto('/');
  await page.getByRole('link', { name: 'Interrogation Services' }).click();
  // await page.getByRole('link', { name: 'Security Hub' }).click();
  await expect(page.getByRole('link', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) }).last()).toBeVisible();
  await page.getByRole('link', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) }).last().click();
  // await page.getByRole('gridcell', { name: 'sha256:3f42abf2' }).click();
  await page.waitForTimeout(1000);
  // await page.getByRole('link', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) }).last().click();
  // await expect(page.getByRole('link', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) })).toBeVisible();
  await expect(page.locator('h2')).toContainText(summary.dangerous_artifacts[0].digest.substring(0, 12));

  // top 5 dangerous artifacts jump test
  await page.goto('/');
  await page.getByRole('link', { name: 'Interrogation Services' }).click();
  await expect(page.locator('app-vulnerability-summary')).toContainText(summary.dangerous_artifacts[0].digest.substring(0, 12));
  await page.getByRole('link', { name: summary.dangerous_artifacts[0].digest.substring(0, 12) }).last().click();
  await page.waitForTimeout(1000);
  await expect(page.locator('h2')).toContainText(summary.dangerous_artifacts[0].digest.substring(0, 12));

  // top 5 dangerous artifacts jump test 2
  await page.goto('/');
  await page.getByRole('link', { name: 'Interrogation Services' }).click();
  await expect(page.locator('app-vulnerability-summary')).toContainText(summary.dangerous_artifacts[1].digest.substring(0, 12));
  await page.getByRole('link', { name: summary.dangerous_artifacts[1].digest.substring(0, 12) }).last().click();
  await page.waitForTimeout(1000);
  await expect(page.locator('h2')).toContainText(summary.dangerous_artifacts[1].digest.substring(0, 12));
  // -- jump tests done --
  await page.goto('/');
  await page.getByRole('link', { name: 'Interrogation Services' }).click();

  // Check that there is no such artifact in the security hub after deleting the artifact
  // delete index repo and then check security hub

  await page.getByRole('link', { name: 'Projects' }).click();
  await page.getByRole('link', { name: project }).click();
  // await expect(page.locator('#clr-dg-row381')).toContainText(project + '/' + index_repo);
  // delete index repo
  const rowRegex = new RegExp(`Select\\s+Select\\s+${project}/${index_repo}`, 'i');

// Wait for row to appear and click its checkbox (label)
const row = page.getByRole('row', { name: rowRegex });
await row.waitFor({ state: 'visible', timeout: 10000 }); // ensure it's visible
await row.locator('label').click();
await page.waitForTimeout(5000);
await page.getByRole('button', { name: 'Delete' }).click();
await page.getByRole('button', { name: 'DELETE', exact: true }).click();

  // await page.getByRole('row', { name: 'Select Select aproject-1764149771585/index1 1 4 11/26/25, 3:07 PM' }).locator('label').click();
  // await page.getByRole('button', { name: 'Delete' }).click();
  // await page.getByRole('button', { name: 'DELETE', exact: true }).click();

  await page.getByRole('link', { name: 'Interrogation Services' }).click();
  await expect(page.locator('app-vulnerability-summary')).not.toContainText(project + '/' + index_repo);

  await page.locator('app-vulnerability-filter').getByRole('combobox').selectOption('repository_name');
  await page.locator('app-vulnerability-filter').getByRole('textbox').click();
  await page.locator('app-vulnerability-filter').getByRole('textbox').fill(project + '/' + index_repo);
  await page.getByRole('button', { name: 'SEARCH' }).click();
  await expect(page.locator('clr-dg-placeholder')).toContainText('We could not find any vulnerability');

  // await page.getByRole('row', { name: /Select Select ${project} + '/' + ${image}/ }).locator('label').click();
    // logout
    await page.goto('/');
    await page.getByRole('button', { name: 'admin', exact: true }).waitFor();
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});

test('Manual Scan All', async ({ page }) => {
  test.setTimeout(3 * 60 * 1000); // 3 minuites
  const project = `project${Date.now()}`
  const image = 'redis'
  const sha256 = 'e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c';

  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // create a project
  await page.getByRole('button', { name: 'New Project' }).click();
  await page.locator('#create_project_name').click();
  await page.locator('#create_project_name').fill(project);
  await page.getByRole('button', { name: 'OK' }).click();

  await pushImage(
      ip,
      user,
      pwd,
      project,
      image, 
      sha256
    )

  // switch to vulnerabilities page
  await page.getByRole('link', {name: 'Interrogation Services'}).click();
  await page.getByRole('link', {name: 'Vulnerability', exact: true}).click();
  await expect(page.getByRole('button', {name: 'SCAN NOW'})).toBeVisible(); 

  // Start scan and wait for it to be finished
  await page.getByRole('button', {name: 'SCAN NOW'}).click();
  await expect(page.getByRole('button', {name: 'SCAN NOW'})).toBeEnabled({timeout: 2 * 60 * 1000});

  // go into the repo
  await page.getByRole('link', {name: 'Projects'}).click();
  await page.getByRole('link', {name: project}).click();
  await page.getByRole('link', {name: project + '/' + image}).click();

  // Check that the vulnerabilites are visible in list row
  await scanResultShouldDisplayInListRow(page, sha256);

  // Check the repo scan details
  await page.getByRole('link', {name: 'sha256'}).click();
  await viewRepoScanDetails(page, ['Critical']);
})

test('Scan a Tag in the Repo', async ({ page }) => {
  test.setTimeout(3 * 60 * 1000); //3 minuites
  const project = `project${Date.now()}`;
  const imageNamespace = 'vmware';
  const image = 'photon';
  const tag = '1.0';
  
  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // create a project
  await page.getByRole('button', { name: 'New Project' }).click();
  await page.locator('#create_project_name').click();
  await page.locator('#create_project_name').fill(project);
  await page.getByRole('button', { name: 'OK' }).click();

  await pushImageWithTag(
    ip,
    user,
    pwd,
    project,
    image,
    tag,
    'latest',
    'docker.io',
    imageNamespace,
  );

  // go into the repo
  await page.getByRole('link', {name: 'Projects'}).click();
  await page.getByRole('link', {name: project}).click();
  await page.getByRole('link', {name: project + '/' + image}).click();

  //scan the repo
  await page.waitForTimeout(1000);
  await page.getByRole('gridcell', { name: 'Select Select' }).locator('label').click();
  await page.waitForTimeout(1000);
  await page.getByRole('checkbox', { name: 'Select', exact: true }).check();
  await page.waitForTimeout(5000);
  await page.getByRole('button', { name: 'Scan vulnerability' }).click();
  await page.getByRole('gridcell', { name: /Total/ }).waitFor();

  await scanResultShouldDisplayInListRow(page, tag);
  
  // pull image 
  await pullImage(ip, user, pwd, project, image, tag);
})

test('Scan As An Unprivileged User', async ({ page }) => {
  test.setTimeout(2 * 60 * 1000); //5 mins
  const project = `project${Date.now()}`;
  const image = 'hello-world';

  // login to harbor
  await page.goto('/');
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('admin');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Harbor12345');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // create a project(public)
  await page.getByRole('button', { name: 'New Project' }).click();
  await page.locator('#create_project_name').click();
  await page.locator('#create_project_name').fill(project);
  await page.locator("input[name='public'] ~ label.clr-control-label").check();
  await page.getByRole('button', { name: 'OK' }).click();

  await pushImage(ip, user, pwd, project, image);

  // logout
  await page.goto('/');
  await page.getByRole('button', { name: 'admin', exact: true }).waitFor();
  await page.getByRole('button', { name: 'admin', exact: true }).click();
  await page.getByRole('menuitem', { name: 'Log Out' }).click();

  // create a new user (unpriveledged)
  await page.goto('/');
  await page.getByRole('link', {name: 'Sign up for an account'}).click();
  await page.locator('#username').click();
  await page.locator('#username').fill('User@01');
  await page.locator('#email').click();
  await page.locator('#email').fill('test_01@example.com');
  await page.getByRole('textbox', { name: 'First and last name*' }).click();
  await page.getByRole('textbox', { name: 'First and last name*' }).fill('User01');
  await page.getByRole('textbox', { name: 'Password*', exact: true}).click();
  await page.getByRole('textbox', { name: 'Password*', exact: true}).fill('Test1@01');
  await page.getByRole('textbox', { name: 'Confirm Password*' }).click();
  await page.getByRole('textbox', { name: 'Confirm Password*' }).fill('Test1@01');
  await page.getByRole('button', { name: 'SIGN UP' }).click();

  // login to the new user
  await page.getByRole('textbox', { name: 'Username' }).click();
  await page.getByRole('textbox', { name: 'Username' }).fill('User@01');
  await page.getByRole('textbox', { name: 'Password' }).click();
  await page.getByRole('textbox', { name: 'Password' }).fill('Test1@01');
  await page.getByRole('button', { name: 'LOG IN' }).click();

  // go into the repo
  await page.getByRole('link', {name: 'Projects'}).click();
  await page.getByRole('link', {name: project}).click();
  await page.getByRole('link', {name: project + '/' + image}).click();

  // Select artifact
  await page.waitForTimeout(1000);
  await page.getByRole('gridcell', { name: 'Select Select' }).locator('label').click();
  await page.waitForTimeout(1000);
  await page.getByRole('checkbox', { name: 'Select', exact: true }).check();

  // Scan should be disabled
  await expect(page.getByRole('button', { name: 'Scan vulnerability' })).toBeDisabled();

})


/* PLAYWRIGHT UTILITY FUNCTIONS */

async function viewRepoScanDetails(page: Page, vuln_levels: string[]) {
  const vuln_table = page.locator('hbr-artifact-vulnerabilities');
  await expect(vuln_table).toBeVisible({timeout : 10000});
  for(const vuln_level of vuln_levels) {
    const row = vuln_table
      .getByRole('row')
      .filter({hasText: vuln_level})
      .first();

      await row.scrollIntoViewIfNeeded();
      await expect(row).toBeVisible();
      console.log(`checked for ${vuln_level}`);
    }
  await page.getByRole('tab', {name: ' Build History '}).click();
}

async function scanResultShouldDisplayInListRow(
  page: Page, 
  tagOrDigest: string, 
  hasNoVulnerability: boolean = false
): Promise<void> {
  
  if (hasNoVulnerability) {
    // Look for row containing both 'No vulnerability' and the tag/digest and tooltip
    const vuln_cell = page.getByRole('row')
      .filter({ hasText: tagOrDigest })
      .getByRole('gridcell', { name: 'No vulnerability' });
    
    await expect(vuln_cell).toBeVisible();
    await vuln_cell.hover({timeout: 5000});
    const tooltip = vuln_cell.getByRole('tooltip');
    await expect(tooltip).toBeVisible({ timeout: 10000 });
    
  } else {
    // Case 2: Image HAS vulnerabilities
    const vuln_cell = page.getByRole('row')
      .filter({ hasText: tagOrDigest })
      .getByRole('gridcell', { name: /Total.*Fixable/ })
    
    // Verify the vulnerability summary is visible in the row and the tooltip
    await expect(vuln_cell).toBeVisible();
    await vuln_cell.hover({timeout: 5000});
    const tooltip = vuln_cell.getByRole('tooltip');
    await expect(tooltip).toBeVisible({ timeout: 10000 });
  }
}


// /**
//  * Executes a shell command and waits until it succeeds.
//  * Throws an error if the command fails.
//  */
// function runCommand(command: string): void {
//   console.log(`\n$ ${command}`);
//   try {
//     const output = execSync(command, { stdio: 'inherit' }); // inherit = live logs
//     console.log(output?.toString() || '');
//     } catch (error) {
//     console.error(`❌ Command failed: ${command}`);
//     throw error;
//   }
// }
export async function getVulnerabilitySummaryFromAPI(ip, user, password) {
  // Encode credentials for Basic Auth
  const credentials = Buffer.from(`${user}:${password}`).toString('base64');

  // API endpoint (mirrors your Robot curl command)
  const url = `https://${ip}/api/v2.0/security/summary?with_dangerous_cve=true&with_dangerous_artifact=true`;

  // Perform the request (insecure curl -> we skip TLS validation)
  // If you're using node-fetch or running in Node, set NODE_TLS_REJECT_UNAUTHORIZED=0
  process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Authorization': `Basic ${credentials}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch vulnerability summary: ${response.status} ${response.statusText}`);
  }

  const json = await response.json();
  return json;
}

export function runCommand(command: string): string {
  console.log(`\n$ ${command}`);

  try {
    // Run command and capture output (stdout + stderr)
    const output = execSync(command, {
      encoding: 'utf-8',  // ensures string output
      stdio: ['pipe', 'pipe', 'pipe'], // capture all streams
    });

    console.log('✅ Command output:\n', output.trim()); // print captured output
    return output.trim(); // return for further processing
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
  tag1: string = 'latest',
  localRegistry: string = LOCAL_REGISTRY,
  localNamespace: string = LOCAL_REGISTRY_NAMESPACE,
): void {
  console.log(`\n🚀 Running docker push for ${image}...`);

  const sourceImage = `${localRegistry}/${localNamespace}/${image}:${tag1}`;
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
}

async function pushImage(  ip: string,
  user: string,
  pwd: string,
  project: string,
  image: string,
  sha256?: string,
  needPullFirst: boolean = true,
  isRobot: boolean = false,
  localRegistry: string = LOCAL_REGISTRY,
  localNamespace: string = LOCAL_REGISTRY_NAMESPACE
): Promise<void> {

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
      runCommand(`docker pull ${sourceImage}`);
    }

    const username = isRobot 
      ? `robot$${project}+${user}` 
      : user;
    
    runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);

    const sourceImageForTag = needPullFirst 
      ? `${localRegistry}/${localNamespace}/${imageInUse}`
      : imageInUse;
    
    const targetImage = `${ip}/${project}/${imageInUseWithTag}`;
    
    console.log(`Tagging ${sourceImageForTag} as ${targetImage}...`);
    runCommand(`docker tag ${sourceImageForTag} ${targetImage}`);

    console.log(`Pushing ${targetImage}...`);
    runCommand(`docker push ${targetImage}`);
    console.log('Push successful');

  } finally {
    runCommand(`docker logout ${ip}`);
  }
}

async function pullImage(
    ip: string,
    user: string,
    pwd: string,
    project: string,
    image: string,
    tag?: string,
    isRobot = false
): Promise<string> {

  console.log(`\nRunning docker pull ${image}...`);

  const imageWithTag = tag ? `${image}:${tag}` : image;

  const username = isRobot 
    ? `robot$${project}+${user}` 
    : user;
  
  dockerLogin(ip, username, pwd);

  const output = runCommand(`docker pull ${ip}/${project}/${imageWithTag}`);

  return output;
}

function dockerLogin(ip: string, username: string, password: string) {
  console.log(`Logging in to ${ip}...`);
  runCommand(`docker login -u '${username}' -p '${password}' ${ip}`);
}

function dockerLogout(ip: string) {
  runCommand(`docker logout ${ip}`);
}