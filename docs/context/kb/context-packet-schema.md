# Context Packet Schema

Context packets are vendor-neutral execution inputs for bounded workers. They
do not replace manifests, goal ledgers, route history, proof traces, or
`todo.md`.

Required fields:

- identity: `schema_version`, `packet_id`, `task_id`, `objective`;
- bounded context: `source_files`, `constraints`, `out_of_scope`;
- result contract: `acceptance_criteria`, `proof_targets`;
- delegation: `model_tier`, `model_tier_reason`, `allowed_tools`,
  `broad_search_policy`, `escalation_triggers`;
- optional deterministic prefetch: `memory_files`, `prefetch`.

Execution telemetry is a separate result linked by `packet_id`; it is not
mutable packet input. Optional telemetry fields normalize runtime output:

- `runtime`, `model`, `predicted_tier`, `actual_tier`, `turns`;
- `input_tokens`, `output_tokens`, `cache_read_tokens`,
  `cache_write_tokens`;
- `rework_count`, `escalation_reason`, `proof_result`,
  `packet_sufficiency`;
- `effective_token_model: raw-v1`.

Raw fields are authoritative. Weighted cost formulas are reporting adapters,
must be versioned, and never replace proof results.

Validate:

```powershell
go run ./cmd/kbcheck context-packet --packet <packet.json>
go run ./cmd/kbcheck execution-telemetry --telemetry <telemetry.json>
```
