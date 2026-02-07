package ink

import (
	"fmt"
	"sort"
	"strings"
)

// InkListItem represents an item in an ink list.
type InkListItem struct {
	OriginName string
	ItemName   string
}

// FullName returns the full name of the item.
func (i InkListItem) FullName() string {
	if i.OriginName == "" {
		return i.ItemName
	}
	return i.OriginName + "." + i.ItemName
}

// String returns the string representation of the item.
func (i InkListItem) String() string {
	return i.FullName()
}

// NewInkListItem creates a new InkListItem.
func NewInkListItem(originName, itemName string) InkListItem {
	return InkListItem{
		OriginName: originName,
		ItemName:   itemName,
	}
}

// InkList represents a list value in Ink.
// It maps InkListItem to int (value).
// Note: In C#, currently supported underlying type is int.
type InkList struct {
	Items map[InkListItem]int
	// Origins of the list items (definitions)
	Origins []*ListDefinition
}

// NewInkList creates a new empty InkList.
func NewInkList() *InkList {
	return &InkList{
		Items:   make(map[InkListItem]int),
		Origins: make([]*ListDefinition, 0),
	}
}

// Add adds an item to the list.
func (il *InkList) Add(item InkListItem, value int) {
	il.Items[item] = value
}

// Contains checks if the list contains an item.
func (il *InkList) Contains(item InkListItem) bool {
	_, ok := il.Items[item]
	return ok
}

// Remove removes an item from the list.
func (il *InkList) Remove(item InkListItem) {
	delete(il.Items, item)
}

// Union returns a new InkList containing items from both lists.
func (il *InkList) Union(other *InkList) *InkList {
	newItems := make(map[InkListItem]int)
	for k, v := range il.Items {
		newItems[k] = v
	}
	for k, v := range other.Items {
		newItems[k] = v
	}

	// Merge origins
	newOrigins := make([]*ListDefinition, 0, len(il.Origins)+len(other.Origins))
	newOrigins = append(newOrigins, il.Origins...)
	for _, o := range other.Origins {
		found := false
		for _, existing := range newOrigins {
			if existing == o {
				found = true
				break
			}
		}
		if !found {
			newOrigins = append(newOrigins, o)
		}
	}

	return &InkList{Items: newItems, Origins: newOrigins}
}

// Subtract returns a new InkList with items from the second list removed from the first.
func (il *InkList) Subtract(other *InkList) *InkList {
	newItems := make(map[InkListItem]int)
	for k, v := range il.Items {
		if !other.Contains(k) {
			newItems[k] = v
		}
	}

	// Keep origins from the first list (base)
	// Usually subtraction keeps the "type" of the original list.
	newOrigins := make([]*ListDefinition, len(il.Origins))
	copy(newOrigins, il.Origins)

	return &InkList{Items: newItems, Origins: newOrigins}
}

// Intersect returns a new InkList with items present in both lists.
func (il *InkList) Intersect(other *InkList) *InkList {
	newItems := make(map[InkListItem]int)
	for k, v := range il.Items {
		if other.Contains(k) {
			newItems[k] = v
		}
	}
	return &InkList{Items: newItems, Origins: il.Origins}
}

// Has returns true if this list contains all items from the other list.
func (il *InkList) Has(other *InkList) bool {
	for k := range other.Items {
		if !il.Contains(k) {
			return false
		}
	}
	return true
}

// -- Value Implementation for ListValue --

// ListValue wraps an InkList as a Runtime Value.
type ListValue struct {
	value[*InkList]
}

// NewListValue creates a new ListValue.
func NewListValue(list *InkList) *ListValue {
	lv := &ListValue{}
	lv.Value = list
	return lv
}

// GetValueType returns the type of the value.
func (lv *ListValue) GetValueType() ValueType {
	return ValueTypeList
}

// IsTruthy returns true if the list is not empty.
func (lv *ListValue) IsTruthy() bool {
	return len(lv.Value.Items) > 0
}

// Cast converts the list to a new type.
func (lv *ListValue) Cast(newType ValueType) (Value, error) {
	if newType == lv.GetValueType() {
		return lv, nil
	}

	switch newType {
	case ValueTypeInt:
		// Max item value? Or empty 0 / 1?
		// Ink logic: defined as Max Item Value if distinct, else...
		// Simplified: If empty 0, else max value of items
		maxVal := 0
		first := true
		for _, v := range lv.Value.Items {
			if first || v > maxVal {
				maxVal = v
				first = false
			}
		}
		if first { // empty
			return NewIntValue(0), nil
		}
		return NewIntValue(maxVal), nil

	case ValueTypeFloat:
		val, _ := lv.Cast(ValueTypeInt)
		if iVal, ok := val.(*IntValue); ok {
			return NewFloatValue(float64(iVal.Value)), nil
		}

	case ValueTypeString:
		type itemPair struct {
			k InkListItem
			v int
		}
		var sorted []itemPair
		for k, v := range lv.Value.Items {
			sorted = append(sorted, itemPair{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].v < sorted[j].v
		})

		var sb strings.Builder
		for i, pair := range sorted {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(pair.k.ItemName)
		}
		return NewStringValue(sb.String()), nil
	}

	return nil, fmt.Errorf("cannot cast ListValue to %v", newType)
}
