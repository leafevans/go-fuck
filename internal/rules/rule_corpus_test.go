package rules

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// ruleCorpusCase is a command line and the output it produced. The corpus is
// shared by the deterministic rule tests and by the fuzz seeds, so every rule
// keeps seeing the failure modes of the CLIs it was written for.
type ruleCorpusCase struct {
	script string
	output string
}

// ruleCorpus returns realistic failing commands with the output of current CLI
// versions, plus deliberately malformed input. It is intentionally kept small:
// every rule is run against every case, so the corpus is a regression net for
// panics and hangs, not a full test of the rules (each rule has its own test).
func ruleCorpus() []ruleCorpusCase {
	long := strings.Repeat("x", 2048)
	return []ruleCorpusCase{
		{"", ""},
		{" ", "\x00\x01\xff"},
		{"\x00\x00", "\x1b[31merror\x1b[0m\r\n"},
		{long, long},
		{"git \\ \" '", "\t\n"},
		{"git", ""},
		{"git brnch", "git: 'brnch' is not a git command. See 'git --help'."},
		{"git comit -m x", "git: 'comit' is not a git command. See 'git --help'."},
		{"git push", "fatal: The current branch main has no upstream branch.\nTo push the current branch and set the remote as upstream, use\n\n    git push --set-upstream origin main\n"},
		{"git push origin main", "! [rejected]        main -> main (fetch first)\nerror: failed to push some refs to 'origin'\nhint: Updates were rejected because the remote contains work that you do not have locally.\n"},
		{"git pull", "fatal: Not possible to fast-forward, aborting."},
		{"git checkout feat", "error: pathspec 'feat' did not match any file(s) known to git"},
		{"git switch feat2", "fatal: invalid reference: feat2"},
		{"git branch -d merged", "error: The branch 'merged' is not fully merged."},
		{"git merge other", "fatal: refusing to merge unrelated histories"},
		{"git stash", "No local changes to save"},
		{"git clone repo", "fatal: destination path 'repo' already exists and is not an empty directory."},
		{"git remote add origin url", "fatal: remote origin already exists."},
		{"git tag -f v1", "fatal: tag 'v1' already exists"},
		{"git rm file", "error: the following file has staged content different from both the file and the HEAD\n(use -f to force removal)"},
		{"git add", "fatal: No pathspec was given"},
		{"git diff --cached", "fatal: bad revision 'cached'"},
		{"git lfs status", "This repository is configured for Git LFS but 'git-lfs' was not found on your path.\nIf you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push."},
		{"git bisect bid", "You need to start by \"git bisect start\""},
		{"git submodule update", "fatal: No url found for submodule path 'vendor' in .gitmodules"},
		{"apt-get install pkg", "E: Unable to locate package pkg"},
		{"apt-get instal", "E: Invalid operation instal"},
		{"apt install pkg", "E: Package 'pkg' has no installation candidate"},
		{"apt list --instaled", "E: Invalid operation --instaled"},
		{"apt-get upgrade", "WARNING: You are doing an upgrade without an update."},
		{"yum install pkg", "No such command: install. Please use /usr/bin/yum --help"},
		{"dnf install pkg", "No such command: install. Please use 'dnf --help'"},
		{"pacman -S pkg", "error: target not found: pkg"},
		{"pacman -Xy", "error: invalid option '-X'"},
		{"brew install pkg", "Error: No available formula with the name \"pkg\"."},
		{"brew instal pkg", "Error: Unknown command: instal\n\nRun 'brew help' to see usage."},
		{"npm install", "npm ERR! code ENOENT\nnpm ERR! syscall open\nnpm ERR! path /package.json"},
		{"npm run buld", "npm ERR! Missing script: \"buld\"\n\nTo see a list of scripts, run:\n  npm run\n\nnpm ERR! A complete log of this run can be found in:"},
		{"npm instll", "npm ERR! Unknown command: \"instll\"\n\nnpm ERR! where <command> is one of: access, adduser, audit, bin, bugs, c, cache, ci, cit\n"},
		{"npm ci", "npm ERR! npm ci can only install packages when your package.json and package-lock.json or npm-shrinkwrap.json are in sync. Please update your lock file with npm install before continuing."},
		{"yarn run buld", "error Command \"buld\" not found."},
		{"yarn instal", "error instal is not a yarn command.\ninfo The yarn command list: add, audit, autoclean, bin, cache, check, config"},
		{"pnpm buld", "ERR_PNPM_COMMAND_NOT_FOUND  Command \"buld\" not found"},
		{"pip instal pkg", "ERROR: unknown command \"instal\" - maybe you meant \"install\"?"},
		{"pip install pkg", "ERROR: Could not find a version that satisfies the requirement pkg (from versions: none)"},
		{"pip3 instal pkg", "ERROR: unknown command \"instal\" - maybe you meant \"install\"?"},
		{"poetry instal", "ERROR: unknown command \"instal\""},
		{"uv instal", "error: unrecognized subcommand 'instal'\n\n  tip: a similar subcommand exists: 'install'"},
		{"uv pip instl x", "error: unrecognized subcommand 'instl'"},
		{"cargo bulid", "error: no such command: 'bulid'\n\n\tDid you mean 'build'?"},
		{"cargo build", "error: failed to parse manifest at 'Cargo.toml'"},
		{"cargo test --benches", "error: no bench target named 'missing'"},
		{"rustup instal stable", "error: unknown subcommand 'instal'"},
		{"go bulid ./...", "go: unknown command \"bulid\""},
		{"go build ./...", "go: missing go.sum entry for module providing package example.com/pkg"},
		{"go run main.go missing.go", "go: no such file or directory missing.go"},
		{"go mod init", "go: go.mod file already exists"},
		{"dotnet bulid", "Could not execute because the specified command or file was not found."},
		{"mvn instal", "[ERROR] No plugin found for prefix 'instal'"},
		{"mvn test", "The specified lifecycle phase 'tst' is not a valid phase"},
		{"gradle tst", "Task 'tst' not found in root project 'demo'."},
		{"flutter bulid", "Could not find a command named \"bulid\"."},
		{"dart bulid", "Could not find a command named \"bulid\"."},
		{"docker ps", "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?"},
		{"docker compose up", "no such service: web"},
		{"docker run img", "docker: Error response from daemon: pull access denied for img."},
		{"kubectl get pod", "error: the server doesn't have a resource type \"pod\""},
		{"kubectl config use-context prod", "error: no context exists with the name: \"prod\""},
		{"helm instal", "Error: unknown command \"instal\" for \"helm\""},
		{"kustomize buld", "Error: unknown command \"buld\""},
		{"minikube start", "Exiting due to PROVIDER_DOCKER_NOT_RUNNING"},
		{"kind create cluster", "ERROR: failed to create cluster: node(s) already exist"},
		{"terraform init", "Error: Failed to install provider"},
		{"terraform stat", "Terraform has no command named \"stat\"."},
		{"gh pr creat", "unknown command \"creat\" for \"gh pr\""},
		{"gh pr view --hlep", "unknown flag: --hlep"},
		{"glab mr creat", "unknown command \"creat\" for \"glab mr\""},
		{"just task", "error: Justfile does not contain recipe 'task'."},
		{"task missing", "task: Failed to run task \"missing\": task does not exist"},
		{"make tst", "make: *** No rule to make target 'tst'.  Stop."},
		{"chmod u+x file", "chmod: cannot access 'file': No such file or directory"},
		{"cd nonexistent", "cd: no such file or directory: nonexistent"},
		{"rm dir", "rm: cannot remove 'dir': Is a directory"},
		{"cp a b", "cp: cannot stat 'a': No such file or directory"},
		{"cp -r a b", "cp: omitting directory 'a'"},
		{"mkdir dir", "mkdir: cannot create directory 'dir': File exists"},
		{"ln a b", "ln: failed to create hard link 'b': File exists"},
		{"touch --file", "touch: unrecognized option '--file'"},
		{"cat dir", "cat: dir: Is a directory"},
		{"ls -lah", "ls: cannot access 'nope': No such file or directory"},
		{"grep foo", "grep: foo: No such file or directory"},
		{"grep -r foo .", "grep: warning: foo: No such file or directory"},
		{"ssh host", "Host key verification failed."},
		{"scp a b", "ssh: Could not resolve hostname b: Name or service not known"},
		{"tar xf a.tar", "tar: a.tar: Cannot open: No such file or directory"},
		{"unzip archive", "unzip: cannot find or open archive.zip"},
		{"sudo foo", "sudo: foo: command not found"},
		{"systemctl restart foo", "Failed to restart foo.service: Unit foo.service not found."},
		{"ifconfig eth9", "eth9: error fetching interface information: Device not found"},
		{"whois example", "No match for \"EXAMPLE\"."},
		{"hostscli websites", "Error: No such command \"websites\""},
		{"rails server", "Migrations are pending. To resolve this issue, run:\n\n        bin/rails db:migrate"},
		{"python -m pip instal x", "No module named pip"},
		{"python script.py", "python: can't open file 'script.py': [Errno 2] No such file or directory"},
		{"node app.js", "Error: Cannot find module '/app.js'"},
		{"node script.js", "ReferenceError: foo is not defined"},
		{"vagrant up", "The machine with the name 'default' was not found"},
		{"history", "  1  git status"},
		{"cd ..", ""},
		{"fuck", ""},
		{"sed -i s/a/b file", "sed: -e expression #1, char 7: unterminated 's' command"},
		{"man ls", "No manual entry for ls"},
		{"ag foo", "ag: unrecognized option '--foo'"},
		{"tsuru app-log", "Error: unknown command \"app-log\""},
		{"heroku logs -a app2", "Multiple apps? Specify one using --app"},
		{"react-native run-android", "error Command \"run-android\" unrecognized"},
		{"adb instal", "adb: usage: unknown command instal"},
		{"gulp tst", "Task 'tst' was not found."},
		{"grunt tst", "Warning: Task \"tst\" not found."},
		{"composer instal", "Command \"instal\" is not defined."},
		{"gem instal rails", "ERROR:  While executing gem ... (Gem::CommandLineError)"},
		{"bundle instal", "Could not find gem 'rails' in locally installed gems."},
		{"conda install pkg", "PackagesNotFoundError: The following packages are not available"},
		{"df -h", "df: no such file or directory"},
		{"kill -9", "kill: usage: kill [-s sigspec | -n signum]"},
		{"port 8080", "lsof: not found"},
	}
}

