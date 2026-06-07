> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Minimal Local Load Profile

This profile records prototype baseline behavior under small, controlled local load. It is local/test-only evidence, not a production throughput claim.

Run against a running localnet:

```powershell
.\scripts\localnet\load-profile.ps1 `
  -OutputDir .localnet `
  -Binary build\aetrad.exe `
  -ChainId aetra-local-1 `
  -Scenario mixed `
  -Count 12 `
  -RatePerSecond 2
```

JSON output for reports:

```powershell
.\scripts\localnet\load-profile.ps1 -Scenario mixed -Count 12 -RatePerSecond 2 -Json
```

## Scenarios

| Scenario | Operations |
| --- | --- |
| `bank` | repeated `tx bank send node0 node1 1naet` |
| `tokenfactory` | setup one factory denom, then repeated admin mints |
| `dex` | setup one factory denom and pool, then repeated `swap-exact-in` |
| `mixed` | cycles bank send, tokenfactory mint, DEX swap |

The script measures latency from CLI broadcast through tx inclusion query. Summary fields include `start_height`, `end_height`, `blocks_progressed`, `successes`, `failures`, `failure_rate`, `observed_tps`, per-operation counts, and latency `min/avg/p95/max`.

## Safety

- The script only accepts chain IDs containing `local` and connects to `127.0.0.1:$RPCPort`.
- It uses local test key names from `.localnet` and does not print mnemonics, private keys, environment values, or validator keys.
- It performs no consensus logic changes and does not alter module code.
- Failed txs are recorded in `failure_samples` with truncated errors.

## Smoke And Manual Profiles

Tiny smoke:

```powershell
.\tests\e2e\load_profile_smoke.ps1
```

Default 3-validator observation:

```powershell
.\scripts\localnet\load-profile.ps1 -Scenario mixed -Count 30 -RatePerSecond 3 -Json
```

Manual 5-validator observation:

```powershell
.\scripts\localnet\init.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\scripts\localnet\start.ps1 -OutputDir .localnet-5 -ValidatorCount 5 -NoInit
.\scripts\localnet\load-profile.ps1 -OutputDir .localnet-5 -Scenario mixed -Count 50 -RatePerSecond 3 -Json
```

Heavy load, public-network load, and high-cardinality query load are outside this prototype gate and require a separate scalability plan.
