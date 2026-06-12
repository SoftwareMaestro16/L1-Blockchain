package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxDNLEntries		= 4096
	MaxDNLRoutes		= 8192
	MaxDNLCache		= 4096
	DefaultDNLQueryLimit	= uint32(32)
	MaxDNLQueryLimit	= uint32(256)
)

type DNLServiceDiscoveryEntry struct {
	EntryID		string
	ServiceID	string
	ZoneID		string
	InterfaceHash	string
	EndpointHash	string
	RouteID		string
	ExpiryHeight	uint64
	ProofHash	string
	EntryHash	string
}

type DNLRoutingTableEntry struct {
	RouteID		string
	ZoneID		string
	ServiceID	string
	NextHopNodeID	string
	OverlayID	string
	Priority	uint32
	WeightBps	uint32
	ExpiryHeight	uint64
	EntryHash	string
}

type DNLCacheEntry struct {
	CacheKey	string
	QueryHash	string
	ResponseHash	string
	ExpiryHeight	uint64
	ProofHash	string
	EntryHash	string
}

type DNLState struct {
	ServiceEntries	[]DNLServiceDiscoveryEntry
	RoutingTable	[]DNLRoutingTableEntry
	CacheEntries	[]DNLCacheEntry
	Height		uint64
	ServicesRoot	string
	RoutingRoot	string
	LookupRoot	string
	CacheRoot	string
	StateRoot	string
}

type DNLQuery struct {
	ServiceID	string
	ZoneID		string
	InterfaceHash	string
	CurrentHeight	uint64
	Limit		uint32
	RequireProof	bool
}

type DNLProof struct {
	Key		string
	ValueHash	string
	StateRoot	string
	Height		uint64
	Path		[]string
	ProofHash	string
}

type DNLDiscoveryResponse struct {
	QueryHash	string
	Entries		[]DNLServiceDiscoveryEntry
	Routes		[]DNLRoutingTableEntry
	Proof		DNLProof
	ExpiryHeight	uint64
	ResponseHash	string
}

type DNLAdvisoryObservation struct {
	ObservedNodeID	string
	ServiceID	string
	ZoneID		string
	EndpointHash	string
	ObservedHeight	uint64
	ObservationHash	string
}

type DNLStateEntry struct {
	Key		string
	ValueHash	string
	EntryHash	string
}

func NewDNLServiceDiscoveryEntry(entry DNLServiceDiscoveryEntry) (DNLServiceDiscoveryEntry, error) {
	entry = NormalizeDNLServiceDiscoveryEntry(entry)
	if entry.EntryID == "" {
		entry.EntryID = ComputeDNLServiceEntryID(entry)
	}
	if entry.EntryHash == "" {
		entry.EntryHash = ComputeDNLServiceEntryHash(entry)
	}
	return entry, entry.Validate()
}

func NewDNLRoutingTableEntry(entry DNLRoutingTableEntry) (DNLRoutingTableEntry, error) {
	entry = NormalizeDNLRoutingTableEntry(entry)
	if entry.RouteID == "" {
		entry.RouteID = ComputeDNLRouteID(entry)
	}
	if entry.EntryHash == "" {
		entry.EntryHash = ComputeDNLRouteEntryHash(entry)
	}
	return entry, entry.Validate()
}

func NewDNLCacheEntry(entry DNLCacheEntry) (DNLCacheEntry, error) {
	entry = NormalizeDNLCacheEntry(entry)
	if entry.CacheKey == "" {
		entry.CacheKey = ComputeDNLCacheKey(entry.QueryHash, entry.ResponseHash)
	}
	if entry.EntryHash == "" {
		entry.EntryHash = ComputeDNLCacheEntryHash(entry)
	}
	return entry, entry.Validate()
}

func BuildDNLState(entries []DNLServiceDiscoveryEntry, routes []DNLRoutingTableEntry, cache []DNLCacheEntry, height uint64) (DNLState, error) {
	state := DNLState{
		ServiceEntries:	normalizeDNLServiceEntries(entries),
		RoutingTable:	normalizeDNLRoutes(routes),
		CacheEntries:	normalizeDNLCacheEntries(cache),
		Height:		height,
	}
	if err := state.ValidateFormat(); err != nil {
		return DNLState{}, err
	}
	state.ServicesRoot = ComputeDNLServiceRoot(state.ServiceEntries)
	state.RoutingRoot = ComputeDNLRoutingRoot(state.RoutingTable)
	state.LookupRoot = ComputeDNLLookupRoot(state.ServiceEntries, state.RoutingTable)
	state.CacheRoot = ComputeDNLCacheRoot(state.CacheEntries)
	state.StateRoot = ComputeDNLStateRoot(state)
	return state, state.Validate()
}

