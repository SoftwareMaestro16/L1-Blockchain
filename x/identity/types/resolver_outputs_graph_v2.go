package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityGraphNodePrefix	= IdentityZonePrefix + "/graph/node"
	IdentityGraphEdgePrefix	= IdentityZonePrefix + "/graph/edge"
	IdentityGraphRootPrefix	= IdentityZonePrefix + "/graph/root"
)

type IdentityResolverOutputType string
type IdentityGraphNodeType string
type IdentityGraphEdgeType string

const (
	IdentityResolverOutputAccountAddress		IdentityResolverOutputType	= "account_address"
	IdentityResolverOutputZoneEndpoint		IdentityResolverOutputType	= "zone_endpoint"
	IdentityResolverOutputServiceEndpoint		IdentityResolverOutputType	= "service_endpoint"
	IdentityResolverOutputContractEndpoint		IdentityResolverOutputType	= "contract_endpoint"
	IdentityResolverOutputCompositeIdentityObject	IdentityResolverOutputType	= "composite_identity_object"

	IdentityGraphNodeDomain				IdentityGraphNodeType	= "domain"
	IdentityGraphNodeAccount			IdentityGraphNodeType	= "account"
	IdentityGraphNodeService			IdentityGraphNodeType	= "service"
	IdentityGraphNodeContract			IdentityGraphNodeType	= "contract"
	IdentityGraphNodeZoneEndpoint			IdentityGraphNodeType	= "zone_endpoint"
	IdentityGraphNodeCompositeIdentityObject	IdentityGraphNodeType	= "composite_identity_object"

	IdentityGraphEdgeOwns		IdentityGraphEdgeType	= "owns"
	IdentityGraphEdgeResolvesTo	IdentityGraphEdgeType	= "resolves_to"
	IdentityGraphEdgeDelegatesTo	IdentityGraphEdgeType	= "delegates_to"
	IdentityGraphEdgeBoundTo	IdentityGraphEdgeType	= "bound_to"
	IdentityGraphEdgeReverseOf	IdentityGraphEdgeType	= "reverse_of"
	IdentityGraphEdgeServiceFor	IdentityGraphEdgeType	= "service_for"
	IdentityGraphEdgeContractFor	IdentityGraphEdgeType	= "contract_for"
)

type CompositeIdentityObjectV2 struct {
	ObjectID	string
	NameHash	string
	ComponentIDs	[]string
	MetadataHash	string
	ObjectHash	string
}

type IdentityResolverOutputV2 struct {
	OutputID	string
	NameHash	string
	OutputType	IdentityResolverOutputType
	Account		sdk.AccAddress
	ZoneEndpoint	string
	Service		*ServiceEndpointV2
	Contract	*ContractTargetV2
	Composite	*CompositeIdentityObjectV2
	Height		uint64
	OutputHash	string
}

type IdentityNode struct {
	IdentityID	string
	NodeType	IdentityGraphNodeType
	NameHash	string
	Owner		sdk.AccAddress
	Label		string
	PayloadHash	string
	NodeHash	string
}

type IdentityEdge struct {
	IdentityID	string
	TargetType	IdentityGraphNodeType
	TargetID	string
	EdgeType	IdentityGraphEdgeType
	TargetHash	string
	Height		uint64
	EdgeHash	string
}

type IdentityGraphRoot struct {
	Height			uint64
	NodeRoot		string
	EdgeRoot		string
	ResolverOutputRoot	string
	RootHash		string
}

type IdentityGraphStateV2 struct {
	Nodes	[]IdentityNode
	Edges	[]IdentityEdge
	Outputs	[]IdentityResolverOutputV2
	Height	uint64
	Root	IdentityGraphRoot
}

func IdentityGraphNodeKey(identityID string) (string, error) {
	if err := validateIdentityGraphID("identity graph node id", identityID); err != nil {
		return "", err
	}
	return IdentityGraphNodePrefix + "/" + identityID, nil
}

