# Upgrading go-fuck

This file documents the compatibility promises of go-fuck and what changes between release series.
For the changes of every individual release see the [GitHub releases](https://github.com/leafevans/go-fuck/releases).

## Compatibility promise

Starting with **1.0.0**:

- **Settings.** The keys in `settings.json` keep their name, type and meaning. Unknown keys are
  ignored, so a settings file written by a newer go-fuck still loads in an older version. Invalid
  JSON or a value of the wrong type is reported and the file is not applied. New optional keys may be
  added in minor releases.
- **Environment variables.** The `GOFUCK_*` prefix and the legacy `THEFUCK_*` prefix keep working
  as before. The precedence stays: built-in defaults < `settings.json` < environment < command line
  flags.
- **JSON output.** `--json` is a stable machine readable contract. `schema_version` is `1` for
  the whole 1.x series. The fields `command`, `output`, `output_truncated` and
  `suggestions[].command`, `suggestions[].priority`, `suggestions[].side_effect` keep their
  meaning. New optional fields may be added in minor releases; consumers must ignore unknown fields.
  The document is written to stdout and diagnostics stay on stderr.
- **Command line.** The flags and behavior documented in the README keep working. A breaking change
  requires a major version.
- **Instant mode.** The experimental instant mode stays opt-in and may change during the 1.x series.
- **Versioning.** Releases follow [Semantic Versioning](https://semver.org/); the policy is described
  in [CONTRIBUTING.md](CONTRIBUTING.md#release-and-versioning).

## Upgrading from 0.x to 1.0.0

No configuration or JSON migration is required:

- `settings.json` stays valid; no keys were removed or renamed.
- `--json` behaves like 0.12.0, including the `schema_version: 1` field that was introduced in
  0.12.0.
- The command line flags and the shell setup are unchanged.
- The release version changes from `0.x.y` to `1.0.0`; `go-fuck --version` prints the new
  version.
- The experimental instant mode is unchanged; it still needs the
  `--enable-experimental-instant-mode` flag.

Upgrade with the channel you installed go-fuck from:

| Channel           | Command                                                      |
| ----------------- | ------------------------------------------------------------ |
| Homebrew          | `brew upgrade go-fuck`                                       |
| Scoop             | `scoop update go-fuck`                                       |
| AUR               | `yay -Syu` or `paru -Syu`                                    |
| Go toolchain      | `go install github.com/leafevans/go-fuck/cmd/go-fuck@latest` |
| Prebuilt archives | download the new archive and replace the binary              |

`go-fuck --setup` is idempotent; it does not have to be run again after upgrading.

## Future releases

- **Patch** releases contain bug fixes and documentation changes.
- **Minor** releases add compatible features, rules and optional JSON fields.
- **Major** releases may change documented behavior, settings or the JSON contract in a breaking
  way; this file gets an upgrade section for them.
