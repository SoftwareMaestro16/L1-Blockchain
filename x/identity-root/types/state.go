package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	SubdomainPolicyOwnerOnly	= "owner_only"
	SubdomainPolicyPublic		= "public"
	SubdomainPolicyDisabled		= "disabled"

	DefaultResolverRoot	= "0000000000000000000000000000000000000000000000000000000000000000"
	DefaultRootNamespace	= "aet"

	DomainRentPayerOwner	= "owner"
	DomainRentPayerProtocol	= "protocol"
)

type IdentityRootParams struct {
	RootNamespace			string
	RegistrationPeriod		uint64
	RenewalPeriod			uint64
	MaxNameBytes			uint32
	MaxRecords			uint32
	MaxReservedNames		uint32
	DomainRentRatePerByteBlock	uint64
	DefaultDomainRentPayerPolicy	string
	NFTBindingEnabled		bool
	AllowPublicSubdomains		bool
	Auction				AuctionParams
}

type AuctionParams struct {
	Enabled			bool
	CommitBlocks		uint64
	RevealBlocks		uint64
	MinimumBidAmount	uint64
}

type IdentityRootState struct {
	Records		[]NameRecord
	Resolvers	[]ResolverRecord
	ReverseRecords	[]ReverseRecord
	NFTBindings	[]IdentityNFTBindingReference
	RootAuthorities	[]RootAuthority
	ReservedNames	[]ReservedName
}

type NameRecord struct {
	Name			string
	ParentName		string
	Owner			string
	ResolverRoot		string
	ExpiryHeight		uint64
	RenewalHeight		uint64
	SubdomainPolicy		string
	NFTBinding		IdentityNFTBindingReference
	StorageRentDebt		uint64
	LastStorageChargeHeight	uint64
	RentPayerPolicy		string
	CreatedHeight		uint64
	UpdatedHeight		uint64
}

type ResolverRecord struct {
	Name		string
	ResolverRoot	string
	UpdatedHeight	uint64
}

type ReverseRecord struct {
	Address		string
	Name		string
	Owner		string
	UpdatedHeight	uint64
}

type IdentityNFTBindingReference struct {
	Name	string
	Enabled	bool
	ClassID	string
	NFTID	string
	Owner	string
}

type RootAuthority struct {
	Authority	string
	Role		string
}

type ReservedName struct {
	Name		string
	Authority	string
	Reason		string
}

type MsgRegisterName struct {
	Owner		string
	Name		string
	Height		uint64
	ResolverRoot	string
	SubdomainPolicy	string
	NFTBinding	IdentityNFTBindingReference
}

type MsgRenewName struct {
	Owner	string
	Name	string
	Height	uint64
}

type MsgTransferName struct {
	Owner		string
	Name		string
	NewOwner	string
	Height		uint64
	NewNFTBinding	IdentityNFTBindingReference
}

type MsgSetResolver struct {
	Owner		string
	Name		string
	ResolverRoot	string
	Height		uint64
}

type MsgSetReverseRecord struct {
	Owner	string
	Address	string
	Name	string
	Height	uint64
}

type MsgCreateSubdomain struct {
	Owner		string
	ParentName	string
	Label		string
	SubdomainOwner	string
	Height		uint64
	ResolverRoot	string
	SubdomainPolicy	string
	NFTBinding	IdentityNFTBindingReference
}

type MsgReserveName struct {
	Authority	string
	Name		string
	Reason		string
}

type MsgReleaseReservedName struct {
	Authority	string
	Name		string
}

func DefaultIdentityRootParams() IdentityRootParams {
	return IdentityRootParams{
		RootNamespace:			DefaultRootNamespace,
		RegistrationPeriod:		1_000_000,
		RenewalPeriod:			1_000_000,
		MaxNameBytes:			253,
		MaxRecords:			100_000,
		MaxReservedNames:		10_000,
		DomainRentRatePerByteBlock:	1,
		DefaultDomainRentPayerPolicy:	DomainRentPayerOwner,
		Auction: AuctionParams{
			CommitBlocks:		100,
			RevealBlocks:		100,
			MinimumBidAmount:	1,
		},
	}
}

