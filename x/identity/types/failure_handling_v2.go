package types

type IdentityLightClientFailureHandlingV2 struct {
	FailureCode					IdentityLightClientFailureCodeV2
	RejectTargetUse					bool
	RequestFreshProof				bool
	DirectAddressFallbackRequiresConfirmation	bool
	DirectAddressFallbackAllowed			bool
	CacheInvalid					bool
	InvalidatedCache				*ResolutionCacheRecordV2
}

type IdentityWalletResolutionStatusV2 struct {
	VerifiedStatusVisible	bool
	AutoFillTargetAllowed	bool
	FreshnessWarning	bool
	RejectTargetUse		bool
	FailureCode		IdentityLightClientFailureCodeV2
}

func HandleIdentityLightClientFailureV2(err error, cached *ResolutionCacheRecordV2, directAddressConfirmed bool) IdentityLightClientFailureHandlingV2 {
	code, found := IdentityLightClientFailureCodeFromErrorV2(err)
	if !found {
		code = IdentityLightClientErrProofInvalid
	}
	out := IdentityLightClientFailureHandlingV2{
		FailureCode:		code,
		RejectTargetUse:	true,
		RequestFreshProof:	true,
		DirectAddressFallbackRequiresConfirmation:	true,
		DirectAddressFallbackAllowed:			directAddressConfirmed,
	}
	if cached != nil {
		invalidated := InvalidateResolutionCacheRecordV2ForDomainMutation(*cached, cached.SourceVersion, cached.ParentEpoch, cached.ChildEpoch)
		out.CacheInvalid = true
		out.InvalidatedCache = &invalidated
	}
	return out
}

func EvaluateIdentityWalletResolutionStatusV2(target *IdentityLightClientVerifiedTargetV2, proofErr error, localHeight uint64, freshnessThreshold uint64) IdentityWalletResolutionStatusV2 {
	if proofErr != nil || target == nil {
		code, found := IdentityLightClientFailureCodeFromErrorV2(proofErr)
		if !found {
			code = IdentityLightClientErrProofInvalid
		}
		return IdentityWalletResolutionStatusV2{
			VerifiedStatusVisible:	false,
			AutoFillTargetAllowed:	false,
			RejectTargetUse:	true,
			FailureCode:		code,
		}
	}
	status := IdentityWalletResolutionStatusV2{
		VerifiedStatusVisible:	true,
		AutoFillTargetAllowed:	true,
	}
	if freshnessThreshold > 0 && localHeight > target.ProofHeight && localHeight-target.ProofHeight > freshnessThreshold {
		status.FreshnessWarning = true
	}
	return status
}

func queryFailureCodeV2(code IdentityQueryCodeV2) IdentityLightClientFailureCodeV2 {
	switch code {
	case IdentityQueryInvalidRequest:
		return IdentityLightClientErrInvalidName
	case IdentityQueryNotFound:
		return IdentityLightClientErrDomainNotFound
	case IdentityQueryExpired:
		return IdentityLightClientErrDomainExpired
	case IdentityQueryVerificationFailed:
		return IdentityLightClientErrProofInvalid
	default:
		return IdentityLightClientErrProofInvalid
	}
}
