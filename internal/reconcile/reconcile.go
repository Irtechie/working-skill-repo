package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const (
	LedgerSchemaVersion = 1
	PlanSchemaVersion   = 1

	ModeDryRun = "dry-run"
	ModePlan   = "plan"

	EvidenceAvailable   = "available"
	EvidenceUnavailable = "unavailable"
	EvidenceStale       = "stale"
	EvidenceDisputed    = "disputed"

	PredicatePass        = "pass"
	PredicateFail        = "fail"
	PredicateUnavailable = "unavailable"
	PredicateStale       = "stale"

	ArtifactSession     = "session"
	ArtifactWorktree    = "worktree"
	ArtifactBranch      = "branch"
	ArtifactPullRequest = "pull-request"
	ArtifactQueueClaim  = "queue-claim"
	ArtifactReceipt     = "receipt"

	ClassificationPreserveActive = "preserve-active"
	ClassificationProtected      = "protected"
	ClassificationRoutineRetire  = "routine-retire"
	ClassificationSafeSupersede  = "safe-supersede"
	ClassificationSalvage        = "salvage"
	ClassificationQuarantine     = "quarantine"
	ClassificationHumanRequired  = "human-required"

	ActionMerge           = "merge"
	ActionPRClose         = "pr-close"
	ActionLocalRefRetire  = "local-ref-retire"
	ActionRemoteRefRetire = "remote-ref-retire"
	ActionWorktreeRetire  = "worktree-retire"
	ActionSessionRetire   = "session-record-retire"
	ActionSalvage         = "salvage"
	ActionProtectedWriter = "protected-writer"

	DedupIdenticalBlob          = "identical-blob-v1"
	DedupIdenticalTree          = "identical-tree-v1"
	DedupSameBasePatch          = "same-base-full-index-binary-patch-v1"
	DedupCommitAncestry         = "commit-ancestry-v1"
	DedupRemoteTopicContainment = "remote-topic-containment-v1"
	DedupRemoteDefaultAncestry  = "remote-default-ancestry-v1"
	DedupProviderMergeTree      = "provider-merge-identity+exact-tree-v1"
	DedupProviderMergePatch     = "provider-merge-identity+same-base-full-index-binary-patch-v1"
)