func QueryDNL(state DNLState, query DNLQuery) (DNLDiscoveryResponse, error) {
	query = NormalizeDNLQuery(query)
	if err := query.Validate(); err != nil {
		return DNLDiscoveryResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return DNLDiscoveryResponse{}, err
	}
	entries := make([]DNLServiceDiscoveryEntry, 0)
	for _, entry := range state.ServiceEntries {
		if query.CurrentHeight > 0 && query.CurrentHeight > entry.ExpiryHeight {
			continue
		}
		if query.ServiceID != "" && entry.ServiceID != query.ServiceID {
			continue
		}
		if query.ZoneID != "" && entry.ZoneID != query.ZoneID {
			continue
		}
		if query.InterfaceHash != "" && entry.InterfaceHash != query.InterfaceHash {
			continue
		}
		entries = append(entries, entry)
	}
	sortDNLServiceEntriesByLookup(entries)
	limit := query.Limit
	if limit == 0 {
		limit = DefaultDNLQueryLimit
	}
	if len(entries) > int(limit) {
		entries = entries[:limit]
	}
	routes := make([]DNLRoutingTableEntry, 0)
	for _, entry := range entries {
		for _, route := range state.RoutingTable {
			if query.CurrentHeight > 0 && query.CurrentHeight > route.ExpiryHeight {
				continue
			}
			if route.RouteID == entry.RouteID || route.ServiceID == entry.ServiceID && route.ZoneID == entry.ZoneID {
				routes = append(routes, route)
			}
		}
	}
	routes = uniqueSortedDNLRoutes(routes)
	response := DNLDiscoveryResponse{
		QueryHash:	ComputeDNLQueryHash(query),
		Entries:	entries,
		Routes:		routes,
		ExpiryHeight:	minDNLExpiry(entries, routes),
	}
	if query.RequireProof {
		proofKey := DNLQueryProofKey(query)
		proof, err := QueryDNLProof(state, proofKey)
		if err != nil {
			return DNLDiscoveryResponse{}, err
		}
		response.Proof = proof
	}
	response.ResponseHash = ComputeDNLDiscoveryResponseHash(response)
	return response, response.Validate(query.RequireProof)
}

func QueryDNLProof(state DNLState, key string) (DNLProof, error) {
	if err := state.Validate(); err != nil {
		return DNLProof{}, err
	}
	records, err := dnlStateEntries(state)
	if err != nil {
		return DNLProof{}, err
	}
	valueHash := ""
	path := make([]string, 0, len(records))
	for _, record := range records {
		if record.Key == key {
			valueHash = record.ValueHash
			continue
		}
		path = append(path, record.EntryHash)
	}
	if valueHash == "" {
		return DNLProof{}, fmt.Errorf("networking DNL proof key %s not found", key)
	}
	sortStrings(path)
	proof := DNLProof{Key: key, ValueHash: valueHash, StateRoot: state.StateRoot, Height: state.Height, Path: path}
	proof.ProofHash = ComputeDNLProofHash(proof)
	return proof, proof.Validate()
}

func BuildDNLCacheEntryFromResponse(response DNLDiscoveryResponse) (DNLCacheEntry, error) {
	if err := response.Validate(response.Proof.ProofHash != ""); err != nil {
		return DNLCacheEntry{}, err
	}
	return NewDNLCacheEntry(DNLCacheEntry{
		QueryHash:	response.QueryHash,
		ResponseHash:	response.ResponseHash,
		ExpiryHeight:	response.ExpiryHeight,
		ProofHash:	response.Proof.ProofHash,
	})
}

func NewDNLAdvisoryObservation(observation DNLAdvisoryObservation) (DNLAdvisoryObservation, error) {
	observation = NormalizeDNLAdvisoryObservation(observation)
	if observation.ObservationHash == "" {
		observation.ObservationHash = ComputeDNLAdvisoryObservationHash(observation)
	}
	return observation, observation.Validate()
}

func RejectDNLAdvisoryObservationForConsensus(observation DNLAdvisoryObservation) error {
	if err := observation.Validate(); err != nil {
		return err
	}
	return errors.New("networking DNL node-local observations are advisory and must not be used for consensus state")
}

