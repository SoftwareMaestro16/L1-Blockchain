package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxDRTAdvertisements = 4096
	DefaultDRTQueryLimit = uint32(32)
	MaxDRTQueryLimit     = uint32(256)
)

type DRTObjectType string

const (
	DRTObjectNode                    DRTObjectType = "node"
	DRTObjectExecutionZone           DRTObjectType = "execution_zone"
	DRTObjectServiceEndpoint         DRTObjectType = "service_endpoint"
	DRTObjectRPCEndpoint             DRTObjectType = "rpc_endpoint"
	DRTObjectStorageProvider         DRTObjectType = "storage_provider"
	DRTObjectRoutingEntryPoint       DRTObjectType = "routing_entry_point"
	DRTObjectOverlayMembershipRecord DRTObjectType = "overlay_membership_record"
	DRTObjectStreamProvider          DRTObjectType = "stream_provider"
)

type DRTAdvertisement struct {
	AdvertisementID   string
	ObjectType        DRTObjectType
	ObjectID          string
	Discovery         DiscoveryRecord
	OverlayID         string
	ZoneID            string
	ServiceID         string
	EndpointHash      string
	StakeWeight       uint64
	PeerScoreBps      uint32
	LeaseStartHeight  uint64
	LeaseExpireHeight uint64
}

type DRTQuery struct {
	ObjectType     DRTObjectType
	ObjectID       string
	OverlayID      string
	ZoneID         string
	ServiceID      string
	MinStakeWeight uint64
	Limit          uint32
	CurrentHeight  uint64
}

type DRTBucket struct {
	BucketID       uint32
	Advertisements []DRTAdvertisement
}

type DistributedRoutingTable struct {
	Advertisements []DRTAdvertisement
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
	ad.Discovery.Record = NormalizeNodeRecord(ad.Discovery.Record)
	ad.Discovery.ProofHash = normalizeHashText(ad.Discovery.ProofHash)
	ad.OverlayID = normalizeHashText(ad.OverlayID)
	ad.ZoneID = strings.TrimSpace(ad.ZoneID)
	ad.ServiceID = strings.TrimSpace(ad.ServiceID)
	ad.EndpointHash = normalizeHashText(ad.EndpointHash)
	return ad
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

func (table DistributedRoutingTable) Prune(currentHeight uint64) DistributedRoutingTable {
	next := DistributedRoutingTable{}
	for _, ad := range table.Advertisements {
		ad = NormalizeDRTAdvertisement(ad)
		if currentHeight > 0 && currentHeight > ad.LeaseExpireHeight {
			continue
		}
		next.Advertisements = append(next.Advertisements, ad)
	}
	sortDRTAdvertisements(next.Advertisements)
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
	return nil
}

func (table DistributedRoutingTable) Clone() DistributedRoutingTable {
	return DistributedRoutingTable{Advertisements: cloneDRTAdvertisements(table.Advertisements)}
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

func cloneDRTAdvertisements(advertisements []DRTAdvertisement) []DRTAdvertisement {
	out := make([]DRTAdvertisement, len(advertisements))
	for i, ad := range advertisements {
		out[i] = NormalizeDRTAdvertisement(ad)
	}
	sortDRTAdvertisements(out)
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

func drtBucketID(localNodeID, remoteNodeID string, bucketCount uint32) uint32 {
	return uint32(hashBytes("aetheris-drt-bucket-v1", []byte(localNodeID+"/"+remoteNodeID))[0]) % bucketCount
}
