package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// gitCheckoutPathspec matches the path that git did not find.
var gitCheckoutPathspec = newRegexp(`error: pathspec '([^']*)' did not match any file\(s\) known to git`)

// gitCheckoutBranches returns the local and remote branch names. It is a
// variable so that tests do not need git to be installed.
var gitCheckoutBranches = func(script string) []string {
	output, err := gitRepositoryOutput(script, "checkout", "branch", "-a", "--no-color", "--no-column")
	if err != nil {
		return nil
	}
	return gitCheckoutParseBranches(string(output))
}

// gitCheckoutParseBranches parses the output of "git branch -a".
func gitCheckoutParseBranches(commandOutput string) []string {
	branches := []string{}
	for line := range strings.SplitSeq(commandOutput, "\n") {
		if strings.Contains(line, "->") {
			continue
		}
		if strings.HasPrefix(line, "*") {
			if fields := strings.Split(line, " "); len(fields) > 1 {
				line = fields[1]
			}
		}
		if strings.HasPrefix(strings.TrimSpace(line), "remotes/") {
			line = strings.Join(strings.Split(line, "/")[2:], "/")
		}
		if branch := strings.TrimSpace(line); branch != "" {
			branches = append(branches, branch)
		}
	}
	return branches
}

// gitCheckoutMatch reports whether git checkout failed because the branch or
// file is unknown.
//
// Deviation: the reference implementation matches any git command whose output
// contains the pathspec error, so it also fires for "git commit <file>" and
// competes with the git_add rule there. go-fuck limits this rule to the
// checkout command it is named after.
func gitCheckoutMatch(command *types.Command) bool {
	if !gitHasSubcommand(command.Script, "checkout") {
		return false
	}
	return strings.Contains(command.Output, "did not match any file(s) known to git") &&
		!strings.Contains(command.Output, "Did you forget to 'git add'?")
}

// gitCheckoutNewCommand suggests the closest branch or creating a new one.
func gitCheckoutNewCommand(command *types.Command) ([]string, error) {
	pathspecMatch := gitCheckoutPathspec.FindStringSubmatch(command.Output)
	if pathspecMatch == nil {
		return nil, nil
	}
	missingFile := pathspecMatch[1]
	closestBranch := utils.GetClosest(missingFile, gitCheckoutBranches(command.Script), 0.6, false)

	newCommands := []string{}
	if closestBranch != "" {
		if corrected, ok := gitReplaceArgument(command.Script, missingFile, closestBranch, false); ok {
			newCommands = append(newCommands, corrected)
		}
	}

	corrected, ok := gitInsertAfterSubcommand(command.Script, "checkout", "-b")
	if !ok {
		return nil, nil
	}
	return append(newCommands, corrected), nil
}

func init() {
	register(types.NewRule("git_checkout",
		specific.GitSupportMatch(gitCheckoutMatch),
		specific.GitSupportNew(gitCheckoutNewCommand)))
}
