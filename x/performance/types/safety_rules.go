package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultShardProofHorizon = uint64(128)
)

type ConsensusDeterminismPolicy struct {
	ConsensusPathName		string
	UsesExternalAPIs		bool
	UsesLocalClock			bool
	UsesRandomShardPlacement	bool
	UsesNondeterministicMapOrder	bool
	UsesMempoolOnlyExecutionData	bool
	UsesFloatingPointMath		bool
	UsesConsensusTime		bool
	SortsStateTransitionMaps	bool
	EncodesProposalMempoolInputs	bool
	DeterminismPolicyHash		string
}

type RoutingSafetyPath struct {
	PathID		string
	RouteHash	string
	Cost		string
	HopCount	uint32
	LiquidityBps	uint32
	CongestionBps	uint32
	TieBreakKey	string
}

type RoutingFailureAccounting struct {
	RouteID		string
	OriginalValue	string
	BouncedValue	string
	BurnedValue	string
	ReceiptHash	string
	ReceiptStatus	string
	AccountingHash	string
}

type RoutingSafetyInput struct {
	RoutingEpoch		uint64
	RoutingTableHash	string
	RoutingMetricsRoot	string
	CandidatePaths		[]RoutingSafetyPath
	SelectedPathID		string
	FailureAccounting	[]RoutingFailureAccounting
	SafetyHash		string
}

type InFlightShardMessage struct {
	MessageID		string
	SourceShardID		string
	DestinationShardID	string
	DeliveryEpoch		uint64
	ExpiryHeight		uint64
	MessageHash		string
}

type ShardLayoutTransition struct {
	ZoneID			string
	PreviousLayoutEpoch	uint64
	NextLayoutEpoch		uint64
	CurrentHeight		uint64
	ActivationHeight	uint64
	EpochBoundaryHeight	uint64
	SplitMergeDecisionHash	string
	MigrationRoot		string
	OldLayoutHash		string
	NewLayoutHash		string
	ProofHorizon		uint64
	InFlightMessages	[]InFlightShardMessage
	TransitionHash		string
}

func BuildConsensusDeterminismPolicy(policy ConsensusDeterminismPolicy) (ConsensusDeterminismPolicy, error) {
	policy = policy.Normalize()
	policy.DeterminismPolicyHash = ComputeConsensusDeterminismPolicyHash(policy)
	return policy, policy.Validate()
}

func (p ConsensusDeterminismPolicy) Normalize() ConsensusDeterminismPolicy {
	p.ConsensusPathName = strings.TrimSpace(p.ConsensusPathName)
	p.DeterminismPolicyHash = normalizeLowerHex(p.DeterminismPolicyHash)
	return p
}

func (p ConsensusDeterminismPolicy) Validate() error {
	policy := p.Normalize()
	if err := validateExecutionToken("safety consensus path name", policy.ConsensusPathName); err != nil {
		return err
	}
	if policy.UsesExternalAPIs {
		return errors.New("safety consensus execution must not call external APIs")
	}
	if policy.UsesLocalClock && !policy.UsesConsensusTime {
		return errors.New("safety consensus execution must not use local clock outside consensus time")
	}
	if policy.UsesRandomShardPlacement {
		return errors.New("safety shard placement must not be random")
	}
	if policy.UsesNondeterministicMapOrder && !policy.SortsStateTransitionMaps {
		return errors.New("safety state transitions must not use nondeterministic map iteration")
	}
	if policy.UsesMempoolOnlyExecutionData && !policy.EncodesProposalMempoolInputs {
		return errors.New("safety mempool-only data must not affect execution results")
	}
	if policy.UsesFloatingPointMath {
		return errors.New("safety consensus execution must not use floating point math")
	}
	if err := validateHexHash("safety determinism policy hash", policy.DeterminismPolicyHash); err != nil {
		return err
	}
	if policy.DeterminismPolicyHash != ComputeConsensusDeterminismPolicyHash(policy) {
		return errors.New("safety determinism policy hash mismatch")
	}
	return nil
}

