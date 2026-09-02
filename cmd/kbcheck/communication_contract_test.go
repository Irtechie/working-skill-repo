package main

import (
	"os"
	"path/filepath"
	"strings"
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
			docAnchor("**Done**"),
			docAnchor("**Agent continues**"),
			docAnchor("**You need to decide**"),
			docConcept("comprehension and decision effort"),
			docConcept("user-facing return boundaries"),
			docConcept("internal reasoning, tool calls, and subagent exchanges"),
			docConcept("practical meaning first"),
			docConcept("what is blocked"),
			docConcept("agent-owned decision into review work"),
			docConcept("table for repeated-field comparison"),
			docConcept("real screenshot"),
			docConcept("interactive-workflow-workbench-light"),
			docConcept("visual must remove mental reconstruction"),
			docConcept("optional visual capability is unavailable"),
		},
		".github/skills/kb-cognitive/SKILL.md": {
			docAnchor("## Return Boundary"),
			docAnchor("## Format Selection"),
			docAnchor("## Response Responsibility"),
			docAnchor("references/response-patterns.md"),
			docConcept("talk to me like a person"),
			docConcept("low cognitive burden"),
			docConcept("smallest format"),
			docConcept("visual must earn its space"),
			docConcept("authorize, supply, or decide it"),
			docConcept("default and continue unless overridden"),
			docConcept("practical meaning first"),
			docConcept("internal reasoning"),
			docConcept("interactive-workflow-workbench-light"),
			docAnchor("`interactive-workflow-workbench`"),
			docConcept("fall back to the best static format"),
			docConcept("never triggers merely because a PR exists"),
			docConcept("plain human language"),
		},
		".github/skills/kb-cognitive/references/response-patterns.md": {
			docAnchor("## 1. Simple Outcome: Use Plain Language"),
			docAnchor("## 3. Hard Decision: Keep the Ask Together"),
			docAnchor("## 4. Repeated Fields: Use a Table"),
			docAnchor("## 5. Workflow or Dependency: Use One Diagram"),
			docAnchor("## 7. Return Boundary: Make Control Ownership Explicit"),
			docAnchor("## 8. Visual Escalation: Stop at the Cheapest Useful Format"),
			docAnchor("```mermaid"),
			docConcept("Order by effect on the person or application"),
			docConcept("Agent continues"),
			docConcept("interactive-workflow-workbench-light"),
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
			docAnchor("Control: Done | Agent continues | You need to decide"),
			docAnchor("Implementation: complete|incomplete"),
			docConcept("roll a narrow gate up"),
			docConcept("optional capability/platform blocked"),
			docConcept("Wait for review is the safe default"),
			docConcept("Do not ask a terminal integration question"),
			docConcept("safe repair or retry remains"),
			docConcept("Do not relabel the underlying technical failure as human-owned"),
			docConcept("continuing requires a human disposition decision"),
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
			docAnchor("### Presentation ladder"),
			docAnchor("What changed / Why it matters"),
			docAnchor("Needs reviewer attention"),
			docAnchor("Handled by agent"),
			docConcept("no reviewer decision required"),
			docConcept("low-burden first screen"),
			docConcept("companion design, research"),
			docConcept("real screenshot"),
			docConcept("interactive-workflow-workbench-light"),
			docAnchor("`interactive-workflow-workbench`"),
			docConcept("optional visual capability is unavailable"),
			docConcept("Do not generate HTML merely because a PR exists"),
		},
		".github/skills/kb-executive-brief/SKILL.md": {
			docAnchor("hard_response_required"),
			docAnchor("soft_preference"),
			docAnchor("no_response_needed"),
			docAnchor("go run ./cmd/kbbrief"),
			docConcept("low-cognitive-burden executive brief"),
			docConcept("at least three meaningful nodes and two edges"),
			docConcept("interactive-workflow-workbench-light"),
			docAnchor("`interactive-workflow-workbench`"),
		},
		"docs/context/operations/low-burden-review-artifacts.md": {
			docAnchor("https://www.humanlayer.dev/blog/advanced-context-engineering"),
			docAnchor("Review Responsibility"),
			docConcept("goal is not fewer words"),
			docConcept("Exact hard questions"),
			docConcept("executive first screen"),
			docConcept("mental alignment"),
			docConcept("generated Markdown is a projection"),
			docConcept("real screenshot"),
			docConcept("interactive-workflow-workbench-light"),
			docConcept("optional visual"),
		},
		"docs/context/architecture/kb-workflow.md": {
			docAnchor("### Blocker Lifecycle"),
			docAnchor("## Cognitive Return Boundary"),
			docConcept("`paused` is rejected as a gate"),
			docConcept("Only dependent work stops"),
			docConcept("User interruption has higher priority"),
			docConcept("Parent pause and stop authority is non-overridable"),
			docConcept("`To resume: /kb-goal <objective>` command"),
			docConcept("interactive-workflow-workbench-light"),
			docConcept("internal reasoning and tool exchange"),
		},
	})

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(agents), "`Done | Next | Blocked`") {
		t.Error("AGENTS.md retains legacy control-state vocabulary")
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
