# Mutating startup

Use this path before an edit, test wave, explicit setup or refresh, delegation,
or cleanup.

1. Resolve the repository and ensure project memory is present. Missing
   `todo.md` or `docs/context/PROJECT.md` invokes `kb-map-bootstrap`.
2. Publish the shared work-queue claim before mutation. A missing session ID or
   conflicting active writer stops the route.
3. For explicit cleanup, prefer the capability-probed reconciler; otherwise use
   `go run ./cmd/kbcheck terminal-cleanup --action sweep --session-id
   <current-project-session-id> --root <project-root>`.

Cleanup preserves the current session/worktree, primary checkout, all dirt,
locked or moved worktrees, active claims, rewritten or uncontained commits, and
unresolved paths. It fails closed when authoritative remote-default evidence is
unavailable. A cleanup-only blocker does not stop unrelated work.

No read-only route may run these steps implicitly.
