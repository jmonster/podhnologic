#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# shellcheck disable=SC1091
source "$SCRIPT_DIR/versions.env"

ALL_TARGETS=(darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64)
TARGETS=()
CLEAN=0
JOBS="${FFMPEG_JOBS:-}"
CACHE_ROOT="${FFMPEG_CACHE_ROOT:-$SCRIPT_DIR/cache}"
WORK_ROOT="${FFMPEG_WORK_ROOT:-$SCRIPT_DIR/work}"
DIST_ROOT="${FFMPEG_DIST_ROOT:-$SCRIPT_DIR/dist}"

usage() {
	cat <<EOF
Usage: ${0##*/} [--target TARGET ...] [--all] [--clean] [--jobs N]

Targets:
  darwin-amd64
  darwin-arm64
  linux-amd64
  linux-arm64
  windows-amd64
  windows-arm64

Outputs land under: $DIST_ROOT/<target>
EOF
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

log() {
	printf '[ffmpeg:%s] %s\n' "${CURRENT_TARGET:-build}" "$*"
}

default_jobs() {
	if command -v nproc >/dev/null 2>&1; then
		nproc
		return
	fi
	case "$(uname -s)" in
		Darwin) sysctl -n hw.ncpu ;;
		*) getconf _NPROCESSORS_ONLN ;;
	esac
}

default_target() {
	local os arch
	os="$(uname -s)"
	arch="$(uname -m)"

	case "$os:$arch" in
		Darwin:arm64|Darwin:aarch64) echo darwin-arm64 ;;
		Darwin:x86_64) echo darwin-amd64 ;;
		Linux:x86_64|Linux:amd64) echo linux-amd64 ;;
		Linux:arm64|Linux:aarch64) echo linux-arm64 ;;
		MINGW*:*|MSYS*:*|CYGWIN*:*|Windows_NT:*) echo windows-amd64 ;;
		*) die "unsupported host: ${os}/${arch}" ;;
	esac
}

target_os() {
	case "$1" in
		darwin-*) echo darwin ;;
		linux-*) echo linux ;;
		windows-*) echo mingw32 ;;
		*) die "unsupported target: $1" ;;
	esac
}

target_arch() {
	case "$1" in
		*-amd64) echo x86_64 ;;
		*-arm64) echo aarch64 ;;
		*) die "unsupported target: $1" ;;
	esac
}

target_cross_prefix() {
	case "$1" in
		darwin-*) echo "" ;;
		linux-amd64) echo x86_64-linux-gnu- ;;
		linux-arm64) echo aarch64-linux-gnu- ;;
		windows-amd64) echo x86_64-w64-mingw32- ;;
		windows-arm64) echo aarch64-w64-mingw32- ;;
		*) die "unsupported target: $1" ;;
	esac
}

target_autotools_host() {
	case "$1" in
		darwin-amd64) echo x86_64-apple-darwin ;;
		darwin-arm64) echo aarch64-apple-darwin ;;
		linux-amd64) echo x86_64-linux-gnu ;;
		linux-arm64) echo aarch64-linux-gnu ;;
		windows-amd64) echo x86_64-w64-mingw32 ;;
		windows-arm64) echo aarch64-w64-mingw32 ;;
		*) die "unsupported target: $1" ;;
	esac
}

is_native_target() {
	[[ "$1" == "$(default_target)" ]]
}

join_csv() {
	local IFS=,
	printf '%s' "$*"
}

hash_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

