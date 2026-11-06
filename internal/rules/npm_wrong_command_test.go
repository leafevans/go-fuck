package rules

import "testing"

// npmWrongCommandUsageOutput is the npm help output printed for an unknown
// command. It contains the "verison" typo of the original test fixture.
const npmWrongCommandUsageOutput = `
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
    update, upgrade, v, verison, version, view, whoami

npm <cmd> -h     quick help on <cmd>
npm -l           display full usage info
npm faq          commonly asked questions
npm help <term>  search for help on <term>
npm help npm     involved overview

Specify configs in the ini-formatted file:
    /home/nvbn/.npmrc
or on the command line via: npm <command> --key value
Config info can be viewed via: npm help config

npm@2.14.7 /opt/node/lib/node_modules/npm
`

// npmWrongCommandModernOutput is the npm 7 and newer output for an unknown
// command. It no longer contains the command list; npm points at "npm help".
const npmWrongCommandModernOutput = `Unknown command: "isntall"

To see a list of supported npm commands, run:
  npm help
`

// npmWrongCommandModernHelpOutput is the "npm help" output with the commands
// that npm 7 and newer list.
const npmWrongCommandModernHelpOutput = `npm <command>

Usage:

npm install        install all the dependencies in your project
npm install <foo>  add the <foo> dependency to your project

All commands:

    access, adduser, audit, bugs, cache, ci, completion,
    config, dedupe, deprecate, diff, dist-tag, docs, doctor,
    edit, exec, explain, explore, find-dupes, fund, get, help,
    help-search, init, install, install-ci-test, install-test,
    link, ll, login, logout, ls, org, outdated, owner, pack,
    ping, pkg, prefix, profile, prune, publish, query, rebuild,
    repo, restart, root, run, sbom, search, set, shrinkwrap,
    star, stars, start, stop, team, test, token, trust,
    undeprecate, uninstall, unpublish, unstar, update, version,
    view, whoami

Specify configs in the ini-formatted file:
    /home/user/.npmrc
`

func TestRuleNpmWrongCommand(t *testing.T) {
	resetTestState(t)

	rule := Get("npm_wrong_command")
	if rule == nil {
		t.Fatal("npm_wrong_command rule is not registered")
	}

	originalHelp := npmWrongCommandHelpCommands
	npmWrongCommandHelpCommands = func() []string {
		return npmWrongCommandParseHelp(npmWrongCommandModernHelpOutput)
	}
	t.Cleanup(func() { npmWrongCommandHelpCommands = originalHelp })

	// npm 7 and newer no longer embed the command list in the error output.
	assertMatch(t, rule, "npm isntall", npmWrongCommandModernOutput)
	assertNotMatch(t, rule, "npm", npmWrongCommandModernOutput)
	assertNotMatch(t, rule, "test isntall", npmWrongCommandModernOutput)
	assertNewCommands(t, rule, "npm isntall -g gulp", npmWrongCommandModernOutput,
		[]string{"npm install -g gulp"})

	for _, script := range []string{
		"npm urgrdae",
		"npm urgrade -g",
		"npm -f urgrade -g",
		"npm urg",
	} {
		assertMatch(t, rule, script, npmWrongCommandUsageOutput)
	}

	assertNotMatch(t, rule, "npm urgrade", "")
	assertNotMatch(t, rule, "npm", npmWrongCommandUsageOutput)
	assertNotMatch(t, rule, "test urgrade", npmWrongCommandUsageOutput)
	assertNotMatch(t, rule, "npm -e", npmWrongCommandUsageOutput)

	assertNewCommands(t, rule, "npm urgrade", npmWrongCommandUsageOutput,
		[]string{"npm upgrade"})
	assertNewCommands(t, rule, "npm -g isntall gulp", npmWrongCommandUsageOutput,
		[]string{"npm -g install gulp"})
	assertNewCommands(t, rule, "npm isntall -g gulp", npmWrongCommandUsageOutput,
		[]string{"npm install -g gulp"})
}
