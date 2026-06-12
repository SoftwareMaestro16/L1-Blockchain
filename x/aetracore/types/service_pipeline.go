package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ServicePipelinePhase string
type ServiceCallKind string
type ServiceCallStatus string
type ServicePaymentStatus string

const (
	ServicePhasePrepareProposal	ServicePipelinePhase	= "PREPARE_PROPOSAL"
	ServicePhaseProcessProposal	ServicePipelinePhase	= "PROCESS_PROPOSAL"
	ServicePhaseDeliverTxCompat	ServicePipelinePhase	= "DELIVER_TX_COMPAT"
	ServicePhaseFinalizeBlock	ServicePipelinePhase	= "FINALIZE_BLOCK"
	ServicePhaseEndBlock		ServicePipelinePhase	= "END_BLOCK"

	ServiceCallKindOnChain		ServiceCallKind	= "ON_CHAIN_CALL"
	ServiceCallKindOffChainReceipt	ServiceCallKind	= "OFF_CHAIN_RECEIPT"
	ServiceCallKindMixedDispute	ServiceCallKind	= "MIXED_DISPUTE"
	ServiceCallKindMixedSettlement	ServiceCallKind	= "MIXED_SETTLEMENT"
	ServiceCallKindCallback		ServiceCallKind	= "CALLBACK"
	ServiceCallKindRetry		ServiceCallKind	= "RETRY"

	ServiceCallStatusAccepted	ServiceCallStatus	= "ACCEPTED"
	ServiceCallStatusExecuted	ServiceCallStatus	= "EXECUTED"
	ServiceCallStatusFailed		ServiceCallStatus	= "FAILED"
	ServiceCallStatusExpired	ServiceCallStatus	= "EXPIRED"
	ServiceCallStatusChallenged	ServiceCallStatus	= "CHALLENGED"
	ServiceCallStatusSettled	ServiceCallStatus	= "SETTLED"
	ServiceCallStatusReverted	ServiceCallStatus	= "REVERTED"

	ServicePaymentStatusNone	ServicePaymentStatus	= "NONE"
	ServicePaymentStatusReserved	ServicePaymentStatus	= "RESERVED"
	ServicePaymentStatusSettled	ServicePaymentStatus	= "SETTLED"
	ServicePaymentStatusRefunded	ServicePaymentStatus	= "REFUNDED"
	ServicePaymentStatusEscrowed	ServicePaymentStatus	= "ESCROWED"
)

type ServiceConsensusContext struct {
	ChainID	string
	Height	uint64
}

type ServiceCallEnvelope struct {
	CallID			string
	ServiceID		string
	Caller			string
	Nonce			uint64
	IdempotencyKey		string
	MethodID		string
	InterfaceHash		string
	PayloadHash		string
	PaymentDenom		string
	MaxFeeAmount		string
	ProofRequirement	ServiceVerificationModel
	Kind			ServiceCallKind
	CreatedHeight		uint64
	DeadlineHeight		uint64
	PriorityClass		uint32
	Callback		bool
	RetryOf			string
	Dispute			bool
	StateReadSet		[]string
	StateWriteSet		[]string
}

type ServiceExecutionGroup struct {
	GroupID		string
	StateKeyScope	string
	Calls		[]ServiceCallEnvelope
}

type ServiceProposalPlan struct {
	Height		uint64
	ChainID		string
	Groups		[]ServiceExecutionGroup
	RegistryRoot	string
	InterfaceRoot	string
	PlanHash	string
}

type ServiceExecutionOutcome struct {
	CallID		string
	Status		ServiceCallStatus
	ResponseHash	string
	ProofHash	string
	PaymentStatus	ServicePaymentStatus
	GasUsed		uint64
	ProviderID	string
	ExecutedHeight	uint64
	AnchoredHeight	uint64
	ErrorCode	string
}

type ServiceCallReceipt struct {
	CallID		string
	ServiceID	string
	MethodID	string
	Caller		string
	Status		ServiceCallStatus
	RequestHash	string
	ResponseHash	string
	ProofHash	string
	PaymentStatus	ServicePaymentStatus
	GasUsed		uint64
	ProviderID	string
	ExecutedHeight	uint64
	AnchoredHeight	uint64
	ErrorCode	string
	ReceiptHash	string
}

type ServiceFinalization struct {
	Height			uint64
	RegistryRoot		string
	InterfaceRoot		string
	ServiceReceiptsRoot	string
	Receipts		[]ServiceCallReceipt
	FinalizationHash	string
}

type ProviderReputationDelta struct {
	ProviderID	string
	Successes	uint64
	Failures	uint64
}

type ServicePipelineMetrics struct {
	ActiveServices		uint64
	ExpiredServices		uint64
	ReceiptCount		uint64
	FailedCallCount		uint64
	SettledPaymentCount	uint64
	ProviderReceiptCount	uint64
}

type ServiceEndBlockMaintenance struct {
	Height			uint64
	ExpiredServiceIDs	[]string
	CleanupServiceIDs	[]string
	ReputationDeltas	[]ProviderReputationDelta
	Metrics			ServicePipelineMetrics
	MaintenanceHash		string
}

type ServiceStateTransition struct {
	CurrentStateRoot	string
	NextStateRoot		string
	Call			ServiceCallEnvelope
	Context			ServiceConsensusContext
	StateReadSet		[]string
	StateWriteSet		[]string
	ExternalCalls		[]string
	UsesWallClock		bool
	LiveAvailabilityChecks	bool
	IterationLimit		uint64
	ProofVerificationGas	uint64
	DirectCrossZoneWrites	[]string
}

