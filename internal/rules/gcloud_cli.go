package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

var gcloudInvalidChoice = newRegexp(`^ERROR: \(gcloud((?:\.[a-z][a-z0-9-]*)*)\) Invalid choice: '([a-z][a-z0-9-]*)'\.$`)
var gcloudCommandWord = newRegexp("^[a-z][a-z0-9-]*$")

func gcloudCorrections(command *types.Command) []string {
	if !strings.Contains(command.Output, "Maybe you meant:") {
		return nil
	}
	words := simpleCommandWords(command.Script)
	if len(words) < 2 || words[0].text != "gcloud" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(command.Output, "\r\n", "\n"), "\n")
	for i, line := range lines {
		match := gcloudInvalidChoice.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if i+1 >= len(lines) || lines[i+1] != "Maybe you meant:" {
			return nil
		}
		prefix := strings.Split("gcloud"+match[1], ".")
		index := len(prefix)
		if len(words) <= index || words[index].text != match[2] {
			return nil
		}
		for j, part := range prefix {
			if words[j].text != part {
				return nil
			}
		}
		var results []string
		seen := make(map[string]bool)
		for _, option := range lines[i+2:] {
			if !strings.HasPrefix(option, "  ") {
				break
			}
			parts := strings.Fields(option)
			if len(parts) != index+1 {
				continue
			}
			same := true
			for j, part := range prefix {
				if parts[j] != part {
					same = false
					break
				}
			}
			replacement := parts[index]
			if same && gcloudCommandWord.MatchString(replacement) && !seen[replacement] {
				results = append(results, replaceSubcommand(command.Script, words, index, match[2], replacement)...)
				seen[replacement] = true
			}
		}
		return results
	}
	return nil
}

func init() {
	register(types.NewRule("gcloud_cli", func(command *types.Command) bool {
		return len(gcloudCorrections(command)) > 0
	}, func(command *types.Command) ([]string, error) {
		return gcloudCorrections(command), nil
	}))
}
