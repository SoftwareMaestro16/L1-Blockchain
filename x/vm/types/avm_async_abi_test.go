package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMAsyncCallPlanIsMessageDrivenAndNonBlocking(t *testing.T) {
	msg := mustAVMAsyncCallMessage(t)
	promise := mustAVMPromise(t, msg.ID)
	plan, err := NewAVMAsyncCallPlan(AVMAsyncCallPlan{
		Height:			msg.CreatedHeight,
		CallerContract:		"contract-a",
		Message:		msg,
		Promise:		promise,
		AwaitNonBlocking:	true,
		PersistedPromiseID:	promise.PromiseID,
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeAVMMessageRoot([]AVMAsyncMessage{msg}), plan.OutboxRoot)
	require.Equal(t, ComputeAVMPromiseRoot([]AVMPromiseState{promise}), plan.PromiseRoot)
	require.Equal(t, promise.PromiseID, plan.PersistedPromiseID)

	blocking := plan
	blocking.AwaitNonBlocking = false
	blocking.PlanHash = ComputeAVMAsyncCallPlanHash(blocking)
	require.ErrorContains(t, blocking.Validate(), "non-blocking")

	local := plan
	local.Message.SourceZone = local.Message.DestinationZone
	local.Message.ID = DeriveAVMAsyncMessageID(local.Message)
	local.Promise.MessageID = local.Message.ID
	local.Promise.PromiseHash = ComputeAVMPromiseHash(local.Promise)
	local.OutboxRoot = ComputeAVMMessageRoot([]AVMAsyncMessage{local.Message})
	local.PromiseRoot = ComputeAVMPromiseRoot([]AVMPromiseState{local.Promise})
	local.PlanHash = ComputeAVMAsyncCallPlanHash(local)
	require.ErrorContains(t, local.Validate(), "cross-zone")
}

func TestAVMPromiseResolutionAndTimeoutAreFutureMessages(t *testing.T) {
	msg := mustAVMAsyncCallMessage(t)
	promise := mustAVMPromise(t, msg.ID)
	resolutionMsg := mustAVMPromiseMessage(t, AVMPayloadTypePromiseResolution, 13, 20)
	resolution, err := NewAVMPromiseResolution(AVMPromiseResolution{
		PromiseID:		promise.PromiseID,
		OriginalMessageID:	promise.MessageID,
		ResolutionMessage:	resolutionMsg,
		Status:			AVMPromiseResolved,
		ReceiptHash:		ComputeAVMBytesHash([]byte("receipt")),
		ReturnHash:		ComputeAVMBytesHash([]byte("return")),
		DeliveryHeight:		13,
	})
	require.NoError(t, err)
	updated, err := ApplyAVMPromiseResolution(promise, resolution)
	require.NoError(t, err)
	require.Equal(t, AVMPromiseResolved, updated.Status)
	require.Equal(t, resolution.ReceiptHash, updated.ReceiptHash)

	early := resolution
	early.DeliveryHeight = promise.CreatedHeight
	early.ResolutionHash = ComputeAVMPromiseResolutionHash(early)
	_, err = ApplyAVMPromiseResolution(promise, early)
	require.ErrorContains(t, err, "future height")

	timeoutMsg := mustAVMPromiseMessage(t, AVMPayloadTypePromiseTimeout, promise.ExpiryHeight, promise.ExpiryHeight+5)
	timeout, err := ScheduleAVMPromiseTimeout(promise, timeoutMsg)
	require.NoError(t, err)
	require.Equal(t, promise.ExpiryHeight, timeout.DueHeight)
	require.Equal(t, promise.Contract, timeout.Contract)

	badTimeout := timeout
	badTimeout.TimeoutMessage.CreatedHeight = promise.ExpiryHeight - 1
	badTimeout.TimeoutMessage.ID = DeriveAVMAsyncMessageID(badTimeout.TimeoutMessage)
	badTimeout.TaskHash = ComputeAVMPromiseTimeoutTaskHash(badTimeout)
	require.ErrorContains(t, badTimeout.Validate(), "before due height")
}

func TestAVMABIIntrospectionBindsToCodeHashAndValidatesCalls(t *testing.T) {
	descriptor := mustAVMABIIntrospection(t, 9)
	code, err := NewAVMCodeRecord(AVMCodeRecord{
		CodeID:			descriptor.CodeID,
		CodeHash:		descriptor.CodeHash,
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		ABIHash:		descriptor.InterfaceHash,
		Deployer:		"deployer-a",
		CreatedAtHeight:	10,
		CodeBytesRef:		"store:v2/code/code_id/00000000000000000009",
		MeteringProfile:	AVMMeteringProfileDefault,
		Enabled:		true,
	})
	require.NoError(t, err)
	require.NoError(t, BindAVMABIToCode(code, descriptor))

	call, err := NewAVMABIMethodCall(AVMABIMethodCall{
		InterfaceHash:		descriptor.InterfaceHash,
		MethodSelector:		"transfer",
		ArgumentEncoding:	"protobuf",
		ArgumentHash:		ComputeAVMBytesHash([]byte("args")),
		Funds:			[]AVMCallFund{{Denom: "naet", Amount: sdkmath.NewInt(10)}},
	})
	require.NoError(t, err)
	require.NoError(t, ValidateAVMABIMethodCall(descriptor, call))

	badSelector := call
	badSelector.MethodSelector = "missing"
	badSelector.CallHash = ComputeAVMABIMethodCallHash(badSelector)
	require.ErrorContains(t, ValidateAVMABIMethodCall(descriptor, badSelector), "selector")

	badEncoding := call
	badEncoding.ArgumentEncoding = "json"
	badEncoding.CallHash = ComputeAVMABIMethodCallHash(badEncoding)
	require.ErrorContains(t, ValidateAVMABIMethodCall(descriptor, badEncoding), "encoding")

	underfunded := call
	underfunded.Funds = []AVMCallFund{{Denom: "naet", Amount: sdkmath.NewInt(1)}}
	underfunded.CallHash = ComputeAVMABIMethodCallHash(underfunded)
	require.ErrorContains(t, ValidateAVMABIMethodCall(descriptor, underfunded), "below requirement")
}

func TestAVMABIIdentityBindingIsProofCommitted(t *testing.T) {
	descriptor := mustAVMABIIntrospection(t, 11)
	binding, err := NewAVMABIIdentityBinding(AVMABIIdentityBinding{
		Name:			"wallet.aet",
		InterfaceHash:		descriptor.InterfaceHash,
		ResolverRecordHash:	ComputeAVMBytesHash([]byte("resolver-record")),
	})
	require.NoError(t, err)
	require.NoError(t, binding.Validate())
	require.Equal(t, ComputeAVMABIIdentityBindingHash(binding), binding.BindingHash)

	bad := binding
	bad.Name = "wallet"
	bad.BindingHash = ComputeAVMABIIdentityBindingHash(bad)
	require.ErrorContains(t, bad.Validate(), ".aet")
}

func mustAVMAsyncCallMessage(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:		"aetra-local",
		Source:			"contract-a",
		Destination:		"contract-b",
		Payload:		[]byte("remote-call"),
		GasLimit:		100,
		ExpiryHeight:		20,
		RetryPolicy:		DefaultAVMRetryPolicy(20),
		BounceFlag:		true,
		SourceZone:		zonestypes.ZoneID("CONTRACT"),
		DestinationZone:	zonestypes.ZoneID("IDENTITY"),
		SenderNonce:		8,
		PayloadType:		"contract.async_call",
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		1,
		CreatedHeight:		10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVMPromiseMessage(t *testing.T, payloadType string, createdHeight, expiryHeight uint64) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:		"aetra-local",
		Source:			"contract-b",
		Destination:		"contract-a",
		Payload:		[]byte(payloadType),
		GasLimit:		100,
		ExpiryHeight:		expiryHeight,
		RetryPolicy:		DefaultAVMRetryPolicy(expiryHeight),
		BounceFlag:		true,
		SourceZone:		zonestypes.ZoneID("IDENTITY"),
		DestinationZone:	zonestypes.ZoneID("CONTRACT"),
		SenderNonce:		createdHeight,
		PayloadType:		payloadType,
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		1,
		CreatedHeight:		createdHeight,
	})
	require.NoError(t, err)
	return msg
}

