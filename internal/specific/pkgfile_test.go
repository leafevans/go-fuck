package specific

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestParsePkgFile(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		want         []string
	}{
		{"empty", "", []string{}},
		{"whitespace", " \r\n\t\n", []string{}},
		{"packages with blank lines", "extra/make /usr/bin/make\r\n  \r\n\t\ncore/base\n", []string{"extra/make", "core/base"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePkgFile([]byte(tc.output), nil)
			if got.err != nil || !reflect.DeepEqual(got.packages, tc.want) {
				t.Fatalf("parsePkgFile() = %q, %v, want %q, nil", got.packages, got.err, tc.want)
			}
		})
	}

	wantErr := errors.New("pkgfile could not start")
	if got := parsePkgFile(nil, wantErr); !errors.Is(got.err, wantErr) || len(got.packages) != 0 {
		t.Fatalf("parsePkgFile() = %+v, want the original error", got)
	}
}

func TestPkgFileExitHelper(t *testing.T) {
	if os.Getenv("GOFUCK_PKGFILE_EXIT_HELPER") != "1" {
		return
	}
	if os.Getenv("GOFUCK_PKGFILE_STDERR") == "1" {
		fmt.Fprintln(os.Stderr, "package database is unavailable")
	}
	os.Exit(1)
}

func TestParsePkgFileDistinguishesNoMatchFromFailure(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, stderr := range []string{"0", "1"} {
		t.Run("stderr="+stderr, func(t *testing.T) {
			cmd := exec.Command(executable, "-test.run=^TestPkgFileExitHelper$")
			cmd.Env = append(os.Environ(), "GOFUCK_PKGFILE_EXIT_HELPER=1", "GOFUCK_PKGFILE_STDERR="+stderr, "GOCOVERDIR="+t.TempDir())
			output, runErr := cmd.Output()
			var exitErr *exec.ExitError
			if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("helper error = %v, want exit status 1", runErr)
			}
			got := parsePkgFile(output, runErr)
			if stderr == "0" {
				if got.err != nil || len(got.packages) != 0 {
					t.Fatalf("no-match lookup = %+v", got)
				}
			} else if !errors.Is(got.err, runErr) {
				t.Fatalf("lookup error = %v, want %v", got.err, runErr)
			}
			if got := parsePkgFile([]byte("partial"), runErr); !errors.Is(got.err, runErr) {
				t.Fatalf("partial lookup error = %v, want %v", got.err, runErr)
			}
		})
	}
}

func TestGetPkgFileNormalizesCacheKeysAndIsolatesResults(t *testing.T) {
	resetSpecificState(t)
	original := getPkgFile
	t.Cleanup(func() { getPkgFile = original })
	stored := []string{"extra/make"}
	calls := 0
	getPkgFile = func(command string) pkgFileResult {
		calls++
		if command != "make" {
			t.Fatalf("lookup key = %q, want make", command)
		}
		return pkgFileResult{packages: stored}
	}
	for _, command := range []string{"make", " sudo   make --version ", "'make' --help", "sudo\tmake"} {
		got, err := GetPkgFile(command)
		if err != nil || !reflect.DeepEqual(got, stored) {
			t.Fatalf("GetPkgFile(%q) = %q, %v", command, got, err)
		}
		got[0] = "changed"
		if stored[0] != "extra/make" {
			t.Fatal("caller modified the memoized package list")
		}
	}
	before := calls
	for _, command := range []string{"", " \t\n", "sudo", "sudo ''", "''"} {
		if got, err := GetPkgFile(command); err != nil || len(got) != 0 {
			t.Fatalf("GetPkgFile(%q) = %q, %v, want no lookup", command, got, err)
		}
	}
	if calls != before {
		t.Fatal("an empty executable reached pkgfile")
	}
}

func TestLookupPkgFileMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if got := lookupPkgFile("make"); got.err == nil || len(got.packages) != 0 {
		t.Fatalf("lookupPkgFile() = %+v, want an executable lookup error", got)
	}
	if got := lookupPkgFile(""); got.err != nil || len(got.packages) != 0 {
		t.Fatalf("lookupPkgFile(empty) = %+v, want no lookup", got)
	}
}
