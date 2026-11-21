package rules

import (
	"os/exec"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
)

// gitRepositoryOutput passes only repository selectors to a read-only Git
// query. Other global options can change the operation itself.
func gitRepositoryOutput(script, subcommand string, args ...string) ([]byte, error) {
	shell := shells.Current()
	start, _, ok := specific.GitSubcommandSpan(script, subcommand, shell)
	if !ok {
		return nil, nil
	}
	prefix := shell.SplitCommand(script[:start])
	if len(prefix) == 0 {
		return nil, nil
	}
	for index := 1; index < len(prefix); index++ {
		arg := prefix[index]
		switch {
		case arg == "-C", arg == "--git-dir", arg == "--work-tree":
			index++
			if index == len(prefix) {
				return nil, nil
			}
		case strings.HasPrefix(arg, "--git-dir="), strings.HasPrefix(arg, "--work-tree="):
		default:
			return nil, nil
		}
	}
	return exec.Command("git", append(prefix[1:], args...)...).Output()
}

func gitHasOperation(script, subcommand string, operations ...string) bool {
	first := true
	for _, word := range specific.GitCommandWords(script, shells.Current()) {
		if !first {
			return slices.Contains(operations, word)
		}
		if word != subcommand {
			return false
		}
		first = false
	}
	return false
}

func gitHasSubcommand(script, subcommand string) bool {
	_, _, ok := specific.GitSubcommandSpan(script, subcommand, shells.Current())
	return ok
}

func gitReplaceArgument(script, before, after string, includePaths bool) (string, bool) {
	shell := shells.Current()
	start, end, ok := specific.GitArgumentSpan(script, before, shell, includePaths)
	if !ok {
		return "", false
	}
	return script[:start] + shell.Quote(after) + script[end:], true
}

// gitInsertAfterSubcommand adds flags without changing unrelated arguments.
func gitInsertAfterSubcommand(script, subcommand, flags string) (string, bool) {
	_, end, ok := specific.GitSubcommandSpan(script, subcommand, shells.Current())
	if !ok {
		return "", false
	}
	return script[:end] + " " + flags + script[end:], true
}
