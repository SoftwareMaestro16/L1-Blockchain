package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	ServicePaymentModelPrefix	= ServiceStorePrefix + "payments/models"
	ServicePaymentEscrowPrefix	= ServiceStorePrefix + "payments/escrow"
	ServicePaymentStreamPrefix	= ServiceStorePrefix + "payments/streams"
	ServicePaymentMeterPrefix	= ServiceStorePrefix + "payments/meters"
	ServicePaymentSettlementPrefix	= ServiceStorePrefix + "payments/settlements"
)

type ServicePaymentModel struct {
	ServiceID		string
	SupportedDenoms		[]string
	DefaultDenom		string
	PricingUnit		coretypes.ServicePricingUnit
	SettlementMode		coretypes.ServicePaymentSettlementMode
	UnitAmount		string
	MaxAmountOptional	string
	FailurePolicy		coretypes.ServiceFailureBehavior
	ProtocolNative		bool
	KnownBeforeSigning	bool
	UpdatedHeight		uint64
	ModelHash		string
}

type ServiceEscrow struct {
	EscrowID	string
	ServiceID	string
	Payer		string
	Denom		string
	Amount		string
	LockedHeight	uint64
	ExpiryHeight	uint64
	LockHash	string
}

type PaymentStream struct {
	StreamID	string
	ServiceID	string
	Payer		string
	Denom		string
	RatePerHeight	string
	StartHeight	uint64
	EndHeight	uint64
	PaidThrough	uint64
	StreamHash	string
}

type MeteredUsage struct {
	MeterID		string
	ServiceID	string
	CallID		string
	UsageReceipt	PaymentUsageReceipt
	AmountDue	string
	RecordedHeight	uint64
	UsageHash	string
}

type PaymentSettlement struct {
	CallID			string
	ServiceID		string
	EnvelopeHash		string
	QuoteHash		string
	AmountSettled		string
	Denom			string
	Status			coretypes.ServicePaymentStatus
	FailurePolicy		coretypes.ServiceFailureBehavior
	SettlementHeight	uint64
	SettlementHash		string
}

type FinancialZonePaymentRoute struct {
	RouteID		string
	ServiceID	string
	Payer		string
	Denom		string
	Amount		string
	BankKeeper	string
	FinancialZone	string
	RouteHash	string
}

type ServicePaymentState struct {
	Models		[]ServicePaymentModel
	Escrows		[]ServiceEscrow
	Streams		[]PaymentStream
	Meters		[]MeteredUsage
	Settlements	[]PaymentSettlement
	Height		uint64
	StateRoot	string
}

type ServicePaymentProof struct {
	Key		string
	ValueHash	string
	StateRoot	string
	Height		uint64
	ProofHash	string
}

type QueryPaymentModel struct {
	ServiceID string
}

type QueryPaymentProof struct {
	Key string
}

type QueryPaymentModelResponse struct {
	Model	ServicePaymentModel
	Found	bool
}

type QueryPaymentProofResponse struct {
	Proof	ServicePaymentProof
	Found	bool
}

func PaymentModelStateKey(serviceID string) (string, error) {
	if err := validateInterfaceToken("services payment model service id", serviceID); err != nil {
		return "", err
	}
	return ServicePaymentModelPrefix + "/" + serviceID, nil
}

func ServiceEscrowStateKey(escrowID string) (string, error) {
	if err := validateInterfaceToken("services payment escrow id", escrowID); err != nil {
		return "", err
	}
	return ServicePaymentEscrowPrefix + "/" + escrowID, nil
}

func PaymentStreamStateKey(streamID string) (string, error) {
	if err := validateInterfaceToken("services payment stream id", streamID); err != nil {
		return "", err
	}
	return ServicePaymentStreamPrefix + "/" + streamID, nil
}

func MeteredUsageStateKey(meterID string) (string, error) {
	if err := validateInterfaceToken("services payment meter id", meterID); err != nil {
		return "", err
	}
	return ServicePaymentMeterPrefix + "/" + meterID, nil
}

func PaymentSettlementStateKey(callID string) (string, error) {
	if err := coretypes.ValidateHash("services payment settlement call id", callID); err != nil {
		return "", err
	}
	return ServicePaymentSettlementPrefix + "/" + callID, nil
}

func NewServicePaymentModel(model ServicePaymentModel) (ServicePaymentModel, error) {
	if model.ModelHash != "" {
		return ServicePaymentModel{}, errors.New("services payment model hash must be empty before construction")
	}
	model = canonicalServicePaymentModel(model)
	if err := model.ValidateFormat(); err != nil {
		return ServicePaymentModel{}, err
	}
	model.ModelHash = ComputeServicePaymentModelHash(model)
	return model, model.Validate()
}

