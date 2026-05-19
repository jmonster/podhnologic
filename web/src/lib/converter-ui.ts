import { createAudioConverter, type AudioFormat } from './audio-ffmpeg';

const form = document.querySelector<HTMLFormElement>('#converter-form');
const convertButton = document.querySelector<HTMLButtonElement>('#convert-button');
const fileInput = document.querySelector<HTMLInputElement>('#files');
const formatInput = document.querySelector<HTMLSelectElement>('#format');
const bitrateInput = document.querySelector<HTMLInputElement>('#bitrate');
const sampleRateInput = document.querySelector<HTMLSelectElement>('#sample-rate');
const channelsInput = document.querySelector<HTMLSelectElement>('#channels');
const coreBaseInput = document.querySelector<HTMLInputElement>('#core-base');
const statusOutput = document.querySelector<HTMLOutputElement>('#status');
const progressEl = document.querySelector<HTMLProgressElement>('#progress');
const logEl = document.querySelector<HTMLPreElement>('#log');
const resultsEl = document.querySelector<HTMLUListElement>('#results');

if (
  !form ||
  !convertButton ||
  !fileInput ||
  !formatInput ||
  !bitrateInput ||
  !sampleRateInput ||
  !channelsInput ||
  !coreBaseInput ||
  !statusOutput ||
  !progressEl ||
  !logEl ||
  !resultsEl
) {
  throw new Error('Audio converter UI did not mount.');
}

const defaultCoreBaseUrl = coreBaseInput.defaultValue || '/ffmpeg-core';

let converter: ReturnType<typeof createAudioConverter> | null = null;
let converterKey = '';
let activeUrls: string[] = [];

const setStatus = (message: string) => {
  statusOutput.textContent = message;
};

const resetResults = () => {
  for (const url of activeUrls) {
    URL.revokeObjectURL(url);
  }
  activeUrls = [];
  resultsEl.replaceChildren();
};

const setEmptyResults = (message: string) => {
  const item = document.createElement('li');
  item.className = 'status';
  item.textContent = message;
  resultsEl.replaceChildren(item);
};

const appendLog = (message: string) => {
  if (message.includes('Aborted()')) {
    return;
  }
  logEl.textContent += `${message}\n`;
  logEl.scrollTop = logEl.scrollHeight;
};

const getConverter = (coreBaseUrl: string) => {
  const key = coreBaseUrl;
  if (!converter || converterKey !== key) {
    converter = createAudioConverter({
      coreBaseUrl,
      onLog: appendLog,
      onProgress: (ratio) => {
        progressEl.value = ratio;
      },
    });
    converterKey = key;
  }

  return converter;
};

form.addEventListener('submit', async (event) => {
  event.preventDefault();

  const files = Array.from(fileInput.files ?? []);
  if (files.length === 0) {
    setStatus('Choose at least one file.');
    return;
  }

  resetResults();
  logEl.textContent = '';
  progressEl.value = 0;

  const options = {
    format: formatInput.value as AudioFormat,
    bitrateKbps: Number(bitrateInput.value || 192),
    sampleRate: Number(sampleRateInput.value),
    channels: Number(channelsInput.value) as 1 | 2,
  };
  const coreBaseUrl = String(coreBaseInput.value || defaultCoreBaseUrl);

  const target = getConverter(coreBaseUrl);

  convertButton.disabled = true;
  fileInput.disabled = true;
  formatInput.disabled = true;
  bitrateInput.disabled = true;
  sampleRateInput.disabled = true;
  channelsInput.disabled = true;
  coreBaseInput.disabled = true;

  try {
    setStatus(`Loading core from ${coreBaseUrl}...`);
    await target.load();

    const converted = [];
    for (const [index, file] of files.entries()) {
      setStatus(`Converting ${index + 1} of ${files.length}: ${file.name}`);
      converted.push(await target.convert(file, options));
    }

    setStatus(`Ready: ${converted.length} output file${converted.length === 1 ? '' : 's'}.`);

    for (const output of converted) {
      const url = URL.createObjectURL(output.blob);
      activeUrls.push(url);

      const item = document.createElement('li');
      item.className = 'result';

      const details = document.createElement('div');
      const name = document.createElement('strong');
      name.textContent = output.name;
      const mime = document.createElement('span');
      mime.textContent = output.mimeType;
      details.append(name, mime);

      const link = document.createElement('a');
      link.className = 'button secondary';
      link.href = url;
      link.download = output.name;
      link.textContent = 'Download';

      item.append(details, link);
      resultsEl.appendChild(item);
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error || 'Conversion failed.');
    setStatus(message);
    setEmptyResults('No downloadable output.');
  } finally {
    convertButton.disabled = false;
    fileInput.disabled = false;
    formatInput.disabled = false;
    bitrateInput.disabled = false;
    sampleRateInput.disabled = false;
    channelsInput.disabled = false;
    coreBaseInput.disabled = false;
  }
});
