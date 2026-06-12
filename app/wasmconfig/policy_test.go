package wasmconfig

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

var (
	governanceAddr	= testAEAddress(0x01)
	uploaderAddr	= testAEAddress(0x02)
	ownerAddr	= testAEAddress(0x03)
	contractAddr	= testAEAddress(0x04)
	attackerAddr	= testAEAddress(0x05)
)

func TestDefaultPolicyIsDisabledAndPinnedToCompatibleWasmd(t *testing.T) {
	policy := DefaultPolicy()

	require.False(t, policy.Enabled)
	require.Equal(t, "v0.70.2", RecommendedWasmdVersion)
	require.Equal(t, "v3.0.6", RecommendedWasmVMVersion)
	require.Equal(t, "v0.54", RecommendedSDKMinor)
	require.Equal(t, UploadPermissionGovernanceOnly, policy.UploadPermission)
	require.Equal(t, InstantiatePermissionCodeOwnerOnly, policy.InstantiatePermission)
	require.NoError(t, policy.Validate())
	require.ErrorContains(t, CanUpload(governanceAddr, policy), "disabled by feature gate")
}

func TestPolicyDefinesPhase11ReadinessSurface(t *testing.T) {
	policy := DefaultPolicy()

	require.False(t, policy.Enabled)
	require.Equal(t, UploadPermissionGovernanceOnly, policy.UploadPermission)
	require.Equal(t, InstantiatePermissionCodeOwnerOnly, policy.InstantiatePermission)
	require.Equal(t, AdminPolicyRequired, policy.AdminPolicy)
	require.True(t, policy.MigrationsEnabled)
	require.Equal(t, uint64(800*1024), policy.MaxContractSizeBytes)
	require.Equal(t, uint64(3*1024*1024), policy.MaxProposalContractSizeBytes)
	require.Equal(t, uint64(20_000_000), policy.MaxInstantiateGas)
	require.Equal(t, uint64(20_000_000), policy.MaxExecuteGasPerTx)
	require.Equal(t, uint64(3_000_000), policy.SmartQueryGasLimit)
	require.Equal(t, uint64(20_000_000), policy.SimulationGasLimit)
	require.Equal(t, uint64(140_000), policy.GasMultiplier)
	require.Equal(t, uint32(100), policy.MemoryCacheSizeMiB)
	require.Equal(t, uint64(256*1024), policy.MaxQueryResponseBytes)
	require.Equal(t, uint32(8), policy.MaxQueryDepth)
	require.Equal(t, PinnedCodePolicyDisabled, policy.PinnedCodePolicy)
	require.Equal(t, uint32(0), policy.MaxPinnedCodes)
	require.Equal(t, uint64(1_000), policy.ContractUploadFeeNaet)
	require.Equal(t, uint64(1), policy.StoragePricePerByteEpochNaet)
	require.NoError(t, policy.Validate())
}

func TestPolicyRejectsUnsafeLimits(t *testing.T) {
	policy := enabledPolicy()

	tooLargeCode := policy
	tooLargeCode.MaxContractSizeBytes = DefaultMaxContractSizeBytes + 1
	require.ErrorContains(t, tooLargeCode.Validate(), "max contract size")

	tooLargeInstantiateGas := policy
	tooLargeInstantiateGas.MaxInstantiateGas = maxInstantiateGas + 1
	require.ErrorContains(t, tooLargeInstantiateGas.Validate(), "instantiate gas")

	tooLargeExecuteGas := policy
	tooLargeExecuteGas.MaxExecuteGasPerTx = maxExecuteGasPerTx + 1
	require.ErrorContains(t, tooLargeExecuteGas.Validate(), "execute gas")

	tooLargeQueryGas := policy
	tooLargeQueryGas.SmartQueryGasLimit = maxSmartQueryGasLimit + 1
	require.ErrorContains(t, tooLargeQueryGas.Validate(), "smart query gas")

	unbenchmarkedGasMultiplier := policy
	unbenchmarkedGasMultiplier.GasMultiplier = DefaultGasMultiplier + 1
	require.ErrorContains(t, unbenchmarkedGasMultiplier.Validate(), "gas multiplier")

	tooMuchCache := policy
	tooMuchCache.MemoryCacheSizeMiB = maxMemoryCacheSizeMiB + 1
	require.ErrorContains(t, tooMuchCache.Validate(), "memory cache")

	tooLargeResponse := policy
	tooLargeResponse.MaxQueryResponseBytes = maxQueryResponseBytes + 1
	require.ErrorContains(t, tooLargeResponse.Validate(), "query response")

	tooDeepQuery := policy
	tooDeepQuery.MaxQueryDepth = maxQueryDepth + 1
	require.ErrorContains(t, tooDeepQuery.Validate(), "query depth")

	badPinned := policy
	badPinned.PinnedCodePolicy = PinnedCodePolicyGovernanceOnly
	badPinned.MaxPinnedCodes = 0
	require.ErrorContains(t, badPinned.Validate(), "pinning is enabled")

	badUploadFee := policy
	badUploadFee.ContractUploadFeeNaet = 0
	require.ErrorContains(t, badUploadFee.Validate(), "upload fee")

	badStoragePrice := policy
	badStoragePrice.StoragePricePerByteEpochNaet = 0
	require.ErrorContains(t, badStoragePrice.Validate(), "storage price")
}

