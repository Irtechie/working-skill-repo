---
name: learn
description: Extract reusable patterns from recent work into instincts. Run after completing features, fixing bugs, or at session end to capture what the project learned.
---

# /learn — Extract Patterns into Instincts

Analyze recent work (observations, git history, solutions) and extract reusable "instincts" — small learned behaviors with confidence scoring.

## When to Use

- After completing a feature or fixing a bug
- At the end of a coding session
- When you want to capture a pattern you noticed
- Periodically to keep the project's knowledge current

## Execution Flow

### Step 1: Gather Evidence

Run these in parallel to collect data:

1. **Git history** — `git log --oneline -20` for recent commits
2. **Recent diffs** — `git diff HEAD~5..HEAD --stat` for what changed
3. **Observations** — Read `.atv/observations.jsonl` for tool use patterns from hooks
4. **Existing instincts** — Read `.atv/instincts/project.yaml` (create if missing)
5. **Solutions** — Read `docs/solutions/` for documented patterns

### Step 2: Analyze Patterns

Look for recurring patterns across the evidence:

**Code style patterns:**
- Error handling conventions (wrapping, custom types, sentinel errors)
- Naming conventions (variable, function, file naming)
- Import organization preferences
- Comment style and documentation patterns

**Workflow patterns:**
- Test-first vs test-after behavior
- Commit granularity preferences
- Branch naming conventions
- Review practices

**Architecture patterns:**
- Package/module organization
- Dependency injection style
- Interface usage patterns
- Configuration management approach

**Tool usage patterns** (from observations.jsonl):
- Frequently used shell commands
- Common file editing sequences
- Preferred build/test commands

### Step 2.5: Apply Recency Decay

Before creating or updating instincts, apply time-based decay to all existing entries:

1. For each instinct in `.atv/instincts/project.yaml`:
   - Calculate days since `last_seen`
   - Apply decay: `new_confidence = confidence × 0.5^(days_since_last_seen / 90)`
   - Update the confidence value in place

2. **Archive stale instincts:**
   - If decayed confidence falls below 0.3, move the instinct to `.atv/instincts/archive/YYYY-MM-DD-archived.yaml`
   - Add `archived_reason: confidence decayed below 0.3 (last seen: <date>)`
   - Remove from `project.yaml`

3. Write updated confidence values back to `project.yaml` before proceeding to Step 3.

**Half-life rationale:** 90 days balances stability (project conventions rarely change weekly) with freshness (patterns unused for 6+ months are likely obsolete). At 90 days, an unobserved instinct at 0.85 decays to:
- 30 days: 0.68
- 90 days: 0.43
- 180 days: 0.21 (archived)

### Step 3: Create or Update Instincts

For each new pattern discovered, create an instinct entry.
For patterns that match existing instincts, increase confidence and observation count.

**Instinct format** (YAML in `.atv/instincts/project.yaml`):

```yaml
instincts:
  - id: kebab-case-unique-id
    trigger: "when [specific situation]"
    behavior: "do [specific action]"
    confidence: 0.5
    domain: code-style|testing|architecture|error-handling|workflow|tooling
    observations: 1
    first_seen: YYYY-MM-DD
    last_seen: YYYY-MM-DD
    evidence:
      - "commit abc123: wrapped all errors with fmt.Errorf"
      - "observed 3 times in observations.jsonl"
```

**Confidence rules:**
- New instinct starts at 0.5
- Each additional observation: +0.1 (capped at 0.95)
- Contradictory evidence: -0.15
- No observations for 30 days: -0.1
- Minimum: 0.1 (below this, remove the instinct)

**Important constraints:**
- Maximum 50 active instincts per project
- Each instinct must be atomic — one trigger, one behavior
- Triggers must be specific (not "when writing code")
- Behaviors must be actionable (not "write good code")
- Evidence must cite specific commits or observations

### Step 4: Write Results

1. Write updated `.atv/instincts/project.yaml`
2. Ensure `.atv/instincts/` directory exists

### Step 5: Report

Show a summary:

```
Learning complete!

New instincts:
  + always-wrap-errors (0.5) — wrap errors with fmt.Errorf using %w
  + table-driven-tests (0.5) — use table-driven test pattern for Go tests

Updated instincts:
  ↑ prefer-early-returns (0.6 → 0.7) — 1 new observation
  ↑ run-tests-before-commit (0.7 → 0.8) — 2 new observations

Ready to evolve (confidence > 0.8):
  ★ run-tests-before-commit — consider /evolve to generate a skill

Total: X instincts (Y new, Z updated)
Instinct file: .atv/instincts/project.yaml
```

## Notes

- Instincts are project-scoped and committed to the repo — the whole team benefits
- Run `/instincts` to see all learned patterns
- Run `/evolve` when instincts reach high confidence to generate full skills
- The observer hooks in `.github/hooks/copilot-hooks.json` automatically capture tool use data
