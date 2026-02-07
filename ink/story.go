package ink

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Story is the main entry point for the ink runtime.
type Story struct {
	MainContent       *Container
	state             *StoryState
	ListDefinitions   *ListDefinitionsOrigin
	externalFunctions map[string]ExternalFunction
}

// ExternalFunction represents a bound external function.
type ExternalFunction func(args []any) (any, error)

// NewStory creates a new Story object from a JSON string.
func NewStory(jsonString string) (*Story, error) {
	var root map[string]any
	if err := json.Unmarshal([]byte(jsonString), &root); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	// Wait, cannot reference story before declaration.
	// We allocated `story` variable after parsing.
	// Refactor:

	listDefsOrigin := NewListDefinitionsOrigin(nil)
	if listDefsToken, ok := root["listDefs"]; ok {
		if listDefsMap, ok := listDefsToken.(map[string]any); ok {
			defs := make([]*ListDefinition, 0)
			for name, itemsToken := range listDefsMap {
				itemsMap, ok := itemsToken.(map[string]any)
				if !ok {
					continue
				}
				items := make(map[string]int)
				for itemName, itemVal := range itemsMap {
					if val, ok := itemVal.(float64); ok {
						items[itemName] = int(val)
					}
				}
				defs = append(defs, NewListDefinition(name, items))
			}
			listDefsOrigin = NewListDefinitionsOrigin(defs)
		}
	}

	rootContainer, err := JObjectToRuntime(root)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root container: %w", err)
	}

	story := &Story{
		MainContent:       rootContainer,
		ListDefinitions:   listDefsOrigin,
		externalFunctions: make(map[string]ExternalFunction),
	}

	story.state = NewStoryState(story)

	story.ResetGlobals()

	return story, nil
}

// ResetGlobals runs the global declaration section to initialize variables.
func (s *Story) ResetGlobals() {
	if _, ok := s.MainContent.NamedContent["global decl"]; ok {
		s.ChoosePathString("global decl")
		s.ContinueMaximally()
	}
	s.state.GoToStart()
}

func (s *Story) GetListDefinitions() *ListDefinitionsOrigin {
	return s.ListDefinitions
}

// Continue continues the story evaluation.
func (s *Story) Continue() (string, error) {
	err := s.continueInternal(0)
	if err != nil {
		return "", err
	}
	return s.CurrentText(), nil
}

// ContinueMaximally continues the story until it stops (e.g. at a choice or end).
func (s *Story) ContinueMaximally() (string, error) {
	var sb strings.Builder
	for s.canContinueInternal() {
		text, err := s.Continue()
		if err != nil {
			return sb.String(), err
		}
		sb.WriteString(text)
	}
	return sb.String(), nil
}

// CurrentText returns the current output text.
// CurrentText returns the current output text.
func (s *Story) CurrentText() string {
	var sb strings.Builder
	glueActive := false

	for _, obj := range s.state.GetOutputStream() {
		var txt string
		isNewline := false
		isInlineWhitespace := false
		isGlue := false

		if _, ok := obj.(*Glue); ok {
			isGlue = true
		} else {
			switch v := obj.(type) {
			case *StringValue:
				txt = v.Value
				isNewline = v.isNewline
				isInlineWhitespace = v.isInlineWhitespace
			case *IntValue:
				txt = fmt.Sprintf("%d", v.Value)
			case *FloatValue:
				txt = fmt.Sprintf("%v", v.Value)
			case *ListValue:
				val, _ := v.Cast(ValueTypeString)
				if sv, ok := val.(*StringValue); ok {
					txt = sv.Value
				}
			case *BoolValue:
				txt = v.String()
			case *Void:
				continue
			default:
				continue
			}
		}

		if isGlue {
			// Glue logic: Remove immediately preceding newline from the buffer
			str := sb.String()
			lastNonWS := -1
			for i := len(str) - 1; i >= 0; i-- {
				c := str[i]
				if c != ' ' && c != '\t' && c != '\n' {
					lastNonWS = i
					break
				}
			}

			if lastNonWS < len(str)-1 {
				suffix := str[lastNonWS+1:]
				if strings.ContainsRune(suffix, '\n') {
					newSuffix := strings.ReplaceAll(suffix, "\n", "")
					sb.Reset()
					sb.WriteString(str[:lastNonWS+1])
					sb.WriteString(newSuffix)
				}
			}
			glueActive = true
		} else {
			if isNewline {
				if glueActive {
					continue
				}
			} else {
				if !isInlineWhitespace {
					glueActive = false
				}
			}
			sb.WriteString(txt)
		}
	}
	return sb.String()
}

