package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestComputeSlashingPenaltyScalesStakeExposureBySeverityRoleAndImpact(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:			"penalty-1",
		ValidatorID:			"val-a",
		SeverityLevel:			SlashSeverityHigh,
		StakeExposureNaet:		sdkmath.NewInt(10_000),
		RoleWeightBps:			5_000,
		RepeatOffenseMultiplierBps:	10_000,
		TaskImpactBps:			8_000,
		SafetyImpactBps:		10_000,
		LivenessImpactBps:		10_000,
		SelfStakeNaet:			sdkmath.NewInt(4_000),
		Nominations:			[]Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(6_000)}},
		RewardConfiscationNaet:		sdkmath.NewInt(250),
		TemporaryJailEpochs:		2,
		RoleSuspensions:		[]ValidatorRole{ValidatorRoleVerifier},
		FutureElectionScorePenaltyBps:	1_000,
		EvidenceHeight:			50,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1_200), penalty.ScaledPenaltyBps)
	require.Equal(t, sdkmath.NewInt(1_200), penalty.StakeSlashNaet)
	require.Equal(t, sdkmath.NewInt(480), penalty.ValidatorStakeSlashNaet)
	require.Equal(t, sdkmath.NewInt(720), penalty.DelegatorProportionalSlash)
	require.Equal(t, []NominatorSlash{{NominatorID: "nom-a", SlashedNaet: sdkmath.NewInt(720)}}, penalty.DelegatorSlashes)
	require.Equal(t, sdkmath.NewInt(250), penalty.RewardConfiscationNaet)
	require.Equal(t, uint64(2), penalty.TemporaryJailEpochs)
	require.Equal(t, []ValidatorRole{ValidatorRoleVerifier}, penalty.RoleSuspensions)
	require.Len(t, penalty.PenaltyHash, PosHashHexLength)
	require.NoError(t, penalty.Validate())
}

func TestSlashSeverityClassesUseProtocolNames(t *testing.T) {
	require.Equal(t, []string{
		SlashSeverityMinorLivenessFault,
		SlashSeverityMajorLivenessFault,
		SlashSeverityRepeatedLivenessFault,
		SlashSeverityInvalidTaskExecution,
		SlashSeverityInvalidStateTransition,
		SlashSeverityRepeatedInvalidProposal,
		SlashSeverityRepeatedTimestampViolation,
		SlashSeverityEquivocation,
		SlashSeverityDoubleSign,
		SlashSeverityEvidenceFraud,
	}, SlashSeverityClasses())

	expected := map[string]uint32{
		SlashSeverityMinorLivenessFault:		5,
		SlashSeverityMajorLivenessFault:		25,
		SlashSeverityRepeatedLivenessFault:		100,
		SlashSeverityInvalidTaskExecution:		750,
		SlashSeverityInvalidStateTransition:		1_500,
		SlashSeverityRepeatedInvalidProposal:		25,
		SlashSeverityRepeatedTimestampViolation:	25,
		SlashSeverityEquivocation:			2_000,
		SlashSeverityDoubleSign:			500,
		SlashSeverityEvidenceFraud:			7_500,
	}
	for severity, bps := range expected {
		actual, err := DefaultSeverityBps(severity)
		require.NoError(t, err)
		require.Equal(t, bps, actual)
	}
}

func TestRepeatedObjectiveProposalAndTimestampViolationsCanJailSlashWithoutTombstone(t *testing.T) {
	for _, tc := range []struct {
		name		string
		severity	string
	}{
		{name: "invalid-proposal", severity: SlashSeverityRepeatedInvalidProposal},
		{name: "timestamp-violation", severity: SlashSeverityRepeatedTimestampViolation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := candidate("val-a", 100_000, 0)
			result, err := ExecuteSlashing(SlashingExecutionInput{
				EvidenceID:	tc.name + "-evidence-1",
				ExecutedHeight:	100,
				CurrentEpoch:	10,
				Candidate:	candidate,
				PenaltyInput: SlashingPenaltyInput{
					PenaltyID:		tc.name + "-penalty-1",
					SeverityLevel:		tc.severity,
					StakeExposureNaet:	sdkmath.NewInt(100_000),
					SelfStakeNaet:		candidate.SelfStakeNaet,
					TemporaryJailEpochs:	24,
					EvidenceHeight:		90,
				},
				RoutingInput: SlashingPenaltyRoutingInput{
					BurnBps: BasisPoints,
				},
			})
			require.NoError(t, err)
			require.Equal(t, uint32(25), result.SeverityMatrix[tc.severity])
			require.Equal(t, sdkmath.NewInt(250), result.Record.SlashAmount)
			require.Equal(t, uint64(34), result.Record.JailUntilEpochOptional)
			require.True(t, result.UpdatedCandidate.Jailed)
			require.False(t, result.UpdatedCandidate.Tombstoned)
		})
	}
}

