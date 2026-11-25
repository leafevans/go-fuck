package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/conf"
)

func runFishAlias(t *testing.T, corrected, arguments string) (string, string) {
	t.Helper()
	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish is not available")
	}
	aliasSyntaxRequireShell(t, fish, "--no-config", "-c", "exit 0")
	resetShellsState(t)
	conf.Current().AlterHistory = false
	dir := t.TempDir()
	writeGoFuckStub(t, dir)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GOFUCK_ARGS_FILE", filepath.Join(dir, "args.txt"))
	correction := filepath.Join(dir, "correction.txt")
	if err := os.WriteFile(correction, []byte(corrected), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_STUB_OUTPUT", correction)
	script := filepath.Join(dir, "run.fish")
	if err := os.WriteFile(script, []byte(NewFish().AppAlias("fuck")+"\nfuck "+arguments+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(fish, "--no-config", script).CombinedOutput()
	if err != nil {
		t.Fatalf("fish alias failed: %v\n%s", err, output)
	}
	recorded, err := os.ReadFile(filepath.Join(dir, "args.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(output), string(recorded)
}

func TestFishAliasPassesInformationCommands(t *testing.T) {
	for _, flag := range []string{"--version", "--alias=myfuck", "--shell-logger=/tmp/example.log"} {
		t.Run(flag, func(t *testing.T) {
			output, recorded := runFishAlias(t, "raw-information-marker\n", flag)
			if output != "raw-information-marker\n" || !strings.Contains(recorded, "ARGS:"+flag+"\n") {
				t.Fatalf("information option was not passed through cleanly: output=%q, arguments=%q", output, recorded)
			}
		})
	}
}

func TestFishAliasRunsMultilineCorrection(t *testing.T) {
	for _, script := range []string{
		"printf '%s\\n' fish-first\nprintf '%s\\n' fish-second",
		"printf '%s\\n' 'fish-first\nfish-second'",
	} {
		t.Run(script, func(t *testing.T) {
			output, _ := runFishAlias(t, script, "-y")
			if output != "fish-first\nfish-second\n" {
				t.Fatalf("multi-line correction was changed or cut short: %q", output)
			}
		})
	}
}

func TestFishAliasIgnoresFailedCorrectionOutput(t *testing.T) {
	t.Setenv("GOFUCK_STUB_EXIT_CODE", "1")
	output, _ := runFishAlias(t, "printf '%s\\n' failed-correction-was-executed", "-y")
	if output != "" {
		t.Fatalf("failed correction output was evaluated: %q", output)
	}
}
