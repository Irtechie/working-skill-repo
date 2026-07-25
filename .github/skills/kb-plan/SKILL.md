---
name: kb-plan
description: "Break a brainstorm or feature into vertical-slice task plans with dependency DAG, verification strategy, and HITL flags. Default planning workflow for end-to-end vertical slices instead of horizontal phases. Use when the user says 'kb plan', 'plan', 'create a plan', 'plan this', 'slice this', 'break into vertical slices', or wants independently-grabbable tasks."
argument-hint: "[brainstorm path, feature description, or PRD]"
---

# KB Plan - Vertical Slice Decomposition

<!-- Inspired by mattpocock/skills to-issues - credit: github.com/mattpocock/skills -->

Break work into independently executable **vertical slices** (tracer bullets). Each slice cuts through all relevant layers end-to-end. Avoid horizontal phases.

## Quick Start

1. Read the brainstorm, PRD, or feature description.
2. Draft thin end-to-end slices with dependencies and verification modes.
3. Review the breakdown yourself against the source material; ask the user only for blocking decisions.
4. Write one KB manifest plus one plan file per slice.
5. Create or update the manifest `gate_ledger`; `plan-to-work` must be
   `passed` before `kb-work` may execute.
6. After writing the manifest, continue to `kb-work <manifest-path>` only when
   execution was requested or an orchestrator called this plan. Otherwise ask
   once and print the exact next command.
7. Stage or commit only the generated files when the user explicitly asked for a commit.

## Interaction Method

`kb-gate` owns blocking-question policy. Record safe, reversible assumptions;
route unresolved `ask-now`, `research-first`, or material scope/architecture/
safety/verification decisions back through that gate.

Always produce the manifest and slice plans first. Continue directly to
`kb-work <manifest-path>` when the user requested execution or an orchestrator
called planning. Otherwise ask once whether to continue; if not, print the
exact next command.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Check `todo.md` and `docs/brainstorms/` for the active or most recent brainstorm. If exactly one likely source exists, use it and record the assumption. If multiple plausible sources exist, ask which one to use. If none exist, ask: "What feature or work should I slice?"

**If input is a brainstorm path:** Read it thoroughly. This is the source of truth for what to build. Carry forward all decisions, scope boundaries, and requirements.

**If input is a handoff path:** Do source discovery before planning:

1. Read the handoff.
2. Check the handoff for explicit `Brainstorm:`, `Requirements:`, `Source:`, `Manifest:`, or `Plan:` pointers.
3. Check `todo.md` for a source pointer tied to that handoff or feature.
4. Look for matching existing source artifacts under project-root paths only:
   - `docs/brainstorms/*<topic>*`
   - `docs/requirements/*<topic>*` if that folder exists
   - `docs/plans/*<topic>*`
5. If a matching brainstorm or requirements doc exists, read it and use it as the planning source of truth. The handoff becomes restart context, not the primary source.
6. If a matching manifest already exists, ask whether to resume with `kb-work` instead of creating a duplicate plan.
7. If no source exists and the handoff is concrete enough, plan from the handoff and record `source: handoff`.
8. If no source exists and the handoff leaves material product or architecture decisions open, stop and route to `kb-brainstorm`.

**If input is a feature description:** Proceed directly to decomposition.

## Core Rules

### Vertical Slices Only

Each slice must deliver a narrow but complete path through every relevant layer: schema, service, API, UI, tests, docs, or ops as applicable. A completed slice is demoable or verifiable on its own.

```text
WRONG (horizontal phases):
  Task 1: Create database schema
  Task 2: Build service layer
  Task 3: Add API routes
  Task 4: Build frontend

RIGHT (vertical slices):
  Task 1: Award points on lesson completion + show on dashboard
  Task 2: Track streaks (builds on task 1)
  Task 3: Add level progression display
```

### Enabling Slices Are Acceptable

Some work is legitimately enabling infrastructure: migrations, auth plumbing, shared config, repo setup. Allow enabling slices only when:

- They unlock a named downstream slice
- They are the smallest viable prerequisite
- The slice names its immediate consumer(s)

### Live-Steering Slices

For recurring, scheduled, or trend-improvement work routed from `kb-goal`,
include the control-loop fields in the manifest or the first slice plan:

- set point: the invariant, threshold, or direction being driven;
- sensor: the command, query, test, or review signal that measures the gap;
- controller: how the next small reviewable increment is selected;
- actuator: the KB lane, coding agent, or workflow that applies the change;
- disturbances: outside changes the loop must tolerate;
- dampener: optional regression gate that keeps the measured problem from
  getting worse while the loop improves it;
