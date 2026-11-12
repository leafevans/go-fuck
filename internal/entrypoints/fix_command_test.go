package entrypoints

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/utils"
)

// resetRawCommandState isolates the process state used by getRawCommand.
func resetRawCommandState(t *testing.T, pathDir string) {
	t.Helper()
	t.Setenv("TF_HISTORY", "")
	t.Setenv("TF_ALIAS", "")
	t.Setenv("PATH", pathDir)
	shells.Set(shells.NewGeneric())
	conf.Reset()
	utils.ResetCaches()
	t.Cleanup(func() {
		shells.Set(shells.NewGeneric())
		conf.Reset()
		utils.ResetCaches()
	})
}

// fakeEntrypointExecutable creates an executable stub in dir and returns its
// file name, including the platform specific suffix.
func fakeEntrypointExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	fileName := name
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte{}, 0o755); err != nil {
		t.Fatalf("create fake executable %q: %v", name, err)
	}
	return fileName
}

// assertRawCommand checks the command returned by getRawCommand.
func assertRawCommand(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getRawCommand() = %q, want %q", got, want)
	}
}

func TestWarnUnknownSettings(t *testing.T) {
	resetEntrypointsState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	warnUnknownSettings(&conf.Settings{UnknownKeys: []string{"nope", "rule"}})
	output := buffer.String()
	for _, want := range []string{"nope", "rule", "ignored"} {
		if !strings.Contains(output, want) {
			t.Fatalf("warnUnknownSettings() output = %q, want it to contain %q", output, want)
		}
	}
}

func TestGetRawCommandForceCommand(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	t.Setenv("TF_HISTORY", "command from history\n")

	args := &argumentparser.Args{
		ForceCommand: "git push --force",
		Command:      []string{"ignored"},
	}
	assertRawCommand(t, getRawCommand(args), []string{"git push --force"})
}

func TestGetRawCommandExplicitArguments(t *testing.T) {
	resetRawCommandState(t, t.TempDir())

	args := &argumentparser.Args{Command: []string{"git", "status"}}
	assertRawCommand(t, getRawCommand(args), []string{"git", "status"})
}

func TestGetRawCommandHistoryNewestEntryWins(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	t.Setenv("TF_HISTORY", "git status\nmake test")

	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"make test"})
}

func TestGetRawCommandHistorySkipsAliasCalls(t *testing.T) {
	resetRawCommandState(t, t.TempDir())

	t.Setenv("TF_HISTORY", "git status\nfuck")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	t.Setenv("TF_HISTORY", "git status\nfuck\nfuck -y")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	t.Setenv("TF_HISTORY", "fuck\nfuck -y")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)

	// A custom alias name is skipped as well.
	t.Setenv("TF_ALIAS", "please")
	t.Setenv("TF_HISTORY", "git status\nplease")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})
}

func TestGetRawCommandHistoryAcceptsExecutables(t *testing.T) {
	dir := t.TempDir()
	// The executable name is similar to the alias name, so the history entry
	// is only accepted because it is a known executable.
	fileName := fakeEntrypointExecutable(t, dir, "myaliasx")
	resetRawCommandState(t, dir)

	t.Setenv("TF_ALIAS", "myalias")
	t.Setenv("TF_HISTORY", fileName)
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{fileName})
}

func TestGetRawCommandHistorySkipsExecutableLookalikes(t *testing.T) {
	resetRawCommandState(t, t.TempDir())

	// The entry looks like the alias and is not a known executable, so it is
	// skipped and the history runs out.
	t.Setenv("TF_ALIAS", "myalias")
	t.Setenv("TF_HISTORY", "myaliasx")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}

func TestGetRawCommandMarkedBashHistory(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewBash())

	t.Setenv("TF_HISTORY", rawCommandMarker+"\t if true; then\necho hi\nfi\n\t fuck")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}),
		[]string{"if true; then\necho hi\nfi"})

	t.Setenv("TF_HISTORY", rawCommandMarker+"\t if true; then\n    echo hi\nfi\n\t fuck")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}),
		[]string{"if true; then\n    echo hi\nfi"})

	// The alias call is not always in the history, for example when it is
	// ignored by HISTCONTROL.
	t.Setenv("TF_HISTORY", rawCommandMarker+"\t git status")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	t.Setenv("TF_HISTORY", rawCommandMarker+"\t git status\n\t fuck --yes")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	t.Setenv("TF_HISTORY", rawCommandMarker+"\t fuck")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)

	t.Setenv("TF_HISTORY", rawCommandMarker)
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}

func TestGetRawCommandMarkedZshHistory(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewZsh())

	t.Setenv("TF_HISTORY", rawCommandMarker+"if true; then\necho hi\nfi\n")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}),
		[]string{"if true; then\necho hi\nfi"})

	t.Setenv("TF_HISTORY", rawCommandMarker+"git status")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	t.Setenv("TF_HISTORY", rawCommandMarker+"fuck --yes")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)

	t.Setenv("TF_HISTORY", rawCommandMarker)
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}

func TestGetRawCommandPowerShellMultilineHistory(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewPowershell())

	// A backtick continuation is stored as one history entry with embedded
	// line breaks and a trailing line break.
	t.Setenv("TF_HISTORY", "Get-Content `\nmissing.tx\n")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}),
		[]string{"Get-Content `\nmissing.tx"})

	// A block is stored the same way.
	t.Setenv("TF_HISTORY", "if ($true) {\nWrite-Output hi\n}\n")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}),
		[]string{"if ($true) {\nWrite-Output hi\n}"})

	// A single line entry keeps working, including a trailing line break.
	t.Setenv("TF_HISTORY", "git status\r\n")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})

	// A previous call to the alias itself is not a command to fix.
	t.Setenv("TF_HISTORY", "fuck\n")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}
