package utils

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// resetUtilsState resets the global state used by the utils tests.
func resetUtilsState(t *testing.T) {
	t.Helper()
	conf.Reset()
	shells.Set(shells.NewGeneric())
	ResetCaches()
	t.Cleanup(func() {
		conf.Reset()
		shells.Set(shells.NewGeneric())
		ResetCaches()
	})
}

// historyShell is a shell with a fixed history.
type historyShell struct {
	*shells.Generic
	entries []string
}

// GetHistory returns the fixed history entries.
func (h *historyShell) GetHistory() []string {
	return h.entries
}

// fakeExecutableDir creates a directory with the given files and returns it.
//
// On Windows the executable suffix is required by exec.LookPath, so both the
// plain name and the ".exe" variant are created.
func fakeExecutableDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	create := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o755); err != nil {
			t.Fatalf("create fake executable %q: %v", name, err)
		}
	}
	for _, name := range names {
		create(name)
		if runtime.GOOS == "windows" {
			create(name + ".exe")
		}
	}
	return dir
}

func TestSequenceRatio(t *testing.T) {
	cases := []struct {
		left, right string
		expected    float64
	}{
		{"bisect strt", "start", 0.5},
		{"bisect strt", "reset", 0.375},
		{"bisect strt", "terms", 0.125},
		{"st", "status", 0.5},
		{"st", "reset", 0.5714285714285714},
		{"st", "stage", 0.5714285714285714},
		{"brnch", "branch", 0.9090909090909091},
		{"vom", "vim", 0.6666666666666666},
		{"vom", "go", 0.4},
		{"got", "go", 0.8},
		{"got", "git", 0.6666666666666666},
		{"re", "grep", 0.6666666666666666},
		{"re", "rm", 0.5},
		{"cone", "clone", 0.8888888888888888},
		{"cafe", "café", 0.75},
		{"", "", 1},
	}
	for _, testCase := range cases {
		got := SequenceRatio(testCase.left, testCase.right)
		if math.Abs(got-testCase.expected) > 1e-9 {
			t.Errorf("SequenceRatio(%q, %q) = %v, want %v",
				testCase.left, testCase.right, got, testCase.expected)
		}
	}
}

func TestGetCloseMatches(t *testing.T) {
	resetUtilsState(t)
	cases := []struct {
		word          string
		possibilities []string
		cutoff        float64
		expected      []string
	}{
		{"st", []string{"status", "reset", "stage", "stash", "stats"}, 0.1,
			[]string{"stats", "stash", "stage"}},
		{"bisect strt", []string{"start", "reset", "terms"}, 0.1,
			[]string{"start", "reset", "terms"}},
		{"brnch", []string{"branch", "status"}, 0.6, []string{"branch"}},
		{"vom", []string{"vim", "fsck", "git", "go", "python"}, 0.1,
			[]string{"vim", "go", "python"}},
		{"got", []string{"vim", "fsck", "git", "go", "python"}, 0.1,
			[]string{"go", "git", "python"}},
		{"gti", []string{"vim", "fsck", "git", "go", "python"}, 0.1,
			[]string{"git", "go", "vim"}},
		{"re", []string{"rebase", "reset", "grep", "rm"}, 0.1,
			[]string{"grep", "reset", "rm"}},
		{"cone", []string{"clone"}, 0.1, []string{"clone"}},
	}
	for _, testCase := range cases {
		got := GetCloseMatches(testCase.word, testCase.possibilities, 3, testCase.cutoff)
		if strings.Join(got, "|") != strings.Join(testCase.expected, "|") {
			t.Errorf("GetCloseMatches(%q) = %q, want %q",
				testCase.word, got, testCase.expected)
		}
	}
}

func TestGetCloseMatchesLimit(t *testing.T) {
	resetUtilsState(t)
	conf.Current().NumCloseMatches = 1
	got := GetCloseMatches("re", []string{"rebase", "reset", "grep", "rm"}, 0, 0.1)
	if len(got) != 1 || got[0] != "grep" {
		t.Fatalf("GetCloseMatches() = %q, want [grep]", got)
	}
}

