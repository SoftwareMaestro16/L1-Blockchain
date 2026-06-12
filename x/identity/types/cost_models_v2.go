package types

import (
	"bytes"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	DefaultIdentityResolverFieldUpdateFeeNaet	= int64(2_000_000)
	DefaultIdentityResolverProofIndexImpactFeeNaet	= int64(1_000_000)
	DefaultIdentityResolverBatchCostFloorNaet	= int64(1_000_000)
	DefaultIdentityResolverRemovedByteCreditBps	= uint32(2_500)
	DefaultIdentityInlineInterfaceStorageMultiplier	= uint32(3)

	DefaultIdentitySubdomainMinimalChildRecordFeeNaet	= int64(500_000_000)
	DefaultIdentitySubdomainLabelByteFeeNaet		= int64(50_000_000)
	DefaultIdentitySubdomainZonePolicyUnitFeeNaet		= int64(250_000_000)

	IdentityDelegatedBillingParentV2	= IdentityDelegatedBillingPolicyV2("parent")
	IdentityDelegatedBillingDelegateV2	= IdentityDelegatedBillingPolicyV2("delegate")
)

type IdentityResolverUpdateCostParamsV2 struct {
	PayloadParams				IdentityResolverPayloadSafetyParamsV2
	FieldUpdateFee				sdkmath.Int
	ProofIndexImpactFee			sdkmath.Int
	BatchCostFloor				sdkmath.Int
	RemovedByteCreditBps			uint32
	InlineInterfaceStorageMultiplier	uint32
	ChurnSurchargeEnabled			bool
}

type IdentityResolverUpdateCostRequestV2 struct {
	Before			UnifiedResolutionRecordV2
	After			UnifiedResolutionRecordV2
	UpdatedFields		[]string
	UpdatesInWindow		uint32
	ProofIndexWrites	uint64
}

type IdentityResolverStorageDeltaV2 struct {
	BeforeBytes		uint64
	AfterBytes		uint64
	AddedBytes		uint64
	RemovedBytes		uint64
	NetGrowthBytes		uint64
	InlineBytes		uint64
	BillableBytes		uint64
	RemovedByteCredit	sdkmath.Int
}

type IdentityResolverUpdateFeeQuoteV2 struct {
	Denom			string
	FieldCount		uint64
	BaseUpdateFee		sdkmath.Int
	FieldUpdateFee		sdkmath.Int
	AddedStorageFee		sdkmath.Int
	RemovedByteCredit	sdkmath.Int
	InlineInterfaceFee	sdkmath.Int
	ChurnSurcharge		sdkmath.Int
	ChurnMultiplierBps	uint32
	ProofIndexImpactFee	sdkmath.Int
	Total			sdkmath.Int
	StorageDelta		IdentityResolverStorageDeltaV2
	DeterministicFormula	string
}

type IdentityResolverBatchFeeQuoteV2 struct {
	Denom				string
	ItemQuotes			[]IdentityResolverUpdateFeeQuoteV2
	IndividualTotal			sdkmath.Int
	BatchFloorTotal			sdkmath.Int
	Total				sdkmath.Int
	BatchSize			uint64
	NotCheaperThanIndividual	bool
	DeterministicFormula		string
}

type IdentityDelegatedBillingPolicyV2 string

type IdentitySubdomainCostParamsV2 struct {
	PricingParams		IdentityPricingParamsV2
	MinimalChildRecordFee	sdkmath.Int
	LabelByteFee		sdkmath.Int
	ZonePolicyUnitFee	sdkmath.Int
	StorageFeePerByte	sdkmath.Int
	DetachedRequiresPayment	bool
}

type IdentitySubdomainCostRequestV2 struct {
	State			IdentityState
	Policy			SubdomainCreationPolicyV2
	ResolverPayloadBytes	uint64
	ZonePolicyComplexity	uint64
	BillingPolicy		IdentityDelegatedBillingPolicyV2
}

