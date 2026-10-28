package rules

import "testing"

const chocoInstallPackageNotFoundError = "Chocolatey v0.10.15\n" +
	"Installing the following packages:\n" +
	"logstitcher\n" +
	"By installing you accept licenses for the packages.\n" +
	"logstitcher not installed. The package was not found with the source(s) listed.\n" +
	" Source(s): 'https://chocolatey.org/api/v2/'\n" +
	" NOTE: When you specify explicit sources, it overrides default sources.\n" +
	"If the package version is a prerelease and you didn't specify `--pre`,\n" +
	" the package may not be found.\n" +
	"Please see https://chocolatey.org/docs/troubleshooting for more\n" +
	" assistance.\n" +
	"\n" +
	"Chocolatey installed 0/1 packages. 1 packages failed.\n" +
	" See the log for details (C:\\ProgramData\\chocolatey\\logs\\chocolatey.log).\n" +
	"\n" +
	"Failures\n" +
	" - logstitcher - logstitcher not installed. The package was not found with the source(s) listed.\n" +
	" Source(s): 'https://chocolatey.org/api/v2/'\n" +
	" NOTE: When you specify explicit sources, it overrides default sources.\n" +
	"If the package version is a prerelease and you didn't specify `--pre`,\n" +
	" the package may not be found.\n" +
	"Please see https://chocolatey.org/docs/troubleshooting for more\n" +
	" assistance.\n"

func TestRuleChocoInstall(t *testing.T) {
	resetTestState(t)
	rule := Get("choco_install")
	if rule == nil {
		t.Fatal("choco_install rule is not registered")
	}

	matchScripts := []string{
		"choco install logstitcher",
		"cinst logstitcher",
		"choco install logstitcher -y",
		"cinst logstitcher -y",
		"choco install logstitcher -y -n=test",
		"cinst logstitcher -y -n=test",
		"choco install logstitcher -y -n=test /env",
		"cinst logstitcher -y -n=test /env",
		"choco install chocolatey -y",
		"cinst chocolatey -y",
	}
	for _, script := range matchScripts {
		assertMatch(t, rule, script, chocoInstallPackageNotFoundError)
	}

	notMatchScripts := []string{
		"choco /?",
		"choco upgrade logstitcher",
		"cup logstitcher",
		"choco upgrade logstitcher -y",
		"cup logstitcher -y",
		"choco upgrade logstitcher -y -n=test",
		"cup logstitcher -y -n=test",
		"choco upgrade logstitcher -y -n=test /env",
		"cup logstitcher -y -n=test /env",
		"choco upgrade chocolatey -y",
		"cup chocolatey -y",
		"choco uninstall logstitcher",
		"cuninst logstitcher",
		"choco uninstall logstitcher -y",
		"cuninst logstitcher -y",
		"choco uninstall logstitcher -y -n=test",
		"cuninst logstitcher -y -n=test",
		"choco uninstall logstitcher -y -n=test /env",
		"cuninst logstitcher -y -n=test /env",
		"choco uninstall chocolatey -y",
		"cuninst chocolatey -y",
	}
	for _, script := range notMatchScripts {
		assertNotMatch(t, rule, script, "")
	}

	newCommandCases := []struct {
		before string
		after  string
	}{
		{"choco install logstitcher", "choco install logstitcher.install"},
		{"cinst logstitcher", "cinst logstitcher.install"},
		{"choco install logstitcher -y", "choco install logstitcher.install -y"},
		{"cinst logstitcher -y", "cinst logstitcher.install -y"},
		{"choco install logstitcher -y -n=test", "choco install logstitcher.install -y -n=test"},
		{"cinst logstitcher -y -n=test", "cinst logstitcher.install -y -n=test"},
		{"choco install logstitcher -y -n=test /env", "choco install logstitcher.install -y -n=test /env"},
		{"cinst logstitcher -y -n=test /env", "cinst logstitcher.install -y -n=test /env"},
		{"choco install chocolatey -y", "choco install chocolatey.install -y"},
		{"cinst chocolatey -y", "cinst chocolatey.install -y"},
	}
	for _, testCase := range newCommandCases {
		assertNewCommands(t, rule, testCase.before, "", []string{testCase.after})
	}
}
