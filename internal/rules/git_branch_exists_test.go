package rules

import (
	"strings"
	"testing"
)

// gitBranchExistsExpected builds the commands the rule should return.
func gitBranchExistsExpected(branchName string) []string {
	templates := []string{
		"git branch -d {0} && git branch {0}",
		"git branch -d {0} && git checkout -b {0}",
		"git branch -D {0} && git branch {0}",
		"git branch -D {0} && git checkout -b {0}",
		"git checkout {0}",
	}
	results := make([]string, len(templates))
	for index, template := range templates {
		results[index] = strings.ReplaceAll(template, "{0}", branchName)
	}
	return results
}

func TestRuleGitBranchExists(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_exists")
	if rule == nil {
		t.Fatal("git_branch_exists rule is not registered")
	}

	output := func(srcBranchName string) string {
		return "fatal: A branch named '" + srcBranchName + "' already exists."
	}

	matchCases := []struct {
		script        string
		srcBranchName string
	}{
		{"git branch foo", "foo"},
		{"git checkout bar", "bar"},
		{"git checkout -b \"let's-push-this\"", "let's-push-this"},
	}
	for _, testCase := range matchCases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, output(testCase.srcBranchName))
		})
	}

	for _, script := range []string{
		"git branch foo",
		"git checkout bar",
		"git checkout -b \"let's-push-this\"",
	} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	escapedBranchName := "let" + backslash + "'s-push-this"
	commandCases := []struct {
		script        string
		srcBranchName string
		branchName    string
	}{
		{"git branch foo", "foo", "foo"},
		{"git checkout bar", "bar", "bar"},
		{"git checkout -b \"let's-push-this\"", "let's-push-this", escapedBranchName},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, output(testCase.srcBranchName),
				gitBranchExistsExpected(testCase.branchName))
		})
	}
}
