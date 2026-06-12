package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMInterfaceDescriptorCommitsMethodsEventsAsyncHandlersAndGets(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)

	require.NoError(t, descriptor.Validate())
	require.Equal(t, ComputeAVMInterfaceHash(descriptor), descriptor.InterfaceHash)
	require.Equal(t, "v1.0.0", descriptor.InterfaceVersion)
	require.Equal(t, AVMInterfaceTargetActor, descriptor.TargetType)
	require.Equal(t, AVMInterfaceSchemaJSONSchema, descriptor.SchemaEncoding)
	require.Len(t, descriptor.MethodDescriptors, 2)
	require.Len(t, descriptor.EventDescriptors, 1)
	require.Len(t, descriptor.AsyncHandlerDescriptors, 1)
	require.Len(t, descriptor.GetMethodDescriptors, 1)
	require.Equal(t, "avm/interfaces/"+descriptor.InterfaceHash, AVMInterfaceDescriptorKey(descriptor.InterfaceHash))

	mutated := descriptor
	mutated.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	mutated.MethodDescriptors[0].GasHint++
	require.NotEqual(t, descriptor.InterfaceHash, ComputeAVMInterfaceHash(mutated))
}

func TestAVMMethodDescriptorSupportsExecutionModesAndOptionalRequirements(t *testing.T) {
	for _, mode := range []AVMInterfaceExecutionMode{
		AVMInterfaceExecutionSync,
		AVMInterfaceExecutionAsync,
		AVMInterfaceExecutionScheduled,
		AVMInterfaceExecutionGet,
	} {
		method := AVMMethodDescriptor{
			MethodID:			"method-" + string(mode),
			Name:				"method_" + string(mode),
			InputSchemaHash:		engineHash("input-" + string(mode)),
			OutputSchemaHash:		engineHash("output-" + string(mode)),
			ExecutionMode:			mode,
			GasHint:			10,
			PaymentRequirementOptional:	"naet",
			ProofRequirementOptional:	"state-proof",
		}
		require.NoError(t, method.Validate())
	}

	badMode := AVMMethodDescriptor{
		MethodID:		"bad",
		Name:			"bad",
		InputSchemaHash:	engineHash("input"),
		OutputSchemaHash:	engineHash("output"),
		ExecutionMode:		AVMInterfaceExecutionMode("streaming"),
		GasHint:		1,
	}
	require.ErrorContains(t, badMode.Validate(), "execution mode")

	noGas := badMode
	noGas.ExecutionMode = AVMInterfaceExecutionSync
	noGas.GasHint = 0
	require.ErrorContains(t, noGas.Validate(), "gas hint")
}

func TestAVMInterfaceRegistryRootAndDuplicateRejection(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)
	second, err := NewAVMInterfaceDescriptor(AVMInterfaceDescriptor{
		InterfaceVersion:	"v1.0.1",
		Owner:			"native-bank",
		TargetType:		AVMInterfaceTargetNativeModule,
		MethodDescriptors: []AVMMethodDescriptor{{
			MethodID:		"bank.send",
			Name:			"MsgSend",
			InputSchemaHash:	engineHash("bank-send-input"),
			OutputSchemaHash:	engineHash("bank-send-output"),
			ExecutionMode:		AVMInterfaceExecutionSync,
			GasHint:		20,
		}},
		SchemaEncoding:		AVMInterfaceSchemaProtobuf,
		MetadataHashOptional:	engineHash("bank-metadata"),
	})
	require.NoError(t, err)

	registry, err := NewAVMInterfaceRegistry(AVMInterfaceRegistry{Interfaces: []AVMInterfaceDescriptor{second, descriptor}})
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, ComputeAVMInterfaceRegistryRoot(registry), registry.Root)

	duplicate := registry
	duplicate.Interfaces = append([]AVMInterfaceDescriptor(nil), registry.Interfaces...)
	duplicate.Interfaces = append(duplicate.Interfaces, descriptor)
	duplicate.Root = ComputeAVMInterfaceRegistryRoot(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")
}

