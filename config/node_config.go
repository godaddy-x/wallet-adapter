package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseNodeConfigJSON parses ow_symbol_node_config.nodeConfig.
// Scalar fields become MapConfig (numbers coerced to strings); verifyAPIs is extracted separately.
func ParseNodeConfigJSON(raw string) (MapConfig, []string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return MapConfig{}, nil, nil
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		return nil, nil, err
	}
	var verify []string
	if v, ok := generic["verifyAPIs"]; ok {
		if err := json.Unmarshal(v, &verify); err != nil {
			return nil, nil, fmt.Errorf("verifyAPIs must be a JSON string array: %w", err)
		}
		delete(generic, "verifyAPIs")
	}
	kv := make(MapConfig, len(generic))
	for k, rm := range generic {
		var s string
		if err := json.Unmarshal(rm, &s); err == nil {
			kv[k] = s
			continue
		}
		var n json.Number
		if err := json.Unmarshal(rm, &n); err == nil {
			kv[k] = n.String()
			continue
		}
		var b bool
		if err := json.Unmarshal(rm, &b); err == nil {
			if b {
				kv[k] = "true"
			} else {
				kv[k] = "false"
			}
		}
	}
	out := make([]string, 0, len(verify))
	for _, u := range verify {
		u = strings.TrimSpace(u)
		if u != "" {
			out = append(out, u)
		}
	}
	return kv, out, nil
}
