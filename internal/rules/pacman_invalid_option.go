package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pacmanInvalidOptionRegexp matches the option that must be uppercase when it
// is combined with other options.
var pacmanInvalidOptionRegexp = regexp.MustCompile(" -[dfqrstuv]")

// pacmanInvalidOptionSupportedOptions lists the options that may be combined
// with a pacman action.
const pacmanInvalidOptionSupportedOptions = "surqfdvt"

// pacmanInvalidOptionMatch reports whether pacman rejected an option that must
// be uppercase.
func pacmanInvalidOptionMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Output, "error: invalid option '-") {
		return false
	}
	for _, option := range pacmanInvalidOptionSupportedOptions {
		if strings.Contains(command.Script, " -"+string(option)) {
			return true
		}
	}
	return false
}

// pacmanInvalidOptionNewCommand uppercases the invalid option.
func pacmanInvalidOptionNewCommand(command *types.Command) ([]string, error) {
	option := pacmanInvalidOptionRegexp.FindString(command.Script)
	if option == "" {
		return nil, nil
	}
	return []string{strings.ReplaceAll(command.Script, option, strings.ToUpper(option))}, nil
}

func init() {
	// The original rule assigns the whole archlinux_env() tuple to
	// enabled_by_default. A non-empty Python tuple is always truthy, so the
	// rule is enabled on every platform; the flag is kept at its default for
	// compatibility.
	register(types.NewRule("pacman_invalid_option",
		specific.SudoSupportMatch(utils.ForApp([]string{"pacman"}, 0, pacmanInvalidOptionMatch)),
		pacmanInvalidOptionNewCommand))
}
