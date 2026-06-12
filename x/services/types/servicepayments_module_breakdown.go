package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServicePaymentsStateObject string
type XServicePaymentsMessageName string
type XServicePaymentsQueryName string
type XServicePaymentsFailureMode string
type XServicePaymentsIntegrationPoint string

const (
	XServicePaymentsStatePaymentModel	XServicePaymentsStateObject	= "PaymentModel"
	XServicePaymentsStatePaymentEnvelope	XServicePaymentsStateObject	= "PaymentEnvelope"
	XServicePaymentsStateServiceEscrow	XServicePaymentsStateObject	= "ServiceEscrow"
	XServicePaymentsStatePaymentStream	XServicePaymentsStateObject	= "PaymentStream"
	XServicePaymentsStateMeteredUsage	XServicePaymentsStateObject	= "MeteredUsage"
	XServicePaymentsStatePaymentSettlement	XServicePaymentsStateObject	= "PaymentSettlement"

	XServicePaymentsMsgSetServicePaymentModel	XServicePaymentsMessageName	= "MsgSetServicePaymentModel"
	XServicePaymentsMsgCreateServiceEscrow		XServicePaymentsMessageName	= "MsgCreateServiceEscrow"
	XServicePaymentsMsgSettleServiceEscrow		XServicePaymentsMessageName	= "MsgSettleServiceEscrow"
	XServicePaymentsMsgOpenPaymentStream		XServicePaymentsMessageName	= "MsgOpenPaymentStream"
	XServicePaymentsMsgClosePaymentStream		XServicePaymentsMessageName	= "MsgClosePaymentStream"
	XServicePaymentsMsgSubmitMeteredUsage		XServicePaymentsMessageName	= "MsgSubmitMeteredUsage"

	XServicePaymentsQueryPaymentModel	XServicePaymentsQueryName	= "QueryPaymentModel"
	XServicePaymentsQueryServiceEscrow	XServicePaymentsQueryName	= "QueryServiceEscrow"
	XServicePaymentsQueryPaymentStream	XServicePaymentsQueryName	= "QueryPaymentStream"
	XServicePaymentsQueryMeteredUsage	XServicePaymentsQueryName	= "QueryMeteredUsage"
	XServicePaymentsQueryPaymentSettlement	XServicePaymentsQueryName	= "QueryPaymentSettlement"

	XServicePaymentsFailureEscrowUnderfunded		XServicePaymentsFailureMode	= "escrow_underfunded"
	XServicePaymentsFailureUsageReceiptInvalid		XServicePaymentsFailureMode	= "usage_receipt_invalid"
	XServicePaymentsFailureStreamSettlementExceedsMax	XServicePaymentsFailureMode	= "stream_settlement_exceeds_maximum"
	XServicePaymentsFailureModelChangedAfterSigning		XServicePaymentsFailureMode	= "payment_model_changed_after_call_signing"

	XServicePaymentsIntegrationBankOrFinancialZone	XServicePaymentsIntegrationPoint	= "bank_or_financial_zone"
	XServicePaymentsIntegrationServices		XServicePaymentsIntegrationPoint	= "x/services"
	XServicePaymentsIntegrationServiceCalls		XServicePaymentsIntegrationPoint	= "x/servicecalls"
	XServicePaymentsIntegrationPayments		XServicePaymentsIntegrationPoint	= "x/payments"
)

type XServicePaymentsFailureCoverage struct {
	Mode	XServicePaymentsFailureMode
	Guard	string
	Scope	string
}

type XServicePaymentsModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServicePaymentsStateObject
	Messages		[]XServicePaymentsMessageName
	Queries			[]XServicePaymentsQueryName
	FailureModes		[]XServicePaymentsFailureCoverage
	IntegrationPoints	[]XServicePaymentsIntegrationPoint
	BreakdownHash		string
}

type MsgSetServicePaymentModel struct {
	Authority	string
	Model		ServicePaymentModel
	MessageHash	string
}

