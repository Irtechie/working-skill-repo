---
result: session-model-routing-release
date: 2026-07-11
status: passed
---

# Session Model Routing Release Proof

The final staged baseline is created with `git write-tree` plus `git commit-tree`
and checked out into a fresh detached worktree. The delivery commit and PR
record its proven tree hash because embedding that hash here would change the
tree it identifies.
The concurrent GHCP/AIC follow-on paths (`cmd/amrbench`,
`evals/amr-model-benchmark`, and the July 11 GHCP plans/context) were absent.

## Exact Release Gate

Command:

```text
go run ./cmd/kbcheck local-release
```

Result: exit 0.

```text
KB release gate: profile=local-release ok=true
passed [required/deterministic-local] kb-check-all: kbcheck core
passed [required/deterministic-local] git-diff-check: git diff --check
passed [required/deterministic-local] skill-sync-report: kbcheck skill-sync-report
passed [optional/static-report] skill-surface-minimality: kbcheck minimality
passed [required/deterministic-local] model-routing-initial-pilot: kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
```

## Focused Evidence

- `node --test ./bin/kb-install.test.mjs`: 19/19 passed.
- `node --test ./bin/check-release-tag.test.mjs`: 3/3 passed.
- `go test ./... -count=1`: passed before staged-tree isolation.
- Required skill sync: 43/43 match for Codex, Copilot, and shared agents; zero
  required issues.
- No-paid routing evidence: not promoted, zero supported cohorts, zero paid calls.
- Fresh Windows checkout hashes for both routing fixtures and `kb-models`
  matched the working source after `.gitattributes` established LF bytes.
- The process containment test proves a grandchild started, then proves it could
  not write its delayed survivor sentinel after timeout.
- The model-routing production proof uses the same bounded-output,
  whole-process-tree containment; its focused timeout and release tests pass.
- Windows private-state bootstrap accepts Builtin Administrators only as a
  transitional owner, then requires current-user ownership and the exact
  protected DACL. Short/long temp-path aliases canonicalize before containment;
  symlinks, escapes, ADS names, reserved device names, and ambiguous suffixes
  remain fail-closed.
- Runtime cache files (`__pycache__`, `.pyc`, `.pyo`, `.DS_Store`, `Thumbs.db`)
  do not affect skill identity; a real source-file addition still changes it.

The final standalone-bundle cleanup removes all ATV and GitHub-workflow
integration. Focused router, model-routing, sync, manifest, and diff proofs are
rerun after that removal; the delivery commit and PR record the final scope.
