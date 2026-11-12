package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// runCoverFunc runs "go tool cover -func=<profile>" and returns its output. It
// is a variable so that tests can replace it with a stub.
var runCoverFunc = func(profile string) (string, error) {
	output, err := exec.Command("go", "tool", "cover", "-func="+profile).CombinedOutput()
	return string(output), err
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run reports the total coverage and enforces the minimum. It returns the
// process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("coveragecheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profile := flags.String("profile", "coverage.out", "path of the profile written by go test -coverprofile")
	minimum := flags.Float64("min", 0, "minimum total coverage in percent; zero only reports")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	percentage, err := totalCoverage(*profile)
	if err != nil {
		fmt.Fprintln(stderr, "coveragecheck:", err)
		return 1
	}

	fmt.Fprintf(stdout, "total coverage: %.1f%%\n", percentage)
	if *minimum > 0 && percentage < *minimum {
		fmt.Fprintf(stderr, "coveragecheck: %.1f%% is below the required minimum of %.1f%%\n",
			percentage, *minimum)
		return 1
	}
	return 0
}

// totalCoverage returns the total statement coverage of a profile.
func totalCoverage(profile string) (float64, error) {
	output, err := runCoverFunc(profile)
	if err != nil {
		return 0, fmt.Errorf("go tool cover failed: %w (%s)",
			err, strings.TrimSpace(output))
	}

	percentage, ok := parseTotal(output)
	if !ok {
		return 0, errors.New("go tool cover did not report a total coverage line")
	}
	return percentage, nil
}

// parseTotal extracts the percentage from the final "total:" line, for example
//
//	total:		(statements)			82.6%
func parseTotal(output string) (float64, bool) {
	lines := strings.Split(output, "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(lines[index])
		if !strings.HasPrefix(line, "total:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		last := strings.TrimSuffix(fields[len(fields)-1], "%")
		percentage, err := strconv.ParseFloat(last, 64)
		if err != nil {
			continue
		}
		return percentage, true
	}
	return 0, false
}
