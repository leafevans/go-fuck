package rules

import (
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// phpSMatch reports whether php was started with the -s flag, which enables
// syntax highlighting instead of the built-in web server.
func phpSMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return utils.Contains(parts, "-s") && parts[len(parts)-1] != "-s"
}

// phpSNewCommand replaces the mistyped -s flag with the web server flag.
func phpSNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "-s", "-S")}, nil
}

func init() {
	// Python: for_app('php', at_least=2).
	register(types.NewRule("php_s",
		utils.ForApp([]string{"php"}, 2, phpSMatch),
		phpSNewCommand))
}
