> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Module Account and Bank Movement Audit

Purpose: every custom module mint, burn, send, and reserve custody path must have explicit ownership, validation, error handling, and regression coverage.

## Module Accounts

| Module account | Permissions | Used for | Blocked recipient |
| --- | --- | --- | --- |
| `tokenfactory` | `Minter`, `Burner` | Temporary custody while minting to recipients and burning admin-owned factory tokens. | Yes |
| `dex` | `Minter`, `Burner` | Pool reserve custody and LP mint/burn operations. | Yes |
| `fees` | none | Fee policy validation only; no custom bank movement. | Yes |

The app-level bank keeper receives `BlockedAddresses()` for all module accounts except governance. Custom modules must not treat module accounts as normal user recipients.

## Atomicity Rule

Multi-step custom bank movements run inside a handler-local cached context via `x/internal/tx.AtomicStateChange`. The cache is written only after every bank movement, custom state write, and event emission succeeds. This matches BaseApp rollback expectations and also protects direct keeper tests from partial bank writes on late failure.

Validation must happen before the cached mutation block when possible. Custom state writes happen after required bank movements inside the same cached block.

## Inventory

| Function | Module | Movement | Denom/amount validation | State ordering | Error handling and rollback | Tests |
| --- | --- | --- | --- | --- | --- | --- |
| `CreateDenom` | tokenfactory | No coin movement; writes denom authority metadata and bank metadata. | Creator Bech32, subdenom format, duplicate denom, native spoofing. | Denom metadata is written after validation. | Store errors propagate; no supply mutation. | `TestCreateDenomRejectsNativeTokenSpoofing`, metadata tests. |
| `Mint` | tokenfactory | `MintCoins(tokenfactory, amount)` then `SendCoinsFromModuleToAccount(tokenfactory, mint_to, amount)`. | Existing denom, current admin signer, recipient Bech32, positive valid coin. | No tokenfactory state write. Event emitted in cached block after bank success. | All bank errors checked; cached block prevents minted supply leak if recipient send fails. | `TestTokenfactoryCreateMintBurnAdminFlow`, `TestMintToBlockedModuleAddressDoesNotLeakSupply`, unauthorized admin tests. |
| `Burn` | tokenfactory | `SendCoinsFromAccountToModule(burn_from, tokenfactory, amount)` then `BurnCoins(tokenfactory, amount)`. | Existing denom, current admin signer, `burn_from == sender`, positive valid coin. | No tokenfactory state write. Event emitted in cached block after bank success. | All bank errors checked; cached block prevents token custody changes if burn fails. | `TestAdminCanBurnOwnFactoryTokens`, unauthorized burn tests. |
| `ChangeAdmin` | tokenfactory | No coin movement. | Existing denom, current admin signer, new admin Bech32. | Admin metadata written after validation. | Store errors propagate; no supply mutation. | `TestAdminTransferRejectsOldAdminAndPreservesSupply`. |
| `CreatePool` | DEX | User sends both pool coins to `dex`; `dex` mints LP; `dex` sends LP to creator. | Creator Bech32, canonical distinct positive coins, duplicate pair check, native-spoof rejection, positive initial shares. | Pool, pair index, and next id write after all bank movements in cached block. | All bank/store errors checked; cached block prevents reserve custody or LP supply partial writes. | create-pool tests, duplicate pair tests, accounting invariant test. |
| `AddLiquidity` | DEX | User sends both pool coins to `dex`; `dex` mints LP; `dex` sends LP to depositor. | Existing pool, matching denoms, positive coins, valid pool state, non-negative `min_shares`, slippage bound. | Pool reserves and total shares write after bank movements in cached block. | All bank/store errors checked; insufficient funds leaves pool and LP balance unchanged. | `TestAddLiquidityInsufficientFundsLeavesPoolAndBalancesUnchanged`, lifecycle tests. |
| `RemoveLiquidity` | DEX | User sends LP to `dex`; `dex` burns LP; `dex` sends reserve coins to user. | Existing pool, LP denom match, positive shares, valid pool state, shares <= total shares, positive withdrawal amounts. | Pool reserves and total shares write after bank movements in cached block. | All bank/store errors checked; reserve/module mismatch does not burn user LP or update pool. | `TestRemoveLiquidityReserveMismatchDoesNotBurnSharesOrUpdatePool`, accounting invariant test. |
| `SwapExactAmountIn` | DEX | User sends input coin to `dex`; `dex` sends output coin to user. | Existing pool, positive valid input, pool denom match, valid pool state, non-negative `min_amount_out`, slippage bound. | Pool reserves write after bank movements in cached block. | All bank/store errors checked; failed bank movement leaves pool unchanged under cached block. | lifecycle/slippage tests, accounting invariant test. |
| Ante fee validation | fees | No custom bank movement; SDK ante handles fee deduction. | FeeTx shape, non-empty positive fee, single allowed denom `naet`, params validation. | No fees module state write in ante. | Wrong/malformed fee fails before tx execution. | fees ante unit and e2e tests. |

## Audit Result

- No unchecked custom bank movement errors remain in `x/tokenfactory` or `x/dex`.
- `tokenfactory` and `dex` are the only custom modules with mint/burn permissions.
- DEX reserve custody is the `dex` module account; LP supply must equal pool `total_shares`.
- `fees` has no module-account bank movement and no mint/burn permission.
- Critical/High accounting regressions require a test or a documented release blocker before merge.
