package policy

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"deadend-lab/pkg/dee"
)

// Forbidden substrings: if present in cmd/ or pkg/ Go source, or in runtime logs,
// could indicate accidental secret logging. Goal is to prevent future additions; not proving perfect secrecy.
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

// Max hex dump length allowed in logs (prevents accidental raw key/session dumps).
const maxHexInLogs = 64

// TestNoSecretsInRuntimeLogs runs a SAFE scenario and fails if log output contains
// forbidden secret patterns or overly long hex dumps. Catches future regressions.
func TestNoSecretsInRuntimeLogs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	initMsg, initSession, err := dee.HandshakeInit(dee.Safe, nil)
	if err != nil {
		t.Fatalf("HandshakeInit: %v", err)
	}
	respMsg, respSession, err := dee.HandshakeResp(dee.Safe, initMsg, nil)
	if err != nil {
		t.Fatalf("HandshakeResp: %v", err)
	}
	if err := initSession.HandshakeComplete(respMsg); err != nil {
		t.Fatalf("HandshakeComplete: %v", err)
	}
	plaintext := []byte("test message")
	ct, err := initSession.Encrypt(plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	header := respSession.WireHeader(0)
	_, err = respSession.Decrypt(ct, header)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	out := buf.String()
	for _, pat := range forbiddenSecretLogPatterns {
		if strings.Contains(out, pat) {
			t.Errorf("forbidden secret pattern %q found in runtime logs", pat)
		}
	}
	if re := regexp.MustCompile(`[0-9a-fA-F]{65,}`); re.MatchString(out) {
		t.Errorf("runtime logs contain hex dump longer than %d chars", maxHexInLogs)
	}
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
