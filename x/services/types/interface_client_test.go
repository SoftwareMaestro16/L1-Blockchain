package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestInterfaceDrivenUXFlowRequiresVerificationDisplayAndConfirmation(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()

	flow, err := NewInterfaceDrivenUXFlow(descriptor, "query")
	require.NoError(t, err)
	require.Equal(t, descriptor.ServiceID, flow.ServiceID)
	require.Equal(t, descriptor.Interface.InterfaceHash, flow.InterfaceHash)
	require.Equal(t, []InterfaceUXStep{
		InterfaceUXUserInput,
		InterfaceUXResolveService,
		InterfaceUXFetchInterface,
		InterfaceUXVerifyInterfaceHash,
		InterfaceUXGenerateForm,
		InterfaceUXBuildCall,
		InterfaceUXExecuteCall,
		InterfaceUXVerifyReceipt,
	}, flow.Steps)
	require.True(t, flow.DisplayPaymentAndTrustModel)
	require.True(t, flow.RequireUserSigningConfirmation)
	require.False(t, flow.MetadataGrantsAuthorization)
	require.NotEmpty(t, flow.PaymentModel)
	require.Equal(t, descriptor.Verification.TrustModel, flow.TrustModel)
	require.Equal(t, coretypes.ServiceVerificationConsensusReceipt, flow.VerificationModel)
	require.NoError(t, flow.Validate())

	bad := flow
	bad.RequireUserSigningConfirmation = false
	bad.FlowHash = ComputeInterfaceDrivenUXFlowHash(bad)
	require.ErrorContains(t, bad.Validate(), "confirmation")

	metadataAuth := flow
	metadataAuth.MetadataGrantsAuthorization = true
	metadataAuth.FlowHash = ComputeInterfaceDrivenUXFlowHash(metadataAuth)
	require.ErrorContains(t, metadataAuth.Validate(), "metadata")
}

func TestInterfaceRegistrationUpdateMessagesAndProofQuery(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	schema, err := NewInterfaceSchemaFormat(descriptor.Interface, "query")
	require.NoError(t, err)

	register, err := NewMsgRegisterInterface(coretypes.DefaultAuthority, descriptor.Interface, schema)
	require.NoError(t, err)
	require.Equal(t, "MsgRegisterInterface", register.ServiceRegistryMessageName())
	require.Equal(t, coretypes.DefaultAuthority, register.ServiceRegistrySigner())
	require.Equal(t, ComputeInterfaceRegistryMessageHash(register), register.MessageHash)
	require.NoError(t, register.ValidateBasic())

	next := descriptor.Interface
	next.Version = 2
	next.InterfaceName = "l1.services.v2.Portable"
	next.MetadataHash = testInterfaceHash("interface/metadata/v2")
	next.InterfaceHash = coretypes.ComputeServiceInterfaceHash(next)
	nextSchema, err := NewInterfaceSchemaFormat(next, "query")
	require.NoError(t, err)

	update, err := NewMsgUpdateInterface(coretypes.DefaultAuthority, descriptor.Interface.InterfaceHash, next, nextSchema, 1)
	require.NoError(t, err)
	require.Equal(t, "MsgUpdateInterface", update.ServiceRegistryMessageName())
	require.Equal(t, ComputeInterfaceRegistryMessageHash(update), update.MessageHash)
	require.NoError(t, update.ValidateBasic())

	anchor, err := coretypes.NewServiceAnchorFromDescriptor(descriptor)
	require.NoError(t, err)
	state, err := coretypes.NewServiceRegistryState(
		[]coretypes.ServiceDescriptor{descriptor},
		[]coretypes.ServiceAnchor{anchor},
		nil,
		nil,
		nil,
		nil,
		20,
	)
	require.NoError(t, err)

	proof, err := QueryInterfaceProofFromState(state, QueryInterfaceProof{InterfaceHash: descriptor.Interface.InterfaceHash})
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.Equal(t, descriptor.Interface.InterfaceHash, proof.Proof.InterfaceHash)
	require.Equal(t, state.StateRoot, proof.Proof.RegistryRoot)
	require.Equal(t, uint64(20), proof.Proof.ProofHeight)
	require.NoError(t, proof.Proof.Validate())

	missing, err := QueryInterfaceProofFromState(state, QueryInterfaceProof{InterfaceHash: testInterfaceHash("missing/interface")})
	require.NoError(t, err)
	require.False(t, missing.Found)
}

