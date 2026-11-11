#!/usr/bin/env bash
#
# check-shell-integration.sh verifies the generated shell aliases and the
# --setup configuration for bash, zsh, fish and tcsh. It is used by CI on Linux
# and can be run locally when those shells are installed:
#
#   go build -trimpath -o /tmp/go-fuck ./cmd/go-fuck
#   bash scripts/check-shell-integration.sh /tmp/go-fuck
#
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: $0 PATH_TO_GO_FUCK" >&2
  exit 2
fi

bin=$1
case "$bin" in
  /*) ;;
  *) bin="$(cd "$(dirname "$bin")" && pwd)/$(basename "$bin")" ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# marker_literal is the octal escaped form of the invisible prompt marker that
# the generated marked alias assigns to __gofuck_mark.
marker_literal='\342\200\213'

fail() {
  echo "shell integration: $*" >&2
  exit 1
}

env -u NO_COLOR "$bin" --version >/dev/null 2>&1 || fail "go-fuck --version failed"

check_alias() {
  shell=$1
  case "$shell" in
    bash)
      env TF_SHELL=bash "$bin" --alias >"$work/alias-bash"
      bash --norc -c 'eval "$(<"$1")"; type fuck >/dev/null' _ "$work/alias-bash" ||
        fail "bash cannot evaluate the generated alias"
      ;;
    zsh)
      env TF_SHELL=zsh "$bin" --alias >"$work/alias-zsh"
      zsh -f -c 'eval "$(<"$1")"; type fuck >/dev/null' _ "$work/alias-zsh" ||
        fail "zsh cannot evaluate the generated alias"
      ;;
    fish)
      env TF_SHELL=fish "$bin" --alias >"$work/alias-fish"
      fish --no-config -c 'source $argv[1]; functions -q fuck' "$work/alias-fish" ||
        fail "fish cannot evaluate the generated alias"
      ;;
    tcsh)
      env TF_SHELL=tcsh "$bin" --alias >"$work/alias-tcsh"
      tcsh -f -c "eval `cat '$work/alias-tcsh'`; alias fuck >/dev/null" ||
        fail "tcsh cannot evaluate the generated alias"
      ;;
    *)
      fail "unknown shell $shell"
      ;;
  esac
}

check_instant_alias() {
  shell=$1
  case "$shell" in
    bash)
      env TF_SHELL=bash "$bin" --alias --enable-experimental-instant-mode >"$work/instant-bash"
      bash -n "$work/instant-bash" || fail "bash cannot parse the instant mode initialization"
      ;;
    zsh)
      env TF_SHELL=zsh "$bin" --alias --enable-experimental-instant-mode >"$work/instant-zsh"
      zsh -n "$work/instant-zsh" || fail "zsh cannot parse the instant mode initialization"
      ;;
  esac
}

# check_marked_alias verifies the marked alias of the logged instant mode shell.
#
# The marked alias is the alias inside the logged shell: it prepends the
# invisible prompt marker to PS1 so that go-fuck can locate the output of the
# last command in the terminal log. The checks generate the marked alias for a
# logged shell, verify that the shell parses it and check that the marker is
# present, then evaluate the alias with a realistic prompt and verify that the
# alias function is defined.
check_marked_alias() {
  shell=$1
  case "$shell" in
    bash)
      env GOFUCK_INSTANT_MODE=True TF_SHELL=bash "$bin" --alias --enable-experimental-instant-mode >"$work/marked-bash"
      grep -qF '__gofuck_mark=' "$work/marked-bash" ||
        fail "the bash marked alias does not define the prompt marker"
      grep -qF "$marker_literal" "$work/marked-bash" ||
        fail "the bash marked alias does not contain the encoded prompt marker"
      bash -n "$work/marked-bash" || fail "bash cannot parse the marked alias"
      # A bira-like multi line prompt with ANSI colors.
      prompt='\[\e[0;32m\]╭─\[\e[0m\]\[\e[1;34m\]\u@\h\[\e[0m\] \[\e[0;32m\]\w\[\e[0m\]\n\[\e[0;32m\]╰─$ \[\e[0m\]'
      bash --norc -c 'PS1="$1"; eval "$(<"$2")"; type fuck >/dev/null' _ "$prompt" "$work/marked-bash" ||
        fail "bash cannot evaluate the marked alias with a bira-like prompt"
      ;;
    zsh)
      env GOFUCK_INSTANT_MODE=True TF_SHELL=zsh "$bin" --alias --enable-experimental-instant-mode >"$work/marked-zsh"
      grep -qF '__gofuck_mark=' "$work/marked-zsh" ||
        fail "the zsh marked alias does not define the prompt marker"
      grep -qF "$marker_literal" "$work/marked-zsh" ||
        fail "the zsh marked alias does not contain the encoded prompt marker"
      zsh -n "$work/marked-zsh" || fail "zsh cannot parse the marked alias"
      # A powerlevel-like prompt with %F colors.
      prompt='%F{green}╭─%f%F{cyan}%n@%m%f %F{blue}%~%f
%F{green}╰─$ %f'
      zsh -f -c 'PROMPT="$1"; eval "$(cat "$2")"; type fuck >/dev/null' _ "$prompt" "$work/marked-zsh" ||
        fail "zsh cannot evaluate the marked alias with a powerlevel-like prompt"
      ;;
  esac
}

check_setup() {
  shell=$1
  config=$2
  home="$work/home-$shell"
  mkdir -p "$(dirname "$home/$config")"
  : >"$home/$config"
  env HOME="$home" XDG_CONFIG_HOME="$home/.config" TF_SHELL="$shell" \
    "$bin" --setup >"$work/setup-$shell" 2>&1 ||
    fail "$shell --setup failed"
  grep -q -- '--alias' "$home/$config" ||
    fail "$shell --setup did not write the alias into $config"
  env HOME="$home" XDG_CONFIG_HOME="$home/.config" TF_SHELL="$shell" \
    "$bin" --setup >>"$work/setup-$shell" 2>&1 ||
    fail "the second $shell --setup failed"
  grep -q 'already configured' "$work/setup-$shell" ||
    fail "$shell --setup is not idempotent"
}

check_alias bash
check_alias zsh
check_alias fish
check_alias tcsh
check_instant_alias bash
check_instant_alias zsh
check_marked_alias bash
check_marked_alias zsh
check_setup bash .bashrc
check_setup zsh .zshrc
check_setup fish .config/fish/config.fish
check_setup tcsh .tcshrc

echo "shell integration: all checks passed"