func IdentityGraphEdgeKey(identityID string, targetType IdentityGraphNodeType, targetID string) (string, error) {
	if err := validateIdentityGraphID("identity graph edge identity id", identityID); err != nil {
		return "", err
	}
	if !IsIdentityGraphNodeType(targetType) {
		return "", fmt.Errorf("unknown identity graph edge target type %q", targetType)
	}
	if err := validateIdentityGraphID("identity graph edge target id", targetID); err != nil {
		return "", err
	}
	return IdentityGraphEdgePrefix + "/" + identityID + "/" + string(targetType) + "/" + targetID, nil
}

func IdentityGraphRootKey(height uint64) (string, error) {
	if height == 0 {
		return "", errors.New("identity graph root height must be positive")
	}
	return fmt.Sprintf("%s/%020d", IdentityGraphRootPrefix, height), nil
}

func NewCompositeIdentityObjectV2(object CompositeIdentityObjectV2) (CompositeIdentityObjectV2, error) {
	if object.ObjectHash != "" {
		return CompositeIdentityObjectV2{}, errors.New("identity composite object hash must be empty before construction")
	}
	object.ComponentIDs = sortedUniqueStrings(object.ComponentIDs)
	if err := object.ValidateFormat(); err != nil {
		return CompositeIdentityObjectV2{}, err
	}
	object.ObjectHash = ComputeCompositeIdentityObjectV2Hash(object)
	return object, object.Validate()
}

func NewIdentityResolverOutputV2(output IdentityResolverOutputV2) (IdentityResolverOutputV2, error) {
	if output.OutputHash != "" {
		return IdentityResolverOutputV2{}, errors.New("identity resolver output hash must be empty before construction")
	}
	if err := output.ValidateFormat(); err != nil {
		return IdentityResolverOutputV2{}, err
	}
	output.OutputHash = ComputeIdentityResolverOutputV2Hash(output)
	return output, output.Validate()
}

func NewIdentityNode(node IdentityNode) (IdentityNode, error) {
	if node.NodeHash != "" {
		return IdentityNode{}, errors.New("identity graph node hash must be empty before construction")
	}
	if err := node.ValidateFormat(); err != nil {
		return IdentityNode{}, err
	}
	node.NodeHash = ComputeIdentityNodeHash(node)
	return node, node.Validate()
}

func NewIdentityEdge(edge IdentityEdge) (IdentityEdge, error) {
	if edge.EdgeHash != "" {
		return IdentityEdge{}, errors.New("identity graph edge hash must be empty before construction")
	}
	if err := edge.ValidateFormat(); err != nil {
		return IdentityEdge{}, err
	}
	edge.EdgeHash = ComputeIdentityEdgeHash(edge)
	return edge, edge.Validate()
}

func BuildIdentityGraphStateV2(nodes []IdentityNode, edges []IdentityEdge, outputs []IdentityResolverOutputV2, height uint64) (IdentityGraphStateV2, error) {
	state := IdentityGraphStateV2{
		Nodes:		normalizeIdentityNodes(nodes),
		Edges:		normalizeIdentityEdges(edges),
		Outputs:	normalizeIdentityResolverOutputsV2(outputs),
		Height:		height,
	}
	root, err := BuildIdentityGraphRoot(height, state.Nodes, state.Edges, state.Outputs)
	if err != nil {
		return IdentityGraphStateV2{}, err
	}
	state.Root = root
	return state, state.Validate()
}

func BuildIdentityGraphRoot(height uint64, nodes []IdentityNode, edges []IdentityEdge, outputs []IdentityResolverOutputV2) (IdentityGraphRoot, error) {
	if height == 0 {
		return IdentityGraphRoot{}, errors.New("identity graph root height must be positive")
	}
	root := IdentityGraphRoot{
		Height:			height,
		NodeRoot:		ComputeIdentityNodeRoot(nodes),
		EdgeRoot:		ComputeIdentityEdgeRoot(edges),
		ResolverOutputRoot:	ComputeIdentityResolverOutputRootV2(outputs),
	}
	root.RootHash = ComputeIdentityGraphRootHash(root)
	return root, root.Validate()
}

