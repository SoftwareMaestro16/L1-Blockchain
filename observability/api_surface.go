package observability

import (
	"fmt"
	"sort"
)

const (
	RequiredAPIModuleStakingPolicy	= "aetra-staking-policy"
	RequiredAPIModuleEconomics	= "aetra-economics"
	RequiredAPIModuleValidatorScore	= "aetra-validator-score"
)

const (
	CommandCategoryQuery	= "query"
	CommandCategoryTx	= "tx"
)

const (
	RequiredAPISurfaceCLIQuery		= "cli_query"
	RequiredAPISurfaceCLITx			= "cli_tx"
	RequiredAPISurfaceProtobuf		= "protobuf_definition"
	RequiredAPISurfaceGRPCService		= "grpc_service"
	RequiredAPISurfaceGRPCQuery		= "grpc_query"
	RequiredAPISurfaceRESTGateway		= "rest_gateway_mapping_where_supported"
	RequiredAPISurfaceRESTQuery		= "rest_query_where_applicable"
	RequiredAPISurfaceEvents		= "events"
	RequiredAPISurfaceResponseExample	= "response_examples"
	RequiredAPISurfaceQueryTests		= "query_tests_where_feasible"
	RequiredAPISurfaceExamplesInDocs	= "examples_in_docs"
	RequiredAPISurfaceJSONOutput		= "json_output"
	RequiredAPISurfaceClearErrors		= "clear_errors"
	RequiredAPISurfaceHeightQuery		= "height_query_where_applicable"
	RequiredAPISurfacePagination		= "pagination_where_applicable"
	RequiredAPISurfaceBoundedAttrs		= "bounded_event_attributes"
	RequiredAPISurfaceStableResponses	= "stable_query_responses"
)

const (
	RequiredAPIEventValidatorCapCrossing		= "validator_cap_crossing"
	RequiredAPIEventDelegationOverflow		= "delegation_overflow"
	RequiredAPIEventRewardMultiplierChange		= "reward_multiplier_change"
	RequiredAPIEventFeeBurn				= "fee_burn"
	RequiredAPIEventTreasuryAllocation		= "treasury_allocation"
	RequiredAPIEventInflationUpdate			= "inflation_update"
	RequiredAPIEventAPREstimateEpochUpdate		= "apr_estimate_update_by_epoch"
	RequiredAPIEventValidatorScoreUpdate		= "validator_score_update"
	RequiredAPIEventDowntimeOffense			= "downtime_offense"
	RequiredAPIEventSlash				= "slash_event"
	RequiredAPIEventJailUnjail			= "jail_unjail"
	RequiredAPIEventGovernanceParamActivation	= "governance_param_activation"
)

const (
	RequiredAPIEventAttrValidator	= "validator"
	RequiredAPIEventAttrDelegator	= "delegator"
	RequiredAPIEventAttrAmount	= "amount"
	RequiredAPIEventAttrDenom	= "denom"
	RequiredAPIEventAttrHeight	= "height"
	RequiredAPIEventAttrEpoch	= "epoch"
	RequiredAPIEventAttrOldValue	= "old_value"
	RequiredAPIEventAttrNewValue	= "new_value"
	RequiredAPIEventAttrReason	= "reason"
	RequiredAPIEventAttrModule	= "module"
)

type CLICommandSpec struct {
	Module			string
	Category		string
	Command			string
	JSONOutput		bool
	HeightQuery		bool
	Pagination		bool
	ClearErrors		bool
	ExamplesInDocs		bool
	SignerValidation	bool
	AuthorityValidation	bool
}

type APISurfaceModuleSpec struct {
	Module			string
	CLICommands		[]CLICommandSpec
	ProtobufDefinition	bool
	GRPCService		bool
	GRPCQuery		bool
	RESTGatewayMapping	bool
	RESTQuery		bool
	Events			bool
	ResponseExamples	bool
	QueryTests		bool
	BoundedAttrs		bool
	StableResponses		bool
	ExamplesInDocs		bool
	Required		bool
}

type APISurfaceReadinessReport struct {
	Modules		[]APISurfaceModuleSpec
	RequiredCount	int
	ReadyCount	int
	Failed		[]string
	Ready		bool
}

type APIEventSpec struct {
	ID		string
	Module		string
	Attributes	[]string
	StableName	bool
	Bounded		bool
	Indexed		bool
	Tested		bool
}