// --- Internal Story Logic ---
func (s *Story) continueInternal(millisecsLimitAsync float64) error {
	// Ensure root has path? GetPath initializes it if nil.
	s.MainContent.GetPath()

	s.state.ResetOutput()

	// Step loop
	for s.canContinueInternal() {
		err := s.step()
		if err != nil {
			return err
		}
		if s.state.OutputStreamEndsInNewline() {
			break
		}
	}

	// Move generated choices to current choices
	if len(s.state.GeneratedChoices) > 0 {
		s.state.CurrentChoices = append(s.state.CurrentChoices, s.state.GeneratedChoices...)
		s.state.GeneratedChoices = make([]*Choice, 0)
	}

	return nil
}

// CanContinueInternal checks if the story logic can continue stepping.
func (s *Story) canContinueInternal() bool {
	return !s.state.GetCurrentPointer().IsNull()
}

// CanContinue checks if the story has more content to yield immediately.
// If true, the user can call Continue().
func (s *Story) CanContinue() bool {
	return s.canContinueInternal() && len(s.state.CurrentChoices) == 0
}

func (s *Story) processChoice(choicePoint *ChoicePoint) *Choice {
	showChoice := true

	// 1. Condition
	if choicePoint.HasCondition {
		conditionValue := s.state.PopEvaluationStack()
		if !s.IsTruthy(conditionValue) {
			showChoice = false
		}
	}

	startText := ""
	choiceOnlyText := ""
	// Tags TODO

	// 2. Choice Only Content (Pop from stack)
	if choicePoint.HasChoiceOnlyContent {
		obj := s.state.PopEvaluationStack()
		if strVal, ok := obj.(*StringValue); ok {
			choiceOnlyText = strVal.Value
		}
		// TODO: Handle tags which might be popped too?
	}

	// 3. Start Content
	if choicePoint.HasStartContent {
		obj := s.state.PopEvaluationStack()
		if strVal, ok := obj.(*StringValue); ok {
			startText = strVal.Value
		}
	}

	if !showChoice {
		return nil
	}

	text := startText + choiceOnlyText

	choice := NewChoice()
	choice.Text = text
	choice.SetPathStringOnChoice(choicePoint.PathStringOnChoice)

	// Resolve Absolute Path for Target
	// ChoicePoint path is the base.
	// TargetPath string is relative to it (usually).
	cpPath := choicePoint.GetPath()
	relPath := NewPathFromString(choicePoint.PathStringOnChoice)
	choice.TargetPath = cpPath.PathByAppendingPath(relPath)

	fmt.Printf("DEBUG: ProcessChoice: CP Path: %s, Rel: %s, Result: %s\n", cpPath, relPath, choice.TargetPath)

	choice.Index = len(s.state.GeneratedChoices)
	choice.SourcePath = cpPath.String()

	// Flags
	choice.OriginalThreadIndex = len(s.state.GetCallStack().Threads) - 1
	choice.IsInvisibleDefault = choicePoint.IsInvisibleDefault

	s.state.GeneratedChoices = append(s.state.GeneratedChoices, choice)

	return choice
}

func (s *Story) GetCurrentChoices() []*Choice {
	return s.state.CurrentChoices
}

// ChoosePathString moves the instruction pointer to the path given by the string.
func (s *Story) ChoosePathString(path string) error {
	p := NewPathFromString(path)
	pointer := s.PointerAtPath(p)
	if pointer.IsNull() {
		return fmt.Errorf("path not found: %s", path)
	}
	s.state.SetCurrentPointer(pointer)
	s.state.CurrentChoices = make([]*Choice, 0)
	return nil
}

