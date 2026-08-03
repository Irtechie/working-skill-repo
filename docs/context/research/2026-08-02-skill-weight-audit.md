# Skill weight audit

`date: 2026-08-02`

The same deterministic loop applied to tests, applied to all 44 skills. The
test audit asked "which statements does only this test cover". The skill
analogue is "which prose does only this skill contain, and who actually
references it".

## Method

For each skill: strip YAML frontmatter, normalise `SKILL.md` to lowercase
word sequences, and form 8-word shingles. A shingle is **unique** when it
appears in no other skill. Nearest neighbour is the highest Jaccard
similarity against every other skill. Inbound references are counted by
scanning the whole repo for the skill name, excluding the skill's own
directory.

### Validating the measurement before trusting it

A near-zero overlap result is exactly what a broken matcher produces, so the
tool was checked against a known-sharing case before any conclusion was drawn:

```text
total distinct shingles:            50,217
shingles shared by >= 2 skills:        176   (0.35%)
most widely shared (5 skills each):
  "go run cmd kbcheck accept --check check json --trace kb trace jsonl"
```

The matcher does detect sharing. What it finds shared are **canonical proof
commands**, which is correct duplication, not prose bloat.

## Headline result: there is no consolidation win

44 skills, 619.1 KB total.

| Measure | Value |
|---|---|
| Highest pairwise similarity between any two skills | 0.7% |
| Lowest unique-content fraction of any skill | 97% |
| Shingles shared by two or more skills | 0.35% |

Every skill is 97-100% textually unique. **The "merge duplicated skills"
hypothesis is dead on evidence.** No byte-level reduction is available from
removing redundant prose, and shrinking the bundle by merging skills would
delete distinct content, not duplication.

This is the opposite shape from the test audit, where cost was concentrated
and removable. Here the surface is already lean per skill; weight comes from
the *number* of skills and their routing clarity, not from repeated text.

### What this measurement cannot tell you

Shingles measure **textual** duplication, not **conceptual** overlap. Two
skills can describe adjacent responsibilities in entirely different words and
score zero similarity while still being hard for an agent to choose between.
That ambiguity, not byte count, is the real weight cost of a skill bundle.

The conceptually adjacent pairs the numbers cannot adjudicate:

| Pair | Question for semantic review |
|---|---|
| `kb-fix` / `kb-troubleshoot` | When is a bug "known" vs "needs diagnosis"? Can an agent tell at routing time? |
| `kb-qa` / `kb-functional-test` | Both own test quality. Which owns the decision to write a functional test? |
| `kb-check` / `kb-repair` | Verification vs surgical fix loop. Is the boundary observable? |
| `kb-map` / `kb-map-bootstrap` | Bootstrap is invoked by map; confirm it is never independently routable. |
| `kb-review` / `kb-finalize` | finalize invokes review. Confirm no overlap in what each gates. |
| `kb-simplify` / `kb-architecture-deepening` | Both restructure committed code. Boundary is stated; verify it holds. |

## Real defect found: caller claims that cannot be true

`config/skill-guidance-audit.json` records a `callers` list per skill used to
justify retention. Comparing every claim against actual references found 8 of
44 skills where a claimed caller does not reference the skill. They are not
all the same severity.

### Impossible (config defect)

| Skill | Claimed caller | Reality |
|---|---|---|
| `todo-create` | `kb-review` | declares `disable-model-invocation: true` -- it **cannot** be invoked by a skill |
| `todo-triage` | `kb-finalize` | declares `disable-model-invocation: true` -- it **cannot** be invoked by a skill |

These are not stale references. A skill that disables model invocation is
user-only by construction, so no skill caller can ever exist. The config
justifies keeping both with a relationship that is impossible.

### Transitive recorded as direct (precision issue)

| Skill | Claimed | Actual path |
|---|---|---|
| `kb-review` | `kb-complete` | `kb-complete` -> `kb-finalize` -> `kb-review` |
| `kb-repair` | `kb-work` | `kb-work` -> `kb-qa` -> `kb-repair` |

The routing works; the config overstates directness. Low severity.

### Unwired (worth a decision)

