package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitPushUpstream matches the commands suggested by git.
var gitPushUpstream = regexp.MustCompile("git push (.*)")

// gitPushMatch reports whether git suggested setting an upstream branch.
func gitPushMatch(command *types.Command) bool {
	if !utils.Contains(command.ScriptParts(), "push") {
		return false
	}
	return strings.Contains(command.Output, "git push --set-upstream")
}

// gitPushNewCommand uses the upstream command suggested by git.
func gitPushNewCommand(command *types.Command) ([]string, error) {
	parts := append([]string{}, command.ScriptParts()...)
	upstreamIndex := utils.Index(parts, "--set-upstream")
	if upstreamIndex < 0 {
		upstreamIndex = utils.Index(parts, "-u")
	}

	if upstreamIndex >= 0 {
		parts = append(parts[:upstreamIndex], parts[upstreamIndex+1:]...)
		// In case of "git push -u" there is no argument to remove.
		if len(parts) > upstreamIndex {
			parts = append(parts[:upstreamIndex], parts[upstreamIndex+1:]...)
		}
	} else {
		// Only the repository and refspec are allowed here; git's suggestion
		// contains them, so they would be duplicated otherwise.
		pushIndex := utils.Index(parts, "push") + 1
		for len(parts) > pushIndex && parts[len(parts)-1][0] != '-' {
			parts = parts[:len(parts)-1]
		}
	}

	matches := gitPushUpstream.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	arguments := strings.ReplaceAll(
		strings.TrimSpace(matches[len(matches)-1][1]), "'", backslash+"'")
	script := utils.ReplaceArgument(strings.Join(parts, " "), "push", "push "+arguments)
	return []string{script}, nil
}

func init() {
	register(types.NewRule("git_push",
		specific.GitSupportMatch(gitPushMatch),
		specific.GitSupportNew(gitPushNewCommand)))
}
