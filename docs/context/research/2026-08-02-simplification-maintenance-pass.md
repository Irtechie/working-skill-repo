# Simplification Maintenance Pass

Checked: 2026-08-02
Budget mode: standard

## Question

How should an explicitly invoked maintenance skill decide *which* accumulated
complexity to attack, *when* duplication should become an abstraction, and *how*
to execute the change without producing an unreviewable sweep?

## Findings

### Ranking: health alone is not enough; weight by change frequency

CodeScene's behavioral code analysis combines two dimensions: a quality
dimension (code health) and a relevance dimension (change frequency from
`git log`). A hotspot is the overlap — low health *and* high churn. Tornhill's
framing is the decisive argument against pure static ranking:

> "What if all those modules in react-reconciler have been stable for 5 years?
> ... Starting to speculatively refactor the code there is not only a technical
> risk; it could also be a wasted opportunity."

Fowler states the same principle qualitatively: "crufty but stable areas of code
can be left alone. In contrast, areas of high activity need a zero-tolerance
attitude to cruft, because the interest payments are cripplingly high."

Quantitative support comes from Tornhill & Borg, "Code Red" (IEEE Software,
2022): across 39 proprietary codebases and 30,737 files, low-quality code
contained **15x more defects**, issue resolution took **124% more time**, and
maximum cycle times were **9x longer**.

### Abstraction: the decision rule is knowledge, not text similarity

DRY's original definition (Hunt & Thomas) is about *knowledge*: "Every piece of
knowledge must have a single, unambiguous, authoritative representation within a
system." Two identical-looking blocks encoding different concerns are not a DRY
violation; two differently-worded blocks encoding one business rule are.

Sandi Metz supplies the counterweight — "duplication is far cheaper than the
wrong abstraction" — and describes the decay cycle: extract, then a near-miss
requirement arrives, then a parameter, then a conditional, repeating until the
abstraction is incomprehensible. Her remedy is to run it backwards: re-inline
into callers, keep only the branch each caller needs. Kent Dodds' AHA (Avoid
Hasty Abstractions) adds the timing rule — wait until "the commonalities scream
at you," because you cannot predict the shape of future change.

This yields an asymmetry worth encoding: **under-abstraction is cheap to fix,
over-abstraction is expensive.** A maintenance pass should therefore look for
*wrong* abstractions to inline as eagerly as it looks for duplication to extract.

### Execution: separate structural from behavioral change

Kent Beck's *Tidy First?* separates structural commits from behavioral commits so
each can be reviewed, bisected, and reasoned about independently. His Canon TDD
post names the two relevant mistakes directly: "refactoring further than
necessary for this session" and "abstracting too soon. Duplication is a hint, not
a command."

Michael Feathers supplies the safety precondition for untested areas —
characterization tests that "document your system's actual behavior, not check
for the behavior you wish your system had."

Naming the specific Fowler catalog refactoring (Extract Function, Inline
Function, Extract Class, Replace Conditional with Polymorphism, Consolidate
Conditional Expression, Remove Dead Code) makes each proposal reviewable.

### AI-specific risk: the target and the danger are the same code

Borg et al., "Code for Machines, Not Just Humans" (2026), find "a meaningful
association between CodeHealth and semantic preservation after AI refactoring."
Low-health code is both the highest-value target *and* the place an AI is most
likely to silently break behavior. Health must be read as a **risk** signal, not
only a priority signal. CodeScene's own ACE tool validates LLM refactorings for
correctness before surfacing them, which implies a non-trivial raw error rate.

Fowler's rabbit-hole warning applies with extra force to agents: "as you fix one
thing you spot another, and another, and before long you're deep in yak hair."

### Prior art: this pattern does not exist yet

- **Claude Code** ships `/code-review`, `/security-review`, `/debug`, `/verify`.
  All are diff-scoped. There is no `/simplify`, `/tidy`, `/janitor`, or
  `/tech-debt` skill.
- **obra/superpowers** has brainstorming, planning, TDD, code review, debugging,
  and branch-finishing skills — but no simplification lane. Its planning skill is
  explicitly anti-sweep: "Don't propose unrelated refactoring."
- **Aider** has `/lint` but no maintenance workflow. **Cursor** has no native
  refactor command (could not be fully verified).

