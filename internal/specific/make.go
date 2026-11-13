package specific

import (
	"os"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// makefileNames are the file names GNU make searches, in order.
var makefileNames = []string{"GNUmakefile", "makefile", "Makefile"}

// MakeTargets returns the targets defined in the Makefile of the current
// working directory, sorted by name.
//
// Only the first Makefile that exists is parsed; included files and variable
// expanded target names are not resolved. The result is memoized because the
// file do not change while go-fuck runs.
var MakeTargets = utils.Memoize(func() []string {
	for _, name := range makefileNames {
		content, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		return parseMakeTargets(string(content))
	}
	return nil
})

// parseMakeTargets returns the literal targets of a Makefile content.
func parseMakeTargets(content string) []string {
	targets := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ".PHONY:") {
			for _, target := range strings.Fields(strings.TrimPrefix(line, ".PHONY:")) {
				if validMakeTarget(target) {
					targets[target] = true
				}
			}
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 || strings.HasPrefix(line[colon:], ":=") || strings.HasPrefix(line[colon:], "::=") {
			continue
		}
		targetPart := line[:colon]
		if comment := strings.IndexByte(targetPart, '#'); comment >= 0 {
			targetPart = targetPart[:comment]
		}
		for _, target := range strings.Fields(targetPart) {
			if validMakeTarget(target) {
				targets[target] = true
			}
		}
	}

	result := make([]string, 0, len(targets))
	for target := range targets {
		result = append(result, target)
	}
	sort.Strings(result)
	return result
}

// validMakeTarget reports whether name looks like a literal Make target.
func validMakeTarget(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") || strings.ContainsAny(name, "$%#") {
		return false
	}
	for _, character := range name {
		switch {
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case character >= '0' && character <= '9':
		case character == '_', character == '-', character == '.', character == '/':
		default:
			return false
		}
	}
	return true
}
