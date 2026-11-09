# Security policy

## Supported versions

Only the latest release receives security fixes. The `main` branch contains the
next release and is supported on a best effort basis.

## Reporting a vulnerability

Please **do not open a public issue** for a security problem.

Report it privately through GitHub's private vulnerability reporting:

<https://github.com/leafevans/go-fuck/security/advisories/new>

Include the go-fuck version (`go-fuck --version`), the operating system, the
shell and the steps that reproduce the problem. If the report involves command
output or terminal logs, remove passwords, tokens and other sensitive data.

You can expect an initial response within a few days. Please give me a chance to
release a fix before you disclose the problem publicly.

## What is in scope

go-fuck never executes a suggestion itself: it prints the corrected command and
the shell alias evaluates it. Security relevant issues are therefore mainly
about what go-fuck reads, what it suggests and how it is built:

- a rule that suggests a destructive command for a failure that does not need it
  (for example a `git reset --hard` or a recursive removal),
- a rule that suggests a command containing attacker controlled text from a
  command output,
- leaked secrets or terminal content in debug output, logs or the instant mode
  terminal log,
- command execution or injection problems while re-running a failed command, and
- problems with the release artifacts, the checksums or the generated shell
  aliases.

Incorrect or missing suggestions are not security issues; report them with the
[bug report](https://github.com/leafevans/go-fuck/issues/new?template=bug_report.yml)
template instead.

## How go-fuck handles data

- The default mode re-runs the failed command to capture its output and merges
  the configured `env` overrides (see the README) into its environment.
- Instant mode reads the output from the terminal log written by the shell
  logger instead of re-running the command.
- Debug output (`--debug`) and the rule failure messages go to standard error and
  may contain the captured output.
- The binary is built from the tagged source by GoReleaser with the Go standard
  library only. The release publishes checksums, and the distribution manifests
  (Homebrew, Scoop, AUR) pin them.
