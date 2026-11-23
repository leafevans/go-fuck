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
	for _, want := range []string{"go-fuck help", "Usage", "Description", "Options", "-h, --help", "-v, --version"} {
		if !strings.Contains(help, want) {
			t.Fatalf("plain help does not contain %q:\n%s", want, help)
		}
	}
	for _, line := range strings.Split(help, "\n") {
		if !strings.Contains(line, "-h, --help") {
			continue
		}
		if column := strings.Index(line, "show this help and exit"); column != 2+helpOptionWidth {
			t.Fatalf("plain help description starts at column %d, want %d:\n%s",
				column, 2+helpOptionWidth, help)
		}
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
	for _, want := range []string{"go-fuck", "help", "Usage", "Options", "-v, --version", "Magnificent app"} {
		if !strings.Contains(help, want) {
			t.Fatalf("styled help does not contain %q:\n%s", want, help)
		}
	}
	if !strings.Contains(help, "show version information and exit") {
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
