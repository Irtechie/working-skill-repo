package graphrouting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Packet struct {
	SchemaVersion int          `json:"schema_version"`
	PacketID      string       `json:"packet_id"`
	Repository    Repository   `json:"repository"`
	Seeds         Seeds        `json:"seeds"`
	Edges         []Edge       `json:"edges"`
	DirectImpact  []ImpactNode `json:"direct_impact"`
	ReverseImpact []ImpactNode `json:"reverse_impact"`
	Tests         []string     `json:"tests"`
	Docs          []string     `json:"docs"`
	Fallback      Fallback     `json:"fallback"`
	Budget        Budget       `json:"budget"`
	Limitations   []string     `json:"limitations"`
	GeneratedBy   string       `json:"generated_by"`
	GeneratedAt   string       `json:"generated_at"`
}

type Repository struct {
	Identity             string `json:"identity"`
	Root                 string `json:"root"`
	VCS                  string `json:"vcs"`
	Revision             string `json:"revision"`
	DirtyFingerprint     string `json:"dirty_fingerprint"`
	WorktreeFingerprint  string `json:"worktree_fingerprint"`
	Freshness            string `json:"freshness"`
	FreshnessExplanation string `json:"freshness_explanation"`
}

type Seeds struct {
	Files   []string `json:"files"`
	Symbols []Symbol `json:"symbols"`
}

type Symbol struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Kind string     `json:"kind"`
	Span SourceSpan `json:"span"`
}

type Edge struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Type        string       `json:"type"`
	Evidence    string       `json:"evidence"`
	Confidence  string       `json:"confidence"`
	Provenance  string       `json:"provenance"`
	Span        SourceSpan   `json:"span"`
	Limitations []string     `json:"limitations"`
	LoadBearing bool         `json:"load_bearing"`
	Metadata    EdgeMetadata `json:"metadata"`
}

type EdgeMetadata struct {
	Provider string `json:"provider,omitempty"`
	Query    string `json:"query,omitempty"`
}

type SourceSpan struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type ImpactNode struct {
	Path       string   `json:"path"`
	Symbol     string   `json:"symbol,omitempty"`
	Reason     string   `json:"reason"`
	Confidence string   `json:"confidence"`
	Edges      []string `json:"edges"`
}

type Fallback struct {
	Mode   string `json:"mode"`
	Reason string `json:"reason"`
}

type Budget struct {
	MaxEdges  int  `json:"max_edges"`
	MaxBytes  int  `json:"max_bytes"`
	Truncated bool `json:"truncated"`
}

type Result struct {
	OK      bool     `json:"ok"`
	Issues  []string `json:"issues,omitempty"`
	Summary Summary  `json:"summary"`
}

type Summary struct {
	PacketID       string         `json:"packet_id"`
	Revision       string         `json:"revision"`
	Freshness      string         `json:"freshness"`
	Fallback       string         `json:"fallback"`
	Edges          int            `json:"edges"`
	DirectImpact   int            `json:"direct_impact"`
	ReverseImpact  int            `json:"reverse_impact"`
	EvidenceCounts map[string]int `json:"evidence_counts"`
	Truncated      bool           `json:"truncated"`
}

func Load(path string) (Packet, error) {
	file, err := os.Open(path)
	if err != nil {
		return Packet{}, err
	}
	defer file.Close()
	return Decode(file)
}

func Decode(reader io.Reader) (Packet, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var packet Packet
	if err := decoder.Decode(&packet); err != nil {
		return Packet{}, err
	}
	return packet, nil
}

