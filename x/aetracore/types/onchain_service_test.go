package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildModuleMsgServiceDescriptor(t *testing.T) {
	wrapper := testOnChainWrapper()
	descriptor, err := BuildModuleMsgServiceDescriptor(wrapper)
	require.NoError(t, err)

	require.Equal(t, ServiceTypeOnChain, descriptor.ServiceType)
	require.Equal(t, ServiceLocationModule, descriptor.Execution.Location)
	require.Equal(t, "dex", descriptor.Execution.ModuleRoute)
	require.True(t, descriptor.Execution.Deterministic)
	require.Equal(t, ServiceReceiptCommittedAndProof, descriptor.Execution.ReceiptPolicy)
	require.Equal(t, ServiceStorageOnChain, descriptor.Storage.Model)
	require.Equal(t, RootType(StateProofRootType), descriptor.Storage.StateRootType)
	require.True(t, descriptor.Storage.ProofRequired)
	require.Equal(t, ServiceTrustConsensusExecuted, descriptor.Verification.TrustModel)
	require.Equal(t, ServiceVerificationConsensusReceipt, descriptor.Verification.Model)
	require.Len(t, descriptor.Interface.Methods, 2)
	require.Equal(t, "quote", descriptor.Interface.Methods[0].MethodID)
	require.Equal(t, DefaultGasPolicy, descriptor.Interface.Methods[0].GasModel)
	require.NoError(t, descriptor.Validate())

	query, err := NewServiceStateProofQuery(descriptor, 9, testHash("dex/state/key"), testHash("dex/state/root"))
	require.NoError(t, err)
	require.Equal(t, descriptor.ServiceID, query.ServiceID)
	require.NotEmpty(t, query.QueryHash)
	require.NoError(t, query.Validate())
}

func TestBuildContractServiceDescriptor(t *testing.T) {
	wrapper := testOnChainWrapper()
	wrapper.ServiceID = "identity-contract-resolver"
	wrapper.ZoneID = ZoneIDContract
	wrapper.ModuleName = ""
	wrapper.ContractAddress = "contract.identity.resolver.v1"
	wrapper.EndpointKey = "contract.identity.resolver"
	wrapper.AvailabilityHash = testHash("contract/availability")

	descriptor, err := BuildContractServiceDescriptor(wrapper)
	require.NoError(t, err)

	require.Equal(t, ServiceLocationContract, descriptor.Execution.Location)
	require.Empty(t, descriptor.Execution.ModuleRoute)
	require.Equal(t, wrapper.ContractAddress, descriptor.Execution.ContractAddress)
	require.Equal(t, wrapper.ContractAddress, descriptor.Execution.Target)
	require.NoError(t, descriptor.Validate())
}

func TestOnChainServiceWrapperRejectsInvalidTargetOrGasModel(t *testing.T) {
	wrapper := testOnChainWrapper()
	wrapper.ModuleName = ""
	wrapper.ContractAddress = ""
	_, err := BuildOnChainServiceDescriptor(wrapper)
	require.ErrorContains(t, err, "requires module name or contract address")

	wrapper = testOnChainWrapper()
	wrapper.Methods[0].GasModel = ""
	_, err = BuildModuleMsgServiceDescriptor(wrapper)
	require.ErrorContains(t, err, "gas model")
}

func TestNewOnChainServiceReceiptCommitsModuleExecution(t *testing.T) {
	state := EmptyState(TestnetParams())
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	descriptor, err := BuildModuleMsgServiceDescriptor(testOnChainWrapper())
	require.NoError(t, err)
	state, err = RegisterServiceDescriptor(state, descriptor)
	require.NoError(t, err)

	ctx := ServiceConsensusContext{ChainID: "aetra-testnet", Height: 12}
	call := ServiceCallEnvelope{
		ServiceID:		descriptor.ServiceID,
		Caller:			DefaultAuthority,
		Nonce:			1,
		MethodID:		"swap",
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		PayloadHash:		testHash("dex/swap/payload"),
		PaymentDenom:		NativeFeePolicyID,
		MaxFeeAmount:		"10",
		ProofRequirement:	ServiceVerificationConsensusReceipt,
		Kind:			ServiceCallKindOnChain,
		CreatedHeight:		11,
		DeadlineHeight:		14,
		StateReadSet:		[]string{"dex/pool/atom-naet"},
		StateWriteSet:		[]string{"dex/account/settlement"},
	}
	result := ExecutionResult{Success: true, ResultHash: testHash("dex/swap/result")}

	receipt, err := NewOnChainServiceReceipt(ctx, state, call, result, 17_000)
	require.NoError(t, err)
	require.Equal(t, ServiceCallStatusExecuted, receipt.Status)
	require.Equal(t, ServicePaymentStatusSettled, receipt.PaymentStatus)
	require.Equal(t, uint64(17_000), receipt.GasUsed)
	require.Equal(t, ctx.Height, receipt.ExecutedHeight)
	require.Equal(t, ctx.Height, receipt.AnchoredHeight)
	require.NotEmpty(t, receipt.ReceiptHash)
	require.NoError(t, receipt.Validate())
}

func testOnChainWrapper() OnChainServiceWrapper {
	return OnChainServiceWrapper{
		ServiceID:		"dex-swap",
		Owner:			DefaultAuthority,
		ZoneID:			ZoneIDFinancial,
		ModuleName:		"dex",
		InterfaceID:		"l1.dex.v1.Msg",
		InterfaceName:		"l1.dex.v1.Msg",
		EndpointKey:		"dex.msg",
		Version:		1,
		AvailabilityHash:	testHash("dex/availability"),
		StateRootType:		StateProofRootType,
		ReceiptPolicy:		ServiceReceiptCommittedAndProof,
		PaymentDenom:		NativeFeePolicyID,
		PaymentAmount:		"0",
		MetadataHash:		testHash("dex/metadata"),
		CreatedHeight:		1,
		UpdatedHeight:		1,
		ExpiryHeight:		100,
		Methods: []OnChainServiceMethod{
			testOnChainMethod("swap"),
			testOnChainMethod("quote"),
		},
	}
}

func testOnChainMethod(methodID string) OnChainServiceMethod {
	return OnChainServiceMethod{
		MethodID:		methodID,
		Name:			methodID,
		InputSchemaHash:	testHash(methodID + "/input"),
		OutputSchemaHash:	testHash(methodID + "/output"),
		GasModel:		DefaultGasPolicy,
		RequiredPaymentModel:	DefaultOnChainPaymentModel,
		FailurePolicy:		ServiceFailureRevert,
		TimeoutHeightDelta:	3,
		IdempotencyRequired:	strings.HasPrefix(methodID, "quote"),
	}
}
