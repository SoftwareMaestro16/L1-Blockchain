package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	AetherisNextTopologyVersion = uint64(1)

	topologyNodeCore              = "core/aether"
	topologyNodeFinancial         = "zone/financial"
	topologyNodeIdentity          = "zone/identity"
	topologyNodeApplication       = "zone/application"
	topologyNodeContract          = "zone/contract"
	topologyNodeFinancialShards   = "shards/financial"
	topologyNodeIdentityShards    = "shards/identity"
	topologyNodeApplicationShards = "shards/application"
	topologyNodeContractShards    = "shards/contract"

	topologyRelationSchedules = "schedules"
	topologyRelationOwns      = "owns"
	topologyRelationAsyncCall = "async_call"
)

type TopologyNodeKind string

const (
	TopologyNodeCore       TopologyNodeKind = "CORE"
	TopologyNodeZone       TopologyNodeKind = "ZONE"
	TopologyNodeShardGroup TopologyNodeKind = "SHARD_GROUP"
)

type AetherisNextTopologyNode struct {
	NodeID       string
	Kind         TopologyNodeKind
	ZoneID       ZoneID
	Label        string
	Capabilities []string
}

type AetherisNextTopologyEdge struct {
	FromNodeID string
	ToNodeID   string
	Relation   string
}

type AetherisNextTopologyPlan struct {
	Version      uint64
	Nodes        []AetherisNextTopologyNode
	Edges        []AetherisNextTopologyEdge
	TopologyHash string
}

func DefaultAetherisNextTopology() (AetherisNextTopologyPlan, error) {
	plan := AetherisNextTopologyPlan{
		Version: AetherisNextTopologyVersion,
		Nodes: []AetherisNextTopologyNode{
			{NodeID: topologyNodeCore, Kind: TopologyNodeCore, Label: "Aether Core", Capabilities: []string{"consensus", "finality", "global-root", "message-root", "proof-registry", "scheduler", "validator-set", "zone-commitments"}},
			{NodeID: topologyNodeFinancial, Kind: TopologyNodeZone, ZoneID: ZoneIDFinancial, Label: "Financial Zone", Capabilities: []string{"bank-fees", "dex-factory", "payment-settlement"}},
			{NodeID: topologyNodeIdentity, Kind: TopologyNodeZone, ZoneID: ZoneIDIdentity, Label: "Identity Zone", Capabilities: []string{".aet-resolver", "nft-ownership", "resolver-proofs"}},
			{NodeID: topologyNodeApplication, Kind: TopologyNodeZone, ZoneID: ZoneIDApplication, Label: "Application Zone", Capabilities: []string{"schedulers", "workflows"}},
			{NodeID: topologyNodeContract, Kind: TopologyNodeZone, ZoneID: ZoneIDContract, Label: "Contract Zone", Capabilities: []string{"avm-2.0", "contracts"}},
			{NodeID: topologyNodeFinancialShards, Kind: TopologyNodeShardGroup, ZoneID: ZoneIDFinancial, Label: "Financial Compute Shards", Capabilities: []string{"account-routing", "pool-routing"}},
			{NodeID: topologyNodeIdentityShards, Kind: TopologyNodeShardGroup, ZoneID: ZoneIDIdentity, Label: "Identity Compute Shards", Capabilities: []string{"name-hash-routing", "reverse-routing"}},
			{NodeID: topologyNodeApplicationShards, Kind: TopologyNodeShardGroup, ZoneID: ZoneIDApplication, Label: "Application Compute Shards", Capabilities: []string{"scheduler-buckets", "workflow-routing"}},
			{NodeID: topologyNodeContractShards, Kind: TopologyNodeShardGroup, ZoneID: ZoneIDContract, Label: "Contract Compute Shards", Capabilities: []string{"contract-address-routing", "storage-prefix-routing"}},
		},
		Edges: []AetherisNextTopologyEdge{
			{FromNodeID: topologyNodeCore, ToNodeID: topologyNodeFinancial, Relation: topologyRelationSchedules},
			{FromNodeID: topologyNodeCore, ToNodeID: topologyNodeIdentity, Relation: topologyRelationSchedules},
			{FromNodeID: topologyNodeCore, ToNodeID: topologyNodeApplication, Relation: topologyRelationSchedules},
			{FromNodeID: topologyNodeFinancial, ToNodeID: topologyNodeFinancialShards, Relation: topologyRelationOwns},
			{FromNodeID: topologyNodeIdentity, ToNodeID: topologyNodeIdentityShards, Relation: topologyRelationOwns},
			{FromNodeID: topologyNodeApplication, ToNodeID: topologyNodeApplicationShards, Relation: topologyRelationOwns},
			{FromNodeID: topologyNodeFinancialShards, ToNodeID: topologyNodeContract, Relation: topologyRelationAsyncCall},
			{FromNodeID: topologyNodeIdentityShards, ToNodeID: topologyNodeContract, Relation: topologyRelationAsyncCall},
			{FromNodeID: topologyNodeApplicationShards, ToNodeID: topologyNodeContract, Relation: topologyRelationAsyncCall},
			{FromNodeID: topologyNodeContract, ToNodeID: topologyNodeContractShards, Relation: topologyRelationOwns},
		},
	}
	return NewAetherisNextTopologyPlan(plan.Nodes, plan.Edges)
}

