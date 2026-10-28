package rules

import "testing"

// azCLINoSuggestions is the az output without a suggestion.
const azCLINoSuggestions = `az provider: error: the following arguments are required: _subcommand
usage: az provider [-h] {list,show,register,unregister,operation} ...
`

// azCLIMisspelledCommand is the az output for a misspelled command.
const azCLIMisspelledCommand = `az: 'providers' is not in the 'az' command group. See 'az --help'.

The most similar choice to 'providers' is:
    provider
`

// azCLIMisspelledSubcommand is the az output for a misspelled subcommand.
const azCLIMisspelledSubcommand = `az provider: 'lis' is not in the 'az provider' command group. See 'az provider --help'.

The most similar choice to 'lis' is:
    list
`

func TestRuleAzCLI(t *testing.T) {
	resetTestState(t)

	rule := Get("az_cli")
	if rule == nil {
		t.Fatal("az_cli rule is not registered")
	}

	assertMatch(t, rule, "az providers", azCLIMisspelledCommand)
	assertMatch(t, rule, "az provider lis", azCLIMisspelledSubcommand)

	assertNotMatch(t, rule, "az provider", azCLINoSuggestions)

	assertNewCommands(t, rule, "az providers list", azCLIMisspelledCommand,
		[]string{"az provider list"})
	assertNewCommands(t, rule, "az provider lis", azCLIMisspelledSubcommand,
		[]string{"az provider list"})
}
