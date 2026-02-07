package ink

import (
	"fmt"
)

// CallStackElement represents a single frame in the call stack.
type CallStackElement struct {
	CurrentPointer                  Pointer
	InExpressionEvaluation          bool
	TemporaryVariables              map[string]RuntimeObject
	Type                            PushPopType
	EvaluationStackHeightWhenPushed int
	FunctionStartInOutputStream     int
}

// NewCallStackElement creates a new CallStackElement.
func NewCallStackElement(type_ PushPopType, pointer Pointer, inExpressionEvaluation bool) *CallStackElement {
	e := &CallStackElement{
		CurrentPointer:         Pointer{Container: pointer.Container, Index: pointer.Index}, // Copy
		InExpressionEvaluation: inExpressionEvaluation,
		TemporaryVariables:     make(map[string]RuntimeObject),
		Type:                   type_,
	}
	return e
}

// Copy creates a deep copy of the element.
func (e *CallStackElement) Copy() *CallStackElement {
	copy := &CallStackElement{
		CurrentPointer:                  Pointer{Container: e.CurrentPointer.Container, Index: e.CurrentPointer.Index},
		InExpressionEvaluation:          e.InExpressionEvaluation,
		TemporaryVariables:              make(map[string]RuntimeObject, len(e.TemporaryVariables)),
		Type:                            e.Type,
		EvaluationStackHeightWhenPushed: e.EvaluationStackHeightWhenPushed,
		FunctionStartInOutputStream:     e.FunctionStartInOutputStream,
	}
	for k, v := range e.TemporaryVariables {
		copy.TemporaryVariables[k] = v
	}
	return copy
}

// CallStackThread represents a thread of execution in the story.
type CallStackThread struct {
	CallStack       []*CallStackElement
	PreviousPointer Pointer
	ThreadIndex     int
}

// NewCallStackThread creates a new thread.
func NewCallStackThread() *CallStackThread {
	return &CallStackThread{
		CallStack:       make([]*CallStackElement, 0),
		PreviousPointer: Pointer{Container: nil, Index: -1}, // Null pointer
	}
}

// Copy creates a deep copy of the thread.
func (t *CallStackThread) Copy() *CallStackThread {
	copy := NewCallStackThread()
	copy.ThreadIndex = t.ThreadIndex
	for _, e := range t.CallStack {
		copy.CallStack = append(copy.CallStack, e.Copy())
	}
	copy.PreviousPointer = Pointer{Container: t.PreviousPointer.Container, Index: t.PreviousPointer.Index}
	return copy
}

// CallStack handles the call stack for the story.
type CallStack struct {
	Threads       []*CallStackThread
	ThreadCounter int
	StartOfRoot   Pointer
}

// NewCallStack creates a new CallStack instantiated with the story's main content.
func NewCallStack(mainContentContainer *Container) *CallStack {
	cs := &CallStack{
		StartOfRoot: StartOf(mainContentContainer),
	}
	cs.Reset()
	return cs
}

// Reset resets the call stack to its initial state.
func (cs *CallStack) Reset() {
	cs.Threads = []*CallStackThread{NewCallStackThread()}
	cs.Threads[0].CallStack = append(cs.Threads[0].CallStack, NewCallStackElement(PushPopTypeTunnel, cs.StartOfRoot, false))
}

// Copy creates a deep copy of the CallStack.
func (cs *CallStack) Copy() *CallStack {
	copy := &CallStack{
		Threads:       make([]*CallStackThread, len(cs.Threads)),
		ThreadCounter: cs.ThreadCounter,
		StartOfRoot:   Pointer{Container: cs.StartOfRoot.Container, Index: cs.StartOfRoot.Index},
	}
	for i, t := range cs.Threads {
		copy.Threads[i] = t.Copy()
	}
	return copy
}

// CurrentThread returns the current active thread.
func (cs *CallStack) CurrentThread() *CallStackThread {
	return cs.Threads[len(cs.Threads)-1]
}

// PushThread clones the current thread and pushes it onto the thread stack.
// This is used for "forking" the story flow (e.g. for `<- thread` commands).
func (cs *CallStack) PushThread() {
	currentThread := cs.CurrentThread()
	newThread := currentThread.Copy()
	cs.ThreadCounter++
	newThread.ThreadIndex = cs.ThreadCounter
	cs.Threads = append(cs.Threads, newThread)
}

