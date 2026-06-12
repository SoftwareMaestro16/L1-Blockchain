package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	StatusActive		= "active"
	StatusPaused		= "paused"
	StatusDeprecated	= "deprecated"

	EventTypeRegistered	= "registered"
	EventTypeUpdated	= "updated"
	EventTypePaused		= "paused"
	EventTypeResumed	= "resumed"
	EventTypeDeprecated	= "deprecated"

	MaxEntitiesV1		= uint32(256)
	MaxCapabilitiesV1	= uint32(64)
	MaxDependenciesV1	= uint32(64)
	MaxModuleNameBytesV1	= uint32(96)
	MaxCapabilityBytesV1	= uint32(128)
	MaxDependencyEdgesV1	= uint32(1_024)
	DefaultModuleVersion	= uint64(1)
	ModuleConfig		= "config"
	ModuleConstitution	= "constitution"
)

type Params struct {
	Authority		string
	MaxEntities		uint32
	MaxCapabilities		uint32
	MaxDependencies		uint32
	MaxModuleNameBytes	uint32
	MaxCapabilityBytes	uint32
	MaxDependencyEdges	uint32
	RequiredModules		[]string
}

type SystemEntity struct {
	Name					string
	ModuleName				string
	ModuleAccountAddress			string
	RawAddress				string
	UserFriendlyAddress			string
	AuthorityAddress			string
	Status					string
	Core					bool
	CanHoldFunds				bool
	CanReceiveUserFunds			bool
	CanSendFunds				bool
	Capabilities				[]string
	Version					uint64
	Dependencies				[]string
	Required				bool
	PrivilegedCallsAllowedWhilePaused	bool
}

type State struct {
	Entities		[]SystemEntity
	UserControlledAccounts	[]string
}

type DependencyEdge struct {
	ModuleName	string
	DependsOn	string
}

type SystemEntityEvent struct {
	Type		string
	ModuleName	string
	Status		string
	Height		uint64
}

type MsgRegisterSystemEntity struct {
	Authority	string
	Entity		SystemEntity
}

type MsgUpdateSystemEntity struct {
	Authority	string
	Entity		SystemEntity
}

type MsgPauseSystemEntity struct {
	Authority			string
	ModuleName			string
	Height				uint64
	AllowPrivilegedCallsWhilePaused	bool
}

type MsgResumeSystemEntity struct {
	Authority	string
	ModuleName	string
	Height		uint64
}

type MsgDeprecateSystemEntity struct {
	Authority	string
	ModuleName	string
	Height		uint64
}

func DefaultParams() Params {
	requiredModules := make([]string, 0, len(addressing.AllSystemAddresses()))
	for _, address := range addressing.AllSystemAddresses() {
		requiredModules = append(requiredModules, address.ModuleName)
	}
	sort.Strings(requiredModules)
	return Params{
		Authority:		prototype.DefaultAuthority,
		MaxEntities:		MaxEntitiesV1,
		MaxCapabilities:	MaxCapabilitiesV1,
		MaxDependencies:	MaxDependenciesV1,
		MaxModuleNameBytes:	MaxModuleNameBytesV1,
		MaxCapabilityBytes:	MaxCapabilityBytesV1,
		MaxDependencyEdges:	MaxDependencyEdgesV1,
		RequiredModules:	requiredModules,
	}
}

