package types

import (
	"errors"
	"fmt"
	"math"
)

const (
	DefaultStorageFeeDenom		= "uaet"
	DefaultStorageBytePrice		= uint64(1)
	DefaultStorageChunkBytePrice	= uint64(1)
	DefaultStorageReceiptFee	= uint64(10)
	DefaultStorageMinimumFee	= uint64(100)
	DefaultLazyFetchMaxBytes	= uint64(4 * 1024 * 1024)
	LazyFetchBoundaryVersionV1	= uint64(1)
)

type StorageReplicationPolicyRule struct {
	Policy			string
	MinReplicas		uint32
	MaxReplicas		uint32
	MinAvailabilityBps	uint32
	RequiresErasureGroup	bool
}

type StorageValidationParams struct {
	ChunkParams			StorageChunkParams
	AllowedAccessPolicies		[]string
	ReplicationPolicyRules		[]StorageReplicationPolicyRule
	RequireProofBackedRetrieval	bool
}

type StorageFeeParams struct {
	Denom		string
	ObjectBytePrice	uint64
	ChunkBytePrice	uint64
	ReceiptFee	uint64
	MinimumFee	uint64
}

type StorageFeeQuote struct {
	Payer		string
	ObjectID	string
	Denom		string
	ObjectBytes	uint64
	ChunkBytes	uint64
	ReceiptCount	uint32
	FeeAmount	uint64
	QuoteHash	string
}

type LazyFetchRequest struct {
	ObjectID		string
	ContentHash		string
	ChunkIndexOptional	uint32
	ChunkProofRoot		string
	Requester		string
	MaxBytes		uint64
	RequestHeight		uint64
	Version			uint64
	RequestHash		string
}

type LazyFetchResultBoundary struct {
	RequestHash	string
	Provider	string
	PayloadHash	string
	Proof		StorageChunkInclusionProof
	ResultHeight	uint64
	ResultHash	string
}

func DefaultStorageValidationParams() StorageValidationParams {
	return StorageValidationParams{
		ChunkParams:	DefaultStorageChunkParams(),
		AllowedAccessPolicies: []string{
			AccessPolicyPermissioned,
			AccessPolicyPrivate,
			AccessPolicyPublicRead,
		},
		ReplicationPolicyRules: []StorageReplicationPolicyRule{
			{Policy: ReplicationPolicyErasure, MinReplicas: 4, MaxReplicas: 256, MinAvailabilityBps: 9900, RequiresErasureGroup: true},
			{Policy: ReplicationPolicyMultiZone, MinReplicas: 3, MaxReplicas: 64, MinAvailabilityBps: 9900},
			{Policy: ReplicationPolicyRegional, MinReplicas: 2, MaxReplicas: 16, MinAvailabilityBps: 9500},
			{Policy: ReplicationPolicySingle, MinReplicas: 1, MaxReplicas: 1, MinAvailabilityBps: 9000},
		},
		RequireProofBackedRetrieval:	true,
	}
}

func DefaultStorageFeeParams() StorageFeeParams {
	return StorageFeeParams{
		Denom:			DefaultStorageFeeDenom,
		ObjectBytePrice:	DefaultStorageBytePrice,
		ChunkBytePrice:		DefaultStorageChunkBytePrice,
		ReceiptFee:		DefaultStorageReceiptFee,
		MinimumFee:		DefaultStorageMinimumFee,
	}
}

