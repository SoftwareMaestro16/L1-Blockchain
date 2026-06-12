package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	identitytypes "github.com/sovereign-l1/l1/x/identity/types"
	memotypes "github.com/sovereign-l1/l1/x/memo/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
)

const (
	StageCheckTxDecode	= "checktx.decode"
	StageCheckTxSignatures	= "checktx.validate_signatures"
	StageCheckTxFees	= "checktx.validate_fees"
	StageCheckTxMemo	= "checktx.validate_memo"
	StageCheckTxStateless	= "checktx.stateless_checks"
	StageDeliverAnte	= "deliver.ante"
	StageExecutionContext	= "deliver.execution_context"
	StageResolverLookup	= "deliver.resolver_lookup"
	StageReputationLimits	= "deliver.reputation_limits"
	StageModuleDispatch	= "deliver.module_dispatch"
	StageAsyncEnqueue	= "deliver.async_enqueue"
	StageEventEmit		= "deliver.event_emit"
	StageStateWrite		= "deliver.state_write"

	RouteBankTransfer	= "bank_transfer"
	RouteResolverPayment	= "resolver_payment"
	RouteSBTProofRevoke	= "sbt_proof_revoke"
	RouteContractCall	= "contract_call"
	RouteDomainAuction	= "domain_auction_bid"
	RouteDomainRenewal	= "domain_renewal"

	VMRouteNone	= ""
	VMRouteAVM	= "avm"
	VMRouteCosmWasm	= "cosmwasm"
)

type ExecutionEnvelope struct {
	TxHash			[]byte
	Sender			sdk.AccAddress
	Receiver		sdk.AccAddress
	Route			string
	VMRoute			string
	GasLimit		uint64
	Fee			sdk.Coins
	Memo			memotypes.TxMetadata
	ResolverDomain		string
	ResolverRecord		*identitytypes.ResolverRecord
	DomainRecord		*identitytypes.DomainRecord
	Identity		*reputationtypes.IdentityReputation
	SenderStake		sdkmath.Int
	BlockGasConsumed	uint64
	BlockTxCount		uint64
	SenderTxCount		uint64
	QueuedMessages		uint32
	AsyncMessages		[]async.MessageEnvelope
	ModuleEvents		[]string
	BlockHeight		uint64
	TimestampUnix		int64
}

type PipelineParams struct {
	FeeParams		feestypes.Params
	MemoParams		memotypes.MemoParams
	ReputationPolicy	reputationtypes.UsagePolicy
	AsyncParams		async.Params
}

type ExecutionTrace struct {
	Steps []TraceStep
}

type TraceStep struct {
	Stage	string
	Detail	string
}

type PipelineResult struct {
	Envelope	ExecutionEnvelope
	FeeQuote	feestypes.FeeQuote
	MemoFee		sdk.Coin
	ResolvedTarget	sdk.AccAddress
	Trace		ExecutionTrace
	Events		[]string
	StateWrite	bool
	AsyncQueued	bool
}

func DefaultPipelineParams() PipelineParams {
	return PipelineParams{
		FeeParams:		feestypes.DefaultParams(),
		MemoParams:		memotypes.DefaultMemoParams(),
		ReputationPolicy:	reputationtypes.DefaultUsagePolicy(),
		AsyncParams:		async.DefaultParams(),
	}
}

func CheckTx(envelope ExecutionEnvelope, params PipelineParams) (PipelineResult, error) {
	trace := ExecutionTrace{}
	trace.Add(StageCheckTxDecode, "decoded execution envelope")
	if err := ValidateExecutionEnvelope(envelope, params, false); err != nil {
		return PipelineResult{}, traceError(trace, err)
	}
	trace.Add(StageCheckTxSignatures, "signature validation must precede execution")
	quote, err := feestypes.ValidateAdmission(params.FeeParams, feestypes.AdmissionInput{
		Fee:			envelope.Fee,
		GasLimit:		envelope.GasLimit,
		BlockGasConsumed:	envelope.BlockGasConsumed,
		BlockTxCount:		envelope.BlockTxCount,
		SenderTxCount:		envelope.SenderTxCount,
		SenderStake:		envelope.SenderStake,
	})
	if err != nil {
		return PipelineResult{}, traceError(trace, err)
	}
	trace.Add(StageCheckTxFees, quote.RequiredFee.String())
	memoFee, denom, err := memotypes.MemoFee(envelope.Memo, params.MemoParams, reputationScore(envelope), memotypes.DefaultCongestionBps)
	if err != nil {
		return PipelineResult{}, traceError(trace, err)
	}
	trace.Add(StageCheckTxMemo, fmt.Sprintf("%s%s", memoFee.String(), denom))
	trace.Add(StageCheckTxStateless, "stateless checks passed")
	return PipelineResult{
		Envelope:	envelope,
		FeeQuote:	quote,
		MemoFee:	sdk.NewCoin(denom, memoFee),
		Trace:		trace,
	}, nil
}