func PrepareServiceProposal(ctx ServiceConsensusContext, state CoreState, calls []ServiceCallEnvelope) (ServiceProposalPlan, error) {
	if err := ctx.Validate(); err != nil {
		return ServiceProposalPlan{}, err
	}
	if err := state.Validate(); err != nil {
		return ServiceProposalPlan{}, err
	}
	ordered := make([]ServiceCallEnvelope, len(calls))
	for i, call := range calls {
		call = NormalizeServiceCall(ctx, call)
		if err := ValidateServiceCallForState(ctx, state, call); err != nil {
			return ServiceProposalPlan{}, err
		}
		ordered[i] = call
	}
	sortServiceCalls(ordered)

	groupsByScope := make(map[string]int)
	groups := make([]ServiceExecutionGroup, 0)
	for _, call := range ordered {
		scope := ComputeServiceCallStateScope(call)
		groupIndex, found := groupsByScope[scope]
		if !found {
			groupIndex = len(groups)
			groupsByScope[scope] = groupIndex
			groups = append(groups, ServiceExecutionGroup{
				GroupID:	ComputeServiceExecutionGroupID(scope),
				StateKeyScope:	scope,
			})
		}
		groups[groupIndex].Calls = append(groups[groupIndex].Calls, call)
	}
	sortServiceExecutionGroups(groups)
	registryRoot, err := ComputeServiceRoot(state.ServiceDescriptors)
	if err != nil {
		return ServiceProposalPlan{}, err
	}
	interfaceRoot, err := ComputeServiceInterfaceRoot(state.ServiceDescriptors)
	if err != nil {
		return ServiceProposalPlan{}, err
	}
	plan := ServiceProposalPlan{
		Height:		ctx.Height,
		ChainID:	ctx.ChainID,
		Groups:		groups,
		RegistryRoot:	registryRoot,
		InterfaceRoot:	interfaceRoot,
	}
	plan.PlanHash = ComputeServiceProposalPlanHash(plan)
	return plan, ValidateServiceProposalPlan(ctx, state, plan)
}

func ProcessServiceProposal(ctx ServiceConsensusContext, state CoreState, plan ServiceProposalPlan) error {
	return ValidateServiceProposalPlan(ctx, state, plan)
}

func FinalizeServiceProposal(ctx ServiceConsensusContext, state CoreState, plan ServiceProposalPlan, outcomes []ServiceExecutionOutcome) (ServiceFinalization, error) {
	if err := ValidateServiceProposalPlan(ctx, state, plan); err != nil {
		return ServiceFinalization{}, err
	}
	outcomeByCall := make(map[string]ServiceExecutionOutcome, len(outcomes))
	for _, outcome := range outcomes {
		outcome = NormalizeServiceExecutionOutcome(ctx, outcome)
		if _, found := outcomeByCall[outcome.CallID]; found {
			return ServiceFinalization{}, fmt.Errorf("duplicate aetracore service outcome %s", outcome.CallID)
		}
		outcomeByCall[outcome.CallID] = outcome
	}
	receipts := make([]ServiceCallReceipt, 0, len(outcomes))
	for _, call := range plan.OrderedCalls() {
		outcome, found := outcomeByCall[call.CallID]
		if !found {
			return ServiceFinalization{}, fmt.Errorf("missing aetracore service outcome %s", call.CallID)
		}
		if err := ValidateServiceExecutionOutcome(ctx, call, outcome); err != nil {
			return ServiceFinalization{}, err
		}
		receipt, err := NewServiceCallReceipt(call, outcome)
		if err != nil {
			return ServiceFinalization{}, err
		}
		receipts = append(receipts, receipt)
	}
	receiptsRoot, err := ComputeServiceReceiptsRoot(receipts)
	if err != nil {
		return ServiceFinalization{}, err
	}
	finalization := ServiceFinalization{
		Height:			ctx.Height,
		RegistryRoot:		plan.RegistryRoot,
		InterfaceRoot:		plan.InterfaceRoot,
		ServiceReceiptsRoot:	receiptsRoot,
		Receipts:		cloneServiceCallReceipts(receipts),
	}
	finalization.FinalizationHash = ComputeServiceFinalizationHash(finalization)
	return finalization, finalization.Validate()
}

func EndBlockServiceMaintenance(state CoreState, height uint64, receipts []ServiceCallReceipt, limit uint64) (ServiceEndBlockMaintenance, error) {
	if height == 0 {
		return ServiceEndBlockMaintenance{}, errors.New("aetracore service endblock height must be positive")
	}
	if limit == 0 {
		return ServiceEndBlockMaintenance{}, errors.New("aetracore service endblock cleanup limit must be positive")
	}
	if err := state.Validate(); err != nil {
		return ServiceEndBlockMaintenance{}, err
	}
	exported := state.Export()
	expired := make([]string, 0)
	activeCount := uint64(0)
	for _, descriptor := range exported.ServiceDescriptors {
		if descriptor.Enabled && descriptor.Status == ServiceStatusActive && (descriptor.ExpiryHeight == 0 || descriptor.ExpiryHeight > height) {
			activeCount++
		}
		if descriptor.ExpiryHeight != 0 && descriptor.ExpiryHeight <= height {
			expired = append(expired, descriptor.ServiceID)
			if uint64(len(expired)) == limit {
				break
			}
		}
	}
	sort.Strings(expired)
	receiptCopies := cloneServiceCallReceipts(receipts)
	sortServiceCallReceipts(receiptCopies)
	for _, receipt := range receiptCopies {
		if err := receipt.Validate(); err != nil {
			return ServiceEndBlockMaintenance{}, err
		}
	}
	maintenance := ServiceEndBlockMaintenance{
		Height:			height,
		ExpiredServiceIDs:	append([]string(nil), expired...),
		CleanupServiceIDs:	append([]string(nil), expired...),
		ReputationDeltas:	BuildProviderReputationDeltas(receiptCopies),
		Metrics: ServicePipelineMetrics{
			ActiveServices:		activeCount,
			ExpiredServices:	uint64(len(expired)),
			ReceiptCount:		uint64(len(receiptCopies)),
			FailedCallCount:	countFailedServiceReceipts(receiptCopies),
			SettledPaymentCount:	countSettledServicePayments(receiptCopies),
			ProviderReceiptCount:	countProviderReceipts(receiptCopies),
		},
	}
	maintenance.MaintenanceHash = ComputeServiceEndBlockMaintenanceHash(maintenance)
	return maintenance, maintenance.Validate()
}

