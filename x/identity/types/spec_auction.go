package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func StartSealedAuction(state IdentityState, name string, startHeight uint64) (IdentityState, Auction, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Auction{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, Auction{}, err
	}
	if startHeight == 0 {
		return IdentityState{}, Auction{}, errors.New("identity auction start height must be positive")
	}
	if !IsDomainAvailable(state, normalized, startHeight) {
		return IdentityState{}, Auction{}, errors.New("identity auction domain is not available")
	}
	if _, found := findAuction(state, normalized); found {
		return IdentityState{}, Auction{}, errors.New("identity auction already exists")
	}
	auction := Auction{
		Name:			normalized,
		CommitStartHeight:	startHeight,
		RevealStartHeight:	startHeight + state.Params.AuctionCommitBlocks,
		RevealEndHeight:	startHeight + state.Params.AuctionCommitBlocks + state.Params.AuctionRevealBlocks,
		Phase:			AuctionPhaseCommit,
	}
	next := state.Clone()
	next.Auctions = append(next.Auctions, auction)
	sortAuctions(next.Auctions)
	return next, auction, next.Validate()
}

func CommitAuctionBid(state IdentityState, name string, bidder sdk.AccAddress, commitmentHash string, height uint64) (IdentityState, Auction, error) {
	state = normalizeIdentityStateParams(state)
	auction, err := requireAuctionPhase(state, name, height, AuctionPhaseCommit)
	if err != nil {
		return IdentityState{}, Auction{}, err
	}
	if err := validateSpecAddress("identity auction bidder", bidder); err != nil {
		return IdentityState{}, Auction{}, err
	}
	if err := validateHexHash("identity auction commitment", commitmentHash); err != nil {
		return IdentityState{}, Auction{}, err
	}
	for _, commit := range auction.Commitments {
		if commit.CommitmentHash == commitmentHash {
			return IdentityState{}, Auction{}, errors.New("identity auction duplicate commitment")
		}
	}
	auction.Commitments = append(auction.Commitments, AuctionCommitment{Name: auction.Name, Bidder: cloneSpecAddress(bidder), CommitmentHash: commitmentHash, CommitHeight: height})
	sortAuctionCommitments(auction.Commitments)
	next := state.Clone()
	next.Auctions = upsertAuction(next.Auctions, auction)
	sortAuctions(next.Auctions)
	return next, auction, next.Validate()
}

func RevealAuctionBid(state IdentityState, name string, bidder sdk.AccAddress, bid uint64, salt string, height uint64) (IdentityState, Auction, error) {
	state = normalizeIdentityStateParams(state)
	auction, err := requireAuctionPhase(state, name, height, AuctionPhaseReveal)
	if err != nil {
		return IdentityState{}, Auction{}, err
	}
	commitment, err := ComputeAuctionCommitment(auction.Name, bidder, bid, salt)
	if err != nil {
		return IdentityState{}, Auction{}, err
	}
	if !auctionHasCommitment(auction, bidder, commitment) {
		return IdentityState{}, Auction{}, errors.New("identity auction commitment not found")
	}
	for _, reveal := range auction.Reveals {
		if reveal.CommitmentHash == commitment {
			return IdentityState{}, Auction{}, errors.New("identity auction commitment already revealed")
		}
	}
	auction.Reveals = append(auction.Reveals, AuctionReveal{Name: auction.Name, Bidder: cloneSpecAddress(bidder), Bid: bid, Salt: salt, RevealHeight: height, CommitmentHash: commitment})
	sortAuctionReveals(auction.Reveals)
	next := state.Clone()
	next.Auctions = upsertAuction(next.Auctions, auction)
	sortAuctions(next.Auctions)
	return next, auction, next.Validate()
}

func FinalizeSealedAuction(state IdentityState, name string, height uint64) (IdentityState, Auction, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Auction{}, err
	}
	auction, found := findAuction(state, name)
	if !found {
		return IdentityState{}, Auction{}, errors.New("identity auction not found")
	}
	if height < auction.RevealEndHeight {
		return IdentityState{}, Auction{}, errors.New("identity auction reveal phase is not over")
	}
	if auction.Phase == AuctionPhaseFinalized {
		return IdentityState{}, Auction{}, errors.New("identity auction already finalized")
	}
	if len(auction.Reveals) == 0 {
		return IdentityState{}, Auction{}, errors.New("identity auction has no revealed bids")
	}
	winner := chooseAuctionWinner(auction.Reveals)
	auction.Phase = AuctionPhaseFinalized
	auction.Winner = cloneSpecAddress(winner.Bidder)
	auction.WinningBid = winner.Bid
	auction.WinningCommitment = winner.CommitmentHash
	auction.Refunds = buildAuctionRefunds(auction, winner)
	next := state.Clone()
	next.Auctions = upsertAuction(next.Auctions, auction)
	sortAuctions(next.Auctions)
	return next, auction, next.Validate()
}