func mustAVMABIIntrospection(t *testing.T, codeID uint64) AVMABIIntrospectionDescriptor {
	t.Helper()
	method := AVM2MethodDescriptor{
		Selector:		"transfer",
		Name:			"transfer",
		ArgumentSchemaHash:	ComputeAVMBytesHash([]byte("transfer-args")),
		ReturnSchemaHash:	ComputeAVMBytesHash([]byte("transfer-return")),
		ArgumentEncoding:	"protobuf",
		GasHint:		100,
	}
	method.MethodHash = ComputeAVM2MethodDescriptorHash(method)
	event := AVM2EventDescriptor{
		Name:		"sent",
		SchemaHash:	ComputeAVMBytesHash([]byte("sent-schema")),
	}
	event.EventHash = ComputeAVM2EventDescriptorHash(event)
	errDesc := AVMErrorDescriptor{
		Code:		"insufficient_funds",
		SchemaHash:	ComputeAVMBytesHash([]byte("error-schema")),
	}
	errDesc.ErrorHash = ComputeAVMErrorDescriptorHash(errDesc)
	fund := AVMFundRequirement{
		Selector:	"transfer",
		Denom:		"naet",
		Minimum:	sdkmath.NewInt(5),
	}
	fund.RequirementHash = ComputeAVMFundRequirementHash(fund)
	hint := AVMGasHint{
		Selector:	"transfer",
		Estimate:	100,
	}
	hint.HintHash = ComputeAVMGasHintHash(hint)
	descriptor, err := NewAVMABIIntrospectionDescriptor(AVMABIIntrospectionDescriptor{
		ABIVersion:		1,
		CodeID:			codeID,
		CodeHash:		ComputeAVMBytesHash([]byte(AVMContractCodeStateKey(codeID))),
		Methods:		[]AVM2MethodDescriptor{method},
		Events:			[]AVM2EventDescriptor{event},
		Errors:			[]AVMErrorDescriptor{errDesc},
		RequiredFunds:		[]AVMFundRequirement{fund},
		GasHints:		[]AVMGasHint{hint},
		IdentityNameOptional:	"wallet.aet",
	})
	require.NoError(t, err)
	return descriptor
}
