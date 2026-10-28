package rules

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// dirtyUntarEntry describes one file in a test archive.
type dirtyUntarEntry struct {
	name    string
	content string
	isDir   bool
}

// dirtyUntarWriteArchive creates a plain tar archive with the given entries.
// When compress is true the archive is gzip compressed.
func dirtyUntarWriteArchive(t *testing.T, path string, entries []dirtyUntarEntry, compress bool) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive %q: %v", path, err)
	}
	defer file.Close()

	var writer io.Writer = file
	var gzipWriter *gzip.Writer
	if compress {
		gzipWriter = gzip.NewWriter(file)
		writer = gzipWriter
	}

	tarWriter := tar.NewWriter(writer)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Mode: 0o644}
		if entry.isDir {
			header.Typeflag = tar.TypeDir
			header.Mode = 0o755
		} else {
			header.Size = int64(len(entry.content))
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write archive header %q: %v", entry.name, err)
		}
		if !entry.isDir {
			if _, err := tarWriter.Write([]byte(entry.content)); err != nil {
				t.Fatalf("write archive content %q: %v", entry.name, err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if gzipWriter != nil {
		if err := gzipWriter.Close(); err != nil {
			t.Fatalf("close gzip writer: %v", err)
		}
	}
}

// dirtyUntarFixture creates a temporary working directory that contains the
// extracted archive contents and changes into it.
func dirtyUntarFixture(t *testing.T, filename string) string {
	t.Helper()
	directory := t.TempDir()
	entries := []dirtyUntarEntry{
		{name: "a", content: "*"},
		{name: "b", content: "*"},
		{name: "c", content: "*"},
		{name: "d", isDir: true},
		{name: "d/e", content: "*"},
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	// The original fixture extracts the archive before running the tests.
	for _, file := range []string{"a", "b", "c"} {
		if err := os.WriteFile(file, []byte("*"), 0o644); err != nil {
			t.Fatalf("create extracted file %q: %v", file, err)
		}
	}
	if err := os.Mkdir("d", 0o755); err != nil {
		t.Fatalf("create extracted directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join("d", "e"), []byte("*"), 0o644); err != nil {
		t.Fatalf("create extracted file: %v", err)
	}

	dirtyUntarWriteArchive(t, filename, entries, false)
	return directory
}

// dirtyUntarDirectoryEntries returns the sorted names in the current
// directory.
func dirtyUntarDirectoryEntries(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read working directory: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

func TestRuleDirtyUntar(t *testing.T) {
	resetTestState(t)
	rule := Get("dirty_untar")
	if rule == nil {
		t.Fatal("dirty_untar rule is not registered")
	}

	filenameTemplates := []struct {
		// typed is the archive name as typed by the user.
		typed string
		// unquoted is the archive name on disk.
		unquoted string
		// quoted is the expected directory name in the corrected command.
		quoted string
	}{
		{"foo%s", "foo%s", "foo%s"},
		{"\"foo bar%s\"", "foo bar%s", "'foo bar%s'"},
	}
	scriptTemplates := []string{
		"tar xvf %s",
		"tar -xvf %s",
		"tar --extract -f %s",
	}

	t.Run("match", func(t *testing.T) {
		for _, extension := range dirtyUntarExtensions {
			for _, filename := range filenameTemplates {
				for _, scriptTemplate := range scriptTemplates {
					name := extension + "/" + filename.unquoted + "/" + scriptTemplate
					t.Run(name, func(t *testing.T) {
						script := strings.Replace(scriptTemplate, "%s",
							strings.Replace(filename.typed, "%s", extension, 1), 1)
						assertMatch(t, rule, script, "")
					})
				}
			}
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, script := range []string{
			"tar xvf",
			"tar xf foo.tar -C bar",
			"tar cf foo.tar",
			"tar cvf foo.tar",
			"ls foo.tar",
		} {
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, extension := range dirtyUntarExtensions {
			for _, filename := range filenameTemplates {
				for _, scriptTemplate := range scriptTemplates {
					name := extension + "/" + filename.unquoted + "/" + scriptTemplate
					t.Run(name, func(t *testing.T) {
						typed := strings.Replace(filename.typed, "%s", extension, 1)
						directory := strings.Replace(filename.quoted, "%s", "", 1)
						script := strings.Replace(scriptTemplate, "%s", typed, 1)
						want := "mkdir -p " + directory + " && " +
							strings.Replace(scriptTemplate, "%s", typed, 1) +
							" -C " + directory
						assertNewCommands(t, rule, script, "", []string{want})
					})
				}
			}
		}
	})

	t.Run("side effect", func(t *testing.T) {
		for _, extension := range []string{".tar", ".tar.gz", ".tgz"} {
			for _, filename := range filenameTemplates {
				name := extension + "/" + filename.unquoted
				t.Run(name, func(t *testing.T) {
					archive := strings.Replace(filename.unquoted, "%s", extension, 1)
					dirtyUntarFixture(t, archive)
					typed := strings.Replace(filename.typed, "%s", extension, 1)
					script := strings.Replace("tar xvf %s", "%s", typed, 1)

					command := testCommand(script, "")
					if err := rule.SideEffect(command, ""); err != nil {
						t.Fatalf("side effect returned an error: %v", err)
					}
					want := []string{"d", archive}
					if got := dirtyUntarDirectoryEntries(t); !dirtyUntarEqualStrings(got, want) {
						t.Fatalf("side effect left %q, want %q", got, want)
					}
				})
			}
		}
	})

	t.Run("side effect keeps files outside the working directory", func(t *testing.T) {
		dirtyUntarFixture(t, "foo.tar")
		outsideDirectory := t.TempDir()
		outsideFile := filepath.Join(outsideDirectory, "outside.txt")
		if err := os.WriteFile(outsideFile, []byte("*"), 0o644); err != nil {
			t.Fatalf("create outside file: %v", err)
		}
		dirtyUntarWriteArchive(t, "foo.tar", []dirtyUntarEntry{
			{name: outsideFile, content: "*"},
		}, false)

		command := testCommand("tar xvf foo.tar", "")
		if err := rule.SideEffect(command, ""); err != nil {
			t.Fatalf("side effect returned an error: %v", err)
		}
		if _, err := os.Stat(outsideFile); err != nil {
			t.Fatalf("side effect removed a file outside the working directory: %v", err)
		}
	})

	t.Run("side effect with a gzip archive", func(t *testing.T) {
		directory := t.TempDir()
		previous, err := os.Getwd()
		if err != nil {
			t.Fatalf("get working directory: %v", err)
		}
		if err := os.Chdir(directory); err != nil {
			t.Fatalf("change working directory: %v", err)
		}
		t.Cleanup(func() { _ = os.Chdir(previous) })

		if err := os.WriteFile("a", []byte("*"), 0o644); err != nil {
			t.Fatalf("create extracted file: %v", err)
		}
		if err := os.WriteFile("b", []byte("*"), 0o644); err != nil {
			t.Fatalf("create extracted file: %v", err)
		}
		dirtyUntarWriteArchive(t, "foo.tar.gz", []dirtyUntarEntry{
			{name: "a", content: "*"},
			{name: "b", content: "*"},
		}, true)

		command := testCommand("tar xvf foo.tar.gz", "")
		if err := rule.SideEffect(command, ""); err != nil {
			t.Fatalf("side effect returned an error: %v", err)
		}
		want := []string{"foo.tar.gz"}
		if got := dirtyUntarDirectoryEntries(t); !dirtyUntarEqualStrings(got, want) {
			t.Fatalf("side effect left %q, want %q", got, want)
		}
	})
}

// equalStrings reports whether two string slices are equal.
func dirtyUntarEqualStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
