package main

import "testing"

func TestAdoptAcceptsExplicitCommitAuthority(t *testing.T) {
	base := []string{"plan-worktree", "--action", "adopt", "--manifest", "plan.md", "--owner-token", "owner", "--commit-authorized"}
	if _, err := parse(base); err == nil {
		t.Fatal("incomplete authority accepted")
	}
	opts, err := parse(append(base, "--commit-authorized-by", "user", "--commit-approval-ref", "p2d"))
	if err != nil {
		t.Fatal(err)
	}
	if !opts.commitAuthorized || opts.sliceLeaseAction != "adopt" {
		t.Fatal("authority not preserved")
	}
}
