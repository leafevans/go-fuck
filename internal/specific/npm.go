package specific

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/internal/utils"
)

// npmScriptLine matches the indented script lines of "npm run-script" output.
var npmScriptLine = regexp.MustCompile("^  [^ ]+")

// PackageScripts returns the script names of the package.json file in the
// current working directory, sorted by name.
//
// It is memoized because rules ask for the scripts while matching and while
// building the corrected command. A missing or unparsable package.json yields
// no scripts.
var PackageScripts = utils.Memoize(func() []string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil || len(pkg.Scripts) == 0 {
		return nil
	}
	scripts := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		scripts = append(scripts, name)
	}
	sort.Strings(scripts)
	return scripts
})

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
