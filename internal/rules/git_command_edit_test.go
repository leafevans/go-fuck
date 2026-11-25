package rules

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/shells"
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

func TestGitCommandEditingPreservesGlobalOptions(t *testing.T) {
	resetTestState(t)
	for _, tc := range []struct {
		name   string
		edit   func() (string, bool)
		script string
		want   string
	}{
		{
			name: "branch option",
			edit: func() (string, bool) {
				return gitReplaceArgument("git -C 'repo dir' branch -d feature", "-d", "-D", false)
			},
			script: "git -C 'repo dir' branch -d feature",
			want:   "git -C 'repo dir' branch -D feature",
		},
		{
			name: "nested operation",
			edit: func() (string, bool) {
				return gitReplaceOperation("git -C 'repo dir' stash saev message", "stash", "saev", "save")
			},
			script: "git -C 'repo dir' stash saev message",
			want:   "git -C 'repo dir' stash save message",
		},
		{
			name: "new subcommand",
			edit: func() (string, bool) {
				return gitCommandWithSubcommand("git -C 'repo dir' checkout feature", "checkout", "branch feature")
			},
			script: "git -C 'repo dir' checkout feature",
			want:   "git -C 'repo dir' branch feature",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.edit()
			if !ok || got != tc.want {
				t.Fatalf("edit(%q) = %q, %v; want %q", tc.script, got, ok, tc.want)
			}
		})
	}
	if _, ok := gitReplaceOperation("git remote remove add", "remote", "add", "set-url"); ok {
		t.Fatal("a non-operation argument was replaced")
	}
	if _, ok := gitReplaceSubcommand("git status", "pull", "clone"); ok {
		t.Fatal("a missing subcommand was replaced")
	}
	if _, ok := gitCommandWithSubcommand("git status", "checkout", "branch feature"); ok {
		t.Fatal("a command was built for a missing subcommand")
	}
	if _, _, ok := gitOperationWordSpan("git switch", "stash"); ok {
		t.Fatal("the wrong subcommand produced an operation span")
	}
	if _, _, ok := gitOperationWordSpan("git remote", "remote"); ok {
		t.Fatal("a subcommand without an operation produced a span")
	}
	for _, script := range []string{"", "git status"} {
		if output, err := gitRepositoryOutput(script, "branch", "branch", "--list"); err != nil || output != nil {
			t.Fatalf("gitRepositoryOutput(%q) = %q, %v", script, output, err)
		}
	}
}
