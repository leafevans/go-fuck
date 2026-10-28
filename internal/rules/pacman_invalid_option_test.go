package rules

import (
	"strings"
	"testing"
)

const pacmanInvalidOptionGoodOutput = `community/shared_meataxe 1.0-3
    A set of programs for working with matrix representations over finite fields
`

const pacmanInvalidOptionBadOutput = "error: invalid option '-"

func TestRulePacmanInvalidOption(t *testing.T) {
	resetTestState(t)
	rule := Get("pacman_invalid_option")
	if rule == nil {
		t.Fatal("pacman_invalid_option rule is not registered")
	}

	for _, option := range "SURQFDVT" {
		assertNotMatch(t, rule, "pacman -"+string(option)+"s meat", pacmanInvalidOptionGoodOutput)
	}
	for _, option := range "azxcbnm" {
		assertNotMatch(t, rule, "pacman -"+string(option)+"v meat", pacmanInvalidOptionBadOutput)
	}
	for _, option := range "surqfdvt" {
		script := "pacman -" + string(option) + "v meat"
		assertMatch(t, rule, script, pacmanInvalidOptionBadOutput)
		assertNewCommands(t, rule, script, "",
			[]string{"pacman -" + strings.ToUpper(string(option)) + "v meat"})
	}
}
