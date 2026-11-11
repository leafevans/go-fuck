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
# SHA256 checksum and replaces the binary.
set -eu

repo=leafevans/go-fuck
version=${1:-latest}
install_dir=${2:-"$HOME/.local/bin"}

if [ "$version" = "latest" ]; then
  tag=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" |
    grep -o '"tag_name": *"[^"]*"' | head -n 1 | cut -d'"' -f4)
  version=${tag#v}
  if [ -z "$version" ]; then
    echo "update-go-fuck: cannot resolve the latest version" >&2
    exit 1
  fi
fi

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "update-go-fuck: unsupported operating system" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "update-go-fuck: unsupported architecture" >&2; exit 1 ;;
esac

tag="v$version"
base="go-fuck_${version}_${os}_${arch}"
url="https://github.com/$repo/releases/download/$tag"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading go-fuck $version ($os/$arch)"
curl -fsSL -o "$tmp/$base.tar.gz" "$url/$base.tar.gz"
curl -fsSL -o "$tmp/checksums.txt" "$url/checksums.txt"

(cd "$tmp" && grep "  $base.tar.gz" checksums.txt | sha256sum -c - >/dev/null) || {
  echo "update-go-fuck: checksum verification failed" >&2
  exit 1
}

tar -xzf "$tmp/$base.tar.gz" -C "$tmp"
mkdir -p "$install_dir"
install -m 755 "$tmp/go-fuck" "$install_dir/go-fuck"
echo "go-fuck $version installed in $install_dir"
