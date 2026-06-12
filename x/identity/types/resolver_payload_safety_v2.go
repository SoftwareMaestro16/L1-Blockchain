package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	DefaultIdentityResolverUpdateBaseFeeNaet	= int64(5_000_000)
	DefaultIdentityResolverChurnWindowBlocks	= uint64(100)
	DefaultIdentityResolverChurnFreeUpdates		= uint32(3)
	DefaultIdentityResolverChurnMultiplierStepBps	= uint32(2_500)
	DefaultIdentityResolverChurnMaxMultiplierBps	= uint32(40_000)
	IdentityResolverPayloadSafetyFormulaV2		= "base_fee*churn_multiplier + payload_bytes*storage_fee_per_byte"
)

type IdentityResolverPayloadSafetyParamsV2 struct {
	MaxRecordBytes		uint64
	StorageFeePerByte	sdkmath.Int
	UpdateBaseFee		sdkmath.Int
	ChurnWindowBlocks	uint64
	FreeUpdatesPerWindow	uint32
	ChurnMultiplierStepBps	uint32
	MaxChurnMultiplierBps	uint32
}

type IdentityResolverPayloadFeeQuoteV2 struct {
	PayloadBytes		uint64
	StorageFee		sdkmath.Int
	BaseUpdateFee		sdkmath.Int
	ChurnMultiplierBps	uint32
	TotalUpdateFee		sdkmath.Int
	Formula			string
}

func DefaultIdentityResolverPayloadSafetyParamsV2() IdentityResolverPayloadSafetyParamsV2 {
	return IdentityResolverPayloadSafetyParamsV2{
		MaxRecordBytes:		MaxUnifiedPayloadBytesV2,
		StorageFeePerByte:	sdkmath.NewInt(DefaultIdentityResolverStorageFeePerByte),
		UpdateBaseFee:		sdkmath.NewInt(DefaultIdentityResolverUpdateBaseFeeNaet),
		ChurnWindowBlocks:	DefaultIdentityResolverChurnWindowBlocks,
		FreeUpdatesPerWindow:	DefaultIdentityResolverChurnFreeUpdates,
		ChurnMultiplierStepBps:	DefaultIdentityResolverChurnMultiplierStepBps,
		MaxChurnMultiplierBps:	DefaultIdentityResolverChurnMaxMultiplierBps,
	}
}

func ValidateIdentityResolverPayloadSafetyParamsV2(params IdentityResolverPayloadSafetyParamsV2) error {
	if params.MaxRecordBytes == 0 || params.MaxRecordBytes > MaxUnifiedPayloadBytesV2 {
		return fmt.Errorf("identity v2 resolver payload max record bytes must be in 1..%d", MaxUnifiedPayloadBytesV2)
	}
	if params.StorageFeePerByte.IsNil() || params.StorageFeePerByte.IsNegative() {
		return fmt.Errorf("identity v2 resolver payload storage fee per byte must not be negative")
	}
	if params.UpdateBaseFee.IsNil() || params.UpdateBaseFee.IsNegative() {
		return fmt.Errorf("identity v2 resolver payload update base fee must not be negative")
	}
	if params.ChurnWindowBlocks == 0 {
		return fmt.Errorf("identity v2 resolver payload churn window is required")
	}
	if params.MaxChurnMultiplierBps < DomainDistributionDenominatorBps {
		return fmt.Errorf("identity v2 resolver payload max churn multiplier must be at least 10000 bps")
	}
	return nil
}

