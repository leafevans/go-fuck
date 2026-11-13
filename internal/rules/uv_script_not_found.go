package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// uvFailedSpawn matches the name that uv could not spawn.
var uvFailedSpawn = regexp.MustCompile(`Failed to spawn: ([^\r\n]+)`)

// uvScriptCandidates returns the names of the files in the current directory.
var uvScriptCandidates = func() []string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	return names
}

// uvScriptClosest returns the file that is closest to name, or an empty string.
func uvScriptClosest(name string) string {
	return utils.GetClosest(name, uvScriptCandidates(), 0.6, false)
}

// uvScriptNotFoundMatch reports whether uv run failed to spawn a missing file.
func uvScriptNotFoundMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 || !strings.EqualFold(parts[1], "run") {
		return false
	}
	if !strings.Contains(command.Output, "Failed to spawn:") ||
		!strings.Contains(command.Output, "program not found") {
		return false
	}
	match := uvFailedSpawn.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	name := strings.TrimSpace(match[1])
	return name != "" && uvScriptClosest(name) != ""
}

// uvScriptNotFoundNewCommand replaces the missing file with the closest one.
func uvScriptNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := uvFailedSpawn.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	name := strings.TrimSpace(match[1])
	closest := uvScriptClosest(name)
	if closest == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, name, closest)}, nil
}

func init() {
	register(types.NewRule("uv_script_not_found",
		utils.ForApp([]string{"uv"}, 1, uvScriptNotFoundMatch),
		uvScriptNotFoundNewCommand))
}
