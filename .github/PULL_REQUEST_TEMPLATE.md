## Summary

-

## Security And Consensus Checklist

- [ ] Manual security audit scope is documented; automated tooling is treated as a pre-audit guardrail only.
- [ ] Consensus determinism reviewed: no nondeterministic map iteration, platform-dependent arithmetic, goroutines, random/time sources, or unbounded ABCI paths in consensus code.
- [ ] Critical/high findings from govulncheck, gosec, CodeQL, secret scanning, dependency review, or manual review are fixed, or have a linked issue, severity assessment, owner, and documented decision.
- [ ] False positives are scoped narrowly; no rule is globally disabled without a documented reason.
- [ ] Secrets/logging reviewed: no private keys, mnemonics, tokens, local node homes, validator material, or sensitive payloads are committed or logged.
- [ ] Dependency impact reviewed, including transitive Go module and GitHub Actions changes.
- [ ] Migration/upgrade impact reviewed for state schema, genesis, params, protobuf, CLI, and release artifact compatibility.

## Checks

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `buf lint`
- [ ] `buf generate`
- [ ] `go mod verify`
- [ ] `govulncheck` triage passed or findings are documented above.
- [ ] `gosec` high-severity scan passed or findings are documented above.
- [ ] `gitleaks` passed with redacted output only.
- [ ] Security scan artifacts or CI run links are attached to the PR.
