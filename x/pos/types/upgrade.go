package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	DefaultDelegationActivationEpochs	= uint64(1)
	DefaultEvidenceWindowEpochs		= uint64(4)
	DefaultMinTaskGroupValidators		= uint32(3)
	DefaultMaxTaskGroupValidators		= uint32(21)
	DefaultReporterRewardBps		= uint32(500)
	MaxReporterRewardBps			= uint32(2_000)

	PosHashHexLength	= 64
	PosEmptyRootHash	= "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	DefaultWorkloadClass	= "general"
	maxPosTokenLength	= 128
)

const (
	CentralizationWarningValidatorShare		= "validator_voting_power_concentration"
	CentralizationWarningTopNShare			= "top_n_voting_power_concentration"
	CentralizationWarningStakeSaturation		= "stake_saturation_ratio"
	CentralizationWarningDelegationRisk		= "delegation_risk_concentration"
	CentralizationWarningRewardDampeningActive	= "reward_dampening_active"
	CentralizationWarningTaskAssignmentShare	= "task_assignment_concentration"
	CentralizationWarningBootstrapEligible		= "bootstrap_eligible_reliable_validator"

	ConcentrationAlertSeverityWarning	= "warning"
	ConcentrationAlertSeverityCritical	= "critical"
)

type EpochPhase string

const (
	EpochPhaseDelegation	EpochPhase	= "delegation"
	EpochPhaseElection	EpochPhase	= "election"
	EpochPhaseAssignment	EpochPhase	= "assignment"
	EpochPhaseActive	EpochPhase	= "active"
	EpochPhaseSettlement	EpochPhase	= "settlement"
	EpochPhaseClosed	EpochPhase	= "closed"
)

type SettlementStatus string

const (
	SettlementStatusPending		SettlementStatus	= "pending"
	SettlementStatusFinalized	SettlementStatus	= "finalized"
)

type ValidatorRole string

const (
	ValidatorRoleValidator		ValidatorRole	= "validator"
	ValidatorRoleProposer		ValidatorRole	= "proposer"
	ValidatorRoleBlockProducer	ValidatorRole	= "block_producer"
	ValidatorRoleVerifier		ValidatorRole	= "verifier"
	ValidatorRoleEvidenceReporter	ValidatorRole	= "evidence_reporter"
	ValidatorRoleCollator		ValidatorRole	= "collator"
	ValidatorRoleDelegationOperator	ValidatorRole	= "delegation_operator"
	ValidatorRoleFisherman		ValidatorRole	= "fisherman"
	ValidatorRoleEvidenceReviewer	ValidatorRole	= "evidence_reviewer"
)

const (
	RoleStatusEligible	= "eligible"
	RoleStatusAssigned	= "assigned"
	RoleStatusSuspended	= "suspended"
	RoleStatusInactive	= "inactive"
)

const (
	CollatorStatusRegistered	= "registered"
	CollatorStatusActive		= "active"
	CollatorStatusSuspended		= "suspended"
	CollatorStatusRetired		= "retired"
)

type RoleRecord struct {
	ValidatorAddress	string
	Role			ValidatorRole
	EpochID			uint64
	Status			string
	EligibilityScore	uint32
	Capacity		ValidatorCapacity
	AssignedTaskCount	uint32
	PerformanceScore	uint32
}

type RoleRule struct {
	Role			ValidatorRole
	Description		string
	RequiresValidator	bool
	RequiresMinimumStake	bool
	RequiresDeposit		bool
	RequiresAuthorization	bool
	RequiresFeeDisclosure	bool
	RequiresRiskPolicy	bool
	CanFinalize		bool
	RewardWeightBps		uint32
	MinimumPerformanceBps	uint32
	MinimumEligibilityBps	uint32
}

type RoleRegistry struct {
	Rules []RoleRule
}

type RoleEligibilityInput struct {
	Params				Params
	Role				ValidatorRole
	ActorAddress			string
	Candidate			Candidate
	DepositNaet			sdkmath.Int
	DelegationOperatorAuthorized	bool
	FeesDisclosed			bool
	RiskPolicyDisclosed		bool
}

type RolePerformanceMetrics struct {
	ValidatorAddress	string
	Role			ValidatorRole
	EpochID			uint64
	AssignedTasks		uint32
	CompletedTasks		uint32
	FaultedTasks		uint32
	MissedTasks		uint32
	PerformanceScore	uint32
}

type RoleRewardInput struct {
	EpochID			uint64
	TotalRewardsNaet	sdkmath.Int
	Records			[]RoleRecord
	Weights			[]RoleRewardWeight
}

type CollatorRecord struct {
	CollatorID		string
	OperatorAddress		string
	SupportedWorkloads	[]WorkloadType
	BondOptional		sdkmath.Int
	Reputation		uint32
	Status			string
	RegisteredEpoch		uint64
}

type CollatorCandidateOutputInput struct {
	EpochID			uint64
	Collator		CollatorRecord
	Task			WorkloadTask
	TaskGroupIDOptional	string
	TransactionRoot		string
	StateTransitionRoot	string
	ProofBundleRoot		string
}

type CollatorCandidateOutput struct {
	EpochID				uint64
	CollatorID			string
	OperatorAddress			string
	TaskID				string
	TaskGroupIDOptional		string
	WorkloadID			string
	WorkloadType			WorkloadType
	TransactionRoot			string
	StateTransitionRoot		string
	ProofBundleRoot			string
	RequiresValidatorVerification	bool
	ValidatorSignatures		[]string
	Finalized			bool
	CandidateOutputHash		string
}

type CollatorRegistry struct {
	EpochID		uint64
	Collators	[]CollatorRecord
	RegistryRoot	string
}

const (
	CollatorVerificationResultValid		= "valid"
	CollatorVerificationResultInvalid	= "invalid"
	CollatorVerificationResultAbstain	= "abstain"
)

type CollatorOutputVerification struct {
	OutputHash		string
	ValidatorAddress	string
	Result			string
	SignatureHash		string
	VerifiedHeight		int64
}

type CollatorOutputVerificationResult struct {
	OutputHash		string
	ValidVotes		uint32
	InvalidVotes		uint32
	AbstainVotes		uint32
	TotalValidators		uint32
	ParticipationBps	uint32
	DecisionThresholdBps	uint32
	Accepted		bool
	Rejected		bool
	ValidSignatureHashes	[]string
	VerificationRoot	string
}

type ValidatorCapacity struct {
	MaxTaskGroups		uint32
	SupportedWorkloads	[]WorkloadType
	ZoneSupport		[]string
	HardwareClassOptional	string
	NetworkClassOptional	string
	AvailabilityCommitment	uint32
}

type EpochPhaseDurations struct {
	DelegationSeconds	uint64
	ElectionSeconds		uint64
	AssignmentSeconds	uint64
	ActiveValidationSeconds	uint64
	SettlementSeconds	uint64
}

type EpochSeedSource string

const (
	EpochSeedSourcePreviousSeedValidatorSet	EpochSeedSource	= "previous_seed_validator_set"
	EpochSeedSourceCometBFTBlockID		EpochSeedSource	= "cometbft_block_id"
	EpochSeedSourceExternalBeacon		EpochSeedSource	= "external_beacon"
)

type EpochLifecycleStep struct {
	Phase		EpochPhase
	Name		string
	DurationKey	string
}

type EpochRecord struct {
	EpochID			uint64
	StartHeight		uint64
	EndHeight		uint64
	Phase			EpochPhase
	Seed			string
	ValidatorSetHash	string
	TaskGroupRoot		string
	PerformanceRoot		string
	RewardRoot		string
	SlashRoot		string
	SettlementStatus	SettlementStatus
}

type EpochSettlementRoots struct {
	PerformanceRoot	string
	RewardRoot	string
	SlashRoot	string
}

type WorkloadTask struct {
	TaskID			string
	WorkloadID		string
	WorkloadType		WorkloadType
	ZoneID			string
	ShardID			string
	WorkloadClass		string
	RequiredValidators	uint32
	Roles			[]ValidatorRole
	ExcludedValidators	[]string
}

type WorkloadType string

const (
	WorkloadTypeGlobalConsensus		WorkloadType	= "global_consensus"
	WorkloadTypeZoneExecution		WorkloadType	= "zone_execution"
	WorkloadTypeShardExecution		WorkloadType	= "shard_execution"
	WorkloadTypeProofVerification		WorkloadType	= "proof_verification"
	WorkloadTypeEvidenceVerification	WorkloadType	= "evidence_verification"
	WorkloadTypeDataAvailability		WorkloadType	= "data_availability"
	WorkloadTypeServiceValidation		WorkloadType	= "service_validation"
)

type TaskAssignment struct {
	TaskID		string
	WorkloadID	string
	WorkloadType	WorkloadType
	ZoneID		string
	ShardID		string
	WorkloadClass	string
	Role		ValidatorRole
	Validators	[]string
	AssignmentHash	string
}

type TaskAssignmentSet struct {
	EpochID		uint64
	Seed		string
	Assignments	[]TaskAssignment
	Root		string
}

type TaskGroup struct {
	EpochID			uint64
	TaskGroupID		string
	WorkloadID		string
	WorkloadType		WorkloadType
	ValidatorMembers	[]string
	ProposerOrder		[]string
	VerifierSet		[]string
	MinimumGroupSize	uint32
	StakeWeightRoot		string
	AssignmentSeed		string
	ActivationHeight	uint64
	ExpiryHeight		uint64
}

type TaskGroupSet struct {
	EpochID	uint64
	Seed	string
	Groups	[]TaskGroup
	Root	string
}

type CapacityFaultEvidence struct {
	ValidatorID		string
	WorkloadID		string
	WorkloadType		WorkloadType
	AssignmentEpoch		uint64
	EvidenceHeight		int64
	UsedForAssignment	bool
	Finalized		bool
}

const (
	EvidenceTypeDoubleSignProof			= "double_sign_proof"
	EvidenceTypeInvalidStateTransitionProof		= "invalid_state_transition_proof"
	EvidenceTypeEquivocationProof			= "equivocation_proof"
	EvidenceTypeDowntimeProof			= "downtime_proof"
	EvidenceTypeInvalidTaskExecutionProof		= "invalid_task_execution_proof"
	EvidenceTypeInvalidCollatorOutputProof		= "invalid_collator_output_proof"
	EvidenceTypeInvalidProofAcceptance		= "invalid_proof_acceptance"
	EvidenceTypeFalseCapacityDeclaration		= "false_capacity_declaration"
	EvidenceTypeInvalidEvidenceSubmission		= "invalid_evidence_submission"
	EvidenceStatusSubmitted				= "submitted"
	EvidenceStatusInVerification			= "in_verification"
	EvidenceStatusAccepted				= "accepted"
	EvidenceStatusVerified				= "verified"
	EvidenceStatusRejected				= "rejected"
	EvidenceStatusExpired				= "expired"
	EvidenceStatusFinalized				= "finalized"
	EvidenceStatusSlashed				= "slashed"
	DefaultEvidenceVerificationQuorumBps		= uint32(6_700)
	DefaultEvidenceFinalityQuorumBps		= uint32(6_700)
	DefaultDoubleSignSlashBps			= uint32(5_000)
	DefaultInvalidStateTransitionSlashBps		= uint32(1_500)
	DefaultEquivocationSlashBps			= uint32(2_000)
	DefaultDowntimeSlashBps				= uint32(100)
	DefaultInvalidTaskExecutionSlashBps		= uint32(750)
	DefaultInvalidCollatorOutputSlashBps		= uint32(500)
	DefaultInvalidProofAcceptanceSlashBps		= uint32(1_000)
	DefaultFalseCapacityDeclarationSlashBps		= uint32(500)
	DefaultInvalidEvidenceSubmissionSlashBps	= uint32(250)
)

type EvidenceRecord struct {
	EvidenceID		string
	EvidenceType		string
	AccusedValidator	string
	Reporter		string
	EpochID			uint64
	TaskGroupIDOptional	string
	ObjectHash		string
	ProofPayloadHash	string
	SubmittedHeight		int64
	Status			string
	VerificationGroupID	string
	DecisionHeight		int64
	PenaltyIDOptional	string
}

type EvidenceVerificationGroupInput struct {
	Params			Params
	Epoch			EpochRecord
	ActiveValidators	[]ScoredValidator
	Evidence		EvidenceRecord
	MinimumGroupSize	uint32
	DecisionThresholdBps	uint32
}

type EvidenceVerificationGroup struct {
	EvidenceID		string
	EpochID			uint64
	VerificationGroupID	string
	Members			[]string
	ExcludedValidators	[]string
	MinimumGroupSize	uint32
	DecisionThresholdBps	uint32
	AssignmentSeed		string
	GroupHash		string
}

type StructuredEvidenceRecord struct {
	EvidenceID		string
	EvidenceType		string
	ReporterID		string
	AccusedValidatorID	string
	SubjectID		string
	EvidenceHash		string
	EvidenceHeight		int64
	EvidenceEpoch		uint64
	SubmittedHeight		int64
	VerificationGroupID	string
	Status			string
	StructuredRecordHash	string
}

type EvidenceSlashPolicy struct {
	EvidenceType		string
	Misbehavior		string
	SlashFractionBps	uint32
}

type EvidenceVerificationVote struct {
	EvidenceID	string
	ReviewerID	string
	Accepted	bool
	SignatureHash	string
	VoteHeight	int64
}

type EvidenceVerificationResult struct {
	EvidenceID		string
	AcceptedVotes		uint32
	RejectedVotes		uint32
	TotalReviewers		uint32
	ParticipationBps	uint32
	QuorumBps		uint32
	Accepted		bool
	Rejected		bool
	Status			string
	VerificationRoot	string
	VerificationGroup	string
}

type EvidenceFinalityVote struct {
	EvidenceID	string
	ValidatorID	string
	Approve		bool
	VotingPowerBps	uint32
	SignatureHash	string
	FinalityHeight	int64
}

type EvidenceFinalityDecision struct {
	EvidenceID		string
	AcceptedPowerBps	uint32
	RejectedPowerBps	uint32
	QuorumBps		uint32
	Finalized		bool
	Accepted		bool
	Status			string
	FinalityVoteRoot	string
	FinalityVoteCount	uint32
}

type DelegationIntent struct {
	NominatorID		string
	ValidatorID		string
	StakeNaet		sdkmath.Int
	RequestedEpoch		uint64
	MaxCommissionBps	uint32
	MinPerformanceScoreBps	uint32
}

type DelegationActivation struct {
	ValidatorID	string
	Nominations	[]Nomination
	ActivatedAt	uint64
	IntentCount	uint32
	TotalStake	sdkmath.Int
	ActivationKey	string
}

type UnbondingRiskWindow struct {
	UnbondingEpochs		uint64
	SlashableWindowEpochs	uint64
	TotalRiskEpochs		uint64
}

type UnbondingRiskRecord struct {
	DelegatorID		string
	ValidatorID		string
	AmountNaet		sdkmath.Int
	RequestedEpoch		uint64
	ExitEpoch		uint64
	SlashableUntilEpoch	uint64
	RiskHistoryKey		string
}

type RedelegationRiskRecord struct {
	DelegatorID			string
	SourceValidatorID		string
	DestinationValidatorID		string
	AmountNaet			sdkmath.Int
	RequestedEpoch			uint64
	ActivationEpoch			uint64
	SourceSlashableUntilEpoch	uint64
	RiskHistoryKey			string
}

type SelfBondChangeRecord struct {
	ValidatorID		string
	PreviousBondNaet	sdkmath.Int
	NewBondNaet		sdkmath.Int
	RequestedEpoch		uint64
	ActivationEpoch		uint64
}

type PendingUnbondingSlashExposureInput struct {
	Record		UnbondingRiskRecord
	FaultEpoch	uint64
	EvidenceEpoch	uint64
}

const (
	RiskWindowStatusActive	= "active"
	RiskWindowStatusExited	= "exited"
	RiskWindowStatusExpired	= "expired"
	RiskWindowStatusSlashed	= "slashed"
)

type RiskWindowRecord struct {
	StakeOwner		string
	ValidatorAddress	string
	AmountNaet		sdkmath.Int
	StartEpoch		uint64
	EndEpoch		uint64
	SlashableUntilEpoch	uint64
	RiskHistoryRoot		string
	Status			string
}

type SlashExposureQuery struct {
	StakeOwner		string
	ValidatorAddress	string
	FaultEpoch		uint64
	EvidenceEpoch		uint64
}

type SlashExposureQueryResult struct {
	StakeOwner		string
	ValidatorAddress	string
	FaultEpoch		uint64
	EvidenceEpoch		uint64
	ExposureNaet		sdkmath.Int
	MatchingWindows		[]RiskWindowRecord
}

type RejectedDelegationIntent struct {
	Intent	DelegationIntent
	Reason	string
}

type EvidenceCase struct {
	EvidenceID		string
	ReporterID		string
	ValidatorID		string
	Misbehavior		string
	SlashFractionBps	uint32
	EvidenceHeight		int64
	EvidenceEpoch		uint64
	Finalized		bool
}

type EvidenceSettlement struct {
	EvidenceID		string
	ReporterID		string
	Slash			SlashDistribution
	ReporterRewardNaet	sdkmath.Int
	BurnNaet		sdkmath.Int
	SettlementHash		string
}

type RoleRewardWeight struct {
	Role		ValidatorRole
	WeightBps	uint32
}

type AssignmentOutcome struct {
	TaskID		string
	Role		ValidatorRole
	ValidatorID	string
	Completed	bool
	Faulted		bool
	WorkUnits	uint64
}

type ValidatorWorkloadReward struct {
	ValidatorID	string
	RewardNaet	sdkmath.Int
	WorkUnits	uint64
}

type WorkloadRewardInput struct {
	EpochID			uint64
	TotalRewardsNaet	sdkmath.Int
	RoleWeights		[]RoleRewardWeight
	Outcomes		[]AssignmentOutcome
}

type WorkloadRewardSettlement struct {
	EpochID		uint64
	Rewards		[]ValidatorWorkloadReward
	RemainderNaet	sdkmath.Int
	RewardRoot	string
	CompletedUnits	uint64
}

type PerformanceFactorInput struct {
	CompletedTasks		uint64
	MissedTasks		uint64
	CorrectVerifications	uint64
	IncorrectVerifications	uint64
	AvailableWindows	uint64
	CommittedWindows	uint64
}

type UptimeFactorInput struct {
	SignedBlocks			uint64
	TotalBlocks			uint64
	TaskParticipations		uint64
	MissedTaskParticipations	uint64
}

type LatencyFactorInput struct {
	CommittedWindow	bool
	AdvisoryOnly	bool
	TargetMillis	uint64
	P95Millis	uint64
}

type ReliabilityIndexInput struct {
	PriorIndexBps		uint32
	SlashEvents		uint64
	DowntimeEpochs		uint64
	MissedTasks		uint64
	RejectedEvidence	uint64
	RecoveryEpochs		uint64
}

type CorrectnessScoreInput struct {
	ValidSignatures		uint64
	InvalidSignatures	uint64
	ValidTaskOutputs	uint64
	InvalidTaskOutputs	uint64
	AcceptedEvidence	uint64
	EvidencePenaltyWeight	uint64
}

type TaskCompletionRateInput struct {
	CompletedAssignedTasks	uint64
	ExpectedAssignedTasks	uint64
}

type PerformanceRewardInput struct {
	EpochID			uint64
	ValidatorID		string
	BaseEmissionNaet	sdkmath.Int
	UptimeScoreBps		uint32
	LatencyScoreBps		uint32
	CorrectnessScoreBps	uint32
	TaskCompletionRateBps	uint32
}

type PerformanceRewardRecord struct {
	EpochID			uint64
	ValidatorID		string
	BaseEmissionNaet	sdkmath.Int
	UptimeScoreBps		uint32
	LatencyScoreBps		uint32
	CorrectnessScoreBps	uint32
	TaskCompletionRateBps	uint32
	RewardNaet		sdkmath.Int
	RewardHash		string
}

type PerformanceRecord struct {
	EpochID			uint64
	OperatorAddress		string
	Role			ValidatorRole
	AssignedTasks		uint64
	CompletedTasks		uint64
	MissedTasks		uint64
	InvalidTasks		uint64
	UptimeScoreBps		uint32
	LatencyScoreBps		uint32
	CorrectnessScoreBps	uint32
	TaskCompletionRateBps	uint32
	RewardMultiplierBps	uint32
}

type PerformanceRecordInput struct {
	EpochID			uint64
	OperatorAddress		string
	Role			ValidatorRole
	AssignedTasks		uint64
	CompletedTasks		uint64
	MissedTasks		uint64
	InvalidTasks		uint64
	UptimeScoreBps		uint32
	LatencyScoreBps		uint32
	CorrectnessScoreBps	uint32
}

type PerformanceDampeningInput struct {
	Record					PerformanceRecord
	CurrentRewardNaet			sdkmath.Int
	FutureElectionScoreBps			uint32
	DelegationAttractivenessBps		uint32
	RoleEligibilityBps			uint32
	CollatorAssignmentProbabilityBps	uint32
}

type PerformanceDampeningResult struct {
	EpochID					uint64
	OperatorAddress				string
	Role					ValidatorRole
	RewardMultiplierBps			uint32
	CurrentRewardNaet			sdkmath.Int
	FutureElectionScoreBps			uint32
	DelegationAttractivenessBps		uint32
	RoleEligibilityBps			uint32
	CollatorAssignmentProbabilityBps	uint32
}

type EconomicSecurityInput struct {
	Validators		[]ScoredValidator
	RiskWindows		[]RiskWindowRecord
	StakeAtRiskNaet		sdkmath.Int
	TopN			uint32
	ParticipatingValidators	uint64
	EligibleValidators	uint64
	AcceptedSlashEvents	uint64
	DetectedFaultEvents	uint64
	AcceptedEvidence	uint64
	SubmittedEvidence	uint64
	CompletedTasks		uint64
	ExpectedTasks		uint64
}

type DelegationRiskBucket struct {
	ValidatorAddress	string
	ExposureNaet		sdkmath.Int
	RiskWindowCount		uint64
}

type EconomicSecurityMetrics struct {
	TotalBondedStakeNaet		sdkmath.Int
	EffectiveStakeNaet		sdkmath.Int
	TotalStakeAtRiskNaet		sdkmath.Int
	StakeSaturationRatioBps		uint32
	TopN				uint32
	TopNVotingPowerConcentrationBps	uint32
	ParticipationRateBps		uint32
	SlashingEfficiencyBps		uint32
	EvidenceAcceptanceRateBps	uint32
	AverageValidatorScore		sdkmath.Int
	DelegationRiskDistribution	[]DelegationRiskBucket
	TaskCompletionRateBps		uint32
	SecurityNaet			sdkmath.Int
}

type SecurityMetricQuery struct {
	Input EconomicSecurityInput
}

type SecurityMetricQueryResult struct {
	Metrics EconomicSecurityMetrics
}

type CentralizationControlParams struct {
	MaxValidatorShareBps		uint32
	MaxTopNConcentrationBps		uint32
	MaxStakeSaturationRatioBps	uint32
	MaxDelegationRiskBucketBps	uint32
	MinBootstrapPerformanceBps	uint32
	MinBootstrapReliabilityBps	uint32
	MaxTaskAssignmentShareBps	uint32
	BootstrapMaxVotingPowerShareBps	uint32
}

type CentralizationTaskAssignment struct {
	TaskGroupID		string
	ValidatorAddress	string
	AssignmentCount		uint64
}

type CentralizationValidatorControl struct {
	ValidatorAddress	string
	VotingPowerShareBps	uint32
	EffectiveStakeNaet	sdkmath.Int
	SaturatedStakeNaet	sdkmath.Int
	RewardDampeningBps	uint32
	BootstrapEligible	bool
	Warnings		[]string
}

type DelegationRiskWarning struct {
	ValidatorAddress	string
	ExposureNaet		sdkmath.Int
	ExposureShareBps	uint32
	ThresholdBps		uint32
}

type TaskAssignmentDiversityReport struct {
	TotalAssignments	uint64
	MaxValidatorAssignments	uint64
	MaxValidatorAddress	string
	MaxAssignmentShareBps	uint32
	DiversityScoreBps	uint32
	Warnings		[]string
}

type ConcentrationInvariantAlert struct {
	AlertType	string
	Severity	string
	ObservedBps	uint32
	ThresholdBps	uint32
}

type CentralizationDashboardInput struct {
	SecurityInput	EconomicSecurityInput
	ControlParams	CentralizationControlParams
	TaskAssignments	[]CentralizationTaskAssignment
}

type CentralizationDashboardData struct {
	Metrics			EconomicSecurityMetrics
	ValidatorControls	[]CentralizationValidatorControl
	DelegationRiskWarnings	[]DelegationRiskWarning
	TaskAssignmentDiversity	TaskAssignmentDiversityReport
	Alerts			[]ConcentrationInvariantAlert
}

type StakeConcentrationSimulationInput struct {
	Params			Params
	Candidates		[]Candidate
	TargetValidatorID	string
	AddedDelegatedStakeNaet	sdkmath.Int
	TopN			uint32
}

type StakeConcentrationSimulationResult struct {
	Before				EconomicSecurityMetrics
	After				EconomicSecurityMetrics
	TopNConcentrationDeltaBps	int32
	TargetEffectiveStakeDeltaNaet	sdkmath.Int
	Alerts				[]ConcentrationInvariantAlert
}

type StakeSplittingSimulationInput struct {
	Params		Params
	Candidate	Candidate
	SplitCount	uint32
	TopN		uint32
}

type StakeSplittingSimulationResult struct {
	SingleEffectiveStakeNaet	sdkmath.Int
	SplitEffectiveStakeNaet		sdkmath.Int
	EffectiveStakeGainNaet		sdkmath.Int
	SingleConcentrationBps		uint32
	SplitConcentrationBps		uint32
}

type CosmosSDKExtensionMode string

const (
	CosmosSDKExtensionModeExtend	CosmosSDKExtensionMode	= "extend"
	CosmosSDKExtensionModeReplace	CosmosSDKExtensionMode	= "replace"
)

type CosmosSDKModuleExtension struct {
	ModuleName		string
	ModulePath		string
	ExtensionMode		CosmosSDKExtensionMode
	PreservedInterfaces	[]string
	AddedState		[]string
	RewardInputs		[]string
}

type PosModuleRequirement struct {
	ModuleName	string
	ModulePath	string
	Required	bool
}

type PosCompatibilityMiddleware struct {
	Name		string
	Layer		PosLayer
	Extends		[]string
	ReadsModules	[]string
	WritesModules	[]string
}

type CosmosSDKCompatibilityManifest struct {
	Extensions	[]CosmosSDKModuleExtension
	Modules		[]PosModuleRequirement
	Middleware	[]PosCompatibilityMiddleware
	Root		string
}

type PosModuleBoundary struct {
	ModuleName	string
	ModulePath	string
	Owns		[]string
	ReadsModules	[]string
	WritesModules	[]string
	QueryEndpoints	[]string
}

type PosModuleBoundaryManifest struct {
	Boundaries	[]PosModuleBoundary
	Root		string
}

type KeeperInterfaceSpec struct {
	KeeperName		string
	ModuleName		string
	InterfaceName		string
	IntegrationPoint	string
	Reads			[]string
	Writes			[]string
}

type KeeperHookSpec struct {
	SourceKeeper		string
	HookName		string
	Trigger			string
	TargetModules		[]string
	PreservesBaseState	bool
	DeterministicOrder	bool
}

type RewardMultiplierIntegration struct {
	SourceModule		string
	DistributionKeeper	string
	MintKeeper		string
	MultiplierField		string
	RewardInputs		[]string
}

type MigrationHandlerSpec struct {
	ModuleName			string
	FromVersion			uint64
	ToVersion			uint64
	PreservesExistingStakingState	bool
	ExportsGenesis			bool
	ImportsGenesis			bool
}

type ModuleExportImportSpec struct {
	ModuleName		string
	ExportsGenesis		bool
	ImportsGenesis		bool
	DeterministicEncoding	bool
}

type KeeperIntegrationManifest struct {
	KeeperInterfaces	[]KeeperInterfaceSpec
	StakingLifecycleHooks	[]KeeperHookSpec
	SlashingHooks		[]KeeperHookSpec
	RewardIntegrations	[]RewardMultiplierIntegration
	MigrationHandlers	[]MigrationHandlerSpec
	ExportImport		[]ModuleExportImportSpec
	Root			string
}

type StateKeySpec struct {
	Domain		string
	Name		string
	Template	string
	Components	[]string
}

type StateModelManifest struct {
	Keys	[]StateKeySpec
	Root	string
}

func (p CentralizationControlParams) Validate() error {
	checks := []struct {
		name	string
		value	uint32
	}{
		{name: "max_validator_share_bps", value: p.MaxValidatorShareBps},
		{name: "max_top_n_concentration_bps", value: p.MaxTopNConcentrationBps},
		{name: "max_stake_saturation_ratio_bps", value: p.MaxStakeSaturationRatioBps},
		{name: "max_delegation_risk_bucket_bps", value: p.MaxDelegationRiskBucketBps},
		{name: "min_bootstrap_performance_bps", value: p.MinBootstrapPerformanceBps},
		{name: "min_bootstrap_reliability_bps", value: p.MinBootstrapReliabilityBps},
		{name: "max_task_assignment_share_bps", value: p.MaxTaskAssignmentShareBps},
		{name: "bootstrap_max_voting_power_share_bps", value: p.BootstrapMaxVotingPowerShareBps},
	}
	for _, check := range checks {
		if check.value == 0 || check.value > BasisPoints {
			return fmt.Errorf("%s must be within 1..%d bps", check.name, BasisPoints)
		}
	}
	return nil
}

type PosLayer string

const (
	PosLayerEconomicConsensus	PosLayer	= "economic_consensus"
	PosLayerTaskAssignment		PosLayer	= "task_assignment"
	PosLayerValidatorExecution	PosLayer	= "validator_execution"
	PosLayerStakingCapital		PosLayer	= "staking_capital"
	PosLayerBaseCometBFT		PosLayer	= "base_cometbft"
)

type PosLayerSpec struct {
	Layer			PosLayer
	Responsibilities	[]string
	DependsOn		[]PosLayer
}

type LayeredPosArchitecture struct {
	Layers	[]PosLayerSpec
	Root	string
}

