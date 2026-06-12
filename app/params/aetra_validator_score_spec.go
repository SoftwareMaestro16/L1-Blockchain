package params

import (
	"fmt"
	"sort"
)

const (
	AetraValidatorScoreModuleName	= "x/aetra-validator-score"

	AetraValidatorScorePurposePublicAccountability	= "public_accountability_without_subjective_consensus_control"

	AetraValidatorScoreResponsibilityTrackUptime			= "track_validator_uptime"
	AetraValidatorScoreResponsibilityTrackMissedBlockWindows	= "track_missed_block_windows"
	AetraValidatorScoreResponsibilityTrackJailHistory		= "track_jail_history"
	AetraValidatorScoreResponsibilityTrackSlashingHistory		= "track_slashing_history"
	AetraValidatorScoreResponsibilityTrackCommissionBehavior	= "track_commission_behavior"
	AetraValidatorScoreResponsibilityTrackSelfBondRatio		= "track_self_bond_ratio"
	AetraValidatorScoreResponsibilityTrackGovernanceParticipation	= "track_governance_participation"
	AetraValidatorScoreResponsibilityTrackConcentrationStatus	= "track_concentration_status"
	AetraValidatorScoreResponsibilityProducePublicScore		= "produce_public_score"
	AetraValidatorScoreResponsibilityExplorerFriendlyQueries	= "expose_explorer_friendly_queries"

	AetraValidatorScoreGuardNoSubjectiveCensorship		= "score_must_not_be_subjective_censorship_mechanism"
	AetraValidatorScoreGuardInformationalFirst		= "score_informational_first"
	AetraValidatorScoreGuardObjectiveRewardOnly		= "reward_affecting_only_from_objective_chain_data"
	AetraValidatorScoreGuardConsensusOverrideDisabled	= "consensus_override_disabled_by_default"
	AetraValidatorScoreGuardObjectiveInputsDeterministic	= "objective_inputs_must_be_deterministic"

	AetraValidatorScoreStateParams		= "Params"
	AetraValidatorScoreStateValidatorScore	= "ValidatorScore"

	AetraValidatorScoreStateParamUptimeWindow		= "UptimeWindow"
	AetraValidatorScoreStateParamUptimeWeightBps		= "UptimeWeightBps"
	AetraValidatorScoreStateParamSlashHistoryWeightBps	= "SlashHistoryWeightBps"
	AetraValidatorScoreStateParamGovernanceWeightBps	= "GovernanceWeightBps"
	AetraValidatorScoreStateParamSelfBondWeightBps		= "SelfBondWeightBps"
	AetraValidatorScoreStateParamConcentrationWeightBps	= "ConcentrationWeightBps"
	AetraValidatorScoreStateParamMinScore			= "MinScore"
	AetraValidatorScoreStateParamMaxScore			= "MaxScore"
	AetraValidatorScoreStateParamRewardModifierEnabled	= "RewardModifierEnabled"
	AetraValidatorScoreStateParamMaxRewardPenaltyBps	= "MaxRewardPenaltyBps"

	AetraValidatorScoreStateScoreOperatorAddress	= "OperatorAddress"
	AetraValidatorScoreStateScoreScore		= "Score"
	AetraValidatorScoreStateScoreUptimeScore	= "UptimeScore"
	AetraValidatorScoreStateScoreSlashScore		= "SlashScore"
	AetraValidatorScoreStateScoreGovernanceScore	= "GovernanceScore"
	AetraValidatorScoreStateScoreSelfBondScore	= "SelfBondScore"
	AetraValidatorScoreStateScoreConcentrationScore	= "ConcentrationScore"
	AetraValidatorScoreStateScoreMissedBlocks	= "MissedBlocks"
	AetraValidatorScoreStateScoreSignedBlocks	= "SignedBlocks"
	AetraValidatorScoreStateScoreJailCount		= "JailCount"
	AetraValidatorScoreStateScoreSlashCount		= "SlashCount"
	AetraValidatorScoreStateScoreLastUpdatedHeight	= "LastUpdatedHeight"

	AetraValidatorScoreRequirementDeterministic			= "score_deterministic"
	AetraValidatorScoreRequirementChainStateOnly			= "score_based_only_on_chain_state"
	AetraValidatorScoreRequirementExplainable			= "score_explainable"
	AetraValidatorScoreRequirementQueryable				= "score_queryable"
	AetraValidatorScoreRequirementBounded				= "score_bounded"
	AetraValidatorScoreRequirementExportImportSafe			= "score_export_import_safe"
	AetraValidatorScoreRequirementOverflowUnderflowResistant	= "score_resistant_to_overflow_underflow"
	AetraValidatorScoreRequiredTestPerfectUptime			= "perfect_uptime_score"
	AetraValidatorScoreRequiredTestPartialUptime			= "partial_uptime_score"
	AetraValidatorScoreRequiredTestMissedBlockPenalty		= "missed_block_penalty"
	AetraValidatorScoreRequiredTestJailPenalty			= "jail_penalty"
	AetraValidatorScoreRequiredTestSlashPenalty			= "slash_penalty"
	AetraValidatorScoreRequiredTestGovernanceParticipation		= "governance_participation_score"
	AetraValidatorScoreRequiredTestConcentrationPenalty		= "concentration_penalty"
	AetraValidatorScoreRequiredTestRewardModifierBounded		= "reward_modifier_bounded"
	AetraValidatorScoreRequiredTestScoreCannotGoBelowMin		= "score_cannot_go_below_min"
	AetraValidatorScoreRequiredTestScoreCannotExceedMax		= "score_cannot_exceed_max"
	AetraValidatorScoreRequiredTestExportImport			= "export_import"
	AetraValidatorScoreRequiredTestDeterministicRecomputation	= "deterministic_recomputation"
)

