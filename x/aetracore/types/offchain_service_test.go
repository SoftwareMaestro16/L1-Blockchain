package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildOffChainServiceDescriptorSchema(t *testing.T) {
	schema, err := BuildOffChainServiceDescriptor(testOffChainDefinition())
	require.NoError(t, err)

	require.Equal(t, ServiceTypeOffChain, schema.Descriptor.ServiceType)
	require.Equal(t, ServiceLocationExternal, schema.Descriptor.Execution.Location)
	require.Equal(t, ExecutionModeAsync, schema.Descriptor.Execution.Mode)
	require.Equal(t, uint64(30), schema.Descriptor.Execution.ResultExpiry)
	require.Equal(t, ServiceVerificationProofAnchored, schema.Descriptor.Verification.Model)
	require.True(t, schema.Descriptor.Verification.RequestSigningRequired)
	require.True(t, schema.Descriptor.Verification.ResponseSigningRequired)
	require.NotEmpty(t, schema.Descriptor.Interface.InterfaceHash)
	require.Equal(t, OffChainEndpointRPC, schema.Endpoint.EndpointType)
	require.Equal(t, "provider.indexer.1", schema.Endpoint.ProviderKey)
	require.NotEmpty(t, schema.SchemaHash)
	require.NoError(t, schema.Validate())
}

func TestOffChainSignedAdvertisementAllowsOwnerOrProvider(t *testing.T) {
	schema, err := BuildOffChainServiceDescriptor(testOffChainDefinition())
	require.NoError(t, err)

	ownerAd, err := NewOffChainSignedAdvertisement(schema, DefaultAuthority)
	require.NoError(t, err)
	require.NoError(t, ownerAd.ValidateForSchema(schema))

	providerAd, err := NewOffChainSignedAdvertisement(schema, schema.Endpoint.ProviderKey)
	require.NoError(t, err)
	require.NoError(t, providerAd.ValidateForSchema(schema))

	providerAd.Signer = "provider.other"
	require.ErrorContains(t, providerAd.ValidateForSchema(schema), "signed by owner or provider")
}

func TestOffChainSignedRequestRequiresReplaySafeNonceOrIdempotency(t *testing.T) {
	request := testOffChainRequest()
	request.Nonce = 0
	request.IdempotencyKey = ""
	_, err := NewOffChainSignedRequest(request, request.Caller)
	require.ErrorContains(t, err, "replay-safe nonce or idempotency key")

	request.Nonce = 0
	request.IdempotencyKey = "query-idempotent-1"
	signed, err := NewOffChainSignedRequest(request, request.Caller)
	require.NoError(t, err)
	require.NotEmpty(t, signed.Request.RequestHash)
	require.NotEmpty(t, signed.SignatureHash)
	require.NoError(t, signed.Validate())
}

func TestOffChainSignedResponseAndReceiptAnchorBindProviderHeightAndHashes(t *testing.T) {
	request := testOffChainRequest()
	signedRequest, err := NewOffChainSignedRequest(request, request.Caller)
	require.NoError(t, err)

	response := OffChainServiceResponse{
		CallID:			request.CallID,
		ServiceID:		request.ServiceID,
		MethodID:		request.MethodID,
		RequestHash:		signedRequest.Request.RequestHash,
		ResponseHash:		testHash("indexer/response"),
		ProviderKey:		request.ProviderKey,
		Height:			42,
		ResultExpiryHeight:	72,
		SettlementUse:		true,
	}
	signedResponse, err := NewOffChainSignedResponse(response, response.ProviderKey)
	require.NoError(t, err)
	require.NoError(t, signedResponse.Validate())

	anchor, err := NewOffChainReceiptAnchorMessage(response, "")
	require.NoError(t, err)
	require.Equal(t, response.RequestHash, anchor.RequestHash)
	require.Equal(t, response.ResponseHash, anchor.ResponseHash)
	require.Equal(t, response.ProviderKey, anchor.ProviderKey)
	require.Equal(t, response.Height, anchor.Height)
	require.NotEmpty(t, anchor.ProofHash)
	require.NotEmpty(t, anchor.AnchorHash)
	require.NoError(t, anchor.Validate())

	anchor.ProviderKey = "provider.other"
	require.ErrorContains(t, anchor.Validate(), "proof hash mismatch")
}

