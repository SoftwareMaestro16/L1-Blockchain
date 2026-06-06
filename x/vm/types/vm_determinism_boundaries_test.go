package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestVMDeterminismProfileRejectsNondeterministicInputs(t *testing.T) {
	profile, err := DefaultVMDeterminismProfile(RuntimeAVM)
	require.NoError(t, err)
	require.NoError(t, profile.Validate())
	require.Equal(t, ComputeVMDeterminismProfileHash(profile), profile.ProfileHash)

	external := profile
	external.NoExternalAPICalls = false
	external.ProfileHash = ComputeVMDeterminismProfileHash(external)
	require.ErrorContains(t, external.Validate(), "external API")

	random := profile
	random.NoTimeBasedRandomness = false
	random.ProfileHash = ComputeVMDeterminismProfileHash(random)
	require.ErrorContains(t, random.Validate(), "time-based randomness")

	unsorted := profile
	unsorted.SortedMessageApplication = false
	unsorted.ProfileHash = ComputeVMDeterminismProfileHash(unsorted)
	require.ErrorContains(t, unsorted.Validate(), "sorted message")

	unboundedIteration := profile
	unboundedIteration.MaxIterationCount = 0
	unboundedIteration.ProfileHash = ComputeVMDeterminismProfileHash(unboundedIteration)
	require.ErrorContains(t, unboundedIteration.Validate(), "bounded iteration")

	unmeteredProof := profile
	unmeteredProof.MeteredProofVerification = false
	unmeteredProof.ProfileHash = ComputeVMDeterminismProfileHash(unmeteredProof)
	require.ErrorContains(t, unmeteredProof.Validate(), "metered proof")
}

func TestAVMAdapterBoundaryRequiresGasStoreMessageAndProofSyscalls(t *testing.T) {
	boundary, err := NewAVMAdapterBoundary(AVMAdapterBoundary{
		BytecodeKind: AVMBytecodeIntermediateIR,
	}, zonestypes.ZoneIDContract)
	require.NoError(t, err)
	require.NoError(t, boundary.ValidateForZone(zonestypes.ZoneIDContract))
	require.Equal(t, RuntimeAVM, boundary.Runtime)
	require.Equal(t, DefaultAVMStoreKey, boundary.StoreKey)
	require.Equal(t, ContractZoneKVPrefix(zonestypes.ZoneIDContract), boundary.KVPrefix)
	require.True(t, boundary.StoreV2Backed)
	require.Equal(t, AVMGasClassCrossZoneRouting, boundary.MessageSyscall.GasClass)
	require.Equal(t, AVMGasClassProofVerification, boundary.ProofVerificationSyscall.GasClass)

	wrongPrefix := boundary
	wrongPrefix.KVPrefix = ContractZoneKVPrefix(zonestypes.ZoneIDApplication)
	wrongPrefix.BoundaryHash = ComputeAVMAdapterBoundaryHash(wrongPrefix)
	require.ErrorContains(t, wrongPrefix.ValidateForZone(zonestypes.ZoneIDContract), "KV prefix")

	unmeteredMessage := boundary
	unmeteredMessage.MessageSyscall.Metered = false
	unmeteredMessage.BoundaryHash = ComputeAVMAdapterBoundaryHash(unmeteredMessage)
	require.ErrorContains(t, unmeteredMessage.ValidateForZone(zonestypes.ZoneIDContract), "metered")

	wrongProofClass := boundary
	wrongProofClass.ProofVerificationSyscall.GasClass = AVMGasClassStorage
	wrongProofClass.BoundaryHash = ComputeAVMAdapterBoundaryHash(wrongProofClass)
	require.ErrorContains(t, wrongProofClass.ValidateForZone(zonestypes.ZoneIDContract), "proof verification gas")
}

func TestCosmWasmAdapterBoundaryEnforcesIsolationPrefixAndCrossZoneProofs(t *testing.T) {
	boundary, err := NewCosmWasmAdapterBoundary(CosmWasmAdapterBoundary{}, zonestypes.ZoneIDContract)
	require.NoError(t, err)
	require.NoError(t, boundary.Validate())
	require.Equal(t, RuntimeCosmWasm, boundary.Runtime)
	require.True(t, boundary.IsolatedAdapterModule)
	require.True(t, boundary.ExplicitStorageKeyPrefix)
	require.True(t, boundary.CrossZoneMessagesOrProofs)
	require.False(t, boundary.DirectNonContractState)
	require.False(t, boundary.ExternalNetwork)
	require.Equal(t, ContractZoneKVPrefix(zonestypes.ZoneIDContract), boundary.StoreAdapter.KeyPrefix)

	directState := boundary
	directState.DirectNonContractState = true
	directState.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(directState)
	require.ErrorContains(t, directState.Validate(), "non-contract zone state")

	noCrossZoneProof := boundary
	noCrossZoneProof.CrossZoneMessagesOrProofs = false
	noCrossZoneProof.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(noCrossZoneProof)
	require.ErrorContains(t, noCrossZoneProof.Validate(), "messages or proofs")

	network := boundary
	network.ExternalNetwork = true
	network.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(network)
	require.ErrorContains(t, network.Validate(), "external network")

	badHost := boundary
	badHost.HostFunctions = append([]AVMWASMHostFunction(nil), boundary.HostFunctions...)
	badHost.HostFunctions[0].Deterministic = false
	badHost.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(badHost)
	require.ErrorContains(t, badHost.Validate(), "deterministic")
}

func TestVMAdapterBoundaryManifestCommitsAllCriticalBoundaries(t *testing.T) {
	manifest, err := NewVMAdapterBoundaryManifest(VMAdapterBoundaryManifest{
		ZoneID: zonestypes.ZoneIDContract,
	})
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Equal(t, ComputeVMAdapterBoundaryManifestHash(manifest), manifest.ManifestHash)
	require.Equal(t, ComputeVMDeterminismProfileHash(manifest.DeterminismProfile), manifest.DeterminismProfile.ProfileHash)
	require.Equal(t, ComputeAVMAdapterBoundaryHash(manifest.AVM), manifest.AVM.BoundaryHash)
	require.Equal(t, ComputeCosmWasmAdapterBoundaryHash(manifest.CosmWasm), manifest.CosmWasm.BoundaryHash)

	tampered := manifest
	tampered.AVM.StoreV2Backed = false
	tampered.AVM.BoundaryHash = ComputeAVMAdapterBoundaryHash(tampered.AVM)
	tampered.ManifestHash = ComputeVMAdapterBoundaryManifestHash(tampered)
	require.ErrorContains(t, tampered.Validate(), "Store v2-backed")

	zoneMismatch := manifest
	zoneMismatch.CosmWasm.StoreAdapter.ZoneID = zonestypes.ZoneIDApplication
	zoneMismatch.CosmWasm.StoreAdapter.KeyPrefix = ContractZoneKVPrefix(zonestypes.ZoneIDApplication)
	zoneMismatch.CosmWasm.StoreAdapter.AdapterHash = ComputeAVMWASMStoreV2KVAdapterHash(zoneMismatch.CosmWasm.StoreAdapter)
	zoneMismatch.CosmWasm.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(zoneMismatch.CosmWasm)
	zoneMismatch.ManifestHash = ComputeVMAdapterBoundaryManifestHash(zoneMismatch)
	require.ErrorContains(t, zoneMismatch.Validate(), "zone mismatch")
}
