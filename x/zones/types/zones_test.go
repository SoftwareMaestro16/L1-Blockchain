package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndActivateCoreZones(t *testing.T) {
	state := EmptyState()
	var err error
	for _, zone := range []Zone{
		testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 10),
		testZone(ZoneIDIdentity, ZoneKindIdentity, VMPolicyNativeModule, 10),
		testZone(ZoneIDApplication, ZoneKindApplication, VMPolicyAVM, 10),
		testZone(ZoneIDContract, ZoneKindContract, VMPolicyCosmWasmGated, 10),
	} {
		state, err = RegisterZone(state, zone)
		require.NoError(t, err)
	}

	_, err = ActivateZone(state, ZoneIDFinancial, 9)
	require.ErrorContains(t, err, "before height 10")

	for _, id := range []ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract} {
		state, err = ActivateZone(state, id, 10)
		require.NoError(t, err)
	}

	require.Equal(t, []ZoneID{ZoneIDApplication, ZoneIDContract, ZoneIDFinancial, ZoneIDIdentity}, state.ActiveZones)
	require.NoError(t, state.Validate())
}

func TestDuplicateZoneRejected(t *testing.T) {
	state, err := RegisterZone(EmptyState(), testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.NoError(t, err)

	_, err = RegisterZone(state, testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.ErrorContains(t, err, "already registered")

	duplicated := state.Clone()
	duplicated.Zones = append(duplicated.Zones, duplicated.Zones[0])
	require.ErrorContains(t, duplicated.Validate(), "duplicate zone")
}

func TestZoneCannotActivateBeforeActivationHeight(t *testing.T) {
	state, err := RegisterZone(EmptyState(), testZone(ZoneIDIdentity, ZoneKindIdentity, VMPolicyNativeModule, 100))
	require.NoError(t, err)

	_, err = ActivateZone(state, ZoneIDIdentity, 99)
	require.ErrorContains(t, err, "cannot activate before height")
}

func TestCommitmentHashStable(t *testing.T) {
	commitmentA, err := NewZoneCommitment(
		ZoneIDFinancial,
		1,
		hash("state"),
		hash("receipt"),
		hash("message"),
		hash("execution"),
		"",
	)
	require.NoError(t, err)
	commitmentB, err := NewZoneCommitment(
		ZoneIDFinancial,
		1,
		hash("state"),
		hash("receipt"),
		hash("message"),
		hash("execution"),
		"",
	)
	require.NoError(t, err)
	require.Equal(t, commitmentA.CommitmentHash, commitmentB.CommitmentHash)

	changed, err := NewZoneCommitment(
		ZoneIDFinancial,
		1,
		hash("state-changed"),
		hash("receipt"),
		hash("message"),
		hash("execution"),
		"",
	)
	require.NoError(t, err)
	require.NotEqual(t, commitmentA.CommitmentHash, changed.CommitmentHash)
}

func TestCommitmentChainDetectsTampering(t *testing.T) {
	state, err := RegisterZone(EmptyState(), testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.NoError(t, err)
	first := newCommitment(t, ZoneIDFinancial, 1, "")
	state, err = AppendCommitment(state, first)
	require.NoError(t, err)

	second := newCommitment(t, ZoneIDFinancial, 2, first.CommitmentHash)
	state, err = AppendCommitment(state, second)
	require.NoError(t, err)

	tampered := state.Clone()
	tampered.Commitments[1].StateRoot = hash("tampered")
	require.ErrorContains(t, tampered.Validate(), "hash mismatch")

	missingPrevious := EmptyState()
	missingPrevious, err = RegisterZone(missingPrevious, testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.NoError(t, err)
	_, err = AppendCommitment(missingPrevious, second)
	require.ErrorContains(t, err, "missing previous commitment")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	state := EmptyState()
	var err error
	for _, zone := range []Zone{
		testZone(ZoneIDContract, ZoneKindContract, VMPolicyCosmWasmGated, 1),
		testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1),
		testZone(ZoneIDApplication, ZoneKindApplication, VMPolicyAVM, 1),
	} {
		state, err = RegisterZone(state, zone)
		require.NoError(t, err)
	}
	state, err = ActivateZone(state, ZoneIDFinancial, 1)
	require.NoError(t, err)
	first := newCommitment(t, ZoneIDFinancial, 1, "")
	state, err = AppendCommitment(state, first)
	require.NoError(t, err)
	state, err = AppendCommitment(state, newCommitment(t, ZoneIDFinancial, 2, first.CommitmentHash))
	require.NoError(t, err)

	exported := state.Export()
	imported, err := ImportState(exported)
	require.NoError(t, err)

	require.Equal(t, exported, imported)
	require.Equal(t, []ZoneID{ZoneIDFinancial}, imported.ActiveZones)
	require.Equal(t, []ZoneID{ZoneIDApplication, ZoneIDContract, ZoneIDFinancial}, []ZoneID{
		imported.Zones[0].ID,
		imported.Zones[1].ID,
		imported.Zones[2].ID,
	})
}

func TestNonNaetFeePolicyRejected(t *testing.T) {
	zone := testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1)
	zone.FeePolicy = "testtoken"

	_, err := RegisterZone(EmptyState(), zone)
	require.ErrorContains(t, err, "naet")
}

func TestUnknownVMRejected(t *testing.T) {
	zone := testZone(ZoneIDContract, ZoneKindContract, VMPolicy("EVM"), 1)

	_, err := RegisterZone(EmptyState(), zone)
	require.ErrorContains(t, err, "unknown zone VM policy")
}

func TestInvalidRootFormatsRejected(t *testing.T) {
	zone := testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1)
	zone.GenesisStateHash = "not-a-root"
	_, err := RegisterZone(EmptyState(), zone)
	require.ErrorContains(t, err, "genesis state hash")

	_, err = NewZoneCommitment(ZoneIDFinancial, 1, "not-a-root", hash("receipt"), hash("message"), hash("execution"), "")
	require.ErrorContains(t, err, "state root")
}

func TestNonCanonicalOrderingRejected(t *testing.T) {
	state := ZoneRegistryState{
		Zones: []Zone{
			testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1),
			testZone(ZoneIDApplication, ZoneKindApplication, VMPolicyAVM, 1),
		},
	}
	require.ErrorContains(t, state.Validate(), "sorted canonically")

	state = EmptyState()
	var err error
	state, err = RegisterZone(state, testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.NoError(t, err)
	state, err = RegisterZone(state, testZone(ZoneIDApplication, ZoneKindApplication, VMPolicyAVM, 1))
	require.NoError(t, err)
	financial := newCommitment(t, ZoneIDFinancial, 1, "")
	application := newCommitment(t, ZoneIDApplication, 1, "")
	state.Commitments = []ZoneCommitment{financial, application}
	require.ErrorContains(t, state.Validate(), "sorted canonically")
}

func TestCommitmentForUnknownZoneRejected(t *testing.T) {
	commitment := newCommitment(t, ZoneIDFinancial, 1, "")

	_, err := AppendCommitment(EmptyState(), commitment)
	require.ErrorContains(t, err, "not registered")
}

func TestDuplicateCommitmentRejected(t *testing.T) {
	state, err := RegisterZone(EmptyState(), testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	require.NoError(t, err)
	first := newCommitment(t, ZoneIDFinancial, 1, "")
	state, err = AppendCommitment(state, first)
	require.NoError(t, err)

	_, err = AppendCommitment(state, first)
	require.ErrorContains(t, err, "height must increase")
}

func testZone(id ZoneID, kind ZoneKind, vm VMPolicy, activationHeight uint64) Zone {
	return Zone{
		ID:			id,
		Kind:			kind,
		VMPolicy:		vm,
		FeePolicy:		FeePolicyNaet,
		GenesisStateHash:	hash(string(id) + "-genesis"),
		StateTransitionID:	"transition-" + string(id),
		UpgradePolicy:		UpgradePolicyGovernance,
		DataAvailabilityPolicy:	DataAvailabilityCoreCommitment,
		AuditStatus:		AuditStatusExperimental,
		ActivationHeight:	activationHeight,
	}
}

func newCommitment(t *testing.T, id ZoneID, height uint64, previous string) ZoneCommitment {
	t.Helper()
	commitment, err := NewZoneCommitment(
		id,
		height,
		hash(fmt.Sprintf("%s-state-%020d", id, height)),
		hash(fmt.Sprintf("%s-receipt-%020d", id, height)),
		hash(fmt.Sprintf("%s-message-%020d", id, height)),
		hash(fmt.Sprintf("%s-execution-%020d", id, height)),
		previous,
	)
	require.NoError(t, err)
	return commitment
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
