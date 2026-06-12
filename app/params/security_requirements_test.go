package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultSecurityRequirementsPass(t *testing.T) {
	consensus := DefaultConsensusSafetyRequirements()
	economic := DefaultEconomicSafetyRequirements()
	report := BuildSecurityRequirementsReport(consensus, economic)

	require.True(t, report.Passed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.ConsensusRequired, report.ConsensusPassed)
	require.Equal(t, report.EconomicRequired, report.EconomicPassed)
	require.Equal(t, report.PermissionRequired, report.PermissionPassed)
	require.NoError(t, ValidateSecurityRequirements(consensus, economic))
}

func TestConsensusSafetyRejectsNonDeterministicRisks(t *testing.T) {
	economic := DefaultEconomicSafetyRequirements()

	cases := []struct {
		name	string
		mutate	func(*ConsensusSafetyRequirements)
		failure	string
	}{
		{"state transitions", func(r *ConsensusSafetyRequirements) { r.DeterministicStateTransitions = false }, SecurityRequirementDeterministicStateTransitions},
		{"external calls", func(r *ConsensusSafetyRequirements) { r.NoNonDeterministicExternalCalls = false }, SecurityRequirementNoExternalConsensusCalls},
		{"wall clock", func(r *ConsensusSafetyRequirements) { r.NoWallClockDependencyExceptBlockTime = false }, SecurityRequirementNoWallClockStateTransitions},
		{"floating point", func(r *ConsensusSafetyRequirements) { r.NoFloatingPointAccounting = false }, SecurityRequirementNoFloatingPointAccounting},
		{"map iteration", func(r *ConsensusSafetyRequirements) { r.NoUnorderedMapIterationAffectingState = false }, SecurityRequirementNoUnorderedMapStateEffects},
		{"serialization", func(r *ConsensusSafetyRequirements) { r.DeterministicSerialization = false }, SecurityRequirementDeterministicSerialization},
		{"export import", func(r *ConsensusSafetyRequirements) { r.ExportImportEqualityTests = false }, SecurityRequirementExportImportEqualityTests},
		{"app hash", func(r *ConsensusSafetyRequirements) { r.AppHashStabilityTests = false }, SecurityRequirementAppHashStabilityTests},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			consensus := DefaultConsensusSafetyRequirements()
			tc.mutate(&consensus)
			report := BuildSecurityRequirementsReport(consensus, economic)
			require.False(t, report.Passed)
			require.Contains(t, report.Failed, tc.failure)
			require.Error(t, ValidateSecurityRequirements(consensus, economic))
		})
	}
}

func TestEconomicSafetyRejectsInvariantGaps(t *testing.T) {
	consensus := DefaultConsensusSafetyRequirements()

	cases := []struct {
		name	string
		mutate	func(*EconomicSafetyRequirements)
		failure	string
	}{
		{"unbounded mint", func(r *EconomicSafetyRequirements) { r.NoUnboundedMint = false }, SecurityRequirementNoUnboundedMint},
		{"module mint burn", func(r *EconomicSafetyRequirements) { r.NoUnauthorizedModuleAccountMintBurn = false }, SecurityRequirementNoUnauthorizedModuleMintBurn},
		{"supply", func(r *EconomicSafetyRequirements) { r.SupplyInvariants = false }, SecurityRequirementSupplyInvariants},
		{"fee split", func(r *EconomicSafetyRequirements) { r.FeeSplitInvariants = false }, SecurityRequirementFeeSplitInvariants},
		{"delegation shares", func(r *EconomicSafetyRequirements) { r.DelegationShareInvariants = false }, SecurityRequirementDelegationShareInvariants},
		{"rewards", func(r *EconomicSafetyRequirements) { r.RewardDistributionInvariants = false }, SecurityRequirementRewardDistributionInvariants},
		{"slash underflow", func(r *EconomicSafetyRequirements) { r.SlashingCannotUnderflowStake = false }, SecurityRequirementSlashingNoStakeUnderflow},
		{"jailed rewards", func(r *EconomicSafetyRequirements) { r.JailedValidatorsCannotReceiveRewards = false }, SecurityRequirementJailedValidatorRewardExclusion},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			economic := DefaultEconomicSafetyRequirements()
			tc.mutate(&economic)
			report := BuildSecurityRequirementsReport(consensus, economic)
			require.False(t, report.Passed)
			require.Contains(t, report.Failed, tc.failure)
			require.Error(t, ValidateSecurityRequirements(consensus, economic))
		})
	}
}

