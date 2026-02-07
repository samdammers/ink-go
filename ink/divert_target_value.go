package ink

import "fmt"

// DivertTargetValue represents a value that is a divert target (a path).
type DivertTargetValue struct {
	*BaseRuntimeObject
	TargetPath *Path
}

// NewDivertTargetValue creates a new DivertTargetValue.
func NewDivertTargetValue(targetPath *Path) *DivertTargetValue {
	return &DivertTargetValue{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		TargetPath:        targetPath,
	}
}

// GetTargetPath returns the path of the divert target.
func (d *DivertTargetValue) GetTargetPath() *Path {
	return d.TargetPath
}

// GetValueType returns the type of the value (ValueTypeDivertTarget).
func (d *DivertTargetValue) GetValueType() ValueType {
	return ValueTypeDivertTarget
}

// IsTruthy returns false.
func (d *DivertTargetValue) IsTruthy() bool {
	return false // errors in java
}

// Cast returns the value as a different type.
func (d *DivertTargetValue) Cast(newType ValueType) (Value, error) {
	if newType == d.GetValueType() {
		return d, nil
	}
	return nil, fmt.Errorf("cannot cast DivertTargetValue")
}

// GetValueObject returns the target path.
func (d *DivertTargetValue) GetValueObject() any {
	return d.TargetPath
}
