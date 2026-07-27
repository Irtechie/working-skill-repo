# Low-Cognitive-Burden Communication Result

Status: base release complete; adaptive-format follow-up synchronized, with the
full release gate still blocked by the known Windows worktree-test hang.

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
- `kb-compact` now selects the smallest useful format from plain language,
  ranked bullets, tables, decision blocks, and workflow diagrams.
- Lazy-loaded examples show when each format helps and when a visual would only
  add noise.

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
- PASS: the adaptive-format follow-up's focused communication-contract test,
  skill lint, independent response forward-test, and five-root `kb-compact`
  hash comparison.
- BLOCKED: the follow-up `local-release` reruns did not complete. Both advanced
  into `go test ./...` and stalled in
  `TestPlanWorktreeSelftestExercisesDisposableLifecycle`; no failure was
  attributed to the communication files.

## Release Boundary

The base communication, review-artifact, executive-brief, and lazy PR-workbench
layers are locally released. The adaptive-format follow-up is synchronized but
not release-gate complete until the existing Windows worktree-test hang is
resolved. Shareable HTML uses a separate downloadable artifact branch; Pages is
optional. Commit, push, PR creation, PR edits, artifact publication, and Pages
deployment remain separate mutations and were not performed.
