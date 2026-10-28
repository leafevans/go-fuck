package entrypoints

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
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