verify_sha256() {
	local file="$1"
	local expected="$2"
	local actual

	actual="$(hash_sha256 "$file")"
	[[ "$actual" == "$expected" ]]
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

download_file() {
	local url="$1"
	local dest="$2"
	local sha256="$3"
	local tmp="${dest}.tmp"

	if [[ -f "$dest" ]] && verify_sha256 "$dest" "$sha256"; then
		return
	fi

	rm -f "$tmp"
	curl -fsSL --retry 3 --retry-delay 1 "$url" -o "$tmp"
	verify_sha256 "$tmp" "$sha256" || {
		rm -f "$tmp"
		die "sha256 mismatch for $url"
	}
	mv "$tmp" "$dest"
}

ensure_source_tree() {
	local name="$1"
	local version="$2"
	local url="$3"
	local sha256="$4"
	local source_root="$5"

	local archive_name archive_path source_dir
	archive_name="${url##*/}"
	archive_path="$CACHE_ROOT/$archive_name"
	source_dir="$source_root/$name-$version"

	if [[ -d "$source_dir" ]]; then
		return
	fi

	mkdir -p "$CACHE_ROOT" "$source_root"
	log "fetching $name $version"
	download_file "$url" "$archive_path" "$sha256"
	log "extracting $name $version"
	tar -xf "$archive_path" -C "$source_root"
	[[ -d "$source_dir" ]] || die "failed to extract $name $version"
}

target_toolchain() {
	local target="$1"
	local cross_prefix

	if is_native_target "$target"; then
		case "$(uname -s)" in
			Darwin)
				CC="${FFMPEG_CC:-clang}"
				AR="${FFMPEG_AR:-ar}"
				RANLIB="${FFMPEG_RANLIB:-ranlib}"
				STRIP="${FFMPEG_STRIP:-strip}"
				PKG_CONFIG="${FFMPEG_PKG_CONFIG:-pkg-config}"
				;;
			*)
				CC="${FFMPEG_CC:-cc}"
				AR="${FFMPEG_AR:-ar}"
				RANLIB="${FFMPEG_RANLIB:-ranlib}"
				STRIP="${FFMPEG_STRIP:-strip}"
				PKG_CONFIG="${FFMPEG_PKG_CONFIG:-pkg-config}"
				;;
		esac
		CROSS_PREFIX=""
	else
		cross_prefix="${FFMPEG_CROSS_PREFIX:-$(target_cross_prefix "$target")}"
		if [[ -n "$cross_prefix" ]]; then
			CC="${FFMPEG_CC:-${cross_prefix}gcc}"
			AR="${FFMPEG_AR:-${cross_prefix}ar}"
			RANLIB="${FFMPEG_RANLIB:-${cross_prefix}ranlib}"
			STRIP="${FFMPEG_STRIP:-${cross_prefix}strip}"
			if [[ "$target" == windows-* ]]; then
				OBJCOPY="${FFMPEG_OBJCOPY:-${cross_prefix}objcopy}"
			fi
		else
			CC="${FFMPEG_CC:-clang}"
			AR="${FFMPEG_AR:-ar}"
			RANLIB="${FFMPEG_RANLIB:-ranlib}"
			STRIP="${FFMPEG_STRIP:-strip}"
			if [[ "$target" == windows-* ]]; then
				OBJCOPY="${FFMPEG_OBJCOPY:-objcopy}"
			fi
		fi
		PKG_CONFIG="${FFMPEG_PKG_CONFIG:-pkg-config}"
		CROSS_PREFIX="$cross_prefix"
	fi

	[[ -n "${CC:-}" ]] || die "missing compiler for $target"
	[[ -n "${AR:-}" ]] || die "missing archiver for $target"
	[[ -n "${RANLIB:-}" ]] || die "missing ranlib for $target"
	[[ -n "${STRIP:-}" ]] || die "missing strip for $target"
	if [[ "$target" == windows-* ]]; then
		[[ -n "${OBJCOPY:-}" ]] || die "missing objcopy for $target"
	fi
	[[ -n "${PKG_CONFIG:-}" ]] || die "missing pkg-config for $target"

	require_cmd "$CC"
	require_cmd "$AR"
	require_cmd "$RANLIB"
	require_cmd "$STRIP"
	if [[ "$target" == windows-* ]]; then
		require_cmd "$OBJCOPY"
	fi
	require_cmd "$PKG_CONFIG"
}

