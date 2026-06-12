package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/accounts"
	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/app/params"
	economicstypes "github.com/sovereign-l1/l1/x/aetra-economics/types"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	AppInvariantBankSupply				= "bank_supply_matches_total_balances"
	AppInvariantModuleAccounts			= "module_accounts_exist_and_match_permissions"
	AppInvariantValidatorSet			= "validator_set_matches_staking_state"
	AppInvariantSystemAddresses			= "system_addresses_are_reserved_and_blocked"
	AppInvariantFeeAccounting			= "fee_accounting_reconciles"
	AppInvariantStorageAccounting			= "storage_accounting_reconciles"
	AppInvariantGenesisExport			= "genesis_export_is_valid_and_deterministic"
	AppInvariantPoolSharesSum			= "pool_shares_sum_equals_total_shares"
	AppInvariantPoolRuntimeAccounting		= "pool_active_stake_rewards_unbondings_consistent"
	AppInvariantValidatorInsuranceBounds		= "validator_entry_self_stake_insurance_bounds"
	AppInvariantEconomicsFeeSplit			= "fee_split_sums_to_100_percent"
	AppInvariantEconomicsAccounting			= "burn_treasury_rewards_accounting"
	AppInvariantAVMQueueReceipts			= "avm_contract_queue_receipts_consistency"
	AppInvariantReservedSystemAddressPolicy		= "reserved_system_address_ownership_blocking"
	AppInvariantBankModuleAccounting		= nativeaccounttypes.InvariantModuleBankAccountingConsistent
	AppInvariantDirectUserDelegationRejected	= nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected
	AppInvariantProtocolCriticalStorageRent		= nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent
	AppInvariantSystemStorageReserveRunway		= nativeaccounttypes.InvariantSystemStorageReserveRunway
	AppInvariantProtocolCriticalExecutable		= nativeaccounttypes.InvariantProtocolCriticalExecutableUnderUnderfunding
	AppInvariantSystemRentTopUpBeforeUserWork	= nativeaccounttypes.InvariantSystemRentTopUpBeforeUserFreeze

	AppInvariantEmissionCap			= "emission_cap_total_minted_below_constitutional_max"
	AppInvariantBurnAccounting		= "burn_accounting_reconciles_denom_burned_vs_bank_supply"
	AppInvariantTreasuryAccounting		= "treasury_accounting_reconciles_bank_balance"
	AppInvariantFeeCollectorAccounting	= "fee_collector_accounting_reconciles_bank_balance"
	AppInvariantRentReserveBalance		= "storage_rent_reserve_balance_above_protocol_minimum"
	AppInvariantNoNativeAppAssetModules	= "no_native_app_asset_modules_wired"
)

// AppInvariant binds a consensus invariant ID to live application state.
type AppInvariant struct {
	ID		string
	Description	string
	Check		func(ctx sdk.Context) error
}

// AppInvariantFailure is a deterministic runtime invariant failure report.
type AppInvariantFailure struct {
	ID	string
	Error	string
}

// AppInvariantRouteRegistry is the app-owned invariant registry used by tests,
// CLI checks, and SDK-compatible route registration.
type AppInvariantRouteRegistry struct {
	routes map[string]sdk.Invariant
}

func NewAppInvariantRouteRegistry() *AppInvariantRouteRegistry {
	return &AppInvariantRouteRegistry{routes: map[string]sdk.Invariant{}}
}

func (r *AppInvariantRouteRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	key := invariantRouteKey(moduleName, route)
	r.routes[key] = invar
}

func (r *AppInvariantRouteRegistry) Routes() []string {
	out := make([]string, 0, len(r.routes))
	for route := range r.routes {
		out = append(out, route)
	}
	sort.Strings(out)
	return out
}

func (r *AppInvariantRouteRegistry) Run(ctx sdk.Context) []AppInvariantFailure {
	routes := r.Routes()
	failures := make([]AppInvariantFailure, 0)
	for _, route := range routes {
		msg, broken := r.routes[route](ctx)
		if broken {
			failures = append(failures, AppInvariantFailure{ID: route, Error: msg})
		}
	}
	return failures
}

func invariantRouteKey(moduleName, route string) string {
	return strings.TrimSpace(moduleName) + "/" + strings.TrimSpace(route)
}

func CriticalAppInvariantRoutes() []string {
	ids := RequiredAppInvariantIDs()
	routes := make([]string, 0, len(ids))
	for _, id := range ids {
		routes = append(routes, invariantRouteKey("aetra", id))
	}
	return routes
}

func RequiredAppInvariantIDs() []string {
	required := append([]string(nil), nativeaccounttypes.RequiredNativeAccountInvariantIDs()...)
	required = append(required,
		AppInvariantBankSupply,
		AppInvariantModuleAccounts,
		AppInvariantValidatorSet,
		AppInvariantSystemAddresses,
		AppInvariantFeeAccounting,
		AppInvariantStorageAccounting,
		AppInvariantGenesisExport,
		AppInvariantPoolSharesSum,
		AppInvariantPoolRuntimeAccounting,
		AppInvariantValidatorInsuranceBounds,
		AppInvariantEconomicsFeeSplit,
		AppInvariantEconomicsAccounting,
		AppInvariantAVMQueueReceipts,
		AppInvariantReservedSystemAddressPolicy,
		AppInvariantEmissionCap,
		AppInvariantBurnAccounting,
		AppInvariantTreasuryAccounting,
		AppInvariantFeeCollectorAccounting,
		AppInvariantRentReserveBalance,
		AppInvariantNoNativeAppAssetModules,
	)
	sort.Strings(required)
	return required
}

func (app *L1App) registerCriticalInvariantRoutes() {
	registry := NewAppInvariantRouteRegistry()
	app.ModuleManager.RegisterInvariants(registry)
	if err := app.RegisterAppInvariants(registry); err != nil {
		panic(err)
	}
	app.invariantRegistry = registry
}

