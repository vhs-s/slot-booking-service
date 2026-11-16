const path = require('node:path');
const { defineConfig, devices } = require('@playwright/test');

const resultsDir = path.resolve(__dirname, '../results/gui');

module.exports = defineConfig({
  testDir: './tests',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  forbidOnly: true,
  timeout: 30_000,
  expect: { timeout: 5_000 },
  outputDir: path.join(resultsDir, 'test-results'),
  reporter: [
    ['list'],
    ['html', { outputFolder: path.join(resultsDir, 'playwright-report'), open: 'never' }],
    ['json', { outputFile: path.join(resultsDir, 'results.json') }],
    ['./reporter.cjs'],
  ],
  use: {
    baseURL: process.env.GUI_BASE_URL || 'http://localhost:8080',
    locale: 'ru-RU',
    timezoneId: 'Europe/Moscow',
    viewport: { width: 1280, height: 900 },
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
    trace: 'on',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 900 } } }],
});
