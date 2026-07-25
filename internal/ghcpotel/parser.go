package ghcpotel

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"
)

const SchemaVersion = "ghcp-otel-jsonl-v1"

const (
	maxExportBytes = 16 << 20
	maxExportRows  = 100_000
	maxExportSpans = 50_000
)

type Usage struct {
	SchemaVersion    string   `json:"schema_version"`
	AIUNano          int64    `json:"aiu_nano"`
	AIUAvailable     bool     `json:"aiu_available"`
	InputTokens      int64    `json:"input_tokens"`
	OutputTokens     int64    `json:"output_tokens"`
	CacheReadTokens  int64    `json:"cache_read_tokens"`
	CacheWriteTokens int64    `json:"cache_write_tokens"`
	Calls            int      `json:"calls"`
	ActualModels     []string `json:"actual_models"`
	RequestedModels  []string `json:"requested_models"`
}

type spanIdentity struct {
	trace string
	span  string
}

type spanRecord struct {
	parent      string
	fingerprint string
	name        string
	attributes  map[string]any
}

func Parse(reader io.Reader) (Usage, error) {
	usage := Usage{SchemaVersion: SchemaVersion}
	spans := make(map[spanIdentity]spanRecord)
	scanner := bufio.NewScanner(io.LimitReader(reader, maxExportBytes+1))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	totalBytes := 0
	rows := 0
	for line := 1; scanner.Scan(); line++ {
		rows++
		totalBytes += len(scanner.Bytes()) + 1
		if rows > maxExportRows || totalBytes > maxExportBytes {
			return Usage{}, fmt.Errorf("OTel export exceeds row or byte limit")
		}
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		row, err := decodeObject(scanner.Bytes())
		if err != nil {
			return Usage{}, fmt.Errorf("line %d: %w", line, err)
		}
		if err := rejectContentKeys(row); err != nil {
			return Usage{}, fmt.Errorf("line %d: %w", line, err)
		}
		if text(row["type"]) != "span" {
			continue
		}
		traceID := firstText(row, "traceId", "trace_id", "traceID")
		spanID := firstText(row, "spanId", "span_id", "spanID")
		if traceID == "" || spanID == "" {
			return Usage{}, fmt.Errorf("line %d: span requires trace and span identity", line)
		}
		parentID := firstText(row, "parentSpanId", "parent_span_id", "parentSpanID")
		canonical, err := json.Marshal(row)
		if err != nil {
			return Usage{}, fmt.Errorf("line %d: canonicalize span: %w", line, err)
		}
		fingerprint := fmt.Sprintf("%x", sha256.Sum256(canonical))
		identity := spanIdentity{trace: traceID, span: spanID}
		if previous, exists := spans[identity]; exists {
			if previous.fingerprint != fingerprint {
				return Usage{}, fmt.Errorf("line %d: conflicting duplicate trace=%q span=%q", line, traceID, spanID)
			}
			continue
		}
		if len(spans) >= maxExportSpans {
			return Usage{}, fmt.Errorf("OTel export exceeds span limit")
		}
		attributes, ok := row["attributes"].(map[string]any)
		if !ok {
			attributes = map[string]any{}
		}
		spans[identity] = spanRecord{
			parent: parentID, fingerprint: fingerprint,
			name: text(row["name"]), attributes: attributes,
		}
	}
	if err := scanner.Err(); err != nil {
		return Usage{}, err
	}
	if len(spans) == 0 {
		return Usage{}, fmt.Errorf("OTel export contains no spans")
	}
	children := make(map[spanIdentity]int)
	roots := make(map[string]int)
	for identity, span := range spans {
		if span.parent == "" {
			roots[identity.trace]++
			continue
		}
		if span.parent == identity.span {
			return Usage{}, fmt.Errorf("self-parent cycle trace=%q span=%q", identity.trace, identity.span)
		}
		parent := spanIdentity{trace: identity.trace, span: span.parent}
		if _, exists := spans[parent]; !exists {
			return Usage{}, fmt.Errorf("missing parent trace=%q span=%q", identity.trace, span.parent)
		}
		children[parent]++
	}
	for identity := range spans {
		if roots[identity.trace] == 0 {
			return Usage{}, fmt.Errorf("trace %q has no root span", identity.trace)
		}
	}
	if err := validateSpanGraphAcyclic(spans); err != nil {
		return Usage{}, err
	}

	actualModels := make(map[string]struct{})
	requestedModels := make(map[string]struct{})
	allPriced := true
	for identity, span := range spans {
		if !strings.HasPrefix(span.name, "chat ") || children[identity] != 0 {
			continue
		}
		attributes := span.attributes
		actual := text(attributes["gen_ai.response.model"])
		if actual == "" {
			return Usage{}, fmt.Errorf("leaf chat span actual model is required")
		}
		actualModels[actual] = struct{}{}
		if requested := text(attributes["gen_ai.request.model"]); requested != "" {
			requestedModels[requested] = struct{}{}
		}

		input, err := requiredNonNegativeInteger(attributes, "gen_ai.usage.input_tokens")
		if err != nil {
			return Usage{}, err
		}
		output, err := requiredNonNegativeInteger(attributes, "gen_ai.usage.output_tokens")
		if err != nil {
			return Usage{}, err
		}
		cacheRead, err := nonNegativeInteger(attributes, "gen_ai.usage.cache_read.input_tokens")
		if err != nil {
			return Usage{}, err
		}
		cacheWrite, err := nonNegativeInteger(attributes, "gen_ai.usage.cache_creation.input_tokens")
		if err != nil {
			return Usage{}, err
		}
		usage.InputTokens, err = checkedAdd(usage.InputTokens, input)
		if err != nil {
			return Usage{}, err
		}
		usage.OutputTokens, err = checkedAdd(usage.OutputTokens, output)
		if err != nil {
			return Usage{}, err
		}
		usage.CacheReadTokens, err = checkedAdd(usage.CacheReadTokens, cacheRead)
		if err != nil {
			return Usage{}, err
		}
		usage.CacheWriteTokens, err = checkedAdd(usage.CacheWriteTokens, cacheWrite)
		if err != nil {
			return Usage{}, err
		}
		usage.Calls++

		nano, priced, err := exactAIUNano(attributes)
		if err != nil {
			return Usage{}, err
		}
		if text(attributes["gen_ai.provider.name"]) == "github" && !priced {
			return Usage{}, fmt.Errorf("GitHub leaf chat span has no exact AIU field")
		}
		if !priced {
			allPriced = false
			continue
		}
		usage.AIUNano, err = checkedAdd(usage.AIUNano, nano)
		if err != nil {
			return Usage{}, err
		}
	}
	if usage.Calls == 0 {
		return Usage{}, fmt.Errorf("OTel export contains no leaf chat spans")
	}
	usage.AIUAvailable = usage.Calls > 0 && allPriced
	usage.ActualModels = sortedKeys(actualModels)
	usage.RequestedModels = sortedKeys(requestedModels)
	return usage, nil
}