target_cflags() {
	case "$1" in
		darwin-amd64) echo "-O2 -fPIC -arch x86_64" ;;
		darwin-arm64) echo "-O2 -fPIC -arch arm64" ;;
		windows-*) echo "-O2" ;;
		*) echo "-O2 -fPIC" ;;
	esac
}

target_ldflags() {
	case "$1" in
		darwin-amd64) echo "-arch x86_64" ;;
		darwin-arm64) echo "-arch arm64" ;;
		*) echo "" ;;
	esac
}

build_lame() {
	local target="$1"
	local source_dir="$2"
	local prefix="$3"
	local host="$4"
	local cflags="$5"
	local cppflags="$6"
	local ldflags="$7"

	log "configuring lame"
	pushd "$source_dir" >/dev/null
	make distclean >/dev/null 2>&1 || true

	local configure_args=(
		--prefix="$prefix"
		--disable-shared
		--enable-static
		--with-pic
		--disable-frontend
		--disable-decoder
	)
	if ! is_native_target "$target"; then
		configure_args+=(--host="$host")
	fi

	env \
		CC="$CC" \
		AR="$AR" \
		RANLIB="$RANLIB" \
		STRIP="$STRIP" \
		CFLAGS="$cflags" \
		CPPFLAGS="$cppflags" \
		LDFLAGS="$ldflags" \
		PKG_CONFIG="$PKG_CONFIG" \
		./configure \
		"${configure_args[@]}"

	log "building lame"
	make -j"$JOBS"
	make install
	popd >/dev/null
}

build_opus() {
	local target="$1"
	local source_dir="$2"
	local prefix="$3"
	local host="$4"
	local cflags="$5"
	local cppflags="$6"
	local ldflags="$7"

	log "configuring opus"
	pushd "$source_dir" >/dev/null
	make distclean >/dev/null 2>&1 || true

	local configure_args=(
		--prefix="$prefix"
		--disable-shared
		--enable-static
		--with-pic
		--disable-extra-programs
		--disable-doc
		--disable-asm
		--disable-rtcd
	)
	if [[ "$target" == windows-* ]]; then
		configure_args+=(--disable-hardening)
	fi
	if ! is_native_target "$target"; then
		configure_args+=(--host="$host")
	fi
	if [[ "$target" == windows-* ]]; then
		cflags="$cflags -D_FORTIFY_SOURCE=0"
	fi

	env \
		CC="$CC" \
		AR="$AR" \
		RANLIB="$RANLIB" \
		STRIP="$STRIP" \
		CFLAGS="$cflags" \
		CPPFLAGS="$cppflags" \
		LDFLAGS="$ldflags" \
		PKG_CONFIG="$PKG_CONFIG" \
		./configure \
		"${configure_args[@]}"

	log "building opus"
	make -j"$JOBS"
	make install
	popd >/dev/null
}

build_zlib() {
	local target="$1"
	local source_dir="$2"
	local prefix="$3"
	local cflags="$4"
	local cppflags="$5"

	log "configuring zlib"
	pushd "$source_dir" >/dev/null
	make distclean >/dev/null 2>&1 || make clean >/dev/null 2>&1 || true

	if [[ "$target" == windows-* ]]; then
		log "building zlib"
		make -f win32/Makefile.gcc -j"$JOBS" \
			PREFIX="$CROSS_PREFIX" \
			CC="$CC" \
			AR="$AR" \
			RC="${CROSS_PREFIX}windres" \
			CFLAGS="$cflags $cppflags" \
			libz.a
		mkdir -p "$prefix/include" "$prefix/lib" "$prefix/lib/pkgconfig"
		cp zlib.h zconf.h "$prefix/include/"
		cp libz.a "$prefix/lib/"
		cat >"$prefix/lib/pkgconfig/zlib.pc" <<EOF
prefix=$prefix
exec_prefix=\${prefix}
libdir=\${prefix}/lib
includedir=\${prefix}/include

Name: zlib
Description: zlib compression library
Version: $ZLIB_VERSION
Libs: -L\${libdir} -lz
Cflags: -I\${includedir}
EOF
	else
		env \
			CC="$CC" \
			AR="$AR" \
			RANLIB="$RANLIB" \
			CFLAGS="$cflags $cppflags" \
			./configure \
			--static \
			--prefix="$prefix"

		log "building zlib"
		make -j"$JOBS"
		make install
	fi

	popd >/dev/null
}

