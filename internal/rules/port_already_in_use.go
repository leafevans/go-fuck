package rules

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// portAlreadyInUsePatterns match the port of the "address already in use"
// errors of the supported applications.
var portAlreadyInUsePatterns = []*lazyRegexp{
	newRegexp(`bind on address \('.*', (?P<port>\d+)\)`),
	newRegexp(`Unable to bind [^ ]*:(?P<port>\d+)`),
	newRegexp(`can't listen on port (?P<port>\d+)`),
	newRegexp(`listen EADDRINUSE [^ ]*:(?P<port>\d+)`),
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

// portAlreadyInUseParsePID accepts a single positive TCP listener PID. Multiple
// socket rows for the same process are harmless; different owners are ambiguous.
func portAlreadyInUseParsePID(output string) string {
	pid := ""
	for _, line := range strings.Split(output, newline) {
		fields := strings.Fields(line)
		if len(fields) < 10 || fields[len(fields)-1] != "(LISTEN)" || fields[7] != "TCP" {
			continue
		}
		candidate := fields[1]
		n, err := strconv.ParseUint(candidate, 10, 31)
		if err != nil || n == 0 {
			return ""
		}
		if pid != "" && pid != candidate {
			return ""
		}
		pid = candidate
	}
	return pid
}

// portAlreadyInUsePID queries listeners, not connected clients or UDP sockets.
// Do not cache PIDs: a process can exit between matching and suggestion creation.
var portAlreadyInUsePID = func(port string) string {
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil || n == 0 {
		return ""
	}
	output, err := exec.Command("lsof", "-nP", "-a", "-iTCP:"+port, "-sTCP:LISTEN").Output()
	if err != nil {
		return ""
	}
	return portAlreadyInUseParsePID(string(output))
}

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
	if pid == "" {
		return nil, nil
	}
	return []string{shells.Current().And("kill "+pid, command.Script)}, nil
}

func init() {
	rule := types.NewRule("port_already_in_use",
		portAlreadyInUseMatch,
		portAlreadyInUseNewCommand)
	// Deviation: occupying a port does not establish consent to terminate its
	// owner. Preserve the legacy correction only through explicit opt-in.
	rule.EnabledByDefault = false
	rule.Reason = "Terminate the TCP listener; it may belong to another application"
	register(rule)
}
