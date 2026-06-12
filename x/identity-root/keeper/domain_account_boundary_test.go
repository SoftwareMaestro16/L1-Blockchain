package keeper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	identitytypes "github.com/sovereign-l1/l1/x/identity-root/types"
)

func TestAccountMetadataReferencesDomainAliasWithoutOwningDomainRecord(t *testing.T) {
	owner := ownerA
	account := accountExportFixture{
		AddressUser:	owner,
		Metadata: accountMetadataFixture{
			DomainAlias:	"alice.aet",
			CreatedHeight:	10,
		},
	}

	bz, err := json.Marshal(account)
	require.NoError(t, err)
	require.Contains(t, string(bz), `"domain_alias":"alice.aet"`)
	require.NotContains(t, string(bz), `"owner"`)
	require.NotContains(t, string(bz), `"resolver_root"`)
	require.NotContains(t, string(bz), `"expiry_height"`)

	registry := setupKeeper(t)
	_, err = registry.RegisterName(identitytypes.MsgRegisterName{Owner: owner, Name: "alice", Height: 10})
	require.NoError(t, err)
	exported := registry.ExportGenesis()
	require.Len(t, exported.State.Records, 1)
	require.Equal(t, owner, exported.State.Records[0].Owner)
}

func TestDomainOwnerQueryDoesNotReadAccountMetadata(t *testing.T) {
	registry := setupKeeper(t)
	_, err := registry.RegisterName(identitytypes.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)
	record, found, err := registry.NameRecord("alice.aet")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ownerA, record.Owner)
	require.Empty(t, record.NFTBinding.ClassID)
}

type accountExportFixture struct {
	AddressUser	string			`json:"address_user"`
	Metadata	accountMetadataFixture	`json:"metadata,omitempty"`
}

type accountMetadataFixture struct {
	DomainAlias	string	`json:"domain_alias,omitempty"`
	CreatedHeight	uint64	`json:"created_height,omitempty"`
}
