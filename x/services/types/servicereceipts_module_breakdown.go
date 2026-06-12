package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServiceReceiptsStateObject string
type XServiceReceiptsMessageName string
type XServiceReceiptsQueryName string
type XServiceReceiptsFailureMode string
type XServiceReceiptsIntegrationPoint string
type ServiceReceiptRecordKind string

const (
	XServiceReceiptsStateReceiptParams	XServiceReceiptsStateObject	= "ReceiptParams"
	XServiceReceiptsStateReceiptRecord	XServiceReceiptsStateObject	= "ReceiptRecord"
	XServiceReceiptsStateReceiptRoot	XServiceReceiptsStateObject	= "ReceiptRoot"
	XServiceReceiptsStateReceiptTombstone	XServiceReceiptsStateObject	= "ReceiptTombstone"

	XServiceReceiptsMsgAnchorReceipt	XServiceReceiptsMessageName	= "MsgAnchorReceipt"
	XServiceReceiptsMsgPruneReceipt		XServiceReceiptsMessageName	= "MsgPruneReceipt"

	XServiceReceiptsQueryReceipt		XServiceReceiptsQueryName	= "QueryReceipt"
	XServiceReceiptsQueryReceiptProof	XServiceReceiptsQueryName	= "QueryReceiptProof"
	XServiceReceiptsQueryReceiptRoot	XServiceReceiptsQueryName	= "QueryReceiptRoot"
	XServiceReceiptsQueryReceiptsByService	XServiceReceiptsQueryName	= "QueryReceiptsByService"

	XServiceReceiptsFailureDuplicateReceipt			XServiceReceiptsFailureMode	= "duplicate_receipt"
	XServiceReceiptsFailureMissingExecutedOnChainReceipt	XServiceReceiptsFailureMode	= "missing_receipt_for_executed_on_chain_call"
	XServiceReceiptsFailureReceiptHashMismatch		XServiceReceiptsFailureMode	= "receipt_hash_mismatch"
	XServiceReceiptsFailurePrunedBeforeProofHorizon		XServiceReceiptsFailureMode	= "receipt_pruned_before_proof_horizon"

	XServiceReceiptsIntegrationAllServiceModules	XServiceReceiptsIntegrationPoint	= "all_service_modules"
	XServiceReceiptsIntegrationProofRegistry	XServiceReceiptsIntegrationPoint	= "proof_registry"
	XServiceReceiptsIntegrationStoreV2		XServiceReceiptsIntegrationPoint	= "store_v2"

	ServiceReceiptRecordKindCall		ServiceReceiptRecordKind	= "call"
	ServiceReceiptRecordKindPayment		ServiceReceiptRecordKind	= "payment"
	ServiceReceiptRecordKindProvider	ServiceReceiptRecordKind	= "provider"
	ServiceReceiptRecordKindService		ServiceReceiptRecordKind	= "service"
	ServiceReceiptRecordKindStorage		ServiceReceiptRecordKind	= "storage"
)

type ReceiptTombstone = ServiceReceiptTombstone

type XServiceReceiptsFailureCoverage struct {
	Mode	XServiceReceiptsFailureMode
	Guard	string
	Scope	string
}

type XServiceReceiptsModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServiceReceiptsStateObject
	Messages		[]XServiceReceiptsMessageName
	Queries			[]XServiceReceiptsQueryName
	FailureModes		[]XServiceReceiptsFailureCoverage
	IntegrationPoints	[]XServiceReceiptsIntegrationPoint
	BreakdownHash		string
}

type ReceiptRecord struct {
	ReceiptID		string
	ServiceID		string
	CallID			string
	Kind			ServiceReceiptRecordKind
	Status			ServiceReceiptStatusText
	ReceiptHash		string
	Height			uint64
	RetainUntilHeight	uint64
	RecordHash		string
}

type ReceiptRoot struct {
	RootKind	ServiceReceiptRecordKind
	Height		uint64
	RecordCount	uint64
	RootHash	string
	CommitmentHash	string
}

type ReceiptParams struct {
	ProofHorizon	uint64
	PruneBatchSize	uint64
	ParamsHash	string
}

