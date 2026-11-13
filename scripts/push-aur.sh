#!/usr/bin/env bash
#
# push-aur.sh publishes the AUR package of a go-fuck release.
#
# It validates packaging/aur/PKGBUILD and packaging/aur/.SRCINFO, clones
# ssh://aur@aur.archlinux.org/go-fuck.git, skips the push when both files are
# already published, commits "upgpkg: go-fuck X.Y.Z-pkgrel" and pushes. Run it
# on Linux, macOS or WSL; see --help for the options.
#
set -euo pipefail

package=go-fuck
aur_remote="ssh://aur@aur.archlinux.org/$package.git"
doh_query='https://cloudflare-dns.com/dns-query?name=aur.archlinux.org&type=A'
ssh_command='ssh -o HostKeyAlias=aur.archlinux.org -o StrictHostKeyChecking=accept-new'
max_attempts=3

script_dir=$(cd -- "$(dirname "$0")" && pwd)
repo_root=$(cd -- "$script_dir/.." && pwd)

version=
source_dir="$repo_root/packaging/aur"
clone_dir=
direct_ip=false
dry_run=false

temp_root=
pkgrel=1

cleanup() {
  if [ -n "${temp_root:-}" ]; then
    rm -rf "$temp_root"
  fi
}
trap cleanup EXIT

die() {
  echo "push-aur: $*" >&2
  exit 1
}

usage_error() {
  echo "push-aur: $*" >&2
  echo "push-aur: run 'bash scripts/push-aur.sh --help' for usage" >&2
  exit 2
}

note() {
  echo "push-aur: $*"
}

usage() {
  cat <<'EOF'
push-aur.sh publishes the AUR package of a go-fuck release.

Usage: push-aur.sh [OPTIONS]

Options:
  -v, --version X.Y.Z   release version to publish (default: pkgver in
                        packaging/aur/PKGBUILD); a leading "v" is accepted
      --source-dir DIR  directory that contains the canonical PKGBUILD and
                        .SRCINFO (default: packaging/aur in this repository)
      --clone-dir DIR   clone the AUR repository into DIR and keep it; when
                        omitted a temporary clone is removed after the run
      --direct-ip       resolve aur.archlinux.org over Cloudflare DNS-over-HTTPS
                        and connect to the address with
                        HostKeyAlias=aur.archlinux.org; use this when the
                        network's answer for aur.archlinux.org blocks SSH
      --dry-run         print the plan without cloning, committing or pushing
  -h, --help            print this help

The script checks that PKGBUILD and .SRCINFO declare the same version and
release, clones the AUR repository and exits without pushing when both files
are already published. Otherwise it copies the files, commits
"upgpkg: go-fuck X.Y.Z-pkgrel", pushes to
ssh://aur@aur.archlinux.org/go-fuck.git and removes a temporary clone. The
clone and the push are retried up to three times; --direct-ip resolves the
address again on every attempt. The script needs bash and git, plus curl for
--direct-ip, and an SSH key registered with the AUR:

  bash scripts/push-aur.sh --version X.Y.Z
  bash scripts/push-aur.sh --version X.Y.Z --direct-ip
  bash scripts/push-aur.sh --dry-run
EOF
}

# read_pkgbuild_value prints the value of a variable assignment in a PKGBUILD.
read_pkgbuild_value() {
  # $1: file, $2: variable name
  sed -n "s/^$2=\([^[:space:]]*\).*/\1/p" "$1" | head -n 1
}

# read_srcinfo_value prints the value of a field of a .SRCINFO file.
read_srcinfo_value() {
  # $1: file, $2: field name
  sed -n "s/^[[:space:]]*$2[[:space:]]*=[[:space:]]*//p" "$1" | head -n 1
}

# resolve_doh prints the first IPv4 address of aur.archlinux.org from the
# Cloudflare DNS-over-HTTPS JSON endpoint. jq is not required.
resolve_doh() {
  if ! command -v curl >/dev/null 2>&1; then
    echo "push-aur: --direct-ip requires curl" >&2
    return 1
  fi
  local body address
  if ! body=$(curl -fsS --max-time 15 -H 'accept: application/dns-json' "$doh_query"); then
    echo "push-aur: DNS-over-HTTPS lookup for aur.archlinux.org failed" >&2
    return 1
  fi
  address=$(printf '%s\n' "$body" | tr '}' '\n' | awk -F'"' '
    /"type"[[:space:]]*:[[:space:]]*1[,}]/ {
      for (i = 1; i <= NF; i++) {
        if ($i == "data") {
          print $(i + 2)
          exit
        }
      }
    }')
  case "$address" in
    '' | *[!0-9.]*)
      return 1
      ;;
  esac
  local dots=${address//[^.]/}
  if [ "${#dots}" -ne 3 ]; then
    return 1
  fi
  printf '%s\n' "$address"
}

# git_ssh runs a git command with the HostKeyAlias SSH options in --direct-ip
# mode and with the regular SSH configuration otherwise.
git_ssh() {
  if [ "$direct_ip" = true ]; then
    GIT_SSH_COMMAND="$ssh_command" "$@"
  else
    "$@"
  fi
}

