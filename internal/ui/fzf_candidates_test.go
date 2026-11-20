package ui

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

func fzfEffectOne(*types.Command, string) error { return nil }
func fzfEffectTwo(*types.Command, string) error { return nil }

func TestFzfCandidatesPreserveCommandIdentity(t *testing.T) {
	commands := []*types.CorrectedCommand{
		{Script: "git status"},
		{Script: "git status", SideEffect: fzfEffectOne},
		{Script: "git status", SideEffect: fzfEffectTwo},
		{Script: "git status", SideEffect: fzfEffectOne},
		{Script: "git status  (+side effect)"},
		{Script: "git status  (+side effect)  [2]"},
		{Script: "echo 'one\ntwo'"},
		{Script: `echo 'one\ntwo'`},
		{Script: "echo \x1b[2J\r\t"},
	}
	labels, byLabel := fzfCandidates(commands)
	want := []*types.CorrectedCommand{commands[0], commands[1], commands[2], commands[4], commands[5], commands[6], commands[7], commands[8]}
	if len(labels) != len(want) || len(byLabel) != len(want) {
		t.Fatalf("candidates lost distinct commands: %q", labels)
	}
	for i, label := range labels {
		if byLabel[label] != want[i] {
			t.Fatalf("label %q selects the wrong command", label)
		}
		if strings.ContainsAny(label, "\r\n\t\x1b") {
			t.Fatalf("unsafe label: %q", label)
		}
		if want[i].SideEffect != nil && !strings.Contains(label, "(+side effect)") {
			t.Fatalf("missing side-effect marker: %q", label)
		}
	}
}

func TestSelectWithFzfPreservesMultilineScriptAndSideEffect(t *testing.T) {
	resetUIState(t)
	conf.Current().Fzf = true
	commands := []*types.CorrectedCommand{
		{Script: "echo 'one\ntwo'"},
		{Script: "echo 'one\ntwo'", SideEffect: fzfEffectOne},
	}
	withFzfRunner(t, func(labels []string) (string, error) {
		if len(labels) != 2 {
			t.Fatalf("distinct side effects were merged: %q", labels)
		}
		return labels[1], nil
	})
	if selected := SelectCommand(commands); selected != commands[1] {
		t.Fatalf("selection did not preserve the original command: %v", selected)
	}
}
