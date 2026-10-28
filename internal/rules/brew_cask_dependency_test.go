package rules

import "testing"

const brewCaskDependencyOutput = `sshfs: OsxfuseRequirement unsatisfied!

You can install with Homebrew-Cask:
  brew cask install osxfuse

You can download from:
  https://osxfuse.github.io/
Error: An unsatisfied requirement failed this build.`

func TestRuleBrewCaskDependency(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_cask_dependency")
	if rule == nil {
		t.Fatal("brew_cask_dependency rule is not registered")
	}

	assertMatch(t, rule, "brew install sshfs", brewCaskDependencyOutput)

	assertNotMatch(t, rule, "brew link sshfs", brewCaskDependencyOutput)
	assertNotMatch(t, rule, "cat output", brewCaskDependencyOutput)
	assertNotMatch(t, rule, "brew install sshfs", "")

	assertNewCommands(t, rule, "brew install sshfs", brewCaskDependencyOutput,
		[]string{"brew cask install osxfuse && brew install sshfs"})
}
