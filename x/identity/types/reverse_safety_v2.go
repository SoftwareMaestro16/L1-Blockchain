package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityReverseWalletDisplayStateV2 string

const (
	IdentityReverseWalletDisplayNoneV2	IdentityReverseWalletDisplayStateV2	= "none"
	IdentityReverseWalletDisplayClaimedV2	IdentityReverseWalletDisplayStateV2	= "claimed_unverified"
	IdentityReverseWalletDisplayVerifiedV2	IdentityReverseWalletDisplayStateV2	= "verified"
	IdentityReverseWalletDisplayInvalidV2	IdentityReverseWalletDisplayStateV2	= "invalid"
	IdentityReverseWalletDisplayExpiredV2	IdentityReverseWalletDisplayStateV2	= "expired"
	IdentityReverseWalletDisplaySeparatedV2	IdentityReverseWalletDisplayStateV2	= "separate_unverified"
	IdentityReverseInvalidationPrimaryV2						= "primary_resolver_change"
	IdentityReverseInvalidationTransferV2						= "domain_transfer"
	IdentityReverseInvalidationExpiryV2						= "domain_expiry"
	IdentityReverseInvalidationResolverV2						= "resolver_update"
	IdentityReverseProofFormatVersionV2						= uint64(1)
)

type IdentityReverseWalletDisplayV2 struct {
	Address				sdk.AccAddress
	Name				string
	State				IdentityReverseWalletDisplayStateV2
	Verified			bool
	DisplayAsCanonical		bool
	DisplaySeparately		bool
	RequiresForwardProof		bool
	ForwardConsistencyVerified	bool
	Warning				string
}

type VerifiedReverseResolutionProofV2 struct {
	ProofVersion	uint64
	Address		sdk.AccAddress
	Name		string
	NameHash	string
	Height		uint64
	Record		ReverseResolutionRecordV2
	Proof		IdentityResolutionProofFormatV2
	ProofHash	string
}

