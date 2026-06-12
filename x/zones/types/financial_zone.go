package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const (
	FinancialZonePrefix	= "zone/financial"

	FinancialAccountsPrefix			= FinancialZonePrefix + "/accounts"
	FinancialBalancesPrefix			= FinancialZonePrefix + "/balances"
	FinancialFeeBucketPrefix		= FinancialZonePrefix + "/fees/buckets"
	FinancialShardFeeBucketPrefix		= FinancialZonePrefix + "/fees/shards"
	FinancialContractAssetDenomPrefix	= FinancialZonePrefix + "/contract-assets/denoms"
	FinancialContractAssetAuthorityPrefix	= FinancialZonePrefix + "/contract-assets/authority"
	FinancialDEXPoolPrefix			= FinancialZonePrefix + "/dex/pools"
	FinancialDEXOrderPrefix			= FinancialZonePrefix + "/dex/orders"
	FinancialPaymentChannelPrefix		= FinancialZonePrefix + "/payments/channels"
	FinancialPaymentConditionPrefix		= FinancialZonePrefix + "/payments/conditions"
	FinancialTransferEscrowPrefix		= FinancialZonePrefix + "/payments/escrow"
	FinancialMessageHandlerRoute		= FinancialZonePrefix + "/handler"
)

type FinancialZoneComponent string
type FinancialMessageKind string
type FinancialProofKind string
type FinancialShardRoutingMode string
type FinancialTransferEscrowStatus string

const (
	FinancialComponentBank			FinancialZoneComponent	= "bank"
	FinancialComponentFees			FinancialZoneComponent	= "fees"
	FinancialComponentDEX			FinancialZoneComponent	= "dex"
	FinancialComponentContractAssets	FinancialZoneComponent	= "contract_assets"
	FinancialComponentPayment		FinancialZoneComponent	= "payments"
	FinancialComponentMessageRouter		FinancialZoneComponent	= "message_router"
	FinancialComponentProofs		FinancialZoneComponent	= "proofs"

	FinancialMessageTransfer		FinancialMessageKind	= "MsgFinancialTransfer"
	FinancialMessageMintFactoryDenom	FinancialMessageKind	= "MsgMintFactoryDenom"
	FinancialMessageBurnFactoryDenom	FinancialMessageKind	= "MsgBurnFactoryDenom"
	FinancialMessageDexSwap			FinancialMessageKind	= "MsgDexSwap"
	FinancialMessageDexSettle		FinancialMessageKind	= "MsgDexSettle"
	FinancialMessagePaymentSettle		FinancialMessageKind	= "MsgPaymentSettle"
	FinancialMessagePaymentDispute		FinancialMessageKind	= "MsgPaymentDispute"

	FinancialProofBalance		FinancialProofKind	= "QueryBalance"
	FinancialProofBalancesByOwner	FinancialProofKind	= "QueryBalancesByOwner"
	FinancialProofFeeBucket		FinancialProofKind	= "QueryFeeBucket"
	FinancialProofFactoryDenom	FinancialProofKind	= "QueryFactoryDenom"
	FinancialProofPool		FinancialProofKind	= "QueryPool"
	FinancialProofPaymentChannel	FinancialProofKind	= "QueryPaymentChannel"
	FinancialProofPaymentCondition	FinancialProofKind	= "QueryPaymentCondition"

	FinancialRouteAccountAddress	FinancialShardRoutingMode	= "account-address"
	FinancialRoutePoolID		FinancialShardRoutingMode	= "pool-id"
	FinancialRoutePaymentChannel	FinancialShardRoutingMode	= "payment-channel"
	FinancialRouteExplicitGlobal	FinancialShardRoutingMode	= "explicit-global"

	FinancialEscrowDebited	FinancialTransferEscrowStatus	= "DEBITED"
	FinancialEscrowSettled	FinancialTransferEscrowStatus	= "SETTLED"
	FinancialEscrowRefunded	FinancialTransferEscrowStatus	= "REFUNDED"
)

type FinancialZoneKeeperBoundary struct {
	ZoneID		ZoneID
	OwnsPrefixes	[]string
	Components	[]FinancialZoneComponent
	MessageHandlers	[]FinancialMessageKind
	ProofKinds	[]FinancialProofKind
}

type FinancialBalance struct {
	Address	string
	Denom	string
	Amount	uint64
}

type FinancialFeeBucket struct {
	BucketID	string
	Denom		string
	Amount		uint64
}

type FinancialShardFeeBucket struct {
	ShardID	uint32
	Denom	string
	Amount	uint64
}

type FinancialFactoryDenom struct {
	Denom		string
	Authority	string
	Supply		uint64
}

type FinancialDEXPool struct {
	PoolID		uint64
	BaseDenom	string
	QuoteDenom	string
	BaseReserve	uint64
	QuoteReserve	uint64
}

type FinancialDEXOrder struct {
	OrderID		string
	PoolID		uint64
	Owner		string
	InputDenom	string
	OutputDenom	string
	InputAmount	uint64
	Status		string
}

type FinancialPaymentChannel struct {
	ChannelID	string
	Payer		string
	Receiver	string
	Denom		string
	EscrowAmount	uint64
	Finalized	bool
	Disputed	bool
	FinalizedHeight	uint64
}

type FinancialPaymentCondition struct {
	ConditionID	string
	ChannelID	string
	Amount		uint64
	Resolved	bool
	Disputed	bool
}

type FinancialTransferEscrow struct {
	TransferID		string
	FromAddress		string
	ToAddress		string
	Denom			string
	Amount			uint64
	LayoutEpoch		uint64
	SourceShardID		uint32
	ReceiverShardID		uint32
	Status			FinancialTransferEscrowStatus
	DebitReceiptHash	string
	CreditReceiptHash	string
	RefundReceiptHash	string
	EscrowHash		string
}

type FinancialShardRoute struct {
	ZoneID		ZoneID
	LayoutEpoch	uint64
	ShardCount	uint32
	ShardID		uint32
	RoutingMode	FinancialShardRoutingMode
	RouteKey	string
	StateKey	string
	RouteHash	string
}

type FinancialZoneState struct {
	Height			uint64
	Accounts		[]string
	Balances		[]FinancialBalance
	FeeBuckets		[]FinancialFeeBucket
	ShardFeeBuckets		[]FinancialShardFeeBucket
	FactoryDenoms		[]FinancialFactoryDenom
	DEXPools		[]FinancialDEXPool
	DEXOrders		[]FinancialDEXOrder
	PaymentChannels		[]FinancialPaymentChannel
	PaymentConditions	[]FinancialPaymentCondition
	TransferEscrows		[]FinancialTransferEscrow
	Receipts		[]ZoneReceipt
}

type FinancialZoneMessage struct {
	Kind		FinancialMessageKind
	AccountKey	string
	CounterpartyKey	string
	PoolID		uint64
	OrderID		string
	Denom		string
	OutputDenom	string
	BucketID	string
	ChannelID	string
	ConditionID	string
	Authority	string
	Amount		uint64
	PayloadHash	string
	Sequence	uint64
	GasLimit	uint64
}

type FinancialZoneRoots struct {
	Height				uint64
	AccountRoot			string
	BalanceRoot			string
	FeeBucketRoot			string
	ContractAssetRoot		string
	DEXRoot				string
	PaymentRoot			string
	InboxRoot			string
	OutboxRoot			string
	ReceiptRoot			string
	ExecutionRoot			string
	ProofRoot			string
	FinancialStateRoot		string
	DeprecatedBankRoot		string
	DeprecatedDEXPoolRoot		string
	DeprecatedSettlementRoot	string
}

