package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func init() {
	register(types.NewRule(
		"cargo_unknown_argument",
		utils.ForApp([]string{"cargo"}, 0, clapArgumentMatch),
		clapArgumentNewCommand,
	))
}
