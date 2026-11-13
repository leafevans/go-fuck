package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func init() {
	register(types.NewRule(
		"flutter_unknown_command",
		utils.ForApp([]string{"flutter"}, 0, dartUnknownCommandMatch),
		dartUnknownCommandNewCommand,
	))
}
