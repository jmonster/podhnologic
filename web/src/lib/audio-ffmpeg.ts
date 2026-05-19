import { FFmpeg } from '@ffmpeg/ffmpeg';
import { fetchFile, toBlobURL } from '@ffmpeg/util';

export type AudioFormat = 'aac' | 'alac' | 'flac' | 'mp3' | 'opus' | 'wav';

export interface AudioCodecOptions {
  format: AudioFormat;
  bitrateKbps: number;
  sampleRate: number;
  channels: 1 | 2;
}

export interface AudioConverterConfig {
  coreBaseUrl: string;
  onLog?: (message: string) => void;
  onProgress?: (ratio: number) => void;
}

export interface ConvertedAudioFile {
  name: string;
  mimeType: string;
  blob: Blob;
}

interface CoreResources {
  coreURL: string;
  wasmURL: string;
}

const OUTPUTS: Record<
  AudioFormat,
  { codec: string; extension: string; mimeType: string; bitrateCapable: boolean; extraArgs?: string[] }
> = {
  aac: { codec: 'aac', extension: 'm4a', mimeType: 'audio/mp4', bitrateCapable: true },
  alac: { codec: 'alac', extension: 'm4a', mimeType: 'audio/mp4', bitrateCapable: false },
  flac: { codec: 'flac', extension: 'flac', mimeType: 'audio/flac', bitrateCapable: false },
  mp3: { codec: 'libmp3lame', extension: 'mp3', mimeType: 'audio/mpeg', bitrateCapable: true },
  opus: {
    codec: 'opus',
    extension: 'opus',
    mimeType: 'audio/ogg',
    bitrateCapable: true,
    extraArgs: ['-strict', 'experimental'],
  },
  wav: { codec: 'pcm_s16le', extension: 'wav', mimeType: 'audio/wav', bitrateCapable: false },
};

export function createAudioConverter(config: AudioConverterConfig) {
  let coreResources: CoreResources | null = null;
  let supportCheck: Promise<void> | null = null;

  const load = async () => {
    supportCheck ??= assertSupportedCore();
    await supportCheck;
  };

  const convert = async (file: File, options: AudioCodecOptions) => {
    await load();

    const ffmpeg = await createLoadedFFmpeg();
    const inputName = `${safeStem(file.name)}-${crypto.randomUUID()}`;
    const outputSpec = OUTPUTS[options.format];
    const outputName = `${inputName}.${outputSpec.extension}`;

    await ffmpeg.writeFile(inputName, await fetchFile(file));

    try {
      const execArgs = [
        '-y',
        '-i',
        inputName,
        '-map',
        '0:a:0',
        '-c:a',
        outputSpec.codec,
        '-ar',
        String(options.sampleRate),
        '-ac',
        String(options.channels),
      ];

      if (outputSpec.bitrateCapable) {
        execArgs.push('-b:a', `${options.bitrateKbps}k`);
      }
      if (outputSpec.extraArgs) {
        execArgs.push(...outputSpec.extraArgs);
      }

      execArgs.push(outputName);
      const exitCode = await ffmpeg.exec(execArgs);
      if (exitCode !== 0) {
        throw new Error(`FFmpeg exited with code ${exitCode}.`);
      }

      const data = (await ffmpeg.readFile(outputName)) as Uint8Array;
      const bytes = new Uint8Array(data.length);
      bytes.set(data);
      const blob = new Blob([bytes], { type: outputSpec.mimeType });

      return {
        name: `${safeStem(file.name)}.${outputSpec.extension}`,
        mimeType: outputSpec.mimeType,
        blob,
      } satisfies ConvertedAudioFile;
    } finally {
      await Promise.allSettled([ffmpeg.deleteFile(inputName), ffmpeg.deleteFile(outputName)]);
      ffmpeg.terminate();
    }
  };

  return { load, convert };

  async function createLoadedFFmpeg(capturedLogs?: string[]) {
    const ffmpeg = new FFmpeg();

    ffmpeg.on('log', ({ message }) => {
      capturedLogs?.push(message);
      config.onLog?.(message);
    });

    ffmpeg.on('progress', ({ progress }) => {
      config.onProgress?.(progress);
    });

    coreResources ??= await resolveCoreResources(config.coreBaseUrl);
    await ffmpeg.load(coreResources);
    return ffmpeg;
  }

  async function assertSupportedCore() {
    const messages: string[] = [];
    let ffmpeg: FFmpeg | null = null;

    try {
      ffmpeg = await createLoadedFFmpeg(messages);
      const exitCode = await ffmpeg.exec(['-version']);
      if (exitCode !== 0) {
        throw new Error(`FFmpeg version check exited with code ${exitCode}.`);
      }
    } finally {
      ffmpeg?.terminate();
    }

    const versionLine = messages.find((message) => message.startsWith('ffmpeg version'));
    const versionMajor = versionLine?.match(/ffmpeg version (?:n)?(\d+)(?:\.|$)/)?.[1];
    const avutilLine = messages.find((message) => /^\s*libavutil\s+/.test(message));
    const avutilMajor = avutilLine?.match(/^\s*libavutil\s+(\d+)\./)?.[1];

    if (
      (!versionMajor || Number(versionMajor) < 8) &&
      (!avutilMajor || Number(avutilMajor) < 60)
    ) {
      const loaded = versionLine && avutilLine ? `${versionLine}; ${avutilLine}` : versionLine ?? 'unknown version';
      throw new Error(`Browser FFmpeg core must be 8.x or newer; loaded ${loaded}.`);
    }
  }
}

async function resolveCoreResources(coreBaseUrl: string): Promise<CoreResources> {
  const base = coreBaseUrl.replace(/\/+$/, '');
  const coreURL = await toBlobURL(`${base}/ffmpeg-core.js`, 'text/javascript');

  return {
    coreURL,
    wasmURL: await toBlobURL(`${base}/ffmpeg-core.wasm`, 'application/wasm'),
  };
}

function safeStem(name: string) {
  const stem = name.replace(/\.[^.]+$/, '');
  return stem.replace(/[^a-zA-Z0-9._-]+/g, '-').replace(/^-+|-+$/g, '') || 'audio';
}