type AetraValidatorScoreSpecEvidence struct {
	ModuleName	string

	PublicAccountabilityWithoutSubjectiveConsensusControl	bool
}

type AetraValidatorScoreSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraValidatorScoreResponsibilitiesEvidence struct {
	ModuleName	string

	TracksValidatorUptime		bool
	TracksMissedBlockWindows	bool
	TracksJailHistory		bool
	TracksSlashingHistory		bool
	TracksCommissionBehavior	bool
	TracksSelfBondRatio		bool
	TracksGovernanceParticipation	bool
	TracksConcentrationStatus	bool
	ProducesPublicScore		bool
	ExposesExplorerFriendlyQueries	bool
}

type AetraValidatorScoreResponsibilitiesReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraValidatorScoreSubjectiveControlEvidence struct {
	ModuleName	string

	NoSubjectiveCensorshipMechanism		bool
	InformationalFirst			bool
	RewardAffectingOnlyObjectiveData	bool
	ConsensusOverrideDisabledDefault	bool
	ObjectiveInputsDeterministic		bool
}

type AetraValidatorScoreSubjectiveControlReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraValidatorScoreStateSpecEvidence struct {
	ModuleName	string

	ParamsFields		[]string
	ValidatorScoreFields	[]string
}

type AetraValidatorScoreStateSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraValidatorScoreRequirementsEvidence struct {
	ModuleName	string

	Deterministic			bool
	BasedOnlyOnChainState		bool
	Explainable			bool
	Queryable			bool
	Bounded				bool
	ExportImportSafe		bool
	OverflowUnderflowResistant	bool
}

type AetraValidatorScoreRequirementsReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraValidatorScoreTestingRequirementsEvidence struct {
	ModuleName	string

	PerfectUptimeScore		bool
	PartialUptimeScore		bool
	MissedBlockPenalty		bool
	JailPenalty			bool
	SlashPenalty			bool
	GovernanceParticipationScore	bool
	ConcentrationPenalty		bool
	RewardModifierBounded		bool
	ScoreCannotGoBelowMin		bool
	ScoreCannotExceedMax		bool
	ExportImport			bool
	DeterministicRecomputation	bool
}

type AetraValidatorScoreTestingRequirementsReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraValidatorScoreSpecEvidence() AetraValidatorScoreSpecEvidence {
	return AetraValidatorScoreSpecEvidence{
		ModuleName:	AetraValidatorScoreModuleName,

		PublicAccountabilityWithoutSubjectiveConsensusControl:	true,
	}
}

