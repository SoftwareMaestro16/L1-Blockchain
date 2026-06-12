package async

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func BenchmarkQueueProcessBlock(b *testing.B) {
	source := benchmarkAddr(1)
	body := []byte("bench")

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		executor := benchmarkExecutor(b)
		destination := benchmarkDeployContract(b, executor, source, nil)
		if err := executor.RegisterHandler(destination, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			_ = msg
			return ExecutionResult{NewState: contract.State, ResultCode: ResultOK}
		}); err != nil {
			b.Fatal(err)
		}
		for j := uint32(0); j < executor.params.MaxMessagesPerBlock; j++ {
			if err := executor.EnqueueMessage(benchmarkMessage(source, destination, uint64(j), body)); err != nil {
				b.Fatal(err)
			}
		}
		if receipts, err := executor.ProcessBlock(uint64(i + 1)); err != nil {
			b.Fatal(err)
		} else if len(receipts) != int(executor.params.MaxMessagesPerBlock) {
			b.Fatalf("expected %d receipts, got %d", executor.params.MaxMessagesPerBlock, len(receipts))
		}
	}
}

func BenchmarkContractStateExportImport(b *testing.B) {
	executor := benchmarkExecutor(b)
	deployer := benchmarkAddr(1)
	for i := 0; i < 16; i++ {
		benchmarkDeployContract(b, executor, deployer, bytes.Repeat([]byte{byte(i)}, 1024))
	}
	exported := executor.ExportState()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := ExportedState{
			Params:			exported.Params,
			Contracts:		cloneContractSlice(exported.Contracts),
			Queue:			cloneQueuedMessages(exported.Queue),
			Inbox:			cloneQueuedMap(exported.Inbox),
			Outbox:			cloneQueuedMap(exported.Outbox),
			DeadLetters:		cloneDeadLetters(exported.DeadLetters),
			Receipts:		cloneReceipts(exported.Receipts),
			NextSequence:		exported.NextSequence,
			NextTxIndex:		exported.NextTxIndex,
			NextDeadLetterSequence:	exported.NextDeadLetterSequence,
			BlockHeight:		exported.BlockHeight,
			Metrics:		exported.Metrics,
		}
		if _, err := ImportState(state); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkExecutor(b *testing.B) *Executor {
	b.Helper()
	executor, err := NewExecutor(DefaultParams())
	if err != nil {
		b.Fatal(err)
	}
	return executor
}

func benchmarkDeployContract(b *testing.B, executor *Executor, deployer sdk.AccAddress, state []byte) sdk.AccAddress {
	b.Helper()
	codeHash := bytes.Repeat([]byte{0x42}, CodeHashLength)
	address, err := executor.DeployContract(deployer, codeHash, []byte{byte(len(executor.contracts) + 1)}, state, sdkmath.NewInt(10_000))
	if err != nil {
		b.Fatal(err)
	}
	return address
}

func benchmarkMessage(source, destination sdk.AccAddress, queryID uint64, body []byte) MessageEnvelope {
	return MessageEnvelope{
		Source:		source,
		Destination:	destination,
		Value:		sdk.NewCoin(appparams.BaseDenom, sdkmath.NewInt(1)),
		Opcode:		1,
		QueryID:	queryID + 1,
		Body:		body,
		Bounce:		true,
		GasLimit:	100_000,
		ForwardFee:	sdk.NewCoin(appparams.BaseDenom, DefaultParams().ForwardingFee),
	}
}

func benchmarkAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}

func cloneContractSlice(contracts []ContractAccount) []ContractAccount {
	out := make([]ContractAccount, len(contracts))
	for i, contract := range contracts {
		out[i] = cloneContract(contract)
	}
	return out
}
