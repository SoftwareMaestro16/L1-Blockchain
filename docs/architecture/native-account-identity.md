# Native Account Identity

User-facing account, validator, and consensus addresses are always `AE...`.
The raw `4:...` format remains an internal/proof key format and is not accepted
by user-facing message or query validation.

Examples:

```text
account_address = AEAAAQAAAAAAAAAAAAAAAAAiIiIiIiIiIiIiIiIiIiIiIiIi
validator_address = AEAAAQAAAAAAAAAAAAAAAAAzMzMzMzMzMzMzMzMzMzMzMzMz
consensus_address = AEAAAQAAAAAAAAAAAAAAAAA0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0
raw_internal_key = 4:0000000000000000000000002222222222222222222222222222222222222222
```

Legacy `aevaloper` and `aevalcons`, foreign Bech32, old raw prefixes, mixed-case
raw addresses, and malformed `AE...` strings are rejected at user-facing
message and query boundaries.

Before activation, a derivable `AE...` address is a virtual inactive account.
Querying it returns an inactive non-persistent view. It is not exported in
genesis, does not accrue storage rent, and can only submit `MsgActivateAccount`.
`MsgActivateAccount` is the normal first persistent state creation path; a
controlled migration is the only other allowed persistent write reason.