type IdentitySubdomainFeeQuoteV2 struct {
	Denom			string
	ChildName		string
	ParentStatus		DomainLifecycleStatus
	ChildLabelLength	uint64
	DelegationType		SubdomainDelegationTypeV2
	Detached		bool
	BillingPolicy		IdentityDelegatedBillingPolicyV2
	ChargedAddress		sdk.AccAddress
	MinimalChildRecordFee	sdkmath.Int
	LabelLengthFee		sdkmath.Int
	ResolverPayloadFee	sdkmath.Int
	ZonePolicyComplexityFee	sdkmath.Int
	DetachedRegistrationFee	sdkmath.Int
	DetachedRenewalFee	sdkmath.Int
	Total			sdkmath.Int
	ExpiryDurationBlocks	uint64
	ChildExpiryHeight	uint64
	ParentExpiryHeight	uint64
	DeterministicFormula	string
}

func DefaultIdentityResolverUpdateCostParamsV2() IdentityResolverUpdateCostParamsV2 {
	return IdentityResolverUpdateCostParamsV2{
		PayloadParams:				DefaultIdentityResolverPayloadSafetyParamsV2(),
		FieldUpdateFee:				sdkmath.NewInt(DefaultIdentityResolverFieldUpdateFeeNaet),
		ProofIndexImpactFee:			sdkmath.NewInt(DefaultIdentityResolverProofIndexImpactFeeNaet),
		BatchCostFloor:				sdkmath.NewInt(DefaultIdentityResolverBatchCostFloorNaet),
		RemovedByteCreditBps:			DefaultIdentityResolverRemovedByteCreditBps,
		InlineInterfaceStorageMultiplier:	DefaultIdentityInlineInterfaceStorageMultiplier,
		ChurnSurchargeEnabled:			true,
	}
}

func ValidateIdentityResolverUpdateCostParamsV2(params IdentityResolverUpdateCostParamsV2) error {
	if err := ValidateIdentityResolverPayloadSafetyParamsV2(params.PayloadParams); err != nil {
		return err
	}
	for _, amount := range []struct {
		label	string
		value	sdkmath.Int
	}{
		{label: "field update fee", value: params.FieldUpdateFee},
		{label: "proof index impact fee", value: params.ProofIndexImpactFee},
		{label: "batch cost floor", value: params.BatchCostFloor},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity resolver update cost %s must not be negative", amount.label)
		}
	}
	if params.RemovedByteCreditBps > DomainDistributionDenominatorBps {
		return errors.New("identity resolver update removed byte credit bps must be <= 10000")
	}
	if params.InlineInterfaceStorageMultiplier == 0 {
		return errors.New("identity resolver update inline interface multiplier is required")
	}
	return nil
}

func CalculateIdentityResolverStorageDeltaV2(before UnifiedResolutionRecordV2, after UnifiedResolutionRecordV2, params IdentityResolverUpdateCostParamsV2) (IdentityResolverStorageDeltaV2, error) {
	if err := ValidateIdentityResolverUpdateCostParamsV2(params); err != nil {
		return IdentityResolverStorageDeltaV2{}, err
	}
	if after.NameHash != "" {
		if err := ValidateUnifiedResolverPayloadSafetyV2(after, params.PayloadParams); err != nil {
			return IdentityResolverStorageDeltaV2{}, err
		}
	}
	beforeBytes := EstimateUnifiedResolverPayloadBytesV2(before)
	afterBytes := EstimateUnifiedResolverPayloadBytesV2(after)
	delta := IdentityResolverStorageDeltaV2{
		BeforeBytes:		beforeBytes,
		AfterBytes:		afterBytes,
		InlineBytes:		EstimateIdentityInlineInterfaceBytesV2(after),
		RemovedByteCredit:	sdkmath.ZeroInt(),
	}
	switch {
	case afterBytes >= beforeBytes:
		delta.AddedBytes = afterBytes - beforeBytes
		delta.NetGrowthBytes = delta.AddedBytes
	default:
		delta.RemovedBytes = beforeBytes - afterBytes
		delta.RemovedByteCredit = ceilBps(params.PayloadParams.StorageFeePerByte.Mul(sdkmath.NewIntFromUint64(delta.RemovedBytes)), params.RemovedByteCreditBps)
	}
	delta.BillableBytes = delta.NetGrowthBytes + delta.InlineBytes*uint64(params.InlineInterfaceStorageMultiplier-1)
	return delta, nil
}