func TestOffChainEndpointRenewalExtendsExpiryAndRequiresProviderSignature(t *testing.T) {
	schema, err := BuildOffChainServiceDescriptor(testOffChainDefinition())
	require.NoError(t, err)

	renewal, err := NewOffChainEndpointRenewal(schema, "https://indexer2.aetra.local/v1", 120, schema.Endpoint.ProviderKey)
	require.NoError(t, err)
	require.Equal(t, uint64(120), renewal.ExpiryHeight)
	require.NotEmpty(t, renewal.AdvertisementHash)
	require.NotEmpty(t, renewal.SignatureHash)
	require.NoError(t, renewal.ValidateForSchema(schema))

	renewal.Signer = DefaultAuthority
	require.ErrorContains(t, renewal.ValidateForSchema(schema), "signed by provider")

	_, err = NewOffChainEndpointRenewal(schema, "https://indexer3.aetra.local/v1", schema.Endpoint.ExpiryHeight, schema.Endpoint.ProviderKey)
	require.ErrorContains(t, err, "extend endpoint expiry")
}

func TestOffChainServiceDescriptorRejectsMissingResultExpiry(t *testing.T) {
	definition := testOffChainDefinition()
	definition.ResultExpiry = 0
	_, err := BuildOffChainServiceDescriptor(definition)
	require.ErrorContains(t, err, "result expiry")
}

func testOffChainDefinition() OffChainServiceDefinition {
	return OffChainServiceDefinition{
		ServiceID:		"indexer-feed-v2",
		Owner:			DefaultAuthority,
		ZoneID:			ZoneIDApplication,
		Endpoint:		"https://indexer.aetra.local/v1",
		EndpointType:		OffChainEndpointRPC,
		ProviderKey:		"provider.indexer.1",
		RequestSigningPolicy:	OffChainRequestCallerSigned,
		ResponseSigningPolicy:	OffChainResponseProviderSigned,
		ProofAnchorPolicy:	OffChainProofAnchorOptional,
		AvailabilityPolicy:	OffChainAvailabilitySignedAdvertisement,
		ResultExpiry:		30,
		InterfaceID:		"l1.indexer.v2.Query",
		InterfaceName:		"l1.indexer.v2.Query",
		EndpointKey:		"indexer.query.v2",
		Version:		1,
		AvailabilityHash:	testHash("indexer/v2/availability"),
		ProviderRoot:		testHash("indexer/v2/providers"),
		PaymentDenom:		NativeFeePolicyID,
		PaymentAmount:		"1",
		MetadataHash:		testHash("indexer/v2/metadata"),
		CreatedHeight:		1,
		UpdatedHeight:		1,
		ExpiryHeight:		90,
		Methods: []OffChainServiceMethod{
			testOffChainMethod("balance"),
			testOffChainMethod("txs"),
		},
	}
}

func testOffChainMethod(methodID string) OffChainServiceMethod {
	return OffChainServiceMethod{
		MethodID:		methodID,
		Name:			methodID,
		InputSchemaHash:	testHash(methodID + "/offchain/input"),
		OutputSchemaHash:	testHash(methodID + "/offchain/output"),
		RequiredPaymentModel:	DefaultOnChainPaymentModel,
		VerificationModel:	ServiceVerificationProofAnchored,
		TimeoutHeightDelta:	5,
		IdempotencyRequired:	true,
		CallbackSupported:	true,
		FailurePolicy:		ServiceFailureRetry,
	}
}

func testOffChainRequest() OffChainServiceRequest {
	return OffChainServiceRequest{
		CallID:		testHash("indexer/call/1"),
		ServiceID:	"indexer-feed-v2",
		MethodID:	"balance",
		Caller:		DefaultAuthority,
		Nonce:		7,
		IdempotencyKey:	"balance-query-7",
		PayloadHash:	testHash("indexer/request/payload"),
		ProviderKey:	"provider.indexer.1",
		DeadlineHeight:	50,
	}
}