type Ledger struct {
	SchemaVersion int          `json:"schema_version"`
	Cutoff        time.Time    `json:"cutoff"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Actor         string       `json:"actor"`
	SessionID     string       `json:"session_id,omitempty"`
	Repositories  []Repository `json:"repositories"`
	Limitations   []string     `json:"limitations,omitempty"`
}

type Repository struct {
	ID                 string     `json:"id"`
	Root               string     `json:"root"`
	CommonDir          string     `json:"common_dir,omitempty"`
	PrimaryWorktree    string     `json:"primary_worktree,omitempty"`
	CurrentWorktree    string     `json:"current_worktree,omitempty"`
	CurrentBranch      string     `json:"current_branch,omitempty"`
	DefaultBranch      string     `json:"default_branch,omitempty"`
	DefaultBranchState string     `json:"default_branch_state"`
	Worktrees          []Worktree `json:"worktrees,omitempty"`
	Branches           []Branch   `json:"branches,omitempty"`
	Remotes            []Remote   `json:"remotes,omitempty"`
	// RemoteAuthority is nil when the caller never opted into a network probe.
	// That is a different statement from authority being unavailable, so it must
	// be absent rather than an empty object: omitempty does not suppress a
	// struct, and "state": "" reads as a failed probe rather than no probe.
	RemoteAuthority *RemoteAuthority `json:"remote_authority,omitempty"`
	QueueClaims     []QueueClaim     `json:"queue_claims,omitempty"`
	Receipts        []Receipt        `json:"receipts,omitempty"`
	Evidence        []Evidence       `json:"evidence"`
	Artifacts       []Artifact       `json:"artifacts"`
}

type Worktree struct {
	ID                string              `json:"id"`
	Path              string              `json:"path"`
	GitDir            string              `json:"git_dir,omitempty"`
	Branch            string              `json:"branch,omitempty"`
	Head              string              `json:"head,omitempty"`
	IsCurrent         bool                `json:"is_current"`
	IsPrimary         bool                `json:"is_primary"`
	IsDefault         bool                `json:"is_default"`
	Locked            bool                `json:"locked"`
	AdminRoundTrip    bool                `json:"admin_round_trip"`
	Dirt              Dirt                `json:"dirt"`
	ProtectionReasons []string            `json:"protection_reasons,omitempty"`
	Predicates        []PredicateEvidence `json:"predicates,omitempty"`
}

type Dirt struct {
	Tracked   []string `json:"tracked"`
	Untracked []string `json:"untracked"`
	Ignored   []string `json:"ignored"`
}

type Branch struct {
	Ref          string       `json:"ref"`
	SHA          string       `json:"sha"`
	UpdatedAt    time.Time    `json:"updated_at,omitempty"`
	CheckedOutAt string       `json:"checked_out_at,omitempty"`
	IsDefault    bool         `json:"is_default"`
	IsRemote     bool         `json:"is_remote"`
	Evidence     []Evidence   `json:"evidence,omitempty"`
	DedupProofs  []DedupProof `json:"dedup_proofs,omitempty"`
}

type Remote struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type QueueClaim struct {
	WorkID    string    `json:"work_id"`
	SessionID string    `json:"session_id,omitempty"`
	Status    string    `json:"status"`
	Branch    string    `json:"branch,omitempty"`
	Worktree  string    `json:"worktree,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Receipt struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Status    string    `json:"status,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Evidence struct {
	Name          string    `json:"name"`
	State         string    `json:"state"`
	Source        string    `json:"source"`
	ObservedAt    time.Time `json:"observed_at"`
	FreshUntil    time.Time `json:"fresh_until,omitempty"`
	Authoritative bool      `json:"authoritative"`
	Value         string    `json:"value,omitempty"`
	Limitation    string    `json:"limitation,omitempty"`
}

type Artifact struct {
	ID                string              `json:"id"`
	Kind              string              `json:"kind"`
	RepositoryID      string              `json:"repository_id"`
	Path              string              `json:"path,omitempty"`
	Ref               string              `json:"ref,omitempty"`
	SHA               string              `json:"sha,omitempty"`
	ObservedAt        time.Time           `json:"observed_at"`
	UpdatedAt         time.Time           `json:"updated_at,omitempty"`
	Dirt              Dirt                `json:"dirt"`
	ProtectionReasons []string            `json:"protection_reasons,omitempty"`
	Predicates        []PredicateEvidence `json:"predicates,omitempty"`
	DedupProofs       []DedupProof        `json:"dedup_proofs,omitempty"`
	UniqueWork        bool                `json:"unique_work"`
	SalvageSafe       bool                `json:"salvage_safe"`
	Ambiguity         string              `json:"ambiguity,omitempty"`
}

type PredicateEvidence struct {
	Name          string    `json:"name"`
	State         string    `json:"state"`
	Source        string    `json:"source"`
	ObservedAt    time.Time `json:"observed_at"`
	Authoritative bool      `json:"authoritative"`
	Limitation    string    `json:"limitation,omitempty"`
}

type DedupProof struct {
	Algorithm     string    `json:"algorithm"`
	Identity      string    `json:"identity"`
	MergeBase     string    `json:"merge_base,omitempty"`
	PathSet       []string  `json:"path_set,omitempty"`
	Source        string    `json:"source,omitempty"`
	ObservedAt    time.Time `json:"observed_at"`
	Authoritative bool      `json:"authoritative"`
}

func MarshalStable(value any) ([]byte, error) {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

func FingerprintLedger(ledger Ledger) (string, error) {
	normalized := normalizeLedger(ledger)
	content, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func normalizeLedger(ledger Ledger) Ledger {
	sort.Strings(ledger.Limitations)
	sort.Slice(ledger.Repositories, func(i, j int) bool {
		return ledger.Repositories[i].ID < ledger.Repositories[j].ID
	})
	for index := range ledger.Repositories {
		repository := &ledger.Repositories[index]
		sort.Slice(repository.Worktrees, func(i, j int) bool {
			return repository.Worktrees[i].Path < repository.Worktrees[j].Path
		})
		sort.Slice(repository.Branches, func(i, j int) bool {
			return repository.Branches[i].Ref < repository.Branches[j].Ref
		})
		sort.Slice(repository.Remotes, func(i, j int) bool {
			if repository.Remotes[i].Name == repository.Remotes[j].Name {
				return repository.Remotes[i].URL < repository.Remotes[j].URL
			}
			return repository.Remotes[i].Name < repository.Remotes[j].Name
		})
		sort.Slice(repository.QueueClaims, func(i, j int) bool {
			return repository.QueueClaims[i].WorkID < repository.QueueClaims[j].WorkID
		})
		sort.Slice(repository.Receipts, func(i, j int) bool {
			return repository.Receipts[i].ID < repository.Receipts[j].ID
		})
		sort.Slice(repository.Evidence, func(i, j int) bool {
			return repository.Evidence[i].Name < repository.Evidence[j].Name
		})
		sort.Slice(repository.Artifacts, func(i, j int) bool {
			return repository.Artifacts[i].ID < repository.Artifacts[j].ID
		})
		for artifactIndex := range repository.Artifacts {
			normalizeArtifact(&repository.Artifacts[artifactIndex])
		}
	}
	return ledger
}

func normalizeArtifact(artifact *Artifact) {
	sort.Strings(artifact.ProtectionReasons)
	sort.Strings(artifact.Dirt.Tracked)
	sort.Strings(artifact.Dirt.Untracked)
	sort.Strings(artifact.Dirt.Ignored)
	sort.Slice(artifact.Predicates, func(i, j int) bool {
		return artifact.Predicates[i].Name < artifact.Predicates[j].Name
	})
	sort.Slice(artifact.DedupProofs, func(i, j int) bool {
		if artifact.DedupProofs[i].Algorithm == artifact.DedupProofs[j].Algorithm {
			return artifact.DedupProofs[i].Identity < artifact.DedupProofs[j].Identity
		}
		return artifact.DedupProofs[i].Algorithm < artifact.DedupProofs[j].Algorithm
	})
}

func repositoryID(commonDir string) string {
	sum := sha256.Sum256([]byte(commonDir))
	return fmt.Sprintf("git-common-dir:%x", sum[:12])
}