func TestAVMInterfaceSchemaBindingsQueriesAndSDKCodegen(t *testing.T) {
	contractDescriptor := testAVMInterfaceDescriptor(t)
	serviceDescriptor, err := NewAVMInterfaceDescriptor(AVMInterfaceDescriptor{
		InterfaceVersion:	"v2.0.0",
		Owner:			"rpc-service",
		TargetType:		AVMInterfaceTargetService,
		MethodDescriptors: []AVMMethodDescriptor{{
			MethodID:		"rpc.submit",
			Name:			"submit",
			InputSchemaHash:	engineHash("rpc-submit-input"),
			OutputSchemaHash:	engineHash("rpc-submit-output"),
			ExecutionMode:		AVMInterfaceExecutionSync,
			GasHint:		11,
		}},
		SchemaEncoding:	AVMInterfaceSchemaProtobuf,
	})
	require.NoError(t, err)
	contractSchema := testAVMInterfaceSchema(t, contractDescriptor)
	serviceSchema := testAVMInterfaceSchema(t, serviceDescriptor)
	contractBinding, err := NewAVMInterfaceBinding(AVMInterfaceBinding{
		TargetID:	"contract-actor-1",
		TargetType:	AVMInterfaceTargetContract,
		InterfaceHash:	contractDescriptor.InterfaceHash,
	})
	require.NoError(t, err)
	serviceBinding, err := NewAVMInterfaceBinding(AVMInterfaceBinding{
		TargetID:	"rpc-service",
		TargetType:	AVMInterfaceTargetService,
		InterfaceHash:	serviceDescriptor.InterfaceHash,
	})
	require.NoError(t, err)
	registry, err := NewAVMInterfaceRegistry(AVMInterfaceRegistry{
		Interfaces:	[]AVMInterfaceDescriptor{serviceDescriptor, contractDescriptor},
		Schemas:	[]AVMInterfaceSchema{serviceSchema, contractSchema},
		Bindings:	[]AVMInterfaceBinding{serviceBinding, contractBinding},
	})
	require.NoError(t, err)

	queriedContract, binding, err := QueryAVMInterfaceByContract(registry, "contract-actor-1")
	require.NoError(t, err)
	require.Equal(t, contractDescriptor.InterfaceHash, queriedContract.InterfaceHash)
	require.Equal(t, contractBinding.BindingHash, binding.BindingHash)
	queriedService, _, err := QueryAVMInterfaceByService(registry, "rpc-service")
	require.NoError(t, err)
	require.Equal(t, serviceDescriptor.InterfaceHash, queriedService.InterfaceHash)

	codegen, err := NewAVMSDKCodeGenerationFormat(AVMSDKCodeGenerationFormat{
		InterfaceHash:		contractDescriptor.InterfaceHash,
		Format:			AVMInterfaceSDKTypeScript,
		PackageName:		"aetra_actor",
		MethodBindings:		[]string{"execute", "schedule"},
		GetMethodBindings:	[]string{"balance"},
	})
	require.NoError(t, err)
	require.NoError(t, codegen.Validate())
}

func TestAVMInterfaceHashVerificationSchemaMismatchAndVersionChanges(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)
	require.NoError(t, VerifyAVMInterfaceHash(descriptor, descriptor.InterfaceHash))

	versioned := descriptor
	versioned.InterfaceVersion = "v1.0.1"
	versioned.InterfaceHash = ComputeAVMInterfaceHash(versioned)
	require.NotEqual(t, descriptor.InterfaceHash, versioned.InterfaceHash)
	require.ErrorContains(t, VerifyAVMInterfaceHash(versioned, descriptor.InterfaceHash), "verification failed")

	schema := testAVMInterfaceSchema(t, descriptor)
	require.NoError(t, VerifyAVMInterfaceSchema(descriptor, schema))
	schema.DescriptorRoot = engineHash("wrong-descriptor-root")
	schema.SchemaHash = ComputeAVMInterfaceSchemaHash(schema)
	require.NotEqual(t, ComputeAVMInterfaceDescriptorRoot(descriptor), schema.DescriptorRoot)
	require.NoError(t, schema.Validate())
	require.ErrorContains(t, VerifyAVMInterfaceSchema(descriptor, schema), "descriptor root mismatch")

	registry, err := NewAVMInterfaceRegistry(AVMInterfaceRegistry{
		Interfaces:	[]AVMInterfaceDescriptor{descriptor},
		Schemas:	[]AVMInterfaceSchema{schema},
	})
	require.NoError(t, err)
	require.Equal(t, schema.SchemaHash, registry.Schemas[0].SchemaHash)
}

