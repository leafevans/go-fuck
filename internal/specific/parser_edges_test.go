package specific

import (
	"slices"
	"testing"
)

func TestParseComposeServicesEdgeCases(t *testing.T) {
	content := "version: \"3\"\r\n" +
		"services:\r\n" +
		"  web:\r\n" +
		"    image: nginx\r\n" +
		"  bad name:\r\n" +
		"    image: invalid\r\n" +
		"  no-colon\r\n" +
		"    image: invalid\r\n" +
		"  api-v1:\r\n" +
		"  # comment\r\n" +
		"volumes:\r\n" +
		"  data:\r\n"
	if got, want := parseComposeServices(content), []string{"api-v1", "web"}; !slices.Equal(got, want) {
		t.Fatalf("parseComposeServices() = %q, want %q", got, want)
	}
}

func TestParseMakeTargetsEdgeCases(t *testing.T) {
	content := ".PHONY: clean test # declared targets\r\n" +
		"build test: dependencies # aliases\r\n" +
		"VERSION = 1.0\r\n" +
		"export CGO_ENABLED := 0\r\n" +
		"%.o: %.c\r\n" +
		"\tcc -c $< -o $@\r\n"
	if got, want := parseMakeTargets(content), []string{"build", "clean", "test"}; !slices.Equal(got, want) {
		t.Fatalf("parseMakeTargets() = %q, want %q", got, want)
	}
}