func NewAetherisNextTopologyPlan(nodes []AetherisNextTopologyNode, edges []AetherisNextTopologyEdge) (AetherisNextTopologyPlan, error) {
	plan := AetherisNextTopologyPlan{
		Version: AetherisNextTopologyVersion,
		Nodes:   cloneTopologyNodes(nodes),
		Edges:   append([]AetherisNextTopologyEdge(nil), edges...),
	}
	sortTopologyNodes(plan.Nodes)
	sortTopologyEdges(plan.Edges)
	if err := plan.ValidateFormat(); err != nil {
		return AetherisNextTopologyPlan{}, err
	}
	plan.TopologyHash = ComputeAetherisNextTopologyHash(plan)
	return plan, plan.ValidateHash()
}

func DefaultAetherisNextZoneDescriptors() []ZoneDescriptor {
	return []ZoneDescriptor{
		nextZoneDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial", 8, []string{"async-inbox", "async-outbox", "cross-shard-transfer", "fee-accumulator"}, []string{"account", "balance", "message", "payment", "receipt"}),
		nextZoneDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity", 4, []string{"async-inbox", "async-outbox", "identity-lookup"}, []string{"domain", "identity", "message", "receipt", "resolver"}),
		nextZoneDescriptor(ZoneIDApplication, ZoneTypeApplication, "application", 4, []string{"app-async-output", "async-inbox", "async-outbox", "deterministic-scheduler"}, []string{"app", "message", "receipt", "scheduler"}),
		nextZoneDescriptor(ZoneIDContract, ZoneTypeContract, "contract", 8, []string{"async-inbox", "async-outbox", "contract-call", "promise"}, []string{"contract", "message", "receipt", "vm"}),
	}
}

func DefaultAetherisNextShardLayouts(activationHeight uint64) ([]ShardLayout, error) {
	if activationHeight == 0 {
		return nil, errors.New("aethercore next topology activation height must be positive")
	}
	specs := []struct {
		zoneID ZoneID
		count  uint32
	}{
		{ZoneIDFinancial, 4},
		{ZoneIDIdentity, 2},
		{ZoneIDApplication, 2},
		{ZoneIDContract, 4},
	}
	layouts := make([]ShardLayout, 0, len(specs))
	for _, spec := range specs {
		shards := make([]ShardDescriptor, 0, spec.count)
		for i := uint32(0); i < spec.count; i++ {
			shardID := ShardID(fmt.Sprintf("%d", i))
			shards = append(shards, ShardDescriptor{
				ShardID:          shardID,
				StatePrefix:      fmt.Sprintf("zone/%s/shard/%s", spec.zoneID, shardID),
				ActivationHeight: activationHeight,
				ValidatorSetHash: hashParts("aetheris-next-validator-set", string(spec.zoneID), string(shardID)),
				Available:        true,
			})
		}
		layout, err := NewShardLayout(spec.zoneID, 1, activationHeight, hashParts("aetheris-next-routing-seed", string(spec.zoneID), "1"), shards)
		if err != nil {
			return nil, err
		}
		layouts = append(layouts, layout)
	}
	sortShardLayouts(layouts)
	return layouts, nil
}

