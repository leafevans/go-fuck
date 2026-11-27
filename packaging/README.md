# Distribution packaging

This directory (together with `Formula/` and `bucket/`) makes the repository a
multi-channel distribution for go-fuck.

| Channel                | File                                                            | Install                                                                                                                   |
| ---------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Homebrew (macOS/Linux) | `Formula/go-fuck.rb`                                            | `brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck && brew trust leafevans/go-fuck && brew install go-fuck` |
| Scoop (Windows)        | `bucket/go-fuck.json`                                           | `scoop bucket add go-fuck https://github.com/leafevans/go-fuck` then `scoop install go-fuck`                              |
| AUR (Arch Linux)       | `packaging/aur/`                                                | `yay -S go-fuck`, or clone `https://aur.archlinux.org/go-fuck.git` and run `makepkg -si`                                  |
| Prebuilt binaries      | GitHub releases                                                 | download and put `go-fuck` on `PATH`                                                                                      |
| Go toolchain           | `go install github.com/leafevans/go-fuck/v4/cmd/go-fuck@latest` |                                                                                                                           |

## Homebrew

The repository itself is a Homebrew tap: users run

```bash
brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
brew trust leafevans/go-fuck
brew install go-fuck
```

The formula installs the prebuilt release archive for the current platform and
architecture and links the `go-fuck` binary. It does not need the Go toolchain.

CI installs the formula on macOS with `brew tap` + `brew install` and runs
`go-fuck --version` as a smoke test. To check it locally:

```bash
brew style Formula/go-fuck.rb
brew audit --strict leafevans/go-fuck/go-fuck
brew install --formula Formula/go-fuck.rb   # local smoke test
```

The formula is Ruby and follows Homebrew's style: double quoted strings and no
frozen string literal comment (homebrew-core excludes `Formula/**` from that
cop). The repository `.rubocop.yml` configures editors accordingly, so Ruby LSP
does not apply the RuboCop defaults (single quotes, frozen literal comment) to
the formula; `brew style` passes its own configuration and is not affected.

## Scoop

The repository is also a Scoop bucket: users run

```powershell
scoop bucket add go-fuck https://github.com/leafevans/go-fuck
scoop install go-fuck
```

`bucket/go-fuck.json` installs `go-fuck.exe` from the release zip and exposes it
as `go-fuck`. `checkver`/`autoupdate` point at the GitHub releases and the
`checksums.txt` asset, so the manifest can be refreshed by a maintainer with
`scoop checkver` once the bucket is checked out.

## AUR

