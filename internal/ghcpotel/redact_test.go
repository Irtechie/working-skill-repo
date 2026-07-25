package ghcpotel

import (
	"bytes"
	"strings"
	"testing"
)

func TestRedactSchemaRemovesValuesAndRejectsSensitiveKeys(t *testing.T) {
	raw := `{"type":"span","traceId":"user-trace","spanId":"span","name":"chat model","attributes":{"gen_ai.response.model":"model","gen_ai.usage.input_tokens":2,"user.name":"mark","absolute.path":"E:\\secret"}}`
	var output bytes.Buffer
	if err := RedactSchema(strings.NewReader(raw), &output); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, forbidden := range []string{"user-trace", "chat model", "mark", `E:\\secret`} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("redacted schema contains %q: %s", forbidden, got)
		}
	}
	if !strings.Contains(got, `"gen_ai.response.model":"string"`) ||
		!strings.Contains(got, `"gen_ai.usage.input_tokens":"number"`) {
		t.Fatalf("schema types missing: %s", got)
	}
}

func TestRedactSchemaRejectsPromptAndToolPayloadKeys(t *testing.T) {
	for _, key := range []string{"gen_ai.prompt", "tool.payload", "message.content"} {
		input := `{"type":"span","attributes":{"` + key + `":"secret"}}`
		if err := RedactSchema(strings.NewReader(input), &bytes.Buffer{}); err == nil {
			t.Fatalf("sensitive key %q passed", key)
		}
	}
}
