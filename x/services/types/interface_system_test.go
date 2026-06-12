package types

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestFormalServiceInterfaceProjectsExtendedFields(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	require.NoError(t, err)
	require.Equal(t, descriptor.Interface.InterfaceHash, definition.InterfaceHash)
	require.Equal(t, descriptor.Interface.InterfaceName, definition.InterfaceName)
	require.Equal(t, uint64(1), definition.Version)
	require.Len(t, definition.Methods, 3)
	require.Equal(t, []string{"BalanceChanged", "StreamOpened"}, definition.Events)
	require.Equal(t, []string{"InvalidRequest"}, definition.Errors)
	require.Equal(t, "owner-or-provider", definition.AuthModel)
	require.Equal(t, "prepaid:naet:1", definition.PaymentRequirements)
	require.Equal(t, "json-schema-v1", definition.SchemaEncoding)
	require.Equal(t, descriptor.Interface.MetadataHash, definition.MetadataHash)
	require.Equal(t, descriptor.Interface.CreatedHeight, definition.CreatedHeight)
	require.True(t, definition.SupportsExecutionType(coretypes.ServiceMethodSync))
	require.True(t, definition.SupportsExecutionType(coretypes.ServiceMethodAsync))
	require.True(t, definition.SupportsExecutionType(coretypes.ServiceMethodEvented))
	require.NoError(t, definition.Validate())
}

func TestPrepareServiceInterfaceCallUnifiesExecutionLocations(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()

	query, err := PrepareServiceInterfaceCall(descriptor, "query", coretypes.DefaultAuthority, 1, testInterfaceHash("payload/query"), 20)
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceMethodSync, query.ExecutionType)
	require.Equal(t, coretypes.DefaultGasPolicy, query.GasModel)
	require.False(t, query.EventStream)
	require.Equal(t, "prepaid:naet:1", query.PaymentRequirements)
	require.NoError(t, query.Validate())

	submit, err := PrepareServiceInterfaceCall(descriptor, "submit", coretypes.DefaultAuthority, 2, testInterfaceHash("payload/submit"), 20)
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceMethodAsync, submit.ExecutionType)
	require.False(t, submit.EventStream)
	require.Equal(t, coretypes.ServiceVerificationSignedResult, submit.VerificationModel)

	stream, err := PrepareServiceInterfaceCall(descriptor, "stream", coretypes.DefaultAuthority, 3, testInterfaceHash("payload/stream"), 20)
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceMethodEvented, stream.ExecutionType)
	require.True(t, stream.EventStream)
	require.Equal(t, ComputeServiceInterfaceCallPreparationHash(stream), stream.PreparationHash)
}

func TestPrepareServiceInterfaceCallRejectsMalformedRequests(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()

	_, err := PrepareServiceInterfaceCall(descriptor, "missing", coretypes.DefaultAuthority, 1, testInterfaceHash("payload"), 20)
	require.ErrorContains(t, err, "not found")

	_, err = PrepareServiceInterfaceCall(descriptor, "query", coretypes.DefaultAuthority, 0, testInterfaceHash("payload"), 20)
	require.ErrorContains(t, err, "nonce")

	_, err = PrepareServiceInterfaceCall(descriptor, "query", coretypes.DefaultAuthority, 1, "bad", 20)
	require.ErrorContains(t, err, "payload")

	definition, err := NewFormalServiceInterface(descriptor.Interface)
	require.NoError(t, err)
	broken := definition
	broken.Methods[0].MethodHash = testInterfaceHash("wrong-method")
	require.ErrorContains(t, broken.Validate(), "method hash mismatch")
}

func TestServiceMethodSchemaIncludesTimeoutIdempotencyAndCallback(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	require.NoError(t, err)

	method, found := definition.MethodByName("submit")
	require.True(t, found)
	require.Equal(t, uint64(10), method.TimeoutPolicy.TimeoutHeightDelta)
	require.Equal(t, coretypes.ServiceFailureRetry, method.TimeoutPolicy.FailureBehavior)
	require.True(t, method.IdempotencyRequired)
	require.True(t, method.CallbackSupported)
	require.NotEmpty(t, method.TimeoutPolicy.PolicyHash)
	require.NoError(t, method.Validate())

	changed := method
	changed.IdempotencyRequired = false
	require.NotEqual(t, method.MethodHash, ComputeServiceInterfaceMethodSchemaHash(changed))
}

func TestServiceInterfaceValidationRules(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	require.NoError(t, ValidateServiceInterfaceForDescriptor(descriptor))

	duplicateName := descriptor
	duplicateName.Interface.Methods[1].Name = duplicateName.Interface.Methods[0].Name
	duplicateName.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(duplicateName.Interface)
	require.ErrorContains(t, ValidateServiceInterfaceForDescriptor(duplicateName), "duplicate")

	badEncoding := testInterfaceSystemDescriptor()
	badEncoding.Interface.SchemaEncoding = "yaml-v1"
	badEncoding.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(badEncoding.Interface)
	require.ErrorContains(t, ValidateServiceInterfaceForDescriptor(badEncoding), "not supported")

	offChain := testInterfaceSystemDescriptor()
	offChain.ServiceType = coretypes.ServiceTypeOffChain
	offChain.Execution.Location = coretypes.ServiceLocationExternal
	offChain.Verification.TrustModel = coretypes.ServiceTrustFullyTrusted
	offChain.Verification.Model = coretypes.ServiceVerificationSignedResult
	offChain.Interface.Methods[0].VerificationModel = coretypes.ServiceVerificationAdvisory
	offChain.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(offChain.Interface)
	require.ErrorContains(t, ValidateServiceInterfaceForDescriptor(offChain), "response verification")
}

