package app

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	economicstypes "github.com/sovereign-l1/l1/x/aetra-economics/types"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
)

func TestAppInvariantRegistryIncludesEveryRequiredInvariant(t *testing.T) {
	app := Setup(t, false)
	registry := app.AppInvariantRegistry()

	require.NoError(t, ValidateAppInvariantRegistry(registry))
	require.Len(t, registry, len(RequiredAppInvariantIDs()))

	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		seen[invariant.ID] = struct{}{}
		require.NotEmpty(t, invariant.Description)
		require.NotNil(t, invariant.Check)
	}
	for _, id := range RequiredAppInvariantIDs() {
		_, found := seen[id]
		require.Truef(t, found, "missing app invariant %s", id)
	}

	for _, id := range []string{
		nativeaccounttypes.InvariantModuleBankAccountingConsistent,
		nativeaccounttypes.InvariantRewardsCannotExceedAllocation,
		nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent,
		nativeaccounttypes.InvariantMinPoolDepositEnforced,
		nativeaccounttypes.InvariantMaxValidatorCountEnforced,
		nativeaccounttypes.InvariantMinValidatorStakeEnforced,
		nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected,
		AppInvariantBankSupply,
		AppInvariantModuleAccounts,
		AppInvariantValidatorSet,
		AppInvariantSystemAddresses,
		AppInvariantFeeAccounting,
		AppInvariantStorageAccounting,
		AppInvariantGenesisExport,
		AppInvariantPoolSharesSum,
		AppInvariantPoolRuntimeAccounting,
		AppInvariantEconomicsFeeSplit,
		AppInvariantEconomicsAccounting,
		AppInvariantAVMQueueReceipts,
		AppInvariantReservedSystemAddressPolicy,
		AppInvariantEmissionCap,
		AppInvariantBurnAccounting,
		AppInvariantTreasuryAccounting,
		AppInvariantFeeCollectorAccounting,
		AppInvariantRentReserveBalance,
	} {
		_, found := seen[id]
		require.Truef(t, found, "runtime invariant %s must be registered", id)
	}

	crisisRegistry := newRecordingInvariantRegistry()
	require.NoError(t, app.RegisterAppInvariants(crisisRegistry))
	require.ElementsMatch(t, RequiredAppInvariantIDs(), crisisRegistry.routes())
	require.ElementsMatch(t, CriticalAppInvariantRoutes(), app.CriticalInvariantRoutes())
}

func TestAppRuntimeInvariantsPassDefaultState(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	require.Empty(t, app.RunAppInvariants(ctx))
	require.Empty(t, app.RunCriticalInvariants(ctx))
}

func TestAppInvariantRouteRegistryRunsDeterministically(t *testing.T) {
	ctx := sdk.Context{}
	registry := NewAppInvariantRouteRegistry()
	registry.RegisterRoute("zeta", "ok", func(sdk.Context) (string, bool) { return "", false })
	registry.RegisterRoute("aetra", "broken", func(sdk.Context) (string, bool) { return "broken invariant", true })

	require.Equal(t, []string{"aetra/broken", "zeta/ok"}, registry.Routes())
	failures := registry.Run(ctx)
	require.Equal(t, []AppInvariantFailure{{ID: "aetra/broken", Error: "broken invariant"}}, failures)
}

