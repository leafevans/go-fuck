// Package specific contains helpers for rules that target a specific
// application or package manager.
package specific

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitAliasExpansion matches the alias expansion line printed by git when
// GIT_TRACE is enabled.
var gitAliasExpansion = regexp.MustCompile(`trace: alias expansion: ([^ ]*) => ([^\n]*)`)

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
	if command.Output == "" || !strings.Contains(command.Output, "trace: alias expansion:") {
		return command
	}
	match := gitAliasExpansion.FindStringSubmatch(command.Output)
	if match == nil {
		return command
	}

	alias, expansion := match[1], match[2]
	parts := shells.Current().SplitCommand(expansion)
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = shells.Current().Quote(part)
	}
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(alias) + `\b`)
	return command.WithScript(pattern.ReplaceAllString(command.Script, strings.Join(quoted, " ")))
}
