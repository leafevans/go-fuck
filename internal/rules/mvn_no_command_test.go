package rules

import "testing"

// mvnNoCommandOutput is the maven output printed when no goal is specified.
const mvnNoCommandOutput = "[ERROR] No goals have been specified for this build. You must specify a valid lifecycle phase or a goal in the format <plugin-prefix>:<goal> or <plugin-group-id>:<plugin-artifact-id>[:<plugin-version>]:<goal>. Available lifecycle phases are: validate, initialize, generate-sources, process-sources, generate-resources, process-resources, compile, process-classes, generate-test-sources, process-test-sources, generate-test-resources, process-test-resources, test-compile, process-test-classes, test, prepare-package, package, pre-integration-test, integration-test, post-integration-test, verify, install, deploy, pre-clean, clean, post-clean, pre-site, site, post-site, site-deploy. -> [Help 1]"

func TestRuleMvnNoCommand(t *testing.T) {
	resetTestState(t)

	rule := Get("mvn_no_command")
	if rule == nil {
		t.Fatal("mvn_no_command rule is not registered")
	}

	assertMatch(t, rule, "mvn", mvnNoCommandOutput)

	mvnCleanOutput := `
[INFO] Scanning for projects...[INFO]
[INFO] ------------------------------------------------------------------------
[INFO] Building test 0.2
[INFO] ------------------------------------------------------------------------
[INFO] 
[INFO] --- maven-clean-plugin:2.5:clean (default-clean) @ test ---
[INFO] Deleting /home/mlk/code/test/target
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time: 0.477s
[INFO] Finished at: Wed Aug 26 13:05:47 BST 2015
[INFO] Final Memory: 6M/240M
[INFO] ------------------------------------------------------------------------
`
	assertNotMatch(t, rule, "mvn clean", mvnCleanOutput)
	assertNotMatch(t, rule, "mvn --help", "")
	assertNotMatch(t, rule, "mvn -v", "")

	assertNewCommands(t, rule, "mvn", mvnNoCommandOutput,
		[]string{"mvn clean package", "mvn clean install"})
	assertNewCommands(t, rule, "mvn -N", mvnNoCommandOutput,
		[]string{"mvn -N clean package", "mvn -N clean install"})
}
