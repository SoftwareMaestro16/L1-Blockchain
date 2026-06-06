package types

import (
	"bytes"
	"errors"
	"fmt"
)

func (s IdentityState) Validate() error {
	params := normalizeIdentityParams(s.Params)
	if params.RenewalWindowBlocks > params.RegistrationPeriodBlocks {
		return errors.New("identity renewal window must not exceed registration period")
	}
	if err := validateDomains(s.Domains); err != nil {
		return err
	}
	if err := validateDomainNFTs(s.DomainNFTs, s.Domains); err != nil {
		return err
	}
	if err := validateDomainCommits(s.Commits); err != nil {
		return err
	}
	if err := validateResolvers(s.Resolvers, s.Domains); err != nil {
		return err
	}
	if err := validateReverseRecords(s.ReverseRecords); err != nil {
		return err
	}
	if err := validateSubdomains(s.Subdomains, s.Domains); err != nil {
		return err
	}
	if err := validateAuctions(s.Auctions); err != nil {
		return err
	}
	return validateResolverIntents(s.PendingResolverUpdates)
}

func validateDomains(domains []Domain) error {
	seen := make(map[string]struct{}, len(domains))
	for i, domain := range domains {
		if err := validateDomain(domain); err != nil {
			return err
		}
		if _, found := seen[domain.Name]; found {
			return errors.New("duplicate identity domain")
		}
		seen[domain.Name] = struct{}{}
		if i > 0 && domains[i-1].Name >= domain.Name {
			return errors.New("identity domains must be sorted canonically")
		}
	}
	return nil
}

func validateDomain(domain Domain) error {
	normalized, err := NormalizeAETDomain(domain.Name)
	if err != nil {
		return err
	}
	if domain.Name != normalized {
		return errors.New("identity domain name must be normalized")
	}
	if err := validateSpecAddress("identity domain owner", domain.Owner); err != nil {
		return err
	}
	if domain.NFTID == "" {
		return errors.New("identity domain NFT id is required")
	}
	expectedNFT, err := DomainNFTID(domain.Name)
	if err != nil {
		return err
	}
	if domain.NFTID != expectedNFT {
		return errors.New("identity domain NFT id mismatch")
	}
	if domain.RegisteredHeight == 0 || domain.ExpiryHeight <= domain.RegisteredHeight {
		return errors.New("identity domain expiry must be after registration")
	}
	if domain.UpdatedHeight < domain.RegisteredHeight {
		return errors.New("identity domain updated height must not precede registration")
	}
	if domain.ParentName != "" {
		if _, err := NormalizeAETDomain(domain.ParentName); err != nil {
			return err
		}
	}
	return nil
}

func validateDomainNFTs(nfts []DomainNFT, domains []Domain) error {
	seen := make(map[string]struct{}, len(nfts))
	for i, nft := range nfts {
		if err := validateDomainNFT(nft); err != nil {
			return err
		}
		if _, found := seen[nft.ID]; found {
			return errors.New("duplicate identity domain NFT")
		}
		seen[nft.ID] = struct{}{}
		if i > 0 && nfts[i-1].ID >= nft.ID {
			return errors.New("identity domain NFTs must be sorted canonically")
		}
		domain, found := findDomain(IdentityState{Domains: domains}, nft.Domain)
		if !found || domain.NFTID != nft.ID || !bytes.Equal(domain.Owner, nft.Owner) {
			return errors.New("identity domain NFT ownership mismatch")
		}
	}
	return nil
}

func validateDomainNFT(nft DomainNFT) error {
	if nft.ID == "" {
		return errors.New("identity domain NFT id is required")
	}
	expected, err := DomainNFTID(nft.Domain)
	if err != nil {
		return err
	}
	if nft.ID != expected {
		return errors.New("identity domain NFT id mismatch")
	}
	if err := validateSpecAddress("identity domain NFT owner", nft.Owner); err != nil {
		return err
	}
	if nft.MintHeight == 0 {
		return errors.New("identity domain NFT mint height must be positive")
	}
	if nft.TransferHeight != 0 && nft.TransferHeight < nft.MintHeight {
		return errors.New("identity domain NFT transfer height must not precede mint")
	}
	return nil
}

func validateDomainCommits(commits []DomainCommit) error {
	seen := make(map[string]struct{}, len(commits))
	for i, commit := range commits {
		if err := validateDomainCommit(commit); err != nil {
			return err
		}
		key := commit.Name + "/" + commit.CommitmentHash
		if _, found := seen[key]; found {
			return errors.New("duplicate identity domain commit")
		}
		seen[key] = struct{}{}
		if i > 0 && compareDomainCommits(commits[i-1], commit) >= 0 {
			return errors.New("identity domain commits must be sorted canonically")
		}
	}
	return nil
}

