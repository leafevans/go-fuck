package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// composeNoSuchService matches the error of docker compose for a service that
// is not defined in the Compose file, for example: no such service: buidl.
var composeNoSuchService = regexp.MustCompile(`no such service: "?([^"\s:]+)"?`)

// composeUnknownServiceMatch reports whether docker compose failed because the
// service does not exist.
func composeUnknownServiceMatch(command *types.Command) bool {
	if !strings.Contains(command.Script, "compose") {
		return false
	}
	if composeNoSuchService.FindStringSubmatch(command.Output) == nil {
		return false
	}
	return len(specific.ComposeServices()) > 0
}

// composeUnknownServiceNewCommand suggests the Compose services that are close
// to the misspelled one.
func composeUnknownServiceNewCommand(command *types.Command) ([]string, error) {
	match := composeNoSuchService.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], specific.ComposeServices()), nil
}

func init() {
	register(types.NewRule(
		"docker_compose_unknown_service",
		utils.ForApp([]string{"docker"}, 0, composeUnknownServiceMatch),
		composeUnknownServiceNewCommand,
	))
}
