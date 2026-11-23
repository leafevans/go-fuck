// Package argumentparser parses the go-fuck command line.
//
// The generated shell aliases insert an argument placeholder that separates
// go-fuck's own arguments from the arguments of the wrapped command. Arguments
// after the placeholder are moved to the beginning and a "--" separator is
// inserted so that options of the corrected command are not parsed by go-fuck.
package argumentparser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/conf"
	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/logs"
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
	// OutputFrom is the file that contains the output of the failed command,
	// or "-" for standard input. It replaces re-running the command.
	OutputFrom string
	// NoRerun disables running the original command to collect output.
	NoRerun bool
	// Preview shows suggestions without re-running the original command or applying a fix.
	Preview bool
	// Explain prints the capture and execution plan without re-running the
	// original command, applying side effects or executing a correction.
	Explain bool
	// CaptureMode overrides the configured capture policy.
	CaptureMode string
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
		case arg == "--capture-mode":
			if i+1 >= len(rest) {
				return nil, fmt.Errorf("argument --capture-mode: expected one argument")
			}
			args.CaptureMode = rest[i+1]
			i++
		case strings.HasPrefix(arg, "--capture-mode="):
			args.CaptureMode = strings.TrimPrefix(arg, "--capture-mode=")
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
		case arg == "--output-from":
			if i+1 >= len(rest) {
				return nil, fmt.Errorf("argument --output-from: expected one argument")
			}
			args.OutputFrom = rest[i+1]
			i++
		case strings.HasPrefix(arg, "--output-from="):
			args.OutputFrom = strings.TrimPrefix(arg, "--output-from=")
		case arg == "--json":
			args.JSON = true
		case arg == "--no-rerun":
			args.NoRerun = true
		case arg == "--preview":
			args.Preview = true
		case arg == "--explain":
			args.Explain = true
		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unrecognized arguments: %s", arg)
		default:
			args.Command = append(args.Command, arg)
		}
	}
	if args.Yes && args.Repeat {
		return nil, fmt.Errorf("argument -r/--repeat: not allowed with argument -y/--yes/--yeah/--hard")
	}
	if args.OutputFrom != "" && args.ForceCommand == "" {
		return nil, fmt.Errorf("argument --output-from: requires argument --force-command")
	}
	if args.CaptureMode != "" {
		switch args.CaptureMode {
		case "auto", "instant", "rerun":
		default:
			return nil, fmt.Errorf("argument --capture-mode: invalid choice: %s (choose from auto, instant, rerun)", args.CaptureMode)
		}
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
	{"-h, --help", "show this help and exit"},
	{"-v, --version", "show version information and exit"},
	{"-a, --alias[=ALIAS]", "print the shell alias; ALIAS overrides TF_ALIAS"},
	{"--setup", "write the alias to the shell configuration file"},
	{"-l, --shell-logger FILE", "log shell output to FILE"},
	{"--enable-experimental-instant-mode", "configure instant mode during alias setup"},
	{"-d, --debug", "print debug information to stderr"},
	{"--doctor", "print a diagnostic report of the shell integration"},
	{"--list-rules", "print the names of the built-in rules and exit"},
	{"--force-command CMD", "fix CMD instead of the previous command"},
	{"--output-from FILE", "read failed command output from FILE (or stdin with -) instead of re-running it"},
	{"--no-rerun", "do not re-run the original command to collect its output"},
	{"--capture-mode MODE", "select output capture: auto, instant or rerun"},
	{"--preview", "show suggestions without re-running the original or applying a fix"},
	{"--explain", "explain capture, rules and execution order without executing"},
	{"--json", "print suggestions without executing them"},
	{"-y, --yes, --yeah, --hard", "execute the fixed command without confirmation"},
	{"-r, --repeat", "repeat when the corrected command fails"},
	{"command ...", "command to fix"},
}

// helpOptionWidth is the width of the option column so that the descriptions
// line up.
const helpOptionWidth = 37

// Usage returns the one line usage message.
func Usage() string {
	return UsageFor(os.Stderr)
}

// UsageFor returns the one line usage message for w.
func UsageFor(w io.Writer) string {
	return usageText(logs.IsTerminal(w) && logs.ColorsEnabled())
}

// usageFlags returns the command-line flags without the executable name.
func usageFlags() string {
	return "[-h] [-v] [-a [ALIAS]] [-l FILE] [--setup] " +
		"[--enable-experimental-instant-mode] [-d] [--doctor] [--list-rules] " +
		"[--force-command CMD] [--output-from FILE] [--no-rerun] " +
		"[--capture-mode MODE] [--preview] [--explain] [--json] [-y | -r] [command ...]"
}

// usageText returns the usage line, optionally styled with the ember theme.
func usageText(styled bool) string {
	rest := consts.ExecutableName() + " " + usageFlags()
	if !styled {
		return "usage: " + rest
	}
	executable, flags, _ := strings.Cut(rest, " ")
	return logs.Accent("usage:") + " " + logs.Strong(executable) + " " + logs.Muted(flags)
}

// Help returns the full help message.
func Help() string {
	return HelpFor(os.Stdout)
}

// HelpFor returns the full help message for w.
func HelpFor(w io.Writer) string {
	return helpText(logs.IsTerminal(w) && logs.ColorsEnabled())
}

// helpText returns the full help message, optionally styled with the ember
// theme.
func helpText(styled bool) string {
	app := consts.ExecutableName()
	title := app + " help"
	usage := app + " " + usageFlags()
	if styled {
		title = logs.Strong(app) + " " + logs.Accent("help")
		usage = logs.Strong(app) + " " + logs.Muted(usageFlags())
	}
	heading := func(text string) string {
		if styled {
			return logs.Accent(text)
		}
		return text
	}
	lines := []string{
		title,
		"",
		heading("Usage"),
		"  " + usage,
		"",
		heading("Description"),
		"  Magnificent app which corrects your previous console command.",
		"",
		heading("Options"),
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
