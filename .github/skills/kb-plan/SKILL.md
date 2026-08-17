---
name: kb-plan
description: "Turn approved requirements into independently verifiable vertical slices, a dependency DAG, and a gated KB manifest."
argument-hint: "[requirements path, feature description, or handoff]"
---

# KB Plan - Vertical Slice Decomposition

Plan thin end-to-end behavior, not horizontal implementation phases.

## Input

<input> #$ARGUMENTS </input>

Prefer a requirements source. For a handoff, resolve its requirements or
manifest pointer first. Resume an existing matching manifest instead of
creating a duplicate. A direct description may proceed only when it is
specific, low-risk, and contains no unresolved product or architecture choice;
otherwise invoke `kb-brainstorm`.

## Requirements Assurance

Planning cannot launder brainstorm ambiguity.

Before slicing, perform the main-agent requirements check:

- goals, non-goals, and acceptance criteria are explicit;
- no contradictions or unresolved `ask-now` or `research-first` items remain;
- dependencies are evidenced or labeled assumptions;
- failure, recovery, trust, migration, and integration behavior is sufficient;
- verification can detect the stated failure modes.

Fix clear document defects.
Write or update the `brainstorm-to-plan` gate as `blocked` or `needs-human`
when the source still requires user input.

Invoke `document-review mode:headless <requirements-path>` only when one
material uncertainty remains. Do not invoke `document-review` for a source
whose self-check is complete. The skill selects exactly one best-fit reviewer
for the full source; never run a reviewer per slice. Resolve P0/P1 before
decomposition. Record a matching review receipt or a specific
`not_required_reason`.

## Slice Design

Each slice must:

1. Deliver one narrow observable outcome across every relevant layer.
2. Be independently executable after its blockers.
3. Name acceptance criteria and test scenarios.
4. Forecast `expected_files` without pretending it is a write allowlist.
5. Declare `test_level`, `functional_risk`, and `execution_class`.
6. Declare `model_tier` as minimum capability, never a provider/model name.
7. Carry `model_requirements`, `escalation_triggers`, and `token_budget` at the
   precision the declared tier requires.
8. Carry `proof_check` or a narrow `no_check_reason`.
9. Mark HITL only for authority, private input, irreversible risk, or
   subjective judgment.

Choose the lowest tier that satisfies reasoning, context, tools, trust, and
risk:

| Tier | Planning classifier |
|---|---|
| `small` | Narrow mechanical change, explicit acceptance, local proof, no cross-boundary or security decision |
| `medium` | Ordinary vertical slice, focused integration, or bounded UI/API workflow |
| `large` | Architecture, auth/security/data migration, multi-subsystem work, unresolved product intent, or broad debugging |

The tier classifies minimum execution capability. `kb-work` resolves the actual
callable route from live evidence; planning never hard-codes a model.

### Execution Contract Precision

A slice must be executable **unaided** by the tier it names: that tier reaches
the acceptance criteria without asking a question, inventing an unstated
decision, or being rescued by a stronger model. A plan that only a stronger
model could execute is misclassified, not merely terse.

Precision scales inversely with tier. The lower the tier, the more the plan
must say:

| Tier | The slice must hand the executor |
|---|---|
| `large` | Objective, constraints, acceptance criteria, proof command. Approach is the executor's to choose. |
| `medium` | The above, plus ordered steps and the pass criterion for each step, plus the files and boundaries in play. Implementation detail stays open. |
| `small` | The above, plus the exact edit sites, the precise expected observable output, and one unambiguous proof command. No judgment call may remain. |

Assigning a lower tier is not a saving; it is a promise to do more
specification work. If a slice cannot be specified to that level, the tier is
wrong - raise it rather than thinning the contract.

Every slice therefore records, in the slice itself:

- `model_requirements` - capabilities the tier must have (tool use, context
  size, structured output, long-horizon reasoning), never a model name;
- `escalation_triggers` - the observable conditions under which the executor
  must stop and escalate instead of guessing;
- `token_budget` - the output budget the executor is expected to work within.

State the budget to the executor. Scoring work against a budget, criterion, or
step it was never shown is not a measurement; it measures the plan.

Enabling slices are allowed only when they are the smallest prerequisite for a
named downstream slice. Prefer behavior-first slices over schema/service/UI
phases.

### Qualification Evidence Plans (Opt-In)

Ordinary KB plans do not need semantic-plan grading. Opt in only when a plan
will be admitted as evidence for a model-tier qualification decision:

```yaml
qualification_plan_contract: true
qualification_plan:
  record_path: docs/plans/<name>-qualification-plan.json
  record_sha256: <sha256 of the strict JSON record>
```

The sidecar binds the exact plan and requirements-wide review by repo-relative
path and SHA-256. The bound Markdown plan must declare the exact reviewed
invariant IDs before any prose that follows:

```yaml
qualification_invariants:
  - stable-invariant-id
```

The strict JSON record must contain that exact invariant set. Each nontrivial
invariant must choose exactly one checkable path:

- repository-specific guidance with a contained source path and hash, stable
  anchor, mechanism or hazard, concrete executor action, and proof target; or
- an uncertainty-driven raise from the target tier to a higher supported tier
  with a specific reason.

Acceptance-criterion restatements, generic warnings, worker selection, and
stronger model names are not mechanism guidance. `document-review` owns
plan-sufficiency judgment; `kbcheck manifest-contract` validates paths, hashes,
structure, and review bindings. Do not create a DDR-specific planner.

## Verification

| Need | Verification |
|---|---|
| Behavior or logic | TDD with protected oracle when practical |
| Cross-boundary wiring | Integration |
| User/API/CLI/browser journey | Functional |
| Config/scaffolding | Verification-only |
| Subjective design approval | HITL |

Use `kb-functional-test` when a unit test could pass while the real workflow is
broken. UI-reachable behavior requires rendered UI proof. Deterministic proof,
not reviewer confidence, marks a slice done.

## Dependency DAG

- Every blocker ID must exist.
- The graph must be acyclic.
- Independent slices may share a ready set only when write/resource claims are
  disjoint and execution isolation exists.
- Serialize migration, destructive, shared UI/browser, and overlapping path
  work.
- Keep the manifest/workstream as the worktree unit; do not create a worktree
  per slice.

## Output

Write one manifest and one file per slice:

```text
docs/plans/YYYY-MM-DD-NNN-kb-<topic>-manifest.md
docs/plans/YYYY-MM-DD-NNN-<type>-<slice>-plan.md
```

New manifests use:

```yaml
manifest_schema: 3
pre_slice_review_contract: true
objective_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
pre_slice_review:
  status: passed|not-required
  source: <requirements path or direct-chat>
  source_sha256: <hash>
  mode: requirements-wide
  review_id: <id>
  reviewed_at: <RFC3339>
  review_artifact: <path>
  review_artifact_sha256: <hash>
  persona_evidence_json: '<exactly one selected reviewer and reason>'
  selected_personas_json: '<one reviewer>'
  completed_personas_json: '<same one reviewer>'
  failed_personas_json: '[]'
  findings_resolved: 0
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 0
  not_required_reason: <required when not-required>
# Include qualification_plan_contract and qualification_plan only when this
# plan itself will be admitted as model-tier qualification evidence.
```

The manifest also records `done_check`, ordered slices, gate ledger,
plan-run-worktree policy, and exact delivery authority. Each slice plan records
the fields from Slice Design plus owner, status, blocker lifecycle fields,
scope boundary, and protected oracles.

Load `references/manifest-template.md` only while writing the manifest and
`references/slice-template.md` only while writing a slice.

## Validate and Gate

1. Validate source traceability and every requirement-to-slice mapping.
2. Validate the DAG and context packets.
3. Run `manifest-contract` when available.
4. Write `plan-to-work: passed` only with objective evidence and
   `allowed_next_action: kb-work <manifest>`.
5. Once `plan-to-work` passes, invoke `w2d <manifest>` directly. Passing the
   gate is the authorization; do not ask the user to confirm execution. Stop and
   return the exact command only when the gate fails or a hard question blocks
   the plan.
   - `w2d` carries the manifest through work, finalization, and PR delivery.
   - When the caller was `kb-complete`, return control to it instead so stored
     delivery policy still owns the endpoint.

## Stop Rules

- Do not slice unresolved product intent.
- Do not put model names, aliases, adapters, endpoints, or transports in plans.
- Do not write patches, function bodies, or fixture answers into a plan. Name
  the target and the proof; the executor writes the code.
- Do not score an executor against a criterion, budget, or step the plan did
  not state.
- Do not require reviewer fan-out.
- Do not mark `plan-to-work` passed from prose confidence.
- Do not commit unless the user authorized local commits.

## Lazy References

- `references/manifest-template.md` - manifest schema and gate example.
- `references/slice-template.md` - slice schema and acceptance format.
- `references/context-packet.md` - bounded worker context packet.
- `references/gate-ledger.md` - gate fields and checker use.
