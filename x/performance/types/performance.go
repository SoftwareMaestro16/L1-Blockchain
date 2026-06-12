package types

import (
	"errors"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	ModuleName	= "performance"
	StoreKey	= ModuleName

	QueryEndpointPerformanceRecord	= "QueryPerformanceRecord"
	QueryEndpointOperatorHistory	= "QueryOperatorPerformanceHistory"
	QueryEndpointRolePerformance	= "QueryRolePerformance"
	QueryEndpointRewardMultiplier	= "QueryRewardMultiplier"
	QueryEndpointDistributionReward	= "QueryPerformanceDistributionReward"
)

type RoleMetricCollectionInput struct {
	EpochID				uint64
	OperatorAddress			string
	Role				postypes.ValidatorRole
	AssignedTasks			uint64
	CompletedTasks			uint64
	MissedTasks			uint64
	InvalidTasks			uint64
	SignedBlocks			uint64
	TotalBlocks			uint64
	TaskParticipations		uint64
	MissedTaskParticipations	uint64
	CommittedLatencyWindow		bool
	LatencyAdvisoryOnly		bool
	LatencyTargetMillis		uint64
	LatencyP95Millis		uint64
	ValidSignatures			uint64
	InvalidSignatures		uint64
	ValidTaskOutputs		uint64
	InvalidTaskOutputs		uint64
	AcceptedEvidence		uint64
	EvidencePenaltyWeight		uint64
}

type PerformanceSnapshot struct {
	Records []postypes.PerformanceRecord
}

type QueryPerformanceRecordRequest struct {
	EpochID		uint64
	OperatorAddress	string
	Role		postypes.ValidatorRole
}

type QueryPerformanceRecordResponse struct {
	Record postypes.PerformanceRecord
}

type QueryOperatorPerformanceHistoryRequest struct {
	OperatorAddress	string
	Limit		uint32
}

type QueryOperatorPerformanceHistoryResponse struct {
	Records []postypes.PerformanceRecord
}

type QueryRolePerformanceRequest struct {
	EpochID	uint64
	Role	postypes.ValidatorRole
}

type QueryRolePerformanceResponse struct {
	Records []postypes.PerformanceRecord
}

type QueryRewardMultiplierRequest struct {
	EpochID		uint64
	OperatorAddress	string
	Role		postypes.ValidatorRole
}

type QueryRewardMultiplierResponse struct {
	RewardMultiplierBps uint32
}

type PerformanceDistributionInput struct {
	Performance	postypes.PerformanceRecord
	Reward		postypes.RewardInput
}

type PerformanceDistributionResult struct {
	Performance		postypes.PerformanceRecord
	OriginalRewardNaet	sdkmath.Int
	DampenedRewardNaet	sdkmath.Int
	Distribution		postypes.RewardDistribution
}

func CollectRolePerformanceMetrics(input RoleMetricCollectionInput) (postypes.PerformanceRecord, error) {
	input.OperatorAddress = strings.TrimSpace(input.OperatorAddress)
	if err := validateTaskAccounting(input.AssignedTasks, input.CompletedTasks, input.MissedTasks, input.InvalidTasks); err != nil {
		return postypes.PerformanceRecord{}, err
	}
	uptime, err := postypes.ComputeUptimeFactor(postypes.UptimeFactorInput{
		SignedBlocks:			input.SignedBlocks,
		TotalBlocks:			input.TotalBlocks,
		TaskParticipations:		input.TaskParticipations,
		MissedTaskParticipations:	input.MissedTaskParticipations,
	})
	if err != nil {
		return postypes.PerformanceRecord{}, err
	}
	latency, err := postypes.ComputeLatencyFactor(postypes.LatencyFactorInput{
		CommittedWindow:	input.CommittedLatencyWindow,
		AdvisoryOnly:		input.LatencyAdvisoryOnly,
		TargetMillis:		input.LatencyTargetMillis,
		P95Millis:		input.LatencyP95Millis,
	})
	if err != nil {
		return postypes.PerformanceRecord{}, err
	}
	correctness, err := postypes.ComputeCorrectnessScore(postypes.CorrectnessScoreInput{
		ValidSignatures:	input.ValidSignatures,
		InvalidSignatures:	input.InvalidSignatures,
		ValidTaskOutputs:	input.ValidTaskOutputs,
		InvalidTaskOutputs:	input.InvalidTaskOutputs,
		AcceptedEvidence:	input.AcceptedEvidence,
		EvidencePenaltyWeight:	input.EvidencePenaltyWeight,
	})
	if err != nil {
		return postypes.PerformanceRecord{}, err
	}
	return postypes.BuildPerformanceRecord(postypes.PerformanceRecordInput{
		EpochID:		input.EpochID,
		OperatorAddress:	input.OperatorAddress,
		Role:			input.Role,
		AssignedTasks:		input.AssignedTasks,
		CompletedTasks:		input.CompletedTasks,
		MissedTasks:		input.MissedTasks,
		InvalidTasks:		input.InvalidTasks,
		UptimeScoreBps:		uptime,
		LatencyScoreBps:	latency,
		CorrectnessScoreBps:	correctness,
	})
}

