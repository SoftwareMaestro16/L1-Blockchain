# Aetheris Core Roadmap

`ROADMAP.md` is a local operator planning file and must not be committed. It is intentionally listed in `.gitignore`.

## Git Policy

- Work is pushed directly with `git push` only when the operator explicitly asks for push.
- Do not create pull requests.
- Do not use GitHub compare/PR screens as the delivery path.
- Keep release/security branches only when they are explicitly requested; otherwise work on the requested branch and push it directly.

## Implementation Principle

This roadmap describes Aetheris architecture only. External ecosystems may be useful during private research, but the public plan should define Aetheris-native algorithms, standards, storage, messages, validation, tests, and rollout gates.

Core principles:

- Contract-owned state.
- Message-based execution.
- Explicit wallet, token, NFT, and SBT standards.
- Replay-safe contract wallets.
- Deterministic async queues.
- Native-only base chain fees in `naet`.

## Current Baseline

- Codebase is Cosmos SDK based.
- Core modules include `auth`, `bank`, `staking`, `distribution`, `gov`, custom `fees`, custom `tokenfactory`, custom `dex`, observability, localnet scripts, and gated CosmWasm readiness.
- The daemon is being renamed to `aetherisd`.
- Native coin is being renamed to Aetheris:
  - token name: `Aetheris`
  - ticker/display denom: `AET`
  - base/staking/fee denom: `naet`
  - exponent: `9`
  - supply model: uncapped PoS inflation/rewards, no fixed max supply.
- Address policy target:
  - raw: `4:` plus 64 lowercase hex characters, total length 66.
  - userfriendly: 48 chars, base64url alphabet `A-Z a-z 0-9 - _`, starts with `AE`.
  - zero address: `4:0000000000000000000000000000000000000000000000000000000000000000`.
  - old public `0:`, `ORB`, and `orb1` formats must be rejected outside explicit migration tooling.

## Architecture Direction

Aetheris should move toward a contract-first L1:

- Native chain state remains deterministic and block-based.
- User-visible entities become smart-contract entities where practical:
  - wallet contract
  - token master contract
  - token wallet contract
  - NFT collection contract
  - NFT item contract
  - SBT item contract
  - DEX pool/router contracts after the base chain is hardened
  - governance-controlled system contracts where suitable
- Protocol fees are always paid in native `naet`.
- User-created tokens can exist, trade, and be transferred, but cannot pay base chain transaction fees.
- Async execution is a major design track. Because Cosmos SDK delivery is synchronous today, implement this in stages:
  - Stage A: deterministic async message queue inside one chain.
  - Stage B: contract outbox/inbox, bounce, refund, and continuation semantics.
  - Stage C: benchmark and harden queue processing.
  - Stage D: only then evaluate partitioning/sharding-like architecture. Do not claim production sharding until Aetheris has its own consensus-safe design.

## Phase 1: Core Rename, Token, And Address Cleanup

Tasks:

- Finish public rename from Orbitalis to Aetheris across app, CLI, scripts, docs, workflows, release packaging, tests, and security docs.
- Rename user-facing daemon/docs/scripts from `orbitalisd` to `aetherisd`.
- Rename default home from `.orbitalis` to `.aetheris`.
- Rename local chain id from `orbitalis-local-1` to `aetheris-local-1`.
- Keep Go module path unchanged unless repository/module rename is separately requested.
- Replace `norb` with `naet` in staking, fees, mint, bank metadata, genesis, localnet scripts, e2e tests, docs, and workflows.
- Replace display `ORB` with `AET`.
- Keep native metadata canonical:
  - base denom: `naet`
  - display: `AET`
  - symbol: `AET`
  - exponent: `9`
- Enforce raw address output as `4:` plus lowercase hex.
- Enforce userfriendly output as `AE...` base64url.
- Reject old public `0:`, `ORB`, and `orb1` formats in normal validation/output.
- Keep any old-format support only in explicit migration tools, not in public protocol validation.

Acceptance:

- `go test ./...` passes.
- `go vet ./...` passes.
- Address codec tests prove valid `4:` and `AE...` pass.
- Negative tests prove old `0:`, `ORB...`, and `orb1...` fail.
- Localnet starts with `aetherisd`, `aetheris-local-1`, and `naet`.

## Phase 2: Refactor Boundaries

Goal: stop mixing product logic, test helpers, scripts, and documentation assumptions.

Tasks:

- Split protocol functionality from test helpers:
  - production modules stay under `app`, `x`, `cmd`;
  - integration fixtures stay under `tests`;
  - script helper libraries stay under `scripts/localnet/lib`;
  - release/docs smoke data stays under `docs` or `.work`.
- Separate functional modules from standards:
  - native coin policy in `app/params`;
  - address policy in `app/addressing`;
  - token standard specs in a future `x/contracts/standards` or `x/aetherisvm/standards`;
  - module-specific keeper logic remains in each `x/<module>`.
- Keep test-only denoms such as `testtoken` out of native token metadata and fee policy.
- Add shared validation helpers for:
  - user address
  - authority address
  - optional admin address
  - contract address
  - zero address rejection
  - fee denom validation
- Add a test matrix file that maps each module to:
  - unit tests
  - keeper tests
  - integration tests
  - e2e smoke tests
  - adversarial tests

Acceptance:

- It is clear which code is production protocol logic and which code is test/script support.
- Native AET policy is not duplicated through hardcoded strings.
- Module tests import shared helpers instead of reimplementing address/denom checks.

## Phase 3: Zero Address And Address Safety

Policy:

- Zero address is forbidden by default.
- Zero address is not a normal burn sink.
- If a burn sink is ever required, define an explicit protocol sink account/contract with no signer semantics and separate validation.

Tasks:

- Reject zero address as signer.
- Reject zero address as token admin.
- Reject zero address as mint recipient.
- Reject zero address as burn-from address.
- Reject zero address as DEX liquidity provider, swap trader, or swap recipient.
- Reject zero address in fees params, authority params, genesis accounts, and module genesis state.
- Reject empty, malformed, mixed-case raw, wrong-length raw, wrong-prefix userfriendly, and non-base64url userfriendly addresses.
- Keep regression tests proving valid `4:` and `AE...` addresses still pass.

Acceptance:

- Zero address cannot become signer, admin, receiver, authority, fee collector, genesis account, DEX participant, or token holder unless a future explicit sink rule is added.

## Phase 4: Account And Transaction Safety

Tasks:

- Audit account creation paths.
- Verify sequence/nonce replay protection.
- Add replay-like transaction tests.
- Add invalid signer tests for bank, staking, gov, fees, tokenfactory, DEX, and future contract VM messages.
- Add malformed protobuf transaction tests.
- Add insufficient funds tests.
- Add wrong chain-id signing tests.
- Add missing fee, malformed fee, wrong fee denom, and below-min fee tests.
- Document and test ante handler order.
- Add consensus panic tests for malformed custom module inputs.
- Add event consistency tests for important state transitions.

Acceptance:

- Same signed tx cannot execute twice.
- Invalid signer cannot mutate state.
- Bad tx fails before state transition.
- Error paths are deterministic and do not panic the app.

## Phase 5: Fee Model Hardening

Policy:

- Base chain fees are paid only in `naet`.
- User-created tokens, token wallets, DEX LP tokens, NFT assets, SBT assets, and test tokens cannot pay protocol fees.
- Gasless/user-friendly flows may exist only through relayers that pay `naet` on-chain and optionally collect other tokens off-chain or inside a separate contract flow.

Tasks:

- Keep `allowed_fee_denoms = ["naet"]`.
- Reject non-`naet` fees even if the user owns the token.
- Bound fee params.
- Add governance authority tests for fee params.
- Add fee collector/module account balance tests.
- Add fee deduction integration tests.
- Add fee split tests if custom distribution remains.
- Decide localnet/public-testnet zero-fee policy:
  - localnet may allow validator minimum gas price `0naet`;
  - delivered tx policy should still require deterministic protocol fees unless explicitly disabled for dev-only chains.
- Document fee accounting in README and security docs.

Acceptance:

- Fees are deterministic, native-only, and cannot be bypassed by token spoofing or malformed txs.

## Phase 6: PoS And Staking Correctness

Tasks:

- Test validator creation with `naet`.
- Test delegation, unbonding, redelegation, validator power updates, and CometBFT propagation.
- Test slashing params and downtime signing info.
- Test staking rewards and distribution.
- Test validator set updates after delegation changes.
- Test 3-validator and 5-validator localnet genesis.
- Test restart persistence after staking state changes.
- Document uncapped PoS supply:
  - AET has no fixed max supply;
  - `naet` issuance comes from configured inflation/reward policy;
  - reward policy must be deterministic and governance-bounded.

Acceptance:

- Local 3-validator network produces blocks.
- Validator power updates after staking tx.
- Restart preserves staking state and consensus progress.

## Phase 7: User Token Standard