func TestGetClosest(t *testing.T) {
	resetUtilsState(t)
	if got := GetClosestDefault("brnch", []string{"branch", "status"}); got != "branch" {
		t.Fatalf("GetClosestDefault() = %q, want branch", got)
	}
	if got := GetClosest("st", []string{"status", "reset"}, 0.6, true); got != "status" {
		t.Fatalf("GetClosest() = %q, want status", got)
	}
	if got := GetClosest("zzz", []string{"status"}, 0.6, false); got != "" {
		t.Fatalf("GetClosest() = %q, want an empty string", got)
	}
	if got := GetClosest("zzz", []string{"status"}, 0.6, true); got != "status" {
		t.Fatalf("GetClosest() = %q, want status", got)
	}
}

func TestReplaceArgument(t *testing.T) {
	cases := []struct {
		script, from, to, expected string
	}{
		{"apt-get instol vim", "instol", "install", "apt-get install vim"},
		{"git brnch", "brnch", "branch", "git branch"},
		{"git brnch", "missing", "branch", "git brnch"},
	}
	for _, testCase := range cases {
		if got := ReplaceArgument(testCase.script, testCase.from, testCase.to); got != testCase.expected {
			t.Errorf("ReplaceArgument(%q, %q, %q) = %q, want %q",
				testCase.script, testCase.from, testCase.to, got, testCase.expected)
		}
	}
}

func TestReplaceCommand(t *testing.T) {
	resetUtilsState(t)
	command := types.NewCommand("git brnch", "")
	got := ReplaceCommand(command, "brnch", []string{"branch", "status"})
	if len(got) == 0 || got[0] != "git branch" {
		t.Fatalf("ReplaceCommand() = %q, want a branch suggestion", got)
	}
}

func TestGetAllMatchedCommands(t *testing.T) {
	cases := []struct {
		output   string
		expected []string
	}{
		{
			"git: 'cone' is not a git command. See 'git --help'.\n\nDid you mean one of these?\n\tclone",
			[]string{"clone"},
		},
		{
			"git: 're' is not a git command. See 'git --help'.\n\nDid you mean one of these?\n\trebase\n\treset\n\tgrep\n\trm",
			[]string{"rebase", "reset", "grep", "rm"},
		},
	}
	for _, testCase := range cases {
		got := GetAllMatchedCommands(testCase.output, "Did you mean")
		if strings.Join(got, "|") != strings.Join(testCase.expected, "|") {
			t.Errorf("GetAllMatchedCommands() = %q, want %q", got, testCase.expected)
		}
	}

	got := GetAllMatchedCommands(
		"git: 're' is not a git command. See 'git --help'.\n\nDid you mean one of these?\n\trebase\n\treset",
		"The most similar command", "Did you mean")
	if strings.Join(got, "|") != "rebase|reset" {
		t.Fatalf("GetAllMatchedCommands() with multiple separators = %q", got)
	}
}

func TestIsApp(t *testing.T) {
	resetUtilsState(t)
	cases := []struct {
		script   string
		names    []string
		expected bool
	}{
		{"/usr/bin/git diff", []string{"git", "hub"}, true},
		{"/bin/hdfs dfs -rm foo", []string{"hdfs"}, true},
		{"git diff", []string{"git", "hub"}, true},
		{"hub diff", []string{"git", "hub"}, true},
		{"hg diff", []string{"git", "hub"}, false},
	}
	for _, testCase := range cases {
		got := IsApp(types.NewCommand(testCase.script, ""), 0, testCase.names...)
		if got != testCase.expected {
			t.Errorf("IsApp(%q, %q) = %v, want %v",
				testCase.script, testCase.names, got, testCase.expected)
		}
	}

	if !IsApp(types.NewCommand("git diff", ""), 1, "git") {
		t.Error("IsApp() with atLeast=1 should match")
	}
	if IsApp(types.NewCommand("git", ""), 1, "git") {
		t.Error("IsApp() with atLeast=1 should not match a bare command")
	}
}

func TestForApp(t *testing.T) {
	resetUtilsState(t)
	match := ForApp([]string{"git", "hub"}, 0, func(*types.Command) bool { return true })
	if !match(types.NewCommand("git diff", "")) {
		t.Error("ForApp() should call the wrapped function")
	}
	if match(types.NewCommand("hg diff", "")) {
		t.Error("ForApp() should not call the wrapped function")
	}
}

