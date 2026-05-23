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
5. Stop after writing the manifest unless the user invoked `klfg` or explicitly asked to execute.
6. Stage or commit only the generated files when the user explicitly asked for a commit.

## Interaction Method

Default to non-interactive planning when the source material is clear. Use the platform's blocking question tool only when an answer changes behavior, scope, acceptance criteria, risk, or verification.

When assumptions are safe and reversible, record them in the manifest instead of stopping. Ask one concise question only for material uncertainty.

Phase boundary: `kb-plan` produces a manifest and slice plans. It does not automatically invoke `kb-work` unless the user explicitly asked for execution or an orchestrator such as `klfg` called it.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Check `todo.md` and `docs/brainstorms/` for the active or most recent brainstorm. If exactly one likely source exists, use it and record the assumption. If multiple plausible sources exist, ask which one to use. If none exist, ask: "What feature or work should I slice?"

**If input is a brainstorm path:** Read it thoroughly. This is the source of truth for what to build. Carry forward all decisions, scope boundaries, and requirements.

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

### Every Slice Has a Verification Strategy

| Mode | When | Gate |
|------|------|------|
| `tdd` | Behavior changes, business logic | Failing test -> implement -> passes |
| `integration` | Wiring, cross-boundary, API contracts | Integration test proves path works |
| `functional` | User-visible workflow, UI/API/CLI journey, escaped bug | Workflow-level check proves the user path |
| `verification-only` | Config, scaffolding, ops | Builds pass, no regression |
| `hitl` | UX taste, design judgment | Human confirms acceptable |

## Process

### 1. Understand the Source Material

Read the brainstorm/PRD/description. Extract:

- What behaviors need to exist
- What the user-visible outcomes are
- What constraints/dependencies exist
- What's explicitly out of scope

### 1.5. Research (Parallel)

Run lightweight research to ground slice design in reality:

- `repo-research-analyst`: similar features, routes, components, tests, commands, conventions.
- `learnings-researcher`: relevant `docs/solutions/`, prior fixes, and known failures.
- Local memory check: `docs/context/PROJECT.md`, relevant `docs/context/architecture/*`, and `docs/context/research/*`.

Run independent agents/reads/searches in parallel when the platform supports it. If named agents are unavailable, do the same work with native search/read tools.

**Research decision:** Based on findings, decide if external research is needed.

- **High-risk topics** (security, payments, external APIs, data privacy) → always research externally
- **Strong local patterns exist** → skip external research
- **Unfamiliar territory** → research externally

**If external research is warranted**, use `kb-research` and write a reusable research note before finalizing slices.

Optional specialist checks before finalizing slices:

- `spec-flow-analyzer` when the feature has multi-step user flows or unclear edge cases.
- `security-sentinel` or `security-reviewer` when auth, permissions, secrets, public endpoints, payments, PII, external callbacks, or user input are involved.
- `adversarial-reviewer` when the plan is architecture-shaping, high-risk, or large enough that assumption/cascade failures are likely.
- `architecture-strategist` when subsystem boundaries, framework migration, or long-term architecture direction are being set.

Carry research findings forward into slice plans — each slice should reference relevant patterns, gotchas, and file paths discovered here.

### 2. Draft Vertical Slices

Break the work into thin end-to-end slices. For each slice, determine:

- **Title** - short descriptive name
- **What it delivers** - end-to-end behavior description
- **Verification mode** - tdd / integration / verification-only / hitl
- **Blocked by** - which other slices must complete first, or none
- **HITL flag** - does this need human judgment? Most should be `false` if the brainstorm was thorough.
- **Expected files** - which files this slice will create or modify, with operation type. Used by `kb-work` for diff-scope verification and edit-safety.

Each entry in `expected_files` should specify:
  - `path` — the file path
  - `op` — `create`, `edit`, or `delete`
  - `scope` — one-line description of what specifically changes (for `edit` operations)

This prevents agents from regenerating files from the plan spec instead of surgically editing current code.

### 3. Validate the Breakdown

Check the proposed breakdown against:

- Granularity: each slice is independently executable and reviewable.
- Dependencies: blockers are necessary, not accidental.
- Verification: each slice has agent-runnable tests/checks where possible.
- Functional coverage: user-visible or cross-boundary slices include a narrow functional check unless explicitly justified.
- HITL: human flags are limited to credentials, external systems, subjective approval, or true decisions.
- Expected files: each slice declares likely touched files and scope.

