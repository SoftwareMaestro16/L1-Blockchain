package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type PaymentsState struct {
	Channels			[]ChannelRecord
	Edges				[]ChannelEdge
	VirtualChannels			[]VirtualChannel
	Settlements			[]SettlementRecord
	Batches				[]SettlementBatch
	CustodyLocks			[]CustodyLock
	ClosedChannels			[]ClosedChannelTombstone
	ConditionClaims			[]ConditionClaimRecord
	ValidatorPaymentServices	[]ValidatorPaymentServiceMetadata
	ValidatorWatchRegistries	[]ValidatorWatchRegistration
	FeeSchedule			PaymentFeeSchedule
	FeeMultipliers			[]PaymentFeeMultiplier
	FeeCharges			[]PaymentFeeCharge
	FeeRefunds			[]PaymentFeeRefund
	SecurityReserveHooks		[]SecurityReserveAllocationHook
	InclusionLatencies		[]SettlementInclusionLatency
	AsyncFinalizationQueue		[]AsyncFinalizationJob
	AsyncPromiseExpiryQueue		[]AsyncPromiseExpiryJob
	AsyncCompletions		[]AsyncSettlementCompletion
	Events				[]PaymentEvent
}

func EmptyState() PaymentsState {
	return PaymentsState{
		Channels:			[]ChannelRecord{},
		Edges:				[]ChannelEdge{},
		VirtualChannels:		[]VirtualChannel{},
		Settlements:			[]SettlementRecord{},
		Batches:			[]SettlementBatch{},
		CustodyLocks:			[]CustodyLock{},
		ClosedChannels:			[]ClosedChannelTombstone{},
		ConditionClaims:		[]ConditionClaimRecord{},
		ValidatorPaymentServices:	[]ValidatorPaymentServiceMetadata{},
		ValidatorWatchRegistries:	[]ValidatorWatchRegistration{},
		FeeSchedule:			DefaultPaymentFeeSchedule(),
		FeeMultipliers:			[]PaymentFeeMultiplier{},
		FeeCharges:			[]PaymentFeeCharge{},
		FeeRefunds:			[]PaymentFeeRefund{},
		SecurityReserveHooks:		[]SecurityReserveAllocationHook{},
		InclusionLatencies:		[]SettlementInclusionLatency{},
		AsyncFinalizationQueue:		[]AsyncFinalizationJob{},
		AsyncPromiseExpiryQueue:	[]AsyncPromiseExpiryJob{},
		AsyncCompletions:		[]AsyncSettlementCompletion{},
		Events:				[]PaymentEvent{},
	}
}

func setChannelFinality(channel ChannelRecord, finality ChannelFinality, height uint64, events *[]PaymentEvent) (ChannelRecord, error) {
	channel = channel.Normalize()
	if height == 0 {
		return ChannelRecord{}, errors.New("payments finality transition height must be positive")
	}
	if !IsChannelFinality(finality) {
		return ChannelRecord{}, fmt.Errorf("unknown payments channel finality %q", finality)
	}
	previous := channel.Finality
	if previous == finality {
		return channel, nil
	}
	channel.Finality = finality
	if err := validateChannelFinalityForStatus(channel); err != nil {
		return ChannelRecord{}, err
	}
	if events != nil {
		*events = append(*events, ChannelFinalityTransitionEvent(channel, previous, finality, height))
	}
	return channel.Normalize(), nil
}

func finalityForPendingClose(channel ChannelRecord) ChannelFinality {
	return DerivedChannelFinality(channel)
}

func finalityForSettledChannel(channel ChannelRecord) ChannelFinality {
	channel = channel.Normalize()
	if len(channel.PendingClose.Penalties) > 0 || len(channel.PendingClose.PenaltyAllocations) > 0 {
		return ChannelFinalityPenalized
	}
	return ChannelFinalitySettled
}

func ConfigurePaymentFeeSchedule(state PaymentsState, schedule PaymentFeeSchedule) (PaymentsState, error) {
	state = state.Export()
	schedule = schedule.Normalize()
	if err := schedule.Validate(); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.FeeSchedule = schedule
	return next, next.Validate()
}

func SetPaymentFeeMultiplier(state PaymentsState, multiplier PaymentFeeMultiplier) (PaymentsState, error) {
	state = state.Export()
	multiplier = multiplier.Normalize()
	if multiplier.UpdatedHeight == 0 {
		return PaymentsState{}, errors.New("payments fee multiplier height must be positive")
	}
	if err := multiplier.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if multiplier.MultiplierBps > state.FeeSchedule.Normalize().MaxMultiplierBps {
		return PaymentsState{}, errors.New("payments fee multiplier exceeds schedule maximum")
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.FeeMultipliers {
		if existing.Normalize().FeeClass == multiplier.FeeClass {
			next.FeeMultipliers[i] = multiplier
			replaced = true
			break
		}
	}
	if !replaced {
		next.FeeMultipliers = append(next.FeeMultipliers, multiplier)
	}
	sortPaymentFeeMultipliers(next.FeeMultipliers)
	return next, next.Validate()
}

func RequiredPaymentFee(state PaymentsState, feeClass PaymentFeeClass, channel ChannelRecord) (string, uint64, uint32, error) {
	state = state.Export()
	schedule := state.FeeSchedule.Normalize()
	if err := schedule.Validate(); err != nil {
		return "", 0, 0, err
	}
	if feeClass == PaymentFeeClassChannelOpen {
		formula, err := ComputeChannelOpenFeeFormula(state, channel)
		if err != nil {
			return "", 0, 0, err
		}
		return formula.TotalFee, formula.StorageBytes, formula.MultiplierBps, nil
	}
	baseText, err := paymentFeeBaseAmount(schedule, feeClass)
	if err != nil {
		return "", 0, 0, err
	}
	base, err := parseNonNegativeInt("payments base fee", baseText)
	if err != nil {
		return "", 0, 0, err
	}
	storageBytes := paymentStorageFootprint(feeClass, channel)
	if schedule.StorageFeeEnabled && storageBytes > 0 {
		byteFee, err := parseNonNegativeInt("payments storage byte fee", schedule.StorageByteFee)
		if err != nil {
			return "", 0, 0, err
		}
		base = base.Add(byteFee.Mul(sdkmath.NewIntFromUint64(storageBytes)))
	}
	multiplier := feeMultiplierForClass(state, feeClass, schedule)
	required := base.Mul(sdkmath.NewInt(int64(multiplier)))
	denom := sdkmath.NewInt(10_000)
	if !required.IsZero() {
		required = required.Add(denom.Sub(sdkmath.OneInt())).Quo(denom)
	}
	return required.String(), storageBytes, multiplier, nil
}

func ComputeChannelOpenFeeFormula(state PaymentsState, channel ChannelRecord) (ChannelOpenFeeFormula, error) {
	state = state.Export()
	schedule := state.FeeSchedule.Normalize()
	if err := schedule.Validate(); err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	channel = channel.Normalize()
	base, err := parseNonNegativeInt("payments channel open base fee", schedule.ChannelOpenFee)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	perParticipant, err := parseNonNegativeInt("payments channel open per participant fee", schedule.ChannelOpenPerParticipantFee)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	participantCount := uint64(len(channel.Participants))
	participantFee := perParticipant.Mul(sdkmath.NewIntFromUint64(participantCount))
	storageBytes := EstimateChannelOpenStorageFootprint(channel)
	byteFee, err := parseNonNegativeInt("payments channel open storage byte fee", schedule.StorageByteFee)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	storageFee := sdkmath.ZeroInt()
	if schedule.StorageFeeEnabled {
		storageFee = byteFee.Mul(sdkmath.NewIntFromUint64(storageBytes))
	}
	conditionalSurcharge, err := parseNonNegativeInt("payments conditional capability surcharge", schedule.ConditionalCapabilitySurcharge)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	if !channel.ConditionalPayments {
		conditionalSurcharge = sdkmath.ZeroInt()
	}
	if _, err := parseNonNegativeInt("payments virtual channel anchor surcharge", schedule.VirtualChannelAnchorSurcharge); err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	virtualSurcharge := sdkmath.ZeroInt()
	routingDeposit, err := parseNonNegativeInt("payments routing advertisement deposit", schedule.RoutingAdvertisementDeposit)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	if !channel.RoutingAdvertised {
		routingDeposit = sdkmath.ZeroInt()
	}
	rentReserve := sdkmath.ZeroInt()
	if schedule.RenewalPeriod > 0 {
		rentPerBlock, err := parseNonNegativeInt("payments storage rent per block", schedule.StorageRentPerBlock)
		if err != nil {
			return ChannelOpenFeeFormula{}, err
		}
		rentReserve = rentPerBlock.Mul(sdkmath.NewIntFromUint64(schedule.RenewalPeriod))
	}
	subtotal := base.Add(participantFee).Add(storageFee).Add(conditionalSurcharge).Add(virtualSurcharge).Add(routingDeposit).Add(rentReserve)
	minFee, err := parseNonNegativeInt("payments open fee minimum", schedule.OpenFeeMin)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	if subtotal.LT(minFee) {
		subtotal = minFee
	}
	maxFee, err := parseNonNegativeInt("payments open fee maximum", schedule.OpenFeeMax)
	if err != nil {
		return ChannelOpenFeeFormula{}, err
	}
	if !maxFee.IsZero() && subtotal.GT(maxFee) {
		subtotal = maxFee
	}
	multiplier := feeMultiplierForClass(state, PaymentFeeClassChannelOpen, schedule)
	total := subtotal.Mul(sdkmath.NewInt(int64(multiplier)))
	denom := sdkmath.NewInt(10_000)
	if !total.IsZero() {
		total = total.Add(denom.Sub(sdkmath.OneInt())).Quo(denom)
	}
	return ChannelOpenFeeFormula{
		Denom:			NativeDenom,
		BaseFee:		base.String(),
		ParticipantFee:		participantFee.String(),
		ParticipantCount:	participantCount,
		StorageByteFee:		byteFee.String(),
		StorageBytes:		storageBytes,
		StorageFee:		storageFee.String(),
		ConditionalSurcharge:	conditionalSurcharge.String(),
		VirtualAnchorSurcharge:	virtualSurcharge.String(),
		RoutingDeposit:		routingDeposit.String(),
		RentReserve:		rentReserve.String(),
		MultiplierBps:		multiplier,
		MinFee:			schedule.OpenFeeMin,
		MaxFee:			schedule.OpenFeeMax,
		TotalFee:		total.String(),
	}, nil
}

func ChargePaymentFee(state PaymentsState, feeClass PaymentFeeClass, channel ChannelRecord, payer, objectID, amountPaid string, height uint64) (PaymentsState, PaymentFeeCharge, error) {
	state = state.Export()
	if height == 0 {
		return PaymentsState{}, PaymentFeeCharge{}, errors.New("payments fee charge height must be positive")
	}
	if err := addressing.ValidateUserAddress("payments fee payer", payer); err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	amountPaid = strings.TrimSpace(amountPaid)
	if amountPaid == "" {
		amountPaid = "0"
	}
	if err := validateNonNegativeInt("payments fee paid", amountPaid); err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	required, storageBytes, multiplier, err := RequiredPaymentFee(state, feeClass, channel)
	if err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	paid, err := parseNonNegativeInt("payments fee paid", amountPaid)
	if err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	requiredAmount, err := parseNonNegativeInt("payments required fee", required)
	if err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	if paid.LT(requiredAmount) {
		return PaymentsState{}, PaymentFeeCharge{}, fmt.Errorf("payments %s fee below required amount %s", feeClass, required)
	}
	channelID := normalizeOptionalHash(channel.ChannelID)
	objectID = strings.TrimSpace(objectID)
	charge := PaymentFeeCharge{
		FeeID:		HashParts("payment-fee", string(feeClass), channelID, objectID, payer, amountPaid, fmt.Sprintf("%020d", height)),
		FeeClass:	feeClass,
		ChannelID:	channelID,
		ObjectID:	objectID,
		Payer:		strings.TrimSpace(payer),
		Denom:		NativeDenom,
		Amount:		amountPaid,
		RequiredAmount:	required,
		StorageBytes:	storageBytes,
		MultiplierBps:	multiplier,
		Height:		height,
	}.Normalize()
	if err := charge.Validate(); err != nil {
		return PaymentsState{}, PaymentFeeCharge{}, err
	}
	next := state.Clone()
	next.FeeCharges = append(next.FeeCharges, charge)
	sortPaymentFeeCharges(next.FeeCharges)
	return next, charge, next.Validate()
}

func RefundPaymentFee(state PaymentsState, feeID, recipient, reason string, height uint64) (PaymentsState, PaymentFeeRefund, error) {
	state = state.Export()
	feeID = normalizeHash(feeID)
	index := -1
	var charge PaymentFeeCharge
	for i, existing := range state.FeeCharges {
		existing = existing.Normalize()
		if existing.FeeID == feeID {
			index = i
			charge = existing
			break
		}
	}
	if index < 0 {
		return PaymentsState{}, PaymentFeeRefund{}, errors.New("payments fee charge not found")
	}
	if charge.Refunded {
		return PaymentsState{}, PaymentFeeRefund{}, errors.New("payments fee already refunded")
	}
	if charge.Amount == "0" {
		return state, PaymentFeeRefund{}, nil
	}
	refund := PaymentFeeRefund{
		RefundID:	HashParts("payment-fee-refund", feeID, recipient, reason, fmt.Sprintf("%020d", height)),
		FeeID:		feeID,
		Recipient:	recipient,
		Denom:		NativeDenom,
		Amount:		charge.Amount,
		Reason:		reason,
		Height:		height,
	}.Normalize()
	if err := refund.Validate(); err != nil {
		return PaymentsState{}, PaymentFeeRefund{}, err
	}
	next := state.Clone()
	next.FeeCharges[index].Refunded = true
	next.FeeRefunds = append(next.FeeRefunds, refund)
	sortPaymentFeeCharges(next.FeeCharges)
	sortPaymentFeeRefunds(next.FeeRefunds)
	return next, refund, next.Validate()
}

func EstimateSettlementMessageGas(input SettlementArbitrationInput, schedule SettlementGasCostSchedule) (SettlementGasEstimate, error) {
	input = input.Normalize()
	schedule = schedule.Normalize()
	if err := schedule.Validate(); err != nil {
		return SettlementGasEstimate{}, err
	}
	if !IsSettlementArbitrationOperation(input.Operation) {
		return SettlementGasEstimate{}, fmt.Errorf("unknown payments settlement gas operation %q", input.Operation)
	}
	base, err := settlementBaseGasForOperation(input.Operation, schedule)
	if err != nil {
		return SettlementGasEstimate{}, err
	}
	signatureCount := uint64(len(input.SignedState.Normalize().Signatures))
	if !input.Claim.IsZero() {
		signatureCount++
		if input.Claim.ReceiverAckOptional.SignatureHash != "" {
			signatureCount++
		}
	}
	conditionCount := uint64(len(input.ConditionProofs))
	fraudProofCount := uint64(0)
	if input.FraudProof.Normalize().ProofID != "" {
		fraudProofCount = 1
	}
	penaltyAllocationCount := uint64(0)
	if input.Operation == SettlementArbitrationPenaltyRouting && input.FraudProof.Normalize().ProofID != "" {
		penaltyAllocationCount = 1
	}
	stateBytes := estimateSettlementStateBytes(input)
	estimate := SettlementGasEstimate{
		Operation:		input.Operation,
		BaseGas:		base,
		SignatureGas:		signatureCount * schedule.PerSignatureGas,
		ConditionGas:		conditionCount * schedule.PerConditionGas,
		FraudProofGas:		fraudProofCount * schedule.PerFraudProofGas,
		PenaltyAllocationGas:	penaltyAllocationCount * schedule.PerPenaltyAllocationGas,
		StateByteGas:		stateBytes * schedule.PerStateByteGas,
	}
	estimate.TotalGas = estimate.BaseGas + estimate.SignatureGas + estimate.ConditionGas + estimate.FraudProofGas + estimate.PenaltyAllocationGas + estimate.StateByteGas
	return estimate, nil
}

func RecordSecurityReserveAllocationHooks(state PaymentsState, channelID string, proof FraudProof, allocations []PenaltyAllocation, height uint64, enabled bool) (PaymentsState, []SecurityReserveAllocationHook, error) {
	state = state.Export()
	if !enabled {
		return state, nil, nil
	}
	channel, found := state.ChannelByID(channelID)
	if !found {
		return PaymentsState{}, nil, errors.New("payments security reserve hook channel not found")
	}
	hooks, err := BuildSecurityReserveAllocationHooks(channel, proof, allocations, height)
	if err != nil {
		return PaymentsState{}, nil, err
	}
	if len(hooks) == 0 {
		return state, nil, nil
	}
	next := state.Clone()
	next.SecurityReserveHooks = append(next.SecurityReserveHooks, hooks...)
	sortSecurityReserveAllocationHooks(next.SecurityReserveHooks)
	return next, hooks, next.Validate()
}

func BuildSecurityReserveAllocationHooks(channel ChannelRecord, proof FraudProof, allocations []PenaltyAllocation, height uint64) ([]SecurityReserveAllocationHook, error) {
	channel = channel.Normalize()
	proof = proof.Normalize()
	if height == 0 {
		return nil, errors.New("payments security reserve hook height must be positive")
	}
	out := []SecurityReserveAllocationHook{}
	for _, allocation := range normalizePenaltyAllocations(allocations) {
		if allocation.Route != PenaltyRouteSecurityReserve {
			continue
		}
		commitment := HashParts("security-reserve-allocation", channel.ChannelID, proof.ProofID, allocation.Offender, allocation.Amount, fmt.Sprintf("%020d", height))
		hook := SecurityReserveAllocationHook{
			HookID:		HashParts("security-reserve-hook", commitment),
			ChannelID:	channel.ChannelID,
			ProofID:	proof.ProofID,
			Offender:	allocation.Offender,
			Denom:		NativeDenom,
			Amount:		allocation.Amount,
			Height:		height,
			Route:		PenaltyRouteSecurityReserve,
			Allocation:	commitment,
		}.Normalize()
		if err := hook.ValidateForChannel(channel); err != nil {
			return nil, err
		}
		out = append(out, hook)
	}
	sortSecurityReserveAllocationHooks(out)
	return out, nil
}

func RecordSettlementInclusionLatency(state PaymentsState, operationID, channelID string, operation SettlementArbitrationOperation, submittedHeight, includedHeight, sloThreshold uint64) (PaymentsState, SettlementInclusionLatency, error) {
	state = state.Export()
	channelID = normalizeHash(channelID)
	if _, found := state.ChannelByID(channelID); !found {
		return PaymentsState{}, SettlementInclusionLatency{}, errors.New("payments inclusion latency channel not found")
	}
	record := SettlementInclusionLatency{
		RecordID:		HashParts("settlement-inclusion-latency", operationID, channelID, string(operation), fmt.Sprintf("%020d", submittedHeight), fmt.Sprintf("%020d", includedHeight)),
		OperationID:		operationID,
		ChannelID:		channelID,
		Operation:		operation,
		SubmittedHeight:	submittedHeight,
		IncludedHeight:		includedHeight,
		SLOThreshold:		sloThreshold,
	}.Normalize()
	if err := record.Validate(state.Channels); err != nil {
		return PaymentsState{}, SettlementInclusionLatency{}, err
	}
	next := state.Clone()
	next.InclusionLatencies = append(next.InclusionLatencies, record)
	sortSettlementInclusionLatencies(next.InclusionLatencies)
	return next, record, next.Validate()
}

func settlementBaseGasForOperation(operation SettlementArbitrationOperation, schedule SettlementGasCostSchedule) (uint64, error) {
	switch operation {
	case SettlementArbitrationOpen, SettlementArbitrationCollateralCustody:
		return schedule.OpenGas, nil
	case SettlementArbitrationCooperativeClose:
		return schedule.CooperativeCloseGas, nil
	case SettlementArbitrationUnilateralClose:
		return schedule.UnilateralCloseGas, nil
	case SettlementArbitrationDispute:
		return schedule.DisputeGas, nil
	case SettlementArbitrationFraudProof:
		return schedule.FraudProofGas, nil
	case SettlementArbitrationConditionResolution:
		return schedule.ConditionResolutionGas, nil
	case SettlementArbitrationPenaltyRouting:
		return schedule.PenaltyRoutingGas, nil
	case SettlementArbitrationFinalSettlement:
		return schedule.FinalSettlementGas, nil
	case SettlementArbitrationReplayProtection:
		return schedule.ReplayProtectionGas, nil
	default:
		return 0, fmt.Errorf("unknown payments settlement gas operation %q", operation)
	}
}

func estimateSettlementStateBytes(input SettlementArbitrationInput) uint64 {
	input = input.Normalize()
	state := input.SignedState.Normalize()
	size := uint64(len(input.ChannelID) + len(state.StateHash) + len(state.PreviousStateHash) + len(state.ConditionRoot) + len(state.ParticipantSetHash))
	size += uint64(len(state.Balances) * 80)
	size += uint64(len(state.Signatures) * 128)
	size += uint64(len(input.ConditionProofs) * 96)
	proof := input.FraudProof.Normalize()
	if proof.ProofID != "" {
		size += uint64(len(proof.ProofID) + len(proof.EvidenceHash) + len(proof.PenaltyAmount) + 256)
	}
	if !input.Claim.IsZero() {
		claim := input.Claim.Normalize()
		size += uint64(len(claim.StateHash) + len(claim.ClaimedAmount) + 128)
	}
	if size == 0 {
		return 1
	}
	return size
}

func RefreshAsyncExecutionQueues(state PaymentsState, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments async queue refresh height must be positive")
	}
	next := state.Clone()
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		finalizeHeight, ok := PendingFinalizationHeightForChannel(channel)
		if !ok {
			continue
		}
		jobID := asyncFinalizationJobID(channel.ChannelID, finalizeHeight)
		if _, found := asyncFinalizationJobByID(next.AsyncFinalizationQueue, jobID); found {
			continue
		}
		next.AsyncFinalizationQueue = append(next.AsyncFinalizationQueue, AsyncFinalizationJob{
			JobID:		jobID,
			ChannelID:	channel.ChannelID,
			FinalizeHeight:	finalizeHeight,
			EnqueuedHeight:	currentHeight,
		}.Normalize())
	}
	sortAsyncFinalizationJobs(next.AsyncFinalizationQueue)
	return next, next.Validate()
}

