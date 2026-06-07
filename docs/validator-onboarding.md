# Validator Onboarding

This guide is for a clean public testnet validator join. Localnet examples use PowerShell paths; public operators must replace local paths, chain id, peers, and keyring backend with launch values.

## Build

```powershell
git clone https://github.com/SoftwareMaestro16/L1-Blockchain.git
cd L1-Blockchain
.\scripts\build-aetrad.ps1
build\aetrad.exe version --long --output json
```

Verify that the commit matches the published testnet release commit.

## Initialize Node

```powershell
$CHAIN_ID = "<testnet-chain-id>"
$HOME = "$env:USERPROFILE\.aetra"
build\aetrad.exe init <moniker> --chain-id $CHAIN_ID --home $HOME
```

Replace `$HOME\config\genesis.json` with the published genesis file, then validate:

```powershell
build\aetrad.exe genesis validate-genesis $HOME\config\genesis.json --home $HOME
```

Configure peers and persistent peers from the launch announcement. Do not reuse localnet keys.

## Create Validator Key

Use a secure keyring backend for public testnet:

```powershell
build\aetrad.exe keys add <key-name> --home $HOME --keyring-backend os
build\aetrad.exe keys show <key-name> -a --home $HOME --keyring-backend os
```

Store mnemonic backup offline. Never commit mnemonics, keyrings, `priv_validator_key.json`, or node keys.

## Sync

Start from genesis:

```powershell
build\aetrad.exe start --home $HOME
```

Or use state sync from the published trust height/hash and RPC server list:

```powershell
# Edit $HOME\config\config.toml with enable=true, rpc_servers, trust_height, trust_hash, trust_period.
build\aetrad.exe start --home $HOME
```

Check sync status:

```powershell
build\aetrad.exe status --node tcp://127.0.0.1:26657 --output json
```

The node is caught up when `catching_up` is false.

## Create Validator

Fund the validator account from the faucet or launch allocation first. Then create the validator using `naet`:

```powershell
$VAL_PUBKEY = build\aetrad.exe comet show-validator --home $HOME
build\aetrad.exe tx staking create-validator `
  --amount 100000000naet `
  --pubkey $VAL_PUBKEY `
  --moniker <moniker> `
  --chain-id $CHAIN_ID `
  --from <key-name> `
  --home $HOME `
  --keyring-backend os `
  --fees 1000000naet `
  --commission-rate 0.05 `
  --commission-max-rate 0.20 `
  --commission-max-change-rate 0.01 `
  --min-self-delegation 1 `
  --node tcp://127.0.0.1:26657 `
  -y
```

Verify:

```powershell
build\aetrad.exe query staking validators --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query tendermint-validator-set --node tcp://127.0.0.1:26657 --output json
```

## Operations

Monitor:

- latest block height,
- validator voting power,
- missed block counter,
- disk usage,
- process restart count,
- peer count,
- RPC/indexer lag if serving public endpoints.

Before restart, stop cleanly and preserve `$HOME\data`, `$HOME\config\priv_validator_key.json`, and `$HOME\config\node_key.json`.

## CosmWasm Contract Smoke

If and only if the launch config explicitly enables CosmWasm, deploy the smoke contract:

```powershell
.\tests\e2e\cosmwasm_smoke.ps1 -EnableWasm -ContractWasm .\artifacts\cw_template.wasm -Node tcp://127.0.0.1:26657 -ChainId $CHAIN_ID -AppHome $HOME -From <key-name>
```

If wasm is not enabled, the disabled-by-default check must pass:

```powershell
.\tests\e2e\cosmwasm_smoke.ps1 -Node tcp://127.0.0.1:26657
```

