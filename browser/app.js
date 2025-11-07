import { FFmpeg } from 'https://unpkg.com/@ffmpeg/ffmpeg@0.12.10/dist/esm/index.js';
import { fetchFile, toBlobURL } from 'https://unpkg.com/@ffmpeg/util@0.12.1/dist/esm/index.js';

class PodhnologicBrowser {
  constructor() {
    this.ffmpeg = new FFmpeg();
    this.inputDirHandle = null;
    this.outputDirHandle = null;
    this.audioFiles = [];
    this.isProcessing = false;
    this.workers = [];
    this.maxWorkers = navigator.hardwareConcurrency || 4;

    this.initializeUI();
  }

  initializeUI() {
    // Button handlers
    document.getElementById('selectInput').addEventListener('click', () => this.selectInputDirectory());
    document.getElementById('selectOutput').addEventListener('click', () => this.selectOutputDirectory());
    document.getElementById('startConversion').addEventListener('click', () => this.startConversion());

    // iPod mode handler
    document.getElementById('ipod').addEventListener('change', (e) => {
      if (e.target.checked && document.getElementById('codec').value !== 'alac') {
        document.getElementById('codec').value = 'aac';
      }
    });

    // Parallel jobs handler
    document.getElementById('parallelJobs').addEventListener('input', (e) => {
      this.maxWorkers = parseInt(e.target.value);
    });

    this.log('Ready to convert your music collection!', 'info');
  }

  async selectInputDirectory() {
    try {
      this.inputDirHandle = await window.showDirectoryPicker({
        mode: 'read'
      });

      document.getElementById('inputInfo').textContent = `Selected: ${this.inputDirHandle.name}`;
      this.log(`Selected input directory: ${this.inputDirHandle.name}`, 'info');

      // Scan for audio files
      await this.scanAudioFiles();
      this.updateStartButton();
    } catch (error) {
      if (error.name !== 'AbortError') {
        this.log(`Error selecting input directory: ${error.message}`, 'error');
      }
    }
  }

  async selectOutputDirectory() {
    try {
      this.outputDirHandle = await window.showDirectoryPicker({
        mode: 'readwrite'
      });

      document.getElementById('outputInfo').textContent = `Selected: ${this.outputDirHandle.name}`;
      this.log(`Selected output directory: ${this.outputDirHandle.name}`, 'info');
      this.updateStartButton();
    } catch (error) {
      if (error.name !== 'AbortError') {
        this.log(`Error selecting output directory: ${error.message}`, 'error');
      }
    }
  }

  async scanAudioFiles() {
    this.audioFiles = [];
    const audioExtensions = ['.mp3', '.wav', '.flac', '.aac', '.opus', '.m4a'];

    const scanDirectory = async (dirHandle, relativePath = '') => {
      for await (const entry of dirHandle.values()) {
        if (entry.kind === 'file') {
          const ext = entry.name.substring(entry.name.lastIndexOf('.')).toLowerCase();
          if (audioExtensions.includes(ext)) {
            this.audioFiles.push({
              handle: entry,
              relativePath: relativePath ? `${relativePath}/${entry.name}` : entry.name,
              dirHandle: dirHandle
            });
          }
        } else if (entry.kind === 'directory') {
          await scanDirectory(entry, relativePath ? `${relativePath}/${entry.name}` : entry.name);
        }
      }
    };

    await scanDirectory(this.inputDirHandle);

    document.getElementById('inputFileCount').textContent = this.audioFiles.length;
    document.getElementById('inputFileInfo').classList.add('active');

    this.log(`Found ${this.audioFiles.length} audio files`, 'success');
  }

  updateStartButton() {
    const button = document.getElementById('startConversion');
    button.disabled = !(this.inputDirHandle && this.outputDirHandle && this.audioFiles.length > 0);
  }

