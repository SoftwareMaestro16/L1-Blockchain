package wasmconfig

const (
	ModuleName = "wasm"
	StoreKey   = ModuleName

	RecommendedWasmdVersion  = "v0.70.2"
	RecommendedWasmVMVersion = "v3.0.6"
	RecommendedSDKMinor      = "v0.54"

	DefaultMaxContractSizeBytes         uint64 = 800 * 1024
	DefaultMaxProposalContractSizeBytes uint64 = 3 * 1024 * 1024
	DefaultSmartQueryGasLimit           uint64 = 3_000_000
	DefaultSimulationGasLimit           uint64 = 20_000_000
	DefaultGasMultiplier                uint64 = 140_000
	DefaultMemoryCacheSizeMiB           uint32 = 100
	DefaultMaxQueryResponseBytes        uint64 = 256 * 1024
	DefaultMaxQueryDepth                uint32 = 8
	DefaultMaxPinnedCodes               uint32 = 0

	maxSmartQueryGasLimit uint64 = 10_000_000
	maxSimulationGasLimit uint64 = 100_000_000
	maxMemoryCacheSizeMiB uint32 = 256
	maxQueryResponseBytes uint64 = 1024 * 1024
	maxQueryDepth         uint32 = 16
	maxPinnedCodes        uint32 = 128
)

type UploadPermission string

const (
	UploadPermissionGovernanceOnly UploadPermission = "governance-only"
	UploadPermissionAllowlist      UploadPermission = "allowlist"
)

type InstantiatePermission string

const (
	InstantiatePermissionCodeOwnerOnly InstantiatePermission = "code-owner-only"
	InstantiatePermissionEverybody     InstantiatePermission = "everybody"
)

type AdminPolicy string

const (
	AdminPolicyRequired AdminPolicy = "required"
)

type PinnedCodePolicy string

const (
	PinnedCodePolicyDisabled       PinnedCodePolicy = "disabled"
	PinnedCodePolicyGovernanceOnly PinnedCodePolicy = "governance-only"
)

type Policy struct {
	Enabled                      bool
	UploadPermission             UploadPermission
	InstantiatePermission        InstantiatePermission
	AdminPolicy                  AdminPolicy
	MigrationsEnabled            bool
	PinnedCodePolicy             PinnedCodePolicy
	GovernanceAuthority          string
	UploadAllowlist              []string
	MaxContractSizeBytes         uint64
	MaxProposalContractSizeBytes uint64
	SmartQueryGasLimit           uint64
	SimulationGasLimit           uint64
	GasMultiplier                uint64
	MemoryCacheSizeMiB           uint32
	MaxQueryResponseBytes        uint64
	MaxQueryDepth                uint32
	MaxPinnedCodes               uint32
}

func DefaultPolicy() Policy {
	return Policy{
		Enabled:                      false,
		UploadPermission:             UploadPermissionGovernanceOnly,
		InstantiatePermission:        InstantiatePermissionCodeOwnerOnly,
		AdminPolicy:                  AdminPolicyRequired,
		MigrationsEnabled:            true,
		PinnedCodePolicy:             PinnedCodePolicyDisabled,
		MaxContractSizeBytes:         DefaultMaxContractSizeBytes,
		MaxProposalContractSizeBytes: DefaultMaxProposalContractSizeBytes,
		SmartQueryGasLimit:           DefaultSmartQueryGasLimit,
		SimulationGasLimit:           DefaultSimulationGasLimit,
		GasMultiplier:                DefaultGasMultiplier,
		MemoryCacheSizeMiB:           DefaultMemoryCacheSizeMiB,
		MaxQueryResponseBytes:        DefaultMaxQueryResponseBytes,
		MaxQueryDepth:                DefaultMaxQueryDepth,
		MaxPinnedCodes:               DefaultMaxPinnedCodes,
	}
}

func validUploadPermission(permission UploadPermission) bool {
	switch permission {
	case UploadPermissionGovernanceOnly, UploadPermissionAllowlist:
		return true
	default:
		return false
	}
}

func validInstantiatePermission(permission InstantiatePermission) bool {
	switch permission {
	case InstantiatePermissionCodeOwnerOnly, InstantiatePermissionEverybody:
		return true
	default:
		return false
	}
}

func validPinnedCodePolicy(policy PinnedCodePolicy) bool {
	switch policy {
	case PinnedCodePolicyDisabled, PinnedCodePolicyGovernanceOnly:
		return true
	default:
		return false
	}
}