func EnqueueExpiredPromise(state PaymentsState, promise ConditionalPromise, resolver string, currentHeight uint64) (PaymentsState, AsyncPromiseExpiryJob, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, AsyncPromiseExpiryJob{}, errors.New("payments async promise enqueue height must be positive")
	}
	promise = promise.Normalize()
	channel, found := state.ChannelByID(promise.ChannelID)
	if !found {
		return PaymentsState{}, AsyncPromiseExpiryJob{}, errors.New("payments channel not found")
	}
	if err := promise.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, AsyncPromiseExpiryJob{}, err
	}
	resolver = strings.TrimSpace(resolver)
	if resolver == "" {
		resolver = promise.Source
	}
	if !containsString(channel.Participants, resolver) {
		return PaymentsState{}, AsyncPromiseExpiryJob{}, errors.New("payments async promise resolver must be participant")
	}
	expireAfterHeight := promise.TimeoutHeight + 1
	jobID := asyncPromiseExpiryJobID(channel.ChannelID, promise.PromiseID, expireAfterHeight)
	if existing, found := asyncPromiseExpiryJobByID(state.AsyncPromiseExpiryQueue, jobID); found {
		return state, existing, nil
	}
	job := AsyncPromiseExpiryJob{
		JobID:			jobID,
		ChannelID:		channel.ChannelID,
		PromiseID:		promise.PromiseID,
		Promise:		promise,
		Resolver:		resolver,
		ExpireAfterHeight:	expireAfterHeight,
		EnqueuedHeight:		currentHeight,
	}.Normalize()
	if err := job.Validate(); err != nil {
		return PaymentsState{}, AsyncPromiseExpiryJob{}, err
	}
	next := state.Clone()
	next.AsyncPromiseExpiryQueue = append(next.AsyncPromiseExpiryQueue, job)
	sortAsyncPromiseExpiryJobs(next.AsyncPromiseExpiryQueue)
	return next, job, next.Validate()
}

func ProcessAsyncExecutionQueues(state PaymentsState, currentHeight, maxFinalizations, maxPromiseExpiries uint64) (PaymentsState, AsyncExecutionResult, error) {
	if currentHeight == 0 {
		return PaymentsState{}, AsyncExecutionResult{}, errors.New("payments async process height must be positive")
	}
	next, err := RefreshAsyncExecutionQueues(state, currentHeight)
	if err != nil {
		return PaymentsState{}, AsyncExecutionResult{}, err
	}
	result := AsyncExecutionResult{}
	for _, queued := range next.AsyncFinalizationQueue {
		if maxFinalizations > 0 && result.ProcessedFinalizations >= maxFinalizations {
			break
		}
		job := queued.Normalize()
		if job.Completed || job.FinalizeHeight > currentHeight {
			continue
		}
		result.ProcessedFinalizations++
		channel, found := next.ChannelByID(job.ChannelID)
		if !found {
			next = markAsyncFinalizationFailed(next, job.JobID, currentHeight, "payments channel not found")
			result.FailedJobIDs = append(result.FailedJobIDs, job.JobID)
			continue
		}
		if channel.Status == ChannelStatusSettled {
			settlementHash := latestSettlementHashForChannel(next.Settlements, channel.ChannelID)
			if settlementHash == "" {
				settlementHash = channel.OpeningStateHash
			}
			next = markAsyncFinalizationCompleted(next, job.JobID, settlementHash, currentHeight)
			next = appendAsyncCompletion(next, job.JobID, "finalization", channel.ChannelID, channel.ChannelID, settlementHash, currentHeight, &result)
			result.CompletedJobIDs = append(result.CompletedJobIDs, job.JobID)
			continue
		}
		var settlement SettlementRecord
		next, settlement, err = FinalizeSettlement(next, job.ChannelID, currentHeight)
		if err != nil {
			next = markAsyncFinalizationFailed(next, job.JobID, currentHeight, err.Error())
			result.FailedJobIDs = append(result.FailedJobIDs, job.JobID)
			continue
		}
		next = markAsyncFinalizationCompleted(next, job.JobID, settlement.SettlementHash, currentHeight)
		next = appendAsyncCompletion(next, job.JobID, "finalization", settlement.ChannelID, settlement.ChannelID, settlement.SettlementHash, currentHeight, &result)
		result.CompletedJobIDs = append(result.CompletedJobIDs, job.JobID)
	}
	for _, queued := range next.AsyncPromiseExpiryQueue {
		if maxPromiseExpiries > 0 && result.ProcessedPromiseExpiries >= maxPromiseExpiries {
			break
		}
		job := queued.Normalize()
		if job.Completed || job.ExpireAfterHeight > currentHeight {
			continue
		}
		result.ProcessedPromiseExpiries++
		channel, found := next.ChannelByID(job.ChannelID)
		if !found {
			next = markAsyncPromiseExpiryFailed(next, job.JobID, currentHeight, "payments channel not found")
			result.FailedJobIDs = append(result.FailedJobIDs, job.JobID)
			continue
		}
		if promiseWasSettled(channel, job.PromiseID, next.ConditionClaims) {
			resultHash := HashParts("async-promise-already-settled", job.ChannelID, job.PromiseID)
			next = markAsyncPromiseExpiryCompleted(next, job.JobID, resultHash, currentHeight)
			next = appendAsyncCompletion(next, job.JobID, "promise-expiry", job.ChannelID, job.PromiseID, resultHash, currentHeight, &result)
			result.CompletedJobIDs = append(result.CompletedJobIDs, job.JobID)
			continue
		}
		var resolutions []ConditionResolution
		next, resolutions, _, err = ExpireConditionalPromises(next, PromiseExpiryRequest{
			ChannelID:	job.ChannelID,
			Promises:	[]ConditionalPromise{job.Promise},
			Resolver:	job.Resolver,
			CurrentHeight:	currentHeight,
		})
		if err != nil {
			next = markAsyncPromiseExpiryFailed(next, job.JobID, currentHeight, err.Error())
			result.FailedJobIDs = append(result.FailedJobIDs, job.JobID)
			continue
		}
		resultHash := HashParts("async-promise-expiry", job.ChannelID, job.PromiseID, resolutions[0].EvidenceHash)
		next = markAsyncPromiseExpiryCompleted(next, job.JobID, resultHash, currentHeight)
		next = appendAsyncCompletion(next, job.JobID, "promise-expiry", job.ChannelID, job.PromiseID, resultHash, currentHeight, &result)
		result.CompletedJobIDs = append(result.CompletedJobIDs, job.JobID)
	}
	sortStrings(result.CompletedJobIDs)
	sortStrings(result.FailedJobIDs)
	sortStrings(result.EmittedCompletionIDs)
	return next, result, next.Validate()
}

func OpenChannelFromRequest(state PaymentsState, req ChannelOpenRequest) (PaymentsState, PaymentEvent, error) {
	channel, err := BuildChannelFromOpenRequest(req)
	if err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	next, event, err := openChannelRecord(state, channel)
	return next, event, err
}

func OpenChannel(state PaymentsState, channel ChannelRecord) (PaymentsState, error) {
	next, _, err := openChannelRecord(state, channel)
	return next, err
}

func openChannelRecord(state PaymentsState, channel ChannelRecord) (PaymentsState, PaymentEvent, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	channel = channel.Normalize()
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments new channel must start open")
	}
	if _, found := state.ChannelByID(channel.ChannelID); found {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments channel already exists")
	}
	if err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationOpen,
		ChannelID:	channel.ChannelID,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if err := channel.LatestState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if channel.OpeningStateHash == "" {
		channel.OpeningStateHash = channel.LatestState.StateHash
	}
	if channel.OpeningStateHash != channel.LatestState.StateHash {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments opening state hash mismatch")
	}
	channel.FinalizedNonce = 0
	channel.Finality = ChannelFinalityOpen
	if err := channel.Validate(); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	lock := CustodyLock{ChannelID: channel.ChannelID, Denom: NativeDenom, Amount: channel.Collateral}.Normalize()
	if err := lock.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if _, found := state.CustodyLockByChannel(channel.ChannelID); found {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments custody lock already exists")
	}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassChannelOpen, channel, channel.Participants[0], channel.ChannelID, channel.OpeningFeePaid, channel.OpenHeight)
	if err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	event := ChannelOpenEvent(channel)
	next := chargedState.Clone()
	next.Channels = append(next.Channels, channel)
	next.CustodyLocks = append(next.CustodyLocks, lock)
	next.Events = append(next.Events, event)
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(channel, "", ChannelFinalityOpen, channel.OpenHeight))
	sortChannels(next.Channels)
	sortCustodyLocks(next.CustodyLocks)
	return next, event, next.Validate()
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
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassRoutingAdvertisement, channel, edge.From, edgeKey(edge), edge.AdvertisementFeePaid, channel.OpenHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	next.Edges = append(next.Edges, edge)
	sortEdges(next.Edges)
	return next, next.Validate()
}

func ApplyGossipEnvelope(store TopologyStore, state PaymentsState, envelope SignedGossipEnvelope, currentHeight uint64) (TopologyStore, error) {
	store = store.Normalize()
	state = state.Export()
	envelope = envelope.Normalize()
	if err := envelope.ValidateForState(state, currentHeight); err != nil {
		return PenalizeInvalidGossip(store, gossipPenaltyNode(envelope), currentHeight), err
	}
	message, err := BuildGossipMessage(envelope.Message)
	if err != nil {
		return PenalizeInvalidGossip(store, gossipPenaltyNode(envelope), currentHeight), err
	}
	envelope.Message = message
	envelope.MessageHash = message.MessageID
	if envelope.ReceivedAt == 0 {
		envelope.ReceivedAt = currentHeight
	}
	next := store
	next.Messages = upsertGossipEnvelope(next.Messages, envelope)
	if edge, ok := message.ToChannelEdge(); ok {
		next.Edges = upsertTopologyEdge(next.Edges, edge)
	}
	next.Reputation = addGossipReputation(next.Reputation, message.NodeID, message.ReputationDelta, false, currentHeight)
	return next.Normalize(), next.Validate()
}

func PenalizeInvalidGossip(store TopologyStore, nodeID string, currentHeight uint64) TopologyStore {
	store = store.Normalize()
	nodeID = strings.TrimSpace(nodeID)
	if currentHeight == 0 || nodeID == "" {
		return store
	}
	store.Reputation = addGossipReputation(store.Reputation, nodeID, -InvalidGossipPenalty, true, currentHeight)
	return store.Normalize()
}

func PruneTopologyStore(store TopologyStore, currentHeight uint64) (TopologyStore, error) {
	if currentHeight == 0 {
		return TopologyStore{}, errors.New("payments topology prune height must be positive")
	}
	store = store.Normalize()
	next := TopologyStore{
		Messages:	make([]SignedGossipEnvelope, 0, len(store.Messages)),
		Edges:		make([]ChannelEdge, 0, len(store.Edges)),
		Reputation:	append([]GossipReputation(nil), store.Reputation...),
		LastPrunedAt:	currentHeight,
	}
	for _, envelope := range store.Messages {
		envelope = envelope.Normalize()
		if envelope.Message.ValidUntilHeight >= currentHeight {
			next.Messages = append(next.Messages, envelope)
		}
	}
	for _, edge := range store.Edges {
		edge = edge.Normalize()
		if edge.ExpiresHeight == 0 || edge.ExpiresHeight >= currentHeight {
			next.Edges = append(next.Edges, edge)
		}
	}
	next = next.Normalize()
	return next, next.Validate()
}

func RoutingScoreForEdge(store TopologyStore, edge ChannelEdge) int64 {
	store = store.Normalize()
	edge = edge.Normalize()
	score := int64(0)
	for _, reputation := range store.Reputation {
		reputation = reputation.Normalize()
		if reputation.NodeID == edge.From {
			score += reputation.Score
		}
	}
	return score
}

func AcceptSignedState(state PaymentsState, channelID string, nextState ChannelState, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments state update height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	nextState = nextState.Normalize()
	if err := nextState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, err
	}
	if nextState.Nonce <= channel.LatestState.Nonce {
		return PaymentsState{}, errors.New("payments channel state nonce must strictly increase")
	}
	if err := ValidatePreviousHashContinuity(channel, nextState); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.LatestState = nextState
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func AcceptAsyncCheckpoint(state PaymentsState, channelID string, checkpoint ChannelState, deltas []AsyncPaymentDelta, submitter string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments async checkpoint height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeAsync {
		return PaymentsState{}, errors.New("payments checkpoint requires async channel")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments async checkpoint submitter must be participant")
	}
	checkpoint = checkpoint.Normalize()
	if err := checkpoint.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if checkpoint.CheckpointNonce <= channel.LatestState.CheckpointNonce {
		return PaymentsState{}, errors.New("payments async checkpoint nonce must increase")
	}
	proof := AsyncDeltaDisputeProof{
		ProofID:		HashParts("async-checkpoint-proof", checkpoint.StateHash),
		ChannelID:		channel.ChannelID,
		CheckpointState:	checkpoint,
		Deltas:			deltas,
		EvidenceHash:		HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, deltas)),
	}
	if err := proof.ValidateForChannel(channel, currentHeight); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.LatestState = checkpoint
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func RegisterUpdateCheckpoint(state PaymentsState, req ChannelUpdateRequest) (PaymentsState, ChannelUpdateResult, error) {
	state = state.Export()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, ChannelUpdateResult{}, errors.New("payments channel not found")
	}
	result, err := ValidateOffchainUpdate(channel, req)
	if err != nil {
		return PaymentsState{}, ChannelUpdateResult{}, err
	}
	if !req.Normalize().RegisterCheckpoint {
		return state, result, nil
	}
	var next PaymentsState
	if channel.ChannelType == ChannelTypeAsync || len(req.Normalize().AsyncDeltas) > 0 {
		next, err = AcceptAsyncCheckpoint(state, channel.ChannelID, req.Normalize().State, req.Normalize().AsyncDeltas, req.Normalize().Submitter, req.Normalize().CurrentHeight)
	} else {
		next, err = AcceptSignedState(state, channel.ChannelID, req.Normalize().State, req.Normalize().CurrentHeight)
	}
	if err != nil {
		return PaymentsState{}, ChannelUpdateResult{}, err
	}
	next, _, err = ChargePaymentFee(next, PaymentFeeClassChannelCheckpoint, channel, req.Normalize().Submitter, req.Normalize().State.StateHash, req.Normalize().CheckpointFeePaid, req.Normalize().CurrentHeight)
	if err != nil {
		return PaymentsState{}, ChannelUpdateResult{}, err
	}
	result.CheckpointRegistered = true
	return next, result, nil
}

func RevealPromisePreimage(state PaymentsState, req PreimageRevealRequest) (PaymentsState, []ConditionResolution, error) {
	state = state.Export()
	req = req.Normalize()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, nil, errors.New("payments channel not found")
	}
	if err := req.ValidateForChannel(channel, state.ConditionClaims); err != nil {
		return PaymentsState{}, nil, err
	}
	preimageHash := HashParts(req.Preimage)
	resolutions := make([]ConditionResolution, 0, len(req.Promises))
	next := state.Clone()
	for _, promise := range normalizeConditionalPromises(req.Promises) {
		evidenceHash := HashParts("promise-preimage", promise.PromiseID, preimageHash)
		resolution := ConditionResolution{
			ConditionID:	promise.PromiseID,
			Resolver:	req.Revealer,
			Recipient:	promise.Destination,
			Amount:		promise.Amount,
			Expired:	false,
			EvidenceHash:	evidenceHash,
		}.Normalize()
		resolutions = append(resolutions, resolution)
		next.ConditionClaims = append(next.ConditionClaims, ConditionClaimRecord{
			ChainID:	channel.ChainID,
			ChannelID:	channel.ChannelID,
			ConditionID:	promise.PromiseID,
			EvidenceHash:	evidenceHash,
			PreimageHash:	preimageHash,
			ResolvedHeight:	req.CurrentHeight,
			ExpiresHeight:	req.CurrentHeight + DefaultReplayHorizon,
		}.Normalize())
	}
	sortConditionClaimRecords(next.ConditionClaims)
	return next, normalizeConditionResolutions(resolutions), next.Validate()
}

func ExpireConditionalPromises(state PaymentsState, req PromiseExpiryRequest) (PaymentsState, []ConditionResolution, ConditionRootUpdate, error) {
	state = state.Export()
	req = req.Normalize()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, nil, ConditionRootUpdate{}, errors.New("payments channel not found")
	}
	if err := req.ValidateForChannel(channel, state.ConditionClaims); err != nil {
		return PaymentsState{}, nil, ConditionRootUpdate{}, err
	}
	_, update, err := BuildConditionRootAfterExpiry(channel.LatestState, req.Promises)
	if err != nil {
		return PaymentsState{}, nil, ConditionRootUpdate{}, err
	}
	resolutions := make([]ConditionResolution, 0, len(req.Promises))
	next := state.Clone()
	for _, promise := range normalizeConditionalPromises(req.Promises) {
		evidenceHash := HashParts("promise-expiry", promise.PromiseID, fmt.Sprintf("%020d", req.CurrentHeight))
		resolution := ConditionResolution{
			ConditionID:	promise.PromiseID,
			Resolver:	req.Resolver,
			Recipient:	promise.Source,
			Amount:		promise.Amount,
			Expired:	true,
			EvidenceHash:	evidenceHash,
		}.Normalize()
		resolutions = append(resolutions, resolution)
		next.ConditionClaims = append(next.ConditionClaims, ConditionClaimRecord{
			ChainID:	channel.ChainID,
			ChannelID:	channel.ChannelID,
			ConditionID:	promise.PromiseID,
			EvidenceHash:	evidenceHash,
			ResolvedHeight:	req.CurrentHeight,
			ExpiresHeight:	req.CurrentHeight + DefaultReplayHorizon,
		}.Normalize())
	}
	sortConditionClaimRecords(next.ConditionClaims)
	return next, normalizeConditionResolutions(resolutions), update, next.Validate()
}

func BatchSettleLinkedPromises(state PaymentsState, req BatchConditionSettlementRequest) (PaymentsState, BatchConditionSettlementResult, error) {
	state = state.Export()
	req = req.Normalize()
	if err := req.ValidateForState(state, state.ConditionClaims); err != nil {
		return PaymentsState{}, BatchConditionSettlementResult{}, err
	}
	proof := req.LinkageProof.Normalize()
	evidenceHash := proof.EvidenceHash
	if evidenceHash == "" {
		evidenceHash = HashParts("batch-condition-settlement", proof.RouteID, string(req.Mode), fmt.Sprintf("%020d", req.CurrentHeight))
	}
	preimageHash := ""
	if req.Mode == ConditionSettlementModePreimage {
		preimageHash = HashParts(req.Preimage)
	}
	updates, err := conditionRootUpdatesForPromises(state, proof.Promises)
	if err != nil {
		return PaymentsState{}, BatchConditionSettlementResult{}, err
	}
	feeChannel, found := state.ChannelByID(proof.Promises[0].ChannelID)
	if !found {
		return PaymentsState{}, BatchConditionSettlementResult{}, errors.New("payments condition fee channel not found")
	}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassConditionalPromiseSettlement, feeChannel, req.Resolver, evidenceHash, req.SettlementFeePaid, req.CurrentHeight)
	if err != nil {
		return PaymentsState{}, BatchConditionSettlementResult{}, err
	}
	next := chargedState.Clone()
	resolutions := make([]ConditionResolution, 0, len(proof.Promises))
	feeClaims := make([]RouteFeeClaim, 0, len(proof.Promises)-1)
	for i, promise := range proof.Promises {
		channel, _ := state.ChannelByID(promise.ChannelID)
		resolution := ConditionResolution{
			ConditionID:	promise.PromiseID,
			Resolver:	req.Resolver,
			Recipient:	promise.Destination,
			Amount:		promise.Amount,
			Expired:	false,
			EvidenceHash:	HashParts("batch-condition-resolution", evidenceHash, promise.PromiseID),
		}
		if req.Mode == ConditionSettlementModeExpiry {
			resolution.Recipient = promise.Source
			resolution.Expired = true
		}
		resolution = resolution.Normalize()
		resolutions = append(resolutions, resolution)
		next.ConditionClaims = append(next.ConditionClaims, ConditionClaimRecord{
			ChainID:	channel.ChainID,
			ChannelID:	channel.ChannelID,
			ConditionID:	promise.PromiseID,
			EvidenceHash:	resolution.EvidenceHash,
			PreimageHash:	preimageHash,
			ResolvedHeight:	req.CurrentHeight,
			ExpiresHeight:	req.CurrentHeight + DefaultReplayHorizon,
		}.Normalize())
		if req.Mode != ConditionSettlementModePreimage || i == 0 {
			continue
		}
		fee, err := parseNonNegativeInt("payments route fee claim amount", promise.Fee)
		if err != nil {
			return PaymentsState{}, BatchConditionSettlementResult{}, err
		}
		if fee.IsZero() {
			continue
		}
		feeClaims = append(feeClaims, RouteFeeClaim{
			ChannelID:	promise.ChannelID,
			PromiseID:	promise.PromiseID,
			Recipient:	promise.Source,
			Amount:		promise.Fee,
			EvidenceHash:	HashParts("route-fee-claim", evidenceHash, promise.PromiseID, promise.Source),
		}.Normalize())
	}
	sortConditionClaimRecords(next.ConditionClaims)
	result := BatchConditionSettlementResult{
		RouteID:		proof.RouteID,
		Resolutions:		resolutions,
		FeeClaims:		feeClaims,
		ConditionRootUpdates:	updates,
		EvidenceHash:		evidenceHash,
	}.Normalize()
	if err := result.Validate(); err != nil {
		return PaymentsState{}, BatchConditionSettlementResult{}, err
	}
	return next, result, next.Validate()
}

func SubmitClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	return SubmitCloseWithRequest(state, ChannelCloseRequest{
		ChannelID:	channelID,
		ClosingState:	closingState,
		CloseReason:	CloseReasonUnilateral,
		Submitter:	submitter,
		CurrentHeight:	currentHeight,
		SettlementFee:	settlementFee,
	})
}

