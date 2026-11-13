package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// goMissingModError is the error that Go prints when no module file exists.
const goMissingModError = "go.mod file not found in current directory or any parent directory"

// goModuleName returns the module name derived from the current directory.
var goModuleName = func() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	name := filepath.Base(dir)
	var builder strings.Builder
	for _, character := range name {
		switch {
		case character >= 'a' && character <= 'z',
			character >= 'A' && character <= 'Z',
			character >= '0' && character <= '9',
			character == '-', character == '_', character == '.':
			builder.WriteRune(character)
		default:
			builder.WriteRune('-')
		}
	}
	name = strings.Trim(builder.String(), "-._")
	if name == "" {
		return "app"
	}
	return name
}

// goMissingModMatch reports whether the command failed because no go.mod
// exists.
func goMissingModMatch(command *types.Command) bool {
	return strings.Contains(command.Output, goMissingModError)
}

// goMissingModNewCommand initializes the module and retries.
func goMissingModNewCommand(command *types.Command) ([]string, error) {
	initialized := "go mod init " + goModuleName()
	return []string{shells.Current().And(initialized, command.Script)}, nil
}

func init() {
	register(types.NewRule("go_mod_init",
		utils.ForApp([]string{"go"}, 0, goMissingModMatch),
		goMissingModNewCommand))
}
