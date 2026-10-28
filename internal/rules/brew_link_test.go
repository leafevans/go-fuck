package rules

import "testing"

const brewLinkOutput = `Error: Could not symlink bin/gcp
Target /usr/local/bin/gcp
already exists. You may want to remove it:
  rm '/usr/local/bin/gcp'

To force the link and overwrite all conflicting files:
  brew link --overwrite coreutils

To list all files that would be deleted:
  brew link --overwrite --dry-run coreutils
`

func TestRuleBrewLink(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_link")
	if rule == nil {
		t.Fatal("brew_link rule is not registered")
	}

	assertMatch(t, rule, "brew link coreutils", brewLinkOutput)
	assertMatch(t, rule, "brew ln coreutils", brewLinkOutput)
	assertNotMatch(t, rule, "brew link coreutils", "")

	assertNewCommands(t, rule, "brew link coreutils", brewLinkOutput,
		[]string{"brew link --overwrite --dry-run coreutils"})
}
