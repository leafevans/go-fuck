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
	if strings.Contains(alias, consts.UserCommandMark) {
		t.Fatalf("generated alias must not embed the raw marker characters: %q", alias)
	}
	for _, want := range []string{
		"__gofuck_mark=$'",
		`PS1="\[${__gofuck_mark}\]$PS1"`,
		"__gofuck_mark_prompt",
		"PROMPT_COMMAND",
		"function fuck",
	} {
		if !strings.Contains(alias, want) {
			t.Fatalf("bash marked alias does not contain %q: %q", want, alias)
		}
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
	if strings.Contains(alias, consts.UserCommandMark) {
		t.Fatalf("generated alias must not embed the raw marker characters: %q", alias)
	}
	for _, want := range []string{
		"__gofuck_mark=$'",
		`PS1="%{${__gofuck_mark}%}$PS1"`,
		"__gofuck_mark_prompt",
		"precmd_functions",
		"fuck () {",
	} {
		if !strings.Contains(alias, want) {
			t.Fatalf("zsh marked alias does not contain %q: %q", want, alias)
		}
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
