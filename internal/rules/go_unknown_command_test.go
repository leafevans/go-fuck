package rules

import "testing"

// goUnknownCommandOutput is the output of a misspelled go command.
const goUnknownCommandOutput = `go bulid: unknown command
Run 'go help' for usage.`

// goUnknownCommandStderr is the help output of "go" of the Python tests. The
// "Additional help topics" section is omitted because it is not parsed.
const goUnknownCommandStderr = `Go is a tool for managing Go source code.

Usage:

	go <command> [arguments]

The commands are:

	bug         start a bug report
	build       compile packages and dependencies
	clean       remove object files and cached files
	doc         show documentation for package or symbol
	env         print Go environment information
	fix         update packages to use new APIs
	fmt         gofmt (reformat) package sources
	generate    generate Go files by processing source
	get         add dependencies to current module and install them
	install     compile and install packages and dependencies
	list        list packages or modules
	mod         module maintenance
	run         compile and run Go program
	test        test packages
	tool        run specified go tool
	version     print Go version
	vet         report likely mistakes in packages

Use "go help <command>" for more information about a command.
`

// goUnknownCommandUseHelp replaces the go help lookup with the Python fixture.
func goUnknownCommandUseHelp(t *testing.T) {
	t.Helper()
	original := goUnknownCommandCommands
	goUnknownCommandCommands = func() []string {
		return goUnknownCommandParseCommands(goUnknownCommandStderr)
	}
	t.Cleanup(func() { goUnknownCommandCommands = original })
}

func TestRuleGoUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("go_unknown_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "go bulid", goUnknownCommandOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "go run", "go run: no go files listed")
	})

	t.Run("get new command", func(t *testing.T) {
		goUnknownCommandUseHelp(t)
		assertNewCommands(t, rule, "go bulid", goUnknownCommandOutput, []string{"go build"})
	})
}
