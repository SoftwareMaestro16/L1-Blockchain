package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestComputeSlashingPenaltyScalesStakeExposureBySeverityRoleAndImpact(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:                     "penalty-1",
		ValidatorID:                   "val-a",
		SeverityLevel:                 SlashSeverityHigh,
		StakeExposureNaet:             sdkmath.NewInt(10_000),
		RoleWeightBps:                 5_000,
		RepeatOffenseMultiplierBps:    10_000,
		TaskImpactBps:                 8_000,
		SafetyImpactBps:               10_000,
		LivenessImpactBps:             10_000,
		SelfStakeNaet:                 sdkmath.NewInt(4_000),
		Nominations:                   []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(6_000)}},
		RewardConfiscationNaet:        sdkmath.NewInt(250),
		TemporaryJailEpochs:           2,
		RoleSuspensions:               []ValidatorRole{ValidatorRoleVerifier},
		FutureElectionScorePenaltyBps: 1_000,
		EvidenceHeight:                50,
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
		SlashSeverityEquivocation,
		SlashSeverityDoubleSign,
		SlashSeverityEvidenceFraud,
	}, SlashSeverityClasses())

	expected := map[string]uint32{
		SlashSeverityMinorLivenessFault:     100,
		SlashSeverityMajorLivenessFault:     500,
		SlashSeverityRepeatedLivenessFault:  1_000,
		SlashSeverityInvalidTaskExecution:   750,
		SlashSeverityInvalidStateTransition: 1_500,
		SlashSeverityEquivocation:           2_000,
		SlashSeverityDoubleSign:             5_000,
		SlashSeverityEvidenceFraud:          7_500,
	}
	for severity, bps := range expected {
		actual, err := DefaultSeverityBps(severity)
		require.NoError(t, err)
		require.Equal(t, bps, actual)
	}
}

func TestComputeSlashingPenaltyCapsAtStakeExposureAndAppliesTombstoneIdentityInvalidation(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:                     "penalty-critical",
		ValidatorID:                   "val-a",
		SeverityLevel:                 SlashSeverityCritical,
		SeverityBps:                   9_000,
		StakeExposureNaet:             sdkmath.NewInt(1_000),
		RoleWeightBps:                 10_000,
		RepeatOffenseMultiplierBps:    10_000,
		TaskImpactBps:                 10_000,
		SafetyImpactBps:               10_000,
		LivenessImpactBps:             10_000,
		SelfStakeNaet:                 sdkmath.NewInt(1_000),
		PermanentTombstone:            true,
		IdentityInvalidation:          true,
		FutureElectionScorePenaltyBps: 10_000,
		EvidenceHeight:                60,
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
		PenaltyID:              "penalty-routing",
		ValidatorID:            "val-a",
		SeverityLevel:          SlashSeverityDoubleSign,
		StakeExposureNaet:      sdkmath.NewInt(10_000),
		SelfStakeNaet:          sdkmath.NewInt(10_000),
		RewardConfiscationNaet: sdkmath.NewInt(1_000),
		EvidenceHeight:         70,
	})
	require.NoError(t, err)
	routing, err := RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:                penalty,
		ReporterID:             "reporter-a",
		AffectedPoolIDOptional: "pool-a",
		BurnBps:                3_000,
		ReporterRewardBps:      2_000,
		ProtocolTreasuryBps:    4_000,
		CompensationBps:        1_000,
		ReporterRewardCapBps:   1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(6_000), routing.TotalPenaltyNaet)
	require.Equal(t, sdkmath.NewInt(1_800), routing.BurnNaet)
	require.Equal(t, sdkmath.NewInt(600), routing.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(2_400), routing.ProtocolTreasuryNaet)
	require.Equal(t, sdkmath.NewInt(600), routing.CompensationNaet)
	require.Equal(t, sdkmath.NewInt(600), routing.ResidualNaet)
	require.Equal(t, routing.TotalPenaltyNaet, routing.BurnNaet.Add(routing.ReporterRewardNaet).Add(routing.ProtocolTreasuryNaet).Add(routing.CompensationNaet).Add(routing.ResidualNaet))
	require.Len(t, routing.RoutingHash, PosHashHexLength)
}

func TestRouteSlashingPenaltyRejectsMissingPoolsAndInvalidBps(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:         "penalty-routing-bad",
		ValidatorID:       "val-a",
		SeverityLevel:     SlashSeverityMinorLivenessFault,
		StakeExposureNaet: sdkmath.NewInt(10_000),
		SelfStakeNaet:     sdkmath.NewInt(10_000),
		EvidenceHeight:    70,
	})
	require.NoError(t, err)
	_, err = RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:         penalty,
		ReporterID:      "reporter-a",
		CompensationBps: 1_000,
	})
	require.ErrorContains(t, err, "affected pool")

	_, err = RouteSlashingPenalty(SlashingPenaltyRoutingInput{
		Penalty:                penalty,
		ReporterID:             "reporter-a",
		AffectedPoolIDOptional: "pool-a",
		BurnBps:                9_000,
		ReporterRewardBps:      2_000,
	})
	require.ErrorContains(t, err, "routing bps")
}

func TestApplySlashingPenaltyJailsSuspendsRolesAndReducesElectionReliability(t *testing.T) {
	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:                     "penalty-roles",
		ValidatorID:                   "val-a",
		SeverityLevel:                 SlashSeverityMedium,
		StakeExposureNaet:             sdkmath.NewInt(1_000),
		RoleWeightBps:                 10_000,
		SelfStakeNaet:                 sdkmath.NewInt(500),
		Nominations:                   []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(500)}},
		TemporaryJailEpochs:           3,
		RoleSuspensions:               []ValidatorRole{ValidatorRoleVerifier},
		FutureElectionScorePenaltyBps: 2_000,
		EvidenceHeight:                61,
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
		PenaltyID:         "penalty-bad",
		ValidatorID:       "val-a",
		SeverityLevel:     "unknown",
		StakeExposureNaet: sdkmath.NewInt(1_000),
		SelfStakeNaet:     sdkmath.NewInt(1_000),
	})
	require.ErrorContains(t, err, "unsupported slash severity")

	_, err = ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:         "penalty-bad",
		ValidatorID:       "val-a",
		SeverityLevel:     SlashSeverityLow,
		StakeExposureNaet: sdkmath.NewInt(1_000),
		RoleWeightBps:     BasisPoints + 1,
		SelfStakeNaet:     sdkmath.NewInt(1_000),
	})
	require.ErrorContains(t, err, "role weight")

	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:         "penalty-good",
		ValidatorID:       "val-a",
		SeverityLevel:     SlashSeverityLow,
		StakeExposureNaet: sdkmath.NewInt(1_000),
		SelfStakeNaet:     sdkmath.NewInt(1_000),
	})
	require.NoError(t, err)
	penalty.StakeSlashNaet = penalty.StakeSlashNaet.AddRaw(1)
	require.ErrorContains(t, penalty.Validate(), "components")
}
