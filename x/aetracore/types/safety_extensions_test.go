package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendedSafetyRulesSpecCoversShardVMAndProofSafety(t *testing.T) {
	require.NoError(t, ValidateExtendedSafetyRulesCoverage())

	spec, err := DefaultExtendedSafetyRulesSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.ShardSafety, 5)
	require.Len(t, spec.VMSafety, 6)
	require.Len(t, spec.ProofSafety, 5)
	require.NotEmpty(t, spec.Root)

	shard := map[ShardSafetyRuleID]ShardSafetyRule{}
	for _, rule := range spec.ShardSafety {
		require.NoError(t, rule.Validate())
		shard[rule.RuleID] = rule
	}
	require.Contains(t, shard[ShardSafetyEpochBoundary].Rule, "epoch boundaries")
	require.Contains(t, shard[ShardSafetyReproducibleDecision].Evidence, "SourceMetricsHash")
	require.Contains(t, shard[ShardSafetyDeliveryEpoch].Evidence, "DeliveryEpoch")
	require.Contains(t, shard[ShardSafetyMigrationRoot].Rule, "migration root")
	require.Contains(t, shard[ShardSafetyProofHorizon].Enforcement, "proof horizon")

	vm := map[VMSafetyRuleID]VMSafetyRule{}
	for _, rule := range spec.VMSafety {
		require.NoError(t, rule.Validate())
		vm[rule.RuleID] = rule
	}
	require.Contains(t, vm[VMSafetyGasMetering].Rule, "Gas metering")
	require.Contains(t, vm[VMSafetyBoundedIteration].Rule, "bounded")
	require.Contains(t, vm[VMSafetyMeteredProofs].Rule, "Proof verification")
	require.Contains(t, vm[VMSafetyForwardingFeeReserve].Rule, "forwarding fee")
	require.Contains(t, vm[VMSafetyNoRemoteMutation].Rule, "remote zone")
	require.Contains(t, vm[VMSafetyDeterministicTimeout].Rule, "Promise timeouts")

	proof := map[ProofSafetyRuleID]ProofSafetyRule{}
	for _, rule := range spec.ProofSafety {
		require.NoError(t, rule.Validate())
		proof[rule.RuleID] = rule
	}
	require.Contains(t, proof[ProofSafetyTrustedHeight].Rule, "trusted header height")
	require.Contains(t, proof[ProofSafetyZoneShardIDs].Rule, "zone and shard")
	require.Contains(t, proof[ProofSafetyObjectKeyRootType].Rule, "object key")
	require.Contains(t, proof[ProofSafetyExplicitAbsence].Rule, "Non-existence")
	require.Contains(t, proof[ProofSafetySupportedVersion].Rule, "unsupported")
}

func TestExtendedSafetyRulesSpecRootIsCanonicalAndRejectsTamper(t *testing.T) {
	spec, err := DefaultExtendedSafetyRulesSpec()
	require.NoError(t, err)

	shard := append([]ShardSafetyRule(nil), ShardSafetyRules()...)
	vm := append([]VMSafetyRule(nil), VMSafetyRules()...)
	proof := append([]ProofSafetyRule(nil), ProofSafetyRules()...)
	slices.Reverse(shard)
	slices.Reverse(vm)
	slices.Reverse(proof)
	reordered, err := BuildExtendedSafetyRulesSpec(shard, vm, proof)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reordered.Root)
	require.Equal(t, spec.ShardSafety, reordered.ShardSafety)
	require.Equal(t, spec.VMSafety, reordered.VMSafety)
	require.Equal(t, spec.ProofSafety, reordered.ProofSafety)

	_, err = BuildExtendedSafetyRulesSpec([]ShardSafetyRule{ShardSafetyRules()[0], ShardSafetyRules()[0]}, VMSafetyRules(), ProofSafetyRules())
	require.ErrorContains(t, err, "duplicate")

	tampered := VMSafetyRules()[0]
	tampered.Enforcement = strings.ReplaceAll(tampered.Enforcement, "gas", "free")
	require.ErrorContains(t, tampered.Validate(), "hash mismatch")
}

func TestShardSafetyEvidenceRequiresEpochMigrationAndProofHorizon(t *testing.T) {
	evidence := ShardSafetyEvidence{
		SourceLayoutEpoch:	7,
		TargetLayoutEpoch:	8,
		ActivationHeight:	110,
		DecisionHeight:		100,
		DecisionHash:		hashParts("decision"),
		CommittedMetricsRoot:	hashParts("metrics"),
		DeliveryEpoch:		8,
		MigrationRoot:		hashParts("migration"),
		OldLayoutHash:		hashParts("old-layout"),
		ProofHorizonUntil:	200,
	}
	evidence.EvidenceHash = ComputeShardSafetyEvidenceHash(evidence)
	require.NoError(t, ValidateShardSafetyEvidence(evidence))

	sameEpoch := evidence
	sameEpoch.TargetLayoutEpoch = sameEpoch.SourceLayoutEpoch
	sameEpoch.EvidenceHash = ComputeShardSafetyEvidenceHash(sameEpoch)
	require.ErrorContains(t, ValidateShardSafetyEvidence(sameEpoch), "future")

	earlyDelivery := evidence
	earlyDelivery.DeliveryEpoch = 7
	earlyDelivery.EvidenceHash = ComputeShardSafetyEvidenceHash(earlyDelivery)
	require.ErrorContains(t, ValidateShardSafetyEvidence(earlyDelivery), "delivery epoch")

	missingMigration := evidence
	missingMigration.MigrationRoot = ""
	missingMigration.EvidenceHash = ComputeShardSafetyEvidenceHash(missingMigration)
	require.ErrorContains(t, ValidateShardSafetyEvidence(missingMigration), "migration root")

	shortHorizon := evidence
	shortHorizon.ProofHorizonUntil = 109
	shortHorizon.EvidenceHash = ComputeShardSafetyEvidenceHash(shortHorizon)
	require.ErrorContains(t, ValidateShardSafetyEvidence(shortHorizon), "proof horizon")
}

