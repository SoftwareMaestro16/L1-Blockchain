package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityZoneID		= "IDENTITY_ZONE"
	IdentityZonePrefix	= "identity"
	IdentityZoneStoreV2	= IdentityStoreV2Prefix
	IdentityZoneShardKey	= "name_hash"

	IdentityZoneDomainPrefix	= IdentityZonePrefix + "/domains"
	IdentityZoneResolverPrefix	= IdentityZonePrefix + "/resolvers"
	IdentityZoneReversePrefix	= IdentityZonePrefix + "/reverse"
	IdentityZoneNFTBindingPrefix	= IdentityZonePrefix + "/nft_bindings"
	IdentityZoneGrantPrefix		= IdentityZonePrefix + "/grants"
	IdentityZoneAuctionPrefix	= IdentityZonePrefix + "/auctions"
	IdentityZoneProofIndexPrefix	= IdentityZonePrefix + "/proofs/index"
	IdentityZoneResolverHookPrefix	= IdentityZonePrefix + "/resolver_vm/hooks"
	IdentityZoneResolutionGraphRoot	= IdentityZonePrefix + "/resolution_graph"
)

type IdentityZoneMessageKind string
type IdentityZoneProofKind string

const (
	IdentityMessageLookupRequest	IdentityZoneMessageKind	= "identity.lookup_request"
	IdentityMessageLookupResponse	IdentityZoneMessageKind	= "identity.lookup_response"
	IdentityMessageAuctionFinalize	IdentityZoneMessageKind	= "identity.auction_finalize"

	IdentityMessageRegisterIdentity		IdentityZoneMessageKind	= "MsgRegisterIdentity"
	IdentityMessageRenewIdentity		IdentityZoneMessageKind	= "MsgRenewIdentity"
	IdentityMessageTransferIdentity		IdentityZoneMessageKind	= "MsgTransferIdentity"
	IdentityMessageUpdateResolver		IdentityZoneMessageKind	= "MsgUpdateResolver"
	IdentityMessageSetReverse		IdentityZoneMessageKind	= "MsgSetReverse"
	IdentityMessageGrantIdentityPermission	IdentityZoneMessageKind	= "MsgGrantIdentityPermission"
	IdentityMessageRevokeIdentityPermission	IdentityZoneMessageKind	= "MsgRevokeIdentityPermission"
	IdentityMessageStartIdentityAuction	IdentityZoneMessageKind	= "MsgStartIdentityAuction"
	IdentityMessageFinalizeIdentityAuction	IdentityZoneMessageKind	= "MsgFinalizeIdentityAuction"

	IdentityProofDomain		IdentityZoneProofKind	= "QueryDomain"
	IdentityProofOwnershipProof	IdentityZoneProofKind	= "QueryOwnershipProof"
	IdentityProofIdentityGraph	IdentityZoneProofKind	= "QueryIdentityGraph"
	IdentityProofIdentityRoot	IdentityZoneProofKind	= "QueryIdentityRoot"

	IdentityProofResolver	IdentityZoneProofKind	= "resolver"
	IdentityProofReverse	IdentityZoneProofKind	= "reverse"
	IdentityProofNFTBinding	IdentityZoneProofKind	= "nft_binding"
	IdentityProofAuction	IdentityZoneProofKind	= "auction"
)

type IdentityZoneStateMachineDescriptor struct {
	ZoneID		string
	StorePrefix	string
	ShardStrategy	string
	MessageHandlers	[]IdentityZoneMessageKind
	ProofQueries	[]IdentityZoneProofKind
}

type IdentityLookupMessage struct {
	RequestID	string
	QueryDomain	string
	RecordKey	string
	SourceZone	string
	SourceShard	string
	ReplyTo		string
	Height		uint64
	PayloadHash	string
	MessageHash	string
}

type IdentityResponseReceipt struct {
	RequestID	string
	QueryDomain	string
	ResolverDomain	string
	ResponseHash	string
	Height		uint64
	Success		bool
	ReceiptHash	string
}

type IdentityAuctionFinalizationReceipt struct {
	Domain		string
	WinnerHash	string
	Height		uint64
	AuctionHash	string
	ReceiptHash	string
}

type IdentityResolverVMHook struct {
	HookID		string
	NameHash	string
	RecordKey	string
	InputHash	string
	OutputHash	string
	GasLimit	uint64
	Version		uint64
	HookHash	string
}

type IdentityResolutionGraphNode struct {
	NodeID		string
	NameHash	string
	RecordKey	string
	TargetHash	string
}

type IdentityResolutionGraphEdge struct {
	FromNodeID	string
	ToNodeID	string
	EdgeKind	string
}

type IdentityResolutionGraph struct {
	Height		uint64
	Nodes		[]IdentityResolutionGraphNode
	Edges		[]IdentityResolutionGraphEdge
	GraphHash	string
}

