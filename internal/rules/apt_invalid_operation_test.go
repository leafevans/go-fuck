package rules

import (
	"strings"
	"testing"
)

const aptInvalidOperationAptHelp = `apt 1.0.10.2ubuntu1 for amd64 compiled on Oct  5 2015 15:55:05
Usage: apt [options] command

CLI for apt.
Basic commands:
 list - list packages based on package names
 search - search in package descriptions
 show - show package details

 update - update list of available packages

 install - install packages
 remove  - remove packages

 upgrade - upgrade the system by installing/upgrading packages
 full-upgrade - upgrade the system by removing/installing/upgrading packages

 edit-sources - edit the source information file
`

var aptInvalidOperationAptOperations = []string{
	"list", "search", "show", "update", "install", "remove",
	"upgrade", "full-upgrade", "edit-sources",
}

const aptInvalidOperationAptGetHelp = `apt 1.0.10.2ubuntu1 for amd64 compiled on Oct  5 2015 15:55:05
Usage: apt-get [options] command
       apt-get [options] install|remove pkg1 [pkg2 ...]
       apt-get [options] source pkg1 [pkg2 ...]

apt-get is a simple command line interface for downloading and
installing packages. The most frequently used commands are update
and install.

Commands:
   update - Retrieve new lists of packages
   upgrade - Perform an upgrade
   install - Install new packages (pkg is libc6 not libc6.deb)
   remove - Remove packages
   autoremove - Remove automatically all unused packages
   purge - Remove packages and config files
   source - Download source archives
   build-dep - Configure build-dependencies for source packages
   dist-upgrade - Distribution upgrade, see apt-get(8)
   dselect-upgrade - Follow dselect selections
   clean - Erase downloaded archive files
   autoclean - Erase old downloaded archive files
   check - Verify that there are no broken dependencies
   changelog - Download and display the changelog for the given package
   download - Download the binary package into the current directory

Options:
  -h  This help text.
  -q  Loggable output - no progress indicator
  -qq No output except for errors
  -d  Download only - do NOT install or unpack archives
  -s  No-act. Perform ordering simulation
  -y  Assume Yes to all queries and do not prompt
  -f  Attempt to correct a system with broken dependencies in place
  -m  Attempt to continue if archives are unlocatable
  -u  Show a list of upgraded packages as well
  -b  Build the source package after fetching it
  -V  Show verbose version numbers
  -c=? Read this configuration file
  -o=? Set an arbitrary configuration option, eg -o dir::cache=/tmp
See the apt-get(8), sources.list(5) and apt.conf(5) manual
pages for more information and options.
                       This APT has Super Cow Powers.
`

var aptInvalidOperationAptGetOperations = []string{
	"update", "upgrade", "install", "remove", "autoremove",
	"purge", "source", "build-dep", "dist-upgrade",
	"dselect-upgrade", "clean", "autoclean", "check",
	"changelog", "download",
}

const aptInvalidOperationNewAptGetHelp = `apt 1.6.12 (amd64)
Usage: apt-get [options] command
       apt-get [options] install|remove pkg1 [pkg2 ...]
       apt-get [options] source pkg1 [pkg2 ...]

apt-get is a command line interface for retrieval of packages
and information about them from authenticated sources and
for installation, upgrade and removal of packages together
with their dependencies.

Most used commands:
  update - Retrieve new lists of packages
  upgrade - Perform an upgrade
  install - Install new packages (pkg is libc6 not libc6.deb)
  remove - Remove packages
  purge - Remove packages and config files
  autoremove - Remove automatically all unused packages
  dist-upgrade - Distribution upgrade, see apt-get(8)
  dselect-upgrade - Follow dselect selections
  build-dep - Configure build-dependencies for source packages
  clean - Erase downloaded archive files
  autoclean - Erase old downloaded archive files
  check - Verify that there are no broken dependencies
  source - Download source archives
  download - Download the binary package into the current directory
  changelog - Download and display the changelog for the given package

See apt-get(8) for more information about the available commands.
Configuration options and syntax is detailed in apt.conf(5).
Information about how to configure sources can be found in sources.list(5).
Package and version choices can be expressed via apt_preferences(5).
Security details are available in apt-secure(8).
                                        This APT has Super Cow Powers.
`

var aptInvalidOperationNewAptGetOperations = []string{
	"update", "upgrade", "install", "remove", "purge",
	"autoremove", "dist-upgrade", "dselect-upgrade",
	"build-dep", "clean", "autoclean", "check",
	"source", "download", "changelog",
}

func TestRuleAptInvalidOperation(t *testing.T) {
	resetTestState(t)
	rule := Get("apt_invalid_operation")
	if rule == nil {
		t.Fatal("apt_invalid_operation rule is not registered")
	}

	for _, script := range []string{"apt", "apt-get", "apt-cache"} {
		assertMatch(t, rule, script, "E: Invalid operation saerch")
	}
	assertMatch(t, rule, "sudo apt-get", "E: Invalid operation isntall")

	assertNotMatch(t, rule, "vim", "E: Invalid operation vim")
	assertNotMatch(t, rule, "apt-get", "")

	originalHelp := aptInvalidOperationHelp
	t.Cleanup(func() { aptInvalidOperationHelp = originalHelp })

	operationsCases := []struct {
		app        string
		help       string
		operations []string
	}{
		{"apt", aptInvalidOperationAptHelp, aptInvalidOperationAptOperations},
		{"apt-get", aptInvalidOperationAptGetHelp, aptInvalidOperationAptGetOperations},
		{"apt-get", aptInvalidOperationNewAptGetHelp, aptInvalidOperationNewAptGetOperations},
	}
	for _, testCase := range operationsCases {
		t.Run("operations_"+testCase.app, func(t *testing.T) {
			aptInvalidOperationHelp = func(string) string { return testCase.help }
			got := strings.Join(aptInvalidOperationOperations(testCase.app), ",")
			want := strings.Join(testCase.operations, ",")
			if got != want {
				t.Errorf("_get_operations(%q) = %q, want %q", testCase.app, got, want)
			}
		})
	}

	newCommandCases := []struct {
		script string
		output string
		help   string
		want   string
	}{
		{"apt-get isntall vim", "E: Invalid operation isntall",
			aptInvalidOperationAptGetHelp, "apt-get install vim"},
		{"apt saerch vim", "E: Invalid operation saerch",
			aptInvalidOperationAptHelp, "apt search vim"},
		{"apt uninstall vim", "E: Invalid operation uninstall",
			aptInvalidOperationAptHelp, "apt remove vim"},
	}
	for _, testCase := range newCommandCases {
		t.Run("new_"+testCase.script, func(t *testing.T) {
			aptInvalidOperationHelp = func(string) string { return testCase.help }
			got, err := rule.GetNewCommand(testCommand(testCase.script, testCase.output))
			if err != nil {
				t.Fatalf("rule %q returned an error: %v", rule.Name, err)
			}
			if len(got) == 0 {
				t.Fatalf("rule %q returned no commands", rule.Name)
			}
			if got[0] != testCase.want {
				t.Errorf("rule %q returned %q, want first command %q", rule.Name, got, testCase.want)
			}
		})
	}
}
