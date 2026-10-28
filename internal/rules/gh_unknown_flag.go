package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// ghUnknownFlag matches the error line of the GitHub CLI for an unknown
	// flag, for example: unknown flag: --stat.
	ghUnknownFlag = regexp.MustCompile(`unknown flag: (--[A-Za-z0-9-]+)`)
	// ghFlag matches a long flag in the Flags section of the usage output.
	ghFlag = regexp.MustCompile(`(?m)^\s+(?:-\w,\s+)?(--[A-Za-z0-9-]+)`)
)

// ghUnknownFlagMatch matches the error that the GitHub CLI prints for an
// unknown flag when the usage output lists the available flags.
func ghUnknownFlagMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "unknown flag: --") &&
		strings.Contains(output, "Flags:")
}

// ghUnknownFlagNewCommand replaces the unknown flag with the closest flag of
// the usage output.
func ghUnknownFlagNewCommand(command *types.Command) ([]string, error) {
	match := ghUnknownFlag.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	flags := []string{}
	for _, flagMatch := range ghFlag.FindAllStringSubmatch(command.Output, -1) {
		flags = append(flags, flagMatch[1])
	}
	closest := utils.GetClosestDefault(broken, flags)
	if closest == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, broken, closest)}, nil
}

func init() {
	register(types.NewRule(
		"gh_unknown_flag",
		utils.ForApp([]string{"gh"}, 0, ghUnknownFlagMatch),
		ghUnknownFlagNewCommand,
	))
}