func SubmitCloseWithRequest(state PaymentsState, req ChannelCloseRequest) (PaymentsState, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, errors.New("payments close height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	closingState := req.ClosingStateWithSignatures()
	if err := req.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	pending := PendingClose{
		Submitter:		req.Submitter,
		SubmittedHeight:	req.CurrentHeight,
		SettleAfterHeight:	req.CurrentHeight + channel.DisputePeriod,
		CloseReason:		req.CloseReason,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		req.SettlementFee,
		State:			closingState,
	}
	if err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationUnilateralClose,
		ChannelID:	channel.ChannelID,
		SignedState:	pending.State,
		CurrentHeight:	req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := pending.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if pending.State.Nonce < channel.FinalizedNonce {
		return PaymentsState{}, errors.New("payments close state nonce is below finalized nonce")
	}
	if pending.State.Nonce < channel.LatestState.Nonce {
		return PaymentsState{}, errors.New("payments close state nonce is below latest accepted nonce")
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusPendingClose
	nextChannel.PendingClose = pending
	nextChannel.LatestState = pending.State
	if nextChannel.DisputedNonce < pending.State.Nonce {
		nextChannel.DisputedNonce = pending.State.Nonce
	}
	feeClass := PaymentFeeClassUnilateralClose
	if req.CloseReason == CloseReasonCooperative {
		feeClass = PaymentFeeClassCooperativeClose
	}
	chargedState, _, err := ChargePaymentFee(state, feeClass, channel, req.Submitter, pending.State.StateHash, req.SettlementFee, req.CurrentHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func ForcedClose(state PaymentsState, channelID string, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments forced close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments forced close submitter must be participant")
	}
	timeoutHeight := channel.LatestState.TimeoutHeight
	if channel.ChannelType == ChannelTypeAsync && channel.LatestState.ExpiryHeight != 0 {
		timeoutHeight = channel.LatestState.ExpiryHeight
	}
	if channel.ChannelType == ChannelTypeUnidirectional && channel.ExpirationHeight != 0 {
		timeoutHeight = channel.ExpirationHeight
	}
	if timeoutHeight == 0 || currentHeight <= timeoutHeight {
		return PaymentsState{}, errors.New("payments forced close timeout has not expired")
	}
	pending := PendingClose{
		Submitter:		submitter,
		SubmittedHeight:	currentHeight,
		SettleAfterHeight:	currentHeight + channel.DisputePeriod,
		CloseReason:		CloseReasonTimeout,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		settlementFee,
		State:			channel.LatestState.Normalize(),
	}
	if err := pending.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusPendingClose
	nextChannel.PendingClose = pending
	if nextChannel.DisputedNonce < pending.State.Nonce {
		nextChannel.DisputedNonce = pending.State.Nonce
	}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassUnilateralClose, channel, submitter, pending.State.StateHash, settlementFee, currentHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func CooperativeClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close submitter must be participant")
	}
	closingState = closingState.Normalize()
	if err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationCooperativeClose,
		ChannelID:	channel.ChannelID,
		SignedState:	closingState,
		CurrentHeight:	currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := closingState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if closingState.Nonce < channel.LatestState.Nonce {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close state nonce is below latest accepted nonce")
	}
	finalBalances, err := applySettlementAdjustments(closingState.Balances, nil, nil, settlementFee, submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		closingState.StateHash,
		Nonce:			closingState.Nonce,
		FinalBalances:		finalBalances,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		settlementFee,
		SettledHeight:		currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.LatestState = closingState
	nextChannel.PendingClose = PendingClose{}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassCooperativeClose, channel, submitter, closingState.StateHash, settlementFee, currentHeight)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	next := chargedState.Clone()
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, channel.Finality, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func ReceiverClose(state PaymentsState, channelID string, claim UnidirectionalClaim, receiver string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeUnidirectional {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close requires unidirectional channel")
	}
	receiver = normalizeAddress(receiver)
	if receiver != channel.Receiver {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close submitter must be receiver")
	}
	claim = claim.Normalize()
	if err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationUnilateralClose,
		ChannelID:	channel.ChannelID,
		Claim:		claim,
		CurrentHeight:	currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := claim.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := validateUnidirectionalClaimProgress(channel.LatestClaim, claim); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if currentHeight > claim.ExpirationHeight+channel.DisputePeriod {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close claim has expired")
	}
	finalBalances, err := finalBalancesForUnidirectionalClaim(channel, claim, settlementFee, receiver)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		claim.StateHash,
		Nonce:			claim.Nonce,
		FinalBalances:		finalBalances,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		settlementFee,
		SettledHeight:		currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.LatestClaim = claim
	nextChannel.PendingClose = PendingClose{}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassUnilateralClose, channel, receiver, claim.StateHash, settlementFee, currentHeight)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	next := chargedState.Clone()
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, channel.Finality, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func PayerReclaim(state PaymentsState, channelID string, payer string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments payer reclaim height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeUnidirectional {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments payer reclaim requires unidirectional channel")
	}
	payer = normalizeAddress(payer)
	if payer != channel.Payer {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments reclaim submitter must be payer")
	}
	expirationHeight := channel.ExpirationHeight
	claim := channel.LatestClaim.Normalize()
	if !claim.IsZero() {
		expirationHeight = claim.ExpirationHeight
	}
	if currentHeight <= expirationHeight+channel.DisputePeriod {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments reclaim is still in dispute window")
	}
	stateHash := channel.OpeningStateHash
	nonce := channel.LatestState.Nonce
	var finalBalances []Balance
	var err error
	if claim.IsZero() {
		finalBalances, err = applySettlementAdjustments([]Balance{
			{Participant: channel.Payer, Amount: channel.Collateral},
			{Participant: channel.Receiver, Amount: "0"},
		}, nil, nil, settlementFee, payer)
	} else {
		stateHash = claim.StateHash
		nonce = claim.Nonce
		finalBalances, err = finalBalancesForUnidirectionalClaim(channel, claim, settlementFee, payer)
	}
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		stateHash,
		Nonce:			nonce,
		FinalBalances:		finalBalances,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		settlementFee,
		SettledHeight:		currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassUnilateralClose, channel, payer, stateHash, settlementFee, currentHeight)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	next := chargedState.Clone()
	var transitionErr error
	nextChannel, transitionErr = setChannelFinality(nextChannel, ChannelFinalityExpired, currentHeight, &next.Events)
	if transitionErr != nil {
		return PaymentsState{}, SettlementRecord{}, transitionErr
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityExpired, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func DisputeClose(state PaymentsState, channelID string, newerState ChannelState, submitter string, currentHeight uint64) (PaymentsState, error) {
	channel, found := state.Export().ChannelByID(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	return DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:		channelID,
		ClosingStateReference:	channel.PendingClose.State.StateHash,
		NewerState:		newerState,
		Submitter:		submitter,
		CurrentHeight:		currentHeight,
	})
}

func DisputeChannel(state PaymentsState, req ChannelDisputeRequest) (PaymentsState, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, errors.New("payments dispute height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments channel is not pending close")
	}
	if req.ClosingStateReference != channel.PendingClose.State.StateHash {
		return PaymentsState{}, errors.New("payments dispute closing state reference mismatch")
	}
	if req.CurrentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments dispute window has closed")
	}
	if err := (SettlementArbitrationInput{
		Operation:		SettlementArbitrationDispute,
		ChannelID:		channel.ChannelID,
		SignedState:		req.NewerState,
		ConditionProofs:	req.ConditionProofs,
		CurrentHeight:		req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := req.NewerState.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if !stateStrongerThan(req.NewerState, channel.PendingClose.State) {
		return PaymentsState{}, errors.New("payments dispute state must be newer or stronger")
	}
	if !containsString(channel.Participants, req.Submitter) {
		return PaymentsState{}, errors.New("payments dispute submitter must be participant")
	}
	if err := rejectReusedConditionClaims(state, channel, req.ConditionProofs); err != nil {
		return PaymentsState{}, err
	}
	if err := validateConditionResolutionsForState(req.NewerState, channel, req.ConditionProofs, false); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.State = req.NewerState
	nextChannel.PendingClose.SubmittedHeight = req.CurrentHeight
	if nextChannel.PendingClose.DisputeCount < MaxDisputeExtensions {
		nextChannel.PendingClose.SettleAfterHeight = req.CurrentHeight + channel.DisputePeriod
		nextChannel.PendingClose.DisputeCount++
	}
	nextChannel.PendingClose.ConditionProofs = mergeConditionResolutions(nextChannel.PendingClose.ConditionProofs, req.ConditionProofs)
	if nextChannel.DisputedNonce < req.NewerState.Nonce {
		nextChannel.DisputedNonce = req.NewerState.Nonce
	}
	if req.FraudProof.ProofID != "" {
		if err := req.FraudProof.ValidateForChannel(channel); err != nil {
			return PaymentsState{}, err
		}
		penalties, allocations, err := BuildFraudPenaltyRouting(channel, req.FraudProof, FraudPenaltyPolicy{})
		if err != nil {
			return PaymentsState{}, err
		}
		nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, req.FraudProof)
		nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalties...)
		nextChannel.PendingClose.PenaltyAllocations = append(nextChannel.PendingClose.PenaltyAllocations, allocations...)
	}
	nextChannel.LatestState = req.NewerState
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassDispute, channel, req.Submitter, req.NewerState.StateHash, req.DisputeFeePaid, req.CurrentHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Events = append(next.Events, ChannelDisputeEvent(nextChannel, req.Submitter, req.CurrentHeight))
	sortChannels(next.Channels)
	return next, next.Validate()
}

func SubmitWatchDispute(state PaymentsState, submission WatchDisputeSubmission) (PaymentsState, error) {
	state = state.Export()
	submission = submission.Normalize()
	channel, found := state.ChannelByID(submission.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if err := submission.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	return DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:		submission.ChannelID,
		ClosingStateReference:	submission.ClosingStateReference,
		NewerState:		submission.NewerState,
		Submitter:		submission.Delegator,
		CurrentHeight:		submission.CurrentHeight,
	})
}

func RegisterValidatorPaymentService(state PaymentsState, metadata ValidatorPaymentServiceMetadata) (PaymentsState, error) {
	state = state.Export()
	metadata = metadata.Normalize()
	metadata.MetadataHash = ComputeValidatorPaymentServiceMetadataHash(metadata)
	if err := metadata.Validate(); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.ValidatorPaymentServices {
		if existing.Normalize().ValidatorAddress == metadata.ValidatorAddress {
			next.ValidatorPaymentServices[i] = metadata
			replaced = true
			break
		}
	}
	if !replaced {
		next.ValidatorPaymentServices = append(next.ValidatorPaymentServices, metadata)
	}
	sortValidatorPaymentServices(next.ValidatorPaymentServices)
	return next, next.Validate()
}

func RegisterValidatorWatchService(state PaymentsState, registration ValidatorWatchRegistration) (PaymentsState, error) {
	state = state.Export()
	registration = registration.Normalize()
	metadata, found := state.ValidatorPaymentServiceByValidator(registration.ValidatorAddress)
	if !found {
		return PaymentsState{}, errors.New("payments validator service not found")
	}
	registration.ServiceAddress = metadata.ServiceAddress
	registration.MetadataHash = metadata.MetadataHash
	if registration.MinDelegation == "" || registration.MinDelegation == "0" {
		registration.MinDelegation = metadata.MinDelegation
	}
	if err := registration.Validate(metadata); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.ValidatorWatchRegistries {
		existing = existing.Normalize()
		if existing.ValidatorAddress == registration.ValidatorAddress && existing.Delegator == registration.Delegator {
			next.ValidatorWatchRegistries[i] = registration
			replaced = true
			break
		}
	}
	if !replaced {
		next.ValidatorWatchRegistries = append(next.ValidatorWatchRegistries, registration)
	}
	sortValidatorWatchRegistrations(next.ValidatorWatchRegistries)
	return next, next.Validate()
}

func SubmitValidatorAssistedDispute(state PaymentsState, submission ValidatorAssistedDisputeSubmission) (PaymentsState, error) {
	state = state.Export()
	submission = submission.Normalize()
	metadata, found := state.ValidatorPaymentServiceByValidator(submission.ValidatorAddress)
	if !found {
		return PaymentsState{}, errors.New("payments validator service not found")
	}
	channel, found := state.ChannelByID(submission.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if err := submission.ValidateForChannel(channel, metadata); err != nil {
		return PaymentsState{}, err
	}
	if _, found := state.ValidatorWatchRegistration(submission.ValidatorAddress, submission.Delegator); !found {
		return PaymentsState{}, errors.New("payments validator watch registration not found")
	}
	next, err := SubmitWatchDispute(state, WatchDisputeSubmission{
		WatchService:		metadata.ServiceAddress,
		Delegator:		submission.Delegator,
		ChannelID:		submission.ChannelID,
		ClosingStateReference:	submission.ClosingStateReference,
		NewerState:		submission.NewerState,
		CurrentHeight:		submission.CurrentHeight,
		EvidenceHash:		submission.EvidenceHash,
	})
	if err != nil {
		return PaymentsState{}, err
	}
	next.Events = append(next.Events, ValidatorAssistedDisputeEvent(metadata, channel, submission.Delegator, submission.CurrentHeight))
	return next, next.Validate()
}

func SubmitFraudProof(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64) (PaymentsState, error) {
	return SubmitFraudProofWithPolicy(state, channelID, proof, currentHeight, FraudPenaltyPolicy{})
}

func SubmitFraudProofWithPolicy(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64, policy FraudPenaltyPolicy) (PaymentsState, error) {
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
	if err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationFraudProof,
		ChannelID:	channel.ChannelID,
		FraudProof:	proof,
		CurrentHeight:	currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := proof.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range channel.PendingClose.FraudProofs {
		if existing.ProofID == proof.ProofID {
			return PaymentsState{}, errors.New("payments duplicate fraud proof")
		}
	}
	penalties, allocations, err := BuildFraudPenaltyRouting(channel, proof, policy)
	if err != nil {
		return PaymentsState{}, err
	}
	chargedState, charge, err := ChargePaymentFee(state, PaymentFeeClassFraudProofVerification, channel, proof.SubmittedBy, proof.ProofID, proof.VerificationFeePaid, currentHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	refundedState := chargedState
	if charge.Amount != "0" {
		refundedState, _, err = RefundPaymentFee(chargedState, charge.FeeID, proof.SubmittedBy, "accepted-fraud-proof", currentHeight)
		if err != nil {
			return PaymentsState{}, err
		}
	}
	hookedState, _, err := RecordSecurityReserveAllocationHooks(refundedState, channel.ChannelID, proof, allocations, currentHeight, policy.Normalize().SecurityReserveHook)
	if err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, proof)
	nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalties...)
	nextChannel.PendingClose.PenaltyAllocations = append(nextChannel.PendingClose.PenaltyAllocations, allocations...)
	next := hookedState.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityPenalized, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func FraudClose(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close requires pending close")
	}
	if len(channel.PendingClose.FraudProofs) == 0 || len(channel.PendingClose.Penalties) == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close requires accepted proof")
	}
	if err := (SettlementArbitrationInput{
		Operation:		SettlementArbitrationFinalSettlement,
		ChannelID:		channel.ChannelID,
		SignedState:		channel.PendingClose.State,
		ConditionProofs:	channel.PendingClose.ConditionProofs,
		CurrentHeight:		currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := rejectReusedConditionClaims(state, channel, channel.PendingClose.ConditionProofs); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	finalBalances, err := applySettlementAdjustments(channel.PendingClose.State.Balances, channel.PendingClose.Penalties, channel.PendingClose.PenaltyAllocations, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		channel.PendingClose.State.StateHash,
		Nonce:			channel.PendingClose.State.Nonce,
		FinalBalances:		finalBalances,
		SettlementFeeDenom:	channel.PendingClose.SettlementFeeDenom,
		SettlementFee:		channel.PendingClose.SettlementFee,
		Penalties:		channel.PendingClose.Penalties,
		PenaltyAllocations:	channel.PendingClose.PenaltyAllocations,
		SettledHeight:		currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	next := state.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityFinalizable, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	settledFinality := finalityForSettledChannel(nextChannel)
	nextChannel.Finality = settledFinality
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityFinalizable, settledFinality, currentHeight))
	nextChannel.PendingClose = PendingClose{}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, channel.PendingClose.ConditionProofs, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	sortConditionClaimRecords(next.ConditionClaims)
	return next, settlement, next.Validate()
}

func FinalizeSettlement(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	return FinalizeSettlementWithRequest(state, FinalSettlementRequest{ChannelID: channelID, CurrentHeight: currentHeight})
}

func FinalizeSettlementWithRequest(state PaymentsState, req FinalSettlementRequest) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not pending close")
	}
	if req.CurrentHeight < channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement is still in dispute window")
	}
	resolutions := mergeConditionResolutions(channel.PendingClose.ConditionProofs, req.ResolvedConditions)
	if err := (SettlementArbitrationInput{
		Operation:		SettlementArbitrationFinalSettlement,
		ChannelID:		channel.ChannelID,
		SignedState:		channel.PendingClose.State,
		ConditionProofs:	resolutions,
		CurrentHeight:		req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := rejectReusedConditionClaims(state, channel, resolutions); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := validateConditionResolutionsForState(channel.PendingClose.State, channel, resolutions, true); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	baseBalances, err := settlementBalancesWithConditions(channel.PendingClose.State, channel, resolutions)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	finalBalances, err := applySettlementAdjustments(baseBalances, channel.PendingClose.Penalties, channel.PendingClose.PenaltyAllocations, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		channel.PendingClose.State.StateHash,
		Nonce:			channel.PendingClose.State.Nonce,
		FinalBalances:		finalBalances,
		SettlementFeeDenom:	channel.PendingClose.SettlementFeeDenom,
		SettlementFee:		channel.PendingClose.SettlementFee,
		Penalties:		channel.PendingClose.Penalties,
		PenaltyAllocations:	channel.PendingClose.PenaltyAllocations,
		SettledHeight:		req.CurrentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	next := state.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityFinalizable, req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	settledFinality := finalityForSettledChannel(nextChannel)
	nextChannel.Finality = settledFinality
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityFinalizable, settledFinality, req.CurrentHeight))
	nextChannel.PendingClose = PendingClose{}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, resolutions, req.CurrentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	sortConditionClaimRecords(next.ConditionClaims)
	return next, settlement, next.Validate()
}

func OpenVirtualChannel(state PaymentsState, vc VirtualChannel) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	vc = vc.Normalize()
	if _, found := state.VirtualChannelByID(vc.VirtualChannelID); found {
		return PaymentsState{}, errors.New("payments virtual channel already exists")
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return PaymentsState{}, err
	}
	var parentChainID string
	for _, parentID := range vc.ParentChannelIDs {
		channel, found := state.ChannelByID(parentID)
		if !found || channel.Status != ChannelStatusOpen {
			return PaymentsState{}, errors.New("payments virtual channel requires open parents")
		}
		if parentChainID == "" {
			parentChainID = channel.ChainID
		} else if parentChainID != channel.ChainID {
			return PaymentsState{}, errors.New("payments virtual channel parents must share chain id")
		}
		if !containsString(channel.Participants, vc.Endpoints[0]) && !containsString(channel.Participants, vc.Endpoints[1]) {
			return PaymentsState{}, errors.New("payments virtual channel parent path must touch an endpoint")
		}
		if reserved, err := parentReservedCapacity(channel); err != nil {
			return PaymentsState{}, err
		} else if reserved.LT(capacity) {
			return PaymentsState{}, errors.New("payments virtual channel capacity exceeds parent reserved capacity")
		}
		if err := validateVirtualParentTimeout(vc, channel, 0); err != nil {
			return PaymentsState{}, err
		}
	}
	if vc.ChainID == "" {
		vc.ChainID = parentChainID
		vc.AnchorCommitment = ""
		vc.StateHash = ""
	}
	if vc.ChainID != parentChainID {
		return PaymentsState{}, errors.New("payments virtual channel chain id mismatch")
	}
	if vc.AnchorCommitment == "" || vc.StateHash == "" || vc.ParentRouteID == "" || vc.IntermediarySetHash == "" {
		preservedSignatures := vc.Signatures
		vc.Signatures = nil
		built, err := BuildVirtualChannel(vc)
		if err != nil {
			return PaymentsState{}, err
		}
		vc = built
		vc.Signatures = preservedSignatures
	}
	if err := ValidateVirtualChannelActivation(vc); err != nil {
		return PaymentsState{}, err
	}
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassVirtualChannelAnchor, feeChannelForVirtual(vc), vc.EndpointA, vc.VirtualChannelID, vc.AnchorFeePaid, vc.ExpiresHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	next.VirtualChannels = append(next.VirtualChannels, vc)
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func OpenVirtualChannelWithProof(state PaymentsState, proof VirtualActivationProof) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	proof = proof.Normalize()
	vc := proof.VirtualChannel.Normalize()
	if _, found := state.VirtualChannelByID(vc.VirtualChannelID); found {
		return PaymentsState{}, errors.New("payments virtual channel already exists")
	}
	parentChainID, err := validateVirtualParentAccounting(state, vc, proof.RouteTimeoutHeight, proof.AggregatedCapacity, proof.ParentReserves)
	if err != nil {
		return PaymentsState{}, err
	}
	if vc.ChainID == "" {
		vc.ChainID = parentChainID
		vc.AnchorCommitment = ""
		vc.StateHash = ""
	}
	if vc.ChainID != parentChainID {
		return PaymentsState{}, errors.New("payments virtual channel chain id mismatch")
	}
	if vc.AnchorCommitment == "" || vc.StateHash == "" || vc.ParentRouteID == "" || vc.IntermediarySetHash == "" {
		preservedSignatures := vc.Signatures
		vc.Signatures = nil
		built, err := BuildVirtualChannel(vc)
		if err != nil {
			return PaymentsState{}, err
		}
		vc = built
		vc.Signatures = preservedSignatures
	}
	proof.VirtualChannel = vc.Normalize()
	if proof.ProofHash == "" {
		proof.ProofHash = ComputeVirtualActivationProofHash(proof)
	}
	if err := ValidateVirtualActivationProof(proof); err != nil {
		return PaymentsState{}, err
	}
	for _, reserve := range proof.ParentReserves {
		channel, _ := state.ChannelByID(reserve.ParentChannelID)
		parentReserved, err := parentReservedCapacity(channel)
		if err != nil {
			return PaymentsState{}, err
		}
		reserveCapacity, err := virtualReserveAccountingAmount(reserve, proof.AggregatedCapacity)
		if err != nil {
			return PaymentsState{}, err
		}
		if parentReserved.LT(reserveCapacity) {
			return PaymentsState{}, errors.New("payments virtual reserve exceeds parent reserved capacity")
		}
		if !containsString(channel.Participants, reserve.ReservedBy) {
			return PaymentsState{}, errors.New("payments virtual reserve signer must be parent participant")
		}
	}
	proof.VirtualChannel.ParentReserveCommitments = virtualActivationReserveCommitments(proof)
	chargedState, _, err := ChargePaymentFee(state, PaymentFeeClassVirtualChannelAnchor, feeChannelForVirtual(proof.VirtualChannel), proof.VirtualChannel.EndpointA, proof.VirtualChannel.VirtualChannelID, proof.VirtualChannel.AnchorFeePaid, proof.RouteTimeoutHeight)
	if err != nil {
		return PaymentsState{}, err
	}
	next := chargedState.Clone()
	next.VirtualChannels = append(next.VirtualChannels, proof.VirtualChannel)
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func CloseVirtualChannelWithProof(state PaymentsState, proof VirtualCloseProof, currentHeight uint64) (PaymentsState, VirtualChannel, []VirtualReserveRelease, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, VirtualChannel{}, nil, err
	}
	if currentHeight == 0 {
		return PaymentsState{}, VirtualChannel{}, nil, errors.New("payments virtual close height must be positive")
	}
	index, current, found := state.VirtualChannelIndex(proof.VirtualChannelID)
	if !found {
		return PaymentsState{}, VirtualChannel{}, nil, errors.New("payments virtual channel not found")
	}
	finalState, err := buildVirtualUpdateForCurrent(current, proof.FinalState)
	if err != nil {
		return PaymentsState{}, VirtualChannel{}, nil, err
	}
	proof.FinalState = finalState
	if proof.ProofHash == "" {
		proof.ProofHash = ComputeVirtualCloseProofHash(proof)
	}
	if err := ValidateVirtualCloseProof(proof, current, currentHeight); err != nil {
		return PaymentsState{}, VirtualChannel{}, nil, err
	}
	closed := finalState.Normalize()
	closed.Status = VirtualChannelStatusSettled
	releases, err := virtualReserveReleasesFromClose(proof, current)
	if err != nil {
		return PaymentsState{}, VirtualChannel{}, nil, err
	}
	next := state.Clone()
	next.VirtualChannels = append(next.VirtualChannels[:index], next.VirtualChannels[index+1:]...)
	sortVirtualChannels(next.VirtualChannels)
	return next, closed.Normalize(), releases, next.Validate()
}

