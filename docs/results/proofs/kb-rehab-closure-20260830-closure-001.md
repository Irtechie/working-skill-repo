# Closure-001 Proof: Work-Reality Removal Safety

Tree: `rehab-closure-rationalization` before the slice commit.

Command:

```powershell
go test ./cmd/kbcheck -run 'WorkReality|RehabRemoval' -count=1
```

Result: pass.

The protected behavior is asymmetric:

- an uncontained `in_progress` row remains byte-for-byte unchanged;
- a terminal contained row without a resolving artifact remains unchanged;
- a terminal contained row with a landed resolving artifact is removed.

The first regression test failed against the previous behavior, which rewrote
`🔧 in_progress` to `🔒 blocked` and appended a `kb-rehab` note to the Link
cell. The repaired path reports `preserved` with zero writes instead.
