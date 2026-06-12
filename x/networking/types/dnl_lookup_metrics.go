package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultDNLRecursiveDepth	= uint32(2)
	MaxDNLRecursiveDepth		= uint32(8)
	MaxDNLRecursiveHops		= 512
	MaxDNLMetricSnapshots		= 8192
)

type DNLRecursiveLookupRequest struct {
	ServiceID		string
	ZoneID			string
	InterfaceHash		string
	CurrentHeight		uint64
	MaxDepth		uint32
	Limit			uint32
	ConsensusRoutingOnly	bool
	AllowNodeLocalHints	bool
}

type DNLLookupHop struct {
	Depth		uint32
	LookupKey	string
	NodeID		string
	ServiceID	string
	ZoneID		string
	RouteID		string
	ProofHash	string
	HopHash		string
}

type DNLRecursiveLookupResponse struct {
	RequestHash	string
	Hops		[]DNLLookupHop
	Entries		[]DNLServiceDiscoveryEntry
	Routes		[]DNLRoutingTableEntry
	Proofs		[]DNLProof
	ExpiryHeight	uint64
	ResponseHash	string
}

type DNLRoutingMetricSnapshot struct {
	SnapshotID		string
	RouteID			string
	NodeID			string
	ZoneID			string
	ServiceID		string
	LatencyMillis		uint64
	GasCost			uint64
	ReliabilityScoreBps	uint32
	CongestionWeightBps	uint32
	ZoneSupport		bool
	ServiceSupport		bool
	Committed		bool
	SnapshotHeight		uint64
	SnapshotHash		string
}

type DNLLiveRoutingHint struct {
	RouteID		string
	NodeID		string
	LatencyMillis	uint64
	ObservedHeight	uint64
	HintHash	string
}

type DNLRouteSelection struct {
	Route			DNLRoutingTableEntry
	Metric			DNLRoutingMetricSnapshot
	Score			uint64
	UsedCommittedTable	bool
	UsedLiveHint		bool
	SelectionHash		string
}

