package types

import (
	"errors"
	"fmt"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	StorageModelEphemeral		= "ephemeral"
	StorageModelPersistentOnChain	= "persistent_on_chain"
	StorageModelDistributedOffChain	= "distributed_off_chain"
	StorageModelHybrid		= "hybrid"

	StorageRetrievalNone			= "none"
	StorageRetrievalInline			= "inline"
	StorageRetrievalOnChainState		= "on_chain_state"
	StorageRetrievalContentAddressed	= "content_addressed"
	StorageRetrievalProviderRPC		= "provider_rpc"
	StorageRetrievalHybridEndpoint		= "hybrid_endpoint"

	StorageVerificationNone			= "none"
	StorageVerificationStateRoot		= "state_root"
	StorageVerificationContentHash		= "content_hash"
	StorageVerificationChunkProof		= "chunk_proof"
	StorageVerificationHybridCommitment	= "hybrid_commitment"

	StorageRetentionNone		= "none"
	StorageRetentionHeight		= "height"
	StorageRetentionExpiry		= "expiry"
	StorageRetentionPermanent	= "permanent"
)

type StorageDeclaration struct {
	StorageModel		string
	ContentHashOptional	string
	StateRootOptional	string
	ContentLocationOptional	string
	RetrievalMethod		string
	VerificationMethod	string
	RetentionPolicy		string
	AccessPolicy		string
	AccessReceiptOptional	string
	MaxPayloadBytes		uint64
	DeclarationHash		string
}

type StorageConsensusValidationPlan struct {
	DeclarationHash		string
	RequiredCommitments	[]string
	RequiresOffchainPayload	bool
	MaxPayloadBytes		uint64
	PlanHash		string
}

func NewStorageDeclaration(declaration StorageDeclaration) (StorageDeclaration, error) {
	if declaration.DeclarationHash != "" {
		return StorageDeclaration{}, errors.New("storage declaration hash must be empty before construction")
	}
	declaration = canonicalStorageDeclaration(declaration)
	if err := declaration.ValidateFormat(); err != nil {
		return StorageDeclaration{}, err
	}
	declaration.DeclarationHash = ComputeStorageDeclarationHash(declaration)
	return declaration, declaration.Validate()
}

func NewHybridStorageDeclaration(object StorageObject, stateRoot, contentLocation, retrievalMethod, accessReceipt string, maxPayloadBytes uint64) (StorageDeclaration, error) {
	if err := object.Validate(); err != nil {
		return StorageDeclaration{}, err
	}
	if retrievalMethod == "" {
		retrievalMethod = StorageRetrievalHybridEndpoint
	}
	return NewStorageDeclaration(StorageDeclaration{
		StorageModel:			StorageModelHybrid,
		ContentHashOptional:		object.ContentHash,
		StateRootOptional:		stateRoot,
		ContentLocationOptional:	contentLocation,
		RetrievalMethod:		retrievalMethod,
		VerificationMethod:		StorageVerificationHybridCommitment,
		RetentionPolicy:		StorageRetentionExpiry,
		AccessPolicy:			object.AccessPolicy,
		AccessReceiptOptional:		accessReceipt,
		MaxPayloadBytes:		maxPayloadBytes,
	})
}

func NewStorageDeclarationFromServiceDescriptor(descriptor coretypes.ServiceStorageDescriptor) (StorageDeclaration, error) {
	declaration := StorageDeclaration{
		StorageModel:		storageModelFromServiceDescriptor(descriptor.Model),
		ContentHashOptional:	descriptor.ContentHash,
		StateRootOptional:	descriptor.StateRoot,
		RetrievalMethod:	descriptor.RetrievalMethod,
		VerificationMethod:	descriptor.VerificationMethod,
		RetentionPolicy:	descriptor.RetentionPolicy,
		AccessPolicy:		descriptor.AccessPolicy,
		MaxPayloadBytes:	descriptor.MaxPayloadBytes,
	}
	if declaration.ContentHashOptional == "" && descriptor.Model != coretypes.ServiceStorageOnChain {
		declaration.ContentHashOptional = descriptor.CommitmentHash
	}
	if declaration.StateRootOptional == "" && (descriptor.Model == coretypes.ServiceStorageOnChain || descriptor.Model == coretypes.ServiceStorageHybridCommitment) {
		declaration.StateRootOptional = descriptor.CommitmentHash
	}
	if declaration.RetrievalMethod == "" {
		declaration.RetrievalMethod = defaultStorageRetrievalForModel(declaration.StorageModel)
	}
	if declaration.VerificationMethod == "" {
		declaration.VerificationMethod = defaultStorageVerificationForModel(declaration.StorageModel, descriptor.ProofRequired)
	}
	if declaration.RetentionPolicy == "" {
		declaration.RetentionPolicy = defaultStorageRetentionForModel(declaration.StorageModel)
	}
	if declaration.AccessPolicy == "" {
		declaration.AccessPolicy = AccessPolicyPermissioned
	}
	if declaration.MaxPayloadBytes == 0 {
		declaration.MaxPayloadBytes = DefaultMaxStateBytes
	}
	if descriptor.Model == coretypes.ServiceStorageHybridCommitment && declaration.ContentLocationOptional == "" {
		declaration.ContentLocationOptional = "descriptor/" + descriptor.CommitmentHash
	}
	return NewStorageDeclaration(declaration)
}