func NewServicePaymentModelFromDescriptor(descriptor ServiceDescriptor) (ServicePaymentModel, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServicePaymentModel{}, err
	}
	maxAmount := descriptor.Payment.MaxAmount
	if maxAmount == "" {
		maxAmount = descriptor.Payment.Amount
	}
	supportedDenoms := []string{descriptor.Payment.Denom}
	if isProtocolNativePaymentDescriptor(descriptor) {
		supportedDenoms = append(supportedDenoms, coretypes.NativeFeePolicyID)
	}
	return NewServicePaymentModel(ServicePaymentModel{
		ServiceID:		descriptor.ServiceID,
		SupportedDenoms:	supportedDenoms,
		DefaultDenom:		descriptor.Payment.Denom,
		PricingUnit:		descriptor.Payment.PricingUnit,
		SettlementMode:		descriptor.Payment.SettlementMode,
		UnitAmount:		descriptor.Payment.Amount,
		MaxAmountOptional:	maxAmount,
		FailurePolicy:		descriptor.Execution.FailureBehavior,
		ProtocolNative:		isProtocolNativePaymentDescriptor(descriptor),
		KnownBeforeSigning:	true,
		UpdatedHeight:		descriptor.UpdatedHeight,
	})
}

func (model ServicePaymentModel) ValidateFormat() error {
	model = canonicalServicePaymentModel(model)
	if _, err := PaymentModelStateKey(model.ServiceID); err != nil {
		return err
	}
	if len(model.SupportedDenoms) == 0 {
		return errors.New("services payment model supported denoms are required")
	}
	if err := validatePaymentDenoms(model.SupportedDenoms); err != nil {
		return err
	}
	if !stringInSet(model.DefaultDenom, model.SupportedDenoms) {
		return errors.New("services payment model default denom must be supported")
	}
	if model.ProtocolNative && !stringInSet(coretypes.NativeFeePolicyID, model.SupportedDenoms) {
		return errors.New("services protocol-native payment model must support naet")
	}
	if !coretypes.IsServicePricingUnit(model.PricingUnit) {
		return fmt.Errorf("unknown services payment model pricing unit %q", model.PricingUnit)
	}
	if !coretypes.IsServicePaymentSettlementMode(model.SettlementMode) {
		return fmt.Errorf("unknown services payment model settlement mode %q", model.SettlementMode)
	}
	if err := validatePositivePaymentAmount("services payment model unit amount", model.UnitAmount); err != nil {
		return err
	}
	if model.MaxAmountOptional != "" {
		if err := validatePositivePaymentAmount("services payment model max amount", model.MaxAmountOptional); err != nil {
			return err
		}
		if comparePaymentAmounts(model.MaxAmountOptional, model.UnitAmount) < 0 {
			return errors.New("services payment model max amount must cover unit amount")
		}
	}
	if !coretypes.IsServiceFailureBehavior(model.FailurePolicy) {
		return fmt.Errorf("unknown services payment model failure policy %q", model.FailurePolicy)
	}
	if !model.KnownBeforeSigning {
		return errors.New("services payment model must be known before call signing")
	}
	if model.UpdatedHeight == 0 {
		return errors.New("services payment model updated height is required")
	}
	if model.ModelHash != "" {
		return coretypes.ValidateHash("services payment model hash", model.ModelHash)
	}
	return nil
}

func (model ServicePaymentModel) Validate() error {
	model = canonicalServicePaymentModel(model)
	if err := model.ValidateFormat(); err != nil {
		return err
	}
	if model.ModelHash == "" {
		return errors.New("services payment model hash is required")
	}
	if expected := ComputeServicePaymentModelHash(model); model.ModelHash != expected {
		return fmt.Errorf("services payment model hash mismatch: expected %s", expected)
	}
	return nil
}

func NewServiceEscrow(escrow ServiceEscrow) (ServiceEscrow, error) {
	if escrow.LockHash != "" {
		return ServiceEscrow{}, errors.New("services escrow hash must be empty before construction")
	}
	escrow = canonicalServiceEscrow(escrow)
	if err := escrow.ValidateFormat(); err != nil {
		return ServiceEscrow{}, err
	}
	escrow.LockHash = ComputeServiceEscrowHash(escrow)
	return escrow, escrow.Validate()
}

