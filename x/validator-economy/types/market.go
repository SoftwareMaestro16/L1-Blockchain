package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

type ValidatorMarketState struct {
	Params            postypes.Params
	Candidates        []postypes.Candidate
	Delegations       []DelegationRecord
	ScoreRecords      []ValidatorScoreRecord
	SlashHistory      []ValidatorSlashHistoryRecord
	CommissionHistory []ValidatorCommissionRecord
}

type ValidatorSlashHistoryRecord struct {
	EpochID              uint64
	Height               int64
	Validator            string
	Misbehavior          string
	SlashFractionBps     uint32
	SelfBondSlashedNaet  sdkmath.Int
	DelegatorSlashedNaet sdkmath.Int
	TotalSlashedNaet     sdkmath.Int
}

type ValidatorCommissionRecord struct {
	EpochID       uint64
	Height        int64
	Validator     string
	CommissionBps uint32
}

type ValidatorRisk struct {
	Validator              string
	SlashEventCount        uint32
	TotalSlashedNaet       sdkmath.Int
	LatestReliabilityBps   uint32
	RiskScoreBps           uint32
	DelegatorRiskInherited bool
}

type ValidatorEffectiveYield struct {
	Validator              string
	RawStakeNaet           sdkmath.Int
	RewardWeightNaet       sdkmath.Int
	GrossYieldBps          uint32
	NetYieldBps            uint32
	CommissionBps          uint32
	SaturationDampeningBps uint32
}

type DelegationRiskExposure struct {
	Delegator              string
	Validator              string
	Amount                 sdkmath.Int
	RiskAppetite           string
	AdvisoryRiskProfile    bool
	SlashEventsInherited   []ValidatorSlashHistoryRecord
	ProjectedSlashNaet     sdkmath.Int
	FirstLossProtectedNaet sdkmath.Int
	HistoricalSlashNaet    sdkmath.Int
}

type SlashPropagationInput struct {
	Validator         string
	SelfBondNaet      sdkmath.Int
	Delegations       []DelegationRecord
	SlashFractionBps  uint32
	SelfBondFirstLoss bool
	EvidenceHeight    int64
	Misbehavior       string
	EpochID           uint64
}

type SlashPropagationResult struct {
	Validator             string
	SelfBondSlashedNaet   sdkmath.Int
	DelegatorSlashes      []DelegatorSlashExposure
	TotalDelegatorSlashed sdkmath.Int
	TotalSlashedNaet      sdkmath.Int
}

type DelegatorSlashExposure struct {
	Delegator   string
	Validator   string
	Amount      sdkmath.Int
	SlashedNaet sdkmath.Int
	RiskTranche string
}

func NewValidatorMarketState(params postypes.Params, candidates []postypes.Candidate, delegations []DelegationRecord, scoreRecords []ValidatorScoreRecord, slashHistory []ValidatorSlashHistoryRecord, commissionHistory []ValidatorCommissionRecord) (ValidatorMarketState, error) {
	if err := params.Validate(); err != nil {
		return ValidatorMarketState{}, err
	}
	candidateCopies := make([]postypes.Candidate, len(candidates))
	for i, candidate := range candidates {
		if err := candidate.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		candidateCopies[i] = candidate
	}
	delegationState, err := NewDelegationCapitalState(params, delegations)
	if err != nil {
		return ValidatorMarketState{}, err
	}
	scoreState, err := NewScoreComponentState(scoreRecords)
	if err != nil {
		return ValidatorMarketState{}, err
	}
	slashCopies := make([]ValidatorSlashHistoryRecord, len(slashHistory))
	for i, record := range slashHistory {
		record.Validator = strings.TrimSpace(record.Validator)
		if err := record.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		slashCopies[i] = record
	}
	sortSlashHistory(slashCopies)
	commissionCopies := make([]ValidatorCommissionRecord, len(commissionHistory))
	for i, record := range commissionHistory {
		record.Validator = strings.TrimSpace(record.Validator)
		if err := record.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		commissionCopies[i] = record
	}
	sortCommissionHistory(commissionCopies)
	return ValidatorMarketState{
		Params:            params,
		Candidates:        candidateCopies,
		Delegations:       delegationState.Records,
		ScoreRecords:      scoreState.Records,
		SlashHistory:      slashCopies,
		CommissionHistory: commissionCopies,
	}, nil
}

