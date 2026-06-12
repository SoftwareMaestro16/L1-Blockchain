package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestUnifiedServiceCallBuildsEnvelopeAndMixedRouting(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	descriptor := testInterfaceSystemDescriptor()

	call, err := NewUnifiedServiceCall(
		ctx,
		descriptor,
		"submit",
		coretypes.DefaultAuthority,
		7,
		testInterfaceHash("submit/payload"),
		"9",
		testInterfaceHash("submit/signature"),
		12,
		"portable-service-submit-idem",
		"portable-service/callback",
	)
	require.NoError(t, err)
	require.Equal(t, descriptor.ServiceID, call.TargetService)
	require.Equal(t, "submit", call.Method)
	require.Equal(t, descriptor.Interface.InterfaceHash, call.InterfaceHash)
	require.Equal(t, coretypes.ServiceLocationHybrid, call.ExecutionLocation)
	require.Equal(t, coretypes.ServiceCallKindOffChainReceipt, call.Kind)
	require.True(t, call.Payment.Reserve)
	require.Equal(t, descriptor.Payment.Denom, call.Payment.Denom)
	require.NoError(t, call.ValidateBasic(ctx))

	envelope := call.ToServiceCallEnvelope()
	envelope = coretypes.NormalizeServiceCall(ctx, envelope)
	require.Equal(t, call.CallID, envelope.CallID)
	require.Equal(t, call.TargetService, envelope.ServiceID)
	require.True(t, envelope.Callback)
	require.NoError(t, envelope.ValidateBasic(ctx))

	plan, err := RouteUnifiedServiceCall(ctx, descriptor, call)
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceTypeMixed, plan.ServiceType)
	require.True(t, plan.OffChainExecutionRequired)
	require.True(t, plan.CommitResultOnChain)
	require.True(t, plan.ReserveFundsBeforeExecution)
	require.True(t, plan.VerifyResultProofBeforeAccept)
	require.True(t, plan.DisputeEligible)
	require.Contains(t, plan.Routes, UnifiedRouteServiceNetwork)
	require.Contains(t, plan.Routes, UnifiedRouteOnChainCommitment)
	require.Contains(t, plan.Routes, UnifiedRouteOnChainPayment)
	require.Contains(t, plan.Routes, UnifiedRouteProofVerification)
	require.NoError(t, plan.Validate())
}

func TestUnifiedServiceCallRoutesOnChainAndOffChainServices(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	onChain := testUnifiedOnChainDescriptor()

	onChainCall, err := NewUnifiedServiceCall(
		ctx,
		onChain,
		"query",
		coretypes.DefaultAuthority,
		1,
		testInterfaceHash("query/payload"),
		"3",
		testInterfaceHash("query/signature"),
		10,
		"identity-query-idem",
		"",
	)
	require.NoError(t, err)
	onChainPlan, err := RouteUnifiedServiceCall(ctx, onChain, onChainCall)
	require.NoError(t, err)
	require.True(t, onChainPlan.OnChainExecutionRequired)
	require.False(t, onChainPlan.OffChainExecutionRequired)
	require.Contains(t, onChainPlan.Routes, UnifiedRouteDeliverTx)
	require.Contains(t, onChainPlan.Routes, UnifiedRouteFinalizeBlock)
	require.Equal(t, UnifiedRouteFinalizeBlock, onChainPlan.ConsensusAcceptanceRoute)

	offChain := testUnifiedOffChainDescriptor()
	offChainCall, err := NewUnifiedServiceCall(
		ctx,
		offChain,
		"submit",
		coretypes.DefaultAuthority,
		2,
		testInterfaceHash("offchain/payload"),
		"4",
		testInterfaceHash("offchain/signature"),
		10,
		"indexer-submit-idem",
		"",
	)
	require.NoError(t, err)
	offChainPlan, err := RouteUnifiedServiceCall(ctx, offChain, offChainCall)
	require.NoError(t, err)
	require.False(t, offChainPlan.OnChainExecutionRequired)
	require.True(t, offChainPlan.OffChainExecutionRequired)
	require.False(t, offChainPlan.CommitResultOnChain)
	require.Contains(t, offChainPlan.Routes, UnifiedRouteServiceNetwork)
	require.Equal(t, UnifiedRouteServiceNetwork, offChainPlan.ConsensusAcceptanceRoute)
}

