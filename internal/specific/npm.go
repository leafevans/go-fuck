package specific

import (
	json "encoding/json/v2"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/utils"
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

// npmRunScriptOutput runs "npm run-script", which lists the scripts of the
// current project. It is a variable so that tests do not need npm.
var npmRunScriptOutput = func() ([]byte, error) {
	return exec.Command("npm", "run-script").Output()
}

// NpmGetScripts returns the custom npm scripts of the current project.
//
// The result is memoized because npm is started once per go-fuck invocation.
// npm 6 lists every script under "Scripts available in ... via ...", while
// npm 7 and newer split the list into the lifecycle scripts and the scripts
// available via "npm run". The package.json file is used as a fallback so the
// suggestions keep working when npm prints an unexpected format or is missing.
var NpmGetScripts = utils.Memoize(npmScripts)

// npmScripts returns the scripts reported by npm, falling back to the
// package.json file.
func npmScripts() []string {
	output, err := npmRunScriptOutput()
	if err == nil {
		if scripts := parseNpmScripts(string(output)); len(scripts) > 0 {
			return scripts
		}
	}
	return PackageScripts()
}

// parseNpmScripts extracts the script names from the output of "npm run-script"
// and "npm run", accepting the npm 6 and npm 7+ layouts.
func parseNpmScripts(output string) []string {
	var scripts []string
	listing := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Lifecycle scripts included in") ||
			strings.HasPrefix(trimmed, "Scripts available in") ||
			strings.Contains(trimmed, "available via ") {
			listing = true
			continue
		}
		if !listing {
			continue
		}
		if npmScriptLine.MatchString(line) {
			scripts = append(scripts, strings.Fields(line)[0])
			continue
		}
		if trimmed == "" || strings.HasPrefix(line, "    ") {
			// A blank line or the description under a script name.
			continue
		}
		// Any other non-indented line ends the listing.
		listing = false
	}
	return scripts
}
