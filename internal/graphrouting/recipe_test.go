package graphrouting

import (
	"path/filepath"
	"testing"
)

func TestTraversalRecipeSelectsRequiredIntents(t *testing.T) {
	set, err := LoadTraversalRecipeSet(recipeFixturePath())
	if err != nil {
		t.Fatalf("load recipes: %v", err)
	}
	if issues := ValidateTraversalRecipeSet(set); len(issues) > 0 {
		t.Fatalf("recipe fixture issues: %v", issues)
	}
	for _, intent := range []string{"api-change", "bug", "deletion", "security-flow", "ui-behavior"} {
		recipe, ok := SelectTraversalRecipe(set, intent)
		if !ok {
			t.Fatalf("missing intent %s", intent)
		}
		if recipe.MaxDepth <= 0 || recipe.MaxEdges <= 0 || recipe.MaxBytes <= 0 {
			t.Fatalf("recipe %s has invalid budget: %#v", intent, recipe)
		}
	}
}

func TestFlowBudgetTruncatesWithContinuation(t *testing.T) {
	recipe := TraversalRecipe{ID: "security-flow", Intent: "security-flow", MaxEdges: 2, Limitations: []string{"bounded local flow only"}}
	edges := []Edge{{Type: "SOURCE"}, {Type: "GUARD"}, {Type: "SINK"}}
	result := ApplyFlowBudget(recipe, edges)
	if !result.Truncated || len(result.Edges) != 2 {
		t.Fatalf("budget did not truncate: %#v", result)
	}
	if result.Continuation == "" || len(result.Limitations) == 0 {
		t.Fatalf("truncation did not preserve continuation metadata: %#v", result)
	}
}

func recipeFixturePath() string {
	return filepath.Join("..", "..", "evals", "graph-routing", "traversal-recipes.json")
}
