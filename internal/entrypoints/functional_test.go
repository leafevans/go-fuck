package entrypoints

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/utils"
)

// functionalQuotationCommand contains both a single and a double quote, so
// that the quotation_marks rule matches deterministically on Windows and unix.
const functionalQuotationCommand = "echo 'hello" + string(rune(34)) + ""

// functionalCorrectedCommand is the correction produced by quotation_marks.
const functionalCorrectedCommand = "echo " + string(rune(34)) + "hello" + string(rune(34)) + ""

// functionalAlias is the default alias name (TF_ALIAS is cleared by the
// isolation helper).
const functionalAlias = "fuck"

// functionalIsolateEnvironment isolates the config, shell and settings
// environment of a functional test.
func functionalIsolateEnvironment(t *testing.T) string {
	t.Helper()
	configDir := resetEntrypointsState(t)

	// Shell state exported by the wrapping shell must not leak into the test.
	t.Setenv("TF_ALIAS", "")
	t.Setenv("TF_SHELL_ALIASES", "")

	// Settings can also be overridden by GOFUCK_*/THEFUCK_* environment
	// variables. Remove the ones that are set in the developer environment so
	// that the flow is deterministic, and restore them afterwards.
	for _, entry := range os.Environ() {
		name, _, found := strings.Cut(entry, "=")
		if !found || (!strings.HasPrefix(name, "GOFUCK_") && !strings.HasPrefix(name, "THEFUCK_")) {
			continue
		}
		previous, _ := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("unset %q: %v", name, err)
		}
		t.Cleanup(func() {
			if err := os.Setenv(name, previous); err != nil {
				t.Errorf("restore %q: %v", name, err)
			}
		})
	}

	utils.ResetCaches()
	return configDir
}

// functionalRequirePlatformShell skips the test when the system shell used to
// re-run the failed command is not available.
func functionalRequirePlatformShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = "cmd.exe"
		}
		if _, err := exec.LookPath(comspec); err != nil {
			t.Skipf("the Windows command interpreter %q is not available: %v", comspec, err)
		}
		return
	}
	if info, err := os.Stat("/bin/sh"); err != nil || info.IsDir() {
		t.Skipf("/bin/sh is not available: %v", err)
	}
}

func TestFunctionalFixCommand(t *testing.T) {
	functionalIsolateEnvironment(t)
	functionalRequirePlatformShell(t)

	code, stdout, _ := runMain(t, "go-fuck", "-y", "--force-command", functionalQuotationCommand)
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if stdout != functionalCorrectedCommand {
		t.Fatalf("stdout = %q, want %q", stdout, functionalCorrectedCommand)
	}
}

func TestFunctionalNoRuleMatch(t *testing.T) {
	functionalIsolateEnvironment(t)
	functionalRequirePlatformShell(t)

	code, stdout, _ := runMain(t, "go-fuck", "-y", "--force-command", "echo plain")
	if code != 1 {
		t.Fatalf("Main() = %d, want 1", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no corrected command", stdout)
	}
}

func TestFunctionalRepeatMode(t *testing.T) {
	functionalIsolateEnvironment(t)
	functionalRequirePlatformShell(t)

	// -y and -r are mutually exclusive in the argument parser, so the
	// confirmation prompt is disabled through the settings environment instead.
	t.Setenv("GOFUCK_REQUIRE_CONFIRMATION", "false")

	code, stdout, _ := runMain(t, "go-fuck", "-r", "--force-command", functionalQuotationCommand)
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}

	prefix := functionalCorrectedCommand + " || " + functionalAlias +
		" --repeat --force-command "
	want := prefix + shells.Quote(functionalCorrectedCommand)
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if !strings.Contains(stdout, "--repeat") {
		t.Fatalf("stdout = %q, want the repeat suffix", stdout)
	}
}

func TestFunctionalHistoryCommand(t *testing.T) {
	functionalIsolateEnvironment(t)
	functionalRequirePlatformShell(t)

	// The newest history entry is a call to the alias, the previous entry is
	// the failed command that has to be fixed.
	t.Setenv("TF_HISTORY", functionalQuotationCommand+string(rune(10))+functionalAlias)

	code, stdout, _ := runMain(t, "go-fuck", "-y")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if stdout != functionalCorrectedCommand {
		t.Fatalf("stdout = %q, want %q", stdout, functionalCorrectedCommand)
	}
}