func AcceptVirtualChannelUpdate(state PaymentsState, nextVC VirtualChannel, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments virtual update height must be positive")
	}
	index, current, found := state.VirtualChannelIndex(nextVC.VirtualChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments virtual channel not found")
	}
	if current.Status != VirtualChannelStatusOpen {
		return PaymentsState{}, errors.New("payments virtual channel is not open")
	}
	if currentHeight >= current.ExpiresHeight {
		return PaymentsState{}, errors.New("payments virtual channel update expired")
	}
	nextVC, err := buildVirtualUpdateForCurrent(current, nextVC)
	if err != nil {
		return PaymentsState{}, err
	}
	if err := validateVirtualEndpointUpdate(current, nextVC); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.VirtualChannels[index] = nextVC
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func SubmitVirtualChannelDispute(state PaymentsState, proof VirtualChannelDisputeProof, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments virtual dispute height must be positive")
	}
	index, current, found := state.VirtualChannelIndex(proof.VirtualChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments virtual channel not found")
	}
	if currentHeight > current.ExpiresHeight {
		return PaymentsState{}, errors.New("payments virtual dispute expired")
	}
	latest, err := buildVirtualUpdateForCurrent(current, proof.LatestState)
	if err != nil {
		return PaymentsState{}, err
	}
	proof.LatestState = latest
	if proof.EvidenceHash == "" {
		proof.EvidenceHash = ComputeVirtualDisputeEvidenceHash(proof)
	}
	if err := ValidateVirtualChannelDisputeProof(proof, current); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.VirtualChannels[index] = latest
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func CloseVirtualChannel(state PaymentsState, virtualChannelID string, currentHeight uint64) (PaymentsState, VirtualChannel, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, VirtualChannel{}, err
	}
	if currentHeight == 0 {
		return PaymentsState{}, VirtualChannel{}, errors.New("payments virtual close height must be positive")
	}
	virtualChannelID = normalizeHash(virtualChannelID)
	index := -1
	var closed VirtualChannel
	for i, vc := range state.VirtualChannels {
		vc = vc.Normalize()
		if vc.VirtualChannelID == virtualChannelID {
			index = i
			closed = vc
			break
		}
	}
	if index < 0 {
		return PaymentsState{}, VirtualChannel{}, errors.New("payments virtual channel not found")
	}
	closed.Status = VirtualChannelStatusSettled
	next := state.Clone()
	next.VirtualChannels = append(next.VirtualChannels[:index], next.VirtualChannels[index+1:]...)
	sortVirtualChannels(next.VirtualChannels)
	return next, closed.Normalize(), next.Validate()
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
		node	string
		edges	[]ChannelEdge
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

func SelectPaymentRoute(state PaymentsState, store TopologyStore, req RouteSelectionRequest) (ScoredRoute, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return ScoredRoute{}, err
	}
	store = store.Normalize()
	if err := store.Validate(); err != nil {
		return ScoredRoute{}, err
	}
	req = req.Normalize()
	if err := req.Validate(); err != nil {
		return ScoredRoute{}, err
	}
	amount, err := parsePositiveInt("payments scored route amount", req.Amount)
	if err != nil {
		return ScoredRoute{}, err
	}
	route, err := selectPaymentRouteWithPolicy(state, store, req, amount)
	if err != nil {
		return ScoredRoute{}, err
	}
	sim, err := SimulateRoute(route, req)
	if err != nil {
		return ScoredRoute{}, err
	}
	if !sim.Attemptable {
		return ScoredRoute{}, errors.New(sim.Reason)
	}
	return route, nil
}

func ApplyCongestionSnapshot(policy RoutePolicy, snapshot CongestionSnapshot) (RoutePolicy, error) {
	policy = policy.Normalize()
	snapshot = snapshot.Normalize()
	if err := snapshot.Validate(); err != nil {
		return RoutePolicy{}, err
	}
	stats := EdgeRoutingStats{
		ChannelID:		snapshot.ChannelID,
		From:			snapshot.From,
		To:			snapshot.To,
		SuccessRateBps:		10_000 - snapshot.ChannelUpdateFailureRateBps,
		LiquidityUpdatedHeight:	snapshot.LiquidityUpdatedHeight,
		CongestionBps:		snapshot.ChannelUpdateFailureRateBps,
		NodeAvailabilityBps:	10_000,
		FailureCount:		uint32(snapshot.ChannelUpdateFailureRateBps / 1_000),
		PendingConditionCount:	snapshot.PendingConditionCount,
		AvgResolutionLatency:	snapshot.AvgResolutionLatency,
		RetryCount:		snapshot.RouteRetryCount,
		ReservePressureBps:	snapshot.ReservePressureBps,
		NodeQueueDelay:		snapshot.NodeQueueDelay,
		LastUpdatedHeight:	snapshot.ObservedHeight,
	}
	if snapshot.NodeQueueDelay > 0 {
		stats.NodeAvailabilityBps = 10_000 - uint32Min(10_000, uint32(snapshot.NodeQueueDelay))
	}
	policy.EdgeStats = upsertRouteStats(policy.EdgeStats, stats)
	return policy.Normalize(), policy.Validate()
}

func ApplyRouteFailureReport(policy RoutePolicy, report RouteFailureReport) (RoutePolicy, error) {
	policy = policy.Normalize()
	report = report.Normalize()
	if err := report.Validate(); err != nil {
		return RoutePolicy{}, err
	}
	stats, found := routeStatsForEdge(policy, ChannelEdge{ChannelID: report.ChannelID, From: report.From, To: report.To})
	if !found {
		stats = EdgeRoutingStats{
			ChannelID:		report.ChannelID,
			From:			report.From,
			To:			report.To,
			SuccessRateBps:		10_000,
			NodeAvailabilityBps:	10_000,
			LiquidityUpdatedHeight:	report.ObservedHeight,
		}
	}
	stats.FailureCount++
	stats.RetryCount++
	stats.LastFailureHeight = report.ObservedHeight
	stats.LastUpdatedHeight = report.ObservedHeight
	switch report.FailureClass {
	case RouteFailureCapacity:
		stats.ReservePressureBps = uint32Max(stats.ReservePressureBps, 8_000)
	case RouteFailureTimeout:
		stats.TimeoutMargin = 1
	case RouteFailureCongestion:
		stats.CongestionBps = uint32Max(stats.CongestionBps, 8_000)
		stats.PendingConditionCount++
	case RouteFailureLiquidityStale:
		stats.LiquidityUpdatedHeight = 1
	case RouteFailureNodeUnavailable:
		stats.NodeAvailabilityBps = 1_000
	case RouteFailurePolicyRejected:
		stats.CongestionBps = uint32Max(stats.CongestionBps, 5_000)
	default:
		stats.CongestionBps = uint32Max(stats.CongestionBps, 4_000)
	}
	policy.EdgeStats = upsertRouteStats(policy.EdgeStats, stats)
	return policy.Normalize(), policy.Validate()
}

func DecayRoutePolicyPenalties(policy RoutePolicy, currentHeight uint64) RoutePolicy {
	policy = policy.Normalize()
	if currentHeight == 0 {
		return policy
	}
	for i, stats := range policy.EdgeStats {
		policy.EdgeStats[i] = decayEdgeRoutingStats(stats, currentHeight, policy.DecayHalfLife)
	}
	return policy.Normalize()
}

func RetryPaymentRoute(state PaymentsState, store TopologyStore, req RouteRetryRequest) (RouteRetryResult, error) {
	req = req.Normalize()
	if err := req.Validate(); err != nil {
		return RouteRetryResult{}, err
	}
	selection := req.Selection
	selection.Policy = DecayRoutePolicyPenalties(selection.Policy, selection.CurrentHeight)
	for _, failure := range req.Failures {
		var err error
		selection.Policy, err = ApplyRouteFailureReport(selection.Policy, failure)
		if err != nil {
			return RouteRetryResult{}, err
		}
		if req.Policy.ExcludeFailedEdges {
			selection.Policy.ExcludedChannels = append(selection.Policy.ExcludedChannels, failure.ChannelID)
		}
	}
	if uint32(len(req.Failures)) >= req.Policy.MaxAttempts {
		return RouteRetryResult{Attempts: uint32(len(req.Failures)), Retryable: false, Reason: "payments route retry attempts exhausted"}, nil
	}
	route, err := SelectPaymentRoute(state, store, selection)
	if err != nil {
		return RouteRetryResult{Attempts: uint32(len(req.Failures)) + 1, Retryable: false, Reason: err.Error()}, err
	}
	return RouteRetryResult{
		Route:		route,
		Attempts:	uint32(len(req.Failures)) + 1,
		Retryable:	true,
		PolicyHash:	routePolicyHash(selection.Policy),
	}, nil
}

func ClassifyRouteFailure(reason string) RouteFailureClass {
	reason = strings.ToLower(strings.TrimSpace(reason))
	switch {
	case strings.Contains(reason, "capacity") || strings.Contains(reason, "reserve"):
		return RouteFailureCapacity
	case strings.Contains(reason, "timeout") || strings.Contains(reason, "expired"):
		return RouteFailureTimeout
	case strings.Contains(reason, "congestion") || strings.Contains(reason, "queue"):
		return RouteFailureCongestion
	case strings.Contains(reason, "stale") || strings.Contains(reason, "fresh"):
		return RouteFailureLiquidityStale
	case strings.Contains(reason, "availability") || strings.Contains(reason, "unavailable"):
		return RouteFailureNodeUnavailable
	case strings.Contains(reason, "policy") || strings.Contains(reason, "fee"):
		return RouteFailurePolicyRejected
	default:
		return RouteFailureUnknown
	}
}

func CalculateHopRoutingFee(req HopFeeCalculationRequest) (RoutingHopFee, error) {
	policy := req.Policy.Normalize()
	if err := policy.ValidateAtHeight(req.CurrentHeight); err != nil {
		return RoutingHopFee{}, err
	}
	amount, err := parsePositiveInt("payments routing hop fee amount", req.Amount)
	if err != nil {
		return RoutingHopFee{}, err
	}
	base, err := parseNonNegativeInt("payments routing base hop fee", policy.BaseHopFee)
	if err != nil {
		return RoutingHopFee{}, err
	}
	proportional := sdkmath.ZeroInt()
	if policy.ProportionalFeeBps > 0 {
		proportional = amount.Mul(sdkmath.NewInt(int64(policy.ProportionalFeeBps)))
		denom := sdkmath.NewInt(10_000)
		proportional = proportional.Add(denom.Sub(sdkmath.OneInt())).Quo(denom)
	}
	reservation, err := parseNonNegativeInt("payments routing liquidity reservation fee", policy.LiquidityReservationFee)
	if err != nil {
		return RoutingHopFee{}, err
	}
	virtualSetup := sdkmath.ZeroInt()
	if req.IncludeVirtualSetup {
		virtualSetup, err = parseNonNegativeInt("payments routing virtual setup fee", policy.VirtualChannelSetupFee)
		if err != nil {
			return RoutingHopFee{}, err
		}
	}
	congestion, err := parseNonNegativeInt("payments routing congestion surcharge", policy.CongestionSurcharge)
	if err != nil {
		return RoutingHopFee{}, err
	}
	failurePenaltyUnit, err := parseNonNegativeInt("payments routing failure penalty", policy.FailurePenalty)
	if err != nil {
		return RoutingHopFee{}, err
	}
	failurePenalty := failurePenaltyUnit.Mul(sdkmath.NewInt(int64(req.RepeatedInvalidAttempts)))
	total := base.Add(proportional).Add(reservation).Add(virtualSetup).Add(congestion).Add(failurePenalty)
	maxHopFee, err := parseNonNegativeInt("payments routing max hop fee", policy.MaxHopFee)
	if err != nil {
		return RoutingHopFee{}, err
	}
	if !maxHopFee.IsZero() && total.GT(maxHopFee) {
		return RoutingHopFee{}, errors.New("payments routing hop fee exceeds policy maximum")
	}
	return RoutingHopFee{
		Denom:				NativeDenom,
		BaseHopFee:			base.String(),
		ProportionalFee:		proportional.String(),
		LiquidityReservationFee:	reservation.String(),
		VirtualChannelSetupFee:		virtualSetup.String(),
		CongestionSurcharge:		congestion.String(),
		FailurePenalty:			failurePenalty.String(),
		RepeatedInvalidAttempts:	req.RepeatedInvalidAttempts,
		TotalFee:			total.String(),
		PolicyHash:			policy.PolicyHash,
	}, nil
}

func ValidateRouteFeeCeiling(route ScoredRoute, policy RoutePolicy) error {
	route = route.Normalize()
	policy = policy.Normalize()
	if strings.TrimSpace(policy.MaxFeeAmount) == "" {
		return nil
	}
	totalFee, err := parseNonNegativeInt("payments route total fee", route.TotalFee)
	if err != nil {
		return err
	}
	maxFee, err := parseNonNegativeInt("payments route policy max fee", policy.MaxFeeAmount)
	if err != nil {
		return err
	}
	if totalFee.GT(maxFee) {
		return errors.New("payments route fee exceeds policy ceiling")
	}
	return nil
}

func ValidateConditionLinkageFeeCeiling(proof ConditionLinkageProof, policy RoutePolicy) error {
	proof = proof.Normalize()
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return err
	}
	totalFees, err := parseNonNegativeInt("payments linked route declared fees", proof.TotalFees)
	if err != nil {
		return err
	}
	promiseFees := sdkmath.ZeroInt()
	for i := 1; i < len(proof.Promises); i++ {
		fee, err := parseNonNegativeInt("payments linked route promise fee", proof.Promises[i].Fee)
		if err != nil {
			return err
		}
		promiseFees = promiseFees.Add(fee)
	}
	if promiseFees.GT(totalFees) {
		return errors.New("payments linked route promise fee overcharge")
	}
	if strings.TrimSpace(policy.MaxFeeAmount) == "" {
		return nil
	}
	maxFee, err := parseNonNegativeInt("payments linked route policy max fee", policy.MaxFeeAmount)
	if err != nil {
		return err
	}
	if totalFees.GT(maxFee) || promiseFees.GT(maxFee) {
		return errors.New("payments linked route fee exceeds policy ceiling")
	}
	return nil
}

func SimulateRoute(route ScoredRoute, req RouteSelectionRequest) (RouteSimulationResult, error) {
	route = route.Normalize()
	req = req.Normalize()
	if err := route.Validate(); err != nil {
		return RouteSimulationResult{}, err
	}
	if err := req.Validate(); err != nil {
		return RouteSimulationResult{}, err
	}
	amount, err := parsePositiveInt("payments route simulation amount", route.Amount)
	if err != nil {
		return RouteSimulationResult{}, err
	}
	if route.Edges[0].From != req.From {
		return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation source mismatch", TotalFee: route.TotalFee}, nil
	}
	if route.Edges[len(route.Edges)-1].To != req.To {
		return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation destination mismatch", TotalFee: route.TotalFee}, nil
	}
	for i, edge := range route.Edges {
		edge = edge.Normalize()
		if i > 0 && route.Edges[i-1].To != edge.From {
			return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation discontinuity", TotalFee: route.TotalFee}, nil
		}
		capacity, err := parsePositiveInt("payments route simulation capacity", edge.Capacity)
		if err != nil {
			return RouteSimulationResult{}, err
		}
		if !edge.Active || capacity.LT(amount) {
			return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation capacity below amount", TotalFee: route.TotalFee}, nil
		}
		if edge.ExpiresHeight > 0 && req.CurrentHeight > edge.ExpiresHeight {
			return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation edge expired", TotalFee: route.TotalFee}, nil
		}
	}
	if err := ValidateRouteFeeCeiling(route, req.Policy); err != nil {
		if !strings.Contains(err.Error(), "policy ceiling") {
			return RouteSimulationResult{}, err
		}
		return RouteSimulationResult{Route: route, Attemptable: false, Reason: "payments route simulation fee exceeds policy", TotalFee: route.TotalFee}, nil
	}
	return RouteSimulationResult{Route: route, Attemptable: true, TotalFee: route.TotalFee}, nil
}

func SplitPaymentRoute(state PaymentsState, store TopologyStore, req RouteSelectionRequest) (MultiPathRoute, error) {
	req = req.Normalize()
	if err := req.Validate(); err != nil {
		return MultiPathRoute{}, err
	}
	if !req.Policy.EnableMultiPath {
		route, err := SelectPaymentRoute(state, store, req)
		if err != nil {
			return MultiPathRoute{}, err
		}
		return buildMultiPathRoute([]ScoredRoute{route})
	}
	amount, err := parsePositiveInt("payments multipath amount", req.Amount)
	if err != nil {
		return MultiPathRoute{}, err
	}
	maxSplits := req.Policy.Normalize().MaxSplits
	remaining := amount
	parts := make([]ScoredRoute, 0, maxSplits)
	excludedChannels := append([]string(nil), req.Policy.ExcludedChannels...)
	for split := 0; split < maxSplits && remaining.IsPositive(); split++ {
		splitsLeft := maxSplits - split
		chunk := remaining.Quo(sdkmath.NewInt(int64(splitsLeft)))
		if chunk.IsZero() {
			chunk = remaining
		}
		if remaining.Mod(sdkmath.NewInt(int64(splitsLeft))).IsPositive() {
			chunk = chunk.Add(sdkmath.NewInt(1))
		}
		partReq := req
		partReq.Amount = chunk.String()
		partReq.Policy.ExcludedChannels = append([]string(nil), excludedChannels...)
		route, err := SelectPaymentRoute(state, store, partReq)
		if err != nil {
			if len(parts) == 0 {
				return MultiPathRoute{}, err
			}
			break
		}
		parts = append(parts, route)
		remaining = remaining.Sub(chunk)
		for _, edge := range route.Edges {
			excludedChannels = append(excludedChannels, edge.ChannelID)
		}
	}
	if remaining.IsPositive() {
		return MultiPathRoute{}, errors.New("payments multipath route capacity insufficient")
	}
	return buildMultiPathRoute(parts)
}

func BuildForwardingPackets(route ScoredRoute, routeSeed string, routeNonce uint64, timeoutHeight uint64) ([]ForwardingPacket, error) {
	route = route.Normalize()
	if err := route.Validate(); err != nil {
		return nil, err
	}
	if timeoutHeight == 0 {
		return nil, errors.New("payments forwarding timeout height must be positive")
	}
	rootRouteID, err := DeriveRouteID(routeSeed, routeNonce)
	if err != nil {
		return nil, err
	}
	packets := make([]ForwardingPacket, len(route.Edges))
	nextHash := ""
	for i := len(route.Edges) - 1; i >= 0; i-- {
		edge := route.Edges[i].Normalize()
		hopRouteID, err := DeriveHopRouteID(rootRouteID, i, edge.ChannelID)
		if err != nil {
			return nil, err
		}
		hopPaymentID, err := DeriveHopPaymentID(rootRouteID, i, edge.ChannelID)
		if err != nil {
			return nil, err
		}
		packet := ForwardingPacket{
			RouteID:	hopRouteID,
			HopPaymentID:	hopPaymentID,
			ChannelID:	edge.ChannelID,
			ForwardingNode:	edge.From,
			NextNode:	edge.To,
			Amount:		route.Amount,
			FeeAmount:	edge.FeeAmount,
			TimeoutHeight:	timeoutHeight,
			NextPacketHash:	nextHash,
		}.Normalize()
		packet.PacketHash = ComputeForwardingPacketHash(packet)
		packet.PacketID = HashParts("forwarding-packet-id", packet.PacketHash)
		packet = packet.Normalize()
		if err := packet.Validate(); err != nil {
			return nil, err
		}
		packets[i] = packet
		nextHash = packet.PacketHash
	}
	return packets, nil
}

