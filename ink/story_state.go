package ink

import (
	"math/rand"
	"time"
)

// StoryState represents the state of the story.
// All story state information is included here.
type StoryState struct {
	// State
	VariablesState  *VariablesState
	CallStack       *CallStack
	EvaluationStack []RuntimeObject
	DivertedPointer Pointer

	VisitCounts map[*Container]int // Pointer map, effectively identity map
	TurnIndices map[*Container]int

	CurrentTurnIndex int
	StorySeed        int
	PreviousRandom   int
	DidSafeExit      bool

	Story *Story

	CurrentChoices   []*Choice
	GeneratedChoices []*Choice

	CurrentFlow    *Flow
	NamedFlows     map[string]*Flow
	AliveFlowNames []string

	OutputStreamDirty     bool
	OutputStreamTagsDirty bool
	CurrentTags           []string
	CurrentErrors         []string
	CurrentWarnings       []string

	InThreadGeneration bool
}

// NewStoryState creates a new StoryState.
func NewStoryState(story *Story) *StoryState {
	ss := &StoryState{
		Story:            story,
		VisitCounts:      make(map[*Container]int),
		TurnIndices:      make(map[*Container]int),
		CurrentTurnIndex: -1,
		EvaluationStack:  make([]RuntimeObject, 0),
		CurrentChoices:   make([]*Choice, 0),
		GeneratedChoices: make([]*Choice, 0),
		NamedFlows:       make(map[string]*Flow),
		AliveFlowNames:   []string{}, // Will be populated
	}

	// Seed random
	rand.Seed(time.Now().UnixNano())
	ss.StorySeed = rand.Intn(100)
	ss.PreviousRandom = 0

	ss.GoToStart()
	return ss
}

func (ss *StoryState) GoToStart() {
	ss.CallStack = NewCallStack(ss.Story.MainContent)
	// Provide method to access ListDefs in Story
	if ss.VariablesState == nil {
		ss.VariablesState = NewVariablesState(ss.CallStack, ss.Story.GetListDefinitions())
	} else {
		ss.VariablesState.CallStack = ss.CallStack
	}

	// Initial Flow
	ss.CurrentFlow = NewFlow("DEFAULT_FLOW", ss.Story)
	ss.NamedFlows["DEFAULT_FLOW"] = ss.CurrentFlow

	ss.OutputStreamDirty = true
	ss.AliveFlowNames = []string{"DEFAULT_FLOW"}

	ss.VisitCounts = make(map[*Container]int)
	ss.TurnIndices = make(map[*Container]int)
	ss.CurrentTurnIndex = -1

	// Start
	ss.CallStack.Reset()
	ss.CurrentFlow.CallStack = ss.CallStack
	ss.OutputStreamDirty = true
}

func (ss *StoryState) GetCallStack() *CallStack {
	if ss.CurrentFlow == nil {
		return nil
	}
	return ss.CurrentFlow.CallStack
}

// GetCurrentPointer returns the current pointer.
func (ss *StoryState) GetCurrentPointer() Pointer {
	return ss.GetCallStack().CurrentElement().CurrentPointer
}

// SetCurrentPointer sets the current pointer.
func (ss *StoryState) SetCurrentPointer(p Pointer) {
	ss.GetCallStack().CurrentElement().CurrentPointer = p
}

// SetPreviousPointer sets the previous pointer.
func (ss *StoryState) SetPreviousPointer(p Pointer) {
	ss.GetCallStack().CurrentThread().PreviousPointer = p
}

// GetDivertedPointer returns the diverted pointer.
func (ss *StoryState) GetDivertedPointer() Pointer {
	return ss.DivertedPointer
}

// SetDivertedPointer sets the diverted pointer.
func (ss *StoryState) SetDivertedPointer(p Pointer) {
	ss.DivertedPointer = p
}

// PopCallStack pops the call stack.
func (ss *StoryState) PopCallStack(type_ PushPopType) error {
	// Security Check: Clean up stack after tunnels to prevent "Dirty Stack" bugs
	var heightWhenPushed int
	if type_ == PushPopTypeTunnel {
		heightWhenPushed = ss.GetCallStack().CurrentElement().EvaluationStackHeightWhenPushed
	}

	err := ss.GetCallStack().Pop(type_)
	if err != nil {
		return err
	}

	// For tunnels, we forcefully clean up any debris left on the evaluation stack
	if type_ == PushPopTypeTunnel {
		if len(ss.EvaluationStack) > heightWhenPushed {
			ss.EvaluationStack = ss.EvaluationStack[:heightWhenPushed]
		}
	}

	return nil
}