func DefaultAetherisNextIdentityResolverService(createdHeight uint64) (ServiceDescriptor, error) {
	if createdHeight == 0 {
		return ServiceDescriptor{}, errors.New("aethercore next resolver service height must be positive")
	}
	interfaceID := "l1.identity.v2.Resolver"
	method := ServiceMethodDescriptor{
		MethodID:             "resolve",
		Name:                 "resolve",
		InputSchemaHash:      hashParts("aetheris-next-identity-resolver", "resolve", "input"),
		OutputSchemaHash:     hashParts("aetheris-next-identity-resolver", "resolve", "output"),
		ExecutionType:        ServiceMethodSync,
		RequiredPaymentModel: "naet-fixed",
		GasModel:             DefaultGasPolicy,
		VerificationModel:    ServiceVerificationConsensusReceipt,
		TimeoutHeightDelta:   10,
		IdempotencyRequired:  true,
		FailureBehavior:      ServiceFailureRevert,
	}
	iface := ServiceInterfaceDescriptor{
		InterfaceID:    interfaceID,
		InterfaceName:  interfaceID,
		Version:        2,
		SchemaEncoding: "json-schema-v1",
		Methods:        []ServiceMethodDescriptor{method},
		Events:         []string{"identity.resolved"},
		Errors:         []string{"identity.not_found"},
		AuthModel:      "aetheris-account",
		PaymentModel:   "naet-fixed",
		MetadataHash:   hashParts("aetheris-next-identity-resolver", "metadata"),
		CreatedHeight:  createdHeight,
	}
	iface = CanonicalServiceInterfaceDescriptor(iface)
	iface.InterfaceHash = ComputeServiceInterfaceHash(iface)
	service := ServiceDescriptor{
		ServiceID:        "identity-resolver",
		Owner:            DefaultAuthority,
		ServiceType:      ServiceTypeOnChain,
		ZoneID:           ZoneIDIdentity,
		InterfaceID:      interfaceID,
		EndpointKey:      "identity.query",
		Version:          2,
		AvailabilityHash: hashParts("aetheris-next-identity-resolver", "availability"),
		Enabled:          true,
		Status:           ServiceStatusActive,
		ExpiryHeight:     createdHeight + DefaultRootHistory,
		CreatedHeight:    createdHeight,
		UpdatedHeight:    createdHeight,
		Interface:        iface,
		Execution: ServiceExecutionDescriptor{
			Location:        ServiceLocationModule,
			Target:          "identity.query",
			ModuleRoute:     "identity",
			Mode:            ExecutionModeSync,
			Deterministic:   true,
			FailureBehavior: ServiceFailureRevert,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:       "identity-resolver",
			IdentityName:      "identity.aet",
			MetadataHash:      hashParts("aetheris-next-identity-resolver", "discovery"),
			CacheExpiryHeight: createdHeight + DefaultRootHistory - 1,
			SignaturePolicy:   "owner-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode: ServicePaymentOnChain,
			Denom:          NativeFeePolicyID,
			Amount:         "0",
			PricingUnit:    ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model:         ServiceStorageOnChain,
			StateRootType: ResolverProofRootType,
			ProofRequired: true,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel: ServiceTrustConsensusExecuted,
			Model:      ServiceVerificationConsensusReceipt,
		},
	}
	service = CanonicalServiceDescriptor(service)
	return service, service.Validate()
}

func BuildAetherisNextTopologyState(params AetherCoreParams, activationHeight uint64, routingEpoch uint64, routingHeight uint64) (CoreState, AetherisNextTopologyPlan, error) {
	if err := params.RequireEnabled(); err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	if activationHeight == 0 || routingEpoch == 0 || routingHeight == 0 {
		return CoreState{}, AetherisNextTopologyPlan{}, errors.New("aethercore next topology bootstrap heights and epochs must be positive")
	}
	plan, err := DefaultAetherisNextTopology()
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	state := EmptyState(params)
	for _, descriptor := range DefaultAetherisNextZoneDescriptors() {
		state, err = RegisterZoneDescriptor(state, descriptor)
		if err != nil {
			return CoreState{}, AetherisNextTopologyPlan{}, err
		}
	}
	service, err := DefaultAetherisNextIdentityResolverService(activationHeight)
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	state, err = RegisterServiceDescriptor(state, service)
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	layouts, err := DefaultAetherisNextShardLayouts(activationHeight)
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	for _, layout := range layouts {
		state, err = RegisterShardLayout(state, layout)
		if err != nil {
			return CoreState{}, AetherisNextTopologyPlan{}, err
		}
	}
	table, err := BuildRoutingTableCommitment(routingEpoch, routingHeight, layouts)
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	state, err = CommitRoutingTable(state, table)
	if err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	if err := ValidateAetherisNextTopologyState(state, routingHeight); err != nil {
		return CoreState{}, AetherisNextTopologyPlan{}, err
	}
	return state.Export(), plan, nil
}

