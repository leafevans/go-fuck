package main

import "testing"

// TestTotalCoverageWithoutTotalLine covers the missing total line branch of
// totalCoverage.
func TestTotalCoverageWithoutTotalLine(t *testing.T) {
	stubCoverage(t, "example.com/a/a.go:1:\tmain\t100.0%\n", nil)
	if _, err := totalCoverage("coverage.out"); err == nil {
		t.Fatal("totalCoverage() error = nil, want the missing total line error")
	}
}

// TestParseTotalSkipsInvalidPercentage covers the invalid percentage branch of
// parseTotal.
func TestParseTotalSkipsInvalidPercentage(t *testing.T) {
	output := "total:\t\t(statements)\t\tnot-a-number%\n"
	if _, ok := parseTotal(output); ok {
		t.Fatal("parseTotal() accepted an invalid percentage")
	}
}
