package system

import "testing"

// TestExpandHomeWithoutUserHomeDir covers the home directory error of
// ExpandHome.
func TestExpandHomeWithoutUserHomeDir(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if got := ExpandHome("~/config"); got != "~/config" {
		t.Fatalf("ExpandHome() = %q, want the input unchanged", got)
	}
}
