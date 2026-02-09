package policy

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Binary/large file extensions to skip.
var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".pdf": true, ".zip": true, ".tar": true, ".gz": true, ".bz2": true,
	".xz": true, ".7z": true, ".exe": true, ".dll": true, ".so": true,
	".dylib": true, ".bin": true,
}

// Path prefixes to exclude.
var excludePaths = []string{"bin/", "dist/", "vendor/", "node_modules/", "tests/policy/"}

func TestNoVendorToolMentions(t *testing.T) {
	files := trackedTextFiles(t)
	// Cursor, ChatGPT, OpenAI, Claude, Copilot - word boundaries
	re := regexp.MustCompile(`(?i)(^|[^A-Za-z])(Cursor|ChatGPT|OpenAI|Claude|Copilot)([^A-Za-z]|$)`)
	for _, f := range files {
		checkFile(t, f, re, "vendor/tool mention")
	}
}

func TestNoPromptArtifacts(t *testing.T) {
	files := trackedTextFiles(t)
	re := regexp.MustCompile(`(?i)(BEGIN PROMPT|END PROMPT|SYSTEM PROMPT|DEVELOPER PROMPT|You are the Implementation Agent|Master Prompt|paste this into Cursor|prompt artifacts)`)
	for _, f := range files {
		checkFile(t, f, re, "prompt/instruction artifact")
	}
}

func TestNoEmojis(t *testing.T) {
	files := trackedTextFiles(t)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("read %s: %v", f, err)
			continue
		}
		s := string(content)
		for i, r := range s {
			if isEmoji(r) {
				t.Errorf("emoji U+%04X at %s:%d", r, f, lineOf(s, i))
				break
			}
		}
	}
}

func trackedTextFiles(t *testing.T) []string {
	t.Helper()
	modDir := mustGoModDir(t)
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = modDir
	out, err := cmd.Output()
	if err != nil {
		t.Skip("not a git repo")
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "go.sum") || strings.HasSuffix(line, "go.mod") {
			continue
		}
		ext := filepath.Ext(line)
		if binaryExts[ext] {
			continue
		}
		lineSlash := filepath.ToSlash(line)
		skip := false
		for _, p := range excludePaths {
			if strings.HasPrefix(lineSlash, p) || strings.Contains(lineSlash, "/"+p) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		full := filepath.Join(modDir, line)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		files = append(files, full)
	}
	return files
}

func checkFile(t *testing.T, path string, re *regexp.Regexp, label string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %v", path, err)
		return
	}
	if re.Match(content) {
		t.Errorf("forbidden %s in %s", label, path)
	}
}

func isEmoji(r rune) bool {
	// Common emoji Unicode blocks (best-effort)
	return (r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0x1F600 && r <= 0x1F64F) ||
		(r >= 0x1F900 && r <= 0x1F9FF)
}

func lineOf(s string, pos int) int {
	n := 1
	for i := 0; i < pos && i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}
