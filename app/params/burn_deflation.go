package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	BurnSourceFees		= "fees"
	BurnSourceSlashing	= "slashing"

	BurnEventFundsRemoved	= "burn_funds_removed"

	DefaultFeeSpikeDeflationThresholdBps	= int64(20_000)
	DefaultBondedStakeSafetyThresholdBps	= int64(5_000)
	DefaultDeflationReserveCapNaet		= int64(10_000_000_000)
	DefaultSecurityRewardFloorNaet		= int64(0)
	DefaultNetIssuanceFloorNaet		= int64(0)
)

type BurnMechanicsParams struct {
	EpochBurnCapNaet		sdkmath.Int
	BurnFloorNaet			sdkmath.Int
	NetIssuanceFloorNaet		sdkmath.Int
	DeflationReserveCapNaet		sdkmath.Int
	SecurityRewardFloorNaet		sdkmath.Int
	FeeSpikeThresholdBps		int64
	BondedStakeSafetyThresholdBps	int64
	GovernanceAllowsBelowFloor	bool
}

type BurnAccountingEvent struct {
	Type				string
	EpochID				uint64
	BlockHeight			uint64
	Source				string
	BurnedNaet			sdkmath.Int
	CumulativeBurnedNaet		sdkmath.Int
	EpochBurnedNaet			sdkmath.Int
	RemovedFromSpendableSupply	bool
}

type DeflationGuardTelemetry struct {
	Active				bool
	Reasons				[]string
	ProposedBurnNaet		sdkmath.Int
	FinalBurnNaet			sdkmath.Int
	ReducedBurnNaet			sdkmath.Int
	DivertedToReserveNaet		sdkmath.Int
	DivertedToSecurityRewardsNaet	sdkmath.Int
	NetIssuanceFloorNaet		sdkmath.Int
	NetIssuanceAfterBurnNaet	sdkmath.Int
	EpochBurnCapNaet		sdkmath.Int
	SecurityRewardFloorNaet		sdkmath.Int
}

type BurnIntegratedFeeDistributionInput struct {
	EpochID				uint64
	BlockHeight			uint64
	CollectedFeesNaet		sdkmath.Int
	BurnRatioBps			int64
	CommunityPoolRatioBps		int64
	StateMaintenanceReserveBps	int64
	SecurityReserveRatioBps		int64
	ExistingEpochBurnedNaet		sdkmath.Int
	GrossMintedNaet			sdkmath.Int
	CumulativeBurnedNaet		sdkmath.Int
	FeeSpikeBps			int64
	BondedStakeRatioBps		int64
	Params				BurnMechanicsParams
}

type BurnIntegratedFeeDistributionOutput struct {
	CollectedFeesNaet		sdkmath.Int
	BurnNaet			sdkmath.Int
	ValidatorRewardNaet		sdkmath.Int
	CommunityPoolNaet		sdkmath.Int
	StateMaintenanceReserveNaet	sdkmath.Int
	SecurityReserveNaet		sdkmath.Int
	DeflationReserveNaet		sdkmath.Int
	CumulativeBurnedNaet		sdkmath.Int
	EpochBurnedNaet			sdkmath.Int
	Events				[]BurnAccountingEvent
	DeflationGuard			DeflationGuardTelemetry
}

type BurnIntegratedSlashingDistributionInput struct {
	EpochID			uint64
	BlockHeight		uint64
	PenaltyNaet		sdkmath.Int
	BurnRatioBps		int64
	TreasuryRatioBps	int64
	ReporterRewardBps	int64
	ExistingEpochBurnedNaet	sdkmath.Int
	GrossMintedNaet		sdkmath.Int
	CumulativeBurnedNaet	sdkmath.Int
	Params			BurnMechanicsParams
}

