// Command benchguard compares the allocation metrics of the benchmarks with a
// committed baseline.
//
// The timing of a benchmark depends on the machine, but its allocated bytes and
// objects per operation are deterministic for the same code and Go version.
// The guard therefore compares only those two numbers, which is enough to catch
// a regression such as a comparator that allocates for every comparison.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// metrics are the deterministic benchmark results of one benchmark.
type metrics struct {
	bytes  int64
	allocs int64
}

// benchmarkLine matches the output of "go test -bench . -benchmem". The CPU
// suffix of the benchmark name (-16, -4, ...) differs between machines and is
// not part of the name.
var benchmarkLine = regexp.MustCompile(
	`^(Benchmark\S+?)(?:-\d+)?\s+\d+\s+[0-9.]+ ns/op\s+(\d+) B/op\s+(\d+) allocs/op`)

// parseBenchmarks reads "go test -bench" output.
func parseBenchmarks(reader io.Reader) (map[string]metrics, error) {
	result := map[string]metrics{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		match := benchmarkLine.FindStringSubmatch(strings.TrimSpace(scanner.Text()))
		if match == nil {
			continue
		}
		bytes, err := strconv.ParseInt(match[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse the bytes of %s: %w", match[1], err)
		}
		allocs, err := strconv.ParseInt(match[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse the allocations of %s: %w", match[1], err)
		}
		result[match[1]] = metrics{bytes: bytes, allocs: allocs}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// parseBaseline reads the baseline file: one "name bytes allocs" line per
// benchmark, blank lines and lines starting with # are ignored.
func parseBaseline(reader io.Reader) (map[string]metrics, error) {
	result := map[string]metrics{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid baseline line %q", line)
		}
		bytes, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid baseline bytes in %q", line)
		}
		allocs, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid baseline allocations in %q", line)
		}
		result[fields[0]] = metrics{bytes: bytes, allocs: allocs}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// benchmarkNames returns the sorted names of a result map.
func benchmarkNames(result map[string]metrics) []string {
	names := make([]string, 0, len(result))
	for name := range result {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// allowed returns the limits that still count as the baseline. The small
// absolute allowance keeps rounding and runtime changes from failing the guard,
// while a change that allocates per comparison exceeds it by far.
func allowed(baseline metrics) metrics {
	allocs := baseline.allocs * 105 / 100
	if baseline.allocs > 0 {
		allocs++
	}
	return metrics{
		bytes:  baseline.bytes*105/100 + 128,
		allocs: allocs,
	}
}

// check compares the benchmark results with the baseline and returns one
// message per benchmark that regressed. A benchmark missing from the output is
// reported as well, because a renamed benchmark must update the baseline.
func check(bench, baseline map[string]metrics) []string {
	problems := []string{}
	for _, name := range benchmarkNames(baseline) {
		want := baseline[name]
		got, ok := bench[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: missing from the benchmark output", name))
			continue
		}
		limit := allowed(want)
		if got.bytes > limit.bytes || got.allocs > limit.allocs {
			problems = append(problems, fmt.Sprintf(
				"%s: %d B/op, %d allocs/op; the baseline allows %d B/op, %d allocs/op",
				name, got.bytes, got.allocs, limit.bytes, limit.allocs))
		}
	}
	return problems
}

// writeBaseline writes the benchmark results in the baseline format.
func writeBaseline(writer io.Writer, bench map[string]metrics) error {
	if _, err := fmt.Fprintln(writer, "# benchmark baseline: name, bytes per op, allocations per op"); err != nil {
		return err
	}
	for _, name := range benchmarkNames(bench) {
		if _, err := fmt.Fprintf(writer, "%s %d %d\n", name, bench[name].bytes, bench[name].allocs); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run checks or updates the allocation baseline and returns the process exit
// code. Keeping the CLI separate from main makes its file and error handling
// testable without terminating the test process.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("benchguard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	benchPath := flags.String("bench", "benchmark.txt", "benchmark output of go test -bench")
	baselinePath := flags.String("baseline", "benchmarks/baseline.txt", "baseline file to compare with")
	update := flags.Bool("update", false, "rewrite the baseline from the benchmark output")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	benchFile, err := os.Open(*benchPath)
	if err != nil {
		fmt.Fprintln(stderr, "benchguard:", err)
		return 2
	}
	bench, err := parseBenchmarks(benchFile)
	benchFile.Close()
	if err != nil {
		fmt.Fprintln(stderr, "benchguard:", err)
		return 2
	}
	if len(bench) == 0 {
		fmt.Fprintln(stderr, "benchguard: the benchmark output has no results")
		return 2
	}

	if *update {
		baseline, err := os.Create(*baselinePath)
		if err != nil {
			fmt.Fprintln(stderr, "benchguard:", err)
			return 2
		}
		if err := writeBaseline(baseline, bench); err != nil {
			baseline.Close()
			fmt.Fprintln(stderr, "benchguard:", err)
			return 2
		}
		if err := baseline.Close(); err != nil {
			fmt.Fprintln(stderr, "benchguard:", err)
			return 2
		}
		fmt.Fprintf(stdout, "benchguard: wrote %d benchmarks to %s\n", len(bench), *baselinePath)
		return 0
	}

	baselineFile, err := os.Open(*baselinePath)
	if err != nil {
		fmt.Fprintln(stderr, "benchguard:", err)
		return 2
	}
	baseline, err := parseBaseline(baselineFile)
	baselineFile.Close()
	if err != nil {
		fmt.Fprintln(stderr, "benchguard:", err)
		return 2
	}

	problems := check(bench, baseline)
	if len(problems) > 0 {
		for _, problem := range problems {
			fmt.Fprintln(stderr, "benchguard:", problem)
		}
		return 1
	}
	fmt.Fprintf(stdout, "benchguard: %d benchmarks are within the baseline\n", len(baseline))
	return 0
}
