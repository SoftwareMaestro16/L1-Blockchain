package types

import (
	"bytes"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func findDomain(state IdentityState, name string) (Domain, bool) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return Domain{}, false
	}
	for _, domain := range state.Domains {
		if domain.Name == normalized {
			return cloneDomain(domain), true
		}
	}
	return Domain{}, false
}

func findResolver(state IdentityState, name string) (ResolverRecord, bool) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return ResolverRecord{}, false
	}
	for _, resolver := range state.Resolvers {
		if resolver.Domain == normalized {
			return cloneResolver(resolver), true
		}
	}
	return ResolverRecord{}, false
}

func findActiveCommit(state IdentityState, name string, height uint64) (DomainCommit, bool) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return DomainCommit{}, false
	}
	for _, commit := range state.Commits {
		if commit.Name == normalized && commit.ExpiresHeight >= height {
			return cloneDomainCommit(commit), true
		}
	}
	return DomainCommit{}, false
}

func findActiveCommitByHash(state IdentityState, name string, owner sdk.AccAddress, commitment string, height uint64) (DomainCommit, bool) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return DomainCommit{}, false
	}
	for _, commit := range state.Commits {
		if commit.Name == normalized && bytes.Equal(commit.Owner, owner) && commit.CommitmentHash == commitment && commit.ExpiresHeight >= height {
			return cloneDomainCommit(commit), true
		}
	}
	return DomainCommit{}, false
}

func upsertDomain(domains []Domain, domain Domain) []Domain {
	out := cloneDomains(domains)
	for i := range out {
		if out[i].Name == domain.Name {
			out[i] = cloneDomain(domain)
			return out
		}
	}
	return append(out, cloneDomain(domain))
}

func upsertDomainNFT(nfts []DomainNFT, nft DomainNFT) []DomainNFT {
	out := cloneDomainNFTs(nfts)
	for i := range out {
		if out[i].ID == nft.ID {
			out[i] = cloneDomainNFT(nft)
			return out
		}
	}
	return append(out, cloneDomainNFT(nft))
}

func upsertResolver(records []ResolverRecord, record ResolverRecord) []ResolverRecord {
	out := cloneResolvers(records)
	for i := range out {
		if out[i].Domain == record.Domain {
			out[i] = cloneResolver(record)
			return out
		}
	}
	return append(out, cloneResolver(record))
}

func upsertReverse(records []ReverseRecord, record ReverseRecord) []ReverseRecord {
	out := cloneReverseRecords(records)
	for i := range out {
		if bytes.Equal(out[i].Address, record.Address) {
			out[i] = cloneReverseRecord(record)
			return out
		}
	}
	return append(out, cloneReverseRecord(record))
}

func removeDomainCommit(commits []DomainCommit, target DomainCommit) []DomainCommit {
	out := make([]DomainCommit, 0, len(commits))
	for _, commit := range commits {
		if commit.Name == target.Name && commit.CommitmentHash == target.CommitmentHash && bytes.Equal(commit.Owner, target.Owner) {
			continue
		}
		out = append(out, cloneDomainCommit(commit))
	}
	return out
}

func transferNFT(nfts []DomainNFT, id string, newOwner sdk.AccAddress, height uint64) []DomainNFT {
	out := cloneDomainNFTs(nfts)
	for i := range out {
		if out[i].ID == id {
			out[i].Owner = cloneSpecAddress(newOwner)
			out[i].TransferHeight = height
		}
	}
	return out
}

func transferResolverOwnership(records []ResolverRecord, domain string, newOwner sdk.AccAddress, height uint64) []ResolverRecord {
	out := cloneResolvers(records)
	for i := range out {
		if out[i].Domain == domain {
			out[i].Owner = cloneSpecAddress(newOwner)
			out[i].UpdatedAtUnix = int64(height)
		}
	}
	return out
}

func removePendingResolverUpdates(intents []ResolverUpdateIntent, domain string) []ResolverUpdateIntent {
	out := make([]ResolverUpdateIntent, 0, len(intents))
	for _, intent := range intents {
		if intent.Domain == domain {
			continue
		}
		out = append(out, cloneResolverIntent(intent))
	}
	return out
}

func sortDomains(domains []Domain) {
	sort.SliceStable(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
}

func sortDomainNFTs(nfts []DomainNFT) {
	sort.SliceStable(nfts, func(i, j int) bool { return nfts[i].ID < nfts[j].ID })
}

func sortDomainCommits(commits []DomainCommit) {
	sort.SliceStable(commits, func(i, j int) bool { return compareDomainCommits(commits[i], commits[j]) < 0 })
}

func sortResolvers(records []ResolverRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Domain < records[j].Domain })
}

func sortReverseRecords(records []ReverseRecord) {
	sort.SliceStable(records, func(i, j int) bool { return compareReverse(records[i], records[j]) < 0 })
}

func sortSubdomains(records []SubdomainRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Name < records[j].Name })
}

func sortResolverIntents(intents []ResolverUpdateIntent) {
	sort.SliceStable(intents, func(i, j int) bool { return compareResolverIntents(intents[i], intents[j]) < 0 })
}

func compareDomainCommits(left, right DomainCommit) int {
	if left.Name != right.Name {
		return stringsCompare(left.Name, right.Name)
	}
	if c := compareAddress(left.Owner, right.Owner); c != 0 {
		return c
	}
	return stringsCompare(left.CommitmentHash, right.CommitmentHash)
}

func compareReverse(left, right ReverseRecord) int {
	return compareAddress(left.Address, right.Address)
}

func compareResolverIntents(left, right ResolverUpdateIntent) int {
	if left.Domain != right.Domain {
		return stringsCompare(left.Domain, right.Domain)
	}
	if c := compareAddress(left.Actor, right.Actor); c != 0 {
		return c
	}
	if left.Nonce < right.Nonce {
		return -1
	}
	if left.Nonce > right.Nonce {
		return 1
	}
	return 0
}

func stringsCompare(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