func (r ValidatorSlashHistoryRecord) Validate(params postypes.Params) error {
	if r.EpochID == 0 {
		return errors.New("slash history epoch id is required")
	}
	if r.Height < 0 {
		return errors.New("slash history height cannot be negative")
	}
	if err := validateEconomyToken("slash history validator", r.Validator); err != nil {
		return err
	}
	if !postypes.IsSlashableMisbehavior(r.Misbehavior) {
		return fmt.Errorf("unsupported slash history misbehavior %q", r.Misbehavior)
	}
	if r.SlashFractionBps == 0 || r.SlashFractionBps > postypes.BasisPoints {
		return fmt.Errorf("slash history fraction must be within 1..%d bps", postypes.BasisPoints)
	}
	if r.SelfBondSlashedNaet.IsNegative() || r.DelegatorSlashedNaet.IsNegative() || r.TotalSlashedNaet.IsNegative() {
		return errors.New("slash history amounts cannot be negative")
	}
	if !r.SelfBondSlashedNaet.Add(r.DelegatorSlashedNaet).Equal(r.TotalSlashedNaet) {
		return errors.New("slash history total must equal self bond plus delegator slashes")
	}
	return params.Validate()
}

func (r ValidatorCommissionRecord) Validate(params postypes.Params) error {
	if r.EpochID == 0 {
		return errors.New("commission history epoch id is required")
	}
	if r.Height < 0 {
		return errors.New("commission history height cannot be negative")
	}
	if err := validateEconomyToken("commission history validator", r.Validator); err != nil {
		return err
	}
	if r.CommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("commission history bps must be <= %d", params.MaxCommissionBps)
	}
	return nil
}

func PropagateSlash(input SlashPropagationInput) (SlashPropagationResult, error) {
	if err := validateEconomyToken("slash validator", input.Validator); err != nil {
		return SlashPropagationResult{}, err
	}
	if input.SelfBondNaet.IsNegative() {
		return SlashPropagationResult{}, errors.New("self bond cannot be negative")
	}
	if input.SlashFractionBps == 0 || input.SlashFractionBps > postypes.BasisPoints {
		return SlashPropagationResult{}, fmt.Errorf("slash fraction must be within 1..%d bps", postypes.BasisPoints)
	}
	validatorDelegations := filterDelegationsForValidator(input.Delegations, input.Validator)
	totalDelegated := sdkmath.ZeroInt()
	for _, record := range validatorDelegations {
		totalDelegated = totalDelegated.Add(record.Amount)
	}
	totalStake := input.SelfBondNaet.Add(totalDelegated)
	targetSlash := mulIntBps(totalStake, input.SlashFractionBps)
	result := SlashPropagationResult{
		Validator:             strings.TrimSpace(input.Validator),
		SelfBondSlashedNaet:   mulIntBps(input.SelfBondNaet, input.SlashFractionBps),
		DelegatorSlashes:      make([]DelegatorSlashExposure, 0, len(validatorDelegations)),
		TotalDelegatorSlashed: sdkmath.ZeroInt(),
		TotalSlashedNaet:      sdkmath.ZeroInt(),
	}
	if input.SelfBondFirstLoss && targetSlash.GT(result.SelfBondSlashedNaet) {
		if input.SelfBondNaet.GTE(targetSlash) {
			result.SelfBondSlashedNaet = targetSlash
		} else {
			result.SelfBondSlashedNaet = input.SelfBondNaet
		}
	}
	remaining := targetSlash.Sub(result.SelfBondSlashedNaet)
	for _, record := range validatorDelegations {
		slashed := sdkmath.ZeroInt()
		if remaining.IsPositive() && totalDelegated.IsPositive() {
			slashed = shareByStake(remaining, record.Amount, totalDelegated)
		}
		result.DelegatorSlashes = append(result.DelegatorSlashes, DelegatorSlashExposure{
			Delegator:   record.Delegator,
			Validator:   record.Validator,
			Amount:      record.Amount,
			SlashedNaet: slashed,
			RiskTranche: record.RiskTrancheOptional,
		})
		result.TotalDelegatorSlashed = result.TotalDelegatorSlashed.Add(slashed)
	}
	result.TotalSlashedNaet = result.SelfBondSlashedNaet.Add(result.TotalDelegatorSlashed)
	return result, nil
}