func DefaultFinancialZoneKeeperBoundary() FinancialZoneKeeperBoundary {
	return FinancialZoneKeeperBoundary{
		ZoneID:	ZoneIDFinancial,
		OwnsPrefixes: []string{
			FinancialAccountsPrefix,
			FinancialBalancesPrefix,
			FinancialContractAssetAuthorityPrefix,
			FinancialContractAssetDenomPrefix,
			FinancialDEXOrderPrefix,
			FinancialDEXPoolPrefix,
			FinancialFeeBucketPrefix,
			FinancialShardFeeBucketPrefix,
			FinancialPaymentChannelPrefix,
			FinancialPaymentConditionPrefix,
			FinancialTransferEscrowPrefix,
		},
		Components: []FinancialZoneComponent{
			FinancialComponentBank,
			FinancialComponentFees,
			FinancialComponentDEX,
			FinancialComponentContractAssets,
			FinancialComponentPayment,
			FinancialComponentMessageRouter,
			FinancialComponentProofs,
		},
		MessageHandlers: []FinancialMessageKind{
			FinancialMessageTransfer,
			FinancialMessageMintFactoryDenom,
			FinancialMessageBurnFactoryDenom,
			FinancialMessageDexSwap,
			FinancialMessageDexSettle,
			FinancialMessagePaymentSettle,
			FinancialMessagePaymentDispute,
		},
		ProofKinds: []FinancialProofKind{
			FinancialProofBalance,
			FinancialProofBalancesByOwner,
			FinancialProofFeeBucket,
			FinancialProofFactoryDenom,
			FinancialProofPool,
			FinancialProofPaymentChannel,
			FinancialProofPaymentCondition,
		},
	}
}

