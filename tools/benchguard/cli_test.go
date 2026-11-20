package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name     string
		bench    string
		baseline string
		update   bool
		missing  string
		code     int
		stdout   string
		stderr   string
	}{
		{name: "pass", bench: benchmarkOutput, baseline: "BenchmarkFoo 464 5\nBenchmarkBar 100 2\n", stdout: "2 benchmarks are within the baseline"},
		{name: "regression", bench: benchmarkOutput, baseline: "BenchmarkFoo 0 0\n", code: 1, stderr: "BenchmarkFoo: 464 B/op, 5 allocs/op"},
		{name: "missing benchmark", bench: benchmarkOutput, baseline: "BenchmarkMissing 0 0\n", code: 1, stderr: "BenchmarkMissing: missing"},
		{name: "missing input", missing: "bench", code: 2, stderr: "benchguard:"},
		{name: "missing baseline", bench: benchmarkOutput, missing: "baseline", code: 2, stderr: "benchguard:"},
		{name: "invalid baseline", bench: benchmarkOutput, baseline: "BenchmarkFoo invalid 5\n", code: 2, stderr: "invalid baseline bytes"},
		{name: "no results", bench: "PASS\n", code: 2, stderr: "has no results"},
		{name: "overflow", bench: "BenchmarkFoo-8 1 1 ns/op 9223372036854775808 B/op 1 allocs/op\n", code: 2, stderr: "parse the bytes"},
		{name: "update", bench: benchmarkOutput, update: true, missing: "baseline", stdout: "wrote 2 benchmarks"},
		{name: "unwritable baseline", bench: benchmarkOutput, update: true, missing: "parent", code: 2, stderr: "benchguard:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			benchPath, baselinePath := filepath.Join(dir, "bench.txt"), filepath.Join(dir, "baseline.txt")
			if tc.missing != "bench" {
				if err := os.WriteFile(benchPath, []byte(tc.bench), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.missing != "baseline" && tc.missing != "parent" {
				if err := os.WriteFile(baselinePath, []byte(tc.baseline), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.missing == "parent" {
				baselinePath = filepath.Join(dir, "missing", "baseline.txt")
			}
			args := []string{"-bench", benchPath, "-baseline", baselinePath}
			if tc.update {
				args = append(args, "-update")
			}
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != tc.code {
				t.Fatalf("run() = %d, want %d; stderr=%q", code, tc.code, stderr.String())
			}
			for _, stream := range []struct{ got, want string }{{stdout.String(), tc.stdout}, {stderr.String(), tc.stderr}} {
				if (stream.want == "" && stream.got != "") || !strings.Contains(stream.got, stream.want) {
					t.Errorf("stream = %q, want %q", stream.got, stream.want)
				}
			}
			if tc.update && tc.code == 0 {
				data, err := os.ReadFile(baselinePath)
				if err != nil {
					t.Fatal(err)
				}
				got, err := parseBaseline(bytes.NewReader(data))
				want, _ := parseBenchmarks(strings.NewReader(tc.bench))
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("updated baseline = %v, %v, want %v", got, err, want)
				}
			} else if tc.missing == "" {
				data, err := os.ReadFile(baselinePath)
				if err != nil || string(data) != tc.baseline {
					t.Fatalf("checking changed the baseline: %q, %v", data, err)
				}
			}
		})
	}
}

func TestRunFlags(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"-h"}, 0, "Usage of benchguard"},
		{[]string{"-unknown"}, 2, "flag provided but not defined"},
		{[]string{"-update=maybe"}, 2, "invalid boolean value"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(tc.args, &stdout, &stderr); code != tc.code || stdout.Len() != 0 || !strings.Contains(stderr.String(), tc.want) {
				t.Fatalf("run() = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

type failedWriter struct {
	err       error
	remaining int
}

func (w *failedWriter) Write(data []byte) (int, error) {
	if w.remaining == 0 {
		return 0, w.err
	}
	w.remaining--
	return len(data), nil
}

func TestParsersReportReadErrors(t *testing.T) {
	want := errors.New("input read failed")
	for name, parse := range map[string]func(io.Reader) (map[string]metrics, error){"benchmarks": parseBenchmarks, "baseline": parseBaseline} {
		t.Run(name, func(t *testing.T) {
			if _, err := parse(failedReader{want}); !errors.Is(err, want) {
				t.Fatalf("parser error = %v, want %v", err, want)
			}
		})
	}
}

func TestParsersRejectOverflow(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		parse func(io.Reader) (map[string]metrics, error)
		want  string
	}{
		{"benchmark bytes", "BenchmarkFoo-8 1 1 ns/op 9223372036854775808 B/op 1 allocs/op", parseBenchmarks, "parse the bytes"},
		{"benchmark allocations", "BenchmarkFoo-8 1 1 ns/op 1 B/op 9223372036854775808 allocs/op", parseBenchmarks, "parse the allocations"},
		{"baseline allocations", "BenchmarkFoo 1 9223372036854775808", parseBaseline, "invalid baseline allocations"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.parse(strings.NewReader(tc.input)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("parser error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestWriteBaselineReportsWriteErrors(t *testing.T) {
	want := errors.New("baseline write failed")
	for _, writes := range []int{0, 1} {
		writer := &failedWriter{err: want, remaining: writes}
		if err := writeBaseline(writer, map[string]metrics{"BenchmarkFoo": {bytes: 1}}); !errors.Is(err, want) {
			t.Fatalf("writeBaseline() after %d writes: %v, want %v", writes, err, want)
		}
	}
}
