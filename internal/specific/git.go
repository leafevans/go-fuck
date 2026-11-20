// Package specific contains helpers for rules that target a specific
// application or package manager.
package specific

import (
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
var gitWord = regexp.MustCompile(`(?:[^ \t\r\n'"\\]|\\.|'[^']*'|"(?:[^"\\]|\\.)*")+`)

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
	start, end, found := gitAliasSpan(command.Script, alias, shell)
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

// gitAliasSpan locates only the subcommand, skipping Git's global options and
// their values. An alias mentioned in a path, option value or later argument
// must never be expanded.
func gitAliasSpan(script, alias string, shell shells.Shell) (int, int, bool) {
	previousEnd := 0
	skipValue := false
	for index, span := range gitWord.FindAllStringIndex(script, -1) {
		if strings.TrimSpace(script[previousEnd:span[0]]) != "" {
			return 0, 0, false
		}
		previousEnd = span[1]
		if index == 0 {
			continue
		}
		parts := shell.SplitCommand(script[span[0]:span[1]])
		if len(parts) != 1 {
			return 0, 0, false
		}
		if skipValue {
			skipValue = false
			continue
		}
		word := parts[0]
		if word == "--" {
			return 0, 0, false
		}
		if strings.HasPrefix(word, "-") {
			switch word {
			case "-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env", "--super-prefix", "--attr-source":
				skipValue = true
			}
			continue
		}
		return span[0], span[1], word == alias
	}
	return 0, 0, false
}
