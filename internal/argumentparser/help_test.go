package argumentparser

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
)

func TestPlainHelp(t *testing.T) {
	help := helpText(false)
	if strings.Contains(help, "\x1b[") {
		t.Fatalf("plain help contains escape sequences: %q", help)
	}
	for _, want := range []string{"usage: go-fuck", "Magnificent app", "options:", "-v, --version"} {
		if !strings.Contains(help, want) {
			t.Fatalf("plain help does not contain %q:\n%s", want, help)
		}
	}
	if !strings.Contains(help, "  -v, --version                        show program's version number and exit") {
		t.Fatalf("plain help options are not aligned:\n%s", help)
	}
}

func TestStyledHelp(t *testing.T) {
	conf.Reset()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "truecolor")

	help := helpText(true)
	if !strings.Contains(help, "\x1b[") {
		t.Fatalf("styled help has no escape sequences: %q", help)
	}
	for _, want := range []string{"usage:", "options:", "-v, --version", "Magnificent app"} {
		if !strings.Contains(help, want) {
			t.Fatalf("styled help does not contain %q:\n%s", want, help)
		}
	}
	if !strings.Contains(help, "show program's version number and exit") {
		t.Fatalf("styled help lost an option description:\n%s", help)
	}
}

func TestStyledUsage(t *testing.T) {
	conf.Reset()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "truecolor")

	usage := usageText(true)
	if !strings.Contains(usage, "\x1b[") || !strings.Contains(usage, "usage:") {
		t.Fatalf("styled usage is not styled: %q", usage)
	}
	if plain := usageText(false); strings.Contains(plain, "\x1b[") || !strings.Contains(plain, "usage: go-fuck") {
		t.Fatalf("plain usage is wrong: %q", plain)
	}
}
