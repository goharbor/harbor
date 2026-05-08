# Playwright E2E Tests

This directory contains Playwright tests for the Harbor portal. The current suite focuses on the Security Hub/Trivy migration and a login smoke test.

## Prerequisites

-   Node.js and npm
-   Docker daemon access
-   A running Harbor instance with the test CA trusted by the browser, Docker, and Node.js

## Environment

| Variable                   | Required | Description                                                         |
| -------------------------- | -------- | ------------------------------------------------------------------- |
| `BASE_URL`                 | Yes      | Harbor UI URL, for example `https://192.168.1.100`                  |
| `IP`                       | Yes      | Harbor registry host/IP used by Docker commands                     |
| `HARBOR_ADMIN`             | No       | Admin username, defaults to `admin`                                 |
| `HARBOR_PASSWORD`          | No       | Admin password, defaults to `Harbor12345`                           |
| `LOCAL_REGISTRY`           | No       | Source registry for test images, defaults to `registry.goharbor.io` |
| `LOCAL_REGISTRY_NAMESPACE` | No       | Source namespace, defaults to `harbor-ci`                           |

For self-signed local Harbor certificates, set `NODE_EXTRA_CA_CERTS` to the Harbor CA certificate path before running Playwright.

## Run Locally

```bash
cd src/portal
npm ci
npx playwright install chromium
BASE_URL=https://<harbor-host> IP=<harbor-host> npx playwright test
```

Run one spec while iterating:

```bash
npx playwright test e2e/trivy.spec.ts
```

## Test Patterns

-   Prefer role, label, and text locators before CSS selectors.
-   Let Playwright auto-wait through actions and assertions instead of adding fixed sleeps.
-   Keep each spec independent by creating its own project/resources.
-   Use `npm ci` in CI so the lockfile controls dependency resolution.
-   Do not pass passwords in shell command strings or log command output that may contain credentials. Docker login must use `--password-stdin`.
-   Keep CI permissions minimal and upload Playwright traces, videos, screenshots, and HTML reports only as failure/debug artifacts.
