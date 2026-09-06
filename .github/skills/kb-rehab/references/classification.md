# Classification and preservation

| Result | Evidence and next action |
|---|---|
| preserve-live | Current/default/occupied worktree, active or uncertain claim, or credential path. Preserve; resolve the owning scope without taking it over. |
| review-needed | Unknown ownership, failed/missing proof, or missing exact manifest/ref/tip binding. Review without deletion. |
| salvage | Exact accepted artifact still matches its hash. Keep the source; its owner may copy/commit the selected work separately. |
| discard-confirmed-junk | Explicitly rejected local branch with a restoration-proven bundle, or explicitly rejected generated output with matching provenance and a restored recovery copy. Report the archive and observed action. |
| deliver-pr | Exact contained manifest/hash/ref/tip pairing; delegate fresh proof/review and PR work to kb-complete in its own workspace. This is pending delivery. |
| merge-eligible | Current explicit merge intent or configured auto-after-checks permits a pending owner attempt. Actual forge checks remain required; completed stays false. |
| retained-blocked | Current authority, ownership, policy, hash, restore, native capability or forge evidence does not permit the specific operation. Preserve and name the next owner. |

Ancestry counts and patch equivalence are review signals, not deletion proof.
Squash/rebase equivalence alone cannot authorize retirement. Already-landed
claims require actual forge merge evidence plus content preservation checked
by the owning delivery/reconciliation mechanism. The portable helper instead
supports the narrower explicit-rejection/restoration route.

No policy file is required for portable survey/preservation. Unknown remote
or malformed policy removes destructive eligibility, not unrelated task
progress. No command from an unverified manifest is executed by the helper.

Age, ignore status, zero bytes and failed tests never classify junk. Expired
claims are review signals; the portable helper conservatively preserves every
nonterminal claim and every occupied worktree. The current/source worktree is
never removed. Exact generated-file deletion is restricted to `.kb/generated/`
or `.kb/tmp/`, with explicit rejection and independently readable hash-bound
provenance; a producer label alone grants no authority.

Archives live under `<git-common-dir>/.copilot-kb/recovery/dispositions/`.
Branch recovery includes `branch.bundle`, a restored bare repository, selected
artifact copies and restored copies, and a receipt listing hashes and excluded
credential paths. Originals stay in place. No credentials are copied or
removed; an excluded credential does not become junk. A restore mismatch or
changed ref blocks retirement. Ref deletion compares the expected old tip.
