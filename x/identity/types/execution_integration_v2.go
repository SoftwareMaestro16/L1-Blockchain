package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityAuditMemoPrefixV2		= "aet:v2"
	MaxIdentityAuditMemoBytesV2		= 512
	DefaultIdentityFreshnessWindowV2	= uint64(20)
)

type IdentityStaleProofPolicyV2 struct {
	CurrentHeight		uint64
	ProofHeight		uint64
	FreshUntilHeight	uint64
	FreshnessThreshold	uint64
}

type IdentitySendByNameRequestV2 struct {
	Name			string
	State			IdentityState
	Height			uint64
	RecordTTL		uint64
	CurrentHeight		uint64
	FreshnessThreshold	uint64
	IncludeAuditMemo	bool

	ExpectedChainID	string
	TrustedHeader	IdentityTrustedHeaderV2
	Proof		*IdentityResolutionProofFormatV2
}

type IdentitySendByNameResultV2 struct {
	OriginalName		string
	NormalizedName		string
	Address			sdk.AccAddress
	ProofVerified		bool
	ProofHeight		uint64
	ProofStatus		string
	RecordVersion		uint64
	FreshUntilHeight	uint64
	StaleProofWarning	bool
	AuditMemo		string
	WalletDisplayLabel	string
}

type IdentityInvokeByNameRequestV2 struct {
	Name			string
	TargetID		string
	InterfaceID		string
	ExpectedInterfaceHash	string
	Method			string
	PayloadHash		string
	State			IdentityState
	Height			uint64
	RecordTTL		uint64
	CurrentHeight		uint64
	FreshnessThreshold	uint64

	ExpectedChainID	string
	TrustedHeader	IdentityTrustedHeaderV2
	Proof		*IdentityResolutionProofFormatV2
}

type IdentityInvokeByNameResultV2 struct {
	OriginalName			string
	NormalizedName			string
	TargetID			string
	ContractAddress			sdk.AccAddress
	Entrypoint			string
	Method				string
	PayloadHash			string
	InterfaceID			string
	InterfaceHash			string
	InterfaceDescriptorVerified	bool
	ProofVerified			bool
	ProofHeight			uint64
	RecordVersion			uint64
	FreshUntilHeight		uint64
	StaleProofWarning		bool
	StaleInterfaceDescriptorWarning	bool
	RequiresInterfaceConfirmation	bool
	SimulationRequiredBeforeSigning	bool
	ResolvedTarget			ContractTargetV2
	VerifiedInterfaceDescriptor	*InterfaceDescriptorV2
}

func BuildIdentitySendByNameV2(request IdentitySendByNameRequestV2) (IdentitySendByNameResultV2, error) {
	normalized, err := NormalizeAETDomain(request.Name)
	if err != nil {
		return IdentitySendByNameResultV2{}, err
	}
	if request.RecordTTL == 0 {
		request.RecordTTL = 1
	}
	if request.CurrentHeight == 0 {
		request.CurrentHeight = request.Height
	}
	out := IdentitySendByNameResultV2{
		OriginalName:	request.Name,
		NormalizedName:	normalized,
		ProofStatus:	"local_state",
	}
	if request.Proof != nil {
		target, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
			ExpectedChainID:	request.ExpectedChainID,
			RequestedName:		normalized,
			TrustedHeader:		request.TrustedHeader,
			Proof:			*request.Proof,
			TargetType:		IdentityResolutionTargetPrimary,
			CurrentHeight:		request.CurrentHeight,
			AllowRenewalWindow:	true,
			NormalizationVersion:	NameNormalizationVersionV2,
		})
		if err != nil {
			return IdentitySendByNameResultV2{}, err
		}
		out.Address = cloneSpecAddress(target.Address)
		out.ProofVerified = true
		out.ProofStatus = "verified"
		out.ProofHeight = target.ProofHeight
		out.RecordVersion = target.RecordVersion
		out.FreshUntilHeight = target.FreshUntilHeight
	} else {
		resolution, err := ResolveIdentityRecordRecursive(request.State, normalized, request.Height)
		if err != nil {
			return IdentitySendByNameResultV2{}, err
		}
		if len(resolution.Record.Primary) == 0 {
			return IdentitySendByNameResultV2{}, errors.New("identity send-by-name primary address is not resolved")
		}
		out.Address = cloneSpecAddress(resolution.Record.Primary)
		out.ProofHeight = request.Height
		out.RecordVersion = ResolverRecordVersionV2(resolution.Record)
		out.FreshUntilHeight = request.Height + request.RecordTTL
	}
	out.StaleProofWarning = EvaluateIdentityStaleProofWarningV2(IdentityStaleProofPolicyV2{
		CurrentHeight:		request.CurrentHeight,
		ProofHeight:		out.ProofHeight,
		FreshUntilHeight:	out.FreshUntilHeight,
		FreshnessThreshold:	request.FreshnessThreshold,
	})
	out.WalletDisplayLabel = fmt.Sprintf("%s -> %s", out.NormalizedName, hex.EncodeToString(out.Address))
	if request.IncludeAuditMemo {
		memo, err := BuildIdentityResolutionAuditMemoV2(out.NormalizedName, out.Address, out.ProofHeight, out.RecordVersion)
		if err != nil {
			return IdentitySendByNameResultV2{}, err
		}
		out.AuditMemo = memo
	}
	return out, nil
}

