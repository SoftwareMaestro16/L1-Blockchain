> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Tx Event Contract

Events are supporting evidence for e2e, CLI debugging, and future indexers. They do not replace final state queries. Every e2e assertion that checks a custom event must also verify committed state through a query or keeper/bank state in tests.

## Rules

- Event types and attribute keys are public prototype API.
- Attribute order in code is deterministic and fixed; tests search by key so future SDK event additions do not break the contract.
- Attributes may contain bounded user-controlled strings only after normal tx validation: Bech32 addresses, SDK denoms, integer amounts, and pool IDs.
- Events must not include mnemonics, private keys, local paths, environment values, raw transaction bytes, logs, or unbounded payloads.
- Events must not iterate maps or depend on wall clock, randomness, memory addresses, goroutines, or external APIs.
- Failed tx paths must not emit success events. Rejected fee-denom txs are evidenced by stable error text and unchanged state, not by a committed custom event.
- Events are low-cardinality evidence, not metrics. Do not add account-specific labels to telemetry from these events without a separate observability review.

## Tokenfactory Events

| Tx | Event type | Attributes |
| --- | --- | --- |
| `MsgCreateDenom` | `tokenfactory_create_denom` | `denom`, `creator`, `admin` |
| `MsgMint` | `tokenfactory_mint` | `denom`, `sender`, `amount`, `mint_to_address` |
| `MsgBurn` | `tokenfactory_burn` | `denom`, `sender`, `amount`, `burn_from_address` |
| `MsgChangeAdmin` | `tokenfactory_change_admin` | `denom`, `sender`, `new_admin` |

State query verification:

- `query tokenfactory denom <denom>`
- `query bank denom-metadata <denom>`
- `query bank balance <addr> <denom>`
- `query bank total-supply-of <denom>`

## DEX Events

| Tx | Event type | Attributes |
| --- | --- | --- |
| `MsgCreatePool` | `dex_create_pool` | `pool_id`, `creator`, `denom0`, `denom1`, `amount0`, `amount1`, `lp_denom`, `minted_shares` |
| `MsgAddLiquidity` | `dex_add_liquidity` | `pool_id`, `depositor`, `denom0`, `denom1`, `amount0`, `amount1`, `lp_denom`, `minted_shares` |
| `MsgRemoveLiquidity` | `dex_remove_liquidity` | `pool_id`, `withdrawer`, `lp_denom`, `shares`, `denom0`, `denom1`, `amount0`, `amount1` |
| `MsgSwapExactAmountIn` | `dex_swap_exact_amount_in` | `pool_id`, `trader`, `token_in`, `token_out` |

State query verification:

- `query dex pool <pool_id>`
- `query bank balance <addr> <lp_denom>`
- `query bank balance <addr> <token_denom>`
- keeper tests compare pool reserves with the DEX module account balances and LP supply.

## Fees Events

| Tx | Event type | Attributes |
| --- | --- | --- |
| `MsgUpdateParams` | `fees_update_params` | `authority`, `allowed_fee_denom`, `validator_rewards_ratio`, `community_pool_ratio` |

State query verification:

- `query fees params`
- REST `/l1/fees/v1/params`

Fee-denom rejection:

- Wrong fee denom, multi-denom fee, malformed fee, and non-`FeeTx` rejection paths are ante failures.
- They must not be treated as committed custom events.
- E2E checks must assert the stable error message, then use state queries to prove no expected state changed.

## E2E Assertions

`scripts/localnet/lib/cli.ps1` exposes `Assert-LocalnetTxEvent` for committed tx query results. Current e2e assertions cover:

- tokenfactory create/mint/burn/change-admin in `tests/e2e/tokenfactory_smoke.ps1`;
- DEX create-pool/add-liquidity/swap/remove-liquidity in `tests/e2e/dex_smoke.ps1`;
- fee rejection as error evidence in `tests/e2e/fees_ante_smoke.ps1` and `tests/e2e/prototype_acceptance.ps1`.

## Regression Tests

| Layer | Evidence |
| --- | --- |
| Event constants | `x/tokenfactory/types/events.go`, `x/dex/types/events.go`, `x/fees/types/events.go` |
| Integration/unit event assertions | `x/tokenfactory/keeper/msg_server_test.go`, `x/dex/keeper/msg_server_test.go`, `x/fees/keeper/msg_server_test.go` |
| E2E event assertions | `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/dex_smoke.ps1` |
| Event contract doc guard | `tests/scripts/event_contract_doc_test.ps1` |

## Change Control

Changing an event type or attribute key is a public-interface change for the prototype. It requires:

1. updating this document,
2. updating module constants,
3. updating integration and e2e assertions,
4. running `go test ./...`, `go vet ./...`, `buf lint`, and the relevant e2e smoke,
5. recording any indexer-impacting change in release notes.
