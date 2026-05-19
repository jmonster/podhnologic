#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck disable=SC1091
source "$SCRIPT_DIR/versions.env"

FFMPEG_GIT_REF="${FFMPEG_GIT_REF:-n${FFMPEG_VERSION}}"
REF_DIR="${FFMPEG_WASM_REF:-/tmp/podhnologic-ffmpegwasm}"
WORK_DIR="${FFMPEG_BROWSER_CORE_WORK:-$SCRIPT_DIR/work/browser-core}"
EXPORT_DIR="${FFMPEG_BROWSER_CORE_EXPORT:-$WORK_DIR/export}"
OUT_DIR="${FFMPEG_BROWSER_CORE_DIST:-$SCRIPT_DIR/dist/browser-core}"
EXTRA_CFLAGS="${FFMPEG_BROWSER_CORE_CFLAGS:--O2 -msimd128}"
EXTRA_LDFLAGS="${FFMPEG_BROWSER_CORE_LDFLAGS:-}"
RUNTIME="${FFMPEG_CONTAINER_RUNTIME:-}"
IMAGE_TAG="${FFMPEG_BROWSER_CORE_IMAGE:-podhnologic-ffmpeg-browser-core:${FFMPEG_VERSION}}"
DOCKERFILE="${FFMPEG_BROWSER_CORE_DOCKERFILE:-$SCRIPT_DIR/browser-core.Dockerfile}"

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

log() {
	printf '[ffmpeg-browser] %s\n' "$*"
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

usage() {
	cat <<EOF
Usage: ${0##*/} [--clean]

Build the browser FFmpeg core from FFmpeg ${FFMPEG_VERSION} and the official
ffmpeg.wasm wrapper source tree.

Inputs:
  FFMPEG_WASM_REF             Reference checkout of ffmpeg.wasm
                              default: $REF_DIR
  FFMPEG_CONTAINER_RUNTIME    podman or docker

Outputs:
  $OUT_DIR/ffmpeg-core.js
  $OUT_DIR/ffmpeg-core.wasm
EOF
}

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
	usage
	exit 0
fi

if [[ ${1:-} == "--clean" ]]; then
	rm -rf "$WORK_DIR" "$OUT_DIR"
	exit 0
fi

if [[ -z "$RUNTIME" ]]; then
	if command -v podman >/dev/null 2>&1; then
		RUNTIME=podman
	elif command -v docker >/dev/null 2>&1; then
		RUNTIME=docker
	else
		die "missing container runtime: set FFMPEG_CONTAINER_RUNTIME, or install podman/docker"
	fi
fi

[[ -d "$REF_DIR" ]] || die "missing ffmpeg.wasm reference checkout: $REF_DIR"
[[ -f "$DOCKERFILE" ]] || die "missing browser core Dockerfile: $DOCKERFILE"
[[ -f "$SCRIPT_DIR/browser-ffmpeg-wasm.sh" ]] || die "missing browser core link script: $SCRIPT_DIR/browser-ffmpeg-wasm.sh"

require_cmd find

mkdir -p "$WORK_DIR" "$OUT_DIR"
cp "$SCRIPT_DIR/browser-ffmpeg-wasm.sh" "$REF_DIR/build/podhnologic-ffmpeg-wasm.sh"

log "building FFmpeg core ${FFMPEG_VERSION} from $REF_DIR"
log "container runtime: $RUNTIME"

rm -rf "$EXPORT_DIR"

if [[ "$RUNTIME" == "podman" ]]; then
	"$RUNTIME" build \
		--tag "$IMAGE_TAG" \
		--build-arg EXTRA_CFLAGS="$EXTRA_CFLAGS" \
		--build-arg EXTRA_LDFLAGS="$EXTRA_LDFLAGS" \
		--build-arg FFMPEG_ST=1 \
		--build-arg FFMPEG_MT= \
		--build-arg FFMPEG_GIT_REF="$FFMPEG_GIT_REF" \
		--build-arg LAME_TARBALL_URL="$LAME_TARBALL_URL" \
		--build-arg LAME_TARBALL_SHA256="$LAME_TARBALL_SHA256" \
		--build-arg ZLIB_TARBALL_URL="$ZLIB_TARBALL_URL" \
		--build-arg ZLIB_TARBALL_SHA256="$ZLIB_TARBALL_SHA256" \
		-f "$DOCKERFILE" \
		"$REF_DIR"

	container_id="$("$RUNTIME" create "$IMAGE_TAG")"
	trap '"$RUNTIME" rm -f "$container_id" >/dev/null 2>&1 || true' EXIT
	"$RUNTIME" cp "$container_id:/dist" "$EXPORT_DIR"
else
	"$RUNTIME" buildx build \
		--build-arg EXTRA_CFLAGS="$EXTRA_CFLAGS" \
		--build-arg EXTRA_LDFLAGS="$EXTRA_LDFLAGS" \
		--build-arg FFMPEG_ST=1 \
		--build-arg FFMPEG_MT= \
		--build-arg FFMPEG_GIT_REF="$FFMPEG_GIT_REF" \
		--build-arg LAME_TARBALL_URL="$LAME_TARBALL_URL" \
		--build-arg LAME_TARBALL_SHA256="$LAME_TARBALL_SHA256" \
		--build-arg ZLIB_TARBALL_URL="$ZLIB_TARBALL_URL" \
		--build-arg ZLIB_TARBALL_SHA256="$ZLIB_TARBALL_SHA256" \
		-o "$EXPORT_DIR" \
		-f "$DOCKERFILE" \
		"$REF_DIR"
fi

core_js="$(find "$EXPORT_DIR" -type f -path '*/esm/ffmpeg-core.js' | sort | head -n 1)"
if [[ -z "$core_js" ]]; then
	core_js="$(find "$EXPORT_DIR" -type f -name 'ffmpeg-core.js' | sort | head -n 1)"
fi
[[ -n "$core_js" ]] || die "no ffmpeg-core.js found under $EXPORT_DIR"

core_dir="$(dirname "$core_js")"
core_wasm="$core_dir/ffmpeg-core.wasm"

[[ -f "$core_wasm" ]] || die "no ffmpeg-core.wasm found next to $core_js"

rm -f "$OUT_DIR/ffmpeg-core.js" "$OUT_DIR/ffmpeg-core.wasm" "$OUT_DIR/ffmpeg-core.worker.js"
cp "$core_js" "$OUT_DIR/ffmpeg-core.js"
cp "$core_wasm" "$OUT_DIR/ffmpeg-core.wasm"

if [[ -f "$core_dir/ffmpeg-core.worker.js" ]]; then
	cp "$core_dir/ffmpeg-core.worker.js" "$OUT_DIR/ffmpeg-core.worker.js"
fi

log "wrote browser core to $OUT_DIR"
