package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ChainStatusActive	= "active"
	ChainStatusPaused	= "paused"
	ChainStatusDisabled	= "disabled"

	ClientTypeIBC		= "ibc"
	ClientTypeLightClient	= "light_client"
	ClientTypeBridge	= "bridge"
	ClientTypeExternal	= "external"

	TrustLevelLow		= "low"
	TrustLevelMedium	= "medium"
	TrustLevelHigh		= "high"
)

type CrossChainRegistryParams struct {
	MaxChains		uint32
	MaxChannels		uint32
	MaxRoutes		uint32
	MaxRiskPolicies		uint32
	MaxRiskScore		uint32
	MaxFinalityBlocks	uint64
	MaxFinalityTimeSeconds	uint64
	MaxTimeoutBlocks	uint64
	MaxTimeoutSeconds	uint64
}

type CrossChainRegistryState struct {
	Chains		[]RegisteredChain
	Channels	[]ChannelRecord
	BridgeRoutes	[]BridgeRoute
	RiskPolicies	[]RiskPolicy
}

type RegisteredChain struct {
	ChainID			string
	Status			string
	ClientType		string
	TrustLevel		string
	ChannelIDs		[]string
	LightClientRef		string
	RiskScore		uint32
	Finality		FinalityAssumptions
	Timeout			TimeoutParameters
	RiskPolicyID		string
	RegisteredHeight	uint64
	UpdatedHeight		uint64
}

type FinalityAssumptions struct {
	ConfirmationBlocks	uint64
	FinalitySeconds		uint64
}

type TimeoutParameters struct {
	TimeoutBlocks	uint64
	TimeoutSeconds	uint64
}

type ChannelRecord struct {
	ChannelID		string
	ChainID			string
	CounterpartyChainID	string
	ClientID		string
	Active			bool
	RegisteredHeight	uint64
}

type BridgeRoute struct {
	RouteID		string
	SourceChainID	string
	TargetChainID	string
	ChannelID	string
	BridgeID	string
	Enabled		bool
}

type RiskPolicy struct {
	PolicyID		string
	ChainID			string
	MaxRiskScore		uint32
	AllowedClientTypes	[]string
	MinTrustLevel		string
	RequireLightClient	bool
}

type MsgRegisterChain struct {
	Authority	string
	Chain		RegisteredChain
	RiskPolicy	RiskPolicy
}

type MsgUpdateChainStatus struct {
	Authority	string
	ChainID		string
	Status		string
	Height		uint64
}

type MsgRegisterChannel struct {
	Authority	string
	Channel		ChannelRecord
	Routes		[]BridgeRoute
}

type MsgUpdateRiskPolicy struct {
	Authority	string
	ChainID		string
	Policy		RiskPolicy
	Height		uint64
}

type MsgRemoveChain struct {
	Authority	string
	ChainID		string
	Height		uint64
}

func DefaultCrossChainRegistryParams() CrossChainRegistryParams {
	return CrossChainRegistryParams{
		MaxChains:		4_096,
		MaxChannels:		100_000,
		MaxRoutes:		100_000,
		MaxRiskPolicies:	4_096,
		MaxRiskScore:		100,
		MaxFinalityBlocks:	1_000_000,
		MaxFinalityTimeSeconds:	30 * 24 * 60 * 60,
		MaxTimeoutBlocks:	1_000_000,
		MaxTimeoutSeconds:	30 * 24 * 60 * 60,
	}
}

func EmptyCrossChainRegistryState() CrossChainRegistryState {
	return CrossChainRegistryState{
		Chains:		[]RegisteredChain{},
		Channels:	[]ChannelRecord{},
		BridgeRoutes:	[]BridgeRoute{},
		RiskPolicies:	[]RiskPolicy{},
	}
}

func (p CrossChainRegistryParams) Validate() error {
	if p.MaxChains == 0 || p.MaxChannels == 0 || p.MaxRoutes == 0 || p.MaxRiskPolicies == 0 {
		return errors.New("cross-chain registry limits must be positive")
	}
	if p.MaxRiskScore == 0 || p.MaxFinalityBlocks == 0 || p.MaxFinalityTimeSeconds == 0 || p.MaxTimeoutBlocks == 0 || p.MaxTimeoutSeconds == 0 {
		return errors.New("cross-chain registry bounds must be positive")
	}
	return nil
}

