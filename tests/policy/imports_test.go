package policy

import (
	"os/exec"
	"strings"
	"testing"
)

const drbgImport = "deadend-lab/internal/drbg"

// allowlist: packages that MAY import internal/drbg (vector generation only).
var drbgAllowlist = map[string]bool{
	"deadend-lab/internal/drbg":           true,
	"deadend-lab/internal/vectorgenerate": true,
}

func TestNoNonVectorPackageImportsDRBG(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./...").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		pkg := strings.TrimSpace(parts[0])
		imports := parts[1]
		if strings.Contains(imports, drbgImport) {
			if !drbgAllowlist[pkg] {
				t.Errorf("package %q imports %s but is not in allowlist; deterministic DRBG must be vector-only", pkg, drbgImport)
			}
		}
	}
}