func ValidateAetraValidatorScoreSpec(evidence AetraValidatorScoreSpecEvidence) error {
	report := BuildAetraValidatorScoreSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreSpecReport(evidence AetraValidatorScoreSpecEvidence) AetraValidatorScoreSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScorePurposePublicAccountability, evidence.PublicAccountabilityWithoutSubjectiveConsensusControl},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreResponsibilitiesEvidence() AetraValidatorScoreResponsibilitiesEvidence {
	return AetraValidatorScoreResponsibilitiesEvidence{
		ModuleName:	AetraValidatorScoreModuleName,

		TracksValidatorUptime:		true,
		TracksMissedBlockWindows:	true,
		TracksJailHistory:		true,
		TracksSlashingHistory:		true,
		TracksCommissionBehavior:	true,
		TracksSelfBondRatio:		true,
		TracksGovernanceParticipation:	true,
		TracksConcentrationStatus:	true,
		ProducesPublicScore:		true,
		ExposesExplorerFriendlyQueries:	true,
	}
}

func ValidateAetraValidatorScoreResponsibilities(evidence AetraValidatorScoreResponsibilitiesEvidence) error {
	report := BuildAetraValidatorScoreResponsibilitiesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score responsibilities failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreResponsibilitiesReport(evidence AetraValidatorScoreResponsibilitiesEvidence) AetraValidatorScoreResponsibilitiesReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreResponsibilityTrackUptime, evidence.TracksValidatorUptime},
		{AetraValidatorScoreResponsibilityTrackMissedBlockWindows, evidence.TracksMissedBlockWindows},
		{AetraValidatorScoreResponsibilityTrackJailHistory, evidence.TracksJailHistory},
		{AetraValidatorScoreResponsibilityTrackSlashingHistory, evidence.TracksSlashingHistory},
		{AetraValidatorScoreResponsibilityTrackCommissionBehavior, evidence.TracksCommissionBehavior},
		{AetraValidatorScoreResponsibilityTrackSelfBondRatio, evidence.TracksSelfBondRatio},
		{AetraValidatorScoreResponsibilityTrackGovernanceParticipation, evidence.TracksGovernanceParticipation},
		{AetraValidatorScoreResponsibilityTrackConcentrationStatus, evidence.TracksConcentrationStatus},
		{AetraValidatorScoreResponsibilityProducePublicScore, evidence.ProducesPublicScore},
		{AetraValidatorScoreResponsibilityExplorerFriendlyQueries, evidence.ExposesExplorerFriendlyQueries},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreResponsibilitiesReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreSubjectiveControlEvidence() AetraValidatorScoreSubjectiveControlEvidence {
	return AetraValidatorScoreSubjectiveControlEvidence{
		ModuleName:	AetraValidatorScoreModuleName,

		NoSubjectiveCensorshipMechanism:	true,
		InformationalFirst:			true,
		RewardAffectingOnlyObjectiveData:	true,
		ConsensusOverrideDisabledDefault:	true,
		ObjectiveInputsDeterministic:		true,
	}
}

func ValidateAetraValidatorScoreSubjectiveControl(evidence AetraValidatorScoreSubjectiveControlEvidence) error {
	report := BuildAetraValidatorScoreSubjectiveControlReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score subjective control failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreSubjectiveControlReport(evidence AetraValidatorScoreSubjectiveControlEvidence) AetraValidatorScoreSubjectiveControlReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreGuardNoSubjectiveCensorship, evidence.NoSubjectiveCensorshipMechanism},
		{AetraValidatorScoreGuardInformationalFirst, evidence.InformationalFirst},
		{AetraValidatorScoreGuardObjectiveRewardOnly, evidence.RewardAffectingOnlyObjectiveData},
		{AetraValidatorScoreGuardConsensusOverrideDisabled, evidence.ConsensusOverrideDisabledDefault},
		{AetraValidatorScoreGuardObjectiveInputsDeterministic, evidence.ObjectiveInputsDeterministic},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreSubjectiveControlReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreStateSpecEvidence() AetraValidatorScoreStateSpecEvidence {
	return AetraValidatorScoreStateSpecEvidence{
		ModuleName:	AetraValidatorScoreModuleName,
		ParamsFields: []string{
			AetraValidatorScoreStateParamUptimeWindow,
			AetraValidatorScoreStateParamUptimeWeightBps,
			AetraValidatorScoreStateParamSlashHistoryWeightBps,
			AetraValidatorScoreStateParamGovernanceWeightBps,
			AetraValidatorScoreStateParamSelfBondWeightBps,
			AetraValidatorScoreStateParamConcentrationWeightBps,
			AetraValidatorScoreStateParamMinScore,
			AetraValidatorScoreStateParamMaxScore,
			AetraValidatorScoreStateParamRewardModifierEnabled,
			AetraValidatorScoreStateParamMaxRewardPenaltyBps,
		},
		ValidatorScoreFields: []string{
			AetraValidatorScoreStateScoreOperatorAddress,
			AetraValidatorScoreStateScoreScore,
			AetraValidatorScoreStateScoreUptimeScore,
			AetraValidatorScoreStateScoreSlashScore,
			AetraValidatorScoreStateScoreGovernanceScore,
			AetraValidatorScoreStateScoreSelfBondScore,
			AetraValidatorScoreStateScoreConcentrationScore,
			AetraValidatorScoreStateScoreMissedBlocks,
			AetraValidatorScoreStateScoreSignedBlocks,
			AetraValidatorScoreStateScoreJailCount,
			AetraValidatorScoreStateScoreSlashCount,
			AetraValidatorScoreStateScoreLastUpdatedHeight,
		},
	}
}

