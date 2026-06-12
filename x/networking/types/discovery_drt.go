package types

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxDRTAdvertisements	= 4096
	DefaultDRTQueryLimit	= uint32(32)
	MaxDRTQueryLimit	= uint32(256)
)

type DRTObjectType string

const (
	DRTObjectNode				DRTObjectType	= "node"
	DRTObjectExecutionZone			DRTObjectType	= "execution_zone"
	DRTObjectServiceEndpoint		DRTObjectType	= "service_endpoint"
	DRTObjectRPCEndpoint			DRTObjectType	= "rpc_endpoint"
	DRTObjectStorageProvider		DRTObjectType	= "storage_provider"
	DRTObjectRoutingEntryPoint		DRTObjectType	= "routing_entry_point"
	DRTObjectOverlayMembershipRecord	DRTObjectType	= "overlay_membership_record"
	DRTObjectStreamProvider			DRTObjectType	= "stream_provider"
)

type DRTAdvertisement struct {
	AdvertisementID		string
	ObjectType		DRTObjectType
	ObjectID		string
	Discovery		DiscoveryRecord
	OverlayID		string
	ZoneID			string
	ServiceID		string
	EndpointHash		string
	StakeWeight		uint64
	PeerScoreBps		uint32
	LeaseStartHeight	uint64
	LeaseExpireHeight	uint64
}

type DRTQuery struct {
	ObjectType	DRTObjectType
	ObjectID	string
	OverlayID	string
	ZoneID		string
	ServiceID	string
	MinStakeWeight	uint64
	Limit		uint32
	CurrentHeight	uint64
}

type DRTBucket struct {
	BucketID	uint32
	Advertisements	[]DRTAdvertisement
}

type DistributedRoutingTable struct {
	Advertisements	[]DRTAdvertisement
	Records		[]DiscoveryRecord
	Revocations	[]DiscoveryRevocation
}

type DiscoveryRevocation struct {
	RevocationID		string
	RecordID		string
	OwnerNodeID		string
	AdvertisementHash	string
	RevokedHeight		uint64
	Signature		[]byte
}

type DiscoverySignatureChainEntry struct {
	NodeID		string
	PublicKey	[]byte
	Signature	[]byte
}

type DiscoveryOnChainProof struct {
	ProofHash	string
	ProofHeight	uint64
	StateRoot	string
}

type DiscoveryResponse struct {
	ResponseID	string
	QueryHash	string
	MatchedRecords	[]DiscoveryRecord
	SignatureChain	[]DiscoverySignatureChainEntry
	OnChainProof	DiscoveryOnChainProof
	ExpiryHeight	uint64
	SourceNodeID	string
	SourceSignature	[]byte
	ResultHash	string
	AdvisoryOnly	bool
	GeneratedHeight	uint64
}

func EmptyDistributedRoutingTable() DistributedRoutingTable {
	return DistributedRoutingTable{}
}

func NewDRTAdvertisement(ad DRTAdvertisement) (DRTAdvertisement, error) {
	ad = NormalizeDRTAdvertisement(ad)
	if ad.ObjectID == "" {
		ad.ObjectID = ComputeDRTObjectID(ad)
	}
	if ad.AdvertisementID == "" {
		ad.AdvertisementID = ComputeDRTAdvertisementID(ad)
	}
	if err := ad.Validate(nil, 0); err != nil {
		return DRTAdvertisement{}, err
	}
	return ad, nil
}

func NormalizeDRTAdvertisement(ad DRTAdvertisement) DRTAdvertisement {
	ad.AdvertisementID = normalizeHashText(ad.AdvertisementID)
	ad.ObjectType = DRTObjectType(strings.ToLower(strings.TrimSpace(string(ad.ObjectType))))
	ad.ObjectID = normalizeHashText(ad.ObjectID)
	ad.Discovery = NormalizeDiscoveryRecord(ad.Discovery)
	ad.OverlayID = normalizeHashText(ad.OverlayID)
	ad.ZoneID = strings.TrimSpace(ad.ZoneID)
	ad.ServiceID = strings.TrimSpace(ad.ServiceID)
	ad.EndpointHash = normalizeHashText(ad.EndpointHash)
	return ad
}

func NormalizeDiscoveryRecord(record DiscoveryRecord) DiscoveryRecord {
	record.RecordID = normalizeHashText(record.RecordID)
	record.RecordType = DRTObjectType(strings.ToLower(strings.TrimSpace(string(record.RecordType))))
	record.OwnerNodeID = normalizeHashText(record.OwnerNodeID)
	record.TargetID = normalizeHashText(record.TargetID)
	record.AdvertisementHash = normalizeHashText(record.AdvertisementHash)
	record.ZoneID = strings.TrimSpace(record.ZoneID)
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.OverlayID = normalizeHashText(record.OverlayID)
	record.ProofHash = normalizeHashText(record.ProofHash)
	record.Record = NormalizeNodeRecord(record.Record)
	record.Signature = cloneBytes(record.Signature)
	return record
}

func NewSignedDiscoveryRecord(record DiscoveryRecord, privateKey ed25519.PrivateKey, networkSalt []byte) (DiscoveryRecord, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return DiscoveryRecord{}, errors.New("networking discovery record private key must be ed25519")
	}
	if len(networkSalt) == 0 {
		return DiscoveryRecord{}, errors.New("networking discovery record network salt is required")
	}
	record = NormalizeDiscoveryRecord(record)
	if err := record.Record.Validate(networkSalt, 0); err != nil {
		return DiscoveryRecord{}, err
	}
	pubKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok || string(pubKey) != string(record.Record.NodePubKey) {
		return DiscoveryRecord{}, errors.New("networking discovery record signer must own node record")
	}
	if record.OwnerNodeID == "" {
		record.OwnerNodeID = record.Record.NodeID
	}
	if record.ExpiresHeight == 0 {
		record.ExpiresHeight = record.Record.ExpiresHeight
	}
	if record.RecordID == "" {
		record.RecordID = ComputeDiscoveryRecordID(record)
	}
	payload, err := DiscoveryRecordSigningPayload(record)
	if err != nil {
		return DiscoveryRecord{}, err
	}
	record.Signature = ed25519.Sign(privateKey, payload)
	if err := ValidateSignedDiscoveryRecord(record, networkSalt, 0); err != nil {
		return DiscoveryRecord{}, err
	}
	return record, nil
}

