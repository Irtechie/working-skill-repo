package reconcile

import (
	"fmt"
	"sort"
	"strings"
)

// RemoteAuthorityState names how much a caller may conclude from a remote.
const (
	RemoteAuthorityAuthoritative = "authoritative"
	RemoteAuthorityUnavailable   = "unavailable"
)

// RemoteAuthoritySource is the probe every caller must attribute, so a report
// cannot claim fresh authority while reading a cached ref.
const RemoteAuthoritySource = "git ls-remote --symref"

// RemoteAuthority is the single answer to "what is the authoritative default
// branch right now, and may an irreversible decision rely on it".
//
// A remote-tracking ref is a cache. Containment computed against a cache proves
// only what the cache last saw, which is why a cached observation never reaches
// state authoritative here.
type RemoteAuthority struct {
	State         string `json:"state"`
	Remote        string `json:"remote,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	SHA           string `json:"sha,omitempty"`
	Limitation    string `json:"limitation,omitempty"`
}

// Authoritative reports whether an irreversible decision may rely on this.
func (authority RemoteAuthority) Authoritative() bool {
	return authority.State == RemoteAuthorityAuthoritative &&
		authority.DefaultBranch != "" && authority.SHA != ""
}

// ResolveRemoteAuthority establishes fresh remote authority over the network.
//
// Every remote must agree on the default branch and its SHA. Disagreement is
// not resolved by preferring one remote: an ambiguous answer is unavailable,
// because a caller about to delete a ref cannot be told a probable default.
//
// This is the adapter whose absence makes an irreversible decision report
// authoritative-containment-unavailable. Callers that must not touch the
// network should not call it and should treat cached refs as cached.
func ResolveRemoteAuthority(root string, remotes []Remote) RemoteAuthority {
	if len(remotes) == 0 {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Limitation: "no configured remote; containment cannot be proven",
		}
	}
	names := make([]string, 0, len(remotes))
	for _, remote := range remotes {
		if name := strings.TrimSpace(remote.Name); name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var resolved RemoteAuthority
	for _, name := range names {
		candidate := resolveRemoteAuthorityFor(root, name)
		if !candidate.Authoritative() {
			return candidate
		}
		if resolved.State == "" {
			resolved = candidate
			continue
		}
		if resolved.SHA != candidate.SHA || resolved.DefaultBranch != candidate.DefaultBranch {
			return RemoteAuthority{
				State: RemoteAuthorityUnavailable,
				Limitation: fmt.Sprintf("remotes disagree on the default branch: %s/%s vs %s/%s",
					resolved.Remote, resolved.DefaultBranch, candidate.Remote, candidate.DefaultBranch),
			}
		}
	}
	if resolved.State == "" {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Limitation: "no usable remote name; containment cannot be proven",
		}
	}
	return resolved
}

// resolveRemoteAuthorityFor proves one remote's default branch is both
// advertised and locally present at the advertised commit.
func resolveRemoteAuthorityFor(root, remote string) RemoteAuthority {
	advertised, err := gitOutput(root, "ls-remote", "--symref", remote, "HEAD")
	if err != nil {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Remote:     remote,
			Limitation: fmt.Sprintf("remote %s is unreachable", remote),
		}
	}
	branch, sha := ParseSymrefAdvertisement(advertised)
	if branch == "" || sha == "" {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Remote:     remote,
			Limitation: fmt.Sprintf("remote %s advertised no resolvable default branch", remote),
		}
	}
	if _, err := gitOutput(root, "fetch", "--no-tags", remote, branch); err != nil {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Remote:     remote,
			Limitation: fmt.Sprintf("fetch of %s/%s failed", remote, branch),
		}
	}
	fetched, err := gitOutput(root, "rev-parse", "FETCH_HEAD")
	if err != nil {
		return RemoteAuthority{
			State:      RemoteAuthorityUnavailable,
			Remote:     remote,
			Limitation: fmt.Sprintf("fetched head of %s/%s is unresolvable", remote, branch),
		}
	}
	// The advertisement and the fetch are two round trips. A default branch that
	// moved between them means this run observed two different worlds, so it
	// cannot authorize an irreversible decision about either.
	if fetched = strings.TrimSpace(fetched); fetched != sha {
		return RemoteAuthority{
			State:  RemoteAuthorityUnavailable,
			Remote: remote,
			Limitation: fmt.Sprintf("remote %s default moved between advertisement %s and fetch %s",
				remote, ShortSHA(sha), ShortSHA(fetched)),
		}
	}
	return RemoteAuthority{
		State:         RemoteAuthorityAuthoritative,
		Remote:        remote,
		DefaultBranch: branch,
		SHA:           sha,
	}
}

// ParseSymrefAdvertisement reads the default branch and SHA out of
// `git ls-remote --symref <remote> HEAD`.
func ParseSymrefAdvertisement(output string) (branch, sha string) {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		if fields[0] == "ref:" && len(fields) >= 3 && fields[2] == "HEAD" {
			branch = strings.TrimPrefix(fields[1], "refs/heads/")
			continue
		}
		if fields[1] == "HEAD" {
			sha = fields[0]
		}
	}
	return branch, sha
}

// ShortSHA abbreviates a commit for a human-facing limitation string.
func ShortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
