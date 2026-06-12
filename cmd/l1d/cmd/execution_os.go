package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	identitytypes "github.com/sovereign-l1/l1/x/identity/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	shardsim "github.com/sovereign-l1/l1/x/sharding/sim"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	executionOSProfileBase			= "base"
	executionOSProfileSim			= "execution-os-sim"
	executionOSProfileZonesPrototype	= "zones-prototype"
	executionOSProfileMeshPrototype		= "mesh-prototype"
	executionOSProfileIdentityPrototype	= "identity-prototype"
)

var executionOSProfiles = []string{
	executionOSProfileBase,
	executionOSProfileSim,
	executionOSProfileZonesPrototype,
	executionOSProfileMeshPrototype,
	executionOSProfileIdentityPrototype,
}

type executionOSReport struct {
	Profile		string			`json:"profile"`
	Load		executionOSLoadReport	`json:"load"`
	Routing		executionOSRouteReport	`json:"routing"`
	Sharding	executionOSShardReport	`json:"sharding"`
	Zones		executionOSZonesReport	`json:"zones"`
	Mesh		executionOSMeshReport	`json:"mesh"`
	Identity	executionOSIDReport	`json:"identity"`
	RestartSafe	bool			`json:"restart_safe"`
	FeatureGated	bool			`json:"feature_gated"`
	ProductionLive	bool			`json:"production_live"`
}

type executionOSLoadReport struct {
	ScoreBps	uint32	`json:"score_bps"`
	Band		string	`json:"band"`
	WindowHeight	uint64	`json:"window_height"`
}

type executionOSRouteReport struct {
	MsgType		string	`json:"msg_type"`
	TxClass		string	`json:"tx_class"`
	ZoneID		string	`json:"zone_id"`
	ShardID		uint32	`json:"shard_id"`
	ActiveShards	uint32	`json:"active_shards"`
}

type executionOSShardReport struct {
	WorkchainID		int32		`json:"workchain_id"`
	ActiveShardCount	uint32		`json:"active_shard_count"`
	ShardIDs		[]string	`json:"shard_ids"`
	RoutingEpoch		uint64		`json:"routing_epoch"`
	DataAvailable		bool		`json:"data_available"`
}

type executionOSZonesReport struct {
	ActiveZones	[]string	`json:"active_zones"`
	CommitmentRoots	[]string	`json:"commitment_roots"`
}

type executionOSMeshReport struct {
	MessageID		string	`json:"message_id"`
	ReceiptStatus		string	`json:"receipt_status"`
	ReceiptHash		string	`json:"receipt_hash"`
	ReplayMarkerCount	int	`json:"replay_marker_count"`
	PendingMessages		int	`json:"pending_messages"`
}

type executionOSIDReport struct {
	Domain		string	`json:"domain"`
	NFTID		string	`json:"nft_id"`
	ResolvedAddress	string	`json:"resolved_address"`
	Lifecycle	string	`json:"lifecycle"`
}

type executionOSDiagnostics struct {
	Profile			string			`json:"profile"`
	Source			string			`json:"source"`
	FeatureGates		map[string]featureGate	`json:"feature_gates"`
	CurrentLoadScoreBps	uint32			`json:"current_load_score_bps"`
	LoadWindowHeight	uint64			`json:"load_window_height"`
	ActiveZones		[]string		`json:"active_zones"`
	ActiveShards		[]zoneShardSummary	`json:"active_shards"`
	PendingMeshMessages	int			`json:"pending_mesh_messages"`
	ReplayMarkerCount	int			`json:"replay_marker_count"`
	MeshReceiptCount	int			`json:"mesh_receipt_count"`
	ZoneCommitmentRoots	[]string		`json:"zone_commitment_roots"`
	IdentityPrototype	bool			`json:"identity_prototype"`
	ProductionLive		bool			`json:"production_live"`
}

type featureGate struct {
	Enabled		bool	`json:"enabled"`
	TestnetProfile	bool	`json:"testnet_profile"`
}

