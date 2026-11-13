package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cargoAvailableTargets returns the target names that cargo lists after the
// given section marker.
func cargoAvailableTargets(output, marker string) []string {
	index := strings.Index(output, marker)
	if index < 0 {
		return nil
	}
	var targets []string
	for _, line := range strings.Split(output[index+len(marker):], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, "    ") {
			break
		}
		targets = append(targets, trimmed)
	}
	return targets
}

// cargoClosestTarget returns the closest target to missing, or the only
// target when there is exactly one.
func cargoClosestTarget(missing string, targets []string) string {
	target := utils.GetClosest(missing, targets, 0.6, false)
	if target == "" && len(targets) == 1 {
		target = targets[0]
	}
	return target
}
