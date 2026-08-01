# KB Gate Ledger

## Contents

- Gate Record
- Question Gate proof
- Phase transitions
- Validation

Use a gate ledger when a workflow has multiple phases, expensive tests, model
runs, benchmark waves, or any path where "almost done" can waste time.

The ledger is the phase-transition source of truth. Chat history, summaries,
and agent confidence do not satisfy a gate.

## Gate Record

Each manifest should contain a `gate_ledger` list in frontmatter:

```yaml
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - docs/brainstorms/<requirements>.md
      - "Question Gate classification completed"
      - "Outstanding Questions / Resolve Before Planning is empty"
      - "No unresolved ask-now or research-first items remain"
      - "Safe assumptions, deferred planning questions, and parked items are recorded"
    proof:
      - docs/brainstorms/<requirements>.md
    blockers: []
    passed_at: "YYYY-MM-DDTHH:MM:SSZ"
    allowed_next_action: "kb-plan <requirements-path>"
```

## Status Values

| Status | Meaning | May Advance? |
|---|---|---|
| `pending` | Gate has not been checked | No |
| `blocked` | Agent, dependency, or external issue is a real impasse for this gate | No |
| `needs-human` | Only a human can authorize, supply, or judge the missing input | No |
| `quarantined` | Known issue excluded from scope with explicit evidence and owner | Yes, only outside quarantined scope |
| `passed` | Required evidence exists and blockers are resolved/quarantined | Yes |

`needs-human` is the gate-ledger equivalent of a `human-required` slice.
`paused` is not a gate result. A user pause stops execution while preserving
the current gate result; it must not be rewritten as `blocked`.

## Required Fields

- `gate_id`: stable phase or slice gate name.
- `owner_skill`: skill responsible for satisfying the gate.
- `status`: one of the status values above.
- `required_evidence`: concrete files, commands, trace exports, screenshots,
  test logs, or explicit checklist assertions.
- `proof`: paths or command records proving the evidence exists.
- `blockers`: unresolved blockers with severity and owner.
- `passed_at`: timestamp, only when status is `passed` or `quarantined`.
- `allowed_next_action`: the only next command/phase allowed after this gate.

## Blocker Lifecycle Contract

New manifests should set:

```yaml
blocker_lifecycle_contract: true
```

Every gate then declares `gate_scope` as `implementation`, `integration`,
`release`, `deployment`, or `optional-capability`.

A `blocked` or `needs-human` gate also records:

```yaml
attempted:
  - "<repair or ownership check already performed>"
responsibility: agent|human|external|dependency
affected_scope: "<smallest scope that actually cannot advance>"
resume_condition: "<observable condition that clears the gate>"
recheck: "<cheapest command or inspection that proves whether it still exists>"
checked_at: "<RFC3339 timestamp with timezone>"
propagation: current-gate-only|dependent-slices-only
```

A `quarantined` gate clears blocker lifecycle fields and instead records:

```yaml
quarantined_scope: "<exact excluded rows, files, platform, or capability>"
quarantine_owner: "<owner responsible for the excluded issue>"
quarantine_evidence:
  - "<current evidence path or command result>"
forbidden_claims:
  - "<claim the next phase must not make>"
```

Rules:

- Use `needs-human` only with `responsibility: human`.
- Use `blocked` for `agent`, `external`, or `dependency` impasses only after
  safe agent repair is exhausted.
- Recheck the named sensor before repeating a blocker in a handoff or status
  report. A nonterminal blocker older than 72 hours is stale.
- Release, deployment, and optional-capability blockers use
  `current-gate-only`. They do not downgrade implementation or unrelated
  capabilities.
- Remove blocker lifecycle metadata when a gate becomes `passed` or
  `quarantined`. A quarantined gate must retain the separate scope, owner,
  evidence, and forbidden-claims fields above; a fully passed gate clears them.
- Nonadvanceable gates must not retain `passed_at`.

## Hard Rules

1. A phase cannot advance while its gate is `pending`, `blocked`, or
   `needs-human`.
2. A gate cannot be marked `passed` unless every `required_evidence` item has a
   corresponding `proof` item.
3. P0/P1 issues block unless fixed, reclassified with evidence, or quarantined
   outside the next phase's scope.
4. Quarantine is not a pass. It must name the excluded rows/files/scope,
   evidence path, owner, and what claims are forbidden.
5. The next skill must re-read the ledger before starting. If the prior gate is
   missing or stale, stop and repair the ledger before doing work.
6. If new evidence contradicts a passed gate, downgrade it to `blocked` and set
   `allowed_next_action` to the repair/review step.
7. A blocked slice stops only dependent slices. Continue unrelated ready work.
8. A user pause preserves the current technical status and authorizes no new
   mutation unless the user also requested a handoff/state write.

## Question Gate Evidence

For `brainstorm-to-plan`, the ledger must distinguish these classes:

- `ask-now`: human answer required before planning;
- `research-first`: source/external research required before planning;
- `safe-assumption`: reversible assumption with evidence and proof hook;
- `defer-to-planning`: technical detail planning may resolve;
- `parked`: out of current scope with forbidden claims recorded.

The gate may pass only when no `ask-now` or `research-first` items remain.
Unlabeled material assumptions count as blockers.

## Minimal Gates

For KB workflows, use these gate IDs:

- `brainstorm-to-plan`
- `plan-to-work`
- `slice-<id>-to-done`
- `work-to-complete`
- `complete-to-ship`

For benchmark waves, add wave-specific gates:

- `gate-a-preflight`
- `gate-b-canary`
- `gate-c-pilot`
- `gate-c-cleanup-or-quarantine`
- `gate-d-final-wave`

## Gate Check Output

Every gate check should report:

```text
Gate: <gate_id>
Status: passed|blocked|needs-human|quarantined
Evidence: <N required, N proven, M missing>
Blockers: P0=<n> P1=<n> P2=<n> P3=<n> P4=<n>
Allowed next action: <command or none>
```

## Deterministic Check

When a manifest exists, run the bundled checker before phase advancement:

```powershell
python .github\skills\kb-gate\scripts\check_gate_ledger.py <manifest-path> --gate <gate-id> --allowed-next "<expected next action>"
```

Use `--allow-quarantine` only when the next phase explicitly excludes the
quarantined scope from scoring, shipping, or claims.
