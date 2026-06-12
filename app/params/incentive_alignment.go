package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	ParticipantClassValidator	= "validator"
	ParticipantClassDelegator	= "delegator"
	ParticipantClassUser		= "user"
	ParticipantClassProtocolReserve	= "protocol_reserve"

	IncentiveResourceConsensusSecurity	= "consensus_security"
	IncentiveResourceExecution		= "execution"
	IncentiveResourceStorage		= "storage"
	IncentiveResourceSpamRisk		= "spam_risk"
	IncentiveResourceSlashingRisk		= "slashing_risk"
	IncentiveResourceProtocolReserve	= "protocol_reserve"

	ParameterChangeInflation	= "inflation"
	ParameterChangeBurn		= "burn"
	ParameterChangeFeeSplit		= "fee_split"
	ParameterChangeStorageRent	= "storage_rent"
	ParameterChangeConcentration	= "concentration"
)

type ParticipantIncentiveEntry struct {
	ParticipantClass	string
	ResourceOrRisk		string
	RewardsNaet		sdkmath.Int
	PenaltiesNaet		sdkmath.Int
	FeesPaidNaet		sdkmath.Int
	ReserveContributionNaet	sdkmath.Int
	NetPositionNaet		sdkmath.Int
	Explanation		string
}

type ParticipantIncentiveMapInput struct {
	ValidatorRewardsNaet	sdkmath.Int
	DelegatorRewardsNaet	sdkmath.Int
	UserFeesPaidNaet	sdkmath.Int
	ExecutionFeesNaet	sdkmath.Int
	StorageFeesNaet		sdkmath.Int
	SpamSurchargeNaet	sdkmath.Int
	ValidatorSlashedNaet	sdkmath.Int
	DelegatorSlashedNaet	sdkmath.Int
	ReserveFundingNaet	sdkmath.Int
	BurnedNaet		sdkmath.Int
}

type ParticipantIncentiveMapReport struct {
	Entries				[]ParticipantIncentiveEntry
	TotalRewardsNaet		sdkmath.Int
	TotalPenaltiesNaet		sdkmath.Int
	TotalFeesPaidNaet		sdkmath.Int
	TotalReserveContributionNaet	sdkmath.Int
	Passed				bool
	Failed				[]string
}

type EpochEconomicReportInput struct {
	EpochID				uint64
	StartingSupplyNaet		sdkmath.Int
	EndingSupplyNaet		sdkmath.Int
	GrossIssuedNaet			sdkmath.Int
	BurnedNaet			sdkmath.Int
	FeesCollectedNaet		sdkmath.Int
	ValidatorRewardsNaet		sdkmath.Int
	DelegatorRewardsNaet		sdkmath.Int
	SlashedNaet			sdkmath.Int
	ReserveInflowNaet		sdkmath.Int
	ReserveOutflowNaet		sdkmath.Int
	StateGrowthBytes		int64
	ValidatorConcentrationBps	int64
	ParticipantInput		ParticipantIncentiveMapInput
}

type EpochEconomicReport struct {
	EpochID				uint64
	StartingSupplyNaet		sdkmath.Int
	EndingSupplyNaet		sdkmath.Int
	ExpectedEndingSupplyNaet	sdkmath.Int
	GrossIssuedNaet			sdkmath.Int
	BurnedNaet			sdkmath.Int
	NetIssuanceNaet			sdkmath.Int
	FeesCollectedNaet		sdkmath.Int
	RewardsDistributedNaet		sdkmath.Int
	SlashedFundsNaet		sdkmath.Int
	ReserveInflowNaet		sdkmath.Int
	ReserveOutflowNaet		sdkmath.Int
	NetReserveChangeNaet		sdkmath.Int
	StateGrowthBytes		int64
	ValidatorConcentrationBps	int64
	ParticipantIncentives		ParticipantIncentiveMapReport
	GovernanceSummary		string
	Reconciled			bool
	Failed				[]string
}

