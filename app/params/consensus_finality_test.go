package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsensusFinalityReportAcceptsRequiredTargets(t *testing.T) {
	report := validConsensusFinalityReport()

	require.NoError(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report))
}

func TestConsensusFinalityReportRejectsUnstableHundredValidatorLocalnet(t *testing.T) {
	report := validConsensusFinalityReport()
	report.LocalnetStable = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "100-128 validator localnet")
}

func TestConsensusFinalityReportRejectsOneSecondBlocks(t *testing.T) {
	report := validConsensusFinalityReport()
	report.ObservedBlockTimeMinSeconds = 1
	report.ObservedBlockTimeMaxSeconds = 2

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "1-2 second block targets")
}

func TestConsensusFinalityReportRejectsFinalityOutsideBounds(t *testing.T) {
	report := validConsensusFinalityReport()
	report.NormalFinalitySeconds = 16
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "normal finality")

	report = validConsensusFinalityReport()
	report.StressFinalitySeconds = 91
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "stress finality")

	report = validConsensusFinalityReport()
	report.WorstFinalitySeconds = 121
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "worst finality")
}

func TestConsensusFinalityReportRequiresDegradedLivenessWithHealthyTwoThirds(t *testing.T) {
	report := validConsensusFinalityReport()
	report.HealthyVotingPowerBps = AetraHealthyVotingPowerBps
	report.LivenessPreserved = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "2/3 voting power")
}

func TestConsensusFinalityReportRequiresTestnetReportMeasurements(t *testing.T) {
	report := validConsensusFinalityReport()
	report.IncludedInTestnetReport = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "testnet reports")
}

func TestConsensusFinalityReportValidatesMatureBlockTimeRange(t *testing.T) {
	report := validConsensusFinalityReport()
	report.ValidatorCount = 300
	report.ObservedBlockTimeMinSeconds = 7
	report.ObservedBlockTimeMaxSeconds = 8

	require.NoError(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report))
}

func validConsensusFinalityReport() ConsensusFinalityReport {
	return ConsensusFinalityReport{
		ValidatorCount:			100,
		BlocksObserved:			100,
		LocalnetStable:			true,
		LoadProfileExecuted:		true,
		ObservedBlockTimeMinSeconds:	5,
		ObservedBlockTimeMaxSeconds:	6,
		NormalFinalitySeconds:		10,
		StressFinalitySeconds:		60,
		WorstFinalitySeconds:		90,
		DegradedScenarioExecuted:	true,
		HealthyVotingPowerBps:		AetraHealthyVotingPowerBps,
		LivenessPreserved:		true,
		IncludedInTestnetReport:	true,
	}
}

func TestLocalnetPerformanceProfileMatchesExpectedValues(t *testing.T) {
	p := LocalnetPerformanceProfile()
	require.Equal(t, NetworkLocalnet, p.Name)
	require.Equal(t, 2, p.BlockTimeSeconds)
	require.Equal(t, int64(50_000_000), p.MaxBlockGas)
	require.Equal(t, 500, p.MaxTxPerBlock)
	require.NoError(t, p.Validate())
}

func TestTestnetPerformanceProfileMatchesExpectedValues(t *testing.T) {
	p := TestnetPerformanceProfile()
	require.Equal(t, NetworkTestnet, p.Name)
	require.Equal(t, 5, p.BlockTimeSeconds)
	require.Equal(t, int64(100_000_000), p.MaxBlockGas)
	require.Equal(t, 1_000, p.MaxTxPerBlock)
	require.NoError(t, p.Validate())
}

func TestMainnetPerformanceProfileMatchesExpectedValues(t *testing.T) {
	p := MainnetPerformanceProfile()
	require.Equal(t, NetworkMainnet, p.Name)
	require.Equal(t, 6, p.BlockTimeSeconds)
	require.Equal(t, int64(200_000_000), p.MaxBlockGas)
	require.Equal(t, 2_000, p.MaxTxPerBlock)
	require.NoError(t, p.Validate())
}

func TestDefaultPerformanceProfileSelectsByName(t *testing.T) {
	localnet, err := DefaultPerformanceProfile(NetworkLocalnet)
	require.NoError(t, err)
	require.Equal(t, LocalnetPerformanceProfile(), localnet)

	testnet, err := DefaultPerformanceProfile(NetworkTestnet)
	require.NoError(t, err)
	require.Equal(t, TestnetPerformanceProfile(), testnet)

	mainnet, err := DefaultPerformanceProfile(NetworkMainnet)
	require.NoError(t, err)
	require.Equal(t, MainnetPerformanceProfile(), mainnet)

	_, err = DefaultPerformanceProfile("unknown")
	require.ErrorContains(t, err, "unknown network profile")
}

func TestPerformanceProfileRejectsInvalidValues(t *testing.T) {
	p := PerformanceProfile{Name: "custom", BlockTimeSeconds: 0, MaxBlockGas: 100, MaxTxPerBlock: 0}
	require.ErrorContains(t, p.Validate(), "block time")

	p = PerformanceProfile{Name: "custom", BlockTimeSeconds: 5, MaxBlockGas: 100, MaxTxPerBlock: 50}
	require.ErrorContains(t, p.Validate(), "max block gas")

	p = PerformanceProfile{Name: "custom", BlockTimeSeconds: 5, MaxBlockGas: 50_000_000, MaxTxPerBlock: 0}
	require.ErrorContains(t, p.Validate(), "max tx per block")

	p = PerformanceProfile{Name: "", BlockTimeSeconds: 5, MaxBlockGas: 50_000_000, MaxTxPerBlock: 500}
	require.ErrorContains(t, p.Validate(), "name is required")
}
