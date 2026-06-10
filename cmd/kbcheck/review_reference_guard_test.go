package main

import (
	"path/filepath"
	"testing"
)

func TestReviewReferenceGuardPassesForMatchingSharedPair(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
  "review_reference_guard": {
    "shared_pairs": [
      {"id":"shared","left":"left.md","right":"right.md","owner":"review","reason":"same doctrine"}
    ],
    "intentional_forks": [
      {"id":"fork","left":"fork-left.md","right":"fork-right.md","owner":"review","reason":"different lane"}
    ]
  }
}`)
	writeFile(t, filepath.Join(root, "left.md"), "same\n")
	writeFile(t, filepath.Join(root, "right.md"), "same\n")
	writeFile(t, filepath.Join(root, "fork-left.md"), "ce\n")
	writeFile(t, filepath.Join(root, "fork-right.md"), "kb\n")

	result, err := computeReviewReferenceGuard(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeReviewReferenceGuard returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected guard to pass, got issues %#v", result.Issues)
	}
}

func TestReviewReferenceGuardRejectsSharedDrift(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
  "review_reference_guard": {
    "shared_pairs": [
      {"id":"shared","left":"left.md","right":"right.md","owner":"review","reason":"same doctrine"}
    ]
  }
}`)
	writeFile(t, filepath.Join(root, "left.md"), "one\n")
	writeFile(t, filepath.Join(root, "right.md"), "two\n")

	result, err := computeReviewReferenceGuard(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeReviewReferenceGuard returned error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected guard to reject drift")
	}
	if result.Issues[0].Kind != "shared-reference-drift" {
		t.Fatalf("expected shared-reference-drift, got %#v", result.Issues)
	}
}

func TestReviewReferenceGuardRejectsUndocumentedFork(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
  "review_reference_guard": {
    "shared_pairs": [
      {"id":"shared","left":"left.md","right":"right.md","owner":"review","reason":"same doctrine"}
    ],
    "intentional_forks": [
      {"id":"fork","left":"fork-left.md","right":"fork-right.md"}
    ]
  }
}`)
	writeFile(t, filepath.Join(root, "left.md"), "same\n")
	writeFile(t, filepath.Join(root, "right.md"), "same\n")
	writeFile(t, filepath.Join(root, "fork-left.md"), "ce\n")
	writeFile(t, filepath.Join(root, "fork-right.md"), "kb\n")

	result, err := computeReviewReferenceGuard(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeReviewReferenceGuard returned error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected guard to reject undocumented fork")
	}
	kinds := map[string]bool{}
	for _, issue := range result.Issues {
		kinds[issue.Kind] = true
	}
	if !kinds["intentional-fork-missing-owner"] || !kinds["intentional-fork-missing-reason"] {
		t.Fatalf("expected missing owner/reason issues, got %#v", result.Issues)
	}
}

func TestReviewReferenceGuardSweepPassesForClassifiedCommonReference(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
  "review_reference_guard": {
    "reference_roots": [
      {"id":"ce","path":"ce/references"},
      {"id":"kb","path":"kb/references"}
    ],
    "shared_pairs": [
      {"id":"known","left":"ce/references/known.md","right":"kb/references/known.md","owner":"review","reason":"shared"}
    ]
  }
}`)
	writeFile(t, filepath.Join(root, "ce", "references", "known.md"), "same\n")
	writeFile(t, filepath.Join(root, "kb", "references", "known.md"), "same\n")

	result, err := computeReviewReferenceGuard(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeReviewReferenceGuard returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected classified common reference to pass, got %#v", result.Issues)
	}
}

func TestReviewReferenceGuardSweepRejectsUnclassifiedCommonReference(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
  "review_reference_guard": {
    "reference_roots": [
      {"id":"ce","path":"ce/references"},
      {"id":"kb","path":"kb/references"}
    ],
    "shared_pairs": [
      {"id":"known","left":"ce/references/known.md","right":"kb/references/known.md","owner":"review","reason":"shared"}
    ]
  }
}`)
	writeFile(t, filepath.Join(root, "ce", "references", "known.md"), "same\n")
	writeFile(t, filepath.Join(root, "kb", "references", "known.md"), "same\n")
	writeFile(t, filepath.Join(root, "ce", "references", "new-rubric.md"), "ce\n")
	writeFile(t, filepath.Join(root, "kb", "references", "new-rubric.md"), "kb\n")

	result, err := computeReviewReferenceGuard(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeReviewReferenceGuard returned error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected unclassified common reference to fail")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Kind == "unclassified-common-reference" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unclassified-common-reference issue, got %#v", result.Issues)
	}
}
