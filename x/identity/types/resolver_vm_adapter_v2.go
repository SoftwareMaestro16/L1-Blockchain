package types

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultVMResolverGasLimit	uint64	= 1_000_000
	DefaultVMResolverMemoryBytes	uint64	= 64 * 1024
	DefaultVMResolverMaxOutputBytes	uint64	= 4 * 1024
	DefaultVMResolverProofChecks	uint32	= 8
	DefaultVMResolverRecursionDepth	uint32	= 2
)

type IdentityProofRequirementV2 string
type IdentityIntegrationTaskIDV2 string
type IdentityIntegrationPriorityV2 string

const (
	IdentityProofRequirementDomainOwnership		IdentityProofRequirementV2	= "domain_ownership"
	IdentityProofRequirementNFTBinding		IdentityProofRequirementV2	= "nft_binding"
	IdentityProofRequirementDomainStatus		IdentityProofRequirementV2	= "domain_status"
	IdentityProofRequirementExpiry			IdentityProofRequirementV2	= "expiry"
	IdentityProofRequirementResolverRecord		IdentityProofRequirementV2	= "resolver_record"
	IdentityProofRequirementReverseLookup		IdentityProofRequirementV2	= "reverse_lookup"
	IdentityProofRequirementDelegationGrants	IdentityProofRequirementV2	= "delegation_or_grants"
	IdentityProofRequirementAuctionFinality		IdentityProofRequirementV2	= "auction_finalization"

	IdentityIntegrationPriorityP0	IdentityIntegrationPriorityV2	= "P0"
	IdentityIntegrationPriorityP1	IdentityIntegrationPriorityV2	= "P1"
	IdentityIntegrationPriorityP2	IdentityIntegrationPriorityV2	= "P2"

	IdentityTaskIsolatedZone		IdentityIntegrationTaskIDV2	= "isolated-identity-zone-state-machine"
	IdentityTaskCrossZoneLookupMessages	IdentityIntegrationTaskIDV2	= "cross-zone-identity-lookup-messages"
	IdentityTaskResolverProofAPIs		IdentityIntegrationTaskIDV2	= "resolver-proof-apis"
	IdentityTaskVMResolverAdapter		IdentityIntegrationTaskIDV2	= "vm-resolver-adapter"
	IdentityTaskReverseLookupProof		IdentityIntegrationTaskIDV2	= "reverse-lookup-proof"
	IdentityTaskCacheInvalidationMessages	IdentityIntegrationTaskIDV2	= "identity-cache-invalidation-messages"
	IdentityTaskWalletSDKHelpers		IdentityIntegrationTaskIDV2	= "wallet-sdk-send-invoke-by-name"
)

type NativeResolverRecordV2 struct {
	Name			string
	NameHash		string
	Owner			sdk.AccAddress
	ResolverRecordVersion	uint64
	ExpiryHeight		uint64
	TargetType		IdentityLookupTargetType
	TargetValue		[]byte
	ProofKey		string
	RecordHash		string
}

type VMResolverExecutionLimitsV2 struct {
	GasLimit		uint64
	MemoryBytes		uint64
	MaxOutputBytes		uint64
	MaxProofChecks		uint32
	MaxRecursionDepth	uint32
}

type VMResolverContractContextV2 struct {
	NameHash		string
	DomainOwner		sdk.AccAddress
	DomainStatus		DomainRecordV2Status
	DomainExpiryHeight	uint64
	ResolverRecordVersion	uint64
	CodeID			uint64
	ContractAddress		string
	TargetType		IdentityLookupTargetType
	Height			uint64
	Limits			VMResolverExecutionLimitsV2
	ContextHash		string
}

type VMResolverContractOutputV2 struct {
	NameHash		string
	TargetType		IdentityLookupTargetType
	ResolvedValue		[]byte
	ResolverRecordVersion	uint64
	GasUsed			uint64
	MemoryUsedBytes		uint64
	ProofChecks		uint32
	RecursionDepth		uint32
	OwnerOverride		sdk.AccAddress
	Status			IdentityResolutionStatus
	OutputHash		string
}

type VMResolverContractProofV2 struct {
	Height			uint64
	Name			string
	NameHash		string
	CodeID			uint64
	ContractAddress		string
	ResolverRoot		string
	NativeRecordHash	string
	OutputHash		string
	ProofHash		string
}

type VMResolverFallbackPolicyV2 struct {
	AllowNativeFallback	bool
	FallbackStatus		IdentityResolutionStatus
}

