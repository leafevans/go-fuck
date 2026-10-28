package rules

import "testing"

// ifconfigDeviceNotFoundStdout is the output of "ifconfig -a" of the Python
// tests.
const ifconfigDeviceNotFoundStdout = `
wlp2s0    Link encap:Ethernet  HWaddr 5c:51:4f:7c:58:5d
          inet addr:192.168.0.103  Bcast:192.168.0.255  Mask:255.255.255.0
          inet6 addr: fe80::be23:69b9:96d2:6d39/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:23581604 errors:0 dropped:0 overruns:0 frame:0
          TX packets:17017655 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:16148429061 (16.1 GB)  TX bytes:7067533695 (7.0 GB)
`

// ifconfigDeviceNotFoundUseInterfaces replaces the ifconfig lookup with the
// interfaces parsed from the Python fixture.
func ifconfigDeviceNotFoundUseInterfaces(t *testing.T) {
	t.Helper()
	original := ifconfigPossibleInterfaces
	ifconfigPossibleInterfaces = func() []string {
		return ifconfigParseInterfaces(ifconfigDeviceNotFoundStdout)
	}
	t.Cleanup(func() { ifconfigPossibleInterfaces = original })
}

func TestRuleIfconfigDeviceNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("ifconfig_device_not_found")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "ifconfig wlan0", "wlan0: error fetching interface information: Device not found")
		assertMatch(t, rule, "ifconfig -s eth0", "eth0: error fetching interface information: Device not found")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "config wlan0", "wlan0: error fetching interface information: Device not found")
		assertNotMatch(t, rule, "ifconfig eth0", "")
	})

	t.Run("get new command", func(t *testing.T) {
		ifconfigDeviceNotFoundUseInterfaces(t)
		assertNewCommands(t, rule, "ifconfig wlan0", "wlan0: error fetching interface information: Device not found",
			[]string{"ifconfig wlp2s0"})
		assertNewCommands(t, rule, "ifconfig -s wlan0", "wlan0: error fetching interface information: Device not found",
			[]string{"ifconfig -s wlp2s0"})
	})
}