func RenewDiscoveryRecord(record DiscoveryRecord, expiresHeight uint64, privateKey ed25519.PrivateKey, networkSalt []byte) (DiscoveryRecord, error) {
	record = NormalizeDiscoveryRecord(record)
	if expiresHeight <= record.ExpiresHeight {
		return DiscoveryRecord{}, errors.New("networking discovery renewal expiry must increase")
	}
	record.ExpiresHeight = expiresHeight
	record.RecordID = ""
	record.Signature = nil
	return NewSignedDiscoveryRecord(record, privateKey, networkSalt)
}

func ComputeDiscoveryRecordID(record DiscoveryRecord) string {
	record = NormalizeDiscoveryRecord(record)
	return HashParts(
		"discovery-record",
		string(record.RecordType),
		record.OwnerNodeID,
		record.TargetID,
		record.AdvertisementHash,
		record.ZoneID,
		record.ServiceID,
		record.OverlayID,
		fmt.Sprintf("%d", record.ExpiresHeight),
		record.ProofHash,
	)
}

func DiscoveryRecordSigningPayload(record DiscoveryRecord) ([]byte, error) {
	record = NormalizeDiscoveryRecord(record)
	record.Signature = nil
	return json.Marshal(record)
}

func IsObjectDiscoveryRecord(record DiscoveryRecord) bool {
	record = NormalizeDiscoveryRecord(record)
	return record.RecordID != "" ||
		record.RecordType != "" ||
		record.OwnerNodeID != "" ||
		record.TargetID != "" ||
		record.AdvertisementHash != "" ||
		record.ExpiresHeight != 0 ||
		len(record.Signature) > 0
}

func ValidateSignedDiscoveryRecord(record DiscoveryRecord, networkSalt []byte, currentHeight uint64) error {
	record = NormalizeDiscoveryRecord(record)
	if err := ValidateHash("networking discovery record id", record.RecordID); err != nil {
		return err
	}
	if record.RecordID != ComputeDiscoveryRecordID(record) {
		return errors.New("networking discovery record id mismatch")
	}
	if !IsDRTObjectType(record.RecordType) {
		return fmt.Errorf("unknown networking discovery record type %q", record.RecordType)
	}
	if len(networkSalt) > 0 {
		if err := record.Record.Validate(networkSalt, currentHeight); err != nil {
			return err
		}
	} else if err := record.Record.ValidateBasic(); err != nil {
		return err
	}
	if record.OwnerNodeID != record.Record.NodeID {
		return errors.New("networking discovery record owner must match node record")
	}
	if err := ValidateHash("networking discovery target id", record.TargetID); err != nil {
		return err
	}
	if err := ValidateHash("networking discovery advertisement hash", record.AdvertisementHash); err != nil {
		return err
	}
	if record.OverlayID != "" {
		if err := ValidateHash("networking discovery overlay id", record.OverlayID); err != nil {
			return err
		}
	}
	if record.ProofHash != "" {
		if err := ValidateHash("networking discovery proof hash", record.ProofHash); err != nil {
			return err
		}
		if record.ProofHeight == 0 {
			return errors.New("networking discovery proof height must be positive")
		}
		if record.ProofHeight > record.ExpiresHeight {
			return errors.New("networking discovery proof cannot outlive record")
		}
	}
	if record.ExpiresHeight == 0 || record.ExpiresHeight > record.Record.ExpiresHeight {
		return errors.New("networking discovery record expiry must be positive and not outlive node record")
	}
	if currentHeight > 0 && currentHeight > record.ExpiresHeight {
		return errors.New("networking discovery record is expired")
	}
	if len(record.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking discovery record signature must be %d bytes", ed25519.SignatureSize)
	}
	if err := validateDiscoveryRecordCompatibility(record); err != nil {
		return err
	}
	payload, err := DiscoveryRecordSigningPayload(record)
	if err != nil {
		return err
	}
	if !ed25519.Verify(record.Record.NodePubKey, payload, record.Signature) {
		return errors.New("networking discovery record signature verification failed")
	}
	return nil
}

func NewDiscoveryResponse(response DiscoveryResponse, sourcePrivateKey ed25519.PrivateKey, networkSalt []byte) (DiscoveryResponse, error) {
	if len(sourcePrivateKey) != ed25519.PrivateKeySize {
		return DiscoveryResponse{}, errors.New("networking discovery response private key must be ed25519")
	}
	response = NormalizeDiscoveryResponse(response)
	sourcePubKey, ok := sourcePrivateKey.Public().(ed25519.PublicKey)
	if !ok {
		return DiscoveryResponse{}, errors.New("networking discovery response public key must be ed25519")
	}
	if response.SourceNodeID == "" {
		response.SourceNodeID = ComputeNodeID(sourcePubKey, networkSalt)
	}
	if response.ExpiryHeight == 0 {
		response.ExpiryHeight = minDiscoveryRecordExpiry(response.MatchedRecords)
	}
	if response.ResultHash == "" {
		resultHash, err := ComputeDiscoveryResponseResultHash(response.MatchedRecords)
		if err != nil {
			return DiscoveryResponse{}, err
		}
		response.ResultHash = resultHash
	}
	if response.ResponseID == "" {
		response.ResponseID = ComputeDiscoveryResponseID(response)
	}
	payload, err := DiscoveryResponseSigningPayload(response)
	if err != nil {
		return DiscoveryResponse{}, err
	}
	response.SourceSignature = ed25519.Sign(sourcePrivateKey, payload)
	if err := response.Validate(sourcePubKey, networkSalt, response.GeneratedHeight); err != nil {
		return DiscoveryResponse{}, err
	}
	return response, nil
}

func BuildDiscoveryResponse(table DistributedRoutingTable, query DRTQuery, source NodeRecord, sourcePrivateKey ed25519.PrivateKey, networkSalt []byte, onChainProof DiscoveryOnChainProof, currentHeight uint64) (DiscoveryResponse, error) {
	records := table.findRecordsForQuery(query, currentHeight)
	response := DiscoveryResponse{
		QueryHash:		ComputeDRTQueryHash(query),
		MatchedRecords:		records,
		SignatureChain:		DiscoverySignatureChain(records),
		OnChainProof:		NormalizeDiscoveryOnChainProof(onChainProof),
		SourceNodeID:		source.NodeID,
		AdvisoryOnly:		onChainProof.ProofHash == "",
		GeneratedHeight:	currentHeight,
	}
	return NewDiscoveryResponse(response, sourcePrivateKey, networkSalt)
}

