package rules

import "testing"

func TestRuleNixosCmdNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("nixos_cmd_not_found")
	if rule == nil {
		t.Fatal("nixos_cmd_not_found rule is not registered")
	}

	assertMatch(t, rule, "vim", "nix-env -iA nixos.vim")
	assertNotMatch(t, rule, "vim", "")
	assertNotMatch(t, rule, "", "")

	assertNewCommands(t, rule, "vim", "nix-env -iA nixos.vim",
		[]string{"nix-env -iA nixos.vim && vim"})
	assertNewCommands(t, rule, "pacman", "nix-env -iA nixos.pacman",
		[]string{"nix-env -iA nixos.pacman && pacman"})
}
