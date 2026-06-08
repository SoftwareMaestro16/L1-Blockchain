# Public Testnet E2E Smoke Commands

This command list is the operator-facing smoke path for public testnet
readiness. Run the readiness report first; it fails fast when a required
runtime module is only prototype/spec state.

```powershell
.\scripts\testnet\public-testnet-readiness-report.ps1
.\scripts\testnet\public-testnet-readiness-report.ps1 -OutputFormat Json
```

Localnet profiles:

```powershell
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 3
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 5
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 10
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All
```

Focused e2e smoke commands:

```powershell
.\tests\e2e\export_import_smoke.ps1
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\query_surface_smoke.ps1
.\tests\e2e\execution_os_smoke.ps1
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\load_profile_smoke.ps1
```

Runtime module checks:

```powershell
go test ./x/native-account/... ./x/contracts/... ./x/nominator-pool/... ./x/storage-rent/...
go test ./x/aetravm/avm ./x/aetravm/async ./x/vm/types
```

Public testnet is not ready if direct delegation is enabled, if AVM/contracts
or native-account are only types/spec packages, if official pool staking cannot
deposit/claim/unbond through the pool path, or if storage rent cannot freeze
recoverable user/contract state while leaving protocol-critical state executable.