func TestUnifiedServiceCallRejectsRoutingRuleViolations(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	descriptor := testInterfaceSystemDescriptor()
	call, err := NewUnifiedServiceCall(
		ctx,
		descriptor,
		"submit",
		coretypes.DefaultAuthority,
		1,
		testInterfaceHash("submit/payload"),
		"9",
		testInterfaceHash("submit/signature"),
		12,
		"submit-idem",
		"",
	)
	require.NoError(t, err)

	noID := call
	noID.IdempotencyKey = ""
	noID.CallID = coretypes.NormalizeServiceCall(ctx, noID.ToServiceCallEnvelope()).CallID
	noID.UnifiedCallHash = ComputeUnifiedServiceCallHash(noID)
	require.ErrorContains(t, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, noID), "idempotency")

	wrongProof := call
	wrongProof.ProofRequirement = coretypes.ServiceVerificationAdvisory
	wrongProof.CallID = coretypes.NormalizeServiceCall(ctx, wrongProof.ToServiceCallEnvelope()).CallID
	wrongProof.UnifiedCallHash = ComputeUnifiedServiceCallHash(wrongProof)
	require.ErrorContains(t, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, wrongProof), "proof requirement")

	wrongPayment := call
	wrongPayment.Payment.Denom = "otherdenom"
	wrongPayment.Payment.PaymentHash = ComputeUnifiedServicePaymentHash(wrongPayment.Payment)
	wrongPayment.CallID = coretypes.NormalizeServiceCall(ctx, wrongPayment.ToServiceCallEnvelope()).CallID
	wrongPayment.UnifiedCallHash = ComputeUnifiedServiceCallHash(wrongPayment)
	require.ErrorContains(t, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, wrongPayment), "payment denom")

	expired := call
	expired.DeadlineHeight = descriptor.ExpiryHeight + 1
	expired.TimeoutHeight = expired.DeadlineHeight - expired.CreatedHeight
	expired.CallID = coretypes.NormalizeServiceCall(ctx, expired.ToServiceCallEnvelope()).CallID
	expired.UnifiedCallHash = ComputeUnifiedServiceCallHash(expired)
	require.ErrorContains(t, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, expired), "deadline")

	badRoute := UnifiedCallRoutingPlan{
		CallID:				call.CallID,
		TargetService:			call.TargetService,
		MethodID:			call.MethodID,
		ServiceType:			coretypes.ServiceTypeMixed,
		ExecutionLocation:		call.ExecutionLocation,
		Kind:				call.Kind,
		Routes:				[]UnifiedServiceRoute{UnifiedRouteOnChainCommitment, UnifiedRouteOnChainPayment, UnifiedRouteProofVerification},
		ReserveFundsBeforeExecution:	true,
		VerifyResultProofBeforeAccept:	true,
		CommitResultOnChain:		true,
		PaymentRouteRequired:		true,
		OffChainExecutionRequired:	true,
		ConsensusAcceptanceRoute:	UnifiedRouteOnChainCommitment,
	}
	badRoute.RoutingHash = ComputeUnifiedCallRoutingPlanHash(badRoute)
	require.ErrorContains(t, badRoute.Validate(), "mixed call")
}

func testUnifiedOnChainDescriptor() coretypes.ServiceDescriptor {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.ServiceID = "identity-unified"
	descriptor.ServiceType = coretypes.ServiceTypeOnChain
	descriptor.ZoneID = coretypes.ZoneIDIdentity
	descriptor.EndpointKey = "identity-unified.endpoint"
	descriptor.Version = 1
	descriptor.CreatedHeight = 10
	descriptor.UpdatedHeight = 10
	descriptor.ExpiryHeight = 100
	descriptor.Interface.InterfaceID = "l1.services.v1.IdentityUnified"
	descriptor.Interface.InterfaceName = "l1.services.v1.IdentityUnified"
	descriptor.Interface.Version = 1
	descriptor.Interface.Methods = []coretypes.ServiceMethodDescriptor{
		testInterfaceMethod("query", coretypes.ServiceMethodSync, coretypes.ServiceVerificationConsensusReceipt, coretypes.DefaultGasPolicy),
	}
	descriptor.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(descriptor.Interface)
	descriptor.InterfaceID = descriptor.Interface.InterfaceID
	descriptor.Execution = coretypes.ServiceExecutionDescriptor{
		Location:		coretypes.ServiceLocationModule,
		Target:			"identity.unified",
		ModuleRoute:		"identity",
		Mode:			coretypes.ExecutionModeSync,
		Deterministic:		true,
		FailureBehavior:	coretypes.ServiceFailureRevert,
		ReceiptPolicy:		coretypes.ServiceReceiptCommitted,
	}
	descriptor.Storage = coretypes.ServiceStorageDescriptor{Model: coretypes.ServiceStorageOnChain, StateRootType: coretypes.StateProofRootType, ProofRequired: true}
	descriptor.Verification = coretypes.ServiceVerificationDescriptor{TrustModel: coretypes.ServiceTrustConsensusExecuted, Model: coretypes.ServiceVerificationConsensusReceipt}
	return coretypes.CanonicalServiceDescriptor(descriptor)
}

func testUnifiedOffChainDescriptor() coretypes.ServiceDescriptor {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.ServiceID = "indexer-unified"
	descriptor.ServiceType = coretypes.ServiceTypeOffChain
	descriptor.EndpointKey = "indexer-unified.endpoint"
	descriptor.Version = 1
	descriptor.Interface.InterfaceID = "l1.services.v1.IndexerUnified"
	descriptor.Interface.InterfaceName = "l1.services.v1.IndexerUnified"
	descriptor.Interface.Version = 1
	descriptor.Interface.Methods = []coretypes.ServiceMethodDescriptor{
		testInterfaceMethod("submit", coretypes.ServiceMethodAsync, coretypes.ServiceVerificationSignedResult, ""),
	}
	descriptor.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(descriptor.Interface)
	descriptor.InterfaceID = descriptor.Interface.InterfaceID
	descriptor.Execution = coretypes.ServiceExecutionDescriptor{
		Location:		coretypes.ServiceLocationExternal,
		Target:			"indexer.unified",
		Endpoint:		"https://indexer-unified.aetra.local/v1",
		Mode:			coretypes.ExecutionModeAsync,
		FailureBehavior:	coretypes.ServiceFailureRetry,
		ResultExpiry:		20,
		ReceiptPolicy:		coretypes.ServiceReceiptCommitted,
	}
	descriptor.Storage = coretypes.ServiceStorageDescriptor{Model: coretypes.ServiceStorageEphemeral}
	descriptor.Verification = coretypes.ServiceVerificationDescriptor{
		TrustModel:			coretypes.ServiceTrustFullyTrusted,
		Model:				coretypes.ServiceVerificationSignedResult,
		RequestSigningRequired:		true,
		ResponseSigningRequired:	true,
	}
	return coretypes.CanonicalServiceDescriptor(descriptor)
}
