> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# State Export/Import Acceptance

Prototype state export/import is consensus-critical. The acceptance target is: after local prototype flows, exported genesis validates, contains no local secrets, preserves custom module state, and rejects corrupted import data with a clear validation error.

## Covered State

The acceptance smoke runs bank, staking, tokenfactory, and DEX flows before export:

| Area | Export check |
| --- | --- |
| Chain header | `chain_id` matches the local chain id |
| Fees | `app_state.fees.params.allowed_fee_denoms` is exactly `naet` |
| Tokenfactory | created `factory/{admin}/{subdenom}` denom and admin are present |
| DEX | pool `1` preserves denoms, reserves, `total_shares`, and `lp/1` |
| Bank | account balances preserve factory token, LP token, and funded `naet` |
| Staking | `bond_denom` is `naet`; the delegated validator/delegator entry exists |
| Security | exported JSON does not contain mnemonic, private key, keyring, seed, wallet, or validator key markers |

Run it from the repo root:

```powershell
.\tests\e2e\export_import_smoke.ps1
```

The script uses `.localnet-export-import` and shifted ports by default, then writes exported genesis under `.work\genesis\export-import\node0-export.json`. Both paths are runtime paths and must remain untracked.

## Corrupted Import

The smoke copies the exported state, corrupts `app_state.dex.pools[0].reserve0`, and expects:

```powershell
build\aetrad.exe genesis validate-genesis .work\genesis\export-import\node0-export-corrupt.json --home .localnet-export-import\node0\aetrad
```

to fail with an `invalid` or `reserve0` validation error. The app unit test `TestStateImportRejectsCorruptedExportedPrototypeData` covers the same risk through `BasicModuleManager.ValidateGenesis`.

## Unit Round Trip

`TestStateExportImportPreservesPrototypeModuleData` creates non-empty tokenfactory and DEX state, exports app state, validates module genesis, imports it into a fresh app through `InitChain`, and queries the imported keepers for the same denom, pool, and balances.

## Current Limit

The smoke validates importability with `genesis validate-genesis` and app-level `InitChain`. It does not yet start a multi-validator localnet from the exported genesis because the exported validator set must be paired with matching private validator keys and node topology. A restart-from-export localnet is a future migration/release drill, not a prerequisite for this prototype acceptance gate.
