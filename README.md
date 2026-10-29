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
8. [Instant mode](#instant-mode)
9. [Machine readable output](#machine-readable-output)
10. [Settings](#settings)
11. [Confirmation keys](#confirmation-keys)
12. [Rules](#rules)
13. [AI fallback](#ai-fallback-optional)
14. [Creating custom rules](#creating-custom-rules)
15. [Differences from thefuck](#differences-from-thefuck)
16. [Contributing](#contributing)
17. [License](#license)
18. [Acknowledgements](#acknowledgements)

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

### Benchmarks

The correction hot path and the shell history parser have Go benchmarks
(`make bench`); they run on every supported platform and do not depend on the
installed tools. Measured on an AMD Ryzen 7 7730U (16 logical CPUs) with Go 1.26:

| Benchmark | Result |
| --- | --- |
| Full correction path (select rules, match them, order suggestions) | ~76 µs/op |
| Matching every rule against a captured output | ~65 µs/op |
| Matching rules when the command output could not be captured | ~0.5 µs/op |
| Parsing a 20,000 line shell history file | ~0.67 ms/op |

The rule engine adds tens of microseconds; the wall clock time of a correction
is dominated by deliberately re-running the failed command. An interpreter based
implementation pays for interpreter and module import startup before the first
rule runs, which a native binary does not need.

## Features

- Corrects the previous console command by matching its output against a set of rules.
- Shell integration for **bash**, **zsh**, **fish**, **tcsh** and **PowerShell**.
- Machine readable JSON output for editor plugins, scripts and AI agents with `--json`.
- Interactive confirmation prompt with keyboard navigation, or skip it with `-y`/`--yes`.
- Repeats automatically on failure with `-r`/`--repeat`, and can fix an arbitrary command with
  `--force-command`.
- A JSON settings file with defaults for every option, plus environment variable overrides using the
  `GOFUCK_*` prefix and the legacy `THEFUCK_*` prefix.
- Rule priorities, disabled rules and per-rule settings.
- An optional AI fallback for commands that no rule matches (disabled by default; see
  [AI fallback](#ai-fallback-optional)).
- No third-party Go dependencies: the whole project uses the standard library.

## Requirements

- **Go 1.23 or newer** to build from source. Go 1.22 and older produce macOS binaries without the
  LC_UUID Mach-O load command, which macOS 26 refuses to start.
- **No runtime dependencies.** The compiled binary is self-contained.
- A supported shell for the alias: bash, zsh, fish, tcsh or PowerShell. On Windows, PowerShell is
  recommended; a `cmd.exe` shim is provided as well.

## Installation

### With Homebrew (macOS and Linux)

The repository doubles as a Homebrew tap:

```bash
brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck
brew trust leafevans/go-fuck
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
[microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs). winget-pkgs
requires a professional command name, so the package is published as
`leafevans.gof` and installs the `gof` command. The project itself keeps its
name. Once the manifest is accepted, install with:

```powershell
winget install leafevans.gof
```

Then configure the shell with `gof --alias` (or `gof --setup`); the generated
`fuck` function calls `gof`.

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

### Automatic setup

`go-fuck --setup` appends the line below to the detected shell's configuration file when that file
exists (bash, zsh, fish and tcsh) and prints the command that reloads it. Shells that cannot be
configured automatically, such as PowerShell, and shells whose configuration file does not exist yet
print the line to add instead. The command is idempotent: when the line is already present go-fuck
reports that the shell is already configured and leaves the file untouched.

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
| `--json` | print the suggestions as JSON and exit (useful for scripts and editors; call go-fuck directly rather than through the alias) |
| `-d`, `--debug` | print debug information to stderr |
| `-v`, `--version` | print build and environment details (a colored banner in a terminal, a single line when piped) |
| `-h`, `--help` | print the full option list |
| `--setup` | add the alias to the detected shell's configuration file (bash, zsh, fish, tcsh) |
| `--enable-experimental-instant-mode` | print the initialization code of the experimental instant mode (bash/zsh on Linux and macOS, PowerShell on Windows) |

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

## Instant mode

The experimental instant mode reads the output of the previous command from a terminal log instead
of re-running the command. This is useful for slow, interactive or side effecting commands. It is
implemented for **bash and zsh on Linux and macOS** and for **PowerShell on Windows**; other shells
and platforms warn and fall back to the regular alias.

Enable it by adding the experimental flag to the alias initialization in your shell configuration.

Linux and macOS:

```bash
# ~/.bashrc (or ~/.zshrc for zsh)
eval "$(go-fuck --alias --enable-experimental-instant-mode)"
```

Windows (Windows PowerShell 5.1 and PowerShell 7):

```powershell
# $PROFILE
iex "$(gof --alias --enable-experimental-instant-mode)"
```

Restart the shell. The initialization replaces the terminal shell with a logged subshell: commands
keep working as usual, but go-fuck reads their output from the terminal log, so `fuck` does not
execute the failed command a second time. The log is a temporary file that is removed when the
logged shell exits.

On Windows the logged shell runs on a pseudo console (ConPTY), which requires Windows 10 version
1809 or newer. When ConPTY is not available, the regular alias is installed instead of closing the
terminal.

Limitations:

- The shell profile runs twice per terminal start: once in the terminal shell and once in the
  logged subshell. Slow profiles delay the first prompt twice.
- Changing `PS1` after the alias initialization breaks the command detection in the log. On
  Windows the alias wraps the `prompt` function, so the instant mode line has to be the last
  prompt definition in the profile.
- Full screen programs such as `vim` or `tmux` can confuse the log; fall back to the regular
  mode if a suggestion looks wrong.
- On Windows the terminal size is polled, so a resized window reaches the logged shell with a
  short delay.
- The log keeps the newest 1 MiB of terminal output, matching the reference ring buffer.
- The reference implementation renders the log with pyte. On Linux and macOS go-fuck strips ANSI
  escape sequences and applies carriage return overwrites; on Windows the ConPTY stream is rendered
  with a small screen model that supports cursor addressing, erase sequences and scrolling. Full
  screen programs can still differ from a real terminal.

The `instant_mode` setting and `GOFUCK_INSTANT_MODE=true` are read automatically inside the
logged shell; there is no need to set them manually.

## Machine readable output

`--json` prints the correction result as JSON instead of opening the interactive prompt. It is
intended for editor plugins, scripts and AI agents; call go-fuck directly rather than through the
shell alias, because the alias evaluates whatever go-fuck prints.

```console
$ go-fuck --json --force-command "git brnch"
{
  "command": "git brnch",
  "suggestions": [
    {
      "command": "git branch",
      "priority": 1000,
      "side_effect": false
    },
    {
      "command": "git brnch",
      "priority": 5000,
      "side_effect": false
    }
  ]
}
```

The suggestions are ordered as they would be offered interactively. `priority` is the rule
priority (lower values are tried first). `side_effect` reports whether the interactive mode would
run a side effect, for example `git add`, before the corrected command. The history rule can repeat
the original command as a fallback suggestion; consumers that only want real corrections can skip
suggestions equal to `command`.

`output` is the captured standard output and standard error of the failed command; it is omitted
when the output could not be captured. Output larger than 32 KiB is cut at a UTF-8 boundary and
`output_truncated` is set to `true`. The example above omits the `output` field for brevity.

In JSON mode go-fuck does not execute the corrected command, does not apply side effects and does
not update the shell history; the caller decides what to do with the suggestions. The exit code is
0 once the JSON was written, even when no rule matched and `suggestions` is empty.
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
| `fzf` | boolean | `false` | Use [fzf](https://github.com/junegunn/fzf) as the selection prompt when it is installed. Falls back to the built-in prompt automatically. |
| `no_colors` | boolean | `false` | Disable colored output. |
| `debug` | boolean | `false` | Print debug information to stderr. |
| `priority` | object of integer | `{}` | Per-rule priority overrides. Lower values are tried first; the built-in default is `1000`. |
| `history_limit` | integer | `0` | Maximum number of shell history lines that are read. `0` means no limit. |
| `alter_history` | boolean | `true` | Write the corrected command back to the shell history. |
| `wait_slow_command` | integer | `15` | Seconds to wait for the output of commands listed in `slow_commands`. |
| `slow_commands` | array of strings | `["lein", "react-native", "gradle", "./gradlew", "vagrant"]` | Commands that are expected to be slow to re-run. |
| `repeat` | boolean | `false` | If the corrected command fails, run go-fuck again. |
| `instant_mode` | boolean | `false` | Read the previous command output from the shell logger terminal log instead of re-running the command. See [Instant mode](#instant-mode). |
| `num_close_matches` | integer | `3` | Maximum number of close matches suggested for misspelled commands. |
| `env` | object of string | `{"LC_ALL": "C", "LANG": "C", "GIT_TRACE": "1"}` | Environment variables added when the failed command is re-run. |
| `excluded_search_path_prefixes` | array of strings | `[]` | `PATH` prefixes that are skipped when searching for executables. |
| `ai` | boolean | `false` | Enable the optional AI fallback when no rule matched. See [AI fallback](#ai-fallback-optional). |
| `ai_endpoint` | string | `""` | OpenAI compatible API base URL, for example `https://api.openai.com/v1`. |
| `ai_model` | string | `""` | Chat completion model used by the AI fallback. |
| `ai_timeout` | integer | `10` | Seconds to wait for the AI endpoint. |
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
| `FZF` | boolean | `GOFUCK_FZF=true` | `fzf` |
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
| `AI` | boolean | `GOFUCK_AI=true` | `ai` |
| `AI_ENDPOINT` | string | `GOFUCK_AI_ENDPOINT=https://api.openai.com/v1` | `ai_endpoint` |
| `AI_MODEL` | string | `GOFUCK_AI_MODEL=gpt-4o-mini` | `ai_model` |
| `AI_TIMEOUT` | integer seconds | `GOFUCK_AI_TIMEOUT=20` | `ai_timeout` |

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

### fzf integration

If [fzf](https://github.com/junegunn/fzf) is installed, go-fuck can use it as the selection prompt.
Enable it once in `settings.json`:

```json
{
  "fzf": true
}
```

or per shell with `export GOFUCK_FZF=true`. When fzf runs, the suggestions are listed with a
`fuck>` prompt (`Enter` selects, `Esc` aborts); go-fuck falls back to its built-in prompt
automatically when fzf is not installed or fails. `-y`/`require_confirmation: false` still skips
the selection entirely.

## Rules

The built-in rules are compiled into the binary, and declarative JSON rules can be loaded from the
configuration directory (see [Creating custom rules](#creating-custom-rules)). Each rule decides
whether it applies to a failed command and returns one or more corrected commands. 169 rules are
ported from the reference implementation; go-fuck also adds rules for tools that thefuck does not
handle: the GitHub CLI, kubectl, uv, cargo, Helm, pnpm, the Go toolchain and Node.js. The list below
is generated from the compiled rules:

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
- `cargo_unknown_argument`
- `cargo_unknown_command`
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
- `gh_unknown_command`
- `gh_unknown_flag`
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
- `go_install_version`
- `go_missing_module`
- `go_run`
- `go_unknown_command`
- `go_unknown_flag`
- `gradle_no_task`
- `gradle_wrapper`
- `grep_arguments_order`
- `grep_recursive`
- `grunt_task_not_found`
- `gulp_not_task`
- `has_exists_script`
- `helm_unknown_command`
- `heroku_multiple_apps`
- `heroku_not_command`
- `history`
- `hostscli`
- `ifconfig_device_not_found`
- `java`
- `javac`
- `kubectl_unknown_command`
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
- `node_missing_module`
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
- `pnpm_command_not_found`
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
- `uv_unknown_argument`
- `uv_unknown_subcommand`
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

## AI fallback (optional)

When no built-in or custom rule matches, go-fuck can ask an OpenAI compatible chat completion
endpoint for a correction. The fallback is disabled by default and sends the failed command and the
tail of its output (up to 4,000 bytes) to the configured endpoint:

```json
{
  "ai": true,
  "ai_endpoint": "https://api.openai.com/v1",
  "ai_model": "gpt-4o-mini",
  "ai_timeout": 10
}
```

The API key is read from the `GOFUCK_AI_API_KEY` environment variable (or `OPENAI_API_KEY` when it
is not set) and is never stored in `settings.json`. The same values can be set with `GOFUCK_AI`,
`GOFUCK_AI_ENDPOINT`, `GOFUCK_AI_MODEL` and `GOFUCK_AI_TIMEOUT`.

A returned command is offered exactly like a rule suggestion, so the normal confirmation prompt
applies (`-y` skips it). Request failures and timeouts are reported as a warning and go-fuck falls
back to "No fucks given". The fallback runs only when no rule matched, so the common path is not
affected.

> **Privacy:** only enable the fallback if you are fine with sending failed commands and their
> output to the configured endpoint.

## Creating custom rules

There are two ways to add a rule.

### JSON rules (no rebuild)

go-fuck loads extra rules from JSON files in the `rules` directory inside its configuration
directory:

| Operating system | Directory |
| --- | --- |
| Linux / BSD | `~/.config/go-fuck/rules` (or `$XDG_CONFIG_HOME/go-fuck/rules`) |
| macOS | `~/Library/Application Support/go-fuck/rules` |
| Windows | `%AppData%\go-fuck\rules` |

Each file contains one rule:

```json
{
  "name": "git_remote_typo",
  "enabled": true,
  "priority": 900,
  "match": {
    "script": "^git remote add (\\w+)$",
    "output": "No such remote '([^']+)'"
  },
  "commands": [
    "git remote remove {1}"
  ]
}
```

- `name` is required, must match `^[a-z0-9][a-z0-9_.]*$` and must not collide with a built-in rule.
- `match.script` and `match.output` are regular expressions; at least one is required. When
  `match.output` is present the command output must be available, otherwise the rule is skipped.
- `commands` contains the suggested corrections. `{0}`, `{1}`, ... are replaced with the capture
  groups of `match.output`; `{script}` is replaced with the original command line.
- `enabled` (default `true`) and `priority` (default `1000`) behave like the flags of built-in
  rules and can be overridden with the `rules`, `exclude_rules` and `priority` settings.

The directory is read on every invocation, so a new rule is active the next time you run `fuck`;
no restart or rebuild is needed. Rules are purely declarative: they cannot run commands as side
effects or access the network. Invalid files are reported on stderr and skipped without affecting
the built-in rules.

### Compiled-in rules

Adding a rule to the binary still follows the porting workflow:

1. Read [PORTING.md](PORTING.md). It describes the rule interface, the helpers and the conventions.
2. Create `internal/rules/<rule_name>.go` and `internal/rules/<rule_name>_test.go`, port the test
   cases from thefuck and register the rule from an `init` function.
3. Run the checks described in [CONTRIBUTING.md](CONTRIBUTING.md).
4. Regenerate the rule list in this file with `make rules-list`.

JSON rules are the lightweight option for local and machine specific corrections; compiled-in rules
are the right place for generally useful rules and for rules that need to inspect files or run
external commands.

## Differences from thefuck

go-fuck is a reimplementation, not a drop-in replacement for every thefuck feature:

- **Complete rule port plus extras.** All 169 rules of the reference implementation are ported,
  plus go-fuck specific rules for the GitHub CLI, kubectl, uv, cargo, Helm, pnpm, the Go toolchain
  and Node.js; the list above is authoritative. The porting conventions are documented in
  [PORTING.md](PORTING.md).
- **Instant mode covers Linux, macOS and Windows.** The experimental instant mode reads the
  previous command output from a terminal log instead of re-running the command. It is implemented
  for bash and zsh on Linux and macOS and for PowerShell on Windows (hosted on ConPTY); other
  platforms and shells warn and fall back to the regular alias. The reference implementation
  emulates the terminal with pyte to render the log; go-fuck uses a simplified renderer (ANSI
  stripping on Linux and macOS, a small screen model for the ConPTY stream on Windows), which can
  differ for full screen programs.
- **Shell logger.** `-l`/`--shell-logger` starts a shell under a pseudo terminal and logs its
  output. It is available on Linux and macOS (POSIX pseudo terminals) and on Windows (ConPTY);
  other platforms report that the shell logger is not supported.
- **In-memory memoization instead of an on-disk cache.** go-fuck is a short-lived process, so
  rule results are memoized only for the duration of one invocation. Thefuck's `@cache` persisted
  values between runs.
- **Rules are compiled in, with declarative JSON extensions.** The ported rule set is compiled into
  the binary. Additional rules can be dropped as JSON files into the configuration directory (see
  [Creating custom rules](#creating-custom-rules)); there is no Python rule API, no third-party rule
  package format and no way for a custom rule to run code or side effects.
- **No Python-apt database.** thefuck's `apt_get` rule uses the database installed by
  `python-commandnotfound`. go-fuck has no Python runtime, so the apt rules are ported without that
  database.
- **Settings use JSON.** Settings live in `settings.json` instead of `settings.py`, and the
  legacy `THEFUCK_*` environment variables are still understood alongside the new `GOFUCK_*` ones.
- **fzf integration is an extra.** thefuck has no fuzzy finder integration; `GOFUCK_FZF=true` /
  `"fzf": true` opts in to using fzf as the selection prompt.
- **Optional AI fallback.** thefuck has no model based fallback. go-fuck can ask an OpenAI
  compatible endpoint for a correction when no rule matched; it is opt-in, disabled by default and
  only sends the failed command and its output tail.

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