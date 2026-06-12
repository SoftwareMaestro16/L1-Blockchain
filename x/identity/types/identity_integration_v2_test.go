package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityIntegrationAccessPathsCoverSectionEightOne(t *testing.T) {
	paths := IdentityIntegrationAccessPathsV2()
	require.Len(t, paths, 4)

	seen := make(map[IdentityAccessPathV2]struct{}, len(paths))
	for _, path := range paths {
		require.NotEmpty(t, path.UseCase)
		require.NotEmpty(t, path.ConsensusRequirement)
		seen[path.Path] = struct{}{}
	}

	for _, path := range []IdentityAccessPathV2{
		IdentityAccessPathCrossZoneMessage,
		IdentityAccessPathProofQuery,
		IdentityAccessPathVerifiedCache,
		IdentityAccessPathPreSigning,
	} {
		_, found := seen[path]
		require.True(t, found, "missing access path %s", path)
	}
}

func TestIdentityLookupMessagesAcceptSectionEightTwoTargetsAndStatuses(t *testing.T) {
	for _, target := range []IdentityLookupTargetType{
		IdentityLookupTargetAccount,
		IdentityLookupTargetContract,
		IdentityLookupTargetService,
		IdentityLookupTargetPayment,
		IdentityLookupTargetMetadata,
	} {
		req, err := NewMsgResolveIdentity(MsgResolveIdentity{
			RequestID:	identityHash("request", string(target)),
			Requester:	"contract:caller",
			SourceZoneID:	"CONTRACT_ZONE",
			TargetName:	"alice.aet",
			TargetType:	target,
			ProofRequired:	true,
			ReplyTo:	"contract:caller",
			ExpiryHeight:	50,
		})
		require.NoError(t, err)
		require.Equal(t, target, req.TargetType)
	}

	for _, status := range []IdentityResolutionStatus{
		IdentityResolutionStatusResolved,
		IdentityResolutionStatusNotFound,
		IdentityResolutionStatusExpired,
		IdentityResolutionStatusUnauthorized,
		IdentityResolutionStatusFailed,
	} {
		resolvedValue := ""
		if status == IdentityResolutionStatusResolved {
			resolvedValue = hex.EncodeToString([]byte("awallet1"))
		}
		result, err := NewMsgIdentityResolutionResult(MsgIdentityResolutionResult{
			RequestID:		identityHash("request-status", string(status)),
			Name:			"alice.aet",
			TargetType:		IdentityLookupTargetAccount,
			ResolvedValue:		resolvedValue,
			ResolverRecordVersion:	7,
			ProofHashOptional:	identityHash("proof", string(status)),
			Status:			status,
			ExpiryHeight:		60,
		})
		require.NoError(t, err)
		require.Equal(t, status, result.Status)
	}
}

func TestIdentityLookupExecutionPlanIsReadOnlyAndExpiryBound(t *testing.T) {
	req := testIdentityResolveRequestV2(t, true)
	plan, err := NewIdentityLookupExecutionPlanV2(req, 10, identityHash("request-receipt-root"))
	require.NoError(t, err)
	require.True(t, plan.ReadOnly)
	require.False(t, plan.MutatesIdentityState)
	require.Equal(t, IdentityZoneID, plan.IdentityZoneID)
	require.Equal(t, req.SourceZoneID, plan.ReplyToZoneID)
	require.NoError(t, plan.Validate())

	mutating := plan
	mutating.MutatesIdentityState = true
	mutating.PlanHash = ComputeIdentityLookupExecutionPlanHashV2(mutating)
	require.ErrorContains(t, mutating.Validate(), "read-only")

	_, err = NewIdentityLookupExecutionPlanV2(req, req.ExpiryHeight+1, "")
	require.ErrorContains(t, err, "expired")
}

func TestIdentityAsyncResolutionEnvelopeRequiresProofWhenRequested(t *testing.T) {
	req := testIdentityResolveRequestV2(t, true)
	result := testIdentityResolutionResultV2(t, req, IdentityResolutionStatusResolved, true)
	envelope, err := NewIdentityAsyncResolutionEnvelopeV2(req, result, identityHash("request-root"), identityHash("result-root"))
	require.NoError(t, err)
	require.Equal(t, IdentityZoneID, envelope.SourceZoneID)
	require.Equal(t, req.SourceZoneID, envelope.DestinationZoneID)
	require.Equal(t, req.ReplyTo, envelope.ReplyTo)
	require.NoError(t, envelope.Validate())

	withoutProof := result
	withoutProof.ProofHashOptional = ""
	withoutProof.ResultHash = ComputeMsgIdentityResolutionResultHash(withoutProof)
	_, err = NewIdentityAsyncResolutionEnvelopeV2(req, withoutProof, "", "")
	require.ErrorContains(t, err, "proof is required")

	otherRequest := req
	otherRequest.RequestID = identityHash("other-request")
	otherRequest.MessageHash = ComputeMsgResolveIdentityHash(otherRequest)
	_, err = NewIdentityAsyncResolutionEnvelopeV2(otherRequest, result, "", "")
	require.ErrorContains(t, err, "request id mismatch")
}