func (app *L1App) CriticalInvariantRoutes() []string {
	if app.invariantRegistry == nil {
		return nil
	}
	return app.invariantRegistry.Routes()
}

func (app *L1App) RunCriticalInvariants(ctx sdk.Context) []AppInvariantFailure {
	if app.invariantRegistry == nil {
		return []AppInvariantFailure{{ID: "registry", Error: "critical invariant registry is not initialized"}}
	}
	return app.invariantRegistry.Run(ctx)
}

func ValidateAppInvariantRegistry(registry []AppInvariant) error {
	required := RequiredAppInvariantIDs()
	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		if strings.TrimSpace(invariant.ID) == "" || strings.TrimSpace(invariant.Description) == "" || invariant.Check == nil {
			return errors.New("app invariant id, description, and check are required")
		}
		if _, found := seen[invariant.ID]; found {
			return fmt.Errorf("duplicate app invariant %s", invariant.ID)
		}
		seen[invariant.ID] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("app invariant registry missing %s", id)
		}
	}
	return nil
}

func (app *L1App) AppInvariantRegistry() []AppInvariant {
	modelRegistry := nativeaccounttypes.DefaultNativeAccountInvariantRegistry()
	registry := make([]AppInvariant, 0, len(modelRegistry)+14)
	for _, modelInvariant := range modelRegistry {
		id := modelInvariant.ID
		registry = append(registry, AppInvariant{
			ID:		id,
			Description:	modelInvariant.Description,
			Check: func(ctx sdk.Context) error {
				return app.checkAppInvariant(ctx, id)
			},
		})
	}
	for _, invariant := range []AppInvariant{
		{ID: AppInvariantBankSupply, Description: "bank total supply must match every account and module-account balance"},
		{ID: AppInvariantModuleAccounts, Description: "module accounts must exist at deterministic module addresses with configured permissions"},
		{ID: AppInvariantValidatorSet, Description: "staking bonded validator set and last validator set must be bounded and internally consistent"},
		{ID: AppInvariantSystemAddresses, Description: "reserved AE/raw system addresses must be unique, canonical, and receive-blocked when required"},
		{ID: AppInvariantFeeAccounting, Description: "fee collection and economics accounting must reconcile to collected fees"},
		{ID: AppInvariantStorageAccounting, Description: "storage rent accounting must preserve protocol-critical state and reserve runway"},
		{ID: AppInvariantGenesisExport, Description: "app genesis export must validate and be deterministic across repeated exports"},
		{ID: AppInvariantPoolSharesSum, Description: "liquid staking pool share records must sum to each pool total shares"},
		{ID: AppInvariantPoolRuntimeAccounting, Description: "liquid staking active stake, rewards, unbondings, and allocations must reconcile"},
		{ID: AppInvariantValidatorInsuranceBounds, Description: "validator registry self-stake and active insurance bounds must hold"},
		{ID: AppInvariantEconomicsFeeSplit, Description: "economics fee split must sum to 100 percent"},
		{ID: AppInvariantEconomicsAccounting, Description: "burn, treasury, and validator reward accounting must reconcile by epoch"},
		{ID: AppInvariantAVMQueueReceipts, Description: "AVM contract state roots and internal message queue records must be canonical"},
		{ID: AppInvariantReservedSystemAddressPolicy, Description: "reserved system address mappings and receive blocking policy must hold"},
		{ID: AppInvariantEmissionCap, Description: "total minted accounting must not exceed constitutional max inflation"},
		{ID: AppInvariantBurnAccounting, Description: "burn keeper burned amount must reconcile with bank supply"},
		{ID: AppInvariantTreasuryAccounting, Description: "treasury module accounting state must match feecollector_treasury bank balance"},
		{ID: AppInvariantFeeCollectorAccounting, Description: "fee-collector module accounting state must match feecollector bank balance"},
		{ID: AppInvariantRentReserveBalance, Description: "feecollector_storage_rent_reserve bank balance must be >= protocol minimum"},
		{ID: AppInvariantNoNativeAppAssetModules, Description: "no native app-asset modules (future_avm_standard, prototype_only, disabled) are wired in the app module manager"},
	} {
		id := invariant.ID
		description := invariant.Description
		registry = append(registry, AppInvariant{
			ID:		id,
			Description:	description,
			Check: func(ctx sdk.Context) error {
				return app.checkAppInvariant(ctx, id)
			},
		})
	}
	sort.SliceStable(registry, func(i, j int) bool { return registry[i].ID < registry[j].ID })
	return registry
}

func (app *L1App) RegisterAppInvariants(registry sdk.InvariantRegistry) error {
	invariants := app.AppInvariantRegistry()
	if err := ValidateAppInvariantRegistry(invariants); err != nil {
		return err
	}
	for _, invariant := range invariants {
		id := invariant.ID
		check := invariant.Check
		registry.RegisterRoute("aetra", id, func(ctx sdk.Context) (string, bool) {
			if err := check(ctx); err != nil {
				return sdk.FormatInvariant("aetra", id, err.Error()), true
			}
			return "", false
		})
	}
	return nil
}

func (app *L1App) RunAppInvariant(ctx sdk.Context, id string) error {
	for _, invariant := range app.AppInvariantRegistry() {
		if invariant.ID == id {
			return invariant.Check(ctx)
		}
	}
	return fmt.Errorf("unknown app invariant %s", id)
}

func (app *L1App) RunAppInvariants(ctx sdk.Context) []AppInvariantFailure {
	registry := app.AppInvariantRegistry()
	failures := make([]AppInvariantFailure, 0)
	for _, invariant := range registry {
		if err := invariant.Check(ctx); err != nil {
			failures = append(failures, AppInvariantFailure{ID: invariant.ID, Error: err.Error()})
		}
	}
	sort.SliceStable(failures, func(i, j int) bool { return failures[i].ID < failures[j].ID })
	return failures
}

