import { defineConfig, devices } from '@playwright/test';

const customerBaseURL = process.env.CUSTOMER_BASE_URL || 'http://localhost:3000';
const adminBaseURL = process.env.ADMIN_BASE_URL || 'http://localhost:3001';

export default defineConfig({
  testDir: './apps',
  testMatch: '**/*.e2e.ts',
  timeout: 30000,
  expect: {
    timeout: 5000,
  },
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
  ],
  use: {
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'customer-chromium',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: customerBaseURL,
      },
    },
    {
      name: 'admin-chromium',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: adminBaseURL,
      },
    },
  ],
});
