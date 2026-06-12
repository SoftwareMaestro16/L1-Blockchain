package types

import (
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"sort"
	"strings"
)

type Params struct {
	Authority			string		`json:"authority"`
	UptimeWindowBlocks		uint64		`json:"uptime_window_blocks"`
	JailPenaltyBps			uint32		`json:"jail_penalty_bps"`
	SlashEventPenaltyBps		uint32		`json:"slash_event_penalty_bps"`
	SelfBondFullScoreBps		uint32		`json:"self_bond_full_score_bps"`
	MinRewardMultiplierBps		uint32		`json:"min_reward_multiplier_bps"`
	ObjectiveRewardModifierEnabled	bool		`json:"objective_reward_modifier_enabled"`
	ConsensusOverrideEnabled	bool		`json:"consensus_override_enabled"`
	Weights				ScoreWeights	`json:"weights"`
}

type ScoreWeights struct {
	UptimeBps		uint32	`json:"uptime_bps"`
	MissedBlocksBps		uint32	`json:"missed_blocks_bps"`
	JailBps			uint32	`json:"jail_bps"`
	SlashHistoryBps		uint32	`json:"slash_history_bps"`
	SelfBondBps		uint32	`json:"self_bond_bps"`
	CommissionBps		uint32	`json:"commission_bps"`
	GovernanceBps		uint32	`json:"governance_bps"`
	DecentralizationBps	uint32	`json:"decentralization_bps"`
	IdentityBps		uint32	`json:"identity_bps"`
}

type SlashEvent struct {
	Height		uint64	`json:"height"`
	FractionBps	uint32	`json:"fraction_bps"`
	Reason		string	`json:"reason"`
}

type CommissionPoint struct {
	Epoch		uint64	`json:"epoch"`
	CommissionBps	uint32	`json:"commission_bps"`
}

type ValidatorMetricInput struct {
	OperatorAddress			string			`json:"operator_address"`
	SignedBlocks			uint64			`json:"signed_blocks"`
	MissedBlocks			uint64			`json:"missed_blocks"`
	UptimeWindow			uint64			`json:"uptime_window"`
	JailEvents			uint64			`json:"jail_events"`
	SlashEvents			[]SlashEvent		`json:"slash_events"`
	SelfBond			uint64			`json:"self_bond"`
	TotalBond			uint64			`json:"total_bond"`
	CommissionHistory		[]CommissionPoint	`json:"commission_history"`
	GovernanceVotes			uint64			`json:"governance_votes"`
	GovernanceProposals		uint64			`json:"governance_proposals"`
	ConcentrationBps		uint32			`json:"concentration_bps"`
	ConcentrationStatus		string			`json:"concentration_status"`
	IdentityMetadataComplete	bool			`json:"identity_metadata_complete"`
}

type ValidatorScore struct {
	OperatorAddress			string	`json:"operator_address"`
	Epoch				uint64	`json:"epoch"`
	UptimeWindow			uint64	`json:"uptime_window"`
	SignedBlocks			uint64	`json:"signed_blocks"`
	MissedBlocks			uint64	`json:"missed_blocks"`
	JailEvents			uint64	`json:"jail_events"`
	SlashEventCount			uint64	`json:"slash_event_count"`
	SelfBondBps			uint32	`json:"self_bond_bps"`
	LastCommissionBps		uint32	`json:"last_commission_bps"`
	GovernanceParticipationBps	uint32	`json:"governance_participation_bps"`
	ConcentrationBps		uint32	`json:"concentration_bps"`
	ConcentrationStatus		string	`json:"concentration_status"`
	IdentityMetadataComplete	bool	`json:"identity_metadata_complete"`
	UptimeScoreBps			uint32	`json:"uptime_score_bps"`
	MissedBlockScoreBps		uint32	`json:"missed_block_score_bps"`
	JailScoreBps			uint32	`json:"jail_score_bps"`
	SlashHistoryScoreBps		uint32	`json:"slash_history_score_bps"`
	SelfBondScoreBps		uint32	`json:"self_bond_score_bps"`
	CommissionScoreBps		uint32	`json:"commission_score_bps"`
	GovernanceScoreBps		uint32	`json:"governance_score_bps"`
	DecentralizationScoreBps	uint32	`json:"decentralization_score_bps"`
	IdentityScoreBps		uint32	`json:"identity_score_bps"`
	OverallScoreBps			uint32	`json:"overall_score_bps"`
	RewardMultiplierBps		uint32	`json:"reward_multiplier_bps"`
	InformationalOnly		bool	`json:"informational_only"`
	ConsensusOverrideAllowed	bool	`json:"consensus_override_allowed"`
}