`kb-gate` claims callers `kb-plan`, `kb-work`, and `kb-complete`. None of the
three reference it, directly or transitively. Its own description says it
governs "brainstorm to plan, plan to work, work to complete", but the skills
that actually reference it are `kb-brainstorm`, `kb-epic`, `kb-finalize`, and
`kb-review`. Either the plan/work/complete gates are enforced somewhere else,
or that policy is documented but not wired.

Same shape, lower confidence: `kb-regression-snapshot` (claims `kb-work`,
`kb-finalize`; actual `kb-eval-map`), `kb-handoff` (claims `kb-goal`; actual
`kb-task`, `kb-troubleshoot`), `kb-research` (claims `kb-plan`; actual
`ce-compound`, `kb-brainstorm`, `kb-epic`, `kb-fix`).

## Verdict legend

| Verdict | Meaning |
|---|---|
| `keep` | Textually unique, referenced or deliberately user-only. No action. |
| `fix-config` | Retention justified by a caller relationship that is wrong or impossible. |
| `semantic-review` | Conceptually adjacent to another skill; numbers cannot decide. |
| `wire-or-document` | Documented routing role that nothing actually invokes. |

## Per-skill data

Sorted by size. `uniq%` = share of this skill's 8-word shingles that appear
in no other skill. `in:skill/doc/code` = inbound references by source.

| KB | lines | uniq% | nearest (sim) | in:skill/doc/code | verdict | skill |
|---:|---:|---:|---|---|---|---|
| 68 | 226 | 100 | `kb-ship` (0.2%) | 2/9/4 | `keep` | `pr-review-workbench` |
| 60.2 | 291 | 100 | `ce-compound` (0.2%) | 2/14/4 | `keep` | `ce-compound-refresh` *(user-only)* |
| 55.6 | 177 | 99 | `kb-goal` (0.3%) | 5/25/5 | `wire-or-document` | `kb-gate` |
| 38.5 | 467 | 100 | `ce-compound-refresh` (0.2%) | 6/18/5 | `keep` | `ce-compound` |
| 28.9 | 221 | 99 | `kb-map` (0.3%) | 5/16/4 | `semantic-review` | `kb-map-bootstrap` |
| 26.6 | 314 | 99 | `kb-map` (0.2%) | 10/46/12 | `keep` | `kb-start` |
| 25.7 | 228 | 99 | `kb-land` (0.2%) | 23/152/47 | `keep` | `kb-work` |
| 23.9 | 151 | 99 | `kb-finalize` (0.5%) | 6/47/10 | `semantic-review` | `kb-review` |
| 23.6 | 307 | 99 | `kb-map-bootstrap` (0.3%) | 16/53/19 | `semantic-review` | `kb-map` |
| 20.3 | 363 | 99 | `kb-gate` (0.3%) | 3/34/9 | `keep` | `kb-goal` |
| 15 | 320 | 100 | `kb-configure` (0.1%) | 15/82/21 | `keep` | `kb-complete` |
| 14.5 | 278 | 98 | `kb-functional-test` (0.6%) | 6/18/7 | `semantic-review` | `kb-qa` |
| 13.1 | 122 | 100 | `` (0%) | 1/26/7 | `wire-or-document` | `kb-regression-snapshot` |
| 12.9 | 208 | 100 | `kb-map` (0%) | 2/18/12 | `keep` | `kb-models` |
| 12.7 | 200 | 98 | `kb-qa` (0.6%) | 8/25/8 | `semantic-review` | `kb-functional-test` |
| 11.3 | 260 | 99 | `kb-check` (0.6%) | 20/91/18 | `keep` | `learn` |
| 10.6 | 117 | 100 | `` (0%) | 6/39/15 | `keep` | `document-review` |
| 10.3 | 242 | 100 | `kb-fix` (0.2%) | 3/17/11 | `keep` | `kb-epic` |
| 10.2 | 142 | 99 | `kb-fix` (0.4%) | 4/18/7 | `semantic-review` | `kb-troubleshoot` |
| 9.7 | 192 | 97 | `kb-repair` (0.7%) | 14/98/17 | `semantic-review` | `kb-check` |
| 9 | 205 | 100 | `` (0%) | 19/124/36 | `keep` | `kb-plan` |
| 8.8 | 187 | 99 | `kb-work` (0.2%) | 5/32/12 | `keep` | `kb-ship` |
| 8.3 | 231 | 100 | `kb-functional-test` (0.1%) | 2/17/3 | `keep` | `kb-eval-map` |
| 8.1 | 110 | 100 | `kb-executive-brief` (0.1%) | 4/9/4 | `keep` | `kb-compact` |
| 8.1 | 136 | 99 | `kb-gate` (0.2%) | 10/41/14 | `keep` | `kb-brainstorm` |
| 7.9 | 193 | 99 | `kb-review` (0.5%) | 14/35/12 | `semantic-review` | `kb-finalize` |
| 7 | 168 | 98 | `kb-check` (0.7%) | 3/17/6 | `semantic-review` | `kb-repair` |
| 6.1 | 152 | 100 | `kb-architecture-deepening` (0.4%) | 1/3/3 | `semantic-review` | `kb-simplify` |
| 5.8 | 109 | 100 | `kb-goal` (0%) | 2/15/3 | `keep` | `kb-task` |
| 5.8 | 118 | 100 | `` (0%) | 1/11/3 | `keep` | `kb-memory-review` |
| 5.6 | 150 | 100 | `learn` (0.1%) | 3/50/5 | `keep` | `evolve` |
| 5.4 | 118 | 100 | `` (0%) | 2/5/4 | `wire-or-document` | `kb-handoff` |
| 5.4 | 50 | 100 | `` (0%) | 0/8/3 | `fix-config` | `todo-create` *(user-only)* |
| 5 | 146 | 100 | `` (0%) | 0/7/3 | `keep` | `repo-critic` |
| 4.8 | 119 | 99 | `kb-work` (0.2%) | 3/14/8 | `keep` | `kb-land` |
| 4.4 | 71 | 100 | `` (0%) | 0/8/3 | `keep` | `safe-shell-quoting` |
| 4.3 | 115 | 100 | `` (0%) | 3/12/4 | `keep` | `kb-first-principles` |
| 4.1 | 100 | 100 | `kb-complete` (0.1%) | 1/9/9 | `keep` | `kb-configure` |
| 3.6 | 82 | 98 | `kb-troubleshoot` (0.4%) | 7/19/26 | `semantic-review` | `kb-fix` |
| 3.1 | 74 | 100 | `kb-compact` (0.1%) | 0/8/5 | `keep` | `kb-executive-brief` |
| 2.2 | 62 | 99 | `kb-simplify` (0.4%) | 2/9/5 | `semantic-review` | `kb-architecture-deepening` |
| 1.9 | 76 | 100 | `` (0%) | 7/11/5 | `wire-or-document` | `kb-research` |
| 1.5 | 40 | 100 | `` (0%) | 2/74/4 | `keep` | `tdd` |
| 1.4 | 50 | 100 | `` (0%) | 0/8/3 | `fix-config` | `todo-triage` *(user-only)* |



