package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// scmCorrectionWrongSCMPatterns maps an SCM to the error it prints when the
// current directory is managed by another SCM.
var scmCorrectionWrongSCMPatterns = map[string]string{
	"git": "fatal: Not a git repository",
	"hg":  "abort: no repository found",
}

// scmCorrectionPathToSCM maps a repository directory to the SCM that manages
// it, in lookup order.
var scmCorrectionPathToSCM = []struct {
	path string
	scm  string
}{
	{".git", "git"},
	{".hg", "hg"},
}

// scmCorrectionActualSCM returns the SCM that manages the current directory, or
// an empty string when none is detected. It is a variable so that tests do not
// depend on the working directory.
var scmCorrectionActualSCM = func() string {
	for _, candidate := range scmCorrectionPathToSCM {
		if info, err := os.Stat(candidate.path); err == nil && info.IsDir() {
			return candidate.scm
		}
	}
	return ""
}

// scmCorrectionMatch reports whether the command was run in a directory that
// belongs to another SCM.
func scmCorrectionMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}

	pattern, ok := scmCorrectionWrongSCMPatterns[parts[0]]
	if !ok {
		return false
	}
	return strings.Contains(command.Output, pattern) && scmCorrectionActualSCM() != ""
}

// scmCorrectionNewCommand runs the same arguments with the actual SCM.
func scmCorrectionNewCommand(command *types.Command) ([]string, error) {
	scm := scmCorrectionActualSCM()
	parts := command.ScriptParts()
	if scm == "" || len(parts) == 0 {
		return nil, nil
	}

	arguments := append([]string{scm}, parts[1:]...)
	return []string{strings.Join(arguments, " ")}, nil
}

func init() {
	register(types.NewRule("scm_correction",
		utils.ForApp([]string{"git", "hg"}, 0, scmCorrectionMatch),
		scmCorrectionNewCommand))
}