func BuildRoutingSafetyInput(input RoutingSafetyInput) (RoutingSafetyInput, error) {
	input = input.Normalize()
	input.SafetyHash = ComputeRoutingSafetyHash(input)
	return input, input.Validate()
}

func (i RoutingSafetyInput) Normalize() RoutingSafetyInput {
	i.RoutingTableHash = normalizeLowerHex(i.RoutingTableHash)
	i.RoutingMetricsRoot = normalizeLowerHex(i.RoutingMetricsRoot)
	i.SelectedPathID = strings.TrimSpace(i.SelectedPathID)
	for idx := range i.CandidatePaths {
		i.CandidatePaths[idx] = i.CandidatePaths[idx].Normalize()
	}
	sort.SliceStable(i.CandidatePaths, func(left, right int) bool {
		return compareRoutingSafetyPath(i.CandidatePaths[left], i.CandidatePaths[right]) < 0
	})
	for idx := range i.FailureAccounting {
		i.FailureAccounting[idx] = i.FailureAccounting[idx].Normalize()
	}
	sort.SliceStable(i.FailureAccounting, func(left, right int) bool {
		return i.FailureAccounting[left].RouteID < i.FailureAccounting[right].RouteID
	})
	i.SafetyHash = normalizeLowerHex(i.SafetyHash)
	return i
}

func (i RoutingSafetyInput) Validate() error {
	input := i.Normalize()
	if input.RoutingEpoch == 0 {
		return errors.New("safety routing epoch must be positive")
	}
	if err := validateHexHash("safety routing table hash", input.RoutingTableHash); err != nil {
		return err
	}
	if err := validateHexHash("safety routing metrics root", input.RoutingMetricsRoot); err != nil {
		return err
	}
	if len(input.CandidatePaths) == 0 {
		return errors.New("safety routing requires candidate paths")
	}
	seen := make(map[string]struct{}, len(input.CandidatePaths))
	for idx, path := range input.CandidatePaths {
		if err := path.Validate(); err != nil {
			return err
		}
		if _, found := seen[path.PathID]; found {
			return errors.New("safety routing duplicate path id")
		}
		seen[path.PathID] = struct{}{}
		if idx > 0 && compareRoutingSafetyPath(input.CandidatePaths[idx-1], path) >= 0 {
			return errors.New("safety routing paths must be sorted by deterministic tie-break")
		}
	}
	if input.SelectedPathID != input.CandidatePaths[0].PathID {
		return errors.New("safety routing selected path must match deterministic best path")
	}
	for _, accounting := range input.FailureAccounting {
		if err := accounting.Validate(); err != nil {
			return err
		}
	}
	if input.SafetyHash != ComputeRoutingSafetyHash(input) {
		return errors.New("safety routing hash mismatch")
	}
	return nil
}

func (p RoutingSafetyPath) Normalize() RoutingSafetyPath {
	p.PathID = strings.TrimSpace(p.PathID)
	p.RouteHash = normalizeLowerHex(p.RouteHash)
	p.Cost = strings.TrimSpace(p.Cost)
	p.TieBreakKey = strings.TrimSpace(p.TieBreakKey)
	return p
}

func (p RoutingSafetyPath) Validate() error {
	path := p.Normalize()
	if err := validateExecutionToken("safety routing path id", path.PathID); err != nil {
		return err
	}
	if err := validateHexHash("safety routing path hash", path.RouteHash); err != nil {
		return err
	}
	if _, err := parsePerformanceNonNegativeInt("safety routing path cost", path.Cost); err != nil {
		return err
	}
	if path.HopCount == 0 {
		return errors.New("safety routing path hop count must be positive")
	}
	if path.LiquidityBps > 10_000 || path.CongestionBps > 10_000 {
		return errors.New("safety routing path bps values must be <= 10000")
	}
	if path.TieBreakKey == "" {
		return errors.New("safety routing path tie-break key is required")
	}
	return nil
}

