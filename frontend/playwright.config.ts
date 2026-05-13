import { defineConfig, devices } from '@playwright/test';

const baseURL = 'http://localhost:3000';
const webServerURL = 'http://127.0.0.1:3000';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: 1,
  use: {
    baseURL,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'Desktop Chrome',
      use: {
        ...devices['Desktop Chrome'],
        ...(process.env.CI ? {} : { channel: 'chrome' as const }),
      },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 15'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: webServerURL,
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