// TestRuleCorpusDoesNotPanic runs every rule against the corpus through the
// normal matching path. A panic in a rule can be recovered by the corrector,
// but it still means the rule is broken, so it fails the suite instead.
func TestRuleCorpusDoesNotPanic(t *testing.T) {
	resetTestState(t)
	registered := All()
	if len(registered) == 0 {
		t.Fatal("no rules are registered")
	}
	corpus := ruleCorpus()
	for _, testCase := range corpus {
		command := types.NewCommand(testCase.script, testCase.output)
		for _, rule := range registered {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Errorf("rule %q panicked on script %q output %q: %v",
							rule.Name, testCase.script, testCase.output, recovered)
					}
				}()
				if rule.Match(command) {
					_, _ = rule.GetNewCommand(command)
				}
			}()
		}
	}
}

// TestRuleConstructorsDoNotPanic calls the command builders directly with
// malformed input, including scripts that the rule matchers normally reject.
// Several rules build their suggestions without checking the command shape, so
// this guards against index out of range panics such as the empty npm script.
func TestRuleConstructorsDoNotPanic(t *testing.T) {
	resetTestState(t)
	registered := All()
	if len(registered) == 0 {
		t.Fatal("no rules are registered")
	}
	commands := []*types.Command{
		types.NewCommand("", ""),
		types.NewCommand("\x00", "\x00"),
		types.NewCommand("npm", "npm ERR! Unknown command: \"x\"\n\nnpm ERR! where <command> is one of: install, run, test\n"),
	}
	for _, command := range commands {
		for _, rule := range registered {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Errorf("rule %q panicked on script %q output %q: %v",
							rule.Name, command.Script, command.Output, recovered)
					}
				}()
				_, _ = rule.GetNewCommand(command)
			}()
		}
	}
}
