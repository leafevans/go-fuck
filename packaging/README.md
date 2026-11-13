# Distribution packaging

This directory (together with `Formula/` and `bucket/`) makes the repository a
multi-channel distribution for go-fuck.

| Channel                | File                                                         | Install                                                                                                                   |
| ---------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| Homebrew (macOS/Linux) | `Formula/go-fuck.rb`                                         | `brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck && brew trust leafevans/go-fuck && brew install go-fuck` |
| Scoop (Windows)        | `bucket/go-fuck.json`                                        | `scoop bucket add go-fuck https://github.com/leafevans/go-fuck` then `scoop install go-fuck`                              |
| AUR (Arch Linux)       | `packaging/aur/`                                             | `yay -S go-fuck`, or clone `https://aur.archlinux.org/go-fuck.git` and run `makepkg -si`                                  |
| Prebuilt binaries      | GitHub releases                                              | download and put `go-fuck` on `PATH`                                                                                      |
| Go toolchain           | `go install github.com/leafevans/go-fuck/v2/cmd/go-fuck@latest` |                                                                                                                           |

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

There is no CI job for the AUR because Arch tooling is not available on the
GitHub runners. The `packaging` tests check that `pkgver`, `pkgrel` and the
archive checksum in `PKGBUILD` and `.SRCINFO` agree with each other and with
the Homebrew formula version.

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
./scripts/update-go-fuck.ps1 -Version 2.11.0 # a specific release
./scripts/update-go-fuck.ps1 -SkipWsl
```

```bash
./scripts/update-go-fuck.sh                   # latest release
./scripts/update-go-fuck.sh 2.11.0 ~/.local/bin
```

The WSL part of the PowerShell script runs `update-go-fuck.sh`, so keep the two
files next to each other. The updater is a convenience for local installs; the
Homebrew, Scoop and AUR packages update through their own package managers.

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
    commit it (also possible through `make packaging VERSION=X.Y.Z
CHECKSUMS=checksums.txt AUR_SOURCE=go-fuck-X.Y.Z.tar.gz`). The whole download
    and regeneration step is wrapped by
    `scripts/update-manifests.ps1 -Version X.Y.Z`; pass `-Commit` to commit the
    regenerated files.
3. Publish the regenerated `PKGBUILD` and `.SRCINFO` to the AUR:
    ```bash
    bash scripts/push-aur.sh --version X.Y.Z
    ```
    The script skips the push when the AUR already carries the same files, so
    it is safe to rerun. Add `--direct-ip` when SSH to `aur.archlinux.org` is
    blocked, or `--address IP` when the DNS-over-HTTPS endpoint is blocked too
    (see [Publishing to the AUR](#publishing-to-the-aur)).
4. Run `go run ./tools/gen-rule-list -check` and the full test suite before
   tagging the follow-up commit.
5. Check that `go install github.com/leafevans/go-fuck/v2/cmd/go-fuck@latest`
   resolves the new tag; the major version is part of the Go module path.

The version and checksums are intentionally committed as plain text so that the
release process stays reviewable in the repository history.
