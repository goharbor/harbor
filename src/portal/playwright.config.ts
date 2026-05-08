import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    globalTimeout: 60 * 60 * 1000,
    testDir: './e2e',
    fullyParallel: false,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: process.env.CI
        ? [['github'], ['html', { open: 'never' }]]
        : [['list'], ['html', { open: 'never' }]],
    outputDir: 'test-results',
    expect: {
        timeout: 10 * 1000,
    },
    use: {
        baseURL: process.env.BASE_URL,
        headless: true,
        ignoreHTTPSErrors: true,
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
