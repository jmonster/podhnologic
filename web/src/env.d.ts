/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly PUBLIC_FFMPEG_CORE_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
