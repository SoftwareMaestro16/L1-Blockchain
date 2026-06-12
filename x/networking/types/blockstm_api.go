package types

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingAPIEndpoint string

const (
	APIEndpointNodeNetworkingQuery		NetworkingAPIEndpoint	= "grpc_node_networking_query"
	APIEndpointOverlayDiagnostics		NetworkingAPIEndpoint	= "rest_overlay_diagnostics"
	APIEndpointStreamDiagnostics		NetworkingAPIEndpoint	= "rpc_stream_diagnostics"
	APIEndpointDiscoveryProof		NetworkingAPIEndpoint	= "grpc_discovery_proof"
	APIEndpointRouteHint			NetworkingAPIEndpoint	= "rest_route_hint"
	APIEndpointStateSyncStreamEndpoint	NetworkingAPIEndpoint	= "state_sync_stream_endpoint"
)

type BlockSTMNetworkAssistInput struct {
	Height				uint64
	Schedules			[]ExecutionMessageSchedule
	Hints				[]ANAProposalHint
	CrossZoneMessages		[]ExecutionZoneMessage
	DecidesCommittedConflicts	bool
}

type BlockSTMNetworkGroup struct {
	GroupID			string
	ZoneID			string
	ShardID			string
	BlockSTMGroupID		string
	ExecutionOverlayID	string
	RouteHint		RouteHint
	Priority		uint32
	TransactionIDs		[]string
	MessageIDs		[]string
	ExecutionQueueID	string
}

type BlockSTMCrossZoneQueueDelivery struct {
	DeliveryID		string
	QueueID			string
	SourceZone		string
	DestinationZone		string
	SourceSequence		uint64
	MessageID		string
	MessageHash		string
	ExecutionOverlayID	string
}

type BlockSTMNetworkAssistPlan struct {
	Height					uint64
	Groups					[]BlockSTMNetworkGroup
	CrossZoneDeliveries			[]BlockSTMCrossZoneQueueDelivery
	PrioritizesExecutionOverlayTraffic	bool
	PropagatesZoneShardRouteHints		bool
	DeliversCrossZoneExecutionQueues	bool
	DecidesCommittedConflicts		bool
}

type NetworkingQueryServiceDescriptor struct {
	Endpoints		[]NetworkingAPIEndpoint
	PreserveGRPC		bool
	PreserveREST		bool
	PreserveRPC		bool
	ProofAttachedDiscovery	bool
	StateSyncStreamEndpoint	bool
	RouteHintEndpoint	bool
}

type NodeNetworkingQueryRequest struct {
	CurrentHeight	uint64
	NetworkSalt	[]byte
	Role		NodeRole
	ZoneID		string
	ServiceID	string
	IncludeExpired	bool
}

type NodeNetworkingQueryResponse struct {
	Nodes		[]NodeRecord
	PeerCountByRole	[]PeerRoleCountMetric
	ResultHash	string
}

type OverlayDiagnosticsRequest struct {
	OverlayID	string
	CurrentHeight	uint64
}

type OverlayDiagnosticsResponse struct {
	OverlayID		string
	OverlayType		OverlayType
	MembershipSize		uint64
	GraphHash		string
	QueuedMessages		uint64
	RouteFailureRateBps	uint32
	ResultHash		string
}

type StreamDiagnosticsRequest struct {
	StreamID	string
	PayloadType	StreamingPayloadType
}

type StreamDiagnosticsResponse struct {
	StreamID		string
	SessionID		string
	PayloadType		StreamingPayloadType
	State			StreamSessionState
	ThroughputBytesBps	uint64
	BytesSent		uint64
	BytesAcknowledged	uint64
	StallCount		uint64
	BackpressureActive	bool
	CompletionBps		uint32
	ResultHash		string
}

type DiscoveryProofAPIRequest struct {
	Query		DRTQuery
	CurrentHeight	uint64
	OnChainProof	DiscoveryOnChainProof
}

type DiscoveryProofAPIResponse struct {
	Response	DiscoveryResponse
	ResultHash	string
}

