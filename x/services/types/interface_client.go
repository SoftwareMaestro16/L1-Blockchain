package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type InterfaceUXStep string

const (
	InterfaceUXUserInput		InterfaceUXStep	= "user_input"
	InterfaceUXResolveService	InterfaceUXStep	= "resolve_service"
	InterfaceUXFetchInterface	InterfaceUXStep	= "fetch_interface"
	InterfaceUXVerifyInterfaceHash	InterfaceUXStep	= "verify_interface_hash"
	InterfaceUXGenerateForm		InterfaceUXStep	= "generate_form"
	InterfaceUXBuildCall		InterfaceUXStep	= "build_call"
	InterfaceUXExecuteCall		InterfaceUXStep	= "execute_call"
	InterfaceUXVerifyReceipt	InterfaceUXStep	= "verify_receipt"

	WalletSchemaFormatJSONV1	= "wallet-json-v1"
	CLISchemaFormatJSONV1		= "cli-json-v1"
)

var interfaceDrivenUXSteps = []InterfaceUXStep{
	InterfaceUXUserInput,
	InterfaceUXResolveService,
	InterfaceUXFetchInterface,
	InterfaceUXVerifyInterfaceHash,
	InterfaceUXGenerateForm,
	InterfaceUXBuildCall,
	InterfaceUXExecuteCall,
	InterfaceUXVerifyReceipt,
}

type InterfaceSchemaFormat struct {
	InterfaceHash		string
	MethodID		string
	SchemaEncoding		string
	InputSchemaHash		string
	OutputSchemaHash	string
	WalletFormat		string
	CLIFormat		string
	MetadataHash		string
	FormatHash		string
}

type InterfaceDrivenUXFlow struct {
	ServiceID			string
	InterfaceHash			string
	MethodName			string
	PaymentModel			string
	TrustModel			coretypes.ServiceTrustModel
	VerificationModel		coretypes.ServiceVerificationModel
	Steps				[]InterfaceUXStep
	DisplayPaymentAndTrustModel	bool
	RequireUserSigningConfirmation	bool
	MetadataGrantsAuthorization	bool
	FlowHash			string
}

type MsgRegisterInterface struct {
	Authority	string
	Interface	ServiceInterface
	Schema		InterfaceSchemaFormat
	MessageHash	string
}

type MsgUpdateInterface struct {
	Authority		string
	PreviousInterfaceHash	string
	Interface		ServiceInterface
	Schema			InterfaceSchemaFormat
	ExpectedVersion		uint64
	MessageHash		string
}

type QueryInterfaceProof struct {
	InterfaceHash string
}

type InterfaceProof struct {
	InterfaceHash	string
	RegistryRoot	string
	RecordHash	string
	ProofHeight	uint64
	ProofHash	string
}

type QueryInterfaceProofResponse struct {
	Proof	InterfaceProof
	Found	bool
}

type SDKInterfaceVerification struct {
	ServiceID		string
	InterfaceHash		string
	Definition		FormalServiceInterface
	PaymentModel		string
	TrustModel		coretypes.ServiceTrustModel
	Verification		coretypes.ServiceVerificationModel
	VerificationHash	string
}

type WalletCLISchema struct {
	ServiceID		string
	InterfaceHash		string
	MethodID		string
	MethodName		string
	SchemaEncoding		string
	InputSchemaHash		string
	OutputSchemaHash	string
	WalletFormat		string
	CLIFormat		string
	PaymentModel		string
	TrustModel		coretypes.ServiceTrustModel
	Verification		coretypes.ServiceVerificationModel
	SchemaHash		string
}

type InterfaceCompatibilityReport struct {
	PreviousInterfaceHash	string
	NextInterfaceHash	string
	PreviousVersion		uint64
	NextVersion		uint64
	Compatible		bool
	BreakingChanges		[]string
	AddedMethods		[]string
	ReportHash		string
}

