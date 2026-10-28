package rules

import "testing"

// dockerImageBeingUsedByContainerError is the docker error message used by the
// Python tests.
const dockerImageBeingUsedByContainerError = "Error response from daemon: conflict: unable to delete cd809b04b6ff (cannot be forced) - image is being used by running container e5e2591040d1"

func TestRuleDockerImageBeingUsedByContainer(t *testing.T) {
	resetTestState(t)
	rule := Get("docker_image_being_used_by_container")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "docker image rm -f cd809b04b6ff", dockerImageBeingUsedByContainerError)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "docker image rm -f cd809b04b6ff", "bash: docker: command not found")
	})

	t.Run("not docker command", func(t *testing.T) {
		assertNotMatch(t, rule, "git image rm -f cd809b04b6ff", dockerImageBeingUsedByContainerError)
	})

	t.Run("get new command", func(t *testing.T) {
		// The container id is the last word of the (multi line) output.
		output := "\n        Error response from daemon: conflict: unable to delete cd809b04b6ff (cannot be forced) - image\n        is being used by running container e5e2591040d1\n        "
		assertNewCommands(t, rule, "docker image rm -f cd809b04b6ff", output,
			[]string{"docker container rm -f e5e2591040d1 && docker image rm -f cd809b04b6ff"})
	})
}
