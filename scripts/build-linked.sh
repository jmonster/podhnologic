#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
BINARY_NAME="${BINARY_NAME:-podhnologic}"
VERSION="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
GO_BIN="${GO_BIN:-$(command -v go || true)}"

if [[ -z "$GO_BIN" ]]; then
	printf 'missing required command: go\n' >&2
	exit 1
fi

host_is_linux_arch() {
	local arch="$1"
	[[ "$(uname -s)" == "Linux" ]] || return 1
	case "$arch:$(uname -m)" in
		amd64:x86_64|amd64:amd64|arm64:aarch64|arm64:arm64) return 0 ;;
		*) return 1 ;;
	esac
}

target="${1:-}"
if [[ -z "$target" ]]; then
	case "$(uname -s):$(uname -m)" in
		Darwin:arm64|Darwin:aarch64) target="darwin-arm64" ;;
		Darwin:x86_64) target="darwin-amd64" ;;
		Linux:x86_64|Linux:amd64) target="linux-amd64" ;;
		Linux:arm64|Linux:aarch64) target="linux-arm64" ;;
		MINGW*:*|MSYS*:*|CYGWIN*:*|Windows_NT:*) target="windows-amd64" ;;
		*) printf 'unsupported host: %s/%s\n' "$(uname -s)" "$(uname -m)" >&2; exit 1 ;;
	esac
fi

case "$target" in
	darwin-amd64) goos=darwin; goarch=amd64; ext=""; cc="${CC:-clang}"; target_cflags="-arch x86_64"; target_ldflags="-arch x86_64" ;;
	darwin-arm64) goos=darwin; goarch=arm64; ext=""; cc="${CC:-clang}"; target_cflags="-arch arm64"; target_ldflags="-arch arm64" ;;
	linux-amd64)
		goos=linux; goarch=amd64; ext=""; target_cflags=""; target_ldflags=""
		if host_is_linux_arch amd64; then cc="${CC:-cc}"; else cc="${CC:-x86_64-linux-gnu-gcc}"; fi
		;;
	linux-arm64)
		goos=linux; goarch=arm64; ext=""; target_cflags=""; target_ldflags=""
		if host_is_linux_arch arm64; then cc="${CC:-cc}"; else cc="${CC:-aarch64-linux-gnu-gcc}"; fi
		;;
	windows-amd64) goos=windows; goarch=amd64; ext=".exe"; cc="${CC:-x86_64-w64-mingw32-gcc}"; target_cflags=""; target_ldflags="" ;;
	windows-arm64) goos=windows; goarch=arm64; ext=".exe"; cc="${CC:-aarch64-w64-mingw32-gcc}"; target_cflags=""; target_ldflags="" ;;
	*) printf 'unsupported target: %s\n' "$target" >&2; exit 1 ;;
esac

dist_root="${FFMPEG_DIST_ROOT:-$ROOT_DIR/scripts/ffmpeg/dist}"
prefix="${FFMPEG_PREFIX:-$dist_root/$target}"
if [[ ! -f "$prefix/lib/libpodhnologicffmpeg.a" ]]; then
	if [[ -n "${FFMPEG_PREFIX:-}" ]]; then
		printf 'missing linked ffmpeg archive at %s\n' "$prefix/lib/libpodhnologicffmpeg.a" >&2
		exit 1
	fi
	"$ROOT_DIR/scripts/ffmpeg/build-native.sh" --target "$target"
fi

export PKG_CONFIG_PATH="$prefix/lib/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
export PKG_CONFIG_LIBDIR="$prefix/lib/pkgconfig${PKG_CONFIG_LIBDIR:+:$PKG_CONFIG_LIBDIR}"
export CC="$cc"
export CGO_CFLAGS="${CGO_CFLAGS:-} $target_cflags -I$prefix/include"
export CGO_LDFLAGS="${CGO_LDFLAGS:-} $target_ldflags -L$prefix/lib -lpodhnologicffmpeg $(pkg-config --static --libs libavfilter libavformat libavcodec libswresample libswscale libavutil)"
export CGO_ENABLED=1
export GOOS="$goos"
export GOARCH="$goarch"

mkdir -p "$BUILD_DIR"
"$GO_BIN" build \
	-a \
	-tags "linkedffmpeg_cgo linkedffmpeg_hidden" \
	-trimpath \
	-ldflags "-s -w -X main.Version=${VERSION}" \
	-o "$BUILD_DIR/${BINARY_NAME}-${target}${ext}" \
	"$ROOT_DIR/cmd/podhnologic"

printf '%s\n' "$BUILD_DIR/${BINARY_NAME}-${target}${ext}"
