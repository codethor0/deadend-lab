package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Forbidden substrings: if present in cmd/ or pkg/ Go source, could indicate
// accidental secret logging. Goal is to prevent future additions; not proving perfect secrecy.
var forbiddenSecretLogPatterns = []string{
	"session_key",
	"k_enc",
	"k_mac",
	"k_rekey",
	"shared_secret",
	"exported_key",
	`priv=`,
	"private_key",
	"seed=",
	"handshake_secret",
	"key=",
	"plaintext=",
}

func TestNoSecretPatternsInLabAndCryptoCode(t *testing.T) {
	modDir := mustGoModDir(t)
	for _, sub := range []string{"cmd", "pkg"} {
		root := filepath.Join(modDir, sub)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			s := string(content)
			for _, pat := range forbiddenSecretLogPatterns {
				if strings.Contains(s, pat) {
					t.Errorf("forbidden secret pattern %q found in %s", pat, path)
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}
