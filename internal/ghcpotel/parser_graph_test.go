package ghcpotel

import (
	"strings"
	"testing"
)

func TestParseCountsOnlyLeafChatSpans(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"span","traceId":"t","spanId":"root","name":"invoke_agent","attributes":{}}`,
		`{"type":"span","traceId":"t","spanId":"parent","parentSpanId":"root","name":"chat aggregate","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","github.copilot.nano_aiu":10}}`,
		`{"type":"span","traceId":"t","spanId":"leaf","parentSpanId":"parent","name":"chat leaf","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","gen_ai.usage.input_tokens":0,"gen_ai.usage.output_tokens":0,"github.copilot.nano_aiu":3}}`,
	}, "\n")
	usage, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if usage.Calls != 1 || usage.AIUNano != 3 {
		t.Fatalf("usage=%+v", usage)
	}
}

func TestParseRejectsEmptyAndCyclicExports(t *testing.T) {
	if _, err := Parse(strings.NewReader("")); err == nil {
		t.Fatal("empty export passed")
	}
	cyclic := strings.Join([]string{
		`{"type":"span","traceId":"t","spanId":"a","parentSpanId":"b","name":"chat a","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","github.copilot.nano_aiu":1}}`,
		`{"type":"span","traceId":"t","spanId":"b","parentSpanId":"a","name":"chat b","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","github.copilot.nano_aiu":1}}`,
	}, "\n")
	if _, err := Parse(strings.NewReader(cyclic)); err == nil {
		t.Fatal("cyclic export passed")
	}
}

func TestParseRejectsWholeExportBeyondLimit(t *testing.T) {
	row := `{"type":"span","traceId":"t","spanId":"s","name":"chat leaf","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","github.copilot.nano_aiu":1}}` + "\n"
	input := strings.Repeat(row, maxExportBytes/len(row)+2)
	if _, err := Parse(strings.NewReader(input)); err == nil {
		t.Fatal("oversized export passed")
	}

}

func TestParseAcceptsIntegralDecimalNanoAIU(t *testing.T) {
	input := `{"type":"span","traceId":"t","spanId":"s","name":"chat model","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"model","gen_ai.usage.input_tokens":1,"gen_ai.usage.output_tokens":1,"github.copilot.nano_aiu":15370000000.0}}`
	usage, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if usage.AIUNano != 15_370_000_000 || usage.Calls != 1 {
		t.Fatalf("usage=%+v", usage)
	}
}