func (ctx ServiceConsensusContext) Validate() error {
	if err := validatePolicyID("aetracore service chain id", ctx.ChainID); err != nil {
		return err
	}
	if ctx.Height == 0 {
		return errors.New("aetracore service consensus height must be positive")
	}
	return nil
}

func NormalizeServiceCall(ctx ServiceConsensusContext, call ServiceCallEnvelope) ServiceCallEnvelope {
	call.ServiceID = strings.TrimSpace(call.ServiceID)
	call.Caller = strings.TrimSpace(call.Caller)
	call.IdempotencyKey = strings.TrimSpace(call.IdempotencyKey)
	call.MethodID = strings.TrimSpace(call.MethodID)
	call.InterfaceHash = strings.ToLower(strings.TrimSpace(call.InterfaceHash))
	call.PayloadHash = strings.ToLower(strings.TrimSpace(call.PayloadHash))
	call.PaymentDenom = strings.TrimSpace(call.PaymentDenom)
	call.MaxFeeAmount = strings.TrimSpace(call.MaxFeeAmount)
	call.RetryOf = strings.ToLower(strings.TrimSpace(call.RetryOf))
	call.StateReadSet = normalizeServiceStringSet(call.StateReadSet)
	call.StateWriteSet = normalizeServiceStringSet(call.StateWriteSet)
	if call.CallID == "" && call.ServiceID != "" && call.Caller != "" && call.PayloadHash != "" {
		call.CallID = ComputeServiceCallID(ctx, call)
	} else {
		call.CallID = strings.ToLower(strings.TrimSpace(call.CallID))
	}
	return call
}

func ValidateServiceCallForState(ctx ServiceConsensusContext, state CoreState, call ServiceCallEnvelope) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	call = NormalizeServiceCall(ctx, call)
	if err := call.ValidateBasic(ctx); err != nil {
		return err
	}
	descriptor, found := state.ServiceByID(call.ServiceID)
	if !found {
		return fmt.Errorf("aetracore service call references unknown service %s", call.ServiceID)
	}
	if !descriptor.Enabled || descriptor.Status != ServiceStatusActive {
		return fmt.Errorf("aetracore service %s is not active", call.ServiceID)
	}
	if descriptor.ExpiryHeight != 0 && call.DeadlineHeight > descriptor.ExpiryHeight {
		return errors.New("aetracore service call deadline must not exceed service expiry")
	}
	if call.InterfaceHash != descriptor.Interface.InterfaceHash {
		return errors.New("aetracore service call interface hash mismatch")
	}
	method, found := descriptor.Interface.MethodByID(call.MethodID)
	if !found {
		return fmt.Errorf("aetracore service method %s is not registered", call.MethodID)
	}
	if method.IdempotencyRequired && call.IdempotencyKey == "" {
		return errors.New("aetracore service call requires idempotency key")
	}
	if call.ProofRequirement != method.VerificationModel {
		return errors.New("aetracore service call proof requirement mismatch")
	}
	if call.PaymentDenom != descriptor.Payment.Denom {
		return errors.New("aetracore service call payment denom mismatch")
	}
	if !serviceCallKindAllowed(descriptor.ServiceType, call.Kind) {
		return fmt.Errorf("aetracore service call kind %s is incompatible with %s service", call.Kind, descriptor.ServiceType)
	}
	return nil
}

func (call ServiceCallEnvelope) ValidateBasic(ctx ServiceConsensusContext) error {
	if err := validatePolicyID("aetracore service call service id", call.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service call caller", call.Caller); err != nil {
		return err
	}
	if call.Nonce == 0 {
		return errors.New("aetracore service call nonce must be positive")
	}
	if err := validatePolicyID("aetracore service call method id", call.MethodID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service call interface hash", call.InterfaceHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service call payload hash", call.PayloadHash); err != nil {
		return err
	}
	if call.CreatedHeight == 0 || call.CreatedHeight > ctx.Height {
		return errors.New("aetracore service call created height is invalid")
	}
	if call.DeadlineHeight == 0 || call.DeadlineHeight < ctx.Height {
		return errors.New("aetracore service call is expired")
	}
	if !IsServiceVerificationModel(call.ProofRequirement) {
		return fmt.Errorf("unknown aetracore service call proof requirement %q", call.ProofRequirement)
	}
	if !IsServiceCallKind(call.Kind) {
		return fmt.Errorf("unknown aetracore service call kind %q", call.Kind)
	}
	if err := validatePolicyID("aetracore service call payment denom", call.PaymentDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore service call max fee", call.MaxFeeAmount); err != nil {
		return err
	}
	if call.RetryOf != "" {
		if err := ValidateHash("aetracore service retry call id", call.RetryOf); err != nil {
			return err
		}
	}
	if err := validateServiceStateKeys("aetracore service read set", call.StateReadSet); err != nil {
		return err
	}
	if err := validateServiceStateKeys("aetracore service write set", call.StateWriteSet); err != nil {
		return err
	}
	expected := ComputeServiceCallID(ctx, call)
	if call.CallID != expected {
		return fmt.Errorf("aetracore service call id mismatch: expected %s", expected)
	}
	return nil
}

