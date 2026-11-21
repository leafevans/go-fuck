package entrypoints

import (
	"archive/zip"
	"bytes"
	json "encoding/json/v2"
	"os"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/ui"
)

func TestPreviewJSONDoesNotRerun(t *testing.T) {
	resetEntrypointsState(t)
	t.Chdir(t.TempDir())
	t.Setenv("GOFUCK_WAIT_COMMAND", "60")
	const command = "echo original > ran"
	code, stdout, stderr := runMain(t, "go-fuck", "--preview", "--json", "--force-command", command)
	var result jsonResult
	if err := json.Unmarshal([]byte(stdout), &result); code != 0 || err != nil || result.Command != command {
		t.Fatalf("preview = (%d, %q, %q), JSON error: %v", code, stdout, stderr, err)
	}
	if _, err := os.Stat("ran"); !os.IsNotExist(err) {
		t.Fatalf("preview ran the original command: %v", err)
	}
}

func TestPreviewLeavesArchiveFilesUntouched(t *testing.T) {
	resetEntrypointsState(t)
	t.Chdir(t.TempDir())
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, name := range []string{"first.txt", "second.txt"} {
		if _, err := writer.Create(name); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte("original"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("files.zip", archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)
	previousFzf := ui.FzfRunner
	t.Cleanup(func() { ui.FzfRunner = previousFzf })
	ui.FzfRunner = func([]string) (string, error) {
		t.Fatal("preview opened fzf")
		return "", nil
	}
	t.Setenv("GOFUCK_FZF", "true")
	const command = "unzip files.zip"
	for _, options := range [][]string{{"--yes"}, {"--repeat"}} {
		args := append([]string{"go-fuck", "--preview", "--force-command", command}, options...)
		code, stdout, stderr := runMain(t, args...)
		if code != 0 || stdout != "" || !strings.Contains(diagnostics.String(), "Before running: Remove files named in the ZIP") {
			t.Fatalf("preview = (%d, %q, %q), diagnostics: %q", code, stdout, stderr, diagnostics.String())
		}
	}
	code, stdout, _ := runMain(t, "go-fuck", "--preview", "--json", "--force-command", command)
	var result jsonResult
	if err := json.Unmarshal([]byte(stdout), &result); code != 0 || err != nil || len(result.Suggestions) == 0 {
		t.Fatalf("archive JSON = (%d, %q), error: %v", code, stdout, err)
	}
	got := result.Suggestions[0]
	if got.Rule != "dirty_unzip" || got.Reason == "" || !got.SideEffect || got.SideEffectDescription == "" {
		t.Fatalf("missing ZIP rule metadata: %+v", got)
	}
	for _, name := range []string{"first.txt", "second.txt"} {
		if contents, err := os.ReadFile(name); string(contents) != "original" || err != nil {
			t.Fatalf("preview changed %s: %q, %v", name, contents, err)
		}
	}
}