Goal: support user-created fungible tokens through an Aetheris-native master/wallet contract model, not as fee denoms.

Working standard name: Aetheris Fungible Token Standard, `AFT-1`.

Contract model:

- `token_master` is the root contract for one token.
- `token_wallet` is a separate contract/account for each holder of that token.
- Native AET is not an `AFT-1` token. AET remains the base chain coin and fee/staking denom.

`token_master` storage:

- token id/address
- owner/admin address
- total supply
- mintable flag
- burnable flag
- metadata/content reference
- token name
- token symbol
- decimals
- token wallet code hash or code reference
- admin transfer/renounce state
- optional allowlist/denylist policy if added later

`token_master` messages:

- mint
- burn notification from token wallet
- change admin
- renounce admin
- change metadata
- close minting
- get supply data
- get wallet address for owner
- upgrade wallet code only if explicitly allowed by policy

`token_wallet` storage:

- master address
- owner address
- balance
- wallet code/version
- pending inbound/outbound query ids if needed for async safety

`token_wallet` messages:

- transfer
- internal transfer
- receive transfer
- burn
- notify owner
- send excess/refund
- bounce handling for failed sends

Rules:

- Wallet contract address should be derived from `(token_master, owner, token_wallet_code_hash)` so wallets are predictable.
- Mint deploys or credits the recipient token wallet.
- Transfer moves balance from sender token wallet to recipient token wallet through deterministic internal messages.
- Burn decreases wallet balance first, then notifies master to decrement total supply.
- Master supply and wallet balances must not diverge.
- Token metadata must not spoof native AET metadata.
- Token balances are never accepted as protocol fee payment.

Tests:

- create token master
- mint to first holder
- derive holder wallet address
- transfer between holders
- deploy missing recipient wallet
- burn and supply decrement
- admin transfer and renounce
- non-admin mint rejected
- native `AET`/`naet` spoof rejected
- non-`naet` fee rejected for all token operations
- bounce restores or finalizes state deterministically
- malformed message rejected
- replayed wallet message rejected

Acceptance:

- Any user can create a token such as a USDT-like asset through the token standard.
- Each holder has an independent token wallet contract.
- Only `naet` pays protocol fees.
- Supply accounting is deterministic under async execution.

## Phase 8: NFT And SBT Standards

Goal: support NFT collections, NFT items, and native non-transferable SBT items as contract standards.

Working standard names:

- Aetheris NFT Standard, `ANFT-1`.
- Aetheris Soulbound Token Standard, `ASBT-1`.

`nft_collection` storage:

- collection owner/admin
- collection metadata/content
- next item index
- item code hash or code reference
- standard version
- mutable metadata flag
- collection royalty policy if added
- SBT policy flag if collection is soulbound-only

`nft_collection` messages:

- mint item
- batch mint with bounded count
- change collection metadata
- change owner/admin
- get collection data
- get item address by index
- prove item belongs to collection

`nft_item` storage:

- collection address
- item index
- owner address
- item metadata/content
- initialized flag
- transfer policy

`nft_item` messages:

- transfer
- ownership notification
- excess/refund
- get item data
- bounce handling where needed

`sbt_item` extension storage:

- collection address
- item index
- immutable owner address
- content
- authority address
- revoked_at timestamp or zero
- revoke reason/content if needed

`sbt_item` messages:

- prove ownership
- ownership proof response
- request current owner
- revoke by authority
- destroy only if policy allows it
- transfer must be rejected

Rules:

- Collection is source of truth for item address derivation.
- Item must prove collection membership through collection-derived address.
- NFT transfer requires current owner authorization.
- SBT owner is immutable after mint.
- SBT revoke does not transfer ownership.
- Metadata changes must be bounded and authorization-checked.
- Batch minting must have strict limits to prevent block/queue DoS.

Tests:

- create collection
- mint NFT item
- verify item address from collection
- transfer NFT item
- unauthorized transfer rejected
- malformed collection address rejected
- SBT mint
- SBT transfer rejected
- SBT prove ownership
- SBT revoke by authority
- unauthorized revoke rejected
- metadata spoofing rejected
- batch mint bounded

Acceptance:

- NFT and SBT assets are first-class contract entities.
- Collection/item verification is deterministic.
- SBTs cannot be transferred.

## Phase 9: Wallet Contract Standard

Goal: support replay-safe Aetheris contract wallets while keeping consensus rules deterministic.

Working standard name: Aetheris Wallet Standard, `AW-1`.

Wallet storage:

