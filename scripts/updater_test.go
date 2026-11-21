// Package scripts tests the release updaters without network or user state.
package scripts

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const fixtureVersion = "9.8.7"

func TestMain(m *testing.M) {
	// A copy of this test executable acts as a native wsl.exe. Unlike a
	// PowerShell function mock, it observes the actual bytes on the pipe.
	if os.Getenv("GO_FUCK_UPDATER_STDIN_HELPER") == "1" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// The WSL command removes CR bytes before passing the stream to sh.
		normalized := bytes.ReplaceAll(data, []byte("\r"), nil)
		if !bytes.HasPrefix(normalized, []byte("#!/bin/sh\n")) {
			prefix := normalized[:min(16, len(normalized))]
			fmt.Fprintf(os.Stderr, "WSL script must start with a shebang without a BOM; got %x\n", prefix)
			os.Exit(1)
		}
		log := strings.Join(os.Args[1:], " ") + "\n" + string(data)
		if err := os.WriteFile(filepath.Join(os.Getenv("UPDATER_FIXTURES"), "wsl.log"), []byte(log), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if os.Getenv("UPDATER_SCENARIO") == "native-wsl-failure" {
			os.Exit(23)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type updaterCase struct {
	name    string
	version string
	wantErr string
}

var updaterCases = []updaterCase{
	{name: "success"},
	{name: "prefixed-version", version: "v" + fixtureVersion},
	{name: "latest", version: "latest"},
	{name: "retry"},
	{name: "checksum-retry"},
	{name: "download-failure", wantErr: "failed after 3 attempts"},
	{name: "checksum-mismatch", wantErr: "checksum verification failed"},
	{name: "missing-checksum", wantErr: "exactly one"},
	{name: "duplicate-checksum", wantErr: "exactly one"},
	{name: "invalid-checksum", wantErr: "checksum"},
	{name: "missing-binary", wantErr: "go-fuck"},
	{name: "corrupt-archive", wantErr: ""},
	{name: "stage-failure", wantErr: "simulated stage failure"},
	{name: "replace-failure", wantErr: "simulated replace failure"},
	{name: "invalid-version", version: "1.2.3; invalid", wantErr: "invalid version"},
	{name: "multiline-version", version: "1.2.3\n4.5.6", wantErr: "invalid version"},
}

type updaterFixture struct {
	root, work, install string
}

func writeFixture(t *testing.T, name string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(name, data, mode); err != nil {
		t.Fatal(err)
	}
}

func newUpdaterFixture(t *testing.T, scenario, platform, arch string) updaterFixture {
	t.Helper()
	f := updaterFixture{root: t.TempDir()}
	f.work = filepath.Join(f.root, "work")
	f.install = filepath.Join(f.root, "install dir [stable]")
	for _, dir := range []string{f.work, f.install} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	binary, ext := "go-fuck", ".tar.gz"
	if platform == "windows" {
		binary, ext = "go-fuck.exe", ".zip"
	}
	writeFixture(t, filepath.Join(f.install, binary), []byte("old binary"), 0o755)
	name := binary
	if scenario == "missing-binary" {
		name = "README.md"
	}
	var archive bytes.Buffer
	if platform == "windows" {
		zw := zip.NewWriter(&archive)
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("new binary")); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gz := gzip.NewWriter(&archive)
		tw := tar.NewWriter(gz)
		header := &tar.Header{Name: name, Mode: 0o755, Size: int64(len("new binary"))}
		if scenario == "symlink-binary" {
			header.Typeflag, header.Linkname, header.Size = tar.TypeSymlink, "README.md", 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Size != 0 {
			if _, err := tw.Write([]byte("new binary")); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	data := archive.Bytes()
	if scenario == "corrupt-archive" {
		data = []byte("not an archive")
	}
	writeFixture(t, filepath.Join(f.root, "archive"), data, 0o644)
	var linuxArchive bytes.Buffer
	gz := gzip.NewWriter(&linuxArchive)
	tw := tar.NewWriter(gz)
	linuxHeader := &tar.Header{Name: "go-fuck", Mode: 0o755, Size: int64(len("new binary"))}
	if err := tw.WriteHeader(linuxHeader); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("new binary")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	linuxData := linuxArchive.Bytes()
	writeFixture(t, filepath.Join(f.root, "linux-archive"), linuxData, 0o644)
	asset := fmt.Sprintf("go-fuck_%s_%s_%s%s", fixtureVersion, platform, arch, ext)
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	linuxAsset := fmt.Sprintf("go-fuck_%s_linux_amd64.tar.gz", fixtureVersion)
	linuxHash := fmt.Sprintf("%x", sha256.Sum256(linuxData))
	switch scenario {
	case "checksum-mismatch":
		hash = strings.Repeat("0", 64)
	case "invalid-checksum":
		hash = "invalid"
	case "wsl-fallback-checksum":
		linuxHash = strings.Repeat("0", 64)
	}
	line := hash + "  " + asset + "\n"
	if linuxAsset != asset {
		line += linuxHash + "  " + linuxAsset + "\n"
	}
	switch scenario {
	case "missing-checksum":
		line = hash + "  " + asset + ".extra\n"
	case "duplicate-checksum":
		line += line
	}
	writeFixture(t, filepath.Join(f.root, "checksums"), []byte(line), 0o644)
	return f
}

func (f updaterFixture) env(scenario string) []string {
	return append(os.Environ(),
		"UPDATER_FIXTURES="+f.root, "UPDATER_SCENARIO="+scenario,
		"TMPDIR="+f.work, "TEMP="+f.work, "TMP="+f.work)
}

func (f updaterFixture) check(t *testing.T, tc updaterCase, output []byte, err error, windows bool) {
	t.Helper()
	wantFailure := tc.wantErr != "" || tc.name == "corrupt-archive"
	if (err != nil) != wantFailure {
		t.Fatalf("updater error = %v, want failure %v\n%s", err, wantFailure, output)
	}
	if tc.wantErr != "" && !strings.Contains(string(output), tc.wantErr) {
		t.Errorf("missing diagnostic %q:\n%s", tc.wantErr, output)
	}
	binary := "go-fuck"
	if windows {
		binary += ".exe"
	}
	data, readErr := os.ReadFile(filepath.Join(f.install, binary))
	if readErr != nil {
		t.Fatal(readErr)
	}
	want := "new binary"
	if wantFailure && tc.name != "native-wsl-failure" && tc.name != "wsl-fallback-checksum" {
		want = "old binary"
	}
	if string(data) != want {
		t.Errorf("installed binary = %q, want %q", data, want)
	}
	entries, readErr := os.ReadDir(f.work)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Errorf("temporary downloads were not cleaned up: %v", entries)
	}
	entries, readErr = os.ReadDir(f.install)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 1 {
		t.Errorf("staging files or backups were not cleaned up: %v", entries)
	}
	if !windows && !wantFailure {
		info, err := os.Stat(filepath.Join(f.install, binary))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o111 != 0o111 {
			t.Errorf("installed binary is not executable: %v", info.Mode())
		}
	}
	if strings.Contains(tc.name, "version") && wantFailure {
		if _, err := os.Stat(filepath.Join(f.root, "requests.log")); !os.IsNotExist(err) {
			t.Error("invalid version triggered a network request")
		}
	}
	if tc.name == "retry" || tc.name == "checksum-retry" || tc.name == "download-failure" {
		requests, readErr := os.ReadFile(filepath.Join(f.root, "requests.log"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		want := 3 // One failed attempt, one successful retry, then the other asset.
		if got := len(strings.Fields(string(requests))); got != want {
			t.Errorf("download attempts = %d, want %d: %s", got, want, requests)
		}
	}
}

func TestPowerShellUpdater(t *testing.T) {
	updater, err := filepath.Abs("update-go-fuck.ps1")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs("test-update-go-fuck.ps1")
	if err != nil {
		t.Fatal(err)
	}
	engines := []string{"pwsh"}
	if runtime.GOOS == "windows" {
		engines = append(engines, "powershell.exe")
	}
	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			program, err := exec.LookPath(engine)
			if err != nil && engine == "powershell.exe" {
				program, err = exec.LookPath(filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", engine))
			}
			if err != nil {
				t.Skip("PowerShell is not available")
			}
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			if out, err := exec.CommandContext(ctx, program, "-NoProfile", "-NonInteractive", "-Command", "exit 0").CombinedOutput(); err != nil {
				t.Skipf("PowerShell cannot start: %v\n%s", err, out)
			}
			cases := append([]updaterCase{}, updaterCases...)
			cases = append(cases, updaterCase{name: "arm64"}, updaterCase{name: "wsl"}, updaterCase{name: "wsl-timeout"},
				updaterCase{name: "wsl-fallback-checksum", wantErr: "checksum verification failed"},
				updaterCase{name: "native-wsl"},
				updaterCase{name: "native-wsl-failure", wantErr: "the WSL update failed"})
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					arch := "amd64"
					if tc.name == "arm64" {
						arch = "arm64"
					}
					f := newUpdaterFixture(t, tc.name, "windows", arch)
					version := tc.version
					if version == "" {
						version = fixtureVersion
					}
					ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, program, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
						"-File", driver, "-Updater", updater, "-InstallDir", f.install, "-Version", version)
					cmd.Env = f.env(tc.name)
					if strings.HasPrefix(tc.name, "native-wsl") {
						bin := filepath.Join(f.root, "native")
						if err := os.Mkdir(bin, 0o755); err != nil {
							t.Fatal(err)
						}
						exe, err := os.Executable()
						if err != nil {
							t.Fatal(err)
						}
						data, err := os.ReadFile(exe)
						if err != nil {
							t.Fatal(err)
						}
						writeFixture(t, filepath.Join(bin, "wsl.exe"), data, 0o755)
						cmd.Env = append(cmd.Env, "GO_FUCK_UPDATER_STDIN_HELPER=1", "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
					}
					if engine == "powershell.exe" {
						// An editor hosted by PowerShell 7 can export its Core modules
						// into Desktop PowerShell, hiding the 5.1 utility functions.
						modules := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "Modules")
						modules += ";" + filepath.Join(os.Getenv("ProgramFiles"), "WindowsPowerShell", "Modules")
						cmd.Env = append(cmd.Env, "PSModulePath="+modules)
					}
					output, err := cmd.CombinedOutput()
					f.check(t, tc, output, err, true)
					if strings.Contains(tc.name, "wsl") {
						data, err := os.ReadFile(filepath.Join(f.root, "wsl.log"))
						if err != nil {
							t.Fatal(err)
						}
						for _, want := range []string{"Test Distro", "--exec sh -c", "tr -d", "go-fuck " + fixtureVersion, "#!/bin/sh"} {
							if !strings.Contains(string(data), want) {
								t.Errorf("WSL input missing %q: %s", want, data)
							}
						}
						if tc.name == "wsl-timeout" {
							for _, want := range []string{"--exec uname -m", "--exec bash -c", "sha256sum --check", "go-fuck $version using Go"} {
								if !strings.Contains(string(data), want) {
									t.Errorf("WSL fallback input missing %q: %s", want, data)
								}
							}
						}
						if tc.name == "wsl-fallback-checksum" && strings.Contains(string(data), "--exec bash -c") {
							t.Errorf("checksum failure reached the WSL installer: %s", data)
						}
					}
				})
			}
		})
	}
}