func NormalizeDiscoveryResponse(response DiscoveryResponse) DiscoveryResponse {
	response.ResponseID = normalizeHashText(response.ResponseID)
	response.QueryHash = normalizeHashText(response.QueryHash)
	response.MatchedRecords = cloneDiscoveryRecords(response.MatchedRecords)
	response.SignatureChain = cloneDiscoverySignatureChain(response.SignatureChain)
	response.OnChainProof = NormalizeDiscoveryOnChainProof(response.OnChainProof)
	response.SourceNodeID = normalizeHashText(response.SourceNodeID)
	response.SourceSignature = cloneBytes(response.SourceSignature)
	response.ResultHash = normalizeHashText(response.ResultHash)
	return response
}

func NormalizeDiscoverySignatureChainEntry(entry DiscoverySignatureChainEntry) DiscoverySignatureChainEntry {
	entry.NodeID = normalizeHashText(entry.NodeID)
	entry.PublicKey = cloneBytes(entry.PublicKey)
	entry.Signature = cloneBytes(entry.Signature)
	return entry
}

func NormalizeDiscoveryOnChainProof(proof DiscoveryOnChainProof) DiscoveryOnChainProof {
	proof.ProofHash = normalizeHashText(proof.ProofHash)
	proof.StateRoot = normalizeHashText(proof.StateRoot)
	return proof
}

func ComputeDiscoveryResponseResultHash(records []DiscoveryRecord) (string, error) {
	records = cloneDiscoveryRecords(records)
	parts := []string{"discovery-response-result"}
	for _, record := range records {
		if err := ValidateHash("networking discovery response record id", record.RecordID); err != nil {
			return "", err
		}
		parts = append(parts, record.RecordID)
	}
	return HashParts(parts...), nil
}

func ComputeDiscoveryResponseID(response DiscoveryResponse) string {
	response = NormalizeDiscoveryResponse(response)
	return HashParts(
		"discovery-response",
		response.QueryHash,
		response.ResultHash,
		fmt.Sprintf("%d", response.ExpiryHeight),
		response.SourceNodeID,
		response.OnChainProof.ProofHash,
		response.OnChainProof.StateRoot,
		fmt.Sprintf("%t", response.AdvisoryOnly),
		fmt.Sprintf("%d", response.GeneratedHeight),
	)
}

func ComputeDRTQueryHash(query DRTQuery) string {
	query = normalizeDRTQuery(query)
	return HashParts(
		"drt-query",
		string(query.ObjectType),
		query.ObjectID,
		query.OverlayID,
		query.ZoneID,
		query.ServiceID,
		fmt.Sprintf("%d", query.MinStakeWeight),
		fmt.Sprintf("%d", query.Limit),
		fmt.Sprintf("%d", query.CurrentHeight),
	)
}

func DiscoveryResponseSigningPayload(response DiscoveryResponse) ([]byte, error) {
	response = NormalizeDiscoveryResponse(response)
	response.SourceSignature = nil
	return json.Marshal(response)
}

func (response DiscoveryResponse) Validate(sourcePubKey ed25519.PublicKey, networkSalt []byte, currentHeight uint64) error {
	response = NormalizeDiscoveryResponse(response)
	if err := ValidateHash("networking discovery response id", response.ResponseID); err != nil {
		return err
	}
	if response.ResponseID != ComputeDiscoveryResponseID(response) {
		return errors.New("networking discovery response id mismatch")
	}
	if err := ValidateHash("networking discovery response query hash", response.QueryHash); err != nil {
		return err
	}
	if len(sourcePubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking discovery response source public key must be %d bytes", ed25519.PublicKeySize)
	}
	expectedSource := ComputeNodeID(sourcePubKey, networkSalt)
	if response.SourceNodeID != expectedSource {
		return errors.New("networking discovery response source node mismatch")
	}
	if response.GeneratedHeight == 0 || response.ExpiryHeight == 0 || response.GeneratedHeight > response.ExpiryHeight {
		return errors.New("networking discovery response heights are invalid")
	}
	if currentHeight > 0 && currentHeight > response.ExpiryHeight {
		return errors.New("networking discovery response is expired")
	}
	resultHash, err := ComputeDiscoveryResponseResultHash(response.MatchedRecords)
	if err != nil {
		return err
	}
	if response.ResultHash != resultHash {
		return errors.New("networking discovery response result hash mismatch")
	}
	if response.ExpiryHeight > minDiscoveryRecordExpiry(response.MatchedRecords) {
		return errors.New("networking discovery response expiry exceeds matched records")
	}
	for _, record := range response.MatchedRecords {
		if err := ValidateSignedDiscoveryRecord(record, networkSalt, currentHeight); err != nil {
			return err
		}
	}
	if err := ValidateDiscoverySignatureChain(response.SignatureChain, response.MatchedRecords); err != nil {
		return err
	}
	if response.OnChainProof.ProofHash != "" {
		if err := response.OnChainProof.Validate(response.ResultHash, response.ExpiryHeight, currentHeight); err != nil {
			return err
		}
	} else if !response.AdvisoryOnly {
		return errors.New("networking unproofed discovery response must be advisory")
	}
	if len(response.SourceSignature) != ed25519.SignatureSize {
		return fmt.Errorf("networking discovery response source signature must be %d bytes", ed25519.SignatureSize)
	}
	payload, err := DiscoveryResponseSigningPayload(response)
	if err != nil {
		return err
	}
	if !ed25519.Verify(sourcePubKey, payload, response.SourceSignature) {
		return errors.New("networking discovery response source signature verification failed")
	}
	return nil
}

func (proof DiscoveryOnChainProof) Validate(resultHash string, expiryHeight, currentHeight uint64) error {
	proof = NormalizeDiscoveryOnChainProof(proof)
	if err := ValidateHash("networking discovery response proof hash", proof.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking discovery response proof state root", proof.StateRoot); err != nil {
		return err
	}
	if proof.ProofHeight == 0 {
		return errors.New("networking discovery response proof height must be positive")
	}
	if expiryHeight > 0 && proof.ProofHeight > expiryHeight {
		return errors.New("networking discovery response proof cannot outlive response")
	}
	if currentHeight > 0 && proof.ProofHeight > currentHeight {
		return errors.New("networking discovery response proof height is in the future")
	}
	expected := ComputeDiscoveryOnChainProofHash(resultHash, proof.StateRoot, proof.ProofHeight)
	if proof.ProofHash != expected {
		return errors.New("networking discovery response proof mismatch")
	}
	return nil
}

