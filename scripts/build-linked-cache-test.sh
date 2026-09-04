#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

mkdir -p "$tmp_dir/scripts" "$tmp_dir/cmd/podhnologic" "$tmp_dir/bin"
mkdir -p "$tmp_dir/scripts/ffmpeg/bridge"
cp "$root_dir/scripts/ffmpeg/versions.env" "$root_dir/scripts/ffmpeg/build-source.sh" "$root_dir/scripts/ffmpeg/build-native.sh" "$tmp_dir/scripts/ffmpeg/"
cp "$root_dir/scripts/ffmpeg/bridge/linked_ffmpeg_bridge.c" "$tmp_dir/scripts/ffmpeg/bridge/"
cp "$root_dir/scripts/build-linked.sh" "$tmp_dir/scripts/"

prefix="$tmp_dir/prefix"
mkdir -p "$prefix/include" "$prefix/lib/pkgconfig" "$prefix/share/podhnologic"
printf ffmpeg >"$prefix/lib/libpodhnologicffmpeg.a"
printf avcodec >"$prefix/lib/libavcodec.a"
printf '%s\n' 'LGPL version 2.1 or later' >"$prefix/share/podhnologic/ffmpeg-license.txt"
printf '%s\n' '--enable-static' >"$prefix/share/podhnologic/ffmpeg-configure-args.txt"

fingerprint() {
	FFMPEG_FINGERPRINT_ONLY=1 "$tmp_dir/scripts/ffmpeg/build-source.sh" \
		--print-fingerprint --target darwin-arm64 --prefix "$prefix"
}
printf '%s\n' "$(fingerprint)" >"$prefix/share/podhnologic/native-build-fingerprint.txt"

printf '#!/usr/bin/env bash\nset -e\nprintf "%%s\\n" "$@" >"${GO_LOG_DIR}/go-${GO_RUN:-unknown}.args"\nif [[ "${FAIL_GO:-0}" == 1 ]]; then exit 42; fi\n' >"$tmp_dir/bin/go"
printf '#!/usr/bin/env bash\nexit 0\n' >"$tmp_dir/bin/pkg-config"
chmod +x "$tmp_dir/bin/go" "$tmp_dir/bin/pkg-config"

run_linked() {
	local run="$1"
	GO_RUN="$run" GO_LOG_DIR="$tmp_dir" FFMPEG_PREFIX="$prefix" GO_BIN="$tmp_dir/bin/go" PATH="$tmp_dir/bin:$PATH" \
		"$tmp_dir/scripts/build-linked.sh" darwin-arm64 >/dev/null
}

run_linked 1
grep -qx -- '-a' "$tmp_dir/go-1.args"
run_linked 2
! grep -qx -- '-a' "$tmp_dir/go-2.args"

printf changed >>"$prefix/lib/libavcodec.a"
run_linked 3
grep -qx -- '-a' "$tmp_dir/go-3.args"

sidecar="$tmp_dir/build/.podhnologic-darwin-arm64.native-archive.sha256"
sidecar_before="$(<"$sidecar")"
printf another-change >>"$prefix/lib/libavcodec.a"
if FAIL_GO=1 run_linked 4; then
	printf 'failed fake Go unexpectedly succeeded\n' >&2
	exit 1
fi
[[ "$(<"$sidecar")" == "$sidecar_before" ]]

base_fingerprint="$(fingerprint)"
for input in bridge/linked_ffmpeg_bridge.c versions.env build-source.sh; do
	file="$tmp_dir/scripts/ffmpeg/$input"
	cp "$file" "$tmp_dir/original"
	printf '\n# cache regression input change\n' >>"$file"
	[[ "$(fingerprint)" != "$base_fingerprint" ]]
	if run_linked stale >"$tmp_dir/stale.log" 2>&1; then
		printf 'stale prefix unexpectedly accepted after changing %s\n' "$input" >&2
		exit 1
	fi
	grep -q 'missing current LGPL-only build metadata' "$tmp_dir/stale.log"
	mv "$tmp_dir/original" "$file"
done
[[ "$(fingerprint)" == "$base_fingerprint" ]]
[[ -z "$(find "$tmp_dir/build" -name 'podhnologic-*' -print)" ]]
printf 'build-linked cache checks passed\n'
