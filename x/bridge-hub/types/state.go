package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	BridgeRiskNormal	= "normal"
	BridgeRiskElevated	= "elevated"
	BridgeRiskCritical	= "critical"

	BridgeEventPending	= "pending"
	BridgeEventFinalized	= "finalized"
	BridgeEventRejected	= "rejected"

	ProofPolicyLightClient	= "light_client"
	ProofPolicyMultisig	= "multisig"
	ProofPolicyZK		= "zk"

	DefaultFeeScale	= uint32(10_000)
)

type BridgeHubParams struct {
	MaxBridges		uint32
	MaxAssetMappings	uint32
	MaxPendingEvents	uint32
	MaxOperators		uint32
	FeeScale		uint32
	DefaultDailyLimit	uint64
}

type BridgeHubState struct {
	Bridges		[]BridgeRecord
	AssetMappings	[]AssetMapping
	Events		[]BridgeEvent
}

type BridgeRecord struct {
	BridgeID		string
	SourceChain		string
	TargetChain		string
	Operators		[]string
	RiskStatus		string
	Paused			bool
	ProofPolicy		string
	DailyLimit		uint64
	DailyUsed		uint64
	LimitWindowStart	uint64
	FeePolicy		BridgeFeePolicy
	RegisteredHeight	uint64
	UpdatedHeight		uint64
}

type BridgeFeePolicy struct {
	FeeBps		uint32
	Collector	string
	MinimumFee	uint64
}

type AssetMapping struct {
	BridgeID	string
	SourceAsset	string
	TargetAsset	string
	Decimals	uint32
	Enabled		bool
}

type BridgeEvent struct {
	EventID		string
	BridgeID	string
	SourceChain	string
	Asset		string
	Amount		uint64
	ProofPolicy	string
	ProofRoot	string
	SubmittedBy	string
	SubmittedHeight	uint64
	FinalizedHeight	uint64
	Status		string
}

type MsgRegisterBridge struct {
	Authority	string
	Bridge		BridgeRecord
}

type MsgPauseBridge struct {
	Authority	string
	BridgeID	string
	Height		uint64
}

type MsgResumeBridge struct {
	Authority	string
	BridgeID	string
	Height		uint64
}

type MsgRegisterAssetMapping struct {
	Authority	string
	Mapping		AssetMapping
}

type MsgUpdateBridgeLimits struct {
	Authority	string
	BridgeID	string
	DailyLimit	uint64
	Height		uint64
}

type MsgSubmitBridgeEvent struct {
	Submitter	string
	Event		BridgeEvent
}

type MsgFinalizeBridgeEvent struct {
	Authority	string
	EventID		string
	Height		uint64
}

func DefaultBridgeHubParams() BridgeHubParams {
	return BridgeHubParams{
		MaxBridges:		1_024,
		MaxAssetMappings:	100_000,
		MaxPendingEvents:	100_000,
		MaxOperators:		128,
		FeeScale:		DefaultFeeScale,
		DefaultDailyLimit:	1_000_000,
	}
}

func EmptyBridgeHubState() BridgeHubState {
	return BridgeHubState{
		Bridges:	[]BridgeRecord{},
		AssetMappings:	[]AssetMapping{},
		Events:		[]BridgeEvent{},
	}
}

func (p BridgeHubParams) Validate() error {
	if p.MaxBridges == 0 || p.MaxAssetMappings == 0 || p.MaxPendingEvents == 0 || p.MaxOperators == 0 {
		return errors.New("bridge hub limits must be positive")
	}
	if p.FeeScale == 0 {
		return errors.New("bridge hub fee scale must be positive")
	}
	return nil
}

func (s BridgeHubState) Export() BridgeHubState {
	out := BridgeHubState{
		Bridges:	cloneBridges(s.Bridges),
		AssetMappings:	cloneMappings(s.AssetMappings),
		Events:		cloneEvents(s.Events),
	}
	SortBridges(out.Bridges)
	SortMappings(out.AssetMappings)
	SortEvents(out.Events)
	return out
}