type GovernanceParameterImpactInput struct {
	ParameterName				string
	CurrentValueBps				int64
	ProposedValueBps			int64
	CurrentEpochReport			EpochEconomicReportInput
	ProjectedEpochs				uint32
	RequirePreUpgradeSimulation		bool
	ConsensusRewardAccountingPreserved	bool
}

type GovernanceDashboardRow struct {
	Metric		string
	CurrentNaet	sdkmath.Int
	ProjectedNaet	sdkmath.Int
	DeltaNaet	sdkmath.Int
	CurrentBps	int64
	ProjectedBps	int64
}

type GovernanceParameterImpactReport struct {
	ParameterName				string
	CurrentValueBps				int64
	ProposedValueBps			int64
	DeltaBps				int64
	ProjectedReports			[]EpochEconomicReport
	ProjectedSupplyDeltaNaet		sdkmath.Int
	ProjectedReserveDeltaNaet		sdkmath.Int
	ProjectedRewardDeltaNaet		sdkmath.Int
	PreUpgradeSimulationIncluded		bool
	ConsensusRewardAccountingPreserved	bool
	ActivationAllowed			bool
	DashboardRows				[]GovernanceDashboardRow
	Failed					[]string
}

func BuildParticipantIncentiveMap(input ParticipantIncentiveMapInput) (ParticipantIncentiveMapReport, error) {
	if err := validateParticipantIncentiveInput(input); err != nil {
		return ParticipantIncentiveMapReport{}, err
	}

	validatorRewards := normalizeInt(input.ValidatorRewardsNaet)
	delegatorRewards := normalizeInt(input.DelegatorRewardsNaet)
	validatorSlashed := normalizeInt(input.ValidatorSlashedNaet)
	delegatorSlashed := normalizeInt(input.DelegatorSlashedNaet)
	userFees := normalizeInt(input.UserFeesPaidNaet).
		Add(normalizeInt(input.ExecutionFeesNaet)).
		Add(normalizeInt(input.StorageFeesNaet)).
		Add(normalizeInt(input.SpamSurchargeNaet))
	reserveFunding := normalizeInt(input.ReserveFundingNaet)

	entries := []ParticipantIncentiveEntry{
		buildParticipantEntry(
			ParticipantClassValidator,
			IncentiveResourceConsensusSecurity+","+IncentiveResourceSlashingRisk,
			validatorRewards,
			validatorSlashed,
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			"validator rewards pay consensus security and validator slash penalties price operator risk",
		),
		buildParticipantEntry(
			ParticipantClassDelegator,
			IncentiveResourceConsensusSecurity+","+IncentiveResourceSlashingRisk,
			delegatorRewards,
			delegatorSlashed,
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			"delegator rewards and slash exposure make validator choice risk-adjusted",
		),
		buildParticipantEntry(
			ParticipantClassUser,
			IncentiveResourceExecution+","+IncentiveResourceStorage+","+IncentiveResourceSpamRisk,
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			userFees,
			sdkmath.ZeroInt(),
			"users pay deterministic execution, storage, and anti-spam costs for resources consumed",
		),
		buildParticipantEntry(
			ParticipantClassProtocolReserve,
			IncentiveResourceProtocolReserve,
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			reserveFunding,
			"protocol reserves receive configured maintenance and security funding",
		),
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ParticipantClass < entries[j].ParticipantClass
	})

	totalRewards := validatorRewards.Add(delegatorRewards)
	totalPenalties := validatorSlashed.Add(delegatorSlashed)
	totalReserve := reserveFunding
	failed := make([]string, 0)
	if normalizeInt(input.BurnedNaet).IsNegative() {
		failed = append(failed, "burned_naet_negative")
	}

	return ParticipantIncentiveMapReport{
		Entries:			entries,
		TotalRewardsNaet:		totalRewards,
		TotalPenaltiesNaet:		totalPenalties,
		TotalFeesPaidNaet:		userFees,
		TotalReserveContributionNaet:	totalReserve,
		Passed:				len(failed) == 0,
		Failed:				failed,
	}, nil
}

