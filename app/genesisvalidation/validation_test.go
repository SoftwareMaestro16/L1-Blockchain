package genesisvalidation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestValidateAuthGenesisRequiresState(t *testing.T) {
	err := ValidateAuthGenesis(testCodec(), State{})
	require.ErrorContains(t, err, "missing auth genesis state")
}

func TestValidateStakingGenesisRequiresState(t *testing.T) {
	err := ValidateStakingGenesis(testCodec(), State{}, "naet")
	require.ErrorContains(t, err, "missing staking genesis state")
}

func testCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}
