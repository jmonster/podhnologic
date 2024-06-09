#!/usr/bin/env node

import fs from 'fs'
import path from 'path'
import { exec, execSync } from 'child_process'
import os from 'os'
import { argv } from 'process'

const inputDir = argv.includes('--input') ? argv[argv.indexOf('--input') + 1] : null
const outputDir = argv.includes('--output') ? argv[argv.indexOf('--output') + 1] : null
const ffmpegPath = argv.includes('--ffmpeg') ? argv[argv.indexOf('--ffmpeg') + 1] : 'ffmpeg'
const dryRun = argv.includes('--dry-run')
const ipod = argv.includes('--ipod')
let codec = argv.includes('--codec') ? argv[argv.indexOf('--codec') + 1] : ipod ? 'aac' : null

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
  const result = execSync('ffmpeg -h encoder=aac_at').toString()
  return !result.includes('Unknown encoder') && !result.includes('is not recognized')
}

const getCodecParams = (codec, metadata) => {
  let baseParams = ''

  if (ipod) {
    let essentialMetadata = [
      'title',
      'artist',
      metadata && metadata.album_artist ? 'album_artist' : 'album',
      'track',
      'disc',
      'composer',
      'genre',
      'year',
    ].filter((key) => key !== 'album_artist') // Omit 'album_artist'

    baseParams = `-map_metadata -1` // Start with stripping all metadata

    essentialMetadata.forEach((key) => {
      if (metadata && metadata[key]) {
        baseParams += ` -metadata ${key}="${metadata[key]}"`
      }
    })
  }

  // Codec specific parameters
  switch (codec) {
    case 'alac':
      return `${baseParams} -c:a alac -c:v copy`
    case 'flac':
      return `${baseParams} -c:a flac -c:v copy`
    case 'wav':
      return `${baseParams} -c:a pcm_s16le -vn`
    case 'ogg':
      return `${baseParams} -c:a libvorbis -q:a 8 -vn`
    case 'aac':
      return `${baseParams} -c:a ${ffmpegHasAACAT() ? 'aac_at' : 'aac'} -b:a 256k -c:v copy`
    case 'mp3':
      return `${baseParams} -c:a libmp3lame -q:a 0`
    default:
      throw new Error(`Unsupported codec: ${codec}`)
  }
}

function extractMetadata(filePath) {
  return new Promise((resolve, reject) => {
    exec(`ffprobe -v quiet -print_format json -show_format -show_streams "${filePath}"`, (error, stdout) => {
      if (error) {
        reject(error)
      } else {
        resolve(JSON.parse(stdout))
      }
    })
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
  // const outputFilePathWithCodec = outputFilePath.replace(/\.[^/.]+$/, path.extname(inputFilePath))

  const command = `${ffmpegPath} -i "${inputFilePath}" ${codecParams} "${outputFilePathWithCodec}"`

  return new Promise((resolve, reject) => {
    const process = exec(command, (error, stdout, stderr) => {
      if (error) {
        console.error(`Error: ${error.message}`)
        reject(new Error(`Conversion failed for ${inputFilePath}`))
      }
      if (stderr) console.error(`Error: ${stderr}`)
      resolve()
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
  for await (const file of walk(inputDir)) {
    if (isAudioFile(file)) {
      const metadata = extractMetadata(file) // Assume extractMetadata is a function you will define
      const relativePath = path.relative(inputDir, file)
      const outputFilePath = path.join(outputDir, relativePath)
      const codecParams = getCodecParams(codec, metadata)
      fileQueue.push({ inputFilePath: file, outputFilePath, codecParams })
    }
  }

  await Promise.all(
    Array.from({ length: numThreads }, async () => {
      while (fileQueue.length > 0) {
        const { inputFilePath, outputFilePath, codecParams } = fileQueue.shift()
        await convertFile(inputFilePath, outputFilePath, codecParams)
      }
    })
  )
}

async function main() {
  if (dryRun) console.log('Dry run enabled. No files will be converted.')
  console.log(`Using ${numThreads} threads.`)
  await processFiles(inputDir, outputDir)
  console.log('All tasks completed.')
}

main().catch(console.error)
