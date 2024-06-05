const fs = require("fs");
const path = require("path");
const { exec, execSync } = require("child_process");
const os = require("os");
const { argv } = require("process");

const inputDir = argv.includes("--input")
  ? argv[argv.indexOf("--input") + 1]
  : null;
const outputDir = argv.includes("--output")
  ? argv[argv.indexOf("--output") + 1]
  : null;
const dryRun = argv.includes("--dry-run");
let codec = argv.includes("--codec") ? argv[argv.indexOf("--codec") + 1] : null;
const ipod = argv.includes("--ipod");

// if no codec but ipod, set codec to aac
if (!codec && ipod) {
  codec = "aac";
}

if (!inputDir || !outputDir || !codec) {
  console.error(
    "Usage: node script.js --input <inputDir> --output <outputDir> --codec [flac|alac|aac|wav|mp3|ogg] [--dry-run] [--ipod]"
  );
  process.exit(1);
}

const numThreads = os.cpus().length;

const isAudioFile = (file) => {
  const audioExtensions = [".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"];
  return audioExtensions.includes(path.extname(file).toLowerCase());
};

// Function to check if ffmpeg supports aac_at codec -- only on Apple
const ffmpegHasAACAT = () => {
  const result = execSync("ffmpeg -h encoder=aac_at").toString();
  return !(
    result.includes("Unknown encoder") || result.includes("is not recognized")
  );
};

const getCodecParams = (codec) => {
  switch (codec) {
    case "alac":
      return "-c:a alac -map 0:v:0 -map 1:a:0 -map_metadata:s:v:0 0:s:v:0 -map_metadata:s:a:0 1:s:a:0 -c:v copy";
    case "flac":
      return "-c:a flac -map 0:v:0 -map 1:a:0 -map_metadata:s:v:0 0:s:v:0 -map_metadata:s:a:0 1:s:a:0 -c:v copy";
    case "wav":
      return "-c:a pcm_s16le -vn";
    case "ogg":
      return `-c:a libvorbis -q:a 8 -vn`;
    case "aac":
      return `-c:a ${
        ffmpegHasAACAT() ? "aac_at" : "aac"
      } -b:a 256k -map 0:v:0 -map 1:a:0 -map_metadata:s:v:0 0:s:v:0 -map_metadata:s:a:0 1:s:a:0 -c:v copy`;
    case "mp3":
      return "-c:a libmp3lame -q:a 0";
    default:
      throw new Error(`Unsupported codec: ${codec}`);
  }
};

const codecParams = getCodecParams(codec);

const convertFile = async (inputFilePath, outputFilePath) => {
  if (dryRun) {
    console.log(`Dry run: Converting ${inputFilePath} to ${outputFilePath}`);
    return;
  }

  if (fs.existsSync(outputFilePath)) {
    console.log(`File exists, skipping: ${outputFilePath}`);
    return;
  }

  const outputDir = path.dirname(outputFilePath);
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const outputExtension = (() => {
    switch (codec) {
      case "alac":
        return ".m4a";
      case "flac":
        return ".flac";
      case "wav":
        return ".wav";
      case "ogg":
        return ".ogg";
      case "aac":
        return ".m4a";
      case "mp3":
        return ".mp3";
      default:
        return path.extname(inputFilePath);
    }
  })();

  const outputFilePathWithCodec = outputFilePath.replace(
    /\.[^/.]+$/,
    outputExtension
  );

  const command = `ffmpeg -i "${inputFilePath}" -i "${inputFilePath}" ${codecParams} "${outputFilePathWithCodec}"`;

  return new Promise((resolve, reject) => {
    const process = exec(command);

    process.stdout.on("data", (data) => console.log(data));
    process.stderr.on("data", (data) => console.error(data));
    process.on("exit", (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Conversion failed for ${inputFilePath}`));
      }
    });
  });
};

async function* walk(dir) {
  const files = await fs.promises.readdir(dir, { withFileTypes: true });
  for (const file of files) {
    const res = path.resolve(dir, file.name);
    if (file.isDirectory()) {
      yield* walk(res);
    } else {
      yield res;
    }
  }
}

const processFiles = async (inputDir, outputDir) => {
  const fileQueue = [];
  for await (const file of walk(inputDir)) {
    if (isAudioFile(file)) {
      const relativePath = path.relative(inputDir, file);
      const outputFilePath = path.join(outputDir, relativePath);
      fileQueue.push({ inputFilePath: file, outputFilePath });
    }
  }

  const workers = Array.from({ length: numThreads }, async () => {
    while (fileQueue.length > 0) {
      const { inputFilePath, outputFilePath } = fileQueue.shift();
      try {
        await convertFile(inputFilePath, outputFilePath);
      } catch (error) {
        console.error(error);
      }
    }
  });

  await Promise.all(workers);
};

const main = async () => {
  if (dryRun) {
    console.log(`Dry run enabled. No files will be converted.`);
  }
  console.log(`Using ${numThreads} threads.`);

  await processFiles(inputDir, outputDir);

  console.log("All tasks completed.");
};

main().catch((error) => console.error(error));
