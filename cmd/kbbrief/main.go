package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const schemaVersion = 1

var nodeIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

type brief struct {
	SchemaVersion   int      `json:"schema_version"`
	Title           string   `json:"title"`
	Outcome         string   `json:"outcome"`
	Response        response `json:"response"`
	KeyPoints       []string `json:"key_points,omitempty"`
	HandledByAgent  []string `json:"handled_by_agent,omitempty"`
	Verification    []string `json:"verification,omitempty"`
	RisksOrLater    []string `json:"risks_or_later,omitempty"`
	CompanionSource string   `json:"companion_source,omitempty"`
	Visual          *visual  `json:"visual,omitempty"`
}

type response struct {
	Class          string `json:"class"`
	Ask            string `json:"ask,omitempty"`
	WhyYou         string `json:"why_you,omitempty"`
	Blocked        string `json:"blocked,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
	Default        string `json:"default,omitempty"`
}

type visual struct {
	Mode  string       `json:"mode"`
	Nodes []visualNode `json:"nodes,omitempty"`
	Edges []visualEdge `json:"edges,omitempty"`
}

type visualNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type visualEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

func main() {
	input := flag.String("input", "", "source executive-brief JSON")
	output := flag.String("output", "", "optional generated Markdown path")
	flag.Parse()

	if strings.TrimSpace(*input) == "" {
		fmt.Fprintln(os.Stderr, "kbbrief requires -input <brief.json>")
		os.Exit(2)
	}
	source, err := os.Open(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input: %v\n", err)
		os.Exit(1)
	}
	defer source.Close()

	data, err := decodeBrief(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid brief: %v\n", err)
		os.Exit(1)
	}
	rendered, err := renderBrief(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render brief: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(*output) == "" {
		_, _ = os.Stdout.Write(rendered)
		return
	}
	if err := writeOutput(*output, rendered); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func decodeBrief(reader io.Reader) (brief, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var value brief
	if err := decoder.Decode(&value); err != nil {
		return brief{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return brief{}, fmt.Errorf("input contains more than one JSON value")
		}
		return brief{}, err
	}
	if err := validateBrief(value); err != nil {
		return brief{}, err
	}
	return value, nil
}

func validateBrief(value brief) error {
	if value.SchemaVersion != schemaVersion {
		return fmt.Errorf("schema_version must be %d", schemaVersion)
	}
	if clean(value.Title) == "" || clean(value.Outcome) == "" {
		return fmt.Errorf("title and outcome are required")
	}
	if len(value.KeyPoints) > 5 {
		return fmt.Errorf("key_points must contain at most 5 items")
	}
	for name, items := range map[string][]string{
		"key_points":       value.KeyPoints,
		"handled_by_agent": value.HandledByAgent,
		"verification":     value.Verification,
		"risks_or_later":   value.RisksOrLater,
	} {
		for _, item := range items {
			if clean(item) == "" {
				return fmt.Errorf("%s cannot contain blank items", name)
			}
		}
	}
	switch value.Response.Class {
	case "hard_response_required":
		if clean(value.Response.Ask) == "" || clean(value.Response.WhyYou) == "" ||
			clean(value.Response.Blocked) == "" || clean(value.Response.Recommendation) == "" {
			return fmt.Errorf("hard_response_required needs ask, why_you, blocked, and recommendation")
		}
		if clean(value.Response.Default) != "" {
			return fmt.Errorf("hard_response_required cannot declare an agent-owned default")
		}
	case "soft_preference":
		if clean(value.Response.Default) == "" {
			return fmt.Errorf("soft_preference needs default")
		}
		if clean(value.Response.WhyYou) != "" || clean(value.Response.Blocked) != "" {
			return fmt.Errorf("soft_preference cannot claim a human-only blocker")
		}
	case "no_response_needed":
		if clean(value.Response.Ask) != "" || clean(value.Response.WhyYou) != "" ||
			clean(value.Response.Blocked) != "" || clean(value.Response.Default) != "" {
			return fmt.Errorf("no_response_needed cannot contain an ask, blocker, or default")
		}
	default:
		return fmt.Errorf("unsupported response class %q", value.Response.Class)
	}
	return validateVisual(value.Visual)
}

func validateVisual(value *visual) error {
	if value == nil {
		return nil
	}
	switch value.Mode {
	case "auto", "always", "none":
	default:
		return fmt.Errorf("visual.mode must be auto, always, or none")
	}
	if value.Mode == "none" {
		if len(value.Nodes) > 0 || len(value.Edges) > 0 {
			return fmt.Errorf("visual.mode none cannot contain nodes or edges")
		}
		return nil
	}
	if len(value.Nodes) == 0 && len(value.Edges) == 0 {
		return nil
	}
	nodes := map[string]bool{}
	for _, node := range value.Nodes {
		if !nodeIDPattern.MatchString(node.ID) {
			return fmt.Errorf("invalid visual node id %q", node.ID)
		}
		if clean(node.Label) == "" {
			return fmt.Errorf("visual node %q needs a label", node.ID)
		}
		if nodes[node.ID] {
			return fmt.Errorf("duplicate visual node id %q", node.ID)
		}
		nodes[node.ID] = true
	}
	for _, edge := range value.Edges {
		if !nodes[edge.From] || !nodes[edge.To] {
			return fmt.Errorf("visual edge %q -> %q references an unknown node", edge.From, edge.To)
		}
	}
	if value.Mode == "always" && (len(value.Nodes) < 2 || len(value.Edges) < 1) {
		return fmt.Errorf("visual.mode always needs at least 2 nodes and 1 edge")
	}
	return nil
}

func renderBrief(value brief) ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(&output, "# %s\n\n", clean(value.Title))
	switch value.Response.Class {
	case "hard_response_required":
		fmt.Fprintf(&output, "> **Response required: Yes.** %s\n\n", clean(value.Response.Ask))
	case "soft_preference":
		fmt.Fprintf(&output, "> **Response required: Optional.** I will use the stated default unless you override it.\n\n")
	case "no_response_needed":
		fmt.Fprintf(&output, "> **Response required: No.**\n\n")
	default:
		return nil, fmt.Errorf("unsupported response class %q", value.Response.Class)
	}

	fmt.Fprintf(&output, "## Outcome\n\n%s\n\n", clean(value.Outcome))
	if len(value.KeyPoints) > 0 {
		output.WriteString("## What matters now\n\n")
		writeNumbered(&output, value.KeyPoints)
	}

	switch value.Response.Class {
	case "hard_response_required":
		output.WriteString("## Your decision\n\n")
		fmt.Fprintf(&output, "- **Ask:** %s\n", clean(value.Response.Ask))
		fmt.Fprintf(&output, "- **Why you:** %s\n", clean(value.Response.WhyYou))
		fmt.Fprintf(&output, "- **Blocked:** %s\n", clean(value.Response.Blocked))
		fmt.Fprintf(&output, "- **Recommendation:** %s\n\n", clean(value.Response.Recommendation))
	case "soft_preference":
		output.WriteString("## Default I will use\n\n")
		fmt.Fprintf(&output, "%s\n\n", clean(value.Response.Default))
	}

	if shouldRenderVisual(value.Visual) {
		output.WriteString("## Flow\n\n```mermaid\nflowchart LR\n")
		nodes := append([]visualNode(nil), value.Visual.Nodes...)
		sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
		for _, node := range nodes {
			fmt.Fprintf(&output, "  %s[\"%s\"]\n", node.ID, mermaidLabel(node.Label))
		}
		for _, edge := range value.Visual.Edges {
			if clean(edge.Label) == "" {
				fmt.Fprintf(&output, "  %s --> %s\n", edge.From, edge.To)
			} else {
				fmt.Fprintf(&output, "  %s -->|%s| %s\n", edge.From, mermaidLabel(edge.Label), edge.To)
			}
		}
		output.WriteString("```\n\n")
	}

	writeSection(&output, "Handled by agent", value.HandledByAgent)
	writeSection(&output, "Verification", value.Verification)
	writeSection(&output, "Risks / later", value.RisksOrLater)
	if clean(value.CompanionSource) != "" {
		fmt.Fprintf(&output, "## Details\n\nSource: %s\n", clean(value.CompanionSource))
	}
	return output.Bytes(), nil
}

func shouldRenderVisual(value *visual) bool {
	if value == nil || value.Mode == "none" {
		return false
	}
	if value.Mode == "always" {
		return true
	}
	return len(value.Nodes) >= 3 && len(value.Edges) >= 2
}

func writeNumbered(output *bytes.Buffer, items []string) {
	for index, item := range items {
		fmt.Fprintf(output, "%d. %s\n", index+1, clean(item))
	}
	output.WriteByte('\n')
}

func writeSection(output *bytes.Buffer, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(output, "## %s\n\n", title)
	for _, item := range items {
		fmt.Fprintf(output, "- %s\n", clean(item))
	}
	output.WriteByte('\n')
}

func clean(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func mermaidLabel(value string) string {
	replacer := strings.NewReplacer(
		`&`, `&amp;`,
		`"`, `&quot;`,
		`<`, `&lt;`,
		`>`, `&gt;`,
		`|`, `&#124;`,
	)
	return replacer.Replace(clean(value))
}

func writeOutput(path string, content []byte) error {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absolute, content, 0o644)
}
