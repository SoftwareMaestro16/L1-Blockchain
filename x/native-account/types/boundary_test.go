package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultBoundariesValidateAndDeclareNativeAccountOwner(t *testing.T) {
	boundaries := DefaultBoundaries()

	require.NoError(t, ValidateBoundaries(boundaries))

	var native Boundary
	for _, boundary := range boundaries {
		if boundary.Path == ModulePath {
			native = boundary
			break
		}
	}
	require.Equal(t, ModulePath, native.Path)
	require.Contains(t, native.Owner, "native account state")
	require.Contains(t, native.Owner, "auth policy")
	require.Contains(t, native.OwnedState, "account/by_user")
	require.Contains(t, native.OwnedState, "account/storage")
	require.Contains(t, native.RejectedWrites, "private keys")
	require.Contains(t, native.RejectedWrites, "seed phrases")
	require.Contains(t, native.RejectedWrites, "token balances")
	require.Contains(t, native.RejectedWrites, "NFT inventories")
	require.Contains(t, native.RejectedWrites, "DEX positions")
}

func TestRejectedCrossModuleWritesCoverSecurityBoundaries(t *testing.T) {
	require.True(t, IsRejectedCrossModuleWrite("app/addressing", ModulePath, "account state"))
	require.True(t, IsRejectedCrossModuleWrite("x/identity", ModulePath, "auth policy"))
	require.True(t, IsRejectedCrossModuleWrite("x/storage-rent", ModulePath, "automatic wallet deletion"))
	require.True(t, IsRejectedCrossModuleWrite("x/storage-rent", "protocol-critical/system state", "rent freeze"))
	require.True(t, IsRejectedCrossModuleWrite("x/nominator-pool", "x/validator-*", "user-selected validator delegation"))
	require.True(t, IsRejectedCrossModuleWrite("x/fees", ModulePath, "duplicated wallet balance"))
	require.True(t, IsRejectedCrossModuleWrite("x/contracts, x/vm, x/aetravm/*", ModulePath, "sequence bypass"))

	require.False(t, IsRejectedCrossModuleWrite(ModulePath, "x/reputation", "reputation id reference"))
}

func TestTokenNFTAndDEXBehaviorRemainContractRouted(t *testing.T) {
	routes := DefaultAssetRoutes()

	require.NoError(t, ValidateAssetRoutes(routes))

	for _, route := range routes {
		switch route.Behavior {
		case "fungible token", "NFT", "DEX":
			require.False(t, route.NativeModuleAllowed, route.Behavior)
			require.Contains(t, strings.ToLower(route.Route), "contract", route.Behavior)
		}
	}
	require.NoError(t, ValidateNoNativeAssetModules([]string{"auth", "bank", "staking", ModuleName, "fees"}))
	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"auth", "tokenfactory"}), "native asset module tokenfactory is not allowed")
	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"dex"}), "native asset module dex is not allowed")
	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"nft"}), "native asset module nft is not allowed")
}

func TestBoundaryManifestIsDeterministic(t *testing.T) {
	lines := BoundaryManifestLines()
	require.NotEmpty(t, lines)

	hash := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	require.Equal(t, "dcbb084c6986c3039f54d72e51fc4ca6f4ba61ca18fd6e1f7d993d158c9090f2", hex.EncodeToString(hash[:]))
}