func QuoteIdentityResolverUpdateFeeV2(req IdentityResolverUpdateCostRequestV2, params IdentityResolverUpdateCostParamsV2) (IdentityResolverUpdateFeeQuoteV2, error) {
	if err := ValidateIdentityResolverUpdateCostParamsV2(params); err != nil {
		return IdentityResolverUpdateFeeQuoteV2{}, err
	}
	delta, err := CalculateIdentityResolverStorageDeltaV2(req.Before, req.After, params)
	if err != nil {
		return IdentityResolverUpdateFeeQuoteV2{}, err
	}
	fieldCount := uint64(len(uniqueResolverUpdateFieldsV2(req.UpdatedFields)))
	if fieldCount == 0 {
		fieldCount = CountResolverUpdatedFieldsV2(req.Before, req.After)
	}
	multiplier := uint32(DomainDistributionDenominatorBps)
	if params.ChurnSurchargeEnabled {
		multiplier = IdentityResolverChurnMultiplierBpsV2(req.UpdatesInWindow, params.PayloadParams)
	}
	baseWithChurn := params.PayloadParams.UpdateBaseFee.MulRaw(int64(multiplier)).QuoRaw(int64(DomainDistributionDenominatorBps))
	churnSurcharge := baseWithChurn.Sub(params.PayloadParams.UpdateBaseFee)
	if churnSurcharge.IsNegative() {
		churnSurcharge = sdkmath.ZeroInt()
	}
	fieldFee := params.FieldUpdateFee.Mul(sdkmath.NewIntFromUint64(fieldCount))
	addedStorageFee := params.PayloadParams.StorageFeePerByte.Mul(sdkmath.NewIntFromUint64(delta.NetGrowthBytes))
	inlineFee := params.PayloadParams.StorageFeePerByte.Mul(sdkmath.NewIntFromUint64(delta.InlineBytes)).MulRaw(int64(params.InlineInterfaceStorageMultiplier - 1))
	proofFee := params.ProofIndexImpactFee.Mul(sdkmath.NewIntFromUint64(req.ProofIndexWrites))
	total := params.PayloadParams.UpdateBaseFee.
		Add(fieldFee).
		Add(addedStorageFee).
		Add(inlineFee).
		Add(churnSurcharge).
		Add(proofFee)
	if delta.RemovedByteCredit.IsPositive() {
		total = total.Sub(delta.RemovedByteCredit)
		if total.LT(params.PayloadParams.UpdateBaseFee) {
			total = params.PayloadParams.UpdateBaseFee
		}
	}
	return IdentityResolverUpdateFeeQuoteV2{
		Denom:			appparams.BaseDenom,
		FieldCount:		fieldCount,
		BaseUpdateFee:		params.PayloadParams.UpdateBaseFee,
		FieldUpdateFee:		fieldFee,
		AddedStorageFee:	addedStorageFee,
		RemovedByteCredit:	delta.RemovedByteCredit,
		InlineInterfaceFee:	inlineFee,
		ChurnSurcharge:		churnSurcharge,
		ChurnMultiplierBps:	multiplier,
		ProofIndexImpactFee:	proofFee,
		Total:			total,
		StorageDelta:		delta,
		DeterministicFormula:	"base_update_fee + unique_updated_fields*field_fee + net_storage_growth*fee_per_byte + inline_schema_bytes*(multiplier-1)*fee_per_byte + churn_surcharge + proof_index_writes*fee - removed_byte_credit_floor_base",
	}, nil
}

func QuoteIdentityResolverBatchUpdateFeesV2(requests []IdentityResolverUpdateCostRequestV2, params IdentityResolverUpdateCostParamsV2) (IdentityResolverBatchFeeQuoteV2, error) {
	if err := ValidateIdentityResolverUpdateCostParamsV2(params); err != nil {
		return IdentityResolverBatchFeeQuoteV2{}, err
	}
	quote := IdentityResolverBatchFeeQuoteV2{
		Denom:			appparams.BaseDenom,
		ItemQuotes:		make([]IdentityResolverUpdateFeeQuoteV2, 0, len(requests)),
		IndividualTotal:	sdkmath.ZeroInt(),
		BatchFloorTotal:	params.BatchCostFloor.Mul(sdkmath.NewIntFromUint64(uint64(len(requests)))),
		BatchSize:		uint64(len(requests)),
		DeterministicFormula:	"max(sum(individual_update_fees), batch_size*batch_cost_floor)",
	}
	for _, request := range requests {
		item, err := QuoteIdentityResolverUpdateFeeV2(request, params)
		if err != nil {
			return IdentityResolverBatchFeeQuoteV2{}, err
		}
		quote.ItemQuotes = append(quote.ItemQuotes, item)
		quote.IndividualTotal = quote.IndividualTotal.Add(item.Total)
	}
	quote.Total = quote.IndividualTotal
	if quote.Total.LT(quote.BatchFloorTotal) {
		quote.Total = quote.BatchFloorTotal
	}
	quote.NotCheaperThanIndividual = !quote.Total.LT(quote.IndividualTotal)
	return quote, nil
}

