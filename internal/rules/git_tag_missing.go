package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitTagMissingPattern matches the error that git prints when a tag passed to
// "git tag" does not exist.
var gitTagMissingPattern = regexp.MustCompile(`(?m)^error: tag '([^']+)' not found\.$`)

// gitTagNamesCommand lists the tag names. It is a variable so that tests do
// not depend on the repository that runs them.
var gitTagNamesCommand = func() []string {
	output, err := exec.Command("git", "tag", "--list").Output()
	if err != nil {
		return nil
	}
	return strings.Fields(string(output))
}

// gitTagMissingMatch reports whether "git tag" failed because the tag does not
// exist and another tag is close to it.
func gitTagMissingMatch(command *types.Command) bool {
	if !gitTagCommandUsesName(command.Script) {
		return false
	}
	missing, ok := gitTagMissingName(command.Output)
	if !ok || missing == "" {
		return false
	}
	closest := utils.GetClosest(missing, gitTagNamesCommand(), 0.6, false)
	return closest != "" && closest != missing
}

// gitTagMissingNewCommand replaces the missing tag with the closest tag.
func gitTagMissingNewCommand(command *types.Command) ([]string, error) {
	if !gitTagCommandUsesName(command.Script) {
		return nil, nil
	}
	missing, ok := gitTagMissingName(command.Output)
	if !ok || missing == "" {
		return nil, nil
	}
	closest := utils.GetClosest(missing, gitTagNamesCommand(), 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	corrected := utils.ReplaceArgument(command.Script, missing, closest)
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

// gitTagCommandUsesName reports whether script is a "git tag" subcommand that
// takes an existing tag name.
func gitTagCommandUsesName(script string) bool {
	for _, prefix := range []string{
		"tag -d ",
		"tag --delete ",
		"tag -v ",
		"tag --verify ",
	} {
		if strings.Contains(script, prefix) {
			return true
		}
	}
	return false
}

// gitTagMissingName returns the missing tag name from a git error.
func gitTagMissingName(output string) (string, bool) {
	if match := gitTagMissingPattern.FindStringSubmatch(output); match != nil {
		return match[1], true
	}
	return "", false
}

func init() {
	register(types.NewRule("git_tag_missing",
		utils.ForApp([]string{"git"}, 0, gitTagMissingMatch),
		gitTagMissingNewCommand))
}