func GenerateEpochEconomicReport(input EpochEconomicReportInput) (EpochEconomicReport, error) {
	if err := validateEpochEconomicReportInput(input); err != nil {
		return EpochEconomicReport{}, err
	}

	participantInput := input.ParticipantInput
	if participantInput == (ParticipantIncentiveMapInput{}) {
		participantInput = ParticipantIncentiveMapInput{
			ValidatorRewardsNaet:	input.ValidatorRewardsNaet,
			DelegatorRewardsNaet:	input.DelegatorRewardsNaet,
			UserFeesPaidNaet:	input.FeesCollectedNaet,
			ValidatorSlashedNaet:	input.SlashedNaet,
			ReserveFundingNaet:	normalizeInt(input.ReserveInflowNaet).Sub(normalizeInt(input.ReserveOutflowNaet)),
			BurnedNaet:		input.BurnedNaet,
		}
	}
	participantReport, err := BuildParticipantIncentiveMap(participantInput)
	if err != nil {
		return EpochEconomicReport{}, err
	}

	starting := normalizeInt(input.StartingSupplyNaet)
	grossIssued := normalizeInt(input.GrossIssuedNaet)
	burned := normalizeInt(input.BurnedNaet)
	expectedEnding := starting.Add(grossIssued).Sub(burned)
	ending := normalizeInt(input.EndingSupplyNaet)
	if ending.IsZero() {
		ending = expectedEnding
	}

	failed := make([]string, 0)
	if !ending.Equal(expectedEnding) {
		failed = append(failed, "supply_accounting_mismatch")
	}
	if !participantReport.Passed {
		failed = append(failed, participantReport.Failed...)
	}

	rewards := normalizeInt(input.ValidatorRewardsNaet).Add(normalizeInt(input.DelegatorRewardsNaet))
	netReserve := normalizeInt(input.ReserveInflowNaet).Sub(normalizeInt(input.ReserveOutflowNaet))
	report := EpochEconomicReport{
		EpochID:			input.EpochID,
		StartingSupplyNaet:		starting,
		EndingSupplyNaet:		ending,
		ExpectedEndingSupplyNaet:	expectedEnding,
		GrossIssuedNaet:		grossIssued,
		BurnedNaet:			burned,
		NetIssuanceNaet:		grossIssued.Sub(burned),
		FeesCollectedNaet:		normalizeInt(input.FeesCollectedNaet),
		RewardsDistributedNaet:		rewards,
		SlashedFundsNaet:		normalizeInt(input.SlashedNaet),
		ReserveInflowNaet:		normalizeInt(input.ReserveInflowNaet),
		ReserveOutflowNaet:		normalizeInt(input.ReserveOutflowNaet),
		NetReserveChangeNaet:		netReserve,
		StateGrowthBytes:		input.StateGrowthBytes,
		ValidatorConcentrationBps:	input.ValidatorConcentrationBps,
		ParticipantIncentives:		participantReport,
		Reconciled:			len(failed) == 0,
		Failed:				failed,
	}
	report.GovernanceSummary = fmt.Sprintf(
		"epoch=%d issuance=%s burn=%s net=%s fees=%s rewards=%s slashed=%s reserve_delta=%s state_growth_bytes=%d validator_concentration_bps=%d",
		report.EpochID,
		report.GrossIssuedNaet.String(),
		report.BurnedNaet.String(),
		report.NetIssuanceNaet.String(),
		report.FeesCollectedNaet.String(),
		report.RewardsDistributedNaet.String(),
		report.SlashedFundsNaet.String(),
		report.NetReserveChangeNaet.String(),
		report.StateGrowthBytes,
		report.ValidatorConcentrationBps,
	)
	return report, nil
}

