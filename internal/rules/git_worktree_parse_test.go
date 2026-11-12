package rules

import "testing"

func TestParseWorktreePaths(t *testing.T) {
	output := "worktree C:/repo/main\nHEAD 1234\nbranch refs/heads/main\n\n" +
		"worktree C:/repo/feature\nHEAD 1234\nbranch refs/heads/feature\n\n"
	got := parseWorktreePaths(output)
	if len(got) != 1 || got[0] != "C:/repo/feature" {
		t.Fatalf("parseWorktreePaths() = %q, want the secondary worktree", got)
	}
}
