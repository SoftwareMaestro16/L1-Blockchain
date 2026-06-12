package observability

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func normalizeLabels(labels Labels) (Labels, string) {
	if len(labels) == 0 {
		return nil, ""
	}
	normalized := make(Labels, len(labels))
	for key, value := range labels {
		if !allowedLabelKey(key) {
			continue
		}
		normalized[key] = sanitizeLabelValue(key, value)
	}
	if len(normalized) == 0 {
		return nil, ""
	}
	keys := make([]string, 0, len(normalized))
	for key := range normalized {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+normalized[key])
	}
	return normalized, strings.Join(parts, "\xff")
}

func allowedLabelKey(key string) bool {
	switch key {
	case "action", "component", "denom", "module", "reason", "result", "type":
		return true
	default:
		return false
	}
}

func sanitizeLabelValue(key, value string) string {
	if value == "" || len(value) > 64 {
		return "redacted"
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_./:-", r) {
			continue
		}
		return "redacted"
	}
	if !allowedLabelValue(key, value) {
		return "redacted"
	}
	return value
}

func allowedLabelValue(key, value string) bool {
	values, ok := allowedLabelValues[key]
	if !ok {
		return false
	}
	_, ok = values[value]
	return ok
}

var allowedLabelValues = map[string]map[string]struct{}{
	"action": {
		"add_liquidity":		{},
		"ante":				{},
		"create_pool":			{},
		"finalize_block":		{},
		"record_collected_fees":	{},
		"remove_liquidity":		{},
		"swap":				{},
		"swap_exact_amount_in":		{},
		"update_params":		{},
	},
	"component":	{"metrics": {}},
	"denom":	{"naet": {}},
	"module":	{"app": {}, "dex": {}, "fees": {}, "contract-assets": {}},
	"reason": {
		"error":				{},
		"evidence_manipulation":		{},
		"equivocation":				{},
		"invalid":				{},
		"invalid_fee":				{},
		"key_compromise_response_failure":	{},
		"major_downtime":			{},
		"minor_downtime":			{},
		"missing_fee_tx":			{},
		"next_error":				{},
		"repeated_downtime":			{},
	},
	"result":	{"accepted": {}, "finalized": {}, "success": {}},
	"type":		{"alloc": {}},
}

func formatLabels(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s=%q`, key, labels[key]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func formatFloat(value float64) string {
	if invalidNumber(value) {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func invalidNumber(value float64) bool {
	return math.IsNaN(value) || math.IsInf(value, 0)
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

func knownMetric(name string) bool {
	for _, def := range Definitions {
		if def.Name == name {
			return true
		}
	}
	return false
}