---

# Part 2: Semantic pass (2026-08-02, later same day)

Part 1 above measured **text overlap** and found a null (max 1% similarity).
That null is real but nearly worthless: 8-word shingles detect copy-paste, not
the same policy restated in different words. Part 1 should not have been read
as "nothing to gain."

This part re-ran the loop against *concepts* and *portability*. It found no
large consolidation win either -- but it found a **correctness** class that the
weight framing had entirely hidden.

## Method

1. Heading census across all 44 skills (437 headings, 368 distinct).
2. Concept probes: regex families for the same policy however worded.
3. Portability scan: which skills invoke artifacts that do not ship.
4. Four subagents deep-read all 44 skills in four clusters.
5. **Every subagent claim re-verified by hand before acceptance.**

## Concept census

Text overlap said "unique." Concept overlap says otherwise:

| Concept | Skills carrying it | Total mentions |
|---|---:|---:|
| proof-before-claim | 34 / 44 | 216 |
| handoff lifecycle | 25 | 122 |
| todo.md board contract | 20 | 68 |
| scope control | 17 | 36 |
| agent-owned verification | 16 | 29 |
| kb-map / memory preflight | 14 | 42 |
| blocker / human-required | 13 | 32 |
| deterministic proof commands | 13 | 30 |
| P0/P1 severity ladder | 9 | 59 |
| delivery authorization | 7 | 22 |

Naming drift also exists: `Use When`, `When To Run`, and `When to Use` are
three headings for one idea.

**This repetition is mostly NOT waste.** See the portability finding: skills
ship standalone, so a skill that omits its own policy is broken, not lean.