The Arch package lives in the AUR as
[`go-fuck`](https://aur.archlinux.org/packages/go-fuck). The canonical files
are `packaging/aur/PKGBUILD` and `packaging/aur/.SRCINFO`; the package builds
from the GitHub tag archive (`archive/refs/tags/vX.Y.Z.tar.gz`) with the Arch
Go toolchain and installs `/usr/bin/go-fuck`.

The package targets `x86_64`, the architecture of the official Arch Linux
repositories. On other architectures (for example Arch Linux ARM) use
`go install`, the prebuilt release binaries or build from source.

The `packaging` tests check that `pkgver`, `pkgrel` and the archive checksum
in `PKGBUILD` and `.SRCINFO` agree with each other and with the Homebrew
formula version. The **AUR verification** workflow can validate either the
published AUR recipe or the pending recipe in this repository. Pass the full
package version (for example `2.31.22-1`) and choose `repository` before
publishing or `published` afterward. It uses an `archlinux:base-devel`
container to compare `.SRCINFO`, build the package with `makepkg`, install it,
and verify both the package and CLI versions.

To update the package for a release:

1. Download the tag archive:

   ```bash
   curl -L -o go-fuck-X.Y.Z.tar.gz https://github.com/leafevans/go-fuck/archive/refs/tags/vX.Y.Z.tar.gz
   ```
2. Regenerate the package files. `tools/update-packaging` rewrites
   `PKGBUILD` and `.SRCINFO` when `-aur-source` is passed: it sets `pkgver`,
   resets `pkgrel` to 1 and computes the checksum of the archive.

   ```bash
   go run ./tools/update-packaging -version X.Y.Z \
     -checksums checksums.txt -aur-source go-fuck-X.Y.Z.tar.gz
   ```

   A packaging-only change (no new version) is not generated; bump `pkgrel`
   manually in both files instead.
3. Publish the two files to the AUR with `scripts/push-aur.sh` (see
   [Publishing to the AUR](#publishing-to-the-aur)):

   ```bash
   bash scripts/push-aur.sh --version X.Y.Z
   ```

### Publishing to the AUR

`scripts/push-aur.sh` is the preferred way to publish `packaging/aur/PKGBUILD`
and `packaging/aur/.SRCINFO`. It works on Linux, macOS, WSL and Git Bash on
Windows and needs bash, git and an SSH key registered with the AUR:

```bash
bash scripts/push-aur.sh --version X.Y.Z                # clone, commit and push
bash scripts/push-aur.sh --version X.Y.Z --identity-file ~/.ssh/nox_aur
bash scripts/push-aur.sh --version X.Y.Z --direct-ip    # blocked SSH fallback
bash scripts/push-aur.sh --version X.Y.Z --address IP   # DNS-over-HTTPS also blocked
bash scripts/push-aur.sh --dry-run                      # print the plan, no network access
```

The script reads `pkgver` from `PKGBUILD` when `--version` is omitted and
checks that `PKGBUILD` and `.SRCINFO` declare the same `pkgver` and
`pkgrel` and match the requested release. It clones
`ssh://aur@aur.archlinux.org/go-fuck.git` into a temporary directory and
exits successfully without pushing when both files are already published, so
it is safe to rerun. Otherwise it copies the files, commits
`upgpkg: go-fuck X.Y.Z-pkgrel`, pushes `HEAD:master` and removes the
temporary clone. The clone and the push are retried up to three times.
Pass `--identity-file` when the AUR key has a non-default filename; that
option uses `IdentitiesOnly=yes` so unrelated SSH identities are not offered.
`--clone-dir DIR` clones into `DIR` and keeps it. `--direct-ip` resolves the
current address again on every attempt; `--address IP` uses the given IPv4
address as is and skips the DNS-over-HTTPS lookup. The two options are
mutually exclusive.

The manual equivalent is:

```bash
git clone ssh://aur@aur.archlinux.org/go-fuck.git
cp packaging/aur/PKGBUILD packaging/aur/.SRCINFO go-fuck/
cd go-fuck
git commit -am "upgpkg: go-fuck X.Y.Z-pkgrel"
git push
```

### When SSH to aur.archlinux.org is blocked

Some networks resolve `aur.archlinux.org` to a proxy address that closes port 22 (and
443), so `git clone ssh://aur@aur.archlinux.org/...` fails with "Connection closed" or
"Connection reset". The AUR Git service is still reachable directly; resolve the current address over
DNS-over-HTTPS and keep the host key name with `HostKeyAlias`:

`scripts/push-aur.sh --direct-ip` automates this and resolves the address
again on every attempt; to do it by hand:

```bash
IP=$(curl -s -H 'accept: application/dns-json' \
  'https://cloudflare-dns.com/dns-query?name=aur.archlinux.org&type=A' |
  jq -r '.Answer[] | select(.type == 1) | .data' | head -n1)
export GIT_SSH_COMMAND="ssh -o HostKeyAlias=aur.archlinux.org"
git clone "ssh://aur@$IP/go-fuck.git"
# ...copy PKGBUILD and .SRCINFO, commit, then:
git push
```

Some networks block the DNS-over-HTTPS endpoint as well. In that case pass the
AUR IPv4 address explicitly with `--address`:

```bash
bash scripts/push-aur.sh --version X.Y.Z --address 209.126.35.78
```

`--address IP` skips the DNS lookup, connects to the given IPv4 address for the
clone and for every push retry, and keeps the same
`HostKeyAlias=aur.archlinux.org` SSH option, so the host key check still works.
It cannot be combined with `--direct-ip`.

The address can change, so resolve it again on every push instead of hard-coding it; `--direct-ip`
retries the resolution up to three times when the connection fails. With `--address` re-check the
address when connections start failing, because the script uses it as given. The AUR RPC endpoint
(`https://aur.archlinux.org/rpc/?v=5&type=info&arg[]=go-fuck`) works over HTTPS and reports the
published version; it can lag behind the Git repository for a few minutes.

## Local updater

`scripts/update-go-fuck.ps1` (Windows, and WSL unless `-SkipWsl` is passed) and
`scripts/update-go-fuck.sh` (Linux and macOS) download a release, verify the
SHA256 checksums and replace a locally installed binary:

```powershell
./scripts/update-go-fuck.ps1                 # latest release, Windows and WSL
./scripts/update-go-fuck.ps1 -Version 3.0.1 # a specific release
./scripts/update-go-fuck.ps1 -SkipWsl
```

```bash
./scripts/update-go-fuck.sh                   # latest release
./scripts/update-go-fuck.sh 3.0.1 ~/.local/bin
```

The WSL part of the PowerShell script runs `update-go-fuck.sh`, so keep the two
files next to each other. It executes the script with `sh` rather than the
user's login shell. When WSL cannot reach GitHub, the Windows side downloads and
verifies the Linux archive before WSL verifies and installs it atomically. The
updater is a convenience for local installs; the Homebrew, Scoop and AUR
packages update through their own package managers.

## Release checklist

When publishing a new release `vX.Y.Z`:

1. Push the tag and wait for the Release workflow to finish.
2. Download `checksums.txt` and the tag archive from the release, then
   regenerate the manifests and the AUR package:

   ```bash
   go run ./tools/update-packaging -version X.Y.Z \
     -checksums checksums.txt -aur-source go-fuck-X.Y.Z.tar.gz
   ```

   The tool checks that the checksums are for the requested version and that
   all six release assets are present, then rewrites the Homebrew formula, the
   Scoop manifest and (with `-aur-source`) the AUR `PKGBUILD` and
   `.SRCINFO` with a fresh `pkgrel=1`, in one step. Review the diff and
   commit it (also possible through `task packaging VERSION=X.Y.Z
CHECKSUMS=checksums.txt AUR_SOURCE=go-fuck-X.Y.Z.tar.gz`). The whole download
   and regeneration step is wrapped by
   `scripts/update-manifests.ps1 -Version X.Y.Z`; pass `-Commit` to commit the
   regenerated files. `-PushAur` publishes the AUR package in the same step
   through Git Bash or WSL, `-AurDryRun` only prints its publish plan, and
   `-AurAddress IP` passes an explicit address when the DNS-over-HTTPS
   lookup is blocked.
3. Build and smoke-test the pending AUR recipe, publish it, then verify the
   published recipe and the public RPC version. See the independently
   retryable steps below.
4. Run `go run ./tools/gen-rule-list -check` and the full test suite on the
   follow-up commit. The existing release tag does not move.
5. Verify the tagged Go install and, for a new latest stable release, `@latest`.

The version and checksums are intentionally committed as plain text so that the
release process stays reviewable in the repository history.

## Retryable release verification

Replace `X.Y.Z` with the release version and `X.Y.Z-1` with the full AUR
package version (use the actual `pkgrel` for a packaging-only update). Retry
only the failed step; do not recreate the GitHub tag or regenerate manifests
just to rerun a check.

### Push the AUR recipe

From the repository root with the reviewed recipe files for this release and
an authorized AUR SSH key, preview the plan first:

```bash
bash scripts/push-aur.sh --version X.Y.Z --dry-run
```

The dry run checks the local version fields but does not test SSH or publish.
When ready to publish, run `bash scripts/push-aur.sh --version X.Y.Z` on its own
(no need to rerun `update-manifests.ps1`). Retry that same command after fixing
an SSH or network failure; if both files already match the remote, the script
exits successfully without pushing. See [Publishing to the AUR](#publishing-to-the-aur)
for key and connection options. Do not retry an old recipe over a newer release.

### Build and install with makepkg

Dispatch the [AUR verification workflow](../.github/workflows/aur-verify.yml)
with `version=X.Y.Z-1`. Use `source=repository` on the branch containing the
updated recipe before publishing; use `source=published` to build the public AUR
Git checkout after publishing. For example, from the GitHub repository checkout:

```bash
gh workflow run aur-verify.yml --ref RECIPE_BRANCH -f source=repository -f version=X.Y.Z-1
gh workflow run aur-verify.yml -f source=published -f version=X.Y.Z-1
```

Check the workflow result, not just the dispatch. It compares `.SRCINFO`,
builds and installs the package, and checks that `pacman -Q go-fuck` reports
`go-fuck X.Y.Z-1` and `/usr/bin/go-fuck --version` starts with
`go-fuck X.Y.Z using Go `. If the build fails, use the job log to distinguish
a recipe/checksum error from a dependency or network failure, then dispatch
again with the same `source` and package version on the intended ref.

To reproduce locally, start in a fresh directory with both recipe files from
`packaging/aur/` or a fresh HTTPS clone of the published AUR repository. Use a
disposable Arch Linux x86_64 environment with `base-devel`; run as a normal
user with permission to install packages via `sudo`:

```bash
(
  set -euo pipefail
  makepkg --printsrcinfo | diff -u .SRCINFO -
  makepkg --syncdeps --install --clean
  test "$(pacman -Q go-fuck)" = "go-fuck X.Y.Z-1"
  [[ $(/usr/bin/go-fuck --version) == "go-fuck X.Y.Z using Go "* ]]
)
```

### Check the AUR RPC version

After publishing, use this read-only HTTPS check (requires `curl` and `jq`):

```bash
(
  set -o pipefail
  curl -fsS --get https://aur.archlinux.org/rpc/ \
    --data-urlencode v=5 --data-urlencode type=info --data-urlencode "arg[]=go-fuck" |
    jq -e --arg expected X.Y.Z-1 \
      '.resultcount == 1 and .results[0].Name == "go-fuck" and .results[0].Version == $expected'
)
```

Expect `true` and exit status zero. If AUR Git is current but RPC still shows
an older version, wait a few minutes for indexing and retry just the RPC
check. RPC confirms public metadata, not that the recipe builds.

### Check Go install

Use a Go toolchain supported by the tag. Run this Bash check on Linux or macOS
with a fresh `GOBIN` so an older binary on `PATH` cannot pass. It checks both the
embedded module version and the CLI output:

```bash
(
  set -euo pipefail
  pkg=github.com/leafevans/go-fuck/v4/cmd/go-fuck
  want=vX.Y.Z
  ref="$want"  # change to latest for the separate @latest check
  bin=$(mktemp -d)
  trap 'rm -rf -- "$bin"' EXIT
  GOBIN=$(cd "$bin" && pwd -P)
  export GOBIN
  go install "$pkg@$ref"
  go version -m "$GOBIN/go-fuck" | awk -v want="$want" \
    '$1 == "mod" && $2 == "github.com/leafevans/go-fuck/v4" && $3 == want { ok = 1 } END { exit !ok }'
  [[ $("$GOBIN/go-fuck" --version) == "go-fuck X.Y.Z using Go "* ]]
)
```

Once the pinned tag passes, repeat with `ref=latest` (leave `want` unchanged)
if the tag should be the latest version on the `/v4` module path. If the pinned
install works but `@latest` still resolves an older version, check your
`GOPROXY` and retry only `@latest` after the proxy has refreshed. Neither check
depends on AUR publication.
