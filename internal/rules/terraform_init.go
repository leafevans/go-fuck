package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// terraformInitMatch reports whether terraform has to be initialized first.
func terraformInitMatch(command *types.Command) bool {
	output := strings.ToLower(command.Output)
	return strings.Contains(output, "this module is not yet installed") ||
		strings.Contains(output, "initialization required")
}

// terraformInitNewCommand initializes terraform and retries the command.
func terraformInitNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("terraform init", command.Script)}, nil
}

func init() {
	register(types.NewRule("terraform_init",
		utils.ForApp([]string{"terraform"}, 0, terraformInitMatch),
		terraformInitNewCommand))
}
