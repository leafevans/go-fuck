package rules

import (
	"fmt"
	"os"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
)

// fixFileTest is one output format covered by the original test suite.
type fixFileTest struct {
	script string
	file   string
	line   string
	col    string
	output string
}

// fixFileTests lists the error output formats of the supported tools.
var fixFileTests = []fixFileTest{
	{
		script: "gcc a.c",
		file:   "a.c",
		line:   "3",
		col:    "1",
		output: "\na.c: In function 'main':\na.c:3:1: error: expected expression before '}' token\n }\n  ^\n",
	},
	{
		script: "clang a.c",
		file:   "a.c",
		line:   "3",
		col:    "1",
		output: "\na.c:3:1: error: expected expression\n}\n^\n",
	},
	{
		script: "perl a.pl",
		file:   "a.pl",
		line:   "3",
		col:    "",
		output: "\nsyntax error at a.pl line 3, at EOF\nExecution of a.pl aborted due to compilation errors.\n",
	},
	{
		script: "perl a.pl",
		file:   "a.pl",
		line:   "2",
		col:    "",
		output: "\nSearch pattern not terminated at a.pl line 2.\n",
	},
	{
		script: "sh a.sh",
		file:   "a.sh",
		line:   "2",
		col:    "",
		output: "\na.sh: line 2: foo: command not found\n",
	},
	{
		script: "zsh a.sh",
		file:   "a.sh",
		line:   "2",
		col:    "",
		output: "\na.sh:2: command not found: foo\n",
	},
	{
		script: "bash a.sh",
		file:   "a.sh",
		line:   "2",
		col:    "",
		output: "\na.sh: line 2: foo: command not found\n",
	},
	{
		script: "rustc a.rs",
		file:   "a.rs",
		line:   "2",
		col:    "5",
		output: "\na.rs:2:5: 2:6 error: unexpected token: `+`\na.rs:2     +\n           ^\n",
	},
	{
		script: "cargo build",
		file:   "src/lib.rs",
		line:   "3",
		col:    "5",
		output: "\n   Compiling test v0.1.0 (file:///tmp/fix-error/test)\n   src/lib.rs:3:5: 3:6 error: unexpected token: `+`\n   src/lib.rs:3     +\n                    ^\nCould not compile `test`.\n\nTo learn more, run the command again with --verbose.\n",
	},
	{
		script: "python a.py",
		file:   "a.py",
		line:   "2",
		col:    "",
		output: "\n  File \"a.py\", line 2\n      +\n          ^\nSyntaxError: invalid syntax\n",
	},
	{
		script: "python a.py",
		file:   "a.py",
		line:   "8",
		col:    "",
		output: "\nTraceback (most recent call last):\n  File \"a.py\", line 8, in <module>\n    match(\"foo\")\n  File \"a.py\", line 5, in match\n    m = re.search(None, command)\n  File \"/usr/lib/python3.4/re.py\", line 170, in search\n    return _compile(pattern, flags).search(string)\n  File \"/usr/lib/python3.4/re.py\", line 293, in _compile\n    raise TypeError(\"first argument must be string or compiled pattern\")\nTypeError: first argument must be string or compiled pattern\n",
	},
	{
		script: "python caf\u00e9.py",
		file:   "caf\u00e9.py",
		line:   "8",
		col:    "",
		output: "\nTraceback (most recent call last):\n  File \"caf\u00e9.py\", line 8, in <module>\n    match(\"foo\")\n  File \"caf\u00e9.py\", line 5, in match\n    m = re.search(None, command)\n  File \"/usr/lib/python3.4/re.py\", line 170, in search\n    return _compile(pattern, flags).search(string)\n  File \"/usr/lib/python3.4/re.py\", line 293, in _compile\n    raise TypeError(\"first argument must be string or compiled pattern\")\nTypeError: first argument must be string or compiled pattern\n",
	},
	{
		script: "ruby a.rb",
		file:   "a.rb",
		line:   "3",
		col:    "",
		output: "\na.rb:3: syntax error, unexpected keyword_end\n",
	},
	{
		script: "lua a.lua",
		file:   "a.lua",
		line:   "2",
		col:    "",
		output: "\nlua: a.lua:2: unexpected symbol near '+'\n",
	},
	{
		script: "fish a.sh",
		file:   "/tmp/fix-error/a.sh",
		line:   "2",
		col:    "",
		output: "\nfish: Unknown command 'foo'\n/tmp/fix-error/a.sh (line 2): foo\n                              ^\n",
	},
	{
		script: "./a",
		file:   "./a",
		line:   "2",
		col:    "",
		output: "\nawk: ./a:2: BEGIN { print \"Hello, world!\" + }\nawk: ./a:2:                                 ^ syntax error\n",
	},
	{
		script: "llc a.ll",
		file:   "a.ll",
		line:   "1",
		col:    "2",
		output: "\nllc: a.ll:1:2: error: expected top-level entity\n+\n^\n",
	},
	{
		script: "go build a.go",
		file:   "a.go",
		line:   "1",
		col:    "2",
		output: "\ncan't load package:\na.go:1:2: expected 'package', found '+'\n",
	},
	{
		script: "make",
		file:   "Makefile",
		line:   "2",
		col:    "",
		output: "\nbidule\nmake: bidule: Command not found\nMakefile:2: recipe for target 'target' failed\nmake: *** [target] Error 127\n",
	},
	{
		script: "git st",
		file:   "/home/martin/.config/git/config",
		line:   "1",
		col:    "",
		output: "\nfatal: bad config file line 1 in /home/martin/.config/git/config\n",
	},
	{
		script: "node fuck.js asdf qwer",
		file:   "/Users/pablo/Workspace/barebones/fuck.js",
		line:   "2",
		col:    "5",
		output: "\n/Users/pablo/Workspace/barebones/fuck.js:2\nconole.log(arg);  // this should read console.log(arg);\n^\nReferenceError: conole is not defined\n    at /Users/pablo/Workspace/barebones/fuck.js:2:5\n    at Array.forEach (native)\n    at Object.<anonymous> (/Users/pablo/Workspace/barebones/fuck.js:1:85)\n    at Module._compile (module.js:460:26)\n    at Object.Module._extensions..js (module.js:478:10)\n    at Module.load (module.js:355:32)\n    at Function.Module._load (module.js:310:12)\n    at Function.Module.runMain (module.js:501:10)\n    at startup (node.js:129:16)\n    at node.js:814:3\n",
	},
	{
		script: "pep8",
		file:   "./tests/rules/test_systemctl.py",
		line:   "17",
		col:    "80",
		output: "\n./tests/rules/test_systemctl.py:17:80: E501 line too long (93 > 79 characters)\n./tests/rules/test_systemctl.py:18:80: E501 line too long (103 > 79 characters)\n./tests/rules/test_whois.py:20:80: E501 line too long (89 > 79 characters)\n./tests/rules/test_whois.py:22:80: E501 line too long (83 > 79 characters)\n",
	},
	{
		script: "pytest",
		file:   "/home/thefuck/tests/rules/test_fix_file.py",
		line:   "218",
		col:    "",
		output: "\nmonkeypatch = <_pytest.monkeypatch.monkeypatch object at 0x7fdb76a25b38>\ntest = ('fish a.sh', '/tmp/fix-error/a.sh', 2, None, '', \"\\nfish: Unknown command 'foo'\\n/tmp/fix-error/a.sh (line 2): foo\\n                              ^\\n\")\n\n    @pytest.mark.parametrize('test', tests)\n    @pytest.mark.usefixtures('no_memoize')\n    def test_get_new_command(monkeypatch, test):\n>       mocker.patch('os.path.isfile', return_value=True)\nE       NameError: name 'mocker' is not defined\n\n/home/thefuck/tests/rules/test_fix_file.py:218: NameError\n",
	},
}

