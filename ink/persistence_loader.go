package ink

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// LoadState loads the story state from a JSON string.
func (s *Story) LoadState(jsonStr string) error {
	var dto StoryStateDto
	err := json.Unmarshal([]byte(jsonStr), &dto)
	if err != nil {
		return err
	}

	return s.restoreStoryState(&dto)
}

//nolint:gocognit
func (s *Story) restoreStoryState(dto *StoryStateDto) error {
	// Restore VariablesState
	// Note: We should probably reset global variables first or ensure we are overwriting.
	// The DTO contains the current state of variables.

	// Re-initialize VariablesState to ensure clean slate (though NewStoryState usually does this).
	// We iterate keys in DTO and set them.
	// But first, we need a working jTokenToRuntimeObject.

	s.state.VariablesState.GlobalVariables = make(map[string]RuntimeObject)
	for k, vDto := range dto.VariablesState {
		val, err := s.jTokenToRuntimeObject(vDto)
		if err != nil {
			return fmt.Errorf("failed to load variable '%s': %w", k, err)
		}
		s.state.VariablesState.GlobalVariables[k] = val
	}

	// Restore EvalStack
	s.state.EvaluationStack = make([]RuntimeObject, len(dto.EvalStack))
	for i, vDto := range dto.EvalStack {
		val, err := s.jTokenToRuntimeObject(vDto)
		if err != nil {
			return fmt.Errorf("failed to load eval stack item at index %d: %w", i, err)
		}
		s.state.EvaluationStack[i] = val
	}

	// Restore DivertedPointer
	if dto.CurrentDivertTarget != "" {
		p := s.PointerAtPath(NewPathFromString(dto.CurrentDivertTarget))
		// Warning or error if null? Usually implies corrupted path or changed story.
		// For robustness, we accept it might be null if story changed, but ideally valid.
		s.state.DivertedPointer = p
	} else {
		s.state.DivertedPointer = NullPointer
	}

	// Restore VisitCounts
	s.state.VisitCounts = make(map[*Container]int)
	for k, v := range dto.VisitCounts {
		p := s.PointerAtPath(NewPathFromString(k))
		if !p.IsNull() {
			if c, ok := p.Resolve().(*Container); ok {
				s.state.VisitCounts[c] = v
			}
		}
	}

	// Restore TurnIndices
	s.state.TurnIndices = make(map[*Container]int)
	for k, v := range dto.TurnIndices {
		p := s.PointerAtPath(NewPathFromString(k))
		if !p.IsNull() {
			if c, ok := p.Resolve().(*Container); ok {
				s.state.TurnIndices[c] = v
			}
		}
	}

	s.state.CurrentTurnIndex = dto.TurnIdx
	s.state.StorySeed = dto.StorySeed
	s.state.PreviousRandom = dto.PreviousRandom

	// Restore Flows
	s.state.NamedFlows = make(map[string]*Flow)
	for name, flowDto := range dto.Flows {
		flow, err := s.restoreFlow(&flowDto, name)
		if err != nil {
			return fmt.Errorf("failed to restore flow '%s': %w", name, err)
		}
		s.state.NamedFlows[name] = flow
	}

	// Set Current Flow
	if currFlow, ok := s.state.NamedFlows[dto.CurrentFlowName]; ok {
		s.state.CurrentFlow = currFlow
		s.state.CurrentChoices = currFlow.CurrentChoices
	} else {
		// Default fallback if not found? Should generally exist.
		// If explicit "DEFAULT_FLOW" is missing, we might need to create it?
		// Usually DTO contains it.
		return fmt.Errorf("current flow '%s' not found in saved flows", dto.CurrentFlowName)
	}

	s.state.OutputStreamDirty = true
	return nil
}

