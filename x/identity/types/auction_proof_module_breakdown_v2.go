package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AuctionModuleStateObjectV2 string
type AuctionModuleMessageNameV2 string
type AuctionModuleQueryNameV2 string
type AuctionModuleFailureModeV2 string
type AuctionModuleIntegrationPointV2 string

const (
	AuctionModulePathV2	= "auction-module"

	AuctionModuleStateAuctionRecord		AuctionModuleStateObjectV2	= "AuctionRecord"
	AuctionModuleStateBidCommitment		AuctionModuleStateObjectV2	= "BidCommitment"
	AuctionModuleStateRevealedBid		AuctionModuleStateObjectV2	= "RevealedBid"
	AuctionModuleStateAuctionParams		AuctionModuleStateObjectV2	= "AuctionParams"
	AuctionModuleStateAuctionFeeSplit	AuctionModuleStateObjectV2	= "AuctionFeeSplit"

	AuctionModuleMsgStartAuction		AuctionModuleMessageNameV2	= "MsgStartAuction"
	AuctionModuleMsgCommitBid		AuctionModuleMessageNameV2	= "MsgCommitBid"
	AuctionModuleMsgRevealBid		AuctionModuleMessageNameV2	= "MsgRevealBid"
	AuctionModuleMsgFinalizeAuction		AuctionModuleMessageNameV2	= "MsgFinalizeAuction"
	AuctionModuleMsgCancelExpiredAuction	AuctionModuleMessageNameV2	= "MsgCancelExpiredAuction"
	AuctionModuleMsgClaimAuctionRefund	AuctionModuleMessageNameV2	= "MsgClaimAuctionRefund"

	AuctionModuleQueryAuction		AuctionModuleQueryNameV2	= "QueryAuction"
	AuctionModuleQueryAuctionByName		AuctionModuleQueryNameV2	= "QueryAuctionByName"
	AuctionModuleQueryBidCommitment		AuctionModuleQueryNameV2	= "QueryBidCommitment"
	AuctionModuleQueryAuctionParams		AuctionModuleQueryNameV2	= "QueryAuctionParams"
	AuctionModuleQueryAuctionPriceFloor	AuctionModuleQueryNameV2	= "QueryAuctionPriceFloor"

	AuctionModuleFailureBidReplay			AuctionModuleFailureModeV2	= "bid_replay"
	AuctionModuleFailureRevealOutsideWindow		AuctionModuleFailureModeV2	= "reveal_outside_window"
	AuctionModuleFailureNonDeterministicTie		AuctionModuleFailureModeV2	= "non_deterministic_tie_handling"
	AuctionModuleFailureFinalizeBeforeRevealEnd	AuctionModuleFailureModeV2	= "finalization_before_reveal_end"
	AuctionModuleFailureIncorrectRefundAccounting	AuctionModuleFailureModeV2	= "incorrect_refund_accounting"

	AuctionModuleIntegrationIdentityCore				AuctionModuleIntegrationPointV2	= "identity_core"
	AuctionModuleIntegrationBankModule				AuctionModuleIntegrationPointV2	= "bank_module"
	AuctionModuleIntegrationFeeModule				AuctionModuleIntegrationPointV2	= "fee_module"
	AuctionModuleIntegrationTreasuryBurnRewardsCommunityPool	AuctionModuleIntegrationPointV2	= "treasury_burn_rewards_community_pool"
	AuctionModuleIntegrationStoreV2					AuctionModuleIntegrationPointV2	= "store_v2"
)

type ProofModuleStateObjectV2 string
type ProofModuleMessageNameV2 string
type ProofModuleQueryNameV2 string
type ProofModuleFailureModeV2 string
type ProofModuleIntegrationPointV2 string

