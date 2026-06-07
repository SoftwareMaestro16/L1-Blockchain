# Official Liquid Staking UX

Normal user staking in Aetra is pool/index based:

```text
User -> Liquid Staking Contract -> Pool Contract -> Validators
```

The wallet, CLI, and user-facing API should present `deposit to official liquid staking pool` as the default staking action. The user supplies an `AE...` account address, an official pool id or contract, and an AET amount. The user does not supply or choose a validator address on the normal staking path.

Behavior:

- deposits below the governance-configured pool minimum are rejected;
- accepted deposits mint deterministic pool shares or a receipt token amount;
- the receipt represents a claim on pool assets and rewards, not ownership of any validator;
- the pool aggregates user deposits and sends validator-sized allocations through official pool accounting rules;
- validator rewards and slashing exposure return to the pool and are applied by pool share;
- direct user delegation to validators is disabled by default and exists only as an explicit governance-enabled operator path.

Example:

```text
user deposits 100 AET into official pool
pool total becomes 10,000 AET
allocation engine assigns deterministic validator weights
official pool contract injects pooled stake to validators
user reward share = user_pool_shares / total_pool_shares
```

Addresses:

- user-facing account, pool, and validator addresses are always `AE...`;
- raw/internal addresses are `4:...`;
- `aevaloper` and `aevalcons` are not used in user-facing staking APIs.
