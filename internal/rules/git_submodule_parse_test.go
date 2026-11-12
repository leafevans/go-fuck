package rules

import "testing"

func TestParseSubmodulePaths(t *testing.T) {
	output := " 1234abcd vendor/foo (v1.0.0)\n-5678cdef vendor/bar\n"
	got := parseSubmodulePaths(output)
	if len(got) != 2 || got[0] != "vendor/foo" || got[1] != "vendor/bar" {
		t.Fatalf("parseSubmodulePaths() = %q", got)
	}
}