type RouteHintAPIRequest struct {
	Message		NetworkMessage
	Descriptor	OverlayDescriptor
	Graph		RoutingGraph
	RequestedHint	RouteHint
	ClientZoneID	string
	ClientShardID	string
	ServiceID	string
	StorageKeyHash	string
}

type RouteHintAPIResponse struct {
	OverlayID		string
	OverlayType		OverlayType
	Hint			RouteHint
	Channel			ChannelClass
	AdvisoryOnly		bool
	DeterministicHintHash	string
	ResultHash		string
}

func BuildBlockSTMNetworkAssistPlan(input BlockSTMNetworkAssistInput) (BlockSTMNetworkAssistPlan, error) {
	if input.Height == 0 {
		return BlockSTMNetworkAssistPlan{}, errors.New("networking BlockSTM assist height must be positive")
	}
	if input.DecidesCommittedConflicts {
		return BlockSTMNetworkAssistPlan{}, errors.New("networking cannot decide committed BlockSTM conflicts")
	}
	hintsBySchedule, err := normalizeAndIndexANAProposalHints(input.Hints)
	if err != nil {
		return BlockSTMNetworkAssistPlan{}, err
	}
	schedules, err := normalizeCommittedExecutionSchedules(input.Schedules)
	if err != nil {
		return BlockSTMNetworkAssistPlan{}, err
	}
	groups := make([]BlockSTMNetworkGroup, 0, len(schedules))
	for _, schedule := range schedules {
		group := blockSTMGroupFromSchedule(schedule, hintsBySchedule[schedule.ScheduleID])
		groups = append(groups, group)
	}
	sortBlockSTMGroups(groups)
	deliveries, err := buildBlockSTMCrossZoneDeliveries(input.CrossZoneMessages)
	if err != nil {
		return BlockSTMNetworkAssistPlan{}, err
	}
	plan := BlockSTMNetworkAssistPlan{
		Height:					input.Height,
		Groups:					groups,
		CrossZoneDeliveries:			deliveries,
		PrioritizesExecutionOverlayTraffic:	true,
		PropagatesZoneShardRouteHints:		len(groups) > 0,
		DeliversCrossZoneExecutionQueues:	len(deliveries) > 0,
	}
	if err := plan.Validate(); err != nil {
		return BlockSTMNetworkAssistPlan{}, err
	}
	return plan, nil
}

