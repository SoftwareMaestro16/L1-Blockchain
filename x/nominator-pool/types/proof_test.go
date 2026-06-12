package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
)

func TestStakingProofMetadataStableForPoolObjects(t *testing.T) {
	account := proofAEAddress(0x44)
	appHash := "app-root-ref"
	rootHash := "nominator-pool-root-ref"

	cases := []struct {
		name		string
		req		StakingProofRequest
		storeKey	string
		stateKey	string
		hash		string
	}{
		{
			name:		"deposit",
			req:		StakingProofRequest{Kind: StakingProofDeposit, Height: 10, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: rootHash},
			storeKey:	StoreKey,
			stateKey:	PoolDepositProofStateKey("pool-a", account),
			hash:		"fb2b1fd776560789390b65940072a55ae1a7411e8bf7847750917c3dcb8737f7",
		},
		{
			name:		"share",
			req:		StakingProofRequest{Kind: StakingProofShare, Height: 10, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: rootHash},
			storeKey:	StoreKey,
			stateKey:	PoolShareProofStateKey("pool-a", account),
			hash:		"c2077ec0a74aec831500eecb163ecafecd4401ca1ab5ff1cffc4d67c36f0df5f",
		},
		{
			name:		"allocation",
			req:		StakingProofRequest{Kind: StakingProofAllocation, Height: 10, PoolID: "pool-a", Epoch: 7, AppHash: appHash, RootHash: rootHash},
			storeKey:	StoreKey,
			stateKey:	PoolAllocationProofStateKey("pool-a", 7),
			hash:		"801f5215804d3b647d3b299e97d742021b25316986e75a0a88132a663051e2f0",
		},
		{
			name:		"reward",
			req:		StakingProofRequest{Kind: StakingProofReward, Height: 10, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: rootHash},
			storeKey:	StoreKey,
			stateKey:	PoolRewardProofStateKey("pool-a", account),
			hash:		"6c43015042f5f4914ecc38eda532fdf32a07d5e9a426b85fd574c4a4ddeefce1",
		},
		{
			name:		"reputation",
			req:		StakingProofRequest{Kind: StakingProofReputation, Height: 10, Account: account, AppHash: appHash, RootHash: "reputation-root-ref"},
			storeKey:	reputationtypes.StoreKey,
			stateKey:	StakeReputationProofStateKey(account),
			hash:		"48fc1bec057d30bc2b351d7f9a82c0473aa3aea7357dd898512cfb5cb457f874",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metadata, err := BuildStakingProofMetadata(tc.req)
			require.NoError(t, err)
			require.Equal(t, tc.storeKey, metadata.StoreKey)
			require.Equal(t, tc.stateKey, metadata.StateKey)
			require.Equal(t, tc.req.Height, metadata.Height)
			require.Equal(t, tc.req.AppHash, metadata.AppHash)
			require.Equal(t, tc.req.RootHash, metadata.RootHash)
			require.Len(t, metadata.ProofPath, 2)
			require.True(t, metadata.BoundedLookup)
			require.Equal(t, tc.hash, metadata.MetadataHash)
		})
	}
}

func TestStakingProofRejectsMissingRootMetadataAndUnboundedFlag(t *testing.T) {
	_, err := BuildStakingProofMetadata(StakingProofRequest{Kind: StakingProofShare, Height: 1, PoolID: "pool-a", Account: proofAEAddress(0x45)})
	require.ErrorContains(t, err, "app hash and root hash")

	metadata, err := BuildStakingProofMetadata(StakingProofRequest{
		Kind:		StakingProofShare,
		Height:		1,
		PoolID:		"pool-a",
		Account:	proofAEAddress(0x45),
		AppHash:	"app",
		RootHash:	"root",
	})
	require.NoError(t, err)
	metadata.BoundedLookup = false
	metadata.MetadataHash = ComputeStakingProofMetadataHash(metadata)
	require.ErrorContains(t, metadata.Validate(), "bounded")
}

func proofAEAddress(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bz))
}