type IdentityCrossZoneBinding struct {
	NameHash	string
	ZoneID		string
	BindingKey	string
	BindingRoot	string
	Height		uint64
	BindingHash	string
}

type IdentityZoneProofIndexEntry struct {
	Height		uint64
	NameHash	string
	ProofKind	IdentityZoneProofKind
	ProofRoot	string
	ProofHash	string
	IndexHash	string
}

type IdentityZoneRoots struct {
	Height			uint64
	DomainRoot		string
	ResolverRoot		string
	ReverseRoot		string
	NFTBindingRoot		string
	GrantRoot		string
	AuctionRoot		string
	ProofIndexRoot		string
	ResolverHookRoot	string
	GraphRoot		string
	CrossZoneRoot		string
	StateRoot		string
}

func DefaultIdentityZoneStateMachineDescriptor() IdentityZoneStateMachineDescriptor {
	return IdentityZoneStateMachineDescriptor{
		ZoneID:		IdentityZoneID,
		StorePrefix:	IdentityStoreV2Prefix,
		ShardStrategy:	IdentityZoneShardKey,
		MessageHandlers: []IdentityZoneMessageKind{
			IdentityMessageRegisterIdentity,
			IdentityMessageRenewIdentity,
			IdentityMessageTransferIdentity,
			IdentityMessageUpdateResolver,
			IdentityMessageSetReverse,
			IdentityMessageGrantIdentityPermission,
			IdentityMessageRevokeIdentityPermission,
			IdentityMessageStartIdentityAuction,
			IdentityMessageFinalizeIdentityAuction,
			IdentityMessageLookupRequest,
			IdentityMessageLookupResponse,
			IdentityMessageAuctionFinalize,
		},
		ProofQueries: []IdentityZoneProofKind{
			IdentityProofDomain,
			IdentityProofResolver,
			IdentityProofReverse,
			IdentityProofNFTBinding,
			IdentityProofAuction,
			IdentityProofOwnershipProof,
			IdentityProofIdentityGraph,
			IdentityProofIdentityRoot,
		},
	}
}

func (d IdentityZoneStateMachineDescriptor) Validate() error {
	if d.ZoneID != IdentityZoneID {
		return errors.New("identity zone descriptor must use IDENTITY_ZONE")
	}
	if d.StorePrefix != IdentityStoreV2Prefix {
		return fmt.Errorf("identity zone store prefix must be %q", IdentityStoreV2Prefix)
	}
	if d.ShardStrategy != IdentityZoneShardKey {
		return fmt.Errorf("identity zone shard strategy must be %q", IdentityZoneShardKey)
	}
	if len(d.MessageHandlers) == 0 || len(d.ProofQueries) == 0 {
		return errors.New("identity zone descriptor requires handlers and proof queries")
	}
	for _, handler := range d.MessageHandlers {
		if !IsIdentityZoneMessageKind(handler) {
			return fmt.Errorf("unknown identity zone message handler %q", handler)
		}
	}
	for _, query := range d.ProofQueries {
		if !IsIdentityZoneProofKind(query) {
			return fmt.Errorf("unknown identity zone proof query %q", query)
		}
	}
	return nil
}

func IdentityZoneDomainKeyByHash(nameHash string) (string, error) {
	if err := validateHexHash("identity zone domain name hash", nameHash); err != nil {
		return "", err
	}
	return IdentityZoneDomainPrefix + "/" + nameHash, nil
}

func IdentityZoneDomainKey(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return IdentityZoneDomainKeyByHash(nameHash)
}

func IdentityZoneResolverKeyByHash(nameHash string) (string, error) {
	if err := validateHexHash("identity zone resolver name hash", nameHash); err != nil {
		return "", err
	}
	return IdentityZoneResolverPrefix + "/" + nameHash, nil
}

func IdentityZoneResolverKey(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return IdentityZoneResolverKeyByHash(nameHash)
}

func IdentityZoneReverseKey(address sdk.AccAddress) (string, error) {
	if err := validateSpecAddress("identity zone reverse address", address); err != nil {
		return "", err
	}
	return IdentityZoneReversePrefix + "/" + fmt.Sprintf("%x", []byte(address)), nil
}

func IdentityZoneNFTBindingKeyByHash(nameHash string) (string, error) {
	if err := validateHexHash("identity zone nft binding name hash", nameHash); err != nil {
		return "", err
	}
	return IdentityZoneNFTBindingPrefix + "/" + nameHash, nil
}

func IdentityZoneGrantKey(nameHash string, grantee sdk.AccAddress, scope DelegationScopeV2) (string, error) {
	if err := validateHexHash("identity zone grant name hash", nameHash); err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity zone grant grantee", grantee); err != nil {
		return "", err
	}
	if err := validateDelegationScopeV2(scope); err != nil {
		return "", err
	}
	return IdentityZoneGrantPrefix + "/" + nameHash + "/" + fmt.Sprintf("%x", []byte(grantee)) + "/" + string(scope), nil
}