func DefaultLayeredPosArchitecture() LayeredPosArchitecture {
	layers := []PosLayerSpec{
		{
			Layer:	PosLayerEconomicConsensus,
			Responsibilities: []string{
				"validator scoring",
				"performance incentives",
				"stake saturation",
				"role-specific reward weights",
				"slashing severity",
				"reporter incentives",
				"treasury, burn, and stabilization routing",
			},
			DependsOn:	[]PosLayer{PosLayerTaskAssignment, PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer:	PosLayerTaskAssignment,
			Responsibilities: []string{
				"workload grouping",
				"shard validator groups",
				"zone validator groups",
				"evidence verification subsets",
				"collator and verifier assignments",
			},
			DependsOn:	[]PosLayer{PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer:	PosLayerValidatorExecution,
			Responsibilities: []string{
				"block production",
				"state transition verification",
				"cross-domain proof verification",
				"signature production",
				"fault rejection",
			},
			DependsOn:	[]PosLayer{PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer:	PosLayerStakingCapital,
			Responsibilities: []string{
				"validators",
				"delegators",
				"bonded stake",
				"unbonding",
				"redelegation",
				"capital risk preferences",
				"commission and delegation market metadata",
			},
			DependsOn:	[]PosLayer{PosLayerBaseCometBFT},
		},
		{
			Layer:	PosLayerBaseCometBFT,
			Responsibilities: []string{
				"finality",
				"proposal and vote protocol",
				"validator public key set",
				"consensus safety and liveness",
			},
		},
	}
	architecture := LayeredPosArchitecture{Layers: layers}
	architecture.Root = ComputeLayeredPosArchitectureRoot(layers)
	return architecture
}

func DefaultCosmosSDKCompatibilityManifest() CosmosSDKCompatibilityManifest {
	manifest := CosmosSDKCompatibilityManifest{
		Extensions: []CosmosSDKModuleExtension{
			{
				ModuleName:		"staking",
				ModulePath:		"x/staking",
				ExtensionMode:		CosmosSDKExtensionModeExtend,
				PreservedInterfaces:	[]string{"ValidatorI", "Delegation", "Redelegation", "UnbondingDelegation", "staking keeper hooks"},
				AddedState:		[]string{"delegation activation epoch", "validator score references", "risk window references", "capacity declarations"},
			},
			{
				ModuleName:		"slashing",
				ModulePath:		"x/slashing",
				ExtensionMode:		CosmosSDKExtensionModeExtend,
				PreservedInterfaces:	[]string{"ValidatorSigningInfo", "tombstone", "jail", "missed block bitmap"},
				AddedState:		[]string{"severity matrix", "role suspension", "future election score penalty", "delegator slash exposure"},
			},
			{
				ModuleName:		"distribution",
				ModulePath:		"x/distribution",
				ExtensionMode:		CosmosSDKExtensionModeExtend,
				PreservedInterfaces:	[]string{"delegator rewards", "validator outstanding rewards", "fee pool"},
				AddedState:		[]string{"role reward weights", "performance reward multiplier", "reporter reward routing"},
				RewardInputs:		[]string{"uptime score", "correctness score", "task completion rate", "role weight"},
			},
			{
				ModuleName:		"mint",
				ModulePath:		"x/mint",
				ExtensionMode:		CosmosSDKExtensionModeExtend,
				PreservedInterfaces:	[]string{"mint params", "minter", "fee collector emission"},
				AddedState:		[]string{"epoch reward budget", "workload-aware emission inputs", "security metric feedback"},
				RewardInputs:		[]string{"base emission", "participation rate", "security score", "performance budget"},
			},
		},
		Modules: []PosModuleRequirement{
			{ModuleName: "epoch", ModulePath: "x/epoch", Required: true},
			{ModuleName: "validator_economy", ModulePath: "x/validator-economy", Required: true},
			{ModuleName: "taskgroups", ModulePath: "x/taskgroups", Required: true},
			{ModuleName: "evidence", ModulePath: "x/evidence", Required: true},
			{ModuleName: "performance", ModulePath: "x/performance", Required: true},
			{ModuleName: "delegation_market", ModulePath: "x/delegation-market", Required: false},
			{ModuleName: "collators", ModulePath: "x/collators", Required: false},
			{ModuleName: "fishermen", ModulePath: "x/fishermen", Required: false},
			{ModuleName: "security_metrics", ModulePath: "x/security-metrics", Required: false},
		},
		Middleware: []PosCompatibilityMiddleware{
			{Name: "epoch_management", Layer: PosLayerStakingCapital, Extends: []string{"staking", "slashing"}, ReadsModules: []string{"epoch", "staking"}, WritesModules: []string{"epoch"}},
			{Name: "validator_scoring", Layer: PosLayerEconomicConsensus, Extends: []string{"staking"}, ReadsModules: []string{"staking", "slashing", "performance", "validator_economy"}, WritesModules: []string{"validator_economy"}},
			{Name: "task_assignment", Layer: PosLayerTaskAssignment, Extends: []string{"staking"}, ReadsModules: []string{"epoch", "validator_economy", "staking"}, WritesModules: []string{"taskgroups"}},
			{Name: "performance_accounting", Layer: PosLayerEconomicConsensus, Extends: []string{"distribution", "mint"}, ReadsModules: []string{"performance", "taskgroups", "staking"}, WritesModules: []string{"performance", "distribution"}},
			{Name: "evidence_slashing", Layer: PosLayerEconomicConsensus, Extends: []string{"slashing", "distribution"}, ReadsModules: []string{"evidence", "taskgroups", "staking"}, WritesModules: []string{"evidence", "slashing", "distribution"}},
		},
	}
	manifest.Root = ComputeCosmosSDKCompatibilityRoot(manifest)
	return manifest
}

func RequiredPoSModuleNames(manifest CosmosSDKCompatibilityManifest) []string {
	out := make([]string, 0)
	for _, module := range manifest.Modules {
		if module.Required {
			out = append(out, module.ModuleName)
		}
	}
	return out
}

func OptionalPoSModuleNames(manifest CosmosSDKCompatibilityManifest) []string {
	out := make([]string, 0)
	for _, module := range manifest.Modules {
		if !module.Required {
			out = append(out, module.ModuleName)
		}
	}
	return out
}

func (m CosmosSDKCompatibilityManifest) Validate() error {
	if len(m.Extensions) == 0 {
		return errors.New("cosmos sdk compatibility extensions are required")
	}
	if len(m.Modules) == 0 {
		return errors.New("pos compatibility modules are required")
	}
	if len(m.Middleware) == 0 {
		return errors.New("pos compatibility middleware is required")
	}
	extensionByName := make(map[string]struct{}, len(m.Extensions))
	for _, extension := range m.Extensions {
		if err := extension.Validate(); err != nil {
			return err
		}
		if extension.ExtensionMode != CosmosSDKExtensionModeExtend {
			return fmt.Errorf("cosmos sdk module %s must be extended, not replaced", extension.ModuleName)
		}
		if _, found := extensionByName[extension.ModuleName]; found {
			return fmt.Errorf("duplicate cosmos sdk extension %s", extension.ModuleName)
		}
		extensionByName[extension.ModuleName] = struct{}{}
	}
	for _, required := range []string{"staking", "slashing", "distribution", "mint"} {
		if _, found := extensionByName[required]; !found {
			return fmt.Errorf("required cosmos sdk extension %s is missing", required)
		}
	}
	moduleByName := make(map[string]PosModuleRequirement, len(m.Modules))
	for _, module := range m.Modules {
		if err := module.Validate(); err != nil {
			return err
		}
		if _, found := moduleByName[module.ModuleName]; found {
			return fmt.Errorf("duplicate pos compatibility module %s", module.ModuleName)
		}
		moduleByName[module.ModuleName] = module
	}
	for _, required := range []string{"epoch", "validator_economy", "taskgroups", "evidence", "performance"} {
		module, found := moduleByName[required]
		if !found || !module.Required {
			return fmt.Errorf("required pos module %s is missing", required)
		}
	}
	for _, optional := range []string{"delegation_market", "collators", "fishermen", "security_metrics"} {
		module, found := moduleByName[optional]
		if !found || module.Required {
			return fmt.Errorf("optional pos module %s is missing or marked required", optional)
		}
	}
	for _, middleware := range m.Middleware {
		if err := middleware.Validate(extensionByName, moduleByName); err != nil {
			return err
		}
	}
	if err := validatePosHash("cosmos sdk compatibility root", m.Root); err != nil {
		return err
	}
	if expected := ComputeCosmosSDKCompatibilityRoot(m); expected != m.Root {
		return errors.New("cosmos sdk compatibility root mismatch")
	}
	return nil
}

func (e CosmosSDKModuleExtension) Validate() error {
	if err := validatePosToken("cosmos sdk extension module name", e.ModuleName); err != nil {
		return err
	}
	if err := validatePosToken("cosmos sdk extension module path", e.ModulePath); err != nil {
		return err
	}
	if e.ExtensionMode != CosmosSDKExtensionModeExtend && e.ExtensionMode != CosmosSDKExtensionModeReplace {
		return fmt.Errorf("unsupported cosmos sdk extension mode %s", e.ExtensionMode)
	}
	if len(e.PreservedInterfaces) == 0 {
		return fmt.Errorf("cosmos sdk extension %s must preserve baseline interfaces", e.ModuleName)
	}
	if len(e.AddedState) == 0 && len(e.RewardInputs) == 0 {
		return fmt.Errorf("cosmos sdk extension %s must add state or reward inputs", e.ModuleName)
	}
	for _, value := range e.PreservedInterfaces {
		if err := validatePosResponsibility("cosmos sdk preserved interface", value); err != nil {
			return err
		}
	}
	for _, value := range e.AddedState {
		if err := validatePosResponsibility("cosmos sdk added state", value); err != nil {
			return err
		}
	}
	for _, value := range e.RewardInputs {
		if err := validatePosResponsibility("cosmos sdk reward input", value); err != nil {
			return err
		}
	}
	return nil
}

func (r PosModuleRequirement) Validate() error {
	if err := validatePosToken("pos compatibility module name", r.ModuleName); err != nil {
		return err
	}
	return validatePosToken("pos compatibility module path", r.ModulePath)
}

func (m PosCompatibilityMiddleware) Validate(extensions map[string]struct{}, modules map[string]PosModuleRequirement) error {
	if err := validatePosToken("pos compatibility middleware name", m.Name); err != nil {
		return err
	}
	if err := validatePosLayer(m.Layer); err != nil {
		return err
	}
	if len(m.Extends) == 0 {
		return fmt.Errorf("pos compatibility middleware %s must extend at least one sdk module", m.Name)
	}
	for _, extension := range m.Extends {
		if err := validatePosToken("pos compatibility middleware extension", extension); err != nil {
			return err
		}
		if _, found := extensions[extension]; !found {
			return fmt.Errorf("pos compatibility middleware %s extends unknown sdk module %s", m.Name, extension)
		}
	}
	if len(m.ReadsModules) == 0 {
		return fmt.Errorf("pos compatibility middleware %s must read at least one module", m.Name)
	}
	referencedModules := append([]string{}, m.ReadsModules...)
	referencedModules = append(referencedModules, m.WritesModules...)
	for _, module := range referencedModules {
		if err := validatePosToken("pos compatibility middleware module", module); err != nil {
			return err
		}
		if _, sdkFound := extensions[module]; sdkFound {
			continue
		}
		if _, posFound := modules[module]; !posFound {
			return fmt.Errorf("pos compatibility middleware %s references unknown module %s", m.Name, module)
		}
	}
	return nil
}

func ComputeCosmosSDKCompatibilityRoot(manifest CosmosSDKCompatibilityManifest) string {
	return posHashRoot("aetheris-pos-cosmos-sdk-compatibility-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(manifest.Extensions)))
		for _, extension := range manifest.Extensions {
			posWritePart(w, extension.ModuleName)
			posWritePart(w, extension.ModulePath)
			posWritePart(w, string(extension.ExtensionMode))
			posWriteStringSlice(w, extension.PreservedInterfaces)
			posWriteStringSlice(w, extension.AddedState)
			posWriteStringSlice(w, extension.RewardInputs)
		}
		posWriteUint64(w, uint64(len(manifest.Modules)))
		for _, module := range manifest.Modules {
			posWritePart(w, module.ModuleName)
			posWritePart(w, module.ModulePath)
			posWriteUint64(w, boolAsUint64(module.Required))
		}
		posWriteUint64(w, uint64(len(manifest.Middleware)))
		for _, middleware := range manifest.Middleware {
			posWritePart(w, middleware.Name)
			posWritePart(w, string(middleware.Layer))
			posWriteStringSlice(w, middleware.Extends)
			posWriteStringSlice(w, middleware.ReadsModules)
			posWriteStringSlice(w, middleware.WritesModules)
		}
	})
}

func DefaultPoSModuleBoundaryManifest() PosModuleBoundaryManifest {
	manifest := PosModuleBoundaryManifest{
		Boundaries: []PosModuleBoundary{
			{
				ModuleName:	"epoch",
				ModulePath:	"x/epoch",
				Owns:		[]string{"epoch lifecycle", "phase transitions", "epoch seed", "epoch queries"},
				ReadsModules:	[]string{"staking"},
				WritesModules:	[]string{"epoch"},
				QueryEndpoints:	[]string{"QueryCurrentEpoch", "QueryEpochHistory"},
			},
			{
				ModuleName:	"validator_economy",
				ModulePath:	"x/validator-economy",
				Owns:		[]string{"validator score", "effective stake", "stake saturation", "election ranking", "role eligibility"},
				ReadsModules:	[]string{"staking", "slashing", "performance"},
				WritesModules:	[]string{"validator_economy"},
				QueryEndpoints:	[]string{"QueryValidatorScore", "QueryElectionRanking", "QueryValidatorSaturation", "QueryRoleEligibility"},
			},
			{
				ModuleName:	"taskgroups",
				ModulePath:	"x/taskgroups",
				Owns:		[]string{"workload registry", "task group assignment", "proposer rotation", "verification groups"},
				ReadsModules:	[]string{"epoch", "validator_economy", "staking"},
				WritesModules:	[]string{"taskgroups"},
				QueryEndpoints:	[]string{"QueryWorkloadRegistry", "QueryTaskGroup", "QueryProposerRotation", "QueryVerificationGroup"},
			},
			{
				ModuleName:	"evidence",
				ModulePath:	"x/evidence",
				Owns:		[]string{"structured evidence records", "evidence deposits", "verification group decisions", "reporter rewards"},
				ReadsModules:	[]string{"taskgroups", "staking", "slashing"},
				WritesModules:	[]string{"evidence", "slashing", "distribution"},
				QueryEndpoints:	[]string{"QueryEvidenceRecord", "QueryEvidenceDeposit", "QueryEvidenceDecision", "QueryReporterRewards"},
			},
			{
				ModuleName:	"performance",
				ModulePath:	"x/performance",
				Owns:		[]string{"uptime", "latency", "correctness", "task completion", "reward multipliers"},
				ReadsModules:	[]string{"taskgroups", "staking", "distribution"},
				WritesModules:	[]string{"performance", "distribution"},
				QueryEndpoints:	[]string{"QueryPerformanceRecord", "QueryOperatorPerformanceHistory", "QueryRolePerformance", "QueryRewardMultiplier"},
			},
		},
	}
	manifest.Root = ComputePoSModuleBoundaryRoot(manifest)
	return manifest
}

func (m PosModuleBoundaryManifest) Validate(compatibility CosmosSDKCompatibilityManifest) error {
	if err := compatibility.Validate(); err != nil {
		return err
	}
	if len(m.Boundaries) == 0 {
		return errors.New("pos module boundaries are required")
	}
	knownModules := make(map[string]struct{})
	for _, extension := range compatibility.Extensions {
		knownModules[extension.ModuleName] = struct{}{}
	}
	for _, module := range compatibility.Modules {
		knownModules[module.ModuleName] = struct{}{}
	}
	required := RequiredPoSModuleNames(compatibility)
	boundaryByName := make(map[string]PosModuleBoundary, len(m.Boundaries))
	owned := make(map[string]string)
	for _, boundary := range m.Boundaries {
		if err := boundary.Validate(knownModules); err != nil {
			return err
		}
		if _, found := boundaryByName[boundary.ModuleName]; found {
			return fmt.Errorf("duplicate pos module boundary %s", boundary.ModuleName)
		}
		boundaryByName[boundary.ModuleName] = boundary
		for _, item := range boundary.Owns {
			if owner, found := owned[item]; found {
				return fmt.Errorf("pos boundary ownership %q overlaps between %s and %s", item, owner, boundary.ModuleName)
			}
			owned[item] = boundary.ModuleName
		}
	}
	for _, moduleName := range required {
		if _, found := boundaryByName[moduleName]; !found {
			return fmt.Errorf("required pos module boundary %s is missing", moduleName)
		}
	}
	if err := validatePosHash("pos module boundary root", m.Root); err != nil {
		return err
	}
	if expected := ComputePoSModuleBoundaryRoot(m); expected != m.Root {
		return errors.New("pos module boundary root mismatch")
	}
	return nil
}

func (b PosModuleBoundary) Validate(knownModules map[string]struct{}) error {
	if err := validatePosToken("pos module boundary name", b.ModuleName); err != nil {
		return err
	}
	if err := validatePosToken("pos module boundary path", b.ModulePath); err != nil {
		return err
	}
	if len(b.Owns) == 0 {
		return fmt.Errorf("pos module boundary %s must own at least one responsibility", b.ModuleName)
	}
	for _, item := range b.Owns {
		if err := validatePosResponsibility("pos module boundary ownership", item); err != nil {
			return err
		}
	}
	if len(b.QueryEndpoints) == 0 {
		return fmt.Errorf("pos module boundary %s must expose query endpoints", b.ModuleName)
	}
	for _, endpoint := range b.QueryEndpoints {
		if err := validatePosToken("pos module boundary query endpoint", endpoint); err != nil {
			return err
		}
	}
	referenced := append([]string{}, b.ReadsModules...)
	referenced = append(referenced, b.WritesModules...)
	for _, moduleName := range referenced {
		if err := validatePosToken("pos module boundary referenced module", moduleName); err != nil {
			return err
		}
		if _, found := knownModules[moduleName]; !found {
			return fmt.Errorf("pos module boundary %s references unknown module %s", b.ModuleName, moduleName)
		}
	}
	return nil
}

func PoSModuleBoundaryByName(manifest PosModuleBoundaryManifest, moduleName string) (PosModuleBoundary, bool) {
	for _, boundary := range manifest.Boundaries {
		if boundary.ModuleName == moduleName {
			return boundary, true
		}
	}
	return PosModuleBoundary{}, false
}

func ComputePoSModuleBoundaryRoot(manifest PosModuleBoundaryManifest) string {
	return posHashRoot("aetheris-pos-module-boundaries-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(manifest.Boundaries)))
		for _, boundary := range manifest.Boundaries {
			posWritePart(w, boundary.ModuleName)
			posWritePart(w, boundary.ModulePath)
			posWriteStringSlice(w, boundary.Owns)
			posWriteStringSlice(w, boundary.ReadsModules)
			posWriteStringSlice(w, boundary.WritesModules)
			posWriteStringSlice(w, boundary.QueryEndpoints)
		}
	})
}

func DefaultKeeperIntegrationManifest() KeeperIntegrationManifest {
	manifest := KeeperIntegrationManifest{
		KeeperInterfaces: []KeeperInterfaceSpec{
			{KeeperName: "staking", ModuleName: "staking", InterfaceName: "StakingKeeper", IntegrationPoint: "validator and delegation state", Reads: []string{"validators", "delegations", "redelegations", "unbonding delegations"}, Writes: []string{"staking hooks only"}},
			{KeeperName: "slashing", ModuleName: "slashing", InterfaceName: "SlashingKeeper", IntegrationPoint: "jail tombstone and slash execution", Reads: []string{"validator signing info", "missed block bitmap"}, Writes: []string{"jail", "tombstone", "slash execution"}},
			{KeeperName: "distribution", ModuleName: "distribution", InterfaceName: "DistributionKeeper", IntegrationPoint: "reward allocation", Reads: []string{"fee pool", "outstanding rewards"}, Writes: []string{"validator rewards", "delegator rewards", "reporter rewards"}},
			{KeeperName: "mint", ModuleName: "mint", InterfaceName: "MintKeeper", IntegrationPoint: "epoch reward budget", Reads: []string{"minter", "mint params"}, Writes: []string{"epoch reward budget"}},
			{KeeperName: "bank", ModuleName: "bank", InterfaceName: "BankKeeper", IntegrationPoint: "deposits reporter rewards and penalty routing", Reads: []string{"module balances", "account balances"}, Writes: []string{"evidence deposits", "reporter rewards", "penalty routing"}},
			{KeeperName: "gov", ModuleName: "gov", InterfaceName: "GovernanceKeeper", IntegrationPoint: "parameter updates", Reads: []string{"governance authority", "parameter proposals"}, Writes: []string{"pos params", "economy params", "security params"}},
		},
		StakingLifecycleHooks: []KeeperHookSpec{
			{SourceKeeper: "staking", HookName: "AfterValidatorCreated", Trigger: "validator registration", TargetModules: []string{"epoch", "validator_economy"}, PreservesBaseState: true, DeterministicOrder: true},
			{SourceKeeper: "staking", HookName: "AfterValidatorBonded", Trigger: "validator bonded", TargetModules: []string{"epoch", "validator_economy", "taskgroups"}, PreservesBaseState: true, DeterministicOrder: true},
			{SourceKeeper: "staking", HookName: "AfterDelegationModified", Trigger: "delegation modified", TargetModules: []string{"epoch", "validator_economy"}, PreservesBaseState: true, DeterministicOrder: true},
			{SourceKeeper: "staking", HookName: "BeforeDelegationRemoved", Trigger: "delegation exit", TargetModules: []string{"epoch", "validator_economy"}, PreservesBaseState: true, DeterministicOrder: true},
		},
		SlashingHooks: []KeeperHookSpec{
			{SourceKeeper: "slashing", HookName: "AfterValidatorSlashed", Trigger: "slash execution", TargetModules: []string{"performance", "validator_economy"}, PreservesBaseState: true, DeterministicOrder: true},
			{SourceKeeper: "slashing", HookName: "AfterValidatorJailed", Trigger: "validator jail", TargetModules: []string{"performance", "validator_economy", "taskgroups"}, PreservesBaseState: true, DeterministicOrder: true},
			{SourceKeeper: "slashing", HookName: "AfterValidatorTombstoned", Trigger: "validator tombstone", TargetModules: []string{"performance", "validator_economy", "taskgroups"}, PreservesBaseState: true, DeterministicOrder: true},
		},
		RewardIntegrations: []RewardMultiplierIntegration{
			{SourceModule: "performance", DistributionKeeper: "distribution", MintKeeper: "mint", MultiplierField: "reward_multiplier_bps", RewardInputs: []string{"uptime", "latency", "correctness", "task completion"}},
		},
		MigrationHandlers: []MigrationHandlerSpec{
			{ModuleName: "epoch", FromVersion: 1, ToVersion: 2, PreservesExistingStakingState: true, ExportsGenesis: true, ImportsGenesis: true},
			{ModuleName: "validator_economy", FromVersion: 1, ToVersion: 2, PreservesExistingStakingState: true, ExportsGenesis: true, ImportsGenesis: true},
			{ModuleName: "taskgroups", FromVersion: 1, ToVersion: 2, PreservesExistingStakingState: true, ExportsGenesis: true, ImportsGenesis: true},
			{ModuleName: "evidence", FromVersion: 1, ToVersion: 2, PreservesExistingStakingState: true, ExportsGenesis: true, ImportsGenesis: true},
			{ModuleName: "performance", FromVersion: 1, ToVersion: 2, PreservesExistingStakingState: true, ExportsGenesis: true, ImportsGenesis: true},
		},
		ExportImport: []ModuleExportImportSpec{
			{ModuleName: "epoch", ExportsGenesis: true, ImportsGenesis: true, DeterministicEncoding: true},
			{ModuleName: "validator_economy", ExportsGenesis: true, ImportsGenesis: true, DeterministicEncoding: true},
			{ModuleName: "taskgroups", ExportsGenesis: true, ImportsGenesis: true, DeterministicEncoding: true},
			{ModuleName: "evidence", ExportsGenesis: true, ImportsGenesis: true, DeterministicEncoding: true},
			{ModuleName: "performance", ExportsGenesis: true, ImportsGenesis: true, DeterministicEncoding: true},
		},
	}
	manifest.Root = ComputeKeeperIntegrationRoot(manifest)
	return manifest
}

func (m KeeperIntegrationManifest) Validate(compatibility CosmosSDKCompatibilityManifest, boundaries PosModuleBoundaryManifest) error {
	if err := compatibility.Validate(); err != nil {
		return err
	}
	if err := boundaries.Validate(compatibility); err != nil {
		return err
	}
	knownModules := knownKeeperIntegrationModules(compatibility)
	if len(m.KeeperInterfaces) == 0 {
		return errors.New("keeper interfaces are required")
	}
	keepers := make(map[string]KeeperInterfaceSpec, len(m.KeeperInterfaces))
	for _, keeper := range m.KeeperInterfaces {
		if err := keeper.Validate(knownModules); err != nil {
			return err
		}
		if _, found := keepers[keeper.KeeperName]; found {
			return fmt.Errorf("duplicate keeper interface %s", keeper.KeeperName)
		}
		keepers[keeper.KeeperName] = keeper
	}
	for _, required := range []string{"staking", "slashing", "distribution", "mint", "bank", "gov"} {
		if _, found := keepers[required]; !found {
			return fmt.Errorf("required keeper interface %s is missing", required)
		}
	}
	if err := validateHookSet("staking lifecycle", m.StakingLifecycleHooks, "staking", knownModules); err != nil {
		return err
	}
	if err := validateHookSet("slashing", m.SlashingHooks, "slashing", knownModules); err != nil {
		return err
	}
	if len(m.RewardIntegrations) == 0 {
		return errors.New("distribution reward multiplier integration is required")
	}
	for _, integration := range m.RewardIntegrations {
		if err := integration.Validate(knownModules); err != nil {
			return err
		}
	}
	requiredPoS := RequiredPoSModuleNames(compatibility)
	if err := validateMigrationHandlers(m.MigrationHandlers, requiredPoS, knownModules); err != nil {
		return err
	}
	if err := validateExportImportSupport(m.ExportImport, requiredPoS, knownModules); err != nil {
		return err
	}
	if err := validatePosHash("keeper integration root", m.Root); err != nil {
		return err
	}
	if expected := ComputeKeeperIntegrationRoot(m); expected != m.Root {
		return errors.New("keeper integration root mismatch")
	}
	return nil
}

func (s KeeperInterfaceSpec) Validate(knownModules map[string]struct{}) error {
	if err := validatePosToken("keeper name", s.KeeperName); err != nil {
		return err
	}
	if err := validatePosToken("keeper module name", s.ModuleName); err != nil {
		return err
	}
	if _, found := knownModules[s.ModuleName]; !found {
		return fmt.Errorf("keeper %s references unknown module %s", s.KeeperName, s.ModuleName)
	}
	if err := validatePosToken("keeper interface name", s.InterfaceName); err != nil {
		return err
	}
	if err := validatePosResponsibility("keeper integration point", s.IntegrationPoint); err != nil {
		return err
	}
	if len(s.Reads) == 0 && len(s.Writes) == 0 {
		return fmt.Errorf("keeper %s must declare reads or writes", s.KeeperName)
	}
	for _, value := range append(append([]string{}, s.Reads...), s.Writes...) {
		if err := validatePosResponsibility("keeper access", value); err != nil {
			return err
		}
	}
	return nil
}

func (h KeeperHookSpec) Validate(expectedSource string, knownModules map[string]struct{}) error {
	if h.SourceKeeper != expectedSource {
		return fmt.Errorf("%s hook source must be %s", h.HookName, expectedSource)
	}
	if err := validatePosToken("keeper hook source", h.SourceKeeper); err != nil {
		return err
	}
	if err := validatePosToken("keeper hook name", h.HookName); err != nil {
		return err
	}
	if err := validatePosResponsibility("keeper hook trigger", h.Trigger); err != nil {
		return err
	}
	if len(h.TargetModules) == 0 {
		return fmt.Errorf("keeper hook %s must target at least one module", h.HookName)
	}
	for _, moduleName := range h.TargetModules {
		if err := validatePosToken("keeper hook target module", moduleName); err != nil {
			return err
		}
		if _, found := knownModules[moduleName]; !found {
			return fmt.Errorf("keeper hook %s references unknown module %s", h.HookName, moduleName)
		}
	}
	if !h.PreservesBaseState {
		return fmt.Errorf("keeper hook %s must preserve base sdk state", h.HookName)
	}
	if !h.DeterministicOrder {
		return fmt.Errorf("keeper hook %s must use deterministic order", h.HookName)
	}
	return nil
}

func (r RewardMultiplierIntegration) Validate(knownModules map[string]struct{}) error {
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "reward source module", value: r.SourceModule},
		{name: "reward distribution keeper", value: r.DistributionKeeper},
		{name: "reward mint keeper", value: r.MintKeeper},
		{name: "reward multiplier field", value: r.MultiplierField},
	} {
		if err := validatePosToken(item.name, item.value); err != nil {
			return err
		}
	}
	for _, moduleName := range []string{r.SourceModule, r.DistributionKeeper, r.MintKeeper} {
		if _, found := knownModules[moduleName]; !found {
			return fmt.Errorf("reward integration references unknown module %s", moduleName)
		}
	}
	if r.SourceModule != "performance" || r.DistributionKeeper != "distribution" || r.MintKeeper != "mint" {
		return errors.New("reward multiplier integration must connect performance to distribution and mint")
	}
	if len(r.RewardInputs) == 0 {
		return errors.New("reward multiplier integration inputs are required")
	}
	for _, input := range r.RewardInputs {
		if err := validatePosResponsibility("reward multiplier input", input); err != nil {
			return err
		}
	}
	return nil
}

func ComputeKeeperIntegrationRoot(manifest KeeperIntegrationManifest) string {
	return posHashRoot("aetheris-pos-keeper-integration-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(manifest.KeeperInterfaces)))
		for _, keeper := range manifest.KeeperInterfaces {
			posWritePart(w, keeper.KeeperName)
			posWritePart(w, keeper.ModuleName)
			posWritePart(w, keeper.InterfaceName)
			posWritePart(w, keeper.IntegrationPoint)
			posWriteStringSlice(w, keeper.Reads)
			posWriteStringSlice(w, keeper.Writes)
		}
		posWriteHookSpecs(w, manifest.StakingLifecycleHooks)
		posWriteHookSpecs(w, manifest.SlashingHooks)
		posWriteUint64(w, uint64(len(manifest.RewardIntegrations)))
		for _, integration := range manifest.RewardIntegrations {
			posWritePart(w, integration.SourceModule)
			posWritePart(w, integration.DistributionKeeper)
			posWritePart(w, integration.MintKeeper)
			posWritePart(w, integration.MultiplierField)
			posWriteStringSlice(w, integration.RewardInputs)
		}
		posWriteUint64(w, uint64(len(manifest.MigrationHandlers)))
		for _, migration := range manifest.MigrationHandlers {
			posWritePart(w, migration.ModuleName)
			posWriteUint64(w, migration.FromVersion)
			posWriteUint64(w, migration.ToVersion)
			posWriteUint64(w, boolAsUint64(migration.PreservesExistingStakingState))
			posWriteUint64(w, boolAsUint64(migration.ExportsGenesis))
			posWriteUint64(w, boolAsUint64(migration.ImportsGenesis))
		}
		posWriteUint64(w, uint64(len(manifest.ExportImport)))
		for _, spec := range manifest.ExportImport {
			posWritePart(w, spec.ModuleName)
			posWriteUint64(w, boolAsUint64(spec.ExportsGenesis))
			posWriteUint64(w, boolAsUint64(spec.ImportsGenesis))
			posWriteUint64(w, boolAsUint64(spec.DeterministicEncoding))
		}
	})
}

func validateHookSet(label string, hooks []KeeperHookSpec, source string, knownModules map[string]struct{}) error {
	if len(hooks) == 0 {
		return fmt.Errorf("%s hooks are required", label)
	}
	seen := make(map[string]struct{}, len(hooks))
	for _, hook := range hooks {
		if err := hook.Validate(source, knownModules); err != nil {
			return err
		}
		if _, found := seen[hook.HookName]; found {
			return fmt.Errorf("duplicate %s hook %s", label, hook.HookName)
		}
		seen[hook.HookName] = struct{}{}
	}
	return nil
}

func validateMigrationHandlers(handlers []MigrationHandlerSpec, requiredModules []string, knownModules map[string]struct{}) error {
	if len(handlers) == 0 {
		return errors.New("migration handlers are required")
	}
	byModule := make(map[string]MigrationHandlerSpec, len(handlers))
	for _, handler := range handlers {
		if err := handler.Validate(knownModules); err != nil {
			return err
		}
		if _, found := byModule[handler.ModuleName]; found {
			return fmt.Errorf("duplicate migration handler %s", handler.ModuleName)
		}
		byModule[handler.ModuleName] = handler
	}
	for _, moduleName := range requiredModules {
		if _, found := byModule[moduleName]; !found {
			return fmt.Errorf("migration handler for %s is missing", moduleName)
		}
	}
	return nil
}

func (m MigrationHandlerSpec) Validate(knownModules map[string]struct{}) error {
	if err := validatePosToken("migration module name", m.ModuleName); err != nil {
		return err
	}
	if _, found := knownModules[m.ModuleName]; !found {
		return fmt.Errorf("migration references unknown module %s", m.ModuleName)
	}
	if m.FromVersion == 0 || m.ToVersion <= m.FromVersion {
		return fmt.Errorf("migration %s must advance module version", m.ModuleName)
	}
	if !m.PreservesExistingStakingState {
		return fmt.Errorf("migration %s must preserve existing staking state", m.ModuleName)
	}
	if !m.ExportsGenesis || !m.ImportsGenesis {
		return fmt.Errorf("migration %s must preserve export and import support", m.ModuleName)
	}
	return nil
}

func validateExportImportSupport(specs []ModuleExportImportSpec, requiredModules []string, knownModules map[string]struct{}) error {
	if len(specs) == 0 {
		return errors.New("export import support is required")
	}
	byModule := make(map[string]ModuleExportImportSpec, len(specs))
	for _, spec := range specs {
		if err := spec.Validate(knownModules); err != nil {
			return err
		}
		if _, found := byModule[spec.ModuleName]; found {
			return fmt.Errorf("duplicate export import support %s", spec.ModuleName)
		}
		byModule[spec.ModuleName] = spec
	}
	for _, moduleName := range requiredModules {
		if _, found := byModule[moduleName]; !found {
			return fmt.Errorf("export import support for %s is missing", moduleName)
		}
	}
	return nil
}

func (s ModuleExportImportSpec) Validate(knownModules map[string]struct{}) error {
	if err := validatePosToken("export import module name", s.ModuleName); err != nil {
		return err
	}
	if _, found := knownModules[s.ModuleName]; !found {
		return fmt.Errorf("export import references unknown module %s", s.ModuleName)
	}
	if !s.ExportsGenesis || !s.ImportsGenesis {
		return fmt.Errorf("module %s must support export and import", s.ModuleName)
	}
	if !s.DeterministicEncoding {
		return fmt.Errorf("module %s export import encoding must be deterministic", s.ModuleName)
	}
	return nil
}

func knownKeeperIntegrationModules(compatibility CosmosSDKCompatibilityManifest) map[string]struct{} {
	known := make(map[string]struct{})
	for _, extension := range compatibility.Extensions {
		known[extension.ModuleName] = struct{}{}
	}
	for _, module := range compatibility.Modules {
		known[module.ModuleName] = struct{}{}
	}
	known["bank"] = struct{}{}
	known["gov"] = struct{}{}
	return known
}

