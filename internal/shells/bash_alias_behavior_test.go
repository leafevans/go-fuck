package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// TestBashAliasPassesInformationCommands checks that the bash alias runs
// go-fuck directly for the information commands instead of evaluating their
// output as a corrected command.
func TestBashAliasPassesInformationCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the behavior test needs a POSIX host")
	}
	resetShellsState(t)

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	aliasSyntaxRequireShell(t, bash, "-c", "exit 0")

	dir := t.TempDir()
	stub := filepath.Join(dir, consts.ExecutableName())
	content := "#!/bin/sh\nprintf 'ARG:%s\\n' \"$*\"\n"
	if err := os.WriteFile(stub, []byte(content), 0o755); err != nil {
		t.Fatalf("write the go-fuck stub: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	script := NewBash().AppAlias("fuck") + "\nfuck --list-rules\n"
	output, err := exec.Command(bash, "-c", script).CombinedOutput()
	if err != nil {
		t.Fatalf("bash failed to run the alias: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "ARG:--list-rules") {
		t.Fatalf("bash output = %q, want the go-fuck invocation with --list-rules", output)
	}
}

// TestBashAliasEvaluatesCorrections checks that the regular correction flow
// still evaluates the corrected command printed by go-fuck.
func TestBashAliasEvaluatesCorrections(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the behavior test needs a POSIX host")
	}
	resetShellsState(t)

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	aliasSyntaxRequireShell(t, bash, "-c", "exit 0")

	dir := t.TempDir()
	stub := filepath.Join(dir, consts.ExecutableName())
	content := "#!/bin/sh\nprintf 'echo CORRECTED\\n'\n"
	if err := os.WriteFile(stub, []byte(content), 0o755); err != nil {
		t.Fatalf("write the go-fuck stub: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	script := NewBash().AppAlias("fuck") + "\nfuck\n"
	output, err := exec.Command(bash, "-c", script).CombinedOutput()
	if err != nil {
		t.Fatalf("bash failed to run the alias: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "CORRECTED") {
		t.Fatalf("bash output = %q, want the evaluated correction", output)
	}
}
