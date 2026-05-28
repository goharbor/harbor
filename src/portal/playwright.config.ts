import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './e2e',
    timeout: 30 * 1000,
    fullyParallel: false,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: process.env.CI
        ? [
              ['list'],
              ['github'],
              ['html', { open: 'never' }],
              ['json', { outputFile: 'test-results/results.json' }],
          ]
        : 'list',
    outputDir: 'test-results',
    expect: {
        timeout: 10 * 1000,
    },
    use: {
        baseURL: process.env.BASE_URL,
        headless: true,
        ignoreHTTPSErrors: true,
        actionTimeout: 30 * 1000,
        navigationTimeout: 30 * 1000,
        screenshot: 'only-on-failure',
        trace: process.env.CI ? 'off' : 'retain-on-failure',
        video: process.env.CI ? 'retain-on-failure' : 'off',
    },
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
});
