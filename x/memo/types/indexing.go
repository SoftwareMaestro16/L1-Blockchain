package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	AssetTypeNative		= "native"
	AssetTypeSBT		= "sbt"
	AssetTypeContract	= "contract"
	AssetTypeDomain		= "domain"

	EventTypeBankTransfer		= "bank_transfer"
	EventTypeResolverPayment	= "resolver_payment"
	EventTypeSBTProofRevoke		= "sbt_proof_revoke"
	EventTypeContractCall		= "contract_call"
	EventTypeDomainAuctionBid	= "domain_auction_bid"
	EventTypeDomainRenewal		= "domain_renewal"
	EventTypeMemoAttached		= "memo_attached"

	StoragePolicyFullMemo		= "full_memo_onchain"
	StoragePolicyHashOnlyOnchain	= "hash_only_onchain"
)

type MemoStoragePolicy struct {
	Mode		string
	MaxOnChainBytes	uint32
}

type MemoStoreRecord struct {
	TxHash		[]byte
	Sender		sdk.AccAddress
	Receiver	sdk.AccAddress
	Contract	sdk.AccAddress
	AssetType	string
	RelatedDomain	string
	EventType	string
	Memo		string
	MemoHash	[]byte
	BlockHeight	int64
	TimestampUnix	int64
}

type MemoIndex struct {
	ByTxHash	map[string][]MemoStoreRecord
	BySender	map[string][]MemoStoreRecord
	ByReceiver	map[string][]MemoStoreRecord
	ByDomain	map[string][]MemoStoreRecord
	ByContract	map[string][]MemoStoreRecord
	ByAsset		map[string][]MemoStoreRecord
	ByEventType	map[string][]MemoStoreRecord
}

type EventMemoAttached struct {
	TxHash		[]byte
	From		sdk.AccAddress
	To		sdk.AccAddress
	Domain		string
	MemoHash	[]byte
	Memo		string
}

func DefaultMemoStoragePolicy(params MemoParams) MemoStoragePolicy {
	return MemoStoragePolicy{Mode: StoragePolicyFullMemo, MaxOnChainBytes: params.MaxMemoBytes}
}

func BuildMemoStoreRecord(metadata TxMetadata, params MemoParams, policy MemoStoragePolicy, base MemoStoreRecord) (MemoStoreRecord, EventMemoAttached, error) {
	if err := ValidateTxMetadata(metadata, params); err != nil {
		return MemoStoreRecord{}, EventMemoAttached{}, err
	}
	if err := ValidateMemoStoragePolicy(policy, params); err != nil {
		return MemoStoreRecord{}, EventMemoAttached{}, err
	}
	if err := validateMemoRecordBase(base); err != nil {
		return MemoStoreRecord{}, EventMemoAttached{}, err
	}
	record := base
	record.MemoHash = metadata.MemoHash
	if len(record.MemoHash) == 0 && metadata.Memo != "" {
		record.MemoHash = MemoHash(metadata.Memo)
	}
	if policy.Mode == StoragePolicyFullMemo && len([]byte(metadata.Memo)) <= int(policy.MaxOnChainBytes) {
		record.Memo = metadata.Memo
	}
	if policy.Mode == StoragePolicyHashOnlyOnchain {
		record.Memo = ""
	}
	event := DeterministicMemoEvent(record)
	return record, event, nil
}

func ValidateMemoStoragePolicy(policy MemoStoragePolicy, params MemoParams) error {
	if policy.MaxOnChainBytes > params.MaxMemoBytes {
		return errors.New("memo storage policy exceeds configured memo byte limit")
	}
	switch policy.Mode {
	case StoragePolicyFullMemo, StoragePolicyHashOnlyOnchain:
		return nil
	default:
		return fmt.Errorf("unknown memo storage policy %q", policy.Mode)
	}
}

func IndexMemoRecords(records []MemoStoreRecord) (MemoIndex, error) {
	index := MemoIndex{
		ByTxHash:	make(map[string][]MemoStoreRecord),
		BySender:	make(map[string][]MemoStoreRecord),
		ByReceiver:	make(map[string][]MemoStoreRecord),
		ByDomain:	make(map[string][]MemoStoreRecord),
		ByContract:	make(map[string][]MemoStoreRecord),
		ByAsset:	make(map[string][]MemoStoreRecord),
		ByEventType:	make(map[string][]MemoStoreRecord),
	}
	for _, record := range records {
		if err := ValidateMemoStoreRecord(record); err != nil {
			return MemoIndex{}, err
		}
		appendIndex(index.ByTxHash, string(record.TxHash), record)
		appendIndex(index.BySender, string(record.Sender), record)
		if len(record.Receiver) > 0 {
			appendIndex(index.ByReceiver, string(record.Receiver), record)
		}
		if record.RelatedDomain != "" {
			appendIndex(index.ByDomain, record.RelatedDomain, record)
		}
		if len(record.Contract) > 0 {
			appendIndex(index.ByContract, string(record.Contract), record)
		}
		appendIndex(index.ByAsset, record.AssetType, record)
		appendIndex(index.ByEventType, record.EventType, record)
	}
	sortMemoIndex(index)
	return index, nil
}

