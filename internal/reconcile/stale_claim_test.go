package reconcile

import (
	"testing"
	"time"
)

// A claim is evidence that a session is working right now. These tests pin the
// decay of that evidence, because an abandoned claim that protects forever is a
// stuck sensor rather than a safety property.

func claimAt(status string, age time.Duration, now time.Time) QueueClaim {
	claim := QueueClaim{
		WorkID: "w", Status: status,
		Branch: "topic", Worktree: "/repo/wt",
	}
	if age >= 0 {
		claim.UpdatedAt = now.Add(-age)
	}
	return claim
}

func TestFreshClaimStillProtects(t *testing.T) {
	now := time.Now()
	for _, status := range []string{"in_progress", "active", "paused"} {
		if !claimProtectsWorktree(claimAt(status, time.Hour, now), now) {
			t.Fatalf("%s claim one hour old must still protect", status)
		}
	}
}

func TestClaimStopsProtectingOnceStale(t *testing.T) {
	now := time.Now()
	// The observed case: a worktree stranded for 163 hours because the claim
	// naming it never expired.
	if claimProtectsWorktree(claimAt("in_progress", 163*time.Hour, now), now) {
		t.Fatal("a claim stale by days must stop protecting its worktree")
	}
	if claimProtectsWorktree(claimAt("in_progress", StaleClaimAfter, now), now) {
		t.Fatal("a claim exactly at the window must have expired")
	}
	if !claimProtectsWorktree(claimAt("in_progress", StaleClaimAfter-time.Minute, now), now) {
		t.Fatal("a claim just inside the window must still protect")
	}
}

func TestUnstampedClaimKeepsProtecting(t *testing.T) {
	now := time.Now()
	// An unstamped claim proves nothing about its own age, so it cannot be
	// aged out. Fail closed toward preservation.
	if !claimProtectsWorktree(claimAt("in_progress", -1, now), now) {
		t.Fatal("a claim with no UpdatedAt must keep protecting")
	}
}

func TestTerminalClaimNeverProtects(t *testing.T) {
	now := time.Now()
	for _, status := range []string{"done", "blocked", "queued", "delivery-integrated", ""} {
		if claimProtectsWorktree(claimAt(status, time.Minute, now), now) {
			t.Fatalf("%q claim must not protect however fresh", status)
		}
	}
}

func TestStaleClaimExpiryWithdrawsOnlyTheClaimProtection(t *testing.T) {
	now := time.Now()
	repository := Repository{
		Worktrees: []Worktree{{
			Path: "/repo/wt", Branch: "topic",
			ProtectionReasons: []string{"tracked-dirt"},
		}},
		QueueClaims: []QueueClaim{claimAt("in_progress", 163*time.Hour, now)},
	}
	applyQueueProtections(&repository, now)
	reasons := repository.Worktrees[0].ProtectionReasons
	for _, reason := range reasons {
		if reason == "active-claim" {
			t.Fatal("stale claim must not add active-claim")
		}
	}
	// Expiry withdraws one reason. It authorizes nothing: every independent
	// protection has to survive untouched.
	if len(reasons) != 1 || reasons[0] != "tracked-dirt" {
		t.Fatalf("independent protections must survive; got %v", reasons)
	}
}

func TestFreshClaimAddsProtectionByBranchAndByPath(t *testing.T) {
	now := time.Now()
	byBranch := Repository{
		Worktrees:   []Worktree{{Path: "/other", Branch: "topic"}},
		QueueClaims: []QueueClaim{claimAt("in_progress", time.Hour, now)},
	}
	applyQueueProtections(&byBranch, now)
	if !contains(byBranch.Worktrees[0].ProtectionReasons, "active-claim") {
		t.Fatal("fresh claim must protect the worktree it names by branch")
	}
	byPath := Repository{
		Worktrees:   []Worktree{{Path: "/repo/wt", Branch: "other"}},
		QueueClaims: []QueueClaim{claimAt("in_progress", time.Hour, now)},
	}
	applyQueueProtections(&byPath, now)
	if !contains(byPath.Worktrees[0].ProtectionReasons, "active-claim") {
		t.Fatal("fresh claim must protect the worktree it names by path")
	}
}