type APIEventReadinessReport struct {
	Events		[]APIEventSpec
	RequiredCount	int
	ReadyCount	int
	Failed		[]string
	Ready		bool
}

func DefaultAPISurfaceModuleSpecs() []APISurfaceModuleSpec {
	return []APISurfaceModuleSpec{
		apiSurfaceModule(RequiredAPIModuleStakingPolicy, true),
		apiSurfaceModule(RequiredAPIModuleEconomics, true),
		apiSurfaceModule(RequiredAPIModuleValidatorScore, true),
	}
}

func ValidateAPISurfaceReadiness(modules []APISurfaceModuleSpec) error {
	report := BuildAPISurfaceReadinessReport(modules)
	if !report.Ready {
		return fmt.Errorf("api surface readiness failed: %v", report.Failed)
	}
	return nil
}

func DefaultAPIEventSpecs() []APIEventSpec {
	return []APIEventSpec{
		apiEvent(RequiredAPIEventValidatorCapCrossing, RequiredAPIModuleStakingPolicy),
		apiEvent(RequiredAPIEventDelegationOverflow, RequiredAPIModuleStakingPolicy),
		apiEvent(RequiredAPIEventRewardMultiplierChange, RequiredAPIModuleStakingPolicy),
		apiEvent(RequiredAPIEventFeeBurn, RequiredAPIModuleEconomics),
		apiEvent(RequiredAPIEventTreasuryAllocation, RequiredAPIModuleEconomics),
		apiEvent(RequiredAPIEventInflationUpdate, RequiredAPIModuleEconomics),
		apiEvent(RequiredAPIEventAPREstimateEpochUpdate, RequiredAPIModuleEconomics),
		apiEvent(RequiredAPIEventValidatorScoreUpdate, RequiredAPIModuleValidatorScore),
		apiEvent(RequiredAPIEventDowntimeOffense, RequiredAPIModuleValidatorScore),
		apiEvent(RequiredAPIEventSlash, "slashing"),
		apiEvent(RequiredAPIEventJailUnjail, "slashing"),
		apiEvent(RequiredAPIEventGovernanceParamActivation, "governance"),
	}
}

func ValidateAPIEventReadiness(events []APIEventSpec) error {
	report := BuildAPIEventReadinessReport(events)
	if !report.Ready {
		return fmt.Errorf("api event readiness failed: %v", report.Failed)
	}
	return nil
}