func DeliverTx(envelope ExecutionEnvelope, params PipelineParams) (PipelineResult, error) {
	result, err := CheckTx(envelope, params)
	if err != nil {
		return PipelineResult{}, err
	}
	trace := result.Trace
	trace.Add(StageDeliverAnte, "ante accepted")
	trace.Add(StageExecutionContext, "execution context created")
	resolved, err := ResolveEnvelopeTarget(envelope)
	if err != nil {
		return PipelineResult{}, traceError(trace, err)
	}
	if len(resolved) > 0 {
		trace.Add(StageResolverLookup, "resolved domain target")
		result.ResolvedTarget = resolved
	}
	if envelope.Identity != nil {
		if err := reputationtypes.ValidateIdentityTxUsage(envelope.Identity, uint32(envelope.SenderTxCount), true, true, params.ReputationPolicy); err != nil {
			return PipelineResult{}, traceError(trace, err)
		}
		if err := reputationtypes.ValidateIdentityAsyncQueueUsage(envelope.Identity, envelope.QueuedMessages, true, params.ReputationPolicy); err != nil {
			return PipelineResult{}, traceError(trace, err)
		}
		trace.Add(StageReputationLimits, "reputation limits accepted")
	}
	if IsVMRoute(envelope.Route) {
		if err := ValidateVMRoute(envelope.VMRoute); err != nil {
			return PipelineResult{}, traceError(trace, err)
		}
	}
	trace.Add(StageModuleDispatch, envelope.Route)
	if len(envelope.AsyncMessages) > 0 {
		if len(envelope.AsyncMessages) > int(params.AsyncParams.MaxMessagesPerTx) {
			return PipelineResult{}, traceError(trace, fmt.Errorf("async messages per tx must be <= %d", params.AsyncParams.MaxMessagesPerTx))
		}
		for _, msg := range envelope.AsyncMessages {
			if err := msg.Validate(params.AsyncParams); err != nil {
				return PipelineResult{}, traceError(trace, err)
			}
		}
		result.AsyncQueued = true
		trace.Add(StageAsyncEnqueue, "async messages accepted for enqueue")
	}
	events := DeterministicEvents(envelope)
	trace.Add(StageEventEmit, "deterministic events collected")
	trace.Add(StageStateWrite, "state write follows successful dispatch")
	result.Trace = trace
	result.Events = events
	result.StateWrite = true
	return result, nil
}

func ValidateExecutionEnvelope(envelope ExecutionEnvelope, params PipelineParams, requireTxHash bool) error {
	if requireTxHash && len(envelope.TxHash) == 0 {
		return errors.New("execution tx hash is required")
	}
	if len(envelope.Sender) == 0 {
		return errors.New("execution sender is required")
	}
	if err := addressing.RejectZeroAddress("execution sender", envelope.Sender); err != nil {
		return err
	}
	if len(envelope.Receiver) > 0 {
		if err := addressing.RejectZeroAddress("execution receiver", envelope.Receiver); err != nil {
			return err
		}
	}
	if !IsExecutionRoute(envelope.Route) {
		return fmt.Errorf("invalid execution route %q", envelope.Route)
	}
	if envelope.GasLimit == 0 {
		return errors.New("execution gas limit must be positive")
	}
	if err := memotypes.ValidateTxMetadata(envelope.Memo, params.MemoParams); err != nil {
		return err
	}
	if envelope.Identity != nil {
		if err := reputationtypes.ValidateIdentityReputation(envelope.Identity); err != nil {
			return err
		}
	}
	return nil
}

func ResolveEnvelopeTarget(envelope ExecutionEnvelope) (sdk.AccAddress, error) {
	if envelope.ResolverDomain == "" {
		return nil, nil
	}
	if envelope.ResolverRecord == nil || envelope.DomainRecord == nil {
		return nil, errors.New("resolver domain requires resolver and domain records")
	}
	return identitytypes.ResolvePaymentTarget(*envelope.ResolverRecord, *envelope.DomainRecord, envelope.TimestampUnix)
}

func DeterministicEvents(envelope ExecutionEnvelope) []string {
	events := append([]string(nil), envelope.ModuleEvents...)
	if envelope.Memo.Memo != "" || len(envelope.Memo.MemoHash) > 0 {
		events = append(events, memotypes.EventTypeMemoAttached)
	}
	if envelope.ResolverDomain != "" {
		events = append(events, identitytypes.ResolverEventSet)
	}
	sort.Strings(events)
	return events
}

func (t *ExecutionTrace) Add(stage, detail string) {
	t.Steps = append(t.Steps, TraceStep{Stage: stage, Detail: detail})
}

func IsExecutionRoute(route string) bool {
	switch route {
	case RouteBankTransfer,
		RouteResolverPayment,
		RouteSBTProofRevoke,
		RouteContractCall,
		RouteDomainAuction,
		RouteDomainRenewal:
		return true
	default:
		return false
	}
}

func IsVMRoute(route string) bool {
	return route == RouteContractCall
}

func ValidateVMRoute(route string) error {
	switch route {
	case VMRouteAVM, VMRouteCosmWasm:
		return nil
	default:
		return fmt.Errorf("invalid VM route %q", route)
	}
}

func reputationScore(envelope ExecutionEnvelope) uint8 {
	return uint8(reputationtypes.IdentityScore(envelope.Identity) / 100)
}

func traceError(trace ExecutionTrace, err error) error {
	return fmt.Errorf("%w; trace_steps=%d", err, len(trace.Steps))
}