func ValidateStorageObjectAgainstChunks(object StorageObject, descriptors []StorageChunkDescriptor, params StorageValidationParams) error {
	params = normalizeStorageValidationParams(params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := object.Validate(); err != nil {
		return err
	}
	if err := validateStorageChunkDescriptors(descriptors, params.ChunkParams); err != nil {
		return err
	}
	if !isAllowedStorageAccessPolicy(object.AccessPolicy, params.AllowedAccessPolicies) {
		return fmt.Errorf("storage access policy %q is not parameter-allowed", object.AccessPolicy)
	}
	if _, found := findStorageReplicationPolicyRule(object.ReplicationPolicy, params.ReplicationPolicyRules); !found {
		return fmt.Errorf("storage replication policy %q is not parameter-allowed", object.ReplicationPolicy)
	}
	contentHash := ComputeStorageContentHashFromChunks(descriptors)
	if object.ContentHash != contentHash {
		return errors.New("storage object content_hash must match chunk root commitment")
	}
	descriptorRoots := normalizeStorageHashes(storageChunkDescriptorRoots(descriptors))
	if !equalStringSlices(object.ChunkRoots, descriptorRoots) {
		return errors.New("storage object chunk roots mismatch chunk descriptors")
	}
	size, err := SumStorageChunkSizes(descriptors)
	if err != nil {
		return err
	}
	if object.Size != size {
		return errors.New("storage object size must equal sum of chunk sizes")
	}
	return nil
}

func ValidateStorageReplicationPolicyAgainstParams(commitment ReplicationStatusCommitment, descriptors []StorageChunkDescriptor, params StorageValidationParams) error {
	params = normalizeStorageValidationParams(params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := commitment.Validate(); err != nil {
		return err
	}
	rule, found := findStorageReplicationPolicyRule(commitment.ReplicationPolicy, params.ReplicationPolicyRules)
	if !found {
		return fmt.Errorf("storage replication policy %q is not parameter-allowed", commitment.ReplicationPolicy)
	}
	if commitment.ReplicaCount < rule.MinReplicas || commitment.ReplicaCount > rule.MaxReplicas {
		return fmt.Errorf("storage replication replica count for %s must be between %d and %d", commitment.ReplicationPolicy, rule.MinReplicas, rule.MaxReplicas)
	}
	if commitment.AvailabilityBps < rule.MinAvailabilityBps {
		return fmt.Errorf("storage replication availability bps for %s must be >= %d", commitment.ReplicationPolicy, rule.MinAvailabilityBps)
	}
	if rule.RequiresErasureGroup {
		for _, descriptor := range descriptors {
			if descriptor.ErasureGroupOptional == "" {
				return errors.New("storage erasure-coded replication requires erasure group metadata")
			}
		}
	}
	return nil
}

func ValidateStorageReceiptReferencesRegisteredObject(receipt StorageAccessReceipt, state StorageStateV2) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	object, found := QueryStorageObject(state, receipt.ObjectID)
	if !found {
		return fmt.Errorf("storage receipt references unregistered object %s", receipt.ObjectID)
	}
	if receipt.ContentHash != object.ContentHash {
		return errors.New("storage receipt content hash mismatch registered object")
	}
	if receipt.PolicyHash != ComputeStoragePolicyHash(object.ReplicationPolicy, object.AccessPolicy, object.StorageClass) {
		return errors.New("storage receipt policy hash mismatch registered object")
	}
	return nil
}

func ValidateStorageRetrievalWithProof(state StorageStateV2, objectID string, proof StorageChunkInclusionProof, params StorageValidationParams) error {
	params = normalizeStorageValidationParams(params)
	object, found := QueryStorageObject(state, objectID)
	if !found {
		return fmt.Errorf("storage retrieval proof references unregistered object %s", objectID)
	}
	if err := VerifyStorageChunkInclusionInObject(proof, object); err != nil {
		return err
	}
	descriptor, found := QueryChunkDescriptor(state, objectID, proof.ChunkIndex)
	if !found {
		return fmt.Errorf("storage retrieval proof references missing chunk %d", proof.ChunkIndex)
	}
	if err := descriptor.Validate(params.ChunkParams); err != nil {
		return err
	}
	if descriptor.ChunkHash != proof.ChunkHash {
		return errors.New("storage retrieval proof chunk hash mismatch descriptor")
	}
	if descriptor.ChunkProofRoot != proof.ChunkProofRoot {
		return errors.New("storage retrieval proof root mismatch descriptor")
	}
	return nil
}

func ValidateStorageStateV2Rules(state StorageStateV2, params StorageValidationParams) error {
	params = normalizeStorageValidationParams(params)
	if err := state.Validate(); err != nil {
		return err
	}
	for _, object := range state.Objects {
		descriptors := storageDescriptorsForObject(state.Chunks, object.ObjectID)
		if err := ValidateStorageObjectAgainstChunks(object, descriptors, params); err != nil {
			return err
		}
		replication, found := storageReplicationForObject(state.Replications, object.ObjectID)
		if !found {
			return fmt.Errorf("storage object %s requires replication commitment", object.ObjectID)
		}
		if err := ValidateStorageReplicationPolicyAgainstParams(replication, descriptors, params); err != nil {
			return err
		}
	}
	for _, receipt := range state.AccessReceipts {
		if err := ValidateStorageReceiptReferencesRegisteredObject(receipt, state); err != nil {
			return err
		}
		if params.RequireProofBackedRetrieval && receipt.RetrievalProof == "" {
			return errors.New("storage retrieval validation requires proof-backed chunks")
		}
	}
	return nil
}

func SumStorageChunkSizes(descriptors []StorageChunkDescriptor) (uint64, error) {
	var total uint64
	for _, descriptor := range descriptors {
		if descriptor.ChunkSize > math.MaxUint64-total {
			return 0, errors.New("storage chunk size sum overflow")
		}
		total += descriptor.ChunkSize
	}
	return total, nil
}

func NewStorageFeeQuote(payer string, object StorageObject, descriptors []StorageChunkDescriptor, receiptCount uint32, params StorageFeeParams) (StorageFeeQuote, error) {
	params = normalizeStorageFeeParams(params)
	if err := params.Validate(); err != nil {
		return StorageFeeQuote{}, err
	}
	if err := validateStorageToken("storage fee payer", payer); err != nil {
		return StorageFeeQuote{}, err
	}
	if err := object.Validate(); err != nil {
		return StorageFeeQuote{}, err
	}
	chunkBytes, err := SumStorageChunkSizes(descriptors)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	objectFee, err := checkedMul(object.Size, params.ObjectBytePrice)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	chunkFee, err := checkedMul(chunkBytes, params.ChunkBytePrice)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	receiptFee, err := checkedMul(uint64(receiptCount), params.ReceiptFee)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	total, err := checkedAdd(objectFee, chunkFee)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	total, err = checkedAdd(total, receiptFee)
	if err != nil {
		return StorageFeeQuote{}, err
	}
	if total < params.MinimumFee {
		total = params.MinimumFee
	}
	quote := StorageFeeQuote{
		Payer:		payer,
		ObjectID:	object.ObjectID,
		Denom:		params.Denom,
		ObjectBytes:	object.Size,
		ChunkBytes:	chunkBytes,
		ReceiptCount:	receiptCount,
		FeeAmount:	total,
	}
	quote.QuoteHash = ComputeStorageFeeQuoteHash(quote)
	return quote, quote.Validate()
}

func NewLazyFetchRequest(request LazyFetchRequest) (LazyFetchRequest, error) {
	if request.RequestHash != "" {
		return LazyFetchRequest{}, errors.New("lazy fetch request hash must be empty before construction")
	}
	if request.Version == 0 {
		request.Version = LazyFetchBoundaryVersionV1
	}
	if err := request.ValidateFormat(); err != nil {
		return LazyFetchRequest{}, err
	}
	request.RequestHash = ComputeLazyFetchRequestHash(request)
	return request, request.Validate()
}

func NewLazyFetchResultBoundary(result LazyFetchResultBoundary) (LazyFetchResultBoundary, error) {
	if result.ResultHash != "" {
		return LazyFetchResultBoundary{}, errors.New("lazy fetch result hash must be empty before construction")
	}
	if err := result.ValidateFormat(); err != nil {
		return LazyFetchResultBoundary{}, err
	}
	result.ResultHash = ComputeLazyFetchResultBoundaryHash(result)
	return result, result.Validate()
}

func ValidateLazyFetchBoundaryOutsideConsensus(request LazyFetchRequest, result LazyFetchResultBoundary, object StorageObject) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if err := result.Validate(); err != nil {
		return err
	}
	if err := object.Validate(); err != nil {
		return err
	}
	if result.RequestHash != request.RequestHash {
		return errors.New("lazy fetch result request hash mismatch")
	}
	if request.ObjectID != object.ObjectID || request.ContentHash != object.ContentHash {
		return errors.New("lazy fetch request object mismatch")
	}
	if result.Proof.ObjectID != object.ObjectID || result.Proof.ContentHash != object.ContentHash {
		return errors.New("lazy fetch proof object mismatch")
	}
	if request.ChunkProofRoot != "" && request.ChunkProofRoot != result.Proof.ChunkProofRoot {
		return errors.New("lazy fetch proof root mismatch request")
	}
	return VerifyStorageChunkInclusionInObject(result.Proof, object)
}

