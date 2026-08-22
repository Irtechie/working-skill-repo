package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeliveryStateProjectionIsDecisionFirst(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "delivery.json")
	writeFile(t, path, `{
  "schema_version": 1,
  "state": "awaiting-review",
  "recommendation": "Wait for the required review, then merge this PR.",
  "proof": ["go test ./cmd/kbcheck"],
  "blocker": "one required approval",
  "pull_request_url": "https://example.test/pr/7",
  "screenshots": ["artifacts/settings.png"]
}`)
	var out, errOut bytes.Buffer
	if code := run([]string{"delivery-state", "--delivery-receipt", path}, &out, &errOut); code != 0 {
		t.Fatalf("delivery-state failed code=%d stderr=%s", code, errOut.String())
	}
	got := out.String()
	for _, want := range []string{
		"Delivery: awaiting-review", "Next: Wait for the required review, then merge this PR.",
		"Proof: go test ./cmd/kbcheck", "Blocker: one required approval",
		"PR: https://example.test/pr/7", "Screenshots: artifacts/settings.png",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("projection missing %q: %s", want, got)
		}
	}
}

func TestDeliveryStateRejectsInvalidLifecycleClaims(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct{ body, want string }{
		"open-pr-without-url":    {`{"schema_version":1,"state":"awaiting-review","recommendation":"wait","proof":["test"]}`, "requires pull_request_url"},
		"integrated-without-sha": {`{"schema_version":1,"state":"delivery-integrated","recommendation":"done","proof":["test"]}`, "requires merged_sha"},
		"unknown-state":          {`{"schema_version":1,"state":"done","recommendation":"done","proof":["test"]}`, "state must be"},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "delivery.json")
			writeFile(t, path, test.body)
			var out, errOut bytes.Buffer
			if code := run([]string{"delivery-state", "--delivery-receipt", path}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), test.want) {
				t.Fatalf("want invalid lifecycle rejection %q, code=%d stdout=%s stderr=%s", test.want, code, out.String(), errOut.String())
			}
		})
	}
}