func (s CrossChainRegistryState) Export() CrossChainRegistryState {
	out := CrossChainRegistryState{
		Chains:		cloneChains(s.Chains),
		Channels:	cloneChannels(s.Channels),
		BridgeRoutes:	cloneRoutes(s.BridgeRoutes),
		RiskPolicies:	clonePolicies(s.RiskPolicies),
	}
	SortChains(out.Chains)
	SortChannels(out.Channels)
	SortBridgeRoutes(out.BridgeRoutes)
	SortRiskPolicies(out.RiskPolicies)
	return out
}

func (s CrossChainRegistryState) Validate(params CrossChainRegistryParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Chains)) > params.MaxChains {
		return errors.New("cross-chain registry chain count exceeds limit")
	}
	if uint32(len(s.Channels)) > params.MaxChannels {
		return errors.New("cross-chain registry channel count exceeds limit")
	}
	if uint32(len(s.BridgeRoutes)) > params.MaxRoutes {
		return errors.New("cross-chain registry route count exceeds limit")
	}
	if uint32(len(s.RiskPolicies)) > params.MaxRiskPolicies {
		return errors.New("cross-chain registry risk policy count exceeds limit")
	}

	chains := map[string]RegisteredChain{}
	for _, chain := range s.Chains {
		if err := chain.Validate(params); err != nil {
			return err
		}
		if _, found := chains[chain.ChainID]; found {
			return fmt.Errorf("duplicate chain id %q", chain.ChainID)
		}
		chains[chain.ChainID] = chain
	}

	policies := map[string]RiskPolicy{}
	for _, policy := range s.RiskPolicies {
		if err := policy.Validate(params); err != nil {
			return err
		}
		if _, found := chains[policy.ChainID]; !found {
			return fmt.Errorf("risk policy %q references unknown chain %q", policy.PolicyID, policy.ChainID)
		}
		if _, found := policies[policy.PolicyID]; found {
			return fmt.Errorf("duplicate risk policy id %q", policy.PolicyID)
		}
		policies[policy.PolicyID] = policy
	}

	channels := map[string]ChannelRecord{}
	for _, channel := range s.Channels {
		if err := channel.Validate(); err != nil {
			return err
		}
		if _, found := chains[channel.ChainID]; !found {
			return fmt.Errorf("channel %q references unknown chain %q", channel.ChannelID, channel.ChainID)
		}
		if channel.CounterpartyChainID != "" {
			if _, found := chains[channel.CounterpartyChainID]; !found {
				return fmt.Errorf("channel %q references unknown counterparty chain %q", channel.ChannelID, channel.CounterpartyChainID)
			}
		}
		if _, found := channels[channel.ChannelID]; found {
			return fmt.Errorf("duplicate channel id %q", channel.ChannelID)
		}
		channels[channel.ChannelID] = channel
	}

	for _, chain := range s.Chains {
		if chain.Status != ChainStatusActive {
			continue
		}
		if chain.RiskPolicyID == "" {
			return fmt.Errorf("risk policy cannot be empty for active chain %q", chain.ChainID)
		}
		policy, found := policies[chain.RiskPolicyID]
		if !found || policy.ChainID != chain.ChainID || policy.IsEmpty() {
			return fmt.Errorf("risk policy cannot be empty for active chain %q", chain.ChainID)
		}
		if chain.RiskScore > policy.MaxRiskScore {
			return fmt.Errorf("chain %q risk score exceeds policy", chain.ChainID)
		}
		if policy.RequireLightClient && chain.LightClientRef == "" {
			return fmt.Errorf("chain %q risk policy requires light client reference", chain.ChainID)
		}
		if !policy.AllowsClientType(chain.ClientType) {
			return fmt.Errorf("chain %q client type is not allowed by risk policy", chain.ChainID)
		}
		if trustRank(chain.TrustLevel) < trustRank(policy.MinTrustLevel) {
			return fmt.Errorf("chain %q trust level is below risk policy", chain.ChainID)
		}
	}

	seenRoutes := map[string]struct{}{}
	for _, route := range s.BridgeRoutes {
		if err := route.Validate(); err != nil {
			return err
		}
		if _, found := seenRoutes[route.RouteID]; found {
			return fmt.Errorf("duplicate bridge route id %q", route.RouteID)
		}
		seenRoutes[route.RouteID] = struct{}{}
		source, found := chains[route.SourceChainID]
		if !found {
			return fmt.Errorf("bridge route %q references unknown source chain %q", route.RouteID, route.SourceChainID)
		}
		target, found := chains[route.TargetChainID]
		if !found {
			return fmt.Errorf("bridge route %q references unknown target chain %q", route.RouteID, route.TargetChainID)
		}
		channel, found := channels[route.ChannelID]
		if !found {
			return fmt.Errorf("bridge route %q references unknown channel %q", route.RouteID, route.ChannelID)
		}
		if route.Enabled && (source.Status != ChainStatusActive || target.Status != ChainStatusActive || !channel.Active) {
			return fmt.Errorf("active bridge route %q requires active chain and channel", route.RouteID)
		}
	}
	return nil
}

