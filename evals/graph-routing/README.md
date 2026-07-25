# Graph Routing Eval

This deterministic corpus scores graph-routing correctness, retrieval cost, and
local multi-session safety separately. It is provider-optional: required checks
do not install, start, or call Graphify, SCIP, CPG, CodeQL, vectors, MCP, or any
paid/model service.

Promotion requires:

- impacted file/symbol recall at or above the threshold in
  `expected-results.json`;
- false positives per retrieved token at or below the configured ceiling;
- zero safety invariant failures;
- stale or unavailable optional providers recorded as `file-native` fallback or
  `skipped-unavailable`;
- no uncited exact edges or hidden dynamic limitations.

Canonical readiness proof:

```powershell
go run ./cmd/kbcheck graph-routing-eval --require-ready
```