func RecursiveDNLLookup(dnl DNLState, routing DNLRoutingState, request DNLRecursiveLookupRequest) (DNLRecursiveLookupResponse, error) {
	request = NormalizeDNLRecursiveLookupRequest(request)
	if err := request.Validate(); err != nil {
		return DNLRecursiveLookupResponse{}, err
	}
	if err := dnl.Validate(); err != nil {
		return DNLRecursiveLookupResponse{}, err
	}
	if err := routing.Validate(); err != nil {
		return DNLRecursiveLookupResponse{}, err
	}
	if request.ConsensusRoutingOnly && request.AllowNodeLocalHints {
		return DNLRecursiveLookupResponse{}, errors.New("networking DNL consensus lookup cannot use node-local hints")
	}

	type pendingQuery struct {
		query	DNLQuery
		depth	uint32
	}
	queue := []pendingQuery{{
		query: DNLQuery{
			ServiceID:	request.ServiceID,
			ZoneID:		request.ZoneID,
			InterfaceHash:	request.InterfaceHash,
			CurrentHeight:	request.CurrentHeight,
			Limit:		request.Limit,
			RequireProof:	true,
		},
		depth:	0,
	}}
	seenQueries := map[string]struct{}{}
	entriesByID := map[string]DNLServiceDiscoveryEntry{}
	routesByID := map[string]DNLRoutingTableEntry{}
	proofsByHash := map[string]DNLProof{}
	hops := make([]DNLLookupHop, 0)
	var lastLookupErr error

	for len(queue) > 0 && len(hops) < MaxDNLRecursiveHops {
		next := queue[0]
		queue = queue[1:]
		queryHash := ComputeDNLQueryHash(next.query)
		if _, found := seenQueries[queryHash]; found {
			continue
		}
		seenQueries[queryHash] = struct{}{}

		queryWithoutProof := next.query
		queryWithoutProof.RequireProof = false
		response, err := QueryDNL(dnl, queryWithoutProof)
		if err != nil {
			lastLookupErr = err
			continue
		}
		attachmentProof := NewDNLRootAttachmentProof(dnl, "dnl/recursive/"+queryHash, response.ResponseHash)
		proofsByHash[attachmentProof.ProofHash] = attachmentProof
		queryProof, err := QueryDNLProof(dnl, DNLQueryProofKey(next.query))
		if err == nil {
			proofsByHash[queryProof.ProofHash] = queryProof
		}
		for _, entry := range response.Entries {
			if request.CurrentHeight > entry.ExpiryHeight {
				continue
			}
			entriesByID[entry.EntryID] = entry
			key, err := DNLServiceKey(entry.ServiceID, entry.ZoneID, entry.EntryID)
			if err == nil {
				proof, err := QueryDNLProof(dnl, key)
				if err == nil {
					proofsByHash[proof.ProofHash] = proof
				}
			}
		}
		for _, route := range response.Routes {
			if request.CurrentHeight > route.ExpiryHeight {
				continue
			}
			node, found := QueryRoutingNode(routing, route.NextHopNodeID)
			if !found || request.CurrentHeight > node.ExpiresHeight {
				continue
			}
			if !containsString(node.ZonesSupported, route.ZoneID) || !containsString(node.ServiceIDs, route.ServiceID) {
				continue
			}
			if _, found := routesByID[route.RouteID]; found {
				continue
			}
			routesByID[route.RouteID] = route
			routeProofHash := ""
			key, err := DNLRouteKey(route.ZoneID, route.RouteID)
			if err == nil {
				proof, err := QueryDNLProof(dnl, key)
				if err == nil {
					proofsByHash[proof.ProofHash] = proof
					routeProofHash = proof.ProofHash
				}
			}
			if routeProofHash == "" && queryProof.ProofHash != "" {
				routeProofHash = queryProof.ProofHash
			}
			if routeProofHash == "" {
				routeProofHash = attachmentProof.ProofHash
			}
			hop, err := NewDNLLookupHop(DNLLookupHop{
				Depth:		next.depth,
				LookupKey:	DNLQueryProofKey(next.query),
				NodeID:		route.NextHopNodeID,
				ServiceID:	route.ServiceID,
				ZoneID:		route.ZoneID,
				RouteID:	route.RouteID,
				ProofHash:	routeProofHash,
			})
			if err != nil {
				return DNLRecursiveLookupResponse{}, err
			}
			hops = append(hops, hop)
			if next.depth+1 < request.MaxDepth {
				for _, serviceID := range node.ServiceIDs {
					child := DNLQuery{
						ServiceID:	serviceID,
						ZoneID:		route.ZoneID,
						CurrentHeight:	request.CurrentHeight,
						Limit:		request.Limit,
						RequireProof:	true,
					}
					queue = append(queue, pendingQuery{query: child, depth: next.depth + 1})
				}
			}
		}
	}

	entries := make([]DNLServiceDiscoveryEntry, 0, len(entriesByID))
	entryIDs := make([]string, 0, len(entriesByID))
	for entryID := range entriesByID {
		entryIDs = append(entryIDs, entryID)
	}
	sortStrings(entryIDs)
	for _, entryID := range entryIDs {
		entries = append(entries, entriesByID[entryID])
	}
	routes := make([]DNLRoutingTableEntry, 0, len(routesByID))
	routeIDs := make([]string, 0, len(routesByID))
	for routeID := range routesByID {
		routeIDs = append(routeIDs, routeID)
	}
	sortStrings(routeIDs)
	for _, routeID := range routeIDs {
		routes = append(routes, routesByID[routeID])
	}
	proofs := make([]DNLProof, 0, len(proofsByHash))
	proofHashes := make([]string, 0, len(proofsByHash))
	for proofHash := range proofsByHash {
		proofHashes = append(proofHashes, proofHash)
	}
	sortStrings(proofHashes)
	for _, proofHash := range proofHashes {
		proofs = append(proofs, proofsByHash[proofHash])
	}
	sortDNLLookupHops(hops)
	response := DNLRecursiveLookupResponse{
		RequestHash:	ComputeDNLRecursiveLookupRequestHash(request),
		Hops:		hops,
		Entries:	entries,
		Routes:		routes,
		Proofs:		proofs,
		ExpiryHeight:	minDNLExpiry(entries, routes),
	}
	response.ResponseHash = ComputeDNLRecursiveLookupResponseHash(response)
	if len(response.Proofs) == 0 && lastLookupErr != nil {
		return DNLRecursiveLookupResponse{}, lastLookupErr
	}
	return response, response.Validate()
}

