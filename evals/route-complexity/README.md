# Route Complexity Evals

These fixtures test the KB router contract without live model calls.

Each fixture describes:

- `platforms` - `codex`, `ghcp`, or both.
- `prompt` - user request shape.
- `repo_state` - assumed local context.
- `expected.route` - the smallest correct KB lane.
- `expected.complexity_tier` - `small`, `standard`, or `large`.
- `expected.max_user_questions` - maximum acceptable avoidable questions before action.
- `expected.artifacts` - durable artifacts expected from the route.
- `expected.proof` - proof the route must eventually produce.
- `complexity_signals` - deterministic inputs for the first-pass rubric.
- `guards` - whether the fixture protects against over-planning, under-planning, or skipped gates.

The first runner validates fixture structure and rubric consistency only. Live model route benchmarking is a later layer.
