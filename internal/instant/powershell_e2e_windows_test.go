//go:build windows

package instant

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

func TestPowerShellInstantCorrection(t *testing.T) {
	if !Supported() {
		t.Skip("ConPTY is not available")
	}
	var shells []string
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		shells = append(shells, pwsh)
	}
	legacy := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := os.Stat(legacy); err == nil {
		shells = append(shells, legacy)
	}
	if len(shells) == 0 {
		t.Skip("PowerShell is not available")
	}
	bin := t.TempDir()
	exe := filepath.Join(bin, "go-fuck.exe")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", exe, "../../cmd/go-fuck")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	for _, shell := range shells {
		t.Run(filepath.Base(shell), func(t *testing.T) {
			for _, tc := range []struct{ name, key string }{
				{"accept without replay", "\r"},
				// Use the UI cancel key, not Ctrl+C: ConPTY can deliver
				// Ctrl+C as a console interrupt before raw input is ready.
				{"cancel without execution", "q"},
			} {
				t.Run(tc.name, func(t *testing.T) { testPowerShellInstantCorrection(t, bin, shell, tc.key) })
			}
		})
	}
}

func testPowerShellInstantCorrection(t *testing.T, bin, shell, key string) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TF_SHELL", "powershell")
	t.Setenv("GOFUCK_REQUIRE_CONFIRMATION", "true")
	t.Setenv("GOFUCK_FZF", "false")
	t.Setenv("GOFUCK_CAPTURE_MODE", "auto") // Would replay if the terminal log were missed.
	t.Setenv(consts.InstantModeEnv, "True")
	logPath := filepath.Join(dir, "instant.log")
	t.Setenv(consts.OutputLogEnv, logPath)

	// The rule cannot match unless it sees the error emitted by the original
	// command. Its correction has an independent, observable side effect.
	const rule = `{"name":"e2e_probe","priority":1,"match":{"script":"^Add-Content original.marker","output":"e2e-captured-failure"},"commands":["Set-Content corrected.marker corrected"]}`
	rules := filepath.Join(dir, "go-fuck", "rules")
	if err := os.MkdirAll(rules, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rules, "probe.json"), []byte(rule), 0o600); err != nil {
		t.Fatal(err)
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	// Start the same logged PowerShell that instant initialization uses, with
	// profiles disabled and history confined to this test directory.
	const prompt = "e2e-ps> "
	setup := "Set-PSReadLineOption -HistorySaveStyle SaveNothing; function prompt { '" + prompt + "' }; Invoke-Expression ((go-fuck --alias --capture-mode=instant) -join [Environment]::NewLine)"
	commandLine := quoteWindowsCommandLineArg(shell) + " -NoLogo -NoProfile -NoExit -Command " + syscall.EscapeArg(setup)
	input, keys := io.Pipe()
	defer keys.Close()
	defer input.Close()
	transcript := &powerShellTranscript{changed: make(chan struct{}, 1)}
	finished := make(chan error, 1)
	go func() {
		_, err := runLoggedShell(commandLine, input, io.Discard, io.MultiWriter(NewRingWriter(logFile), transcript))
		finished <- err
		close(finished)
	}()
	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()
	waitFor := func(text string, start int) string {
		t.Helper()
		for {
			out := transcript.text()
			if end := strings.Index(out[start:], text); end >= 0 {
				return out[:start+end+len(text)]
			}
			select {
			case err := <-finished:
				t.Fatalf("PowerShell exited (%v) waiting for %q; transcript: %q", err, text, out)
			case <-timer.C:
				t.Fatalf("timed out waiting for %q; transcript: %q", text, out)
			case <-transcript.changed:
			}
		}
	}
	send := func(text string) {
		t.Helper()
		if _, err := io.WriteString(keys, text); err != nil {
			t.Fatal(err)
		}
	}
	out := waitFor(consts.UserCommandMark+prompt, 0)
	// Generate the diagnostic via throw, not a synthetic terminal log. Split
	// its text so the command echo alone cannot satisfy the output regex.
	send("Add-Content original.marker once; throw ('e2e-captured-' + 'failure')\r")
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	send("fuck\r")
	out = waitFor("ctrl+c", len(out))
	if !strings.Contains(out, "Set-Content corrected.marker corrected") {
		t.Fatalf("confirmation did not show the expected correction: %q", out)
	}
	send(key)
	if key == "q" {
		out = waitFor("Aborted", len(out))
	}
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	// Query the actual logged output for two invocations with identical input.
	send("$sample = 'first'; Write-Output $sample\r")
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	send("Write-Output $sample\r")
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	send("$sample = 'second'\r")
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	send("Write-Output $sample\r")
	out = waitFor(consts.UserCommandMark+prompt, len(out))
	if got, ok := LastOutput(logPath, "Write-Output $sample"); !ok || got != "second" {
		t.Fatalf("latest output = %q, %t; want second", got, ok)
	}
	send("go-fuck --capture-mode=instant --force-command 'Add-Content absent.marker replay'\r")
	out = waitFor("No suggestion without the original command output", len(out))
	waitFor(consts.UserCommandMark+prompt, len(out))
	if _, err := os.Stat("absent.marker"); !os.IsNotExist(err) {
		t.Fatalf("instant miss unexpectedly ran the original: %v", err)
	}
	send("exit 0\r")
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-timer.C:
		t.Fatalf("PowerShell did not exit; transcript: %q", transcript.text())
	}

	if data, err := os.ReadFile("original.marker"); err != nil || strings.Join(strings.Fields(string(data)), " ") != "once" {
		t.Fatalf("original marker = %q (%v), want exactly one execution", data, err)
	}
	data, err := os.ReadFile("corrected.marker")
	if key == "q" {
		if !os.IsNotExist(err) {
			t.Fatalf("cancel left a correction marker: %q (%v)", data, err)
		}
	} else if err != nil || strings.TrimSpace(string(data)) != "corrected" {
		t.Fatalf("accepted correction marker = %q (%v), want corrected", data, err)
	}
}

type powerShellTranscript struct {
	mu      sync.Mutex
	out     strings.Builder
	changed chan struct{}
}

func (s *powerShellTranscript) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.out.Write(p)
	select {
	case s.changed <- struct{}{}:
	default:
	}
	return n, err
}

func (s *powerShellTranscript) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.out.String()
}
