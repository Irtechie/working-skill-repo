# Codex Critic Audit — Adoption, Token Rent, and Proof Claims

Date: 2026-07-13
Scope: `README.md`, all `.github/skills/**/SKILL.md`, `kb-plan`, `kb-work`, and routing/AMR claims under `docs/` and `.github/skills/`
Mode: read-only source audit; this report proposes changes but applies none

Revision note: P0-ADOPT-02 was tightened after the user reported that `npx` is
now blocked. Current evidence shows the GitHub package syntax is still supported
by npm in general, but the documented route is operationally blocked in the
tested environment and should not remain the canonical on-ramp.

## Audit standard

A finding exists only when it has a falsifiable failed criterion, an exact
location, observed evidence, a smallest fix, and a P0-P2 severity. Operative
skill sections were classified as:

- **A — load-bearing:** a deterministic checker or a current repo artifact
  demonstrates the instruction's output/behavior.
- **B — duplicated:** another named surface owns the same normative content.
- **C — aspirational:** no enforcing check and no current artifact demonstrating
  the instruction were found.

Headings inside fenced templates are template content, not separate operative
sections.

## Summary

| Area | P0 | P1 | P2 | Total |
|---|---:|---:|---:|---:|
| Adoption surface | 2 | 1 | 0 | 3 |
| Token rent | 0 | 9 | 1 | 10 |
| Proof-claim hygiene | 1 | 0 | 0 | 1 |
| **Total** | **3** | **10** | **1** | **14** |

## P0 findings

### P0-ADOPT-01 — The first-use promise does not answer “why” before process/status detail

failed_criterion: After the first descriptive sentence, a competent stranger can
explain both what the project does and why they would want it.

failure_location: `README.md:3-12`, especially line 3; the expansion of `KB` is
deferred to `README.md:278-280`.

evidence: Line 3 says only `Portable KB workflow skills for GitHub Copilot and
Codex.` This identifies the artifact type and hosts, but not the user outcome.
The first outcome list appears at lines 10-12, after status and provenance, and
`KB` is not expanded until line 278. A stranger stopping after the opening
description cannot yet say why this is preferable to ordinary agent prompts.

smallest_fix: Move the existing `KB means Kanban-Based` definition to the opening
and add one concrete outcome sentence before the status line using only existing
terms: recover project memory, choose a bounded lane, execute, and prove the
result. Do not add a new concept or architecture section.

severity: P0

### P0-ADOPT-02 — The canonical `npx github:` install route is operationally blocked

failed_criterion: The README's canonical first-use command completes in the
supported environment, or the README routes users to an installation path that
does.

failure_location: `README.md:17-31`, `README.md:649-701`, and
`package.json:26-28`.

