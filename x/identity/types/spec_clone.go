package types

func cloneDomain(domain Domain) Domain {
	domain.Owner = cloneSpecAddress(domain.Owner)
	return domain
}

func cloneDomains(domains []Domain) []Domain {
	out := make([]Domain, len(domains))
	for i, domain := range domains {
		out[i] = cloneDomain(domain)
	}
	return out
}

func cloneDomainNFT(nft DomainNFT) DomainNFT {
	nft.Owner = cloneSpecAddress(nft.Owner)
	return nft
}

func cloneDomainNFTs(nfts []DomainNFT) []DomainNFT {
	out := make([]DomainNFT, len(nfts))
	for i, nft := range nfts {
		out[i] = cloneDomainNFT(nft)
	}
	return out
}

func cloneDomainCommit(commit DomainCommit) DomainCommit {
	commit.Owner = cloneSpecAddress(commit.Owner)
	return commit
}

func cloneDomainCommits(commits []DomainCommit) []DomainCommit {
	out := make([]DomainCommit, len(commits))
	for i, commit := range commits {
		out[i] = cloneDomainCommit(commit)
	}
	return out
}

func cloneResolver(record ResolverRecord) ResolverRecord {
	record.Owner = cloneSpecAddress(record.Owner)
	record.Primary = cloneSpecAddress(record.Primary)
	record.Contract = cloneSpecAddress(record.Contract)
	record.Records = cloneResolverRecords(record.Records)
	record.Metadata = append([]byte(nil), record.Metadata...)
	return record
}

func cloneResolvers(records []ResolverRecord) []ResolverRecord {
	out := make([]ResolverRecord, len(records))
	for i, record := range records {
		out[i] = cloneResolver(record)
	}
	return out
}

func cloneReverseRecord(record ReverseRecord) ReverseRecord {
	record.Address = cloneSpecAddress(record.Address)
	return record
}

func cloneReverseRecords(records []ReverseRecord) []ReverseRecord {
	out := make([]ReverseRecord, len(records))
	for i, record := range records {
		out[i] = cloneReverseRecord(record)
	}
	return out
}

func cloneSubdomain(record SubdomainRecord) SubdomainRecord {
	record.Owner = cloneSpecAddress(record.Owner)
	return record
}

func cloneSubdomains(records []SubdomainRecord) []SubdomainRecord {
	out := make([]SubdomainRecord, len(records))
	for i, record := range records {
		out[i] = cloneSubdomain(record)
	}
	return out
}

func cloneResolverIntent(intent ResolverUpdateIntent) ResolverUpdateIntent {
	intent.Actor = cloneSpecAddress(intent.Actor)
	return intent
}

func cloneResolverIntents(intents []ResolverUpdateIntent) []ResolverUpdateIntent {
	out := make([]ResolverUpdateIntent, len(intents))
	for i, intent := range intents {
		out[i] = cloneResolverIntent(intent)
	}
	return out
}

func cloneAuction(auction Auction) Auction {
	auction.Commitments = cloneAuctionCommitments(auction.Commitments)
	auction.Reveals = cloneAuctionReveals(auction.Reveals)
	auction.Winner = cloneSpecAddress(auction.Winner)
	auction.Refunds = cloneAuctionRefunds(auction.Refunds)
	return auction
}

func cloneAuctions(auctions []Auction) []Auction {
	out := make([]Auction, len(auctions))
	for i, auction := range auctions {
		out[i] = cloneAuction(auction)
	}
	return out
}

func cloneAuctionCommitment(commitment AuctionCommitment) AuctionCommitment {
	commitment.Bidder = cloneSpecAddress(commitment.Bidder)
	return commitment
}

func cloneAuctionCommitments(commitments []AuctionCommitment) []AuctionCommitment {
	out := make([]AuctionCommitment, len(commitments))
	for i, commitment := range commitments {
		out[i] = cloneAuctionCommitment(commitment)
	}
	return out
}

func cloneAuctionReveal(reveal AuctionReveal) AuctionReveal {
	reveal.Bidder = cloneSpecAddress(reveal.Bidder)
	return reveal
}

func cloneAuctionReveals(reveals []AuctionReveal) []AuctionReveal {
	out := make([]AuctionReveal, len(reveals))
	for i, reveal := range reveals {
		out[i] = cloneAuctionReveal(reveal)
	}
	return out
}

func cloneAuctionRefund(refund AuctionRefundReceipt) AuctionRefundReceipt {
	refund.Bidder = cloneSpecAddress(refund.Bidder)
	return refund
}

func cloneAuctionRefunds(refunds []AuctionRefundReceipt) []AuctionRefundReceipt {
	out := make([]AuctionRefundReceipt, len(refunds))
	for i, refund := range refunds {
		out[i] = cloneAuctionRefund(refund)
	}
	return out
}