type MsgAnchorReceipt struct {
	Authority		string
	Receipt			ServiceReceipt
	ExpectedReceiptHash	string
	MessageHash		string
}

type MsgPruneReceipt struct {
	Authority	string
	ReceiptID	string
	CurrentHeight	uint64
	MessageHash	string
}

type QueryReceipt struct {
	ReceiptID string
}

type QueryReceiptResponse struct {
	Record	ReceiptRecord
	Found	bool
}

type QueryReceiptsByService struct {
	ServiceID string
}

type QueryReceiptsByServiceResponse struct {
	Records		[]ReceiptRecord
	ResponseHash	string
}

type QueryReceiptRoot struct {
	RootKind	ServiceReceiptRecordKind
	Height		uint64
}

type QueryReceiptRootResponse struct {
	Root	ReceiptRoot
	Found	bool
}

type QueryReceiptProof struct {
	ReceiptID string
}

type ServiceReceiptProofRecord struct {
	ReceiptID	string
	ServiceID	string
	ReceiptHash	string
	RootHash	string
	ProofHeight	uint64
	ProofHashes	[]string
	ProofHash	string
}

type QueryReceiptProofResponse struct {
	Proof	ServiceReceiptProofRecord
	Found	bool
}

func DefaultXServiceReceiptsModuleBreakdown() (XServiceReceiptsModuleBreakdown, error) {
	breakdown := XServiceReceiptsModuleBreakdown{
		ModulePath:	ServiceModuleReceipts,
		Purpose: []string{
			"proof_queryable_roots",
			"receipt_records",
			"receipt_tombstones",
			"storage_payment_provider_receipts",
		},
		StateObjects: []XServiceReceiptsStateObject{
			XServiceReceiptsStateReceiptParams,
			XServiceReceiptsStateReceiptRecord,
			XServiceReceiptsStateReceiptRoot,
			XServiceReceiptsStateReceiptTombstone,
		},
		Messages: []XServiceReceiptsMessageName{
			XServiceReceiptsMsgAnchorReceipt,
			XServiceReceiptsMsgPruneReceipt,
		},
		Queries: []XServiceReceiptsQueryName{
			XServiceReceiptsQueryReceipt,
			XServiceReceiptsQueryReceiptProof,
			XServiceReceiptsQueryReceiptRoot,
			XServiceReceiptsQueryReceiptsByService,
		},
		FailureModes: []XServiceReceiptsFailureCoverage{
			newXServiceReceiptsFailureCoverage(XServiceReceiptsFailureDuplicateReceipt, "AnchorReceiptRecord", ServiceStoreV2ReceiptPrefix),
			newXServiceReceiptsFailureCoverage(XServiceReceiptsFailureMissingExecutedOnChainReceipt, "ValidateExecutedOnChainCallsHaveReceipts", ServiceStoreV2ReceiptPrefix),
			newXServiceReceiptsFailureCoverage(XServiceReceiptsFailureReceiptHashMismatch, "MsgAnchorReceipt.ValidateBasic", ServiceStoreV2ReceiptPrefix),
			newXServiceReceiptsFailureCoverage(XServiceReceiptsFailurePrunedBeforeProofHorizon, "PruneReceiptRecord", ServiceStoreV2ReceiptPrefix),
		},
		IntegrationPoints: []XServiceReceiptsIntegrationPoint{
			XServiceReceiptsIntegrationAllServiceModules,
			XServiceReceiptsIntegrationProofRegistry,
			XServiceReceiptsIntegrationStoreV2,
		},
	}
	return NewXServiceReceiptsModuleBreakdown(breakdown)
}

