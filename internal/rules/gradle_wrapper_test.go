package rules

import "testing"

// gradleWrapperStub replaces the executable lookup and the gradlew file check.
func gradleWrapperStub(t *testing.T, whichResult string, exists bool) {
	t.Helper()
	originalWhich := gradleWrapperWhich
	originalExists := gradleWrapperExists
	gradleWrapperWhich = func(string) string { return whichResult }
	gradleWrapperExists = func() bool { return exists }
	t.Cleanup(func() {
		gradleWrapperWhich = originalWhich
		gradleWrapperExists = originalExists
	})
}

func TestRuleGradleWrapper(t *testing.T) {
	resetTestState(t)

	rule := Get("gradle_wrapper")
	if rule == nil {
		t.Fatal("gradle_wrapper rule is not registered")
	}

	gradleWrapperStub(t, "", true)
	assertMatch(t, rule, "gradle tasks", "gradle: not found")
	assertMatch(t, rule, "gradle build", "gradle: not found")

	gradleWrapperStub(t, "", false)
	assertNotMatch(t, rule, "gradle tasks", "gradle: not found")

	gradleWrapperStub(t, "/usr/bin/gradle", true)
	assertNotMatch(t, rule, "gradle tasks", "command not found")

	gradleWrapperStub(t, "", true)
	assertNotMatch(t, rule, "npm tasks", "npm: not found")

	assertNewCommands(t, rule, "gradle assemble", "",
		[]string{"./gradlew assemble"})
	assertNewCommands(t, rule, "gradle --help", "",
		[]string{"./gradlew --help"})
	assertNewCommands(t, rule, "gradle build -c", "",
		[]string{"./gradlew build -c"})
}