func DefaultStateModelManifest() StateModelManifest {
	manifest := StateModelManifest{Keys: []StateKeySpec{
		{Domain: "epoch", Name: "current", Template: "epoch/current"},
		{Domain: "epoch", Name: "records", Template: "epoch/records/{epoch_id}", Components: []string{"epoch_id"}},
		{Domain: "epoch", Name: "phase", Template: "epoch/phase/{epoch_id}", Components: []string{"epoch_id"}},
		{Domain: "epoch", Name: "seed", Template: "epoch/seed/{epoch_id}", Components: []string{"epoch_id"}},
		{Domain: "validator_economy", Name: "scores", Template: "valecon/scores/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "validator_economy", Name: "effective_stake", Template: "valecon/effective_stake/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "validator_economy", Name: "saturation", Template: "valecon/saturation/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "validator_economy", Name: "roles", Template: "valecon/roles/{epoch_id}/{validator}/{role}", Components: []string{"epoch_id", "validator", "role"}},
		{Domain: "taskgroups", Name: "groups", Template: "taskgroups/groups/{epoch_id}/{task_group_id}", Components: []string{"epoch_id", "task_group_id"}},
		{Domain: "taskgroups", Name: "workloads", Template: "taskgroups/workloads/{workload_id}", Components: []string{"workload_id"}},
		{Domain: "taskgroups", Name: "assignments", Template: "taskgroups/assignments/{epoch_id}/{validator}/{task_group_id}", Components: []string{"epoch_id", "validator", "task_group_id"}},
		{Domain: "taskgroups", Name: "proposer", Template: "taskgroups/proposer/{epoch_id}/{slot}/{task_group_id}", Components: []string{"epoch_id", "slot", "task_group_id"}},
		{Domain: "evidence", Name: "records", Template: "evidence/records/{evidence_id}", Components: []string{"evidence_id"}},
		{Domain: "evidence", Name: "by_accused", Template: "evidence/by_accused/{validator}/{evidence_id}", Components: []string{"validator", "evidence_id"}},
		{Domain: "evidence", Name: "by_reporter", Template: "evidence/by_reporter/{reporter}/{evidence_id}", Components: []string{"reporter", "evidence_id"}},
		{Domain: "evidence", Name: "verification_groups", Template: "evidence/verification_groups/{evidence_id}", Components: []string{"evidence_id"}},
		{Domain: "evidence", Name: "deposits", Template: "evidence/deposits/{evidence_id}", Components: []string{"evidence_id"}},
		{Domain: "performance", Name: "records", Template: "performance/records/{epoch_id}/{operator}/{role}", Components: []string{"epoch_id", "operator", "role"}},
		{Domain: "performance", Name: "uptime", Template: "performance/uptime/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "performance", Name: "correctness", Template: "performance/correctness/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "performance", Name: "tasks", Template: "performance/tasks/{epoch_id}/{validator}", Components: []string{"epoch_id", "validator"}},
		{Domain: "risk", Name: "unbonding", Template: "risk/unbonding/{delegator}/{validator}/{creation_height}", Components: []string{"delegator", "validator", "creation_height"}},
		{Domain: "risk", Name: "redelegation", Template: "risk/redelegation/{delegator}/{src_validator}/{dst_validator}/{epoch_id}", Components: []string{"delegator", "src_validator", "dst_validator", "epoch_id"}},
		{Domain: "risk", Name: "exposure", Template: "risk/exposure/{epoch_id}/{validator}/{delegator}", Components: []string{"epoch_id", "validator", "delegator"}},
	}}
	manifest.Root = ComputeStateModelRoot(manifest)
	return manifest
}

func (m StateModelManifest) Validate() error {
	if len(m.Keys) == 0 {
		return errors.New("state model keys are required")
	}
	seenTemplates := make(map[string]struct{}, len(m.Keys))
	seenNames := make(map[string]struct{}, len(m.Keys))
	for _, key := range m.Keys {
		if err := key.Validate(); err != nil {
			return err
		}
		if _, found := seenTemplates[key.Template]; found {
			return fmt.Errorf("duplicate state key template %s", key.Template)
		}
		seenTemplates[key.Template] = struct{}{}
		qualified := key.Domain + "/" + key.Name
		if _, found := seenNames[qualified]; found {
			return fmt.Errorf("duplicate state key name %s", qualified)
		}
		seenNames[qualified] = struct{}{}
	}
	if err := validatePosHash("state model root", m.Root); err != nil {
		return err
	}
	if expected := ComputeStateModelRoot(m); expected != m.Root {
		return errors.New("state model root mismatch")
	}
	return nil
}

func (s StateKeySpec) Validate() error {
	if err := validatePosToken("state key domain", s.Domain); err != nil {
		return err
	}
	if err := validatePosToken("state key name", s.Name); err != nil {
		return err
	}
	if strings.TrimSpace(s.Template) != s.Template || s.Template == "" {
		return errors.New("state key template is required and must not have surrounding whitespace")
	}
	if strings.Contains(s.Template, "//") {
		return fmt.Errorf("state key template %s must not contain empty segments", s.Template)
	}
	for _, component := range s.Components {
		if err := validatePosToken("state key component", component); err != nil {
			return err
		}
		if !strings.Contains(s.Template, "{"+component+"}") {
			return fmt.Errorf("state key component %s is not present in template %s", component, s.Template)
		}
	}
	return nil
}

func ComputeStateModelRoot(manifest StateModelManifest) string {
	return posHashRoot("aetheris-pos-state-model-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(manifest.Keys)))
		for _, key := range manifest.Keys {
			posWritePart(w, key.Domain)
			posWritePart(w, key.Name)
			posWritePart(w, key.Template)
			posWriteStringSlice(w, key.Components)
		}
	})
}

func EpochCurrentKey() string	{ return "epoch/current" }
func EpochRecordKey(epochID uint64) string {
	return stateKey("epoch", "records", uint64StateComponent(epochID))
}
func EpochPhaseKey(epochID uint64) string {
	return stateKey("epoch", "phase", uint64StateComponent(epochID))
}
func EpochSeedKey(epochID uint64) string {
	return stateKey("epoch", "seed", uint64StateComponent(epochID))
}
func ValidatorScoreKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("valecon", "scores", uint64StateComponent(epochID), validator)
}
func ValidatorEffectiveStakeKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("valecon", "effective_stake", uint64StateComponent(epochID), validator)
}
func ValidatorSaturationKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("valecon", "saturation", uint64StateComponent(epochID), validator)
}
func ValidatorRoleKey(epochID uint64, validator string, role ValidatorRole) (string, error) {
	return stateKeyChecked("valecon", "roles", uint64StateComponent(epochID), validator, string(role))
}
func TaskGroupKey(epochID uint64, taskGroupID string) (string, error) {
	return stateKeyChecked("taskgroups", "groups", uint64StateComponent(epochID), taskGroupID)
}
func WorkloadKey(workloadID string) (string, error) {
	return stateKeyChecked("taskgroups", "workloads", workloadID)
}
func TaskAssignmentKey(epochID uint64, validator string, taskGroupID string) (string, error) {
	return stateKeyChecked("taskgroups", "assignments", uint64StateComponent(epochID), validator, taskGroupID)
}
func ProposerKey(epochID uint64, slot uint64, taskGroupID string) (string, error) {
	return stateKeyChecked("taskgroups", "proposer", uint64StateComponent(epochID), uint64StateComponent(slot), taskGroupID)
}
func EvidenceRecordKey(evidenceID string) (string, error) {
	return stateKeyChecked("evidence", "records", evidenceID)
}
func EvidenceByAccusedKey(validator string, evidenceID string) (string, error) {
	return stateKeyChecked("evidence", "by_accused", validator, evidenceID)
}
func EvidenceByReporterKey(reporter string, evidenceID string) (string, error) {
	return stateKeyChecked("evidence", "by_reporter", reporter, evidenceID)
}
func EvidenceVerificationGroupKey(evidenceID string) (string, error) {
	return stateKeyChecked("evidence", "verification_groups", evidenceID)
}
func EvidenceDepositKey(evidenceID string) (string, error) {
	return stateKeyChecked("evidence", "deposits", evidenceID)
}
func PerformanceRecordKey(epochID uint64, operator string, role ValidatorRole) (string, error) {
	return stateKeyChecked("performance", "records", uint64StateComponent(epochID), operator, string(role))
}
func PerformanceUptimeKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("performance", "uptime", uint64StateComponent(epochID), validator)
}
func PerformanceCorrectnessKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("performance", "correctness", uint64StateComponent(epochID), validator)
}
func PerformanceTasksKey(epochID uint64, validator string) (string, error) {
	return stateKeyChecked("performance", "tasks", uint64StateComponent(epochID), validator)
}
func RiskUnbondingKey(delegator string, validator string, creationHeight uint64) (string, error) {
	return stateKeyChecked("risk", "unbonding", delegator, validator, uint64StateComponent(creationHeight))
}
func RiskRedelegationKey(delegator string, sourceValidator string, destinationValidator string, epochID uint64) (string, error) {
	return stateKeyChecked("risk", "redelegation", delegator, sourceValidator, destinationValidator, uint64StateComponent(epochID))
}
func RiskExposureKey(epochID uint64, validator string, delegator string) (string, error) {
	return stateKeyChecked("risk", "exposure", uint64StateComponent(epochID), validator, delegator)
}

func (a LayeredPosArchitecture) Validate() error {
	if len(a.Layers) != len(DefaultPosLayerOrder()) {
		return errors.New("layered pos architecture must define all layers")
	}
	expectedOrder := DefaultPosLayerOrder()
	seen := make(map[PosLayer]int, len(a.Layers))
	for i, layer := range a.Layers {
		if layer.Layer != expectedOrder[i] {
			return fmt.Errorf("pos layer %d must be %s", i, expectedOrder[i])
		}
		if _, found := seen[layer.Layer]; found {
			return fmt.Errorf("duplicate pos layer %s", layer.Layer)
		}
		seen[layer.Layer] = i
		if err := layer.Validate(); err != nil {
			return err
		}
	}
	for _, layer := range a.Layers {
		layerIndex := seen[layer.Layer]
		for _, dependency := range layer.DependsOn {
			dependencyIndex, found := seen[dependency]
			if !found {
				return fmt.Errorf("pos layer %s depends on unknown layer %s", layer.Layer, dependency)
			}
			if dependencyIndex <= layerIndex {
				return fmt.Errorf("pos layer %s must depend only on lower layers", layer.Layer)
			}
		}
	}
	if err := validatePosHash("layered pos architecture root", a.Root); err != nil {
		return err
	}
	if expected := ComputeLayeredPosArchitectureRoot(a.Layers); a.Root != expected {
		return errors.New("layered pos architecture root mismatch")
	}
	return nil
}

func (s PosLayerSpec) Validate() error {
	if err := validatePosLayer(s.Layer); err != nil {
		return err
	}
	if len(s.Responsibilities) == 0 {
		return fmt.Errorf("pos layer %s responsibilities are required", s.Layer)
	}
	for _, responsibility := range s.Responsibilities {
		if err := validatePosResponsibility("pos layer responsibility", responsibility); err != nil {
			return err
		}
	}
	seen := make(map[PosLayer]struct{}, len(s.DependsOn))
	for _, dependency := range s.DependsOn {
		if err := validatePosLayer(dependency); err != nil {
			return err
		}
		if dependency == s.Layer {
			return fmt.Errorf("pos layer %s cannot depend on itself", s.Layer)
		}
		if _, found := seen[dependency]; found {
			return fmt.Errorf("duplicate dependency %s for pos layer %s", dependency, s.Layer)
		}
		seen[dependency] = struct{}{}
	}
	return nil
}

func DefaultPosLayerOrder() []PosLayer {
	return []PosLayer{
		PosLayerEconomicConsensus,
		PosLayerTaskAssignment,
		PosLayerValidatorExecution,
		PosLayerStakingCapital,
		PosLayerBaseCometBFT,
	}
}

func ComputeLayeredPosArchitectureRoot(layers []PosLayerSpec) string {
	return posHashRoot("aetheris-pos-layered-architecture-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(layers)))
		for _, layer := range layers {
			posWritePart(w, string(layer.Layer))
			posWriteUint64(w, uint64(len(layer.Responsibilities)))
			for _, responsibility := range layer.Responsibilities {
				posWritePart(w, responsibility)
			}
			posWriteUint64(w, uint64(len(layer.DependsOn)))
			for _, dependency := range layer.DependsOn {
				posWritePart(w, string(dependency))
			}
		}
	})
}

func DefaultEpochLifecycle() []EpochLifecycleStep {
	return []EpochLifecycleStep{
		{Phase: EpochPhaseDelegation, Name: "delegation phase", DurationKey: "delegation_phase_duration"},
		{Phase: EpochPhaseElection, Name: "validator election", DurationKey: "election_phase_duration"},
		{Phase: EpochPhaseAssignment, Name: "task group assignment", DurationKey: "assignment_phase_duration"},
		{Phase: EpochPhaseActive, Name: "active validation", DurationKey: "active_validation_duration"},
		{Phase: EpochPhaseSettlement, Name: "settlement + reward + slash finality", DurationKey: "settlement_phase_duration"},
	}
}

func ValidateEpochLifecycle(lifecycle []EpochLifecycleStep) error {
	expected := DefaultEpochLifecycle()
	if len(lifecycle) != len(expected) {
		return errors.New("epoch lifecycle must define every active phase")
	}
	seen := make(map[EpochPhase]struct{}, len(lifecycle))
	for i, step := range lifecycle {
		if step.Phase != expected[i].Phase {
			return fmt.Errorf("epoch lifecycle step %d must be %s", i, expected[i].Phase)
		}
		if _, found := seen[step.Phase]; found {
			return fmt.Errorf("duplicate epoch lifecycle phase %s", step.Phase)
		}
		seen[step.Phase] = struct{}{}
		if strings.TrimSpace(step.Name) != step.Name || step.Name == "" {
			return fmt.Errorf("epoch lifecycle phase %s name is required", step.Phase)
		}
		if strings.TrimSpace(step.DurationKey) != step.DurationKey || step.DurationKey == "" {
			return fmt.Errorf("epoch lifecycle phase %s duration key is required", step.Phase)
		}
	}
	return nil
}

func NextEpochPhase(phase EpochPhase) (EpochPhase, bool, error) {
	switch phase {
	case EpochPhaseDelegation:
		return EpochPhaseElection, false, nil
	case EpochPhaseElection:
		return EpochPhaseAssignment, false, nil
	case EpochPhaseAssignment:
		return EpochPhaseActive, false, nil
	case EpochPhaseActive:
		return EpochPhaseSettlement, false, nil
	case EpochPhaseSettlement:
		return EpochPhaseClosed, true, nil
	case EpochPhaseClosed:
		return EpochPhaseClosed, true, nil
	default:
		return "", false, fmt.Errorf("unsupported epoch phase %q", phase)
	}
}

func ValidateEpochPhaseTransition(from EpochPhase, to EpochPhase) error {
	next, _, err := NextEpochPhase(from)
	if err != nil {
		return err
	}
	if next != to {
		return fmt.Errorf("invalid epoch phase transition from %s to %s", from, to)
	}
	return nil
}

func (s EpochSeedSource) Validate() error {
	switch s {
	case EpochSeedSourcePreviousSeedValidatorSet, EpochSeedSourceCometBFTBlockID, EpochSeedSourceExternalBeacon:
		return nil
	default:
		return fmt.Errorf("unsupported epoch seed source %q", s)
	}
}

func (p Params) EffectiveEpochSeedSource() EpochSeedSource {
	if p.EpochSeedSource == "" {
		return EpochSeedSourcePreviousSeedValidatorSet
	}
	return p.EpochSeedSource
}

func MaxValidatorSetChanges(params Params, activeValidatorCount uint32) (uint32, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if activeValidatorCount == 0 {
		return 0, errors.New("active validator count must be positive")
	}
	changes := (uint64(activeValidatorCount)*uint64(params.MaxValidatorSetChangeRateBps) + uint64(BasisPoints) - 1) / uint64(BasisPoints)
	if changes == 0 {
		return 1, nil
	}
	if changes > uint64(activeValidatorCount) {
		return activeValidatorCount, nil
	}
	return uint32(changes), nil
}

func BuildElectionCandidates(params Params, electionEpoch uint64, candidates []Candidate, intents []DelegationIntent) ([]Candidate, []RejectedDelegationIntent, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	out := make([]Candidate, len(candidates))
	indexByID := make(map[string]int, len(candidates))
	for i, candidate := range candidates {
		cloned := cloneCandidate(candidate)
		if err := cloned.Validate(params); err != nil {
			return nil, nil, err
		}
		if _, found := indexByID[cloned.ValidatorID]; found {
			return nil, nil, fmt.Errorf("duplicate candidate %q", cloned.ValidatorID)
		}
		indexByID[cloned.ValidatorID] = i
		out[i] = cloned
	}
	activations, rejected, err := ActivateDelegationIntents(params, electionEpoch, out, intents)
	if err != nil {
		return nil, nil, err
	}
	for _, activation := range activations {
		idx, found := indexByID[activation.ValidatorID]
		if !found {
			continue
		}
		out[idx].Nominations = mergeNominations(out[idx].Nominations, activation.Nominations)
		out[idx].DelegatedStakeNaet = sumNominations(out[idx].Nominations)
	}
	return out, rejected, nil
}

func ComputePerformanceFactor(input PerformanceFactorInput) (uint32, error) {
	completionDenom := input.CompletedTasks + input.MissedTasks
	if completionDenom < input.CompletedTasks {
		return 0, errors.New("performance task count overflow")
	}
	correctnessDenom := input.CorrectVerifications + input.IncorrectVerifications
	if correctnessDenom < input.CorrectVerifications {
		return 0, errors.New("performance verification count overflow")
	}
	if input.CommittedWindows < input.AvailableWindows {
		return 0, errors.New("available windows cannot exceed committed windows")
	}
	completion := ratioBps(input.CompletedTasks, completionDenom)
	correctness := ratioBps(input.CorrectVerifications, correctnessDenom)
	availability := ratioBps(input.AvailableWindows, input.CommittedWindows)
	score := uint64(4_000)*uint64(completion) +
		uint64(4_000)*uint64(correctness) +
		uint64(2_000)*uint64(availability)
	return uint32(score / uint64(BasisPoints)), nil
}

func ComputeUptimeFactor(input UptimeFactorInput) (uint32, error) {
	if input.TotalBlocks < input.SignedBlocks {
		return 0, errors.New("signed blocks cannot exceed total blocks")
	}
	totalTaskParticipations := input.TaskParticipations + input.MissedTaskParticipations
	if totalTaskParticipations < input.TaskParticipations {
		return 0, errors.New("task participation count overflow")
	}
	blocks := ratioBps(input.SignedBlocks, input.TotalBlocks)
	tasks := ratioBps(input.TaskParticipations, totalTaskParticipations)
	score := uint64(7_000)*uint64(blocks) + uint64(3_000)*uint64(tasks)
	return uint32(score / uint64(BasisPoints)), nil
}

func ComputeLatencyFactor(input LatencyFactorInput) (uint32, error) {
	if !input.CommittedWindow {
		return 0, errors.New("latency factor requires committed measurement window")
	}
	if input.AdvisoryOnly {
		return BasisPoints, nil
	}
	if input.TargetMillis == 0 {
		return 0, errors.New("latency target must be positive")
	}
	if input.P95Millis == 0 || input.P95Millis <= input.TargetMillis {
		return BasisPoints, nil
	}
	return uint32(sdkmath.NewIntFromUint64(input.TargetMillis).MulRaw(int64(BasisPoints)).Quo(sdkmath.NewIntFromUint64(input.P95Millis)).Uint64()), nil
}

func ComputeReliabilityIndex(input ReliabilityIndexInput) (uint32, error) {
	if input.PriorIndexBps > BasisPoints {
		return 0, fmt.Errorf("prior reliability index must be <= %d bps", BasisPoints)
	}
	index := input.PriorIndexBps
	if index == 0 {
		index = BasisPoints
	}
	penalty, err := reliabilityPenalty(input)
	if err != nil {
		return 0, err
	}
	if penalty >= uint64(index) {
		index = 0
	} else {
		index -= uint32(penalty)
	}
	recovery := input.RecoveryEpochs * 100
	if recovery > uint64(BasisPoints-index) {
		return BasisPoints, nil
	}
	return index + uint32(recovery), nil
}

func ComputeCorrectnessScore(input CorrectnessScoreInput) (uint32, error) {
	penaltyWeight := input.EvidencePenaltyWeight
	if penaltyWeight == 0 {
		penaltyWeight = 2
	}
	validUnits, err := checkedAddUint64(input.ValidSignatures, input.ValidTaskOutputs, "correctness valid unit overflow")
	if err != nil {
		return 0, err
	}
	invalidUnits, err := checkedAddUint64(input.InvalidSignatures, input.InvalidTaskOutputs, "correctness invalid unit overflow")
	if err != nil {
		return 0, err
	}
	evidenceFaults, overflow := mulUint64Overflow(input.AcceptedEvidence, penaltyWeight)
	if overflow {
		return 0, errors.New("correctness evidence penalty overflow")
	}
	faultUnits, err := checkedAddUint64(invalidUnits, evidenceFaults, "correctness fault unit overflow")
	if err != nil {
		return 0, err
	}
	totalUnits, err := checkedAddUint64(validUnits, faultUnits, "correctness total unit overflow")
	if err != nil {
		return 0, err
	}
	return ratioBps(validUnits, totalUnits), nil
}

func ComputeTaskCompletionRate(input TaskCompletionRateInput) (uint32, error) {
	if input.CompletedAssignedTasks > input.ExpectedAssignedTasks {
		return 0, errors.New("completed assigned tasks cannot exceed expected tasks")
	}
	return ratioBps(input.CompletedAssignedTasks, input.ExpectedAssignedTasks), nil
}

func ComputePerformanceBasedReward(input PerformanceRewardInput) (PerformanceRewardRecord, error) {
	input.ValidatorID = strings.TrimSpace(input.ValidatorID)
	if input.EpochID == 0 {
		return PerformanceRewardRecord{}, errors.New("performance reward epoch id is required")
	}
	if err := validatePosToken("performance reward validator id", input.ValidatorID); err != nil {
		return PerformanceRewardRecord{}, err
	}
	if input.BaseEmissionNaet.IsNil() {
		input.BaseEmissionNaet = sdkmath.ZeroInt()
	}
	if input.BaseEmissionNaet.IsNegative() {
		return PerformanceRewardRecord{}, errors.New("performance reward base emission cannot be negative")
	}
	if err := validatePerformanceRewardBps(input.UptimeScoreBps, input.LatencyScoreBps, input.CorrectnessScoreBps, input.TaskCompletionRateBps); err != nil {
		return PerformanceRewardRecord{}, err
	}
	reward := input.BaseEmissionNaet
	reward = mulIntBps(reward, input.UptimeScoreBps)
	reward = mulIntBps(reward, input.LatencyScoreBps)
	reward = mulIntBps(reward, input.CorrectnessScoreBps)
	reward = mulIntBps(reward, input.TaskCompletionRateBps)
	record := PerformanceRewardRecord{
		EpochID:		input.EpochID,
		ValidatorID:		input.ValidatorID,
		BaseEmissionNaet:	input.BaseEmissionNaet,
		UptimeScoreBps:		input.UptimeScoreBps,
		LatencyScoreBps:	input.LatencyScoreBps,
		CorrectnessScoreBps:	input.CorrectnessScoreBps,
		TaskCompletionRateBps:	input.TaskCompletionRateBps,
		RewardNaet:		reward,
	}
	record.RewardHash = ComputePerformanceRewardHash(record)
	return record, record.Validate()
}

func (r PerformanceRewardRecord) Validate() error {
	if r.EpochID == 0 {
		return errors.New("performance reward epoch id is required")
	}
	if err := validatePosToken("performance reward validator id", r.ValidatorID); err != nil {
		return err
	}
	if r.BaseEmissionNaet.IsNil() || r.BaseEmissionNaet.IsNegative() {
		return errors.New("performance reward base emission cannot be nil or negative")
	}
	if r.RewardNaet.IsNil() || r.RewardNaet.IsNegative() {
		return errors.New("performance reward amount cannot be nil or negative")
	}
	if r.RewardNaet.GT(r.BaseEmissionNaet) {
		return errors.New("performance reward cannot exceed base emission")
	}
	if err := validatePerformanceRewardBps(r.UptimeScoreBps, r.LatencyScoreBps, r.CorrectnessScoreBps, r.TaskCompletionRateBps); err != nil {
		return err
	}
	if err := validatePosHash("performance reward hash", r.RewardHash); err != nil {
		return err
	}
	if expected := ComputePerformanceRewardHash(r); expected != r.RewardHash {
		return errors.New("performance reward hash mismatch")
	}
	return nil
}

func ComputePerformanceRewardHash(record PerformanceRewardRecord) string {
	return posHashRoot("aetheris-pos-performance-reward-v1", func(w posByteWriter) {
		posWriteUint64(w, record.EpochID)
		posWritePart(w, record.ValidatorID)
		posWritePart(w, record.BaseEmissionNaet.String())
		posWriteUint64(w, uint64(record.UptimeScoreBps))
		posWriteUint64(w, uint64(record.LatencyScoreBps))
		posWriteUint64(w, uint64(record.CorrectnessScoreBps))
		posWriteUint64(w, uint64(record.TaskCompletionRateBps))
		posWritePart(w, record.RewardNaet.String())
	})
}

func PerformanceRecordFieldNames() []string {
	return []string{
		"epoch_id",
		"operator_address",
		"role",
		"assigned_tasks",
		"completed_tasks",
		"missed_tasks",
		"invalid_tasks",
		"uptime_score",
		"latency_score",
		"correctness_score",
		"task_completion_rate",
		"reward_multiplier",
	}
}

func BuildPerformanceRecord(input PerformanceRecordInput) (PerformanceRecord, error) {
	input.OperatorAddress = strings.TrimSpace(input.OperatorAddress)
	if input.AssignedTasks < input.CompletedTasks {
		return PerformanceRecord{}, errors.New("completed tasks cannot exceed assigned tasks")
	}
	if input.AssignedTasks < input.MissedTasks {
		return PerformanceRecord{}, errors.New("missed tasks cannot exceed assigned tasks")
	}
	if input.AssignedTasks < input.InvalidTasks {
		return PerformanceRecord{}, errors.New("invalid tasks cannot exceed assigned tasks")
	}
	observed, err := checkedAddUint64(input.CompletedTasks, input.MissedTasks, "performance observed task overflow")
	if err != nil {
		return PerformanceRecord{}, err
	}
	observed, err = checkedAddUint64(observed, input.InvalidTasks, "performance observed task overflow")
	if err != nil {
		return PerformanceRecord{}, err
	}
	if observed > input.AssignedTasks {
		return PerformanceRecord{}, errors.New("performance task counts exceed assigned tasks")
	}
	taskCompletion, err := ComputeTaskCompletionRate(TaskCompletionRateInput{
		CompletedAssignedTasks:	input.CompletedTasks,
		ExpectedAssignedTasks:	input.AssignedTasks,
	})
	if err != nil {
		return PerformanceRecord{}, err
	}
	multiplier, err := computeRewardMultiplierBps(input.UptimeScoreBps, input.LatencyScoreBps, input.CorrectnessScoreBps, taskCompletion)
	if err != nil {
		return PerformanceRecord{}, err
	}
	record := PerformanceRecord{
		EpochID:		input.EpochID,
		OperatorAddress:	input.OperatorAddress,
		Role:			input.Role,
		AssignedTasks:		input.AssignedTasks,
		CompletedTasks:		input.CompletedTasks,
		MissedTasks:		input.MissedTasks,
		InvalidTasks:		input.InvalidTasks,
		UptimeScoreBps:		input.UptimeScoreBps,
		LatencyScoreBps:	input.LatencyScoreBps,
		CorrectnessScoreBps:	input.CorrectnessScoreBps,
		TaskCompletionRateBps:	taskCompletion,
		RewardMultiplierBps:	multiplier,
	}
	return record, record.Validate()
}

func (r PerformanceRecord) Validate() error {
	if r.EpochID == 0 {
		return errors.New("performance record epoch id is required")
	}
	if err := validatePosToken("performance record operator address", r.OperatorAddress); err != nil {
		return err
	}
	if err := validateValidatorRole(r.Role); err != nil {
		return err
	}
	if r.AssignedTasks < r.CompletedTasks {
		return errors.New("completed tasks cannot exceed assigned tasks")
	}
	if r.AssignedTasks < r.MissedTasks {
		return errors.New("missed tasks cannot exceed assigned tasks")
	}
	if r.AssignedTasks < r.InvalidTasks {
		return errors.New("invalid tasks cannot exceed assigned tasks")
	}
	observed, err := checkedAddUint64(r.CompletedTasks, r.MissedTasks, "performance observed task overflow")
	if err != nil {
		return err
	}
	observed, err = checkedAddUint64(observed, r.InvalidTasks, "performance observed task overflow")
	if err != nil {
		return err
	}
	if observed > r.AssignedTasks {
		return errors.New("performance task counts exceed assigned tasks")
	}
	if err := validatePerformanceRewardBps(r.UptimeScoreBps, r.LatencyScoreBps, r.CorrectnessScoreBps, r.TaskCompletionRateBps); err != nil {
		return err
	}
	if r.RewardMultiplierBps > BasisPoints {
		return fmt.Errorf("reward multiplier must be <= %d bps", BasisPoints)
	}
	expectedMultiplier, err := computeRewardMultiplierBps(r.UptimeScoreBps, r.LatencyScoreBps, r.CorrectnessScoreBps, r.TaskCompletionRateBps)
	if err != nil {
		return err
	}
	if r.RewardMultiplierBps != expectedMultiplier {
		return errors.New("performance record reward multiplier mismatch")
	}
	return nil
}

func ApplyPerformanceDampening(input PerformanceDampeningInput) (PerformanceDampeningResult, error) {
	if err := input.Record.Validate(); err != nil {
		return PerformanceDampeningResult{}, err
	}
	if input.CurrentRewardNaet.IsNil() {
		input.CurrentRewardNaet = sdkmath.ZeroInt()
	}
	if input.CurrentRewardNaet.IsNegative() {
		return PerformanceDampeningResult{}, errors.New("performance dampening current reward cannot be negative")
	}
	if err := validatePerformanceRewardBps(
		input.FutureElectionScoreBps,
		input.DelegationAttractivenessBps,
		input.RoleEligibilityBps,
		input.CollatorAssignmentProbabilityBps,
	); err != nil {
		return PerformanceDampeningResult{}, err
	}
	multiplier := input.Record.RewardMultiplierBps
	result := PerformanceDampeningResult{
		EpochID:				input.Record.EpochID,
		OperatorAddress:			input.Record.OperatorAddress,
		Role:					input.Record.Role,
		RewardMultiplierBps:			multiplier,
		CurrentRewardNaet:			mulIntBps(input.CurrentRewardNaet, multiplier),
		FutureElectionScoreBps:			mulBps(input.FutureElectionScoreBps, multiplier),
		DelegationAttractivenessBps:		mulBps(input.DelegationAttractivenessBps, multiplier),
		RoleEligibilityBps:			mulBps(input.RoleEligibilityBps, multiplier),
		CollatorAssignmentProbabilityBps:	mulBps(input.CollatorAssignmentProbabilityBps, multiplier),
	}
	if input.Record.Role != ValidatorRoleCollator {
		result.CollatorAssignmentProbabilityBps = input.CollatorAssignmentProbabilityBps
	}
	return result, nil
}

