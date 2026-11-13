package specific

import (
	"errors"
	"reflect"
	"testing"
)

// npm6ScriptOutput is the script list of npm 6.
const npm6ScriptOutput = "Scripts available in audit@1.0.0 via `npm run-script`:\n" +
	"  build\n" +
	"    echo build\n" +
	"  test\n" +
	"    echo test\n"

// npm11ScriptOutput is the script list of npm 7 and newer, which splits the
// lifecycle scripts from the scripts available via npm run.
const npm11ScriptOutput = "Lifecycle scripts included in audit@1.0.0:\n" +
	"  test\n" +
	"    echo test\n" +
	"available via `npm run`:\n" +
	"  build\n" +
	"    echo build\n"

func TestParseNpmScripts(t *testing.T) {
	testCases := []struct {
		name   string
		output string
		want   []string
	}{
		{"npm 6", npm6ScriptOutput, []string{"build", "test"}},
		{"npm 7 and newer", npm11ScriptOutput, []string{"test", "build"}},
		{"empty", "", nil},
		{"unrelated", "added 3 packages\n", nil},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := parseNpmScripts(testCase.output); !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("parseNpmScripts() = %q, want %q", got, testCase.want)
			}
		})
	}
}

// stubNpmRunScript replaces the npm invocation of the package.
func stubNpmRunScript(t *testing.T, output string, err error) {
	t.Helper()
	original := npmRunScriptOutput
	npmRunScriptOutput = func() ([]byte, error) { return []byte(output), err }
	t.Cleanup(func() { npmRunScriptOutput = original })
}

// stubPackageScripts replaces the package.json script reader of the package.
func stubPackageScripts(t *testing.T, scripts []string) {
	t.Helper()
	original := PackageScripts
	PackageScripts = func() []string { return scripts }
	t.Cleanup(func() { PackageScripts = original })
}

func TestNpmScripts(t *testing.T) {
	t.Run("reads the npm list", func(t *testing.T) {
		stubNpmRunScript(t, npm11ScriptOutput, nil)
		stubPackageScripts(t, []string{"from-package-json"})
		if got := npmScripts(); !reflect.DeepEqual(got, []string{"test", "build"}) {
			t.Fatalf("npmScripts() = %q, want the npm list", got)
		}
	})

	t.Run("falls back for a failing npm", func(t *testing.T) {
		stubNpmRunScript(t, "", errors.New("npm is missing"))
		stubPackageScripts(t, []string{"build", "test"})
		if got := npmScripts(); !reflect.DeepEqual(got, []string{"build", "test"}) {
			t.Fatalf("npmScripts() = %q, want the package.json scripts", got)
		}
	})

	t.Run("falls back for an unknown output format", func(t *testing.T) {
		stubNpmRunScript(t, "something new\n", nil)
		stubPackageScripts(t, []string{"lint"})
		if got := npmScripts(); !reflect.DeepEqual(got, []string{"lint"}) {
			t.Fatalf("npmScripts() = %q, want the package.json scripts", got)
		}
	})
}
