package rules

import "github.com/leafevans/go-fuck/internal/types"

// cdCsMatch reports whether the command is the "cs" typo of "cd".
func cdCsMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 0 && parts[0] == "cs"
}

// cdCsNewCommand replaces the "cs" typo with "cd".
func cdCsNewCommand(command *types.Command) ([]string, error) {
	if len(command.Script) < 2 {
		return nil, nil
	}
	return []string{"cd" + command.Script[2:]}, nil
}

func init() {
	rule := types.NewRule("cd_cs", cdCsMatch, cdCsNewCommand)
	// Python: priority = 900
	rule.Priority = 900
	register(rule)
}