func (s ValidatorMarketState) QueryValidatorRisk(validator string) (ValidatorRisk, bool) {
	validator = strings.TrimSpace(validator)
	history := s.QueryValidatorSlashHistory(validator)
	if len(history) == 0 && latestScoreRecord(s.ScoreRecords, validator).ValidatorAddress == "" {
		return ValidatorRisk{}, false
	}
	totalSlashed := sdkmath.ZeroInt()
	for _, record := range history {
		totalSlashed = totalSlashed.Add(record.TotalSlashedNaet)
	}
	latestScore := latestScoreRecord(s.ScoreRecords, validator)
	reliability := latestScore.ReliabilityIndex
	if reliability == 0 {
		reliability = postypes.BasisPoints
	}
	risk := ValidatorRisk{
		Validator:              validator,
		SlashEventCount:        uint32(len(history)),
		TotalSlashedNaet:       totalSlashed,
		LatestReliabilityBps:   reliability,
		DelegatorRiskInherited: len(history) > 0,
	}
	risk.RiskScoreBps = minBps(uint64(len(history))*1_000 + uint64(postypes.BasisPoints-reliability))
	return risk, true
}

func (s ValidatorMarketState) QueryValidatorEffectiveYield(validator string, annualRewardsNaet sdkmath.Int) (ValidatorEffectiveYield, bool, error) {
	if annualRewardsNaet.IsNegative() {
		return ValidatorEffectiveYield{}, false, errors.New("annual rewards cannot be negative")
	}
	candidate, found := s.findCandidate(validator)
	if !found {
		return ValidatorEffectiveYield{}, false, nil
	}
	preview, err := postypes.PreviewStakeSaturation(s.Params, candidate)
	if err != nil {
		return ValidatorEffectiveYield{}, false, err
	}
	commission := latestCommissionBps(s.CommissionHistory, validator, candidate.CommissionBps)
	gross := shareBps(annualRewardsNaet, preview.BondedStakeNaet)
	net := uint32((uint64(gross) * uint64(postypes.BasisPoints-commission)) / uint64(postypes.BasisPoints))
	return ValidatorEffectiveYield{
		Validator:              strings.TrimSpace(validator),
		RawStakeNaet:           preview.BondedStakeNaet,
		RewardWeightNaet:       preview.RewardWeightNaet,
		GrossYieldBps:          gross,
		NetYieldBps:            net,
		CommissionBps:          commission,
		SaturationDampeningBps: shareBps(preview.RewardWeightNaet, preview.BondedStakeNaet),
	}, true, nil
}

func (s ValidatorMarketState) QueryValidatorSaturation(validator string) (postypes.StakeSaturationPreview, bool, error) {
	candidate, found := s.findCandidate(validator)
	if !found {
		return postypes.StakeSaturationPreview{}, false, nil
	}
	preview, err := postypes.PreviewStakeSaturation(s.Params, candidate)
	if err != nil {
		return postypes.StakeSaturationPreview{}, false, err
	}
	return preview, true, nil
}

func (s ValidatorMarketState) QueryDelegationRiskExposure(delegator string, validator string, slashFractionBps uint32, selfBondFirstLoss bool) (DelegationRiskExposure, bool, error) {
	delegator = strings.TrimSpace(delegator)
	validator = strings.TrimSpace(validator)
	record, found := s.findDelegation(delegator, validator)
	if !found {
		return DelegationRiskExposure{}, false, nil
	}
	candidate, candidateFound := s.findCandidate(validator)
	if !candidateFound {
		return DelegationRiskExposure{}, false, fmt.Errorf("validator %q is not in market candidates", validator)
	}
	propagation, err := PropagateSlash(SlashPropagationInput{
		Validator:         validator,
		SelfBondNaet:      candidate.SelfStakeNaet,
		Delegations:       s.Delegations,
		SlashFractionBps:  slashFractionBps,
		SelfBondFirstLoss: selfBondFirstLoss,
	})
	if err != nil {
		return DelegationRiskExposure{}, false, err
	}
	projected := sdkmath.ZeroInt()
	firstLossProtected := sdkmath.ZeroInt()
	proportionalWithoutFirstLoss := mulIntBps(record.Amount, slashFractionBps)
	for _, slash := range propagation.DelegatorSlashes {
		if slash.Delegator == delegator && slash.Validator == validator {
			projected = slash.SlashedNaet
			break
		}
	}
	if proportionalWithoutFirstLoss.GT(projected) {
		firstLossProtected = proportionalWithoutFirstLoss.Sub(projected)
	}
	history := s.QueryValidatorSlashHistory(validator)
	historical := sdkmath.ZeroInt()
	for _, event := range history {
		historical = historical.Add(shareByStake(event.DelegatorSlashedNaet, record.Amount, s.totalDelegatedAtValidator(validator)))
	}
	return DelegationRiskExposure{
		Delegator:              delegator,
		Validator:              validator,
		Amount:                 record.Amount,
		RiskAppetite:           record.RiskAppetite,
		AdvisoryRiskProfile:    true,
		SlashEventsInherited:   history,
		ProjectedSlashNaet:     projected,
		FirstLossProtectedNaet: firstLossProtected,
		HistoricalSlashNaet:    historical,
	}, true, nil
}

