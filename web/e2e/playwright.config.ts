import { defineConfig, devices } from '@playwright/test';

const PORT = 4199;

export default defineConfig({
  testDir: '.',
  timeout: 30_000,
  fullyParallel: false, // all specs share one mbsecli instance/workdir
  reporter: 'list',
  use: {
    baseURL: `http://localhost:${PORT}`,
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: `bash ./start-server.sh ${PORT}`,
    url: `http://localhost:${PORT}`,
    reuseExistingServer: !process.env.CI,
    timeout: 20_000,
  },
});