func ValidateServiceProposalPlan(ctx ServiceConsensusContext, state CoreState, plan ServiceProposalPlan) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	if err := state.Validate(); err != nil {
		return err
	}
	if plan.Height != ctx.Height || plan.ChainID != ctx.ChainID {
		return errors.New("aetracore service proposal context mismatch")
	}
	if err := ValidateHash("aetracore service proposal registry root", plan.RegistryRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service proposal interface root", plan.InterfaceRoot); err != nil {
		return err
	}
	if expected := ComputeServiceProposalPlanHash(plan); plan.PlanHash != expected {
		return fmt.Errorf("aetracore service proposal hash mismatch: expected %s", expected)
	}
	registryRoot, err := ComputeServiceRoot(state.ServiceDescriptors)
	if err != nil {
		return err
	}
	if registryRoot != plan.RegistryRoot {
		return errors.New("aetracore service proposal registry root mismatch")
	}
	interfaceRoot, err := ComputeServiceInterfaceRoot(state.ServiceDescriptors)
	if err != nil {
		return err
	}
	if interfaceRoot != plan.InterfaceRoot {
		return errors.New("aetracore service proposal interface root mismatch")
	}
	seenCalls := make(map[string]struct{})
	senderLastNonce := make(map[string]uint64)
	writeOwners := make(map[string]string)
	for i, group := range plan.Groups {
		if err := group.Validate(ctx, state); err != nil {
			return err
		}
		if i > 0 && plan.Groups[i-1].GroupID >= group.GroupID {
			return errors.New("aetracore service proposal groups must be sorted canonically")
		}
		for _, call := range group.Calls {
			if _, found := seenCalls[call.CallID]; found {
				return fmt.Errorf("duplicate aetracore service call %s", call.CallID)
			}
			seenCalls[call.CallID] = struct{}{}
			senderKey := call.ServiceID + "/" + call.Caller
			if lastNonce := senderLastNonce[senderKey]; lastNonce >= call.Nonce {
				return errors.New("aetracore service calls from same sender must be nonce ordered")
			}
			senderLastNonce[senderKey] = call.Nonce
			for _, key := range call.StateWriteSet {
				if owner, found := writeOwners[key]; found && owner != group.GroupID {
					return errors.New("aetracore service proposal groups have BlockSTM write conflict")
				}
				writeOwners[key] = group.GroupID
			}
		}
	}
	return nil
}

func (g ServiceExecutionGroup) Validate(ctx ServiceConsensusContext, state CoreState) error {
	if err := ValidateHash("aetracore service execution group id", g.GroupID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service execution group scope", g.StateKeyScope); err != nil {
		return err
	}
	if g.GroupID != ComputeServiceExecutionGroupID(g.StateKeyScope) {
		return errors.New("aetracore service execution group id mismatch")
	}
	if len(g.Calls) == 0 {
		return errors.New("aetracore service execution group requires calls")
	}
	for i, call := range g.Calls {
		call = NormalizeServiceCall(ctx, call)
		if err := ValidateServiceCallForState(ctx, state, call); err != nil {
			return err
		}
		if ComputeServiceCallStateScope(call) != g.StateKeyScope {
			return errors.New("aetracore service execution group scope mismatch")
		}
		if i > 0 && compareServiceCalls(g.Calls[i-1], call) >= 0 {
			return errors.New("aetracore service execution group calls must be sorted canonically")
		}
	}
	return nil
}

func (p ServiceProposalPlan) OrderedCalls() []ServiceCallEnvelope {
	calls := make([]ServiceCallEnvelope, 0)
	for _, group := range p.Groups {
		calls = append(calls, cloneServiceCalls(group.Calls)...)
	}
	sortServiceCalls(calls)
	return calls
}

func NormalizeServiceExecutionOutcome(ctx ServiceConsensusContext, outcome ServiceExecutionOutcome) ServiceExecutionOutcome {
	outcome.CallID = strings.ToLower(strings.TrimSpace(outcome.CallID))
	outcome.ResponseHash = strings.ToLower(strings.TrimSpace(outcome.ResponseHash))
	outcome.ProofHash = strings.ToLower(strings.TrimSpace(outcome.ProofHash))
	outcome.ProviderID = strings.TrimSpace(outcome.ProviderID)
	outcome.ErrorCode = strings.TrimSpace(outcome.ErrorCode)
	if outcome.ExecutedHeight == 0 {
		outcome.ExecutedHeight = ctx.Height
	}
	if outcome.AnchoredHeight == 0 {
		outcome.AnchoredHeight = ctx.Height
	}
	if outcome.PaymentStatus == "" {
		outcome.PaymentStatus = ServicePaymentStatusNone
	}
	return outcome
}

