#!/usr/bin/env bash
# Feed separate input lines to interactive zsh so it records real history events.
set -eu
bin=$1
mkdir "$2/zsh-history"
cd "$2/zsh-history"
mkdir bin
printf '%s\n' '#!/bin/sh' 'echo "fixture: fd failed" >&2' 'exit 1' >bin/fd
printf '%s\n' '#!/bin/sh' 'echo "fdd ran"' >bin/fdd
chmod +x bin/fd bin/fdd

env -i PATH="$PWD/bin:$(dirname "$bin"):/usr/bin:/bin" HOME="$PWD" ZDOTDIR="$PWD" \
  HISTFILE="$PWD/history" TF_SHELL=zsh GOFUCK_RULES=history GOFUCK_ALTER_HISTORY=false \
  timeout 15 zsh -fis >actual <<'ZSH'
HISTSIZE=100 SAVEHIST=100
setopt EXTENDED_HISTORY INC_APPEND_HISTORY
eval "$(go-fuck --alias f --capture-mode=rerun)"
fdd
fd
f --yes
f --yes >repeated.out 2>repeated.err
exit
ZSH

# Consecutive corrections skip alias invocations and find the same failed command.
printf '%s\n' 'fdd ran' 'fdd ran' | diff -u - actual
printf '%s\n' 'fdd ran' | diff -u - repeated.out
grep -Fq 'Suggest a similar command from shell history' repeated.err