func BuildAPIEventReadinessReport(events []APIEventSpec) APIEventReadinessReport {
	if events == nil {
		events = DefaultAPIEventSpecs()
	}
	events = normalizeAPIEvents(events)
	requiredEvents := requiredAPIEvents()
	seen := map[string]APIEventSpec{}
	failed := make([]string, 0)
	requiredCount := 0
	readyCount := 0

	for _, event := range events {
		if event.ID == "" {
			failed = append(failed, "event_id_required")
			continue
		}
		if _, duplicate := seen[event.ID]; duplicate {
			failed = append(failed, event.ID+":duplicate_event")
		}
		seen[event.ID] = event
		if !requiredEvents[event.ID] {
			failed = append(failed, event.ID+":unknown_event")
		}
		requiredCount++
		eventFailures := validateAPIEvent(event)
		failed = append(failed, eventFailures...)
		if len(eventFailures) == 0 {
			readyCount++
		}
	}
	for id := range requiredEvents {
		if _, ok := seen[id]; !ok {
			failed = append(failed, id+":missing_event")
		}
	}

	sort.Strings(failed)
	return APIEventReadinessReport{
		Events:		events,
		RequiredCount:	requiredCount,
		ReadyCount:	readyCount,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func apiEvent(id, module string) APIEventSpec {
	return APIEventSpec{
		ID:		id,
		Module:		module,
		Attributes:	requiredAPIEventAttributes(),
		StableName:	true,
		Bounded:	true,
		Indexed:	true,
		Tested:		true,
	}
}

func BuildAPISurfaceReadinessReport(modules []APISurfaceModuleSpec) APISurfaceReadinessReport {
	if modules == nil {
		modules = DefaultAPISurfaceModuleSpecs()
	}
	modules = normalizeAPISurfaceModules(modules)
	required := requiredAPIModules()
	seen := map[string]APISurfaceModuleSpec{}
	failed := make([]string, 0)
	requiredCount := 0
	readyCount := 0

	for _, module := range modules {
		if module.Module == "" {
			failed = append(failed, "module_required")
			continue
		}
		if _, duplicate := seen[module.Module]; duplicate {
			failed = append(failed, module.Module+":duplicate_module")
		}
		seen[module.Module] = module
		if !required[module.Module] {
			failed = append(failed, module.Module+":unknown_module")
		}
		if module.Required {
			requiredCount++
		}
		moduleFailures := validateAPISurfaceModule(module)
		failed = append(failed, moduleFailures...)
		if module.Required && len(moduleFailures) == 0 {
			readyCount++
		}
	}
	for id := range required {
		if _, ok := seen[id]; !ok {
			failed = append(failed, id+":missing_module")
		}
	}

	sort.Strings(failed)
	return APISurfaceReadinessReport{
		Modules:	modules,
		RequiredCount:	requiredCount,
		ReadyCount:	readyCount,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func apiSurfaceModule(module string, rest bool) APISurfaceModuleSpec {
	return APISurfaceModuleSpec{
		Module:	module,
		CLICommands: []CLICommandSpec{
			apiSurfaceCLICommand(module, CommandCategoryQuery),
			apiSurfaceCLICommand(module, CommandCategoryTx),
		},
		ProtobufDefinition:	true,
		GRPCService:		true,
		GRPCQuery:		true,
		RESTGatewayMapping:	rest,
		RESTQuery:		rest,
		Events:			true,
		ResponseExamples:	true,
		QueryTests:		true,
		BoundedAttrs:		true,
		StableResponses:	true,
		ExamplesInDocs:		true,
		Required:		true,
	}
}

func apiSurfaceCLICommand(module, category string) CLICommandSpec {
	command := "aetrad " + category + " " + module + " ..."
	return CLICommandSpec{
		Module:			module,
		Category:		category,
		Command:		command,
		JSONOutput:		true,
		HeightQuery:		category == CommandCategoryQuery,
		Pagination:		category == CommandCategoryQuery,
		ClearErrors:		true,
		ExamplesInDocs:		true,
		SignerValidation:	category == CommandCategoryTx,
		AuthorityValidation:	category == CommandCategoryTx,
	}
}

func validateAPISurfaceModule(module APISurfaceModuleSpec) []string {
	failed := make([]string, 0)
	if module.Required {
		if !module.ProtobufDefinition {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceProtobuf+":missing")
		}
		if !module.GRPCService {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceGRPCService+":missing")
		}
		if !module.GRPCQuery {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceGRPCQuery+":missing")
		}
		if !module.RESTGatewayMapping {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceRESTGateway+":missing")
		}
		if !module.RESTQuery {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceRESTQuery+":missing")
		}
		if !module.Events {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceEvents+":missing")
		}
		if !module.ResponseExamples {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceResponseExample+":missing")
		}
		if !module.QueryTests {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceQueryTests+":missing")
		}
		if !module.BoundedAttrs {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceBoundedAttrs+":missing")
		}
		if !module.StableResponses {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceStableResponses+":missing")
		}
		if !module.ExamplesInDocs {
			failed = append(failed, module.Module+":"+RequiredAPISurfaceExamplesInDocs+":missing")
		}
	}

	commandsByCategory := map[string]CLICommandSpec{}
	for _, command := range module.CLICommands {
		if command.Module != module.Module {
			failed = append(failed, module.Module+":cli_command_module_mismatch")
		}
		if command.Category != CommandCategoryQuery && command.Category != CommandCategoryTx {
			failed = append(failed, module.Module+":"+command.Category+":unknown_cli_category")
		}
		if _, duplicate := commandsByCategory[command.Category]; duplicate {
			failed = append(failed, module.Module+":"+command.Category+":duplicate_cli_command")
		}
		commandsByCategory[command.Category] = command
		failed = append(failed, validateCLICommand(command)...)
	}
	for _, category := range []string{CommandCategoryQuery, CommandCategoryTx} {
		if _, ok := commandsByCategory[category]; !ok {
			failed = append(failed, module.Module+":cli_"+category+":missing")
		}
	}
	return failed
}

