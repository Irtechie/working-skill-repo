# PR Review Workbench Trial

Status: integration, real-PR trial, and downloadable-artifact route complete.
Fresh local preview is unavailable in the Codex in-app browser because its
security policy blocks `file://` navigation; downloadable HTML was selected as
the sufficient delivery boundary.

## What Became Easier To Understand

The two PRs publish the same feature at different layers:

- [Public marketplace PR #1](https://github.com/Irtechie/agent-marketplace-public/pull/1)
  is the lean distribution: 10 files, +854/-0, pinned at
  `f7c7cb399dbbea630a8c39b8e949fbe46baab9f0`.
- [Source marketplace PR #2](https://github.com/All-The-Vibes/skill-marketplace/pull/2)
  is the source and release system: 31 files, +2156/-2, pinned at
  `09952004339dd2f78c0b0c817a1ac285e56405a6`.

The first screen exposes that distinction without requiring the reviewer to
reconstruct it from 41 changed files. Each artifact shows one decision state,
three behavior claims with source anchors, at most five primary facts, and one
next action. Both are `ready for human decision`; that means the evidence packet
is complete enough to assess, not that either PR is correct or should be
approved.

## Artifacts

- `.kb/pr-review-workbench/agent-marketplace-public-pr-1.html`
- `.kb/pr-review-workbench/agent-marketplace-public.json`
- `.kb/pr-review-workbench/skill-marketplace-pr-2.html`
- `.kb/pr-review-workbench/skill-marketplace.json`

These trial files are intentionally ephemeral. At real PR time, a shareable
copy goes on a separate `pr-review-artifacts` branch under the PR number and
reviewed SHA. The PR links the GitHub file page with **Download, then open
locally**. The artifact never goes on the PR branch because that would change
the reviewed head SHA.

## Proof And Limits

- PASS: the adopted source package's 16 Python tests.
- PASS: repo-owned focused Go contract test renders the fixture and checks the
  decision state, five-fact ceiling, one next action, restrictive CSP, original
  PR link, and absence of inline event handlers.
- PASS: both live packets are commit-pinned, source-anchored, and rendered as
  self-contained HTML.
- LIMIT: the Codex in-app browser rejected local `file://` navigation. No
  alternate browser workaround was used after that policy decision.
- NOT DONE: no GitHub review, PR edit, artifact-branch push, Pages deployment,
  commit, or push.

Neither PR reported CI checks in the collected GitHub evidence. Reviewers still
need to judge the anchored behavioral claims and any repository-specific
release requirements.

## Visual Redesign

The initial trial proved the evidence contract but still looked like a summary
page with tabs. The upgraded renderer now includes:

- a visible PR decision topology with a pass path and repair branch;
- an application-impact spine ordered by source-backed effect on the running
  product, downstream consumers, proof, and then supporting churn;
- clickable workflow nodes backed by a coordinated evidence inspector;
- a five-step guided review from immutable scope through human decision;
- an app-impact drill-down with proportional churn and file lists;
- a separate behavioral-evidence view with source anchors, dataset state,
  checks, and blockers;
- responsive light/dark presentation with no external assets or storage.

The rich page is not the default for every PR. Small, low-risk, single-area
changes keep the compact PR first screen. Automatic rich rendering requires a
commit-pinned impact analysis or another material runtime, downstream,
security/data/API/deployment, reconstruction, or evidence-repair reason. Path
groups are an honest fallback only; they do not qualify a PR automatically.

Research:
`docs/context/research/2026-07-26-pr-review-workbench-visual-patterns.md`.

## Fresh Visual-Redesign Proof

- PASS: `python -m py_compile` for the renderer.
- PASS: `go test ./cmd/kbcheck -run TestPRReviewWorkbenchContract -count=1`.
- PASS: both real PR pages regenerated with source-backed impact packets.
- PASS: JavaScript parsing; one decision state, four primary facts, one next
  action, five guided steps, zero external scripts, and zero inline handlers.
- PASS: `go run ./cmd/kbcheck skill-lint` with zero errors,
  `git diff --check`, and `skill-sync-report` with 138/138 required matches.
- BLOCKED OUTSIDE THIS WORKSTREAM: `core` and `local-release` reach the unchanged
  `cmd/kbrouter` suite, where twelve tests pass a nonexistent temp project path
  into stricter canonicalization and fail with
  `The system cannot find the path specified`.
- BLOCKED OUTSIDE THIS WORKSTREAM: `local-release` also reports the existing
  model-routing initial-pilot evidence path as missing.