func ValidateMemoStoreRecord(record MemoStoreRecord) error {
	if err := validateMemoRecordBase(record); err != nil {
		return err
	}
	if record.Memo != "" && len(record.MemoHash) == 0 {
		return errors.New("memo hash is required when memo is stored")
	}
	if len(record.MemoHash) > 0 && len(record.MemoHash) != MemoHashBytes {
		return fmt.Errorf("memo hash must be %d bytes", MemoHashBytes)
	}
	return nil
}

func MemoSearchIndexAffectsConsensus() bool {
	return false
}

func DeterministicMemoEvent(record MemoStoreRecord) EventMemoAttached {
	return EventMemoAttached{
		TxHash:		cloneBytes(record.TxHash),
		From:		cloneAddress(record.Sender),
		To:		cloneAddress(record.Receiver),
		Domain:		record.RelatedDomain,
		MemoHash:	cloneBytes(record.MemoHash),
		Memo:		record.Memo,
	}
}

func validateMemoRecordBase(record MemoStoreRecord) error {
	if len(record.TxHash) == 0 {
		return errors.New("memo tx hash is required")
	}
	if len(record.Sender) == 0 {
		return errors.New("memo sender is required")
	}
	if err := addressing.RejectZeroAddress("memo sender", record.Sender); err != nil {
		return err
	}
	if len(record.Receiver) > 0 {
		if err := addressing.RejectZeroAddress("memo receiver", record.Receiver); err != nil {
			return err
		}
	}
	if len(record.Contract) > 0 {
		if err := addressing.RejectZeroAddress("memo contract", record.Contract); err != nil {
			return err
		}
	}
	if !IsAssetType(record.AssetType) {
		return fmt.Errorf("invalid memo asset type %q", record.AssetType)
	}
	if !IsEventType(record.EventType) {
		return fmt.Errorf("invalid memo event type %q", record.EventType)
	}
	if record.BlockHeight < 0 {
		return errors.New("memo block height must be non-negative")
	}
	if record.TimestampUnix < 0 {
		return errors.New("memo timestamp must be non-negative")
	}
	return nil
}

func IsAssetType(assetType string) bool {
	switch assetType {
	case AssetTypeNative, AssetTypeSBT, AssetTypeContract, AssetTypeDomain:
		return true
	default:
		return false
	}
}

func IsEventType(eventType string) bool {
	switch eventType {
	case EventTypeBankTransfer,
		EventTypeResolverPayment,
		EventTypeSBTProofRevoke,
		EventTypeContractCall,
		EventTypeDomainAuctionBid,
		EventTypeDomainRenewal,
		EventTypeMemoAttached:
		return true
	default:
		return false
	}
}

func appendIndex(index map[string][]MemoStoreRecord, key string, record MemoStoreRecord) {
	index[key] = append(index[key], record)
}

func sortMemoIndex(index MemoIndex) {
	for _, bucket := range memoIndexBuckets(index) {
		keys := make([]string, 0, len(bucket))
		for key := range bucket {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sort.SliceStable(bucket[key], func(i, j int) bool {
				if bucket[key][i].BlockHeight == bucket[key][j].BlockHeight {
					return string(bucket[key][i].TxHash) < string(bucket[key][j].TxHash)
				}
				return bucket[key][i].BlockHeight < bucket[key][j].BlockHeight
			})
		}
	}
}

func memoIndexBuckets(index MemoIndex) []map[string][]MemoStoreRecord {
	return []map[string][]MemoStoreRecord{
		index.ByTxHash,
		index.BySender,
		index.ByReceiver,
		index.ByDomain,
		index.ByContract,
		index.ByAsset,
		index.ByEventType,
	}
}

func cloneBytes(bz []byte) []byte {
	if len(bz) == 0 {
		return nil
	}
	return append([]byte(nil), bz...)
}

func cloneAddress(addr sdk.AccAddress) sdk.AccAddress {
	if len(addr) == 0 {
		return nil
	}
	return append(sdk.AccAddress(nil), addr...)
}