func EmptyIdentityRootState() IdentityRootState {
	return IdentityRootState{
		Records:		[]NameRecord{},
		Resolvers:		[]ResolverRecord{},
		ReverseRecords:		[]ReverseRecord{},
		NFTBindings:		[]IdentityNFTBindingReference{},
		RootAuthorities:	[]RootAuthority{},
		ReservedNames:		[]ReservedName{},
	}
}

func (p IdentityRootParams) Validate() error {
	root, err := NormalizeRootNamespace(p.RootNamespace)
	if err != nil {
		return err
	}
	if root == "" {
		return errors.New("identity root namespace is required")
	}
	if p.RegistrationPeriod == 0 || p.RenewalPeriod == 0 {
		return errors.New("identity root registration and renewal periods must be positive")
	}
	if p.MaxNameBytes == 0 || p.MaxRecords == 0 || p.MaxReservedNames == 0 {
		return errors.New("identity root limits must be positive")
	}
	if p.DomainRentRatePerByteBlock == 0 {
		return errors.New("identity domain storage rent rate must be positive")
	}
	if !IsDomainRentPayerPolicy(p.DefaultDomainRentPayerPolicy) {
		return errors.New("identity domain rent payer policy is invalid")
	}
	if p.Auction.Enabled && (p.Auction.CommitBlocks == 0 || p.Auction.RevealBlocks == 0 || p.Auction.MinimumBidAmount == 0) {
		return errors.New("identity root auction parameters must be positive when enabled")
	}
	return nil
}

func (s IdentityRootState) Export() IdentityRootState {
	out := IdentityRootState{
		Records:		cloneRecords(s.Records),
		Resolvers:		cloneResolvers(s.Resolvers),
		ReverseRecords:		cloneReverseRecords(s.ReverseRecords),
		NFTBindings:		cloneBindings(s.NFTBindings),
		RootAuthorities:	cloneAuthorities(s.RootAuthorities),
		ReservedNames:		cloneReserved(s.ReservedNames),
	}
	SortRecords(out.Records)
	SortResolvers(out.Resolvers)
	SortReverseRecords(out.ReverseRecords)
	SortBindings(out.NFTBindings)
	SortAuthorities(out.RootAuthorities)
	SortReserved(out.ReservedNames)
	return out
}