build_ffmpeg() {
	local target="$1"
	local source_dir="$2"
	local prefix="$3"
	local cflags="$4"
	local cppflags="$5"
	local ldflags="$6"
	local ffmpeg_arch="$7"
	local ffmpeg_os="$8"
	local cross_prefix="$9"

	log "configuring ffmpeg"
	pushd "$source_dir" >/dev/null
	make distclean >/dev/null 2>&1 || true

	local audio_decoders audio_demuxers audio_encoders audio_parsers cover_art_decoders pcm_adpcm_decoders
	audio_demuxers="aa,aac,aax,ac3,aiff,ape,asf,au,caf,dsf,dts,eac3,flac,hca,matroska,mov,mp3,mpc,mpc8,ogg,oma,shorten,tak,tta,voc,w64,wav,wv,xwma"
	audio_decoders="aac,aac_latm,ac3,alac,ape,atrac1,atrac3,atrac3al,atrac3p,atrac3pal,atrac9,cook,dca,dsd_lsbf,dsd_lsbf_planar,dsd_msbf,dsd_msbf_planar,eac3,flac,hca,mace3,mace6,metasound,mp1,mp1float,mp2,mp2float,mp3,mp3adu,mp3adufloat,mp3float,mp3on4,mp3on4float,mpc7,mpc8,opus,qdm2,qdmc,ra_144,ra_288,ralf,shorten,tak,tta,vorbis,wavpack,wmalossless,wmapro,wmav1,wmav2"
	audio_encoders="aac,alac,flac,libmp3lame,libopus,pcm_alaw,pcm_mulaw,pcm_s16be,pcm_s16le,pcm_s24be,pcm_s24le,pcm_s32be,pcm_s32le,pcm_f32be,pcm_f32le,pcm_u8"
	audio_parsers="aac,aac_latm,ac3,cook,dca,flac,mpegaudio,opus,tak,vorbis"
	cover_art_decoders="bmp,mjpeg,png"
	pcm_adpcm_decoders="adpcm_4xm,adpcm_adx,adpcm_afc,adpcm_agm,adpcm_aica,adpcm_argo,adpcm_circus,adpcm_ct,adpcm_dtk,adpcm_ea,adpcm_ea_maxis_xa,adpcm_ea_r1,adpcm_ea_r2,adpcm_ea_r3,adpcm_ea_xas,adpcm_g722,adpcm_g726,adpcm_g726le,adpcm_ima_acorn,adpcm_ima_alp,adpcm_ima_amv,adpcm_ima_apc,adpcm_ima_apm,adpcm_ima_cunning,adpcm_ima_dat4,adpcm_ima_dk3,adpcm_ima_dk4,adpcm_ima_ea_eacs,adpcm_ima_ea_sead,adpcm_ima_escape,adpcm_ima_hvqm2,adpcm_ima_hvqm4,adpcm_ima_iss,adpcm_ima_magix,adpcm_ima_moflex,adpcm_ima_mtf,adpcm_ima_oki,adpcm_ima_pda,adpcm_ima_qt,adpcm_ima_qt_at,adpcm_ima_rad,adpcm_ima_smjpeg,adpcm_ima_ssi,adpcm_ima_wav,adpcm_ima_ws,adpcm_ima_xbox,adpcm_ms,adpcm_mtaf,adpcm_n64,adpcm_psx,adpcm_psxc,adpcm_sanyo,adpcm_sbpro_2,adpcm_sbpro_3,adpcm_sbpro_4,adpcm_swf,adpcm_thp,adpcm_thp_le,adpcm_vima,adpcm_xa,adpcm_xmd,adpcm_yamaha,adpcm_zork,pcm_alaw,pcm_bluray,pcm_dvd,pcm_f16le,pcm_f24le,pcm_f32be,pcm_f32le,pcm_f64be,pcm_f64le,pcm_lxf,pcm_mulaw,pcm_s16be,pcm_s16be_planar,pcm_s16le,pcm_s16le_planar,pcm_s24be,pcm_s24daud,pcm_s24le,pcm_s24le_planar,pcm_s32be,pcm_s32le,pcm_s32le_planar,pcm_s64be,pcm_s64le,pcm_s8,pcm_s8_planar,pcm_sga,pcm_u16be,pcm_u16le,pcm_u24be,pcm_u24le,pcm_u32be,pcm_u32le,pcm_u8,pcm_vidc"

	if [[ "$ffmpeg_os" == "darwin" ]]; then
		audio_encoders="$audio_encoders,aac_at"
	fi

	local configure_args=(
		--prefix="$prefix"
		--pkg-config="$PKG_CONFIG"
		--pkg-config-flags=--static
		--arch="$ffmpeg_arch"
		--target-os="$ffmpeg_os"
		--enable-gpl
		--disable-debug
		--disable-doc
		--disable-network
		--disable-autodetect
		--disable-asm
		--disable-avdevice
		--disable-shared
		--enable-static
		--enable-pic
		--enable-small
		--disable-everything
		--enable-swresample
		--enable-protocol=file,pipe
		--enable-zlib
		--enable-demuxer="$audio_demuxers"
		--enable-muxer=adts,flac,ipod,mov,mp3,ogg,opus,mp4,wav
		--enable-decoder="$audio_decoders,$cover_art_decoders,$pcm_adpcm_decoders"
		--enable-encoder="$audio_encoders"
		--enable-parser="$audio_parsers"
		--enable-bsf=aac_adtstoasc
		--enable-filter=aformat,anull,aresample,atrim,crop,format,hflip,null,rotate,transpose,trim,vflip
		--enable-libmp3lame
		--enable-libopus
	)

	if [[ "$ffmpeg_os" == "darwin" ]]; then
		configure_args+=(--enable-audiotoolbox)
	fi

	local env_args=(
		CC="$CC"
		AR="$AR"
		RANLIB="$RANLIB"
		STRIP="$STRIP"
		PKG_CONFIG="$PKG_CONFIG"
		PKG_CONFIG_PATH="$prefix/lib/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
		PKG_CONFIG_LIBDIR="$prefix/lib/pkgconfig${PKG_CONFIG_LIBDIR:+:$PKG_CONFIG_LIBDIR}"
		CFLAGS="$cflags"
		CPPFLAGS="$cppflags"
		LDFLAGS="$ldflags"
	)

	if ! is_native_target "$target"; then
		configure_args+=(--enable-cross-compile)
		if [[ -n "$cross_prefix" ]]; then
			configure_args+=(--cross-prefix="$cross_prefix")
		fi
	fi

	env "${env_args[@]}" ./configure "${configure_args[@]}"

	log "building ffmpeg"
	make -j"$JOBS"
	make install
	popd >/dev/null
}