  async initializeFFmpeg() {
    if (!this.ffmpeg.loaded) {
      this.log('Loading FFmpeg (this may take a moment)...', 'info');
      this.updateStatus('Loading FFmpeg...');

      const baseURL = 'https://unpkg.com/@ffmpeg/core@0.12.6/dist/esm';

      await this.ffmpeg.load({
        coreURL: await toBlobURL(`${baseURL}/ffmpeg-core.js`, 'text/javascript'),
        wasmURL: await toBlobURL(`${baseURL}/ffmpeg-core.wasm`, 'application/wasm'),
      });

      this.log('FFmpeg loaded successfully!', 'success');
    }
  }

  getCodecParams(codec, ipodMode, noLyrics) {
    const params = [];

    // Base parameters
    params.push('-map', '0:a');

    // Codec-specific parameters
    switch (codec) {
      case 'alac':
        params.push('-c:a', 'alac');
        if (ipodMode) {
          params.push('-sample_fmt', 's16p', '-ar', '44100');
        }
        break;
      case 'aac':
        params.push('-c:a', 'aac', '-b:a', '256k');
        if (ipodMode) {
          params.push('-ar', '44100');
        }
        break;
      case 'flac':
        params.push('-c:a', 'flac');
        break;
      case 'wav':
        params.push('-c:a', 'pcm_s16le');
        break;
      case 'opus':
        params.push('-c:a', 'libopus', '-b:a', '128k');
        break;
      case 'mp3':
        params.push('-c:a', 'libmp3lame', '-q:a', '0');
        break;
    }

    // iPod optimizations
    if (ipodMode && ['aac', 'alac'].includes(codec)) {
      params.push('-movflags', '+faststart');
    }

    return params;
  }

  getOutputExtension(codec) {
    const extensions = {
      alac: '.m4a',
      aac: '.m4a',
      flac: '.flac',
      wav: '.wav',
      opus: '.opus',
      mp3: '.mp3'
    };
    return extensions[codec] || '.m4a';
  }

  async convertFile(fileInfo, codec, ipodMode, noLyrics) {
    const inputFile = await fileInfo.handle.getFile();
    const outputExt = this.getOutputExtension(codec);
    const outputName = fileInfo.relativePath.replace(/\.[^/.]+$/, outputExt);

    try {
      // Check if output file already exists
      const outputExists = await this.checkOutputExists(outputName);
      if (outputExists) {
        this.log(`Skipping ${fileInfo.relativePath} (already exists)`, 'warning');
        return { skipped: true };
      }

      // Write input file to FFmpeg virtual filesystem
      await this.ffmpeg.writeFile(inputFile.name, await fetchFile(inputFile));

      // Get conversion parameters
      const params = this.getCodecParams(codec, ipodMode, noLyrics);

      // Run conversion
      const outputFileName = `output${outputExt}`;
      await this.ffmpeg.exec([
        '-i', inputFile.name,
        ...params,
        outputFileName
      ]);

      // Read output file
      const data = await this.ffmpeg.readFile(outputFileName);

      // Write to output directory
      await this.writeOutputFile(outputName, data);

      // Cleanup FFmpeg virtual filesystem
      await this.ffmpeg.deleteFile(inputFile.name);
      await this.ffmpeg.deleteFile(outputFileName);

      this.log(`✓ ${fileInfo.relativePath} → ${outputName}`, 'success');
      return { success: true };

    } catch (error) {
      this.log(`✗ Failed to convert ${fileInfo.relativePath}: ${error.message}`, 'error');
      return { error: error.message };
    }
  }

  async checkOutputExists(relativePath) {
    try {
      const pathParts = relativePath.split('/');
      let currentHandle = this.outputDirHandle;

      // Navigate through directories
      for (let i = 0; i < pathParts.length - 1; i++) {
        try {
          currentHandle = await currentHandle.getDirectoryHandle(pathParts[i]);
        } catch {
          return false; // Directory doesn't exist, so file doesn't exist
        }
      }

      // Check if file exists
      const fileName = pathParts[pathParts.length - 1];
      try {
        await currentHandle.getFileHandle(fileName);
        return true;
      } catch {
        return false;
      }
    } catch {
      return false;
    }
  }