func NewDNLLookupHop(hop DNLLookupHop) (DNLLookupHop, error) {
	hop = NormalizeDNLLookupHop(hop)
	if hop.HopHash == "" {
		hop.HopHash = ComputeDNLLookupHopHash(hop)
	}
	return hop, hop.Validate()
}

func NewDNLRootAttachmentProof(state DNLState, key, valueHash string) DNLProof {
	proof := DNLProof{
		Key:		strings.TrimSpace(key),
		ValueHash:	normalizeHashText(valueHash),
		StateRoot:	state.StateRoot,
		Height:		state.Height,
	}
	proof.ProofHash = ComputeDNLProofHash(proof)
	return proof
}

func NewDNLRoutingMetricSnapshot(snapshot DNLRoutingMetricSnapshot) (DNLRoutingMetricSnapshot, error) {
	snapshot = NormalizeDNLRoutingMetricSnapshot(snapshot)
	if snapshot.SnapshotID == "" {
		snapshot.SnapshotID = ComputeDNLRoutingMetricSnapshotID(snapshot)
	}
	if snapshot.SnapshotHash == "" {
		snapshot.SnapshotHash = ComputeDNLRoutingMetricSnapshotHash(snapshot)
	}
	return snapshot, snapshot.Validate()
}

func NewDNLLiveRoutingHint(hint DNLLiveRoutingHint) (DNLLiveRoutingHint, error) {
	hint = NormalizeDNLLiveRoutingHint(hint)
	if hint.HintHash == "" {
		hint.HintHash = ComputeDNLLiveRoutingHintHash(hint)
	}
	return hint, hint.Validate()
}

func SelectConsensusRouteFromCommittedMetrics(state DNLRoutingState, epoch uint64, metrics []DNLRoutingMetricSnapshot, request DNLRecursiveLookupRequest) (DNLRouteSelection, error) {
	request = NormalizeDNLRecursiveLookupRequest(request)
	if err := state.Validate(); err != nil {
		return DNLRouteSelection{}, err
	}
	if err := request.Validate(); err != nil {
		return DNLRouteSelection{}, err
	}
	table, found := LookupRoutingTable(state, epoch)
	if !found {
		return DNLRouteSelection{}, errors.New("networking DNL consensus routing requires committed routing table")
	}
	if err := table.Validate(); err != nil {
		return DNLRouteSelection{}, err
	}
	metricByRoute := map[string]DNLRoutingMetricSnapshot{}
	for _, metric := range normalizeDNLMetricSnapshots(metrics) {
		if err := metric.Validate(); err != nil {
			return DNLRouteSelection{}, err
		}
		if !metric.Committed {
			return DNLRouteSelection{}, errors.New("networking DNL consensus routing metrics must be committed")
		}
		metricByRoute[metric.RouteID] = metric
	}
	candidates := make([]DNLRouteSelection, 0)
	for _, route := range table.Routes {
		if request.CurrentHeight > route.ExpiryHeight {
			continue
		}
		if request.ServiceID != "" && route.ServiceID != request.ServiceID {
			continue
		}
		if request.ZoneID != "" && route.ZoneID != request.ZoneID {
			continue
		}
		node, found := QueryRoutingNode(state, route.NextHopNodeID)
		if !found || request.CurrentHeight > node.ExpiresHeight {
			continue
		}
		if !containsString(node.ZonesSupported, route.ZoneID) || !containsString(node.ServiceIDs, route.ServiceID) {
			continue
		}
		metric, found := metricByRoute[route.RouteID]
		if !found {
			continue
		}
		if metric.NodeID != route.NextHopNodeID || metric.ZoneID != route.ZoneID || metric.ServiceID != route.ServiceID {
			continue
		}
		selection := DNLRouteSelection{
			Route:			route,
			Metric:			metric,
			Score:			ComputeDNLRouteMetricScore(route, metric),
			UsedCommittedTable:	true,
			UsedLiveHint:		false,
		}
		selection.SelectionHash = ComputeDNLRouteSelectionHash(selection)
		candidates = append(candidates, selection)
	}
	if len(candidates) == 0 {
		return DNLRouteSelection{}, errors.New("networking DNL consensus routing has no committed candidates")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Route.RouteID < candidates[j].Route.RouteID
	})
	return candidates[0], candidates[0].Validate(true)
}