type zoneShardSummary struct {
	ZoneID		string	`json:"zone_id"`
	ActiveShards	uint32	`json:"active_shards"`
}

func NewExecutionOSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"execution-os",
		Short:	"Aetra modular execution OS operator tools",
	}
	cmd.AddCommand(
		newExecutionOSProfilesCmd(),
		newExecutionOSSmokeCmd(),
		newExecutionOSDiagnosticsCmd(),
	)
	return cmd
}

func newExecutionOSProfilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:	"profiles",
		Short:	"List supported local execution OS profiles",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeJSON(cmd, struct {
				Profiles []string `json:"profiles"`
			}{Profiles: append([]string(nil), executionOSProfiles...)})
		},
	}
}

func newExecutionOSSmokeCmd() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:	"smoke",
		Short:	"Run a deterministic execution OS simulator smoke scenario",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateExecutionOSProfile(profile); err != nil {
				return err
			}
			report, err := buildExecutionOSSmokeReport(profile)
			if err != nil {
				return err
			}
			return writeJSON(cmd, report)
		},
	}
	cmd.Flags().StringVar(&profile, "profile", executionOSProfileSim, "execution OS profile to simulate")
	return cmd
}

func newExecutionOSDiagnosticsCmd() *cobra.Command {
	var profile string
	var genesisPath string
	cmd := &cobra.Command{
		Use:	"diagnostics",
		Short:	"Inspect execution OS prototype state from genesis or simulator defaults",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateExecutionOSProfile(profile); err != nil {
				return err
			}
			diagnostics, err := buildExecutionOSDiagnostics(profile, genesisPath)
			if err != nil {
				return err
			}
			return writeJSON(cmd, diagnostics)
		},
	}
	cmd.Flags().StringVar(&profile, "profile", executionOSProfileBase, "localnet execution OS profile")
	cmd.Flags().StringVar(&genesisPath, "genesis", "", "optional genesis.json path to inspect")
	return cmd
}

func buildExecutionOSSmokeReport(profile string) (executionOSReport, error) {
	loadResult, simState, err := runShardingSmoke()
	if err != nil {
		return executionOSReport{}, err
	}
	route, err := routingtypes.Route(routingtypes.RouteInput{
		MsgType:		routingtypes.MsgTypeBankSend,
		FeeDenom:		routingtypes.NativeFeeDenom,
		FeeClass:		99,
		ReputationClass:	99,
		AdmissionHeight:	12,
		TxHash:			hashBytes("operator-smoke-tx"),
		RoutingEpoch:		simState.LoadStates[shardsim.BaseWorkchain].RoutingEpoch,
		ActiveShards: map[routingtypes.ZoneID]uint32{
			routingtypes.ZoneFinancial: uint32(len(simState.Shards)),
		},
		Locality: routingtypes.Locality{
			AccountKey:	[]byte("operator-account"),
			AssetDenom:	"naet",
		},
	})
	if err != nil {
		return executionOSReport{}, err
	}
	zoneState, err := buildZonesSmokeState()
	if err != nil {
		return executionOSReport{}, err
	}
	meshState, meshMsg, meshReceipt, err := runMeshSmoke()
	if err != nil {
		return executionOSReport{}, err
	}
	identityState, domain, resolved, err := runIdentitySmoke()
	if err != nil {
		return executionOSReport{}, err
	}
	lifecycle, err := identitytypes.DomainLifecycle(identityState, domain.Name, domain.RegisteredHeight+2)
	if err != nil {
		return executionOSReport{}, err
	}
	loadState := simState.LoadStates[shardsim.BaseWorkchain]
	return executionOSReport{
		Profile:	profile,
		Load: executionOSLoadReport{
			ScoreBps:	loadResult.LoadScoreBps,
			Band:		string(loadResult.Band),
			WindowHeight:	loadResult.EMA.WindowHeight,
		},
		Routing: executionOSRouteReport{
			MsgType:	routingtypes.MsgTypeBankSend,
			TxClass:	string(route.TxClass),
			ZoneID:		string(route.ZoneID),
			ShardID:	uint32(route.ShardID),
			ActiveShards:	route.ActiveShards,
		},
		Sharding: executionOSShardReport{
			WorkchainID:		shardsim.BaseWorkchain,
			ActiveShardCount:	loadState.ActiveShardCount,
			ShardIDs:		shardKeys(simState),
			RoutingEpoch:		loadState.RoutingEpoch,
			DataAvailable:		allShardsAvailable(simState),
		},
		Zones: executionOSZonesReport{
			ActiveZones:		zoneIDs(zoneState.ActiveZones),
			CommitmentRoots:	zoneCommitmentRoots(zoneState.Commitments),
		},
		Mesh: executionOSMeshReport{
			MessageID:		meshMsg.MessageID,
			ReceiptStatus:		string(meshReceipt.Status),
			ReceiptHash:		meshReceipt.ReceiptHash,
			ReplayMarkerCount:	len(meshState.ReplayMarkers),
			PendingMessages:	0,
		},
		Identity: executionOSIDReport{
			Domain:			domain.Name,
			NFTID:			domain.NFTID,
			ResolvedAddress:	hex.EncodeToString(resolved),
			Lifecycle:		string(lifecycle),
		},
		RestartSafe:	true,
		FeatureGated:	profile != executionOSProfileBase,
		ProductionLive:	false,
	}, nil
}

