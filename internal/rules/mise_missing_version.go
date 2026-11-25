package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

var (
	miseMissingVersionColor = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	miseMissingVersionShim  = regexp.MustCompile(`^(?:mise ERROR |ERROR )?(Tool|Tools) not installed for shim: ([a-zA-Z0-9][a-zA-Z0-9_.+-]*)$`)
	// Only pass plain tool IDs and concrete resolved versions to the shell.
	miseMissingVersionSpec = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.+:/-]*@v?[0-9]+\.[0-9][a-zA-Z0-9_.+-]*$`)
)

// miseMissingVersionTools reads the complete missing-install block from mise's
// shim diagnostic. An unset default is a different case and is not corrected.
func miseMissingVersionTools(command *types.Command) []string {
	if !strings.Contains(command.Output, "not installed for shim:") {
		return nil
	}
	words := simpleCommandWords(command.Script)
	if len(words) == 0 || words[0].text != command.Script[words[0].start:words[0].end] {
		return nil
	}

	output := miseMissingVersionColor.ReplaceAllString(command.Output, "")
	if strings.Contains(output, "is not trusted") || strings.Contains(output, "mise trust") {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	for i, line := range lines {
		match := miseMissingVersionShim.FindStringSubmatch(line)
		if match == nil || match[2] != words[0].text {
			continue
		}

		var tools []string
		for _, detail := range lines[i+1:] {
			if detail == "Install all missing tools with: mise install" {
				if match[1] == "Tool" && len(tools) == 1 || match[1] == "Tools" && len(tools) > 1 {
					return tools
				}
				return nil
			}
			spec, ok := strings.CutPrefix(detail, "Missing tool version: ")
			if !ok || !miseMissingVersionSpec.MatchString(spec) {
				return nil
			}
			tools = append(tools, spec)
		}
		return nil
	}
	return nil
}

func miseMissingVersionMatch(command *types.Command) bool {
	return len(miseMissingVersionTools(command)) != 0
}

func miseMissingVersionNewCommand(command *types.Command) ([]string, error) {
	tools := miseMissingVersionTools(command)
	if len(tools) == 0 {
		return nil, nil
	}
	// Install only the versions named in the error, not every missing tool in
	// the project. Never add a version to config or change its trust status.
	return []string{shells.Current().And("mise install "+strings.Join(tools, " "), command.Script)}, nil
}

func init() {
	rule := types.NewRule("mise_missing_version", miseMissingVersionMatch, miseMissingVersionNewCommand)
	// Installers can download and execute code, so this rule is opt-in.
	rule.EnabledByDefault = false
	register(rule)
}
