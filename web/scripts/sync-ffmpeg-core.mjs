import { access, copyFile, mkdir, rm } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const source = join(root, '..', 'scripts', 'ffmpeg', 'dist', 'browser-core');
const target = join(root, 'public', 'ffmpeg-core');

async function copyRequired(name) {
  const sourceFile = join(source, name);

  try {
    await access(sourceFile);
  } catch {
    throw new Error(`Missing ${sourceFile}. Run: bash ../scripts/ffmpeg/build-browser-core.sh`);
  }

  await copyFile(sourceFile, join(target, name));
}

async function copyOptional(name) {
  const sourceFile = join(source, name);

  try {
    await access(sourceFile);
  } catch {
    return;
  }

  await copyFile(sourceFile, join(target, name));
}

await mkdir(target, { recursive: true });

await Promise.all([
  rm(join(target, 'ffmpeg-core.js'), { force: true }),
  rm(join(target, 'ffmpeg-core.wasm'), { force: true }),
  rm(join(target, 'ffmpeg-core.worker.js'), { force: true }),
]);

await Promise.all([
  copyRequired('ffmpeg-core.js'),
  copyRequired('ffmpeg-core.wasm'),
]);

await copyOptional('ffmpeg-core.worker.js');
