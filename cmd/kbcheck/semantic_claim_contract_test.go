package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemanticClaimRepoContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	required := map[string][]string{
		".github/skills/kb-start/scripts/work_queue.ps1": {
			"semantic_resources", "active_owner", "global_authority", "resource_conflict",
		},
		".github/skills/kb-start/SKILL.md": {
			"kbreconcile", "declared semantic resources", "authoritative adapter capability",
		},
		".github/skills/kb-finalize/SKILL.md": {
			"register completion state and evidence", "never deletes the current worktree",
		},
		".github/skills/kb-complete/SKILL.md": {
			"`local-durable`", "`awaiting-review`", "`delivery-integrated`",
			"delivery, physical cleanup, ref retirement, and host session retirement",
		},
	}
	for relative, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.ToLower(strings.Join(strings.Fields(string(content)), " "))
		for _, phrase := range phrases {
			if !strings.Contains(normalized, strings.ToLower(strings.Join(strings.Fields(phrase), " "))) {
				t.Errorf("%s weakens semantic claim contract: missing %q", relative, phrase)
			}
		}
	}
}

func TestDeliveryChainLifecycleStatesRemainSeparate(t *testing.T) {
	for _, test := range []struct {
		status, mode string
		want         bool
	}{
		{"local-durable", "local", true},
		{"awaiting-review", "pr", true},
		{"delivery-integrated", "direct", true},
		{"awaiting-review", "direct", false},
		{"delivery-integrated", "pr", false},
	} {
		if got := terminalClaimMatchesDelivery(test.status, test.mode); got != test.want {
			t.Errorf("status=%s mode=%s got=%t want=%t", test.status, test.mode, got, test.want)
		}
	}
}
