package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
)

var _ reputationpb.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) reputationpb.MsgServer	{ return msgServer{Keeper: k} }

func (m msgServer) UpdateReputationParams(ctx context.Context, msg *reputationpb.MsgUpdateReputationParams) (*reputationpb.MsgUpdateReputationParamsResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	var params types.ReputationParams
	if err := json.Unmarshal([]byte(msg.ParamsJson), &params); err != nil {
		return nil, err
	}
	if err := authorizeFromState(state, msg.Authority); err != nil {
		return nil, err
	}
	params.Authority = strings.TrimSpace(params.Authority)
	if params.Authority == "" {
		params.Authority = state.Params.Authority
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	state.Params = params
	return &reputationpb.MsgUpdateReputationParamsResponse{}, m.SetState(ctx, state)
}

func (m msgServer) ApplyReputationPenalty(ctx context.Context, msg *reputationpb.MsgApplyReputationPenalty) (*reputationpb.MsgApplyReputationPenaltyResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	if err := authorizeFromState(state, msg.Authority); err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	userAddr := addressing.FormatAccAddress(subject)

	switch msg.SubjectType {
	case types.SubjectAccount:
		return m.applyIdentityPenalty(ctx, state, userAddr, msg)
	case types.SubjectValidator:
		return m.applyValidatorPenalty(ctx, state, userAddr, msg)
	default:
		return nil, errors.New("unsupported subject type for penalty; only account and validator are supported")
	}
}

func (m msgServer) applyIdentityPenalty(ctx context.Context, state types.ConsolidatedReputationState, userAddr string, msg *reputationpb.MsgApplyReputationPenalty) (*reputationpb.MsgApplyReputationPenaltyResponse, error) {
	id, found := types.FindIdentity(state, userAddr)
	if !found {
		id = *types.NewIdentityReputation(userAddr)
	}

	switch msg.Component {
	case types.ComponentSpam:
		id.RecordSpam(msg.Epoch)
	case types.ComponentSlashing:
		id.RecordSlashEvent(msg.Epoch)
	case types.ComponentMissedBlock:
		id.RecordFailedTx(msg.Epoch)
	default:
		id.RecordFailedTx(msg.Epoch)
	}

	id.Score = types.ComputeIdentityScore(&id)
	id.Confidence = types.ComputeConfidence(&id)
	id.LastUpdateHeight = msg.Epoch

	state = types.UpsertIdentity(state, id)
	respJSON, _ := mustJSON(id)
	return &reputationpb.MsgApplyReputationPenaltyResponse{RecordJson: respJSON}, m.SetState(ctx, state)
}

func (m msgServer) applyValidatorPenalty(ctx context.Context, state types.ConsolidatedReputationState, userAddr string, msg *reputationpb.MsgApplyReputationPenalty) (*reputationpb.MsgApplyReputationPenaltyResponse, error) {
	vs, found := types.FindValidatorScore(state, userAddr)
	if !found {
		vs = *types.NewValidatorScore(userAddr)
	}

	switch msg.Component {
	case types.ComponentMissedBlock:
		vs.MissedBlocksPenalty += uint32(msg.Amount)
	case types.ComponentSlashing:
		vs.SlashingPenalty += uint32(msg.Amount)
	}
	vs.TotalScore = types.ComputeValidatorTotalScore(&vs)
	vs.LastUpdateHeight = msg.Epoch

	state = types.UpsertValidatorScore(state, vs)
	respJSON, _ := mustJSON(vs)
	return &reputationpb.MsgApplyReputationPenaltyResponse{RecordJson: respJSON}, m.SetState(ctx, state)
}

func (m msgServer) ApplyReputationReward(ctx context.Context, msg *reputationpb.MsgApplyReputationReward) (*reputationpb.MsgApplyReputationRewardResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	if err := authorizeFromState(state, msg.Authority); err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	userAddr := addressing.FormatAccAddress(subject)

	switch msg.SubjectType {
	case types.SubjectAccount:
		return m.applyIdentityReward(ctx, state, userAddr, msg)
	case types.SubjectValidator:
		return m.applyValidatorReward(ctx, state, userAddr, msg)
	default:
		return nil, errors.New("unsupported subject type for reward")
	}
}

func (m msgServer) applyIdentityReward(ctx context.Context, state types.ConsolidatedReputationState, userAddr string, msg *reputationpb.MsgApplyReputationReward) (*reputationpb.MsgApplyReputationRewardResponse, error) {
	id, found := types.FindIdentity(state, userAddr)
	if !found {
		id = *types.NewIdentityReputation(userAddr)
	}

	switch msg.Component {
	case types.ComponentUptime:
		id.RecordSuccessfulTx(msg.Epoch)
	case types.ComponentRecovery:
		id.RecordRecoveryEvent(msg.Epoch)
	case types.ComponentVolume:
		id.RecordSuccessfulTx(msg.Epoch)
	default:
		id.RecordSuccessfulTx(msg.Epoch)
	}

	id.Score = types.ComputeIdentityScore(&id)
	id.Confidence = types.ComputeConfidence(&id)
	id.LastUpdateHeight = msg.Epoch

	state = types.UpsertIdentity(state, id)
	respJSON, _ := mustJSON(id)
	return &reputationpb.MsgApplyReputationRewardResponse{RecordJson: respJSON}, m.SetState(ctx, state)
}

func (m msgServer) applyValidatorReward(ctx context.Context, state types.ConsolidatedReputationState, userAddr string, msg *reputationpb.MsgApplyReputationReward) (*reputationpb.MsgApplyReputationRewardResponse, error) {
	vs, found := types.FindValidatorScore(state, userAddr)
	if !found {
		vs = *types.NewValidatorScore(userAddr)
	}

	switch msg.Component {
	case types.ComponentUptime:
		vs.UptimeScore += uint32(msg.Amount)
	case types.ComponentRecovery:
		vs.CommissionBehavior += uint32(msg.Amount)
	default:
		vs.UptimeScore += uint32(msg.Amount)
	}
	vs.TotalScore = types.ComputeValidatorTotalScore(&vs)
	vs.LastUpdateHeight = msg.Epoch

	state = types.UpsertValidatorScore(state, vs)
	respJSON, _ := mustJSON(vs)
	return &reputationpb.MsgApplyReputationRewardResponse{RecordJson: respJSON}, m.SetState(ctx, state)
}

func (m msgServer) RecomputeReputation(ctx context.Context, msg *reputationpb.MsgRecomputeReputation) (*reputationpb.MsgRecomputeReputationResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	if err := authorizeFromState(state, msg.Authority); err != nil {
		return nil, err
	}
	subject, err := parseAddress(msg.Subject)
	if err != nil {
		return nil, err
	}
	userAddr := addressing.FormatAccAddress(subject)

	switch msg.SubjectType {
	case types.SubjectAccount:
		id, found := types.FindIdentity(state, userAddr)
		if !found {
			return nil, errors.New("identity reputation not found")
		}
		id.Score = types.ComputeIdentityScore(&id)
		id.Confidence = types.ComputeConfidence(&id)
		state = types.UpsertIdentity(state, id)
		respJSON, _ := mustJSON(id)
		return &reputationpb.MsgRecomputeReputationResponse{RecordJson: respJSON}, m.SetState(ctx, state)
	case types.SubjectValidator:
		vs, found := types.FindValidatorScore(state, userAddr)
		if !found {
			return nil, errors.New("validator score not found")
		}
		vs.TotalScore = types.ComputeValidatorTotalScore(&vs)
		state = types.UpsertValidatorScore(state, vs)
		respJSON, _ := mustJSON(vs)
		return &reputationpb.MsgRecomputeReputationResponse{RecordJson: respJSON}, m.SetState(ctx, state)
	default:
		return nil, errors.New("unsupported subject type for recompute")
	}
}

func (k Keeper) ClaimStakeReputation(ctx context.Context, msg types.MsgClaimStakeReputation) (types.ReputationClaim, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return types.ReputationClaim{}, err
	}
	if err := authorizeFromState(state, msg.Authority); err != nil {
		return types.ReputationClaim{}, err
	}
	account, err := addressing.ParseUserAddress("stake reputation account", msg.Account)
	if err != nil {
		return types.ReputationClaim{}, err
	}
	userAddr := addressing.FormatAccAddress(account)

	if strings.TrimSpace(msg.PoolID) == "" {
		return types.ReputationClaim{}, errors.New("stake reputation pool id is required")
	}
	if msg.TimestampUnix == 0 {
		return types.ReputationClaim{}, errors.New("stake reputation timestamp must be positive")
	}
	if msg.PoolShares > 0 && msg.PoolTotalShares == 0 {
		return types.ReputationClaim{}, errors.New("stake reputation total pool shares must be positive when shares are positive")
	}

	id, found := types.FindIdentity(state, userAddr)
	if !found {
		id = *types.NewIdentityReputation(userAddr)
	}

	duration := msg.TimestampUnix
	if id.LastUpdateTime > 0 && msg.TimestampUnix > uint64(id.LastUpdateTime) {
		duration = msg.TimestampUnix - uint64(id.LastUpdateTime)
	}

	seconds := duration
	stakeAmount := msg.PoolActiveStake

	if seconds > 0 && stakeAmount > 0 {
		stakeTimeSeconds := seconds
		if stakeTimeSeconds > 365*24*3600 {
			stakeTimeSeconds = 365 * 24 * 3600
		}
		id.RecordStakeTime(stakeTimeSeconds, msg.TimestampUnix)
		id.Score = types.ComputeIdentityScore(&id)
		id.Confidence = types.ComputeConfidence(&id)
		id.LastUpdateHeight = msg.TimestampUnix
	}

	state = types.UpsertIdentity(state, id)

	claim := types.ReputationClaim{
		Account:		userAddr,
		Score:			id.Score,
		Confidence:		id.Confidence,
		StakeTimeAccumulator:	id.StakeTimeAccumulator,
		ClaimHeight:		msg.TimestampUnix,
	}
	return claim, k.SetState(ctx, state)
}

func authorizeFromState(state types.ConsolidatedReputationState, authority string) error {
	authority = strings.TrimSpace(authority)
	if err := addressing.ValidateAuthorityAddress("reputation message authority", authority); err != nil {
		return err
	}
	if authority != state.Params.Authority {
		return errors.New("reputation message requires authority")
	}
	return nil
}
