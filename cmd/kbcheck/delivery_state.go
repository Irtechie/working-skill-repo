package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const deliveryStateSchemaVersion = 1

// deliveryStateReceipt is the durable boundary between lifecycle machinery
// and a person deciding what happens next.
type deliveryStateReceipt struct {
	SchemaVersion  int      `json:"schema_version"`
	State          string   `json:"state"`
	Recommendation string   `json:"recommendation"`
	Proof          []string `json:"proof"`
	Blocker        string   `json:"blocker,omitempty"`
	PullRequestURL string   `json:"pull_request_url,omitempty"`
	MergedSHA      string   `json:"merged_sha,omitempty"`
	Screenshots    []string `json:"screenshots,omitempty"`
}

func runDeliveryStateCommand(root string, opts options, stdout, stderr io.Writer) int {
	receipt, issue := loadDeliveryStateReceipt(resolveInputPath(root, opts.deliveryReceiptPath))
	if issue != "" {
		fmt.Fprintln(stderr, issue)
		return 2
	}
	if opts.json {
		writeJSON(stdout, receipt)
		return 0
	}
	writeDeliveryStateProjection(stdout, receipt)
	return 0
}

func loadDeliveryStateReceipt(path string) (deliveryStateReceipt, string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return deliveryStateReceipt{}, fmt.Sprintf("read delivery receipt: %v", err)
	}
	var receipt deliveryStateReceipt
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return deliveryStateReceipt{}, fmt.Sprintf("parse delivery receipt: %v", err)
	}
	if issue := validateDeliveryStateReceipt(receipt); issue != "" {
		return deliveryStateReceipt{}, issue
	}
	return receipt, ""
}

func validateDeliveryStateReceipt(receipt deliveryStateReceipt) string {
	if receipt.SchemaVersion != deliveryStateSchemaVersion {
		return fmt.Sprintf("delivery receipt schema_version must be %d", deliveryStateSchemaVersion)
	}
	if strings.TrimSpace(receipt.Recommendation) == "" {
		return "delivery receipt recommendation is required"
	}
	if len(receipt.Proof) == 0 {
		return "delivery receipt proof is required"
	}
	for _, proof := range receipt.Proof {
		if strings.TrimSpace(proof) == "" {
			return "delivery receipt proof entries must be non-empty"
		}
	}
	switch receipt.State {
	case "local-durable":
		if receipt.PullRequestURL != "" || receipt.MergedSHA != "" {
			return "local-durable receipt cannot contain PR or merged SHA"
		}
	case "awaiting-review":
		if strings.TrimSpace(receipt.PullRequestURL) == "" {
			return "awaiting-review receipt requires pull_request_url"
		}
		if strings.TrimSpace(receipt.MergedSHA) != "" {
			return "awaiting-review receipt cannot contain merged_sha"
		}
	case "delivery-integrated":
		if strings.TrimSpace(receipt.MergedSHA) == "" {
			return "delivery-integrated receipt requires merged_sha"
		}
	default:
		return "delivery receipt state must be local-durable, awaiting-review, or delivery-integrated"
	}
	return ""
}

func writeDeliveryStateProjection(stdout io.Writer, receipt deliveryStateReceipt) {
	fmt.Fprintf(stdout, "Delivery: %s\n", receipt.State)
	fmt.Fprintf(stdout, "Next: %s\n", receipt.Recommendation)
	fmt.Fprintf(stdout, "Proof: %s\n", strings.Join(receipt.Proof, "; "))
	if receipt.Blocker != "" {
		fmt.Fprintf(stdout, "Blocker: %s\n", receipt.Blocker)
	}
	if receipt.PullRequestURL != "" {
		fmt.Fprintf(stdout, "PR: %s\n", receipt.PullRequestURL)
	}
	if receipt.MergedSHA != "" {
		fmt.Fprintf(stdout, "Merged: %s\n", receipt.MergedSHA)
	}
	if len(receipt.Screenshots) > 0 {
		fmt.Fprintf(stdout, "Screenshots: %s\n", strings.Join(receipt.Screenshots, "; "))
	}
}