func (declaration StorageDeclaration) ValidateFormat() error {
	declaration = canonicalStorageDeclaration(declaration)
	if !IsStorageDeclarationModel(declaration.StorageModel) {
		return fmt.Errorf("unknown storage declaration model %q", declaration.StorageModel)
	}
	if !IsStorageRetrievalMethod(declaration.RetrievalMethod) {
		return fmt.Errorf("unknown storage retrieval method %q", declaration.RetrievalMethod)
	}
	if !IsStorageVerificationMethod(declaration.VerificationMethod) {
		return fmt.Errorf("unknown storage verification method %q", declaration.VerificationMethod)
	}
	if !IsStorageRetentionPolicy(declaration.RetentionPolicy) {
		return fmt.Errorf("unknown storage retention policy %q", declaration.RetentionPolicy)
	}
	if !IsStorageAccessPolicy(declaration.AccessPolicy) {
		return fmt.Errorf("unknown storage access policy %q", declaration.AccessPolicy)
	}
	if declaration.ContentHashOptional != "" {
		if err := validateStorageHash("storage declaration content hash", declaration.ContentHashOptional); err != nil {
			return err
		}
	}
	if declaration.StateRootOptional != "" {
		if err := validateStorageHash("storage declaration state root", declaration.StateRootOptional); err != nil {
			return err
		}
	}
	if declaration.ContentLocationOptional != "" {
		if err := validateStorageToken("storage declaration content location", declaration.ContentLocationOptional); err != nil {
			return err
		}
	}
	if declaration.AccessReceiptOptional != "" {
		if err := validateStorageHash("storage declaration access receipt", declaration.AccessReceiptOptional); err != nil {
			return err
		}
	}
	if declaration.MaxPayloadBytes == 0 || declaration.MaxPayloadBytes > MaxStorageObjectSize {
		return fmt.Errorf("storage declaration max payload bytes must be between 1 and %d", MaxStorageObjectSize)
	}
	if declaration.DeclarationHash != "" {
		if err := validateStorageHash("storage declaration hash", declaration.DeclarationHash); err != nil {
			return err
		}
	}
	return validateStorageDeclarationModelRules(declaration)
}

func (declaration StorageDeclaration) Validate() error {
	declaration = canonicalStorageDeclaration(declaration)
	if err := declaration.ValidateFormat(); err != nil {
		return err
	}
	if declaration.DeclarationHash == "" {
		return errors.New("storage declaration hash is required")
	}
	if declaration.DeclarationHash != ComputeStorageDeclarationHash(declaration) {
		return errors.New("storage declaration hash mismatch")
	}
	return nil
}

func BuildStorageConsensusValidationPlan(declaration StorageDeclaration) (StorageConsensusValidationPlan, error) {
	if err := declaration.Validate(); err != nil {
		return StorageConsensusValidationPlan{}, err
	}
	commitments := make([]string, 0, 2)
	if declaration.ContentHashOptional != "" {
		commitments = append(commitments, declaration.ContentHashOptional)
	}
	if declaration.StateRootOptional != "" {
		commitments = append(commitments, declaration.StateRootOptional)
	}
	plan := StorageConsensusValidationPlan{
		DeclarationHash:		declaration.DeclarationHash,
		RequiredCommitments:		normalizeStorageHashes(commitments),
		RequiresOffchainPayload:	false,
		MaxPayloadBytes:		declaration.MaxPayloadBytes,
	}
	plan.PlanHash = ComputeStorageConsensusValidationPlanHash(plan)
	return plan, plan.Validate()
}

