package rules

import "testing"

// vagrantUpSSHOutput is the output of a failing "vagrant ssh".
const vagrantUpSSHOutput = "VM must be running to open SSH connection. Run `vagrant up`\nto start the virtual machine."

// vagrantUpRDPOutput is the output of a failing "vagrant rdp".
const vagrantUpRDPOutput = "VM must be created before running this command. Run `vagrant up` first."

// vagrantUpNotMatchOutput is the output of "vagrant ssh" with an unknown
// machine.
const vagrantUpNotMatchOutput = "The machine with the name 'jeff' was not found configured for this Vagrant environment."

// vagrantUpInitOutput tells the user to initialize the environment instead.
const vagrantUpInitOutput = "A Vagrant environment or target machine is required to run this command. Run `vagrant init` to create a new Vagrant environment. Or, get an ID of a target machine from `vagrant global-status` to run this command on. A final option is to change to a directory with a Vagrantfile and to try again."

func TestRuleVagrantUp(t *testing.T) {
	resetTestState(t)
	rule := Get("vagrant_up")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "vagrant ssh", vagrantUpSSHOutput)
		assertMatch(t, rule, "vagrant ssh devbox", vagrantUpSSHOutput)
		assertMatch(t, rule, "vagrant rdp", vagrantUpRDPOutput)
		assertMatch(t, rule, "vagrant rdp devbox", vagrantUpRDPOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "vagrant ssh", "")
		assertNotMatch(t, rule, "vagrant ssh jeff", vagrantUpNotMatchOutput)
		assertNotMatch(t, rule, "vagrant ssh", vagrantUpInitOutput)
		assertNotMatch(t, rule, "", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "vagrant ssh", vagrantUpSSHOutput,
			[]string{"vagrant up && vagrant ssh"})
		assertNewCommands(t, rule, "vagrant ssh devbox", vagrantUpSSHOutput,
			[]string{"vagrant up devbox && vagrant ssh devbox", "vagrant up && vagrant ssh devbox"})
		assertNewCommands(t, rule, "vagrant rdp", vagrantUpRDPOutput,
			[]string{"vagrant up && vagrant rdp"})
		assertNewCommands(t, rule, "vagrant rdp devbox", vagrantUpRDPOutput,
			[]string{"vagrant up devbox && vagrant rdp devbox", "vagrant up && vagrant rdp devbox"})
	})
}