func BuildIdentityResolverOutputsFromUnifiedRecordV2(record UnifiedResolutionRecordV2, height uint64) ([]IdentityResolverOutputV2, error) {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return nil, err
	}
	if height == 0 {
		return nil, errors.New("identity resolver output height must be positive")
	}
	outputs := []IdentityResolverOutputV2{}
	if len(record.PrimaryAddress) > 0 {
		outputs = append(outputs, IdentityResolverOutputV2{
			OutputID:	"account/" + fmt.Sprintf("%x", []byte(record.PrimaryAddress)),
			NameHash:	record.NameHash,
			OutputType:	IdentityResolverOutputAccountAddress,
			Account:	cloneSpecAddress(record.PrimaryAddress),
			Height:		height,
		})
	}
	if record.RoutingMetadata.ZoneID != "" || record.RoutingMetadata.RouteID != "" {
		endpoint := strings.Join([]string{record.RoutingMetadata.ZoneID, record.RoutingMetadata.ShardID, record.RoutingMetadata.RouteID}, "/")
		outputs = append(outputs, IdentityResolverOutputV2{
			OutputID:	"zone/" + identityHash(endpoint),
			NameHash:	record.NameHash,
			OutputType:	IdentityResolverOutputZoneEndpoint,
			ZoneEndpoint:	endpoint,
			Height:		height,
		})
	}
	for _, service := range record.ServiceEndpoints {
		outputs = append(outputs, IdentityResolverOutputV2{
			OutputID:	"service/" + serviceEndpointIDV2(service),
			NameHash:	record.NameHash,
			OutputType:	IdentityResolverOutputServiceEndpoint,
			Service:	cloneServiceEndpointPtrV2(service),
			Height:		height,
		})
	}
	for _, contract := range record.ContractTargets {
		outputs = append(outputs, IdentityResolverOutputV2{
			OutputID:	"contract/" + contractTargetIDV2(contract),
			NameHash:	record.NameHash,
			OutputType:	IdentityResolverOutputContractEndpoint,
			Contract:	cloneContractTargetPtrV2(contract),
			Height:		height,
		})
	}
	constructed := make([]IdentityResolverOutputV2, 0, len(outputs))
	for _, output := range outputs {
		next, err := NewIdentityResolverOutputV2(output)
		if err != nil {
			return nil, err
		}
		constructed = append(constructed, next)
	}
	return normalizeIdentityResolverOutputsV2(constructed), nil
}

func BuildIdentityGraphFromResolverOutputsV2(nameHash string, owner sdk.AccAddress, outputs []IdentityResolverOutputV2, height uint64) (IdentityGraphStateV2, error) {
	domainNode, err := NewIdentityNode(IdentityNode{
		IdentityID:	"domain/" + nameHash,
		NodeType:	IdentityGraphNodeDomain,
		NameHash:	nameHash,
		Owner:		cloneSpecAddress(owner),
		Label:		"domain",
		PayloadHash:	nameHash,
	})
	if err != nil {
		return IdentityGraphStateV2{}, err
	}
	nodes := []IdentityNode{domainNode}
	edges := []IdentityEdge{}
	for _, output := range normalizeIdentityResolverOutputsV2(outputs) {
		node, edge, err := identityGraphNodeAndEdgeFromOutputV2(domainNode.IdentityID, output)
		if err != nil {
			return IdentityGraphStateV2{}, err
		}
		nodes = append(nodes, node)
		edges = append(edges, edge)
	}
	return BuildIdentityGraphStateV2(nodes, edges, outputs, height)
}

func (object CompositeIdentityObjectV2) ValidateFormat() error {
	if err := validateIdentityGraphID("identity composite object id", object.ObjectID); err != nil {
		return err
	}
	if err := validateHexHash("identity composite object name hash", object.NameHash); err != nil {
		return err
	}
	if len(object.ComponentIDs) == 0 {
		return errors.New("identity composite object requires components")
	}
	for i, componentID := range object.ComponentIDs {
		if err := validateIdentityGraphID("identity composite object component id", componentID); err != nil {
			return err
		}
		if i > 0 && object.ComponentIDs[i-1] >= componentID {
			return errors.New("identity composite object components must be sorted canonically")
		}
	}
	if err := validateHexHash("identity composite object metadata hash", object.MetadataHash); err != nil {
		return err
	}
	if object.ObjectHash != "" {
		return validateHexHash("identity composite object hash", object.ObjectHash)
	}
	return nil
}

