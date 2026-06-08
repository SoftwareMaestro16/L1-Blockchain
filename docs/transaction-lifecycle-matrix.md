> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Transaction Lifecycle Matrix

This matrix traces every prototype transaction from operator CLI input to final state query. It is the transaction-level companion to [Prototype Contract](prototype-contract.md), [Operator Commands](operator-commands.md), [Prototype Tx Event Contract](event-contract.md), [Prototype Test Pyramid](test-pyramid.md), and the module flow docs.

Rules:

- Transaction behavior changes require module-boundary review and targeted tests in the row being changed.
- Public proto changes require the buf lint/generation workflow before merge.
- All example fees use `1000000naet`; `AET`, factory denoms, LP denoms, and `testtoken` are not valid fee denoms.
- Transaction paths must use direct key lookups or bounded structures. No transaction path may require a full store scan.
- Custom module events are supporting evidence only; state queries remain the source of truth.

## Security Review Lens

The Cosmos-specific review for every row must cover:

- signer mismatch or missing authorization,
- invalid Bech32 account/operator addresses,
- invalid or spoofed denoms,
- insufficient funds or failed bank keeper calls,
- duplicate state and replay/sequence failure,
- malformed amount, pool, param, or query fields,
- ABCI panic and nondeterministic state-write risk.

## Lifecycle Matrix