func RejectLiveRoutingHintForConsensus(hint DNLLiveRoutingHint) error {
	if err := hint.Validate(); err != nil {
		return err
	}
	return errors.New("networking DNL live latency measurements are advisory until committed")
}

func ValidateDNLMetricConsensusUse(snapshot DNLRoutingMetricSnapshot, usedForConsensus bool) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if usedForConsensus && !snapshot.Committed {
		return errors.New("networking DNL consensus routing requires committed metric snapshot")
	}
	return nil
}

func ComputeDNLMetricRoot(snapshots []DNLRoutingMetricSnapshot) string {
	ordered := normalizeDNLMetricSnapshots(snapshots)
	parts := []string{"dnl-routing-metric-root", fmt.Sprintf("%d", len(ordered))}
	for _, snapshot := range ordered {
		parts = append(parts, snapshot.SnapshotHash)
	}
	return HashParts(parts...)
}

func NormalizeDNLRecursiveLookupRequest(request DNLRecursiveLookupRequest) DNLRecursiveLookupRequest {
	request.ServiceID = strings.TrimSpace(request.ServiceID)
	request.ZoneID = strings.TrimSpace(request.ZoneID)
	request.InterfaceHash = normalizeHashText(request.InterfaceHash)
	if request.MaxDepth == 0 {
		request.MaxDepth = DefaultDNLRecursiveDepth
	}
	if request.Limit == 0 {
		request.Limit = DefaultDNLQueryLimit
	}
	return request
}

func NormalizeDNLLookupHop(hop DNLLookupHop) DNLLookupHop {
	hop.LookupKey = strings.TrimSpace(hop.LookupKey)
	hop.NodeID = normalizeHashText(hop.NodeID)
	hop.ServiceID = strings.TrimSpace(hop.ServiceID)
	hop.ZoneID = strings.TrimSpace(hop.ZoneID)
	hop.RouteID = normalizeHashText(hop.RouteID)
	hop.ProofHash = normalizeHashText(hop.ProofHash)
	hop.HopHash = normalizeHashText(hop.HopHash)
	return hop
}

func NormalizeDNLRoutingMetricSnapshot(snapshot DNLRoutingMetricSnapshot) DNLRoutingMetricSnapshot {
	snapshot.SnapshotID = normalizeHashText(snapshot.SnapshotID)
	snapshot.RouteID = normalizeHashText(snapshot.RouteID)
	snapshot.NodeID = normalizeHashText(snapshot.NodeID)
	snapshot.ZoneID = strings.TrimSpace(snapshot.ZoneID)
	snapshot.ServiceID = strings.TrimSpace(snapshot.ServiceID)
	snapshot.SnapshotHash = normalizeHashText(snapshot.SnapshotHash)
	return snapshot
}

func NormalizeDNLLiveRoutingHint(hint DNLLiveRoutingHint) DNLLiveRoutingHint {
	hint.RouteID = normalizeHashText(hint.RouteID)
	hint.NodeID = normalizeHashText(hint.NodeID)
	hint.HintHash = normalizeHashText(hint.HintHash)
	return hint
}

