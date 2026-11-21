package rules

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

func TestGitRepositoryOutputUsesSelectedRepository(t *testing.T) {
	resetTestState(t)
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", dir},
		{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--allow-empty", "-m", "initial"},
		{"-C", dir, "branch", "quality-check"},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	script := "git -C " + shells.Current().Quote(dir) + " branch -d quality-chec"
	output, err := gitRepositoryOutput(script, "branch", "branch", "--list", "--format=%(refname:short)")
	if err != nil || !strings.Contains(string(output), "quality-check") {
		t.Fatalf("repository query = %q, %v", output, err)
	}
	if output, err := gitRepositoryOutput("git --help branch -d x", "branch", "branch", "--list"); err != nil || len(output) != 0 {
		t.Fatalf("unsafe global option produced %q, %v", output, err)
	}
}
