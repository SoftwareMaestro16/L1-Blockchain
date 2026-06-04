# Security Tooling And Dependency Hygiene

Security automation is a pre-audit guardrail. It does not replace manual review of consensus-critical state transitions, message authorization, validator economics, or upgrade paths.

## Required CI Gates

- `govulncheck -scan=package -format json ./...` plus `scripts/security/govulncheck_triage.go` blocks new untriaged Go vulnerability IDs and uploads the raw JSON plus triage summary.
- `gosec -conf .gosec.json -exclude-generated -severity high -confidence medium ./...` blocks high-confidence/high-severity Go security findings and uploads SARIF output.
- `gitleaks detect --source . --config .gitleaks.toml --redact --no-banner --verbose` blocks committed secrets while redacting matches in logs and uploads redacted JSON output.
- CodeQL runs `security-and-quality` queries for Go and uploads code scanning results.
- Dependency Review blocks pull requests that introduce high or critical vulnerable dependencies.
- Dependabot and Renovate monitor Go modules and GitHub Actions.

Baseline expectation for a normal PR:

- No untriaged govulncheck IDs.
- No high-severity/high-confidence gosec findings.
- No committed secrets or private runtime material.
- No new high/critical vulnerable dependencies.
- No unresolved high/critical CodeQL or manual audit finding without a linked issue, severity assessment, owner, and documented decision.

## Local Commands

Use the repository toolchain first when it exists:

```powershell
$env:PATH = "$PWD\.work\tools\go1.25.11\go\bin;$PWD\.work\tools\bin;$env:PATH"
go test ./...
go vet ./...
go mod verify
New-Item -ItemType Directory -Force .work\security | Out-Null
New-Item -ItemType Directory -Force .work\tools\bin | Out-Null
$env:GOBIN = "$PWD\.work\tools\bin"
go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
cmd /c "govulncheck -scan=package -format json ./... > .work\security\govulncheck.json"
go run .\scripts\security\govulncheck_triage.go -triage security\govulncheck-triage.json -input .work\security\govulncheck.json
go install github.com/securego/gosec/v2/cmd/gosec@v2.22.10
gosec -conf .gosec.json -exclude-generated -severity high -confidence medium ./...
go install github.com/zricethezav/gitleaks/v8@v8.30.1
gitleaks detect --source . --config .gitleaks.toml --redact --no-banner --verbose
```

Proto checks remain separate:

```powershell
$env:PATH = "$PWD\.work\tools\bin;$env:PATH"
buf lint
buf generate
```

## Triage Policy

- Critical/high findings block push and merge unless there is a linked issue, severity assessment, owner, and documented decision.
- Current accepted govulncheck findings are listed in `security/govulncheck-triage.json`; adding a new ID requires fixing it or adding owner/issue/decision first.
- False positives must be scoped narrowly. Prefer a code comment or path-specific allowlist over disabling a rule globally.
- Generated `.pb.go` and `.pb.gw.go` files are excluded from gosec/CodeQL/Gitleaks scans; the source `.proto` files remain in scope.
- Secret scanning must never print full findings. Use redacted output only.
- Dependency updates with security labels are reviewed before routine feature work.
- Run `govulncheck -scan=symbol ./...` manually for deep reachability triage on a machine with enough memory before external audit milestones.

Every suppression or accepted finding must include:

- Finding ID or rule name.
- Affected dependency, package, or path.
- Assessed severity and exploitability for Orbitalis.
- Linked issue or audit ticket.
- Owner and review date.
- Decision: fix now, accept temporarily, or wait for upstream.

## Update Policy

- Dependabot is the default routine PR source for weekly patch/minor Go module and GitHub Actions updates.
- Renovate keeps a dependency dashboard for Go modules and GitHub Actions. Routine Renovate PRs require dashboard approval to avoid duplicating Dependabot PRs.
- Major updates are separate, require explicit Renovate dashboard approval, and must include consensus, serialization, CLI, migration, and release-artifact review.
- Security advisories are allowed to cut ahead of routine schedules.
- Dead dependencies are removed only after `go mod tidy`, `go test ./...`, and `go vet ./...` agree with the change.

## Suppression Rules

- `.gosec.json` contains no project suppressions; suppression must be inline and justified when needed.
- `.gitleaks.toml` extends default rules and only allowlists generated, build, and local runtime paths.
- CodeQL ignores generated protobuf, build, and localnet cache paths only.
- Dependabot ignores semver-major version-update PRs because Renovate handles those through explicit dashboard approval.