func ValidateAetraValidatorScoreStateSpec(evidence AetraValidatorScoreStateSpecEvidence) error {
	report := BuildAetraValidatorScoreStateSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score state spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreStateSpecReport(evidence AetraValidatorScoreStateSpecEvidence) AetraValidatorScoreStateSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	requiredParams := requiredAetraValidatorScoreParamsFields()
	requiredScore := requiredAetraValidatorScoreValidatorScoreFields()

	passedParams, failedParams := validateAetraValidatorScoreCatalog(AetraValidatorScoreStateParams, evidence.ParamsFields, requiredParams)
	passedScore, failedScore := validateAetraValidatorScoreCatalog(AetraValidatorScoreStateValidatorScore, evidence.ValidatorScoreFields, requiredScore)

	failed = append(failed, failedParams...)
	failed = append(failed, failedScore...)
	sort.Strings(failed)
	return AetraValidatorScoreStateSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredParams) + len(requiredScore),
		Passed:		passedParams + passedScore,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraValidatorScoreParamsFields() []string {
	return []string{
		AetraValidatorScoreStateParamUptimeWindow,
		AetraValidatorScoreStateParamUptimeWeightBps,
		AetraValidatorScoreStateParamSlashHistoryWeightBps,
		AetraValidatorScoreStateParamGovernanceWeightBps,
		AetraValidatorScoreStateParamSelfBondWeightBps,
		AetraValidatorScoreStateParamConcentrationWeightBps,
		AetraValidatorScoreStateParamMinScore,
		AetraValidatorScoreStateParamMaxScore,
		AetraValidatorScoreStateParamRewardModifierEnabled,
		AetraValidatorScoreStateParamMaxRewardPenaltyBps,
	}
}

func requiredAetraValidatorScoreValidatorScoreFields() []string {
	return []string{
		AetraValidatorScoreStateScoreOperatorAddress,
		AetraValidatorScoreStateScoreScore,
		AetraValidatorScoreStateScoreUptimeScore,
		AetraValidatorScoreStateScoreSlashScore,
		AetraValidatorScoreStateScoreGovernanceScore,
		AetraValidatorScoreStateScoreSelfBondScore,
		AetraValidatorScoreStateScoreConcentrationScore,
		AetraValidatorScoreStateScoreMissedBlocks,
		AetraValidatorScoreStateScoreSignedBlocks,
		AetraValidatorScoreStateScoreJailCount,
		AetraValidatorScoreStateScoreSlashCount,
		AetraValidatorScoreStateScoreLastUpdatedHeight,
	}
}

func DefaultAetraValidatorScoreRequirementsEvidence() AetraValidatorScoreRequirementsEvidence {
	return AetraValidatorScoreRequirementsEvidence{
		ModuleName:	AetraValidatorScoreModuleName,

		Deterministic:			true,
		BasedOnlyOnChainState:		true,
		Explainable:			true,
		Queryable:			true,
		Bounded:			true,
		ExportImportSafe:		true,
		OverflowUnderflowResistant:	true,
	}
}