type VMResolverAdapterResultV2 struct {
	NativeRecord	NativeResolverRecordV2
	Context		VMResolverContractContextV2
	Output		VMResolverContractOutputV2
	Proof		VMResolverContractProofV2
	UsedFallback	bool
	Status		IdentityResolutionStatus
	ResultHash	string
}

type IdentityProofRequirementDescriptorV2 struct {
	Requirement		IdentityProofRequirementV2
	VerifiedCondition	string
}

type IdentityIntegrationTaskV2 struct {
	Priority		IdentityIntegrationPriorityV2
	TaskID			IdentityIntegrationTaskIDV2
	Target			string
	AcceptanceCriteria	string
}

func DefaultVMResolverExecutionLimitsV2() VMResolverExecutionLimitsV2 {
	return VMResolverExecutionLimitsV2{
		GasLimit:		DefaultVMResolverGasLimit,
		MemoryBytes:		DefaultVMResolverMemoryBytes,
		MaxOutputBytes:		DefaultVMResolverMaxOutputBytes,
		MaxProofChecks:		DefaultVMResolverProofChecks,
		MaxRecursionDepth:	DefaultVMResolverRecursionDepth,
	}
}

func NewNativeResolverRecordV2(domain DomainRecordV2, resolver ResolverRecord, targetType IdentityLookupTargetType, targetValue []byte, proofKey string) (NativeResolverRecordV2, error) {
	if !IsIdentityLookupTargetType(targetType) {
		return NativeResolverRecordV2{}, fmt.Errorf("unknown identity lookup target type %q", targetType)
	}
	if len(targetValue) == 0 {
		return NativeResolverRecordV2{}, errors.New("native resolver target value is required")
	}
	if !bytes.Equal(resolver.Owner, domain.Owner) {
		return NativeResolverRecordV2{}, errors.New("native resolver owner must match domain owner")
	}
	if proofKey == "" {
		return NativeResolverRecordV2{}, errors.New("native resolver proof key is required")
	}
	record := NativeResolverRecordV2{
		Name:			domain.Name,
		NameHash:		domain.NameHash,
		Owner:			cloneSpecAddress(domain.Owner),
		ResolverRecordVersion:	ResolverRecordVersionV2(resolver),
		ExpiryHeight:		domain.ExpiryHeight,
		TargetType:		targetType,
		TargetValue:		append([]byte(nil), targetValue...),
		ProofKey:		proofKey,
	}
	record.RecordHash = ComputeNativeResolverRecordHashV2(record)
	return record, record.Validate()
}

func NewVMResolverContractContextV2(native NativeResolverRecordV2, domain DomainRecordV2, codeID uint64, contractAddress string, height uint64, limits VMResolverExecutionLimitsV2) (VMResolverContractContextV2, error) {
	if err := native.Validate(); err != nil {
		return VMResolverContractContextV2{}, err
	}
	if codeID == 0 {
		return VMResolverContractContextV2{}, errors.New("VM resolver code id must be positive")
	}
	if contractAddress == "" {
		return VMResolverContractContextV2{}, errors.New("VM resolver contract address is required")
	}
	if height == 0 {
		return VMResolverContractContextV2{}, errors.New("VM resolver context height must be positive")
	}
	if limits.GasLimit == 0 {
		limits = DefaultVMResolverExecutionLimitsV2()
	}
	ctx := VMResolverContractContextV2{
		NameHash:		native.NameHash,
		DomainOwner:		cloneSpecAddress(domain.Owner),
		DomainStatus:		domain.Status,
		DomainExpiryHeight:	domain.ExpiryHeight,
		ResolverRecordVersion:	native.ResolverRecordVersion,
		CodeID:			codeID,
		ContractAddress:	contractAddress,
		TargetType:		native.TargetType,
		Height:			height,
		Limits:			limits,
	}
	ctx.ContextHash = ComputeVMResolverContractContextHashV2(ctx)
	return ctx, ctx.Validate()
}

func NewVMResolverContractOutputV2(output VMResolverContractOutputV2) (VMResolverContractOutputV2, error) {
	output.ResolvedValue = append([]byte(nil), output.ResolvedValue...)
	output.OwnerOverride = cloneSpecAddress(output.OwnerOverride)
	output.OutputHash = ComputeVMResolverContractOutputHashV2(output)
	return output, output.ValidateFormat()
}

