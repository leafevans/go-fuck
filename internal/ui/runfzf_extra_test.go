package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// fzfHelperEnv selects the behavior of the test binary when it is executed as a
// fake fzf. The test binary is copied because a shell script cannot be executed
// as a program on Windows.
const fzfHelperEnv = "GOFUCK_UI_FZF_HELPER"

// TestMain lets the test binary act as a fake fzf executable.
func TestMain(m *testing.M) {
	if mode := os.Getenv(fzfHelperEnv); mode != "" {
		os.Exit(runFzfHelper(mode))
	}
	os.Exit(m.Run())
}

// runFzfHelper implements the fake fzf modes.
func runFzfHelper(mode string) int {
	switch mode {
	case "select":
		fmt.Print("git status" + newline)
		return 0
	case "abort":
		return 1
	case "interrupt":
		return 130
	case "error":
		return 3
	}
	return 2
}

// installFakeFzf copies the test binary into a temporary directory under the
// name of the fzf executable and points PATH at it.
func installFakeFzf(t *testing.T) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatalf("read the test binary: %v", err)
	}
	name := "fzf"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o755); err != nil {
		t.Fatalf("write the fake fzf executable: %v", err)
	}
	t.Setenv("PATH", dir)
	utils.ResetCaches()
	t.Cleanup(utils.ResetCaches)
}

// enableColors turns the theme colors on for one test.
func enableColors(t *testing.T) {
	t.Helper()
	conf.Current().NoColors = false
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "")
	t.Setenv("WT_SESSION", "")
}

func TestRunFzfSelection(t *testing.T) {
	resetUIState(t)
	installFakeFzf(t)
	enableColors(t)
	t.Setenv(fzfHelperEnv, "select")

	selected, err := runFzf([]string{"ls", "git status"})
	if err != nil {
		t.Fatalf("runFzf() error = %v", err)
	}
	if selected != "git status" {
		t.Fatalf("runFzf() = %q, want %q", selected, "git status")
	}
}

func TestRunFzfAbortExitCodes(t *testing.T) {
	for _, mode := range []string{"abort", "interrupt"} {
		t.Run(mode, func(t *testing.T) {
			resetUIState(t)
			installFakeFzf(t)
			t.Setenv(fzfHelperEnv, mode)

			if _, err := runFzf([]string{"ls"}); !errors.Is(err, errFzfAborted) {
				t.Fatalf("runFzf() error = %v, want errFzfAborted", err)
			}
		})
	}
}

func TestRunFzfReportsExecutionError(t *testing.T) {
	resetUIState(t)
	installFakeFzf(t)
	t.Setenv(fzfHelperEnv, "error")

	_, err := runFzf([]string{"ls"})
	if err == nil {
		t.Fatal("runFzf() error = nil, want the fzf failure")
	}
	if errors.Is(err, errFzfAborted) {
		t.Fatal("runFzf() reported an abort for a failed fzf run")
	}
}

func TestSelectWithFzfDeduplicatesCandidates(t *testing.T) {
	resetUIState(t)
	conf.Current().Fzf = true
	withFzfRunner(t, func(candidates []string) (string, error) {
		if len(candidates) != 2 || candidates[0] != "ls" || candidates[1] != "cd" {
			t.Fatalf("fzf candidates = %q, want [ls cd]", candidates)
		}
		return "cd", nil
	})

	commands := []*types.CorrectedCommand{
		{Script: "ls"},
		{Script: "cd"},
		{Script: "ls"},
	}
	selected := SelectCommand(commands)
	if selected == nil || selected.Script != "cd" {
		t.Fatalf("SelectCommand() = %v, want cd", selected)
	}
}

func TestSelectCommandPreviousAction(t *testing.T) {
	resetUIState(t)
	setKeys(t, consts.KeyUp, "\n")
	commands := []*types.CorrectedCommand{
		{Script: "ls"},
		{Script: "cd"},
	}
	selected := SelectCommand(commands)
	if selected == nil || selected.Script != "cd" {
		t.Fatalf("SelectCommand() = %v, want cd", selected)
	}
}
