package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// Python uses lookbehind assertions, which Go's regexp package does not
// support, so the surrounding text is matched and the command name captured
// instead.
var (
	// terraformNoCommandMistake matches the unknown command reported by
	// terraform.
	terraformNoCommandMistake = regexp.MustCompile(`Terraform has no command named "([^"]+)".`)
	// terraformNoCommandFix matches the command terraform suggests.
	terraformNoCommandFix = regexp.MustCompile(`Did you mean "([^"]+)"\?`)
)

// terraformNoCommandMatch reports whether terraform did not recognize a
// command but suggests a fix.
func terraformNoCommandMatch(command *types.Command) bool {
	return terraformNoCommandMistake.MatchString(command.Output) &&
		terraformNoCommandFix.MatchString(command.Output)
}

// terraformNoCommandNewCommand replaces the unknown command with the suggested
// one.
func terraformNoCommandNewCommand(command *types.Command) ([]string, error) {
	mistake := terraformNoCommandMistake.FindStringSubmatch(command.Output)
	fix := terraformNoCommandFix.FindStringSubmatch(command.Output)
	if mistake == nil || fix == nil {
		return nil, nil
	}
	return []string{strings.ReplaceAll(command.Script, mistake[1], fix[1])}, nil
}

func init() {
	register(types.NewRule("terraform_no_command",
		utils.ForApp([]string{"terraform"}, 0, terraformNoCommandMatch),
		terraformNoCommandNewCommand))
}