func buildExecutionOSDiagnostics(profile, genesisPath string) (executionOSDiagnostics, error) {
	diag := executionOSDiagnostics{
		Profile:	profile,
		Source:		"simulator-defaults",
		FeatureGates: map[string]featureGate{
			"load":		{},
			"routing":	{},
			"zones":	{},
			"mesh":		{},
		},
		IdentityPrototype:	profile == executionOSProfileIdentityPrototype,
		ProductionLive:		false,
	}
	if genesisPath == "" {
		return diag, nil
	}
	bz, err := os.ReadFile(genesisPath)
	if err != nil {
		return executionOSDiagnostics{}, err
	}
	var genesis struct {
		AppState map[string]json.RawMessage `json:"app_state"`
	}
	if err := json.Unmarshal(bz, &genesis); err != nil {
		return executionOSDiagnostics{}, err
	}
	diag.Source = genesisPath
	if raw := genesis.AppState["load"]; len(raw) > 0 {
		var gs loadkeeper.GenesisState
		if err := json.Unmarshal(raw, &gs); err != nil {
			return executionOSDiagnostics{}, err
		}
		if err := gs.Validate(); err != nil {
			return executionOSDiagnostics{}, err
		}
		diag.FeatureGates["load"] = featureGate{Enabled: gs.Params.Enabled, TestnetProfile: gs.Params.TestnetProfile}
		diag.CurrentLoadScoreBps = gs.EMA.LoadScoreBps
		diag.LoadWindowHeight = gs.EMA.WindowHeight
	}
	if raw := genesis.AppState["routing"]; len(raw) > 0 {
		var gs routingkeeper.GenesisState
		if err := json.Unmarshal(raw, &gs); err != nil {
			return executionOSDiagnostics{}, err
		}
		if err := gs.Validate(); err != nil {
			return executionOSDiagnostics{}, err
		}
		diag.FeatureGates["routing"] = featureGate{Enabled: gs.Params.Enabled, TestnetProfile: gs.Params.TestnetProfile}
		diag.ActiveShards = make([]zoneShardSummary, len(gs.Shards))
		for i, shard := range gs.Shards {
			diag.ActiveShards[i] = zoneShardSummary{ZoneID: string(shard.ZoneID), ActiveShards: shard.ActiveShards}
		}
	}
	if raw := genesis.AppState["zones"]; len(raw) > 0 {
		var gs zoneskeeper.GenesisState
		if err := json.Unmarshal(raw, &gs); err != nil {
			return executionOSDiagnostics{}, err
		}
		if err := gs.Validate(); err != nil {
			return executionOSDiagnostics{}, err
		}
		diag.FeatureGates["zones"] = featureGate{Enabled: gs.Params.Enabled, TestnetProfile: gs.Params.TestnetProfile}
		diag.ActiveZones = zoneIDs(gs.State.ActiveZones)
		diag.ZoneCommitmentRoots = zoneCommitmentRoots(gs.State.Commitments)
	}
	if raw := genesis.AppState["mesh"]; len(raw) > 0 {
		var gs meshkeeper.GenesisState
		if err := json.Unmarshal(raw, &gs); err != nil {
			return executionOSDiagnostics{}, err
		}
		if err := gs.Validate(); err != nil {
			return executionOSDiagnostics{}, err
		}
		diag.FeatureGates["mesh"] = featureGate{Enabled: gs.Params.Enabled, TestnetProfile: gs.Params.TestnetProfile}
		diag.ReplayMarkerCount = len(gs.State.ReplayMarkers)
		diag.MeshReceiptCount = len(gs.State.Receipts)
	}
	sort.SliceStable(diag.ActiveShards, func(i, j int) bool {
		return diag.ActiveShards[i].ZoneID < diag.ActiveShards[j].ZoneID
	})
	return diag, nil
}