No surveyed tool presents a ranked menu of maintenance targets and loops on it.
The closest interaction model is superpowers' `requesting-code-review`, which
returns severity-tiered findings for one diff.

## Sources

| Claim | Source | Type |
|---|---|---|
| Hotspot = code health x change frequency | https://codescene.com/blog/prioritize-technical-debt-by-impact/ | Primary |
| 15x defects, 124% more time, 9x cycle time | Tornhill & Borg, "Code Red", arXiv:2203.04374, IEEE Software 39(4) 2022 | Primary, peer-reviewed |
| Code health predicts AI semantic preservation | Borg et al., "Code for Machines, Not Just Humans", arXiv 2026 | Primary |
| Validated LLM refactoring (ACE) | Tornhill et al., arXiv 2025 | Primary |
| Stable cruft can be left alone | https://martinfowler.com/bliki/TechnicalDebt.html | Primary |
| Rabbit-hole warning | https://martinfowler.com/bliki/OpportunisticRefactoring.html | Primary |
| "duplication is far cheaper than the wrong abstraction" | https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction | Primary |
| AHA — Avoid Hasty Abstractions | https://kentcdodds.com/blog/aha-programming | Primary |
| DRY is about knowledge | Hunt & Thomas via https://en.wikipedia.org/wiki/Don%27t_repeat_yourself | Secondary |
| Structural vs behavioral commits | Beck, *Tidy First?* (O'Reilly 2023) | Secondary |
| "Duplication is a hint, not a command" | https://newsletter.kentbeck.com/p/canon-tdd | Primary |
| Characterization tests | https://michaelfeathers.silvrback.com/characterization-testing | Primary |
| Refactoring catalog | https://refactoring.com/catalog/ | Primary |
| Clean As You Code | https://docs.sonarsource.com/sonarqube-server/user-guide/about-new-code.md | Primary |
| Claude Code bundled skills | https://code.claude.com/docs/en/skills | Primary |
| superpowers skill inventory | https://github.com/obra/superpowers | Primary |
| Aider commands | https://aider.chat/docs/usage/commands.html | Primary |

## Applies When

- Designing or changing `kb-simplify`.
- Deciding whether a proposed extraction is justified, or whether an existing
  abstraction should be inlined instead.
- Arguing about whether a cleanup pass should be automatic or explicit.
- Choosing how to rank maintenance work in any KB lane.

## Stale When

- CodeScene publishes independently replicated hotspot evidence, or a
  third party contradicts the Code Red findings.
- Claude Code, Cursor, or superpowers ship a first-party simplification skill
  worth comparing against.
- `kb-simplify` gains automated invocation, which would contradict the explicit
  invocation finding here.

## Rejected Approaches

- **Pure static complexity ranking.** Surfaces ugly-but-stable code nobody
  touches. Rejected on Tornhill and Fowler.
- **Automatic invocation inside `kb-work` or `kb-complete`.** Every surveyed tool
  keeps maintenance sweeps human-triggered, and superpowers explicitly forbids
  unrelated refactoring during planned work.
- **Applying all findings in one diff.** The dominant reported failure mode is
  the unreviewable sweep.
- **A deterministic `kbcheck` complexity budget as the primary mechanism.**
  Considered and dropped for now: judgment about whether an abstraction is
  *wrong* is not expressible as a file/dependency count, and building scoring
  machinery to police overengineering is itself the failure it targets.
- **Extract-on-second-occurrence.** Contradicts Rule of Three and AHA.

## Impact On Current Project

Creates `.github/skills/kb-simplify/SKILL.md` as an explicitly invoked
maintenance lane. It fills the gap `kb-architecture-deepening` already names but
does not own ("AI residue cleanup ... That is cleanup/deslop territory").

Lane boundaries after this change:

- `kb-review` — one profile against one diff; dimension 4 covers avoidable
  complexity at review time.
- `kb-architecture-deepening` — should the architecture change; max 3 candidates;
  pre-implementation.
- `kb-simplify` — already-committed code; max 6 ranked targets; execute-one loop.

Supersedes the earlier idea of a per-slice `complexity_budget` contract in
`kb-plan`/`kbcheck`. See Rejected Approaches.
