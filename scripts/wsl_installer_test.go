package scripts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// wslInstallerScript extracts the here-string that performs the WSL fallback
// install. Keeping the production script inline avoids a third file that users
// must install, while these tests execute the exact shipped text.
func wslInstallerScript(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("update-go-fuck.ps1")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	start := strings.Index(text, "$installScript = @'\n")
	if start < 0 {
		t.Fatal("WSL installer here-string was not found")
	}
	start += len("$installScript = @'\n")
	end := strings.Index(text[start:], "\n'@")
	if end < 0 {
		t.Fatal("WSL installer here-string terminator was not found")
	}
	return []byte(text[start:start+end] + "\n")
}

func installerArchive(t *testing.T, body []byte, symlink bool) []byte {
	t.Helper()
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	header := &tar.Header{Name: "go-fuck", Mode: 0o755, Size: int64(len(body))}
	if symlink {
		header.Typeflag, header.Linkname, header.Size = tar.TypeSymlink, "old", 0
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if header.Size != 0 {
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}

func TestWslFallbackInstaller(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the extracted installer runs in the Linux WSL environment")
	}
	versionScript := []byte("#!/bin/sh\nprintf 'go-fuck " + fixtureVersion + " using Go test\\n'\n")
	wrongVersionScript := []byte("#!/bin/sh\nprintf 'go-fuck 1.2.3 using Go test\\n'\n")

	for _, tc := range []struct {
		name      string
		body      []byte
		archive   []byte
		symlink   bool
		hash      string
		failMV    bool
		wantError string
	}{
		{name: "success", body: versionScript},
		{name: "checksum-mismatch", body: versionScript, hash: strings.Repeat("0", 64), wantError: "FAILED"},
		{name: "corrupt-archive", body: versionScript, archive: []byte("not an archive"), wantError: "tar"},
		{name: "wrong-version", body: wrongVersionScript, wantError: "unexpected version"},
		{name: "symlink", symlink: true, wantError: "regular go-fuck binary"},
		{name: "mv-failure", body: versionScript, failMV: true, wantError: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			installDir := filepath.Join(home, ".local", "bin")
			if err := os.MkdirAll(installDir, 0o755); err != nil {
				t.Fatal(err)
			}
			oldPath := filepath.Join(installDir, "go-fuck")
			writeFixture(t, oldPath, []byte("old binary"), 0o755)

			archive := tc.archive
			if archive == nil {
				archive = installerArchive(t, tc.body, tc.symlink)
			}
			archivePath := filepath.Join(root, "release.tar.gz")
			writeFixture(t, archivePath, archive, 0o644)
			hash := tc.hash
			if hash == "" {
				hash = fmt.Sprintf("%x", sha256.Sum256(archive))
			}

			path := os.Getenv("PATH")
			if tc.failMV {
				bin := filepath.Join(root, "bin")
				if err := os.Mkdir(bin, 0o755); err != nil {
					t.Fatal(err)
				}
				writeFixture(t, filepath.Join(bin, "mv"), []byte("#!/bin/sh\nexit 1\n"), 0o755)
				path = bin + string(os.PathListSeparator) + path
			}
			cmd := exec.Command("/bin/sh", "-s", archivePath, hash, fixtureVersion)
			cmd.Stdin = bytes.NewReader(wslInstallerScript(t))
			cmd.Env = append(os.Environ(), "HOME="+home, "PATH="+path)
			output, err := cmd.CombinedOutput()
			if (err != nil) != (tc.wantError != "" || tc.failMV) {
				t.Fatalf("installer error = %v, want failure %v\n%s", err, tc.wantError != "" || tc.failMV, output)
			}
			if tc.wantError != "" && !strings.Contains(string(output), tc.wantError) {
				t.Errorf("missing diagnostic %q:\n%s", tc.wantError, output)
			}
			installed, err := os.ReadFile(oldPath)
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "success" {
				if !bytes.Equal(installed, versionScript) {
					t.Fatalf("installed binary = %q, want the release binary", installed)
				}
			} else if string(installed) != "old binary" {
				t.Fatalf("failed install replaced the old binary with %q", installed)
			}
			entries, err := os.ReadDir(installDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Errorf("staging files were not cleaned up: %v", entries)
			}
		})
	}
}