- scope gate, batch size, WIP bound, and steering-memory path.

Do not force this framing onto one-shot feature work. Do not invent separate
sensor, controller, and actuator artifacts when the repo's real toolchain fuses
them; record the fused component and the selection policy. HumanLayer-style CI,
Bun, CodeLayer, or GitHub Actions runners are examples, not KB defaults.

### Every Slice Has a Verification Strategy

| Mode | When | Gate |
|------|------|------|
| `tdd` | Behavior changes, business logic | Define protected oracle first when practical -> prove RED -> implement -> unchanged oracle passes |
| `integration` | Wiring, cross-boundary, API contracts | Integration test proves path works |
| `functional` | User-visible workflow, UI/API/CLI journey, escaped bug | Workflow-level check proves the user path |
| `verification-only` | Config, scaffolding, ops | Builds pass, no regression |
| `hitl` | UX taste, design judgment | Human confirms acceptable |

Also record `test_level` and `functional_risk` for each slice. `kb-functional-test` owns this classification:

- `test_level`: `none`, `unit`, `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`
- `functional_risk`: `none`, `narrow`, `broad`, or `full`

Use `unit` only when a unit test can genuinely prove the changed behavior. Use functional levels when a unit test could pass while the user-visible, API, CLI, browser, persistence, auth/session, streaming, or integration path is broken.

## Process

### 1. Understand the Source Material

Read the brainstorm/PRD/description. Extract:

- What behaviors need to exist
- What the user-visible outcomes are
- What constraints/dependencies exist
- What's explicitly out of scope
- Question Gate state: unresolved `ask-now`/`research-first` blockers, safe
  assumptions, deferred planning questions, and parked forbidden claims.

Planning cannot launder brainstorm ambiguity.

If the source has unresolved `ask-now` or `research-first` items, stop before
decomposition. Write or update the `brainstorm-to-plan` gate as `blocked` or
`needs-human` and set `allowed_next_action` to the smallest repair action, such
as `kb-brainstorm <requirements-path>`.

### 1.5. Research (Parallel)

Use the `kb-map` context already loaded. When material uncertainty remains—
especially security, payments, external APIs, privacy, or unfamiliar framework
behavior—invoke `kb-research` and point affected slices to its evidence. Carry
relevant active landmines into constraints and verification.

### 2. Draft Vertical Slices

Break the work into thin end-to-end slices. For each slice, determine:

- **Title** - short descriptive name
- **What it delivers** - end-to-end behavior description
- **Verification mode** - tdd / integration / verification-only / hitl. For `tdd`, record the oracle path/command before implementation whenever practical.
- **Test level** - none / unit / integration / functional-api / functional-cli / functional-browser / full
- **Functional risk** - none / narrow / broad / full
- **Model tier** - the `small` / `medium` / `large` correction and authority
  tier required if the first implementation attempt fails; Planner is a
  separate orchestration role
- **Model tier reason** - one falsifiable explanation tied to uncertainty,
  blast radius, coupling, reversibility, authority, or verification burden
- **Model requirements** - capabilities, tools, context, risk, and proof shape the work-time selector must consider
- **Escalation triggers** - observable conditions that require a higher tier
- **Workspace isolation intent** - `shared-serial` or `worktree-required`,
  conflict domains, and shared resources. Plans record intent only, never live
  paths, branch names, session IDs, cleanup state, or owner tokens.
- **Blocked by** - which other slices must complete first, or none
- **HITL flag** - does this need human judgment? Most should be `false` if the brainstorm was thorough.
- **Expected files** - best current forecast of files this slice may create or modify, with operation type. Used by `kb-work` as an orientation and review-scope seed, not as a literal allowlist.
- **Impact forecast** - when `kb-map` provided an impact packet, carry forward
  impacted files/symbols, tests, docs, freshness, fallback, and limitations.
  Stale or missing graph evidence must become an explicit file-native fallback,
  not a silent confidence claim.
- **Context packet** - for non-trivial slices, the bounded execution payload:
  memory/source files already checked, deterministic prefetch, constraints,
  acceptance/proof targets, minimum execution tier, allowed tools/search
  policy, and escalation triggers. Tiny doc-only or mechanical slices may omit it with a
  one-line reason.