func (a RoutingFailureAccounting) Normalize() RoutingFailureAccounting {
	a.RouteID = strings.TrimSpace(a.RouteID)
	a.OriginalValue = strings.TrimSpace(a.OriginalValue)
	a.BouncedValue = strings.TrimSpace(a.BouncedValue)
	a.BurnedValue = strings.TrimSpace(a.BurnedValue)
	a.ReceiptHash = normalizeLowerHex(a.ReceiptHash)
	a.ReceiptStatus = strings.TrimSpace(a.ReceiptStatus)
	a.AccountingHash = normalizeLowerHex(a.AccountingHash)
	return a
}

func (a RoutingFailureAccounting) Validate() error {
	accounting := a.Normalize()
	if err := validateExecutionToken("safety failed route id", accounting.RouteID); err != nil {
		return err
	}
	original, err := parsePerformanceNonNegativeInt("safety failed route original value", accounting.OriginalValue)
	if err != nil {
		return err
	}
	bounced, err := parsePerformanceNonNegativeInt("safety failed route bounced value", accounting.BouncedValue)
	if err != nil {
		return err
	}
	burned, err := parsePerformanceNonNegativeInt("safety failed route burned value", accounting.BurnedValue)
	if err != nil {
		return err
	}
	if burned.IsPositive() && accounting.ReceiptHash == "" {
		return errors.New("safety failed route cannot burn value without receipt")
	}
	if bounced.Add(burned).GT(original) {
		return errors.New("safety failed route bounce cannot create extra value")
	}
	if accounting.ReceiptHash != "" {
		if err := validateHexHash("safety failed route receipt hash", accounting.ReceiptHash); err != nil {
			return err
		}
	}
	if accounting.ReceiptStatus == "" {
		return errors.New("safety failed route receipt status is required")
	}
	if accounting.AccountingHash != ComputeRoutingFailureAccountingHash(accounting) {
		return errors.New("safety failed route accounting hash mismatch")
	}
	return nil
}

func BuildShardLayoutTransition(transition ShardLayoutTransition) (ShardLayoutTransition, error) {
	transition = transition.Normalize()
	if transition.ProofHorizon == 0 {
		transition.ProofHorizon = DefaultShardProofHorizon
	}
	transition.TransitionHash = ComputeShardLayoutTransitionHash(transition)
	return transition, transition.Validate()
}

func (t ShardLayoutTransition) Normalize() ShardLayoutTransition {
	t.ZoneID = strings.TrimSpace(t.ZoneID)
	t.SplitMergeDecisionHash = normalizeLowerHex(t.SplitMergeDecisionHash)
	t.MigrationRoot = normalizeLowerHex(t.MigrationRoot)
	t.OldLayoutHash = normalizeLowerHex(t.OldLayoutHash)
	t.NewLayoutHash = normalizeLowerHex(t.NewLayoutHash)
	for idx := range t.InFlightMessages {
		t.InFlightMessages[idx] = t.InFlightMessages[idx].Normalize()
	}
	sort.SliceStable(t.InFlightMessages, func(left, right int) bool {
		return t.InFlightMessages[left].MessageID < t.InFlightMessages[right].MessageID
	})
	t.TransitionHash = normalizeLowerHex(t.TransitionHash)
	return t
}

