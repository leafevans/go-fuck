package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run checks the formatting of the given paths with gofmt and returns the
// process exit code.
func run(paths []string, stdout, stderr io.Writer) int {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	cmd := exec.Command("gofmt", append([]string{"-s", "-l"}, paths...)...)
	var output, commandError bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &commandError

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(stderr, "checkfmt: gofmt was not found on PATH; install Go or add its bin directory to PATH")
			return 1
		}
		fmt.Fprintf(stderr, "checkfmt: gofmt failed: %v\n%s", err, commandError.String())
		return 1
	}

	unformatted := strings.TrimSpace(output.String())
	if unformatted == "" {
		fmt.Fprintln(stdout, "gofmt: all Go files are formatted")
		return 0
	}
	fmt.Fprintln(stderr, "gofmt: the following files are not formatted; run \"make fmt\":")
	fmt.Fprintln(stderr, unformatted)
	return 1
}