type BurnIntegratedSlashingDistributionOutput struct {
	PenaltyNaet		sdkmath.Int
	BurnNaet		sdkmath.Int
	TreasuryNaet		sdkmath.Int
	ReporterRewardNaet	sdkmath.Int
	ValidatorPoolNaet	sdkmath.Int
	DeflationReserveNaet	sdkmath.Int
	CumulativeBurnedNaet	sdkmath.Int
	EpochBurnedNaet		sdkmath.Int
	Events			[]BurnAccountingEvent
	DeflationGuard		DeflationGuardTelemetry
}

type BurnAccountingInvariantReport struct {
	Passed	bool
	Failed	[]string
}

type BurnSupplyQueryInput struct {
	CumulativeBurnedNaet	sdkmath.Int
	Events			[]BurnAccountingEvent
	CurrentBlockHeight	uint64
	RecentWindowBlocks	uint64
}

type BurnSupplyQueryOutput struct {
	CumulativeBurnedNaet		sdkmath.Int
	RecentBurnedNaet		sdkmath.Int
	RecentBurnRateNaetPerBlock	sdkmath.Int
	EventCount			int
}

func DefaultBurnMechanicsParams() BurnMechanicsParams {
	return BurnMechanicsParams{
		BurnFloorNaet:			sdkmath.ZeroInt(),
		NetIssuanceFloorNaet:		sdkmath.NewInt(DefaultNetIssuanceFloorNaet),
		DeflationReserveCapNaet:	sdkmath.NewInt(DefaultDeflationReserveCapNaet),
		SecurityRewardFloorNaet:	sdkmath.NewInt(DefaultSecurityRewardFloorNaet),
		FeeSpikeThresholdBps:		DefaultFeeSpikeDeflationThresholdBps,
		BondedStakeSafetyThresholdBps:	DefaultBondedStakeSafetyThresholdBps,
	}
}

func ComputeBurnIntegratedFeeDistribution(input BurnIntegratedFeeDistributionInput) (BurnIntegratedFeeDistributionOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return BurnIntegratedFeeDistributionOutput{}, err
	}
	if input.EpochID == 0 {
		return BurnIntegratedFeeDistributionOutput{}, fmt.Errorf("epoch_id must be positive")
	}
	if err := validateBurnFeeInput(input); err != nil {
		return BurnIntegratedFeeDistributionOutput{}, err
	}

	fees := normalizeInt(input.CollectedFeesNaet)
	proposedBurn := ApplyBps(fees, input.BurnRatioBps)
	community := ApplyBps(fees, input.CommunityPoolRatioBps)
	stateReserve := ApplyBps(fees, input.StateMaintenanceReserveBps)
	securityReserve := ApplyBps(fees, input.SecurityReserveRatioBps)
	validator := fees.Sub(proposedBurn).Sub(community).Sub(stateReserve).Sub(securityReserve)

	guard := applyDeflationGuard(deflationGuardInput{
		Source:			BurnSourceFees,
		ProposedBurnNaet:	proposedBurn,
		ExistingEpochBurned:	normalizeInt(input.ExistingEpochBurnedNaet),
		GrossMintedNaet:	normalizeInt(input.GrossMintedNaet),
		CurrentSecurityReward:	validator,
		FeeSpikeBps:		input.FeeSpikeBps,
		BondedStakeRatioBps:	input.BondedStakeRatioBps,
		Params:			params,
	})
	validator = validator.Add(guard.DivertedToSecurityRewardsNaet)
	deflationReserve := guard.DivertedToReserveNaet

	output := BurnIntegratedFeeDistributionOutput{
		CollectedFeesNaet:		fees,
		BurnNaet:			guard.FinalBurnNaet,
		ValidatorRewardNaet:		validator,
		CommunityPoolNaet:		community,
		StateMaintenanceReserveNaet:	stateReserve,
		SecurityReserveNaet:		securityReserve,
		DeflationReserveNaet:		deflationReserve,
		EpochBurnedNaet:		normalizeInt(input.ExistingEpochBurnedNaet).Add(guard.FinalBurnNaet),
		CumulativeBurnedNaet:		normalizeInt(input.CumulativeBurnedNaet).Add(guard.FinalBurnNaet),
		DeflationGuard:			guard,
	}
	output.Events = burnAccountingEvents(input.EpochID, input.BlockHeight, BurnSourceFees, output.BurnNaet, output.CumulativeBurnedNaet, output.EpochBurnedNaet)
	return output, nil
}

