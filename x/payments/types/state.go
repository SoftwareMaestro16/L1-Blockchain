package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

type PaymentsState struct {
	Channels        []ChannelRecord
	Edges           []ChannelEdge
	VirtualChannels []VirtualChannel
	Settlements     []SettlementRecord
	Batches         []SettlementBatch
}

func EmptyState() PaymentsState {
	return PaymentsState{
		Channels:        []ChannelRecord{},
		Edges:           []ChannelEdge{},
		VirtualChannels: []VirtualChannel{},
		Settlements:     []SettlementRecord{},
		Batches:         []SettlementBatch{},
	}
}

func OpenChannel(state PaymentsState, channel ChannelRecord) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	channel = channel.Normalize()
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments new channel must start open")
	}
	if _, found := state.ChannelByID(channel.ChannelID); found {
		return PaymentsState{}, errors.New("payments channel already exists")
	}
	if err := channel.LatestState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, err
	}
	if channel.OpeningStateHash == "" {
		channel.OpeningStateHash = channel.LatestState.StateHash
	}
	if channel.OpeningStateHash != channel.LatestState.StateHash {
		return PaymentsState{}, errors.New("payments opening state hash mismatch")
	}
	channel.FinalizedNonce = 0
	if err := channel.Validate(); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.Channels = append(next.Channels, channel)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func RegisterRoutingEdge(state PaymentsState, edge ChannelEdge) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	edge = edge.Normalize()
	if err := edge.Validate(); err != nil {
		return PaymentsState{}, err
	}
	channel, found := state.ChannelByID(edge.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments routing edge requires open channel")
	}
	if !containsString(channel.Participants, edge.From) || !containsString(channel.Participants, edge.To) {
		return PaymentsState{}, errors.New("payments routing edge endpoints must be channel participants")
	}
	if _, found := state.EdgeByKey(edge.ChannelID, edge.From, edge.To); found {
		return PaymentsState{}, errors.New("payments routing edge already exists")
	}
	next := state.Clone()
	next.Edges = append(next.Edges, edge)
	sortEdges(next.Edges)
	return next, next.Validate()
}

func SubmitClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	pending := PendingClose{
		Submitter:         submitter,
		SubmittedHeight:   currentHeight,
		SettleAfterHeight: currentHeight + channel.DisputePeriod,
		SettlementFee:     settlementFee,
		State:             closingState.Normalize(),
	}
	if err := pending.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if pending.State.Nonce < channel.FinalizedNonce {
		return PaymentsState{}, errors.New("payments close state nonce is below finalized nonce")
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusPendingClose
	nextChannel.PendingClose = pending
	nextChannel.LatestState = pending.State
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func DisputeClose(state PaymentsState, channelID string, newerState ChannelState, submitter string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments dispute height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments channel is not pending close")
	}
	if currentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments dispute window has closed")
	}
	newerState = newerState.Normalize()
	if err := newerState.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if newerState.Nonce <= channel.PendingClose.State.Nonce {
		return PaymentsState{}, errors.New("payments dispute state nonce must be newer")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments dispute submitter must be participant")
	}
	nextChannel := channel
	nextChannel.PendingClose.State = newerState
	nextChannel.PendingClose.SubmittedHeight = currentHeight
	nextChannel.PendingClose.SettleAfterHeight = currentHeight + channel.DisputePeriod
	nextChannel.LatestState = newerState
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func SubmitFraudProof(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments fraud proof height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments fraud proof requires pending close")
	}
	if currentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments fraud proof window has closed")
	}
	proof = proof.Normalize()
	if err := proof.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range channel.PendingClose.FraudProofs {
		if existing.ProofID == proof.ProofID {
			return PaymentsState{}, errors.New("payments duplicate fraud proof")
		}
	}
	penalty := Penalty{Offender: proof.OffendingSigner, Recipient: proof.SubmittedBy, Amount: proof.PenaltyAmount}.Normalize()
	if err := penalty.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, proof)
	nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalty)
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func FinalizeSettlement(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not pending close")
	}
	if currentHeight < channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement is still in dispute window")
	}
	finalBalances, err := applySettlementAdjustments(channel.PendingClose.State.Balances, channel.PendingClose.Penalties, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChannelID:     channel.ChannelID,
		StateHash:     channel.PendingClose.State.StateHash,
		Nonce:         channel.PendingClose.State.Nonce,
		FinalBalances: finalBalances,
		SettlementFee: channel.PendingClose.SettlementFee,
		Penalties:     channel.PendingClose.Penalties,
		SettledHeight: currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	return next, settlement, next.Validate()
}