func validateDomainCommit(commit DomainCommit) error {
	if normalized, err := NormalizeAETDomain(commit.Name); err != nil {
		return err
	} else if commit.Name != normalized {
		return errors.New("identity commit domain name must be normalized")
	}
	if err := validateSpecAddress("identity commit owner", commit.Owner); err != nil {
		return err
	}
	if err := validateHexHash("identity commit hash", commit.CommitmentHash); err != nil {
		return err
	}
	if commit.CommitHeight == 0 || commit.ExpiresHeight <= commit.CommitHeight {
		return errors.New("identity commit expiry must be after commit height")
	}
	return nil
}

func validateResolvers(records []ResolverRecord, domains []Domain) error {
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := ValidateResolverRecord(record); err != nil {
			return err
		}
		authority, found := findResolverAuthorityDomain(IdentityState{Domains: domains}, record.Domain)
		if !found {
			return errors.New("identity resolver authority domain not found")
		}
		if !bytes.Equal(record.Owner, authority.Owner) {
			return errors.New("identity resolver owner must match registry owner")
		}
		if _, found := seen[record.Domain]; found {
			return errors.New("duplicate identity resolver")
		}
		seen[record.Domain] = struct{}{}
		if i > 0 && records[i-1].Domain >= record.Domain {
			return errors.New("identity resolvers must be sorted canonically")
		}
	}
	return nil
}

func validateReverseRecords(records []ReverseRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := validateSpecAddress("identity reverse address", record.Address); err != nil {
			return err
		}
		if normalized, err := NormalizeAETDomain(record.Domain); err != nil {
			return err
		} else if record.Domain != normalized {
			return errors.New("identity reverse domain must be normalized")
		}
		key := string(record.Address)
		if _, found := seen[key]; found {
			return errors.New("duplicate identity reverse record")
		}
		seen[key] = struct{}{}
		if i > 0 && compareReverse(records[i-1], record) >= 0 {
			return errors.New("identity reverse records must be sorted canonically")
		}
	}
	return nil
}

func validateSubdomains(records []SubdomainRecord, domains []Domain) error {
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		if err := validateSubdomain(record, domains); err != nil {
			return err
		}
		if _, found := seen[record.Name]; found {
			return errors.New("duplicate identity subdomain")
		}
		seen[record.Name] = struct{}{}
		if i > 0 && records[i-1].Name >= record.Name {
			return errors.New("identity subdomains must be sorted canonically")
		}
	}
	return nil
}

func validateSubdomain(record SubdomainRecord, domains []Domain) error {
	parent, err := NormalizeAETDomain(record.ParentName)
	if err != nil {
		return err
	}
	if record.ParentName != parent {
		return errors.New("identity subdomain parent must be normalized")
	}
	child, err := NormalizeAETDomain(record.Name)
	if err != nil {
		return err
	}
	if record.Name != child {
		return errors.New("identity subdomain name must be normalized")
	}
	if !stringsHasSuffixLabel(child, parent) {
		return errors.New("identity subdomain must be below parent")
	}
	if err := validateSpecAddress("identity subdomain owner", record.Owner); err != nil {
		return err
	}
	if record.CreatedHeight == 0 {
		return errors.New("identity subdomain created height must be positive")
	}
	domain, found := findDomain(IdentityState{Domains: domains}, child)
	if !found || !bytes.Equal(domain.Owner, record.Owner) {
		return errors.New("identity subdomain registry mismatch")
	}
	return nil
}

func validateResolverIntents(intents []ResolverUpdateIntent) error {
	seen := make(map[string]struct{}, len(intents))
	for i, intent := range intents {
		if normalized, err := NormalizeAETDomain(intent.Domain); err != nil {
			return err
		} else if intent.Domain != normalized {
			return errors.New("identity pending resolver domain must be normalized")
		}
		if err := validateSpecAddress("identity pending resolver actor", intent.Actor); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s/%020d", intent.Domain, string(intent.Actor), intent.Nonce)
		if _, found := seen[key]; found {
			return errors.New("duplicate identity pending resolver update")
		}
		seen[key] = struct{}{}
		if i > 0 && compareResolverIntents(intents[i-1], intent) >= 0 {
			return errors.New("identity pending resolver updates must be sorted canonically")
		}
	}
	return nil
}

func validateHexHash(field, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be 64 lowercase hex chars", field)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be 64 lowercase hex chars", field)
	}
	return nil
}

func stringsHasSuffixLabel(child, parent string) bool {
	return len(child) > len(parent) && child[len(child)-len(parent):] == parent && child[len(child)-len(parent)-1] == '.'
}