func TestSlashingEvidenceTimingCoversHeightManipulationEdges(t *testing.T) {
	valid := SlashingEvidenceTimingInput{
		FaultHeight:			100,
		EvidenceHeight:			110,
		CurrentHeight:			120,
		MaxEvidenceAgeBlocks:		100,
		ValidatorBondedHeight:		90,
		ValidatorUnbondingHeight:	115,
		UnbondingEvidenceWindowBlocks:	30,
	}
	require.NoError(t, ValidateSlashingEvidenceTiming(valid))

	expired := valid
	expired.EvidenceHeight = 201
	expired.CurrentHeight = 201
	require.ErrorContains(t, ValidateSlashingEvidenceTiming(expired), "expired")

	beforeBonded := valid
	beforeBonded.FaultHeight = 80
	require.ErrorContains(t, ValidateSlashingEvidenceTiming(beforeBonded), "bonded height")

	evidenceBeforeFault := valid
	evidenceBeforeFault.EvidenceHeight = 99
	require.ErrorContains(t, ValidateSlashingEvidenceTiming(evidenceBeforeFault), "precede fault")

	afterUnbondingFault := valid
	afterUnbondingFault.FaultHeight = 115
	afterUnbondingFault.EvidenceHeight = 116
	require.ErrorContains(t, ValidateSlashingEvidenceTiming(afterUnbondingFault), "after validator unbonding")

	lateUnbondingEvidence := valid
	lateUnbondingEvidence.EvidenceHeight = 146
	lateUnbondingEvidence.CurrentHeight = 146
	require.ErrorContains(t, ValidateSlashingEvidenceTiming(lateUnbondingEvidence), "unbonding risk window")
}

func TestProgressiveDowntimeSlashingDefaultsStayWithinAetraRanges(t *testing.T) {
	first, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"downtime-first",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityMinorLivenessFault,
		StakeExposureNaet:	sdkmath.NewInt(100_000),
		SelfStakeNaet:		sdkmath.NewInt(100_000),
		TemporaryJailEpochs:	1,
		EvidenceHeight:		10,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5), first.SeverityBps)
	require.Equal(t, sdkmath.NewInt(50), first.StakeSlashNaet)
	require.Equal(t, uint64(1), first.TemporaryJailEpochs)

	repeat, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"downtime-repeat",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityMajorLivenessFault,
		StakeExposureNaet:	sdkmath.NewInt(100_000),
		SelfStakeNaet:		sdkmath.NewInt(100_000),
		TemporaryJailEpochs:	24,
		EvidenceHeight:		20,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(25), repeat.SeverityBps)
	require.Equal(t, sdkmath.NewInt(250), repeat.StakeSlashNaet)
	require.Equal(t, uint64(24), repeat.TemporaryJailEpochs)

	chronic, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:			"downtime-chronic",
		ValidatorID:			"val-a",
		SeverityLevel:			SlashSeverityRepeatedLivenessFault,
		StakeExposureNaet:		sdkmath.NewInt(100_000),
		SelfStakeNaet:			sdkmath.NewInt(100_000),
		TemporaryJailEpochs:		72,
		FutureElectionScorePenaltyBps:	500,
		EvidenceHeight:			30,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(100), chronic.SeverityBps)
	require.Equal(t, sdkmath.NewInt(1_000), chronic.StakeSlashNaet)
	require.Equal(t, uint64(72), chronic.TemporaryJailEpochs)
	require.Equal(t, uint32(500), chronic.FutureElectionScorePenaltyBps)
}

