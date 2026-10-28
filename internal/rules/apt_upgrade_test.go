package rules

import "testing"

const aptUpgradeMatchOutput = `
Listing... Done
heroku/stable 6.15.2-1 amd64 [upgradable from: 6.14.43-1]
resolvconf/zesty-updates,zesty-updates 1.79ubuntu4.1 all [upgradable from: 1.79ubuntu4]
squashfs-tools/zesty-updates 1:4.3-3ubuntu2.17.04.1 amd64 [upgradable from: 1:4.3-3ubuntu2]
unattended-upgrades/zesty-updates,zesty-updates 0.93.1ubuntu2.4 all [upgradable from: 0.93.1ubuntu2.3]
`

const aptUpgradeNoMatchOutput = `
Listing... Done
`

func TestRuleAptUpgrade(t *testing.T) {
	resetTestState(t)
	rule := Get("apt_upgrade")
	if rule == nil {
		t.Fatal("apt_upgrade rule is not registered")
	}

	assertMatch(t, rule, "apt list --upgradable", aptUpgradeMatchOutput)
	assertMatch(t, rule, "sudo apt list --upgradable", aptUpgradeMatchOutput)

	assertNotMatch(t, rule, "apt list --upgradable", aptUpgradeNoMatchOutput)
	assertNotMatch(t, rule, "sudo apt list --upgradable", aptUpgradeNoMatchOutput)

	assertNewCommands(t, rule, "apt list --upgradable", aptUpgradeMatchOutput,
		[]string{"apt upgrade"})
	assertNewCommands(t, rule, "sudo apt list --upgradable", aptUpgradeMatchOutput,
		[]string{"sudo apt upgrade"})
}
