# CosmWasm Readiness

CosmWasm stays disabled in Aetra until the base chain hardening phases pass
and the app wiring explicitly opts into `x/wasm`.

## Version Decision

Use `github.com/CosmWasm/wasmd v0.70.2` as the first integration candidate. Its module metadata targets Go `1.25.9`, Cosmos SDK `v0.54.0`, CometBFT `v0.39.0`, and `github.com/CosmWasm/wasmvm/v3 v3.0.6`, which is the closest match to the current Aetra SDK `v0.54.x` base. Older wasmd lines inspected for this decision target SDK `v0.50` or `v0.53` and should not be used for the initial integration.

Do not add the real `x/wasm` keeper until the integration branch also wires the required keeper dependencies, wasm store keys, module account permissions, ante decorators, genesis validation, export, and e2e contract smoke tests.

## Feature Gate

Default policy:

- `wasm.enabled = false`
- no `wasm` store key in the app
- no `wasm` genesis state
- no `tx wasm` upload, instantiate, execute, or migrate surface

The future gate must be explicit in config or startup flags. Localnet and public testnet configs must not silently enable wasm because a dependency was added.

Only the configured governance authority can enable or disable the gate. The
authority address must be a valid non-zero Aetra authority address, and
policy updates must validate the full next policy before it can become active.

## Permissions

Upload:

- Public testnet and mainnet default: governance-only.
- Dev testnet exception: allowlist upload is allowed only with a non-empty allowlist of valid non-zero Aetra user addresses.

Instantiate:

- Default: code-owner-only.
- Dev testnet may use everybody-instantiates only after upload remains governance-only or allowlisted.

Admin and migration:

- Contract admin must be a valid non-zero Aetra address.
- Migration is admin-only.
- Empty admin, zero admin, unauthorized migrate, and admin takeover attempts are rejected.

Pinned code:

- Default: disabled.
- If used later, pinning is governance-only and bounded by `max_pinned_codes`.
- Pinning cannot be enabled with `max_pinned_codes = 0`.

## Limits

Initial guarded limits:

- max stored contract size: `800 KiB`
- max proposal contract size: `3 MiB`
- smart query gas limit: `3,000,000`
- simulation gas limit: `20,000,000`
- gas multiplier: `140,000`
- memory cache: `100 MiB`, hard cap `256 MiB`
- smart query response limit: `256 KiB`, hard cap `1 MiB`
- smart query depth limit: `8`, hard cap `16`
- pinned code count: `0` by default, hard cap `128`

Query limits are part of the feature gate. Smart query gas, simulation gas, and
query response/depth limits must remain bounded before any public CosmWasm
enablement.

Changing the gas multiplier or increasing limits requires benchmarks, adversarial tests, and a security checklist update.

Fee and address safety:

- All CosmWasm upload, instantiate, execute, query, and migrate transactions
  must pay protocol fees in `naet`.
- CosmWasm contracts cannot pay base-chain protocol fees with user-created
  tokens, LP tokens, NFT/SBT assets, or test denoms.
- Contract admin, instantiate recipient, execute actor, migrate actor, and
  contract address inputs must reject empty, malformed, and zero addresses.

## Tests

Current readiness tests cover:

- default app has no `wasm` store or genesis state,
- disabled feature gate blocks upload,
- governance-only upload,
- allowlist upload with malformed, empty, and zero-address rejection,
- instantiate, execute, and migrate authorization policy,
- unauthorized migrate and zero-admin takeover rejection.
- gas limit and query limit enforcement,
- contract code size enforcement,
- pinned code disabled/governance-only policy,
- governance-only enable/disable policy,
- non-`naet` fee rejection,
- zero admin, zero recipient, and zero contract address rejection.

When real `x/wasm` is wired, add keeper/app tests for `MsgStoreCode`, `MsgInstantiateContract`, `MsgExecuteContract`, `MsgMigrateContract`, genesis import/export, max-size rejection, gas-limit rejection, and unauthorized admin changes.

## Localnet Smoke

After the real gated `x/wasm` integration exists:

```powershell
.\tests\e2e\cosmwasm_smoke.ps1 -EnableWasm -ContractWasm .\artifacts\cw_template.wasm -ExecuteMsg "{}" -QueryMsg "{}"
```

Without `-EnableWasm`, the smoke script asserts that `query wasm params` is unavailable, proving wasm is still disabled by default.

When migration is explicitly permitted by policy, include `-MigrateWasm` and
`-MigrateMsg`; otherwise the smoke deliberately skips migration.

## Rust Contract Flow

Build a simple test contract:

```powershell
rustup target add wasm32-unknown-unknown
cargo install cargo-generate
cargo install cosmwasm-check
cargo generate --git https://github.com/CosmWasm/cw-template.git --name aetra-smoke
cd aetra-smoke
cargo wasm
cosmwasm-check .\target\wasm32-unknown-unknown\release\aetra_smoke.wasm
```

Deploy only on a localnet or testnet where the explicit wasm gate is enabled:

```powershell
build\aetrad.exe tx wasm store .\target\wasm32-unknown-unknown\release\aetra_smoke.wasm --from node0 --home .localnet\node0\aetrad --keyring-backend test --chain-id aetra-local-1 --node tcp://127.0.0.1:26657 --fees 1000000naet -y
build\aetrad.exe tx wasm instantiate 1 "{}" --from node0 --label aetra-smoke --admin <admin-address> --home .localnet\node0\aetrad --keyring-backend test --chain-id aetra-local-1 --node tcp://127.0.0.1:26657 --fees 1000000naet -y
build\aetrad.exe tx wasm execute <contract-address> "{}" --from node0 --home .localnet\node0\aetrad --keyring-backend test --chain-id aetra-local-1 --node tcp://127.0.0.1:26657 --fees 1000000naet -y
build\aetrad.exe query wasm contract-state smart <contract-address> "{}" --node tcp://127.0.0.1:26657
```
