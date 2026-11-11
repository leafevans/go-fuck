package entrypoints

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/instant"
)

func TestDoctorInstantCheck(t *testing.T) {
	if !instant.Supported() {
		t.Skip("instant mode is not supported on this platform")
	}
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	if got := doctorInstantCheck(); got.Status != doctorInfo || !strings.Contains(got.Message, "not enabled") {
		t.Fatalf("disabled instant check = %+v", got)
	}

	missing := filepath.Join(t.TempDir(), "missing.log")
	t.Setenv(consts.OutputLogEnv, missing)
	if got := doctorInstantCheck(); got.Status != doctorWarn {
		t.Fatalf("missing log check = %+v, want warn", got)
	}

	empty := filepath.Join(t.TempDir(), "empty.log")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	t.Setenv(consts.OutputLogEnv, empty)
	if got := doctorInstantCheck(); got.Status != doctorInfo || !strings.Contains(got.Message, "no output yet") {
		t.Fatalf("empty log check = %+v", got)
	}

	withoutMarker := filepath.Join(t.TempDir(), "no-marker.log")
	if err := os.WriteFile(withoutMarker, []byte("some output\n"), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	t.Setenv(consts.OutputLogEnv, withoutMarker)
	if got := doctorInstantCheck(); got.Status != doctorWarn || !strings.Contains(got.Message, "prompt marker") {
		t.Fatalf("unmarked log check = %+v, want warn", got)
	}

	valid := filepath.Join(t.TempDir(), "valid.log")
	if err := os.WriteFile(valid, []byte(consts.UserCommandMark+"PS> "), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	t.Setenv(consts.OutputLogEnv, valid)
	if got := doctorInstantCheck(); got.Status != doctorOK {
		t.Fatalf("valid log check = %+v, want ok", got)
	}
}
