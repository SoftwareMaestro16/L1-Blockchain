package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestNativeAccountInvariantRegistryIncludesEveryRequiredInvariant(t *testing.T) {
	registry := DefaultNativeAccountInvariantRegistry()

	require.NoError(t, ValidateNativeAccountInvariantRegistry(registry))
	require.Len(t, registry, len(RequiredNativeAccountInvariantIDs()))

	registered := map[string]bool{}
	for _, invariant := range registry {
		registered[invariant.ID] = true
		require.NotEmpty(t, invariant.Description)
	}
	for _, id := range RequiredNativeAccountInvariantIDs() {
		require.True(t, registered[id], id)
	}
}

func TestNativeAccountInvariantFailureFixturesProduceExpectedErrors(t *testing.T) {
	fixtures := map[string]NativeAccountInvariantInput{
		InvariantNoPrivateKeyOnChain:		mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ExportedPayloads = []string{"private_key:bad"} }),
		InvariantNoSeedPhraseOnChain:		mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ExportedPayloads = []string{"seed phrase:bad"} }),
		InvariantAEAddressRoundtripStable:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.AEAddressRoundtripStable = false }),
		InvariantRawAddressRoundtripStable:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.RawAddressRoundtripStable = false }),
		InvariantAccountActivationIdempotency:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ActivationAttempts["AE-duplicate"] = 2 }),
		InvariantAccountCannotActivateTwice:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ActivationAttempts["AE-duplicate"] = 2 }),
		InvariantPoolStakeDoesNotCreateCoins:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.TotalSupply = 1 }),
		InvariantRewardsCannotExceedAllocation:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.RewardsAccrued = in.RewardBudget + 1 }),
		InvariantMaxValidatorCountEnforced:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ActiveValidatorCount = in.MaxValidatorCount + 1 }),
		InvariantMinValidatorStakeEnforced:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ValidatorStakes = []uint64{in.MinValidatorStake - 1} }),
		InvariantValidatorSelfStakeRatioEnforced: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.ValidatorSelfStake = 1
			in.ValidatorTotalStake = 1_000_000
			in.MinSelfStakeBps = 4_000
		}),
		InvariantMinPoolDepositEnforced:		mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.PoolDeposits = []uint64{in.MinPoolDeposit - 1} }),
		InvariantDirectUserValidatorDelegationRejected:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.DirectUserDelegationSeen = true }),
		InvariantUnbondingCannotReleaseEarly: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.UnbondingEntries = []InvariantUnbondingEntry{{ReleaseHeight: 100, CurrentHeight: 99, Released: true}}
		}),
		InvariantReputationRequiresStakeTime:		mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.StakeTimeDelta = 0; in.ReputationDelta = 1 }),
		InvariantJailedValidatorNoPositiveBonus:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.JailedBonusAmount = 1 }),
		InvariantExportImportPreservesState:		mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ExportImportStable = false }),
		InvariantModuleBankAccountingConsistent: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.BankAccountingScenarios[0].ModuleAccounting = in.BankAccountingScenarios[0].BankBalance + 1
		}),
		InvariantProtocolCriticalStateNotFrozenByRent:	mutateInvariantInput(func(in *NativeAccountInvariantInput) { in.ProtocolCriticalFrozenByRent = true }),
		InvariantSystemStorageReserveRunway: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.SystemReserveRunwayBlocks = 1
			in.MinSystemReserveRunwayBlocks = 10
			in.SystemReserveAlertRaised = false
		}),
		InvariantSystemRentTopUpBeforeUserFreeze: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.SystemTopUpOrder = []string{"user_freeze_processing", "system_rent_top_up"}
		}),
		InvariantProtocolCriticalExecutableUnderUnderfunding: mutateInvariantInput(func(in *NativeAccountInvariantInput) {
			in.SystemRentUnderfunded = true
			in.ProtocolCriticalExecutable = false
		}),
	}

	for id, input := range fixtures {
		t.Run(id, func(t *testing.T) {
			err := RunNativeAccountInvariant(id, input)
			require.Error(t, err)
		})
	}
}

func TestNativeAccountBankModuleAccountingInvariantPassesAfterCoreFlows(t *testing.T) {
	input := validInvariantInput(t)
	input.BankAccountingScenarios = []InvariantAccountingScenario{
		{Name: "pool deposit", BankBalance: 1_000, ModuleAccounting: 1_000},
		{Name: "reward claim", BankBalance: 950, ModuleAccounting: 950},
		{Name: "pool unbonding", BankBalance: 900, ModuleAccounting: 900},
		{Name: "storage rent payment", BankBalance: 875, ModuleAccounting: 875},
		{Name: "allocation rebalance", BankBalance: 875, ModuleAccounting: 875},
		{Name: "contract execution", BankBalance: 860, ModuleAccounting: 860},
	}

	require.NoError(t, RunNativeAccountInvariant(InvariantModuleBankAccountingConsistent, input))
	require.Empty(t, RunAllNativeAccountInvariants(input))
}

