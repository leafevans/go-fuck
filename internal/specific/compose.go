package specific

import (
	"os"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// composeFileNames are the Compose file names that docker compose searches, in
// order.
var composeFileNames = []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}

// ComposeServices returns the service names of the Compose file in the current
// working directory, sorted by name.
//
// Only the first Compose file that exists is parsed and only the simple
// "services:" block is understood; YAML anchors and included files are not
// resolved. The result is memoized because the file does not change while
// go-fuck runs.
var ComposeServices = utils.Memoize(func() []string {
	for _, name := range composeFileNames {
		content, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		return parseComposeServices(string(content))
	}
	return nil
})

// parseComposeServices returns the service names of a Compose file content.
func parseComposeServices(content string) []string {
	services := []string{}
	inServices := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if !inServices {
			if indent == 0 && strings.HasPrefix(trimmed, "services:") {
				inServices = true
			}
			continue
		}
		if indent == 0 {
			break
		}
		if indent != 2 {
			continue
		}
		name, _, found := strings.Cut(trimmed, ":")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" || strings.ContainsAny(name, " \t'\"$#{}") || strings.HasPrefix(name, "-") {
			continue
		}
		services = append(services, name)
	}
	sort.Strings(services)
	return services
}