func TestVMSafetyEvidenceRequiresMeteringFeeAndDeterministicTimeout(t *testing.T) {
	evidence := VMSafetyEvidence{
		GasTableHash:			hashParts("gas-table"),
		GasLimit:			1_000,
		GasUsed:			500,
		MaxStorageIterationItems:	64,
		ProofVerificationGas:		20,
		ForwardingFeeReserved:		5,
		CreatedMessageCount:		1,
		PromiseTimeoutHeight:		120,
		ConsensusHeight:		100,
	}
	evidence.EvidenceHash = ComputeVMSafetyEvidenceHash(evidence)
	require.NoError(t, ValidateVMSafetyEvidence(evidence))

	noGas := evidence
	noGas.GasLimit = 0
	noGas.EvidenceHash = ComputeVMSafetyEvidenceHash(noGas)
	require.ErrorContains(t, ValidateVMSafetyEvidence(noGas), "gas metering")

	unboundedIteration := evidence
	unboundedIteration.MaxStorageIterationItems = 0
	unboundedIteration.EvidenceHash = ComputeVMSafetyEvidenceHash(unboundedIteration)
	require.ErrorContains(t, ValidateVMSafetyEvidence(unboundedIteration), "bounded")

	noProofGas := evidence
	noProofGas.ProofVerificationGas = 0
	noProofGas.EvidenceHash = ComputeVMSafetyEvidenceHash(noProofGas)
	require.ErrorContains(t, ValidateVMSafetyEvidence(noProofGas), "proof verification gas")

	noFee := evidence
	noFee.ForwardingFeeReserved = 0
	noFee.EvidenceHash = ComputeVMSafetyEvidenceHash(noFee)
	require.ErrorContains(t, ValidateVMSafetyEvidence(noFee), "forwarding fee")

	remoteMutation := evidence
	remoteMutation.SynchronousRemoteMutation = true
	remoteMutation.EvidenceHash = ComputeVMSafetyEvidenceHash(remoteMutation)
	require.ErrorContains(t, ValidateVMSafetyEvidence(remoteMutation), "remote zone")

	badTimeout := evidence
	badTimeout.PromiseTimeoutHeight = 100
	badTimeout.EvidenceHash = ComputeVMSafetyEvidenceHash(badTimeout)
	require.ErrorContains(t, ValidateVMSafetyEvidence(badTimeout), "future consensus height")
}

func TestProofSafetyEvidenceRequiresHeightScopeKeyAbsenceAndVersion(t *testing.T) {
	evidence := ProofSafetyEvidence{
		ProofVersion:		UniversalProofVersionV1,
		TrustedHeaderHeight:	55,
		ProofHeight:		55,
		ZoneID:			ZoneIDIdentity,
		ShardID:		"2",
		RootType:		ResolverProofRootType,
		ObjectKey:		[]byte("identity/resolvers/alice.aet"),
		NonExistenceProof:	true,
		AbsenceMarker:		[]byte("between:alice:bob"),
	}
	evidence.EvidenceHash = ComputeProofSafetyEvidenceHash(evidence)
	require.NoError(t, ValidateProofSafetyEvidence(evidence))

	badVersion := evidence
	badVersion.ProofVersion = UniversalProofVersionV1 + 1
	badVersion.EvidenceHash = ComputeProofSafetyEvidenceHash(badVersion)
	require.ErrorContains(t, ValidateProofSafetyEvidence(badVersion), "unsupported")

	badHeight := evidence
	badHeight.ProofHeight = 56
	badHeight.EvidenceHash = ComputeProofSafetyEvidenceHash(badHeight)
	require.ErrorContains(t, ValidateProofSafetyEvidence(badHeight), "trusted header")

	missingShard := evidence
	missingShard.ShardID = ""
	missingShard.EvidenceHash = ComputeProofSafetyEvidenceHash(missingShard)
	require.ErrorContains(t, ValidateProofSafetyEvidence(missingShard), "shard id")

	missingKey := evidence
	missingKey.ObjectKey = nil
	missingKey.EvidenceHash = ComputeProofSafetyEvidenceHash(missingKey)
	require.ErrorContains(t, ValidateProofSafetyEvidence(missingKey), "object key")

	missingAbsence := evidence
	missingAbsence.AbsenceMarker = nil
	missingAbsence.EvidenceHash = ComputeProofSafetyEvidenceHash(missingAbsence)
	require.ErrorContains(t, ValidateProofSafetyEvidence(missingAbsence), "explicit marker")
}