func Validate(packet Packet) Result {
	issues := []string{}
	require := func(ok bool, issue string) {
		if !ok {
			issues = append(issues, issue)
		}
	}

	require(packet.SchemaVersion == 1, "schema_version must be 1")
	require(nonBlank(packet.PacketID), "packet_id is required")
	require(nonBlank(packet.Repository.Identity), "repository.identity is required")
	require(nonBlank(packet.Repository.Root), "repository.root is required")
	require(nonBlank(packet.Repository.VCS), "repository.vcs is required")
	require(nonBlank(packet.Repository.Revision), "repository.revision is required")
	require(nonBlank(packet.Repository.WorktreeFingerprint), "repository.worktree_fingerprint is required")
	require(knownFreshness(packet.Repository.Freshness), "repository.freshness must be fresh, dirty, stale, or unknown")
	if packet.Repository.Freshness == "stale" {
		require(packet.Fallback.Mode != "" && packet.Fallback.Mode != "authoritative-provider", "stale packets must record an explicit non-authoritative fallback")
		require(nonBlank(packet.Repository.FreshnessExplanation), "stale packets require repository.freshness_explanation")
	}
	require(len(packet.Seeds.Files) > 0 || len(packet.Seeds.Symbols) > 0, "at least one seed file or symbol is required")
	for _, path := range packet.Seeds.Files {
		require(boundedRepoPath(path), "seed file must be repository-relative without traversal or globs: "+path)
	}
	for _, symbol := range packet.Seeds.Symbols {
		require(nonBlank(symbol.ID), "seed symbol id is required")
		if symbol.Span.Path != "" {
			require(validSpan(symbol.Span), "seed symbol span is invalid: "+symbol.ID)
		}
	}
	require(nonBlank(packet.Fallback.Mode), "fallback.mode is required")
	require(nonBlank(packet.Fallback.Reason), "fallback.reason is required")
	require(packet.Budget.MaxEdges > 0, "budget.max_edges must be positive")
	require(packet.Budget.MaxBytes > 0, "budget.max_bytes must be positive")
	require(len(packet.Limitations) > 0, "limitations must not be empty")

	for i, edge := range packet.Edges {
		prefix := fmt.Sprintf("edge[%d]", i)
		require(nonBlank(edge.From), prefix+".from is required")
		require(nonBlank(edge.To), prefix+".to is required")
		require(nonBlank(edge.Type), prefix+".type is required")
		require(knownEvidence(edge.Evidence), prefix+".evidence is invalid")
		require(knownConfidence(edge.Confidence), prefix+".confidence is invalid")
		require(nonBlank(edge.Provenance), prefix+".provenance is required")
		if edge.Evidence == "exact" || (edge.LoadBearing && (edge.Evidence == "exact" || edge.Evidence == "observed")) {
			require(validSpan(edge.Span), prefix+" load-bearing exact/observed edge requires source span")
		} else if edge.Span.Path != "" {
			require(validSpan(edge.Span), prefix+".span is invalid")
		}
		if edge.Evidence == "llm-inferred" {
			require(edge.Confidence != "exact", prefix+" llm-inferred evidence cannot claim exact confidence")
			require(len(edge.Limitations) > 0, prefix+" llm-inferred evidence requires limitations")
		}
	}
	for _, node := range append([]ImpactNode{}, append(packet.DirectImpact, packet.ReverseImpact...)...) {
		require(boundedRepoPath(node.Path), "impact path must be repository-relative without traversal or globs: "+node.Path)
		require(nonBlank(node.Reason), "impact reason is required for "+node.Path)
		require(knownConfidence(node.Confidence), "impact confidence is invalid for "+node.Path)
	}
	for _, path := range append([]string{}, append(packet.Tests, packet.Docs...)...) {
		require(boundedRepoPath(path), "test/doc path must be repository-relative without traversal or globs: "+path)
	}

	return Result{OK: len(issues) == 0, Issues: issues, Summary: Summarize(packet)}
}

func Summarize(packet Packet) Summary {
	counts := map[string]int{}
	for _, edge := range packet.Edges {
		counts[edge.Evidence]++
	}
	return Summary{
		PacketID:       packet.PacketID,
		Revision:       packet.Repository.Revision,
		Freshness:      packet.Repository.Freshness,
		Fallback:       packet.Fallback.Mode,
		Edges:          len(packet.Edges),
		DirectImpact:   len(packet.DirectImpact),
		ReverseImpact:  len(packet.ReverseImpact),
		EvidenceCounts: counts,
		Truncated:      packet.Budget.Truncated,
	}
}

func FormatSummary(summary Summary) string {
	kinds := make([]string, 0, len(summary.EvidenceCounts))
	for kind := range summary.EvidenceCounts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	parts := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		parts = append(parts, fmt.Sprintf("%s=%d", kind, summary.EvidenceCounts[kind]))
	}
	return fmt.Sprintf("graph-route: ok packet=%s revision=%s freshness=%s fallback=%s edges=%d direct=%d reverse=%d evidence=%s truncated=%t",
		summary.PacketID, summary.Revision, summary.Freshness, summary.Fallback, summary.Edges, summary.DirectImpact, summary.ReverseImpact, strings.Join(parts, ","), summary.Truncated)
}

func nonBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func knownFreshness(value string) bool {
	switch value {
	case "fresh", "dirty", "stale", "unknown":
		return true
	default:
		return false
	}
}

func knownEvidence(value string) bool {
	switch value {
	case "exact", "observed", "structural", "heuristic", "llm-inferred":
		return true
	default:
		return false
	}
}

func knownConfidence(value string) bool {
	switch value {
	case "exact", "high", "medium", "low", "unknown":
		return true
	default:
		return false
	}
}

func validSpan(span SourceSpan) bool {
	return boundedRepoPath(span.Path) && span.StartLine > 0 && span.EndLine >= span.StartLine
}

func boundedRepoPath(value string) bool {
	trimmed := strings.TrimSpace(value)
	clean := filepath.Clean(trimmed)
	return trimmed != "" &&
		clean != "." &&
		!filepath.IsAbs(clean) &&
		clean != ".." &&
		!strings.HasPrefix(clean, ".."+string(filepath.Separator)) &&
		!strings.ContainsAny(trimmed, "*?[]")
}
