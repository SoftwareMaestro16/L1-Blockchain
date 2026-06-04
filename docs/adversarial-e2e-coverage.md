# Adversarial And E2E Coverage

This suite is split into fast PR gates and heavier localnet checks. The attacker model for custom modules lives in `tests/adversarial/ATTACKER_MODEL.md`.

## Fast PR Gate

- `go test ./app ./observability ./x/... ./tests/adversarial`
  - Keeps existing module tests mandatory.
  - Adds malformed protobuf, unauthorized tokenfactory actions, DEX manipulation, fee abuse, governance abuse, and invalid-fee spam regression tests.
- `go test ./tests/integration`
  - Covers signed transaction lifecycle, invalid signer rejection before mutation, replay-like sequence protection, and tokenfactory-to-DEX-to-fees cross-module state.
- `go vet ./...`
- `buf lint`
- `buf generate`
- `go build -o build\orbitalisd.exe .\cmd\l1d`
- `tests\e2e\adversarial_smoke.ps1`
  - Starts a 3-validator localnet.
  - Submits malformed broadcast bytes, repeated wrong-fee-denom transactions, and a DEX same-denom pool attempt.
  - Verifies the network keeps producing blocks after rejected adversarial inputs.

## Nightly / Manual Gate

- `tests\e2e\localnet_smoke.ps1 -ValidatorCounts 3,5,10 -TimeoutSeconds 180`
  - Runs scaled localnets with restart persistence, bank send, tokenfactory create/mint, DEX pool/swap, fee queries, and negative configuration paths.

## Security Rules

- Do not upload `.localnet*`, keyring data, private validator keys, mnemonics, or generated chain data as CI artifacts.
- Do not print environment variables, GitHub secrets, or node-home file contents in CI logs.
- Keep test data generators deterministic and bounded; use `SpamCount` and validator-count parameters instead of unbounded loops.
- New security regressions must first add a failing adversarial or integration test, then the module fix.

## Residual Gaps

- Query pagination DoS tests should be expanded when DEX and tokenfactory query contracts become paginated/versioned.
- Governance proposal execution through the full `x/gov` lifecycle should be added after custom params modules expose stable proposal fixtures.
- Malformed peer-level P2P packets are handled by CometBFT and are outside this application test suite.
