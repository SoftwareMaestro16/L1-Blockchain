package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVM2AsyncCallPlanIsMessageDrivenAndNonBlocking(t *testing.T) {
	msg := mustAVM2AsyncCallMessage(t)
	promise := mustAVM2Promise(t, msg.ID)
	plan, err := NewAVM2AsyncCallPlan(AVM2AsyncCallPlan{
		Height:             msg.CreatedHeight,
		CallerContract:     "contract-a",
		Message:            msg,
		Promise:            promise,
		AwaitNonBlocking:   true,
		PersistedPromiseID: promise.PromiseID,
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeAVM2MessageRoot([]AVMAsyncMessage{msg}), plan.OutboxRoot)
	require.Equal(t, ComputeAVM2PromiseRoot([]AVM2PromiseState{promise}), plan.PromiseRoot)
	require.Equal(t, promise.PromiseID, plan.PersistedPromiseID)

	blocking := plan
	blocking.AwaitNonBlocking = false
	blocking.PlanHash = ComputeAVM2AsyncCallPlanHash(blocking)
	require.ErrorContains(t, blocking.Validate(), "non-blocking")

	local := plan
	local.Message.SourceZone = local.Message.DestinationZone
	local.Message.ID = DeriveAVMAsyncMessageID(local.Message)
	local.Promise.MessageID = local.Message.ID
	local.Promise.PromiseHash = ComputeAVM2PromiseHash(local.Promise)
	local.OutboxRoot = ComputeAVM2MessageRoot([]AVMAsyncMessage{local.Message})
	local.PromiseRoot = ComputeAVM2PromiseRoot([]AVM2PromiseState{local.Promise})
	local.PlanHash = ComputeAVM2AsyncCallPlanHash(local)
	require.ErrorContains(t, local.Validate(), "cross-zone")
}

func TestAVM2PromiseResolutionAndTimeoutAreFutureMessages(t *testing.T) {
	msg := mustAVM2AsyncCallMessage(t)
	promise := mustAVM2Promise(t, msg.ID)
	resolutionMsg := mustAVM2PromiseMessage(t, AVM2PayloadTypePromiseResolution, 13, 20)
	resolution, err := NewAVM2PromiseResolution(AVM2PromiseResolution{
		PromiseID:         promise.PromiseID,
		OriginalMessageID: promise.MessageID,
		ResolutionMessage: resolutionMsg,
		Status:            AVM2PromiseResolved,
		ReceiptHash:       ComputeAVM2BytesHash([]byte("receipt")),
		ReturnHash:        ComputeAVM2BytesHash([]byte("return")),
		DeliveryHeight:    13,
	})
	require.NoError(t, err)
	updated, err := ApplyAVM2PromiseResolution(promise, resolution)
	require.NoError(t, err)
	require.Equal(t, AVM2PromiseResolved, updated.Status)
	require.Equal(t, resolution.ReceiptHash, updated.ReceiptHash)

	early := resolution
	early.DeliveryHeight = promise.CreatedHeight
	early.ResolutionHash = ComputeAVM2PromiseResolutionHash(early)
	_, err = ApplyAVM2PromiseResolution(promise, early)
	require.ErrorContains(t, err, "future height")

	timeoutMsg := mustAVM2PromiseMessage(t, AVM2PayloadTypePromiseTimeout, promise.ExpiryHeight, promise.ExpiryHeight+5)
	timeout, err := ScheduleAVM2PromiseTimeout(promise, timeoutMsg)
	require.NoError(t, err)
	require.Equal(t, promise.ExpiryHeight, timeout.DueHeight)
	require.Equal(t, promise.Contract, timeout.Contract)

	badTimeout := timeout
	badTimeout.TimeoutMessage.CreatedHeight = promise.ExpiryHeight - 1
	badTimeout.TimeoutMessage.ID = DeriveAVMAsyncMessageID(badTimeout.TimeoutMessage)
	badTimeout.TaskHash = ComputeAVM2PromiseTimeoutTaskHash(badTimeout)
	require.ErrorContains(t, badTimeout.Validate(), "before due height")
}

func TestAVM2ABIIntrospectionBindsToCodeHashAndValidatesCalls(t *testing.T) {
	descriptor := mustAVM2ABIIntrospection(t, 9)
	code, err := NewAVM2CodeRecord(AVM2CodeRecord{
		CodeID:                descriptor.CodeID,
		CodeHash:              descriptor.CodeHash,
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		ABIHash:               descriptor.InterfaceHash,
		Deployer:              "deployer-a",
		CreatedAtHeight:       10,
		CodeBytesRef:          "store:v2/code/code_id/00000000000000000009",
		MeteringProfile:       AVM2MeteringProfileDefault,
		Enabled:               true,
	})
	require.NoError(t, err)
	require.NoError(t, BindAVM2ABIToCode(code, descriptor))

	call, err := NewAVM2ABIMethodCall(AVM2ABIMethodCall{
		InterfaceHash:    descriptor.InterfaceHash,
		MethodSelector:   "transfer",
		ArgumentEncoding: "protobuf",
		ArgumentHash:     ComputeAVM2BytesHash([]byte("args")),
		Funds:            []AVM2CallFund{{Denom: "naet", Amount: sdkmath.NewInt(10)}},
	})
	require.NoError(t, err)
	require.NoError(t, ValidateAVM2ABIMethodCall(descriptor, call))

	badSelector := call
	badSelector.MethodSelector = "missing"
	badSelector.CallHash = ComputeAVM2ABIMethodCallHash(badSelector)
	require.ErrorContains(t, ValidateAVM2ABIMethodCall(descriptor, badSelector), "selector")

	badEncoding := call
	badEncoding.ArgumentEncoding = "json"
	badEncoding.CallHash = ComputeAVM2ABIMethodCallHash(badEncoding)
	require.ErrorContains(t, ValidateAVM2ABIMethodCall(descriptor, badEncoding), "encoding")

	underfunded := call
	underfunded.Funds = []AVM2CallFund{{Denom: "naet", Amount: sdkmath.NewInt(1)}}
	underfunded.CallHash = ComputeAVM2ABIMethodCallHash(underfunded)
	require.ErrorContains(t, ValidateAVM2ABIMethodCall(descriptor, underfunded), "below requirement")
}

func TestAVM2ABIIdentityBindingIsProofCommitted(t *testing.T) {
	descriptor := mustAVM2ABIIntrospection(t, 11)
	binding, err := NewAVM2ABIIdentityBinding(AVM2ABIIdentityBinding{
		Name:               "wallet.aet",
		InterfaceHash:      descriptor.InterfaceHash,
		ResolverRecordHash: ComputeAVM2BytesHash([]byte("resolver-record")),
	})
	require.NoError(t, err)
	require.NoError(t, binding.Validate())
	require.Equal(t, ComputeAVM2ABIIdentityBindingHash(binding), binding.BindingHash)

	bad := binding
	bad.Name = "wallet"
	bad.BindingHash = ComputeAVM2ABIIdentityBindingHash(bad)
	require.ErrorContains(t, bad.Validate(), ".aet")
}

func mustAVM2AsyncCallMessage(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:         "aetra-local",
		Source:          "contract-a",
		Destination:     "contract-b",
		Payload:         []byte("remote-call"),
		GasLimit:        100,
		ExpiryHeight:    20,
		RetryPolicy:     DefaultAVMRetryPolicy(20),
		BounceFlag:      true,
		SourceZone:      zonestypes.ZoneID("CONTRACT"),
		DestinationZone: zonestypes.ZoneID("IDENTITY"),
		SenderNonce:     8,
		PayloadType:     "contract.async_call",
		ValueNAET:       1,
		ForwardingFee:   1,
		Priority:        1,
		CreatedHeight:   10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVM2PromiseMessage(t *testing.T, payloadType string, createdHeight, expiryHeight uint64) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:         "aetra-local",
		Source:          "contract-b",
		Destination:     "contract-a",
		Payload:         []byte(payloadType),
		GasLimit:        100,
		ExpiryHeight:    expiryHeight,
		RetryPolicy:     DefaultAVMRetryPolicy(expiryHeight),
		BounceFlag:      true,
		SourceZone:      zonestypes.ZoneID("IDENTITY"),
		DestinationZone: zonestypes.ZoneID("CONTRACT"),
		SenderNonce:     createdHeight,
		PayloadType:     payloadType,
		ValueNAET:       1,
		ForwardingFee:   1,
		Priority:        1,
		CreatedHeight:   createdHeight,
	})
	require.NoError(t, err)
	return msg
}

