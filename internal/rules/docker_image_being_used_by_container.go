package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dockerImageBeingUsedByContainerMatch reports whether docker refused to remove
// an image because a running container still uses it.
func dockerImageBeingUsedByContainerMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "image is being used by running container")
}

// dockerImageBeingUsedByContainerNewCommand removes the container before the
// image removal is retried.
func dockerImageBeingUsedByContainerNewCommand(command *types.Command) ([]string, error) {
	// Python: command.output.strip().split(' ')[-1] -- the container id is the
	// last word of the error message.
	fields := strings.Fields(command.Output)
	if len(fields) == 0 {
		return nil, nil
	}
	containerID := fields[len(fields)-1]
	script := shells.Current().And("docker container rm -f "+containerID, command.Script)
	return []string{script}, nil
}

func init() {
	register(types.NewRule("docker_image_being_used_by_container",
		utils.ForApp([]string{"docker"}, 0, dockerImageBeingUsedByContainerMatch),
		dockerImageBeingUsedByContainerNewCommand))
}
