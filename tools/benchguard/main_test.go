package main

import (
	"bytes"
	"strings"
	"testing"
)

const benchmarkOutput = `goos: linux
goarch: amd64
pkg: example.com/demo
cpu: Test CPU
BenchmarkFoo-16         	     100	      1234 ns/op	     464 B/op	       5 allocs/op
BenchmarkBar-8          	      50	       999 ns/op	     100 B/op	       2 allocs/op
PASS
ok  	example.com/demo	0.123s
`

func TestParseBenchmarks(t *testing.T) {
	bench, err := parseBenchmarks(strings.NewReader(benchmarkOutput))
	if err != nil {
		t.Fatalf("parseBenchmarks() error = %v", err)
	}
	if len(bench) != 2 {
		t.Fatalf("parsed %d benchmarks, want 2: %+v", len(bench), bench)
	}
	if got := bench["BenchmarkFoo"]; got.bytes != 464 || got.allocs != 5 {
		t.Fatalf("BenchmarkFoo = %+v, want 464 bytes and 5 allocations", got)
	}
	if got := bench["BenchmarkBar"]; got.bytes != 100 || got.allocs != 2 {
		t.Fatalf("BenchmarkBar = %+v, want 100 bytes and 2 allocations", got)
	}
}

func TestParseBenchmarksWithoutResults(t *testing.T) {
	bench, err := parseBenchmarks(strings.NewReader("PASS\n"))
	if err != nil {
		t.Fatalf("parseBenchmarks() error = %v", err)
	}
	if len(bench) != 0 {
		t.Fatalf("parseBenchmarks() = %+v, want no results", bench)
	}
}

func TestParseBaseline(t *testing.T) {
	baseline, err := parseBaseline(strings.NewReader("# comment\n\nBenchmarkFoo 464 5\nBenchmarkBar 100 2\n"))
	if err != nil {
		t.Fatalf("parseBaseline() error = %v", err)
	}
	if len(baseline) != 2 {
		t.Fatalf("parsed %d entries, want 2: %+v", len(baseline), baseline)
	}
	if got := baseline["BenchmarkFoo"]; got.bytes != 464 || got.allocs != 5 {
		t.Fatalf("BenchmarkFoo = %+v, want 464 bytes and 5 allocations", got)
	}
}

func TestParseBaselineRejectsBrokenLines(t *testing.T) {
	if _, err := parseBaseline(strings.NewReader("BenchmarkFoo 464\n")); err == nil {
		t.Fatal("parseBaseline() accepted a line with two fields")
	}
	if _, err := parseBaseline(strings.NewReader("BenchmarkFoo many 5\n")); err == nil {
		t.Fatal("parseBaseline() accepted a non-numeric byte count")
	}
}

func TestCheckPassesWithinTolerance(t *testing.T) {
	baseline := map[string]metrics{"BenchmarkFoo": {bytes: 464, allocs: 5}}
	current := map[string]metrics{"BenchmarkFoo": {bytes: 470, allocs: 5}}
	if problems := check(current, baseline); len(problems) != 0 {
		t.Fatalf("check() = %v, want no problems", problems)
	}
}

func TestCheckReportsRegression(t *testing.T) {
	baseline := map[string]metrics{"BenchmarkFoo": {bytes: 464, allocs: 5}}
	current := map[string]metrics{"BenchmarkFoo": {bytes: 5000, allocs: 60}}
	problems := check(current, baseline)
	if len(problems) != 1 {
		t.Fatalf("check() = %v, want one problem", problems)
	}
	if !strings.Contains(problems[0], "5000 B/op") || !strings.Contains(problems[0], "60 allocs/op") {
		t.Fatalf("problem %q does not name the measured values", problems[0])
	}
}

func TestCheckReportsMissingBenchmark(t *testing.T) {
	baseline := map[string]metrics{"BenchmarkFoo": {bytes: 464, allocs: 5}}
	problems := check(map[string]metrics{}, baseline)
	if len(problems) != 1 || !strings.Contains(problems[0], "missing") {
		t.Fatalf("check() = %v, want a missing benchmark problem", problems)
	}
}

func TestWriteBaseline(t *testing.T) {
	bench := map[string]metrics{
		"BenchmarkZebra": {bytes: 20, allocs: 1},
		"BenchmarkAlpha": {bytes: 10, allocs: 2},
	}
	var buffer bytes.Buffer
	if err := writeBaseline(&buffer, bench); err != nil {
		t.Fatalf("writeBaseline() error = %v", err)
	}
	want := "# benchmark baseline: name, bytes per op, allocations per op\n" +
		"BenchmarkAlpha 10 2\nBenchmarkZebra 20 1\n"
	if buffer.String() != want {
		t.Fatalf("writeBaseline() = %q, want %q", buffer.String(), want)
	}
}
