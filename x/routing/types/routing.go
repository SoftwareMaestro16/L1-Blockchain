package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	NativeFeeDenom	= "naet"

	MaxFeeClass		FeeClass	= 10
	MaxReputationClass	ReputationClass	= 10
)

const (
	MsgTypeSoftwareUpgrade	= "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"
	MsgTypeCancelUpgrade	= "/cosmos.upgrade.v1beta1.MsgCancelUpgrade"

	MsgTypeCreateValidator		= "/cosmos.staking.v1beta1.MsgCreateValidator"
	MsgTypeEditValidator		= "/cosmos.staking.v1beta1.MsgEditValidator"
	MsgTypeDelegate			= "/cosmos.staking.v1beta1.MsgDelegate"
	MsgTypeUndelegate		= "/cosmos.staking.v1beta1.MsgUndelegate"
	MsgTypeBeginRedelegate		= "/cosmos.staking.v1beta1.MsgBeginRedelegate"
	MsgTypeCancelUnbonding		= "/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation"
	MsgTypeUnjail			= "/cosmos.slashing.v1beta1.MsgUnjail"
	MsgTypeSubmitEvidence		= "/cosmos.evidence.v1beta1.MsgSubmitEvidence"
	MsgTypeGovSubmitProposal	= "/cosmos.gov.v1.MsgSubmitProposal"
	MsgTypeGovVote			= "/cosmos.gov.v1.MsgVote"
	MsgTypeGovVoteWeighted		= "/cosmos.gov.v1.MsgVoteWeighted"
	MsgTypeGovDeposit		= "/cosmos.gov.v1.MsgDeposit"
	MsgTypeGovExecLegacyContent	= "/cosmos.gov.v1.MsgExecLegacyContent"
	MsgTypeGovUpdateParams		= "/cosmos.gov.v1.MsgUpdateParams"
	MsgTypeMintUpdateParams		= "/cosmos.mint.v1beta1.MsgUpdateParams"
	MsgTypeDistributionWithdraw	= "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"
	MsgTypeDistributionSetAddr	= "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress"
	MsgTypeDistributionWithdrawV	= "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission"

	MsgTypeBankSend		= "/cosmos.bank.v1beta1.MsgSend"
	MsgTypeBankMultiSend	= "/cosmos.bank.v1beta1.MsgMultiSend"

	MsgTypeIdentityRegister	= "/l1.identity.v1.MsgRegisterDomain"
	MsgTypeIdentityRenew	= "/l1.identity.v1.MsgRenewDomain"
	MsgTypeIdentityResolver	= "/l1.identity.v1.MsgSetResolver"
	MsgTypeIdentityReverse	= "/l1.identity.v1.MsgSetReverse"

	MsgTypeWasmStore	= "/cosmwasm.wasm.v1.MsgStoreCode"
	MsgTypeWasmInstantiate	= "/cosmwasm.wasm.v1.MsgInstantiateContract"
	MsgTypeWasmExecute	= "/cosmwasm.wasm.v1.MsgExecuteContract"
	MsgTypeWasmMigrate	= "/cosmwasm.wasm.v1.MsgMigrateContract"
	MsgTypeAVMDeploy	= "/aetra.avm.v1.MsgDeploy"
	MsgTypeAVMExecute	= "/aetra.avm.v1.MsgExecute"

	MsgTypeAsyncSend	= "/l1.messaging.v1.MsgSendAsync"
	MsgTypeAsyncEnqueue	= "/l1.queue.v1.MsgEnqueue"
	MsgTypeAsyncDeliver	= "/l1.aetravm.async.v1.MsgDeliver"

	MsgTypeMemoAttach	= "/l1.memo.v1.MsgAttachMemo"
	MsgTypePermissionsSet	= "/l1.permissions.v1.MsgSetPermission"
	MsgTypeWorkflowExecute	= "/l1.workflow.v1.MsgExecute"
	MsgTypeMarketOrder	= "/l1.market.v1.MsgSubmitOrder"
)

type TxClass string

const (
	TxClassCriticalSystem		TxClass	= "CRITICAL_SYSTEM"
	TxClassStakingGovSecurity	TxClass	= "STAKING_GOV_SECURITY"
	TxClassFinancial		TxClass	= "FINANCIAL"
	TxClassIdentity			TxClass	= "IDENTITY"
	TxClassContract			TxClass	= "CONTRACT"
	TxClassApplication		TxClass	= "APPLICATION"
	TxClassAsyncMessage		TxClass	= "ASYNC_MESSAGE"
)

