# Playwright End-to-End Tests for Harbor Registry Portal

This document explains how to install, configure, and run the Playwright E2E tests for the Harbor Registry Portal.
It covers running the tests directly on your machine as well as running them inside the Docker image.


## 1. Prerequisites

You only need **Node.js** and **npm** installed on your system.
No other system-level packages are required unless you want to run the tests through Docker.


## 2. Local Installation

1. Navigate to the portal source directory:

   ```
   cd src/portal
   ```

2. Install dependencies:

   ```
   npm install
   ```

   This installs Playwright and all test dependencies.

3. Install Playwright browsers (first-time only):

   ```
   npx playwright install
   ```

---

## 3. Required Environment Variables

The tests require **two** environment variables to be set before running:

```
export BASE_URL=https://registry.harbor.com
export IP=registry.harbor.com
```

Explanation:

* **BASE_URL** → URL of the Harbor instance under test
* **IP** → Host/IP of the registry, used by the tests for validations and networking logic

> Make sure these variables are exported in the same terminal session where you run the tests.

---

## 4. Running the Tests Locally

### Run the full test suite:

```
npx playwright test
```

### Run tests with the interactive UI (recommended for development):

```
npx playwright test --ui
```

### After tests run:

* An **HTML report** is generated.
* Videos of failures are automatically saved and embedded inside the report.
* Live Tracing/debugging is available when using UI mode.
* Traces are recorded on test failures.

This is extremely useful for debugging and for test development.

---

## 5. Running Tests inside Docker

The repository includes a Dockerfile located at:

```
src/portal/e2e/Dockerfile
```

### Build the Docker image:

From inside `src/portal`:

```
docker build -t playwright-tests . -f ./e2e/Dockerfile
```

### Run the tests using Docker:

```
docker run \
  -e BASE_URL="http://localhost:43403" \
  -e IP="localhost:43403" \
  -e NETWORK_TYPE=public \
  -v /dev/shm:/dev/shm \
  -it --privileged \
  localhost/playwright-tests:latest
```

Notes:

* `/dev/shm` is mounted to avoid browser memory issues.
* `--privileged` is required because the image includes Docker Engine (Docker-in-Docker).
* `NETWORK_TYPE` is an optional parameter used by certain tests.

---

## 6. CI/CD Integration

These tests run in CI/CD as part of the automated test pipeline.
The CI pipeline also embeds:

* Videos
* Traces
* Screenshots
* HTML reports

This ensures full visibility whenever a test fails.

---

## 7. Summary

* You can run the tests **locally** with just `npm install` and `npx playwright test`.
* You can run the tests inside **Docker** using the provided Dockerfile.
* Two required environment variables must be set:

  * `BASE_URL`
  * `IP`
* UI mode is available for test development:

  ```
  npx playwright test --ui
  ```
* HTML reports and videos of failures are automatically generated.

This setup provides a clean and reliable Playwright testing workflow for Harbor Registry Portal.
