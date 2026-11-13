package utils

import (
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

// aliasShell is a shell with fixed aliases.
type aliasShell struct {
	*shells.Generic
	aliases map[string]string
}

// GetAliases returns the fixed alias map.
func (a *aliasShell) GetAliases() map[string]string {
	return a.aliases
}

// TestWhichUsesCache covers the memoized branch of Which.
func TestWhichUsesCache(t *testing.T) {
	resetUtilsState(t)
	t.Setenv("PATH", fakeExecutableDir(t, "go-fuck-test-cached"))

	first := Which("go-fuck-test-cached")
	if first == "" {
		t.Fatal("Which() did not find the executable")
	}
	if second := Which("go-fuck-test-cached"); second != first {
		t.Fatalf("Which() = %q after the cache, want %q", second, first)
	}
}

// TestIsExecutableIncludesAliases covers IsExecutable and the shell alias part
// of GetAllExecutables.
func TestIsExecutableIncludesAliases(t *testing.T) {
	resetUtilsState(t)
	t.Setenv("TF_ALIAS", "fuck")
	t.Setenv("PATH", fakeExecutableDir(t, "go-fuck-test-listed"))
	shells.Set(&aliasShell{
		Generic: shells.NewGeneric(),
		aliases: map[string]string{"fuck": "go-fuck", "ll": "ls -l"},
	})
	ResetCaches()

	if !IsExecutable("go-fuck-test-listed") {
		t.Fatal("IsExecutable() = false for an executable in PATH")
	}
	if IsExecutable("go-fuck-test-missing") {
		t.Fatal("IsExecutable() = true for a missing executable")
	}

	found := map[string]bool{}
	for _, executable := range GetAllExecutables() {
		found[executable] = true
	}
	if !found["ll"] {
		t.Fatalf("GetAllExecutables() = %q, want it to contain the ll alias", GetAllExecutables())
	}
	if found["fuck"] {
		t.Fatalf("GetAllExecutables() = %q, want it to exclude the alias itself", GetAllExecutables())
	}
}
