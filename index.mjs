#!/usr/bin/env node

import fs from 'fs'
import os from 'os'
import path from 'path'
import { argv } from 'process'
import { exec, execSync } from 'child_process'

const inputDir = argv.includes('--input') ? argv[argv.indexOf('--input') + 1] : null
const outputDir = argv.includes('--output') ? argv[argv.indexOf('--output') + 1] : null
const ffmpegPath = argv.includes('--ffmpeg') ? argv[argv.indexOf('--ffmpeg') + 1] : 'ffmpeg'
const dryRun = argv.includes('--dry-run')
const ipod = argv.includes('--ipod')
const codec = argv.includes('--codec') ? argv[argv.indexOf('--codec') + 1] : ipod ? 'aac' : null

if (!inputDir || !outputDir || !codec) {
  console.error(
    'Usage: node script.js --input <inputDir> --output <outputDir> --codec [flac|alac|aac|wav|mp3|ogg] [--dry-run] [--ipod] [--ffmpeg /opt/homebrew/bin/ffmpeg]'
  )
  process.exit(1)
}

const numThreads = os.cpus().length
const audioExtensions = ['.mp3', '.wav', '.flac', '.aac', '.ogg', '.m4a']

const isAudioFile = (file) => audioExtensions.includes(path.extname(file).toLowerCase())

