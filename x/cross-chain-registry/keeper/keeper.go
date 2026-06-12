package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	RegistryParams	types.CrossChainRegistryParams
	State		types.CrossChainRegistryState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		prototype.DefaultParams(),
		RegistryParams:	types.DefaultCrossChainRegistryParams(),
		State:		types.EmptyCrossChainRegistryState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("cross-chain registry prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.RegistryParams)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) RegisterChain(msg types.MsgRegisterChain) (types.RegisteredChain, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.RegisteredChain{}, err
	}
	chain := msg.Chain.Normalize()
	if chain.RegisteredHeight == 0 {
		return types.RegisteredChain{}, errors.New("chain registration height must be positive")
	}
	chain.UpdatedHeight = chain.RegisteredHeight
	if _, _, found := chainIndex(k.genesis.State.Chains, chain.ChainID); found {
		return types.RegisteredChain{}, errors.New("chain already registered")
	}

	policy := msg.RiskPolicy.Normalize()
	if policy.ChainID == "" {
		policy.ChainID = chain.ChainID
	}
	hasPolicyInput := policy.PolicyID != "" || policy.MaxRiskScore != 0 || len(policy.AllowedClientTypes) != 0 || policy.MinTrustLevel != ""
	if hasPolicyInput {
		if policy.ChainID == "" {
			policy.ChainID = chain.ChainID
		}
		if chain.RiskPolicyID == "" {
			chain.RiskPolicyID = policy.PolicyID
		}
	}

	next := cloneGenesis(k.genesis)
	next.State.Chains = append(next.State.Chains, chain)
	if hasPolicyInput {
		next.State.RiskPolicies = append(next.State.RiskPolicies, policy)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.RegisteredChain{}, err
	}
	k.genesis = next
	return chain.Normalize(), nil
}

func (k *Keeper) UpdateChainStatus(msg types.MsgUpdateChainStatus) (types.RegisteredChain, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.RegisteredChain{}, err
	}
	if msg.Height == 0 {
		return types.RegisteredChain{}, errors.New("chain status update height must be positive")
	}
	if !types.IsChainStatus(msg.Status) {
		return types.RegisteredChain{}, errors.New("chain status is invalid")
	}
	idx, chain, found := chainIndex(k.genesis.State.Chains, msg.ChainID)
	if !found {
		return types.RegisteredChain{}, errors.New("chain not found")
	}
	chain.Status = msg.Status
	chain.UpdatedHeight = msg.Height

	next := cloneGenesis(k.genesis)
	next.State.Chains[idx] = chain.Normalize()
	if msg.Status != types.ChainStatusActive {
		for i := range next.State.BridgeRoutes {
			if next.State.BridgeRoutes[i].SourceChainID == chain.ChainID || next.State.BridgeRoutes[i].TargetChainID == chain.ChainID {
				next.State.BridgeRoutes[i].Enabled = false
			}
		}
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.RegisteredChain{}, err
	}
	k.genesis = next
	return chain.Normalize(), nil
}

func (k *Keeper) RegisterChannel(msg types.MsgRegisterChannel) (types.ChannelRecord, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ChannelRecord{}, err
	}
	channel := msg.Channel.Normalize()
	if _, _, found := channelIndex(k.genesis.State.Channels, channel.ChannelID); found {
		return types.ChannelRecord{}, errors.New("channel already registered")
	}
	chainIdx, chain, found := chainIndex(k.genesis.State.Chains, channel.ChainID)
	if !found {
		return types.ChannelRecord{}, errors.New("channel references unknown chain")
	}
	if channel.CounterpartyChainID != "" {
		if _, _, found := chainIndex(k.genesis.State.Chains, channel.CounterpartyChainID); !found {
			return types.ChannelRecord{}, errors.New("channel references unknown counterparty chain")
		}
	}
	routes := cloneRoutes(msg.Routes)
	for _, route := range routes {
		if _, _, found := routeIndex(k.genesis.State.BridgeRoutes, route.RouteID); found {
			return types.ChannelRecord{}, errors.New("bridge route already registered")
		}
		if route.ChannelID != channel.ChannelID {
			return types.ChannelRecord{}, errors.New("bridge route channel mismatch")
		}
	}

	next := cloneGenesis(k.genesis)
	next.State.Channels = append(next.State.Channels, channel)
	chain.ChannelIDs = append(chain.ChannelIDs, channel.ChannelID)
	next.State.Chains[chainIdx] = chain.Normalize()
	next.State.BridgeRoutes = append(next.State.BridgeRoutes, routes...)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ChannelRecord{}, err
	}
	k.genesis = next
	return channel, nil
}

func (k *Keeper) UpdateRiskPolicy(msg types.MsgUpdateRiskPolicy) (types.RiskPolicy, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.RiskPolicy{}, err
	}
	if msg.Height == 0 {
		return types.RiskPolicy{}, errors.New("risk policy update height must be positive")
	}
	chainIdx, chain, found := chainIndex(k.genesis.State.Chains, msg.ChainID)
	if !found {
		return types.RiskPolicy{}, errors.New("chain not found")
	}
	policy := msg.Policy.Normalize()
	if policy.ChainID == "" {
		policy.ChainID = chain.ChainID
	}
	if policy.ChainID != chain.ChainID {
		return types.RiskPolicy{}, errors.New("risk policy chain mismatch")
	}
	if policy.PolicyID == "" {
		policy.PolicyID = chain.ChainID + "-risk"
	}
	chain.RiskPolicyID = policy.PolicyID
	chain.UpdatedHeight = msg.Height

	next := cloneGenesis(k.genesis)
	next.State.Chains[chainIdx] = chain.Normalize()
	if policyIdx, _, found := policyIndex(next.State.RiskPolicies, policy.PolicyID); found {
		next.State.RiskPolicies[policyIdx] = policy
	} else {
		next.State.RiskPolicies = append(next.State.RiskPolicies, policy)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.RiskPolicy{}, err
	}
	k.genesis = next
	return policy, nil
}