func (escrow ServiceEscrow) ValidateFormat() error {
	escrow = canonicalServiceEscrow(escrow)
	if _, err := ServiceEscrowStateKey(escrow.EscrowID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services escrow service id", escrow.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("services escrow payer", escrow.Payer); err != nil {
		return err
	}
	if err := validateInterfaceToken("services escrow denom", escrow.Denom); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("services escrow amount", escrow.Amount); err != nil {
		return err
	}
	if escrow.LockedHeight == 0 || escrow.ExpiryHeight <= escrow.LockedHeight {
		return errors.New("services escrow height range is invalid")
	}
	if escrow.LockHash != "" {
		return coretypes.ValidateHash("services escrow hash", escrow.LockHash)
	}
	return nil
}

func (escrow ServiceEscrow) Validate() error {
	escrow = canonicalServiceEscrow(escrow)
	if err := escrow.ValidateFormat(); err != nil {
		return err
	}
	if escrow.LockHash == "" {
		return errors.New("services escrow hash is required")
	}
	if expected := ComputeServiceEscrowHash(escrow); escrow.LockHash != expected {
		return fmt.Errorf("services escrow hash mismatch: expected %s", expected)
	}
	return nil
}

func NewPaymentStream(stream PaymentStream) (PaymentStream, error) {
	if stream.StreamHash != "" {
		return PaymentStream{}, errors.New("services payment stream hash must be empty before construction")
	}
	stream = canonicalPaymentStream(stream)
	if err := stream.ValidateFormat(); err != nil {
		return PaymentStream{}, err
	}
	stream.StreamHash = ComputePaymentStreamHash(stream)
	return stream, stream.Validate()
}

func (stream PaymentStream) ValidateFormat() error {
	stream = canonicalPaymentStream(stream)
	if _, err := PaymentStreamStateKey(stream.StreamID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment stream service id", stream.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("services payment stream payer", stream.Payer); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment stream denom", stream.Denom); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("services payment stream rate", stream.RatePerHeight); err != nil {
		return err
	}
	if stream.StartHeight == 0 || stream.EndHeight <= stream.StartHeight {
		return errors.New("services payment stream height range is invalid")
	}
	if stream.PaidThrough < stream.StartHeight || stream.PaidThrough > stream.EndHeight {
		return errors.New("services payment stream paid-through height is outside range")
	}
	if stream.StreamHash != "" {
		return coretypes.ValidateHash("services payment stream hash", stream.StreamHash)
	}
	return nil
}

func (stream PaymentStream) Validate() error {
	stream = canonicalPaymentStream(stream)
	if err := stream.ValidateFormat(); err != nil {
		return err
	}
	if stream.StreamHash == "" {
		return errors.New("services payment stream hash is required")
	}
	if expected := ComputePaymentStreamHash(stream); stream.StreamHash != expected {
		return fmt.Errorf("services payment stream hash mismatch: expected %s", expected)
	}
	return nil
}

func NewMeteredUsage(usage MeteredUsage) (MeteredUsage, error) {
	if usage.UsageHash != "" {
		return MeteredUsage{}, errors.New("services metered usage hash must be empty before construction")
	}
	usage = canonicalMeteredUsage(usage)
	if err := usage.ValidateFormat(); err != nil {
		return MeteredUsage{}, err
	}
	usage.UsageHash = ComputeMeteredUsageHash(usage)
	return usage, usage.Validate()
}

func (usage MeteredUsage) ValidateFormat() error {
	usage = canonicalMeteredUsage(usage)
	if _, err := MeteredUsageStateKey(usage.MeterID); err != nil {
		return err
	}
	if err := usage.UsageReceipt.Validate(); err != nil {
		return err
	}
	if usage.ServiceID != usage.UsageReceipt.ServiceID || usage.CallID != usage.UsageReceipt.CallID {
		return errors.New("services metered usage receipt mismatch")
	}
	if err := validatePositivePaymentAmount("services metered usage amount due", usage.AmountDue); err != nil {
		return err
	}
	if usage.RecordedHeight == 0 {
		return errors.New("services metered usage recorded height is required")
	}
	if usage.UsageHash != "" {
		return coretypes.ValidateHash("services metered usage hash", usage.UsageHash)
	}
	return nil
}

func (usage MeteredUsage) Validate() error {
	usage = canonicalMeteredUsage(usage)
	if err := usage.ValidateFormat(); err != nil {
		return err
	}
	if usage.UsageHash == "" {
		return errors.New("services metered usage hash is required")
	}
	if expected := ComputeMeteredUsageHash(usage); usage.UsageHash != expected {
		return fmt.Errorf("services metered usage hash mismatch: expected %s", expected)
	}
	return nil
}

