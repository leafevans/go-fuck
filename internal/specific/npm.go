package specific

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/utils"
)

// npmScriptLine matches the indented script lines of "npm run-script" output.
var npmScriptLine = regexp.MustCompile("^  [^ ]+")

// NpmGetScripts returns the custom npm scripts of the current project.
//
// The result is memoized because npm is started once per go-fuck invocation.
var NpmGetScripts = utils.Memoize(func() []string {
	output, err := exec.Command("npm", "run-script").Output()
	if err != nil {
		return nil
	}

	scripts := []string{}
	shouldYield := false
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "available via ") {
			shouldYield = true
			continue
		}
		if shouldYield && npmScriptLine.MatchString(line) {
			scripts = append(scripts, strings.Fields(line)[0])
		}
	}
	return scripts
})