func DNLServiceKey(serviceID, zoneID, entryID string) (string, error) {
	if err := validateIdentifierSet("DNL service id", []string{serviceID}, MaxServiceIDBytes); err != nil {
		return "", err
	}
	if err := validateIdentifierSet("DNL zone id", []string{zoneID}, MaxZoneIDBytes); err != nil {
		return "", err
	}
	if err := ValidateHash("networking DNL service entry id", normalizeHashText(entryID)); err != nil {
		return "", err
	}
	return "dnl/services/" + zoneID + "/" + serviceID + "/" + normalizeHashText(entryID), nil
}

func DNLRouteKey(zoneID, routeID string) (string, error) {
	if err := validateIdentifierSet("DNL route zone id", []string{zoneID}, MaxZoneIDBytes); err != nil {
		return "", err
	}
	if err := ValidateHash("networking DNL route id", normalizeHashText(routeID)); err != nil {
		return "", err
	}
	return "dnl/routes/" + zoneID + "/" + normalizeHashText(routeID), nil
}

func DNLCacheKey(cacheKey string) (string, error) {
	cacheKey = normalizeHashText(cacheKey)
	if err := ValidateHash("networking DNL cache key", cacheKey); err != nil {
		return "", err
	}
	return "dnl/cache/" + cacheKey, nil
}

func DNLQueryProofKey(query DNLQuery) string {
	query = NormalizeDNLQuery(query)
	return "dnl/lookup/" + ComputeDNLQueryHash(DNLQuery{
		ServiceID:	query.ServiceID,
		ZoneID:		query.ZoneID,
		InterfaceHash:	query.InterfaceHash,
		CurrentHeight:	0,
		Limit:		0,
		RequireProof:	false,
	})
}

func NormalizeDNLServiceDiscoveryEntry(entry DNLServiceDiscoveryEntry) DNLServiceDiscoveryEntry {
	entry.EntryID = normalizeHashText(entry.EntryID)
	entry.ServiceID = strings.TrimSpace(entry.ServiceID)
	entry.ZoneID = strings.TrimSpace(entry.ZoneID)
	entry.InterfaceHash = normalizeHashText(entry.InterfaceHash)
	entry.EndpointHash = normalizeHashText(entry.EndpointHash)
	entry.RouteID = normalizeHashText(entry.RouteID)
	entry.ProofHash = normalizeHashText(entry.ProofHash)
	entry.EntryHash = normalizeHashText(entry.EntryHash)
	return entry
}

func NormalizeDNLRoutingTableEntry(entry DNLRoutingTableEntry) DNLRoutingTableEntry {
	entry.RouteID = normalizeHashText(entry.RouteID)
	entry.ZoneID = strings.TrimSpace(entry.ZoneID)
	entry.ServiceID = strings.TrimSpace(entry.ServiceID)
	entry.NextHopNodeID = normalizeHashText(entry.NextHopNodeID)
	entry.OverlayID = normalizeHashText(entry.OverlayID)
	entry.EntryHash = normalizeHashText(entry.EntryHash)
	return entry
}

func NormalizeDNLCacheEntry(entry DNLCacheEntry) DNLCacheEntry {
	entry.CacheKey = normalizeHashText(entry.CacheKey)
	entry.QueryHash = normalizeHashText(entry.QueryHash)
	entry.ResponseHash = normalizeHashText(entry.ResponseHash)
	entry.ProofHash = normalizeHashText(entry.ProofHash)
	entry.EntryHash = normalizeHashText(entry.EntryHash)
	return entry
}

func NormalizeDNLQuery(query DNLQuery) DNLQuery {
	query.ServiceID = strings.TrimSpace(query.ServiceID)
	query.ZoneID = strings.TrimSpace(query.ZoneID)
	query.InterfaceHash = normalizeHashText(query.InterfaceHash)
	return query
}

func NormalizeDNLAdvisoryObservation(observation DNLAdvisoryObservation) DNLAdvisoryObservation {
	observation.ObservedNodeID = normalizeHashText(observation.ObservedNodeID)
	observation.ServiceID = strings.TrimSpace(observation.ServiceID)
	observation.ZoneID = strings.TrimSpace(observation.ZoneID)
	observation.EndpointHash = normalizeHashText(observation.EndpointHash)
	observation.ObservationHash = normalizeHashText(observation.ObservationHash)
	return observation
}

