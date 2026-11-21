package types

import "reflect"

type correctedCommandKey struct {
	script     string
	sideEffect uintptr
}

// CorrectedCommandSet tracks command identity using the same semantics as
// CorrectedCommand.Equal. Priority and rule name do not distinguish commands.
// The zero value is ready to use.
type CorrectedCommandSet struct {
	first      map[string]uintptr
	additional map[correctedCommandKey]struct{}
}

// NewCorrectedCommandSet returns a set sized for the expected number of commands.
func NewCorrectedCommandSet(size int) CorrectedCommandSet {
	return CorrectedCommandSet{first: make(map[string]uintptr, size)}
}

// Add records command and reports whether it was not already in the set.
func (s *CorrectedCommandSet) Add(command *CorrectedCommand) bool {
	if s.first == nil {
		s.first = make(map[string]uintptr)
	}
	effect := sideEffectPointer(command.SideEffect)
	previous, exists := s.first[command.Script]
	if !exists {
		s.first[command.Script] = effect
		return true
	}
	if previous == effect {
		return false
	}
	// Most scripts occur once. Allocate a second index only when distinct
	// side effects share a script, without scanning the preceding commands.
	if s.additional == nil {
		s.additional = make(map[correctedCommandKey]struct{})
	}
	key := correctedCommandKey{command.Script, effect}
	if _, exists := s.additional[key]; exists {
		return false
	}
	s.additional[key] = struct{}{}
	return true
}

// sideEffectPointer mirrors the function identity used by Equal. Function code
// pointers do not distinguish closures that share code and capture other data.
func sideEffectPointer(effect SideEffectFunc) uintptr {
	if effect == nil {
		return 0
	}
	return reflect.ValueOf(effect).Pointer()
}