func IdentityZoneAuctionKey(auctionID string) (string, error) {
	if err := validateHexHash("identity zone auction id", auctionID); err != nil {
		return "", err
	}
	return IdentityZoneAuctionPrefix + "/" + auctionID, nil
}

func IdentityZoneProofIndexKey(height uint64, nameHash string) (string, error) {
	if height == 0 {
		return "", errors.New("identity zone proof index height must be positive")
	}
	if err := validateHexHash("identity zone proof index name hash", nameHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%020d/%s", IdentityZoneProofIndexPrefix, height, nameHash), nil
}

func NewIdentityLookupMessage(msg IdentityLookupMessage) (IdentityLookupMessage, error) {
	if msg.MessageHash != "" {
		return IdentityLookupMessage{}, errors.New("identity lookup message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return IdentityLookupMessage{}, err
	}
	msg.MessageHash = ComputeIdentityLookupMessageHash(msg)
	return msg, msg.Validate()
}

func (m IdentityLookupMessage) ValidateFormat() error {
	if _, err := NormalizeAETDomain(m.QueryDomain); err != nil {
		return err
	}
	if m.RequestID == "" || m.SourceZone == "" || m.SourceShard == "" || m.ReplyTo == "" {
		return errors.New("identity lookup message route fields are required")
	}
	if m.Height == 0 {
		return errors.New("identity lookup message height must be positive")
	}
	if m.RecordKey != "" {
		if err := ValidateResolverGrantKey(m.RecordKey); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity lookup payload hash", m.PayloadHash); err != nil {
		return err
	}
	if m.MessageHash != "" {
		return validateHexHash("identity lookup message hash", m.MessageHash)
	}
	return nil
}

func (m IdentityLookupMessage) Validate() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.MessageHash == "" {
		return errors.New("identity lookup message hash is required")
	}
	if m.MessageHash != ComputeIdentityLookupMessageHash(m) {
		return errors.New("identity lookup message hash mismatch")
	}
	return nil
}

func NewIdentityResponseReceipt(receipt IdentityResponseReceipt) (IdentityResponseReceipt, error) {
	if receipt.ReceiptHash != "" {
		return IdentityResponseReceipt{}, errors.New("identity response receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return IdentityResponseReceipt{}, err
	}
	receipt.ReceiptHash = ComputeIdentityResponseReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (r IdentityResponseReceipt) ValidateFormat() error {
	if _, err := NormalizeAETDomain(r.QueryDomain); err != nil {
		return err
	}
	if r.ResolverDomain != "" {
		if _, err := NormalizeAETDomain(r.ResolverDomain); err != nil {
			return err
		}
	}
	if r.RequestID == "" {
		return errors.New("identity response request id is required")
	}
	if r.Height == 0 {
		return errors.New("identity response height must be positive")
	}
	if err := validateHexHash("identity response hash", r.ResponseHash); err != nil {
		return err
	}
	if r.ReceiptHash != "" {
		return validateHexHash("identity response receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r IdentityResponseReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("identity response receipt hash is required")
	}
	if r.ReceiptHash != ComputeIdentityResponseReceiptHash(r) {
		return errors.New("identity response receipt hash mismatch")
	}
	return nil
}

func NewIdentityAuctionFinalizationReceipt(receipt IdentityAuctionFinalizationReceipt) (IdentityAuctionFinalizationReceipt, error) {
	if receipt.ReceiptHash != "" {
		return IdentityAuctionFinalizationReceipt{}, errors.New("identity auction receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return IdentityAuctionFinalizationReceipt{}, err
	}
	receipt.ReceiptHash = ComputeIdentityAuctionFinalizationReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (r IdentityAuctionFinalizationReceipt) ValidateFormat() error {
	if _, err := NormalizeAETDomain(r.Domain); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("identity auction receipt height must be positive")
	}
	if err := validateHexHash("identity auction winner hash", r.WinnerHash); err != nil {
		return err
	}
	if err := validateHexHash("identity auction hash", r.AuctionHash); err != nil {
		return err
	}
	if r.ReceiptHash != "" {
		return validateHexHash("identity auction receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r IdentityAuctionFinalizationReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("identity auction receipt hash is required")
	}
	if r.ReceiptHash != ComputeIdentityAuctionFinalizationReceiptHash(r) {
		return errors.New("identity auction receipt hash mismatch")
	}
	return nil
}

func BuildIdentityResolverProof(state IdentityState, name string) (IdentityInclusionProof, error) {
	key, err := IdentityResolverStoreKey(name)
	if err != nil {
		return IdentityInclusionProof{}, err
	}
	return BuildIdentityProof(state, key)
}

func BuildIdentityReverseLookupProof(state IdentityState, address sdk.AccAddress) (IdentityInclusionProof, error) {
	key, err := IdentityReverseStoreKey(address)
	if err != nil {
		return IdentityInclusionProof{}, err
	}
	return BuildIdentityProof(state, key)
}

func CheckIdentityNFTBinding(state IdentityState, name string, height uint64) error {
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return err
	}
	nft, found := findDomainNFTByID(state, domain.NFTID)
	if !found {
		return errors.New("identity zone nft binding not found")
	}
	if nft.Domain != domain.Name || nft.ID != domain.NFTID {
		return errors.New("identity zone nft binding domain mismatch")
	}
	if string(nft.Owner) != string(domain.Owner) {
		return errors.New("identity zone nft binding owner mismatch")
	}
	return nil
}

func ComputeIdentityLookupMessageHash(msg IdentityLookupMessage) string {
	return identityHash(
		"identity-zone-lookup-message-v1",
		msg.RequestID,
		msg.QueryDomain,
		msg.RecordKey,
		msg.SourceZone,
		msg.SourceShard,
		msg.ReplyTo,
		fmt.Sprintf("%020d", msg.Height),
		msg.PayloadHash,
	)
}

func ComputeIdentityResponseReceiptHash(receipt IdentityResponseReceipt) string {
	return identityHash(
		"identity-zone-response-receipt-v1",
		receipt.RequestID,
		receipt.QueryDomain,
		receipt.ResolverDomain,
		receipt.ResponseHash,
		fmt.Sprintf("%020d", receipt.Height),
		fmt.Sprintf("%t", receipt.Success),
	)
}

func ComputeIdentityAuctionFinalizationReceiptHash(receipt IdentityAuctionFinalizationReceipt) string {
	return identityHash(
		"identity-zone-auction-finalization-v1",
		receipt.Domain,
		receipt.WinnerHash,
		fmt.Sprintf("%020d", receipt.Height),
		receipt.AuctionHash,
	)
}

func NewIdentityResolverVMHook(hook IdentityResolverVMHook) (IdentityResolverVMHook, error) {
	if hook.HookHash != "" {
		return IdentityResolverVMHook{}, errors.New("identity resolver VM hook hash must be empty before construction")
	}
	if err := hook.ValidateFormat(); err != nil {
		return IdentityResolverVMHook{}, err
	}
	hook.HookHash = ComputeIdentityResolverVMHookHash(hook)
	return hook, hook.Validate()
}

func (h IdentityResolverVMHook) ValidateFormat() error {
	if h.HookID == "" {
		return errors.New("identity resolver VM hook id is required")
	}
	if err := validateHexHash("identity resolver VM hook name hash", h.NameHash); err != nil {
		return err
	}
	if err := ValidateResolverGrantKey(h.RecordKey); err != nil {
		return err
	}
	if err := validateHexHash("identity resolver VM hook input hash", h.InputHash); err != nil {
		return err
	}
	if err := validateHexHash("identity resolver VM hook output hash", h.OutputHash); err != nil {
		return err
	}
	if h.GasLimit == 0 {
		return errors.New("identity resolver VM hook gas limit must be positive")
	}
	if h.Version == 0 {
		return errors.New("identity resolver VM hook version must be positive")
	}
	if h.HookHash != "" {
		return validateHexHash("identity resolver VM hook hash", h.HookHash)
	}
	return nil
}

func (h IdentityResolverVMHook) Validate() error {
	if err := h.ValidateFormat(); err != nil {
		return err
	}
	if h.HookHash == "" {
		return errors.New("identity resolver VM hook hash is required")
	}
	if h.HookHash != ComputeIdentityResolverVMHookHash(h) {
		return errors.New("identity resolver VM hook hash mismatch")
	}
	return nil
}

func NewIdentityResolutionGraph(graph IdentityResolutionGraph) (IdentityResolutionGraph, error) {
	if graph.GraphHash != "" {
		return IdentityResolutionGraph{}, errors.New("identity resolution graph hash must be empty before construction")
	}
	graph.Nodes = normalizeIdentityResolutionGraphNodes(graph.Nodes)
	graph.Edges = normalizeIdentityResolutionGraphEdges(graph.Edges)
	if err := graph.ValidateFormat(); err != nil {
		return IdentityResolutionGraph{}, err
	}
	graph.GraphHash = ComputeIdentityResolutionGraphHash(graph)
	return graph, graph.Validate()
}

func (g IdentityResolutionGraph) ValidateFormat() error {
	if g.Height == 0 {
		return errors.New("identity resolution graph height must be positive")
	}
	if len(g.Nodes) == 0 {
		return errors.New("identity resolution graph requires at least one node")
	}
	for i, node := range g.Nodes {
		if err := node.Validate(); err != nil {
			return err
		}
		if i > 0 && g.Nodes[i-1].NodeID >= node.NodeID {
			return errors.New("identity resolution graph nodes must be sorted canonically")
		}
	}
	for i, edge := range g.Edges {
		if err := edge.Validate(); err != nil {
			return err
		}
		if i > 0 && compareIdentityGraphEdges(g.Edges[i-1], edge) >= 0 {
			return errors.New("identity resolution graph edges must be sorted canonically")
		}
	}
	if g.GraphHash != "" {
		return validateHexHash("identity resolution graph hash", g.GraphHash)
	}
	return nil
}

func (g IdentityResolutionGraph) Validate() error {
	if err := g.ValidateFormat(); err != nil {
		return err
	}
	if g.GraphHash == "" {
		return errors.New("identity resolution graph hash is required")
	}
	if g.GraphHash != ComputeIdentityResolutionGraphHash(g) {
		return errors.New("identity resolution graph hash mismatch")
	}
	return nil
}

func (n IdentityResolutionGraphNode) Validate() error {
	if n.NodeID == "" {
		return errors.New("identity resolution graph node id is required")
	}
	if err := validateHexHash("identity resolution graph node name hash", n.NameHash); err != nil {
		return err
	}
	if n.RecordKey != "" {
		if err := ValidateResolverGrantKey(n.RecordKey); err != nil {
			return err
		}
	}
	return validateHexHash("identity resolution graph node target hash", n.TargetHash)
}

func (e IdentityResolutionGraphEdge) Validate() error {
	if e.FromNodeID == "" || e.ToNodeID == "" || e.EdgeKind == "" {
		return errors.New("identity resolution graph edge route fields are required")
	}
	return nil
}

func NewIdentityCrossZoneBinding(binding IdentityCrossZoneBinding) (IdentityCrossZoneBinding, error) {
	if binding.BindingHash != "" {
		return IdentityCrossZoneBinding{}, errors.New("identity cross-zone binding hash must be empty before construction")
	}
	if err := binding.ValidateFormat(); err != nil {
		return IdentityCrossZoneBinding{}, err
	}
	binding.BindingHash = ComputeIdentityCrossZoneBindingHash(binding)
	return binding, binding.Validate()
}

func (b IdentityCrossZoneBinding) ValidateFormat() error {
	if err := validateHexHash("identity cross-zone binding name hash", b.NameHash); err != nil {
		return err
	}
	if b.ZoneID == "" || b.BindingKey == "" {
		return errors.New("identity cross-zone binding zone id and key are required")
	}
	if err := validateHexHash("identity cross-zone binding root", b.BindingRoot); err != nil {
		return err
	}
	if b.Height == 0 {
		return errors.New("identity cross-zone binding height must be positive")
	}
	if b.BindingHash != "" {
		return validateHexHash("identity cross-zone binding hash", b.BindingHash)
	}
	return nil
}

func (b IdentityCrossZoneBinding) Validate() error {
	if err := b.ValidateFormat(); err != nil {
		return err
	}
	if b.BindingHash == "" {
		return errors.New("identity cross-zone binding hash is required")
	}
	if b.BindingHash != ComputeIdentityCrossZoneBindingHash(b) {
		return errors.New("identity cross-zone binding hash mismatch")
	}
	return nil
}

func NewIdentityZoneProofIndexEntry(entry IdentityZoneProofIndexEntry) (IdentityZoneProofIndexEntry, error) {
	if entry.IndexHash != "" {
		return IdentityZoneProofIndexEntry{}, errors.New("identity proof index hash must be empty before construction")
	}
	if err := entry.ValidateFormat(); err != nil {
		return IdentityZoneProofIndexEntry{}, err
	}
	entry.IndexHash = ComputeIdentityZoneProofIndexEntryHash(entry)
	return entry, entry.Validate()
}

func (e IdentityZoneProofIndexEntry) ValidateFormat() error {
	if e.Height == 0 {
		return errors.New("identity proof index height must be positive")
	}
	if err := validateHexHash("identity proof index name hash", e.NameHash); err != nil {
		return err
	}
	if !IsIdentityZoneProofKind(e.ProofKind) {
		return fmt.Errorf("unknown identity proof index kind %q", e.ProofKind)
	}
	if err := validateHexHash("identity proof index root", e.ProofRoot); err != nil {
		return err
	}
	if err := validateHexHash("identity proof index proof hash", e.ProofHash); err != nil {
		return err
	}
	if e.IndexHash != "" {
		return validateHexHash("identity proof index hash", e.IndexHash)
	}
	return nil
}

func (e IdentityZoneProofIndexEntry) Validate() error {
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.IndexHash == "" {
		return errors.New("identity proof index hash is required")
	}
	if e.IndexHash != ComputeIdentityZoneProofIndexEntryHash(e) {
		return errors.New("identity proof index hash mismatch")
	}
	return nil
}

func BuildIdentityZoneRoots(
	height uint64,
	state IdentityState,
	hooks []IdentityResolverVMHook,
	graphs []IdentityResolutionGraph,
	bindings []IdentityCrossZoneBinding,
	proofs []IdentityZoneProofIndexEntry,
) (IdentityZoneRoots, error) {
	if height == 0 {
		return IdentityZoneRoots{}, errors.New("identity zone root height must be positive")
	}
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityZoneRoots{}, err
	}
	roots := IdentityZoneRoots{
		Height:			height,
		DomainRoot:		ComputeIdentityZoneDomainRoot(state.Domains),
		ResolverRoot:		ComputeIdentityZoneResolverRoot(state.Resolvers),
		ReverseRoot:		ComputeIdentityZoneReverseRoot(state.ReverseRecords),
		NFTBindingRoot:		ComputeIdentityZoneNFTBindingRoot(state.DomainNFTs),
		GrantRoot:		ComputeIdentityZoneGrantRoot(nil),
		AuctionRoot:		ComputeIdentityZoneAuctionRoot(state.Auctions),
		ProofIndexRoot:		ComputeIdentityZoneProofIndexRoot(proofs),
		ResolverHookRoot:	ComputeIdentityResolverVMHookRoot(hooks),
		GraphRoot:		ComputeIdentityResolutionGraphRoot(graphs),
		CrossZoneRoot:		ComputeIdentityCrossZoneBindingRoot(bindings),
	}
	roots.StateRoot = identityHash(
		"identity-zone-state-root-v1",
		stateRoot,
		roots.DomainRoot,
		roots.ResolverRoot,
		roots.ReverseRoot,
		roots.NFTBindingRoot,
		roots.GrantRoot,
		roots.AuctionRoot,
		roots.ProofIndexRoot,
		roots.ResolverHookRoot,
		roots.GraphRoot,
		roots.CrossZoneRoot,
	)
	return roots, roots.Validate()
}

func (r IdentityZoneRoots) Validate() error {
	if r.Height == 0 {
		return errors.New("identity zone roots height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "identity zone domain root", value: r.DomainRoot},
		{name: "identity zone resolver root", value: r.ResolverRoot},
		{name: "identity zone reverse root", value: r.ReverseRoot},
		{name: "identity zone nft binding root", value: r.NFTBindingRoot},
		{name: "identity zone grant root", value: r.GrantRoot},
		{name: "identity zone auction root", value: r.AuctionRoot},
		{name: "identity zone proof index root", value: r.ProofIndexRoot},
		{name: "identity zone resolver hook root", value: r.ResolverHookRoot},
		{name: "identity zone graph root", value: r.GraphRoot},
		{name: "identity zone cross-zone root", value: r.CrossZoneRoot},
		{name: "identity zone state root", value: r.StateRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func QueryIdentityZoneLightClientProof(state IdentityState, kind IdentityZoneProofKind, name string, height uint64) (IdentityZoneProofIndexEntry, error) {
	if height == 0 {
		return IdentityZoneProofIndexEntry{}, errors.New("identity zone proof query height must be positive")
	}
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return IdentityZoneProofIndexEntry{}, err
	}
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityZoneProofIndexEntry{}, err
	}
	var proofHash string
	switch kind {
	case IdentityProofDomain, IdentityProofOwnershipProof:
		proof, err := BuildIdentityProof(state, mustIdentityZoneDomainLegacyKey(name))
		if err != nil {
			return IdentityZoneProofIndexEntry{}, err
		}
		proofHash = proof.LeafHash
	case IdentityProofResolver:
		proof, err := BuildIdentityResolverProof(state, name)
		if err != nil {
			return IdentityZoneProofIndexEntry{}, err
		}
		proofHash = proof.LeafHash
	default:
		return IdentityZoneProofIndexEntry{}, fmt.Errorf("identity zone light-client proof kind %q is not name-based", kind)
	}
	return NewIdentityZoneProofIndexEntry(IdentityZoneProofIndexEntry{
		Height:		height,
		NameHash:	nameHash,
		ProofKind:	kind,
		ProofRoot:	stateRoot,
		ProofHash:	proofHash,
	})
}

func QueryIdentityZoneReverseLookupProof(state IdentityState, address sdk.AccAddress, height uint64) (IdentityZoneProofIndexEntry, IdentityInclusionProof, error) {
	if height == 0 {
		return IdentityZoneProofIndexEntry{}, IdentityInclusionProof{}, errors.New("identity reverse proof height must be positive")
	}
	proof, err := BuildIdentityReverseLookupProof(state, address)
	if err != nil {
		return IdentityZoneProofIndexEntry{}, IdentityInclusionProof{}, err
	}
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityZoneProofIndexEntry{}, IdentityInclusionProof{}, err
	}
	nameHash := identityHash("identity-zone-reverse-name", fmt.Sprintf("%x", []byte(address)))
	entry, err := NewIdentityZoneProofIndexEntry(IdentityZoneProofIndexEntry{
		Height:		height,
		NameHash:	nameHash,
		ProofKind:	IdentityProofReverse,
		ProofRoot:	stateRoot,
		ProofHash:	proof.LeafHash,
	})
	return entry, proof, err
}

func ComputeIdentityResolverVMHookHash(hook IdentityResolverVMHook) string {
	return identityHash(
		"identity-resolver-vm-hook-v1",
		hook.HookID,
		hook.NameHash,
		hook.RecordKey,
		hook.InputHash,
		hook.OutputHash,
		fmt.Sprintf("%020d", hook.GasLimit),
		fmt.Sprintf("%020d", hook.Version),
	)
}

func ComputeIdentityResolutionGraphHash(graph IdentityResolutionGraph) string {
	nodes := normalizeIdentityResolutionGraphNodes(graph.Nodes)
	edges := normalizeIdentityResolutionGraphEdges(graph.Edges)
	parts := []string{"identity-resolution-graph-v1", fmt.Sprintf("%020d", graph.Height), fmt.Sprintf("%020d", len(nodes))}
	for _, node := range nodes {
		parts = append(parts, node.NodeID, node.NameHash, node.RecordKey, node.TargetHash)
	}
	parts = append(parts, fmt.Sprintf("%020d", len(edges)))
	for _, edge := range edges {
		parts = append(parts, edge.FromNodeID, edge.ToNodeID, edge.EdgeKind)
	}
	return identityHash(parts...)
}

func ComputeIdentityCrossZoneBindingHash(binding IdentityCrossZoneBinding) string {
	return identityHash(
		"identity-cross-zone-binding-v1",
		binding.NameHash,
		binding.ZoneID,
		binding.BindingKey,
		binding.BindingRoot,
		fmt.Sprintf("%020d", binding.Height),
	)
}

func ComputeIdentityZoneProofIndexEntryHash(entry IdentityZoneProofIndexEntry) string {
	return identityHash(
		"identity-zone-proof-index-entry-v1",
		fmt.Sprintf("%020d", entry.Height),
		entry.NameHash,
		string(entry.ProofKind),
		entry.ProofRoot,
		entry.ProofHash,
	)
}

func ComputeIdentityZoneDomainRoot(domains []Domain) string {
	ordered := cloneDomains(domains)
	sortDomains(ordered)
	parts := []string{"identity-zone-domain-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, domain := range ordered {
		parts = append(parts, domain.Name, string(domain.Owner), domain.NFTID, fmt.Sprintf("%020d", domain.ExpiryHeight), fmt.Sprintf("%020d", domain.UpdatedHeight))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneResolverRoot(resolvers []ResolverRecord) string {
	ordered := cloneResolvers(resolvers)
	sortResolvers(ordered)
	parts := []string{"identity-zone-resolver-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, resolver := range ordered {
		parts = append(parts, resolver.Domain, string(resolver.Owner), string(resolver.Primary), string(resolver.Contract), resolver.ZoneEndpoint, fmt.Sprintf("%020d", resolver.UpdatedAtUnix))
		keys := make([]string, 0, len(resolver.Records))
		for key := range resolver.Records {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			parts = append(parts, key, string(resolver.Records[key]))
		}
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneReverseRoot(records []ReverseRecord) string {
	ordered := cloneReverseRecords(records)
	sortReverseRecords(ordered)
	parts := []string{"identity-zone-reverse-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, record := range ordered {
		parts = append(parts, fmt.Sprintf("%x", []byte(record.Address)), record.Domain, fmt.Sprintf("%020d", record.UpdatedAtUnix))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneNFTBindingRoot(nfts []DomainNFT) string {
	ordered := cloneDomainNFTs(nfts)
	sortDomainNFTs(ordered)
	parts := []string{"identity-zone-nft-binding-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, nft := range ordered {
		parts = append(parts, nft.ID, nft.Domain, string(nft.Owner), fmt.Sprintf("%020d", nft.MintHeight), fmt.Sprintf("%020d", nft.TransferHeight))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneGrantRoot(grants []DelegationRecordV2) string {
	ordered := append([]DelegationRecordV2(nil), grants...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].NameHash != ordered[j].NameHash {
			return ordered[i].NameHash < ordered[j].NameHash
		}
		if string(ordered[i].Delegate) != string(ordered[j].Delegate) {
			return string(ordered[i].Delegate) < string(ordered[j].Delegate)
		}
		return ordered[i].Scope < ordered[j].Scope
	})
	parts := []string{"identity-zone-grant-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, grant := range ordered {
		parts = append(parts, grant.NameHash, string(grant.Delegate), string(grant.Scope), fmt.Sprintf("%020d", grant.ScopeBits), fmt.Sprintf("%020d", grant.ExpiresAtHeight), fmt.Sprintf("%020d", grant.CreatedAtHeight), fmt.Sprintf("%020d", grant.TimeLockedUntilHeight), fmt.Sprintf("%020d", effectiveDelegationVersionV2(grant)), fmt.Sprintf("%t", grant.CanTransferParent))
		parts = append(parts, grant.Permissions...)
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneAuctionRoot(auctions []Auction) string {
	ordered := cloneAuctions(auctions)
	sortAuctions(ordered)
	parts := []string{"identity-zone-auction-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, auction := range ordered {
		parts = append(parts, auction.Name, string(auction.Phase), string(auction.Winner), fmt.Sprintf("%020d", auction.WinningBid), auction.WinningCommitment)
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneProofIndexRoot(entries []IdentityZoneProofIndexEntry) string {
	ordered := append([]IdentityZoneProofIndexEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Height != ordered[j].Height {
			return ordered[i].Height < ordered[j].Height
		}
		if ordered[i].NameHash != ordered[j].NameHash {
			return ordered[i].NameHash < ordered[j].NameHash
		}
		return ordered[i].ProofKind < ordered[j].ProofKind
	})
	parts := []string{"identity-zone-proof-index-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, entry.IndexHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityResolverVMHookRoot(hooks []IdentityResolverVMHook) string {
	ordered := append([]IdentityResolverVMHook(nil), hooks...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].HookID < ordered[j].HookID })
	parts := []string{"identity-resolver-vm-hook-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, hook := range ordered {
		parts = append(parts, hook.HookHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityResolutionGraphRoot(graphs []IdentityResolutionGraph) string {
	ordered := append([]IdentityResolutionGraph(nil), graphs...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].GraphHash < ordered[j].GraphHash })
	parts := []string{"identity-resolution-graph-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, graph := range ordered {
		parts = append(parts, graph.GraphHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityCrossZoneBindingRoot(bindings []IdentityCrossZoneBinding) string {
	ordered := append([]IdentityCrossZoneBinding(nil), bindings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].NameHash != ordered[j].NameHash {
			return ordered[i].NameHash < ordered[j].NameHash
		}
		if ordered[i].ZoneID != ordered[j].ZoneID {
			return ordered[i].ZoneID < ordered[j].ZoneID
		}
		return ordered[i].BindingKey < ordered[j].BindingKey
	})
	parts := []string{"identity-cross-zone-binding-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, binding := range ordered {
		parts = append(parts, binding.BindingHash)
	}
	return identityHash(parts...)
}

func IsIdentityZoneMessageKind(kind IdentityZoneMessageKind) bool {
	switch kind {
	case IdentityMessageRegisterIdentity,
		IdentityMessageRenewIdentity,
		IdentityMessageTransferIdentity,
		IdentityMessageUpdateResolver,
		IdentityMessageSetReverse,
		IdentityMessageGrantIdentityPermission,
		IdentityMessageRevokeIdentityPermission,
		IdentityMessageStartIdentityAuction,
		IdentityMessageFinalizeIdentityAuction,
		IdentityMessageLookupRequest,
		IdentityMessageLookupResponse,
		IdentityMessageAuctionFinalize:
		return true
	default:
		return false
	}
}

func IsIdentityZoneProofKind(kind IdentityZoneProofKind) bool {
	switch kind {
	case IdentityProofDomain,
		IdentityProofResolver,
		IdentityProofReverse,
		IdentityProofNFTBinding,
		IdentityProofAuction,
		IdentityProofOwnershipProof,
		IdentityProofIdentityGraph,
		IdentityProofIdentityRoot:
		return true
	default:
		return false
	}
}

func mustIdentityZoneDomainLegacyKey(name string) string {
	key, err := IdentityDomainStoreKey(name)
	if err != nil {
		return ""
	}
	return key
}

func normalizeIdentityResolutionGraphNodes(nodes []IdentityResolutionGraphNode) []IdentityResolutionGraphNode {
	out := append([]IdentityResolutionGraphNode(nil), nodes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func normalizeIdentityResolutionGraphEdges(edges []IdentityResolutionGraphEdge) []IdentityResolutionGraphEdge {
	out := append([]IdentityResolutionGraphEdge(nil), edges...)
	sort.SliceStable(out, func(i, j int) bool { return compareIdentityGraphEdges(out[i], out[j]) < 0 })
	return out
}

func compareIdentityGraphEdges(left, right IdentityResolutionGraphEdge) int {
	if left.FromNodeID < right.FromNodeID {
		return -1
	}
	if left.FromNodeID > right.FromNodeID {
		return 1
	}
	if left.ToNodeID < right.ToNodeID {
		return -1
	}
	if left.ToNodeID > right.ToNodeID {
		return 1
	}
	if left.EdgeKind < right.EdgeKind {
		return -1
	}
	if left.EdgeKind > right.EdgeKind {
		return 1
	}
	return 0
}
