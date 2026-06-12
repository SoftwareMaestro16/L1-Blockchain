package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDefaultXServiceInterfaceModuleBreakdownCoversSection152(t *testing.T) {
	breakdown, err := DefaultXServiceInterfaceModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModuleInterface, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServiceInterfaceStateInterface)
	require.Contains(t, breakdown.StateObjects, XServiceInterfaceStateMethod)
	require.Contains(t, breakdown.StateObjects, XServiceInterfaceStateEvent)
	require.Contains(t, breakdown.StateObjects, XServiceInterfaceStateError)
	require.Contains(t, breakdown.StateObjects, XServiceInterfaceStateInterfaceVersion)

	require.Contains(t, breakdown.Messages, XServiceInterfaceMsgRegisterInterface)
	require.Contains(t, breakdown.Messages, XServiceInterfaceMsgUpdateInterface)
	require.Contains(t, breakdown.Messages, XServiceInterfaceMsgDeprecateInterface)

	require.Contains(t, breakdown.Queries, XServiceInterfaceQueryInterface)
	require.Contains(t, breakdown.Queries, XServiceInterfaceQueryMethod)
	require.Contains(t, breakdown.Queries, XServiceInterfaceQueryInterfaceProof)
	require.Contains(t, breakdown.Queries, XServiceInterfaceQueryInterfacesByOwner)

	require.Contains(t, breakdown.IntegrationPoints, XServiceInterfaceIntegrationServices)
	require.Contains(t, breakdown.IntegrationPoints, XServiceInterfaceIntegrationWalletSDK)
	require.Contains(t, breakdown.IntegrationPoints, XServiceInterfaceIntegrationCLI)
	require.Contains(t, breakdown.IntegrationPoints, XServiceInterfaceIntegrationContractAdapter)
}

func TestXServiceInterfaceModuleBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServiceInterfaceModuleBreakdown()
	require.NoError(t, err)
	breakdown.Queries = removeXServiceInterfaceQueryForTest(breakdown.Queries, XServiceInterfaceQueryMethod)
	breakdown.BreakdownHash = ComputeXServiceInterfaceModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "query")

	breakdown, err = DefaultXServiceInterfaceModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServiceInterfaceMessageForTest(breakdown.Messages, XServiceInterfaceMsgDeprecateInterface)
	breakdown.BreakdownHash = ComputeXServiceInterfaceModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")
}

func TestXServiceInterfaceFailureModesUseExecutableGuards(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	schema, err := NewInterfaceSchemaFormat(descriptor.Interface, "query")
	require.NoError(t, err)

	badSchema := schema
	badSchema.InterfaceHash = testInterfaceHash("x-serviceinterface/wrong-schema-interface")
	badSchema.FormatHash = ComputeInterfaceSchemaFormatHash(badSchema)
	_, err = NewMsgRegisterInterface(coretypes.DefaultAuthority, descriptor.Interface, badSchema)
	require.ErrorContains(t, err, "schema hash mismatch")

	collision := descriptor.Interface
	collision.Methods = append([]coretypes.ServiceMethodDescriptor(nil), collision.Methods...)
	collision.Methods[1].MethodID = collision.Methods[0].MethodID
	collision.InterfaceHash = coretypes.ComputeServiceInterfaceHash(collision)
	_, err = NewFormalServiceInterface(collision)
	require.ErrorContains(t, err, "duplicate")

	badEncoding := descriptor.Interface
	badEncoding.SchemaEncoding = "yaml-v1"
	badEncoding.InterfaceHash = coretypes.ComputeServiceInterfaceHash(badEncoding)
	_, err = NewFormalServiceInterface(badEncoding)
	require.ErrorContains(t, err, "not supported")

	breakingSameHash := descriptor.Interface
	breakingSameHash.Version = 2
	err = ValidateServiceInterfaceVersionChange(descriptor.Interface, breakingSameHash)
	require.ErrorContains(t, err, "new interface hash")
}

