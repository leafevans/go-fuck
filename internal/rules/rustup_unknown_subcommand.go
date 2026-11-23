package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

func init() {
	register(types.NewRule(
		"rustup_unknown_subcommand",
		utils.ForApp([]string{"rustup"}, 0, clapSubcommandMatch),
		clapSubcommandNewCommand,
	))
}