func (p BlockSTMNetworkAssistPlan) Validate() error {
	if p.Height == 0 {
		return errors.New("networking BlockSTM assist plan height must be positive")
	}
	if p.DecidesCommittedConflicts {
		return errors.New("networking BlockSTM assist must not decide committed conflicts")
	}
	if !p.PrioritizesExecutionOverlayTraffic {
		return errors.New("networking BlockSTM assist must prioritize execution overlay traffic")
	}
	if len(p.Groups) == 0 {
		return errors.New("networking BlockSTM assist requires execution groups")
	}
	for _, group := range p.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
	}
	for _, delivery := range p.CrossZoneDeliveries {
		if err := delivery.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (g BlockSTMNetworkGroup) Validate() error {
	group := normalizeBlockSTMNetworkGroup(g)
	if err := ValidateHash("networking BlockSTM assist group id", group.GroupID); err != nil {
		return err
	}
	if group.GroupID != ComputeBlockSTMNetworkGroupID(group) {
		return errors.New("networking BlockSTM assist group id mismatch")
	}
	if group.ZoneID == "" || group.ShardID == "" {
		return errors.New("networking BlockSTM assist group requires zone and shard")
	}
	if err := ValidateHash("networking BlockSTM group id", group.BlockSTMGroupID); err != nil {
		return err
	}
	if err := ValidateHash("networking BlockSTM execution overlay id", group.ExecutionOverlayID); err != nil {
		return err
	}
	if err := ValidateHash("networking BlockSTM execution queue id", group.ExecutionQueueID); err != nil {
		return err
	}
	if group.Priority != PriorityForChannel(ChannelExecution) {
		return errors.New("networking BlockSTM assist must use execution channel priority")
	}
	if group.RouteHint.ZoneID != group.ZoneID || group.RouteHint.ShardID != group.ShardID {
		return errors.New("networking BlockSTM route hint must target group zone and shard")
	}
	if len(group.TransactionIDs) == 0 && len(group.MessageIDs) == 0 {
		return errors.New("networking BlockSTM assist group requires transactions or messages")
	}
	return nil
}

func (d BlockSTMCrossZoneQueueDelivery) Validate() error {
	delivery := normalizeBlockSTMCrossZoneDelivery(d)
	if err := ValidateHash("networking BlockSTM cross-zone delivery id", delivery.DeliveryID); err != nil {
		return err
	}
	if delivery.DeliveryID != ComputeBlockSTMCrossZoneDeliveryID(delivery) {
		return errors.New("networking BlockSTM cross-zone delivery id mismatch")
	}
	if err := ValidateHash("networking BlockSTM cross-zone queue id", delivery.QueueID); err != nil {
		return err
	}
	if delivery.SourceZone == "" || delivery.DestinationZone == "" || delivery.SourceZone == delivery.DestinationZone {
		return errors.New("networking BlockSTM cross-zone delivery requires distinct zones")
	}
	if delivery.SourceSequence == 0 {
		return errors.New("networking BlockSTM cross-zone delivery requires sequence")
	}
	if err := ValidateHash("networking BlockSTM cross-zone message id", delivery.MessageID); err != nil {
		return err
	}
	if err := ValidateHash("networking BlockSTM cross-zone message hash", delivery.MessageHash); err != nil {
		return err
	}
	if err := ValidateHash("networking BlockSTM cross-zone execution overlay id", delivery.ExecutionOverlayID); err != nil {
		return err
	}
	return nil
}

func DefaultNetworkingQueryServiceDescriptor() NetworkingQueryServiceDescriptor {
	return NetworkingQueryServiceDescriptor{
		Endpoints: []NetworkingAPIEndpoint{
			APIEndpointNodeNetworkingQuery,
			APIEndpointOverlayDiagnostics,
			APIEndpointStreamDiagnostics,
			APIEndpointDiscoveryProof,
			APIEndpointRouteHint,
			APIEndpointStateSyncStreamEndpoint,
		},
		PreserveGRPC:			true,
		PreserveREST:			true,
		PreserveRPC:			true,
		ProofAttachedDiscovery:		true,
		StateSyncStreamEndpoint:	true,
		RouteHintEndpoint:		true,
	}
}

func (d NetworkingQueryServiceDescriptor) Validate() error {
	required := []NetworkingAPIEndpoint{
		APIEndpointNodeNetworkingQuery,
		APIEndpointOverlayDiagnostics,
		APIEndpointStreamDiagnostics,
		APIEndpointDiscoveryProof,
		APIEndpointRouteHint,
		APIEndpointStateSyncStreamEndpoint,
	}
	for _, endpoint := range required {
		if !hasNetworkingAPIEndpoint(d.Endpoints, endpoint) {
			return fmt.Errorf("networking API descriptor missing endpoint %s", endpoint)
		}
	}
	if !d.PreserveGRPC || !d.PreserveREST || !d.PreserveRPC {
		return errors.New("networking API descriptor must preserve gRPC, REST, and RPC APIs")
	}
	if !d.ProofAttachedDiscovery {
		return errors.New("networking API descriptor must expose proof-attached discovery")
	}
	if !d.StateSyncStreamEndpoint {
		return errors.New("networking API descriptor must expose state sync stream endpoints")
	}
	if !d.RouteHintEndpoint {
		return errors.New("networking API descriptor must expose client route hints")
	}
	return nil
}

func BuildNodeNetworkingQueryResponse(req NodeNetworkingQueryRequest, records []NodeRecord) (NodeNetworkingQueryResponse, error) {
	if req.CurrentHeight == 0 {
		return NodeNetworkingQueryResponse{}, errors.New("networking node query height must be positive")
	}
	if len(req.NetworkSalt) == 0 {
		return NodeNetworkingQueryResponse{}, errors.New("networking node query requires network salt")
	}
	req.ZoneID = strings.TrimSpace(req.ZoneID)
	req.ServiceID = strings.TrimSpace(req.ServiceID)
	out := make([]NodeRecord, 0, len(records))
	for _, record := range records {
		record = NormalizeNodeRecord(record)
		validateHeight := req.CurrentHeight
		if req.IncludeExpired {
			validateHeight = 0
		}
		if err := record.Validate(req.NetworkSalt, validateHeight); err != nil {
			if req.IncludeExpired && strings.Contains(err.Error(), "expired") {

			} else {
				return NodeNetworkingQueryResponse{}, err
			}
		}
		if req.Role != "" && !hasRole(record.Roles, req.Role) {
			continue
		}
		if req.ZoneID != "" && !containsString(record.ZonesSupported, req.ZoneID) {
			continue
		}
		if req.ServiceID != "" && !containsString(record.ServicesSupported, req.ServiceID) {
			continue
		}
		out = append(out, record)
	}
	sortNodeRecords(out)
	peerCounts, err := ComputePeerCountByRole(out)
	if err != nil {
		return NodeNetworkingQueryResponse{}, err
	}
	response := NodeNetworkingQueryResponse{Nodes: out, PeerCountByRole: peerCounts}
	response.ResultHash = ComputeNodeNetworkingQueryResultHash(response)
	return response, nil
}

func BuildOverlayDiagnosticsResponse(req OverlayDiagnosticsRequest, descriptors []OverlayDescriptor, memberships []OverlayMembershipRecord, graph RoutingGraph, metrics []L3OverlayMetrics, failures []RouteFailureSample) (OverlayDiagnosticsResponse, error) {
	overlayID := normalizeHashText(req.OverlayID)
	if err := ValidateHash("networking overlay diagnostics overlay id", overlayID); err != nil {
		return OverlayDiagnosticsResponse{}, err
	}
	if req.CurrentHeight == 0 {
		return OverlayDiagnosticsResponse{}, errors.New("networking overlay diagnostics height must be positive")
	}
	desc, found, err := findOverlayDescriptor(overlayID, descriptors)
	if err != nil {
		return OverlayDiagnosticsResponse{}, err
	}
	if !found {
		return OverlayDiagnosticsResponse{}, errors.New("networking overlay diagnostics descriptor not found")
	}
	graph = NormalizeRoutingGraph(graph)
	if graph.OverlayID == "" {
		graph.OverlayID = overlayID
		graph.GraphHash = ComputeRoutingGraphHash(graph)
	}
	if err := graph.Validate(desc); err != nil {
		return OverlayDiagnosticsResponse{}, err
	}
	membershipSize := uint64(0)
	for _, membership := range memberships {
		membership.OverlayID = normalizeHashText(membership.OverlayID)
		if membership.OverlayID == overlayID && (membership.ExpiresHeight == 0 || membership.ExpiresHeight >= req.CurrentHeight) {
			membershipSize++
		}
	}
	queued := uint64(0)
	for _, metric := range metrics {
		if normalizeHashText(metric.OverlayID) == overlayID {
			queued += metric.QueuedCount
		}
	}
	filteredFailures := make([]RouteFailureSample, 0, len(failures))
	for _, sample := range failures {
		if normalizeHashText(sample.OverlayID) == overlayID {
			filteredFailures = append(filteredFailures, sample)
		}
	}
	failureRate, err := ComputeRouteFailureRate(filteredFailures)
	if err != nil {
		return OverlayDiagnosticsResponse{}, err
	}
	response := OverlayDiagnosticsResponse{
		OverlayID:		overlayID,
		OverlayType:		desc.OverlayType,
		MembershipSize:		membershipSize,
		GraphHash:		graph.GraphHash,
		QueuedMessages:		queued,
		RouteFailureRateBps:	failureRate,
	}
	response.ResultHash = ComputeOverlayDiagnosticsResultHash(response)
	return response, nil
}

func BuildStreamDiagnosticsResponse(req StreamDiagnosticsRequest, sessions []StreamSession, metrics []StreamMetrics) (StreamDiagnosticsResponse, error) {
	streamID := normalizeHashText(req.StreamID)
	if err := ValidateHash("networking stream diagnostics stream id", streamID); err != nil {
		return StreamDiagnosticsResponse{}, err
	}
	for _, session := range sessions {
		session = session.Normalize()
		if session.StreamID != streamID {
			continue
		}
		if err := session.Validate(); err != nil {
			return StreamDiagnosticsResponse{}, err
		}
		if req.PayloadType != "" && session.PayloadType != req.PayloadType {
			return StreamDiagnosticsResponse{}, errors.New("networking stream diagnostics payload type mismatch")
		}
		metric := StreamMetrics{
			StreamID:		streamID,
			PayloadType:		session.PayloadType,
			State:			session.State,
			BytesSent:		session.BytesSent,
			BytesAcknowledged:	session.BytesAcknowledged,
			InFlightBytes:		session.BytesSent - session.BytesAcknowledged,
			AvailableWindow:	StreamAvailableWindow(session),
			CompletionBps:		uint32(0),
		}
		for _, candidate := range metrics {
			if normalizeHashText(candidate.StreamID) == streamID {
				if err := validateStreamMetrics(candidate); err != nil {
					return StreamDiagnosticsResponse{}, err
				}
				metric = candidate
				break
			}
		}
		response := StreamDiagnosticsResponse{
			StreamID:		streamID,
			SessionID:		session.SessionID,
			PayloadType:		metric.PayloadType,
			State:			metric.State,
			ThroughputBytesBps:	metric.ThroughputBytesBps,
			BytesSent:		metric.BytesSent,
			BytesAcknowledged:	metric.BytesAcknowledged,
			StallCount:		metric.StallCount,
			BackpressureActive:	metric.BackpressureActive,
			CompletionBps:		metric.CompletionBps,
		}
		response.ResultHash = ComputeStreamDiagnosticsResultHash(response)
		return response, nil
	}
	return StreamDiagnosticsResponse{}, errors.New("networking stream diagnostics stream not found")
}

func BuildDiscoveryProofAPIResponse(req DiscoveryProofAPIRequest, table DistributedRoutingTable, source NodeRecord, sourcePrivateKey ed25519.PrivateKey, networkSalt []byte) (DiscoveryProofAPIResponse, error) {
	if req.CurrentHeight == 0 {
		return DiscoveryProofAPIResponse{}, errors.New("networking discovery proof API height must be positive")
	}
	query := normalizeDRTQuery(req.Query)
	query.CurrentHeight = req.CurrentHeight
	response, err := BuildDiscoveryResponse(table, query, source, sourcePrivateKey, networkSalt, req.OnChainProof, req.CurrentHeight)
	if err != nil {
		return DiscoveryProofAPIResponse{}, err
	}
	sourcePubKey, ok := sourcePrivateKey.Public().(ed25519.PublicKey)
	if !ok {
		return DiscoveryProofAPIResponse{}, errors.New("networking discovery proof API source public key must be ed25519")
	}
	if err := response.Validate(sourcePubKey, networkSalt, req.CurrentHeight); err != nil {
		return DiscoveryProofAPIResponse{}, err
	}
	return DiscoveryProofAPIResponse{Response: response, ResultHash: response.ResultHash}, nil
}

func BuildRouteHintAPIResponse(req RouteHintAPIRequest) (RouteHintAPIResponse, error) {
	msg := req.Message.Normalize()
	if msg.ReplaySafeID == "" {
		msg.ReplaySafeID = ComputeNetworkMessageID(msg)
	}
	if err := msg.ValidateHardRules(); err != nil {
		return RouteHintAPIResponse{}, err
	}
	desc := NormalizeOverlayDescriptor(req.Descriptor)
	if err := desc.ValidateBasic(); err != nil {
		return RouteHintAPIResponse{}, err
	}
	graph := NormalizeRoutingGraph(req.Graph)
	if graph.OverlayID == "" {
		graph.OverlayID = desc.OverlayID
		graph.GraphHash = ComputeRoutingGraphHash(graph)
	}
	if err := graph.Validate(desc); err != nil {
		return RouteHintAPIResponse{}, err
	}
	hint := RouteHint{
		ZoneID:		strings.TrimSpace(req.ClientZoneID),
		ShardID:	strings.TrimSpace(req.ClientShardID),
		ServiceID:	strings.TrimSpace(req.ServiceID),
		StorageKeyHash:	normalizeHashText(req.StorageKeyHash),
	}
	if hint.ZoneID == "" {
		hint.ZoneID = strings.TrimSpace(req.RequestedHint.ZoneID)
	}
	if hint.ShardID == "" {
		hint.ShardID = strings.TrimSpace(req.RequestedHint.ShardID)
	}
	if hint.ServiceID == "" {
		hint.ServiceID = strings.TrimSpace(req.RequestedHint.ServiceID)
	}
	if hint.StorageKeyHash == "" {
		hint.StorageKeyHash = normalizeHashText(req.RequestedHint.StorageKeyHash)
	}
	if graph.DeterministicHintHash != "" {
		hint.DeterministicHintHash = graph.DeterministicHintHash
	}
	if hint.StorageKeyHash != "" {
		if err := ValidateHash("networking route hint API storage key hash", hint.StorageKeyHash); err != nil {
			return RouteHintAPIResponse{}, err
		}
	}
	response := RouteHintAPIResponse{
		OverlayID:		desc.OverlayID,
		OverlayType:		desc.OverlayType,
		Hint:			hint,
		Channel:		msg.Channel,
		AdvisoryOnly:		!msg.ConsensusEffect,
		DeterministicHintHash:	hint.DeterministicHintHash,
	}
	response.ResultHash = ComputeRouteHintAPIResultHash(response)
	return response, nil
}

func ComputeBlockSTMNetworkGroupID(group BlockSTMNetworkGroup) string {
	group = normalizeBlockSTMNetworkGroup(group)
	parts := []string{
		"blockstm-network-group",
		group.ZoneID,
		group.ShardID,
		group.BlockSTMGroupID,
		group.ExecutionOverlayID,
		group.ExecutionQueueID,
	}
	parts = append(parts, group.TransactionIDs...)
	parts = append(parts, group.MessageIDs...)
	return HashParts(parts...)
}

func ComputeBlockSTMCrossZoneDeliveryID(delivery BlockSTMCrossZoneQueueDelivery) string {
	delivery = normalizeBlockSTMCrossZoneDelivery(delivery)
	return HashParts(
		"blockstm-cross-zone-delivery",
		delivery.QueueID,
		delivery.SourceZone,
		delivery.DestinationZone,
		fmt.Sprintf("%d", delivery.SourceSequence),
		delivery.MessageID,
		delivery.MessageHash,
		delivery.ExecutionOverlayID,
	)
}

func ComputeNodeNetworkingQueryResultHash(response NodeNetworkingQueryResponse) string {
	parts := []string{"node-networking-query-result"}
	for _, record := range response.Nodes {
		parts = append(parts, normalizeHashText(record.NodeID))
	}
	for _, count := range response.PeerCountByRole {
		parts = append(parts, string(count.Role), fmt.Sprintf("%d", count.Count))
	}
	return HashParts(parts...)
}

func ComputeOverlayDiagnosticsResultHash(response OverlayDiagnosticsResponse) string {
	return HashParts(
		"overlay-diagnostics-result",
		response.OverlayID,
		string(response.OverlayType),
		fmt.Sprintf("%d", response.MembershipSize),
		response.GraphHash,
		fmt.Sprintf("%d", response.QueuedMessages),
		fmt.Sprintf("%d", response.RouteFailureRateBps),
	)
}

func ComputeStreamDiagnosticsResultHash(response StreamDiagnosticsResponse) string {
	return HashParts(
		"stream-diagnostics-result",
		response.StreamID,
		response.SessionID,
		string(response.PayloadType),
		string(response.State),
		fmt.Sprintf("%d", response.ThroughputBytesBps),
		fmt.Sprintf("%d", response.BytesSent),
		fmt.Sprintf("%d", response.BytesAcknowledged),
		fmt.Sprintf("%d", response.StallCount),
		fmt.Sprintf("%t", response.BackpressureActive),
		fmt.Sprintf("%d", response.CompletionBps),
	)
}

func ComputeRouteHintAPIResultHash(response RouteHintAPIResponse) string {
	return HashParts(
		"route-hint-api-result",
		response.OverlayID,
		string(response.OverlayType),
		string(response.Channel),
		response.Hint.ZoneID,
		response.Hint.ShardID,
		response.Hint.ServiceID,
		response.Hint.StorageKeyHash,
		response.Hint.DeterministicHintHash,
		fmt.Sprintf("%t", response.AdvisoryOnly),
	)
}

func blockSTMGroupFromSchedule(schedule ExecutionMessageSchedule, hints []ANAProposalHint) BlockSTMNetworkGroup {
	group := BlockSTMNetworkGroup{
		ZoneID:			schedule.ZoneID,
		ShardID:		schedule.ShardID,
		BlockSTMGroupID:	HashParts("blockstm-network-group-id", schedule.ZoneID, schedule.ShardID, schedule.ScheduleID),
		ExecutionOverlayID:	HashParts("blockstm-execution-overlay", schedule.ZoneID, schedule.ShardID),
		Priority:		PriorityForChannel(ChannelExecution),
		TransactionIDs:		append([]string(nil), schedule.TransactionIDs...),
		MessageIDs:		append([]string(nil), schedule.MessageIDs...),
		ExecutionQueueID:	HashParts("blockstm-execution-queue", schedule.ZoneID, schedule.ShardID),
		RouteHint: RouteHint{
			ZoneID:		schedule.ZoneID,
			ShardID:	schedule.ShardID,
		},
	}
	for _, hint := range hints {
		if hint.BlockSTMGroupID != "" {
			group.BlockSTMGroupID = hint.BlockSTMGroupID
		}
		if hint.DeterministicHintProof != "" {
			group.RouteHint.DeterministicHintHash = hint.DeterministicHintProof
		}
	}
	group = normalizeBlockSTMNetworkGroup(group)
	group.GroupID = ComputeBlockSTMNetworkGroupID(group)
	return group
}

func buildBlockSTMCrossZoneDeliveries(messages []ExecutionZoneMessage) ([]BlockSTMCrossZoneQueueDelivery, error) {
	deliveries := make([]BlockSTMCrossZoneQueueDelivery, 0, len(messages))
	for _, msg := range messages {
		msg = NormalizeExecutionZoneMessage(msg)
		if msg.Message.Type != MeshMessageCrossZone {
			continue
		}
		if err := msg.Message.ValidateBasic(0); err != nil {
			return nil, err
		}
		if err := msg.CrossZone.Validate(0); err != nil {
			return nil, err
		}
		if msg.CrossZone.DestinationZone != msg.Message.DestinationZone {
			return nil, errors.New("networking BlockSTM cross-zone destination mismatch")
		}
		delivery := BlockSTMCrossZoneQueueDelivery{
			QueueID:		HashParts("blockstm-cross-zone-queue", msg.CrossZone.DestinationZone, msg.ShardID),
			SourceZone:		msg.CrossZone.SourceZone,
			DestinationZone:	msg.CrossZone.DestinationZone,
			SourceSequence:		msg.CrossZone.SourceSequence,
			MessageID:		msg.Message.MessageID,
			MessageHash:		msg.CrossZone.MessageHash,
			ExecutionOverlayID:	msg.ExecutionOverlayID,
		}
		delivery.DeliveryID = ComputeBlockSTMCrossZoneDeliveryID(delivery)
		deliveries = append(deliveries, delivery)
	}
	sort.SliceStable(deliveries, func(i, j int) bool {
		if deliveries[i].DestinationZone != deliveries[j].DestinationZone {
			return deliveries[i].DestinationZone < deliveries[j].DestinationZone
		}
		if deliveries[i].SourceZone != deliveries[j].SourceZone {
			return deliveries[i].SourceZone < deliveries[j].SourceZone
		}
		return deliveries[i].SourceSequence < deliveries[j].SourceSequence
	})
	return deliveries, nil
}

func normalizeBlockSTMNetworkGroup(group BlockSTMNetworkGroup) BlockSTMNetworkGroup {
	group.GroupID = normalizeHashText(group.GroupID)
	group.ZoneID = strings.TrimSpace(group.ZoneID)
	group.ShardID = strings.TrimSpace(group.ShardID)
	group.BlockSTMGroupID = normalizeHashText(group.BlockSTMGroupID)
	group.ExecutionOverlayID = normalizeHashText(group.ExecutionOverlayID)
	group.RouteHint.ZoneID = strings.TrimSpace(group.RouteHint.ZoneID)
	group.RouteHint.ShardID = strings.TrimSpace(group.RouteHint.ShardID)
	group.RouteHint.ServiceID = strings.TrimSpace(group.RouteHint.ServiceID)
	group.RouteHint.StorageKeyHash = normalizeHashText(group.RouteHint.StorageKeyHash)
	group.RouteHint.DeterministicHintHash = normalizeHashText(group.RouteHint.DeterministicHintHash)
	group.TransactionIDs = normalizeHashSet(group.TransactionIDs)
	group.MessageIDs = normalizeHashSet(group.MessageIDs)
	group.ExecutionQueueID = normalizeHashText(group.ExecutionQueueID)
	return group
}

func normalizeBlockSTMCrossZoneDelivery(delivery BlockSTMCrossZoneQueueDelivery) BlockSTMCrossZoneQueueDelivery {
	delivery.DeliveryID = normalizeHashText(delivery.DeliveryID)
	delivery.QueueID = normalizeHashText(delivery.QueueID)
	delivery.SourceZone = strings.TrimSpace(delivery.SourceZone)
	delivery.DestinationZone = strings.TrimSpace(delivery.DestinationZone)
	delivery.MessageID = normalizeHashText(delivery.MessageID)
	delivery.MessageHash = normalizeHashText(delivery.MessageHash)
	delivery.ExecutionOverlayID = normalizeHashText(delivery.ExecutionOverlayID)
	return delivery
}

func sortBlockSTMGroups(groups []BlockSTMNetworkGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].ZoneID != groups[j].ZoneID {
			return groups[i].ZoneID < groups[j].ZoneID
		}
		if groups[i].ShardID != groups[j].ShardID {
			return groups[i].ShardID < groups[j].ShardID
		}
		return groups[i].GroupID < groups[j].GroupID
	})
}