func EvaluateVMResolverContractOutputV2(native NativeResolverRecordV2, ctx VMResolverContractContextV2, output VMResolverContractOutputV2, resolverRoot string, fallback VMResolverFallbackPolicyV2) (VMResolverAdapterResultV2, error) {
	if err := native.Validate(); err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	if err := ctx.Validate(); err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	if resolverRoot == "" {
		return VMResolverAdapterResultV2{}, errors.New("VM resolver proof resolver root is required")
	}
	if err := validateHexHash("VM resolver proof resolver root", resolverRoot); err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	if err := ValidateVMResolverNativeAuthorityV2(native, ctx); err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	resolvedOutput, usedFallback, status, err := resolveVMResolverOutputOrFallback(native, ctx, output, fallback)
	if err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	proof, err := NewVMResolverContractProofV2(VMResolverContractProofV2{
		Height:			ctx.Height,
		Name:			native.Name,
		NameHash:		native.NameHash,
		CodeID:			ctx.CodeID,
		ContractAddress:	ctx.ContractAddress,
		ResolverRoot:		resolverRoot,
		NativeRecordHash:	native.RecordHash,
		OutputHash:		resolvedOutput.OutputHash,
	})
	if err != nil {
		return VMResolverAdapterResultV2{}, err
	}
	result := VMResolverAdapterResultV2{
		NativeRecord:	native,
		Context:	ctx,
		Output:		resolvedOutput,
		Proof:		proof,
		UsedFallback:	usedFallback,
		Status:		status,
	}
	result.ResultHash = ComputeVMResolverAdapterResultHashV2(result)
	return result, result.Validate()
}

func NewVMResolverContractProofV2(proof VMResolverContractProofV2) (VMResolverContractProofV2, error) {
	proof.ProofHash = ComputeVMResolverContractProofHashV2(proof)
	return proof, proof.Validate()
}

func IdentityProofRequirementsV2() []IdentityProofRequirementDescriptorV2 {
	return []IdentityProofRequirementDescriptorV2{
		{Requirement: IdentityProofRequirementDomainOwnership, VerifiedCondition: "current owner, ownership version, and domain key match the committed Identity Zone root"},
		{Requirement: IdentityProofRequirementNFTBinding, VerifiedCondition: "domain ownership token or binding exists and matches the domain record"},
		{Requirement: IdentityProofRequirementDomainStatus, VerifiedCondition: "domain lifecycle status is committed at proof height"},
		{Requirement: IdentityProofRequirementExpiry, VerifiedCondition: "record is valid at proof height and has not passed expiry or grace rules"},
		{Requirement: IdentityProofRequirementResolverRecord, VerifiedCondition: "resolver value, version, target type, and expiry match committed resolver state"},
		{Requirement: IdentityProofRequirementReverseLookup, VerifiedCondition: "address-to-name mapping is committed and authorized by forward record or owner"},
		{Requirement: IdentityProofRequirementDelegationGrants, VerifiedCondition: "delegate permissions, scopes, and expiry authorize the requested action"},
		{Requirement: IdentityProofRequirementAuctionFinality, VerifiedCondition: "auction outcome, winner, settlement, and finalization height are committed when relevant"},
	}
}

func IdentityIntegrationTasksV2() []IdentityIntegrationTaskV2 {
	return []IdentityIntegrationTaskV2{
		{Priority: IdentityIntegrationPriorityP0, TaskID: IdentityTaskIsolatedZone, Target: "x/identity v2 zone adapter", AcceptanceCriteria: "domain, resolver, reverse, delegation, auction, and NFT binding state commit under Identity Zone roots"},
		{Priority: IdentityIntegrationPriorityP0, TaskID: IdentityTaskCrossZoneLookupMessages, Target: "message layer and Identity Zone", AcceptanceCriteria: "MsgResolveIdentity and MsgIdentityResolutionResult are async, receipt-backed, expiry-bounded, and replay-safe"},
		{Priority: IdentityIntegrationPriorityP0, TaskID: IdentityTaskResolverProofAPIs, Target: "Identity Zone queries", AcceptanceCriteria: "resolver proofs verify target value, type, version, expiry, and resolver root"},
		{Priority: IdentityIntegrationPriorityP1, TaskID: IdentityTaskVMResolverAdapter, Target: "Contract Zone and Identity Zone", AcceptanceCriteria: "resolver contracts execute with bounded gas and cannot override native ownership or lifecycle state"},
		{Priority: IdentityIntegrationPriorityP1, TaskID: IdentityTaskReverseLookupProof, Target: "Identity proof API", AcceptanceCriteria: "address-to-name proofs verify reverse record, owner authorization, expiry, and forward binding when required"},
		{Priority: IdentityIntegrationPriorityP1, TaskID: IdentityTaskCacheInvalidationMessages, Target: "message layer and cache state", AcceptanceCriteria: "resolver cache entries expire or invalidate through committed height, version, and invalidation messages"},
		{Priority: IdentityIntegrationPriorityP2, TaskID: IdentityTaskWalletSDKHelpers, Target: "client SDK", AcceptanceCriteria: "SDK resolves .aet names, binds proof or resolved value into the transaction, and handles expiry"},
	}
}