func (s BridgeHubState) Validate(params BridgeHubParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Bridges)) > params.MaxBridges {
		return errors.New("bridge hub bridge count exceeds limit")
	}
	if uint32(len(s.AssetMappings)) > params.MaxAssetMappings {
		return errors.New("bridge hub asset mapping count exceeds limit")
	}
	bridgeIDs := map[string]BridgeRecord{}
	for _, bridge := range s.Bridges {
		if err := bridge.Validate(params); err != nil {
			return err
		}
		if _, found := bridgeIDs[bridge.BridgeID]; found {
			return fmt.Errorf("duplicate bridge id %q", bridge.BridgeID)
		}
		bridgeIDs[bridge.BridgeID] = bridge
	}
	mappingBySource := map[string]AssetMapping{}
	mappingByTarget := map[string]AssetMapping{}
	seenMappings := map[string]struct{}{}
	for _, mapping := range s.AssetMappings {
		if err := mapping.Validate(); err != nil {
			return err
		}
		if _, found := bridgeIDs[mapping.BridgeID]; !found {
			return fmt.Errorf("bridge asset mapping references unknown bridge %q", mapping.BridgeID)
		}
		sourceKey := mapping.BridgeID + "\x00" + mapping.SourceAsset
		targetKey := mapping.BridgeID + "\x00" + mapping.TargetAsset
		exactKey := sourceKey + "\x00" + mapping.TargetAsset
		if _, found := seenMappings[exactKey]; found {
			return fmt.Errorf("duplicate bridge asset mapping %q", mapping.SourceAsset)
		}
		if previous, found := mappingBySource[sourceKey]; found && previous.TargetAsset != mapping.TargetAsset {
			return fmt.Errorf("conflicting source asset mapping %q", mapping.SourceAsset)
		}
		if previous, found := mappingByTarget[targetKey]; found && previous.SourceAsset != mapping.SourceAsset {
			return fmt.Errorf("conflicting target asset mapping %q", mapping.TargetAsset)
		}
		mappingBySource[sourceKey] = mapping
		mappingByTarget[targetKey] = mapping
		seenMappings[exactKey] = struct{}{}
	}
	pending := uint32(0)
	seenEvents := map[string]struct{}{}
	for _, event := range s.Events {
		if err := event.Validate(); err != nil {
			return err
		}
		bridge, found := bridgeIDs[event.BridgeID]
		if !found {
			return fmt.Errorf("bridge event references unknown bridge %q", event.BridgeID)
		}
		if event.ProofPolicy != bridge.ProofPolicy {
			return fmt.Errorf("bridge event proof policy must match registered chain policy for %q", event.EventID)
		}
		if _, found := seenEvents[event.EventID]; found {
			return fmt.Errorf("duplicate bridge event %q", event.EventID)
		}
		seenEvents[event.EventID] = struct{}{}
		if event.Status == BridgeEventPending {
			pending++
		}
	}
	if pending > params.MaxPendingEvents {
		return errors.New("bridge hub pending event count exceeds limit")
	}
	return nil
}

func (b BridgeRecord) Normalize(params BridgeHubParams) BridgeRecord {
	b.BridgeID = strings.TrimSpace(b.BridgeID)
	b.SourceChain = strings.TrimSpace(b.SourceChain)
	b.TargetChain = strings.TrimSpace(b.TargetChain)
	b.Operators = normalizeStrings(b.Operators)
	b.RiskStatus = strings.TrimSpace(b.RiskStatus)
	if b.RiskStatus == "" {
		b.RiskStatus = BridgeRiskNormal
	}
	b.ProofPolicy = strings.TrimSpace(b.ProofPolicy)
	if b.DailyLimit == 0 {
		b.DailyLimit = params.DefaultDailyLimit
	}
	b.FeePolicy = b.FeePolicy.Normalize()
	return b
}

func (b BridgeRecord) Validate(params BridgeHubParams) error {
	b = b.Normalize(params)
	if b.BridgeID == "" || b.SourceChain == "" || b.TargetChain == "" {
		return errors.New("bridge id and chains are required")
	}
	if len(b.Operators) == 0 || uint32(len(b.Operators)) > params.MaxOperators {
		return errors.New("bridge operators must be present and within limit")
	}
	if !IsBridgeRiskStatus(b.RiskStatus) {
		return errors.New("bridge risk status is invalid")
	}
	if !IsProofPolicy(b.ProofPolicy) {
		return errors.New("bridge proof policy is invalid")
	}
	if b.DailyLimit == 0 {
		return errors.New("bridge daily limit must be positive")
	}
	if b.DailyUsed > b.DailyLimit {
		return errors.New("bridge daily used exceeds limit")
	}
	if b.RegisteredHeight == 0 || b.UpdatedHeight == 0 {
		return errors.New("bridge heights must be positive")
	}
	return b.FeePolicy.Validate(params)
}

func (f BridgeFeePolicy) Normalize() BridgeFeePolicy {
	f.Collector = strings.TrimSpace(f.Collector)
	return f
}

func (f BridgeFeePolicy) Validate(params BridgeHubParams) error {
	f = f.Normalize()
	if f.FeeBps > params.FeeScale {
		return errors.New("bridge fee bps exceeds fee scale")
	}
	if f.FeeBps > 0 && f.Collector == "" {
		return errors.New("bridge fee collector is required")
	}
	return nil
}

func (m AssetMapping) Normalize() AssetMapping {
	m.BridgeID = strings.TrimSpace(m.BridgeID)
	m.SourceAsset = strings.TrimSpace(m.SourceAsset)
	m.TargetAsset = strings.TrimSpace(m.TargetAsset)
	return m
}

