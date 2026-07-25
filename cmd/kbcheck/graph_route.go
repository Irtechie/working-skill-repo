package main

import (
	"fmt"
	"io"

	"github.com/Irtechie/working-skill-repo/internal/graphrouting"
)

func runGraphRouteCommand(root string, opts options, stdout, stderr io.Writer) int {
	packetPath := resolveInputPath(root, opts.packetPath)
	packet, err := graphrouting.Load(packetPath)
	if err != nil {
		fmt.Fprintf(stderr, "graph-route: %v\n", err)
		return 1
	}
	result := graphrouting.Validate(packet)
	if annotationIssues := graphrouting.ValidateTraversalAnnotations(packet); len(annotationIssues) > 0 {
		result.OK = false
		result.Issues = append(result.Issues, annotationIssues...)
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK {
		fmt.Fprintln(stdout, graphrouting.FormatSummary(result.Summary))
	} else {
		for _, issue := range result.Issues {
			fmt.Fprintln(stderr, issue)
		}
	}
	if !result.OK {
		return 2
	}
	return 0
}