func runShardingSmoke() (loadtypes.Result, shardsim.MasterchainState, error) {
	sim, err := shardsim.New([]shardsim.Validator{
		{Address: "val-a", Power: 100},
		{Address: "val-b", Power: 100},
		{Address: "val-c", Power: 100},
	}, "operator-smoke")
	if err != nil {
		return loadtypes.Result{}, shardsim.MasterchainState{}, err
	}
	if err := sim.AddWorkchain(shardsim.WorkchainConfig{
		ID:			shardsim.BaseWorkchain,
		AllowedVMs:		[]string{"AVM", "COSMWASM_GATED", "NATIVE_MODULE"},
		FeeDenom:		shardsim.FeeDenomNaet,
		AddressFormat:		"ae",
		GenesisStateHash:	hashString("operator-workchain-genesis"),
		UpgradePolicy:		"GOVERNANCE",
	}); err != nil {
		return loadtypes.Result{}, shardsim.MasterchainState{}, err
	}
	params := loadtypes.DefaultParams()
	params.AlphaNumerator = 1
	params.AlphaDenominator = 1
	params.MaxDeltaBps = loadtypes.BasisPoints
	policy := shardsim.ShardActivationPolicy{
		WorkchainID:		shardsim.BaseWorkchain,
		LoadParams:		params,
		PartialShardCount:	2,
		MaxShardCount:		4,
		CooldownBlocks:		2,
		RoutingEpoch:		1,
	}
	transition, err := sim.UpdateLoadAndShards(policy, loadtypes.Metrics{
		CanonicalMempoolSize:		params.TargetMempoolSize,
		UsedBlockGas:			params.TargetBlockGas,
		AverageInclusionDelayBlocks:	params.TargetLatencyBlocks,
		FailedTxCount:			1,
		TotalTxCount:			1,
		ExecutionStepCount:		params.TargetExecutionSteps,
	}, 10)
	if err != nil {
		return loadtypes.Result{}, shardsim.MasterchainState{}, err
	}
	return transition.LoadResult, sim.Export(), nil
}