func DefaultIdentitySubdomainCostParamsV2() IdentitySubdomainCostParamsV2 {
	return IdentitySubdomainCostParamsV2{
		PricingParams:			DefaultIdentityPricingParamsV2(),
		MinimalChildRecordFee:		sdkmath.NewInt(DefaultIdentitySubdomainMinimalChildRecordFeeNaet),
		LabelByteFee:			sdkmath.NewInt(DefaultIdentitySubdomainLabelByteFeeNaet),
		ZonePolicyUnitFee:		sdkmath.NewInt(DefaultIdentitySubdomainZonePolicyUnitFeeNaet),
		StorageFeePerByte:		sdkmath.NewInt(DefaultIdentityResolverStorageFeePerByte),
		DetachedRequiresPayment:	true,
	}
}

func ValidateIdentitySubdomainCostParamsV2(params IdentitySubdomainCostParamsV2) error {
	if err := ValidateIdentityPricingParamsV2(params.PricingParams); err != nil {
		return err
	}
	for _, amount := range []struct {
		label	string
		value	sdkmath.Int
	}{
		{label: "minimal child record fee", value: params.MinimalChildRecordFee},
		{label: "label byte fee", value: params.LabelByteFee},
		{label: "zone policy unit fee", value: params.ZonePolicyUnitFee},
		{label: "storage fee per byte", value: params.StorageFeePerByte},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity subdomain cost %s must not be negative", amount.label)
		}
	}
	return nil
}

func QuoteIdentitySubdomainCreationFeeV2(req IdentitySubdomainCostRequestV2, params IdentitySubdomainCostParamsV2) (IdentitySubdomainFeeQuoteV2, error) {
	if err := ValidateIdentitySubdomainCostParamsV2(params); err != nil {
		return IdentitySubdomainFeeQuoteV2{}, err
	}
	state := normalizeIdentityStateParams(req.State)
	childName, err := ValidateSubdomainCreationV2(state, req.Policy)
	if err != nil {
		return IdentitySubdomainFeeQuoteV2{}, err
	}
	parent, err := requireActiveDomain(state, req.Policy.ParentName, req.Policy.Height)
	if err != nil {
		return IdentitySubdomainFeeQuoteV2{}, err
	}
	parentStatus, err := DomainLifecycle(state, parent.Name, req.Policy.Height)
	if err != nil {
		return IdentitySubdomainFeeQuoteV2{}, err
	}
	delegationType := req.Policy.DelegationType
	if delegationType == "" {
		delegationType = SubdomainDelegationOwnerControlledV2
	}
	childExpiry := req.Policy.ChildExpiryHeight
	if childExpiry == 0 {
		childExpiry = parent.ExpiryHeight
	}
	billing := normalizeDelegatedBillingPolicyV2(req.BillingPolicy)
	charged := cloneSpecAddress(parent.Owner)
	if delegationType == SubdomainDelegationDelegateControlledV2 && billing == IdentityDelegatedBillingDelegateV2 {
		charged = cloneSpecAddress(req.Policy.Actor)
	}
	labelFee := params.LabelByteFee.Mul(sdkmath.NewIntFromUint64(uint64(len(req.Policy.Label))))
	resolverFee := params.StorageFeePerByte.Mul(sdkmath.NewIntFromUint64(req.ResolverPayloadBytes))
	zoneFee := params.ZonePolicyUnitFee.Mul(sdkmath.NewIntFromUint64(req.ZonePolicyComplexity))
	quote := IdentitySubdomainFeeQuoteV2{
		Denom:				appparams.BaseDenom,
		ChildName:			childName,
		ParentStatus:			parentStatus,
		ChildLabelLength:		uint64(len(req.Policy.Label)),
		DelegationType:			delegationType,
		Detached:			req.Policy.DetachedPaid,
		BillingPolicy:			billing,
		ChargedAddress:			charged,
		MinimalChildRecordFee:		params.MinimalChildRecordFee,
		LabelLengthFee:			labelFee,
		ResolverPayloadFee:		resolverFee,
		ZonePolicyComplexityFee:	zoneFee,
		DetachedRegistrationFee:	sdkmath.ZeroInt(),
		DetachedRenewalFee:		sdkmath.ZeroInt(),
		ExpiryDurationBlocks:		childExpiry - req.Policy.Height,
		ChildExpiryHeight:		childExpiry,
		ParentExpiryHeight:		parent.ExpiryHeight,
		DeterministicFormula:		"minimal_child_record_fee + label_bytes*label_fee + resolver_payload_bytes*storage_fee + zone_policy_complexity*unit_fee + optional_detached_registration_and_renewal",
	}
	quote.Total = quote.MinimalChildRecordFee.Add(quote.LabelLengthFee).Add(quote.ResolverPayloadFee).Add(quote.ZonePolicyComplexityFee)
	if req.Policy.DetachedPaid {
		if params.DetachedRequiresPayment && !req.Policy.IndependentPayment {
			return IdentitySubdomainFeeQuoteV2{}, errors.New("identity subdomain detached pricing requires independent payment")
		}
		detached, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
			Name:			childName,
			DurationBlocks:		quote.ExpiryDurationBlocks,
			ResolverPayloadBytes:	req.ResolverPayloadBytes,
			SubdomainMode:		IdentitySubdomainModeDetachedPaidV2,
		}, params.PricingParams)
		if err != nil {
			return IdentitySubdomainFeeQuoteV2{}, err
		}
		renewal, err := QuoteIdentityRenewalPriceV2(childName, 1, req.ResolverPayloadBytes, params.PricingParams)
		if err != nil {
			return IdentitySubdomainFeeQuoteV2{}, err
		}
		quote.DetachedRegistrationFee = detached.Total
		quote.DetachedRenewalFee = renewal.Total
		quote.Total = quote.Total.Add(quote.DetachedRegistrationFee)
	}
	return quote, nil
}

