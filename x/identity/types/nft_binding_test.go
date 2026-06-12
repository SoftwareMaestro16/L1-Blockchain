package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainNFTBindingRequiresRegistryAndNFTModuleOwnerAgreement(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	binding, err := NewDomainNFTBinding(record.Name, record.NFTItemID, addr(1), 12)
	require.NoError(t, err)

	require.NoError(t, ValidateDomainNFTBinding(binding, DomainNFTBindingContext{
		RegistryOwner:	record.Owner,
		NFTModuleOwner:	addr(1),
		CurrentHeight:	12,
	}))
	require.ErrorContains(t, ValidateDomainNFTBinding(binding, DomainNFTBindingContext{
		RegistryOwner:	record.Owner,
		NFTModuleOwner:	addr(9),
		CurrentHeight:	12,
	}), "nft module owner")
}

func TestDomainNFTBindingTransferUpdatesRegistryAndBindingAtomically(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	binding, err := NewDomainNFTBinding(record.Name, record.NFTItemID, addr(1), 12)
	require.NoError(t, err)

	nextRecord, nextBinding, err := TransferDomainNFTBindingAtomic(record, binding, addr(9), 20)
	require.NoError(t, err)
	require.Equal(t, addr(9), nextRecord.Owner)
	require.Equal(t, addr(9), nextBinding.Owner)
	require.Equal(t, uint64(20), nextRecord.UpdatedAtHeight)
	require.Equal(t, uint64(20), nextBinding.LastVerifiedHeight)
}

func TestDomainNFTBindingBrokenBindingRestrictsUntilRepair(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	binding, err := NewDomainNFTBinding(record.Name, record.NFTItemID, addr(1), 12)
	require.NoError(t, err)
	binding.Owner = addr(9)

	restricted, err := RestrictDomainRecordV2ForBrokenBinding(record, binding, addr(9), 21)
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2GraceLocked, restricted.Status)
	require.NotZero(t, restricted.Flags&DomainRecordV2FlagRestricted)

	_, _, err = RepairDomainNFTBinding(restricted, binding, addr(1), addr(9), 22)
	require.ErrorContains(t, err, "cannot be repaired")

	restricted.Owner = addr(9)
	repaired, repairedBinding, err := RepairDomainNFTBinding(restricted, binding, addr(9), addr(9), 23)
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, repaired.Status)
	require.Zero(t, repaired.Flags&DomainRecordV2FlagRestricted)
	require.Equal(t, addr(9), repairedBinding.Owner)
	require.Equal(t, uint64(23), repairedBinding.LastVerifiedHeight)
}