func (plan StorageConsensusValidationPlan) Validate() error {
	if err := validateStorageHash("storage consensus plan declaration hash", plan.DeclarationHash); err != nil {
		return err
	}
	if len(plan.RequiredCommitments) == 0 {
		return errors.New("storage consensus plan requires at least one commitment")
	}
	for i, commitment := range plan.RequiredCommitments {
		if err := validateStorageHash("storage consensus plan commitment", commitment); err != nil {
			return err
		}
		if i > 0 && plan.RequiredCommitments[i-1] >= commitment {
			return errors.New("storage consensus plan commitments must be sorted canonically")
		}
	}
	if plan.RequiresOffchainPayload {
		return errors.New("storage consensus plan must not require off-chain payload retrieval")
	}
	if plan.MaxPayloadBytes == 0 || plan.MaxPayloadBytes > MaxStorageObjectSize {
		return fmt.Errorf("storage consensus plan max payload bytes must be between 1 and %d", MaxStorageObjectSize)
	}
	if err := validateStorageHash("storage consensus plan hash", plan.PlanHash); err != nil {
		return err
	}
	if plan.PlanHash != ComputeStorageConsensusValidationPlanHash(plan) {
		return errors.New("storage consensus plan hash mismatch")
	}
	return nil
}

func ComputeStorageDeclarationHash(declaration StorageDeclaration) string {
	declaration = canonicalStorageDeclaration(declaration)
	return storageHashParts(
		"storage-declaration-v1",
		declaration.StorageModel,
		declaration.ContentHashOptional,
		declaration.StateRootOptional,
		declaration.ContentLocationOptional,
		declaration.RetrievalMethod,
		declaration.VerificationMethod,
		declaration.RetentionPolicy,
		declaration.AccessPolicy,
		declaration.AccessReceiptOptional,
		fmt.Sprintf("%020d", declaration.MaxPayloadBytes),
	)
}

func ComputeStorageConsensusValidationPlanHash(plan StorageConsensusValidationPlan) string {
	commitments := normalizeStorageHashes(plan.RequiredCommitments)
	parts := []string{
		"storage-consensus-validation-plan-v1",
		plan.DeclarationHash,
		fmt.Sprintf("%t", plan.RequiresOffchainPayload),
		fmt.Sprintf("%020d", plan.MaxPayloadBytes),
		fmt.Sprintf("%020d", len(commitments)),
	}
	parts = append(parts, commitments...)
	return storageHashParts(parts...)
}

func IsStorageDeclarationModel(model string) bool {
	switch model {
	case StorageModelEphemeral, StorageModelPersistentOnChain, StorageModelDistributedOffChain, StorageModelHybrid:
		return true
	default:
		return false
	}
}

func IsStorageRetrievalMethod(method string) bool {
	switch method {
	case StorageRetrievalNone, StorageRetrievalInline, StorageRetrievalOnChainState,
		StorageRetrievalContentAddressed, StorageRetrievalProviderRPC, StorageRetrievalHybridEndpoint:
		return true
	default:
		return false
	}
}

func IsStorageVerificationMethod(method string) bool {
	switch method {
	case StorageVerificationNone, StorageVerificationStateRoot, StorageVerificationContentHash,
		StorageVerificationChunkProof, StorageVerificationHybridCommitment:
		return true
	default:
		return false
	}
}

func IsStorageRetentionPolicy(policy string) bool {
	switch policy {
	case StorageRetentionNone, StorageRetentionHeight, StorageRetentionExpiry, StorageRetentionPermanent:
		return true
	default:
		return false
	}
}

func storageModelFromServiceDescriptor(model coretypes.ServiceStorageModel) string {
	switch model {
	case coretypes.ServiceStorageEphemeral, coretypes.ServiceStorageNone:
		return StorageModelEphemeral
	case coretypes.ServiceStorageOnChain:
		return StorageModelPersistentOnChain
	case coretypes.ServiceStorageDistributedOffChain:
		return StorageModelDistributedOffChain
	case coretypes.ServiceStorageHybridCommitment:
		return StorageModelHybrid
	default:
		return string(model)
	}
}

