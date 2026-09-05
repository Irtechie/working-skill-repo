# Session hygiene

Load this reference only at an actual restart decision or when the session is
ending with a dirty tree. Do not interrupt active implementation merely to run
it.

When ending a dirty mutating session, preserve work on its own branch:

```shell
go run ./cmd/kbcheck session-preserve --action apply --session-id <session-id> --json
```

The WIP commit is never pushed, merged, or treated as completion proof. It
refuses the default branch, detached HEAD, and in-progress Git operations, and
reports excluded build artifacts or oversized files. Durability is not delivery.
