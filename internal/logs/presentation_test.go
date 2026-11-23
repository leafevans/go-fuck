package logs

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
)

func TestRendererUsesSharedStatusVocabulary(t *testing.T) {
	renderer := Renderer{profile: colorProfileTrueColor}
	for _, status := range []Status{StatusOK, StatusInfo, StatusWarn, StatusFail} {
		line := renderer.Diagnostic(status, "message")
		if !strings.Contains(line, string(status)) || !strings.Contains(line, "message") {
			t.Fatalf("Diagnostic(%q) = %q", status, line)
		}
		if !strings.Contains(line, "\x1b[") {
			t.Fatalf("Diagnostic(%q) has no styles: %q", status, line)
		}
	}
}

func TestRendererDisablesStylesForBuffers(t *testing.T) {
	buffer := resetLogsState(t)
	setBannerEnvironment(t, "xterm-256color", "truecolor", "", "")
	conf.Set(conf.Default())
	renderer := NewRenderer(buffer)

	if renderer.Styled() {
		t.Fatal("NewRenderer(buffer).Styled() = true, want false")
	}
	if line := renderer.Diagnostic(StatusFail, "boom"); strings.Contains(line, "\x1b[") {
		t.Fatalf("buffered diagnostic contains ANSI sequences: %q", line)
	}
}
