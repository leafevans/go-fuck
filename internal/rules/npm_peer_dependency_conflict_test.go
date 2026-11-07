package rules

import "testing"

// npmPeerDependencyConflictOutput is the npm 7 and newer output for a peer
// dependency conflict. npm 7 and 8 prefix the lines with "npm ERR!" and npm 9
// and newer with "npm error".
func npmPeerDependencyConflictOutput(prefix string) string {
	return prefix + " code ERESOLVE\n" +
		prefix + " ERESOLVE unable to resolve dependency tree\n" +
		prefix + "\n" +
		prefix + " While resolving: probe@1.0.0\n" +
		prefix + " Found: react@17.0.2\n" +
		prefix + " node_modules/react\n" +
		prefix + "   react@\"17.0.2\" from the root project\n" +
		prefix + "\n" +
		prefix + " Could not resolve dependency:\n" +
		prefix + " peer react@\"^18.2.0\" from react-dom@18.2.0\n" +
		prefix + " node_modules/react-dom\n" +
		prefix + "   react-dom@\"18.2.0\" from the root project\n" +
		prefix + "\n" +
		prefix + " Fix the upstream dependency conflict, or retry\n" +
		prefix + " this command with --force or --legacy-peer-deps to accept an incorrect (and potentially broken) dependency resolution.\n"
}

func TestRuleNpmPeerDependencyConflict(t *testing.T) {
	resetTestState(t)
	rule := Get("npm_peer_dependency_conflict")
	if rule == nil {
		t.Fatal("npm_peer_dependency_conflict rule is not registered")
	}

	assertMatch(t, rule, "npm install react@17.0.2 react-dom@18.2.0", npmPeerDependencyConflictOutput("npm error"))
	assertMatch(t, rule, "npm install", npmPeerDependencyConflictOutput("npm ERR!"))
	assertMatch(t, rule, "npm i react@17.0.2", npmPeerDependencyConflictOutput("npm error"))
	assertNotMatch(t, rule, "pnpm install", npmPeerDependencyConflictOutput("npm error"))
	assertNotMatch(t, rule, "npm install --legacy-peer-deps react@17.0.2", npmPeerDependencyConflictOutput("npm error"))
	assertNotMatch(t, rule, "npm install --force react@17.0.2", npmPeerDependencyConflictOutput("npm error"))
	assertNotMatch(t, rule, "npm install react@17.0.2", "")

	assertNewCommands(t, rule, "npm install react@17.0.2 react-dom@18.2.0", npmPeerDependencyConflictOutput("npm error"),
		[]string{"npm install react@17.0.2 react-dom@18.2.0 --legacy-peer-deps"})
}
