package ink

import "fmt"

type DivertTargetValue struct {
	*BaseRuntimeObject
	TargetPath *Path
}

func NewDivertTargetValue(targetPath *Path) *DivertTargetValue {
	return &DivertTargetValue{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		TargetPath:        targetPath,
	}
}

func (d *DivertTargetValue) GetTargetPath() *Path {
	return d.TargetPath
}

func (d *DivertTargetValue) GetValueType() ValueType {
	return ValueTypeDivertTarget
}

func (d *DivertTargetValue) IsTruthy() bool {
	return false // errors in java
}

func (d *DivertTargetValue) Cast(newType ValueType) (Value, error) {
	if newType == d.GetValueType() {
		return d, nil
	}
	return nil, fmt.Errorf("cannot cast DivertTargetValue")
}

func (d *DivertTargetValue) GetValueObject() any {
	return d.TargetPath
}
