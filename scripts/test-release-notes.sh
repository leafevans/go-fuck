#!/usr/bin/env bash
set -euo pipefail

script=$(cd "$(dirname "$0")" && pwd)/release-notes-output.sh
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
cd "$work"
mkdir release-notes
export GITHUB_OUTPUT="$work/output"
printf '## Changelog\n\n- Fixed quoting.\n' > release-notes/v1.2.3.md
GITHUB_REF_NAME=v1.2.3 bash "$script"
printf 'path=release-notes/v1.2.3.md\n' | cmp - "$GITHUB_OUTPUT"

cp release-notes/v1.2.3.md release-notes/v1x2x3.md
printf ' \n' > release-notes/v1.2.4.md
printf '## Changelog\n' > release-notes/v1.2.5.md
for tag in v1x2x3 v1.2.4 v1.2.5 v1.2.6; do
    : > "$GITHUB_OUTPUT"
    if GITHUB_REF_NAME=$tag bash "$script"; then
        printf 'Unexpected success for %s\n' "$tag" >&2
        exit 1
    fi
    test ! -s "$GITHUB_OUTPUT"
done
