package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type governorContractFile struct {
	Path    string
	Needles []string
}

func runWorkflowGovernorSelftest(root string, stdout, stderr io.Writer) int {
	files := []governorContractFile{
		{
			Path: ".github/skills/kb-brainstorm/SKILL.md",
			Needles: []string{
				"## Question Gate",
				"`ask-now`",
				"`research-first`",
				"`safe-assumption`",
				"`defer-to-planning`",
				"`parked`",
				"Do not hand off with unresolved `ask-now` or `research-first` items.",
			},
		},
		{
			Path: ".github/skills/kb-brainstorm/references/requirements-template.md",
			Needles: []string{
				"[ask-now]",
				"[research-first]",
				"[safe-assumption]",
				"[defer-to-planning]",
				"[parked]",
			},
		},
		{
			Path: ".github/skills/kb-gate/SKILL.md",
			Needles: []string{
				"## Question Gate Classes",
				"unresolved `ask-now` items",
				"unresolved `research-first` items",
				"`safe-assumption` is not a loophole.",
			},
		},
		{
			Path: ".github/skills/kb-gate/references/gate-ledger.md",
			Needles: []string{
				"Question Gate classification completed",
				"No unresolved ask-now or research-first items remain",
				"Unlabeled material assumptions count as blockers.",
			},
		},
		{
			Path: ".github/skills/kb-plan/SKILL.md",
			Needles: []string{
				"Planning cannot launder brainstorm ambiguity.",
				"unresolved `ask-now` or `research-first` items",
				"Write or update the `brainstorm-to-plan` gate as `blocked` or",
			},
		},
		{
			Path: ".github/skills/kb-epic/SKILL.md",
			Needles: []string{
				"Use the shared Question Gate classes from `kb-gate`",
				"collect\n  `ask-now` questions across all brainstorm-needed workstreams",
				"Resolve `research-first` items with research where possible",
			},
		},
		{
			Path: ".github/skills/klfg/SKILL.md",
			Needles: []string{
				"`klfg` is the strict workflow governor.",
				"Brainstorm cannot advance while `ask-now` or `research-first` items remain.",
				"Complete cannot emit DONE without a passed or quarantined `complete-to-ship`",
			},
		},
	}

	missing := []string{}
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s: read failed: %v", file.Path, err))
			continue
		}
		text := string(content)
		for _, needle := range file.Needles {
			if !strings.Contains(text, needle) {
				missing = append(missing, fmt.Sprintf("%s: missing %q", file.Path, needle))
			}
		}
	}

	if len(missing) > 0 {
		for _, issue := range missing {
			fmt.Fprintln(stderr, issue)
		}
		return 1
	}

	fmt.Fprintln(stdout, "Workflow governor selftest: question and phase gate contract present.")
	return 0
}