func (app *L1App) checkAppInvariant(ctx sdk.Context, id string) error {
	input, err := app.nativeInvariantInput(ctx)
	if err != nil {
		return err
	}
	if isNativeAccountInvariantID(id) {
		if err := nativeaccounttypes.RunNativeAccountInvariant(id, input); err != nil {
			return err
		}
	}

	switch id {
	case nativeaccounttypes.InvariantModuleBankAccountingConsistent:
		return app.assertBankModuleAccountingInvariant(ctx)
	case AppInvariantBankSupply:
		return app.assertBankSupplyInvariant(ctx)
	case AppInvariantModuleAccounts:
		return app.assertModuleAccountsInvariant(ctx)
	case AppInvariantValidatorSet:
		return app.assertValidatorSetInvariant(ctx)
	case AppInvariantSystemAddresses:
		return app.assertReservedSystemAddressInvariant(ctx)
	case AppInvariantFeeAccounting:
		return app.assertEconomicsAccountingInvariant(ctx)
	case AppInvariantStorageAccounting:
		return app.assertStorageRentInvariant(ctx)
	case AppInvariantGenesisExport:
		return app.assertGenesisExportInvariant(ctx)
	case AppInvariantPoolSharesSum:
		return app.assertPoolSharesSumInvariant(ctx)
	case AppInvariantPoolRuntimeAccounting:
		return app.assertPoolRuntimeAccountingInvariant(ctx)
	case AppInvariantEconomicsFeeSplit:
		return app.assertEconomicsFeeSplitInvariant(ctx)
	case AppInvariantEconomicsAccounting:
		return app.assertEconomicsAccountingInvariant(ctx)
	case AppInvariantAVMQueueReceipts:
		return app.assertAVMQueueReceiptsInvariant(ctx)
	case AppInvariantReservedSystemAddressPolicy:
		return app.assertReservedSystemAddressInvariant(ctx)
	case AppInvariantValidatorInsuranceBounds:
		return app.assertValidatorInsuranceBoundsInvariant(ctx)
	case AppInvariantEmissionCap:
		return app.assertEmissionCapInvariant(ctx)
	case AppInvariantBurnAccounting:
		return app.assertBurnAccountingInvariant(ctx)
	case AppInvariantTreasuryAccounting:
		return app.assertTreasuryAccountingInvariant(ctx)
	case AppInvariantFeeCollectorAccounting:
		return app.assertFeeCollectorAccountingInvariant(ctx)
	case AppInvariantRentReserveBalance:
		return app.assertRentReserveBalanceInvariant(ctx)
	case AppInvariantNoNativeAppAssetModules:
		return app.assertNoNativeAppAssetModulesInvariant(ctx)
	case nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected:
		return app.assertDirectUserDelegationRejectedInvariant(ctx)
	case nativeaccounttypes.InvariantMaxValidatorCountEnforced,
		nativeaccounttypes.InvariantMinValidatorStakeEnforced,
		nativeaccounttypes.InvariantValidatorSelfStakeRatioEnforced:
		return app.assertValidatorEntryInvariant(ctx)
	case nativeaccounttypes.InvariantPoolStakeDoesNotCreateCoins,
		nativeaccounttypes.InvariantRewardsCannotExceedAllocation,
		nativeaccounttypes.InvariantMinPoolDepositEnforced,
		nativeaccounttypes.InvariantUnbondingCannotReleaseEarly,
		nativeaccounttypes.InvariantReputationRequiresStakeTime,
		nativeaccounttypes.InvariantJailedValidatorNoPositiveBonus,
		nativeaccounttypes.InvariantExportImportPreservesState:
		return app.assertNominatorPoolInvariant(ctx)
	case nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent,
		nativeaccounttypes.InvariantSystemStorageReserveRunway,
		nativeaccounttypes.InvariantSystemRentTopUpBeforeUserFreeze,
		nativeaccounttypes.InvariantProtocolCriticalExecutableUnderUnderfunding:
		return app.assertStorageRentInvariant(ctx)
	default:
		return nil
	}
}

func isNativeAccountInvariantID(id string) bool {
	for _, nativeID := range nativeaccounttypes.RequiredNativeAccountInvariantIDs() {
		if id == nativeID {
			return true
		}
	}
	return false
}