func (entry DNLServiceDiscoveryEntry) Validate() error {
	entry = NormalizeDNLServiceDiscoveryEntry(entry)
	if err := ValidateHash("networking DNL service entry id", entry.EntryID); err != nil {
		return err
	}
	if entry.EntryID != ComputeDNLServiceEntryID(entry) {
		return errors.New("networking DNL service entry id mismatch")
	}
	if err := validateIdentifierSet("DNL service id", []string{entry.ServiceID}, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL zone id", []string{entry.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL interface hash", entry.InterfaceHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL endpoint hash", entry.EndpointHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL route id", entry.RouteID); err != nil {
		return err
	}
	if entry.ExpiryHeight == 0 {
		return errors.New("networking DNL service entry expiry height must be positive")
	}
	if err := ValidateHash("networking DNL service proof hash", entry.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL service entry hash", entry.EntryHash); err != nil {
		return err
	}
	if entry.EntryHash != ComputeDNLServiceEntryHash(entry) {
		return errors.New("networking DNL service entry hash mismatch")
	}
	return nil
}

func (entry DNLRoutingTableEntry) Validate() error {
	entry = NormalizeDNLRoutingTableEntry(entry)
	if err := ValidateHash("networking DNL route id", entry.RouteID); err != nil {
		return err
	}
	if entry.RouteID != ComputeDNLRouteID(entry) {
		return errors.New("networking DNL route id mismatch")
	}
	if err := validateIdentifierSet("DNL route zone id", []string{entry.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL route service id", []string{entry.ServiceID}, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL route next hop", entry.NextHopNodeID); err != nil {
		return err
	}
	if entry.OverlayID != "" {
		if err := ValidateHash("networking DNL route overlay", entry.OverlayID); err != nil {
			return err
		}
	}
	if entry.WeightBps > BasisPoints {
		return fmt.Errorf("networking DNL route weight must be <= %d bps", BasisPoints)
	}
	if entry.ExpiryHeight == 0 {
		return errors.New("networking DNL route expiry height must be positive")
	}
	if err := ValidateHash("networking DNL route entry hash", entry.EntryHash); err != nil {
		return err
	}
	if entry.EntryHash != ComputeDNLRouteEntryHash(entry) {
		return errors.New("networking DNL route entry hash mismatch")
	}
	return nil
}

func (entry DNLCacheEntry) Validate() error {
	entry = NormalizeDNLCacheEntry(entry)
	if err := ValidateHash("networking DNL cache key", entry.CacheKey); err != nil {
		return err
	}
	if entry.CacheKey != ComputeDNLCacheKey(entry.QueryHash, entry.ResponseHash) {
		return errors.New("networking DNL cache key mismatch")
	}
	if err := ValidateHash("networking DNL cache query hash", entry.QueryHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL cache response hash", entry.ResponseHash); err != nil {
		return err
	}
	if entry.ExpiryHeight == 0 {
		return errors.New("networking DNL cache expiry height must be positive")
	}
	if err := ValidateHash("networking DNL cache proof hash", entry.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL cache entry hash", entry.EntryHash); err != nil {
		return err
	}
	if entry.EntryHash != ComputeDNLCacheEntryHash(entry) {
		return errors.New("networking DNL cache entry hash mismatch")
	}
	return nil
}

func (state DNLState) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("networking DNL state height must be positive")
	}
	if len(state.ServiceEntries) > MaxDNLEntries {
		return fmt.Errorf("networking DNL service entries must be <= %d", MaxDNLEntries)
	}
	if len(state.RoutingTable) > MaxDNLRoutes {
		return fmt.Errorf("networking DNL routes must be <= %d", MaxDNLRoutes)
	}
	if len(state.CacheEntries) > MaxDNLCache {
		return fmt.Errorf("networking DNL cache entries must be <= %d", MaxDNLCache)
	}
	if err := validateDNLServiceEntries(state.ServiceEntries); err != nil {
		return err
	}
	if err := validateDNLRoutes(state.RoutingTable, state.ServiceEntries); err != nil {
		return err
	}
	if err := validateDNLCacheEntries(state.CacheEntries, state.Height); err != nil {
		return err
	}
	if state.StateRoot != "" {
		if err := ValidateHash("networking DNL services root", state.ServicesRoot); err != nil {
			return err
		}
		if err := ValidateHash("networking DNL routing root", state.RoutingRoot); err != nil {
			return err
		}
		if err := ValidateHash("networking DNL lookup root", state.LookupRoot); err != nil {
			return err
		}
		if err := ValidateHash("networking DNL cache root", state.CacheRoot); err != nil {
			return err
		}
		if err := ValidateHash("networking DNL state root", state.StateRoot); err != nil {
			return err
		}
	}
	return nil
}

