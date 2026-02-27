# Playwright End-to-End Tests for Harbor Registry Portal

This document explains how to install, configure, and run the Playwright E2E tests for the Harbor Registry Portal. It also covers patterns for contributing new tests migrated from Robot Framework.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Environment Variables](#environment-variables)
- [Running Tests](#running-tests)
- [Project Structure](#project-structure)
- [Migration from Robot Framework](#migration-from-robot-framework)
- [Patterns and Best Practices](#patterns-and-best-practices)
- [External Dependencies](#external-dependencies)
- [CI/CD Integration](#cicd-integration)

---

## Quick Start

```bash
# 1. Install dependencies
cd src/portal
npm install
npx playwright install

# 2. Set environment variables
export BASE_URL=https://your-harbor-instance.com
export IP=your-harbor-instance.com

# 3. Run tests
npx playwright test
```

---

## Prerequisites

- **Node.js** and **npm**
- **Docker** (for tests requiring image push/pull)
- Access to a Harbor instance

---

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `BASE_URL` | Yes | Harbor UI URL | `https://192.168.1.100` |
| `IP` | Yes | Harbor host/IP (used for docker login) | `192.168.1.100` |
| `HARBOR_ADMIN` | No | Admin username | `admin` |
| `HARBOR_PASSWORD` | No | Admin password | `Harbor12345` |
| `LOCAL_REGISTRY` | No | Source registry for test images | `registry.goharbor.io` |
| `LOCAL_REGISTRY_NAMESPACE` | No | Namespace in source registry | `harbor-ci` |
| `WEBHOOK_ENDPOINT_UI` | No | Webhook server address (for webhook tests) | `localhost:8084` |

---

## Running Tests

### Run all tests
```bash
npx playwright test
```

### Run specific test file
```bash
npx playwright test trivy.spec.ts
npx playwright test webhook.spec.ts
```

### Run with UI mode (recommended for development)
```bash
npx playwright test --ui
```

### Run in Docker
```bash
# Build
docker build -t playwright-tests . -f ./e2e/Dockerfile

# Run
docker run \
  -e BASE_URL="https://harbor.example.com" \
  -e IP="harbor.example.com" \
  -v /dev/shm:/dev/shm \
  -it --privileged \
  playwright-tests:latest
```

---

## Project Structure

```
src/portal/e2e/
├── *.spec.ts                    # Test files
├── scripts/                     # Shell scripts for Docker operations
│   └── docker_push_manifest_list.sh
├── webhook-server/              # Webhook receiver for testing
│   ├── docker-compose.yml
│   └── README.md
├── Dockerfile                   # Docker image for CI
└── README.md                    # This file
```

---

## Migration from Robot Framework

We are migrating tests from `tests/robot-cases/` to Playwright. The detailed migration guide is in `CLAUDE.md` at the repository root.

### High-Level Process

1. **Find the Robot test case** in `tests/robot-cases/Group1-Nightly/`
2. **Trace keyword definitions** in `tests/resources/Harbor-Pages/`
3. **Understand the test flow** - what actions, what verifications
4. **Write Playwright equivalent** following the patterns below

### Where to Find Robot Keywords

| Category | Location |
|----------|----------|
| Page-specific keywords | `tests/resources/Harbor-Pages/*.robot` |
| Element locators | `tests/resources/Harbor-Pages/*_Elements.robot` |
| Utility keywords | `tests/resources/Util.robot` |
| Docker operations | `tests/resources/Docker-Util.robot` |
| Test case bodies | `tests/resources/TestCaseBody.robot` |

---

## Patterns and Best Practices

### 1. Test File Structure

```typescript
import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

// Constants at the top
const ip = process.env.IP;
const user = process.env.HARBOR_ADMIN || 'admin';
const pwd = process.env.HARBOR_PASSWORD || 'Harbor12345';

test.describe('Feature Name', () => {
  test.setTimeout(60 * 60 * 1000); // Set appropriate timeout

  test('should do something specific', async ({ page }) => {
    // Test implementation
  });
});

// Helper functions at the bottom
function helperFunction() { }
```

### 2. Login Pattern

```typescript
await page.goto('/');
await page.getByRole('textbox', { name: 'Username' }).fill(user);
await page.getByRole('textbox', { name: 'Password' }).fill(pwd);
await page.getByRole('button', { name: 'LOG IN' }).click();
```

### 3. Create Project Pattern

```typescript
await page.getByRole('button', { name: 'New Project' }).click();
await page.locator('#create_project_name').fill(projectName);
await page.getByRole('button', { name: 'OK' }).click();
```

### 4. Selector Preferences

Prefer in this order:
1. **Role-based** - `getByRole('button', { name: 'OK' })`
2. **Text-based** - `getByText('Submit')`
3. **Label-based** - `getByLabel('Username')`
4. **ID-based** - `locator('#element-id')`
5. **XPath** - `locator('//xpath')` (last resort)

### 5. Docker Operations

Use synchronous `execSync` for Docker commands:

```typescript
function runCommand(command: string): string {
  const output = execSync(command, { encoding: 'utf-8' });
  return output.trim();
}

function pushImageWithTag(ip, user, pwd, project, image, tag) {
  runCommand(`docker pull ${sourceImage}`);
  runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);
  runCommand(`docker tag ${sourceImage} ${targetImage}`);
  runCommand(`docker push ${targetImage}`);
  runCommand(`docker logout ${ip}`);
}
```

### 6. Multi-Page/Tab Pattern

For tests requiring multiple browser tabs:

```typescript
const context = await browser.newContext();
const page1 = await context.newPage();
const page2 = await context.newPage();

// Switch between pages
await page1.bringToFront();
// ... do something
await page2.bringToFront();
// ... do something else

await context.close();
```

### 7. Waiting Strategies

```typescript
// Prefer auto-waiting (built into most Playwright methods)
await page.getByRole('button', { name: 'Save' }).click();

// Explicit wait for element
await page.locator('#loading').waitFor({ state: 'hidden' });

// Wait for specific text
await expect(page.locator('body')).toContainText('Success');

// Avoid fixed timeouts, but use when necessary
await page.waitForTimeout(2000);
```

### 8. Assertions

```typescript
// Visibility
await expect(page.locator('#element')).toBeVisible();
await expect(page.locator('#element')).not.toBeVisible();

// Text content
await expect(page.locator('#message')).toContainText('Success');
await expect(page.locator('#count')).toHaveText('5');

// Element count
await expect(page.locator('.item')).toHaveCount(3);
```

### 9. Naming Conventions

- Test files: `feature-name.spec.ts`
- Test descriptions: Start with "should" - `'should create a new project'`
- Helper functions: camelCase - `pushImageWithTag()`
- Constants: UPPER_SNAKE_CASE - `LOCAL_REGISTRY`

---

## External Dependencies

### Webhook Server

For webhook tests, start the webhook receiver:

```bash
cd src/portal/e2e/webhook-server
docker compose up -d

# Then run tests with
WEBHOOK_ENDPOINT_UI=localhost:8084 npx playwright test webhook.spec.ts
```

### Docker Daemon

Tests that push/pull images require Docker daemon access. When running in Docker, use `--privileged` flag.

---

## CI/CD Integration

Tests run automatically in CI with:
- HTML reports
- Video recordings of failures
- Trace files for debugging
- Screenshots on failure

---

## Contributing New Tests

1. **Check existing patterns** - Look at `trivy.spec.ts` and `webhook.spec.ts`
2. **Document in CLAUDE.md** - Add keyword mappings before writing code
3. **Follow the patterns** - Use helper functions, proper selectors
4. **Test locally** - Run with `--ui` mode for debugging
5. **Keep tests independent** - Each test should create its own resources

---

## Troubleshooting

### Tests timing out
- Increase timeout: `test.setTimeout(60 * 60 * 1000)`
- Check Harbor instance is accessible
- Verify environment variables are set

### Docker commands failing
- Ensure Docker daemon is running
- Check network connectivity to Harbor
- Verify credentials are correct

### Selectors not finding elements
- Use Playwright Inspector: `npx playwright test --debug`
- Check if element is in iframe
- Verify page has fully loaded