func NewPaymentSettlement(settlement PaymentSettlement) (PaymentSettlement, error) {
	if settlement.SettlementHash != "" {
		return PaymentSettlement{}, errors.New("services payment settlement hash must be empty before construction")
	}
	settlement = canonicalPaymentSettlement(settlement)
	if err := settlement.ValidateFormat(); err != nil {
		return PaymentSettlement{}, err
	}
	settlement.SettlementHash = ComputePaymentSettlementHash(settlement)
	return settlement, settlement.Validate()
}

func (settlement PaymentSettlement) ValidateFormat() error {
	settlement = canonicalPaymentSettlement(settlement)
	if _, err := PaymentSettlementStateKey(settlement.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment settlement service id", settlement.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services payment settlement envelope hash", settlement.EnvelopeHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services payment settlement quote hash", settlement.QuoteHash); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("services payment settlement amount", settlement.AmountSettled); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment settlement denom", settlement.Denom); err != nil {
		return err
	}
	if !coretypes.IsServicePaymentStatus(settlement.Status) {
		return fmt.Errorf("unknown services payment settlement status %q", settlement.Status)
	}
	if !coretypes.IsServiceFailureBehavior(settlement.FailurePolicy) {
		return fmt.Errorf("unknown services payment settlement failure policy %q", settlement.FailurePolicy)
	}
	if settlement.SettlementHeight == 0 {
		return errors.New("services payment settlement height is required")
	}
	if settlement.SettlementHash != "" {
		return coretypes.ValidateHash("services payment settlement hash", settlement.SettlementHash)
	}
	return nil
}

func (settlement PaymentSettlement) Validate() error {
	settlement = canonicalPaymentSettlement(settlement)
	if err := settlement.ValidateFormat(); err != nil {
		return err
	}
	if settlement.SettlementHash == "" {
		return errors.New("services payment settlement hash is required")
	}
	if expected := ComputePaymentSettlementHash(settlement); settlement.SettlementHash != expected {
		return fmt.Errorf("services payment settlement hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildServicePaymentState(models []ServicePaymentModel, escrows []ServiceEscrow, streams []PaymentStream, meters []MeteredUsage, settlements []PaymentSettlement, height uint64) (ServicePaymentState, error) {
	if height == 0 {
		return ServicePaymentState{}, errors.New("services payment state height is required")
	}
	state := ServicePaymentState{
		Models:		normalizeServicePaymentModels(models),
		Escrows:	normalizeServiceEscrows(escrows),
		Streams:	normalizePaymentStreams(streams),
		Meters:		normalizeMeteredUsages(meters),
		Settlements:	normalizePaymentSettlements(settlements),
		Height:		height,
	}
	if err := state.ValidateFormat(); err != nil {
		return ServicePaymentState{}, err
	}
	state.StateRoot = ComputeServicePaymentStateRoot(state)
	return state, state.Validate()
}

func (state ServicePaymentState) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("services payment state height is required")
	}
	if err := validateUniquePaymentModels(state.Models); err != nil {
		return err
	}
	for _, escrow := range state.Escrows {
		if err := escrow.Validate(); err != nil {
			return err
		}
	}
	for _, stream := range state.Streams {
		if err := stream.Validate(); err != nil {
			return err
		}
	}
	for _, usage := range state.Meters {
		if err := usage.Validate(); err != nil {
			return err
		}
	}
	for _, settlement := range state.Settlements {
		if err := settlement.Validate(); err != nil {
			return err
		}
	}
	if state.StateRoot != "" {
		return coretypes.ValidateHash("services payment state root", state.StateRoot)
	}
	return nil
}

func (state ServicePaymentState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("services payment state root is required")
	}
	if expected := ComputeServicePaymentStateRoot(state); state.StateRoot != expected {
		return fmt.Errorf("services payment state root mismatch: expected %s", expected)
	}
	return nil
}

func QueryServicePaymentModelFromState(state ServicePaymentState, q QueryPaymentModel) (QueryPaymentModelResponse, error) {
	if err := validateInterfaceToken("services payment model query service id", q.ServiceID); err != nil {
		return QueryPaymentModelResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryPaymentModelResponse{}, err
	}
	for _, model := range state.Models {
		if model.ServiceID == q.ServiceID {
			return QueryPaymentModelResponse{Model: model, Found: true}, nil
		}
	}
	return QueryPaymentModelResponse{Found: false}, nil
}