func (app *L1App) nativeInvariantInput(ctx sdk.Context) (nativeaccounttypes.NativeAccountInvariantInput, error) {
	params, err := app.StakingKeeper.GetParams(ctx)
	if err != nil {
		return nativeaccounttypes.NativeAccountInvariantInput{}, err
	}
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nativeaccounttypes.NativeAccountInvariantInput{}, err
	}

	activeValidators := uint64(0)
	validatorStakes := make([]uint64, 0, len(validators))
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			activeValidators++
		}
		if validator.Tokens.IsPositive() {
			validatorStakes = append(validatorStakes, safeUint64(validator.Tokens, math.MaxUint64))
		}
	}

	supply := app.BankKeeper.GetSupply(ctx, params.BondDenom)
	totalSupply := safeUint64(supply.Amount, 0)

	minValidatorStake := uint64(0)

	emissionParams, emissionErr := app.EmissionsKeeper.GetParams(ctx)
	var rewardBudget uint64
	if emissionErr == nil && emissionParams.MaxAnnualInflationBps > 0 && emissionParams.EpochsPerYear > 0 {
		refSupply := safeUint64(emissionParams.AnnualReferenceSupply.Amount, 0)
		rewardBudget = (refSupply * uint64(emissionParams.MaxAnnualInflationBps) / 10000) / emissionParams.EpochsPerYear
	}

	srGenesis, srErr := app.StorageRentKeeper.ExportGenesisState(ctx)
	var systemReserveRunwayBlocks, minSystemReserveRunwayBlocks uint64
	protocolCriticalExecutable := true
	systemRentUnderfunded := true
	if srErr == nil {
		reserve := srGenesis.State.SystemReserve
		systemReserveRunwayBlocks = reserve.LastRunwayBlocks
		if systemReserveRunwayBlocks == 0 {
			systemReserveRunwayBlocks = reserve.WarningRunwayBlocks
		}
		minSystemReserveRunwayBlocks = reserve.CriticalRunwayBlocks
		protocolCriticalExecutable = reserve.ProtocolCriticalExecutable
		result := reserve.Evaluate()
		systemRentUnderfunded = result.Alert == storagerenttypes.SystemRentAlertInvariant ||
			result.Alert == storagerenttypes.SystemRentAlertCritical ||
			result.Alert == storagerenttypes.SystemRentAlertWarning
	}

	aeRoundtripStable, rawRoundtripStable := appAddressRoundtripStable()
	return nativeaccounttypes.NativeAccountInvariantInput{
		AEAddressRoundtripStable:	aeRoundtripStable,
		RawAddressRoundtripStable:	rawRoundtripStable,
		ActivationAttempts:		map[string]uint64{},
		TotalSupply:			totalSupply,
		RewardBudget:			rewardBudget,
		ActiveValidatorCount:		activeValidators,
		MaxValidatorCount:		uint64(params.MaxValidators),
		ValidatorStakes:		validatorStakes,
		MinValidatorStake:		minValidatorStake,
		ExportImportStable:		true,
		SystemReserveRunwayBlocks:	systemReserveRunwayBlocks,
		MinSystemReserveRunwayBlocks:	minSystemReserveRunwayBlocks,
		SystemTopUpOrder:		[]string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:	protocolCriticalExecutable,
		SystemRentUnderfunded:		systemRentUnderfunded,
	}, nil
}

func (app *L1App) assertBankModuleAccountingInvariant(ctx sdk.Context) error {
	return validateBankSupplyMatchesBalances(app.BankKeeper.ExportGenesis(ctx))
}

func (app *L1App) assertBankSupplyInvariant(ctx sdk.Context) error {
	genesis := app.BankKeeper.ExportGenesis(ctx)
	if err := validateBankSupplyMatchesBalances(genesis); err != nil {
		return err
	}
	totalSupply, _, err := app.BankKeeper.GetPaginatedTotalSupply(ctx, nil)
	if err != nil {
		return err
	}
	declaredSupply := sdk.NewCoins(genesis.Supply...)
	if !totalSupply.Equal(declaredSupply) {
		return fmt.Errorf("bank keeper supply mismatch: keeper=%s exported=%s", totalSupply.String(), declaredSupply.String())
	}
	return nil
}

func (app *L1App) assertModuleAccountsInvariant(ctx sdk.Context) error {
	expected := accounts.ModuleAccountPermissions()
	for moduleName, expectedPerms := range expected {
		moduleAcc := app.AccountKeeper.GetModuleAccount(ctx, moduleName)
		if moduleAcc == nil {
			return fmt.Errorf("module account %s is missing", moduleName)
		}
		expectedAddress := app.AccountKeeper.GetModuleAddress(moduleName)
		if expectedAddress == nil {
			return fmt.Errorf("module account %s has no deterministic address", moduleName)
		}
		if !bytes.Equal(moduleAcc.GetAddress().Bytes(), expectedAddress.Bytes()) {
			return fmt.Errorf("module account %s address mismatch: account=%s expected=%s", moduleName, moduleAcc.GetAddress().String(), expectedAddress.String())
		}
		accountI := app.AccountKeeper.GetAccount(ctx, expectedAddress)
		if accountI == nil {
			return fmt.Errorf("module account %s not stored at module address", moduleName)
		}
		storedModule, ok := accountI.(sdk.ModuleAccountI)
		if !ok {
			return fmt.Errorf("module account %s stored as non-module account", moduleName)
		}
		if storedModule.GetName() != moduleName {
			return fmt.Errorf("module account %s name mismatch: stored=%s", moduleName, storedModule.GetName())
		}
		if !sameStringSet(storedModule.GetPermissions(), expectedPerms) {
			return fmt.Errorf("module account %s permissions mismatch: stored=%v expected=%v", moduleName, sortedStrings(storedModule.GetPermissions()), sortedStrings(expectedPerms))
		}
	}
	return nil
}

func (app *L1App) assertValidatorSetInvariant(ctx sdk.Context) error {
	params, err := app.StakingKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	bonded, err := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return err
	}
	lastValidators, err := app.StakingKeeper.GetLastValidators(ctx)
	if err != nil {
		return err
	}
	if uint32(len(bonded)) > params.MaxValidators {
		return fmt.Errorf("bonded validator set size %d exceeds max %d", len(bonded), params.MaxValidators)
	}
	if len(lastValidators) > len(bonded) && uint32(len(lastValidators)) > params.MaxValidators {
		return fmt.Errorf("last validator set size %d exceeds max %d", len(lastValidators), params.MaxValidators)
	}
	seen := make(map[string]struct{}, len(bonded))
	for _, validator := range bonded {
		if validator.Status != stakingtypes.Bonded {
			return fmt.Errorf("validator %s appears in bonded set with status %s", validator.OperatorAddress, validator.Status.String())
		}
		if !validator.Tokens.IsPositive() {
			return fmt.Errorf("bonded validator %s has non-positive tokens", validator.OperatorAddress)
		}
		if validator.ConsensusPower(sdk.DefaultPowerReduction) <= 0 {
			return fmt.Errorf("bonded validator %s has non-positive consensus power", validator.OperatorAddress)
		}
		if _, found := seen[validator.OperatorAddress]; found {
			return fmt.Errorf("duplicate bonded validator %s", validator.OperatorAddress)
		}
		seen[validator.OperatorAddress] = struct{}{}
	}
	for _, validator := range lastValidators {
		if strings.TrimSpace(validator.OperatorAddress) == "" {
			return errors.New("last validator set contains empty operator address")
		}
		if validator.ConsensusPower(sdk.DefaultPowerReduction) < 0 {
			return fmt.Errorf("last validator %s has negative consensus power", validator.OperatorAddress)
		}
	}
	return nil
}

