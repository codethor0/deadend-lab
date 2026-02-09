#!/usr/bin/env bash
# Rewrite git history to replace cursoragent with Thor Thor, then force-push + re-tag.
# Run from a terminal where you have SSH push access to codethor0/deadend-lab.
set -euo pipefail

REPO_SSH="git@github.com:codethor0/deadend-lab.git"
REPO_HTTPS="https://github.com/codethor0/deadend-lab.git"
WORKDIR="$(mktemp -d)"
NAME="Thor Thor"
EMAIL="codethor@gmail.com"
TAG="v0.1.0"

echo "== Clone fresh =="
git clone "$REPO_SSH" "$WORKDIR/deadend-lab"
cd "$WORKDIR/deadend-lab"

echo "== Require git-filter-repo =="
command -v git-filter-repo >/dev/null 2>&1 || { echo "Missing git-filter-repo. Install: brew install git-filter-repo"; exit 1; }

echo "== Rewrite authors/committers matching cursoragent/Cursor Agent =="
git filter-repo --force --commit-callback '
import re
def norm(b):
  return (b or b"").decode("utf-8", "ignore").strip()

patterns = [
  re.compile(r"^cursoragent$", re.I),
  re.compile(r"cursor\s*agent", re.I),
]

an, ae = norm(commit.author_name), norm(commit.author_email)
cn, ce = norm(commit.committer_name), norm(commit.committer_email)

def bad(n, e):
  s = (n + " " + e).strip()
  return any(p.search(n) or p.search(s) for p in patterns)

if bad(an, ae):
  commit.author_name = b"Thor Thor"
  commit.author_email = b"codethor@gmail.com"
if bad(cn, ce):
  commit.committer_name = b"Thor Thor"
  commit.committer_email = b"codethor@gmail.com"
'

echo "== Gate: verify-clean, rc, vectors, git diff =="
make verify-clean
make rc
make vectors
git diff --exit-code

echo "== Force-push rewritten history to main =="
git remote add origin "$REPO_HTTPS" 2>/dev/null || true
GIT_TERMINAL_PROMPT=0 git push --force origin HEAD:main

echo "== Move tag to new history =="
GIT_TERMINAL_PROMPT=0 git push --delete origin "$TAG" 2>/dev/null || true
git tag -f -a "$TAG" -m "deadend-lab research preview $TAG"
GIT_TERMINAL_PROMPT=0 git push -f origin "$TAG"

echo "OK: history rewritten + main force-pushed + tag re-pushed."
echo "Run: gh release edit v0.1.0 ... if you need to update release notes."