func (s IdentityRootState) Validate(params IdentityRootParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Records)) > params.MaxRecords {
		return errors.New("identity root record count exceeds limit")
	}
	if uint32(len(s.ReservedNames)) > params.MaxReservedNames {
		return errors.New("identity root reserved name count exceeds limit")
	}
	records := map[string]NameRecord{}
	for _, record := range s.Records {
		if err := record.Validate(params); err != nil {
			return err
		}
		record = record.Normalize(params)
		if _, found := records[record.Name]; found {
			return fmt.Errorf("duplicate identity name %q", record.Name)
		}
		records[record.Name] = record
	}
	authorities := map[string]struct{}{}
	for _, authority := range s.RootAuthorities {
		if err := authority.Validate(); err != nil {
			return err
		}
		if _, found := authorities[authority.Authority]; found {
			return fmt.Errorf("duplicate identity root authority %q", authority.Authority)
		}
		authorities[authority.Authority] = struct{}{}
	}
	reserved := map[string]struct{}{}
	for _, name := range s.ReservedNames {
		if err := name.Validate(params); err != nil {
			return err
		}
		name = name.Normalize(params)
		if _, found := reserved[name.Name]; found {
			return fmt.Errorf("duplicate reserved identity name %q", name.Name)
		}
		reserved[name.Name] = struct{}{}
	}
	for _, record := range s.Records {
		record = record.Normalize(params)
		if _, isReserved := reserved[record.Name]; isReserved {
			if _, isAuthority := authorities[record.Owner]; !isAuthority {
				return fmt.Errorf("reserved identity name %q cannot be owned by normal user", record.Name)
			}
		}
		if record.ParentName != "" {
			parent, found := records[record.ParentName]
			if !found {
				return fmt.Errorf("identity subdomain %q references missing parent", record.Name)
			}
			if parent.SubdomainPolicy == SubdomainPolicyOwnerOnly && record.Owner != parent.Owner {
				return fmt.Errorf("identity subdomain %q must follow parent ownership policy", record.Name)
			}
			if parent.SubdomainPolicy == SubdomainPolicyDisabled {
				return fmt.Errorf("identity subdomain %q is disabled by parent policy", record.Name)
			}
		}
	}
	for _, resolver := range s.Resolvers {
		if err := resolver.Validate(params); err != nil {
			return err
		}
		resolver = resolver.Normalize(params)
		if _, found := records[resolver.Name]; !found {
			return fmt.Errorf("identity resolver references unknown name %q", resolver.Name)
		}
	}
	for _, reverse := range s.ReverseRecords {
		if err := reverse.Validate(params); err != nil {
			return err
		}
		reverse = reverse.Normalize(params)
		record, found := records[reverse.Name]
		if !found {
			return fmt.Errorf("identity reverse record references unknown name %q", reverse.Name)
		}
		if reverse.Owner != record.Owner {
			return fmt.Errorf("identity reverse owner must match name owner for %q", reverse.Name)
		}
	}
	for _, binding := range s.NFTBindings {
		if err := binding.Validate(params); err != nil {
			return err
		}
		binding = binding.Normalize(params)
		record, found := records[binding.Name]
		if !found {
			return fmt.Errorf("identity NFT binding references unknown name %q", binding.Name)
		}
		if params.NFTBindingEnabled && binding.Owner != record.Owner {
			return fmt.Errorf("identity NFT binding owner must match name owner for %q", binding.Name)
		}
	}
	return nil
}

func (r NameRecord) Normalize(params IdentityRootParams) NameRecord {
	r.Name, _ = NormalizeName(r.Name, params.RootNamespace)
	r.ParentName, _ = NormalizeOptionalName(r.ParentName, params.RootNamespace)
	r.Owner = strings.TrimSpace(r.Owner)
	r.ResolverRoot = normalizeResolverRoot(r.ResolverRoot)
	r.SubdomainPolicy = strings.TrimSpace(r.SubdomainPolicy)
	if r.SubdomainPolicy == "" {
		r.SubdomainPolicy = SubdomainPolicyOwnerOnly
	}
	r.RentPayerPolicy = strings.TrimSpace(r.RentPayerPolicy)
	if r.RentPayerPolicy == "" {
		r.RentPayerPolicy = params.DefaultDomainRentPayerPolicy
	}
	if r.LastStorageChargeHeight == 0 && r.CreatedHeight != 0 {
		r.LastStorageChargeHeight = r.CreatedHeight
	}
	r.NFTBinding = r.NFTBinding.Normalize(params)
	return r
}

func (r NameRecord) Validate(params IdentityRootParams) error {
	r = r.Normalize(params)
	if err := ValidateName(r.Name, params); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("identity name owner", r.Owner); err != nil {
		return err
	}
	if err := ValidateResolverRoot(r.ResolverRoot); err != nil {
		return err
	}
	if r.ExpiryHeight == 0 || r.CreatedHeight == 0 || r.UpdatedHeight == 0 {
		return errors.New("identity name heights must be positive")
	}
	if r.UpdatedHeight < r.CreatedHeight {
		return errors.New("identity name updated height cannot precede creation")
	}
	if !IsSubdomainPolicy(r.SubdomainPolicy) {
		return errors.New("identity subdomain policy is invalid")
	}
	if !IsDomainRentPayerPolicy(r.RentPayerPolicy) {
		return errors.New("identity domain rent payer policy is invalid")
	}
	if r.LastStorageChargeHeight == 0 {
		return errors.New("identity domain last storage charge height must be positive")
	}
	if r.LastStorageChargeHeight < r.CreatedHeight {
		return errors.New("identity domain storage charge height cannot precede creation")
	}
	if params.NFTBindingEnabled {
		if !r.NFTBinding.Enabled {
			return errors.New("identity NFT binding is required when enabled")
		}
		if r.NFTBinding.Owner != r.Owner {
			return errors.New("identity NFT binding owner must match name owner")
		}
	}
	return nil
}