func (app *L1App) assertGenesisExportInvariant(ctx sdk.Context) error {
	first, err := app.ExportAppStateAndValidators(false, nil, nil)
	if err != nil {
		return err
	}
	second, err := app.ExportAppStateAndValidators(false, nil, nil)
	if err != nil {
		return err
	}
	if !bytes.Equal(first.AppState, second.AppState) {
		return errors.New("app genesis export is nondeterministic")
	}
	var exported GenesisState
	if err := json.Unmarshal(first.AppState, &exported); err != nil {
		return err
	}
	return app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exported)
}

func validateBankSupplyMatchesBalances(genesis *banktypes.GenesisState) error {
	if genesis == nil {
		return errors.New("bank genesis state is nil")
	}
	balanceSupply := sdk.NewCoins()
	for _, balance := range genesis.Balances {
		if strings.TrimSpace(balance.Address) == "" {
			return errors.New("bank balance address is required")
		}
		if !balance.Coins.IsValid() {
			return fmt.Errorf("bank balance for %s contains invalid coins", balance.Address)
		}
		balanceSupply = balanceSupply.Add(balance.Coins...)
	}
	declaredSupply := sdk.NewCoins(genesis.Supply...)
	if !balanceSupply.Equal(declaredSupply) {
		return fmt.Errorf("bank supply mismatch: balances=%s declared=%s", balanceSupply.String(), declaredSupply.String())
	}
	return nil
}

func (app *L1App) assertDirectUserDelegationRejectedInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	return validateDirectUserDelegationPolicy(genesis.Params)
}

func validateDirectUserDelegationPolicy(params nominatorpooltypes.Params) error {
	if params.DirectUserDelegationEnabled || params.DirectUserValidatorDelegationEnabled {
		return errors.New("direct user validator delegation is enabled by staking params")
	}
	return nil
}

func (app *L1App) assertValidatorEntryInvariant(ctx sdk.Context) error {
	params, err := app.StakingKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	activeValidators := uint32(0)
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			activeValidators++
		}
		if activeValidators > params.MaxValidators {
			return fmt.Errorf("active validators %d exceed max %d", activeValidators, params.MaxValidators)
		}
		if validator.Tokens.IsPositive() && validator.Tokens.LT(validator.MinSelfDelegation) {
			return fmt.Errorf("validator %s stake %s below min self delegation %s", validator.OperatorAddress, validator.Tokens.String(), validator.MinSelfDelegation.String())
		}
	}
	return nil
}

func (app *L1App) assertNominatorPoolInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	return genesis.Validate()
}

func (app *L1App) assertPoolSharesSumInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	return validatePoolSharesSum(genesis.State)
}

func (app *L1App) assertPoolRuntimeAccountingInvariant(ctx sdk.Context) error {
	genesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := genesis.Validate(); err != nil {
		return err
	}
	return validatePoolRuntimeAccounting(genesis.State)
}

func validatePoolSharesSum(state nominatorpooltypes.State) error {
	poolShares := make(map[string]uint64)
	for _, share := range state.PoolShares {
		next, err := checkedAddUint64(poolShares[share.PoolID], share.Shares)
		if err != nil {
			return fmt.Errorf("pool %s share sum overflow", share.PoolID)
		}
		poolShares[share.PoolID] = next
	}
	for _, pool := range state.LiquidStakingPools {
		if got := poolShares[pool.PoolID]; got != pool.TotalShares {
			return fmt.Errorf("liquid staking pool %s share sum mismatch: shares=%d total=%d", pool.PoolID, got, pool.TotalShares)
		}
	}
	for _, pool := range state.Pools {
		total := uint64(0)
		for _, share := range pool.DelegatorShares {
			next, err := checkedAddUint64(total, share.Shares)
			if err != nil {
				return fmt.Errorf("nominator pool %s delegator share sum overflow", pool.PoolID)
			}
			total = next
		}
		if total != pool.TotalShares {
			return fmt.Errorf("nominator pool %s share sum mismatch: delegator=%d total=%d", pool.PoolID, total, pool.TotalShares)
		}
	}
	return nil
}

func validatePoolRuntimeAccounting(state nominatorpooltypes.State) error {
	allocationByPool := make(map[string]uint64)
	unbondingByPool := make(map[string]uint64)
	pendingRewardsByPool := make(map[string]uint64)
	for _, allocation := range state.PoolValidatorAllocations {
		total, err := checkedAddUint64(allocation.ActiveStake, allocation.PendingStake)
		if err != nil {
			return fmt.Errorf("pool %s allocation overflow", allocation.PoolID)
		}
		total, err = checkedAddUint64(total, allocation.UnbondingStake)
		if err != nil {
			return fmt.Errorf("pool %s allocation overflow", allocation.PoolID)
		}
		allocationByPool[allocation.PoolID], err = checkedAddUint64(allocationByPool[allocation.PoolID], total)
		if err != nil {
			return fmt.Errorf("pool %s allocation sum overflow", allocation.PoolID)
		}
	}
	for _, request := range state.PoolUnbondingRequests {
		if request.Status == nominatorpooltypes.WithdrawalStatusPending {
			var err error
			unbondingByPool[request.PoolID], err = checkedAddUint64(unbondingByPool[request.PoolID], request.Amount)
			if err != nil {
				return fmt.Errorf("pool %s unbonding sum overflow", request.PoolID)
			}
		}
	}
	for _, share := range state.PoolShares {
		var err error
		pendingRewardsByPool[share.PoolID], err = checkedAddUint64(pendingRewardsByPool[share.PoolID], share.PendingRewards)
		if err != nil {
			return fmt.Errorf("pool %s pending rewards overflow", share.PoolID)
		}
	}
	for _, pool := range state.LiquidStakingPools {
		activeAndUnbonding, err := checkedAddUint64(pool.TotalActiveStake, pool.TotalUnbonding)
		if err != nil {
			return fmt.Errorf("liquid staking pool %s active plus unbonding overflow", pool.PoolID)
		}
		if allocationByPool[pool.PoolID] > activeAndUnbonding {
			return fmt.Errorf("liquid staking pool %s allocations exceed active plus unbonding stake", pool.PoolID)
		}
		if unbondingByPool[pool.PoolID] != pool.TotalUnbonding {
			return fmt.Errorf("liquid staking pool %s unbonding mismatch: requests=%d total=%d", pool.PoolID, unbondingByPool[pool.PoolID], pool.TotalUnbonding)
		}
		if pendingRewardsByPool[pool.PoolID] > pool.TotalDeposited {
			return fmt.Errorf("liquid staking pool %s pending rewards exceed deposited principal", pool.PoolID)
		}
		if activeAndUnbonding > pool.TotalDeposited {
			return fmt.Errorf("liquid staking pool %s active plus unbonding stake exceeds deposits", pool.PoolID)
		}
	}
	return nil
}