func EstimateIdentityInlineInterfaceBytesV2(record UnifiedResolutionRecordV2) uint64 {
	var total uint64
	for _, descriptor := range record.InterfaceDescriptors {
		total += uint64(len(descriptor.SchemaInlineOptional))
	}
	return total
}

func CountResolverUpdatedFieldsV2(before UnifiedResolutionRecordV2, after UnifiedResolutionRecordV2) uint64 {
	var count uint64
	if !bytes.Equal(before.PrimaryAddress, after.PrimaryAddress) {
		count++
	}
	if len(before.ContractTargets) != len(after.ContractTargets) {
		count++
	}
	if len(before.ServiceEndpoints) != len(after.ServiceEndpoints) {
		count++
	}
	if len(before.InterfaceDescriptors) != len(after.InterfaceDescriptors) || EstimateIdentityInlineInterfaceBytesV2(before) != EstimateIdentityInlineInterfaceBytesV2(after) {
		count++
	}
	if before.RoutingMetadata.RouteID != after.RoutingMetadata.RouteID ||
		before.RoutingMetadata.TargetType != after.RoutingMetadata.TargetType ||
		before.RoutingMetadata.PreferredTarget != after.RoutingMetadata.PreferredTarget ||
		before.RoutingMetadata.ZoneID != after.RoutingMetadata.ZoneID {
		count++
	}
	if len(before.ExecutionHints) != len(after.ExecutionHints) {
		count++
	}
	if count == 0 && EstimateUnifiedResolverPayloadBytesV2(before) != EstimateUnifiedResolverPayloadBytesV2(after) {
		count = 1
	}
	return count
}

func uniqueResolverUpdateFieldsV2(fields []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, found := seen[field]; found {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func normalizeDelegatedBillingPolicyV2(policy IdentityDelegatedBillingPolicyV2) IdentityDelegatedBillingPolicyV2 {
	switch policy {
	case IdentityDelegatedBillingDelegateV2:
		return IdentityDelegatedBillingDelegateV2
	default:
		return IdentityDelegatedBillingParentV2
	}
}
