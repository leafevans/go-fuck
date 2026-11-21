package rules

import "testing"

func TestParseSubmodulePaths(t *testing.T) {
	output := "160000 1234abcd 0\tvendor/foo\x00100644 abc12345 0\tfile\x00160000 5678cdef 0\tvendor/bar baz\x00"
	got := parseSubmodulePaths(output)
	if len(got) != 2 || got[0] != "vendor/foo" || got[1] != "vendor/bar baz" {
		t.Fatalf("parseSubmodulePaths() = %q", got)
	}
}
