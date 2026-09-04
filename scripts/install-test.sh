#!/bin/sh
set -eu

# Isolated checks for install.sh's upgrade confirmation. The fake tools below
# prevent downloads and installation from touching the network or the user's
# home directory.
ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

fake_bin="$TEST_DIR/bin"
mkdir -p "$fake_bin"

printf '#!/bin/sh\nexit 0\n' > "$fake_bin/podhnologic"
chmod 755 "$fake_bin/podhnologic"
printf '#!/bin/sh\nprintf downloader-reached >&2\nexit 99\n' > "$fake_bin/curl"
chmod 755 "$fake_bin/curl"

set +e
output=$(env PATH="$fake_bin:/usr/bin:/bin" SHELL=/bin/sh \
    sh "$ROOT_DIR/install.sh" </dev/null 2>&1)
status=$?
set -e

[ "$status" -eq 1 ] || {
    printf 'expected noninteractive upgrade to fail with status 1, got %s\n%s\n' "$status" "$output" >&2
    exit 1
}
case "$output" in
    *"no controlling terminal"*"--yes"*) ;;
    *) printf 'missing noninteractive guidance:\n%s\n' "$output" >&2; exit 1 ;;
esac

pty_output="$TEST_DIR/pty-output"
printf 'n\n' | script -q "$pty_output" env \
    PATH="$fake_bin:/usr/bin:/bin" SHELL=/bin/sh sh -c 'cat "$1" | sh' sh "$ROOT_DIR/install.sh" \
    > "$TEST_DIR/pty-transcript" 2>&1
pty_transcript=$(sed 's/\r//g' "$TEST_DIR/pty-output")
case "$pty_transcript" in
    *"Update to the latest version?"*"Installation cancelled."*) ;;
    *) printf 'interactive PTY prompt test failed:\n%s\n' "$pty_transcript" >&2; exit 1 ;;
esac

set +e
yes_output=$(env PATH="$fake_bin:/usr/bin:/bin" SHELL=/bin/sh \
    sh "$ROOT_DIR/install.sh" --yes </dev/null 2>&1)
yes_status=$?
set -e
[ "$yes_status" -eq 99 ] || {
    printf 'expected --yes run to reach fake downloader (status 99), got %s\n%s\n' "$yes_status" "$yes_output" >&2
    exit 1
}
case "$yes_output" in
    *downloader-reached*) ;;
    *) printf -- '--yes did not reach downloader:\n%s\n' "$yes_output" >&2; exit 1 ;;
esac

help_output=$(env PATH="$fake_bin:/usr/bin:/bin" \
    sh "$ROOT_DIR/install.sh" --help 2>&1)
case "$help_output" in
    *"--yes"*) ;;
    *) printf 'missing --yes help:\n%s\n' "$help_output" >&2; exit 1 ;;
esac

printf '%s\n' 'installer confirmation tests passed'