func ValidateAetherisNextTopologyState(state CoreState, height uint64) error {
	if height == 0 {
		return errors.New("aethercore next topology validation height must be positive")
	}
	if err := state.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := state.Validate(); err != nil {
		return err
	}
	if _, err := DefaultAetherisNextTopology(); err != nil {
		return err
	}
	required := DefaultAetherisNextZoneDescriptors()
	layouts := make([]ShardLayout, 0, len(required))
	for _, descriptor := range required {
		actual, found := state.ZoneDescriptorByID(descriptor.ZoneID)
		if !found {
			return fmt.Errorf("aethercore next topology missing required zone %s", descriptor.ZoneID)
		}
		actual = CanonicalZoneDescriptor(actual)
		descriptor = CanonicalZoneDescriptor(descriptor)
		if !actual.Enabled {
			return fmt.Errorf("aethercore next topology zone %s is disabled", descriptor.ZoneID)
		}
		if actual.ZoneType != descriptor.ZoneType {
			return fmt.Errorf("aethercore next topology zone %s type mismatch", descriptor.ZoneID)
		}
		if actual.ModuleName != descriptor.ModuleName {
			return fmt.Errorf("aethercore next topology zone %s module mismatch", descriptor.ZoneID)
		}
		layout, found := state.LatestShardLayout(descriptor.ZoneID, height)
		if !found {
			return fmt.Errorf("aethercore next topology missing shard layout for zone %s", descriptor.ZoneID)
		}
		layouts = append(layouts, layout)
	}
	if err := ensureNativeAETResolver(state); err != nil {
		return err
	}
	table, found := state.LatestRoutingTableAtHeight(height)
	if !found {
		return errors.New("aethercore next topology requires committed routing table")
	}
	return validateRoutingTableCoversLayouts(table, layouts)
}

func (p AetherisNextTopologyPlan) ValidateFormat() error {
	if p.Version != AetherisNextTopologyVersion {
		return errors.New("aethercore next topology version mismatch")
	}
	if len(p.Nodes) == 0 {
		return errors.New("aethercore next topology requires nodes")
	}
	if len(p.Edges) == 0 {
		return errors.New("aethercore next topology requires edges")
	}
	seenNodes := make(map[string]AetherisNextTopologyNode, len(p.Nodes))
	var previousNode string
	for i, node := range p.Nodes {
		node = canonicalTopologyNode(node)
		if err := node.Validate(); err != nil {
			return err
		}
		if _, found := seenNodes[node.NodeID]; found {
			return fmt.Errorf("duplicate aethercore next topology node %s", node.NodeID)
		}
		seenNodes[node.NodeID] = node
		if i > 0 && previousNode >= node.NodeID {
			return errors.New("aethercore next topology nodes must be sorted canonically")
		}
		previousNode = node.NodeID
	}
	var previousEdge AetherisNextTopologyEdge
	for i, edge := range p.Edges {
		if err := edge.Validate(seenNodes); err != nil {
			return err
		}
		if i > 0 && compareTopologyEdges(previousEdge, edge) >= 0 {
			return errors.New("aethercore next topology edges must be sorted canonically")
		}
		previousEdge = edge
	}
	if p.TopologyHash != "" {
		return ValidateHash("aethercore next topology hash", p.TopologyHash)
	}
	return nil
}

func (p AetherisNextTopologyPlan) ValidateHash() error {
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeAetherisNextTopologyHash(p)
	if p.TopologyHash != expected {
		return fmt.Errorf("aethercore next topology hash mismatch: expected %s", expected)
	}
	return nil
}

