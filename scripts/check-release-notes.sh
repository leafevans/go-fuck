#!/usr/bin/env bash
set -euo pipefail

notes=${1:?usage: check-release-notes.sh FILE}
if [[ ! -f "$notes" ]]; then
    printf 'Release notes not found: %s\\n' "$notes" >&2
    exit 1
fi

# Release notes must contain at least one visible changelog bullet.
if ! awk '
    {
        rest = $0
        line = ""
        while (1) {
            if (comment) {
                end = index(rest, "-->")
                if (!end) break
                rest = substr(rest, end + 3)
                comment = 0
            }
            start = index(rest, "<!--")
            if (!start) {
                line = line rest
                break
            }
            line = line substr(rest, 1, start - 1)
            rest = substr(rest, start + 4)
            comment = 1
        }
        if (line ~ /^[-*+][[:blank:]]+.*[[:alnum:]]/) found = 1
    }
    END { exit !found }
' "$notes"; then
    printf 'Release notes need at least one non-empty Markdown bullet: %s\\n' "$notes" >&2
    exit 1
fi
