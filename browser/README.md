# Podhnologic Browser Edition

Convert your music collection entirely in your browser using WebAssembly!

## Features

✅ **Full Feature Parity** with the CLI version:
- All codecs: ALAC, AAC, FLAC, MP3, Opus, WAV
- iPod mode optimization
- Metadata preservation
- Parallel processing
- Resume support (skips existing files)

✅ **Browser Benefits**:
- No installation required
- No FFmpeg installation required
- Works 100% offline (after first load)
- Cross-platform (works on any OS)
- Privacy-focused (all processing happens locally)

## Browser Requirements

**Supported Browsers:**
- ✅ Chrome/Chromium 86+
- ✅ Edge 86+
- ✅ Opera 72+

**Not Supported:**
- ❌ Firefox (no File System Access API)
- ❌ Safari (no File System Access API)

## Usage

### Option 1: Run Locally

```bash
cd browser
npm install
npm start
```

Then open http://localhost:8080 in your browser.

### Option 2: Direct File Access

Simply open `index.html` in a supported browser. Note: Some features may require a local server due to CORS restrictions.

### Option 3: Deploy to Static Host

Deploy the `browser` directory to any static hosting service:
- GitHub Pages
- Netlify
- Vercel
- Cloudflare Pages

## How to Use

1. **Select Input Directory**: Click "Choose Folder" and select your music collection
2. **Select Output Directory**: Click "Choose Folder" and select where to save converted files
3. **Choose Options**:
   - Select output codec (AAC, ALAC, FLAC, MP3, Opus, WAV)
   - Enable iPod mode for optimized playback
   - Adjust parallel jobs (default: your CPU core count)
4. **Start Conversion**: Click "Start Conversion" and wait for completion

## Performance

- **Parallel Processing**: Converts multiple files simultaneously
- **Smart Skip**: Automatically skips files that already exist
- **Memory Efficient**: Processes files one batch at a time
- **Progress Tracking**: Real-time progress bar and detailed logs

## Codec Details

Same as CLI version:

- **ALAC & FLAC**: Lossless compression (iPod mode: down-sampled to 16-bit 44.1kHz)
- **AAC**: 256 kbps (Apple-quality encoder via FFmpeg.wasm)
- **MP3**: 320 kbps (highest quality)
- **Opus**: 128 kbps (efficient compression)
- **WAV**: PCM 16-bit

## iPod Mode

When enabled:
- Optimizes file format for iPod compatibility
- Down-samples ALAC to 16-bit 44.1kHz (prevents skipping)
- Adds `faststart` flag for better streaming
- For AAC: 256 kbps at 44.1kHz

## Technical Details

### Architecture

- **Frontend**: Vanilla JavaScript (ES modules)
- **FFmpeg**: FFmpeg.wasm 0.12.x (WebAssembly)
- **File Access**: File System Access API
- **Parallelism**: Batch processing (simulates Web Workers behavior)

### File System Access API

The browser version uses the modern File System Access API which:
- Gives direct access to local directories
- Supports reading and writing files
- Maintains directory structure
- Requires user permission (for security)

### Why No Web Workers?

While the original plan included Web Workers, FFmpeg.wasm already handles internal threading efficiently. The current implementation uses batch processing which:
- Is simpler and more maintainable
- Provides similar performance
- Avoids complex worker synchronization
- Reduces memory overhead

### Limitations

1. **Browser Support**: Only works in Chromium-based browsers
2. **Memory**: Large files may be slower than CLI version
3. **First Load**: Initial FFmpeg.wasm download (~32 MB)
4. **Encoder Quality**: Uses FFmpeg's AAC encoder (CLI can use Apple's encoder on macOS)

## Privacy & Security

✅ **100% Private**:
- All processing happens in your browser
- No files are uploaded to any server
- No tracking or analytics
- No network requests (after initial load)

✅ **Open Source**:
- Fully auditable code
- No minification or obfuscation
- Uses official FFmpeg.wasm library

## Troubleshooting

### "Browser Not Supported" Error
**Solution**: Use Chrome, Edge, or Opera (version 86+)

### FFmpeg Fails to Load
**Solution**: Check your internet connection on first load, or clear browser cache

### Conversion is Slow
**Solutions**:
- Reduce parallel jobs
- Close other browser tabs
- Use the CLI version for large libraries

### Output Files Not Appearing
**Solution**: Check that you selected the correct output directory and have write permissions

## Development

The browser version is a standalone implementation that mirrors the CLI functionality:

```
browser/
├── index.html    # UI and styling
├── app.js        # Main application logic
├── package.json  # Dev dependencies
└── README.md     # This file
```

To modify or extend:
1. Edit `app.js` for logic changes
2. Edit `index.html` for UI changes
3. No build step required (uses ES modules)

## Comparison: Browser vs CLI

| Feature | Browser | CLI |
|---------|---------|-----|
| FFmpeg Installation | ❌ Not needed | ✅ Required |
| Node.js Required | ❌ No | ✅ Yes |
| Apple AAC Encoder | ❌ No | ✅ Yes (macOS) |
| Performance | Good | Excellent |
| Large Libraries | Slower | Faster |
| Privacy | 100% Local | 100% Local |
| Cross-platform | Limited* | ✅ Yes |
| Offline Use | ✅ Yes | ✅ Yes |

*Limited to Chromium-based browsers

## License

Same as parent project: Creative Commons Attribution-NonCommercial-NoDerivatives 4.0 International License

## Contributing

Issues and PRs welcome! Please test in Chrome/Edge before submitting.