func ValidateForwardingPacket(packet ForwardingPacket, expectedForwarder string, replayRecords []ForwardingPacketReplayRecord, currentHeight uint64) error {
	packet = packet.Normalize()
	if err := packet.Validate(); err != nil {
		return err
	}
	expectedForwarder = strings.TrimSpace(expectedForwarder)
	if expectedForwarder != "" && packet.ForwardingNode != expectedForwarder {
		return errors.New("payments forwarding packet forwarder mismatch")
	}
	if currentHeight == 0 {
		return errors.New("payments forwarding packet validation height must be positive")
	}
	if currentHeight > packet.TimeoutHeight {
		return errors.New("payments forwarding packet is expired")
	}
	for _, record := range normalizeForwardingReplayRecords(replayRecords) {
		if currentHeight > record.ExpiresHeight {
			continue
		}
		if record.PacketID == packet.PacketID {
			return errors.New("payments forwarding packet replay detected")
		}
		if record.RouteID == packet.RouteID {
			return errors.New("payments forwarding route id replay detected")
		}
		if record.HopPaymentID == packet.HopPaymentID {
			return errors.New("payments forwarding payment id replay detected")
		}
	}
	return nil
}

func RecordForwardingPacket(replayRecords []ForwardingPacketReplayRecord, packet ForwardingPacket, currentHeight uint64) ([]ForwardingPacketReplayRecord, error) {
	if err := ValidateForwardingPacket(packet, packet.ForwardingNode, replayRecords, currentHeight); err != nil {
		return nil, err
	}
	records := append(normalizeForwardingReplayRecords(replayRecords), ForwardingPacketReplayRecord{
		PacketID:	packet.PacketID,
		RouteID:	packet.RouteID,
		HopPaymentID:	packet.HopPaymentID,
		RecordedHeight:	currentHeight,
		ExpiresHeight:	currentHeight + DefaultReplayHorizon,
	}.Normalize())
	sortForwardingReplayRecords(records)
	return records, nil
}

func PruneForwardingReplayRecords(replayRecords []ForwardingPacketReplayRecord, currentHeight uint64) []ForwardingPacketReplayRecord {
	out := make([]ForwardingPacketReplayRecord, 0, len(replayRecords))
	for _, record := range normalizeForwardingReplayRecords(replayRecords) {
		if currentHeight <= record.ExpiresHeight {
			out = append(out, record)
		}
	}
	sortForwardingReplayRecords(out)
	return out
}

func PrivacySafeForwardingLog(packet ForwardingPacket, currentHeight uint64) (ForwardingLogRecord, error) {
	packet = packet.Normalize()
	if err := packet.Validate(); err != nil {
		return ForwardingLogRecord{}, err
	}
	if currentHeight == 0 {
		return ForwardingLogRecord{}, errors.New("payments forwarding log height must be positive")
	}
	record := ForwardingLogRecord{
		PacketID:	packet.PacketID,
		RouteID:	packet.RouteID,
		HopPaymentID:	packet.HopPaymentID,
		ChannelID:	packet.ChannelID,
		ForwardingNode:	packet.ForwardingNode,
		NextNodeHash:	HashParts("forwarding-next-node", packet.NextNode),
		AmountHash:	HashParts("forwarding-amount", packet.Amount, packet.FeeAmount),
		RecordedHeight:	currentHeight,
	}.Normalize()
	return record, record.Validate()
}

func ImportState(state PaymentsState) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	return state, nil
}

func BuildStoreV2Layout(state PaymentsState) (StoreV2Layout, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return StoreV2Layout{}, err
	}
	layout := StoreV2Layout{Version: StoreV2MigrationVersion}
	seenConditionKeys := map[string]struct{}{}
	appendCondition := func(record StoreV2ConditionRecord) {
		record = record.Normalize()
		if _, found := seenConditionKeys[record.Key]; found {
			return
		}
		seenConditionKeys[record.Key] = struct{}{}
		layout.Conditions = append(layout.Conditions, record)
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		compact := compactStoreV2Channel(channel)
		participantKeys := make([]string, 0, len(channel.Participants))
		for _, participant := range channel.Participants {
			index := StoreV2ParticipantChannelRecord{
				Key:		StoreV2ParticipantChannelKey(participant, channel.ChannelID),
				Version:	StoreV2MigrationVersion,
				Participant:	participant,
				ChannelID:	channel.ChannelID,
			}.Normalize()
			layout.ParticipantChannels = append(layout.ParticipantChannels, index)
			participantKeys = append(participantKeys, index.Key)
		}
		sortStrings(participantKeys)
		record := StoreV2ChannelRecord{
			Key:				StoreV2ChannelKey(channel.ChannelID),
			Version:			StoreV2MigrationVersion,
			ChannelID:			channel.ChannelID,
			Channel:			compact,
			LatestStateHash:		channel.LatestState.StateHash,
			LatestStateNonce:		channel.LatestState.Nonce,
			ParticipantIndexKeys:		participantKeys,
			RoutingAdvertisementKey:	StoreV2RoutingKeyForChannel(channel),
		}
		layout.Channels = append(layout.Channels, record.Normalize())
		addLatestHashCheckpoint := true
		for _, condition := range channel.LatestState.Conditions {
			appendCondition(storeV2ConditionFromPayment(channel, condition, false))
		}
		if channel.PendingClose.State.StateHash != "" {
			if channel.PendingClose.State.Nonce == channel.LatestState.Nonce {
				addLatestHashCheckpoint = false
			}
			pending := StoreV2PendingCloseRecord{
				Key:		StoreV2PendingCloseKey(channel.ChannelID),
				Version:	StoreV2MigrationVersion,
				ChannelID:	channel.ChannelID,
				Close:		channel.PendingClose,
			}.Normalize()
			layout.PendingCloses = append(layout.PendingCloses, pending)
			layout.Channels[len(layout.Channels)-1].PendingCloseKey = pending.Key
			layout.ChannelStates = append(layout.ChannelStates, StoreV2ChannelStateRecord{
				Key:			StoreV2ChannelStateKey(channel.ChannelID, channel.PendingClose.State.Nonce),
				Version:		StoreV2MigrationVersion,
				ChannelID:		channel.ChannelID,
				Nonce:			channel.PendingClose.State.Nonce,
				StateHash:		channel.PendingClose.State.StateHash,
				FullState:		channel.PendingClose.State,
				SubmittedOnChain:	true,
				CheckpointHeight:	channel.PendingClose.SubmittedHeight,
			}.Normalize())
			for _, condition := range channel.PendingClose.State.Conditions {
				appendCondition(storeV2ConditionFromPayment(channel, condition, false))
			}
			for _, proof := range channel.PendingClose.FraudProofs {
				layout.FraudProofs = append(layout.FraudProofs, StoreV2FraudProofRecord{
					Key:		StoreV2FraudProofKey(proof.ProofID),
					Version:	StoreV2MigrationVersion,
					ProofID:	proof.ProofID,
					ChannelID:	channel.ChannelID,
					Proof:		proof,
				}.Normalize())
			}
		}
		if addLatestHashCheckpoint {
			layout.ChannelStates = append(layout.ChannelStates, StoreV2ChannelStateRecord{
				Key:			StoreV2ChannelStateKey(channel.ChannelID, channel.LatestState.Nonce),
				Version:		StoreV2MigrationVersion,
				ChannelID:		channel.ChannelID,
				Nonce:			channel.LatestState.Nonce,
				StateHash:		channel.LatestState.StateHash,
				FullState:		compactStoreV2State(channel.LatestState),
				SubmittedOnChain:	false,
				CheckpointHeight:	channel.OpenHeight,
			}.Normalize())
		}
	}
	for _, vc := range state.VirtualChannels {
		vc = vc.Normalize()
		layout.VirtualChannels = append(layout.VirtualChannels, StoreV2VirtualChannelRecord{
			Key:			StoreV2VirtualChannelKey(vc.VirtualChannelID),
			Version:		StoreV2MigrationVersion,
			VirtualChannelID:	vc.VirtualChannelID,
			Channel:		vc,
			AnchorHash:		vc.AnchorCommitment,
		}.Normalize())
	}
	for _, tombstone := range state.ClosedChannels {
		tombstone = tombstone.Normalize()
		layout.SettlementTombstones = append(layout.SettlementTombstones, StoreV2SettlementTombstoneRecord{
			Key:			StoreV2SettlementTombstoneKey(tombstone.ChannelID),
			Version:		StoreV2MigrationVersion,
			ChannelID:		tombstone.ChannelID,
			Tombstone:		tombstone,
			PruneAfterHeight:	tombstone.ExpiresHeight,
		}.Normalize())
	}
	layout = layout.Normalize()
	if err := layout.Validate(); err != nil {
		return StoreV2Layout{}, err
	}
	return layout, nil
}

func PruneStoreV2Layout(layout StoreV2Layout, currentHeight uint64) (StoreV2Layout, error) {
	if currentHeight == 0 {
		return StoreV2Layout{}, errors.New("payments store v2 prune height must be positive")
	}
	layout = layout.Normalize()
	pruned := layout
	pruned.SettlementTombstones = pruned.SettlementTombstones[:0]
	for _, tombstone := range layout.SettlementTombstones {
		tombstone = tombstone.Normalize()
		if tombstone.PruneAfterHeight == 0 || tombstone.PruneAfterHeight >= currentHeight {
			pruned.SettlementTombstones = append(pruned.SettlementTombstones, tombstone)
		}
	}
	pruned.Conditions = pruned.Conditions[:0]
	for _, condition := range layout.Conditions {
		condition = condition.Normalize()
		if !condition.Settled && (condition.ExpiresHeight == 0 || condition.ExpiresHeight >= currentHeight) {
			pruned.Conditions = append(pruned.Conditions, condition)
		}
	}
	pruned = pruned.Normalize()
	return pruned, pruned.Validate()
}

func QueryStoreV2ParticipantChannels(layout StoreV2Layout, req ParticipantChannelPageRequest) (ParticipantChannelPageResponse, error) {
	layout = layout.Normalize()
	address := strings.TrimSpace(req.Address)
	if err := addressing.ValidateUserAddress("payments store v2 participant query address", address); err != nil {
		return ParticipantChannelPageResponse{}, err
	}
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}
	matches := []StoreV2ParticipantChannelRecord{}
	for _, entry := range layout.ParticipantChannels {
		entry = entry.Normalize()
		if entry.Participant == address {
			matches = append(matches, entry)
		}
	}
	total := uint64(len(matches))
	if req.Offset >= total {
		return ParticipantChannelPageResponse{Entries: []StoreV2ParticipantChannelRecord{}, Total: total}, nil
	}
	end := req.Offset + limit
	if end > total {
		end = total
	}
	next := uint64(0)
	if end < total {
		next = end
	}
	return ParticipantChannelPageResponse{
		Entries:	append([]StoreV2ParticipantChannelRecord(nil), matches[req.Offset:end]...),
		NextOffset:	next,
		Total:		total,
	}, nil
}

func BuildAdaptiveSyncSnapshot(state PaymentsState, height uint64) (AdaptiveSyncSnapshot, error) {
	if height == 0 {
		return AdaptiveSyncSnapshot{}, errors.New("payments adaptive sync snapshot height must be positive")
	}
	state = state.Export()
	if err := state.Validate(); err != nil {
		return AdaptiveSyncSnapshot{}, err
	}
	layout, err := BuildStoreV2Layout(state)
	if err != nil {
		return AdaptiveSyncSnapshot{}, err
	}
	snapshot := AdaptiveSyncSnapshot{
		Key:				StoreV2AdaptiveSnapshotKey(height),
		Version:			StoreV2MigrationVersion,
		Height:				height,
		Layout:				layout,
		ConsensusOnly:			true,
		RoutingTopologyExcluded:	true,
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status == ChannelStatusPendingClose && channel.PendingClose.State.StateHash != "" {
			if channel.PendingClose.DisputeCount > 0 || channel.Finality == ChannelFinalityInDispute {
				snapshot.ActiveDisputes = append(snapshot.ActiveDisputes, AdaptiveSyncActiveDisputeIndex{
					Key:			StoreV2ActiveDisputeKey(channel.ChannelID),
					ChannelID:		channel.ChannelID,
					PendingStateHash:	channel.PendingClose.State.StateHash,
					PendingNonce:		channel.PendingClose.State.Nonce,
					SubmittedHeight:	channel.PendingClose.SubmittedHeight,
					SettleAfterHeight:	channel.PendingClose.SettleAfterHeight,
					DisputeCount:		channel.PendingClose.DisputeCount,
					Submitter:		channel.PendingClose.Submitter,
				}.Normalize())
			}
			if pendingHeight, ok := PendingFinalizationHeightForChannel(channel); ok {
				snapshot.PendingFinalizations = append(snapshot.PendingFinalizations, AdaptiveSyncPendingFinalizationIndex{
					Key:			StoreV2PendingFinalizationKey(channel.ChannelID),
					ChannelID:		channel.ChannelID,
					PendingHeight:		pendingHeight,
					Finality:		channel.Finality,
					PendingStateHash:	channel.PendingClose.State.StateHash,
					PendingNonce:		channel.PendingClose.State.Nonce,
				}.Normalize())
			}
		}
	}
	for _, event := range state.Events {
		event = event.Normalize()
		snapshot.WatcherReplayEvents = append(snapshot.WatcherReplayEvents, AdaptiveSyncWatcherReplayEvent{
			Key:		StoreV2WatcherReplayEventKey(event.Height, event.EventID),
			Event:		event,
			EventHash:	AdaptiveSyncEventHash(event),
		}.Normalize())
	}
	snapshot = snapshot.Normalize()
	snapshot.SnapshotHash = ComputeAdaptiveSyncSnapshotHash(snapshot)
	if err := snapshot.Validate(); err != nil {
		return AdaptiveSyncSnapshot{}, err
	}
	return snapshot, nil
}

func RecoverAdaptiveSyncSafety(snapshot AdaptiveSyncSnapshot) (AdaptiveSyncRecoveryState, error) {
	snapshot = snapshot.Normalize()
	if err := snapshot.Validate(); err != nil {
		return AdaptiveSyncRecoveryState{}, err
	}
	recovered := AdaptiveSyncRecoveryState{RecoveredFromSnapshotHash: snapshot.SnapshotHash}
	for _, channel := range snapshot.Layout.Channels {
		recovered.ActiveChannelIDs = append(recovered.ActiveChannelIDs, channel.ChannelID)
		if channel.PendingCloseKey != "" {
			recovered.PendingCloseChannelIDs = append(recovered.PendingCloseChannelIDs, channel.ChannelID)
		}
	}
	for _, condition := range snapshot.Layout.Conditions {
		if !condition.Settled {
			recovered.UnresolvedConditionIDs = append(recovered.UnresolvedConditionIDs, condition.ConditionID)
		}
	}
	for _, vc := range snapshot.Layout.VirtualChannels {
		recovered.VirtualChannelIDs = append(recovered.VirtualChannelIDs, vc.VirtualChannelID)
	}
	for _, tombstone := range snapshot.Layout.SettlementTombstones {
		recovered.SettlementTombstoneIDs = append(recovered.SettlementTombstoneIDs, tombstone.ChannelID)
	}
	for _, dispute := range snapshot.ActiveDisputes {
		recovered.ActiveDisputeChannelIDs = append(recovered.ActiveDisputeChannelIDs, dispute.ChannelID)
	}
	for _, pending := range snapshot.PendingFinalizations {
		recovered.PendingFinalizationIDs = append(recovered.PendingFinalizationIDs, pending.ChannelID)
	}
	for _, event := range snapshot.WatcherReplayEvents {
		recovered.WatcherReplayEventIDs = append(recovered.WatcherReplayEventIDs, event.Event.EventID)
	}
	sortStrings(recovered.ActiveChannelIDs)
	sortStrings(recovered.PendingCloseChannelIDs)
	sortStrings(recovered.UnresolvedConditionIDs)
	sortStrings(recovered.VirtualChannelIDs)
	sortStrings(recovered.SettlementTombstoneIDs)
	sortStrings(recovered.ActiveDisputeChannelIDs)
	sortStrings(recovered.PendingFinalizationIDs)
	sortStrings(recovered.WatcherReplayEventIDs)
	return recovered, nil
}

func (s PaymentsState) Export() PaymentsState {
	out := s.Clone()
	sortChannels(out.Channels)
	sortEdges(out.Edges)
	sortVirtualChannels(out.VirtualChannels)
	sortSettlements(out.Settlements)
	sortBatches(out.Batches)
	sortCustodyLocks(out.CustodyLocks)
	sortClosedChannelTombstones(out.ClosedChannels)
	sortConditionClaimRecords(out.ConditionClaims)
	sortValidatorPaymentServices(out.ValidatorPaymentServices)
	sortValidatorWatchRegistrations(out.ValidatorWatchRegistries)
	sortPaymentFeeMultipliers(out.FeeMultipliers)
	sortPaymentFeeCharges(out.FeeCharges)
	sortPaymentFeeRefunds(out.FeeRefunds)
	sortSecurityReserveAllocationHooks(out.SecurityReserveHooks)
	sortSettlementInclusionLatencies(out.InclusionLatencies)
	sortAsyncFinalizationJobs(out.AsyncFinalizationQueue)
	sortAsyncPromiseExpiryJobs(out.AsyncPromiseExpiryQueue)
	sortAsyncSettlementCompletions(out.AsyncCompletions)
	return out
}