func (object CompositeIdentityObjectV2) Validate() error {
	if err := object.ValidateFormat(); err != nil {
		return err
	}
	if object.ObjectHash == "" {
		return errors.New("identity composite object hash is required")
	}
	if object.ObjectHash != ComputeCompositeIdentityObjectV2Hash(object) {
		return errors.New("identity composite object hash mismatch")
	}
	return nil
}

func (output IdentityResolverOutputV2) ValidateFormat() error {
	if err := validateIdentityGraphID("identity resolver output id", output.OutputID); err != nil {
		return err
	}
	if err := validateHexHash("identity resolver output name hash", output.NameHash); err != nil {
		return err
	}
	if !IsIdentityResolverOutputType(output.OutputType) {
		return fmt.Errorf("unknown identity resolver output type %q", output.OutputType)
	}
	if output.Height == 0 {
		return errors.New("identity resolver output height must be positive")
	}
	switch output.OutputType {
	case IdentityResolverOutputAccountAddress:
		return validateSpecAddress("identity resolver output account", output.Account)
	case IdentityResolverOutputZoneEndpoint:
		return validateUnifiedRecordValue("identity resolver output zone endpoint", output.ZoneEndpoint, MaxUnifiedEndpointBytes)
	case IdentityResolverOutputServiceEndpoint:
		if output.Service == nil {
			return errors.New("identity resolver output service endpoint is required")
		}
		return validateServiceEndpointsV2([]ServiceEndpointV2{*output.Service}, output.Service.TTL)
	case IdentityResolverOutputContractEndpoint:
		if output.Contract == nil {
			return errors.New("identity resolver output contract endpoint is required")
		}
		return validateContractTargetsV2([]ContractTargetV2{*output.Contract}, nil)
	case IdentityResolverOutputCompositeIdentityObject:
		if output.Composite == nil {
			return errors.New("identity resolver output composite object is required")
		}
		return output.Composite.Validate()
	default:
		return fmt.Errorf("unknown identity resolver output type %q", output.OutputType)
	}
}

func (output IdentityResolverOutputV2) Validate() error {
	if err := output.ValidateFormat(); err != nil {
		return err
	}
	if output.OutputHash == "" {
		return errors.New("identity resolver output hash is required")
	}
	if output.OutputHash != ComputeIdentityResolverOutputV2Hash(output) {
		return errors.New("identity resolver output hash mismatch")
	}
	return nil
}

func (node IdentityNode) ValidateFormat() error {
	if err := validateIdentityGraphID("identity graph node id", node.IdentityID); err != nil {
		return err
	}
	if !IsIdentityGraphNodeType(node.NodeType) {
		return fmt.Errorf("unknown identity graph node type %q", node.NodeType)
	}
	if node.NameHash != "" {
		if err := validateHexHash("identity graph node name hash", node.NameHash); err != nil {
			return err
		}
	}
	if len(node.Owner) > 0 {
		if err := validateSpecAddress("identity graph node owner", node.Owner); err != nil {
			return err
		}
	}
	if node.Label != "" {
		if err := validateUnifiedRecordKey("identity graph node label", node.Label); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity graph node payload hash", node.PayloadHash); err != nil {
		return err
	}
	if node.NodeHash != "" {
		return validateHexHash("identity graph node hash", node.NodeHash)
	}
	return nil
}

func (node IdentityNode) Validate() error {
	if err := node.ValidateFormat(); err != nil {
		return err
	}
	if node.NodeHash == "" {
		return errors.New("identity graph node hash is required")
	}
	if node.NodeHash != ComputeIdentityNodeHash(node) {
		return errors.New("identity graph node hash mismatch")
	}
	return nil
}

