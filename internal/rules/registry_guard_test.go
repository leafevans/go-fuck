package rules

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// expectedRuleNames is the complete rule set of the reference implementation:
// the module names of thefuck/thefuck/rules/*.py, excluding __init__.py.
//
// The list is hardcoded on purpose. The guard fails when a reference rule is
// missing from the port and when an unexpected rule is added, so that the Go
// implementation stays a faithful and complete reimplementation.
//
// Note: the unusual reference module file "test.py.py" maps to the rule name
// "test.py".
var expectedRuleNames = []string{
	"adb_unknown_command",
	"ag_literal",
	"apt_get",
	"apt_get_search",
	"apt_invalid_operation",
	"apt_list_upgradable",
	"apt_upgrade",
	"aws_cli",
	"az_cli",
	"brew_cask_dependency",
	"brew_install",
	"brew_link",
	"brew_reinstall",
	"brew_uninstall",
	"brew_unknown_command",
	"brew_update_formula",
	"cargo",
	"cargo_no_command",
	"cat_dir",
	"cd_correction",
	"cd_cs",
	"cd_mkdir",
	"cd_parent",
	"chmod_x",
	"choco_install",
	"composer_not_command",
	"conda_mistype",
	"cp_create_destination",
	"cp_omitting_directory",
	"cpp11",
	"dirty_untar",
	"dirty_unzip",
	"django_south_ghost",
	"django_south_merge",
	"dnf_no_such_command",
	"docker_image_being_used_by_container",
	"docker_login",
	"docker_not_command",
	"dry",
	"fab_command_not_found",
	"fix_alt_space",
	"fix_file",
	"gem_unknown_command",
	"git_add",
	"git_add_force",
	"git_bisect_usage",
	"git_branch_0flag",
	"git_branch_delete",
	"git_branch_delete_checked_out",
	"git_branch_exists",
	"git_branch_list",
	"git_checkout",
	"git_clone_git_clone",
	"git_clone_missing",
	"git_commit_add",
	"git_commit_amend",
	"git_commit_reset",
	"git_diff_no_index",
	"git_diff_staged",
	"git_fix_stash",
	"git_flag_after_filename",
	"git_help_aliased",
	"git_hook_bypass",
	"git_lfs_mistype",
	"git_main_master",
	"git_merge",
	"git_merge_unrelated",
	"git_not_command",
	"git_pull",
	"git_pull_clone",
	"git_pull_uncommitted_changes",
	"git_push",
	"git_push_different_branch_names",
	"git_push_force",
	"git_push_pull",
	"git_push_without_commits",
	"git_rebase_merge_dir",
	"git_rebase_no_changes",
	"git_remote_delete",
	"git_remote_seturl_add",
	"git_rm_local_modifications",
	"git_rm_recursive",
	"git_rm_staged",
	"git_stash",
	"git_stash_pop",
	"git_tag_force",
	"git_two_dashes",
	"go_run",
	"go_unknown_command",
	"gradle_no_task",
	"gradle_wrapper",
	"grep_arguments_order",
	"grep_recursive",
	"grunt_task_not_found",
	"gulp_not_task",
	"has_exists_script",
	"heroku_multiple_apps",
	"heroku_not_command",
	"history",
	"hostscli",
	"ifconfig_device_not_found",
	"java",
	"javac",
	"lein_not_task",
	"ln_no_hard_link",
	"ln_s_order",
	"long_form_help",
	"ls_all",
	"ls_lah",
	"man",
	"man_no_space",
	"mercurial",
	"missing_space_before_subcommand",
	"mkdir_p",
	"mvn_no_command",
	"mvn_unknown_lifecycle_phase",
	"nixos_cmd_not_found",
	"no_command",
	"no_such_file",
	"npm_missing_script",
	"npm_run_script",
	"npm_wrong_command",
	"omnienv_no_such_command",
	"open",
	"pacman",
	"pacman_invalid_option",
	"pacman_not_found",
	"path_from_history",
	"php_s",
	"pip_install",
	"pip_unknown_command",
	"port_already_in_use",
	"prove_recursively",
	"python_command",
	"python_execute",
	"python_module_error",
	"quotation_marks",
	"rails_migrations_pending",
	"react_native_command_unrecognized",
	"remove_shell_prompt_literal",
	"remove_trailing_cedilla",
	"rm_dir",
	"rm_root",
	"scm_correction",
	"sed_unterminated_s",
	"sl_ls",
	"ssh_known_hosts",
	"sudo",
	"sudo_command_from_user_path",
	"switch_lang",
	"systemctl",
	"terraform_init",
	"terraform_no_command",
	"test.py",
	"tmux",
	"touch",
	"tsuru_login",
	"tsuru_not_command",
	"unknown_command",
	"unsudo",
	"vagrant_up",
	"whois",
	"workon_doesnt_exists",
	"wrong_hyphen_before_subcommand",
	"yarn_alias",
	"yarn_command_not_found",
	"yarn_command_replaced",
	"yarn_help",
	"yum_invalid_operation",
}