func TestServiceInterfaceVersionChangeRequiresNewHash(t *testing.T) {
	previous := testInterfaceSystemDescriptor().Interface
	next := previous
	next.Version = 2
	next.InterfaceName = "l1.services.v2.Portable"
	next.InterfaceHash = coretypes.ComputeServiceInterfaceHash(next)

	require.NoError(t, ValidateServiceInterfaceVersionChange(previous, next))

	bad := previous
	bad.Version = 2
	require.ErrorContains(t, ValidateServiceInterfaceVersionChange(previous, bad), "new interface hash")
}

func testInterfaceSystemDescriptor() coretypes.ServiceDescriptor {
	methods := []coretypes.ServiceMethodDescriptor{
		testInterfaceMethod("query", coretypes.ServiceMethodSync, coretypes.ServiceVerificationConsensusReceipt, coretypes.DefaultGasPolicy),
		testInterfaceMethod("submit", coretypes.ServiceMethodAsync, coretypes.ServiceVerificationSignedResult, ""),
		testInterfaceMethod("stream", coretypes.ServiceMethodEvented, coretypes.ServiceVerificationSignedResult, ""),
	}
	iface := coretypes.ServiceInterfaceDescriptor{
		InterfaceID:	"l1.services.v1.Portable",
		InterfaceName:	"l1.services.v1.Portable",
		Version:	1,
		SchemaEncoding:	"json-schema-v1",
		Methods:	methods,
		Events:		[]string{"StreamOpened", "BalanceChanged"},
		Errors:		[]string{"InvalidRequest"},
		AuthModel:	"owner-or-provider",
		PaymentModel:	"prepaid:naet:1",
		MetadataHash:	testInterfaceHash("interface/metadata"),
		CreatedHeight:	7,
	}
	iface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(iface)
	descriptor := coretypes.ServiceDescriptor{
		ServiceID:		"portable-service",
		Owner:			coretypes.DefaultAuthority,
		ServiceType:		coretypes.ServiceTypeMixed,
		ZoneID:			coretypes.ZoneIDApplication,
		InterfaceID:		iface.InterfaceID,
		EndpointKey:		"portable.endpoint",
		Version:		1,
		AvailabilityHash:	testInterfaceHash("portable/availability"),
		Enabled:		true,
		Status:			coretypes.ServiceStatusActive,
		ExpiryHeight:		100,
		CreatedHeight:		7,
		UpdatedHeight:		7,
		Interface:		iface,
		Execution: coretypes.ServiceExecutionDescriptor{
			Location:		coretypes.ServiceLocationHybrid,
			Target:			"portable.target",
			Endpoint:		"https://portable.aetra.local/v1",
			Mode:			coretypes.ExecutionModeAsync,
			FailureBehavior:	coretypes.ServiceFailureChallenge,
			ResultExpiry:		30,
			ChallengeWindow:	10,
		},
		Discovery: coretypes.ServiceDiscoveryDescriptor{
			ServiceName:		"portable-service",
			ProviderRoot:		testInterfaceHash("portable/providers"),
			MetadataHash:		testInterfaceHash("portable/metadata"),
			CacheExpiryHeight:	90,
			SignaturePolicy:	"provider-signature-v1",
		},
		Payment: coretypes.ServicePaymentDescriptor{
			SettlementMode:	coretypes.ServicePaymentPrepaid,
			Denom:		coretypes.NativeFeePolicyID,
			Amount:		"1",
			PricingUnit:	coretypes.ServicePricingPerCall,
		},
		Storage:	coretypes.ServiceStorageDescriptor{Model: coretypes.ServiceStorageHybridCommitment, CommitmentHash: testInterfaceHash("portable/storage"), ProofRequired: true},
		Verification: coretypes.ServiceVerificationDescriptor{
			TrustModel:			coretypes.ServiceTrustHybridChallengeable,
			Model:				coretypes.ServiceVerificationChallengeWindow,
			RequestSigningRequired:		true,
			ResponseSigningRequired:	true,
			ChallengeWindow:		10,
			FaultPolicy:			coretypes.ServiceFailureChallenge,
		},
	}
	return coretypes.CanonicalServiceDescriptor(descriptor)
}

func testInterfaceMethod(methodID string, executionType coretypes.ServiceMethodExecutionType, verification coretypes.ServiceVerificationModel, gasModel string) coretypes.ServiceMethodDescriptor {
	return coretypes.ServiceMethodDescriptor{
		MethodID:		methodID,
		Name:			methodID,
		InputSchemaHash:	testInterfaceHash(methodID + "/input"),
		OutputSchemaHash:	testInterfaceHash(methodID + "/output"),
		ExecutionType:		executionType,
		RequiredPaymentModel:	"prepaid:naet:1",
		GasModel:		gasModel,
		VerificationModel:	verification,
		TimeoutHeightDelta:	10,
		IdempotencyRequired:	true,
		CallbackSupported:	executionType == coretypes.ServiceMethodAsync,
		FailureBehavior:	coretypes.ServiceFailureRetry,
	}
}

func testInterfaceHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
