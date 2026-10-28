package specific

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// resetSpecificState isolates the shell and the executable cache of the tests.
func resetSpecificState(t *testing.T) {
	t.Helper()
	shells.Set(shells.NewGeneric())
	utils.ResetCaches()
	t.Cleanup(func() {
		shells.Set(shells.NewGeneric())
		utils.ResetCaches()
	})
}

// fakeSpecificExecutables creates executable stubs in a temporary directory
// and points PATH at it, so that utils.Which finds only the stubs.
func fakeSpecificExecutables(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			path += ".exe"
		}
		if err := os.WriteFile(path, []byte{}, 0o755); err != nil {
			t.Fatalf("create fake executable %q: %v", name, err)
		}
	}
	t.Setenv("PATH", dir)
	return dir
}

// writeSpecificScript writes an executable shell script that fakes an external
// tool. It is only used by tests that are skipped on Windows.
func writeSpecificScript(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake executable %q: %v", path, err)
	}
}

func TestSudoSupportMatch(t *testing.T) {
	onlyGitPush := func(command *types.Command) bool {
		return command.Script == "git push"
	}
	wrapped := SudoSupportMatch(onlyGitPush)

	if !wrapped(types.NewCommand("git push", "")) {
		t.Error("SudoSupportMatch() did not match the plain command")
	}
	if !wrapped(types.NewCommand("sudo git push", "")) {
		t.Error("SudoSupportMatch() did not strip the sudo prefix")
	}
	if wrapped(types.NewCommand("sudo git pull", "")) {
		t.Error("SudoSupportMatch() matched the wrong command")
	}
}

func TestSudoSupportNew(t *testing.T) {
	force := func(command *types.Command) ([]string, error) {
		return []string{command.Script + " --force"}, nil
	}
	wrapped := SudoSupportNew(force)

	commands, err := wrapped(types.NewCommand("git push", ""))
	if err != nil {
		t.Fatalf("SudoSupportNew() error = %v", err)
	}
	if len(commands) != 1 || commands[0] != "git push --force" {
		t.Fatalf("SudoSupportNew() = %q, want [git push --force]", commands)
	}

	commands, err = wrapped(types.NewCommand("sudo git push", ""))
	if err != nil {
		t.Fatalf("SudoSupportNew() error = %v", err)
	}
	if len(commands) != 1 || commands[0] != "sudo git push --force" {
		t.Fatalf("SudoSupportNew() = %q, want [sudo git push --force]", commands)
	}

	lookupErr := errors.New("lookup failed")
	failing := func(command *types.Command) ([]string, error) {
		return []string{"partial"}, lookupErr
	}
	commands, err = SudoSupportNew(failing)(types.NewCommand("sudo git push", ""))
	if !errors.Is(err, lookupErr) {
		t.Fatalf("SudoSupportNew() error = %v, want %v", err, lookupErr)
	}
	if len(commands) != 1 || commands[0] != "partial" {
		t.Fatalf("SudoSupportNew() = %q, want [partial]", commands)
	}
}

func TestGitSupportMatch(t *testing.T) {
	resetSpecificState(t)

	var matchedScript string
	matchAll := func(command *types.Command) bool {
		matchedScript = command.Script
		return true
	}
	wrapped := GitSupportMatch(matchAll)

	if wrapped(types.NewCommand("ls", "")) {
		t.Error("GitSupportMatch() matched a non git command")
	}
	if !wrapped(types.NewCommand("git status", "")) {
		t.Error("GitSupportMatch() did not match git")
	}
	if !wrapped(types.NewCommand("hub status", "")) {
		t.Error("GitSupportMatch() did not match hub")
	}

	// A git alias reported through GIT_TRACE is expanded and quoted so that
	// the wrapped rule can match on the real command.
	aliasOutput := "trace: alias expansion: cm => commit -m \"hello world\"\n"
	if !wrapped(types.NewCommand("git cm", aliasOutput)) {
		t.Fatal("GitSupportMatch() did not match a git alias")
	}
	if matchedScript != "git commit -m 'hello world'" {
		t.Fatalf("GitSupportMatch() script = %q, want the expanded alias", matchedScript)
	}

	if !wrapped(types.NewCommand("git st", "fatal: not a git repository\n")) {
		t.Fatal("GitSupportMatch() did not match ordinary git output")
	}
	if matchedScript != "git st" {
		t.Fatalf("GitSupportMatch() script = %q, want it unchanged", matchedScript)
	}
}