func ValidateServiceExecutionOutcome(ctx ServiceConsensusContext, call ServiceCallEnvelope, outcome ServiceExecutionOutcome) error {
	if outcome.CallID != call.CallID {
		return errors.New("aetracore service outcome call id mismatch")
	}
	if !IsServiceCallStatus(outcome.Status) {
		return fmt.Errorf("unknown aetracore service call status %q", outcome.Status)
	}
	if !IsServicePaymentStatus(outcome.PaymentStatus) {
		return fmt.Errorf("unknown aetracore service payment status %q", outcome.PaymentStatus)
	}
	if outcome.ResponseHash != "" {
		if err := ValidateHash("aetracore service response hash", outcome.ResponseHash); err != nil {
			return err
		}
	}
	if outcome.ProofHash != "" {
		if err := ValidateHash("aetracore service proof hash", outcome.ProofHash); err != nil {
			return err
		}
	}
	if outcome.ExecutedHeight == 0 || outcome.ExecutedHeight > ctx.Height {
		return errors.New("aetracore service outcome executed height is invalid")
	}
	if outcome.AnchoredHeight == 0 || outcome.AnchoredHeight > ctx.Height {
		return errors.New("aetracore service outcome anchored height is invalid")
	}
	if outcome.ProviderID != "" {
		if err := validatePolicyID("aetracore service provider id", outcome.ProviderID); err != nil {
			return err
		}
	}
	if outcome.ErrorCode != "" {
		if err := validatePolicyID("aetracore service error code", outcome.ErrorCode); err != nil {
			return err
		}
	}
	if call.Kind == ServiceCallKindOffChainReceipt && outcome.ProofHash == "" && call.ProofRequirement != ServiceVerificationSignedResult {
		return errors.New("aetracore off-chain service receipt requires proof hash unless signed-result verified")
	}
	if call.Kind == ServiceCallKindMixedDispute && outcome.Status != ServiceCallStatusChallenged && outcome.Status != ServiceCallStatusSettled {
		return errors.New("aetracore mixed dispute outcome must challenge or settle")
	}
	if call.Kind == ServiceCallKindMixedSettlement && outcome.Status != ServiceCallStatusSettled {
		return errors.New("aetracore mixed settlement outcome must settle")
	}
	return nil
}

func NewServiceCallReceipt(call ServiceCallEnvelope, outcome ServiceExecutionOutcome) (ServiceCallReceipt, error) {
	receipt := ServiceCallReceipt{
		CallID:		call.CallID,
		ServiceID:	call.ServiceID,
		MethodID:	call.MethodID,
		Caller:		call.Caller,
		Status:		outcome.Status,
		RequestHash:	call.PayloadHash,
		ResponseHash:	outcome.ResponseHash,
		ProofHash:	outcome.ProofHash,
		PaymentStatus:	outcome.PaymentStatus,
		GasUsed:	outcome.GasUsed,
		ProviderID:	outcome.ProviderID,
		ExecutedHeight:	outcome.ExecutedHeight,
		AnchoredHeight:	outcome.AnchoredHeight,
		ErrorCode:	outcome.ErrorCode,
	}
	receipt.ReceiptHash = ComputeServiceCallReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (r ServiceCallReceipt) Validate() error {
	if err := ValidateHash("aetracore service receipt call id", r.CallID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service receipt service id", r.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service receipt method id", r.MethodID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service receipt caller", r.Caller); err != nil {
		return err
	}
	if !IsServiceCallStatus(r.Status) {
		return fmt.Errorf("unknown aetracore service receipt status %q", r.Status)
	}
	if err := ValidateHash("aetracore service receipt request hash", r.RequestHash); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore service receipt response hash", r.ResponseHash); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore service receipt proof hash", r.ProofHash); err != nil {
		return err
	}
	if !IsServicePaymentStatus(r.PaymentStatus) {
		return fmt.Errorf("unknown aetracore service receipt payment status %q", r.PaymentStatus)
	}
	if r.ProviderID != "" {
		if err := validatePolicyID("aetracore service receipt provider id", r.ProviderID); err != nil {
			return err
		}
	}
	if r.ExecutedHeight == 0 || r.AnchoredHeight == 0 {
		return errors.New("aetracore service receipt heights must be positive")
	}
	if r.ErrorCode != "" {
		if err := validatePolicyID("aetracore service receipt error code", r.ErrorCode); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore service receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallReceiptHash(r); r.ReceiptHash != expected {
		return fmt.Errorf("aetracore service receipt hash mismatch: expected %s", expected)
	}
	return nil
}

func (f ServiceFinalization) Validate() error {
	if f.Height == 0 {
		return errors.New("aetracore service finalization height must be positive")
	}
	if err := ValidateHash("aetracore service finalization registry root", f.RegistryRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service finalization interface root", f.InterfaceRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service finalization receipts root", f.ServiceReceiptsRoot); err != nil {
		return err
	}
	for _, receipt := range f.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	receiptsRoot, err := ComputeServiceReceiptsRoot(f.Receipts)
	if err != nil {
		return err
	}
	if receiptsRoot != f.ServiceReceiptsRoot {
		return errors.New("aetracore service finalization receipts root mismatch")
	}
	if err := ValidateHash("aetracore service finalization hash", f.FinalizationHash); err != nil {
		return err
	}
	if expected := ComputeServiceFinalizationHash(f); f.FinalizationHash != expected {
		return fmt.Errorf("aetracore service finalization hash mismatch: expected %s", expected)
	}
	return nil
}

func (m ServiceEndBlockMaintenance) Validate() error {
	if m.Height == 0 {
		return errors.New("aetracore service maintenance height must be positive")
	}
	if err := validateSortedStringSet("aetracore expired service id", m.ExpiredServiceIDs); err != nil {
		return err
	}
	if err := validateSortedStringSet("aetracore cleanup service id", m.CleanupServiceIDs); err != nil {
		return err
	}
	for i, delta := range m.ReputationDeltas {
		if err := delta.Validate(); err != nil {
			return err
		}
		if i > 0 && m.ReputationDeltas[i-1].ProviderID >= delta.ProviderID {
			return errors.New("aetracore provider reputation deltas must be sorted canonically")
		}
	}
	if err := ValidateHash("aetracore service maintenance hash", m.MaintenanceHash); err != nil {
		return err
	}
	if expected := ComputeServiceEndBlockMaintenanceHash(m); m.MaintenanceHash != expected {
		return fmt.Errorf("aetracore service maintenance hash mismatch: expected %s", expected)
	}
	return nil
}

func (d ProviderReputationDelta) Validate() error {
	if err := validatePolicyID("aetracore provider reputation id", d.ProviderID); err != nil {
		return err
	}
	if d.Successes == 0 && d.Failures == 0 {
		return errors.New("aetracore provider reputation delta must change score")
	}
	return nil
}

func (t ServiceStateTransition) Validate(state CoreState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	if err := t.Context.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service STF current state root", t.CurrentStateRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service STF next state root", t.NextStateRoot); err != nil {
		return err
	}
	t.Call = NormalizeServiceCall(t.Context, t.Call)
	if err := ValidateServiceCallForState(t.Context, state, t.Call); err != nil {
		return err
	}
	if err := validateServiceStateKeys("aetracore service STF read set", t.StateReadSet); err != nil {
		return err
	}
	if err := validateServiceStateKeys("aetracore service STF write set", t.StateWriteSet); err != nil {
		return err
	}
	if len(t.ExternalCalls) > 0 {
		return errors.New("aetracore service STF must not perform external network calls")
	}
	if t.UsesWallClock {
		return errors.New("aetracore service STF must not use nondeterministic wall-clock time")
	}
	if t.LiveAvailabilityChecks {
		return errors.New("aetracore service STF must not perform live service availability checks")
	}
	if t.IterationLimit == 0 {
		return errors.New("aetracore service STF iteration limit must be bounded")
	}
	if serviceProofRequiresMetering(t.Call.ProofRequirement) && t.ProofVerificationGas == 0 {
		return errors.New("aetracore service STF proof verification must be metered")
	}
	if len(t.DirectCrossZoneWrites) > 0 {
		return errors.New("aetracore service STF must not perform direct cross-zone writes")
	}
	return nil
}