func (t ShardLayoutTransition) Validate() error {
	transition := t.Normalize()
	if err := validateExecutionToken("safety shard transition zone id", transition.ZoneID); err != nil {
		return err
	}
	if transition.PreviousLayoutEpoch == 0 || transition.NextLayoutEpoch == 0 {
		return errors.New("safety shard layout epochs must be positive")
	}
	if transition.NextLayoutEpoch != transition.PreviousLayoutEpoch+1 {
		return errors.New("safety shard layout changes must advance one epoch")
	}
	if transition.CurrentHeight == 0 || transition.ActivationHeight == 0 || transition.EpochBoundaryHeight == 0 {
		return errors.New("safety shard transition heights must be positive")
	}
	if transition.ActivationHeight != transition.EpochBoundaryHeight {
		return errors.New("safety shard layout changes only at epoch boundaries")
	}
	if transition.CurrentHeight > transition.ActivationHeight {
		return errors.New("safety shard transition cannot activate in the past")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"safety shard split merge decision hash", transition.SplitMergeDecisionHash},
		{"safety shard migration root", transition.MigrationRoot},
		{"safety shard old layout hash", transition.OldLayoutHash},
		{"safety shard new layout hash", transition.NewLayoutHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if transition.OldLayoutHash == transition.NewLayoutHash {
		return errors.New("safety shard transition requires a new layout hash")
	}
	if transition.ProofHorizon == 0 {
		return errors.New("safety shard proof horizon must be positive")
	}
	for idx, msg := range transition.InFlightMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if msg.DeliveryEpoch != transition.NextLayoutEpoch {
			return errors.New("safety in-flight message delivery epoch must match next layout epoch")
		}
		if msg.ExpiryHeight < transition.ActivationHeight {
			return errors.New("safety in-flight message expires before layout activation")
		}
		if idx > 0 && transition.InFlightMessages[idx-1].MessageID >= msg.MessageID {
			return errors.New("safety in-flight messages must be sorted canonically")
		}
	}
	if transition.TransitionHash != ComputeShardLayoutTransitionHash(transition) {
		return errors.New("safety shard transition hash mismatch")
	}
	return nil
}

func (m InFlightShardMessage) Normalize() InFlightShardMessage {
	m.MessageID = normalizeLowerHex(m.MessageID)
	m.SourceShardID = strings.TrimSpace(m.SourceShardID)
	m.DestinationShardID = strings.TrimSpace(m.DestinationShardID)
	m.MessageHash = normalizeLowerHex(m.MessageHash)
	return m
}

