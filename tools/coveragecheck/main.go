// Command coveragecheck reports and optionally enforces the total statement
// coverage of a Go coverage profile.
//
// It delegates the calculation to "go tool cover -func", which is the same
// implementation that produces the numbers developers see locally, and only
// parses its total line. This keeps the CI check in sync with the Go toolchain
// without adding any dependency to the project.
//
// Usage:
//
//	go run ./tools/coveragecheck -profile coverage.out [-min 80]
package main

import (
	"errors"
	"flag"
	"fmt"
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
	profile := flag.String("profile", "coverage.out", "path of the profile written by go test -coverprofile")
	minimum := flag.Float64("min", 0, "minimum total coverage in percent; zero only reports")
	flag.Parse()

	percentage, err := totalCoverage(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coveragecheck:", err)
		os.Exit(1)
	}

	fmt.Printf("total coverage: %.1f%%\n", percentage)
	if *minimum > 0 && percentage < *minimum {
		fmt.Fprintf(os.Stderr, "coveragecheck: %.1f%% is below the required minimum of %.1f%%\n",
			percentage, *minimum)
		os.Exit(1)
	}
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
			return 0, false
		}
		return percentage, true
	}
	return 0, false
}
