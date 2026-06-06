package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinancialZoneBoundaryRoutesStateUnderSpecPrefixes(t *testing.T) {
	boundary := DefaultFinancialZoneKeeperBoundary()
	require.NoError(t, boundary.Validate())
	require.Equal(t, ZoneIDFinancial, boundary.ZoneID)
	require.Contains(t, boundary.OwnsPrefixes, FinancialBalancesPrefix)
	require.Contains(t, boundary.MessageHandlers, FinancialMessagePaymentSettle)
	require.Contains(t, boundary.ProofKinds, FinancialProofPaymentCondition)

	accountKey, err := FinancialAccountKey("alice")
	require.NoError(t, err)
	require.Equal(t, "financial/accounts/alice", accountKey)

	balanceKey, err := FinancialBalanceKey("alice", "naet")
	require.NoError(t, err)
	require.Equal(t, "financial/balances/alice/naet", balanceKey)

	feeKey, err := FinancialFeeBucketKey("base-fee")
	require.NoError(t, err)
	require.Equal(t, "financial/fees/buckets/base-fee", feeKey)

	denomKey, err := FinancialFactoryDenomKey("factory/alice/token")
	require.NoError(t, err)
	require.Contains(t, denomKey, FinancialTokenFactoryDenomPrefix+"/")

	authorityKey, err := FinancialTokenAuthorityKey("factory/alice/token")
	require.NoError(t, err)
	require.Contains(t, authorityKey, FinancialTokenFactoryAuthorityPrefix+"/")

	poolKey, err := FinancialDEXPoolKey(42)
	require.NoError(t, err)
	require.Equal(t, "financial/dex/pools/00000000000000000042", poolKey)

	orderKey, err := FinancialDEXOrderKey("order-1")
	require.NoError(t, err)
	require.Equal(t, "financial/dex/orders/order-1", orderKey)

	channelKey, err := FinancialPaymentChannelKey("channel-1")
	require.NoError(t, err)
	require.Equal(t, "financial/payments/channels/channel-1", channelKey)

	conditionKey, err := FinancialPaymentConditionKey("condition-1")
	require.NoError(t, err)
	require.Equal(t, "financial/payments/conditions/condition-1", conditionKey)
}

func TestFinancialZoneMessageDrivenTransferIngressAndFeeAccounting(t *testing.T) {
	state := FinancialZoneState{
		Accounts: []string{"bob", "alice"},
		Balances: []FinancialBalance{
			{Address: "bob", Denom: "naet", Amount: 5},
			{Address: "alice", Denom: "naet", Amount: 100},
		},
	}
	msg := FinancialZoneMessage{
		Kind:            FinancialMessageTransfer,
		AccountKey:      "alice",
		CounterpartyKey: "bob",
		Denom:           "naet",
		Amount:          40,
		PayloadHash:     hash("financial-transfer"),
		Sequence:        9,
		GasLimit:        700,
	}
	zoneMsg, err := msg.ZoneMessage()
	require.NoError(t, err)
	require.Equal(t, ZoneIDFinancial, zoneMsg.ZoneID)
	require.Equal(t, string(FinancialMessageTransfer), zoneMsg.MessageType)

	next, receipt, err := ApplyFinancialMessage(state, msg, 12)
	require.NoError(t, err)
	require.Equal(t, ZoneReceiptStatusSuccess, receipt.Status)
	require.Equal(t, msg.PayloadHash, receipt.ItemHash)
	require.Equal(t, uint64(700), receipt.GasUsed)
	require.Equal(t, []string{"alice", "bob"}, next.Accounts)
	require.Equal(t, uint64(60), next.Balances[0].Amount)
	require.Equal(t, uint64(45), next.Balances[1].Amount)

	next, err = CreditFinancialFeeBucket(next, "base-fee", "naet", 3)
	require.NoError(t, err)
	next, err = CreditFinancialFeeBucket(next, "base-fee", "naet", 2)
	require.NoError(t, err)
	require.Equal(t, uint64(5), next.FeeBuckets[0].Amount)
	require.NotEmpty(t, ComputeFinancialZoneStateRoot(next))
}

