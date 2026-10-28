package rules

import "testing"

// adbUnknownCommandOutput returns the help output that adb prints for an
// unknown command.
func adbUnknownCommandOutput() string {
	return `Android Debug Bridge version 1.0.31

 -d                            - directs command to the only connected USB device
                                 returns an error if more than one USB device is present.
 -e                            - directs command to the only running emulator.
                                 returns an error if more than one emulator is running.
 -s <specific device>          - directs command to the device or emulator with the given
                                 serial number or qualifier. Overrides ANDROID_SERIAL
                                 environment variable.
`
}

func TestRuleAdbUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("adb_unknown_command")
	if rule == nil {
		t.Fatal("adb_unknown_command rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		output := adbUnknownCommandOutput()
		assertMatch(t, rule, "adb lgcat", output)
		assertMatch(t, rule, "adb puhs", output)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "git branch foo", "")
		assertNotMatch(t, rule, "abd push", "")
	})

	t.Run("new command", func(t *testing.T) {
		output := adbUnknownCommandOutput()
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"adb puhs test.bin /sdcard/test.bin", []string{"adb push test.bin /sdcard/test.bin"}},
			{"adb -s 1111 logcta", []string{"adb -s 1111 logcat"}},
			{"adb -P 666 pulll /sdcard/test.bin", []string{"adb -P 666 pull /sdcard/test.bin"}},
			{"adb -d logcatt", []string{"adb -d logcat"}},
			{"adb -e reboott", []string{"adb -e reboot"}},
		} {
			assertNewCommands(t, rule, testCase.script, output, testCase.want)
		}
	})
}
