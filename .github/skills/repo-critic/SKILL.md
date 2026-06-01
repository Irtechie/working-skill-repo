---
name: repo-critic
description: "Claims-vs-code evidence review for repos, skills, eval harnesses, README claims, commit messages, and tests. Use when the user asks to audit claims, verify whether implementation matches documentation, find overclaims, review evaluator honesty, compare README/skills against code, or run a proof-focused critic pass before shipping."
argument-hint: "[ref/path/scope to review]"
---

# Repo Critic

Find the gap between what the repo claims and what the checked artifact
observably does.

Be evidence-first about claims and precise about proof. The goal is not tone;
it is forcing claims to match code, tests, and executed behavior.

## Prime Directive

Verify the observed artifact, not the convenient one.

1. State the exact ref/SHA reviewed at the top of the report.
2. Confirm the tree before making claims:
   - `git rev-parse HEAD`
   - `git status --short --branch`
   - for remote refs, prefer `git ls-tree -r <ref>` and `git show <ref>:<path>`
     over trusting the current checkout.
3. Distinguish:
   - **read**: file or line inspected;
   - **run**: command executed and exit observed;
   - **unverified**: not read or not run.
4. Never treat "test exists" as "test asserts the behavior." Read the
   assertion or mark it unverified.

## Review Targets

Use this skill on:

- README, docs, skill descriptions, and commit messages;
- eval harnesses, scorers, fixtures, baselines, trace wrappers, and proof gates;
- workflow skills that claim autonomous execution, review, learning, or
  completion behavior;
- migration claims such as "ported to Go", "cross-platform", "live eval",
  "observed trace", "quality scorer", or "runtime enforced";
- "required" surface claims such as skill counts, reviewer-agent sets, install
  targets, and sync promises.

## Probes

Apply only probes relevant to the target. Do not pad the report.

### A. Costume Validation

- **Computed or self-authored?** Does the score come from real output or
  deterministic measurement, or from numbers typed into fixtures?
- **Advertised modes?** For every enum/flag/schema mode, name the code path that
  honors it.
- **Observed vs reported trace?** Enforce forbidden command/tool rules against
  externally observed trace when available. Required commands/files may be
  model-reported intent; forbidden safety invariants must not rely only on a
  model self-report.

### B. Discovery vs Execution

- Does the new path execute checks, or only discover and name them?
- Does parity prove behavior, or only command names and exit codes?
- Is "cross-platform" a real non-PowerShell path, or a Go launcher for a
  PowerShell engine?

### C. Executed vs Narrated Logic

- Where does numeric logic execute: code or prose?
- Are confidence, decay, scoring, WIP, and thresholds computed by code or
  narrated in a skill body?
- Search for swallowed failures in gates: `catch {}`, bare `try`, `|| true`,
  silent redirects, and ignored exit codes.

### D. Token Economics

- What is resident at peak context: one active card, or the whole manifest and
  prior slice history?
- Is there a hard WIP cap?
- Does bookkeeping run in code or in prose the model must execute in context?
- Is the hot path, especially routers like `kb-start`, lean enough for frequent
  loading?

### E. Claims-vs-Code Drift

- Pull concrete claims from README, docs, skills, or commits.
- For each claim, cite the code/test/doc line that makes it true or mark it
  `OVERCLAIM`.
- For vendored/forked skills, decide whether divergence is documented override
  or silent drift.
- For "required" surfaces, enumerate the load-bearing set or flag the claim as
  unjustified.

## Severity Defaults

- `P0`: safety/proof gate can silently lie; forbidden behavior enforced only by
  self-report; marketed feature is only narrated prose.
- `P1`: core proof/quality/release claim is materially overstated or underbuilt.
- `P2`: portability, parity, WIP, or token-economics claim is too broad.
- `P3`: documented surface is unclear or likely to drift.
- `P4`: wording nit that does not change behavior or trust.

P0/P1 block completion.

## Fix Classification

For every finding, choose the cheapest honest fix:

- `OVERCLAIM`: fix the words. Rename, narrow, or remove the claim.
- `UNDERBUILD`: fix the code. Wire the judge, capture the trace, execute the
  math, add the assertion, or enforce the gate.

Do not prescribe a rewrite when relabeling is honest. Do not relabel when the
repo genuinely needs the feature it claims.

## Output Contract

Start with:

```text
Reviewed ref: <sha/ref>
Working tree: <clean/dirty summary>
```

Then list findings first, one issue at a time:

```text
P1 - <short title>
Claim: <exact claim>
Observed: <what code/test/doc actually does>
Evidence: <file:line and/or command with exit>
Required action: <OVERCLAIM|UNDERBUILD> - <smallest honest fix>
```

End with:

```text
Verified:
- <commands run or files read that materially support non-findings>

Not verified:
- <anything not read or not run>
```

No finding without evidence. If unsure, say `unverified`.
