# Architecture Cleanup Notes

This note records the refactor baseline and audit constraints for the app wiring cleanup. The intent is to keep behavior and public APIs stable while making future module additions easier to review.

## Baseline

Run before the refactor:

```powershell
$env:PATH = "$PWD\.work\tools\go1.25.11\go\bin;$PWD\.work\tools\bin;$env:PATH"
go test ./...
go vet ./...
buf lint
```

All three checks passed before code changes.

## Before And After

Before:
- `app/app.go` mixed application construction with KV store key inventory and module account permissions.
- `app/services.go` mixed service registration with store key accessors and module account helpers.
- `GetMaccPerms()` cloned the permissions map but reused permission slices.
- `GetStoreKeys()` returned map iteration order, which is not stable for tests and tooling.
- Repository docs did not distinguish generated-code size from authored-code size.

After:
- `app/module_accounts.go` owns module account permissions and blocked-address policy.
- `app/store_keys.go` owns KV store key construction and store key accessors.
- `GetMaccPerms()` returns a defensive copy of both the map and permission slices.
- `GetStoreKeys()` returns keys sorted by `StoreKey.Name()`.
- Governance docs define generated-code, file-size, and panic review policy.

## Panic Review

Startup-only panics are allowed when the process cannot safely continue:
- SDK address-prefix initialization and node-home discovery.
- Interface registry signing-context validation.
- BaseApp streaming service registration.
- Tx config construction.
- Module service and reflection service registration.
- Loading the latest store version during application startup.

Genesis/export panics are tolerated only while the code is part of import/export tooling and not a user transaction path:
- Custom keeper `InitGenesis` and `ExportGenesis` store codec failures.
- Zero-height export preparation in `app/export.go`.

Runtime handlers should return typed errors instead of panicking. In this refactor, no permission checks, authority checks, denom validation, fee policy, or DEX math were changed. Future changes to those paths require a separate security review and targeted tests.

## Generated Code And File Size

Generated files such as `*.pb.go` and `*.pulsar.go` are exempt from the 300-500 line authored-code target. They must not be edited manually. Change the `.proto` source and regenerate instead.

Security and lint tools may exclude generated code when the tool cannot handle it directly, but suppressions must be documented with the generator and reason. Do not disable rules globally to hide findings in authored code.

## Future Module Checklist

When adding a module:
- Add its store key in `app/store_keys.go`.
- Add module account permissions in `app/module_accounts.go` only if it needs a module account.
- Wire keeper dependencies in `app/keepers.go` using narrow expected keeper interfaces.
- Register app modules and deterministic order in `app/modules.go`.
- Update `docs/module-boundaries.md` before adding public messages or keeper dependencies.
- Add tests for authority, denom validation, state import/export, and cross-module flows touched by the module.