func (c RegisteredChain) Normalize() RegisteredChain {
	c.ChainID = strings.TrimSpace(c.ChainID)
	c.Status = strings.TrimSpace(c.Status)
	if c.Status == "" {
		c.Status = ChainStatusPaused
	}
	c.ClientType = strings.TrimSpace(c.ClientType)
	c.TrustLevel = strings.TrimSpace(c.TrustLevel)
	if c.TrustLevel == "" {
		c.TrustLevel = TrustLevelLow
	}
	c.ChannelIDs = normalizeStrings(c.ChannelIDs)
	c.LightClientRef = strings.TrimSpace(c.LightClientRef)
	c.RiskPolicyID = strings.TrimSpace(c.RiskPolicyID)
	return c
}

func (c RegisteredChain) Validate(params CrossChainRegistryParams) error {
	c = c.Normalize()
	if c.ChainID == "" {
		return errors.New("chain id is required")
	}
	if !IsChainStatus(c.Status) {
		return errors.New("chain status is invalid")
	}
	if !IsClientType(c.ClientType) {
		return errors.New("chain client type is invalid")
	}
	if !IsTrustLevel(c.TrustLevel) {
		return errors.New("chain trust level is invalid")
	}
	if c.RiskScore > params.MaxRiskScore {
		return errors.New("chain risk score exceeds maximum")
	}
	if err := c.Finality.Validate(params); err != nil {
		return err
	}
	if err := c.Timeout.Validate(params); err != nil {
		return err
	}
	if c.RegisteredHeight == 0 || c.UpdatedHeight == 0 {
		return errors.New("chain heights must be positive")
	}
	return nil
}

func (f FinalityAssumptions) Validate(params CrossChainRegistryParams) error {
	if f.ConfirmationBlocks == 0 || f.FinalitySeconds == 0 {
		return errors.New("finality parameters must be positive")
	}
	if f.ConfirmationBlocks > params.MaxFinalityBlocks || f.FinalitySeconds > params.MaxFinalityTimeSeconds {
		return errors.New("finality parameters exceed bounds")
	}
	return nil
}

func (t TimeoutParameters) Validate(params CrossChainRegistryParams) error {
	if t.TimeoutBlocks == 0 || t.TimeoutSeconds == 0 {
		return errors.New("timeout parameters must be positive")
	}
	if t.TimeoutBlocks > params.MaxTimeoutBlocks || t.TimeoutSeconds > params.MaxTimeoutSeconds {
		return errors.New("timeout parameters exceed bounds")
	}
	return nil
}

func (c ChannelRecord) Normalize() ChannelRecord {
	c.ChannelID = strings.TrimSpace(c.ChannelID)
	c.ChainID = strings.TrimSpace(c.ChainID)
	c.CounterpartyChainID = strings.TrimSpace(c.CounterpartyChainID)
	c.ClientID = strings.TrimSpace(c.ClientID)
	return c
}

func (c ChannelRecord) Validate() error {
	c = c.Normalize()
	if c.ChannelID == "" || c.ChainID == "" || c.ClientID == "" {
		return errors.New("channel id, chain id, and client id are required")
	}
	if c.RegisteredHeight == 0 {
		return errors.New("channel registered height must be positive")
	}
	return nil
}

func (r BridgeRoute) Normalize() BridgeRoute {
	r.RouteID = strings.TrimSpace(r.RouteID)
	r.SourceChainID = strings.TrimSpace(r.SourceChainID)
	r.TargetChainID = strings.TrimSpace(r.TargetChainID)
	r.ChannelID = strings.TrimSpace(r.ChannelID)
	r.BridgeID = strings.TrimSpace(r.BridgeID)
	return r
}