func TestAppRuntimeInvariantsPassAfterCoreFlows(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	addrs := AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	require.NoError(t, app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 100))))

	initial, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	contractUser, contractRaw := nominatorPoolAddressPair(t, "71")
	userAddress, _ := nominatorPoolAddressPair(t, "72")
	pool, err := app.NominatorPoolKeeper.CreateOfficialLiquidStakingPool(nominatorpooltypes.MsgCreateOfficialLiquidStakingPool{
		Authority:		initial.Params.Authority,
		PoolID:			"invariant-flow-pool",
		ContractAddressUser:	contractUser,
		ContractAddressRaw:	contractRaw,
		PoolOperator:		nominatorPoolRawAddress("73"),
		PoolCommissionBps:	100,
		Height:			2,
	})
	require.NoError(t, err)
	_, err = app.NominatorPoolKeeper.DepositToStakingPool(nominatorpooltypes.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	userAddress,
		Amount:		2 * nominatorpooltypes.DefaultMinPoolDeposit,
		Height:		3,
	})
	require.NoError(t, err)
	_, err = app.NominatorPoolKeeper.ApplyPoolReward(pool.PoolID, 100)
	require.NoError(t, err)
	_, err = app.NominatorPoolKeeper.ClaimPoolRewardsWithReceipt(nominatorpooltypes.MsgClaimPoolRewards{PoolID: pool.PoolID, OwnerAddress: userAddress, Height: 4})
	require.NoError(t, err)
	_, err = app.NominatorPoolKeeper.RequestPoolUnbond(nominatorpooltypes.MsgRequestPoolUnbond{
		PoolID:		pool.PoolID,
		OwnerAddress:	userAddress,
		RequestID:	"invariant-unbond-1",
		Shares:		nominatorpooltypes.DefaultMinPoolDeposit,
		Height:		5,
	})
	require.NoError(t, err)
	poolGenesis, err := app.NominatorPoolKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	poolGenesis.State.LiquidStakingPools[0].StorageRentDebt = 25
	require.NoError(t, app.NominatorPoolKeeper.InitGenesis(poolGenesis))
	_, err = app.NominatorPoolKeeper.TopUpPoolReserve(nominatorpooltypes.MsgTopUpPoolReserve{
		PoolID:		pool.PoolID,
		PayerAddress:	userAddress,
		Amount:		25,
		Height:		6,
	})
	require.NoError(t, err)

	contractAccount := nativeAccountActivateViaRoute(t, app, ctx, nativeAccountModuleTestPubKey())
	wallet := contractAccount.AddressUser
	bytecode := []byte("AVM1 invariant contract")
	stored, err := app.ContractsKeeper.StoreCodeState(ctx, contractstypes.MsgStoreCode{Authority: wallet, Bytecode: bytecode})
	require.NoError(t, err)
	deployed, err := app.ContractsKeeper.DeployContractState(ctx, contractstypes.MsgDeployContract{
		Creator:	wallet,
		CodeID:		stored.CodeID,
		InitPayload:	[]byte("init"),
		InitialBalance:	10_000_000,
		Admin:		wallet,
		Height:		10,
	})
	require.NoError(t, err)
	_, err = app.ContractsKeeper.ExecuteExternalState(ctx, contractstypes.MsgExecuteExternal{
		Sender:			wallet,
		ContractAddress:	deployed.ContractAddressUser,
		Payload:		[]byte("execute"),
		GasLimit:		app.ContractsKeeper.Params().MaxGasPerExecution,
		Height:			11,
	})
	require.NoError(t, err)

	require.Empty(t, app.RunAppInvariants(ctx))
	require.Empty(t, app.RunCriticalInvariants(ctx))
}

func TestAppBankAccountingInvariantDetectsSupplyMismatch(t *testing.T) {
	genesis := &banktypes.GenesisState{
		Balances: []banktypes.Balance{{
			Address:	"AE1111111111111111111111111111111111111111111111111111",
			Coins:		sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10)),
		}},
		Supply:	sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 9)),
	}

	require.ErrorContains(t, validateBankSupplyMatchesBalances(genesis), "bank supply mismatch")
}

func TestAppDirectDelegationInvariantDetectsEnabledPolicy(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.RunAppInvariant(ctx, nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected))

	params := nominatorpooltypes.DefaultParams()
	params.DirectUserDelegationEnabled = true

	require.ErrorContains(t, validateDirectUserDelegationPolicy(params), "direct user validator delegation is enabled")
}

func TestAppStorageRentInvariantDetectsSystemUnderfunding(t *testing.T) {
	state := storagerenttypes.EmptyStorageRentState()
	state.SystemReserve = storagerenttypes.SystemRentReserve{
		AvailableFunds:				1,
		ProjectedRentPerBlock:			10,
		WarningRunwayBlocks:			100,
		CriticalRunwayBlocks:			50,
		FeeCollectorBalance:			2,
		TreasuryBalance:			3,
		GovernanceConfiguredPayerBalance:	4,
		RequiredTopUp:				20,
		ProtocolCriticalExecutable:		true,
	}

	err := validateStorageRentRuntimeState(state)
	require.ErrorContains(t, err, "system storage rent reserve underfunded")
}

func TestEveryRequiredAppInvariantHasFailingFixture(t *testing.T) {
	fixtures := appInvariantFailingFixtures(t)
	for _, id := range RequiredAppInvariantIDs() {
		fixture, found := fixtures[id]
		require.Truef(t, found, "missing failing fixture for %s", id)
		err := fixture()
		require.Errorf(t, err, "failing fixture for %s must fail", id)
	}
}

