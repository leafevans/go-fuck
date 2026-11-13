package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitDubiousOwnershipError matches the error that git prints when the
// repository is owned by another user.
const gitDubiousOwnershipError = "detected dubious ownership in repository at"

// gitDubiousOwnershipCommand matches the exemption command that git prints
// itself. Git shell quotes the path for us, so the rule reuses the command as
// printed instead of quoting the path again.
var gitDubiousOwnershipCommand = regexp.MustCompile(`(?m)^\tgit config --global --add safe\.directory (.*)$`)

// gitDubiousOwnershipMatch reports whether git refused to use a repository
// because of its ownership.
func gitDubiousOwnershipMatch(command *types.Command) bool {
	return strings.Contains(command.Output, gitDubiousOwnershipError)
}

// gitDubiousOwnershipNewCommand adds the directory to safe.directory with the
// command that git suggests and retries the failed command.
func gitDubiousOwnershipNewCommand(command *types.Command) ([]string, error) {
	match := gitDubiousOwnershipCommand.FindStringSubmatch(command.Output)
	if match == nil || strings.TrimSpace(match[1]) == "" {
		return nil, nil
	}
	exemption := "git config --global --add safe.directory " + strings.TrimSpace(match[1])
	return []string{shells.Current().And(exemption, command.Script)}, nil
}

func init() {
	register(types.NewRule("git_dubious_ownership",
		specific.GitSupportMatch(gitDubiousOwnershipMatch),
		specific.GitSupportNew(gitDubiousOwnershipNewCommand)))
}