func ComputeBurnIntegratedSlashingDistribution(input BurnIntegratedSlashingDistributionInput) (BurnIntegratedSlashingDistributionOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return BurnIntegratedSlashingDistributionOutput{}, err
	}
	if input.EpochID == 0 {
		return BurnIntegratedSlashingDistributionOutput{}, fmt.Errorf("epoch_id must be positive")
	}
	flow, err := ComputeSlashingEconomyFlow(SlashingEconomyFlowInput{
		PenaltyNaet:		input.PenaltyNaet,
		BurnRatioBps:		input.BurnRatioBps,
		TreasuryRatioBps:	input.TreasuryRatioBps,
		ReporterRewardBps:	input.ReporterRewardBps,
	})
	if err != nil {
		return BurnIntegratedSlashingDistributionOutput{}, err
	}

	guard := applyDeflationGuard(deflationGuardInput{
		Source:			BurnSourceSlashing,
		ProposedBurnNaet:	flow.BurnNaet,
		ExistingEpochBurned:	normalizeInt(input.ExistingEpochBurnedNaet),
		GrossMintedNaet:	normalizeInt(input.GrossMintedNaet),
		Params:			params,
	})

	output := BurnIntegratedSlashingDistributionOutput{
		PenaltyNaet:		flow.PenaltyNaet,
		BurnNaet:		guard.FinalBurnNaet,
		TreasuryNaet:		flow.TreasuryNaet,
		ReporterRewardNaet:	flow.ReporterRewardNaet,
		ValidatorPoolNaet:	flow.ValidatorPoolNaet,
		DeflationReserveNaet:	guard.DivertedToReserveNaet,
		EpochBurnedNaet:	normalizeInt(input.ExistingEpochBurnedNaet).Add(guard.FinalBurnNaet),
		CumulativeBurnedNaet:	normalizeInt(input.CumulativeBurnedNaet).Add(guard.FinalBurnNaet),
		DeflationGuard:		guard,
	}
	output.Events = burnAccountingEvents(input.EpochID, input.BlockHeight, BurnSourceSlashing, output.BurnNaet, output.CumulativeBurnedNaet, output.EpochBurnedNaet)
	return output, nil
}

func ValidateBurnFeeDistributionInvariants(output BurnIntegratedFeeDistributionOutput, params BurnMechanicsParams) BurnAccountingInvariantReport {
	params = params.withDefaults()
	failed := make([]string, 0)
	total := normalizeInt(output.BurnNaet).
		Add(normalizeInt(output.ValidatorRewardNaet)).
		Add(normalizeInt(output.CommunityPoolNaet)).
		Add(normalizeInt(output.StateMaintenanceReserveNaet)).
		Add(normalizeInt(output.SecurityReserveNaet)).
		Add(normalizeInt(output.DeflationReserveNaet))
	if !normalizeInt(output.CollectedFeesNaet).Equal(total) {
		failed = append(failed, "fee_distribution_not_conservative")
	}
	if !normalizeInt(params.EpochBurnCapNaet).IsZero() && normalizeInt(output.EpochBurnedNaet).GT(normalizeInt(params.EpochBurnCapNaet)) {
		failed = append(failed, "epoch_burn_cap_exceeded")
	}
	if len(output.Events) == 0 && normalizeInt(output.BurnNaet).IsPositive() {
		failed = append(failed, "burn_event_missing")
	}
	for _, event := range output.Events {
		if !event.RemovedFromSpendableSupply {
			failed = append(failed, "burn_not_removed_from_spendable_supply")
		}
	}
	if !params.GovernanceAllowsBelowFloor && normalizeInt(output.DeflationGuard.NetIssuanceAfterBurnNaet).LT(normalizeInt(params.NetIssuanceFloorNaet)) {
		failed = append(failed, "net_issuance_below_floor")
	}
	return BurnAccountingInvariantReport{Passed: len(failed) == 0, Failed: failed}
}

