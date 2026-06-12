package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceRequiredTestKind string
type ServiceUnitTestCaseID string
type ServiceIntegrationTestCaseID string
type ServiceInvariantTestCaseID string
type ServiceFuzzTestCaseID string
type ServicePerformanceTestCaseID string

const (
	ServiceRequiredTestUnit		ServiceRequiredTestKind	= "unit"
	ServiceRequiredTestIntegration	ServiceRequiredTestKind	= "integration"
	ServiceRequiredTestInvariant	ServiceRequiredTestKind	= "invariant"
	ServiceRequiredTestFuzz		ServiceRequiredTestKind	= "fuzz"
	ServiceRequiredTestPerformance	ServiceRequiredTestKind	= "performance"

	ServiceUnitCallIDDerivation	ServiceUnitTestCaseID	= "call_id_derivation"
	ServiceUnitDescriptorHash	ServiceUnitTestCaseID	= "descriptor_hash_calculation"
	ServiceUnitIdempotencyKey	ServiceUnitTestCaseID	= "idempotency_key_behavior"
	ServiceUnitInterfaceHash	ServiceUnitTestCaseID	= "interface_hash_calculation"
	ServiceUnitMethodIDValidation	ServiceUnitTestCaseID	= "method_id_validation"
	ServiceUnitNonceReplay		ServiceUnitTestCaseID	= "nonce_replay_rejection"
	ServiceUnitPaymentModel		ServiceUnitTestCaseID	= "payment_model_validation"
	ServiceUnitReceiptHash		ServiceUnitTestCaseID	= "receipt_hash_calculation"
	ServiceUnitServiceIDValidation	ServiceUnitTestCaseID	= "service_id_validation"
	ServiceUnitTrustModel		ServiceUnitTestCaseID	= "trust_model_validation"

	ServiceIntegrationAnchorOffChainResult		ServiceIntegrationTestCaseID	= "anchor_off_chain_service_result"
	ServiceIntegrationChallengeMixedResult		ServiceIntegrationTestCaseID	= "challenge_mixed_service_result"
	ServiceIntegrationExecuteOnChainCall		ServiceIntegrationTestCaseID	= "execute_on_chain_service_call"
	ServiceIntegrationGenerateReceiptProof		ServiceIntegrationTestCaseID	= "generate_and_verify_service_receipt_proof"
	ServiceIntegrationRegisterFogProvider		ServiceIntegrationTestCaseID	= "register_provider_for_fog_market_service"
	ServiceIntegrationRegisterInterfaceBinding	ServiceIntegrationTestCaseID	= "register_interface_and_bind_to_service"
	ServiceIntegrationRegisterMixedService		ServiceIntegrationTestCaseID	= "register_mixed_service"
	ServiceIntegrationRegisterOffChainAnchor	ServiceIntegrationTestCaseID	= "register_off_chain_service_anchor"
	ServiceIntegrationRegisterOnChainService	ServiceIntegrationTestCaseID	= "register_on_chain_service"
	ServiceIntegrationResolveAETBinding		ServiceIntegrationTestCaseID	= "resolve_service_through_aet_binding"
	ServiceIntegrationSettleEscrowPayment		ServiceIntegrationTestCaseID	= "settle_escrow_payment"

	ServiceInvariantActiveServiceInterface		ServiceInvariantTestCaseID	= "active_service_does_not_reference_missing_interface"
	ServiceInvariantCallReceiptServiceMethod	ServiceInvariantTestCaseID	= "call_receipt_references_existing_service_and_method"
	ServiceInvariantDescriptorStoredHash		ServiceInvariantTestCaseID	= "service_descriptor_hash_matches_stored_descriptor"
	ServiceInvariantExpiredServiceRejectsCalls	ServiceInvariantTestCaseID	= "expired_service_cannot_accept_new_calls"
	ServiceInvariantInterfaceRegisteredHash		ServiceInvariantTestCaseID	= "service_interface_hash_matches_registered_interface"
	ServiceInvariantPaymentSettlementEscrowLimit	ServiceInvariantTestCaseID	= "payment_settlement_does_not_exceed_escrow"
	ServiceInvariantProviderCollateralNonNegative	ServiceInvariantTestCaseID	= "provider_collateral_cannot_be_negative"
	ServiceInvariantReceiptRootCommittedReceipts	ServiceInvariantTestCaseID	= "receipt_root_includes_all_committed_receipts"

	ServiceFuzzDuplicateIdempotencyKeys	ServiceFuzzTestCaseID	= "duplicate_idempotency_keys"
	ServiceFuzzDuplicateNonces		ServiceFuzzTestCaseID	= "duplicate_nonces"
	ServiceFuzzForgedProviderSignatures	ServiceFuzzTestCaseID	= "forged_provider_signatures"
	ServiceFuzzInvalidDisputeProofs		ServiceFuzzTestCaseID	= "invalid_dispute_proofs"
	ServiceFuzzInvalidResultAnchors		ServiceFuzzTestCaseID	= "invalid_result_anchors"
	ServiceFuzzLargePayloadCalls		ServiceFuzzTestCaseID	= "large_payload_calls"
	ServiceFuzzMalformedDescriptors		ServiceFuzzTestCaseID	= "malformed_descriptors"
	ServiceFuzzMalformedInterfaceSchemas	ServiceFuzzTestCaseID	= "malformed_interface_schemas"
	ServiceFuzzPaymentEdgeCases		ServiceFuzzTestCaseID	= "payment_edge_cases"

	ServicePerformanceBlockSTMConflictRate		ServicePerformanceTestCaseID	= "blockstm_conflict_rate_for_independent_services"
	ServicePerformanceInterfaceLookupLatency	ServicePerformanceTestCaseID	= "interface_lookup_latency"
	ServicePerformanceOnChainExecutionThroughput	ServicePerformanceTestCaseID	= "on_chain_service_call_execution_throughput"
	ServicePerformanceProviderLookupLatency		ServicePerformanceTestCaseID	= "provider_lookup_latency"
	ServicePerformanceReceiptAnchoringThroughput	ServicePerformanceTestCaseID	= "receipt_anchoring_throughput"
	ServicePerformanceReceiptProofLatency		ServicePerformanceTestCaseID	= "receipt_proof_generation_latency"
	ServicePerformanceRegistryLookupLatency		ServicePerformanceTestCaseID	= "service_registry_direct_lookup_latency"
	ServicePerformanceServiceCallEnqueue		ServicePerformanceTestCaseID	= "service_call_enqueue_throughput"
)

