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

// TestCommitAttribution enforces that every commit (author and committer) uses the maintainer identity.
// No Co-authored-by trailers are allowed.
func TestCommitAttribution(t *testing.T) {
	modDir := mustGoModDir(t)
	cmd := exec.Command("git", "log", "--all", "--format=%H|%an|%ae|%cn|%ce")
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
		if ae != allowedAuthorEmail {
			t.Errorf("commit %s: author email %q != allowed %q", hash[:12], ae, allowedAuthorEmail)
		}
		if an != allowedAuthorName {
			t.Errorf("commit %s: author name %q != allowed %q", hash[:12], an, allowedAuthorName)
		}
		if ce != allowedAuthorEmail {
			t.Errorf("commit %s: committer email %q != allowed %q", hash[:12], ce, allowedAuthorEmail)
		}
		if cn != allowedAuthorName {
			t.Errorf("commit %s: committer name %q != allowed %q", hash[:12], cn, allowedAuthorName)
		}
	}
}

// TestNoCoAuthoredByTrailer ensures no commit message contains a Co-authored-by trailer.
func TestNoCoAuthoredByTrailer(t *testing.T) {
	modDir := mustGoModDir(t)
	cmd := exec.Command("git", "log", "--all", "--format=%B")
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
