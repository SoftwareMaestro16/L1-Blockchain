package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	configtypes "github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/constitution/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestOrdinaryConfigChangeFailsOutsideConstitutionalBounds(t *testing.T) {
	k := NewKeeper()
	err := k.ValidateOrdinaryConfigChange(configtypes.ConfigChange{
		ID:		"bad-block-gas",
		Key:		configtypes.KeyConsensusMaxBlockGas,
		Value:		"1000000001",
		Operation:	configtypes.OperationSet,
		Status:		configtypes.ChangeStatusPending,
		SubmittedBy:	prototype.DefaultAuthority,
	})
	require.ErrorContains(t, err, "constitutional max block gas")
}

func TestConstitutionalUpdateRequiresSpecialFlowAndDelay(t *testing.T) {
	k := NewKeeper()
	proposed := types.DefaultConstitution().Normalize()
	proposed.MaxBlockGas = proposed.MaxBlockGas + 1

	amendment, err := k.ProposeConstitutionAmendment(types.MsgProposeConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		Amendment: types.Amendment{
			ID:		"raise-block-gas",
			Proposed:	proposed,
			Reason:		"capacity",
		},
	}, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(110), amendment.ExecutableHeight)

	_, _, err = k.ExecuteConstitutionAmendment(types.MsgExecuteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-block-gas",
	}, 109)
	require.ErrorContains(t, err, "approved")

	_, err = k.VoteConstitutionAmendment(types.MsgVoteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-block-gas",
		Support:	types.VoteSupportYes,
		VotingPowerBps:	6_700,
	}, 20)
	require.NoError(t, err)

	_, _, err = k.ExecuteConstitutionAmendment(types.MsgExecuteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-block-gas",
	}, 109)
	require.ErrorContains(t, err, "delay")

	updated, executed, err := k.ExecuteConstitutionAmendment(types.MsgExecuteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-block-gas",
	}, 110)
	require.NoError(t, err)
	require.Equal(t, types.AmendmentStatusExecuted, executed.Status)
	require.Equal(t, proposed.MaxBlockGas, updated.MaxBlockGas)
}

func TestEmergencyPauseExpiresAutomatically(t *testing.T) {
	k := NewKeeper()
	require.NoError(t, k.ActivateEmergencyPause(prototype.DefaultAuthority, 10, 5))
	require.True(t, k.Constitution().IsEmergencyPaused(15))
	require.False(t, k.ExpireEmergencyPause(15))
	require.True(t, k.ExpireEmergencyPause(16))
	require.False(t, k.Constitution().IsEmergencyPaused(16))
	require.ErrorContains(t, k.ActivateEmergencyPause(prototype.DefaultAuthority, 20, 2_000), "duration")
}

func TestExportImportPreservesAmendmentQueue(t *testing.T) {
	source := NewKeeper()
	for _, id := range []string{"z", "a"} {
		_, err := source.ProposeConstitutionAmendment(types.MsgProposeConstitutionAmendment{
			Authority:	prototype.DefaultAuthority,
			Amendment: types.Amendment{
				ID:		id,
				Proposed:	types.DefaultConstitution().Normalize(),
			},
		}, 1)
		require.NoError(t, err)
	}

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Equal(t, []string{"a", "z"}, []string{exported.State.PendingAmendments[0].ID, exported.State.PendingAmendments[1].ID})

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func TestMaliciousAuthorityCannotBypassProtectedModuleList(t *testing.T) {
	k := NewKeeper()
	err := k.ValidateOrdinaryConfigChange(configtypes.ConfigChange{
		ID:		"disable-config",
		Key:		"module/enabled/config",
		Value:		"false",
		Operation:	configtypes.OperationSet,
		Status:		configtypes.ChangeStatusPending,
		SubmittedBy:	prototype.DefaultAuthority,
	})
	require.ErrorContains(t, err, "protected modules")

	_, err = k.ProposeConstitutionAmendment(types.MsgProposeConstitutionAmendment{
		Authority:	"4:0000000000000000000000000000000000000000000000000000000000000002",
		Amendment: types.Amendment{
			ID:		"malicious",
			Proposed:	types.DefaultConstitution().Normalize(),
		},
	}, 1)
	require.ErrorContains(t, err, "governance authority")
}
