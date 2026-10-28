package rules

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dirtyUntarExtensions lists the archive file extensions that tar extracts.
var dirtyUntarExtensions = []string{
	".tar", ".tar.Z", ".tar.bz2", ".tar.gz", ".tar.lz",
	".tar.lzma", ".tar.xz", ".taz", ".tb2", ".tbz", ".tbz2",
	".tgz", ".tlz", ".txz", ".tz",
}

// dirtyUntarIsExtract reports whether the tar command extracts an archive.
func dirtyUntarIsExtract(script string) bool {
	if strings.Contains(script, "--extract") {
		return true
	}
	// Python: cmd = cmd.split(); return len(cmd) > 1 and 'x' in cmd[1]
	fields := strings.Fields(script)
	return len(fields) > 1 && strings.Contains(fields[1], "x")
}

// dirtyUntarFile returns the archive name and the directory it unpacks into.
func dirtyUntarFile(parts []string) (filename, directory string, ok bool) {
	for _, part := range parts {
		for _, extension := range dirtyUntarExtensions {
			if strings.HasSuffix(part, extension) {
				return part, part[:len(part)-len(extension)], true
			}
		}
	}
	return "", "", false
}

// dirtyUntarMatch reports whether tar extracted an archive into the current
// directory without the -C flag.
func dirtyUntarMatch(command *types.Command) bool {
	if strings.Contains(command.Script, "-C") || !dirtyUntarIsExtract(command.Script) {
		return false
	}
	_, _, ok := dirtyUntarFile(command.ScriptParts())
	return ok
}

// dirtyUntarNewCommand extracts the archive into its own directory.
func dirtyUntarNewCommand(command *types.Command) ([]string, error) {
	_, directory, ok := dirtyUntarFile(command.ScriptParts())
	if !ok {
		return nil, nil
	}
	quotedDirectory := shells.Current().Quote(directory)
	script := shells.Current().And(
		"mkdir -p "+quotedDirectory,
		command.Script+" -C "+quotedDirectory)
	return []string{script}, nil
}

// dirtyUntarSideEffect removes the files that the first tar invocation
// extracted, so that they do not clash with the files extracted into the new
// directory.
func dirtyUntarSideEffect(oldCommand *types.Command, script string) error {
	filename, _, ok := dirtyUntarFile(oldCommand.ScriptParts())
	if !ok {
		return nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader, err := dirtyUntarArchiveReader(file)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	archive := tar.NewReader(reader)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Directories are not removed: it cannot be known whether they
		// existed before the archive was extracted.
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if !dirtyUntarInsideDirectory(header.Name, workingDirectory) {
			// It is unsafe to overwrite files outside of the current
			// directory.
			continue
		}
		// The error is ignored for the same reason as above.
		_ = os.Remove(header.Name)
	}
	return nil
}

// dirtyUntarArchiveReader returns a reader for the archive contents,
// decompressing gzip and bzip2 archives.
//
// Deviation: the original implementation supports the tar.Z, tar.lz, tar.lzma
// and tar.xz formats through Python's tarfile module. Go's standard library
// only provides gzip and bzip2 decompression, so files extracted from the
// other compressed formats are not cleaned up. Matching and the corrected
// commands are not affected.
func dirtyUntarArchiveReader(file *os.File) (io.Reader, error) {
	header := make([]byte, 2)
	read, err := io.ReadFull(file, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	header = header[:read]
	switch {
	case bytes.HasPrefix(header, []byte{0x1f, 0x8b}):
		return gzip.NewReader(file)
	case bytes.HasPrefix(header, []byte("BZh")):
		return bzip2.NewReader(file), nil
	default:
		return file, nil
	}
}

// dirtyUntarInsideDirectory reports whether name resolves to a path inside
// directory.
func dirtyUntarInsideDirectory(name, directory string) bool {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absolute, directory)
}

func init() {
	rule := types.NewRule("dirty_untar",
		utils.ForApp([]string{"tar"}, 0, dirtyUntarMatch),
		dirtyUntarNewCommand)
	rule.SideEffect = dirtyUntarSideEffect
	register(rule)
}
