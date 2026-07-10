package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type contextPacket struct {
	SchemaVersion      int      `json:"schema_version"`
	PacketID           string   `json:"packet_id"`
	TaskID             string   `json:"task_id"`
	Objective          string   `json:"objective"`
	MemoryFiles        []string `json:"memory_files"`
	SourceFiles        []string `json:"source_files"`
	Prefetch           []string `json:"prefetch"`
	Constraints        []string `json:"constraints"`
	OutOfScope         []string `json:"out_of_scope"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	ProofTargets       []string `json:"proof_targets"`
	ModelTier          string   `json:"model_tier"`
	ModelTierReason    string   `json:"model_tier_reason"`
	AllowedTools       []string `json:"allowed_tools"`
	BroadSearchPolicy  string   `json:"broad_search_policy"`
	EscalationTriggers []string `json:"escalation_triggers"`
}

type contextPacketResult struct {
	OK     bool     `json:"ok"`
	Issues []string `json:"issues"`
}

func runContextPacketCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.packetPath)
	var packet contextPacket
	if err := readJSONFile(path, &packet); err != nil {
		if opts.json {
			writeJSON(stdout, contextPacketResult{OK: false, Issues: []string{err.Error()}})
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 1
	}
	result := validateContextPacket(packet)
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK {
		fmt.Fprintf(stdout, "context packet: ok id=%s tier=%s\n", packet.PacketID, packet.ModelTier)
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

func validateContextPacket(packet contextPacket) contextPacketResult {
	issues := []string{}
	require := func(ok bool, message string) {
		if !ok {
			issues = append(issues, message)
		}
	}
	require(packet.SchemaVersion == 1, "schema_version must be 1")
	require(strings.TrimSpace(packet.PacketID) != "", "packet_id is required")
	require(strings.TrimSpace(packet.TaskID) != "", "task_id is required")
	require(strings.TrimSpace(packet.Objective) != "", "objective is required")
	require(nonBlankStrings(packet.SourceFiles), "source_files must contain bounded paths")
	require(nonBlankStrings(packet.Constraints), "constraints must not be empty")
	require(nonBlankStrings(packet.OutOfScope), "out_of_scope must not be empty")
	require(nonBlankStrings(packet.AcceptanceCriteria), "acceptance_criteria must not be empty")
	require(nonBlankStrings(packet.ProofTargets), "proof_targets must not be empty")
	require(validModelTier(packet.ModelTier), "model_tier must be tiny, small, medium, or large")
	require(strings.TrimSpace(packet.ModelTierReason) != "", "model_tier_reason is required")
	require(nonBlankStrings(packet.AllowedTools), "allowed_tools must not be empty")
	require(nonBlankStrings(packet.EscalationTriggers), "escalation_triggers must not be empty")
	require(contains([]string{"forbidden", "bounded", "allowed"}, packet.BroadSearchPolicy), "broad_search_policy must be forbidden, bounded, or allowed")
	if packet.BroadSearchPolicy != "allowed" {
		require(boundedRepoPaths(packet.SourceFiles), "source_files must be relative paths without traversal or globs when broad search is not allowed")
		require(boundedRepoPaths(packet.MemoryFiles), "memory_files must be relative paths without traversal or globs when broad search is not allowed")
	}
	return contextPacketResult{OK: len(issues) == 0, Issues: issues}
}

func nonBlankStrings(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func boundedRepoPaths(values []string) bool {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		clean := filepath.Clean(trimmed)
		if trimmed == "" || clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.ContainsAny(value, "*?[]") {
			return false
		}
	}
	return true
}

func runContextPacketSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "context-packet-selftest-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(temp)
	valid := contextPacket{
		SchemaVersion: 1, PacketID: "packet-1", TaskID: "task-1", Objective: "prove packet",
		SourceFiles: []string{"a.go"}, Constraints: []string{"no daemon"}, OutOfScope: []string{"runtime"}, AcceptanceCriteria: []string{"passes"},
		ProofTargets: []string{"go test ./..."}, ModelTier: "small", ModelTierReason: "bounded",
		AllowedTools: []string{"rg"}, BroadSearchPolicy: "bounded", EscalationTriggers: []string{"scope expands"},
	}
	if result := validateContextPacket(valid); !result.OK {
		fmt.Fprintf(stderr, "valid packet failed: %v\n", result.Issues)
		return 1
	}
	valid.SourceFiles = nil
	result := validateContextPacket(valid)
	if result.OK {
		fmt.Fprintf(stderr, "invalid packet accepted: %v\n", result.Issues)
		return 1
	}
	encoded, _ := json.Marshal(valid)
	if len(encoded) == 0 {
		fmt.Fprintln(stderr, "usage encoding failed")
		return 1
	}
	fmt.Fprintln(stdout, "context packet selftest passed")
	return 0
}
