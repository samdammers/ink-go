package ink

// StoryStateDto represents the state of the story for JSON serialization.
// It strictly mirrors the structure used in the Ink JSON save format.
type StoryStateDto struct {
	Flows               map[string]FlowDto     `json:"flows"`
	CurrentFlowName     string                 `json:"currentFlowName"`
	VariablesState      map[string]interface{} `json:"variablesState"`
	EvalStack           []interface{}          `json:"evalStack"`
	CurrentDivertTarget string                 `json:"currentDivertTarget,omitempty"`
	VisitCounts         map[string]int         `json:"visitCounts"`
	TurnIndices         map[string]int         `json:"turnIndices"`
	TurnIdx             int                    `json:"turnIdx"`
	StorySeed           int                    `json:"storySeed"`
	PreviousRandom      int                    `json:"previousRandom"`
	InkSaveVersion      int                    `json:"inkSaveVersion"`
	InkFormatVersion    int                    `json:"inkFormatVersion"`
}

// FlowDto represents a saved Flow.
type FlowDto struct {
	CallStack      CallStackDto                  `json:"callstack"`
	OutputStream   []interface{}                 `json:"outputStream"`
	ChoiceThreads  map[string]CallStackThreadDto `json:"choiceThreads,omitempty"`
	CurrentChoices []ChoiceDto                   `json:"currentChoices,omitempty"`
}

// ChoiceDto represents a saved Choice.
type ChoiceDto struct {
	Text                string   `json:"text"`
	Index               int      `json:"index"`
	OriginalChoicePath  string   `json:"originalChoicePath"`
	OriginalThreadIndex int      `json:"originalThreadIndex"`
	TargetPath          string   `json:"targetPath"`
	Tags                []string `json:"tags,omitempty"`
}

// CallStackDto represents the saved state of the CallStack.
type CallStackDto struct {
	Threads       []CallStackThreadDto `json:"threads"`
	ThreadCounter int                  `json:"threadCounter"`
}

// CallStackThreadDto represents a saved thread within the CallStack.
type CallStackThreadDto struct {
	CallStack             []CallStackElementDto `json:"callstack"`
	ThreadIndex           int                   `json:"threadIndex"`
	PreviousContentObject string                `json:"previousContentObject,omitempty"`
}

// CallStackElementDto represents a saved element on the call stack.
type CallStackElementDto struct {
	CPath              string                 `json:"cPath,omitempty"`
	Idx                int                    `json:"idx"`
	Exp                bool                   `json:"exp"`
	Type               int                    `json:"type"`
	TemporaryVariables map[string]interface{} `json:"temp,omitempty"`
}