func (b FinancialZoneKeeperBoundary) Validate() error {
	if b.ZoneID != ZoneIDFinancial {
		return errors.New("financial zone boundary must use FINANCIAL_ZONE")
	}
	if len(b.OwnsPrefixes) == 0 {
		return errors.New("financial zone boundary prefixes are required")
	}
	for i, prefix := range b.OwnsPrefixes {
		if err := validateRuntimeToken("financial zone prefix", prefix, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && b.OwnsPrefixes[i-1] >= prefix {
			return errors.New("financial zone prefixes must be sorted canonically")
		}
	}
	for _, handler := range b.MessageHandlers {
		if !IsFinancialMessageKind(handler) {
			return fmt.Errorf("unknown financial boundary message handler %q", handler)
		}
	}
	for _, proof := range b.ProofKinds {
		if !IsFinancialProofKind(proof) {
			return fmt.Errorf("unknown financial boundary proof kind %q", proof)
		}
	}
	if len(b.Components) == 0 || len(b.MessageHandlers) == 0 || len(b.ProofKinds) == 0 {
		return errors.New("financial zone boundary must declare components, handlers, and proof kinds")
	}
	return nil
}

func FinancialAccountKey(address string) (string, error) {
	if err := validateRuntimeToken("financial account address", address, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialAccountsPrefix + "/" + address, nil
}

func FinancialBankAccountKey(address string) (string, error) {
	return FinancialAccountKey(address)
}

func FinancialBalanceKey(address, denom string) (string, error) {
	if _, err := FinancialAccountKey(address); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("financial balance denom", denom, MaxZoneNamespaceLength); err != nil {
		return "", err
	}
	return FinancialBalancesPrefix + "/" + address + "/" + denom, nil
}

func FinancialBankTransferRoute(fromAccount string, toAccount string) (string, error) {
	fromKey, err := FinancialAccountKey(fromAccount)
	if err != nil {
		return "", err
	}
	toKey, err := FinancialAccountKey(toAccount)
	if err != nil {
		return "", err
	}
	return hashRuntimeParts("financial-bank-transfer-route-v1", fromKey, toKey), nil
}

func FinancialFeeBucketKey(bucketID string) (string, error) {
	if err := validateRuntimeToken("financial fee bucket", bucketID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialFeeBucketPrefix + "/" + bucketID, nil
}

func FinancialFactoryDenomKey(denom string) (string, error) {
	if err := validateRuntimeToken("financial contract asset denom", denom, MaxZoneNamespaceLength); err != nil {
		return "", err
	}
	return FinancialContractAssetDenomPrefix + "/" + hashRuntimeParts("financial-denom", denom), nil
}

func FinancialTokenAuthorityKey(denom string) (string, error) {
	if err := validateRuntimeToken("financial contract asset denom", denom, MaxZoneNamespaceLength); err != nil {
		return "", err
	}
	return FinancialContractAssetAuthorityPrefix + "/" + hashRuntimeParts("financial-denom-authority", denom), nil
}

func FinancialDEXPoolKey(poolID uint64) (string, error) {
	if poolID == 0 {
		return "", errors.New("financial dex pool id must be positive")
	}
	return fmt.Sprintf("%s/%020d", FinancialDEXPoolPrefix, poolID), nil
}

func FinancialDEXOrderKey(orderID string) (string, error) {
	if err := validateRuntimeToken("financial dex order id", orderID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialDEXOrderPrefix + "/" + orderID, nil
}

func FinancialPaymentChannelKey(channelID string) (string, error) {
	if err := validateRuntimeToken("financial payment channel id", channelID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialPaymentChannelPrefix + "/" + channelID, nil
}

func FinancialPaymentConditionKey(conditionID string) (string, error) {
	if err := validateRuntimeToken("financial payment condition id", conditionID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialPaymentConditionPrefix + "/" + conditionID, nil
}

func FinancialTransferEscrowKey(transferID string) (string, error) {
	if err := validateRuntimeToken("financial transfer escrow id", transferID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return FinancialTransferEscrowPrefix + "/" + transferID, nil
}

func FinancialShardFeeBucketKey(shardID uint32, denom string) (string, error) {
	if err := validateRuntimeToken("financial shard fee denom", denom, MaxZoneNamespaceLength); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%010d/%s", FinancialShardFeeBucketPrefix, shardID, denom), nil
}

func FinancialPaymentSettlementKey(channelID string) (string, error) {
	return FinancialPaymentChannelKey(channelID)
}

func RouteFinancialAccountShard(address string, shardCount uint32, layoutEpoch uint64) (FinancialShardRoute, error) {
	key, err := FinancialAccountKey(address)
	if err != nil {
		return FinancialShardRoute{}, err
	}
	return routeFinancialStateKey(FinancialRouteAccountAddress, key, address, shardCount, layoutEpoch)
}

func RouteFinancialBalanceShard(address, denom string, shardCount uint32, layoutEpoch uint64) (FinancialShardRoute, error) {
	key, err := FinancialBalanceKey(address, denom)
	if err != nil {
		return FinancialShardRoute{}, err
	}
	return routeFinancialStateKey(FinancialRouteAccountAddress, key, address, shardCount, layoutEpoch)
}

func RouteFinancialDEXPoolShard(poolID uint64, shardCount uint32, layoutEpoch uint64) (FinancialShardRoute, error) {
	key, err := FinancialDEXPoolKey(poolID)
	if err != nil {
		return FinancialShardRoute{}, err
	}
	return routeFinancialStateKey(FinancialRoutePoolID, key, fmt.Sprintf("%020d", poolID), shardCount, layoutEpoch)
}

func RouteFinancialPaymentChannelShard(channelID string, shardCount uint32, layoutEpoch uint64) (FinancialShardRoute, error) {
	key, err := FinancialPaymentChannelKey(channelID)
	if err != nil {
		return FinancialShardRoute{}, err
	}
	return routeFinancialStateKey(FinancialRoutePaymentChannel, key, channelID, shardCount, layoutEpoch)
}

func (m FinancialZoneMessage) Validate() error {
	if !IsFinancialMessageKind(m.Kind) {
		return fmt.Errorf("unknown financial message kind %q", m.Kind)
	}
	if m.Amount == 0 {
		return errors.New("financial message amount must be positive")
	}
	if err := ValidateHash("financial message payload hash", m.PayloadHash); err != nil {
		return err
	}
	switch m.Kind {
	case FinancialMessageTransfer:
		if _, err := FinancialBalanceKey(m.AccountKey, m.Denom); err != nil {
			return err
		}
		_, err := FinancialBalanceKey(m.CounterpartyKey, m.Denom)
		return err
	case FinancialMessageMintFactoryDenom:
		if _, err := FinancialFactoryDenomKey(m.Denom); err != nil {
			return err
		}
		_, err := FinancialBalanceKey(m.AccountKey, m.Denom)
		return err
	case FinancialMessageBurnFactoryDenom:
		if _, err := FinancialFactoryDenomKey(m.Denom); err != nil {
			return err
		}
		_, err := FinancialBalanceKey(m.AccountKey, m.Denom)
		return err
	case FinancialMessageDexSwap:
		if _, err := FinancialDEXPoolKey(m.PoolID); err != nil {
			return err
		}
		_, err := FinancialDEXOrderKey(m.OrderID)
		return err
	case FinancialMessageDexSettle:
		_, err := FinancialDEXOrderKey(m.OrderID)
		return err
	case FinancialMessagePaymentSettle:
		_, err := FinancialPaymentChannelKey(m.ChannelID)
		return err
	case FinancialMessagePaymentDispute:
		_, err := FinancialPaymentConditionKey(m.ConditionID)
		return err
	default:
		return nil
	}
}

func (m FinancialZoneMessage) ZoneMessage() (ZoneMessage, error) {
	if err := m.Validate(); err != nil {
		return ZoneMessage{}, err
	}
	gasLimit := m.GasLimit
	if gasLimit == 0 {
		gasLimit = 1
	}
	return ZoneMessage{
		ZoneID:		ZoneIDFinancial,
		MessageType:	string(m.Kind),
		Source:		FinancialMessageHandlerRoute,
		Destination:	FinancialMessageHandlerRoute + "/" + string(m.Kind),
		GasLimit:	gasLimit,
		PayloadHash:	m.PayloadHash,
		Sequence:	m.Sequence,
	}, nil
}

func FinancialMessageHandler(kind FinancialMessageKind) (string, error) {
	if !IsFinancialMessageKind(kind) {
		return "", fmt.Errorf("unknown financial message kind %q", kind)
	}
	return FinancialMessageHandlerRoute + "/" + string(kind), nil
}

func FinancialProofRequest(kind FinancialProofKind, key string, height uint64, root string, limit uint32) (ZoneProofRequest, error) {
	if !IsFinancialProofKind(kind) {
		return ZoneProofRequest{}, fmt.Errorf("unknown financial proof kind %q", kind)
	}
	if err := validateRuntimeToken("financial proof key", key, MaxZoneProofKeyLength); err != nil {
		return ZoneProofRequest{}, err
	}
	req := ZoneProofRequest{
		ZoneID:	ZoneIDFinancial,
		Height:	height,
		Kind:	ZoneProofKindState,
		Key:	string(kind) + "/" + key,
		Root:	root,
		Limit:	limit,
	}
	return req, req.Validate()
}

func ApplyFinancialMessage(state FinancialZoneState, msg FinancialZoneMessage, height uint64) (FinancialZoneState, ZoneReceipt, error) {
	if height == 0 {
		return FinancialZoneState{}, ZoneReceipt{}, errors.New("financial message height must be positive")
	}
	if err := msg.Validate(); err != nil {
		return FinancialZoneState{}, ZoneReceipt{}, err
	}
	next := state.Normalize()
	next.Height = height
	var err error
	switch msg.Kind {
	case FinancialMessageTransfer:
		next, err = applyFinancialTransfer(next, msg.AccountKey, msg.CounterpartyKey, msg.Denom, msg.Amount)
	case FinancialMessageMintFactoryDenom:
		next, err = mintFinancialFactoryDenom(next, msg.Denom, msg.Authority, msg.AccountKey, msg.Amount)
	case FinancialMessageBurnFactoryDenom:
		next, err = burnFinancialFactoryDenom(next, msg.Denom, msg.AccountKey, msg.Amount)
	case FinancialMessageDexSwap:
		next, err = applyFinancialDEXSwap(next, msg)
	case FinancialMessageDexSettle:
		next, err = settleFinancialDEXOrder(next, msg.OrderID)
	case FinancialMessagePaymentSettle:
		next, err = finalizeFinancialPaymentChannel(next, msg.ChannelID, height)
	case FinancialMessagePaymentDispute:
		next, err = disputeFinancialPaymentCondition(next, msg.ConditionID)
	default:
		err = fmt.Errorf("unknown financial message kind %q", msg.Kind)
	}
	if err != nil {
		return FinancialZoneState{}, ZoneReceipt{}, err
	}
	receipt, err := NewZoneReceipt(ZoneReceipt{
		ZoneID:		ZoneIDFinancial,
		Height:		height,
		ItemHash:	msg.PayloadHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	msg.GasLimit,
		ResultHash:	ComputeFinancialZoneStateRoot(next),
		Sequence:	msg.Sequence,
	})
	if err != nil {
		return FinancialZoneState{}, ZoneReceipt{}, err
	}
	next.Receipts = append(next.Receipts, receipt)
	return next.Normalize(), receipt, nil
}

func CreditFinancialFeeBucket(state FinancialZoneState, bucketID, denom string, amount uint64) (FinancialZoneState, error) {
	if amount == 0 {
		return FinancialZoneState{}, errors.New("financial fee bucket credit amount must be positive")
	}
	if _, err := FinancialFeeBucketKey(bucketID); err != nil {
		return FinancialZoneState{}, err
	}
	if err := validateRuntimeToken("financial fee bucket denom", denom, MaxZoneNamespaceLength); err != nil {
		return FinancialZoneState{}, err
	}
	next := state.Normalize()
	for i, bucket := range next.FeeBuckets {
		if bucket.BucketID == bucketID && bucket.Denom == denom {
			sum, err := checkedAdd(bucket.Amount, amount, "financial fee bucket amount")
			if err != nil {
				return FinancialZoneState{}, err
			}
			next.FeeBuckets[i].Amount = sum
			return next.Normalize(), nil
		}
	}
	next.FeeBuckets = append(next.FeeBuckets, FinancialFeeBucket{BucketID: bucketID, Denom: denom, Amount: amount})
	return next.Normalize(), nil
}

func CreditFinancialShardFeeBucket(state FinancialZoneState, shardID uint32, denom string, amount uint64) (FinancialZoneState, error) {
	if amount == 0 {
		return FinancialZoneState{}, errors.New("financial shard fee credit amount must be positive")
	}
	if _, err := FinancialShardFeeBucketKey(shardID, denom); err != nil {
		return FinancialZoneState{}, err
	}
	next := state.Normalize()
	for i, bucket := range next.ShardFeeBuckets {
		if bucket.ShardID == shardID && bucket.Denom == denom {
			sum, err := checkedAdd(bucket.Amount, amount, "financial shard fee bucket amount")
			if err != nil {
				return FinancialZoneState{}, err
			}
			next.ShardFeeBuckets[i].Amount = sum
			return next.Normalize(), nil
		}
	}
	next.ShardFeeBuckets = append(next.ShardFeeBuckets, FinancialShardFeeBucket{ShardID: shardID, Denom: denom, Amount: amount})
	return next.Normalize(), nil
}

func AggregateFinancialShardFees(state FinancialZoneState, aggregateBucketID string) (FinancialZoneState, string, error) {
	if _, err := FinancialFeeBucketKey(aggregateBucketID); err != nil {
		return FinancialZoneState{}, "", err
	}
	next := state.Normalize()
	for _, bucket := range next.ShardFeeBuckets {
		updated, err := CreditFinancialFeeBucket(next, aggregateBucketID, bucket.Denom, bucket.Amount)
		if err != nil {
			return FinancialZoneState{}, "", err
		}
		next = updated
	}
	root := ComputeFinancialFeeBucketRoot(next.FeeBuckets)
	return next.Normalize(), root, nil
}

func OpenFinancialCrossShardTransferEscrow(state FinancialZoneState, from, to, denom string, amount uint64, shardCount uint32, layoutEpoch uint64, nonce uint64) (FinancialZoneState, FinancialTransferEscrow, error) {
	if nonce == 0 {
		return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow nonce must be positive")
	}
	source, err := RouteFinancialBalanceShard(from, denom, shardCount, layoutEpoch)
	if err != nil {
		return FinancialZoneState{}, FinancialTransferEscrow{}, err
	}
	receiver, err := RouteFinancialBalanceShard(to, denom, shardCount, layoutEpoch)
	if err != nil {
		return FinancialZoneState{}, FinancialTransferEscrow{}, err
	}
	if source.ShardID == receiver.ShardID {
		return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial cross-shard transfer requires distinct shards")
	}
	next, err := adjustFinancialBalance(state, from, denom, amount, false)
	if err != nil {
		return FinancialZoneState{}, FinancialTransferEscrow{}, err
	}
	transferID := hashRuntimeParts("financial-cross-shard-transfer-id-v1", from, to, denom, fmt.Sprint(amount), fmt.Sprint(layoutEpoch), fmt.Sprint(nonce))
	receiptHash := hashRuntimeParts("financial-cross-shard-debit-receipt-v1", transferID, fmt.Sprint(source.ShardID), fmt.Sprint(receiver.ShardID))
	escrow := FinancialTransferEscrow{
		TransferID:		transferID,
		FromAddress:		from,
		ToAddress:		to,
		Denom:			denom,
		Amount:			amount,
		LayoutEpoch:		layoutEpoch,
		SourceShardID:		source.ShardID,
		ReceiverShardID:	receiver.ShardID,
		Status:			FinancialEscrowDebited,
		DebitReceiptHash:	receiptHash,
	}
	escrow.EscrowHash = ComputeFinancialTransferEscrowHash(escrow)
	if err := escrow.ValidateHash(); err != nil {
		return FinancialZoneState{}, FinancialTransferEscrow{}, err
	}
	next.TransferEscrows = append(next.TransferEscrows, escrow)
	return next.Normalize(), escrow, nil
}

func SettleFinancialCrossShardTransferEscrow(state FinancialZoneState, transferID string, committedDebitReceiptHash string) (FinancialZoneState, FinancialTransferEscrow, error) {
	next := state.Normalize()
	for i, escrow := range next.TransferEscrows {
		if escrow.TransferID != transferID {
			continue
		}
		if err := escrow.ValidateHash(); err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		if escrow.Status != FinancialEscrowDebited {
			return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow is not debited")
		}
		if escrow.DebitReceiptHash != committedDebitReceiptHash {
			return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow debit receipt mismatch")
		}
		credited, err := adjustFinancialBalance(next, escrow.ToAddress, escrow.Denom, escrow.Amount, true)
		if err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		next = credited.Normalize()
		escrow.Status = FinancialEscrowSettled
		escrow.CreditReceiptHash = hashRuntimeParts("financial-cross-shard-credit-receipt-v1", escrow.TransferID, committedDebitReceiptHash)
		escrow.EscrowHash = ComputeFinancialTransferEscrowHash(escrow)
		if err := escrow.ValidateHash(); err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		next.TransferEscrows[i] = escrow
		return next.Normalize(), escrow, nil
	}
	return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow not found")
}

func RefundFinancialCrossShardTransferEscrow(state FinancialZoneState, transferID string, committedDebitReceiptHash string) (FinancialZoneState, FinancialTransferEscrow, error) {
	next := state.Normalize()
	for i, escrow := range next.TransferEscrows {
		if escrow.TransferID != transferID {
			continue
		}
		if err := escrow.ValidateHash(); err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		if escrow.Status != FinancialEscrowDebited {
			return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow is not debited")
		}
		if escrow.DebitReceiptHash != committedDebitReceiptHash {
			return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow debit receipt mismatch")
		}
		refunded, err := adjustFinancialBalance(next, escrow.FromAddress, escrow.Denom, escrow.Amount, true)
		if err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		next = refunded.Normalize()
		escrow.Status = FinancialEscrowRefunded
		escrow.RefundReceiptHash = hashRuntimeParts("financial-cross-shard-refund-receipt-v1", escrow.TransferID, committedDebitReceiptHash)
		escrow.EscrowHash = ComputeFinancialTransferEscrowHash(escrow)
		if err := escrow.ValidateHash(); err != nil {
			return FinancialZoneState{}, FinancialTransferEscrow{}, err
		}
		next.TransferEscrows[i] = escrow
		return next.Normalize(), escrow, nil
	}
	return FinancialZoneState{}, FinancialTransferEscrow{}, errors.New("financial transfer escrow not found")
}

func BuildFinancialDEXSettlementReceipt(height uint64, order FinancialDEXOrder, resultHash string, sequence uint64) (ZoneReceipt, error) {
	if height == 0 {
		return ZoneReceipt{}, errors.New("financial dex receipt height must be positive")
	}
	if err := order.Validate(); err != nil {
		return ZoneReceipt{}, err
	}
	if err := ValidateHash("financial dex settlement result hash", resultHash); err != nil {
		return ZoneReceipt{}, err
	}
	itemHash := hashRuntimeParts("financial-dex-settlement-receipt-v1", order.OrderID, fmt.Sprint(order.PoolID), order.Status)
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:		ZoneIDFinancial,
		Height:		height,
		ItemHash:	itemHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	0,
		ResultHash:	resultHash,
		Sequence:	sequence,
	})
}

func BuildFinancialPaymentFinalizationHook(height uint64, channel FinancialPaymentChannel, resultHash string, sequence uint64) (ZoneReceipt, error) {
	if height == 0 {
		return ZoneReceipt{}, errors.New("financial payment hook height must be positive")
	}
	if err := channel.Validate(); err != nil {
		return ZoneReceipt{}, err
	}
	if !channel.Finalized {
		return ZoneReceipt{}, errors.New("financial payment finalization hook requires finalized channel")
	}
	if err := ValidateHash("financial payment finalization result hash", resultHash); err != nil {
		return ZoneReceipt{}, err
	}
	itemHash := hashRuntimeParts("financial-payment-finalization-hook-v1", channel.ChannelID, fmt.Sprint(channel.FinalizedHeight))
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:		ZoneIDFinancial,
		Height:		height,
		ItemHash:	itemHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	0,
		ResultHash:	resultHash,
		Sequence:	sequence,
	})
}

func BuildFinancialZoneRoot(roots FinancialZoneRoots) (ZoneRoot, error) {
	if err := roots.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	stateRoot := roots.FinancialStateRoot
	if stateRoot == "" {
		stateRoot = hashRuntimeParts(
			"aetra-financial-zone-state-v2",
			roots.AccountRoot,
			roots.BalanceRoot,
			roots.FeeBucketRoot,
			roots.ContractAssetRoot,
			roots.DEXRoot,
			roots.PaymentRoot,
		)
	}
	root := ZoneRoot{
		ZoneID:			ZoneIDFinancial,
		Height:			roots.Height,
		ZoneStateRoot:		stateRoot,
		InboxRoot:		roots.InboxRoot,
		OutboxRoot:		roots.OutboxRoot,
		ReceiptRoot:		roots.ReceiptRoot,
		EventRoot:		EmptyRootHash(),
		ExecutionResultRoot:	roots.ExecutionRoot,
		ProofRoot:		roots.ProofRoot,
	}
	root.RootHash = ComputeZoneRootHash(root)
	return root, root.Validate()
}

func BuildFinancialZoneRootFromState(height uint64, state FinancialZoneState, queues ZoneMessageQueues, proofRoot string) (ZoneRoot, error) {
	if height == 0 {
		return ZoneRoot{}, errors.New("financial zone root height must be positive")
	}
	if err := queues.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	if queues.ZoneID != ZoneIDFinancial {
		return ZoneRoot{}, errors.New("financial zone root queue route mismatch")
	}
	normalized := state.Normalize()
	roots := FinancialZoneRoots{
		Height:			height,
		AccountRoot:		ComputeFinancialAccountsRoot(normalized.Accounts),
		BalanceRoot:		ComputeFinancialBalancesRoot(normalized.Balances),
		FeeBucketRoot:		ComputeFinancialFeeBucketRoot(normalized.FeeBuckets),
		ContractAssetRoot:	ComputeFinancialContractAssetRoot(normalized.FactoryDenoms),
		DEXRoot:		ComputeFinancialDEXRoot(normalized.DEXPools, normalized.DEXOrders),
		PaymentRoot:		ComputeFinancialPaymentRoot(normalized.PaymentChannels, normalized.PaymentConditions, normalized.TransferEscrows),
		InboxRoot:		queues.InboxRoot(),
		OutboxRoot:		queues.OutboxRoot(),
		ReceiptRoot:		ComputeZoneReceiptRoot(normalized.Receipts),
		ExecutionRoot:		ComputeZoneExecutionResultRoot(normalized.Receipts),
		ProofRoot:		proofRoot,
		FinancialStateRoot:	ComputeFinancialZoneStateRoot(normalized),
	}
	return BuildFinancialZoneRoot(roots)
}

func (r FinancialZoneRoots) Validate() error {
	if r.Height == 0 {
		return errors.New("financial zone root height must be positive")
	}
	if r.AccountRoot == "" && r.DeprecatedBankRoot != "" {
		r.AccountRoot = r.DeprecatedBankRoot
	}
	if r.DEXRoot == "" && r.DeprecatedDEXPoolRoot != "" {
		r.DEXRoot = r.DeprecatedDEXPoolRoot
	}
	if r.PaymentRoot == "" && r.DeprecatedSettlementRoot != "" {
		r.PaymentRoot = r.DeprecatedSettlementRoot
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "financial account root", value: r.AccountRoot},
		{name: "financial balance root", value: r.BalanceRoot},
		{name: "financial fee bucket root", value: r.FeeBucketRoot},
		{name: "financial contract asset root", value: r.ContractAssetRoot},
		{name: "financial dex root", value: r.DEXRoot},
		{name: "financial payment root", value: r.PaymentRoot},
		{name: "financial inbox root", value: r.InboxRoot},
		{name: "financial outbox root", value: r.OutboxRoot},
		{name: "financial receipt root", value: r.ReceiptRoot},
		{name: "financial execution root", value: r.ExecutionRoot},
		{name: "financial proof root", value: r.ProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.FinancialStateRoot != "" {
		return ValidateHash("financial state root", r.FinancialStateRoot)
	}
	return nil
}

func (s FinancialZoneState) Normalize() FinancialZoneState {
	s.Accounts = normalizeFinancialAccounts(s.Accounts)
	s.Balances = normalizeFinancialBalances(s.Balances)
	s.FeeBuckets = normalizeFinancialFeeBuckets(s.FeeBuckets)
	s.ShardFeeBuckets = normalizeFinancialShardFeeBuckets(s.ShardFeeBuckets)
	s.FactoryDenoms = normalizeFinancialFactoryDenoms(s.FactoryDenoms)
	s.DEXPools = normalizeFinancialDEXPools(s.DEXPools)
	s.DEXOrders = normalizeFinancialDEXOrders(s.DEXOrders)
	s.PaymentChannels = normalizeFinancialPaymentChannels(s.PaymentChannels)
	s.PaymentConditions = normalizeFinancialPaymentConditions(s.PaymentConditions)
	s.TransferEscrows = normalizeFinancialTransferEscrows(s.TransferEscrows)
	s.Receipts = cloneZoneReceipts(s.Receipts)
	return s
}

func (s FinancialZoneState) Validate() error {
	normalized := s.Normalize()
	for i, account := range normalized.Accounts {
		if _, err := FinancialAccountKey(account); err != nil {
			return err
		}
		if i > 0 && normalized.Accounts[i-1] >= account {
			return errors.New("financial accounts must be sorted canonically")
		}
	}
	for _, balance := range normalized.Balances {
		if err := balance.Validate(); err != nil {
			return err
		}
	}
	for _, bucket := range normalized.FeeBuckets {
		if err := bucket.Validate(); err != nil {
			return err
		}
	}
	for _, bucket := range normalized.ShardFeeBuckets {
		if err := bucket.Validate(); err != nil {
			return err
		}
	}
	for _, denom := range normalized.FactoryDenoms {
		if err := denom.Validate(); err != nil {
			return err
		}
	}
	for _, pool := range normalized.DEXPools {
		if err := pool.Validate(); err != nil {
			return err
		}
	}
	for _, order := range normalized.DEXOrders {
		if err := order.Validate(); err != nil {
			return err
		}
	}
	for _, channel := range normalized.PaymentChannels {
		if err := channel.Validate(); err != nil {
			return err
		}
	}
	for _, condition := range normalized.PaymentConditions {
		if err := condition.Validate(); err != nil {
			return err
		}
	}
	for _, escrow := range normalized.TransferEscrows {
		if err := escrow.ValidateHash(); err != nil {
			return err
		}
	}
	for _, receipt := range normalized.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if receipt.ZoneID != ZoneIDFinancial {
			return errors.New("financial receipt route mismatch")
		}
	}
	return nil
}

func (b FinancialBalance) Validate() error {
	_, err := FinancialBalanceKey(b.Address, b.Denom)
	return err
}

func (b FinancialFeeBucket) Validate() error {
	if _, err := FinancialFeeBucketKey(b.BucketID); err != nil {
		return err
	}
	return validateRuntimeToken("financial fee bucket denom", b.Denom, MaxZoneNamespaceLength)
}

func (b FinancialShardFeeBucket) Validate() error {
	_, err := FinancialShardFeeBucketKey(b.ShardID, b.Denom)
	return err
}

func (d FinancialFactoryDenom) Validate() error {
	if _, err := FinancialFactoryDenomKey(d.Denom); err != nil {
		return err
	}
	return validateRuntimeToken("financial contract asset authority", d.Authority, MaxZoneEndpointLength)
}

func (p FinancialDEXPool) Validate() error {
	if _, err := FinancialDEXPoolKey(p.PoolID); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial dex base denom", p.BaseDenom, MaxZoneNamespaceLength); err != nil {
		return err
	}
	return validateRuntimeToken("financial dex quote denom", p.QuoteDenom, MaxZoneNamespaceLength)
}

func (o FinancialDEXOrder) Validate() error {
	if _, err := FinancialDEXOrderKey(o.OrderID); err != nil {
		return err
	}
	if _, err := FinancialDEXPoolKey(o.PoolID); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial dex order owner", o.Owner, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial dex input denom", o.InputDenom, MaxZoneNamespaceLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial dex output denom", o.OutputDenom, MaxZoneNamespaceLength); err != nil {
		return err
	}
	return validateRuntimeToken("financial dex order status", o.Status, MaxZoneEndpointLength)
}

func (c FinancialPaymentChannel) Validate() error {
	if _, err := FinancialPaymentChannelKey(c.ChannelID); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial payment payer", c.Payer, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial payment receiver", c.Receiver, MaxZoneEndpointLength); err != nil {
		return err
	}
	return validateRuntimeToken("financial payment denom", c.Denom, MaxZoneNamespaceLength)
}