func NewXServiceReceiptsModuleBreakdown(breakdown XServiceReceiptsModuleBreakdown) (XServiceReceiptsModuleBreakdown, error) {
	breakdown = canonicalXServiceReceiptsModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServiceReceiptsModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServiceReceiptsModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func NewReceiptParams(proofHorizon, pruneBatchSize uint64) (ReceiptParams, error) {
	params := ReceiptParams{ProofHorizon: proofHorizon, PruneBatchSize: pruneBatchSize}
	params.ParamsHash = ComputeReceiptParamsHash(params)
	return params, params.Validate()
}

func NewReceiptRecordFromServiceReceipt(receipt ServiceReceipt, kind ServiceReceiptRecordKind, proofHorizon uint64) (ReceiptRecord, error) {
	if err := receipt.Validate(); err != nil {
		return ReceiptRecord{}, err
	}
	if proofHorizon == 0 {
		return ReceiptRecord{}, errors.New("x/servicereceipts proof horizon must be positive")
	}
	status, err := CanonicalServiceReceiptStatus(receipt.Status)
	if err != nil {
		return ReceiptRecord{}, err
	}
	record := ReceiptRecord{
		ReceiptID:		receipt.CallID,
		ServiceID:		receipt.ServiceID,
		CallID:			receipt.CallID,
		Kind:			kind,
		Status:			status,
		ReceiptHash:		receipt.ReceiptHash,
		Height:			receipt.AnchoredHeight,
		RetainUntilHeight:	receipt.AnchoredHeight + proofHorizon,
	}
	record.RecordHash = ComputeReceiptRecordHash(record)
	return record, record.Validate()
}

func BuildReceiptRoot(kind ServiceReceiptRecordKind, height uint64, records []ReceiptRecord) (ReceiptRoot, error) {
	ordered := cloneReceiptRecords(records)
	sortReceiptRecords(ordered)
	if err := validateReceiptRecords(ordered); err != nil {
		return ReceiptRoot{}, err
	}
	root := ReceiptRoot{
		RootKind:	kind,
		Height:		height,
		RecordCount:	uint64(len(ordered)),
		RootHash:	ComputeReceiptRecordRootHash(kind, ordered),
	}
	root.CommitmentHash = ComputeReceiptRootCommitmentHash(root)
	return root, root.Validate()
}

func NewMsgAnchorReceipt(authority string, receipt ServiceReceipt, expectedReceiptHash string) (MsgAnchorReceipt, error) {
	msg := MsgAnchorReceipt{
		Authority:		strings.TrimSpace(authority),
		Receipt:		receipt,
		ExpectedReceiptHash:	strings.ToLower(strings.TrimSpace(expectedReceiptHash)),
	}
	msg.MessageHash = ComputeMsgAnchorReceiptHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgPruneReceipt(authority string, record ReceiptRecord, currentHeight uint64) (MsgPruneReceipt, error) {
	msg := MsgPruneReceipt{
		Authority:	strings.TrimSpace(authority),
		ReceiptID:	strings.ToLower(strings.TrimSpace(record.ReceiptID)),
		CurrentHeight:	currentHeight,
	}
	msg.MessageHash = ComputeMsgPruneReceiptHash(msg)
	return msg, msg.ValidateForRecord(record)
}

func AnchorReceiptRecord(records []ReceiptRecord, msg MsgAnchorReceipt, kind ServiceReceiptRecordKind, proofHorizon uint64) ([]ReceiptRecord, ReceiptRecord, error) {
	if err := validateReceiptRecords(cloneReceiptRecords(records)); err != nil {
		return nil, ReceiptRecord{}, err
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, ReceiptRecord{}, err
	}
	record, err := NewReceiptRecordFromServiceReceipt(msg.Receipt, kind, proofHorizon)
	if err != nil {
		return nil, ReceiptRecord{}, err
	}
	for _, existing := range records {
		if existing.ReceiptID == record.ReceiptID {
			return nil, ReceiptRecord{}, errors.New("x/servicereceipts duplicate receipt")
		}
	}
	next := append(cloneReceiptRecords(records), record)
	sortReceiptRecords(next)
	return next, record, validateReceiptRecords(next)
}

func PruneReceiptRecord(records []ReceiptRecord, params ReceiptParams, msg MsgPruneReceipt) ([]ReceiptRecord, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if err := validateReceiptRecords(cloneReceiptRecords(records)); err != nil {
		return nil, err
	}
	found := false
	next := make([]ReceiptRecord, 0, len(records))
	for _, record := range records {
		if record.ReceiptID != msg.ReceiptID {
			next = append(next, record)
			continue
		}
		found = true
		if err := msg.ValidateForRecord(record); err != nil {
			return nil, err
		}
		if msg.CurrentHeight < record.RetainUntilHeight {
			return nil, errors.New("x/servicereceipts receipt pruned before proof horizon")
		}
	}
	if !found {
		return nil, errors.New("x/servicereceipts receipt not found")
	}
	sortReceiptRecords(next)
	return next, validateReceiptRecords(next)
}

func ValidateExecutedOnChainCallsHaveReceipts(calls []UnifiedServiceCall, records []ReceiptRecord) error {
	receiptByCall := map[string]struct{}{}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		receiptByCall[record.CallID] = struct{}{}
	}
	for _, call := range calls {
		if call.Kind != coretypes.ServiceCallKindOnChain {
			continue
		}
		if call.CallID == "" {
			return errors.New("x/servicereceipts on-chain call id is required")
		}
		if _, found := receiptByCall[call.CallID]; !found {
			return errors.New("x/servicereceipts missing receipt for executed on-chain call")
		}
	}
	return nil
}

func QueryReceiptFromRecords(records []ReceiptRecord, query QueryReceipt) (QueryReceiptResponse, error) {
	if err := validateReceiptRecords(cloneReceiptRecords(records)); err != nil {
		return QueryReceiptResponse{}, err
	}
	if err := coretypes.ValidateHash("x/servicereceipts query receipt id", query.ReceiptID); err != nil {
		return QueryReceiptResponse{}, err
	}
	for _, record := range records {
		if record.ReceiptID == query.ReceiptID {
			return QueryReceiptResponse{Record: record, Found: true}, nil
		}
	}
	return QueryReceiptResponse{}, nil
}

func QueryReceiptsByServiceFromRecords(records []ReceiptRecord, query QueryReceiptsByService) (QueryReceiptsByServiceResponse, error) {
	if err := validateReceiptRecords(cloneReceiptRecords(records)); err != nil {
		return QueryReceiptsByServiceResponse{}, err
	}
	if err := validateInterfaceToken("x/servicereceipts query service id", query.ServiceID); err != nil {
		return QueryReceiptsByServiceResponse{}, err
	}
	out := []ReceiptRecord{}
	for _, record := range records {
		if record.ServiceID == query.ServiceID {
			out = append(out, record)
		}
	}
	sortReceiptRecords(out)
	return QueryReceiptsByServiceResponse{Records: out, ResponseHash: ComputeReceiptsByServiceResponseHash(query.ServiceID, out)}, nil
}

func QueryReceiptRootFromRecords(records []ReceiptRecord, query QueryReceiptRoot) (QueryReceiptRootResponse, error) {
	root, err := BuildReceiptRoot(query.RootKind, query.Height, records)
	if err != nil {
		return QueryReceiptRootResponse{}, err
	}
	return QueryReceiptRootResponse{Root: root, Found: true}, nil
}

func QueryReceiptProofFromRecords(records []ReceiptRecord, query QueryReceiptProof) (QueryReceiptProofResponse, error) {
	if err := validateReceiptRecords(cloneReceiptRecords(records)); err != nil {
		return QueryReceiptProofResponse{}, err
	}
	if err := coretypes.ValidateHash("x/servicereceipts proof receipt id", query.ReceiptID); err != nil {
		return QueryReceiptProofResponse{}, err
	}
	root, err := BuildReceiptRoot(ServiceReceiptRecordKindService, maxReceiptRecordHeight(records), records)
	if err != nil {
		return QueryReceiptProofResponse{}, err
	}
	for _, record := range records {
		if record.ReceiptID != query.ReceiptID {
			continue
		}
		proof := ServiceReceiptProofRecord{
			ReceiptID:	record.ReceiptID,
			ServiceID:	record.ServiceID,
			ReceiptHash:	record.ReceiptHash,
			RootHash:	root.RootHash,
			ProofHeight:	root.Height,
			ProofHashes:	receiptRecordSiblingHashes(records, record.ReceiptID),
		}
		proof.ProofHash = ComputeServiceReceiptProofRecordHash(proof)
		return QueryReceiptProofResponse{Proof: proof, Found: true}, proof.Validate()
	}
	return QueryReceiptProofResponse{}, nil
}

func (breakdown XServiceReceiptsModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModuleReceipts {
		return errors.New("x/servicereceipts module path must be x/servicereceipts")
	}
	if err := validateSortedTokens("x/servicereceipts purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServiceReceiptsEnumSet("state", breakdown.StateObjects, requiredXServiceReceiptsStates(), IsXServiceReceiptsStateObject); err != nil {
		return err
	}
	if err := validateXServiceReceiptsEnumSet("message", breakdown.Messages, requiredXServiceReceiptsMessages(), IsXServiceReceiptsMessageName); err != nil {
		return err
	}
	if err := validateXServiceReceiptsEnumSet("query", breakdown.Queries, requiredXServiceReceiptsQueries(), IsXServiceReceiptsQueryName); err != nil {
		return err
	}
	if err := validateXServiceReceiptsFailureCoverages(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServiceReceiptsEnumSet("integration", breakdown.IntegrationPoints, requiredXServiceReceiptsIntegrations(), IsXServiceReceiptsIntegrationPoint); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/servicereceipts breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServiceReceiptsModuleBreakdown) Validate() error {
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/servicereceipts breakdown hash is required")
	}
	if expected := ComputeXServiceReceiptsModuleBreakdownHash(breakdown); breakdown.BreakdownHash != expected {
		return fmt.Errorf("x/servicereceipts breakdown hash mismatch: expected %s", expected)
	}
	return nil
}

func (coverage XServiceReceiptsFailureCoverage) Validate() error {
	if !IsXServiceReceiptsFailureMode(coverage.Mode) {
		return fmt.Errorf("x/servicereceipts unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/servicereceipts failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.Scope + "/_") {
		return errors.New("x/servicereceipts failure scope must use services store prefix")
	}
	return nil
}

func (record ReceiptRecord) Validate() error {
	if err := coretypes.ValidateHash("x/servicereceipts receipt id", record.ReceiptID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/servicereceipts service id", record.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts call id", record.CallID); err != nil {
		return err
	}
	if !IsServiceReceiptRecordKind(record.Kind) {
		return fmt.Errorf("x/servicereceipts unknown record kind %q", record.Kind)
	}
	if !IsServiceReceiptStatusText(record.Status) {
		return fmt.Errorf("x/servicereceipts unknown status %q", record.Status)
	}
	if err := coretypes.ValidateHash("x/servicereceipts receipt hash", record.ReceiptHash); err != nil {
		return err
	}
	if record.Height == 0 || record.RetainUntilHeight < record.Height {
		return errors.New("x/servicereceipts record heights are invalid")
	}
	if err := coretypes.ValidateHash("x/servicereceipts record hash", record.RecordHash); err != nil {
		return err
	}
	if expected := ComputeReceiptRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("x/servicereceipts record hash mismatch: expected %s", expected)
	}
	return nil
}

func (root ReceiptRoot) Validate() error {
	if !IsServiceReceiptRecordKind(root.RootKind) {
		return fmt.Errorf("x/servicereceipts unknown root kind %q", root.RootKind)
	}
	if root.Height == 0 {
		return errors.New("x/servicereceipts root height must be positive")
	}
	if err := coretypes.ValidateHash("x/servicereceipts root hash", root.RootHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts root commitment hash", root.CommitmentHash); err != nil {
		return err
	}
	if expected := ComputeReceiptRootCommitmentHash(root); root.CommitmentHash != expected {
		return fmt.Errorf("x/servicereceipts root commitment hash mismatch: expected %s", expected)
	}
	return nil
}

func (params ReceiptParams) Validate() error {
	if params.ProofHorizon == 0 {
		return errors.New("x/servicereceipts proof horizon must be positive")
	}
	if params.PruneBatchSize == 0 {
		return errors.New("x/servicereceipts prune batch size must be positive")
	}
	if err := coretypes.ValidateHash("x/servicereceipts params hash", params.ParamsHash); err != nil {
		return err
	}
	if expected := ComputeReceiptParamsHash(params); params.ParamsHash != expected {
		return fmt.Errorf("x/servicereceipts params hash mismatch: expected %s", expected)
	}
	return nil
}

func (msg MsgAnchorReceipt) ValidateBasic() error {
	if err := addressing.ValidateAuthorityAddress("x/servicereceipts anchor authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Receipt.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts expected receipt hash", msg.ExpectedReceiptHash); err != nil {
		return err
	}
	if msg.ExpectedReceiptHash != msg.Receipt.ReceiptHash {
		return errors.New("x/servicereceipts receipt hash mismatch")
	}
	if err := coretypes.ValidateHash("x/servicereceipts anchor message hash", msg.MessageHash); err != nil {
		return err
	}
	if expected := ComputeMsgAnchorReceiptHash(msg); msg.MessageHash != expected {
		return fmt.Errorf("x/servicereceipts anchor message hash mismatch: expected %s", expected)
	}
	return nil
}

func (msg MsgPruneReceipt) ValidateForRecord(record ReceiptRecord) error {
	if err := addressing.ValidateAuthorityAddress("x/servicereceipts prune authority", msg.Authority); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts prune receipt id", msg.ReceiptID); err != nil {
		return err
	}
	if msg.ReceiptID != record.ReceiptID {
		return errors.New("x/servicereceipts prune receipt mismatch")
	}
	if msg.CurrentHeight == 0 {
		return errors.New("x/servicereceipts prune height must be positive")
	}
	if err := coretypes.ValidateHash("x/servicereceipts prune message hash", msg.MessageHash); err != nil {
		return err
	}
	if expected := ComputeMsgPruneReceiptHash(msg); msg.MessageHash != expected {
		return fmt.Errorf("x/servicereceipts prune message hash mismatch: expected %s", expected)
	}
	return nil
}