func OpenVirtualChannel(state PaymentsState, vc VirtualChannel) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	vc = vc.Normalize()
	if vc.AnchorCommitment == "" {
		vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	}
	if err := vc.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if _, found := state.VirtualChannelByID(vc.VirtualChannelID); found {
		return PaymentsState{}, errors.New("payments virtual channel already exists")
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return PaymentsState{}, err
	}
	for _, parentID := range vc.ParentChannelIDs {
		channel, found := state.ChannelByID(parentID)
		if !found || channel.Status != ChannelStatusOpen {
			return PaymentsState{}, errors.New("payments virtual channel requires open parents")
		}
		if !containsString(channel.Participants, vc.Endpoints[0]) && !containsString(channel.Participants, vc.Endpoints[1]) {
			return PaymentsState{}, errors.New("payments virtual channel parent path must touch an endpoint")
		}
		if channelCapacity, err := parsePositiveInt("payments channel collateral", channel.Collateral); err != nil {
			return PaymentsState{}, err
		} else if channelCapacity.LT(capacity) {
			return PaymentsState{}, errors.New("payments virtual channel capacity exceeds parent capacity")
		}
	}
	next := state.Clone()
	next.VirtualChannels = append(next.VirtualChannels, vc)
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func AddSettlementBatch(state PaymentsState, batch SettlementBatch) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	batch = batch.Normalize()
	if err := batch.Validate(); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range state.Batches {
		if existing.BatchID == batch.BatchID {
			return PaymentsState{}, errors.New("payments settlement batch already exists")
		}
	}
	for _, op := range batch.Operations {
		channel, found := state.ChannelByID(op.ChannelID)
		if !found {
			return PaymentsState{}, errors.New("payments settlement batch references unknown channel")
		}
		if op.Nonce < channel.FinalizedNonce {
			return PaymentsState{}, errors.New("payments settlement batch operation nonce below finalized nonce")
		}
	}
	next := state.Clone()
	next.Batches = append(next.Batches, batch)
	sortBatches(next.Batches)
	return next, next.Validate()
}

func RoutePayment(state PaymentsState, from, to, amountText string, currentHeight uint64, maxHops int) ([]ChannelEdge, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return nil, err
	}
	amount, err := parsePositiveInt("payments route amount", amountText)
	if err != nil {
		return nil, err
	}
	if maxHops <= 0 || maxHops > MaxRoutingHops {
		maxHops = MaxRoutingHops
	}
	candidates := activeEdgesForAmount(state.Edges, amount, currentHeight)
	sortEdges(candidates)
	type path struct {
		node  string
		edges []ChannelEdge
	}
	queue := []path{{node: from}}
	visitedDepth := map[string]int{from: 0}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if len(current.edges) >= maxHops {
			continue
		}
		for _, edge := range candidates {
			if edge.From != current.node {
				continue
			}
			nextEdges := append([]ChannelEdge(nil), current.edges...)
			nextEdges = append(nextEdges, edge)
			if edge.To == to {
				return nextEdges, nil
			}
			if depth, seen := visitedDepth[edge.To]; seen && depth <= len(nextEdges) {
				continue
			}
			visitedDepth[edge.To] = len(nextEdges)
			queue = append(queue, path{node: edge.To, edges: nextEdges})
		}
	}
	return nil, errors.New("payments route not found")
}

func ImportState(state PaymentsState) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	return state, nil
}

func (s PaymentsState) Export() PaymentsState {
	out := s.Clone()
	sortChannels(out.Channels)
	sortEdges(out.Edges)
	sortVirtualChannels(out.VirtualChannels)
	sortSettlements(out.Settlements)
	sortBatches(out.Batches)
	return out
}

func (s PaymentsState) Clone() PaymentsState {
	out := PaymentsState{
		Channels:        make([]ChannelRecord, len(s.Channels)),
		Edges:           make([]ChannelEdge, len(s.Edges)),
		VirtualChannels: make([]VirtualChannel, len(s.VirtualChannels)),
		Settlements:     make([]SettlementRecord, len(s.Settlements)),
		Batches:         make([]SettlementBatch, len(s.Batches)),
	}
	for i, channel := range s.Channels {
		out.Channels[i] = channel.Normalize()
	}
	for i, edge := range s.Edges {
		out.Edges[i] = edge.Normalize()
	}
	for i, vc := range s.VirtualChannels {
		out.VirtualChannels[i] = vc.Normalize()
	}
	for i, settlement := range s.Settlements {
		out.Settlements[i] = settlement.Normalize()
	}
	for i, batch := range s.Batches {
		out.Batches[i] = batch.Normalize()
	}
	return out
}