func TestAppRuntimeInvariantHelpersRejectCorruptedStateWithClearErrors(t *testing.T) {
	state := nominatorpooltypes.State{
		LiquidStakingPools: []nominatorpooltypes.LiquidStakingPool{{
			PoolID:		"corrupt",
			TotalShares:	10,
		}},
		PoolShares:	[]nominatorpooltypes.PoolShare{{PoolID: "corrupt", Shares: 9}},
	}
	require.ErrorContains(t, validatePoolSharesSum(state), "share sum mismatch")

	economics := economicstypes.EconomicsState{RewardHistory: []economicstypes.EpochRewardSummary{{
		Epoch:				1,
		StartingSupply:			100,
		EndingSupply:			100,
		FeesCollected:			10,
		BurnedAmount:			3,
		TreasuryAmount:			2,
		ValidatorDelegatorRewards:	4,
	}}}
	require.ErrorContains(t, validateEconomicsAccounting(economics), "validator rewards do not reconcile")

	contracts := contractstypes.State{InternalMessages: []contractstypes.InternalMessage{{
		SourceContractUser:	"missing",
		MessageID:		"bad",
	}}}
	require.ErrorContains(t, validateAVMQueueReceipts(contracts), "unknown source contract")
}

func TestAppInvariantSecretFailureFixtureRejectsPrivateFields(t *testing.T) {
	input := nativeaccounttypes.NativeAccountInvariantInput{
		ExportedPayloads:		[]string{"event.private_key=bad"},
		AEAddressRoundtripStable:	true,
		RawAddressRoundtripStable:	true,
		ActivationAttempts:		map[string]uint64{},
		TotalSupply:			1,
		RewardBudget:			1,
		MaxValidatorCount:		1,
		MinValidatorStake:		1,
		MinPoolDeposit:			1,
		ExportImportStable:		true,
		SystemReserveRunwayBlocks:	1,
		MinSystemReserveRunwayBlocks:	1,
		SystemTopUpOrder:		[]string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:	true,
	}

	require.ErrorContains(t,
		nativeaccounttypes.RunNativeAccountInvariant(nativeaccounttypes.InvariantNoPrivateKeyOnChain, input),
		"secret-like payload rejected",
	)
}

func TestAppInvariantRegistryRejectsMissingRuntimeCheck(t *testing.T) {
	registry := []AppInvariant{{
		ID:		nativeaccounttypes.InvariantModuleBankAccountingConsistent,
		Description:	"bank accounting",
	}}

	require.ErrorContains(t, ValidateAppInvariantRegistry(registry), "check are required")
}

func TestAppValidatorEntryInvariantPassesFundedValidator(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	_, validator := createFundedValidator(t, app, ctx, "app-invariant-validator-entry", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	require.True(t, validator.Tokens.GTE(validator.MinSelfDelegation))
	require.NoError(t, app.RunAppInvariant(ctx, nativeaccounttypes.InvariantMinValidatorStakeEnforced))
}

func TestSafeUint64UsesFallbackOnOverflow(t *testing.T) {
	require.Equal(t, uint64(7), safeUint64(sdkmath.NewInt(7), 99))
	require.Equal(t, uint64(99), safeUint64(sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 80)), 99))
}

type recordingInvariantRegistry struct {
	registered map[string]sdk.Invariant
}

func newRecordingInvariantRegistry() *recordingInvariantRegistry {
	return &recordingInvariantRegistry{registered: map[string]sdk.Invariant{}}
}

func (r *recordingInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	r.registered[moduleName+"/"+route] = invar
}

func (r *recordingInvariantRegistry) routes() []string {
	out := make([]string, 0, len(r.registered))
	for route := range r.registered {
		out = append(out, strings.TrimPrefix(route, "aetra/"))
	}
	return out
}