func TestForAppNew(t *testing.T) {
	resetUtilsState(t)
	newCommand := ForAppNew([]string{"git"}, 0, func(command *types.Command) ([]string, error) {
		return []string{command.Script + "!"}, nil
	})
	got, err := newCommand(types.NewCommand("git diff", ""))
	if err != nil || len(got) != 1 || got[0] != "git diff!" {
		t.Fatalf("ForAppNew() = %q, %v", got, err)
	}
	got, err = newCommand(types.NewCommand("hg diff", ""))
	if err != nil || got != nil {
		t.Fatalf("ForAppNew() = %q, %v, want no commands", got, err)
	}
}

func TestGetAllExecutables(t *testing.T) {
	resetUtilsState(t)
	dir := fakeExecutableDir(t, "vim", "fsck")
	t.Setenv("PATH", dir)
	ResetCaches()

	executables := GetAllExecutables()
	found := map[string]bool{}
	for _, executable := range executables {
		found[executable] = true
	}
	if !found["vim"] || !found["fsck"] {
		t.Fatalf("GetAllExecutables() = %q, want vim and fsck", executables)
	}
}

func TestGetAllExecutablesExcludesPaths(t *testing.T) {
	resetUtilsState(t)
	first := fakeExecutableDir(t, "vim")
	second := fakeExecutableDir(t, "fsck")
	t.Setenv("PATH", first+string(os.PathListSeparator)+second)
	conf.Current().ExcludedSearchPathPrefixes = []string{second}
	ResetCaches()

	for _, executable := range GetAllExecutables() {
		if executable == "fsck" {
			t.Fatalf("GetAllExecutables() included an executable from an excluded path")
		}
	}
}

func TestGetValidHistoryWithoutCurrent(t *testing.T) {
	resetUtilsState(t)
	dir := fakeExecutableDir(t, "ls", "diff", "café")
	t.Setenv("PATH", dir)
	ResetCaches()

	shell := &historyShell{
		Generic: shells.NewGeneric(),
		entries: []string{"le cat", "fuck", "ls cat", "diff x", "nocommand x", "café ô"},
	}
	shells.Set(shell)

	cases := []struct {
		script   string
		expected []string
	}{
		{"le cat", []string{"ls cat", "diff x", "café ô"}},
		{"diff x", []string{"ls cat", "café ô"}},
		{"fuck", []string{"ls cat", "diff x", "café ô"}},
		{"cafe ô", []string{"ls cat", "diff x", "café ô"}},
	}
	for _, testCase := range cases {
		got := GetValidHistoryWithoutCurrent(types.NewCommand(testCase.script, ""))
		if strings.Join(got, "|") != strings.Join(testCase.expected, "|") {
			t.Errorf("GetValidHistoryWithoutCurrent(%q) = %q, want %q",
				testCase.script, got, testCase.expected)
		}
	}
}

func TestMemoize(t *testing.T) {
	calls := 0
	memoized := Memoize(func() int {
		calls++
		return 42
	})
	first, second := memoized(), memoized()
	if first != 42 || second != 42 {
		t.Fatal("Memoize() returned a wrong value")
	}
	if calls != 1 {
		t.Fatalf("Memoize() called the function %d times, want 1", calls)
	}
}

func TestMemoize1(t *testing.T) {
	calls := 0
	memoized := Memoize1(func(key string) int {
		calls++
		return len(key)
	})
	if memoized("ab") != 2 || memoized("abc") != 3 || memoized("ab") != 2 {
		t.Fatal("Memoize1() returned a wrong value")
	}
	if calls != 2 {
		t.Fatalf("Memoize1() called the function %d times, want 2", calls)
	}
}

func TestContainsAndIndex(t *testing.T) {
	if !Contains([]string{"a", "b"}, "b") || Contains([]string{"a"}, "c") {
		t.Fatal("Contains() returned a wrong value")
	}
	if Index([]string{"a", "b"}, "b") != 1 || Index([]string{"a"}, "c") != -1 {
		t.Fatal("Index() returned a wrong value")
	}
}

func TestWhich(t *testing.T) {
	resetUtilsState(t)
	if Which("definitely-not-an-executable") != "" {
		t.Fatal("Which() found a missing executable")
	}
	dir := fakeExecutableDir(t, "go-fuck-test-executable")
	t.Setenv("PATH", dir)
	if Which("go-fuck-test-executable") == "" {
		t.Fatal("Which() did not find an existing executable")
	}
}

func TestGetAlias(t *testing.T) {
	t.Setenv("TF_ALIAS", "please")
	if got := GetAlias(); got != "please" {
		t.Fatalf("GetAlias() = %q, want please", got)
	}
}