Each entry in `expected_files` should specify:
  - `path` — the file path
  - `op` — `create`, `edit`, or `delete`
  - `scope` — one-line description of what specifically changes (for `edit` operations)

This helps agents start surgically instead of rediscovering the whole repo. It cannot perfectly predict implementation reality; `kb-work` records discovered files in the scope ledger when current code requires touching files not forecast here.

Impact forecasts are also forecasts. They guide conflict domains, proof
selection, and review scope, but they do not replace source reading or
functional proof.

When the consuming repo includes `cmd/kbcheck`, validate a JSON packet with:

```powershell
go run ./cmd/kbcheck context-packet --packet <packet.json>
```

If the validator is not installed, verify the required packet fields directly
and record `packet-validator: unavailable`; do not pretend deterministic
validation ran. The skill bundle does not require consumers to install the Go
maintainer harness.

Packets are execution inputs, not another task database. Manifest status,
goal/run state, proof traces, and `todo.md` continue to own lifecycle state.

### 3. Validate the Breakdown

Check the proposed breakdown against:

- Granularity: each slice is independently executable and reviewable.
- Dependencies: blockers are necessary, not accidental.
- Verification: each slice has agent-runnable tests/checks where possible.
- Functional coverage: user-visible or cross-boundary slices include a narrow functional check unless explicitly justified.
- Test-level classification: each slice says whether unit, integration, API/CLI/browser functional, or full-suite proof is required.
- HITL: human flags are limited to credentials, external systems, subjective approval, or true decisions.
- Expected files: each slice declares likely touched files and scope, with enough specificity to guide the first edit. Do not pretend the list is exhaustive when current code may reveal adjacent files.
- Context packet: material slices provide bounded context or record why a tiny
  slice does not need one. A packet must not embed raw chat history or broad
  tool catalogs.

Ask the user only when a material decision remains. Otherwise proceed and record assumptions.

Run `kb-gate` before writing final plans when validation surfaces P0/P1/P2/P3 issues. P0/P1 block work, but the agent should rectify safe/actionable blockers before asking the user. For P2/P3, ask whether to rectify all fixable issues before moving on.

Before handing off to `kb-work`, write a `plan-to-work` gate in the manifest.
Load `kb-gate/references/gate-ledger.md` if needed. The gate must include proof
for: manifest path, every slice plan path, dependency DAG validation, acceptance
criteria, `expected_files`, verification mode, `test_level`, `functional_risk`,
`model_tier`, model requirements, escalation triggers, HITL classification, any protected oracle policy,
and any objective-contract fields. If any proof is missing, set
`status: blocked` and do not invoke `kb-work`.

### Model Tier Contract

`model_tier` records the minimum execution capability the orchestrator judges
necessary for the slice. It is not a permanent worker assignment and not a
proof level. Verification requirements stay the same regardless of tier.
`kb-work` decides whether the current orchestrator must retain execution because
its reasoning, context, tools, or authority are required, or whether one
qualified worker may execute the bounded slice. The plan does not record a
model, route alias, provider, or `attempt_tier`.

The planner never chooses a native model, extra-route alias, provider, adapter,
endpoint, or transport. The current master resolves live native routes and any
saved project source preference immediately before work. The actual route
belongs in the receipt. Only run-scoped `require <model>` hard-pins.

| Tier | Good fit | Do not assign |
|---|---|---|
| `small` | narrow mechanical code edits, straightforward tests, local docs updates with clear examples | ambiguous architecture, cross-boundary behavior, user-visible workflows without stronger review |
| `medium` | ordinary vertical slices, focused refactors, integration wiring with clear acceptance criteria | high-risk architecture/security/data migrations, unresolved product calls |
| `large` | decomposition, hard debugging, architecture/security decisions, broad migrations, final synthesis/review | tasks with no executable proof path |

Every runnable slice must include `model_tier_reason`, non-empty
`model_requirements`, and observable `escalation_triggers`. A tier label without
those fields fails `manifest-contract`. The reason must state why that minimum
capability is needed, not merely name a task type.

Legacy `tiny` remains readable as a `small`-lane hint only. When unsure, choose
the higher tier. Subjective design direction, philosophy/policy judgment,
unresolved architecture, weak proof, and security/auth/data-boundary decisions
normally require the current orchestrator or HITL. Straightforward code is not
enough by itself to justify delegation or a lower tier.

### Workspace Isolation Contract

For manifests that may execute mutating slices concurrently, add
`workspace_isolation_contract` to the manifest and give each runnable slice:

- `workspace_mode: shared-serial` when the coordinator should run the slice in
  the current checkout, one mutator at a time.
- `workspace_mode: worktree-required` when the slice may run beside another
  mutating slice, the source checkout is dirty, or isolation is explicitly part
  of the acceptance criteria.
- `conflict_domains` describing files, prefixes, generated outputs, graph/index
  namespaces, browser/port/database resources, skills, or lifecycle surfaces.
- `shared_resources` for anything that must be serialized or explicitly
  namespaced, such as `git:integration-owner`, `graph:index`,
  `browser:4110`, or `sync:global-skills`.

The planner does not choose a worktree path, branch, session, cleanup command,
or integration order. `kb-work` resolves those from live Git state after it
acquires the atomic slice lease.

If graph routing is part of the plan, add `impact_packet_contract` only when
the manifest can point to packet files or explicit `no_impact_packet_reason`
fallbacks. Keep legacy manifests readable by omitting that opt-in contract when
no packet is available.

### Objective Done Contract

For goal-like, autonomous, long-running, or "continue until done" work, add an
objective contract to the manifest. This makes completion an observable check,
not an agent assertion.

- Set `objective_contract: true`.
- Add a top-level `done_check` with the command, artifact, or gate that proves
  the whole objective is done.
- Add `proof_check` to every slice that can be machine-checked.
- Use `no_check_reason` only for `verification-only` or `none` slices where no
  executable proof exists; the reason must be explicit and human-auditable.
- Keep manifests route-neutral. `kb-plan` records `model_tier`; `kb-work`
  resolves and chooses the actual route at dispatch time and records it in the
  run receipt. Legacy `model_route` values may remain readable as hints only.

In the manifest template, the model selection contract makes ownership
explicit. It forbids automatic downward routing and silent fallback between the
current orchestrator and delegated execution. AMR experiments are separate from
normal manifest execution.

If no honest objective-level check exists yet, do not fake one. Either plan a
slice that creates the check first, or record a human-approved exception before
`kb-work` starts.

### 4. Generate Plan Files

Create a manifest and individual slice plans.

#### Manifest: `docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md`

```yaml
---
type: kb-manifest
kb_id: kb-YYYY-MM-DD-<name>
brainstorm_path: docs/brainstorms/<source-file>.md
created: YYYY-MM-DD
status: active
workflow_shape: "<direct-chat|single-skill-edit|skill-bundle-change|pipeline-change|multi-stream-epic>"
objective_contract: true
done_check:
  kind: command_exit
  command: "<single command, gate, or artifact check that proves the whole KB objective is done>"
  expect: 0
  why: "<what completion claim this proves>"
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
  amr_required: false
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "<requirements path exists>"
      - "Question Gate classification exists"
      - "Resolve Before Planning is empty"
      - "no unresolved ask-now or research-first items remain"
      - "safe assumptions, deferred planning questions, and parked items are recorded"
    proof:
      - docs/brainstorms/<source-file>.md
    blockers: []
    passed_at: "<timestamp>"
    allowed_next_action: "kb-plan <requirements-path>"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "<manifest path exists>"
      - "<all slice plan paths exist>"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier"
      - "objective_contract manifests have done_check and each slice has proof_check or a justified no_check_reason"
    proof:
      - docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md
      - docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
    blockers: []
    passed_at: "<timestamp>"
    allowed_next_action: "kb-work <manifest-path>"
slices:
  - id: slice-001
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: none
    model_tier: medium
    model_tier_reason: "<why this authority tier is required>"
    model_requirements: ["<tools/context/risk/proof capability>"]
    escalation_triggers: ["<observable reason to move higher or re-plan>"]
    workspace_mode: shared-serial
    conflict_domains: ["<file:path-or-resource>"]
    shared_resources: ["git:integration-owner"]
    proof_check:
      kind: command_exit
      command: "<narrowest deterministic command or artifact check for this slice>"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles:
      - path: "tests/<behavior>.test.<ext>"
        role: "behavior oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-002
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-002-<type>-<name>-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-browser
    functional_risk: narrow
    model_tier: large
    proof_check:
      kind: command_exit
      command: "<narrowest deterministic command or artifact check for this slice>"
      expect: 0
    hitl: false
    status: pending
    notes: ""
---

# KB: <Feature Name>

## Origin
Brainstorm: `<brainstorm_path>`

## Workflow Shape

`<workflow_shape>` - why this shape fits.

## Slice Overview
| # | Slice | Blocked By | Verification | HITL | Status |
|---|-------|------------|--------------|------|--------|
| 1 | <title> | - | tdd | no | pending |
| 2 | <title> | slice-001 | tdd | no | pending |
| 3 | <title> | - | integration | no | pending |
```