func (k *Keeper) RemoveChain(msg types.MsgRemoveChain) error {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("chain removal height must be positive")
	}
	if _, _, found := chainIndex(k.genesis.State.Chains, msg.ChainID); !found {
		return errors.New("chain not found")
	}
	next := cloneGenesis(k.genesis)
	next.State.Chains = filterChains(next.State.Chains, msg.ChainID)
	next.State.Channels = filterChannels(next.State.Channels, msg.ChainID)
	next.State.BridgeRoutes = filterRoutes(next.State.BridgeRoutes, msg.ChainID)
	next.State.RiskPolicies = filterPolicies(next.State.RiskPolicies, msg.ChainID)
	for i := range next.State.Chains {
		next.State.Chains[i].ChannelIDs = filterStrings(next.State.Chains[i].ChannelIDs, removedChannelIDs(k.genesis.State.Channels, msg.ChainID))
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) RegisteredChain(chainID string) (types.RegisteredChain, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.RegisteredChain{}, false, err
	}
	_, chain, found := chainIndex(k.genesis.State.Chains, chainID)
	return chain, found, nil
}

func (k Keeper) RegisteredChains() ([]types.RegisteredChain, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.Export().Chains, nil
}

func (k Keeper) Channel(channelID string) (types.ChannelRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.ChannelRecord{}, false, err
	}
	_, channel, found := channelIndex(k.genesis.State.Channels, channelID)
	return channel, found, nil
}

func (k Keeper) BridgeRoute(routeID string) (types.BridgeRoute, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.BridgeRoute{}, false, err
	}
	_, route, found := routeIndex(k.genesis.State.BridgeRoutes, routeID)
	return route, found, nil
}

func (k Keeper) RiskPolicy(chainID string) (types.RiskPolicy, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.RiskPolicy{}, false, err
	}
	_, chain, found := chainIndex(k.genesis.State.Chains, chainID)
	if !found || chain.RiskPolicyID == "" {
		return types.RiskPolicy{}, false, nil
	}
	_, policy, found := policyIndex(k.genesis.State.RiskPolicies, chain.RiskPolicyID)
	return policy, found, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func chainIndex(chains []types.RegisteredChain, chainID string) (int, types.RegisteredChain, bool) {
	for i, chain := range chains {
		if chain.ChainID == chainID {
			return i, chain, true
		}
	}
	return -1, types.RegisteredChain{}, false
}

func channelIndex(channels []types.ChannelRecord, channelID string) (int, types.ChannelRecord, bool) {
	for i, channel := range channels {
		if channel.ChannelID == channelID {
			return i, channel, true
		}
	}
	return -1, types.ChannelRecord{}, false
}

func routeIndex(routes []types.BridgeRoute, routeID string) (int, types.BridgeRoute, bool) {
	for i, route := range routes {
		if route.RouteID == routeID {
			return i, route, true
		}
	}
	return -1, types.BridgeRoute{}, false
}

func policyIndex(policies []types.RiskPolicy, policyID string) (int, types.RiskPolicy, bool) {
	for i, policy := range policies {
		if policy.PolicyID == policyID {
			return i, policy, true
		}
	}
	return -1, types.RiskPolicy{}, false
}

func cloneRoutes(routes []types.BridgeRoute) []types.BridgeRoute {
	out := append([]types.BridgeRoute(nil), routes...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func filterChains(chains []types.RegisteredChain, chainID string) []types.RegisteredChain {
	out := make([]types.RegisteredChain, 0, len(chains))
	for _, chain := range chains {
		if chain.ChainID != chainID {
			out = append(out, chain)
		}
	}
	return out
}

func filterChannels(channels []types.ChannelRecord, chainID string) []types.ChannelRecord {
	out := make([]types.ChannelRecord, 0, len(channels))
	for _, channel := range channels {
		if channel.ChainID != chainID && channel.CounterpartyChainID != chainID {
			out = append(out, channel)
		}
	}
	return out
}

func filterRoutes(routes []types.BridgeRoute, chainID string) []types.BridgeRoute {
	out := make([]types.BridgeRoute, 0, len(routes))
	for _, route := range routes {
		if route.SourceChainID != chainID && route.TargetChainID != chainID {
			out = append(out, route)
		}
	}
	return out
}

func filterPolicies(policies []types.RiskPolicy, chainID string) []types.RiskPolicy {
	out := make([]types.RiskPolicy, 0, len(policies))
	for _, policy := range policies {
		if policy.ChainID != chainID {
			out = append(out, policy)
		}
	}
	return out
}

func removedChannelIDs(channels []types.ChannelRecord, chainID string) map[string]struct{} {
	removed := map[string]struct{}{}
	for _, channel := range channels {
		if channel.ChainID == chainID || channel.CounterpartyChainID == chainID {
			removed[channel.ChannelID] = struct{}{}
		}
	}
	return removed
}

func filterStrings(values []string, removed map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, found := removed[value]; !found {
			out = append(out, value)
		}
	}
	return out
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
