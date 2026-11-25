package rules

import "testing"

func TestRuleGcloudCLI(t *testing.T) {
	resetTestState(t)
	rule := Get("gcloud_cli")
	// Diagnostic from https://github.com/nvbn/thefuck/pull/1082.
	output := "ERROR: (gcloud.compute) Invalid choice: 'instance'.\nMaybe you meant:\n  gcloud compute instance-groups\n  gcloud compute instance-templates\n  gcloud compute instances\n  gcloud compute target-instances\n\nTo search the help text of gcloud commands, run:\n  gcloud help -- SEARCH_TERMS\n"
	command := "gcloud compute instance list --project=demo"
	assertMatch(t, rule, command, output)
	assertNewCommands(t, rule, command, output, []string{
		"gcloud compute instance-groups list --project=demo",
		"gcloud compute instance-templates list --project=demo",
		"gcloud compute instances list --project=demo",
		"gcloud compute target-instances list --project=demo",
	})
	for _, command := range []string{"aws compute instance list", "gcloud comute instance list", "gcloud compute list instance", "gcloud compute instance list; echo done"} {
		assertNotMatch(t, rule, command, output)
	}
	assertNotMatch(t, rule, command, "ERROR: (gcloud.compute) Invalid choice: 'instance'.\nUsage: gcloud compute")
	assertNotMatch(t, rule, command, "ERROR: (gcloud.compute) Invalid choice: 'instance'.\nMaybe you meant:\n  gcloud config set\n  gcloud compute instances;\n  gcloud compute instances delete")
	assertNewCommands(t, rule, "gcloud cmp instances list",
		"ERROR: (gcloud) Invalid choice: 'cmp'.\r\nMaybe you meant:\r\n  gcloud compute\r\n  gcloud compute\r\n",
		[]string{"gcloud compute instances list"})
}
