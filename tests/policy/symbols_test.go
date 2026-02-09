package policy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var forbiddenDeterministicSymbols = []string{
	"HandshakeInitDeterministic",
	"HandshakeRespDeterministic",
}

// allowlist: directories that MAY reference deterministic handshake (vector-only).
var deterministicAllowlist = map[string]bool{
	"internal/vectorgenerate": true,
	"cmd/vectors-gen":         true,
	"pkg/dee":                 true, // defines the funcs
	"tests/policy":            true, // enforcement test; symbols only as string literals
}

func TestDeterministicHandshakeNotReferencedOutsideVectorPath(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{.Dir}}", "./...").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	dirs := strings.Fields(strings.TrimSpace(string(out)))
	for _, dir := range dirs {
		rel, err := filepath.Rel(mustGoModDir(t), dir)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if deterministicAllowlist[rel] {
			continue
		}
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			s := string(content)
			for _, sym := range forbiddenDeterministicSymbols {
				if strings.Contains(s, sym) {
					t.Errorf("forbidden symbol %q found in %s (outside vector path)", sym, path)
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("walk: %v", err)
		}
	}
}

func mustGoModDir(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err != nil {
		t.Fatalf("go list -m: %v", err)
	}
	return strings.TrimSpace(string(out))
}
