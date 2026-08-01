# Review Process

## 1. Bind Scope and Evidence

Resolve the base and integrated tree, authoritative intent, requirements hash,
proof receipt hash, and exact changed paths. Use caller-provided scope when it
is bound to the integrated tree. Otherwise use `resolve-base.sh` and include
tracked working-tree changes. Identify untracked files as excluded scope.

Stop when integrated deterministic proof is missing or stale.

## 2. Classify Review Need

Skip only when every path is docs-only, proven generated output, or
mechanically constrained and fully covered by deterministic proof. Unknown,
runtime, behavior, contract, configuration, persistence, trust, API, CLI, and
UI changes require semantic review.

## 3. Select One Profile

Use `persona-catalog.md`. Select the profile for the highest-consequence
dominant risk. Do not add profiles. A specialist prompt retains the universal
four-question contract.

## 4. Dispatch or Fallback

Send one read-only reviewer the intent, tree and receipt fingerprints, risk
classification, selected profile reason, file list, diff, universal questions,
and findings schema. If no reviewer agent is available, run one local
structured pass and label it `review-mode: local-fallback`.

## 5. Validate Findings

Require concrete evidence, correct line references, calibrated severity, and a
specific fix or owner. Suppress lint output, protected-artifact cleanup, and
speculation. With one reviewer there is no merge swarm or cross-persona vote.

## 6. Fix and Confirm

Apply only `safe_auto` fixes in modes that permit mutation. Any code-affecting
fix invalidates affected proof and the review receipt. Rerun proof, then run one
bounded confirmation with the same profile. Never add a second profile.

## 7. Write Receipt

Bind the final receipt to base tree, integrated tree, requirements hash, proof
receipt hash, policy version, risk, profile, mode, scope, counts, and residual
risks. A later tree or requirements change invalidates it.