func validateSpanGraphAcyclic(spans map[spanIdentity]spanRecord) error {
	const (
		visiting = 1
		complete = 2
	)
	state := make(map[spanIdentity]uint8, len(spans))
	for start := range spans {
		if state[start] == complete {
			continue
		}
		var path []spanIdentity
		current := start
		for {
			if state[current] == visiting {
				return fmt.Errorf("trace %q contains a parent cycle", start.trace)
			}
			if state[current] == complete {
				break
			}
			state[current] = visiting
			path = append(path, current)
			parent := spans[current].parent
			if parent == "" {
				break
			}
			current = spanIdentity{trace: current.trace, span: parent}
		}
		for _, identity := range path {
			state[identity] = complete
		}
	}
	return nil
}

func decodeObject(content []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("malformed OTel row: %w", err)
	}
	if decoder.More() {
		return nil, fmt.Errorf("multiple JSON values in one row")
	}
	return value, nil
}

func exactAIUNano(attributes map[string]any) (int64, bool, error) {
	aiu, hasAIU := attributes["github.copilot.aiu"]
	nano, hasNano := attributes["github.copilot.nano_aiu"]
	if hasAIU && hasNano {
		return 0, false, fmt.Errorf("chat span has conflicting AIU fields")
	}
	if hasNano {
		value, err := exactInteger(nano)
		if err != nil || value < 0 {
			return 0, false, fmt.Errorf("nano_aiu must be a non-negative integer")
		}
		return value, true, nil
	}
	if !hasAIU {
		return 0, false, nil
	}
	raw, ok := numberText(aiu)
	if !ok {
		return 0, false, fmt.Errorf("aiu must be numeric")
	}
	rational, ok := new(big.Rat).SetString(raw)
	if !ok || rational.Sign() < 0 {
		return 0, false, fmt.Errorf("aiu must be a non-negative decimal")
	}
	rational.Mul(rational, big.NewRat(1_000_000_000, 1))
	if !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, false, fmt.Errorf("aiu is not exactly representable in nano units")
	}
	return rational.Num().Int64(), true, nil
}

func nonNegativeInteger(attributes map[string]any, key string) (int64, error) {
	raw, exists := attributes[key]
	if !exists {
		return 0, nil
	}
	value, err := exactInteger(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}

func requiredNonNegativeInteger(attributes map[string]any, key string) (int64, error) {
	if _, exists := attributes[key]; !exists {
		return 0, fmt.Errorf("%s is required for exact accounting", key)
	}
	return nonNegativeInteger(attributes, key)
}

func exactInteger(value any) (int64, error) {
	raw, ok := numberText(value)
	if !ok {
		return 0, fmt.Errorf("not an integer")
	}
	rational, ok := new(big.Rat).SetString(raw)
	if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, fmt.Errorf("not an integer")
	}
	return rational.Num().Int64(), nil
}

func numberText(value any) (string, bool) {
	switch typed := value.(type) {
	case json.Number:
		return typed.String(), true
	case string:
		return typed, true
	default:
		return "", false
	}
}

func checkedAdd(left, right int64) (int64, error) {
	sum := new(big.Int).Add(big.NewInt(left), big.NewInt(right))
	if !sum.IsInt64() {
		return 0, fmt.Errorf("telemetry counter overflow")
	}
	return sum.Int64(), nil
}

func text(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func firstText(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := text(row[key]); value != "" {
			return value
		}
	}
	return ""
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
