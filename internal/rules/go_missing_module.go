package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goMissingPackage matches the "go get" command that the Go toolchain prints
// to suggest the missing module.
var goMissingPackage = regexp.MustCompile(`(?m)^\tgo get (.+)$`)

// goMissingModuleMatch matches the error that the Go toolchain prints when a
// package is imported but not provided by any required module.
func goMissingModuleMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "no required module provides package") &&
		strings.Contains(output, "to add it:")
}

// goMissingModuleNewCommand adds the missing modules with go get and reruns
// the failed command.
func goMissingModuleNewCommand(command *types.Command) ([]string, error) {
	packages := []string{}
	for _, match := range goMissingPackage.FindAllStringSubmatch(command.Output, -1) {
		packages = append(packages, strings.TrimSpace(match[1]))
	}
	if len(packages) == 0 {
		return nil, nil
	}
	fixed := shells.Current().And("go get "+strings.Join(packages, " "), command.Script)
	return []string{fixed}, nil
}

func init() {
	register(types.NewRule(
		"go_missing_module",
		utils.ForApp([]string{"go"}, 0, goMissingModuleMatch),
		goMissingModuleNewCommand,
	))
}
