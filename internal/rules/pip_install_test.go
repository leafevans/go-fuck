package rules

import "testing"

func TestRulePipInstall(t *testing.T) {
	resetTestState(t)

	rule := Get("pip_install")
	if rule == nil {
		t.Fatal("pip_install rule is not registered")
	}

	response1 := "\n    Could not install packages due to an EnvironmentError: " +
		"[Errno 13] Permission denied: " +
		"'/Library/Python/2.7/site-packages/entrypoints.pyc'\n" +
		"Consider using the `--user` option or check the permissions.\n"
	assertMatch(t, rule, "pip install -r requirements.txt", response1)

	response2 := `
Collecting bacon
  Downloading https://files.pythonhosted.org/packages/b2/81/19fb79139ee71c8bc4e5a444546f318e2b87253b8939ec8a7e10d63b7341/bacon-0.3.1.zip (11.0MB)
    100% |████████████████████████████████| 11.0MB 3.0MB/s
Installing collected packages: bacon
  Running setup.py install for bacon ... done
Successfully installed bacon-0.3.1
`
	assertNotMatch(t, rule, "pip install bacon", response2)

	assertNewCommands(t, rule, "pip install -r requirements.txt", "",
		[]string{"pip install --user -r requirements.txt"})
	assertNewCommands(t, rule, "pip install bacon", "",
		[]string{"pip install --user bacon"})
	assertNewCommands(t, rule, "pip install --user -r requirements.txt", "",
		[]string{"sudo pip install -r requirements.txt"})

	// The rule supports sudo like the original implementation.
	assertNewCommands(t, rule, "sudo pip install bacon", "",
		[]string{"sudo pip install --user bacon"})
}
