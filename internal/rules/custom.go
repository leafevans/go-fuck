package rules

import (
	"errors"
	"fmt"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

// RegisterCustom adds a rule loaded from a user rule file.
func RegisterCustom(rule *types.Rule) error {
	if rule == nil {
		return errors.New("custom rule is nil")
	}
	if !addRule(rule) {
		return fmt.Errorf("rule %q is already registered", rule.Name)
	}
	return nil
}
