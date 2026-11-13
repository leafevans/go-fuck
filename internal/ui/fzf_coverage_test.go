package ui

import (
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func TestRunFzfUnavailable(t *testing.T) {
	t.Setenv("PATH", "")
	utils.ResetCaches()
	t.Cleanup(utils.ResetCaches)
	if _, err := runFzf([]string{"git status"}); err != errFzfUnavailable {
		t.Fatalf("runFzf() error = %v, want errFzfUnavailable", err)
	}
}