func (edge IdentityEdge) ValidateFormat() error {
	if _, err := IdentityGraphEdgeKey(edge.IdentityID, edge.TargetType, edge.TargetID); err != nil {
		return err
	}
	if !IsIdentityGraphEdgeType(edge.EdgeType) {
		return fmt.Errorf("unknown identity graph edge type %q", edge.EdgeType)
	}
	if err := validateHexHash("identity graph edge target hash", edge.TargetHash); err != nil {
		return err
	}
	if edge.Height == 0 {
		return errors.New("identity graph edge height must be positive")
	}
	if edge.EdgeHash != "" {
		return validateHexHash("identity graph edge hash", edge.EdgeHash)
	}
	return nil
}

func (edge IdentityEdge) Validate() error {
	if err := edge.ValidateFormat(); err != nil {
		return err
	}
	if edge.EdgeHash == "" {
		return errors.New("identity graph edge hash is required")
	}
	if edge.EdgeHash != ComputeIdentityEdgeHash(edge) {
		return errors.New("identity graph edge hash mismatch")
	}
	return nil
}

func (root IdentityGraphRoot) Validate() error {
	if root.Height == 0 {
		return errors.New("identity graph root height must be positive")
	}
	if _, err := IdentityGraphRootKey(root.Height); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "identity graph node root", value: root.NodeRoot},
		{name: "identity graph edge root", value: root.EdgeRoot},
		{name: "identity graph resolver output root", value: root.ResolverOutputRoot},
		{name: "identity graph root hash", value: root.RootHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if root.RootHash != ComputeIdentityGraphRootHash(root) {
		return errors.New("identity graph root hash mismatch")
	}
	return nil
}

func (state IdentityGraphStateV2) Validate() error {
	if state.Height == 0 {
		return errors.New("identity graph state height must be positive")
	}
	if err := validateIdentityNodes(state.Nodes); err != nil {
		return err
	}
	if err := validateIdentityEdges(state.Edges, state.Nodes); err != nil {
		return err
	}
	if err := validateIdentityResolverOutputsV2(state.Outputs); err != nil {
		return err
	}
	if err := state.Root.Validate(); err != nil {
		return err
	}
	expected, err := BuildIdentityGraphRoot(state.Height, state.Nodes, state.Edges, state.Outputs)
	if err != nil {
		return err
	}
	if state.Root != expected {
		return errors.New("identity graph state root mismatch")
	}
	return nil
}

func ComputeCompositeIdentityObjectV2Hash(object CompositeIdentityObjectV2) string {
	components := sortedUniqueStrings(object.ComponentIDs)
	parts := []string{"identity-composite-object-v1", object.ObjectID, object.NameHash, object.MetadataHash, fmt.Sprintf("%020d", len(components))}
	parts = append(parts, components...)
	return identityHash(parts...)
}

func ComputeIdentityResolverOutputV2Hash(output IdentityResolverOutputV2) string {
	parts := []string{"identity-resolver-output-v1", output.OutputID, output.NameHash, string(output.OutputType), fmt.Sprintf("%020d", output.Height)}
	switch output.OutputType {
	case IdentityResolverOutputAccountAddress:
		parts = append(parts, fmt.Sprintf("%x", []byte(output.Account)))
	case IdentityResolverOutputZoneEndpoint:
		parts = append(parts, output.ZoneEndpoint)
	case IdentityResolverOutputServiceEndpoint:
		if output.Service != nil {
			parts = append(parts, serviceEndpointIDV2(*output.Service), output.Service.Endpoint, output.Service.ServiceType, output.Service.Transport, fmt.Sprintf("%020d", output.Service.TTL))
		}
	case IdentityResolverOutputContractEndpoint:
		if output.Contract != nil {
			parts = append(parts, contractTargetIDV2(*output.Contract), fmt.Sprintf("%x", []byte(contractTargetAddressV2(*output.Contract))), output.Contract.CodeID, output.Contract.Entrypoint)
		}
	case IdentityResolverOutputCompositeIdentityObject:
		if output.Composite != nil {
			parts = append(parts, output.Composite.ObjectHash)
		}
	}
	return identityHash(parts...)
}

