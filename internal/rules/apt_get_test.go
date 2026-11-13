package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

func TestRuleAptGet(t *testing.T) {
	resetTestState(t)
	rule := Get("apt_get")
	if rule == nil {
		t.Fatal("apt_get rule is not registered")
	}

	originalWhich := aptGetWhich
	originalPackages := aptGetPackages
	t.Cleanup(func() {
		aptGetWhich = originalWhich
		aptGetPackages = originalPackages
	})

	vimPackages := []string{"vim", "vim-tiny"}
	convertPackages := []string{"imagemagick", "graphicsmagick-imagemagick-compat"}

	matchCases := []struct {
		script string
		output string
	}{
		{"vim", "vim: command not found"},
		{"sudo vim", "vim: command not found"},
		{"vim", "The program 'vim' is currently not installed. You can install it by typing: sudo apt install vim"},
	}
	for _, testCase := range matchCases {
		t.Run(testCase.script, func(t *testing.T) {
			aptGetWhich = func(string) string { return "" }
			aptGetPackages = func(string) []string { return vimPackages }
			assertMatch(t, rule, testCase.script, testCase.output)
		})
	}

	notMatchCases := []struct {
		script   string
		output   string
		packages []string
		which    string
	}{
		{"a_bad_cmd", "a_bad_cmd: command not found", nil, ""},
		{"vim", "", nil, ""},
		{"", "", nil, ""},
		{"vim", "vim: command not found", []string{"vim"}, "/usr/bin/vim"},
		{"sudo vim", "vim: command not found", []string{"vim"}, "/usr/bin/vim"},
	}
	for _, testCase := range notMatchCases {
		t.Run("not_"+testCase.script, func(t *testing.T) {
			aptGetWhich = func(string) string { return testCase.which }
			aptGetPackages = func(string) []string { return testCase.packages }
			assertNotMatch(t, rule, testCase.script, testCase.output)
		})
	}

	newCommandCases := []struct {
		script   string
		packages []string
		want     string
	}{
		{"vim", vimPackages, "sudo apt-get install vim && vim"},
		{"convert", convertPackages, "sudo apt-get install imagemagick && convert"},
		{"sudo vim", vimPackages, "sudo apt-get install vim && sudo vim"},
		{"sudo convert", convertPackages, "sudo apt-get install imagemagick && sudo convert"},
	}
	for _, testCase := range newCommandCases {
		t.Run("new_"+testCase.script, func(t *testing.T) {
			aptGetWhich = func(string) string { return "" }
			aptGetPackages = func(string) []string { return testCase.packages }
			assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
		})
	}
}

func TestAptGetHelpers(t *testing.T) {
	resetTestState(t)
	originalWhich := aptGetWhich
	originalPackages := aptGetPackages
	t.Cleanup(func() { aptGetWhich = originalWhich; aptGetPackages = originalPackages })

	if got := aptGetExecutable(types.NewCommand("", "")); got != "" {
		t.Fatalf("aptGetExecutable(empty) = %q", got)
	}
	if got := aptGetExecutable(types.NewCommand("sudo", "")); got != "" {
		t.Fatalf("aptGetExecutable(sudo) = %q", got)
	}
	if got := aptGetExecutable(types.NewCommand("sudo make", "")); got != "make" {
		t.Fatalf("aptGetExecutable(sudo make) = %q", got)
	}
	if got := aptGetExecutable(types.NewCommand("make", "")); got != "make" {
		t.Fatalf("aptGetExecutable(make) = %q", got)
	}

	aptGetWhich = func(string) string { return "" }
	aptGetPackages = func(string) []string { return nil }
	if aptGetMatch(types.NewCommand("make", "command not found")) {
		t.Fatal("aptGetMatch() matched without a package")
	}
	aptGetPackages = func(string) []string { return []string{"make"} }
	if !aptGetMatch(types.NewCommand("make", "command not found")) {
		t.Fatal("aptGetMatch() did not match an installable package")
	}
	aptGetWhich = func(string) string { return "/usr/bin/make" }
	if aptGetMatch(types.NewCommand("make", "command not found")) {
		t.Fatal("aptGetMatch() matched an available executable")
	}

	aptGetPackages = func(string) []string { return nil }
	if got, _ := aptGetNewCommand(types.NewCommand("make", "command not found")); len(got) != 0 {
		t.Fatalf("aptGetNewCommand(no package) = %q", got)
	}
	aptGetPackages = func(string) []string { return []string{"make"} }
	if got, _ := aptGetNewCommand(types.NewCommand("make", "command not found")); len(got) != 1 ||
		got[0] != "sudo apt-get install make && make" {
		t.Fatalf("aptGetNewCommand() = %q", got)
	}

	t.Setenv("PATH", "")
	utils.ResetCaches()
	_ = aptGetHandlerPath()
}
