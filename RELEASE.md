# Release Instructions

## Prerequisites

### 1. Developer ID Certificate

You need a "Developer ID Application" certificate from Apple Developer Program.

Verify it's installed:
```bash
security find-identity -v -p codesigning | grep "Developer ID Application"
```

### 2. Notarytool Credentials

Store your Apple ID credentials for notarization (one-time setup):
```bash
xcrun notarytool store-credentials "notarytool-profile" \
  --apple-id "YOUR_APPLE_ID" \
  --team-id "YOUR_TEAM_ID" \
  --password "APP_SPECIFIC_PASSWORD"
```

Create app-specific password at: https://appleid.apple.com/account/manage

### 3. Custom FFmpeg from Pixelbrite

Build FFmpeg for each architecture from the pixelbrite repo:
```bash
cd /Volumes/thunderware/GitHub/pixelbrite/ffkit

# Build for arm64 (Apple Silicon)
./build-ffmpeg.sh darwin arm64

# Build for amd64 (Intel)
./build-ffmpeg.sh darwin amd64
```

Binaries will be output to: `/Volumes/thunderware/GitHub/pixelbrite/dist/`

---

## Build & Release

### Step 1: Copy FFmpeg Binaries

```bash
# Create binaries directory
mkdir -p binaries/{darwin-arm64,darwin-amd64}

# Copy from pixelbrite dist
cp /Volumes/thunderware/GitHub/pixelbrite/dist/ffmpeg-darwin-arm64/ffmpeg binaries/darwin-arm64/
cp /Volumes/thunderware/GitHub/pixelbrite/dist/ffmpeg-darwin-arm64/ffprobe binaries/darwin-arm64/

cp /Volumes/thunderware/GitHub/pixelbrite/dist/ffmpeg-darwin-amd64/ffmpeg binaries/darwin-amd64/
cp /Volumes/thunderware/GitHub/pixelbrite/dist/ffmpeg-darwin-amd64/ffprobe binaries/darwin-amd64/

chmod +x binaries/darwin-*/*
```

### Step 2: Build Go Binaries

```bash
# Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o build/podhnologic-darwin-arm64 .

# Intel (amd64)
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o build/podhnologic-darwin-amd64 .
```

### Step 3: Sign Binaries

```bash
# Sign arm64
codesign --force --options runtime \
  --sign "Developer ID Application: YOUR_TEAM_NAME (YOUR_TEAM_ID)" \
  build/podhnologic-darwin-arm64

# Sign amd64
codesign --force --options runtime \
  --sign "Developer ID Application: YOUR_TEAM_NAME (YOUR_TEAM_ID)" \
  build/podhnologic-darwin-amd64

# Verify signatures
codesign -dv build/podhnologic-darwin-arm64
codesign -dv build/podhnologic-darwin-amd64
```

### Step 4: Notarize

```bash
# Create zips for notarization
ditto -c -k --keepParent build/podhnologic-darwin-arm64 build/podhnologic-darwin-arm64.zip
ditto -c -k --keepParent build/podhnologic-darwin-amd64 build/podhnologic-darwin-amd64.zip

# Submit for notarization (uses stored credentials)
xcrun notarytool submit build/podhnologic-darwin-arm64.zip \
  --keychain-profile "notarytool-profile" --wait

xcrun notarytool submit build/podhnologic-darwin-amd64.zip \
  --keychain-profile "notarytool-profile" --wait

# Staple notarization tickets to binaries
xcrun stapler staple build/podhnologic-darwin-arm64
xcrun stapler staple build/podhnologic-darwin-amd64
```

### Step 5: Verify

```bash
# Should show "accepted" and "source=Notarized Developer ID"
spctl --assess --verbose build/podhnologic-darwin-arm64
spctl --assess --verbose build/podhnologic-darwin-amd64
```

### Step 6: Create GitHub Release

```bash
# Tag the release
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z

# Create release with signed binaries
gh release create vX.Y.Z \
  build/podhnologic-darwin-arm64 \
  build/podhnologic-darwin-amd64 \
  --title "vX.Y.Z" \
  --notes "Signed and notarized macOS binaries"
```

---

## Linux Builds (unsigned)

Linux doesn't require code signing. Cross-compile from macOS:

```bash
# Ensure Linux FFmpeg binaries are in binaries/linux-*/
# Then build:
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o build/podhnologic-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o build/podhnologic-linux-arm64 .
```

---

## Windows Builds (unsigned)

Windows doesn't require code signing. Cross-compile from macOS:

```bash
# Ensure Windows FFmpeg binaries are in binaries/windows-amd64/
# Then build:
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o build/podhnologic-windows-amd64.exe .
```

Note: Users may see a SmartScreen warning ("Unknown publisher"). They can click "More info" → "Run anyway".

---

## Upload All Binaries to Release

```bash
gh release upload vX.Y.Z \
  build/podhnologic-linux-amd64 \
  build/podhnologic-linux-arm64 \
  build/podhnologic-windows-amd64.exe
```

---

## Troubleshooting

### Notarization fails
Check the log for details:
```bash
xcrun notarytool log SUBMISSION_ID --keychain-profile "notarytool-profile"
```

### "Developer ID Application" certificate not found
1. Go to https://developer.apple.com/account/resources/certificates/list
2. Create a new "Developer ID Application" certificate
3. Download and double-click to install in Keychain

### App-specific password issues
Regenerate at https://appleid.apple.com/account/manage under "Sign-In and Security" > "App-Specific Passwords"
