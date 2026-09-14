package game

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// AgentToolDefinitionByName returns the canonical tool definition used by HTTP and MCP.
func AgentToolDefinitionByName(name string) (AgentToolDefinition, bool) {
	for _, def := range AgentToolDefinitions() {
		if def.Name == name {
			return def, true
		}
	}
	return AgentToolDefinition{}, false
}

// ValidateAgentArguments validates the top-level JSON object against the subset of
// JSON Schema used by AgentToolDefinitions. Keeping validation close to the canonical
// schemas prevents HTTP/MCP adapters from drifting away from the documented contract.
func ValidateAgentArguments(def AgentToolDefinition, args map[string]any) error {
	schema := def.InputSchema
	props, _ := schema["properties"].(map[string]any)
	required := map[string]bool{}
	switch list := schema["required"].(type) {
	case []string:
		for _, key := range list {
			required[key] = true
		}
	case []any:
		for _, raw := range list {
			if key, ok := raw.(string); ok {
				required[key] = true
			}
		}
	}
	for key := range required {
		if _, ok := args[key]; !ok {
			return fmt.Errorf("缺少必填参数: %s", key)
		}
	}
	if extra, _ := schema["additionalProperties"].(bool); !extra {
		for key := range args {
			if _, ok := props[key]; !ok {
				return fmt.Errorf("未知参数: %s", key)
			}
		}
	}
	for key, value := range args {
		raw, ok := props[key]
		if !ok {
			continue
		}
		prop, _ := raw.(map[string]any)
		if err := validateAgentValue(key, value, prop); err != nil {
			return err
		}
	}
	return nil
}

func validateAgentValue(key string, value any, schema map[string]any) error {
	typ, _ := schema["type"].(string)
	switch typ {
	case "string":
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("参数 %s 必须是 string", key)
		}
		n := len([]rune(s))
		if min, ok := numberToInt(schema["minLength"]); ok && n < min {
			return fmt.Errorf("参数 %s 长度不能小于 %d", key, min)
		}
		if max, ok := numberToInt(schema["maxLength"]); ok && n > max {
			return fmt.Errorf("参数 %s 长度不能大于 %d", key, max)
		}
		if p, ok := schema["pattern"].(string); ok && p != "" {
			re, err := regexp.Compile(p)
			if err == nil && !re.MatchString(s) {
				return fmt.Errorf("参数 %s 不符合格式 %s", key, p)
			}
		}
		if rawEnum, ok := schema["enum"]; ok && !enumContains(rawEnum, s) {
			return fmt.Errorf("参数 %s 不是允许值", key)
		}
	case "integer":
		n, ok := valueToInt(value)
		if !ok {
			return fmt.Errorf("参数 %s 必须是 integer", key)
		}
		if min, ok := numberToInt(schema["minimum"]); ok && n < min {
			return fmt.Errorf("参数 %s 不能小于 %d", key, min)
		}
		if max, ok := numberToInt(schema["maximum"]); ok && n > max {
			return fmt.Errorf("参数 %s 不能大于 %d", key, max)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("参数 %s 必须是 boolean", key)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("参数 %s 必须是 array", key)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("参数 %s 必须是 object", key)
		}
	}
	return nil
}

func enumContains(raw any, value string) bool {
	switch list := raw.(type) {
	case []string:
		for _, item := range list {
			if item == value {
				return true
			}
		}
	case []any:
		for _, item := range list {
			if s, ok := item.(string); ok && s == value {
				return true
			}
		}
	}
	return false
}

func numberToInt(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	return valueToInt(v)
}

func valueToInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		if n != float64(int(n)) {
			return 0, false
		}
		return int(n), true
	case float32:
		if n != float32(int(n)) {
			return 0, false
		}
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

// AgentToolRisk returns the risk marker embedded in the tool definition.
func AgentToolRisk(name string) string {
	def, ok := AgentToolDefinitionByName(name)
	if !ok {
		return "unknown"
	}
	if v, ok := def.Meta["ashen-crown/risk"].(string); ok {
		return strings.ToLower(v)
	}
	return "unknown"
}

func AgentToolReadOnly(name string) bool {
	def, ok := AgentToolDefinitionByName(name)
	if !ok {
		return false
	}
	v, _ := def.Annotations["readOnlyHint"].(bool)
	return v
}