func (s PaymentsState) Validate() error {
	if err := validateChannels(s.Channels); err != nil {
		return err
	}
	if err := validateEdges(s.Channels, s.Edges); err != nil {
		return err
	}
	if err := validateVirtualChannels(s.Channels, s.VirtualChannels); err != nil {
		return err
	}
	if err := validateSettlements(s.Channels, s.Settlements); err != nil {
		return err
	}
	return validateBatches(s.Channels, s.Batches)
}

func (s PaymentsState) ChannelByID(channelID string) (ChannelRecord, bool) {
	_, channel, found := s.ChannelIndex(channelID)
	return channel, found
}

func (s PaymentsState) ChannelIndex(channelID string) (int, ChannelRecord, bool) {
	needle := normalizeHash(channelID)
	for i, channel := range s.Channels {
		channel = channel.Normalize()
		if channel.ChannelID == needle {
			return i, channel, true
		}
	}
	return 0, ChannelRecord{}, false
}

func (s PaymentsState) EdgeByKey(channelID, from, to string) (ChannelEdge, bool) {
	channelID = normalizeHash(channelID)
	for _, edge := range s.Edges {
		edge = edge.Normalize()
		if edge.ChannelID == channelID && edge.From == from && edge.To == to {
			return edge, true
		}
	}
	return ChannelEdge{}, false
}

func (s PaymentsState) VirtualChannelByID(id string) (VirtualChannel, bool) {
	needle := normalizeHash(id)
	for _, vc := range s.VirtualChannels {
		vc = vc.Normalize()
		if vc.VirtualChannelID == needle {
			return vc, true
		}
	}
	return VirtualChannel{}, false
}

func validateChannels(channels []ChannelRecord) error {
	seen := make(map[string]struct{}, len(channels))
	var previous string
	for i, channel := range channels {
		channel = channel.Normalize()
		if err := channel.Validate(); err != nil {
			return err
		}
		if _, found := seen[channel.ChannelID]; found {
			return errors.New("payments duplicate channel")
		}
		seen[channel.ChannelID] = struct{}{}
		if i > 0 && previous >= channel.ChannelID {
			return errors.New("payments channels must be sorted canonically")
		}
		previous = channel.ChannelID
	}
	return nil
}

func validateEdges(channels []ChannelRecord, edges []ChannelEdge) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(edges))
	var previous string
	for i, edge := range edges {
		edge = edge.Normalize()
		if err := edge.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[edge.ChannelID]
		if !found {
			return errors.New("payments routing edge references unknown channel")
		}
		if channel.Status != ChannelStatusOpen {
			return errors.New("payments routing edge references non-open channel")
		}
		if !containsString(channel.Participants, edge.From) || !containsString(channel.Participants, edge.To) {
			return errors.New("payments routing edge endpoints must be channel participants")
		}
		key := edgeKey(edge)
		if _, found := seen[key]; found {
			return errors.New("payments duplicate routing edge")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("payments routing edges must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateVirtualChannels(channels []ChannelRecord, virtualChannels []VirtualChannel) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(virtualChannels))
	var previous string
	for i, vc := range virtualChannels {
		vc = vc.Normalize()
		if err := vc.Validate(); err != nil {
			return err
		}
		for _, parentID := range vc.ParentChannelIDs {
			if _, found := channelByID[parentID]; !found {
				return errors.New("payments virtual channel references unknown parent")
			}
		}
		if _, found := seen[vc.VirtualChannelID]; found {
			return errors.New("payments duplicate virtual channel")
		}
		seen[vc.VirtualChannelID] = struct{}{}
		if i > 0 && previous >= vc.VirtualChannelID {
			return errors.New("payments virtual channels must be sorted canonically")
		}
		previous = vc.VirtualChannelID
	}
	return nil
}

func validateSettlements(channels []ChannelRecord, settlements []SettlementRecord) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(settlements))
	var previous string
	for i, settlement := range settlements {
		settlement = settlement.Normalize()
		channel, found := channelByID[settlement.ChannelID]
		if !found {
			return errors.New("payments settlement references unknown channel")
		}
		if err := settlement.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[settlement.ChannelID]; found {
			return errors.New("payments duplicate settlement")
		}
		seen[settlement.ChannelID] = struct{}{}
		if i > 0 && previous >= settlement.ChannelID {
			return errors.New("payments settlements must be sorted canonically")
		}
		previous = settlement.ChannelID
	}
	return nil
}