func (r ResolverRecord) Normalize(params IdentityRootParams) ResolverRecord {
	r.Name, _ = NormalizeName(r.Name, params.RootNamespace)
	r.ResolverRoot = normalizeResolverRoot(r.ResolverRoot)
	return r
}

func (r ResolverRecord) Validate(params IdentityRootParams) error {
	r = r.Normalize(params)
	if err := ValidateName(r.Name, params); err != nil {
		return err
	}
	if err := ValidateResolverRoot(r.ResolverRoot); err != nil {
		return err
	}
	if r.UpdatedHeight == 0 {
		return errors.New("identity resolver updated height must be positive")
	}
	return nil
}

func (r ReverseRecord) Normalize(params IdentityRootParams) ReverseRecord {
	r.Address = strings.TrimSpace(r.Address)
	r.Name, _ = NormalizeName(r.Name, params.RootNamespace)
	r.Owner = strings.TrimSpace(r.Owner)
	return r
}

func (r ReverseRecord) Validate(params IdentityRootParams) error {
	r = r.Normalize(params)
	if err := ValidateUserFacingAEAddress("identity reverse address", r.Address); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("identity reverse owner", r.Owner); err != nil {
		return err
	}
	if err := ValidateName(r.Name, params); err != nil {
		return err
	}
	if r.UpdatedHeight == 0 {
		return errors.New("identity reverse updated height must be positive")
	}
	return nil
}

func (b IdentityNFTBindingReference) Normalize(params IdentityRootParams) IdentityNFTBindingReference {
	b.Name, _ = NormalizeOptionalName(b.Name, params.RootNamespace)
	b.ClassID = strings.TrimSpace(b.ClassID)
	b.NFTID = strings.TrimSpace(b.NFTID)
	b.Owner = strings.TrimSpace(b.Owner)
	return b
}

func (b IdentityNFTBindingReference) Validate(params IdentityRootParams) error {
	b = b.Normalize(params)
	if !b.Enabled {
		return nil
	}
	if b.Name != "" {
		if err := ValidateName(b.Name, params); err != nil {
			return err
		}
	}
	if b.ClassID == "" || b.NFTID == "" {
		return errors.New("identity NFT binding class id, nft id, and owner are required")
	}
	if err := ValidateUserFacingAEAddress("identity NFT binding owner", b.Owner); err != nil {
		return err
	}
	return nil
}

func (a RootAuthority) Validate() error {
	a.Authority = strings.TrimSpace(a.Authority)
	a.Role = strings.TrimSpace(a.Role)
	if a.Authority == "" || a.Role == "" {
		return errors.New("identity root authority and role are required")
	}
	return nil
}

func (r ReservedName) Normalize(params IdentityRootParams) ReservedName {
	r.Name, _ = NormalizeName(r.Name, params.RootNamespace)
	r.Authority = strings.TrimSpace(r.Authority)
	r.Reason = strings.TrimSpace(r.Reason)
	return r
}

func (r ReservedName) Validate(params IdentityRootParams) error {
	r = r.Normalize(params)
	if err := ValidateName(r.Name, params); err != nil {
		return err
	}
	if r.Authority == "" || r.Reason == "" {
		return errors.New("identity reserved name authority and reason are required")
	}
	return nil
}

func NormalizeRootNamespace(root string) (string, error) {
	root = strings.ToLower(strings.TrimSpace(root))
	root = strings.TrimPrefix(root, ".")
	root = strings.TrimSuffix(root, ".")
	if root == "" {
		return "", errors.New("identity root namespace is required")
	}
	if strings.Contains(root, ".") {
		return "", errors.New("identity root namespace must be a single label")
	}
	if err := validateLabel(root); err != nil {
		return "", err
	}
	return root, nil
}

