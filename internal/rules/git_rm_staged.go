package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRmStagedMatch reports whether git rm refused to delete a file with
// staged changes.
func gitRmStagedMatch(command *types.Command) bool {
	return strings.Contains(command.Script, " rm ") &&
		strings.Contains(command.Output,
			"error: the following file has changes staged in the index") &&
		strings.Contains(command.Output,
			"use --cached to keep the file, or -f to force removal")
}

// gitRmStagedNewCommand suggests keeping the file with --cached or deleting it
// with -f.
func gitRmStagedNewCommand(command *types.Command) ([]string, error) {
	parts := append([]string{}, command.ScriptParts()...)
	index := utils.Index(parts, "rm")
	if index < 0 {
		return nil, nil
	}
	index++

	cachedParts := make([]string, 0, len(parts)+1)
	cachedParts = append(cachedParts, parts[:index]...)
	cachedParts = append(cachedParts, "--cached")
	cachedParts = append(cachedParts, parts[index:]...)
	cached := strings.Join(cachedParts, " ")

	cachedParts[index] = "-f"
	forced := strings.Join(cachedParts, " ")
	return []string{cached, forced}, nil
}

func init() {
	register(types.NewRule("git_rm_staged",
		specific.GitSupportMatch(gitRmStagedMatch),
		specific.GitSupportNew(gitRmStagedNewCommand)))
}
