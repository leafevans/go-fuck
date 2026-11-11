package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goMissingFilePackage matches the Go toolchain error for an argument that is
// not provided by any required module. A mistyped file name without its .go
// suffix is treated as a package path and ends up here.
var goMissingFilePackage = regexp.MustCompile(`(?m)no required module provides package ([^\s;]+); to add it:`)

// goMissingFileAttribute matches the Windows error the Go toolchain prints when
// the file argument of a command does not exist.
var goMissingFileAttribute = regexp.MustCompile(`(?m)GetFileAttributesEx (.+?): The system cannot find the file specified\.`)

// goMissingFileMatch reports whether a go command failed because a file
// argument is mistyped and another file in the same directory is close to it.
func goMissingFileMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "no required module provides package ") &&
		!strings.Contains(command.Output, "GetFileAttributesEx ") {
		return false
	}
	_, _, matched := goMissingFileArgument(command)
	return matched
}

// goMissingFileNewCommand replaces the mistyped file argument with the closest
// existing Go file.
func goMissingFileNewCommand(command *types.Command) ([]string, error) {
	argument, replacement, matched := goMissingFileArgument(command)
	if !matched {
		return nil, nil
	}
	index := strings.LastIndex(command.Script, argument)
	if index < 0 {
		return nil, nil
	}
	return []string{command.Script[:index] + replacement + command.Script[index+len(argument):]}, nil
}

// goMissingFileArgument returns the mistyped file argument of a go command, the
// closest existing Go file and whether the argument is worth replacing.
func goMissingFileArgument(command *types.Command) (argument string, replacement string, matched bool) {
	if match := goMissingFilePackage.FindStringSubmatch(command.Output); match != nil {
		// Package paths contain slashes; only a single path segment can be a
		// mistyped file name.
		if !strings.ContainsAny(match[1], "/\\") {
			if closest, ok := goFileTypoCorrection(match[1]); ok {
				return match[1], closest, true
			}
		}
	}
	if match := goMissingFileAttribute.FindStringSubmatch(command.Output); match != nil {
		if closest, ok := goFileTypoCorrection(match[1]); ok {
			return match[1], closest, true
		}
	}
	return "", "", false
}

// goFileTypoCorrection returns the closest existing Go file for a mistyped file
// argument and reports whether the argument has to be replaced.
func goFileTypoCorrection(argument string) (string, bool) {
	if argument == "" || !strings.Contains(argument, ".") {
		return "", false
	}
	base := filepath.Base(argument)
	if base == "" || base == "." || base == ".." {
		return "", false
	}
	names, err := goDirectoryEntries(filepath.Dir(argument))
	if err != nil {
		return "", false
	}
	candidates := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasSuffix(strings.ToLower(name), ".go") {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return "", false
	}

	// When the extension itself is mistyped (main.g, main.goo) only a candidate
	// with the same stem is a plausible correction. Anything else would guess a
	// file the user did not type.
	extension := strings.ToLower(filepath.Ext(base))
	if extension != ".go" {
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		for _, candidate := range candidates {
			candidateStem := strings.TrimSuffix(candidate, filepath.Ext(candidate))
			if strings.EqualFold(stem, candidateStem) {
				return argument[:len(argument)-len(base)] + candidate, true
			}
		}
		return "", false
	}

	closest := utils.GetClosest(base, candidates, 0.6, false)
	if closest == "" || closest == base {
		return "", false
	}
	return argument[:len(argument)-len(base)] + closest, true
}

func init() {
	register(types.NewRule("go_missing_file",
		utils.ForApp([]string{"go"}, 0, goMissingFileMatch),
		goMissingFileNewCommand))
}
