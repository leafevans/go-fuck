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