func GenerateGovernanceParameterImpactReport(input GovernanceParameterImpactInput) (GovernanceParameterImpactReport, error) {
	if err := validateGovernanceParameterImpactInput(input); err != nil {
		return GovernanceParameterImpactReport{}, err
	}

	failed := make([]string, 0)
	simulationIncluded := input.ProjectedEpochs > 0
	if input.RequirePreUpgradeSimulation && !simulationIncluded {
		failed = append(failed, "pre_upgrade_simulation_required")
	}
	if !input.ConsensusRewardAccountingPreserved {
		failed = append(failed, "consensus_reward_accounting_not_preserved")
	}

	current, err := GenerateEpochEconomicReport(input.CurrentEpochReport)
	if err != nil {
		return GovernanceParameterImpactReport{}, err
	}

	projected := make([]EpochEconomicReport, 0, input.ProjectedEpochs)
	nextInput := input.CurrentEpochReport
	startingSupply := current.StartingSupplyNaet
	rewardDelta := sdkmath.ZeroInt()
	reserveDelta := sdkmath.ZeroInt()
	for epoch := uint32(0); epoch < input.ProjectedEpochs; epoch++ {
		nextInput.EpochID = input.CurrentEpochReport.EpochID + uint64(epoch) + 1
		nextInput.StartingSupplyNaet = startingSupply
		nextInput.EndingSupplyNaet = sdkmath.ZeroInt()
		applyGovernanceProjection(&nextInput, input.ParameterName, input.ProposedValueBps)
		nextInput.ParticipantInput = ParticipantIncentiveMapInput{
			ValidatorRewardsNaet:	nextInput.ValidatorRewardsNaet,
			DelegatorRewardsNaet:	nextInput.DelegatorRewardsNaet,
			UserFeesPaidNaet:	nextInput.FeesCollectedNaet,
			ValidatorSlashedNaet:	nextInput.SlashedNaet,
			ReserveFundingNaet:	normalizeInt(nextInput.ReserveInflowNaet).Sub(normalizeInt(nextInput.ReserveOutflowNaet)),
			BurnedNaet:		nextInput.BurnedNaet,
		}

		report, err := GenerateEpochEconomicReport(nextInput)
		if err != nil {
			return GovernanceParameterImpactReport{}, err
		}
		projected = append(projected, report)
		startingSupply = report.EndingSupplyNaet
		rewardDelta = rewardDelta.Add(report.RewardsDistributedNaet.Sub(current.RewardsDistributedNaet))
		reserveDelta = reserveDelta.Add(report.NetReserveChangeNaet.Sub(current.NetReserveChangeNaet))
		if !report.Reconciled {
			failed = append(failed, "projected_epoch_not_reconciled")
		}
	}

	supplyDelta := sdkmath.ZeroInt()
	if len(projected) > 0 {
		supplyDelta = projected[len(projected)-1].EndingSupplyNaet.Sub(current.StartingSupplyNaet)
	}
	rows := governanceDashboardRows(current, projected, input.CurrentValueBps, input.ProposedValueBps)

	return GovernanceParameterImpactReport{
		ParameterName:				input.ParameterName,
		CurrentValueBps:			input.CurrentValueBps,
		ProposedValueBps:			input.ProposedValueBps,
		DeltaBps:				input.ProposedValueBps - input.CurrentValueBps,
		ProjectedReports:			projected,
		ProjectedSupplyDeltaNaet:		supplyDelta,
		ProjectedReserveDeltaNaet:		reserveDelta,
		ProjectedRewardDeltaNaet:		rewardDelta,
		PreUpgradeSimulationIncluded:		simulationIncluded,
		ConsensusRewardAccountingPreserved:	input.ConsensusRewardAccountingPreserved,
		ActivationAllowed:			len(failed) == 0,
		DashboardRows:				rows,
		Failed:					failed,
	}, nil
}