func InvalidateReverseRecordsForResolverUpdateV2(state IdentityState, name string, height uint64, primaryChanged bool, authorizedAliasKeys []string) (IdentityState, []ReverseRecord, error) {
	if !primaryChanged {
		return InvalidateReverseRecordsForDomainV2(state, name, height, authorizedAliasKeys)
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, nil, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	next := state.Clone()
	kept := make([]ReverseRecord, 0, len(state.ReverseRecords))
	invalidated := make([]ReverseRecord, 0)
	for _, reverse := range state.ReverseRecords {
		if reverse.Domain == normalized {
			invalidated = append(invalidated, reverse)
			continue
		}
		kept = append(kept, cloneReverseRecord(reverse))
	}
	next.ReverseRecords = kept
	sortIdentityState(&next)
	if err := next.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	return next, invalidated, nil
}

func InvalidateReverseRecordsForDomainTransferV2(state IdentityState, name string, height uint64) (IdentityState, []ReverseRecord, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, nil, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	next := state.Clone()
	kept := make([]ReverseRecord, 0, len(state.ReverseRecords))
	invalidated := make([]ReverseRecord, 0)
	for _, reverse := range state.ReverseRecords {
		if reverse.Domain == normalized {
			invalidated = append(invalidated, reverse)
			continue
		}
		kept = append(kept, cloneReverseRecord(reverse))
	}
	next.ReverseRecords = kept
	sortIdentityState(&next)
	if err := next.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	return next, invalidated, nil
}

func InvalidateReverseRecordsForDomainExpiryV2(state IdentityState, height uint64) (IdentityState, []ReverseRecord, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	next := state.Clone()
	kept := make([]ReverseRecord, 0, len(state.ReverseRecords))
	invalidated := make([]ReverseRecord, 0)
	for _, reverse := range state.ReverseRecords {
		domain, found := findDomain(state, reverse.Domain)
		if !found || domain.ExpiryHeight <= height {
			invalidated = append(invalidated, reverse)
			continue
		}
		kept = append(kept, cloneReverseRecord(reverse))
	}
	next.ReverseRecords = kept
	sortIdentityState(&next)
	if err := next.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	return next, invalidated, nil
}

func BuildVerifiedReverseResolutionProofV2(state IdentityState, address sdk.AccAddress, chainID string, appHash string, height uint64, ttl uint64, authorizedAliasKeys []string) (VerifiedReverseResolutionProofV2, error) {
	record, proof, err := reverseProofForFormatV2(state, address, height, authorizedAliasKeys)
	if err != nil {
		return VerifiedReverseResolutionProofV2{}, err
	}
	if !record.Verified {
		return VerifiedReverseResolutionProofV2{}, errors.New("identity v2 verified reverse proof requires verified record")
	}
	formatProof, err := BuildIdentityResolutionProofFormatV2(state, chainID, appHash, record.Name, IdentityProofQueryResolveReverse, height, ttl, address)
	if err != nil {
		return VerifiedReverseResolutionProofV2{}, err
	}
	if formatProof.ReverseRecordOptional == nil || formatProof.ReverseRecordProofOptional == nil {
		return VerifiedReverseResolutionProofV2{}, errors.New("identity v2 verified reverse proof missing reverse record")
	}
	if proof.RootHash != formatProof.AppHash {
		return VerifiedReverseResolutionProofV2{}, errors.New("identity v2 verified reverse proof root mismatch")
	}
	out := VerifiedReverseResolutionProofV2{
		ProofVersion:	IdentityReverseProofFormatVersionV2,
		Address:	cloneSpecAddress(address),
		Name:		record.Name,
		NameHash:	record.NameHash,
		Height:		height,
		Record:		record,
		Proof:		formatProof,
	}
	out.ProofHash = ComputeVerifiedReverseResolutionProofHashV2(out)
	return out, nil
}

func ValidateVerifiedReverseResolutionProofV2(proof VerifiedReverseResolutionProofV2, request IdentityLightClientVerificationRequestV2) error {
	if proof.ProofVersion != IdentityReverseProofFormatVersionV2 {
		return errors.New("identity v2 unsupported verified reverse proof version")
	}
	if !proof.Record.Verified {
		return errors.New("identity v2 verified reverse proof record is unverified")
	}
	if proof.ProofHash == "" || proof.ProofHash != ComputeVerifiedReverseResolutionProofHashV2(proof) {
		return errors.New("identity v2 verified reverse proof hash mismatch")
	}
	request.Proof = proof.Proof
	request.RequireReverseResolution = true
	request.TargetType = IdentityResolutionTargetPrimary
	request.TargetKey = ResolverKeyPrimary
	_, err := VerifyIdentityResolutionProofLightClientV2(request)
	return err
}

func ComputeVerifiedReverseResolutionProofHashV2(proof VerifiedReverseResolutionProofV2) string {
	return identityHash(
		"identity-v2-verified-reverse-proof",
		formatUint64V2(proof.ProofVersion),
		string(proof.Address),
		proof.Name,
		proof.NameHash,
		formatUint64V2(proof.Height),
		proof.Record.NameHash,
		ComputeIdentityResolutionProofCommitmentHashV2(proof.Proof),
	)
}

func BuildIdentityReverseWalletDisplayStateV2(state IdentityState, address sdk.AccAddress, height uint64, authorizedAliasKeys []string) IdentityReverseWalletDisplayV2 {
	out := IdentityReverseWalletDisplayV2{Address: cloneSpecAddress(address), State: IdentityReverseWalletDisplayNoneV2}
	for _, reverse := range state.ReverseRecords {
		if !addressesEqual(reverse.Address, address) {
			continue
		}
		out.Name = reverse.Domain
		record, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err != nil {
			out.State = IdentityReverseWalletDisplayInvalidV2
			out.Warning = err.Error()
			return out
		}
		if record.ExpiryHeight <= height {
			out.State = IdentityReverseWalletDisplayExpiredV2
			out.DisplaySeparately = true
			out.Warning = "reverse record is expired"
			return out
		}
		if err := ValidateReverseResolutionRecordV2(state, record, height, authorizedAliasKeys); err != nil {
			unverified := record
			unverified.Verified = false
			if claimErr := ValidateReverseResolutionRecordV2(state, unverified, height, authorizedAliasKeys); claimErr == nil {
				out.State = IdentityReverseWalletDisplaySeparatedV2
				out.DisplaySeparately = true
				out.RequiresForwardProof = true
				out.Warning = "reverse record is an unverified claim"
				return out
			}
			out.State = IdentityReverseWalletDisplayInvalidV2
			out.Warning = err.Error()
			return out
		}
		out.State = IdentityReverseWalletDisplayVerifiedV2
		out.Verified = true
		out.DisplayAsCanonical = true
		out.ForwardConsistencyVerified = true
		return out
	}
	return out
}

func BuildIdentityReverseWalletDisplayFromRecordV2(state IdentityState, record ReverseResolutionRecordV2, height uint64, authorizedAliasKeys []string) IdentityReverseWalletDisplayV2 {
	out := IdentityReverseWalletDisplayV2{
		Address:	cloneSpecAddress(record.Address),
		Name:		record.Name,
	}
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		out.State = IdentityReverseWalletDisplayInvalidV2
		out.Warning = err.Error()
		return out
	}
	if record.ExpiryHeight <= height {
		out.State = IdentityReverseWalletDisplayExpiredV2
		out.DisplaySeparately = true
		out.Warning = "reverse record is expired"
		return out
	}
	if !record.Verified {
		out.State = IdentityReverseWalletDisplaySeparatedV2
		out.DisplaySeparately = true
		out.RequiresForwardProof = true
		out.Warning = "reverse record is an unverified claim"
		return out
	}
	if err := ValidateReverseResolutionRecordV2(state, record, height, authorizedAliasKeys); err != nil {
		out.State = IdentityReverseWalletDisplayInvalidV2
		out.Warning = err.Error()
		return out
	}
	out.State = IdentityReverseWalletDisplayVerifiedV2
	out.Verified = true
	out.DisplayAsCanonical = true
	out.ForwardConsistencyVerified = true
	return out
}

func formatUint64V2(value uint64) string {
	return fmt.Sprintf("%020d", value)
}
