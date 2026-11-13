package entrypoints

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/rules"
)

func TestLoadCustomRules(t *testing.T) {
	resetEntrypointsState(t)
	rulesDir := filepath.Join(conf.UserDir(), "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	valid := `{"name":"test_custom_demo","enabled":true,"match":{"script":"^git demo$"},"commands":["git status"]}`
	files := map[string]string{
		"demo.json":      valid,
		"invalid.json":   "{not json",
		"duplicate.json": valid,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(rulesDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	loadCustomRules()

	if rules.Get("test_custom_demo") == nil {
		t.Fatal("loadCustomRules() did not register the valid custom rule")
	}
	output := buffer.String()
	if !strings.Contains(output, "Custom rule") {
		t.Fatalf("loadCustomRules() output = %q, want custom rule warnings", output)
	}

	// A missing rules directory is not an error.
	if err := os.RemoveAll(rulesDir); err != nil {
		t.Fatal(err)
	}
	buffer.Reset()
	loadCustomRules()
	if buffer.Len() != 0 {
		t.Fatalf("loadCustomRules() with a missing directory warned: %q", buffer.String())
	}
}
