// Package specific contains helpers for rules that target a specific
// application or package manager.
package specific

import (
	"iter"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitAliasExpansion matches the alias expansion line printed by git when
// GIT_TRACE is enabled.
var gitAliasExpansion = regexp.MustCompile(`trace: alias expansion: ([^ \t\r\n]+) => ([^\r\n]+)`)

// gitWord locates shell words while retaining their byte offsets, so expanding
// a subcommand does not rewrite the quoting of unrelated arguments.
var gitWord = regexp.MustCompile(`^(?:[^ \t\r\n;&|<>()'"\\]|\\.|'[^']*'|"(?:[^"\\]|\\.)*")+`)

var gitFishWord = regexp.MustCompile(`^(?:[^ \t\r\n;&|<>()'"\\]|\\.|'(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*")+`)

var gitPowerShellWord = regexp.MustCompile("^(?:[^ \\t\\r\\n;&|<>()'\"`]|`.|'(?:[^']|'')*'|\"(?:[^\"`]|`.)*\")+")

// GitSupportMatch wraps a match function so that it only runs for git (or hub)
// commands and with git aliases expanded.
func GitSupportMatch(fn func(*types.Command) bool) func(*types.Command) bool {
	return func(command *types.Command) bool {
		if !utils.IsApp(command, 0, "git", "hub") {
			return false
		}
		return fn(expandGitAliases(command))
	}
}

// GitSupportNew wraps a correction function so that it only runs for git (or
// hub) commands and with git aliases expanded.
func GitSupportNew(fn func(*types.Command) ([]string, error)) func(*types.Command) ([]string, error) {
	return func(command *types.Command) ([]string, error) {
		if !utils.IsApp(command, 0, "git", "hub") {
			return nil, nil
		}
		return fn(expandGitAliases(command))
	}
}

// expandGitAliases replaces a git alias with the command it expands to.
func expandGitAliases(command *types.Command) *types.Command {
	return expandGitAliasesForShell(command, shells.Current())
}

func expandGitAliasesForShell(command *types.Command, shell shells.Shell) *types.Command {
	if command.Output == "" || !strings.Contains(command.Output, "trace: alias expansion:") {
		return command
	}
	match := gitAliasExpansion.FindStringSubmatch(command.Output)
	if match == nil {
		return command
	}

	alias, expansion := match[1], match[2]
	start, end, found := GitSubcommandSpan(command.Script, alias, shell)
	if !found {
		return command
	}
	// Git trace uses shell quoting independently of the caller's shell. Parse
	// it strictly before quoting the expanded arguments for the current shell.
	parts, err := shells.Split(expansion)
	if err != nil || len(parts) == 0 || strings.HasPrefix(parts[0], "!") {
		return command
	}
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = shell.Quote(part)
	}
	return command.WithScript(command.Script[:start] + strings.Join(quoted, " ") + command.Script[end:])
}

// GitSubcommandSpan locates the expected subcommand, skipping Git's global
// options and their values.
func GitSubcommandSpan(script, expected string, shell shells.Shell) (int, int, bool) {
	for span, word := range GitCommandWords(script, shell) {
		return span[0], span[1], word == expected
	}
	return 0, 0, false
}

// GitOptionSpan locates an option after the subcommand but before pathspecs.
func GitOptionSpan(script, expected string, shell shells.Shell) (int, int, bool) {
	return GitArgumentSpan(script, expected, shell, false)
}

// GitArgumentSpan locates a unique argument following the Git subcommand.
// Arguments after the pathspec separator are included only when requested.
func GitArgumentSpan(script, expected string, shell shells.Shell, includePaths bool) (int, int, bool) {
	first := true
	var start, end int
	separator := false
	valid := walkGitWords(script, shell, includePaths, func(span [2]int, word string) bool {
		if first {
			first = false
			return true
		}
		if !separator && word == "--" {
			separator = true
			return true
		}
		if word == expected {
			if end != 0 {
				end = 0
				return false
			}
			start, end = span[0], span[1]
		}
		return true
	})
	if !valid || end == 0 {
		return 0, 0, false
	}
	return start, end, true
}

// GitCommandWords yields the subcommand and arguments before the pathspec separator.
// Each span identifies the original shell word in script.
func GitCommandWords(script string, shell shells.Shell) iter.Seq2[[2]int, string] {
	return func(yield func([2]int, string) bool) { walkGitWords(script, shell, false, yield) }
}

// GitCommandPathWords also yields the separator and literal path arguments.
func GitCommandPathWords(script string, shell shells.Shell) iter.Seq2[[2]int, string] {
	return func(yield func([2]int, string) bool) { walkGitWords(script, shell, true, yield) }
}

// gitShellWord reads just one word. Ordinary unquoted arguments are slices of
// the input. Delegate quoted and escaped words to the caller's shell.
func gitShellWord(rest string, shell shells.Shell) (string, int) {
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case ' ', '\t', '\r', '\n', ';', '&', '|', '<', '>', '(', ')':
			return rest[:i], i
		case '\'', '"', '\\', '`':
			pattern := gitWord
			switch shell.(type) {
			case *shells.Fish:
				pattern = gitFishWord
			case *shells.Powershell:
				pattern = gitPowerShellWord
			}
			raw := pattern.FindString(rest)
			if raw == "" || len(raw) < len(rest) && !strings.ContainsRune(" \t\r\n;&|<>()", rune(rest[len(raw)])) {
				return "", 0
			}
			parts := shell.SplitCommand(raw)
			if len(parts) != 1 {
				return "", 0
			}
			return parts[0], len(raw)
		}
	}
	return rest, len(rest)
}

func walkGitWords(script string, shell shells.Shell, includePaths bool, yield func([2]int, string) bool) bool {
	pos := 0
	skipValue, global := false, true
	for index := 0; pos < len(script); index++ {
		for pos < len(script) && (script[pos] == ' ' || script[pos] == '\t') {
			pos++
		}
		if pos == len(script) || script[pos] == '#' {
			break
		}
		word, size := gitShellWord(script[pos:], shell)
		if size == 0 {
			return false
		}
		span := [2]int{pos, pos + size}
		pos += size
		if index == 0 {
			continue
		}
		if skipValue {
			skipValue = false
			continue
		}
		if word == "--" && !includePaths {
			break
		}
		if global && strings.HasPrefix(word, "-") {
			switch word {
			case "-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env", "--super-prefix", "--attr-source":
				skipValue = true
			}
			continue
		}
		global = false
		if !yield(span, word) {
			break
		}
	}
	return true
}
