package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServiceInterfaceStateObject string
type XServiceInterfaceMessageName string
type XServiceInterfaceQueryName string
type XServiceInterfaceFailureMode string
type XServiceInterfaceIntegrationPoint string

const (
	XServiceInterfaceStateInterface		XServiceInterfaceStateObject	= "ServiceInterface"
	XServiceInterfaceStateMethod		XServiceInterfaceStateObject	= "ServiceMethod"
	XServiceInterfaceStateEvent		XServiceInterfaceStateObject	= "ServiceEvent"
	XServiceInterfaceStateError		XServiceInterfaceStateObject	= "ServiceError"
	XServiceInterfaceStateInterfaceVersion	XServiceInterfaceStateObject	= "InterfaceVersion"

	XServiceInterfaceMsgRegisterInterface	XServiceInterfaceMessageName	= "MsgRegisterInterface"
	XServiceInterfaceMsgUpdateInterface	XServiceInterfaceMessageName	= "MsgUpdateInterface"
	XServiceInterfaceMsgDeprecateInterface	XServiceInterfaceMessageName	= "MsgDeprecateInterface"

	XServiceInterfaceQueryInterface		XServiceInterfaceQueryName	= "QueryInterface"
	XServiceInterfaceQueryMethod		XServiceInterfaceQueryName	= "QueryMethod"
	XServiceInterfaceQueryInterfaceProof	XServiceInterfaceQueryName	= "QueryInterfaceProof"
	XServiceInterfaceQueryInterfacesByOwner	XServiceInterfaceQueryName	= "QueryInterfacesByOwner"

	XServiceInterfaceFailureSchemaHashMismatch		XServiceInterfaceFailureMode	= "schema_hash_mismatch"
	XServiceInterfaceFailureMethodIDCollision		XServiceInterfaceFailureMode	= "method_id_collision"
	XServiceInterfaceFailureUnsupportedSchemaEncoding	XServiceInterfaceFailureMode	= "unsupported_schema_encoding"
	XServiceInterfaceFailureBreakingUpdateWithoutNewHash	XServiceInterfaceFailureMode	= "breaking_update_without_new_interface_hash"

	XServiceInterfaceIntegrationServices		XServiceInterfaceIntegrationPoint	= "x/services"
	XServiceInterfaceIntegrationWalletSDK		XServiceInterfaceIntegrationPoint	= "wallet_sdk"
	XServiceInterfaceIntegrationCLI			XServiceInterfaceIntegrationPoint	= "cli"
	XServiceInterfaceIntegrationContractAdapter	XServiceInterfaceIntegrationPoint	= "contract_adapter"
)

type XServiceInterfaceFailureCoverage struct {
	Mode	XServiceInterfaceFailureMode
	Guard	string
	Scope	string
}

type XServiceInterfaceModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServiceInterfaceStateObject
	Messages		[]XServiceInterfaceMessageName
	Queries			[]XServiceInterfaceQueryName
	FailureModes		[]XServiceInterfaceFailureCoverage
	IntegrationPoints	[]XServiceInterfaceIntegrationPoint
	BreakdownHash		string
}

type MsgDeprecateInterface struct {
	Authority	string
	Marker		ServiceInterfaceDeprecationMarker
	MsgHash		string
}

type QueryMethod struct {
	InterfaceHash	string
	MethodID	string
}

type QueryMethodResponse struct {
	Method	ServiceInterfaceMethodSchema
	Found	bool
}

type QueryInterfacesByOwner struct {
	Owner string
}

type QueryInterfacesByOwnerResponse struct {
	Interfaces	[]FormalServiceInterface
	Total		uint64
	ResponseHash	string
}

type ServiceInterfaceContractAdapterSchema struct {
	InterfaceHash	string
	MethodID	string
	ABIHash		string
	Encoding	string
	AdapterHash	string
}