func NormalizeName(name, root string) (string, error) {
	root, err := NormalizeRootNamespace(root)
	if err != nil {
		return "", err
	}
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, ".")
	name = strings.TrimPrefix(name, ".")
	if name == "" {
		return "", errors.New("identity name is required")
	}
	if name == root {
		return name, nil
	}
	if !strings.HasSuffix(name, "."+root) {
		name += "." + root
	}
	return name, nil
}

func NormalizeOptionalName(name, root string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", nil
	}
	return NormalizeName(name, root)
}

func ValidateName(name string, params IdentityRootParams) error {
	name, err := NormalizeName(name, params.RootNamespace)
	if err != nil {
		return err
	}
	if uint32(len(name)) > params.MaxNameBytes {
		return errors.New("identity name exceeds max bytes")
	}
	labels := strings.Split(name, ".")
	for _, label := range labels {
		if err := validateLabel(label); err != nil {
			return err
		}
	}
	root, _ := NormalizeRootNamespace(params.RootNamespace)
	if labels[len(labels)-1] != root {
		return errors.New("identity name must be under root namespace")
	}
	return nil
}

func ParentName(name, root string) (string, error) {
	name, err := NormalizeName(name, root)
	if err != nil {
		return "", err
	}
	parts := strings.Split(name, ".")
	if len(parts) <= 2 {
		return "", nil
	}
	return strings.Join(parts[1:], "."), nil
}

func ChildName(label, parentName, root string) (string, error) {
	parentName, err := NormalizeName(parentName, root)
	if err != nil {
		return "", err
	}
	label = strings.ToLower(strings.TrimSpace(label))
	if err := validateLabel(label); err != nil {
		return "", err
	}
	return label + "." + parentName, nil
}

func IsActive(record NameRecord, height uint64) bool {
	return height > 0 && record.ExpiryHeight > height
}

func IsSubdomainPolicy(policy string) bool {
	switch policy {
	case SubdomainPolicyOwnerOnly, SubdomainPolicyPublic, SubdomainPolicyDisabled:
		return true
	default:
		return false
	}
}

func IsDomainRentPayerPolicy(policy string) bool {
	switch policy {
	case DomainRentPayerOwner, DomainRentPayerProtocol:
		return true
	default:
		return false
	}
}

func ValidateUserFacingAEAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("%s must use AE user-facing address format", field)
	}
	return addressing.ValidateUserAddress(field, text)
}

func DomainStorageSize(record NameRecord) uint64 {
	record.ResolverRoot = normalizeResolverRoot(record.ResolverRoot)
	return uint64(len(record.Name) + len(record.ParentName) + len(record.Owner) + len(record.ResolverRoot) + len(record.SubdomainPolicy) + len(record.RentPayerPolicy))
}

func DomainStorageRentDelta(record NameRecord, params IdentityRootParams, height uint64) (uint64, error) {
	record = record.Normalize(params)
	if height < record.LastStorageChargeHeight {
		return 0, errors.New("identity domain rent height cannot go backwards")
	}
	elapsed := height - record.LastStorageChargeHeight
	size := DomainStorageSize(record)
	if size != 0 && elapsed > ^uint64(0)/size {
		return 0, errors.New("identity domain storage rent overflow")
	}
	usage := size * elapsed
	if params.DomainRentRatePerByteBlock != 0 && usage > ^uint64(0)/params.DomainRentRatePerByteBlock {
		return 0, errors.New("identity domain storage rent overflow")
	}
	return usage * params.DomainRentRatePerByteBlock, nil
}

func ValidateResolverRoot(root string) error {
	root = normalizeResolverRoot(root)
	if len(root) != 64 {
		return errors.New("identity resolver root must be 32-byte hex")
	}
	if _, err := hex.DecodeString(root); err != nil {
		return fmt.Errorf("identity resolver root must be hex: %w", err)
	}
	return nil
}