func (r BridgeRoute) Validate() error {
	r = r.Normalize()
	if r.RouteID == "" || r.SourceChainID == "" || r.TargetChainID == "" || r.ChannelID == "" || r.BridgeID == "" {
		return errors.New("bridge route identifiers are required")
	}
	if r.SourceChainID == r.TargetChainID {
		return errors.New("bridge route source and target chains must differ")
	}
	return nil
}

func (p RiskPolicy) Normalize() RiskPolicy {
	p.PolicyID = strings.TrimSpace(p.PolicyID)
	p.ChainID = strings.TrimSpace(p.ChainID)
	p.AllowedClientTypes = normalizeStrings(p.AllowedClientTypes)
	p.MinTrustLevel = strings.TrimSpace(p.MinTrustLevel)
	return p
}

func (p RiskPolicy) Validate(params CrossChainRegistryParams) error {
	p = p.Normalize()
	if p.PolicyID == "" || p.ChainID == "" {
		return errors.New("risk policy id and chain id are required")
	}
	if p.MaxRiskScore == 0 || p.MaxRiskScore > params.MaxRiskScore {
		return errors.New("risk policy max risk score is invalid")
	}
	if len(p.AllowedClientTypes) == 0 {
		return errors.New("risk policy allowed client types cannot be empty")
	}
	for _, clientType := range p.AllowedClientTypes {
		if !IsClientType(clientType) {
			return errors.New("risk policy client type is invalid")
		}
	}
	if !IsTrustLevel(p.MinTrustLevel) {
		return errors.New("risk policy minimum trust level is invalid")
	}
	return nil
}

func (p RiskPolicy) IsEmpty() bool {
	p = p.Normalize()
	return p.PolicyID == "" || p.ChainID == "" || p.MaxRiskScore == 0 || len(p.AllowedClientTypes) == 0 || p.MinTrustLevel == ""
}

func (p RiskPolicy) AllowsClientType(clientType string) bool {
	for _, allowed := range p.Normalize().AllowedClientTypes {
		if allowed == clientType {
			return true
		}
	}
	return false
}

func IsChainStatus(status string) bool {
	switch status {
	case ChainStatusActive, ChainStatusPaused, ChainStatusDisabled:
		return true
	default:
		return false
	}
}

func IsClientType(clientType string) bool {
	switch clientType {
	case ClientTypeIBC, ClientTypeLightClient, ClientTypeBridge, ClientTypeExternal:
		return true
	default:
		return false
	}
}

func IsTrustLevel(trustLevel string) bool {
	switch trustLevel {
	case TrustLevelLow, TrustLevelMedium, TrustLevelHigh:
		return true
	default:
		return false
	}
}

func SortChains(chains []RegisteredChain) {
	sort.SliceStable(chains, func(i, j int) bool { return chains[i].ChainID < chains[j].ChainID })
}

func SortChannels(channels []ChannelRecord) {
	sort.SliceStable(channels, func(i, j int) bool {
		if channels[i].ChainID != channels[j].ChainID {
			return channels[i].ChainID < channels[j].ChainID
		}
		return channels[i].ChannelID < channels[j].ChannelID
	})
}

func SortBridgeRoutes(routes []BridgeRoute) {
	sort.SliceStable(routes, func(i, j int) bool { return routes[i].RouteID < routes[j].RouteID })
}

func SortRiskPolicies(policies []RiskPolicy) {
	sort.SliceStable(policies, func(i, j int) bool { return policies[i].PolicyID < policies[j].PolicyID })
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func trustRank(trustLevel string) int {
	switch trustLevel {
	case TrustLevelHigh:
		return 3
	case TrustLevelMedium:
		return 2
	case TrustLevelLow:
		return 1
	default:
		return 0
	}
}

func cloneChains(chains []RegisteredChain) []RegisteredChain {
	out := append([]RegisteredChain(nil), chains...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneChannels(channels []ChannelRecord) []ChannelRecord {
	out := append([]ChannelRecord(nil), channels...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneRoutes(routes []BridgeRoute) []BridgeRoute {
	out := append([]BridgeRoute(nil), routes...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func clonePolicies(policies []RiskPolicy) []RiskPolicy {
	out := append([]RiskPolicy(nil), policies...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}
