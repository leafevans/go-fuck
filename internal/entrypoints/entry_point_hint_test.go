package entrypoints

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/logs"
)

// TestMainEntryPointWordHint reports the flag when the word of an information
// command is passed as the command to fix.
func TestMainEntryPointWordHint(t *testing.T) {
	for _, tc := range []struct{ name, word, want string }{
		{"setup", "setup", "run go-fuck --setup to add the alias"},
		{"doctor", "doctor", "run go-fuck --doctor to print"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetEntrypointsState(t)
			t.Setenv(consts.AliasEnv, "")
			var diagnostics bytes.Buffer
			logs.SetOutput(&diagnostics)

			code, stdout, _ := runMain(t, "go-fuck", tc.word)
			if code != 1 {
				t.Fatalf("Main(%q) = %d, want 1", tc.word, code)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(diagnostics.String(), tc.want) {
				t.Fatalf("diagnostics = %q, want %q", diagnostics.String(), tc.want)
			}
		})
	}
}

// TestMainEntryPointWordHintForcedCommand checks the hint for a forced command
// without re-running anything.
func TestMainEntryPointWordHintForcedCommand(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv(consts.AliasEnv, "")
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)

	code, stdout, _ := runMain(t, "go-fuck", "--no-rerun", "--force-command", "doctor")
	if code != 1 {
		t.Fatalf("Main(--force-command doctor) = %d, want 1", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(diagnostics.String(), "run go-fuck --doctor") {
		t.Fatalf("diagnostics = %q, want the doctor flag", diagnostics.String())
	}
}

// TestMainEntryPointWordThroughAliasIsCorrection keeps the alias semantics: the
// alias passes its arguments as the command to fix, so the word is not
// rewritten into a flag.
func TestMainEntryPointWordThroughAliasIsCorrection(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv(consts.AliasEnv, "fuck")
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)

	code, stdout, _ := runMain(t, "go-fuck", "--no-rerun", "--force-command", "doctor")
	if code != 1 {
		t.Fatalf("Main(--force-command doctor) through the alias = %d, want 1", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if strings.Contains(diagnostics.String(), "--doctor") {
		t.Fatalf("diagnostics = %q, want no flag hint", diagnostics.String())
	}
	if !strings.Contains(diagnostics.String(), "No suggestion") {
		t.Fatalf("diagnostics = %q, want the plain failure", diagnostics.String())
	}
}
