package rules

import (
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

func init() {
	register(types.NewRule(
		"uv_unknown_subcommand",
		utils.ForApp([]string{"uv"}, 0, clapSubcommandMatch),
		clapSubcommandNewCommand,
	))
}