func validateCLICommand(command CLICommandSpec) []string {
	failed := make([]string, 0)
	prefix := command.Module + ":" + command.Category
	expected := "aetrad " + command.Category + " " + command.Module + " ..."
	if command.Command != expected {
		failed = append(failed, prefix+":unexpected_command")
	}
	if !command.JSONOutput {
		failed = append(failed, prefix+":"+RequiredAPISurfaceJSONOutput+":missing")
	}
	if !command.ClearErrors {
		failed = append(failed, prefix+":"+RequiredAPISurfaceClearErrors+":missing")
	}
	if !command.ExamplesInDocs {
		failed = append(failed, prefix+":"+RequiredAPISurfaceExamplesInDocs+":missing")
	}
	if command.Category == CommandCategoryQuery {
		if !command.HeightQuery {
			failed = append(failed, prefix+":"+RequiredAPISurfaceHeightQuery+":missing")
		}
		if !command.Pagination {
			failed = append(failed, prefix+":"+RequiredAPISurfacePagination+":missing")
		}
	}
	if command.Category == CommandCategoryTx {
		if !command.SignerValidation {
			failed = append(failed, prefix+":signer_validation:missing")
		}
		if !command.AuthorityValidation {
			failed = append(failed, prefix+":authority_validation:missing")
		}
	}
	return failed
}

func normalizeAPISurfaceModules(modules []APISurfaceModuleSpec) []APISurfaceModuleSpec {
	out := append([]APISurfaceModuleSpec{}, modules...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Module < out[j].Module })
	return out
}

func normalizeAPIEvents(events []APIEventSpec) []APIEventSpec {
	out := append([]APIEventSpec{}, events...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func validateAPIEvent(event APIEventSpec) []string {
	failed := make([]string, 0)
	if event.Module == "" {
		failed = append(failed, event.ID+":module:missing")
	}
	if !event.StableName {
		failed = append(failed, event.ID+":stable_name:missing")
	}
	if !event.Bounded {
		failed = append(failed, event.ID+":bounded_attributes:missing")
	}
	if !event.Indexed {
		failed = append(failed, event.ID+":indexer_compatible:missing")
	}
	if !event.Tested {
		failed = append(failed, event.ID+":tests:missing")
	}
	attrs := map[string]int{}
	for _, attr := range event.Attributes {
		attrs[attr]++
	}
	for _, attr := range requiredAPIEventAttributes() {
		if attrs[attr] == 0 {
			failed = append(failed, event.ID+":attr_"+attr+":missing")
		}
		if attrs[attr] > 1 {
			failed = append(failed, event.ID+":attr_"+attr+":duplicate")
		}
	}
	for attr := range attrs {
		if !requiredAPIEventAttributeSet()[attr] {
			failed = append(failed, event.ID+":attr_"+attr+":unexpected")
		}
	}
	return failed
}

func requiredAPIModules() map[string]bool {
	return map[string]bool{
		RequiredAPIModuleStakingPolicy:		true,
		RequiredAPIModuleEconomics:		true,
		RequiredAPIModuleValidatorScore:	true,
	}
}

func requiredAPIEvents() map[string]bool {
	return map[string]bool{
		RequiredAPIEventValidatorCapCrossing:		true,
		RequiredAPIEventDelegationOverflow:		true,
		RequiredAPIEventRewardMultiplierChange:		true,
		RequiredAPIEventFeeBurn:			true,
		RequiredAPIEventTreasuryAllocation:		true,
		RequiredAPIEventInflationUpdate:		true,
		RequiredAPIEventAPREstimateEpochUpdate:		true,
		RequiredAPIEventValidatorScoreUpdate:		true,
		RequiredAPIEventDowntimeOffense:		true,
		RequiredAPIEventSlash:				true,
		RequiredAPIEventJailUnjail:			true,
		RequiredAPIEventGovernanceParamActivation:	true,
	}
}

func requiredAPIEventAttributes() []string {
	return []string{
		RequiredAPIEventAttrValidator,
		RequiredAPIEventAttrDelegator,
		RequiredAPIEventAttrAmount,
		RequiredAPIEventAttrDenom,
		RequiredAPIEventAttrHeight,
		RequiredAPIEventAttrEpoch,
		RequiredAPIEventAttrOldValue,
		RequiredAPIEventAttrNewValue,
		RequiredAPIEventAttrReason,
		RequiredAPIEventAttrModule,
	}
}

func requiredAPIEventAttributeSet() map[string]bool {
	out := map[string]bool{}
	for _, attr := range requiredAPIEventAttributes() {
		out[attr] = true
	}
	return out
}
