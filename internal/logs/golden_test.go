package logs

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
)

// updateGolden rewrites the golden files instead of comparing against them.
var updateGolden = flag.Bool("update-golden", false, "rewrite the testdata golden files")

// assertGolden compares got with the golden file below testdata.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	if got != string(want) {
		t.Fatalf("output does not match testdata/%s; run go test ./internal/logs -run TestGolden -update-golden", name)
	}
}

// TestGoldenBanner locks the version banner output for every color profile.
func TestGoldenBanner(t *testing.T) {
	cases := []struct {
		name      string
		term      string
		colorTerm string
		wantFile  string
	}{
		{"none", "dumb", "", "banner_none.golden"},
		{"ansi", "xterm", "", "banner_ansi.golden"},
		{"256", "xterm-256color", "", "banner_256.golden"},
		{"truecolor", "xterm-256color", "truecolor", "banner_truecolor.golden"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_ = resetLogsState(t)
			conf.Reset()
			setBannerEnvironment(t, testCase.term, testCase.colorTerm, "", "")
			assertGolden(t, testCase.wantFile, renderBanner(bannerInfo(), false))
		})
	}
}

// goldenMessages renders the interactive diagnostics used by the golden test.
func goldenMessages(t *testing.T, truecolor bool) string {
	t.Helper()
	buffer := resetLogsState(t)
	if truecolor {
		setBannerEnvironment(t, "xterm-256color", "truecolor", "", "")
		conf.Current().NoColors = false
	}
	ConfirmText("git branch", false)
	buffer.WriteString(newline)
	ShowCorrectedCommand("git add .", true)
	Warn("instant mode is not supported for Fish Shell")
	Failed("No fucks given")
	conf.Current().Debug = true
	Debug("trying rule: git_push")
	return buffer.String()
}

// TestGoldenMessages locks the confirmation prompt and the diagnostics for the
// plain and the 24 bit color profiles.
func TestGoldenMessages(t *testing.T) {
	cases := []struct {
		name      string
		truecolor bool
		wantFile  string
	}{
		{"none", false, "messages_none.golden"},
		{"truecolor", true, "messages_truecolor.golden"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assertGolden(t, testCase.wantFile, goldenMessages(t, testCase.truecolor))
		})
	}
}
