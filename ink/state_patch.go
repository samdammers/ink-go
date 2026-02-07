package ink

// StatePatch is used to apply changes to the state.
type StatePatch struct {
	Globals          map[string]RuntimeObject
	ChangedVariables map[string]struct{} // Set
	VisitCounts      map[*Container]int
	TurnIndices      map[*Container]int
}

// NewStatePatch creates a new StatePatch.
func NewStatePatch(toCopy *StatePatch) *StatePatch {
	sp := &StatePatch{
		Globals:          make(map[string]RuntimeObject),
		ChangedVariables: make(map[string]struct{}),
		VisitCounts:      make(map[*Container]int),
		TurnIndices:      make(map[*Container]int),
	}

	if toCopy != nil {
		for k, v := range toCopy.Globals {
			sp.Globals[k] = v
		}
		for k := range toCopy.ChangedVariables {
			sp.ChangedVariables[k] = struct{}{}
		}
		for k, v := range toCopy.VisitCounts {
			sp.VisitCounts[k] = v // Pointer keys, but container pointers sort of persistent
		}
		for k, v := range toCopy.TurnIndices {
			sp.TurnIndices[k] = v
		}
	}
	return sp
}

// GetGlobals returns the global variables in the patch.
func (sp *StatePatch) GetGlobals() map[string]RuntimeObject {
	return sp.Globals
}

// GetChangedVariables returns the names of the changed variables.
func (sp *StatePatch) GetChangedVariables() []string {
	keys := make([]string, 0, len(sp.ChangedVariables))
	for k := range sp.ChangedVariables {
		keys = append(keys, k)
	}
	return keys
}
