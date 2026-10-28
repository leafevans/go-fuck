package rules

import "testing"

var pacmanNotFoundPackages = []string{"extra/llvm", "extra/llvm35"}

func TestRulePacmanNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("pacman_not_found")
	if rule == nil {
		t.Fatal("pacman_not_found rule is not registered")
	}

	originalPkgFile := pacmanNotFoundPkgFile
	t.Cleanup(func() { pacmanNotFoundPkgFile = originalPkgFile })
	pacmanNotFoundPkgFile = func(string) ([]string, error) {
		return pacmanNotFoundPackages, nil
	}

	scripts := []string{
		"yay -S llc",
		"pikaur -S llc",
		"yaourt -S llc",
		"pacman llc",
		"sudo pacman llc",
	}
	for _, script := range scripts {
		assertMatch(t, rule, script, "error: target not found: llc")
		assertNotMatch(t, rule, script, "")
	}
	assertNotMatch(t, rule, "vim llc", "error: target not found: llc")

	newCommandCases := []struct {
		script string
		want   []string
	}{
		{"yay -S llc", []string{"yay -S extra/llvm", "yay -S extra/llvm35"}},
		{"pikaur -S llc", []string{"pikaur -S extra/llvm", "pikaur -S extra/llvm35"}},
		{"yaourt -S llc", []string{"yaourt -S extra/llvm", "yaourt -S extra/llvm35"}},
		{"pacman -S llc", []string{"pacman -S extra/llvm", "pacman -S extra/llvm35"}},
		{"sudo pacman -S llc", []string{"sudo pacman -S extra/llvm", "sudo pacman -S extra/llvm35"}},
	}
	for _, testCase := range newCommandCases {
		assertNewCommands(t, rule, testCase.script, "error: target not found: llc", testCase.want)
	}
}
