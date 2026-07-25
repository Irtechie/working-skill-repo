package ghcpotel

import (
	"strings"
	"testing"
)

func TestParseDeduplicatesByTraceAndSpanAndIgnoresAgentAggregate(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"span","traceId":"trace-a","spanId":"leaf","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.request.model":"requested","gen_ai.response.model":"actual","gen_ai.usage.input_tokens":100,"gen_ai.usage.output_tokens":25,"github.copilot.nano_aiu":2500000000}}`,
		`{"type":"span","traceId":"trace-a","spanId":"leaf","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.request.model":"requested","gen_ai.response.model":"actual","gen_ai.usage.input_tokens":100,"gen_ai.usage.output_tokens":25,"github.copilot.nano_aiu":2500000000}}`,
		`{"type":"span","traceId":"trace-b","spanId":"leaf","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"fallback","gen_ai.usage.input_tokens":0,"gen_ai.usage.output_tokens":0,"github.copilot.aiu":0.5}}`,
		`{"type":"span","traceId":"trace-a","spanId":"agent","name":"invoke_agent","attributes":{"github.copilot.nano_aiu":3000000000}}`,
	}, "\n")

	usage, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if usage.Calls != 2 || usage.AIUNano != 3_000_000_000 || !usage.AIUAvailable {
		t.Fatalf("usage=%+v", usage)
	}
	if usage.InputTokens != 100 || usage.OutputTokens != 25 {
		t.Fatalf("tokens=%+v", usage)
	}
	if strings.Join(usage.ActualModels, ",") != "actual,fallback" ||
		strings.Join(usage.RequestedModels, ",") != "requested" {
		t.Fatalf("models=%+v", usage)
	}
}

func TestParseRejectsIncompleteConflictingAndInexactAccounting(t *testing.T) {
	tests := map[string]string{
		"missing price":    `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","gen_ai.usage.input_tokens":1,"gen_ai.usage.output_tokens":1}}`,
		"missing input":    `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","gen_ai.usage.output_tokens":1,"github.copilot.nano_aiu":1}}`,
		"missing output":   `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","gen_ai.usage.input_tokens":1,"github.copilot.nano_aiu":1}}`,
		"two price fields": `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.aiu":1,"github.copilot.nano_aiu":1000000000}}`,
		"fractional nano":  `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":1.5}}`,
		"sub nano aiu":     `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.aiu":0.0000000001}}`,
		"missing trace":    `{"type":"span","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":1}}`,
		"missing actual":   `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.request.model":"requested","github.copilot.nano_aiu":1}}`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(input)); err == nil {
				t.Fatal("invalid accounting passed")
			}
		})
	}
}

func TestParseAllowsMissingCacheTokenFieldsAsZero(t *testing.T) {
	input := `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","gen_ai.usage.input_tokens":7,"gen_ai.usage.output_tokens":3,"github.copilot.nano_aiu":1}}`
	usage, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if usage.InputTokens != 7 || usage.OutputTokens != 3 || usage.CacheReadTokens != 0 || usage.CacheWriteTokens != 0 {
		t.Fatalf("usage=%+v", usage)
	}
}

func TestParseRejectsConflictingDuplicateAndMissingParent(t *testing.T) {
	conflict := strings.Join([]string{
		`{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":1}}`,
		`{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":2}}`,
	}, "\n")
	if _, err := Parse(strings.NewReader(conflict)); err == nil {
		t.Fatal("conflicting duplicate passed")
	}

	orphan := `{"type":"span","traceId":"t","spanId":"s","parentSpanId":"missing","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":1}}`
	if _, err := Parse(strings.NewReader(orphan)); err == nil {
		t.Fatal("missing parent passed")
	}
}

func TestParseRejectsContentBearingTelemetry(t *testing.T) {
	input := `{"type":"span","traceId":"t","spanId":"s","name":"chat x","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","gen_ai.prompt":"secret","github.copilot.nano_aiu":1}}`
	if _, err := Parse(strings.NewReader(input)); err == nil {
		t.Fatal("content-bearing telemetry passed")
	}
}
