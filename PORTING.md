# Porting rules from thefuck

This document describes how the Python rules of [thefuck](https://github.com/nvbn/thefuck)
are ported to Go in this project. Read it completely before porting a rule, and
follow the conventions strictly so that all rules stay consistent.

## Layout

* One rule per file: `internal/rules/<rule_name>.go`.
* One test file per rule: `internal/rules/<rule_name>_test.go`.
* The rule name must be identical to the Python module name (for example
  `git_push`).
* Never edit shared files (`registry.go`, `testhelpers_test.go`,
  `README.md`, `go.mod`) unless the task explicitly asks for it.
* Never run `git` commands. The Team Lead commits.

## Rule anatomy

A rule consists of:

1. a `match` function that reports whether the rule applies,
2. a `get_new_command` function that returns one or more corrected commands,
3. an optional side effect that runs before the corrected command,
4. an `init` function that registers the rule.

```go
package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// lsAllMatch reports whether ls was called without -a.
func lsAllMatch(command *types.Command) bool {
	return strings.HasPrefix(command.Script, "ls ") &&
		!strings.Contains(command.Script, "-a")
}

// lsAllNewCommand adds the -a flag.
func lsAllNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " -a"}, nil
}

func init() {
	register(types.NewRule("ls_all", lsAllMatch, lsAllNewCommand))
}
```

Rule flags are set on the returned rule:

```go
func init() {
	rule := types.NewRule("brew_install", brewInstallMatch, brewInstallNewCommand)
	rule.EnabledByDefault = specific.BrewAvailable() // Python: enabled_by_default = brew_available
	rule.RequiresOutput = false                       // Python: requires_output = False
	rule.Priority = 900                               // Python: priority = 900
	rule.SideEffect = dirtyUntarSideEffect            // Python: def side_effect(...)
	register(rule)
}
```

Defaults applied by `types.NewRule`: `EnabledByDefault = true`,
`RequiresOutput = true`, `Priority = 1000`.

## API mapping

| Python (thefuck)                              | Go (go-fuck)                                              |
| --------------------------------------------- | --------------------------------------------------------- |
| `command.script`                             | `command.Script`                                          |
| `command.output`                             | `command.Output`                                          |
| `command.script_parts`                       | `command.ScriptParts()`                                   |
| `command.update(script=x)`                    | `command.WithScript(x)`                                   |
| `command.update(output=x)`                    | `command.WithOutput(x)`                                   |
| `settings.wait_command`                       | `conf.Current().WaitCommand`                              |
| `settings.get('key')`                         | `conf.Current().String("key", default)`                   |
| `shell.and_(a, b)`                            | `shells.Current().And(a, b)`                              |
| `shell.or_(a, b)`                             | `shells.Current().Or(a, b)`                               |
| `shell.quote(x)`                              | `shells.Current().Quote(x)`                               |
| `shell.split_command(x)`                      | `shells.Current().SplitCommand(x)`                        |
| `is_app(command, 'git')`                      | `utils.IsApp(command, 0, "git")`                          |
| `for_app('cd')` (match only)                  | `utils.ForApp([]string{"cd"}, 0, matchFn)`                |
| `for_app('cd', at_least=2)`                   | `utils.ForApp([]string{"cd"}, 2, matchFn)`                |
| `for_app(...)` on `get_new_command`           | `utils.ForAppNew([]string{"cd"}, 0, newCommandFn)`        |
| `@git_support` on match                       | `specific.GitSupportMatch(matchFn)`                       |
| `@git_support` on get_new_command             | `specific.GitSupportNew(newCommandFn)`                    |
| `@sudo_support` on match                      | `specific.SudoSupportMatch(matchFn)`                      |
| `@sudo_support` on get_new_command            | `specific.SudoSupportNew(newCommandFn)`                    |
| `which('brew')`                               | `utils.Which("brew") != ""`                               |
| `apt_available` / `brew_available` / ...      | `specific.AptAvailable()` / `specific.BrewAvailable()` / ... |
| `get_brew_path_prefix()`                      | `specific.GetBrewPathPrefix()`                            |
| `get_pkgfile(cmd)`, `archlinux_env()`        | `specific.GetPkgFile(cmd)`, `specific.ArchlinuxEnv()`    |
| `get_scripts()` (npm)                         | `specific.NpmGetScripts()`                                |
| `get_all_executables()`                       | `utils.GetAllExecutables()`                               |
| `get_valid_history_without_current(cmd)`      | `utils.GetValidHistoryWithoutCurrent(cmd)`                 |
| `get_alias()`                                 | `conf.Alias()`                                            |
| `get_all_matched_commands(out, 'Did you mean')` | `utils.GetAllMatchedCommands(out, "Did you mean")`       |
| `replace_argument(script, a, b)`              | `utils.ReplaceArgument(script, a, b)`                     |
| `replace_command(cmd, broken, matched)`       | `utils.ReplaceCommand(cmd, broken, matched)`               |
| `get_close_matches(word, options)`            | `utils.GetCloseMatches(word, options, 0, 0.6)`             |
| `get_closest(word, options)`                  | `utils.GetClosestDefault(word, options)`                  |
| `get_closest(..., fallback_to_first=False)`   | `utils.GetClosest(word, options, 0.6, false)`             |
| `memoize`                                    | `utils.Memoize(fn)` / `utils.Memoize1(fn)`               |
| `cache('~/.bashrc')(fn)`                      | `utils.Memoize(fn)` (see deviation note below)            |
| `@default_settings({...})`                    | `conf.Current().String("key", default)`                    |

Notes:

* `utils.GetCloseMatches` takes the cutoff as the last argument because Go has
  no default arguments. Python's default cutoff is `0.6`.
* `utils.ForApp` is generic: pass a match function returning `bool` or a
  correction function returning a single value. Multi-value results need
  `utils.ForAppNew`.
* Python's on-disk result cache (`@cache`) is replaced by in-process
  memoization. go-fuck is a short-lived process, so the cache only saves
  duplicate work inside one invocation. This is a documented deviation.
* `thefuck`'s `six`, `pathlib`, `urllib` and `shutil` usages map to the Go
  standard library (`os`, `net/url`, `os/exec`, `path/filepath`, ...).
* If a rule depends on a Python-only library, reimplement the behavior with an
  equivalent command or make the rule disabled by default. Add a comment that
  starts with `// Deviation:` explaining the difference.

## Subprocesses

Rules that call external programs must use `os/exec` and must be testable
without the program being installed. Declare the dependency as a package level
function variable and call it from the rule functions:

```go
// gradleTasks returns the tasks reported by gradle.
var gradleTasks = utils.Memoize(func() []string {
	output, err := exec.Command("gradle", "tasks").Output()
	if err != nil {
		return nil
	}
	// ... parse output
})

func init() {
	...
}
```

Tests overwrite the variable and restore it with `t.Cleanup`. Do not start
processes in `init` unless the Python rule evaluates availability at import
time (use `specific.*Available()` for that).

## Side effects

Python side effects receive `(old_cmd, corrected_script)` and may raise. In Go:

```go
func dirtyUntarSideEffect(oldCommand *types.Command, script string) error {
	// ... clean up files extracted by the first tar invocation
	return nil
}
```

Return an error instead of panicking. Never print to stdout from a side effect.

## Tests

Port the corresponding test file from `thefuck/tests/rules/test_<name>.py`.
Keep the fixture strings byte for byte where possible. Use table driven tests
and `t.Run` subtests.

Use the helpers from `internal/rules/testhelpers_test.go`:

* `resetTestState(t)` – call at the beginning of every test; resets settings,
  shell and caches.
* `testCommand(script, output)` – creates a command with a known output.
* `assertMatch(t, rule, script, output)`, `assertNotMatch(...)` – check
  `match`.
* `assertNewCommands(t, rule, script, output, want)` – checks the corrected
  commands returned by `get_new_command`.
* `fakeExecutables(t, "git", "brew")` – creates stubs in a temporary directory
  and prepends it to `PATH`.

Do not redefine those helper names. Rule specific helpers must be prefixed with
the rule name, for example `gitPushBitbucketOutput()`.

Test rules:

```go
func TestRuleGitPush(t *testing.T) {
	resetTestState(t)
	rule := Get("git_push")
	...
}
```

Run them with:

```
go build ./...
go test ./internal/rules -run 'TestRule' -count=1
go vet ./...
gofmt -l .
```

Tests must not depend on the machine: no real network, no installed shell
plugins, no user configuration. Use `t.Setenv`, `t.TempDir` and the injected
function variables instead.

## Definition of done

* The Go file exists and is registered with the exact Python rule name.
* `match` and `get_new_command` behave like the Python implementation for all
  cases covered by the Python tests.
* The Python test cases are ported and pass.
* `gofmt`, `go vet` and `go build ./...` are clean.
* All comments and identifiers are in English.
* The rule name appears in `README.md` (the README rule list is generated by
  the Team Lead; do not edit the README yourself).
