> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Mempool And CheckTx Negative Flow

Malformed or underfunded transactions must fail without panic and without unintended state writes. This gate documents the expected rejection phase and state invariant for common prototype failures.

Run:

```powershell
.\tests\e2e\mempool_negative_smoke.ps1
```

The smoke uses `.localnet-mempool-negative` and shifted ports by default, creates one tokenfactory denom and one DEX pool, then submits invalid transactions over normal CLI/broadcast paths.

## Negative Matrix

| Scenario | Expected phase | Expected error signal | State invariant |
| --- | --- | --- | --- |
| Wrong fee denom on bank send | `CheckTx` / ante | `fee denom testtoken not accepted; use naet` | receiver balance unchanged; no tx execution |
| Insufficient bank send amount | `DeliverTx` | `insufficient`, `spendable`, or `funds` | receiver balance unchanged; SDK fee deduction for failed DeliverTx is allowed |
| Signed tx replay / invalid sequence | `CheckTx` / ante | `sequence`, `account sequence`, or signature verification failure | second broadcast does not repeat bank send |
| Unauthorized tokenfactory mint | `DeliverTx` | `only denom admin can mint` | factory token supply and recipient balance unchanged |
| Malformed tokenfactory subdenom | CLI validation or `DeliverTx` | `subdenom`, `invalid`, or length rule text | tokenfactory denom count unchanged |
| Duplicate DEX pool | `DeliverTx` | `pool already exists` | pool count and existing pool state unchanged |
| Malformed DEX denom | CLI validation or `DeliverTx` | invalid coin/denom text | pool count unchanged |

## Security Notes

- `CheckTx` failures are cheap ante or CLI construction failures and must not execute module handlers.
- `DeliverTx` failures may pay normal SDK fees, but custom module state and target balances must remain unchanged.
- Rejection paths rely on bounded parsing, direct key lookups, and cached state changes in tokenfactory/DEX handlers.
- Errors may identify the rejected field or policy, but must not print local paths, key material, mnemonics, or private validator data.

## Related Tests

- Fee shape and denom ante coverage: `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1`.
- Tokenfactory authorization/state coverage: `x/tokenfactory/keeper/msg_server_test.go`.
- DEX insufficient-funds, duplicate, slippage, and malformed pool coverage: `x/dex/keeper/msg_server_test.go`, `tests/e2e/dex_smoke.ps1`.
- Staking wrong denom, insufficient funds, malformed validator, and replay coverage: `app/pos_test.go`, `tests/e2e/pos_smoke.ps1`.
