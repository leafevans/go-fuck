package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// quoteRoundTripValues are the values that used to break one shell or another.
var quoteRoundTripValues = []string{
	"a b",
	"a'b",
	"a\"b",
	"a\\b",
	"a\\'b",
	"a!b",
	"$HOME",
	"a*b",
	"line1\nline2",
	"a b'c\"d\\e!f",
}

// shellCommandForQuoteRoundTrip returns the command line that prints the first
// argument of a script file for the given shell.
func shellQuoteCommand(shellName, quoted string) string {
	if shellName == "tcsh" {
		// tcsh has no printf builtin.
		return "/usr/bin/printf '%s' " + quoted + "\n"
	}
	return "printf '%s' " + quoted + "\n"
}

// runQuoteRoundTrip checks that the quoted form reaches the shell as exactly
// one argument with the original bytes.
func runQuoteRoundTrip(t *testing.T, shellName string, quote func(string) string, program string, args ...string) {
	t.Helper()
	dir := t.TempDir()
	for index, value := range quoteRoundTripValues {
		command := shellQuoteCommand(shellName, quote(value))
		script := filepath.Join(dir, shellName+"-"+string(rune('a'+index))+".txt")
		if err := os.WriteFile(script, []byte(command), 0o644); err != nil {
			t.Fatal(err)
		}
		commandArgs := append(append([]string{}, args...), script)
		output, err := exec.Command(program, commandArgs...).Output()
		if err != nil {
			t.Fatalf("%s could not run %q: %v", shellName, value, err)
		}
		if string(output) != value {
			t.Errorf("%s quote round trip for %q returned %q", shellName, value, string(output))
		}
	}
}

func shellPath(t *testing.T, names ...string) string {
	t.Helper()
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func TestQuoteRoundTripBash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the round trip test needs a POSIX host")
	}
	program := shellPath(t, "bash")
	if program == "" {
		t.Skip("bash is not available")
	}
	runQuoteRoundTrip(t, "bash", NewBash().Quote, program, "--norc")
}

func TestQuoteRoundTripZsh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the round trip test needs a POSIX host")
	}
	program := shellPath(t, "zsh")
	if program == "" {
		t.Skip("zsh is not available")
	}
	runQuoteRoundTrip(t, "zsh", NewZsh().Quote, program, "-f")
}

func TestQuoteRoundTripFish(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the round trip test needs a POSIX host")
	}
	program := shellPath(t, "fish")
	if program == "" {
		t.Skip("fish is not available")
	}
	runQuoteRoundTrip(t, "fish", NewFish().Quote, program, "--no-config")
}

func TestQuoteRoundTripTcsh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the round trip test needs a POSIX host")
	}
	program := shellPath(t, "tcsh")
	if program == "" {
		t.Skip("tcsh is not available")
	}
	runQuoteRoundTrip(t, "tcsh", NewTcsh().Quote, program, "-f")
}
func TestFishQuote(t *testing.T) {
	testCases := []struct {
		value string
		want  string
	}{
		{"", "''"},
		{"echo hi", "'echo hi'"},
		{`a'b`, `'a\'b'`},
		{`a!b`, `'a!b'`},
	}
	shell := NewFish()
	for _, testCase := range testCases {
		if got := shell.Quote(testCase.value); got != testCase.want {
			t.Errorf("Quote(%q) = %q, want %q", testCase.value, got, testCase.want)
		}
	}
}

func TestTcshQuote(t *testing.T) {
	testCases := []struct {
		value string
		want  string
	}{
		{"", "''"},
		{"echo hi", "'echo hi'"},
		{`a!b`, `'a\!b'`},
		{`a'b`, `'a'\''b'`},
	}
	shell := NewTcsh()
	for _, testCase := range testCases {
		if got := shell.Quote(testCase.value); got != testCase.want {
			t.Errorf("Quote(%q) = %q, want %q", testCase.value, got, testCase.want)
		}
	}
}