- signature_allowed flag
- seqno
- wallet_id
- public_key
- owner address if separate from public key
- extensions dictionary/map
- optional recovery authority policy
- optional subscription/standing-payment policy

Wallet messages:

- external signed command
- internal extension command
- update signature_allowed
- install extension
- remove extension
- multi-send message batch
- query wallet state

Rules:

- `seqno` prevents replay.
- `wallet_id` prevents cross-wallet replay.
- `valid_until` prevents old signed commands.
- Signature validation happens before gas acceptance for external commands.
- Extension auth is explicit and revocable.
- Multi-send is bounded.
- Wallet cannot silently pay protocol fees in non-`naet`.
- A relayer can submit a command, but on-chain fee payment is still `naet`.

Tests:

- deploy wallet
- signed send
- replayed seqno rejected
- wrong wallet_id rejected
- expired valid_until rejected
- invalid signature rejected
- extension install/remove
- extension authorized send
- unauthorized extension rejected
- multi-send bounded
- relayer flow pays `naet`

Acceptance:

- Contract wallets can safely authorize transactions and contract messages.
- Replay and wrong-wallet attacks fail before state mutation.

## Phase 10: Async Smart Contract Execution

Goal: bring Aetheris-native asynchronous contract message semantics into the chain without breaking Cosmos SDK determinism.

Important constraint:

- Current Cosmos SDK transaction execution is synchronous per delivered tx.
- Aetheris can add deterministic async contract semantics inside blocks, but production partitioning/sharding requires a deeper consensus architecture and must be treated as a later R&D track.

Core concepts:

- contract account
- contract state
- incoming message
- outgoing message
- message queue
- bounce message
- refund/excess message
- logical time per contract
- bounded per-block queue processing
- deterministic ordering
- gas/fee accounting per message
- failure result code

Tasks:

- Design contract address derivation.
- Design contract state storage.
- Design message envelope:
  - source
  - destination
  - value in `naet`
  - opcode
  - query_id
  - body
  - bounce flag
  - created logical time
  - expiration/deadline if needed
- Design outbox/inbox storage.
- Design queue ordering.
- Design bounce behavior:
  - when to bounce
  - what state rolls back
  - what state remains final
  - how excess/refunds are returned
- Design fee/gas model:
  - execution gas
  - storage fee if used
  - message forwarding fee
  - contract deployment cost
  - all protocol fees in `naet`
- Add deterministic processing limits:
  - max messages per tx
  - max messages per block
  - max recursion/depth
  - max body size
  - max state size
  - max contract deploys per tx/block
- Add observability:
  - queued messages
  - processed messages
  - bounced messages
  - failed executions
  - gas used
  - queue lag

Acceptance:

- A contract can emit an internal message to another contract.
- The recipient executes in deterministic order.
- Failed sends produce deterministic bounce/refund behavior.
- Queue limits prevent DoS.
- State export/import preserves queue state exactly.

## Phase 11: VM Direction

Near-term:

- Keep CosmWasm as the gated VM candidate already reflected in `app/wasmconfig`.
- Keep it disabled by default.
- Enable only with explicit config/feature gate.
- Define upload permissions, instantiate permissions, admin/migration policy, gas limits, contract size limits, memory/cache limits, and query limits.

Aetheris VM R&D:

- Define whether Aetheris contracts are:
  - CosmWasm contracts with Aetheris async/message standards; or
  - a future Aetheris VM with its own execution, storage, and message ABI; or
  - both, with strict compatibility boundaries.
- Do not introduce a new VM without:
  - binary serialization spec
  - message ABI
  - storage ABI
  - gas schedule
  - deterministic execution proof
  - fuzz tests
  - upgrade/migration policy
  - adversarial audit

Contract standard requirements:

- Explicit storage schema.
- Explicit inbound messages.
- Explicit outbound messages.
- Explicit getters.
- Explicit unknown-message policy.
- Explicit bounce behavior.
- Explicit fee behavior.
- Explicit deployment behavior.

Acceptance:

- CosmWasm readiness does not weaken base chain security.
- Aetheris async VM research has a written spec before implementation.
- Contract standards can be tested independent of the VM choice.

## Phase 12: DEX Direction

Current:

- `x/dex` is part of the blockchain and should remain in the repository.
- It can stay as a native module while the contract VM matures.

Future:

- Move toward contract-based pools/routers when async contract execution is safe.
- Keep native DEX module as reference or migration bridge until contract DEX is audited.

Tasks:

