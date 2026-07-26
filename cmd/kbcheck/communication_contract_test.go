package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLowCognitiveBurdenCommunicationContract(t *testing.T) {
	root := communicationContractRepoRoot(t)
	required := map[string][]string{
		"AGENTS.md": {
			"Optimize for comprehension and decision effort, not the fewest words.",
			"hard response required",
			"soft preference",
			"no response needed",
			"what is blocked",
			"Never turn an agent-owned decision into review work for the user.",
		},
		".github/skills/kb-compact/SKILL.md": {
			"talk to me like a person",
			"low cognitive burden",
			"## Response Responsibility",
			"Only the user can authorize, supply, or decide it",
			"State the default and continue unless overridden",
			"Use plain human language",
		},
		".github/skills/kb-gate/SKILL.md": {
			"Response required:",
			"Why you:",
			"Blocked:",
			"Recommendation:",
			"Do not ask the user to choose among agent-owned fixes.",
		},
		".github/skills/kb-ship/SKILL.md": {
			"What changed / Why it matters",
			"Needs reviewer attention",
			"Handled by agent",
			"None — no reviewer decision required",
			"low-burden first screen",
			"companion design, research,",
		},
		".github/skills/kb-executive-brief/SKILL.md": {
			"low-cognitive-burden executive brief",
			"hard_response_required",
			"soft_preference",
			"no_response_needed",
			"at least three meaningful nodes and two edges",
			"go run ./cmd/kbbrief",
		},
		"docs/context/operations/low-burden-review-artifacts.md": {
			"The goal is not fewer words.",
			"Review Responsibility",
			"Exact hard questions",
			"executive first screen",
			"mental alignment",
			"https://www.humanlayer.dev/blog/advanced-context-engineering",
			"generated Markdown is a projection",
		},
		"README.md": {
			"Optimize for comprehension and decision effort, not the fewest words.",
			"Plannotator's bro skill",
			"hard response only the user can",
			"low-burden PR first screen",
			"go run ./cmd/kbbrief",
		},
	}

	for path, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(content)
		for _, phrase := range phrases {
			if !strings.Contains(text, phrase) {
				t.Errorf("%s missing communication contract %q", path, phrase)
			}
		}
	}
}

func communicationContractRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}