func ComputeIdentityNodeHash(node IdentityNode) string {
	return identityHash("identity-graph-node-v1", node.IdentityID, string(node.NodeType), node.NameHash, fmt.Sprintf("%x", []byte(node.Owner)), node.Label, node.PayloadHash)
}

func ComputeIdentityEdgeHash(edge IdentityEdge) string {
	return identityHash("identity-graph-edge-v1", edge.IdentityID, string(edge.TargetType), edge.TargetID, string(edge.EdgeType), edge.TargetHash, fmt.Sprintf("%020d", edge.Height))
}

func ComputeIdentityNodeRoot(nodes []IdentityNode) string {
	ordered := normalizeIdentityNodes(nodes)
	parts := []string{"identity-graph-node-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, node := range ordered {
		parts = append(parts, node.NodeHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityEdgeRoot(edges []IdentityEdge) string {
	ordered := normalizeIdentityEdges(edges)
	parts := []string{"identity-graph-edge-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, edge := range ordered {
		parts = append(parts, edge.EdgeHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityResolverOutputRootV2(outputs []IdentityResolverOutputV2) string {
	ordered := normalizeIdentityResolverOutputsV2(outputs)
	parts := []string{"identity-resolver-output-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, output := range ordered {
		parts = append(parts, output.OutputHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityGraphRootHash(root IdentityGraphRoot) string {
	return identityHash("identity-graph-root-v1", fmt.Sprintf("%020d", root.Height), root.NodeRoot, root.EdgeRoot, root.ResolverOutputRoot)
}

func IsIdentityResolverOutputType(outputType IdentityResolverOutputType) bool {
	switch outputType {
	case IdentityResolverOutputAccountAddress,
		IdentityResolverOutputZoneEndpoint,
		IdentityResolverOutputServiceEndpoint,
		IdentityResolverOutputContractEndpoint,
		IdentityResolverOutputCompositeIdentityObject:
		return true
	default:
		return false
	}
}

func IsIdentityGraphNodeType(nodeType IdentityGraphNodeType) bool {
	switch nodeType {
	case IdentityGraphNodeDomain,
		IdentityGraphNodeAccount,
		IdentityGraphNodeService,
		IdentityGraphNodeContract,
		IdentityGraphNodeZoneEndpoint,
		IdentityGraphNodeCompositeIdentityObject:
		return true
	default:
		return false
	}
}

func IsIdentityGraphEdgeType(edgeType IdentityGraphEdgeType) bool {
	switch edgeType {
	case IdentityGraphEdgeOwns,
		IdentityGraphEdgeResolvesTo,
		IdentityGraphEdgeDelegatesTo,
		IdentityGraphEdgeBoundTo,
		IdentityGraphEdgeReverseOf,
		IdentityGraphEdgeServiceFor,
		IdentityGraphEdgeContractFor:
		return true
	default:
		return false
	}
}

func identityGraphNodeAndEdgeFromOutputV2(domainID string, output IdentityResolverOutputV2) (IdentityNode, IdentityEdge, error) {
	var nodeType IdentityGraphNodeType
	var edgeType IdentityGraphEdgeType
	switch output.OutputType {
	case IdentityResolverOutputAccountAddress:
		nodeType, edgeType = IdentityGraphNodeAccount, IdentityGraphEdgeResolvesTo
	case IdentityResolverOutputZoneEndpoint:
		nodeType, edgeType = IdentityGraphNodeZoneEndpoint, IdentityGraphEdgeResolvesTo
	case IdentityResolverOutputServiceEndpoint:
		nodeType, edgeType = IdentityGraphNodeService, IdentityGraphEdgeServiceFor
	case IdentityResolverOutputContractEndpoint:
		nodeType, edgeType = IdentityGraphNodeContract, IdentityGraphEdgeContractFor
	case IdentityResolverOutputCompositeIdentityObject:
		nodeType, edgeType = IdentityGraphNodeCompositeIdentityObject, IdentityGraphEdgeBoundTo
	default:
		return IdentityNode{}, IdentityEdge{}, fmt.Errorf("unknown identity resolver output type %q", output.OutputType)
	}
	node, err := NewIdentityNode(IdentityNode{
		IdentityID:	string(nodeType) + "/" + output.OutputID,
		NodeType:	nodeType,
		NameHash:	output.NameHash,
		Label:		string(output.OutputType),
		PayloadHash:	output.OutputHash,
	})
	if err != nil {
		return IdentityNode{}, IdentityEdge{}, err
	}
	edge, err := NewIdentityEdge(IdentityEdge{
		IdentityID:	domainID,
		TargetType:	nodeType,
		TargetID:	output.OutputID,
		EdgeType:	edgeType,
		TargetHash:	node.NodeHash,
		Height:		output.Height,
	})
	if err != nil {
		return IdentityNode{}, IdentityEdge{}, err
	}
	return node, edge, nil
}

func validateIdentityGraphID(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > 160 {
		return fmt.Errorf("%s must not exceed 160 bytes", fieldName)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func normalizeIdentityResolverOutputsV2(outputs []IdentityResolverOutputV2) []IdentityResolverOutputV2 {
	out := append([]IdentityResolverOutputV2(nil), outputs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].OutputID < out[j].OutputID })
	return out
}

func normalizeIdentityNodes(nodes []IdentityNode) []IdentityNode {
	out := append([]IdentityNode(nil), nodes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].IdentityID < out[j].IdentityID })
	return out
}

func normalizeIdentityEdges(edges []IdentityEdge) []IdentityEdge {
	out := append([]IdentityEdge(nil), edges...)
	sort.SliceStable(out, func(i, j int) bool {
		left, _ := IdentityGraphEdgeKey(out[i].IdentityID, out[i].TargetType, out[i].TargetID)
		right, _ := IdentityGraphEdgeKey(out[j].IdentityID, out[j].TargetType, out[j].TargetID)
		return left < right
	})
	return out
}

func validateIdentityResolverOutputsV2(outputs []IdentityResolverOutputV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, output := range outputs {
		if err := output.Validate(); err != nil {
			return err
		}
		if _, found := seen[output.OutputID]; found {
			return fmt.Errorf("duplicate identity resolver output %s", output.OutputID)
		}
		seen[output.OutputID] = struct{}{}
		if previous != "" && previous >= output.OutputID {
			return errors.New("identity resolver outputs must be sorted canonically")
		}
		previous = output.OutputID
	}
	return nil
}

func validateIdentityNodes(nodes []IdentityNode) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, node := range nodes {
		if err := node.Validate(); err != nil {
			return err
		}
		if _, found := seen[node.IdentityID]; found {
			return fmt.Errorf("duplicate identity graph node %s", node.IdentityID)
		}
		seen[node.IdentityID] = struct{}{}
		if previous != "" && previous >= node.IdentityID {
			return errors.New("identity graph nodes must be sorted canonically")
		}
		previous = node.IdentityID
	}
	return nil
}

func validateIdentityEdges(edges []IdentityEdge, nodes []IdentityNode) error {
	nodeHashes := map[string]string{}
	for _, node := range nodes {
		nodeHashes[node.IdentityID] = node.NodeHash
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			return err
		}
		key, _ := IdentityGraphEdgeKey(edge.IdentityID, edge.TargetType, edge.TargetID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate identity graph edge %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("identity graph edges must be sorted canonically")
		}
		targetNodeID := string(edge.TargetType) + "/" + edge.TargetID
		if targetHash, found := nodeHashes[targetNodeID]; found && targetHash != edge.TargetHash {
			return fmt.Errorf("identity graph edge target hash mismatch for %s", targetNodeID)
		}
		previous = key
	}
	return nil
}

func cloneServiceEndpointPtrV2(endpoint ServiceEndpointV2) *ServiceEndpointV2 {
	copied := endpoint
	return &copied
}

func cloneContractTargetPtrV2(target ContractTargetV2) *ContractTargetV2 {
	copied := target
	copied.Address = cloneSpecAddress(target.Address)
	copied.ContractAddress = cloneSpecAddress(target.ContractAddress)
	return &copied
}