// expectedExtraRuleNames lists rules that go-fuck adds on top of the reference
// implementation. They are tested and documented like the ported rules, but
// they are not part of the thefuck rule set.
var expectedExtraRuleNames = []string{
	"cargo_unknown_argument",
	"cargo_unknown_command",
	"gh_unknown_command",
	"gh_unknown_flag",
	"go_install_version",
	"go_missing_module",
	"go_unknown_flag",
	"helm_unknown_command",
	"kubectl_unknown_command",
	"node_missing_module",
	"pnpm_command_not_found",
	"uv_unknown_argument",
	"uv_unknown_subcommand",
}

// allExpectedRuleNames returns the reference rules plus the go-fuck specific
// extra rules.
func allExpectedRuleNames() []string {
	names := make([]string, 0, len(expectedRuleNames)+len(expectedExtraRuleNames))
	names = append(names, expectedRuleNames...)
	names = append(names, expectedExtraRuleNames...)
	return names
}

// registryGuardHelperFiles lists the files of this package that do not
// implement or test a single rule. They are skipped by the file guards.
var registryGuardHelperFiles = map[string]bool{
	"registry.go":            true,
	"testhelpers_test.go":    true,
	"readme_test.go":         true,
	"registry_guard_test.go": true,
	"benchmark_test.go":      true,
	"cobra_suggestions.go":   true,
}

// ruleNamePattern matches the characters that a rule name may contain. Rule
// names are derived from Python module names.
var ruleNamePattern = regexp.MustCompile("^[a-z0-9][a-z0-9_.]*$")

// ruleRegistryNames returns the names of all registered rules.
func ruleRegistryNames() []string {
	names := []string{}
	for _, rule := range All() {
		names = append(names, rule.Name)
	}
	return names
}

// TestRuleRegistryCompleteness guards that every rule of the reference
// implementation is registered and that no rule is missing or unexpected.
func TestRuleRegistryCompleteness(t *testing.T) {
	resetTestState(t)

	registered := map[string]bool{}
	for _, name := range ruleRegistryNames() {
		registered[name] = true
	}

	expectedNames := allExpectedRuleNames()
	expected := map[string]bool{}
	missing := []string{}
	for _, name := range expectedNames {
		expected[name] = true
		if !registered[name] {
			missing = append(missing, name)
		}
	}

	unexpected := []string{}
	for name := range registered {
		if !expected[name] {
			unexpected = append(unexpected, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)

	if len(missing) > 0 || len(unexpected) > 0 {
		t.Fatalf("rule registry mismatch: %d of %d reference and extra rules registered; missing: [%s]; unexpected: [%s]",
			len(expectedNames)-len(missing), len(expectedNames),
			strings.Join(missing, ", "), strings.Join(unexpected, ", "))
	}
}

// TestRuleRegistryUniqueNames guards that no rule name is registered twice.
func TestRuleRegistryUniqueNames(t *testing.T) {
	resetTestState(t)

	expectedSeen := map[string]bool{}
	for _, name := range allExpectedRuleNames() {
		if expectedSeen[name] {
			t.Fatalf("expectedRuleNames contains a duplicate: %q", name)
		}
		expectedSeen[name] = true
	}

	seen := map[string]bool{}
	for _, rule := range All() {
		if seen[rule.Name] {
			t.Fatalf("rule %q is registered more than once", rule.Name)
		}
		seen[rule.Name] = true
	}
}

// TestRuleRegistryRuleQuality guards that every registered rule is usable and
// has sensible metadata.
func TestRuleRegistryRuleQuality(t *testing.T) {
	resetTestState(t)

	for _, rule := range All() {
		t.Run(rule.Name, func(t *testing.T) {
			if rule.Name == "" {
				t.Fatal("rule has an empty name")
			}
			if !ruleNamePattern.MatchString(rule.Name) {
				t.Errorf("rule name %q contains unexpected characters", rule.Name)
			}
			if rule.Match == nil {
				t.Error("rule has a nil Match function")
			}
			if rule.GetNewCommand == nil {
				t.Error("rule has a nil GetNewCommand function")
			}
			if rule.Priority <= 0 {
				t.Errorf("rule priority = %d, want a positive priority", rule.Priority)
			}
		})
	}
}

// TestRuleRegistryFiles guards that every reference rule has an implementation
// file and a test file, and that the package does not contain stray rule
// files.
func TestRuleRegistryFiles(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read rules package directory: %v", err)
	}

	files := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			files[entry.Name()] = true
		}
	}

	missing := []string{}
	expectedFiles := map[string]bool{}
	for _, name := range allExpectedRuleNames() {
		// A rule name maps to <name>.go and <name>_test.go, so the unusual
		// "test.py" rule maps to "test.py.go" and "test.py_test.go".
		ruleFile := name + ".go"
		testFile := name + "_test.go"
		expectedFiles[ruleFile] = true
		expectedFiles[testFile] = true
		if !files[ruleFile] {
			missing = append(missing, ruleFile)
		}
		if !files[testFile] {
			missing = append(missing, testFile)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("%d rule files are missing: [%s]",
			len(missing), strings.Join(missing, ", "))
	}

	unexpected := []string{}
	for name := range files {
		if !strings.HasSuffix(name, ".go") || expectedFiles[name] || registryGuardHelperFiles[name] {
			continue
		}
		unexpected = append(unexpected, name)
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		t.Fatalf("unexpected Go files in the rules package: [%s]; port the rule, rename the file, or add the helper to registryGuardHelperFiles",
			strings.Join(unexpected, ", "))
	}
}
