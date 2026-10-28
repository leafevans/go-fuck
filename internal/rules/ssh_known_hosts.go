package rules

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// sshKnownHostsCommands lists the ssh commands that report host key errors.
var sshKnownHostsCommands = []string{"ssh", "scp"}

// sshKnownHostsPatterns match the warnings about a changed host key.
var sshKnownHostsPatterns = []*regexp.Regexp{
	regexp.MustCompile("WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!"),
	regexp.MustCompile("WARNING: POSSIBLE DNS SPOOFING DETECTED!"),
	regexp.MustCompile(`Warning: the \S+ host key for '([^']+)' differs from the key for the IP address '([^']+)'`),
}

// sshKnownHostsOffendingPatterns match the lines that name the offending key.
var sshKnownHostsOffendingPatterns = []*regexp.Regexp{
	regexp.MustCompile(`Offending (key for IP|\S+ key) in ([^:]+):(\d+)`),
	regexp.MustCompile(`Matching host key in ([^:]+):(\d+)`),
}

// sshKnownHostsOffending describes one known_hosts line that must be removed.
type sshKnownHostsOffending struct {
	path string
	line int
}

// sshKnownHostsMatch reports whether ssh reported a changed host key.
func sshKnownHostsMatch(command *types.Command) bool {
	if command.Script == "" {
		return false
	}
	matchedCommand := false
	for _, name := range sshKnownHostsCommands {
		if strings.HasPrefix(command.Script, name) {
			matchedCommand = true
			break
		}
	}
	if !matchedCommand {
		return false
	}

	for _, pattern := range sshKnownHostsPatterns {
		if pattern.MatchString(command.Output) {
			return true
		}
	}
	return false
}

// sshKnownHostsNewCommand returns the original command unchanged: the fix is
// the side effect that removes the outdated key.
func sshKnownHostsNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script}, nil
}

// sshKnownHostsOffendingEntries extracts the files and line numbers of the
// outdated known_hosts entries.
func sshKnownHostsOffendingEntries(commandOutput string) []sshKnownHostsOffending {
	entries := []sshKnownHostsOffending{}
	for _, pattern := range sshKnownHostsOffendingPatterns {
		for _, match := range pattern.FindAllStringSubmatch(commandOutput, -1) {
			path, line := match[len(match)-2], match[len(match)-1]
			number, err := strconv.Atoi(line)
			if err != nil {
				continue
			}
			entries = append(entries, sshKnownHostsOffending{path: path, line: number})
		}
	}
	return entries
}

// sshKnownHostsSideEffect removes the outdated keys from the known_hosts
// files.
func sshKnownHostsSideEffect(oldCommand *types.Command, script string) error {
	for _, entry := range sshKnownHostsOffendingEntries(oldCommand.Output) {
		content, err := os.ReadFile(entry.path)
		if err != nil {
			return err
		}

		lines := strings.SplitAfter(string(content), newline)
		// A trailing newline produces an empty final element that is not a
		// line of its own.
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		if entry.line < 1 || entry.line > len(lines) {
			continue
		}
		lines = append(lines[:entry.line-1], lines[entry.line:]...)
		if err := os.WriteFile(entry.path, []byte(strings.Join(lines, "")), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rule := types.NewRule("ssh_known_hosts",
		utils.ForApp(sshKnownHostsCommands, 0, sshKnownHostsMatch),
		sshKnownHostsNewCommand)
	rule.SideEffect = sshKnownHostsSideEffect
	register(rule)
}
