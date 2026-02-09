#!/usr/bin/env bash
# Audit and optionally rewrite git history; force-push and re-tag.
# Run from repo root with clean tree. Requires: git-filter-repo, gh (for HTTPS push).
set -euo pipefail

REPO="codethor0/deadend-lab"
REPO_HTTPS="https://github.com/${REPO}.git"
WORKDIR="$(mktemp -d)"
NAME="Thor Thor"
EMAIL="codethor@gmail.com"
TAG="v0.1.0"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT"
test -z "$(git status --porcelain)" || { echo "Working tree dirty. Commit or stash first."; exit 1; }
git remote get-url origin | grep -q "deadend-lab" || { echo "Origin does not point at deadend-lab."; exit 1; }

echo "== Audit: authors =="
git log --all --format='%an <%ae>' | sort | uniq -c | sort -rn
echo ""
echo "== Audit: committers =="
git log --all --format='%cn <%ce>' | sort | uniq -c | sort -rn
echo ""
echo "== Audit: annotated tag objects =="
for t in $(git tag -l 2>/dev/null); do
  obj=$(git rev-parse "$t" 2>/dev/null)
  typ=$(git cat-file -t "$obj" 2>/dev/null)
  if [ "$typ" = "tag" ]; then
    echo "Tag $t (object $obj):"
    git cat-file -p "$obj" 2>/dev/null | grep -E '^(tagger|author|committer) ' || true
  fi
done

BAD_AUTH=$(git log --all --format='%an <%ae>' | sort -u | grep -v "^${NAME} <${EMAIL}>$" || true)
BAD_COMM=$(git log --all --format='%cn <%ce>' | sort -u | grep -v "^${NAME} <${EMAIL}>$" || true)
if [ -z "$BAD_AUTH" ] && [ -z "$BAD_COMM" ]; then
  echo "All identities compliant. No rewrite needed."
  echo "Run: git fetch origin && ./scripts/pre-push-gate.sh"
  exit 0
fi

echo "== Non-compliant identities found. Clone and rewrite. =="
git clone "$REPO_HTTPS" "$WORKDIR/deadend-lab"
cd "$WORKDIR/deadend-lab"
command -v git-filter-repo >/dev/null 2>&1 || { echo "Missing git-filter-repo. Install: brew install git-filter-repo"; exit 1; }

git filter-repo --force --commit-callback '
def norm(b):
  return (b or b"").decode("utf-8", "ignore").strip()
ok_name = b"Thor Thor"
ok_email = b"codethor@gmail.com"
an, ae = commit.author_name, commit.author_email
cn, ce = commit.committer_name, commit.committer_email
if norm(an) != "Thor Thor" or norm(ae) != "codethor@gmail.com":
  commit.author_name = ok_name
  commit.author_email = ok_email
if norm(cn) != "Thor Thor" or norm(ce) != "codethor@gmail.com":
  commit.committer_name = ok_name
  commit.committer_email = ok_email
' --tag-callback '
def norm(b):
  return (b or b"").decode("utf-8", "ignore").strip()
ok_name = b"Thor Thor"
ok_email = b"codethor@gmail.com"
if hasattr(tag, "tagger_name") and tag.tagger_name:
  if norm(tag.tagger_name) != "Thor Thor" or norm(tag.tagger_email) != "codethor@gmail.com":
    tag.tagger_name = ok_name
    tag.tagger_email = ok_email
'

echo "== Gate =="
make verify-clean
make rc
make vectors
git diff --exit-code
./scripts/pre-push-gate.sh

echo "== Force-push main and tag =="
git remote add origin "$REPO_HTTPS" 2>/dev/null || true
GIT_TERMINAL_PROMPT=0 git push --force origin HEAD:main
GIT_TERMINAL_PROMPT=0 git push --delete origin "$TAG" 2>/dev/null || true
git tag -f -a "$TAG" -m "deadend-lab research preview $TAG"
GIT_TERMINAL_PROMPT=0 git push -f origin "$TAG"

echo "OK. Post-push checks:"
echo "  git fetch origin && git log --format='%an <%ae>' | sort | uniq -c"
echo "  git show --no-patch --decorate $TAG"
echo "  https://github.com/${REPO}/graphs/contributors"