func (proof ServiceReceiptProofRecord) Validate() error {
	if err := coretypes.ValidateHash("x/servicereceipts proof receipt id", proof.ReceiptID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/servicereceipts proof service id", proof.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts proof receipt hash", proof.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicereceipts proof root hash", proof.RootHash); err != nil {
		return err
	}
	if proof.ProofHeight == 0 {
		return errors.New("x/servicereceipts proof height must be positive")
	}
	if err := validateSortedTokens("x/servicereceipts proof sibling hash", proof.ProofHashes); err != nil {
		return err
	}
	for _, hash := range proof.ProofHashes {
		if err := coretypes.ValidateHash("x/servicereceipts proof sibling hash", hash); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("x/servicereceipts proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptProofRecordHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("x/servicereceipts proof hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeXServiceReceiptsModuleBreakdownHash(breakdown XServiceReceiptsModuleBreakdown) string {
	breakdown = canonicalXServiceReceiptsModuleBreakdown(breakdown)
	parts := []string{"aetra-x-servicereceipts-breakdown-v1", breakdown.ModulePath}
	parts = appendStringParts(parts, "purpose", breakdown.Purpose)
	for _, state := range breakdown.StateObjects {
		parts = append(parts, "state", string(state))
	}
	for _, msg := range breakdown.Messages {
		parts = append(parts, "message", string(msg))
	}
	for _, query := range breakdown.Queries {
		parts = append(parts, "query", string(query))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.Scope)
	}
	for _, integration := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(integration))
	}
	return servicesHashParts(parts...)
}

