package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// lnNoHardLinkPrefix matches the leading ln command.
var lnNoHardLinkPrefix = regexp.MustCompile("^ln ")

// lnNoHardLinkMatch reports whether ln failed because hard links are not
// allowed for directories.
func lnNoHardLinkMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 0 &&
		strings.HasSuffix(command.Output, "hard link not allowed for directory") &&
		parts[0] == "ln"
}

// lnNoHardLinkNewCommand creates a symbolic link instead.
func lnNoHardLinkNewCommand(command *types.Command) ([]string, error) {
	return []string{lnNoHardLinkPrefix.ReplaceAllString(command.Script, "ln -s ")}, nil
}

func init() {
	register(types.NewRule("ln_no_hard_link",
		specific.SudoSupportMatch(lnNoHardLinkMatch),
		specific.SudoSupportNew(lnNoHardLinkNewCommand)))
}
