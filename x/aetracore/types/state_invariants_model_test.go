package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateInvariantModelCoversSectionElevenFour(t *testing.T) {
	require.NoError(t, ValidateStateInvariantModel())

	model, err := DefaultStateInvariantModel()
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Len(t, model.Invariants, 8)
	require.NotEmpty(t, model.Root)

	byID := map[StateInvariantID]StateInvariantDescriptor{}
	for _, invariant := range model.Invariants {
		require.NoError(t, invariant.Validate())
		byID[invariant.InvariantID] = invariant
	}

	require.Contains(t, byID[StateInvariantZoneCommitmentMatchesState].Invariant, "zone commitment")
	require.Contains(t, byID[StateInvariantZoneCommitmentMatchesState].Enforcement, "Recompute zone state")
	require.Contains(t, byID[StateInvariantShardRootsIncludedInZoneRoot].Enforcement, "shard_roots_root")
	require.Contains(t, byID[StateInvariantOutputMessagesInOutboxRoot].Enforcement, "outbox entries")
	require.Contains(t, byID[StateInvariantConsumedMessagesHaveOneReceipt].Enforcement, "exactly one")
	require.Contains(t, byID[StateInvariantCrossZoneValueConserved].Invariant, "conserve naet")
	require.Contains(t, byID[StateInvariantPaymentCollateralMatchesBankLock].Enforcement, "locked balances")
	require.Contains(t, byID[StateInvariantIdentityDomainOwnershipBinding].Enforcement, "NFT binding")
	require.Contains(t, byID[StateInvariantContractStorageProofZoneBinding].Enforcement, "Contract Zone root")
}

func TestStateInvariantModelRootIsCanonicalAcrossInputOrder(t *testing.T) {
	model, err := DefaultStateInvariantModel()
	require.NoError(t, err)

	reordered := append([]StateInvariantDescriptor(nil), StateInvariantDescriptors()...)
	slices.Reverse(reordered)
	reorderedModel, err := BuildStateInvariantModel(reordered)
	require.NoError(t, err)
	require.Equal(t, model.Root, reorderedModel.Root)
	require.Equal(t, model.Invariants, reorderedModel.Invariants)
}

func TestStateInvariantEvidenceRequiresAllSafetyChecks(t *testing.T) {
	evidence := validStateInvariantEvidence(t)
	require.NoError(t, evidence.Validate())
	require.NotEmpty(t, evidence.EvidenceHash)

	broken := evidence
	broken.ConsumedMessagesHaveOneReceipt = false
	broken.EvidenceHash = ""
	_, err := BuildStateInvariantEvidence(broken)
	require.ErrorContains(t, err, "exactly one receipt")

	tampered := evidence
	tampered.PaymentCollateralRoot = hashParts("tampered-payment-collateral")
	require.ErrorContains(t, tampered.Validate(), "evidence hash mismatch")
}

func TestStateInvariantModelRejectsMalformedDescriptors(t *testing.T) {
	duplicate, err := BuildStateInvariantModel([]StateInvariantDescriptor{
		StateInvariantDescriptors()[0],
		StateInvariantDescriptors()[0],
	})
	require.ErrorContains(t, err, "duplicate state invariant")
	require.Empty(t, duplicate.Root)

	_, err = BuildStateInvariantDescriptor(StateInvariantDescriptor{
		InvariantID:	StateInvariantID("unknown"),
		Invariant:	"unknown invariant",
		Enforcement:	"unknown enforcement",
	})
	require.ErrorContains(t, err, "unknown aetracore state invariant")

	noEnforcement := StateInvariantDescriptor{
		InvariantID:	StateInvariantCrossZoneValueConserved,
		Invariant:	"Every cross-zone value transfer must conserve naet.",
	}
	_, err = BuildStateInvariantDescriptor(noEnforcement)
	require.ErrorContains(t, err, "enforcement is required")

	tampered := StateInvariantDescriptors()[0]
	tampered.Invariant = strings.ReplaceAll(tampered.Invariant, "match", "ignore")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}

func validStateInvariantEvidence(t *testing.T) StateInvariantEvidence {
	t.Helper()
	evidence, err := BuildStateInvariantEvidence(StateInvariantEvidence{
		Height:					177,
		ZoneCommitmentMatchesExecutedState:	true,
		ShardRootsIncludedInZoneRoot:		true,
		OutputMessagesIncludedInSourceOutbox:	true,
		ConsumedMessagesHaveOneReceipt:		true,
		CrossZoneValueTransferConservesNaet:	true,
		PaymentCollateralMatchesLockedBalance:	true,
		IdentityDomainsHaveOwnershipBinding:	true,
		ContractStorageProofsBindToZoneRoot:	true,
		ZoneCommitmentRoot:			hashParts("state-invariant-zone-commitment-root"),
		ShardRootsRoot:				hashParts("state-invariant-shard-roots-root"),
		MessageOutboxRoot:			hashParts("state-invariant-message-outbox-root"),
		MessageReceiptRoot:			hashParts("state-invariant-message-receipt-root"),
		ValueConservationRoot:			hashParts("state-invariant-value-conservation-root"),
		PaymentCollateralRoot:			hashParts("state-invariant-payment-collateral-root"),
		IdentityOwnershipRoot:			hashParts("state-invariant-identity-ownership-root"),
		ContractStorageProofRoot:		hashParts("state-invariant-contract-storage-proof-root"),
	})
	require.NoError(t, err)
	return evidence
}