func QueryServicePaymentProofFromState(state ServicePaymentState, q QueryPaymentProof) (QueryPaymentProofResponse, error) {
	if strings.TrimSpace(q.Key) == "" {
		return QueryPaymentProofResponse{}, errors.New("services payment proof query key is required")
	}
	if err := state.Validate(); err != nil {
		return QueryPaymentProofResponse{}, err
	}
	valueHash, found := servicePaymentValueHashForKey(state, q.Key)
	if !found {
		return QueryPaymentProofResponse{Found: false}, nil
	}
	proof := ServicePaymentProof{Key: q.Key, ValueHash: valueHash, StateRoot: state.StateRoot, Height: state.Height}
	proof.ProofHash = ComputeServicePaymentProofHash(proof)
	return QueryPaymentProofResponse{Proof: proof, Found: true}, proof.Validate()
}

func (proof ServicePaymentProof) Validate() error {
	if strings.TrimSpace(proof.Key) == "" {
		return errors.New("services payment proof key is required")
	}
	if err := coretypes.ValidateHash("services payment proof value hash", proof.ValueHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services payment proof state root", proof.StateRoot); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("services payment proof height is required")
	}
	if err := coretypes.ValidateHash("services payment proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeServicePaymentProofHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("services payment proof hash mismatch: expected %s", expected)
	}
	return nil
}

func EnsureEnvelopeMatchesPaymentModel(envelope PaymentEnvelope, model ServicePaymentModel) error {
	if err := envelope.Validate(); err != nil {
		return err
	}
	if err := model.Validate(); err != nil {
		return err
	}
	if envelope.PayeeService != model.ServiceID {
		return errors.New("services payment envelope service mismatch")
	}
	if !stringInSet(envelope.Denom, model.SupportedDenoms) {
		return errors.New("services payment denom is not supported for service")
	}
	if envelope.Denom != model.DefaultDenom && model.ProtocolNative && envelope.Denom != coretypes.NativeFeePolicyID {
		return errors.New("services protocol-native payment must use supported native denom")
	}
	if envelope.PricingUnit != model.PricingUnit || envelope.SettlementMode != model.SettlementMode {
		return errors.New("services payment envelope model mismatch")
	}
	if comparePaymentAmounts(envelope.Amount, model.UnitAmount) < 0 {
		return errors.New("services payment envelope amount is below model unit amount")
	}
	return nil
}

func CreateServiceEscrowFromEnvelope(envelope PaymentEnvelope, height uint64) (ServiceEscrow, error) {
	if err := envelope.Validate(); err != nil {
		return ServiceEscrow{}, err
	}
	if envelope.SettlementMode != coretypes.ServicePaymentEscrow || envelope.EscrowIDOptional == "" {
		return ServiceEscrow{}, errors.New("services escrow creation requires escrow payment envelope")
	}
	return NewServiceEscrow(ServiceEscrow{
		EscrowID:	envelope.EscrowIDOptional,
		ServiceID:	envelope.PayeeService,
		Payer:		envelope.Payer,
		Denom:		envelope.Denom,
		Amount:		envelope.Amount,
		LockedHeight:	height,
		ExpiryHeight:	envelope.ExpiryHeight,
	})
}

func CreatePaymentStreamFromEnvelope(envelope PaymentEnvelope, startHeight, endHeight uint64) (PaymentStream, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentStream{}, err
	}
	if envelope.SettlementMode != coretypes.ServicePaymentStreaming || envelope.StreamIDOptional == "" {
		return PaymentStream{}, errors.New("services stream creation requires streaming payment envelope")
	}
	return NewPaymentStream(PaymentStream{
		StreamID:	envelope.StreamIDOptional,
		ServiceID:	envelope.PayeeService,
		Payer:		envelope.Payer,
		Denom:		envelope.Denom,
		RatePerHeight:	envelope.Amount,
		StartHeight:	startHeight,
		EndHeight:	endHeight,
		PaidThrough:	startHeight,
	})
}

func RecordMeteredUsageFromQuote(meterID string, quote PaymentModelQuote, receipt PaymentUsageReceipt, height uint64) (MeteredUsage, error) {
	if err := quote.Validate(); err != nil {
		return MeteredUsage{}, err
	}
	if !quote.RequiresUsageReceipt {
		return MeteredUsage{}, errors.New("services metered usage requires usage receipt quote")
	}
	return NewMeteredUsage(MeteredUsage{
		MeterID:	meterID,
		ServiceID:	quote.Envelope.PayeeService,
		CallID:		receipt.CallID,
		UsageReceipt:	receipt,
		AmountDue:	quote.AmountDue,
		RecordedHeight:	height,
	})
}

