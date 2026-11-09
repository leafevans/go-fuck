package rules

import (
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

func init() {
	register(types.NewRule(
		"rustup_unknown_subcommand",
		utils.ForApp([]string{"rustup"}, 0, clapSubcommandMatch),
		clapSubcommandNewCommand,
	))
}