// TryExitFunctionEvaluationFromGame tries to exit function evaluation from game.
func (ss *StoryState) TryExitFunctionEvaluationFromGame() {
	if ss.GetCallStack().CanPopType(PushPopTypeFunctionEvaluationFromGame) {
		ss.GetCallStack().Pop(PushPopTypeFunctionEvaluationFromGame)
	}
}

// GetInExpressionEvaluation returns if we are in expression evaluation.
func (ss *StoryState) GetInExpressionEvaluation() bool {
	return ss.GetCallStack().CurrentElement().InExpressionEvaluation
}

// SetInExpressionEvaluation sets if we are in expression evaluation.
func (ss *StoryState) SetInExpressionEvaluation(active bool) {
	ss.GetCallStack().CurrentElement().InExpressionEvaluation = active
}

// PushEvaluationStack pushes an object to the evaluation stack.
func (ss *StoryState) PushEvaluationStack(obj RuntimeObject) {
	ss.EvaluationStack = append(ss.EvaluationStack, obj)
}

// PopEvaluationStack pops an object from the evaluation stack.
func (ss *StoryState) PopEvaluationStack() RuntimeObject {
	if len(ss.EvaluationStack) == 0 {
		return nil
	}
	obj := ss.EvaluationStack[len(ss.EvaluationStack)-1]
	ss.EvaluationStack = ss.EvaluationStack[:len(ss.EvaluationStack)-1]
	return obj
}

// PeekEvaluationStack peeks at the top of the evaluation stack.
func (ss *StoryState) PeekEvaluationStack() RuntimeObject {
	if len(ss.EvaluationStack) == 0 {
		return nil
	}
	return ss.EvaluationStack[len(ss.EvaluationStack)-1]
}

// OutputStreamEndsInNewline checks if the output stream ends in a newline.
func (ss *StoryState) OutputStreamEndsInNewline() bool {
	if len(ss.CurrentFlow.OutputStream) == 0 {
		return false
	}
	lastObj := ss.CurrentFlow.OutputStream[len(ss.CurrentFlow.OutputStream)-1]
	if strVal, ok := lastObj.(*StringValue); ok {
		return strVal.isNewline || strVal.Value == "\n"
	}
	return false
}

// PushToOutputStream pushes an object to the output stream.
func (ss *StoryState) PushToOutputStream(obj RuntimeObject) {
	// Javas implementation simply adds to list
	ss.CurrentFlow.OutputStream = append(ss.CurrentFlow.OutputStream, obj)
	ss.OutputStreamDirty = true
}

// GetOutputStream returns the current flow's output stream.
func (ss *StoryState) GetOutputStream() []RuntimeObject {
	return ss.CurrentFlow.OutputStream
}

// GetGeneratedChoices returns the list of choices generated in the current chunk.
func (ss *StoryState) GetGeneratedChoices() []*Choice {
	return ss.CurrentFlow.CurrentChoices
}

// AddGeneratedChoice adds a choice.
func (ss *StoryState) AddGeneratedChoice(choice *Choice) {
	ss.CurrentFlow.CurrentChoices = append(ss.CurrentFlow.CurrentChoices, choice)
}

// HasError returns true if there are errors.
func (ss *StoryState) HasError() bool {
	return len(ss.CurrentErrors) > 0
}

// HasWarning returns true if there are warnings.
func (ss *StoryState) HasWarning() bool {
	return len(ss.CurrentWarnings) > 0
}

func (ss *StoryState) GetCurrentErrors() []string {
	return ss.CurrentErrors
}

func (ss *StoryState) GetCurrentWarnings() []string {
	return ss.CurrentWarnings
}

func (ss *StoryState) ResetOutput() {
	ss.CurrentFlow.OutputStream = make([]RuntimeObject, 0)
	ss.CurrentFlow.CurrentChoices = make([]*Choice, 0)
	ss.OutputStreamDirty = true
}

func (ss *StoryState) SetDidSafeExit(didSafeExit bool) {
	ss.DidSafeExit = didSafeExit
}

func (ss *StoryState) GetVariablesState() *VariablesState {
	return ss.VariablesState
}

// VisitCountForContainer returns the visit count for a container.
func (ss *StoryState) VisitCountForContainer(container *Container) int {
	if count, ok := ss.VisitCounts[container]; ok {
		return count
	}
	return 0
}

func (ss *StoryState) IncrementVisitCountForContainer(container *Container) {
	count := 0
	if c, ok := ss.VisitCounts[container]; ok {
		count = c
	}
	count++
	ss.VisitCounts[container] = count
}

func (ss *StoryState) RecordTurnIndexVisitToContainer(container *Container) {
	ss.TurnIndices[container] = ss.CurrentTurnIndex
}
