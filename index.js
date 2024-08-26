#!/usr/bin/env node

const fs = require('fs')
const os = require('os')
const path = require('path')
const { argv } = require('process')
const { exec, execSync } = require('child_process')

const inputDir = argv.includes('--input') ? argv[argv.indexOf('--input') + 1] : null
const outputDir = argv.includes('--output') ? argv[argv.indexOf('--output') + 1] : null
const ffmpegPath = argv.includes('--ffmpeg') ? argv[argv.indexOf('--ffmpeg') + 1] : 'ffmpeg'
const dryRun = argv.includes('--dry-run')
const ipod = argv.includes('--ipod')
const codec = argv.includes('--codec') ? argv[argv.indexOf('--codec') + 1] : ipod ? 'aac' : null
const noLyrics = argv.includes('--no-lyrics')
const trimMetadata = argv.includes('--trim-only')

if (!inputDir || !outputDir) {
  console.error(
    'Usage:\n --input <inputDir>\n --output <outputDir>\n [--codec [flac|alac|aac|opus|wav|mp3]]\n [--ipod]\n [--no-lyrics]\n [--ffmpeg /opt/homebrew/bin/ffmpeg]\n [--trim-only]\n [--dry-run]'
  )
  process.exit(1)
}

const numThreads = os.cpus().length
const audioExtensions = ['.mp3', '.wav', '.flac', '.aac', '.opus', '.m4a']

const isAudioFile = (file) => audioExtensions.includes(path.extname(file).toLowerCase())

const ffmpegHasAACAT = (() => {
  let cachedResult = null
  return () => {
    if (cachedResult === null) {
      try {
        execSync(`${ffmpegPath} -h encoder=aac_at`, { stdio: 'ignore' })
        cachedResult = true
      } catch {
        cachedResult = false
      }
    }
    return cachedResult
  }
})()

const escapeShellArg = (arg) => {
  if (process.platform === 'win32') {
    return `"${arg.replace(/(["%])/g, '^$1')}"`
  }
  return `"${arg.replace(/(["$`\\])/g, '\\$1')}"`
}

const extractMetadata = (filePath) => {
  return new Promise((resolve, reject) => {
    const ffprobePath = ffmpegPath.replace(/ffmpeg$/, 'ffprobe')
    exec(`"${ffprobePath}" -v quiet -print_format json -show_format -show_streams "${filePath}"`, (error, stdout) => {
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
    .map((key) => (normalizedTags[key] ? `-metadata ${key}=${escapeShellArg(normalizedTags[key])}` : ''))
    .filter(Boolean)
    .join(' ')

  // trim only?
  if (trimMetadata) {
    // Copy the original file without re-encoding, but only keep desired metadata
    return `-c copy -map 0 -map_metadata -1 ${desiredMetadata}`
  }

  const baseParams = `-map 0 -map_metadata -1 ${desiredMetadata}`
  const videoParams = '-c:v copy'

  const ipod_alacParams = ipod ? '-sample_fmt s16p -ar 44100 -movflags +faststart -disposition:a 0' : ''
  const ipod_aacParams = ipod ? '-ar 44100 -movflags +faststart -disposition:a 0' : ''
  const codecParams = {
    alac: `-c:a alac ${videoParams} ${ipod_alacParams}`,
    aac: `-c:a ${ffmpegHasAACAT() ? 'aac_at' : 'aac'} -b:a 256k ${videoParams} ${ipod_aacParams}`,
    flac: `-c:a flac ${videoParams}`,
    wav: '-c:a pcm_s16le -vn',
    opus: '-c:a libopus -b:a 128k -vn',
    mp3: '-c:a libmp3lame -q:a 0',
  }

  return `${baseParams} ${codecParams[codec]}`
}

async function convertFile(inputFilePath, outputFilePath, codecParams) {
  const codecToFileExtension = {
    alac: '.m4a',
    flac: '.flac',
    wav: '.wav',
    opus: '.opus',
    aac: '.m4a',
    mp3: '.mp3',
  }

  const outputExtension = trimMetadata ? path.extname(inputFilePath) : codecToFileExtension[codec] || path.extname(inputFilePath)
  const outputFilePathWithCodec = outputFilePath.replace(/\.[^/.]+$/, outputExtension)
  const command = `${ffmpegPath} -i "${inputFilePath}" ${codecParams} "${outputFilePathWithCodec}" > /dev/null 2>&1`

  if (dryRun) {
    console.log(`[dry run] converting ${inputFilePath} to ${outputFilePath} with the following command`)
    console.log('\x1b[92m%s\x1b[0m', `${command}\n`)
    return
  }

  if (fs.existsSync(outputFilePathWithCodec)) {
    console.log(`File exists, skipping: ${outputFilePathWithCodec}`)
    return
  }

  console.debug('\x1b[92m%s\x1b[0m', command, '\n')

  try {
    await fs.promises.mkdir(path.dirname(outputFilePathWithCodec), { recursive: true })
    await new Promise((resolve, reject) => {
      exec(command, (error, stdout, stderr) => {
        if (error) {
          console.error(`Error: ${error.message}`)
          reject(new Error(`Conversion failed for ${inputFilePath}`))
        } else {
          if (stderr) console.error(`Error: ${stderr}`)
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
        const codecParams = getCodecParams(codec, metadata, ipod)
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