func findOverlayDescriptor(overlayID string, descriptors []OverlayDescriptor) (OverlayDescriptor, bool, error) {
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if err := desc.ValidateBasic(); err != nil {
			return OverlayDescriptor{}, false, err
		}
		if desc.OverlayID == overlayID {
			return desc, true, nil
		}
	}
	return OverlayDescriptor{}, false, nil
}

func validateStreamMetrics(metric StreamMetrics) error {
	metric.StreamID = normalizeHashText(metric.StreamID)
	if err := ValidateHash("networking stream metrics stream id", metric.StreamID); err != nil {
		return err
	}
	if !IsStreamingPayloadType(metric.PayloadType) {
		return fmt.Errorf("unknown networking stream metrics payload type %q", metric.PayloadType)
	}
	if !IsStreamSessionState(metric.State) {
		return fmt.Errorf("unknown networking stream metrics state %q", metric.State)
	}
	if metric.BytesAcknowledged > metric.BytesSent {
		return errors.New("networking stream metrics acknowledged bytes exceed sent bytes")
	}
	if metric.CompletionBps > BasisPoints {
		return fmt.Errorf("networking stream metrics completion must be <= %d bps", BasisPoints)
	}
	return nil
}

func hasNetworkingAPIEndpoint(endpoints []NetworkingAPIEndpoint, required NetworkingAPIEndpoint) bool {
	for _, endpoint := range endpoints {
		if endpoint == required {
			return true
		}
	}
	return false
}
