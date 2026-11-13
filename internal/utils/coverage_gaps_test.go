package utils

import (
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
)

// TestExecutableDirectorySkipsEntryPoints covers the entry point skip and the
// executable suffix trimming of the PATH scan.
func TestExecutableDirectorySkipsEntryPoints(t *testing.T) {
	resetUtilsState(t)
	t.Setenv("TF_ALIAS", "fuck")
	t.Setenv("PATH", fakeExecutableDir(t, "go-fuck", "thefuck", "fuck", "real-tool", "tool.exe"))
	ResetCaches()

	found := map[string]bool{}
	for _, name := range GetAllExecutables() {
		found[name] = true
	}
	for _, excluded := range []string{"go-fuck", "thefuck", "fuck"} {
		if found[excluded] {
			t.Fatalf("GetAllExecutables() = %q, want it to exclude the entry point %q",
				GetAllExecutables(), excluded)
		}
	}
	if !found["real-tool"] {
		t.Fatalf("GetAllExecutables() = %q, want it to contain real-tool", GetAllExecutables())
	}
	if !found["tool.exe"] {
		t.Fatalf("GetAllExecutables() = %q, want it to contain tool.exe", GetAllExecutables())
	}
}

// TestGetCloseMatchesWithoutLimit covers the zero limit branch of
// GetCloseMatches.
func TestGetCloseMatchesWithoutLimit(t *testing.T) {
	conf.Reset()
	t.Cleanup(conf.Reset)
	conf.Current().NumCloseMatches = 0

	if got := GetCloseMatches("x", []string{"y"}, 0, 0.1); got != nil {
		t.Fatalf("GetCloseMatches() = %q, want nil", got)
	}
}
