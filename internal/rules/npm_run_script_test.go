package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/specific"
)

// npmRunScriptUsageOutput is the npm help output printed for an unknown
// command.
const npmRunScriptUsageOutput = `
Usage: npm <command>

where <command> is one of:
    access, add-user, adduser, apihelp, author, bin, bugs, c,
    cache, completion, config, ddp, dedupe, deprecate, dist-tag,
    dist-tags, docs, edit, explore, faq, find, find-dupes, get,
    help, help-search, home, i, info, init, install, issues, la,
    link, list, ll, ln, login, logout, ls, outdated, owner,
    pack, ping, prefix, prune, publish, r, rb, rebuild, remove,
    repo, restart, rm, root, run-script, s, se, search, set,
    show, shrinkwrap, star, stars, start, stop, t, tag, team,
    test, tst, un, uninstall, unlink, unpublish, unstar, up,
    update, upgrade, v, version, view, whoami

npm <cmd> -h     quick help on <cmd>
npm -l           display full usage info
npm faq          commonly asked questions
npm help <term>  search for help on <term>
npm help npm     involved overview

Specify configs in the ini-formatted file:
    /home/nvbn/.npmrc
or on the command line via: npm <command> --key value
Config info can be viewed via: npm help config

`

// npmRunScriptStubScripts replaces the npm scripts with the fixture values.
func npmRunScriptStubScripts(t *testing.T) {
	t.Helper()
	original := specific.NpmGetScripts
	specific.NpmGetScripts = func() []string {
		return []string{"build", "develop", "watch-test"}
	}
	t.Cleanup(func() { specific.NpmGetScripts = original })
}

func TestRuleNpmRunScript(t *testing.T) {
	resetTestState(t)
	npmRunScriptStubScripts(t)

	rule := Get("npm_run_script")
	if rule == nil {
		t.Fatal("npm_run_script rule is not registered")
	}

	for _, script := range []string{"npm watch-test", "npm develop"} {
		assertMatch(t, rule, script, npmRunScriptUsageOutput)
	}

	assertNotMatch(t, rule, "npm test", "TEST FAIL")
	assertNotMatch(t, rule, "npm watch-test", "TEST FAIL")
	assertNotMatch(t, rule, "npm test", npmRunScriptUsageOutput)
	assertNotMatch(t, rule, "vim watch-test", npmRunScriptUsageOutput)

	assertNewCommands(t, rule, "npm watch-test", npmRunScriptUsageOutput,
		[]string{"npm run-script watch-test"})
	assertNewCommands(t, rule, "npm -i develop", npmRunScriptUsageOutput,
		[]string{"npm run-script -i develop"})
	assertNewCommands(t, rule, "npm -i watch-script --path ..", npmRunScriptUsageOutput,
		[]string{"npm run-script -i watch-script --path .."})
}
