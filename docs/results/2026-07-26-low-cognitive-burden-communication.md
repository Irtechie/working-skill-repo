# Low-Cognitive-Burden Communication Result

Status: released locally and synchronized to all required global skill roots.

## Implemented

- Ambient and `kb-compact` behavior optimizes comprehension and decision effort,
  not minimum word count.
- Questions are classified as hard response required, soft preference, or no
  response needed.
- Hard questions name the exact ask, why the human owns it, what blocks, and
  the recommendation. Agent-owned defaults continue without creating review
  work.
- `kb-ship` now requires an executive PR first screen and links material
  reasoning to a source-owned companion document.
- `kb-executive-brief` and `cmd/kbbrief` generate responsibility-first Markdown
  and evidence-backed Mermaid from strict source-owned JSON.
- The newer plan-run shipping protections found in global `kb-ship` were merged
  into the repository before the communication edits.

## Proof

- PASS:
  `go test ./cmd/kbcheck -run TestLowCognitiveBurdenCommunicationContract -count=1`
- PASS: focused `kbbrief` and communication-contract tests covering strict
  input, responsibility classes, visual thresholds, valid references, golden
  output, and safe regeneration.
- PASS: `go run ./cmd/kbcheck skill-lint` with zero errors and twelve unchanged
  warnings.
- PASS: `git diff --check` for the communication-contract files.
- PASS: pre-slice regression verification, 17/17 snapshots.
- PASS: `go run ./cmd/kbcheck core` with 38 checks after reconciling the
  completed plan-worktree branch with proof-governor work.
- PASS: `go run ./cmd/kbcheck local-release`.
- PASS: `go run ./cmd/kbcheck skill-sync-report` with 138 comparisons and zero
  required issues; `kb-executive-brief`, `kb-ship`, and
  `pr-review-workbench` are hash-identical across the repository, Codex,
  Copilot, and shared-agent roots.

## Release Boundary

The complete communication, review-artifact, executive-brief, and lazy
PR-workbench layers are locally released and synchronized. Shareable HTML uses
a separate downloadable artifact branch; Pages is optional. Commit, push, PR
creation, PR edits, artifact publication, and Pages deployment remain separate
mutations and were not performed.
