package rules

import (
	"slices"
	"testing"
)

// dotnetUnknownCommandOutput is the English output of the .NET CLI for a
// misspelled top level command.
func dotnetUnknownCommandOutput(command string) string {
	return `Could not execute because the specified command or file was not found.
Possible reasons for this include:
  * You misspelled a built-in dotnet command.
  * You intended to execute a .NET program, but dotnet-` + command + ` does not exist.
  * You intended to run a global tool, but a dotnet-prefixed executable with this name could not be found on the PATH.
`
}

// dotnetUnknownCommandLocalizedOutput is the Chinese output of the .NET CLI for
// the same mistake. The message is localized, the dotnet-<command> token is not.
func dotnetUnknownCommandLocalizedOutput(command string) string {
	return `无法执行，因为找不到指定的命令或文件。
可能造成此问题的原因包括:
 *内置 dotnet 命令拼写错误。
 *你打算执行 .NET 程序，但 dotnet-` + command + ` 不存在。
 *你打算运行全局工具，但在 PATH 上找不到具有此名称且带有 dotnet 前缀的可执行文件。
`
}

// dotnetUnknownCommandToolOutput is the output for a misspelled nested
// subcommand. It does not contain the top level dotnet-<command> token.
const dotnetUnknownCommandToolOutput = `'instal' was not matched. Did you mean one of the following?
install

Unrecognized command or argument 'instal'.

Description:
  Install or work with tools that extend the .NET experience.

Usage:
  dotnet tool [command] [options]

Options:
  -?, -h, --help  Show command line help.

Commands:
  install <packageId>    Install global or local tool.
  uninstall <PACKAGE_ID> Uninstall a global or local tool.
`

// dotnetUnknownCommandEnglishHelp is an excerpt of "dotnet --help".
const dotnetUnknownCommandEnglishHelp = `Usage: dotnet [runtime-options] [path-to-application] [arguments]

Execute a .NET application.

runtime-options:
  --additionalprobingpath <path>   Path containing probing policy and assemblies to probe for.
  --additional-deps <path>         Path to additional deps.json file.
  --depsfile                       Path to <application>.deps.json file.

path-to-application:
  The path to an application .dll file to execute.

Usage: dotnet [sdk-options] [command] [command-options] [arguments]

Execute a .NET SDK command.

sdk-options:
  -d|--diagnostics  Enable diagnostic output.
  -h|--help         Show command line help.
  --info            Display .NET information.
  --list-runtimes   Display the installed runtimes.
  --list-sdks       Display the installed SDKs.
  --version         Display .NET SDK version in use.

SDK commands:
  build             Build a .NET project.
  build-server      Interact with servers started by a build.
  clean             Clean build outputs of a .NET project.
  format            Apply style preferences to a project or solution.
  help              Opens the reference page in a browser for the specified command.
  msbuild           Run Microsoft Build Engine (MSBuild) commands.
  new               Create a new .NET project or file.
  nuget             Provides additional NuGet commands.
  pack              Create a NuGet package.
  package           Search for, add, remove, or list PackageReferences for a .NET project.
  publish           Publish a .NET project for deployment.
  reference         Add, remove, or list ProjectReferences for a .NET project.
  restore           Restore dependencies specified in a .NET project.
  run               Build and run a .NET project output.
  sdk               Manage .NET SDK installation.
  solution          Modify Visual Studio solution files.
  store             Store the specified assemblies in the runtime package store.
  test              Run unit tests using the test runner specified in a .NET project.
  tool              Install or manage tools that extend the .NET experience.
  vstest            Run Microsoft Test Engine (VSTest) commands.
  workload          Manage optional workloads.

Additional commands from bundled tools:
  dev-certs         Create and manage development certificates.
  fsi               Start F# Interactive / execute F# scripts.
  user-jwts         Manage JSON Web Tokens in development.
  user-secrets      Manage development user secrets.
  watch             Start a file watcher that runs a command when files change.

Run 'dotnet [command] --help' for more information on a command.
`

// dotnetUnknownCommandUseHelp replaces the dotnet help lookup with a fixture.
func dotnetUnknownCommandUseHelp(t *testing.T, help string) {
	t.Helper()
	original := dotnetUnknownCommandCommands
	dotnetUnknownCommandCommands = func() []string {
		return dotnetUnknownCommandParseCommands(help)
	}
	t.Cleanup(func() { dotnetUnknownCommandCommands = original })
}

func TestRuleDotnetUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("dotnet_unknown_command")
	if rule == nil {
		t.Fatal("dotnet_unknown_command rule is not registered")
	}

	commands := dotnetUnknownCommandParseCommands(dotnetUnknownCommandEnglishHelp)
	for _, want := range []string{"build", "build-server", "dev-certs", "watch"} {
		if !slices.Contains(commands, want) {
			t.Errorf("dotnet help parser = %q, want it to contain %q", commands, want)
		}
	}
	for _, unwanted := range []string{"--additionalprobingpath", "path-to-application"} {
		if slices.Contains(commands, unwanted) {
			t.Errorf("dotnet help parser = %q, should not contain %q", commands, unwanted)
		}
	}

	dotnetUnknownCommandUseHelp(t, dotnetUnknownCommandEnglishHelp)

	assertMatch(t, rule, "dotnet buidl", dotnetUnknownCommandOutput("buidl"))
	assertMatch(t, rule, "dotnet bulid", dotnetUnknownCommandOutput("bulid"))
	assertMatch(t, rule, "dotnet buidl", dotnetUnknownCommandLocalizedOutput("buidl"))
	assertNotMatch(t, rule, "dotnet tool instal", dotnetUnknownCommandToolOutput)
	assertNotMatch(t, rule, "dotnet run", "Hello, world!")
	assertNotMatch(t, rule, "dotnet --version", "10.0.403")
	assertNotMatch(t, rule, "go buidl", dotnetUnknownCommandOutput("buidl"))

	assertNewCommands(t, rule, "dotnet buidl", dotnetUnknownCommandOutput("buidl"), []string{"dotnet build"})
	assertNewCommands(t, rule, "dotnet buidl", dotnetUnknownCommandLocalizedOutput("buidl"), []string{"dotnet build"})
}

func TestDotnetUnknownCommandFallback(t *testing.T) {
	resetTestState(t)

	original := dotnetUnknownCommandCommands
	dotnetUnknownCommandCommands = func() []string {
		return append([]string{}, dotnetUnknownCommandFallbackCommands...)
	}
	t.Cleanup(func() { dotnetUnknownCommandCommands = original })

	rule := Get("dotnet_unknown_command")
	assertNewCommands(t, rule, "dotnet buidl", dotnetUnknownCommandOutput("buidl"), []string{"dotnet build"})
}