func TestComputeSlashingPenaltyCapsAtStakeExposureAndAppliesTombstoneIdentityInvalidation(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:			"penalty-critical",
		ValidatorID:			"val-a",
		SeverityLevel:			SlashSeverityCritical,
		SeverityBps:			9_000,
		StakeExposureNaet:		sdkmath.NewInt(1_000),
		RoleWeightBps:			10_000,
		RepeatOffenseMultiplierBps:	10_000,
		TaskImpactBps:			10_000,
		SafetyImpactBps:		10_000,
		LivenessImpactBps:		10_000,
		SelfStakeNaet:			sdkmath.NewInt(1_000),
		PermanentTombstone:		true,
		IdentityInvalidation:		true,
		FutureElectionScorePenaltyBps:	10_000,
		EvidenceHeight:			60,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(900), penalty.StakeSlashNaet)
	require.True(t, penalty.PermanentTombstone)
	require.True(t, penalty.IdentityInvalidation)

	candidate := candidate("val-a", 1_000_000_000, 0)
	candidate.ReliabilityIndexBps = BasisPoints
	candidate.Roles = []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleBlockProducer}
	applied, err := ApplySlashingPenaltyToCandidate(candidate, penalty)
	require.NoError(t, err)
	require.True(t, applied.Tombstoned)
	require.Equal(t, uint32(0), applied.ReliabilityIndexBps)
	require.Equal(t, sdkmath.NewInt(999_999_100), applied.SelfStakeNaet)
}

func TestRouteSlashingPenaltySplitsBurnReporterTreasuryAndCompensation(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-routing",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityDoubleSign,
		StakeExposureNaet:	sdkmath.NewInt(10_000),
		SelfStakeNaet:		sdkmath.NewInt(10_000),
		RewardConfiscationNaet:	sdkmath.NewInt(1_000),
		EvidenceHeight:		70,
	})
	require.NoError(t, err)
	routing, err := RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:		penalty,
		ReporterID:		"reporter-a",
		AffectedPoolIDOptional:	"pool-a",
		BurnBps:		3_000,
		ReporterRewardBps:	2_000,
		ProtocolTreasuryBps:	4_000,
		CompensationBps:	1_000,
		ReporterRewardCapBps:	1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_500), routing.TotalPenaltyNaet)
	require.Equal(t, sdkmath.NewInt(450), routing.BurnNaet)
	require.Equal(t, sdkmath.NewInt(150), routing.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(600), routing.ProtocolTreasuryNaet)
	require.Equal(t, sdkmath.NewInt(150), routing.CompensationNaet)
	require.Equal(t, sdkmath.NewInt(150), routing.ResidualNaet)
	require.Equal(t, routing.TotalPenaltyNaet, routing.BurnNaet.Add(routing.ReporterRewardNaet).Add(routing.ProtocolTreasuryNaet).Add(routing.CompensationNaet).Add(routing.ResidualNaet))
	require.Len(t, routing.RoutingHash, PosHashHexLength)
}

