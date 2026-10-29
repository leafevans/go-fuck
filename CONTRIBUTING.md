# Contributing to go-fuck

Thanks for taking the time to contribute! This document explains how to set up
the project, which checks CI runs and the conventions used in the repository.

## Prerequisites

- [Go](https://go.dev/dl/) 1.22 or newer.
- [GNU make](https://www.gnu.org/software/make/) for the convenience targets. It is optional; every
  target documents the equivalent Go command below.
- [golangci-lint](https://golangci-lint.run/) v2 for `make lint` (CI installs it automatically).
- `git` and a GitHub account for opening pull requests.

The project uses the Go standard library only, so no module downloads are needed to build or test it.

## Getting started

```bash
git clone https://github.com/leafevans/go-fuck.git
cd go-fuck
go build ./...
go test ./... -count=1
```

## Make targets

| Target | Description |
| --- | --- |
| `make` / `make help` | print the list of targets |
| `make build` | build `go-fuck` (`go-fuck.exe` on Windows) |
| `make test` | run the test suite |
| `make test-race` | run the test suite with the race detector (needs a C compiler) |
| `make cover` | run the test suite with coverage and print the total |
| `make bench` | run the benchmarks (`go test ./internal/... -run '^$' -bench . -benchmem`) |
| `make vet` | run `go vet` |
| `make fmt` | format all Go files with `gofmt -s -w` |
| `make fmt-check` | fail when a Go file is not formatted |
| `make lint` | run `golangci-lint` with `.golangci.yml` |
| `make install` | run `go install ./cmd/go-fuck` |
| `make clean` | remove the binary and the GoReleaser output |
| `make packaging` | regenerate Homebrew/Scoop/winget manifests for `VERSION` from `CHECKSUMS` |
| `make packaging-check` | validate that the committed distribution manifests are consistent |
| `make rules-list` | regenerate the rule list in `README.md` |
| `make rules-list-check` | fail when the `README.md` rule list is out of date |

The recipes avoid POSIX-only commands so that they work on Windows as well: small helper programs in
`tools/` replace shell tricks such as `rm -rf` and `test -z "$(gofmt -l .)"`. If make is not
available, run the underlying commands:

```bash
go build ./...
go test ./... -count=1
go test ./internal/... -run '^$' -bench . -benchmem
go vet ./...
gofmt -s -w .
go run ./tools/checkfmt
go run ./tools/gen-rule-list
```

## Project layout

| Path | Contents |
| --- | --- |
| `cmd/go-fuck` | the command's `main` package |
| `internal/` | implementation packages (rules, shells, settings, UI, ...) |
| `internal/rules/` | one file and one test file per correction rule |
| `tools/` | small helper programs used by the Makefile and CI |
| `scripts/` | `cmd.exe` and PowerShell shims |
| `.github/workflows/` | CI and release workflows |
| `PORTING.md` | how the rules are ported from thefuck |
| `README.md` | user documentation; the rule list is generated |

## Code style

- **Formatting.** Run `gofmt -s -w .` (`make fmt`). CI fails when any file is not formatted.
  Comments and identifiers must be in English.
- **Vetting.** `go vet ./...` must be clean.
- **Linting.** `make lint` runs `golangci-lint` with the repository configuration; keep the
  build green.
- **Standard library only.** Do not add third-party module dependencies. `go.mod` intentionally has
  no `require` block. If something cannot be done with the standard library, open an issue first.
- **Errors.** Return errors instead of panicking. Wrap errors with `%w` so callers can inspect them
  with `errors.Is`/`errors.As`. Rules must not print to stdout; stdout is reserved for the
  corrected command that the shell alias evaluates. Use the helpers in `internal/logs` for user
  facing messages, which go to stderr.
- **Documentation.** Every package has a package comment, and exported identifiers are documented
  following the usual Go doc conventions.
- **Platform specifics.** Keep OS specific code in files with `_windows.go`/`_unix.go` suffixes
  and build tags instead of runtime `if runtime.GOOS == ...` checks where practical.
- **Tests.** Prefer table driven tests with `t.Run` subtests. Keep test fixtures identical to the
  Python tests where possible.

## Adding a rule

Rules that ship with go-fuck are compiled into the binary, so a new rule is a Go file with a test
file. Local, machine specific rules can be dropped as JSON files into the configuration directory
instead; see [Creating custom rules](README.md#creating-custom-rules). The workflow below is for
rules that ship with go-fuck. The complete guide, including the rule interface, helper functions and
the test helpers, is [PORTING.md](PORTING.md). The short version:

1. Read PORTING.md completely.
2. Add `internal/rules/<rule_name>.go` with the match function, the correction function and an
   `init` function that registers the rule. Use the rule name from the Python module.
3. Add `internal/rules/<rule_name>_test.go`, porting the test cases from
   `thefuck/tests/rules/test_<rule_name>.py`.
4. Run `go test ./internal/rules -run 'TestRule' -count=1`, `go build ./...` and
   `go vet ./...`.
5. Regenerate the rule list with `make rules-list` and commit the `README.md` change.
6. Do not edit files shared by several rules (`registry.go`, `testhelpers_test.go`) unless the
   change applies to every rule.

## Testing

- Run the full suite with `go test ./... -count=1`.
- Run only the rule tests with `go test ./internal/rules -run 'TestRule' -count=1`.
- Use `go test -race ./...` when a C compiler is available.
- Tests must pass on Linux, macOS and Windows; CI runs all three. Do not depend on the host shell,
  installed programs, network access or the user's configuration.
- Use `t.Setenv`, `t.TempDir` and the injectable function variables documented in PORTING.md
  instead of real state. Never read or write the real go-fuck settings or shell history.
- If a rule spawns processes, declare the command as a package level function variable so tests can
  replace it (see PORTING.md).

## Commit messages

The project uses [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<optional scope>): <subject>

<optional body>

<optional footer>
```

Allowed types:

- `feat` — a new user facing feature (for example a ported rule)
- `fix` — a bug fix
- `docs` — documentation only
- `test` — tests only
- `chore` — build, tooling and maintenance work
- `refactor` — a code change that neither fixes a bug nor adds a feature

Guidelines:

- Use the imperative mood in the subject line: "add git_push rule", not "added git_push rule".
- Keep the subject at 72 characters or less and do not end it with a period.
- Explain the **why** in the body when the change is not obvious. Wrap lines at 72 characters.
- Reference issues in the footer, for example `Fixes #123`.
- One logical change per commit.

Examples:

```text
feat(rules): port git_push from thefuck

fix(shells): keep TF_SHELL when the parent process is a wrapper

docs: document the GOFUCK_* environment variables

chore(ci): test on Go 1.22 and stable
```

## Pull request checklist

- [ ] `go build ./...` succeeds.
- [ ] `go test ./... -count=1` succeeds.
- [ ] `go vet ./...` is clean.
- [ ] `make fmt-check` (or `gofmt -s -l .`) reports nothing.
- [ ] `make lint` (or the CI lint job) is clean.
- [ ] New behavior is covered by tests; ported rules include the Python test cases.
- [ ] `make rules-list` was run when rules were added or renamed, and
      `make rules-list-check` passes.
- [ ] User facing changes are reflected in `README.md` and `PORTING.md` when relevant.
- [ ] Commits follow Conventional Commits and the branch is based on the latest `main`.
- [ ] CI is green on Linux, macOS and Windows.

## Reporting bugs

Open an issue with:

- the operating system and shell,
- the output of `go-fuck --version`,
- the command that failed and its output,
- what go-fuck suggested and what you expected instead,
- the output of `go-fuck --debug` when useful.

## License

By contributing you agree that your contributions are licensed under the MIT license used by this
project; see [LICENSE](LICENSE).