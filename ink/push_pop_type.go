package ink

// PushPopType is the type of push/pop operation.
type PushPopType int

// PushPopType constants.
const (
	PushPopTypeTunnel PushPopType = iota
	PushPopTypeFunction
	PushPopTypeFunctionEvaluationFromGame
)