func (s ValidatorMarketState) QueryDelegationActivationEpoch(delegator string, validator string) (uint64, bool) {
	record, found := s.findDelegation(strings.TrimSpace(delegator), strings.TrimSpace(validator))
	if !found {
		return 0, false
	}
	return record.ActivationEpoch, true
}

func (s ValidatorMarketState) QueryValidatorCommissionHistory(validator string) []ValidatorCommissionRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorCommissionRecord, 0)
	for _, record := range s.CommissionHistory {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortCommissionHistory(out)
	return out
}

func (s ValidatorMarketState) QueryValidatorSlashHistory(validator string) []ValidatorSlashHistoryRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorSlashHistoryRecord, 0)
	for _, record := range s.SlashHistory {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortSlashHistory(out)
	return out
}

func (s ValidatorMarketState) QueryValidatorPerformanceHistory(validator string) []ValidatorScoreRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorScoreRecord, 0)
	for _, record := range s.ScoreRecords {
		if record.ValidatorAddress == validator {
			out = append(out, record)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].EpochID < out[j].EpochID
	})
	return out
}

func (s ValidatorMarketState) findCandidate(validator string) (postypes.Candidate, bool) {
	validator = strings.TrimSpace(validator)
	for _, candidate := range s.Candidates {
		if strings.TrimSpace(candidate.ValidatorID) == validator {
			return candidate, true
		}
	}
	return postypes.Candidate{}, false
}

func (s ValidatorMarketState) findDelegation(delegator string, validator string) (DelegationRecord, bool) {
	for _, record := range s.Delegations {
		if record.Delegator == delegator && record.Validator == validator {
			return record, true
		}
	}
	return DelegationRecord{}, false
}

func (s ValidatorMarketState) totalDelegatedAtValidator(validator string) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, record := range s.Delegations {
		if record.Validator == validator {
			total = total.Add(record.Amount)
		}
	}
	return total
}

func filterDelegationsForValidator(records []DelegationRecord, validator string) []DelegationRecord {
	validator = strings.TrimSpace(validator)
	out := make([]DelegationRecord, 0)
	for _, record := range records {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortDelegationRecords(out)
	return out
}

func latestScoreRecord(records []ValidatorScoreRecord, validator string) ValidatorScoreRecord {
	validator = strings.TrimSpace(validator)
	var latest ValidatorScoreRecord
	for _, record := range records {
		if record.ValidatorAddress == validator && record.EpochID >= latest.EpochID {
			latest = record
		}
	}
	return latest
}

func latestCommissionBps(records []ValidatorCommissionRecord, validator string, fallback uint32) uint32 {
	validator = strings.TrimSpace(validator)
	latestEpoch := uint64(0)
	commission := fallback
	for _, record := range records {
		if record.Validator == validator && record.EpochID >= latestEpoch {
			latestEpoch = record.EpochID
			commission = record.CommissionBps
		}
	}
	return commission
}

func sortSlashHistory(records []ValidatorSlashHistoryRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID < records[j].EpochID
		}
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		return records[i].Validator < records[j].Validator
	})
}

func sortCommissionHistory(records []ValidatorCommissionRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID < records[j].EpochID
		}
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		return records[i].Validator < records[j].Validator
	})
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	return value.MulRaw(int64(bps)).QuoRaw(int64(postypes.BasisPoints))
}

func shareByStake(amount sdkmath.Int, stake sdkmath.Int, totalStake sdkmath.Int) sdkmath.Int {
	if !amount.IsPositive() || !stake.IsPositive() || !totalStake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	return amount.Mul(stake).Quo(totalStake)
}

func minBps(value uint64) uint32 {
	if value > uint64(postypes.BasisPoints) {
		return postypes.BasisPoints
	}
	return uint32(value)
}