func (s *Story) ChooseChoiceIndex(index int) error {
	if index < 0 || index >= len(s.state.CurrentChoices) {
		return fmt.Errorf("choice out of range")
	}

	choice := s.state.CurrentChoices[index]

	// Allow thread jumping etc (Simplified for now)

	// Divert
	s.state.SetCurrentPointer(s.PointerAtPath(choice.TargetPath))
	s.state.CurrentChoices = make([]*Choice, 0)

	return nil
}

func (s *Story) step() error {
	shouldAddToStream := true

	// Get current pointer
	pointer := s.state.GetCurrentPointer()
	if pointer.IsNull() {
		// End of content
		return nil
	}

	// Step directly to the first element of content in a container (if necessary)
	obj := pointer.Resolve()

	// Handle container entry (drilling down)
	// Iterate while we are pointing at a container
	// In Java: while(containerToEnter != nil) ...
	for obj != nil {
		container, isContainer := obj.(*Container)
		if !isContainer {
			break
		}

		// Mark container as being entered
		s.state.IncrementVisitCountForContainer(container)

		// No content? the most we can do is step past it
		if len(container.Content) == 0 {
			break
		}

		// Enter container
		pointer = StartOf(container)

		obj = pointer.Resolve()
	}

	s.state.SetCurrentPointer(pointer)

	// Verification after entry
	// If the container was empty, we stopped at the container itself.
	// But the pointer is still pointing at the container in the parent.
	// The nextContent() call effectively steps over it.

	// TODO: Profiler step

	// Is the current content Object:
	// - Normal content
	// - Or a logic/flow statement - if so, do it
	// - Stop flow if we hit a stack pop when we're unable to pop

	currentContentObj := pointer.Resolve()
	isLogicOrFlowControl := s.PerformLogicAndFlowControl(currentContentObj)

	// Has flow been forced to end by flow control above?
	if s.state.GetCurrentPointer().IsNull() {
		return nil
	}

	if isLogicOrFlowControl {
		shouldAddToStream = false
	}

	// TODO: Choice with condition?

	// If the container has no content, then it will be the "content" itself,
	// but we skip over it.
	if _, ok := currentContentObj.(*Container); ok {
		shouldAddToStream = false
	}

	// Content to add to evaluation stack or the output stream
	if shouldAddToStream {
		// TODO: VariablePointerValue context fix

		// Expression evaluation content
		if s.state.GetInExpressionEvaluation() {
			s.state.PushEvaluationStack(currentContentObj)
		} else {
			// Output stream content
			s.state.PushToOutputStream(currentContentObj)
		}
	}

	// Increment the content pointer, following diverts if necessary
	err := s.NextContent()
	if err != nil {
		return err
	}

	// TODO: StartThread command check

	return nil
}

func (s *Story) NextContent() error {
	// Setting previousContentObject is critical for VisitChangedContainersDueToDivert
	s.state.SetPreviousPointer(s.state.GetCurrentPointer())

	// Divert step?
	if !s.state.GetDivertedPointer().IsNull() {
		s.state.SetCurrentPointer(s.state.GetDivertedPointer())
		s.state.SetDivertedPointer(NullPointer)

		// TODO: visitChangedContainersDueToDivert()

		// Diverted location has valid content?
		if !s.state.GetCurrentPointer().IsNull() {
			return nil
		}

		// Otherwise drop down and attempt to increment
	}

	successfulPointerIncrement := s.IncrementContentPointer()

	// Ran out of content? Try to auto-exit from a function,
	// or finish evaluating the content of a thread
	if !successfulPointerIncrement {
		didPop := false

		if s.state.GetCallStack().CanPopType(PushPopTypeFunction) {
			s.state.PopCallStack(PushPopTypeFunction)

			if s.state.GetInExpressionEvaluation() {
				s.state.PushEvaluationStack(NewVoid())
			}
			didPop = true
		} else if s.state.GetCallStack().CanPop() {
			// Auto-pop ANY other type (Tunnel, etc)
			// We effectively finished the content of a container that was pushed to the stack
			s.state.PopCallStack(s.state.GetCallStack().CurrentElement().Type)
			didPop = true
		} else {
			s.state.TryExitFunctionEvaluationFromGame()
		}

		// Step past the point where we last called out

		if didPop && !s.state.GetCurrentPointer().IsNull() {
			return s.NextContent()
		}
	}
	return nil
}

