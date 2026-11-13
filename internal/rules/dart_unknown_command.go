package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func init() {
	register(types.NewRule(
		"dart_unknown_command",
		utils.ForApp([]string{"dart"}, 0, dartUnknownCommandMatch),
		dartUnknownCommandNewCommand,
	))
}