type PublicValidatorMetrics struct {
	OperatorAddress			string	`json:"operator_address"`
	Epoch				uint64	`json:"epoch"`
	UptimeBps			uint32	`json:"uptime_bps"`
	MissedBlocks			uint64	`json:"missed_blocks"`
	JailEvents			uint64	`json:"jail_events"`
	SlashEventCount			uint64	`json:"slash_event_count"`
	SelfBondBps			uint32	`json:"self_bond_bps"`
	LastCommissionBps		uint32	`json:"last_commission_bps"`
	GovernanceParticipationBps	uint32	`json:"governance_participation_bps"`
	ConcentrationBps		uint32	`json:"concentration_bps"`
	ConcentrationStatus		string	`json:"concentration_status"`
	IdentityMetadataComplete	bool	`json:"identity_metadata_complete"`
	OverallScoreBps			uint32	`json:"overall_score_bps"`
	RewardMultiplierBps		uint32	`json:"reward_multiplier_bps"`
	InformationalOnly		bool	`json:"informational_only"`
}

type GenesisState struct {
	Params	Params			`json:"params"`
	Epoch	uint64			`json:"epoch"`
	Metrics	[]ValidatorMetricInput	`json:"metrics"`
	Scores	[]ValidatorScore	`json:"scores"`
}

type MsgUpdateValidatorScoreParams struct {
	Authority	string	`json:"authority"`
	Params		Params	`json:"params"`
}