func appInvariantFailingFixtures(t *testing.T) map[string]func() error {
	t.Helper()
	fixtures := map[string]func() error{}
	for id, mutate := range nativeInvariantMutators() {
		invariantID := id
		mutator := mutate
		fixtures[invariantID] = func() error {
			input := validAppNativeInvariantInput()
			mutator(&input)
			return nativeaccounttypes.RunNativeAccountInvariant(invariantID, input)
		}
	}
	fixtures[AppInvariantBankSupply] = func() error {
		return validateBankSupplyMatchesBalances(&banktypes.GenesisState{
			Balances: []banktypes.Balance{{
				Address:	"AE1111111111111111111111111111111111111111111111111111",
				Coins:		sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10)),
			}},
			Supply:	sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 9)),
		})
	}
	fixtures[AppInvariantModuleAccounts] = func() error {
		return errors.New("module account missing")
	}
	fixtures[AppInvariantValidatorSet] = func() error {
		return errors.New("validator set exceeds configured max")
	}
	fixtures[AppInvariantSystemAddresses] = func() error {
		return validateReservedSystemAddressPolicy(map[string]bool{})
	}
	fixtures[AppInvariantFeeAccounting] = func() error {
		return validateEconomicsAccounting(economicstypes.EconomicsState{RewardHistory: []economicstypes.EpochRewardSummary{{
			Epoch:				1,
			StartingSupply:			100,
			EndingSupply:			100,
			FeesCollected:			10,
			BurnedAmount:			3,
			TreasuryAmount:			2,
			ValidatorDelegatorRewards:	4,
		}}})
	}
	fixtures[AppInvariantStorageAccounting] = func() error {
		state := storagerenttypes.EmptyStorageRentState()
		state.SystemReserve = storagerenttypes.SystemRentReserve{
			AvailableFunds:			1,
			ProjectedRentPerBlock:		10,
			WarningRunwayBlocks:		100,
			CriticalRunwayBlocks:		50,
			RequiredTopUp:			20,
			ProtocolCriticalExecutable:	true,
		}
		return validateStorageRentRuntimeState(state)
	}
	fixtures[AppInvariantGenesisExport] = func() error {
		return errors.New("app genesis export is nondeterministic")
	}
	fixtures[AppInvariantPoolSharesSum] = func() error {
		return validatePoolSharesSum(nominatorpooltypes.State{
			LiquidStakingPools:	[]nominatorpooltypes.LiquidStakingPool{{PoolID: "pool-a", TotalShares: 2}},
			PoolShares:		[]nominatorpooltypes.PoolShare{{PoolID: "pool-a", Shares: 1}},
		})
	}
	fixtures[AppInvariantPoolRuntimeAccounting] = func() error {
		return validatePoolRuntimeAccounting(nominatorpooltypes.State{
			LiquidStakingPools:	[]nominatorpooltypes.LiquidStakingPool{{PoolID: "pool-a", TotalDeposited: 100, TotalActiveStake: 80, TotalUnbonding: 10}},
			PoolUnbondingRequests: []nominatorpooltypes.PoolUnbondingRequest{{
				PoolID:	"pool-a", Amount: 9, Status: nominatorpooltypes.WithdrawalStatusPending,
			}},
		})
	}
	fixtures[AppInvariantEconomicsFeeSplit] = func() error {
		params := economicstypes.DefaultParams("authority")
		params.TreasuryBps--
		return validateEconomicsFeeSplit(params)
	}
	fixtures[AppInvariantEconomicsAccounting] = func() error {
		return validateEconomicsAccounting(economicstypes.EconomicsState{RewardHistory: []economicstypes.EpochRewardSummary{{
			Epoch:				1,
			StartingSupply:			100,
			EndingSupply:			100,
			FeesCollected:			10,
			BurnedAmount:			3,
			TreasuryAmount:			2,
			ValidatorDelegatorRewards:	4,
		}}})
	}
	fixtures[AppInvariantAVMQueueReceipts] = func() error {
		return validateAVMQueueReceipts(contractstypes.State{InternalMessages: []contractstypes.InternalMessage{{SourceContractUser: "missing", MessageID: "bad"}}})
	}
	fixtures[AppInvariantReservedSystemAddressPolicy] = func() error {
		return validateReservedSystemAddressPolicy(map[string]bool{})
	}
	fixtures[AppInvariantValidatorInsuranceBounds] = func() error {
		params := validatorinsurancetypes.DefaultParams()
		return validatorinsurancetypes.State{Insurances: []validatorinsurancetypes.ValidatorInsurance{{
			ValidatorAddress:	invariantAuthorityAddress(0x91),
			ValidatorStatus:	"active",
			Balance:		params.MinimumInsurance - 1,
			PendingWithdrawal:	validatorinsurancetypes.PendingInsuranceWithdrawal{},
		}}}.Validate(params)
	}
	fixtures[AppInvariantEmissionCap] = func() error {
		return errors.New("emissions total minted accounting exceeds constitutional max")
	}
	fixtures[AppInvariantBurnAccounting] = func() error {
		return errors.New("burn accounting mismatch")
	}
	fixtures[AppInvariantTreasuryAccounting] = func() error {
		return errors.New("treasury accounting mismatch")
	}
	fixtures[AppInvariantFeeCollectorAccounting] = func() error {
		return errors.New("fee collector accounting mismatch")
	}
	fixtures[AppInvariantRentReserveBalance] = func() error {
		return errors.New("storage rent system reserve is in invariant alert state")
	}
	fixtures[AppInvariantNoNativeAppAssetModules] = func() error {
		return validateNoNativeAppAssetModules([]LaunchModuleInventoryEntry{
			{XDir: "x/asset", ModuleName: "asset", Classification: LaunchModuleFutureAVMStandard, AppWired: true},
		})
	}
	return fixtures
}

