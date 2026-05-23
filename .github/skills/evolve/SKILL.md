---
name: evolve
description: Promote mature instincts (confidence > 0.8) into full Copilot skills that get auto-discovered. Clusters related instincts and generates SKILL.md files in .github/skills/.
---

# /evolve — Promote Instincts to Skills

Transform mature instincts into full Copilot skills. When an instinct reaches high confidence through repeated observation, it has proven valuable enough to become a permanent part of the project's skill set.

## When to Use

- When `/instincts` shows patterns with ★ (confidence > 0.8)
- When you want to formalize a recurring pattern into a discoverable skill
- Periodically to graduate well-established project conventions

## Execution Flow

### Step 1: Identify Candidates

Read `.atv/instincts/project.yaml` and filter for:
- Confidence > 0.85
- Observations > 5
- `last_seen` within the last 90 days (rejects stale instincts even if confidence is high)
- Not already evolved (check `.atv/instincts/archive/`)

If no candidates found:
```
No instincts ready to evolve yet.
Run /learn to build confidence, or check /instincts for current status.
Instincts need confidence > 0.85, 5+ observations, and activity within 90 days to evolve.
```

### Step 2: Cluster Related Instincts

Group candidates by domain. Each cluster becomes one skill:

```
Evolution candidates:

  Cluster 1: "Go Error Handling" (error-handling domain)
    ★ always-wrap-errors      0.9  15 obs
    ★ sentinel-errors         0.85  8 obs
    → Will generate: .github/skills/learned-go-error-handling/SKILL.md

  Cluster 2: "Testing Conventions" (testing domain)
    ★ table-driven-tests      0.85  12 obs
    → Will generate: .github/skills/learned-testing-conventions/SKILL.md

Proceed with evolution? (Copilot will generate the skills)
```

### Step 3: Generate SKILL.md

For each cluster, generate a `SKILL.md` file:

```yaml
---
name: learned-[domain-name]
description: "[Auto-generated] Project conventions for [domain] learned from [N] observations across [M] sessions."
---
```

The skill content should:
1. Describe when to apply these conventions
2. List each instinct as a concrete guideline with examples
3. Reference the evidence (commit hashes, observation counts)
4. Mark as auto-generated so humans know to review

**Naming convention:** `learned-` prefix so generated skills are visually distinct from hand-written ones.

**Output path:** `.github/skills/learned-<domain>/SKILL.md`

### Step 4: Archive Evolved Instincts

Move evolved instincts from `project.yaml` to `.atv/instincts/archive/evolved-YYYY-MM-DD.yaml` with metadata:
```yaml
evolved_to: .github/skills/learned-go-error-handling/SKILL.md
evolved_at: 2026-04-06
```

### Step 5: Report

```
Evolution complete!

Generated skills:
  ✅ .github/skills/learned-go-error-handling/SKILL.md (from 2 instincts)
  ✅ .github/skills/learned-testing-conventions/SKILL.md (from 1 instinct)

Archived 3 instincts to .atv/instincts/archive/

These skills will be auto-discovered by Copilot in the next session.
Review the generated files and adjust as needed — they're a starting point.

Remaining instincts: X active (run /instincts to see)
```

## Important Notes

- Generated skills use `learned-` prefix — easy to spot and edit
- Always review generated skills before committing — they're a draft
- Evolved instincts are archived, not deleted, so history is preserved
- If a generated skill doesn't feel right, delete it and the instincts will re-accumulate