func SettlePerformanceDistribution(input PerformanceDistributionInput) (PerformanceDistributionResult, error) {
	if err := input.Performance.Validate(); err != nil {
		return PerformanceDistributionResult{}, err
	}
	if strings.TrimSpace(input.Reward.ValidatorID) != input.Performance.OperatorAddress {
		return PerformanceDistributionResult{}, errors.New("performance distribution validator mismatch")
	}
	if input.Reward.TotalRewardsNaet.IsNil() {
		input.Reward.TotalRewardsNaet = sdkmath.ZeroInt()
	}
	if input.Reward.TotalRewardsNaet.IsNegative() {
		return PerformanceDistributionResult{}, errors.New("performance distribution reward cannot be negative")
	}
	dampened := mulIntBps(input.Reward.TotalRewardsNaet, input.Performance.RewardMultiplierBps)
	reward := input.Reward
	reward.TotalRewardsNaet = dampened
	distribution, err := postypes.DistributeRewards(reward)
	if err != nil {
		return PerformanceDistributionResult{}, err
	}
	return PerformanceDistributionResult{
		Performance:		input.Performance,
		OriginalRewardNaet:	input.Reward.TotalRewardsNaet,
		DampenedRewardNaet:	dampened,
		Distribution:		distribution,
	}, nil
}

func NewPerformanceSnapshot(records []postypes.PerformanceRecord) (PerformanceSnapshot, error) {
	out := make([]postypes.PerformanceRecord, len(records))
	for i, record := range records {
		if err := record.Validate(); err != nil {
			return PerformanceSnapshot{}, err
		}
		out[i] = record
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareRecords(out[i], out[j]) < 0
	})
	return PerformanceSnapshot{Records: out}, nil
}

func QueryPerformanceRecord(snapshot PerformanceSnapshot, req QueryPerformanceRecordRequest) (QueryPerformanceRecordResponse, error) {
	req.OperatorAddress = strings.TrimSpace(req.OperatorAddress)
	if req.EpochID == 0 {
		return QueryPerformanceRecordResponse{}, errors.New("performance query epoch id is required")
	}
	if req.OperatorAddress == "" {
		return QueryPerformanceRecordResponse{}, errors.New("performance query operator address is required")
	}
	for _, record := range snapshot.Records {
		if record.EpochID == req.EpochID && record.OperatorAddress == req.OperatorAddress && record.Role == req.Role {
			return QueryPerformanceRecordResponse{Record: record}, nil
		}
	}
	return QueryPerformanceRecordResponse{}, errors.New("performance record not found")
}

func QueryOperatorPerformanceHistory(snapshot PerformanceSnapshot, req QueryOperatorPerformanceHistoryRequest) (QueryOperatorPerformanceHistoryResponse, error) {
	req.OperatorAddress = strings.TrimSpace(req.OperatorAddress)
	if req.OperatorAddress == "" {
		return QueryOperatorPerformanceHistoryResponse{}, errors.New("performance history operator address is required")
	}
	records := make([]postypes.PerformanceRecord, 0)
	for _, record := range snapshot.Records {
		if record.OperatorAddress == req.OperatorAddress {
			records = append(records, record)
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID > records[j].EpochID
		}
		return records[i].Role < records[j].Role
	})
	if req.Limit > 0 && uint32(len(records)) > req.Limit {
		records = records[:req.Limit]
	}
	return QueryOperatorPerformanceHistoryResponse{Records: records}, nil
}

func QueryRolePerformance(snapshot PerformanceSnapshot, req QueryRolePerformanceRequest) (QueryRolePerformanceResponse, error) {
	if req.EpochID == 0 {
		return QueryRolePerformanceResponse{}, errors.New("role performance query epoch id is required")
	}
	records := make([]postypes.PerformanceRecord, 0)
	for _, record := range snapshot.Records {
		if record.EpochID == req.EpochID && record.Role == req.Role {
			records = append(records, record)
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].RewardMultiplierBps != records[j].RewardMultiplierBps {
			return records[i].RewardMultiplierBps > records[j].RewardMultiplierBps
		}
		return records[i].OperatorAddress < records[j].OperatorAddress
	})
	return QueryRolePerformanceResponse{Records: records}, nil
}

func QueryRewardMultiplier(snapshot PerformanceSnapshot, req QueryRewardMultiplierRequest) (QueryRewardMultiplierResponse, error) {
	res, err := QueryPerformanceRecord(snapshot, QueryPerformanceRecordRequest{
		EpochID:		req.EpochID,
		OperatorAddress:	req.OperatorAddress,
		Role:			req.Role,
	})
	if err != nil {
		return QueryRewardMultiplierResponse{}, err
	}
	return QueryRewardMultiplierResponse{RewardMultiplierBps: res.Record.RewardMultiplierBps}, nil
}

func validateTaskAccounting(assigned uint64, completed uint64, missed uint64, invalid uint64) error {
	if completed > assigned {
		return errors.New("completed tasks cannot exceed assigned tasks")
	}
	if missed > assigned {
		return errors.New("missed tasks cannot exceed assigned tasks")
	}
	if invalid > assigned {
		return errors.New("invalid tasks cannot exceed assigned tasks")
	}
	if completed+missed < completed {
		return errors.New("performance task count overflow")
	}
	observed := completed + missed
	if observed+invalid < observed {
		return errors.New("performance task count overflow")
	}
	if observed+invalid > assigned {
		return errors.New("performance task counts exceed assigned tasks")
	}
	return nil
}

func compareRecords(left postypes.PerformanceRecord, right postypes.PerformanceRecord) int {
	if left.EpochID != right.EpochID {
		return compareUint64(left.EpochID, right.EpochID)
	}
	if left.OperatorAddress != right.OperatorAddress {
		if left.OperatorAddress < right.OperatorAddress {
			return -1
		}
		return 1
	}
	if left.Role < right.Role {
		return -1
	}
	if left.Role > right.Role {
		return 1
	}
	return 0
}

func compareUint64(left uint64, right uint64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	return value.MulRaw(int64(bps)).QuoRaw(10_000)
}
