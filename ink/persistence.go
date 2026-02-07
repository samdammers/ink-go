package ink

import (
	"encoding/json"
	"fmt"
)

// ToJSON serializes the story state to a JSON string.
func (s *Story) ToJSON() (string, error) {
	dto, err := s.stateToDto()
	if err != nil {
		return "", err
	}

	// Marshal to JSON
	// We use standard Marshal. For Indent, user can unmarshal and marshal again if needed.
	// Or we can use MarshalIndent if readable output is preferred (often the case for Save Games).
	// But let's stick to standard compact or maybe indent?
	// Ink code often produces non-indented JSON for size, but for debugging indentation is nice.
	// Java implementation uses SimpleJson which defaults to compact unless configured?
	// Let's use compact for now.
	bytes, err := json.Marshal(dto)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Story) stateToDto() (*StoryStateDto, error) {
	ss := s.state

	dto := &StoryStateDto{
		InkSaveVersion:   10, // kInkSaveStateVersion
		InkFormatVersion: 21, // inkVersionCurrent
		Flows:            make(map[string]FlowDto),
		VariablesState:   make(map[string]interface{}),
		VisitCounts:      make(map[string]int),
		TurnIndices:      make(map[string]int),
	}

	// Flows
	for name, flow := range ss.NamedFlows {
		flowDto, err := flowToDto(flow)
		if err != nil {
			return nil, err
		}
		dto.Flows[name] = flowDto
	}

	// Current Flow Name
	if ss.CurrentFlow != nil {
		dto.CurrentFlowName = ss.CurrentFlow.Name
	} else {
		dto.CurrentFlowName = "DEFAULT_FLOW"
	}

	// VariablesState
	// Note: We might want option to skip default values (optimization), but for now we save all.
	for name, val := range ss.VariablesState.GlobalVariables {
		// Optimization: Check if value equals default global value?
		// We skip this optimization for simplicity for now as per instructions (Map -> Dto).
		vDto, err := runtimeObjectToInterface(val)
		if err != nil {
			return nil, err
		}
		dto.VariablesState[name] = vDto
	}

	// EvalStack
	evalStackDto, err := runtimeListToDto(ss.EvaluationStack)
	if err != nil {
		return nil, err
	}
	dto.EvalStack = evalStackDto

	// DivertedPointer
	if !ss.DivertedPointer.IsNull() {
		dto.CurrentDivertTarget = ss.DivertedPointer.Path().String()
	}

	// VisitCounts
	for container, count := range ss.VisitCounts {
		if container != nil && container.GetPath() != nil {
			dto.VisitCounts[container.GetPath().String()] = count
		}
	}

	// TurnIndices
	for container, idx := range ss.TurnIndices {
		if container != nil && container.GetPath() != nil {
			dto.TurnIndices[container.GetPath().String()] = idx
		}
	}

	dto.TurnIdx = ss.CurrentTurnIndex
	dto.StorySeed = ss.StorySeed
	dto.PreviousRandom = ss.PreviousRandom

	return dto, nil
}

func flowToDto(flow *Flow) (FlowDto, error) {
	dto := FlowDto{
		OutputStream:  make([]interface{}, 0),
		ChoiceThreads: make(map[string]CallStackThreadDto),
	}

	// CallStack
	csDto, err := callStackToDto(flow.CallStack)
	if err != nil {
		return dto, err
	}
	dto.CallStack = csDto

	// OutputStream
	streamDto, err := runtimeListToDto(flow.OutputStream)
	if err != nil {
		return dto, err
	}
	dto.OutputStream = streamDto

	// ChoiceThreads
	// Logic: "Has to come BEFORE the choices themselves are written out since the originalThreadIndex of each choice needs to be set"
	// In the DTO model, we just save the map directly.
	// We iterate through flow.CurrentChoices.
	// Java logic:
	// for (Choice c : currentChoices) {
	//    c.originalThreadIndex = c.getThreadAtGeneration().threadIndex;
	//    if (callStack.getThreadWithIndex(c.originalThreadIndex) == null) {
	//        if (!hasChoiceThreads) ...
	//        writer.writePropertyStart(c.originalThreadIndex);
	//        c.getThreadAtGeneration().writeJson(writer);
	//    }
	// }
	// The key is c.originalThreadIndex (int). Go map keys are strings for JSON usually, but DTO defined as map[string]CallStackThreadDto.
	// So we convert int to string.

	for _, c := range flow.CurrentChoices {
		// Update originalThreadIndex based on ThreadAtGeneration
		if c.ThreadAtGeneration != nil {
			c.OriginalThreadIndex = c.ThreadAtGeneration.ThreadIndex
		}

		// Check if thread is NOT in callstack
		found := false
		for _, t := range flow.CallStack.Threads {
			if t.ThreadIndex == c.OriginalThreadIndex {
				found = true
				break
			}
		}

		if !found && c.ThreadAtGeneration != nil {
			tDto, err := threadToDto(c.ThreadAtGeneration)
			if err != nil {
				return dto, err
			}
			dto.ChoiceThreads[fmt.Sprintf("%d", c.OriginalThreadIndex)] = tDto
		}
	}

	// CurrentChoices
	dto.CurrentChoices = make([]ChoiceDto, len(flow.CurrentChoices))
	for i, c := range flow.CurrentChoices {
		dto.CurrentChoices[i] = choiceToDto(c)
	}

	return dto, nil
}

func choiceToDto(c *Choice) ChoiceDto {
	dto := ChoiceDto{
		Text:                c.Text,
		Index:               c.Index,
		OriginalChoicePath:  c.SourcePath,
		OriginalThreadIndex: c.OriginalThreadIndex,
		Tags:                c.Tags,
	}

	if c.TargetPath != nil {
		dto.TargetPath = c.TargetPath.String()
	}

	return dto
}