func DiscoverySignatureChain(records []DiscoveryRecord) []DiscoverySignatureChainEntry {
	records = cloneDiscoveryRecords(records)
	chain := make([]DiscoverySignatureChainEntry, len(records))
	for i, record := range records {
		chain[i] = DiscoverySignatureChainEntry{
			NodeID:		record.OwnerNodeID,
			PublicKey:	cloneBytes(record.Record.NodePubKey),
			Signature:	cloneBytes(record.Signature),
		}
	}
	return chain
}

func ValidateDiscoverySignatureChain(chain []DiscoverySignatureChainEntry, records []DiscoveryRecord) error {
	records = cloneDiscoveryRecords(records)
	if len(chain) != len(records) {
		return errors.New("networking discovery signature chain length mismatch")
	}
	normalized := cloneDiscoverySignatureChain(chain)
	for i, record := range records {
		entry := normalized[i]
		if entry.NodeID != record.OwnerNodeID {
			return errors.New("networking discovery signature chain owner mismatch")
		}
		if string(entry.PublicKey) != string(record.Record.NodePubKey) {
			return errors.New("networking discovery signature chain public key mismatch")
		}
		if string(entry.Signature) != string(record.Signature) {
			return errors.New("networking discovery signature chain signature mismatch")
		}
	}
	return nil
}

func ComputeDRTObjectID(ad DRTAdvertisement) string {
	ad = NormalizeDRTAdvertisement(ad)
	return HashParts(
		"drt-object",
		string(ad.ObjectType),
		ad.Discovery.Record.NodeID,
		ad.OverlayID,
		ad.ZoneID,
		ad.ServiceID,
		ad.EndpointHash,
	)
}

func ComputeDRTAdvertisementID(ad DRTAdvertisement) string {
	ad = NormalizeDRTAdvertisement(ad)
	return HashParts(
		"drt-advertisement",
		string(ad.ObjectType),
		ad.ObjectID,
		ad.Discovery.Record.NodeID,
		ad.OverlayID,
		ad.ZoneID,
		ad.ServiceID,
		ad.EndpointHash,
		fmt.Sprintf("%d", ad.StakeWeight),
		fmt.Sprintf("%d", ad.PeerScoreBps),
		fmt.Sprintf("%d", ad.LeaseStartHeight),
		fmt.Sprintf("%d", ad.LeaseExpireHeight),
		ad.Discovery.ProofHash,
	)
}

func ComputeDRTIndexRoot(advertisements []DRTAdvertisement) (string, error) {
	if len(advertisements) == 0 {
		return HashParts("drt-index-root", "empty"), nil
	}
	normalized := cloneDRTAdvertisements(advertisements)
	sortDRTAdvertisements(normalized)
	parts := []string{"drt-index-root"}
	for _, ad := range normalized {
		if err := ad.Validate(nil, 0); err != nil {
			return "", err
		}
		parts = append(parts, ad.AdvertisementID)
	}
	return HashParts(parts...), nil
}

func (ad DRTAdvertisement) Validate(networkSalt []byte, currentHeight uint64) error {
	ad = NormalizeDRTAdvertisement(ad)
	if err := ValidateHash("networking DRT advertisement id", ad.AdvertisementID); err != nil {
		return err
	}
	if ad.AdvertisementID != ComputeDRTAdvertisementID(ad) {
		return errors.New("networking DRT advertisement id mismatch")
	}
	if !IsDRTObjectType(ad.ObjectType) {
		return fmt.Errorf("unknown networking DRT object type %q", ad.ObjectType)
	}
	if err := ValidateHash("networking DRT object id", ad.ObjectID); err != nil {
		return err
	}
	if ad.ObjectID != ComputeDRTObjectID(ad) {
		return errors.New("networking DRT object id mismatch")
	}
	if len(networkSalt) > 0 {
		if err := ad.Discovery.Validate(networkSalt, currentHeight); err != nil {
			return err
		}
	} else if err := ad.Discovery.Record.ValidateBasic(); err != nil {
		return err
	}
	if ad.LeaseStartHeight == 0 || ad.LeaseExpireHeight == 0 {
		return errors.New("networking DRT lease heights must be positive")
	}
	if ad.LeaseStartHeight > ad.LeaseExpireHeight {
		return errors.New("networking DRT lease start cannot exceed expiry")
	}
	if ad.LeaseExpireHeight > ad.Discovery.Record.ExpiresHeight {
		return errors.New("networking DRT lease cannot outlive node record")
	}
	if currentHeight > 0 && currentHeight > ad.LeaseExpireHeight {
		return errors.New("networking DRT advertisement is expired")
	}
	if ad.PeerScoreBps > BasisPoints {
		return fmt.Errorf("networking DRT peer score must be <= %d bps", BasisPoints)
	}
	if ad.OverlayID != "" {
		if err := ValidateHash("networking DRT overlay id", ad.OverlayID); err != nil {
			return err
		}
	}
	if ad.EndpointHash != "" {
		if err := ValidateHash("networking DRT endpoint hash", ad.EndpointHash); err != nil {
			return err
		}
	}
	return validateDRTObjectCompatibility(ad)
}

func (table DistributedRoutingTable) Add(ad DRTAdvertisement, networkSalt []byte, currentHeight uint64) (DistributedRoutingTable, error) {
	ad, err := NewDRTAdvertisement(ad)
	if err != nil {
		return DistributedRoutingTable{}, err
	}
	if err := ad.Validate(networkSalt, currentHeight); err != nil {
		return DistributedRoutingTable{}, err
	}
	next := table.Clone()
	replaced := false
	key := drtAdvertisementKey(ad)
	for i, existing := range next.Advertisements {
		if drtAdvertisementKey(existing) == key {
			next.Advertisements[i] = ad
			replaced = true
			break
		}
	}
	if !replaced {
		next.Advertisements = append(next.Advertisements, ad)
	}
	sortDRTAdvertisements(next.Advertisements)
	return next, next.Validate(networkSalt, currentHeight)
}

