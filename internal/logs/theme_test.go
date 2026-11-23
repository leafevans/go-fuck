package logs

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestThemeStyles(t *testing.T) {
	for name, format := range map[string]func(string) string{
		"Accent":  Accent,
		"Muted":   Muted,
		"Strong":  Strong,
		"Success": Success,
		"Failure": Failure,
	} {
		if got := format("value"); !strings.Contains(got, "value") {
			t.Fatalf("%s(value) = %q", name, got)
		}
	}
	_ = ColorsEnabled()
}

func TestTraceWrappedError(t *testing.T) {
	err := fmt.Errorf("outer: %w", errors.New("inner"))
	got := trace(err)
	if !strings.Contains(got, "outer") || !strings.Contains(got, "inner") {
		t.Fatalf("trace() = %q", got)
	}
	if got := trace(nil); got != "<nil>" {
		t.Fatalf("trace(nil) = %q", got)
	}
}