func (state DNLState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("networking DNL state root is required")
	}
	expected := DNLState{
		ServiceEntries:	state.ServiceEntries,
		RoutingTable:	state.RoutingTable,
		CacheEntries:	state.CacheEntries,
		Height:		state.Height,
		ServicesRoot:	ComputeDNLServiceRoot(state.ServiceEntries),
		RoutingRoot:	ComputeDNLRoutingRoot(state.RoutingTable),
		LookupRoot:	ComputeDNLLookupRoot(state.ServiceEntries, state.RoutingTable),
		CacheRoot:	ComputeDNLCacheRoot(state.CacheEntries),
	}
	expected.StateRoot = ComputeDNLStateRoot(expected)
	if state.ServicesRoot != expected.ServicesRoot || state.RoutingRoot != expected.RoutingRoot || state.LookupRoot != expected.LookupRoot || state.CacheRoot != expected.CacheRoot || state.StateRoot != expected.StateRoot {
		return errors.New("networking DNL state root mismatch")
	}
	return nil
}

func (query DNLQuery) Validate() error {
	query = NormalizeDNLQuery(query)
	if query.ServiceID != "" {
		if err := validateIdentifierSet("DNL query service id", []string{query.ServiceID}, MaxServiceIDBytes); err != nil {
			return err
		}
	}
	if query.ZoneID != "" {
		if err := validateIdentifierSet("DNL query zone id", []string{query.ZoneID}, MaxZoneIDBytes); err != nil {
			return err
		}
	}
	if query.InterfaceHash != "" {
		if err := ValidateHash("networking DNL query interface hash", query.InterfaceHash); err != nil {
			return err
		}
	}
	if query.CurrentHeight == 0 {
		return errors.New("networking DNL query current height must be positive")
	}
	if query.Limit > MaxDNLQueryLimit {
		return fmt.Errorf("networking DNL query limit must be <= %d", MaxDNLQueryLimit)
	}
	if query.ServiceID == "" && query.ZoneID == "" && query.InterfaceHash == "" {
		return errors.New("networking DNL query requires service, zone, or interface")
	}
	return nil
}

func (proof DNLProof) Validate() error {
	if proof.Key == "" {
		return errors.New("networking DNL proof key is required")
	}
	if err := ValidateHash("networking DNL proof value hash", proof.ValueHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL proof state root", proof.StateRoot); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("networking DNL proof height must be positive")
	}
	for _, item := range proof.Path {
		if err := ValidateHash("networking DNL proof path item", item); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking DNL proof hash", proof.ProofHash); err != nil {
		return err
	}
	if proof.ProofHash != ComputeDNLProofHash(proof) {
		return errors.New("networking DNL proof hash mismatch")
	}
	return nil
}

func (response DNLDiscoveryResponse) Validate(requireProof bool) error {
	if err := ValidateHash("networking DNL response query hash", response.QueryHash); err != nil {
		return err
	}
	if response.ExpiryHeight == 0 {
		return errors.New("networking DNL response expiry height must be positive")
	}
	if requireProof {
		if err := response.Proof.Validate(); err != nil {
			return err
		}
	}
	if err := validateDNLServiceEntries(response.Entries); err != nil {
		return err
	}
	if err := validateDNLRoutes(normalizeDNLRoutes(response.Routes), response.Entries); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL response hash", response.ResponseHash); err != nil {
		return err
	}
	if response.ResponseHash != ComputeDNLDiscoveryResponseHash(response) {
		return errors.New("networking DNL response hash mismatch")
	}
	return nil
}

func (observation DNLAdvisoryObservation) Validate() error {
	observation = NormalizeDNLAdvisoryObservation(observation)
	if err := ValidateHash("networking DNL advisory node id", observation.ObservedNodeID); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL advisory service id", []string{observation.ServiceID}, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL advisory zone id", []string{observation.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL advisory endpoint hash", observation.EndpointHash); err != nil {
		return err
	}
	if observation.ObservedHeight == 0 {
		return errors.New("networking DNL advisory observation height must be positive")
	}
	if err := ValidateHash("networking DNL advisory observation hash", observation.ObservationHash); err != nil {
		return err
	}
	if observation.ObservationHash != ComputeDNLAdvisoryObservationHash(observation) {
		return errors.New("networking DNL advisory observation hash mismatch")
	}
	return nil
}

