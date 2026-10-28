package rules

import (
	"fmt"
	"testing"
)

// pythonModuleErrorOutput builds the python traceback of a missing module.
func pythonModuleErrorOutput(filename, moduleName string) string {
	return fmt.Sprintf(`Traceback (most recent call last):
  File "%s", line 1, in <module>
    import %s
ModuleNotFoundError: No module named '%s'`, filename, moduleName, moduleName)
}

func TestRulePythonModuleError(t *testing.T) {
	resetTestState(t)
	rule := Get("python_module_error")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "python some_script.py", pythonModuleErrorOutput("some_script.py", "more_itertools"))
		assertMatch(t, rule, "./some_other_script.py", pythonModuleErrorOutput("some_other_script.py", "a_module"))
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "python hello_world.py", "Hello World")
		assertNotMatch(t, rule, "./hello_world.py", `Traceback (most recent call last):
  File "hello_world.py", line 1, in <module>
    pritn("Hello World")
NameError: name 'pritn' is not defined`)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "python some_script.py", pythonModuleErrorOutput("some_script.py", "more_itertools"),
			[]string{"pip install more_itertools && python some_script.py"})
		assertNewCommands(t, rule, "./some_other_script.py", pythonModuleErrorOutput("some_other_script.py", "a_module"),
			[]string{"pip install a_module && ./some_other_script.py"})
	})
}