type ZoneID string

const (
	ZoneAetraCore	ZoneID	= "AETHER_CORE"
	ZoneFinancial	ZoneID	= "FINANCIAL_ZONE"
	ZoneIdentity	ZoneID	= "IDENTITY_ZONE"
	ZoneContract	ZoneID	= "CONTRACT_ZONE"
	ZoneApplication	ZoneID	= "APPLICATION_ZONE"
)

type ShardID uint32

type ReputationClass uint32

type FeeClass uint32

type Locality struct {
	AccountKey		[]byte
	ContractAddress		[]byte
	Domain			string
	AssetDenom		string
	AsyncDestination	[]byte
}

type RouteInput struct {
	MsgType		string
	FeeDenom	string
	FeeClass	FeeClass
	ReputationClass	ReputationClass
	AdmissionHeight	uint64
	TxHash		[]byte
	RoutingEpoch	uint64
	ActiveShards	map[ZoneID]uint32
	Locality	Locality
}

type RouteDecision struct {
	TxClass		TxClass
	ZoneID		ZoneID
	ShardID		ShardID
	ActiveShards	uint32
	PrimaryActor	[]byte
	PriorityKey	PriorityKey
}

type PriorityKey struct {
	PriorityClass	uint32
	FeeClass	FeeClass
	ReputationClass	ReputationClass
	AdmissionHeight	uint64
	TxHash		[]byte
}

func ClassifyTx(msgType string) (TxClass, error) {
	switch normalizeMsgType(msgType) {
	case MsgTypeSoftwareUpgrade, MsgTypeCancelUpgrade:
		return TxClassCriticalSystem, nil
	case MsgTypeCreateValidator,
		MsgTypeEditValidator,
		MsgTypeDelegate,
		MsgTypeUndelegate,
		MsgTypeBeginRedelegate,
		MsgTypeCancelUnbonding,
		MsgTypeUnjail,
		MsgTypeSubmitEvidence,
		MsgTypeGovSubmitProposal,
		MsgTypeGovVote,
		MsgTypeGovVoteWeighted,
		MsgTypeGovDeposit,
		MsgTypeGovExecLegacyContent,
		MsgTypeGovUpdateParams,
		MsgTypeMintUpdateParams,
		MsgTypeDistributionWithdraw,
		MsgTypeDistributionSetAddr,
		MsgTypeDistributionWithdrawV:
		return TxClassStakingGovSecurity, nil
	case MsgTypeBankSend,
		MsgTypeBankMultiSend:
		return TxClassFinancial, nil
	case MsgTypeIdentityRegister, MsgTypeIdentityRenew, MsgTypeIdentityResolver, MsgTypeIdentityReverse:
		return TxClassIdentity, nil
	case MsgTypeWasmStore, MsgTypeWasmInstantiate, MsgTypeWasmExecute, MsgTypeWasmMigrate, MsgTypeAVMDeploy, MsgTypeAVMExecute:
		return TxClassContract, nil
	case MsgTypeAsyncSend, MsgTypeAsyncEnqueue, MsgTypeAsyncDeliver:
		return TxClassAsyncMessage, nil
	case MsgTypeMemoAttach, MsgTypePermissionsSet, MsgTypeWorkflowExecute, MsgTypeMarketOrder:
		return TxClassApplication, nil
	default:
		return "", fmt.Errorf("unknown routing message type %q", msgType)
	}
}