func SettlePaymentFromQuote(callID string, quote PaymentModelQuote, status coretypes.ServicePaymentStatus, failurePolicy coretypes.ServiceFailureBehavior, height uint64) (PaymentSettlement, error) {
	if err := quote.Validate(); err != nil {
		return PaymentSettlement{}, err
	}
	amount := quote.AmountDue
	if status != PaymentStatusForFailurePolicy(failurePolicy) && status != coretypes.ServicePaymentStatusSettled && status != coretypes.ServicePaymentStatusEscrowed {
		return PaymentSettlement{}, errors.New("services payment settlement status does not follow failure policy")
	}
	return NewPaymentSettlement(PaymentSettlement{
		CallID:			callID,
		ServiceID:		quote.Envelope.PayeeService,
		EnvelopeHash:		quote.Envelope.EnvelopeHash,
		QuoteHash:		quote.QuoteHash,
		AmountSettled:		amount,
		Denom:			quote.Envelope.Denom,
		Status:			status,
		FailurePolicy:		failurePolicy,
		SettlementHeight:	height,
	})
}

func PaymentStatusForFailurePolicy(failurePolicy coretypes.ServiceFailureBehavior) coretypes.ServicePaymentStatus {
	switch failurePolicy {
	case coretypes.ServiceFailureRefund, coretypes.ServiceFailureRevert, coretypes.ServiceFailureFallbackOnChain:
		return coretypes.ServicePaymentStatusRefunded
	case coretypes.ServiceFailureChallenge, coretypes.ServiceFailureSlashProvider:
		return coretypes.ServicePaymentStatusEscrowed
	case coretypes.ServiceFailurePartialSettle:
		return coretypes.ServicePaymentStatusSettled
	case coretypes.ServiceFailureRetry:
		return coretypes.ServicePaymentStatusReserved
	default:
		return coretypes.ServicePaymentStatusNone
	}
}

func BuildFinancialZonePaymentRoute(envelope PaymentEnvelope, bankKeeper, financialZone string) (FinancialZonePaymentRoute, error) {
	if err := envelope.Validate(); err != nil {
		return FinancialZonePaymentRoute{}, err
	}
	route := FinancialZonePaymentRoute{
		RouteID:	servicesHashParts("aetra-services-financial-route-id-v1", envelope.EnvelopeHash),
		ServiceID:	envelope.PayeeService,
		Payer:		envelope.Payer,
		Denom:		envelope.Denom,
		Amount:		envelope.Amount,
		BankKeeper:	strings.TrimSpace(bankKeeper),
		FinancialZone:	strings.TrimSpace(financialZone),
	}
	if err := validateInterfaceToken("services financial route bank keeper", route.BankKeeper); err != nil {
		return FinancialZonePaymentRoute{}, err
	}
	if err := validateInterfaceToken("services financial route zone", route.FinancialZone); err != nil {
		return FinancialZonePaymentRoute{}, err
	}
	route.RouteHash = ComputeFinancialZonePaymentRouteHash(route)
	return route, route.Validate()
}