func (r NativeResolverRecordV2) Validate() error {
	normalized, err := NormalizeAETDomain(r.Name)
	if err != nil {
		return err
	}
	if r.Name != normalized {
		return errors.New("native resolver record name must be normalized")
	}
	if expected, err := DomainRecordV2NameHash(r.Name); err != nil {
		return err
	} else if r.NameHash != expected {
		return errors.New("native resolver record name_hash mismatch")
	}
	if err := validateSpecAddress("native resolver owner", r.Owner); err != nil {
		return err
	}
	if r.ResolverRecordVersion == 0 || r.ExpiryHeight == 0 {
		return errors.New("native resolver record version and expiry are required")
	}
	if !IsIdentityLookupTargetType(r.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", r.TargetType)
	}
	if len(r.TargetValue) == 0 {
		return errors.New("native resolver target value is required")
	}
	if r.ProofKey == "" {
		return errors.New("native resolver proof key is required")
	}
	if err := validateHexHash("native resolver record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeNativeResolverRecordHashV2(r) {
		return errors.New("native resolver record hash mismatch")
	}
	return nil
}

func (l VMResolverExecutionLimitsV2) Validate() error {
	if l.GasLimit == 0 || l.MemoryBytes == 0 || l.MaxOutputBytes == 0 || l.MaxProofChecks == 0 || l.MaxRecursionDepth == 0 {
		return errors.New("VM resolver limits must be positive")
	}
	return nil
}