func (params StorageValidationParams) Validate() error {
	if err := params.ChunkParams.Validate(); err != nil {
		return err
	}
	if len(params.AllowedAccessPolicies) == 0 {
		return errors.New("storage validation requires allowed access policies")
	}
	seenAccess := map[string]struct{}{}
	previousAccess := ""
	for _, policy := range params.AllowedAccessPolicies {
		if !IsStorageAccessPolicy(policy) {
			return fmt.Errorf("unknown storage access policy %q", policy)
		}
		if _, found := seenAccess[policy]; found {
			return fmt.Errorf("duplicate storage access policy %s", policy)
		}
		seenAccess[policy] = struct{}{}
		if previousAccess != "" && previousAccess >= policy {
			return errors.New("storage access policies must be sorted canonically")
		}
		previousAccess = policy
	}
	if len(params.ReplicationPolicyRules) == 0 {
		return errors.New("storage validation requires replication policy rules")
	}
	seenReplication := map[string]struct{}{}
	previousReplication := ""
	for _, rule := range params.ReplicationPolicyRules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seenReplication[rule.Policy]; found {
			return fmt.Errorf("duplicate storage replication policy %s", rule.Policy)
		}
		seenReplication[rule.Policy] = struct{}{}
		if previousReplication != "" && previousReplication >= rule.Policy {
			return errors.New("storage replication policy rules must be sorted canonically")
		}
		previousReplication = rule.Policy
	}
	return nil
}