func (s PaymentsState) Clone() PaymentsState {
	out := PaymentsState{
		Channels:			make([]ChannelRecord, len(s.Channels)),
		Edges:				make([]ChannelEdge, len(s.Edges)),
		VirtualChannels:		make([]VirtualChannel, len(s.VirtualChannels)),
		Settlements:			make([]SettlementRecord, len(s.Settlements)),
		Batches:			make([]SettlementBatch, len(s.Batches)),
		CustodyLocks:			make([]CustodyLock, len(s.CustodyLocks)),
		ClosedChannels:			make([]ClosedChannelTombstone, len(s.ClosedChannels)),
		ConditionClaims:		make([]ConditionClaimRecord, len(s.ConditionClaims)),
		ValidatorPaymentServices:	make([]ValidatorPaymentServiceMetadata, len(s.ValidatorPaymentServices)),
		ValidatorWatchRegistries:	make([]ValidatorWatchRegistration, len(s.ValidatorWatchRegistries)),
		FeeSchedule:			s.FeeSchedule.Normalize(),
		FeeMultipliers:			make([]PaymentFeeMultiplier, len(s.FeeMultipliers)),
		FeeCharges:			make([]PaymentFeeCharge, len(s.FeeCharges)),
		FeeRefunds:			make([]PaymentFeeRefund, len(s.FeeRefunds)),
		SecurityReserveHooks:		make([]SecurityReserveAllocationHook, len(s.SecurityReserveHooks)),
		InclusionLatencies:		make([]SettlementInclusionLatency, len(s.InclusionLatencies)),
		AsyncFinalizationQueue:		make([]AsyncFinalizationJob, len(s.AsyncFinalizationQueue)),
		AsyncPromiseExpiryQueue:	make([]AsyncPromiseExpiryJob, len(s.AsyncPromiseExpiryQueue)),
		AsyncCompletions:		make([]AsyncSettlementCompletion, len(s.AsyncCompletions)),
		Events:				make([]PaymentEvent, len(s.Events)),
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
	for i, lock := range s.CustodyLocks {
		out.CustodyLocks[i] = lock.Normalize()
	}
	for i, tombstone := range s.ClosedChannels {
		out.ClosedChannels[i] = tombstone.Normalize()
	}
	for i, claim := range s.ConditionClaims {
		out.ConditionClaims[i] = claim.Normalize()
	}
	for i, metadata := range s.ValidatorPaymentServices {
		out.ValidatorPaymentServices[i] = metadata.Normalize()
	}
	for i, registration := range s.ValidatorWatchRegistries {
		out.ValidatorWatchRegistries[i] = registration.Normalize()
	}
	for i, multiplier := range s.FeeMultipliers {
		out.FeeMultipliers[i] = multiplier.Normalize()
	}
	for i, charge := range s.FeeCharges {
		out.FeeCharges[i] = charge.Normalize()
	}
	for i, refund := range s.FeeRefunds {
		out.FeeRefunds[i] = refund.Normalize()
	}
	for i, hook := range s.SecurityReserveHooks {
		out.SecurityReserveHooks[i] = hook.Normalize()
	}
	for i, latency := range s.InclusionLatencies {
		out.InclusionLatencies[i] = latency.Normalize()
	}
	for i, job := range s.AsyncFinalizationQueue {
		out.AsyncFinalizationQueue[i] = job.Normalize()
	}
	for i, job := range s.AsyncPromiseExpiryQueue {
		out.AsyncPromiseExpiryQueue[i] = job.Normalize()
	}
	for i, completion := range s.AsyncCompletions {
		out.AsyncCompletions[i] = completion.Normalize()
	}
	for i, event := range s.Events {
		out.Events[i] = event.Normalize()
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
	if err := validateBatches(s.Channels, s.Batches); err != nil {
		return err
	}
	if err := validateCustodyLocks(s.Channels, s.CustodyLocks); err != nil {
		return err
	}
	if err := ValidateLockedCollateralForFinality(s); err != nil {
		return err
	}
	if err := validateClosedChannelTombstones(s.Channels, s.ClosedChannels); err != nil {
		return err
	}
	if err := validateConditionClaimRecords(s.Channels, s.ConditionClaims); err != nil {
		return err
	}
	if err := validateValidatorPaymentServices(s.ValidatorPaymentServices); err != nil {
		return err
	}
	if err := validateValidatorWatchRegistrations(s.ValidatorPaymentServices, s.ValidatorWatchRegistries); err != nil {
		return err
	}
	if err := s.FeeSchedule.Normalize().Validate(); err != nil {
		return err
	}
	if err := validatePaymentFeeMultipliers(s.FeeSchedule.Normalize(), s.FeeMultipliers); err != nil {
		return err
	}
	if err := validatePaymentFeeCharges(s.FeeCharges); err != nil {
		return err
	}
	if err := validatePaymentFeeRefunds(s.FeeCharges, s.FeeRefunds); err != nil {
		return err
	}
	if err := validateSecurityReserveAllocationHooks(s.Channels, s.SecurityReserveHooks); err != nil {
		return err
	}
	if err := validateSettlementInclusionLatencies(s.Channels, s.InclusionLatencies); err != nil {
		return err
	}
	if err := validateAsyncFinalizationJobs(s.Channels, s.AsyncFinalizationQueue); err != nil {
		return err
	}
	if err := validateAsyncPromiseExpiryJobs(s.Channels, s.AsyncPromiseExpiryQueue); err != nil {
		return err
	}
	if err := validateAsyncSettlementCompletions(s.Channels, s.AsyncFinalizationQueue, s.AsyncPromiseExpiryQueue, s.AsyncCompletions); err != nil {
		return err
	}
	return validatePaymentEvents(s.Channels, s.Events)
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

func (s PaymentsState) ValidatorPaymentServiceByValidator(validatorAddress string) (ValidatorPaymentServiceMetadata, bool) {
	validatorAddress = strings.TrimSpace(validatorAddress)
	for _, metadata := range s.ValidatorPaymentServices {
		metadata = metadata.Normalize()
		if metadata.ValidatorAddress == validatorAddress {
			return metadata, true
		}
	}
	return ValidatorPaymentServiceMetadata{}, false
}

func (s PaymentsState) ValidatorWatchRegistration(validatorAddress, delegator string) (ValidatorWatchRegistration, bool) {
	validatorAddress = strings.TrimSpace(validatorAddress)
	delegator = strings.TrimSpace(delegator)
	for _, registration := range s.ValidatorWatchRegistries {
		registration = registration.Normalize()
		if registration.ValidatorAddress == validatorAddress && registration.Delegator == delegator {
			return registration, true
		}
	}
	return ValidatorWatchRegistration{}, false
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
	_, vc, found := s.VirtualChannelIndex(id)
	return vc, found
}

func (s PaymentsState) VirtualChannelIndex(id string) (int, VirtualChannel, bool) {
	needle := normalizeHash(id)
	for i, vc := range s.VirtualChannels {
		vc = vc.Normalize()
		if vc.VirtualChannelID == needle {
			return i, vc, true
		}
	}
	return 0, VirtualChannel{}, false
}

func (s PaymentsState) StateHashDebug(channelID string) (StateHashDebug, error) {
	channel, found := s.Export().ChannelByID(channelID)
	if !found {
		return StateHashDebug{}, errors.New("payments channel not found")
	}
	debug := StateHashDebug{
		ChannelID:			channel.ChannelID,
		Status:				channel.Status,
		LatestNonce:			channel.LatestState.Nonce,
		LatestStateHash:		channel.LatestState.StateHash,
		ComputedLatestStateHash:	ComputeStateHash(channel.LatestState),
		FinalizedNonce:			channel.FinalizedNonce,
		DisputedNonce:			channel.DisputedNonce,
	}
	if channel.PendingClose.State.StateHash != "" {
		debug.PendingNonce = channel.PendingClose.State.Nonce
		debug.PendingStateHash = channel.PendingClose.State.StateHash
		debug.ComputedPendingStateHash = ComputeStateHash(channel.PendingClose.State)
	}
	return debug, nil
}

func (s PaymentsState) CustodyLockByChannel(channelID string) (CustodyLock, bool) {
	needle := normalizeHash(channelID)
	for _, lock := range s.CustodyLocks {
		lock = lock.Normalize()
		if lock.ChannelID == needle {
			return lock, true
		}
	}
	return CustodyLock{}, false
}

func (s PaymentsState) PendingFinalizationHeight(channelID string) (uint64, bool, error) {
	state := s.Export()
	channel, found := state.ChannelByID(channelID)
	if !found {
		return 0, false, errors.New("payments channel not found")
	}
	height, ok := PendingFinalizationHeightForChannel(channel)
	return height, ok, nil
}

func AdvanceChannelFinality(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments finality advance height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	nextFinality := FinalityAfterPendingClose(channel, currentHeight)
	if nextFinality == channel.Finality {
		return state, nil
	}
	next := state.Clone()
	nextChannel, err := setChannelFinality(channel, nextFinality, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func ValidateLockedCollateralForFinality(state PaymentsState) error {
	state = state.Export()
	lockByChannel := make(map[string]CustodyLock, len(state.CustodyLocks))
	for _, lock := range state.CustodyLocks {
		lock = lock.Normalize()
		lockByChannel[lock.ChannelID] = lock
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		lock, locked := lockByChannel[channel.ChannelID]
		switch channel.Finality {
		case ChannelFinalitySettled, ChannelFinalityPenalized:
			if channel.Status == ChannelStatusSettled {
				if locked {
					return errors.New("payments settled finality must not retain custody lock")
				}
				continue
			}
		}
		if !locked {
			return errors.New("payments unsettled finality must retain custody lock")
		}
		if err := lock.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	return nil
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

func validateCustodyLocks(channels []ChannelRecord, locks []CustodyLock) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(locks))
	var previous string
	for i, lock := range locks {
		lock = lock.Normalize()
		channel, found := channelByID[lock.ChannelID]
		if !found {
			return errors.New("payments custody lock references unknown channel")
		}
		if channel.Status == ChannelStatusSettled {
			return errors.New("payments settled channel must not retain custody lock")
		}
		if err := lock.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[lock.ChannelID]; found {
			return errors.New("payments duplicate custody lock")
		}
		seen[lock.ChannelID] = struct{}{}
		if i > 0 && previous >= lock.ChannelID {
			return errors.New("payments custody locks must be sorted canonically")
		}
		previous = lock.ChannelID
	}
	for _, channel := range channelByID {
		if channel.Status == ChannelStatusSettled {
			continue
		}
		if _, found := seen[channel.ChannelID]; !found {
			return errors.New("payments channel custody lock is required")
		}
	}
	return nil
}

func validateClosedChannelTombstones(channels []ChannelRecord, tombstones []ClosedChannelTombstone) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(tombstones))
	var previous string
	for i, tombstone := range tombstones {
		tombstone = tombstone.Normalize()
		if err := tombstone.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[tombstone.ChannelID]
		if !found {
			return errors.New("payments tombstone references unknown channel")
		}
		if channel.Status != ChannelStatusSettled {
			return errors.New("payments tombstone requires settled channel")
		}
		if tombstone.ChainID != channel.ChainID || tombstone.FinalizedNonce != channel.FinalizedNonce {
			return errors.New("payments tombstone channel domain mismatch")
		}
		if _, found := seen[tombstone.ChannelID]; found {
			return errors.New("payments duplicate closed channel tombstone")
		}
		seen[tombstone.ChannelID] = struct{}{}
		if i > 0 && previous >= tombstone.ChannelID {
			return errors.New("payments closed channel tombstones must be sorted canonically")
		}
		previous = tombstone.ChannelID
	}
	for _, channel := range channelByID {
		if channel.Status != ChannelStatusSettled {
			continue
		}
		if _, found := seen[channel.ChannelID]; !found {
			return errors.New("payments settled channel tombstone is required")
		}
	}
	return nil
}

func validateConditionClaimRecords(channels []ChannelRecord, claims []ConditionClaimRecord) error {
	channelByID := channelMap(channels)
	seenCondition := make(map[string]struct{}, len(claims))
	seenEvidence := make(map[string]struct{}, len(claims))
	var previous string
	for i, claim := range claims {
		claim = claim.Normalize()
		if err := claim.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[claim.ChannelID]
		if !found {
			return errors.New("payments condition claim references unknown channel")
		}
		if claim.ChainID != channel.ChainID {
			return errors.New("payments condition claim channel domain mismatch")
		}
		conditionKey := conditionClaimKey(claim.ChannelID, claim.ConditionID)
		evidenceKey := conditionEvidenceKey(claim.ChannelID, claim.EvidenceHash)
		if _, found := seenCondition[conditionKey]; found {
			return errors.New("payments duplicate condition claim")
		}
		if _, found := seenEvidence[evidenceKey]; found {
			return errors.New("payments duplicate condition evidence claim")
		}
		seenCondition[conditionKey] = struct{}{}
		seenEvidence[evidenceKey] = struct{}{}
		sortKey := conditionKey + "/" + claim.EvidenceHash
		if i > 0 && previous >= sortKey {
			return errors.New("payments condition claims must be sorted canonically")
		}
		previous = sortKey
	}
	return nil
}

func validateValidatorPaymentServices(services []ValidatorPaymentServiceMetadata) error {
	seen := make(map[string]struct{}, len(services))
	var previous string
	for i, metadata := range services {
		metadata = metadata.Normalize()
		if err := metadata.Validate(); err != nil {
			return err
		}
		if metadata.MetadataHash == "" {
			return errors.New("payments validator service metadata hash is required")
		}
		if _, found := seen[metadata.ValidatorAddress]; found {
			return errors.New("payments duplicate validator payment service")
		}
		seen[metadata.ValidatorAddress] = struct{}{}
		if i > 0 && previous >= metadata.ValidatorAddress {
			return errors.New("payments validator services must be sorted canonically")
		}
		previous = metadata.ValidatorAddress
	}
	return nil
}

func validateValidatorWatchRegistrations(services []ValidatorPaymentServiceMetadata, registrations []ValidatorWatchRegistration) error {
	serviceByValidator := make(map[string]ValidatorPaymentServiceMetadata, len(services))
	for _, metadata := range services {
		metadata = metadata.Normalize()
		serviceByValidator[metadata.ValidatorAddress] = metadata
	}
	seen := make(map[string]struct{}, len(registrations))
	var previous string
	for i, registration := range registrations {
		registration = registration.Normalize()
		metadata, found := serviceByValidator[registration.ValidatorAddress]
		if !found {
			return errors.New("payments validator watch registration references unknown service")
		}
		if err := registration.Validate(metadata); err != nil {
			return err
		}
		key := validatorWatchRegistrationKey(registration.ValidatorAddress, registration.Delegator)
		if _, found := seen[key]; found {
			return errors.New("payments duplicate validator watch registration")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("payments validator watch registrations must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validatePaymentFeeMultipliers(schedule PaymentFeeSchedule, multipliers []PaymentFeeMultiplier) error {
	seen := make(map[PaymentFeeClass]struct{}, len(multipliers))
	var previous string
	for i, multiplier := range multipliers {
		multiplier = multiplier.Normalize()
		if err := multiplier.Validate(); err != nil {
			return err
		}
		if multiplier.MultiplierBps > schedule.MaxMultiplierBps {
			return errors.New("payments fee multiplier exceeds schedule maximum")
		}
		if _, found := seen[multiplier.FeeClass]; found {
			return errors.New("payments duplicate fee multiplier")
		}
		seen[multiplier.FeeClass] = struct{}{}
		key := string(multiplier.FeeClass)
		if i > 0 && previous >= key {
			return errors.New("payments fee multipliers must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validatePaymentFeeCharges(charges []PaymentFeeCharge) error {
	seen := make(map[string]struct{}, len(charges))
	var previous string
	for i, charge := range charges {
		charge = charge.Normalize()
		if err := charge.Validate(); err != nil {
			return err
		}
		if _, found := seen[charge.FeeID]; found {
			return errors.New("payments duplicate fee charge")
		}
		seen[charge.FeeID] = struct{}{}
		if i > 0 && previous >= charge.FeeID {
			return errors.New("payments fee charges must be sorted canonically")
		}
		previous = charge.FeeID
	}
	return nil
}

func validatePaymentFeeRefunds(charges []PaymentFeeCharge, refunds []PaymentFeeRefund) error {
	chargeByID := make(map[string]PaymentFeeCharge, len(charges))
	for _, charge := range charges {
		charge = charge.Normalize()
		chargeByID[charge.FeeID] = charge
	}
	seen := make(map[string]struct{}, len(refunds))
	var previous string
	for i, refund := range refunds {
		refund = refund.Normalize()
		if err := refund.Validate(); err != nil {
			return err
		}
		charge, found := chargeByID[refund.FeeID]
		if !found {
			return errors.New("payments fee refund references unknown charge")
		}
		if !charge.Refunded {
			return errors.New("payments fee refund requires refunded charge marker")
		}
		if refund.Amount != charge.Amount {
			return errors.New("payments fee refund amount must match charge")
		}
		if _, found := seen[refund.RefundID]; found {
			return errors.New("payments duplicate fee refund")
		}
		seen[refund.RefundID] = struct{}{}
		if i > 0 && previous >= refund.RefundID {
			return errors.New("payments fee refunds must be sorted canonically")
		}
		previous = refund.RefundID
	}
	return nil
}

func validateSecurityReserveAllocationHooks(channels []ChannelRecord, hooks []SecurityReserveAllocationHook) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(hooks))
	var previous string
	for i, hook := range hooks {
		hook = hook.Normalize()
		channel, found := channelByID[hook.ChannelID]
		if !found {
			return errors.New("payments security reserve hook channel not found")
		}
		if err := hook.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[hook.HookID]; found {
			return errors.New("payments duplicate security reserve hook")
		}
		seen[hook.HookID] = struct{}{}
		if i > 0 && previous >= hook.HookID {
			return errors.New("payments security reserve hooks must be sorted canonically")
		}
		previous = hook.HookID
	}
	return nil
}

func validateSettlementInclusionLatencies(channels []ChannelRecord, records []SettlementInclusionLatency) error {
	seen := make(map[string]struct{}, len(records))
	var previous string
	for i, record := range records {
		record = record.Normalize()
		if err := record.Validate(channels); err != nil {
			return err
		}
		if _, found := seen[record.RecordID]; found {
			return errors.New("payments duplicate settlement inclusion latency")
		}
		seen[record.RecordID] = struct{}{}
		if i > 0 && previous >= record.RecordID {
			return errors.New("payments settlement inclusion latencies must be sorted canonically")
		}
		previous = record.RecordID
	}
	return nil
}

func validateAsyncFinalizationJobs(channels []ChannelRecord, jobs []AsyncFinalizationJob) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(jobs))
	var previous string
	for i, job := range jobs {
		job = job.Normalize()
		if err := job.Validate(); err != nil {
			return err
		}
		if _, found := channelByID[job.ChannelID]; !found {
			return errors.New("payments async finalization references unknown channel")
		}
		if _, found := seen[job.JobID]; found {
			return errors.New("payments duplicate async finalization job")
		}
		seen[job.JobID] = struct{}{}
		if i > 0 && previous >= job.JobID {
			return errors.New("payments async finalization jobs must be sorted canonically")
		}
		previous = job.JobID
	}
	return nil
}

func validateAsyncPromiseExpiryJobs(channels []ChannelRecord, jobs []AsyncPromiseExpiryJob) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(jobs))
	var previous string
	for i, job := range jobs {
		job = job.Normalize()
		if err := job.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[job.ChannelID]
		if !found {
			return errors.New("payments async promise expiry references unknown channel")
		}
		if err := job.Promise.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[job.JobID]; found {
			return errors.New("payments duplicate async promise expiry job")
		}
		seen[job.JobID] = struct{}{}
		if i > 0 && previous >= job.JobID {
			return errors.New("payments async promise expiry jobs must be sorted canonically")
		}
		previous = job.JobID
	}
	return nil
}

func validateAsyncSettlementCompletions(channels []ChannelRecord, finalizationJobs []AsyncFinalizationJob, expiryJobs []AsyncPromiseExpiryJob, completions []AsyncSettlementCompletion) error {
	channelByID := channelMap(channels)
	jobIDs := make(map[string]struct{}, len(finalizationJobs)+len(expiryJobs))
	for _, job := range finalizationJobs {
		jobIDs[job.Normalize().JobID] = struct{}{}
	}
	for _, job := range expiryJobs {
		jobIDs[job.Normalize().JobID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(completions))
	var previous string
	for i, completion := range completions {
		completion = completion.Normalize()
		if err := completion.Validate(); err != nil {
			return err
		}
		if _, found := channelByID[completion.ChannelID]; !found {
			return errors.New("payments async completion references unknown channel")
		}
		if _, found := jobIDs[completion.JobID]; !found {
			return errors.New("payments async completion references unknown job")
		}
		if _, found := seen[completion.CompletionID]; found {
			return errors.New("payments duplicate async completion")
		}
		seen[completion.CompletionID] = struct{}{}
		if i > 0 && previous >= completion.CompletionID {
			return errors.New("payments async completions must be sorted canonically")
		}
		previous = completion.CompletionID
	}
	return nil
}

func validatePaymentEvents(channels []ChannelRecord, events []PaymentEvent) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(events))
	openEventByChannel := make(map[string]struct{}, len(channels))
	for _, event := range events {
		event = event.Normalize()
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := channelByID[event.ChannelID]; !found {
			return errors.New("payments event references unknown channel")
		}
		if _, found := seen[event.EventID]; found {
			return errors.New("payments duplicate event")
		}
		seen[event.EventID] = struct{}{}
		if event.EventType == "channel-open" {
			openEventByChannel[event.ChannelID] = struct{}{}
		}
	}
	for _, channel := range channelByID {
		if _, found := openEventByChannel[channel.ChannelID]; !found {
			return errors.New("payments channel-open event is required")
		}
	}
	return nil
}

func applySettlementAdjustments(balances []Balance, penalties []Penalty, allocations []PenaltyAllocation, feeText, feePayer string) ([]Balance, error) {
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
	for _, allocation := range normalizePenaltyAllocations(allocations) {
		amount, err := parsePositiveInt("payments penalty allocation amount", allocation.Amount)
		if err != nil {
			return nil, err
		}
		offenderBalance, found := amounts[allocation.Offender]
		if !found || offenderBalance.LT(amount) {
			return nil, errors.New("payments penalty allocation exceeds offender balance")
		}
		amounts[allocation.Offender] = offenderBalance.Sub(amount)
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

func settlementBalancesWithConditions(state ChannelState, channel ChannelRecord, resolutions []ConditionResolution) ([]Balance, error) {
	state = state.Normalize()
	if len(state.Conditions) == 0 {
		return state.Balances, nil
	}
	amounts := make(map[string]sdkmath.Int, len(state.Balances))
	for _, balance := range normalizeBalances(state.Balances) {
		amount, err := parseNonNegativeInt("payments settlement base balance", balance.Amount)
		if err != nil {
			return nil, err
		}
		amounts[balance.Participant] = amount
	}
	reserveByParticipant := map[string]sdkmath.Int{}
	if state.ChannelType == ChannelTypeBidirectional {
		reserveA, err := parseNonNegativeInt("payments settlement reserve a", state.ReserveA)
		if err != nil {
			return nil, err
		}
		reserveB, err := parseNonNegativeInt("payments settlement reserve b", state.ReserveB)
		if err != nil {
			return nil, err
		}
		reserveByParticipant[state.ParticipantA] = reserveA
		reserveByParticipant[state.ParticipantB] = reserveB
	}
	resolutionByID := make(map[string]ConditionResolution, len(resolutions))
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		resolutionByID[resolution.ConditionID] = resolution
	}
	for _, condition := range state.Conditions {
		condition = condition.Normalize()
		resolution, found := resolutionByID[condition.ConditionID]
		if !found {
			return nil, errors.New("payments condition is unresolved")
		}
		amount, err := parsePositiveInt("payments condition amount", condition.Amount)
		if err != nil {
			return nil, err
		}
		reserve := reserveByParticipant[condition.Payer]
		if reserve.LT(amount) {
			return nil, errors.New("payments condition exceeds reserved balance")
		}
		reserveByParticipant[condition.Payer] = reserve.Sub(amount)
		recipient := resolution.Recipient
		amounts[recipient] = amounts[recipient].Add(amount)
	}
	for participant, reserve := range reserveByParticipant {
		amounts[participant] = amounts[participant].Add(reserve)
	}
	out := make([]Balance, 0, len(amounts))
	for participant, amount := range amounts {
		if !containsString(channel.Participants, participant) {
			return nil, errors.New("payments settlement condition participant must be in channel")
		}
		out = append(out, Balance{Participant: participant, Amount: amount.String()})
	}
	return normalizeBalances(out), nil
}

func rejectReusedConditionClaims(state PaymentsState, channel ChannelRecord, resolutions []ConditionResolution) error {
	channel = channel.Normalize()
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		conditionKey := conditionClaimKey(channel.ChannelID, resolution.ConditionID)
		evidenceKey := conditionEvidenceKey(channel.ChannelID, resolution.EvidenceHash)
		for _, existing := range state.ConditionClaims {
			existing = existing.Normalize()
			if existing.ChainID != channel.ChainID || existing.ChannelID != channel.ChannelID {
				continue
			}
			if conditionClaimKey(existing.ChannelID, existing.ConditionID) == conditionKey {
				return errors.New("payments condition claim has already been used")
			}
			if conditionEvidenceKey(existing.ChannelID, existing.EvidenceHash) == evidenceKey {
				return errors.New("payments condition evidence claim has already been used")
			}
		}
	}
	return nil
}

func appendSettlementReplayRecords(state *PaymentsState, channel ChannelRecord, settlement SettlementRecord, resolutions []ConditionResolution, height uint64) {
	channel = channel.Normalize()
	settlement = settlement.Normalize()
	tombstone := ClosedChannelTombstone{
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		FinalizedNonce:	settlement.Nonce,
		StateHash:	settlement.StateHash,
		ClosedHeight:	height,
		ExpiresHeight:	height + DefaultReplayHorizon,
	}.Normalize()
	state.ClosedChannels = upsertClosedChannelTombstone(state.ClosedChannels, tombstone)
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		state.ConditionClaims = append(state.ConditionClaims, ConditionClaimRecord{
			ChainID:	channel.ChainID,
			ChannelID:	channel.ChannelID,
			ConditionID:	resolution.ConditionID,
			EvidenceHash:	resolution.EvidenceHash,
			ResolvedHeight:	height,
			ExpiresHeight:	height + DefaultReplayHorizon,
		}.Normalize())
	}
}

func conditionRootUpdatesForPromises(state PaymentsState, promises []ConditionalPromise) ([]ConditionRootUpdate, error) {
	grouped := make(map[string][]ConditionalPromise)
	for _, promise := range normalizeConditionalPromises(promises) {
		grouped[promise.ChannelID] = append(grouped[promise.ChannelID], promise)
	}
	channelIDs := make([]string, 0, len(grouped))
	for channelID := range grouped {
		channelIDs = append(channelIDs, channelID)
	}
	sort.Strings(channelIDs)
	updates := make([]ConditionRootUpdate, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		channel, found := state.ChannelByID(channelID)
		if !found {
			return nil, errors.New("payments condition root update channel not found")
		}
		if len(channel.LatestState.Conditions) == 0 {
			continue
		}
		_, update, err := BuildConditionRootAfterExpiry(channel.LatestState, grouped[channelID])
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return normalizeConditionRootUpdates(updates), nil
}

func upsertClosedChannelTombstone(tombstones []ClosedChannelTombstone, next ClosedChannelTombstone) []ClosedChannelTombstone {
	out := make([]ClosedChannelTombstone, 0, len(tombstones)+1)
	replaced := false
	for _, tombstone := range tombstones {
		tombstone = tombstone.Normalize()
		if tombstone.ChannelID == next.ChannelID {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, tombstone)
	}
	if !replaced {
		out = append(out, next)
	}
	sortClosedChannelTombstones(out)
	return out
}

func finalBalancesForUnidirectionalClaim(channel ChannelRecord, claim UnidirectionalClaim, settlementFee, feePayer string) ([]Balance, error) {
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return nil, err
	}
	claimed, err := parseNonNegativeInt("payments claimed amount", claim.ClaimedAmount)
	if err != nil {
		return nil, err
	}
	if claimed.GT(collateral) {
		return nil, errors.New("payments claimed amount exceeds locked collateral")
	}
	return applySettlementAdjustments([]Balance{
		{Participant: channel.Payer, Amount: collateral.Sub(claimed).String()},
		{Participant: channel.Receiver, Amount: claimed.String()},
	}, nil, nil, settlementFee, feePayer)
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

func parentReservedCapacity(channel ChannelRecord) (sdkmath.Int, error) {
	channel = channel.Normalize()
	reserveA, err := parseNonNegativeInt("payments virtual parent reserve a", channel.LatestState.ReserveA)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}
	reserveB, err := parseNonNegativeInt("payments virtual parent reserve b", channel.LatestState.ReserveB)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}
	return reserveA.Add(reserveB), nil
}

