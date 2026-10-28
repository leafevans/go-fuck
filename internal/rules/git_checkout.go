package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCheckoutPathspec matches the path that git did not find.
var gitCheckoutPathspec = regexp.MustCompile(`error: pathspec '([^']*)' did not match any file\(s\) known to git`)

// gitCheckoutBranches returns the local and remote branch names. It is a
// variable so that tests do not need git to be installed.
var gitCheckoutBranches = func() []string {
	output, err := exec.Command("git", "branch", "-a", "--no-color", "--no-column").Output()
	if err != nil {
		return nil
	}
	return gitCheckoutParseBranches(string(output))
}

// gitCheckoutParseBranches parses the output of "git branch -a".
func gitCheckoutParseBranches(commandOutput string) []string {
	branches := []string{}
	for _, line := range strings.Split(commandOutput, "\n") {
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

// gitCheckoutMatch reports whether git failed because the branch or file is
// unknown.
func gitCheckoutMatch(command *types.Command) bool {
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
	closestBranch := utils.GetClosest(missingFile, gitCheckoutBranches(), 0.6, false)

	newCommands := []string{}
	if closestBranch != "" {
		newCommands = append(newCommands,
			utils.ReplaceArgument(command.Script, missingFile, closestBranch))
	}

	parts := command.ScriptParts()
	if len(parts) > 1 && parts[1] == "checkout" {
		newCommands = append(newCommands,
			utils.ReplaceArgument(command.Script, "checkout", "checkout -b"))
	}

	if len(newCommands) == 0 {
		newCommands = append(newCommands,
			shells.Current().And("git branch "+missingFile, command.Script))
	}
	return newCommands, nil
}

func init() {
	register(types.NewRule("git_checkout",
		specific.GitSupportMatch(gitCheckoutMatch),
		specific.GitSupportNew(gitCheckoutNewCommand)))
}
