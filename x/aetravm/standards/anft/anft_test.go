package anft

import (
	"bytes"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCreateCollectionMintNFTAndVerifyItemAddress(t *testing.T) {
	admin := testAddr(1)
	owner := testAddr(2)
	state := newTestState(t, admin)

	item, err := state.MintNFT(admin, owner, testMetadata("Genesis NFT"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), item.Index)
	require.Equal(t, owner, item.Owner)
	require.Equal(t, ItemKindNFT, item.Kind)
	require.True(t, item.Transferable)

	expected, err := DeriveItemAddress(state.Collection.Address, item.Index, state.Collection.ItemCodeHash)
	require.NoError(t, err)
	require.Equal(t, expected, item.Address)

	ok, err := state.ProveItemBelongsToCollection(item.Address, item.Index)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, uint64(1), state.Collection.NextItemIndex)
}

func TestCollectionRoyaltyPolicyBounded(t *testing.T) {
	collection := testCollection(testAddr(1))
	collection.RoyaltyPolicy = RoyaltyPolicy{
		Enabled:     true,
		Recipient:   testAddr(5),
		BasisPoints: MaxRoyaltyBasisPoints,
	}
	_, err := NewState(collection)
	require.NoError(t, err)

	collection.RoyaltyPolicy.BasisPoints = MaxRoyaltyBasisPoints + 1
	_, err = NewState(collection)
	require.ErrorContains(t, err, "royalty basis points")

	collection.RoyaltyPolicy = RoyaltyPolicy{Enabled: true, Recipient: sdk.AccAddress(make([]byte, 20)), BasisPoints: 100}
	_, err = NewState(collection)
	require.ErrorContains(t, err, "must not be zero")

	collection.RoyaltyPolicy = RoyaltyPolicy{Enabled: false, Recipient: testAddr(5)}
	_, err = NewState(collection)
	require.ErrorContains(t, err, "disabled royalty policy")
}

func TestTransferNFTRequiresCurrentOwner(t *testing.T) {
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	attacker := testAddr(4)
	state := newTestState(t, admin)

	item, err := state.MintNFT(admin, alice, testMetadata("Transferable NFT"))
	require.NoError(t, err)
	require.ErrorContains(t, state.TransferNFT(attacker, item.Address, bob), "current owner")
	require.NoError(t, state.TransferNFT(alice, item.Address, bob))

	owner, err := state.RequestCurrentOwner(item.Address)
	require.NoError(t, err)
	require.Equal(t, bob, owner)
	require.NoError(t, state.ValidateCollectionMembership())
}

func TestMalformedCollectionAddressRejected(t *testing.T) {
	_, err := DeriveItemAddress(sdk.AccAddress(make([]byte, 20)), 0, testCodeHash())
	require.ErrorContains(t, err, "must not be zero")

	collection := testCollection(testAddr(1))
	collection.Address = sdk.AccAddress(make([]byte, 20))
	_, err = NewState(collection)
	require.ErrorContains(t, err, "must not be zero")
}

func TestSBTMintTransferProofAndRevoke(t *testing.T) {
	admin := testAddr(1)
	owner := testAddr(2)
	authority := testAddr(3)
	newOwner := testAddr(4)
	state := newTestState(t, admin)

	item, err := state.MintSBT(admin, owner, authority, testMetadata("Credential SBT"))
	require.NoError(t, err)
	require.Equal(t, ItemKindSBT, item.Kind)
	require.Equal(t, owner, item.ImmutableOwner)
	require.False(t, item.Transferable)

	require.ErrorContains(t, state.TransferSBT(owner, item.Address, newOwner), "SBT transfer must be rejected")
	require.ErrorContains(t, state.TransferNFT(owner, item.Address, newOwner), "SBT transfer must be rejected")

	proof, err := state.ProveSBTOwnership(item.Address, owner)
	require.NoError(t, err)
	require.Equal(t, owner, proof.Owner)
	require.False(t, proof.Revoked)

	require.NoError(t, state.RevokeSBT(authority, item.Address, 1_700_000_000, "credential revoked"))
	proof, err = state.ProveSBTOwnership(item.Address, owner)
	require.NoError(t, err)
	require.True(t, proof.Revoked)

	itemAfterRevoke := state.Items[string(item.Address)]
	require.Equal(t, owner, itemAfterRevoke.Owner)
	require.Equal(t, owner, itemAfterRevoke.ImmutableOwner)
	require.Equal(t, int64(1_700_000_000), itemAfterRevoke.RevokedAt)
}

func TestUnauthorizedSBTRevokeRejected(t *testing.T) {
	admin := testAddr(1)
	owner := testAddr(2)
	authority := testAddr(3)
	attacker := testAddr(4)
	state := newTestState(t, admin)

	item, err := state.MintSBT(admin, owner, authority, testMetadata("Credential SBT"))
	require.NoError(t, err)
	require.ErrorContains(t, state.RevokeSBT(attacker, item.Address, 1, "bad"), "only SBT authority")
	require.Equal(t, int64(0), state.Items[string(item.Address)].RevokedAt)
}