func NewInterfaceSchemaFormat(iface ServiceInterface, methodName string) (InterfaceSchemaFormat, error) {
	definition, err := NewFormalServiceInterface(iface)
	if err != nil {
		return InterfaceSchemaFormat{}, err
	}
	method, found := definition.MethodByName(methodName)
	if !found {
		return InterfaceSchemaFormat{}, fmt.Errorf("services interface method %s not found", methodName)
	}
	format := InterfaceSchemaFormat{
		InterfaceHash:		definition.InterfaceHash,
		MethodID:		method.MethodID,
		SchemaEncoding:		definition.SchemaEncoding,
		InputSchemaHash:	method.InputSchemaHash,
		OutputSchemaHash:	method.OutputSchemaHash,
		WalletFormat:		WalletSchemaFormatJSONV1,
		CLIFormat:		CLISchemaFormatJSONV1,
		MetadataHash:		definition.MetadataHash,
	}
	format.FormatHash = ComputeInterfaceSchemaFormatHash(format)
	return format, format.Validate()
}

func (format InterfaceSchemaFormat) Validate() error {
	if err := coretypes.ValidateHash("services interface schema interface hash", format.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface schema method id", format.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface schema encoding", format.SchemaEncoding); err != nil {
		return err
	}
	if !IsSupportedServiceSchemaEncoding(format.SchemaEncoding) {
		return fmt.Errorf("services interface schema encoding %q is not supported", format.SchemaEncoding)
	}
	if err := coretypes.ValidateHash("services interface schema input", format.InputSchemaHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface schema output", format.OutputSchemaHash); err != nil {
		return err
	}
	if !isSupportedWalletSchemaFormat(format.WalletFormat) {
		return fmt.Errorf("services interface wallet schema format %q is not supported", format.WalletFormat)
	}
	if !isSupportedCLISchemaFormat(format.CLIFormat) {
		return fmt.Errorf("services interface cli schema format %q is not supported", format.CLIFormat)
	}
	if format.MetadataHash != "" {
		if err := coretypes.ValidateHash("services interface schema metadata", format.MetadataHash); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services interface schema format hash", format.FormatHash); err != nil {
		return err
	}
	if expected := ComputeInterfaceSchemaFormatHash(format); format.FormatHash != expected {
		return fmt.Errorf("services interface schema format hash mismatch: expected %s", expected)
	}
	return nil
}

func NewInterfaceDrivenUXFlow(descriptor ServiceDescriptor, methodName string) (InterfaceDrivenUXFlow, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := ValidateServiceInterfaceForDescriptor(descriptor); err != nil {
		return InterfaceDrivenUXFlow{}, err
	}
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	if err != nil {
		return InterfaceDrivenUXFlow{}, err
	}
	method, found := definition.MethodByName(methodName)
	if !found {
		return InterfaceDrivenUXFlow{}, fmt.Errorf("services interface method %s not found", methodName)
	}
	flow := InterfaceDrivenUXFlow{
		ServiceID:			descriptor.ServiceID,
		InterfaceHash:			definition.InterfaceHash,
		MethodName:			method.Name,
		PaymentModel:			registryPaymentModelFromDescriptor(descriptor),
		TrustModel:			descriptor.Verification.TrustModel,
		VerificationModel:		method.VerificationModel,
		Steps:				append([]InterfaceUXStep(nil), interfaceDrivenUXSteps...),
		DisplayPaymentAndTrustModel:	true,
		RequireUserSigningConfirmation:	true,
		MetadataGrantsAuthorization:	false,
	}
	flow.FlowHash = ComputeInterfaceDrivenUXFlowHash(flow)
	return flow, flow.Validate()
}

func (flow InterfaceDrivenUXFlow) Validate() error {
	if err := validateInterfaceToken("services interface ux service id", flow.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface ux interface hash", flow.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface ux method name", flow.MethodName); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface ux payment model", flow.PaymentModel); err != nil {
		return err
	}
	if !coretypes.IsServiceTrustModel(flow.TrustModel) {
		return fmt.Errorf("services interface ux unknown trust model %q", flow.TrustModel)
	}
	if !coretypes.IsServiceVerificationModel(flow.VerificationModel) {
		return fmt.Errorf("services interface ux unknown verification model %q", flow.VerificationModel)
	}
	if len(flow.Steps) != len(interfaceDrivenUXSteps) {
		return errors.New("services interface ux flow must include the canonical client steps")
	}
	for i, step := range flow.Steps {
		if step != interfaceDrivenUXSteps[i] {
			return errors.New("services interface ux flow steps are not canonical")
		}
	}
	if !flow.DisplayPaymentAndTrustModel {
		return errors.New("services interface ux must display payment and trust model")
	}
	if !flow.RequireUserSigningConfirmation {
		return errors.New("services interface ux requires explicit user signing confirmation")
	}
	if flow.MetadataGrantsAuthorization {
		return errors.New("services interface ux metadata must not grant authorization")
	}
	if err := coretypes.ValidateHash("services interface ux flow hash", flow.FlowHash); err != nil {
		return err
	}
	if expected := ComputeInterfaceDrivenUXFlowHash(flow); flow.FlowHash != expected {
		return fmt.Errorf("services interface ux flow hash mismatch: expected %s", expected)
	}
	return nil
}

