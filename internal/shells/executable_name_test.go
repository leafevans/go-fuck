package shells

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

func TestAliasesUseExecutableName(t *testing.T) {
	resetShellsState(t)
	consts.SetExecutableName("gof")
	t.Cleanup(func() { consts.SetExecutableName(consts.AppName) })

	aliases := map[string]string{
		"bash":       NewBash().AppAlias("fuck"),
		"zsh":        NewZsh().AppAlias("fuck"),
		"fish":       NewFish().AppAlias("fuck"),
		"tcsh":       NewTcsh().AppAlias("fuck"),
		"powershell": NewPowershell().AppAlias("fuck"),
		"generic":    NewGeneric().AppAlias("fuck"),
	}
	for name, alias := range aliases {
		if strings.Contains(alias, consts.AppName) {
			t.Errorf("%s alias still contains %q: %s", name, consts.AppName, alias)
		}
		if !strings.Contains(alias, "gof") {
			t.Errorf("%s alias does not contain the executable name: %s", name, alias)
		}
	}

	configurations := map[string]*Configuration{
		"bash":       NewBash().HowToConfigure(),
		"zsh":        NewZsh().HowToConfigure(),
		"fish":       NewFish().HowToConfigure(),
		"tcsh":       NewTcsh().HowToConfigure(),
		"powershell": NewPowershell().HowToConfigure(),
	}
	for name, configuration := range configurations {
		if configuration == nil || !strings.Contains(configuration.Content, "gof") {
			t.Errorf("%s configuration does not use the executable name: %+v", name, configuration)
		}
	}
}