func TestPermissionSafetyRejectsAuthorityAndWiringGaps(t *testing.T) {
	consensus := DefaultConsensusSafetyRequirements()
	economic := DefaultEconomicSafetyRequirements()

	cases := []struct {
		name	string
		mutate	func(*PermissionSafetyRequirements)
		failure	string
	}{
		{"module permissions startup", func(r *PermissionSafetyRequirements) { r.ModuleAccountPermissionsValidatedAtStartup = false }, SecurityRequirementModulePermissionsStartup},
		{"reserved user tx signer", func(r *PermissionSafetyRequirements) { r.ReservedAddressesCannotSignUserTxs = false }, SecurityRequirementReservedCannotSignUserTxs},
		{"blocked receives", func(r *PermissionSafetyRequirements) { r.BlockedAddressesCannotReceiveUserFunds = false }, SecurityRequirementBlockedCannotReceiveFunds},
		{"governance authority", func(r *PermissionSafetyRequirements) { r.GovernanceAuthorityChecked = false }, SecurityRequirementGovernanceAuthorityChecked},
		{"params authority", func(r *PermissionSafetyRequirements) { r.ParamsAuthorityChecked = false }, SecurityRequirementParamsAuthorityChecked},
		{"keeper wiring", func(r *PermissionSafetyRequirements) { r.KeeperWiringTests = false }, SecurityRequirementKeeperWiringTests},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			permission := DefaultPermissionSafetyRequirements()
			tc.mutate(&permission)
			report := BuildSecurityRequirementsReport(consensus, economic, permission)
			require.False(t, report.Passed)
			require.Contains(t, report.Failed, tc.failure)
			require.Error(t, ValidateSecurityRequirements(consensus, economic, permission))
		})
	}
}

func TestPermissionSafetyRequiresExplicitAllowlistAndReservedCatalog(t *testing.T) {
	consensus := DefaultConsensusSafetyRequirements()
	economic := DefaultEconomicSafetyRequirements()

	permission := DefaultPermissionSafetyRequirements()
	permission.ExplicitReceiveAllowlistRequired = false
	require.Contains(t, BuildSecurityRequirementsReport(consensus, economic, permission).Failed, SecurityRequirementBlockedCannotReceiveFunds)

	permission = DefaultPermissionSafetyRequirements()
	permission.UserTxSignerValidationUsesReservedCatalog = false
	require.Contains(t, BuildSecurityRequirementsReport(consensus, economic, permission).Failed, SecurityRequirementReservedCannotSignUserTxs)

	permission = DefaultPermissionSafetyRequirements()
	permission.ReservedAddressCatalogValidated = false
	require.Contains(t, BuildSecurityRequirementsReport(consensus, economic, permission).Failed, SecurityRequirementKeeperWiringTests)
}

func TestSlashingCannotUnderflowStake(t *testing.T) {
	require.NoError(t, ValidateSlashingDoesNotUnderflowStake(1_000, 250))
	require.NoError(t, ValidateSlashingDoesNotUnderflowStake(1_000, 1_000))
	require.ErrorContains(t, ValidateSlashingDoesNotUnderflowStake(1_000, 1_001), "underflow")
	require.ErrorContains(t, ValidateSlashingDoesNotUnderflowStake(-1, 0), "non-negative")
	require.ErrorContains(t, ValidateSlashingDoesNotUnderflowStake(1, -1), "non-negative")
}

func TestJailedValidatorsCannotReceiveActiveRewards(t *testing.T) {
	eligible := ValidatorRewardEligibility{
		ValidatorAddress:	"aetravaloper1example",
		ActiveSet:		true,
		RewardNaet:		100,
	}
	require.NoError(t, ValidateActiveValidatorRewardEligibility(eligible))

	eligible.Jailed = true
	require.ErrorContains(t, ValidateActiveValidatorRewardEligibility(eligible), "jailed validators")

	eligible.Jailed = false
	eligible.Tombstoned = true
	require.ErrorContains(t, ValidateActiveValidatorRewardEligibility(eligible), "tombstoned validators")

	eligible.Tombstoned = false
	eligible.ActiveSet = false
	require.ErrorContains(t, ValidateActiveValidatorRewardEligibility(eligible), "inactive validators")

	eligible.RewardNaet = 0
	require.NoError(t, ValidateActiveValidatorRewardEligibility(eligible))

	eligible.RewardNaet = -1
	require.ErrorContains(t, ValidateActiveValidatorRewardEligibility(eligible), "negative")
}