func (rule StorageReplicationPolicyRule) Validate() error {
	if !IsStorageReplicationPolicy(rule.Policy) {
		return fmt.Errorf("unknown storage replication policy %q", rule.Policy)
	}
	if rule.MinReplicas == 0 || rule.MaxReplicas < rule.MinReplicas {
		return errors.New("storage replication rule replica bounds are invalid")
	}
	if rule.MinAvailabilityBps > 10000 {
		return errors.New("storage replication rule availability bps must be <= 10000")
	}
	return nil
}

func (params StorageFeeParams) Validate() error {
	if err := validateStorageToken("storage fee denom", params.Denom); err != nil {
		return err
	}
	if params.ObjectBytePrice == 0 {
		return errors.New("storage fee object byte price must be positive")
	}
	if params.ChunkBytePrice == 0 {
		return errors.New("storage fee chunk byte price must be positive")
	}
	return nil
}

func (quote StorageFeeQuote) Validate() error {
	if err := validateStorageToken("storage fee payer", quote.Payer); err != nil {
		return err
	}
	if err := validateStorageToken("storage fee object id", quote.ObjectID); err != nil {
		return err
	}
	if err := validateStorageToken("storage fee denom", quote.Denom); err != nil {
		return err
	}
	if quote.ObjectBytes == 0 || quote.ChunkBytes == 0 {
		return errors.New("storage fee quote bytes must be positive")
	}
	if quote.FeeAmount == 0 {
		return errors.New("storage fee quote amount must be positive")
	}
	if err := validateStorageHash("storage fee quote hash", quote.QuoteHash); err != nil {
		return err
	}
	if quote.QuoteHash != ComputeStorageFeeQuoteHash(quote) {
		return errors.New("storage fee quote hash mismatch")
	}
	return nil
}

func (request LazyFetchRequest) ValidateFormat() error {
	if err := validateStorageToken("lazy fetch object id", request.ObjectID); err != nil {
		return err
	}
	if err := validateStorageHash("lazy fetch content hash", request.ContentHash); err != nil {
		return err
	}
	if request.ChunkProofRoot != "" {
		if err := validateStorageHash("lazy fetch chunk proof root", request.ChunkProofRoot); err != nil {
			return err
		}
	}
	if err := validateStorageToken("lazy fetch requester", request.Requester); err != nil {
		return err
	}
	if request.MaxBytes == 0 || request.MaxBytes > DefaultLazyFetchMaxBytes {
		return fmt.Errorf("lazy fetch max bytes must be between 1 and %d", DefaultLazyFetchMaxBytes)
	}
	if request.RequestHeight == 0 {
		return errors.New("lazy fetch request height must be positive")
	}
	if request.Version == 0 {
		return errors.New("lazy fetch version must be positive")
	}
	if request.RequestHash != "" {
		return validateStorageHash("lazy fetch request hash", request.RequestHash)
	}
	return nil
}

func (request LazyFetchRequest) Validate() error {
	if err := request.ValidateFormat(); err != nil {
		return err
	}
	if request.RequestHash == "" {
		return errors.New("lazy fetch request hash is required")
	}
	if request.RequestHash != ComputeLazyFetchRequestHash(request) {
		return errors.New("lazy fetch request hash mismatch")
	}
	return nil
}

