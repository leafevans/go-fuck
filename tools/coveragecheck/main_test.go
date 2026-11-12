package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseTotal(t *testing.T) {
	output := "example.com/a/a.go:1:\tmain\t100.0%\n" +
		"total:\t\t(statements)\t\t82.6%\n"
	percentage, ok := parseTotal(output)
	if !ok {
		t.Fatal("parseTotal() did not find the total line")
	}
	if percentage != 82.6 {
		t.Fatalf("parseTotal() = %v, want 82.6", percentage)
	}
}

func TestParseTotalMissing(t *testing.T) {
	if _, ok := parseTotal("example.com/a/a.go:1:\tmain\t100.0%\n"); ok {
		t.Fatal("parseTotal() should fail without a total line")
	}
}

func TestTotalCoverage(t *testing.T) {
	original := runCoverFunc
	t.Cleanup(func() { runCoverFunc = original })

	runCoverFunc = func(string) (string, error) {
		return "total:\t\t(statements)\t\t75.5%\n", nil
	}
	percentage, err := totalCoverage("coverage.out")
	if err != nil {
		t.Fatalf("totalCoverage() error = %v", err)
	}
	if percentage != 75.5 {
		t.Fatalf("totalCoverage() = %v, want 75.5", percentage)
	}
}

func TestTotalCoverageToolFailure(t *testing.T) {
	original := runCoverFunc
	t.Cleanup(func() { runCoverFunc = original })

	runCoverFunc = func(string) (string, error) {
		return "cover: open coverage.out: no such file", errors.New("exit status 1")
	}
	if _, err := totalCoverage("coverage.out"); err == nil {
		t.Fatal("totalCoverage() should report the tool failure")
	}
}

func stubCoverage(t *testing.T, output string, err error) {
	t.Helper()
	original := runCoverFunc
	t.Cleanup(func() { runCoverFunc = original })
	runCoverFunc = func(string) (string, error) { return output, err }
}

func TestRunReportsCoverage(t *testing.T) {
	stubCoverage(t, "total:\t\t(statements)\t\t85.1%\n", nil)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run(nil, stdout, stderr); code != 0 {
		t.Fatalf("run() = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "85.1%") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunEnforcesMinimum(t *testing.T) {
	stubCoverage(t, "total:\t\t(statements)\t\t80.0%\n", nil)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run([]string{"-min", "84"}, stdout, stderr); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "below the required minimum") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunToolFailure(t *testing.T) {
	stubCoverage(t, "cover: open coverage.out: no such file", errors.New("exit status 1"))
	if code := run(nil, &bytes.Buffer{}, &bytes.Buffer{}); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	if code := run([]string{"-unknown"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}