type MsgUpdateValidatorScores struct {
	Authority	string			`json:"authority"`
	Epoch		uint64			`json:"epoch"`
	Metrics		[]ValidatorMetricInput	`json:"metrics"`
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct{ Params Params }

type QueryValidatorScoreRequest struct{ OperatorAddress string }
type QueryValidatorScoreResponse struct{ Score ValidatorScore }

type QueryPublicValidatorMetricsRequest struct{ OperatorAddress string }
type QueryPublicValidatorMetricsResponse struct{ Metrics PublicValidatorMetrics }

type QueryAllValidatorScoresRequest struct{}
type QueryAllValidatorScoresResponse struct{ Scores []ValidatorScore }

func DefaultParams(authority string) Params {
	return Params{
		Authority:			authority,
		UptimeWindowBlocks:		10_000,
		JailPenaltyBps:			1_000,
		SlashEventPenaltyBps:		500,
		SelfBondFullScoreBps:		1_000,
		MinRewardMultiplierBps:		7_000,
		ObjectiveRewardModifierEnabled:	true,
		ConsensusOverrideEnabled:	false,
		Weights: ScoreWeights{
			UptimeBps:		2_500,
			MissedBlocksBps:	1_000,
			JailBps:		1_000,
			SlashHistoryBps:	1_500,
			SelfBondBps:		1_000,
			CommissionBps:		1_000,
			GovernanceBps:		1_000,
			DecentralizationBps:	800,
			IdentityBps:		200,
		},
	}
}

func DefaultGenesisState(authority string) GenesisState {
	return GenesisState{
		Params:		DefaultParams(authority),
		Metrics:	[]ValidatorMetricInput{},
		Scores:		[]ValidatorScore{},
	}
}

func ComputeValidatorScores(params Params, epoch uint64, inputs []ValidatorMetricInput) ([]ValidatorScore, error) {
	if err := params.Validate(); err != nil {
		return nil, ErrInvalidParams.Wrap(err.Error())
	}
	canonical, err := CanonicalMetricInputs(inputs)
	if err != nil {
		return nil, err
	}
	scores := make([]ValidatorScore, 0, len(canonical))
	for _, input := range canonical {
		score := scoreValidator(params, epoch, input)
		scores = append(scores, score)
	}
	return scores, nil
}

func CanonicalMetricInputs(inputs []ValidatorMetricInput) ([]ValidatorMetricInput, error) {
	canonical := append([]ValidatorMetricInput(nil), inputs...)
	for i := range canonical {
		canonical[i].OperatorAddress = strings.TrimSpace(canonical[i].OperatorAddress)
		canonical[i].ConcentrationStatus = normalizeConcentrationStatus(canonical[i].ConcentrationStatus)
		canonical[i].SlashEvents = canonicalSlashEvents(canonical[i].SlashEvents)
		canonical[i].CommissionHistory = canonicalCommissionHistory(canonical[i].CommissionHistory)
		if err := canonical[i].Validate(); err != nil {
			return nil, ErrInvalidScore.Wrap(err.Error())
		}
	}
	sort.Slice(canonical, func(i, j int) bool {
		return canonical[i].OperatorAddress < canonical[j].OperatorAddress
	})
	for i := 1; i < len(canonical); i++ {
		if canonical[i-1].OperatorAddress == canonical[i].OperatorAddress {
			return nil, ErrInvalidScore.Wrap("duplicate validator metric input")
		}
	}
	return canonical, nil
}

func PublicMetricsFromScore(score ValidatorScore) PublicValidatorMetrics {
	return PublicValidatorMetrics{
		OperatorAddress:		score.OperatorAddress,
		Epoch:				score.Epoch,
		UptimeBps:			score.UptimeScoreBps,
		MissedBlocks:			score.MissedBlocks,
		JailEvents:			score.JailEvents,
		SlashEventCount:		score.SlashEventCount,
		SelfBondBps:			score.SelfBondBps,
		LastCommissionBps:		score.LastCommissionBps,
		GovernanceParticipationBps:	score.GovernanceParticipationBps,
		ConcentrationBps:		score.ConcentrationBps,
		ConcentrationStatus:		score.ConcentrationStatus,
		IdentityMetadataComplete:	score.IdentityMetadataComplete,
		OverallScoreBps:		score.OverallScoreBps,
		RewardMultiplierBps:		score.RewardMultiplierBps,
		InformationalOnly:		score.InformationalOnly,
	}
}

func (p Params) Validate() error {
	if strings.TrimSpace(p.Authority) == "" {
		return errors.New("authority must be non-empty")
	}
	if p.UptimeWindowBlocks == 0 {
		return errors.New("uptime window blocks must be positive")
	}
	if p.JailPenaltyBps == 0 || p.JailPenaltyBps > BasisPoints {
		return fmt.Errorf("jail penalty must be between 1 and %d bps", BasisPoints)
	}
	if p.SlashEventPenaltyBps > BasisPoints {
		return fmt.Errorf("slash event penalty cannot exceed %d bps", BasisPoints)
	}
	if p.SelfBondFullScoreBps == 0 || p.SelfBondFullScoreBps > BasisPoints {
		return fmt.Errorf("self bond full score threshold must be between 1 and %d bps", BasisPoints)
	}
	if p.MinRewardMultiplierBps > BasisPoints {
		return fmt.Errorf("minimum reward multiplier cannot exceed %d bps", BasisPoints)
	}
	if err := p.Weights.Validate(); err != nil {
		return err
	}
	return nil
}

func (w ScoreWeights) Validate() error {
	weights := []struct {
		name	string
		value	uint32
	}{
		{"uptime", w.UptimeBps},
		{"missed_blocks", w.MissedBlocksBps},
		{"jail", w.JailBps},
		{"slash_history", w.SlashHistoryBps},
		{"self_bond", w.SelfBondBps},
		{"commission", w.CommissionBps},
		{"governance", w.GovernanceBps},
		{"decentralization", w.DecentralizationBps},
		{"identity", w.IdentityBps},
	}
	var total uint64
	for _, weight := range weights {
		if weight.value > BasisPoints {
			return fmt.Errorf("score weight %s cannot exceed %d bps", weight.name, BasisPoints)
		}
		total += uint64(weight.value)
	}
	if total != uint64(BasisPoints) {
		return fmt.Errorf("score weights must total %d bps", BasisPoints)
	}
	return nil
}

func (s SlashEvent) Validate() error {
	if s.Height == 0 {
		return errors.New("slash event height must be positive")
	}
	if s.FractionBps == 0 || s.FractionBps > BasisPoints {
		return fmt.Errorf("slash event fraction must be between 1 and %d bps", BasisPoints)
	}
	if strings.TrimSpace(s.Reason) == "" {
		return errors.New("slash event reason must be non-empty")
	}
	return nil
}

func (c CommissionPoint) Validate() error {
	if c.CommissionBps > BasisPoints {
		return fmt.Errorf("commission cannot exceed %d bps", BasisPoints)
	}
	return nil
}

func (v ValidatorMetricInput) Validate() error {
	if strings.TrimSpace(v.OperatorAddress) == "" {
		return errors.New("operator address must be non-empty")
	}
	if v.UptimeWindow == 0 {
		return errors.New("uptime window must be positive")
	}
	if v.SignedBlocks+v.MissedBlocks < v.SignedBlocks {
		return errors.New("signed plus missed blocks overflowed")
	}
	if v.SignedBlocks+v.MissedBlocks > v.UptimeWindow {
		return errors.New("signed plus missed blocks cannot exceed uptime window")
	}
	if v.TotalBond == 0 {
		return errors.New("total bond must be positive")
	}
	if v.SelfBond > v.TotalBond {
		return errors.New("self bond cannot exceed total bond")
	}
	if v.GovernanceProposals > 0 && v.GovernanceVotes > v.GovernanceProposals {
		return errors.New("governance votes cannot exceed governance proposals")
	}
	if v.ConcentrationBps > BasisPoints {
		return fmt.Errorf("concentration cannot exceed %d bps", BasisPoints)
	}
	switch normalizeConcentrationStatus(v.ConcentrationStatus) {
	case ConcentrationStatusNormal, ConcentrationStatusNearCap, ConcentrationStatusOverloaded:
	default:
		return fmt.Errorf("unknown concentration status %q", v.ConcentrationStatus)
	}
	for i, slash := range v.SlashEvents {
		if err := slash.Validate(); err != nil {
			return fmt.Errorf("slash event %d: %w", i, err)
		}
	}
	for i, point := range v.CommissionHistory {
		if err := point.Validate(); err != nil {
			return fmt.Errorf("commission point %d: %w", i, err)
		}
		if i > 0 && v.CommissionHistory[i-1].Epoch == point.Epoch {
			return errors.New("commission history contains duplicate epoch")
		}
	}
	return nil
}

func (s ValidatorScore) Validate(params Params) error {
	if strings.TrimSpace(s.OperatorAddress) == "" {
		return errors.New("operator address must be non-empty")
	}
	bpsFields := []struct {
		name	string
		value	uint32
	}{
		{"self_bond_bps", s.SelfBondBps},
		{"last_commission_bps", s.LastCommissionBps},
		{"governance_participation_bps", s.GovernanceParticipationBps},
		{"concentration_bps", s.ConcentrationBps},
		{"uptime_score_bps", s.UptimeScoreBps},
		{"missed_block_score_bps", s.MissedBlockScoreBps},
		{"jail_score_bps", s.JailScoreBps},
		{"slash_history_score_bps", s.SlashHistoryScoreBps},
		{"self_bond_score_bps", s.SelfBondScoreBps},
		{"commission_score_bps", s.CommissionScoreBps},
		{"governance_score_bps", s.GovernanceScoreBps},
		{"decentralization_score_bps", s.DecentralizationScoreBps},
		{"identity_score_bps", s.IdentityScoreBps},
		{"overall_score_bps", s.OverallScoreBps},
		{"reward_multiplier_bps", s.RewardMultiplierBps},
	}
	for _, field := range bpsFields {
		if field.value > BasisPoints {
			return fmt.Errorf("%s cannot exceed %d bps", field.name, BasisPoints)
		}
	}
	if s.RewardMultiplierBps < params.MinRewardMultiplierBps {
		return errors.New("reward multiplier below configured minimum")
	}
	if s.ConsensusOverrideAllowed && !params.ConsensusOverrideEnabled {
		return errors.New("score cannot allow consensus override when params disable it")
	}
	return nil
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	metrics, err := CanonicalMetricInputs(g.Metrics)
	if err != nil {
		return err
	}
	if len(metrics) != len(g.Metrics) {
		return errors.New("metrics canonicalization changed metric count")
	}
	for i := range metrics {
		if !reflect.DeepEqual(metrics[i], g.Metrics[i]) {
			return errors.New("metrics must be sorted canonically")
		}
	}
	if len(g.Scores) != 0 {
		scores := append([]ValidatorScore(nil), g.Scores...)
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].OperatorAddress < scores[j].OperatorAddress
		})
		for i, score := range scores {
			if err := score.Validate(g.Params); err != nil {
				return fmt.Errorf("score %d: %w", i, err)
			}
			if i > 0 && scores[i-1].OperatorAddress == score.OperatorAddress {
				return errors.New("duplicate validator score")
			}
			if scores[i] != g.Scores[i] {
				return errors.New("scores must be sorted canonically")
			}
		}
	}
	return nil
}