const (
	ProofVerificationModulePathV2	= "proof-verification-module"

	ProofModuleStateProofParams		ProofModuleStateObjectV2	= "ProofParams"
	ProofModuleStateProofSchemaVersion	ProofModuleStateObjectV2	= "ProofSchemaVersion"
	ProofModuleStateResolutionCacheRecord	ProofModuleStateObjectV2	= "ResolutionCacheRecord"
	ProofModuleStateProofPathDescriptor	ProofModuleStateObjectV2	= "ProofPathDescriptor"

	ProofModuleMsgInvalidateResolutionCache	ProofModuleMessageNameV2	= "MsgInvalidateResolutionCache"

	ProofModuleQueryResolutionProof			ProofModuleQueryNameV2	= "QueryResolutionProof"
	ProofModuleQueryRecursiveResolutionProof	ProofModuleQueryNameV2	= "QueryRecursiveResolutionProof"
	ProofModuleQueryReverseResolutionProof		ProofModuleQueryNameV2	= "QueryReverseResolutionProof"
	ProofModuleQueryNonExistenceProof		ProofModuleQueryNameV2	= "QueryNonExistenceProof"
	ProofModuleQueryProofSchema			ProofModuleQueryNameV2	= "QueryProofSchema"

	ProofModuleFailureInconsistentRecordVersions	ProofModuleFailureModeV2	= "proof_inconsistent_record_versions"
	ProofModuleFailureMissingDelegationConstraint	ProofModuleFailureModeV2	= "recursive_path_omits_delegation_constraint"
	ProofModuleFailureStaleCacheAfterInvalidation	ProofModuleFailureModeV2	= "cache_proof_remains_after_invalidation"
	ProofModuleFailureMalformedNameNonExistence	ProofModuleFailureModeV2	= "non_existence_proof_for_malformed_name"
	ProofModuleFailureProofHeightPruned		ProofModuleFailureModeV2	= "proof_height_unavailable_after_pruning"

	ProofModuleIntegrationIdentityCore	ProofModuleIntegrationPointV2	= "identity_core"
	ProofModuleIntegrationResolverModule	ProofModuleIntegrationPointV2	= "resolver_module"
	ProofModuleIntegrationSubdomainModule	ProofModuleIntegrationPointV2	= "subdomain_module"
	ProofModuleIntegrationStoreV2ProofAPIs	ProofModuleIntegrationPointV2	= "store_v2_proof_apis"
	ProofModuleIntegrationAdaptiveSync	ProofModuleIntegrationPointV2	= "adaptive_sync_state_snapshots"
)

type AuctionModuleFailureCoverageV2 struct {
	Mode		AuctionModuleFailureModeV2
	Guard		string
	StoreScope	string
}

type AuctionModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]AuctionModuleStateObjectV2
	Messages		[]AuctionModuleMessageNameV2
	Queries			[]AuctionModuleQueryNameV2
	FailureModes		[]AuctionModuleFailureCoverageV2
	IntegrationPoints	[]AuctionModuleIntegrationPointV2
	BackingPrimitives	[]string
	StoreKeys		[]string
	BreakdownHash		string
}

type ProofModuleFailureCoverageV2 struct {
	Mode		ProofModuleFailureModeV2
	Guard		string
	StoreScope	string
}

type ProofVerificationModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]ProofModuleStateObjectV2
	Messages		[]ProofModuleMessageNameV2
	Queries			[]ProofModuleQueryNameV2
	FailureModes		[]ProofModuleFailureCoverageV2
	IntegrationPoints	[]ProofModuleIntegrationPointV2
	BackingPrimitives	[]string
	StoreKeys		[]string
	BreakdownHash		string
}

type ProofModuleSchemaDescriptorV2 struct {
	SchemaVersion		uint64
	ResolutionFields	[]string
	RecursiveFields		[]string
	SupportedQueries	[]ProofModuleQueryNameV2
	OptionalMessage		ProofModuleMessageNameV2
	DescriptorHash		string
}

func DefaultAuctionModuleBreakdownV2() (AuctionModuleBreakdownV2, error) {
	breakdown := AuctionModuleBreakdownV2{
		ModulePath:	AuctionModulePathV2,
		Purpose: []string{
			"deterministic_winner_selection",
			"fee_split_accounting",
			"refund_accounting",
			"sealed_bid_commit_reveal",
			"timestamp_based_auction_flow",
		},
		StateObjects:	requiredAuctionModuleStateObjectsV2(),
		Messages:	requiredAuctionModuleMessagesV2(),
		Queries:	requiredAuctionModuleQueriesV2(),
		FailureModes: []AuctionModuleFailureCoverageV2{
			{Mode: AuctionModuleFailureBidReplay, Guard: "CommitAuctionBid", StoreScope: IdentityStoreV2SpecAuctionsPrefix},
			{Mode: AuctionModuleFailureFinalizeBeforeRevealEnd, Guard: "FinalizeSealedAuctionFairV2", StoreScope: IdentityStoreV2SpecAuctionsPrefix},
			{Mode: AuctionModuleFailureIncorrectRefundAccounting, Guard: "ValidateAuctionModuleRefundAccountingV2", StoreScope: IdentityStoreV2SpecAuctionsPrefix},
			{Mode: AuctionModuleFailureNonDeterministicTie, Guard: "ValidateAuctionModuleDeterministicWinnerV2", StoreScope: IdentityStoreV2SpecAuctionsPrefix},
			{Mode: AuctionModuleFailureRevealOutsideWindow, Guard: "RevealAuctionBid", StoreScope: IdentityStoreV2SpecAuctionsPrefix},
		},
		IntegrationPoints:	requiredAuctionModuleIntegrationPointsV2(),
		BackingPrimitives: []string{
			"ComputeAuctionCommitmentV2",
			"DescribeIdentityAuctionStateMachineV2",
			"FinalizeSealedAuctionFairV2",
			"SplitIdentityAuctionFeeV2",
			"StartSealedAuction",
		},
		StoreKeys: []string{
			IdentityStoreV2SpecAuctionsByNamePrefix,
			IdentityStoreV2SpecAuctionsPrefix,
		},
	}
	return NewAuctionModuleBreakdownV2(breakdown)
}