// Function to check if ffmpeg has AAC_AT encoder and cache the result
const ffmpegHasAACAT = (() => {
  let cachedResult = null
  return () => {
    if (cachedResult === null) {
      try {
        const result = execSync(`${ffmpegPath} -h encoder=aac_at > /dev/null 2>&1`).toString()
        cachedResult = !result.includes('Unknown encoder') && !result.includes('is not recognized')
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
    exec(
      `${ffmpegPath.replace('ffmpeg', 'ffprobe')} -v quiet -print_format json -show_format -show_streams "${filePath}"`,
      (error, stdout) => {
        if (error) {
          reject(error)
        } else {
          resolve(JSON.parse(stdout))
        }
      }
    )
  })
}

const getCodecParams = (codec, metadata, ipod) => {
  // Normalize metadata keys to lowercase for case-insensitive mapping
  const normalizedTags = Object.keys(metadata.format.tags || {}).reduce((acc, key) => {
    acc[key.toLowerCase()] = metadata.format.tags[key]
    return acc
  }, {})

  // Desired metadata keys in lowercase
  const desiredMetadataKeys = ['title', 'artist', 'album', 'date', 'track', 'genre', 'disc']

  // Construct metadata string for desired attributes
  const desiredMetadata = desiredMetadataKeys
    .map((key) => (normalizedTags[key] ? `-metadata ${key}=${escapeShellArg(normalizedTags[key])}` : ''))
    .filter(Boolean)
    .join(' ') // Join metadata attributes with spaces

  // Base parameters for all/most codecs
  const baseParams = `-map 0 -map_metadata -1 ${desiredMetadata}`
  const videoParams = '-c:v copy'

  // Codec-specific parameters
  const codecParams = {
    alac: `-c:a alac ${videoParams} ${ipod ? '-sample_fmt s16p -ar 44100 -movflags +faststart -disposition:a 0' : ''}`, // ALAC codec params with video copy
    flac: `-c:a flac ${videoParams}`,
    wav: '-c:a pcm_s16le -vn',
    ogg: '-c:a libvorbis -q:a 8 -vn',
    aac: `-c:a ${ffmpegHasAACAT() ? 'aac_at' : 'aac'} -b:a 256k ${videoParams}`,
    mp3: '-c:a libmp3lame -q:a 0',
  }

  return `${baseParams} ${codecParams[codec]}` // Return full command parameters for the codec
}

async function convertFile(inputFilePath, outputFilePath, codecParams) {
  const codecToFileExtension = {
    // Mapping from codec to file extension
    alac: '.m4a',
    flac: '.flac',
    wav: '.wav',
    ogg: '.ogg',
    aac: '.m4a',
    mp3: '.mp3',
  }

  const outputExtension = codecToFileExtension[codec] || path.extname(inputFilePath)
  const outputFilePathWithCodec = outputFilePath.replace(/\.[^/.]+$/, outputExtension)
  const command = `${ffmpegPath} -i "${inputFilePath}" ${codecParams} "${outputFilePathWithCodec}" > /dev/null 2>&1` // Construct ffmpeg command

  if (dryRun) {
    // If dry-run mode, just log the command without executing
    console.log(`[dry run] converting ${inputFilePath} to ${outputFilePath} with the following command`)
    console.log('\x1b[92m%s\x1b[0m', `${command}\n`)
    return
  }

  if (fs.existsSync(outputFilePathWithCodec)) {
    console.log(`File exists, skipping: ${outputFilePathWithCodec}`)
    return
  }

  fs.mkdirSync(path.dirname(outputFilePathWithCodec), { recursive: true })

  console.debug('\x1b[92m%s\x1b[0m', command, '\n') // Debug log for the command

  return new Promise((resolve, reject) => {
    exec(command, (error, stdout, stderr) => {
      // Execute the command
      if (error) {
        // Log and reject if there's an error
        console.error(`Error: ${error.message}`)
        reject(new Error(`Conversion failed for ${inputFilePath}`))
      } else {
        if (stderr) console.error(`Error: ${stderr}`) // Log stderr if any
        resolve() // Resolve the promise if successful
      }
    })
  })
}

async function* walk(dir) {
  const files = await fs.promises.readdir(dir, { withFileTypes: true }) // Read directory contents
  for (const file of files) {
    const res = path.resolve(dir, file.name) // Resolve file path
    if (file.isDirectory()) yield* walk(res) // Recursively walk directories
    else yield res // Yield file path if it's a file
  }
}

async function processFiles(inputDir, outputDir) {
  const fileQueue = [] // Queue for files to process
  let activeWorkers = 0 // Count of active worker threads

  const processFile = async (file) => {
    if (isAudioFile(file)) {
      // Check if the file is an audio file
      try {
        const metadata = await extractMetadata(file) // Extract metadata from the file
        const relativePath = path.relative(inputDir, file) // Get relative path for the output file
        const outputFilePath = path.join(outputDir, relativePath) // Construct output file path
        const codecParams = getCodecParams(codec, metadata, ipod) // Get codec parameters
        await convertFile(file, outputFilePath, codecParams) // Convert the file
      } catch (error) {
        // Log any errors
        console.error(`Failed to process file: ${file}, Error: ${error.message}`)
      }
    }
  }

  const worker = async () => {
    while (fileQueue.length > 0) {
      // Process files while the queue is not empty
      const file = fileQueue.shift() // Get the next file from the queue
      if (file) await processFile(file) // Process the file
    }
    activeWorkers-- // Decrement active worker count when done
  }

  for await (const file of walk(inputDir)) {
    // Walk the input directory
    fileQueue.push(file) // Add each file to the queue
    if (activeWorkers < numThreads) {
      // Start a new worker if there are fewer than the maximum number of active workers
      activeWorkers++
      worker().catch(console.error) // Start worker and catch any errors
    }
  }

  await new Promise((resolve) => {
    const interval = setInterval(() => {
      // Check if all workers are done
      if (activeWorkers === 0 && fileQueue.length === 0) {
        clearInterval(interval) // Clear the interval when done
        resolve() // Resolve the promise
      }
    }, 100)
  })
}

async function main() {
  if (dryRun) console.log('Dry run enabled. No files will be converted.') // Log if dry run mode is enabled
  console.log(`Using ${numThreads} threads.`) // Log the number of threads being used
  await processFiles(inputDir, outputDir) // Process the files
  console.log('All tasks completed.') // Log completion
}

main().catch(console.error) // Run the main function and catch any errors