func ComputeReceiptRecordHash(record ReceiptRecord) string {
	return servicesHashParts(
		"aetra-x-servicereceipts-record-v1",
		record.ReceiptID,
		record.ServiceID,
		record.CallID,
		string(record.Kind),
		string(record.Status),
		record.ReceiptHash,
		fmt.Sprint(record.Height),
		fmt.Sprint(record.RetainUntilHeight),
	)
}

func ComputeReceiptRecordRootHash(kind ServiceReceiptRecordKind, records []ReceiptRecord) string {
	ordered := cloneReceiptRecords(records)
	sortReceiptRecords(ordered)
	parts := []string{"aetra-x-servicereceipts-root-v1", string(kind), fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return servicesHashParts(parts...)
}

func ComputeReceiptRootCommitmentHash(root ReceiptRoot) string {
	return servicesHashParts("aetra-x-servicereceipts-root-commitment-v1", string(root.RootKind), fmt.Sprint(root.Height), fmt.Sprint(root.RecordCount), root.RootHash)
}

func ComputeReceiptParamsHash(params ReceiptParams) string {
	return servicesHashParts("aetra-x-servicereceipts-params-v1", fmt.Sprint(params.ProofHorizon), fmt.Sprint(params.PruneBatchSize))
}

func ComputeMsgAnchorReceiptHash(msg MsgAnchorReceipt) string {
	return servicesHashParts("aetra-x-servicereceipts-msg-anchor-v1", msg.Authority, msg.Receipt.CallID, msg.ExpectedReceiptHash)
}

func ComputeMsgPruneReceiptHash(msg MsgPruneReceipt) string {
	return servicesHashParts("aetra-x-servicereceipts-msg-prune-v1", msg.Authority, msg.ReceiptID, fmt.Sprint(msg.CurrentHeight))
}

func ComputeReceiptsByServiceResponseHash(serviceID string, records []ReceiptRecord) string {
	ordered := cloneReceiptRecords(records)
	sortReceiptRecords(ordered)
	parts := []string{"aetra-x-servicereceipts-by-service-v1", serviceID, fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceReceiptProofRecordHash(proof ServiceReceiptProofRecord) string {
	hashes := append([]string(nil), proof.ProofHashes...)
	sort.Strings(hashes)
	parts := []string{"aetra-x-servicereceipts-proof-v1", proof.ReceiptID, proof.ServiceID, proof.ReceiptHash, proof.RootHash, fmt.Sprint(proof.ProofHeight)}
	parts = append(parts, hashes...)
	return servicesHashParts(parts...)
}

func IsXServiceReceiptsStateObject(value XServiceReceiptsStateObject) bool {
	switch value {
	case XServiceReceiptsStateReceiptParams, XServiceReceiptsStateReceiptRecord, XServiceReceiptsStateReceiptRoot, XServiceReceiptsStateReceiptTombstone:
		return true
	default:
		return false
	}
}

func IsXServiceReceiptsMessageName(value XServiceReceiptsMessageName) bool {
	switch value {
	case XServiceReceiptsMsgAnchorReceipt, XServiceReceiptsMsgPruneReceipt:
		return true
	default:
		return false
	}
}

func IsXServiceReceiptsQueryName(value XServiceReceiptsQueryName) bool {
	switch value {
	case XServiceReceiptsQueryReceipt, XServiceReceiptsQueryReceiptProof, XServiceReceiptsQueryReceiptRoot, XServiceReceiptsQueryReceiptsByService:
		return true
	default:
		return false
	}
}

func IsXServiceReceiptsFailureMode(value XServiceReceiptsFailureMode) bool {
	switch value {
	case XServiceReceiptsFailureDuplicateReceipt, XServiceReceiptsFailureMissingExecutedOnChainReceipt, XServiceReceiptsFailureReceiptHashMismatch, XServiceReceiptsFailurePrunedBeforeProofHorizon:
		return true
	default:
		return false
	}
}

func IsXServiceReceiptsIntegrationPoint(value XServiceReceiptsIntegrationPoint) bool {
	switch value {
	case XServiceReceiptsIntegrationAllServiceModules, XServiceReceiptsIntegrationProofRegistry, XServiceReceiptsIntegrationStoreV2:
		return true
	default:
		return false
	}
}

func IsServiceReceiptRecordKind(value ServiceReceiptRecordKind) bool {
	switch value {
	case ServiceReceiptRecordKindCall, ServiceReceiptRecordKindPayment, ServiceReceiptRecordKindProvider, ServiceReceiptRecordKindService, ServiceReceiptRecordKindStorage:
		return true
	default:
		return false
	}
}

func newXServiceReceiptsFailureCoverage(mode XServiceReceiptsFailureMode, guard, scope string) XServiceReceiptsFailureCoverage {
	return XServiceReceiptsFailureCoverage{Mode: mode, Guard: guard, Scope: scope}
}

func canonicalXServiceReceiptsModuleBreakdown(breakdown XServiceReceiptsModuleBreakdown) XServiceReceiptsModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	sort.Strings(breakdown.Purpose)
	sort.SliceStable(breakdown.StateObjects, func(i, j int) bool { return breakdown.StateObjects[i] < breakdown.StateObjects[j] })
	sort.SliceStable(breakdown.Messages, func(i, j int) bool { return breakdown.Messages[i] < breakdown.Messages[j] })
	sort.SliceStable(breakdown.Queries, func(i, j int) bool { return breakdown.Queries[i] < breakdown.Queries[j] })
	sort.SliceStable(breakdown.FailureModes, func(i, j int) bool { return breakdown.FailureModes[i].Mode < breakdown.FailureModes[j].Mode })
	sort.SliceStable(breakdown.IntegrationPoints, func(i, j int) bool { return breakdown.IntegrationPoints[i] < breakdown.IntegrationPoints[j] })
	breakdown.BreakdownHash = strings.ToLower(strings.TrimSpace(breakdown.BreakdownHash))
	return breakdown
}

func validateXServiceReceiptsEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/servicereceipts expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/servicereceipts unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/servicereceipts %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/servicereceipts duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/servicereceipts missing %s %s", label, value)
		}
	}
	return nil
}