#### Individual Slice Plans: `docs/plans/YYYY-MM-DD-NNN-<type>-<name>-plan.md`

Each slice plan uses standard ATV plan format with additional frontmatter:

```yaml
---
kb_id: kb-YYYY-MM-DD-<name>
slice_id: slice-NNN
title: "<title>"
blockers: []
verification: tdd
test_level: unit
functional_risk: none
model_tier: medium
model_tier_reason: "<why this authority tier is required>"
model_requirements: ["<tools/context/risk/proof capability>"]
escalation_triggers: ["<observable reason to move higher or re-plan>"]
workspace_mode: shared-serial
conflict_domains: ["<file:path-or-resource>"]
shared_resources: ["git:integration-owner"]
proof_check:
  kind: command_exit
  command: "<narrowest deterministic command or artifact check for this slice>"
  expect: 0
hitl: false
expected_files:
  - path: ""
    op: edit
    scope: "what specifically changes"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---
```

The plan body should include:

- What to build, expressed as end-to-end behavior
- Acceptance criteria
- Minimum execution tier and why its reasoning, context, tools, trust, and
  authority are sufficient for the slice
- Expected files (must match `expected_files` in frontmatter as the initial forecast; actual touched files may expand during `kb-work` when justified by the acceptance criteria and recorded in the scope ledger)
- Test scenarios specific enough for TDD or integration verification
- Proof check: the command, artifact, browser/API/CLI assertion, or accepted
  trace that must exist before `kb-work` marks the slice done
- Protected oracle candidates when expected behavior is known before implementation: tests, fixtures, scorers, snapshots, or contract files that should be written or selected first, proven RED when practical, and protected from mutation with SHA before implementation continues
- Test inputs needed to run those scenarios without asking the user to manually test later
- Scope boundary: what this slice does not include
- Dependencies and why they are needed
- HITL question if `hitl: true`

If verification needs realistic input values, include them in frontmatter:

```yaml
test_inputs:
  - name: "<input name>"
    source: user|fixture|env|generated
    required_for: "<acceptance criterion or QA step>"
    value: "<literal value, fixture path, env var name, or TODO-human>"
```

Only mark `hitl: true` when the human step is truly required. Do not use HITL for checks the agent can run with provided inputs.

Use `protected_oracles` when a slice has a known behavior target before
implementation. Each entry should name the oracle file, its role, and the update
policy. `kb-work` fills or verifies the SHA after RED/protection. If the correct
oracle cannot be known until implementation reveals the interface, leave
`protected_oracles: []` and explain the verification strategy in the plan body.

### 5. Update Todo and Handoffs

Update the existing `todo.md` with a compact manifest pointer and slice status.
If project memory is missing, route to `kb-map`; `kb-map-bootstrap` owns the
templates and layout. Create an active handoff only when another session needs
a restart packet, and link the manifest instead of copying its contents.

### 6. Validate Output

- Confirm every `blockers` entry references an existing slice ID.
- Confirm no dependency cycles exist.
- Confirm every slice has a verification mode and acceptance criteria.
- Confirm the manifest has a `plan-to-work` gate with `status: passed` or `status: blocked`; never leave it absent or pending.
- Confirm every generated plan path is listed in the manifest.
- Confirm the manifest body table matches the YAML frontmatter.

## Success Criteria

- The manifest is a valid DAG with no missing blockers or cycles.
- Each slice is independently grabbable and has a clear verification gate.
- Enabling slices name their immediate downstream consumers.
- Generated paths are precise enough for `kb-work` to resume without rediscovery.
- No unrelated existing plans are staged or changed.

## Integration with Other Skills

- **Input from:** `kb-brainstorm` or a clear feature description.
- **Deepening:** Use `kb-research` only for individual slices with material unresolved uncertainty.
- **Execution:** `kb-work` runs all slices in order when invoked, or can pick up one slice at a time
- **Verification:** Each `tdd` slice carries protected-oracle proof in the manifest; load the standalone `tdd` skill only for explicit test-first coaching.
- **Protected oracles:** Known behavior targets can be frozen before implementation so tests, fixtures, scorers, snapshots, or contracts cannot be rewritten silently
