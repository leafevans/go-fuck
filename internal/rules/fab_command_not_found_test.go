package rules

import "testing"

// fabCommandNotFoundOutput is the fab output of the Python tests.
const fabCommandNotFoundOutput = `
Warning: Command(s) not found:
    extenson
    deloyp

Available commands:

    update_config
    prepare_extension
    Template               A string class for supporting $-substitutions.
    deploy
    glob                   Return a list of paths matching a pathname pattern.
    install_web
    set_version
`

func TestRuleFabCommandNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("fab_command_not_found")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "fab extenson", fabCommandNotFoundOutput)
		assertMatch(t, rule, "fab deloyp", fabCommandNotFoundOutput)
		assertMatch(t, rule, "fab extenson deloyp", fabCommandNotFoundOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "gulp extenson", fabCommandNotFoundOutput)
		assertNotMatch(t, rule, "fab deloyp", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "fab extenson", fabCommandNotFoundOutput,
			[]string{"fab prepare_extension"})
		assertNewCommands(t, rule, "fab extenson:version=2016", fabCommandNotFoundOutput,
			[]string{"fab prepare_extension:version=2016"})
		assertNewCommands(t, rule, "fab extenson:version=2016 install_web set_version:val=0.5.0", fabCommandNotFoundOutput,
			[]string{"fab prepare_extension:version=2016 install_web set_version:val=0.5.0"})
		assertNewCommands(t, rule, "fab extenson:version=2016 deloyp:beta=true -H the.fuck", fabCommandNotFoundOutput,
			[]string{"fab prepare_extension:version=2016 deploy:beta=true -H the.fuck"})
	})
}
