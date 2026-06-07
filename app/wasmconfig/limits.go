package wasmconfig

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

func ValidateContractCodeSize(size uint64, proposal bool, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	limit := p.MaxContractSizeBytes
	if proposal {
		limit = p.MaxProposalContractSizeBytes
	}
	if size == 0 {
		return errors.New("wasm contract code must not be empty")
	}
	if size > limit {
		return fmt.Errorf("wasm contract code size must be <= %d bytes", limit)
	}
	return nil
}

func ValidateProtocolFees(fees sdk.Coins) error {
	return feestypes.ValidateFeeCoins(feestypes.DefaultParams(), fees, true)
}

func EnforceContractUploadFee(fees sdk.Coins, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := ValidateProtocolFees(fees); err != nil {
		return err
	}
	required := sdkmath.NewIntFromUint64(p.ContractUploadFeeNaet)
	if fees.AmountOf(feestypes.BondDenom).LT(required) {
		return fmt.Errorf("wasm contract upload fee must be at least %s%s", required.String(), feestypes.BondDenom)
	}
	return nil
}

func EnforceGasLimit(wanted uint64, limit uint64, label string) error {
	if wanted == 0 {
		return fmt.Errorf("wasm %s gas must be positive", label)
	}
	if wanted > limit {
		return fmt.Errorf("wasm %s gas must be <= %d", label, limit)
	}
	return nil
}

func EnforceInstantiateGasLimit(gas uint64, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	return EnforceGasLimit(gas, p.MaxInstantiateGas, "instantiate")
}

func EnforceExecuteGasLimit(gas uint64, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	return EnforceGasLimit(gas, p.MaxExecuteGasPerTx, "execute")
}

func EnforceQueryLimit(gas uint64, responseBytes uint64, depth uint32, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := EnforceGasLimit(gas, p.SmartQueryGasLimit, "query"); err != nil {
		return err
	}
	if responseBytes > p.MaxQueryResponseBytes {
		return fmt.Errorf("wasm query response bytes must be <= %d", p.MaxQueryResponseBytes)
	}
	if depth > p.MaxQueryDepth {
		return fmt.Errorf("wasm query depth must be <= %d", p.MaxQueryDepth)
	}
	return nil
}

func CalculateStoragePrice(storageBytes uint64, epochs uint64, p Policy) (sdk.Coin, error) {
	if err := requireEnabled(p); err != nil {
		return sdk.Coin{}, err
	}
	if storageBytes == 0 || epochs == 0 {
		return sdk.Coin{}, errors.New("wasm storage pricing requires positive bytes and epochs")
	}
	amount := sdkmath.NewIntFromUint64(p.StoragePricePerByteEpochNaet).
		Mul(sdkmath.NewIntFromUint64(storageBytes)).
		Mul(sdkmath.NewIntFromUint64(epochs))
	return sdk.NewCoin(feestypes.BondDenom, amount), nil
}

func validateLimits(p Policy) error {
	if p.MaxContractSizeBytes == 0 || p.MaxContractSizeBytes > DefaultMaxContractSizeBytes {
		return fmt.Errorf("wasm max contract size must be between 1 and %d bytes", DefaultMaxContractSizeBytes)
	}
	if p.MaxProposalContractSizeBytes < p.MaxContractSizeBytes ||
		p.MaxProposalContractSizeBytes > DefaultMaxProposalContractSizeBytes {
		return fmt.Errorf("wasm proposal contract size must be between max contract size and %d bytes", DefaultMaxProposalContractSizeBytes)
	}
	if p.MaxInstantiateGas == 0 || p.MaxInstantiateGas > maxInstantiateGas {
		return fmt.Errorf("wasm instantiate gas limit must be between 1 and %d", maxInstantiateGas)
	}
	if p.MaxExecuteGasPerTx == 0 || p.MaxExecuteGasPerTx > maxExecuteGasPerTx {
		return fmt.Errorf("wasm execute gas per tx must be between 1 and %d", maxExecuteGasPerTx)
	}
	if p.SmartQueryGasLimit == 0 || p.SmartQueryGasLimit > maxSmartQueryGasLimit {
		return fmt.Errorf("wasm smart query gas limit must be between 1 and %d", maxSmartQueryGasLimit)
	}
	if p.SimulationGasLimit == 0 || p.SimulationGasLimit > maxSimulationGasLimit {
		return fmt.Errorf("wasm simulation gas limit must be between 1 and %d", maxSimulationGasLimit)
	}
	if p.GasMultiplier != DefaultGasMultiplier {
		return fmt.Errorf("wasm gas multiplier must remain %d until benchmarked otherwise", DefaultGasMultiplier)
	}
	if p.MemoryCacheSizeMiB > maxMemoryCacheSizeMiB {
		return fmt.Errorf("wasm memory cache size must be at most %d MiB", maxMemoryCacheSizeMiB)
	}
	if p.MaxQueryResponseBytes == 0 || p.MaxQueryResponseBytes > maxQueryResponseBytes {
		return fmt.Errorf("wasm query response bytes must be between 1 and %d", maxQueryResponseBytes)
	}
	if p.MaxQueryDepth == 0 || p.MaxQueryDepth > maxQueryDepth {
		return fmt.Errorf("wasm query depth must be between 1 and %d", maxQueryDepth)
	}
	if p.MaxPinnedCodes > maxPinnedCodes {
		return fmt.Errorf("wasm pinned code count must be <= %d", maxPinnedCodes)
	}
	if p.ContractUploadFeeNaet == 0 || p.ContractUploadFeeNaet > maxContractUploadFeeNaet {
		return fmt.Errorf("wasm contract upload fee must be between 1 and %d naet", maxContractUploadFeeNaet)
	}
	if p.StoragePricePerByteEpochNaet == 0 || p.StoragePricePerByteEpochNaet > maxStoragePricePerByteEpochNaet {
		return fmt.Errorf("wasm storage price per byte epoch must be between 1 and %d naet", maxStoragePricePerByteEpochNaet)
	}
	if p.PinnedCodePolicy == PinnedCodePolicyGovernanceOnly && p.MaxPinnedCodes == 0 {
		return errors.New("wasm pinned code max count must be positive when pinning is enabled")
	}
	return nil
}
