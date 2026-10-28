package rules

import "testing"

// reactNativeCommandOutput builds the unrecognized command error message.
func reactNativeCommandOutput(name string) string {
	return "Unrecognized command '" + name + "'"
}

// reactNativeCommandStdout is the "react-native --help" output of the Python
// tests.
const reactNativeCommandStdout = `
Scanning 615 folders for symlinks in /home/nvbn/work/zcho/BookkaWebView/node_modules (6ms)

  Usage: react-native [options] [command]


  Options:

    -V, --version  output the version number
    -h, --help     output usage information


  Commands:

    start [options]                    starts the webserver
    run-ios [options]                  builds your app and starts it on iOS simulator
    run-android [options]              builds your app and starts it on a connected Android emulator or device
    new-library [options]              generates a native library bridge
    bundle [options]                   builds the javascript bundle for offline use
    unbundle [options]                 builds javascript as "unbundle" for offline use
    eject [options]                    Re-create the iOS and Android folders and native code
    link [options] [packageName]       links all native dependencies (updates native build files)
    unlink [options] <packageName>     unlink native dependency
    install [options] <packageName>    install and link native dependencies
    uninstall [options] <packageName>  uninstall and unlink native dependencies
    upgrade [options]                  upgrade your app's template files to the latest version; run this after updating the react-native version in your package.json and running npm install
    log-android [options]              starts adb logcat
    log-ios [options]                  starts iOS device syslog tail
`

// reactNativeCommandUseHelp replaces the react-native help lookup with the
// Python fixture.
func reactNativeCommandUseHelp(t *testing.T) {
	t.Helper()
	original := reactNativeCommandCommands
	reactNativeCommandCommands = func() []string {
		return reactNativeCommandParseCommands(reactNativeCommandStdout)
	}
	t.Cleanup(func() { reactNativeCommandCommands = original })
}

func TestRuleReactNativeCommandUnrecognized(t *testing.T) {
	resetTestState(t)
	rule := Get("react_native_command_unrecognized")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "react-native star", reactNativeCommandOutput("star"))
		assertMatch(t, rule, "react-native android-logs", reactNativeCommandOutput("android-logs"))
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "gradle star", reactNativeCommandOutput("star"))
		assertNotMatch(t, rule, "react-native start", "")
	})

	t.Run("get new command", func(t *testing.T) {
		reactNativeCommandUseHelp(t)
		newCommands, err := rule.GetNewCommand(testCommand("react-native star", reactNativeCommandOutput("star")))
		if err != nil {
			t.Fatalf("rule %q returned an error: %v", rule.Name, err)
		}
		if len(newCommands) == 0 || newCommands[0] != "react-native start" {
			t.Fatalf("rule %q returned %q, want the first command %q", rule.Name, newCommands, "react-native start")
		}

		newCommands, err = rule.GetNewCommand(testCommand("react-native logsandroid -f", reactNativeCommandOutput("logsandroid")))
		if err != nil {
			t.Fatalf("rule %q returned an error: %v", rule.Name, err)
		}
		if len(newCommands) == 0 || newCommands[0] != "react-native log-android -f" {
			t.Fatalf("rule %q returned %q, want the first command %q", rule.Name, newCommands, "react-native log-android -f")
		}
	})
}
