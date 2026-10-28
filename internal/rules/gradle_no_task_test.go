package rules

import "testing"

// gradleNoTaskOutputNotFound returns the gradle output for an unknown task.
func gradleNoTaskOutputNotFound(task string) string {
	return `

FAILURE: Build failed with an exception.

* What went wrong:
Task '` + task + `' not found in root project 'org.rerenderer_example.snake'.

* Try:
Run gradlew tasks to get a list of available tasks. Run with --stacktrace option to get the stack trace. Run with --info or --debug option to get more log output.
`
}

// gradleNoTaskOutputAmbiguous returns the gradle output for an ambiguous task.
func gradleNoTaskOutputAmbiguous(task string) string {
	return `

FAILURE: Build failed with an exception.

* What went wrong:
Task '` + task + `' is ambiguous in root project 'org.rerenderer_example.snake'. Candidates are: 'assembleRelease', 'assembleReleaseUnitTest'.

* Try:
Run gradlew tasks to get a list of available tasks. Run with --stacktrace option to get the stack trace. Run with --info or --debug option to get more log output.
`
}

// gradleNoTaskTasksOutput is the "gradle tasks" output of the original test
// fixture.
const gradleNoTaskTasksOutput = `
:tasks

------------------------------------------------------------
All tasks runnable from root project
------------------------------------------------------------

Android tasks
-------------
androidDependencies - Displays the Android dependencies of the project.
signingReport - Displays the signing info for each variant.
sourceSets - Prints out all the source sets defined in this project.

Build tasks
-----------
assemble - Assembles all variants of all applications and secondary packages.
assembleAndroidTest - Assembles all the Test applications.
assembleDebug - Assembles all Debug builds.
assembleRelease - Assembles all Release builds.
build - Assembles and tests this project.
buildDependents - Assembles and tests this project and all projects that depend on it.
buildNeeded - Assembles and tests this project and all projects it depends on.
compileDebugAndroidTestSources
compileDebugSources
compileDebugUnitTestSources
compileReleaseSources
compileReleaseUnitTestSources
extractDebugAnnotations - Extracts Android annotations for the debug variant into the archive file
extractReleaseAnnotations - Extracts Android annotations for the release variant into the archive file
mockableAndroidJar - Creates a version of android.jar that's suitable for unit tests.

Build Setup tasks
-----------------
init - Initializes a new Gradle build. [incubating]
wrapper - Generates Gradle wrapper files. [incubating]

Help tasks
----------
components - Displays the components produced by root project 'org.rerenderer_example.snake'. [incubating]
dependencies - Displays all dependencies declared in root project 'org.rerenderer_example.snake'.
dependencyInsight - Displays the insight into a specific dependency in root project 'org.rerenderer_example.snake'.
help - Displays a help message.
model - Displays the configuration model of root project 'org.rerenderer_example.snake'. [incubating]
projects - Displays the sub-projects of root project 'org.rerenderer_example.snake'.
properties - Displays the properties of root project 'org.rerenderer_example.snake'.
tasks - Displays the tasks runnable from root project 'org.rerenderer_example.snake' (some of the displayed tasks may belong to subprojects).

Install tasks
-------------
installDebug - Installs the Debug build.
installDebugAndroidTest - Installs the android (on device) tests for the Debug build.
installRelease - Installs the Release build.
uninstallAll - Uninstall all applications.
uninstallDebug - Uninstalls the Debug build.
uninstallDebugAndroidTest - Uninstalls the android (on device) tests for the Debug build.
uninstallRelease - Uninstalls the Release build.

React tasks
-----------
bundleDebugJsAndAssets - bundle JS and assets for Debug.
bundleReleaseJsAndAssets - bundle JS and assets for Release.

Verification tasks
------------------
check - Runs all checks.
clean - Deletes the build directory.
connectedAndroidTest - Installs and runs instrumentation tests for all flavors on connected devices.
connectedCheck - Runs all device checks on currently connected devices.
connectedDebugAndroidTest - Installs and runs the tests for debug on connected devices.
deviceAndroidTest - Installs and runs instrumentation tests using all Device Providers.
deviceCheck - Runs all device checks using Device Providers and Test Servers.
lint - Runs lint on all variants.
lintDebug - Runs lint on the Debug build.
lintRelease - Runs lint on the Release build.
test - Run unit tests for all variants.
testDebugUnitTest - Run unit tests for the debug build.
testReleaseUnitTest - Run unit tests for the release build.

Other tasks
-----------
assembleDefault
copyDownloadableDepsToLibs
jarDebugClasses
jarReleaseClasses

To see all tasks and more detail, run gradlew tasks --all

To see more detail about a task, run gradlew help --task <task>

BUILD SUCCESSFUL

Total time: 1.936 secs
`

// gradleNoTaskStubTasks replaces the gradle task lookup with the fixture.
func gradleNoTaskStubTasks(t *testing.T) {
	t.Helper()
	original := gradleTasks
	gradleTasks = func(string) []string { return gradleNoTaskParseTasks(gradleNoTaskTasksOutput) }
	t.Cleanup(func() { gradleTasks = original })
}

// gradleNoTaskAssertFirstCommand checks the best suggestion. The original test
// only compares the first corrected command as well.
func gradleNoTaskAssertFirstCommand(t *testing.T, script, output, want string) {
	t.Helper()
	rule := Get("gradle_no_task")
	got, err := rule.GetNewCommand(testCommand(script, output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 {
		t.Fatalf("rule %q returned no corrected commands for %q", rule.Name, script)
	}
	if got[0] != want {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, want)
	}
}

func TestRuleGradleNoTask(t *testing.T) {
	resetTestState(t)
	gradleNoTaskStubTasks(t)

	rule := Get("gradle_no_task")
	if rule == nil {
		t.Fatal("gradle_no_task rule is not registered")
	}

	assertMatch(t, rule, "./gradlew assembler", gradleNoTaskOutputAmbiguous("assembler"))
	assertMatch(t, rule, "./gradlew instar", gradleNoTaskOutputNotFound("instar"))
	assertMatch(t, rule, "gradle assembler", gradleNoTaskOutputAmbiguous("assembler"))
	assertMatch(t, rule, "gradle instar", gradleNoTaskOutputNotFound("instar"))

	assertNotMatch(t, rule, "./gradlew assemble", "")
	assertNotMatch(t, rule, "gradle assemble", "")
	assertNotMatch(t, rule, "npm assembler", gradleNoTaskOutputAmbiguous("assembler"))
	assertNotMatch(t, rule, "npm instar", gradleNoTaskOutputNotFound("instar"))

	gradleNoTaskAssertFirstCommand(t, "./gradlew assembler",
		gradleNoTaskOutputAmbiguous("assembler"), "./gradlew assemble")
	gradleNoTaskAssertFirstCommand(t, "./gradlew instardebug",
		gradleNoTaskOutputNotFound("instardebug"), "./gradlew installDebug")
	gradleNoTaskAssertFirstCommand(t, "gradle assembler",
		gradleNoTaskOutputAmbiguous("assembler"), "gradle assemble")
	gradleNoTaskAssertFirstCommand(t, "gradle instardebug",
		gradleNoTaskOutputNotFound("instardebug"), "gradle installDebug")
}
