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
- **Instant mode.** Instant mode is opt-in and documented in [README.md](README.md#instant-mode).
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
- The instant mode is unchanged; it still needs the `--enable-experimental-instant-mode` flag
  (the flag name is historical).

Upgrade with the channel you installed go-fuck from:

| Channel           | Command                                                      |
| ----------------- | ------------------------------------------------------------ |
| Homebrew          | `brew upgrade go-fuck`                                       |
| Scoop             | `scoop update go-fuck`                                       |
| AUR               | `yay -Syu` or `paru -Syu`                                    |
| Go toolchain      | `go install github.com/leafevans/go-fuck/cmd/go-fuck@latest` |
| Prebuilt archives | download the new archive and replace the binary              |

`go-fuck --setup` is idempotent; it does not have to be run again after upgrading.

## Upgrading from 1.x to 2.0.0

The optional AI fallback was removed. The built-in rules are the only correction source again: the
fallback required sending the failed command and its output to a third-party endpoint, and the
extra round trip added latency to exactly the cases where no rule matched.

- The settings keys `ai`, `ai_endpoint`, `ai_model` and `ai_timeout` are ignored. Existing
  `settings.json` files stay valid; the keys can be deleted.
- The environment variables `GOFUCK_AI`, `GOFUCK_AI_ENDPOINT`, `GOFUCK_AI_MODEL`,
  `GOFUCK_AI_TIMEOUT`, `GOFUCK_AI_API_KEY` and `OPENAI_API_KEY` are no longer read.
- No replacement is needed: when no rule matches, go-fuck reports "No fucks given" as before.
- The `--json` contract, the command line flags and the shell setup are unchanged (the
  `schema_version` stays `1`), so no migration is required.

## Upgrading from 2.1 to 2.2.0

No migration is required. Instant mode is now documented as an opt-in feature instead of an
experimental one; the `--enable-experimental-instant-mode` flag name is kept for compatibility and
the generated shell initialization lines are unchanged. This release also adds end-to-end tests for
the terminal logger on Linux, macOS and Windows and makes `go-fuck --doctor` validate the active
instant mode log.

## Upgrading from 2.13 to 2.14.0

go-fuck now uses `encoding/json/v2` for every JSON document and therefore needs **Go 1.27 or newer**
to build from source. The prebuilt release archives, Homebrew, Scoop and AUR packages are unaffected,
and `go install github.com/leafevans/go-fuck/cmd/go-fuck@latest` keeps working because
`GOTOOLCHAIN=auto` (the default) downloads Go 1.27 when the installed toolchain is older. Only
builds with `GOTOOLCHAIN=local` and an older Go need an upgrade.

The settings and `--json` contracts keep their keys, meaning and `schema_version: 1`, but
`settings.json` is parsed with the stricter v2 rules:

- **Member names are case-sensitive.** A hand-written key such as `"Rules"` used to match `rules`
  because v1 compared names case-insensitively; it is now treated as an unknown key and ignored.
  Files written by `go-fuck` use the documented lower-case names and are unaffected.
- **Duplicate keys are rejected.** A file with the same key twice is reported and not applied; v1
  silently kept the last value. Remove the duplicate to upgrade.
- **Invalid UTF-8 is rejected** instead of being replaced with U+FFFD.
- Unknown keys are still ignored, and the `--json` document keeps its fields and layout. The
  encoder no longer escapes `<`, `>` and `&` as `\u003c`, `\u003e` and `\u0026` inside strings;
  both forms are valid JSON and decode to the same values. Invalid UTF-8 bytes in captured command
  output are replaced with U+FFFD before encoding, so `--json` always emits a valid document.

## Future releases

- **Patch** releases contain bug fixes and documentation changes.
- **Minor** releases add compatible features, rules and optional JSON fields.
- **Major** releases may change documented behavior, settings or the JSON contract in a breaking
  way; this file gets an upgrade section for them.
