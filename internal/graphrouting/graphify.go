package graphrouting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const GraphifyProvider = "graphify"

type GraphifyStructuralProvider struct {
	GraphPath string
	Recipe    TraversalRecipe
}

func (provider GraphifyStructuralProvider) Name() string {
	return GraphifyProvider
}

func (provider GraphifyStructuralProvider) ResolveStructuralEdges(request ProviderRequest) (ProviderLookup, error) {
	return ResolveGraphifyStructuralEdges(provider.GraphPath, provider.Recipe, request)
}

type GraphifyGraph struct {
	SchemaVersion       int                 `json:"schema_version"`
	Provider            string              `json:"provider"`
	Repository          GraphifyRepository  `json:"repository"`
	MultigraphCollapsed bool                `json:"multigraph_collapsed"`
	Nodes               []GraphifyNode      `json:"nodes"`
	Edges               []GraphifyGraphEdge `json:"edges"`
}

type GraphifyRepository struct {
	Identity            string `json:"identity"`
	Revision            string `json:"revision"`
	WorktreeFingerprint string `json:"worktree_fingerprint"`
}

type GraphifyNode struct {
	ID     string     `json:"id"`
	Path   string     `json:"path"`
	Symbol string     `json:"symbol"`
	Span   SourceSpan `json:"span"`
}

type GraphifyGraphEdge struct {
	From     string     `json:"from"`
	To       string     `json:"to"`
	Type     string     `json:"type"`
	Inferred bool       `json:"inferred"`
	Span     SourceSpan `json:"span"`
}

func LoadGraphifyGraph(path string) (GraphifyGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		return GraphifyGraph{}, err
	}
	defer file.Close()
	return DecodeGraphifyGraph(file)
}

func DecodeGraphifyGraph(reader io.Reader) (GraphifyGraph, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var graph GraphifyGraph
	if err := decoder.Decode(&graph); err != nil {
		return GraphifyGraph{}, err
	}
	return graph, nil
}

func ResolveGraphifyStructuralEdges(path string, recipe TraversalRecipe, request ProviderRequest) (ProviderLookup, error) {
	request = normalizeProviderRequest(request)
	graph, err := LoadGraphifyGraph(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fallbackProviderLookup(request, GraphifyProvider, "graphify output unavailable; falling back to file-native lookup", err.Error()), nil
		}
		return ProviderLookup{}, err
	}
	return resolveGraphifyStructuralEdgesFromGraph(graph, path, recipe, request)
}

