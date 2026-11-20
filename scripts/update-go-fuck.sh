#!/bin/sh
#
# update-go-fuck.sh installs the latest go-fuck release on Linux or macOS.
#
# Usage: update-go-fuck.sh [VERSION] [INSTALL_DIR]
#
#   VERSION      release version, for example 2.11.0 (default: latest)
#   INSTALL_DIR  directory for the go-fuck binary (default: ~/.local/bin)
#
# The script downloads the release archive and checksums.txt, verifies the
# SHA256 checksum and atomically replaces the binary. Each request has a
# connection timeout and a bounded transfer time, with up to three attempts.
set -eu

repo=leafevans/go-fuck
version=${1:-latest}
install_dir=${2:-"$HOME/.local/bin"}

fail() {
  echo "update-go-fuck: $*" >&2
  exit 1
}

valid_version() {
  case "$1" in ""|*[!0-9.]*) return 1 ;; esac
  printf '%s\n' "$1" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'
}

# Run in a subshell so retry state cannot overwrite the caller's variables.
download() (
  output=$1
  source_url=$2
  timeout=$3
  shift 3
  attempt=1
  while [ "$attempt" -le 3 ]; do
    if curl -fsS --connect-timeout 10 --max-time "$timeout" \
        --user-agent go-fuck-updater "$@" -o "$output" "$source_url"; then
      return 0
    fi
    rm -f "$output"
    if [ "$attempt" -eq 3 ]; then
      fail "download failed after 3 attempts: $source_url"
    fi
    echo "update-go-fuck: download failed (attempt $attempt/3); retrying: $source_url" >&2
    sleep 2
    attempt=$((attempt + 1))
  done
)

[ "$#" -le 2 ] || fail "usage: update-go-fuck.sh [VERSION] [INSTALL_DIR]"
if [ "$version" != "latest" ]; then
  version=${version#v}
  valid_version "$version" || fail "invalid version: expected X.Y.Z or vX.Y.Z"
fi

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported operating system" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) fail "unsupported architecture" ;;
esac

if command -v sha256sum >/dev/null 2>&1; then
  checksum_tool=sha256sum
elif command -v shasum >/dev/null 2>&1; then
  checksum_tool=shasum
else
  fail "SHA256 verification requires sha256sum or shasum"
fi

tmp=$(mktemp -d)
staged=
cleanup() {
  rm -rf "$tmp"
  if [ -n "$staged" ]; then rm -f "$staged"; fi
}
trap cleanup 0
trap 'exit 130' INT
trap 'exit 143' HUP TERM

if [ "$version" = "latest" ]; then
  tag=
  if download "$tmp/release.json" "https://api.github.com/repos/$repo/releases/latest" 20 -L; then
    tag=$(grep -o '"tag_name": *"[^"]*"' "$tmp/release.json" | head -n 1 | cut -d'"' -f4)
  fi
  if ! valid_version "${tag#v}"; then
    # The API is rate limited for unauthenticated requests; the releases/latest
    # page redirects to /releases/tag/vX.Y.Z.
    download "$tmp/headers" "https://github.com/$repo/releases/latest" 20 -I ||
      fail "cannot resolve the latest version"
    tag=$(awk 'tolower($1) == "location:" { print $2 }' "$tmp/headers" | tr -d '\r' | head -n 1)
    case "$tag" in
      "https://github.com/$repo/releases/tag/v"*) ;;
      *) fail "cannot resolve the latest version" ;;
    esac
    tag=${tag##*/}
  fi
  version=${tag#v}
  valid_version "$version" || fail "cannot resolve the latest version"
fi

tag="v$version"
base="go-fuck_${version}_${os}_${arch}"
asset="$base.tar.gz"
url="https://github.com/$repo/releases/download/$tag"

echo "downloading go-fuck $version ($os/$arch)"
download "$tmp/$asset" "$url/$asset" 120 -L
download "$tmp/checksums.txt" "$url/checksums.txt" 20 -L

# Match exactly one filename, rather than a substring or a regular expression.
expected=$(awk -v asset="$asset" 'NF == 2 && $2 == asset { hash = $1; count++ }
  END { if (count != 1) exit 1; print hash }' "$tmp/checksums.txt") ||
  fail "expected exactly one checksum for $asset"
printf '%s\n' "$expected" | grep -Eq '^[[:xdigit:]]{64}$' ||
  fail "invalid SHA256 checksum for $asset"
expected=$(printf '%s' "$expected" | tr 'A-F' 'a-f')
if [ "$checksum_tool" = "sha256sum" ]; then
  actual=$(sha256sum "$tmp/$asset")
else
  actual=$(shasum -a 256 "$tmp/$asset")
fi
actual=${actual%% *}
[ "$expected" = "$actual" ] || fail "checksum verification failed for $asset"

tar -xzf "$tmp/$asset" -C "$tmp" go-fuck || fail "cannot extract go-fuck from $asset"
[ -f "$tmp/go-fuck" ] && [ ! -L "$tmp/go-fuck" ] || fail "archive does not contain a regular go-fuck binary"
[ ! -d "$install_dir/go-fuck" ] || fail "installation target is a directory: $install_dir/go-fuck"
mkdir -p "$install_dir"
# Stage on the destination filesystem so a failed install preserves the old
# executable and the final rename replaces it atomically.
staged=$(mktemp "$install_dir/.go-fuck.XXXXXX")
install -m 755 "$tmp/go-fuck" "$staged" || fail "cannot stage the new binary"
mv -f "$staged" "$install_dir/go-fuck" || fail "cannot replace the installed binary"
staged=
echo "go-fuck $version installed in $install_dir"
