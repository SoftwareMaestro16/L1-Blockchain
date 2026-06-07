> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Testnet Incident Response

Use this runbook for public testnet incidents. It does not replace chain governance or mainnet incident policy.

## Severity

Critical:

- chain halt or repeated consensus failure,
- AppHash divergence,
- widespread validator key compromise,
- exploit causing unauthorized mint, burn, fee bypass, admin takeover, or DEX accounting corruption.

High:

- many validators missing blocks,
- state sync or snapshot publication is wrong,
- faucet drains or distributes excessive funds,
- explorer/indexer shows misleading finalized state,
- public RPC is unavailable for most users.

Medium:

- single validator outage,
- delayed indexer,
- documentation issue that blocks onboarding but does not affect consensus.

## First Response

1. Freeze risky automation: faucet, indexer writes, deployment bots, and nonessential scripts.
2. Preserve evidence: node logs, tx hashes, block heights, genesis hash, app version, commit, and config diffs.
3. Identify whether the issue is consensus, networking, app logic, operator config, faucet, explorer, or documentation.
4. Communicate status with exact height, UTC time, affected services, and next update time.
5. Do not rotate validator keys or delete data directories until evidence is copied and the recovery path is agreed.

## Consensus Halt

Collect:

```powershell
build\aetrad.exe status --node <rpc> --output json
build\aetrad.exe query block --node <rpc> --output json
build\aetrad.exe query tendermint-validator-set --node <rpc> --output json
```

Then:

- compare latest height and app hash across at least three validators,
- inspect recent logs for panic, out-of-gas, invalid evidence, or slashing messages,
- stop non-validator load generators,
- decide whether restart, config rollback, or coordinated upgrade is needed.

## Suspected Fund Or Admin Exploit

Collect:

- tx hash and full `query tx` JSON,
- affected account balances,
- affected module params,
- tokenfactory denom metadata,
- DEX pool state and module balances,
- fee params.

Do not patch around the exploit without adding a regression test in the relevant package or e2e smoke.

## Faucet Incident

Immediate actions:

- disable faucet web/API service,
- stop faucet key usage,
- record last successful tx hash,
- query faucet balance,
- rotate faucet credentials if the key may be exposed.

The faucet is off-chain only; it must not be granted mint authority.

## Snapshot Or State-Sync Incident

If a published snapshot is bad:

- remove the download link,
- publish affected height/hash/checksum,
- provide fallback snapshot or genesis sync instructions,
- verify archive excludes keyrings, `priv_validator_key.json`, `node_key.json`, and local secrets.

If state sync trust data is bad:

- revoke the trust height/hash announcement,
- publish corrected values from two independent RPC servers,
- tell joining validators to reset data before retrying.

## Communication

Each update should include:

- severity,
- started at UTC time,
- affected height range,
- affected services,
- operator action required,
- next update time,
- owner.

After resolution, publish a postmortem with root cause, impact, fix, tests added, and follow-up owner.

