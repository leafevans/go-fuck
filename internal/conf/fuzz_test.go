package conf

import (
	"sort"
	"testing"
)

// FuzzSettingsFile checks that arbitrary settings files cannot panic the
// parsing and application code, and that the unknown-key diagnostics keep
// their contract: sorted, unique and limited to keys that are not part of the
// settings schema.
func FuzzSettingsFile(f *testing.F) {
	seeds := []string{
		`{"rules": ["git_push"], "wait_command": 3}`,
		`{"strict_settings": true, "rule": ["typo"]}`,
		`{"_comment": "x", "settings": {"fixlinecmd": "y"}}`,
		`{"priority": {"git_push": 900}, "env": {"A": "B"}}`,
		`null`,
		`[]`,
		`{`,
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		schema, unknown, err := parseSettingsFile(data)
		if err != nil {
			return
		}
		if schema == nil {
			t.Fatalf("parseSettingsFile(%q) returned a nil schema without an error", data)
		}
		if !sort.StringsAreSorted(unknown) {
			t.Fatalf("unknown keys are not sorted: %q", unknown)
		}
		seen := map[string]bool{}
		for _, key := range unknown {
			if key == "" {
				t.Fatalf("unknown keys contain an empty name: %q", unknown)
			}
			if knownSettingKeys[key] {
				t.Fatalf("known setting %q reported as unknown", key)
			}
			if seen[key] {
				t.Fatalf("unknown keys contain %q twice", key)
			}
			seen[key] = true
		}

		settings := Default()
		settings.UnknownKeys = unknown
		settings.applyFile(schema)
	})
}
