package rules

import "testing"

// yarnCommandNotFoundOutput returns the output yarn prints for an unknown
// command.
func yarnCommandNotFoundOutput(name string) string {
	return "error Command \"" + name + "\" not found.\n"
}

// yarnCommandNotFoundHelpOutput is the "yarn --help" output of the original
// test fixture.
const yarnCommandNotFoundHelpOutput = `

  Usage: yarn [command] [flags]

  Options:

    -h, --help                      output usage information
    -V, --version                   output the version number
    --verbose                       output verbose messages on internal operations
    --offline                       trigger an error if any required dependencies are not available in local cache
    --prefer-offline                use network only if dependencies are not available in local cache
    --strict-semver                 
    --json                          
    --ignore-scripts                don't run lifecycle scripts
    --har                           save HAR output of network traffic
    --ignore-platform               ignore platform checks
    --ignore-engines                ignore engines check
    --ignore-optional               ignore optional dependencies
    --force                         ignore all caches
    --no-bin-links                  don't generate bin links when setting up packages
    --flat                          only allow one version of a package
    --prod, --production [prod]     
    --no-lockfile                   don't read or generate a lockfile
    --pure-lockfile                 don't generate a lockfile
    --frozen-lockfile               don't generate a lockfile and fail if an update is needed
    --link-duplicates               create hardlinks to the repeated modules in node_modules
    --global-folder <path>          
    --modules-folder <path>         rather than installing modules into the node_modules folder relative to the cwd, output them here
    --cache-folder <path>           specify a custom folder to store the yarn cache
    --mutex <type>[:specifier]      use a mutex to ensure only one yarn instance is executing
    --no-emoji                      disable emoji in output
    --proxy <host>                  
    --https-proxy <host>            
    --no-progress                   disable progress bar
    --network-concurrency <number>  maximum number of concurrent network requests

  Commands:

    - access
    - add
    - bin
    - cache
    - check
    - clean
    - config
    - generate-lock-entry
    - global
    - import
    - info
    - init
    - install
    - licenses
    - link
    - list
    - login
    - logout
    - outdated
    - owner
    - pack
    - publish
    - remove
    - run
    - tag
    - team
    - unlink
    - upgrade
    - upgrade-interactive
    - version
    - versions
    - why

  Run ` + "`yarn help COMMAND`" + ` for more information on specific commands.
  Visit https://yarnpkg.com/en/docs/cli/ to learn more about Yarn.
`

// yarnCommandNotFoundStubTasks replaces the yarn commands with the fixture.
func yarnCommandNotFoundStubTasks(t *testing.T) {
	t.Helper()
	original := yarnAllTasks
	yarnAllTasks = func() []string {
		return yarnCommandNotFoundParseTasks(yarnCommandNotFoundHelpOutput)
	}
	t.Cleanup(func() { yarnAllTasks = original })
}

// yarnCommandNotFoundAssertFirstCommand checks the best suggestion. The
// original test only compares the first corrected command as well.
func yarnCommandNotFoundAssertFirstCommand(t *testing.T, script, output, want string) {
	t.Helper()
	rule := Get("yarn_command_not_found")
	got, err := rule.GetNewCommand(testCommand(script, output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 {
		t.Fatalf("rule %q returned no corrected commands for %q", rule.Name, script)
	}
	if got[0] != want {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, want)
	}
}

func TestRuleYarnCommandNotFound(t *testing.T) {
	resetTestState(t)
	yarnCommandNotFoundStubTasks(t)

	rule := Get("yarn_command_not_found")
	if rule == nil {
		t.Fatal("yarn_command_not_found rule is not registered")
	}

	assertMatch(t, rule, "yarn whyy webpack", yarnCommandNotFoundOutput("whyy"))
	assertNotMatch(t, rule, "npm nuild", yarnCommandNotFoundOutput("nuild"))
	assertNotMatch(t, rule, "yarn install", "")

	yarnCommandNotFoundAssertFirstCommand(t, "yarn whyy webpack",
		yarnCommandNotFoundOutput("whyy"), "yarn why webpack")
	assertNewCommands(t, rule, "yarn require lodash", yarnCommandNotFoundOutput("require"),
		[]string{"yarn add lodash"})
}
