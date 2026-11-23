package ui

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/conf"
	"github.com/leafevans/go-fuck/v3/internal/logs"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

func TestExplainCommandsShowsNonExecutingPlan(t *testing.T) {
	conf.Reset()
	t.Cleanup(conf.Reset)
	var diagnostics bytes.Buffer
	logs.SetOutput(&diagnostics)
	t.Cleanup(func() { logs.SetOutput(io.Discard) })

	sideEffect := func(*types.Command, string) error { return nil }
	command := types.NewCommandWithoutOutput("git commit")
	ExplainCommands(command, []*types.CorrectedCommand{{
		Script:                "git commit --amend",
		RuleName:              "git_commit_amend",
		Reason:                "The commit failed before creating a commit.",
		SideEffect:            sideEffect,
		SideEffectDescription: "stage tracked changes",
	}})

	output := diagnostics.String()
	for _, want := range []string{
		"go-fuck explain", "git commit", "capture", "none",
		"git_commit_amend", "Before:", "stage tracked changes",
		"Execute:", "git commit --amend",
	} {
		if !strings.Contains(strings.ToLower(output), strings.ToLower(want)) {
			t.Fatalf("explain output %q does not contain %q", output, want)
		}
	}
}
