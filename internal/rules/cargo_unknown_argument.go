package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

func init() {
	register(types.NewRule(
		"cargo_unknown_argument",
		utils.ForApp([]string{"cargo"}, 0, clapArgumentMatch),
		clapArgumentNewCommand,
	))
}
