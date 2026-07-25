package graphrouting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const SCIPFixtureProvider = "scip-fixture"

type SCIPExactSymbolProvider struct {
	IndexPath string
}

func (provider SCIPExactSymbolProvider) Name() string {
	return SCIPFixtureProvider
}

func (provider SCIPExactSymbolProvider) ResolveExactSymbols(request ProviderRequest) (ProviderLookup, error) {
	return ResolveSCIPExactSymbols(provider.IndexPath, request)
}

type SCIPExactSymbolIndex struct {
	SchemaVersion int              `json:"schema_version"`
	Provider      string           `json:"provider"`
	Repository    SCIPRepository   `json:"repository"`
	Symbols       []SCIPSymbol     `json:"symbols"`
	Relations     []SCIPRelation   `json:"relations"`
	Diagnostics   []SCIPDiagnostic `json:"diagnostics,omitempty"`
}

type SCIPRepository struct {
	Identity            string `json:"identity"`
	Root                string `json:"root"`
	Revision            string `json:"revision"`
	WorktreeFingerprint string `json:"worktree_fingerprint"`
}

type SCIPSymbol struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Kind string     `json:"kind"`
	Span SourceSpan `json:"span"`
}

type SCIPRelation struct {
	From string     `json:"from"`
	To   string     `json:"to"`
	Type string     `json:"type"`
	Span SourceSpan `json:"span"`
}

type SCIPDiagnostic struct {
	Severity string     `json:"severity"`
	Message  string     `json:"message"`
	Span     SourceSpan `json:"span"`
}

func LoadSCIPExactSymbolIndex(path string) (SCIPExactSymbolIndex, error) {
	file, err := os.Open(path)
	if err != nil {
		return SCIPExactSymbolIndex{}, err
	}
	defer file.Close()
	return DecodeSCIPExactSymbolIndex(file)
}

func DecodeSCIPExactSymbolIndex(reader io.Reader) (SCIPExactSymbolIndex, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var index SCIPExactSymbolIndex
	if err := decoder.Decode(&index); err != nil {
		return SCIPExactSymbolIndex{}, err
	}
	return index, nil
}

func ResolveSCIPExactSymbols(indexPath string, request ProviderRequest) (ProviderLookup, error) {
	request = normalizeProviderRequest(request)
	index, err := LoadSCIPExactSymbolIndex(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fallbackProviderLookup(request, SCIPFixtureProvider, "exact-symbol index unavailable; falling back to file-native lookup", err.Error()), nil
		}
		return ProviderLookup{}, err
	}
	return resolveSCIPExactSymbolsFromIndex(index, indexPath, request)
}

