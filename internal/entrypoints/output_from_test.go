package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/logs"
)

// outputFromResult mirrors the part of the machine readable output used here.
type outputFromResult struct {
	Command     string `json:"command"`
	Output      string `json:"output"`
	Suggestions []struct {
		Command string `json:"command"`
		Rule    string `json:"rule"`
	} `json:"suggestions"`
}

// assertOutputFromSuggestion checks that the supplied output was used by the
// rules instead of re-running the command.
func assertOutputFromSuggestion(t *testing.T, stdout string) {
	t.Helper()
	var result outputFromResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if result.Command != "cat testdata" {
		t.Fatalf("command = %q, want cat testdata", result.Command)
	}
	if !strings.Contains(result.Output, "Is a directory") {
		t.Fatalf("output = %q, want the supplied output", result.Output)
	}
	for _, suggestion := range result.Suggestions {
		if suggestion.Command == "ls testdata" && suggestion.Rule == "cat_dir" {
			return
		}
	}
	t.Fatalf("suggestions = %+v, want ls testdata from cat_dir", result.Suggestions)
}

func TestFixCommandOutputFromFile(t *testing.T) {
	resetEntrypointsState(t)
	path := filepath.Join(t.TempDir(), "failed.log")
	if err := os.WriteFile(path, []byte("cat: testdata: Is a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--force-command", "cat testdata", "--output-from", path)
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	assertOutputFromSuggestion(t, stdout)
}

func TestFixCommandOutputFromStdin(t *testing.T) {
	resetEntrypointsState(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = original
		_ = reader.Close()
	})
	go func() {
		_, _ = writer.WriteString("cat: testdata: Is a directory\n")
		_ = writer.Close()
	}()

	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--force-command", "cat testdata", "--output-from", "-")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	assertOutputFromSuggestion(t, stdout)
}

func TestFixCommandOutputFromMissingFile(t *testing.T) {
	resetEntrypointsState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	code, _, _ := runMain(t, "go-fuck", "--json", "--force-command", "cat testdata", "--output-from", filepath.Join(t.TempDir(), "missing.log"))
	if code != 1 {
		t.Fatalf("Main() = %d, want 1", code)
	}
	if !strings.Contains(buffer.String(), "Can't read the command output") {
		t.Fatalf("log output = %q, want the read error", buffer.String())
	}
}

func TestOutputFromRequiresForceCommand(t *testing.T) {
	resetEntrypointsState(t)
	code, _, stderr := runMain(t, "go-fuck", "--output-from", "failed.log")
	if code != 2 {
		t.Fatalf("Main() = %d, want 2; stderr = %q", code, stderr)
	}
	if !strings.Contains(stderr, "--output-from: requires argument --force-command") {
		t.Fatalf("stderr = %q, want the argument error", stderr)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Fatalf("stderr = %q, want the usage line", stderr)
	}
}

// TestFixCommandOutputFromWindowsText checks that a byte order mark and CRLF
// line endings in the supplied output are normalized like captured output.
func TestFixCommandOutputFromWindowsText(t *testing.T) {
	resetEntrypointsState(t)
	path := filepath.Join(t.TempDir(), "failed.log")
	content := "\uFEFFcat: testdata: Is a directory\r\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--force-command", "cat testdata", "--output-from", path)
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	assertOutputFromSuggestion(t, stdout)
}
