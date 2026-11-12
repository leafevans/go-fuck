package entrypoints

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"testing"
	"unicode/utf8"
)

// FuzzJSONOutput checks that arbitrary command lines, captured output and
// suggestions always produce a valid JSON document within the output limit.
func FuzzJSONOutput(f *testing.F) {
	f.Add("git brnch", "git: 'brnch' is not a git command.\n", "git branch")
	f.Add("", "", "")
	f.Add(string([]byte{0xff}), string([]byte{0xff, 0xfe}), string([]byte{0xff}))
	f.Fuzz(func(t *testing.T, command, output, suggestion string) {
		result := jsonResult{
			SchemaVersion: jsonSchemaVersion,
			Command:       validUTF8(command),
			Suggestions: []jsonSuggestion{{
				Command:  validUTF8(suggestion),
				Priority: 1000,
			}},
		}
		result.Output, _ = truncatedOutput(output)

		data, err := json.Marshal(result, jsontext.WithIndent("  "))
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		if !utf8.Valid(data) {
			t.Fatalf("the JSON document is not valid UTF-8: %q", data)
		}
		if len(result.Output) > jsonOutputLimit {
			t.Fatalf("len(Output) = %d, want at most %d", len(result.Output), jsonOutputLimit)
		}
	})
}
