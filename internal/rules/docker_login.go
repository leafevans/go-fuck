package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dockerLoginMatch reports whether docker failed because the registry requires
// authentication.
func dockerLoginMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "docker") &&
		strings.Contains(command.Output, "access denied") &&
		strings.Contains(command.Output, "may require 'docker login'")
}

// dockerLoginNewCommand logs in to the registry and retries the command.
func dockerLoginNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("docker login", command.Script)}, nil
}

func init() {
	register(types.NewRule("docker_login",
		utils.ForApp([]string{"docker"}, 0, dockerLoginMatch),
		dockerLoginNewCommand))
}
