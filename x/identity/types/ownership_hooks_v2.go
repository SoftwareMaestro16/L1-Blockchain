package types

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityNFTTransferHookModeV2 string

const (
	IdentityNFTTransferHookRejectOnRegistryMismatchV2	IdentityNFTTransferHookModeV2	= "reject_on_registry_mismatch"
	IdentityNFTTransferHookUpdateRegistryV2			IdentityNFTTransferHookModeV2	= "update_registry"
)

func ApplyIdentityNFTTransferHookStateV2(state IdentityState, nftID string, newOwner sdk.AccAddress, height uint64, mode IdentityNFTTransferHookModeV2) (IdentityState, Domain, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	if err := validateSpecAddress("identity nft transfer hook new owner", newOwner); err != nil {
		return IdentityState{}, Domain{}, err
	}
	if nftID == "" {
		return IdentityState{}, Domain{}, errors.New("identity nft transfer hook nft id is required")
	}
	nft, found := findDomainNFTByID(state, nftID)
	if !found {
		return IdentityState{}, Domain{}, errors.New("identity nft transfer hook nft binding not found")
	}
	domain, found := findDomain(state, nft.Domain)
	if !found {
		return IdentityState{}, Domain{}, errors.New("identity nft transfer hook registry domain not found")
	}
	switch mode {
	case IdentityNFTTransferHookRejectOnRegistryMismatchV2:
		if !bytes.Equal(domain.Owner, newOwner) {
			return IdentityState{}, Domain{}, errors.New("identity nft transfer hook rejected because registry owner would diverge")
		}
		return state.Export(), domain, nil
	case IdentityNFTTransferHookUpdateRegistryV2:
		next := state.Clone()
		domain.Owner = cloneSpecAddress(newOwner)
		domain.UpdatedHeight = height
		next.Domains = upsertDomain(next.Domains, domain)
		next.DomainNFTs = transferNFT(next.DomainNFTs, nftID, newOwner, height)
		next.Resolvers = transferResolverOwnership(next.Resolvers, state.Domains, domain.Name, newOwner, height)
		next.PendingResolverUpdates = removePendingResolverUpdates(next.PendingResolverUpdates, state.Domains, domain.Name)
		var err error
		next, _, err = InvalidateReverseRecordsForDomainV2(next, domain.Name, height, nil)
		if err != nil {
			return IdentityState{}, Domain{}, err
		}
		sortIdentityState(&next)
		return next, domain, next.Validate()
	default:
		return IdentityState{}, Domain{}, errors.New("identity nft transfer hook mode is unsupported")
	}
}

func ApplyIdentityRegistryTransferHookStateV2(state IdentityState, name string, actor sdk.AccAddress, newOwner sdk.AccAddress, height uint64) (IdentityState, Domain, error) {
	return TransferDomainNFT(state, name, actor, newOwner, height)
}

func ApplyIdentityNFTTransferHookV2(record DomainRecordV2, binding DomainNFTBinding, nftModuleNewOwner sdk.AccAddress, height uint64, mode IdentityNFTTransferHookModeV2) (DomainRecordV2, DomainNFTBinding, error) {
	switch mode {
	case IdentityNFTTransferHookRejectOnRegistryMismatchV2:
		if !addressesEqual(record.Owner, nftModuleNewOwner) {
			return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft transfer hook rejected because registry owner would diverge")
		}
		return record, binding, nil
	case IdentityNFTTransferHookUpdateRegistryV2:
		return TransferDomainNFTBindingAtomic(record, binding, nftModuleNewOwner, height)
	default:
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft transfer hook mode is unsupported")
	}
}

func ApplyIdentityRegistryTransferHookV2(record DomainRecordV2, binding DomainNFTBinding, actor sdk.AccAddress, newOwner sdk.AccAddress, height uint64) (DomainRecordV2, DomainNFTBinding, error) {
	return TransferDomainNFTBindingWithInvariantsV2(record, binding, actor, newOwner, height)
}
