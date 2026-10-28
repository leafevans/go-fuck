package rules

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// resetTestState resets the process wide state that rules depend on. It must be
// called at the beginning of every rule test.
func resetTestState(t *testing.T) {
	t.Helper()
	conf.Reset()
	shells.Set(shells.NewGeneric())
	utils.ResetCaches()
}

// testCommand returns a command with a known output.
func testCommand(script, output string) *types.Command {
	return types.NewCommand(script, output)
}

// assertMatch checks that the rule matches the command.
func assertMatch(t *testing.T, rule *types.Rule, script, output string) {
	t.Helper()
	if !rule.Match(testCommand(script, output)) {
		t.Errorf("rule %q should match command %q with output %q", rule.Name, script, output)
	}
}

// assertNotMatch checks that the rule does not match the command.
func assertNotMatch(t *testing.T, rule *types.Rule, script, output string) {
	t.Helper()
	if rule.Match(testCommand(script, output)) {
		t.Errorf("rule %q should not match command %q with output %q", rule.Name, script, output)
	}
}

// assertNewCommands checks the corrected commands of the rule.
func assertNewCommands(t *testing.T, rule *types.Rule, script, output string, want []string) {
	t.Helper()
	got, err := rule.GetNewCommand(testCommand(script, output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) != len(want) {
		t.Fatalf("rule %q returned %q, want %q", rule.Name, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rule %q returned %q, want %q", rule.Name, got, want)
		}
	}
}

// fakeExecutables creates executable stubs in a temporary directory and
// prepends the directory to PATH, so that utils.Which finds them.
func fakeExecutables(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			path += ".exe"
		}
		if err := os.WriteFile(path, []byte{}, 0o755); err != nil {
			t.Fatalf("create fake executable %q: %v", name, err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}
