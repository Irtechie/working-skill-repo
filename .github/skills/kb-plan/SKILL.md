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
3. Confirm the breakdown with the user when the platform supports blocking questions.
4. Write one KB manifest plus one plan file per slice.
5. Stage or commit only the generated files when the user explicitly asked for a commit.

## Interaction Method

Use the platform's blocking question tool when available. Ask one question at a time and prefer concise single-select choices. If no blocking question tool exists, ask concise direct questions.

If the user asked for non-interactive planning, make conservative assumptions and record them in the manifest.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Check `docs/brainstorms/` for recent brainstorm documents. If found, ask which one to use. Otherwise ask: "What feature or work would you like to decompose into vertical slices?"

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

- Use the **repo-research-analyst** agent to understand existing patterns related to the feature
- Use the **learnings-researcher** agent to check `docs/solutions/` for relevant institutional knowledge

Launch both in parallel. Focus on: similar features, established conventions, documented gotchas.

**Research decision:** Based on findings, decide if external research is needed.

- **High-risk topics** (security, payments, external APIs, data privacy) → always research externally
- **Strong local patterns exist** → skip external research
- **Unfamiliar territory** → research externally

**If external research is warranted**, also run in parallel:

- Use the **best-practices-researcher** agent for industry patterns
- Use the **framework-docs-researcher** agent for framework-specific guidance

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

### 3. Present and Quiz the User

Show the proposed breakdown as a numbered list. Ask:

- Does the granularity feel right: too coarse, too fine, or right?
- Are dependency relationships correct?
- Should any slices be merged or split?
- Are verification modes correct?
- Are any HITL flags wrong?

Iterate until approved unless the user asked for non-interactive planning.

### 4. Generate Plan Files

Create a manifest and individual slice plans.

#### Manifest: `docs/plans/YYYY-MM-DD-000-kanban-<name>-manifest.md`

```yaml
---
type: kanban-manifest
kanban_id: kb-YYYY-MM-DD-<name>
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
kanban_id: kb-YYYY-MM-DD-<name>
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
---
```

The plan body should include:

- What to build, expressed as end-to-end behavior
- Acceptance criteria
- Expected files (must match `expected_files` in frontmatter — these are the files this slice is allowed to touch)
- Test scenarios specific enough for TDD or integration verification
- Scope boundary: what this slice does not include
- Dependencies and why they are needed
- HITL question if `hitl: true`

### 5. Update the Board

After generating plan files, update `docs/kanban.md` — the human-visible board and multi-agent handoff file.

**If `docs/kanban.md` doesn't exist**, create it with this template:

```markdown
# <Project> — Kanban Board

> The board is the source of truth — not chat history.
> Last updated: <ISO timestamp>

## Rules
- Done features → `kanban-done.md` immediately. Keep this file lean.
- One agent per slice. Claim by setting 🔧. Do not work unclaimed slices.
- Discovered work → Parked first. Human promotes to active.
- Human-required items stay visible until a person completes them.
```

**Add a feature section** for the new KB workflow:

```markdown
---

## 🔧 <Feature Name> (kb-YYYY-MM-DD-name)

Source: `docs/brainstorms/<file>.md`
Manifest: `docs/plans/<manifest>.md`

| # | Slice | Blocked By | Verification | Status |
|---|-------|------------|--------------|--------|
| 1 | <title> | - | tdd | ⬜ pending |
| 2 | <title> | slice-001 | tdd | ⬜ pending |

Done criteria: All N slices done or skipped with reason. Archive to `kanban-done.md`.
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
git add docs/plans/YYYY-MM-DD-000-kanban-<name>-manifest.md docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
git commit -m "kb-plan: decompose <feature> into N vertical slices"
```

## Success Criteria

- The manifest is a valid DAG with no missing blockers or cycles.
- Each slice is independently grabbable and has a clear verification gate.
- Enabling slices name their immediate downstream consumers.
- Generated paths are precise enough for `kb-work` to resume without rediscovery.
- No unrelated existing plans are staged or changed.

## Integration with Other Skills

- **Input from:** `kb-brainstorm`, `deepen-brainstorm`
- **Deepening:** Run `deepen-plan` on individual slices, one at a time
- **Execution:** `kb-work` runs all slices in order, or `kb-work` can pick up one slice at a time
- **Verification:** Each slice uses `tdd` skill principles when verification mode is `tdd`