// fixFileSetEditor sets EDITOR for the duration of the test.
func fixFileSetEditor(t *testing.T, editor string) {
	t.Helper()
	t.Setenv("EDITOR", editor)
}

// fixFileUnsetEditor removes EDITOR for the duration of the test.
func fixFileUnsetEditor(t *testing.T) {
	t.Helper()
	original, ok := os.LookupEnv("EDITOR")
	if err := os.Unsetenv("EDITOR"); err != nil {
		t.Fatalf("unset EDITOR: %v", err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv("EDITOR", original)
			return
		}
		_ = os.Unsetenv("EDITOR")
	})
}

// fixFileSetup installs the fake file check.
func fixFileSetup(t *testing.T, isFile bool) {
	t.Helper()
	original := fixFileIsFile
	fixFileIsFile = func(string) bool { return isFile }
	t.Cleanup(func() { fixFileIsFile = original })
}

func TestRuleFixFile(t *testing.T) {
	resetTestState(t)
	rule := Get("fix_file")
	if rule == nil {
		t.Fatal("fix_file rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		fixFileSetup(t, true)
		fixFileSetEditor(t, "dummy_editor")
		for _, testCase := range fixFileTests {
			assertMatch(t, rule, "", testCase.output)
		}
	})

	t.Run("no editor", func(t *testing.T) {
		fixFileSetup(t, true)
		fixFileUnsetEditor(t)
		for _, testCase := range fixFileTests {
			assertNotMatch(t, rule, "", testCase.output)
		}
	})

	t.Run("not a file", func(t *testing.T) {
		fixFileSetup(t, false)
		fixFileSetEditor(t, "dummy_editor")
		for _, testCase := range fixFileTests {
			assertNotMatch(t, rule, "", testCase.output)
		}
	})

	t.Run("default command", func(t *testing.T) {
		fixFileSetup(t, true)
		fixFileSetEditor(t, "dummy_editor")
		for _, testCase := range fixFileTests {
			want := fmt.Sprintf("dummy_editor %s +%s && %s", testCase.file, testCase.line, testCase.script)
			assertNewCommands(t, rule, testCase.script, testCase.output, []string{want})
		}
	})

	t.Run("column command", func(t *testing.T) {
		fixFileSetup(t, true)
		fixFileSetEditor(t, "dummy_editor")
		conf.Current().SetExtra("fixcolcmd", "{editor} {file} +{line}:{col}")
		t.Cleanup(func() { conf.Current().SetExtra("fixcolcmd", "") })

		for _, testCase := range fixFileTests {
			var want string
			if testCase.col != "" {
				want = fmt.Sprintf("dummy_editor %s +%s:%s && %s",
					testCase.file, testCase.line, testCase.col, testCase.script)
			} else {
				want = fmt.Sprintf("dummy_editor %s +%s && %s", testCase.file, testCase.line, testCase.script)
			}
			assertNewCommands(t, rule, testCase.script, testCase.output, []string{want})
		}
	})
}
