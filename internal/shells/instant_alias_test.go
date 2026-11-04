package shells

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// isolateInstantMode clears the instant mode environment for one test.
func isolateInstantMode(t *testing.T) {
	t.Helper()
	t.Setenv(consts.InstantModeEnv, "")
	t.Setenv(consts.LegacyInstantModeEnv, "")
}

func TestBashInstantModeInitAlias(t *testing.T) {
	isolateInstantMode(t)
	alias := NewBash().InstantModeAlias("fuck")
	for _, want := range []string{"--shell-logger", "GOFUCK_INSTANT_MODE=True", "GOFUCK_OUTPUT_LOG=", "rm -f", "exit"} {
		if !strings.Contains(alias, want) {
			t.Fatalf("instant mode alias does not contain %q: %q", want, alias)
		}
	}
	if strings.Contains(alias, "function fuck") {
		t.Fatalf("initialization code should not define the alias: %q", alias)
	}
}

func TestBashInstantModeMarkedAlias(t *testing.T) {
	isolateInstantMode(t)
	t.Setenv(consts.InstantModeEnv, "true")
	alias := NewBash().InstantModeAlias("fuck")
	if !strings.Contains(alias, consts.UserCommandMark) {
		t.Fatalf("logged shell alias does not mark the prompt: %q", alias)
	}
	wantMarked := "PS1=\"\\[" + consts.UserCommandMark + "\r\\]$PS1\""
	if !strings.Contains(alias, wantMarked) {
		t.Fatalf("bash prompt marker is not wrapped as a non printing sequence: %q", alias)
	}
	if !strings.Contains(alias, "function fuck") {
		t.Fatalf("logged shell alias is missing the alias function: %q", alias)
	}
}

func TestZshInstantModeInitAlias(t *testing.T) {
	isolateInstantMode(t)
	alias := NewZsh().InstantModeAlias("fuck")
	if !strings.Contains(alias, "--shell-logger") || !strings.Contains(alias, "GOFUCK_INSTANT_MODE=True") {
		t.Fatalf("instant mode alias is incomplete: %q", alias)
	}
}

func TestZshInstantModeMarkedAlias(t *testing.T) {
	isolateInstantMode(t)
	t.Setenv(consts.LegacyInstantModeEnv, "true")
	alias := NewZsh().InstantModeAlias("fuck")
	if !strings.Contains(alias, consts.UserCommandMark) || !strings.Contains(alias, "fuck () {") {
		t.Fatalf("logged shell alias is incomplete: %q", alias)
	}
	wantMarked := "PS1=\"%{" + consts.UserCommandMark + "\r%}$PS1\""
	if !strings.Contains(alias, wantMarked) {
		t.Fatalf("zsh prompt marker is not wrapped as a non printing sequence: %q", alias)
	}
}

func TestInstantModeLogPathIsUnique(t *testing.T) {
	isolateInstantMode(t)
	first := NewBash().InstantModeAlias("fuck")
	second := NewBash().InstantModeAlias("fuck")
	if first == second {
		t.Fatal("every instant mode session must use a fresh log path")
	}
}