func TestAllowlistRejectsMalformedEmptyAndZeroAddress(t *testing.T) {
	policy := enabledPolicy()
	policy.UploadPermission = UploadPermissionAllowlist

	policy.UploadAllowlist = nil
	require.ErrorContains(t, policy.Validate(), "allowlist must not be empty")

	policy.UploadAllowlist = []string{"ae1malformed"}
	require.Error(t, policy.Validate())

	policy.UploadAllowlist = []string{addressing.ZeroUserFriendly}
	require.ErrorContains(t, policy.Validate(), "must not be zero address")
}

func TestGovernanceOnlyUploadRequiresAuthority(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, CanUpload(governanceAddr, policy))
	require.ErrorContains(t, CanUpload(uploaderAddr, policy), "requires governance authority")

	policy.GovernanceAuthority = addressing.ZeroUserFriendly
	require.ErrorContains(t, CanUpload(governanceAddr, policy), "must not be zero address")
}

func TestGovernanceAuthorityEnablesAndDisablesCosmWasmOnlyIntentionally(t *testing.T) {
	current := DefaultPolicy()
	next := DefaultPolicy()
	next.Enabled = true
	next.GovernanceAuthority = governanceAddr

	require.NoError(t, CanUpdatePolicy(governanceAddr, current, next))
	require.ErrorContains(t, CanUpdatePolicy(attackerAddr, current, next), "requires governance authority")

	disabled := next
	disabled.Enabled = false
	require.NoError(t, CanUpdatePolicy(governanceAddr, next, disabled))

	badAuthority := next
	badAuthority.GovernanceAuthority = addressing.ZeroUserFriendly
	require.ErrorContains(t, CanUpdatePolicy(governanceAddr, current, badAuthority), "must not be zero address")
}

func TestAllowlistUploadInstantiateExecuteMigrateLifecycle(t *testing.T) {
	policy := enabledPolicy()
	policy.UploadPermission = UploadPermissionAllowlist
	policy.InstantiatePermission = InstantiatePermissionEverybody
	policy.UploadAllowlist = []string{uploaderAddr}
	require.NoError(t, policy.Validate())

	require.NoError(t, CanUpload(uploaderAddr, policy))
	require.NoError(t, CanInstantiate(ownerAddr, uploaderAddr, policy))
	require.NoError(t, ValidateInstantiateAddresses(ownerAddr, ownerAddr, policy))
	require.NoError(t, CanExecute(ownerAddr, contractAddr, policy))
	require.NoError(t, CanMigrate(ownerAddr, ownerAddr, policy))
	policy.MigrationsEnabled = false
	require.ErrorContains(t, CanMigrate(ownerAddr, ownerAddr, policy), "disabled by governance")
}

func TestInstantiateOwnerOnlyAndMigrationRejectAdminTakeover(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, CanInstantiate(ownerAddr, ownerAddr, policy))
	require.ErrorContains(t, CanInstantiate(attackerAddr, ownerAddr, policy), "requires code owner")
	require.ErrorContains(t, CanMigrate(attackerAddr, ownerAddr, policy), "requires contract admin")
	require.ErrorContains(t, CanMigrate(ownerAddr, "", policy), "empty address string")
	require.ErrorContains(t, CanMigrate(ownerAddr, addressing.ZeroUserFriendly, policy), "must not be zero address")
}

