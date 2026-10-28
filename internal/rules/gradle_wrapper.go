package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gradleWrapperWhich looks up an executable. It is a variable so that tests do
// not depend on the installed tools.
var gradleWrapperWhich = utils.Which

// gradleWrapperExists reports whether the current directory contains the gradle
// wrapper script. It is a variable so that tests do not depend on files.
var gradleWrapperExists = func() bool {
	info, err := os.Stat("gradlew")
	return err == nil && !info.IsDir()
}

// gradleWrapperMatch reports whether gradle is missing but the wrapper script
// is available.
func gradleWrapperMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	return gradleWrapperWhich(parts[0]) == "" &&
		strings.Contains(command.Output, "not found") &&
		gradleWrapperExists()
}

// gradleWrapperNewCommand runs the wrapper script instead.
func gradleWrapperNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	return []string{"./gradlew " + strings.Join(parts[1:], " ")}, nil
}

func init() {
	register(types.NewRule("gradle_wrapper",
		utils.ForApp([]string{"gradle"}, 0, gradleWrapperMatch),
		gradleWrapperNewCommand))
}
