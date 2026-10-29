package entrypoints

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

// jsonTestResult mirrors the machine readable output of --json.
type jsonTestResult struct {
	Command     string               `json:"command"`
	Suggestions []jsonTestSuggestion `json:"suggestions"`
}

// jsonTestSuggestion mirrors one suggestion of the machine readable output.
type jsonTestSuggestion struct {
	Command    string `json:"command"`
	Priority   int    `json:"priority"`
	SideEffect bool   `json:"side_effect"`
}

func TestMainJSONOutput(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("GOFUCK_WAIT_COMMAND", "60")

	command := "echo 'hello" + string(rune(34))
	code, stdout, _ := runMain(t, "go-fuck", "--json", "--force-command", command)
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if result.Command == "" {
		t.Fatalf("JSON output has no command: %s", stdout)
	}

	want := "echo " + string(rune(34)) + "hello" + string(rune(34))
	for _, suggestion := range result.Suggestions {
		if suggestion.Command == want {
			return
		}
	}
	t.Fatalf("suggestions = %+v, want it to contain %q", result.Suggestions, want)
}

func TestMainJSONWithoutSuggestions(t *testing.T) {
	resetEntrypointsState(t)
	code, stdout, _ := captureOutput(t, func() int { return printJSON(nil, nil) })
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"suggestions": []`) {
		t.Fatalf("stdout = %q, want an empty suggestions array", stdout)
	}
}

func TestPrintJSONIncludesSuggestionDetails(t *testing.T) {
	command := types.NewCommand("git commit", "")
	suggestion := &types.CorrectedCommand{
		Script:   "git commit --amend",
		Priority: 2000,
		SideEffect: func(*types.Command, string) error {
			return nil
		},
	}
	code, stdout, _ := captureOutput(t, func() int {
		return printJSON(command, []*types.CorrectedCommand{suggestion})
	})
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if result.Command != "git commit" || len(result.Suggestions) != 1 {
		t.Fatalf("unexpected JSON output: %s", stdout)
	}
	got := result.Suggestions[0]
	if got.Command != "git commit --amend" || got.Priority != 2000 || !got.SideEffect {
		t.Fatalf("unexpected suggestion JSON: %+v", got)
	}
}
