package rules

import "testing"

const composerNotCommandOutput = "\n" +
	"\n" +
	"                                    \n" +
	"  [InvalidArgumentException]        \n" +
	"  Command \"udpate\" is not defined.  \n" +
	"  Did you mean this?                \n" +
	"      update                        \n" +
	"                                    \n" +
	"\n" +
	"\n"

const composerNotCommandOneOfThisOutput = "\n" +
	"\n" +
	"                                   \n" +
	"  [InvalidArgumentException]       \n" +
	"  Command \"pdate\" is not defined.  \n" +
	"  Did you mean one of these?       \n" +
	"      selfupdate                   \n" +
	"      self-update                  \n" +
	"      update                       \n" +
	"                                   \n" +
	"\n" +
	"\n"

const composerNotCommandRequireInsteadOfInstallOutput = "Invalid argument package. " +
	"Use \"composer require package\" instead to add packages to your composer.json."

func TestRuleComposerNotCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("composer_not_command")
	if rule == nil {
		t.Fatal("composer_not_command rule is not registered")
	}

	assertMatch(t, rule, "composer udpate", composerNotCommandOutput)
	assertMatch(t, rule, "composer pdate", composerNotCommandOneOfThisOutput)
	assertMatch(t, rule, "composer install package", composerNotCommandRequireInsteadOfInstallOutput)
	assertNotMatch(t, rule, "ls update", composerNotCommandOutput)

	assertNewCommands(t, rule, "composer udpate", composerNotCommandOutput,
		[]string{"composer update"})
	assertNewCommands(t, rule, "composer pdate", composerNotCommandOneOfThisOutput,
		[]string{"composer selfupdate"})
	assertNewCommands(t, rule, "composer install package", composerNotCommandRequireInsteadOfInstallOutput,
		[]string{"composer require package"})
}