func (c FinancialPaymentCondition) Validate() error {
	if _, err := FinancialPaymentConditionKey(c.ConditionID); err != nil {
		return err
	}
	if _, err := FinancialPaymentChannelKey(c.ChannelID); err != nil {
		return err
	}
	return nil
}

func (e FinancialTransferEscrow) ValidateFormat() error {
	if _, err := FinancialTransferEscrowKey(e.TransferID); err != nil {
		return err
	}
	if _, err := FinancialBalanceKey(e.FromAddress, e.Denom); err != nil {
		return err
	}
	if _, err := FinancialBalanceKey(e.ToAddress, e.Denom); err != nil {
		return err
	}
	if e.Amount == 0 {
		return errors.New("financial transfer escrow amount must be positive")
	}
	if e.LayoutEpoch == 0 {
		return errors.New("financial transfer escrow layout epoch must be positive")
	}
	if e.SourceShardID == e.ReceiverShardID {
		return errors.New("financial transfer escrow requires distinct shards")
	}
	switch e.Status {
	case FinancialEscrowDebited:
		if e.CreditReceiptHash != "" || e.RefundReceiptHash != "" {
			return errors.New("financial debited escrow must not include terminal receipts")
		}
	case FinancialEscrowSettled:
		if e.CreditReceiptHash == "" {
			return errors.New("financial settled escrow requires credit receipt")
		}
	case FinancialEscrowRefunded:
		if e.RefundReceiptHash == "" {
			return errors.New("financial refunded escrow requires refund receipt")
		}
	default:
		return fmt.Errorf("unknown financial transfer escrow status %q", e.Status)
	}
	if err := ValidateHash("financial transfer escrow debit receipt", e.DebitReceiptHash); err != nil {
		return err
	}
	if e.CreditReceiptHash != "" {
		if err := ValidateHash("financial transfer escrow credit receipt", e.CreditReceiptHash); err != nil {
			return err
		}
	}
	if e.RefundReceiptHash != "" {
		if err := ValidateHash("financial transfer escrow refund receipt", e.RefundReceiptHash); err != nil {
			return err
		}
	}
	if e.EscrowHash != "" {
		return ValidateHash("financial transfer escrow hash", e.EscrowHash)
	}
	return nil
}

