package specific

import (
	"strings"
	"testing"
)

// makefileFixture is a small Makefile used by the parser tests.
const makefileFixture = `# build and test targets
.PHONY: build test lint
build:
	go build ./...
test: build
	go test ./...
lint:
	gofmt -l .
clean:
	rm -f go-fuck
VERSION := 1.0
export CGO_ENABLED := 0
%.o: %.c
	cc -c $< -o $@
build test: dependencies
`

func TestParseMakeTargets(t *testing.T) {
	targets := parseMakeTargets(makefileFixture)
	joined := strings.Join(targets, " ")
	for _, want := range []string{"build", "clean", "lint", "test"} {
		if !strings.Contains(" "+joined+" ", " "+want+" ") {
			t.Fatalf("parseMakeTargets() = %q, want it to contain %q", targets, want)
		}
	}
	for _, unwanted := range []string{".PHONY", "VERSION", "CGO_ENABLED", "%.o", "dependencies"} {
		if strings.Contains(" "+joined+" ", " "+unwanted+" ") {
			t.Fatalf("parseMakeTargets() = %q, want it to exclude %q", targets, unwanted)
		}
	}
}

func TestParseMakeTargetsEmpty(t *testing.T) {
	if targets := parseMakeTargets(""); len(targets) != 0 {
		t.Fatalf("parseMakeTargets() = %q, want no targets", targets)
	}
}
