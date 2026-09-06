# Mutating startup

Use this path before an edit, test wave, explicit setup or refresh, delegation,
or cleanup.

1. Resolve the repository and ensure project memory is present. Missing
   `todo.md` or `docs/context/PROJECT.md` invokes `kb-map-bootstrap`.
2. Publish the shared work-queue claim before mutation. A missing session ID or
   conflicting active writer stops the route.
3. For ordinary work, use [unfinished-work continuation](unfinished-work.md):
   notice backlog once and prepare independent work without awaiting cleanup.
4. For explicit cleanup, invoke the installed `kb-rehab` owner. Use
   `go run ./cmd/kbcheck terminal-cleanup --action sweep --session-id
   <current-project-session-id> --root <project-root>` only when the consumer
   actually provides that command and its capability/authority predicates pass.
   Missing native tooling selects the installed portable path; it cannot turn
   unrelated cleanup into a prerequisite for the current feature.

Cleanup preserves the current session/worktree, primary checkout, all dirt,
locked or moved worktrees, active claims, rewritten or uncontained commits, and
unresolved paths. It fails closed when authoritative remote-default evidence is
unavailable. A cleanup-only blocker does not stop unrelated work.

No read-only route may run these steps implicitly.
