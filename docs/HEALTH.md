# Aetra Testnet Health Check Documentation

## Overview

This document defines health endpoints, commands, and monitoring procedures for Aetra testnet validators and operators.

## Health Check Categories

### 1. Process Alive

**Command:**
```bash
curl http://localhost:26657/status
```

**Expected:**
- HTTP 200 response
- JSON body with `result.sync_info` field present

**Alert Conditions:**
- HTTP non-200 response
- Connection refused
- Timeout

### 2. RPC Status

**Command:**
```bash
aetrad status --node http://localhost:26657
```

**Expected Response Fields:**
```json
{
  "NodeInfo": {
    "version": "...",
    "network": "aetra-testnet-1"
  },
  "SyncInfo": {
    "latest_block_height": "12345",
    "catching_up": false
  },
  "ValidatorInfo": {
    "address": "...",
    "voting_power": "1000000"
  }
}
```

### 3. Block Height Increasing

**Command:**
```bash
# Poll twice with 10 second interval
HEIGHT1=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
sleep 10
HEIGHT2=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
[ "$HEIGHT2" -gt "$HEIGHT1" ]
```

**Expected:**
- `HEIGHT2 > HEIGHT1` within 10 seconds

**Alert Conditions:**
- Height not increasing
- Height decreasing

### 4. Catching Up False

**Command:**
```bash
curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up'
```

**Expected:**
```
false
```

**Alert Conditions:**
- `true` - node is syncing
- Response parsing error

### 5. Peer Count

**Command:**
```bash
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
```

**Expected:**
- Minimum 1 peer for testnet
- Recommended 3+ peers for production

**Alert Conditions:**
- 0 peers (isolated)
- Peer count lower than expected

### 6. Validator Signing Info

**Command:**
```bash
# Get validator consensus address
curl -s http://localhost:26657/dumps | jq -r '.validators[] | select(.address == "YOUR_VALOPER")'

# Check signing status
curl -s http://localhost:26657/validators | jq '.result.validators[].jailed'
```

**Expected:**
- `jailed: false`
- `tombstoned: false`
- `missed_blocks` within acceptable range

### 7. App Invariant Command

**Command:**
```bash
aetrad export --for-export > /dev/null 2>&1
echo $?
```

**Expected:**
- Exit code 0
- No panic or assertion failure

## Health Check Script

Use the included health script for automated monitoring:

```powershell
# Basic health check
.\scripts\localnet\health.ps1 -OutputDir .localnet

# With JSON output for monitoring systems
.\scripts\localnet\health.ps1 -OutputDir .localnet -Json

# Specify validator count
.\scripts\localnet\health.ps1 -OutputDir .localnet -ValidatorCount 4

# Extended check with log tail
.\scripts\localnet\health.ps1 -OutputDir .localnet -LogTailLines 100
```

## Prometheus Metrics

Expose metrics at `http://localhost:26660/metrics`:

| Metric | Description |
|--------|-------------|
| `aetrad_block_height` | Current block height |
| `aetrad_validator_voting_power` | Validator voting power |
| `aetrad_peers` | Number of connected peers |
| `aetrad_missed_blocks` | Missed blocks in current window |

## Health Check Interval Recommendations

| Check | Interval | Timeout |
|-------|----------|---------|
| Process Alive | 30s | 5s |
| RPC Status | 30s | 10s |
| Block Height | 60s | 30s |
| Catching Up | 60s | 10s |
| Peer Count | 5m | 30s |
| Invariant Check | 1h | 300s |

## Alert Thresholds

| Condition | Severity | Action |
|-----------|----------|--------|
| Node unreachable > 60s | Critical | Restart process |
| Catching up > 5min | Warning | Check peer connections |
| Peers < 2 | Warning | Check network configuration |
| Height not increasing > 120s | Critical | Check consensus |
| Missed blocks > 50% | Critical | Check validator status |

## Network Peer List

See `docs/testnet/peers.example.json` for peer configuration template.

## Troubleshooting

### Node Not Producing Blocks
1. Check `aetrad status --node <rpc>` for catching_up status
2. Verify peer connections with `curl http://localhost:26657/net_info`
3. Check validator signing info at `curl http://localhost:26657/signing_info`

### Peers Not Connecting
1. Verify firewall rules allow port 26656
2. Check seed nodes are accessible
3. Verify node ID and IP in peer list

### Invariant Check Failing
1. Check logs for panic or assertion messages
2. Verify genesis configuration
3. Run `aetrad export` to identify state corruption

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-06-09 | Initial health check documentation |