# Security Policy

## RESEARCH / CTF ONLY - NOT PRODUCTION CRYPTO

This project is a **research and CTF (Capture The Flag) harness** for studying Dead-End Encryption (DEE) and StegoPQ handshake patterns. It is **NOT intended for production use**.

**Supported scope:**
- SAFE and NAIVE mode lab demonstrations
- Attack harness and demo commands
- Deterministic vector generation for testing

**Out of scope:**
- Production cryptographic implementations
- Secure key management
- Compliance with FIPS or other certification requirements

## Reporting Vulnerabilities

If you discover a vulnerability in this research harness:

1. **Preferred:** Use [Private vulnerability reporting](https://github.com/codethor0/deadend-lab/security/advisories/new) on GitHub for sensitive disclosures.
2. **For research/CTF issues:** Open a GitHub Issue in this repository.
3. **Alternative:** Email codethor@gmail.com if the finding could affect other research tooling or documentation.

We do not offer bug bounties. This is a learning and research project.

## Repository Security Configuration

### Branch Protection
- **Main branch protection**: Enabled
- **Force pushes**: Disabled
- **Branch deletion**: Disabled
- **Required reviews**: At least 1 approval
- **Linear history**: Required
- **Status checks**: Required before merge
- **Conversation resolution**: Required

### Automated Security
- **CodeQL scanning**: Enabled for Go
- **Secret scanning**: Enabled with custom patterns
- **Dependabot alerts**: Enabled (security notifications only)
- **Private vulnerability reporting**: Enabled
- **Security advisories**: Enabled

### Access Control
- **Commit attribution policy**: All commits must be by Thor Thor <codethor@gmail.com>
- **No Co-authored-by trailers**: Enforced by pre-commit hook and CI
- **Contributor control**: Only verified maintainers