func ComputeEconomicSecurityMetrics(input EconomicSecurityInput) (EconomicSecurityMetrics, error) {
	if len(input.Validators) == 0 {
		return EconomicSecurityMetrics{}, errors.New("economic security validators are required")
	}
	if input.TopN == 0 {
		return EconomicSecurityMetrics{}, errors.New("economic security top-n must be positive")
	}
	if input.ParticipatingValidators > input.EligibleValidators {
		return EconomicSecurityMetrics{}, errors.New("participating validators cannot exceed eligible validators")
	}
	if input.AcceptedSlashEvents > input.DetectedFaultEvents {
		return EconomicSecurityMetrics{}, errors.New("accepted slash events cannot exceed detected fault events")
	}
	if input.AcceptedEvidence > input.SubmittedEvidence {
		return EconomicSecurityMetrics{}, errors.New("accepted evidence cannot exceed submitted evidence")
	}
	if input.StakeAtRiskNaet.IsNil() {
		input.StakeAtRiskNaet = sdkmath.ZeroInt()
	}
	if input.StakeAtRiskNaet.IsNegative() {
		return EconomicSecurityMetrics{}, errors.New("stake at risk cannot be negative")
	}
	taskCompletion, err := ComputeTaskCompletionRate(TaskCompletionRateInput{
		CompletedAssignedTasks:	input.CompletedTasks,
		ExpectedAssignedTasks:	input.ExpectedTasks,
	})
	if err != nil {
		return EconomicSecurityMetrics{}, err
	}

	totalBonded := sdkmath.ZeroInt()
	effectiveStake := sdkmath.ZeroInt()
	saturatedStake := sdkmath.ZeroInt()
	totalScore := sdkmath.ZeroInt()
	totalVotingPower := sdkmath.ZeroInt()
	for _, validator := range input.Validators {
		if err := validateScoredValidatorForSecurity(validator); err != nil {
			return EconomicSecurityMetrics{}, err
		}
		totalBonded = totalBonded.Add(validator.TotalStakeNaet)
		effectiveStake = effectiveStake.Add(validator.EffectiveStakeNaet)
		saturatedStake = saturatedStake.Add(validator.ScoreComponents.SaturatedStakeNaet)
		totalScore = totalScore.Add(validator.Score)
		totalVotingPower = totalVotingPower.Add(validator.VotingPowerNaet)
	}

	stakeAtRisk := input.StakeAtRiskNaet
	riskDistribution, riskExposure, err := BuildDelegationRiskDistribution(input.RiskWindows)
	if err != nil {
		return EconomicSecurityMetrics{}, err
	}
	if !stakeAtRisk.IsPositive() {
		stakeAtRisk = riskExposure
	}
	if !stakeAtRisk.IsPositive() {
		stakeAtRisk = totalBonded
	}

	participation := ratioBps(input.ParticipatingValidators, input.EligibleValidators)
	slashingEfficiency := ratioBps(input.AcceptedSlashEvents, input.DetectedFaultEvents)
	security := mulIntBps(stakeAtRisk, participation)
	security = mulIntBps(security, slashingEfficiency)

	return EconomicSecurityMetrics{
		TotalBondedStakeNaet:			totalBonded,
		EffectiveStakeNaet:			effectiveStake,
		TotalStakeAtRiskNaet:			stakeAtRisk,
		StakeSaturationRatioBps:		intRatioBps(saturatedStake, totalBonded),
		TopN:					input.TopN,
		TopNVotingPowerConcentrationBps:	TopNVotingPowerConcentrationBps(input.Validators, input.TopN),
		ParticipationRateBps:			participation,
		SlashingEfficiencyBps:			slashingEfficiency,
		EvidenceAcceptanceRateBps:		ratioBps(input.AcceptedEvidence, input.SubmittedEvidence),
		AverageValidatorScore:			totalScore.QuoRaw(int64(len(input.Validators))),
		DelegationRiskDistribution:		riskDistribution,
		TaskCompletionRateBps:			taskCompletion,
		SecurityNaet:				security,
	}, nil
}

func QuerySecurityMetrics(query SecurityMetricQuery) (SecurityMetricQueryResult, error) {
	metrics, err := ComputeEconomicSecurityMetrics(query.Input)
	if err != nil {
		return SecurityMetricQueryResult{}, err
	}
	return SecurityMetricQueryResult{Metrics: metrics}, nil
}

func DefaultCentralizationControlParams(params Params) CentralizationControlParams {
	maxValidatorShare := params.MaxVotingPowerBps
	if maxValidatorShare == 0 {
		maxValidatorShare = DefaultMaxVotingPowerBps
	}
	return CentralizationControlParams{
		MaxValidatorShareBps:			maxValidatorShare,
		MaxTopNConcentrationBps:		6_700,
		MaxStakeSaturationRatioBps:		2_500,
		MaxDelegationRiskBucketBps:		5_000,
		MinBootstrapPerformanceBps:		9_000,
		MinBootstrapReliabilityBps:		9_000,
		MaxTaskAssignmentShareBps:		5_000,
		BootstrapMaxVotingPowerShareBps:	maxValidatorShare / 2,
	}
}

func BuildCentralizationDashboard(input CentralizationDashboardInput) (CentralizationDashboardData, error) {
	params := input.ControlParams
	if err := params.Validate(); err != nil {
		return CentralizationDashboardData{}, err
	}
	metrics, err := ComputeEconomicSecurityMetrics(input.SecurityInput)
	if err != nil {
		return CentralizationDashboardData{}, err
	}
	taskDiversity, err := BuildTaskAssignmentDiversityReport(input.TaskAssignments, params.MaxTaskAssignmentShareBps)
	if err != nil {
		return CentralizationDashboardData{}, err
	}
	alerts := ValidateConcentrationInvariants(metrics, params)
	if taskDiversity.MaxAssignmentShareBps > params.MaxTaskAssignmentShareBps {
		alerts = append(alerts, ConcentrationInvariantAlert{
			AlertType:	CentralizationWarningTaskAssignmentShare,
			Severity:	ConcentrationAlertSeverityWarning,
			ObservedBps:	taskDiversity.MaxAssignmentShareBps,
			ThresholdBps:	params.MaxTaskAssignmentShareBps,
		})
	}
	return CentralizationDashboardData{
		Metrics:			metrics,
		ValidatorControls:		BuildCentralizationValidatorControls(input.SecurityInput.Validators, metrics, params),
		DelegationRiskWarnings:		BuildDelegationRiskWarnings(metrics.DelegationRiskDistribution, metrics.TotalStakeAtRiskNaet, params.MaxDelegationRiskBucketBps),
		TaskAssignmentDiversity:	taskDiversity,
		Alerts:				alerts,
	}, nil
}

func BuildCentralizationValidatorControls(validators []ScoredValidator, metrics EconomicSecurityMetrics, params CentralizationControlParams) []CentralizationValidatorControl {
	out := make([]CentralizationValidatorControl, 0, len(validators))
	for _, validator := range validators {
		share := intRatioBps(validator.VotingPowerNaet, metrics.EffectiveStakeNaet)
		dampening := RewardDampeningAboveSoftCapBps(share, params.MaxValidatorShareBps)
		warnings := make([]string, 0)
		if share > params.MaxValidatorShareBps {
			warnings = append(warnings, CentralizationWarningValidatorShare)
		}
		if validator.ScoreComponents.SaturatedStakeNaet.IsPositive() {
			warnings = append(warnings, CentralizationWarningStakeSaturation)
		}
		if dampening > 0 {
			warnings = append(warnings, CentralizationWarningRewardDampeningActive)
		}
		bootstrap := IsBootstrapEligibleReliableValidator(validator, share, params)
		if bootstrap {
			warnings = append(warnings, CentralizationWarningBootstrapEligible)
		}
		out = append(out, CentralizationValidatorControl{
			ValidatorAddress:	validator.ValidatorID,
			VotingPowerShareBps:	share,
			EffectiveStakeNaet:	validator.EffectiveStakeNaet,
			SaturatedStakeNaet:	validator.ScoreComponents.SaturatedStakeNaet,
			RewardDampeningBps:	dampening,
			BootstrapEligible:	bootstrap,
			Warnings:		warnings,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].VotingPowerShareBps != out[j].VotingPowerShareBps {
			return out[i].VotingPowerShareBps > out[j].VotingPowerShareBps
		}
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})
	return out
}

func BuildDelegationRiskWarnings(distribution []DelegationRiskBucket, totalStakeAtRisk sdkmath.Int, thresholdBps uint32) []DelegationRiskWarning {
	warnings := make([]DelegationRiskWarning, 0)
	for _, bucket := range distribution {
		share := intRatioBps(bucket.ExposureNaet, totalStakeAtRisk)
		if share <= thresholdBps {
			continue
		}
		warnings = append(warnings, DelegationRiskWarning{
			ValidatorAddress:	bucket.ValidatorAddress,
			ExposureNaet:		bucket.ExposureNaet,
			ExposureShareBps:	share,
			ThresholdBps:		thresholdBps,
		})
	}
	return warnings
}

func BuildTaskAssignmentDiversityReport(assignments []CentralizationTaskAssignment, maxShareBps uint32) (TaskAssignmentDiversityReport, error) {
	if maxShareBps == 0 || maxShareBps > BasisPoints {
		return TaskAssignmentDiversityReport{}, fmt.Errorf("max task assignment share must be within 1..%d bps", BasisPoints)
	}
	byValidator := make(map[string]uint64)
	total := uint64(0)
	for _, assignment := range assignments {
		if err := validatePosToken("task assignment validator address", assignment.ValidatorAddress); err != nil {
			return TaskAssignmentDiversityReport{}, err
		}
		if err := validatePosToken("task group id", assignment.TaskGroupID); err != nil {
			return TaskAssignmentDiversityReport{}, err
		}
		if assignment.AssignmentCount == 0 {
			return TaskAssignmentDiversityReport{}, errors.New("task assignment count must be positive")
		}
		nextTotal, err := checkedAddUint64(total, assignment.AssignmentCount, "task assignment count overflow")
		if err != nil {
			return TaskAssignmentDiversityReport{}, err
		}
		total = nextTotal
		nextValidatorTotal, err := checkedAddUint64(byValidator[assignment.ValidatorAddress], assignment.AssignmentCount, "validator task assignment count overflow")
		if err != nil {
			return TaskAssignmentDiversityReport{}, err
		}
		byValidator[assignment.ValidatorAddress] = nextValidatorTotal
	}
	report := TaskAssignmentDiversityReport{TotalAssignments: total, DiversityScoreBps: BasisPoints}
	for validator, count := range byValidator {
		if count > report.MaxValidatorAssignments || count == report.MaxValidatorAssignments && validator < report.MaxValidatorAddress {
			report.MaxValidatorAddress = validator
			report.MaxValidatorAssignments = count
		}
	}
	if total > 0 {
		report.MaxAssignmentShareBps = ratioBps(report.MaxValidatorAssignments, total)
		if report.MaxAssignmentShareBps <= BasisPoints {
			report.DiversityScoreBps = BasisPoints - report.MaxAssignmentShareBps
		}
	}
	if report.MaxAssignmentShareBps > maxShareBps {
		report.Warnings = append(report.Warnings, CentralizationWarningTaskAssignmentShare)
	}
	return report, nil
}

func ValidateConcentrationInvariants(metrics EconomicSecurityMetrics, params CentralizationControlParams) []ConcentrationInvariantAlert {
	alerts := make([]ConcentrationInvariantAlert, 0)
	if metrics.TopNVotingPowerConcentrationBps > params.MaxTopNConcentrationBps {
		alerts = append(alerts, concentrationAlert(CentralizationWarningTopNShare, metrics.TopNVotingPowerConcentrationBps, params.MaxTopNConcentrationBps))
	}
	if metrics.StakeSaturationRatioBps > params.MaxStakeSaturationRatioBps {
		alerts = append(alerts, concentrationAlert(CentralizationWarningStakeSaturation, metrics.StakeSaturationRatioBps, params.MaxStakeSaturationRatioBps))
	}
	for _, bucket := range metrics.DelegationRiskDistribution {
		share := intRatioBps(bucket.ExposureNaet, metrics.TotalStakeAtRiskNaet)
		if share > params.MaxDelegationRiskBucketBps {
			alerts = append(alerts, concentrationAlert(CentralizationWarningDelegationRisk, share, params.MaxDelegationRiskBucketBps))
			break
		}
	}
	return alerts
}

func RewardDampeningAboveSoftCapBps(votingPowerShareBps uint32, softCapBps uint32) uint32 {
	if softCapBps >= BasisPoints || votingPowerShareBps <= softCapBps {
		return 0
	}
	over := votingPowerShareBps - softCapBps
	denom := uint64(BasisPoints - softCapBps)
	return uint32((uint64(over) * uint64(BasisPoints)) / denom)
}

func IsBootstrapEligibleReliableValidator(validator ScoredValidator, votingPowerShareBps uint32, params CentralizationControlParams) bool {
	if votingPowerShareBps > params.BootstrapMaxVotingPowerShareBps {
		return false
	}
	if validator.ScoreComponents.PerformanceFactorBps < params.MinBootstrapPerformanceBps {
		return false
	}
	if validator.ScoreComponents.ReliabilityIndexBps < params.MinBootstrapReliabilityBps {
		return false
	}
	return validator.ScoreComponents.SaturatedStakeNaet.IsZero()
}

func SimulateStakeConcentration(input StakeConcentrationSimulationInput) (StakeConcentrationSimulationResult, error) {
	if err := input.Params.Validate(); err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	if input.AddedDelegatedStakeNaet.IsNil() || !input.AddedDelegatedStakeNaet.IsPositive() {
		return StakeConcentrationSimulationResult{}, errors.New("added delegated stake must be positive")
	}
	if err := validatePosToken("target validator id", input.TargetValidatorID); err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	beforeValidators, err := ScoreCandidatesForSecurity(input.Params, input.Candidates)
	if err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	afterCandidates := make([]Candidate, len(input.Candidates))
	for i, candidate := range input.Candidates {
		afterCandidates[i] = cloneCandidate(candidate)
		if afterCandidates[i].ValidatorID == input.TargetValidatorID {
			afterCandidates[i].DelegatedStakeNaet = afterCandidates[i].DelegatedStakeNaet.Add(input.AddedDelegatedStakeNaet)
			afterCandidates[i].Nominations = append(afterCandidates[i].Nominations, Nomination{
				NominatorID:	"simulated-concentration-delegator",
				StakeNaet:	input.AddedDelegatedStakeNaet,
			})
		}
	}
	afterValidators, err := ScoreCandidatesForSecurity(input.Params, afterCandidates)
	if err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	before, err := securityMetricsFromValidators(beforeValidators, input.TopN)
	if err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	after, err := securityMetricsFromValidators(afterValidators, input.TopN)
	if err != nil {
		return StakeConcentrationSimulationResult{}, err
	}
	beforeTarget, found := scoredValidatorByID(beforeValidators, input.TargetValidatorID)
	if !found {
		return StakeConcentrationSimulationResult{}, errors.New("target validator not found")
	}
	afterTarget, found := scoredValidatorByID(afterValidators, input.TargetValidatorID)
	if !found {
		return StakeConcentrationSimulationResult{}, errors.New("target validator not found")
	}
	return StakeConcentrationSimulationResult{
		Before:				before,
		After:				after,
		TopNConcentrationDeltaBps:	int32(after.TopNVotingPowerConcentrationBps) - int32(before.TopNVotingPowerConcentrationBps),
		TargetEffectiveStakeDeltaNaet:	afterTarget.EffectiveStakeNaet.Sub(beforeTarget.EffectiveStakeNaet),
		Alerts:				ValidateConcentrationInvariants(after, DefaultCentralizationControlParams(input.Params)),
	}, nil
}

func SimulateStakeSplitting(input StakeSplittingSimulationInput) (StakeSplittingSimulationResult, error) {
	if err := input.Params.Validate(); err != nil {
		return StakeSplittingSimulationResult{}, err
	}
	if input.SplitCount < 2 {
		return StakeSplittingSimulationResult{}, errors.New("split count must be at least two")
	}
	single, err := ScoreCandidate(input.Params, input.Candidate)
	if err != nil {
		return StakeSplittingSimulationResult{}, err
	}
	totalStake := input.Candidate.SelfStakeNaet.Add(input.Candidate.DelegatedStakeNaet)
	baseStake := totalStake.QuoRaw(int64(input.SplitCount))
	remainder := totalStake.Sub(baseStake.MulRaw(int64(input.SplitCount)))
	splitCandidates := make([]Candidate, input.SplitCount)
	for i := range splitCandidates {
		stake := baseStake
		if i == 0 {
			stake = stake.Add(remainder)
		}
		split := cloneCandidate(input.Candidate)
		split.ValidatorID = fmt.Sprintf("%s-split-%03d", input.Candidate.ValidatorID, i)
		split.SelfStakeNaet = stake
		split.DelegatedStakeNaet = sdkmath.ZeroInt()
		split.Nominations = nil
		splitCandidates[i] = split
	}
	splitValidators, err := ScoreCandidatesForSecurity(input.Params, splitCandidates)
	if err != nil {
		return StakeSplittingSimulationResult{}, err
	}
	singleMetrics, err := securityMetricsFromValidators([]ScoredValidator{single}, input.TopN)
	if err != nil {
		return StakeSplittingSimulationResult{}, err
	}
	splitMetrics, err := securityMetricsFromValidators(splitValidators, input.TopN)
	if err != nil {
		return StakeSplittingSimulationResult{}, err
	}
	splitEffective := sdkmath.ZeroInt()
	for _, validator := range splitValidators {
		splitEffective = splitEffective.Add(validator.EffectiveStakeNaet)
	}
	return StakeSplittingSimulationResult{
		SingleEffectiveStakeNaet:	single.EffectiveStakeNaet,
		SplitEffectiveStakeNaet:	splitEffective,
		EffectiveStakeGainNaet:		splitEffective.Sub(single.EffectiveStakeNaet),
		SingleConcentrationBps:		singleMetrics.TopNVotingPowerConcentrationBps,
		SplitConcentrationBps:		splitMetrics.TopNVotingPowerConcentrationBps,
	}, nil
}

func ScoreCandidatesForSecurity(params Params, candidates []Candidate) ([]ScoredValidator, error) {
	if len(candidates) == 0 {
		return nil, errors.New("security scoring candidates are required")
	}
	validators := make([]ScoredValidator, len(candidates))
	for i, candidate := range candidates {
		scored, err := ScoreCandidate(params, candidate)
		if err != nil {
			return nil, err
		}
		validators[i] = scored
	}
	return validators, nil
}

func securityMetricsFromValidators(validators []ScoredValidator, topN uint32) (EconomicSecurityMetrics, error) {
	return ComputeEconomicSecurityMetrics(EconomicSecurityInput{
		Validators:			validators,
		TopN:				topN,
		ParticipatingValidators:	uint64(len(validators)),
		EligibleValidators:		uint64(len(validators)),
		AcceptedSlashEvents:		1,
		DetectedFaultEvents:		1,
		AcceptedEvidence:		1,
		SubmittedEvidence:		1,
		CompletedTasks:			1,
		ExpectedTasks:			1,
	})
}

func scoredValidatorByID(validators []ScoredValidator, validatorID string) (ScoredValidator, bool) {
	for _, validator := range validators {
		if validator.ValidatorID == validatorID {
			return validator, true
		}
	}
	return ScoredValidator{}, false
}

func concentrationAlert(alertType string, observedBps uint32, thresholdBps uint32) ConcentrationInvariantAlert {
	severity := ConcentrationAlertSeverityWarning
	if thresholdBps > 0 && observedBps >= minUint32(BasisPoints, thresholdBps*2) {
		severity = ConcentrationAlertSeverityCritical
	}
	return ConcentrationInvariantAlert{
		AlertType:	alertType,
		Severity:	severity,
		ObservedBps:	observedBps,
		ThresholdBps:	thresholdBps,
	}
}

func BuildDelegationRiskDistribution(windows []RiskWindowRecord) ([]DelegationRiskBucket, sdkmath.Int, error) {
	byValidator := make(map[string]DelegationRiskBucket)
	totalExposure := sdkmath.ZeroInt()
	for _, window := range windows {
		if err := window.Validate(); err != nil {
			return nil, sdkmath.Int{}, err
		}
		if window.Status == RiskWindowStatusExpired {
			continue
		}
		bucket := byValidator[window.ValidatorAddress]
		bucket.ValidatorAddress = window.ValidatorAddress
		if bucket.ExposureNaet.IsNil() {
			bucket.ExposureNaet = sdkmath.ZeroInt()
		}
		bucket.ExposureNaet = bucket.ExposureNaet.Add(window.AmountNaet)
		bucket.RiskWindowCount++
		byValidator[window.ValidatorAddress] = bucket
		totalExposure = totalExposure.Add(window.AmountNaet)
	}
	distribution := make([]DelegationRiskBucket, 0, len(byValidator))
	for _, bucket := range byValidator {
		distribution = append(distribution, bucket)
	}
	sort.SliceStable(distribution, func(i, j int) bool {
		if !distribution[i].ExposureNaet.Equal(distribution[j].ExposureNaet) {
			return distribution[i].ExposureNaet.GT(distribution[j].ExposureNaet)
		}
		return distribution[i].ValidatorAddress < distribution[j].ValidatorAddress
	})
	return distribution, totalExposure, nil
}

func TopNVotingPowerConcentrationBps(validators []ScoredValidator, topN uint32) uint32 {
	if len(validators) == 0 || topN == 0 {
		return 0
	}
	ordered := make([]ScoredValidator, len(validators))
	copy(ordered, validators)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].VotingPowerNaet.Equal(ordered[j].VotingPowerNaet) {
			return ordered[i].VotingPowerNaet.GT(ordered[j].VotingPowerNaet)
		}
		return ordered[i].ValidatorID < ordered[j].ValidatorID
	})
	total := sdkmath.ZeroInt()
	for _, validator := range ordered {
		if validator.VotingPowerNaet.IsNil() || !validator.VotingPowerNaet.IsPositive() {
			continue
		}
		total = total.Add(validator.VotingPowerNaet)
	}
	if !total.IsPositive() {
		return 0
	}
	limit := int(topN)
	if limit > len(ordered) {
		limit = len(ordered)
	}
	top := sdkmath.ZeroInt()
	for i := 0; i < limit; i++ {
		if ordered[i].VotingPowerNaet.IsNil() || !ordered[i].VotingPowerNaet.IsPositive() {
			continue
		}
		top = top.Add(ordered[i].VotingPowerNaet)
	}
	return intRatioBps(top, total)
}

func validateScoredValidatorForSecurity(validator ScoredValidator) error {
	if err := validatePosToken("economic security validator id", validator.ValidatorID); err != nil {
		return err
	}
	if validator.TotalStakeNaet.IsNil() || validator.TotalStakeNaet.IsNegative() {
		return errors.New("validator total stake cannot be nil or negative")
	}
	if validator.EffectiveStakeNaet.IsNil() || validator.EffectiveStakeNaet.IsNegative() {
		return errors.New("validator effective stake cannot be nil or negative")
	}
	if validator.VotingPowerNaet.IsNil() || validator.VotingPowerNaet.IsNegative() {
		return errors.New("validator voting power cannot be nil or negative")
	}
	if validator.Score.IsNil() || validator.Score.IsNegative() {
		return errors.New("validator score cannot be nil or negative")
	}
	if validator.ScoreComponents.SaturatedStakeNaet.IsNil() || validator.ScoreComponents.SaturatedStakeNaet.IsNegative() {
		return errors.New("validator saturated stake cannot be nil or negative")
	}
	return nil
}

func reliabilityPenalty(input ReliabilityIndexInput) (uint64, error) {
	penalty := sdkmath.NewIntFromUint64(input.SlashEvents).MulRaw(2_000)
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.DowntimeEpochs).MulRaw(500))
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.MissedTasks).MulRaw(100))
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.RejectedEvidence).MulRaw(250))
	if !penalty.LTE(sdkmath.NewIntFromUint64(uint64(BasisPoints))) {
		return uint64(BasisPoints), nil
	}
	return penalty.Uint64(), nil
}

func EpochRecordFieldNames() []string {
	return []string{
		"epoch_id",
		"start_height",
		"end_height",
		"phase",
		"seed",
		"validator_set_hash",
		"task_group_root",
		"performance_root",
		"reward_root",
		"slash_root",
		"settlement_status",
	}
}

func EpochPhaseValues() []EpochPhase {
	return []EpochPhase{
		EpochPhaseDelegation,
		EpochPhaseElection,
		EpochPhaseAssignment,
		EpochPhaseActive,
		EpochPhaseSettlement,
		EpochPhaseClosed,
	}
}

func ValidateValidatorSetChangeActivation(epoch EpochRecord, activationHeight uint64) error {
	if err := epoch.Validate(); err != nil {
		return err
	}
	if activationHeight != epoch.StartHeight {
		return fmt.Errorf("validator set changes must activate at epoch boundary height %d", epoch.StartHeight)
	}
	return nil
}

func ValidateConsecutiveEpochs(previous EpochRecord, next EpochRecord) error {
	if err := previous.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if next.EpochID != previous.EpochID+1 {
		return errors.New("next epoch id must increment by one")
	}
	if next.StartHeight != previous.EndHeight+1 {
		return errors.New("next epoch must start at previous end height plus one")
	}
	return nil
}

func DelegationEffectiveElectionEpoch(params Params, requestedEpoch uint64) (uint64, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	return requestedEpoch + params.DelegationActivationEpochs, nil
}

func DelegationAffectsElection(params Params, requestedEpoch uint64, electionEpoch uint64) (bool, error) {
	effectiveEpoch, err := DelegationEffectiveElectionEpoch(params, requestedEpoch)
	if err != nil {
		return false, err
	}
	return electionEpoch >= effectiveEpoch, nil
}

func EvidenceWithinSlashableWindow(params Params, evidenceEpoch uint64, currentEpoch uint64) (bool, error) {
	if err := params.Validate(); err != nil {
		return false, err
	}
	if currentEpoch < evidenceEpoch {
		return false, errors.New("current epoch cannot be before evidence epoch")
	}
	return currentEpoch-evidenceEpoch <= params.EvidenceWindowEpochs, nil
}

func UnbondingRiskWindowForParams(params Params) (UnbondingRiskWindow, error) {
	if err := params.Validate(); err != nil {
		return UnbondingRiskWindow{}, err
	}
	unbondingEpochs := ceilDivUint64(params.UnbondingSeconds, params.EpochDurationSeconds)
	totalRiskEpochs, err := checkedAddUint64(unbondingEpochs, params.EvidenceWindowEpochs, "unbonding risk window overflow")
	if err != nil {
		return UnbondingRiskWindow{}, err
	}
	return UnbondingRiskWindow{
		UnbondingEpochs:	unbondingEpochs,
		SlashableWindowEpochs:	params.EvidenceWindowEpochs,
		TotalRiskEpochs:	totalRiskEpochs,
	}, nil
}

func BeginUnbondingRisk(params Params, delegatorID string, validatorID string, amount sdkmath.Int, requestedEpoch uint64) (UnbondingRiskRecord, error) {
	if requestedEpoch == 0 {
		return UnbondingRiskRecord{}, errors.New("unbonding requested epoch is required")
	}
	if err := validatePosToken("unbonding delegator id", strings.TrimSpace(delegatorID)); err != nil {
		return UnbondingRiskRecord{}, err
	}
	if err := validatePosToken("unbonding validator id", strings.TrimSpace(validatorID)); err != nil {
		return UnbondingRiskRecord{}, err
	}
	if amount.IsNil() || !amount.IsPositive() {
		return UnbondingRiskRecord{}, errors.New("unbonding amount must be positive")
	}
	window, err := UnbondingRiskWindowForParams(params)
	if err != nil {
		return UnbondingRiskRecord{}, err
	}
	exitEpoch, err := checkedAddUint64(requestedEpoch, window.UnbondingEpochs, "unbonding exit epoch overflow")
	if err != nil {
		return UnbondingRiskRecord{}, err
	}
	slashableUntil, err := checkedAddUint64(requestedEpoch, window.TotalRiskEpochs, "unbonding slashable epoch overflow")
	if err != nil {
		return UnbondingRiskRecord{}, err
	}
	record := UnbondingRiskRecord{
		DelegatorID:		strings.TrimSpace(delegatorID),
		ValidatorID:		strings.TrimSpace(validatorID),
		AmountNaet:		amount,
		RequestedEpoch:		requestedEpoch,
		ExitEpoch:		exitEpoch,
		SlashableUntilEpoch:	slashableUntil,
	}
	record.RiskHistoryKey = ComputeUnbondingRiskHistoryKey(record)
	return record, record.Validate()
}

func (r UnbondingRiskRecord) Validate() error {
	if err := validatePosToken("unbonding delegator id", r.DelegatorID); err != nil {
		return err
	}
	if err := validatePosToken("unbonding validator id", r.ValidatorID); err != nil {
		return err
	}
	if r.AmountNaet.IsNil() || !r.AmountNaet.IsPositive() {
		return errors.New("unbonding amount must be positive")
	}
	if r.RequestedEpoch == 0 {
		return errors.New("unbonding requested epoch is required")
	}
	if r.ExitEpoch <= r.RequestedEpoch {
		return errors.New("unbonding exit epoch must be after requested epoch")
	}
	if r.SlashableUntilEpoch < r.ExitEpoch {
		return errors.New("unbonding slashable window must cover exit epoch")
	}
	if err := validatePosHash("unbonding risk history key", r.RiskHistoryKey); err != nil {
		return err
	}
	if expected := ComputeUnbondingRiskHistoryKey(r); expected != r.RiskHistoryKey {
		return errors.New("unbonding risk history key mismatch")
	}
	return nil
}

func CreateRedelegationRiskRecord(params Params, delegatorID string, sourceValidatorID string, destinationValidatorID string, amount sdkmath.Int, requestedEpoch uint64) (RedelegationRiskRecord, error) {
	if requestedEpoch == 0 {
		return RedelegationRiskRecord{}, errors.New("redelegation requested epoch is required")
	}
	delegatorID = strings.TrimSpace(delegatorID)
	sourceValidatorID = strings.TrimSpace(sourceValidatorID)
	destinationValidatorID = strings.TrimSpace(destinationValidatorID)
	if err := validatePosToken("redelegation delegator id", delegatorID); err != nil {
		return RedelegationRiskRecord{}, err
	}
	if err := validatePosToken("redelegation source validator id", sourceValidatorID); err != nil {
		return RedelegationRiskRecord{}, err
	}
	if err := validatePosToken("redelegation destination validator id", destinationValidatorID); err != nil {
		return RedelegationRiskRecord{}, err
	}
	if sourceValidatorID == destinationValidatorID {
		return RedelegationRiskRecord{}, errors.New("redelegation destination must differ from source")
	}
	if amount.IsNil() || !amount.IsPositive() {
		return RedelegationRiskRecord{}, errors.New("redelegation amount must be positive")
	}
	if err := params.Validate(); err != nil {
		return RedelegationRiskRecord{}, err
	}
	activationEpoch, err := checkedAddUint64(requestedEpoch, params.DelegationActivationEpochs, "redelegation activation epoch overflow")
	if err != nil {
		return RedelegationRiskRecord{}, err
	}
	window, err := UnbondingRiskWindowForParams(params)
	if err != nil {
		return RedelegationRiskRecord{}, err
	}
	sourceSlashableUntil, err := checkedAddUint64(requestedEpoch, window.TotalRiskEpochs, "redelegation source slashable epoch overflow")
	if err != nil {
		return RedelegationRiskRecord{}, err
	}
	record := RedelegationRiskRecord{
		DelegatorID:			delegatorID,
		SourceValidatorID:		sourceValidatorID,
		DestinationValidatorID:		destinationValidatorID,
		AmountNaet:			amount,
		RequestedEpoch:			requestedEpoch,
		ActivationEpoch:		activationEpoch,
		SourceSlashableUntilEpoch:	sourceSlashableUntil,
	}
	record.RiskHistoryKey = ComputeRedelegationRiskHistoryKey(record)
	return record, record.Validate()
}

func (r RedelegationRiskRecord) Validate() error {
	if err := validatePosToken("redelegation delegator id", r.DelegatorID); err != nil {
		return err
	}
	if err := validatePosToken("redelegation source validator id", r.SourceValidatorID); err != nil {
		return err
	}
	if err := validatePosToken("redelegation destination validator id", r.DestinationValidatorID); err != nil {
		return err
	}
	if r.SourceValidatorID == r.DestinationValidatorID {
		return errors.New("redelegation destination must differ from source")
	}
	if r.AmountNaet.IsNil() || !r.AmountNaet.IsPositive() {
		return errors.New("redelegation amount must be positive")
	}
	if r.RequestedEpoch == 0 {
		return errors.New("redelegation requested epoch is required")
	}
	if r.ActivationEpoch <= r.RequestedEpoch {
		return errors.New("redelegation activation epoch must be after requested epoch")
	}
	if r.SourceSlashableUntilEpoch < r.ActivationEpoch {
		return errors.New("redelegation source risk window must cover activation epoch")
	}
	if err := validatePosHash("redelegation risk history key", r.RiskHistoryKey); err != nil {
		return err
	}
	if expected := ComputeRedelegationRiskHistoryKey(r); expected != r.RiskHistoryKey {
		return errors.New("redelegation risk history key mismatch")
	}
	return nil
}

func PlanSelfBondChange(params Params, validatorID string, previousBond sdkmath.Int, newBond sdkmath.Int, requestedEpoch uint64) (SelfBondChangeRecord, error) {
	if err := params.Validate(); err != nil {
		return SelfBondChangeRecord{}, err
	}
	validatorID = strings.TrimSpace(validatorID)
	if err := validatePosToken("self bond validator id", validatorID); err != nil {
		return SelfBondChangeRecord{}, err
	}
	if previousBond.IsNil() || previousBond.IsNegative() || newBond.IsNil() || newBond.IsNegative() {
		return SelfBondChangeRecord{}, errors.New("self bond amounts cannot be nil or negative")
	}
	if requestedEpoch == 0 {
		return SelfBondChangeRecord{}, errors.New("self bond requested epoch is required")
	}
	activationEpoch, err := checkedAddUint64(requestedEpoch, params.DelegationActivationEpochs, "self bond activation epoch overflow")
	if err != nil {
		return SelfBondChangeRecord{}, err
	}
	record := SelfBondChangeRecord{
		ValidatorID:		validatorID,
		PreviousBondNaet:	previousBond,
		NewBondNaet:		newBond,
		RequestedEpoch:		requestedEpoch,
		ActivationEpoch:	activationEpoch,
	}
	return record, record.Validate()
}

