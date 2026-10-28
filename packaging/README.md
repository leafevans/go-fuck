# Distribution packaging

This directory (together with `Formula/` and `bucket/`) makes the repository a
multi-channel distribution for go-fuck.

| Channel | File | Install |
| --- | --- | --- |
| Homebrew (macOS/Linux) | `Formula/go-fuck.rb` | `brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck && brew install go-fuck` |
| Scoop (Windows) | `bucket/go-fuck.json` | `scoop bucket add go-fuck https://github.com/leafevans/go-fuck` then `scoop install go-fuck` |
| winget (Windows) | `packaging/winget/` | `winget install leafevans.go-fuck` (after the manifest is accepted upstream) |
| Prebuilt binaries | GitHub releases | download and put `go-fuck` on `PATH` |
| Go toolchain | `go install github.com/leafevans/go-fuck/cmd/go-fuck@latest` | |

## Homebrew

The repository itself is a Homebrew tap: users run

```bash
brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
brew install go-fuck
```

The formula installs the prebuilt release archive for the current platform and
architecture and links the `go-fuck` binary. It does not need the Go toolchain.

To check the formula on a machine with Homebrew:

```bash
brew style Formula/go-fuck.rb
brew audit --strict leafevans/go-fuck/go-fuck
brew install --formula Formula/go-fuck.rb   # local smoke test
```

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

## winget

The manifests in `packaging/winget/` follow the multi-file manifest layout:

- `leafevans.go-fuck.yaml` – version manifest
- `leafevans.go-fuck.installer.yaml` – zip installers for `x64` and `arm64`
- `leafevans.go-fuck.locale.en-US.yaml` – metadata

To publish them once a release exists:

1. Fork `microsoft/winget-pkgs`.
2. Copy the three files into `manifests/l/leafevans/go-fuck/<version>/`
   (rename them to `leafevans.go-fuck.yaml` etc. as they already are).
3. Validate and test locally:
   ```powershell
   winget validate --manifest packaging/winget
   winget install --manifest packaging/winget
   ```
4. Open a pull request against `microsoft/winget-pkgs`.

After the manifest is merged, users install it with
`winget install leafevans.go-fuck`.

## Release checklist

When publishing a new release `vX.Y.Z`:

1. Push the tag and wait for the Release workflow to finish.
2. Download `checksums.txt` from the release and regenerate the manifests:
   ```bash
   go run ./tools/update-packaging -version X.Y.Z -checksums checksums.txt
   ```
   The tool checks that the checksums are for the requested version and that
   all six release assets are present, then rewrites the Homebrew formula, the
   Scoop manifest and the winget manifests in one step. Review the diff and
   commit it (also possible through `make packaging VERSION=X.Y.Z
   CHECKSUMS=checksums.txt`).
3. Submit the updated winget manifests to `microsoft/winget-pkgs`.
4. Run `go run ./tools/gen-rule-list -check` and the full test suite before
   tagging the follow-up commit.

The version and checksums are intentionally committed as plain text so that the
release process stays reviewable in the repository history.