# clone_repository clones the AUR repository into clone_path and retries
# transient failures. With --direct-ip the address is resolved again for every
# attempt.
clone_repository() {
  local attempt=1 url
  while [ "$attempt" -le "$max_attempts" ]; do
    rm -rf "$clone_path"
    if [ "$direct_ip" = true ]; then
      if ! url=$(resolve_doh); then
        echo "push-aur: cannot resolve aur.archlinux.org over DNS-over-HTTPS (attempt $attempt/$max_attempts)" >&2
        attempt=$((attempt + 1))
        if [ "$attempt" -le "$max_attempts" ]; then
          sleep 2
        fi
        continue
      fi
      note "resolved aur.archlinux.org to $url (attempt $attempt/$max_attempts)"
      url="ssh://aur@$url/$package.git"
    else
      url=$aur_remote
    fi
    if git_ssh git clone --quiet "$url" "$clone_path"; then
      if [ "$direct_ip" = true ]; then
        # The clone was fetched over the resolved address; keep the canonical
        # remote in a kept clone.
        git -C "$clone_path" remote set-url origin "$aur_remote"
      fi
      return 0
    fi
    echo "push-aur: cloning the AUR repository failed (attempt $attempt/$max_attempts)" >&2
    attempt=$((attempt + 1))
    if [ "$attempt" -le "$max_attempts" ]; then
      echo "push-aur: retrying in 2 seconds" >&2
      sleep 2
    fi
  done
  rm -rf "$clone_path"
  return 1
}

# push_package pushes the commit of the clone with the same retry policy.
push_package() {
  local attempt=1 url
  while [ "$attempt" -le "$max_attempts" ]; do
    if [ "$direct_ip" = true ]; then
      if ! url=$(resolve_doh); then
        echo "push-aur: cannot resolve aur.archlinux.org over DNS-over-HTTPS (attempt $attempt/$max_attempts)" >&2
      else
        note "resolved aur.archlinux.org to $url (attempt $attempt/$max_attempts)"
        if GIT_SSH_COMMAND="$ssh_command" git -C "$clone_path" push --quiet "ssh://aur@$url/$package.git" HEAD:master; then
          return 0
        fi
      fi
    elif git -C "$clone_path" push --quiet origin HEAD:master; then
      return 0
    fi
    echo "push-aur: pushing to the AUR failed (attempt $attempt/$max_attempts)" >&2
    attempt=$((attempt + 1))
    if [ "$attempt" -le "$max_attempts" ]; then
      echo "push-aur: retrying in 2 seconds" >&2
      sleep 2
    fi
  done
  return 1
}

# print_plan prints what a real run would do. It never touches the network.
print_plan() {
  echo "push-aur: dry run: no network requests, nothing is committed or pushed"
  echo "  package:   $package"
  echo "  version:   $version-$pkgrel"
  echo "  source:    $pkgbuild"
  echo "             $srcinfo"
  echo "  commit:    upgpkg: $package $version-$pkgrel"
  echo "  remote:    $aur_remote"
  echo "  clone:     $clone_description"
  if [ "$direct_ip" = true ]; then
    echo "  transport: resolve aur.archlinux.org over Cloudflare DNS-over-HTTPS and"
    echo "             connect with HostKeyAlias=aur.archlinux.org"
  else
    echo "  transport: direct SSH to aur.archlinux.org"
  fi
  echo "  retries:   up to $max_attempts attempts for the clone and the push"
  echo "  steps:"
  echo "    1. clone the AUR repository"
  echo "    2. exit successfully when PKGBUILD and .SRCINFO are already published"
  echo "    3. copy both files, commit and push HEAD:master"
  if [ -n "$clone_dir" ]; then
    echo "  cleanup:   the clone is kept in $clone_description"
  else
    echo "  cleanup:   the temporary clone is removed when the script exits"
  fi
  if command -v git >/dev/null 2>&1; then
    echo "  git:       $(command -v git)"
  else
    echo "  git:       not found (a real run needs git)"
  fi
  if [ "$direct_ip" = true ]; then
    if command -v curl >/dev/null 2>&1; then
      echo "  curl:      $(command -v curl)"
    else
      echo "  curl:      not found (--direct-ip needs curl)"
    fi
  fi
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -v | --version)
      [ "$#" -ge 2 ] || usage_error "missing argument for $1"
      version=$2
      shift 2
      ;;
    --version=*)
      version=${1#*=}
      shift
      ;;
    --source-dir)
      [ "$#" -ge 2 ] || usage_error "missing argument for $1"
      source_dir=$2
      shift 2
      ;;
    --source-dir=*)
      source_dir=${1#*=}
      shift
      ;;
    --clone-dir)
      [ "$#" -ge 2 ] || usage_error "missing argument for $1"
      clone_dir=$2
      shift 2
      ;;
    --clone-dir=*)
      clone_dir=${1#*=}
      shift
      ;;
    --direct-ip)
      direct_ip=true
      shift
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      usage_error "unknown option: $1"
      ;;
    *)
      usage_error "unexpected argument: $1"
      ;;
  esac
