# Consensus Determinism Audit

Branch: `security/determinism-hardening`

## Scope

- `app/abci.go`
- `app/services.go`
- `app/export.go`
- `PreBlocker`, `BeginBlocker`, `EndBlocker`
- custom module keepers and msg servers under `x/tokenfactory`, `x/dex`, `x/fees`
- genesis import/export paths
- vote extension handlers

## Checklist

| Pattern | Finding | Impact | Resolution |
| --- | --- | --- | --- |
| `crypto/rand` / `rand` | `app/abci.go` used `crypto/rand` in dummy vote extensions | Voting path, not AppHash; unsafe for production semantics | Replaced with deterministic SHA-256 payload and disabled unless `orbitalis.vote_extensions.deterministic_for_testing` is explicitly enabled |
| wall-clock time | No `time.Now` in non-test `app` or `x` consensus paths | None | Static regression test blocks new wall-clock usage |
| goroutines/select | No goroutines or `select` in non-generated consensus paths | None | Static regression test blocks new goroutines/select in `app` and `x` |
| map iteration | Static config maps exist in app setup; no unordered map iteration writes consensus state | App setup/off-chain config, not AppHash | No state-machine change required |
| KV iteration | Custom modules use prefixed KVStore iterators | AppHash/export order | Iterator order is key ordered; tokenfactory order-insensitivity test added |
| panic on malformed genesis bytes | `InitChainer` panicked on invalid app state JSON | Genesis import/node startup | Changed to return an error |
| keeper `MustUnmarshal` | Present on data loaded from module stores | AppHash if disk/state is externally corrupted; not reachable by valid txs | Existing corrupted DEX state tests cover no-panic msg execution; deeper store corruption hardening remains a future maintenance item |
| platform-dependent `int/uint` | Generated protobuf code uses `int/uint`; app/test helpers use `int` for test counts | Generated/test/off-chain | No consensus-state dependency found |
| external APIs | No external network/API calls in `app` or `x` state transitions | None | No change required |

## Execution Bounds

- Added tests do not introduce unbounded module scans at block execution time.
- Static source scanning runs only in tests.
- Deterministic vote extension payload is O(len(hash)) and currently 32-byte SHA-256 output.
