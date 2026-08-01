# Model Tier Qualification

`kbcheck model-tier-eval` is an experimental, offline classifier for one strict,
redacted evidence document:

```powershell
go run ./cmd/kbcheck model-tier-eval --evidence <repo-relative-evidence.json> --json
```

The first policy supports only `medium` with threshold revision `medium-v1`.
Qualification requires 30 admitted unique fixtures, five materially independent
families, no family above 20%, at least one admitted holdout family, and zero
admitted model failures. A model-output schema failure or unchanged-proof
failure after valid execution is a model failure. Plan, oracle, preflight,
route, no-response, indeterminate, and proof-infrastructure outcomes are visible
exclusions and never improve the denominator.

Evidence must bind a repo-contained preregistered cohort manifest, one immutable
execution fingerprint, unique attempt and fixture identities, family dimensions,
and repo-contained artifact paths with recomputed SHA-256 hashes. The cohort
manifest must carry a valid Ed25519 signature from a signer listed in
`config/model-tier-trust.json`. Receipt facts are `reproduced` or `unsupported`;
unsupported authority cannot qualify, so self-asserted trust labels and forged
signatures cannot qualify.

The scorer rejects unknown or duplicate fields, path escape, symlinks,
oversized/deep input, sensitive fields or values, mixed fingerprints, omitted
or replayed attempts, and unsupported tiers. It performs no network access,
route discovery, inference, credential lookup, or private-key operation.
Evidence must not contain endpoints, credentials, prompts, response bodies,
protected tests, oracle contents, solutions, or host-local paths.

Results are `qualified`, `not-qualified`, or `inconclusive` and remain scoped to
the recorded task families, tools, context/risk envelope, route revision, and
evidence date. They are evidence artifacts, not automatic routing promotion.
Provider revisions that cannot be observed expire after 30 days.

`fixtures.json` is consumed by `TestModelTierEvalDeterministicCorpus` and covers
qualification, model failures, all exclusions, threshold shortfalls,
correlation, stale and unsupported evidence, replay, selective omission,
unknown fields, mixed fingerprints, sensitive values, and unsupported tiers.

Focused proof:

```powershell
go test ./cmd/kbcheck -run 'ModelTierEval' -count=1
```