func ComputeDNLServiceEntryID(entry DNLServiceDiscoveryEntry) string {
	entry = NormalizeDNLServiceDiscoveryEntry(entry)
	return HashParts("dnl-service-entry-id", entry.ServiceID, entry.ZoneID, entry.InterfaceHash, entry.EndpointHash, entry.RouteID)
}

func ComputeDNLServiceEntryHash(entry DNLServiceDiscoveryEntry) string {
	entry = NormalizeDNLServiceDiscoveryEntry(entry)
	return HashParts("dnl-service-entry", entry.EntryID, entry.ServiceID, entry.ZoneID, entry.InterfaceHash, entry.EndpointHash, entry.RouteID, fmt.Sprintf("%d", entry.ExpiryHeight), entry.ProofHash)
}

func ComputeDNLRouteID(entry DNLRoutingTableEntry) string {
	entry = NormalizeDNLRoutingTableEntry(entry)
	return HashParts("dnl-route-id", entry.ZoneID, entry.ServiceID, entry.NextHopNodeID, entry.OverlayID, fmt.Sprintf("%d", entry.Priority))
}

func ComputeDNLRouteEntryHash(entry DNLRoutingTableEntry) string {
	entry = NormalizeDNLRoutingTableEntry(entry)
	return HashParts("dnl-route-entry", entry.RouteID, entry.ZoneID, entry.ServiceID, entry.NextHopNodeID, entry.OverlayID, fmt.Sprintf("%d", entry.Priority), fmt.Sprintf("%d", entry.WeightBps), fmt.Sprintf("%d", entry.ExpiryHeight))
}

func ComputeDNLCacheKey(queryHash, responseHash string) string {
	return HashParts("dnl-cache-key", normalizeHashText(queryHash), normalizeHashText(responseHash))
}

func ComputeDNLCacheEntryHash(entry DNLCacheEntry) string {
	entry = NormalizeDNLCacheEntry(entry)
	return HashParts("dnl-cache-entry", entry.CacheKey, entry.QueryHash, entry.ResponseHash, fmt.Sprintf("%d", entry.ExpiryHeight), entry.ProofHash)
}

func ComputeDNLQueryHash(query DNLQuery) string {
	query = NormalizeDNLQuery(query)
	return HashParts("dnl-query", query.ServiceID, query.ZoneID, query.InterfaceHash, fmt.Sprintf("%d", query.CurrentHeight), fmt.Sprintf("%d", query.Limit), fmt.Sprintf("%t", query.RequireProof))
}

