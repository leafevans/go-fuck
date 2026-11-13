//go:build linux || darwin

package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// runBashScript writes a shell script and runs it with bash.
func runBashScript(t *testing.T, script string) string {
	t.Helper()
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash is not available: %v", err)
	}
	path := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write script: %v", err)
	}
	out, err := exec.Command(bash, "--noprofile", "--norc", path).CombinedOutput()
	if err != nil {
		t.Fatalf("bash failed: %v\n%s", err, out)
	}
	return string(out)
}

// TestBashPromptHookKeepsMarker checks that a prompt function that rebuilds
// PS1 does not remove the instant mode marker, because the hook re-adds it.
func TestBashPromptHookKeepsMarker(t *testing.T) {
	isolateInstantMode(t)
	t.Setenv(consts.InstantModeEnv, "true")
	alias := NewBash().InstantModeAlias("fuck")

	script := "PS1='$ '\n" +
		"PROMPT_COMMAND='PS1=\"$ \"'\n" +
		alias + "\n" +
		"eval \"$PROMPT_COMMAND\"\n" +
		"printf '%s' \"$PS1\"\n"
	out := runBashScript(t, script)

	if !strings.Contains(out, consts.UserCommandMark+"\r") {
		t.Fatalf("the prompt hook did not restore the marker: %q", out)
	}
	if !strings.Contains(out, "$ ") {
		t.Fatalf("the prompt text was lost: %q", out)
	}
}

// TestZshPromptHookKeepsMarker checks the zsh precmd hook.
func TestZshPromptHookKeepsMarker(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skipf("zsh is not available: %v", err)
	}
	isolateInstantMode(t)
	t.Setenv(consts.InstantModeEnv, "true")
	alias := NewZsh().InstantModeAlias("fuck")

	script := "__reset_prompt() { PS1='$ ' }\n" +
		"precmd_functions=(__reset_prompt)\n" +
		alias + "\n" +
		"for function in $precmd_functions; do $function; done\n" +
		"printf '%s' \"$PS1\"\n"
	path := filepath.Join(t.TempDir(), "script.zsh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write script: %v", err)
	}
	out, err := exec.Command(zsh, "-f", path).CombinedOutput()
	if err != nil {
		t.Fatalf("zsh failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), consts.UserCommandMark+"\r") {
		t.Fatalf("the prompt hook did not restore the marker: %q", out)
	}
}
