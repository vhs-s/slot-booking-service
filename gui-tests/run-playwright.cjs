const path = require('node:path');
const { spawnSync } = require('node:child_process');

// Keep browser binaries in this project, including when npm scripts run on Windows.
const result = spawnSync(
  process.execPath,
  [require.resolve('@playwright/test/cli'), ...process.argv.slice(2)],
  {
    cwd: __dirname,
    env: { ...process.env, PLAYWRIGHT_BROWSERS_PATH: path.join(__dirname, '.browsers') },
    stdio: 'inherit',
    windowsHide: true,
  },
);
if (result.error) console.error(result.error.message);
process.exitCode = result.status ?? 1;