func (e FinancialTransferEscrow) ValidateHash() error {
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeFinancialTransferEscrowHash(e)
	if e.EscrowHash != expected {
		return errors.New("financial transfer escrow hash mismatch")
	}
	return nil
}

func ComputeFinancialZoneStateRoot(state FinancialZoneState) string {
	normalized := state.Normalize()
	return hashRuntimeParts(
		"aetra-financial-zone-state-v2",
		ComputeFinancialAccountsRoot(normalized.Accounts),
		ComputeFinancialBalancesRoot(normalized.Balances),
		ComputeFinancialFeeBucketRoot(normalized.FeeBuckets),
		ComputeFinancialContractAssetRoot(normalized.FactoryDenoms),
		ComputeFinancialDEXRoot(normalized.DEXPools, normalized.DEXOrders),
		ComputeFinancialPaymentRoot(normalized.PaymentChannels, normalized.PaymentConditions, normalized.TransferEscrows),
	)
}

func ComputeFinancialAccountsRoot(accounts []string) string {
	ordered := normalizeFinancialAccounts(accounts)
	parts := []string{"aetra-financial-accounts-root-v1", fmt.Sprint(len(ordered))}
	parts = append(parts, ordered...)
	return hashRuntimeParts(parts...)
}

func ComputeFinancialBalancesRoot(balances []FinancialBalance) string {
	ordered := normalizeFinancialBalances(balances)
	parts := []string{"aetra-financial-balances-root-v1", fmt.Sprint(len(ordered))}
	for _, balance := range ordered {
		parts = append(parts, balance.Address, balance.Denom, fmt.Sprint(balance.Amount))
	}
	return hashRuntimeParts(parts...)
}

