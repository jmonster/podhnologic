#!/bin/bash
# Download static ffmpeg binaries for embedding
# This script downloads pre-built static ffmpeg binaries for all supported platforms

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARIES_DIR="$SCRIPT_DIR/../binaries"

# Create binaries directory structure
mkdir -p "$BINARIES_DIR"/{darwin-amd64,darwin-arm64,linux-amd64,linux-arm64,windows-amd64}

echo "Downloading static ffmpeg binaries..."
echo "This may take a while as the files are large."
echo ""

# macOS (both Intel and ARM use universal binaries from evermeet.cx)
echo "Downloading macOS binaries..."
curl -L -o /tmp/ffmpeg-mac.zip "https://evermeet.cx/ffmpeg/getrelease/ffmpeg/zip"
curl -L -o /tmp/ffprobe-mac.zip "https://evermeet.cx/ffmpeg/getrelease/ffprobe/zip"

unzip -q /tmp/ffmpeg-mac.zip -d "$BINARIES_DIR/darwin-amd64/"
unzip -q /tmp/ffprobe-mac.zip -d "$BINARIES_DIR/darwin-amd64/"

unzip -q /tmp/ffmpeg-mac.zip -d "$BINARIES_DIR/darwin-arm64/"
unzip -q /tmp/ffprobe-mac.zip -d "$BINARIES_DIR/darwin-arm64/"

rm /tmp/ffmpeg-mac.zip /tmp/ffprobe-mac.zip
chmod +x "$BINARIES_DIR/darwin-amd64"/* "$BINARIES_DIR/darwin-arm64"/*
echo "✓ macOS binaries complete"
echo ""

# Linux amd64
echo "Downloading Linux amd64..."
curl -L -o /tmp/ffmpeg-linux-amd64.tar.xz \
    "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz"

tar -xJf /tmp/ffmpeg-linux-amd64.tar.xz -C /tmp
find /tmp/ffmpeg-master-latest-linux64-gpl -name "ffmpeg" -o -name "ffprobe" | while read binary; do
    cp "$binary" "$BINARIES_DIR/linux-amd64/"
    chmod +x "$BINARIES_DIR/linux-amd64/$(basename "$binary")"
done

rm -rf /tmp/ffmpeg-linux-amd64.tar.xz /tmp/ffmpeg-master-latest-linux64-gpl
echo "✓ Linux amd64 complete"
echo ""

# Linux arm64
echo "Downloading Linux arm64..."
curl -L -o /tmp/ffmpeg-linux-arm64.tar.xz \
    "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz"

tar -xJf /tmp/ffmpeg-linux-arm64.tar.xz -C /tmp
find /tmp/ffmpeg-master-latest-linuxarm64-gpl -name "ffmpeg" -o -name "ffprobe" | while read binary; do
    cp "$binary" "$BINARIES_DIR/linux-arm64/"
    chmod +x "$BINARIES_DIR/linux-arm64/$(basename "$binary")"
done

rm -rf /tmp/ffmpeg-linux-arm64.tar.xz /tmp/ffmpeg-master-latest-linuxarm64-gpl
echo "✓ Linux arm64 complete"
echo ""

# Windows amd64
echo "Downloading Windows amd64..."
curl -L -o /tmp/ffmpeg-windows.zip \
    "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"

unzip -q /tmp/ffmpeg-windows.zip -d /tmp
find /tmp/ffmpeg-master-latest-win64-gpl -name "ffmpeg.exe" -o -name "ffprobe.exe" | while read binary; do
    cp "$binary" "$BINARIES_DIR/windows-amd64/"
done

rm -rf /tmp/ffmpeg-windows.zip /tmp/ffmpeg-master-latest-win64-gpl
echo "✓ Windows amd64 complete"
echo ""

echo "All binaries downloaded successfully!"
echo ""
echo "Directory structure:"
find "$BINARIES_DIR" -type f -exec ls -lh {} \;
echo ""
echo "Total size:"
du -sh "$BINARIES_DIR"
