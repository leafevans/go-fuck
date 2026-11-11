package rules

import "testing"

const (
	// goMissingFilePathOutput is the go output for a file argument that the
	// toolchain resolved as a package path.
	goMissingFilePathOutput = "no required module provides package main.g; to add it:\n\tgo get main.g\n"
	// goMissingFileAttributeOutput is the go output for a missing file argument
	// on Windows.
	goMissingFileAttributeOutput = "GetFileAttributesEx mian.go: The system cannot find the file specified.\n"
)

func TestRuleGoMissingFile(t *testing.T) {
	resetTestState(t)
	rule := Get("go_missing_file")
	if rule == nil {
		t.Fatal("go_missing_file rule is not registered")
	}

	stubGoDirectoryEntries(t, []string{"main.go", "notes.txt", "solution.go"})
	assertMatch(t, rule, "go run main.g", goMissingFilePathOutput)
	assertMatch(t, rule, "go build main.g", goMissingFilePathOutput)
	assertMatch(t, rule, "go vet main.g", goMissingFilePathOutput)
	assertMatch(t, rule, "go test -run TestX main.g", goMissingFilePathOutput)
	assertMatch(t, rule, "go run mian.go", goMissingFileAttributeOutput)
	assertMatch(t, rule, "go run .\\mian.go", "GetFileAttributesEx .\\mian.go: The system cannot find the file specified.\n")
	assertNotMatch(t, rule, "go run main.g", "")
	assertNotMatch(t, rule, "cargo run main.g", goMissingFilePathOutput)
	assertNotMatch(t, rule, "go run main.g", "stat C:\\repo\\main.g: directory not found\n")
	assertNotMatch(t, rule, "go run nothere.g",
		"no required module provides package nothere.g; to add it:\n\tgo get nothere.g\n")
	assertMatch(t, rule, "go run main.goo",
		"no required module provides package main.goo; to add it:\n\tgo get main.goo\n")
	assertNotMatch(t, rule, "go build ./...",
		"main.go:3:8: no required module provides package github.com/definitely/not-a-real-package-xyz; to add it:\n"+
			"\tgo get github.com/definitely/not-a-real-package-xyz\n")

	assertNewCommands(t, rule, "go run main.g", goMissingFilePathOutput,
		[]string{"go run main.go"})
	assertNewCommands(t, rule, "go run main.goo",
		"no required module provides package main.goo; to add it:\n\tgo get main.goo\n",
		[]string{"go run main.go"})
	assertNewCommands(t, rule, "go test -run TestX main.g", goMissingFilePathOutput,
		[]string{"go test -run TestX main.go"})
	assertNewCommands(t, rule, "go run mian.go", goMissingFileAttributeOutput,
		[]string{"go run main.go"})
	assertNewCommands(t, rule, "go run .\\mian.go",
		"GetFileAttributesEx .\\mian.go: The system cannot find the file specified.\n",
		[]string{"go run .\\main.go"})
}