Ask the user only when a material decision remains. Otherwise proceed and record assumptions.

Run `kb-gate` before writing final plans when validation surfaces P0/P1/P2/P3 issues. P0/P1 block work, but the agent should rectify safe/actionable blockers before asking the user. For P2/P3, ask whether to rectify all fixable issues before moving on.

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
slices:
  - id: slice-001
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
    blockers: []
    verification: tdd
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-002
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-002-<type>-<name>-plan.md
    blockers: [slice-001]
    verification: tdd
    hitl: false
    status: pending
    notes: ""
---

# KB: <Feature Name>

## Origin
Brainstorm: `<brainstorm_path>`

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
hitl: false
expected_files:
  - path: ""
    op: edit
    scope: "what specifically changes"
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
- Expected files (must match `expected_files` in frontmatter — these are the files this slice is allowed to touch)
- Test scenarios specific enough for TDD or integration verification
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

### 5. Update Todo and Handoffs

After generating plan files, update `todo.md` — the human-visible live execution board.
Create or update a compact handoff file under `docs/handoffs/active/` only when a future session needs a restart packet.

**If `todo.md` doesn't exist**, create it with this template:

```markdown
# Todo

## Rules
- Keep this file current and small.
- Active, blocked, parked, and human-required work belongs here.
- Completed work moves to `todo-done.md`.
- Detailed handoffs live under `docs/handoffs/`; link them here instead of pasting full content.
- Refresh cold or parked work older than 72 hours before execution.
- When all active todos are done, check the handoff queue.

## Objective

## Current Focus

## Current Truth

## Active Work

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

## Parked / Cold Storage

## Blocked

## Work Log
```

**Add an active work section** for the new KB workflow:

```markdown
### <Feature Name> (kb-YYYY-MM-DD-name)

Source: `docs/brainstorms/<file>.md`
Manifest: `docs/plans/<manifest>.md`

| # | Slice | Blocked By | Verification | Status |
|---|-------|------------|--------------|--------|
| 1 | <title> | - | tdd | ⬜ pending |
| 2 | <title> | slice-001 | tdd | ⬜ pending |

Done criteria: All N slices done or skipped with reason. Archive summary to `todo-done.md`.
```

**If a restart packet is needed**, create `docs/handoffs/active/YYYY-MM-DD-<feature>.md`:

```markdown
# <Feature Handoff>

Created: YYYY-MM-DD
Last refreshed: YYYY-MM-DD
Status: active
Suggested route: kb-work

## Intent

## Current State

## Next Agent Action

## Human Required

## Pointers

- Project map: docs/context/PROJECT.md
- Manifest: docs/plans/<manifest>.md
- Todo: todo.md

## Staleness Check

Refresh before execution if older than 72 hours.

## Completion Criteria
```

**Board status markers** (superset of manifest statuses):

| Marker | Meaning |
|--------|---------|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified |
| 🔒 blocked | Cannot proceed — reason noted |
| ⊘ skipped | Intentionally skipped with reason |
| 🧑 manual | Needs human action (HITL) |

**Standing sections** (add once, keep across features):

- **💡 Feature Ideas** — not yet brainstormed, human promotes to active
- **📋 Queued Improvements** — approved but not yet planned
- **🧊 Parked / Cold Storage** — discovered work, do not execute until promoted
- **🛑 Human Required** — items only a person can complete (logins, approvals, decisions)
- **📝 Work Log** — short dated entries for cross-session visibility

Omit empty sections. These conventions come from `todo_rules.md` and apply here.

### 6. Validate Output

- Confirm every `blockers` entry references an existing slice ID.
- Confirm no dependency cycles exist.
- Confirm every slice has a verification mode and acceptance criteria.
- Confirm every generated plan path is listed in the manifest.
- Confirm the manifest body table matches the YAML frontmatter.

### 6. Optional Commit

Commit only when the user explicitly asked for it. Stage only the generated manifest and slice plan files, never the whole `docs/plans/` directory:

```bash
git add docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md todo.md docs/handoffs/active/YYYY-MM-DD-<feature>.md
git commit -m "kb-plan: decompose <feature> into N vertical slices"
```

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
- **Verification:** Each slice uses `tdd` skill principles when verification mode is `tdd`