func validateVirtualParentAccounting(state PaymentsState, vc VirtualChannel, routeTimeoutHeight uint64, aggregated bool, reserves []VirtualParentReserve) (string, error) {
	vc = vc.Normalize()
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return "", err
	}
	requiredByParent := map[string]sdkmath.Int{}
	if aggregated {
		for _, reserve := range normalizeVirtualParentReserves(reserves) {
			amount, err := virtualReserveAccountingAmount(reserve, true)
			if err != nil {
				return "", err
			}
			current := requiredByParent[reserve.ParentChannelID]
			if current.IsNil() {
				current = sdkmath.ZeroInt()
			}
			requiredByParent[reserve.ParentChannelID] = current.Add(amount)
		}
	}
	var parentChainID string
	for _, parentID := range vc.ParentChannelIDs {
		channel, found := state.ChannelByID(parentID)
		if !found || channel.Status != ChannelStatusOpen {
			return "", errors.New("payments virtual channel requires open parents")
		}
		if parentChainID == "" {
			parentChainID = channel.ChainID
		} else if parentChainID != channel.ChainID {
			return "", errors.New("payments virtual channel parents must share chain id")
		}
		if !containsString(channel.Participants, vc.Endpoints[0]) && !containsString(channel.Participants, vc.Endpoints[1]) && !sharesAny(channel.Participants, vc.Intermediaries) {
			return "", errors.New("payments virtual channel parent path must touch route participants")
		}
		reserved, err := parentReservedCapacity(channel)
		if err != nil {
			return "", err
		}
		required := capacity
		if aggregated {
			required = requiredByParent[parentID]
			if required.IsNil() || !required.IsPositive() {
				return "", errors.New("payments virtual aggregated reserve missing parent split")
			}
		}
		if reserved.LT(required) {
			return "", errors.New("payments virtual channel capacity exceeds parent reserved capacity")
		}
		if err := validateVirtualParentTimeout(vc, channel, routeTimeoutHeight); err != nil {
			return "", err
		}
	}
	return parentChainID, nil
}

func validateVirtualParentTimeout(vc VirtualChannel, channel ChannelRecord, routeTimeoutHeight uint64) error {
	vc = vc.Normalize()
	channel = channel.Normalize()
	parentSafetyHeight := channel.LatestState.TimeoutHeight
	if parentSafetyHeight == 0 {
		parentSafetyHeight = channel.OpenHeight + channel.CloseDelay + channel.DisputePeriod
	}
	safetyExpiry := vc.ExpiresHeight + channel.CloseDelay + channel.DisputePeriod
	if safetyExpiry < vc.ExpiresHeight || safetyExpiry >= parentSafetyHeight {
		return errors.New("payments virtual channel expiry must be earlier than parent safety timeout")
	}
	if routeTimeoutHeight > 0 {
		if routeTimeoutHeight > parentSafetyHeight {
			return errors.New("payments virtual route timeout exceeds parent safety timeout")
		}
		if safetyExpiry >= routeTimeoutHeight {
			return errors.New("payments virtual channel expiry must be earlier than route timeout")
		}
	}
	return nil
}

func virtualReserveAccountingAmount(reserve VirtualParentReserve, aggregated bool) (sdkmath.Int, error) {
	reserve = reserve.Normalize()
	if aggregated {
		return parsePositiveInt("payments virtual reserve split amount", reserve.SplitAmount)
	}
	return parsePositiveInt("payments virtual reserve capacity", reserve.Capacity)
}

func buildVirtualUpdateForCurrent(current VirtualChannel, nextVC VirtualChannel) (VirtualChannel, error) {
	current = current.Normalize()
	nextVC = nextVC.Normalize()
	nextVC.ChainID = current.ChainID
	nextVC.ParentRouteID = current.ParentRouteID
	nextVC.ParentChannelIDs = append([]string(nil), current.ParentChannelIDs...)
	nextVC.ParentReserveCommitments = append([]string(nil), current.ParentReserveCommitments...)
	nextVC.Endpoints = append([]string(nil), current.Endpoints...)
	nextVC.EndpointA = current.EndpointA
	nextVC.EndpointB = current.EndpointB
	nextVC.Intermediaries = append([]string(nil), current.Intermediaries...)
	nextVC.IntermediarySetHash = current.IntermediarySetHash
	nextVC.Capacity = current.Capacity
	nextVC.RoutingFeeAmount = current.RoutingFeeAmount
	nextVC.ExpiresHeight = current.ExpiresHeight
	nextVC.Status = current.Status
	nextVC.AnchorCommitment = ""
	nextVC.StateHash = ""
	preservedSignatures := nextVC.Signatures
	nextVC.Signatures = nil
	built, err := BuildVirtualChannel(nextVC)
	if err != nil {
		return VirtualChannel{}, err
	}
	built.Signatures = preservedSignatures
	return built.Normalize(), nil
}

func validateVirtualEndpointUpdate(current VirtualChannel, nextVC VirtualChannel) error {
	if nextVC.Normalize().Nonce <= current.Normalize().Nonce {
		return errors.New("payments virtual update nonce must strictly increase")
	}
	return validateVirtualEndpointSignedState(current, nextVC, true)
}

func validateVirtualEndpointSignedState(current VirtualChannel, nextVC VirtualChannel, requireNewer bool) error {
	current = current.Normalize()
	nextVC = nextVC.Normalize()
	if nextVC.VirtualChannelID != current.VirtualChannelID {
		return errors.New("payments virtual update channel mismatch")
	}
	if nextVC.ChainID != current.ChainID || nextVC.ParentRouteID != current.ParentRouteID {
		return errors.New("payments virtual update domain mismatch")
	}
	if strings.Join(nextVC.ParentChannelIDs, "/") != strings.Join(current.ParentChannelIDs, "/") {
		return errors.New("payments virtual update parent channel mismatch")
	}
	if strings.Join(nextVC.ParentReserveCommitments, "/") != strings.Join(current.ParentReserveCommitments, "/") {
		return errors.New("payments virtual update reserve commitment mismatch")
	}
	if strings.Join(nextVC.Endpoints, "/") != strings.Join(current.Endpoints, "/") || strings.Join(nextVC.Intermediaries, "/") != strings.Join(current.Intermediaries, "/") {
		return errors.New("payments virtual update route participant mismatch")
	}
	if nextVC.Capacity != current.Capacity || nextVC.RoutingFeeAmount != current.RoutingFeeAmount || nextVC.ExpiresHeight != current.ExpiresHeight {
		return errors.New("payments virtual update immutable field mismatch")
	}
	if requireNewer && nextVC.Nonce <= current.Nonce {
		return errors.New("payments virtual update nonce must strictly increase")
	}
	if !requireNewer && nextVC.Nonce < current.Nonce {
		return errors.New("payments virtual signed state nonce is stale")
	}
	if err := nextVC.ValidateCore(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(nextVC.Signatures))
	for _, sig := range nextVC.Signatures {
		sig = sig.Normalize()
		if err := ValidateVirtualChannelSignature(sig, nextVC); err != nil {
			return err
		}
		if !containsString(nextVC.Endpoints, sig.Signer) {
			return errors.New("payments virtual update requires endpoint signatures only")
		}
		seen[sig.Signer] = struct{}{}
	}
	for _, endpoint := range nextVC.Endpoints {
		if _, found := seen[endpoint]; !found {
			return errors.New("payments virtual update missing endpoint signature")
		}
	}
	return nil
}

func virtualReserveReleasesFromClose(proof VirtualCloseProof, current VirtualChannel) ([]VirtualReserveRelease, error) {
	proof = proof.Normalize()
	current = current.Normalize()
	commitments := proof.ParentReserveCommitments
	if len(commitments) == 0 {
		commitments = current.ParentReserveCommitments
	}
	if len(commitments) == 0 {
		for _, parentID := range current.ParentChannelIDs {
			commitments = append(commitments, HashParts("virtual-reserve-release", current.VirtualChannelID, parentID))
		}
	}
	capacity, err := parsePositiveInt("payments virtual capacity", current.Capacity)
	if err != nil {
		return nil, err
	}
	releases := make([]VirtualReserveRelease, 0, len(commitments))
	for i, commitment := range commitments {
		parentID := ""
		if i < len(current.ParentChannelIDs) {
			parentID = current.ParentChannelIDs[i]
		} else {
			parentID = current.ParentChannelIDs[len(current.ParentChannelIDs)-1]
		}
		amount := current.Capacity
		if len(commitments) > len(current.ParentChannelIDs) {
			share := capacity.QuoRaw(int64(len(commitments)))
			amount = share.String()
		}
		release := VirtualReserveRelease{
			SegmentID:		HashParts("virtual-release-segment", current.VirtualChannelID, commitment),
			VirtualChannelID:	current.VirtualChannelID,
			ParentChannelID:	parentID,
			ReserveCommitment:	commitment,
			Capacity:		amount,
			BalanceA:		proof.FinalState.BalanceA,
			BalanceB:		proof.FinalState.BalanceB,
			FeeAmount:		current.RoutingFeeAmount,
			ReleaseHeight:		proof.ReleaseHeight,
		}
		release.ReleaseHash = HashParts("virtual-reserve-release", release.SegmentID, release.VirtualChannelID, release.ParentChannelID, release.ReserveCommitment, release.Capacity, release.BalanceA, release.BalanceB, fmt.Sprintf("%020d", release.ReleaseHeight))
		releases = append(releases, release.Normalize())
	}
	sort.SliceStable(releases, func(i, j int) bool {
		return releases[i].SegmentID < releases[j].SegmentID
	})
	return releases, nil
}

func sharesAny(left, right []string) bool {
	for _, value := range left {
		if containsString(right, value) {
			return true
		}
	}
	return false
}

func virtualActivationReserveCommitments(proof VirtualActivationProof) []string {
	proof = proof.Normalize()
	out := make([]string, 0, len(proof.ParentReserves))
	for _, reserve := range proof.ParentReserves {
		out = append(out, reserve.ReserveCommitment)
	}
	return normalizeHashSlice(out)
}

type routeSearchPath struct {
	node	string
	edges	[]ChannelEdge
	cost	sdkmath.Int
	fee	sdkmath.Int
}

func selectPaymentRouteWithPolicy(state PaymentsState, store TopologyStore, req RouteSelectionRequest, amount sdkmath.Int) (ScoredRoute, error) {
	policy := req.Policy.Normalize()
	candidates := candidateRoutingEdges(state, store, amount, req.CurrentHeight, policy)
	if len(candidates) == 0 {
		return ScoredRoute{}, errors.New("payments scored route has no eligible edges")
	}
	sortEdges(candidates)
	queue := []routeSearchPath{{node: req.From, cost: sdkmath.ZeroInt(), fee: sdkmath.ZeroInt()}}
	bestByNode := map[string]sdkmath.Int{req.From: sdkmath.ZeroInt()}
	for len(queue) > 0 {
		sortRouteQueue(queue)
		current := queue[0]
		queue = queue[1:]
		if current.node == req.To && len(current.edges) > 0 {
			return buildScoredRoute(current.edges, amount, current.fee, current.cost)
		}
		if len(current.edges) >= policy.MaxHops {
			continue
		}
		for _, edge := range candidates {
			edge = edge.Normalize()
			if edge.From != current.node || routeContainsNode(current.edges, edge.To) {
				continue
			}
			weight, fee, err := routeEdgeWeight(store, edge, amount, req.CurrentHeight, policy)
			if err != nil {
				return ScoredRoute{}, err
			}
			nextCost := current.cost.Add(weight)
			nextFee := current.fee.Add(fee)
			if policy.MaxFeeAmount != "" {
				maxFee, err := parseNonNegativeInt("payments route policy max fee", policy.MaxFeeAmount)
				if err != nil {
					return ScoredRoute{}, err
				}
				if nextFee.GT(maxFee) {
					continue
				}
			}
			nextEdges := append([]ChannelEdge(nil), current.edges...)
			nextEdges = append(nextEdges, edge)
			if best, found := bestByNode[edge.To]; found && !nextCost.LT(best) {
				continue
			}
			bestByNode[edge.To] = nextCost
			queue = append(queue, routeSearchPath{node: edge.To, edges: nextEdges, cost: nextCost, fee: nextFee})
		}
	}
	return ScoredRoute{}, errors.New("payments scored route not found")
}

func candidateRoutingEdges(state PaymentsState, store TopologyStore, amount sdkmath.Int, currentHeight uint64, policy RoutePolicy) []ChannelEdge {
	combined := make([]ChannelEdge, 0, len(state.Edges)+len(store.Edges))
	for _, edge := range state.Edges {
		combined = upsertTopologyEdge(combined, edge)
	}
	for _, edge := range store.Edges {
		combined = upsertTopologyEdge(combined, edge)
	}
	active := activeEdgesForAmount(combined, amount, currentHeight)
	out := make([]ChannelEdge, 0, len(active))
	for _, edge := range active {
		edge = edge.Normalize()
		if routePolicyExcludesEdge(policy, edge) {
			continue
		}
		if !edgeEffectiveCapacityCovers(policy, edge, amount) {
			continue
		}
		channel, found := state.ChannelByID(edge.ChannelID)
		if !found || channel.Status != ChannelStatusOpen {
			continue
		}
		if !containsString(channel.Participants, edge.From) || !containsString(channel.Participants, edge.To) {
			continue
		}
		out = append(out, edge)
	}
	sortEdges(out)
	return out
}

func routeEdgeWeight(store TopologyStore, edge ChannelEdge, amount sdkmath.Int, currentHeight uint64, policy RoutePolicy) (sdkmath.Int, sdkmath.Int, error) {
	fee, err := parseNonNegativeInt("payments route edge fee", edge.FeeAmount)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
	}
	if policy.ProportionalFeeBps > 0 {
		fee = fee.Add(amount.Mul(sdkmath.NewInt(int64(policy.ProportionalFeeBps))).Quo(sdkmath.NewInt(10_000)))
	}
	cost := fee
	for _, penaltyText := range []string{policy.HopPenalty} {
		penalty, err := parseNonNegativeInt("payments route fixed penalty", penaltyText)
		if err != nil {
			return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
		}
		cost = cost.Add(penalty)
	}
	stats, hasStats := routeStatsForEdge(policy, edge)
	if hasStats {
		if stats.CongestionBps > 0 {
			cost = cost.Add(routeScaledPenalty(policy.CongestionPenalty, stats.CongestionBps))
		}
		if stats.FailureCount > 0 {
			penalty, err := parseNonNegativeInt("payments route failure penalty", policy.FailurePenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty.Mul(sdkmath.NewInt(int64(stats.FailureCount))))
		}
		if stats.SuccessRateBps > 0 && stats.SuccessRateBps < 10_000 {
			cost = cost.Add(routeScaledPenalty(policy.SuccessPenalty, 10_000-stats.SuccessRateBps))
		}
		if stats.NodeAvailabilityBps > 0 && stats.NodeAvailabilityBps < 10_000 {
			cost = cost.Add(routeScaledPenalty(policy.AvailabilityPenalty, 10_000-stats.NodeAvailabilityBps))
		}
		if stats.LiquidityUpdatedHeight > 0 && currentHeight > stats.LiquidityUpdatedHeight+policy.StaleLiquidityAfter {
			penalty, err := parseNonNegativeInt("payments stale liquidity penalty", policy.StaleLiquidityPenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty)
		}
		if stats.TimeoutMargin > 0 && stats.TimeoutMargin < policy.RequiredTimeoutMargin {
			penalty, err := parseNonNegativeInt("payments timeout margin penalty", policy.TimeoutPenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty)
		}
		if stats.ReservePressureBps > 0 {
			cost = cost.Add(routeScaledPenalty(policy.ReservePressurePenalty, stats.ReservePressureBps))
		}
		if stats.PendingConditionCount > 0 {
			penalty, err := parseNonNegativeInt("payments pending condition penalty", policy.PendingConditionPenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty.Mul(sdkmath.NewInt(int64(stats.PendingConditionCount))))
		}
		if stats.AvgResolutionLatency > 0 {
			penalty, err := parseNonNegativeInt("payments condition latency penalty", policy.LatencyPenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty.Mul(sdkmath.NewInt(int64(stats.AvgResolutionLatency))).Quo(sdkmath.NewInt(100)))
		}
		if stats.RetryCount > 0 {
			penalty, err := parseNonNegativeInt("payments route retry penalty", policy.FailurePenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty.Mul(sdkmath.NewInt(int64(stats.RetryCount))))
		}
		if stats.NodeQueueDelay > 0 {
			penalty, err := parseNonNegativeInt("payments node queue delay penalty", policy.QueueDelayPenalty)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
			}
			cost = cost.Add(penalty.Mul(sdkmath.NewInt(int64(stats.NodeQueueDelay))).Quo(sdkmath.NewInt(100)))
		}
	}
	reputation := RoutingScoreForEdge(store, edge)
	if reputation < 0 {
		cost = cost.Add(sdkmath.NewInt(-reputation))
	}
	return cost, fee, nil
}

func routeScaledPenalty(penaltyText string, multiplier uint32) sdkmath.Int {
	penalty, err := parseNonNegativeInt("payments route scaled penalty", penaltyText)
	if err != nil {
		return sdkmath.ZeroInt()
	}
	return penalty.Mul(sdkmath.NewInt(int64(multiplier))).Quo(sdkmath.NewInt(10_000))
}

func routePolicyExcludesEdge(policy RoutePolicy, edge ChannelEdge) bool {
	policy = policy.Normalize()
	edge = edge.Normalize()
	if containsString(policy.ExcludedChannels, edge.ChannelID) {
		return true
	}
	return containsString(policy.ExcludedNodes, edge.From) || containsString(policy.ExcludedNodes, edge.To)
}

func routeStatsForEdge(policy RoutePolicy, edge ChannelEdge) (EdgeRoutingStats, bool) {
	key := routeStatsKey(EdgeRoutingStats{ChannelID: edge.ChannelID, From: edge.From, To: edge.To})
	for _, stats := range policy.Normalize().EdgeStats {
		if routeStatsKey(stats) == key {
			return stats.Normalize(), true
		}
	}
	return EdgeRoutingStats{}, false
}

func edgeEffectiveCapacityCovers(policy RoutePolicy, edge ChannelEdge, amount sdkmath.Int) bool {
	edge = edge.Normalize()
	capacity, err := parsePositiveInt("payments congested edge capacity", edge.Capacity)
	if err != nil {
		return false
	}
	stats, found := routeStatsForEdge(policy, edge)
	if !found {
		return !capacity.LT(amount)
	}
	reductionBps := uint32Max(stats.CongestionBps, stats.ReservePressureBps)
	if stats.PendingConditionCount > 0 {
		reductionBps = uint32Max(reductionBps, uint32Min(10_000, stats.PendingConditionCount*500))
	}
	if reductionBps == 0 {
		return !capacity.LT(amount)
	}
	allowedBps := uint32(10_000 - reductionBps)
	maxCongestedBps := policy.Normalize().MaxCongestedPaymentBps
	if maxCongestedBps > 0 && allowedBps > maxCongestedBps {
		allowedBps = maxCongestedBps
	}
	effective := capacity.Mul(sdkmath.NewInt(int64(allowedBps))).Quo(sdkmath.NewInt(10_000))
	return !effective.LT(amount)
}

func routeStatsKey(stats EdgeRoutingStats) string {
	stats = stats.Normalize()
	return fmt.Sprintf("%s/%s/%s", stats.ChannelID, stats.From, stats.To)
}

func upsertRouteStats(stats []EdgeRoutingStats, next EdgeRoutingStats) []EdgeRoutingStats {
	next = next.Normalize()
	key := routeStatsKey(next)
	out := make([]EdgeRoutingStats, 0, len(stats)+1)
	replaced := false
	for _, existing := range stats {
		existing = existing.Normalize()
		if routeStatsKey(existing) == key {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, next)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return routeStatsKey(out[i]) < routeStatsKey(out[j])
	})
	return out
}

func decayEdgeRoutingStats(stats EdgeRoutingStats, currentHeight, halfLife uint64) EdgeRoutingStats {
	stats = stats.Normalize()
	if halfLife == 0 || stats.LastUpdatedHeight == 0 || currentHeight <= stats.LastUpdatedHeight {
		return stats
	}
	periods := (currentHeight - stats.LastUpdatedHeight) / halfLife
	if periods == 0 {
		return stats
	}
	for ; periods > 0; periods-- {
		stats.CongestionBps /= 2
		stats.FailureCount /= 2
		stats.PendingConditionCount /= 2
		stats.AvgResolutionLatency /= 2
		stats.RetryCount /= 2
		stats.ReservePressureBps /= 2
		stats.NodeQueueDelay /= 2
		if stats.NodeAvailabilityBps < 10_000 {
			stats.NodeAvailabilityBps += (10_000 - stats.NodeAvailabilityBps) / 2
		}
		if stats.SuccessRateBps < 10_000 {
			stats.SuccessRateBps += (10_000 - stats.SuccessRateBps) / 2
		}
	}
	stats.LastUpdatedHeight = currentHeight
	return stats.Normalize()
}

func routeFailureKey(report RouteFailureReport) string {
	report = report.Normalize()
	return fmt.Sprintf("%s/%s/%s/%s/%020d", report.ChannelID, report.From, report.To, report.FailureClass, report.ObservedHeight)
}