done
if [ "$#" -gt 0 ]; then
  usage_error "unexpected argument: $1"
fi

if [ ! -d "$source_dir" ]; then
  die "source directory not found: $source_dir"
fi
source_dir=$(cd "$source_dir" && pwd)
pkgbuild=$source_dir/PKGBUILD
srcinfo=$source_dir/.SRCINFO
if [ ! -f "$pkgbuild" ]; then
  die "missing $source_dir/PKGBUILD"
fi
if [ ! -f "$srcinfo" ]; then
  die "missing $source_dir/.SRCINFO"
fi

pkgbuild_version=$(read_pkgbuild_value "$pkgbuild" pkgver)
pkgbuild_pkgrel=$(read_pkgbuild_value "$pkgbuild" pkgrel)
srcinfo_version=$(read_srcinfo_value "$srcinfo" pkgver)
srcinfo_pkgrel=$(read_srcinfo_value "$srcinfo" pkgrel)

if [ -z "$pkgbuild_version" ]; then
  die "PKGBUILD does not declare pkgver"
fi
if [ -z "$pkgbuild_pkgrel" ]; then
  die "PKGBUILD does not declare pkgrel"
fi
if [ -z "$srcinfo_version" ]; then
  die ".SRCINFO does not declare pkgver"
fi
if [ -z "$srcinfo_pkgrel" ]; then
  die ".SRCINFO does not declare pkgrel"
fi
if [ "$pkgbuild_version" != "$srcinfo_version" ]; then
  die "PKGBUILD pkgver ($pkgbuild_version) and .SRCINFO pkgver ($srcinfo_version) disagree"
fi
if [ "$pkgbuild_pkgrel" != "$srcinfo_pkgrel" ]; then
  die "PKGBUILD pkgrel ($pkgbuild_pkgrel) and .SRCINFO pkgrel ($srcinfo_pkgrel) disagree"
fi

version=${version#v}
if [ -z "$version" ]; then
  version=$pkgbuild_version
fi
if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  die "invalid version '$version': expected X.Y.Z"
fi
if [ "$pkgbuild_version" != "$version" ]; then
  die "the package files are at version $pkgbuild_version, but $version was requested; regenerate them first (see packaging/README.md)"
fi
if ! [[ "$pkgbuild_pkgrel" =~ ^[1-9][0-9]*$ ]]; then
  die "invalid pkgrel '$pkgbuild_pkgrel' in PKGBUILD"
fi
pkgrel=$pkgbuild_pkgrel

if [ -z "$clone_dir" ]; then
  clone_description="temporary directory (removed after the run)"
else
  case "$clone_dir" in
    /*) ;;
    *) clone_dir=$PWD/$clone_dir ;;
  esac
  clone_description=$clone_dir
fi

if [ "$dry_run" = true ]; then
  print_plan
  exit 0
fi

command -v git >/dev/null 2>&1 || die "git is required to publish the AUR package"

if [ -n "$clone_dir" ]; then
  if [ -e "$clone_dir" ]; then
    die "clone directory already exists: $clone_dir (remove it or omit --clone-dir)"
  fi
  clone_path=$clone_dir
else
  temp_root=$(mktemp -d "${TMPDIR:-/tmp}/$package-aur.XXXXXXXX") || die "cannot create a temporary directory"
  clone_path=$temp_root/$package
fi

note "publishing $package $version-$pkgrel from $source_dir"

if ! clone_repository; then
  die "cannot clone the AUR repository; check the network and your AUR SSH key"
fi

if cmp -s "$pkgbuild" "$clone_path/PKGBUILD" &&
  cmp -s "$srcinfo" "$clone_path/.SRCINFO"; then
  published_version=$(read_pkgbuild_value "$clone_path/PKGBUILD" pkgver)
  published_pkgrel=$(read_pkgbuild_value "$clone_path/PKGBUILD" pkgrel)
  note "the AUR already publishes $package $published_version-$published_pkgrel"
  note "the local PKGBUILD and .SRCINFO are already published; nothing to push"
  exit 0
fi

cp "$pkgbuild" "$clone_path/PKGBUILD"
cp "$srcinfo" "$clone_path/.SRCINFO"

if ! git -C "$clone_path" add PKGBUILD .SRCINFO; then
  die "cannot stage PKGBUILD and .SRCINFO in $clone_path"
fi

message="upgpkg: $package $version-$pkgrel"
if ! git -C "$clone_path" commit --quiet -m "$message"; then
  die "cannot commit '$message'; set git user.name and user.email"
fi
commit=$(git -C "$clone_path" rev-parse --short HEAD)
note "committed '$message' ($commit)"

if ! push_package; then
  if [ -n "$clone_dir" ]; then
    die "cannot push to the AUR; the commit remains in $clone_path"
  fi
  die "cannot push $package $version-$pkgrel to the AUR"
fi

note "pushed $package $version-$pkgrel to the AUR ($commit)"
if [ -n "$clone_dir" ]; then
  note "the clone was kept in $clone_path"
fi
note "the AUR package page and RPC endpoint can lag behind the Git repository for a few minutes"