evidence: The README gives
`npx github:Irtechie/working-skill-repo --target all --profile core`. The exact
command stalled for more than 90 seconds. A second bounded probe,
`npm view github:Irtechie/working-skill-repo name version`, also produced no
output and required termination. The npm debug log stops at
`fetch manifest working-skill-repo@github:Irtechie/working-skill-repo`.
Meanwhile, `git ls-remote https://github.com/Irtechie/working-skill-repo.git HEAD`
completed and returned the remote HEAD, so GitHub itself was reachable. The repo
has no `.npmrc`, project `allow-git` resolves to `all`, and current npm
documentation still lists `github:user/repo` as supported with `allow-git`
defaulting to `all` ([npm install documentation](https://docs.npmjs.com/cli/v11/commands/npm-install/));
therefore the evidence does not support a universal npm ban.
It does prove that the documented route is blocked/unusable in the tested
environment. The cloned local installer was then verified with
`node ./bin/kb-install.mjs --target all --profile core --dry-run --install-root
<temp> --router skip`; it exited 0, planned all three target roots, and wrote
nothing.

smallest_fix: Remove `npx github:` from `Start Here` and make the verified
clone-plus-local installer the canonical path. Keep `npx` only in an explicitly
non-canonical section after a bounded release test proves it again. State the
existing `Node.js 18+` requirement beside the local command.

severity: P0

### P0-PROOF-01 — Review guidance asserts cost/latency savings without a landed result

failed_criterion: No skill or reference asserts routing cost, token, or latency
improvement unless a landed artifact under `docs/results/` or `docs/reports/`
measures that exact claim.

failure_location: `.github/skills/kb-review/references/review-process.md:218-226`
and `.github/skills/ce-review/references/review-process.md:218-226`.

evidence: Both files say persona agents `should use cheaper/faster models to
reduce cost and latency`, direct the platform to use its `cheapest capable`
model, and name `haiku` and `gpt-4o-mini`. A search of `docs/results/` and
`docs/reports/` for review cost, review latency, those model names, and the
quoted claim found no backing result. The only matching routing release proof is
`docs/results/2026-07-11-session-model-routing-release-proof.md`, which proves a
deterministic initial-pilot release gate, not review cost or latency reduction.
This contradicts the conservative posture at `README.md:172-173`,
`README.md:223-225`, and `README.md:248-249`.

smallest_fix: Delete the duplicated model-tiering blocks at lines 218-226 in both
references. Preserve the fallback rule that review must still work when no exact
route is available, but do not name a price/performance outcome or stale model.

severity: P0

## P1 findings

### P1-ADOPT-03 — Four execution terms appear before a concrete definition

failed_criterion: Every coined execution term in the adoption path is defined
with a concrete example at or before first use.

failure_location: `README.md:183-194`; `HITL` first appears at
`README.md:366` without expansion.

evidence: `ready-set` and `scope-lease` first appear at line 183. A ready set is
partly explained later at lines 350-352, but `scope-lease` is never defined with
an example in the README. `route-bound receipt` and `dispatch-proven` appear at
lines 193-194 without showing the minimum fields or a pass/fail example. `HITL`
appears at line 366 without `human-in-the-loop`. DDR and AMR do not fail this
criterion: DDR is defined at lines 113-117 and AMR is expanded and bounded at
lines 48-50 and 152-173. The image alt text at lines 119 and 159 is descriptive
and makes no performance claim.

smallest_fix: Add one parenthetical at first use for each existing term:
dependency-unblocked slices for `ready-set`, one writer for an overlapping path
for `scope-lease`, exact route identity plus linked work proof for
`dispatch-proven`, and `human-in-the-loop (HITL)`. Do not create a glossary.

severity: P1

### P1-RENT-01 — `kb-plan` interaction policy has no observable enforcement

failed_criterion: The `Interaction Method` section is either demonstrated by a
trace/check or consolidated into the gate that owns blocking-question policy.

failure_location: `.github/skills/kb-plan/SKILL.md:26-58`.

evidence: The section prescribes when to ask, when to assume, and exact follow-up
chat text. `kbcheck manifest-contract` and gate-ledger checks validate artifact
state, not whether the planner asked only one material question or printed the
specified text. No interaction trace/result proving this section was found.
`kb-gate` already owns Question Gate classes and phase advancement at
`.github/skills/kb-gate/SKILL.md:81-120`.

smallest_fix: Consolidate the blocking-question policy into `kb-gate` ownership
and delete the duplicated examples from `kb-plan`, retaining only the phase
boundary and exact next-skill decision.

severity: P1

### P1-RENT-02 — `kb-plan` research orchestration is prose-only

failed_criterion: `Research (Parallel)` produces an artifact, gate field, or
trace that proves which local/specialist/external research ran or why it was
skipped.

failure_location: `.github/skills/kb-plan/SKILL.md:165-192`.

evidence: The section names agents and decision rules, but the manifest template
at lines 320-433 has no research-evidence field and `manifest-contract` does not
validate one. Current plans can contain research pointers, but no deterministic
surface establishes that this section was followed for each plan.

smallest_fix: Delete the orchestration prose from `kb-plan` and retain one
delegation sentence to `kb-research` only when material uncertainty remains.

severity: P1

### P1-RENT-03 — `kb-plan` embeds 143 lines of board/handoff templates owned elsewhere

failed_criterion: Standard project-memory layout and board conventions have one
authoritative template owner.

failure_location: `.github/skills/kb-plan/SKILL.md:508-650`.

evidence: This range embeds full `todo.md`, `todo-done.md`, active-work, handoff,
status-marker, and standing-section templates: 143 lines and 765 whitespace-token
estimate. `kb-map-bootstrap` already owns missing-memory setup and standard
layout at `.github/skills/kb-map-bootstrap/SKILL.md:13-18` and `:37-69`; the
project contract also says `kb-map` invokes bootstrap when `todo.md` or
`PROJECT.md` is missing. Keeping a second template in `kb-plan` creates drift.

smallest_fix: Keep lines 503-507 (update the board; create a handoff only when a
restart packet is needed), delete lines 508-650, and route missing memory to
`kb-map`/`kb-map-bootstrap`.

severity: P1

### P1-RENT-04 — `kb-plan` optional-commit prose is not an enforceable staging guard

failed_criterion: The optional commit section is enforced by a scoped staging
check or owned by the delivery skill instead of relying on an example command.

failure_location: `.github/skills/kb-plan/SKILL.md:660-667`.

evidence: The section says to stage only generated files, but its sample command
can still stage paths that were not generated in the current run and no checker
binds the list to a plan receipt. Commit/delivery policy is already owned by
`kb-ship` and ambient repo instructions. No result artifact demonstrates this
section as an independent guard.

smallest_fix: Delete the example `git add`/`git commit` block and keep the single
existing rule at Quick Start line 24: commit only when explicitly requested.

severity: P1

### P1-RENT-05 — `kb-work` thin-plan deepening has no proof surface

failed_criterion: `Deepen If Thin` records a checkable before/after condition or
is removed from the executor hot path.

failure_location: `.github/skills/kb-work/SKILL.md:298-304`.

evidence: The trigger is fewer than three acceptance criteria or no test
scenarios, but no command checks it, no manifest field records the deepening
result, and no current result artifact demonstrates compliance.

smallest_fix: Delete this section and let an invalid/thin slice fail the existing
preflight/plan contract and return to `kb-plan`.

severity: P1

### P1-RENT-06 — `kb-work` duplicates the functional-test classifier

failed_criterion: Test-level classification and browser minimum proof are
defined once by their owning skill.

failure_location: `.github/skills/kb-work/SKILL.md:306-342` and
`.github/skills/kb-work/SKILL.md:423-435`.

evidence: These ranges repeat `test_level`, `functional_risk`, UI extension
rules, browser interaction, screenshot cleanup, and escalation criteria already
owned by `.github/skills/kb-functional-test/SKILL.md:19-50` and `:70-186`.
The duplicated `kb-work` ranges total 50 lines and roughly 500 whitespace-token
estimate.

smallest_fix: Keep the requirement to invoke `kb-functional-test` and record its
classification; delete the repeated classifier table and browser procedure from
`kb-work`.

severity: P1

### P1-RENT-07 — `System-Wide Test Check` is an unenforced checklist

failed_criterion: The system-wide test section emits a traceable check/result or
is delegated to a verification owner.

failure_location: `.github/skills/kb-work/SKILL.md:520-532`.

evidence: The section asks four reasoning questions and even says one branch
`take[s] 10 seconds`, but writes no manifest field, invokes no command, and has
no landed result artifact. The same risk class is already handled by
`kb-functional-test`, `kb-check`, and `kb-qa`.

smallest_fix: Delete lines 520-532 and rely on the declared `test_level`,
functional proof, and QA gates.

severity: P1

### P1-RENT-08 — The destructive-command section calls prose “enforcement”

failed_criterion: A section that says `This is enforcement` has an executable
interceptor/check or does not claim enforcement.

failure_location: `.github/skills/kb-work/SKILL.md:573-595`, especially line
594.

evidence: The blocklist is Markdown read by an agent; `kb-work` invokes no
command interceptor or runtime hook before shell calls. The text therefore
cannot prove that a blocked pattern was detected. Platform and repo ambient
safety instructions already govern destructive commands.

smallest_fix: Delete the pseudo-enforcement block from `kb-work` and retain one
sentence deferring destructive operations to platform/repo approval policy.
Do not weaken the actual approval requirement.

severity: P1

### P1-RENT-09 — Figma sync is conditional prose with no availability or result contract

failed_criterion: The Figma section records tool/design availability and a
comparison result or stays outside the executor hot path.

failure_location: `.github/skills/kb-work/SKILL.md:609-619`.

evidence: The section says to use `figma-design-sync` repeatedly until the
implementation matches, but does not define how designs are discovered, how
match is measured, what artifact is recorded, or what happens when the agent is
absent. No current result artifact was found for this contract.

smallest_fix: Delete this section from `kb-work`; UI acceptance criteria and
`kb-qa` remain authoritative. Invoke a design-specific tool only when the slice
itself names one.

severity: P1

## P2 findings

### P2-RENT-10 — `surface-report` line totals differ from conventional file line counts

failed_criterion: The surface-report `lines` metric matches the repo's ordinary
newline/line-count measurement for newline-terminated `SKILL.md` files.

failure_location: `cmd/kbcheck/skill_validators.go:827-831` and
`cmd/kbcheck/report_validators.go:451-458`.

evidence: Conventional `Get-Content` counts 8,541 lines across 43 `SKILL.md`
files. The source-equivalent `surface-report` algorithm splits the full text on
newlines and counts the terminal empty element, producing 8,584: exactly +43,
one per newline-terminated file. The token estimate is unaffected at 63,388
because it trims whitespace first. Direct `go run ./cmd/kbcheck surface-report
--json` was attempted twice during this audit but stalled behind concurrent Go
processes; the cross-check above reproduces the checked-in implementation
exactly.

smallest_fix: Change `countLines` to ignore one terminal empty split element, or
rename/document the metric as split rows. Preserve baseline compatibility
explicitly when making that change.

severity: P2

## Adoption walk result

The first blocking point is line 3: it says what kind of repository this is but
does not yet say why a user wants it. Lines 8-12 eventually provide outcomes.
The first copy-paste action begins at line 19 and is not fully diagnosable from
the docs because Node 18+, success output, and a bounded verification step are
missing.

Terms checked at first use:

| Term | First use | Concrete definition at/before use | Result |
|---|---:|---|---|
| KB | 3 | No; expanded at 278 | Finding P0-ADOPT-01 |
| AMR | 48 | Yes; expanded and bounded at 48-50 | Pass |
| DDR | 113 | Yes; defined at 115-117 | Pass |
| ready-set | 183 | No; partial explanation at 350-352 | Finding P1-ADOPT-03 |
| scope-lease | 183 | No concrete README example | Finding P1-ADOPT-03 |
| route-bound receipt | 193 | No minimum-field example | Finding P1-ADOPT-03 |
| dispatch-proven | 194 | No pass/fail example | Finding P1-ADOPT-03 |
| HITL | 366 | No expansion | Finding P1-ADOPT-03 |
| proof spine | 383 | Yes; commands and RED/GREEN rule at 384-388 | Pass |

## Token-rent measurements

Measurement commands:

```powershell
Get-ChildItem .github/skills -Recurse -Filter SKILL.md -File
# conventional physical count: (Get-Content <file>).Count
# surface-report semantics: split full text on \r?\n; token estimate = whitespace words after TrimSpace
```

- Skill files: **43**
- Conventional physical lines: **8,541**
- `surface-report`-semantic rows: **8,584**
- `surface-report` token estimate: **63,388**
- `kb-work`: **737** report rows / **7,066** token estimate
- `kb-plan`: **684** report rows / **4,330** token estimate

Per-skill counts, largest first:

| Rank | Skill | Report rows | Token estimate |
|---:|---|---:|---:|
| 1 | kb-work | 737 | 7,066 |
| 2 | kb-plan | 684 | 4,330 |
| 3 | kb-brainstorm | 567 | 4,784 |
| 4 | ce-compound | 468 | 3,536 |
| 5 | kb-finalize | 447 | 3,570 |
| 6 | document-review | 318 | 2,240 |
| 7 | kb-goal | 314 | 1,864 |
| 8 | ce-compound-refresh | 291 | 3,402 |
| 9 | kb-map | 279 | 1,855 |
| 10 | learn | 251 | 1,502 |
| 11 | kb-review | 248 | 2,394 |
| 12 | kb-epic | 242 | 1,490 |
| 13 | kb-qa | 239 | 1,875 |
| 14 | ce-review | 236 | 2,305 |
| 15 | kb-eval-map | 231 | 1,208 |
| 16 | kb-map-bootstrap | 221 | 1,317 |
| 17 | kb-start | 211 | 1,834 |
| 18 | kb-functional-test | 187 | 1,725 |
| 19 | kb-models | 164 | 1,391 |
| 20 | kb-repair | 156 | 1,014 |
| 21 | evolve | 150 | 753 |
| 22 | repo-critic | 146 | 770 |
| 23 | kb-complete | 135 | 689 |
| 24 | kb-troubleshoot | 132 | 1,357 |
| 25 | kb-gate | 125 | 965 |
| 26 | kb-ship | 123 | 780 |
| 27 | kb-memory-review | 118 | 826 |
| 28 | kb-first-principles | 115 | 677 |
| 29 | kb-handoff | 108 | 737 |
| 30 | kb-check | 105 | 672 |
| 31 | kb-task | 102 | 781 |
| 32 | kb-regression-snapshot | 100 | 450 |
| 33 | kb-land | 95 | 516 |
| 34 | kb-configure | 91 | 473 |
| 35 | kb-research | 76 | 267 |
| 36 | kb-fix | 75 | 475 |
| 37 | kb-architecture-deepening | 62 | 335 |
| 38 | kb-compact | 54 | 308 |
| 39 | todo-create | 50 | 264 |
| 40 | todo-triage | 50 | 212 |
| 41 | tdd | 40 | 220 |
| 42 | klfg | 22 | 80 |
| 43 | kb-finish | 19 | 79 |

The ranking uses the checked-in surface-report row semantics. The stated top ten
are also the exact ten largest from the conventional physical-line command.
Before using this table as a baseline, fix or explicitly accept P2-RENT-10.

## `kb-plan` section classification

| Operative section | Lines | Class | Demonstration or finding |
|---|---:|:---:|---|
| Quick Start | 13-24 | A | Current manifests contain slice plans and passing `plan-to-work` gates |
| Interaction Method | 26-58 | C | P1-RENT-01 |
| Input | 60-82 | A | Current manifests record brainstorm/source/handoff origins |
| Vertical Slices Only | 86-102 | A | Current manifests contain dependency-linked end-to-end slices |
| Enabling Slices Are Acceptable | 103-110 | A | Current manifests name blockers and downstream consumers |
| Live-Steering Slices | 111-129 | A | `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md` records the contract |
| Every Slice Has a Verification Strategy | 130-145 | A | `manifest-contract` validates verification/test fields |
| Understand the Source Material | 149-164 | A | Gate ledger records source and unresolved-question state |
| Research (Parallel) | 165-192 | C | P1-RENT-02 |
| Draft Vertical Slices | 194-236 | A | Manifest and slice schemas demonstrate all required fields |
| Validate the Breakdown | 238-263 | A | `plan-to-work` gate and manifest checker enforce the contract |
| Model Tier Contract | 265-290 | A | `manifest-contract` and current manifests validate tiers |
| Objective Done Contract | 292-314 | A | `done_check`/`proof_check` are deterministically validated |
| Generate Plan Files | 316-501 | A | Current plan artifacts instantiate the templates |
| Update Todo/Handoffs action | 503-507 | A | Current board and active/done handoffs demonstrate the action |
| Embedded Todo/Handoff templates | 508-650 | B | P1-RENT-03 |
| Validate Output | 651-659 | A | Manifest/gate checkers cover these fields |
| Optional Commit | 660-667 | C | P1-RENT-04 |
| Success Criteria | 669-675 | A | Manifest/gate validation covers the criteria |
| Integration with Other Skills | 677-683 | A | Current archived manifests show brainstorm -> plan -> work routing |

## `kb-work` section classification

| Operative section | Lines | Class | Demonstration or finding |
|---|---:|:---:|---|
| Quick Start | 13-21 | A | Current work manifests contain terminal gates and finalization proof |
| Input | 23-33 | A | Current manifests/handoffs use the named routing surfaces |
| Continuous Completion Loop | 35-72 | A | Completed manifests and finalization artifacts demonstrate terminal handling |
| Pre-Flight | 74-101 | A | Gate-ledger, manifest-contract, and current board checks exist |
| Board Sync Protocol | 103-126 | A | Current `todo.md` and manifests carry matching status/notes |
| Run-State Events | 127-148 | A | `.kb/runs/*/route-history.jsonl` and `kbcheck run-state` exist |
| Live Route Selection | 150-230 | A | Router commands and the session-routing release proof exist |
| Ready-Set Ordering | 231-247 | A | `kbcheck ready-set-selftest` is in the core contract |
| Continuous Execution Rule | 252-273 | A | Current manifests show continued ready-slice progress and terminal gates |
| HITL Flag | 275-296 | A | Current board/manifest contain human-required and blocked states |
| Deepen If Thin | 298-304 | C | P1-RENT-05 |
| Test-Level Classification | 306-342, 423-435 | B | P1-RENT-06 |
| Adaptive Model Routing | 343-421 | A | No-paid conformance/finalization artifacts demonstrate disabled/fail-closed behavior |
| Regression Snapshot Gate | 437-446 | A | Current manifests record snapshot paths/results |
| Scope Forecast and Ledger | 448-477 | A | Current manifest notes record forecast/discovery counts |
| Execute | 479-498 | A | Current slice result artifacts and receipts demonstrate execution |
| Protected Oracle Gate | 500-518 | A | Current manifests record protected SHA256 values |
| System-Wide Test Check | 520-532 | C | P1-RENT-07 |
| Diff-Scope Verification | 533-571 | A | Current manifests record `scope-check` and `scope-discovery` notes |
| Destructive Command Guard | 573-595 | C | P1-RENT-08 |
| KB QA | 596-608 | A | Current manifests record `qa-lint` and `qa-browser` outcomes |
| Figma Design Sync | 609-619 | C | P1-RENT-09 |
| Verify and Update | 620-668 | A | Slice-to-done gates and proof notes exist in current manifests |
| Completion | 670-696 | A | Work-to-complete and finalization proof artifacts exist |
| Failure Handling | 698-710 | A | Current blocked/human-required handoffs demonstrate failure states |
| Resume Support | 712-718 | A | Statused manifests and active handoffs are resumable artifacts |
| Success Criteria | 720-725 | A | Gate/diff/manifest checks cover the criteria |
| Integration | 727-736 | A | Current manifests invoke the named verification/finalization owners |

## Proof-claim hygiene result

The README complies with the falsification posture in the architecture under
audit:

- `README.md:172-173` says AMR makes no savings claim until paired evidence wins
  without reducing correctness.
- `README.md:223-225` limits the present claim to existing routes/gates/checks
  and explicitly disclaims measured token savings.
- `README.md:248-249` says the no-paid artifact has zero supported cohorts and
  makes no live cost, latency, token, or savings claim.
- The DDR and AMR image alt text at lines 119 and 159 describes the diagrams and
  does not assert a result.
- `docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md:73-75`
  explicitly labels its evidence a deterministic fixture, not a live-model
  benchmark.

The sole contradictory claim found is P0-PROOF-01 in the duplicated review
references. Historical plans that describe desired measurements or forbid
future unsupported claims are proposals, not shipped result assertions, and
were not misclassified as benchmark results.

## Proposed diff manifest — do not apply in this audit

| Path | Proposed smallest change | Approximate deletion |
|---|---|---:|
| `README.md` | Move the existing KB expansion to the opening; define four existing terms at first use; replace canonical `npx github:` startup with the verified clone/local installer and state Node 18+ | small net change |
| `.github/skills/kb-review/references/review-process.md` | Delete lines 218-226 unsupported model/cost block | 9 lines / 158 estimate |
| `.github/skills/ce-review/references/review-process.md` | Delete lines 218-226 duplicate unsupported block | 9 lines / 158 estimate |
| `.github/skills/kb-plan/SKILL.md` | Delete embedded memory templates at 508-650; route missing layout to `kb-map` | 143 lines / 765 estimate |
| `.github/skills/kb-plan/SKILL.md` | Consolidate interaction policy into `kb-gate`; delete prose-only research orchestration and optional commit example | about 60 lines |
| `.github/skills/kb-work/SKILL.md` | Delete duplicated test/browser classifier; retain invocation and recorded result | about 50 lines / 500 estimate |
| `.github/skills/kb-work/SKILL.md` | Delete prose-only deepening, system-wide checklist, pseudo-enforcement block, and Figma loop | 54 lines / about 500 estimate |
| `cmd/kbcheck/skill_validators.go` | Make terminal-newline treatment explicit and baseline-safe | 1 small function change |

## What must not change

1. Keep the explicit no-savings statements and zero-supported-cohort boundary;
   deleting them would let model-price intuition masquerade as measured benefit.
2. Keep deterministic work proof authoritative over routing receipts; a receipt
   proves attribution, not correctness.
3. Keep DDR as planning difficulty plus work-time selection and AMR as the
   narrower optional attempt; do not re-split or re-litigate the landed
   consolidation.
4. Keep manifest gates, protected oracles, `done_check`/`proof_check`, and
   fail-closed completion. These are the measurable spine, not ceremony filler.
5. Keep endpoints, credentials, approvals, and source preferences user-local;
   adoption simplification must not move private routing state into plans or
   shared skills.

## Highest deletion return: three smallest fixes

The estimate below uses the same whitespace-token approximation as
`surface-report`; minutes are bounded editing estimates, not measured labor.

1. Delete the duplicated unsupported review model-tier blocks in both reference
   files: **316 estimate / 2 minutes = 158 per minute**.
2. Delete `kb-work`'s unenforced System-Wide Test checklist and rely on the named
   verification owners: **213 estimate / 2 minutes = 106.5 per minute**.
3. Delete `kb-plan`'s embedded board/handoff templates and route missing memory
   to `kb-map`: **765 estimate / 8 minutes = 95.6 per minute**.
