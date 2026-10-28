package rules

import "testing"

// mvnUnknownLifecyclePhaseOutput returns the maven output for an unknown
// lifecycle phase. The available phases are the ones of the original fixture.
func mvnUnknownLifecyclePhaseOutput(phase string) string {
	return "[ERROR] Unknown lifecycle phase \"" + phase + "\". You must specify a valid lifecycle phase or a goal in the format <plugin-prefix>:<goal> or <plugin-group-id>:<plugin-artifact-id>[:<plugin-version>]:<goal>. Available lifecycle phases are: validate, initialize, generate-sources, process-sources, generate-resources, process-resources, compile, process-classes, generate-test-sources, process-test-sources, generate-test-resources, process-test-resources, test-compile, process-test-classes, test, prepare-package, package, pre-integration-test, integration-test, post-integration-test, verify, install, deploy, pre-clean, clean, post-clean, pre-site, site, post-site, site-deploy. -> [Help 1]"
}

func TestRuleMvnUnknownLifecyclePhase(t *testing.T) {
	resetTestState(t)

	rule := Get("mvn_unknown_lifecycle_phase")
	if rule == nil {
		t.Fatal("mvn_unknown_lifecycle_phase rule is not registered")
	}

	assertMatch(t, rule, "mvn cle", mvnUnknownLifecyclePhaseOutput("cle"))

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

	assertNewCommands(t, rule, "mvn cle", mvnUnknownLifecyclePhaseOutput("cle"),
		[]string{"mvn clean", "mvn compile"})
	assertNewCommands(t, rule, "mvn claen package",
		mvnUnknownLifecyclePhaseOutput("claen"),
		[]string{"mvn clean package"})
}
