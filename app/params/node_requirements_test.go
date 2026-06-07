package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultPublicTestnetHardwareProfileMatchesMediumTarget(t *testing.T) {
	profile := DefaultPublicTestnetHardwareProfile()

	require.Equal(t, 4, profile.CPUCores)
	require.Equal(t, 16, profile.RAMGB)
	require.Equal(t, AetraValidatorStorageClass, profile.Storage)
	require.Equal(t, 100, profile.NetworkMbps)
	require.True(t, profile.LowPacketLoss)
	require.Equal(t, AetraValidatorRecommendedOS, profile.OS)
	require.NoError(t, ValidatePublicTestnetHardwareProfile(profile))
}

func TestPublicTestnetHardwareRejectsExtremeOrWeakProfiles(t *testing.T) {
	profile := DefaultPublicTestnetHardwareProfile()
	profile.CPUCores = 2
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "4-8 modern cores")

	profile = DefaultPublicTestnetHardwareProfile()
	profile.CPUCores = 32
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "4-8 modern cores")

	profile = DefaultPublicTestnetHardwareProfile()
	profile.RAMGB = 8
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "16-32 GB")

	profile = DefaultPublicTestnetHardwareProfile()
	profile.Storage = "HDD"
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "NVMe SSD")

	profile = DefaultPublicTestnetHardwareProfile()
	profile.NetworkMbps = 50
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "100 Mbps+")

	profile = DefaultPublicTestnetHardwareProfile()
	profile.LowPacketLoss = false
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "low packet loss")
}

func TestWindowsIsLocalDevelopmentOnly(t *testing.T) {
	profile := DefaultLocalDevelopmentHardwareProfile()

	require.NoError(t, ValidateLocalDevelopmentHardwareProfile(profile))
	require.ErrorContains(t, ValidatePublicTestnetHardwareProfile(profile), "Windows is local tooling only")
}

func TestMainnetHardwareRequiresCompletedLoadTesting(t *testing.T) {
	profile := DefaultPublicTestnetHardwareProfile()
	require.ErrorContains(t, ValidateMainnetHardwareProfile(profile), "after load testing")

	profile.MainnetLoadTestingComplete = true
	require.NoError(t, ValidateMainnetHardwareProfile(profile))
}

func TestDefaultStateManagementReadinessCoversRequiredNodeFeatures(t *testing.T) {
	readiness := DefaultNodeStateManagementReadiness()

	require.True(t, readiness.StateSyncSupported)
	require.True(t, readiness.SnapshotsSupported)
	require.Contains(t, readiness.PruningProfiles, AetraPruningProfileDefault)
	require.Contains(t, readiness.PruningProfiles, AetraPruningProfileArchive)
	require.Contains(t, readiness.PruningProfiles, AetraPruningProfileAggressive)
	require.True(t, readiness.ArchiveNodeProfile)
	require.True(t, readiness.ExportImportReliable)
	require.True(t, readiness.RestartSafety)
	require.True(t, readiness.DeterministicAppHashAcrossRestarts)
	require.True(t, readiness.ValidatorSetupDocumented)
	require.True(t, readiness.SentryArchitectureDocumented)
	require.NoError(t, ValidateNodeStateManagementReadiness(readiness))
}

func TestStateManagementReadinessRejectsMissingRequiredGates(t *testing.T) {
	readiness := DefaultNodeStateManagementReadiness()
	readiness.StateSyncSupported = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "state sync support")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.SnapshotsSupported = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "snapshots")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.PruningProfiles = []string{AetraPruningProfileDefault}
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "pruning profiles")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.ArchiveNodeProfile = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "archive node profile")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.ExportImportReliable = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "export/import reliability")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.RestartSafety = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "restart safety")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.DeterministicAppHashAcrossRestarts = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "deterministic app hash")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.ValidatorSetupDocumented = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "documented validator setup")

	readiness = DefaultNodeStateManagementReadiness()
	readiness.SentryArchitectureDocumented = false
	require.ErrorContains(t, ValidateNodeStateManagementReadiness(readiness), "documented sentry architecture")
}
