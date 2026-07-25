package graphrouting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type TraversalRecipeSet struct {
	SchemaVersion int               `json:"schema_version"`
	Recipes       []TraversalRecipe `json:"recipes"`
}

type TraversalRecipe struct {
	ID             string   `json:"id"`
	Intent         string   `json:"intent"`
	SeedKinds      []string `json:"seed_kinds"`
	EdgeOrder      []string `json:"edge_order"`
	Direction      string   `json:"direction"`
	MaxDepth       int      `json:"max_depth"`
	MaxEdges       int      `json:"max_edges"`
	MaxBytes       int      `json:"max_bytes"`
	StopConditions []string `json:"stop_conditions"`
	Limitations    []string `json:"limitations"`
	Fallback       string   `json:"fallback"`
}

type TraversalBudgetResult struct {
	Edges        []Edge
	Truncated    bool
	Continuation string
	Limitations  []string
}

func LoadTraversalRecipeSet(path string) (TraversalRecipeSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return TraversalRecipeSet{}, err
	}
	defer file.Close()
	return DecodeTraversalRecipeSet(file)
}

func DecodeTraversalRecipeSet(reader io.Reader) (TraversalRecipeSet, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var set TraversalRecipeSet
	if err := decoder.Decode(&set); err != nil {
		return TraversalRecipeSet{}, err
	}
	return set, nil
}

func ValidateTraversalRecipeSet(set TraversalRecipeSet) []string {
	issues := []string{}
	if set.SchemaVersion != 1 {
		issues = append(issues, "traversal recipe schema_version must be 1")
	}
	seen := map[string]bool{}
	required := map[string]bool{
		"api-change":    false,
		"bug":           false,
		"deletion":      false,
		"security-flow": false,
		"ui-behavior":   false,
	}
	for i, recipe := range set.Recipes {
		prefix := fmt.Sprintf("recipe[%d]", i)
		if strings.TrimSpace(recipe.ID) == "" {
			issues = append(issues, prefix+".id is required")
		}
		if seen[recipe.ID] {
			issues = append(issues, prefix+".id duplicates another recipe")
		}
		seen[recipe.ID] = true
		if _, ok := required[recipe.Intent]; ok {
			required[recipe.Intent] = true
		}
		if len(recipe.SeedKinds) == 0 {
			issues = append(issues, prefix+".seed_kinds is required")
		}
		if len(recipe.EdgeOrder) == 0 {
			issues = append(issues, prefix+".edge_order is required")
		}
		if recipe.MaxDepth <= 0 || recipe.MaxEdges <= 0 || recipe.MaxBytes <= 0 {
			issues = append(issues, prefix+" budgets must be positive")
		}
		if len(recipe.Limitations) == 0 {
			issues = append(issues, prefix+".limitations is required")
		}
		if strings.TrimSpace(recipe.Fallback) == "" {
			issues = append(issues, prefix+".fallback is required")
		}
	}
	for intent, ok := range required {
		if !ok {
			issues = append(issues, "missing required traversal intent: "+intent)
		}
	}
	return issues
}

func SelectTraversalRecipe(set TraversalRecipeSet, intent string) (TraversalRecipe, bool) {
	for _, recipe := range set.Recipes {
		if recipe.Intent == intent {
			return recipe, true
		}
	}
	return TraversalRecipe{}, false
}

func ApplyFlowBudget(recipe TraversalRecipe, candidates []Edge) TraversalBudgetResult {
	maxEdges := recipe.MaxEdges
	if maxEdges <= 0 || maxEdges > len(candidates) {
		maxEdges = len(candidates)
	}
	result := TraversalBudgetResult{Edges: append([]Edge{}, candidates[:maxEdges]...)}
	if maxEdges < len(candidates) {
		result.Truncated = true
		result.Continuation = fmt.Sprintf("truncated after %d/%d edges for recipe %s", maxEdges, len(candidates), recipe.ID)
		result.Limitations = append(result.Limitations, "over-budget traversal requires continuation or file-native verification")
	}
	result.Limitations = append(result.Limitations, recipe.Limitations...)
	return result
}

func ValidateTraversalAnnotations(packet Packet) []string {
	issues := []string{}
	for i, edge := range packet.Edges {
		prefix := fmt.Sprintf("edge[%d]", i)
		if edge.Metadata.Provider == "graphify" && edge.Evidence == "exact" {
			issues = append(issues, prefix+" graphify structural provider cannot claim exact evidence")
		}
		if strings.Contains(strings.ToUpper(edge.Type), "INFERRED") && edge.Confidence == "exact" {
			issues = append(issues, prefix+" inferred traversal edge cannot claim exact confidence")
		}
	}
	return issues
}
