package policy

import (
	"os/exec"
	"strings"
	"testing"
)

const (
	allowedAuthorEmail = "codethor@gmail.com"
	allowedAuthorName  = "Thor Thor"
)

// allowedCommit is true if author+committer match maintainer or allowed automation.
func allowedCommit(an, ae, cn, ce string) bool {
	if an == allowedAuthorName && ae == allowedAuthorEmail && cn == allowedAuthorName && ce == allowedAuthorEmail {
		return true
	}
	// Dependabot dependency update commits (author: dependabot[bot], committer: GitHub)
	if an == "dependabot[bot]" && cn == "GitHub" {
		return true
	}
	// GitHub squash-merge: author Thor Thor (verified email), committer GitHub
	if an == allowedAuthorName && cn == "GitHub" && strings.Contains(ae, "codethor") {
		return true
	}
	return false
}

// TestCommitAttribution enforces that every commit uses maintainer identity or allowed automation (Dependabot).
// No Co-authored-by trailers are allowed.
func TestCommitAttribution(t *testing.T) {
	modDir := mustGoModDir(t)
	cmd := exec.Command("git", "log", "main", "--format=%H|%an|%ae|%cn|%ce")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		t.Skip("not a git repo or git unavailable")
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) != 5 {
			t.Errorf("malformed attrib line: %s", line)
			continue
		}
		hash, an, ae, cn, ce := parts[0], parts[1], parts[2], parts[3], parts[4]
		if !allowedCommit(an, ae, cn, ce) {
			t.Errorf("commit %s: disallowed author %q <%s> committer %q <%s>", hash[:12], an, ae, cn, ce)
		}
	}
}

// TestNoCoAuthoredByTrailer ensures no commit on main contains a Co-authored-by trailer.
func TestNoCoAuthoredByTrailer(t *testing.T) {
	modDir := mustGoModDir(t)
	cmd := exec.Command("git", "log", "main", "--format=%B")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		t.Skip("not a git repo or git unavailable")
	}
	body := string(out)
	if strings.Contains(strings.ToLower(body), "co-authored-by:") {
		t.Error("at least one commit message contains Co-authored-by trailer; not allowed")
	}
}