type ServiceRequiredTestCase struct {
	Kind		ServiceRequiredTestKind
	UnitCase	ServiceUnitTestCaseID
	Integration	ServiceIntegrationTestCaseID
	Invariant	ServiceInvariantTestCaseID
	Fuzz		ServiceFuzzTestCaseID
	Performance	ServicePerformanceTestCaseID
	PackagePath	string
	TestName	string
	EvidenceHash	string
	CaseHash	string
}

type ServiceRequiredTestCoverage struct {
	UnitTests		[]ServiceRequiredTestCase
	IntegrationTests	[]ServiceRequiredTestCase
	InvariantTests		[]ServiceRequiredTestCase
	FuzzTests		[]ServiceRequiredTestCase
	PerformanceTests	[]ServiceRequiredTestCase
	CoverageHash		string
}

func DefaultServiceRequiredTestCoverage() (ServiceRequiredTestCoverage, error) {
	unit := []ServiceRequiredTestCase{
		newServiceUnitTestCase(ServiceUnitCallIDDerivation, "TestUnifiedCallIDDerivation", "ComputeUnifiedServiceCallID"),
		newServiceUnitTestCase(ServiceUnitDescriptorHash, "TestServiceDescriptorHashCalculation", "ComputeServiceDescriptorHash"),
		newServiceUnitTestCase(ServiceUnitIdempotencyKey, "TestServiceCallIdempotencyBehavior", "NewServiceCallIdempotencyRecord"),
		newServiceUnitTestCase(ServiceUnitInterfaceHash, "TestServiceInterfaceHashCalculation", "ComputeFormalServiceInterfaceHash"),
		newServiceUnitTestCase(ServiceUnitMethodIDValidation, "TestServiceMethodIDValidation", "ServiceInterfaceMethodSchema.Validate"),
		newServiceUnitTestCase(ServiceUnitNonceReplay, "TestServiceNonceReplayRejection", "ValidateServiceCallAnte"),
		newServiceUnitTestCase(ServiceUnitPaymentModel, "TestServicePaymentModelValidation", "ServicePaymentModel.Validate"),
		newServiceUnitTestCase(ServiceUnitReceiptHash, "TestServiceReceiptHashCalculation", "ComputeServiceCallReceiptHash"),
		newServiceUnitTestCase(ServiceUnitServiceIDValidation, "TestServiceIDValidation", "ServiceDescriptor.Validate"),
		newServiceUnitTestCase(ServiceUnitTrustModel, "TestServiceTrustModelValidation", "ServiceTrustModelSecurityRule.Validate"),
	}
	integration := []ServiceRequiredTestCase{
		newServiceIntegrationTestCase(ServiceIntegrationAnchorOffChainResult, "TestAnchorOffChainServiceResult", "MsgAnchorReceipt"),
		newServiceIntegrationTestCase(ServiceIntegrationChallengeMixedResult, "TestChallengeMixedServiceResult", "NewServiceChallengeFlow"),
		newServiceIntegrationTestCase(ServiceIntegrationExecuteOnChainCall, "TestExecuteOnChainServiceCall", "ValidateUnifiedServiceCallForDescriptor"),
		newServiceIntegrationTestCase(ServiceIntegrationGenerateReceiptProof, "TestGenerateAndVerifyServiceReceiptProof", "QueryReceiptProof"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterFogProvider, "TestRegisterProviderForFogMarketService", "MsgRegisterProvider"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterInterfaceBinding, "TestRegisterInterfaceAndBindToService", "MsgRegisterInterface/MsgUpdateService"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterMixedService, "TestRegisterMixedService", "MsgRegisterService"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterOffChainAnchor, "TestRegisterOffChainServiceAnchor", "ServiceAnchor"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterOnChainService, "TestRegisterOnChainService", "MsgRegisterService"),
		newServiceIntegrationTestCase(ServiceIntegrationResolveAETBinding, "TestResolveServiceThroughAETBinding", "IdentityServiceBinding"),
		newServiceIntegrationTestCase(ServiceIntegrationSettleEscrowPayment, "TestSettleEscrowPayment", "MsgSettleServiceEscrow"),
	}
	invariant := []ServiceRequiredTestCase{
		newServiceInvariantTestCase(ServiceInvariantActiveServiceInterface, "TestInvariantActiveServiceInterfaceExists", "ValidateRegistryInvariants"),
		newServiceInvariantTestCase(ServiceInvariantCallReceiptServiceMethod, "TestInvariantCallReceiptReferencesServiceMethod", "ServiceReceipt.Validate"),
		newServiceInvariantTestCase(ServiceInvariantDescriptorStoredHash, "TestInvariantDescriptorHashMatchesStoredDescriptor", "ComputeServiceDescriptorHash"),
		newServiceInvariantTestCase(ServiceInvariantExpiredServiceRejectsCalls, "TestInvariantExpiredServiceRejectsCalls", "ValidateXServicesDescriptorUsableForCall"),
		newServiceInvariantTestCase(ServiceInvariantInterfaceRegisteredHash, "TestInvariantInterfaceHashMatchesRegisteredInterface", "ComputeServiceInterfaceHash"),
		newServiceInvariantTestCase(ServiceInvariantPaymentSettlementEscrowLimit, "TestInvariantPaymentSettlementDoesNotExceedEscrow", "ValidateServiceEscrowFunding"),
		newServiceInvariantTestCase(ServiceInvariantProviderCollateralNonNegative, "TestInvariantProviderCollateralNonNegative", "ProviderCollateral.Validate"),
		newServiceInvariantTestCase(ServiceInvariantReceiptRootCommittedReceipts, "TestInvariantReceiptRootIncludesCommittedReceipts", "BuildReceiptRoot"),
	}
	fuzz := []ServiceRequiredTestCase{
		newServiceFuzzTestCase(ServiceFuzzDuplicateIdempotencyKeys, "FuzzServiceDuplicateIdempotencyKeys", "NewServiceCallIdempotencyRecord"),
		newServiceFuzzTestCase(ServiceFuzzDuplicateNonces, "FuzzServiceDuplicateNonces", "ValidateServiceCallAnte"),
		newServiceFuzzTestCase(ServiceFuzzForgedProviderSignatures, "FuzzServiceForgedProviderSignatures", "ValidateProviderFaultProof"),
		newServiceFuzzTestCase(ServiceFuzzInvalidDisputeProofs, "FuzzServiceInvalidDisputeProofs", "ServiceFaultProof.Validate"),
		newServiceFuzzTestCase(ServiceFuzzInvalidResultAnchors, "FuzzServiceInvalidResultAnchors", "MsgAnchorReceipt.ValidateBasic"),
		newServiceFuzzTestCase(ServiceFuzzLargePayloadCalls, "FuzzServiceLargePayloadCalls", "UnifiedServiceCall.ValidateBasic"),
		newServiceFuzzTestCase(ServiceFuzzMalformedDescriptors, "FuzzServiceMalformedDescriptors", "ServiceDescriptor.Validate"),
		newServiceFuzzTestCase(ServiceFuzzMalformedInterfaceSchemas, "FuzzServiceMalformedInterfaceSchemas", "ServiceInterfaceDefinition.Validate"),
		newServiceFuzzTestCase(ServiceFuzzPaymentEdgeCases, "FuzzServicePaymentEdgeCases", "ServicePaymentModel.Validate"),
	}
	performance := []ServiceRequiredTestCase{
		newServicePerformanceTestCase(ServicePerformanceBlockSTMConflictRate, "BenchmarkServiceBlockSTMConflictRate", "ServiceBlockSTMOperationsConflict"),
		newServicePerformanceTestCase(ServicePerformanceInterfaceLookupLatency, "BenchmarkServiceInterfaceLookupLatency", "QueryInterface"),
		newServicePerformanceTestCase(ServicePerformanceOnChainExecutionThroughput, "BenchmarkOnChainServiceCallExecutionThroughput", "ValidateUnifiedServiceCallForDescriptor"),
		newServicePerformanceTestCase(ServicePerformanceProviderLookupLatency, "BenchmarkProviderLookupLatency", "QueryProvidersByService"),
		newServicePerformanceTestCase(ServicePerformanceReceiptAnchoringThroughput, "BenchmarkReceiptAnchoringThroughput", "AnchorReceiptRecord"),
		newServicePerformanceTestCase(ServicePerformanceReceiptProofLatency, "BenchmarkReceiptProofGenerationLatency", "QueryReceiptProof"),
		newServicePerformanceTestCase(ServicePerformanceRegistryLookupLatency, "BenchmarkServiceRegistryDirectLookupLatency", "QueryService"),
		newServicePerformanceTestCase(ServicePerformanceServiceCallEnqueue, "BenchmarkServiceCallEnqueueThroughput", "AcceptUnifiedServiceCall"),
	}
	return NewServiceRequiredTestCoverage(unit, integration, invariant, fuzz, performance)
}