func Route(input RouteInput) (RouteDecision, error) {
	if strings.TrimSpace(input.FeeDenom) != NativeFeeDenom {
		return RouteDecision{}, fmt.Errorf("routing fee denom must be %s", NativeFeeDenom)
	}
	if len(input.TxHash) == 0 {
		return RouteDecision{}, errors.New("routing tx hash is required")
	}
	if err := input.Locality.Validate(); err != nil {
		return RouteDecision{}, err
	}
	txClass, err := ClassifyTx(input.MsgType)
	if err != nil {
		return RouteDecision{}, err
	}
	zone, err := ZoneForClass(txClass)
	if err != nil {
		return RouteDecision{}, err
	}
	primaryActor, err := input.Locality.PrimaryActor(txClass)
	if err != nil {
		return RouteDecision{}, err
	}

	decision := RouteDecision{
		TxClass:	txClass,
		ZoneID:		zone,
		PrimaryActor:	cloneBytes(primaryActor),
		PriorityKey:	BuildPriorityKey(txClass, input.FeeClass, input.ReputationClass, input.AdmissionHeight, input.TxHash),
	}
	if zone == ZoneAetraCore {
		decision.ActiveShards = 1
		return decision, nil
	}
	if len(primaryActor) == 0 {
		return RouteDecision{}, fmt.Errorf("routing primary actor is required for %s", txClass)
	}
	activeShards := input.ActiveShards[zone]
	if activeShards == 0 {
		return RouteDecision{}, fmt.Errorf("routing active shards for %s must be positive", zone)
	}
	decision.ActiveShards = activeShards
	decision.ShardID = AssignShard(zone, primaryActor, input.RoutingEpoch, activeShards)
	return decision, nil
}

func ZoneForClass(txClass TxClass) (ZoneID, error) {
	switch txClass {
	case TxClassCriticalSystem, TxClassStakingGovSecurity:
		return ZoneAetraCore, nil
	case TxClassFinancial:
		return ZoneFinancial, nil
	case TxClassIdentity:
		return ZoneIdentity, nil
	case TxClassContract:
		return ZoneContract, nil
	case TxClassApplication, TxClassAsyncMessage:
		return ZoneApplication, nil
	default:
		return "", fmt.Errorf("missing zone for tx class %q", txClass)
	}
}

func AssignShard(zone ZoneID, primaryActor []byte, routingEpoch uint64, activeShards uint32) ShardID {
	if activeShards == 0 {
		return 0
	}
	h := sha256.New()
	h.Write([]byte("aetra-routing-v1"))
	h.Write([]byte{0})
	h.Write([]byte(zone))
	h.Write([]byte{0})
	h.Write(primaryActor)
	h.Write([]byte{0})
	var epoch [8]byte
	binary.BigEndian.PutUint64(epoch[:], routingEpoch)
	h.Write(epoch[:])
	sum := h.Sum(nil)
	value := binary.BigEndian.Uint64(sum[:8])
	return ShardID(value % uint64(activeShards))
}

func BuildPriorityKey(txClass TxClass, feeClass FeeClass, reputation ReputationClass, admissionHeight uint64, txHash []byte) PriorityKey {
	return PriorityKey{
		PriorityClass:		PriorityClassForTx(txClass),
		FeeClass:		BoundFeeClass(feeClass),
		ReputationClass:	BoundReputationClass(reputation),
		AdmissionHeight:	admissionHeight,
		TxHash:			cloneBytes(txHash),
	}
}

func PriorityClassForTx(txClass TxClass) uint32 {
	switch txClass {
	case TxClassCriticalSystem:
		return 0
	case TxClassStakingGovSecurity:
		return 1
	case TxClassFinancial:
		return 2
	case TxClassIdentity:
		return 3
	case TxClassContract:
		return 4
	case TxClassAsyncMessage:
		return 5
	case TxClassApplication:
		return 6
	default:
		return 100
	}
}

func BoundFeeClass(feeClass FeeClass) FeeClass {
	if feeClass > MaxFeeClass {
		return MaxFeeClass
	}
	return feeClass
}

func BoundReputationClass(reputation ReputationClass) ReputationClass {
	if reputation > MaxReputationClass {
		return MaxReputationClass
	}
	return reputation
}

func ComparePriority(left, right PriorityKey) int {
	if left.PriorityClass != right.PriorityClass {
		return compareUint32Ascending(left.PriorityClass, right.PriorityClass)
	}
	if left.FeeClass != right.FeeClass {
		return compareUint32Descending(uint32(left.FeeClass), uint32(right.FeeClass))
	}
	if left.ReputationClass != right.ReputationClass {
		return compareUint32Descending(uint32(left.ReputationClass), uint32(right.ReputationClass))
	}
	if left.AdmissionHeight != right.AdmissionHeight {
		return compareUint64Ascending(left.AdmissionHeight, right.AdmissionHeight)
	}
	return bytes.Compare(left.TxHash, right.TxHash)
}

func SortPriorityKeys(keys []PriorityKey) []PriorityKey {
	out := make([]PriorityKey, len(keys))
	for i, key := range keys {
		key.TxHash = cloneBytes(key.TxHash)
		out[i] = key
	}
	sort.SliceStable(out, func(i, j int) bool {
		return ComparePriority(out[i], out[j]) < 0
	})
	return out
}

