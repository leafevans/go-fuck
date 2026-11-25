package rules

import (
	"strings"
	"testing"
)

// Verbatim diagnostic from Bazel's Windows CI log in
// https://github.com/bazelbuild/bazel/issues/3071 (line endings normalized).
const bazelMissingTargetUpstreamOutput = "ERROR: no such target '//py_test_dir/tensorflow/python:session_test': target 'session_test' not declared in package 'py_test_dir/tensorflow/python' (did you mean 'session_ops'?) defined by C:/jenkins/workspace/tensorflow/bazel_version/head/platform_name/windows-x86_64/py_test_dir/tensorflow/python/BUILD."

func TestRuleBazelMissingTarget(t *testing.T) {
	resetTestState(t)
	rule := Get("bazel_missing_target")
	const oldLabel = "//py_test_dir/tensorflow/python:session_test"
	const newLabel = "//py_test_dir/tensorflow/python:session_ops"
	// The hint list follows Bazel 8.0.0's TargetSuggesterTest.java.
	const modern = "ERROR: Skipping '//phone:pixel_test': no such target '//phone:pixel_test': target 'pixel_test' not declared in package 'phone' defined by /tmp/work space/phone/BUILD.bazel (did you mean pixel_5_test, or pixel_6_test?)"
	for _, tc := range []struct {
		name, script, output string
		want                 []string
	}{
		{"upstream diagnostic", "bazel test " + oldLabel, bazelMissingTargetUpstreamOutput, []string{"bazel test " + newLabel}},
		{"preserve quoting flags and whitespace", "bazel --bazelrc='a b' test\t  '" + oldLabel + "' --test_filter=\"session test\" --config=ci", "INFO: Loading\r\n" + bazelMissingTargetUpstreamOutput + "\r\n", []string{"bazel --bazelrc='a b' test\t  '" + newLabel + "' --test_filter=\"session test\" --config=ci"}},
		{"modern list in upstream order", "bazel build --config=dev \"//phone:pixel_test\"", modern, []string{"bazel build --config=dev \"//phone:pixel_5_test\"", "bazel build --config=dev \"//phone:pixel_6_test\""}},
		{"local root package", "bazel run //:bni -- --flag", "ERROR: no such target '//:bni': target 'bni' not declared in package '' defined by /tmp/BUILD (did you mean bin?)", []string{"bazel run //:bin -- --flag"}},
		{"same package explicit and deduplicated", "bazel test //phone:pixel_test", strings.Replace(modern, "pixel_5_test, or pixel_6_test", "//phone:pixel_5_test, :pixel_5_test, @@external//phone:pixel_6_test, or //other:pixel_6_test", 1), []string{"bazel test //phone:pixel_5_test"}},
		{"no explicit suggestion", "bazel test " + oldLabel, strings.Replace(bazelMissingTargetUpstreamOutput, " (did you mean 'session_ops'?)", "", 1), nil},
		{"different package", "bazel test " + oldLabel, strings.Replace(bazelMissingTargetUpstreamOutput, "'session_ops'", "'//other:session_ops'", 1), nil},
		{"external target", "bazel test @external" + oldLabel, strings.Replace(bazelMissingTargetUpstreamOutput, oldLabel, "@external"+oldLabel, 1), nil},
		{"wildcard", "bazel test " + oldLabel, strings.Replace(bazelMissingTargetUpstreamOutput, "'session_ops'", "'all'", 1), nil},
		{"shell syntax in suggestion", "bazel test " + oldLabel, strings.Replace(bazelMissingTargetUpstreamOutput, "'session_ops'", "'session_ops; echo foo'", 1), nil},
		{"flag value only", "bazel test --platforms=" + oldLabel, bazelMissingTargetUpstreamOutput, nil},
		{"ambiguous flag value", "bazel test --platforms " + oldLabel, bazelMissingTargetUpstreamOutput, nil},
		{"another target", "bazel test " + newLabel, bazelMissingTargetUpstreamOutput, nil},
		{"duplicate operand", "bazel test " + oldLabel + " " + oldLabel, bazelMissingTargetUpstreamOutput, nil},
		{"dependency diagnostic", "bazel test " + oldLabel, bazelMissingTargetUpstreamOutput + " and referenced by '//other:dep'", nil},
		{"run program argument", "bazel run //:main -- " + oldLabel, bazelMissingTargetUpstreamOutput, nil},
		{"not Bazel", "echo bazel test " + oldLabel, bazelMissingTargetUpstreamOutput, nil},
		{"compound command", "bazel test " + oldLabel + " && echo done", bazelMissingTargetUpstreamOutput, nil},
		{"multiple unresolved operands", "bazel test " + oldLabel + " //phone:pixel_test", bazelMissingTargetUpstreamOutput + "\n" + modern, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.want) != 0 {
				assertMatch(t, rule, tc.script, tc.output)
			} else {
				assertNotMatch(t, rule, tc.script, tc.output)
			}
			assertNewCommands(t, rule, tc.script, tc.output, tc.want)
		})
	}
}