  async writeOutputFile(relativePath, data) {
    const pathParts = relativePath.split('/');
    let currentHandle = this.outputDirHandle;

    // Create directories as needed
    for (let i = 0; i < pathParts.length - 1; i++) {
      currentHandle = await currentHandle.getDirectoryHandle(pathParts[i], { create: true });
    }

    // Write file
    const fileName = pathParts[pathParts.length - 1];
    const fileHandle = await currentHandle.getFileHandle(fileName, { create: true });
    const writable = await fileHandle.createWritable();
    await writable.write(data);
    await writable.close();
  }

  async startConversion() {
    if (this.isProcessing) return;

    this.isProcessing = true;
    document.getElementById('startConversion').disabled = true;
    document.getElementById('startConversion').innerHTML = 'Converting... <span class="loading"></span>';
    document.getElementById('progressContainer').style.display = 'block';

    const codec = document.getElementById('codec').value;
    const ipodMode = document.getElementById('ipod').checked;
    const noLyrics = document.getElementById('noLyrics').checked;

    try {
      // Initialize FFmpeg
      await this.initializeFFmpeg();

      this.updateStatus('Converting files...');
      this.log(`Starting conversion of ${this.audioFiles.length} files...`, 'info');
      this.log(`Codec: ${codec.toUpperCase()} | iPod Mode: ${ipodMode ? 'Yes' : 'No'} | Parallel Jobs: ${this.maxWorkers}`, 'info');

      // Process files with parallelism
      let completed = 0;
      let skipped = 0;
      let failed = 0;

      // Process files in batches
      for (let i = 0; i < this.audioFiles.length; i += this.maxWorkers) {
        const batch = this.audioFiles.slice(i, i + this.maxWorkers);
        const promises = batch.map(file => this.convertFile(file, codec, ipodMode, noLyrics));
        const results = await Promise.all(promises);

        // Update counters
        results.forEach(result => {
          if (result.success) completed++;
          else if (result.skipped) skipped++;
          else if (result.error) failed++;
        });

        // Update progress
        const progress = Math.round(((i + batch.length) / this.audioFiles.length) * 100);
        this.updateProgress(progress);
      }

      // Completion summary
      this.updateProgress(100);
      this.updateStatus('Conversion complete!');
      this.log('', 'info');
      this.log('═══════════════════════════════════════', 'info');
      this.log(`Conversion Complete!`, 'success');
      this.log(`✓ Successfully converted: ${completed} files`, 'success');
      if (skipped > 0) this.log(`⊘ Skipped (already exist): ${skipped} files`, 'warning');
      if (failed > 0) this.log(`✗ Failed: ${failed} files`, 'error');
      this.log('═══════════════════════════════════════', 'info');

    } catch (error) {
      this.log(`Fatal error: ${error.message}`, 'error');
      this.updateStatus('Conversion failed');
    } finally {
      this.isProcessing = false;
      document.getElementById('startConversion').disabled = false;
      document.getElementById('startConversion').textContent = 'Start Conversion';
    }
  }

  updateProgress(percent) {
    const fill = document.getElementById('progressFill');
    fill.style.width = `${percent}%`;
    fill.textContent = `${percent}%`;
  }

  updateStatus(text) {
    const status = document.getElementById('status');
    const statusText = document.getElementById('statusText');
    status.classList.add('active');
    statusText.textContent = text;
  }

  log(message, type = 'info') {
    const container = document.getElementById('logContainer');
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    entry.textContent = message;
    container.appendChild(entry);
    container.scrollTop = container.scrollHeight;
  }
}

// Check for File System Access API support
if (!('showDirectoryPicker' in window)) {
  document.body.innerHTML = `
    <div class="container">
      <h1>⚠️ Browser Not Supported</h1>
      <p style="text-align: center; margin-top: 20px;">
        Your browser doesn't support the File System Access API.<br><br>
        Please use a modern browser like:<br>
        <strong>Chrome, Edge, or Opera</strong><br><br>
        (Safari and Firefox don't support this feature yet)
      </p>
    </div>
  `;
} else {
  // Initialize the application
  new PodhnologicBrowser();
}