func (route FinancialZonePaymentRoute) Validate() error {
	if err := coretypes.ValidateHash("services financial route id", route.RouteID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services financial route service id", route.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("services financial route payer", route.Payer); err != nil {
		return err
	}
	if err := validateInterfaceToken("services financial route denom", route.Denom); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("services financial route amount", route.Amount); err != nil {
		return err
	}
	if err := validateInterfaceToken("services financial route bank keeper", route.BankKeeper); err != nil {
		return err
	}
	if err := validateInterfaceToken("services financial route zone", route.FinancialZone); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services financial route hash", route.RouteHash); err != nil {
		return err
	}
	if expected := ComputeFinancialZonePaymentRouteHash(route); route.RouteHash != expected {
		return fmt.Errorf("services financial route hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServicePaymentModelHash(model ServicePaymentModel) string {
	model = canonicalServicePaymentModel(model)
	parts := []string{"aetra-services-payment-model-v1", model.ServiceID, model.DefaultDenom, string(model.PricingUnit), string(model.SettlementMode), model.UnitAmount, model.MaxAmountOptional, string(model.FailurePolicy), fmt.Sprint(model.ProtocolNative), fmt.Sprint(model.KnownBeforeSigning), fmt.Sprint(model.UpdatedHeight), fmt.Sprint(len(model.SupportedDenoms))}
	parts = append(parts, model.SupportedDenoms...)
	return servicesHashParts(parts...)
}

func ComputeServiceEscrowHash(escrow ServiceEscrow) string {
	escrow = canonicalServiceEscrow(escrow)
	return servicesHashParts("aetra-services-payment-escrow-v1", escrow.EscrowID, escrow.ServiceID, escrow.Payer, escrow.Denom, escrow.Amount, fmt.Sprint(escrow.LockedHeight), fmt.Sprint(escrow.ExpiryHeight))
}

func ComputePaymentStreamHash(stream PaymentStream) string {
	stream = canonicalPaymentStream(stream)
	return servicesHashParts("aetra-services-payment-stream-v1", stream.StreamID, stream.ServiceID, stream.Payer, stream.Denom, stream.RatePerHeight, fmt.Sprint(stream.StartHeight), fmt.Sprint(stream.EndHeight), fmt.Sprint(stream.PaidThrough))
}

func ComputeMeteredUsageHash(usage MeteredUsage) string {
	usage = canonicalMeteredUsage(usage)
	return servicesHashParts("aetra-services-payment-metered-usage-v1", usage.MeterID, usage.ServiceID, usage.CallID, usage.UsageReceipt.ReceiptHash, usage.AmountDue, fmt.Sprint(usage.RecordedHeight))
}

func ComputePaymentSettlementHash(settlement PaymentSettlement) string {
	settlement = canonicalPaymentSettlement(settlement)
	return servicesHashParts("aetra-services-payment-settlement-v1", settlement.CallID, settlement.ServiceID, settlement.EnvelopeHash, settlement.QuoteHash, settlement.AmountSettled, settlement.Denom, string(settlement.Status), string(settlement.FailurePolicy), fmt.Sprint(settlement.SettlementHeight))
}

func ComputeServicePaymentStateRoot(state ServicePaymentState) string {
	parts := []string{"aetra-services-payment-state-root-v1", fmt.Sprint(state.Height)}
	for _, model := range normalizeServicePaymentModels(state.Models) {
		parts = append(parts, model.ModelHash)
	}
	for _, escrow := range normalizeServiceEscrows(state.Escrows) {
		parts = append(parts, escrow.LockHash)
	}
	for _, stream := range normalizePaymentStreams(state.Streams) {
		parts = append(parts, stream.StreamHash)
	}
	for _, meter := range normalizeMeteredUsages(state.Meters) {
		parts = append(parts, meter.UsageHash)
	}
	for _, settlement := range normalizePaymentSettlements(state.Settlements) {
		parts = append(parts, settlement.SettlementHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServicePaymentProofHash(proof ServicePaymentProof) string {
	return servicesHashParts("aetra-services-payment-proof-v1", proof.Key, proof.ValueHash, proof.StateRoot, fmt.Sprint(proof.Height))
}

func ComputeFinancialZonePaymentRouteHash(route FinancialZonePaymentRoute) string {
	return servicesHashParts("aetra-services-financial-route-v1", route.RouteID, route.ServiceID, route.Payer, route.Denom, route.Amount, route.BankKeeper, route.FinancialZone)
}

func canonicalServicePaymentModel(model ServicePaymentModel) ServicePaymentModel {
	model.ServiceID = strings.TrimSpace(model.ServiceID)
	model.SupportedDenoms = normalizeDescriptorTokens(model.SupportedDenoms)
	model.DefaultDenom = strings.TrimSpace(model.DefaultDenom)
	model.UnitAmount = strings.TrimSpace(model.UnitAmount)
	model.MaxAmountOptional = strings.TrimSpace(model.MaxAmountOptional)
	model.ModelHash = strings.ToLower(strings.TrimSpace(model.ModelHash))
	return model
}

func canonicalServiceEscrow(escrow ServiceEscrow) ServiceEscrow {
	escrow.EscrowID = strings.TrimSpace(escrow.EscrowID)
	escrow.ServiceID = strings.TrimSpace(escrow.ServiceID)
	escrow.Payer = strings.TrimSpace(escrow.Payer)
	escrow.Denom = strings.TrimSpace(escrow.Denom)
	escrow.Amount = strings.TrimSpace(escrow.Amount)
	escrow.LockHash = strings.ToLower(strings.TrimSpace(escrow.LockHash))
	return escrow
}

func canonicalPaymentStream(stream PaymentStream) PaymentStream {
	stream.StreamID = strings.TrimSpace(stream.StreamID)
	stream.ServiceID = strings.TrimSpace(stream.ServiceID)
	stream.Payer = strings.TrimSpace(stream.Payer)
	stream.Denom = strings.TrimSpace(stream.Denom)
	stream.RatePerHeight = strings.TrimSpace(stream.RatePerHeight)
	stream.StreamHash = strings.ToLower(strings.TrimSpace(stream.StreamHash))
	return stream
}

func canonicalMeteredUsage(usage MeteredUsage) MeteredUsage {
	usage.MeterID = strings.TrimSpace(usage.MeterID)
	usage.ServiceID = strings.TrimSpace(usage.ServiceID)
	usage.CallID = strings.ToLower(strings.TrimSpace(usage.CallID))
	usage.AmountDue = strings.TrimSpace(usage.AmountDue)
	usage.UsageHash = strings.ToLower(strings.TrimSpace(usage.UsageHash))
	return usage
}

func canonicalPaymentSettlement(settlement PaymentSettlement) PaymentSettlement {
	settlement.CallID = strings.ToLower(strings.TrimSpace(settlement.CallID))
	settlement.ServiceID = strings.TrimSpace(settlement.ServiceID)
	settlement.EnvelopeHash = strings.ToLower(strings.TrimSpace(settlement.EnvelopeHash))
	settlement.QuoteHash = strings.ToLower(strings.TrimSpace(settlement.QuoteHash))
	settlement.AmountSettled = strings.TrimSpace(settlement.AmountSettled)
	settlement.Denom = strings.TrimSpace(settlement.Denom)
	settlement.SettlementHash = strings.ToLower(strings.TrimSpace(settlement.SettlementHash))
	return settlement
}

func isProtocolNativePaymentDescriptor(descriptor ServiceDescriptor) bool {
	if descriptor.ServiceType == coretypes.ServiceTypeOnChain {
		return true
	}
	switch descriptor.ZoneID {
	case coretypes.ZoneIDAetraCore, coretypes.ZoneIDFinancial, coretypes.ZoneIDIdentity:
		return true
	default:
		return false
	}
}

func normalizeServicePaymentModels(models []ServicePaymentModel) []ServicePaymentModel {
	out := append([]ServicePaymentModel(nil), models...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ServiceID < out[j].ServiceID })
	return out
}

func normalizeServiceEscrows(escrows []ServiceEscrow) []ServiceEscrow {
	out := append([]ServiceEscrow(nil), escrows...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].EscrowID < out[j].EscrowID })
	return out
}

func normalizePaymentStreams(streams []PaymentStream) []PaymentStream {
	out := append([]PaymentStream(nil), streams...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].StreamID < out[j].StreamID })
	return out
}

