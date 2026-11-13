// Package argumentparser parses the go-fuck command line.
//
// The generated shell aliases insert an argument placeholder that separates
// go-fuck's own arguments from the arguments of the wrapped command. Arguments
// after the placeholder are moved to the beginning and a "--" separator is
// inserted so that options of the corrected command are not parsed by go-fuck.
package argumentparser

import (
	"fmt"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
)

// Args holds the parsed command line arguments.
type Args struct {
	// Version reports whether --version was requested.
	Version bool
	// Help reports whether --help was requested.
	Help bool
	// Setup reports whether --setup was requested.
	Setup bool
	// Debug enables debug output.
	Debug bool
	// Doctor reports whether --doctor was requested.
	Doctor bool
	// ListRules reports whether --list-rules was requested.
	ListRules bool
	// Yes executes the corrected command without confirmation.
	Yes bool
	// Repeat re-runs go-fuck when the corrected command fails.
	Repeat bool
	// EnableExperimentalInstantMode requests the instant mode. The flag name is
	// kept for compatibility.
	EnableExperimentalInstantMode bool
	// Alias is the alias name requested with --alias.
	Alias string
	// AliasRequested reports whether --alias was given.
	AliasRequested bool
	// ShellLogger is the output file of the shell logger.
	ShellLogger string
	// ForceCommand is the command to fix when --force-command is given.
	ForceCommand string
	// JSON prints the suggestions as JSON instead of running the interactive
	// selection and the corrected command.
	JSON bool
	// Command is the command to fix, as passed by the shell alias.
	Command []string
}

// Parse parses the command line arguments, including the program name.
func Parse(argv []string) (*Args, error) {
	args := &Args{}
	if len(argv) == 0 {
		return args, nil
	}

	rest := prepareArguments(argv[1:])
	positional := false
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		switch {
		case positional:
			args.Command = append(args.Command, arg)
		case arg == "--":
			positional = true
		case arg == "-v" || arg == "--version":
			args.Version = true
		case arg == "-h" || arg == "--help":
			args.Help = true
		case arg == "--setup":
			args.Setup = true
		case arg == "-d" || arg == "--debug":
			args.Debug = true
		case arg == "--doctor":
			args.Doctor = true
		case arg == "--list-rules":
			args.ListRules = true
		case arg == "-y" || arg == "--yes" || arg == "--yeah" || arg == "--hard":
			args.Yes = true
		case arg == "-r" || arg == "--repeat":
			args.Repeat = true
		case arg == "--enable-experimental-instant-mode":
			args.EnableExperimentalInstantMode = true
		case arg == "-a" || arg == "--alias":
			args.AliasRequested = true
			args.Alias = conf.Alias()
			if i+1 < len(rest) && !strings.HasPrefix(rest[i+1], "-") {
				args.Alias = rest[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--alias="):
			args.AliasRequested = true
			args.Alias = strings.TrimPrefix(arg, "--alias=")
		case arg == "-l" || arg == "--shell-logger":
			if i+1 >= len(rest) {
				return nil, fmt.Errorf("argument -l/--shell-logger: expected one argument")
			}
			args.ShellLogger = rest[i+1]
			i++
		case strings.HasPrefix(arg, "--shell-logger="):
			args.ShellLogger = strings.TrimPrefix(arg, "--shell-logger=")
		case arg == "--force-command":
			if i+1 >= len(rest) {
				return nil, fmt.Errorf("argument --force-command: expected one argument")
			}
			args.ForceCommand = rest[i+1]
			i++
		case strings.HasPrefix(arg, "--force-command="):
			args.ForceCommand = strings.TrimPrefix(arg, "--force-command=")
		case arg == "--json":
			args.JSON = true
		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unrecognized arguments: %s", arg)
		default:
			args.Command = append(args.Command, arg)
		}
	}
	if args.Yes && args.Repeat {
		return nil, fmt.Errorf("argument -r/--repeat: not allowed with argument -y/--yes/--yeah/--hard")
	}
	return args, nil
}

// prepareArguments moves the arguments after the placeholder to the beginning
// and inserts the "--" separator before the command arguments.
func prepareArguments(argv []string) []string {
	for index, arg := range argv {
		if arg == consts.ArgumentPlaceholder || arg == consts.LegacyArgumentPlaceholder {
			prepared := append([]string{}, argv[index+1:]...)
			prepared = append(prepared, "--")
			return append(prepared, argv[:index]...)
		}
	}
	if len(argv) > 0 && !strings.HasPrefix(argv[0], "-") && argv[0] != "--" {
		return append([]string{"--"}, argv...)
	}
	return argv
}

// helpOption is one option row of the help output.
type helpOption struct {
	flags       string
	description string
}

// helpOptions lists the command line options in help order.
var helpOptions = []helpOption{
	{"-v, --version", "show program's version number and exit"},
	{"-a, --alias [ALIAS]", "[custom-alias-name] prints alias for current shell"},
	{"--setup", "configure the shell alias in the shell configuration file"},
	{"-l, --shell-logger SHELL_LOGGER", "log shell output to the file"},
	{"--enable-experimental-instant-mode", "enable instant mode for the alias initialization"},
	{"-h, --help", "show this help message and exit"},
	{"-d, --debug", "enable debug output"},
	{"--doctor", "print a diagnostic report of the shell integration"},
	{"--list-rules", "print the names of the built-in rules and exit"},
	{"--force-command FORCE_COMMAND", "fix this command instead of the previous one"},
	{"--json", "print the suggestions as JSON and exit"},
	{"-y, --yes, --yeah, --hard", "execute fixed command without confirmation"},
	{"-r, --repeat", "repeat on failure"},
	{"command", "command that should be fixed"},
}

// helpOptionWidth is the width of the option column so that the descriptions
// line up.
const helpOptionWidth = 37

// Usage returns the one line usage message.
func Usage() string {
	return usageText(logs.OutputIsTerminal() && logs.ColorsEnabled())
}

// usageText returns the usage line, optionally styled with the ember theme.
func usageText(styled bool) string {
	rest := consts.ExecutableName() + " [-v] [-a [ALIAS]] [-l SHELL_LOGGER] [--setup] " +
		"[--enable-experimental-instant-mode] [-h] [-d] [--doctor] [--list-rules] " +
		"[--force-command FORCE_COMMAND] [--json] [-y | -r] [command ...]"
	if !styled {
		return "usage: " + rest
	}
	executable, flags, _ := strings.Cut(rest, " ")
	return logs.Accent("usage:") + " " + logs.Strong(executable) + " " + logs.Muted(flags)
}

// Help returns the full help message.
func Help() string {
	return helpText(logs.OutputIsTerminal() && logs.ColorsEnabled())
}

// helpText returns the full help message, optionally styled with the ember
// theme.
func helpText(styled bool) string {
	lines := []string{
		usageText(styled),
		"",
		"Magnificent app which corrects your previous console command.",
		"",
		"options:",
	}
	if styled {
		lines[2] = logs.Muted(lines[2])
		lines[4] = logs.Accent(lines[4])
	}
	for _, option := range helpOptions {
		flags := fmt.Sprintf("%-*s", helpOptionWidth, option.flags)
		if styled {
			lines = append(lines, "  "+logs.Accent(flags)+logs.Muted(option.description))
			continue
		}
		lines = append(lines, "  "+flags+option.description)
	}
	lines = append(lines, "")
	return strings.Join(lines, string(rune(10)))
}