func SortRecords(records []NameRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Name < records[j].Name })
}

func SortResolvers(resolvers []ResolverRecord) {
	sort.SliceStable(resolvers, func(i, j int) bool { return resolvers[i].Name < resolvers[j].Name })
}

func SortReverseRecords(records []ReverseRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Address < records[j].Address })
}

func SortBindings(bindings []IdentityNFTBindingReference) {
	sort.SliceStable(bindings, func(i, j int) bool {
		if bindings[i].Name != bindings[j].Name {
			return bindings[i].Name < bindings[j].Name
		}
		return bindings[i].NFTID < bindings[j].NFTID
	})
}

func SortAuthorities(authorities []RootAuthority) {
	sort.SliceStable(authorities, func(i, j int) bool { return authorities[i].Authority < authorities[j].Authority })
}

func SortReserved(names []ReservedName) {
	sort.SliceStable(names, func(i, j int) bool { return names[i].Name < names[j].Name })
}

func normalizeResolverRoot(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return DefaultResolverRoot
	}
	return root
}

func validateLabel(label string) error {
	if label == "" || len(label) > 63 {
		return errors.New("identity name label length is invalid")
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return errors.New("identity name label cannot start or end with hyphen")
	}
	for _, ch := range label {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			continue
		}
		return errors.New("identity name label contains invalid character")
	}
	return nil
}

func cloneRecords(records []NameRecord) []NameRecord {
	out := append([]NameRecord(nil), records...)
	for i := range out {
		out[i].Name = strings.ToLower(strings.TrimSpace(out[i].Name))
		out[i].ParentName = strings.ToLower(strings.TrimSpace(out[i].ParentName))
		out[i].Owner = strings.TrimSpace(out[i].Owner)
		out[i].ResolverRoot = normalizeResolverRoot(out[i].ResolverRoot)
		out[i].SubdomainPolicy = strings.TrimSpace(out[i].SubdomainPolicy)
	}
	return out
}

func cloneResolvers(records []ResolverRecord) []ResolverRecord {
	out := append([]ResolverRecord(nil), records...)
	for i := range out {
		out[i].Name = strings.ToLower(strings.TrimSpace(out[i].Name))
		out[i].ResolverRoot = normalizeResolverRoot(out[i].ResolverRoot)
	}
	return out
}

func cloneReverseRecords(records []ReverseRecord) []ReverseRecord {
	out := append([]ReverseRecord(nil), records...)
	for i := range out {
		out[i].Address = strings.TrimSpace(out[i].Address)
		out[i].Name = strings.ToLower(strings.TrimSpace(out[i].Name))
		out[i].Owner = strings.TrimSpace(out[i].Owner)
	}
	return out
}

func cloneBindings(bindings []IdentityNFTBindingReference) []IdentityNFTBindingReference {
	out := append([]IdentityNFTBindingReference(nil), bindings...)
	for i := range out {
		out[i].Name = strings.ToLower(strings.TrimSpace(out[i].Name))
		out[i].ClassID = strings.TrimSpace(out[i].ClassID)
		out[i].NFTID = strings.TrimSpace(out[i].NFTID)
		out[i].Owner = strings.TrimSpace(out[i].Owner)
	}
	return out
}

func cloneAuthorities(authorities []RootAuthority) []RootAuthority {
	out := append([]RootAuthority(nil), authorities...)
	for i := range out {
		out[i].Authority = strings.TrimSpace(out[i].Authority)
		out[i].Role = strings.TrimSpace(out[i].Role)
	}
	return out
}

func cloneReserved(names []ReservedName) []ReservedName {
	out := append([]ReservedName(nil), names...)
	for i := range out {
		out[i].Name = strings.ToLower(strings.TrimSpace(out[i].Name))
		out[i].Authority = strings.TrimSpace(out[i].Authority)
		out[i].Reason = strings.TrimSpace(out[i].Reason)
	}
	return out
}
