package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/system"
)

// yarnHelpOutput is the yarn output that points to the command documentation.
const yarnHelpOutput = `

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

  Visit https://yarnpkg.com/en/docs/cli/clean for documentation about this command.
`

func TestRuleYarnHelp(t *testing.T) {
	resetTestState(t)

	rule := Get("yarn_help")
	if rule == nil {
		t.Fatal("yarn_help rule is not registered")
	}

	assertMatch(t, rule, "yarn help clean", yarnHelpOutput)

	assertNotMatch(t, rule, "yarn clean", yarnHelpOutput)
	assertNotMatch(t, rule, "npm help clean", yarnHelpOutput)
	assertNotMatch(t, rule, "yarn help clean", "")

	url := "https://yarnpkg.com/en/docs/cli/clean"
	assertNewCommands(t, rule, "yarn help clean", yarnHelpOutput,
		[]string{system.OpenCommand(url)})
}