func (result LazyFetchResultBoundary) ValidateFormat() error {
	if err := validateStorageHash("lazy fetch result request hash", result.RequestHash); err != nil {
		return err
	}
	if err := validateStorageToken("lazy fetch provider", result.Provider); err != nil {
		return err
	}
	if err := validateStorageHash("lazy fetch payload hash", result.PayloadHash); err != nil {
		return err
	}
	if err := result.Proof.Validate(); err != nil {
		return err
	}
	if result.ResultHeight == 0 {
		return errors.New("lazy fetch result height must be positive")
	}
	if result.ResultHash != "" {
		return validateStorageHash("lazy fetch result hash", result.ResultHash)
	}
	return nil
}

func (result LazyFetchResultBoundary) Validate() error {
	if err := result.ValidateFormat(); err != nil {
		return err
	}
	if result.ResultHash == "" {
		return errors.New("lazy fetch result hash is required")
	}
	if result.ResultHash != ComputeLazyFetchResultBoundaryHash(result) {
		return errors.New("lazy fetch result hash mismatch")
	}
	return nil
}

func ComputeStorageFeeQuoteHash(quote StorageFeeQuote) string {
	return storageHashParts(
		"storage-fee-quote-v1",
		quote.Payer,
		quote.ObjectID,
		quote.Denom,
		fmt.Sprintf("%020d", quote.ObjectBytes),
		fmt.Sprintf("%020d", quote.ChunkBytes),
		fmt.Sprintf("%020d", quote.ReceiptCount),
		fmt.Sprintf("%020d", quote.FeeAmount),
	)
}

func ComputeLazyFetchRequestHash(request LazyFetchRequest) string {
	return storageHashParts(
		"lazy-fetch-request-v1",
		request.ObjectID,
		request.ContentHash,
		fmt.Sprintf("%020d", request.ChunkIndexOptional),
		request.ChunkProofRoot,
		request.Requester,
		fmt.Sprintf("%020d", request.MaxBytes),
		fmt.Sprintf("%020d", request.RequestHeight),
		fmt.Sprintf("%020d", request.Version),
	)
}

func ComputeLazyFetchResultBoundaryHash(result LazyFetchResultBoundary) string {
	return storageHashParts(
		"lazy-fetch-result-boundary-v1",
		result.RequestHash,
		result.Provider,
		result.PayloadHash,
		result.Proof.ProofHash,
		fmt.Sprintf("%020d", result.ResultHeight),
	)
}

func normalizeStorageValidationParams(params StorageValidationParams) StorageValidationParams {
	defaults := DefaultStorageValidationParams()
	if params.ChunkParams.MaxChunkBytes == 0 {
		params.ChunkParams = defaults.ChunkParams
	}
	if len(params.AllowedAccessPolicies) == 0 {
		params.AllowedAccessPolicies = defaults.AllowedAccessPolicies
	}
	if len(params.ReplicationPolicyRules) == 0 {
		params.ReplicationPolicyRules = defaults.ReplicationPolicyRules
	}
	return params
}

func normalizeStorageFeeParams(params StorageFeeParams) StorageFeeParams {
	if params.Denom == "" && params.ObjectBytePrice == 0 && params.ChunkBytePrice == 0 && params.ReceiptFee == 0 && params.MinimumFee == 0 {
		return DefaultStorageFeeParams()
	}
	return params
}

func isAllowedStorageAccessPolicy(policy string, allowed []string) bool {
	for _, candidate := range allowed {
		if candidate == policy {
			return true
		}
	}
	return false
}

func findStorageReplicationPolicyRule(policy string, rules []StorageReplicationPolicyRule) (StorageReplicationPolicyRule, bool) {
	for _, rule := range rules {
		if rule.Policy == policy {
			return rule, true
		}
	}
	return StorageReplicationPolicyRule{}, false
}

func storageDescriptorsForObject(records []StorageChunkDescriptorRecord, objectID string) []StorageChunkDescriptor {
	out := make([]StorageChunkDescriptor, 0)
	for _, record := range records {
		if record.ObjectID == objectID {
			out = append(out, record.Descriptor)
		}
	}
	return out
}

func storageReplicationForObject(replications []ReplicationStatusCommitment, objectID string) (ReplicationStatusCommitment, bool) {
	for _, replication := range replications {
		if replication.ObjectID == objectID {
			return replication, true
		}
	}
	return ReplicationStatusCommitment{}, false
}

func checkedMul(left, right uint64) (uint64, error) {
	if left != 0 && right > math.MaxUint64/left {
		return 0, errors.New("storage fee multiplication overflow")
	}
	return left * right, nil
}

func checkedAdd(left, right uint64) (uint64, error) {
	if right > math.MaxUint64-left {
		return 0, errors.New("storage fee addition overflow")
	}
	return left + right, nil
}