func defaultStorageRetrievalForModel(model string) string {
	switch model {
	case StorageModelPersistentOnChain:
		return StorageRetrievalOnChainState
	case StorageModelDistributedOffChain:
		return StorageRetrievalContentAddressed
	case StorageModelHybrid:
		return StorageRetrievalHybridEndpoint
	default:
		return StorageRetrievalInline
	}
}

func defaultStorageVerificationForModel(model string, proofRequired bool) string {
	switch model {
	case StorageModelPersistentOnChain:
		return StorageVerificationStateRoot
	case StorageModelDistributedOffChain:
		if proofRequired {
			return StorageVerificationChunkProof
		}
		return StorageVerificationContentHash
	case StorageModelHybrid:
		return StorageVerificationHybridCommitment
	default:
		return StorageVerificationNone
	}
}

func defaultStorageRetentionForModel(model string) string {
	switch model {
	case StorageModelEphemeral:
		return StorageRetentionNone
	case StorageModelPersistentOnChain:
		return StorageRetentionPermanent
	default:
		return StorageRetentionExpiry
	}
}

func canonicalStorageDeclaration(declaration StorageDeclaration) StorageDeclaration {
	declaration.StorageModel = strings.TrimSpace(declaration.StorageModel)
	declaration.ContentHashOptional = strings.ToLower(strings.TrimSpace(declaration.ContentHashOptional))
	declaration.StateRootOptional = strings.ToLower(strings.TrimSpace(declaration.StateRootOptional))
	declaration.ContentLocationOptional = strings.TrimSpace(declaration.ContentLocationOptional)
	declaration.RetrievalMethod = strings.TrimSpace(declaration.RetrievalMethod)
	declaration.VerificationMethod = strings.TrimSpace(declaration.VerificationMethod)
	declaration.RetentionPolicy = strings.TrimSpace(declaration.RetentionPolicy)
	declaration.AccessPolicy = strings.TrimSpace(declaration.AccessPolicy)
	declaration.AccessReceiptOptional = strings.ToLower(strings.TrimSpace(declaration.AccessReceiptOptional))
	declaration.DeclarationHash = strings.ToLower(strings.TrimSpace(declaration.DeclarationHash))
	return declaration
}

func validateStorageDeclarationModelRules(declaration StorageDeclaration) error {
	switch declaration.StorageModel {
	case StorageModelEphemeral:
		if declaration.RetrievalMethod != StorageRetrievalInline && declaration.RetrievalMethod != StorageRetrievalNone {
			return errors.New("ephemeral storage must use inline or none retrieval")
		}
		if declaration.VerificationMethod != StorageVerificationNone && declaration.VerificationMethod != StorageVerificationContentHash {
			return errors.New("ephemeral storage must use none or content-hash verification")
		}
	case StorageModelPersistentOnChain:
		if declaration.StateRootOptional == "" {
			return errors.New("persistent on-chain storage requires state root")
		}
		if declaration.RetrievalMethod != StorageRetrievalOnChainState {
			return errors.New("persistent on-chain storage requires on-chain retrieval")
		}
		if declaration.VerificationMethod != StorageVerificationStateRoot {
			return errors.New("persistent on-chain storage requires state-root verification")
		}
	case StorageModelDistributedOffChain:
		if declaration.ContentHashOptional == "" {
			return errors.New("distributed off-chain storage requires content hash")
		}
		if declaration.RetrievalMethod != StorageRetrievalContentAddressed && declaration.RetrievalMethod != StorageRetrievalProviderRPC {
			return errors.New("distributed off-chain storage requires content-addressed or provider-rpc retrieval")
		}
		if declaration.VerificationMethod != StorageVerificationContentHash && declaration.VerificationMethod != StorageVerificationChunkProof {
			return errors.New("distributed off-chain storage requires content-hash or chunk-proof verification")
		}
	case StorageModelHybrid:
		if declaration.ContentHashOptional == "" || declaration.StateRootOptional == "" {
			return errors.New("hybrid storage requires on-chain commitment and off-chain content hash")
		}
		if declaration.ContentLocationOptional == "" {
			return errors.New("hybrid storage requires off-chain content location")
		}
		if declaration.RetrievalMethod != StorageRetrievalHybridEndpoint && declaration.RetrievalMethod != StorageRetrievalContentAddressed {
			return errors.New("hybrid storage requires hybrid endpoint or content-addressed retrieval")
		}
		if declaration.VerificationMethod != StorageVerificationHybridCommitment {
			return errors.New("hybrid storage requires hybrid commitment verification")
		}
	}
	return nil
}