func (s *Story) IncrementContentPointer() bool {
	successfulIncrement := true

	pointer := s.state.GetCurrentPointer() // Copy
	pointer.Index++

	// Each time we step off the end, we fall out to the next container
	for {
		// Check if index is valid for current container
		if pointer.Container == nil || pointer.Index < len(pointer.Container.Content) {
			break
		}

		successfulIncrement = false

		parent := pointer.Container.GetParent()
		nextAncestor, ok := parent.(*Container)
		if !ok || nextAncestor == nil {
			break
		}

		// Find index of current container in ancestor
		// We don't have indexOf easily unless we scan.
		// Optimized: Container usually knows its index? No.
		// Scan for now.
		indexInAncestor := -1
		for i, c := range nextAncestor.Content {
			if c == pointer.Container {
				indexInAncestor = i
				break
			}
		}

		if indexInAncestor == -1 {
			break
		}

		pointer = NewPointer(nextAncestor, indexInAncestor)

		// Increment to next content in outer container
		pointer.Index++

		successfulIncrement = true
	}

	if !successfulIncrement {
		pointer = NullPointer
	}

	s.state.SetCurrentPointer(pointer)

	return successfulIncrement
}

func (s *Story) PerformLogicAndFlowControl(content RuntimeObject) bool {
	if content == nil {
		return false
	}

	// Control Commands (Eval, Stack, etc)
	if evalCommand, ok := content.(*ControlCommand); ok {
		switch evalCommand.CommandType {
		case CommandTypeEvalStart:
			s.state.SetInExpressionEvaluation(true)
			return true
		case CommandTypeEvalEnd:
			s.state.SetInExpressionEvaluation(false)
			return true
		case CommandTypeEvalOutput:
			if len(s.state.EvaluationStack) > 0 {
				output := s.state.PopEvaluationStack()
				if _, isVoid := output.(*Void); !isVoid {
					s.state.PushToOutputStream(output)
				}
			}
			return true
		case CommandTypeNoOp:
			return true
		case CommandTypeDuplicate:
			s.state.PushEvaluationStack(s.state.PeekEvaluationStack())
			return true
		case CommandTypePopEvaluatedValue:
			s.state.PopEvaluationStack()
			return true
		case CommandTypePopFunction, CommandTypePopTunnel:
			pushPopType := PushPopTypeFunction
			if evalCommand.CommandType == CommandTypePopTunnel {
				pushPopType = PushPopTypeTunnel

				// Tunnel override: Check if there's a divert target on the evaluation stack
				if len(s.state.EvaluationStack) > 0 {
					peek := s.state.PeekEvaluationStack()
					if divertVal, ok := peek.(*DivertTargetValue); ok {
						s.state.PopEvaluationStack() // Consume it
						// Pop from the call stack to exit the tunnel
						err := s.state.PopCallStack(pushPopType) // Ignore error? Or should we?
						if err != nil {
							// If we can't pop, maybe we weren't in a tunnel?
							// But we still divert.
							// But standard behavior implies we consume the tunnel context.
						}
						// Divert to the target
						s.state.SetDivertedPointer(s.PointerAtPath(divertVal.GetTargetPath()))
						return true
					}
				}
			}

			// Pop from the call stack
			// For Tunnels, it's just a jump back
			err := s.state.PopCallStack(pushPopType)
			if err != nil {
				return true // Error state?
			}
			return true
		case CommandTypeStartThread:
			s.state.InThreadGeneration = true
			return true
		case CommandTypeDone:
			// If we're in a thread, kill it and return to the main thread
			if s.state.GetCallStack().CanPopThread() {
				s.state.GetCallStack().PopThread()
				// Trick: Set DivertedPointer to current to prevent NextContent from incrementing
				// effectively treating the resume as a jump to the current position.
				s.state.SetDivertedPointer(s.state.GetCurrentPointer())
				return true
			}
			s.state.SetCurrentPointer(NullPointer)
			return true
		}
		return true
	}

	// Divert
	if divert, ok := content.(*Divert); ok {
		if s.state.InThreadGeneration {
			if divert.PushesToStack {
				s.state.CallStack.Push(divert.StackPushType, 0, 0)
			}

			s.state.InThreadGeneration = false
			s.state.CallStack.PushThread()

			// Fork complete. The new thread (Active) will take the divert below.
			// The old thread (Inactive) needs to step over this instruction.
			prevThread := s.state.CallStack.Threads[len(s.state.CallStack.Threads)-2]
			elem := prevThread.CallStack[len(prevThread.CallStack)-1]
			elem.CurrentPointer.Index++
		}

		if divert.IsConditional {
			if len(s.state.EvaluationStack) > 0 {
				cond := s.state.PopEvaluationStack()
				if !s.IsTruthy(cond) {
					return true
				}
			}
		}

		if divert.IsExternal {
			err := s.callExternalFunction(divert.TargetPath.String(), divert.ExternalArgs)
			if err != nil {
				fmt.Printf("Error calling external function %s: %v\n", divert.TargetPath.String(), err)
			}
			return true
		}

		targetPath := divert.TargetPath

		// Variable Divert
		if divert.VariableDivertName != "" {
			val := s.state.GetVariablesState().GetVariableWithName(divert.VariableDivertName)
			if val != nil {
				if dv, ok := val.(*DivertTargetValue); ok {
					targetPath = dv.GetTargetPath()
				}
			}
		}

		if divert.PushesToStack {
			s.state.CallStack.Push(divert.StackPushType, 0, 0)
		}

		if targetPath == nil && divert.VariableDivertName == "" {
			return true // Should act as DONE? or Error?
		}

		if targetPath != nil && targetPath.IsRelative {
			ptr := s.state.GetCurrentPointer()
			if !ptr.IsNull() {
				targetPath = ptr.Path().PathByAppendingPath(targetPath)
			}
		}

		s.state.SetDivertedPointer(s.PointerAtPath(targetPath))
		return true
	}

	// Variable Reference
	if varRef, ok := content.(*VariableReference); ok {
		val := s.state.GetVariablesState().GetVariableWithName(varRef.Name)
		if val == nil {
			// Warning? For now default 0
			// s.state.AddWarning("Variable not found: " + varRef.Name)
			val = NewIntValue(0)
		}
		s.state.PushEvaluationStack(val)
		return true
	}

	// Variable Assignment
	if varAss, ok := content.(*VariableAssignment); ok {
		val := s.state.PopEvaluationStack()
		if val == nil {
			return true // Error?
		}
		err := s.state.GetVariablesState().Assign(varAss, val)
		if err != nil {
			// Panic or error?
			// s.state.AddError(err.Error())
			return true
		}
		return true
	}

	// Native Function Call
	if nativeFunc, ok := content.(*NativeFunctionCall); ok {
		params := make([]RuntimeObject, nativeFunc.NumberOfParameters)
		for i := nativeFunc.NumberOfParameters - 1; i >= 0; i-- {
			params[i] = s.state.PopEvaluationStack()
		}
		result, err := nativeFunc.Call(params)
		if err != nil {
			// Handle error logic
			return true
		}
		s.state.PushEvaluationStack(result)
		return true
	}

	// Choice Point
	if choicePoint, ok := content.(*ChoicePoint); ok {
		s.processChoice(choicePoint)
		return true
	}

	// Divert
	if divert, ok := content.(*Divert); ok {
		if s.state.InThreadGeneration {
			shouldFork := true
			if divert.IsConditional {
				val := s.state.PopEvaluationStack()
				shouldFork = s.IsTruthy(val)
			}

			if shouldFork {
				s.state.GetCallStack().Fork()
				s.state.InThreadGeneration = false

				// Advance the previous thread (now second on stack) past this divert
				// so it continues flow instead of diverting when it resumes.
				prevThread := s.state.GetCallStack().Threads[len(s.state.GetCallStack().Threads)-2]
				prevThread.CallStack[len(prevThread.CallStack)-1].CurrentPointer.Index++
			} else {
				// Condition false, just continue main flow.
				// InThreadGeneration was set, but we consumed it effectively.
				s.state.InThreadGeneration = false
				return true
			}
		} else if divert.IsConditional {
			// Check logic
			val := s.state.PopEvaluationStack()
			if !s.IsTruthy(val) {
				return true
			}
		}

		if divert.HasVariableTarget() {
			varName := divert.GetVariableDivertName()
			varVal := s.state.GetVariablesState().GetVariableWithName(varName)

			if divTarget, ok := varVal.(*DivertTargetValue); ok {
				s.state.SetDivertedPointer(s.PointerAtPath(divTarget.GetTargetPath()))
			} else {
				// Error or other logic
				s.state.SetDivertedPointer(NullPointer) // Fail safely?
			}
		} else if divert.IsExternal {
			funcName := divert.GetTargetPath().String() // Usually stored here? Or VariableDivertName?
			// Checking divert.go: IsExternal is a flag.
			// If external, args is Divert.ExternalArgs.
			// Target Path holds the function name usually.
			err := s.callExternalFunction(funcName, divert.ExternalArgs)
			if err != nil {
				return true // Error handling?
			}
			return true
		} else {
			targetPath := divert.GetTargetPath()
			if targetPath.IsRelative {
				context := divert.GetParent()
				if context != nil {
					// Make absolute by appending to context's path
					// Path logic: parent.Path + targetPath
					parentPath := context.GetPath()
					// If parent path is nil/empty (root context?), handle carefully
					// Usually runtime objects have paths.
					targetPath = parentPath.PathByAppendingPath(targetPath)
				}
			}
			s.state.SetDivertedPointer(s.PointerAtPath(targetPath))
		}

		if divert.PushesToStack {
			s.state.GetCallStack().Push(divert.StackPushType, 0, len(s.state.GetOutputStream()))
		}

		return true
	}

	return false
}

