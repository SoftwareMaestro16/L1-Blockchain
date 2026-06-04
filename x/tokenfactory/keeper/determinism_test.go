package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestGenesisImportOrderDoesNotAffectExportedDenomOrder(t *testing.T) {
	admin := "orb1p9gs69x7h00a62tq30j8tg54g3cx2cghw3q3f8"
	forward := types.GenesisState{Denoms: []types.DenomAuthorityMetadata{
		{Denom: "factory/" + admin + "/silver", Admin: admin},
		{Denom: "factory/" + admin + "/gold", Admin: admin},
	}}
	reversed := types.GenesisState{Denoms: []types.DenomAuthorityMetadata{
		forward.Denoms[1],
		forward.Denoms[0],
	}}

	exportForward := exportTokenfactoryGenesis(t, forward)
	exportReversed := exportTokenfactoryGenesis(t, reversed)

	require.Equal(t, exportForward, exportReversed)
	require.Equal(t, "factory/"+admin+"/gold", exportForward.Denoms[0].Denom)
	require.Equal(t, "factory/"+admin+"/silver", exportForward.Denoms[1].Denom)
}

func exportTokenfactoryGenesis(t *testing.T, gs types.GenesisState) *types.GenesisState {
	t.Helper()

	l1 := l1app.Setup(t, false)
	ctx := l1.NewContext(false)
	l1.TokenFactoryKeeper.InitGenesis(ctx, gs)
	return l1.TokenFactoryKeeper.ExportGenesis(ctx)
}
