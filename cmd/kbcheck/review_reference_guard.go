package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type reviewReferenceGuardConfig struct {
	ReferenceRoots   []reviewReferenceRoot  `json:"reference_roots"`
	SharedPairs      []reviewReferenceEntry `json:"shared_pairs"`
	IntentionalForks []reviewReferenceEntry `json:"intentional_forks"`
}

type reviewReferenceRoot struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type reviewReferenceEntry struct {
	ID     string `json:"id"`
	Left   string `json:"left"`
	Right  string `json:"right"`
	Owner  string `json:"owner"`
	Reason string `json:"reason"`
}

type reviewReferenceGuardResult struct {
	OK               bool                   `json:"ok"`
	ReferenceRoots   int                    `json:"reference_roots"`
	SharedPairCount  int                    `json:"shared_pair_count"`
	IntentionalForks int                    `json:"intentional_forks"`
	IssueCount       int                    `json:"issue_count"`
	Issues           []reviewReferenceIssue `json:"issues"`
}

type reviewReferenceIssue struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func runReviewReferenceGuardCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-quality.json"
	}
	result, err := computeReviewReferenceGuard(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Review reference guard: roots=%d shared=%d forks=%d issues=%d\n", result.ReferenceRoots, result.SharedPairCount, result.IntentionalForks, result.IssueCount)
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s: %s\n", issue.Kind, issue.Path, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

func computeReviewReferenceGuard(root, configPath string) (reviewReferenceGuardResult, error) {
	var result reviewReferenceGuardResult
	config, err := loadSkillQualityConfig(root, configPath)
	if err != nil {
		return result, err
	}
	result.ReferenceRoots = len(config.ReviewReferenceGuard.ReferenceRoots)
	result.SharedPairCount = len(config.ReviewReferenceGuard.SharedPairs)
	result.IntentionalForks = len(config.ReviewReferenceGuard.IntentionalForks)
	classifiedPairs := map[string]bool{}
	if result.SharedPairCount == 0 {
		result.Issues = append(result.Issues, reviewReferenceIssue{
			Kind:    "missing-shared-pairs",
			ID:      "review-reference-guard",
			Path:    configPath,
			Message: "Config must declare at least one shared review-reference pair.",
		})
	}
	for _, pair := range config.ReviewReferenceGuard.SharedPairs {
		classifiedPairs[reviewReferencePairKey(pair.Left, pair.Right)] = true
		validateReviewReferenceMetadata(pair, "shared-pair", &result.Issues)
		left, leftOK := readReviewReference(root, pair.ID, pair.Left, &result.Issues)
		right, rightOK := readReviewReference(root, pair.ID, pair.Right, &result.Issues)
		if !leftOK || !rightOK {
			continue
		}
		leftHash := sha256.Sum256(left)
		rightHash := sha256.Sum256(right)
		if leftHash != rightHash {
			result.Issues = append(result.Issues, reviewReferenceIssue{
				Kind:    "shared-reference-drift",
				ID:      pair.ID,
				Path:    pair.Left + " <-> " + pair.Right,
				Message: "Declared shared review references must hash-match; update both files together or reclassify as an intentional fork with owner and reason.",
			})
		}
	}
	for _, fork := range config.ReviewReferenceGuard.IntentionalForks {
		classifiedPairs[reviewReferencePairKey(fork.Left, fork.Right)] = true
		validateReviewReferenceMetadata(fork, "intentional-fork", &result.Issues)
		readReviewReference(root, fork.ID, fork.Left, &result.Issues)
		readReviewReference(root, fork.ID, fork.Right, &result.Issues)
	}
	validateReviewReferenceSweep(root, config.ReviewReferenceGuard.ReferenceRoots, classifiedPairs, &result.Issues)
	result.IssueCount = len(result.Issues)
	result.OK = result.IssueCount == 0
	return result, nil
}

func validateReviewReferenceMetadata(entry reviewReferenceEntry, kind string, issues *[]reviewReferenceIssue) {
	if strings.TrimSpace(entry.ID) == "" {
		*issues = append(*issues, reviewReferenceIssue{Kind: kind + "-missing-id", Path: entry.Left + " <-> " + entry.Right, Message: "Review reference entry must have an id."})
	}
	if strings.TrimSpace(entry.Left) == "" || strings.TrimSpace(entry.Right) == "" {
		*issues = append(*issues, reviewReferenceIssue{Kind: kind + "-missing-path", ID: entry.ID, Path: entry.Left + " <-> " + entry.Right, Message: "Review reference entry must name left and right paths."})
	}
	if strings.TrimSpace(entry.Owner) == "" {
		*issues = append(*issues, reviewReferenceIssue{Kind: kind + "-missing-owner", ID: entry.ID, Path: entry.Left + " <-> " + entry.Right, Message: "Review reference entry must record an owner."})
	}
	if strings.TrimSpace(entry.Reason) == "" {
		*issues = append(*issues, reviewReferenceIssue{Kind: kind + "-missing-reason", ID: entry.ID, Path: entry.Left + " <-> " + entry.Right, Message: "Review reference entry must record a reason."})
	}
}

func readReviewReference(root, id, path string, issues *[]reviewReferenceIssue) ([]byte, bool) {
	if strings.TrimSpace(path) == "" {
		return nil, false
	}
	content, err := os.ReadFile(resolveRepoPath(root, path))
	if err != nil {
		*issues = append(*issues, reviewReferenceIssue{
			Kind:    "missing-file",
			ID:      id,
			Path:    path,
			Message: fmt.Sprintf("Review reference file is missing or unreadable: %v", err),
		})
		return nil, false
	}
	return content, true
}

func validateReviewReferenceSweep(root string, roots []reviewReferenceRoot, classifiedPairs map[string]bool, issues *[]reviewReferenceIssue) {
	if len(roots) == 0 {
		return
	}
	filesByRoot := make([]map[string]string, len(roots))
	for i, refRoot := range roots {
		if strings.TrimSpace(refRoot.ID) == "" {
			*issues = append(*issues, reviewReferenceIssue{Kind: "reference-root-missing-id", Path: refRoot.Path, Message: "Reference root must have an id."})
		}
		if strings.TrimSpace(refRoot.Path) == "" {
			*issues = append(*issues, reviewReferenceIssue{Kind: "reference-root-missing-path", ID: refRoot.ID, Message: "Reference root must have a path."})
			continue
		}
		files, ok := listReviewReferenceRoot(root, refRoot, issues)
		if !ok {
			continue
		}
		filesByRoot[i] = files
	}
	for i := 0; i < len(roots); i++ {
		if filesByRoot[i] == nil {
			continue
		}
		for j := i + 1; j < len(roots); j++ {
			if filesByRoot[j] == nil {
				continue
			}
			for rel, leftPath := range filesByRoot[i] {
				rightPath, ok := filesByRoot[j][rel]
				if !ok {
					continue
				}
				if classifiedPairs[reviewReferencePairKey(leftPath, rightPath)] {
					continue
				}
				*issues = append(*issues, reviewReferenceIssue{
					Kind:    "unclassified-common-reference",
					ID:      roots[i].ID + ":" + roots[j].ID + ":" + rel,
					Path:    leftPath + " <-> " + rightPath,
					Message: "Common review-reference filename is not classified as shared or intentionally forked.",
				})
			}
		}
	}
}

func listReviewReferenceRoot(root string, refRoot reviewReferenceRoot, issues *[]reviewReferenceIssue) (map[string]string, bool) {
	fullRoot := resolveRepoPath(root, refRoot.Path)
	info, err := os.Stat(fullRoot)
	if err != nil || !info.IsDir() {
		*issues = append(*issues, reviewReferenceIssue{Kind: "missing-reference-root", ID: refRoot.ID, Path: refRoot.Path, Message: "Configured reference root is missing or not a directory."})
		return nil, false
	}
	files := map[string]string{}
	err = filepath.WalkDir(fullRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(fullRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		files[rel] = joinReviewReferencePath(refRoot.Path, rel)
		return nil
	})
	if err != nil {
		*issues = append(*issues, reviewReferenceIssue{Kind: "reference-root-read-error", ID: refRoot.ID, Path: refRoot.Path, Message: err.Error()})
		return nil, false
	}
	return files, true
}

func joinReviewReferencePath(rootPath, rel string) string {
	return normalizeReviewReferencePath(filepath.Join(filepath.FromSlash(rootPath), filepath.FromSlash(rel)))
}

func reviewReferencePairKey(left, right string) string {
	parts := []string{normalizeReviewReferencePath(left), normalizeReviewReferencePath(right)}
	sort.Strings(parts)
	return parts[0] + "\x00" + parts[1]
}

func normalizeReviewReferencePath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
}