func SortDecisions(decisions []RouteDecision) []RouteDecision {
	out := make([]RouteDecision, len(decisions))
	for i, decision := range decisions {
		decision.PrimaryActor = cloneBytes(decision.PrimaryActor)
		decision.PriorityKey.TxHash = cloneBytes(decision.PriorityKey.TxHash)
		out[i] = decision
	}
	sort.SliceStable(out, func(i, j int) bool {
		return ComparePriority(out[i].PriorityKey, out[j].PriorityKey) < 0
	})
	return out
}

func (l Locality) Validate() error {
	if isZeroBytes(l.AccountKey) {
		return errors.New("routing account key must not be zero address")
	}
	if isZeroBytes(l.ContractAddress) {
		return errors.New("routing contract address must not be zero address")
	}
	if isZeroBytes(l.AsyncDestination) {
		return errors.New("routing async destination must not be zero address")
	}
	if strings.TrimSpace(l.Domain) != "" {
		if _, err := normalizeDomain(l.Domain); err != nil {
			return err
		}
	}
	if strings.TrimSpace(l.AssetDenom) != "" {
		if _, err := normalizeAssetDenom(l.AssetDenom); err != nil {
			return err
		}
	}
	return nil
}

func (l Locality) PrimaryActor(txClass TxClass) ([]byte, error) {
	switch txClass {
	case TxClassCriticalSystem, TxClassStakingGovSecurity:
		return cloneBytes(l.AccountKey), nil
	case TxClassFinancial:
		if len(l.AccountKey) > 0 {
			return cloneBytes(l.AccountKey), nil
		}
		if l.AssetDenom != "" {
			denom, err := normalizeAssetDenom(l.AssetDenom)
			if err != nil {
				return nil, err
			}
			return []byte("asset:" + denom), nil
		}
		return nil, nil
	case TxClassIdentity:
		domain, err := normalizeDomain(l.Domain)
		if err != nil {
			return nil, err
		}
		return []byte("domain:" + domain), nil
	case TxClassContract:
		if len(l.ContractAddress) > 0 {
			return cloneBytes(l.ContractAddress), nil
		}
		return nil, nil
	case TxClassApplication:
		if len(l.AccountKey) > 0 {
			return cloneBytes(l.AccountKey), nil
		}
		if l.AssetDenom != "" {
			denom, err := normalizeAssetDenom(l.AssetDenom)
			if err != nil {
				return nil, err
			}
			return []byte("asset:" + denom), nil
		}
		return nil, nil
	case TxClassAsyncMessage:
		return cloneBytes(l.AsyncDestination), nil
	default:
		return nil, fmt.Errorf("unknown tx class %q", txClass)
	}
}

func normalizeMsgType(msgType string) string {
	trimmed := strings.TrimSpace(msgType)
	if trimmed == "" || strings.HasPrefix(trimmed, "/") {
		return trimmed
	}
	return "/" + trimmed
}

func normalizeDomain(domain string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	if normalized == "" {
		return "", errors.New("routing domain key is required")
	}
	if !strings.HasSuffix(normalized, ".aet") {
		return "", errors.New("routing domain key must end with .aet")
	}
	labels := strings.Split(normalized, ".")
	for _, label := range labels {
		if label == "" {
			return "", errors.New("routing domain key contains empty label")
		}
	}
	return normalized, nil
}

func normalizeAssetDenom(denom string) (string, error) {
	normalized := strings.TrimSpace(denom)
	if normalized == "" {
		return "", errors.New("routing asset denom is required")
	}
	for _, r := range normalized {
		if r <= ' ' || r == 0x7f {
			return "", errors.New("routing asset denom must not contain whitespace or control characters")
		}
	}
	return normalized, nil
}

func isZeroBytes(bz []byte) bool {
	if len(bz) == 0 {
		return false
	}
	for _, b := range bz {
		if b != 0 {
			return false
		}
	}
	return true
}

func cloneBytes(bz []byte) []byte {
	if len(bz) == 0 {
		return nil
	}
	return append([]byte(nil), bz...)
}

func compareUint32Ascending(left, right uint32) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func compareUint32Descending(left, right uint32) int {
	return compareUint32Ascending(right, left)
}

func compareUint64Ascending(left, right uint64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
