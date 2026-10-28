package rules

import (
	"archive/zip"
	"os"
	"sort"
	"testing"
)

// dirtyUnzipWriteArchive creates a zip archive with the given entries.
func dirtyUnzipWriteArchive(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive %q: %v", path, err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create archive entry %q: %v", name, err)
		}
		if _, err := entry.Write([]byte(entries[name])); err != nil {
			t.Fatalf("write archive entry %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

// dirtyUnzipFixture creates a temporary working directory that contains the
// extracted archive contents and changes into it.
func dirtyUnzipFixture(t *testing.T, filename string) {
	t.Helper()
	directory := t.TempDir()

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
	if err := os.WriteFile("d/e", []byte("*"), 0o644); err != nil {
		t.Fatalf("create extracted file: %v", err)
	}

	dirtyUnzipWriteArchive(t, filename, map[string]string{
		"a": "1", "b": "2", "c": "3", "d/e": "4",
	})
}

// dirtyUnzipDirectoryEntries returns the sorted names in the current
// directory.
func dirtyUnzipDirectoryEntries(t *testing.T) []string {
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

func TestRuleDirtyUnzip(t *testing.T) {
	resetTestState(t)
	rule := Get("dirty_unzip")
	if rule == nil {
		t.Fatal("dirty_unzip rule is not registered")
	}
	if rule.RequiresOutput {
		t.Error("dirty_unzip must not require command output")
	}

	testCases := []struct {
		script   string
		filename string
	}{
		{"unzip café", "café.zip"},
		{"unzip café.zip", "café.zip"},
		{"unzip foo", "foo.zip"},
		{"unzip foo.zip", "foo.zip"},
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range testCases {
			t.Run(testCase.script, func(t *testing.T) {
				dirtyUnzipFixture(t, testCase.filename)
				assertMatch(t, rule, testCase.script, "")
			})
		}
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "unzip foo", "")
		assertNotMatch(t, rule, "unzip -d foo bar.zip", "")
	})

	t.Run("side effect", func(t *testing.T) {
		for _, testCase := range testCases {
			t.Run(testCase.script, func(t *testing.T) {
				dirtyUnzipFixture(t, testCase.filename)
				command := testCommand(testCase.script, "")
				if err := rule.SideEffect(command, ""); err != nil {
					t.Fatalf("side effect returned an error: %v", err)
				}

				entries := dirtyUnzipDirectoryEntries(t)
				if len(entries) != 2 {
					t.Fatalf("side effect left %q, want the archive and directory d", entries)
				}
				hasDirectory := entries[0] == "d" || entries[1] == "d"
				if !hasDirectory {
					t.Fatalf("side effect left %q, want directory d", entries)
				}
				// Some file systems normalize unicode file names, so the
				// archive is checked by looking it up rather than by comparing
				// the directory entry.
				if _, err := os.Stat(testCase.filename); err != nil {
					t.Fatalf("side effect removed the archive: %v", err)
				}
			})
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   string
		}{
			{"unzip café", "unzip café -d 'café'"},
			{"unzip foo", "unzip foo -d foo"},
			{"unzip 'foo bar.zip'", "unzip 'foo bar.zip' -d 'foo bar'"},
			{"unzip foo.zip", "unzip foo.zip -d foo"},
		} {
			t.Run(testCase.script, func(t *testing.T) {
				assertNewCommands(t, rule, testCase.script, "",
					[]string{testCase.want})
			})
		}
	})
}
