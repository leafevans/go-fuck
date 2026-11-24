package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"github.com/leafevans/go-fuck/v3/internal/logs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafetyReleaseHookOutputModes(t *testing.T) {
	for _, mode := range []string{"--json", "--preview", "--explain"} {
		t.Run(mode, func(t *testing.T) {
			resetEntrypointsState(t)
			t.Chdir(t.TempDir())
			output := filepath.Join(t.TempDir(), "failure")
			if err := os.WriteFile(output, []byte("husky - pre-commit hook exited with code 1 (error)"), 0600); err != nil {
				t.Fatal(err)
			}
			var diagnostics bytes.Buffer
			logs.SetOutput(&diagnostics)
			// Running the original command would create marker. None of these modes
			// may execute the original or its suggested --no-verify correction.
			code, stdout, stderr := runMain(t, "go-fuck", mode, "--force-command", "git commit -m x; echo ran > marker", "--output-from", output)
			if code != 0 {
				t.Fatalf("exit=%d stdout=%q stderr=%q diagnostics=%q", code, stdout, stderr, diagnostics.String())
			}
			if mode == "--json" {
				var result jsonResult
				if err := json.Unmarshal([]byte(stdout), &result); err != nil {
					t.Fatal(err)
				}
				if result.SchemaVersion != 1 {
					t.Fatal("schema changed")
				}
				found := false
				for _, s := range result.Suggestions {
					if s.Rule == "git_hook_bypass" {
						found = true
						if s.Reason == "" {
							t.Fatal("missing explanation")
						}
					}
					if s.Rule == "git_commit_reset" || s.Rule == "git_commit_amend" {
						t.Fatal("history rewrite enabled by default")
					}
				}
				if !found {
					t.Fatal("valid hook correction lost")
				}
			} else if !strings.Contains(diagnostics.String(), "--no-verify") {
				t.Fatalf("missing suggestion: %q", diagnostics.String())
			}
			if _, err := os.Stat("marker"); !os.IsNotExist(err) {
				t.Fatalf("command executed: %v", err)
			}
		})
	}
}