func buildZonesSmokeState() (zonestypes.ZoneRegistryState, error) {
	state := zonestypes.EmptyState()
	for _, zone := range []zonestypes.Zone{
		operatorZone(zonestypes.ZoneIDApplication, zonestypes.ZoneKindApplication, zonestypes.VMPolicyAVM),
		operatorZone(zonestypes.ZoneIDContract, zonestypes.ZoneKindContract, zonestypes.VMPolicyCosmWasmGated),
		operatorZone(zonestypes.ZoneIDFinancial, zonestypes.ZoneKindFinancial, zonestypes.VMPolicyNativeModule),
		operatorZone(zonestypes.ZoneIDIdentity, zonestypes.ZoneKindIdentity, zonestypes.VMPolicyNativeModule),
	} {
		next, err := zonestypes.RegisterZone(state, zone)
		if err != nil {
			return zonestypes.ZoneRegistryState{}, err
		}
		state = next
	}
	for _, id := range []zonestypes.ZoneID{
		zonestypes.ZoneIDApplication,
		zonestypes.ZoneIDContract,
		zonestypes.ZoneIDFinancial,
		zonestypes.ZoneIDIdentity,
	} {
		next, err := zonestypes.ActivateZone(state, id, 1)
		if err != nil {
			return zonestypes.ZoneRegistryState{}, err
		}
		state = next
	}
	first, err := zonestypes.NewZoneCommitment(
		zonestypes.ZoneIDFinancial,
		1,
		hashString("financial-state-1"),
		hashString("financial-receipt-1"),
		hashString("financial-message-1"),
		hashString("financial-execution-1"),
		"",
	)
	if err != nil {
		return zonestypes.ZoneRegistryState{}, err
	}
	return zonestypes.AppendCommitment(state, first)
}

func runMeshSmoke() (meshtypes.MeshState, meshtypes.MeshMessage, meshtypes.MeshReceipt, error) {
	state := meshtypes.EmptyState(meshtypes.DefaultParams())
	var err error
	state, err = meshtypes.RegisterDestination(state, meshtypes.MeshDestination{ZoneID: "CONTRACT_ZONE", ShardID: "0:1", Active: true})
	if err != nil {
		return meshtypes.MeshState{}, meshtypes.MeshMessage{}, meshtypes.MeshReceipt{}, err
	}
	state, err = meshtypes.RegisterDestination(state, meshtypes.MeshDestination{ZoneID: "FINANCIAL_ZONE", ShardID: "0:0", Active: true})
	if err != nil {
		return meshtypes.MeshState{}, meshtypes.MeshMessage{}, meshtypes.MeshReceipt{}, err
	}
	commitment := meshtypes.FinalizedCommitment{
		ZoneID:		"FINANCIAL_ZONE",
		ShardID:	"0:0",
		Height:		90,
		CommitmentHash:	meshtypes.HashParts("source-commitment", "financial", "0:0", "90"),
		MessageRoot:	meshtypes.HashParts("message-root", "financial", "90"),
		ReceiptRoot:	meshtypes.HashParts("receipt-root", "financial", "90"),
	}
	state, err = meshtypes.AddFinalizedCommitment(state, commitment)
	if err != nil {
		return meshtypes.MeshState{}, meshtypes.MeshMessage{}, meshtypes.MeshReceipt{}, err
	}
	msg, err := meshtypes.NewMessage(meshtypes.MeshMessage{
		SourceZone:		"FINANCIAL_ZONE",
		SourceShard:		"0:0",
		DestinationZone:	"CONTRACT_ZONE",
		DestinationShard:	"0:1",
		Nonce:			7,
		Sender:			[]byte("operator-sender"),
		Recipient:		[]byte("operator-contract"),
		AssetCommitment:	meshtypes.HashParts("asset", "100naet"),
		PayloadHash:		meshtypes.HashParts("payload", "execute"),
		TimeoutHeight:		150,
		Finality:		meshtypes.FinalityReference{Height: commitment.Height, CommitmentHash: commitment.CommitmentHash},
		Sequence:		3,
		SourceLogicalTime:	88,
	})
	if err != nil {
		return meshtypes.MeshState{}, meshtypes.MeshMessage{}, meshtypes.MeshReceipt{}, err
	}
	msg.Proof = meshtypes.BuildProof(msg, commitment)
	next, receipt, err := meshtypes.ApplyMessage(state, msg, meshtypes.ExecutionResult{
		Success:	true,
		Code:		0,
		ResultHash:	meshtypes.HashParts("execution", "success"),
	}, 100)
	if err != nil {
		return meshtypes.MeshState{}, meshtypes.MeshMessage{}, meshtypes.MeshReceipt{}, err
	}
	return next, msg, receipt, nil
}