func ResolveIdentityContractTargetByNameV2(state IdentityState, name string, targetID string, height uint64, ttl uint64) (UnifiedResolutionRecordV2, ContractTargetV2, error) {
	if targetID == "" {
		targetID = ResolverKeyContract
	}
	if err := validateUnifiedRecordKey("identity v2 contract target_id", targetID); err != nil {
		return UnifiedResolutionRecordV2{}, ContractTargetV2{}, err
	}
	if ttl == 0 {
		ttl = 1
	}
	record, err := BuildUnifiedResolutionRecordV2(state, name, height, ttl)
	if err != nil {
		return UnifiedResolutionRecordV2{}, ContractTargetV2{}, err
	}
	target, found := ContractTargetFromUnifiedRecordV2(record, targetID)
	if !found {
		return UnifiedResolutionRecordV2{}, ContractTargetV2{}, fmt.Errorf("identity v2 contract target %q is not resolved", targetID)
	}
	if len(contractTargetAddressV2(target)) == 0 {
		return UnifiedResolutionRecordV2{}, ContractTargetV2{}, errors.New("identity v2 contract target address is not resolved")
	}
	return record, target, nil
}

func ContractTargetFromUnifiedRecordV2(record UnifiedResolutionRecordV2, targetID string) (ContractTargetV2, bool) {
	for _, target := range record.ContractTargets {
		if !contractTargetEnabledV2(target) {
			continue
		}
		if contractTargetIDV2(target) == targetID {
			return target, true
		}
	}
	return ContractTargetV2{}, false
}

func VerifyIdentityInterfaceDescriptorForInvokeV2(record UnifiedResolutionRecordV2, interfaceID string, expectedHash string) (*InterfaceDescriptorV2, error) {
	if interfaceID == "" {
		if expectedHash != "" {
			return nil, errors.New("identity v2 expected interface hash requires interface_id")
		}
		return nil, nil
	}
	for _, descriptor := range record.InterfaceDescriptors {
		if descriptor.InterfaceID != interfaceID {
			continue
		}
		schemaHash := interfaceDescriptorSchemaHashV2(descriptor)
		if err := ValidateInterfaceDescriptorHashFormatV2(schemaHash); err != nil {
			return nil, err
		}
		if expectedHash != "" && schemaHash != strings.ToLower(expectedHash) {
			return nil, errors.New("identity v2 interface descriptor hash mismatch")
		}
		verified := descriptor
		return &verified, nil
	}
	return nil, fmt.Errorf("identity v2 interface descriptor %q is not resolved", interfaceID)
}