func (m InFlightShardMessage) Validate() error {
	msg := m.Normalize()
	if err := validateHexHash("safety in-flight message id", msg.MessageID); err != nil {
		return err
	}
	if err := validateExecutionToken("safety in-flight source shard id", msg.SourceShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("safety in-flight destination shard id", msg.DestinationShardID); err != nil {
		return err
	}
	if msg.DeliveryEpoch == 0 || msg.ExpiryHeight == 0 {
		return errors.New("safety in-flight message epoch and expiry must be positive")
	}
	if msg.MessageHash != ComputeInFlightShardMessageHash(msg) {
		return errors.New("safety in-flight message hash mismatch")
	}
	return nil
}

func OldShardLayoutQueryable(queryHeight, activationHeight, proofHorizon uint64) bool {
	if queryHeight < activationHeight {
		return true
	}
	return queryHeight-activationHeight <= proofHorizon
}

func ComputeConsensusDeterminismPolicyHash(policy ConsensusDeterminismPolicy) string {
	policy = policy.Normalize()
	return hashStrings(
		"safety-determinism-policy",
		policy.ConsensusPathName,
		fmt.Sprintf("%t", policy.UsesExternalAPIs),
		fmt.Sprintf("%t", policy.UsesLocalClock),
		fmt.Sprintf("%t", policy.UsesRandomShardPlacement),
		fmt.Sprintf("%t", policy.UsesNondeterministicMapOrder),
		fmt.Sprintf("%t", policy.UsesMempoolOnlyExecutionData),
		fmt.Sprintf("%t", policy.UsesFloatingPointMath),
		fmt.Sprintf("%t", policy.UsesConsensusTime),
		fmt.Sprintf("%t", policy.SortsStateTransitionMaps),
		fmt.Sprintf("%t", policy.EncodesProposalMempoolInputs),
	)
}

func ComputeRoutingSafetyHash(input RoutingSafetyInput) string {
	input = input.Normalize()
	parts := []string{
		"safety-routing",
		fmt.Sprintf("%020d", input.RoutingEpoch),
		input.RoutingTableHash,
		input.RoutingMetricsRoot,
		input.SelectedPathID,
	}
	for _, path := range input.CandidatePaths {
		parts = append(parts, ComputeRoutingSafetyPathHash(path))
	}
	for _, accounting := range input.FailureAccounting {
		parts = append(parts, accounting.AccountingHash)
	}
	return hashStrings(parts...)
}

func ComputeRoutingSafetyPathHash(path RoutingSafetyPath) string {
	path = path.Normalize()
	return hashStrings(
		"safety-routing-path",
		path.PathID,
		path.RouteHash,
		path.Cost,
		fmt.Sprintf("%020d", uint64(path.HopCount)),
		fmt.Sprintf("%020d", uint64(path.LiquidityBps)),
		fmt.Sprintf("%020d", uint64(path.CongestionBps)),
		path.TieBreakKey,
	)
}

func ComputeRoutingFailureAccountingHash(accounting RoutingFailureAccounting) string {
	accounting = accounting.Normalize()
	return hashStrings(
		"safety-routing-failure-accounting",
		accounting.RouteID,
		accounting.OriginalValue,
		accounting.BouncedValue,
		accounting.BurnedValue,
		accounting.ReceiptHash,
		accounting.ReceiptStatus,
	)
}

func ComputeShardLayoutTransitionHash(transition ShardLayoutTransition) string {
	transition = transition.Normalize()
	parts := []string{
		"safety-shard-layout-transition",
		transition.ZoneID,
		fmt.Sprintf("%020d", transition.PreviousLayoutEpoch),
		fmt.Sprintf("%020d", transition.NextLayoutEpoch),
		fmt.Sprintf("%020d", transition.CurrentHeight),
		fmt.Sprintf("%020d", transition.ActivationHeight),
		fmt.Sprintf("%020d", transition.EpochBoundaryHeight),
		transition.SplitMergeDecisionHash,
		transition.MigrationRoot,
		transition.OldLayoutHash,
		transition.NewLayoutHash,
		fmt.Sprintf("%020d", transition.ProofHorizon),
	}
	for _, msg := range transition.InFlightMessages {
		parts = append(parts, msg.MessageHash)
	}
	return hashStrings(parts...)
}

func ComputeInFlightShardMessageHash(msg InFlightShardMessage) string {
	msg = msg.Normalize()
	return hashStrings(
		"safety-in-flight-shard-message",
		msg.MessageID,
		msg.SourceShardID,
		msg.DestinationShardID,
		fmt.Sprintf("%020d", msg.DeliveryEpoch),
		fmt.Sprintf("%020d", msg.ExpiryHeight),
	)
}

func compareRoutingSafetyPath(left, right RoutingSafetyPath) int {
	leftCost, _ := parsePerformanceNonNegativeInt("left routing cost", left.Cost)
	rightCost, _ := parsePerformanceNonNegativeInt("right routing cost", right.Cost)
	if !leftCost.Equal(rightCost) {
		if leftCost.LT(rightCost) {
			return -1
		}
		return 1
	}
	if left.HopCount != right.HopCount {
		if left.HopCount < right.HopCount {
			return -1
		}
		return 1
	}
	if left.CongestionBps != right.CongestionBps {
		if left.CongestionBps < right.CongestionBps {
			return -1
		}
		return 1
	}
	if left.LiquidityBps != right.LiquidityBps {
		if left.LiquidityBps > right.LiquidityBps {
			return -1
		}
		return 1
	}
	if left.TieBreakKey < right.TieBreakKey {
		return -1
	}
	if left.TieBreakKey > right.TieBreakKey {
		return 1
	}
	if left.PathID < right.PathID {
		return -1
	}
	if left.PathID > right.PathID {
		return 1
	}
	return 0
}