func runIdentitySmoke() (identitytypes.IdentityState, identitytypes.Domain, sdk.AccAddress, error) {
	owner := sampleAddress(1)
	primary := sampleAddress(2)
	state := identitytypes.EmptyIdentityState(identitytypes.DefaultIdentityParams())
	commitment, err := identitytypes.ComputeRegistrationCommitment("operator.aet", owner, "operator-salt")
	if err != nil {
		return identitytypes.IdentityState{}, identitytypes.Domain{}, nil, err
	}
	state, err = identitytypes.CommitDomainRegistration(state, "operator.aet", owner, commitment, 10)
	if err != nil {
		return identitytypes.IdentityState{}, identitytypes.Domain{}, nil, err
	}
	state, domain, err := identitytypes.RevealRegisterDomain(state, "operator.aet", owner, "operator-salt", 11)
	if err != nil {
		return identitytypes.IdentityState{}, identitytypes.Domain{}, nil, err
	}
	state, _, err = identitytypes.SetIdentityResolver(state, domain.Name, owner, identitytypes.ResolverUpdate{Primary: primary}, 12)
	if err != nil {
		return identitytypes.IdentityState{}, identitytypes.Domain{}, nil, err
	}
	resolved, err := identitytypes.ResolveIdentityAddress(state, domain.Name, 13)
	if err != nil {
		return identitytypes.IdentityState{}, identitytypes.Domain{}, nil, err
	}
	return state, domain, resolved, nil
}

func operatorZone(id zonestypes.ZoneID, kind zonestypes.ZoneKind, vm zonestypes.VMPolicy) zonestypes.Zone {
	return zonestypes.Zone{
		ID:			id,
		Kind:			kind,
		VMPolicy:		vm,
		FeePolicy:		zonestypes.FeePolicyNaet,
		GenesisStateHash:	hashString(string(id) + "-genesis"),
		StateTransitionID:	"transition-" + string(id),
		UpgradePolicy:		zonestypes.UpgradePolicyGovernance,
		DataAvailabilityPolicy:	zonestypes.DataAvailabilityCoreCommitment,
		AuditStatus:		zonestypes.AuditStatusExperimental,
		ActivationHeight:	1,
	}
}

func shardKeys(state shardsim.MasterchainState) []string {
	keys := make([]string, 0, len(state.Shards))
	for key := range state.Shards {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func allShardsAvailable(state shardsim.MasterchainState) bool {
	if len(state.Shards) == 0 {
		return false
	}
	for _, shard := range state.Shards {
		if !shard.Available {
			return false
		}
	}
	return true
}

func zoneIDs(ids []zonestypes.ZoneID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func zoneCommitmentRoots(commitments []zonestypes.ZoneCommitment) []string {
	out := make([]string, 0, len(commitments))
	for _, commitment := range commitments {
		out = append(out, commitment.CommitmentHash)
	}
	sort.Strings(out)
	return out
}

func sampleAddress(seed byte) sdk.AccAddress {
	out := make([]byte, 20)
	out[19] = seed
	return sdk.AccAddress(out)
}

func hashBytes(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func hashString(value string) string {
	return hex.EncodeToString(hashBytes(value))
}

func validateExecutionOSProfile(profile string) error {
	for _, allowed := range executionOSProfiles {
		if profile == allowed {
			return nil
		}
	}
	return fmt.Errorf("unknown execution OS profile %q", profile)
}

func writeJSON(cmd *cobra.Command, value any) error {
	bz, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if len(bz) == 0 {
		return errors.New("empty JSON output")
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bz))
	return err
}
