//go:build windows

package shells

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

// Explicitly dot-source an isolated initialization script with -NoProfile.
// Changing HOME alone does not isolate every PowerShell profile location.
func TestPowerShellInstantNonInteractiveHosts(t *testing.T) {
	isolateInstantMode(t)
	dir := t.TempDir()
	bin := filepath.Join(dir, "go-fuck.exe")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", bin, "../../cmd/go-fuck").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	t.Setenv("APPDATA", dir)
	t.Setenv("LOCALAPPDATA", dir)
	t.Setenv("TEMP", dir)
	t.Setenv("TMP", dir)
	t.Setenv("TF_SHELL", "powershell")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GOFUCK_CAPTURE_MODE", "instant")
	if err := os.MkdirAll(filepath.Join(dir, "go-fuck"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go-fuck", "settings.json"), []byte("{\"capture_mode\":\"instant\"}"), 0600); err != nil {
		t.Fatal(err)
	}
	initPath := filepath.Join(dir, "init.ps1")
	init := "iex ((& '" + strings.ReplaceAll(bin, "'", "''") + "' --alias --capture-mode=instant) -join [Environment]::NewLine)\n"
	if err := os.WriteFile(initPath, []byte(init), 0600); err != nil {
		t.Fatal(err)
	}
	body := ". '" + strings.ReplaceAll(initPath, "'", "''") + "'; if (-not (Get-Command fuck -CommandType Function -ErrorAction SilentlyContinue)) { exit 91 }; if ($env:GOFUCK_INSTANT_MODE) { exit 92 }; Write-Output 'host-completed'; exit 42"
	script := filepath.Join(dir, "host.ps1")
	if err := os.WriteFile(script, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	words := utf16.Encode([]rune(body))
	bytes := make([]byte, 2*len(words))
	for i, w := range words {
		bytes[2*i], bytes[2*i+1] = byte(w), byte(w>>8)
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	for _, name := range []string{"powershell.exe", "pwsh.exe"} {
		t.Run(name, func(t *testing.T) {
			shell, err := exec.LookPath(name)
			if err != nil {
				t.Skip(err)
			}
			for _, tc := range []struct {
				name string
				args []string
			}{
				{"command", []string{"-Command", body}},
				{"file", []string{"-File", script}},
				{"encoded", []string{"-EncodedCommand", encoded}},
				{"abbreviated", []string{"-c", body}},
				{"noninteractive", []string{"-NonInteractive", "-Command", body}},
				{"noexit-redirected", []string{"-NoExit", "-Command", body}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, shell, append([]string{"-NoLogo", "-NoProfile"}, tc.args...)...)
					cmd.Dir = dir
					cmd.WaitDelay = 2 * time.Second
					// Cancel only this test's process tree, never processes by image name.
					cmd.Cancel = func() error {
						killCtx, killCancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer killCancel()
						_ = exec.CommandContext(killCtx, "taskkill.exe", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run()
						return cmd.Process.Kill()
					}
					out, err := cmd.CombinedOutput()
					if ctx.Err() != nil {
						t.Fatalf("host timed out: %v\n%s", ctx.Err(), out)
					}
					exit, ok := err.(*exec.ExitError)
					if !ok || exit.ExitCode() != 42 || !strings.Contains(string(out), "host-completed") {
						t.Fatalf("host did not preserve execution/exit code: %v\n%s", err, out)
					}
					if logs, _ := filepath.Glob(filepath.Join(dir, "go-fuck-script-log-*")); len(logs) > 0 {
						t.Fatalf("unexpected logs: %v", logs)
					}
				})
			}
		})
	}
}
