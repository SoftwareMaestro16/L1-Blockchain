package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVMSafetyEnvelopeRequiresMeteringBoundsAndForwardingReserve(t *testing.T) {
	envelope := validVMSafetyEnvelope()
	built, err := BuildVMSafetyEnvelope(envelope)
	require.NoError(t, err)
	require.NoError(t, built.Validate())

	noGas := built
	noGas.GasMeteringEnabled = false
	noGas.SafetyHash = ComputeVMSafetyEnvelopeHash(noGas)
	require.ErrorContains(t, noGas.Validate(), "gas metering")

	unboundedStorage := built
	unboundedStorage.StorageIterations = unboundedStorage.StorageIterationLimit + 1
	unboundedStorage.SafetyHash = ComputeVMSafetyEnvelopeHash(unboundedStorage)
	require.ErrorContains(t, unboundedStorage.Validate(), "storage iteration")

	underMeteredProof := built
	underMeteredProof.ProofVerificationGasConsumed = underMeteredProof.ProofVerificationGasReserved + 1
	underMeteredProof.SafetyHash = ComputeVMSafetyEnvelopeHash(underMeteredProof)
	require.ErrorContains(t, underMeteredProof.Validate(), "under-metered")

	noForwardReserve := built
	noForwardReserve.ReservedForwardingFee = "1"
	noForwardReserve.SafetyHash = ComputeVMSafetyEnvelopeHash(noForwardReserve)
	require.ErrorContains(t, noForwardReserve.Validate(), "reserved forwarding fee")
}

func TestVMSafetyRejectsSynchronousRemoteMutationAndNondeterministicTimeout(t *testing.T) {
	envelope := validVMSafetyEnvelope()

	remoteSync := envelope
	remoteSync.RemoteMutationMode = VMRemoteMutationSynchronous
	remoteSync.SafetyHash = ComputeVMSafetyEnvelopeHash(remoteSync)
	require.ErrorContains(t, remoteSync.Validate(), "synchronously mutate remote")

	badTimeout := envelope
	badTimeout.PromiseTimeout.TimeoutHeight++
	badTimeout.PromiseTimeout.PolicyHash = ComputeVMPromiseTimeoutPolicyHash(badTimeout.PromiseTimeout)
	badTimeout.SafetyHash = ComputeVMSafetyEnvelopeHash(badTimeout)
	require.ErrorContains(t, badTimeout.Validate(), "deterministic")
}

func TestProofSafetyBindsTrustedHeaderZoneShardObjectAndRootType(t *testing.T) {
	proof := validProofSafetyEnvelope(false)
	built, err := BuildUniversalProofSafetyEnvelope(proof)
	require.NoError(t, err)
	require.NoError(t, built.Validate())

	wrongHeight := built
	wrongHeight.ProofHeight = built.TrustedHeader.Height + 1
	wrongHeight.ProofHash = ComputeUniversalProofSafetyHash(wrongHeight)
	require.ErrorContains(t, wrongHeight.Validate(), "trusted header height")

	missingZone := built
	missingZone.ZoneID = ""
	missingZone.ProofHash = ComputeUniversalProofSafetyHash(missingZone)
	require.ErrorContains(t, missingZone.Validate(), "zone id")

	badRootType := built
	badRootType.RootType = ProofRootType("unsupported")
	badRootType.ProofHash = ComputeUniversalProofSafetyHash(badRootType)
	require.ErrorContains(t, badRootType.Validate(), "root type")
}

func TestProofSafetyRequiresExplicitNonExistenceAndSupportedVersions(t *testing.T) {
	absence := validProofSafetyEnvelope(true)
	built, err := BuildUniversalProofSafetyEnvelope(absence)
	require.NoError(t, err)
	require.True(t, built.NonExistence)
	require.NoError(t, built.Validate())

	missingAbsence := built
	missingAbsence.AbsenceProofHash = ""
	missingAbsence.ProofHash = ComputeUniversalProofSafetyHash(missingAbsence)
	require.ErrorContains(t, missingAbsence.Validate(), "explicit")

	unsupported := built
	unsupported.ProofVersion = 2
	unsupported.ProofHash = ComputeUniversalProofSafetyHash(unsupported)
	require.ErrorContains(t, unsupported.Validate(), "unsupported")
}

func validVMSafetyEnvelope() VMSafetyEnvelope {
	timeout := VMPromiseTimeoutPolicy{
		CreatedHeight:	100,
		TimeoutHeight:	110,
		DelayBlocks:	10,
	}
	timeout.PolicyHash = ComputeVMPromiseTimeoutPolicyHash(timeout)
	envelope := VMSafetyEnvelope{
		ExecutionID:			"exec-1",
		ZoneID:				"contract",
		ShardID:			"shard-a",
		ContractAddress:		"contract/escrow",
		GasMeteringEnabled:		true,
		GasLimit:			10_000,
		GasConsumed:			9_000,
		StorageIterationLimit:		128,
		StorageIterations:		64,
		ProofVerificationCount:		2,
		ProofVerificationGasReserved:	1_000,
		ProofVerificationGasConsumed:	900,
		MessageCreationCount:		1,
		ReservedForwardingFee:		"5",
		ForwardingFeeRequired:		"3",
		RemoteMutationMode:		VMRemoteMutationAsyncMessage,
		PromiseTimeout:			timeout,
	}
	envelope.SafetyHash = ComputeVMSafetyEnvelopeHash(envelope)
	return envelope
}

func validProofSafetyEnvelope(nonExistence bool) UniversalProofSafetyEnvelope {
	proof := UniversalProofSafetyEnvelope{
		ProofID:		"proof-1",
		ProofVersion:		ProofSafetyVersionV1,
		SupportedVersions:	[]uint32{ProofSafetyVersionV1},
		TrustedHeader: TrustedHeaderBinding{
			Height:		200,
			HeaderHash:	hashStrings("trusted-header"),
			AppHash:	hashStrings("trusted-app"),
		},
		ProofHeight:	200,
		ZoneID:		"identity",
		ShardID:	"shard-a",
		ObjectKey:	"domains/alice.aet",
		RootType:	ProofRootIdentity,
		RootHash:	hashStrings("identity-root"),
		NonExistence:	nonExistence,
	}
	if nonExistence {
		proof.AbsenceProofHash = hashStrings("absence-proof")
	} else {
		proof.ExistenceProofHash = hashStrings("existence-proof")
	}
	proof.ProofHash = ComputeUniversalProofSafetyHash(proof)
	return proof
}