func mustAVM2ABIIntrospection(t *testing.T, codeID uint64) AVM2ABIIntrospectionDescriptor {
	t.Helper()
	method := AVM2MethodDescriptor{
		Selector:           "transfer",
		Name:               "transfer",
		ArgumentSchemaHash: ComputeAVM2BytesHash([]byte("transfer-args")),
		ReturnSchemaHash:   ComputeAVM2BytesHash([]byte("transfer-return")),
		ArgumentEncoding:   "protobuf",
		GasHint:            100,
	}
	method.MethodHash = ComputeAVM2MethodDescriptorHash(method)
	event := AVM2EventDescriptor{
		Name:       "sent",
		SchemaHash: ComputeAVM2BytesHash([]byte("sent-schema")),
	}
	event.EventHash = ComputeAVM2EventDescriptorHash(event)
	errDesc := AVM2ErrorDescriptor{
		Code:       "insufficient_funds",
		SchemaHash: ComputeAVM2BytesHash([]byte("error-schema")),
	}
	errDesc.ErrorHash = ComputeAVM2ErrorDescriptorHash(errDesc)
	fund := AVM2FundRequirement{
		Selector: "transfer",
		Denom:    "naet",
		Minimum:  sdkmath.NewInt(5),
	}
	fund.RequirementHash = ComputeAVM2FundRequirementHash(fund)
	hint := AVM2GasHint{
		Selector: "transfer",
		Estimate: 100,
	}
	hint.HintHash = ComputeAVM2GasHintHash(hint)
	descriptor, err := NewAVM2ABIIntrospectionDescriptor(AVM2ABIIntrospectionDescriptor{
		ABIVersion:           1,
		CodeID:               codeID,
		CodeHash:             ComputeAVM2BytesHash([]byte(AVM2ContractCodeStateKey(codeID))),
		Methods:              []AVM2MethodDescriptor{method},
		Events:               []AVM2EventDescriptor{event},
		Errors:               []AVM2ErrorDescriptor{errDesc},
		RequiredFunds:        []AVM2FundRequirement{fund},
		GasHints:             []AVM2GasHint{hint},
		IdentityNameOptional: "wallet.aet",
	})
	require.NoError(t, err)
	return descriptor
}