func validateBatches(channels []ChannelRecord, batches []SettlementBatch) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(batches))
	var previous string
	for i, batch := range batches {
		batch = batch.Normalize()
		if err := batch.Validate(); err != nil {
			return err
		}
		for _, op := range batch.Operations {
			if _, found := channelByID[op.ChannelID]; !found {
				return errors.New("payments batch references unknown channel")
			}
		}
		if _, found := seen[batch.BatchID]; found {
			return errors.New("payments duplicate batch")
		}
		seen[batch.BatchID] = struct{}{}
		if i > 0 && previous >= batch.BatchID {
			return errors.New("payments batches must be sorted canonically")
		}
		previous = batch.BatchID
	}
	return nil
}

func applySettlementAdjustments(balances []Balance, penalties []Penalty, feeText, feePayer string) ([]Balance, error) {
	amounts := make(map[string]sdkmath.Int, len(balances))
	for _, balance := range normalizeBalances(balances) {
		amount, err := parseNonNegativeInt("payments final balance", balance.Amount)
		if err != nil {
			return nil, err
		}
		amounts[balance.Participant] = amount
	}
	for _, penalty := range normalizePenalties(penalties) {
		amount, err := parsePositiveInt("payments penalty amount", penalty.Amount)
		if err != nil {
			return nil, err
		}
		offenderBalance, found := amounts[penalty.Offender]
		if !found || offenderBalance.LT(amount) {
			return nil, errors.New("payments penalty exceeds offender balance")
		}
		amounts[penalty.Offender] = offenderBalance.Sub(amount)
		amounts[penalty.Recipient] = amounts[penalty.Recipient].Add(amount)
	}
	fee, err := parseNonNegativeInt("payments settlement fee", feeText)
	if err != nil {
		return nil, err
	}
	if fee.IsPositive() {
		balance, found := amounts[feePayer]
		if !found || balance.LT(fee) {
			return nil, errors.New("payments settlement fee exceeds payer balance")
		}
		amounts[feePayer] = balance.Sub(fee)
	}
	out := make([]Balance, 0, len(amounts))
	for participant, amount := range amounts {
		out = append(out, Balance{Participant: participant, Amount: amount.String()})
	}
	return normalizeBalances(out), nil
}

func activeEdgesForAmount(edges []ChannelEdge, amount sdkmath.Int, currentHeight uint64) []ChannelEdge {
	out := make([]ChannelEdge, 0, len(edges))
	for _, edge := range edges {
		edge = edge.Normalize()
		capacity, err := parsePositiveInt("payments routing capacity", edge.Capacity)
		if err != nil {
			continue
		}
		if !edge.Active || capacity.LT(amount) {
			continue
		}
		if edge.ExpiresHeight > 0 && currentHeight > edge.ExpiresHeight {
			continue
		}
		out = append(out, edge)
	}
	return out
}

func filterEdgesForSettledChannel(edges []ChannelEdge, channelID string) []ChannelEdge {
	channelID = normalizeHash(channelID)
	out := make([]ChannelEdge, 0, len(edges))
	for _, edge := range edges {
		if edge.Normalize().ChannelID == channelID {
			continue
		}
		out = append(out, edge)
	}
	return out
}

func channelMap(channels []ChannelRecord) map[string]ChannelRecord {
	out := make(map[string]ChannelRecord, len(channels))
	for _, channel := range channels {
		channel = channel.Normalize()
		out[channel.ChannelID] = channel
	}
	return out
}

func sortChannels(channels []ChannelRecord) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Normalize().ChannelID < channels[j].Normalize().ChannelID
	})
}

func sortEdges(edges []ChannelEdge) {
	sort.SliceStable(edges, func(i, j int) bool {
		return edgeKey(edges[i].Normalize()) < edgeKey(edges[j].Normalize())
	})
}

func sortVirtualChannels(channels []VirtualChannel) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Normalize().VirtualChannelID < channels[j].Normalize().VirtualChannelID
	})
}

func sortSettlements(settlements []SettlementRecord) {
	sort.SliceStable(settlements, func(i, j int) bool {
		return settlements[i].Normalize().ChannelID < settlements[j].Normalize().ChannelID
	})
}

func sortBatches(batches []SettlementBatch) {
	sort.SliceStable(batches, func(i, j int) bool {
		return batches[i].Normalize().BatchID < batches[j].Normalize().BatchID
	})
}

func edgeKey(edge ChannelEdge) string {
	return fmt.Sprintf("%s/%s/%s", edge.ChannelID, edge.From, edge.To)
}
