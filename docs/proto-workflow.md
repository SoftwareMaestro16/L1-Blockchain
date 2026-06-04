# Proto Compatibility And Generation Workflow

Proto files under `proto/` are public contracts for CLI, gRPC, REST, genesis, and module state. Treat them as API changes, not local implementation details.

## Required Workflow

1. Edit source `.proto` files only.
2. Run lint:

   ```powershell
   $env:PATH = "$PWD\.work\tools\bin;$env:PATH"
   buf lint
   ```

3. Generate verification output:

   ```powershell
   buf generate
   ```

   `buf.gen.yaml` writes output to ignored `.work\bufgen`. Do not edit generated Go files by hand.

4. If checked-in generated files must change, copy only the corresponding generated files from `.work\bufgen\github.com\sovereign-l1\l1\x\...\types\` into `x\...\types\`.
5. Verify checked-in generated files exactly match buf output:

   ```powershell
   .\scripts\proto\verify-generated.ps1 -Buf .\.work\tools\bin\buf.exe
   ```

6. Run Go tests for changed messages and any affected CLI/query/e2e smoke.

## Compatibility Rules

- Never reuse field numbers.
- Never change field meaning while keeping the same field number.
- Do not remove fields from a live API without a versioned replacement or an explicit compatibility exception.
- Prefer adding optional fields over changing existing wire shape.
- Breaking REST/gRPC paths require a new versioned API surface or a documented owner decision.
- Keep proto comments concise and focused on public behavior.
- Keep proto-only diffs separate when practical; source/generated/test changes should be easy to review.

## Security Review

For every proto or generated-code change, review:

- Msg signer fields and `GetSigners()` behavior.
- Authority fields such as governance `MsgUpdateParams`.
- Denom/address fields for SDK validation in types or handlers.
- Query request pagination and malformed-request status errors.
- Genesis validation for duplicate state, malformed denoms, invalid params, and impossible module accounting.

Every Critical/High issue found during this review needs a regression test or a documented blocker before merge.

## Expected Checks

```powershell
$env:PATH = "$PWD\.work\tools\go1.25.11\go\bin;$PWD\.work\tools\bin;$env:PATH"
buf lint
.\scripts\proto\verify-generated.ps1 -Buf .\.work\tools\bin\buf.exe
go test ./...
go vet ./...
```
