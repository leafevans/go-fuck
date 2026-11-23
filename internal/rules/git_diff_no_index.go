package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitDiffNoIndexMatch reports whether two files were compared without the
// --no-index flag.
func gitDiffNoIndexMatch(command *types.Command) bool {
	files := 0
	first, paths := true, false
	for _, word := range specific.GitCommandPathWords(command.Script, shells.Current()) {
		switch {
		case first:
			if word != "diff" {
				return false
			}
			first = false
		case paths:
			files++
		case word == "--":
			paths = true
		case strings.HasPrefix(word, "-"):
			// Only count operands when the option cannot consume the next word.
			if !gitDiffStandaloneOption(word) {
				return false
			}
		default:
			files++
		}
	}
	return files == 2
}

func gitDiffStandaloneOption(word string) bool {
	switch word {
	case "-p", "--patch", "--stat", "--name-only", "--name-status", "--shortstat", "--numstat", "--summary", "--raw", "--color", "--no-color", "--color-words", "--word-diff", "--exit-code", "--quiet", "--check", "--minimal", "--binary", "-w", "-b":
		return true
	}
	return strings.HasPrefix(word, "--") && strings.Contains(word, "=")
}

// gitDiffNoIndexNewCommand adds the --no-index flag.
func gitDiffNoIndexNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitInsertAfterSubcommand(command.Script, "diff", "--no-index")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_diff_no_index",
		specific.GitSupportMatch(gitDiffNoIndexMatch),
		specific.GitSupportNew(gitDiffNoIndexNewCommand)))
}
