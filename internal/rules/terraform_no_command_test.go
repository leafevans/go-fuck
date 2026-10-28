package rules

import "testing"

func TestRuleTerraformNoCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("terraform_no_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "terraform appyl",
			`Terraform has no command named "appyl". Did you mean "apply"?`)
		assertMatch(t, rule, "terraform destory",
			`Terraform has no command named "destory". Did you mean "destroy"?`)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "terraform --version", "Terraform v0.12.2")
		assertNotMatch(t, rule, "terraform plan", "No changes. Infrastructure is up-to-date.")
		assertNotMatch(t, rule, "terraform apply", "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "terraform appyl",
			`Terraform has no command named "appyl". Did you mean "apply"?`,
			[]string{"terraform apply"})
		assertNewCommands(t, rule, "terraform destory --some-other-option",
			`Terraform has no command named "destory". Did you mean "destroy"?`,
			[]string{"terraform destroy --some-other-option"})
	})
}
