package keeper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/bridge-hub/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	authority	= prototype.DefaultAuthority
	bridgeID	= "eth-mainnet"
)

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.BridgeParams.DefaultDailyLimit = 100
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func bridge(height uint64) types.BridgeRecord {
	return types.BridgeRecord{
		BridgeID:		bridgeID,
		SourceChain:		"ethereum",
		TargetChain:		"aetra",
		Operators:		[]string{"op2", "op1"},
		ProofPolicy:		types.ProofPolicyLightClient,
		DailyLimit:		100,
		FeePolicy:		types.BridgeFeePolicy{FeeBps: 10, Collector: "treasury"},
		RegisteredHeight:	height,
	}
}

func event(id string, amount, height uint64) types.BridgeEvent {
	return types.BridgeEvent{
		EventID:		id,
		BridgeID:		bridgeID,
		SourceChain:		"ethereum",
		Asset:			"ETH",
		Amount:			amount,
		ProofPolicy:		types.ProofPolicyLightClient,
		ProofRoot:		strings.Repeat("a", 64),
		SubmittedHeight:	height,
	}
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
}

func TestRegisterBridge(t *testing.T) {
	k := setupKeeper(t)
	record, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	require.Equal(t, bridgeID, record.BridgeID)
	require.Equal(t, []string{"op1", "op2"}, record.Operators)

	bridges, err := k.Bridges()
	require.NoError(t, err)
	require.Len(t, bridges, 1)
}

func TestPauseResumeBridge(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)

	paused, err := k.PauseBridge(types.MsgPauseBridge{Authority: authority, BridgeID: bridgeID, Height: 2})
	require.NoError(t, err)
	require.True(t, paused.Paused)
	resumed, err := k.ResumeBridge(types.MsgResumeBridge{Authority: authority, BridgeID: bridgeID, Height: 3})
	require.NoError(t, err)
	require.False(t, resumed.Paused)
}

func TestMappingConflictRejected(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	_, err = k.RegisterAssetMapping(types.MsgRegisterAssetMapping{
		Authority:	authority,
		Mapping:	types.AssetMapping{BridgeID: bridgeID, SourceAsset: "ETH", TargetAsset: "aETH", Decimals: 18, Enabled: true},
	})
	require.NoError(t, err)
	_, err = k.RegisterAssetMapping(types.MsgRegisterAssetMapping{
		Authority:	authority,
		Mapping:	types.AssetMapping{BridgeID: bridgeID, SourceAsset: "ETH", TargetAsset: "wrappedETH", Decimals: 18, Enabled: true},
	})
	require.ErrorContains(t, err, "conflict")
}

func TestDailyLimitEnforced(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-1", 60, 2)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-2", 50, 3)})
	require.NoError(t, err)
	_, err = k.FinalizeBridgeEvent(types.MsgFinalizeBridgeEvent{Authority: authority, EventID: "event-1", Height: 4})
	require.NoError(t, err)
	_, err = k.FinalizeBridgeEvent(types.MsgFinalizeBridgeEvent{Authority: authority, EventID: "event-2", Height: 5})
	require.ErrorContains(t, err, "daily limit")
}

func TestPausedBridgeCannotFinalizeAndDuplicateEventRejected(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-1", 10, 2)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-1", 10, 2)})
	require.ErrorContains(t, err, "already submitted")
	_, err = k.PauseBridge(types.MsgPauseBridge{Authority: authority, BridgeID: bridgeID, Height: 3})
	require.NoError(t, err)
	_, err = k.FinalizeBridgeEvent(types.MsgFinalizeBridgeEvent{Authority: authority, EventID: "event-1", Height: 4})
	require.ErrorContains(t, err, "paused")
}

func TestBridgeEventCannotFinalizeTwice(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-1", 10, 2)})
	require.NoError(t, err)
	finalized, err := k.FinalizeBridgeEvent(types.MsgFinalizeBridgeEvent{Authority: authority, EventID: "event-1", Height: 3})
	require.NoError(t, err)
	require.Equal(t, types.BridgeEventFinalized, finalized.Status)
	_, err = k.FinalizeBridgeEvent(types.MsgFinalizeBridgeEvent{Authority: authority, EventID: "event-1", Height: 4})
	require.ErrorContains(t, err, "twice")
}

func TestExportImportPreservesPendingEvents(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterBridge(types.MsgRegisterBridge{Authority: authority, Bridge: bridge(1)})
	require.NoError(t, err)
	_, err = k.SubmitBridgeEvent(types.MsgSubmitBridgeEvent{Submitter: "relayer", Event: event("event-1", 10, 2)})
	require.NoError(t, err)

	exported := k.ExportGenesis()
	var imported Keeper
	require.NoError(t, imported.InitGenesis(exported))
	events, err := imported.BridgeEvents(bridgeID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, types.BridgeEventPending, events[0].Status)
}
