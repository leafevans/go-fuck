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
check_setup bash .bashrc
check_setup zsh .zshrc
check_setup fish .config/fish/config.fish
check_setup tcsh .tcshrc

echo "shell integration: all checks passed"
