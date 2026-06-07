> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Prototype Demo

This is a guided local-only demo path for showing the working Aetra prototype in one visible sequence. It wraps tested localnet commands and is not a substitute for the e2e acceptance suite.

Preview the steps without building or starting localnet:

```powershell
.\scripts\demo\prototype-demo.ps1 -Check
```

Run the demo:

```powershell
.\scripts\demo\prototype-demo.ps1
```

Use an existing binary:

```powershell
.\scripts\demo\prototype-demo.ps1 -SkipBuild -Binary build\aetrad.exe
```

## What It Shows

1. Builds `aetrad` unless `-SkipBuild` is set.
2. Starts a default 3-validator localnet under `.localnet-demo`.
3. Shows block height through CLI and node info through REST.
4. Sends `1000naet` from `node0` to `node1`.
5. Creates a tokenfactory denom and mints factory tokens to `node0`.
6. Creates a DEX pool and swaps `naet` for the factory token.
7. Queries the DEX pool through REST.
8. Prints final `naet`, factory token, and LP balances.
9. Stops localnet unless `-KeepLocalnet` is set.

## Safety

- Local-only: `ChainId` must contain `local`.
- Uses ignored localnet homes and `--keyring-backend test`.
- Does not print mnemonics, private keys, validator keys, node keys, or environment secrets.
- Does not add production faucet or privileged demo logic.

## Related Tests

The demo path is a readable subset of:

- `tests/e2e/prototype_acceptance.ps1`
- `tests/e2e/native_token_smoke.ps1`
- `tests/e2e/tokenfactory_smoke.ps1`
- `tests/e2e/dex_smoke.ps1`
- `tests/e2e/query_surface_smoke.ps1`
