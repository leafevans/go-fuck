package entrypoints

import (
	"strings"
	"testing"
)

func TestBashHistoryEntrySelection(t *testing.T) {
	for _, tc := range []struct{ history, want string }{
		{"unmarked\r\n", "unmarked"},
		{"preamble\n\t \n", ""},
		{"\t first\n\t second\n\t \n", "second"},
		{"\t echo first\r\necho second\r\n\t ", "echo first\r\necho second"},
	} {
		got := bashMarkedRawCommand(tc.history)
		if tc.want == "" && len(got) == 0 {
			continue
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("bashMarkedRawCommand(%q) = %q, want %q", tc.history, got, tc.want)
		}
	}
}

func FuzzMarkedRawCommand(f *testing.F) {
	seeds := []string{
		"",
		"\t git status",
		"\t if true; then\necho hi\nfi\n\t fuck",
		"if true; then\necho hi\nfi\n",
		"\t fuck --yes",
		"preamble\n\t fuck",
		"\t git status\n\t \n\t fuck",
		strings.Repeat("line\n", 1000),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, payload string) {
		results := [][]string{
			sanitizeMarkedCommands(bashMarkedRawCommand(payload)),
			sanitizeMarkedCommands(zshMarkedRawCommand(payload)),
			sanitizeMarkedCommands(singleMarkedCommand(payload)),
		}
		for _, commands := range results {
			for _, command := range commands {
				if strings.Contains(command, rawCommandMarker) {
					t.Fatalf("command %q contains the internal marker", command)
				}
			}
		}
	})
}
