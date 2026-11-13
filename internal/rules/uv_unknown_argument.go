package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func init() {
	register(types.NewRule(
		"uv_unknown_argument",
		utils.ForApp([]string{"uv"}, 0, clapArgumentMatch),
		clapArgumentNewCommand,
	))
}