func (c VMResolverContractContextV2) Validate() error {
	if err := validateHexHash("VM resolver context name hash", c.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("VM resolver context domain owner", c.DomainOwner); err != nil {
		return err
	}
	if !IsDomainRecordV2Status(c.DomainStatus) {
		return fmt.Errorf("invalid VM resolver domain status %q", c.DomainStatus)
	}
	if c.DomainExpiryHeight == 0 || c.ResolverRecordVersion == 0 || c.CodeID == 0 || c.Height == 0 {
		return errors.New("VM resolver context expiry, version, code id, and height are required")
	}
	if c.Height > c.DomainExpiryHeight {
		return errors.New("VM resolver context domain expired")
	}
	if c.ContractAddress == "" {
		return errors.New("VM resolver contract address is required")
	}
	if !IsIdentityLookupTargetType(c.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", c.TargetType)
	}
	if err := c.Limits.Validate(); err != nil {
		return err
	}
	if err := validateHexHash("VM resolver context hash", c.ContextHash); err != nil {
		return err
	}
	if c.ContextHash != ComputeVMResolverContractContextHashV2(c) {
		return errors.New("VM resolver context hash mismatch")
	}
	return nil
}

func (o VMResolverContractOutputV2) ValidateFormat() error {
	if err := validateHexHash("VM resolver output name hash", o.NameHash); err != nil {
		return err
	}
	if !IsIdentityLookupTargetType(o.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", o.TargetType)
	}
	if o.ResolverRecordVersion == 0 {
		return errors.New("VM resolver output record version is required")
	}
	if !IsIdentityResolutionStatus(o.Status) {
		return fmt.Errorf("unknown VM resolver output status %q", o.Status)
	}
	if o.Status == IdentityResolutionStatusResolved && len(o.ResolvedValue) == 0 {
		return errors.New("VM resolver resolved output requires value")
	}
	if err := validateHexHash("VM resolver output hash", o.OutputHash); err != nil {
		return err
	}
	if o.OutputHash != ComputeVMResolverContractOutputHashV2(o) {
		return errors.New("VM resolver output hash mismatch")
	}
	return nil
}

func (p VMResolverContractProofV2) Validate() error {
	if p.Height == 0 || p.CodeID == 0 {
		return errors.New("VM resolver proof height and code id are required")
	}
	if _, err := NormalizeAETDomain(p.Name); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"VM resolver proof name hash", p.NameHash},
		{"VM resolver proof resolver root", p.ResolverRoot},
		{"VM resolver proof native record hash", p.NativeRecordHash},
		{"VM resolver proof output hash", p.OutputHash},
		{"VM resolver proof hash", p.ProofHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if p.ContractAddress == "" {
		return errors.New("VM resolver proof contract address is required")
	}
	if p.ProofHash != ComputeVMResolverContractProofHashV2(p) {
		return errors.New("VM resolver proof hash mismatch")
	}
	return nil
}

func (r VMResolverAdapterResultV2) Validate() error {
	if err := r.NativeRecord.Validate(); err != nil {
		return err
	}
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if err := r.Output.ValidateFormat(); err != nil {
		return err
	}
	if err := r.Proof.Validate(); err != nil {
		return err
	}
	if r.Status == "" || !IsIdentityResolutionStatus(r.Status) {
		return fmt.Errorf("unknown VM resolver adapter status %q", r.Status)
	}
	if r.ResultHash != ComputeVMResolverAdapterResultHashV2(r) {
		return errors.New("VM resolver adapter result hash mismatch")
	}
	return nil
}

func ValidateVMResolverNativeAuthorityV2(native NativeResolverRecordV2, ctx VMResolverContractContextV2) error {
	if native.NameHash != ctx.NameHash {
		return errors.New("VM resolver authority name_hash mismatch")
	}
	if !bytes.Equal(native.Owner, ctx.DomainOwner) {
		return errors.New("VM resolver authority owner mismatch")
	}
	if native.ResolverRecordVersion != ctx.ResolverRecordVersion {
		return errors.New("VM resolver authority record version mismatch")
	}
	if native.ExpiryHeight != ctx.DomainExpiryHeight {
		return errors.New("VM resolver authority expiry mismatch")
	}
	if native.TargetType != ctx.TargetType {
		return errors.New("VM resolver authority target type mismatch")
	}
	return nil
}

func ValidateVMResolverOutputAgainstNativeV2(native NativeResolverRecordV2, ctx VMResolverContractContextV2, output VMResolverContractOutputV2) error {
	if err := output.ValidateFormat(); err != nil {
		return err
	}
	if output.NameHash != native.NameHash || output.NameHash != ctx.NameHash {
		return errors.New("VM resolver output name_hash mismatch")
	}
	if output.TargetType != native.TargetType || output.TargetType != ctx.TargetType {
		return errors.New("VM resolver output target type mismatch")
	}
	if output.ResolverRecordVersion != native.ResolverRecordVersion {
		return errors.New("VM resolver output record version mismatch")
	}
	if len(output.OwnerOverride) > 0 && !bytes.Equal(output.OwnerOverride, native.Owner) {
		return errors.New("VM resolver contract cannot override native owner")
	}
	if output.GasUsed > ctx.Limits.GasLimit {
		return errors.New("VM resolver output exceeds gas limit")
	}
	if output.MemoryUsedBytes > ctx.Limits.MemoryBytes {
		return errors.New("VM resolver output exceeds memory limit")
	}
	if uint64(len(output.ResolvedValue)) > ctx.Limits.MaxOutputBytes {
		return errors.New("VM resolver output exceeds byte limit")
	}
	if output.ProofChecks > ctx.Limits.MaxProofChecks {
		return errors.New("VM resolver output exceeds proof check limit")
	}
	if output.RecursionDepth > ctx.Limits.MaxRecursionDepth {
		return errors.New("VM resolver output exceeds recursion limit")
	}
	return nil
}

func resolveVMResolverOutputOrFallback(native NativeResolverRecordV2, ctx VMResolverContractContextV2, output VMResolverContractOutputV2, fallback VMResolverFallbackPolicyV2) (VMResolverContractOutputV2, bool, IdentityResolutionStatus, error) {
	if err := ValidateVMResolverOutputAgainstNativeV2(native, ctx, output); err == nil && output.Status == IdentityResolutionStatusResolved {
		return output, false, IdentityResolutionStatusResolved, nil
	}
	if !fallback.AllowNativeFallback {
		if err := ValidateVMResolverOutputAgainstNativeV2(native, ctx, output); err != nil {
			return VMResolverContractOutputV2{}, false, IdentityResolutionStatusFailed, err
		}
		return VMResolverContractOutputV2{}, false, output.Status, errors.New("VM resolver contract failed and native fallback is disabled")
	}
	status := fallback.FallbackStatus
	if status == "" {
		status = IdentityResolutionStatusFailed
	}
	fallbackOutput, err := NewVMResolverContractOutputV2(VMResolverContractOutputV2{
		NameHash:		native.NameHash,
		TargetType:		native.TargetType,
		ResolvedValue:		native.TargetValue,
		ResolverRecordVersion:	native.ResolverRecordVersion,
		GasUsed:		minVMResolverUint64(output.GasUsed, ctx.Limits.GasLimit),
		MemoryUsedBytes:	minVMResolverUint64(output.MemoryUsedBytes, ctx.Limits.MemoryBytes),
		ProofChecks:		minVMResolverUint32(output.ProofChecks, ctx.Limits.MaxProofChecks),
		RecursionDepth:		minVMResolverUint32(output.RecursionDepth, ctx.Limits.MaxRecursionDepth),
		Status:			IdentityResolutionStatusResolved,
	})
	if err != nil {
		return VMResolverContractOutputV2{}, false, status, err
	}
	return fallbackOutput, true, status, nil
}

func ComputeNativeResolverRecordHashV2(record NativeResolverRecordV2) string {
	return identityHash(
		"identity-v2-native-resolver-record",
		record.Name,
		record.NameHash,
		string(record.Owner),
		fmt.Sprintf("%020d", record.ResolverRecordVersion),
		fmt.Sprintf("%020d", record.ExpiryHeight),
		string(record.TargetType),
		string(record.TargetValue),
		record.ProofKey,
	)
}

func ComputeVMResolverContractContextHashV2(ctx VMResolverContractContextV2) string {
	return identityHash(
		"identity-v2-vm-resolver-context",
		ctx.NameHash,
		string(ctx.DomainOwner),
		string(ctx.DomainStatus),
		fmt.Sprintf("%020d", ctx.DomainExpiryHeight),
		fmt.Sprintf("%020d", ctx.ResolverRecordVersion),
		fmt.Sprintf("%020d", ctx.CodeID),
		ctx.ContractAddress,
		string(ctx.TargetType),
		fmt.Sprintf("%020d", ctx.Height),
		fmt.Sprintf("%020d", ctx.Limits.GasLimit),
		fmt.Sprintf("%020d", ctx.Limits.MemoryBytes),
		fmt.Sprintf("%020d", ctx.Limits.MaxOutputBytes),
		fmt.Sprintf("%010d", ctx.Limits.MaxProofChecks),
		fmt.Sprintf("%010d", ctx.Limits.MaxRecursionDepth),
	)
}

func ComputeVMResolverContractOutputHashV2(output VMResolverContractOutputV2) string {
	return identityHash(
		"identity-v2-vm-resolver-output",
		output.NameHash,
		string(output.TargetType),
		string(output.ResolvedValue),
		fmt.Sprintf("%020d", output.ResolverRecordVersion),
		fmt.Sprintf("%020d", output.GasUsed),
		fmt.Sprintf("%020d", output.MemoryUsedBytes),
		fmt.Sprintf("%010d", output.ProofChecks),
		fmt.Sprintf("%010d", output.RecursionDepth),
		string(output.OwnerOverride),
		string(output.Status),
	)
}

func ComputeVMResolverContractProofHashV2(proof VMResolverContractProofV2) string {
	return identityHash(
		"identity-v2-vm-resolver-proof",
		fmt.Sprintf("%020d", proof.Height),
		proof.Name,
		proof.NameHash,
		fmt.Sprintf("%020d", proof.CodeID),
		proof.ContractAddress,
		proof.ResolverRoot,
		proof.NativeRecordHash,
		proof.OutputHash,
	)
}

func ComputeVMResolverAdapterResultHashV2(result VMResolverAdapterResultV2) string {
	return identityHash(
		"identity-v2-vm-resolver-result",
		result.NativeRecord.RecordHash,
		result.Context.ContextHash,
		result.Output.OutputHash,
		result.Proof.ProofHash,
		fmt.Sprintf("%t", result.UsedFallback),
		string(result.Status),
	)
}

func minVMResolverUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}

func minVMResolverUint32(left, right uint32) uint32 {
	if left < right {
		return left
	}
	return right
}