func DefaultState() State {
	authority := prototype.DefaultAuthority
	entities := make([]SystemEntity, 0, len(addressing.AllSystemAddresses()))
	for _, address := range addressing.AllSystemAddresses() {
		entities = append(entities, SystemEntity{
			Name:			address.Name,
			ModuleName:		address.ModuleName,
			ModuleAccountAddress:	address.Raw,
			RawAddress:		address.Raw,
			UserFriendlyAddress:	address.UserFriendly,
			AuthorityAddress:	authority,
			Status:			StatusActive,
			Core:			address.Core,
			CanHoldFunds:		address.CanHoldFunds,
			CanReceiveUserFunds:	address.CanReceiveUserFunds,
			CanSendFunds:		address.CanSendFunds,
			Version:		DefaultModuleVersion,
			Required:		true,
		})
	}
	return State{Entities: SortEntities(entities)}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("system registry authority", p.Authority); err != nil {
		return err
	}
	if p.MaxEntities == 0 || p.MaxEntities > MaxEntitiesV1 {
		return fmt.Errorf("system registry max entities must be between 1 and %d", MaxEntitiesV1)
	}
	if p.MaxCapabilities == 0 || p.MaxCapabilities > MaxCapabilitiesV1 {
		return fmt.Errorf("system registry max capabilities must be between 1 and %d", MaxCapabilitiesV1)
	}
	if p.MaxDependencies == 0 || p.MaxDependencies > MaxDependenciesV1 {
		return fmt.Errorf("system registry max dependencies must be between 1 and %d", MaxDependenciesV1)
	}
	if p.MaxModuleNameBytes == 0 || p.MaxModuleNameBytes > MaxModuleNameBytesV1 {
		return fmt.Errorf("system registry max module name bytes must be between 1 and %d", MaxModuleNameBytesV1)
	}
	if p.MaxCapabilityBytes == 0 || p.MaxCapabilityBytes > MaxCapabilityBytesV1 {
		return fmt.Errorf("system registry max capability bytes must be between 1 and %d", MaxCapabilityBytesV1)
	}
	if p.MaxDependencyEdges == 0 || p.MaxDependencyEdges > MaxDependencyEdgesV1 {
		return fmt.Errorf("system registry max dependency edges must be between 1 and %d", MaxDependencyEdgesV1)
	}
	if len(p.RequiredModules) == 0 {
		return errors.New("system registry required modules must be non-empty")
	}
	return validateSortedUniqueTokens("system registry required module", p.RequiredModules, p.MaxModuleNameBytes)
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("system registry update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("system registry update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := addressing.ValidateReservedSystemAddressCatalog(); err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}
	if len(s.Entities) == 0 {
		return errors.New("system registry entities must be non-empty")
	}
	if uint32(len(s.Entities)) > params.MaxEntities {
		return errors.New("system registry entity limit exceeded")
	}

	byName := make(map[string]SystemEntity, len(s.Entities))
	byAccount := make(map[string]string, len(s.Entities))
	edgeCount := uint32(0)
	for _, entity := range s.Entities {
		normalized := entity.Normalize(params)
		if err := normalized.Validate(params); err != nil {
			return err
		}
		if _, found := byName[normalized.ModuleName]; found {
			return fmt.Errorf("system registry duplicate module %q", normalized.ModuleName)
		}
		accountKey, err := addressing.AddressTextBytesKey(normalized.ModuleAccountAddress)
		if err != nil {
			return fmt.Errorf("system registry module account bytes invalid for %q: %w", normalized.ModuleName, err)
		}
		if other, found := byAccount[accountKey]; found {
			return fmt.Errorf("system registry duplicate module account %q used by %s and %s", normalized.ModuleAccountAddress, other, normalized.ModuleName)
		}
		byName[normalized.ModuleName] = normalized
		byAccount[accountKey] = normalized.ModuleName
		edgeCount += uint32(len(normalized.Dependencies))
	}
	if edgeCount > params.MaxDependencyEdges {
		return errors.New("system registry dependency edge limit exceeded")
	}
	for _, required := range params.RequiredModules {
		entity, found := byName[required]
		if !found {
			return fmt.Errorf("system registry required module %q is missing", required)
		}
		if !entity.Required || entity.Status != StatusActive {
			return fmt.Errorf("system registry required module %q must be active", required)
		}
	}
	for _, entity := range byName {
		for _, dependency := range entity.Dependencies {
			if dependency == entity.ModuleName {
				return fmt.Errorf("system registry module %q cannot depend on itself", entity.ModuleName)
			}
			if _, found := byName[dependency]; !found {
				return fmt.Errorf("system registry dependency %q for module %q is not registered", dependency, entity.ModuleName)
			}
		}
	}
	if err := validateAcyclic(byName); err != nil {
		return err
	}
	if err := ValidateReservedSystemRegistryState(s.Normalize(params)); err != nil {
		return err
	}
	return addressing.ValidateNoUserControlledSystemAddresses(s.UserControlledAccounts)
}

