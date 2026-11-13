package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cdMkdirScript matches the directory argument of a cd command.
var cdMkdirScript = regexp.MustCompile("^cd (.*)")

// cdTarget returns the directory argument of a cd command.
//
// The raw command is used instead of the shell split arguments, because
// splitting treats the backslashes of a Windows path as escapes.
func cdTarget(command *types.Command) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(command.Script, " \t"), "sudo "))
	if !strings.HasPrefix(trimmed, "cd ") {
		return ""
	}
	return strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "cd ")), "\"'")
}

// cdMkdirTargetMissing reports whether the target of the cd command does not
// exist. PowerShell reports an existing file as "does not exist" when
// Set-Location is used on it, and creating a directory with the same name would
// fail, so the correction is only offered for a path that is really absent.
func cdMkdirTargetMissing(command *types.Command) bool {
	target := cdTarget(command)
	if target == "" {
		return true
	}
	if _, err := os.Stat(target); err == nil {
		return false
	}
	return true
}

// cdMkdirMatch reports whether cd failed because the directory is missing.
func cdMkdirMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Script, "cd ") {
		return false
	}
	if !cdMkdirTargetMissing(command) {
		return false
	}
	output := strings.ToLower(command.Output)
	return strings.Contains(output, "no such file or directory") ||
		strings.Contains(output, "cd: can't cd to") ||
		strings.Contains(output, "does not exist")
}

// cdMkdirNewCommand creates the directory and changes into it.
func cdMkdirNewCommand(command *types.Command) ([]string, error) {
	replacement := shells.Current().And("mkdir -p $1", "cd $1")
	return []string{cdMkdirScript.ReplaceAllString(command.Script, replacement)}, nil
}

func init() {
	register(types.NewRule("cd_mkdir",
		specific.SudoSupportMatch(utils.ForApp([]string{"cd"}, 0, cdMkdirMatch)),
		specific.SudoSupportNew(cdMkdirNewCommand)))
}
