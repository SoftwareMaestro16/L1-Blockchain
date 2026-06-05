package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestDenomsQueryReturnsEmptyState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	res, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{})
	require.NoError(t, err)
	require.Empty(t, res.Denoms)
	require.NotNil(t, res.Pagination)
	require.Empty(t, res.Pagination.NextKey)
}

func TestDenomQueryReturnsMetadata(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "querygold",
	})
	require.NoError(t, err)

	res, err := app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: createRes.NewTokenDenom})
	require.NoError(t, err)
	require.Equal(t, createRes.NewTokenDenom, res.Metadata.Denom)
	require.Equal(t, orbitaladdress.FormatAccAddress(admin), res.Metadata.Admin)
}

func TestDenomQueryErrorsAreGrpcStatusCompatible(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	missingDenom := fmt.Sprintf("factory/%s/missing", admin.String())

	_, err := app.TokenFactoryKeeper.Denom(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: "bad denom"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: missingDenom})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestDenomsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.TokenFactoryKeeper.Denoms(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDenomsQueryPaginatesWithNextKey(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]

	denoms := seedDenoms(t, app, ctx, admin.String(), 5)

	first, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &sdkquery.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, first.Denoms, 2)
	require.Equal(t, denoms[:2], []string{first.Denoms[0].Denom, first.Denoms[1].Denom})
	require.NotEmpty(t, first.Pagination.NextKey)

	next, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &sdkquery.PageRequest{Limit: 2, Key: first.Pagination.NextKey},
	})
	require.NoError(t, err)
	require.Len(t, next.Denoms, 2)
	require.Equal(t, denoms[2:4], []string{next.Denoms[0].Denom, next.Denoms[1].Denom})
	require.NotEmpty(t, next.Pagination.NextKey)
}

func TestDenomsQueryDefaultLimitIsBoundedOnLargeState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	seedDenoms(t, app, ctx, admin.String(), types.MaxQueryDenoms+1)

	res, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Denoms, types.DefaultQueryDenoms)
	require.NotEmpty(t, res.Pagination.NextKey)
}

func TestDenomsQueryRejectsInvalidPagination(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	cases := []sdkquery.PageRequest{
		{Limit: types.MaxQueryDenoms + 1},
		{Key: []byte("bad-key")},
		{Offset: 1},
		{CountTotal: true},
		{Reverse: true},
	}

	for _, tc := range cases {
		_, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{Pagination: &tc})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
}

func seedDenoms(t *testing.T, app *l1app.L1App, ctx sdk.Context, admin string, count int) []string {
	t.Helper()
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		denom := fmt.Sprintf("factory/%s/querygold%03d", admin, i)
		require.NoError(t, app.TokenFactoryKeeper.SetDenom(ctx, types.DenomAuthorityMetadata{
			Denom: denom,
			Admin: admin,
		}))
		out = append(out, denom)
	}
	return out
}
