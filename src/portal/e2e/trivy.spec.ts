
import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

// variables
const LOCAL_REGISTRY: string = process.env.LOCAL_REGISTRY || 'registry.goharbor.io';
const LOCAL_REGISTRY_NAMESPACE: string = process.env.LOCAL_REGISTRY_NAMESPACE || 'harbor-ci';
const ip: string = process.env.IP;
const user: string = process.env.HARBOR_ADMIN || 'admin';
const pwd: string = process.env.HARBOR_PASSWORD || 'Harbor12345';

test('login and scan the things', async ({ page }) => {
  test.setTimeout(60 * 60 * 1000); // 60 minutes
const tag = 'v2.2.0';
const digest = 'sha256:7bf979f25c6a6986eab83e100a7b78bd5195c9bcac03e823e64492bb17fa4dad';
const cve_id = 'CVE-2021-22926';
const packageName = 'curl';
const cvss_score_v3_from = 6.5;
const cvss_score_v3_to = 7.5;
const severity = 'High';
const d = '1';
const index_repo = `index${d}`;
const cve_description = 'Description: libcurl-using applications can ask for a specific client certificate to be used in a transfer.';

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
  for (let i = 0; i < expectedCounts.length; i++) {
    await expect(page.locator('app-vulnerability-summary')).toContainText(`${expectedCounts[i]}`);
    // await expect(page.getByText('9015')).toBeVisible();
    // await expect(page.locator('app-vulnerability-summary')).toContainText('9015');
    // await expect(page.locator('app-vulnerability-summary')).toContainText('101785 total with 85911 fixable Critical 9015 High 37632 Medium 42688 Low 11617 n/a 833 None 0 Medium: 41.94%');
    // const locator = page.locator(`(//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[${index}]`);
    // await expect(locator).toBeVisible({ timeout: 15000 });
    // await expect(locator).toHaveText(`${expectedCounts[i]}`);
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

// check the top 5 dangerous CVEs
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
  await page.locator('app-vulnerability-summary').getByRole('link', { name: summary.dangerousArtifacts[0].repository_name }).first().click(); 
  // check if the below element got the right repo and digest

  await expect(page.locator('app-vulnerability-filter form div').filter({ hasText: 'Filter by All Repository Name' }).getByRole('textbox')).toHaveValue(summary.dangerousArtifacts[0].repository_name);
  await expect(page.getByRole('textbox').nth(2)).toHaveValue(summary.dangerousArtifacts[0].digest);
  // check if the table shows the right info
  await expect(page.locator('#clr-dg-row32')).toContainText(summary.dangerousArtifacts[0].repository_name);
  await page.waitForTimeout(3000);
  await expect(page.locator('#clr-dg-row32')).toContainText(summary.dangerousArtifacts[0].digest.substring(0, 12));
  await expect(page.locator('#clr-dg-row32')).toContainText(summary.dangerous_cves[0].version);
  await expect(page.locator('#clr-dg-row32')).toContainText(summary.dangerous_cves[0].cvss_score_v3.toString());
 await page.waitForTimeout(2000);

 // check for the cve id
 
   // search for dangerous cves
  await page.getByRole('link', { name: summary.dangerous_cves[0] }).first().click();
  await page.waitForTimeout(2000);
  // check if the below element got the right cve id

  // await page.getByRole('link', { name: 'CVE-2024-' }).nth(1).click();
  // await page.getByRole('link', { name: 'CVE-2024-' }).first().click();



  // await expect(page.locator('app-vulnerability-summary')).toContainText('101785 total with 85911 fixable Critical 9015 High 37632 Medium 42688 Low 11617 n/a 833 None 0 Medium: 41.94%'); 
  // wait for data to load
  // --- Ensure exactly 5 rows visible ---
  // const rows = page.locator(top5BaseXPath);
  // await expect(rows).toHaveCount(5, { timeout: 20000 }); // retry until 5 are shown

  // --- Loop through API artifacts and validate ---
  // for (let i = 0; i < dangerousArtifacts.length && i < 5; i++) {
  //   const artifact = dangerousArtifacts[i];
  //   const repositoryName = artifact.repository_name;
  //   const shortDigest = artifact.digest.substring(0, 15);

  //   // --- Construct XPath like in Robot ---
  //   const rowXPath = `(${top5BaseXPath})[${i + 1}]`;
  //   const repoLocator = page.locator(`${rowXPath}//a[@title='${repositoryName}']`);
  //   const digestLocator = page.locator(`${rowXPath}//span[contains(text(),'${shortDigest}')]`);

  //   // --- Expect both repository name and digest visible ---
  //   await expect(repoLocator).toBeVisible({ timeout: 15000 });
  //   await expect(digestLocator).toBeVisible({ timeout: 15000 });
  // }

  // await page.getByRole('row', { name: /Select Select ${project} + '/' + ${image}/ }).locator('label').click();
    // logout
    await page.goto('/');
    await page.getByRole('button', { name: 'admin', exact: true }).waitFor();
    await page.getByRole('button', { name: 'admin', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
});



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
  tag1: string = 'latest'
): void {
  console.log(`\n🚀 Running docker push for ${image}...`);

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
}