func buildParticipantEntry(class, resource string, rewards, penalties, feesPaid, reserveContribution sdkmath.Int, explanation string) ParticipantIncentiveEntry {
	net := normalizeInt(rewards).Add(normalizeInt(reserveContribution)).Sub(normalizeInt(penalties)).Sub(normalizeInt(feesPaid))
	return ParticipantIncentiveEntry{
		ParticipantClass:		class,
		ResourceOrRisk:			resource,
		RewardsNaet:			normalizeInt(rewards),
		PenaltiesNaet:			normalizeInt(penalties),
		FeesPaidNaet:			normalizeInt(feesPaid),
		ReserveContributionNaet:	normalizeInt(reserveContribution),
		NetPositionNaet:		net,
		Explanation:			explanation,
	}
}

func validateParticipantIncentiveInput(input ParticipantIncentiveMapInput) error {
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "validator_rewards_naet", value: input.ValidatorRewardsNaet},
		{name: "delegator_rewards_naet", value: input.DelegatorRewardsNaet},
		{name: "user_fees_paid_naet", value: input.UserFeesPaidNaet},
		{name: "execution_fees_naet", value: input.ExecutionFeesNaet},
		{name: "storage_fees_naet", value: input.StorageFeesNaet},
		{name: "spam_surcharge_naet", value: input.SpamSurchargeNaet},
		{name: "validator_slashed_naet", value: input.ValidatorSlashedNaet},
		{name: "delegator_slashed_naet", value: input.DelegatorSlashedNaet},
		{name: "reserve_funding_naet", value: input.ReserveFundingNaet},
		{name: "burned_naet", value: input.BurnedNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	return nil
}

func validateEpochEconomicReportInput(input EpochEconomicReportInput) error {
	if input.EpochID == 0 {
		return fmt.Errorf("epoch_id must be positive")
	}
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "starting_supply_naet", value: input.StartingSupplyNaet},
		{name: "ending_supply_naet", value: input.EndingSupplyNaet},
		{name: "gross_issued_naet", value: input.GrossIssuedNaet},
		{name: "burned_naet", value: input.BurnedNaet},
		{name: "fees_collected_naet", value: input.FeesCollectedNaet},
		{name: "validator_rewards_naet", value: input.ValidatorRewardsNaet},
		{name: "delegator_rewards_naet", value: input.DelegatorRewardsNaet},
		{name: "slashed_naet", value: input.SlashedNaet},
		{name: "reserve_inflow_naet", value: input.ReserveInflowNaet},
		{name: "reserve_outflow_naet", value: input.ReserveOutflowNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	if !normalizeInt(input.StartingSupplyNaet).IsPositive() {
		return fmt.Errorf("starting_supply_naet must be positive")
	}
	if input.StateGrowthBytes < 0 {
		return fmt.Errorf("state_growth_bytes must not be negative")
	}
	if err := validateBps("validator_concentration_bps", input.ValidatorConcentrationBps, 0, BasisPoints); err != nil {
		return err
	}
	return nil
}

