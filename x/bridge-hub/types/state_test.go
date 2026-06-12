package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentWindowStartDeterministic(t *testing.T) {
	require.Equal(t, uint64(1), CurrentWindowStart(1))
	require.Equal(t, uint64(1), CurrentWindowStart(86_400))
	require.Equal(t, uint64(86_401), CurrentWindowStart(86_401))
}

func TestProofPolicyMustMatchRegisteredBridge(t *testing.T) {
	params := DefaultBridgeHubParams()
	state := EmptyBridgeHubState()
	state.Bridges = append(state.Bridges, BridgeRecord{
		BridgeID:		"eth-mainnet",
		SourceChain:		"ethereum",
		TargetChain:		"aetra",
		Operators:		[]string{"op1"},
		ProofPolicy:		ProofPolicyLightClient,
		DailyLimit:		100,
		RegisteredHeight:	1,
		UpdatedHeight:		1,
	})
	state.Events = append(state.Events, BridgeEvent{
		EventID:		"event-1",
		BridgeID:		"eth-mainnet",
		SourceChain:		"ethereum",
		Asset:			"ETH",
		Amount:			1,
		ProofPolicy:		ProofPolicyMultisig,
		ProofRoot:		strings.Repeat("a", 64),
		SubmittedBy:		"relayer",
		SubmittedHeight:	2,
		Status:			BridgeEventPending,
	})

	require.ErrorContains(t, state.Validate(params), "proof policy")
}

func TestAssetMappingConflictInvariant(t *testing.T) {
	params := DefaultBridgeHubParams()
	state := EmptyBridgeHubState()
	state.Bridges = append(state.Bridges, BridgeRecord{
		BridgeID:		"eth-mainnet",
		SourceChain:		"ethereum",
		TargetChain:		"aetra",
		Operators:		[]string{"op1"},
		ProofPolicy:		ProofPolicyLightClient,
		DailyLimit:		100,
		RegisteredHeight:	1,
		UpdatedHeight:		1,
	})
	state.AssetMappings = append(state.AssetMappings,
		AssetMapping{BridgeID: "eth-mainnet", SourceAsset: "ETH", TargetAsset: "aETH", Decimals: 18, Enabled: true},
		AssetMapping{BridgeID: "eth-mainnet", SourceAsset: "ETH", TargetAsset: "wrappedETH", Decimals: 18, Enabled: true},
	)

	require.ErrorContains(t, state.Validate(params), "conflicting")
}