build_link_bridge() {
	local target="$1"
	local source_dir="$2"
	local prefix="$3"
	local cflags="$4"
	local cppflags="$5"
	local bridge_src="$SCRIPT_DIR/bridge/linked_ffmpeg_bridge.c"
	local bridge_dir="$source_dir/podhnologic-bridge"
	local objects=()
	local object

	log "building linked ffmpeg bridge"
	mkdir -p "$bridge_dir"

	pushd "$source_dir" >/dev/null
	"$CC" \
		-I. \
		-I"$prefix/include" \
		-I"$prefix/include/opus" \
		-D_GNU_SOURCE \
		-D_ISOC11_SOURCE \
		-D_FILE_OFFSET_BITS=64 \
		-D_LARGEFILE_SOURCE \
		-DPIC \
		-I./compat/dispatch_semaphore \
		-I./compat/stdbit \
		$cflags \
		$cppflags \
		-std=c17 \
		-fno-common \
		-fomit-frame-pointer \
		-pthread \
		-Wall \
		-Wno-parentheses \
		-Wno-switch \
		-Wno-format-zero-length \
		-Wno-pointer-sign \
		-Wno-unused-const-variable \
		-Wno-bool-operation \
		-Wno-char-subscripts \
		-Wno-implicit-const-int-float-conversion \
		-Wno-microsoft-enum-forward-reference \
		-Werror=implicit-function-declaration \
		-Werror=return-type \
		-Wno-missing-prototypes \
		-Dmain=podhnologic_ffmpeg_main \
		-c fftools/ffmpeg.c \
		-o "$bridge_dir/ffmpeg_main.o"

	"$CC" \
		-I. \
		-I"$prefix/include" \
		-I"$prefix/include/opus" \
		-D_GNU_SOURCE \
		-D_ISOC11_SOURCE \
		-D_FILE_OFFSET_BITS=64 \
		-D_LARGEFILE_SOURCE \
		-DPIC \
		-I./compat/dispatch_semaphore \
		-I./compat/stdbit \
		$cflags \
		$cppflags \
		-std=c17 \
		-fno-common \
		-fomit-frame-pointer \
		-pthread \
		-Wall \
		-Wno-parentheses \
		-Wno-switch \
		-Wno-format-zero-length \
		-Wno-pointer-sign \
		-Wno-unused-const-variable \
		-Wno-bool-operation \
		-Wno-char-subscripts \
		-Wno-implicit-const-int-float-conversion \
		-Wno-microsoft-enum-forward-reference \
		-Werror=implicit-function-declaration \
		-Werror=return-type \
		-Wno-missing-prototypes \
		-Dmain=podhnologic_ffprobe_main \
		-Dprogram_name=podhnologic_ffprobe_program_name \
		-Dprogram_birth_year=podhnologic_ffprobe_program_birth_year \
		-Dshow_help_default=podhnologic_ffprobe_show_help_default \
		-c fftools/ffprobe.c \
		-o "$bridge_dir/ffprobe_main.o"

	if [[ "$target" == windows-* ]]; then
		"$OBJCOPY" --redefine-sym main=podhnologic_ffmpeg_main "$bridge_dir/ffmpeg_main.o" "$bridge_dir/ffmpeg_main.renamed.o"
		mv "$bridge_dir/ffmpeg_main.renamed.o" "$bridge_dir/ffmpeg_main.o"
		"$OBJCOPY" --redefine-sym main=podhnologic_ffprobe_main "$bridge_dir/ffprobe_main.o" "$bridge_dir/ffprobe_main.renamed.o"
		mv "$bridge_dir/ffprobe_main.renamed.o" "$bridge_dir/ffprobe_main.o"
	fi

	"$CC" \
		$cflags \
		$cppflags \
		-D_GNU_SOURCE \
		-std=c17 \
		-Wall \
		-Werror=implicit-function-declaration \
		-Werror=return-type \
		-c "$bridge_src" \
		-o "$bridge_dir/linked_ffmpeg_bridge.o"

	objects+=(
		"$bridge_dir/linked_ffmpeg_bridge.o"
		"$bridge_dir/ffmpeg_main.o"
		"$bridge_dir/ffprobe_main.o"
	)

	while IFS= read -r object; do
		case "$object" in
			fftools/ffmpeg.o|fftools/ffprobe.o) ;;
			*) objects+=("$object") ;;
		esac
	done < <(find fftools -type f -name '*.o' | sort)

	rm -f "$prefix/lib/libpodhnologicffmpeg.a"
	"$AR" rcs "$prefix/lib/libpodhnologicffmpeg.a" "${objects[@]}"
	"$RANLIB" "$prefix/lib/libpodhnologicffmpeg.a"
	popd >/dev/null
}

