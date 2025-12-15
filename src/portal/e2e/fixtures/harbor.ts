import { test as base, Page } from '@playwright/test';

type HarborUser = {
    username: string;
    password: string;
};

type HarborFixtures = {
    harborPage: Page;
    harborUser: HarborUser;
};

export async function login(page: Page, baseURL: string | undefined, creds: HarborUser) {
    await page.goto(baseURL);
    await page.getByRole('textbox', { name: 'Username' }).fill(creds.username);
    await page.getByRole('textbox', { name: 'Password' }).fill(creds.password);
    await page.getByRole('button', { name: 'LOG IN' }).click();
}

export async function logout(page: Page, username: string) {
    await page.getByRole('button', { name: username, exact: true }).waitFor({ state: 'visible', timeout: 5000 });
    await page.getByRole('button', { name: username, exact: true }).click();
    await page.getByRole('menuitem', { name: 'Log Out' }).click();
}

export async function logoutIfPossible(page: Page, username: string) {
    try {
        await page.getByRole('button', { name: username, exact: true }).waitFor({ state: 'visible', timeout: 2000 });
    } catch {
        return;
    }

    try {
        await logout(page, username);
    } catch {
        return;
    }
}

const harborTest = base.extend<HarborFixtures>({
    harborUser: async ({}, use) => {
        await use({
            username: process.env.HARBOR_USERNAME || 'admin',
            password: process.env.HARBOR_PASSWORD || 'Harbor12345',
        });
    },

    harborPage: async ({ page, harborUser }, use) => {
        let baseURL = process.env.HARBOR_BASE_URL;
        await login(page, baseURL, harborUser);
        await use(page);
        await logoutIfPossible(page, harborUser.username);
    },
});

export const test = harborTest;
export const expect = harborTest.expect;