func ValidateBurnSlashingDistributionInvariants(output BurnIntegratedSlashingDistributionOutput, params BurnMechanicsParams) BurnAccountingInvariantReport {
	params = params.withDefaults()
	failed := make([]string, 0)
	total := normalizeInt(output.BurnNaet).
		Add(normalizeInt(output.TreasuryNaet)).
		Add(normalizeInt(output.ReporterRewardNaet)).
		Add(normalizeInt(output.ValidatorPoolNaet)).
		Add(normalizeInt(output.DeflationReserveNaet))
	if !normalizeInt(output.PenaltyNaet).Equal(total) {
		failed = append(failed, "slashing_distribution_not_conservative")
	}
	if !normalizeInt(params.EpochBurnCapNaet).IsZero() && normalizeInt(output.EpochBurnedNaet).GT(normalizeInt(params.EpochBurnCapNaet)) {
		failed = append(failed, "epoch_burn_cap_exceeded")
	}
	for _, event := range output.Events {
		if !event.RemovedFromSpendableSupply {
			failed = append(failed, "burn_not_removed_from_spendable_supply")
		}
	}
	return BurnAccountingInvariantReport{Passed: len(failed) == 0, Failed: failed}
}

func QueryBurnSupply(input BurnSupplyQueryInput) (BurnSupplyQueryOutput, error) {
	cumulative := normalizeInt(input.CumulativeBurnedNaet)
	if cumulative.IsNegative() {
		return BurnSupplyQueryOutput{}, fmt.Errorf("cumulative_burned_naet must not be negative")
	}
	recent := sdkmath.ZeroInt()
	for _, event := range input.Events {
		burned := normalizeInt(event.BurnedNaet)
		if burned.IsNegative() {
			return BurnSupplyQueryOutput{}, fmt.Errorf("event burned_naet must not be negative")
		}
		if input.RecentWindowBlocks == 0 || event.BlockHeight+input.RecentWindowBlocks >= input.CurrentBlockHeight {
			recent = recent.Add(burned)
		}
	}
	rate := sdkmath.ZeroInt()
	if input.RecentWindowBlocks > 0 {
		rate = recent.QuoRaw(int64(input.RecentWindowBlocks))
	}
	return BurnSupplyQueryOutput{
		CumulativeBurnedNaet:		cumulative,
		RecentBurnedNaet:		recent,
		RecentBurnRateNaetPerBlock:	rate,
		EventCount:			len(input.Events),
	}, nil
}

