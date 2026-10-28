// Command checkfmt reports Go files that are not formatted and exits with a
// non-zero status when it finds any.
//
// It is used by "make fmt-check". The usual shell variant
// "test -z \"$(gofmt -l .)\"" is not portable to cmd.exe, which is the default
// shell of make on Windows.
//
// Usage:
//
//	go run ./tools/checkfmt [PATH...]
//
// PATH defaults to the current directory. gofmt must be on PATH.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{"."}
	}

	cmd := exec.Command("gofmt", append([]string{"-s", "-l"}, paths...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(os.Stderr, "checkfmt: gofmt was not found on PATH; install Go or add its bin directory to PATH")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "checkfmt: gofmt failed: %v\n%s", err, stderr.String())
		os.Exit(1)
	}

	unformatted := strings.TrimSpace(stdout.String())
	if unformatted == "" {
		fmt.Println("gofmt: all Go files are formatted")
		return
	}
	fmt.Fprintln(os.Stderr, "gofmt: the following files are not formatted; run \"make fmt\":")
	fmt.Fprintln(os.Stderr, unformatted)
	os.Exit(1)
}
