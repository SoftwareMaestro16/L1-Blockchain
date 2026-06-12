package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyTxStableTypeStrings(t *testing.T) {
	tests := []struct {
		msgType	string
		class	TxClass
	}{
		{MsgTypeSoftwareUpgrade, TxClassCriticalSystem},
		{MsgTypeDelegate, TxClassStakingGovSecurity},
		{"cosmos.bank.v1beta1.MsgSend", TxClassFinancial},
		{MsgTypeIdentityRegister, TxClassIdentity},
		{MsgTypeWasmExecute, TxClassContract},
		{MsgTypeAsyncSend, TxClassAsyncMessage},
		{MsgTypeMemoAttach, TxClassApplication},
	}

	for _, tt := range tests {
		t.Run(tt.msgType, func(t *testing.T) {
			class, err := ClassifyTx(tt.msgType)
			require.NoError(t, err)
			require.Equal(t, tt.class, class)
		})
	}
}

func TestRouteSameInputDeterministic(t *testing.T) {
	input := validFinancialInput()

	left, err := Route(input)
	require.NoError(t, err)
	right, err := Route(input)
	require.NoError(t, err)

	require.Equal(t, left, right)
	require.Equal(t, TxClassFinancial, left.TxClass)
	require.Equal(t, ZoneFinancial, left.ZoneID)
	require.Less(t, uint32(left.ShardID), left.ActiveShards)
}

func TestDifferentMapInsertionOrderProducesSameRouteAndPriority(t *testing.T) {
	leftInput := validFinancialInput()
	leftInput.ActiveShards = map[ZoneID]uint32{
		ZoneFinancial:		8,
		ZoneIdentity:		4,
		ZoneApplication:	16,
	}
	rightInput := validFinancialInput()
	rightInput.ActiveShards = map[ZoneID]uint32{
		ZoneApplication:	16,
		ZoneIdentity:		4,
		ZoneFinancial:		8,
	}

	left, err := Route(leftInput)
	require.NoError(t, err)
	right, err := Route(rightInput)
	require.NoError(t, err)
	require.Equal(t, left, right)

	keys := []PriorityKey{
		BuildPriorityKey(TxClassApplication, 1, 2, 9, hashBytes("c")),
		BuildPriorityKey(TxClassCriticalSystem, 0, 0, 10, hashBytes("a")),
		BuildPriorityKey(TxClassFinancial, 9, 3, 8, hashBytes("b")),
	}
	reversed := []PriorityKey{keys[2], keys[1], keys[0]}

	require.Equal(t, SortPriorityKeys(keys), SortPriorityKeys(reversed))
}

func TestHighPrioritySystemTxSortsBeforeNormalUserTx(t *testing.T) {
	systemInput := validCoreInput()
	systemInput.MsgType = MsgTypeSoftwareUpgrade
	systemInput.TxHash = hashBytes("system")
	systemDecision, err := Route(systemInput)
	require.NoError(t, err)

	userInput := validFinancialInput()
	userInput.FeeClass = MaxFeeClass
	userInput.ReputationClass = MaxReputationClass
	userInput.TxHash = hashBytes("user")
	userDecision, err := Route(userInput)
	require.NoError(t, err)

	ordered := SortDecisions([]RouteDecision{userDecision, systemDecision})
	require.Equal(t, systemDecision.PriorityKey.TxHash, ordered[0].PriorityKey.TxHash)
	require.Equal(t, userDecision.PriorityKey.TxHash, ordered[1].PriorityKey.TxHash)
}

func TestPriorityOrderingUsesFeeReputationHeightAndHash(t *testing.T) {
	oldLowFee := BuildPriorityKey(TxClassFinancial, 1, 1, 10, hashBytes("old-low"))
	newHighFee := BuildPriorityKey(TxClassFinancial, 2, 1, 20, hashBytes("new-high"))
	highRep := BuildPriorityKey(TxClassFinancial, 2, 2, 30, hashBytes("high-rep"))
	olderSame := BuildPriorityKey(TxClassFinancial, 2, 2, 5, hashBytes("older"))
	hashTie := BuildPriorityKey(TxClassFinancial, 2, 2, 5, hashBytes("aaaa"))

	ordered := SortPriorityKeys([]PriorityKey{oldLowFee, newHighFee, highRep, olderSame, hashTie})

	require.Equal(t, hashTie.TxHash, ordered[0].TxHash)
	require.Equal(t, olderSame.TxHash, ordered[1].TxHash)
	require.Equal(t, highRep.TxHash, ordered[2].TxHash)
	require.Equal(t, newHighFee.TxHash, ordered[3].TxHash)
	require.Equal(t, oldLowFee.TxHash, ordered[4].TxHash)
}

func TestFeeAndReputationClassesAreBounded(t *testing.T) {
	key := BuildPriorityKey(TxClassFinancial, MaxFeeClass+999, MaxReputationClass+999, 1, hashBytes("bounded"))

	require.Equal(t, MaxFeeClass, key.FeeClass)
	require.Equal(t, MaxReputationClass, key.ReputationClass)
}

