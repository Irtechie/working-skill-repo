# Skills System

## Purpose

This repo ships 48 `SKILL.md` files under `.github\skills\`. They define the KB
workflow lanes, compound/document-review helpers, todo helpers, and learning/utilities
used by Codex, Copilot, and shared agents.

Supporting reviewer/specialist agents live under `.github\agents\*.agent.md`.

## Grouped Workflow Map

| Group | Skills |
|---|---|
| Entry / routing | `kb-start`, `kb-task`, `kb-map`, `kb-map-bootstrap`, `kb-memory-review`, `kb-goal`, `kb-epic` |
| Requirements / planning | `kb-brainstorm`, `kb-plan`, `kb-gate`, `kb-research`, `kb-first-principles` |
| Execution / repair | `kb-work`, `kb-fix`, `kb-troubleshoot`, `kb-repair`, `tdd` |
| Verification / eval | `kb-check`, `kb-functional-test`, `kb-qa`, `kb-regression-snapshot`, `kb-eval-map` |
| Completion / delivery | `kb-complete`, `kb-finalize`, `kb-review`, `kb-ship`, `kb-land`, `p2d`, `w2d`, `kb-executive-brief`, `pr-review-workbench` |
| Learning / maintenance | `learn`, `evolve`, `kb-cognitive`, `kb-configure`, `kb-models`, `kb-handoff`, `kb-architecture-deepening`, `kb-simplify`, `kb-rehab` |
| Compound / document review | `document-review`, `ce-compound`, `ce-compound-refresh`, `repo-critic` |
| Todo / utility | `todo-create`, `todo-triage`, `safe-shell-quoting`, `gh-copilot-cost-ops` |

Successful planned work continues through this group automatically:
`kb-work -> kb-finalize -> kb-complete`. `kb-complete` then applies the
configured endpoint, invoking `kb-ship` for PR delivery and authorized
`kb-land` for remote-default integration. The chain does not move merge
authority into `kb-work` or `kb-ship`.

## Full Skill Inventory

Each row maps to `.github\skills\<skill>\SKILL.md`.

| Skill | Purpose | Trigger / when to run |
|---|---|---|
| `ce-compound-refresh` | Refresh stale `docs/solutions` learnings against current code | Use after refactors, migrations, dependency upgrades, or drift in learned docs |
| `ce-compound` | Document a recently solved problem into reusable knowledge | Use after a solved bug/pattern worth compounding |
| `document-review` | Optional one-profile requirements/plan review | Use when the main-agent self-check leaves material uncertainty |
| `evolve` | Promote mature instincts into skills | Use when repeated instincts are ready to become durable skills |
| `gh-copilot-cost-ops` | Build cost-ops infrastructure for Copilot usage-based billing | Use for per-PR cost attribution, cost cliff detection, or budget control |
| `kb-architecture-deepening` | Explore where architecture needs deeper modularity | Use for architecture questions, not routine cleanup |
| `kb-brainstorm` | Produce proportional requirements | Use for vague ideas or before planning |
| `kb-check` | Run deterministic repo proof | Use when tests, lint, builds, or scripts should judge correctness |
| `kb-cognitive` | Reduce comprehension effort and select the smallest useful response format | Use when KB docs or outputs are noisy, hard to scan, or need a table/flow to expose relationships |
| `kb-complete` | Single state-aware end-to-end completion command | Use to take work from current state to configured endpoint |
| `kb-configure` | Set optional per-project execution/delivery policy | Use to inspect or change attempts/delivery settings |
| `kb-epic` | Coordinate large multi-workstream initiatives | Use for migrations, rewrites, or large related backlogs |
| `kb-eval-map` | Map repo-native eval surfaces | Use during bootstrap or when eval strategy is unclear |
| `kb-executive-brief` | Generate a low-burden executive brief from structured data | Use for a decision brief or a PR/companion first screen |
| `kb-finalize` | Post-work review/learning/cleanup pipeline | Usually invoked after `kb-work` |
| `kb-first-principles` | Honest pushback / principled disagreement lane | Use when claims are challenged or truth is uncertain |
| `kb-fix` | Small bug-fix lane | Use for narrow known bugs or failing tests |
| `kb-functional-test` | Choose/test functional proof depth | Use when behavior needs e2e/API/browser proof judgment |
| `kb-gate` | Shared phase-gate policy | Use before advancing between KB phases with unresolved risk |
| `kb-goal` | Durable multi-session objective governor | Use for continue-until-done or long-lived objectives |
| `kb-handoff` | Create fresh-session restart packet | Use to pause/resume across sessions |
| `kb-land` | Internal direct-integration phase | Usually invoked by `kb-complete` for direct landing |
| `kb-map-bootstrap` | Deep project-memory bootstrap | Use when memory is missing or badly stale |
| `kb-map` | Cheap project-memory lookup/refresh | Use on startup or when another KB skill needs local context |
| `kb-memory-review` | High-cost memory maintenance pass | Use when memory-maintenance says review is due |
| `kb-models` | Inspect/manage optional model routes | Use to show, discover, or configure optional routes |
| `kb-plan` | Vertical-slice planning workflow | Use when requirements exist and need slice decomposition |
| `kb-qa` | Slice QA gate | Use to run lint/browser/API/CLI checks on slices |
| `kb-regression-snapshot` | Capture/replay passing state | Use after a slice passes and before later slices risk regressions |
| `kb-rehab` | Clean house across a repo and check the reconciliation in | Use to reconcile outstanding work, settle proven-done work, and make the tree current before new work |
| `kb-repair` | Surgical retry loop for QA/lint failures | Invoked by `kb-qa` when checks fail |
| `kb-research` | Reusable research lane | Use when external docs/prior art may change direction |
| `kb-review` | KB-specific structured code review | Use before KB completion or PR-ready delivery |
| `kb-ship` | Checked-in PR delivery lane | Usually invoked by `kb-complete` after review |
| `kb-simplify` | User-invoked maintenance sweep of committed code | Use only on explicit request to simplify, deslop, or de-duplicate; never automatic |
| `kb-start` | Default KB router | Use for fresh sessions, ambiguous work, or “kb” requests |
| `kb-task` | First-principles autonomous task runner | Use for a bounded task to complete end-to-end without choosing a lane |
| `kb-troubleshoot` | Autonomous debugging/self-correction lane | Use when behavior is broken but cause is unclear |
| `kb-work` | Bounded swarm executor for planned slices | Use to execute a `kb-plan` manifest |
| `learn` | Extract recent patterns into instincts | Use after completing work or at session end |
| `p2d` | Plan to done | Use to carry an unplanned idea through `kb-plan` into `w2d` without staged confirmations |
| `pr-review-workbench` | Turn a PR inbox or one PR into a terse, self-contained HTML review workbench | Use to reduce review cognitive load, build a PR walkthrough, or prepare a review decision |
| `repo-critic` | Claims-vs-code evidence critic | Use to audit whether docs/claims match implementation |
| `safe-shell-quoting` | Move fragile shell quoting into scripts | Use for quote-heavy or nested shell commands |
| `tdd` | Explicit test-first compatibility lane | Use for “TDD”, “test first”, or red-green-refactor requests |
| `todo-create` | Create durable work items in `todo.md` | Use when adding active tracked work |
| `todo-triage` | Classify pending findings/todos | Use when deciding what belongs on the active board |
| `w2d` | Work to done | Use after `kb-plan`, or for "take it all the way" / "land it" |

## Closest Files For Common Questions

| Question | Read This |
|---|---|
| “What’s the default user path?” | `.github\skills\kb-start\SKILL.md` |
| “How does project memory work?” | `.github\skills\kb-map\SKILL.md` |
| “How do I bootstrap or refresh repo memory?” | `.github\skills\kb-map-bootstrap\SKILL.md` |
| “What is the proof lane?” | `.github\skills\kb-check\SKILL.md` + `docs/context/architecture/kbcheck.md` |
| “What reviews docs?” | `.github\skills\document-review\SKILL.md` |
| “What reviews code?” | `.github\skills\kb-review\SKILL.md` |

## Removed Capability Parity

| Removed skill | Replacement owner | Preserved behavior | Proof |
|---|---|---|---|
| Generalized duplicate code-review entrypoint | `kb-review` | Standalone and completion review, one broad or replacement specialist profile, report-only/headless/autofix modes | Review contract tests and `kb-review` references |
| Legacy finish alias | `kb-complete` | State-aware plan-to-endpoint completion and configured delivery | Completion route fixtures |
| Legacy strict-pipeline alias | `kb-complete` | Brainstorm, plan, work, finalize, and delivery recovery from current state | Workflow governor and completion tests |

The removed names remain only in `config/removed-skills.json` and factual
historical artifacts. They are not current routes, install targets, or skills.