func normalizeMeteredUsages(usages []MeteredUsage) []MeteredUsage {
	out := append([]MeteredUsage(nil), usages...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].MeterID < out[j].MeterID })
	return out
}

func normalizePaymentSettlements(settlements []PaymentSettlement) []PaymentSettlement {
	out := append([]PaymentSettlement(nil), settlements...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CallID < out[j].CallID })
	return out
}

func validateUniquePaymentModels(models []ServicePaymentModel) error {
	seen := map[string]struct{}{}
	for _, model := range models {
		if err := model.Validate(); err != nil {
			return err
		}
		if _, ok := seen[model.ServiceID]; ok {
			return fmt.Errorf("services duplicate payment model %s", model.ServiceID)
		}
		seen[model.ServiceID] = struct{}{}
	}
	return nil
}

func validatePaymentDenoms(denoms []string) error {
	previous := ""
	for _, denom := range denoms {
		if err := validateInterfaceToken("services payment denom", denom); err != nil {
			return err
		}
		if previous != "" && previous >= denom {
			return errors.New("services payment denoms must be sorted and unique")
		}
		previous = denom
	}
	return nil
}

func stringInSet(value string, values []string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func servicePaymentValueHashForKey(state ServicePaymentState, key string) (string, bool) {
	for _, model := range state.Models {
		modelKey, _ := PaymentModelStateKey(model.ServiceID)
		if key == modelKey {
			return model.ModelHash, true
		}
	}
	for _, escrow := range state.Escrows {
		escrowKey, _ := ServiceEscrowStateKey(escrow.EscrowID)
		if key == escrowKey {
			return escrow.LockHash, true
		}
	}
	for _, stream := range state.Streams {
		streamKey, _ := PaymentStreamStateKey(stream.StreamID)
		if key == streamKey {
			return stream.StreamHash, true
		}
	}
	for _, usage := range state.Meters {
		meterKey, _ := MeteredUsageStateKey(usage.MeterID)
		if key == meterKey {
			return usage.UsageHash, true
		}
	}
	for _, settlement := range state.Settlements {
		settlementKey, _ := PaymentSettlementStateKey(settlement.CallID)
		if key == settlementKey {
			return settlement.SettlementHash, true
		}
	}
	return "", false
}