func (n AetherisNextTopologyNode) Validate() error {
	if err := validatePolicyID("aethercore next topology node id", n.NodeID); err != nil {
		return err
	}
	if n.Kind != TopologyNodeCore && n.Kind != TopologyNodeZone && n.Kind != TopologyNodeShardGroup {
		return fmt.Errorf("unknown aethercore next topology node kind %q", n.Kind)
	}
	if n.Kind == TopologyNodeCore {
		if n.ZoneID != "" {
			return errors.New("aethercore next topology core node must not have zone id")
		}
	} else if err := ValidateZoneID(n.ZoneID); err != nil {
		return err
	}
	if err := validateTopologyLabel("aethercore next topology node label", n.Label); err != nil {
		return err
	}
	return validateCapabilitiesForField("aethercore next topology node capability", n.Capabilities)
}

func (e AetherisNextTopologyEdge) Validate(nodes map[string]AetherisNextTopologyNode) error {
	if err := validatePolicyID("aethercore next topology edge source", e.FromNodeID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore next topology edge destination", e.ToNodeID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore next topology edge relation", e.Relation); err != nil {
		return err
	}
	if e.FromNodeID == e.ToNodeID {
		return errors.New("aethercore next topology edge cannot self-loop")
	}
	if _, found := nodes[e.FromNodeID]; !found {
		return fmt.Errorf("aethercore next topology edge source %s is missing", e.FromNodeID)
	}
	if _, found := nodes[e.ToNodeID]; !found {
		return fmt.Errorf("aethercore next topology edge destination %s is missing", e.ToNodeID)
	}
	return nil
}

func ComputeAetherisNextTopologyHash(plan AetherisNextTopologyPlan) string {
	nodes := cloneTopologyNodes(plan.Nodes)
	edges := append([]AetherisNextTopologyEdge(nil), plan.Edges...)
	sortTopologyNodes(nodes)
	sortTopologyEdges(edges)
	parts := []string{"aetheris-next-topology-v1", fmt.Sprint(plan.Version), fmt.Sprint(len(nodes))}
	for _, node := range nodes {
		node = canonicalTopologyNode(node)
		parts = append(parts, node.NodeID, string(node.Kind), string(node.ZoneID), node.Label)
		parts = appendStringSliceParts(parts, "capabilities", node.Capabilities)
	}
	parts = append(parts, fmt.Sprint(len(edges)))
	for _, edge := range edges {
		parts = append(parts, edge.FromNodeID, edge.ToNodeID, edge.Relation)
	}
	return hashParts(parts...)
}

func nextZoneDescriptor(id ZoneID, zoneType ZoneType, moduleName string, maxShards uint32, messageCapabilities []string, proofCapabilities []string) ZoneDescriptor {
	return CanonicalZoneDescriptor(ZoneDescriptor{
		ZoneID:              id,
		ZoneType:            zoneType,
		ModuleName:          moduleName,
		Enabled:             true,
		StateMachineVersion: 2,
		MempoolPolicyID:     DefaultMempoolPolicy,
		FeePolicyID:         NativeFeePolicyID,
		GasPolicyID:         DefaultGasPolicy,
		MessagePolicyID:     DefaultMessagePolicy,
		ShardLayoutEpoch:    1,
		MaxShards:           maxShards,
		MessageCapabilities: messageCapabilities,
		ProofCapabilities:   proofCapabilities,
	})
}

func canonicalTopologyNode(node AetherisNextTopologyNode) AetherisNextTopologyNode {
	node.Capabilities = append([]string(nil), node.Capabilities...)
	sort.Strings(node.Capabilities)
	return node
}

func cloneTopologyNodes(nodes []AetherisNextTopologyNode) []AetherisNextTopologyNode {
	out := make([]AetherisNextTopologyNode, len(nodes))
	for i, node := range nodes {
		out[i] = canonicalTopologyNode(node)
	}
	return out
}

func sortTopologyNodes(nodes []AetherisNextTopologyNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].NodeID < nodes[j].NodeID
	})
	for i := range nodes {
		nodes[i] = canonicalTopologyNode(nodes[i])
	}
}

func sortTopologyEdges(edges []AetherisNextTopologyEdge) {
	sort.SliceStable(edges, func(i, j int) bool {
		return compareTopologyEdges(edges[i], edges[j]) < 0
	})
}

func compareTopologyEdges(left, right AetherisNextTopologyEdge) int {
	for _, pair := range [][2]string{
		{left.FromNodeID, right.FromNodeID},
		{left.ToNodeID, right.ToNodeID},
		{left.Relation, right.Relation},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

func validateTopologyLabel(fieldName, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxScopeLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxScopeLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}
