package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// jsonTestResult mirrors the machine readable output of --json.
type jsonTestResult struct {
	SchemaVersion   int                  `json:"schema_version"`
	Command         string               `json:"command"`
	Output          string               `json:"output"`
	OutputTruncated bool                 `json:"output_truncated"`
	Suggestions     []jsonTestSuggestion `json:"suggestions"`
}

// jsonTestSuggestion mirrors one suggestion of the machine readable output.
type jsonTestSuggestion struct {
	Command    string `json:"command"`
	Priority   int    `json:"priority"`
	SideEffect bool   `json:"side_effect"`
	Rule       string `json:"rule"`
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
	if result.SchemaVersion != jsonSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", result.SchemaVersion, jsonSchemaVersion)
	}
	if result.Command == "" {
		t.Fatalf("JSON output has no command: %s", stdout)
	}

	want := "echo " + string(rune(34)) + "hello" + string(rune(34))
	for _, suggestion := range result.Suggestions {
		if suggestion.Command == want {
			if suggestion.Rule == "" {
				t.Fatalf("suggestion %q has no rule name", suggestion.Command)
			}
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
		RuleName: "git_commit_amend",
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
	if got.Rule != "git_commit_amend" {
		t.Fatalf("rule = %q, want git_commit_amend", got.Rule)
	}
}
func TestPrintJSONIncludesCommandOutput(t *testing.T) {
	command := types.NewCommand("git psu", "fatal: not a git repository\n")
	code, stdout, _ := captureOutput(t, func() int { return printJSON(command, nil) })
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v%s", err, stdout)
	}
	if result.Output != "fatal: not a git repository\n" {
		t.Fatalf("Output = %q, want the command output", result.Output)
	}
	if result.OutputTruncated {
		t.Fatalf("OutputTruncated = true, want false")
	}
}

func TestPrintJSONReportsWriteError(t *testing.T) {
	resetEntrypointsState(t)
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)

	file, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = file
	t.Cleanup(func() { os.Stdout = previous })

	if code := printJSON(types.NewCommand("git status", ""), nil); code != 1 {
		t.Errorf("printJSON() = %d, want 1", code)
	}
	if !strings.Contains(diagnostics.String(), "Can't write the JSON output") {
		t.Errorf("missing write error diagnostic: %q", diagnostics.String())
	}
}

func TestPrintJSONReplacesInvalidUTF8(t *testing.T) {
	// Captured command output can contain arbitrary bytes. encoding/json/v2
	// rejects invalid UTF-8 instead of replacing it like v1 did, so printJSON
	// has to sanitize the strings before encoding them.
	invalid := string([]byte{'a', 0xff, 0xfe, 'b'})
	command := types.NewCommand(invalid, invalid)
	suggestion := &types.CorrectedCommand{Script: invalid, Priority: 1000}

	code, stdout, _ := captureOutput(t, func() int {
		return printJSON(command, []*types.CorrectedCommand{suggestion})
	})
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}
	if !utf8.ValidString(stdout) {
		t.Fatalf("stdout is not valid UTF-8: %q", stdout)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if result.Command != "a\uFFFDb" {
		t.Fatalf("Command = %q, want the invalid bytes replaced", result.Command)
	}
	if result.Output != "a\uFFFDb" {
		t.Fatalf("Output = %q, want the invalid bytes replaced", result.Output)
	}
	if len(result.Suggestions) != 1 || result.Suggestions[0].Command != "a\uFFFDb" {
		t.Fatalf("Suggestions = %+v, want the invalid bytes replaced", result.Suggestions)
	}
}

func TestPrintJSONTruncatesInvalidUTF8Output(t *testing.T) {
	// The replacement characters may make the sanitized output longer than the
	// original limit, so the result has to be capped again.
	output := strings.Repeat(string([]byte{0xff}), jsonOutputLimit+128)
	command := types.NewCommand("noisy", output)
	code, stdout, _ := captureOutput(t, func() int { return printJSON(command, nil) })
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v%s", err, stdout)
	}
	if !result.OutputTruncated {
		t.Fatal("OutputTruncated = false, want true")
	}
	if len(result.Output) > jsonOutputLimit {
		t.Fatalf("len(Output) = %d, want at most %d", len(result.Output), jsonOutputLimit)
	}
	if !strings.Contains(result.Output, string(utf8.RuneError)) {
		t.Fatalf("Output = %q, want replacement characters", result.Output)
	}
}

func TestPrintJSONTruncatesLongOutput(t *testing.T) {
	output := strings.Repeat("a", jsonOutputLimit+128)
	command := types.NewCommand("noisy", output)
	code, stdout, _ := captureOutput(t, func() int { return printJSON(command, nil) })
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}

	var result jsonTestResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v%s", err, stdout)
	}
	if !result.OutputTruncated {
		t.Fatal("OutputTruncated = false, want true")
	}
	if len(result.Output) != jsonOutputLimit {
		t.Fatalf("len(Output) = %d, want %d", len(result.Output), jsonOutputLimit)
	}
	if len(stdout) > jsonOutputLimit+1024 {
		t.Fatalf("JSON output is not capped: %d bytes", len(stdout))
	}
}

func TestPrintJSONOmitsUnavailableOutput(t *testing.T) {
	command := types.NewCommandWithoutOutput("git psu")
	code, stdout, _ := captureOutput(t, func() int { return printJSON(command, nil) })
	if code != 0 {
		t.Fatalf("printJSON() = %d, want 0", code)
	}
	if strings.Contains(stdout, "\"output\"") {
		t.Fatalf("stdout = %q, want the output field to be omitted", stdout)
	}
}

func TestTruncatedOutputHandlesReplacementGrowth(t *testing.T) {
	input := strings.Repeat("\xff", jsonOutputLimit/2) + strings.Repeat("a", jsonOutputLimit)
	output, truncated := truncatedOutput(input)
	if !truncated {
		t.Fatal("truncatedOutput() did not report truncation")
	}
	if len(output) > jsonOutputLimit {
		t.Fatalf("truncatedOutput() length = %d, want at most %d", len(output), jsonOutputLimit)
	}
	if !utf8.ValidString(output) {
		t.Fatal("truncatedOutput() produced invalid UTF-8")
	}
}

func TestTruncateJSONOutputBoundaries(t *testing.T) {
	exact := strings.Repeat("a", jsonOutputLimit)
	if got, truncated := truncateJSONOutput(exact); truncated || got != exact {
		t.Fatalf("truncateJSONOutput(exact limit) truncated=%t, length=%d", truncated, len(got))
	}

	oversized := strings.Repeat("a", jsonOutputLimit+1)
	if got, truncated := truncateJSONOutput(oversized); !truncated || len(got) != jsonOutputLimit {
		t.Fatalf("truncateJSONOutput(oversized) truncated=%t, length=%d", truncated, len(got))
	}

	splitRune := strings.Repeat("a", jsonOutputLimit-1) + "€"
	if got, truncated := truncateJSONOutput(splitRune); !truncated || !utf8.ValidString(got) ||
		len(got) != jsonOutputLimit-1 {
		t.Fatalf("truncateJSONOutput(split rune) truncated=%t, valid=%t, length=%d",
			truncated, utf8.ValidString(got), len(got))
	}
}
