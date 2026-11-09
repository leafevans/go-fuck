// Package entrypoints implements the go-fuck command line entry points.
package entrypoints

import (
	"fmt"
	"os"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
)

// Main is the entry point of the go-fuck command. It returns the process exit
// code.
func Main() int {
	return recoverEntrypoints(dispatch)
}

// dispatch parses the command line and runs the requested entry point.
func dispatch() int {
	consts.SetExecutableName(os.Args[0])
	system.InitOutput()
	shells.Set(shells.Detect())

	args, err := argumentparser.Parse(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, argumentparser.Usage())
		return 2
	}

	switch {
	case args.Help:
		fmt.Fprintln(os.Stderr, argumentparser.Help())
		return 0
	case args.Version:
		logs.PrintVersion(versionInfo(logs.OutputIsTerminal() || args.Debug), args.Debug)
		return 0
	case args.ListRules:
		return listRules(args)
	case args.Doctor:
		return doctor(args)
	case args.ShellLogger != "":
		return shellLogger(args.ShellLogger)
	case args.AliasRequested:
		return printAlias(args)
	case args.Setup:
		return setup(args)
	case len(args.Command) > 0 || args.ForceCommand != "" || os.Getenv("TF_HISTORY") != "":
		return fixCommand(args)
	default:
		fmt.Fprintln(os.Stderr, argumentparser.Usage())
		return 0
	}
}

// recoverEntrypoints runs fn and converts a panic into a friendly error so
// that an internal error does not print a raw stack trace in the shell.
func recoverEntrypoints(fn func() int) (code int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintln(os.Stderr, consts.ExecutableName()+": internal error: "+fmt.Sprint(recovered))
			fmt.Fprintln(os.Stderr, "Please report this at https://github.com/leafevans/go-fuck/issues")
			code = 1
		}
	}()
	return fn()
}
