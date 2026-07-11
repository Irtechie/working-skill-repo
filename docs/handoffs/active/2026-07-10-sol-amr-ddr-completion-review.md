# Sol Review: One AMR Loop, Skill Consolidation, Then Images

## Purpose

Use a fresh large-model session to kick the tires on one execution loop across
all affected skills. Remove duplicate DDR ownership and malformed tier rules.
Make the default invisible and safe for new users while preserving concise
advanced overrides. Edit skills/code only after the review; images are last.

## Start

```text
kb-start docs/handoffs/active/2026-07-10-sol-amr-ddr-completion-review.md
```

Use `kb-first-principles`, `repo-critic`, and `kb-architecture-deepening`.
Return keep/consolidate/delete and corrected architecture before implementation.

## Current State

- Repo/branch/HEAD: `E:\working-skill-repo` /
  `codex/session-model-routing` / `33957e0`
- Dirty worktree contains concurrent work. Never reset, revert, stage, or
  overwrite unrelated files.
- `kb-plan` says a lower-tier worker may draft, but `kb-work`, `kb-configure`,
  requirements, and `internal/modelrouting/selector.go` treat planned tier as a
  hard floor and forbid automatic downward attempts.
- DDR separately describes lower-cost draft, proof, stronger review, repair,
  and escalation. This duplicates the desired AMR control loop.
- DDR has skill prose and manifest-field validation, not a deterministic runtime.
- `go test ./cmd/kbcheck ./internal/modelrouting` passed. `go test
  ./cmd/kbrouter` fails at
  `TestCatalogOpenAICompatibleDiscoveryUsesTrustedRouteAndBoundedResponse`.

## Read First

1. `.github/skills/kb-plan/SKILL.md`
2. `.github/skills/kb-work/SKILL.md`
3. `.github/skills/kb-work/references/execution-prompt.md`
4. `.github/skills/kb-functional-test/SKILL.md`
5. `.github/skills/kb-configure/SKILL.md`
6. `.github/skills/kb-models/SKILL.md`
7. `cmd/kbcheck/context_packet.go`, `manifest_contract.go`, and `swarm.go`
8. `internal/modelrouting/selector.go` and `cmd/kbrouter/`
9. `README.md`, `docs/context/architecture/kb-workflow.md`, and
   `docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md`

## Architecture Hypothesis to Falsify

Planner records task difficulty/required correction authority, constraints, and
proof—not DDR and not a permanent model. Prefer a clearer name than the current
ambiguous `model_tier`, such as `validation_tier`, only if migration cost earns it.

```text
plan: validation tier = Medium + bounded packet + objective proof
  -> AMR checks whether a lower-tier attempt is safe
     -> no: start with an eligible Medium route
     -> yes: try an eligible Small route
        -> proof passes: keep result; continue ordinary QA/completion
        -> proof fails: preserve valid work and hand a surgical correction to Medium
           -> proof passes: continue
           -> fails/crosses boundary: higher/current driver, re-plan, or HITL
```

The planner does not create DDR. AMR owns attempt, proof-triggered escalation,
and correction routing. Delete DDR as a separate user/configuration concept if
this loop covers its only distinct behavior.

“This is code” is insufficient for a lower-tier attempt. Require settled intent,
bounded files/interfaces/authority, objective proof, and explicit escalation
triggers. HTML from an approved mockup with browser assertions qualifies;
subjective design direction, philosophy/policy judgment, unresolved architecture,
or code without an adequate oracle starts at the planned tier or driver/HITL.

## Surgical Failure Handoff

On failure, do not restart or rewrite the file. Give the planned-tier model the
original packet, current file, compact diff, attempt ledger, and:

```yaml
accepted_result: <what already satisfies the contract>
failed_criterion: <one observed failure>
failure_location: <file + symbol/line/hunk when known>
allowed_change: <smallest behavior/files/hunks>
preserve_invariants: [<interfaces, passing behavior, tests, user decisions>]
relevant_interfaces: [<callers/callees/schema/DOM/API>]
proof: {command: <exact>, exit: <status>, artifact: <path/hash>, failure: <excerpt>}
corrective_diff_only: true
proof_to_rerun: [<focused>, <regression>]
```

The larger model returns a corrective diff plus proof. Broaden only when the
failure cannot be localized, crosses interfaces, invalidates the plan, or the
focused correction fails.

## One-Loop Duplication Audit

Review every owner and apply the deletion test:

| Surface | Question |
|---|---|
| `kb-plan` | Does it record difficulty/proof once without execution policy? |
| `kb-work` | Can it own the complete attempt -> proof -> surgical escalation loop? |
| `kb-functional-test` | Does it classify proof without separately deciding implementation ownership? |
| `kb-configure` | Can DDR modes/fields be deleted, leaving delivery and true advanced preferences? |
| `kb-models` | Are `use`, `require`, preferences, and `ignore routing` sufficient advanced controls? |
| execution prompt | Is there one canonical worker/correction packet rather than repeated prose? |
| `kbrouter` | Does one decision API return attempt route, planned correction tier, and fallback? |
| `kbcheck` | Are self-authored DDR fields removed in favor of observable attempt/proof/escalation evidence? |
| README/architecture/images | Do they explain the same loop without exposing internal ceremony to beginners? |

Return at most three deeper architecture candidates. Prefer the one that deletes
the most configuration, duplicate branches, manifest fields, and skill prose.