func requireAuctionPhase(state IdentityState, name string, height uint64, phase AuctionPhase) (Auction, error) {
	if err := state.Validate(); err != nil {
		return Auction{}, err
	}
	auction, found := findAuction(state, name)
	if !found {
		return Auction{}, errors.New("identity auction not found")
	}
	switch phase {
	case AuctionPhaseCommit:
		if height < auction.CommitStartHeight || height >= auction.RevealStartHeight {
			return Auction{}, errors.New("identity auction is not in commit phase")
		}
	case AuctionPhaseReveal:
		if height < auction.RevealStartHeight || height >= auction.RevealEndHeight {
			return Auction{}, errors.New("identity auction is not in reveal phase")
		}
	default:
		return Auction{}, fmt.Errorf("unsupported identity auction phase %q", phase)
	}
	if auction.Phase == AuctionPhaseFinalized {
		return Auction{}, errors.New("identity auction already finalized")
	}
	if auction.Phase != phase {
		auction.Phase = phase
	}
	return auction, nil
}

func chooseAuctionWinner(reveals []AuctionReveal) AuctionReveal {
	ordered := cloneAuctionReveals(reveals)
	sort.SliceStable(ordered, func(i, j int) bool {
		return compareAuctionWinner(ordered[i], ordered[j]) < 0
	})
	return ordered[0]
}

func compareAuctionWinner(left, right AuctionReveal) int {
	if left.Bid != right.Bid {
		if left.Bid > right.Bid {
			return -1
		}
		return 1
	}
	if left.RevealHeight != right.RevealHeight {
		if left.RevealHeight < right.RevealHeight {
			return -1
		}
		return 1
	}
	return stringsCompare(left.CommitmentHash, right.CommitmentHash)
}

func buildAuctionRefunds(auction Auction, winner AuctionReveal) []AuctionRefundReceipt {
	refunds := make([]AuctionRefundReceipt, 0, len(auction.Reveals)-1)
	for _, reveal := range auction.Reveals {
		if reveal.CommitmentHash == winner.CommitmentHash {
			continue
		}
		receipt := AuctionRefundReceipt{
			ReceiptID:	identityHash("auction-refund", auction.Name, reveal.CommitmentHash),
			Name:		auction.Name,
			Bidder:		cloneSpecAddress(reveal.Bidder),
			Amount:		reveal.Bid,
			CommitmentHash:	reveal.CommitmentHash,
			Reason:		"losing_bid",
		}
		refunds = append(refunds, receipt)
	}
	sortAuctionRefunds(refunds)
	return refunds
}

func validateAuctions(auctions []Auction) error {
	seen := make(map[string]struct{}, len(auctions))
	for i, auction := range auctions {
		if err := validateAuction(auction); err != nil {
			return err
		}
		if _, found := seen[auction.Name]; found {
			return errors.New("duplicate identity auction")
		}
		seen[auction.Name] = struct{}{}
		if i > 0 && auctions[i-1].Name >= auction.Name {
			return errors.New("identity auctions must be sorted canonically")
		}
	}
	return nil
}

func validateAuction(auction Auction) error {
	normalized, err := NormalizeAETDomain(auction.Name)
	if err != nil {
		return err
	}
	if auction.Name != normalized {
		return errors.New("identity auction name must be normalized")
	}
	if auction.CommitStartHeight == 0 || auction.RevealStartHeight <= auction.CommitStartHeight || auction.RevealEndHeight <= auction.RevealStartHeight {
		return errors.New("identity auction phase heights are invalid")
	}
	if auction.Phase != AuctionPhaseCommit && auction.Phase != AuctionPhaseReveal && auction.Phase != AuctionPhaseFinalized {
		return fmt.Errorf("invalid identity auction phase %q", auction.Phase)
	}
	if err := validateAuctionCommitments(auction.Commitments); err != nil {
		return err
	}
	if err := validateAuctionReveals(auction.Reveals); err != nil {
		return err
	}
	if err := validateAuctionRefunds(auction.Refunds); err != nil {
		return err
	}
	if auction.Phase == AuctionPhaseFinalized {
		if err := validateSpecAddress("identity auction winner", auction.Winner); err != nil {
			return err
		}
		if auction.WinningBid == 0 {
			return errors.New("identity auction winning bid must be positive")
		}
		if err := validateHexHash("identity auction winning commitment", auction.WinningCommitment); err != nil {
			return err
		}
	}
	return nil
}