func NewServiceRequiredTestCoverage(unit []ServiceRequiredTestCase, integration []ServiceRequiredTestCase, invariant []ServiceRequiredTestCase, fuzz []ServiceRequiredTestCase, performance []ServiceRequiredTestCase) (ServiceRequiredTestCoverage, error) {
	coverage := ServiceRequiredTestCoverage{
		UnitTests:		cloneServiceRequiredTestCases(unit),
		IntegrationTests:	cloneServiceRequiredTestCases(integration),
		InvariantTests:		cloneServiceRequiredTestCases(invariant),
		FuzzTests:		cloneServiceRequiredTestCases(fuzz),
		PerformanceTests:	cloneServiceRequiredTestCases(performance),
	}
	sortServiceRequiredTestCases(coverage.UnitTests)
	sortServiceRequiredTestCases(coverage.IntegrationTests)
	sortServiceRequiredTestCases(coverage.InvariantTests)
	sortServiceRequiredTestCases(coverage.FuzzTests)
	sortServiceRequiredTestCases(coverage.PerformanceTests)
	if err := coverage.ValidateFormat(); err != nil {
		return ServiceRequiredTestCoverage{}, err
	}
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	return coverage, coverage.Validate()
}

func ValidateServiceRequiredTestEvidence(coverage ServiceRequiredTestCoverage) error {
	if err := coverage.Validate(); err != nil {
		return err
	}
	for _, testCase := range coverage.UnitTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.UnitCase)
		}
	}
	for _, testCase := range coverage.IntegrationTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.Integration)
		}
	}
	for _, testCase := range coverage.InvariantTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.Invariant)
		}
	}
	for _, testCase := range coverage.FuzzTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.Fuzz)
		}
	}
	for _, testCase := range coverage.PerformanceTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.Performance)
		}
	}
	return nil
}

