package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestStateInitHashAndAddressAreDeterministic(t *testing.T) {
	params := DefaultParams()
	owner := stateInitAddress(0x11)
	codeHash := stateInitHash("code")
	left := StateInit{
		ABIVersion:		1,
		CodeID:			codeHash,
		CodeHash:		codeHash,
		InitData:		[]byte("init"),
		Salt:			"salt",
		Owner:			owner,
		InitialStorageRoot:	DefaultInitialStorageRoot,
		InitialBalanceNAET:	100,
		Libraries: []CodeDependency{
			{CodeID: "z", CodeHash: stateInitHash("z")},
			{CodeID: "a", CodeHash: stateInitHash("a")},
		},
		Capabilities:	[]string{"write", "read", "read"},
	}
	right := left
	right.Libraries = []CodeDependency{left.Libraries[1], left.Libraries[0]}
	right.Capabilities = []string{"read", "write"}

	leftHash, err := HashStateInit(left)
	require.NoError(t, err)
	rightHash, err := HashStateInit(right)
	require.NoError(t, err)
	require.Equal(t, leftHash, rightHash)

	leftUser, leftRaw, err := DeriveContractAddressFromStateInit("aetra-test", "zone-1", owner, left, params)
	require.NoError(t, err)
	rightUser, rightRaw, err := DeriveContractAddressFromStateInit("aetra-test", "zone-1", owner, right, params)
	require.NoError(t, err)
	require.Equal(t, leftUser, rightUser)
	require.Equal(t, leftRaw, rightRaw)
	require.True(t, strings.HasPrefix(leftUser, "AE"))
	require.True(t, strings.HasPrefix(leftRaw, "4:"))
	require.NoError(t, ValidateAddressPair("derived contract", leftUser, leftRaw))
}

func TestStateInitAddressChangesWithInitDataAndSalt(t *testing.T) {
	params := DefaultParams()
	owner := stateInitAddress(0x22)
	codeHash := stateInitHash("code")
	base := NewStateInit(owner, codeHash, []byte("init-a"), "salt-a", 0)

	baseAddress, _, err := DeriveContractAddressFromStateInit("", "", owner, base, params)
	require.NoError(t, err)
	changedInit := base
	changedInit.InitData = []byte("init-b")
	initAddress, _, err := DeriveContractAddressFromStateInit("", "", owner, changedInit, params)
	require.NoError(t, err)
	require.NotEqual(t, baseAddress, initAddress)

	changedSalt := base
	changedSalt.Salt = "salt-b"
	saltAddress, _, err := DeriveContractAddressFromStateInit("", "", owner, changedSalt, params)
	require.NoError(t, err)
	require.NotEqual(t, baseAddress, saltAddress)
}

func TestStateInitValidationRejectsBoundsAndMalformedInputs(t *testing.T) {
	params := DefaultParams()
	owner := stateInitAddress(0x33)
	codeHash := stateInitHash("code")
	oversized := NewStateInit(owner, codeHash, make([]byte, params.MaxInitDataBytes+1), "salt", 0)
	require.ErrorContains(t, oversized.Validate(params), "init data")

	badHash := NewStateInit(owner, "not-hex", nil, "salt", 0)
	require.ErrorContains(t, badHash.Validate(params), "code hash")

	tooManyDeps := NewStateInit(owner, codeHash, nil, "salt", 0)
	for i := uint32(0); i < params.MaxStateInitDependencies+1; i++ {
		tooManyDeps.Libraries = append(tooManyDeps.Libraries, CodeDependency{
			CodeID:		stateInitHash(string(rune('a' + i))),
			CodeHash:	stateInitHash(string(rune('z' + i))),
		})
	}
	require.ErrorContains(t, tooManyDeps.Validate(params), "dependency count")

	_, _, err := DeriveContractAddress("", "", addressing.ZeroUserFriendly, codeHash, NewStateInit(owner, codeHash, nil, "salt", 0).InitDataHash(), []byte("salt"))
	require.ErrorContains(t, err, "zero address")
	_, _, err = DeriveContractAddress("", "", owner, "not-hex", NewStateInit(owner, codeHash, nil, "salt", 0).InitDataHash(), []byte("salt"))
	require.ErrorContains(t, err, "code hash")
}

func stateInitAddress(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return addressing.FormatAccAddress(sdk.AccAddress(bz))
}

func stateInitHash(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}