func resolveGraphifyStructuralEdgesFromGraph(graph GraphifyGraph, path string, recipe TraversalRecipe, request ProviderRequest) (ProviderLookup, error) {
	request = normalizeProviderRequest(request)
	if err := validateGraphifyShape(graph); err != nil {
		return ProviderLookup{}, err
	}
	if reason := staleGraphifyReason(graph.Repository, request.Repository); reason != "" {
		return fallbackProviderLookup(request, GraphifyProvider, reason), nil
	}
	nodes := map[string]GraphifyNode{}
	for _, node := range graph.Nodes {
		if !validSpan(node.Span) {
			return ProviderLookup{}, fmt.Errorf("graphify node %s has invalid source span", node.ID)
		}
		nodes[node.ID] = node
	}
	candidates := []Edge{}
	limitations := append([]string{}, recipe.Limitations...)
	if graph.MultigraphCollapsed {
		limitations = append(limitations, "graphify multigraph collapse risk; verify source edges before impact claims")
	}
	for _, graphEdge := range graph.Edges {
		if graphEdge.From != request.Seed.ID && graphEdge.To != request.Seed.ID {
			continue
		}
		from, fromOK := nodes[graphEdge.From]
		to, toOK := nodes[graphEdge.To]
		if !fromOK || !toOK {
			return ProviderLookup{}, fmt.Errorf("graphify edge %s->%s references missing node", graphEdge.From, graphEdge.To)
		}
		if !validSpan(graphEdge.Span) {
			return ProviderLookup{}, fmt.Errorf("graphify edge %s->%s has invalid source span", graphEdge.From, graphEdge.To)
		}
		edgeType := normalizeGraphifyEdgeType(graphEdge.Type)
		evidence := "structural"
		confidence := "high"
		edgeLimitations := append([]string{}, limitations...)
		if graphEdge.Inferred {
			edgeType = "INFERRED_" + edgeType
			evidence = "heuristic"
			confidence = "low"
			edgeLimitations = append(edgeLimitations, "provider marked this edge inferred")
		}
		candidates = append(candidates, Edge{
			From:        from.ID,
			To:          to.ID,
			Type:        edgeType,
			Evidence:    evidence,
			Confidence:  confidence,
			Provenance:  "graphify-output:" + path,
			Span:        graphEdge.Span,
			Limitations: edgeLimitations,
			LoadBearing: false,
			Metadata:    EdgeMetadata{Provider: GraphifyProvider, Query: recipe.ID},
		})
	}
	budgeted := ApplyFlowBudget(recipe, candidates)
	packet := Packet{
		SchemaVersion: 1,
		PacketID:      request.PacketID,
		Repository:    request.Repository,
		Seeds: Seeds{
			Symbols: []Symbol{request.Seed},
		},
		Edges:       budgeted.Edges,
		Fallback:    Fallback{Mode: "file-native", Reason: "verify structural provider edges against source before editing"},
		Budget:      Budget{MaxEdges: recipe.MaxEdges, MaxBytes: recipe.MaxBytes, Truncated: budgeted.Truncated},
		Limitations: budgeted.Limitations,
		GeneratedBy: GraphifyProvider,
	}
	for _, edge := range budgeted.Edges {
		target := nodes[edge.To]
		packet.DirectImpact = append(packet.DirectImpact, ImpactNode{
			Path:       target.Span.Path,
			Symbol:     target.Symbol,
			Reason:     "bounded " + recipe.Intent + " traversal through " + edge.Type,
			Confidence: edge.Confidence,
			Edges:      []string{edge.From + "->" + edge.To},
		})
	}
	return ProviderLookup{
		Packet:        packet,
		Provider:      GraphifyProvider,
		Available:     true,
		Authoritative: false,
	}, nil
}

func validateGraphifyShape(graph GraphifyGraph) error {
	if graph.SchemaVersion != 1 {
		return fmt.Errorf("graphify graph schema_version must be 1")
	}
	if strings.TrimSpace(graph.Provider) == "" {
		return fmt.Errorf("graphify graph provider is required")
	}
	if strings.TrimSpace(graph.Repository.Identity) == "" ||
		strings.TrimSpace(graph.Repository.Revision) == "" ||
		strings.TrimSpace(graph.Repository.WorktreeFingerprint) == "" {
		return fmt.Errorf("graphify repository identity, revision, and worktree_fingerprint are required")
	}
	return nil
}

func staleGraphifyReason(graphRepo GraphifyRepository, requestRepo Repository) string {
	if graphRepo.Identity != requestRepo.Identity {
		return "graphify repository identity mismatch; falling back to file-native lookup"
	}
	if graphRepo.Revision != requestRepo.Revision {
		return "graphify revision is stale; falling back to file-native lookup"
	}
	if graphRepo.WorktreeFingerprint != requestRepo.WorktreeFingerprint {
		return "graphify worktree fingerprint mismatch; falling back to file-native lookup"
	}
	return ""
}

func normalizeGraphifyEdgeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "calls":
		return "CALLS_STATIC"
	case "calls_observed":
		return "CALLS_OBSERVED"
	case "references":
		return "REFERENCES"
	case "implements":
		return "IMPLEMENTS"
	case "overrides":
		return "OVERRIDES"
	case "reads_config":
		return "READS_CONFIG"
	case "generates":
		return "GENERATES"
	case "builds":
		return "BUILDS"
	case "tests":
		return "TESTS"
	case "documents":
		return "DOCUMENTS"
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}