func (e SystemEntity) Validate(params Params) error {
	e.Name = strings.TrimSpace(e.Name)
	if e.Name != "" && len(e.Name) > int(params.MaxModuleNameBytes) {
		return fmt.Errorf("system registry entity name exceeds %d bytes", params.MaxModuleNameBytes)
	}
	if err := validateToken("system registry module name", e.ModuleName, params.MaxModuleNameBytes); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("system registry module account", e.ModuleAccountAddress); err != nil {
		return err
	}
	if strings.TrimSpace(e.RawAddress) != "" {
		if err := addressing.ValidateAuthorityAddress("system registry raw address", e.RawAddress); err != nil {
			return err
		}
	}
	if strings.TrimSpace(e.UserFriendlyAddress) != "" {
		if err := addressing.ValidateAuthorityAddress("system registry user-friendly address", e.UserFriendlyAddress); err != nil {
			return err
		}
	}
	if err := addressing.ValidateAuthorityAddress("system registry entity authority", e.AuthorityAddress); err != nil {
		return err
	}
	if !IsStatus(e.Status) {
		return fmt.Errorf("system registry status %q is invalid", e.Status)
	}
	if e.Version == 0 {
		return errors.New("system registry module version must be positive")
	}
	if uint32(len(e.Capabilities)) > params.MaxCapabilities {
		return errors.New("system registry capability limit exceeded")
	}
	if uint32(len(e.Dependencies)) > params.MaxDependencies {
		return errors.New("system registry dependency limit exceeded")
	}
	if err := validateSortedUniqueTokens("system registry capability", e.Capabilities, params.MaxCapabilityBytes); err != nil {
		return err
	}
	if err := validateSortedUniqueTokens("system registry dependency", e.Dependencies, params.MaxModuleNameBytes); err != nil {
		return err
	}
	if e.Status != StatusPaused && e.PrivilegedCallsAllowedWhilePaused {
		return errors.New("system registry privileged paused calls flag requires paused status")
	}
	return nil
}

func (e SystemEntity) Normalize(params Params) SystemEntity {
	e.Name = strings.TrimSpace(e.Name)
	e.ModuleName = strings.TrimSpace(e.ModuleName)
	e.ModuleAccountAddress = strings.TrimSpace(e.ModuleAccountAddress)
	e.RawAddress = strings.TrimSpace(e.RawAddress)
	e.UserFriendlyAddress = strings.TrimSpace(e.UserFriendlyAddress)
	e.AuthorityAddress = strings.TrimSpace(e.AuthorityAddress)
	e.Status = strings.TrimSpace(e.Status)
	if e.Status == "" {
		e.Status = StatusActive
	}
	if e.Version == 0 {
		e.Version = DefaultModuleVersion
	}
	e.Capabilities = sortedUniqueTokens(e.Capabilities)
	e.Dependencies = sortedUniqueTokens(e.Dependencies)
	if contains(params.RequiredModules, e.ModuleName) {
		e.Required = true
	}
	if e.Status != StatusPaused {
		e.PrivilegedCallsAllowedWhilePaused = false
	}
	return e
}

func (s State) Normalize(params Params) State {
	out := State{
		Entities:		make([]SystemEntity, 0, len(s.Entities)),
		UserControlledAccounts:	append([]string(nil), s.UserControlledAccounts...),
	}
	for _, entity := range s.Entities {
		out.Entities = append(out.Entities, entity.Normalize(params))
	}
	out.Entities = SortEntities(out.Entities)
	return out
}