func ComputeServiceCallID(ctx ServiceConsensusContext, call ServiceCallEnvelope) string {
	return hashParts(
		"aetra-aek-service-call-v1",
		ctx.ChainID,
		call.ServiceID,
		call.Caller,
		fmt.Sprint(call.Nonce),
		call.IdempotencyKey,
		call.PayloadHash,
	)
}

func ComputeServiceCallStateScope(call ServiceCallEnvelope) string {
	readSet := normalizeServiceStringSet(call.StateReadSet)
	writeSet := normalizeServiceStringSet(call.StateWriteSet)
	parts := []string{"aetra-aek-service-state-scope-v1"}
	if len(writeSet) > 0 {
		parts = appendStringSliceParts(parts, "write", writeSet)
		return hashParts(parts...)
	}
	parts = appendStringSliceParts(parts, "read", readSet)
	return hashParts(parts...)
}

func ComputeServiceExecutionGroupID(scope string) string {
	return hashParts("aetra-aek-service-execution-group-v1", scope)
}

func ComputeServiceInterfaceRoot(services []ServiceDescriptor) (string, error) {
	ordered := cloneServiceDescriptors(services)
	sortServiceDescriptors(ordered)
	parts := []string{"aetra-aek-service-interfaces-root-v1", fmt.Sprint(len(ordered))}
	var previous string
	for _, service := range ordered {
		service = CanonicalServiceDescriptor(service)
		if err := service.Validate(); err != nil {
			return "", err
		}
		if previous != "" && previous >= service.ServiceID {
			return "", errors.New("aetracore services must be sorted canonically by service id")
		}
		parts = append(parts, service.ServiceID, service.Interface.InterfaceHash)
		previous = service.ServiceID
	}
	return hashParts(parts...), nil
}

func ComputeServiceProposalPlanHash(plan ServiceProposalPlan) string {
	groups := cloneServiceExecutionGroups(plan.Groups)
	sortServiceExecutionGroups(groups)
	parts := []string{
		"aetra-aek-service-proposal-plan-v1",
		plan.ChainID,
		fmt.Sprint(plan.Height),
		plan.RegistryRoot,
		plan.InterfaceRoot,
		fmt.Sprint(len(groups)),
	}
	for _, group := range groups {
		parts = append(parts, group.GroupID, group.StateKeyScope, fmt.Sprint(len(group.Calls)))
		calls := cloneServiceCalls(group.Calls)
		sortServiceCalls(calls)
		for _, call := range calls {
			parts = append(parts, ComputeServiceCallEnvelopeHash(call))
		}
	}
	return hashParts(parts...)
}

func ComputeServiceCallEnvelopeHash(call ServiceCallEnvelope) string {
	call.StateReadSet = normalizeServiceStringSet(call.StateReadSet)
	call.StateWriteSet = normalizeServiceStringSet(call.StateWriteSet)
	parts := []string{
		"aetra-aek-service-call-envelope-v1",
		call.CallID,
		call.ServiceID,
		call.Caller,
		fmt.Sprint(call.Nonce),
		call.IdempotencyKey,
		call.MethodID,
		call.InterfaceHash,
		call.PayloadHash,
		call.PaymentDenom,
		call.MaxFeeAmount,
		string(call.ProofRequirement),
		string(call.Kind),
		fmt.Sprint(call.CreatedHeight),
		fmt.Sprint(call.DeadlineHeight),
		fmt.Sprint(call.PriorityClass),
		fmt.Sprint(call.Callback),
		call.RetryOf,
		fmt.Sprint(call.Dispute),
	}
	parts = appendStringSliceParts(parts, "read", call.StateReadSet)
	parts = appendStringSliceParts(parts, "write", call.StateWriteSet)
	return hashParts(parts...)
}

