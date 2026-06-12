package avm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeveloperUXManifestBuildsHashCommittedClientMetadata(t *testing.T) {
	manifest := developerUXManifest()
	metadata, err := BuildInterfaceDeveloperMetadata(manifest)
	require.NoError(t, err)
	require.Equal(t, "counter", metadata.Name)
	require.Len(t, metadata.Methods, 1)
	require.Len(t, metadata.Events, 1)
	require.Len(t, metadata.AsyncHandlers, 1)
	require.Len(t, metadata.GetMethods, 1)
	require.Len(t, metadata.CLIBindings, 1)
	require.Len(t, metadata.SDKBindings, 1)
	require.Len(t, metadata.WalletActions, 1)
	require.Equal(t, []string{"counter-inc", "inc"}, metadata.CLIBindings[0].Aliases)

	hash, err := InterfaceHash(manifest)
	require.NoError(t, err)
	require.Equal(t, hash, metadata.MetadataHash)

	module := counterModule()
	module.MetadataHash = metadata.MetadataHash
	require.NoError(t, VerifyInterface(module, manifest))

	missingQuery := counterModule()
	missingQuery.MetadataHash = metadata.MetadataHash
	delete(missingQuery.Exports, EntryQuery)
	require.ErrorContains(t, VerifyInterface(missingQuery, manifest), "get method")

	mutated := manifest
	mutated.WalletActions = append([]InterfaceWalletAction(nil), manifest.WalletActions...)
	mutated.WalletActions[0].ConfirmLabel = "Apply"
	mutatedHash, err := InterfaceHash(mutated)
	require.NoError(t, err)
	require.NotEqual(t, metadata.MetadataHash, mutatedHash)
}

