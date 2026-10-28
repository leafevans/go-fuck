package rules

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dirtyUnzipIsBadZip reports whether the archive contains more than one file,
// which means that it extracts into the current directory instead of a
// directory of its own.
func dirtyUnzipIsBadZip(filename string) bool {
	archive, err := zip.OpenReader(filename)
	if err != nil {
		return false
	}
	defer archive.Close()
	return len(archive.File) > 1
}

// dirtyUnzipZipFile returns the archive named by the unzip command.
//
// unzip works like this:
//
//	unzip [-flags] file[.zip] [file(s) ...] [-x file(s) ...]
//	               ^          ^ files to unzip from the archive
//	               archive to unzip
func dirtyUnzipZipFile(command *types.Command) string {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return ""
	}
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "-") {
			continue
		}
		if strings.HasSuffix(part, ".zip") {
			return part
		}
		return part + ".zip"
	}
	return ""
}

// dirtyUnzipMatch reports whether unzip extracted a multi file archive into the
// current directory.
func dirtyUnzipMatch(command *types.Command) bool {
	if strings.Contains(command.Script, "-d") {
		return false
	}
	zipFile := dirtyUnzipZipFile(command)
	if zipFile == "" {
		return false
	}
	return dirtyUnzipIsBadZip(zipFile)
}

// dirtyUnzipNewCommand extracts the archive into a directory named after it.
func dirtyUnzipNewCommand(command *types.Command) ([]string, error) {
	zipFile := dirtyUnzipZipFile(command)
	if len(zipFile) < 4 {
		return nil, nil
	}
	// Python: shell.quote(_zip_file(command)[:-4])
	directory := zipFile[:len(zipFile)-4]
	script := command.Script + " -d " + shells.Current().Quote(directory)
	return []string{script}, nil
}

// dirtyUnzipSideEffect removes the files that the first unzip invocation
// extracted, so that they do not clash with the files extracted into the new
// directory.
func dirtyUnzipSideEffect(oldCommand *types.Command, script string) error {
	zipFile := dirtyUnzipZipFile(oldCommand)
	if zipFile == "" {
		return nil
	}

	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, file := range archive.File {
		// Directories are not removed: it cannot be known whether they
		// existed before the archive was extracted.
		if file.FileInfo().IsDir() {
			continue
		}
		if !dirtyUnzipInsideDirectory(file.Name, workingDirectory) {
			// It is unsafe to overwrite files outside of the current
			// directory.
			continue
		}
		// The error is ignored for the same reason as above.
		_ = os.Remove(file.Name)
	}
	return nil
}

// dirtyUnzipInsideDirectory reports whether name resolves to a path inside
// directory.
func dirtyUnzipInsideDirectory(name, directory string) bool {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absolute, directory)
}

func init() {
	rule := types.NewRule("dirty_unzip",
		utils.ForApp([]string{"unzip"}, 0, dirtyUnzipMatch),
		dirtyUnzipNewCommand)
	// Python: requires_output = False
	rule.RequiresOutput = false
	rule.SideEffect = dirtyUnzipSideEffect
	register(rule)
}
