#!/usr/bin/env bash
# Exercise a generated short alias against a private, deterministic history.
set -eu
bin=$1
cd "$2"
export HOME=$PWD HISTFILE=$PWD/history XDG_CONFIG_HOME=$PWD/config TF_SHELL=bash
export GOFUCK_RULES=history GOFUCK_ALTER_HISTORY=false
mkdir history-bin
printf '%s\n' '#!/bin/sh' 'case "$*" in *x) echo typo >&2; exit 1;; esac' 'echo "$*"' > history-bin/fd
chmod +x history-bin/fd
ln -s fd history-bin/fzf
export PATH="$PWD/history-bin:$(dirname "$bin"):$PATH"
source <("$bin" --alias f --capture-mode=rerun)
printf '%s\n' 'fd wrong' 'fd wrongx' 'f --yes' 'fzf --filter file' 'git stash' > "$HISTFILE"

# The fake tools echo their arguments, so stdout must contain only the
# corrected command's result. The closer but failed 'fd wrongx' must lose.
check_correction() {
  history -c
  history -s "$1"
  history -s 'f --yes'
  f --yes >actual
  printf '%s\n' "$2" | diff -u - actual
}
check_correction 'fd wrongxx' 'wrong'
check_correction 'fzf --filter filex' '--filter file'

export GOFUCK_RULES=git_not_command:history
printf '%s\n' "git: 'stats' is not a git command. See 'git --help'." '' 'The most similar command is' '    status' > failure.log
f --json --no-rerun --output-from failure.log --force-command 'git stats' > ranked.json
python3 -c 'import json, sys; cmds = [s["command"] for s in json.load(sys.stdin)["suggestions"]]; assert cmds[:1] == ["git status"] and "git stash" in cmds[1:], cmds' < ranked.json