| Tx | Actor | Signer | CLI input | Funds and fee | State writes | Observable events | Verification queries | Tests |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Bank send | funded local account `node0` sends native funds to `node1` or another `AE...` account | `--from node0`; SDK `MsgSend` signer must be sender | `tx bank send node0 $NODE1 1000naet --fees 1000000naet` | sender must hold `1000naet` plus fee; fee denom must be `naet` | `x/bank` sender/receiver balances and fee deduction | SDK tx/message and bank transfer events | `query bank balance $NODE1 naet`; `query tx <hash>` | `tests/e2e/native_token_smoke.ps1`, `tests/e2e/localnet_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1`, `x/fees/keeper/ante_test.go` |
| Official liquid staking deposit | local user deposits `naet` into the official liquid staking pool/index without choosing a validator | `--from node0`; pool deposit signer must be depositor | official pool deposit tx or contract execute with pool id/contract and `5000000naet`; no validator address argument | depositor must hold deposited `naet` plus fee; fee denom must be `naet` | official pool share/receipt state and bounded pool accounting; no user `x/staking` delegation key | deterministic pool deposit/share events plus SDK tx/message events | pool share/reward query; validator allocation query from pool/index state | `app/pos_test.go`, `tests/e2e/pos_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Direct staking delegate rejected | normal user attempts to bypass the pool/index and choose a validator | `--from node0`; SDK `MsgDelegate` signer is still checked by tx auth | `tx staking delegate $VALOPER 5000000naet --fees 1000000naet` | no delegation or bonded-pool movement is allowed | rejected by pool-only staking policy before staking keeper mutation; validator tokens/power stay unchanged | deterministic error: `direct user delegation to validators is disabled; use official liquid staking pool deposit` | `query staking delegation $DELEGATOR $VALOPER` remains empty; validator power unchanged | `app/pos_test.go`, `tests/integration/pos_lifecycle_test.go` |
| Tokenfactory create denom | creator defines a new factory denom under own address | `--from node0`; `MsgCreateDenom.Creator` from CLI account | `tx tokenfactory create-denom gold --fees 1000000naet` | creator pays `naet` fee only; no token funds moved | `x/tokenfactory` denom authority metadata; bank denom metadata for `factory/<creator>/gold` | `tokenfactory_create_denom` plus SDK tx/message events | `query tokenfactory denom $DENOM`; `query bank denom-metadata $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `x/tokenfactory/keeper/query_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Tokenfactory mint | current denom admin mints factory supply to a recipient | `--from <admin>`; `MsgMint.Sender` must equal current admin | `tx tokenfactory mint "1000000$DENOM" $TO --fees 1000000naet` | admin pays `naet` fee; module mints factory denom; recipient may be any valid `AE...` | bank supply and recipient balance for factory denom | `tokenfactory_mint` plus SDK tx/message, bank mint, and transfer events | `query bank balance $TO $DENOM`; `query bank total-supply-of $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Tokenfactory burn | current denom admin burns own factory balance | `--from <admin>`; `MsgBurn.Sender` must equal current admin and `burn_from_address` | `tx tokenfactory burn "250000$DENOM" $ADMIN --fees 1000000naet` | admin must hold burned amount plus `naet` fee | bank supply and admin balance decrease for factory denom | `tokenfactory_burn` plus SDK tx/message, bank send-to-module, and burn events | `query bank balance $ADMIN $DENOM`; `query bank total-supply-of $DENOM` | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1` |
| Tokenfactory change admin | current denom admin transfers authority | `--from <old-admin>`; `MsgChangeAdmin.Sender` must equal current admin | `tx tokenfactory change-admin $DENOM $NEW_ADMIN --fees 1000000naet` | old admin pays `naet` fee only | `x/tokenfactory` denom authority metadata admin field | `tokenfactory_change_admin` plus SDK tx/message events | `query tokenfactory denom $DENOM`; follow-up old-admin failure and new-admin mint success | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1` |
| Fees update params | governance authority updates fee policy params | governance module authority; no operator CLI business logic | gRPC/proposal path for `l1.fees.v1.MsgUpdateParams`; no local operator shortcut | governance tx/proposal fees in `naet`; params must validate before write | `x/fees` params store | `fees_update_params` plus SDK tx/message events | `query fees params`; REST `/l1/fees/v1/params` | `x/fees/keeper/msg_server_test.go`, `x/fees/types/genesis_test.go`, `x/fees/keeper/query_server_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| DEX create pool | liquidity provider creates a constant-product pair | `--from node0`; `MsgCreatePool.Creator` from CLI account | `tx dex create-pool 10000000naet "10000000$DENOM" --fees 1000000naet` | creator must hold both pool coins plus `naet` fee | user coins move to `dex` module account; LP `lp/<pool_id>` minted to creator; pool, pair index, next pool id written | `dex_create_pool` plus SDK tx/message, bank transfer/mint events | `query dex pool 1`; `query bank balance $CREATOR lp/1`; module account balance checks in tests | `x/dex/keeper/msg_server_test.go`, `x/dex/keeper/query_server_test.go`, `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| DEX add liquidity | LP provider adds both pool assets | `--from <depositor>`; `MsgAddLiquidity.Depositor` from CLI account | `tx dex add-liquidity 1 1000000naet "1000000$DENOM" 1000000 --fees 1000000naet` | depositor must hold both deposit coins plus fee; `min_shares` bounds slippage | user coins move to module; LP shares minted; pool reserves and total shares increase | `dex_add_liquidity` plus SDK tx/message, bank transfer/mint events | `query dex pool 1`; `query bank balance $DEPOSITOR lp/1` | `x/dex/keeper/msg_server_test.go`, `tests/e2e/dex_smoke.ps1` |
| DEX remove liquidity | LP holder burns LP shares for underlying assets | `--from <withdrawer>`; `MsgRemoveLiquidity.Withdrawer` from CLI account | `tx dex remove-liquidity 1 1000000lp/1 --fees 1000000naet` | withdrawer must hold LP shares plus fee | LP shares move to module and burn; module sends pool assets; pool reserves and total shares decrease | `dex_remove_liquidity` plus SDK tx/message, bank transfer/burn events | `query dex pool 1`; `query bank balance $WITHDRAWER lp/1`; `query bank balance $WITHDRAWER $DENOM` | `x/dex/keeper/msg_server_test.go`, `tests/e2e/dex_smoke.ps1` |
| DEX swap exact in | trader swaps one pool asset for the other | `--from <trader>`; `MsgSwapExactAmountIn.Trader` from CLI account | `tx dex swap-exact-in 1 100000naet $DENOM 1 --fees 1000000naet` | trader must hold input coin plus fee; `min_amount_out` bounds slippage | input coin moves to module; output coin moves to trader; pool reserves update | `dex_swap_exact_amount_in` plus SDK tx/message and bank transfer events | `query dex pool 1`; `query bank balance $TRADER $DENOM`; `query bank balance $TRADER naet` | `x/dex/keeper/msg_server_test.go`, `x/dex/keeper/math_test.go`, `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |

## Negative And Adversarial Matrix

| Tx | Signer/auth failures | Field and denom failures | Balance/state failures | Replay/sequence coverage | Scale/scan note |
| --- | --- | --- | --- | --- | --- |
| Bank send | SDK rejects missing/wrong sender signature | invalid Bech32 receiver; invalid amount denom; wrong fee denom rejected by `x/fees` ante | insufficient funds fails in bank keeper without receiver mutation | SDK ante sequence; replay explicitly covered by `tests/integration/tx_lifecycle_test.go` and inherited by all SDK txs | direct account balance updates |
| Direct staking delegate rejected | SDK requires delegator signature, then pool-only policy rejects user delegation | validator address and denom do not unlock the direct path | no delegation, unbonding, redelegation, bonded-pool movement, or validator power mutation | `app/pos_test.go` and `tests/integration/pos_lifecycle_test.go` cover msg route and signed tx rejection | bounded message check plus staking key lookup; no full user scan |
| Tokenfactory create denom | creator signer from tx `--from`; malformed creator rejected | invalid subdenom, duplicate denom, native/LP spoofing | duplicate denom state rejected before metadata write | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | denom key lookup; no denom list scan |
| Tokenfactory mint | non-admin sender rejected | missing denom, invalid amount, invalid recipient address | bank mint/send errors returned; supply checked by bank query | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | denom metadata lookup only |
| Tokenfactory burn | non-admin and burn-from mismatch rejected | missing denom, invalid amount, invalid burn address | insufficient balance fails on send-to-module; bank supply decreases only after successful transfer | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | denom metadata lookup only |
| Tokenfactory change admin | non-admin sender rejected | missing denom, invalid new admin | metadata write only after checks | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | denom metadata lookup only |
| Fees update params | invalid authority rejected | empty, duplicate, multi-denom, invalid denom params rejected | params write only after `Validate` | governance/SKD ante sequence inherited; no operator shortcut | single params key |
| DEX create pool | creator signer from tx `--from`; malformed creator rejected | invalid coins, same denom, wrong/uncanonical denom pair | duplicate pair rejected before funds move; insufficient funds returned by bank keeper | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | pair-index lookup; no pool scan |
| DEX add liquidity | depositor signer from tx `--from`; malformed depositor rejected | missing pool, wrong pool denoms, non-positive tokens, invalid `min_shares` | excessive `min_shares`, corrupted pool state, insufficient funds fail before final pool write | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | pool id lookup |
| DEX remove liquidity | withdrawer signer from tx `--from`; malformed withdrawer rejected | missing pool, wrong LP denom, zero shares | shares exceed supply, withdrawal rounds to zero, insufficient LP balance fail safely | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | pool id lookup |
| DEX swap exact in | trader signer from tx `--from`; malformed trader rejected | missing pool, wrong in/out denoms, invalid input, invalid `min_amount_out` | excessive slippage, tiny output, corrupted pool, insufficient input balance fail safely | SDK ante inherited; replay/sequence covered at signed tx layer by `tests/integration/tx_lifecycle_test.go` | pool id lookup |

## Coverage Index

| Layer | Evidence |
| --- | --- |
| CLI construction | `cmd/l1d/cmd/root_test.go`, `x/tokenfactory/client/cli/tx.go`, `x/dex/client/cli/tx.go`, `x/fees/client/cli/query.go` |
| Msg server authorization and state writes | `x/tokenfactory/keeper/msg_server_test.go`, `x/dex/keeper/msg_server_test.go`, `x/fees/keeper/msg_server_test.go`, `app/pos_test.go` |
| Fee ante policy | `x/fees/keeper/ante_test.go`, `tests/integration/tx_lifecycle_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| Query verification | `x/tokenfactory/keeper/query_server_test.go`, `x/dex/keeper/query_server_test.go`, `x/fees/keeper/query_server_test.go`, `tests/e2e/query_surface_smoke.ps1` |
| E2E lifecycle | `tests/e2e/native_token_smoke.ps1`, `tests/e2e/pos_smoke.ps1`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Determinism and audit | `scripts/security/determinism-gate.ps1`, `scripts/security/prototype-audit.ps1`, `docs/security/cosmos-security-checklist.md` |
| Bench/perf | `BenchmarkEmptyBlockFinalizeCommit`, `BenchmarkDexCreatePoolsAndSwap` |

## Gaps

MUST FIX before public release:

- Add a reusable signed-tx replay/sequence e2e helper that can submit the same signed bytes twice and prove the second delivery cannot mutate state.

SHOULD FIX for stronger operator observability:

- Add high-cardinality query pagination load evidence for tokenfactory and DEX before public explorer/API load testing.
- Add per-row transcript artifacts to release evidence so each lifecycle can be audited without rerunning the localnet.
- Add targeted CLI negative tests for malformed Bech32 and malformed coin arguments at command construction boundaries where Cosmos SDK validation does not already cover the path.

NICE TO HAVE:

- Cross-module invariant generator that runs tokenfactory mint/burn and DEX pool operations across many denoms/accounts.
- Multi-OS localnet lifecycle transcript beyond the current Windows-focused operator evidence.