func TestDoubleSignSlashingDefaultsToFivePercentJailAndTombstone(t *testing.T) {
	candidate := candidate("val-a", 10_000, 0)
	result, err := ExecuteSlashing(SlashingExecutionInput{
		EvidenceID:	"double-sign-evidence-1",
		ExecutedHeight:	100,
		CurrentEpoch:	10,
		Candidate:	candidate,
		PenaltyInput: SlashingPenaltyInput{
			PenaltyID:		"double-sign-penalty-1",
			SeverityLevel:		SlashSeverityDoubleSign,
			StakeExposureNaet:	sdkmath.NewInt(10_000),
			SelfStakeNaet:		candidate.SelfStakeNaet,
			TemporaryJailEpochs:	1,
			PermanentTombstone:	true,
			EvidenceHeight:		90,
		},
		RoutingInput: SlashingPenaltyRoutingInput{
			BurnBps: BasisPoints,
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(500), result.SeverityMatrix[SlashSeverityDoubleSign])
	require.Equal(t, SlashSeverityDoubleSign, result.Record.Severity)
	require.Equal(t, sdkmath.NewInt(500), result.Record.SlashAmount)
	require.Equal(t, uint64(11), result.Record.JailUntilEpochOptional)
	require.True(t, result.Record.Tombstone)
	require.True(t, result.UpdatedCandidate.Jailed)
	require.True(t, result.UpdatedCandidate.Tombstoned)
	require.Equal(t, sdkmath.NewInt(9_500), result.UpdatedCandidate.SelfStakeNaet)
}

func TestSlashingRecordMatchesDesignFieldsAndExecutesInvariants(t *testing.T) {
	require.Equal(t, []string{
		"penalty_id",
		"validator_address",
		"evidence_id",
		"severity",
		"stake_exposure",
		"role_weight",
		"slash_amount",
		"delegator_slash_amount",
		"reward_confiscation",
		"jail_until_epoch_optional",
		"tombstone",
		"routing",
		"executed_height",
	}, SlashingRecordFieldNames())
	matrix := SeverityMatrix()
	require.Equal(t, uint32(500), matrix[SlashSeverityDoubleSign])
	require.Equal(t, uint32(7_500), matrix[SlashSeverityEvidenceFraud])

	candidate := candidate("val-a", 10_000, 5_000)
	candidate.ReliabilityIndexBps = BasisPoints
	candidate.Roles = []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleBlockProducer}
	result, err := ExecuteSlashing(SlashingExecutionInput{
		EvidenceID:		"evidence-1",
		ReporterID:		"reporter-a",
		AffectedPoolIDOptional:	"pool-a",
		ExecutedHeight:		100,
		CurrentEpoch:		10,
		Candidate:		candidate,
		PenaltyInput: SlashingPenaltyInput{
			PenaltyID:			"penalty-record-1",
			SeverityLevel:			SlashSeverityInvalidTaskExecution,
			StakeExposureNaet:		sdkmath.NewInt(15_000),
			RoleWeightBps:			10_000,
			SelfStakeNaet:			candidate.SelfStakeNaet,
			Nominations:			candidate.Nominations,
			RewardConfiscationNaet:		sdkmath.NewInt(100),
			TemporaryJailEpochs:		2,
			RoleSuspensions:		[]ValidatorRole{ValidatorRoleVerifier},
			FutureElectionScorePenaltyBps:	1_000,
			EvidenceHeight:			90,
		},
		RoutingInput: SlashingPenaltyRoutingInput{
			BurnBps:		3_000,
			ReporterRewardBps:	1_000,
			ProtocolTreasuryBps:	4_000,
			CompensationBps:	2_000,
			ReporterRewardCapBps:	1_000,
		},
	})
	require.NoError(t, err)
	require.True(t, result.NonNegative)
	require.True(t, result.ExactRouting)
	require.True(t, result.DelegatorExact)
	require.Equal(t, "penalty-record-1", result.Record.PenaltyID)
	require.Equal(t, "val-a", result.Record.ValidatorAddress)
	require.Equal(t, "evidence-1", result.Record.EvidenceID)
	require.Equal(t, SlashSeverityInvalidTaskExecution, result.Record.Severity)
	require.Equal(t, sdkmath.NewInt(15_000), result.Record.StakeExposure)
	require.Equal(t, uint32(10_000), result.Record.RoleWeight)
	require.Equal(t, sdkmath.NewInt(1_125), result.Record.SlashAmount)
	require.Equal(t, sdkmath.NewInt(375), result.Record.DelegatorSlashAmount)
	require.Equal(t, sdkmath.NewInt(100), result.Record.RewardConfiscation)
	require.Equal(t, uint64(12), result.Record.JailUntilEpochOptional)
	require.False(t, result.Record.Tombstone)
	require.Equal(t, int64(100), result.Record.ExecutedHeight)
	require.Len(t, result.Record.RecordHash, PosHashHexLength)
	require.NoError(t, result.Record.Validate())
	require.Equal(t, result.Record.Routing.TotalPenaltyNaet, result.Record.Routing.BurnNaet.Add(result.Record.Routing.ReporterRewardNaet).Add(result.Record.Routing.ProtocolTreasuryNaet).Add(result.Record.Routing.CompensationNaet).Add(result.Record.Routing.ResidualNaet))
	require.False(t, result.UpdatedCandidate.SelfStakeNaet.IsNegative())
	require.False(t, result.UpdatedCandidate.DelegatedStakeNaet.IsNegative())
	require.Equal(t, []ValidatorRole{ValidatorRoleBlockProducer}, result.UpdatedCandidate.Roles)
}

func TestSlashingRecordRejectsRoutingMismatchAndNegativeExecution(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-record-bad",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityDoubleSign,
		StakeExposureNaet:	sdkmath.NewInt(1_000),
		SelfStakeNaet:		sdkmath.NewInt(1_000),
		RewardConfiscationNaet:	sdkmath.NewInt(100),
		EvidenceHeight:		90,
	})
	require.NoError(t, err)
	routing, err := RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:		penalty,
		ReporterID:		"reporter-a",
		AffectedPoolIDOptional:	"pool-a",
	})
	require.NoError(t, err)
	record := NewSlashingRecord("evidence-bad", penalty, routing, 10, 100)
	require.NoError(t, record.Validate())
	record.SlashAmount = record.SlashAmount.AddRaw(1)
	require.ErrorContains(t, record.Validate(), "routing total")

	_, err = ExecuteSlashing(SlashingExecutionInput{
		EvidenceID:	"evidence-bad",
		ExecutedHeight:	-1,
		CurrentEpoch:	10,
		Candidate:	candidate("val-a", 1_000, 0),
		PenaltyInput: SlashingPenaltyInput{
			PenaltyID:		"penalty-record-bad-2",
			SeverityLevel:		SlashSeverityDoubleSign,
			StakeExposureNaet:	sdkmath.NewInt(1_000),
			SelfStakeNaet:		sdkmath.NewInt(1_000),
		},
	})
	require.ErrorContains(t, err, "executed height")
}

