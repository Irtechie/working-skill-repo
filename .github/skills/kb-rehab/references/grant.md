# Granted delivery

This file governs the machine-enforced grant for delivering **`unshipped` feature
pairings** through the reconcile policy engine. It does not govern the lane's own
reconciliation commit, which `kb-rehab` lands directly under the standing
authorization described in that skill's `Authorization` section. Do not read a
refusal here as a reason to leave the rehab bookkeeping unlanded.

A grant is a runtime input, never a stored credential and never a file this lane
writes. It authorizes a bounded, enumerated set of merges for one run and then
expires. Absence of a grant is the normal state.

**Wiring status.** `evaluateRehabDelivery` in `cmd/kbcheck/rehab_delivery.go`
has no CLI entry point: it is reachable only from its tests. Everything in this
file describes a designed subsystem that no shipped command currently invokes.
Treat it as the contract to build against, not as a mechanism running today, and
do not cite one of its refusals as the reason a real run stopped short.

## What a grant cannot do

A grant raises no ceiling. It selects from what the shipped policy already
allows.

| A grant never | Because |
|---|---|
| Enables a disallowed action | `internal/reconcile/policy.go` sets `ActionMerge.Allowed = false` and a per-run budget of `0`. A grant capped above that ceiling is clamped to the ceiling, so the effective cap stays `0`. |
| Authorizes a protected-path merge | A pairing touching `.github/skills/**`, `cmd/**`, `internal/**`, `config/**`, or `scripts/**` stops at an open PR under every grant. |
| Authorizes global skill sync | Merging `.github/skills/**` propagates into `~/.codex/skills/`, `~/.copilot/skills/`, and `~/.agents/skills/` — a write into every future agent session on this host. Every receipt records `Sync: not-authorized`. |
| Deliver a non-`unshipped` pairing | `dead`, `superseded`, `live`, `orphan-work`, `orphan-branch`, and `human-required` are report-only regardless of grant. |
| Supply its own proof | The check gate resolves only from the authoritative default tree. |

Under the shipped policy `merge-eligible` is unreachable. That is the intended
outcome, not a defect: this repository has no forge check adapter, and most
branches touch a protected path. The honest result is a PR a human reviews.

## Record fields

| Field | Rule |
|---|---|
| `schema_version` | Must be `1`. An unknown field anywhere rejects the whole grant. |
| `run_id` | Must equal the current run. A grant from an earlier run is a replay and is void. |
| `operator` | Required. A grant naming no operator is void. |
| `issued_at` / `expires_at` | RFC3339. `expires_at - issued_at` must not exceed the policy `PlanTTL` (`10m`). |
| `evidence_cutoff` | Must equal the cutoff of the report it claims to authorize. A grant cannot be moved onto fresher evidence. |
| `caps.merge` | Clamped to the shipped per-run ceiling. Requesting more never raises it. |
| `owner_identities` | A ref whose tip identity is not named here is reported only and reaches no predicate evaluation. |
| `pairings[]` | Each entry pins `ref`, `tip_sha`, and `pull_request`. An unenumerated ref, or one whose tip moved since issue, is refused. |

## Predicates

The mandatory predicate set is consumed from the shipped `ActionMerge` manifest,
never restated here. Adding a predicate to `policy.go` therefore blocks this
lane until that predicate has an evaluation branch — an unknown name evaluates
as unsatisfied.

An adapter that is absent leaves its predicate unsatisfied. Absence is never a
pass. With no forge adapter, `exact-pr-head-base`, `fresh-mergeability`, and
`required-reviews-green` cannot hold.

`required-checks-green` may be satisfied by the repository-native gate declared
as `native_check_gate` in `config/rehab-policy.json`, but only when that value
was read from the authoritative default tree. The gate runs at most once per
run, and its failure blocks every merge in that run.

## Refusal vocabulary

Every refusal names the specific missing evidence. Report the receipt's reason
verbatim; do not soften it into "not ready".

- `state <name> is never delivered under any grant`
- `tip identity <id> is not named by the grant; reported only`
- `delivered to PR only: unsatisfied mandatory predicates`
- `delivered to PR only: touches protected paths <paths>; never auto-merge-eligible under any grant`
- `delivered to PR only: the grant does not enumerate this ref`
- `delivered to PR only: the enumerated tip SHA no longer matches the observed tip`
- `delivered to PR only: merge cap <n> reached`

## Contention

Delivery evaluation takes the same repository lock namespace `kbreconcile` uses
(`<git-common-dir>/.copilot-kb`, `work-queue.lock`). A contended run records
`contended: true`, decides nothing, and mutates nothing. Never bypass the lock.