func (s State) Entity(moduleName string) (SystemEntity, bool) {
	moduleName = strings.TrimSpace(moduleName)
	for _, entity := range SortEntities(s.Entities) {
		if entity.ModuleName == moduleName {
			return entity, true
		}
	}
	return SystemEntity{}, false
}

func (s State) DependencyGraph() []DependencyEdge {
	edges := []DependencyEdge{}
	for _, entity := range SortEntities(s.Entities) {
		for _, dependency := range sortedUniqueTokens(entity.Dependencies) {
			edges = append(edges, DependencyEdge{ModuleName: entity.ModuleName, DependsOn: dependency})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].ModuleName == edges[j].ModuleName {
			return edges[i].DependsOn < edges[j].DependsOn
		}
		return edges[i].ModuleName < edges[j].ModuleName
	})
	return edges
}

func ValidateReservedSystemRegistryState(state State) error {
	byReservedName := make(map[string]SystemEntity, len(state.Entities))
	for _, entity := range state.Entities {
		if entity.Name != "" {
			byReservedName[entity.Name] = entity
		}
	}

	seenBytes := map[string]string{}
	for _, expected := range addressing.AllSystemAddresses() {
		entity, found := byReservedName[expected.Name]
		if !found {
			return fmt.Errorf("system registry required system entity %q is missing", expected.Name)
		}
		if err := validateReservedSystemEntity(entity, expected); err != nil {
			return err
		}
		key, err := addressing.SystemAddressBytesKey(expected)
		if err != nil {
			return fmt.Errorf("system registry reserved address %q is invalid: %w", expected.Name, err)
		}
		if other, found := seenBytes[key]; found {
			return fmt.Errorf("system registry duplicate reserved address bytes for %q and %q", other, expected.Name)
		}
		seenBytes[key] = expected.Name
	}

	for _, entity := range state.Entities {
		if entity.RawAddress == "" {
			continue
		}
		key, err := addressing.AddressTextBytesKey(entity.RawAddress)
		if err != nil {
			return fmt.Errorf("system registry raw address for %q is invalid: %w", entity.ModuleName, err)
		}
		if expectedName, found := seenBytes[key]; found && entity.Name != expectedName {
			return fmt.Errorf("system registry duplicate reserved address bytes for %q and %q", expectedName, entity.Name)
		}
	}
	return nil
}

func validateReservedSystemEntity(entity SystemEntity, expected addressing.SystemAddress) error {
	if entity.ModuleName != expected.ModuleName {
		return fmt.Errorf("system registry entity %q module_name mismatch: expected %q got %q", expected.Name, expected.ModuleName, entity.ModuleName)
	}
	if entity.ModuleAccountAddress != expected.Raw {
		return fmt.Errorf("system registry entity %q module account mismatch: expected %q got %q", expected.Name, expected.Raw, entity.ModuleAccountAddress)
	}
	if entity.RawAddress != expected.Raw {
		return fmt.Errorf("system registry entity %q raw mismatch: expected %q got %q", expected.Name, expected.Raw, entity.RawAddress)
	}
	if entity.UserFriendlyAddress != expected.UserFriendly {
		return fmt.Errorf("system registry entity %q user_friendly mismatch: expected %q got %q", expected.Name, expected.UserFriendly, entity.UserFriendlyAddress)
	}
	if entity.Core != expected.Core {
		return fmt.Errorf("system registry entity %q core policy mismatch", expected.Name)
	}
	if entity.CanHoldFunds != expected.CanHoldFunds {
		return fmt.Errorf("system registry entity %q can_hold_funds policy mismatch", expected.Name)
	}
	if entity.CanReceiveUserFunds != expected.CanReceiveUserFunds {
		return fmt.Errorf("system registry entity %q can_receive_user_funds policy mismatch", expected.Name)
	}
	if entity.CanSendFunds != expected.CanSendFunds {
		return fmt.Errorf("system registry entity %q can_send_funds policy mismatch", expected.Name)
	}
	if entity.Status != StatusActive || entity.Status != expected.Status {
		return fmt.Errorf("system registry entity %q status must be active", expected.Name)
	}
	if !entity.Required {
		return fmt.Errorf("system registry entity %q must be required", expected.Name)
	}
	rawKey, err := addressing.AddressTextBytesKey(entity.RawAddress)
	if err != nil {
		return fmt.Errorf("system registry entity %q raw invalid: %w", expected.Name, err)
	}
	ufKey, err := addressing.AddressTextBytesKey(entity.UserFriendlyAddress)
	if err != nil {
		return fmt.Errorf("system registry entity %q user_friendly invalid: %w", expected.Name, err)
	}
	if rawKey != ufKey {
		return fmt.Errorf("system registry entity %q raw and user_friendly bytes mismatch", expected.Name)
	}
	rawBytes, err := addressing.Parse(entity.RawAddress)
	if err != nil {
		return fmt.Errorf("system registry entity %q raw invalid: %w", expected.Name, err)
	}
	if addressing.IsZero(rawBytes) {
		return fmt.Errorf("system registry entity %q must not use zero address", expected.Name)
	}
	return nil
}