func scoreValidator(params Params, epoch uint64, input ValidatorMetricInput) ValidatorScore {
	uptimeScore := ratioBps(input.SignedBlocks, input.UptimeWindow)
	missedBlockScore := subtractBps(BasisPoints, ratioBps(input.MissedBlocks, input.UptimeWindow))
	jailScore := subtractBps(BasisPoints, saturatingProductToBps(input.JailEvents, params.JailPenaltyBps))
	slashScore := slashHistoryScore(params, input.SlashEvents)
	selfBondBps := ratioBps(input.SelfBond, input.TotalBond)
	selfBondScore := cappedBps(uint64(selfBondBps) * uint64(BasisPoints) / uint64(params.SelfBondFullScoreBps))
	commissionScore, lastCommission := commissionHistoryScore(input.CommissionHistory)
	governanceScore := governanceParticipationScore(input.GovernanceVotes, input.GovernanceProposals)
	decentralizationScore := decentralizationScore(input.ConcentrationBps, input.ConcentrationStatus)
	identityScore := uint32(8_000)
	if input.IdentityMetadataComplete {
		identityScore = BasisPoints
	}
	overall := weightedOverallScore(params.Weights, uptimeScore, missedBlockScore, jailScore, slashScore, selfBondScore, commissionScore, governanceScore, decentralizationScore, identityScore)
	rewardMultiplier := objectiveRewardMultiplier(params, uptimeScore, missedBlockScore, jailScore, slashScore, decentralizationScore)
	return ValidatorScore{
		OperatorAddress:		input.OperatorAddress,
		Epoch:				epoch,
		UptimeWindow:			input.UptimeWindow,
		SignedBlocks:			input.SignedBlocks,
		MissedBlocks:			input.MissedBlocks,
		JailEvents:			input.JailEvents,
		SlashEventCount:		uint64(len(input.SlashEvents)),
		SelfBondBps:			selfBondBps,
		LastCommissionBps:		lastCommission,
		GovernanceParticipationBps:	governanceScore,
		ConcentrationBps:		input.ConcentrationBps,
		ConcentrationStatus:		normalizeConcentrationStatus(input.ConcentrationStatus),
		IdentityMetadataComplete:	input.IdentityMetadataComplete,
		UptimeScoreBps:			uptimeScore,
		MissedBlockScoreBps:		missedBlockScore,
		JailScoreBps:			jailScore,
		SlashHistoryScoreBps:		slashScore,
		SelfBondScoreBps:		selfBondScore,
		CommissionScoreBps:		commissionScore,
		GovernanceScoreBps:		governanceScore,
		DecentralizationScoreBps:	decentralizationScore,
		IdentityScoreBps:		identityScore,
		OverallScoreBps:		overall,
		RewardMultiplierBps:		rewardMultiplier,
		InformationalOnly:		!params.ObjectiveRewardModifierEnabled,
		ConsensusOverrideAllowed:	params.ConsensusOverrideEnabled,
	}
}