func TestGitSupportNew(t *testing.T) {
	resetSpecificState(t)

	identity := func(command *types.Command) ([]string, error) {
		return []string{command.Script}, nil
	}
	wrapped := GitSupportNew(identity)

	commands, err := wrapped(types.NewCommand("ls", ""))
	if err != nil {
		t.Fatalf("GitSupportNew() error = %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("GitSupportNew() = %q for a non git command, want none", commands)
	}

	commands, err = wrapped(types.NewCommand("git push", ""))
	if err != nil {
		t.Fatalf("GitSupportNew() error = %v", err)
	}
	if len(commands) != 1 || commands[0] != "git push" {
		t.Fatalf("GitSupportNew() = %q, want [git push]", commands)
	}

	aliasOutput := "trace: alias expansion: ps => push --force\n"
	commands, err = wrapped(types.NewCommand("git ps", aliasOutput))
	if err != nil {
		t.Fatalf("GitSupportNew() error = %v", err)
	}
	if len(commands) != 1 || commands[0] != "git push --force" {
		t.Fatalf("GitSupportNew() = %q, want [git push --force]", commands)
	}
}

func TestAvailabilityHelpers(t *testing.T) {
	cases := []struct {
		name        string
		available   func() bool
		executables [][]string
	}{
		{name: "apt", available: AptAvailable, executables: [][]string{{"apt-get"}}},
		{name: "brew", available: BrewAvailable, executables: [][]string{{"brew"}}},
		{name: "dnf", available: DnfAvailable, executables: [][]string{{"dnf"}}},
		{name: "yum", available: YumAvailable, executables: [][]string{{"yum"}}},
		{name: "nix", available: NixAvailable, executables: [][]string{{"nix"}}},
		{name: "npm", available: NpmAvailable, executables: [][]string{{"npm"}}},
		{name: "choco", available: ChocoAvailable, executables: [][]string{{"choco"}, {"cinst"}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, executables := range testCase.executables {
				t.Run("available/"+strings.Join(executables, "_"), func(t *testing.T) {
					fakeSpecificExecutables(t, executables...)
					if !testCase.available() {
						t.Errorf("%s() = false, want true", testCase.name)
					}
				})
			}
			t.Run("unavailable", func(t *testing.T) {
				fakeSpecificExecutables(t)
				if testCase.available() {
					t.Errorf("%s() = true, want false", testCase.name)
				}
			})
		})
	}
}

func TestArchlinuxEnv(t *testing.T) {
	cases := []struct {
		name        string
		executables []string
		pacman      string
		enabled     bool
	}{
		{name: "yay with pkgfile", executables: []string{"yay", "pkgfile"}, pacman: "yay", enabled: true},
		{name: "yay without pkgfile", executables: []string{"yay"}, pacman: "yay"},
		{name: "yay before pikaur", executables: []string{"yay", "pikaur", "pkgfile"}, pacman: "yay", enabled: true},
		{name: "pikaur", executables: []string{"pikaur", "pkgfile"}, pacman: "pikaur", enabled: true},
		{name: "yaourt", executables: []string{"yaourt"}, pacman: "yaourt"},
		{name: "pacman", executables: []string{"pacman"}, pacman: "sudo pacman"},
		{name: "pacman with pkgfile", executables: []string{"pacman", "pkgfile"}, pacman: "sudo pacman", enabled: true},
		{name: "no package manager"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeSpecificExecutables(t, testCase.executables...)
			enabled, pacman := ArchlinuxEnv()
			if pacman != testCase.pacman {
				t.Fatalf("ArchlinuxEnv() pacman = %q, want %q", pacman, testCase.pacman)
			}
			if enabled != testCase.enabled {
				t.Fatalf("ArchlinuxEnv() enabled = %t, want %t", enabled, testCase.enabled)
			}
		})
	}
}

func TestGetPkgFileUsesInjectedLookup(t *testing.T) {
	original := getPkgFile
	t.Cleanup(func() {
		getPkgFile = original
	})

	lookupErr := errors.New("pkgfile failed")
	getPkgFile = func(command string) pkgFileResult {
		switch command {
		case "broken":
			return pkgFileResult{err: lookupErr}
		case "empty":
			return pkgFileResult{}
		default:
			return pkgFileResult{packages: []string{"extra/" + command}}
		}
	}

	packages, err := GetPkgFile("make")
	if err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if len(packages) != 1 || packages[0] != "extra/make" {
		t.Fatalf("GetPkgFile() = %q, want [extra/make]", packages)
	}

	packages, err = GetPkgFile("empty")
	if err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if len(packages) != 0 {
		t.Fatalf("GetPkgFile() = %q, want no packages", packages)
	}

	packages, err = GetPkgFile("broken")
	if !errors.Is(err, lookupErr) {
		t.Fatalf("GetPkgFile() error = %v, want %v", err, lookupErr)
	}
	if len(packages) != 0 {
		t.Fatalf("GetPkgFile() = %q, want no packages", packages)
	}
}

