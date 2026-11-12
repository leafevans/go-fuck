package entrypoints

import (
	"strings"
	"testing"
)

func FuzzMarkedRawCommand(f *testing.F) {
	seeds := []string{
		"",
		"\t git status",
		"\t if true; then\necho hi\nfi\n\t fuck",
		"if true; then\necho hi\nfi\n",
		"\t fuck --yes",
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