func slashHistoryScore(params Params, events []SlashEvent) uint32 {
	var penalty uint64
	for _, event := range events {
		penalty += uint64(params.SlashEventPenaltyBps) + uint64(event.FractionBps)
		if penalty >= uint64(BasisPoints) {
			return 0
		}
	}
	return subtractBps(BasisPoints, uint32(penalty))
}

func commissionHistoryScore(history []CommissionPoint) (uint32, uint32) {
	if len(history) == 0 {
		return BasisPoints, 0
	}
	var maxDelta uint32
	for i := 1; i < len(history); i++ {
		delta := absUint32Delta(history[i].CommissionBps, history[i-1].CommissionBps)
		if delta > maxDelta {
			maxDelta = delta
		}
	}
	return subtractBps(BasisPoints, maxDelta), history[len(history)-1].CommissionBps
}

func governanceParticipationScore(votes, proposals uint64) uint32 {
	if proposals == 0 {
		return BasisPoints
	}
	return ratioBps(votes, proposals)
}

func decentralizationScore(concentrationBps uint32, status string) uint32 {
	score := subtractBps(BasisPoints, concentrationBps)
	switch normalizeConcentrationStatus(status) {
	case ConcentrationStatusNearCap:
		return subtractBps(score, 500)
	case ConcentrationStatusOverloaded:
		return subtractBps(score, 1_500)
	default:
		return score
	}
}