func (request DNLRecursiveLookupRequest) Validate() error {
	request = NormalizeDNLRecursiveLookupRequest(request)
	if request.ServiceID == "" && request.ZoneID == "" && request.InterfaceHash == "" {
		return errors.New("networking DNL recursive lookup requires service, zone, or interface")
	}
	if request.ServiceID != "" {
		if err := validateIdentifierSet("DNL recursive service id", []string{request.ServiceID}, MaxServiceIDBytes); err != nil {
			return err
		}
	}
	if request.ZoneID != "" {
		if err := validateIdentifierSet("DNL recursive zone id", []string{request.ZoneID}, MaxZoneIDBytes); err != nil {
			return err
		}
	}
	if request.InterfaceHash != "" {
		if err := ValidateHash("networking DNL recursive interface hash", request.InterfaceHash); err != nil {
			return err
		}
	}
	if request.CurrentHeight == 0 {
		return errors.New("networking DNL recursive lookup current height must be positive")
	}
	if request.MaxDepth == 0 || request.MaxDepth > MaxDNLRecursiveDepth {
		return fmt.Errorf("networking DNL recursive lookup depth must be between 1 and %d", MaxDNLRecursiveDepth)
	}
	if request.Limit == 0 || request.Limit > MaxDNLQueryLimit {
		return fmt.Errorf("networking DNL recursive lookup limit must be between 1 and %d", MaxDNLQueryLimit)
	}
	return nil
}

func (hop DNLLookupHop) Validate() error {
	hop = NormalizeDNLLookupHop(hop)
	if hop.LookupKey == "" {
		return errors.New("networking DNL lookup hop key is required")
	}
	if err := ValidateHash("networking DNL lookup hop node", hop.NodeID); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL lookup hop service id", []string{hop.ServiceID}, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL lookup hop zone id", []string{hop.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL lookup hop route", hop.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL lookup hop proof", hop.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL lookup hop hash", hop.HopHash); err != nil {
		return err
	}
	if hop.HopHash != ComputeDNLLookupHopHash(hop) {
		return errors.New("networking DNL lookup hop hash mismatch")
	}
	return nil
}

func (response DNLRecursiveLookupResponse) Validate() error {
	if err := ValidateHash("networking DNL recursive request hash", response.RequestHash); err != nil {
		return err
	}
	if response.ExpiryHeight == 0 {
		return errors.New("networking DNL recursive response expiry height must be positive")
	}
	if len(response.Proofs) == 0 {
		return errors.New("networking DNL recursive response requires proofs")
	}
	for _, hop := range response.Hops {
		if err := hop.Validate(); err != nil {
			return err
		}
	}
	if err := validateDNLServiceEntries(response.Entries); err != nil {
		return err
	}
	if err := validateDNLRoutes(normalizeDNLRoutes(response.Routes), response.Entries); err != nil {
		return err
	}
	for _, proof := range response.Proofs {
		if err := proof.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking DNL recursive response hash", response.ResponseHash); err != nil {
		return err
	}
	if response.ResponseHash != ComputeDNLRecursiveLookupResponseHash(response) {
		return errors.New("networking DNL recursive response hash mismatch")
	}
	return nil
}