func SortEntities(entities []SystemEntity) []SystemEntity {
	out := make([]SystemEntity, len(entities))
	copy(out, entities)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ModuleName < out[j].ModuleName
	})
	return out
}

func UpsertEntity(entities []SystemEntity, entity SystemEntity) []SystemEntity {
	next := make([]SystemEntity, 0, len(entities)+1)
	replaced := false
	for _, current := range entities {
		if current.ModuleName == entity.ModuleName {
			next = append(next, entity)
			replaced = true
			continue
		}
		next = append(next, current)
	}
	if !replaced {
		next = append(next, entity)
	}
	return SortEntities(next)
}

func IsStatus(status string) bool {
	switch status {
	case StatusActive, StatusPaused, StatusDeprecated:
		return true
	default:
		return false
	}
}

func PrivilegedCallAllowed(entity SystemEntity) bool {
	return entity.Status == StatusActive || (entity.Status == StatusPaused && entity.PrivilegedCallsAllowedWhilePaused)
}

func validateAcyclic(byName map[string]SystemEntity) error {
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var walk func(string, []string) error
	walk = func(moduleName string, path []string) error {
		if visiting[moduleName] {
			return fmt.Errorf("system registry dependency graph cycle detected at %q", moduleName)
		}
		if visited[moduleName] {
			return nil
		}
		visiting[moduleName] = true
		entity := byName[moduleName]
		for _, dependency := range entity.Dependencies {
			if err := walk(dependency, append(path, moduleName)); err != nil {
				return err
			}
		}
		visiting[moduleName] = false
		visited[moduleName] = true
		return nil
	}
	names := make([]string, 0, len(byName))
	for moduleName := range byName {
		names = append(names, moduleName)
	}
	sort.Strings(names)
	for _, moduleName := range names {
		if err := walk(moduleName, nil); err != nil {
			return err
		}
	}
	return nil
}

func validateToken(label, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s must be non-empty", label)
	}
	if uint32(len(value)) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", label, maxBytes)
	}
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' || ch == ':' {
			continue
		}
		return fmt.Errorf("%s contains invalid character %q", label, ch)
	}
	return nil
}

func validateSortedUniqueTokens(label string, values []string, maxBytes uint32) error {
	normalized := sortedUniqueTokens(values)
	if len(normalized) != len(values) {
		return fmt.Errorf("%s values must be unique", label)
	}
	for i, value := range values {
		if value != normalized[i] {
			return fmt.Errorf("%s values must be sorted deterministically", label)
		}
		if err := validateToken(label, value, maxBytes); err != nil {
			return err
		}
	}
	return nil
}

func sortedUniqueTokens(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func contains(values []string, value string) bool {
	for _, current := range values {
		if current == value {
			return true
		}
	}
	return false
}