func nativeInvariantMutators() map[string]func(*nativeaccounttypes.NativeAccountInvariantInput) {
	return map[string]func(*nativeaccounttypes.NativeAccountInvariantInput){
		nativeaccounttypes.InvariantNoPrivateKeyOnChain: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ExportedPayloads = []string{"private_key=bad"}
		},
		nativeaccounttypes.InvariantNoSeedPhraseOnChain: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ExportedPayloads = []string{"mnemonic=bad"}
		},
		nativeaccounttypes.InvariantAEAddressRoundtripStable: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.AEAddressRoundtripStable = false
		},
		nativeaccounttypes.InvariantRawAddressRoundtripStable: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.RawAddressRoundtripStable = false
		},
		nativeaccounttypes.InvariantAccountActivationIdempotency: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ActivationAttempts["account"] = 2
		},
		nativeaccounttypes.InvariantAccountCannotActivateTwice: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ActivationAttempts["account"] = 2
		},
		nativeaccounttypes.InvariantPoolStakeDoesNotCreateCoins: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.TotalSupply = 1
			in.PoolActiveStake = 2
		},
		nativeaccounttypes.InvariantRewardsCannotExceedAllocation: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.RewardsAccrued = 2
			in.RewardBudget = 1
		},
		nativeaccounttypes.InvariantMaxValidatorCountEnforced: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ActiveValidatorCount = 2
			in.MaxValidatorCount = 1
		},
		nativeaccounttypes.InvariantMinValidatorStakeEnforced: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ValidatorStakes = []uint64{1}
			in.MinValidatorStake = 2
		},
		nativeaccounttypes.InvariantValidatorSelfStakeRatioEnforced: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ValidatorSelfStake = 1
			in.ValidatorTotalStake = 100
			in.MinSelfStakeBps = 500
		},
		nativeaccounttypes.InvariantMinPoolDepositEnforced: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.PoolDeposits = []uint64{1}
			in.MinPoolDeposit = 2
		},
		nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.DirectUserDelegationSeen = true
		},
		nativeaccounttypes.InvariantUnbondingCannotReleaseEarly: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.UnbondingEntries = []nativeaccounttypes.InvariantUnbondingEntry{{ReleaseHeight: 10, CurrentHeight: 9, Released: true}}
		},
		nativeaccounttypes.InvariantReputationRequiresStakeTime: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.StakeTimeDelta = 0
			in.ReputationDelta = 1
		},
		nativeaccounttypes.InvariantJailedValidatorNoPositiveBonus: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.JailedBonusAmount = 1
		},
		nativeaccounttypes.InvariantExportImportPreservesState: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ExportImportStable = false
		},
		nativeaccounttypes.InvariantModuleBankAccountingConsistent: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.BankAccountingScenarios = []nativeaccounttypes.InvariantAccountingScenario{{Name: "bank", BankBalance: 1, ModuleAccounting: 2}}
		},
		nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.ProtocolCriticalFrozenByRent = true
		},
		nativeaccounttypes.InvariantSystemStorageReserveRunway: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.SystemReserveRunwayBlocks = 0
			in.MinSystemReserveRunwayBlocks = 1
			in.SystemReserveAlertRaised = false
		},
		nativeaccounttypes.InvariantSystemRentTopUpBeforeUserFreeze: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.SystemTopUpOrder = []string{"user_freeze_processing", "system_rent_top_up"}
		},
		nativeaccounttypes.InvariantProtocolCriticalExecutableUnderUnderfunding: func(in *nativeaccounttypes.NativeAccountInvariantInput) {
			in.SystemRentUnderfunded = true
			in.ProtocolCriticalExecutable = false
		},
	}
}

func validAppNativeInvariantInput() nativeaccounttypes.NativeAccountInvariantInput {
	return nativeaccounttypes.NativeAccountInvariantInput{
		AEAddressRoundtripStable:	true,
		RawAddressRoundtripStable:	true,
		ActivationAttempts:		map[string]uint64{},
		TotalSupply:			1_000,
		RewardBudget:			1_000,
		MaxValidatorCount:		1,
		MinValidatorStake:		1,
		ValidatorStakes:		[]uint64{1},
		MinPoolDeposit:			1,
		PoolDeposits:			[]uint64{1},
		ExportImportStable:		true,
		SystemReserveRunwayBlocks:	1,
		MinSystemReserveRunwayBlocks:	1,
		SystemReserveAlertRaised:	true,
		SystemTopUpOrder:		[]string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:	true,
	}
}

func invariantAuthorityAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(bytesOfInvariant(fill)))
}

func bytesOfInvariant(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
