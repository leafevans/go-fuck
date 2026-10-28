package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// tsuruLoginMatch reports whether the tsuru session expired.
func tsuruLoginMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "not authenticated") &&
		strings.Contains(command.Output, "session has expired")
}

// tsuruLoginNewCommand logs in and retries the command.
func tsuruLoginNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("tsuru login", command.Script)}, nil
}

func init() {
	register(types.NewRule("tsuru_login",
		utils.ForApp([]string{"tsuru"}, 0, tsuruLoginMatch),
		tsuruLoginNewCommand))
}
