package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cdCorrectionMaxAllowedDiff is the minimum similarity of a directory name
// that is accepted as a correction.
const cdCorrectionMaxAllowedDiff = 0.6

// cdCorrectionGetwd returns the current working directory. It is a variable so
// that tests do not depend on the working directory.
var cdCorrectionGetwd = os.Getwd

// cdCorrectionSubDirs returns the child directories of a directory. It is a
// variable so that tests do not depend on the filesystem.
var cdCorrectionSubDirs = func(parent string) []string {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}

	directories := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}
	return directories
}

// cdCorrectionMatch reports whether cd failed because the directory is missing.
func cdCorrectionMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Script, "cd ") {
		return false
	}
	output := strings.ToLower(command.Output)
	return strings.Contains(output, "no such file or directory") ||
		strings.Contains(output, "cd: can't cd to") ||
		strings.Contains(output, "does not exist")
}

// cdCorrectionNewCommand spell checks the directory path and falls back to the
// cd_mkdir rule when no directory is similar enough.
func cdCorrectionNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}

	separator := string(os.PathSeparator)
	path := parts[1]
	cwd := ""
	if filepath.IsAbs(path) {
		volume := filepath.VolumeName(path)
		cwd = volume + separator
		path = strings.TrimPrefix(path, volume)
	} else {
		workingDirectory, err := cdCorrectionGetwd()
		if err != nil {
			return nil, err
		}
		cwd = workingDirectory
	}
	// Windows accepts both slash styles, so a path typed with forward slashes
	// has to be split the same way as one typed with backslashes. On POSIX
	// systems a backslash is a valid character in a directory name.
	destination := strings.FieldsFunc(path, func(character rune) bool {
		return character == '/' || (separator == "\\" && character == '\\')
	})
	if len(destination) == 0 {
		return nil, nil
	}

	for _, directory := range destination {
		if directory == "." {
			continue
		}
		if directory == ".." {
			cwd = filepath.Dir(cwd)
			continue
		}

		bestMatches := utils.GetCloseMatches(directory, cdCorrectionSubDirs(cwd), 0, cdCorrectionMaxAllowedDiff)
		if len(bestMatches) == 0 {
			return cdMkdirNewCommand(command)
		}
		cwd = filepath.Join(cwd, bestMatches[0])
	}
	return []string{"cd \"" + cwd + "\""}, nil
}

func init() {
	register(types.NewRule("cd_correction",
		specific.SudoSupportMatch(utils.ForApp([]string{"cd"}, 0, cdCorrectionMatch)),
		specific.SudoSupportNew(cdCorrectionNewCommand)))
}
