package rules

import (
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

func init() {
	register(types.NewRule(
		"uv_unknown_argument",
		utils.ForApp([]string{"uv"}, 0, clapArgumentMatch),
		clapArgumentNewCommand,
	))
}