// CurrentElement returns the current element (top of stack) of the current thread.
func (cs *CallStack) CurrentElement() *CallStackElement {
	thread := cs.CurrentThread()
	if len(thread.CallStack) == 0 {
		return nil
	}
	return thread.CallStack[len(thread.CallStack)-1]
}

// CurrentElementIndex returns the index of the current element.
func (cs *CallStack) CurrentElementIndex() int {
	return len(cs.CurrentThread().CallStack) - 1
}

// GetTemporaryVariableWithName returns a temporary variable by name.
func (cs *CallStack) GetTemporaryVariableWithName(name string, contextIndex int) RuntimeObject {
	if contextIndex == -1 {
		contextIndex = cs.CurrentElementIndex() + 1
	}

	thread := cs.CurrentThread()
	contextElement := thread.CallStack[contextIndex-1]

	if val, ok := contextElement.TemporaryVariables[name]; ok {
		return val
	}
	return nil
}

// SetTemporaryVariable sets a temporary variable.
func (cs *CallStack) SetTemporaryVariable(name string, value RuntimeObject, declareNew bool, contextIndex int) error {
	if contextIndex == -1 {
		contextIndex = cs.CurrentElementIndex() + 1
	}

	thread := cs.CurrentThread()
	contextElement := thread.CallStack[contextIndex-1]

	if !declareNew {
		if _, ok := contextElement.TemporaryVariables[name]; !ok {
			return fmt.Errorf("could not find temporary variable to set: %s", name)
		}
	}

	contextElement.TemporaryVariables[name] = value
	return nil
}

// ContextForVariableNamed determines the index of the context where the variable is defined.
// Returns 0 if global, or index > 0 for callstack element index + 1.
func (cs *CallStack) ContextForVariableNamed(name string) int {
	// Current temporary context?
	if _, ok := cs.CurrentElement().TemporaryVariables[name]; ok {
		return cs.CurrentElementIndex() + 1
	}
	// Default to global
	return 0
}

// CanPop returns true if the current thread has more than one element.
func (cs *CallStack) CanPop() bool {
	return len(cs.CurrentThread().CallStack) > 1
}

// CanPopType returns true if can pop and the type matches.
func (cs *CallStack) CanPopType(type_ PushPopType) bool {
	if !cs.CanPop() {
		return false
	}
	// TODO: Handle null type equivalent? Go Enum doesn't handle nil.
	// Assuming type passed is always valid.
	// If caller needs check, they check.
	return cs.CurrentElement().Type == type_
}

// Push pushes a new element onto the stack.
func (cs *CallStack) Push(type_ PushPopType, externalEvaluationStackHeight int, outputStreamLengthWithPushed int) {
	// When changing content pointer during a function call, we usually want to pointer
	// relative to where the function call is.
	element := NewCallStackElement(type_, cs.CurrentElement().CurrentPointer, false)
	element.EvaluationStackHeightWhenPushed = externalEvaluationStackHeight
	element.FunctionStartInOutputStream = outputStreamLengthWithPushed

	// The element we push should point to the instruction AFTER the current one,
	// so that when we return, we continue execution.
	element.CurrentPointer.Index++

	cs.CurrentThread().CallStack = append(cs.CurrentThread().CallStack, element)
}

// Pop pops the top element from the stack.
func (cs *CallStack) Pop(type_ PushPopType) error {
	if cs.CanPopType(type_) {
		thread := cs.CurrentThread()
		thread.CallStack = thread.CallStack[:len(thread.CallStack)-1]
		return nil
	}
	return fmt.Errorf("mismatched push/pop in CallStack")
}

// Fork creates a new thread.
func (cs *CallStack) Fork() {
	thread := cs.CurrentThread().Copy()
	cs.ThreadCounter++
	thread.ThreadIndex = cs.ThreadCounter
	cs.Threads = append(cs.Threads, thread)
}

// GetDepth returns the depth of the current thread's stack.
func (cs *CallStack) GetDepth() int {
	return len(cs.CurrentThread().CallStack)
}

// Elements returns the current thread's callstack.
func (cs *CallStack) Elements() []*CallStackElement {
	return cs.CurrentThread().CallStack
}