func (app *L1App) assertStorageRentInvariant(ctx sdk.Context) error {
	genesis, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := genesis.Validate(); err != nil {
		return err
	}
	return validateStorageRentRuntimeState(genesis.State)
}

func validateStorageRentRuntimeState(state storagerenttypes.StorageRentState) error {
	systemRent := state.SystemReserve.Evaluate()
	if systemRent.Alert == storagerenttypes.SystemRentAlertInvariant {
		return fmt.Errorf("system storage rent reserve underfunded: runway=%d remaining_debt=%d", systemRent.RunwayBlocks, systemRent.RemainingDebt)
	}
	if !systemRent.Executable {
		return errors.New("protocol-critical storage rent path is not executable")
	}
	systemAddresses := map[string]struct{}{}
	for _, systemAddress := range addressing.AllSystemAddresses() {
		systemAddresses[systemAddress.UserFriendly] = struct{}{}
		systemAddresses[systemAddress.Raw] = struct{}{}
	}
	for _, record := range state.Contracts {
		if _, found := systemAddresses[record.ContractAddress]; !found {
			continue
		}
		switch record.Status {
		case storagerenttypes.ContractStatusFrozen, storagerenttypes.ContractStatusDeleted, storagerenttypes.ContractStatusArchived:
			return fmt.Errorf("protocol-critical storage rent record %s has forbidden status %s", record.ContractAddress, record.Status)
		}
	}
	return nil
}

func (app *L1App) assertEconomicsFeeSplitInvariant(ctx sdk.Context) error {
	genesis, err := app.AetraEconomicsKeeper.ExportGenesis()
	if err != nil {
		return err
	}
	return validateEconomicsFeeSplit(genesis.Params)
}

func (app *L1App) assertEconomicsAccountingInvariant(ctx sdk.Context) error {
	genesis, err := app.AetraEconomicsKeeper.ExportGenesis()
	if err != nil {
		return err
	}
	if err := genesis.Validate(); err != nil {
		return err
	}
	return validateEconomicsAccounting(genesis.State)
}

func validateEconomicsFeeSplit(params economicstypes.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint64(params.BurnCurrentBps)+uint64(params.ValidatorRewardBps)+uint64(params.TreasuryBps) != uint64(economicstypes.BasisPoints) {
		return fmt.Errorf("fee split percentages must sum to %d bps", economicstypes.BasisPoints)
	}
	return nil
}

func validateEconomicsAccounting(state economicstypes.EconomicsState) error {
	var lastBurned uint64
	var lastTreasury uint64
	for i, summary := range state.RewardHistory {
		allocated, err := checkedAddUint64(summary.BurnedAmount, summary.TreasuryAmount)
		if err != nil || allocated > summary.FeesCollected {
			return fmt.Errorf("epoch %d fee split exceeds collected fees", summary.Epoch)
		}
		if summary.ValidatorDelegatorRewards != summary.FeesCollected-summary.BurnedAmount-summary.TreasuryAmount {
			return fmt.Errorf("epoch %d validator rewards do not reconcile to fees", summary.Epoch)
		}
		gross, err := checkedAddUint64(summary.MintedRewards, summary.ValidatorDelegatorRewards)
		if err != nil || gross != summary.GrossRewards {
			return fmt.Errorf("epoch %d gross rewards mismatch", summary.Epoch)
		}
		expectedEnding, err := checkedAddUint64(summary.StartingSupply, summary.MintedRewards)
		if err != nil || summary.BurnedAmount > expectedEnding {
			return fmt.Errorf("epoch %d supply arithmetic overflow", summary.Epoch)
		}
		expectedEnding -= summary.BurnedAmount
		if summary.EndingSupply != expectedEnding {
			return fmt.Errorf("epoch %d ending supply mismatch", summary.Epoch)
		}
		if i > 0 {
			if summary.BurnedSupply < lastBurned || summary.TreasuryBalance < lastTreasury {
				return fmt.Errorf("epoch %d burn/treasury cumulative counters decreased", summary.Epoch)
			}
		}
		lastBurned = summary.BurnedSupply
		lastTreasury = summary.TreasuryBalance
	}
	if len(state.RewardHistory) > 0 {
		latest := state.RewardHistory[len(state.RewardHistory)-1]
		if state.BurnedSupply != latest.BurnedSupply {
			return fmt.Errorf("economics burned supply mismatch: state=%d latest=%d", state.BurnedSupply, latest.BurnedSupply)
		}
		if state.TreasuryBalance != latest.TreasuryBalance {
			return fmt.Errorf("economics treasury balance mismatch: state=%d latest=%d", state.TreasuryBalance, latest.TreasuryBalance)
		}
	}
	return nil
}

