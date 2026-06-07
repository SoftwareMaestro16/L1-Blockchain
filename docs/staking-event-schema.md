# Deterministic Staking Events And Receipts

Staking event schema version: `staking-event-v1`.

Each staking event is a fixed-field record, not a map:

- `type`: one of `AccountActivated`, `PoolStakeDeposited`, `PoolSharesMinted`, `PoolAllocationUpdated`, `PoolUnbondingRequested`, `PoolUnbondingCompleted`, `PoolRewardsClaimed`, `StakeReputationClaimed`, `ValidatorRegistered`, `ValidatorUpdated`, `AdvancedStakeDelegated`, `AdvancedStakeUndelegated`, `AdvancedStakeRedelegated`.
- `actor`: user-facing `AE...` account that caused the state transition.
- `pool_contract`: user-facing `AE...` liquid staking contract when the event is pool-scoped.
- `validator`: user-facing `AE...` validator only for pool allocation, validator operator, or advanced staking paths.
- `amount`: base-unit AET amount or claimed reputation delta when applicable.
- `shares`: pool share amount when applicable.
- `height` and `epoch`: consensus height and reward/allocation epoch. Wall-clock time is not part of the schema.
- `sequence`: zero-based transaction-local event order.
- `state_key`: deterministic state key for the proved object.
- `proof_metadata_hash`: optional hash of the state proof metadata when a proof envelope is available.
- `event_hash`: deterministic hash over the fixed fields above.

Receipt schema version: `staking-receipt-v1`.

A staking receipt contains a normalized transaction id, a height, an ordered event list, and a receipt hash. Events are ordered by `height`, then `sequence`, then type/hash as deterministic tie breakers. Valid transaction receipts require contiguous zero-based event sequences.

The schema intentionally excludes private keys, seed phrases, mnemonics, auth secrets, wall-clock timestamps, and arbitrary attribute maps. Indexers should primarily index low-cardinality fields such as event type, height, epoch, actor, pool contract, and validator. `state_key` and proof hashes are payload fields for proof lookup and audit joins.
