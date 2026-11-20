package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/ui"
)

func TestJSONNoRerunLeavesOriginalUntouched(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("GOFUCK_WAIT_COMMAND", "60")
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	t.Chdir(dir)
	const command = "echo original > ran"
	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--no-rerun", "--force-command", command)
	var result jsonResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || code != 0 || result.Command != command || result.SchemaVersion != 1 || result.Suggestions == nil {
		t.Fatalf("preview = (%d, %q, %q), JSON error: %v", code, stdout, stderr, err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("preview touched marker: %v", err)
	}
	code, _, stderr = runMain(t, "go-fuck", "--json", "--force-command", command)
	if _, err := os.Stat(marker); code != 0 || err != nil {
		t.Fatalf("default capture did not run the original: exit=%d stderr=%q stat=%v", code, stderr, err)
	}
}

func TestNoRerunStillMatchesRulesWithoutOutput(t *testing.T) {
	resetEntrypointsState(t)
	const command = "git commit"
	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--no-rerun", "--force-command", command)
	var result jsonResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || code != 0 {
		t.Fatalf("got exit=%d stdout=%q stderr=%q error=%v", code, stdout, stderr, err)
	}
	for _, suggestion := range result.Suggestions {
		if suggestion.Rule == "git_hook_bypass" {
			return
		}
	}
	t.Fatalf("missing output-independent suggestion in %s", stdout)
}

func TestCancelDoesNotReportMissingSuggestions(t *testing.T) {
	resetEntrypointsState(t)
	previous := ui.ReadKey
	t.Cleanup(func() { ui.ReadKey = previous })
	ui.ReadKey = func() (consts.Key, error) { return consts.KeyCtrlC, nil }
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)
	code, stdout, _ := runMain(t, "go-fuck", "--no-rerun", "--force-command", "git commit")
	if code != 1 || stdout != "" || !strings.Contains(diagnostics.String(), "git commit --no-verify") || strings.Contains(diagnostics.String(), "No suggestion") {
		t.Fatalf("got exit=%d stdout=%q diagnostics=%q", code, stdout, diagnostics.String())
	}
}

func TestCorrectionDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
		code int
	}{
		{"missing history", []string{"--no-rerun"}, "No usable command in shell history", 0},
		{"missing output", []string{"--no-rerun", "--force-command", "echo ok"}, "No suggestion without the original command output", 1},
		{"empty known output", []string{"--no-rerun", "--force-command", "echo ok", "--output-from", "empty"}, "No matching correction rule found", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetEntrypointsState(t)
			t.Chdir(t.TempDir())
			if err := os.WriteFile("empty", nil, 0o600); err != nil {
				t.Fatal(err)
			}
			var diagnostics bytes.Buffer
			logs.SetOutput(&diagnostics)
			code, stdout, _ := runMain(t, append([]string{"go-fuck"}, tc.args...)...)
			if code != tc.code || stdout != "" || !strings.Contains(diagnostics.String(), tc.want) {
				t.Fatalf("got exit=%d stdout=%q diagnostics=%q", code, stdout, diagnostics.String())
			}
		})
	}
}