func (m AssetMapping) Validate() error {
	m = m.Normalize()
	if m.BridgeID == "" || m.SourceAsset == "" || m.TargetAsset == "" {
		return errors.New("bridge asset mapping fields are required")
	}
	if m.Decimals > 36 {
		return errors.New("bridge asset mapping decimals exceed limit")
	}
	return nil
}

func (e BridgeEvent) Normalize() BridgeEvent {
	e.EventID = strings.TrimSpace(e.EventID)
	e.BridgeID = strings.TrimSpace(e.BridgeID)
	e.SourceChain = strings.TrimSpace(e.SourceChain)
	e.Asset = strings.TrimSpace(e.Asset)
	e.ProofPolicy = strings.TrimSpace(e.ProofPolicy)
	e.ProofRoot = strings.TrimSpace(e.ProofRoot)
	e.SubmittedBy = strings.TrimSpace(e.SubmittedBy)
	e.Status = strings.TrimSpace(e.Status)
	if e.Status == "" {
		e.Status = BridgeEventPending
	}
	return e
}

func (e BridgeEvent) Validate() error {
	e = e.Normalize()
	if e.EventID == "" || e.BridgeID == "" || e.SourceChain == "" || e.Asset == "" {
		return errors.New("bridge event identifiers are required")
	}
	if e.Amount == 0 {
		return errors.New("bridge event amount must be positive")
	}
	if !IsProofPolicy(e.ProofPolicy) {
		return errors.New("bridge event proof policy is invalid")
	}
	if err := ValidateProofRoot(e.ProofRoot); err != nil {
		return err
	}
	if e.SubmittedBy == "" || e.SubmittedHeight == 0 {
		return errors.New("bridge event submitter and submitted height are required")
	}
	if !IsBridgeEventStatus(e.Status) {
		return errors.New("bridge event status is invalid")
	}
	if e.Status == BridgeEventFinalized && e.FinalizedHeight == 0 {
		return errors.New("finalized bridge event requires finalized height")
	}
	if e.Status != BridgeEventFinalized && e.FinalizedHeight != 0 {
		return errors.New("non-finalized bridge event cannot have finalized height")
	}
	return nil
}

func IsBridgeRiskStatus(status string) bool {
	switch status {
	case BridgeRiskNormal, BridgeRiskElevated, BridgeRiskCritical:
		return true
	default:
		return false
	}
}

func IsBridgeEventStatus(status string) bool {
	switch status {
	case BridgeEventPending, BridgeEventFinalized, BridgeEventRejected:
		return true
	default:
		return false
	}
}

func IsProofPolicy(policy string) bool {
	switch policy {
	case ProofPolicyLightClient, ProofPolicyMultisig, ProofPolicyZK:
		return true
	default:
		return false
	}
}

func ValidateProofRoot(root string) error {
	root = strings.TrimSpace(root)
	if len(root) != 64 {
		return errors.New("bridge proof root must be 32-byte hex")
	}
	if _, err := hex.DecodeString(root); err != nil {
		return fmt.Errorf("bridge proof root must be hex: %w", err)
	}
	return nil
}

func CurrentWindowStart(height uint64) uint64 {
	if height == 0 {
		return 0
	}
	return ((height-1)/86_400)*86_400 + 1
}

func SortBridges(bridges []BridgeRecord) {
	sort.SliceStable(bridges, func(i, j int) bool { return bridges[i].BridgeID < bridges[j].BridgeID })
}

func SortMappings(mappings []AssetMapping) {
	sort.SliceStable(mappings, func(i, j int) bool {
		if mappings[i].BridgeID != mappings[j].BridgeID {
			return mappings[i].BridgeID < mappings[j].BridgeID
		}
		if mappings[i].SourceAsset != mappings[j].SourceAsset {
			return mappings[i].SourceAsset < mappings[j].SourceAsset
		}
		return mappings[i].TargetAsset < mappings[j].TargetAsset
	})
}

func SortEvents(events []BridgeEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].SubmittedHeight != events[j].SubmittedHeight {
			return events[i].SubmittedHeight < events[j].SubmittedHeight
		}
		return events[i].EventID < events[j].EventID
	})
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func cloneBridges(bridges []BridgeRecord) []BridgeRecord {
	out := append([]BridgeRecord(nil), bridges...)
	params := DefaultBridgeHubParams()
	for i := range out {
		out[i] = out[i].Normalize(params)
	}
	return out
}

func cloneMappings(mappings []AssetMapping) []AssetMapping {
	out := append([]AssetMapping(nil), mappings...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneEvents(events []BridgeEvent) []BridgeEvent {
	out := append([]BridgeEvent(nil), events...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}
