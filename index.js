#!/usr/bin/env node

const fs = require('fs')
const os = require('os')
const path = require('path')
const { argv } = require('process')
const { exec, execSync } = require('child_process')
const { spawn } = require('child_process')

const inputDir = argv.includes('--input') ? argv[argv.indexOf('--input') + 1] : null
const outputDir = argv.includes('--output') ? argv[argv.indexOf('--output') + 1] : null
const ffmpegPath = argv.includes('--ffmpeg') ? argv[argv.indexOf('--ffmpeg') + 1] : 'ffmpeg'
const dryRun = argv.includes('--dry-run')
const ipod = argv.includes('--ipod')
const codec = argv.includes('--codec') ? argv[argv.indexOf('--codec') + 1] : ipod ? 'aac' : null
const noLyrics = argv.includes('--no-lyrics')

if (!inputDir || !outputDir || !codec) {
  console.error(
    'Usage:\n --input <inputDir>\n --output <outputDir>\n --codec [flac|alac|aac|opus|wav|mp3]\n [--ipod]\n [--no-lyrics]\n [--ffmpeg /opt/homebrew/bin/ffmpeg]\n [--dry-run]'
  )
  process.exit(1)
}

const numThreads = os.cpus().length
const audioExtensions = ['.mp3', '.wav', '.flac', '.aac', '.opus', '.m4a']

const isAudioFile = (file) => audioExtensions.includes(path.extname(file).toLowerCase())