func TestGasCodeQueryAndPinnedCodePolicies(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, ValidateContractCodeSize(DefaultMaxContractSizeBytes, false, policy))
	require.ErrorContains(t, ValidateContractCodeSize(DefaultMaxContractSizeBytes+1, false, policy), "code size")
	require.NoError(t, ValidateContractCodeSize(DefaultMaxProposalContractSizeBytes, true, policy))
	require.ErrorContains(t, ValidateContractCodeSize(0, false, policy), "must not be empty")

	require.NoError(t, EnforceInstantiateGasLimit(policy.MaxInstantiateGas, policy))
	require.ErrorContains(t, EnforceInstantiateGasLimit(policy.MaxInstantiateGas+1, policy), "instantiate gas")
	require.NoError(t, EnforceExecuteGasLimit(policy.SimulationGasLimit, policy))
	require.ErrorContains(t, EnforceExecuteGasLimit(policy.MaxExecuteGasPerTx+1, policy), "execute gas")
	require.NoError(t, EnforceQueryLimit(policy.SmartQueryGasLimit, policy.MaxQueryResponseBytes, policy.MaxQueryDepth, policy))
	require.ErrorContains(t, EnforceQueryLimit(policy.SmartQueryGasLimit+1, 1, 1, policy), "query gas")
	require.ErrorContains(t, EnforceQueryLimit(1, policy.MaxQueryResponseBytes+1, 1, policy), "query response")
	require.ErrorContains(t, EnforceQueryLimit(1, 1, policy.MaxQueryDepth+1, policy), "query depth")

	require.ErrorContains(t, CanPinCode(governanceAddr, 0, policy), "pinned code is disabled")
	policy.PinnedCodePolicy = PinnedCodePolicyGovernanceOnly
	policy.MaxPinnedCodes = 2
	require.NoError(t, policy.Validate())
	require.NoError(t, CanPinCode(governanceAddr, 1, policy))
	require.ErrorContains(t, CanPinCode(attackerAddr, 1, policy), "requires governance authority")
	require.ErrorContains(t, CanPinCode(governanceAddr, 2, policy), "pinned code count")
}

func TestUploadFeeAndStoragePricingPolicy(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, EnforceContractUploadFee(sdkCoins("naet", int64(policy.ContractUploadFeeNaet)), policy))
	require.ErrorContains(t, EnforceContractUploadFee(sdkCoins("naet", int64(policy.ContractUploadFeeNaet-1)), policy), "upload fee")
	require.Error(t, EnforceContractUploadFee(sdkCoins("testtoken", int64(policy.ContractUploadFeeNaet)), policy))

	price, err := CalculateStoragePrice(10, 3, policy)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("naet", 30), price)
	_, err = CalculateStoragePrice(0, 3, policy)
	require.ErrorContains(t, err, "positive bytes")
	_, err = CalculateStoragePrice(10, 0, policy)
	require.ErrorContains(t, err, "positive bytes")
}

func TestCosmWasmCannotBypassNativeFeeOrZeroAddressPolicy(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, ValidateProtocolFees(naetFee()))
	require.Error(t, ValidateProtocolFees(sdkCoins("testtoken", 1)))
	require.Error(t, ValidateProtocolFees(sdkCoins("naet", 0)))

	require.ErrorContains(t, ValidateInstantiateAddresses(addressing.ZeroUserFriendly, ownerAddr, policy), "must not be zero address")
	require.ErrorContains(t, ValidateInstantiateAddresses(ownerAddr, addressing.ZeroUserFriendly, policy), "must not be zero address")
	require.ErrorContains(t, CanExecute(ownerAddr, addressing.ZeroUserFriendly, policy), "must not be zero address")
}

func enabledPolicy() Policy {
	policy := DefaultPolicy()
	policy.Enabled = true
	policy.GovernanceAuthority = governanceAddr
	return policy
}

func naetFee() sdk.Coins {
	return sdkCoins("naet", 1)
}

func sdkCoins(denom string, amount int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(denom, amount))
}

func testAEAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress{
		fill, fill, fill, fill, fill,
		fill, fill, fill, fill, fill,
		fill, fill, fill, fill, fill,
		fill, fill, fill, fill, fill,
	})
}