func TestAVMInterfaceDescriptorRejectsMalformedDescriptorSets(t *testing.T) {
	descriptor := testAVMInterfaceDescriptor(t)

	empty := descriptor
	empty.MethodDescriptors = nil
	empty.EventDescriptors = nil
	empty.AsyncHandlerDescriptors = nil
	empty.GetMethodDescriptors = nil
	empty.InterfaceHash = ComputeAVMInterfaceHash(empty)
	require.ErrorContains(t, empty.Validate(), "at least one")

	duplicateMethod := descriptor
	duplicateMethod.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	duplicateMethod.MethodDescriptors = append(duplicateMethod.MethodDescriptors, descriptor.MethodDescriptors[0])
	duplicateMethod.InterfaceHash = ComputeAVMInterfaceHash(duplicateMethod)
	require.ErrorContains(t, duplicateMethod.Validate(), "duplicate AVM method")

	badMetadata := descriptor
	badMetadata.MetadataHashOptional = "not-a-hash"
	badMetadata.InterfaceHash = ComputeAVMInterfaceHash(badMetadata)
	require.ErrorContains(t, badMetadata.Validate(), "metadata")

	badEncoding := descriptor
	badEncoding.SchemaEncoding = AVMInterfaceSchemaEncoding("yaml")
	badEncoding.InterfaceHash = ComputeAVMInterfaceHash(badEncoding)
	require.ErrorContains(t, badEncoding.Validate(), "schema encoding")

	authMetadata := descriptor
	authMetadata.MetadataGrantsAuth = true
	authMetadata.InterfaceHash = ComputeAVMInterfaceHash(authMetadata)
	require.ErrorContains(t, authMetadata.Validate(), "authorization")

	getWrites := descriptor
	getWrites.GetMethodDescriptors = append([]AVMGetMethodDescriptor(nil), descriptor.GetMethodDescriptors...)
	getWrites.GetMethodDescriptors[0].ReadOnly = false
	getWrites.InterfaceHash = ComputeAVMInterfaceHash(getWrites)
	require.ErrorContains(t, getWrites.Validate(), "read-only")

	missingCallback := descriptor
	missingCallback.AsyncHandlerDescriptors = append([]AVMAsyncHandlerDescriptor(nil), descriptor.AsyncHandlerDescriptors...)
	missingCallback.AsyncHandlerDescriptors[0].CallbackBehavior = ""
	missingCallback.InterfaceHash = ComputeAVMInterfaceHash(missingCallback)
	require.ErrorContains(t, missingCallback.Validate(), "callback")
}

func testAVMInterfaceDescriptor(t *testing.T) AVMInterfaceDescriptor {
	t.Helper()
	descriptor, err := NewAVMInterfaceDescriptor(AVMInterfaceDescriptor{
		InterfaceVersion:	"v1.0.0",
		Owner:			"actor-contract-1",
		TargetType:		AVMInterfaceTargetActor,
		MethodDescriptors: []AVMMethodDescriptor{
			{
				MethodID:			"actor.execute",
				Name:				"execute",
				InputSchemaHash:		engineHash("execute-input"),
				OutputSchemaHash:		engineHash("execute-output"),
				ExecutionMode:			AVMInterfaceExecutionAsync,
				GasHint:			100,
				PaymentRequirementOptional:	"naet",
				ProofRequirementOptional:	"state-proof",
			},
			{
				MethodID:		"actor.schedule",
				Name:			"schedule",
				InputSchemaHash:	engineHash("schedule-input"),
				OutputSchemaHash:	engineHash("schedule-output"),
				ExecutionMode:		AVMInterfaceExecutionScheduled,
				GasHint:		120,
			},
		},
		EventDescriptors: []AVMEventDescriptor{{
			EventID:	"actor.executed",
			Name:		"ActorExecuted",
			SchemaHash:	engineHash("event-schema"),
		}},
		AsyncHandlerDescriptors: []AVMAsyncHandlerDescriptor{{
			HandlerID:		"actor.handle",
			Name:			"handle",
			InputSchemaHash:	engineHash("handler-input"),
			OutputSchemaHash:	engineHash("handler-output"),
			GasHint:		80,
			RetryPolicyOptional:	"bounded",
			CallbackBehavior:	"emit_receipt",
			TimeoutHeight:		10,
		}},
		GetMethodDescriptors: []AVMGetMethodDescriptor{{
			MethodID:		"actor.balance",
			Name:			"balance",
			InputSchemaHash:	engineHash("get-input"),
			OutputSchemaHash:	engineHash("get-output"),
			GasHint:		5,
			ReadOnly:		true,
		}},
		SchemaEncoding:		AVMInterfaceSchemaJSONSchema,
		MetadataHashOptional:	engineHash("interface-metadata"),
	})
	require.NoError(t, err)
	return descriptor
}

func testAVMInterfaceSchema(t *testing.T, descriptor AVMInterfaceDescriptor) AVMInterfaceSchema {
	t.Helper()
	schema, err := NewAVMInterfaceSchema(AVMInterfaceSchema{
		InterfaceHash:	descriptor.InterfaceHash,
		SchemaEncoding:	descriptor.SchemaEncoding,
		DescriptorRoot:	ComputeAVMInterfaceDescriptorRoot(descriptor),
		UseCases: []AVMInterfaceUseCase{
			AVMInterfaceUseUIGeneration,
			AVMInterfaceUseWalletForms,
			AVMInterfaceUseCLIAutoBinding,
			AVMInterfaceUseRPCIntrospection,
			AVMInterfaceUseSDKCallBuilders,
			AVMInterfaceUseCapabilityDiscovery,
		},
	})
	require.NoError(t, err)
	return schema
}
