package ink

import "fmt"

// CommandType is the type of a control command.
type CommandType int

// CommandType constants.
const (
	CommandTypeNotSet CommandType = iota
	CommandTypeEvalStart
	CommandTypeEvalOutput
	CommandTypeEvalEnd
	CommandTypeDuplicate
	CommandTypePopEvaluatedValue
	CommandTypePopFunction
	CommandTypePopTunnel
	CommandTypeBeginString
	CommandTypeEndString
	CommandTypeNoOp
	CommandTypeChoiceCount
	CommandTypeTurns
	CommandTypeTurnsSince
	CommandTypeReadCount
	CommandTypeRandom
	CommandTypeSeedRandom
	CommandTypeVisitIndex
	CommandTypeSequenceShuffleIndex
	CommandTypeStartThread
	CommandTypeDone
	CommandTypeEnd
	CommandTypeListFromInt
	CommandTypeListRange
	CommandTypeListRandom
	CommandTypeBeginTag
	CommandTypeEndTag
	CommandTypeNewline
)

var commandTypeNames = [...]string{
	"NotSet",
	"EvalStart",
	"EvalOutput",
	"EvalEnd",
	"Duplicate",
	"PopEvaluatedValue",
	"PopFunction",
	"PopTunnel",
	"BeginString",
	"EndString",
	"NoOp",
	"ChoiceCount",
	"Turns",
	"TurnsSince",
	"ReadCount",
	"Random",
	"SeedRandom",
	"VisitIndex",
	"SequenceShuffleIndex",
	"StartThread",
	"Done",
	"End",
	"ListFromInt",
	"ListRange",
	"ListRandom",
	"BeginTag",
	"EndTag",
	"Newline",
}

func (c CommandType) String() string {
	if c >= 0 && int(c) < len(commandTypeNames) {
		return commandTypeNames[c]
	}
	return "Unsupported"
}

// ControlCommand is a special RuntimeObject that represents a command to the
// story engine.
type ControlCommand struct {
	BaseRuntimeObject
	CommandType CommandType
}

// NewControlCommand creates a new control command value.
func NewControlCommand(cmdType CommandType) *ControlCommand {
	return &ControlCommand{CommandType: cmdType}
}

func (c *ControlCommand) String() string {
	return fmt.Sprintf("ControlCommand(%s)", c.CommandType)
}
