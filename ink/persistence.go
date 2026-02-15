package ink

import (
	"encoding/json"
	"fmt"
)

// ToJSON serializes the story state to a JSON string.
func (s *Story) ToJSON() (string, error) {
	dto := s.stateToDto()

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

func (s *Story) stateToDto() *StoryStateDto {
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
		flowDto := flowToDto(flow)
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
		vDto := runtimeObjectToInterface(val)
		dto.VariablesState[name] = vDto
	}

	// EvalStack
	evalStackDto := runtimeListToDto(ss.EvaluationStack)
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

	return dto
}

func flowToDto(flow *Flow) FlowDto {
	dto := FlowDto{
		OutputStream:  make([]interface{}, 0),
		ChoiceThreads: make(map[string]CallStackThreadDto),
	}

	// CallStack
	csDto := callStackToDto(flow.CallStack)
	dto.CallStack = csDto

	// OutputStream
	streamDto := runtimeListToDto(flow.OutputStream)
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
			tDto := threadToDto(c.ThreadAtGeneration)
			dto.ChoiceThreads[fmt.Sprintf("%d", c.OriginalThreadIndex)] = tDto
		}
	}

	// CurrentChoices
	dto.CurrentChoices = make([]ChoiceDto, len(flow.CurrentChoices))
	for i, c := range flow.CurrentChoices {
		dto.CurrentChoices[i] = choiceToDto(c)
	}

	return dto
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

func callStackToDto(cs *CallStack) CallStackDto {
	dto := CallStackDto{
		Threads:       make([]CallStackThreadDto, 0),
		ThreadCounter: cs.ThreadCounter,
	}

	for _, t := range cs.Threads {
		tDto := threadToDto(t)
		dto.Threads = append(dto.Threads, tDto)
	}

	return dto
}

func threadToDto(t *CallStackThread) CallStackThreadDto {
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
			vDto := runtimeObjectToInterface(val)
			elDto.TemporaryVariables[name] = vDto
		}

		dto.CallStack = append(dto.CallStack, elDto)
	}

	if !t.PreviousPointer.IsNull() {
		// Java: previousPointer.resolve().getPath().toString()
		// We resolve the pointer to get the target object, then get its path.
		resolved := t.PreviousPointer.Resolve()
		if resolved != nil && resolved.GetPath() != nil {
			dto.PreviousContentObject = resolved.GetPath().String()
		}
	}

	return dto
}

func runtimeListToDto(list []RuntimeObject) []interface{} {
	res := make([]interface{}, len(list))
	for i, obj := range list {
		v := runtimeObjectToInterface(obj)
		res[i] = v
	}
	return res
}

func runtimeObjectToInterface(obj RuntimeObject) interface{} {
	switch v := obj.(type) {
	case *Container:
		return nil // Containers not typically serialized as values here
	case *BoolValue:
		return v.Value
	case *IntValue:
		return v.Value
	case *FloatValue:
		return v.Value
	case *StringValue:
		if v.GetIsNewline() {
			return "\n"
		}
		return "^" + v.Value
	case *ListValue:
		return inkListToDto(v)
	case *DivertTargetValue:
		pathStr := ""
		if v.TargetPath != nil {
			pathStr = v.TargetPath.String()
		}
		return map[string]string{"^->": pathStr}
	case *VariablePointerValue:
		return map[string]interface{}{
			"^var": v.VariableName(),
			"ci":   v.ContextIndex(),
		}
	case *Glue:
		return "<>"
	case *ControlCommand:
		return controlCommandToString(v.CommandType)
	case *NativeFunctionCall:
		return v.Name
	case *Void:
		return VoidName
	case *ChoicePoint:
		return map[string]interface{}{
			"*":   v.PathStringOnChoice,
			"flg": v.Flags(),
		}
	}
	// Fallback for unknown types or nil
	return nil
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
