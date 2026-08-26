---
kb_id: kb-2026-08-26-kb-rehab-outstanding-work
slice_id: slice-002
title: "kb-rehab lane: triage, markers, and decision packet"
blockers: [slice-001]
verification: verification-plus-integration
test_level: integration
functional_risk: moderate
execution_class: cli
model_tier: medium
model_tier_reason: "Authoring one skill lane and a bounded marker writer over a proven report. The decision taxonomy is inherited from todo-triage and the packet contract from the reconciler; no new vocabulary or product choice is introduced."
model_requirements: ["KB skill authoring against the repo skill contract", "Markdown table editing that preserves existing status-marker vocabulary", "integration testing against a fixture todo.md"]
escalation_triggers:
  - "A marker or removal would be written without the report's containment proof."
  - "The lane would need a decision vocabulary that todo-triage does not already define."
  - "The packet would exceed five grouped items or omit a mandated field."
  - "Ambiguity would default to action rather than preservation."
token_budget: 20000
cost_tier: 2
cost_tier_evidence: >-
  Prior art in this repo. The decision taxonomy - ready, blocked, parked,
  duplicate, delete - is defined in .github/skills/todo-triage/SKILL.md, and the
  status-marker vocabulary is defined in todo.md's own Rules table. The bounded
  decision-packet contract with recommended choice, affected artifacts,
  evidence, irreversible consequence, safe default, and expiry sensor is already
  specified by reconciler R20 and implemented in internal/reconcile/plan.go.
  Ruled out tier 6, a new triage vocabulary and a new packet format, because two
  compatible vocabularies would force operators to learn both and would break
  the existing todo.md contract.
workspace_mode: shared-serial
conflict_domains: ["skills:kb-rehab", "skills:kb-start", "docs:readme", "go:kbcheck-commands"]
shared_resources: ["filesystem:skills-tree", "filesystem:todo.md"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'WorkReality|SkillRepoContract' -count=1"
  expect: 0
hitl: false
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: cmd/kbcheck/work_reality_test.go
    sha256: "29d5fd11517756e463750e36342b8cb49f42680f135373846fd4d36c041be914"
    update_policy: "additive only; a slice-001 assertion may not be weakened or deleted without an explicit plan amendment"
    purpose: "Marker and removal tests assert that no write occurs without the report's containment proof."
expected_files:
  - path: .github/skills/kb-rehab/SKILL.md
    op: create
    scope: "The lane: when it fires, what it reads, the triage decisions, the packet, and explicit delegation to kb-complete and kbreconcile."
  - path: .github/skills/kb-rehab/references/classification.md
    op: create
    scope: "Lazy reference for the lifecycle table and evidence rules."
  - path: cmd/kbcheck/work_reality.go
    op: modify
    scope: "Add the --action mark mode that writes todo.md markers under R7/R8 and emits the packet."
  - path: cmd/kbcheck/work_reality_test.go
    op: modify
    scope: "Prove marker writes, blocked removals, packet bounds, and preservation on unanswered items."
  - path: .github/skills/kb-start/SKILL.md
    op: modify
    scope: "Add the kb-rehab routing row to the ranked routing decision list."
  - path: README.md
    op: modify
    scope: "Add kb-rehab to the installed-skill list."
---

# Slice 002 - The Lane, Markers, and the Packet

## Observable Outcome

`kb-rehab` exists as a routable lane. Running it produces marked `todo.md` rows
for proven-dead and proven-superseded work, and one bounded decision packet for
everything ambiguous. Nothing is delivered and nothing is deleted.

## Ordered Steps

1. **Author `SKILL.md`.** Frontmatter `name` must equal the folder name
   `kb-rehab`, with a description narrow enough that the lane fires only for
   outstanding-work reconciliation and not for ordinary cleanup or review. Name
   `kb-map`, `kbcheck work-reality`, `kb-complete`, and `kbreconcile` explicitly
   as the lanes it calls; ambient discovery is forbidden by the repo skill
   contract.
   *Pass criterion:* `go run ./cmd/kbcheck skill-lint` and the skill repo
   contract test pass.

2. **Keep the skill thin.** `SKILL.md` carries scope, sequencing, escalation,
   and the delegation contract. The lifecycle table and evidence rules move to
   `references/classification.md` as a lazy reference.
   *Pass criterion:* the minimality check does not flag `kb-rehab`, and
   `SKILL.md` contains no rule already stated in `AGENTS.md`.

3. **Implement `--action mark`.** Extend the slice-001 subcommand with a mode
   that writes `todo.md` markers using the existing status-marker vocabulary
   from `todo.md`'s Rules table. Only pairings the report classified `dead` or
   `superseded` may be marked (R7).
   *Pass criterion:* a fixture asserts that a `human-required` or `orphan-*`
   pairing produces no write.

4. **Gate removal.** A declared work item may be removed only when the
   superseding or completing artifact is named in the same edit **and** its
   paired ref holds no uncontained commits. Any uncontained commit blocks
   removal and re-marks the row instead (R8).
   *Pass criterion:* a fixture with an uncontained commit asserts the row is
   re-marked and still present after the run.

5. **Emit the packet.** At most five grouped items. Each carries the recommended
   choice, affected artifacts, exact evidence and uncertainty, irreversible
   consequence, safe default, and expiry/recheck sensor (R9). For a merge
   decision in this repository the irreversible-consequence field names the
   global install targets from `AGENTS.md`.
   *Pass criterion:* a fixture with eight ambiguous pairings emits exactly five
   grouped items and asserts every mandated field is non-empty.

6. **Preserve on silence.** An unanswered packet item leaves the work item and
   the ref untouched (R10).
   *Pass criterion:* a fixture that answers nothing asserts zero `todo.md`
   writes and zero ref changes.

7. **Route it.** Add one row to the `kb-start` ranked routing decision list, and
   add `kb-rehab` to the README installed-skill list.
   *Pass criterion:* the doc contract test passes and the routing row does not
   duplicate an existing rank's signal.

## Acceptance Criteria

- The lane writes no implementation code. It classifies, marks, and packets
  (R11).
- Every `todo.md` write is traceable to a report pairing and its containment
  proof.
- The packet never exceeds five grouped items and never omits a mandated field.
- Unanswered ambiguity preserves both the work item and the ref.
- `kb-start` routes an outstanding-work request to `kb-rehab` without the user
  naming a skill.

## Test Scenarios

| Scenario | Expected |
|---|---|
| Pairing classified `dead` with containment proof | row marked, not removed |
| Pairing classified `superseded` with containment proof | row marked |
| Pairing classified `human-required` | no write, enters packet |
| Removal requested, ref has uncontained commits | removal blocked, row re-marked |
| Removal requested, artifact named, zero uncontained commits | removal permitted |
| Eight ambiguous pairings | exactly five grouped packet items |
| Packet answered with nothing | zero writes |
| Merge decision in this repository | irreversible-consequence names the three global install roots |

## Scope Boundary

No PR, no push, no merge, no deletion, no proof execution. This slice marks and
reports. Delivery and reaping belong to slice 003.

## Requirements Owned

R7, R8, R9, R10, R11.
