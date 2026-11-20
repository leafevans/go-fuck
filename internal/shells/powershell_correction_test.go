package shells

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPowershellAliasRunsEchoCorrections(t *testing.T) {
	for _, script := range []string{
		"echo 'alias-echo-correction'",
		"echolocation",
	} {
		t.Run(script, func(t *testing.T) {
			resetShellsState(t)
			path := filepath.Join(t.TempDir(), "correction.txt")
			if err := os.WriteFile(path, []byte(script+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GOFUCK_STUB_OUTPUT", path)
			output, _ := runPowershellAliasRaw(t,
				"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'broken' } }; function echolocation { Write-Output 'alias-echo-correction' }",
				"fuck")
			if !strings.Contains(string(output), "alias-echo-correction") {
				t.Fatalf("corrected command was not executed: %q", output)
			}
		})
	}
}

func TestPowershellAliasIgnoresFailedCorrectionOutput(t *testing.T) {
	resetShellsState(t)
	path := filepath.Join(t.TempDir(), "correction.txt")
	if err := os.WriteFile(path, []byte("Write-Output 'failed-correction-was-executed'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_STUB_OUTPUT", path)
	t.Setenv("GOFUCK_STUB_EXIT_CODE", "1")
	output, _ := runPowershellAliasRaw(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'broken' } }",
		"fuck")
	if strings.Contains(string(output), "failed-correction-was-executed") {
		t.Fatalf("failed correction output was evaluated: %q", output)
	}
}

func TestPowershellAliasExportsItsOwnName(t *testing.T) {
	resetShellsState(t)
	path := filepath.Join(t.TempDir(), "correction.txt")
	if err := os.WriteFile(path, []byte("Write-Output \"alias-name:$env:TF_ALIAS\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_STUB_OUTPUT", path)
	t.Setenv("TF_ALIAS", "stale-alias")
	output, _ := runPowershellAliasRaw(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'broken' } }",
		"fuck")
	if !strings.Contains(string(output), "alias-name:fuck") {
		t.Fatalf("TF_ALIAS was not set for the correction: %q", output)
	}
}

func TestPowershellAliasSupportsProfileInterpolation(t *testing.T) {
	resetShellsState(t)
	recorded := runPowershellAlias(t, "", `
$ErrorActionPreference = 'Stop'
$source = (Get-Command fuck).ScriptBlock.ToString() -split '\r?\n'
Invoke-Expression "function fuck {$source}"
fuck --version
`)
	if len(recorded) != 1 || strings.TrimSpace(recorded[0].args) != "--version" {
		t.Fatalf("interpolated alias invocation = %+v, want --version", recorded)
	}
}