func validateXServiceReceiptsFailureCoverages(values []XServiceReceiptsFailureCoverage) error {
	required := requiredXServiceReceiptsFailures()
	if len(values) != len(required) {
		return fmt.Errorf("x/servicereceipts expected %d failure entries", len(required))
	}
	seen := map[XServiceReceiptsFailureMode]struct{}{}
	previous := ""
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		current := string(value.Mode)
		if previous != "" && previous >= current {
			return errors.New("x/servicereceipts failure entries must be sorted canonically")
		}
		previous = current
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("x/servicereceipts duplicate failure %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/servicereceipts missing failure %s", value)
		}
	}
	return nil
}

func requiredXServiceReceiptsStates() []XServiceReceiptsStateObject {
	return []XServiceReceiptsStateObject{XServiceReceiptsStateReceiptParams, XServiceReceiptsStateReceiptRecord, XServiceReceiptsStateReceiptRoot, XServiceReceiptsStateReceiptTombstone}
}

func requiredXServiceReceiptsMessages() []XServiceReceiptsMessageName {
	return []XServiceReceiptsMessageName{XServiceReceiptsMsgAnchorReceipt, XServiceReceiptsMsgPruneReceipt}
}

func requiredXServiceReceiptsQueries() []XServiceReceiptsQueryName {
	return []XServiceReceiptsQueryName{XServiceReceiptsQueryReceipt, XServiceReceiptsQueryReceiptProof, XServiceReceiptsQueryReceiptRoot, XServiceReceiptsQueryReceiptsByService}
}

