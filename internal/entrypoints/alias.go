package entrypoints

import (
	"fmt"
	"os"

	"github.com/leafevans/go-fuck/v4/internal/argumentparser"
	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/instant"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/shells"
)

// printAlias prints the alias definition for the current shell.
func printAlias(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{
		Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat,
		CaptureMode: args.CaptureMode,
	}); err != nil {
		logs.Exception("Can't load settings", err)
	}

	mode := conf.Current().CaptureMode
	if args.EnableExperimentalInstantMode {
		mode = "instant"
	}
	shell := shells.Current()
	if mode == "auto" && instant.Supported() {
		if configuration := shell.HowToConfigure(); configuration != nil && configuration.InstantContent != "" {
			mode = "instant"
		}
	}
	if mode == "instant" {
		fmt.Fprintln(os.Stdout, shell.InstantModeAlias(args.Alias))
		return 0
	}

	fmt.Fprintln(os.Stdout, shell.AppAlias(args.Alias))
	return 0
}
