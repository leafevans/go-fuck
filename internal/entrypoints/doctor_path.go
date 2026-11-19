package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// doctorPathCheck reports whether the shell alias resolves to the binary that
// is running.
//
// The alias runs the executable by name, so an older or differently installed
// copy that comes first in PATH is the one the user gets; this is how an old
// 0.x/1.x installation silently shadows a v2 release.
func doctorPathCheck(report doctorReport) doctorCheck {
	check := doctorCheck{Name: "path"}
	if report.Binary == "" {
		check.Status = doctorInfo
		check.Message = "not checked: the executable path is unknown"
		return check
	}

	running := doctorResolvePath(report.Binary)
	candidates := doctorPathCandidates(report.Binary, running)
	if len(candidates) == 0 {
		check.Status = doctorWarn
		check.Message = "the running binary is not on PATH: " + report.Binary
		check.Hint = "the shell alias starts " + filepath.Base(report.Binary) +
			" from PATH, so add " + filepath.Dir(report.Binary) + " to PATH"
		return check
	}

	first := candidates[0]
	if first.path == running {
		if len(candidates) == 1 {
			check.Status = doctorOK
			check.Message = "resolves to this build: " + first.display
			return check
		}
		others := make([]string, 0, len(candidates)-1)
		for _, candidate := range candidates[1:] {
			others = append(others, candidate.display)
		}
		check.Status = doctorInfo
		check.Message = "resolves to this build: " + first.display
		check.Hint = "other copies later in PATH: " + strings.Join(others, ", ")
		return check
	}

	check.Status = doctorWarn
	check.Message = "another " + filepath.Base(report.Binary) + " is earlier in PATH: " + first.display
	check.Hint = "the shell alias starts that binary, not this build; remove or rename it, or put " +
		filepath.Dir(report.Binary) + " first in PATH"
	return check
}

// doctorPathCandidate is one executable found while scanning PATH.
type doctorPathCandidate struct {
	// path is the resolved absolute path used for comparisons.
	path string
	// display is the path as it appears in PATH, for the report.
	display string
}

// doctorPathCandidates scans PATH for executables with the name of the running
// binary and returns them in PATH order, without duplicates.
func doctorPathCandidates(running, resolvedRunning string) []doctorPathCandidate {
	names := doctorCandidateNames(running)
	candidates := make([]doctorPathCandidate, 0, 4)
	seen := map[string]bool{}
	for _, directory := range filepath.SplitList(os.Getenv("PATH")) {
		if directory == "" {
			continue
		}
		for _, name := range names {
			path := filepath.Join(directory, name)
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			resolved := doctorResolvePath(path)
			if seen[resolved] {
				continue
			}
			seen[resolved] = true
			candidates = append(candidates, doctorPathCandidate{path: resolved, display: path})
		}
	}
	return candidates
}

// doctorCandidateNames returns the file names under which the running binary
// can be found in PATH.
func doctorCandidateNames(running string) []string {
	name := filepath.Base(running)
	names := []string{name}
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		names = append(names, name+".exe")
	}
	return names
}

// doctorResolvePath returns an absolute, symlink free path for comparisons.
func doctorResolvePath(path string) string {
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	return path
}