func (r SelfBondChangeRecord) Validate() error {
	if err := validatePosToken("self bond validator id", r.ValidatorID); err != nil {
		return err
	}
	if r.PreviousBondNaet.IsNil() || r.PreviousBondNaet.IsNegative() || r.NewBondNaet.IsNil() || r.NewBondNaet.IsNegative() {
		return errors.New("self bond amounts cannot be nil or negative")
	}
	if r.RequestedEpoch == 0 {
		return errors.New("self bond requested epoch is required")
	}
	if r.ActivationEpoch <= r.RequestedEpoch {
		return errors.New("self bond activation epoch must be after requested epoch")
	}
	return nil
}

func PendingUnbondingSlashExposure(input PendingUnbondingSlashExposureInput) (sdkmath.Int, error) {
	if err := input.Record.Validate(); err != nil {
		return sdkmath.Int{}, err
	}
	if input.FaultEpoch == 0 || input.EvidenceEpoch == 0 {
		return sdkmath.Int{}, errors.New("fault and evidence epochs are required")
	}
	if input.FaultEpoch >= input.Record.ExitEpoch {
		return sdkmath.ZeroInt(), nil
	}
	if input.EvidenceEpoch > input.Record.SlashableUntilEpoch {
		return sdkmath.ZeroInt(), nil
	}
	return input.Record.AmountNaet, nil
}

func RiskWindowFromUnbonding(record UnbondingRiskRecord, currentEpoch uint64) (RiskWindowRecord, error) {
	if err := record.Validate(); err != nil {
		return RiskWindowRecord{}, err
	}
	window := RiskWindowRecord{
		StakeOwner:		record.DelegatorID,
		ValidatorAddress:	record.ValidatorID,
		AmountNaet:		record.AmountNaet,
		StartEpoch:		record.RequestedEpoch,
		EndEpoch:		record.ExitEpoch,
		SlashableUntilEpoch:	record.SlashableUntilEpoch,
		Status:			riskWindowStatus(record.ExitEpoch, record.SlashableUntilEpoch, currentEpoch),
	}
	window.RiskHistoryRoot = ComputeRiskWindowRoot(window)
	return window, window.Validate()
}

func RiskWindowFromRedelegation(record RedelegationRiskRecord, currentEpoch uint64) (RiskWindowRecord, error) {
	if err := record.Validate(); err != nil {
		return RiskWindowRecord{}, err
	}
	window := RiskWindowRecord{
		StakeOwner:		record.DelegatorID,
		ValidatorAddress:	record.SourceValidatorID,
		AmountNaet:		record.AmountNaet,
		StartEpoch:		record.RequestedEpoch,
		EndEpoch:		record.ActivationEpoch,
		SlashableUntilEpoch:	record.SourceSlashableUntilEpoch,
		Status:			riskWindowStatus(record.ActivationEpoch, record.SourceSlashableUntilEpoch, currentEpoch),
	}
	window.RiskHistoryRoot = ComputeRiskWindowRoot(window)
	return window, window.Validate()
}

func (r RiskWindowRecord) Validate() error {
	if err := validatePosToken("risk window stake owner", r.StakeOwner); err != nil {
		return err
	}
	if err := validatePosToken("risk window validator address", r.ValidatorAddress); err != nil {
		return err
	}
	if r.AmountNaet.IsNil() || !r.AmountNaet.IsPositive() {
		return errors.New("risk window amount must be positive")
	}
	if r.StartEpoch == 0 {
		return errors.New("risk window start epoch is required")
	}
	if r.EndEpoch <= r.StartEpoch {
		return errors.New("risk window end epoch must be after start epoch")
	}
	if r.SlashableUntilEpoch < r.EndEpoch {
		return errors.New("risk window slashable epoch must cover end epoch")
	}
	if err := validateRiskWindowStatus(r.Status); err != nil {
		return err
	}
	if err := validatePosHash("risk window history root", r.RiskHistoryRoot); err != nil {
		return err
	}
	if expected := ComputeRiskWindowRoot(r); expected != r.RiskHistoryRoot {
		return errors.New("risk window history root mismatch")
	}
	return nil
}

func QuerySlashExposure(windows []RiskWindowRecord, query SlashExposureQuery) (SlashExposureQueryResult, error) {
	query.StakeOwner = strings.TrimSpace(query.StakeOwner)
	query.ValidatorAddress = strings.TrimSpace(query.ValidatorAddress)
	if err := validatePosToken("slash exposure stake owner", query.StakeOwner); err != nil {
		return SlashExposureQueryResult{}, err
	}
	if err := validatePosToken("slash exposure validator address", query.ValidatorAddress); err != nil {
		return SlashExposureQueryResult{}, err
	}
	if query.FaultEpoch == 0 || query.EvidenceEpoch == 0 {
		return SlashExposureQueryResult{}, errors.New("slash exposure fault and evidence epochs are required")
	}
	result := SlashExposureQueryResult{
		StakeOwner:		query.StakeOwner,
		ValidatorAddress:	query.ValidatorAddress,
		FaultEpoch:		query.FaultEpoch,
		EvidenceEpoch:		query.EvidenceEpoch,
		ExposureNaet:		sdkmath.ZeroInt(),
		MatchingWindows:	make([]RiskWindowRecord, 0),
	}
	for _, window := range windows {
		if err := window.Validate(); err != nil {
			return SlashExposureQueryResult{}, err
		}
		if window.StakeOwner != query.StakeOwner || window.ValidatorAddress != query.ValidatorAddress {
			continue
		}
		if query.FaultEpoch < window.StartEpoch || query.FaultEpoch >= window.EndEpoch {
			continue
		}
		if query.EvidenceEpoch > window.SlashableUntilEpoch {
			continue
		}
		if window.Status == RiskWindowStatusExpired {
			continue
		}
		result.ExposureNaet = result.ExposureNaet.Add(window.AmountNaet)
		result.MatchingWindows = append(result.MatchingWindows, window)
	}
	return result, nil
}

func ComputeRiskWindowRoot(record RiskWindowRecord) string {
	return posHashRoot("aetheris-pos-risk-window-v1", func(w posByteWriter) {
		posWritePart(w, record.StakeOwner)
		posWritePart(w, record.ValidatorAddress)
		posWritePart(w, record.AmountNaet.String())
		posWriteUint64(w, record.StartEpoch)
		posWriteUint64(w, record.EndEpoch)
		posWriteUint64(w, record.SlashableUntilEpoch)
		posWritePart(w, record.Status)
	})
}

func ComputeUnbondingRiskHistoryKey(record UnbondingRiskRecord) string {
	return posHashRoot("aetheris-pos-unbonding-risk-v1", func(w posByteWriter) {
		posWritePart(w, record.DelegatorID)
		posWritePart(w, record.ValidatorID)
		posWritePart(w, record.AmountNaet.String())
		posWriteUint64(w, record.RequestedEpoch)
		posWriteUint64(w, record.ExitEpoch)
		posWriteUint64(w, record.SlashableUntilEpoch)
	})
}

func ComputeRedelegationRiskHistoryKey(record RedelegationRiskRecord) string {
	return posHashRoot("aetheris-pos-redelegation-risk-v1", func(w posByteWriter) {
		posWritePart(w, record.DelegatorID)
		posWritePart(w, record.SourceValidatorID)
		posWritePart(w, record.DestinationValidatorID)
		posWritePart(w, record.AmountNaet.String())
		posWriteUint64(w, record.RequestedEpoch)
		posWriteUint64(w, record.ActivationEpoch)
		posWriteUint64(w, record.SourceSlashableUntilEpoch)
	})
}

func DefaultEpochPhaseDurations(epochDurationSeconds uint64) EpochPhaseDurations {
	delegation := epochDurationSeconds / 4
	election := epochDurationSeconds / 12
	assignment := epochDurationSeconds / 12
	settlement := epochDurationSeconds / 12
	active := epochDurationSeconds - delegation - election - assignment - settlement
	return EpochPhaseDurations{
		DelegationSeconds:		delegation,
		ElectionSeconds:		election,
		AssignmentSeconds:		assignment,
		ActiveValidationSeconds:	active,
		SettlementSeconds:		settlement,
	}
}

func (p Params) EffectivePhaseDurations() EpochPhaseDurations {
	baseDefault := DefaultEpochPhaseDurations(DefaultEpochDurationSeconds)
	if p.PhaseDurations.IsZero() ||
		(p.EpochDurationSeconds != DefaultEpochDurationSeconds && p.PhaseDurations == baseDefault) {
		return DefaultEpochPhaseDurations(p.EpochDurationSeconds)
	}
	return p.PhaseDurations
}

func (d EpochPhaseDurations) IsZero() bool {
	return d.DelegationSeconds == 0 &&
		d.ElectionSeconds == 0 &&
		d.AssignmentSeconds == 0 &&
		d.ActiveValidationSeconds == 0 &&
		d.SettlementSeconds == 0
}

func (d EpochPhaseDurations) TotalSeconds() uint64 {
	return d.DelegationSeconds +
		d.ElectionSeconds +
		d.AssignmentSeconds +
		d.ActiveValidationSeconds +
		d.SettlementSeconds
}

func (d EpochPhaseDurations) Validate(epochDurationSeconds uint64) error {
	if d.DelegationSeconds == 0 {
		return errors.New("delegation phase duration must be positive")
	}
	if d.ElectionSeconds == 0 {
		return errors.New("election phase duration must be positive")
	}
	if d.AssignmentSeconds == 0 {
		return errors.New("assignment phase duration must be positive")
	}
	if d.ActiveValidationSeconds == 0 {
		return errors.New("active validation phase duration must be positive")
	}
	if d.SettlementSeconds == 0 {
		return errors.New("settlement phase duration must be positive")
	}
	if d.TotalSeconds() != epochDurationSeconds {
		return fmt.Errorf("epoch phase durations must sum to %d seconds", epochDurationSeconds)
	}
	return nil
}

func EpochPhaseAt(params Params, epochStartUnixSeconds uint64, nowUnixSeconds uint64) (EpochPhase, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}
	if nowUnixSeconds < epochStartUnixSeconds {
		return "", errors.New("epoch phase time cannot be before epoch start")
	}
	elapsed := nowUnixSeconds - epochStartUnixSeconds
	if elapsed >= params.EpochDurationSeconds {
		return EpochPhaseClosed, nil
	}
	durations := params.EffectivePhaseDurations()
	if elapsed < durations.DelegationSeconds {
		return EpochPhaseDelegation, nil
	}
	elapsed -= durations.DelegationSeconds
	if elapsed < durations.ElectionSeconds {
		return EpochPhaseElection, nil
	}
	elapsed -= durations.ElectionSeconds
	if elapsed < durations.AssignmentSeconds {
		return EpochPhaseAssignment, nil
	}
	elapsed -= durations.AssignmentSeconds
	if elapsed < durations.ActiveValidationSeconds {
		return EpochPhaseActive, nil
	}
	return EpochPhaseSettlement, nil
}

func NewEpochRecord(params Params, epochID uint64, startHeight uint64, endHeight uint64, phase EpochPhase, previousSeed string, validators []ScoredValidator) (EpochRecord, error) {
	if err := params.Validate(); err != nil {
		return EpochRecord{}, err
	}
	if startHeight == 0 || endHeight < startHeight {
		return EpochRecord{}, errors.New("epoch heights must be positive and ordered")
	}
	if err := validateEpochPhase(phase); err != nil {
		return EpochRecord{}, err
	}
	validatorSetHash, err := ComputeValidatorSetHash(validators)
	if err != nil {
		return EpochRecord{}, err
	}
	seed, err := DeriveEpochSeedWithSource(params.EffectiveEpochSeedSource(), epochID, startHeight, previousSeed, validatorSetHash)
	if err != nil {
		return EpochRecord{}, err
	}
	record := EpochRecord{
		EpochID:		epochID,
		StartHeight:		startHeight,
		EndHeight:		endHeight,
		Phase:			phase,
		Seed:			seed,
		ValidatorSetHash:	validatorSetHash,
		TaskGroupRoot:		PosEmptyRootHash,
		PerformanceRoot:	PosEmptyRootHash,
		RewardRoot:		PosEmptyRootHash,
		SlashRoot:		PosEmptyRootHash,
		SettlementStatus:	SettlementStatusPending,
	}
	if err := record.Validate(); err != nil {
		return EpochRecord{}, err
	}
	return record, nil
}

func CloseEpochRecord(record EpochRecord, performanceRoot string, rewardRoot string, slashRoot string) (EpochRecord, error) {
	if err := record.Validate(); err != nil {
		return EpochRecord{}, err
	}
	if record.Phase != EpochPhaseSettlement {
		return EpochRecord{}, errors.New("epoch must be in settlement phase before closing")
	}
	if err := validatePosHash("performance root", performanceRoot); err != nil {
		return EpochRecord{}, err
	}
	if err := validatePosHash("reward root", rewardRoot); err != nil {
		return EpochRecord{}, err
	}
	if err := validatePosHash("slash root", slashRoot); err != nil {
		return EpochRecord{}, err
	}
	record.Phase = EpochPhaseClosed
	record.PerformanceRoot = performanceRoot
	record.RewardRoot = rewardRoot
	record.SlashRoot = slashRoot
	record.SettlementStatus = SettlementStatusFinalized
	return record, record.Validate()
}

func (r EpochRecord) Validate() error {
	if r.StartHeight == 0 || r.EndHeight < r.StartHeight {
		return errors.New("epoch heights must be positive and ordered")
	}
	if err := validateEpochPhase(r.Phase); err != nil {
		return err
	}
	if err := validatePosHash("epoch seed", r.Seed); err != nil {
		return err
	}
	if err := validatePosHash("validator set hash", r.ValidatorSetHash); err != nil {
		return err
	}
	if err := validatePosHash("task group root", r.TaskGroupRoot); err != nil {
		return err
	}
	if err := validatePosHash("performance root", r.PerformanceRoot); err != nil {
		return err
	}
	if err := validatePosHash("reward root", r.RewardRoot); err != nil {
		return err
	}
	if err := validatePosHash("slash root", r.SlashRoot); err != nil {
		return err
	}
	switch r.SettlementStatus {
	case SettlementStatusPending, SettlementStatusFinalized:
	default:
		return errors.New("unsupported settlement status")
	}
	if r.Phase == EpochPhaseClosed && r.SettlementStatus != SettlementStatusFinalized {
		return errors.New("closed epoch must have finalized settlement")
	}
	if r.SettlementStatus == SettlementStatusFinalized && r.Phase != EpochPhaseClosed {
		return errors.New("finalized settlement must close the epoch")
	}
	return nil
}

func ComputeValidatorSetHash(validators []ScoredValidator) (string, error) {
	ordered := cloneScoredValidators(validators)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ValidatorID < ordered[j].ValidatorID
	})
	seen := make(map[string]struct{}, len(ordered))
	for _, validator := range ordered {
		if err := validatePosToken("validator id", validator.ValidatorID); err != nil {
			return "", err
		}
		if validator.VotingPowerNaet.IsNegative() {
			return "", errors.New("validator voting power cannot be negative")
		}
		if validator.Score.IsNegative() {
			return "", errors.New("validator score cannot be negative")
		}
		if _, found := seen[validator.ValidatorID]; found {
			return "", fmt.Errorf("duplicate validator %q", validator.ValidatorID)
		}
		seen[validator.ValidatorID] = struct{}{}
		if err := validateValidatorRoles(validator.Roles); err != nil {
			return "", err
		}
	}
	return posHashRoot("aetheris-pos-validator-set-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(ordered)))
		for _, validator := range ordered {
			posWritePart(w, validator.ValidatorID)
			posWritePart(w, validator.VotingPowerNaet.String())
			posWritePart(w, validator.Score.String())
			for _, role := range normalizedRoles(validator.Roles, AllValidatorRoles()) {
				posWritePart(w, string(role))
			}
		}
	}), nil
}

func DeriveEpochSeed(epochID uint64, startHeight uint64, previousSeed string, validatorSetHash string) (string, error) {
	return DeriveEpochSeedWithSource(EpochSeedSourcePreviousSeedValidatorSet, epochID, startHeight, previousSeed, validatorSetHash)
}

func DeriveEpochSeedWithSource(source EpochSeedSource, epochID uint64, startHeight uint64, previousSeed string, validatorSetHash string) (string, error) {
	if err := source.Validate(); err != nil {
		return "", err
	}
	if startHeight == 0 {
		return "", errors.New("epoch seed start height must be positive")
	}
	if previousSeed == "" {
		previousSeed = PosEmptyRootHash
	}
	if err := validatePosHash("previous epoch seed", previousSeed); err != nil {
		return "", err
	}
	if err := validatePosHash("validator set hash", validatorSetHash); err != nil {
		return "", err
	}
	return posHashRoot("aetheris-pos-epoch-seed-v1", func(w posByteWriter) {
		posWritePart(w, string(source))
		posWriteUint64(w, epochID)
		posWriteUint64(w, startHeight)
		posWritePart(w, previousSeed)
		posWritePart(w, validatorSetHash)
	}), nil
}

func BuildTaskAssignments(params Params, epoch EpochRecord, validators []ScoredValidator, tasks []WorkloadTask) (TaskAssignmentSet, error) {
	if err := params.Validate(); err != nil {
		return TaskAssignmentSet{}, err
	}
	if err := epoch.Validate(); err != nil {
		return TaskAssignmentSet{}, err
	}
	if len(validators) == 0 {
		return TaskAssignmentSet{}, errors.New("task assignment requires active validators")
	}
	validatorSetHash, err := ComputeValidatorSetHash(validators)
	if err != nil {
		return TaskAssignmentSet{}, err
	}
	if validatorSetHash != epoch.ValidatorSetHash {
		return TaskAssignmentSet{}, errors.New("task assignments require committed validator set hash")
	}
	if len(tasks) == 0 {
		return TaskAssignmentSet{
			EpochID:	epoch.EpochID,
			Seed:		epoch.Seed,
			Root:		PosEmptyRootHash,
		}, nil
	}

	orderedTasks := make([]WorkloadTask, len(tasks))
	for i, task := range tasks {
		normalized := normalizeWorkloadTask(params, task)
		if err := normalized.Validate(params); err != nil {
			return TaskAssignmentSet{}, err
		}
		orderedTasks[i] = normalized
	}
	sort.SliceStable(orderedTasks, func(i, j int) bool {
		return compareWorkloadTasks(orderedTasks[i], orderedTasks[j]) < 0
	})

	assignments := make([]TaskAssignment, 0)
	assignedTaskKeys := make(map[string]map[string]struct{})
	for _, task := range orderedTasks {
		key := taskKey(task)
		for _, role := range normalizedRoles(task.Roles, DefaultTaskRoles()) {
			eligible := validatorsForTaskRole(validators, task, role, assignedTaskKeys, key)
			if uint32(len(eligible)) < task.RequiredValidators {
				return TaskAssignmentSet{}, fmt.Errorf("insufficient validators for task %s role %s", task.TaskID, role)
			}
			selected := selectTaskValidatorIDs(epoch.Seed, task, role, eligible, task.RequiredValidators)
			markTaskGroupAssignments(assignedTaskKeys, key, selected)
			assignment := TaskAssignment{
				TaskID:		task.TaskID,
				WorkloadID:	task.WorkloadID,
				WorkloadType:	task.WorkloadType,
				ZoneID:		task.ZoneID,
				ShardID:	task.ShardID,
				WorkloadClass:	task.WorkloadClass,
				Role:		role,
				Validators:	selected,
			}
			assignment.AssignmentHash = ComputeTaskAssignmentHash(epoch.EpochID, epoch.Seed, assignment)
			assignments = append(assignments, assignment)
		}
	}
	sort.SliceStable(assignments, func(i, j int) bool {
		return compareTaskAssignments(assignments[i], assignments[j]) < 0
	})
	root := ComputeTaskAssignmentRoot(epoch.EpochID, epoch.Seed, assignments)
	out := TaskAssignmentSet{EpochID: epoch.EpochID, Seed: epoch.Seed, Assignments: assignments, Root: root}
	return out, out.Validate()
}

func BuildTaskGroups(params Params, epoch EpochRecord, validators []ScoredValidator, tasks []WorkloadTask, activationHeight uint64, expiryHeight uint64) (TaskGroupSet, error) {
	if activationHeight == 0 {
		return TaskGroupSet{}, errors.New("task group activation height is required")
	}
	if expiryHeight <= activationHeight {
		return TaskGroupSet{}, errors.New("task group expiry height must be after activation height")
	}
	assignments, err := BuildTaskAssignments(params, epoch, validators, tasks)
	if err != nil {
		return TaskGroupSet{}, err
	}
	if len(tasks) == 0 {
		return TaskGroupSet{EpochID: epoch.EpochID, Seed: epoch.Seed, Root: PosEmptyRootHash}, nil
	}
	validatorByID := make(map[string]ScoredValidator, len(validators))
	for _, validator := range validators {
		validatorByID[validator.ValidatorID] = validator
	}
	taskByID := make(map[string]WorkloadTask, len(tasks))
	for _, task := range tasks {
		normalized := normalizeWorkloadTask(params, task)
		taskByID[taskKey(normalized)] = normalized
	}
	assignmentsByTask := make(map[string][]TaskAssignment)
	for _, assignment := range assignments.Assignments {
		key := taskKey(WorkloadTask{
			TaskID:		assignment.TaskID,
			WorkloadID:	assignment.WorkloadID,
			WorkloadType:	assignment.WorkloadType,
			ZoneID:		assignment.ZoneID,
			ShardID:	assignment.ShardID,
			WorkloadClass:	assignment.WorkloadClass,
		})
		assignmentsByTask[key] = append(assignmentsByTask[key], assignment)
	}
	groups := make([]TaskGroup, 0, len(taskByID))
	taskKeys := sortedStringKeys(taskByID)
	for _, key := range taskKeys {
		task := taskByID[key]
		taskAssignments := assignmentsByTask[key]
		members := taskGroupMembers(taskAssignments)
		verifiers := taskGroupVerifiers(taskAssignments, members)
		group := TaskGroup{
			EpochID:		epoch.EpochID,
			WorkloadID:		task.WorkloadID,
			WorkloadType:		task.WorkloadType,
			ValidatorMembers:	members,
			ProposerOrder:		taskGroupProposerOrder(epoch.Seed, task, members),
			VerifierSet:		verifiers,
			MinimumGroupSize:	task.RequiredValidators,
			StakeWeightRoot:	ComputeTaskGroupStakeWeightRoot(epoch.EpochID, task, members, validatorByID),
			AssignmentSeed:		epoch.Seed,
			ActivationHeight:	activationHeight,
			ExpiryHeight:		expiryHeight,
		}
		group.TaskGroupID = ComputeTaskGroupID(group)
		groups = append(groups, group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return compareTaskGroups(groups[i], groups[j]) < 0
	})
	out := TaskGroupSet{
		EpochID:	epoch.EpochID,
		Seed:		epoch.Seed,
		Groups:		groups,
		Root:		ComputeTaskGroupRoot(epoch.EpochID, epoch.Seed, groups),
	}
	return out, out.Validate()
}

func (t WorkloadTask) Validate(params Params) error {
	if err := validatePosToken("task id", t.TaskID); err != nil {
		return err
	}
	if err := validatePosToken("task workload id", t.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(t.WorkloadType); err != nil {
		return err
	}
	if err := validatePosToken("task zone id", t.ZoneID); err != nil {
		return err
	}
	if err := validatePosToken("task shard id", t.ShardID); err != nil {
		return err
	}
	if err := validatePosToken("task workload class", t.WorkloadClass); err != nil {
		return err
	}
	if t.RequiredValidators < params.MinTaskGroupValidators {
		return fmt.Errorf("task validators must be at least %d", params.MinTaskGroupValidators)
	}
	if t.RequiredValidators > params.MaxTaskGroupValidators {
		return fmt.Errorf("task validators must be <= %d", params.MaxTaskGroupValidators)
	}
	if err := validateValidatorRoles(t.Roles); err != nil {
		return err
	}
	return validateExcludedValidators(t.ExcludedValidators)
}

func (c ValidatorCapacity) Validate() error {
	if c.MaxTaskGroups == 0 && len(c.SupportedWorkloads) == 0 && len(c.ZoneSupport) == 0 && c.HardwareClassOptional == "" && c.NetworkClassOptional == "" && c.AvailabilityCommitment == 0 {
		return nil
	}
	if c.MaxTaskGroups == 0 {
		return errors.New("validator capacity max task groups must be positive when capacity is declared")
	}
	if c.AvailabilityCommitment > BasisPoints {
		return fmt.Errorf("validator availability commitment must be <= %d bps", BasisPoints)
	}
	if err := validateWorkloadTypes(c.SupportedWorkloads); err != nil {
		return err
	}
	if err := validateZoneSupport(c.ZoneSupport); err != nil {
		return err
	}
	if c.HardwareClassOptional != "" {
		if err := validatePosToken("hardware class", c.HardwareClassOptional); err != nil {
			return err
		}
	}
	if c.NetworkClassOptional != "" {
		if err := validatePosToken("network class", c.NetworkClassOptional); err != nil {
			return err
		}
	}
	return nil
}

func (c ValidatorCapacity) SupportsAssignment(task WorkloadTask, assignedTaskKeys map[string]map[string]struct{}, validatorID string, taskKey string) bool {
	if !c.supportsWorkload(task.WorkloadType) || !c.supportsZone(task.ZoneID) {
		return false
	}
	if c.MaxTaskGroups == 0 {
		return true
	}
	current := assignedTaskKeys[validatorID]
	if _, alreadyAssignedToTask := current[taskKey]; alreadyAssignedToTask {
		return true
	}
	return uint32(len(current)) < c.MaxTaskGroups
}

func (c ValidatorCapacity) supportsWorkload(workloadType WorkloadType) bool {
	if len(c.SupportedWorkloads) == 0 {
		return true
	}
	for _, supported := range c.SupportedWorkloads {
		if supported == workloadType {
			return true
		}
	}
	return false
}

func (c ValidatorCapacity) supportsZone(zoneID string) bool {
	if len(c.ZoneSupport) == 0 {
		return true
	}
	for _, zone := range c.ZoneSupport {
		if zone == zoneID {
			return true
		}
	}
	return false
}

func (e CapacityFaultEvidence) Validate() error {
	if err := validatePosToken("capacity evidence validator id", e.ValidatorID); err != nil {
		return err
	}
	if err := validatePosToken("capacity evidence workload id", e.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(e.WorkloadType); err != nil {
		return err
	}
	if e.AssignmentEpoch == 0 {
		return errors.New("capacity evidence assignment epoch is required")
	}
	if e.EvidenceHeight < 0 {
		return errors.New("capacity evidence height cannot be negative")
	}
	return nil
}

func IsSlashableCapacityFault(evidence CapacityFaultEvidence) (bool, error) {
	if err := evidence.Validate(); err != nil {
		return false, err
	}
	return evidence.Finalized && evidence.UsedForAssignment, nil
}

func NewEvidenceRecord(record EvidenceRecord) (EvidenceRecord, error) {
	record.EvidenceID = strings.TrimSpace(record.EvidenceID)
	record.EvidenceType = strings.TrimSpace(record.EvidenceType)
	record.AccusedValidator = strings.TrimSpace(record.AccusedValidator)
	record.Reporter = strings.TrimSpace(record.Reporter)
	record.TaskGroupIDOptional = strings.TrimSpace(record.TaskGroupIDOptional)
	record.ObjectHash = strings.TrimSpace(record.ObjectHash)
	record.ProofPayloadHash = strings.TrimSpace(record.ProofPayloadHash)
	record.VerificationGroupID = strings.TrimSpace(record.VerificationGroupID)
	record.PenaltyIDOptional = strings.TrimSpace(record.PenaltyIDOptional)
	if record.Status == "" {
		record.Status = EvidenceStatusSubmitted
	}
	return record, record.Validate()
}

func EvidenceRecordFieldNames() []string {
	return []string{
		"evidence_id",
		"evidence_type",
		"accused_validator",
		"reporter",
		"epoch_id",
		"task_group_id_optional",
		"object_hash",
		"proof_payload_hash",
		"submitted_height",
		"status",
		"verification_group_id",
		"decision_height",
		"penalty_id_optional",
	}
}

func EvidenceRecordStatusValues() []string {
	return []string{
		EvidenceStatusSubmitted,
		EvidenceStatusInVerification,
		EvidenceStatusAccepted,
		EvidenceStatusRejected,
		EvidenceStatusExpired,
		EvidenceStatusSlashed,
	}
}

func (e EvidenceRecord) Validate() error {
	if err := validatePosToken("evidence record id", e.EvidenceID); err != nil {
		return err
	}
	if !IsStructuredEvidenceType(e.EvidenceType) {
		return fmt.Errorf("unsupported evidence record type %q", e.EvidenceType)
	}
	if err := validatePosToken("evidence record accused validator", e.AccusedValidator); err != nil {
		return err
	}
	if err := validatePosToken("evidence record reporter", e.Reporter); err != nil {
		return err
	}
	if e.EpochID == 0 {
		return errors.New("evidence record epoch id is required")
	}
	if e.TaskGroupIDOptional != "" {
		if err := validatePosToken("evidence record task group id", e.TaskGroupIDOptional); err != nil {
			return err
		}
	}
	if err := validatePosHash("evidence record object hash", e.ObjectHash); err != nil {
		return err
	}
	if err := validatePosHash("evidence record proof payload hash", e.ProofPayloadHash); err != nil {
		return err
	}
	if e.SubmittedHeight < 0 {
		return errors.New("evidence record submitted height cannot be negative")
	}
	if !isEvidenceRecordStatus(e.Status) {
		return fmt.Errorf("unsupported evidence record status %q", e.Status)
	}
	if e.VerificationGroupID != "" {
		if err := validatePosToken("evidence record verification group id", e.VerificationGroupID); err != nil {
			return err
		}
	}
	if (e.Status == EvidenceStatusInVerification || e.Status == EvidenceStatusAccepted || e.Status == EvidenceStatusRejected || e.Status == EvidenceStatusSlashed) && e.VerificationGroupID == "" {
		return errors.New("evidence record status requires verification group id")
	}
	if e.DecisionHeight < 0 {
		return errors.New("evidence record decision height cannot be negative")
	}
	if e.PenaltyIDOptional != "" {
		if err := validatePosToken("evidence record penalty id", e.PenaltyIDOptional); err != nil {
			return err
		}
	}
	if e.Status == EvidenceStatusSlashed && e.PenaltyIDOptional == "" {
		return errors.New("slashed evidence record requires penalty id")
	}
	if (e.Status == EvidenceStatusAccepted || e.Status == EvidenceStatusRejected || e.Status == EvidenceStatusExpired || e.Status == EvidenceStatusSlashed) && e.DecisionHeight == 0 {
		return errors.New("decided evidence record requires decision height")
	}
	return nil
}

func AssignEvidenceVerificationGroup(record EvidenceRecord, group EvidenceVerificationGroup) (EvidenceRecord, error) {
	if err := record.Validate(); err != nil {
		return EvidenceRecord{}, err
	}
	if err := group.Validate(); err != nil {
		return EvidenceRecord{}, err
	}
	if record.EvidenceID != group.EvidenceID {
		return EvidenceRecord{}, errors.New("evidence verification group record id mismatch")
	}
	if record.EpochID != group.EpochID {
		return EvidenceRecord{}, errors.New("evidence verification group record epoch mismatch")
	}
	next := record
	next.VerificationGroupID = group.VerificationGroupID
	next.Status = EvidenceStatusInVerification
	next.DecisionHeight = 0
	next.PenaltyIDOptional = ""
	return next, next.Validate()
}

func AdvanceEvidenceRecordStatus(record EvidenceRecord, status string, decisionHeight int64, penaltyIDOptional string) (EvidenceRecord, error) {
	if err := record.Validate(); err != nil {
		return EvidenceRecord{}, err
	}
	if !isEvidenceRecordStatus(status) {
		return EvidenceRecord{}, fmt.Errorf("unsupported evidence record status %q", status)
	}
	if !isAllowedEvidenceRecordTransition(record.Status, status) {
		return EvidenceRecord{}, fmt.Errorf("invalid evidence record status transition %s -> %s", record.Status, status)
	}
	next := record
	next.Status = status
	next.DecisionHeight = decisionHeight
	next.PenaltyIDOptional = strings.TrimSpace(penaltyIDOptional)
	return next, next.Validate()
}

func SelectEvidenceVerificationGroup(input EvidenceVerificationGroupInput) (EvidenceVerificationGroup, error) {
	if err := input.Params.Validate(); err != nil {
		return EvidenceVerificationGroup{}, err
	}
	if err := input.Epoch.Validate(); err != nil {
		return EvidenceVerificationGroup{}, err
	}
	if err := input.Evidence.Validate(); err != nil {
		return EvidenceVerificationGroup{}, err
	}
	if input.Evidence.EpochID != input.Epoch.EpochID {
		return EvidenceVerificationGroup{}, errors.New("evidence record epoch does not match verification epoch")
	}
	minimum := input.MinimumGroupSize
	if minimum == 0 {
		minimum = input.Params.MinTaskGroupValidators
	}
	if minimum == 0 {
		return EvidenceVerificationGroup{}, errors.New("evidence verification group minimum size is required")
	}
	threshold := input.DecisionThresholdBps
	if threshold == 0 {
		threshold = DefaultEvidenceVerificationQuorumBps
	}
	if threshold > BasisPoints {
		return EvidenceVerificationGroup{}, fmt.Errorf("evidence decision threshold must be <= %d bps", BasisPoints)
	}
	excluded := evidenceVerificationExclusions(input.Evidence, input.ActiveValidators)
	eligible := make([]string, 0, len(input.ActiveValidators))
	seen := make(map[string]struct{}, len(input.ActiveValidators))
	for _, validator := range input.ActiveValidators {
		validatorID := strings.TrimSpace(validator.ValidatorID)
		if validatorID == "" {
			return EvidenceVerificationGroup{}, errors.New("active validator id is required")
		}
		if _, duplicate := seen[validatorID]; duplicate {
			return EvidenceVerificationGroup{}, fmt.Errorf("duplicate active validator %q", validatorID)
		}
		seen[validatorID] = struct{}{}
		if isExcludedValidator(validatorID, excluded) {
			continue
		}
		eligible = append(eligible, validatorID)
	}
	if uint32(len(eligible)) < minimum {
		return EvidenceVerificationGroup{}, fmt.Errorf("insufficient eligible validators for evidence verification group: need %d got %d", minimum, len(eligible))
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		left := computeEvidenceVerifierSelectionHash(input.Epoch.Seed, input.Evidence.EvidenceID, eligible[i])
		right := computeEvidenceVerifierSelectionHash(input.Epoch.Seed, input.Evidence.EvidenceID, eligible[j])
		if left != right {
			return left < right
		}
		return eligible[i] < eligible[j]
	})
	members := cloneStringSlice(eligible[:minimum])
	sort.Strings(members)
	sort.Strings(excluded)
	group := EvidenceVerificationGroup{
		EvidenceID:		input.Evidence.EvidenceID,
		EpochID:		input.Evidence.EpochID,
		Members:		members,
		ExcludedValidators:	excluded,
		MinimumGroupSize:	minimum,
		DecisionThresholdBps:	threshold,
		AssignmentSeed:		computeEvidenceVerificationAssignmentSeed(input.Epoch.Seed, input.Evidence.EvidenceID),
	}
	group.VerificationGroupID = computeEvidenceVerificationGroupID(group)
	group.GroupHash = computeEvidenceVerificationGroupHash(group)
	return group, group.Validate()
}

func (g EvidenceVerificationGroup) Validate() error {
	if err := validatePosToken("evidence verification group evidence id", g.EvidenceID); err != nil {
		return err
	}
	if g.EpochID == 0 {
		return errors.New("evidence verification group epoch id is required")
	}
	if err := validatePosToken("evidence verification group id", g.VerificationGroupID); err != nil {
		return err
	}
	if len(g.Members) == 0 {
		return errors.New("evidence verification group members are required")
	}
	if g.MinimumGroupSize == 0 || uint32(len(g.Members)) < g.MinimumGroupSize {
		return errors.New("evidence verification group minimum size is not met")
	}
	if g.DecisionThresholdBps == 0 || g.DecisionThresholdBps > BasisPoints {
		return fmt.Errorf("evidence verification decision threshold must be within 1..%d bps", BasisPoints)
	}
	if err := validateSortedUniqueTokens("evidence verification group member", g.Members); err != nil {
		return err
	}
	if err := validateSortedUniqueTokens("evidence verification group exclusion", g.ExcludedValidators); err != nil {
		return err
	}
	for _, member := range g.Members {
		if isExcludedValidator(member, g.ExcludedValidators) {
			return fmt.Errorf("evidence verification group member %q is excluded", member)
		}
	}
	if err := validatePosHash("evidence verification group assignment seed", g.AssignmentSeed); err != nil {
		return err
	}
	expectedID := computeEvidenceVerificationGroupID(g)
	if g.VerificationGroupID != expectedID {
		return errors.New("evidence verification group id mismatch")
	}
	expectedHash := computeEvidenceVerificationGroupHash(g)
	if g.GroupHash != expectedHash {
		return errors.New("evidence verification group hash mismatch")
	}
	return nil
}

func StructuredEvidenceTypes() []string {
	return []string{
		EvidenceTypeDoubleSignProof,
		EvidenceTypeInvalidStateTransitionProof,
		EvidenceTypeEquivocationProof,
		EvidenceTypeDowntimeProof,
		EvidenceTypeInvalidTaskExecutionProof,
		EvidenceTypeInvalidCollatorOutputProof,
		EvidenceTypeInvalidProofAcceptance,
		EvidenceTypeFalseCapacityDeclaration,
		EvidenceTypeInvalidEvidenceSubmission,
	}
}

func IsStructuredEvidenceType(evidenceType string) bool {
	switch evidenceType {
	case EvidenceTypeDoubleSignProof,
		EvidenceTypeInvalidStateTransitionProof,
		EvidenceTypeEquivocationProof,
		EvidenceTypeDowntimeProof,
		EvidenceTypeInvalidTaskExecutionProof,
		EvidenceTypeInvalidCollatorOutputProof,
		EvidenceTypeInvalidProofAcceptance,
		EvidenceTypeFalseCapacityDeclaration,
		EvidenceTypeInvalidEvidenceSubmission:
		return true
	default:
		return false
	}
}

func DefaultEvidenceSlashPolicy(evidenceType string) (EvidenceSlashPolicy, error) {
	switch evidenceType {
	case EvidenceTypeDoubleSignProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorDoubleSign, SlashFractionBps: DefaultDoubleSignSlashBps}, nil
	case EvidenceTypeInvalidStateTransitionProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultInvalidStateTransitionSlashBps}, nil
	case EvidenceTypeEquivocationProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorDoubleSign, SlashFractionBps: DefaultEquivocationSlashBps}, nil
	case EvidenceTypeDowntimeProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorDowntime, SlashFractionBps: DefaultDowntimeSlashBps}, nil
	case EvidenceTypeInvalidTaskExecutionProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultInvalidTaskExecutionSlashBps}, nil
	case EvidenceTypeInvalidCollatorOutputProof:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultInvalidCollatorOutputSlashBps}, nil
	case EvidenceTypeInvalidProofAcceptance:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultInvalidProofAcceptanceSlashBps}, nil
	case EvidenceTypeFalseCapacityDeclaration:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultFalseCapacityDeclarationSlashBps}, nil
	case EvidenceTypeInvalidEvidenceSubmission:
		return EvidenceSlashPolicy{EvidenceType: evidenceType, Misbehavior: MisbehaviorInvalidBlock, SlashFractionBps: DefaultInvalidEvidenceSubmissionSlashBps}, nil
	default:
		return EvidenceSlashPolicy{}, fmt.Errorf("unsupported structured evidence type %q", evidenceType)
	}
}

