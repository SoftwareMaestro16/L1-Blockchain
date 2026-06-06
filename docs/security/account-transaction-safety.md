> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Account And Transaction Safety

This document records the Phase 4 Aetra transaction safety policy.

## Base Chain Safety Before Contracts

AVM and CosmWasm must not be allowed to mutate production state until the base
chain safety gate is green. The gate exists so contract runtime work cannot
weaken signer, address, denom, fee, genesis, or deterministic transaction
invariants.

Required centralized helpers:

- address validation lives in `app/addressing`;
- `ParseUserAddress`, `ParseAuthorityAddress`, `ParseOptionalAdminAddress`,
  `ParseContractAddress`, and `RejectZeroAddress` are the shared entry points;
- native token constants live in `app/params`;
- `BaseDenom = "naet"` is the single source for staking, mint, and fee denom
  policy;
- `ValidateNativeFeeDenomsV1` is the shared native-only fee denom validator.

Required rejection policy:

- zero address is rejected everywhere by default;
- old public `0:`, `orb1`, and `ORB` formats are rejected outside explicit
  migration tooling;
- bad transactions fail before message state mutation;
- invalid signers cannot mutate state;
- malformed tx bytes must not panic consensus code;
- protocol fee accounting records only after the wrapped SDK ante handler
  succeeds.

Required evidence before AVM or CosmWasm state mutation is enabled:

- signed transaction replay test using identical signed bytes;
- wrong chain-id signing test;
- malformed protobuf transaction test;
- invalid signer tests for bank, staking, gov, fees, tokenfactory, DEX, AVM,
  and CosmWasm messages;
- consensus panic tests for every custom module message and genesis type;
- deterministic event contract tests for important state transitions;
- `go test ./...`;
- `go vet ./...`;
- `buf lint`;
- security scans: `govulncheck`, `gosec`, CodeQL, gitleaks, Dependency Review;
- determinism gate with no untriaged high/critical consensus or fund-safety
  findings.

Until AVM and CosmWasm message handlers exist, their required invalid-signer
and panic tests are explicit enablement blockers rather than skipped
production coverage.

## Account Creation Paths

Runtime account creation is handled through the Cosmos SDK account keeper and
standard ante flow. Aetra adds protocol guards around address parsing and
genesis import:

- Auth genesis accounts are rejected when the account address is the Aetra
  zero address.
- Bank genesis balances are rejected when the balance owner is the zero address.
- `SimGenesisAccount.Validate` rejects nil base accounts and zero addresses.
- Bank `MsgSend` and `MsgMultiSend` sender and recipient fields are checked by
  the Aetra ante wrapper before fee validation.
- Custom modules parse actor, authority, admin, contract, and fee collector
  fields with `app/addressing` helpers, so empty, malformed, wrong-prefix,
  mixed-case raw, non-base64url userfriendly, and zero addresses cannot become
  module actors.
- Zero address is not a burn sink. Any future sink must be an explicit protocol
  account or contract with separate validation.

## Replay And Signer Safety

All signed transactions run through the SDK ante handler after the Aetra fee
policy wrapper. SDK ante verifies account number, account sequence, chain ID in
the sign bytes, signer/public-key binding, signature count, signature gas, and
then increments sequence only after signature verification succeeds.

Acceptance evidence:

- Same signed bank tx succeeds once and replay fails with the stale sequence.
- A tx signed for a different chain ID fails before balance mutation.
- A tx signed by an account that is not the message signer fails before balance
  mutation.
- Malformed protobuf tx bytes produce failed tx results without fee accounting.
- Missing fee, malformed fee, wrong fee denom, below-min fee, and insufficient
  fee funds fail before message state transition.

## Ante Handler Order

`app/handlers.go` installs:

1. `x/fees` ante decorator.
2. Cosmos SDK `x/auth/ante.NewAnteHandler`.

The `x/fees` decorator runs first:

1. Reject zero address fee payers, signers, bank send senders/recipients, and
   bank multisend inputs/outputs.
2. At block height `0`, allow genesis `MsgCreateValidator` txs through.
3. Require the tx to implement `sdk.FeeTx`.
4. Load fee params.
5. Require valid native `naet` fees meeting `min_fee_amount`.
6. Call the wrapped SDK ante handler.
7. Only after wrapped ante success and non-simulation, record protocol fee
   accounting.

The SDK ante handler order in Cosmos SDK v0.54.3 is:

1. `SetUpContext`
2. `ExtensionOptions`
3. `ValidateBasic`
4. `TxTimeoutHeight`
5. `ValidateMemo`
6. `ConsumeGasForTxSize`
7. `DeductFee`
8. `SetPubKey`
9. `ValidateSigCount`
10. `SigGasConsume`
11. `SigVerification`
12. `IncrementSequence`

Because Aetra records protocol fee accounting only after the wrapped SDK ante
succeeds, missing fees, invalid fee denoms, insufficient fee funds, invalid
signatures, wrong chain ID, stale sequence, and zero-address actor fields cannot
update protocol fee accounting or message state.

## Required Tests

Run:

```powershell
go test ./tests/integration
go test ./tests/adversarial
go test ./x/fees/...
go test ./...
go vet ./...
buf lint
```

The public-testnet security gate additionally requires the configured security
workflows and `scripts/security/determinism-gate.ps1`.
