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

const ffmpegHasAACAT = () => {
  try {
    const result = execSync(`${ffmpegPath} -h encoder=aac_at > /dev/null 2>&1`).toString()
    return !result.includes('Unknown encoder') && !result.includes('is not recognized')
  } catch {
    return false
  }
}

const escapeShellArg = (arg) => {
  if (process.platform === 'win32') {
    return `"${arg.replace(/(["%])/g, '^$1')}"`
  }
  return `"${arg.replace(/(["$`\\])/g, '\\$1')}"`
}

const getCodecParams = (codec, metadata) => {
  const desiredMetadata = ['title', 'artist', 'album', 'date', 'track', 'genre', 'disc']
    .map((attr) => `-metadata ${attr}=${escapeShellArg(metadata.format.tags[attr] || '')}`)
    .join(' ')

  switch (codec) {
    case 'alac':
      return `-map_metadata -1 ${desiredMetadata} -c:a alac -c:v copy ${ipod ? '-sample_fmt s16p -ar 44100' : ''}`
    case 'flac':
      return `-map_metadata -1 ${desiredMetadata} -c:a flac -c:v copy`
    case 'wav':
      return `-map_metadata -1 ${desiredMetadata} -c:a pcm_s16le -vn`
    case 'ogg':
      return `-map_metadata -1 ${desiredMetadata} -c:a libvorbis -q:a 8 -vn`
    case 'aac':
      return `-map_metadata -1 ${desiredMetadata} -c:a ${ffmpegHasAACAT() ? 'aac_at' : 'aac'} -b:a 256k -c:v copy`
    case 'mp3':
      return `-map_metadata -1 ${desiredMetadata} -c:a libmp3lame -q:a 0`
    default:
      throw new Error(`Unsupported codec: ${codec}`)
  }
}

function extractMetadata(filePath) {
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

async function convertFile(inputFilePath, outputFilePath, codecParams) {
  if (dryRun) {
    console.log(`Dry run: Converting ${inputFilePath} to ${outputFilePath} with params ${codecParams}`)
    return
  }

  if (fs.existsSync(outputFilePath)) {
    console.log(`File exists, skipping: ${outputFilePath}`)
    return
  }

  fs.mkdirSync(path.dirname(outputFilePath), { recursive: true })

  const codecToFileExtension = {
    alac: '.m4a',
    flac: '.flac',
    wav: '.wav',
    ogg: '.ogg',
    aac: '.m4a',
    mp3: '.mp3',
  }

  const outputExtension = codecToFileExtension[codec] || path.extname(inputFilePath)
  const outputFilePathWithCodec = outputFilePath.replace(/\.[^/.]+$/, outputExtension)

  const command = `${ffmpegPath} -i "${inputFilePath}" ${codecParams} "${outputFilePathWithCodec}" > /dev/null 2>&1`
  console.debug(command + '\n')

  return new Promise((resolve, reject) => {
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
        const codecParams = getCodecParams(codec, metadata)
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
