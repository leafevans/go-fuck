# go-fuck

[![CI](https://github.com/leafevans/go-fuck/actions/workflows/ci.yml/badge.svg)](https://github.com/leafevans/go-fuck/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/leafevans/go-fuck.svg)](https://pkg.go.dev/github.com/leafevans/go-fuck)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A blazingly fast reimplementation of [thefuck](https://github.com/nvbn/thefuck) in Go.

`go-fuck` fixes the command you just typed and got wrong. Run the `fuck` alias after a failure and
go-fuck replays the command to capture its error output, matches a rule against it and executes the
corrected command once you confirm the suggestion.

```console
$ git push
fatal: The current branch main has no upstream branch.
To push the current branch and set the remote as upstream, use

    git push --set-upstream origin main

$ fuck
git push --set-upstream origin main [enter/↑/↓/ctrl+c]
...
```

## Contents

1. [Why Go](#why-go)
2. [Features](#features)
3. [Requirements](#requirements)
4. [Installation](#installation)
5. [Shell setup](#shell-setup)
6. [First steps](#first-steps)
7. [How it works](#how-it-works)
8. [Settings](#settings)
9. [Confirmation keys](#confirmation-keys)
10. [Rules](#rules)
11. [Creating custom rules](#creating-custom-rules)
12. [Differences from thefuck](#differences-from-thefuck)
13. [Contributing](#contributing)
14. [License](#license)
15. [Acknowledgements](#acknowledgements)

## Why Go

- **One static binary.** A single `go-fuck` executable (or `go-fuck.exe` on Windows) contains
  everything. No interpreter, no packages, no virtualenv.
- **No runtime dependencies.** You do not need Python, pip or any other software installed to run the
  compiled binary.
- **Fast startup.** Fixing a command should be instantaneous. A native Go binary starts in
  milliseconds, without importing an interpreter and its modules.
- **Cross-platform by default.** The same source builds for Linux, macOS and Windows on amd64 and
  arm64, and the release workflow does exactly that.
- **Simple to contribute.** The code is standard library only, formatted with `gofmt` and tested on
  three operating systems in CI.

## Features

- Corrects the previous console command by matching its output against a set of rules.
- Shell integration for **bash**, **zsh**, **fish**, **tcsh** and **PowerShell**.
- Interactive confirmation prompt with keyboard navigation, or skip it with `-y`/`--yes`.
- Repeats automatically on failure with `-r`/`--repeat`, and can fix an arbitrary command with
  `--force-command`.
- A JSON settings file with defaults for every option, plus environment variable overrides using the
  `GOFUCK_*` prefix and the legacy `THEFUCK_*` prefix.
- Rule priorities, disabled rules and per-rule settings.
- No third-party Go dependencies: the whole project uses the standard library.

## Requirements

- **Go 1.22 or newer** to build from source.
- **No runtime dependencies.** The compiled binary is self-contained.
- A supported shell for the alias: bash, zsh, fish, tcsh or PowerShell. On Windows, PowerShell is
  recommended; a `cmd.exe` shim is provided as well.

## Installation

### With Homebrew (macOS and Linux)

The repository doubles as a Homebrew tap:

```bash
brew tap leafevans/go-fuck
brew install go-fuck
```

The formula installs the prebuilt binary for your platform, so no Go toolchain is
needed. Upgrade with `brew upgrade go-fuck`.

### With Scoop (Windows)

```powershell
scoop bucket add go-fuck https://github.com/leafevans/go-fuck
scoop install go-fuck
```

Upgrade with `scoop update go-fuck`.

### With winget (Windows)

The winget manifest in `packaging/winget/` is submitted to
[microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs). Once it is
accepted, install with:

```powershell
winget install leafevans.go-fuck
```

### With `go install`

```bash
go install github.com/leafevans/go-fuck/cmd/go-fuck@latest
```

The binary is installed into `$(go env GOPATH)/bin` (`%USERPROFILE%\go\bin` on Windows). Make
sure that directory is on your `PATH`.

### Prebuilt binaries

Download the archive for your operating system and architecture from the
[releases page](https://github.com/leafevans/go-fuck/releases), unpack it and place the `go-fuck`
binary somewhere on your `PATH`. Linux and macOS archives contain the raw binary, Windows archives
contain `go-fuck.exe`. Every release ships a `checksums.txt` file that you can use to verify the
download.

### From source

```bash
git clone https://github.com/leafevans/go-fuck.git
cd go-fuck
make build          # or: go build -o go-fuck ./cmd/go-fuck
```

`make install` runs `go install ./cmd/go-fuck` instead. On Windows the build produces
`go-fuck.exe`.

## Shell setup

go-fuck is a normal executable: the shell function is what connects it to your previous command.
`go-fuck --alias` prints the function definition for the current shell, so the setup is the same for
every shell and you always get the version matching your binary.

Add one of the following lines to your shell configuration file, then start a new shell or reload the
file.

go-fuck detects the shell from the `TF_SHELL` environment variable and otherwise from the process
tree. If detection is unreliable (for example when generating the alias from a setup script), set
`TF_SHELL` explicitly: `TF_SHELL=fish go-fuck --alias`.

### bash

Add to `~/.bashrc` (or `~/.bash_profile` on macOS):

```bash
eval "$(go-fuck --alias)"
```

### zsh

Add to `~/.zshrc`:

```zsh
eval "$(go-fuck --alias)"
```

### fish

Add to `~/.config/fish/config.fish`:

```fish
go-fuck --alias | source
```

### tcsh

Add to `~/.tcshrc`:

```tcsh
eval `go-fuck --alias`
```

### PowerShell

Add to your profile (the path is stored in `$PROFILE`):

```powershell
iex "$(go-fuck --alias)"
```

### Custom alias name

Pass the name after `--alias`:

```bash
eval "$(go-fuck --alias FUCK)"
```

The name has to be a valid shell function name. Without an argument, the alias is named after the
`TF_ALIAS` environment variable when it is set, and `fuck` otherwise.

### How the alias works

The generated function exports a few variables, calls go-fuck and evaluates the corrected command that
go-fuck prints on standard output:

| Variable | Set by | Purpose |
| --- | --- | --- |
| `TF_SHELL` | the alias | shell name (`bash`, `zsh`, `fish`, `tcsh`, `powershell`); go-fuck uses it instead of guessing from the process tree |
| `TF_ALIAS` | the alias | name of the alias function; used in messages and by the repeat mechanism |
| `TF_HISTORY` | the bash and zsh alias | the last ten history lines; go-fuck picks the failed command from them |
| `TF_SHELL_ALIASES` | the bash and zsh alias | the shell aliases, expanded before the command is matched |
| `GOFUCK_ARGUMENT_PLACEHOLDER` | the alias | separates go-fuck's own arguments from the arguments of the corrected command; the legacy `THEFUCK_ARGUMENT_PLACEHOLDER` is accepted too |

The fish alias passes the previous command explicitly via its `$history[1]` variable. Because
`TF_SHELL` is always set by the alias, go-fuck uses the right shell-specific behavior even when the
process tree is unreliable (containers, `sudo`, wrappers).

### Windows: cmd.exe and PowerShell shims

The `scripts/` directory contains two small Windows shims that mirror the scripts shipped with
thefuck:

- `scripts/fuck.bat` — for `cmd.exe`. It passes the previous command from the doskey history to
  go-fuck and runs the corrected command in the current console.
- `scripts/fuck.ps1` — for PowerShell. The first call writes the go-fuck initialization to your
  profile and defines `fuck()` for the current session; later calls delegate to the function.

Put the script somewhere on your `PATH` and call `fuck`.

## First steps

```console
$ cd project/src
cd: no such file or directory: project/src

$ fuck
mkdir -p project/src && cd project/src [enter/↑/↓/ctrl+c]
```

```console
$ git push
fatal: The current branch main has no upstream branch.
...

$ fuck
git push --set-upstream origin main [enter/↑/↓/ctrl+c]
```

```console
$ echo "hello'
zsh: unmatched "

$ fuck
echo "hello" [enter/↑/↓/ctrl+c]
```

Common options:

| Option | Meaning |
| --- | --- |
| `-y`, `--yes`, `--yeah`, `--hard` | execute the first suggestion without asking |
| `-r`, `--repeat` | if the corrected command fails, run go-fuck again |
| `--force-command CMD` | fix `CMD` instead of the previous command |
| `-d`, `--debug` | print debug information to stderr |
| `-v`, `--version` | print the go-fuck, Go and shell versions |
| `-h`, `--help` | print the full option list |
| `--enable-experimental-instant-mode` | accepted for compatibility; instant mode is not supported and go-fuck falls back to re-running the command |

## How it works

1. You run a command and it fails. You type `fuck`.
2. The shell function exports the shell name and history and calls `go-fuck`.
3. go-fuck determines the command that failed, expands shell aliases and **re-runs the command**
   with a short timeout (`wait_command`, three seconds by default; commands listed in
   `slow_commands` get `wait_slow_command`, fifteen seconds by default) to capture its output.
4. Every enabled rule is tried in priority order. A matching rule returns one or more corrected
   commands, which are sorted by priority and offered in the confirmation prompt.
5. When you accept a suggestion, go-fuck applies the rule's side effect (if any), optionally writes
   the corrected command to the shell history and prints it on standard output.
6. The shell function evaluates that output, so the corrected command runs in your shell exactly as
   if you had typed it.

The re-run is why go-fuck works with any shell and any command: it does not need a terminal logger or
shell plugin, at the cost of running the failed command a second time.

## Settings

Settings are stored in `settings.json` inside the go-fuck directory of the user configuration
directory returned by Go's `os.UserConfigDir()`:

| Operating system | Path |
| --- | --- |
| Linux / BSD | `~/.config/go-fuck/settings.json` (or `$XDG_CONFIG_HOME/go-fuck/settings.json`) |
| macOS | `~/Library/Application Support/go-fuck/settings.json` |
| Windows | `%AppData%\go-fuck\settings.json` |

The file is created with the default settings on first use. JSON does not support comments, so the
generated file contains an explanatory `_comment` field; it is ignored when the file is loaded.

A minimal file that customizes a few values:

```json
{
  "rules": ["All rules enabled"],
  "exclude_rules": ["git_push"],
  "wait_command": 10,
  "require_confirmation": true,
  "priority": {
    "cd_mkdir": 100
  },
  "settings": {}
}
```

All fields:

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `rules` | array of strings | `["All rules enabled"]` | Rules that are enabled. The sentinel value `"All rules enabled"` enables every rule that is enabled by default; list explicit rule names to enable exactly those rules. |
| `exclude_rules` | array of strings | `[]` | Rules that are never loaded, even when enabled. |
| `wait_command` | integer | `3` | Seconds to wait for the output of the re-run command. |
| `require_confirmation` | boolean | `true` | Ask before executing a corrected command. |
| `no_colors` | boolean | `false` | Disable colored output. |
| `debug` | boolean | `false` | Print debug information to stderr. |
| `priority` | object of integer | `{}` | Per-rule priority overrides. Lower values are tried first; the built-in default is `1000`. |
| `history_limit` | integer | `0` | Maximum number of shell history lines that are read. `0` means no limit. |
| `alter_history` | boolean | `true` | Write the corrected command back to the shell history. |
| `wait_slow_command` | integer | `15` | Seconds to wait for the output of commands listed in `slow_commands`. |
| `slow_commands` | array of strings | `["lein", "react-native", "gradle", "./gradlew", "vagrant"]` | Commands that are expected to be slow to re-run. |
| `repeat` | boolean | `false` | If the corrected command fails, run go-fuck again. |
| `instant_mode` | boolean | `false` | Reserved for the experimental instant mode. go-fuck does not support it and re-runs the command instead. |
| `num_close_matches` | integer | `3` | Maximum number of close matches suggested for misspelled commands. |
| `env` | object of string | `{"LC_ALL": "C", "LANG": "C", "GIT_TRACE": "1"}` | Environment variables added when the failed command is re-run. |
| `excluded_search_path_prefixes` | array of strings | `[]` | `PATH` prefixes that are skipped when searching for executables. |
| `settings` | object | `{} in the default file` | Rule-specific string settings, read by rules with `conf.Current().String(key, default)`. |

### Environment variables

Every setting can be set with an environment variable. The modern prefix is `GOFUCK_`; the legacy
`THEFUCK_` prefix from the original implementation is still recognized, so an existing thefuck
configuration keeps working. For example, `GOFUCK_WAIT_COMMAND=10` and
`THEFUCK_WAIT_COMMAND=10` do the same thing. If both prefixes are set for the same setting, the
legacy `THEFUCK_*` variable is applied last and wins.

| Suffix | Type | Example | Corresponding setting |
| --- | --- | --- | --- |
| `RULES` | colon separated list | `GOFUCK_RULES='sudo:no_command'` | `rules`; the token `DEFAULT_RULES` stands for `"All rules enabled"` |
| `EXCLUDE_RULES` | colon separated list | `GOFUCK_EXCLUDE_RULES='git_push'` | `exclude_rules` |
| `WAIT_COMMAND` | integer seconds | `GOFUCK_WAIT_COMMAND=10` | `wait_command` |
| `REQUIRE_CONFIRMATION` | boolean | `GOFUCK_REQUIRE_CONFIRMATION=false` | `require_confirmation` |
| `NO_COLORS` | boolean | `GOFUCK_NO_COLORS=true` | `no_colors` |
| `DEBUG` | boolean | `GOFUCK_DEBUG=true` | `debug` |
| `PRIORITY` | `rule=priority` list separated by colons | `GOFUCK_PRIORITY='cd_mkdir=100:git_push=900'` | `priority` |
| `HISTORY_LIMIT` | integer | `GOFUCK_HISTORY_LIMIT=2000` | `history_limit` |
| `ALTER_HISTORY` | boolean | `GOFUCK_ALTER_HISTORY=false` | `alter_history` |
| `WAIT_SLOW_COMMAND` | integer seconds | `GOFUCK_WAIT_SLOW_COMMAND=30` | `wait_slow_command` |
| `SLOW_COMMANDS` | colon separated list | `GOFUCK_SLOW_COMMANDS='lein:gradle'` | `slow_commands` |
| `REPEAT` | boolean | `GOFUCK_REPEAT=true` | `repeat` |
| `INSTANT_MODE` | boolean | `GOFUCK_INSTANT_MODE=false` | `instant_mode` |
| `NUM_CLOSE_MATCHES` | integer | `GOFUCK_NUM_CLOSE_MATCHES=5` | `num_close_matches` |
| `EXCLUDED_SEARCH_PATH_PREFIXES` | colon separated list | `GOFUCK_EXCLUDED_SEARCH_PATH_PREFIXES='/snap'` | `excluded_search_path_prefixes` |

The `env` map can only be set in `settings.json`. Booleans accept `true` and `false`
(case-insensitive).

Settings are applied in increasing order of precedence:

1. built-in defaults,
2. `settings.json`,
3. environment variables,
4. command line options such as `-y`, `-d` and `-r`.

## Confirmation keys

By default go-fuck asks before running the corrected command. The prompt shows the suggestion and the
available actions:

| Key | Action |
| --- | --- |
| `Enter` | execute the selected command |
| `↑`, `Ctrl+N`, `k`, `e` | select the previous suggestion |
| `↓`, `Ctrl+P`, `j`, `n` | select the next suggestion |
| `Ctrl+C`, `q` | abort without running anything |

Disable the prompt with `require_confirmation: false` in `settings.json` or by passing `-y`.

## Rules

Rules are compiled into the binary. Each rule decides whether it applies to a failed command and
returns one or more corrected commands. The list below is generated from the compiled rules:

<!-- RULES_LIST_START -->
- `adb_unknown_command`
- `ag_literal`
- `apt_get`
- `apt_get_search`
- `apt_invalid_operation`
- `apt_list_upgradable`
- `apt_upgrade`
- `aws_cli`
- `az_cli`
- `brew_cask_dependency`
- `brew_install`
- `brew_link`
- `brew_reinstall`
- `brew_uninstall`
- `brew_unknown_command`
- `brew_update_formula`
- `cargo`
- `cargo_no_command`
- `cat_dir`
- `cd_correction`
- `cd_cs`
- `cd_mkdir`
- `cd_parent`
- `chmod_x`
- `choco_install`
- `composer_not_command`
- `conda_mistype`
- `cp_create_destination`
- `cp_omitting_directory`
- `cpp11`
- `dirty_untar`
- `dirty_unzip`
- `django_south_ghost`
- `django_south_merge`
- `dnf_no_such_command`
- `docker_image_being_used_by_container`
- `docker_login`
- `docker_not_command`
- `dry`
- `fab_command_not_found`
- `fix_alt_space`
- `fix_file`
- `gem_unknown_command`
- `git_add`
- `git_add_force`
- `git_bisect_usage`
- `git_branch_0flag`
- `git_branch_delete`
- `git_branch_delete_checked_out`
- `git_branch_exists`
- `git_branch_list`
- `git_checkout`
- `git_clone_git_clone`
- `git_clone_missing`
- `git_commit_add`
- `git_commit_amend`
- `git_commit_reset`
- `git_diff_no_index`
- `git_diff_staged`
- `git_fix_stash`
- `git_flag_after_filename`
- `git_help_aliased`
- `git_hook_bypass`
- `git_lfs_mistype`
- `git_main_master`
- `git_merge`
- `git_merge_unrelated`
- `git_not_command`
- `git_pull`
- `git_pull_clone`
- `git_pull_uncommitted_changes`
- `git_push`
- `git_push_different_branch_names`
- `git_push_force`
- `git_push_pull`
- `git_push_without_commits`
- `git_rebase_merge_dir`
- `git_rebase_no_changes`
- `git_remote_delete`
- `git_remote_seturl_add`
- `git_rm_local_modifications`
- `git_rm_recursive`
- `git_rm_staged`
- `git_stash`
- `git_stash_pop`
- `git_tag_force`
- `git_two_dashes`
- `go_run`
- `go_unknown_command`
- `gradle_no_task`
- `gradle_wrapper`
- `grep_arguments_order`
- `grep_recursive`
- `grunt_task_not_found`
- `gulp_not_task`
- `has_exists_script`
- `heroku_multiple_apps`
- `heroku_not_command`
- `history`
- `hostscli`
- `ifconfig_device_not_found`
- `java`
- `javac`
- `lein_not_task`
- `ln_no_hard_link`
- `ln_s_order`
- `long_form_help`
- `ls_all`
- `ls_lah`
- `man`
- `man_no_space`
- `mercurial`
- `missing_space_before_subcommand`
- `mkdir_p`
- `mvn_no_command`
- `mvn_unknown_lifecycle_phase`
- `nixos_cmd_not_found`
- `no_command`
- `no_such_file`
- `npm_missing_script`
- `npm_run_script`
- `npm_wrong_command`
- `omnienv_no_such_command`
- `open`
- `pacman`
- `pacman_invalid_option`
- `pacman_not_found`
- `path_from_history`
- `php_s`
- `pip_install`
- `pip_unknown_command`
- `port_already_in_use`
- `prove_recursively`
- `python_command`
- `python_execute`
- `python_module_error`
- `quotation_marks`
- `rails_migrations_pending`
- `react_native_command_unrecognized`
- `remove_shell_prompt_literal`
- `remove_trailing_cedilla`
- `rm_dir`
- `rm_root`
- `scm_correction`
- `sed_unterminated_s`
- `sl_ls`
- `ssh_known_hosts`
- `sudo`
- `sudo_command_from_user_path`
- `switch_lang`
- `systemctl`
- `terraform_init`
- `terraform_no_command`
- `test.py`
- `tmux`
- `touch`
- `tsuru_login`
- `tsuru_not_command`
- `unknown_command`
- `unsudo`
- `vagrant_up`
- `whois`
- `workon_doesnt_exists`
- `wrong_hyphen_before_subcommand`
- `yarn_alias`
- `yarn_command_not_found`
- `yarn_command_replaced`
- `yarn_help`
- `yum_invalid_operation`
<!-- RULES_LIST_END -->

Regenerate it after adding or renaming a rule:

```bash
go run ./tools/gen-rule-list   # or: make rules-list
make rules-list-check          # fail when README.md is out of date
```

The same names can be used in the `rules`, `exclude_rules` and `priority` settings.

## Creating custom rules

Because the rules are compiled in, adding a rule means adding Go code and rebuilding go-fuck:

1. Read [PORTING.md](PORTING.md). It describes the rule interface, the helpers and the conventions.
2. Create `internal/rules/<rule_name>.go` and `internal/rules/<rule_name>_test.go`, port the test
   cases from thefuck and register the rule from an `init` function.
3. Run the checks described in [CONTRIBUTING.md](CONTRIBUTING.md).
4. Regenerate the rule list in this file with `make rules-list`.

There is no plugin mechanism and no third-party rule package format: thefuck's Python rules in
`~/.config/thefuck/rules` are not loaded, and the only rule set is the one compiled into the binary.
This keeps the binary self-contained and deterministic. Contributions to the port are welcome.

## Differences from thefuck

go-fuck is a reimplementation, not a drop-in replacement for every thefuck feature:

- **Rule coverage.** The rule set is being ported rule by rule. The list above is authoritative; the
  full thefuck rule list is larger. See [PORTING.md](PORTING.md) for the porting status.
- **Instant mode is not supported.** The experimental instant mode of thefuck reads the previous
  output from a terminal log. go-fuck always re-runs the command, using `wait_command` and
  `wait_slow_command` as timeouts. `--enable-experimental-instant-mode` is accepted for
  compatibility but prints a warning and uses the regular alias.
- **No shell logger.** thefuck's `-l`/`--shell-logger` option logs shell output through a socket;
  go-fuck does not implement it and warns when the option is used.
- **In-memory memoization instead of an on-disk cache.** go-fuck is a short-lived process, so
  rule results are memoized only for the duration of one invocation. Thefuck's `@cache` persisted
  values between runs.
- **Rules are compiled in.** There is no runtime rule loading, no third-party rule packages and no
  Python rule API. Custom rules require a rebuild.
- **No Python-apt database.** thefuck's `apt_get` rule uses the database installed by
  `python-commandnotfound`. go-fuck has no Python runtime, so the apt rules are ported without that
  database.
- **Settings use JSON.** Settings live in `settings.json` instead of `settings.py`, and the
  legacy `THEFUCK_*` environment variables are still understood alongside the new `GOFUCK_*` ones.

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) for the development setup, the
`make` targets and the conventions. In short:

```bash
go build ./...
go test ./... -count=1
go vet ./...
make fmt-check
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`,
`docs:`, `test:`, `chore:`, `refactor:`).

## License

MIT. See [LICENSE](LICENSE).

## Acknowledgements

- [thefuck](https://github.com/nvbn/thefuck) by [nvbn](https://github.com/nvbn) and its contributors,
  which this project reimplements. The rules, their behavior and many messages are ported from it.
- The original idea comes from a
  [tweet by @liamosaur](https://twitter.com/liamosaur/status/506975850596536320).

go-fuck is an independent project and is not affiliated with thefuck.