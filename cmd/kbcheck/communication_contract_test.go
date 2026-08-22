package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLowCognitiveBurdenCommunicationContract(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	requireDocContract(t, root, "communication contract", map[string][]contractMatcher{
		"AGENTS.md": {
			docAnchor("hard response required"),
			docAnchor("soft preference"),
			docAnchor("no response needed"),
			docConcept("comprehension and decision effort"),
			docConcept("what is blocked"),
			docConcept("agent-owned decision into review work"),
			docConcept("table for repeated-field comparison"),
			docConcept("visual must remove mental reconstruction"),
		},
		".github/skills/kb-cognitive/SKILL.md": {
			docAnchor("## Format Selection"),
			docAnchor("## Response Responsibility"),
			docAnchor("references/response-patterns.md"),
			docConcept("talk to me like a person"),
			docConcept("low cognitive burden"),
			docConcept("smallest format"),
			docConcept("visual must earn its space"),
			docConcept("authorize, supply, or decide it"),
			docConcept("default and continue unless overridden"),
			docConcept("plain human language"),
		},
		".github/skills/kb-cognitive/references/response-patterns.md": {
			docAnchor("## 1. Simple Outcome: Use Plain Language"),
			docAnchor("## 3. Hard Decision: Keep the Ask Together"),
			docAnchor("## 4. Repeated Fields: Use a Table"),
			docAnchor("## 5. Workflow or Dependency: Use One Diagram"),
			docAnchor("```mermaid"),
			docConcept("Order by effect on the person or application"),
		},
		".github/skills/kb-gate/SKILL.md": {
			docAnchor("Response required:"),
			docAnchor("Why you:"),
			docAnchor("Blocked:"),
			docAnchor("Recommendation:"),
			docConcept("Do not ask the user to choose among agent-owned fixes"),
			docConcept("execution control, not a gate result"),
			docConcept("cheapest owning sensor"),
			docConcept("unrelated ready work continues"),
		},
		".github/skills/kb-work/SKILL.md": {
			docConcept("explicit pause is not technical terminal proof"),
			docConcept("Continue unrelated runnable slices"),
			docConcept("Slice execution fails with progress possible"),
		},
		".github/skills/kb-goal/SKILL.md": {
			docAnchor("## Stop Protocol"),
			docAnchor("To resume: /kb-goal <verbatim Objective text from the ledger>"),
			docConcept("preemptive control signal, never a suggestion"),
			docConcept("Do not start another goal-work"),
			docConcept("End this goal permanently?"),
			docConcept("Parent stop authority is non-overridable"),
			docConcept("Reject queued and late work"),
			docConcept("coerce pause into a false blocker"),
			docConcept("Before marking or repeating a blocker"),
		},
		".github/skills/kb-complete/SKILL.md": {
			docAnchor("Implementation: complete|incomplete"),
			docConcept("roll a narrow gate up"),
			docConcept("optional capability/platform blocked"),
		},
		".github/skills/kb-handoff/SKILL.md": {
			docConcept("before copying any blocker"),
			docConcept("`⏸ paused` only when the user requested"),
		},
		".github/skills/kb-qa/SKILL.md": {
			docConcept("agent/tool dependency"),
			docConcept("agent-owned test work"),
		},
		".github/skills/kb-finalize/SKILL.md": {
			docAnchor("`Delivery: blocked`"),
			docConcept("reporting any blocked or human-required item"),
		},
		".github/skills/kb-ship/SKILL.md": {
			docAnchor("What changed / Why it matters"),
			docAnchor("Needs reviewer attention"),
			docAnchor("Handled by agent"),
			docConcept("no reviewer decision required"),
			docConcept("low-burden first screen"),
			docConcept("companion design, research"),
		},
		".github/skills/kb-executive-brief/SKILL.md": {
			docAnchor("hard_response_required"),
			docAnchor("soft_preference"),
			docAnchor("no_response_needed"),
			docAnchor("go run ./cmd/kbbrief"),
			docConcept("low-cognitive-burden executive brief"),
			docConcept("at least three meaningful nodes and two edges"),
		},
		"docs/context/operations/low-burden-review-artifacts.md": {
			docAnchor("https://www.humanlayer.dev/blog/advanced-context-engineering"),
			docAnchor("Review Responsibility"),
			docConcept("goal is not fewer words"),
			docConcept("Exact hard questions"),
			docConcept("executive first screen"),
			docConcept("mental alignment"),
			docConcept("generated Markdown is a projection"),
		},
		"README.md": {
			docAnchor("go run ./cmd/kbbrief"),
			docConcept("comprehension and decision effort"),
			docConcept("Plannotator's bro skill"),
			docConcept("hard response only the user can"),
			docConcept("low-burden PR first screen"),
			docConcept("user pause stops work immediately"),
			docConcept("after a stop signal, the goal does not dispatch"),
			docConcept("override the parent freeze"),
			docConcept("Every blocker is rechecked before it is repeated"),
		},
		"docs/context/architecture/kb-workflow.md": {
			docAnchor("### Blocker Lifecycle"),
			docConcept("`paused` is rejected as a gate"),
			docConcept("Only dependent work stops"),
			docConcept("User interruption has higher priority"),
			docConcept("Parent pause and stop authority is non-overridable"),
			docConcept("`To resume: /kb-goal <objective>` command"),
		},
	})
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