func TestSDKVerifierCallBuilderAndWalletCLISchema(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()

	verification, err := VerifySDKInterface(descriptor, descriptor.Interface.InterfaceHash)
	require.NoError(t, err)
	require.Equal(t, descriptor.ServiceID, verification.ServiceID)
	require.Equal(t, descriptor.Interface.InterfaceHash, verification.InterfaceHash)
	require.NotEmpty(t, verification.PaymentModel)
	require.Equal(t, descriptor.Verification.TrustModel, verification.TrustModel)
	require.NoError(t, verification.Validate())

	_, err = VerifySDKInterface(descriptor, testInterfaceHash("wrong/interface"))
	require.ErrorContains(t, err, "mismatch")

	_, err = BuildMethodCallFromUserInput(descriptor, "submit", coretypes.DefaultAuthority, 1, testInterfaceHash("payload/submit"), 21, false)
	require.ErrorContains(t, err, "confirmation")

	call, err := BuildMethodCallFromUserInput(descriptor, "submit", coretypes.DefaultAuthority, 1, testInterfaceHash("payload/submit"), 21, true)
	require.NoError(t, err)
	require.Equal(t, "submit", call.MethodName)
	require.Equal(t, coretypes.ServiceMethodAsync, call.ExecutionType)
	require.NoError(t, call.Validate())

	schema, err := BuildWalletCLISchema(descriptor, "submit")
	require.NoError(t, err)
	require.Equal(t, WalletSchemaFormatJSONV1, schema.WalletFormat)
	require.Equal(t, CLISchemaFormatJSONV1, schema.CLIFormat)
	require.Equal(t, "submit", schema.MethodName)
	require.NotEmpty(t, schema.PaymentModel)
	require.Equal(t, descriptor.Verification.TrustModel, schema.TrustModel)
	require.Equal(t, coretypes.ServiceVerificationSignedResult, schema.Verification)
	require.NoError(t, schema.Validate())
}

func TestVersionedInterfaceCompatibilityDetectsBreakingChanges(t *testing.T) {
	previous := testInterfaceSystemDescriptor().Interface
	compatible := previous
	compatible.Version = 2
	compatible.InterfaceName = "l1.services.v2.Portable"
	compatible.Methods = append([]coretypes.ServiceMethodDescriptor(nil), compatible.Methods...)
	compatible.Methods = append(compatible.Methods, testInterfaceMethod("status", coretypes.ServiceMethodSync, coretypes.ServiceVerificationConsensusReceipt, coretypes.DefaultGasPolicy))
	compatible = coretypes.CanonicalServiceInterfaceDescriptor(compatible)
	compatible.InterfaceHash = coretypes.ComputeServiceInterfaceHash(compatible)

	report, err := CheckVersionedInterfaceCompatibility(previous, compatible)
	require.NoError(t, err)
	require.True(t, report.Compatible)
	require.Empty(t, report.BreakingChanges)
	require.Equal(t, []string{"status"}, report.AddedMethods)
	require.NoError(t, report.Validate())

	breaking := previous
	breaking.Version = 2
	breaking.InterfaceName = "l1.services.v2.Portable"
	breaking.Methods = append([]coretypes.ServiceMethodDescriptor(nil), breaking.Methods[:2]...)
	breaking.Methods[0].InputSchemaHash = testInterfaceHash("query/input/v2")
	breaking = coretypes.CanonicalServiceInterfaceDescriptor(breaking)
	breaking.InterfaceHash = coretypes.ComputeServiceInterfaceHash(breaking)

	breakingReport, err := CheckVersionedInterfaceCompatibility(previous, breaking)
	require.NoError(t, err)
	require.False(t, breakingReport.Compatible)
	require.Contains(t, breakingReport.BreakingChanges, "changed input schema query")
	require.Contains(t, breakingReport.BreakingChanges, "removed method submit")
	require.NoError(t, breakingReport.Validate())
}
