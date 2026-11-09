package utils

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// ReplaceArgument replaces a single command line argument.
func ReplaceArgument(script, from, to string) string {
	suffix := " " + from
	if strings.HasSuffix(script, suffix) {
		return script[:len(script)-len(suffix)] + " " + to
	}
	search := " " + from + " "
	if before, after, ok := strings.Cut(script, search); ok {
		return before + " " + to + " " + after
	}
	return script
}

// ReplaceArguments replaces from with every target and returns the commands
// that actually changed.
func ReplaceArguments(script, from string, targets []string) []string {
	corrected := make([]string, 0, len(targets))
	for _, target := range targets {
		replaced := ReplaceArgument(script, from, target)
		if replaced != script {
			corrected = append(corrected, replaced)
		}
	}
	return corrected
}

// ReplaceArgumentsSorted replaces from with every candidate, ordered by its
// similarity to from, and returns the commands that actually changed.
//
// Rules that print several suggestions should use it so that the closest
// candidate is offered first, independently of the order the tool prints.
func ReplaceArgumentsSorted(script, from string, candidates []string) []string {
	return ReplaceArguments(script, from, SortBySimilarity(from, candidates))
}

// ReplaceCommand replaces a misspelled command with its closest matches.
func ReplaceCommand(command *types.Command, broken string, matched []string) []string {
	closeMatches := GetCloseMatches(broken, matched, 0, 0.1)
	results := make([]string, 0, len(closeMatches))
	for _, match := range closeMatches {
		results = append(results, ReplaceArgument(command.Script, broken, strings.TrimSpace(match)))
	}
	return results
}

// FindLineByPrefix returns the first line of output whose trimmed text starts
// with prefix, together with a boolean that reports whether such a line was
// found.
//
// Rules that parse a suggested command should look it up by its prefix instead
// of by its position: the captured output may carry extra empty lines, for
// example when the command was re-run by PowerShell.
func FindLineByPrefix(output, prefix string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return line, true
		}
	}
	return "", false
}

// GetAllMatchedCommands extracts the suggested commands from a command output
// that lists suggestions after one of the given separators.
func GetAllMatchedCommands(commandOutput string, separators ...string) []string {
	results := []string{}
	shouldYield := false
	for _, line := range strings.Split(commandOutput, "\n") {
		matched := false
		for _, separator := range separators {
			if strings.Contains(line, separator) {
				shouldYield = true
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if shouldYield && line != "" {
			results = append(results, strings.TrimSpace(line))
		}
	}
	return results
}
