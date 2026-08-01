# Replacement Profile Catalog

Select one profile per review boundary. A specialist replaces the broad profile
and inherits the universal questions about intent, test validity, correctness,
and code health.

| Profile | Use when dominant risk is |
|---|---|
| `code-review` | General or unknown |
| `security-reviewer` | Exploitable security or trust boundary |
| `data-migrations-reviewer` | Migration, backfill, persistent data |
| `performance-reviewer` | Scaling, query, transform, or I/O cost |
| `reliability-reviewer` | Retry, timeout, async, queue, recovery |
| `api-contract-reviewer` | Public API, serialization, versioning |
| `cli-readiness-reviewer` | CLI contracts and command handlers |
| `thermo-nuclear-code-quality-reviewer` | Structural simplification dominates |

Do not select by file extension alone. Use changed behavior and failure impact.
When multiple domains apply, choose the highest-consequence risk and include
the secondary concern in that single review prompt.
