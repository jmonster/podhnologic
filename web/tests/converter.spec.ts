import { expect, test, type Page } from '@playwright/test';

test('converts a wav file without 404s', async ({ page }) => {
  const monitor = monitorBrowserFailures(page);

  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Audio converter' })).toBeVisible();
  await expect(page.locator('#core-base')).toHaveValue('/ffmpeg-core');

  await page.locator('#files').setInputFiles({
    name: 'tone.wav',
    mimeType: 'audio/wav',
    buffer: makeWavFixture(),
  });
  await page.getByRole('button', { name: 'Convert' }).click();

  await expect(page.locator('#status')).toHaveText('Ready: 1 output file.', { timeout: 120_000 });
  await expect(page.getByRole('link', { name: 'Download' })).toHaveAttribute('download', 'tone.m4a');

  monitor.expectClean();
});

test('converts a wav file to every browser output format', async ({ page }) => {
  const monitor = monitorBrowserFailures(page);

  await page.goto('/');

  for (const format of ['aac', 'alac', 'flac', 'mp3', 'opus', 'wav']) {
    const extension = format === 'aac' || format === 'alac' ? 'm4a' : format;

    await page.locator('#files').setInputFiles({
      name: 'tone.wav',
      mimeType: 'audio/wav',
      buffer: makeWavFixture(),
    });
    await page.locator('#format').selectOption(format);

    await page.getByRole('button', { name: 'Convert' }).click();

    await expect(page.locator('#status')).toHaveText('Ready: 1 output file.', { timeout: 120_000 });
    await expect(page.locator('#log')).toContainText(/libavutil\s+(?:6[0-9]|[7-9][0-9])\./);
    await expect(page.getByRole('link', { name: 'Download' })).toHaveAttribute(
      'download',
      `tone.${extension}`,
    );
  }

  monitor.expectClean();
});

test('converts multiple files without showing emscripten abort noise', async ({ page }) => {
  const monitor = monitorBrowserFailures(page);

  await page.goto('/');
  await page.locator('#files').setInputFiles([
    {
      name: 'one.wav',
      mimeType: 'audio/wav',
      buffer: makeWavFixture(),
    },
    {
      name: 'two.wav',
      mimeType: 'audio/wav',
      buffer: makeWavFixture(),
    },
    {
      name: 'three.wav',
      mimeType: 'audio/wav',
      buffer: makeWavFixture(),
    },
  ]);
  await page.locator('#format').selectOption('wav');

  await page.getByRole('button', { name: 'Convert' }).click();

  await expect(page.locator('#status')).toHaveText('Ready: 3 output files.', { timeout: 120_000 });
  await expect(page.getByRole('link', { name: 'Download' })).toHaveCount(3);
  await expect(page.locator('#log')).not.toContainText('Aborted()');

  monitor.expectClean();
});

function monitorBrowserFailures(page: Page) {
  const failedResponses: string[] = [];
  const pageErrors: string[] = [];
  const consoleErrors: string[] = [];

  page.on('response', (response) => {
    if (response.status() >= 400) {
      failedResponses.push(`${response.status()} ${response.url()}`);
    }
  });
  page.on('pageerror', (error) => {
    pageErrors.push(error.message);
  });
  page.on('console', (message) => {
    if (message.type() === 'error') {
      consoleErrors.push(message.text());
    }
  });

  return {
    expectClean() {
      expect(failedResponses, 'browser responses should not include 4xx/5xx resources').toEqual([]);
      expect(pageErrors, 'page should not throw runtime errors').toEqual([]);
      expect(consoleErrors, 'console should not include errors').toEqual([]);
    },
  };
}

function makeWavFixture() {
  const sampleRate = 8000;
  const samples = 8000;
  const dataBytes = samples * 2;
  const buffer = Buffer.alloc(44 + dataBytes);
  let offset = 0;

  offset = writeAscii(buffer, offset, 'RIFF');
  buffer.writeUInt32LE(36 + dataBytes, offset);
  offset += 4;
  offset = writeAscii(buffer, offset, 'WAVE');
  offset = writeAscii(buffer, offset, 'fmt ');
  buffer.writeUInt32LE(16, offset);
  offset += 4;
  buffer.writeUInt16LE(1, offset);
  offset += 2;
  buffer.writeUInt16LE(1, offset);
  offset += 2;
  buffer.writeUInt32LE(sampleRate, offset);
  offset += 4;
  buffer.writeUInt32LE(sampleRate * 2, offset);
  offset += 4;
  buffer.writeUInt16LE(2, offset);
  offset += 2;
  buffer.writeUInt16LE(16, offset);
  offset += 2;
  offset = writeAscii(buffer, offset, 'data');
  buffer.writeUInt32LE(dataBytes, offset);
  offset += 4;

  for (let i = 0; i < samples; i += 1) {
    const phase = (2 * Math.PI * 440 * i) / sampleRate;
    buffer.writeInt16LE(Math.sin(phase) * 8000, offset);
    offset += 2;
  }

  return buffer;
}

function writeAscii(buffer: Buffer, offset: number, value: string) {
  return offset + buffer.write(value, offset, 'ascii');
}
