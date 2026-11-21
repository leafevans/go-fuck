package rules

import (
	"path/filepath"

	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitWorktreeRemoveMissingPattern matches the error that git prints when
// "git worktree remove" is used for a path that is not a working tree.
var gitWorktreeRemoveMissingPattern = newRegexp(`(?m)^fatal: '([^']+)' is not a working tree$`)

// gitWorktreePathsCommand lists the registered worktree paths. It is a
// variable so that tests do not depend on the repository that runs them.
var gitWorktreePathsCommand = func(script string) []string {
	output, err := gitRepositoryOutput(script, "worktree", "worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	return parseWorktreePaths(string(output))
}

// parseWorktreePaths extracts the secondary worktree paths from porcelain
// output. The first entry is the main working tree, which "git worktree
// remove" cannot remove.
func parseWorktreePaths(output string) []string {
	paths := []string{}
	first := true
	for line := range strings.SplitSeq(output, "\n") {
		path, ok := strings.CutPrefix(line, "worktree ")
		if !ok {
			continue
		}
		if first {
			first = false
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

// canonicalWorktreePath returns a cleaned absolute path, so that a relative
// path typed by the user and the absolute path printed by git can be compared.
func canonicalWorktreePath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(absolute)
}

// closestWorktreePath returns the original worktree path that is closest to
// the path from the error. The directory names are compared instead of the
// whole paths, because all worktrees usually share the same parent directory.
func closestWorktreePath(script, missing string) string {
	paths := gitWorktreePathsCommand(script)
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, filepath.Base(canonicalWorktreePath(path)))
	}
	closest := utils.GetClosest(filepath.Base(canonicalWorktreePath(missing)), names, 0.6, false)
	if closest == "" {
		return ""
	}
	for i, name := range names {
		if name == closest {
			return paths[i]
		}
	}
	return ""
}

// gitWorktreeRemoveMissingMatch reports whether "git worktree remove" failed
// because the path is not a worktree and another worktree path is close to it.
func gitWorktreeRemoveMissingMatch(command *types.Command) bool {
	_, ok := gitWorktreeRemoveMissingCorrection(command)
	return ok
}

// gitWorktreeRemoveMissingNewCommand replaces the missing path with the
// closest registered worktree.
func gitWorktreeRemoveMissingNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitWorktreeRemoveMissingCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func gitWorktreeRemoveMissingCorrection(command *types.Command) (string, bool) {
	if !gitHasOperation(command.Script, "worktree", "remove") {
		return "", false
	}
	missing, ok := gitWorktreeRemoveMissingName(command.Output)
	if !ok || missing == "" {
		return "", false
	}
	start, end, ok := specific.GitArgumentSpan(command.Script, missing, shells.Current(), true)
	if !ok {
		return "", false
	}
	closest := closestWorktreePath(command.Script, missing)
	if closest == "" {
		return "", false
	}
	return command.Script[:start] + shells.Current().Quote(closest) + command.Script[end:], true
}

// gitWorktreeRemoveMissingName returns the path from a worktree error.
func gitWorktreeRemoveMissingName(output string) (string, bool) {
	if match := gitWorktreeRemoveMissingPattern.FindStringSubmatch(output); match != nil {
		return match[1], true
	}
	return "", false
}

func init() {
	register(types.NewRule("git_worktree_remove_missing",
		utils.ForApp([]string{"git"}, 0, gitWorktreeRemoveMissingMatch),
		gitWorktreeRemoveMissingNewCommand))
}