func TestShellUpdater(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the POSIX updater is tested on Linux and macOS")
	}
	updater, err := filepath.Abs("update-go-fuck.sh")
	if err != nil {
		t.Fatal(err)
	}
	cases := append([]updaterCase{}, updaterCases...)
	cases = append(cases, updaterCase{name: "macos"}, updaterCase{name: "latest-fallback", version: "latest"},
		updaterCase{name: "symlink-binary", wantErr: "regular go-fuck binary"})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			platform, arch := "linux", "amd64"
			if tc.name == "macos" {
				platform, arch = "darwin", "arm64"
			}
			f := newUpdaterFixture(t, tc.name, platform, arch)
			bin := filepath.Join(f.root, "bin")
			if err := os.Mkdir(bin, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"awk", "cat", "cp", "cut", "grep", "gzip", "head", "install", "mkdir", "mktemp", "mv", "rm", "tar", "tr", "sh"} {
				program, err := exec.LookPath(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(program, filepath.Join(bin, name)); err != nil {
					t.Fatal(err)
				}
			}
			checksum := "sha256sum"
			if tc.name == "macos" {
				checksum = "shasum"
			}
			program, err := exec.LookPath(checksum)
			if err != nil && checksum == "sha256sum" {
				checksum = "shasum"
				program, err = exec.LookPath(checksum)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(program, filepath.Join(bin, checksum)); err != nil {
				t.Fatal(err)
			}
			writeFixture(t, filepath.Join(bin, "curl"), []byte(mockCurl), 0o755)
			writeFixture(t, filepath.Join(bin, "sleep"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
			uname := "#!/bin/sh\ncase $1 in -s) echo Linux;; -m) echo amd64;; esac\n"
			if platform == "darwin" {
				uname = "#!/bin/sh\ncase $1 in -s) echo Darwin;; -m) echo arm64;; esac\n"
			}
			writeFixture(t, filepath.Join(bin, "uname"), []byte(uname), 0o755)
			for name, scenario := range map[string]string{"install": "stage-failure", "mv": "replace-failure"} {
				if tc.name == scenario {
					if err := os.Remove(filepath.Join(bin, name)); err != nil {
						t.Fatal(err)
					}
					message := "simulated stage failure"
					if name == "mv" {
						message = "simulated replace failure"
					}
					writeFixture(t, filepath.Join(bin, name), []byte("#!/bin/sh\necho '"+message+"' >&2\nexit 1\n"), 0o755)
				}
			}
			version := tc.version
			if version == "" {
				version = fixtureVersion
			}
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "/bin/sh", updater, version, f.install)
			cmd.Env = append(f.env(tc.name), "PATH="+bin)
			output, err := cmd.CombinedOutput()
			f.check(t, tc, output, err, false)
		})
	}
}