build_target() {
	local target="$1"
	local target_root="$WORK_ROOT/$target"
	local source_root="$target_root/src"
	local prefix="$DIST_ROOT/$target"
	local ffmpeg_arch ffmpeg_os cross_prefix cflags cppflags ldflags autotools_host
	local zlib_src lame_src opus_src ffmpeg_src

	CURRENT_TARGET="$target"
	autotools_host="$(target_autotools_host "$target")"
	ffmpeg_arch="$(target_arch "$target")"
	ffmpeg_os="$(target_os "$target")"
	cross_prefix="$(target_cross_prefix "$target")"
	cflags="$(target_cflags "$target")"
	cppflags="-I$prefix/include"
	ldflags="$(target_ldflags "$target") -L$prefix/lib"

	mkdir -p "$source_root" "$prefix"

	require_cmd curl
	require_cmd tar
	require_cmd make
	require_cmd awk

	target_toolchain "$target"

	ensure_source_tree lame "$LAME_VERSION" "$LAME_TARBALL_URL" "$LAME_TARBALL_SHA256" "$source_root"
	ensure_source_tree opus "$OPUS_VERSION" "$OPUS_TARBALL_URL" "$OPUS_TARBALL_SHA256" "$source_root"
	ensure_source_tree zlib "$ZLIB_VERSION" "$ZLIB_TARBALL_URL" "$ZLIB_TARBALL_SHA256" "$source_root"
	ensure_source_tree ffmpeg "$FFMPEG_VERSION" "$FFMPEG_TARBALL_URL" "$FFMPEG_TARBALL_SHA256" "$source_root"

	zlib_src="$source_root/zlib-$ZLIB_VERSION"
	lame_src="$source_root/lame-$LAME_VERSION"
	opus_src="$source_root/opus-$OPUS_VERSION"
	ffmpeg_src="$source_root/ffmpeg-$FFMPEG_VERSION"

	build_zlib "$target" "$zlib_src" "$prefix" "$cflags" "$cppflags"
	build_lame "$target" "$lame_src" "$prefix" "$autotools_host" "$cflags" "$cppflags" "$ldflags"
	build_opus "$target" "$opus_src" "$prefix" "$autotools_host" "$cflags" "$cppflags" "$ldflags"
	build_ffmpeg "$target" "$ffmpeg_src" "$prefix" "$cflags" "$cppflags" "$ldflags" "$ffmpeg_arch" "$ffmpeg_os" "$cross_prefix"
	build_link_bridge "$target" "$ffmpeg_src" "$prefix" "$cflags" "$cppflags"

	log "built bundle at $prefix"
}

while [[ $# -gt 0 ]]; do
	case "$1" in
		--all)
			TARGETS=("${ALL_TARGETS[@]}")
			shift
			;;
		--target)
			[[ $# -ge 2 ]] || die "--target needs an argument"
			TARGETS+=("$2")
			shift 2
			;;
		--clean)
			CLEAN=1
			shift
			;;
		--jobs|-j)
			[[ $# -ge 2 ]] || die "--jobs needs an argument"
			JOBS="$2"
			shift 2
			;;
		--help|-h)
			usage
			exit 0
			;;
		*)
			die "unknown argument: $1"
			;;
	esac
done

if [[ ${#TARGETS[@]} -eq 0 ]]; then
	TARGETS=("$(default_target)")
fi

if [[ -z "$JOBS" ]]; then
	JOBS="$(default_jobs)"
fi

if [[ "$CLEAN" -eq 1 ]]; then
	for target in "${TARGETS[@]}"; do
		rm -rf "$WORK_ROOT/$target" "$DIST_ROOT/$target"
	done
fi

for target in "${TARGETS[@]}"; do
	case " ${ALL_TARGETS[*]} " in
		*" $target "*) ;;
		*) die "unsupported target: $target" ;;
	esac
	build_target "$target"
done
