package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadmeContainsAllRules ensures that every bundled rule is documented in
// the README, mirroring the reference implementation's test suite.
func TestReadmeContainsAllRules(t *testing.T) {
	path := filepath.Join("..", "..", "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(data)

	missing := []string{}
	for _, rule := range All() {
		if !strings.Contains(readme, rule.Name) {
			missing = append(missing, rule.Name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("rules missing from README.md: %s", strings.Join(missing, ", "))
	}
}