func (s *Story) restoreFlow(dto *FlowDto, name string) (*Flow, error) {
	flow := NewFlow(name, s)

	// Restore CallStack
	cs, err := s.restoreCallStack(&dto.CallStack)
	if err != nil {
		return nil, err
	}
	flow.CallStack = cs

	// Restore OutputStream
	flow.OutputStream = make([]RuntimeObject, len(dto.OutputStream))
	for i, vDto := range dto.OutputStream {
		val, err := s.jTokenToRuntimeObject(vDto)
		if err != nil {
			return nil, fmt.Errorf("failed to output stream item at index %d: %w", i, err)
		}
		flow.OutputStream[i] = val
	}

	// Restore Choices
	flow.CurrentChoices = make([]*Choice, len(dto.CurrentChoices))
	for i, cDto := range dto.CurrentChoices {
		c := s.restoreChoice(&cDto)

		// Map thread if exists in ChoiceThreads
		// ChoiceThreads key is originalThreadIndex (int) as string
		// We need to restore the thread and assign it to c.ThreadAtGeneration.

		// Optimisation: choiceThreads map in DTO holds the thread data.
		// We just need to parse it and create the CallStackThread.
		// Note from Java: "Has to come BEFORE the choices themselves are written out"
		// In DTO we have them in parallel.

		if threadDto, ok := dto.ChoiceThreads[fmt.Sprintf("%d", c.OriginalThreadIndex)]; ok {
			threadDto := threadDto // Capture loop variable or map lookups to avoid aliasing issues
			thread, err := s.restoreThread(&threadDto)
			if err != nil {
				return nil, fmt.Errorf("failed to restore choice thread %d: %w", c.OriginalThreadIndex, err)
			}
			c.ThreadAtGeneration = thread
		} else {
			// If not found in choiceThreads, it might be in the main CallStack
			// Java logic: "if (callStack.getThreadWithIndex(c.originalThreadIndex) == null)" -> write it.
			// So if it IS in the callstack, we should find it there.
			for _, t := range flow.CallStack.Threads {
				if t.ThreadIndex == c.OriginalThreadIndex {
					c.ThreadAtGeneration = t.Copy() // Java: choice.setThreadAtGeneration(foundActiveThread.copy());
					break
				}
			}
		}

		flow.CurrentChoices[i] = c
	}

	return flow, nil
}

func (s *Story) restoreChoice(dto *ChoiceDto) *Choice {
	c := NewChoice()
	c.Text = dto.Text
	c.Index = dto.Index
	c.SourcePath = dto.OriginalChoicePath
	c.OriginalThreadIndex = dto.OriginalThreadIndex
	c.Tags = dto.Tags
	c.IsInvisibleDefault = false // Default to false as it is not persisted in the DTO

	if dto.TargetPath != "" {
		c.TargetPath = NewPathFromString(dto.TargetPath)
	}

	return c
}

func (s *Story) restoreCallStack(dto *CallStackDto) (*CallStack, error) {
	cs := NewCallStack(s.MainContent) // Start with default
	cs.Threads = make([]*CallStackThread, len(dto.Threads))
	cs.ThreadCounter = dto.ThreadCounter

	for i, tDto := range dto.Threads {
		t, err := s.restoreThread(&tDto)
		if err != nil {
			return nil, err
		}
		cs.Threads[i] = t
	}

	return cs, nil
}

func (s *Story) restoreThread(dto *CallStackThreadDto) (*CallStackThread, error) {
	t := NewCallStackThread()
	t.ThreadIndex = dto.ThreadIndex

	if dto.PreviousContentObject != "" {
		p := s.PointerAtPath(NewPathFromString(dto.PreviousContentObject))
		if !p.IsNull() {
			t.PreviousPointer = p
		}
	}

	t.CallStack = make([]*CallStackElement, len(dto.CallStack))
	for i, elDto := range dto.CallStack {
		el, err := s.restoreElement(&elDto)
		if err != nil {
			return nil, err
		}
		t.CallStack[i] = el
	}

	return t, nil
}

func (s *Story) restoreElement(dto *CallStackElementDto) (*CallStackElement, error) {
	// Reconstruct pointer
	var p Pointer
	if dto.CPath != "" {
		path := NewPathFromString(dto.CPath)
		p = s.PointerAtPath(path)
		p.Index = dto.Idx

		if p.IsNull() {
			// This might happen if CPath is empty string (root?) handled above.
			// Or if pointer resolution failed.
			// "root" path is empty components.
			if dto.CPath == "" && dto.Idx == 0 {
				p.Container = s.MainContent
				p.Index = 0
			}
		}
	} else {
		p = NullPointer
	}

	el := NewCallStackElement(PushPopType(dto.Type), p, dto.Exp)

	// Restore temps
	for k, vDto := range dto.TemporaryVariables {
		val, err := s.jTokenToRuntimeObject(vDto)
		if err != nil {
			return nil, fmt.Errorf("failed to restore temp var '%s': %w", k, err)
		}
		el.TemporaryVariables[k] = val
	}

	return el, nil
}