func NewMsgRegisterInterface(authority string, iface ServiceInterface, schema InterfaceSchemaFormat) (MsgRegisterInterface, error) {
	msg := MsgRegisterInterface{
		Authority:	strings.TrimSpace(authority),
		Interface:	coretypes.CanonicalServiceInterfaceDescriptor(iface),
		Schema:		schema,
	}
	msg.MessageHash = ComputeInterfaceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgUpdateInterface(authority, previousInterfaceHash string, iface ServiceInterface, schema InterfaceSchemaFormat, expectedVersion uint64) (MsgUpdateInterface, error) {
	msg := MsgUpdateInterface{
		Authority:		strings.TrimSpace(authority),
		PreviousInterfaceHash:	strings.ToLower(strings.TrimSpace(previousInterfaceHash)),
		Interface:		coretypes.CanonicalServiceInterfaceDescriptor(iface),
		Schema:			schema,
		ExpectedVersion:	expectedVersion,
	}
	msg.MessageHash = ComputeInterfaceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func (m MsgRegisterInterface) ServiceRegistryMessageName() string	{ return "MsgRegisterInterface" }
func (m MsgRegisterInterface) ServiceRegistrySigner() string		{ return m.Authority }
func (m MsgRegisterInterface) ValidateBasic() error {
	if err := addressing.ValidateAuthorityAddress("services interface registration authority", m.Authority); err != nil {
		return err
	}
	if err := m.Interface.Validate(); err != nil {
		return err
	}
	if m.Interface.InterfaceHash != coretypes.ComputeServiceInterfaceHash(m.Interface) {
		return errors.New("services interface registration hash must commit to interface fields")
	}
	if err := m.Schema.Validate(); err != nil {
		return err
	}
	if m.Schema.InterfaceHash != m.Interface.InterfaceHash {
		return errors.New("services interface registration schema hash mismatch")
	}
	return validateInterfaceRegistryMessageHash(m, m.MessageHash)
}

func (m MsgUpdateInterface) ServiceRegistryMessageName() string	{ return "MsgUpdateInterface" }
func (m MsgUpdateInterface) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgUpdateInterface) ValidateBasic() error {
	if err := addressing.ValidateAuthorityAddress("services interface update authority", m.Authority); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface update previous hash", m.PreviousInterfaceHash); err != nil {
		return err
	}
	if m.ExpectedVersion == 0 {
		return errors.New("services interface update expected version must be positive")
	}
	if err := m.Interface.Validate(); err != nil {
		return err
	}
	if m.Interface.Version <= m.ExpectedVersion {
		return errors.New("services interface update version must exceed expected version")
	}
	if m.Interface.InterfaceHash == m.PreviousInterfaceHash {
		return errors.New("services interface update must create a new interface hash")
	}
	if m.Interface.InterfaceHash != coretypes.ComputeServiceInterfaceHash(m.Interface) {
		return errors.New("services interface update hash must commit to interface fields")
	}
	if err := m.Schema.Validate(); err != nil {
		return err
	}
	if m.Schema.InterfaceHash != m.Interface.InterfaceHash {
		return errors.New("services interface update schema hash mismatch")
	}
	return validateInterfaceRegistryMessageHash(m, m.MessageHash)
}

func (q QueryInterfaceProof) Validate() error {
	return coretypes.ValidateHash("services query interface proof hash", q.InterfaceHash)
}

