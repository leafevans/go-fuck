package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// sudoPatterns lists the error messages that indicate missing root
// privileges.
var sudoPatterns = []string{
	"permission denied",
	"eacces",
	"pkg: insufficient privileges",
	"you cannot perform this operation unless you are root",
	"non-root users cannot",
	"operation not permitted",
	"not super-user",
	"superuser privilege",
	"root privilege",
	"this command has to be run under the root user.",
	"this operation requires root.",
	"requested operation requires superuser privilege",
	"must be run as root",
	"must run as root",
	"must be superuser",
	"must be root",
	"need to be root",
	"need root",
	"needs to be run as root",
	"only root can ",
	"you don't have access to the history db.",
	"authentication is required",
	"edspermissionerror",
	"you don't have write permissions",
	"use `sudo`",
	"sudorequirederror",
	"error: insufficient privileges",
	"updatedb: can not open a temporary file",
}

// sudoMatch reports whether the command failed because it needs root
// privileges.
func sudoMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) > 0 && !utils.Contains(parts, "&&") && parts[0] == "sudo" {
		return false
	}

	output := strings.ToLower(command.Output)
	for _, pattern := range sudoPatterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

// sudoNewCommand runs the command with sudo.
func sudoNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Script, "&&") {
		parts := command.ScriptParts()
		kept := make([]string, 0, len(parts))
		for _, part := range parts {
			if part != "sudo" {
				kept = append(kept, part)
			}
		}
		return []string{"sudo sh -c " + doubleQuote + strings.Join(kept, " ") + doubleQuote}, nil
	}
	if strings.Contains(command.Script, ">") {
		escaped := strings.ReplaceAll(command.Script, doubleQuote, backslash+doubleQuote)
		return []string{"sudo sh -c " + doubleQuote + escaped + doubleQuote}, nil
	}
	return []string{"sudo " + command.Script}, nil
}

func init() {
	register(types.NewRule("sudo", sudoMatch, sudoNewCommand))
}