func SubmitStructuredEvidence(evidence StructuredEvidenceRecord) (StructuredEvidenceRecord, error) {
	evidence.EvidenceID = strings.TrimSpace(evidence.EvidenceID)
	evidence.EvidenceType = strings.TrimSpace(evidence.EvidenceType)
	evidence.ReporterID = strings.TrimSpace(evidence.ReporterID)
	evidence.AccusedValidatorID = strings.TrimSpace(evidence.AccusedValidatorID)
	evidence.SubjectID = strings.TrimSpace(evidence.SubjectID)
	evidence.EvidenceHash = strings.TrimSpace(evidence.EvidenceHash)
	evidence.VerificationGroupID = strings.TrimSpace(evidence.VerificationGroupID)
	evidence.Status = EvidenceStatusSubmitted
	evidence.StructuredRecordHash = computeStructuredEvidenceHash(evidence)
	return evidence, evidence.Validate()
}

func (e StructuredEvidenceRecord) Validate() error {
	if err := validatePosToken("structured evidence id", e.EvidenceID); err != nil {
		return err
	}
	if !IsStructuredEvidenceType(e.EvidenceType) {
		return fmt.Errorf("unsupported structured evidence type %q", e.EvidenceType)
	}
	if err := validatePosToken("structured evidence reporter id", e.ReporterID); err != nil {
		return err
	}
	if err := validatePosToken("structured evidence accused validator id", e.AccusedValidatorID); err != nil {
		return err
	}
	if err := validatePosToken("structured evidence subject id", e.SubjectID); err != nil {
		return err
	}
	if err := validatePosHash("structured evidence hash", e.EvidenceHash); err != nil {
		return err
	}
	if e.EvidenceHeight < 0 {
		return errors.New("structured evidence height cannot be negative")
	}
	if e.EvidenceEpoch == 0 {
		return errors.New("structured evidence epoch is required")
	}
	if e.SubmittedHeight < 0 {
		return errors.New("structured evidence submitted height cannot be negative")
	}
	if err := validatePosToken("structured evidence verification group id", e.VerificationGroupID); err != nil {
		return err
	}
	if !isEvidenceStatus(e.Status) {
		return fmt.Errorf("unsupported structured evidence status %q", e.Status)
	}
	expectedHash := computeStructuredEvidenceHash(e)
	if e.StructuredRecordHash != expectedHash {
		return errors.New("structured evidence record hash mismatch")
	}
	return nil
}

func VerifyStructuredEvidenceBySubset(evidence StructuredEvidenceRecord, reviewers []string, votes []EvidenceVerificationVote, quorumBps uint32) (EvidenceVerificationResult, error) {
	if quorumBps == 0 {
		quorumBps = DefaultEvidenceVerificationQuorumBps
	}
	if quorumBps > BasisPoints {
		return EvidenceVerificationResult{}, fmt.Errorf("evidence verification quorum must be <= %d bps", BasisPoints)
	}
	if err := evidence.Validate(); err != nil {
		return EvidenceVerificationResult{}, err
	}
	if evidence.Status != EvidenceStatusSubmitted && evidence.Status != EvidenceStatusVerified {
		return EvidenceVerificationResult{}, errors.New("structured evidence must be submitted before subset verification")
	}
	reviewerSet, err := validateEvidenceReviewers(reviewers)
	if err != nil {
		return EvidenceVerificationResult{}, err
	}
	seen := make(map[string]struct{}, len(votes))
	accepted := uint32(0)
	rejected := uint32(0)
	for _, vote := range votes {
		if err := vote.Validate(evidence.EvidenceID); err != nil {
			return EvidenceVerificationResult{}, err
		}
		if _, assigned := reviewerSet[vote.ReviewerID]; !assigned {
			return EvidenceVerificationResult{}, fmt.Errorf("evidence reviewer %q is not assigned to verification subset", vote.ReviewerID)
		}
		if _, found := seen[vote.ReviewerID]; found {
			return EvidenceVerificationResult{}, fmt.Errorf("duplicate evidence verification vote from %q", vote.ReviewerID)
		}
		seen[vote.ReviewerID] = struct{}{}
		if vote.Accepted {
			accepted++
		} else {
			rejected++
		}
	}
	totalReviewers := uint32(len(reviewerSet))
	acceptedBps := ratioBps(uint64(accepted), uint64(totalReviewers))
	rejectedBps := ratioBps(uint64(rejected), uint64(totalReviewers))
	result := EvidenceVerificationResult{
		EvidenceID:		evidence.EvidenceID,
		AcceptedVotes:		accepted,
		RejectedVotes:		rejected,
		TotalReviewers:		totalReviewers,
		ParticipationBps:	ratioBps(uint64(len(votes)), uint64(totalReviewers)),
		QuorumBps:		quorumBps,
		Accepted:		acceptedBps >= quorumBps,
		Rejected:		rejectedBps >= quorumBps,
		Status:			EvidenceStatusSubmitted,
		VerificationGroup:	evidence.VerificationGroupID,
	}
	if result.Accepted {
		result.Status = EvidenceStatusVerified
	} else if result.Rejected {
		result.Status = EvidenceStatusRejected
	}
	result.VerificationRoot = computeEvidenceVerificationRoot(evidence.EvidenceID, votes)
	return result, nil
}

func (v EvidenceVerificationVote) Validate(expectedEvidenceID string) error {
	if v.EvidenceID != expectedEvidenceID {
		return errors.New("evidence verification vote id mismatch")
	}
	if err := validatePosToken("evidence verification reviewer id", v.ReviewerID); err != nil {
		return err
	}
	if err := validatePosHash("evidence verification signature hash", v.SignatureHash); err != nil {
		return err
	}
	if v.VoteHeight < 0 {
		return errors.New("evidence verification vote height cannot be negative")
	}
	return nil
}

func FinalizeStructuredEvidence(evidence StructuredEvidenceRecord, verification EvidenceVerificationResult, votes []EvidenceFinalityVote, quorumBps uint32) (EvidenceFinalityDecision, error) {
	if quorumBps == 0 {
		quorumBps = DefaultEvidenceFinalityQuorumBps
	}
	if quorumBps > BasisPoints {
		return EvidenceFinalityDecision{}, fmt.Errorf("evidence finality quorum must be <= %d bps", BasisPoints)
	}
	if err := evidence.Validate(); err != nil {
		return EvidenceFinalityDecision{}, err
	}
	if verification.EvidenceID != evidence.EvidenceID {
		return EvidenceFinalityDecision{}, errors.New("evidence finality verification id mismatch")
	}
	if !verification.Accepted || verification.Status != EvidenceStatusVerified {
		return EvidenceFinalityDecision{}, errors.New("evidence must be verified by consensus subset before finality vote")
	}
	seen := make(map[string]struct{}, len(votes))
	acceptedPower := uint64(0)
	rejectedPower := uint64(0)
	totalPower := uint64(0)
	for _, vote := range votes {
		if err := vote.Validate(evidence.EvidenceID); err != nil {
			return EvidenceFinalityDecision{}, err
		}
		if _, found := seen[vote.ValidatorID]; found {
			return EvidenceFinalityDecision{}, fmt.Errorf("duplicate evidence finality vote from %q", vote.ValidatorID)
		}
		seen[vote.ValidatorID] = struct{}{}
		totalPower += uint64(vote.VotingPowerBps)
		if totalPower > uint64(BasisPoints) {
			return EvidenceFinalityDecision{}, fmt.Errorf("evidence finality voting power must be <= %d bps", BasisPoints)
		}
		if vote.Approve {
			acceptedPower += uint64(vote.VotingPowerBps)
		} else {
			rejectedPower += uint64(vote.VotingPowerBps)
		}
	}
	decision := EvidenceFinalityDecision{
		EvidenceID:		evidence.EvidenceID,
		AcceptedPowerBps:	uint32(acceptedPower),
		RejectedPowerBps:	uint32(rejectedPower),
		QuorumBps:		quorumBps,
		FinalityVoteRoot:	computeEvidenceFinalityVoteRoot(evidence.EvidenceID, votes),
		FinalityVoteCount:	uint32(len(votes)),
		Status:			EvidenceStatusVerified,
	}
	if uint32(acceptedPower) >= quorumBps {
		decision.Finalized = true
		decision.Accepted = true
		decision.Status = EvidenceStatusFinalized
	} else if uint32(rejectedPower) >= quorumBps {
		decision.Finalized = true
		decision.Status = EvidenceStatusRejected
	}
	return decision, nil
}

func (v EvidenceFinalityVote) Validate(expectedEvidenceID string) error {
	if v.EvidenceID != expectedEvidenceID {
		return errors.New("evidence finality vote id mismatch")
	}
	if err := validatePosToken("evidence finality validator id", v.ValidatorID); err != nil {
		return err
	}
	if v.VotingPowerBps == 0 || v.VotingPowerBps > BasisPoints {
		return fmt.Errorf("evidence finality voting power must be within 1..%d bps", BasisPoints)
	}
	if err := validatePosHash("evidence finality signature hash", v.SignatureHash); err != nil {
		return err
	}
	if v.FinalityHeight < 0 {
		return errors.New("evidence finality vote height cannot be negative")
	}
	return nil
}

func ExecuteStructuredEvidenceSlashing(params Params, currentEpoch uint64, evidence StructuredEvidenceRecord, decision EvidenceFinalityDecision, selfStake sdkmath.Int, nominations []Nomination) (EvidenceSettlement, error) {
	if decision.EvidenceID != evidence.EvidenceID {
		return EvidenceSettlement{}, errors.New("evidence slashing decision id mismatch")
	}
	if !decision.Finalized || !decision.Accepted || decision.Status != EvidenceStatusFinalized {
		return EvidenceSettlement{}, errors.New("evidence must have accepted finality before slashing")
	}
	if err := evidence.Validate(); err != nil {
		return EvidenceSettlement{}, err
	}
	policy, err := DefaultEvidenceSlashPolicy(evidence.EvidenceType)
	if err != nil {
		return EvidenceSettlement{}, err
	}
	return SettleEvidenceCase(params, currentEpoch, EvidenceCase{
		EvidenceID:		evidence.EvidenceID,
		ReporterID:		evidence.ReporterID,
		ValidatorID:		evidence.AccusedValidatorID,
		Misbehavior:		policy.Misbehavior,
		SlashFractionBps:	policy.SlashFractionBps,
		EvidenceHeight:		evidence.EvidenceHeight,
		EvidenceEpoch:		evidence.EvidenceEpoch,
		Finalized:		true,
	}, selfStake, nominations)
}

func (a TaskAssignment) Validate() error {
	if err := validatePosToken("assignment task id", a.TaskID); err != nil {
		return err
	}
	if err := validatePosToken("assignment workload id", a.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(a.WorkloadType); err != nil {
		return err
	}
	if err := validatePosToken("assignment zone id", a.ZoneID); err != nil {
		return err
	}
	if err := validatePosToken("assignment shard id", a.ShardID); err != nil {
		return err
	}
	if err := validatePosToken("assignment workload class", a.WorkloadClass); err != nil {
		return err
	}
	if err := validateValidatorRole(a.Role); err != nil {
		return err
	}
	if len(a.Validators) == 0 {
		return errors.New("assignment validators are required")
	}
	seen := make(map[string]struct{}, len(a.Validators))
	var previous string
	for i, validatorID := range a.Validators {
		if err := validatePosToken("assignment validator id", validatorID); err != nil {
			return err
		}
		if _, found := seen[validatorID]; found {
			return fmt.Errorf("duplicate assignment validator %q", validatorID)
		}
		seen[validatorID] = struct{}{}
		if i > 0 && previous >= validatorID {
			return errors.New("assignment validators must be sorted canonically")
		}
		previous = validatorID
	}
	return validatePosHash("assignment hash", a.AssignmentHash)
}

func (s TaskAssignmentSet) Validate() error {
	if err := validatePosHash("assignment seed", s.Seed); err != nil {
		return err
	}
	if err := validatePosHash("assignment root", s.Root); err != nil {
		return err
	}
	for i, assignment := range s.Assignments {
		if err := assignment.Validate(); err != nil {
			return err
		}
		expectedHash := ComputeTaskAssignmentHash(s.EpochID, s.Seed, assignment)
		if assignment.AssignmentHash != expectedHash {
			return errors.New("assignment hash mismatch")
		}
		if i > 0 && compareTaskAssignments(s.Assignments[i-1], assignment) >= 0 {
			return errors.New("task assignments must be sorted canonically")
		}
	}
	expectedRoot := PosEmptyRootHash
	if len(s.Assignments) > 0 {
		expectedRoot = ComputeTaskAssignmentRoot(s.EpochID, s.Seed, s.Assignments)
	}
	if s.Root != expectedRoot {
		return errors.New("task assignment root mismatch")
	}
	return nil
}

func (g TaskGroup) Validate() error {
	if g.EpochID == 0 {
		return errors.New("task group epoch id is required")
	}
	if err := validatePosToken("task group id", g.TaskGroupID); err != nil {
		return err
	}
	if err := validatePosToken("task group workload id", g.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(g.WorkloadType); err != nil {
		return err
	}
	if len(g.ValidatorMembers) < int(g.MinimumGroupSize) {
		return errors.New("task group members below minimum group size")
	}
	if err := validateCanonicalValidatorIDs("task group member", g.ValidatorMembers); err != nil {
		return err
	}
	if len(g.ProposerOrder) != len(g.ValidatorMembers) {
		return errors.New("task group proposer order must include every member")
	}
	if err := validateValidatorIDSet("task group proposer", g.ProposerOrder, g.ValidatorMembers); err != nil {
		return err
	}
	if len(g.VerifierSet) == 0 {
		return errors.New("task group verifier set is required")
	}
	if err := validateCanonicalValidatorIDs("task group verifier", g.VerifierSet); err != nil {
		return err
	}
	if err := validateValidatorIDSubset("task group verifier", g.VerifierSet, g.ValidatorMembers); err != nil {
		return err
	}
	if err := validatePosHash("task group stake weight root", g.StakeWeightRoot); err != nil {
		return err
	}
	if err := validatePosHash("task group assignment seed", g.AssignmentSeed); err != nil {
		return err
	}
	if g.ActivationHeight == 0 {
		return errors.New("task group activation height is required")
	}
	if g.ExpiryHeight <= g.ActivationHeight {
		return errors.New("task group expiry height must be after activation height")
	}
	if expected := ComputeTaskGroupID(g); g.TaskGroupID != expected {
		return errors.New("task group id mismatch")
	}
	return nil
}

func (s TaskGroupSet) Validate() error {
	if err := validatePosHash("task group set seed", s.Seed); err != nil {
		return err
	}
	if err := validatePosHash("task group set root", s.Root); err != nil {
		return err
	}
	for i, group := range s.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
		if i > 0 && compareTaskGroups(s.Groups[i-1], group) >= 0 {
			return errors.New("task groups must be sorted canonically")
		}
	}
	expectedRoot := PosEmptyRootHash
	if len(s.Groups) > 0 {
		expectedRoot = ComputeTaskGroupRoot(s.EpochID, s.Seed, s.Groups)
	}
	if s.Root != expectedRoot {
		return errors.New("task group root mismatch")
	}
	return nil
}

func ComputeTaskAssignmentHash(epochID uint64, seed string, assignment TaskAssignment) string {
	return posHashRoot("aetheris-pos-task-assignment-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWritePart(w, assignment.TaskID)
		posWritePart(w, assignment.WorkloadID)
		posWritePart(w, string(assignment.WorkloadType))
		posWritePart(w, assignment.ZoneID)
		posWritePart(w, assignment.ShardID)
		posWritePart(w, assignment.WorkloadClass)
		posWritePart(w, string(assignment.Role))
		posWriteUint64(w, uint64(len(assignment.Validators)))
		for _, validatorID := range assignment.Validators {
			posWritePart(w, validatorID)
		}
	})
}

func ComputeTaskAssignmentRoot(epochID uint64, seed string, assignments []TaskAssignment) string {
	return posHashRoot("aetheris-pos-task-assignment-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWriteUint64(w, uint64(len(assignments)))
		for _, assignment := range assignments {
			posWritePart(w, assignment.AssignmentHash)
		}
	})
}

func ComputeTaskGroupID(group TaskGroup) string {
	return posHashRoot("aetheris-pos-task-group-id-v1", func(w posByteWriter) {
		posWriteUint64(w, group.EpochID)
		posWritePart(w, group.WorkloadID)
		posWritePart(w, string(group.WorkloadType))
		posWritePart(w, group.AssignmentSeed)
		posWriteUint64(w, group.ActivationHeight)
		posWriteUint64(w, group.ExpiryHeight)
	})
}

func ComputeTaskGroupStakeWeightRoot(epochID uint64, task WorkloadTask, members []string, validators map[string]ScoredValidator) string {
	return posHashRoot("aetheris-pos-task-group-stake-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, task.TaskID)
		posWritePart(w, task.WorkloadID)
		posWritePart(w, string(task.WorkloadType))
		posWriteUint64(w, uint64(len(members)))
		for _, validatorID := range members {
			validator := validators[validatorID]
			posWritePart(w, validatorID)
			posWritePart(w, validator.ScoreComponents.StakeWeightNaet.String())
			posWritePart(w, validator.VotingPowerNaet.String())
		}
	})
}

func ComputeTaskGroupRoot(epochID uint64, seed string, groups []TaskGroup) string {
	return posHashRoot("aetheris-pos-task-group-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWriteUint64(w, uint64(len(groups)))
		for _, group := range groups {
			posWritePart(w, group.TaskGroupID)
			posWritePart(w, group.StakeWeightRoot)
		}
	})
}

func ActivateDelegationIntents(params Params, electionEpoch uint64, candidates []Candidate, intents []DelegationIntent) ([]DelegationActivation, []RejectedDelegationIntent, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	candidateByID := make(map[string]Candidate, len(candidates))
	for _, candidate := range candidates {
		id := strings.TrimSpace(candidate.ValidatorID)
		if err := validatePosToken("validator id", id); err != nil {
			return nil, nil, err
		}
		if _, found := candidateByID[id]; found {
			return nil, nil, fmt.Errorf("duplicate candidate %q", id)
		}
		candidate.ValidatorID = id
		candidateByID[id] = candidate
	}

	ordered := make([]DelegationIntent, len(intents))
	copy(ordered, intents)
	sort.SliceStable(ordered, func(i, j int) bool {
		return compareDelegationIntents(ordered[i], ordered[j]) < 0
	})

	nominationsByValidator := make(map[string][]Nomination)
	seenNomination := make(map[string]struct{}, len(ordered))
	rejected := make([]RejectedDelegationIntent, 0)
	for _, intent := range ordered {
		if err := intent.Validate(params); err != nil {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: err.Error()})
			continue
		}
		if electionEpoch < intent.RequestedEpoch+params.DelegationActivationEpochs {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "delegation activation delay has not elapsed"})
			continue
		}
		candidate, found := candidateByID[intent.ValidatorID]
		if !found {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator is not in election market"})
			continue
		}
		if candidate.Jailed || candidate.Tombstoned {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator is not eligible for delegation"})
			continue
		}
		if candidate.CommissionBps > intent.MaxCommissionBps {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator commission exceeds delegation risk profile"})
			continue
		}
		if candidate.PerformanceScoreBps < intent.MinPerformanceScoreBps {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator performance below delegation risk profile"})
			continue
		}
		nominationKey := intent.ValidatorID + "\x00" + intent.NominatorID
		if _, found := seenNomination[nominationKey]; found {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "duplicate delegation intent for validator"})
			continue
		}
		seenNomination[nominationKey] = struct{}{}
		nominationsByValidator[intent.ValidatorID] = append(nominationsByValidator[intent.ValidatorID], Nomination{
			NominatorID:	intent.NominatorID,
			StakeNaet:	intent.StakeNaet,
		})
	}

	validatorIDs := make([]string, 0, len(nominationsByValidator))
	for validatorID := range nominationsByValidator {
		validatorIDs = append(validatorIDs, validatorID)
	}
	sort.Strings(validatorIDs)

	activations := make([]DelegationActivation, 0, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		nominations := sortNominations(nominationsByValidator[validatorID])
		totalStake := sumNominations(nominations)
		activation := DelegationActivation{
			ValidatorID:	validatorID,
			Nominations:	nominations,
			ActivatedAt:	electionEpoch,
			IntentCount:	uint32(len(nominations)),
			TotalStake:	totalStake,
			ActivationKey:	computeDelegationActivationKey(electionEpoch, validatorID, nominations),
		}
		activations = append(activations, activation)
	}
	return activations, rejected, nil
}

func (i DelegationIntent) Validate(params Params) error {
	if err := validatePosToken("nominator id", i.NominatorID); err != nil {
		return err
	}
	if err := validatePosToken("validator id", i.ValidatorID); err != nil {
		return err
	}
	if !i.StakeNaet.IsPositive() {
		return errors.New("delegation intent stake must be positive")
	}
	if i.MaxCommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("delegation max commission must be <= %d bps", params.MaxCommissionBps)
	}
	if i.MinPerformanceScoreBps > BasisPoints {
		return fmt.Errorf("delegation minimum performance must be <= %d bps", BasisPoints)
	}
	return nil
}

func SettleEvidenceCase(params Params, currentEpoch uint64, evidence EvidenceCase, selfStake sdkmath.Int, nominations []Nomination) (EvidenceSettlement, error) {
	if err := evidence.Validate(params, currentEpoch); err != nil {
		return EvidenceSettlement{}, err
	}
	slash, err := ComputeSlash(SlashInput{
		ValidatorID:		evidence.ValidatorID,
		Misbehavior:		evidence.Misbehavior,
		SlashFractionBps:	evidence.SlashFractionBps,
		SelfStakeNaet:		selfStake,
		Nominations:		nominations,
		EvidenceHeight:		evidence.EvidenceHeight,
		EvidenceFinalized:	true,
	})
	if err != nil {
		return EvidenceSettlement{}, err
	}
	reporterReward := mulIntBps(slash.TotalSlashedNaet, params.ReporterRewardBps)
	settlement := EvidenceSettlement{
		EvidenceID:		evidence.EvidenceID,
		ReporterID:		evidence.ReporterID,
		Slash:			slash,
		ReporterRewardNaet:	reporterReward,
		BurnNaet:		slash.TotalSlashedNaet.Sub(reporterReward),
	}
	settlement.SettlementHash = computeEvidenceSettlementHash(settlement)
	return settlement, nil
}

func (e EvidenceCase) Validate(params Params, currentEpoch uint64) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validatePosToken("evidence id", e.EvidenceID); err != nil {
		return err
	}
	if err := validatePosToken("evidence reporter id", e.ReporterID); err != nil {
		return err
	}
	if err := validatePosToken("evidence validator id", e.ValidatorID); err != nil {
		return err
	}
	if !IsSlashableMisbehavior(e.Misbehavior) {
		return fmt.Errorf("unsupported misbehavior %q", e.Misbehavior)
	}
	if e.SlashFractionBps == 0 || e.SlashFractionBps > BasisPoints {
		return fmt.Errorf("slash fraction must be within 1..%d bps", BasisPoints)
	}
	if e.EvidenceHeight < 0 {
		return errors.New("evidence height cannot be negative")
	}
	if !e.Finalized {
		return errors.New("evidence must be finalized before settlement")
	}
	withinWindow, err := EvidenceWithinSlashableWindow(params, e.EvidenceEpoch, currentEpoch)
	if err != nil {
		return err
	}
	if !withinWindow {
		return errors.New("evidence is outside slashable window")
	}
	return nil
}

func DefaultRoleRewardWeights() []RoleRewardWeight {
	return []RoleRewardWeight{
		{Role: ValidatorRoleBlockProducer, WeightBps: 3_500},
		{Role: ValidatorRoleVerifier, WeightBps: 3_500},
		{Role: ValidatorRoleCollator, WeightBps: 1_500},
		{Role: ValidatorRoleEvidenceReviewer, WeightBps: 1_500},
	}
}

func SettleWorkloadRewards(input WorkloadRewardInput) (WorkloadRewardSettlement, error) {
	if input.TotalRewardsNaet.IsNegative() {
		return WorkloadRewardSettlement{}, errors.New("workload rewards cannot be negative")
	}
	weights := input.RoleWeights
	if len(weights) == 0 {
		weights = DefaultRoleRewardWeights()
	}
	if err := validateRoleRewardWeights(weights); err != nil {
		return WorkloadRewardSettlement{}, err
	}

	outcomesByRole := make(map[ValidatorRole][]AssignmentOutcome)
	for _, outcome := range input.Outcomes {
		if err := outcome.Validate(); err != nil {
			return WorkloadRewardSettlement{}, err
		}
		outcomesByRole[outcome.Role] = append(outcomesByRole[outcome.Role], outcome)
	}

	rewardByValidator := make(map[string]sdkmath.Int)
	workUnitsByValidator := make(map[string]uint64)
	remainder := sdkmath.ZeroInt()
	completedUnits := uint64(0)

	for _, weight := range weights {
		roleBudget := mulIntBps(input.TotalRewardsNaet, weight.WeightBps)
		roleUnitsByValidator := make(map[string]uint64)
		totalRoleUnits := uint64(0)
		for _, outcome := range outcomesByRole[weight.Role] {
			if !outcome.Completed || outcome.Faulted || outcome.WorkUnits == 0 {
				continue
			}
			roleUnitsByValidator[outcome.ValidatorID] += outcome.WorkUnits
			workUnitsByValidator[outcome.ValidatorID] += outcome.WorkUnits
			totalRoleUnits += outcome.WorkUnits
			completedUnits += outcome.WorkUnits
		}
		if totalRoleUnits == 0 {
			remainder = remainder.Add(roleBudget)
			continue
		}
		validatorIDs := sortedStringKeys(roleUnitsByValidator)
		distributed := sdkmath.ZeroInt()
		for _, validatorID := range validatorIDs {
			reward := roleBudget.MulRaw(int64(roleUnitsByValidator[validatorID])).QuoRaw(int64(totalRoleUnits))
			currentReward, found := rewardByValidator[validatorID]
			if !found {
				currentReward = sdkmath.ZeroInt()
			}
			rewardByValidator[validatorID] = currentReward.Add(reward)
			distributed = distributed.Add(reward)
		}
		remainder = remainder.Add(roleBudget.Sub(distributed))
	}

	validatorIDs := sortedStringKeys(rewardByValidator)
	rewards := make([]ValidatorWorkloadReward, 0, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		rewards = append(rewards, ValidatorWorkloadReward{
			ValidatorID:	validatorID,
			RewardNaet:	rewardByValidator[validatorID],
			WorkUnits:	workUnitsByValidator[validatorID],
		})
	}
	settlement := WorkloadRewardSettlement{
		EpochID:	input.EpochID,
		Rewards:	rewards,
		RemainderNaet:	remainder,
		CompletedUnits:	completedUnits,
	}
	settlement.RewardRoot = computeWorkloadRewardRoot(settlement)
	return settlement, nil
}

