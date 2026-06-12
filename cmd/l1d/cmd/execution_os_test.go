package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/require"

	l1cmd "github.com/sovereign-l1/l1/cmd/l1d/cmd"
)

func TestExecutionOSSmokeCommandOutputsOperatorReport(t *testing.T) {
	rootCmd := l1cmd.NewRootCmd()
	var out bytes.Buffer
	homeDir := t.TempDir()
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{
		"execution-os", "smoke",
		"--profile", "execution-os-sim",
		fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", homeDir))

	var report struct {
		Profile	string	`json:"profile"`
		Load	struct {
			ScoreBps	uint32	`json:"score_bps"`
			Band		string	`json:"band"`
		}	`json:"load"`
		Routing	struct {
			ZoneID string `json:"zone_id"`
		}	`json:"routing"`
		Sharding	struct {
			ActiveShardCount uint32 `json:"active_shard_count"`
		}	`json:"sharding"`
		Mesh	struct {
			ReceiptStatus string `json:"receipt_status"`
		}	`json:"mesh"`
		Identity	struct {
			Domain string `json:"domain"`
		}	`json:"identity"`
		ProductionLive	bool	`json:"production_live"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &report), out.String())
	require.Equal(t, "execution-os-sim", report.Profile)
	require.Greater(t, report.Load.ScoreBps, uint32(0))
	require.NotEmpty(t, report.Load.Band)
	require.Equal(t, "FINANCIAL_ZONE", report.Routing.ZoneID)
	require.GreaterOrEqual(t, report.Sharding.ActiveShardCount, uint32(2))
	require.Equal(t, "SUCCESS", report.Mesh.ReceiptStatus)
	require.Equal(t, "operator.aet", report.Identity.Domain)
	require.False(t, report.ProductionLive)
}

func TestExecutionOSDiagnosticsCommandReadsGenesis(t *testing.T) {
	genesis := `{
  "app_state": {
    "load": {
      "Version": 1,
      "Params": {"Enabled": true, "TestnetProfile": true, "ProductionVersionGate": "", "Authority": "4:0000000000000000000000000000000000000000000000000000000000000001", "DefaultQueryLimit": 50, "MaxQueryLimit": 200},
      "LoadParams": {"WindowBlocks":60,"AlphaNumerator":2,"AlphaDenominator":61,"MaxDeltaBps":500,"TargetMempoolSize":10000,"TargetBlockGas":20000000,"TargetLatencyBlocks":5,"TargetExecutionSteps":20000000,"MempoolSizeWeightBps":2000,"BlockUtilizationWeightBps":3000,"TxLatencyWeightBps":2000,"FailureRateWeightBps":1000,"ExecutionTimeWeightBps":2000},
      "EMA": {"LoadScoreBps": 500, "WindowHeight": 1},
      "History": []
    },
    "routing": {
      "Version": 1,
      "Params": {"Enabled": true, "TestnetProfile": true, "ProductionVersionGate": "", "Authority": "4:0000000000000000000000000000000000000000000000000000000000000001", "DefaultQueryLimit": 50, "MaxQueryLimit": 200},
      "RoutingEpoch": 1,
      "Shards": [{"ZoneID": "FINANCIAL_ZONE", "ActiveShards": 2}]
    },
    "zones": {
      "Version": 1,
      "Params": {"Enabled": false, "TestnetProfile": false, "ProductionVersionGate": "", "Authority": "4:0000000000000000000000000000000000000000000000000000000000000001", "DefaultQueryLimit": 50, "MaxQueryLimit": 200},
      "State": {"Zones": [], "ActiveZones": [], "Commitments": []}
    },
    "mesh": {
      "Version": 1,
      "Params": {"Enabled": false, "TestnetProfile": false, "ProductionVersionGate": "", "Authority": "4:0000000000000000000000000000000000000000000000000000000000000001", "DefaultQueryLimit": 50, "MaxQueryLimit": 200},
      "State": {"CurrentHeight":0,"Params":{"MaxFinalityAge":256},"Destinations":[],"FinalizedCommitments":[],"ReplayMarkers":[],"Receipts":[],"BounceReceipts":[],"RefundReceipts":[]}
    }
  }
}`
	homeDir := t.TempDir()
	genesisPath := filepath.Join(homeDir, "genesis.json")
	require.NoError(t, os.WriteFile(genesisPath, []byte(genesis), 0o600))

	rootCmd := l1cmd.NewRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{
		"execution-os", "diagnostics",
		"--profile", "execution-os-sim",
		"--genesis", genesisPath,
		fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", homeDir))

	var diagnostics struct {
		CurrentLoadScoreBps	uint32	`json:"current_load_score_bps"`
		ActiveShards		[]struct {
			ZoneID		string	`json:"zone_id"`
			ActiveShards	uint32	`json:"active_shards"`
		}	`json:"active_shards"`
		FeatureGates	map[string]struct {
			Enabled		bool	`json:"enabled"`
			TestnetProfile	bool	`json:"testnet_profile"`
		}	`json:"feature_gates"`
		ProductionLive	bool	`json:"production_live"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &diagnostics), out.String())
	require.Equal(t, uint32(500), diagnostics.CurrentLoadScoreBps)
	require.Len(t, diagnostics.ActiveShards, 1)
	require.Equal(t, "FINANCIAL_ZONE", diagnostics.ActiveShards[0].ZoneID)
	require.True(t, diagnostics.FeatureGates["load"].Enabled)
	require.True(t, diagnostics.FeatureGates["routing"].TestnetProfile)
	require.False(t, diagnostics.ProductionLive)
}
