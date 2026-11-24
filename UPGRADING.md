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
  the whole 1.x, 2.x and 3.x series. The fields `command`, `output`, `output_truncated` and
  `suggestions[].command`, `suggestions[].priority`, `suggestions[].side_effect` keep their
  meaning. New optional fields may be added in minor releases; consumers must ignore unknown fields.
  The document is written to stdout and diagnostics stay on stderr.
- **Command line.** The flags and behavior documented in the README keep working. A breaking change
  requires a major version.
- **Output capture.** `capture_mode` selects `auto`, `instant` or `rerun`, documented in
  [README.md](README.md#terminal-capture-and-instant-mode).
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

| Channel           | Command                                                         |
| ----------------- | --------------------------------------------------------------- |
| Homebrew          | `brew upgrade go-fuck`                                          |
| Scoop             | `scoop update go-fuck`                                          |
| AUR               | `yay -Syu` or `paru -Syu`                                       |
| Go toolchain      | `go install github.com/leafevans/go-fuck/v2/cmd/go-fuck@latest` |
| Prebuilt archives | download the new archive and replace the binary                 |

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
and `go install github.com/leafevans/go-fuck/v2/cmd/go-fuck@latest` keeps working because
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

## Upgrading from 2.20 to 2.21.0

The Go module moved to `github.com/leafevans/go-fuck/v2` so that `go install` resolves the current
release series. No settings, command line or `--json` migration is required: only the import path of
the Go packages and the install command change.

- Install with `go install github.com/leafevans/go-fuck/v2/cmd/go-fuck@latest`. The command without
  the `/v2` suffix keeps resolving to the old 1.x series, because a Go module path has to carry the
  major version from 2 on.
- Homebrew, Scoop, the AUR package and the prebuilt archives are unaffected; nothing has to be
  reinstalled beyond the usual upgrade command.

## Upgrading from 2.28 to 2.29.0

No configuration or JSON migration is required. The shell aliases pass the information commands
(`--list-rules`, `--doctor`, `--json`, `--alias`, `--help`, `--version`) through to go-fuck
instead of evaluating their output as a corrected command, so `fuck --list-rules` now prints the rule
list. Aliases generated by 2.28 and older keep working for corrections, but `fuck --list-rules` executes
the rule names in such a shell; old aliases are replaced by the next shell start because the
configuration file evaluates `go-fuck --alias`. To update a running shell immediately, evaluate the alias
again (`eval "$(go-fuck --alias)"` in bash and zsh, `iex "$(go-fuck --alias)"` in PowerShell) or run
`go-fuck --setup`. go-fuck warns when it detects an alias of an older format.

## Upgrading from 2.35.x to 2.36.0

No settings, shell initialization or JSON migration is required. Version 2.36.0 standardizes the
human-readable command interface and output streams:

- `--help` and `--version` now write their normal information to stdout instead of stderr.
- `--setup` success and reload messages now write to stdout; warnings and failures stay on stderr.
- Human-readable reports use one `go-fuck <command>` heading and `OK`/`INFO`/`WARN`/`FAIL` status
  vocabulary. Styling is limited to interactive terminals; pipes, `NO_COLOR` and `no_colors` remain
  plain.
- `--alias` remains raw shell code on stdout, and `--json` plus non-terminal `--list-rules` output
  remain unchanged machine-readable contracts.

Existing shell aliases already pass `--help` and `--version` through to go-fuck, so they do not need
to be regenerated for this change.

## Upgrading from 2.36 to 3.0.0

Version 3.0 changes the default output-capture policy. No JSON migration is required: the schema
stays at version 1 and `capture_source` is an optional field.

- The Go module moved to `github.com/leafevans/go-fuck/v3`. Build from source with
  `go install github.com/leafevans/go-fuck/v3/cmd/go-fuck@latest`; the `/v2` path continues to
  resolve the older 2.x series.
- New installations use `capture_mode: "auto"`. Supported Bash, Zsh and PowerShell setups start a
  logged shell and read the previous command output without re-running it.
- Run `go-fuck --setup` to migrate an existing initialization line. Setup preserves an explicitly
  configured `capture_mode`, otherwise it migrates the shell to `auto`.
- `capture_mode: "instant"` never falls back to re-running. `capture_mode: "rerun"` preserves the
  old behavior explicitly.
- The deprecated `instant_mode` key and `GOFUCK_INSTANT_MODE` variable are still accepted.
  `instant_mode: true` maps to `auto`; `instant_mode: false` maps to `rerun` when no explicit
  `capture_mode` is present.
- `--enable-experimental-instant-mode` remains as a compatibility alias for
  `--capture-mode=instant`.
- Generated aliases use format version 4. Reload or regenerate the alias to skip consecutive alias
  calls when locating the previous real command.

## Unreleased safety changes (version not assigned)

This section describes changes on main, not an available release. The last published
version is still 3.0.2. No settings keys or JSON fields have been removed.

- History-rewriting rules `git_commit_reset` and `git_commit_amend`, and process-termination
  rule `port_already_in_use`, are disabled by default. To retain a legacy rule, add its
  name to your existing `rules` array; preserve other entries and the `DEFAULT_RULES`
  token if present. Do not enable these rules unless you accept their risks.
- `git_hook_bypass` requires a local hook failure diagnostic, not just a hook/tool
  mention. Server-side hook rejection cannot be bypassed using local `--no-verify`.
- Existing-branch corrections reuse the branch instead of deleting and recreating it.
- Missing cp sources, incidental mkdir/rm arguments and SSH authentication failures
  no longer trigger unrelated directory-creation, recursive-removal or sudo suggestions.
- TCP owner lookup excludes clients, invalid PIDs and ambiguous multiple owners;
  PID reuse between suggestion and execution remains possible even after opt-in.
- `schema_version` remains 1. `side_effect` continues to mean an additional pre-run
  callback, not a general guarantee that the suggested command is non-destructive.
  Rule-list default flags and hook output requirements change; clients must not cache
  these values as fixed across releases.

Release gate: the version is intentionally unassigned. The published policy requires a
major release for breaking documented behavior. Decide whether these safety corrections
fall under bug fixes or require a major migration before changing version constants,
module paths, tags or distribution manifests. Validate preview/explain/JSON behavior,
cross-platform CI and Arch packaging before publishing; regenerate manifests only
from the actual release assets and their checksums.

## Future releases

- **Patch** releases contain bug fixes and documentation changes.
- **Minor** releases add compatible features, rules and optional JSON fields.
- **Major** releases may change documented behavior, settings or the JSON contract in a breaking
  way; this file gets an upgrade section for them.
