# Bug Report Template

## Severity
- [ ] CRITICAL: Security property violation, crash, or data loss
- [ ] HIGH: Functional failure under normal conditions
- [ ] MEDIUM: Edge case failure, performance degradation
- [ ] LOW: Cosmetic issue, logging noise

## Category
- [ ] Protocol: Encryption/decryption incorrect
- [ ] Concurrency: Race condition, deadlock
- [ ] Resource: Memory leak, file descriptor leak
- [ ] Input: Malformed payload handling
- [ ] Performance: Timeout, latency spike
- [ ] Logging: Secret material in logs

## Reproduction
Exact command to reproduce:
```
[paste command]
```

Expected behavior:

Actual behavior:

## Root Cause Analysis
[To be filled after investigation]

## Fix Verification
- [ ] Fix implemented
- [ ] Regression test added
- [ ] make verify passes
- [ ] Exploratory test passes