func TestMetadataSpoofingRejected(t *testing.T) {
	tests := []Metadata{
		{Name: "Aetra", Symbol: "COLL", ContentRef: "ipfs://collection"},
		{Name: "Wrapped Asset", Symbol: "AET", ContentRef: "ipfs://collection"},
		{Name: "Wrapped Asset", Symbol: "naet", ContentRef: "ipfs://collection"},
	}

	for _, metadata := range tests {
		t.Run(metadata.Name+"-"+metadata.Symbol, func(t *testing.T) {
			collection := testCollection(testAddr(1))
			collection.Metadata = metadata
			_, err := NewState(collection)
			require.ErrorContains(t, err, "must not spoof native AET/naet")
		})
	}
}

func TestBatchMintBoundedAndAtomic(t *testing.T) {
	admin := testAddr(1)
	state := newTestState(t, admin)

	owners := []sdk.AccAddress{testAddr(2), testAddr(3)}
	metadata := []Metadata{testMetadata("NFT 1"), testMetadata("NFT 2")}
	items, err := state.BatchMintNFT(admin, owners, metadata)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, uint64(2), state.Collection.NextItemIndex)

	tooManyOwners := make([]sdk.AccAddress, MaxBatchMintCount+1)
	tooManyMetadata := make([]Metadata, MaxBatchMintCount+1)
	for i := range tooManyOwners {
		tooManyOwners[i] = testAddr(byte(i + 10))
		tooManyMetadata[i] = testMetadata("Batch NFT")
	}
	_, err = state.BatchMintNFT(admin, tooManyOwners, tooManyMetadata)
	require.ErrorContains(t, err, "batch mint count")
	require.Equal(t, uint64(2), state.Collection.NextItemIndex)

	before := state.Collection.NextItemIndex
	_, err = state.BatchMintNFT(admin, []sdk.AccAddress{testAddr(5), testAddr(6)}, []Metadata{
		testMetadata("Good NFT"),
		{Name: strings.Repeat("x", MaxNameLength+1), ContentRef: "ipfs://bad"},
	})
	require.ErrorContains(t, err, "metadata name length")
	require.Equal(t, before, state.Collection.NextItemIndex)
	require.Len(t, state.Items, 2)
}

func TestMalformedANFTMessagesRejected(t *testing.T) {
	admin := testAddr(1)
	owner := testAddr(2)
	state := newTestState(t, admin)

	_, err := DeriveItemAddress(state.Collection.Address, 0, []byte{1})
	require.ErrorContains(t, err, "item code hash")
	_, err = state.MintNFT(testAddr(9), owner, testMetadata("Nope"))
	require.ErrorContains(t, err, "only collection admin")
	_, err = state.MintNFT(admin, sdk.AccAddress(make([]byte, 20)), testMetadata("Nope"))
	require.ErrorContains(t, err, "must not be zero")
	_, err = state.MintNFT(admin, owner, Metadata{Name: "", ContentRef: "ipfs://empty"})
	require.ErrorContains(t, err, "metadata name")
	_, err = state.MintSBT(admin, owner, sdk.AccAddress(make([]byte, 20)), testMetadata("Nope"))
	require.ErrorContains(t, err, "must not be zero")
}

func TestCollectionMetadataAuthAndBounds(t *testing.T) {
	admin := testAddr(1)
	state := newTestState(t, admin)

	require.ErrorContains(t, state.ChangeCollectionMetadata(testAddr(9), testMetadata("Updated")), "only collection admin")
	require.NoError(t, state.ChangeCollectionMetadata(admin, testMetadata("Updated")))

	state.Collection.MutableMetadata = false
	require.ErrorContains(t, state.ChangeCollectionMetadata(admin, testMetadata("Again")), "immutable")
}

func newTestState(t testing.TB, admin sdk.AccAddress) *State {
	t.Helper()
	state, err := NewState(testCollection(admin))
	require.NoError(t, err)
	return state
}

func testCollection(admin sdk.AccAddress) CollectionState {
	return CollectionState{
		Address:         testAddr(8),
		Admin:           admin,
		Metadata:        testMetadata("Aetra Artifacts"),
		NextItemIndex:   0,
		ItemCodeHash:    testCodeHash(),
		StandardVersion: DefaultVersion,
		MutableMetadata: true,
	}
}

func testMetadata(name string) Metadata {
	return Metadata{
		Name:       name,
		Symbol:     "ART",
		ContentRef: "ipfs://bafy-anft66",
	}
}

func testAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}

func testCodeHash() []byte {
	return bytes.Repeat([]byte{7}, ItemCodeHashLength)
}
