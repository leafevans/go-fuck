package rules

import "testing"

// terraformInitInitializationOutput is the error printed when terraform has not
// been initialized.
const terraformInitInitializationOutput = "Error: Initialization required. Please see the error message above."

// terraformInitModuleOutput is the error printed when the modules are missing.
const terraformInitModuleOutput = "This module is not yet installed. Run \"terraform init\" to install all modules required by this configuration."

func TestRuleTerraformInit(t *testing.T) {
	resetTestState(t)
	rule := Get("terraform_init")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "terraform plan", terraformInitInitializationOutput)
		assertMatch(t, rule, "terraform plan", terraformInitModuleOutput)
		assertMatch(t, rule, "terraform apply", terraformInitInitializationOutput)
		assertMatch(t, rule, "terraform apply", terraformInitModuleOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "terraform --version", "Terraform v0.12.2")
		assertNotMatch(t, rule, "terraform plan", "No changes. Infrastructure is up-to-date.")
		assertNotMatch(t, rule, "terraform apply", "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "terraform plan", "",
			[]string{"terraform init && terraform plan"})
		assertNewCommands(t, rule, "terraform apply", "",
			[]string{"terraform init && terraform apply"})
	})
}