func (table DistributedRoutingTable) Query(query DRTQuery) []DRTAdvertisement {
	query = normalizeDRTQuery(query)
	limit := query.Limit
	if limit == 0 {
		limit = DefaultDRTQueryLimit
	}
	if limit > MaxDRTQueryLimit {
		limit = MaxDRTQueryLimit
	}
	out := make([]DRTAdvertisement, 0)
	for _, ad := range table.Advertisements {
		ad = NormalizeDRTAdvertisement(ad)
		if query.CurrentHeight > 0 && query.CurrentHeight > ad.LeaseExpireHeight {
			continue
		}
		if query.ObjectType != "" && ad.ObjectType != query.ObjectType {
			continue
		}
		if query.ObjectID != "" && ad.ObjectID != query.ObjectID {
			continue
		}
		if query.OverlayID != "" && ad.OverlayID != query.OverlayID {
			continue
		}
		if query.ZoneID != "" && ad.ZoneID != query.ZoneID {
			continue
		}
		if query.ServiceID != "" && ad.ServiceID != query.ServiceID {
			continue
		}
		if ad.StakeWeight < query.MinStakeWeight {
			continue
		}
		out = append(out, ad)
	}
	sortDRTAdvertisementsByRank(out)
	if uint32(len(out)) > limit {
		out = out[:limit]
	}
	return out
}