func ComputeFinancialFeeBucketRoot(buckets []FinancialFeeBucket) string {
	ordered := normalizeFinancialFeeBuckets(buckets)
	parts := []string{"aetra-financial-fee-bucket-root-v1", fmt.Sprint(len(ordered))}
	for _, bucket := range ordered {
		parts = append(parts, bucket.BucketID, bucket.Denom, fmt.Sprint(bucket.Amount))
	}
	return hashRuntimeParts(parts...)
}

func ComputeFinancialContractAssetRoot(denoms []FinancialFactoryDenom) string {
	ordered := normalizeFinancialFactoryDenoms(denoms)
	parts := []string{"aetra-financial-contract-asset-root-v1", fmt.Sprint(len(ordered))}
	for _, denom := range ordered {
		parts = append(parts, denom.Denom, denom.Authority, fmt.Sprint(denom.Supply))
	}
	return hashRuntimeParts(parts...)
}

func ComputeFinancialDEXRoot(pools []FinancialDEXPool, orders []FinancialDEXOrder) string {
	orderedPools := normalizeFinancialDEXPools(pools)
	orderedOrders := normalizeFinancialDEXOrders(orders)
	parts := []string{"aetra-financial-dex-root-v1", fmt.Sprint(len(orderedPools))}
	for _, pool := range orderedPools {
		parts = append(parts, fmt.Sprint(pool.PoolID), pool.BaseDenom, pool.QuoteDenom, fmt.Sprint(pool.BaseReserve), fmt.Sprint(pool.QuoteReserve))
	}
	parts = append(parts, fmt.Sprint(len(orderedOrders)))
	for _, order := range orderedOrders {
		parts = append(parts, order.OrderID, fmt.Sprint(order.PoolID), order.Owner, order.InputDenom, order.OutputDenom, fmt.Sprint(order.InputAmount), order.Status)
	}
	return hashRuntimeParts(parts...)
}