const mockCurl = `#!/bin/sh
set -eu
output= timeout= connect=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) output=$2; shift 2 ;;
    --connect-timeout) connect=$2; shift 2 ;;
    --max-time) timeout=$2; shift 2 ;;
    --user-agent) shift 2 ;;
    -*) shift ;;
    *) url=$1; shift ;;
  esac
done
[ "$connect" = 10 ] || { echo 'missing connection timeout' >&2; exit 1; }
kind=archive
case "$url" in
  */checksums.txt) kind=checksums ;;
  https://api.github.com/*) kind=api ;;
  */releases/latest) kind=redirect ;;
esac
if [ "$kind" = archive ]; then
  [ "$timeout" = 120 ] || exit 1
else
  [ "$timeout" = 20 ] || exit 1
fi
echo "$kind" >> "$UPDATER_FIXTURES/requests.log"
count_file="$UPDATER_FIXTURES/$kind.count"
count=0
if [ -f "$count_file" ]; then count=$(cat "$count_file"); fi
count=$((count + 1))
echo "$count" > "$count_file"
case "$UPDATER_SCENARIO:$kind:$count" in
  retry:archive:1|checksum-retry:checksums:1|download-failure:archive:*)
    echo partial > "$output"; echo 'simulated transfer failure' >&2; exit 18 ;;
esac
case "$kind" in
  api)
    if [ "$UPDATER_SCENARIO" = latest-fallback ]; then
      echo '{"tag_name":"invalid"}' > "$output"
    else
      echo '{"tag_name":"v9.8.7"}' > "$output"
    fi ;;
  redirect) printf 'HTTP/2 302\r\nLocation: https://github.com/leafevans/go-fuck/releases/tag/v9.8.7\r\n' > "$output" ;;
  *) cp "$UPDATER_FIXTURES/$kind" "$output" ;;
esac
`
