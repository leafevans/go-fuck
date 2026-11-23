package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestPowershellQuote(t *testing.T) {
	backtick := "\x60"
	testCases := []struct {
		value string
		want  string
	}{
		{"", "''"},
		{"echo hi", "'echo hi'"},
		{"git commit -m 'hello'", "'git commit -m ''hello'''"},
		{"a'b", "'a''b'"},
		{"echo '$HOME' " + backtick + "t", "'echo ''$HOME'' " + backtick + "t'"},
		{"C:\\Users\\example", "'C:\\Users\\example'"},
	}
	shell := NewPowershell()
	for _, testCase := range testCases {
		if got := shell.Quote(testCase.value); got != testCase.want {
			t.Errorf("Quote(%q) = %q, want %q", testCase.value, got, testCase.want)
		}
	}
}

func TestPowershellSplitCommand(t *testing.T) {
	backtick := "\x60"
	testCases := []struct {
		command string
		want    []string
	}{
		{"cat C:\\Windows", []string{"cat", "C:\\Windows"}},
		{"cat C:\\Program Files", []string{"cat", "C:\\Program", "Files"}},
		{"cat \"C:\\Program Files\"", []string{"cat", "C:\\Program Files"}},
		{"cat 'C:\\Program Files'", []string{"cat", "C:\\Program Files"}},
		{"cat 'a''b'", []string{"cat", "a'b"}},
		{"echo " + backtick + "\"quoted" + backtick + "\"", []string{"echo", "\"quoted\""}},
		{"echo a" + backtick + " b", []string{"echo", "a b"}},
		{"echo  one   two", []string{"echo", "one", "two"}},
		{"", nil},
	}
	shell := NewPowershell()
	for _, testCase := range testCases {
		if got := shell.SplitCommand(testCase.command); !reflect.DeepEqual(got, testCase.want) {
			t.Errorf("SplitCommand(%q) = %q, want %q", testCase.command, got, testCase.want)
		}
	}
}

// TestPowershellSplitCommandMatchesPowerShell checks the splitter against the
// argument list that a real PowerShell builds for the same command line.
func TestPowershellSplitCommandMatchesPowerShell(t *testing.T) {
	powershell := aliasSyntaxPowerShellPath()
	if powershell == "" {
		t.Skip("PowerShell is not available")
	}
	aliasSyntaxRequireShell(t, powershell, "-NoProfile", "-Command", "exit 0")

	backtick := "\x60"
	commands := []string{
		"cat C:\\Windows",
		"cat \"C:\\Program Files\"",
		"cat 'C:\\Program Files'",
		"echo a" + backtick + " b",
	}
	lines := []string{
		"function capture { param([Parameter(ValueFromRemainingArguments=$true)]$rest)",
		"  foreach ($item in $rest) { Write-Output ('ARG:' + $item) }",
		"}",
	}
	for index, command := range commands {
		lines = append(lines, "Write-Output 'CMD:"+strconv.Itoa(index)+"'", "capture "+command)
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "split.ps1")
	if err := os.WriteFile(script, []byte(strings.Join(lines, "\r\n")+"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-NoProfile", "-NonInteractive"}
	if strings.HasSuffix(strings.ToLower(powershell), ".exe") {
		arguments = append(arguments, "-ExecutionPolicy", "Bypass")
	}
	arguments = append(arguments, "-File", script)
	output, err := exec.Command(powershell, arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell failed: %v\n%s", err, output)
	}
	actual := map[int][]string{}
	current := -1
	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "CMD:"):
			index, err := strconv.Atoi(strings.TrimPrefix(line, "CMD:"))
			if err != nil {
				t.Fatalf("bad CMD marker %q", line)
			}
			current = index
			actual[current] = []string{}
		case strings.HasPrefix(line, "ARG:") && current >= 0:
			actual[current] = append(actual[current], strings.TrimPrefix(line, "ARG:"))
		}
	}
	shell := NewPowershell()
	for index, command := range commands {
		want := actual[index]
		got := shell.SplitCommand(command)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SplitCommand(%q) = %q, PowerShell built %q", command, got, want)
		}
	}
}

// TestPowershellQuoteRoundTrips checks that the quoted form reaches a real
// PowerShell as one argument with the original text.
func TestPowershellQuoteRoundTrips(t *testing.T) {
	powershell := aliasSyntaxPowerShellPath()
	if powershell == "" {
		t.Skip("PowerShell is not available")
	}
	aliasSyntaxRequireShell(t, powershell, "-NoProfile", "-Command", "exit 0")

	shell := NewPowershell()
	values := []string{
		"git commit -m 'hello'",
		"echo '$HOME' and a backtick",
		`C:\Program Files\example`,
		"a'b\"c",
	}
	lines := []string{
		"function capture { param([Parameter(ValueFromRemainingArguments=$true)]$rest)",
		"  foreach ($item in $rest) { Write-Output ('ARG:' + $item) }",
		"}",
	}
	for _, value := range values {
		lines = append(lines, "capture "+shell.Quote(value))
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "quote.ps1")
	if err := os.WriteFile(script, []byte(strings.Join(lines, "\r\n")+"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-NoProfile", "-NonInteractive"}
	if strings.HasSuffix(strings.ToLower(powershell), ".exe") {
		arguments = append(arguments, "-ExecutionPolicy", "Bypass")
	}
	arguments = append(arguments, "-File", script)
	output, err := exec.Command(powershell, arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell failed: %v\n%s", err, output)
	}
	got := []string{}
	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, "ARG:") {
			continue
		}
		got = append(got, strings.TrimPrefix(line, "ARG:"))
	}
	if !reflect.DeepEqual(got, values) {
		t.Fatalf("PowerShell received %q, want %q", got, values)
	}
}

// TestPowershellChainWithoutCommands covers the empty command list of
// powershellChain.
func TestPowershellChainWithoutCommands(t *testing.T) {
	if got := NewPowershell().And(); got != "" {
		t.Fatalf("And() = %q, want an empty string", got)
	}
	if got := NewPowershell().Or(); got != "" {
		t.Fatalf("Or() = %q, want an empty string", got)
	}
}

// TestQuotePowerShellString covers quotePowerShellString.
func TestQuotePowerShellString(t *testing.T) {
	if got, want := quotePowerShellString("it's"), "'it''s'"; got != want {
		t.Fatalf("quotePowerShellString() = %q, want %q", got, want)
	}
	if got, want := quotePowerShellString("/tmp/go-fuck.log"), "'/tmp/go-fuck.log'"; got != want {
		t.Fatalf("quotePowerShellString() = %q, want %q", got, want)
	}
}