func ComputeDNLServiceRoot(entries []DNLServiceDiscoveryEntry) string {
	ordered := normalizeDNLServiceEntries(entries)
	parts := []string{"dnl-services-root", fmt.Sprintf("%d", len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, entry.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeDNLRoutingRoot(routes []DNLRoutingTableEntry) string {
	ordered := normalizeDNLRoutes(routes)
	parts := []string{"dnl-routing-root", fmt.Sprintf("%d", len(ordered))}
	for _, route := range ordered {
		parts = append(parts, route.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeDNLLookupRoot(entries []DNLServiceDiscoveryEntry, routes []DNLRoutingTableEntry) string {
	records := dnlLookupRecords(entries, routes)
	parts := []string{"dnl-lookup-root", fmt.Sprintf("%d", len(records))}
	for _, record := range records {
		parts = append(parts, record.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeDNLCacheRoot(entries []DNLCacheEntry) string {
	ordered := normalizeDNLCacheEntries(entries)
	parts := []string{"dnl-cache-root", fmt.Sprintf("%d", len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, entry.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeDNLStateRoot(state DNLState) string {
	return HashParts("dnl-state-root", fmt.Sprintf("%d", state.Height), state.ServicesRoot, state.RoutingRoot, state.LookupRoot, state.CacheRoot)
}

func ComputeDNLProofHash(proof DNLProof) string {
	parts := []string{"dnl-proof", proof.Key, proof.ValueHash, proof.StateRoot, fmt.Sprintf("%d", proof.Height), fmt.Sprintf("%d", len(proof.Path))}
	ordered := append([]string(nil), proof.Path...)
	sortStrings(ordered)
	parts = append(parts, ordered...)
	return HashParts(parts...)
}

func ComputeDNLDiscoveryResponseHash(response DNLDiscoveryResponse) string {
	entries := normalizeDNLServiceEntries(response.Entries)
	routes := normalizeDNLRoutes(response.Routes)
	parts := []string{"dnl-discovery-response", response.QueryHash, fmt.Sprintf("%d", response.ExpiryHeight), response.Proof.ProofHash, fmt.Sprintf("%d", len(entries)), fmt.Sprintf("%d", len(routes))}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	for _, route := range routes {
		parts = append(parts, route.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeDNLAdvisoryObservationHash(observation DNLAdvisoryObservation) string {
	observation = NormalizeDNLAdvisoryObservation(observation)
	return HashParts("dnl-advisory-observation", observation.ObservedNodeID, observation.ServiceID, observation.ZoneID, observation.EndpointHash, fmt.Sprintf("%d", observation.ObservedHeight))
}

func dnlStateEntries(state DNLState) ([]DNLStateEntry, error) {
	records := make([]DNLStateEntry, 0, len(state.ServiceEntries)+len(state.RoutingTable)+len(state.CacheEntries)+16)
	for _, entry := range state.ServiceEntries {
		key, err := DNLServiceKey(entry.ServiceID, entry.ZoneID, entry.EntryID)
		if err != nil {
			return nil, err
		}
		records = append(records, dnlStateEntry(key, entry.EntryHash))
	}
	for _, route := range state.RoutingTable {
		key, err := DNLRouteKey(route.ZoneID, route.RouteID)
		if err != nil {
			return nil, err
		}
		records = append(records, dnlStateEntry(key, route.EntryHash))
	}
	for _, cache := range state.CacheEntries {
		key, err := DNLCacheKey(cache.CacheKey)
		if err != nil {
			return nil, err
		}
		records = append(records, dnlStateEntry(key, cache.EntryHash))
	}
	records = append(records, dnlLookupRecords(state.ServiceEntries, state.RoutingTable)...)
	sortDNLStateEntries(records)
	return records, nil
}

func dnlLookupRecords(entries []DNLServiceDiscoveryEntry, routes []DNLRoutingTableEntry) []DNLStateEntry {
	values := map[string][]string{}
	for _, entry := range normalizeDNLServiceEntries(entries) {
		serviceQuery := DNLQueryProofKey(DNLQuery{ServiceID: entry.ServiceID})
		zoneQuery := DNLQueryProofKey(DNLQuery{ZoneID: entry.ZoneID})
		interfaceQuery := DNLQueryProofKey(DNLQuery{InterfaceHash: entry.InterfaceHash})
		values[serviceQuery] = append(values[serviceQuery], entry.EntryHash)
		values[zoneQuery] = append(values[zoneQuery], entry.EntryHash)
		values[interfaceQuery] = append(values[interfaceQuery], entry.EntryHash)
	}
	for _, route := range normalizeDNLRoutes(routes) {
		zoneQuery := DNLQueryProofKey(DNLQuery{ServiceID: route.ServiceID, ZoneID: route.ZoneID})
		values[zoneQuery] = append(values[zoneQuery], route.EntryHash)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sortStrings(keys)
	out := make([]DNLStateEntry, 0, len(keys))
	for _, key := range keys {
		hashes := normalizeHashList(values[key])
		out = append(out, dnlStateEntry(key, HashParts(append([]string{"dnl-lookup-value", fmt.Sprintf("%d", len(hashes))}, hashes...)...)))
	}
	return out
}

func dnlStateEntry(key, valueHash string) DNLStateEntry {
	return DNLStateEntry{Key: key, ValueHash: valueHash, EntryHash: HashParts("dnl-state-entry", key, valueHash)}
}

func normalizeDNLServiceEntries(entries []DNLServiceDiscoveryEntry) []DNLServiceDiscoveryEntry {
	out := make([]DNLServiceDiscoveryEntry, len(entries))
	for i, entry := range entries {
		out[i] = NormalizeDNLServiceDiscoveryEntry(entry)
	}
	sort.SliceStable(out, func(i, j int) bool { return dnlServiceEntryKey(out[i]) < dnlServiceEntryKey(out[j]) })
	return out
}

func normalizeDNLRoutes(routes []DNLRoutingTableEntry) []DNLRoutingTableEntry {
	out := make([]DNLRoutingTableEntry, len(routes))
	for i, route := range routes {
		out[i] = NormalizeDNLRoutingTableEntry(route)
	}
	sort.SliceStable(out, func(i, j int) bool { return dnlRouteKey(out[i]) < dnlRouteKey(out[j]) })
	return out
}

func normalizeDNLCacheEntries(entries []DNLCacheEntry) []DNLCacheEntry {
	out := make([]DNLCacheEntry, len(entries))
	for i, entry := range entries {
		out[i] = NormalizeDNLCacheEntry(entry)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CacheKey < out[j].CacheKey })
	return out
}

func normalizeHashList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		normalized := normalizeHashText(value)
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sortStrings(out)
	return out
}

func sortDNLStateEntries(entries []DNLStateEntry) {
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
}

func sortDNLServiceEntriesByLookup(entries []DNLServiceDiscoveryEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]
		if left.ZoneID != right.ZoneID {
			return left.ZoneID < right.ZoneID
		}
		if left.ServiceID != right.ServiceID {
			return left.ServiceID < right.ServiceID
		}
		if left.ExpiryHeight != right.ExpiryHeight {
			return left.ExpiryHeight < right.ExpiryHeight
		}
		return left.EntryID < right.EntryID
	})
}

func uniqueSortedDNLRoutes(routes []DNLRoutingTableEntry) []DNLRoutingTableEntry {
	ordered := normalizeDNLRoutes(routes)
	out := make([]DNLRoutingTableEntry, 0, len(ordered))
	seen := map[string]struct{}{}
	for _, route := range ordered {
		if _, found := seen[route.RouteID]; found {
			continue
		}
		seen[route.RouteID] = struct{}{}
		out = append(out, route)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		if out[i].WeightBps != out[j].WeightBps {
			return out[i].WeightBps > out[j].WeightBps
		}
		return out[i].RouteID < out[j].RouteID
	})
	return out
}

func validateDNLServiceEntries(entries []DNLServiceDiscoveryEntry) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		key := dnlServiceEntryKey(entry)
		if _, found := seen[key]; found {
			return errors.New("networking duplicate DNL service entry")
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("networking DNL service entries must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateDNLRoutes(routes []DNLRoutingTableEntry, entries []DNLServiceDiscoveryEntry) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, route := range routes {
		if err := route.Validate(); err != nil {
			return err
		}
		if len(entries) > 0 && !dnlRouteReferencesService(route, entries) {
			return errors.New("networking DNL route must reference committed service entry")
		}
		key := dnlRouteKey(route)
		if _, found := seen[key]; found {
			return errors.New("networking duplicate DNL route")
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("networking DNL routes must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateDNLCacheEntries(entries []DNLCacheEntry, currentHeight uint64) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if currentHeight > 0 && currentHeight > entry.ExpiryHeight {
			return errors.New("networking DNL committed cache entry is expired")
		}
		if _, found := seen[entry.CacheKey]; found {
			return errors.New("networking duplicate DNL cache entry")
		}
		seen[entry.CacheKey] = struct{}{}
		if previous != "" && previous >= entry.CacheKey {
			return errors.New("networking DNL cache entries must be sorted canonically")
		}
		previous = entry.CacheKey
	}
	return nil
}

func dnlRouteReferencesService(route DNLRoutingTableEntry, entries []DNLServiceDiscoveryEntry) bool {
	for _, entry := range entries {
		if entry.RouteID == route.RouteID || entry.ServiceID == route.ServiceID && entry.ZoneID == route.ZoneID {
			return true
		}
	}
	return false
}

func dnlServiceEntryKey(entry DNLServiceDiscoveryEntry) string {
	entry = NormalizeDNLServiceDiscoveryEntry(entry)
	return entry.ZoneID + "/" + entry.ServiceID + "/" + entry.EntryID
}

func dnlRouteKey(entry DNLRoutingTableEntry) string {
	entry = NormalizeDNLRoutingTableEntry(entry)
	return entry.ZoneID + "/" + entry.ServiceID + "/" + entry.RouteID
}

func minDNLExpiry(entries []DNLServiceDiscoveryEntry, routes []DNLRoutingTableEntry) uint64 {
	min := uint64(0)
	for _, entry := range entries {
		if min == 0 || entry.ExpiryHeight < min {
			min = entry.ExpiryHeight
		}
	}
	for _, route := range routes {
		if min == 0 || route.ExpiryHeight < min {
			min = route.ExpiryHeight
		}
	}
	if min == 0 {
		return 1
	}
	return min
}
