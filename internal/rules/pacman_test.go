package rules

import "testing"

var pacmanVimPackages = []string{
	"extra/gvim", "extra/gvim-python3", "extra/vim", "extra/vim-minimal", "extra/vim-python3",
}

func TestRulePacman(t *testing.T) {
	resetTestState(t)
	rule := Get("pacman")
	if rule == nil {
		t.Fatal("pacman rule is not registered")
	}

	originalPkgFile := pacmanPkgFile
	originalInstall := pacmanInstallCommand
	t.Cleanup(func() {
		pacmanPkgFile = originalPkgFile
		pacmanInstallCommand = originalInstall
	})
	pacmanInstallCommand = func() string { return "pacman" }
	pacmanPkgFile = func(command string) ([]string, error) {
		switch command {
		case "vim", "sudo vim":
			return pacmanVimPackages, nil
		case "convert", "sudo convert":
			return []string{"extra/imagemagick"}, nil
		case "sudo":
			return []string{"core/sudo"}, nil
		}
		return nil, nil
	}

	assertMatch(t, rule, "vim", "vim: command not found")
	assertMatch(t, rule, "sudo vim", "sudo: vim: command not found")

	assertNotMatch(t, rule, "vim", "")
	assertNotMatch(t, rule, "", "")
	assertNotMatch(t, rule, "sudo vim", "")

	assertNewCommands(t, rule, "vim", "", []string{
		"pacman -S extra/gvim && vim",
		"pacman -S extra/gvim-python3 && vim",
		"pacman -S extra/vim && vim",
		"pacman -S extra/vim-minimal && vim",
		"pacman -S extra/vim-python3 && vim",
	})
	assertNewCommands(t, rule, "sudo vim", "", []string{
		"pacman -S extra/gvim && sudo vim",
		"pacman -S extra/gvim-python3 && sudo vim",
		"pacman -S extra/vim && sudo vim",
		"pacman -S extra/vim-minimal && sudo vim",
		"pacman -S extra/vim-python3 && sudo vim",
	})
	assertNewCommands(t, rule, "convert", "",
		[]string{"pacman -S extra/imagemagick && convert"})
	assertNewCommands(t, rule, "sudo convert", "",
		[]string{"pacman -S extra/imagemagick && sudo convert"})
	assertNewCommands(t, rule, "sudo", "",
		[]string{"pacman -S core/sudo && sudo"})
}
