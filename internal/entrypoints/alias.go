package entrypoints

import (
	"fmt"
	"os"

	"github.com/leafevans/go-fuck/v2/internal/argumentparser"
	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/shells"
)

// printAlias prints the alias definition for the current shell.
func printAlias(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat}); err != nil {
		logs.Exception("Can't load settings", err)
	}

	if args.EnableExperimentalInstantMode {
		fmt.Fprintln(os.Stdout, shells.Current().InstantModeAlias(args.Alias))
		return 0
	}

	fmt.Fprintln(os.Stdout, shells.Current().AppAlias(args.Alias))
	return 0
}