func (app *L1App) assertAVMQueueReceiptsInvariant(ctx sdk.Context) error {
	genesis, err := app.ContractsKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := genesis.Validate(); err != nil {
		return err
	}
	return validateAVMQueueReceipts(genesis.State)
}

func validateAVMQueueReceipts(state contractstypes.State) error {
	contracts := make(map[string]contractstypes.Contract)
	for _, contract := range state.Contracts {
		expectedRoot := contractstypes.ComputeContractStateRoot(contract)
		if contract.StateRoot != "" && contract.StateRoot != expectedRoot {
			return fmt.Errorf("contract %s state root mismatch", contract.AddressUser)
		}
		contracts[contract.AddressUser] = contract
	}
	seenMessages := make(map[string]struct{}, len(state.InternalMessages))
	for _, msg := range state.InternalMessages {
		if _, found := contracts[msg.SourceContractUser]; !found {
			return fmt.Errorf("internal message %s references unknown source contract %s", msg.MessageID, msg.SourceContractUser)
		}
		expectedID := contractstypes.ComputeInternalMessageID(msg)
		if msg.MessageID != expectedID {
			return fmt.Errorf("internal message id mismatch: expected %s got %s", expectedID, msg.MessageID)
		}
		if _, found := seenMessages[msg.MessageID]; found {
			return fmt.Errorf("duplicate internal message id %s", msg.MessageID)
		}
		seenMessages[msg.MessageID] = struct{}{}
		if msg.Refunded && !msg.Bounce {
			return fmt.Errorf("internal message %s is refunded without bounce semantics", msg.MessageID)
		}
	}
	return nil
}

func (app *L1App) assertValidatorInsuranceBoundsInvariant(ctx sdk.Context) error {
	registryGenesis, err := app.ValidatorRegistryKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	insuranceGenesis, err := app.ValidatorInsuranceKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := registryGenesis.Validate(); err != nil {
		return err
	}
	if err := insuranceGenesis.Validate(); err != nil {
		return err
	}
	return validateValidatorInsuranceBounds(registryGenesis.State, registryGenesis.Params, insuranceGenesis.State, insuranceGenesis.Params)
}

func validateValidatorInsuranceBounds(registryState validatorregistrytypes.State, registryParams validatorregistrytypes.Params, insuranceState validatorinsurancetypes.State, insuranceParams validatorinsurancetypes.Params) error {
	if err := registryState.Validate(registryParams); err != nil {
		return err
	}
	if err := insuranceState.Validate(insuranceParams); err != nil {
		return err
	}
	insuranceByValidator := make(map[string]validatorinsurancetypes.ValidatorInsurance, len(insuranceState.Insurances))
	for _, insurance := range insuranceState.Insurances {
		insuranceByValidator[insurance.ValidatorAddress] = insurance
	}
	for _, validator := range registryState.Validators {
		normalized := validator.Normalize(registryParams)
		if normalized.Status != validatorregistrytypes.StatusActive {
			continue
		}
		mode := validatorregistrytypes.ValidatorFundingPoolBacked
		if normalized.NominatorBond == 0 {
			mode = validatorregistrytypes.ValidatorFundingSolo
		}
		if err := registryParams.ValidateValidatorFunding(validatorregistrytypes.ValidatorFunding{
			Mode:		mode,
			SelfStake:	normalized.SelfBond,
			NominatorBond:	normalized.NominatorBond,
		}); err != nil {
			return fmt.Errorf("active validator %s funding invalid: %w", normalized.OperatorAddress, err)
		}
		if insuranceParams.Enabled {
			insurance, found := insuranceByValidator[normalized.OperatorAddress]
			if !found {
				return fmt.Errorf("active validator %s missing validator insurance", normalized.OperatorAddress)
			}
			if insurance.Balance < insuranceParams.MinimumInsurance {
				return fmt.Errorf("active validator %s insurance below minimum: balance=%d minimum=%d", normalized.OperatorAddress, insurance.Balance, insuranceParams.MinimumInsurance)
			}
		}
	}
	return nil
}

func (app *L1App) assertEmissionCapInvariant(ctx sdk.Context) error {
	totalMinted, err := app.EmissionsKeeper.GetTotalMintedAccounting(ctx)
	if err != nil {
		return err
	}
	emParams, err := app.EmissionsKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	maxAnnualSupply := emParams.AnnualReferenceSupply
	if maxAnnualSupply.IsZero() {
		return nil
	}
	maxMintable := maxAnnualSupply.Amount.Mul(sdkmath.NewInt(int64(emParams.ConstitutionalMaxInflationBps))).Quo(sdkmath.NewInt(10_000))
	if totalMinted.Amount.GT(maxMintable) {
		return fmt.Errorf("emissions total minted accounting %s exceeds constitutional max %s", totalMinted.String(), sdk.NewCoin(maxAnnualSupply.Denom, maxMintable).String())
	}
	return nil
}

func (app *L1App) assertBurnAccountingInvariant(ctx sdk.Context) error {
	burnedEntries, err := app.BurnKeeper.GetAllBurnedByDenom(ctx)
	if err != nil {
		return err
	}
	burnedEpochs, err := app.BurnKeeper.GetAllBurnedByEpoch(ctx)
	if err != nil {
		return err
	}
	baseDenom := params.BaseDenom
	denomSum := sdkmath.ZeroInt()
	for _, entry := range burnedEntries {
		if entry.Denom == baseDenom {
			denomSum = denomSum.Add(entry.Amount[0].Amount)
		}
	}
	epochSum := sdkmath.ZeroInt()
	for _, entry := range burnedEpochs {
		for _, coin := range entry.Amount {
			if coin.Denom == baseDenom {
				epochSum = epochSum.Add(coin.Amount)
			}
		}
	}
	if !denomSum.Equal(epochSum) {
		return fmt.Errorf("burn accounting mismatch: denom sum=%s epoch sum=%s", denomSum.String(), epochSum.String())
	}
	return nil
}