func ComputeServiceCallReceiptHash(receipt ServiceCallReceipt) string {
	return hashParts(
		"aetra-aek-service-call-receipt-v1",
		receipt.CallID,
		receipt.ServiceID,
		receipt.MethodID,
		receipt.Caller,
		string(receipt.Status),
		receipt.RequestHash,
		receipt.ResponseHash,
		receipt.ProofHash,
		string(receipt.PaymentStatus),
		fmt.Sprint(receipt.GasUsed),
		receipt.ProviderID,
		fmt.Sprint(receipt.ExecutedHeight),
		fmt.Sprint(receipt.AnchoredHeight),
		receipt.ErrorCode,
	)
}

func ComputeServiceReceiptsRoot(receipts []ServiceCallReceipt) (string, error) {
	ordered := cloneServiceCallReceipts(receipts)
	sortServiceCallReceipts(ordered)
	parts := []string{"aetra-aek-service-receipts-root-v1", fmt.Sprint(len(ordered))}
	var previous string
	for _, receipt := range ordered {
		if err := receipt.Validate(); err != nil {
			return "", err
		}
		if previous != "" && previous >= receipt.CallID {
			return "", errors.New("aetracore service receipts must be sorted canonically by call id")
		}
		parts = append(parts, receipt.ReceiptHash)
		previous = receipt.CallID
	}
	return hashParts(parts...), nil
}

func ComputeServiceFinalizationHash(finalization ServiceFinalization) string {
	return hashParts(
		"aetra-aek-service-finalization-v1",
		fmt.Sprint(finalization.Height),
		finalization.RegistryRoot,
		finalization.InterfaceRoot,
		finalization.ServiceReceiptsRoot,
		fmt.Sprint(len(finalization.Receipts)),
	)
}

func ComputeServiceEndBlockMaintenanceHash(m ServiceEndBlockMaintenance) string {
	parts := []string{
		"aetra-aek-service-endblock-maintenance-v1",
		fmt.Sprint(m.Height),
		fmt.Sprint(m.Metrics.ActiveServices),
		fmt.Sprint(m.Metrics.ExpiredServices),
		fmt.Sprint(m.Metrics.ReceiptCount),
		fmt.Sprint(m.Metrics.FailedCallCount),
		fmt.Sprint(m.Metrics.SettledPaymentCount),
		fmt.Sprint(m.Metrics.ProviderReceiptCount),
	}
	parts = appendStringSliceParts(parts, "expired", m.ExpiredServiceIDs)
	parts = appendStringSliceParts(parts, "cleanup", m.CleanupServiceIDs)
	parts = append(parts, "reputation", fmt.Sprint(len(m.ReputationDeltas)))
	for _, delta := range m.ReputationDeltas {
		parts = append(parts, delta.ProviderID, fmt.Sprint(delta.Successes), fmt.Sprint(delta.Failures))
	}
	return hashParts(parts...)
}

func (d ServiceInterfaceDescriptor) MethodByID(methodID string) (ServiceMethodDescriptor, bool) {
	for _, method := range d.Methods {
		if method.MethodID == methodID {
			return method, true
		}
	}
	return ServiceMethodDescriptor{}, false
}

func IsServicePipelinePhase(phase ServicePipelinePhase) bool {
	switch phase {
	case ServicePhasePrepareProposal, ServicePhaseProcessProposal, ServicePhaseDeliverTxCompat, ServicePhaseFinalizeBlock, ServicePhaseEndBlock:
		return true
	default:
		return false
	}
}

func IsServiceCallKind(kind ServiceCallKind) bool {
	switch kind {
	case ServiceCallKindOnChain, ServiceCallKindOffChainReceipt, ServiceCallKindMixedDispute, ServiceCallKindMixedSettlement, ServiceCallKindCallback, ServiceCallKindRetry:
		return true
	default:
		return false
	}
}

func IsServiceCallStatus(status ServiceCallStatus) bool {
	switch status {
	case ServiceCallStatusAccepted, ServiceCallStatusExecuted, ServiceCallStatusFailed, ServiceCallStatusExpired,
		ServiceCallStatusChallenged, ServiceCallStatusSettled, ServiceCallStatusReverted:
		return true
	default:
		return false
	}
}

func IsServicePaymentStatus(status ServicePaymentStatus) bool {
	switch status {
	case ServicePaymentStatusNone, ServicePaymentStatusReserved, ServicePaymentStatusSettled, ServicePaymentStatusRefunded, ServicePaymentStatusEscrowed:
		return true
	default:
		return false
	}
}

func serviceCallKindAllowed(serviceType ServiceType, kind ServiceCallKind) bool {
	switch kind {
	case ServiceCallKindCallback, ServiceCallKindRetry:
		return true
	}
	switch serviceType {
	case ServiceTypeOnChain:
		return kind == ServiceCallKindOnChain
	case ServiceTypeOffChain, ServiceTypeFogMarket:
		return kind == ServiceCallKindOffChainReceipt
	case ServiceTypeMixed:
		return kind == ServiceCallKindOffChainReceipt || kind == ServiceCallKindMixedDispute || kind == ServiceCallKindMixedSettlement
	default:
		return false
	}
}

func serviceProofRequiresMetering(model ServiceVerificationModel) bool {
	switch model {
	case ServiceVerificationProofAnchored, ServiceVerificationChallengeWindow, ServiceVerificationEconomicCollateral:
		return true
	default:
		return false
	}
}