func validateGovernanceParameterImpactInput(input GovernanceParameterImpactInput) error {
	if input.ParameterName == "" {
		return fmt.Errorf("parameter_name is required")
	}
	if err := validateBps("current_value_bps", input.CurrentValueBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("proposed_value_bps", input.ProposedValueBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	switch input.ParameterName {
	case ParameterChangeInflation, ParameterChangeBurn, ParameterChangeFeeSplit, ParameterChangeStorageRent, ParameterChangeConcentration:
		return nil
	default:
		return fmt.Errorf("unsupported parameter_name %q", input.ParameterName)
	}
}

func applyGovernanceProjection(input *EpochEconomicReportInput, parameter string, proposedBps int64) {
	switch parameter {
	case ParameterChangeInflation:
		input.GrossIssuedNaet = ApplyBps(normalizeInt(input.StartingSupplyNaet), proposedBps)
		rewards := normalizeInt(input.ValidatorRewardsNaet).Add(normalizeInt(input.DelegatorRewardsNaet))
		if rewards.IsPositive() {
			validatorShareBps := normalizeInt(input.ValidatorRewardsNaet).MulRaw(BasisPoints).Quo(rewards).Int64()
			input.ValidatorRewardsNaet = ApplyBps(input.GrossIssuedNaet, validatorShareBps)
			input.DelegatorRewardsNaet = input.GrossIssuedNaet.Sub(input.ValidatorRewardsNaet)
		}
	case ParameterChangeBurn:
		input.BurnedNaet = ApplyBps(normalizeInt(input.FeesCollectedNaet), proposedBps)
	case ParameterChangeFeeSplit:
		input.ReserveInflowNaet = ApplyBps(normalizeInt(input.FeesCollectedNaet), proposedBps)
		validatorAndDelegator := normalizeInt(input.FeesCollectedNaet).Sub(input.ReserveInflowNaet)
		if validatorAndDelegator.IsNegative() {
			validatorAndDelegator = sdkmath.ZeroInt()
		}
		input.ValidatorRewardsNaet = validatorAndDelegator.QuoRaw(2)
		input.DelegatorRewardsNaet = validatorAndDelegator.Sub(input.ValidatorRewardsNaet)
	case ParameterChangeStorageRent:
		storageFee := sdkmath.NewInt(input.StateGrowthBytes).MulRaw(proposedBps)
		input.FeesCollectedNaet = normalizeInt(input.FeesCollectedNaet).Add(storageFee)
		input.ReserveInflowNaet = normalizeInt(input.ReserveInflowNaet).Add(storageFee)
	case ParameterChangeConcentration:
		input.ValidatorConcentrationBps = clampInt64(proposedBps, 0, BasisPoints)
		if proposedBps > DefaultTopNConcentrationLimitBps {
			dampening := clampInt64(proposedBps-DefaultTopNConcentrationLimitBps, 0, MaxValidatorRewardDampeningBps)
			input.ValidatorRewardsNaet = ApplyBps(normalizeInt(input.ValidatorRewardsNaet), BasisPoints-dampening)
		}
	}
}

func governanceDashboardRows(current EpochEconomicReport, projected []EpochEconomicReport, currentBps, proposedBps int64) []GovernanceDashboardRow {
	last := current
	if len(projected) > 0 {
		last = projected[len(projected)-1]
	}
	return []GovernanceDashboardRow{
		{
			Metric:		"supply",
			CurrentNaet:	current.EndingSupplyNaet,
			ProjectedNaet:	last.EndingSupplyNaet,
			DeltaNaet:	last.EndingSupplyNaet.Sub(current.EndingSupplyNaet),
			CurrentBps:	currentBps,
			ProjectedBps:	proposedBps,
		},
		{
			Metric:		"rewards",
			CurrentNaet:	current.RewardsDistributedNaet,
			ProjectedNaet:	last.RewardsDistributedNaet,
			DeltaNaet:	last.RewardsDistributedNaet.Sub(current.RewardsDistributedNaet),
			CurrentBps:	currentBps,
			ProjectedBps:	proposedBps,
		},
		{
			Metric:		"reserve_delta",
			CurrentNaet:	current.NetReserveChangeNaet,
			ProjectedNaet:	last.NetReserveChangeNaet,
			DeltaNaet:	last.NetReserveChangeNaet.Sub(current.NetReserveChangeNaet),
			CurrentBps:	currentBps,
			ProjectedBps:	proposedBps,
		},
		{
			Metric:		"net_issuance",
			CurrentNaet:	current.NetIssuanceNaet,
			ProjectedNaet:	last.NetIssuanceNaet,
			DeltaNaet:	last.NetIssuanceNaet.Sub(current.NetIssuanceNaet),
			CurrentBps:	currentBps,
			ProjectedBps:	proposedBps,
		},
	}
}
