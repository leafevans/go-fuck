package entrypoints

import (
	"encoding/json"
	"os"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/types"
)

// jsonOutputLimit is the maximum number of output bytes included in the JSON
// result. Editor plugins and AI agents usually only need the tail of the
// output, so the JSON output is capped instead of growing without a bound.
const jsonOutputLimit = 32 * 1024

// jsonSuggestion is one corrected command of the machine readable output.
type jsonSuggestion struct {
	// Command is the corrected command line.
	Command string `json:"command"`
	// Priority orders the suggestions. Lower values are suggested first.
	Priority int `json:"priority"`
	// SideEffect reports whether the normal interactive mode would run a side
	// effect (for example "git add") before the command.
	SideEffect bool `json:"side_effect"`
}

// jsonResult is the machine readable result of a correction request.
type jsonResult struct {
	// Command is the command line that was fixed, with shell aliases expanded.
	Command string `json:"command"`
	// Output is the captured standard output and standard error of the failed
	// command. It is omitted when the output could not be captured and empty
	// when the command printed nothing.
	Output string `json:"output,omitempty"`
	// OutputTruncated reports whether Output was cut at jsonOutputLimit bytes.
	OutputTruncated bool `json:"output_truncated,omitempty"`
	// Suggestions are the corrected commands, in the order in which they would
	// be offered.
	Suggestions []jsonSuggestion `json:"suggestions"`
}

// printJSON writes the correction result as JSON.
//
// Unlike the interactive mode it does not run the corrected command, update
// the shell history or apply rule side effects.
func printJSON(command *types.Command, corrected []*types.CorrectedCommand) int {
	result := jsonResult{Suggestions: make([]jsonSuggestion, 0, len(corrected))}
	if command != nil {
		result.Command = command.Script
		if command.HasOutput() {
			result.Output, result.OutputTruncated = truncateJSONOutput(command.Output)
		}
	}
	for _, suggestion := range corrected {
		result.Suggestions = append(result.Suggestions, jsonSuggestion{
			Command:    suggestion.Script,
			Priority:   suggestion.Priority,
			SideEffect: suggestion.SideEffect != nil,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		logs.Exception("Can't write the JSON output", err)
		return 1
	}
	return 0
}

// truncateJSONOutput limits output to jsonOutputLimit bytes without splitting a
// UTF-8 sequence. The second result reports whether the output was truncated.
func truncateJSONOutput(output string) (string, bool) {
	if len(output) <= jsonOutputLimit {
		return output, false
	}
	end := jsonOutputLimit
	for end > 0 && !utf8.RuneStart(output[end]) {
		end--
	}
	return output[:end], true
}
