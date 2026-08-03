package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateApprovedCatalogPreservesMixedEvidenceAndPromotionException(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "approved-skills.json")
	existing := `{
  "schemaVersion": 1,
  "skills": [{
    "id": "existing-skill",
    "status": "approved",
    "source": {"type": "personal", "path": "skills/existing-skill"},
    "marketplacePath": "skills/existing-skill",
    "sha256": "abc",
    "approvedBy": "owner",
    "approvedAt": "2026-07-20",
    "approvalReason": "proven",
    "promotionException": "owner-approved",
    "evidence": {
      "observations": 8,
      "confidence": 0.97,
      "successfulReuses": 3,
      "proofCommands": ["verify existing"]
    }
  }]
}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	entry := approvedCatalogEntry{
		ID: "new-skill", Status: "approved",
		Source:          map[string]string{"type": "local-reviewed", "path": "skills/new-skill"},
		MarketplacePath: "skills/new-skill", SHA256: "def", ApprovedBy: "owner",
		ApprovedAt: "2026-07-21", ApprovalReason: "approved",
		Evidence: map[string]interface{}{"proofCommands": []string{"verify new"}},
	}
	if err := updateApprovedCatalog(path, entry); err != nil {
		t.Fatalf("updateApprovedCatalog returned error: %v", err)
	}

	var catalog approvedCatalog
	if err := readJSONFile(path, &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Skills) != 2 {
		t.Fatalf("skills=%d, want 2", len(catalog.Skills))
	}
	var got approvedCatalogEntry
	for _, skill := range catalog.Skills {
		if skill.ID == "existing-skill" {
			got = skill
		}
	}
	if got.PromotionException != "owner-approved" {
		t.Fatalf("promotionException=%q", got.PromotionException)
	}
	if got.Evidence["observations"] != float64(8) || got.Evidence["confidence"] != 0.97 || got.Evidence["successfulReuses"] != float64(3) {
		t.Fatalf("mixed evidence was not preserved: %#v", got.Evidence)
	}
}
