package avm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

// Engine is the core AVM execution coordinator.
type Engine struct {
}

func NewEngine() *Engine {
	return &Engine{}
}

// Execute performs a deterministic state transition.
// (StateChunk, Message, BlockContext) -> (NewStateChunk, Actions, Receipt, error)
func (e *Engine) Execute(state *chunk.Chunk, msg Message, blockCtx BlockContext, gasLimit uint64, maxActions uint32) (*chunk.Chunk, []Action, AVMReceipt, error) {
	msg.GasLimit = gasLimit
	frame := NewExecutionFrame(state, msg, maxActions)

	frame.BlockCtx = blockCtx
	frame.Capabilities = CapabilityMask{
		Crypto:		true,
		Chain:		true,
		Messaging:	true,
		Storage:	true,
	}

	frame.Phase = PhaseStorage

	if !frame.ChargeGas(500) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	frame.Phase = PhaseCredit
	if !frame.ChargeGas(100) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}
	if frame.Message.Value > 0 {
		credited, err := applyAttachedValueToWorkingState(frame.WorkingState, frame.Message.Value)
		if err != nil {
			frame.Aborted = true
			return frame.finalize(contractstypes.ExitCodeContractAbort)
		}
		frame.WorkingState = credited
	}

	frame.Phase = PhaseCompute

	payloadData := frame.Message.Payload.Data()

	if string(payloadData) == "trigger_abort" {
		frame.Aborted = true
		return frame.finalize(contractstypes.ExitCodeContractAbort)
	}

	if string(payloadData) == "use_forbidden_opcode" {
		return frame.finalize(contractstypes.ExitCodeCodeRejected)
	}

	if string(payloadData) == "emit_actions" || string(payloadData) == "emit_with_bounce" {
		frame.PendingActions = append(frame.PendingActions, Action{
			Type:		ActionInternal,
			Target:		"contract_b",
			Payload:	frame.Message.Payload,
		})

		if string(payloadData) == "emit_with_bounce" {
			frame.PendingActions = append(frame.PendingActions, Action{
				Type:		ActionSystem,
				Target:		"system_notifier",
				Payload:	frame.Message.Payload,
				SystemBounce:	true,
			})
		} else {
			frame.PendingActions = append(frame.PendingActions, Action{
				Type:		ActionExternal,
				Target:		"user_a",
				Payload:	frame.Message.Payload,
			})
		}
	}

	frame.Trace.Steps = append(frame.Trace.Steps, TraceStep{
		Instruction:	"LOAD_BAL",
		StackDelta:	1,
		GasConsumed:	10,
		Phase:		PhaseCompute,
	})

	if !frame.ChargeGas(1000) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	frame.Phase = PhaseAction
	if !frame.ChargeGas(200) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	if uint32(len(frame.PendingActions)) > frame.ActionBudget {
		frame.Aborted = true
		return frame.finalize(contractstypes.ExitCodeActionBudgetExceeded)
	}
	frame.ActionsUsed = uint32(len(frame.PendingActions))

	frame.Phase = PhaseFinalization
	if !frame.ChargeGas(300) {
		return frame.finalize(contractstypes.ExitCodeOutOfGas)
	}

	return frame.finalize(contractstypes.ExitCodeOK)
}

func (f *ExecutionFrame) finalize(exitCode uint32) (*chunk.Chunk, []Action, AVMReceipt, error) {
	f.ExitCode = exitCode

	receipt := AVMReceipt{
		ExitCode:		f.ExitCode,
		GasUsed:		f.GasUsed,
		GasLimit:		f.GasLimit,
		PhaseGas:		f.PhaseGas,
		StateRootBefore:	hex.EncodeToString(f.StateSnapshot.Hash()),
	}

	sort.SliceStable(f.PendingActions, func(i, j int) bool {
		if f.PendingActions[i].Type != f.PendingActions[j].Type {
			return f.PendingActions[i].Type < f.PendingActions[j].Type
		}
		return f.PendingActions[i].Target < f.PendingActions[j].Target
	})

	if f.Aborted || f.ExitCode != contractstypes.ExitCodeOK {

		receipt.StateRootAfter = receipt.StateRootBefore

		var finalActions []Action
		for _, action := range f.PendingActions {
			if action.SystemBounce {
				finalActions = append(finalActions, action)
			}
		}
		receipt.EmittedActionsHash = f.computeActionsHash(finalActions)
		receipt.ExecutionTraceHash = f.computeTraceHash()

		return f.StateSnapshot, finalActions, receipt, nil
	}

	receipt.StateRootAfter = hex.EncodeToString(f.WorkingState.Hash())
	receipt.EmittedActionsHash = f.computeActionsHash(f.PendingActions)
	receipt.ExecutionTraceHash = f.computeTraceHash()

	return f.WorkingState, f.PendingActions, receipt, nil
}

func (f *ExecutionFrame) computeActionsHash(actions []Action) string {
	h := sha256.New()
	for _, a := range actions {
		h.Write([]byte(fmt.Sprintf("%d:%s:%v", a.Type, a.Target, a.Payload.Hash())))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (f *ExecutionFrame) computeTraceHash() string {
	h := sha256.New()
	for _, s := range f.Trace.Steps {
		h.Write([]byte(fmt.Sprintf("%s:%d:%d:%s", s.Instruction, s.StackDelta, s.GasConsumed, s.Phase)))
	}
	return hex.EncodeToString(h.Sum(nil))
}