func DefaultProofVerificationModuleBreakdownV2() (ProofVerificationModuleBreakdownV2, error) {
	breakdown := ProofVerificationModuleBreakdownV2{
		ModulePath:	ProofVerificationModulePathV2,
		Purpose: []string{
			"full_node_proof_assembly",
			"light_client_verification",
			"non_existence_proofs",
			"recursive_resolution_proofs",
			"wallet_tooling_verification",
		},
		StateObjects:	requiredProofModuleStateObjectsV2(),
		Messages:	requiredProofModuleMessagesV2(),
		Queries:	requiredProofModuleQueriesV2(),
		FailureModes: []ProofModuleFailureCoverageV2{
			{Mode: ProofModuleFailureInconsistentRecordVersions, Guard: "ValidateProofModuleResolutionProofV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: ProofModuleFailureMalformedNameNonExistence, Guard: "BuildProofModuleNonExistenceProofV2", StoreScope: IdentityStoreV2SpecDomainNamesPrefix},
			{Mode: ProofModuleFailureMissingDelegationConstraint, Guard: "ValidateProofModuleRecursiveProofV2", StoreScope: IdentityStoreV2SpecDelegationsPrefix},
			{Mode: ProofModuleFailureProofHeightPruned, Guard: "ValidateProofModuleHeightAvailableV2", StoreScope: IdentityStoreV2Prefix + "/proof_window"},
			{Mode: ProofModuleFailureStaleCacheAfterInvalidation, Guard: "ValidateProofModuleCacheUseV2", StoreScope: IdentityStoreV2SpecResolutionCachePrefix},
		},
		IntegrationPoints:	requiredProofModuleIntegrationPointsV2(),
		BackingPrimitives: []string{
			"BuildIdentityResolutionProofFormatV2",
			"BuildRecursiveResolutionProofV2",
			"ValidateIdentityResolutionProofFormatV2",
			"ValidateRecursiveResolutionProofV2",
			"VerifyIdentityResolutionProofLightClientV2",
		},
		StoreKeys: []string{
			IdentityStoreV2SpecDelegationsPrefix,
			IdentityStoreV2SpecDomainsPrefix,
			IdentityStoreV2SpecNFTBindingsByNamePrefix,
			IdentityStoreV2SpecResolutionCachePrefix,
			IdentityStoreV2SpecResolversPrefix,
			IdentityStoreV2SpecReversePrefix,
		},
	}
	return NewProofVerificationModuleBreakdownV2(breakdown)
}

func NewAuctionModuleBreakdownV2(breakdown AuctionModuleBreakdownV2) (AuctionModuleBreakdownV2, error) {
	breakdown = canonicalAuctionModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return AuctionModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeAuctionModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func NewProofVerificationModuleBreakdownV2(breakdown ProofVerificationModuleBreakdownV2) (ProofVerificationModuleBreakdownV2, error) {
	breakdown = canonicalProofModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return ProofVerificationModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeProofVerificationModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func (breakdown AuctionModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != AuctionModulePathV2 {
		return errors.New("auction module breakdown must describe auction-module")
	}
	if err := validateBreakdownTokenSetV2("auction purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("auction module", "state object", breakdown.StateObjects, requiredAuctionModuleStateObjectsV2(), IsAuctionModuleStateObjectV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("auction module", "message", breakdown.Messages, requiredAuctionModuleMessagesV2(), IsAuctionModuleMessageNameV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("auction module", "query", breakdown.Queries, requiredAuctionModuleQueriesV2(), IsAuctionModuleQueryNameV2); err != nil {
		return err
	}
	if err := validateAuctionModuleFailuresV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("auction module", "integration", breakdown.IntegrationPoints, requiredAuctionModuleIntegrationPointsV2(), IsAuctionModuleIntegrationPointV2); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("auction backing primitive", breakdown.BackingPrimitives, []string{"ComputeAuctionCommitmentV2", "DescribeIdentityAuctionStateMachineV2", "FinalizeSealedAuctionFairV2", "SplitIdentityAuctionFeeV2", "StartSealedAuction"}); err != nil {
		return err
	}
	if err := validateBreakdownStoreKeysV2("auction", breakdown.StoreKeys, []string{IdentityStoreV2SpecAuctionsByNamePrefix, IdentityStoreV2SpecAuctionsPrefix}); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("auction module breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown AuctionModuleBreakdownV2) Validate() error {
	breakdown = canonicalAuctionModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("auction module breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeAuctionModuleBreakdownHashV2(breakdown) {
		return errors.New("auction module breakdown hash mismatch")
	}
	return nil
}

func (breakdown ProofVerificationModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != ProofVerificationModulePathV2 {
		return errors.New("proof verification module breakdown must describe proof-verification-module")
	}
	if err := validateBreakdownTokenSetV2("proof purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("proof verification module", "state object", breakdown.StateObjects, requiredProofModuleStateObjectsV2(), IsProofModuleStateObjectV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("proof verification module", "message", breakdown.Messages, requiredProofModuleMessagesV2(), IsProofModuleMessageNameV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("proof verification module", "query", breakdown.Queries, requiredProofModuleQueriesV2(), IsProofModuleQueryNameV2); err != nil {
		return err
	}
	if err := validateProofModuleFailuresV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("proof verification module", "integration", breakdown.IntegrationPoints, requiredProofModuleIntegrationPointsV2(), IsProofModuleIntegrationPointV2); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("proof backing primitive", breakdown.BackingPrimitives, []string{"BuildIdentityResolutionProofFormatV2", "BuildRecursiveResolutionProofV2", "ValidateIdentityResolutionProofFormatV2", "ValidateRecursiveResolutionProofV2", "VerifyIdentityResolutionProofLightClientV2"}); err != nil {
		return err
	}
	if err := validateBreakdownStoreKeysV2("proof", breakdown.StoreKeys, []string{IdentityStoreV2SpecDelegationsPrefix, IdentityStoreV2SpecDomainsPrefix, IdentityStoreV2SpecNFTBindingsByNamePrefix, IdentityStoreV2SpecResolutionCachePrefix, IdentityStoreV2SpecResolversPrefix, IdentityStoreV2SpecReversePrefix}); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("proof verification module breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown ProofVerificationModuleBreakdownV2) Validate() error {
	breakdown = canonicalProofModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("proof verification module breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeProofVerificationModuleBreakdownHashV2(breakdown) {
		return errors.New("proof verification module breakdown hash mismatch")
	}
	return nil
}

func ValidateAuctionModuleDeterministicWinnerV2(auction Auction, params IdentityAuctionFairnessParamsV2) error {
	if err := validateAuction(auction); err != nil {
		return err
	}
	if err := ValidateIdentityAuctionFairnessParamsV2(params); err != nil {
		return err
	}
	if auction.Phase != AuctionPhaseFinalized {
		return errors.New("auction module deterministic winner requires finalized auction")
	}
	if len(auction.Reveals) == 0 {
		return errors.New("auction module deterministic winner requires revealed bids")
	}
	winner := chooseAuctionWinnerByRuleV2(auction.Reveals, params.TieBreakRule)
	if !addressesEqual(winner.Bidder, auction.Winner) || winner.Bid != auction.WinningBid || winner.CommitmentHash != auction.WinningCommitment {
		return fmt.Errorf("%s: winner does not match configured tie-break rule", AuctionModuleFailureNonDeterministicTie)
	}
	return nil
}

func ValidateAuctionModuleRefundAccountingV2(result IdentityAuctionFairFinalizationV2, params IdentityAuctionFairnessParamsV2) error {
	if err := ValidateIdentityAuctionFairnessParamsV2(params); err != nil {
		return err
	}
	if err := validateAuction(result.Auction); err != nil {
		return err
	}
	winningSplit := result.FeeSplit.Burn.Add(result.FeeSplit.Treasury).Add(result.FeeSplit.Rewards).Add(result.FeeSplit.CommunityPool)
	if !winningSplit.Equal(sdkmath.NewIntFromUint64(result.WinningBid)) {
		return fmt.Errorf("%s: fee split does not equal winning bid", AuctionModuleFailureIncorrectRefundAccounting)
	}
	if uint64(len(result.LosingBidRefunds)) > uint64(len(result.Auction.Reveals)) {
		return fmt.Errorf("%s: too many losing bid refunds", AuctionModuleFailureIncorrectRefundAccounting)
	}
	expectedForfeits := buildAuctionUnrevealedForfeitsV2(result.Auction, params)
	if len(expectedForfeits) != len(result.UnrevealedForfeits) {
		return fmt.Errorf("%s: unrevealed forfeit count mismatch", AuctionModuleFailureIncorrectRefundAccounting)
	}
	for i := range expectedForfeits {
		if expectedForfeits[i].ReceiptID != result.UnrevealedForfeits[i].ReceiptID || expectedForfeits[i].Amount != result.UnrevealedForfeits[i].Amount {
			return fmt.Errorf("%s: unrevealed forfeit mismatch", AuctionModuleFailureIncorrectRefundAccounting)
		}
	}
	return nil
}

func FinalizeAuctionModuleV2(state IdentityState, name string, finalizer sdk.AccAddress, height uint64, params IdentityAuctionFairnessParamsV2) (IdentityState, IdentityAuctionFairFinalizationV2, error) {
	next, result, err := FinalizeSealedAuctionFairV2(state, name, finalizer, height, params)
	if err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	if err := ValidateAuctionModuleDeterministicWinnerV2(result.Auction, params); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	if err := ValidateAuctionModuleRefundAccountingV2(result, params); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	return next, result, nil
}

func ValidateProofModuleResolutionProofV2(proof IdentityResolutionProofFormatV2) error {
	if err := ValidateIdentityResolutionProofFormatV2(proof); err != nil {
		return err
	}
	if proof.ResolverRecord != nil && proof.RecordVersion != proof.ResolverRecord.RecordVersion {
		return fmt.Errorf("%s: proof record_version %d resolver version %d", ProofModuleFailureInconsistentRecordVersions, proof.RecordVersion, proof.ResolverRecord.RecordVersion)
	}
	return nil
}

func ValidateProofModuleRecursiveProofV2(proof RecursiveResolutionProofV2) error {
	if err := ValidateRecursiveResolutionProofV2(proof); err != nil {
		return err
	}
	if len(proof.PathLabels) > 1 && len(proof.PathDomainRecords) == 0 && len(proof.PathDelegationRecords) == 0 {
		return fmt.Errorf("%s: recursive path has no domain or delegation constraints", ProofModuleFailureMissingDelegationConstraint)
	}
	return nil
}

func ValidateProofModuleCacheUseV2(record ResolutionCacheRecordV2, ctx ResolutionCacheUseContextV2) error {
	if err := ValidateResolutionCacheRecordV2Use(record, ctx); err != nil {
		return fmt.Errorf("%s: %w", ProofModuleFailureStaleCacheAfterInvalidation, err)
	}
	return nil
}

func ValidateProofModuleHeightAvailableV2(height uint64, earliestAvailableHeight uint64, latestAvailableHeight uint64) error {
	if height == 0 || earliestAvailableHeight == 0 || latestAvailableHeight == 0 {
		return errors.New("proof module height window values must be positive")
	}
	if earliestAvailableHeight > latestAvailableHeight {
		return errors.New("proof module height window is invalid")
	}
	if height < earliestAvailableHeight || height > latestAvailableHeight {
		return fmt.Errorf("%s: height %d outside [%d,%d]", ProofModuleFailureProofHeightPruned, height, earliestAvailableHeight, latestAvailableHeight)
	}
	return nil
}

func BuildProofModuleResolutionProofV2(state IdentityState, chainID string, appHash string, name string, height uint64, ttl uint64) (IdentityResolutionProofFormatV2, error) {
	proof, err := BuildIdentityResolutionProofFormatV2(state, chainID, appHash, name, IdentityProofQueryResolveRecord, height, ttl, nil)
	if err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	return proof, ValidateProofModuleResolutionProofV2(proof)
}

func BuildProofModuleReverseResolutionProofV2(state IdentityState, chainID string, appHash string, name string, height uint64, ttl uint64, reverseAddress sdk.AccAddress) (IdentityResolutionProofFormatV2, error) {
	proof, err := BuildIdentityResolutionProofFormatV2(state, chainID, appHash, name, IdentityProofQueryResolveReverse, height, ttl, reverseAddress)
	if err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	return proof, ValidateProofModuleResolutionProofV2(proof)
}

func BuildProofModuleNonExistenceProofV2(state IdentityState, chainID string, appHash string, name string, height uint64, ttl uint64) (IdentityResolutionProofFormatV2, error) {
	if _, err := NormalizeAETDomain(name); err != nil {
		return IdentityResolutionProofFormatV2{}, fmt.Errorf("%s: %w", ProofModuleFailureMalformedNameNonExistence, err)
	}
	proof, err := BuildIdentityResolutionProofFormatV2(state, chainID, appHash, name, IdentityProofQueryDomainAbsent, height, ttl, nil)
	if err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	return proof, ValidateProofModuleResolutionProofV2(proof)
}

func BuildProofModuleSchemaDescriptorV2() ProofModuleSchemaDescriptorV2 {
	descriptor := ProofModuleSchemaDescriptorV2{
		SchemaVersion:		IdentityProofSchemaVersionV2,
		ResolutionFields:	append([]string(nil), IdentityResolutionProofFormatV2FieldOrder...),
		RecursiveFields:	append([]string(nil), RecursiveResolutionProofV2FieldOrder...),
		SupportedQueries:	requiredProofModuleQueriesV2(),
		OptionalMessage:	ProofModuleMsgInvalidateResolutionCache,
	}
	descriptor.DescriptorHash = ComputeProofModuleSchemaDescriptorHashV2(descriptor)
	return descriptor
}

func ValidateProofModuleSchemaDescriptorV2(descriptor ProofModuleSchemaDescriptorV2) error {
	if descriptor.SchemaVersion != IdentityProofSchemaVersionV2 {
		return fmt.Errorf("unsupported proof module schema version %d", descriptor.SchemaVersion)
	}
	if err := validateProofModuleFieldOrderV2("resolution", descriptor.ResolutionFields, IdentityResolutionProofFormatV2FieldOrder); err != nil {
		return err
	}
	if err := validateProofModuleFieldOrderV2("recursive", descriptor.RecursiveFields, RecursiveResolutionProofV2FieldOrder); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("proof schema", "query", descriptor.SupportedQueries, requiredProofModuleQueriesV2(), IsProofModuleQueryNameV2); err != nil {
		return err
	}
	if descriptor.OptionalMessage != ProofModuleMsgInvalidateResolutionCache {
		return errors.New("proof schema optional message must be MsgInvalidateResolutionCache")
	}
	if descriptor.DescriptorHash == "" || descriptor.DescriptorHash != ComputeProofModuleSchemaDescriptorHashV2(descriptor) {
		return errors.New("proof schema descriptor hash mismatch")
	}
	return nil
}

func ComputeAuctionModuleBreakdownHashV2(breakdown AuctionModuleBreakdownV2) string {
	breakdown = canonicalAuctionModuleBreakdownV2(breakdown)
	parts := []string{"aetra-auction-module-breakdown-v2", breakdown.ModulePath}
	parts = appendBreakdownStringsV2(parts, "purpose", breakdown.Purpose)
	for _, value := range breakdown.StateObjects {
		parts = append(parts, "state", string(value))
	}
	for _, value := range breakdown.Messages {
		parts = append(parts, "message", string(value))
	}
	for _, value := range breakdown.Queries {
		parts = append(parts, "query", string(value))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	for _, value := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(value))
	}
	parts = appendBreakdownStringsV2(parts, "primitive", breakdown.BackingPrimitives)
	parts = appendBreakdownStringsV2(parts, "store", breakdown.StoreKeys)
	return identityHash(parts...)
}

func ComputeProofVerificationModuleBreakdownHashV2(breakdown ProofVerificationModuleBreakdownV2) string {
	breakdown = canonicalProofModuleBreakdownV2(breakdown)
	parts := []string{"aetra-proof-verification-module-breakdown-v2", breakdown.ModulePath}
	parts = appendBreakdownStringsV2(parts, "purpose", breakdown.Purpose)
	for _, value := range breakdown.StateObjects {
		parts = append(parts, "state", string(value))
	}
	for _, value := range breakdown.Messages {
		parts = append(parts, "message", string(value))
	}
	for _, value := range breakdown.Queries {
		parts = append(parts, "query", string(value))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	for _, value := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(value))
	}
	parts = appendBreakdownStringsV2(parts, "primitive", breakdown.BackingPrimitives)
	parts = appendBreakdownStringsV2(parts, "store", breakdown.StoreKeys)
	return identityHash(parts...)
}

func ComputeProofModuleSchemaDescriptorHashV2(descriptor ProofModuleSchemaDescriptorV2) string {
	descriptor.ResolutionFields = sortedBreakdownStringsV2(descriptor.ResolutionFields)
	descriptor.RecursiveFields = sortedBreakdownStringsV2(descriptor.RecursiveFields)
	descriptor.SupportedQueries = sortedBreakdownTypedStringsV2(descriptor.SupportedQueries)
	parts := []string{"aetra-proof-module-schema-descriptor-v2", fmt.Sprint(descriptor.SchemaVersion)}
	parts = appendBreakdownStringsV2(parts, "resolution", descriptor.ResolutionFields)
	parts = appendBreakdownStringsV2(parts, "recursive", descriptor.RecursiveFields)
	for _, query := range descriptor.SupportedQueries {
		parts = append(parts, "query", string(query))
	}
	parts = append(parts, "optional-message", string(descriptor.OptionalMessage))
	return identityHash(parts...)
}

func IsAuctionModuleStateObjectV2(value AuctionModuleStateObjectV2) bool {
	switch value {
	case AuctionModuleStateAuctionRecord, AuctionModuleStateBidCommitment, AuctionModuleStateRevealedBid, AuctionModuleStateAuctionParams, AuctionModuleStateAuctionFeeSplit:
		return true
	default:
		return false
	}
}

func IsAuctionModuleMessageNameV2(value AuctionModuleMessageNameV2) bool {
	switch value {
	case AuctionModuleMsgStartAuction, AuctionModuleMsgCommitBid, AuctionModuleMsgRevealBid, AuctionModuleMsgFinalizeAuction, AuctionModuleMsgCancelExpiredAuction, AuctionModuleMsgClaimAuctionRefund:
		return true
	default:
		return false
	}
}

func IsAuctionModuleQueryNameV2(value AuctionModuleQueryNameV2) bool {
	switch value {
	case AuctionModuleQueryAuction, AuctionModuleQueryAuctionByName, AuctionModuleQueryBidCommitment, AuctionModuleQueryAuctionParams, AuctionModuleQueryAuctionPriceFloor:
		return true
	default:
		return false
	}
}

func IsAuctionModuleFailureModeV2(value AuctionModuleFailureModeV2) bool {
	switch value {
	case AuctionModuleFailureBidReplay, AuctionModuleFailureRevealOutsideWindow, AuctionModuleFailureNonDeterministicTie, AuctionModuleFailureFinalizeBeforeRevealEnd, AuctionModuleFailureIncorrectRefundAccounting:
		return true
	default:
		return false
	}
}

func IsAuctionModuleIntegrationPointV2(value AuctionModuleIntegrationPointV2) bool {
	switch value {
	case AuctionModuleIntegrationIdentityCore, AuctionModuleIntegrationBankModule, AuctionModuleIntegrationFeeModule, AuctionModuleIntegrationTreasuryBurnRewardsCommunityPool, AuctionModuleIntegrationStoreV2:
		return true
	default:
		return false
	}
}

func IsProofModuleStateObjectV2(value ProofModuleStateObjectV2) bool {
	switch value {
	case ProofModuleStateProofParams, ProofModuleStateProofSchemaVersion, ProofModuleStateResolutionCacheRecord, ProofModuleStateProofPathDescriptor:
		return true
	default:
		return false
	}
}

func IsProofModuleMessageNameV2(value ProofModuleMessageNameV2) bool {
	return value == ProofModuleMsgInvalidateResolutionCache
}

func IsProofModuleQueryNameV2(value ProofModuleQueryNameV2) bool {
	switch value {
	case ProofModuleQueryResolutionProof, ProofModuleQueryRecursiveResolutionProof, ProofModuleQueryReverseResolutionProof, ProofModuleQueryNonExistenceProof, ProofModuleQueryProofSchema:
		return true
	default:
		return false
	}
}

func IsProofModuleFailureModeV2(value ProofModuleFailureModeV2) bool {
	switch value {
	case ProofModuleFailureInconsistentRecordVersions, ProofModuleFailureMissingDelegationConstraint, ProofModuleFailureStaleCacheAfterInvalidation, ProofModuleFailureMalformedNameNonExistence, ProofModuleFailureProofHeightPruned:
		return true
	default:
		return false
	}
}

func IsProofModuleIntegrationPointV2(value ProofModuleIntegrationPointV2) bool {
	switch value {
	case ProofModuleIntegrationIdentityCore, ProofModuleIntegrationResolverModule, ProofModuleIntegrationSubdomainModule, ProofModuleIntegrationStoreV2ProofAPIs, ProofModuleIntegrationAdaptiveSync:
		return true
	default:
		return false
	}
}

func requiredAuctionModuleStateObjectsV2() []AuctionModuleStateObjectV2 {
	return []AuctionModuleStateObjectV2{AuctionModuleStateAuctionFeeSplit, AuctionModuleStateAuctionParams, AuctionModuleStateAuctionRecord, AuctionModuleStateBidCommitment, AuctionModuleStateRevealedBid}
}

func requiredAuctionModuleMessagesV2() []AuctionModuleMessageNameV2 {
	return []AuctionModuleMessageNameV2{AuctionModuleMsgCancelExpiredAuction, AuctionModuleMsgClaimAuctionRefund, AuctionModuleMsgCommitBid, AuctionModuleMsgFinalizeAuction, AuctionModuleMsgRevealBid, AuctionModuleMsgStartAuction}
}

func requiredAuctionModuleQueriesV2() []AuctionModuleQueryNameV2 {
	return []AuctionModuleQueryNameV2{AuctionModuleQueryAuction, AuctionModuleQueryAuctionByName, AuctionModuleQueryAuctionParams, AuctionModuleQueryAuctionPriceFloor, AuctionModuleQueryBidCommitment}
}

func requiredAuctionModuleFailuresV2() []AuctionModuleFailureModeV2 {
	return []AuctionModuleFailureModeV2{AuctionModuleFailureBidReplay, AuctionModuleFailureFinalizeBeforeRevealEnd, AuctionModuleFailureIncorrectRefundAccounting, AuctionModuleFailureNonDeterministicTie, AuctionModuleFailureRevealOutsideWindow}
}

func requiredAuctionModuleIntegrationPointsV2() []AuctionModuleIntegrationPointV2 {
	return []AuctionModuleIntegrationPointV2{AuctionModuleIntegrationBankModule, AuctionModuleIntegrationFeeModule, AuctionModuleIntegrationIdentityCore, AuctionModuleIntegrationStoreV2, AuctionModuleIntegrationTreasuryBurnRewardsCommunityPool}
}

func requiredProofModuleStateObjectsV2() []ProofModuleStateObjectV2 {
	return []ProofModuleStateObjectV2{ProofModuleStateProofParams, ProofModuleStateProofPathDescriptor, ProofModuleStateProofSchemaVersion, ProofModuleStateResolutionCacheRecord}
}

func requiredProofModuleMessagesV2() []ProofModuleMessageNameV2 {
	return []ProofModuleMessageNameV2{ProofModuleMsgInvalidateResolutionCache}
}

func requiredProofModuleQueriesV2() []ProofModuleQueryNameV2 {
	return []ProofModuleQueryNameV2{ProofModuleQueryNonExistenceProof, ProofModuleQueryProofSchema, ProofModuleQueryRecursiveResolutionProof, ProofModuleQueryResolutionProof, ProofModuleQueryReverseResolutionProof}
}

func requiredProofModuleFailuresV2() []ProofModuleFailureModeV2 {
	return []ProofModuleFailureModeV2{ProofModuleFailureInconsistentRecordVersions, ProofModuleFailureMalformedNameNonExistence, ProofModuleFailureMissingDelegationConstraint, ProofModuleFailureProofHeightPruned, ProofModuleFailureStaleCacheAfterInvalidation}
}

func requiredProofModuleIntegrationPointsV2() []ProofModuleIntegrationPointV2 {
	return []ProofModuleIntegrationPointV2{ProofModuleIntegrationAdaptiveSync, ProofModuleIntegrationIdentityCore, ProofModuleIntegrationResolverModule, ProofModuleIntegrationStoreV2ProofAPIs, ProofModuleIntegrationSubdomainModule}
}

func canonicalAuctionModuleBreakdownV2(breakdown AuctionModuleBreakdownV2) AuctionModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedBreakdownStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedBreakdownTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedBreakdownTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedBreakdownTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedAuctionModuleFailuresV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedBreakdownTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.BackingPrimitives = sortedBreakdownStringsV2(breakdown.BackingPrimitives)
	breakdown.StoreKeys = sortedBreakdownStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func canonicalProofModuleBreakdownV2(breakdown ProofVerificationModuleBreakdownV2) ProofVerificationModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedBreakdownStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedBreakdownTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedBreakdownTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedBreakdownTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedProofModuleFailuresV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedBreakdownTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.BackingPrimitives = sortedBreakdownStringsV2(breakdown.BackingPrimitives)
	breakdown.StoreKeys = sortedBreakdownStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateAuctionModuleFailuresV2(values []AuctionModuleFailureCoverageV2) error {
	if len(values) != len(requiredAuctionModuleFailuresV2()) {
		return fmt.Errorf("auction module expected %d failure modes", len(requiredAuctionModuleFailuresV2()))
	}
	seen := map[AuctionModuleFailureModeV2]struct{}{}
	for _, value := range values {
		if !IsAuctionModuleFailureModeV2(value.Mode) {
			return fmt.Errorf("auction module unknown failure mode %q", value.Mode)
		}
		if strings.TrimSpace(value.Guard) == "" || !strings.HasPrefix(value.StoreScope, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("auction module failure %s has invalid guard or store scope", value.Mode)
		}
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("auction module duplicate failure mode %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, required := range requiredAuctionModuleFailuresV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("auction module missing failure mode %s", required)
		}
	}
	return nil
}

func validateProofModuleFailuresV2(values []ProofModuleFailureCoverageV2) error {
	if len(values) != len(requiredProofModuleFailuresV2()) {
		return fmt.Errorf("proof verification module expected %d failure modes", len(requiredProofModuleFailuresV2()))
	}
	seen := map[ProofModuleFailureModeV2]struct{}{}
	for _, value := range values {
		if !IsProofModuleFailureModeV2(value.Mode) {
			return fmt.Errorf("proof verification module unknown failure mode %q", value.Mode)
		}
		if strings.TrimSpace(value.Guard) == "" || !strings.HasPrefix(value.StoreScope, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("proof verification module failure %s has invalid guard or store scope", value.Mode)
		}
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("proof verification module duplicate failure mode %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, required := range requiredProofModuleFailuresV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("proof verification module missing failure mode %s", required)
		}
	}
	return nil
}

func sortedAuctionModuleFailuresV2(values []AuctionModuleFailureCoverageV2) []AuctionModuleFailureCoverageV2 {
	out := append([]AuctionModuleFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedProofModuleFailuresV2(values []ProofModuleFailureCoverageV2) []ProofModuleFailureCoverageV2 {
	out := append([]ProofModuleFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func validateProofModuleFieldOrderV2(label string, got []string, expected []string) error {
	if len(got) != len(expected) {
		return fmt.Errorf("proof schema %s field order expected %d fields", label, len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			return fmt.Errorf("proof schema %s field order mismatch at %d: expected %s got %s", label, i, expected[i], got[i])
		}
	}
	return nil
}