func TestXServiceInterfaceDeprecateFlow(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testServiceInterfaceRegistryState(t, descriptor)
	marker, err := NewServiceInterfaceDeprecationMarker(ServiceInterfaceDeprecationMarker{
		InterfaceHash:			descriptor.Interface.InterfaceHash,
		Version:			descriptor.Interface.Version,
		DeprecatedHeight:		50,
		RetirementHeight:		100,
		ReplacementInterfaceHash:	testInterfaceHash("x-serviceinterface/replacement"),
		Reason:				"replace_with_v2",
	})
	require.NoError(t, err)
	msg, err := NewMsgDeprecateInterface(coretypes.DefaultAuthority, marker)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgDeprecateInterfaceHash(msg), msg.MsgHash)
	require.NoError(t, msg.ValidateBasic())

	applied, err := DeprecateInterfaceInState(state, msg, 45)
	require.NoError(t, err)
	require.Equal(t, marker.MarkerHash, applied.MarkerHash)

	_, err = DeprecateInterfaceInState(state, msg, 55)
	require.ErrorContains(t, err, "after deprecation height")

	missing := marker
	missing.InterfaceHash = testInterfaceHash("x-serviceinterface/missing")
	missing.MarkerHash = ComputeServiceInterfaceDeprecationMarkerHash(missing)
	missingMsg, err := NewMsgDeprecateInterface(coretypes.DefaultAuthority, missing)
	require.NoError(t, err)
	_, err = DeprecateInterfaceInState(state, missingMsg, 45)
	require.ErrorContains(t, err, "not found")
}

func TestXServiceInterfaceQueriesMethodProofAndOwner(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testServiceInterfaceRegistryState(t, descriptor)
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	require.NoError(t, err)

	method, err := QueryMethodFromInterface(definition, QueryMethod{
		InterfaceHash:	descriptor.Interface.InterfaceHash,
		MethodID:	"submit",
	})
	require.NoError(t, err)
	require.True(t, method.Found)
	require.Equal(t, "submit", method.Method.MethodID)

	missing, err := QueryMethodFromInterface(definition, QueryMethod{
		InterfaceHash:	descriptor.Interface.InterfaceHash,
		MethodID:	"missing",
	})
	require.NoError(t, err)
	require.False(t, missing.Found)

	proof, err := QueryInterfaceProofFromState(state, QueryInterfaceProof{InterfaceHash: descriptor.Interface.InterfaceHash})
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.NoError(t, proof.Proof.Validate())

	owner, err := QueryInterfacesByOwnerFromServiceRegistry(state, descriptor.Owner)
	require.NoError(t, err)
	require.Equal(t, uint64(1), owner.Total)
	require.Equal(t, descriptor.Interface.InterfaceHash, owner.Interfaces[0].InterfaceHash)
	require.NoError(t, owner.Validate())

	empty, err := QueryInterfacesByOwnerFromServiceRegistry(state, "other.owner")
	require.NoError(t, err)
	require.Equal(t, uint64(0), empty.Total)
	require.Empty(t, empty.Interfaces)
	require.NoError(t, empty.Validate())
}

func TestXServiceInterfaceWalletCLIAndContractAdapterIntegrations(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()

	schema, err := BuildWalletCLISchema(descriptor, "submit")
	require.NoError(t, err)
	require.Equal(t, WalletSchemaFormatJSONV1, schema.WalletFormat)
	require.Equal(t, CLISchemaFormatJSONV1, schema.CLIFormat)
	require.NoError(t, schema.Validate())

	adapter, err := NewServiceInterfaceContractAdapterSchema(
		descriptor.Interface.InterfaceHash,
		"submit",
		testInterfaceHash("x-serviceinterface/contract-abi"),
		"cosmwasm-json",
	)
	require.NoError(t, err)
	require.NoError(t, adapter.Validate())
	require.NotEmpty(t, adapter.AdapterHash)

	_, err = NewServiceInterfaceContractAdapterSchema(
		descriptor.Interface.InterfaceHash,
		"submit",
		testInterfaceHash("x-serviceinterface/contract-abi"),
		"yaml",
	)
	require.ErrorContains(t, err, "not supported")
}

func testServiceInterfaceRegistryState(t *testing.T, descriptor coretypes.ServiceDescriptor) ServiceRegistryState {
	t.Helper()
	anchor, err := coretypes.NewServiceAnchorFromDescriptor(descriptor)
	require.NoError(t, err)
	state, err := coretypes.NewServiceRegistryState(
		[]coretypes.ServiceDescriptor{descriptor},
		[]coretypes.ServiceAnchor{anchor},
		nil,
		nil,
		nil,
		nil,
		40,
	)
	require.NoError(t, err)
	return state
}

func removeXServiceInterfaceMessageForTest(messages []XServiceInterfaceMessageName, target XServiceInterfaceMessageName) []XServiceInterfaceMessageName {
	out := make([]XServiceInterfaceMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}

func removeXServiceInterfaceQueryForTest(queries []XServiceInterfaceQueryName, target XServiceInterfaceQueryName) []XServiceInterfaceQueryName {
	out := make([]XServiceInterfaceQueryName, 0, len(queries))
	for _, query := range queries {
		if query != target {
			out = append(out, query)
		}
	}
	return out
}