func requiredXServiceReceiptsFailures() []XServiceReceiptsFailureMode {
	return []XServiceReceiptsFailureMode{XServiceReceiptsFailureDuplicateReceipt, XServiceReceiptsFailureMissingExecutedOnChainReceipt, XServiceReceiptsFailureReceiptHashMismatch, XServiceReceiptsFailurePrunedBeforeProofHorizon}
}

func requiredXServiceReceiptsIntegrations() []XServiceReceiptsIntegrationPoint {
	return []XServiceReceiptsIntegrationPoint{XServiceReceiptsIntegrationAllServiceModules, XServiceReceiptsIntegrationProofRegistry, XServiceReceiptsIntegrationStoreV2}
}

func validateReceiptRecords(records []ReceiptRecord) error {
	sortReceiptRecords(records)
	previous := ""
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= record.ReceiptID {
			return errors.New("x/servicereceipts records must be sorted canonically")
		}
		previous = record.ReceiptID
	}
	return nil
}

func cloneReceiptRecords(records []ReceiptRecord) []ReceiptRecord {
	out := make([]ReceiptRecord, len(records))
	copy(out, records)
	return out
}

func sortReceiptRecords(records []ReceiptRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ReceiptID < records[j].ReceiptID })
}

func maxReceiptRecordHeight(records []ReceiptRecord) uint64 {
	var max uint64
	for _, record := range records {
		if record.Height > max {
			max = record.Height
		}
	}
	if max == 0 {
		return 1
	}
	return max
}

func receiptRecordSiblingHashes(records []ReceiptRecord, receiptID string) []string {
	hashes := []string{}
	for _, record := range records {
		if record.ReceiptID == receiptID {
			continue
		}
		hashes = append(hashes, record.RecordHash)
	}
	sort.Strings(hashes)
	return hashes
}