func (table DistributedRoutingTable) Store(record DiscoveryRecord, networkSalt []byte, currentHeight uint64) (DistributedRoutingTable, error) {
	record = NormalizeDiscoveryRecord(record)
	if err := record.Validate(networkSalt, currentHeight); err != nil {
		return DistributedRoutingTable{}, err
	}
	if table.isRevoked(record.RecordID) {
		return DistributedRoutingTable{}, errors.New("networking discovery record is revoked")
	}
	next := table.Clone()
	replaced := false
	for i, existing := range next.Records {
		if discoveryRecordKey(existing) == discoveryRecordKey(record) {
			next.Records[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		next.Records = append(next.Records, record)
	}
	sortDiscoveryRecords(next.Records)
	return next, next.Validate(networkSalt, currentHeight)
}

func (table DistributedRoutingTable) UpdateLease(record DiscoveryRecord, networkSalt []byte, currentHeight uint64) (DistributedRoutingTable, error) {
	record = NormalizeDiscoveryRecord(record)
	if err := record.Validate(networkSalt, currentHeight); err != nil {
		return DistributedRoutingTable{}, err
	}
	found := false
	next := table.Clone()
	records := make([]DiscoveryRecord, 0, len(next.Records)+1)
	for _, existing := range table.Records {
		existing = NormalizeDiscoveryRecord(existing)
		if existing.RecordType == record.RecordType &&
			existing.OwnerNodeID == record.OwnerNodeID &&
			existing.TargetID == record.TargetID &&
			existing.AdvertisementHash == record.AdvertisementHash {
			found = true
			if record.ExpiresHeight <= existing.ExpiresHeight {
				return DistributedRoutingTable{}, errors.New("networking discovery lease update must extend expiry")
			}
			continue
		}
		records = append(records, existing)
	}
	if !found {
		return DistributedRoutingTable{}, errors.New("networking discovery lease update target not found")
	}
	records = append(records, record)
	next.Records = records
	sortDiscoveryRecords(next.Records)
	return next, next.Validate(networkSalt, currentHeight)
}

func (table DistributedRoutingTable) Revoke(revocation DiscoveryRevocation, networkSalt []byte, currentHeight uint64) (DistributedRoutingTable, error) {
	revocation = NormalizeDiscoveryRevocation(revocation)
	record, found := table.discoveryRecordByID(revocation.RecordID)
	if !found {
		return DistributedRoutingTable{}, errors.New("networking discovery revoke target not found")
	}
	if err := revocation.Validate(record, networkSalt, currentHeight); err != nil {
		return DistributedRoutingTable{}, err
	}
	next := table.Clone()
	remaining := make([]DiscoveryRecord, 0, len(next.Records))
	for _, existing := range next.Records {
		if NormalizeDiscoveryRecord(existing).RecordID == revocation.RecordID {
			continue
		}
		remaining = append(remaining, existing)
	}
	next.Records = remaining
	next.Revocations = append(next.Revocations, revocation)
	sortDiscoveryRecords(next.Records)
	sortDiscoveryRevocations(next.Revocations)
	return next, next.Validate(networkSalt, currentHeight)
}

func (table DistributedRoutingTable) FindNode(nodeID string, currentHeight uint64) []DiscoveryRecord {
	nodeID = normalizeHashText(nodeID)
	return table.findRecords(func(record DiscoveryRecord) bool {
		return record.RecordType == DRTObjectNode && (record.TargetID == nodeID || record.OwnerNodeID == nodeID)
	}, currentHeight)
}

func (table DistributedRoutingTable) FindService(serviceID string, currentHeight uint64) []DiscoveryRecord {
	serviceID = strings.TrimSpace(serviceID)
	return table.findRecords(func(record DiscoveryRecord) bool {
		return record.RecordType == DRTObjectServiceEndpoint && record.ServiceID == serviceID
	}, currentHeight)
}

func (table DistributedRoutingTable) FindZone(zoneID string, currentHeight uint64) []DiscoveryRecord {
	zoneID = strings.TrimSpace(zoneID)
	return table.findRecords(func(record DiscoveryRecord) bool {
		return record.RecordType == DRTObjectExecutionZone && record.ZoneID == zoneID
	}, currentHeight)
}

func (table DistributedRoutingTable) FindEndpoint(advertisementHash string, currentHeight uint64) []DiscoveryRecord {
	advertisementHash = normalizeHashText(advertisementHash)
	return table.findRecords(func(record DiscoveryRecord) bool {
		return record.AdvertisementHash == advertisementHash
	}, currentHeight)
}

func (table DistributedRoutingTable) FindStorageProvider(currentHeight uint64) []DiscoveryRecord {
	return table.findRecords(func(record DiscoveryRecord) bool {
		return record.RecordType == DRTObjectStorageProvider
	}, currentHeight)
}

func (table DistributedRoutingTable) findRecordsForQuery(query DRTQuery, currentHeight uint64) []DiscoveryRecord {
	query = normalizeDRTQuery(query)
	return table.findRecords(func(record DiscoveryRecord) bool {
		if query.ObjectType != "" && record.RecordType != query.ObjectType {
			return false
		}
		if query.ObjectID != "" && record.TargetID != query.ObjectID {
			return false
		}
		if query.OverlayID != "" && record.OverlayID != query.OverlayID {
			return false
		}
		if query.ZoneID != "" && record.ZoneID != query.ZoneID {
			return false
		}
		if query.ServiceID != "" && record.ServiceID != query.ServiceID {
			return false
		}
		return true
	}, currentHeight)
}

func (table DistributedRoutingTable) Prune(currentHeight uint64) DistributedRoutingTable {
	next := DistributedRoutingTable{}
	for _, ad := range table.Advertisements {
		ad = NormalizeDRTAdvertisement(ad)
		if currentHeight > 0 && currentHeight > ad.LeaseExpireHeight {
			continue
		}
		next.Advertisements = append(next.Advertisements, ad)
	}
	for _, record := range table.Records {
		record = NormalizeDiscoveryRecord(record)
		if currentHeight > 0 && currentHeight > record.ExpiresHeight {
			continue
		}
		if table.isRevoked(record.RecordID) {
			continue
		}
		next.Records = append(next.Records, record)
	}
	next.Revocations = append([]DiscoveryRevocation(nil), table.Revocations...)
	sortDRTAdvertisements(next.Advertisements)
	sortDiscoveryRecords(next.Records)
	sortDiscoveryRevocations(next.Revocations)
	return next
}

func (table DistributedRoutingTable) Buckets(localNodeID string, objectType DRTObjectType, bucketCount uint32, currentHeight uint64) ([]DRTBucket, error) {
	localNodeID = normalizeHashText(localNodeID)
	if err := ValidateHash("networking DRT local node id", localNodeID); err != nil {
		return nil, err
	}
	objectType = DRTObjectType(strings.ToLower(strings.TrimSpace(string(objectType))))
	if objectType != "" && !IsDRTObjectType(objectType) {
		return nil, fmt.Errorf("unknown networking DRT object type %q", objectType)
	}
	if bucketCount == 0 {
		return nil, errors.New("networking DRT bucket count must be positive")
	}
	buckets := make([]DRTBucket, bucketCount)
	for i := range buckets {
		buckets[i].BucketID = uint32(i)
	}
	for _, ad := range table.Advertisements {
		ad = NormalizeDRTAdvertisement(ad)
		if currentHeight > 0 && currentHeight > ad.LeaseExpireHeight {
			continue
		}
		if objectType != "" && ad.ObjectType != objectType {
			continue
		}
		bucketID := drtBucketID(localNodeID, ad.Discovery.Record.NodeID, bucketCount)
		buckets[bucketID].Advertisements = append(buckets[bucketID].Advertisements, ad)
	}
	for i := range buckets {
		sortDRTAdvertisementsByRank(buckets[i].Advertisements)
	}
	return buckets, nil
}

func (table DistributedRoutingTable) Validate(networkSalt []byte, currentHeight uint64) error {
	if len(table.Advertisements) > MaxDRTAdvertisements {
		return fmt.Errorf("networking DRT advertisements must be <= %d", MaxDRTAdvertisements)
	}
	seen := make(map[string]struct{}, len(table.Advertisements))
	var previous string
	for i, ad := range table.Advertisements {
		ad = NormalizeDRTAdvertisement(ad)
		if err := ad.Validate(networkSalt, currentHeight); err != nil {
			return err
		}
		key := drtAdvertisementKey(ad)
		if _, found := seen[key]; found {
			return errors.New("networking duplicate DRT advertisement")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("networking DRT advertisements must be sorted canonically")
		}
		previous = key
	}
	seenRecords := make(map[string]struct{}, len(table.Records))
	previous = ""
	for i, record := range table.Records {
		record = NormalizeDiscoveryRecord(record)
		if err := record.Validate(networkSalt, currentHeight); err != nil {
			return err
		}
		key := discoveryRecordKey(record)
		if _, found := seenRecords[key]; found {
			return errors.New("networking duplicate discovery record")
		}
		if table.isRevoked(record.RecordID) {
			return errors.New("networking active discovery record is revoked")
		}
		seenRecords[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("networking discovery records must be sorted canonically")
		}
		previous = key
	}
	seenRevocations := make(map[string]struct{}, len(table.Revocations))
	previous = ""
	for i, revocation := range table.Revocations {
		revocation = NormalizeDiscoveryRevocation(revocation)
		if err := ValidateHash("networking discovery revocation id", revocation.RevocationID); err != nil {
			return err
		}
		key := discoveryRevocationKey(revocation)
		if _, found := seenRevocations[key]; found {
			return errors.New("networking duplicate discovery revocation")
		}
		seenRevocations[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("networking discovery revocations must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func (table DistributedRoutingTable) Clone() DistributedRoutingTable {
	return DistributedRoutingTable{
		Advertisements:	cloneDRTAdvertisements(table.Advertisements),
		Records:	cloneDiscoveryRecords(table.Records),
		Revocations:	cloneDiscoveryRevocations(table.Revocations),
	}
}

func IsDRTObjectType(objectType DRTObjectType) bool {
	switch objectType {
	case DRTObjectNode,
		DRTObjectExecutionZone,
		DRTObjectServiceEndpoint,
		DRTObjectRPCEndpoint,
		DRTObjectStorageProvider,
		DRTObjectRoutingEntryPoint,
		DRTObjectOverlayMembershipRecord,
		DRTObjectStreamProvider:
		return true
	default:
		return false
	}
}

func normalizeDRTQuery(query DRTQuery) DRTQuery {
	query.ObjectType = DRTObjectType(strings.ToLower(strings.TrimSpace(string(query.ObjectType))))
	query.ObjectID = normalizeHashText(query.ObjectID)
	query.OverlayID = normalizeHashText(query.OverlayID)
	query.ZoneID = strings.TrimSpace(query.ZoneID)
	query.ServiceID = strings.TrimSpace(query.ServiceID)
	return query
}

func validateDRTObjectCompatibility(ad DRTAdvertisement) error {
	record := ad.Discovery.Record
	switch ad.ObjectType {
	case DRTObjectNode:
		if ad.ObjectID != ComputeDRTObjectID(ad) {
			return errors.New("networking DRT node object id mismatch")
		}
	case DRTObjectExecutionZone:
		if ad.ZoneID == "" {
			return errors.New("networking DRT execution zone requires zone id")
		}
		if err := validateIdentifierSet("zone", []string{ad.ZoneID}, MaxZoneIDBytes); err != nil {
			return err
		}
		if !hasRole(record.Roles, NodeRoleZoneExecution) && !containsString(record.ZonesSupported, ad.ZoneID) {
			return errors.New("networking DRT execution zone requires zone execution role or supported zone")
		}
	case DRTObjectServiceEndpoint:
		if ad.ServiceID == "" || ad.EndpointHash == "" {
			return errors.New("networking DRT service endpoint requires service id and endpoint hash")
		}
		if err := validateIdentifierSet("service", []string{ad.ServiceID}, MaxServiceIDBytes); err != nil {
			return err
		}
		if !hasRole(record.Roles, NodeRoleService) || !containsString(record.ServicesSupported, ad.ServiceID) {
			return errors.New("networking DRT service endpoint requires advertised service role")
		}
	case DRTObjectRPCEndpoint:
		if ad.EndpointHash == "" {
			return errors.New("networking DRT RPC endpoint requires endpoint hash")
		}
		if !hasRole(record.Roles, NodeRoleFull) && !hasRole(record.Roles, NodeRoleArchive) && !hasRole(record.Roles, NodeRoleLightGateway) {
			return errors.New("networking DRT RPC endpoint requires full, archive, or light gateway role")
		}
	case DRTObjectStorageProvider:
		if ad.EndpointHash == "" {
			return errors.New("networking DRT storage provider requires endpoint hash")
		}
		if !hasRole(record.Roles, NodeRoleStorageProvider) {
			return errors.New("networking DRT storage provider requires storage provider role")
		}
	case DRTObjectRoutingEntryPoint:
		if ad.EndpointHash == "" {
			return errors.New("networking DRT routing entry point requires endpoint hash")
		}
		if !hasRole(record.Roles, NodeRoleRouting) {
			return errors.New("networking DRT routing entry point requires routing role")
		}
	case DRTObjectOverlayMembershipRecord:
		if ad.OverlayID == "" {
			return errors.New("networking DRT overlay membership record requires overlay id")
		}
	case DRTObjectStreamProvider:
		if ad.EndpointHash == "" {
			return errors.New("networking DRT stream provider requires endpoint hash")
		}
		if !hasRole(record.Roles, NodeRoleStateSync) && !hasRole(record.Roles, NodeRoleStorageProvider) && !hasRole(record.Roles, NodeRoleFull) {
			return errors.New("networking DRT stream provider requires state sync, storage, or full node role")
		}
	}
	return nil
}

func validateDiscoveryRecordCompatibility(record DiscoveryRecord) error {
	ad := DRTAdvertisement{
		ObjectType:		record.RecordType,
		ObjectID:		record.TargetID,
		Discovery:		DiscoveryRecord{Record: record.Record},
		OverlayID:		record.OverlayID,
		ZoneID:			record.ZoneID,
		ServiceID:		record.ServiceID,
		EndpointHash:		record.AdvertisementHash,
		StakeWeight:		1,
		PeerScoreBps:		0,
		LeaseStartHeight:	1,
		LeaseExpireHeight:	record.ExpiresHeight,
	}
	switch record.RecordType {
	case DRTObjectNode, DRTObjectOverlayMembershipRecord:
		return nil
	default:
		return validateDRTObjectCompatibility(ad)
	}
}

func NewDiscoveryRevocation(record DiscoveryRecord, privateKey ed25519.PrivateKey, revokedHeight uint64) (DiscoveryRevocation, error) {
	record = NormalizeDiscoveryRecord(record)
	if len(privateKey) != ed25519.PrivateKeySize {
		return DiscoveryRevocation{}, errors.New("networking discovery revocation private key must be ed25519")
	}
	pubKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok || string(pubKey) != string(record.Record.NodePubKey) {
		return DiscoveryRevocation{}, errors.New("networking discovery revocation signer must own node record")
	}
	revocation := NormalizeDiscoveryRevocation(DiscoveryRevocation{
		RecordID:		record.RecordID,
		OwnerNodeID:		record.OwnerNodeID,
		AdvertisementHash:	record.AdvertisementHash,
		RevokedHeight:		revokedHeight,
	})
	revocation.RevocationID = ComputeDiscoveryRevocationID(revocation)
	payload, err := DiscoveryRevocationSigningPayload(revocation)
	if err != nil {
		return DiscoveryRevocation{}, err
	}
	revocation.Signature = ed25519.Sign(privateKey, payload)
	if err := revocation.Validate(record, nil, revokedHeight); err != nil {
		return DiscoveryRevocation{}, err
	}
	return revocation, nil
}

func NormalizeDiscoveryRevocation(revocation DiscoveryRevocation) DiscoveryRevocation {
	revocation.RevocationID = normalizeHashText(revocation.RevocationID)
	revocation.RecordID = normalizeHashText(revocation.RecordID)
	revocation.OwnerNodeID = normalizeHashText(revocation.OwnerNodeID)
	revocation.AdvertisementHash = normalizeHashText(revocation.AdvertisementHash)
	revocation.Signature = cloneBytes(revocation.Signature)
	return revocation
}

func ComputeDiscoveryRevocationID(revocation DiscoveryRevocation) string {
	revocation = NormalizeDiscoveryRevocation(revocation)
	return HashParts(
		"discovery-revocation",
		revocation.RecordID,
		revocation.OwnerNodeID,
		revocation.AdvertisementHash,
		fmt.Sprintf("%d", revocation.RevokedHeight),
	)
}

func DiscoveryRevocationSigningPayload(revocation DiscoveryRevocation) ([]byte, error) {
	revocation = NormalizeDiscoveryRevocation(revocation)
	revocation.Signature = nil
	return json.Marshal(revocation)
}

func (r DiscoveryRevocation) Validate(record DiscoveryRecord, networkSalt []byte, currentHeight uint64) error {
	revocation := NormalizeDiscoveryRevocation(r)
	record = NormalizeDiscoveryRecord(record)
	if err := ValidateSignedDiscoveryRecord(record, networkSalt, 0); err != nil {
		return err
	}
	if revocation.RevocationID != ComputeDiscoveryRevocationID(revocation) {
		return errors.New("networking discovery revocation id mismatch")
	}
	if revocation.RecordID != record.RecordID || revocation.OwnerNodeID != record.OwnerNodeID || revocation.AdvertisementHash != record.AdvertisementHash {
		return errors.New("networking discovery revocation target mismatch")
	}
	if revocation.RevokedHeight == 0 {
		return errors.New("networking discovery revocation height must be positive")
	}
	if currentHeight > 0 && revocation.RevokedHeight > currentHeight {
		return errors.New("networking discovery revocation height is in the future")
	}
	if len(revocation.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking discovery revocation signature must be %d bytes", ed25519.SignatureSize)
	}
	payload, err := DiscoveryRevocationSigningPayload(revocation)
	if err != nil {
		return err
	}
	if !ed25519.Verify(record.Record.NodePubKey, payload, revocation.Signature) {
		return errors.New("networking discovery revocation signature verification failed")
	}
	return nil
}

func cloneDRTAdvertisements(advertisements []DRTAdvertisement) []DRTAdvertisement {
	out := make([]DRTAdvertisement, len(advertisements))
	for i, ad := range advertisements {
		out[i] = NormalizeDRTAdvertisement(ad)
	}
	sortDRTAdvertisements(out)
	return out
}

func cloneDiscoveryRecords(records []DiscoveryRecord) []DiscoveryRecord {
	out := make([]DiscoveryRecord, len(records))
	for i, record := range records {
		out[i] = NormalizeDiscoveryRecord(record)
	}
	sortDiscoveryRecords(out)
	return out
}

func cloneDiscoveryRevocations(revocations []DiscoveryRevocation) []DiscoveryRevocation {
	out := make([]DiscoveryRevocation, len(revocations))
	for i, revocation := range revocations {
		out[i] = NormalizeDiscoveryRevocation(revocation)
	}
	sortDiscoveryRevocations(out)
	return out
}

func cloneDiscoverySignatureChain(chain []DiscoverySignatureChainEntry) []DiscoverySignatureChainEntry {
	out := make([]DiscoverySignatureChainEntry, len(chain))
	for i, entry := range chain {
		out[i] = NormalizeDiscoverySignatureChainEntry(entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].NodeID != out[j].NodeID {
			return out[i].NodeID < out[j].NodeID
		}
		return string(out[i].Signature) < string(out[j].Signature)
	})
	return out
}

func sortDRTAdvertisements(advertisements []DRTAdvertisement) {
	sort.SliceStable(advertisements, func(i, j int) bool {
		return drtAdvertisementKey(advertisements[i]) < drtAdvertisementKey(advertisements[j])
	})
}

func sortDRTAdvertisementsByRank(advertisements []DRTAdvertisement) {
	sort.SliceStable(advertisements, func(i, j int) bool {
		left := NormalizeDRTAdvertisement(advertisements[i])
		right := NormalizeDRTAdvertisement(advertisements[j])
		if left.StakeWeight != right.StakeWeight {
			return left.StakeWeight > right.StakeWeight
		}
		if left.PeerScoreBps != right.PeerScoreBps {
			return left.PeerScoreBps > right.PeerScoreBps
		}
		if left.LeaseExpireHeight != right.LeaseExpireHeight {
			return left.LeaseExpireHeight > right.LeaseExpireHeight
		}
		return left.AdvertisementID < right.AdvertisementID
	})
}

func drtAdvertisementKey(ad DRTAdvertisement) string {
	ad = NormalizeDRTAdvertisement(ad)
	return string(ad.ObjectType) + "/" + ad.ObjectID + "/" + ad.Discovery.Record.NodeID
}

func discoveryRecordKey(record DiscoveryRecord) string {
	record = NormalizeDiscoveryRecord(record)
	return string(record.RecordType) + "/" + record.OwnerNodeID + "/" + record.TargetID + "/" + record.RecordID
}

func discoveryRevocationKey(revocation DiscoveryRevocation) string {
	revocation = NormalizeDiscoveryRevocation(revocation)
	return revocation.RecordID + "/" + revocation.RevocationID
}

func sortDiscoveryRecords(records []DiscoveryRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return discoveryRecordKey(records[i]) < discoveryRecordKey(records[j])
	})
}

func sortDiscoveryRevocations(revocations []DiscoveryRevocation) {
	sort.SliceStable(revocations, func(i, j int) bool {
		return discoveryRevocationKey(revocations[i]) < discoveryRevocationKey(revocations[j])
	})
}

func (table DistributedRoutingTable) isRevoked(recordID string) bool {
	recordID = normalizeHashText(recordID)
	for _, revocation := range table.Revocations {
		if NormalizeDiscoveryRevocation(revocation).RecordID == recordID {
			return true
		}
	}
	return false
}

func (table DistributedRoutingTable) discoveryRecordByID(recordID string) (DiscoveryRecord, bool) {
	recordID = normalizeHashText(recordID)
	for _, record := range table.Records {
		record = NormalizeDiscoveryRecord(record)
		if record.RecordID == recordID {
			return record, true
		}
	}
	return DiscoveryRecord{}, false
}

func (table DistributedRoutingTable) findRecords(match func(DiscoveryRecord) bool, currentHeight uint64) []DiscoveryRecord {
	out := make([]DiscoveryRecord, 0)
	for _, record := range table.Records {
		record = NormalizeDiscoveryRecord(record)
		if currentHeight > 0 && currentHeight > record.ExpiresHeight {
			continue
		}
		if table.isRevoked(record.RecordID) {
			continue
		}
		if match(record) {
			out = append(out, record)
		}
	}
	sortDiscoveryRecords(out)
	return out
}

func minDiscoveryRecordExpiry(records []DiscoveryRecord) uint64 {
	if len(records) == 0 {
		return 1
	}
	min := uint64(0)
	for _, record := range records {
		record = NormalizeDiscoveryRecord(record)
		if record.ExpiresHeight == 0 {
			continue
		}
		if min == 0 || record.ExpiresHeight < min {
			min = record.ExpiresHeight
		}
	}
	if min == 0 {
		return 1
	}
	return min
}

func ComputeDiscoveryOnChainProofHash(resultHash, stateRoot string, proofHeight uint64) string {
	return HashParts("discovery-on-chain-proof", normalizeHashText(resultHash), normalizeHashText(stateRoot), fmt.Sprintf("%d", proofHeight))
}

func drtBucketID(localNodeID, remoteNodeID string, bucketCount uint32) uint32 {
	return uint32(hashBytes("aetra-drt-bucket-v1", []byte(localNodeID+"/"+remoteNodeID))[0]) % bucketCount
}