func (snapshot DNLRoutingMetricSnapshot) Validate() error {
	snapshot = NormalizeDNLRoutingMetricSnapshot(snapshot)
	if err := ValidateHash("networking DNL metric snapshot id", snapshot.SnapshotID); err != nil {
		return err
	}
	if snapshot.SnapshotID != ComputeDNLRoutingMetricSnapshotID(snapshot) {
		return errors.New("networking DNL metric snapshot id mismatch")
	}
	if err := ValidateHash("networking DNL metric route id", snapshot.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL metric node id", snapshot.NodeID); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL metric zone id", []string{snapshot.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("DNL metric service id", []string{snapshot.ServiceID}, MaxServiceIDBytes); err != nil {
		return err
	}
	if snapshot.LatencyMillis == 0 {
		return errors.New("networking DNL metric latency must be positive")
	}
	if snapshot.ReliabilityScoreBps > BasisPoints || snapshot.CongestionWeightBps > BasisPoints {
		return fmt.Errorf("networking DNL metric bps must be <= %d", BasisPoints)
	}
	if snapshot.SnapshotHeight == 0 {
		return errors.New("networking DNL metric snapshot height must be positive")
	}
	if !snapshot.ZoneSupport || !snapshot.ServiceSupport {
		return errors.New("networking DNL metric snapshot requires zone and service support")
	}
	if err := ValidateHash("networking DNL metric snapshot hash", snapshot.SnapshotHash); err != nil {
		return err
	}
	if snapshot.SnapshotHash != ComputeDNLRoutingMetricSnapshotHash(snapshot) {
		return errors.New("networking DNL metric snapshot hash mismatch")
	}
	return nil
}

func (hint DNLLiveRoutingHint) Validate() error {
	hint = NormalizeDNLLiveRoutingHint(hint)
	if err := ValidateHash("networking DNL live hint route id", hint.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL live hint node id", hint.NodeID); err != nil {
		return err
	}
	if hint.LatencyMillis == 0 {
		return errors.New("networking DNL live hint latency must be positive")
	}
	if hint.ObservedHeight == 0 {
		return errors.New("networking DNL live hint observed height must be positive")
	}
	if err := ValidateHash("networking DNL live hint hash", hint.HintHash); err != nil {
		return err
	}
	if hint.HintHash != ComputeDNLLiveRoutingHintHash(hint) {
		return errors.New("networking DNL live hint hash mismatch")
	}
	return nil
}

func (selection DNLRouteSelection) Validate(consensus bool) error {
	if err := selection.Route.Validate(); err != nil {
		return err
	}
	if err := selection.Metric.Validate(); err != nil {
		return err
	}
	if selection.Score == 0 {
		return errors.New("networking DNL route selection score must be positive")
	}
	if consensus && (!selection.UsedCommittedTable || selection.UsedLiveHint || !selection.Metric.Committed) {
		return errors.New("networking DNL consensus route selection requires committed table and metrics")
	}
	if err := ValidateHash("networking DNL route selection hash", selection.SelectionHash); err != nil {
		return err
	}
	if selection.SelectionHash != ComputeDNLRouteSelectionHash(selection) {
		return errors.New("networking DNL route selection hash mismatch")
	}
	return nil
}

func ComputeDNLRecursiveLookupRequestHash(request DNLRecursiveLookupRequest) string {
	request = NormalizeDNLRecursiveLookupRequest(request)
	return HashParts("dnl-recursive-lookup-request", request.ServiceID, request.ZoneID, request.InterfaceHash, fmt.Sprintf("%d", request.CurrentHeight), fmt.Sprintf("%d", request.MaxDepth), fmt.Sprintf("%d", request.Limit), fmt.Sprintf("%t", request.ConsensusRoutingOnly), fmt.Sprintf("%t", request.AllowNodeLocalHints))
}

func ComputeDNLLookupHopHash(hop DNLLookupHop) string {
	hop = NormalizeDNLLookupHop(hop)
	return HashParts("dnl-recursive-lookup-hop", fmt.Sprintf("%d", hop.Depth), hop.LookupKey, hop.NodeID, hop.ServiceID, hop.ZoneID, hop.RouteID, hop.ProofHash)
}

func ComputeDNLRecursiveLookupResponseHash(response DNLRecursiveLookupResponse) string {
	entries := normalizeDNLServiceEntries(response.Entries)
	routes := normalizeDNLRoutes(response.Routes)
	hops := normalizeDNLLookupHops(response.Hops)
	proofs := normalizeDNLProofs(response.Proofs)
	parts := []string{"dnl-recursive-lookup-response", response.RequestHash, fmt.Sprintf("%d", response.ExpiryHeight), fmt.Sprintf("%d", len(hops)), fmt.Sprintf("%d", len(entries)), fmt.Sprintf("%d", len(routes)), fmt.Sprintf("%d", len(proofs))}
	for _, hop := range hops {
		parts = append(parts, hop.HopHash)
	}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	for _, route := range routes {
		parts = append(parts, route.EntryHash)
	}
	for _, proof := range proofs {
		parts = append(parts, proof.ProofHash)
	}
	return HashParts(parts...)
}

func ComputeDNLRoutingMetricSnapshotID(snapshot DNLRoutingMetricSnapshot) string {
	snapshot = NormalizeDNLRoutingMetricSnapshot(snapshot)
	return HashParts("dnl-routing-metric-snapshot-id", snapshot.RouteID, snapshot.NodeID, snapshot.ZoneID, snapshot.ServiceID, fmt.Sprintf("%d", snapshot.SnapshotHeight))
}

func ComputeDNLRoutingMetricSnapshotHash(snapshot DNLRoutingMetricSnapshot) string {
	snapshot = NormalizeDNLRoutingMetricSnapshot(snapshot)
	return HashParts("dnl-routing-metric-snapshot", snapshot.SnapshotID, snapshot.RouteID, snapshot.NodeID, snapshot.ZoneID, snapshot.ServiceID, fmt.Sprintf("%d", snapshot.LatencyMillis), fmt.Sprintf("%d", snapshot.GasCost), fmt.Sprintf("%d", snapshot.ReliabilityScoreBps), fmt.Sprintf("%d", snapshot.CongestionWeightBps), fmt.Sprintf("%t", snapshot.ZoneSupport), fmt.Sprintf("%t", snapshot.ServiceSupport), fmt.Sprintf("%t", snapshot.Committed), fmt.Sprintf("%d", snapshot.SnapshotHeight))
}

func ComputeDNLLiveRoutingHintHash(hint DNLLiveRoutingHint) string {
	hint = NormalizeDNLLiveRoutingHint(hint)
	return HashParts("dnl-live-routing-hint", hint.RouteID, hint.NodeID, fmt.Sprintf("%d", hint.LatencyMillis), fmt.Sprintf("%d", hint.ObservedHeight))
}

func ComputeDNLRouteMetricScore(route DNLRoutingTableEntry, metric DNLRoutingMetricSnapshot) uint64 {
	reliability := uint64(metric.ReliabilityScoreBps) * 10_000
	routeWeight := uint64(route.WeightBps) * 100
	latencyPenalty := metric.LatencyMillis * 10
	gasPenalty := metric.GasCost
	congestionPenalty := uint64(metric.CongestionWeightBps) * 100
	priorityPenalty := uint64(route.Priority) * 1_000
	score := reliability + routeWeight
	penalty := latencyPenalty + gasPenalty + congestionPenalty + priorityPenalty
	if score > penalty {
		return score - penalty
	}
	return 1
}

func ComputeDNLRouteSelectionHash(selection DNLRouteSelection) string {
	return HashParts("dnl-route-selection", selection.Route.RouteID, selection.Metric.SnapshotHash, fmt.Sprintf("%d", selection.Score), fmt.Sprintf("%t", selection.UsedCommittedTable), fmt.Sprintf("%t", selection.UsedLiveHint))
}

func normalizeDNLLookupHops(hops []DNLLookupHop) []DNLLookupHop {
	out := make([]DNLLookupHop, len(hops))
	for i, hop := range hops {
		out[i] = NormalizeDNLLookupHop(hop)
	}
	sortDNLLookupHops(out)
	return out
}

func sortDNLLookupHops(hops []DNLLookupHop) {
	sort.SliceStable(hops, func(i, j int) bool {
		if hops[i].Depth != hops[j].Depth {
			return hops[i].Depth < hops[j].Depth
		}
		if hops[i].ZoneID != hops[j].ZoneID {
			return hops[i].ZoneID < hops[j].ZoneID
		}
		if hops[i].ServiceID != hops[j].ServiceID {
			return hops[i].ServiceID < hops[j].ServiceID
		}
		return hops[i].RouteID < hops[j].RouteID
	})
}

func normalizeDNLProofs(proofs []DNLProof) []DNLProof {
	out := append([]DNLProof(nil), proofs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ProofHash < out[j].ProofHash })
	return out
}

func normalizeDNLMetricSnapshots(snapshots []DNLRoutingMetricSnapshot) []DNLRoutingMetricSnapshot {
	out := make([]DNLRoutingMetricSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		out[i] = NormalizeDNLRoutingMetricSnapshot(snapshot)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RouteID != out[j].RouteID {
			return out[i].RouteID < out[j].RouteID
		}
		return out[i].SnapshotID < out[j].SnapshotID
	})
	return out
}