func TestDeveloperUXManifestRejectsInvalidDescriptors(t *testing.T) {
	base := InterfaceManifest{
		Name:		"counter",
		Version:	1,
		Methods: []InterfaceMethod{{
			Name:		"increment",
			Entrypoint:	EntryReceiveInternal,
			Opcode:		1,
			Params: []InterfaceParamDescriptor{{
				Name:		"amount",
				Kind:		InterfaceValueU64,
				Required:	true,
			}},
			Results: []InterfaceResultDescriptor{{
				Name:	"counter",
				Kind:	InterfaceValueU64,
			}},
		}},
	}

	tests := []struct {
		name	string
		mutate	func(*InterfaceManifest)
		want	string
	}{
		{
			name:	"get method must be query",
			mutate: func(manifest *InterfaceManifest) {
				manifest.GetMethods = []InterfaceGetMethod{{
					Name:		"get_counter",
					Entrypoint:	EntryReceiveInternal,
					Selector:	1,
				}}
			},
			want:	"query entrypoint",
		},
		{
			name:	"duplicate callable names are rejected",
			mutate: func(manifest *InterfaceManifest) {
				manifest.GetMethods = []InterfaceGetMethod{{
					Name:		"increment",
					Entrypoint:	EntryQuery,
					Selector:	1,
				}}
			},
			want:	"duplicate",
		},
		{
			name:	"async bounced flag must match entrypoint",
			mutate: func(manifest *InterfaceManifest) {
				manifest.AsyncHandlers = []InterfaceAsyncHandler{{
					Name:		"bounce",
					Entrypoint:	EntryReceiveInternal,
					Opcode:		2,
					Bounced:	true,
				}}
			},
			want:	"bounced flag",
		},
		{
			name:	"unknown SDK binding method is rejected",
			mutate: func(manifest *InterfaceManifest) {
				manifest.SDKBindings = []InterfaceSDKBinding{{
					Method:		"missing",
					Package:	"github.com/sovereign-l1/l1/sdk/avm",
					Service:	"Counter",
					MethodName:	"Increment",
					RequestType:	"IncrementRequest",
					ResponseType:	"IncrementResponse",
				}}
			},
			want:	"not declared",
		},
		{
			name:	"wallet input kind is validated",
			mutate: func(manifest *InterfaceManifest) {
				manifest.WalletActions = []InterfaceWalletAction{{
					Method:		"increment",
					Title:		"Increment counter",
					Risk:		InterfaceWalletRiskLow,
					ConfirmLabel:	"Increment",
					Inputs: []InterfaceParamDescriptor{{
						Name:	"amount",
						Kind:	InterfaceValueKind("map"),
					}},
				}}
			},
			want:	"invalid kind",
		},
		{
			name:	"wallet inputs must match declared parameters",
			mutate: func(manifest *InterfaceManifest) {
				manifest.WalletActions = []InterfaceWalletAction{{
					Method:		"increment",
					Title:		"Increment counter",
					Risk:		InterfaceWalletRiskLow,
					ConfirmLabel:	"Increment",
					Inputs: []InterfaceParamDescriptor{{
						Name:	"missing",
						Kind:	InterfaceValueU64,
					}},
				}}
			},
			want:	"not declared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := base
			tt.mutate(&manifest)
			_, err := InterfaceHash(manifest)
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func developerUXManifest() InterfaceManifest {
	return InterfaceManifest{
		Name:		" counter ",
		Version:	1,
		Methods: []InterfaceMethod{{
			Name:		"increment",
			Entrypoint:	EntryReceiveInternal,
			Opcode:		1,
			Async:		true,
			Description:	"Increment the counter by a bounded amount.",
			Params: []InterfaceParamDescriptor{{
				Name:		"amount",
				Kind:		InterfaceValueU64,
				Required:	true,
				Description:	"Amount to add.",
			}},
			Results: []InterfaceResultDescriptor{{
				Name:	"counter",
				Kind:	InterfaceValueU64,
			}},
		}},
		Events: []InterfaceEvent{{
			Name:	"incremented",
			Opcode:	10,
			Fields: []InterfaceParamDescriptor{{
				Name:		"counter",
				Kind:		InterfaceValueU64,
				Required:	true,
			}},
		}},
		AsyncHandlers: []InterfaceAsyncHandler{{
			Name:		"resume",
			Entrypoint:	EntryReceiveInternal,
			Opcode:		2,
			MessageType:	"counter.resume",
			Idempotent:	true,
			Params: []InterfaceParamDescriptor{{
				Name:		"query_id",
				Kind:		InterfaceValueU64,
				Required:	true,
			}},
		}},
		GetMethods: []InterfaceGetMethod{{
			Name:			"get_counter",
			Entrypoint:		EntryQuery,
			Selector:		1,
			Cacheable:		true,
			MaxResponseBytes:	64,
			Results: []InterfaceResultDescriptor{{
				Name:	"counter",
				Kind:	InterfaceValueU64,
			}},
		}},
		CLIBindings: []InterfaceCLIBinding{{
			Method:		"increment",
			Command:	"counter increment",
			Use:		"counter increment --amount <u64>",
			Aliases:	[]string{"inc", "counter-inc"},
			Examples:	[]string{"aether avm counter increment --amount 1"},
			InputFormat:	"flags",
			OutputFormat:	"json",
		}},
		SDKBindings: []InterfaceSDKBinding{{
			Method:		"increment",
			Package:	"github.com/sovereign-l1/l1/sdk/avm",
			Service:	"Counter",
			MethodName:	"Increment",
			RequestType:	"IncrementRequest",
			ResponseType:	"IncrementResponse",
			Async:		true,
		}},
		WalletActions: []InterfaceWalletAction{{
			Method:		"increment",
			Title:		"Increment counter",
			Description:	"Adds the provided amount to the AVM counter state.",
			Category:	"counter",
			Icon:		"counter",
			Risk:		InterfaceWalletRiskLow,
			ConfirmLabel:	"Increment",
			Inputs: []InterfaceParamDescriptor{{
				Name:		"amount",
				Kind:		InterfaceValueU64,
				Required:	true,
			}},
			Outputs: []InterfaceResultDescriptor{{
				Name:	"counter",
				Kind:	InterfaceValueU64,
			}},
		}},
	}
}
