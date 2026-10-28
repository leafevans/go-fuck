package rules

import (
	"os"

	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// workonGetAllEnvironments returns the names of the virtualenv environments. It
// is a variable so that tests can inject a fixed list.
var workonGetAllEnvironments = utils.Memoize(func() []string {
	return workonReadEnvironments()
})

// workonReadEnvironments reads the environment names from ~/.virtualenvs.
func workonReadEnvironments() []string {
	entries, err := os.ReadDir(system.ExpandHome("~/.virtualenvs"))
	if err != nil {
		return nil
	}

	environments := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			environments = append(environments, entry.Name())
		}
	}
	return environments
}

// workonMatch reports whether the environment passed to workon does not exist.
func workonMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return false
	}
	return !utils.Contains(workonGetAllEnvironments(), parts[1])
}

// workonNewCommand suggests an existing environment or creates a new one.
func workonNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	misspelled := parts[1]
	createNew := "mkvirtualenv " + misspelled

	available := workonGetAllEnvironments()
	if len(available) == 0 {
		return []string{createNew}, nil
	}
	return append(utils.ReplaceCommand(command, misspelled, available), createNew), nil
}

func init() {
	register(types.NewRule("workon_doesnt_exists",
		utils.ForApp([]string{"workon"}, 0, workonMatch),
		workonNewCommand))
}