func TestIdentityVerifiedResolverCacheRequiresProofExpiryAndInvalidation(t *testing.T) {
	req := testIdentityResolveRequestV2(t, true)
	result := testIdentityResolutionResultV2(t, req, IdentityResolutionStatusResolved, true)
	entry, err := NewIdentityVerifiedResolverCacheEntryV2(result, 12, identityHash("trusted-app"), []IdentityCacheInvalidationTriggerV2{
		IdentityCacheInvalidResolverUpdateV2,
		IdentityCacheInvalidDomainTransferV2,
		IdentityCacheInvalidResolverUpdateV2,
	})
	require.NoError(t, err)
	require.Len(t, entry.InvalidationTriggers, 2)
	require.NoError(t, ValidateIdentityVerifiedResolverCacheUseV2(entry, IdentityCacheUseV2{
		Height:			20,
		ResolverRecordVersion:	result.ResolverRecordVersion,
		ProofVerified:		true,
	}))

	err = ValidateIdentityVerifiedResolverCacheUseV2(entry, IdentityCacheUseV2{
		Height:			result.ExpiryHeight + 1,
		ResolverRecordVersion:	result.ResolverRecordVersion,
		ProofVerified:		true,
	})
	require.ErrorContains(t, err, "expired")

	err = ValidateIdentityVerifiedResolverCacheUseV2(entry, IdentityCacheUseV2{
		Height:			20,
		ResolverRecordVersion:	result.ResolverRecordVersion + 1,
		ProofVerified:		true,
	})
	require.ErrorContains(t, err, "version changed")

	err = ValidateIdentityVerifiedResolverCacheUseV2(entry, IdentityCacheUseV2{
		Height:			20,
		ResolverRecordVersion:	result.ResolverRecordVersion,
		ProofVerified:		true,
		InvalidationTrigger:	IdentityCacheInvalidResolverUpdateV2,
	})
	require.ErrorContains(t, err, "invalidated")
}

func TestIdentityPreSigningBindingBindsResolvedValueIntoPayload(t *testing.T) {
	txHash := identityHash("tx-payload")
	first, err := NewIdentityPreSigningResolutionBindingV2("Alice", IdentityLookupTargetAccount, []byte("awallet1"), txHash)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", first.Name)
	require.NoError(t, first.Validate())

	second, err := NewIdentityPreSigningResolutionBindingV2("alice.aet", IdentityLookupTargetAccount, []byte("awallet2"), txHash)
	require.NoError(t, err)
	require.NotEqual(t, first.ResolvedValueHash, second.ResolvedValueHash)
	require.NotEqual(t, first.BoundPayloadHash, second.BoundPayloadHash)
}

func TestIdentityContractResolutionUseRequiresProofOrIdentityOriginReceipt(t *testing.T) {
	req := testIdentityResolveRequestV2(t, true)
	result := testIdentityResolutionResultV2(t, req, IdentityResolutionStatusResolved, true)

	require.NoError(t, ValidateIdentityContractResolutionUseV2(IdentityContractResolutionUseV2{
		Result:		result,
		CurrentHeight:	20,
		ProofVerified:	true,
	}))

	require.NoError(t, ValidateIdentityContractResolutionUseV2(IdentityContractResolutionUseV2{
		Result:			result,
		CurrentHeight:		20,
		MessageOriginZoneID:	IdentityZoneID,
		ReceiptRoot:		identityHash("receipt-root"),
		ReceiptHash:		identityHash("receipt-hash"),
	}))

	err := ValidateIdentityContractResolutionUseV2(IdentityContractResolutionUseV2{
		Result:			result,
		CurrentHeight:		20,
		MessageOriginZoneID:	"APPLICATION_ZONE",
		ReceiptRoot:		identityHash("receipt-root"),
		ReceiptHash:		identityHash("receipt-hash"),
	})
	require.ErrorContains(t, err, "verified proof or Identity Zone origin")

	err = ValidateIdentityContractResolutionUseV2(IdentityContractResolutionUseV2{
		Result:		result,
		CurrentHeight:	result.ExpiryHeight + 1,
		ProofVerified:	true,
	})
	require.ErrorContains(t, err, "expired")
}

func testIdentityResolveRequestV2(t *testing.T, proofRequired bool) MsgResolveIdentity {
	t.Helper()
	msg, err := NewMsgResolveIdentity(MsgResolveIdentity{
		RequestID:	identityHash("request-id", "alice"),
		Requester:	"contract:caller",
		SourceZoneID:	"CONTRACT_ZONE",
		TargetName:	"alice.aet",
		TargetType:	IdentityLookupTargetAccount,
		ProofRequired:	proofRequired,
		ReplyTo:	"contract:caller",
		ExpiryHeight:	50,
	})
	require.NoError(t, err)
	return msg
}

func testIdentityResolutionResultV2(t *testing.T, req MsgResolveIdentity, status IdentityResolutionStatus, withProof bool) MsgIdentityResolutionResult {
	t.Helper()
	proofHash := ""
	if withProof {
		proofHash = identityHash("proof", req.RequestID)
	}
	value := ""
	if status == IdentityResolutionStatusResolved {
		value = hex.EncodeToString([]byte("awallet1"))
	}
	result, err := NewMsgIdentityResolutionResult(MsgIdentityResolutionResult{
		RequestID:		req.RequestID,
		Name:			"alice.aet",
		TargetType:		req.TargetType,
		ResolvedValue:		value,
		ResolverRecordVersion:	7,
		ProofHashOptional:	proofHash,
		Status:			status,
		ExpiryHeight:		80,
	})
	require.NoError(t, err)
	return result
}