func TestFinancialZoneTokenfactoryDexAndPaymentHooks(t *testing.T) {
	state := FinancialZoneState{
		DEXPools: []FinancialDEXPool{
			{PoolID: 1, BaseDenom: "naet", QuoteDenom: "factory/alice/token", BaseReserve: 1_000, QuoteReserve: 2_000},
		},
		PaymentChannels: []FinancialPaymentChannel{
			{ChannelID: "channel-1", Payer: "alice", Receiver: "bob", Denom: "naet", EscrowAmount: 100},
		},
		PaymentConditions: []FinancialPaymentCondition{
			{ConditionID: "condition-1", ChannelID: "channel-1", Amount: 25},
		},
	}

	mint := FinancialZoneMessage{
		Kind:        FinancialMessageMintFactoryDenom,
		AccountKey:  "alice",
		Denom:       "factory/alice/token",
		Authority:   "alice",
		Amount:      500,
		PayloadHash: hash("mint"),
		Sequence:    1,
		GasLimit:    10,
	}
	next, _, err := ApplyFinancialMessage(state, mint, 13)
	require.NoError(t, err)
	require.Equal(t, uint64(500), next.FactoryDenoms[0].Supply)
	require.Equal(t, uint64(500), next.Balances[0].Amount)

	swap := FinancialZoneMessage{
		Kind:        FinancialMessageDexSwap,
		AccountKey:  "alice",
		PoolID:      1,
		OrderID:     "order-1",
		Denom:       "naet",
		OutputDenom: "factory/alice/token",
		Amount:      50,
		PayloadHash: hash("swap"),
		Sequence:    2,
		GasLimit:    20,
	}
	next, _, err = ApplyFinancialMessage(next, swap, 14)
	require.NoError(t, err)
	require.Equal(t, uint64(1_050), next.DEXPools[0].BaseReserve)
	require.Equal(t, "open", next.DEXOrders[0].Status)

	settle := FinancialZoneMessage{
		Kind:        FinancialMessageDexSettle,
		OrderID:     "order-1",
		Amount:      1,
		PayloadHash: hash("settle"),
		Sequence:    3,
		GasLimit:    30,
	}
	next, _, err = ApplyFinancialMessage(next, settle, 15)
	require.NoError(t, err)
	require.Equal(t, "settled", next.DEXOrders[0].Status)
	dexReceipt, err := BuildFinancialDEXSettlementReceipt(15, next.DEXOrders[0], ComputeFinancialZoneStateRoot(next), 4)
	require.NoError(t, err)
	require.Equal(t, ZoneIDFinancial, dexReceipt.ZoneID)

	payment := FinancialZoneMessage{
		Kind:        FinancialMessagePaymentSettle,
		ChannelID:   "channel-1",
		Amount:      1,
		PayloadHash: hash("payment"),
		Sequence:    5,
		GasLimit:    40,
	}
	next, _, err = ApplyFinancialMessage(next, payment, 16)
	require.NoError(t, err)
	require.True(t, next.PaymentChannels[0].Finalized)
	hook, err := BuildFinancialPaymentFinalizationHook(16, next.PaymentChannels[0], ComputeFinancialZoneStateRoot(next), 6)
	require.NoError(t, err)
	require.Equal(t, ZoneReceiptStatusSuccess, hook.Status)

	disputed := FinancialZoneMessage{
		Kind:        FinancialMessagePaymentDispute,
		ConditionID: "condition-1",
		Amount:      1,
		PayloadHash: hash("dispute"),
		Sequence:    7,
		GasLimit:    50,
	}
	next, _, err = ApplyFinancialMessage(next, disputed, 17)
	require.NoError(t, err)
	require.True(t, next.PaymentConditions[0].Disputed)
	require.True(t, next.PaymentChannels[0].Disputed)
}

func TestFinancialZoneStateRootIsCanonicalAndBuildsZoneRoot(t *testing.T) {
	left := FinancialZoneState{
		Accounts: []string{"bob", "alice"},
		Balances: []FinancialBalance{
			{Address: "bob", Denom: "naet", Amount: 2},
			{Address: "alice", Denom: "naet", Amount: 1},
		},
		FeeBuckets: []FinancialFeeBucket{{BucketID: "base", Denom: "naet", Amount: 3}},
		FactoryDenoms: []FinancialFactoryDenom{
			{Denom: "factory/alice/token", Authority: "alice", Supply: 4},
		},
		DEXPools: []FinancialDEXPool{{PoolID: 1, BaseDenom: "naet", QuoteDenom: "uatom", BaseReserve: 5, QuoteReserve: 6}},
		PaymentChannels: []FinancialPaymentChannel{
			{ChannelID: "channel-1", Payer: "alice", Receiver: "bob", Denom: "naet", EscrowAmount: 7},
		},
	}
	right := left
	right.Accounts = []string{"alice", "bob"}
	right.Balances = []FinancialBalance{
		{Address: "alice", Denom: "naet", Amount: 1},
		{Address: "bob", Denom: "naet", Amount: 2},
	}
	require.Equal(t, ComputeFinancialZoneStateRoot(left), ComputeFinancialZoneStateRoot(right))

	queues, err := NewZoneMessageQueues(ZoneIDFinancial, nil, nil)
	require.NoError(t, err)
	root, err := BuildFinancialZoneRootFromState(22, left, queues, EmptyRootHash())
	require.NoError(t, err)
	require.Equal(t, ZoneIDFinancial, root.ZoneID)
	require.Equal(t, ComputeFinancialZoneStateRoot(left), root.ZoneStateRoot)

	req, err := FinancialProofRequest(FinancialProofBalance, "alice/naet", 22, root.RootHash, 8)
	require.NoError(t, err)
	require.Equal(t, ZoneIDFinancial, req.ZoneID)
	require.Equal(t, "QueryBalance/alice/naet", req.Key)
}