func (p BurnMechanicsParams) Validate() error {
	if normalizeInt(p.EpochBurnCapNaet).IsNegative() {
		return fmt.Errorf("epoch_burn_cap_naet must not be negative")
	}
	if normalizeInt(p.BurnFloorNaet).IsNegative() {
		return fmt.Errorf("burn_floor_naet must not be negative")
	}
	if normalizeInt(p.NetIssuanceFloorNaet).IsNegative() {
		return fmt.Errorf("net_issuance_floor_naet must not be negative")
	}
	if normalizeInt(p.DeflationReserveCapNaet).IsNegative() {
		return fmt.Errorf("deflation_reserve_cap_naet must not be negative")
	}
	if normalizeInt(p.SecurityRewardFloorNaet).IsNegative() {
		return fmt.Errorf("security_reward_floor_naet must not be negative")
	}
	if err := validateBps("fee_spike_threshold_bps", p.FeeSpikeThresholdBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	return validateBps("bonded_stake_safety_threshold_bps", p.BondedStakeSafetyThresholdBps, 0, BasisPoints)
}

func (p BurnMechanicsParams) withDefaults() BurnMechanicsParams {
	defaults := DefaultBurnMechanicsParams()
	if p.NetIssuanceFloorNaet.IsNil() {
		p.NetIssuanceFloorNaet = defaults.NetIssuanceFloorNaet
	}
	if p.DeflationReserveCapNaet.IsNil() {
		p.DeflationReserveCapNaet = defaults.DeflationReserveCapNaet
	}
	if p.SecurityRewardFloorNaet.IsNil() {
		p.SecurityRewardFloorNaet = defaults.SecurityRewardFloorNaet
	}
	if p.BurnFloorNaet.IsNil() {
		p.BurnFloorNaet = defaults.BurnFloorNaet
	}
	if p.FeeSpikeThresholdBps == 0 {
		p.FeeSpikeThresholdBps = defaults.FeeSpikeThresholdBps
	}
	if p.BondedStakeSafetyThresholdBps == 0 {
		p.BondedStakeSafetyThresholdBps = defaults.BondedStakeSafetyThresholdBps
	}
	return p
}

type deflationGuardInput struct {
	Source			string
	ProposedBurnNaet	sdkmath.Int
	ExistingEpochBurned	sdkmath.Int
	GrossMintedNaet		sdkmath.Int
	CurrentSecurityReward	sdkmath.Int
	FeeSpikeBps		int64
	BondedStakeRatioBps	int64
	Params			BurnMechanicsParams
}

func applyDeflationGuard(input deflationGuardInput) DeflationGuardTelemetry {
	params := input.Params.withDefaults()
	proposed := normalizeInt(input.ProposedBurnNaet)
	finalBurn := proposed
	toReserve := sdkmath.ZeroInt()
	toSecurity := sdkmath.ZeroInt()
	reasons := make([]string, 0)

	if input.Source == BurnSourceFees {
		securityReward := normalizeInt(input.CurrentSecurityReward)
		securityFloor := normalizeInt(params.SecurityRewardFloorNaet)
		if securityFloor.IsPositive() && securityReward.LT(securityFloor) && finalBurn.IsPositive() {
			needed := securityFloor.Sub(securityReward)
			move := minInt(finalBurn, needed)
			finalBurn = finalBurn.Sub(move)
			toSecurity = toSecurity.Add(move)
			reasons = append(reasons, "security_reward_floor_priority")
		}
		if input.BondedStakeRatioBps > 0 && input.BondedStakeRatioBps < params.BondedStakeSafetyThresholdBps && finalBurn.IsPositive() {
			move := finalBurn
			finalBurn = finalBurn.Sub(move)
			toSecurity = toSecurity.Add(move)
			reasons = append(reasons, "bonded_stake_safety_priority")
		}
	}

	if !normalizeInt(params.EpochBurnCapNaet).IsZero() {
		remainingCap := normalizeInt(params.EpochBurnCapNaet).Sub(normalizeInt(input.ExistingEpochBurned))
		if remainingCap.IsNegative() {
			remainingCap = sdkmath.ZeroInt()
		}
		if finalBurn.GT(remainingCap) {
			excess := finalBurn.Sub(remainingCap)
			finalBurn = remainingCap
			toReserve = toReserve.Add(excess)
			reasons = append(reasons, "burn_cap_applied")
		}
	}

	netFloor := normalizeInt(params.NetIssuanceFloorNaet)
	if !params.GovernanceAllowsBelowFloor && netFloor.IsPositive() {
		netAfterBurn := normalizeInt(input.GrossMintedNaet).Sub(normalizeInt(input.ExistingEpochBurned)).Sub(finalBurn)
		if netAfterBurn.LT(netFloor) {
			reduction := minInt(finalBurn, netFloor.Sub(netAfterBurn))
			finalBurn = finalBurn.Sub(reduction)
			toReserve = toReserve.Add(reduction)
			reasons = append(reasons, "net_issuance_floor_guard")
		}
	}

	if input.FeeSpikeBps > params.FeeSpikeThresholdBps && finalBurn.IsPositive() {
		toReserve = toReserve.Add(finalBurn)
		finalBurn = sdkmath.ZeroInt()
		reasons = append(reasons, "fee_spike_diverted_to_reserve")
	}

	burnFloor := normalizeInt(params.BurnFloorNaet)
	if burnFloor.IsPositive() && finalBurn.LT(burnFloor) && proposed.GTE(burnFloor) {
		reasons = append(reasons, "burn_floor_constrained_by_guard")
	}

	reduced := proposed.Sub(finalBurn)
	if reduced.IsNegative() {
		reduced = sdkmath.ZeroInt()
	}
	netAfterBurn := normalizeInt(input.GrossMintedNaet).Sub(normalizeInt(input.ExistingEpochBurned)).Sub(finalBurn)
	return DeflationGuardTelemetry{
		Active:				len(reasons) > 0,
		Reasons:			reasons,
		ProposedBurnNaet:		proposed,
		FinalBurnNaet:			finalBurn,
		ReducedBurnNaet:		reduced,
		DivertedToReserveNaet:		toReserve,
		DivertedToSecurityRewardsNaet:	toSecurity,
		NetIssuanceFloorNaet:		netFloor,
		NetIssuanceAfterBurnNaet:	netAfterBurn,
		EpochBurnCapNaet:		normalizeInt(params.EpochBurnCapNaet),
		SecurityRewardFloorNaet:	normalizeInt(params.SecurityRewardFloorNaet),
	}
}

func validateBurnFeeInput(input BurnIntegratedFeeDistributionInput) error {
	if normalizeInt(input.CollectedFeesNaet).IsNegative() {
		return fmt.Errorf("collected_fees_naet must not be negative")
	}
	if normalizeInt(input.ExistingEpochBurnedNaet).IsNegative() {
		return fmt.Errorf("existing_epoch_burned_naet must not be negative")
	}
	if normalizeInt(input.GrossMintedNaet).IsNegative() {
		return fmt.Errorf("gross_minted_naet must not be negative")
	}
	if normalizeInt(input.CumulativeBurnedNaet).IsNegative() {
		return fmt.Errorf("cumulative_burned_naet must not be negative")
	}
	if err := validateBps("burn_ratio_bps", input.BurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("community_pool_ratio_bps", input.CommunityPoolRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("state_maintenance_reserve_bps", input.StateMaintenanceReserveBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("security_reserve_ratio_bps", input.SecurityReserveRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("bonded_stake_ratio_bps", input.BondedStakeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if input.FeeSpikeBps < 0 {
		return fmt.Errorf("fee_spike_bps must not be negative")
	}
	if input.BurnRatioBps+input.CommunityPoolRatioBps+input.StateMaintenanceReserveBps+input.SecurityReserveRatioBps > BasisPoints {
		return fmt.Errorf("fee distribution ratios exceed 100%%")
	}
	return nil
}

func burnAccountingEvents(epochID, blockHeight uint64, source string, burned, cumulative, epochBurned sdkmath.Int) []BurnAccountingEvent {
	if !normalizeInt(burned).IsPositive() {
		return nil
	}
	return []BurnAccountingEvent{{
		Type:				BurnEventFundsRemoved,
		EpochID:			epochID,
		BlockHeight:			blockHeight,
		Source:				source,
		BurnedNaet:			normalizeInt(burned),
		CumulativeBurnedNaet:		normalizeInt(cumulative),
		EpochBurnedNaet:		normalizeInt(epochBurned),
		RemovedFromSpendableSupply:	true,
	}}
}

func minInt(a, b sdkmath.Int) sdkmath.Int {
	a = normalizeInt(a)
	b = normalizeInt(b)
	if a.LT(b) {
		return a
	}
	return b
}
