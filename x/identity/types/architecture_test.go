package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultIdentityV2ArchitectureValidatesAndOrdersModules(t *testing.T) {
	arch := DefaultIdentityV2Architecture()
	require.NoError(t, ValidateIdentityV2Architecture(arch))

	order, err := IdentityV2ExecutionOrder(arch)
	require.NoError(t, err)
	require.Equal(t, []IdentityV2Component{
		IdentityV2Core,
		IdentityV2Resolver,
		IdentityV2Subdomain,
		IdentityV2Auction,
		IdentityV2Proof,
		IdentityV2RoutingIntegration,
	}, order)

	hash, err := IdentityV2ArchitectureHash(arch)
	require.NoError(t, err)
	require.Len(t, hash, 64)
	repeated, err := IdentityV2ArchitectureHash(DefaultIdentityV2Architecture())
	require.NoError(t, err)
	require.Equal(t, hash, repeated)
}

func TestIdentityV2ArchitectureStoreOwnership(t *testing.T) {
	arch := DefaultIdentityV2Architecture()

	resolverKey, err := IdentityResolverStoreKey("api.alice.aet")
	require.NoError(t, err)
	owner, found, err := IdentityV2ComponentForStoreKey(arch, resolverKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, IdentityV2Resolver, owner)

	domainKey, err := IdentityDomainStoreKey("alice.aet")
	require.NoError(t, err)
	owner, found, err = IdentityV2ComponentForStoreKey(arch, domainKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, IdentityV2Core, owner)

	auctionKey, err := IdentityAuctionStoreKey("alice.aet")
	require.NoError(t, err)
	owner, found, err = IdentityV2ComponentForStoreKey(arch, auctionKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, IdentityV2Auction, owner)

	_, found, err = IdentityV2ComponentForStoreKey(arch, IdentityStoreV2Prefix+"/unknown/key")
	require.NoError(t, err)
	require.False(t, found)
}

func TestIdentityV2ArchitectureRejectsDependencyCycle(t *testing.T) {
	arch := DefaultIdentityV2Architecture()
	for i := range arch.Modules {
		if arch.Modules[i].Component == IdentityV2Core {
			arch.Modules[i].DependsOn = []IdentityV2Component{IdentityV2RoutingIntegration}
		}
	}

	require.ErrorContains(t, ValidateIdentityV2Architecture(arch), "cycle")
}

func TestIdentityV2ArchitectureRejectsDuplicatePrefixOwner(t *testing.T) {
	arch := DefaultIdentityV2Architecture()
	for i := range arch.Modules {
		if arch.Modules[i].Component == IdentityV2Resolver {
			arch.Modules[i].OwnsPrefixes = append(arch.Modules[i].OwnsPrefixes, IdentityStoreV2DomainPrefix)
		}
	}

	require.ErrorContains(t, ValidateIdentityV2Architecture(arch), "owned by both")
}

func TestIdentityV2ArchitectureRequiresProofAndRoutingEdges(t *testing.T) {
	arch := DefaultIdentityV2Architecture()
	for i := range arch.Modules {
		if arch.Modules[i].Component == IdentityV2Proof {
			arch.Modules[i].DependsOn = []IdentityV2Component{IdentityV2Core, IdentityV2Resolver, IdentityV2Auction}
		}
	}
	require.ErrorContains(t, ValidateIdentityV2Architecture(arch), "proof module")

	arch = DefaultIdentityV2Architecture()
	for i := range arch.Modules {
		if arch.Modules[i].Component == IdentityV2RoutingIntegration {
			arch.Modules[i].DependsOn = []IdentityV2Component{IdentityV2Resolver}
		}
	}
	require.ErrorContains(t, ValidateIdentityV2Architecture(arch), "routing integration")
}