func callStackToDto(cs *CallStack) (CallStackDto, error) {
	dto := CallStackDto{
		Threads:       make([]CallStackThreadDto, 0),
		ThreadCounter: cs.ThreadCounter,
	}

	for _, t := range cs.Threads {
		tDto, err := threadToDto(t)
		if err != nil {
			return dto, err
		}
		dto.Threads = append(dto.Threads, tDto)
	}

	return dto, nil
}

func threadToDto(t *CallStackThread) (CallStackThreadDto, error) {
	dto := CallStackThreadDto{
		CallStack:   make([]CallStackElementDto, 0),
		ThreadIndex: t.ThreadIndex,
	}

	for _, el := range t.CallStack {
		elDto := CallStackElementDto{
			Idx:                el.CurrentPointer.Index,
			Exp:                el.InExpressionEvaluation,
			Type:               int(el.Type),
			TemporaryVariables: make(map[string]interface{}),
		}

		if !el.CurrentPointer.IsNull() {
			if el.CurrentPointer.Container != nil && el.CurrentPointer.Container.GetPath() != nil {
				elDto.CPath = el.CurrentPointer.Container.GetPath().String()
			}
		}

		for name, val := range el.TemporaryVariables {
			vDto, err := runtimeObjectToInterface(val)
			if err != nil {
				return dto, err
			}
			elDto.TemporaryVariables[name] = vDto
		}

		dto.CallStack = append(dto.CallStack, elDto)
	}

	if !t.PreviousPointer.IsNull() {
		// Java: previousPointer.resolve().getPath().toString()
		// Wait, PreviousPointer might be pointing to something specific?
		// Usually resolve() handles index.
		// If pointer is standard (Container + Index), we want the path to the Object pointed to.
		// If index is -1, it's the container's path.
		// If index >= 0, it's container.path + index.
		// Pointer.Resolved() returns RuntimeObject. Object.Path.String().
		resolved := t.PreviousPointer.Resolve()
		if resolved != nil && resolved.GetPath() != nil {
			dto.PreviousContentObject = resolved.GetPath().String()
		}
	}

	return dto, nil
}

func runtimeListToDto(list []RuntimeObject) ([]interface{}, error) {
	res := make([]interface{}, len(list))
	for i, obj := range list {
		v, err := runtimeObjectToInterface(obj)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}

func runtimeObjectToInterface(obj RuntimeObject) (interface{}, error) {
	switch v := obj.(type) {
	case *Container:
		return nil, nil // Containers not typically serialized as values here
	case *BoolValue:
		return v.Value, nil
	case *IntValue:
		return v.Value, nil
	case *FloatValue:
		return v.Value, nil
	case *StringValue:
		if v.GetIsNewline() {
			return "\n", nil
		}
		return "^" + v.Value, nil
	case *ListValue:
		return inkListToDto(v), nil
	case *DivertTargetValue:
		pathStr := ""
		if v.TargetPath != nil {
			pathStr = v.TargetPath.String()
		}
		return map[string]string{"^->": pathStr}, nil
	case *VariablePointerValue:
		return map[string]interface{}{
			"^var": v.VariableName(),
			"ci":   v.ContextIndex(),
		}, nil
	case *Glue:
		return "<>", nil
	case *ControlCommand:
		return controlCommandToString(v.CommandType), nil
	case *NativeFunctionCall:
		return v.Name, nil
	case *Void:
		return VoidName, nil
	case *ChoicePoint:
		return map[string]interface{}{
			"*":   v.PathStringOnChoice,
			"flg": v.Flags(),
		}, nil
	}
	// Fallback for unknown types or nil
	return nil, nil
}

func inkListToDto(l *ListValue) map[string]interface{} {
	rawList := l.Value // List
	listData := make(map[string]int)

	for item, val := range rawList.Items { // Map[ListItem]int
		// key: OriginName.ItemName
		key := "?"
		if item.OriginName != "" {
			key = item.OriginName
		}
		key += "." + item.ItemName
		listData[key] = val
	}

	res := map[string]interface{}{
		"list": listData,
	}

	// Origins if list is empty
	if len(rawList.Items) == 0 && len(rawList.Origins) > 0 {
		names := make([]string, len(rawList.Origins))
		for i, o := range rawList.Origins {
			names[i] = o.Name
		}
		res["origins"] = names
	}

	return res
}

func controlCommandToString(cType CommandType) string {
	if int(cType) >= 1 && int(cType) < len(controlCommandNames) {
		return controlCommandNames[int(cType)]
	}
	return "nop"
}

// controlCommandNames mapping mirroring Java implementation
var controlCommandNames = [...]string{
	"",          // 0: NotSet maps to index -1 in java which is invalid access
	"ev",        // 1: EvalStart
	"out",       // 2: EvalOutput
	"/ev",       // 3: EvalEnd
	"du",        // 4: Duplicate
	"pop",       // 5: PopEvaluatedValue
	"~ret",      // 6: PopFunction
	"->->",      // 7: PopTunnel
	"str",       // 8: BeginString
	"/str",      // 9: EndString
	"nop",       // 10: NoOp
	"choiceCnt", // 11: ChoiceCount
	"turn",      // 12: Turns
	"turns",     // 13: TurnsSince
	"readc",     // 14: ReadCount
	"rnd",       // 15: Random
	"srnd",      // 16: SeedRandom
	"visit",     // 17: VisitIndex
	"seq",       // 18: SequenceShuffleIndex
	"thread",    // 19: StartThread
	"done",      // 20: Done
	"end",       // 21: End
	"listInt",   // 22: ListFromInt
	"range",     // 23: ListRange
	"lrnd",      // 24: ListRandom
	"#",         // 25: BeginTag
	"/#",        // 26: EndTag
}