## The finding that matters: the bundle ships without its tooling

`bin/kb-install.mjs` copies whole skill directories to the chosen roots. It
does **not** copy `cmd/kbcheck`. Verified: no reference to `cmd` or `kbcheck`
exists anywhere in the installer.

`AGENTS.md` is copied only under `if (target === "repo")` (line 1043). The
default target is `all`, which expands to `["codex","copilot","agents"]` --
**`repo` is not included**. So a default personal-machine install receives
skills and agents, but neither `AGENTS.md` nor `cmd/kbcheck`.

Consequence: any skill instruction of the form `go run ./cmd/kbcheck ...` is a
dead command on a default install unless the skill says what to do without it.

| | count |
|---|---:|
| skills referencing `cmd/kbcheck` | 14 |
| total references | 42 |
| guarded by availability/fallback language | 12 |
| **unguarded (dead standalone)** | **30** |
| skills with zero guards | 7 |

Unguarded by skill: `kb-check` 10, `kb-complete` 3, `kb-goal` 3, `kb-repair` 3,
`kb-work` 3, `kb-start` 2, `kb-troubleshoot` 2, `evolve` 1, `kb-gate` 1,
`kb-ship` 1, `learn` 1.

The correct pattern already exists and is documented. `kb-ship` line 17:

> Validate through `kb-gate`; when repo `cmd/kbcheck` exists, also run:

and README lines 603-605 state the design intent explicitly ("Consuming repos
without it use the skill's fail-closed fallback"). The pattern is simply
applied 12 times out of 42.

## Install-path defects

| Defect | Evidence | Impact |
|---|---|---|
| `kb-start` hardcodes one install root | L69: `"$HOME\.copilot\skills\kb-start\scripts\work_queue.ps1"` | Wrong for `~/.codex` and `~/.agents` (2 of 3 default targets) and wrong in-repo. Sits on the pre-mutation hot path. |
| `kb-regression-snapshot` uses repo-relative script paths | L24-26: `.github/skills/kb-regression-snapshot/scripts/...` | Correct in-repo, wrong once installed. Opposite convention to `kb-start`. |

Both conventions cannot be right. Scripts *do* ship (whole directories are
copied), so the skill needs a root-agnostic way to locate its own `scripts/`.

## README claims that do not match the installer

| README | Reality |
|---|---|
| L592 "46 skills" | 44 |
| L599 `.github/instructions/*.instructions.md` installed | never copied at any target |
| L598 `AGENTS.md` in "installed runtime surface" | only `--target repo`; quickstart (L43) uses `--target all` |

No check guards any of these, which is why they drifted.

## Config defect found and fixed

Three audit rows claimed a skill caller for a skill declaring
`disable-model-invocation: true` -- unsatisfiable by construction, not merely
stale. Now enforced by `audit-caller-impossible` in `cmd/kbcheck/skill_guidance.go`.

`todo-create` and `todo-triage` claimed callers that reference them nowhere in
the repo; corrected to user-only. `ce-compound-refresh` was the inverse:
`ce-compound` spends ~40 lines instructing invocation, so the flag was removed
so that guidance is live.

## Subagent reliability (recorded deliberately)

Four subagents were told to quote file:line evidence for every claim. **Three of
four "CRITICAL" findings were false**, and each took under a minute to falsify:

| Claim | Reality |
|---|---|
| `kb-compact` references non-existent `references/response-patterns.md` | The file exists |
| `learn` calls non-existent `cmd/kbcheck learning-adoption` | Subcommand exists (`main.go` L289, L416) |
| `evolve` mandates instinct reads with no bailout | Bailout is at L27-32, six lines below the quoted text |
| `kb-regression-snapshot`: "only SKILL.md ships" | Installer copies whole directories |

Only the deterministic scans survived verification intact. **Do not accept a
subagent finding into a plan without re-running its cheapest check.**

## Verdict on weight

Semantic review confirms Part 1's direction while rejecting its framing:
consolidating skills is not available, because near-duplicate policy text is
what makes each skill survive standalone installation. Measured bloat is roughly
150-250 lines across ~7,500 (2-3%), and realizing it costs more risk than it
returns.

The available work here is correctness, not weight: close the 30 unguarded
tool references, pick one script-path convention, and guard the README claims.
