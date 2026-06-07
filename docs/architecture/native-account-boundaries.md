# Native Account Boundaries

The native wallet/account model lands in `x/native-account`. It is a bounded
native module for account activation, account status, auth policy, account
sequence/account number ownership, account metadata bounds, account storage-rent
references, and account export/import state.

It must not replace Cosmos SDK `x/auth`, `x/bank`, `x/staking`, `x/slashing`,
or `x/distribution`. Those SDK modules remain wired and keep their invariants
until an explicit migration replaces them with compatibility tests.

## Ownership

| Area | Owner |
| --- | --- |
| `app/addressing` | `AE...` user-facing and `4:...` raw/internal formatting only; no state writes. |
| `x/native-account` | Versioned account state, activation, auth policy, status, sequence, account number, storage-rent references. |
| `x/identity` | `.aet` domains, resolver records, ownership indexes, and optional domain NFT binding proofs. |
| `x/reputation` | Deterministic account/contract reputation records and stake-time accumulators. |
| `x/storage-rent` | Rent params, state-size records, rent debt accounting, protocol-payer reserve policy. |
| `x/pos` | PoS calculation specs and compatibility checks around SDK staking/slashing. |
| `x/nominator-pool` / `x/single-nominator-pool` | Pool accounting boundaries; normal users do not choose validators directly. |
| `x/validator-*` | Validator registry, election, insurance, admission, and validator lifecycle metadata. |
| `x/stake-concentration` | Concentration snapshots, power-cap metrics, and anti-centralization signals. |
| `x/fees` / `x/burn` / `x/treasury` | Fee policy, burn accounting, and treasury allocation records using bank keeper movement. |
| `x/contracts`, `x/vm`, `x/aetravm/*` | Contract code/data, VM routing, contract standards, async messages, token/NFT/DEX contract state. |

## Guardrails

- Private keys and seed phrases are never account state.
- Balances stay in the bank/native balance layer.
- Token, NFT, and DEX behavior is contract-routed, not reintroduced as native
  asset modules.
- Cross-module writes into native account auth policy, sequence, status, or
  account number are rejected; future integrations use explicit interfaces.
- Storage rent may charge and report account/contract debt through interfaces,
  but must not automatically delete wallets or freeze protocol-critical state.
- Normal user staking remains routed through the official liquid staking flow,
  not direct user delegation to a chosen validator.
