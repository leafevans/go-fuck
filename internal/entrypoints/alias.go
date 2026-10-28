package entrypoints

import (
	"fmt"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/shells"
)

// printAlias prints the alias definition for the current shell.
func printAlias(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat}); err != nil {
		logs.Exception("Can't load settings", err)
	}

	if args.EnableExperimentalInstantMode {
		logs.Warn("Instant mode is not supported by go-fuck, using the regular alias")
	}

	fmt.Println(shells.Current().AppAlias(args.Alias))
	return 0
}