func QueryInterfaceProofFromState(state ServiceRegistryState, q QueryInterfaceProof) (QueryInterfaceProofResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryInterfaceProofResponse{}, err
	}
	iface, found := state.ServiceInterfaceByHash(q.InterfaceHash)
	if !found {
		return QueryInterfaceProofResponse{Found: false}, nil
	}
	recordHash := ComputeInterfaceProofRecordHash(iface)
	proof := InterfaceProof{
		InterfaceHash:	iface.InterfaceHash,
		RegistryRoot:	state.StateRoot,
		RecordHash:	recordHash,
		ProofHeight:	state.UpdatedHeight,
	}
	proof.ProofHash = ComputeInterfaceProofHash(proof)
	return QueryInterfaceProofResponse{Proof: proof, Found: true}, proof.Validate()
}

func RegisterInterfaceInState(state ServiceRegistryState, msg MsgRegisterInterface, height uint64) (ServiceRegistryState, error) {
	if err := msg.ValidateBasic(); err != nil {
		return ServiceRegistryState{}, err
	}
	if height == 0 {
		height = state.UpdatedHeight + 1
	}
	if height < state.UpdatedHeight {
		return ServiceRegistryState{}, errors.New("services interface registration height must not go backwards")
	}
	if _, found := state.ServiceInterfaceByHash(msg.Interface.InterfaceHash); found {
		return ServiceRegistryState{}, fmt.Errorf("services interface %s already exists", msg.Interface.InterfaceHash)
	}
	return appendInterfaceToState(state, msg.Interface, height)
}

func UpdateInterfaceInState(state ServiceRegistryState, msg MsgUpdateInterface, height uint64) (ServiceRegistryState, error) {
	if err := msg.ValidateBasic(); err != nil {
		return ServiceRegistryState{}, err
	}
	if height == 0 {
		height = state.UpdatedHeight + 1
	}
	if height < state.UpdatedHeight {
		return ServiceRegistryState{}, errors.New("services interface update height must not go backwards")
	}
	previous, found := state.ServiceInterfaceByHash(msg.PreviousInterfaceHash)
	if !found {
		return ServiceRegistryState{}, fmt.Errorf("services interface %s not found", msg.PreviousInterfaceHash)
	}
	if previous.Version != msg.ExpectedVersion {
		return ServiceRegistryState{}, errors.New("services interface update expected version mismatch")
	}
	if _, found := state.ServiceInterfaceByHash(msg.Interface.InterfaceHash); found {
		return ServiceRegistryState{}, fmt.Errorf("services interface %s already exists", msg.Interface.InterfaceHash)
	}
	if err := ValidateServiceInterfaceVersionChange(previous, msg.Interface); err != nil {
		return ServiceRegistryState{}, err
	}
	return appendInterfaceToState(state, msg.Interface, height)
}

func (proof InterfaceProof) Validate() error {
	if err := coretypes.ValidateHash("services interface proof interface hash", proof.InterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface proof registry root", proof.RegistryRoot); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface proof record hash", proof.RecordHash); err != nil {
		return err
	}
	if proof.ProofHeight == 0 {
		return errors.New("services interface proof height must be positive")
	}
	if err := coretypes.ValidateHash("services interface proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeInterfaceProofHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("services interface proof hash mismatch: expected %s", expected)
	}
	return nil
}

func VerifySDKInterface(descriptor ServiceDescriptor, expectedInterfaceHash string) (SDKInterfaceVerification, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	expectedInterfaceHash = strings.ToLower(strings.TrimSpace(expectedInterfaceHash))
	if err := coretypes.ValidateHash("services sdk expected interface hash", expectedInterfaceHash); err != nil {
		return SDKInterfaceVerification{}, err
	}
	if descriptor.Interface.InterfaceHash != expectedInterfaceHash {
		return SDKInterfaceVerification{}, errors.New("services sdk interface hash mismatch")
	}
	if err := ValidateServiceInterfaceForDescriptor(descriptor); err != nil {
		return SDKInterfaceVerification{}, err
	}
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	if err != nil {
		return SDKInterfaceVerification{}, err
	}
	verification := SDKInterfaceVerification{
		ServiceID:	descriptor.ServiceID,
		InterfaceHash:	definition.InterfaceHash,
		Definition:	definition,
		PaymentModel:	registryPaymentModelFromDescriptor(descriptor),
		TrustModel:	descriptor.Verification.TrustModel,
		Verification:	descriptor.Verification.Model,
	}
	verification.VerificationHash = ComputeSDKInterfaceVerificationHash(verification)
	return verification, verification.Validate()
}

