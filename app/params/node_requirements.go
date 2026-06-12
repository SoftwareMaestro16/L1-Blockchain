package params

import (
	"fmt"
	"strings"
)

const (
	AetraValidatorCPUCoreMin	= 4
	AetraValidatorCPUCoreMax	= 8
	AetraValidatorRAMMinGB		= 16
	AetraValidatorRAMMaxGB		= 32
	AetraValidatorMinNetworkMbps	= 100

	AetraValidatorStorageClass		= "NVMe SSD"
	AetraValidatorRecommendedOS		= "Linux"
	AetraValidatorDevelopmentOSSupport	= "Windows local tooling supported for development"

	AetraPruningProfileDefault	= "default"
	AetraPruningProfileArchive	= "nothing"
	AetraPruningProfileAggressive	= "everything"
	AetraPruningProfileCustom	= "custom"
)

type NodeHardwareProfile struct {
	CPUCores			int
	RAMGB				int
	Storage				string
	NetworkMbps			int
	LowPacketLoss			bool
	OS				string
	MainnetLoadTestingComplete	bool
}

type NodeStateManagementReadiness struct {
	StateSyncSupported			bool
	SnapshotsSupported			bool
	PruningProfiles				[]string
	ArchiveNodeProfile			bool
	ExportImportReliable			bool
	RestartSafety				bool
	DeterministicAppHashAcrossRestarts	bool
	ValidatorSetupDocumented		bool
	SentryArchitectureDocumented		bool
}

func DefaultPublicTestnetHardwareProfile() NodeHardwareProfile {
	return NodeHardwareProfile{
		CPUCores:	AetraValidatorCPUCoreMin,
		RAMGB:		AetraValidatorRAMMinGB,
		Storage:	AetraValidatorStorageClass,
		NetworkMbps:	AetraValidatorMinNetworkMbps,
		LowPacketLoss:	true,
		OS:		AetraValidatorRecommendedOS,
	}
}

func DefaultLocalDevelopmentHardwareProfile() NodeHardwareProfile {
	profile := DefaultPublicTestnetHardwareProfile()
	profile.OS = AetraValidatorDevelopmentOSSupport
	return profile
}

func DefaultNodeStateManagementReadiness() NodeStateManagementReadiness {
	return NodeStateManagementReadiness{
		StateSyncSupported:			true,
		SnapshotsSupported:			true,
		PruningProfiles:			DefaultPruningProfiles(),
		ArchiveNodeProfile:			true,
		ExportImportReliable:			true,
		RestartSafety:				true,
		DeterministicAppHashAcrossRestarts:	true,
		ValidatorSetupDocumented:		true,
		SentryArchitectureDocumented:		true,
	}
}

func DefaultPruningProfiles() []string {
	return []string{
		AetraPruningProfileDefault,
		AetraPruningProfileArchive,
		AetraPruningProfileAggressive,
		AetraPruningProfileCustom,
	}
}

func ValidatePublicTestnetHardwareProfile(profile NodeHardwareProfile) error {
	if err := validateMediumValidatorHardware(profile); err != nil {
		return err
	}
	if !strings.EqualFold(profile.OS, AetraValidatorRecommendedOS) {
		return fmt.Errorf("public testnet validator OS must be %s recommended; Windows is local tooling only", AetraValidatorRecommendedOS)
	}
	return nil
}

func ValidateLocalDevelopmentHardwareProfile(profile NodeHardwareProfile) error {
	if err := validateMediumValidatorHardware(profile); err != nil {
		return err
	}
	if !strings.EqualFold(profile.OS, AetraValidatorRecommendedOS) && profile.OS != AetraValidatorDevelopmentOSSupport {
		return fmt.Errorf("local development OS must be %s or %s", AetraValidatorRecommendedOS, AetraValidatorDevelopmentOSSupport)
	}
	return nil
}

func ValidateMainnetHardwareProfile(profile NodeHardwareProfile) error {
	if err := ValidatePublicTestnetHardwareProfile(profile); err != nil {
		return err
	}
	if !profile.MainnetLoadTestingComplete {
		return fmt.Errorf("mainnet hardware requirements must be finalized after load testing")
	}
	return nil
}

func ValidateNodeStateManagementReadiness(readiness NodeStateManagementReadiness) error {
	if !readiness.StateSyncSupported {
		return fmt.Errorf("state sync support is required")
	}
	if !readiness.SnapshotsSupported {
		return fmt.Errorf("snapshots are required")
	}
	if err := validateRequiredPruningProfiles(readiness.PruningProfiles); err != nil {
		return err
	}
	if !readiness.ArchiveNodeProfile {
		return fmt.Errorf("archive node profile is required")
	}
	if !readiness.ExportImportReliable {
		return fmt.Errorf("export/import reliability is required")
	}
	if !readiness.RestartSafety {
		return fmt.Errorf("restart safety is required")
	}
	if !readiness.DeterministicAppHashAcrossRestarts {
		return fmt.Errorf("deterministic app hash across restarts is required")
	}
	if !readiness.ValidatorSetupDocumented {
		return fmt.Errorf("documented validator setup is required")
	}
	if !readiness.SentryArchitectureDocumented {
		return fmt.Errorf("documented sentry architecture is required")
	}
	return nil
}

func validateMediumValidatorHardware(profile NodeHardwareProfile) error {
	if profile.CPUCores < AetraValidatorCPUCoreMin || profile.CPUCores > AetraValidatorCPUCoreMax {
		return fmt.Errorf("validator CPU must stay within %d-%d modern cores", AetraValidatorCPUCoreMin, AetraValidatorCPUCoreMax)
	}
	if profile.RAMGB < AetraValidatorRAMMinGB || profile.RAMGB > AetraValidatorRAMMaxGB {
		return fmt.Errorf("validator RAM must stay within %d-%d GB", AetraValidatorRAMMinGB, AetraValidatorRAMMaxGB)
	}
	if !strings.EqualFold(profile.Storage, AetraValidatorStorageClass) {
		return fmt.Errorf("validator storage must be %s", AetraValidatorStorageClass)
	}
	if profile.NetworkMbps < AetraValidatorMinNetworkMbps {
		return fmt.Errorf("validator network must be stable %d Mbps+", AetraValidatorMinNetworkMbps)
	}
	if !profile.LowPacketLoss {
		return fmt.Errorf("validator network must have low packet loss")
	}
	return nil
}

func validateRequiredPruningProfiles(profiles []string) error {
	required := map[string]bool{
		AetraPruningProfileDefault:	false,
		AetraPruningProfileArchive:	false,
		AetraPruningProfileAggressive:	false,
	}
	for _, profile := range profiles {
		if _, ok := required[profile]; ok {
			required[profile] = true
		}
	}
	for profile, present := range required {
		if !present {
			return fmt.Errorf("pruning profiles must include %q", profile)
		}
	}
	return nil
}
