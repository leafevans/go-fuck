package system

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// fakeSystemExecutable creates an executable stub in dir and returns its path.
func fakeSystemExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	if err := os.WriteFile(path, []byte{}, 0o755); err != nil {
		t.Fatalf("create fake executable %q: %v", path, err)
	}
	return path
}

func TestWhich(t *testing.T) {
	dir := t.TempDir()
	path := fakeSystemExecutable(t, dir, "go-fuck-test-tool")
	t.Setenv("PATH", dir)

	if got := Which("go-fuck-test-tool"); got != path {
		t.Fatalf("Which() = %q, want %q", got, path)
	}
	if got := Which("go-fuck-missing-tool"); got != "" {
		t.Fatalf("Which() = %q for a missing executable, want an empty string", got)
	}
	if got := Which(""); got != "" {
		t.Fatalf("Which(\"\") = %q, want an empty string", got)
	}
}

func TestExpandHome(t *testing.T) {
	home := t.TempDir()
	// os.UserHomeDir reads HOME on Unix like systems and USERPROFILE on
	// Windows.
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if got := ExpandHome("~"); got != home {
		t.Fatalf("ExpandHome(\"~\") = %q, want %q", got, home)
	}

	wantSub := filepath.Join(home, "sub")
	for _, path := range []string{
		"~" + string(filepath.Separator) + "sub",
		"~/sub",
	} {
		if got := ExpandHome(path); got != wantSub {
			t.Fatalf("ExpandHome(%q) = %q, want %q", path, got, wantSub)
		}
	}

	if got := ExpandHome("~other"); got != "~other" {
		t.Fatalf("ExpandHome(\"~other\") = %q, want it unchanged", got)
	}

	relative := filepath.Join("relative", "path")
	if got := ExpandHome(relative); got != relative {
		t.Fatalf("ExpandHome(%q) = %q, want it unchanged", relative, got)
	}

	absolute := filepath.Join(home, "absolute")
	if got := ExpandHome(absolute); got != absolute {
		t.Fatalf("ExpandHome(%q) = %q, want it unchanged", absolute, got)
	}
}

func TestOpenCommand(t *testing.T) {
	const argument = "https://example.com/report page"

	// An empty PATH makes the lookup of optional helpers deterministic.
	t.Setenv("PATH", t.TempDir())

	command := OpenCommand(argument)
	if command == "" {
		t.Fatal("OpenCommand() is empty")
	}
	if !strings.Contains(command, argument) {
		t.Fatalf("OpenCommand() = %q, want it to contain %q", command, argument)
	}
	if runtime.GOOS == "windows" {
		if want := "cmd /c start " + argument; command != want {
			t.Fatalf("OpenCommand() = %q, want %q", command, want)
		}
		return
	}
	if want := "open " + argument; command != want {
		t.Fatalf("OpenCommand() = %q, want %q", command, want)
	}

	// xdg-open is preferred when it is available.
	fakeDir := t.TempDir()
	fakeSystemExecutable(t, fakeDir, "xdg-open")
	t.Setenv("PATH", fakeDir)
	if got, want := OpenCommand(argument), "xdg-open "+argument; got != want {
		t.Fatalf("OpenCommand() = %q, want %q", got, want)
	}
}

func TestInitOutputDoesNotPanic(t *testing.T) {
	InitOutput()
	InitOutput()
}

func TestAncestorNames(t *testing.T) {
	names := AncestorNames()
	if names == nil {
		t.Fatal("AncestorNames() = nil, want a slice")
	}
	if len(names) == 0 {
		t.Fatal("AncestorNames() returned no ancestors")
	}
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			t.Fatalf("AncestorNames() contains an empty process name: %q", names)
		}
	}
}

func TestParseKey(t *testing.T) {
	cases := []struct {
		name    string
		input   []byte
		want    consts.Key
		wantErr bool
	}{
		{name: "up", input: []byte{0x1b, '[', 'A'}, want: consts.KeyUp},
		{name: "down", input: []byte{0x1b, '[', 'B'}, want: consts.KeyDown},
		{name: "unknown escape sequence", input: []byte{0x1b, '[', 'C'}, want: consts.Key("\x1b[C")},
		{name: "escape without bracket", input: []byte{0x1b, 'x'}, want: consts.Key("\x1bx")},
		{name: "escape at end", input: []byte{0x1b}, want: consts.Key("\x1b")},
		{name: "escape sequence cut off", input: []byte{0x1b, '['}, want: consts.Key("\x1b[")},
		{name: "ctrl n", input: []byte{0x0e}, want: consts.KeyCtrlN},
		{name: "ctrl c", input: []byte{0x03}, want: consts.KeyCtrlC},
		{name: "ctrl p", input: []byte{0x10}, want: consts.KeyCtrlP},
		{name: "plain character", input: []byte{'a'}, want: consts.Key("a")},
		{name: "empty input", wantErr: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			index := 0
			read := func() (byte, error) {
				if index >= len(testCase.input) {
					return 0, io.EOF
				}
				value := testCase.input[index]
				index++
				return value, nil
			}

			got, err := parseKey(read)
			if testCase.wantErr {
				if !errors.Is(err, io.EOF) {
					t.Fatalf("parseKey() error = %v, want io.EOF", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseKey() error = %v", err)
			}
			if got != testCase.want {
				t.Fatalf("parseKey() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestStdinByteReader(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		_ = reader.Close()
	})

	if _, err := writer.Write([]byte{'x'}); err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	ch, err := stdinByteReader()
	if err != nil {
		t.Fatalf("stdinByteReader() error = %v", err)
	}
	if ch != 'x' {
		t.Fatalf("stdinByteReader() = %q, want 'x'", ch)
	}

	if _, err := stdinByteReader(); !errors.Is(err, io.EOF) {
		t.Fatalf("stdinByteReader() error = %v, want io.EOF", err)
	}
}