// countCalls returns the number of lines written to the given file.
func countCalls(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("read call counter: %v", err)
	}
	return strings.Count(string(data), "called")
}

func TestGetPkgFileRunsPkgfile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a deterministic fake pkgfile executable cannot be created without a compiler")
	}

	dir := t.TempDir()
	counter := filepath.Join(t.TempDir(), "calls")
	script := "#!/bin/sh\n" +
		"case \"$3\" in\n" +
		"  missing) exit 1 ;;\n" +
		"  broken) echo oops >&2; exit 2 ;;\n" +
		"esac\n" +
		"printf 'pkg/%s\\n' \"$3\"\n" +
		"echo called >> \"$GOFUCK_PKGFILE_COUNTER\"\n"
	writeSpecificScript(t, filepath.Join(dir, "pkgfile"), script)
	t.Setenv("PATH", dir)
	t.Setenv("GOFUCK_PKGFILE_COUNTER", counter)

	packages, err := GetPkgFile("sudo make")
	if err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if len(packages) != 1 || packages[0] != "pkg/make" {
		t.Fatalf("GetPkgFile() = %q, want [pkg/make]", packages)
	}

	// The lookup is memoized, so an identical command does not run pkgfile
	// again. The counter is not compared against a fixed value because the
	// memoization cache survives across -count iterations.
	before := countCalls(t, counter)
	if _, err := GetPkgFile("sudo make"); err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if after := countCalls(t, counter); after != before {
		t.Fatalf("GetPkgFile() ran pkgfile %d times for the same command, want %d", after, before)
	}

	// Leading and trailing spaces and the sudo prefix are stripped before the
	// command is passed to pkgfile.
	packages, err = GetPkgFile("  sudo make  ")
	if err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if len(packages) != 1 || packages[0] != "pkg/make" {
		t.Fatalf("GetPkgFile() = %q, want [pkg/make]", packages)
	}

	// pkgfile exits with 1 and no output when nothing provides the command.
	packages, err = GetPkgFile("missing")
	if err != nil {
		t.Fatalf("GetPkgFile() error = %v", err)
	}
	if len(packages) != 0 {
		t.Fatalf("GetPkgFile() = %q, want no packages", packages)
	}

	// Other exit statuses are reported as errors.
	if _, err := GetPkgFile("broken"); err == nil {
		t.Fatal("GetPkgFile() error = nil, want the pkgfile failure")
	}

	// An empty command is not looked up at all.
	if packages, err := GetPkgFile("   "); err != nil || len(packages) != 0 {
		t.Fatalf("GetPkgFile() = %q, %v for an empty command", packages, err)
	}
}

func TestNpmGetScriptsParsesOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a deterministic fake npm executable cannot be created without a compiler")
	}

	dir := t.TempDir()
	backtick := string(rune(96))
	script := "#!/bin/sh\ncat <<'EOF'\n" +
		"Lifecycle scripts included in demo@1.0.0:\n" +
		"  test\n" +
		"    mocha\n" +
		"\n" +
		"available via " + backtick + "npm run-script" + backtick + ":\n" +
		"  build\n" +
		"    tsc\n" +
		"  lint\n" +
		"    eslint .\n" +
		"EOF\n"
	writeSpecificScript(t, filepath.Join(dir, "npm"), script)
	// The fake npm script uses "cat", which the shell has to find in the
	// standard directories: the PATH below only contains the stub first.
	t.Setenv("PATH", dir+string(os.PathListSeparator)+"/bin"+string(os.PathListSeparator)+"/usr/bin")

	scripts := NpmGetScripts()
	if len(scripts) != 2 || scripts[0] != "build" || scripts[1] != "lint" {
		t.Fatalf("NpmGetScripts() = %q, want [build lint]", scripts)
	}
}

func TestGetBrewPathPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a deterministic fake brew executable cannot be created without a compiler")
	}

	dir := t.TempDir()
	writeSpecificScript(t, filepath.Join(dir, "brew"),
		"#!/bin/sh\nprintf '/opt/homebrew\\n'\n")
	t.Setenv("PATH", dir)

	if got := GetBrewPathPrefix(); got != "/opt/homebrew" {
		t.Fatalf("GetBrewPathPrefix() = %q, want /opt/homebrew", got)
	}
}