func (verification SDKInterfaceVerification) Validate() error {
	if err := validateInterfaceToken("services sdk verification service id", verification.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services sdk verification interface hash", verification.InterfaceHash); err != nil {
		return err
	}
	if err := verification.Definition.Validate(); err != nil {
		return err
	}
	if verification.Definition.InterfaceHash != verification.InterfaceHash {
		return errors.New("services sdk verification definition hash mismatch")
	}
	if err := validateInterfaceToken("services sdk verification payment model", verification.PaymentModel); err != nil {
		return err
	}
	if !coretypes.IsServiceTrustModel(verification.TrustModel) {
		return fmt.Errorf("services sdk verification unknown trust model %q", verification.TrustModel)
	}
	if !coretypes.IsServiceVerificationModel(verification.Verification) {
		return fmt.Errorf("services sdk verification unknown verification model %q", verification.Verification)
	}
	if err := coretypes.ValidateHash("services sdk verification hash", verification.VerificationHash); err != nil {
		return err
	}
	if expected := ComputeSDKInterfaceVerificationHash(verification); verification.VerificationHash != expected {
		return fmt.Errorf("services sdk verification hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildMethodCallFromUserInput(descriptor ServiceDescriptor, methodName, caller string, nonce uint64, payloadHash string, preparedHeight uint64, userConfirmed bool) (ServiceInterfaceCallPreparation, error) {
	if !userConfirmed {
		return ServiceInterfaceCallPreparation{}, errors.New("services interface call requires explicit user confirmation")
	}
	if _, err := VerifySDKInterface(descriptor, descriptor.Interface.InterfaceHash); err != nil {
		return ServiceInterfaceCallPreparation{}, err
	}
	return PrepareServiceInterfaceCall(descriptor, methodName, caller, nonce, payloadHash, preparedHeight)
}

func BuildWalletCLISchema(descriptor ServiceDescriptor, methodName string) (WalletCLISchema, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if _, err := VerifySDKInterface(descriptor, descriptor.Interface.InterfaceHash); err != nil {
		return WalletCLISchema{}, err
	}
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	if err != nil {
		return WalletCLISchema{}, err
	}
	method, found := definition.MethodByName(methodName)
	if !found {
		return WalletCLISchema{}, fmt.Errorf("services interface method %s not found", methodName)
	}
	schema := WalletCLISchema{
		ServiceID:		descriptor.ServiceID,
		InterfaceHash:		definition.InterfaceHash,
		MethodID:		method.MethodID,
		MethodName:		method.Name,
		SchemaEncoding:		definition.SchemaEncoding,
		InputSchemaHash:	method.InputSchemaHash,
		OutputSchemaHash:	method.OutputSchemaHash,
		WalletFormat:		WalletSchemaFormatJSONV1,
		CLIFormat:		CLISchemaFormatJSONV1,
		PaymentModel:		registryPaymentModelFromDescriptor(descriptor),
		TrustModel:		descriptor.Verification.TrustModel,
		Verification:		method.VerificationModel,
	}
	schema.SchemaHash = ComputeWalletCLISchemaHash(schema)
	return schema, schema.Validate()
}

func (schema WalletCLISchema) Validate() error {
	if err := validateInterfaceToken("services wallet cli schema service id", schema.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services wallet cli schema interface hash", schema.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services wallet cli schema method id", schema.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services wallet cli schema method name", schema.MethodName); err != nil {
		return err
	}
	if err := validateInterfaceToken("services wallet cli schema encoding", schema.SchemaEncoding); err != nil {
		return err
	}
	if !IsSupportedServiceSchemaEncoding(schema.SchemaEncoding) {
		return fmt.Errorf("services wallet cli schema encoding %q is not supported", schema.SchemaEncoding)
	}
	if err := coretypes.ValidateHash("services wallet cli input schema", schema.InputSchemaHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services wallet cli output schema", schema.OutputSchemaHash); err != nil {
		return err
	}
	if !isSupportedWalletSchemaFormat(schema.WalletFormat) {
		return fmt.Errorf("services wallet schema format %q is not supported", schema.WalletFormat)
	}
	if !isSupportedCLISchemaFormat(schema.CLIFormat) {
		return fmt.Errorf("services cli schema format %q is not supported", schema.CLIFormat)
	}
	if err := validateInterfaceToken("services wallet cli payment model", schema.PaymentModel); err != nil {
		return err
	}
	if !coretypes.IsServiceTrustModel(schema.TrustModel) {
		return fmt.Errorf("services wallet cli unknown trust model %q", schema.TrustModel)
	}
	if !coretypes.IsServiceVerificationModel(schema.Verification) {
		return fmt.Errorf("services wallet cli unknown verification model %q", schema.Verification)
	}
	if err := coretypes.ValidateHash("services wallet cli schema hash", schema.SchemaHash); err != nil {
		return err
	}
	if expected := ComputeWalletCLISchemaHash(schema); schema.SchemaHash != expected {
		return fmt.Errorf("services wallet cli schema hash mismatch: expected %s", expected)
	}
	return nil
}

func CheckVersionedInterfaceCompatibility(previous ServiceInterface, next ServiceInterface) (InterfaceCompatibilityReport, error) {
	previous = coretypes.CanonicalServiceInterfaceDescriptor(previous)
	next = coretypes.CanonicalServiceInterfaceDescriptor(next)
	if err := ValidateServiceInterfaceVersionChange(previous, next); err != nil {
		return InterfaceCompatibilityReport{}, err
	}
	previousDefinition, err := NewFormalServiceInterface(previous)
	if err != nil {
		return InterfaceCompatibilityReport{}, err
	}
	nextDefinition, err := NewFormalServiceInterface(next)
	if err != nil {
		return InterfaceCompatibilityReport{}, err
	}
	previousMethods := make(map[string]ServiceInterfaceMethodSchema, len(previousDefinition.Methods))
	for _, method := range previousDefinition.Methods {
		previousMethods[method.MethodID] = method
	}
	nextMethods := make(map[string]ServiceInterfaceMethodSchema, len(nextDefinition.Methods))
	for _, method := range nextDefinition.Methods {
		nextMethods[method.MethodID] = method
	}
	report := InterfaceCompatibilityReport{
		PreviousInterfaceHash:	previousDefinition.InterfaceHash,
		NextInterfaceHash:	nextDefinition.InterfaceHash,
		PreviousVersion:	previousDefinition.Version,
		NextVersion:		nextDefinition.Version,
		Compatible:		true,
		BreakingChanges:	[]string{},
		AddedMethods:		[]string{},
	}
	for id, previousMethod := range previousMethods {
		nextMethod, found := nextMethods[id]
		if !found {
			report.BreakingChanges = append(report.BreakingChanges, "removed method "+id)
			continue
		}
		if previousMethod.InputSchemaHash != nextMethod.InputSchemaHash {
			report.BreakingChanges = append(report.BreakingChanges, "changed input schema "+id)
		}
		if previousMethod.OutputSchemaHash != nextMethod.OutputSchemaHash {
			report.BreakingChanges = append(report.BreakingChanges, "changed output schema "+id)
		}
		if previousMethod.ExecutionType != nextMethod.ExecutionType {
			report.BreakingChanges = append(report.BreakingChanges, "changed execution type "+id)
		}
		if previousMethod.VerificationModel != nextMethod.VerificationModel {
			report.BreakingChanges = append(report.BreakingChanges, "changed verification model "+id)
		}
		if previousMethod.PaymentRequirements != nextMethod.PaymentRequirements {
			report.BreakingChanges = append(report.BreakingChanges, "changed payment requirements "+id)
		}
	}
	for id := range nextMethods {
		if _, found := previousMethods[id]; !found {
			report.AddedMethods = append(report.AddedMethods, id)
		}
	}
	sort.Strings(report.BreakingChanges)
	sort.Strings(report.AddedMethods)
	report.Compatible = len(report.BreakingChanges) == 0
	report.ReportHash = ComputeInterfaceCompatibilityReportHash(report)
	return report, report.Validate()
}

func (report InterfaceCompatibilityReport) Validate() error {
	if err := coretypes.ValidateHash("services interface compatibility previous hash", report.PreviousInterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface compatibility next hash", report.NextInterfaceHash); err != nil {
		return err
	}
	if report.PreviousVersion == 0 || report.NextVersion <= report.PreviousVersion {
		return errors.New("services interface compatibility versions are invalid")
	}
	if report.Compatible != (len(report.BreakingChanges) == 0) {
		return errors.New("services interface compatibility flag mismatch")
	}
	if err := validateSortedTokenSet("services interface compatibility breaking change", report.BreakingChanges); err != nil {
		return err
	}
	if err := validateSortedTokenSet("services interface compatibility added method", report.AddedMethods); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface compatibility report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputeInterfaceCompatibilityReportHash(report); report.ReportHash != expected {
		return fmt.Errorf("services interface compatibility report hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeInterfaceSchemaFormatHash(format InterfaceSchemaFormat) string {
	return servicesHashParts(
		"aetra-services-interface-schema-format-v1",
		format.InterfaceHash,
		format.MethodID,
		format.SchemaEncoding,
		format.InputSchemaHash,
		format.OutputSchemaHash,
		format.WalletFormat,
		format.CLIFormat,
		format.MetadataHash,
	)
}

func ComputeInterfaceDrivenUXFlowHash(flow InterfaceDrivenUXFlow) string {
	parts := []string{
		"aetra-services-interface-ux-flow-v1",
		flow.ServiceID,
		flow.InterfaceHash,
		flow.MethodName,
		flow.PaymentModel,
		string(flow.TrustModel),
		string(flow.VerificationModel),
		fmt.Sprint(flow.DisplayPaymentAndTrustModel),
		fmt.Sprint(flow.RequireUserSigningConfirmation),
		fmt.Sprint(flow.MetadataGrantsAuthorization),
	}
	for _, step := range flow.Steps {
		parts = append(parts, string(step))
	}
	return servicesHashParts(parts...)
}

func ComputeInterfaceRegistryMessageHash(msg interface{}) string {
	switch m := msg.(type) {
	case MsgRegisterInterface:
		return servicesHashParts("aetra-services-msg-register-interface-v1", m.Authority, m.Interface.InterfaceHash, m.Schema.FormatHash)
	case MsgUpdateInterface:
		return servicesHashParts("aetra-services-msg-update-interface-v1", m.Authority, m.PreviousInterfaceHash, m.Interface.InterfaceHash, m.Schema.FormatHash, fmt.Sprint(m.ExpectedVersion))
	default:
		return coretypes.EmptyRootHash
	}
}

func ComputeInterfaceProofRecordHash(iface ServiceInterface) string {
	key, _ := coretypes.ServiceInterfaceStateKey(iface.InterfaceHash)
	return servicesHashParts("aetra-services-interface-proof-record-v1", key, coretypes.ComputeServiceInterfaceHash(iface))
}

func ComputeInterfaceProofHash(proof InterfaceProof) string {
	return servicesHashParts(
		"aetra-services-interface-proof-v1",
		proof.InterfaceHash,
		proof.RegistryRoot,
		proof.RecordHash,
		fmt.Sprint(proof.ProofHeight),
	)
}

func ComputeSDKInterfaceVerificationHash(verification SDKInterfaceVerification) string {
	return servicesHashParts(
		"aetra-services-sdk-interface-verification-v1",
		verification.ServiceID,
		verification.InterfaceHash,
		verification.Definition.DefinitionHash,
		verification.PaymentModel,
		string(verification.TrustModel),
		string(verification.Verification),
	)
}

func ComputeWalletCLISchemaHash(schema WalletCLISchema) string {
	return servicesHashParts(
		"aetra-services-wallet-cli-schema-v1",
		schema.ServiceID,
		schema.InterfaceHash,
		schema.MethodID,
		schema.MethodName,
		schema.SchemaEncoding,
		schema.InputSchemaHash,
		schema.OutputSchemaHash,
		schema.WalletFormat,
		schema.CLIFormat,
		schema.PaymentModel,
		string(schema.TrustModel),
		string(schema.Verification),
	)
}

func ComputeInterfaceCompatibilityReportHash(report InterfaceCompatibilityReport) string {
	parts := []string{
		"aetra-services-interface-compatibility-v1",
		report.PreviousInterfaceHash,
		report.NextInterfaceHash,
		fmt.Sprint(report.PreviousVersion),
		fmt.Sprint(report.NextVersion),
		fmt.Sprint(report.Compatible),
	}
	parts = appendStringParts(parts, "breaking", report.BreakingChanges)
	parts = appendStringParts(parts, "added", report.AddedMethods)
	return servicesHashParts(parts...)
}

func validateInterfaceRegistryMessageHash(msg interface{}, messageHash string) error {
	if err := coretypes.ValidateHash("services interface registry message hash", messageHash); err != nil {
		return err
	}
	if expected := ComputeInterfaceRegistryMessageHash(msg); messageHash != expected {
		return fmt.Errorf("services interface registry message hash mismatch: expected %s", expected)
	}
	return nil
}

func appendInterfaceToState(state ServiceRegistryState, iface ServiceInterface, height uint64) (ServiceRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ServiceRegistryState{}, err
	}
	iface = coretypes.CanonicalServiceInterfaceDescriptor(iface)
	if err := iface.Validate(); err != nil {
		return ServiceRegistryState{}, err
	}
	key, err := coretypes.ServiceInterfaceStateKey(iface.InterfaceHash)
	if err != nil {
		return ServiceRegistryState{}, err
	}
	entry := coretypes.ServiceRegistryStateEntry{
		Key:		key,
		Value:		iface.InterfaceHash,
		EntryType:	coretypes.ServiceRegistryStateInterface,
	}
	entry.EntryHash = coretypes.ComputeServiceRegistryStateEntryHash(entry)
	if err := entry.Validate(); err != nil {
		return ServiceRegistryState{}, err
	}
	next := cloneRegistryStateForInterfaceMutation(state)
	next.Interfaces = append(next.Interfaces, iface)
	next.Entries = append(next.Entries, entry)
	sort.SliceStable(next.Interfaces, func(i, j int) bool { return next.Interfaces[i].InterfaceHash < next.Interfaces[j].InterfaceHash })
	sort.SliceStable(next.Entries, func(i, j int) bool { return next.Entries[i].Key < next.Entries[j].Key })
	next.UpdatedHeight = height
	next.StateRoot = coretypes.ComputeServiceRegistryStateRoot(next)
	return next, next.Validate()
}

func cloneRegistryStateForInterfaceMutation(state ServiceRegistryState) ServiceRegistryState {
	state.Descriptors = append([]ServiceDescriptor(nil), state.Descriptors...)
	state.Anchors = append([]ServiceAnchor(nil), state.Anchors...)
	state.Interfaces = append([]ServiceInterface(nil), state.Interfaces...)
	state.OwnerIndex = append([]coretypes.ServiceRegistryStateEntry(nil), state.OwnerIndex...)
	state.NameIndex = append([]coretypes.ServiceRegistryStateEntry(nil), state.NameIndex...)
	state.IdentityBindings = append([]IdentityServiceBinding(nil), state.IdentityBindings...)
	state.Providers = append([]ProviderRecord(nil), state.Providers...)
	state.ExpiryIndex = append([]coretypes.ServiceRegistryStateEntry(nil), state.ExpiryIndex...)
	state.Reputations = append([]ReputationRecord(nil), state.Reputations...)
	state.Receipts = append([]ServiceReceipt(nil), state.Receipts...)
	state.Entries = append([]coretypes.ServiceRegistryStateEntry(nil), state.Entries...)
	return state
}

func validateSortedTokenSet(fieldName string, values []string) error {
	previous := ""
	for _, value := range values {
		if err := validateInterfaceCompatibilityToken(fieldName, value); err != nil {
			return err
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func validateInterfaceCompatibilityToken(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > 160 {
		return fmt.Errorf("%s must be <= 160 bytes", fieldName)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '.' || r == ':' || r == '/' || r == ' ' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func registryPaymentModelFromDescriptor(descriptor ServiceDescriptor) string {
	escrow := "no-escrow"
	if descriptor.Payment.EscrowRequired {
		escrow = "escrow"
	}
	return strings.Join([]string{
		string(descriptor.Payment.SettlementMode),
		descriptor.Payment.Denom,
		descriptor.Payment.Amount,
		string(descriptor.Payment.PricingUnit),
		descriptor.Payment.MaxAmount,
		escrow,
		descriptor.Payment.EscrowID,
		descriptor.Payment.MeterID,
	}, ":")
}

func isSupportedWalletSchemaFormat(format string) bool {
	return format == WalletSchemaFormatJSONV1
}

func isSupportedCLISchemaFormat(format string) bool {
	return format == CLISchemaFormatJSONV1
}