func (s *Story) IsTruthy(obj RuntimeObject) bool {
	if obj == nil {
		return false
	}
	if val, ok := obj.(Value); ok {
		return val.IsTruthy()
	}
	return true // Objects are truthy?
}

func (s *Story) PointerAtPath(path *Path) Pointer {
	if path == nil || len(path.Components) == 0 {
		return NullPointer
	}

	var currentObj RuntimeObject = s.MainContent

	// fmt.Printf("DEBUG: Resolving Path: %s\n", path.String())

	for _, component := range path.Components {
		// If current object is a container, try to find child
		container, isContainer := currentObj.(*Container)

		if !isContainer {
			fmt.Printf("DEBUG: PointerAtPath failed. CurrentObj is not container. Type: %T. Path: %s. Component: %s\n", currentObj, path.String(), component.String())
			return NullPointer
		}

		child, err := container.ContentAtPathComponent(component)
		if err != nil {
			fmt.Printf("DEBUG: PointerAtPath failed at component '%s' (Path: %s). Error: %v. Keys in container: %v\n", component, path.String(), err, keys(container.NamedContent))
			return NullPointer
		}
		currentObj = child
	}

	// If the target is a container, we want to point to the start of its content
	if container, ok := currentObj.(*Container); ok {
		return StartOf(container)
	}

	return s.PointerAtContent(currentObj)
}

func (s *Story) PointerAtContent(obj RuntimeObject) Pointer {
	if obj == nil {
		return NullPointer
	}
	parent := obj.GetParent()
	if parent == nil {
		if c, ok := obj.(*Container); ok {
			return StartOf(c)
		}
		return NullPointer
	}

	container, ok := parent.(*Container)
	if !ok {
		return NullPointer
	}

	for i, c := range container.Content {
		if c == obj {
			return NewPointer(container, i)
		}
	}

	return NullPointer
}
