import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

// package.json has "type": "module", so this file runs as ESM under
// Playwright's TS loader — no __dirname, derive it from import.meta.url.
const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Points at the scratch model directory start-server.sh created for this
// run, so the live-reload test can mutate a throwaway copy of drone.sysml.
function workdirModelPath(): string {
  const marker = path.join(__dirname, '.e2e-workdir');
  const workDir = fs.readFileSync(marker, 'utf-8').trim();
  return path.join(workDir, 'drone.sysml');
}

test.describe('mbsecli visualizer', () => {
  test('renders the object tree from the parsed model', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Airframe', { exact: true })).toBeVisible();
    await expect(page.getByText('Battery', { exact: true })).toBeVisible();
    await expect(page.getByText('MaxWeight', { exact: true })).toBeVisible();
  });

  test('selecting an element populates the inspector', async ({ page }) => {
    await page.goto('/');
    await page.getByText('Airframe', { exact: true }).click();

    // The inspector renders FQN as a labeled, read-only field. Playwright
    // has no display-value locator (that's a Testing Library thing) — walk
    // from the "FQN" label to the next input in document order and assert
    // on its live value, which correctly reflects a React-controlled input.
    const fqnInput = page.locator('xpath=//*[contains(text(),"FQN")]/following::input[1]');
    await expect(fqnInput).toHaveValue('Drone::Airframe');
  });

  test('live-reload: editing the .sysml file on disk updates the tree', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Airframe', { exact: true })).toBeVisible();

    const modelPath = workdirModelPath();
    const original = fs.readFileSync(modelPath, 'utf-8');
    const edited = original.replace(
      'part motor : Motor;',
      'part motor : Motor;\n        part camera : Camera;',
    );
    expect(edited).not.toBe(original); // sanity: the replace actually matched

    try {
      fs.writeFileSync(modelPath, edited);
      await expect(page.getByText('camera', { exact: true })).toBeVisible({ timeout: 5000 });
    } finally {
      fs.writeFileSync(modelPath, original);
    }
  });
});
