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

// gitReplaceSubcommand replaces the top-level Git operation without touching
// a branch, path, or option with the same text.
func gitReplaceSubcommand(script, before, after string) (string, bool) {
	start, end, ok := specific.GitSubcommandSpan(script, before, shells.Current())
	if !ok {
		return "", false
	}
	return script[:start] + shells.Current().Quote(after) + script[end:], true
}

// gitOperationWordSpan returns the first word after the Git subcommand.
func gitOperationWordSpan(script, subcommand string) ([2]int, string, bool) {
	first := true
	for span, word := range specific.GitCommandWords(script, shells.Current()) {
		if first {
			if word != subcommand {
				return [2]int{}, "", false
			}
			first = false
			continue
		}
		return span, word, true
	}
	return [2]int{}, "", false
}

// gitReplaceOperation replaces the nested operation, for example "add" in
// "git remote add origin URL".
func gitReplaceOperation(script, subcommand, before, after string) (string, bool) {
	span, word, ok := gitOperationWordSpan(script, subcommand)
	if !ok || word != before {
		return "", false
	}
	return script[:span[0]] + shells.Current().Quote(after) + script[span[1]:], true
}

// gitInsertAfterSubcommand adds flags without changing unrelated arguments.
func gitInsertAfterSubcommand(script, subcommand, flags string) (string, bool) {
	_, end, ok := specific.GitSubcommandSpan(script, subcommand, shells.Current())
	if !ok {
		return "", false
	}
	return script[:end] + " " + flags + script[end:], true
}

// gitCommandWithSubcommand keeps Git's global repository selectors when
// producing a command with different arguments.
func gitCommandWithSubcommand(script, subcommand, replacement string) (string, bool) {
	start, _, ok := specific.GitSubcommandSpan(script, subcommand, shells.Current())
	if !ok {
		return "", false
	}
	return strings.TrimRight(script[:start], " \t") + " " + replacement, true
}
