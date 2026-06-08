package services

import (
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestAutoCLIOptionsUsesAetraAddressCodecs(t *testing.T) {
	opts := AutoCLIOptions(map[string]any{})

	require.NotNil(t, opts.AddressCodec)
	require.NotNil(t, opts.ValidatorAddressCodec)
	require.NotNil(t, opts.ConsensusAddressCodec)
	require.Empty(t, opts.Modules)
}

func TestAutoCLIOptionsSkipsStakingHistoricalInfoHeightConflict(t *testing.T) {
	opts := AutoCLIOptions(map[string]any{})

	stakingOptions := opts.ModuleOptions[stakingtypes.ModuleName]
	require.NotNil(t, stakingOptions)
	require.NotNil(t, stakingOptions.Query)

	var found bool
	for _, rpc := range stakingOptions.Query.RpcCommandOptions {
		if rpc.RpcMethod == "HistoricalInfo" {
			found = true
			require.True(t, rpc.Skip)
		}
	}
	require.True(t, found)
}