func TestSecretInjectionIntoAccountAuthGenesisAndEventsRejected(t *testing.T) {
	for _, fixture := range []struct {
		name	string
		mutate	func(*NativeAccountInvariantInput)
		wantErr	string
	}{
		{
			name:	"account pubkey private key",
			mutate: func(input *NativeAccountInvariantInput) {
				input.Accounts[0].PubKeys = []string{"private_key:bad"}
			},
			wantErr:	"private keys",
		},
		{
			name:	"account metadata seed",
			mutate: func(input *NativeAccountInvariantInput) {
				input.Accounts[0].Metadata.MetadataHash = "seed_phrase:bad"
			},
			wantErr:	"seed phrases",
		},
		{
			name:	"auth policy private key",
			mutate: func(input *NativeAccountInvariantInput) {
				input.Accounts[0].AuthPolicy.Mode = "private_key_mode"
			},
			wantErr:	"private key",
		},
		{
			name:	"genesis event payload mnemonic",
			mutate: func(input *NativeAccountInvariantInput) {
				input.ExportedPayloads = append(input.ExportedPayloads, `{"event":"mnemonic leaked"}`)
			},
			wantErr:	"mnemonic",
		},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			input := validInvariantInput(t)
			fixture.mutate(&input)
			err := RunNativeAccountInvariant(InvariantNoPrivateKeyOnChain, input)
			if err == nil {
				err = RunNativeAccountInvariant(InvariantNoSeedPhraseOnChain, input)
			}
			require.ErrorContains(t, err, fixture.wantErr)
		})
	}
}

func TestProtocolCriticalRentInvariantsPassWhenReserveAlertsAndTopUpOrderAreCorrect(t *testing.T) {
	input := validInvariantInput(t)
	input.SystemReserveRunwayBlocks = 5
	input.MinSystemReserveRunwayBlocks = 10
	input.SystemReserveAlertRaised = true
	input.SystemRentUnderfunded = true
	input.ProtocolCriticalExecutable = true

	require.NoError(t, RunNativeAccountInvariant(InvariantProtocolCriticalStateNotFrozenByRent, input))
	require.NoError(t, RunNativeAccountInvariant(InvariantSystemStorageReserveRunway, input))
	require.NoError(t, RunNativeAccountInvariant(InvariantSystemRentTopUpBeforeUserFreeze, input))
	require.NoError(t, RunNativeAccountInvariant(InvariantProtocolCriticalExecutableUnderUnderfunding, input))
}

func mutateInvariantInput(mutator func(*NativeAccountInvariantInput)) NativeAccountInvariantInput {
	input := validInvariantInput(nil)
	mutator(&input)
	return input
}

func validInvariantInput(t *testing.T) NativeAccountInvariantInput {
	userAddress, rawAddress := invariantAddressPair(0x91)
	account := Account{
		Version:	AccountVersionV2,
		AddressUser:	userAddress,
		AddressRaw:	rawAddress,
		PubKeys:	[]string{"ed25519:valid"},
		AccountNumber:	1,
		Sequence:	1,
		Status:		AccountStatusActive,
		AuthPolicy:	AuthPolicy{Version: 1, Mode: "single_key"},
		FeatureFlags: []string{
			AccountFeatureInternalMessagesV2,
			AccountFeatureMetadataV2,
			AccountFeatureRecoveryPolicyV2,
		},
		Metadata: AccountMetadata{
			MetadataHash:		"metadata-hash",
			DisplayNameHash:	"display-hash",
			DomainAlias:		"alice.aet",
			CreatedHeight:		1,
		},
		CreatedHeight:			1,
		LastActiveHeight:		2,
		LastStorageChargeHeight:	2,
	}
	if t != nil {
		require.NoError(t, ValidateAccountInvariant(account))
	}
	return NativeAccountInvariantInput{
		ExportedPayloads:	[]string{"account activated"},
		Accounts:		[]Account{account},

		AEAddressRoundtripStable:	true,
		RawAddressRoundtripStable:	true,
		ActivationAttempts:		map[string]uint64{account.AddressUser: 1},

		PoolActiveStake:	1_000,
		PoolUnbonding:		100,
		ValidatorSelfStake:	400_000,
		LiquidBalances:		500,
		TotalSupply:		500_000,

		RewardsAccrued:	50,
		RewardBudget:	100,

		ActiveValidatorCount:	100,
		MaxValidatorCount:	300,
		ValidatorStakes:	[]uint64{1_000_000, 1_100_000},
		MinValidatorStake:	1_000_000,
		ValidatorTotalStake:	1_000_000,
		MinSelfStakeBps:	4_000,
		PoolDeposits:		[]uint64{10, 11},
		MinPoolDeposit:		10,

		UnbondingEntries:	[]InvariantUnbondingEntry{{ReleaseHeight: 100, CurrentHeight: 100, Released: true}},

		StakeTimeDelta:		1,
		ReputationDelta:	1,

		ExportImportStable:	true,
		BankAccountingScenarios: []InvariantAccountingScenario{
			{Name: "pool deposit", BankBalance: 1_000, ModuleAccounting: 1_000},
		},

		SystemReserveRunwayBlocks:	100,
		MinSystemReserveRunwayBlocks:	10,
		SystemTopUpOrder:		[]string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:	true,
	}
}

func invariantAddressPair(fill byte) (string, string) {
	address := sdk.AccAddress(bytes20(fill))
	return addressing.FormatAccAddress(address), addressing.Format(address)
}