func BuildIdentityInvokeByNameV2(request IdentityInvokeByNameRequestV2) (IdentityInvokeByNameResultV2, error) {
	normalized, err := NormalizeAETDomain(request.Name)
	if err != nil {
		return IdentityInvokeByNameResultV2{}, err
	}
	if request.Method == "" {
		return IdentityInvokeByNameResultV2{}, errors.New("identity invoke-by-name method is required")
	}
	if err := validateContractEntrypointV2(request.Method); err != nil {
		return IdentityInvokeByNameResultV2{}, err
	}
	if request.PayloadHash != "" {
		if err := validateHexHash("identity v2 invoke payload hash", request.PayloadHash); err != nil {
			return IdentityInvokeByNameResultV2{}, err
		}
	}
	if request.RecordTTL == 0 {
		request.RecordTTL = 1
	}
	if request.CurrentHeight == 0 {
		request.CurrentHeight = request.Height
	}
	targetID := request.TargetID
	if targetID == "" {
		targetID = ResolverKeyContract
	}

	var record UnifiedResolutionRecordV2
	var target ContractTargetV2
	var proofVerified bool
	var proofHeight uint64
	var recordVersion uint64
	var freshUntil uint64
	if request.Proof != nil {
		verifiedTarget, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
			ExpectedChainID:	request.ExpectedChainID,
			RequestedName:		normalized,
			TrustedHeader:		request.TrustedHeader,
			Proof:			*request.Proof,
			TargetType:		IdentityResolutionTargetContract,
			TargetKey:		targetID,
			CurrentHeight:		request.CurrentHeight,
			AllowRenewalWindow:	true,
			NormalizationVersion:	NameNormalizationVersionV2,
		})
		if err != nil {
			return IdentityInvokeByNameResultV2{}, err
		}
		if request.Proof.ResolverRecord == nil {
			return IdentityInvokeByNameResultV2{}, errors.New("identity v2 invoke proof resolver record is required")
		}
		record = *request.Proof.ResolverRecord
		foundTarget, found := ContractTargetFromUnifiedRecordV2(record, verifiedTarget.TargetKey)
		if !found {
			return IdentityInvokeByNameResultV2{}, errors.New("identity v2 verified contract target is missing from proof record")
		}
		target = foundTarget
		proofVerified = true
		proofHeight = verifiedTarget.ProofHeight
		recordVersion = verifiedTarget.RecordVersion
		freshUntil = verifiedTarget.FreshUntilHeight
	} else {
		var err error
		record, target, err = ResolveIdentityContractTargetByNameV2(request.State, normalized, targetID, request.Height, request.RecordTTL)
		if err != nil {
			return IdentityInvokeByNameResultV2{}, err
		}
		proofHeight = request.Height
		recordVersion = record.RecordVersion
		freshUntil = request.Height + request.RecordTTL
	}

	descriptor, err := VerifyIdentityInterfaceDescriptorForInvokeV2(record, request.InterfaceID, request.ExpectedInterfaceHash)
	if err != nil {
		return IdentityInvokeByNameResultV2{}, err
	}
	interfaceHash := target.InterfaceHash
	if descriptor != nil {
		interfaceHash = interfaceDescriptorSchemaHashV2(*descriptor)
	}
	entrypoint := target.Entrypoint
	if entrypoint == "" {
		entrypoint = request.Method
	}
	out := IdentityInvokeByNameResultV2{
		OriginalName:				request.Name,
		NormalizedName:				normalized,
		TargetID:				contractTargetIDV2(target),
		ContractAddress:			cloneSpecAddress(contractTargetAddressV2(target)),
		Entrypoint:				entrypoint,
		Method:					request.Method,
		PayloadHash:				request.PayloadHash,
		InterfaceID:				request.InterfaceID,
		InterfaceHash:				interfaceHash,
		InterfaceDescriptorVerified:		descriptor != nil,
		ProofVerified:				proofVerified,
		ProofHeight:				proofHeight,
		RecordVersion:				recordVersion,
		FreshUntilHeight:			freshUntil,
		StaleProofWarning:			EvaluateIdentityStaleProofWarningV2(IdentityStaleProofPolicyV2{CurrentHeight: request.CurrentHeight, ProofHeight: proofHeight, FreshUntilHeight: freshUntil, FreshnessThreshold: request.FreshnessThreshold}),
		StaleInterfaceDescriptorWarning:	EvaluateIdentityStaleProofWarningV2(IdentityStaleProofPolicyV2{CurrentHeight: request.CurrentHeight, ProofHeight: record.UpdatedAtHeight, FreshUntilHeight: freshUntil, FreshnessThreshold: request.FreshnessThreshold}),
		ResolvedTarget:				target,
		VerifiedInterfaceDescriptor:		descriptor,
	}
	for _, hint := range record.ExecutionHints {
		out.RequiresInterfaceConfirmation = out.RequiresInterfaceConfirmation || hint.RequiresInterfaceConfirmation
		out.SimulationRequiredBeforeSigning = out.SimulationRequiredBeforeSigning || hint.SimulationRequired
	}
	return out, nil
}

func BuildIdentityResolutionAuditMemoV2(name string, address sdk.AccAddress, proofHeight uint64, recordVersion uint64) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 audit memo address", address); err != nil {
		return "", err
	}
	if proofHeight == 0 {
		return "", errors.New("identity v2 audit memo proof height is required")
	}
	if recordVersion == 0 {
		return "", errors.New("identity v2 audit memo record version is required")
	}
	memo := fmt.Sprintf("%s;name=%s;height=%d;addr=%s;version=%d", IdentityAuditMemoPrefixV2, normalized, proofHeight, hex.EncodeToString(address), recordVersion)
	if len(memo) > MaxIdentityAuditMemoBytesV2 {
		return "", fmt.Errorf("identity v2 audit memo must not exceed %d bytes", MaxIdentityAuditMemoBytesV2)
	}
	return memo, nil
}

func EvaluateIdentityStaleProofWarningV2(policy IdentityStaleProofPolicyV2) bool {
	threshold := policy.FreshnessThreshold
	if threshold == 0 {
		threshold = DefaultIdentityFreshnessWindowV2
	}
	if policy.CurrentHeight == 0 || policy.ProofHeight == 0 {
		return true
	}
	if policy.FreshUntilHeight != 0 && policy.CurrentHeight > policy.FreshUntilHeight {
		return true
	}
	return policy.CurrentHeight > policy.ProofHeight+threshold
}
