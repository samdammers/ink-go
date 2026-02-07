package ink

import (
	"fmt"
	"strings"
)

// JTokenToRuntimeObject converts a JSON token (any) into a RuntimeObject.
func JTokenToRuntimeObject(token any) (RuntimeObject, error) {
	switch v := token.(type) {
	case float64:
		// Convert to int if strictly integer, to match Ink C# behavior which preserves Int types
		if v == float64(int(v)) {
			return NewIntValue(int(v)), nil
		}
		return NewFloatValue(v), nil
	case int:
		return NewIntValue(v), nil
	case bool:
		return NewBoolValue(v), nil
	case string:
		return jStringToRuntimeObject(v)

	case []any:
		return JArrayToContainer(v)

	case map[string]any:
		// Handle maps (diverts, choices, etc.)
		return JMapToRuntimeObject(v)

	case nil:
		return nil, nil
	}

	return nil, fmt.Errorf("failed to convert token to runtime object: %v (type %T)", token, token)
}

// JArrayToContainer converts a JSON array to a Container.
func JArrayToContainer(jArray []any) (*Container, error) {
	container := NewContainer()

	// 1. Parse content (all but last element)
	contentCount := len(jArray)
	if contentCount > 0 {
		lastEl := jArray[contentCount-1]
		// Check if last element is metadata
		if _, ok := lastEl.(map[string]any); ok || lastEl == nil {
			contentCount--
		}
	}

	for i := 0; i < contentCount; i++ {
		v := jArray[i]
		obj, err := JTokenToRuntimeObject(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content at index %d: %w", i, err)
		}
		if obj != nil {
			if err := container.AddContent(obj); err != nil {
				return nil, err
			}
		}
	}

	// 2. Parse metadata from the last element if it exists
	if len(jArray) > 0 {
		lastEl := jArray[len(jArray)-1]
		if md, ok := lastEl.(map[string]any); ok {
			if err := parseContainerMetadata(container, md); err != nil {
				return nil, err
			}
		}
	}

	return container, nil
}

// JMapToRuntimeObject converts a JSON map to a runtime object.
func JMapToRuntimeObject(jMap map[string]any) (RuntimeObject, error) {
	if obj, ok := parseVariableOperation(jMap); ok {
		return obj, nil
	}

	if obj, ok := parseDivert(jMap); ok {
		return obj, nil
	}

	if obj, ok := parseList(jMap); ok {
		return obj, nil
	}

	if obj, ok := parseDivertTargetValue(jMap); ok {
		return obj, nil
	}

	if obj, ok := parseChoicePoint(jMap); ok {
		return obj, nil
	}

	return nil, fmt.Errorf("map tokens not fully implemented yet: %v", jMap)
}

// JObjectToRuntime converts a JSON object to a runtime container.
func JObjectToRuntime(root map[string]any) (*Container, error) {
	rootToken, ok := root["root"]
	if !ok {
		return nil, fmt.Errorf("root object not found in json")
	}

	rootList, ok := rootToken.([]any)
	if !ok {
		return nil, fmt.Errorf("root is not a list")
	}

	return JArrayToContainer(rootList)
}

func jStringToRuntimeObject(v string) (RuntimeObject, error) {
	if strings.HasPrefix(v, "^") {
		return NewStringValue(v[1:]), nil
	}
	if v == "\n" {
		return NewStringValue("\n"), nil
	}
	if v == "<>" {
		return NewGlue(), nil
	}

	if obj, ok := parseDivertShorthand(v); ok {
		return obj, nil
	}

	if obj, ok := parseControlCommand(v); ok {
		return obj, nil
	}

	if v == "L^" {
		v = "^"
	}

	// Native functions
	if NativeFunctionCallNumberOfParameters(v) > 0 || v == NativeFunctionCallNegate || v == NativeFunctionCallNot || v == NativeFunctionCallInt || v == NativeFunctionCallFloat {
		return NewNativeFunctionCall(v), nil
	}

	return nil, fmt.Errorf("unknown token type: string %v", v)
}

func parseDivertShorthand(v string) (RuntimeObject, bool) {
	// Divert shorthand (from test cases)
	if strings.HasPrefix(v, "-> ") {
		target := strings.TrimPrefix(v, "-> ")
		if !strings.HasPrefix(target, ".") {
			target = ".^." + target
		}
		div := NewDivert()
		div.TargetPath = NewPathFromString(target)
		return div, true
	}

	if strings.HasPrefix(v, "->t-> ") {
		target := strings.TrimPrefix(v, "->t-> ")
		if !strings.HasPrefix(target, ".") {
			target = ".^." + target
		}
		div := NewDivertWithPushType(PushPopTypeTunnel)
		div.TargetPath = NewPathFromString(target)
		return div, true
	}
	return nil, false
}

func parseControlCommand(v string) (RuntimeObject, bool) {
	switch v {
	case "done":
		return NewControlCommand(CommandTypeDone), true
	case "ev":
		return NewControlCommand(CommandTypeEvalStart), true
	case "/ev":
		return NewControlCommand(CommandTypeEvalEnd), true
	case "out":
		return NewControlCommand(CommandTypeEvalOutput), true
	case "pop":
		return NewControlCommand(CommandTypePopEvaluatedValue), true
	case "du":
		return NewControlCommand(CommandTypeDuplicate), true
	case "str":
		return NewControlCommand(CommandTypeBeginString), true
	case "/str":
		return NewControlCommand(CommandTypeEndString), true
	case "nop":
		return NewControlCommand(CommandTypeNoOp), true
	case "thread":
		return NewControlCommand(CommandTypeStartThread), true
	case "->->":
		return NewControlCommand(CommandTypePopTunnel), true
	case VoidName:
		return NewVoid(), true
	case "end":
		return NewControlCommand(CommandTypeEnd), true
	}
	return nil, false
}

