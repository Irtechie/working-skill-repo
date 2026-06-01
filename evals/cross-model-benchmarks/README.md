# Cross-Model Benchmarks

These fixtures are prompt packs for later Codex, GHCP, Claude, or other host
comparison runs. The local gate only validates fixture shape and expected
deterministic scoring fields. It does not call live models.

Each fixture file contains cases with:

- `id`: stable case id.
- `category`: route-selection, proof-discipline, or minimalism.
- `prompt`: the user request shown to the model.
- `expected`: deterministic expected route or behavior.
- `forbidden_failures`: failure modes that should be scored as regressions.
- `scoring`: named dimensions with pass criteria.

Live execution must be explicit and should write captured model outputs under
`.atv/eval-runs/` before any comparison report is generated.