func TestRouteSlashingPenaltyRejectsMissingPoolsAndInvalidBps(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-routing-bad",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityMinorLivenessFault,
		StakeExposureNaet:	sdkmath.NewInt(10_000),
		SelfStakeNaet:		sdkmath.NewInt(10_000),
		EvidenceHeight:		70,
	})
	require.NoError(t, err)
	_, err = RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:		penalty,
		ReporterID:		"reporter-a",
		CompensationBps:	1_000,
	})
	require.ErrorContains(t, err, "affected pool")

	_, err = RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:		penalty,
		ReporterID:		"reporter-a",
		AffectedPoolIDOptional:	"pool-a",
		BurnBps:		9_000,
		ReporterRewardBps:	2_000,
	})
	require.ErrorContains(t, err, "routing bps")
}

func TestApplySlashingPenaltyJailsSuspendsRolesAndReducesElectionReliability(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:			"penalty-roles",
		ValidatorID:			"val-a",
		SeverityLevel:			SlashSeverityMedium,
		StakeExposureNaet:		sdkmath.NewInt(1_000),
		RoleWeightBps:			10_000,
		SelfStakeNaet:			sdkmath.NewInt(500),
		Nominations:			[]Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(500)}},
		TemporaryJailEpochs:		3,
		RoleSuspensions:		[]ValidatorRole{ValidatorRoleVerifier},
		FutureElectionScorePenaltyBps:	2_000,
		EvidenceHeight:			61,
	})
	require.NoError(t, err)
	candidate := candidate("val-a", 2_000_000_000, 2_000)
	candidate.ReliabilityIndexBps = 9_000
	candidate.Roles = []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleCollator}

	applied, err := ApplySlashingPenaltyToCandidate(candidate, penalty)
	require.NoError(t, err)
	require.True(t, applied.Jailed)
	require.False(t, applied.Tombstoned)
	require.Equal(t, []ValidatorRole{ValidatorRoleCollator}, applied.Roles)
	require.Equal(t, uint32(7_200), applied.ReliabilityIndexBps)
	require.Equal(t, sdkmath.NewInt(1_999_999_950), applied.SelfStakeNaet)
	require.Equal(t, sdkmath.NewInt(1_950), applied.DelegatedStakeNaet)
}

func TestSlashingPenaltyRejectsUnsafeInputsAndHashTampering(t *testing.T) {
	_, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-bad",
		ValidatorID:		"val-a",
		SeverityLevel:		"unknown",
		StakeExposureNaet:	sdkmath.NewInt(1_000),
		SelfStakeNaet:		sdkmath.NewInt(1_000),
	})
	require.ErrorContains(t, err, "unsupported slash severity")

	_, err = ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-bad",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityLow,
		StakeExposureNaet:	sdkmath.NewInt(1_000),
		RoleWeightBps:		BasisPoints + 1,
		SelfStakeNaet:		sdkmath.NewInt(1_000),
	})
	require.ErrorContains(t, err, "role weight")

	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-good",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityLow,
		StakeExposureNaet:	sdkmath.NewInt(1_000),
		SelfStakeNaet:		sdkmath.NewInt(1_000),
	})
	require.NoError(t, err)
	penalty.StakeSlashNaet = penalty.StakeSlashNaet.AddRaw(1)
	require.ErrorContains(t, penalty.Validate(), "components")
}
