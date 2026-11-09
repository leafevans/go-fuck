package shells

import (
	"io"
	"os"
	"testing"

	"github.com/leafevans/go-fuck/internal/logs"
)

func TestCurrentAndSet(t *testing.T) {
	previous := Current()
	t.Cleanup(func() { Set(previous) })

	Set(NewBash())
	if got := Current().Name(); got != "bash" {
		t.Fatalf("Current().Name() = %q, want bash", got)
	}
	if got := Current().FriendlyName(); got != "Bash" {
		t.Fatalf("Current().FriendlyName() = %q, want Bash", got)
	}
}

func TestGenericDefaults(t *testing.T) {
	previous := Current()
	t.Cleanup(func() { Set(previous) })
	logs.SetOutput(io.Discard)
	t.Cleanup(func() { logs.SetOutput(os.Stderr) })

	generic := NewGeneric()
	if got := generic.Name(); got != "generic" {
		t.Fatalf("Name() = %q, want generic", got)
	}
	if got := generic.FriendlyName(); got != "Generic Shell" {
		t.Fatalf("FriendlyName() = %q, want Generic Shell", got)
	}
	if configuration := generic.HowToConfigure(); configuration != nil {
		t.Fatalf("HowToConfigure() = %+v, want nil", configuration)
	}
	if got := generic.Quote("a b"); got != "'a b'" {
		t.Fatalf("Quote() = %q", got)
	}
	if alias := generic.InstantModeAlias("fuck"); alias == "" {
		t.Fatalf("InstantModeAlias() is empty")
	}
}