type MsgCreateServiceEscrow struct {
	Authority	string
	Envelope	PaymentEnvelope
	Escrow		ServiceEscrow
	MessageHash	string
}

type MsgSettleServiceEscrow struct {
	Authority	string
	EscrowID	string
	Settlement	PaymentSettlement
	MessageHash	string
}

type MsgOpenPaymentStream struct {
	Authority	string
	Envelope	PaymentEnvelope
	Stream		PaymentStream
	MessageHash	string
}

type MsgClosePaymentStream struct {
	Authority	string
	StreamID	string
	CloseHeight	uint64
	AmountSettled	string
	MessageHash	string
}

type MsgSubmitMeteredUsage struct {
	Authority	string
	Usage		MeteredUsage
	MessageHash	string
}

type QueryServiceEscrow struct {
	EscrowID string
}

type QueryServiceEscrowResponse struct {
	Escrow	ServiceEscrow
	Found	bool
}

type QueryPaymentStream struct {
	StreamID string
}

type QueryPaymentStreamResponse struct {
	Stream	PaymentStream
	Found	bool
}

type QueryMeteredUsage struct {
	MeterID string
}

type QueryMeteredUsageResponse struct {
	Usage	MeteredUsage
	Found	bool
}

type QueryPaymentSettlement struct {
	CallID string
}

type QueryPaymentSettlementResponse struct {
	Settlement	PaymentSettlement
	Found		bool
}

type ServicePaymentSignedModelSnapshot struct {
	ServiceID	string
	CallID		string
	ModelHash	string
	SignedHeight	uint64
	SnapshotHash	string
}

func DefaultXServicePaymentsModuleBreakdown() (XServicePaymentsModuleBreakdown, error) {
	breakdown := XServicePaymentsModuleBreakdown{
		ModulePath:	ServiceModulePayments,
		Purpose: []string{
			"define_payment_models",
			"escrow_settlement",
			"metered_usage",
			"payment_streams",
			"service_payment_settlement",
		},
		StateObjects: []XServicePaymentsStateObject{
			XServicePaymentsStatePaymentModel,
			XServicePaymentsStatePaymentEnvelope,
			XServicePaymentsStateServiceEscrow,
			XServicePaymentsStatePaymentStream,
			XServicePaymentsStateMeteredUsage,
			XServicePaymentsStatePaymentSettlement,
		},
		Messages: []XServicePaymentsMessageName{
			XServicePaymentsMsgSetServicePaymentModel,
			XServicePaymentsMsgCreateServiceEscrow,
			XServicePaymentsMsgSettleServiceEscrow,
			XServicePaymentsMsgOpenPaymentStream,
			XServicePaymentsMsgClosePaymentStream,
			XServicePaymentsMsgSubmitMeteredUsage,
		},
		Queries: []XServicePaymentsQueryName{
			XServicePaymentsQueryPaymentModel,
			XServicePaymentsQueryServiceEscrow,
			XServicePaymentsQueryPaymentStream,
			XServicePaymentsQueryMeteredUsage,
			XServicePaymentsQueryPaymentSettlement,
		},
		FailureModes: []XServicePaymentsFailureCoverage{
			newXServicePaymentsFailureCoverage(XServicePaymentsFailureEscrowUnderfunded, "ValidateServiceEscrowFunding", ServicePaymentEscrowPrefix),
			newXServicePaymentsFailureCoverage(XServicePaymentsFailureUsageReceiptInvalid, "MeteredUsage.Validate", ServicePaymentMeterPrefix),
			newXServicePaymentsFailureCoverage(XServicePaymentsFailureStreamSettlementExceedsMax, "ValidatePaymentStreamSettlementLimit", ServicePaymentStreamPrefix),
			newXServicePaymentsFailureCoverage(XServicePaymentsFailureModelChangedAfterSigning, "ValidatePaymentModelSnapshotForCall", ServicePaymentModelPrefix),
		},
		IntegrationPoints: []XServicePaymentsIntegrationPoint{
			XServicePaymentsIntegrationBankOrFinancialZone,
			XServicePaymentsIntegrationServices,
			XServicePaymentsIntegrationServiceCalls,
			XServicePaymentsIntegrationPayments,
		},
	}
	return NewXServicePaymentsModuleBreakdown(breakdown)
}