func TestRouteRejectsZeroActiveShards(t *testing.T) {
	input := validFinancialInput()
	input.ActiveShards = map[ZoneID]uint32{ZoneFinancial: 0}

	_, err := Route(input)
	require.ErrorContains(t, err, "active shards")
}

func TestRouteRejectsUnknownTxClass(t *testing.T) {
	_, err := ClassifyTx("/unknown.Msg")
	require.ErrorContains(t, err, "unknown routing message type")

	_, err = Route(RouteInput{
		MsgType:	"/unknown.Msg",
		FeeDenom:	NativeFeeDenom,
		TxHash:		hashBytes("unknown"),
	})
	require.ErrorContains(t, err, "unknown routing message type")
}

func TestZoneForClassRejectsMissingZone(t *testing.T) {
	_, err := ZoneForClass(TxClass("BOGUS"))
	require.ErrorContains(t, err, "missing zone")
}

func TestRouteRejectsZeroAddressPrimaryActor(t *testing.T) {
	input := validFinancialInput()
	input.Locality.AccountKey = bytes.Repeat([]byte{0}, 20)

	_, err := Route(input)
	require.ErrorContains(t, err, "zero address")
}

func TestRouteRejectsEmptyPrimaryActorForShardedClasses(t *testing.T) {
	input := validFinancialInput()
	input.Locality.AccountKey = nil
	input.Locality.AssetDenom = ""

	_, err := Route(input)
	require.ErrorContains(t, err, "primary actor")
}

func TestRouteRejectsNonNativeFeeDenom(t *testing.T) {
	input := validFinancialInput()
	input.FeeDenom = "testtoken"

	_, err := Route(input)
	require.ErrorContains(t, err, "naet")
}

func TestIdentityDomainNormalizationAndValidation(t *testing.T) {
	input := RouteInput{
		MsgType:		MsgTypeIdentityRegister,
		FeeDenom:		NativeFeeDenom,
		FeeClass:		1,
		ReputationClass:	1,
		AdmissionHeight:	7,
		TxHash:			hashBytes("identity"),
		RoutingEpoch:		3,
		ActiveShards:		map[ZoneID]uint32{ZoneIdentity: 4},
		Locality:		Locality{Domain: " Alice.AET "},
	}
	decision, err := Route(input)
	require.NoError(t, err)
	require.Equal(t, []byte("domain:alice.aet"), decision.PrimaryActor)

	input.Locality.Domain = "alice"
	_, err = Route(input)
	require.ErrorContains(t, err, ".aet")
}

func TestContractAndAsyncLocality(t *testing.T) {
	contractInput := RouteInput{
		MsgType:	MsgTypeWasmExecute,
		FeeDenom:	NativeFeeDenom,
		TxHash:		hashBytes("contract"),
		ActiveShards:	map[ZoneID]uint32{ZoneContract: 2},
		Locality:	Locality{ContractAddress: actor(0x44)},
	}
	contractDecision, err := Route(contractInput)
	require.NoError(t, err)
	require.Equal(t, ZoneContract, contractDecision.ZoneID)
	require.Equal(t, actor(0x44), contractDecision.PrimaryActor)

	asyncInput := RouteInput{
		MsgType:	MsgTypeAsyncSend,
		FeeDenom:	NativeFeeDenom,
		TxHash:		hashBytes("async"),
		ActiveShards:	map[ZoneID]uint32{ZoneApplication: 2},
		Locality:	Locality{AsyncDestination: actor(0x55)},
	}
	asyncDecision, err := Route(asyncInput)
	require.NoError(t, err)
	require.Equal(t, ZoneApplication, asyncDecision.ZoneID)
	require.Equal(t, actor(0x55), asyncDecision.PrimaryActor)
}

func TestRouteRejectsMalformedLocality(t *testing.T) {
	input := validFinancialInput()
	input.Locality.AssetDenom = "bad denom"
	input.Locality.AccountKey = nil

	_, err := Route(input)
	require.ErrorContains(t, err, "asset denom")
}

func validFinancialInput() RouteInput {
	return RouteInput{
		MsgType:		MsgTypeBankSend,
		FeeDenom:		NativeFeeDenom,
		FeeClass:		2,
		ReputationClass:	3,
		AdmissionHeight:	42,
		TxHash:			hashBytes("financial"),
		RoutingEpoch:		9,
		ActiveShards:		map[ZoneID]uint32{ZoneFinancial: 8},
		Locality:		Locality{AccountKey: actor(0x11)},
	}
}

func validCoreInput() RouteInput {
	return RouteInput{
		MsgType:		MsgTypeDelegate,
		FeeDenom:		NativeFeeDenom,
		FeeClass:		1,
		ReputationClass:	1,
		AdmissionHeight:	1,
		TxHash:			hashBytes("core"),
		Locality:		Locality{AccountKey: actor(0x22)},
	}
}

func actor(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, 20)
}

func hashBytes(seed string) []byte {
	raw := make([]byte, 32)
	encoded := hex.EncodeToString([]byte(seed))
	copy(raw, []byte(encoded))
	return raw
}