const ffmpegHasEncoder = (() => {
  const cache = {}
  return (encoder) => {
    if (cache[encoder] === undefined) {
      try {
        const result = execSync(`${ffmpegPath} -h encoder=${encoder}`, { encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore'] })
        // Check if the output contains "Encoder" followed by the encoder name
        cache[encoder] = result.includes(`Encoder ${encoder}`)
      } catch {
        cache[encoder] = false
      }
    }
    return cache[encoder]
  }
})()
const escapeShellArg = (arg) => {
  if (process.platform === 'win32') {
    // For Windows, use double quotes and escape problematic characters
    return `"${arg.replace(/"/g, '""').replace(/%/g, '"%"')}"`
  } else {
    // For Unix-like systems, use single quotes
    return `'${arg.replace(/'/g, "'\\''")}'`
  }
}

const extractMetadata = (filePath) => {
  return new Promise((resolve, reject) => {
    const ffprobePath = ffmpegPath.replace(/ffmpeg$/, 'ffprobe')
    const command = `${escapeShellArg(ffprobePath)} -v quiet -print_format json -show_format -show_streams ${escapeShellArg(filePath)}`
    exec(command, (error, stdout) => {
      if (error) {
        reject(error)
      } else {
        resolve(JSON.parse(stdout))
      }
    })
  })
}

const getCodecParams = (codec, metadata, ipod) => {
  const normalizedTags = Object.keys(metadata.format.tags || {}).reduce((acc, key) => {
    acc[key.toLowerCase()] = metadata.format.tags[key]
    return acc
  }, {})

  const desiredMetadataKeys = ['title', 'artist', 'album', 'date', 'track', 'genre', 'disc']
  if (!noLyrics) {
    desiredMetadataKeys.push('lyrics')
  }

  const desiredMetadata = desiredMetadataKeys
    .map((key) => (normalizedTags[key] ? `-metadata ${key}=${normalizedTags[key]}` : ''))
    .filter(Boolean)

  const baseParams = ['-map', '0', '-map_metadata', '-1', ...desiredMetadata]
  const videoParams = ['-c:v', 'copy']

  const ipod_alacParams = ipod ? ['-sample_fmt', 's16p', '-ar', '44100', '-movflags', '+faststart', '-disposition:a', '0'] : []
  const ipod_aacParams = ipod ? ['-ar', '44100', '-movflags', '+faststart', '-disposition:a', '0'] : []

  let aacCodec = 'aac'
  if (ffmpegHasEncoder('aac_at')) {
    aacCodec = 'aac_at'
  } else if (ffmpegHasEncoder('libfdk_aac')) {
    aacCodec = 'libfdk_aac'
  }

  const codecParams = {
    alac: ['-c:a', 'alac', ...videoParams, ...ipod_alacParams],
    aac: ['-c:a', aacCodec, '-b:a', '256k', ...videoParams, ...ipod_aacParams],
    flac: ['-c:a', 'flac', ...videoParams],
    wav: ['-c:a', 'pcm_s16le', '-vn'],
    opus: ['-c:a', 'libopus', '-b:a', '128k', '-vn'],
    mp3: ['-c:a', 'libmp3lame', '-q:a', '0'],
  }

  return [...baseParams, ...codecParams[codec]]
}

const pathQuote = process.platform === 'win32' ? '"' : "'"

async function convertFile(inputFilePath, outputFilePath, codecParams) {
  const codecToFileExtension = {
    alac: '.m4a',
    flac: '.flac',
    wav: '.wav',
    opus: '.opus',
    aac: '.m4a',
    mp3: '.mp3',
  }

  const outputExtension = codecToFileExtension[codec] || path.extname(inputFilePath)
  const outputFilePathWithCodec = outputFilePath.replace(/\.[^/.]+$/, outputExtension)

  // Construct the command carefully
  const command = ['-i', inputFilePath]

  // Handle metadata and other parameters
  codecParams.forEach((param) => {
    if (param.startsWith('-metadata')) {
      command.push('-metadata')
      command.push(param.substring(10)) // Push the "key=value" part
    } else {
      command.push(param)
    }
  })

  command.push(outputFilePathWithCodec)

  if (dryRun) {
    console.log(`[dry run] converting ${inputFilePath} to ${outputFilePathWithCodec} with the following command`)
    console.log('\x1b[92m%s\x1b[0m', `${ffmpegPath} ${command.map((arg) => (arg.includes(' ') ? `"${arg}"` : arg)).join(' ')}\n`)
    return
  }

  if (fs.existsSync(outputFilePathWithCodec)) {
    console.log(`File exists, skipping: ${outputFilePathWithCodec}`)
    return
  }

  console.log(inputFilePath, '->', outputFilePathWithCodec)

  try {
    await fs.promises.mkdir(path.dirname(outputFilePathWithCodec), { recursive: true })
    await new Promise((resolve, reject) => {
      const process = spawn(ffmpegPath, command, {
        windowsHide: true,
        stdio: ['ignore', 'pipe', 'pipe'],
      })

      let stderr = ''
      process.stderr.on('data', (data) => {
        stderr += data.toString()
      })

      process.on('close', (code) => {
        if (code !== 0) {
          console.error(`FFmpeg process exited with code ${code}`)
          console.error(`FFmpeg stderr: ${stderr}`)
          reject(new Error(`Conversion failed for ${inputFilePath}`))
        } else {
          resolve()
        }
      })
    })
  } catch (error) {
    console.error(`Failed to process file: ${inputFilePath}, Error: ${error.message}`)
  }
}

async function* walk(dir) {
  const files = await fs.promises.readdir(dir, { withFileTypes: true })
  for (const file of files) {
    const res = path.resolve(dir, file.name)
    if (file.isDirectory()) yield* walk(res)
    else yield res
  }
}

async function processFiles(inputDir, outputDir) {
  const fileQueue = []
  let activeWorkers = 0

  const processFile = async (file) => {
    if (isAudioFile(file)) {
      try {
        const metadata = await extractMetadata(file)
        const relativePath = path.relative(inputDir, file)
        const outputFilePath = path.join(outputDir, relativePath)
        const codecParams = getCodecParams(codec, metadata, ipod, noLyrics)
        await convertFile(file, outputFilePath, codecParams)
      } catch (error) {
        console.error(`Failed to process file: ${file}, Error: ${error.message}`)
      }
    }
  }

  const worker = async () => {
    while (fileQueue.length > 0) {
      const file = fileQueue.shift()
      if (file) await processFile(file)
    }
    activeWorkers--
  }

  for await (const file of walk(inputDir)) {
    fileQueue.push(file)
    if (activeWorkers < numThreads) {
      activeWorkers++
      worker().catch(console.error)
    }
  }

  await new Promise((resolve) => {
    const interval = setInterval(() => {
      if (activeWorkers === 0 && fileQueue.length === 0) {
        clearInterval(interval)
        resolve()
      }
    }, 100)
  })
}

async function main() {
  if (dryRun) console.log('Dry run enabled. No files will be converted.')
  console.log(`Using ${numThreads} threads.`)
  await processFiles(inputDir, outputDir)
  console.log('All tasks completed.')
}

main().catch(console.error)
