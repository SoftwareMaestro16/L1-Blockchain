> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Performance And Speed

Phase 15 performance work is measurement-first. Aetra optimizes only when a
Go benchmark, localnet smoke duration, CI timing summary, or load-profile result
shows the before/after cost. Performance changes must not bypass signer, fee, denom, zero-address, authority, or genesis validation.

## Benchmark Baseline

Run targeted baselines with:

```powershell
.\.work\tools\go1.25.11\go\bin\go.exe test -run '^$' -bench 'Benchmark(Parse|ValidateUser|ValidateFee|Dex|Tokenfactory|Queue|ContractState|Token|NFT|SBT|Wallet)' -benchmem ./app/addressing ./x/fees/types ./x/dex/keeper ./x/tokenfactory/keeper ./x/aetherisvm/async ./x/aetherisvm/standards/...
```

Current benchmark coverage:

| Target | Benchmark |
| --- | --- |
| address parsing | `BenchmarkParseRawAddress`, `BenchmarkParseUserFriendlyAddress`, `BenchmarkValidateUserAddress` |
| fee validation | `BenchmarkValidateFeeCoinsAllowedNaet`, `BenchmarkValidateFeeCoinsRejectsNonNaet` |
| DEX keeper paths | `BenchmarkDexSwapExactAmountIn`, `BenchmarkDexSetAndPagePools` |
| tokenfactory keeper paths | `BenchmarkTokenfactoryCreateDenom`, `BenchmarkTokenfactoryMint`, `BenchmarkTokenfactoryGetDenomsPage` |
| async queue processing | `BenchmarkQueueProcessBlock` |
| contract state load/save | `BenchmarkContractStateExportImport` |
| token master and wallet | `BenchmarkTokenMasterMint`, `BenchmarkTokenWalletTransfer` |
| NFT/SBT operations | `BenchmarkNFTMint`, `BenchmarkNFTTransfer`, `BenchmarkSBTProofAndRevoke` |
| wallet contract send | `BenchmarkWalletSignedSend` |

Any optimization PR must paste before/after benchmark numbers and name the
machine or CI runner profile used for the comparison.

## Smoke Duration Baseline

Localnet startup writes a timing summary to:

```text
.localnet/logs/startup-timing.json
```

`scripts/localnet/start.ps1` also prints total, init, process launch, and health
wait milliseconds. CI localnet smoke jobs append a `CI Timing Summary` table to
`GITHUB_STEP_SUMMARY` for adversarial, scaled, and prototype acceptance smoke
runs.

## Repeated Parsing Inventory

Known repeated JSON/protobuf parsing surfaces to measure before optimizing:

- e2e scripts repeatedly parse CLI JSON with `ConvertFrom-Json` for tx results,
  balances, query outputs, health responses, and load-profile summaries;
- localnet health and query helpers repeatedly parse CometBFT status and REST
  responses;
- protobuf tx encode/decode is exercised in integration tests through signed tx
  helpers and malformed protobuf tx tests;
- release/security scripts parse generated JSON summaries and scanner output.

Before caching or sharing parsed values, prove that the cache cannot cross a
state transition, block height, chain-id, signer, fee, denom, or genesis
validation boundary.

## Loop Bounds And Query Limits

User-controlled state iteration must stay bounded:

- tokenfactory list queries use `DefaultQueryDenoms` and `MaxQueryDenoms`;
- DEX pool list queries use `DefaultQueryPools` and `MaxQueryPools`;
- pagination rejects offset, reverse, count_total, wrong-prefix next keys, and
  limits above module maximums through `x/internal/query`;
- AFT wallet accounting, ANFT item membership, AW multi-send, and async queue
  processing keep explicit caps or remain benchmarked blockers before public
  high-cardinality use.

Future token, NFT, contract, and async indexes must add default and max query limits before exposing list endpoints.

## Acceptance Rule

Performance work is acceptable only when:

- before/after numbers are attached;
- no validation shortcut is introduced;
- loop bounds or pagination are preserved;
- localnet and CI smoke timing remain observable.
