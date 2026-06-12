package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const authority = prototype.DefaultAuthority

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func chain(id string, height uint64) types.RegisteredChain {
	return types.RegisteredChain{
		ChainID:	id,
		Status:		types.ChainStatusActive,
		ClientType:	types.ClientTypeLightClient,
		TrustLevel:	types.TrustLevelHigh,
		LightClientRef:	"lc-" + id,
		RiskScore:	10,
		Finality: types.FinalityAssumptions{
			ConfirmationBlocks:	12,
			FinalitySeconds:	180,
		},
		Timeout: types.TimeoutParameters{
			TimeoutBlocks:	1_000,
			TimeoutSeconds:	3_600,
		},
		RiskPolicyID:		id + "-risk",
		RegisteredHeight:	height,
	}
}

func riskPolicy(id string) types.RiskPolicy {
	return types.RiskPolicy{
		PolicyID:		id + "-risk",
		ChainID:		id,
		MaxRiskScore:		50,
		AllowedClientTypes:	[]string{types.ClientTypeLightClient},
		MinTrustLevel:		types.TrustLevelMedium,
		RequireLightClient:	true,
	}
}

func registerChain(t *testing.T, k *Keeper, id string) {
	t.Helper()
	_, err := k.RegisterChain(types.MsgRegisterChain{Authority: authority, Chain: chain(id, 1), RiskPolicy: riskPolicy(id)})
	require.NoError(t, err)
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
}

func TestRegisterChain(t *testing.T) {
	k := setupKeeper(t)
	record, err := k.RegisterChain(types.MsgRegisterChain{Authority: authority, Chain: chain("ethereum", 1), RiskPolicy: riskPolicy("ethereum")})
	require.NoError(t, err)
	require.Equal(t, "ethereum", record.ChainID)

	chains, err := k.RegisteredChains()
	require.NoError(t, err)
	require.Len(t, chains, 1)
}

func TestDuplicateChainRejected(t *testing.T) {
	k := setupKeeper(t)
	registerChain(t, &k, "ethereum")

	_, err := k.RegisterChain(types.MsgRegisterChain{Authority: authority, Chain: chain("ethereum", 2), RiskPolicy: riskPolicy("ethereum")})
	require.ErrorContains(t, err, "already registered")
}

func TestRegisterChannelAndRoute(t *testing.T) {
	k := setupKeeper(t)
	registerChain(t, &k, "aetra")
	registerChain(t, &k, "ethereum")

	channel, err := k.RegisterChannel(types.MsgRegisterChannel{
		Authority:	authority,
		Channel: types.ChannelRecord{
			ChannelID:		"channel-0",
			ChainID:		"ethereum",
			CounterpartyChainID:	"aetra",
			ClientID:		"client-eth",
			Active:			true,
			RegisteredHeight:	2,
		},
		Routes: []types.BridgeRoute{{
			RouteID:	"eth-aet",
			SourceChainID:	"ethereum",
			TargetChainID:	"aetra",
			ChannelID:	"channel-0",
			BridgeID:	"bridge-eth",
			Enabled:	true,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "channel-0", channel.ChannelID)
	route, found, err := k.BridgeRoute("eth-aet")
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, route.Enabled)
}

func TestPauseChainDisablesRoutes(t *testing.T) {
	k := setupKeeper(t)
	registerChain(t, &k, "aetra")
	registerChain(t, &k, "ethereum")
	_, err := k.RegisterChannel(types.MsgRegisterChannel{
		Authority:	authority,
		Channel: types.ChannelRecord{
			ChannelID:		"channel-0",
			ChainID:		"ethereum",
			CounterpartyChainID:	"aetra",
			ClientID:		"client-eth",
			Active:			true,
			RegisteredHeight:	2,
		},
		Routes: []types.BridgeRoute{{
			RouteID:	"eth-aet",
			SourceChainID:	"ethereum",
			TargetChainID:	"aetra",
			ChannelID:	"channel-0",
			BridgeID:	"bridge-eth",
			Enabled:	true,
		}},
	})
	require.NoError(t, err)

	paused, err := k.UpdateChainStatus(types.MsgUpdateChainStatus{Authority: authority, ChainID: "ethereum", Status: types.ChainStatusPaused, Height: 3})
	require.NoError(t, err)
	require.Equal(t, types.ChainStatusPaused, paused.Status)
	route, found, err := k.BridgeRoute("eth-aet")
	require.NoError(t, err)
	require.True(t, found)
	require.False(t, route.Enabled)
}

func TestRegisterActiveChainWithoutRiskPolicyRejected(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterChain(types.MsgRegisterChain{Authority: authority, Chain: chain("ethereum", 1)})
	require.ErrorContains(t, err, "risk policy")
}

func TestExportImportPreservesRegistryOrder(t *testing.T) {
	k := setupKeeper(t)
	registerChain(t, &k, "zeta")
	registerChain(t, &k, "aetra")
	registerChain(t, &k, "ethereum")

	exported := k.ExportGenesis()
	var imported Keeper
	require.NoError(t, imported.InitGenesis(exported))
	chains, err := imported.RegisteredChains()
	require.NoError(t, err)
	require.Equal(t, []string{"aetra", "ethereum", "zeta"}, []string{chains[0].ChainID, chains[1].ChainID, chains[2].ChainID})
}