func resolveSCIPExactSymbolsFromIndex(index SCIPExactSymbolIndex, indexPath string, request ProviderRequest) (ProviderLookup, error) {
	request = normalizeProviderRequest(request)
	if err := validateSCIPIndexShape(index); err != nil {
		return ProviderLookup{}, err
	}
	if reason := staleSCIPIndexReason(index.Repository, request.Repository); reason != "" {
		return fallbackProviderLookup(request, SCIPFixtureProvider, reason), nil
	}

	symbols := make(map[string]SCIPSymbol, len(index.Symbols))
	for _, symbol := range index.Symbols {
		if !validSpan(symbol.Span) {
			return ProviderLookup{}, fmt.Errorf("scip symbol %s has invalid source span", symbol.ID)
		}
		symbols[symbol.ID] = symbol
	}
	seedID := request.Seed.ID
	if seedID == "" {
		seedID = findSCIPSymbolIDByName(symbols, request.Seed.Name)
	}
	if seedID == "" {
		return fallbackProviderLookup(request, SCIPFixtureProvider, "exact-symbol seed not present in index; falling back to file-native lookup"), nil
	}
	seed, ok := symbols[seedID]
	if !ok {
		return fallbackProviderLookup(request, SCIPFixtureProvider, "exact-symbol seed not present in index; falling back to file-native lookup"), nil
	}

	edges := []Edge{}
	direct := []ImpactNode{}
	reverse := []ImpactNode{}
	for _, relation := range index.Relations {
		if relation.From != seedID && relation.To != seedID {
			continue
		}
		if _, ok := symbols[relation.From]; !ok {
			return ProviderLookup{}, fmt.Errorf("scip relation %s->%s references missing from symbol", relation.From, relation.To)
		}
		toSymbol, ok := symbols[relation.To]
		if !ok {
			return ProviderLookup{}, fmt.Errorf("scip relation %s->%s references missing to symbol", relation.From, relation.To)
		}
		if !validSpan(relation.Span) {
			return ProviderLookup{}, fmt.Errorf("scip relation %s->%s has invalid source span", relation.From, relation.To)
		}
		edge := Edge{
			From:        relation.From,
			To:          relation.To,
			Type:        relation.Type,
			Evidence:    "exact",
			Confidence:  "exact",
			Provenance:  "scip-index:" + indexPath,
			Span:        relation.Span,
			LoadBearing: true,
			Metadata:    EdgeMetadata{Provider: SCIPFixtureProvider, Query: seedID},
		}
		edges = append(edges, edge)
		edgeID := relation.From + "->" + relation.To
		if relation.From == seedID {
			direct = append(direct, ImpactNode{
				Path:       toSymbol.Span.Path,
				Symbol:     relation.To,
				Reason:     "exact-symbol " + relation.Type + " from " + seedID,
				Confidence: "exact",
				Edges:      []string{edgeID},
			})
		}
		if relation.To == seedID {
			reverse = append(reverse, ImpactNode{
				Path:       symbols[relation.From].Span.Path,
				Symbol:     relation.From,
				Reason:     "exact-symbol " + relation.Type + " to " + seedID,
				Confidence: "exact",
				Edges:      []string{edgeID},
			})
		}
	}

	packet := Packet{
		SchemaVersion: 1,
		PacketID:      request.PacketID,
		Repository:    request.Repository,
		Seeds: Seeds{
			Symbols: []Symbol{{
				ID:   seed.ID,
				Name: seed.Name,
				Kind: seed.Kind,
				Span: seed.Span,
			}},
		},
		Edges:         edges,
		DirectImpact:  direct,
		ReverseImpact: reverse,
		Fallback:      Fallback{Mode: "authoritative-provider", Reason: "fresh exact-symbol index matched repository identity, revision, and worktree fingerprint"},
		Budget:        Budget{MaxEdges: request.MaxEdges, MaxBytes: request.MaxBytes},
		Limitations:   []string{"fixture-backed exact-symbol adapter; no language server lifecycle is managed"},
		GeneratedBy:   SCIPFixtureProvider,
	}
	return ProviderLookup{
		Packet:        packet,
		Provider:      SCIPFixtureProvider,
		Available:     true,
		Authoritative: true,
	}, nil
}

func validateSCIPIndexShape(index SCIPExactSymbolIndex) error {
	if index.SchemaVersion != 1 {
		return fmt.Errorf("scip index schema_version must be 1")
	}
	if strings.TrimSpace(index.Provider) == "" {
		return fmt.Errorf("scip index provider is required")
	}
	if strings.TrimSpace(index.Repository.Identity) == "" ||
		strings.TrimSpace(index.Repository.Revision) == "" ||
		strings.TrimSpace(index.Repository.WorktreeFingerprint) == "" {
		return fmt.Errorf("scip index repository identity, revision, and worktree_fingerprint are required")
	}
	return nil
}

func staleSCIPIndexReason(indexRepo SCIPRepository, requestRepo Repository) string {
	if indexRepo.Identity != requestRepo.Identity {
		return "exact-symbol index repository identity mismatch; falling back to file-native lookup"
	}
	if indexRepo.Revision != requestRepo.Revision {
		return "exact-symbol index revision is stale; falling back to file-native lookup"
	}
	if indexRepo.WorktreeFingerprint != requestRepo.WorktreeFingerprint {
		return "exact-symbol index worktree fingerprint mismatch; falling back to file-native lookup"
	}
	return ""
}

func findSCIPSymbolIDByName(symbols map[string]SCIPSymbol, name string) string {
	matches := []string{}
	for id, symbol := range symbols {
		if symbol.Name == name {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}
