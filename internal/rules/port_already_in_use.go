package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// portAlreadyInUsePatterns match the port of the "address already in use"
// errors of the supported applications.
var portAlreadyInUsePatterns = []*regexp.Regexp{
	regexp.MustCompile(`bind on address \('.*', (?P<port>\d+)\)`),
	regexp.MustCompile(`Unable to bind [^ ]*:(?P<port>\d+)`),
	regexp.MustCompile(`can't listen on port (?P<port>\d+)`),
	regexp.MustCompile(`listen EADDRINUSE [^ ]*:(?P<port>\d+)`),
}

// portAlreadyInUseParsePort returns the port reported by an error output.
func portAlreadyInUseParsePort(output string) string {
	for _, pattern := range portAlreadyInUsePatterns {
		if match := pattern.FindStringSubmatch(output); match != nil {
			return match[pattern.SubexpIndex("port")]
		}
	}
	return ""
}

// portAlreadyInUseUsedPort returns the port reported by the command output.
var portAlreadyInUseUsedPort = utils.Memoize1(portAlreadyInUseParsePort)

// portAlreadyInUseParsePID extracts the PID from the output of lsof.
func portAlreadyInUseParsePID(output string) string {
	lines := strings.Split(output, newline)
	if len(lines) <= 1 {
		return ""
	}
	fields := strings.Fields(lines[1])
	if len(fields) <= 1 {
		return ""
	}
	return fields[1]
}

// portAlreadyInUsePID returns the PID of the process that listens on the port.
var portAlreadyInUsePID = utils.Memoize1(func(port string) string {
	output, _ := exec.Command("lsof", "-i", ":"+port).Output()
	return portAlreadyInUseParsePID(string(output))
})

// portAlreadyInUseMatch reports whether the port of a failed bind is used by
// another process.
func portAlreadyInUseMatch(command *types.Command) bool {
	port := portAlreadyInUseUsedPort(command.Output)
	return port != "" && portAlreadyInUsePID(port) != ""
}

// portAlreadyInUseNewCommand kills the process that uses the port.
func portAlreadyInUseNewCommand(command *types.Command) ([]string, error) {
	port := portAlreadyInUseUsedPort(command.Output)
	pid := portAlreadyInUsePID(port)
	return []string{shells.Current().And("kill "+pid, command.Script)}, nil
}

func init() {
	rule := types.NewRule("port_already_in_use",
		portAlreadyInUseMatch,
		portAlreadyInUseNewCommand)
	// Python: enabled_by_default = bool(which('lsof')).
	rule.EnabledByDefault = utils.Which("lsof") != ""
	register(rule)
}