func DefaultXServiceInterfaceModuleBreakdown() (XServiceInterfaceModuleBreakdown, error) {
	breakdown := XServiceInterfaceModuleBreakdown{
		ModulePath:	ServiceModuleInterface,
		Purpose: []string{
			"formal_service_interface_schemas",
			"interface_proofs",
			"schema_verification",
			"versioned_interface_lifecycle",
		},
		StateObjects: []XServiceInterfaceStateObject{
			XServiceInterfaceStateInterface,
			XServiceInterfaceStateMethod,
			XServiceInterfaceStateEvent,
			XServiceInterfaceStateError,
			XServiceInterfaceStateInterfaceVersion,
		},
		Messages: []XServiceInterfaceMessageName{
			XServiceInterfaceMsgRegisterInterface,
			XServiceInterfaceMsgUpdateInterface,
			XServiceInterfaceMsgDeprecateInterface,
		},
		Queries: []XServiceInterfaceQueryName{
			XServiceInterfaceQueryInterface,
			XServiceInterfaceQueryMethod,
			XServiceInterfaceQueryInterfaceProof,
			XServiceInterfaceQueryInterfacesByOwner,
		},
		FailureModes: []XServiceInterfaceFailureCoverage{
			newXServiceInterfaceFailureCoverage(XServiceInterfaceFailureSchemaHashMismatch, "MsgRegisterInterface.ValidateBasic", ServiceStoreV2InterfacePrefix),
			newXServiceInterfaceFailureCoverage(XServiceInterfaceFailureMethodIDCollision, "FormalServiceInterface.Validate", ServiceStoreV2MethodIndexPrefix),
			newXServiceInterfaceFailureCoverage(XServiceInterfaceFailureUnsupportedSchemaEncoding, "InterfaceSchemaFormat.Validate", ServiceStoreV2InterfacePrefix),
			newXServiceInterfaceFailureCoverage(XServiceInterfaceFailureBreakingUpdateWithoutNewHash, "ValidateServiceInterfaceVersionChange", ServiceStoreV2InterfacePrefix),
		},
		IntegrationPoints: []XServiceInterfaceIntegrationPoint{
			XServiceInterfaceIntegrationServices,
			XServiceInterfaceIntegrationWalletSDK,
			XServiceInterfaceIntegrationCLI,
			XServiceInterfaceIntegrationContractAdapter,
		},
	}
	return NewXServiceInterfaceModuleBreakdown(breakdown)
}

