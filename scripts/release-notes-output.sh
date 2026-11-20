#!/usr/bin/env bash
set -euo pipefail

tag=${GITHUB_REF_NAME:?missing tag}
if [[ ! $tag =~ ^v[0-9]+[.][0-9]+[.][0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
    printf 'Invalid release tag: %s\n' "$tag" >&2
    exit 1
fi
path="release-notes/$tag.md"
bash "$(dirname "$0")/check-release-notes.sh" "$path"
printf 'path=%s\n' "$path" >> "${GITHUB_OUTPUT:?missing output file}"