func BuildProviderReputationDeltas(receipts []ServiceCallReceipt) []ProviderReputationDelta {
	byProvider := make(map[string]*ProviderReputationDelta)
	for _, receipt := range receipts {
		if receipt.ProviderID == "" {
			continue
		}
		delta, found := byProvider[receipt.ProviderID]
		if !found {
			delta = &ProviderReputationDelta{ProviderID: receipt.ProviderID}
			byProvider[receipt.ProviderID] = delta
		}
		switch receipt.Status {
		case ServiceCallStatusExecuted, ServiceCallStatusSettled, ServiceCallStatusAccepted:
			delta.Successes++
		default:
			delta.Failures++
		}
	}
	out := make([]ProviderReputationDelta, 0, len(byProvider))
	for _, delta := range byProvider {
		out = append(out, *delta)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ProviderID < out[j].ProviderID })
	return out
}

func sortServiceCalls(calls []ServiceCallEnvelope) {
	sort.SliceStable(calls, func(i, j int) bool {
		return compareServiceCalls(calls[i], calls[j]) < 0
	})
}

func compareServiceCalls(left, right ServiceCallEnvelope) int {
	if left.ServiceID == right.ServiceID && left.Caller == right.Caller {
		if left.Nonce < right.Nonce {
			return -1
		}
		if left.Nonce > right.Nonce {
			return 1
		}
	}
	if left.Dispute != right.Dispute {
		if left.Dispute {
			return -1
		}
		return 1
	}
	if rankLeft, rankRight := serviceCallKindRank(left.Kind), serviceCallKindRank(right.Kind); rankLeft != rankRight {
		if rankLeft < rankRight {
			return -1
		}
		return 1
	}
	if left.DeadlineHeight != right.DeadlineHeight {
		if left.DeadlineHeight < right.DeadlineHeight {
			return -1
		}
		return 1
	}
	if feeCmp := compareDecimalString(left.MaxFeeAmount, right.MaxFeeAmount); feeCmp != 0 {
		return -feeCmp
	}
	if left.PriorityClass != right.PriorityClass {
		if left.PriorityClass < right.PriorityClass {
			return -1
		}
		return 1
	}
	if left.ServiceID != right.ServiceID {
		if left.ServiceID < right.ServiceID {
			return -1
		}
		return 1
	}
	if left.Caller != right.Caller {
		if left.Caller < right.Caller {
			return -1
		}
		return 1
	}
	if left.CallID < right.CallID {
		return -1
	}
	if left.CallID > right.CallID {
		return 1
	}
	return 0
}

func serviceCallKindRank(kind ServiceCallKind) int {
	switch kind {
	case ServiceCallKindMixedDispute:
		return 0
	case ServiceCallKindMixedSettlement:
		return 1
	case ServiceCallKindCallback:
		return 2
	case ServiceCallKindRetry:
		return 3
	case ServiceCallKindOnChain:
		return 4
	case ServiceCallKindOffChainReceipt:
		return 5
	default:
		return 99
	}
}

func sortServiceExecutionGroups(groups []ServiceExecutionGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].GroupID < groups[j].GroupID
	})
	for i := range groups {
		sortServiceCalls(groups[i].Calls)
	}
}

func cloneServiceCalls(calls []ServiceCallEnvelope) []ServiceCallEnvelope {
	out := make([]ServiceCallEnvelope, len(calls))
	for i, call := range calls {
		out[i] = call
		out[i].StateReadSet = append([]string(nil), call.StateReadSet...)
		out[i].StateWriteSet = append([]string(nil), call.StateWriteSet...)
	}
	return out
}

func cloneServiceExecutionGroups(groups []ServiceExecutionGroup) []ServiceExecutionGroup {
	out := make([]ServiceExecutionGroup, len(groups))
	for i, group := range groups {
		out[i] = group
		out[i].Calls = cloneServiceCalls(group.Calls)
	}
	return out
}

func cloneServiceCallReceipts(receipts []ServiceCallReceipt) []ServiceCallReceipt {
	out := make([]ServiceCallReceipt, len(receipts))
	copy(out, receipts)
	return out
}

func sortServiceCallReceipts(receipts []ServiceCallReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return receipts[i].CallID < receipts[j].CallID
	})
}

func normalizeServiceStringSet(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
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

func validateServiceStateKeys(fieldName string, keys []string) error {
	return validateSortedStringSet(fieldName, keys)
}

func validateSortedStringSet(fieldName string, values []string) error {
	var previous string
	for i, value := range values {
		if err := validatePolicyID(fieldName, value); err != nil {
			return err
		}
		if i > 0 && previous >= value {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func compareDecimalString(left, right string) int {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if left == "" {
		left = "0"
	}
	if right == "" {
		right = "0"
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func countFailedServiceReceipts(receipts []ServiceCallReceipt) uint64 {
	var count uint64
	for _, receipt := range receipts {
		switch receipt.Status {
		case ServiceCallStatusFailed, ServiceCallStatusExpired, ServiceCallStatusReverted, ServiceCallStatusChallenged:
			count++
		}
	}
	return count
}

func countSettledServicePayments(receipts []ServiceCallReceipt) uint64 {
	var count uint64
	for _, receipt := range receipts {
		if receipt.PaymentStatus == ServicePaymentStatusSettled {
			count++
		}
	}
	return count
}

func countProviderReceipts(receipts []ServiceCallReceipt) uint64 {
	var count uint64
	for _, receipt := range receipts {
		if receipt.ProviderID != "" {
			count++
		}
	}
	return count
}