func routePolicyHash(policy RoutePolicy) string {
	policy = policy.Normalize()
	parts := []string{"route-policy", fmt.Sprintf("%d", policy.MaxHops), fmt.Sprintf("%020d", policy.RequiredTimeoutMargin), fmt.Sprintf("%020d", policy.StaleLiquidityAfter)}
	for _, channelID := range policy.ExcludedChannels {
		parts = append(parts, "excluded-channel", channelID)
	}
	for _, node := range policy.ExcludedNodes {
		parts = append(parts, "excluded-node", node)
	}
	for _, stats := range policy.EdgeStats {
		stats = stats.Normalize()
		parts = append(parts,
			"stats",
			routeStatsKey(stats),
			fmt.Sprintf("%d", stats.SuccessRateBps),
			fmt.Sprintf("%d", stats.CongestionBps),
			fmt.Sprintf("%d", stats.FailureCount),
			fmt.Sprintf("%d", stats.PendingConditionCount),
			fmt.Sprintf("%020d", stats.LastUpdatedHeight),
		)
	}
	return HashParts(parts...)
}

func uint32Max(left, right uint32) uint32 {
	if left > right {
		return left
	}
	return right
}

func uint32Min(left, right uint32) uint32 {
	if left < right {
		return left
	}
	return right
}

func routeContainsNode(edges []ChannelEdge, node string) bool {
	node = strings.TrimSpace(node)
	for _, edge := range edges {
		edge = edge.Normalize()
		if edge.From == node || edge.To == node {
			return true
		}
	}
	return false
}

func sortRouteQueue(queue []routeSearchPath) {
	sort.SliceStable(queue, func(i, j int) bool {
		if !queue[i].cost.Equal(queue[j].cost) {
			return queue[i].cost.LT(queue[j].cost)
		}
		return routePathKey(queue[i].edges) < routePathKey(queue[j].edges)
	})
}

func routePathKey(edges []ChannelEdge) string {
	if len(edges) == 0 {
		return ""
	}
	parts := make([]string, 0, len(edges))
	for _, edge := range edges {
		parts = append(parts, edgeKey(edge.Normalize()))
	}
	return strings.Join(parts, "|")
}

func buildScoredRoute(edges []ChannelEdge, amount, totalFee, totalCost sdkmath.Int) (ScoredRoute, error) {
	if len(edges) == 0 {
		return ScoredRoute{}, errors.New("payments scored route requires edges")
	}
	minCapacity, err := parsePositiveInt("payments scored route capacity", edges[0].Capacity)
	if err != nil {
		return ScoredRoute{}, err
	}
	parts := []string{"scored-route", amount.String(), totalFee.String(), totalCost.String()}
	for _, edge := range edges {
		edge = edge.Normalize()
		capacity, err := parsePositiveInt("payments scored route capacity", edge.Capacity)
		if err != nil {
			return ScoredRoute{}, err
		}
		if capacity.LT(minCapacity) {
			minCapacity = capacity
		}
		parts = append(parts, edgeKey(edge), edge.Capacity, edge.FeeAmount, fmt.Sprintf("%020d", edge.ExpiresHeight))
	}
	route := ScoredRoute{
		Edges:		append([]ChannelEdge(nil), edges...),
		Amount:		amount.String(),
		TotalFee:	totalFee.String(),
		TotalCost:	totalCost.String(),
		MinCapacity:	minCapacity.String(),
	}
	route.ScoreHash = HashParts(parts...)
	route = route.Normalize()
	return route, route.Validate()
}

func buildMultiPathRoute(parts []ScoredRoute) (MultiPathRoute, error) {
	if len(parts) == 0 {
		return MultiPathRoute{}, errors.New("payments multipath route requires parts")
	}
	totalAmount := sdkmath.ZeroInt()
	totalFee := sdkmath.ZeroInt()
	hashParts := []string{"multipath-route"}
	for _, part := range parts {
		part = part.Normalize()
		if err := part.Validate(); err != nil {
			return MultiPathRoute{}, err
		}
		amount, err := parsePositiveInt("payments multipath part amount", part.Amount)
		if err != nil {
			return MultiPathRoute{}, err
		}
		fee, err := parseNonNegativeInt("payments multipath part fee", part.TotalFee)
		if err != nil {
			return MultiPathRoute{}, err
		}
		totalAmount = totalAmount.Add(amount)
		totalFee = totalFee.Add(fee)
		hashParts = append(hashParts, part.ScoreHash)
	}
	out := MultiPathRoute{
		Parts:		append([]ScoredRoute(nil), parts...),
		TotalAmount:	totalAmount.String(),
		TotalFee:	totalFee.String(),
		ScoreHash:	HashParts(hashParts...),
	}
	return out, nil
}

func gossipPenaltyNode(envelope SignedGossipEnvelope) string {
	envelope = envelope.Normalize()
	if envelope.Message.NodeID != "" {
		return envelope.Message.NodeID
	}
	if envelope.Signature.Signer != "" {
		return envelope.Signature.Signer
	}
	return envelope.ReceivedFrom
}

func upsertGossipEnvelope(envelopes []SignedGossipEnvelope, next SignedGossipEnvelope) []SignedGossipEnvelope {
	next = next.Normalize()
	messageID := next.Message.MessageID
	out := make([]SignedGossipEnvelope, 0, len(envelopes)+1)
	replaced := false
	for _, envelope := range envelopes {
		envelope = envelope.Normalize()
		if envelope.Message.MessageID == messageID {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, envelope)
	}
	if !replaced {
		out = append(out, next)
	}
	sortGossipEnvelopes(out)
	return out
}

func upsertTopologyEdge(edges []ChannelEdge, next ChannelEdge) []ChannelEdge {
	next = next.Normalize()
	nextKey := edgeKey(next)
	out := make([]ChannelEdge, 0, len(edges)+1)
	replaced := false
	for _, edge := range edges {
		edge = edge.Normalize()
		if edgeKey(edge) == nextKey {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, edge)
	}
	if !replaced {
		out = append(out, next)
	}
	sortEdges(out)
	return out
}

func addGossipReputation(reputation []GossipReputation, nodeID string, delta int64, invalid bool, height uint64) []GossipReputation {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" || height == 0 {
		return normalizeGossipReputation(reputation)
	}
	out := make([]GossipReputation, 0, len(reputation)+1)
	replaced := false
	for _, record := range reputation {
		record = record.Normalize()
		if record.NodeID == nodeID {
			record.Score += delta
			record.LastUpdateHeight = height
			if invalid {
				record.InvalidGossip++
			}
			out = append(out, record)
			replaced = true
			continue
		}
		out = append(out, record)
	}
	if !replaced {
		record := GossipReputation{NodeID: nodeID, Score: delta, LastUpdateHeight: height}
		if invalid {
			record.InvalidGossip = 1
		}
		out = append(out, record.Normalize())
	}
	return normalizeGossipReputation(out)
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

func filterCustodyLocksForSettledChannel(locks []CustodyLock, channelID string) []CustodyLock {
	channelID = normalizeHash(channelID)
	out := make([]CustodyLock, 0, len(locks))
	for _, lock := range locks {
		if lock.Normalize().ChannelID == channelID {
			continue
		}
		out = append(out, lock)
	}
	return out
}

func stateStrongerThan(candidate, current ChannelState) bool {
	candidate = candidate.Normalize()
	current = current.Normalize()
	if candidate.Nonce > current.Nonce {
		return true
	}
	return candidate.ChannelType == ChannelTypeAsync && candidate.CheckpointNonce > current.CheckpointNonce
}

func mergeConditionResolutions(left, right []ConditionResolution) []ConditionResolution {
	byID := make(map[string]ConditionResolution, len(left)+len(right))
	for _, resolution := range normalizeConditionResolutions(left) {
		byID[resolution.ConditionID] = resolution
	}
	for _, resolution := range normalizeConditionResolutions(right) {
		byID[resolution.ConditionID] = resolution
	}
	out := make([]ConditionResolution, 0, len(byID))
	for _, resolution := range byID {
		out = append(out, resolution)
	}
	return normalizeConditionResolutions(out)
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

func sortGossipEnvelopes(envelopes []SignedGossipEnvelope) {
	sort.SliceStable(envelopes, func(i, j int) bool {
		left := envelopes[i].Normalize()
		right := envelopes[j].Normalize()
		if left.Message.MessageID == right.Message.MessageID {
			return left.ReceivedAt < right.ReceivedAt
		}
		return left.Message.MessageID < right.Message.MessageID
	})
}

func sortGossipReputation(reputation []GossipReputation) {
	sort.SliceStable(reputation, func(i, j int) bool {
		return reputation[i].Normalize().NodeID < reputation[j].Normalize().NodeID
	})
}

func sortForwardingReplayRecords(records []ForwardingPacketReplayRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i].Normalize()
		right := records[j].Normalize()
		if left.RouteID == right.RouteID {
			return left.PacketID < right.PacketID
		}
		return left.RouteID < right.RouteID
	})
}

func normalizeForwardingReplayRecords(records []ForwardingPacketReplayRecord) []ForwardingPacketReplayRecord {
	out := make([]ForwardingPacketReplayRecord, len(records))
	for i, record := range records {
		out[i] = record.Normalize()
	}
	sortForwardingReplayRecords(out)
	return out
}

func normalizeGossipReputation(reputation []GossipReputation) []GossipReputation {
	out := make([]GossipReputation, len(reputation))
	for i, record := range reputation {
		out[i] = record.Normalize()
	}
	sortGossipReputation(out)
	return out
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

func sortCustodyLocks(locks []CustodyLock) {
	sort.SliceStable(locks, func(i, j int) bool {
		return locks[i].Normalize().ChannelID < locks[j].Normalize().ChannelID
	})
}

func sortClosedChannelTombstones(tombstones []ClosedChannelTombstone) {
	sort.SliceStable(tombstones, func(i, j int) bool {
		return tombstones[i].Normalize().ChannelID < tombstones[j].Normalize().ChannelID
	})
}

func sortConditionClaimRecords(claims []ConditionClaimRecord) {
	sort.SliceStable(claims, func(i, j int) bool {
		left := claims[i].Normalize()
		right := claims[j].Normalize()
		return conditionClaimKey(left.ChannelID, left.ConditionID)+"/"+left.EvidenceHash < conditionClaimKey(right.ChannelID, right.ConditionID)+"/"+right.EvidenceHash
	})
}

func sortValidatorPaymentServices(services []ValidatorPaymentServiceMetadata) {
	sort.SliceStable(services, func(i, j int) bool {
		return services[i].Normalize().ValidatorAddress < services[j].Normalize().ValidatorAddress
	})
}

func sortValidatorWatchRegistrations(registrations []ValidatorWatchRegistration) {
	sort.SliceStable(registrations, func(i, j int) bool {
		left := registrations[i].Normalize()
		right := registrations[j].Normalize()
		return validatorWatchRegistrationKey(left.ValidatorAddress, left.Delegator) < validatorWatchRegistrationKey(right.ValidatorAddress, right.Delegator)
	})
}

func sortPaymentFeeMultipliers(multipliers []PaymentFeeMultiplier) {
	sort.SliceStable(multipliers, func(i, j int) bool {
		return string(multipliers[i].Normalize().FeeClass) < string(multipliers[j].Normalize().FeeClass)
	})
}

func sortPaymentFeeCharges(charges []PaymentFeeCharge) {
	sort.SliceStable(charges, func(i, j int) bool {
		return charges[i].Normalize().FeeID < charges[j].Normalize().FeeID
	})
}

func sortPaymentFeeRefunds(refunds []PaymentFeeRefund) {
	sort.SliceStable(refunds, func(i, j int) bool {
		return refunds[i].Normalize().RefundID < refunds[j].Normalize().RefundID
	})
}

func sortSecurityReserveAllocationHooks(hooks []SecurityReserveAllocationHook) {
	sort.SliceStable(hooks, func(i, j int) bool {
		return hooks[i].Normalize().HookID < hooks[j].Normalize().HookID
	})
}

func sortSettlementInclusionLatencies(records []SettlementInclusionLatency) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Normalize().RecordID < records[j].Normalize().RecordID
	})
}

func sortAsyncFinalizationJobs(jobs []AsyncFinalizationJob) {
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Normalize().JobID < jobs[j].Normalize().JobID
	})
}

func sortAsyncPromiseExpiryJobs(jobs []AsyncPromiseExpiryJob) {
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Normalize().JobID < jobs[j].Normalize().JobID
	})
}

func sortAsyncSettlementCompletions(completions []AsyncSettlementCompletion) {
	sort.SliceStable(completions, func(i, j int) bool {
		return completions[i].Normalize().CompletionID < completions[j].Normalize().CompletionID
	})
}

func conditionClaimKey(channelID, conditionID string) string {
	return normalizeHash(channelID) + "/" + normalizeHash(conditionID)
}

func conditionEvidenceKey(channelID, evidenceHash string) string {
	return normalizeHash(channelID) + "/" + normalizeHash(evidenceHash)
}

func validatorWatchRegistrationKey(validatorAddress, delegator string) string {
	return strings.TrimSpace(validatorAddress) + "/" + strings.TrimSpace(delegator)
}

func paymentFeeBaseAmount(schedule PaymentFeeSchedule, feeClass PaymentFeeClass) (string, error) {
	switch feeClass {
	case PaymentFeeClassChannelOpen:
		return schedule.ChannelOpenFee, nil
	case PaymentFeeClassChannelCheckpoint:
		return schedule.ChannelCheckpointFee, nil
	case PaymentFeeClassCooperativeClose:
		return schedule.CooperativeCloseFee, nil
	case PaymentFeeClassUnilateralClose:
		return schedule.UnilateralCloseFee, nil
	case PaymentFeeClassDispute:
		return schedule.DisputeFee, nil
	case PaymentFeeClassFraudProofVerification:
		return schedule.FraudProofVerificationFee, nil
	case PaymentFeeClassConditionalPromiseSettlement:
		return schedule.ConditionalPromiseSettlementFee, nil
	case PaymentFeeClassVirtualChannelAnchor:
		return schedule.VirtualChannelAnchorFee, nil
	case PaymentFeeClassRoutingAdvertisement:
		return schedule.RoutingAdvertisementFee, nil
	default:
		return "", fmt.Errorf("unknown payments fee class %q", feeClass)
	}
}

func paymentStorageFootprint(feeClass PaymentFeeClass, channel ChannelRecord) uint64 {
	channel = channel.Normalize()
	switch feeClass {
	case PaymentFeeClassChannelOpen, PaymentFeeClassChannelCheckpoint, PaymentFeeClassUnilateralClose, PaymentFeeClassDispute:
		return EstimateChannelOpenStorageFootprint(channel)
	case PaymentFeeClassConditionalPromiseSettlement:
		return uint64(len(channel.ChannelID) + len(channel.LatestState.Conditions)*128)
	case PaymentFeeClassVirtualChannelAnchor:
		return uint64(len(channel.ChannelID) + len(channel.Participants)*48 + 128)
	case PaymentFeeClassRoutingAdvertisement:
		return uint64(len(channel.ChannelID) + len(channel.Participants)*48 + 64)
	default:
		return 0
	}
}

func EstimateChannelOpenStorageFootprint(channel ChannelRecord) uint64 {
	channel = channel.Normalize()
	footprint := uint64(128)
	footprint += uint64(len(channel.ChannelID) + len(channel.ChainID) + len(channel.Denom) + len(channel.Collateral))
	footprint += uint64(len(channel.Participants) * 48)
	footprint += uint64(len(channel.RequiredSigners) * 48)
	footprint += uint64(len(channel.OpeningStateHash) + len(channel.LatestState.StateHash) + len(channel.LatestState.ParticipantSetHash))
	footprint += uint64(len(channel.LatestState.Balances) * 64)
	footprint += uint64(len(channel.LatestState.ReserveA) + len(channel.LatestState.ReserveB))
	footprint += uint64(len(channel.LatestState.Conditions) * 160)
	if channel.ConditionalPayments {
		footprint += 64
	}
	if channel.RoutingAdvertised {
		footprint += 96
	}
	return footprint
}

func feeMultiplierForClass(state PaymentsState, feeClass PaymentFeeClass, schedule PaymentFeeSchedule) uint32 {
	multiplier := schedule.Normalize().BaseMultiplierBps
	for _, configured := range state.FeeMultipliers {
		configured = configured.Normalize()
		if configured.FeeClass == feeClass {
			multiplier = configured.MultiplierBps
			break
		}
	}
	return multiplier
}

func feeChannelForVirtual(vc VirtualChannel) ChannelRecord {
	vc = vc.Normalize()
	return ChannelRecord{
		ChainID:	vc.ChainID,
		ChannelID:	vc.VirtualChannelID,
		Participants:	vc.Endpoints,
		LatestState:	ChannelState{StateHash: vc.StateHash},
	}
}

func asyncFinalizationJobID(channelID string, finalizeHeight uint64) string {
	return HashParts("async-finalization-job", normalizeHash(channelID), fmt.Sprintf("%020d", finalizeHeight))
}

func asyncPromiseExpiryJobID(channelID, promiseID string, expireAfterHeight uint64) string {
	return HashParts("async-promise-expiry-job", normalizeHash(channelID), normalizeHash(promiseID), fmt.Sprintf("%020d", expireAfterHeight))
}

func asyncFinalizationJobByID(jobs []AsyncFinalizationJob, jobID string) (AsyncFinalizationJob, bool) {
	jobID = normalizeHash(jobID)
	for _, job := range jobs {
		job = job.Normalize()
		if job.JobID == jobID {
			return job, true
		}
	}
	return AsyncFinalizationJob{}, false
}

func asyncPromiseExpiryJobByID(jobs []AsyncPromiseExpiryJob, jobID string) (AsyncPromiseExpiryJob, bool) {
	jobID = normalizeHash(jobID)
	for _, job := range jobs {
		job = job.Normalize()
		if job.JobID == jobID {
			return job, true
		}
	}
	return AsyncPromiseExpiryJob{}, false
}

func markAsyncFinalizationCompleted(state PaymentsState, jobID, settlementHash string, height uint64) PaymentsState {
	jobID = normalizeHash(jobID)
	for i := range state.AsyncFinalizationQueue {
		if state.AsyncFinalizationQueue[i].Normalize().JobID == jobID {
			state.AsyncFinalizationQueue[i].Completed = true
			state.AsyncFinalizationQueue[i].CompletedHeight = height
			state.AsyncFinalizationQueue[i].SettlementHash = normalizeHash(settlementHash)
			state.AsyncFinalizationQueue[i].LastRunHeight = height
			state.AsyncFinalizationQueue[i].LastError = ""
			state.AsyncFinalizationQueue[i].Attempts++
			break
		}
	}
	sortAsyncFinalizationJobs(state.AsyncFinalizationQueue)
	return state
}

func markAsyncFinalizationFailed(state PaymentsState, jobID string, height uint64, message string) PaymentsState {
	jobID = normalizeHash(jobID)
	for i := range state.AsyncFinalizationQueue {
		if state.AsyncFinalizationQueue[i].Normalize().JobID == jobID {
			state.AsyncFinalizationQueue[i].LastRunHeight = height
			state.AsyncFinalizationQueue[i].LastError = strings.TrimSpace(message)
			state.AsyncFinalizationQueue[i].Attempts++
			break
		}
	}
	sortAsyncFinalizationJobs(state.AsyncFinalizationQueue)
	return state
}

func markAsyncPromiseExpiryCompleted(state PaymentsState, jobID, resolutionHash string, height uint64) PaymentsState {
	jobID = normalizeHash(jobID)
	for i := range state.AsyncPromiseExpiryQueue {
		if state.AsyncPromiseExpiryQueue[i].Normalize().JobID == jobID {
			state.AsyncPromiseExpiryQueue[i].Completed = true
			state.AsyncPromiseExpiryQueue[i].CompletedHeight = height
			state.AsyncPromiseExpiryQueue[i].ResolutionHash = normalizeHash(resolutionHash)
			state.AsyncPromiseExpiryQueue[i].LastRunHeight = height
			state.AsyncPromiseExpiryQueue[i].LastError = ""
			state.AsyncPromiseExpiryQueue[i].Attempts++
			break
		}
	}
	sortAsyncPromiseExpiryJobs(state.AsyncPromiseExpiryQueue)
	return state
}

func markAsyncPromiseExpiryFailed(state PaymentsState, jobID string, height uint64, message string) PaymentsState {
	jobID = normalizeHash(jobID)
	for i := range state.AsyncPromiseExpiryQueue {
		if state.AsyncPromiseExpiryQueue[i].Normalize().JobID == jobID {
			state.AsyncPromiseExpiryQueue[i].LastRunHeight = height
			state.AsyncPromiseExpiryQueue[i].LastError = strings.TrimSpace(message)
			state.AsyncPromiseExpiryQueue[i].Attempts++
			break
		}
	}
	sortAsyncPromiseExpiryJobs(state.AsyncPromiseExpiryQueue)
	return state
}

func appendAsyncCompletion(state PaymentsState, jobID, jobType, channelID, objectID, resultHash string, height uint64, result *AsyncExecutionResult) PaymentsState {
	jobID = normalizeHash(jobID)
	resultHash = normalizeHash(resultHash)
	for _, completion := range state.AsyncCompletions {
		completion = completion.Normalize()
		if completion.JobID == jobID && completion.ResultHash == resultHash {
			return state
		}
	}
	completion := AsyncSettlementCompletion{
		CompletionID:	HashParts("async-settlement-completion", jobID, resultHash, fmt.Sprintf("%020d", height)),
		JobID:		jobID,
		JobType:	jobType,
		ChannelID:	channelID,
		ObjectID:	objectID,
		ResultHash:	resultHash,
		Height:		height,
	}.Normalize()
	state.AsyncCompletions = append(state.AsyncCompletions, completion)
	state.Events = append(state.Events, AsyncSettlementCompletionEvent(completion))
	sortAsyncSettlementCompletions(state.AsyncCompletions)
	if result != nil {
		result.EmittedCompletionIDs = append(result.EmittedCompletionIDs, completion.CompletionID)
	}
	return state
}

func latestSettlementHashForChannel(settlements []SettlementRecord, channelID string) string {
	channelID = normalizeHash(channelID)
	var latest SettlementRecord
	for _, settlement := range settlements {
		settlement = settlement.Normalize()
		if settlement.ChannelID != channelID {
			continue
		}
		if settlement.SettledHeight >= latest.SettledHeight {
			latest = settlement
		}
	}
	return latest.SettlementHash
}

func edgeKey(edge ChannelEdge) string {
	return fmt.Sprintf("%s/%s/%s", edge.ChannelID, edge.From, edge.To)
}