func validateAuctionCommitments(commits []AuctionCommitment) error {
	seen := make(map[string]struct{}, len(commits))
	for i, commit := range commits {
		if err := validateSpecAddress("identity auction commit bidder", commit.Bidder); err != nil {
			return err
		}
		if err := validateHexHash("identity auction commit hash", commit.CommitmentHash); err != nil {
			return err
		}
		if _, found := seen[commit.CommitmentHash]; found {
			return errors.New("duplicate identity auction commitment")
		}
		seen[commit.CommitmentHash] = struct{}{}
		if i > 0 && compareAuctionCommitments(commits[i-1], commit) >= 0 {
			return errors.New("identity auction commitments must be sorted canonically")
		}
	}
	return nil
}

func validateAuctionReveals(reveals []AuctionReveal) error {
	seen := make(map[string]struct{}, len(reveals))
	for i, reveal := range reveals {
		if err := validateSpecAddress("identity auction reveal bidder", reveal.Bidder); err != nil {
			return err
		}
		if reveal.Bid == 0 {
			return errors.New("identity auction reveal bid must be positive")
		}
		if err := validateHexHash("identity auction reveal commitment", reveal.CommitmentHash); err != nil {
			return err
		}
		if _, found := seen[reveal.CommitmentHash]; found {
			return errors.New("duplicate identity auction reveal")
		}
		seen[reveal.CommitmentHash] = struct{}{}
		if i > 0 && compareAuctionReveals(reveals[i-1], reveal) >= 0 {
			return errors.New("identity auction reveals must be sorted canonically")
		}
	}
	return nil
}

func validateAuctionRefunds(refunds []AuctionRefundReceipt) error {
	seen := make(map[string]struct{}, len(refunds))
	for i, refund := range refunds {
		if err := validateSpecAddress("identity auction refund bidder", refund.Bidder); err != nil {
			return err
		}
		if refund.Amount == 0 {
			return errors.New("identity auction refund amount must be positive")
		}
		if err := validateHexHash("identity auction refund commitment", refund.CommitmentHash); err != nil {
			return err
		}
		if err := validateHexHash("identity auction refund receipt id", refund.ReceiptID); err != nil {
			return err
		}
		if _, found := seen[refund.ReceiptID]; found {
			return errors.New("duplicate identity auction refund")
		}
		seen[refund.ReceiptID] = struct{}{}
		if i > 0 && refunds[i-1].ReceiptID >= refund.ReceiptID {
			return errors.New("identity auction refunds must be sorted canonically")
		}
	}
	return nil
}

func findAuction(state IdentityState, name string) (Auction, bool) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return Auction{}, false
	}
	for _, auction := range state.Auctions {
		if auction.Name == normalized {
			return cloneAuction(auction), true
		}
	}
	return Auction{}, false
}

func auctionHasCommitment(auction Auction, bidder sdk.AccAddress, commitment string) bool {
	for _, commit := range auction.Commitments {
		if bytes.Equal(commit.Bidder, bidder) && commit.CommitmentHash == commitment {
			return true
		}
	}
	return false
}

func upsertAuction(auctions []Auction, auction Auction) []Auction {
	out := cloneAuctions(auctions)
	for i := range out {
		if out[i].Name == auction.Name {
			out[i] = cloneAuction(auction)
			return out
		}
	}
	return append(out, cloneAuction(auction))
}

func sortAuctions(auctions []Auction) {
	sort.SliceStable(auctions, func(i, j int) bool { return auctions[i].Name < auctions[j].Name })
}

func sortAuctionCommitments(commits []AuctionCommitment) {
	sort.SliceStable(commits, func(i, j int) bool { return compareAuctionCommitments(commits[i], commits[j]) < 0 })
}

func sortAuctionReveals(reveals []AuctionReveal) {
	sort.SliceStable(reveals, func(i, j int) bool { return compareAuctionReveals(reveals[i], reveals[j]) < 0 })
}

func sortAuctionRefunds(refunds []AuctionRefundReceipt) {
	sort.SliceStable(refunds, func(i, j int) bool { return refunds[i].ReceiptID < refunds[j].ReceiptID })
}

func compareAuctionCommitments(left, right AuctionCommitment) int {
	if left.Name != right.Name {
		return stringsCompare(left.Name, right.Name)
	}
	return stringsCompare(left.CommitmentHash, right.CommitmentHash)
}

func compareAuctionReveals(left, right AuctionReveal) int {
	return stringsCompare(left.CommitmentHash, right.CommitmentHash)
}
