package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/logs"
)

func instantLogFor(t *testing.T, command, commandOutput string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "instant.log")
	content := consts.UserCommandMark + "$ " + command + "\n" + commandOutput + "\n" +
		consts.UserCommandMark + "$ "
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExplainUsesInstantCaptureWithoutExecuting(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv(consts.OutputLogEnv, instantLogFor(t,
		"git brnch", "git: 'brnch' is not a git command. See 'git --help'.\n\n"+
			"The most similar command is\nbranch"))
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)

	code, stdout, _ := runMain(t, "go-fuck", "--explain",
		"--capture-mode", "instant", "--force-command", "git brnch")
	if code != 0 {
		t.Fatalf("Main(--explain) = %d, stdout=%q diagnostics=%q", code, stdout, diagnostics.String())
	}
	if stdout != "" {
		t.Fatalf("explain wrote executable output to stdout: %q", stdout)
	}
	for _, want := range []string{
		"go-fuck explain", "Capture", "instant_log", "Suggestions",
		"git branch", "Rule:", "git_not_command", "Execute:",
	} {
		if !strings.Contains(diagnostics.String(), want) {
			t.Fatalf("explain output %q does not contain %q", diagnostics.String(), want)
		}
	}
}

func TestExplainJSONIncludesCaptureSource(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv(consts.OutputLogEnv, instantLogFor(t,
		"git brnch", "git: 'brnch' is not a git command. See 'git --help'.\n\n"+
			"The most similar command is\nbranch"))

	code, stdout, stderr := runMain(t, "go-fuck", "--json", "--explain",
		"--capture-mode", "instant", "--force-command", "git brnch")
	if code != 0 || stderr != "" {
		t.Fatalf("Main(--json --explain) = %d, stdout=%q stderr=%q", code, stdout, stderr)
	}
	var result jsonResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if result.CaptureSource != "instant_log" {
		t.Fatalf("CaptureSource = %q, want instant_log", result.CaptureSource)
	}
}