func (app *L1App) assertRentReserveBalanceInvariant(ctx sdk.Context) error {
	gs, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	result := gs.State.SystemReserve.Evaluate()
	if result.Alert == storagerenttypes.SystemRentAlertInvariant {
		return fmt.Errorf("storage rent system reserve is in invariant alert state")
	}
	return nil
}

func (app *L1App) assertNoNativeAppAssetModulesInvariant(ctx sdk.Context) error {
	entries := DefaultLaunchModuleInventory()
	byModule := make(map[string]LaunchModuleInventoryEntry, len(entries))
	for _, entry := range entries {
		if entry.ModuleName != "" {
			byModule[entry.ModuleName] = entry
		}
	}
	for moduleName := range app.ModuleManager.Modules {
		entry, found := byModule[moduleName]
		if !found {
			continue
		}
		if !entry.AppWired {
			continue
		}
		switch entry.Classification {
		case LaunchModulePrototypeOnly, LaunchModuleDisabled:
			return fmt.Errorf("module %s is %s and must not be wired in app", moduleName, entry.Classification)
		case LaunchModuleFutureAVMStandard:
			return fmt.Errorf("module %s is future AVM standard and must not be wired as native module", moduleName)
		}
	}
	return nil
}

func validateNoNativeAppAssetModules(entries []LaunchModuleInventoryEntry) error {
	for _, entry := range entries {
		if !entry.AppWired {
			continue
		}
		switch entry.Classification {
		case LaunchModulePrototypeOnly, LaunchModuleDisabled:
			return fmt.Errorf("module %s is %s and must not be wired in app", entry.ModuleName, entry.Classification)
		case LaunchModuleFutureAVMStandard:
			return fmt.Errorf("module %s is future AVM standard and must not be wired as native module", entry.ModuleName)
		}
	}
	return nil
}

func (app *L1App) assertTreasuryAccountingInvariant(ctx sdk.Context) error {
	return app.TreasuryKeeper.AssertTreasuryAccountingInvariant(ctx)
}

func (app *L1App) assertFeeCollectorAccountingInvariant(ctx sdk.Context) error {
	return app.FeeCollectorKeeper.AssertModuleAccountingInvariant(ctx)
}

func (app *L1App) assertReservedSystemAddressInvariant(ctx sdk.Context) error {
	return validateReservedSystemAddressPolicy(app.BankKeeper.GetBlockedAddresses())
}

func validateReservedSystemAddressPolicy(blocked map[string]bool) error {
	seenRaw := map[string]string{}
	seenUser := map[string]string{}
	for _, address := range addressing.AllSystemAddresses() {
		if strings.TrimSpace(address.Raw) == "" || strings.TrimSpace(address.UserFriendly) == "" {
			return fmt.Errorf("reserved system address %s has empty mapping", address.Name)
		}
		if err := addressing.ValidateUserSignerAddress(address.UserFriendly); err == nil {
			return fmt.Errorf("reserved system address %s can be used as user signer", address.Name)
		}
		rawKey, err := addressing.AddressTextBytesKey(address.Raw)
		if err != nil {
			return fmt.Errorf("reserved system address %s raw mapping invalid: %w", address.Name, err)
		}
		userKey, err := addressing.AddressTextBytesKey(address.UserFriendly)
		if err != nil {
			return fmt.Errorf("reserved system address %s user mapping invalid: %w", address.Name, err)
		}
		if rawKey != userKey {
			return fmt.Errorf("reserved system address %s raw/user-friendly mapping mismatch", address.Name)
		}
		if other, found := seenRaw[rawKey]; found {
			return fmt.Errorf("reserved system address %s duplicates raw mapping with %s", address.Name, other)
		}
		seenRaw[rawKey] = address.Name
		if other, found := seenUser[address.UserFriendly]; found {
			return fmt.Errorf("reserved system address %s duplicates user-friendly mapping with %s", address.Name, other)
		}
		seenUser[address.UserFriendly] = address.Name
		acc, err := addressing.ParseAccAddress(address.UserFriendly)
		if err == nil && !address.CanReceiveUserFunds && !blocked[acc.String()] {
			return fmt.Errorf("reserved system address %s is not bank-blocked for user receives", address.Name)
		}
	}
	return nil
}

func appAddressRoundtripStable() (bool, bool) {
	addr := sdk.AccAddress(bytes.Repeat([]byte{0x42}, 20))
	user := addressing.FormatAccAddress(addr)
	pairFromUser, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, user)
	if err != nil || pairFromUser.User != user {
		return false, false
	}
	parsedUser, err := addressing.ParseAccAddress(pairFromUser.User)
	if err != nil || !bytes.Equal(parsedUser.Bytes(), addr.Bytes()) {
		return false, false
	}
	pairFromRaw, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, pairFromUser.Raw)
	if err != nil {
		return true, false
	}
	parsedRaw, err := addressing.ParseAccAddress(pairFromRaw.Raw)
	if err != nil {
		return true, false
	}
	return pairFromRaw.User == user && pairFromRaw.Raw == pairFromUser.Raw && bytes.Equal(parsedRaw.Bytes(), addr.Bytes()), true
}

func safeUint64(value sdkmath.Int, fallback uint64) uint64 {
	if value.IsNil() || !value.IsUint64() {
		return fallback
	}
	return value.Uint64()
}

func checkedAddUint64(a, b uint64) (uint64, error) {
	if math.MaxUint64-a < b {
		return 0, errors.New("uint64 addition overflow")
	}
	return a + b, nil
}

func sameStringSet(left, right []string) bool {
	left = sortedStrings(left)
	right = sortedStrings(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