func NewXServiceInterfaceModuleBreakdown(breakdown XServiceInterfaceModuleBreakdown) (XServiceInterfaceModuleBreakdown, error) {
	breakdown = canonicalXServiceInterfaceModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServiceInterfaceModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServiceInterfaceModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func NewMsgDeprecateInterface(authority string, marker ServiceInterfaceDeprecationMarker) (MsgDeprecateInterface, error) {
	msg := MsgDeprecateInterface{Authority: strings.TrimSpace(authority), Marker: canonicalServiceInterfaceDeprecationMarker(marker)}
	msg.MsgHash = ComputeMsgDeprecateInterfaceHash(msg)
	return msg, msg.ValidateBasic()
}

func DeprecateInterfaceInState(state ServiceRegistryState, msg MsgDeprecateInterface, height uint64) (ServiceInterfaceDeprecationMarker, error) {
	if err := msg.ValidateBasic(); err != nil {
		return ServiceInterfaceDeprecationMarker{}, err
	}
	if height == 0 {
		return ServiceInterfaceDeprecationMarker{}, errors.New("x/serviceinterface deprecation height must be positive")
	}
	if _, found := state.ServiceInterfaceByHash(msg.Marker.InterfaceHash); !found {
		return ServiceInterfaceDeprecationMarker{}, fmt.Errorf("x/serviceinterface interface %s not found", msg.Marker.InterfaceHash)
	}
	if height > msg.Marker.DeprecatedHeight {
		return ServiceInterfaceDeprecationMarker{}, errors.New("x/serviceinterface deprecation cannot be applied after deprecation height")
	}
	return msg.Marker, nil
}

func QueryMethodFromInterface(definition FormalServiceInterface, query QueryMethod) (QueryMethodResponse, error) {
	if err := definition.Validate(); err != nil {
		return QueryMethodResponse{}, err
	}
	if err := query.Validate(); err != nil {
		return QueryMethodResponse{}, err
	}
	if definition.InterfaceHash != query.InterfaceHash {
		return QueryMethodResponse{Found: false}, nil
	}
	method, found := definition.MethodByName(query.MethodID)
	if !found {
		return QueryMethodResponse{Found: false}, nil
	}
	return QueryMethodResponse{Method: method, Found: true}, method.Validate()
}

func QueryInterfacesByOwnerFromServiceRegistry(state ServiceRegistryState, owner string) (QueryInterfacesByOwnerResponse, error) {
	if err := state.Validate(); err != nil {
		return QueryInterfacesByOwnerResponse{}, err
	}
	query := QueryInterfacesByOwner{Owner: strings.TrimSpace(owner)}
	if err := query.Validate(); err != nil {
		return QueryInterfacesByOwnerResponse{}, err
	}
	seen := map[string]struct{}{}
	definitions := []FormalServiceInterface{}
	for _, descriptor := range state.Descriptors {
		if descriptor.Owner != query.Owner {
			continue
		}
		if _, found := seen[descriptor.Interface.InterfaceHash]; found {
			continue
		}
		definition, err := NewFormalServiceInterface(descriptor.Interface)
		if err != nil {
			return QueryInterfacesByOwnerResponse{}, err
		}
		seen[definition.InterfaceHash] = struct{}{}
		definitions = append(definitions, definition)
	}
	sort.SliceStable(definitions, func(i, j int) bool { return definitions[i].InterfaceHash < definitions[j].InterfaceHash })
	response := QueryInterfacesByOwnerResponse{Interfaces: definitions, Total: uint64(len(definitions))}
	response.ResponseHash = ComputeQueryInterfacesByOwnerResponseHash(response)
	return response, response.Validate()
}

func NewServiceInterfaceContractAdapterSchema(interfaceHash, methodID, abiHash, encoding string) (ServiceInterfaceContractAdapterSchema, error) {
	schema := ServiceInterfaceContractAdapterSchema{
		InterfaceHash:	strings.ToLower(strings.TrimSpace(interfaceHash)),
		MethodID:	strings.TrimSpace(methodID),
		ABIHash:	strings.ToLower(strings.TrimSpace(abiHash)),
		Encoding:	strings.TrimSpace(encoding),
	}
	schema.AdapterHash = ComputeServiceInterfaceContractAdapterSchemaHash(schema)
	return schema, schema.Validate()
}

func (breakdown XServiceInterfaceModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModuleInterface {
		return errors.New("x/serviceinterface breakdown must describe x/serviceinterface")
	}
	if err := validateSortedTokens("x/serviceinterface purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServiceInterfaceStateObjects(breakdown.StateObjects); err != nil {
		return err
	}
	if err := validateXServiceInterfaceMessages(breakdown.Messages); err != nil {
		return err
	}
	if err := validateXServiceInterfaceQueries(breakdown.Queries); err != nil {
		return err
	}
	if err := validateXServiceInterfaceFailureCoverage(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServiceInterfaceIntegrationPoints(breakdown.IntegrationPoints); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/serviceinterface breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServiceInterfaceModuleBreakdown) Validate() error {
	breakdown = canonicalXServiceInterfaceModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/serviceinterface breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeXServiceInterfaceModuleBreakdownHash(breakdown) {
		return errors.New("x/serviceinterface breakdown hash mismatch")
	}
	return nil
}

func (coverage XServiceInterfaceFailureCoverage) Validate() error {
	if !IsXServiceInterfaceFailureMode(coverage.Mode) {
		return fmt.Errorf("x/serviceinterface unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/serviceinterface failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.Scope + "/_") {
		return fmt.Errorf("x/serviceinterface failure scope %s must be services store key", coverage.Scope)
	}
	return nil
}

func (msg MsgDeprecateInterface) ValidateBasic() error {
	if err := validateInterfaceToken("x/serviceinterface deprecation authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Marker.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceinterface deprecation message hash", msg.MsgHash); err != nil {
		return err
	}
	if msg.MsgHash != ComputeMsgDeprecateInterfaceHash(msg) {
		return errors.New("x/serviceinterface deprecation message hash mismatch")
	}
	return nil
}

func (query QueryMethod) Validate() error {
	if err := coretypes.ValidateHash("x/serviceinterface query method interface hash", query.InterfaceHash); err != nil {
		return err
	}
	return validateInterfaceToken("x/serviceinterface query method id", query.MethodID)
}

func (query QueryInterfacesByOwner) Validate() error {
	return validateInterfaceToken("x/serviceinterface query owner", query.Owner)
}

func (response QueryInterfacesByOwnerResponse) Validate() error {
	if response.Total != uint64(len(response.Interfaces)) {
		return errors.New("x/serviceinterface owner query total mismatch")
	}
	previous := ""
	for _, definition := range response.Interfaces {
		if err := definition.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= definition.InterfaceHash {
			return errors.New("x/serviceinterface owner query interfaces must be sorted")
		}
		previous = definition.InterfaceHash
	}
	if err := coretypes.ValidateHash("x/serviceinterface owner query response hash", response.ResponseHash); err != nil {
		return err
	}
	if response.ResponseHash != ComputeQueryInterfacesByOwnerResponseHash(response) {
		return errors.New("x/serviceinterface owner query response hash mismatch")
	}
	return nil
}

func (schema ServiceInterfaceContractAdapterSchema) Validate() error {
	if err := coretypes.ValidateHash("x/serviceinterface contract adapter interface hash", schema.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/serviceinterface contract adapter method id", schema.MethodID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceinterface contract adapter abi hash", schema.ABIHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/serviceinterface contract adapter encoding", schema.Encoding); err != nil {
		return err
	}
	if schema.Encoding != "cosmwasm-json" && schema.Encoding != "avm-binary" {
		return fmt.Errorf("x/serviceinterface contract adapter encoding %q is not supported", schema.Encoding)
	}
	if err := coretypes.ValidateHash("x/serviceinterface contract adapter hash", schema.AdapterHash); err != nil {
		return err
	}
	if schema.AdapterHash != ComputeServiceInterfaceContractAdapterSchemaHash(schema) {
		return errors.New("x/serviceinterface contract adapter hash mismatch")
	}
	return nil
}

func IsXServiceInterfaceStateObject(object XServiceInterfaceStateObject) bool {
	switch object {
	case XServiceInterfaceStateInterface, XServiceInterfaceStateMethod, XServiceInterfaceStateEvent, XServiceInterfaceStateError, XServiceInterfaceStateInterfaceVersion:
		return true
	default:
		return false
	}
}

func IsXServiceInterfaceMessageName(message XServiceInterfaceMessageName) bool {
	switch message {
	case XServiceInterfaceMsgRegisterInterface, XServiceInterfaceMsgUpdateInterface, XServiceInterfaceMsgDeprecateInterface:
		return true
	default:
		return false
	}
}

func IsXServiceInterfaceQueryName(query XServiceInterfaceQueryName) bool {
	switch query {
	case XServiceInterfaceQueryInterface, XServiceInterfaceQueryMethod, XServiceInterfaceQueryInterfaceProof, XServiceInterfaceQueryInterfacesByOwner:
		return true
	default:
		return false
	}
}

func IsXServiceInterfaceFailureMode(mode XServiceInterfaceFailureMode) bool {
	switch mode {
	case XServiceInterfaceFailureSchemaHashMismatch, XServiceInterfaceFailureMethodIDCollision, XServiceInterfaceFailureUnsupportedSchemaEncoding, XServiceInterfaceFailureBreakingUpdateWithoutNewHash:
		return true
	default:
		return false
	}
}

func IsXServiceInterfaceIntegrationPoint(point XServiceInterfaceIntegrationPoint) bool {
	switch point {
	case XServiceInterfaceIntegrationServices, XServiceInterfaceIntegrationWalletSDK, XServiceInterfaceIntegrationCLI, XServiceInterfaceIntegrationContractAdapter:
		return true
	default:
		return false
	}
}

func ComputeXServiceInterfaceModuleBreakdownHash(breakdown XServiceInterfaceModuleBreakdown) string {
	breakdown = canonicalXServiceInterfaceModuleBreakdown(breakdown)
	parts := []string{
		"aetra-x-serviceinterface-module-breakdown-v1",
		breakdown.ModulePath,
		"purpose",
		fmt.Sprint(len(breakdown.Purpose)),
	}
	parts = append(parts, breakdown.Purpose...)
	parts = append(parts, "state", fmt.Sprint(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		parts = append(parts, string(object))
	}
	parts = append(parts, "messages", fmt.Sprint(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		parts = append(parts, string(message))
	}
	parts = append(parts, "queries", fmt.Sprint(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		parts = append(parts, string(query))
	}
	parts = append(parts, "failures", fmt.Sprint(len(breakdown.FailureModes)))
	for _, coverage := range breakdown.FailureModes {
		parts = append(parts, string(coverage.Mode), coverage.Guard, coverage.Scope)
	}
	parts = append(parts, "integrations", fmt.Sprint(len(breakdown.IntegrationPoints)))
	for _, point := range breakdown.IntegrationPoints {
		parts = append(parts, string(point))
	}
	return servicesHashParts(parts...)
}

func ComputeMsgDeprecateInterfaceHash(msg MsgDeprecateInterface) string {
	return servicesHashParts(
		"aetra-x-serviceinterface-msg-deprecate-v1",
		msg.Authority,
		msg.Marker.MarkerHash,
	)
}

func ComputeQueryInterfacesByOwnerResponseHash(response QueryInterfacesByOwnerResponse) string {
	definitions := append([]FormalServiceInterface(nil), response.Interfaces...)
	sort.SliceStable(definitions, func(i, j int) bool { return definitions[i].InterfaceHash < definitions[j].InterfaceHash })
	parts := []string{"aetra-x-serviceinterface-owner-query-v1", fmt.Sprint(response.Total)}
	for _, definition := range definitions {
		parts = append(parts, definition.DefinitionHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceInterfaceContractAdapterSchemaHash(schema ServiceInterfaceContractAdapterSchema) string {
	return servicesHashParts(
		"aetra-x-serviceinterface-contract-adapter-v1",
		schema.InterfaceHash,
		schema.MethodID,
		schema.ABIHash,
		schema.Encoding,
	)
}

func newXServiceInterfaceFailureCoverage(mode XServiceInterfaceFailureMode, guard, scope string) XServiceInterfaceFailureCoverage {
	return XServiceInterfaceFailureCoverage{Mode: mode, Guard: guard, Scope: scope}
}

func canonicalXServiceInterfaceModuleBreakdown(breakdown XServiceInterfaceModuleBreakdown) XServiceInterfaceModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedStrings(breakdown.Purpose)
	breakdown.StateObjects = sortedXServiceInterfaceStateObjects(breakdown.StateObjects)
	breakdown.Messages = sortedXServiceInterfaceMessages(breakdown.Messages)
	breakdown.Queries = sortedXServiceInterfaceQueries(breakdown.Queries)
	breakdown.FailureModes = sortedXServiceInterfaceFailureCoverage(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedXServiceInterfaceIntegrationPoints(breakdown.IntegrationPoints)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateXServiceInterfaceStateObjects(objects []XServiceInterfaceStateObject) error {
	required := []XServiceInterfaceStateObject{XServiceInterfaceStateInterface, XServiceInterfaceStateMethod, XServiceInterfaceStateEvent, XServiceInterfaceStateError, XServiceInterfaceStateInterfaceVersion}
	return validateXServiceInterfaceEnumSet("state object", objects, required, IsXServiceInterfaceStateObject)
}

func validateXServiceInterfaceMessages(messages []XServiceInterfaceMessageName) error {
	required := []XServiceInterfaceMessageName{XServiceInterfaceMsgRegisterInterface, XServiceInterfaceMsgUpdateInterface, XServiceInterfaceMsgDeprecateInterface}
	return validateXServiceInterfaceEnumSet("message", messages, required, IsXServiceInterfaceMessageName)
}

func validateXServiceInterfaceQueries(queries []XServiceInterfaceQueryName) error {
	required := []XServiceInterfaceQueryName{XServiceInterfaceQueryInterface, XServiceInterfaceQueryMethod, XServiceInterfaceQueryInterfaceProof, XServiceInterfaceQueryInterfacesByOwner}
	return validateXServiceInterfaceEnumSet("query", queries, required, IsXServiceInterfaceQueryName)
}

func validateXServiceInterfaceFailureCoverage(coverage []XServiceInterfaceFailureCoverage) error {
	required := []XServiceInterfaceFailureMode{XServiceInterfaceFailureSchemaHashMismatch, XServiceInterfaceFailureMethodIDCollision, XServiceInterfaceFailureUnsupportedSchemaEncoding, XServiceInterfaceFailureBreakingUpdateWithoutNewHash}
	if len(coverage) != len(required) {
		return fmt.Errorf("x/serviceinterface expected %d failure modes", len(required))
	}
	seen := map[XServiceInterfaceFailureMode]struct{}{}
	for _, item := range coverage {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.Mode]; found {
			return fmt.Errorf("x/serviceinterface duplicate failure mode %s", item.Mode)
		}
		seen[item.Mode] = struct{}{}
	}
	for _, mode := range required {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("x/serviceinterface missing failure mode %s", mode)
		}
	}
	return nil
}

func validateXServiceInterfaceIntegrationPoints(points []XServiceInterfaceIntegrationPoint) error {
	required := []XServiceInterfaceIntegrationPoint{XServiceInterfaceIntegrationServices, XServiceInterfaceIntegrationWalletSDK, XServiceInterfaceIntegrationCLI, XServiceInterfaceIntegrationContractAdapter}
	return validateXServiceInterfaceEnumSet("integration", points, required, IsXServiceInterfaceIntegrationPoint)
}

func validateXServiceInterfaceEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/serviceinterface expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/serviceinterface unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/serviceinterface %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/serviceinterface duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/serviceinterface missing %s %s", label, value)
		}
	}
	return nil
}

func sortedXServiceInterfaceStateObjects(values []XServiceInterfaceStateObject) []XServiceInterfaceStateObject {
	out := append([]XServiceInterfaceStateObject(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceInterfaceMessages(values []XServiceInterfaceMessageName) []XServiceInterfaceMessageName {
	out := append([]XServiceInterfaceMessageName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceInterfaceQueries(values []XServiceInterfaceQueryName) []XServiceInterfaceQueryName {
	out := append([]XServiceInterfaceQueryName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceInterfaceFailureCoverage(values []XServiceInterfaceFailureCoverage) []XServiceInterfaceFailureCoverage {
	out := append([]XServiceInterfaceFailureCoverage(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].Scope = strings.Trim(strings.TrimSpace(out[i].Scope), "/")
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedXServiceInterfaceIntegrationPoints(values []XServiceInterfaceIntegrationPoint) []XServiceInterfaceIntegrationPoint {
	out := append([]XServiceInterfaceIntegrationPoint(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
