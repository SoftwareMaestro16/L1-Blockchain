package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActiveChainRequiresRiskPolicy(t *testing.T) {
	state := EmptyCrossChainRegistryState()
	state.Chains = append(state.Chains, chain("ethereum", ChainStatusActive))

	require.ErrorContains(t, state.Validate(DefaultCrossChainRegistryParams()), "risk policy")
}

func TestFinalityParametersBounded(t *testing.T) {
	params := DefaultCrossChainRegistryParams()
	state := validState()
	state.Chains[0].Finality.ConfirmationBlocks = params.MaxFinalityBlocks + 1

	require.ErrorContains(t, state.Validate(params), "finality")
}

func TestActiveBridgeRouteRequiresActiveChain(t *testing.T) {
	state := validState()
	state.Chains[1].Status = ChainStatusPaused

	require.ErrorContains(t, state.Validate(DefaultCrossChainRegistryParams()), "active bridge route")
}

func TestDuplicateChainRejectedByInvariant(t *testing.T) {
	state := validState()
	state.Chains = append(state.Chains, state.Chains[0])

	require.ErrorContains(t, state.Validate(DefaultCrossChainRegistryParams()), "duplicate chain")
}

func validState() CrossChainRegistryState {
	state := EmptyCrossChainRegistryState()
	state.Chains = append(state.Chains, chain("aetra", ChainStatusActive), chain("ethereum", ChainStatusActive))
	state.RiskPolicies = append(state.RiskPolicies, policy("aetra"), policy("ethereum"))
	state.Channels = append(state.Channels, ChannelRecord{
		ChannelID:		"channel-0",
		ChainID:		"ethereum",
		CounterpartyChainID:	"aetra",
		ClientID:		"client-eth",
		Active:			true,
		RegisteredHeight:	2,
	})
	state.BridgeRoutes = append(state.BridgeRoutes, BridgeRoute{
		RouteID:	"eth-aet",
		SourceChainID:	"ethereum",
		TargetChainID:	"aetra",
		ChannelID:	"channel-0",
		BridgeID:	"bridge-eth",
		Enabled:	true,
	})
	return state
}

func chain(chainID, status string) RegisteredChain {
	return RegisteredChain{
		ChainID:	chainID,
		Status:		status,
		ClientType:	ClientTypeLightClient,
		TrustLevel:	TrustLevelHigh,
		LightClientRef:	"lc-" + chainID,
		RiskScore:	10,
		Finality: FinalityAssumptions{
			ConfirmationBlocks:	12,
			FinalitySeconds:	180,
		},
		Timeout: TimeoutParameters{
			TimeoutBlocks:	1_000,
			TimeoutSeconds:	3_600,
		},
		RiskPolicyID:		chainID + "-risk",
		RegisteredHeight:	1,
		UpdatedHeight:		1,
	}
}

func policy(chainID string) RiskPolicy {
	return RiskPolicy{
		PolicyID:		chainID + "-risk",
		ChainID:		chainID,
		MaxRiskScore:		50,
		AllowedClientTypes:	[]string{ClientTypeLightClient},
		MinTrustLevel:		TrustLevelMedium,
		RequireLightClient:	true,
	}
}