- Keep current DEX tests on `naet`.
- Harden zero-address checks for pool creator, liquidity provider, withdrawer, trader, and recipient.
- Reject native AET spoofing through LP/factory denoms.
- Add invariant tests:
  - reserves match module balances
  - LP supply matches pool shares
  - swaps preserve constant-product constraints within fee policy
  - slippage bounds are enforced
- Add future contract DEX design:
  - pool contract
  - router contract
  - LP token master/wallet or native LP representation
  - async swap settlement
  - bounce/refund on failed swap path

Acceptance:

- Current DEX is safe as native module.
- Future DEX migration path is contract-compatible.

## Phase 13: Genesis, Export, Migration Safety

Tasks:

- Validate genesis accounts, balances, staking denom, mint denom, fee denom, module params, token masters, token wallets, NFT collections/items, SBT items, and async queues.
- Reject duplicate accounts and duplicate contract addresses.
- Reject malformed account/contract state.
- Reject zero addresses in genesis.
- Add export/import round-trip tests.
- Add deterministic exported state tests.
- Add migration skeletons for custom modules.
- Add upgrade handler discipline docs.
- Add migration-only support for old Orbitalis/ORB/orb1/0: data if needed, clearly isolated from public validation.

Acceptance:

- `DefaultGenesis -> InitGenesis -> ExportGenesis -> ValidateGenesis` is deterministic.
- Malformed input is rejected without avoidable panics.
- Upgrade/migration paths cannot silently reintroduce old public formats.

## Phase 14: Security Audit Pack

Tasks:

- Keep manual audit checklist.
- Keep `govulncheck` workflow.
- Keep `gosec` workflow.
- Keep CodeQL workflow.
- Keep gitleaks secret scanning.
- Keep dependency review.
- Add triage policy for high/critical findings.
- Add Cosmos-specific checks:
  - non-determinism
  - incorrect signers
  - ABCI panics
  - unsafe rounding
  - unbounded iteration
  - malformed genesis
  - replay paths
  - invalid authority paths
- Add contract-specific checks:
  - wallet replay
  - wrong wallet_id
  - extension takeover
  - token supply divergence
  - NFT unauthorized transfer
  - SBT transfer bypass
  - async queue DoS
  - bounce/refund double-spend
  - metadata spoofing
  - admin takeover

Acceptance:

- Public testnet cannot proceed with untriaged high/critical fund-safety, consensus-safety, or secret-leak findings.

## Phase 15: Performance And Speed

Policy:

- Optimize only with benchmarks or smoke-duration data.
- Do not weaken validation, determinism, or auditability for speed.

Benchmark targets:

- ante handler hot paths
- address parsing
- fee validation
- genesis validation
- staking tx flow
- token master operations
- token wallet transfer
- NFT mint/transfer
- SBT proof/revoke
- DEX keeper paths
- async queue processing
- contract state load/save
- localnet startup
- CI smoke duration

Tasks:

- Add Go benchmarks for address parsing and fee validation.
- Add keeper benchmarks for DEX/tokenfactory.
- Add async queue benchmarks once queue exists.
- Add localnet startup timing output.
- Add CI timing summary.
- Identify repeated JSON/protobuf parsing in scripts and tests.
- Bound all loops over user-controlled state.
- Add pagination and query limits for token/NFT/contract indexes.

Acceptance:

- Performance changes have before/after numbers.
- No optimization bypasses signer, fee, denom, zero-address, or genesis validation.

## Phase 16: Public Testnet Preparation

Tasks:

- Harden localnet scripts.
- Add 3-validator and 5-validator profiles.
- Add faucet plan.
- Add explorer/indexer plan.
- Add validator onboarding docs.
- Add minimum hardware docs.
- Add snapshot/state-sync plan.
- Add incident response docs.
- Add launch checklist.
- Add rollback/restart procedure.
- Run full smoke test before public testnet.

Acceptance:

- A validator can join from clean docs, sync, produce/validate blocks, and use enabled modules safely.
- If CosmWasm or async contracts are enabled, a simple contract deployment and token/NFT smoke test must pass first.

## Immediate Next Work

1. Finish tracked Aetheris rename sweep in docs, scripts, workflows, release packaging, and tests.
2. Keep `ROADMAP.md` ignored and local-only.
3. Re-run:
   - `go test ./...`
   - `go vet ./...`
   - `buf lint`
4. Start Phase 2 refactor helpers:
   - central native token constants
   - central address validation
   - central fee denom validation
5. Add the contract standards spec files before implementing VM behavior:
   - wallet standard
   - fungible token standard
   - NFT standard
   - SBT standard
   - async message model
6. Only then implement contract runtime changes.