func (o AssignmentOutcome) Validate() error {
	if err := validatePosToken("assignment outcome task id", o.TaskID); err != nil {
		return err
	}
	if err := validateValidatorRole(o.Role); err != nil {
		return err
	}
	if err := validatePosToken("assignment outcome validator id", o.ValidatorID); err != nil {
		return err
	}
	if o.Completed && o.Faulted {
		return errors.New("assignment outcome cannot be both completed and faulted")
	}
	return nil
}

func ValidatorSupportsRole(candidate Candidate, role ValidatorRole) bool {
	if err := validateValidatorRole(role); err != nil {
		return false
	}
	if len(candidate.Roles) == 0 {
		return true
	}
	for _, candidateRole := range candidate.Roles {
		if candidateRole == role {
			return true
		}
	}
	return false
}

func ValidatorRoleValues() []ValidatorRole {
	return []ValidatorRole{
		ValidatorRoleValidator,
		ValidatorRoleProposer,
		ValidatorRoleVerifier,
		ValidatorRoleEvidenceReporter,
		ValidatorRoleDelegationOperator,
		ValidatorRoleCollator,
		ValidatorRoleFisherman,
	}
}

func RoleRecordFieldNames() []string {
	return []string{
		"validator_address",
		"role",
		"epoch_id",
		"status",
		"eligibility_score",
		"capacity",
		"assigned_task_count",
		"performance_score",
	}
}

func RoleStatusValues() []string {
	return []string{
		RoleStatusEligible,
		RoleStatusAssigned,
		RoleStatusSuspended,
		RoleStatusInactive,
	}
}

func CollatorRecordFieldNames() []string {
	return []string{
		"collator_id",
		"operator_address",
		"supported_workloads",
		"bond_optional",
		"reputation",
		"status",
		"registered_epoch",
	}
}

func CollatorStatusValues() []string {
	return []string{
		CollatorStatusRegistered,
		CollatorStatusActive,
		CollatorStatusSuspended,
		CollatorStatusRetired,
	}
}

func NewRoleRecord(record RoleRecord) (RoleRecord, error) {
	record.ValidatorAddress = strings.TrimSpace(record.ValidatorAddress)
	record.Status = strings.TrimSpace(record.Status)
	if record.Status == "" {
		record.Status = RoleStatusEligible
	}
	return record, record.Validate()
}

func (r RoleRecord) Validate() error {
	if err := validatePosToken("role record validator address", r.ValidatorAddress); err != nil {
		return err
	}
	if err := validateValidatorRole(r.Role); err != nil {
		return err
	}
	if r.EpochID == 0 {
		return errors.New("role record epoch id is required")
	}
	if err := validateRoleStatus(r.Status); err != nil {
		return err
	}
	if r.EligibilityScore > BasisPoints {
		return fmt.Errorf("role record eligibility score must be <= %d bps", BasisPoints)
	}
	if r.PerformanceScore > BasisPoints {
		return fmt.Errorf("role record performance score must be <= %d bps", BasisPoints)
	}
	if err := r.Capacity.Validate(); err != nil {
		return err
	}
	if r.Capacity.MaxTaskGroups > 0 && r.AssignedTaskCount > r.Capacity.MaxTaskGroups {
		return errors.New("role record assigned task count exceeds capacity")
	}
	if r.Status == RoleStatusAssigned && r.AssignedTaskCount == 0 {
		return errors.New("assigned role record requires assigned task count")
	}
	return nil
}

func ValidateRoleRecords(records []RoleRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s|%s|%d", record.ValidatorAddress, record.Role, record.EpochID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate role record %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func NewCollatorRecord(record CollatorRecord) (CollatorRecord, error) {
	record.CollatorID = strings.TrimSpace(record.CollatorID)
	record.OperatorAddress = strings.TrimSpace(record.OperatorAddress)
	record.Status = strings.TrimSpace(record.Status)
	if record.Status == "" {
		record.Status = CollatorStatusRegistered
	}
	if record.BondOptional.IsNil() {
		record.BondOptional = sdkmath.ZeroInt()
	}
	record.SupportedWorkloads = normalizedWorkloadTypes(record.SupportedWorkloads)
	return record, record.Validate()
}

func (c CollatorRecord) Validate() error {
	if err := validatePosToken("collator id", c.CollatorID); err != nil {
		return err
	}
	if err := validatePosToken("collator operator address", c.OperatorAddress); err != nil {
		return err
	}
	if len(c.SupportedWorkloads) == 0 {
		return errors.New("collator supported workloads are required")
	}
	if err := validateWorkloadTypes(c.SupportedWorkloads); err != nil {
		return err
	}
	if c.BondOptional.IsNil() {
		return errors.New("collator bond optional must be set")
	}
	if c.BondOptional.IsNegative() {
		return errors.New("collator bond optional cannot be negative")
	}
	if c.Reputation > BasisPoints {
		return fmt.Errorf("collator reputation must be <= %d bps", BasisPoints)
	}
	if err := validateCollatorStatus(c.Status); err != nil {
		return err
	}
	if c.RegisteredEpoch == 0 {
		return errors.New("collator registered epoch is required")
	}
	return nil
}

func (c CollatorRecord) SupportsWorkload(workloadType WorkloadType) bool {
	if err := validateWorkloadType(workloadType); err != nil {
		return false
	}
	for _, supported := range c.SupportedWorkloads {
		if supported == workloadType {
			return true
		}
	}
	return false
}

func BuildCollatorCandidateOutput(params Params, input CollatorCandidateOutputInput) (CollatorCandidateOutput, error) {
	if err := params.Validate(); err != nil {
		return CollatorCandidateOutput{}, err
	}
	collator, err := NewCollatorRecord(input.Collator)
	if err != nil {
		return CollatorCandidateOutput{}, err
	}
	if collator.Status == CollatorStatusSuspended || collator.Status == CollatorStatusRetired {
		return CollatorCandidateOutput{}, errors.New("collator is not eligible to build candidate outputs")
	}
	task := normalizeWorkloadTask(params, input.Task)
	if err := task.Validate(params); err != nil {
		return CollatorCandidateOutput{}, err
	}
	if !collator.SupportsWorkload(task.WorkloadType) {
		return CollatorCandidateOutput{}, fmt.Errorf("collator does not support workload %q", task.WorkloadType)
	}
	if input.EpochID == 0 {
		return CollatorCandidateOutput{}, errors.New("collator candidate output epoch id is required")
	}
	output := CollatorCandidateOutput{
		EpochID:			input.EpochID,
		CollatorID:			collator.CollatorID,
		OperatorAddress:		collator.OperatorAddress,
		TaskID:				task.TaskID,
		TaskGroupIDOptional:		strings.TrimSpace(input.TaskGroupIDOptional),
		WorkloadID:			task.WorkloadID,
		WorkloadType:			task.WorkloadType,
		TransactionRoot:		input.TransactionRoot,
		StateTransitionRoot:		input.StateTransitionRoot,
		ProofBundleRoot:		input.ProofBundleRoot,
		RequiresValidatorVerification:	true,
		ValidatorSignatures:		nil,
		Finalized:			false,
	}
	output.CandidateOutputHash = ComputeCollatorCandidateOutputHash(output)
	return output, output.Validate()
}

func (o CollatorCandidateOutput) Validate() error {
	if o.EpochID == 0 {
		return errors.New("collator candidate output epoch id is required")
	}
	if err := validatePosToken("collator output collator id", o.CollatorID); err != nil {
		return err
	}
	if err := validatePosToken("collator output operator address", o.OperatorAddress); err != nil {
		return err
	}
	if err := validatePosToken("collator output task id", o.TaskID); err != nil {
		return err
	}
	if o.TaskGroupIDOptional != "" {
		if err := validatePosToken("collator output task group id", o.TaskGroupIDOptional); err != nil {
			return err
		}
	}
	if err := validatePosToken("collator output workload id", o.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(o.WorkloadType); err != nil {
		return err
	}
	if err := validatePosHash("collator output transaction root", o.TransactionRoot); err != nil {
		return err
	}
	if err := validatePosHash("collator output state transition root", o.StateTransitionRoot); err != nil {
		return err
	}
	if err := validatePosHash("collator output proof bundle root", o.ProofBundleRoot); err != nil {
		return err
	}
	if !o.RequiresValidatorVerification {
		return errors.New("collator output requires validator verification")
	}
	for _, signature := range o.ValidatorSignatures {
		if err := validatePosHash("collator output validator signature", signature); err != nil {
			return err
		}
	}
	if o.Finalized && len(o.ValidatorSignatures) == 0 {
		return errors.New("finalized collator output requires validator signatures")
	}
	if err := validatePosHash("collator output hash", o.CandidateOutputHash); err != nil {
		return err
	}
	if expected := ComputeCollatorCandidateOutputHash(o); expected != o.CandidateOutputHash {
		return errors.New("collator candidate output hash mismatch")
	}
	return nil
}

func ComputeCollatorCandidateOutputHash(output CollatorCandidateOutput) string {
	return posHashRoot("aetheris-pos-collator-output-v1", func(w posByteWriter) {
		posWriteUint64(w, output.EpochID)
		posWritePart(w, output.CollatorID)
		posWritePart(w, output.OperatorAddress)
		posWritePart(w, output.TaskID)
		posWritePart(w, output.TaskGroupIDOptional)
		posWritePart(w, output.WorkloadID)
		posWritePart(w, string(output.WorkloadType))
		posWritePart(w, output.TransactionRoot)
		posWritePart(w, output.StateTransitionRoot)
		posWritePart(w, output.ProofBundleRoot)
		posWriteUint64(w, boolAsUint64(output.RequiresValidatorVerification))
	})
}

func NewCollatorRegistry(epochID uint64, collators []CollatorRecord) (CollatorRegistry, error) {
	records := make([]CollatorRecord, len(collators))
	for i, collator := range collators {
		normalized, err := NewCollatorRecord(collator)
		if err != nil {
			return CollatorRegistry{}, err
		}
		records[i] = normalized
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CollatorID < records[j].CollatorID
	})
	registry := CollatorRegistry{EpochID: epochID, Collators: records}
	if len(records) == 0 {
		registry.RegistryRoot = PosEmptyRootHash
	} else {
		registry.RegistryRoot = ComputeCollatorRegistryRoot(registry)
	}
	return registry, registry.Validate()
}

func (r CollatorRegistry) Validate() error {
	if r.EpochID == 0 {
		return errors.New("collator registry epoch id is required")
	}
	seen := make(map[string]struct{}, len(r.Collators))
	for _, collator := range r.Collators {
		if err := collator.Validate(); err != nil {
			return err
		}
		if collator.RegisteredEpoch > r.EpochID {
			return errors.New("collator registered epoch cannot exceed registry epoch")
		}
		if _, found := seen[collator.CollatorID]; found {
			return fmt.Errorf("duplicate collator id %q", collator.CollatorID)
		}
		seen[collator.CollatorID] = struct{}{}
	}
	expectedRoot := PosEmptyRootHash
	if len(r.Collators) > 0 {
		expectedRoot = ComputeCollatorRegistryRoot(r)
	}
	if r.RegistryRoot != expectedRoot {
		return errors.New("collator registry root mismatch")
	}
	return nil
}

func (r CollatorRegistry) CollatorByID(collatorID string) (CollatorRecord, bool, error) {
	if err := validatePosToken("collator id", collatorID); err != nil {
		return CollatorRecord{}, false, err
	}
	if err := r.Validate(); err != nil {
		return CollatorRecord{}, false, err
	}
	for _, collator := range r.Collators {
		if collator.CollatorID == collatorID {
			return collator, true, nil
		}
	}
	return CollatorRecord{}, false, nil
}

func (r CollatorRegistry) ActiveCollatorsForWorkload(workloadType WorkloadType) ([]CollatorRecord, error) {
	if err := validateWorkloadType(workloadType); err != nil {
		return nil, err
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	out := make([]CollatorRecord, 0, len(r.Collators))
	for _, collator := range r.Collators {
		if collator.Status == CollatorStatusActive && collator.SupportsWorkload(workloadType) {
			out = append(out, collator)
		}
	}
	return out, nil
}

func ComputeCollatorRegistryRoot(registry CollatorRegistry) string {
	return posHashRoot("aetheris-pos-collator-registry-v1", func(w posByteWriter) {
		posWriteUint64(w, registry.EpochID)
		posWriteUint64(w, uint64(len(registry.Collators)))
		for _, collator := range registry.Collators {
			posWritePart(w, collator.CollatorID)
			posWritePart(w, collator.OperatorAddress)
			posWriteUint64(w, uint64(len(collator.SupportedWorkloads)))
			for _, workloadType := range collator.SupportedWorkloads {
				posWritePart(w, string(workloadType))
			}
			posWritePart(w, collator.BondOptional.String())
			posWriteUint64(w, uint64(collator.Reputation))
			posWritePart(w, collator.Status)
			posWriteUint64(w, collator.RegisteredEpoch)
		}
	})
}

func NewCollatorOutputVerification(output CollatorCandidateOutput, validatorAddress string, result string, signatureHash string, verifiedHeight int64) (CollatorOutputVerification, error) {
	if err := output.Validate(); err != nil {
		return CollatorOutputVerification{}, err
	}
	verification := CollatorOutputVerification{
		OutputHash:		output.CandidateOutputHash,
		ValidatorAddress:	strings.TrimSpace(validatorAddress),
		Result:			strings.TrimSpace(result),
		SignatureHash:		strings.TrimSpace(signatureHash),
		VerifiedHeight:		verifiedHeight,
	}
	return verification, verification.Validate()
}

func (v CollatorOutputVerification) Validate() error {
	if err := validatePosHash("collator output verification hash", v.OutputHash); err != nil {
		return err
	}
	if err := validatePosToken("collator output verifier", v.ValidatorAddress); err != nil {
		return err
	}
	if err := validateCollatorVerificationResult(v.Result); err != nil {
		return err
	}
	if err := validatePosHash("collator output verification signature", v.SignatureHash); err != nil {
		return err
	}
	if v.VerifiedHeight < 0 {
		return errors.New("collator output verification height cannot be negative")
	}
	return nil
}

func VerifyCollatorOutputByValidators(output CollatorCandidateOutput, validatorSet []string, votes []CollatorOutputVerification, decisionThresholdBps uint32) (CollatorOutputVerificationResult, error) {
	if err := output.Validate(); err != nil {
		return CollatorOutputVerificationResult{}, err
	}
	if decisionThresholdBps == 0 {
		decisionThresholdBps = DefaultEvidenceVerificationQuorumBps
	}
	if decisionThresholdBps > BasisPoints {
		return CollatorOutputVerificationResult{}, fmt.Errorf("collator verification threshold must be <= %d bps", BasisPoints)
	}
	allowed, err := validatorSetMap("collator verification validator", validatorSet)
	if err != nil {
		return CollatorOutputVerificationResult{}, err
	}
	if len(validatorSet) == 0 {
		return CollatorOutputVerificationResult{}, errors.New("collator verification validator set is required")
	}
	seen := make(map[string]struct{}, len(votes))
	result := CollatorOutputVerificationResult{
		OutputHash:		output.CandidateOutputHash,
		TotalValidators:	uint32(len(validatorSet)),
		DecisionThresholdBps:	decisionThresholdBps,
	}
	for _, vote := range votes {
		if err := vote.Validate(); err != nil {
			return CollatorOutputVerificationResult{}, err
		}
		if vote.OutputHash != output.CandidateOutputHash {
			return CollatorOutputVerificationResult{}, errors.New("collator output verification hash mismatch")
		}
		if _, ok := allowed[vote.ValidatorAddress]; !ok {
			return CollatorOutputVerificationResult{}, errors.New("collator output verifier is not in validator set")
		}
		if _, found := seen[vote.ValidatorAddress]; found {
			return CollatorOutputVerificationResult{}, fmt.Errorf("duplicate collator output verification by %q", vote.ValidatorAddress)
		}
		seen[vote.ValidatorAddress] = struct{}{}
		switch vote.Result {
		case CollatorVerificationResultValid:
			result.ValidVotes++
			result.ValidSignatureHashes = append(result.ValidSignatureHashes, vote.SignatureHash)
		case CollatorVerificationResultInvalid:
			result.InvalidVotes++
		case CollatorVerificationResultAbstain:
			result.AbstainVotes++
		}
	}
	result.ParticipationBps = ratioBps(uint64(len(seen)), uint64(len(validatorSet)))
	validBps := ratioBps(uint64(result.ValidVotes), uint64(len(validatorSet)))
	invalidBps := ratioBps(uint64(result.InvalidVotes), uint64(len(validatorSet)))
	result.Accepted = validBps >= decisionThresholdBps
	result.Rejected = invalidBps >= decisionThresholdBps
	if result.Accepted && result.Rejected {
		return CollatorOutputVerificationResult{}, errors.New("collator output verification cannot be both accepted and rejected")
	}
	result.VerificationRoot = ComputeCollatorOutputVerificationRoot(result)
	return result, nil
}

func FinalizeCollatorOutputAfterVerification(output CollatorCandidateOutput, verification CollatorOutputVerificationResult) (CollatorCandidateOutput, error) {
	if err := output.Validate(); err != nil {
		return CollatorCandidateOutput{}, err
	}
	if err := verification.Validate(); err != nil {
		return CollatorCandidateOutput{}, err
	}
	if verification.OutputHash != output.CandidateOutputHash {
		return CollatorCandidateOutput{}, errors.New("collator verification output hash mismatch")
	}
	if !verification.Accepted {
		return CollatorCandidateOutput{}, errors.New("collator output is not accepted by validators")
	}
	out := output
	out.ValidatorSignatures = append([]string(nil), verification.ValidSignatureHashes...)
	out.Finalized = true
	return out, out.Validate()
}

func (r CollatorOutputVerificationResult) Validate() error {
	if err := validatePosHash("collator output verification result hash", r.OutputHash); err != nil {
		return err
	}
	if r.TotalValidators == 0 {
		return errors.New("collator output verification result requires validators")
	}
	if r.ValidVotes+r.InvalidVotes+r.AbstainVotes > r.TotalValidators {
		return errors.New("collator output verification votes exceed validator set")
	}
	if r.ParticipationBps > BasisPoints || r.DecisionThresholdBps > BasisPoints {
		return fmt.Errorf("collator output verification bps must be <= %d", BasisPoints)
	}
	if r.Accepted && r.Rejected {
		return errors.New("collator output verification cannot be both accepted and rejected")
	}
	for _, signature := range r.ValidSignatureHashes {
		if err := validatePosHash("collator output valid signature", signature); err != nil {
			return err
		}
	}
	if err := validatePosHash("collator output verification root", r.VerificationRoot); err != nil {
		return err
	}
	if expected := ComputeCollatorOutputVerificationRoot(r); expected != r.VerificationRoot {
		return errors.New("collator output verification root mismatch")
	}
	return nil
}

func ComputeCollatorOutputVerificationRoot(result CollatorOutputVerificationResult) string {
	return posHashRoot("aetheris-pos-collator-verification-v1", func(w posByteWriter) {
		posWritePart(w, result.OutputHash)
		posWriteUint64(w, uint64(result.ValidVotes))
		posWriteUint64(w, uint64(result.InvalidVotes))
		posWriteUint64(w, uint64(result.AbstainVotes))
		posWriteUint64(w, uint64(result.TotalValidators))
		posWriteUint64(w, uint64(result.ParticipationBps))
		posWriteUint64(w, uint64(result.DecisionThresholdBps))
		posWriteUint64(w, boolAsUint64(result.Accepted))
		posWriteUint64(w, boolAsUint64(result.Rejected))
		posWriteUint64(w, uint64(len(result.ValidSignatureHashes)))
		for _, signature := range result.ValidSignatureHashes {
			posWritePart(w, signature)
		}
	})
}

func BuildInvalidCollatorOutputEvidence(evidenceID string, reporterID string, collator CollatorRecord, output CollatorCandidateOutput, verification CollatorOutputVerificationResult, submittedHeight int64) (StructuredEvidenceRecord, error) {
	if err := collator.Validate(); err != nil {
		return StructuredEvidenceRecord{}, err
	}
	if err := output.Validate(); err != nil {
		return StructuredEvidenceRecord{}, err
	}
	if output.CollatorID != collator.CollatorID {
		return StructuredEvidenceRecord{}, errors.New("invalid collator evidence collator id mismatch")
	}
	if err := verification.Validate(); err != nil {
		return StructuredEvidenceRecord{}, err
	}
	if verification.OutputHash != output.CandidateOutputHash {
		return StructuredEvidenceRecord{}, errors.New("invalid collator evidence output hash mismatch")
	}
	if !verification.Rejected {
		return StructuredEvidenceRecord{}, errors.New("invalid collator evidence requires validator rejection")
	}
	return SubmitStructuredEvidence(StructuredEvidenceRecord{
		EvidenceID:		evidenceID,
		EvidenceType:		EvidenceTypeInvalidCollatorOutputProof,
		ReporterID:		reporterID,
		AccusedValidatorID:	collator.CollatorID,
		SubjectID:		output.CandidateOutputHash,
		EvidenceHash:		ComputeInvalidCollatorOutputEvidenceHash(collator, output, verification),
		EvidenceHeight:		submittedHeight,
		EvidenceEpoch:		output.EpochID,
		SubmittedHeight:	submittedHeight,
		VerificationGroupID:	fmt.Sprintf("collator-output/%s", collator.CollatorID),
	})
}

func ComputeInvalidCollatorOutputEvidenceHash(collator CollatorRecord, output CollatorCandidateOutput, verification CollatorOutputVerificationResult) string {
	return posHashRoot("aetheris-pos-invalid-collator-output-evidence-v1", func(w posByteWriter) {
		posWritePart(w, collator.CollatorID)
		posWritePart(w, output.CandidateOutputHash)
		posWritePart(w, verification.VerificationRoot)
		posWritePart(w, collator.BondOptional.String())
	})
}

func ComputeInvalidCollatorOutputPenalty(collator CollatorRecord, evidence StructuredEvidenceRecord) (sdkmath.Int, error) {
	if err := collator.Validate(); err != nil {
		return sdkmath.Int{}, err
	}
	if err := evidence.Validate(); err != nil {
		return sdkmath.Int{}, err
	}
	if evidence.EvidenceType != EvidenceTypeInvalidCollatorOutputProof {
		return sdkmath.Int{}, errors.New("invalid collator output penalty requires invalid collator evidence")
	}
	if evidence.AccusedValidatorID != collator.CollatorID {
		return sdkmath.Int{}, errors.New("invalid collator output evidence accused collator mismatch")
	}
	if evidence.Status != EvidenceStatusAccepted && evidence.Status != EvidenceStatusFinalized && evidence.Status != EvidenceStatusSlashed {
		return sdkmath.Int{}, errors.New("invalid collator output evidence must be accepted before penalty")
	}
	if !collator.BondOptional.IsPositive() {
		return sdkmath.ZeroInt(), nil
	}
	penalty := collator.BondOptional.MulRaw(int64(DefaultInvalidCollatorOutputSlashBps)).QuoRaw(int64(BasisPoints))
	if penalty.GT(collator.BondOptional) {
		return collator.BondOptional, nil
	}
	return penalty, nil
}

func DefaultRoleRegistry() RoleRegistry {
	return RoleRegistry{Rules: []RoleRule{
		{Role: ValidatorRoleValidator, Description: "participates in consensus security", RequiresValidator: true, RequiresMinimumStake: true, RewardWeightBps: 2_000, MinimumPerformanceBps: 8_000, MinimumEligibilityBps: 8_000, CanFinalize: true},
		{Role: ValidatorRoleProposer, Description: "produces canonical block or task output for slot", RequiresValidator: true, RequiresMinimumStake: true, RewardWeightBps: 1_500, MinimumPerformanceBps: 8_500, MinimumEligibilityBps: 8_500},
		{Role: ValidatorRoleVerifier, Description: "re-executes and signs verification receipts", RequiresValidator: true, RequiresMinimumStake: true, RewardWeightBps: 2_000, MinimumPerformanceBps: 8_500, MinimumEligibilityBps: 8_500},
		{Role: ValidatorRoleEvidenceReporter, Description: "detects and submits faults", RequiresDeposit: true, RewardWeightBps: 1_000, MinimumPerformanceBps: 7_000, MinimumEligibilityBps: 7_000},
		{Role: ValidatorRoleDelegationOperator, Description: "manages delegated capital strategy where authorized", RequiresAuthorization: true, RequiresFeeDisclosure: true, RequiresRiskPolicy: true, RewardWeightBps: 1_000, MinimumPerformanceBps: 8_000, MinimumEligibilityBps: 8_000},
		{Role: ValidatorRoleCollator, Description: "assembles transactions, state transitions, and proof bundles", RequiresValidator: true, RewardWeightBps: 1_000, MinimumPerformanceBps: 8_000, MinimumEligibilityBps: 8_000},
		{Role: ValidatorRoleFisherman, Description: "external fault detector submitting fraud proofs with deposit", RequiresDeposit: true, RewardWeightBps: 500, MinimumPerformanceBps: 6_000, MinimumEligibilityBps: 6_000},
	}}
}

func (r RoleRegistry) Validate() error {
	seen := make(map[ValidatorRole]struct{}, len(r.Rules))
	for _, rule := range r.Rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.Role]; found {
			return fmt.Errorf("duplicate role registry rule %q", rule.Role)
		}
		seen[rule.Role] = struct{}{}
	}
	return nil
}

func (r RoleRegistry) Rule(role ValidatorRole) (RoleRule, bool, error) {
	if err := validateValidatorRole(role); err != nil {
		return RoleRule{}, false, err
	}
	if err := r.Validate(); err != nil {
		return RoleRule{}, false, err
	}
	for _, rule := range r.Rules {
		if rule.Role == role {
			return rule, true, nil
		}
	}
	return RoleRule{}, false, nil
}

func (r RoleRule) Validate() error {
	if err := validateValidatorRole(r.Role); err != nil {
		return err
	}
	if strings.TrimSpace(r.Description) == "" {
		return errors.New("role registry description is required")
	}
	if r.RewardWeightBps > BasisPoints {
		return fmt.Errorf("role reward weight must be <= %d bps", BasisPoints)
	}
	if r.MinimumPerformanceBps > BasisPoints || r.MinimumEligibilityBps > BasisPoints {
		return fmt.Errorf("role minimum scores must be <= %d bps", BasisPoints)
	}
	if r.Role == ValidatorRoleCollator && r.CanFinalize {
		return errors.New("collator role cannot finalize without validator verification")
	}
	return nil
}

func CheckRoleEligibility(registry RoleRegistry, input RoleEligibilityInput) (RoleRecord, error) {
	input.ActorAddress = strings.TrimSpace(input.ActorAddress)
	if err := input.Params.Validate(); err != nil {
		return RoleRecord{}, err
	}
	rule, found, err := registry.Rule(input.Role)
	if err != nil {
		return RoleRecord{}, err
	}
	if !found {
		return RoleRecord{}, fmt.Errorf("missing role registry rule %q", input.Role)
	}
	if rule.RequiresValidator {
		if err := input.Candidate.Validate(input.Params); err != nil {
			return RoleRecord{}, err
		}
		if !ValidatorSupportsRole(input.Candidate, input.Role) && !legacyRoleAliasSupports(input.Candidate, input.Role) {
			return RoleRecord{}, fmt.Errorf("candidate does not support role %q", input.Role)
		}
		if rule.RequiresMinimumStake && input.Candidate.SelfStakeNaet.Add(input.Candidate.DelegatedStakeNaet).LT(input.Params.MinStakeNaet) {
			return RoleRecord{}, errors.New("role requires minimum validator stake")
		}
		input.ActorAddress = input.Candidate.ValidatorID
	}
	if input.ActorAddress == "" {
		input.ActorAddress = strings.TrimSpace(input.Candidate.ValidatorID)
	}
	if input.ActorAddress == "" {
		return RoleRecord{}, errors.New("role eligibility actor address is required")
	}
	if rule.RequiresDeposit && (input.DepositNaet.IsNil() || !input.DepositNaet.IsPositive()) {
		return RoleRecord{}, errors.New("role requires evidence or fraud-proof deposit")
	}
	if rule.RequiresAuthorization && !input.DelegationOperatorAuthorized {
		return RoleRecord{}, errors.New("delegation operator role requires authorization")
	}
	if rule.RequiresFeeDisclosure && !input.FeesDisclosed {
		return RoleRecord{}, errors.New("delegation operator role requires fee disclosure")
	}
	if rule.RequiresRiskPolicy && !input.RiskPolicyDisclosed {
		return RoleRecord{}, errors.New("delegation operator role requires risk policy disclosure")
	}
	performance := normalizeOptionalFactorBps(input.Candidate.PerformanceScoreBps)
	if !rule.RequiresValidator && performance == BasisPoints && input.Candidate.ValidatorID == "" {
		performance = rule.MinimumPerformanceBps
	}
	if performance < rule.MinimumPerformanceBps {
		return RoleRecord{}, errors.New("role performance below requirement")
	}
	uptime := normalizeOptionalFactorBps(input.Candidate.UptimeFactorBps)
	if !rule.RequiresValidator && input.Candidate.ValidatorID == "" {
		uptime = BasisPoints
	}
	eligibility := uint32((uint64(performance) + uint64(uptime)) / 2)
	record, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	input.ActorAddress,
		Role:			input.Role,
		EpochID:		1,
		Status:			RoleStatusEligible,
		EligibilityScore:	eligibility,
		Capacity:		input.Candidate.Capacity,
		PerformanceScore:	performance,
	})
	if err != nil {
		return RoleRecord{}, err
	}
	if record.EligibilityScore < rule.MinimumEligibilityBps {
		return RoleRecord{}, errors.New("role eligibility score below requirement")
	}
	return record, nil
}

func ComputeRolePerformanceMetrics(record RoleRecord, completedTasks uint32, faultedTasks uint32, missedTasks uint32) (RolePerformanceMetrics, error) {
	if err := record.Validate(); err != nil {
		return RolePerformanceMetrics{}, err
	}
	if completedTasks+faultedTasks+missedTasks > record.AssignedTaskCount {
		return RolePerformanceMetrics{}, errors.New("role performance task counts exceed assigned tasks")
	}
	score := uint32(BasisPoints)
	if record.AssignedTaskCount > 0 {
		completedBps := ratioBps(uint64(completedTasks), uint64(record.AssignedTaskCount))
		faultPenalty := roleMinBps(uint64(faultedTasks) * 2_500)
		missedPenalty := roleMinBps(uint64(missedTasks) * 1_000)
		if uint64(faultPenalty)+uint64(missedPenalty) >= uint64(completedBps) {
			score = 0
		} else {
			score = completedBps - faultPenalty - missedPenalty
		}
	}
	return RolePerformanceMetrics{
		ValidatorAddress:	record.ValidatorAddress,
		Role:			record.Role,
		EpochID:		record.EpochID,
		AssignedTasks:		record.AssignedTaskCount,
		CompletedTasks:		completedTasks,
		FaultedTasks:		faultedTasks,
		MissedTasks:		missedTasks,
		PerformanceScore:	score,
	}, nil
}

func SettleRoleRewards(input RoleRewardInput) (WorkloadRewardSettlement, error) {
	outcomes := make([]AssignmentOutcome, 0, len(input.Records))
	for _, record := range input.Records {
		if err := record.Validate(); err != nil {
			return WorkloadRewardSettlement{}, err
		}
		outcomes = append(outcomes, AssignmentOutcome{
			TaskID:		fmt.Sprintf("role/%s/%d", record.Role, record.EpochID),
			Role:		record.Role,
			ValidatorID:	record.ValidatorAddress,
			Completed:	record.Status == RoleStatusAssigned || record.Status == RoleStatusEligible,
			Faulted:	record.Status == RoleStatusSuspended,
			WorkUnits:	uint64(record.PerformanceScore) * uint64(roleMaxUint32(record.AssignedTaskCount, 1)),
		})
	}
	return SettleWorkloadRewards(WorkloadRewardInput{
		EpochID:		input.EpochID,
		TotalRewardsNaet:	input.TotalRewardsNaet,
		RoleWeights:		input.Weights,
		Outcomes:		outcomes,
	})
}