func (s *Story) jTokenToRuntimeObject(token interface{}) (RuntimeObject, error) {
	if token == nil {
		return nil, nil
	}

	switch val := token.(type) {
	case string:
		return s.parseStateString(val)

	case bool:
		return NewBoolValue(val), nil

	case float64:
		// Check if int
		if val == math.Trunc(val) {
			return NewIntValue(int(val)), nil
		}
		return NewFloatValue(val), nil

	case int: // Just in case, though json decodes to float64 usually
		return NewIntValue(val), nil

	case map[string]interface{}:
		if target, ok := val["^->"]; ok {
			return NewDivertTargetValue(NewPathFromString(target.(string))), nil
		}

		if name, ok := val["^var"]; ok {
			ci := -1
			if ciVal, ok := val["ci"]; ok {
				// ciVal is float64 likely
				if f, ok := ciVal.(float64); ok {
					ci = int(f)
				}
			}
			return NewVariablePointerValue(name.(string), ci), nil
		}

		if obj, err := s.parseStateList(val); obj != nil || err != nil {
			return obj, err
		}

		if obj := s.parseStateChoice(val); obj != nil {
			return obj, nil
		}
	}

	return nil, fmt.Errorf("unknown token type: %T", token)
}

func (s *Story) parseStateList(val map[string]interface{}) (RuntimeObject, error) {
	listData, ok := val["list"]
	if !ok {
		return nil, nil
	}

	inkList := NewList()
	listMap, ok := listData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid list data format")
	}

	for key, v := range listMap {
		s.parseStateListItem(inkList, key, v)
	}

	if originsVal, ok := val["origins"]; ok {
		if originsList, ok := originsVal.([]interface{}); ok {
			for _, o := range originsList {
				if name, ok := o.(string); ok {
					if def, ok := s.ListDefinitions.Lists[name]; ok {
						inkList.Origins = append(inkList.Origins, def)
					}
				}
			}
		}
	}

	return NewListValue(inkList), nil
}

func (s *Story) parseStateChoice(val map[string]interface{}) RuntimeObject {
	if pathOnChoice, ok := val["*"]; ok {
		path := pathOnChoice.(string)
		flg := 0
		if f, ok := val["flg"]; ok {
			flg = int(f.(float64))
		}
		cp := NewChoicePoint(false, false, false, false, false)
		cp.SetPathStringOnChoice(path)
		cp.SetFlags(flg)
		return cp
	}
	if pathOnChoice, ok := val["+"]; ok {
		path := pathOnChoice.(string)
		flg := 0
		if f, ok := val["flg"]; ok {
			flg = int(f.(float64))
		}
		cp := NewChoicePoint(false, false, false, false, false)
		cp.SetPathStringOnChoice(path)
		cp.SetFlags(flg)
		return cp
	}
	return nil
}

func (s *Story) parseStateString(val string) (RuntimeObject, error) {
	firstChar := ""
	if len(val) > 0 {
		firstChar = string(val[0])
	}

	if firstChar == "^" {
		return NewStringValue(val[1:]), nil
	}
	if val == "\n" {
		return NewStringValue("\n"), nil
	}
	if val == "<>" {
		return NewGlue(), nil
	}

	for i, name := range controlCommandNames {
		if name == val {
			return NewControlCommand(CommandType(i)), nil
		}
	}

	if val == "L^" {
		val = "^"
	}
	return NewNativeFunctionCall(val), nil
}

func (s *Story) parseStateListItem(inkList *List, key string, v interface{}) {
	itemValFloat, ok := v.(float64)
	if !ok {
		return
	}
	itemVal := int(itemValFloat)

	parts := strings.Split(key, ".")
	var originName, itemName string
	if len(parts) == 2 {
		originName = parts[0]
		itemName = parts[1]
	} else {
		itemName = key
	}

	item := NewListItem(originName, itemName)

	if s.ListDefinitions != nil {
		if def, ok := s.ListDefinitions.Lists[originName]; ok {
			if _, ok := def.Items[itemName]; ok {
				item.OriginName = def.Name
			}
		}
	}

	inkList.Add(item, itemVal)
}
