package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// tcshAliasCase runs the tcsh alias once with the given failed command and
// returns the command go-fuck received and the shell output.
func tcshAliasCase(t *testing.T, tcsh, failedCommand string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stub := filepath.Join(binDir, "go-fuck")
	// The stub records the command go-fuck would correct: TF_HISTORY when the
	// alias passes it, the arguments otherwise.
	stubScript := "#!/bin/sh\n" +
		"raw=\"$TF_HISTORY\"\n" +
		"if [ -z \"$raw\" ]; then raw=\"$*\"; fi\n" +
		"printf 'RAW:[%s]\\n' \"$raw\" >> \"$STUB_OUT\"\n" +
		"printf 'echo CORRECTED_OK\\n'\n"
	if err := os.WriteFile(stub, []byte(stubScript), 0o755); err != nil {
		t.Fatal(err)
	}
	// Files that a glob would expand to, which must not reach go-fuck.
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	aliasPath := filepath.Join(dir, "alias.tcsh")
	if err := os.WriteFile(aliasPath, []byte(NewTcsh().AppAlias("fuck")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rawPath := filepath.Join(dir, "raw.txt")
	setup := strings.Join([]string{
		"set history = 50",
		"setenv PATH " + binDir + ":/usr/bin:/bin",
		"setenv STUB_OUT " + rawPath,
		"source " + aliasPath,
	}, "\n") + "\n"

	command := exec.Command(tcsh, "-f", "-i")
	command.Dir = dir
	command.Stdin = strings.NewReader(setup + failedCommand + "\nfuck\nexit\n")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("tcsh failed: %v\n%s", err, output)
	}
	captured, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("the alias did not run go-fuck: %v\n%s", err, output)
	}
	return string(captured), string(output)
}

// TestTcshAliasPassesRawHistory checks that the alias hands the failed command
// to go-fuck exactly as it was typed.
//
// The first version of the alias used an unquoted command substitution, so the
// shell split the command on blanks and expanded globs before go-fuck saw it:
// "false *" arrived as the files of the directory.
func TestTcshAliasPassesRawHistory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the tcsh alias behavior test needs a POSIX host")
	}
	tcsh, err := exec.LookPath("tcsh")
	if err != nil {
		t.Skip("tcsh is not available")
	}
	testCases := []struct {
		name    string
		command string
		want    string
	}{
		{"glob", "false *", "RAW:[false *]\n"},
		{"quotes", "false \"hello world\"", "RAW:[false \"hello world\"]\n"},
		{"single quote", "false 'hello world'", "RAW:[false 'hello world']\n"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			captured, output := tcshAliasCase(t, tcsh, testCase.command)
			if captured != testCase.want {
				t.Fatalf("go-fuck received %q, want %q", captured, testCase.want)
			}
			if !strings.Contains(output, "CORRECTED_OK") {
				t.Fatalf("the corrected command was not executed:\n%s", output)
			}
		})
	}
}

// TestAliasSyntaxTcsh checks that tcsh can parse the generated alias.
func TestAliasSyntaxTcsh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the tcsh syntax check needs a POSIX host")
	}
	tcsh, err := exec.LookPath("tcsh")
	if err != nil {
		t.Skip("tcsh is not available")
	}
	alias := NewTcsh().AppAlias("fuck")
	aliasSyntaxContains(t, "tcsh", alias)
	if !strings.Contains(alias, "TF_HISTORY") {
		t.Fatalf("tcsh AppAlias() does not pass the history through: %q", alias)
	}
	path := filepath.Join(t.TempDir(), "alias.tcsh")
	if err := os.WriteFile(path, []byte(alias+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(tcsh, "-f", "-n", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("tcsh failed to parse the alias: %v\n%s", err, output)
	}
}
