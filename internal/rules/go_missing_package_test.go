package rules

import (
	"os"
	"testing"
)

// stubGoDirectoryEntries replaces the directory lookup of the rule.
func stubGoDirectoryEntries(t *testing.T, names []string) {
	t.Helper()
	original := goDirectoryEntries
	goDirectoryEntries = func(string) ([]string, error) { return names, nil }
	t.Cleanup(func() { goDirectoryEntries = original })
}

// goMissingDirectoryOutput is the go output for a missing package directory.
const goMissingDirectoryOutput = "stat /repo/internal/missng: directory not found\n"

func TestRuleGoMissingPackage(t *testing.T) {
	resetTestState(t)
	rule := Get("go_missing_package")
	if rule == nil {
		t.Fatal("go_missing_package rule is not registered")
	}

	stubGoDirectoryEntries(t, []string{"missing", "rules"})
	assertMatch(t, rule, "go build ./internal/missng", goMissingDirectoryOutput)
	assertMatch(t, rule, "go vet ./internal/missng", goMissingDirectoryOutput)
	assertMatch(t, rule, "go test ./internal/missng",
		"# ./internal/missng\n"+goMissingDirectoryOutput+"FAIL\t./internal/missng [setup failed]\n")
	assertNotMatch(t, rule, "go build ./internal/missng", "")
	assertNotMatch(t, rule, "cargo build ./internal/missng", goMissingDirectoryOutput)
	assertNotMatch(t, rule, "go build ./internal/other", goMissingDirectoryOutput)

	assertNewCommands(t, rule, "go build ./internal/missng", goMissingDirectoryOutput,
		[]string{"go build ./internal/missing"})
}

func TestGoMissingPackageEdges(t *testing.T) {
	resetTestState(t)
	rule := Get("go_missing_package")
	original := goDirectoryEntries
	t.Cleanup(func() { goDirectoryEntries = original })

	assertNotMatch(t, rule, "go build ./internal/missng", "other output")

	// The lookup fails, so the rule cannot suggest anything.
	goDirectoryEntries = func(string) ([]string, error) { return nil, os.ErrNotExist }
	assertNewCommands(t, rule, "go build ./internal/missng", goMissingDirectoryOutput, nil)

	// The only close name is the missing name itself.
	goDirectoryEntries = func(string) ([]string, error) { return []string{"missng"}, nil }
	assertNewCommands(t, rule, "go build ./internal/missng", goMissingDirectoryOutput, nil)

	// The script does not contain the missing directory name.
	goDirectoryEntries = func(string) ([]string, error) { return []string{"missing"}, nil }
	assertNewCommands(t, rule, "go build other", goMissingDirectoryOutput, nil)

	// A "." directory has no useful correction.
	assertNewCommands(t, rule, "go build .", "stat /repo/.: directory not found", nil)
}
