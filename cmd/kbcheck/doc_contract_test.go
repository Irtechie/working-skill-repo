package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// Documentation contracts guard that a policy is still documented. They must not
// guard the exact wording used to document it, or every clarifying edit becomes a
// build failure.
//
// docAnchor matches structural text that is stable by construction: headings,
// commands, config keys, code spans, and literal output templates.
//
// docConcept matches a policy by its distinctive terms, so rewording the sentence
// keeps the contract green while deleting the policy still fails.

type contractMatcher struct {
	anchor string
	terms  []string
}

func docAnchor(text string) contractMatcher {
	return contractMatcher{anchor: text}
}

func docConcept(terms ...string) contractMatcher {
	return contractMatcher{terms: terms}
}

// negationForms collapses the interchangeable ways a policy says "not", so a
// contract survives "Do not X" becoming "Never X" while still requiring the
// prohibition to be present.
var negationForms = strings.NewReplacer(
	"do not ", "not ",
	"does not ", "not ",
	"did not ", "not ",
	"dont ", "not ",
	"doesnt ", "not ",
	"cannot ", "not ",
	"can not ", "not ",
	"must not ", "not ",
	"should not ", "not ",
	"will not ", "not ",
	"wont ", "not ",
	"never ", "not ",
)

// normalizeContractText collapses the differences that carry no policy meaning:
// letter case, line wrapping, sentence punctuation, markdown emphasis, and the
// interchangeable spellings of a prohibition.
func normalizeContractText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	pendingSpace := false
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			pendingSpace = b.Len() > 0
		case strings.ContainsRune(`.,;!?"'*_()[]`, r):
			// Volatile punctuation and markdown emphasis are not policy. Colons are
			// kept because they carry structure in headings, labels, and config keys.
		default:
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}
			b.WriteRune(unicode.ToLower(r))
		}
	}
	// Replace on a space-terminated copy so a trailing negation ("... cannot")
	// normalizes the same way it would mid-sentence.
	return strings.TrimSpace(negationForms.Replace(b.String() + " "))
}

func (m contractMatcher) describe() string {
	if len(m.terms) > 0 {
		return "concept " + strings.Join(m.terms, " + ")
	}
	return "anchor " + m.anchor
}

// missing reports the first requirement absent from already-normalized text.
func (m contractMatcher) missing(normalized string) (string, bool) {
	if len(m.terms) > 0 {
		for _, term := range m.terms {
			if !strings.Contains(normalized, normalizeContractText(term)) {
				return "concept term " + term, true
			}
		}
		return "", false
	}
	if !strings.Contains(normalized, normalizeContractText(m.anchor)) {
		return "anchor " + m.anchor, true
	}
	return "", false
}

// mustMutateContract removes a policy term so a contract check is proven able to
// fail. It replaces every occurrence and fails loudly when the term is already
// gone, so a stale mutation never masquerades as a passing guard.
func mustMutateContract(t *testing.T, text, target, replacement string) string {
	t.Helper()
	if target == "" {
		t.Fatal("mutation target must not be empty")
	}
	var b strings.Builder
	remaining := text
	replacements := 0
	for {
		index := strings.Index(strings.ToLower(remaining), strings.ToLower(target))
		if index < 0 {
			b.WriteString(remaining)
			break
		}
		b.WriteString(remaining[:index])
		b.WriteString(replacement)
		remaining = remaining[index+len(target):]
		replacements++
	}
	if replacements == 0 {
		t.Fatalf("mutation target %q is no longer present; update the mutation to match current wording", target)
	}
	return b.String()
}

// docContractProblems reports every unmet requirement so the contract logic can
// itself be tested for the ability to fail.
func docContractProblems(root string, required map[string][]contractMatcher) ([]string, error) {
	var problems []string
	for relative, matchers := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		normalized := normalizeContractText(string(content))
		for _, matcher := range matchers {
			if what, absent := matcher.missing(normalized); absent {
				problems = append(problems, relative+": "+what)
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
}

// requireDocContract asserts every matcher holds for its documentation file.
func requireDocContract(t *testing.T, root, label string, required map[string][]contractMatcher) {
	t.Helper()
	problems, err := docContractProblems(root, required)
	if err != nil {
		t.Fatalf("read %s documentation: %v", label, err)
	}
	for _, problem := range problems {
		t.Errorf("missing %s -> %s", label, problem)
	}
}