func SuspendRoleOnFault(records []RoleRecord, validatorAddress string, role ValidatorRole, epochID uint64) ([]RoleRecord, error) {
	if err := validatePosToken("role suspension validator address", validatorAddress); err != nil {
		return nil, err
	}
	if err := validateValidatorRole(role); err != nil {
		return nil, err
	}
	out := make([]RoleRecord, len(records))
	copy(out, records)
	found := false
	for i, record := range out {
		if record.ValidatorAddress == validatorAddress && record.Role == role && record.EpochID == epochID {
			record.Status = RoleStatusSuspended
			record.AssignedTaskCount = 0
			out[i] = record
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("role record not found for suspension")
	}
	return out, ValidateRoleRecords(out)
}

func legacyRoleAliasSupports(candidate Candidate, role ValidatorRole) bool {
	if role == ValidatorRoleProposer {
		return ValidatorSupportsRole(candidate, ValidatorRoleBlockProducer)
	}
	if role == ValidatorRoleEvidenceReporter {
		return ValidatorSupportsRole(candidate, ValidatorRoleEvidenceReviewer)
	}
	return false
}

func roleMinBps(value uint64) uint32 {
	if value >= uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(value)
}

func roleMaxUint32(left uint32, right uint32) uint32 {
	if left >= right {
		return left
	}
	return right
}

func AllValidatorRoles() []ValidatorRole {
	return []ValidatorRole{
		ValidatorRoleValidator,
		ValidatorRoleProposer,
		ValidatorRoleBlockProducer,
		ValidatorRoleVerifier,
		ValidatorRoleEvidenceReporter,
		ValidatorRoleDelegationOperator,
		ValidatorRoleCollator,
		ValidatorRoleFisherman,
		ValidatorRoleEvidenceReviewer,
	}
}

func DefaultTaskRoles() []ValidatorRole {
	return []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
}

func validateEpochPhase(phase EpochPhase) error {
	switch phase {
	case EpochPhaseDelegation, EpochPhaseElection, EpochPhaseAssignment, EpochPhaseActive, EpochPhaseSettlement, EpochPhaseClosed:
		return nil
	default:
		return fmt.Errorf("unsupported epoch phase %q", phase)
	}
}

func validateValidatorRole(role ValidatorRole) error {
	switch role {
	case ValidatorRoleValidator,
		ValidatorRoleProposer,
		ValidatorRoleBlockProducer,
		ValidatorRoleVerifier,
		ValidatorRoleEvidenceReporter,
		ValidatorRoleDelegationOperator,
		ValidatorRoleCollator,
		ValidatorRoleFisherman,
		ValidatorRoleEvidenceReviewer:
		return nil
	default:
		return fmt.Errorf("unsupported validator role %q", role)
	}
}

func validateRoleStatus(status string) error {
	switch status {
	case RoleStatusEligible, RoleStatusAssigned, RoleStatusSuspended, RoleStatusInactive:
		return nil
	default:
		return fmt.Errorf("unsupported role status %q", status)
	}
}

func validateCollatorStatus(status string) error {
	switch status {
	case CollatorStatusRegistered, CollatorStatusActive, CollatorStatusSuspended, CollatorStatusRetired:
		return nil
	default:
		return fmt.Errorf("unsupported collator status %q", status)
	}
}

func validateCollatorVerificationResult(result string) error {
	switch result {
	case CollatorVerificationResultValid, CollatorVerificationResultInvalid, CollatorVerificationResultAbstain:
		return nil
	default:
		return fmt.Errorf("unsupported collator verification result %q", result)
	}
}

func validateRiskWindowStatus(status string) error {
	switch status {
	case RiskWindowStatusActive, RiskWindowStatusExited, RiskWindowStatusExpired, RiskWindowStatusSlashed:
		return nil
	default:
		return fmt.Errorf("unsupported risk window status %q", status)
	}
}

func riskWindowStatus(endEpoch uint64, slashableUntilEpoch uint64, currentEpoch uint64) string {
	if currentEpoch > slashableUntilEpoch {
		return RiskWindowStatusExpired
	}
	if currentEpoch >= endEpoch {
		return RiskWindowStatusExited
	}
	return RiskWindowStatusActive
}

func validatePosLayer(layer PosLayer) error {
	switch layer {
	case PosLayerEconomicConsensus, PosLayerTaskAssignment, PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT:
		return nil
	default:
		return fmt.Errorf("unsupported pos layer %q", layer)
	}
}

func validateValidatorRoles(roles []ValidatorRole) error {
	seen := make(map[ValidatorRole]struct{}, len(roles))
	for _, role := range roles {
		if err := validateValidatorRole(role); err != nil {
			return err
		}
		if _, found := seen[role]; found {
			return fmt.Errorf("duplicate validator role %q", role)
		}
		seen[role] = struct{}{}
	}
	return nil
}

func normalizedRoles(roles []ValidatorRole, defaults []ValidatorRole) []ValidatorRole {
	if len(roles) == 0 {
		roles = defaults
	}
	out := make([]ValidatorRole, len(roles))
	copy(out, roles)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func normalizedWorkloadTypes(workloadTypes []WorkloadType) []WorkloadType {
	out := make([]WorkloadType, len(workloadTypes))
	copy(out, workloadTypes)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func normalizeWorkloadTask(params Params, task WorkloadTask) WorkloadTask {
	out := task
	if out.WorkloadID == "" {
		out.WorkloadID = out.TaskID
	}
	if out.WorkloadType == "" {
		out.WorkloadType = WorkloadTypeServiceValidation
	}
	if out.WorkloadClass == "" {
		out.WorkloadClass = DefaultWorkloadClass
	}
	if out.RequiredValidators == 0 {
		out.RequiredValidators = params.MinTaskGroupValidators
	}
	if len(out.Roles) == 0 {
		out.Roles = DefaultTaskRoles()
	}
	out.Roles = normalizedRoles(out.Roles, DefaultTaskRoles())
	return out
}

func cloneValidatorCapacity(capacity ValidatorCapacity) ValidatorCapacity {
	out := capacity
	out.SupportedWorkloads = make([]WorkloadType, len(capacity.SupportedWorkloads))
	copy(out.SupportedWorkloads, capacity.SupportedWorkloads)
	out.ZoneSupport = make([]string, len(capacity.ZoneSupport))
	copy(out.ZoneSupport, capacity.ZoneSupport)
	return out
}

func validateWorkloadType(workloadType WorkloadType) error {
	switch workloadType {
	case WorkloadTypeGlobalConsensus,
		WorkloadTypeZoneExecution,
		WorkloadTypeShardExecution,
		WorkloadTypeProofVerification,
		WorkloadTypeEvidenceVerification,
		WorkloadTypeDataAvailability,
		WorkloadTypeServiceValidation:
		return nil
	default:
		return fmt.Errorf("unsupported workload type %q", workloadType)
	}
}

func validateWorkloadTypes(workloadTypes []WorkloadType) error {
	seen := make(map[WorkloadType]struct{}, len(workloadTypes))
	for _, workloadType := range workloadTypes {
		if err := validateWorkloadType(workloadType); err != nil {
			return err
		}
		if _, found := seen[workloadType]; found {
			return fmt.Errorf("duplicate supported workload %q", workloadType)
		}
		seen[workloadType] = struct{}{}
	}
	return nil
}

func validateZoneSupport(zones []string) error {
	seen := make(map[string]struct{}, len(zones))
	for _, zone := range zones {
		if err := validatePosToken("zone support", zone); err != nil {
			return err
		}
		if _, found := seen[zone]; found {
			return fmt.Errorf("duplicate zone support %q", zone)
		}
		seen[zone] = struct{}{}
	}
	return nil
}

func validateExcludedValidators(validatorIDs []string) error {
	seen := make(map[string]struct{}, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		if err := validatePosToken("excluded validator id", validatorID); err != nil {
			return err
		}
		if _, found := seen[validatorID]; found {
			return fmt.Errorf("duplicate excluded validator %q", validatorID)
		}
		seen[validatorID] = struct{}{}
	}
	return nil
}

func validatorSetMap(fieldName string, validatorIDs []string) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		if err := validatePosToken(fieldName, validatorID); err != nil {
			return nil, err
		}
		if _, found := seen[validatorID]; found {
			return nil, fmt.Errorf("duplicate %s %q", fieldName, validatorID)
		}
		seen[validatorID] = struct{}{}
	}
	return seen, nil
}

func validatorsForTaskRole(validators []ScoredValidator, task WorkloadTask, role ValidatorRole, assignedTaskKeys map[string]map[string]struct{}, taskKey string) []ScoredValidator {
	out := make([]ScoredValidator, 0, len(validators))
	for _, validator := range validators {
		if ValidatorSupportsRole(validator.Candidate, role) &&
			!isExcludedValidator(validator.ValidatorID, task.ExcludedValidators) &&
			validator.Capacity.SupportsAssignment(task, assignedTaskKeys, validator.ValidatorID, taskKey) {
			out = append(out, validator)
		}
	}
	return out
}

func isExcludedValidator(validatorID string, excluded []string) bool {
	for _, excludedID := range excluded {
		if excludedID == validatorID {
			return true
		}
	}
	return false
}

func selectTaskValidatorIDs(seed string, task WorkloadTask, role ValidatorRole, validators []ScoredValidator, required uint32) []string {
	type rankedValidator struct {
		validatorID	string
		rankHash	string
		score		sdkmath.Int
	}
	ranked := make([]rankedValidator, len(validators))
	for i, validator := range validators {
		ranked[i] = rankedValidator{
			validatorID:	validator.ValidatorID,
			score:		validator.Score,
			rankHash: posHashRoot("aetheris-pos-task-rank-v1", func(w posByteWriter) {
				posWritePart(w, seed)
				posWritePart(w, task.TaskID)
				posWritePart(w, task.WorkloadID)
				posWritePart(w, string(task.WorkloadType))
				posWritePart(w, task.ZoneID)
				posWritePart(w, task.ShardID)
				posWritePart(w, string(role))
				posWritePart(w, validator.ValidatorID)
				posWritePart(w, validator.Score.String())
				posWritePart(w, validator.VotingPowerNaet.String())
			}),
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].rankHash != ranked[j].rankHash {
			return ranked[i].rankHash < ranked[j].rankHash
		}
		if !ranked[i].score.Equal(ranked[j].score) {
			return ranked[i].score.GT(ranked[j].score)
		}
		return ranked[i].validatorID < ranked[j].validatorID
	})
	selected := make([]string, 0, required)
	for i := uint32(0); i < required; i++ {
		selected = append(selected, ranked[i].validatorID)
	}
	sort.Strings(selected)
	return selected
}

func taskKey(task WorkloadTask) string {
	return strings.Join([]string{task.TaskID, task.WorkloadID, string(task.WorkloadType), task.ZoneID, task.ShardID, task.WorkloadClass}, "|")
}

func markTaskGroupAssignments(assignedTaskKeys map[string]map[string]struct{}, taskKey string, validatorIDs []string) {
	for _, validatorID := range validatorIDs {
		if _, found := assignedTaskKeys[validatorID]; !found {
			assignedTaskKeys[validatorID] = make(map[string]struct{})
		}
		assignedTaskKeys[validatorID][taskKey] = struct{}{}
	}
}

func taskGroupMembers(assignments []TaskAssignment) []string {
	seen := make(map[string]struct{})
	for _, assignment := range assignments {
		for _, validatorID := range assignment.Validators {
			seen[validatorID] = struct{}{}
		}
	}
	members := make([]string, 0, len(seen))
	for validatorID := range seen {
		members = append(members, validatorID)
	}
	sort.Strings(members)
	return members
}

func taskGroupVerifiers(assignments []TaskAssignment, members []string) []string {
	seen := make(map[string]struct{})
	for _, assignment := range assignments {
		if assignment.Role == ValidatorRoleVerifier || assignment.Role == ValidatorRoleEvidenceReviewer {
			for _, validatorID := range assignment.Validators {
				seen[validatorID] = struct{}{}
			}
		}
	}
	if len(seen) == 0 {
		for _, validatorID := range members {
			seen[validatorID] = struct{}{}
		}
	}
	verifiers := make([]string, 0, len(seen))
	for validatorID := range seen {
		verifiers = append(verifiers, validatorID)
	}
	sort.Strings(verifiers)
	return verifiers
}

func taskGroupProposerOrder(seed string, task WorkloadTask, members []string) []string {
	type proposerRank struct {
		validatorID	string
		hash		string
	}
	ranks := make([]proposerRank, len(members))
	for i, validatorID := range members {
		ranks[i] = proposerRank{
			validatorID:	validatorID,
			hash: posHashRoot("aetheris-pos-task-group-proposer-v1", func(w posByteWriter) {
				posWritePart(w, seed)
				posWritePart(w, task.TaskID)
				posWritePart(w, task.WorkloadID)
				posWritePart(w, string(task.WorkloadType))
				posWritePart(w, validatorID)
			}),
		}
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		if ranks[i].hash != ranks[j].hash {
			return ranks[i].hash < ranks[j].hash
		}
		return ranks[i].validatorID < ranks[j].validatorID
	})
	out := make([]string, len(ranks))
	for i, rank := range ranks {
		out[i] = rank.validatorID
	}
	return out
}

func validateCanonicalValidatorIDs(fieldName string, validatorIDs []string) error {
	if len(validatorIDs) == 0 {
		return fmt.Errorf("%s ids are required", fieldName)
	}
	seen := make(map[string]struct{}, len(validatorIDs))
	var previous string
	for i, validatorID := range validatorIDs {
		if err := validatePosToken(fieldName+" id", validatorID); err != nil {
			return err
		}
		if _, found := seen[validatorID]; found {
			return fmt.Errorf("duplicate %s id %q", fieldName, validatorID)
		}
		seen[validatorID] = struct{}{}
		if i > 0 && previous >= validatorID {
			return fmt.Errorf("%s ids must be sorted canonically", fieldName)
		}
		previous = validatorID
	}
	return nil
}

func validateValidatorIDSet(fieldName string, values []string, expectedMembers []string) error {
	if len(values) != len(expectedMembers) {
		return fmt.Errorf("%s ids must include every task group member", fieldName)
	}
	if err := validateValidatorIDSubset(fieldName, values, expectedMembers); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s id %q", fieldName, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateValidatorIDSubset(fieldName string, values []string, members []string) error {
	memberSet := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberSet[member] = struct{}{}
	}
	for _, value := range values {
		if _, found := memberSet[value]; !found {
			return fmt.Errorf("%s id %q is not a task group member", fieldName, value)
		}
	}
	return nil
}

func compareWorkloadTasks(left, right WorkloadTask) int {
	if left.TaskID < right.TaskID {
		return -1
	}
	if left.TaskID > right.TaskID {
		return 1
	}
	if left.WorkloadID < right.WorkloadID {
		return -1
	}
	if left.WorkloadID > right.WorkloadID {
		return 1
	}
	if left.WorkloadType < right.WorkloadType {
		return -1
	}
	if left.WorkloadType > right.WorkloadType {
		return 1
	}
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.ShardID < right.ShardID {
		return -1
	}
	if left.ShardID > right.ShardID {
		return 1
	}
	if left.WorkloadClass < right.WorkloadClass {
		return -1
	}
	if left.WorkloadClass > right.WorkloadClass {
		return 1
	}
	return 0
}

func compareTaskAssignments(left, right TaskAssignment) int {
	if cmp := compareWorkloadTasks(
		WorkloadTask{TaskID: left.TaskID, WorkloadID: left.WorkloadID, WorkloadType: left.WorkloadType, ZoneID: left.ZoneID, ShardID: left.ShardID, WorkloadClass: left.WorkloadClass},
		WorkloadTask{TaskID: right.TaskID, WorkloadID: right.WorkloadID, WorkloadType: right.WorkloadType, ZoneID: right.ZoneID, ShardID: right.ShardID, WorkloadClass: right.WorkloadClass},
	); cmp != 0 {
		return cmp
	}
	if left.Role < right.Role {
		return -1
	}
	if left.Role > right.Role {
		return 1
	}
	return 0
}

func compareTaskGroups(left, right TaskGroup) int {
	if left.EpochID < right.EpochID {
		return -1
	}
	if left.EpochID > right.EpochID {
		return 1
	}
	if left.WorkloadID < right.WorkloadID {
		return -1
	}
	if left.WorkloadID > right.WorkloadID {
		return 1
	}
	if left.WorkloadType < right.WorkloadType {
		return -1
	}
	if left.WorkloadType > right.WorkloadType {
		return 1
	}
	if left.TaskGroupID < right.TaskGroupID {
		return -1
	}
	if left.TaskGroupID > right.TaskGroupID {
		return 1
	}
	return 0
}

func compareDelegationIntents(left, right DelegationIntent) int {
	if left.ValidatorID < right.ValidatorID {
		return -1
	}
	if left.ValidatorID > right.ValidatorID {
		return 1
	}
	if left.NominatorID < right.NominatorID {
		return -1
	}
	if left.NominatorID > right.NominatorID {
		return 1
	}
	if left.RequestedEpoch < right.RequestedEpoch {
		return -1
	}
	if left.RequestedEpoch > right.RequestedEpoch {
		return 1
	}
	return 0
}

func computeDelegationActivationKey(epoch uint64, validatorID string, nominations []Nomination) string {
	return posHashRoot("aetheris-pos-delegation-activation-v1", func(w posByteWriter) {
		posWriteUint64(w, epoch)
		posWritePart(w, validatorID)
		posWriteUint64(w, uint64(len(nominations)))
		for _, nomination := range nominations {
			posWritePart(w, nomination.NominatorID)
			posWritePart(w, nomination.StakeNaet.String())
		}
	})
}

func computeEvidenceSettlementHash(settlement EvidenceSettlement) string {
	return posHashRoot("aetheris-pos-evidence-settlement-v1", func(w posByteWriter) {
		posWritePart(w, settlement.EvidenceID)
		posWritePart(w, settlement.ReporterID)
		posWritePart(w, settlement.Slash.ValidatorID)
		posWritePart(w, settlement.Slash.Misbehavior)
		posWritePart(w, settlement.Slash.TotalSlashedNaet.String())
		posWritePart(w, settlement.ReporterRewardNaet.String())
		posWritePart(w, settlement.BurnNaet.String())
		posWriteUint64(w, uint64(settlement.Slash.EvidenceHeight))
	})
}

func computeStructuredEvidenceHash(evidence StructuredEvidenceRecord) string {
	return posHashRoot("aetheris-pos-structured-evidence-v1", func(w posByteWriter) {
		posWritePart(w, evidence.EvidenceID)
		posWritePart(w, evidence.EvidenceType)
		posWritePart(w, evidence.ReporterID)
		posWritePart(w, evidence.AccusedValidatorID)
		posWritePart(w, evidence.SubjectID)
		posWritePart(w, evidence.EvidenceHash)
		posWritePart(w, fmt.Sprintf("%d", evidence.EvidenceHeight))
		posWriteUint64(w, evidence.EvidenceEpoch)
		posWritePart(w, fmt.Sprintf("%d", evidence.SubmittedHeight))
		posWritePart(w, evidence.VerificationGroupID)
		posWritePart(w, evidence.Status)
	})
}

func computeEvidenceRecordHash(record EvidenceRecord) string {
	return posHashRoot("aetheris-pos-evidence-record-v1", func(w posByteWriter) {
		posWritePart(w, record.EvidenceID)
		posWritePart(w, record.EvidenceType)
		posWritePart(w, record.AccusedValidator)
		posWritePart(w, record.Reporter)
		posWriteUint64(w, record.EpochID)
		posWritePart(w, record.TaskGroupIDOptional)
		posWritePart(w, record.ObjectHash)
		posWritePart(w, record.ProofPayloadHash)
		posWritePart(w, fmt.Sprintf("%d", record.SubmittedHeight))
		posWritePart(w, record.Status)
		posWritePart(w, record.VerificationGroupID)
		posWritePart(w, fmt.Sprintf("%d", record.DecisionHeight))
		posWritePart(w, record.PenaltyIDOptional)
	})
}

func computeEvidenceVerificationAssignmentSeed(epochSeed string, evidenceID string) string {
	return posHashRoot("aetheris-pos-evidence-verification-seed-v1", func(w posByteWriter) {
		posWritePart(w, epochSeed)
		posWritePart(w, evidenceID)
	})
}

func computeEvidenceVerifierSelectionHash(epochSeed string, evidenceID string, validatorID string) string {
	return posHashRoot("aetheris-pos-evidence-verifier-rank-v1", func(w posByteWriter) {
		posWritePart(w, epochSeed)
		posWritePart(w, evidenceID)
		posWritePart(w, validatorID)
	})
}

func computeEvidenceVerificationGroupID(group EvidenceVerificationGroup) string {
	return posHashRoot("aetheris-pos-evidence-verification-group-id-v1", func(w posByteWriter) {
		posWritePart(w, group.EvidenceID)
		posWriteUint64(w, group.EpochID)
		posWritePart(w, group.AssignmentSeed)
		posWriteUint64(w, uint64(group.MinimumGroupSize))
		posWriteUint64(w, uint64(group.DecisionThresholdBps))
	})
}

func computeEvidenceVerificationGroupHash(group EvidenceVerificationGroup) string {
	return posHashRoot("aetheris-pos-evidence-verification-group-v1", func(w posByteWriter) {
		posWritePart(w, group.EvidenceID)
		posWriteUint64(w, group.EpochID)
		posWritePart(w, group.VerificationGroupID)
		posWriteUint64(w, uint64(group.MinimumGroupSize))
		posWriteUint64(w, uint64(group.DecisionThresholdBps))
		posWritePart(w, group.AssignmentSeed)
		posWriteUint64(w, uint64(len(group.Members)))
		for _, member := range group.Members {
			posWritePart(w, member)
		}
		posWriteUint64(w, uint64(len(group.ExcludedValidators)))
		for _, excluded := range group.ExcludedValidators {
			posWritePart(w, excluded)
		}
	})
}

func computeEvidenceVerificationRoot(evidenceID string, votes []EvidenceVerificationVote) string {
	ordered := make([]EvidenceVerificationVote, len(votes))
	copy(ordered, votes)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ReviewerID != ordered[j].ReviewerID {
			return ordered[i].ReviewerID < ordered[j].ReviewerID
		}
		return ordered[i].VoteHeight < ordered[j].VoteHeight
	})
	return posHashRoot("aetheris-pos-evidence-verification-root-v1", func(w posByteWriter) {
		posWritePart(w, evidenceID)
		posWriteUint64(w, uint64(len(ordered)))
		for _, vote := range ordered {
			posWritePart(w, vote.ReviewerID)
			posWritePart(w, fmt.Sprintf("%t", vote.Accepted))
			posWritePart(w, vote.SignatureHash)
			posWritePart(w, fmt.Sprintf("%d", vote.VoteHeight))
		}
	})
}

func computeEvidenceFinalityVoteRoot(evidenceID string, votes []EvidenceFinalityVote) string {
	ordered := make([]EvidenceFinalityVote, len(votes))
	copy(ordered, votes)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ValidatorID != ordered[j].ValidatorID {
			return ordered[i].ValidatorID < ordered[j].ValidatorID
		}
		return ordered[i].FinalityHeight < ordered[j].FinalityHeight
	})
	return posHashRoot("aetheris-pos-evidence-finality-root-v1", func(w posByteWriter) {
		posWritePart(w, evidenceID)
		posWriteUint64(w, uint64(len(ordered)))
		for _, vote := range ordered {
			posWritePart(w, vote.ValidatorID)
			posWritePart(w, fmt.Sprintf("%t", vote.Approve))
			posWriteUint64(w, uint64(vote.VotingPowerBps))
			posWritePart(w, vote.SignatureHash)
			posWritePart(w, fmt.Sprintf("%d", vote.FinalityHeight))
		}
	})
}

func computeWorkloadRewardRoot(settlement WorkloadRewardSettlement) string {
	return posHashRoot("aetheris-pos-workload-reward-root-v1", func(w posByteWriter) {
		posWriteUint64(w, settlement.EpochID)
		posWriteUint64(w, uint64(len(settlement.Rewards)))
		for _, reward := range settlement.Rewards {
			posWritePart(w, reward.ValidatorID)
			posWritePart(w, reward.RewardNaet.String())
			posWriteUint64(w, reward.WorkUnits)
		}
		posWritePart(w, settlement.RemainderNaet.String())
		posWriteUint64(w, settlement.CompletedUnits)
	})
}

func validateRoleRewardWeights(weights []RoleRewardWeight) error {
	if len(weights) == 0 {
		return errors.New("role reward weights are required")
	}
	total := uint64(0)
	seen := make(map[ValidatorRole]struct{}, len(weights))
	for _, weight := range weights {
		if err := validateValidatorRole(weight.Role); err != nil {
			return err
		}
		if _, found := seen[weight.Role]; found {
			return fmt.Errorf("duplicate role reward weight %q", weight.Role)
		}
		seen[weight.Role] = struct{}{}
		total += uint64(weight.WeightBps)
	}
	if total != uint64(BasisPoints) {
		return fmt.Errorf("role reward weights must sum to %d bps", BasisPoints)
	}
	return nil
}

func validatePerformanceRewardBps(values ...uint32) error {
	names := []string{"uptime score", "latency score", "correctness score", "task completion rate"}
	for i, value := range values {
		name := "performance reward component"
		if i < len(names) {
			name = names[i]
		}
		if value > BasisPoints {
			return fmt.Errorf("%s must be <= %d bps", name, BasisPoints)
		}
	}
	return nil
}

func computeRewardMultiplierBps(uptimeScoreBps uint32, latencyScoreBps uint32, correctnessScoreBps uint32, taskCompletionRateBps uint32) (uint32, error) {
	if err := validatePerformanceRewardBps(uptimeScoreBps, latencyScoreBps, correctnessScoreBps, taskCompletionRateBps); err != nil {
		return 0, err
	}
	multiplier := mulBps(uptimeScoreBps, latencyScoreBps)
	multiplier = mulBps(multiplier, correctnessScoreBps)
	multiplier = mulBps(multiplier, taskCompletionRateBps)
	return multiplier, nil
}

func mulBps(value uint32, multiplier uint32) uint32 {
	return uint32((uint64(value) * uint64(multiplier)) / uint64(BasisPoints))
}

func checkedAddUint64(left uint64, right uint64, message string) (uint64, error) {
	if ^uint64(0)-left < right {
		return 0, errors.New(message)
	}
	return left + right, nil
}

func mulUint64Overflow(left uint64, right uint64) (uint64, bool) {
	if left == 0 || right == 0 {
		return 0, false
	}
	if left > ^uint64(0)/right {
		return 0, true
	}
	return left * right, false
}

func ceilDivUint64(value uint64, divisor uint64) uint64 {
	if divisor == 0 {
		return 0
	}
	if value == 0 {
		return 0
	}
	return 1 + (value-1)/divisor
}

func isEvidenceStatus(status string) bool {
	switch status {
	case EvidenceStatusSubmitted,
		EvidenceStatusInVerification,
		EvidenceStatusAccepted,
		EvidenceStatusVerified,
		EvidenceStatusRejected,
		EvidenceStatusExpired,
		EvidenceStatusFinalized,
		EvidenceStatusSlashed:
		return true
	default:
		return false
	}
}

func isEvidenceRecordStatus(status string) bool {
	switch status {
	case EvidenceStatusSubmitted, EvidenceStatusInVerification, EvidenceStatusAccepted, EvidenceStatusRejected, EvidenceStatusExpired, EvidenceStatusSlashed:
		return true
	default:
		return false
	}
}

func isAllowedEvidenceRecordTransition(current string, next string) bool {
	if current == next {
		return true
	}
	switch current {
	case EvidenceStatusSubmitted:
		return next == EvidenceStatusInVerification || next == EvidenceStatusExpired
	case EvidenceStatusInVerification:
		return next == EvidenceStatusAccepted || next == EvidenceStatusRejected || next == EvidenceStatusExpired
	case EvidenceStatusAccepted:
		return next == EvidenceStatusSlashed
	default:
		return false
	}
}

func evidenceVerificationExclusions(evidence EvidenceRecord, validators []ScoredValidator) []string {
	validatorIDs := make(map[string]struct{}, len(validators))
	for _, validator := range validators {
		validatorIDs[validator.ValidatorID] = struct{}{}
	}
	excluded := make([]string, 0, 2)
	if _, found := validatorIDs[evidence.AccusedValidator]; found {
		excluded = append(excluded, evidence.AccusedValidator)
	}
	if _, found := validatorIDs[evidence.Reporter]; found && evidence.Reporter != evidence.AccusedValidator {
		excluded = append(excluded, evidence.Reporter)
	}
	sort.Strings(excluded)
	return excluded
}

func validateSortedUniqueTokens(fieldName string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	var previous string
	for i, value := range values {
		if err := validatePosToken(fieldName, value); err != nil {
			return err
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %q", fieldName, value)
		}
		seen[value] = struct{}{}
		if i > 0 && previous >= value {
			return fmt.Errorf("%s values must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func cloneStringSlice(values []string) []string {
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func validateEvidenceReviewers(reviewers []string) (map[string]struct{}, error) {
	if len(reviewers) == 0 {
		return nil, errors.New("evidence verification reviewers are required")
	}
	out := make(map[string]struct{}, len(reviewers))
	for _, reviewerID := range reviewers {
		if err := validatePosToken("evidence verification reviewer id", reviewerID); err != nil {
			return nil, err
		}
		if _, found := out[reviewerID]; found {
			return nil, fmt.Errorf("duplicate evidence verification reviewer %q", reviewerID)
		}
		out[reviewerID] = struct{}{}
	}
	return out, nil
}

func sortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mergeNominations(existing []Nomination, activated []Nomination) []Nomination {
	byNominator := make(map[string]sdkmath.Int, len(existing)+len(activated))
	for _, nomination := range existing {
		current, found := byNominator[nomination.NominatorID]
		if !found {
			current = sdkmath.ZeroInt()
		}
		byNominator[nomination.NominatorID] = current.Add(nomination.StakeNaet)
	}
	for _, nomination := range activated {
		current, found := byNominator[nomination.NominatorID]
		if !found {
			current = sdkmath.ZeroInt()
		}
		byNominator[nomination.NominatorID] = current.Add(nomination.StakeNaet)
	}
	nominatorIDs := sortedStringKeys(byNominator)
	out := make([]Nomination, 0, len(nominatorIDs))
	for _, nominatorID := range nominatorIDs {
		out = append(out, Nomination{NominatorID: nominatorID, StakeNaet: byNominator[nominatorID]})
	}
	return out
}

func ratioBps(numerator uint64, denominator uint64) uint32 {
	if denominator == 0 {
		return BasisPoints
	}
	if numerator >= denominator {
		return BasisPoints
	}
	return uint32((uint64(BasisPoints) * numerator) / denominator)
}

func intRatioBps(numerator sdkmath.Int, denominator sdkmath.Int) uint32 {
	if denominator.IsNil() || !denominator.IsPositive() {
		return BasisPoints
	}
	if numerator.IsNil() || !numerator.IsPositive() {
		return 0
	}
	if numerator.GTE(denominator) {
		return BasisPoints
	}
	return uint32(numerator.MulRaw(int64(BasisPoints)).Quo(denominator).Uint64())
}

func stateKey(parts ...string) string {
	return strings.Join(parts, "/")
}

func stateKeyChecked(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", errors.New("state key parts are required")
	}
	for _, part := range parts {
		if err := validateStateKeyComponent("state key component", part); err != nil {
			return "", err
		}
	}
	return stateKey(parts...), nil
}

func uint64StateComponent(value uint64) string {
	return fmt.Sprintf("%d", value)
}

func validateStateKeyComponent(fieldName string, value string) error {
	if err := validatePosToken(fieldName, value); err != nil {
		return err
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("%s must not contain path separator", fieldName)
	}
	return nil
}

func validatePosToken(fieldName string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxPosTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxPosTokenLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func validatePosResponsibility(fieldName string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxPosTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxPosTokenLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' || r == ' ' || r == '+' || r == ',' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func validatePosHash(fieldName string, value string) error {
	if len(value) != PosHashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, PosHashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, PosHashHexLength)
	}
	return nil
}

type posByteWriter interface {
	Write([]byte) (int, error)
}

func posHashRoot(domain string, write func(posByteWriter)) string {
	h := sha256.New()
	posWritePart(h, domain)
	write(h)
	return hex.EncodeToString(h.Sum(nil))
}

func posWritePart(w posByteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func posWriteStringSlice(w posByteWriter, values []string) {
	posWriteUint64(w, uint64(len(values)))
	for _, value := range values {
		posWritePart(w, value)
	}
}

func posWriteHookSpecs(w posByteWriter, hooks []KeeperHookSpec) {
	posWriteUint64(w, uint64(len(hooks)))
	for _, hook := range hooks {
		posWritePart(w, hook.SourceKeeper)
		posWritePart(w, hook.HookName)
		posWritePart(w, hook.Trigger)
		posWriteStringSlice(w, hook.TargetModules)
		posWriteUint64(w, boolAsUint64(hook.PreservesBaseState))
		posWriteUint64(w, boolAsUint64(hook.DeterministicOrder))
	}
}

func posWriteUint64(w posByteWriter, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}

func boolAsUint64(value bool) uint64 {
	if value {
		return 1
	}
	return 0
}
