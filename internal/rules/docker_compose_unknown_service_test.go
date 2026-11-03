package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/specific"
)

// composeUnknownServiceOutput returns the output of docker compose for a
// service that is not defined in the Compose file.
func composeUnknownServiceOutput(service string) string {
	return "no such service: " + service + "\n"
}

// stubComposeServices replaces the Compose services with fixture values.
func stubComposeServices(t *testing.T) {
	t.Helper()
	original := specific.ComposeServices
	specific.ComposeServices = func() []string {
		return []string{"build", "db", "web"}
	}
	t.Cleanup(func() { specific.ComposeServices = original })
}

func TestRuleDockerComposeUnknownService(t *testing.T) {
	resetTestState(t)
	stubComposeServices(t)
	rule := Get("docker_compose_unknown_service")
	if rule == nil {
		t.Fatal("docker_compose_unknown_service rule is not registered")
	}

	assertMatch(t, rule, "docker compose up buidl", composeUnknownServiceOutput("buidl"))
	assertMatch(t, rule, "docker compose run buidl sh", composeUnknownServiceOutput("buidl"))
	assertNotMatch(t, rule, "docker run buidl", composeUnknownServiceOutput("buidl"))
	assertNotMatch(t, rule, "docker compose up web", "Creating web ... done\n")

	got, err := rule.GetNewCommand(testCommand("docker compose up buidl", composeUnknownServiceOutput("buidl")))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "docker compose up build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "docker compose up build")
	}
}

func TestRuleDockerComposeUnknownServiceWithoutServices(t *testing.T) {
	resetTestState(t)
	original := specific.ComposeServices
	specific.ComposeServices = func() []string { return nil }
	t.Cleanup(func() { specific.ComposeServices = original })

	rule := Get("docker_compose_unknown_service")
	assertNotMatch(t, rule, "docker compose up buidl", composeUnknownServiceOutput("buidl"))
}