func parseContainerMetadata(container *Container, md map[string]any) error {
	for k, v := range md {
		if k == "#n" {
			if name, ok := v.(string); ok {
				container.SetName(name)
			}
			continue
		}
		if k == "#f" {
			parseContainerFlags(container, v)
			continue
		}
		if k == "flg" || k == "listDefs" {
			continue
		}

		if err := parseNamedContent(container, k, v); err != nil {
			return err
		}
	}
	return nil
}

func parseVariableOperation(jMap map[string]any) (RuntimeObject, bool) {
	if v, ok := jMap["VAR?"]; ok {
		return NewVariableReference(v.(string)), true
	}
	if v, ok := jMap["VAR="]; ok {
		varName := v.(string)
		isNewDecl := true
		if _, ok := jMap["re"]; ok {
			isNewDecl = false
		}
		va := NewVariableAssignment(varName, isNewDecl)
		va.isGlobal = true // Usually VAR= is global
		return va, true
	}
	if v, ok := jMap["temp="]; ok {
		varName := v.(string)
		isNewDecl := true
		if _, ok := jMap["re"]; ok {
			isNewDecl = false
		}
		va := NewVariableAssignment(varName, isNewDecl)
		va.isGlobal = false
		return va, true
	}
	return nil, false
}

func parseDivert(jMap map[string]any) (RuntimeObject, bool) {
	if v, ok := jMap["->"]; ok {
		div := NewDivert()
		div.TargetPath = NewPathFromString(v.(string))
		if _, ok := jMap["c"]; ok {
			div.IsConditional = true
		}
		if _, ok := jMap["var"]; ok {
			div.VariableDivertName = v.(string)
			div.TargetPath = nil
		}
		return div, true
	}
	if v, ok := jMap["->t->"]; ok {
		div := NewDivertWithPushType(PushPopTypeTunnel)
		div.TargetPath = NewPathFromString(v.(string))
		if _, ok := jMap["c"]; ok {
			div.IsConditional = true
		}
		return div, true
	}
	if v, ok := jMap["f()"]; ok {
		div := NewDivertWithPushType(PushPopTypeFunction)
		div.TargetPath = NewPathFromString(v.(string))
		return div, true
	}
	if v, ok := jMap["x()"]; ok {
		div := NewDivert()
		div.IsExternal = true
		div.TargetPath = NewPathFromString(v.(string))
		if args, ok := jMap["exArgs"]; ok {
			div.ExternalArgs = int(args.(float64))
		}
		return div, true
	}
	return nil, false
}

func parseList(jMap map[string]any) (RuntimeObject, bool) {
	if v, ok := jMap["list"]; ok {
		if listMap, ok := v.(map[string]any); ok {
			inkList := NewList()
			for rawName, val := range listMap {
				var originName string
				var itemName string
				parts := strings.Split(rawName, ".")
				if len(parts) == 2 {
					originName = parts[0]
					itemName = parts[1]
				} else {
					itemName = rawName
				}
				var itemVal int
				if f, ok := val.(float64); ok {
					itemVal = int(f)
				} else if i, ok := val.(int); ok {
					itemVal = i
				} else if s, ok := val.(string); ok {
					originName = rawName
					itemName = s
					itemVal = 1
				}
				inkList.Add(NewListItem(originName, itemName), itemVal)
			}
			return NewListValue(inkList), true
		}
	}
	return nil, false
}

func parseChoicePoint(jMap map[string]any) (RuntimeObject, bool) {
	if v, ok := jMap["*"]; ok {
		pathString := v.(string)
		cp := NewChoicePoint(false, false, false, false, false)
		cp.SetPathStringOnChoice(pathString)
		if flg, ok := jMap["flg"]; ok {
			cp.SetFlags(int(flg.(float64)))
		}
		return cp, true
	}
	if v, ok := jMap["+"]; ok {
		pathString := v.(string)
		cp := NewChoicePoint(false, false, false, false, false)
		cp.SetPathStringOnChoice(pathString)
		if flg, ok := jMap["flg"]; ok {
			cp.SetFlags(int(flg.(float64)))
		}
		return cp, true
	}
	return nil, false
}

func parseDivertTargetValue(jMap map[string]any) (RuntimeObject, bool) {
	if v, ok := jMap["^->"]; ok {
		if targetStr, ok := v.(string); ok {
			return NewDivertTargetValue(NewPathFromString(targetStr)), true
		}
	}
	return nil, false
}

func parseContainerFlags(container *Container, v any) {
	if flags, ok := v.(float64); ok {
		f := int(flags)
		if (f & 1) > 0 {
			container.VisitsShouldBeCounted = true
		}
		if (f & 2) > 0 {
			container.TurnIndexShouldBeCounted = true
		}
		if (f & 4) > 0 {
			container.CountingAtStartOnly = true
		}
	}
}

func parseNamedContent(container *Container, k string, v any) error {
	childObj, err := JTokenToRuntimeObject(v)
	if err != nil {
		return fmt.Errorf("failed to parse named content '%s': %w", k, err)
	}

	if c, ok := childObj.(*Container); ok {
		c.SetName(k)
	}

	return container.AddNamedContent(k, childObj)
}
