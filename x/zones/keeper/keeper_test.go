package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Zones)
}

func TestInitGenesisRejectsCorruptedState(t *testing.T) {
	gs := DefaultGenesis()
	gs.Version = prototype.CurrentGenesisVersion + 1

	keeper := NewKeeper()
	require.ErrorContains(t, keeper.InitGenesis(gs), "unsupported genesis version")

	gs = DefaultGenesis()
	zone := keeperZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule, 1)
	gs.State.Zones = []zonestypes.Zone{zone, zone}
	require.ErrorContains(t, keeper.InitGenesis(gs), "duplicate zone")
}

func TestFeatureDisabledRejectsMutatingMessages(t *testing.T) {
	keeper := NewKeeper()

	err := keeper.RegisterZone(keeperZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule, 1))
	require.ErrorContains(t, err, "disabled")
}

func TestUnauthorizedAuthorityRejected(t *testing.T) {
	keeper := NewKeeper()

	err := keeper.UpdateParams("4:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", prototype.TestnetParams())
	require.ErrorContains(t, err, "authority")
}

func TestRegisterActivateAndAppendCommitment(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))

	err := keeper.RegisterZone(keeperZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule, 10))
	require.NoError(t, err)
	require.ErrorContains(t, keeper.ActivateZone(zonestypes.ZoneIDFinancial, 9), "before height 10")
	require.NoError(t, keeper.ActivateZone(zonestypes.ZoneIDFinancial, 10))

	first := keeperCommitment(t, zonestypes.ZoneIDFinancial, 1, "")
	require.NoError(t, keeper.AppendCommitment(first))
	require.NoError(t, keeper.AppendCommitment(keeperCommitment(t, zonestypes.ZoneIDFinancial, 2, first.CommitmentHash)))

	exported := keeper.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Equal(t, []zonestypes.ZoneID{zonestypes.ZoneIDFinancial}, exported.State.ActiveZones)
	require.Len(t, exported.State.Commitments, 2)
}

func TestZonesQueryPaginationBoundedAndMalformedSafe(t *testing.T) {
	keeper := NewKeeper()
	params := prototype.TestnetParams()
	params.DefaultQueryLimit = 1
	params.MaxQueryLimit = 2
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, params))

	require.NoError(t, keeper.RegisterZone(keeperZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule, 1)))
	require.NoError(t, keeper.RegisterZone(keeperZone(zonestypes.ZoneIDContract, zonestypes.ZoneKindContract, zonestypes.VMPolicyCosmWasmGated, 1)))
	require.NoError(t, keeper.RegisterZone(keeperZone(zonestypes.ZoneIDApplication, zonestypes.ZoneKindApplication, zonestypes.VMPolicyAVM, 1)))

	first, page, err := keeper.Zones(nil)
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.NotZero(t, page.NextOffset)

	second, page, err := keeper.Zones(&prototype.PageRequest{Offset: page.NextOffset, Limit: 2})
	require.NoError(t, err)
	require.Len(t, second, 2)
	require.Zero(t, page.NextOffset)

	_, _, err = keeper.Zones(&prototype.PageRequest{Offset: 99, Limit: 1})
	require.ErrorContains(t, err, "offset")

	_, _, err = keeper.Zones(&prototype.PageRequest{Limit: params.MaxQueryLimit + 1})
	require.ErrorContains(t, err, "limit")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	require.NoError(t, keeper.RegisterZone(keeperZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule, 1)))
	require.NoError(t, keeper.ActivateZone(zonestypes.ZoneIDFinancial, 1))
	first := keeperCommitment(t, zonestypes.ZoneIDFinancial, 1, "")
	require.NoError(t, keeper.AppendCommitment(first))

	exported := keeper.ExportGenesis()
	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))

	require.Equal(t, exported, imported.ExportGenesis())
}

func TestMigrationNoopPathValidatesCurrentState(t *testing.T) {
	keeper := NewKeeper()

	err := NewMigrator(&keeper).Migrate1to2()
	require.NoError(t, err)
}

func keeperZone(id zonestypes.ZoneID, kind zonestypes.ZoneKind, vm zonestypes.VMPolicy, activationHeight uint64) zonestypes.Zone {
	return zonestypes.Zone{
		ID:			id,
		Kind:			kind,
		VMPolicy:		vm,
		FeePolicy:		zonestypes.FeePolicyNaet,
		GenesisStateHash:	keeperHash(string(id) + "-genesis"),
		StateTransitionID:	"transition-" + string(id),
		UpgradePolicy:		zonestypes.UpgradePolicyGovernance,
		DataAvailabilityPolicy:	zonestypes.DataAvailabilityCoreCommitment,
		AuditStatus:		zonestypes.AuditStatusExperimental,
		ActivationHeight:	activationHeight,
	}
}

func keeperCommitment(t *testing.T, id zonestypes.ZoneID, height uint64, previous string) zonestypes.ZoneCommitment {
	t.Helper()
	commitment, err := zonestypes.NewZoneCommitment(
		id,
		height,
		keeperHash(fmt.Sprintf("%s-state-%020d", id, height)),
		keeperHash(fmt.Sprintf("%s-receipt-%020d", id, height)),
		keeperHash(fmt.Sprintf("%s-message-%020d", id, height)),
		keeperHash(fmt.Sprintf("%s-execution-%020d", id, height)),
		previous,
	)
	require.NoError(t, err)
	return commitment
}

func keeperHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
