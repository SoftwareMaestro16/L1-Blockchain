package appconfig

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestConfigureSDKSetsAetraAddressAndBondDenom(t *testing.T) {
	home := ConfigureSDK(".aetra")

	require.True(t, strings.HasSuffix(home, ".aetra"), home)
	require.Equal(t, "ae", sdk.GetConfig().GetBech32AccountAddrPrefix())
	require.Equal(t, "aevaloper", sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	require.Equal(t, "aevalcons", sdk.GetConfig().GetBech32ConsensusAddrPrefix())
	require.Equal(t, appparams.BaseDenom, sdk.DefaultBondDenom)
}
