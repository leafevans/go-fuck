package rules

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goMissingDirectory matches the package directory that the go command could
// not stat.
var goMissingDirectory = regexp.MustCompile(`stat (.+): directory not found`)

// goDirectoryEntries returns the names in dir.
var goDirectoryEntries = func(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// goMissingPackageMatch reports whether a go command could not find a package
// directory.
func goMissingPackageMatch(command *types.Command) bool {
	match := goMissingDirectory.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	return strings.Contains(command.Script, filepath.Base(match[1]))
}

// goMissingPackageNewCommand replaces the missing directory with the closest
// existing one in the same parent directory.
func goMissingPackageNewCommand(command *types.Command) ([]string, error) {
	match := goMissingDirectory.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	missing := filepath.Base(match[1])
	if missing == "" || missing == "." || missing == string(filepath.Separator) {
		return nil, nil
	}
	names, err := goDirectoryEntries(filepath.Dir(match[1]))
	if err != nil {
		return nil, nil
	}
	closest := utils.GetClosest(missing, names, 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	index := strings.LastIndex(command.Script, missing)
	if index < 0 {
		return nil, nil
	}
	return []string{command.Script[:index] + closest + command.Script[index+len(missing):]}, nil
}

func init() {
	register(types.NewRule("go_missing_package",
		utils.ForApp([]string{"go"}, 0, goMissingPackageMatch),
		goMissingPackageNewCommand))
}