func weightedOverallScore(weights ScoreWeights, uptime, missed, jail, slash, selfBond, commission, governance, decentralization, identity uint32) uint32 {
	total := uint64(uptime)*uint64(weights.UptimeBps) +
		uint64(missed)*uint64(weights.MissedBlocksBps) +
		uint64(jail)*uint64(weights.JailBps) +
		uint64(slash)*uint64(weights.SlashHistoryBps) +
		uint64(selfBond)*uint64(weights.SelfBondBps) +
		uint64(commission)*uint64(weights.CommissionBps) +
		uint64(governance)*uint64(weights.GovernanceBps) +
		uint64(decentralization)*uint64(weights.DecentralizationBps) +
		uint64(identity)*uint64(weights.IdentityBps)
	return uint32(total / uint64(BasisPoints))
}

func objectiveRewardMultiplier(params Params, uptime, missed, jail, slash, decentralization uint32) uint32 {
	if !params.ObjectiveRewardModifierEnabled {
		return BasisPoints
	}
	objective := (uint64(uptime) + uint64(missed) + uint64(jail) + uint64(slash) + uint64(decentralization)) / 5
	if objective < uint64(params.MinRewardMultiplierBps) {
		return params.MinRewardMultiplierBps
	}
	return uint32(objective)
}

func canonicalSlashEvents(events []SlashEvent) []SlashEvent {
	next := append([]SlashEvent(nil), events...)
	for i := range next {
		next[i].Reason = strings.TrimSpace(next[i].Reason)
	}
	sort.Slice(next, func(i, j int) bool {
		if next[i].Height == next[j].Height {
			return next[i].Reason < next[j].Reason
		}
		return next[i].Height < next[j].Height
	})
	return next
}

func canonicalCommissionHistory(history []CommissionPoint) []CommissionPoint {
	next := append([]CommissionPoint(nil), history...)
	sort.Slice(next, func(i, j int) bool {
		return next[i].Epoch < next[j].Epoch
	})
	return next
}

func normalizeConcentrationStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "", ConcentrationStatusNormal:
		return ConcentrationStatusNormal
	case ConcentrationStatusNearCap:
		return ConcentrationStatusNearCap
	case ConcentrationStatusOverloaded:
		return ConcentrationStatusOverloaded
	default:
		return strings.TrimSpace(status)
	}
}

func ratioBps(numerator, denominator uint64) uint32 {
	if denominator == 0 {
		return 0
	}
	if numerator >= denominator {
		return BasisPoints
	}
	hi, lo := bits.Mul64(numerator, uint64(BasisPoints))
	quotient, _ := bits.Div64(hi, lo, denominator)
	return uint32(quotient)
}

func subtractBps(value uint32, penalty uint32) uint32 {
	if penalty >= value {
		return 0
	}
	return value - penalty
}

func cappedBps(value uint64) uint32 {
	if value >= uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(value)
}

func saturatingProductToBps(count uint64, penaltyBps uint32) uint32 {
	if count == 0 || penaltyBps == 0 {
		return 0
	}
	if count > uint64(BasisPoints)/uint64(penaltyBps) {
		return BasisPoints
	}
	return uint32(count * uint64(penaltyBps))
}

func absUint32Delta(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
