package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDocConceptSurvivesRewordingButNotDeletion(t *testing.T) {
	t.Parallel()
	policy := docConcept("second local route")

	reworded := []string{
		"Do not select a second local route.",
		"No second local route is selected.",
		"do not select a second local route",
		"A **second local route** must never be selected.",
		"Never select a\nsecond local route when the parent returns.",
	}
	for _, text := range reworded {
		if what, absent := policy.missing(normalizeContractText(text)); absent {
			t.Errorf("reworded policy should still match %q: missing %s", text, what)
		}
	}

	deleted := []string{
		"The parent returns on first local failure.",
		"Select one route and stop.",
	}
	for _, text := range deleted {
		if _, absent := policy.missing(normalizeContractText(text)); !absent {
			t.Errorf("deleted policy should fail, but %q matched", text)
		}
	}
}

func TestDocConceptRequiresEveryTerm(t *testing.T) {
	t.Parallel()
	matcher := docConcept("agent-owned decision", "review work")

	if _, absent := matcher.missing(normalizeContractText("Never turn an agent-owned decision into review work for the user.")); absent {
		t.Fatal("full policy should match")
	}
	if _, absent := matcher.missing(normalizeContractText("An agent-owned decision belongs to the agent.")); !absent {
		t.Fatal("partial policy must not satisfy a multi-term concept")
	}
}

func TestDocAnchorIgnoresCaseWrappingAndEmphasis(t *testing.T) {
	t.Parallel()
	heading := docAnchor("### Step 2.6: Orchestrator Ownership Decision (DDR)")
	if _, absent := heading.missing(normalizeContractText("### Step 2.6: Orchestrator Ownership Decision (DDR)")); absent {
		t.Fatal("exact heading should match")
	}
	if _, absent := heading.missing(normalizeContractText("###   step 2.6:  orchestrator ownership\ndecision DDR")); absent {
		t.Fatal("anchor should tolerate case, wrapping, and punctuation drift")
	}
	if _, absent := heading.missing(normalizeContractText("### Step 2.6: Orchestrator Ownership")); !absent {
		t.Fatal("truncated heading must fail")
	}

	label := docAnchor("**Native host delegation:**")
	if _, absent := label.missing(normalizeContractText("Native host delegation:")); absent {
		t.Fatal("anchor should tolerate markdown emphasis changes")
	}

	// A colon carries structure, so a short label must not collapse into a common word.
	shortLabel := docAnchor("Blocked:")
	if _, absent := shortLabel.missing(normalizeContractText("Blocked: the remote default is protected.")); absent {
		t.Fatal("labelled anchor should match")
	}
	if _, absent := shortLabel.missing(normalizeContractText("This slice is blocked on review.")); !absent {
		t.Fatal("a bare word must not satisfy a labelled anchor")
	}
}

func TestDocContractDetectsDeletedPolicyButToleratesRewrite(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	required := map[string][]contractMatcher{
		"SKILL.md": {
			docAnchor("## Delivery Boundary"),
			docAnchor("terminal-cleanup --action sweep"),
			docConcept("second local route"),
			docConcept("never merges"),
		},
	}

	original := "# Skill\n\n## Delivery Boundary\n\nRun `terminal-cleanup --action sweep` first.\nDo not select a second local route.\nThis skill never merges a default branch.\n"
	writeFile(t, filepath.Join(root, "SKILL.md"), original)
	problems, err := docContractProblems(root, required)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("baseline documentation should satisfy the contract, got %v", problems)
	}

	rewritten := "# Skill\n\n## Delivery Boundary\n\nStart by running `terminal-cleanup --action sweep`.\n**No second local route** is ever selected, because a\nsecond attempt hides failure. This skill never merges anything.\n"
	writeFile(t, filepath.Join(root, "SKILL.md"), rewritten)
	problems, err = docContractProblems(root, required)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("rewording must not break the contract, got %v", problems)
	}

	deleted := "# Skill\n\n## Delivery Boundary\n\nRun `terminal-cleanup --action sweep` first.\nThis skill never merges a default branch.\n"
	writeFile(t, filepath.Join(root, "SKILL.md"), deleted)
	problems, err = docContractProblems(root, required)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "second local route") {
		t.Fatalf("deleting the policy must fail the contract, got %v", problems)
	}

	renamedHeading := "# Skill\n\n## Delivery\n\nRun `terminal-cleanup --action sweep` first.\nDo not select a second local route.\nThis skill never merges a default branch.\n"
	writeFile(t, filepath.Join(root, "SKILL.md"), renamedHeading)
	problems, err = docContractProblems(root, required)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "## Delivery Boundary") {
		t.Fatalf("removing a structural anchor must fail the contract, got %v", problems)
	}
}

func TestDocConceptTreatsProhibitionSpellingsAsEquivalent(t *testing.T) {
	t.Parallel()
	prohibition := docConcept("not ask the user whether to run browser proof")

	for _, text := range []string{
		"Do not ask the user whether to run browser proof.",
		"Never ask the user whether to run browser proof.",
		"You must not ask the user whether to run browser proof.",
		"Don't ask the user whether to run browser proof.",
	} {
		if what, absent := prohibition.missing(normalizeContractText(text)); absent {
			t.Errorf("prohibition should match %q: missing %s", text, what)
		}
	}

	if _, absent := prohibition.missing(normalizeContractText("Ask the user whether to run browser proof.")); !absent {
		t.Fatal("dropping the prohibition must fail the contract")
	}

	// A concept may end on the negation itself, which only normalizes if the
	// replacement is not anchored to a following space.
	trailing := docConcept("Diff size and path grouping cannot")
	if what, absent := trailing.missing(normalizeContractText("Diff size and path grouping cannot substitute for impact.")); absent {
		t.Fatalf("trailing negation should normalize: missing %s", what)
	}
	if what, absent := trailing.missing(normalizeContractText("Diff size and path grouping never substitute for impact.")); absent {
		t.Fatalf("trailing negation should match an equivalent spelling: missing %s", what)
	}
}

func TestMustMutateContractFailsWhenTargetIsGone(t *testing.T) {
	t.Parallel()
	mutated := mustMutateContract(t, "Do not select a second local route. A second local route hides failure.", "second local route", "second remote route")
	if strings.Contains(strings.ToLower(mutated), "second local route") {
		t.Fatalf("every occurrence must be replaced, got %q", mutated)
	}
}

func TestDocAnchorPreservesCommandAndKeyPrecision(t *testing.T) {
	t.Parallel()
	command := docAnchor("terminal-cleanup --action sweep")
	if _, absent := command.missing(normalizeContractText("Run `terminal-cleanup --action sweep` first.")); absent {
		t.Fatal("command anchor should match inside a code span")
	}
	if _, absent := command.missing(normalizeContractText("Run `terminal-cleanup --action register` first.")); !absent {
		t.Fatal("a different subcommand must fail")
	}

	key := docAnchor("affects_normal_work: false")
	if _, absent := key.missing(normalizeContractText("affects_normal_work: false")); absent {
		t.Fatal("config key anchor should match")
	}
	if _, absent := key.missing(normalizeContractText("affects_normal_work: true")); !absent {
		t.Fatal("flipped config value must fail")
	}
}