func EstimateUnifiedResolverPayloadBytesV2(record UnifiedResolutionRecordV2) uint64 {
	total := len(record.NameHash) + len(record.Owner) + len(record.PrimaryAddress) + len(record.OwnerSignatureOptional)
	for _, target := range record.ContractTargets {
		total += len(target.Key) + len(target.Address) + len(target.CodeID) + len(target.TargetID) + len(target.ContractAddress) +
			len(target.Entrypoint) + len(target.InterfaceHash) + len(target.RequiredFundsPolicy) + 32
	}
	for _, endpoint := range record.ServiceEndpoints {
		total += len(endpoint.Key) + len(endpoint.Endpoint) + len(endpoint.ServiceID) + len(endpoint.ServiceType) +
			len(endpoint.Transport) + len(endpoint.AuthPolicy) + len(endpoint.HealthPathOptional) + len(endpoint.SchemaHashOptional) + 32
	}
	for _, descriptor := range record.InterfaceDescriptors {
		total += len(descriptor.InterfaceID) + len(descriptor.Descriptor) + len(descriptor.SchemaHash) + len(descriptor.SchemaURIOptional) +
			len(descriptor.SchemaInlineOptional) + len(descriptor.Version) + len(descriptor.RenderPolicy) +
			len(descriptor.ContractTargetIDOptional) + len(descriptor.ServiceIDOptional) + 32
		for _, permission := range descriptor.PermissionsRequired {
			total += len(permission)
		}
	}
	total += len(record.RoutingMetadata.ZoneID) + len(record.RoutingMetadata.ShardID) + len(record.RoutingMetadata.VM) +
		len(record.RoutingMetadata.Entrypoint) + len(record.RoutingMetadata.RouteID) + len(record.RoutingMetadata.TargetType) +
		len(record.RoutingMetadata.PreferredTarget) + len(record.RoutingMetadata.ChainContext) + len(record.RoutingMetadata.FeeHint) +
		len(record.RoutingMetadata.MemoPolicy) + 32
	for _, fallback := range record.RoutingMetadata.FallbackTargets {
		total += len(fallback)
	}
	for _, capability := range record.RoutingMetadata.CapabilityRequirements {
		total += len(capability)
	}
	for _, hint := range record.ExecutionHints {
		total += len(hint.Key) + len(hint.Value) + len(hint.PreferredFeeMode) + len(hint.MessageType) + 32
	}
	return uint64(total)
}

func ValidateUnifiedResolverPayloadSafetyV2(record UnifiedResolutionRecordV2, params IdentityResolverPayloadSafetyParamsV2) error {
	if err := ValidateIdentityResolverPayloadSafetyParamsV2(params); err != nil {
		return err
	}
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return err
	}
	payloadBytes := EstimateUnifiedResolverPayloadBytesV2(record)
	if payloadBytes > params.MaxRecordBytes {
		return fmt.Errorf("identity v2 resolver payload size %d exceeds configured maximum %d", payloadBytes, params.MaxRecordBytes)
	}
	return nil
}

func QuoteIdentityResolverPayloadUpdateFeeV2(record UnifiedResolutionRecordV2, updatesInWindow uint32, params IdentityResolverPayloadSafetyParamsV2) (IdentityResolverPayloadFeeQuoteV2, error) {
	if err := ValidateUnifiedResolverPayloadSafetyV2(record, params); err != nil {
		return IdentityResolverPayloadFeeQuoteV2{}, err
	}
	payloadBytes := EstimateUnifiedResolverPayloadBytesV2(record)
	storageFee := params.StorageFeePerByte.Mul(sdkmath.NewIntFromUint64(payloadBytes))
	multiplier := IdentityResolverChurnMultiplierBpsV2(updatesInWindow, params)
	baseFee := params.UpdateBaseFee.MulRaw(int64(multiplier)).QuoRaw(int64(DomainDistributionDenominatorBps))
	return IdentityResolverPayloadFeeQuoteV2{
		PayloadBytes:		payloadBytes,
		StorageFee:		storageFee,
		BaseUpdateFee:		params.UpdateBaseFee,
		ChurnMultiplierBps:	multiplier,
		TotalUpdateFee:		baseFee.Add(storageFee),
		Formula:		IdentityResolverPayloadSafetyFormulaV2,
	}, nil
}

func IdentityResolverChurnMultiplierBpsV2(updatesInWindow uint32, params IdentityResolverPayloadSafetyParamsV2) uint32 {
	multiplier := uint32(DomainDistributionDenominatorBps)
	if updatesInWindow <= params.FreeUpdatesPerWindow {
		return multiplier
	}
	excess := updatesInWindow - params.FreeUpdatesPerWindow
	multiplier += excess * params.ChurnMultiplierStepBps
	if multiplier > params.MaxChurnMultiplierBps {
		return params.MaxChurnMultiplierBps
	}
	return multiplier
}
