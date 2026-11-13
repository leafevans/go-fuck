package entrypoints

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// updateJSONGolden rewrites the JSON golden files instead of comparing them.
var updateJSONGolden = flag.Bool("update-json-golden", false, "rewrite the JSON golden files")

// assertJSONGolden compares got with one JSON golden file below testdata.
func assertJSONGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateJSONGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	if got != string(want) {
		t.Fatalf("output does not match testdata/%s; run go test ./internal/entrypoints -run TestJSONGolden -update-json-golden", name)
	}
}

// TestJSONGolden locks the complete JSON document for the cases that editor
// plugins and AI agents consume.
func TestJSONGolden(t *testing.T) {
	sideEffect := func(*types.Command, string) error { return nil }
	cases := []struct {
		name      string
		command   *types.Command
		corrected []*types.CorrectedCommand
		file      string
	}{
		{
			name: "empty",
			file: "json_empty.golden",
		},
		{
			name:    "suggestions",
			command: types.NewCommand("git brnch", "git: 'brnch' is not a git command.\n"),
			corrected: []*types.CorrectedCommand{
				{Script: "git branch", RuleName: "git_unknown_command", Priority: 1000},
				{Script: "git brnch", RuleName: "history", Priority: 5000},
			},
			file: "json_suggestions.golden",
		},
		{
			name:    "side_effect",
			command: types.NewCommandWithoutOutput("git commit"),
			corrected: []*types.CorrectedCommand{
				{Script: "git commit --amend", RuleName: "git_commit_amend", Priority: 2000, SideEffect: sideEffect},
			},
			file: "json_side_effect.golden",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			code, stdout, stderr := captureOutput(t, func() int {
				return printJSON(testCase.command, testCase.corrected)
			})
			if code != 0 {
				t.Fatalf("printJSON() = %d, want 0", code)
			}
			if stderr != "" {
				t.Fatalf("printJSON() wrote to stderr: %q", stderr)
			}
			assertJSONGolden(t, testCase.file, stdout)
		})
	}
}