func NewXServicePaymentsModuleBreakdown(breakdown XServicePaymentsModuleBreakdown) (XServicePaymentsModuleBreakdown, error) {
	breakdown = canonicalXServicePaymentsModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServicePaymentsModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServicePaymentsModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func NewMsgSetServicePaymentModel(authority string, model ServicePaymentModel) (MsgSetServicePaymentModel, error) {
	msg := MsgSetServicePaymentModel{Authority: strings.TrimSpace(authority), Model: model}
	msg.MessageHash = ComputeMsgSetServicePaymentModelHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgCreateServiceEscrow(authority string, envelope PaymentEnvelope, height uint64) (MsgCreateServiceEscrow, error) {
	escrow, err := CreateServiceEscrowFromEnvelope(envelope, height)
	if err != nil {
		return MsgCreateServiceEscrow{}, err
	}
	msg := MsgCreateServiceEscrow{Authority: strings.TrimSpace(authority), Envelope: envelope, Escrow: escrow}
	msg.MessageHash = ComputeMsgCreateServiceEscrowHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgSettleServiceEscrow(authority string, escrow ServiceEscrow, settlement PaymentSettlement) (MsgSettleServiceEscrow, error) {
	msg := MsgSettleServiceEscrow{Authority: strings.TrimSpace(authority), EscrowID: escrow.EscrowID, Settlement: settlement}
	msg.MessageHash = ComputeMsgSettleServiceEscrowHash(msg)
	return msg, msg.ValidateForEscrow(escrow)
}

func NewMsgOpenPaymentStream(authority string, envelope PaymentEnvelope, startHeight, endHeight uint64) (MsgOpenPaymentStream, error) {
	stream, err := CreatePaymentStreamFromEnvelope(envelope, startHeight, endHeight)
	if err != nil {
		return MsgOpenPaymentStream{}, err
	}
	msg := MsgOpenPaymentStream{Authority: strings.TrimSpace(authority), Envelope: envelope, Stream: stream}
	msg.MessageHash = ComputeMsgOpenPaymentStreamHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgClosePaymentStream(authority string, stream PaymentStream, closeHeight uint64, amountSettled string) (MsgClosePaymentStream, error) {
	msg := MsgClosePaymentStream{Authority: strings.TrimSpace(authority), StreamID: stream.StreamID, CloseHeight: closeHeight, AmountSettled: strings.TrimSpace(amountSettled)}
	msg.MessageHash = ComputeMsgClosePaymentStreamHash(msg)
	return msg, msg.ValidateForStream(stream)
}

func NewMsgSubmitMeteredUsage(authority string, usage MeteredUsage) (MsgSubmitMeteredUsage, error) {
	msg := MsgSubmitMeteredUsage{Authority: strings.TrimSpace(authority), Usage: usage}
	msg.MessageHash = ComputeMsgSubmitMeteredUsageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewServicePaymentSignedModelSnapshot(model ServicePaymentModel, call UnifiedServiceCall, signedHeight uint64) (ServicePaymentSignedModelSnapshot, error) {
	if err := model.Validate(); err != nil {
		return ServicePaymentSignedModelSnapshot{}, err
	}
	if signedHeight == 0 {
		return ServicePaymentSignedModelSnapshot{}, errors.New("x/servicepayments signed height is required")
	}
	snapshot := ServicePaymentSignedModelSnapshot{ServiceID: model.ServiceID, CallID: call.CallID, ModelHash: model.ModelHash, SignedHeight: signedHeight}
	snapshot.SnapshotHash = ComputeServicePaymentSignedModelSnapshotHash(snapshot)
	return snapshot, snapshot.Validate()
}

func ValidateServiceEscrowFunding(escrow ServiceEscrow, amountDue string) error {
	if err := escrow.Validate(); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("x/servicepayments amount due", amountDue); err != nil {
		return err
	}
	if comparePaymentAmounts(escrow.Amount, amountDue) < 0 {
		return errors.New("x/servicepayments escrow underfunded")
	}
	return nil
}

func ValidatePaymentStreamSettlementLimit(stream PaymentStream, amountSettled string, closeHeight uint64) error {
	if err := stream.Validate(); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("x/servicepayments stream settled amount", amountSettled); err != nil {
		return err
	}
	if closeHeight < stream.StartHeight || closeHeight > stream.EndHeight {
		return errors.New("x/servicepayments stream close height outside range")
	}
	elapsed := closeHeight - stream.StartHeight
	maximum := multiplyPaymentAmount(stream.RatePerHeight, elapsed)
	if elapsed == 0 {
		maximum = "0"
	}
	if comparePaymentAmounts(amountSettled, maximum) > 0 {
		return errors.New("x/servicepayments stream settlement exceeds maximum")
	}
	return nil
}

func ValidatePaymentModelSnapshotForCall(snapshot ServicePaymentSignedModelSnapshot, current ServicePaymentModel) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if err := current.Validate(); err != nil {
		return err
	}
	if snapshot.ServiceID != current.ServiceID {
		return errors.New("x/servicepayments signed model service mismatch")
	}
	if snapshot.ModelHash != current.ModelHash {
		return errors.New("x/servicepayments payment model changed after call signing")
	}
	return nil
}

func QueryServiceEscrowFromState(state ServicePaymentState, q QueryServiceEscrow) (QueryServiceEscrowResponse, error) {
	if err := validateInterfaceToken("x/servicepayments escrow query id", q.EscrowID); err != nil {
		return QueryServiceEscrowResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryServiceEscrowResponse{}, err
	}
	for _, escrow := range state.Escrows {
		if escrow.EscrowID == q.EscrowID {
			return QueryServiceEscrowResponse{Escrow: escrow, Found: true}, nil
		}
	}
	return QueryServiceEscrowResponse{Found: false}, nil
}

func QueryPaymentStreamFromState(state ServicePaymentState, q QueryPaymentStream) (QueryPaymentStreamResponse, error) {
	if err := validateInterfaceToken("x/servicepayments stream query id", q.StreamID); err != nil {
		return QueryPaymentStreamResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryPaymentStreamResponse{}, err
	}
	for _, stream := range state.Streams {
		if stream.StreamID == q.StreamID {
			return QueryPaymentStreamResponse{Stream: stream, Found: true}, nil
		}
	}
	return QueryPaymentStreamResponse{Found: false}, nil
}

func QueryMeteredUsageFromState(state ServicePaymentState, q QueryMeteredUsage) (QueryMeteredUsageResponse, error) {
	if err := validateInterfaceToken("x/servicepayments meter query id", q.MeterID); err != nil {
		return QueryMeteredUsageResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryMeteredUsageResponse{}, err
	}
	for _, usage := range state.Meters {
		if usage.MeterID == q.MeterID {
			return QueryMeteredUsageResponse{Usage: usage, Found: true}, nil
		}
	}
	return QueryMeteredUsageResponse{Found: false}, nil
}

func QueryPaymentSettlementFromState(state ServicePaymentState, q QueryPaymentSettlement) (QueryPaymentSettlementResponse, error) {
	if err := coretypes.ValidateHash("x/servicepayments settlement query call id", q.CallID); err != nil {
		return QueryPaymentSettlementResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryPaymentSettlementResponse{}, err
	}
	for _, settlement := range state.Settlements {
		if settlement.CallID == q.CallID {
			return QueryPaymentSettlementResponse{Settlement: settlement, Found: true}, nil
		}
	}
	return QueryPaymentSettlementResponse{Found: false}, nil
}

func (breakdown XServicePaymentsModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModulePayments {
		return errors.New("x/servicepayments breakdown must describe x/servicepayments")
	}
	if err := validateSortedTokens("x/servicepayments purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServicePaymentsStateObjects(breakdown.StateObjects); err != nil {
		return err
	}
	if err := validateXServicePaymentsMessages(breakdown.Messages); err != nil {
		return err
	}
	if err := validateXServicePaymentsQueries(breakdown.Queries); err != nil {
		return err
	}
	if err := validateXServicePaymentsFailureCoverage(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServicePaymentsIntegrationPoints(breakdown.IntegrationPoints); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/servicepayments breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServicePaymentsModuleBreakdown) Validate() error {
	breakdown = canonicalXServicePaymentsModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/servicepayments breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeXServicePaymentsModuleBreakdownHash(breakdown) {
		return errors.New("x/servicepayments breakdown hash mismatch")
	}
	return nil
}

func (coverage XServicePaymentsFailureCoverage) Validate() error {
	if !IsXServicePaymentsFailureMode(coverage.Mode) {
		return fmt.Errorf("x/servicepayments unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/servicepayments failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.Scope + "/_") {
		return fmt.Errorf("x/servicepayments failure scope %s must be services store key", coverage.Scope)
	}
	return nil
}

func (msg MsgSetServicePaymentModel) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicepayments set model authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Model.Validate(); err != nil {
		return err
	}
	return validateServicePaymentMsgHash("set model", msg.MessageHash, ComputeMsgSetServicePaymentModelHash(msg))
}

func (msg MsgCreateServiceEscrow) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicepayments create escrow authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Envelope.Validate(); err != nil {
		return err
	}
	if err := msg.Escrow.Validate(); err != nil {
		return err
	}
	if msg.Envelope.EscrowIDOptional != msg.Escrow.EscrowID || msg.Envelope.EnvelopeHash == "" {
		return errors.New("x/servicepayments escrow envelope mismatch")
	}
	return validateServicePaymentMsgHash("create escrow", msg.MessageHash, ComputeMsgCreateServiceEscrowHash(msg))
}

func (msg MsgSettleServiceEscrow) ValidateForEscrow(escrow ServiceEscrow) error {
	if err := validateInterfaceToken("x/servicepayments settle escrow authority", msg.Authority); err != nil {
		return err
	}
	if msg.EscrowID != escrow.EscrowID {
		return errors.New("x/servicepayments settle escrow id mismatch")
	}
	if err := msg.Settlement.Validate(); err != nil {
		return err
	}
	if err := ValidateServiceEscrowFunding(escrow, msg.Settlement.AmountSettled); err != nil {
		return err
	}
	return validateServicePaymentMsgHash("settle escrow", msg.MessageHash, ComputeMsgSettleServiceEscrowHash(msg))
}

func (msg MsgOpenPaymentStream) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicepayments open stream authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Envelope.Validate(); err != nil {
		return err
	}
	if err := msg.Stream.Validate(); err != nil {
		return err
	}
	if msg.Envelope.StreamIDOptional != msg.Stream.StreamID {
		return errors.New("x/servicepayments stream envelope mismatch")
	}
	return validateServicePaymentMsgHash("open stream", msg.MessageHash, ComputeMsgOpenPaymentStreamHash(msg))
}

func (msg MsgClosePaymentStream) ValidateForStream(stream PaymentStream) error {
	if err := validateInterfaceToken("x/servicepayments close stream authority", msg.Authority); err != nil {
		return err
	}
	if msg.StreamID != stream.StreamID {
		return errors.New("x/servicepayments close stream id mismatch")
	}
	if err := ValidatePaymentStreamSettlementLimit(stream, msg.AmountSettled, msg.CloseHeight); err != nil {
		return err
	}
	return validateServicePaymentMsgHash("close stream", msg.MessageHash, ComputeMsgClosePaymentStreamHash(msg))
}

func (msg MsgSubmitMeteredUsage) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicepayments metered usage authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Usage.Validate(); err != nil {
		return err
	}
	return validateServicePaymentMsgHash("submit usage", msg.MessageHash, ComputeMsgSubmitMeteredUsageHash(msg))
}

func (snapshot ServicePaymentSignedModelSnapshot) Validate() error {
	if err := validateInterfaceToken("x/servicepayments snapshot service id", snapshot.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicepayments snapshot call id", snapshot.CallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicepayments snapshot model hash", snapshot.ModelHash); err != nil {
		return err
	}
	if snapshot.SignedHeight == 0 {
		return errors.New("x/servicepayments snapshot signed height is required")
	}
	if err := coretypes.ValidateHash("x/servicepayments snapshot hash", snapshot.SnapshotHash); err != nil {
		return err
	}
	if snapshot.SnapshotHash != ComputeServicePaymentSignedModelSnapshotHash(snapshot) {
		return errors.New("x/servicepayments snapshot hash mismatch")
	}
	return nil
}

func IsXServicePaymentsStateObject(object XServicePaymentsStateObject) bool {
	switch object {
	case XServicePaymentsStatePaymentModel, XServicePaymentsStatePaymentEnvelope, XServicePaymentsStateServiceEscrow, XServicePaymentsStatePaymentStream, XServicePaymentsStateMeteredUsage, XServicePaymentsStatePaymentSettlement:
		return true
	default:
		return false
	}
}

func IsXServicePaymentsMessageName(message XServicePaymentsMessageName) bool {
	switch message {
	case XServicePaymentsMsgSetServicePaymentModel, XServicePaymentsMsgCreateServiceEscrow, XServicePaymentsMsgSettleServiceEscrow, XServicePaymentsMsgOpenPaymentStream, XServicePaymentsMsgClosePaymentStream, XServicePaymentsMsgSubmitMeteredUsage:
		return true
	default:
		return false
	}
}

func IsXServicePaymentsQueryName(query XServicePaymentsQueryName) bool {
	switch query {
	case XServicePaymentsQueryPaymentModel, XServicePaymentsQueryServiceEscrow, XServicePaymentsQueryPaymentStream, XServicePaymentsQueryMeteredUsage, XServicePaymentsQueryPaymentSettlement:
		return true
	default:
		return false
	}
}

func IsXServicePaymentsFailureMode(mode XServicePaymentsFailureMode) bool {
	switch mode {
	case XServicePaymentsFailureEscrowUnderfunded, XServicePaymentsFailureUsageReceiptInvalid, XServicePaymentsFailureStreamSettlementExceedsMax, XServicePaymentsFailureModelChangedAfterSigning:
		return true
	default:
		return false
	}
}

func IsXServicePaymentsIntegrationPoint(point XServicePaymentsIntegrationPoint) bool {
	switch point {
	case XServicePaymentsIntegrationBankOrFinancialZone, XServicePaymentsIntegrationServices, XServicePaymentsIntegrationServiceCalls, XServicePaymentsIntegrationPayments:
		return true
	default:
		return false
	}
}

func ComputeXServicePaymentsModuleBreakdownHash(breakdown XServicePaymentsModuleBreakdown) string {
	breakdown = canonicalXServicePaymentsModuleBreakdown(breakdown)
	parts := []string{"aetra-x-servicepayments-module-breakdown-v1", breakdown.ModulePath, "purpose", fmt.Sprint(len(breakdown.Purpose))}
	parts = append(parts, breakdown.Purpose...)
	parts = append(parts, "state", fmt.Sprint(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		parts = append(parts, string(object))
	}
	parts = append(parts, "messages", fmt.Sprint(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		parts = append(parts, string(message))
	}
	parts = append(parts, "queries", fmt.Sprint(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		parts = append(parts, string(query))
	}
	parts = append(parts, "failures", fmt.Sprint(len(breakdown.FailureModes)))
	for _, coverage := range breakdown.FailureModes {
		parts = append(parts, string(coverage.Mode), coverage.Guard, coverage.Scope)
	}
	parts = append(parts, "integrations", fmt.Sprint(len(breakdown.IntegrationPoints)))
	for _, point := range breakdown.IntegrationPoints {
		parts = append(parts, string(point))
	}
	return servicesHashParts(parts...)
}

func ComputeMsgSetServicePaymentModelHash(msg MsgSetServicePaymentModel) string {
	return servicesHashParts("aetra-x-servicepayments-msg-set-model-v1", msg.Authority, msg.Model.ServiceID, msg.Model.ModelHash)
}

func ComputeMsgCreateServiceEscrowHash(msg MsgCreateServiceEscrow) string {
	return servicesHashParts("aetra-x-servicepayments-msg-create-escrow-v1", msg.Authority, msg.Envelope.EnvelopeHash, msg.Escrow.LockHash)
}

func ComputeMsgSettleServiceEscrowHash(msg MsgSettleServiceEscrow) string {
	return servicesHashParts("aetra-x-servicepayments-msg-settle-escrow-v1", msg.Authority, msg.EscrowID, msg.Settlement.SettlementHash)
}

func ComputeMsgOpenPaymentStreamHash(msg MsgOpenPaymentStream) string {
	return servicesHashParts("aetra-x-servicepayments-msg-open-stream-v1", msg.Authority, msg.Envelope.EnvelopeHash, msg.Stream.StreamHash)
}

func ComputeMsgClosePaymentStreamHash(msg MsgClosePaymentStream) string {
	return servicesHashParts("aetra-x-servicepayments-msg-close-stream-v1", msg.Authority, msg.StreamID, fmt.Sprint(msg.CloseHeight), msg.AmountSettled)
}

func ComputeMsgSubmitMeteredUsageHash(msg MsgSubmitMeteredUsage) string {
	return servicesHashParts("aetra-x-servicepayments-msg-submit-metered-v1", msg.Authority, msg.Usage.UsageHash)
}

func ComputeServicePaymentSignedModelSnapshotHash(snapshot ServicePaymentSignedModelSnapshot) string {
	return servicesHashParts("aetra-x-servicepayments-model-snapshot-v1", snapshot.ServiceID, snapshot.CallID, snapshot.ModelHash, fmt.Sprint(snapshot.SignedHeight))
}

func validateServicePaymentMsgHash(label string, actual string, expected string) error {
	if err := coretypes.ValidateHash("x/servicepayments "+label+" message hash", actual); err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("x/servicepayments %s message hash mismatch", label)
	}
	return nil
}

func newXServicePaymentsFailureCoverage(mode XServicePaymentsFailureMode, guard, scope string) XServicePaymentsFailureCoverage {
	return XServicePaymentsFailureCoverage{Mode: mode, Guard: guard, Scope: scope}
}

func canonicalXServicePaymentsModuleBreakdown(breakdown XServicePaymentsModuleBreakdown) XServicePaymentsModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedStrings(breakdown.Purpose)
	breakdown.StateObjects = sortedXServicePaymentsStateObjects(breakdown.StateObjects)
	breakdown.Messages = sortedXServicePaymentsMessages(breakdown.Messages)
	breakdown.Queries = sortedXServicePaymentsQueries(breakdown.Queries)
	breakdown.FailureModes = sortedXServicePaymentsFailureCoverage(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedXServicePaymentsIntegrationPoints(breakdown.IntegrationPoints)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateXServicePaymentsStateObjects(objects []XServicePaymentsStateObject) error {
	required := []XServicePaymentsStateObject{XServicePaymentsStatePaymentModel, XServicePaymentsStatePaymentEnvelope, XServicePaymentsStateServiceEscrow, XServicePaymentsStatePaymentStream, XServicePaymentsStateMeteredUsage, XServicePaymentsStatePaymentSettlement}
	return validateXServicePaymentsEnumSet("state object", objects, required, IsXServicePaymentsStateObject)
}

func validateXServicePaymentsMessages(messages []XServicePaymentsMessageName) error {
	required := []XServicePaymentsMessageName{XServicePaymentsMsgSetServicePaymentModel, XServicePaymentsMsgCreateServiceEscrow, XServicePaymentsMsgSettleServiceEscrow, XServicePaymentsMsgOpenPaymentStream, XServicePaymentsMsgClosePaymentStream, XServicePaymentsMsgSubmitMeteredUsage}
	return validateXServicePaymentsEnumSet("message", messages, required, IsXServicePaymentsMessageName)
}

func validateXServicePaymentsQueries(queries []XServicePaymentsQueryName) error {
	required := []XServicePaymentsQueryName{XServicePaymentsQueryPaymentModel, XServicePaymentsQueryServiceEscrow, XServicePaymentsQueryPaymentStream, XServicePaymentsQueryMeteredUsage, XServicePaymentsQueryPaymentSettlement}
	return validateXServicePaymentsEnumSet("query", queries, required, IsXServicePaymentsQueryName)
}

func validateXServicePaymentsFailureCoverage(coverage []XServicePaymentsFailureCoverage) error {
	required := []XServicePaymentsFailureMode{XServicePaymentsFailureEscrowUnderfunded, XServicePaymentsFailureUsageReceiptInvalid, XServicePaymentsFailureStreamSettlementExceedsMax, XServicePaymentsFailureModelChangedAfterSigning}
	if len(coverage) != len(required) {
		return fmt.Errorf("x/servicepayments expected %d failure modes", len(required))
	}
	seen := map[XServicePaymentsFailureMode]struct{}{}
	for _, item := range coverage {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.Mode]; found {
			return fmt.Errorf("x/servicepayments duplicate failure mode %s", item.Mode)
		}
		seen[item.Mode] = struct{}{}
	}
	for _, mode := range required {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("x/servicepayments missing failure mode %s", mode)
		}
	}
	return nil
}

func validateXServicePaymentsIntegrationPoints(points []XServicePaymentsIntegrationPoint) error {
	required := []XServicePaymentsIntegrationPoint{XServicePaymentsIntegrationBankOrFinancialZone, XServicePaymentsIntegrationServices, XServicePaymentsIntegrationServiceCalls, XServicePaymentsIntegrationPayments}
	return validateXServicePaymentsEnumSet("integration", points, required, IsXServicePaymentsIntegrationPoint)
}

func validateXServicePaymentsEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/servicepayments expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/servicepayments unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/servicepayments %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/servicepayments duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/servicepayments missing %s %s", label, value)
		}
	}
	return nil
}

func sortedXServicePaymentsStateObjects(values []XServicePaymentsStateObject) []XServicePaymentsStateObject {
	out := append([]XServicePaymentsStateObject(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicePaymentsMessages(values []XServicePaymentsMessageName) []XServicePaymentsMessageName {
	out := append([]XServicePaymentsMessageName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicePaymentsQueries(values []XServicePaymentsQueryName) []XServicePaymentsQueryName {
	out := append([]XServicePaymentsQueryName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicePaymentsFailureCoverage(values []XServicePaymentsFailureCoverage) []XServicePaymentsFailureCoverage {
	out := append([]XServicePaymentsFailureCoverage(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].Scope = strings.Trim(strings.TrimSpace(out[i].Scope), "/")
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedXServicePaymentsIntegrationPoints(values []XServicePaymentsIntegrationPoint) []XServicePaymentsIntegrationPoint {
	out := append([]XServicePaymentsIntegrationPoint(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
