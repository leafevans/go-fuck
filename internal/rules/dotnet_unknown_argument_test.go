package rules

import "testing"

// dotnetToolUnknownArgument is the output of "dotnet tool instal".
const dotnetToolUnknownArgument = `'instal' was not matched. Did you mean one of the following?
install

Unrecognized command or argument 'instal'.

Description:
  Install or work with tools that extend the .NET experience.

Usage:
  dotnet tool [command] [options]

Options:
  -?, -h, --help  Show command line help.

Commands:
  install <packageId>                           Install global or local tool.
  uninstall <PACKAGE_ID>                        Uninstall a global tool or local tool.
`

// dotnetNugetUnknownArgument is the output of "dotnet nuget foo".
const dotnetNugetUnknownArgument = `'foo' was not matched. Did you mean one of the following?
why
add
-h

Required command was not provided.
Unrecognized command or argument 'foo'.

Usage:
  NuGet.CommandLine.XPlat [command] [options]

Commands:
  why <PROJECT|SOLUTION|FILE> <PACKAGE>  Shows the dependency graph for a particular package for a given project or solution.
  add                                    Add a NuGet source.
`

// dotnetUnknownArgumentNoSuggestions is printed when the misspelled argument
// has no close match.
const dotnetUnknownArgumentNoSuggestions = `'zzz' was not matched. Did you mean one of the following?

Unrecognized command or argument 'zzz'.

Usage:
  dotnet tool [command] [options]
`

func TestRuleDotnetUnknownArgument(t *testing.T) {
	resetTestState(t)
	rule := Get("dotnet_unknown_argument")
	if rule == nil {
		t.Fatal("dotnet_unknown_argument rule is not registered")
	}

	assertMatch(t, rule, "dotnet tool instal", dotnetToolUnknownArgument)
	assertMatch(t, rule, "dotnet workload instal", dotnetToolUnknownArgument)
	assertMatch(t, rule, "dotnet nuget foo", dotnetNugetUnknownArgument)

	assertNotMatch(t, rule, "dotnet tool install", "")
	assertNotMatch(t, rule, "dotnet buidl", "Could not execute because the specified command or file was not found.")
	assertNotMatch(t, rule, "dotnet zzz", dotnetUnknownArgumentNoSuggestions)
	assertNotMatch(t, rule, "git tool instal", dotnetToolUnknownArgument)

	assertNewCommands(t, rule, "dotnet tool instal", dotnetToolUnknownArgument,
		[]string{"dotnet tool install"})
	assertNewCommands(t, rule, "dotnet workload instal", dotnetToolUnknownArgument,
		[]string{"dotnet workload install"})
	assertNewCommands(t, rule, "dotnet nuget foo", dotnetNugetUnknownArgument,
		[]string{"dotnet nuget why", "dotnet nuget add", "dotnet nuget -h"})
}