## User Experience Contract

New user: no questionnaire or DDR terminology. Show at most one compact line:
`Trying Small for bounded code; Medium correction fallback.` Ordinary proof stays
authoritative.

Advanced user: preserve run-scoped `use <model>`, `require <model>`,
`prefer local|hosted`, and `ignore model routing`. Provide an explicit policy to
disable lower-tier attempts without requiring per-project model mapping.

## Required Fixtures and Kill Gate

- Medium bounded code + strong proof -> Small attempt; pass without Medium rewrite.
- Small attempt fails locally -> Medium receives surgical packet and preserves
  accepted hunks.
- Failure crosses an interface -> no forced surgical correction; broaden/re-plan.
- Medium subjective/ambiguous task -> no Small attempt.
- Missing proof or authority expansion -> no lower-tier attempt.
- `require` unavailable -> pause; do not silently choose another model.
- AMR unavailable -> current driver and ordinary proof still work.

Compare against direct planned-tier execution: first-pass correctness, total
tokens/time/cost, repeated hunks, collateral diff, escalation rate, and user
interventions. If correctness regresses or savings/throughput are neutral, delete
the lower-tier attempt machinery rather than optimizing it.

## Images Last

Do not edit `docs/assets/kb-model-selection.png` or
`docs/assets/kb-routing-workflow.png` until the audit, approved edits, and focused
proof pass. Then show one AMR loop; do not diagram DDR as a second subsystem.

## Review Outcome — 2026-07-11

### Keep

- `kb-plan` owns portable difficulty, constraints, context, risk, and objective
  proof. It never persists a model, route, transport, source preference, or
  attempt tier.
- `kb-work` discovers the live catalog immediately before execution and the
  current master selects an eligible bounded worker. Ordinary proof remains the
  acceptance authority; route receipts provide attribution only.
- `kb-models` owns optional user-local OpenAI-compatible/LiteLLM extras and
  personal source priority. Host-native discovery needs no setup.
- Run-only `use`, `require`, `prefer self-hosted|native`, and `ignore model
  routing` remain the advanced controls.

### Consolidate

- DDR's only useful behavior is now the AMR admission/attempt/proof/handoff
  loop. There is no separate DDR subsystem or planner-created DDR artifact.
- AMR may make one exact next-lower attempt only under explicit enabled policy
  and objective admission. Failure stops lower retries and creates a bounded
  surgical handoff for separate ordinary planned-tier execution.
- Model origin, hosting class, trust, dispatch qualification, and receipt-bound
  proof are independent facts. Source preference only orders already-eligible
  routes.

### Delete / Refuse

- Delete DDR terminology, DDR configuration modes, durable attempt tiers,
  static versioned model rosters, and a Planner model tier.
- Refuse philosophy, speculative design, unresolved architecture, weak proof,
  or authority expansion as lower-tier AMR work. Approved HTML can qualify when
  browser proof is objective.
- Refuse automatic live-checkout surgical correction until an isolated
  workspace and compare-and-swap apply runner exist. Current failures record no
  preserved-work savings.
- Refuse live support/promotion claims from self-authored evidence. The current
  release validator accepts only deterministic, zero-paid, zero-supported,
  not-promoted evidence until an external verifier exists.

## Implemented State

- Installer: strict SemVer, HTTPS-only remote roots/redirects, bounded downloads,
  timeout, checksum verification, backup-safe replacement, native platform
  lifecycle tests, and skill-only fallback.
- Release: tag must equal `v${package.json.version}`; publication depends on Go,
  installer, tag-contract, and no-paid evidence verification.
- Required skill copies are hash-synced across Codex, Copilot, shared agents,
  and ATV. Optional thin ATV scaffold/plugin drift remains warning-only.
- Images regenerated last:
  - `docs/assets/kb-model-selection.png` — SHA-256
    `14a1e54521a50068eb066d2de82c5c07abd85cc54caa15bd77fc5c1c0977448f`
  - `docs/assets/kb-routing-workflow.png` — SHA-256
    `1e70181afc66be45e2c4797806929623b9a6e72e63022bb151401e550e994e4f`

## Proof And Gate Result

Passed after the adversarial repair pass:

- `node --test ./bin/kb-install.test.mjs` — 19/19
- `node --test ./bin/check-release-tag.test.mjs` — 3/3
- `go test ./... -count=1`
- manifest contract and `git diff --check`
- `kbcheck skill-sync-report --json` — `ok: true`, zero required issues
- canonical no-paid evidence — not promoted, zero supported cohorts, zero paid
  calls
- exact `go run ./cmd/kbcheck local-release` on the final staged delivery
  candidate — exit 0; proven tree recorded in delivery metadata

The initial wrapper failures were useful and are closed: concurrent follow-on
scope leaked into dirty-tree proof, subprocesses were silent/unbounded, fresh
Windows checkout changed hash-bound bytes, and generated Python caches polluted
skill identity. The final gate runs against an unattached staged commit in a
fresh worktree, excludes GHCP/AIC follow-on paths, enforces LF checkout bytes,
bounds/contains proof subprocesses, ignores runtime caches, and passes exactly.

Concurrent follow-on GHCP/AIC benchmark files and plans are present. Preserve
them, but do not mix their implementation into the model-routing baseline
commit unless separately reviewed and authorized.