func (coverage ServiceRequiredTestCoverage) ValidateFormat() error {
	if err := validateServiceRequiredUnitTests(coverage.UnitTests); err != nil {
		return err
	}
	if err := validateServiceRequiredIntegrationTests(coverage.IntegrationTests); err != nil {
		return err
	}
	if err := validateServiceRequiredInvariantTests(coverage.InvariantTests); err != nil {
		return err
	}
	if err := validateServiceRequiredFuzzTests(coverage.FuzzTests); err != nil {
		return err
	}
	if err := validateServiceRequiredPerformanceTests(coverage.PerformanceTests); err != nil {
		return err
	}
	if coverage.CoverageHash != "" {
		return coretypes.ValidateHash("services required test coverage hash", coverage.CoverageHash)
	}
	return nil
}

func (coverage ServiceRequiredTestCoverage) Validate() error {
	if err := coverage.ValidateFormat(); err != nil {
		return err
	}
	if coverage.CoverageHash == "" {
		return errors.New("services required test coverage hash is required")
	}
	if expected := ComputeServiceRequiredTestCoverageHash(coverage); coverage.CoverageHash != expected {
		return fmt.Errorf("services required test coverage hash mismatch: expected %s", expected)
	}
	return nil
}

func (testCase ServiceRequiredTestCase) Validate() error {
	if !IsServiceRequiredTestKind(testCase.Kind) {
		return fmt.Errorf("services required test unknown kind %q", testCase.Kind)
	}
	switch testCase.Kind {
	case ServiceRequiredTestUnit:
		if !IsServiceUnitTestCaseID(testCase.UnitCase) {
			return fmt.Errorf("services required test unknown unit case %q", testCase.UnitCase)
		}
		if testCase.Integration != "" || testCase.Invariant != "" || testCase.Fuzz != "" || testCase.Performance != "" {
			return errors.New("services required unit test cannot set another case id")
		}
	case ServiceRequiredTestIntegration:
		if !IsServiceIntegrationTestCaseID(testCase.Integration) {
			return fmt.Errorf("services required test unknown integration case %q", testCase.Integration)
		}
		if testCase.UnitCase != "" || testCase.Invariant != "" || testCase.Fuzz != "" || testCase.Performance != "" {
			return errors.New("services required integration test cannot set another case id")
		}
	case ServiceRequiredTestInvariant:
		if !IsServiceInvariantTestCaseID(testCase.Invariant) {
			return fmt.Errorf("services required test unknown invariant case %q", testCase.Invariant)
		}
		if testCase.UnitCase != "" || testCase.Integration != "" || testCase.Fuzz != "" || testCase.Performance != "" {
			return errors.New("services required invariant test cannot set another case id")
		}
	case ServiceRequiredTestFuzz:
		if !IsServiceFuzzTestCaseID(testCase.Fuzz) {
			return fmt.Errorf("services required test unknown fuzz case %q", testCase.Fuzz)
		}
		if testCase.UnitCase != "" || testCase.Integration != "" || testCase.Invariant != "" || testCase.Performance != "" {
			return errors.New("services required fuzz test cannot set another case id")
		}
	case ServiceRequiredTestPerformance:
		if !IsServicePerformanceTestCaseID(testCase.Performance) {
			return fmt.Errorf("services required test unknown performance case %q", testCase.Performance)
		}
		if testCase.UnitCase != "" || testCase.Integration != "" || testCase.Invariant != "" || testCase.Fuzz != "" {
			return errors.New("services required performance test cannot set another case id")
		}
	}
	if err := validateInterfaceToken("services required test package", testCase.PackagePath); err != nil {
		return err
	}
	if err := validateInterfaceToken("services required test name", testCase.TestName); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services required test evidence hash", testCase.EvidenceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services required test case hash", testCase.CaseHash); err != nil {
		return err
	}
	if expected := ComputeServiceRequiredTestCaseHash(testCase); testCase.CaseHash != expected {
		return fmt.Errorf("services required test case hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceRequiredTestCoverageHash(coverage ServiceRequiredTestCoverage) string {
	unit := cloneServiceRequiredTestCases(coverage.UnitTests)
	integration := cloneServiceRequiredTestCases(coverage.IntegrationTests)
	invariant := cloneServiceRequiredTestCases(coverage.InvariantTests)
	fuzz := cloneServiceRequiredTestCases(coverage.FuzzTests)
	performance := cloneServiceRequiredTestCases(coverage.PerformanceTests)
	sortServiceRequiredTestCases(unit)
	sortServiceRequiredTestCases(integration)
	sortServiceRequiredTestCases(invariant)
	sortServiceRequiredTestCases(fuzz)
	sortServiceRequiredTestCases(performance)
	parts := []string{"aetra-services-required-test-coverage-v1"}
	for _, testCase := range unit {
		parts = append(parts, "unit", testCase.CaseHash)
	}
	for _, testCase := range integration {
		parts = append(parts, "integration", testCase.CaseHash)
	}
	for _, testCase := range invariant {
		parts = append(parts, "invariant", testCase.CaseHash)
	}
	for _, testCase := range fuzz {
		parts = append(parts, "fuzz", testCase.CaseHash)
	}
	for _, testCase := range performance {
		parts = append(parts, "performance", testCase.CaseHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRequiredTestCaseHash(testCase ServiceRequiredTestCase) string {
	return servicesHashParts(
		"aetra-services-required-test-case-v1",
		string(testCase.Kind),
		string(testCase.UnitCase),
		string(testCase.Integration),
		string(testCase.Invariant),
		string(testCase.Fuzz),
		string(testCase.Performance),
		testCase.PackagePath,
		testCase.TestName,
		testCase.EvidenceHash,
	)
}

func IsServiceRequiredTestKind(kind ServiceRequiredTestKind) bool {
	switch kind {
	case ServiceRequiredTestUnit, ServiceRequiredTestIntegration, ServiceRequiredTestInvariant, ServiceRequiredTestFuzz, ServiceRequiredTestPerformance:
		return true
	default:
		return false
	}
}

func IsServiceUnitTestCaseID(caseID ServiceUnitTestCaseID) bool {
	switch caseID {
	case ServiceUnitCallIDDerivation, ServiceUnitDescriptorHash, ServiceUnitIdempotencyKey, ServiceUnitInterfaceHash,
		ServiceUnitMethodIDValidation, ServiceUnitNonceReplay, ServiceUnitPaymentModel, ServiceUnitReceiptHash,
		ServiceUnitServiceIDValidation, ServiceUnitTrustModel:
		return true
	default:
		return false
	}
}

func IsServiceIntegrationTestCaseID(caseID ServiceIntegrationTestCaseID) bool {
	switch caseID {
	case ServiceIntegrationAnchorOffChainResult, ServiceIntegrationChallengeMixedResult, ServiceIntegrationExecuteOnChainCall,
		ServiceIntegrationGenerateReceiptProof, ServiceIntegrationRegisterFogProvider, ServiceIntegrationRegisterInterfaceBinding,
		ServiceIntegrationRegisterMixedService, ServiceIntegrationRegisterOffChainAnchor, ServiceIntegrationRegisterOnChainService,
		ServiceIntegrationResolveAETBinding, ServiceIntegrationSettleEscrowPayment:
		return true
	default:
		return false
	}
}

func IsServiceInvariantTestCaseID(caseID ServiceInvariantTestCaseID) bool {
	switch caseID {
	case ServiceInvariantActiveServiceInterface, ServiceInvariantCallReceiptServiceMethod, ServiceInvariantDescriptorStoredHash,
		ServiceInvariantExpiredServiceRejectsCalls, ServiceInvariantInterfaceRegisteredHash, ServiceInvariantPaymentSettlementEscrowLimit,
		ServiceInvariantProviderCollateralNonNegative, ServiceInvariantReceiptRootCommittedReceipts:
		return true
	default:
		return false
	}
}

func IsServiceFuzzTestCaseID(caseID ServiceFuzzTestCaseID) bool {
	switch caseID {
	case ServiceFuzzDuplicateIdempotencyKeys, ServiceFuzzDuplicateNonces, ServiceFuzzForgedProviderSignatures,
		ServiceFuzzInvalidDisputeProofs, ServiceFuzzInvalidResultAnchors, ServiceFuzzLargePayloadCalls,
		ServiceFuzzMalformedDescriptors, ServiceFuzzMalformedInterfaceSchemas, ServiceFuzzPaymentEdgeCases:
		return true
	default:
		return false
	}
}

func IsServicePerformanceTestCaseID(caseID ServicePerformanceTestCaseID) bool {
	switch caseID {
	case ServicePerformanceBlockSTMConflictRate, ServicePerformanceInterfaceLookupLatency, ServicePerformanceOnChainExecutionThroughput,
		ServicePerformanceProviderLookupLatency, ServicePerformanceReceiptAnchoringThroughput, ServicePerformanceReceiptProofLatency,
		ServicePerformanceRegistryLookupLatency, ServicePerformanceServiceCallEnqueue:
		return true
	default:
		return false
	}
}

func newServiceUnitTestCase(caseID ServiceUnitTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:		ServiceRequiredTestUnit,
		UnitCase:	caseID,
		PackagePath:	"x/services/types",
		TestName:	strings.TrimSpace(testName),
		EvidenceHash:	servicesHashParts("aetra-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func newServiceIntegrationTestCase(caseID ServiceIntegrationTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:		ServiceRequiredTestIntegration,
		Integration:	caseID,
		PackagePath:	"x/services/types",
		TestName:	strings.TrimSpace(testName),
		EvidenceHash:	servicesHashParts("aetra-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func newServiceInvariantTestCase(caseID ServiceInvariantTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:		ServiceRequiredTestInvariant,
		Invariant:	caseID,
		PackagePath:	"x/services/types",
		TestName:	strings.TrimSpace(testName),
		EvidenceHash:	servicesHashParts("aetra-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func newServiceFuzzTestCase(caseID ServiceFuzzTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:		ServiceRequiredTestFuzz,
		Fuzz:		caseID,
		PackagePath:	"x/services/types",
		TestName:	strings.TrimSpace(testName),
		EvidenceHash:	servicesHashParts("aetra-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func newServicePerformanceTestCase(caseID ServicePerformanceTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:		ServiceRequiredTestPerformance,
		Performance:	caseID,
		PackagePath:	"x/services/types",
		TestName:	strings.TrimSpace(testName),
		EvidenceHash:	servicesHashParts("aetra-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func validateServiceRequiredUnitTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceUnitTestCaseID{
		ServiceUnitCallIDDerivation,
		ServiceUnitDescriptorHash,
		ServiceUnitIdempotencyKey,
		ServiceUnitInterfaceHash,
		ServiceUnitMethodIDValidation,
		ServiceUnitNonceReplay,
		ServiceUnitPaymentModel,
		ServiceUnitReceiptHash,
		ServiceUnitServiceIDValidation,
		ServiceUnitTrustModel,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d unit tests", len(required))
	}
	seen := map[ServiceUnitTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestUnit {
			return errors.New("services required unit coverage includes non-unit test")
		}
		current := string(testCase.UnitCase)
		if previous != "" && previous >= current {
			return errors.New("services required unit tests must be sorted canonically")
		}
		previous = current
		seen[testCase.UnitCase] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing unit test %s", caseID)
		}
	}
	return nil
}

func validateServiceRequiredInvariantTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceInvariantTestCaseID{
		ServiceInvariantActiveServiceInterface,
		ServiceInvariantCallReceiptServiceMethod,
		ServiceInvariantDescriptorStoredHash,
		ServiceInvariantExpiredServiceRejectsCalls,
		ServiceInvariantInterfaceRegisteredHash,
		ServiceInvariantPaymentSettlementEscrowLimit,
		ServiceInvariantProviderCollateralNonNegative,
		ServiceInvariantReceiptRootCommittedReceipts,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d invariant tests", len(required))
	}
	seen := map[ServiceInvariantTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestInvariant {
			return errors.New("services required invariant coverage includes non-invariant test")
		}
		current := string(testCase.Invariant)
		if previous != "" && previous >= current {
			return errors.New("services required invariant tests must be sorted canonically")
		}
		previous = current
		seen[testCase.Invariant] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing invariant test %s", caseID)
		}
	}
	return nil
}

func validateServiceRequiredFuzzTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceFuzzTestCaseID{
		ServiceFuzzDuplicateIdempotencyKeys,
		ServiceFuzzDuplicateNonces,
		ServiceFuzzForgedProviderSignatures,
		ServiceFuzzInvalidDisputeProofs,
		ServiceFuzzInvalidResultAnchors,
		ServiceFuzzLargePayloadCalls,
		ServiceFuzzMalformedDescriptors,
		ServiceFuzzMalformedInterfaceSchemas,
		ServiceFuzzPaymentEdgeCases,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d fuzz tests", len(required))
	}
	seen := map[ServiceFuzzTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestFuzz {
			return errors.New("services required fuzz coverage includes non-fuzz test")
		}
		current := string(testCase.Fuzz)
		if previous != "" && previous >= current {
			return errors.New("services required fuzz tests must be sorted canonically")
		}
		previous = current
		seen[testCase.Fuzz] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing fuzz test %s", caseID)
		}
	}
	return nil
}

func validateServiceRequiredPerformanceTests(testCases []ServiceRequiredTestCase) error {
	required := []ServicePerformanceTestCaseID{
		ServicePerformanceBlockSTMConflictRate,
		ServicePerformanceInterfaceLookupLatency,
		ServicePerformanceOnChainExecutionThroughput,
		ServicePerformanceProviderLookupLatency,
		ServicePerformanceReceiptAnchoringThroughput,
		ServicePerformanceReceiptProofLatency,
		ServicePerformanceRegistryLookupLatency,
		ServicePerformanceServiceCallEnqueue,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d performance tests", len(required))
	}
	seen := map[ServicePerformanceTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestPerformance {
			return errors.New("services required performance coverage includes non-performance test")
		}
		current := string(testCase.Performance)
		if previous != "" && previous >= current {
			return errors.New("services required performance tests must be sorted canonically")
		}
		previous = current
		seen[testCase.Performance] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing performance test %s", caseID)
		}
	}
	return nil
}

func validateServiceRequiredIntegrationTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceIntegrationTestCaseID{
		ServiceIntegrationAnchorOffChainResult,
		ServiceIntegrationChallengeMixedResult,
		ServiceIntegrationExecuteOnChainCall,
		ServiceIntegrationGenerateReceiptProof,
		ServiceIntegrationRegisterFogProvider,
		ServiceIntegrationRegisterInterfaceBinding,
		ServiceIntegrationRegisterMixedService,
		ServiceIntegrationRegisterOffChainAnchor,
		ServiceIntegrationRegisterOnChainService,
		ServiceIntegrationResolveAETBinding,
		ServiceIntegrationSettleEscrowPayment,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d integration tests", len(required))
	}
	seen := map[ServiceIntegrationTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestIntegration {
			return errors.New("services required integration coverage includes non-integration test")
		}
		current := string(testCase.Integration)
		if previous != "" && previous >= current {
			return errors.New("services required integration tests must be sorted canonically")
		}
		previous = current
		seen[testCase.Integration] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing integration test %s", caseID)
		}
	}
	return nil
}

func cloneServiceRequiredTestCases(testCases []ServiceRequiredTestCase) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, len(testCases))
	copy(out, testCases)
	return out
}

func sortServiceRequiredTestCases(testCases []ServiceRequiredTestCase) {
	sort.SliceStable(testCases, func(i, j int) bool {
		return serviceRequiredTestCaseSortKey(testCases[i]) < serviceRequiredTestCaseSortKey(testCases[j])
	})
}

func serviceRequiredTestCaseSortKey(testCase ServiceRequiredTestCase) string {
	switch testCase.Kind {
	case ServiceRequiredTestUnit:
		return string(testCase.UnitCase)
	case ServiceRequiredTestIntegration:
		return string(testCase.Integration)
	case ServiceRequiredTestInvariant:
		return string(testCase.Invariant)
	case ServiceRequiredTestFuzz:
		return string(testCase.Fuzz)
	case ServiceRequiredTestPerformance:
		return string(testCase.Performance)
	default:
		return ""
	}
}
