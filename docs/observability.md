# Observability

Orbitalis exposes two metrics surfaces in localnet:

- CometBFT Prometheus metrics from `config.toml` instrumentation at each manifest `metrics_url`.
- Orbitalis app/module Prometheus metrics at each manifest `app_metrics_url`.

The app metrics endpoint is opt-in and should be bound to loopback or a private operator network:

```powershell
build\orbitalisd.exe start --home .localnet\node0\orbitalisd --observability-metrics --observability-metrics-addr 127.0.0.1:27660
```

Current app metrics include block height/time, FinalizeBlock duration, approximate per-tx processing latency, custom module errors, DEX pool/liquidity/swap counters, fee accept/reject counters, localnet health, and Go runtime memory/goroutine gauges.

Security rules:

- Metrics labels are bounded to fixed keys and sanitize unexpected values.
- Do not add labels containing account addresses, tx hashes, pool IDs, validator keys, mnemonics, DSNs, or environment variables.
- Scripts must not print generated mnemonics unless `-DebugSecrets` is explicitly set.
- Local chain data and logs remain under ignored `.localnet*` paths.

Consensus rule:

Metrics are process-local side effects only. They must not read or write KVStores, change message execution, alter gas, or influence AppHash.

OpenTelemetry plan:

1. Keep Prometheus text metrics as the stable operator interface for v1.
2. Add OpenTelemetry exporters behind explicit config only after metric names and labels are stable.
3. Preserve low-cardinality labels and keep SDK/comet telemetry disabled or loopback-bound unless operators opt in.
4. Add CI snapshot tests for any new public metric name before enabling it in localnet.