func ComputeFinancialPaymentRoot(channels []FinancialPaymentChannel, conditions []FinancialPaymentCondition, escrows []FinancialTransferEscrow) string {
	orderedChannels := normalizeFinancialPaymentChannels(channels)
	orderedConditions := normalizeFinancialPaymentConditions(conditions)
	orderedEscrows := normalizeFinancialTransferEscrows(escrows)
	parts := []string{"aetra-financial-payment-root-v1", fmt.Sprint(len(orderedChannels))}
	for _, channel := range orderedChannels {
		parts = append(parts, channel.ChannelID, channel.Payer, channel.Receiver, channel.Denom, fmt.Sprint(channel.EscrowAmount), fmt.Sprint(channel.Finalized), fmt.Sprint(channel.Disputed), fmt.Sprint(channel.FinalizedHeight))
	}
	parts = append(parts, fmt.Sprint(len(orderedConditions)))
	for _, condition := range orderedConditions {
		parts = append(parts, condition.ConditionID, condition.ChannelID, fmt.Sprint(condition.Amount), fmt.Sprint(condition.Resolved), fmt.Sprint(condition.Disputed))
	}
	parts = append(parts, fmt.Sprint(len(orderedEscrows)))
	for _, escrow := range orderedEscrows {
		parts = append(parts, escrow.EscrowHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeFinancialTransferEscrowHash(escrow FinancialTransferEscrow) string {
	return hashRuntimeParts(
		"aetra-financial-transfer-escrow-v1",
		escrow.TransferID,
		escrow.FromAddress,
		escrow.ToAddress,
		escrow.Denom,
		fmt.Sprint(escrow.Amount),
		fmt.Sprint(escrow.LayoutEpoch),
		fmt.Sprint(escrow.SourceShardID),
		fmt.Sprint(escrow.ReceiverShardID),
		string(escrow.Status),
		escrow.DebitReceiptHash,
		escrow.CreditReceiptHash,
		escrow.RefundReceiptHash,
	)
}

func IsFinancialMessageKind(kind FinancialMessageKind) bool {
	switch kind {
	case FinancialMessageTransfer,
		FinancialMessageMintFactoryDenom,
		FinancialMessageBurnFactoryDenom,
		FinancialMessageDexSwap,
		FinancialMessageDexSettle,
		FinancialMessagePaymentSettle,
		FinancialMessagePaymentDispute:
		return true
	default:
		return false
	}
}

func IsFinancialProofKind(kind FinancialProofKind) bool {
	switch kind {
	case FinancialProofBalance,
		FinancialProofBalancesByOwner,
		FinancialProofFeeBucket,
		FinancialProofFactoryDenom,
		FinancialProofPool,
		FinancialProofPaymentChannel,
		FinancialProofPaymentCondition:
		return true
	default:
		return false
	}
}

func (r FinancialShardRoute) ValidateHash() error {
	if r.ZoneID != ZoneIDFinancial {
		return errors.New("financial shard route must use FINANCIAL_ZONE")
	}
	if r.LayoutEpoch == 0 {
		return errors.New("financial shard route layout epoch must be positive")
	}
	if r.ShardCount == 0 {
		return errors.New("financial shard route count must be positive")
	}
	if r.ShardID >= r.ShardCount {
		return errors.New("financial shard route shard id out of range")
	}
	switch r.RoutingMode {
	case FinancialRouteAccountAddress, FinancialRoutePoolID, FinancialRoutePaymentChannel, FinancialRouteExplicitGlobal:
	default:
		return fmt.Errorf("unknown financial shard routing mode %q", r.RoutingMode)
	}
	if err := validateRuntimeToken("financial shard route key", r.RouteKey, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("financial shard route state key", r.StateKey, MaxZoneNamespaceLength); err != nil {
		return err
	}
	if err := ValidateHash("financial shard route hash", r.RouteHash); err != nil {
		return err
	}
	if expected := ComputeFinancialShardRouteHash(r); r.RouteHash != expected {
		return errors.New("financial shard route hash mismatch")
	}
	return nil
}

func ComputeFinancialShardRouteHash(route FinancialShardRoute) string {
	return hashRuntimeParts(
		"aetra-financial-shard-route-v1",
		string(route.ZoneID),
		fmt.Sprint(route.LayoutEpoch),
		fmt.Sprint(route.ShardCount),
		fmt.Sprint(route.ShardID),
		string(route.RoutingMode),
		route.RouteKey,
		route.StateKey,
	)
}

func applyFinancialTransfer(state FinancialZoneState, from, to, denom string, amount uint64) (FinancialZoneState, error) {
	next, err := adjustFinancialBalance(state, from, denom, amount, false)
	if err != nil {
		return FinancialZoneState{}, err
	}
	return adjustFinancialBalance(next, to, denom, amount, true)
}

func mintFinancialFactoryDenom(state FinancialZoneState, denom, authority, receiver string, amount uint64) (FinancialZoneState, error) {
	if authority == "" {
		return FinancialZoneState{}, errors.New("financial contract asset authority is required")
	}
	next := state.Normalize()
	found := false
	for i, item := range next.FactoryDenoms {
		if item.Denom != denom {
			continue
		}
		if item.Authority != authority {
			return FinancialZoneState{}, errors.New("financial contract asset authority mismatch")
		}
		supply, err := checkedAdd(item.Supply, amount, "financial contract asset supply")
		if err != nil {
			return FinancialZoneState{}, err
		}
		next.FactoryDenoms[i].Supply = supply
		found = true
		break
	}
	if !found {
		next.FactoryDenoms = append(next.FactoryDenoms, FinancialFactoryDenom{Denom: denom, Authority: authority, Supply: amount})
	}
	return adjustFinancialBalance(next, receiver, denom, amount, true)
}

func burnFinancialFactoryDenom(state FinancialZoneState, denom, owner string, amount uint64) (FinancialZoneState, error) {
	next, err := adjustFinancialBalance(state, owner, denom, amount, false)
	if err != nil {
		return FinancialZoneState{}, err
	}
	for i, item := range next.FactoryDenoms {
		if item.Denom != denom {
			continue
		}
		if item.Supply < amount {
			return FinancialZoneState{}, errors.New("financial contract asset burn exceeds supply")
		}
		next.FactoryDenoms[i].Supply -= amount
		return next.Normalize(), nil
	}
	return FinancialZoneState{}, errors.New("financial contract asset denom not found")
}

func applyFinancialDEXSwap(state FinancialZoneState, msg FinancialZoneMessage) (FinancialZoneState, error) {
	next := state.Normalize()
	for i, pool := range next.DEXPools {
		if pool.PoolID != msg.PoolID {
			continue
		}
		if pool.BaseDenom != msg.Denom || pool.QuoteDenom != msg.OutputDenom {
			return FinancialZoneState{}, errors.New("financial dex swap denom mismatch")
		}
		baseReserve, err := checkedAdd(pool.BaseReserve, msg.Amount, "financial dex base reserve")
		if err != nil {
			return FinancialZoneState{}, err
		}
		next.DEXPools[i].BaseReserve = baseReserve
		next.DEXOrders = append(next.DEXOrders, FinancialDEXOrder{
			OrderID:	msg.OrderID,
			PoolID:		msg.PoolID,
			Owner:		msg.AccountKey,
			InputDenom:	msg.Denom,
			OutputDenom:	msg.OutputDenom,
			InputAmount:	msg.Amount,
			Status:		"open",
		})
		return next.Normalize(), nil
	}
	return FinancialZoneState{}, errors.New("financial dex pool not found")
}

func settleFinancialDEXOrder(state FinancialZoneState, orderID string) (FinancialZoneState, error) {
	next := state.Normalize()
	for i, order := range next.DEXOrders {
		if order.OrderID == orderID {
			next.DEXOrders[i].Status = "settled"
			return next.Normalize(), nil
		}
	}
	return FinancialZoneState{}, errors.New("financial dex order not found")
}

func finalizeFinancialPaymentChannel(state FinancialZoneState, channelID string, height uint64) (FinancialZoneState, error) {
	next := state.Normalize()
	for i, channel := range next.PaymentChannels {
		if channel.ChannelID != channelID {
			continue
		}
		if channel.Disputed {
			return FinancialZoneState{}, errors.New("financial payment channel is disputed")
		}
		next.PaymentChannels[i].Finalized = true
		next.PaymentChannels[i].FinalizedHeight = height
		return next.Normalize(), nil
	}
	return FinancialZoneState{}, errors.New("financial payment channel not found")
}

func disputeFinancialPaymentCondition(state FinancialZoneState, conditionID string) (FinancialZoneState, error) {
	next := state.Normalize()
	for i, condition := range next.PaymentConditions {
		if condition.ConditionID != conditionID {
			continue
		}
		next.PaymentConditions[i].Disputed = true
		for j, channel := range next.PaymentChannels {
			if channel.ChannelID == condition.ChannelID {
				next.PaymentChannels[j].Disputed = true
				break
			}
		}
		return next.Normalize(), nil
	}
	return FinancialZoneState{}, errors.New("financial payment condition not found")
}

func adjustFinancialBalance(state FinancialZoneState, address, denom string, amount uint64, credit bool) (FinancialZoneState, error) {
	if amount == 0 {
		return FinancialZoneState{}, errors.New("financial balance amount must be positive")
	}
	if _, err := FinancialBalanceKey(address, denom); err != nil {
		return FinancialZoneState{}, err
	}
	next := state.Normalize()
	next.Accounts = append(next.Accounts, address)
	for i, balance := range next.Balances {
		if balance.Address != address || balance.Denom != denom {
			continue
		}
		if credit {
			sum, err := checkedAdd(balance.Amount, amount, "financial balance")
			if err != nil {
				return FinancialZoneState{}, err
			}
			next.Balances[i].Amount = sum
		} else {
			if balance.Amount < amount {
				return FinancialZoneState{}, errors.New("financial balance underflow")
			}
			next.Balances[i].Amount -= amount
		}
		return next.Normalize(), nil
	}
	if !credit {
		return FinancialZoneState{}, errors.New("financial balance not found")
	}
	next.Balances = append(next.Balances, FinancialBalance{Address: address, Denom: denom, Amount: amount})
	return next.Normalize(), nil
}

func checkedAdd(left, right uint64, fieldName string) (uint64, error) {
	if right > ^uint64(0)-left {
		return 0, fmt.Errorf("%s overflow", fieldName)
	}
	return left + right, nil
}

func routeFinancialStateKey(mode FinancialShardRoutingMode, stateKey string, routeKey string, shardCount uint32, layoutEpoch uint64) (FinancialShardRoute, error) {
	if shardCount == 0 {
		return FinancialShardRoute{}, errors.New("financial shard count must be positive")
	}
	if layoutEpoch == 0 {
		return FinancialShardRoute{}, errors.New("financial shard layout epoch must be positive")
	}
	if err := validateRuntimeToken("financial shard state key", stateKey, MaxZoneNamespaceLength); err != nil {
		return FinancialShardRoute{}, err
	}
	if err := validateRuntimeToken("financial shard route key", routeKey, MaxZoneEndpointLength); err != nil {
		return FinancialShardRoute{}, err
	}
	hash := hashRuntimeParts("aetra-financial-route-key-v1", string(mode), routeKey, fmt.Sprint(layoutEpoch))
	shardID := uint32(binary.BigEndian.Uint64([]byte(hash[:8])) % uint64(shardCount))
	route := FinancialShardRoute{
		ZoneID:		ZoneIDFinancial,
		LayoutEpoch:	layoutEpoch,
		ShardCount:	shardCount,
		ShardID:	shardID,
		RoutingMode:	mode,
		RouteKey:	routeKey,
		StateKey:	stateKey,
	}
	route.RouteHash = ComputeFinancialShardRouteHash(route)
	return route, route.ValidateHash()
}

func normalizeFinancialAccounts(accounts []string) []string {
	out := append([]string(nil), accounts...)
	sort.Strings(out)
	deduped := out[:0]
	for _, account := range out {
		if len(deduped) == 0 || deduped[len(deduped)-1] != account {
			deduped = append(deduped, account)
		}
	}
	return append([]string(nil), deduped...)
}

func normalizeFinancialBalances(balances []FinancialBalance) []FinancialBalance {
	out := append([]FinancialBalance(nil), balances...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Address != out[j].Address {
			return out[i].Address < out[j].Address
		}
		return out[i].Denom < out[j].Denom
	})
	return out
}

func normalizeFinancialFeeBuckets(buckets []FinancialFeeBucket) []FinancialFeeBucket {
	out := append([]FinancialFeeBucket(nil), buckets...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].BucketID != out[j].BucketID {
			return out[i].BucketID < out[j].BucketID
		}
		return out[i].Denom < out[j].Denom
	})
	return out
}

func normalizeFinancialShardFeeBuckets(buckets []FinancialShardFeeBucket) []FinancialShardFeeBucket {
	out := append([]FinancialShardFeeBucket(nil), buckets...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ShardID != out[j].ShardID {
			return out[i].ShardID < out[j].ShardID
		}
		return out[i].Denom < out[j].Denom
	})
	return out
}

func normalizeFinancialFactoryDenoms(denoms []FinancialFactoryDenom) []FinancialFactoryDenom {
	out := append([]FinancialFactoryDenom(nil), denoms...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Denom < out[j].Denom })
	return out
}

func normalizeFinancialDEXPools(pools []FinancialDEXPool) []FinancialDEXPool {
	out := append([]FinancialDEXPool(nil), pools...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func normalizeFinancialDEXOrders(orders []FinancialDEXOrder) []FinancialDEXOrder {
	out := append([]FinancialDEXOrder(nil), orders...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].OrderID < out[j].OrderID })
	return out
}

func normalizeFinancialPaymentChannels(channels []FinancialPaymentChannel) []FinancialPaymentChannel {
	out := append([]FinancialPaymentChannel(nil), channels...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeFinancialPaymentConditions(conditions []FinancialPaymentCondition) []FinancialPaymentCondition {
	out := append([]FinancialPaymentCondition(nil), conditions...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ConditionID < out[j].ConditionID })
	return out
}

func normalizeFinancialTransferEscrows(escrows []FinancialTransferEscrow) []FinancialTransferEscrow {
	out := append([]FinancialTransferEscrow(nil), escrows...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].TransferID < out[j].TransferID })
	return out
}