func ValidateAetraValidatorScoreRequirements(evidence AetraValidatorScoreRequirementsEvidence) error {
	report := BuildAetraValidatorScoreRequirementsReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreRequirementsReport(evidence AetraValidatorScoreRequirementsEvidence) AetraValidatorScoreRequirementsReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreRequirementDeterministic, evidence.Deterministic},
		{AetraValidatorScoreRequirementChainStateOnly, evidence.BasedOnlyOnChainState},
		{AetraValidatorScoreRequirementExplainable, evidence.Explainable},
		{AetraValidatorScoreRequirementQueryable, evidence.Queryable},
		{AetraValidatorScoreRequirementBounded, evidence.Bounded},
		{AetraValidatorScoreRequirementExportImportSafe, evidence.ExportImportSafe},
		{AetraValidatorScoreRequirementOverflowUnderflowResistant, evidence.OverflowUnderflowResistant},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreRequirementsReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraValidatorScoreTestingRequirementsEvidence() AetraValidatorScoreTestingRequirementsEvidence {
	return AetraValidatorScoreTestingRequirementsEvidence{
		ModuleName:	AetraValidatorScoreModuleName,

		PerfectUptimeScore:		true,
		PartialUptimeScore:		true,
		MissedBlockPenalty:		true,
		JailPenalty:			true,
		SlashPenalty:			true,
		GovernanceParticipationScore:	true,
		ConcentrationPenalty:		true,
		RewardModifierBounded:		true,
		ScoreCannotGoBelowMin:		true,
		ScoreCannotExceedMax:		true,
		ExportImport:			true,
		DeterministicRecomputation:	true,
	}
}

func ValidateAetraValidatorScoreTestingRequirements(evidence AetraValidatorScoreTestingRequirementsEvidence) error {
	report := BuildAetraValidatorScoreTestingRequirementsReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator score testing requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorScoreTestingRequirementsReport(evidence AetraValidatorScoreTestingRequirementsEvidence) AetraValidatorScoreTestingRequirementsReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraValidatorScoreModuleName {
		failed = append(failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	}

	checks := []requirementCheck{
		{AetraValidatorScoreRequiredTestPerfectUptime, evidence.PerfectUptimeScore},
		{AetraValidatorScoreRequiredTestPartialUptime, evidence.PartialUptimeScore},
		{AetraValidatorScoreRequiredTestMissedBlockPenalty, evidence.MissedBlockPenalty},
		{AetraValidatorScoreRequiredTestJailPenalty, evidence.JailPenalty},
		{AetraValidatorScoreRequiredTestSlashPenalty, evidence.SlashPenalty},
		{AetraValidatorScoreRequiredTestGovernanceParticipation, evidence.GovernanceParticipationScore},
		{AetraValidatorScoreRequiredTestConcentrationPenalty, evidence.ConcentrationPenalty},
		{AetraValidatorScoreRequiredTestRewardModifierBounded, evidence.RewardModifierBounded},
		{AetraValidatorScoreRequiredTestScoreCannotGoBelowMin, evidence.ScoreCannotGoBelowMin},
		{AetraValidatorScoreRequiredTestScoreCannotExceedMax, evidence.ScoreCannotExceedMax},
		{AetraValidatorScoreRequiredTestExportImport, evidence.ExportImport},
		{AetraValidatorScoreRequiredTestDeterministicRecomputation, evidence.DeterministicRecomputation},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorScoreTestingRequirementsReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func validateAetraValidatorScoreCatalog(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, item := range required {
		requiredSet[item] = true
	}
	seen := map[string]bool{}
	for _, item := range actual {
		if item == "" {
			failed = append(failed, group+".item_required")
			continue
		}
		if seen[item] {
			failed = append(failed, group+"."+item+":duplicate")
			continue
		}
		seen[item] = true
		if !requiredSet[item] {
			failed = append(failed, group+"."+item+":unexpected")
		}
	}
	passed := 0
	for _, item := range required {
		if seen[item] {
			passed++
		} else {
			failed = append(failed, group+"."+item+":missing")
		}
	}
	return passed, failed
}
